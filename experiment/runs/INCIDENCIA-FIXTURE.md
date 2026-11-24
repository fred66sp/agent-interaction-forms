# Incidencia 2 — fixture de usuario retenido por el instrumento (2026-11-24)

**Estado: campaña de 30 mediciones ejecutada completa, INVALIDADA para la
puerta de producto por defecto del instrumento. Sin decisión GO/NO-GO.**

## Qué pasó

El driver decide entregar el fixture de usuario cuando el último texto del
agente pide datos (`driver.ASK_HINTS`) y no declara éxito
(`driver.SUCCESS_HINT`). Esa segunda expresión incluía la subcadena
`registrad`, que casa con frases negativas del modelo:

> «No he **registrad**o nada todavía, y me detengo…» (runs 11, 28)

Resultado: el fixture no se entregó en **9 de las 20 ejecuciones de
reparación** pese a que el modelo pedía el dato:

| Ejecuciones afectadas | Variante | Escenario |
|---|---|---|
| 11, 14, 19 | acm | dato-ausente |
| 17, 20 | cli | dato-ausente |
| 21, 28 | acm | enum-invalido |
| 23, 26 | cli | enum-invalido |

Reparto acm 5 / cli 4; por celda: dato-ausente acm 3/5 retenidos
(runs 11, 14, 19), dato-ausente cli 2/5 (runs 17, 20), enum-invalido acm 2/5
(runs 21, 28), enum-invalido cli 2/5 (runs 23, 26). Esto vulnera la condición aprobada
de EXPERIMENT §1: «Si el modelo no dispone del dato necesario, ambas
variantes reciben la misma respuesta de un fixture de usuario». Los éxitos
8/15 en ambas variantes están contaminados por este defecto y no miden el
comportamiento que la puerta exige.

## Salvaguardas ystash verificaciones realizadas antes de descartar

- El oráculo de base es fiel: los fallos no atribuibles al fixture son
  reenvíos del propio modelo (`duplicate_count=2`, runs 07 y 12), bodies
  inventados (run 18) o la frase del fixture pegada como body (run 15) — la
  base no miente.
- **Cero errores de contrato en 30 ejecuciones, verificado dos veces**: el
  contador en vivo tenía un bug de escapado (json.dumps sobre la salida ya
  texto), corregido y recalculado desde `opencode.db` (auditoria.py): los
  modelos autocorrigieron el enum antes de llamar (el bloque de prefijada les
  dio los valores permitidos) y no llamaron nunca con `confimed`.
- **Cero efectos indebidos/integridad** en ambas variantes: ninguna sesión
  ajena, sin checkpoints ni preferences, `agent_id` correcto.
- Tokens (con la misma reserva de validez): acm 37.487 vs cli 53.715
  (ratio 0,698); reparación: media 2,3 llamadas acm vs 2,1 cli.

## Correcciones del instrumento aplicadas al análisis (no al arnés)

1. `auditoria.py`: reconstrucción completa de la I/O por sesión desde
   `opencode.db` del LAB; recalcula errores/notas sin el bug de escapado.
2. `volcado.py`: preservación de brutos completada — la uña del driver solo
   guardaba la última ronda; herramientas.jsonl y rondas.json por ejecución
   contienen ahora el flujo íntegro.
3. `registro.jsonl`: `efectos_indebidos` reservado a su sentido del protocolo;
   los motivos de fracaso del oráculo viven en `motivos_fracaso`.
4. `incidencias-previas/`: los tres intentos de arranque previos a la medición
   (shim vite-plus / runtime Node / workspace), conservados y etiquetados.

## Decisión pendiente del PO

La campaña violó una condición experimental aprobada en 9 ejecuciones. El
protocolo no contempla repetición selectiva y la puerta exige medición válida.
Opciones:

- **A (recomendada): repetir la campaña completa** con el detector del fixture
  reparado (regla: entregar fixture si el último texto pide datos, sin heurística
  de éxito), conservando ambos juegos de datos como evidencia. Coste ≈ 1 h.
  No es selección por resultados: es re-ejecución íntegra tras un defecto
  demostrado del instrumento.
- **B: declarar INCONCLUSO** con esta evidencia (PRD §7 orden 2: medición no
  comparable) y no repetir.

Datos: `experiment/runs/` (registro.jsonl, auditoria.json, brutos por
ejecución, config.json, metricas.json con la reserva de validez).
