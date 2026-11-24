# AIF — Agent Interaction Forms
## Product Requirements Document

| Metadato | Valor |
|---|---|
| ID | AIF-PRD |
| Versión | 0.1.4 |
| Fecha | 2026-11-24 |
| Estado de preparación | PARCIAL |
| Fuente | [RESEARCH.md](RESEARCH.md), v0.1.2 |
| Estado de aprobación | Pendiente de revisión del PO |

## 1. Visión y problema

AIF aspira a ser un contrato abierto, ligero e independiente del lenguaje para la interacción determinista entre agentes y herramientas. Una operación completa se ejecuta directamente; una incompleta devuelve los datos necesarios para continuar; una inválida permite corregir sin reconstruir los campos válidos.

El problema es la fricción de descubrir sintaxis, interpretar errores y reconstruir argumentos. ACM demuestra viabilidad, no superioridad: su README informa de 22.900 tokens AIF frente a 12.700 convencional (+80,3 %) y 106.000 db-ai directo, con n=1 por variante. No existe todavía una línea base comparable frente a una CLI JSON competente.

## 2. Usuarios y objetivos

Personas hipotéticas, no validadas mediante entrevistas: agente consumidor que necesita completar operaciones; desarrollador que busca una interfaz reutilizable; operador que necesita preservar las reglas de negocio. La experiencia con ACM motiva estas hipótesis, pero no acredita demanda general.

- OBJ-01: reducir errores de contrato y facilitar su recuperación.
- OBJ-02: mantener una sola invocación cuando los datos bastan.
- OBJ-03: demostrar utilidad frente a una CLI JSON bien diseñada antes de construir infraestructura.

## 3. Alcance v0.1

Tres entregables, condicionados a la puerta de producto: (1) especificación mínima y perfil CLI; (2) implementación de referencia con fixtures de conformidad; (3) PoC aislada con ACM. Go es candidato, no decisión cerrada. No se modifica ACM de producción.

El núcleo no administra sesiones, almacenamiento, claves idempotentes, caducidad, revisiones ni concurrencia. Las garantías existentes de ACM permanecen en ACM; no se eliminan ni se generalizan. El runner, Python como segunda implementación, MCP/HTTP y las garantías de reintento quedan como candidatos v0.2, sujetos a una nueva decisión de producto. Tampoco entran orquestación, GUI ni servicios residentes.

## 4. Requisitos funcionales

Cada requisito conserva su ID de v0.1.0. Las lagunas de numeración (RF-09, RF-10, RF-14, RF-15) son requisitos diferidos o reducidos; sus IDs no se reutilizan. P01–P08 corresponden a RESEARCH §11.

| ID | Requisito MVP | Origen |
|---|---|---|
| RF-01 | Identificar operación, versión y resultado mediante contrato de máquina. | P01 |
| RF-02 | Ejecutar directamente con datos suficientes y válidos. | P02 |
| RF-03 | Ante datos ausentes, devolver los campos pertinentes con obligatoriedad y restricciones explícitas. | P02 |
| RF-04 | Distinguir la ausencia de datos de su invalidez. | P03 |
| RF-05 | Identificar los errores y preservar los valores válidos no sensibles. | P03 |
| RF-06 | Indicar la continuación sin requerir interpretar ayuda general. | P04 |
| RF-07 | Distinguir éxito, necesidad de datos, validación y denegación; informar de fallos de ejecución sin declarar éxito falso. | P03 |
| RF-08 | Revalidar toda entrada y su operación usando reglas confiables del ejecutor, sin confiar en metadatos enviados por el agente. No requiere almacenamiento de formularios. | P06 |
| RF-11 | El dominio decide permisos y ejecuta; AIF no concede autoridad. | P06 |
| RF-12 | Perfil CLI no interactivo con salida JSON y diagnósticos separados. | P07 |
| RF-13 | Fixtures de conformidad ejecutables sin LLM. | P08 |

RF-09/10/15 diferidos (idempotencia, revisión, otros transportes). RF-14 se limita a no prometer garantías de reintento inexistentes; no se implementa un motor de reintentos. Una denegación no invita a corregir para eludir permisos.

## 5. Requisitos no funcionales

| ID | Requisito |
|---|---|
| RNF-01 | Decisiones reproducibles con iguales entradas, reglas y contexto. |
| RNF-02 | Sin LLM, red ni servicio residente obligatorio. |
| RNF-03 | Respuestas compactas y límites de entrada documentados por la especificación. |
| RNF-05 | Sin efectos de dominio por entradas inválidas o denegadas. |
| RNF-06 | No confiar en autoridad declarada por el agente. |
| RNF-07 | No devolver secretos en formularios, errores ni trazas. |
| RNF-08 | Rechazar versiones incompatibles de forma explícita. |
| RNF-10 | Mantener reglas de dominio fuera del núcleo y justificar dependencias. |

RNF-04: el desarrollo y la PoC se ejecutan inicialmente en Linux/WSL; Windows nativo diferido. RNF-09: latencia y tokens se miden en la puerta de §7 ([EXPERIMENT.md](EXPERIMENT.md)); no se promete rendimiento sin medir.

## 6. Criterios de aceptación de la PoC

| ID | Resultado observable | RF |
|---|---|---|
| CA-01 | Entrada completa: una invocación, efecto esperado y resultado correcto. | 01/02 |
| CA-02 | Campo obligatorio ausente: formulario contextual, cero efectos de dominio. | 03/04 |
| CA-03 | Enum inválido: error localizado; se conservan todos los valores válidos no sensibles; corrección válida completa la operación. | 04/05/06 |
| CA-04 | Operación denegada o restricciones manipuladas: no se ejecuta ningún efecto no autorizado. | 07/08/11 |
| CA-05 | Fixtures repetidos producen decisiones semánticamente iguales; salida JSON interpretable sin prompts interactivos. | 01/12/13 |
| CA-06 | Fallo simulado del dominio no se presenta como éxito; versión incompatible rechazada. | 07/01 |

Los fixtures deben pasar al 100 %. Esta numeración sustituye a la de v0.1.0: los criterios de concurrencia, reintento y revisión de estado quedan diferidos a v0.2 y el resto se absorbe en los escenarios anteriores. No se exige una segunda implementación.

## 7. Puerta de producto previa a Specification

El procedimiento, las métricas completas y el plazo están en [EXPERIMENT.md](EXPERIMENT.md). Aquí solo se fija la puerta. Se compara ACM con una CLI JSON competente que comparta ejecutor, validadores y permisos. Los umbrales son criterios de producto propuestos, no significación estadística, y se fijan antes de medir.

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
| D-04: perfil de JSON Schema y política de extensiones | Specification, tras GO |
| D-05: persistencia y caducidad | Diferida; no forma parte del núcleo v0.1 |
| D-06: sintaxis, tipos exactos, códigos de salida y versionado | Specification, tras GO |
| D-07: licencia, nombre, gobernanza y publicación | Pendiente; resolver antes de cualquier publicación |
| D-08: umbrales y protocolo | Aprobados por el PO el 2026-11-24 (umbrales en §7, protocolo en EXPERIMENT.md) |

**PARCIAL:** alcance y criterios definidos; hipótesis de usuarios sin validación externa; línea base comparable y experimento pendientes; auditoría externa de fuentes parcial (véase RESEARCH, adenda); decisión de lenguaje y aprobación del PO pendientes. No listo para implementación.

## 9. Secuencia y control de cambios

RESEARCH → PRD revisado → experimento ACM → GO / NO-GO / INCONCLUSO → Specification → SDD mínimo → PoC y fixtures. Benchmark ampliado solo si la evidencia lo justifica.

v0.1.0: borrador inicial. v0.1.1: recorte del MVP, elimina dependencias implícitas con el Runner, puerta temprana con objetivos medibles, estado PARCIAL. v0.1.2: se corrige la numeración (un ID por requisito, CA contiguos), se recuperan D-02, D-04, D-06 y D-07 y se conserva RNF-04 (Linux/WSL inicial); el protocolo experimental pasa a EXPERIMENT.md; §7 conserva la puerta con cuatro criterios. Un baseline sin margen de mejora da INCONCLUSO, no GO ni NO-GO. Se cierra el hueco de la puerta: cualquier criterio incumplido con medición válida es NO-GO. Se corrigen la referencia a RESEARCH v0.1.2 y la matriz comparativa de RESEARCH §5 (reintento seguro diferido a v0.2). v0.1.3: el PO aprueba el orden de evaluación de la puerta (§7) y la primera llamada prefijada del experimento ([EXPERIMENT.md](EXPERIMENT.md) §4). v0.1.4: el PO aprueba el protocolo y los umbrales (D-08), el modelo `nan/deepseek-v4-flash#medium` y el aislamiento y registro mínimos; la ejecución sigue pendiente de autorización de destino.
