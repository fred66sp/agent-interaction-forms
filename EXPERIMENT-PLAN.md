# AIF — Plan de preparación del experimento

**Estado: superado.** Este plan es histórico: sus decisiones aprobadas se recogen en [EXPERIMENT.md](EXPERIMENT.md) §4 y la guía operativa es [EXPERIMENT-RUNBOOK.md](EXPERIMENT-RUNBOOK.md). Las versiones de aislamiento, registro y tokens de este plan fueron sustituidas por otras más ligeras. Estado original: propuesta bloqueada; no autoriza mediciones. La puerta canónica es [PRD.md](PRD.md) §7 y el procedimiento de partida es [EXPERIMENT.md](EXPERIMENT.md). La aprobación previa del PO es obligatoria.

**Etiquetas:** **[Hecho]** = contenido observado en una fuente o comprobación, no necesariamente comportamiento ejecutado; **[Inferencia]** = conclusión o propuesta no aprobada; **[Desconocido]** = falta evidencia o decisión. Las instrucciones propuestas son **[Inferencia]**. Las rutas de ACM son relativas al árbol `database-virtual-filesystem-main`, facilitado como fuente.

## Resumen en diez líneas

1. **[Hecho]** Se compara ACM con una CLI JSON competente que comparte ejecutor, validadores, campos y permisos; solo ACM ofrece formulario contextual y valores preservados (EXPERIMENT §1).
2. **[Hecho]** Se prevén un modelo fijado, tres escenarios, cinco pares por escenario y sesiones independientes: 15 ejecuciones por variante (EXPERIMENT §1).
3. **[Hecho]** Cada par comparte tarea y estado inicial equivalente; se alterna el orden y se limita cada ejecución a cuatro invocaciones de herramienta y 120 segundos (EXPERIMENT §1).
4. **[Hecho]** Si falta información al modelo, ambas variantes reciben la misma respuesta de un fixture de usuario; pedir un dato necesario no es un error (EXPERIMENT §1).
5. **[Hecho]** Para GO, AIF debe lograr al menos el éxito del baseline y al menos 14 éxitos de 15 (PRD §7).
6. **[Hecho]** Para GO, los tokens totales de AIF deben ser como máximo el 125 % de los del baseline (PRD §7).
7. **[Hecho]** Para GO, AIF debe tener cero efectos no autorizados o inválidos (PRD §7).
8. **[Hecho]** También debe reducir al menos un 30 % los errores de contrato o un 20 % las invocaciones en reparación (PRD §7).
9. **[Hecho]** Medición incompleta, no comparable o sin margen de beneficio en el baseline lleva a INCONCLUSO; un defecto de integridad es siempre NO-GO (PRD §7).
10. **[Hecho]** Sin defecto de integridad, una medición completa y comparable que incumpla un criterio es NO-GO; latencia y camino directo son descriptivos y el PO debe aprobar antes de medir (PRD §§7–8).

### Ambigüedades y diferencias que requieren decisión

- **[Hecho]** EXPERIMENT §3 menciona explícitamente NO-GO cuando hay beneficio pero fallan éxito o tokens; PRD §7 incluye cualquier criterio incumplido con medición completa y comparable y declara que un defecto de integridad siempre es NO-GO. **[Inferencia]** El protocolo no describe de forma igualmente explícita todos esos casos; prevalece PRD, sin modificar EXPERIMENT.
- **[Desconocido]** Ninguna fuente define qué se cuenta como error de contrato, qué ejecuciones integran el denominador de reparación, cómo tratar denominadores cero por separado, ni el tratamiento de un fallo del baseline en el criterio de efectos. Debe fijarse antes de medir.
- **[Desconocido]** No se especifica cómo provocar de forma comparable las condiciones «dato ausente» y «enum inválido» sin enseñar al modelo la corrección, ni si una llamada prefijada por el operador cuenta dentro de las cuatro. Una orden que revele el enum válido sesgaría el resultado.
- **[Hecho]** RESEARCH §10 describe el escenario de dato ausente como «formulario, sin mutación de dominio»; EXPERIMENT §1 también contempla recibir una respuesta de usuario. **[Desconocido]** No queda fijado si el éxito final exige completar la escritura después de ese estado intermedio.
- **[Desconocido]** No se especifican proveedor, modelo disponible, configuración, telemetría de tokens, ni precedencia entre «baseline sin margen» y otros incumplimientos sin defecto de integridad. PRD solo explicita la prioridad absoluta del defecto de integridad.

## 1. Entorno y dependencias

- **[Hecho]** `examples/agent-continuity/tools/acm.py` usa Python, invoca `db-ai` por `PATH`, lee `ACM_ALIAS` y `ACM_AGENT_ID`, y utiliza `tools/content-hash.py`. Sus anotaciones usan `dict | None`, que requieren Python 3.10 o posterior sin evaluación diferida.
- **[Hecho]** `examples/agent-continuity/init.sh` requiere `bash`, `sqlite3`, `python3`, `flock` y `db-ai`; prepara SQLite, política de montaje y configuración bajo `XDG_CONFIG_HOME`, o bajo `HOME` si no se fija. ACM declara el contrato db-ai v0.2.0 en su README.
- **[Hecho]** El lanzador Windows `acm.cmd` llama al intérprete y al adaptador mediante rutas Windows; no es un lanzador Linux. `db-ai.exe` está disponible en el PATH Windows. `go.mod` declara Go 1.27 y `modernc.org/sqlite`; el README de db-ai documenta `go build ./cmd/db-ai`.
- **[Hecho]** En la máquina donde se preparó este documento, `wsl --status` terminó con código 1 e informó de que WSL no está instalado. El usuario indica que WSL sí está disponible en la otra máquina prevista para ejecutar el plan. No se ha comprobado esa máquina ni se ha ejecutado allí ACM o db-ai.
- **[Desconocido]** No se ha identificado el cambio aplicado al binario Windows ni comprobado su equivalencia con el código fuente. No se sabe si Linux necesita un parche equivalente ni si dispone de las dependencias o versiones indicadas.
- **[Inferencia]** En un Linux/WSL habilitado, construir db-ai desde una copia del código con las dependencias ya disponibles, fuera de la instalación original, y usar `python3 tools/acm.py`; comparar el cambio Windows con el código fuente y ejecutar pruebas de contrato aisladas antes de decidir si hace falta un parche. No descargar nada ni modificar la instalación como parte de esta preparación.

## 2. Aislamiento propuesto

**[Inferencia]** En un entorno Linux autorizado, crear una raíz nueva y privada de experimento fuera de cualquier repositorio y de toda memoria real; asignarla a `LAB`. Copiar el código fuente necesario de db-ai y `examples/agent-continuity` como archivos ordinarios, sin enlaces simbólicos ni `.agent-memory`, configuraciones personales, datos reales ni binarios preexistentes. Verificar rutas canónicas y ausencia de enlaces antes de ejecutar cualquier comando; ante una ruta fuera de `LAB`, detenerse.

**[Inferencia]** Construir el binario en `LAB` desde la copia, sin red; fijar `HOME`, `XDG_CONFIG_HOME`, `ACM_HOME`, `ACM_ALIAS`, `ACM_AGENT_ID`, `PATH`, directorios temporales y cachés de compilación dentro de `LAB`. Usar un entorno explícito y mínimo, sin heredar variables de sesión ni configuración de db-ai. Ejecutar `bash init.sh` de la copia, sin `--demo`, con `ACM_HOME` y `XDG_CONFIG_HOME` nuevos. No usar el alias ni la configuración personal aunque tengan el mismo nombre.

**[Hecho]** `init.sh` crea `.agent-memory/memory.db` bajo `ACM_HOME`, escribe la política y el montaje en `XDG_CONFIG_HOME/db-ai/config.json`; `acm.py` guarda los marcadores `acm-forms` junto a la base obtenida de `db-ai mounts`. Las políticas de db-ai no sustituyen permisos de sistema operativo (README de ACM, límites declarados).

**[Inferencia]** Antes de cualquier prueba, comprobar que el único montaje resuelve a la base de `LAB`, que todos los archivos de configuración, marcadores, auditoría y temporales están dentro de `LAB`, que no se puede acceder a la memoria real mediante permisos o aislamiento de proceso, y que la política conserva lectura y escritura originales. Detenerse si no puede demostrarse. Registrar versiones y hashes de las copias y del binario, sin secretos.

**[Inferencia]** Crear una plantilla limpia y una base y configuración independientes por ejecución; clonar la plantilla con SQLite cerrado, incluidos los metadatos equivalentes, y usar marcadores vacíos separados. Preparar en cada copia una sesión abierta con el mismo identificador y objetivo; las conversaciones del modelo siguen siendo independientes. Comprobar equivalencia semántica antes de cada par y diferencias posteriores con SQLite en modo solo lectura. No copiar nunca la base real.

## 3. Baseline y equivalencia

- **[Hecho]** `EVENT_FORMS` define campos y restricciones; `build_form` valida y construye la representación contextual; `submit_event` despacha a `record_observation` y demás funciones de dominio. `record_observation` comprueba sesión, valida texto, calcula hash y conserva deduplicación. `resolve_event` y `cmd_submit` añaden respectivamente formularios y marcadores de reenvío.
- **[Inferencia]** Implementar en la copia un adaptador pequeño para `learning` que acepte un objeto JSON completo por stdin y ofrezca ayuda/esquema explícitos. Importar `EVENT_FORMS` y llamar `build_form` para la misma validación; descartar su formulario, nunca devolver valores recibidos ni metadatos contextuales, y traducir `invalid` a un error JSON con campo, causa y valores permitidos. En entrada válida, extraer los valores normalizados y llamar a `submit_event`, con la misma sesión, actor y `db-ai`.
- **[Inferencia]** El baseline debe aceptar todos los campos, restricciones, valores por defecto y permisos de `learning`, y distinguir fallos de validación, permisos y ejecución. La ayuda se deriva de `EVENT_FORMS`; no duplicar reglas ni quitar mensajes útiles. Reenviar requiere el objeto completo, no estado almacenado ni formulario preservado.
- **[Desconocido]** La equivalencia exacta de reenvíos es delicada: el camino `submit` de ACM usa marcadores `form_id` y el camino directo no los usa. No declarar paridad de idempotencia sin pruebas; no eliminar ni debilitar garantías existentes. Antes de medir, probar con fixtures sin modelo las rutas completas, inválidas, denegadas, duplicadas y de fallo de dominio para ambas variantes; si no son equivalentes en garantías aplicables, bloquear y pedir decisión.

## 4. Escenarios, fixture y oráculo

**[Inferencia]** Usar `learning`, cuyo esquema exige `title` y `body`, permite `topic` y limita `confidence` a `confirmed`, `inference`, `hypothesis`, `proposal` o `unknown` (`acm.py`, `EVENT_FORMS`). Cada ejecución parte de una base limpia con una sesión abierta y sin observaciones con el hash del contenido esperado. Usar textos sintéticos, idénticos dentro del par y diferentes entre pares.

| Escenario | Datos sintéticos propuestos | Condición a observar |
|---|---|---|
| **[Inferencia]** Completa | `title`: «Parser en UTC»; `body`: «Las fechas sin zona se interpretan como UTC.»; `topic`: `parser-fechas`; `confidence`: `confirmed`. | Una operación válida crea exactamente la observación esperada. |
| **[Inferencia]** Obligatorio ausente | Mismos datos, salvo que el usuario inicial no aporta `body`; el fixture responde «Las fechas sin zona se interpretan como UTC.» solo si se pregunta por el contenido. | La falta no debe producir escritura; tras obtener el dato, se evalúa la escritura final si el PO aprueba esa interpretación. |
| **[Inferencia]** Enum inválido | Mismos datos, salvo `confidence`: `confimed`, valor inválido; el fixture responde `confirmed` únicamente si se pregunta qué confianza registrar. | La entrada inválida no escribe; la corrección válida produce una observación esperada. |

**[Desconocido]** Un modelo podría pedir el dato antes de llamar a la herramienta o corregir por sí mismo `confimed`. No se debe forzar artificialmente un error ni excluir silenciosamente esa ejecución. Preacordar con el PO si el escenario prueba la interacción libre o una primera llamada predeterminada, y cómo contar dicha llamada y sus tokens. Si se impone una llamada, aplicar exactamente el mismo contenido lógico, orden y exposición al modelo a ambas variantes; registrar la intervención y no presentar el resultado como comportamiento libre.

**[Inferencia]** Para cada par, preparar una tabla de fixture con pregunta semántica, respuesta literal y condición de entrega; no responder con datos no solicitados, no inventar otros y registrar cada entrega. La solicitud legítima de datos no es un error de contrato. Si no existe respuesta aplicable, registrar abandono o bloqueo, no fabricar un valor.

**[Inferencia]** Tomar una instantánea lógica antes y después de cada llamada. Para una entrada ausente o inválida, exigir cero cambios de dominio en `sessions`, `observations`, `checkpoints` y `preferences`; inspeccionar también marcadores y auditoría como efectos auxiliares separados. Tras completar, exigir exactamente una observación nueva `kind=learning`, sesión y actor previstos, `title`, `body`, `topic_key`, `confidence`, `status=active` y `duplicate_count=1`; verificar también que no cambian otras filas de dominio. Comparar contenido y estado, no UUID ni sellos de tiempo. Un `ok:true` sin esa comprobación no es éxito. Conservar evidencia de todos los intentos, incluso tras error o timeout.

## 5. Registro y cálculo propuestos

**[Inferencia]** Una fila por ejecución en JSONL, con un identificador sin datos personales y un registro de eventos enlazado. Guardar como mínimo:

```json
{
  "variante": "acm|cli_json",
  "escenario": "completa|ausente|enum_invalido",
  "par": "identificador",
  "orden": 1,
  "modelo_id": "identificador exacto",
  "configuracion_modelo": {},
  "tokens_entrada": 0,
  "tokens_salida": 0,
  "tokens_desglose": [],
  "llamadas_descubrimiento": 0,
  "llamadas_operacion": 0,
  "llamadas_totales": 0,
  "errores_contrato": [],
  "necesidades_dato": [],
  "exito_verificado_bd": false,
  "evidencia_bd": "referencia",
  "duracion_total_ms": 0,
  "efectos_indebidos": [],
  "abandono": false,
  "timeout": false,
  "estado_medicion": "completa|incompleta|no_comparable"
}
```

- **[Inferencia]** Contar cada llamada visible del agente, incluso ayuda, esquema y descubrimiento, y cada llamada de operación o reparación; los subprocesos internos de db-ai no son llamadas del agente, pero pueden registrarse aparte. Registrar la petición, respuesta, código, tiempo y categoría por evento, sin secretos. Los cuatro usos y los 120 segundos se aplican desde el inicio de la ejecución hasta su resultado; incluir espera del fixture y descubrimiento. El montaje y la verificación independiente de base quedan fuera del tiempo de tarea.
- **[Inferencia]** Propuesta para aprobación: contar cada respuesta que rechaza una petición por forma, campo, tipo o enum como un error de contrato; no contar petición legítima de datos, denegación de permiso ni avería de infraestructura como tal. Registrar estas últimas por separado. No contar dos veces el mismo rechazo por tener varios campos.
- **[Inferencia]** Propuesta para aprobación: identificar como «caso de reparación» cada tarea que recibió un rechazo de contrato y calcular la media de llamadas totales visibles en esas tareas para cada variante; informar tamaños y composición de los conjuntos y comparar también los pares correspondientes. Si los conjuntos difieren, no ocultarlo: decidir comparabilidad antes de aplicar el umbral. Con denominador cero para una reducción, marcarla no calculable; no inventar porcentajes.
- **[Inferencia]** Sumar tokens por variante incluyendo todas las ejecuciones, sin eliminar fallos; calcular `100 × (1 − errores_AIF / errores_CLI)` y `100 × (1 − media_reparación_AIF / media_reparación_CLI)` cuando el denominador exista. Calcular éxito como recuento verificado sobre 15 y tokens como suma de entrada más salida. Mantener cada componente y datos brutos. Mediana y p95 de duración son descriptivos; registrar aparte el número de operaciones en el camino completo.
- **[Inferencia]** Alternar el orden por par dentro de cada escenario: ACM primero en pares 1, 3 y 5; CLI primero en pares 2 y 4. Registrar el desequilibrio inevitable de cinco pares. Aislar sesión y estado entre variantes; no reutilizar conversación ni caché compartida.
- **[Inferencia]** Abandono, límite de llamadas o timeout no se ocultan y no son éxito; registrar causa y estado de la base. Si falta telemetría, estado u otra evidencia indispensable, marcar medición incompleta y aplicar PRD §7, sin convertirla en fracaso comparable. Un efecto indebido se comunica inmediatamente y se clasifica según el criterio de integridad del PRD.

## 6. Modelo y contabilidad

**[Inferencia]** Candidato a fijar: `openai/gpt-6.1-terra-fast`, identificador del modelo de esta sesión, sujeto a confirmar que el proveedor de evaluación lo ofrece con una revisión inmutable y telemetría completa. No se ha hecho ninguna llamada de medición ni se ha comprobado disponibilidad mediante API.

**[Inferencia]** Fijar antes de medir proveedor, identificador y revisión exactos, misma instrucción de sistema, límite de salida, temperatura 0 si está soportada, configuración de razonamiento y herramientas, versión del cliente y ausencia o política idéntica de caché; registrar cualquier valor no configurable. Una revisión flotante o una configuración no comparable bloquea la medición.

**[Inferencia]** Usar las cifras de uso del proveedor por cada respuesta del modelo y sumar todos los turnos de cada ejecución: instrucciones, tarea, definiciones de herramientas, descubrimiento, salidas de herramientas, fixture y respuestas finales. Separar entrada, salida, entrada cacheada y razonamiento cuando el proveedor lo informe; definir previamente si el proveedor incluye el razonamiento en salida y no sumarlo dos veces. Registrar también contadores brutos y parámetros de caché. Los tokens históricos de RESEARCH §3.1 no son baseline comparable. Si faltan contadores o se facturan categorías de forma no comparable, no sustituirlos por estimaciones sin aprobación; la medición queda incompleta.

## 7. Bloqueos, riesgos y preguntas para el PO

1. **[Desconocido] Bloqueo:** aunque el usuario indica que WSL está disponible en la máquina prevista, no se ha comprobado allí la compatibilidad de ACM, la versión efectiva de db-ai ni la necesidad de parche. Verificar las dependencias en esa máquina y contrastar el parche Windows con el código antes de medir.
2. **[Hecho] Bloqueo:** D-08 exige aprobación previa; la aprobación aún no consta en las fuentes. Aprobar el protocolo, umbrales y decisiones pendientes antes de medir.
3. **[Desconocido] Bloqueo:** disponibilidad real del modelo candidato, revisión fija, configuración y contadores completos. Sin ello no se puede garantizar comparabilidad.
4. **[Desconocido] Bloqueo:** baseline y aislamiento aún no están implementados ni probados; garantías de reenvío y paridad de permisos pendientes de verificación sin modelo.
5. **[Inferencia] Riesgo:** el modelo puede evitar naturalmente los errores del escenario; imponerlos cambia la pregunta experimental. Fijar el diseño sin ocultar la intervención.
6. **[Inferencia] Riesgo:** una muestra pequeña, conjuntos de reparación distintos y orden 3/2 por escenario pueden sesgar resultados; no extrapolar ni presentar p95 como concluyente.
7. **[Inferencia] Riesgo:** XDG aislado por sí solo no impide acceso directo a la memoria real; exigir además aislamiento de permisos y verificación de rutas y efectos.

**Preguntas para aprobación:** **[Desconocido]** ¿Aprueba el PO la definición de error, el denominador de reparación y la regla para denominadores cero? ¿El escenario ausente termina al devolver el formulario o tras completar con el fixture? ¿Se permite una primera llamada prefijada en ambos escenarios de reparación y cómo se contabiliza? ¿Qué resultado corresponde si no hay margen en el baseline y también falla otro criterio sin defecto de integridad? ¿Qué modelo, revisión y proveedor disponibles se autorizan si el candidato no ofrece telemetría comparable?

**[Hecho] Límite de esta preparación:** no se han ejecutado las 30 tareas, llamadas a modelos, pruebas de ACM, creación de bases ni compilación. La evidencia de código no demuestra ejecución en WSL ni seguridad del aislamiento propuesto.
