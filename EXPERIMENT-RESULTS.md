# AIF — Resultados del experimento previo a Specification

| Metadato | Valor |
|---|---|
| ID | AIF-EXP-RESULTS |
| Versión | 1.0.0 |
| Fecha | 2026-11-24 |
| Fuentes | [PRD.md](PRD.md) §7 · [EXPERIMENT.md](EXPERIMENT.md) v0.1.6 · [EXPERIMENT-RUNBOOK.md](EXPERIMENT-RUNBOOK.md) v0.1.0 |
| Estado | Experimento clasificado **INCONCLUSO** por el PO (opción B, 2026-11-24) |
| Decisión de producto | Continuar con el proyecto AIF asumiendo la incertidumbre residual; **no equivale a un GO experimental**. Specification no se inicia en este encargo |

## 1. Qué se ejecutó

Las 30 mediciones del protocolo aprobado (EXPERIMENT v0.1.6, D-08): 3 escenarios
(entrada completa, dato obligatorio ausente, enum inválido) × 5 pares × 2
variantes, orden alternado ACM/CLI por par, sesiones independientes del modelo
`nan/deepseek-v4-flash#medium` vía OpenCode v2.4.1, límite de 4 invocaciones de
herramienta y 120 s por ejecución, fixture de usuario aprobado, prefijada del
operador en los escenarios de reparación (intervención registrada), base ACM
desechable restaurada de plantilla antes de cada ejecución y oráculo de éxito
verificado sobre la base de datos (no sobre `ok:true`).

El arnés congelado en `main` (merge `2e47cc0`) no se modificó. El instrumento de
medida (driver) vivió fuera del repositorio, en el laboratorio `/tmp/opencode/lab`.

## 2. Incidencias de infraestructura previas a la medición

Documentadas en [`experiment/runs/incidencias-previas/`](runs/incidencias-previas/README.md):

1. Runtime de Node de OpenCode no descargable al HOME aislado (sin red).
2. El `opencode` primero en PATH es el conmutador de vite-plus (`vp`), que bajo
   el entorno aislado no delega («Package not found in workspace»).
3. Sin variables de sesión, `vp` interpretó los flags como tareas propias.

Resolución: invocación del binario OpenCode v2.4.1 (el que produjo la telemetría
validada en la Fase A) por ruta absoluta, con caché real compartida (solo
binarios y `models.json`; sesiones y configuración aisladas en el laboratorio).
Ninguna tentativa consumió muestra: abortaron antes de producir salida del modelo.

## 3. Resultados observados [INVALIDADOS como entrada de la puerta — véase §4]

Campaña ejecutada completa: 30/30 mediciones con `estado_medicion=completa`,
sin timeouts ni abandonos de infraestructura.

| Métrica (runbook §7) | AIF (ACM) | CLI JSON baseline |
|---|---|---|
| Éxito verificado en base de datos | 8/15 | 8/15 |
| Tokens totales (entrada+salida) | 37.487 | 53.715 |
| Ratio de tokens | **0,698** (AIF ≤ 125 %: cumple) | referencia |
| Errores de contrato | 0 | 0 (reducción no calculable) |
| Efectos indebidos de integridad | 0 | 0 |
| Invocaciones en reparación (media, 10 casos) | 2,3 | 2,1 (reducción −9,5 %) |
| Duración mediana (ms) | 13.439 | 12.870 |
| Duración p95 (ms, descriptiva) | 19.000 | 30.265 |
| Camino directo (operación sin ayuda previa) | 0 de 5 | 0 de 5 |
| Tokens de caché de lectura (informativo) | 426.496 | 479.360 |

Límites operativos: en las 30 ejecuciones no hubo timeouts ni abortos de
infraestructura; en 14 la ejecución terminó sin éxito por decisión del modelo
(abandono registrado); en 2 (runs 12 y 16, CLI, dato-ausente) el modelo superó
el límite de 4 invocaciones con una quinta llamada, ambas además fallidas
(registrado como exceso de invocaciones; sin re-descuento de la puerta porque
el protocolo no contempla descartes).

Datos por ejecución: [`runs/registro.jsonl`](experiment/runs/registro.jsonl) (14 campos del
runbook §6 más auditoría); descriptivos y puerta:
[`runs/metricas.json`](experiment/runs/metricas.json) (calculados por
[`experiment/metrics.py`](experiment/metrics.py)).

## 4. Incidencia 2 — fixture de usuario retenido (invalidación)

Documentación completa: [`runs/INCIDENCIA-FIXTURE.md`](runs/INCIDENCIA-FIXTURE.md).

El detector del fixture del instrumento clasificó como «éxito» frases negativas
del modelo (la subcadena `registrad` casa con «No he registrado nada todavía»)
y **no entregó el fixture de usuario en 9 de las 20 ejecuciones de reparación**
(AIF 5: runs 11, 14, 19, 21, 28; CLI 4: runs 17, 20, 23, 26) pese a que el
modelo pedía el dato. Esto vulnera la condición aprobada de EXPERIMENT §1 (mismo
fixture para ambas variantes cuando el modelo no dispone del dato) y contamina
directamente la métrica de éxito, criterio de la puerta. Con la decisión del PO
(opción B) la campaña **no se repite**: los resultados de §3 describen lo
ocurrido bajo una condición experimental parcialmente vulnerada y no sustentan
ningún criterio de la puerta.

## 5. Garantías funcionales verificadas vs beneficios cuantitativos no demostrados

**Verificado en esta campaña (independiente del defecto del fixture):**

- Éxito verificable en base de datos: el mecanismo de un extremo a otro funciona
  — una observación `learning` con los campos previstos en 16 de 30 ejecuciones
  (8 AIF y 8 CLI, el acierto del oráculo) y con observación presente pero
  malformada o duplicada en otras 5 (bodies inventados, frase del fixture pegada
  como body, reenvíos con `duplicate_count=2`); cero mutaciones inesperadas en
  las demás.
- Integridad: **cero efectos indebidos** en ambas variantes — ninguna sesión
  ajena, sin checkpoints ni preferences, `agent_id` correcto, entradas inválidas
  sin efecto de dominio (RNF-05).
- Errores de contrato: cero en ambas variantes; los modelos no llamaron nunca
  con el enum inválido (el bloque de la prefijada les dio los valores
  permitidos), lo que hace la reducción de la puerta no calculable.
- Equivalencia de garantías entre variantes sin modelo: 17/17 PASS
  (Fase B, re-validada antes de medir).
- Telemetría completa por respuesta en las 30 ejecuciones.

**No demostrado (beneficios cuantitativos de la puerta):**

- Superioridad o paridad de éxito: 8/15 en ambas variantes queda por debajo del
  umbral 14/15, pero bajo condición parcialmente vulnerada (§4): no interpretable.
- Reducción de errores de contrato: no calculable (ambos a cero): sin margen
  medible de ningún tipo.
- Reducción de invocaciones en reparación: medida en −9,5 % (AIF peor), no
  interpretable por la contaminación del fixture y, aun válida, lejos del 20 %.
- Eficiencia de tokens: ratio 0,698 a favor de AIF, medido íntegramente en las
  30 ejecuciones tal como ocurrieron; sigue siendo una observación bajo
  condición vulnerada (el flujo alterado consume tokens distintos), no un
  criterio de puerta superado.

## 6. Limitaciones

1. **Condición experimental vulnerada** (fixture retenido en 9/20 reparaciones):
   invalida la comparación para la puerta; su opción B la convierte en definitiva
   para este experimento.
2. **Temperatura efectiva en runtime: NO EJECUTADA.** Configurada a 0 en la
   configuración del laboratorio y configurabilidad del proveedor acreditada en
   la Fase A; la comprobación directa exigía desplazar el puerto del servicio
   gestionado de OpenCode (configuración global no autorizada).
3. **Caché de lectura compartida con la máquina**: runtime y binarios de
   OpenCode provienen del caché real del usuario (sin red en el laboratorio);
   el contenido cacheado del prompt del sistema depende de la misma máquina en
   ambas variantes, simétrico, pero no replicable en una máquina distinta.
4. **n=5 pares por escenario**: los umbrales son criterios de producto, no
   significación estadística (EXPERIMENT §3); ninguna métrica admite
   inferencia poblacional.
5. **Instrumento artesanal del operador** (driver + auditoría): dos defectos
   detectados y documentados (contador de errores de contrato por escapado,
   corregido y re-verificado desde la base de sesiones; detector del fixture,
   asumido sin corrección por decisión del PO), lo que acota cuánta confianza
   cabe en cualquier medición automática de interacción agente-herramienta.
6. El éxito quedó además limitado por comportamientos genuinos del modelo
   (reenvíos con `duplicate_count=2`, bodies inventados, fixture pegado literal
   como body), visibles en los brutos.

## 7. Clasificación formal

**INCONCLUSO** (PRD §7, orden de evaluación, regla 2: medición incompleta o no
comparable → INCONCLUSO). Fundamento: la condición experimental no se sostuvo
en 9 de 20 ejecuciones de reparación (§4); la medición no es comparable.

Observación registral (sin valor de decisión): de haber sido la medición válida,
el orden de la puerta habría concluido NO-GO por la regla 3 (éxito 8/15 < 14/15
en ambas variantes) y la cadena habría terminado ahí: con el denominador de
reparación calculable (−9,5 % < 20 %), la regla 5 reafirmaría NO-GO; la regla 4
(sin margen de mejora medible) no se alcanzó porque la reducción de invocaciones
en reparación sí es calculable. Ningún resultado observado apunta a un GO no
registrado.

## 8. Decisión del PO

- **INCONCLUSO**, opción B: no se repite la campaña ni se corrige el instrumento
  (2026-11-24).
- Se continúa el proyecto AIF **asumiendo explícitamente la incertidumbre
  residual**; esta decisión de producto no equivale a un GO experimental y no
  inicia Specification (encargo posterior).

## 9. Verificación independiente

Realizada por agente revisor sin participación en la medición (2026-11-24):
8/8 comprobaciones conformes (conteo 30/30 con cobertura completa, éxito 8/15,
tokens y ratio, errores y efectos indebidos a cero tras auditoría, fixture
retenido confirmado con brutos en runs 11 y 28, clasificación INCONCLUSO según
PRD §7 regla 2, limitaciones declaradas sin silencios, descriptivos
reproducidos). Hallazgos menores resueltos antes del cierre: H1 (observación
registral precisa ahora sobre el orden real de la puerta), H2 (exceso de
invocaciones de los runs 12 y 16 declarado en §3), H3 (desglose por celda
corregido en INCIDENCIA-FIXTURE.md). No verificado por el revisor: la
«equivalencia 17/17» procede de la Fase B (evidencia del tablero tk-3, fuera
del alcance del revisor).

## 10. Trazabilidad y tareas del tablero

- tk-4 (Fase C, medición): ejecutada 30/30; invalidada como entrada de la puerta
  por INCIDENCIA-FIXTURE.md; cierre con comentario de evidencia.
- tk-5 (Fase D, métricas): calculadas deterministamente; no aplicadas a la
  puerta por §4; cierre con comentario de evidencia.
- tk-6 (Fase E, verificación y resultados): verificación independiente 8/8 y
  este informe; cierre con comentario de evidencia.
- tk-1 (épica): se cierra cuando cierran sus tres hijas.

- Registro por ejecución, descriptivos y métricas: `experiment/runs/`
  (`registro.jsonl`, `metricas.json`, `auditoria.json`, `config.json` del
  experimento con versiones y hash del arnés e instrumentos).
- Brutos por ejecución: `runs/<nn>-<variante>-<escenario>-p<n>/`
  (`registro.json`, `opencode.r1.jsonl`, `herramientas.jsonl`, `rondas.json`,
  `snapshot_final.json`, `montaje.json`, `prefijada.json`, `textos_agente.txt`).
- Incidencias: `runs/incidencias-previas/` (arranque) y
  `runs/INCIDENCIA-FIXTURE.md` (invalidación).
- Instrucciones: `experiment/metrics.py` (métricas deterministas),
  auditoría y volcado en el laboratorio (`/tmp/opencode/lab/auditoria.py`,
  `volcado.py`, `driver.py` — fuera del repositorio).
- Tablero: tk-4 (Fase C), tk-5 (Fase D), tk-6 (Fase E) bajo la épica tk-1.
