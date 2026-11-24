# AIF — Protocolo del experimento previo a Specification

| Metadato | Valor |
|---|---|
| ID | AIF-EXP |
| Versión | 0.1.6 |
| Fecha | 2026-11-24 |
| Estado | Aprobado por el PO el 2026-11-24 (PRD D-08). Ejecución fuera de esta sesión, en la máquina de desarrollo; antes de medir debe verificarse la telemetría del modelo ([EXPERIMENT-RUNBOOK.md](EXPERIMENT-RUNBOOK.md) §1) |
| Origen | [PRD.md](PRD.md) §7 (puerta GO/NO-GO) |

Este documento define el procedimiento. La puerta y sus criterios viven en el PRD; si discrepan, prevalece el PRD. Los umbrales se fijan antes de ejecutar.

## 1. Diseño

Sin crear motor nuevo: usar ACM existente en una copia desechable y una CLI JSON de comparación que comparta el mismo ejecutor, validadores, campos y permisos. El baseline ofrece ayuda/esquema accesible y errores JSON accionables, pero no formulario contextual con valores preservados. No degradar artificialmente el baseline. Documentar el adaptador mínimo utilizado y mantener las garantías de negocio equivalentes.

Un modelo fijado (identificador y configuración registrados), tres escenarios (entrada completa, dato obligatorio ausente, enum inválido), cinco pares por escenario: 30 ejecuciones en total. Cada par recibe la misma tarea y estado inicial equivalente; alternar el orden de las variantes y usar sesiones independientes. Límite de cuatro invocaciones de herramienta y 120 s por ejecución.

Si el modelo no dispone del dato necesario, ambas variantes reciben la misma respuesta de un fixture de usuario. Una necesidad legítima de datos no cuenta como error. No ocultar abandonos ni timeouts.

## 2. Qué se registra

Tokens de entrada y salida, llamadas (incluidas las de descubrimiento), errores de contrato, éxito verificable en base de datos, tiempo total y efectos indebidos.

| Métrica | Línea base | Papel | Umbral provisional |
|---|---|---|---|
| Éxito | 15 ejecuciones CLI JSON, pendiente | Puerta | Al menos igual y 14/15 o más |
| Tokens | Total CLI JSON, pendiente; dato histórico no comparable: 12.700 | Puerta | AIF ≤ 125 % del baseline |
| Efectos no autorizados o inválidos | Inspección de ambos estados | Puerta | Cero en AIF |
| Beneficio: errores de contrato | Recuento CLI JSON, pendiente | Puerta (uno de los dos) | Al menos 30 % menos |
| Beneficio: invocaciones en casos de reparación | Media CLI JSON, pendiente | Puerta (uno de los dos) | Al menos 20 % menos |
| Latencia (mediana; p95 descriptivo) | Mediana CLI JSON, pendiente | Descriptiva | Referencia: AIF ≤ 125 % del baseline |
| Camino directo | Cinco casos completos por variante | Descriptiva | Referencia: cinco de cinco en una llamada de operación |

## 3. Interpretación

- **Orden de evaluación:** el de PRD §7, que prevalece. Integridad, completitud, éxito y tokens, y por último beneficio.
- **Sin margen de mejora medible:** si el baseline no presenta errores de contrato ni casos de reparación, y éxito y tokens cumplen, el resultado es INCONCLUSO. No se declara GO por equivalencia ni NO-GO por falta de margen.
- **Éxito o tokens fuera de umbral:** con medición completa y comparable el resultado es NO-GO, haya o no beneficio medible. El PO puede decidir reformular el diseño y repetir el filtro.
- **Muestra:** con cinco pares por escenario los umbrales son criterios de producto, no significación estadística. Es un filtro de inversión, no una prueba de ventaja universal.
- **p95:** descriptivo, no concluyente con esta muestra.

## 4. Decisiones del PO y propuestas pendientes

**Aprobada (2026-11-24): primera llamada prefijada.** En los escenarios «dato ausente» y «enum inválido», el operador ejecuta una primera llamada con el dato ausente o inválido, idéntica en contenido lógico, orden y exposición al modelo en ambas variantes. Cuenta en tokens y en el límite de cuatro invocaciones, y se registra como intervención: el resultado no se presenta como comportamiento libre del modelo. Garantiza que los dos escenarios de reparación (10 ejecuciones por variante) sean casos de reparación medibles.

**Aprobadas (2026-11-24):**

- **Error de contrato:** todo rechazo por forma, campo, tipo o enum, contado una vez; no cuentan la petición legítima de datos ni la denegación.
- **Éxito en «dato ausente»:** exige completar la escritura tras la respuesta del fixture.
- **Precedencia de la puerta:** integridad → completitud → éxito y tokens → beneficio, tal como figura en PRD §7.
- **Modelo:** proveedor `nan` mediante OpenCode, modelo `nan/deepseek-v4-flash` con variante `medium` (fuente: [nan-opencode-delegation](https://github.com/fred66sp/nan-opencode-delegation)). Se usa un único agente, sin el orquestador ni los subagentes del paquete, para no añadir tokens ajenos a la comparación. Pendiente de verificar antes de medir: que el proveedor ofrezca una revisión inmutable, la temperatura configurable y contadores de tokens completos por respuesta; si no, la medición queda incompleta (PRD §7).

- **Aislamiento, registro y tokens mínimos:** directorio temporal con `ACM_HOME` y `XDG_CONFIG_HOME` propios y copia de la base limpia por ejecución; registro con variante, escenario, par, orden, tokens, llamadas, errores de contrato, éxito verificado y tiempo; tokens como entrada más salida, informando de la caché solo si el proveedor la da. Sustituye a las versiones más pesadas de [EXPERIMENT-PLAN.md](EXPERIMENT-PLAN.md) §§2, 5 y 6.

**Propuestas sin aprobar** (origen: [EXPERIMENT-PLAN.md](EXPERIMENT-PLAN.md)):

- Ejecución en la máquina con WSL: no se hace desde esta sesión. Aquí solo se prepara y documenta; quien ejecute allí sigue el runbook.

## 5. Plazo

Dos jornadas desde el inicio de la ejecución en la máquina de desarrollo (la preparación y la documentación se hacen aparte; guía operativa en [EXPERIMENT-RUNBOOK.md](EXPERIMENT-RUNBOOK.md)). Publicar datos y decisión dentro de la segunda jornada o declarar bloqueo.
