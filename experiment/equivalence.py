#!/usr/bin/env python3
"""Equivalencia de garantías entre Variante A (flags) y B (JSON por stdin).

Sin modelo: rutas de garantía y de recuperación por variante contra una base
nueva restaurada de la plantilla, comprobando el contenido y estado reales de
las tablas de dominio con sqlite en modo ro. Solo se toleran diferencias de
forma en la respuesta; los efectos de dominio deben coincidir. Incluye la
prueba negativa del aislamiento de restore() y la coherencia de la ayuda y el
esquema del baseline con las reglas de acm.
"""
import importlib.util
import json
import os
import shutil
import sqlite3
import subprocess
import sys
import tempfile
from pathlib import Path

LAB = Path(os.environ.get("AIF_LAB", "/tmp/opencode/lab"))
TEMPLATE = LAB / "template-clean.db"
ACM = LAB / "tools" / "acm.py"
ADAPTER = LAB / "cli-json.py"
RW = LAB / "home" / ".config"
RO = LAB / "home-ro" / ".config"
NOMOUNT = LAB / "home-nomount" / ".config"
SID = "ses_equivalencia01"
SID_NEW = "ses_equivalencia02"
RESULTS = []


def lab_db(config_dir=RW):
    config = json.loads((config_dir / "db-ai" / "config.json").read_text())
    try:
        return Path(config["mounts"]["memo"]["file"])
    except KeyError as exc:
        raise SystemExit(f"aislamiento: config.json sin el montaje 'memo': {exc}")


DB = lab_db()


def lab_path(path):
    """Ruta real de escritura dentro del laboratorio, o aborto sin escribir.

    resolve() sigue symlinks: un enlace que apunte fuera del laboratorio
    resuelve fuera y se rechaza. Sin límites de carrera: es un arnés local,
    no un sistema multiusuario.
    """
    lab = LAB.resolve()
    resolved = Path(path).resolve(strict=False)
    if resolved != lab and lab not in resolved.parents:
        raise SystemExit(f"aislamiento: '{resolved}' queda fuera de '{lab}'")
    return resolved


def prepare_configs():
    """Deriva de la config rw: una copia en ro y otra sin montajes."""
    base = json.loads((RW / "db-ai" / "config.json").read_text())
    ro = json.loads(json.dumps(base))
    ro["mounts"]["memo"]["mode"] = "ro"
    none = json.loads(json.dumps(base))
    none["mounts"] = {}
    for directory, config in ((RO, ro), (NOMOUNT, none)):
        target = lab_path(directory / "db-ai")
        target.mkdir(parents=True, exist_ok=True)
        (target / "config.json").write_text(json.dumps(config, indent=2))


def env(config_dir):
    keep = {k: v for k, v in os.environ.items()
            if k not in ("OPENCODE_SESSION_ID", "XDG_CONFIG_HOME")}
    keep.update({"HOME": str(LAB / "home"), "XDG_CONFIG_HOME": str(config_dir),
                 "ACM_HOME": str(LAB / "acm-root"), "ACM_ALIAS": "memo",
                 "ACM_AGENT_ID": "exp-agent",
                 "PATH": f"{LAB / 'bin'}:{os.environ.get('PATH', '')}"})
    return keep


def restore(template=TEMPLATE, config_dir=RW):
    """Copia la plantilla a la base del montaje tras validar el aislamiento.

    Toda comprobación precede a la primera escritura o borrado: una config
    maliciosa que apunte fuera del laboratorio (directa o vía symlink) aborta
    sin tocar nada.
    """
    db = lab_path(lab_db(config_dir))
    template = Path(template).resolve(strict=False)
    if template == db:
        raise SystemExit("aislamiento: la plantilla y la base de destino son "
                         "el mismo archivo")
    extras = [lab_path(Path(str(db) + suffix)) for suffix in ("-wal", "-shm")]
    for extra in extras:
        if extra.exists():
            extra.unlink()
    shutil.copyfile(template, db)


# Campos que varían entre ejecuciones (identificadores y sellos de tiempo);
# dentro de una misma ejecución ni el contenido de dominio ni el estado
# cambian sin que el oráculo lo detecte.
VOLATILE = {
    "sessions": ("id", "started_at", "created_at", "updated_at", "ended_at"),
    "observations": ("id", "created_at", "updated_at", "last_seen_at",
                     "review_after"),
    "checkpoints": ("id", "created_at", "updated_at"),
    "preferences": ("id", "created_at", "updated_at"),
}


def snapshot():
    con = sqlite3.connect(f"file:{DB}?mode=ro", uri=True)
    con.row_factory = sqlite3.Row
    try:
        cur = con.cursor()
        return {table: [dict(row) for row in cur.execute(f"select * from {table}")]
                for table in ("sessions", "observations", "checkpoints",
                              "preferences")}
    finally:
        con.close()


def domino(blob):
    """Filas de dominio sin campos volátiles: el contenido comparable."""
    return {table: [{c: row[c] for c in row if c not in VOLATILE[table]}
                    for row in rows]
            for table, rows in blob.items()}


def one_observation(blob, title, body, topic, confidence, dup=1):
    """Exactamente la observación esperada; checkpoints y preferences vacíos."""
    assert not blob["checkpoints"] and not blob["preferences"], blob
    rows = domino({"observations": blob["observations"]})["observations"]
    assert len(rows) == 1, blob
    obs = rows[0]
    assert (obs["kind"], obs["title"], obs["body"], obs["topic_key"],
            obs["confidence"]) == ("learning", title, body, topic,
                                   confidence), obs
    assert (obs["status"], obs["duplicate_count"], obs["session_id"],
            obs["agent_id"]) == ("active", dup, SID, "exp-agent"), obs
    return obs


def open_session(blob, sid):
    """La sesión abierta esperada, sin cambios de contenido."""
    assert len(blob["sessions"]) == 1 and blob["sessions"][0]["id"] == sid, blob
    session = domino({"sessions": blob["sessions"]})["sessions"][0]
    assert session["agent_id"] == "exp-agent" and session["status"] == "open", session
    return session


def run(args, config_dir, stdin=None):
    proc = subprocess.run(args, env=env(config_dir), input=stdin,
                          capture_output=True, text=True, encoding="utf-8")
    return proc.returncode, proc.stdout.strip(), proc.stderr.strip()


def payload(stdout):
    try:
        return json.loads(stdout)
    except json.JSONDecodeError:
        return {}


def start_session(sid):
    code, out, err = run([sys.executable, str(ACM), "start", "--session", sid], RW)
    assert code == 0, f"no se pudo abrir {sid}: {out}{err}"


def write(variant, sid, title, body, topic=None, confidence=None,
          config_dir=RW):
    if variant == "A":
        args = [sys.executable, str(ACM), "learning", "--session", sid,
                "--title", title, "--body", body]
        if topic is not None:
            args += ["--topic", topic]
        if confidence is not None:
            args += ["--confidence", confidence]
        return run(args, config_dir)
    data = {"session": sid, "title": title, "body": body}
    if topic is not None:
        data["topic"] = topic
    if confidence is not None:
        data["confidence"] = confidence
    return run([sys.executable, str(ADAPTER)], config_dir,
               stdin=json.dumps(data))


def route1(variant):
    restore()
    start_session(SID)
    code, out, err = write(variant, SID, "AIF equivalencia", "cuerpo completo",
                           "equivalencia", "confirmed")
    assert code == 0, f"exit {code}: {out}{err}"
    assert payload(out).get("result") == "created", out
    blob = snapshot()
    one_observation(blob, "AIF equivalencia", "cuerpo completo",
                    "equivalencia", "confirmed")
    open_session(blob, SID)
    return ("exit 0, sesion abierta, 1 observacion learning con campos, "
            "actor y sesion, active, dup 1; checkpoints/preferences vacios")


def route2(variant):
    restore()
    before = snapshot()
    code, out, _ = write(variant, SID_NEW, "T inválida", "cuerpo", None, "confimed")
    assert code != 0, f"exit {code}: aceptó confidence inválido"
    data = payload(out)
    assert data.get("ok") is False, out
    assert snapshot() == before, "el rechazo modificó la base"
    if variant == "A":
        form = data.get("form") or {}
        assert form.get("title", {}).get("value") == "T inválida", \
            "A no preserva valores en el formulario"
        note = "exit 2, formulario con valores preservados, base intacta"
    else:
        assert "form" not in data, "B devolvió formulario"
        assert data.get("error") == "invalid_field", out
        assert "T inválida" not in out, "B filtró un valor en la respuesta"
        note = "exit 2, error JSON invalid_field, sin valores, base intacta"
    return note


def route3(variant):
    restore()
    start_session(SID)
    before = snapshot()
    code, out, _ = write(variant, SID, "Denegada", "cuerpo", None, None, RO)
    assert code != 0, "la escritura en ro devolvió éxito"
    data = payload(out)
    assert data.get("error") == "E_DENIED" and code == 4, f"exit {code}: {out}"
    assert snapshot() == before, "la denegación modificó la base"
    return "exit 4 E_DENIED con mount ro, base idéntica (nada en ninguna tabla)"


def route4(variant):
    restore()
    start_session(SID)
    before = snapshot()
    first = write(variant, SID, "Duplicada", "mismo cuerpo", "dup", "inference")
    second = write(variant, SID, "Duplicada", "mismo cuerpo", "dup", "inference")
    assert first[0] == 0 and payload(first[1]).get("result") == "created", first
    assert second[0] == 0, second
    assert payload(second[1]).get("result") == "consolidated", second[1]
    after = snapshot()
    other = {t: r for t, r in after.items() if t != "observations"}
    other_before = {t: r for t, r in before.items() if t != "observations"}
    assert other == other_before, "la consolidación tocó otra tabla"
    one_observation(after, "Duplicada", "mismo cuerpo", "dup", "inference",
                    dup=2)
    blob = after["observations"]
    assert blob[0]["duplicate_count"] == 2 and blob[0]["session_id"] == SID, blob
    return ("segunda ejecución consolidated, 1 fila, duplicate_count 2, "
            "resto de tablas idénticas")


def route5(variant):
    restore()
    start_session(SID)
    before = snapshot()
    code, out, _ = write(variant, SID, "Sin dominio", "cuerpo", None, None,
                         NOMOUNT)
    assert code != 0, "sin montajes devolvió éxito"
    data = payload(out)
    assert data.get("error") in ("E_NOT_FOUND", "E_INTERNAL"), f"exit {code}: {out}"
    assert snapshot() == before, "el fallo de dominio modificó la base"
    again = write(variant, SID, "Sin dominio", "cuerpo", None, None)
    assert again[0] == 0 and payload(again[1]).get("result") == "created", again
    blob = snapshot()
    one_observation(blob, "Sin dominio", "cuerpo", "", "confirmed")
    open_session(blob, SID)
    return (f"exit {code} {data['error']} sin efectos; restaurado el mount, "
            "created con su observación completa")


def continue_via_submit(sid, form_id, values):
    """Mecanismo real de continuación de ACM/AIF: `acm submit` con form_id."""
    data = {"event": "learning", "form_id": form_id, "session": sid, **values}
    return run([sys.executable, str(ACM), "submit", "--session", sid,
                "--data", "-"], RW, stdin=json.dumps(data))


def route6_missing(variant):
    """Recuperación por campo obligatorio ausente, completada en ambas."""
    title, body, topic, conf = ("Reparación ausente", "cuerpo reparado",
                                "recuperacion", "confirmed")
    restore()
    start_session(SID)
    before = snapshot()
    if variant == "A":
        code, out, err = run([sys.executable, str(ACM), "learning",
                              "--session", SID, "--title", title,
                              "--topic", topic, "--confidence", conf], RW)
        assert code == 0, f"exit {code}: {out}{err}"
        data = payload(out)
        assert data.get("result") == "form", out
        form = data["form"]
        assert form["title"]["value"] == title, "A no conserva el título válido"
        assert (form["topic"]["value"], form["confidence"]["value"]) == (
            topic, conf), "A no conserva los valores válidos"
        assert snapshot() == before, "el formulario produjo efectos de dominio"
        code, out, err = continue_via_submit(
            SID, form["form_id"], {"title": title, "body": body,
                                   "topic": topic, "confidence": conf})
    else:
        code, out, _ = run([sys.executable, str(ADAPTER)], RW, stdin=json.dumps(
            {"session": SID, "title": title, "topic": topic,
             "confidence": conf}))
        assert code != 0, out
        data = payload(out)
        assert data.get("error") == "missing_required" \
            and data.get("fields") == ["body"], out
        assert snapshot() == before, "el rechazo produjo efectos de dominio"
        # El baseline no conserva estado: se reconstruye el objeto y se reenvía.
        code, out, _ = write(variant, SID, title, body, topic, conf)
    assert code == 0 and payload(out).get("result") == "created", \
        f"exit {code}: {out}"
    blob = snapshot()
    one_observation(blob, title, body, topic, conf)
    open_session(blob, SID)
    return "primera llamada correcta sin efectos; completada con 1 observación"


def route7_enum(variant):
    """Recuperación por enum inválido, completada en ambas."""
    title, body, topic = "Reparación enum", "cuerpo reparado", "recuperacion"
    restore()
    start_session(SID)
    before = snapshot()
    if variant == "A":
        code, out, err = run([sys.executable, str(ACM), "learning",
                              "--session", SID, "--title", title,
                              "--body", body, "--topic", topic,
                              "--confidence", "confimed"], RW)
        assert code != 0, "aceptó confidence inválido"
        data = payload(out)
        form = data["form"]
        assert form["confidence"]["value"] == "confimed", out
        assert (form["title"]["value"], form["body"]["value"],
                form["topic"]["value"]) == (title, body, topic), \
            "A no conserva los valores válidos"
        assert "valores permitidos" in data["invalid"]["confidence"], out
        assert snapshot() == before, "el rechazo produjo efectos de dominio"
        code, out, err = continue_via_submit(
            SID, form["form_id"], {"title": title, "body": body,
                                   "topic": topic,
                                   "confidence": "confirmed"})
    else:
        code, out, _ = run([sys.executable, str(ADAPTER)], RW, stdin=json.dumps(
            {"session": SID, "title": title, "body": body, "topic": topic,
             "confidence": "confimed"}))
        assert code != 0, out
        data = payload(out)
        assert data.get("error") == "invalid_field", out
        reason = data["fields"]["confidence"]["reason"]
        assert "valores permitidos" in reason, reason
        assert snapshot() == before, "el rechazo produjo efectos de dominio"
        code, out, _ = write(variant, SID, title, body, topic, "confirmed")
    assert code == 0 and payload(out).get("result") == "created", \
        f"exit {code}: {out}"
    blob = snapshot()
    one_observation(blob, title, body, topic, "confirmed")
    open_session(blob, SID)
    return "primera llamada correcta sin efectos; corregida y creada"


def route_errores():
    """La ayuda, el esquema y el motivo coinciden con las reglas de acm."""
    _, schema_out, _ = run([sys.executable, str(ADAPTER), "--schema"], RW)
    allowed = json.loads(schema_out)["fields"]["confidence"]["allowed"]
    _, help_out, _ = run([sys.executable, str(ADAPTER), "--help"], RW)
    assert "Obligatorios: title, body." in help_out, help_out
    assert f"confidence admite: {', '.join(allowed)}" in help_out, help_out
    restore()
    start_session(SID)
    before = snapshot()
    code, out, _ = run([sys.executable, str(ADAPTER)], RW, stdin=json.dumps(
        {"session": SID, "title": "X", "body": "y",
         "confidence": "confimed"}))
    assert code != 0, out
    reason = payload(out)["fields"]["confidence"]["reason"]
    assert reason == "valores permitidos: " + ", ".join(allowed), reason
    assert snapshot() == before, "el rechazo produjo efectos de dominio"
    return ("ayuda y esquema derivados; invalid_field con valores permitidos; "
            "base intacta")


def test_aislamiento():
    """Config maliciosa: nada fuera del laboratorio se escribe o borra."""
    base = Path(tempfile.mkdtemp(prefix="aif-negativa-", dir="/tmp/opencode"))
    victima = base / "victima.db"      # fuera del LAB: solo /tmp/opencode/lab
    plantilla = base / "plantilla.db"
    victima.write_bytes(b"datos-previos")
    plantilla.write_bytes(b"plantilla-negativa")
    config_dir = base / "cfg" / "db-ai"
    config_dir.mkdir(parents=True)
    enlace = LAB / "symlink-negativa.db"
    try:
        os.symlink(victima, enlace)
        # 1) destino directo fuera del laboratorio
        (config_dir / "config.json").write_text(json.dumps(
            {"mounts": {"memo": {"file": str(victima)}}}))
        try:
            restore(plantilla, base / "cfg")
            return "no abortó con destino externo directo"
        except SystemExit:
            if victima.read_bytes() != b"datos-previos":
                return "el destino externo directo quedó modificado"
        # 2) destino dentro del lab vía symlink que apunta fuera
        (config_dir / "config.json").write_text(json.dumps(
            {"mounts": {"memo": {"file": str(enlace)}}}))
        try:
            restore(plantilla, base / "cfg")
            return "no abortó con symlink que escapa"
        except SystemExit:
            if victima.read_bytes() != b"datos-previos":
                return "el destino del symlink quedó modificado"
        # 3) plantilla y destino son el mismo archivo
        (config_dir / "config.json").write_text(json.dumps(
            {"mounts": {"memo": {"file": str(plantilla)}}}))
        try:
            restore(plantilla, base / "cfg")
            return "no abortó con plantilla == destino"
        except SystemExit:
            pass
    finally:
        enlace.unlink(missing_ok=True)
        shutil.rmtree(base, ignore_errors=True)
    return ("destino externo, symlink que escapa y plantilla==destino "
            "abortados sin escribir; víctima sintética intacta")


ROUTES = (
    ("ruta1-completa", route1),
    ("ruta2-invalida", route2),
    ("ruta3-denegada", route3),
    ("ruta4-duplicada", route4),
    ("ruta5-fallo-dominio", route5),
    ("ruta6-recupera-ausente", route6_missing),
    ("ruta7-recupera-enum", route7_enum),
)


def check(label, ok, detail):
    print(f"{'PASS' if ok else 'FAIL'} {label} {detail}")
    RESULTS.append(ok)


def guarded(label, fn, *args, prefix=""):
    try:
        check(label, True, prefix + fn(*args))
    except AssertionError as exc:
        check(label, False, prefix + str(exc))
    except Exception as exc:  # noqa: BLE001 - el guion no debe abortar
        check(label, False, f"{prefix}{type(exc).__name__}: {exc}")


def main():
    prepare_configs()
    restore()
    template = domino(snapshot())

    for variant in ("A", "B"):
        for name, route in ROUTES:
            guarded(name, route, variant, prefix=f"{variant}: ")

    guarded("errores-accionables", route_errores)
    guarded("aislamiento-negativa", test_aislamiento)

    print("INFO forma-de-error A: formulario AIF con valores preservados; "
          "B: error JSON sin valores (diferencia declarada, sin efecto de "
          "dominio)")

    restore()
    clean = domino(snapshot()) == template
    check("estado-final", clean, "base restaurada: contenido de dominio igual "
          "a la plantilla")
    return 0 if all(RESULTS) else 1


if __name__ == "__main__":
    sys.exit(main())
