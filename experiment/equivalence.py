#!/usr/bin/env python3
"""Equivalencia de garantías entre Variante A (flags) y B (JSON por stdin).

Sin modelo: cinco rutas por variante contra una base nueva restaurada de la
plantilla, comprobando el estado real con sqlite en modo ro. Solo se toleran
diferencias de forma en la respuesta; los efectos de dominio deben coincidir.
"""
import json
import os
import shutil
import sqlite3
import subprocess
import sys
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


def lab_db():
    config = json.loads((RW / "db-ai" / "config.json").read_text())
    return Path(config["mounts"]["memo"]["file"])


DB = lab_db()


def prepare_configs():
    """Deriva de la config rw: una copia en ro y otra sin montajes."""
    base = json.loads((RW / "db-ai" / "config.json").read_text())
    ro = json.loads(json.dumps(base))
    ro["mounts"]["memo"]["mode"] = "ro"
    none = json.loads(json.dumps(base))
    none["mounts"] = {}
    for directory, config in ((RO, ro), (NOMOUNT, none)):
        target = directory / "db-ai"
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


def restore():
    shutil.copyfile(TEMPLATE, DB)
    for suffix in ("-wal", "-shm"):
        extra = Path(str(DB) + suffix)
        if extra.exists():
            extra.unlink()


def snapshot():
    con = sqlite3.connect(f"file:{DB}?mode=ro", uri=True)
    try:
        cur = con.cursor()
        counts = {table: cur.execute(f"select count(*) from {table}").fetchone()[0]
                  for table in ("sessions", "observations", "checkpoints",
                                "preferences")}
        observations = cur.execute(
            "select kind, title, body, topic_key, confidence, status,"
            " duplicate_count, session_id from observations").fetchall()
    finally:
        con.close()
    return counts, observations


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
    counts, observations = snapshot()
    assert counts["observations"] == 1, counts
    kind, title, body, topic, confidence, status, dup, sid = observations[0]
    assert (kind, title, body, topic, confidence) == (
        "learning", "AIF equivalencia", "cuerpo completo", "equivalencia",
        "confirmed"), observations[0]
    assert (status, dup, sid) == ("active", 1, SID), observations[0]
    return "exit 0, 1 observación learning con sus campos, active, duplicate_count 1"


def route2(variant):
    restore()
    before, _ = snapshot()
    code, out, _ = write(variant, SID_NEW, "T inválida", "cuerpo", None, "confimed")
    assert code != 0, f"exit {code}: aceptó confidence inválido"
    data = payload(out)
    assert data.get("ok") is False, out
    after, _ = snapshot()
    assert after == before, f"efectos laterales: {before} -> {after}"
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
    code, out, _ = write(variant, SID, "Denegada", "cuerpo", None, None, RO)
    assert code != 0, "la escritura en ro devolvió éxito"
    data = payload(out)
    assert data.get("error") == "E_DENIED" and code == 4, f"exit {code}: {out}"
    counts, observations = snapshot()
    assert counts["observations"] == 0 and not observations, counts
    return "exit 4 E_DENIED con mount ro, cero observaciones"


def route4(variant):
    restore()
    start_session(SID)
    first = write(variant, SID, "Duplicada", "mismo cuerpo", "dup", "inference")
    second = write(variant, SID, "Duplicada", "mismo cuerpo", "dup", "inference")
    assert first[0] == 0 and payload(first[1]).get("result") == "created", first
    assert second[0] == 0, second
    assert payload(second[1]).get("result") == "consolidated", second[1]
    counts, observations = snapshot()
    assert counts["observations"] == 1, counts
    assert observations[0][6] == 2, observations[0]
    return "segunda ejecución consolidated, 1 fila, duplicate_count 2"


def route5(variant):
    restore()
    start_session(SID)
    code, out, _ = write(variant, SID, "Sin dominio", "cuerpo", None, None,
                         NOMOUNT)
    assert code != 0, "sin montajes devolvió éxito"
    data = payload(out)
    assert data.get("error") in ("E_NOT_FOUND", "E_INTERNAL"), f"exit {code}: {out}"
    counts, observations = snapshot()
    assert counts["observations"] == 0, counts
    again = write(variant, SID, "Sin dominio", "cuerpo", None, None)
    assert again[0] == 0 and payload(again[1]).get("result") == "created", again
    return f"exit {code} {data['error']} sin efectos; con el mount restaurado, created"


ROUTES = (
    ("ruta1-completa", route1),
    ("ruta2-invalida", route2),
    ("ruta3-denegada", route3),
    ("ruta4-duplicada", route4),
    ("ruta5-fallo-dominio", route5),
)


def check(label, ok, detail):
    print(f"{'PASS' if ok else 'FAIL'} {label} {detail}")
    RESULTS.append(ok)


def main():
    prepare_configs()
    restore()
    template_counts, _ = snapshot()

    for variant in ("A", "B"):
        for name, route in ROUTES:
            try:
                detail = route(variant)
                check(name, True, f"{variant}: {detail}")
            except AssertionError as exc:
                check(name, False, f"{variant}: {exc}")
            except Exception as exc:  # noqa: BLE001 - el guion no debe abortar
                check(name, False, f"{variant}: {type(exc).__name__}: {exc}")

    print("INFO forma-de-error A: formulario AIF con valores preservados; "
          "B: error JSON sin valores (diferencia declarada, sin efecto de "
          "dominio)")

    restore()
    final_counts, final_observations = snapshot()
    clean = final_counts == template_counts and not final_observations
    check("estado-final", clean, f"base restaurada: {final_counts}")
    return 0 if all(RESULTS) else 1


if __name__ == "__main__":
    sys.exit(main())
