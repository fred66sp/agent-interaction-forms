# AIF — Runbook de ejecución del experimento

| Metadato | Valor |
|---|---|
| ID | AIF-EXP-RUN |
| Versión | 0.1.0 |
| Fecha | 2026-11-24 |
| Destinatario | Quien ejecute el experimento en la máquina de desarrollo (Linux/WSL) |
| Autoridad | [PRD.md](PRD.md) §7 y [EXPERIMENT.md](EXPERIMENT.md) v0.1.6, aprobados por el PO. Si discrepan con este runbook, prevalecen ellos |

Este runbook solo operacionaliza lo aprobado. No añade criterios. Lo que figura como **[Sin verificar]** procede de [EXPERIMENT-PLAN.md](EXPERIMENT-PLAN.md) (lectura de código sin ejecución) y debe comprobarse en la máquina de desarrollo antes de fiarse de ello.

## 0. Reglas

- Trabajar solo en copias. No tocar la instalación de ACM, su base real ni `~/.agent-memory`.
- Sin commits y sin operaciones remotas no autorizadas.
- Un único agente OpenCode con el proveedor `nan`; sin orquestador ni subagentes de nan-opencode-delegation.
- Ante cualquier comprobación previa fallida (§1), detenerse y registrar el resultado como medición incompleta; no improvisar.

## 1. Comprobaciones previas (parar si falla alguna)

1. **Modelo.** `opencode models | grep 'nan/'` lista `nan/deepseek-v4-flash`. Fijar variante `medium`. Registrar versión de OpenCode y del cliente.
2. **Telemetría.** Una sola llamada de prueba con el modelo. Confirmar que OpenCode registra tokens de entrada y de salida por respuesta. Si no hay contadores completos, la medición es incompleta (PRD §7): detener y avisar al PO.
3. **Parámetros del modelo.** Anotar si la temperatura es configurable (usar 0 si lo es) y si existe una revisión inmutable del modelo. Registrar lo que no sea configurable.
4. **ACM en Linux/WSL.** [Sin verificar] `acm.py` requiere Python ≥ 3.10, `db-ai` en `PATH`, `bash`, `sqlite3`, `flock`; `init.sh` los comprueba. El binario Windows de `db-ai` llevaba un parche en `sqlite.go` (`dsn()` para rutas `C:`): comprobar si el código fuente Linux lo necesita. Construir `db-ai` desde una copia del código (`go build ./cmd/db-ai`).
5. **Fuente de ACM.** `database-virtual-filesystem/examples/agent-continuity` (adaptador `tools/acm.py`, `init.sh`, `tools/content-hash.py`).

## 2. Montaje (aislamiento mínimo, aprobado)

1. Crear un directorio temporal `LAB` fuera de cualquier repositorio y copiar ahí `db-ai` y `agent-continuity` como archivos ordinarios.
2. Exportar `HOME`, `XDG_CONFIG_HOME`, `ACM_HOME`, `ACM_ALIAS` y `ACM_AGENT_ID` apuntando dentro de `LAB`, con valores propios (no los personales).
3. Ejecutar `bash init.sh` de la copia, sin `--demo`.
4. Guardar la base limpia como plantilla. **Antes de cada ejecución**, copiarla a una base nueva, abrir una sesión con el mismo identificador y objetivo, y partir de cero observaciones.
5. Comprobar una vez que el montaje resuelve a la base de `LAB` y no a la real.

## 3. Variantes

- **ACM:** `python3 tools/acm.py learning ...` con formulario contextual y valores preservados.
- **CLI JSON (baseline):** adaptador mínimo sobre la misma copia, solo para el evento `learning`. [Sin verificar] Importa `EVENT_FORMS` y `build_form` para la misma validación, acepta un objeto JSON completo por stdin, ofrece ayuda y esquema derivados de `EVENT_FORMS`, y traduce `invalid` a un error JSON con campo, causa y valores permitidos. En entrada válida extrae los valores normalizados y llama a `submit_event` con la misma sesión, actor y `db-ai`. No devuelve formulario, valores recibidos ni metadatos contextuales. No degradarlo: mismos campos, restricciones, defaults y permisos.
- **Antes de medir**, probar sin modelo, en ambas variantes, las rutas completa, inválida, denegada, duplicada y de fallo de dominio. Si las garantías aplicables no son equivalentes, detenerse y pedir decisión al PO.

## 4. Escenarios

Evento `learning` ([Sin verificar]: `title` y `body` obligatorios; `topic` opcional; `confidence` ∈ `confirmed|inference|hypothesis|proposal|unknown`). 3 escenarios × 5 pares × 2 variantes = 30 ejecuciones. Textos sintéticos, idénticos dentro del par y distintos entre pares. Ejemplo del par base:

| Escenario | Datos | Primera llamada prefijada (aprobada) |
|---|---|---|
| Completa | `title` «Parser en UTC»; `body` «Las fechas sin zona se interpretan como UTC.»; `topic` `parser-fechas`; `confidence` `confirmed` | Ninguna: el modelo actúa libremente |
| Dato ausente | Mismos datos, sin `body` | Llamada sin `body` |
| Enum inválido | Mismos datos, `confidence` = `confimed` | Llamada con `confimed` |

**Primera llamada prefijada.** Idéntica en contenido lógico, orden y exposición para ambas variantes; cuenta en tokens y en el límite de cuatro invocaciones; se registra como intervención. [Sin verificar] Si OpenCode permite sembrar la llamada y su resultado en el contexto de la tarea, hacerlo así; si no, el operador la ejecuta y pega su salida de forma idéntica en ambas variantes.

**Fixture de usuario.** Solo responde si el modelo pregunta: «Las fechas sin zona se interpretan como UTC.» (cuerpo) y `confirmed` (confianza). No ofrecer datos no solicitados ni inventar otros. Registrar cada entrega.

**Límites por ejecución:** 4 invocaciones de herramienta y 120 s, desde el inicio hasta el resultado (incluye espera del fixture y descubrimiento). Sesión de modelo independiente por ejecución, sin caché ni conversación compartidas.

**Orden:** alternar por par dentro de cada escenario: ACM primero en los pares 1, 3 y 5; CLI primero en los pares 2 y 4.

## 5. Oráculo de éxito (base de datos, no `ok:true`)

- Tras entrada ausente o inválida: cero cambios en `sessions`, `observations`, `checkpoints` y `preferences`.
- Tras completar (también en «dato ausente», donde el éxito exige completar la escritura): exactamente una observación nueva con `kind=learning`, sesión y actor previstos, `title`, `body`, `topic_key`, `confidence`, `status=active` y `duplicate_count=1`, y ningún otro cambio de dominio. Comparar contenido y estado, no UUID ni sellos de tiempo.
- Abandono, límite de llamadas y timeout no son éxito. Conservar la evidencia de todos los intentos.

## 6. Registro (mínimo, aprobado)

Una línea por ejecución (JSONL o CSV) con: `variante`, `escenario`, `par`, `orden`, `tokens_entrada`, `tokens_salida`, `llamadas` (visibles del agente, incluidas ayuda y esquema; los subprocesos internos de `db-ai` no cuentan), `errores_contrato`, `exito_verificado_bd`, `duracion_ms`, `efectos_indebidos`, `abandono`, `timeout`, `estado_medicion` (`completa|incompleta|no_comparable`). Añadir, una vez por experimento, modelo, variante, versiones y parámetros fijados.

**Error de contrato** (aprobado): todo rechazo por forma, campo, tipo o enum, contado una vez. No cuentan la petición legítima de datos, la denegación ni la avería de infraestructura (estas se anotan aparte).

## 7. Cálculo y decisión

- **Éxito:** recuento verificado sobre 15 por variante.
- **Tokens:** suma de entrada más salida de todas las ejecuciones, sin descartar fallos; informar de la caché solo si el proveedor la da.
- **Errores de contrato:** reducción = `100 × (1 − errores_AIF / errores_CLI)`; no calculable si el denominador es 0.
- **Invocaciones en reparación:** la población la fijan los escenarios aprobados, no una selección posterior por resultados: los diez casos de reparación prefijados por variante (cinco de dato ausente y cinco de enum inválido; EXPERIMENT §4), incluidas las ejecuciones fallidas, abandonadas o agotadas conforme al protocolo; media de llamadas por variante, reducción análoga e informe de tamaños y composición. Una petición legítima de datos no es error de contrato, pero pertenece al escenario de reparación correspondiente.
- **Descriptivos:** mediana de duración (p95 solo descriptivo) y camino directo (5 de 5 en una llamada de operación).
- **Decisión:** aplicar el orden de PRD §7: integridad → medición completa → éxito y tokens → beneficio → GO. Resultados posibles: GO, NO-GO, INCONCLUSO. Decide el PO.

## 8. Entregable

Datos brutos, configuración y un informe breve con el resultado de cada criterio, la decisión resultante, lo que no se pudo comprobar y las desviaciones respecto a este runbook. Plazo: dos jornadas desde el inicio de la ejecución en la máquina de desarrollo.
