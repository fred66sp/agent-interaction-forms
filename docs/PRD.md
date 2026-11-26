# AIF — Agent Interaction Forms
## Product Requirements Document

| Metadato | Valor |
|---|---|
| ID | AIF-PRD |
| Versión | 0.2.0 |
| Fecha | 2026-11-26 |
| Estado de preparación | PARCIAL |
| Fuentes | [RESEARCH.md](RESEARCH.md), v0.1.2 · [PRODUCT.md](PRODUCT.md), v0.2 (visión aprobada por el PO) · [SPECIFICATION.md](SPECIFICATION.md), v0.2.0 (aprobada por el PO) |
| Estado de aprobación | Pendiente de revisión del PO |

## 1. Visión y problema

AIF dota a las herramientas de una interfaz autodescriptiva, determinista y orientada a agentes; su núcleo es un contrato abierto, ligero e independiente del lenguaje para la interacción determinista entre agentes y herramientas ([PRODUCT.md](PRODUCT.md) §1). Una operación completa se ejecuta directamente; una incompleta devuelve los datos que faltan; una inválida permite corregir sin reconstruir los datos ya conocidos y sin efectos de dominio, aunque la conservación de esos datos no certifica su validez. El contrato es único —v0.2— y sin estado: cada petición contiene todo lo que el ejecutor necesita y el ejecutor revalida siempre.

El problema es la fricción de descubrir sintaxis, interpretar errores y reconstruir argumentos. ACM demuestra viabilidad, no superioridad: su README informa de 22.900 tokens AIF frente a 12.700 convencional (+80,3 %) y 106.000 db-ai directo, con n=1 por variante. No existe todavía una línea base comparable frente a una CLI JSON competente.

### 1.1 Visión aprobada y jerarquía documental

La visión de producto está aprobada y vive en [PRODUCT.md](PRODUCT.md) v0.2: qué es AIF, el problema, las dos modalidades de uso, los componentes y el estado actual. Este PRD no la repite; fija los requisitos verificables que la materializan y la referencia.

Jerarquía de fuentes: PRODUCT.md es la visión aprobada; este PRD, los requisitos; [SPECIFICATION.md](SPECIFICATION.md), el contrato normativo. Si la visión y un requisito discrepan, la discrepancia se eleva al PO antes de modificarse. El reparto entre base implementada, visión completa y capacidades pendientes de diseño está en §3.

## 2. Usuarios y objetivos

Personas hipotéticas, no validadas mediante entrevistas: agente consumidor que necesita completar operaciones; desarrollador que busca una interfaz reutilizable; operador que necesita preservar las reglas de negocio. La experiencia con ACM motiva estas hipótesis, pero no acredita demanda general.

- OBJ-01: reducir errores de contrato y facilitar su recuperación.
- OBJ-02: mantener una sola invocación cuando los datos bastan.
- OBJ-03: demostrar utilidad frente a una CLI JSON bien diseñada antes de construir infraestructura.

## 3. Alcance implementado, base técnica y visión de producto

Este apartado separa tres cosas que no deben confundirse: lo ya implementado y verificado (§3.1), la visión del producto completo aprobada en [PRODUCT.md](PRODUCT.md) (§3.2) y lo que sigue pendiente de diseño (§3.3). Las tres conviven; §3.2 y §3.3 no amplían el alcance implementado ni los requisitos de §§4–6.

### 3.1 Base técnica implementada (contrato v0.2)

El alcance inicial se fijó en tres entregables, condicionados a la puerta de producto: (1) especificación mínima y perfil CLI; (2) implementación de referencia con fixtures de conformidad; (3) PoC aislada con ACM. Go figuraba como candidato, no como decisión cerrada; la implementación entregada está en Go y D-03 conserva su estado en §8. No se modifica ACM de producción. La evolución aprobada del contrato a v0.2 es semántica, no estructural: mismos entregables, mismos ocho resultados y mismos miembros.

Entregado y verificado (detalle en [CHANGELOG.md](../CHANGELOG.md)):

| Pieza | Estado |
|---|---|
| Biblioteca Go `aif` | Implementada: materializa el contrato v0.2 sobre `aif`/`operation`/`input`, con los ocho resultados y su contenido exigible, precedencia de `invalid`, comunicación de todos los errores detectados en un único intercambio y conservación de los datos conocidos, presentes, no nulos, no sensibles y no rechazados, sin que la conservación certifique su validez. |
| Contrato v0.2 | [SPECIFICATION.md](SPECIFICATION.md) v0.2.0, evolución autorizada por el PO; v0.1.1 queda como contrato histórico implementado. |
| Fixtures de conformidad | Serie CJ, independientes del lenguaje, con runner determinista; pasan al 100 %. |
| CLI de demostración | `cmd/aifdemo`, conforme al perfil CLI: códigos de salida 0/1/2/3/4 y diagnósticos separados. |
| PoC con ACM | `poc/acm` con base efímera y copia intacta del adaptador (4/4 escenarios PASS). |

Trazabilidad: la puerta de §7 concluyó **INCONCLUSO** y la Specification v0.1.1 se aprobó después por decisión del PO, registrada en el CHANGELOG. La evolución a v0.2 se aprobó con la implementación en curso. §7 conserva su texto original y este PRD no reescribe la secuencia de §9.

El núcleo no administra sesiones, almacenamiento, claves idempotentes, caducidad, revisiones ni concurrencia. Las garantías existentes de ACM permanecen en ACM; no se eliminan ni se generalizan. El runner, Python como segunda implementación, MCP/HTTP y las garantías de reintento quedan fuera del alcance v0.2, como candidatos sujetos a una nueva decisión de producto. Tampoco entran orquestación, GUI ni servicios residentes.

### 3.2 Visión del producto completo

La visión aprobada ([PRODUCT.md](PRODUCT.md) §§4–5) describe un producto único con dos modalidades complementarias:

- **Integración nativa:** la biblioteca AIF se incorpora al código de la aplicación, sin ejecutable intermediario.
- **Integración externa:** un ejecutable AIF actúa como intermediario; un archivo de definición describe las operaciones y sus restricciones.

Y cuatro componentes como responsabilidades conceptuales de ese único producto, no como proyectos separados: AIF Core (biblioteca reutilizable), AIF CLI (ejecutable), AIF Definition (formato declarativo) y AIF Specification (contrato común). Ambas modalidades comparten núcleo y contrato.

Esta sección describe el destino del producto; no añade requisitos a §§4–6 ni decide su implementación.

### 3.3 Capacidades pendientes de diseño

Dos capacidades de la visión siguen **pendientes de diseño**: **AIF CLI como producto** (capa externa sobre herramientas existentes) y **AIF Definition** (formato declarativo que describe operaciones, información necesaria, restricciones e invocación). Son **conceptuales, no disponibles**: no tienen requisitos, formato ni interfaz decididos, y nada de lo escrito aquí debe leerse como decisión de implementación.

> **Decisión pendiente del PO.** Cómo encaja el alcance implementado con las dos modalidades de la visión, y con D-03 —que mantiene el lenguaje como propuesta pese a la implementación Go existente—, no está decidido. Este PRD lo deja señalado y no lo resuelve.

## 4. Requisitos funcionales

Cada requisito conserva su ID de v0.1.0. Las lagunas de numeración (RF-09, RF-10, RF-14, RF-15) son requisitos diferidos o reducidos; sus IDs no se reutilizan. P01–P08 corresponden a RESEARCH §11.

| ID | Requisito MVP | Origen |
|---|---|---|
| RF-01 | Identificar operación, versión y resultado mediante contrato de máquina. | P01 |
| RF-02 | Ejecutar directamente con datos suficientes y válidos. | P02 |
| RF-03 | Ante datos ausentes, devolver los campos pertinentes con obligatoriedad y restricciones explícitas; `schema` se incluye cuando la regla se expresa fielmente con él, pero no con el tipo ni con los valores admitidos. | P02 |
| RF-04 | Distinguir la ausencia de datos de su invalidez. | P03 |
| RF-05 | Comunicar en un único intercambio todos los errores detectados y conservar los datos conocidos, presentes, no nulos, no sensibles y no rechazados; la conservación no certifica su validez. | P03 |
| RF-06 | Indicar la continuación sin requerir interpretar ayuda general. | P04 |
| RF-07 | Distinguir éxito, necesidad de datos, validación y denegación; informar de fallos de ejecución sin declarar éxito falso. | P03 |
| RF-08 | Revalidar toda entrada y su operación en cada petición, sin estado y con reglas confiables del ejecutor, sin confiar en metadatos enviados por el agente; el dominio valida por completo antes de cualquier efecto. No requiere almacenamiento de formularios. | P06 |
| RF-11 | El dominio decide permisos y ejecuta; AIF no concede autoridad. | P06 |
| RF-12 | Perfil CLI no interactivo con salida JSON y diagnósticos separados. | P07 |
| RF-13 | Fixtures de conformidad ejecutables sin LLM. | P08 |

RF-09/10/15 diferidos (idempotencia, revisión, otros transportes). RF-14 se limita a no prometer garantías de reintento inexistentes; no se implementa un motor de reintentos. Una denegación no invita a corregir para eludir permisos.

La validación puede ser progresiva o exhaustiva: una respuesta reparable comunica todos los errores detectados en ese intercambio, no necesariamente todos los posibles. `schema` no se exige cuando las restricciones disponibles no expresan la regla con fidelidad; rige la regla efectiva que valida el dominio, no la descripción.

## 5. Requisitos no funcionales

| ID | Requisito |
|---|---|
| RNF-01 | Decisiones reproducibles con iguales entradas, reglas y contexto. |
| RNF-02 | Sin LLM, red ni servicio residente obligatorio. |
| RNF-03 | Respuestas compactas y límites de entrada documentados por la especificación. |
| RNF-05 | Sin efectos de dominio por entradas inválidas o denegadas; el dominio valida por completo antes de cualquier efecto, incluida la invalidez detectada desde la ejecución. |
| RNF-06 | No confiar en autoridad declarada por el agente. |
| RNF-07 | No devolver secretos en formularios, errores ni trazas. |
| RNF-08 | Rechazar versiones incompatibles de forma explícita. |
| RNF-10 | Mantener reglas de dominio fuera del núcleo y justificar dependencias. |

RNF-04: el desarrollo y la PoC se ejecutan inicialmente en Linux/WSL; Windows nativo diferido. RNF-09: latencia y tokens se miden en la puerta de §7 ([EXPERIMENT.md](experiment/EXPERIMENT.md)); no se promete rendimiento sin medir.

## 6. Criterios de aceptación de la PoC

| ID | Resultado observable | RF |
|---|---|---|
| CA-01 | Entrada completa: una invocación, efecto esperado y resultado correcto. | 01/02 |
| CA-02 | Campo obligatorio ausente: formulario contextual con la obligatoriedad y las restricciones de los campos faltantes, cero efectos de dominio. | 03/04 |
| CA-03 | Varios errores en una petición: se comunican todos los detectados en un único intercambio; se conservan los datos conocidos, presentes, no nulos, no sensibles y no rechazados, sin que la conservación certifique validez; corrección válida completa la operación. | 04/05/06 |
| CA-04 | Operación denegada o restricciones manipuladas: no se ejecuta ningún efecto no autorizado. | 07/08/11 |
| CA-05 | Fixtures repetidos producen decisiones semánticamente iguales; salida JSON interpretable sin prompts interactivos. | 01/12/13 |
| CA-06 | Fallo simulado del dominio no se presenta como éxito; versión incompatible rechazada. | 07/01 |
| CA-07 | Invalidez detectada desde la ejecución: resultado `invalid` reparable sin efectos de dominio; regla sin `schema` fiel se valida igualmente por el dominio. | 05/08 |

Los fixtures deben pasar al 100 %. Esta numeración sustituye a la de v0.1.0: los criterios de concurrencia, reintento y revisión de estado quedan diferidos a una versión posterior y el resto se absorbe en los escenarios anteriores. No se exige una segunda implementación.

## 7. Puerta de producto previa a Specification

El procedimiento, las métricas completas y el plazo están en [EXPERIMENT.md](experiment/EXPERIMENT.md). Aquí solo se fija la puerta. Se compara ACM con una CLI JSON competente que comparta ejecutor, validadores y permisos. Los umbrales son criterios de producto propuestos, no significación estadística, y se fijan antes de medir.

| Criterio de GO | Umbral provisional |
|---|---|
| Éxito | Al menos igual al baseline y 14/15 o más |
| Tokens | AIF ≤ 125 % del baseline |
| Efectos no autorizados o inválidos | Cero en AIF |
| Beneficio | Al menos 30 % menos errores de contrato o 20 % menos invocaciones en reparación |

Orden de evaluación, aprobado por el PO: (1) defecto de integridad → **NO-GO**; (2) medición incompleta o no comparable → **INCONCLUSO**; (3) éxito o tokens fuera de umbral con medición válida → **NO-GO**, aunque el beneficio no sea medible; (4) beneficio no medible (baseline sin errores ni reparaciones) → **INCONCLUSO**; (5) beneficio medible por debajo del umbral → **NO-GO**; (6) los cuatro criterios cumplidos → **GO**. El PO decide continuar, reformular o detener; no se inicia Specification antes de un GO explícito. Latencia y camino directo se registran de forma descriptiva.

## 8. Decisiones y preparación

| Decisión | Estado |
|---|---|
| D-01: contrato independiente del lenguaje; CLI primero | Dirección acordada |
| D-02: investigación y experimento antes de Specification; Specification antes de implementar | Acordada |
| D-03: Go como referencia | Propuesta pendiente, reevaluar tras experimento |
| D-04: perfil de JSON Schema y política de extensiones | Resuelta en la Specification v0.2: `schema` opcional en sintaxis JSON Schema 2020-12, no exigible cuando no expresa fielmente la regla; extensiones del ejecutor permitidas y fuera del contrato |
| D-05: persistencia y caducidad | Diferida; no forma parte del núcleo |
| D-06: sintaxis, tipos exactos, códigos de salida y versionado | Resuelta en la Specification v0.2: versión única `0.2`, rechazo explícito de cualquier otra |
| D-07: licencia, nombre, gobernanza y publicación | Pendiente; resolver antes de cualquier publicación |
| D-08: umbrales y protocolo | Aprobados por el PO el 2026-11-24 (umbrales en §7, protocolo en EXPERIMENT.md) |
| D-09: metodología de desarrollo — sin documento SDD; flujo ODD del proyecto (tablero Tasks-AI, Orchestrator, ACM) | Aprobada por el PO el 2026-11-24, en la revisión de SPECIFICATION v0.1 |
| D-10: protocolo v0.2 único — sin estado entre peticiones, validación progresiva o exhaustiva, todos los errores detectados en un intercambio, invalidez reparable sin efectos y `collected` que conserva sin certificar | Aprobada por el PO; materializada en SPECIFICATION v0.2 |

**PARCIAL:** alcance y criterios definidos; base técnica implementada y verificada sobre el contrato v0.2 (§3.1); hipótesis de usuarios sin validación externa; experimento ejecutado con resultado **INCONCLUSO** por medición no comparable (§7), sin evidencia de superioridad frente a una CLI JSON competente; auditoría externa de fuentes parcial (véase RESEARCH, adenda); aprobación de este PRD pendiente. El producto completo de PRODUCT.md no está listo: AIF CLI como producto y AIF Definition siguen pendientes de diseño (§3.3).

## 9. Secuencia y control de cambios

RESEARCH → PRD revisado → experimento ACM → GO / NO-GO / INCONCLUSO → Specification →
PoC y fixtures. Benchmark ampliado solo si la evidencia lo justifica. No hay etapa de
documento SDD (D-09): la Specification aprobada es la entrada directa de la
implementación, dirigida por el flujo ODD del proyecto.

v0.1.0: borrador inicial. v0.1.1: recorte del MVP, elimina dependencias implícitas con el Runner, puerta temprana con objetivos medibles, estado PARCIAL. v0.1.2: se corrige la numeración (un ID por requisito, CA contiguos), se recuperan D-02, D-04, D-06 y D-07 y se conserva RNF-04 (Linux/WSL inicial); el protocolo experimental pasa a EXPERIMENT.md; §7 conserva la puerta con cuatro criterios. Un baseline sin margen de mejora da INCONCLUSO, no GO ni NO-GO. Se cierra el hueco de la puerta: cualquier criterio incumplido con medición válida es NO-GO. Se corrigen la referencia a RESEARCH v0.1.2 y la matriz comparativa de RESEARCH §5 (reintento seguro diferido a v0.2). v0.1.3: el PO aprueba el orden de evaluación de la puerta (§7) y la primera llamada prefijada del experimento ([EXPERIMENT.md](experiment/EXPERIMENT.md) §4). v0.1.4: el PO aprueba el protocolo y los umbrales (D-08), el modelo `nan/deepseek-v4-flash#medium` y el aislamiento y registro mínimos; la ejecución sigue pendiente de autorización de destino. v0.1.5: decisión del PO (D-09): la secuencia §9 no incluye etapa de documento SDD; el desarrollo sigue el flujo ODD del proyecto (tablero Tasks-AI, Orchestrator, ACM), sin documento sustitutivo ni metodología paralela. v0.1.6: alineación con la visión de producto aprobada en [PRODUCT.md](PRODUCT.md); el PRD separa base técnica implementada (§3.1), visión del producto completo (§3.2) y capacidades pendientes de diseño (§3.3), sin duplicar la visión ni alterar requisitos, decisiones D-01..D-09, puerta de §7 ni secuencia de §9. v0.2.0: se adopta el protocolo AIF 0.2 único aprobado por el PO (D-10): interacción sin estado, validación progresiva o exhaustiva con todos los errores detectados en un único intercambio, `collected` que conserva los datos conocidos, presentes, no nulos, no sensibles y no rechazados sin certificar validez, invalidez reparable sin efectos aunque se detecte al ejecutar, y `schema` no exigible cuando no expresa fielmente la regla; se actualizan RF-03/05/08, RNF-05, CA-02/03/07, D-04, D-06 y el estado de la base técnica; §7 y la historia del experimento no cambian.
