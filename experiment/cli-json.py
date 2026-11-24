#!/usr/bin/env python3
"""Variante B de AIF: el evento `learning` recibe un objeto JSON por stdin.

Solo cambia la interfaz: la validación, las sesiones, el content-hash y la
persistencia son las de acm.py. Sin formulario AIF ni valores preservados.
"""
import importlib.util
import json
import sys
from pathlib import Path

FIELDS = ("title", "body", "topic", "confidence")

def help_text(acm):
    """Ayuda derivada de EVENT_FORMS: no duplica enumeraciones ni reglas."""
    rules = acm.EVENT_FORMS["learning"]
    required = ", ".join(n for n, r in rules.items() if r["required"])
    optional = ", ".join(n for n, r in rules.items() if not r["required"])
    enums = "\n".join(f"  {n} admite: {', '.join(r['allowed'])}"
                      for n, r in rules.items() if "allowed" in r)
    return f"""uso: cli-json.py [--schema] < entrada.json

Evento `learning` con un objeto JSON completo por stdin (sin flags):
  {{"session": "ses_x", "title": "...", "body": "...",
   "topic": "...", "confidence": "confirmed"}}

Obligatorios: {required}. Opcionales: {optional}.
Valores permitidos, según acm.py:
{enums}

Salida: una línea JSON en stdout.
  éxito            {{"ok":true,"result":"created"|"consolidated","id":...}}
  faltan campos    {{"ok":false,"error":"missing_required","fields":[...]}}
  campos inválidos {{"ok":false,"error":"invalid_field","fields":{{...}}}}
  error de dominio {{"ok":false,"error":"<código>","message":"..."}}
Exit: 0 éxito, 2 validación/uso, 3 no encontrado, 4 denegado, 9 interno.
"""


def load_acm():
    here = Path(__file__).resolve().parent
    for candidate in (here / "acm.py", here / "tools" / "acm.py"):
        if candidate.is_file():
            spec = importlib.util.spec_from_file_location("acm_lab", candidate)
            module = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(module)
            return module
    raise SystemExit(emit({"ok": False, "error": "E_INTERNAL",
                           "message": "no se encontró acm.py junto al adaptador"}, 9))


def emit(payload, code=0):
    print(json.dumps(payload, ensure_ascii=False, separators=(",", ":")))
    return code


def schema(acm):
    fields = {}
    for name, rule in acm.EVENT_FORMS["learning"].items():
        entry = {"required": rule["required"]}
        if "allowed" in rule:
            entry["allowed"] = list(rule["allowed"])
        fields[name] = entry
    return {"event": "learning", "input": "objeto JSON completo por stdin",
            "fields": fields}


def main(argv):
    acm = load_acm()
    if argv == ["--help"]:
        print(help_text(acm), end="")
        return 0
    if argv == ["--schema"]:
        return emit(schema(acm))
    if argv:
        return emit({"ok": False, "error": "E_USAGE",
                     "message": "no admite argumentos; la entrada va por "
                                "stdin como JSON (ver --help)"}, 2)

    try:
        payload = json.load(sys.stdin)
    except json.JSONDecodeError as exc:
        return emit({"ok": False, "error": "E_USAGE",
                     "message": f"la entrada no es JSON: {exc}"}, 2)
    if not isinstance(payload, dict):
        return emit({"ok": False, "error": "E_USAGE",
                     "message": "la entrada debe ser un objeto JSON"}, 2)

    unknown = sorted(set(payload) - set(FIELDS) - {"session"})
    if unknown:
        return emit({"ok": False, "error": "E_USAGE",
                     "message": "campos desconocidos: " + ", ".join(unknown)}, 2)

    sid = payload.get("session")
    if not isinstance(sid, str) or not acm.SESSION_RE.fullmatch(sid):
        return emit({"ok": False, "error": "invalid_field",
                     "fields": {"session": {
                         "value": sid,
                         "reason": "identificador de sesión: "
                                   f"{acm.SESSION_RE.pattern}"}}}, 2)

    values = {name: payload[name] for name in FIELDS if name in payload}
    form, invalid = acm.build_form("learning", sid, "", values)
    bad = {name: reason for name, reason in invalid.items()
           if reason != "obligatorio"}
    if bad:
        return emit({"ok": False, "error": "invalid_field",
                     "fields": {name: {"value": values.get(name),
                                       "reason": reason}
                                for name, reason in bad.items()}}, 2)
    if invalid:
        return emit({"ok": False, "error": "missing_required",
                     "fields": list(invalid),
                     "hint": "proporciona los campos listados"}, 2)

    try:
        result = acm.submit_event(
            "learning", sid,
            {name: form[name]["value"] for name in FIELDS})
    except acm.AcmError as exc:
        return emit({"ok": False, "error": exc.code, "message": exc.message},
                    acm.EXIT_CODES.get(exc.code, 1))
    return emit(result)


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
