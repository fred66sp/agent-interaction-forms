#!/usr/bin/env python3
"""Métricas del experimento AIF (runbook §7, gobernadas por PRD §7).

Determinista: lee experiment/runs/registro.jsonl y experiment/runs/config.json
sin tocar los brutos. Imprime el bloque de decisión en el orden de PRD §7:
integridad → medición completa → éxito y tokens → beneficio → GO.

Eficacia y descriptivos:
- tokens: suma de entrada más salida de todas las ejecuciones, sin descartar
  fallos; la caché se informa aparte (el proveedor la da).
- errores de contrato: reducción = 100 × (1 − errores_AIF / errores_CLI);
  no calculable si el denominador es 0.
- invocaciones en reparación: población fijada por el protocolo: los diez
  casos de reparación prefijados por variante (cinco de dato ausente y cinco
  de enum inválido), incluidas las fallidas, abandonadas o agotadas; media de
  llamadas por variante y reducción análoga.
- descriptivos: mediana de duración, p95 descriptivo, y camino directo (5 de
  5 en una llamada de operación) en el escenario completo.
"""
import json
import statistics
import sys
from pathlib import Path

RUNS = Path(__file__).resolve().parent / "runs"
REPAIR = {"dato-ausente", "enum-invalido"}


def load():
    records = []
    log = RUNS / "registro.jsonl"
    for line in log.read_text().splitlines():
        if line.strip():
            records.append(json.loads(line))
    return records


def check(records, condition):
    return [rec["id"] for rec in records if condition(rec)]


def main():
    records = load()
    if len(records) != 30:
        print(f"AVISO: {len(records)} registros de 30; cierre incompleto")
    out = {}

    # Integridad y completitud de la medición
    incompletas = check(records, lambda r: r["estado_medicion"] != "completa")
    no_comparables = check(records, lambda r: r["estado_medicion"]
                           == "no_comparable")
    efectos = [r for r in records if r["efectos_indebidos"]]
    integridad = {
        "efectos_indebidos_AIF": len([r for r in efectos
                                      if r["variante"] == "acm"]),
        "efectos_indebidos_total": len(efectos),
        "detalle": [{"id": r["id"], "variante": r["variante"],
                     "efectos": r["efectos_indebidos"]} for r in efectos],
    }
    medicion = {
        "completas": len(records) - len(incompletas),
        "incompletas": incompletas,
        "no_comparables": no_comparables,
    }

    # Éxito y tokens por variante
    per = {}
    for variant in ("acm", "cli"):
        sub = [r for r in records if r["variante"] == variant]
        ok = sum(1 for r in sub if r["exito_verificado_bd"])
        tokens = sum(r["tokens_entrada"] + r["tokens_salida"] for r in sub)
        periods = [r["tokens_cache_lectura"] for r in sub]
        walk = [r for r in sub if r["escenario"] == "completa"]
        directos = sum(1 for r in walk if r["llamadas"] == 1)
        durations = sorted(r["duracion_ms"] for r in sub)
        p95 = durations[int(0.95 * len(durations)) - 1] if durations else None
        per[variant] = {
            "ejecuciones": len(sub),
            "exito": ok,
            "exito_frac": ok / len(sub) if sub else None,
            "tokens_total": tokens,
            "tokens_media": tokens / len(sub) if sub else None,
            "cache_lectura_total": sum(periods),
            "duracion_mediana_ms": statistics.median(durations) if durations else None,
            "duracion_p95_ms": p95,
            "camino_directo": f"{directos} de {len(walk)}",
            "errores_contrato_total": sum(r["errores_contrato"] for r in sub),
            "ejecuciones_con_error_contrato": sum(
                1 for r in sub if r["errores_contrato"] > 0),
        }

    # Beneficio 1: errores de contrato
    e_aif = per["acm"]["errores_contrato_total"]
    e_cli = per["cli"]["errores_contrato_total"]
    if e_cli > 0:
        reduccion_errores = 100 * (1 - e_aif / e_cli)
    else:
        reduccion_errores = None  # no calculable

    # Beneficio 2: invocaciones en reparación (población del protocolo)
    repair = {}
    for variant in ("acm", "cli"):
        sub = [r for r in records if r["variante"] == variant
               and r["escenario"] in REPAIR]
        calls = [r["llamadas"] for r in sub]
        repair[variant] = {
            "casos": len(sub),  # 10 por variante: 5 ausente + 5 enum
            "media_llamadas": statistics.mean(calls) if calls else None,
            "composicion": {esc: sum(1 for r in sub if r["escenario"] == esc)
                            for esc in REPAIR},
        }
    if repair["cli"]["media_llamadas"]:
        reduccion_repar = 100 * (1 - repair["acm"]["media_llamadas"]
                                 / repair["cli"]["media_llamadas"])
    else:
        reduccion_repar = None

    # Puerta PRD §7 en orden
    decision_rows = []
    if integridad["efectos_indebidos_AIF"] > 0:
        decision = "NO-GO"
        decision_rows.append(f"(1) integridad: {integridad['efectos_indebidos_AIF']} "
                             "efectos indebidos en AIF → NO-GO")
    elif len(medicion["incompletas"]) or len(medicion["no_comparables"]):
        decision = "INCONCLUSO"
        decision_rows.append(f"(2) medición incompleta o no comparable "
                             f"({len(medicion['incompletas'])} incompletas, "
                             f"{len(medicion['no_comparables'])} no comparables) "
                             "→ INCONCLUSO")
    else:
        exito_ok = (per["acm"]["exito"] >= per["cli"]["exito"]
                    and per["acm"]["exito"] >= 14)
        tokens_ratio = (per["acm"]["tokens_total"] / per["cli"]["tokens_total"]
                        if per["cli"]["tokens_total"] else None)
        tokens_ok = tokens_ratio is not None and tokens_ratio <= 1.25
        if not (exito_ok and tokens_ok):
            decision = "NO-GO"
            if not exito_ok:
                decision_rows.append(
                    f"(3) éxito AIF {per['acm']['exito']}/15 vs CLI "
                    f"{per['cli']['exito']}/15 → NO-GO")
            if not tokens_ok:
                decision_rows.append(
                    f"(3) tokens ratio {tokens_ratio:.3f} > 1,25 → NO-GO")
        elif reduccion_errores is None and reduccion_repar is None:
            decision = "INCONCLUSO"
            decision_rows.append("(4) sin margen de mejora medible → INCONCLUSO")
        else:
            benefit_ok = ((reduccion_errores is not None
                           and reduccion_errores >= 30)
                          or (reduccion_repar is not None
                              and reduccion_repar >= 20))
            if not benefit_ok:
                decision = "NO-GO"
                data = [d for d in (
                    f"(5) errores de contrato: reducción "
                    f"{reduccion_errores:.1f}% < 30%"
                    if reduccion_errores is not None else None,
                    f"(5) invocaciones en reparación: reducción "
                    f"{reduccion_repar:.1f}% < 20%"
                    if reduccion_repar is not None else None) if d]
                decision_rows.append(" y ".join(data) + " → NO-GO")
            else:
                decision = "GO"
                decision_rows.append(
                    f"(6) los cuatro criterios cumplidos (errores "
                    f"{reduccion_errores:.1f}% >= 30% o reparación "
                    f"{reduccion_repar:.1f}% >= 20%) → GO")

    report = {
        "medicion": medicion,
        "integridad": integridad,
        "variantes": per,
        "errores_contrato": {"AIF": e_aif, "CLI": e_cli,
                             "reduccion_pct": reduccion_errores},
        "reparacion": repair,
        "reduccion_reparacion_pct": reduccion_repar,
        "tokens_ratio": (per["acm"]["tokens_total"] / per["cli"]["tokens_total"]
                         if per["cli"]["tokens_total"] else None),
        "decision_propuesta": decision,
        "orden_evaluado": decision_rows,
    }
    (RUNS / "metricas.json").write_text(
        json.dumps(report, ensure_ascii=False, indent=2))
    print(json.dumps(report, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
