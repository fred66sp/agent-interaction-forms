# AIF — Agent Interaction Forms

AIF es un contrato abierto, ligero e independiente del lenguaje para la interacción
determinista entre agentes y herramientas. Su objetivo es que una operación completa se
ejecute con una sola invocación, que una incompleta devuelva los datos necesarios para
continuar y que una inválida permita corregir sin reconstruir los campos válidos.

## Cómo funciona

- **Operación completa:** los datos bastan y son válidos; el ejecutor la ejecuta y devuelve
  el resultado mediante un contrato de máquina.
- **Operación incompleta:** devuelve un formulario contextual con los campos pertinentes,
  su obligatoriedad y sus restricciones, y una continuación explícita.
- **Operación inválida:** identifica los errores y conserva los valores válidos no sensibles
  para que el agente corrija sin repetir lo que ya era correcto.

El dominio conserva la autoridad: AIF no concede permisos y el ejecutor revalida toda
entrada con sus propias reglas, sin confiar en los metadatos que declare el agente.

## Estado del proyecto

| Elemento | Estado |
|---|---|
| PRD | v0.1.5; preparación **PARCIAL**, pendiente de aprobación del PO |
| RESEARCH | v0.1.2; investigación inicial, no normativa |
| SPECIFICATION | v0.1.1 **APROBADO** por el PO (2026-11-24), tras la revisión integrada |
| Experimento previo a Specification | Protocolo v0.1.6; ejecutado el 2026-11-24 (30/30 mediciones), resultado **INCONCLUSO** (PRD §7, orden 2) |
| Implementación de referencia | No iniciada; Go es candidato, no decisión cerrada |

La secuencia acordada es RESEARCH → PRD → experimento → GO / NO-GO / INCONCLUSO →
Specification → PoC con fixtures; no hay etapa de documento SDD (decisión D-09 del
PO: el desarrollo sigue el flujo ODD del proyecto). No se inicia la Specification sin un GO
explícito del PO. El alcance v0.1 son tres entregables condicionados a esa puerta:
especificación mínima y perfil CLI, implementación de referencia con fixtures de
conformidad, y PoC aislada con ACM. No se modifica ACM de producción.

## Documentos base

| Documento | Qué contiene |
|---|---|
| [PRD.md](PRD.md) | Visión, usuarios, alcance, requisitos, criterios de aceptación de la PoC y la puerta GO/NO-GO de §7 |
| [RESEARCH.md](RESEARCH.md) | Investigación previa: antecedentes, evidencia propia de ACM, hueco diferencial y riesgos |
| [EXPERIMENT.md](EXPERIMENT.md) | Protocolo aprobado del experimento: diseño, métricas, interpretación y plazo |
| [EXPERIMENT-PLAN.md](EXPERIMENT-PLAN.md) | Plan de preparación original; **histórico y superado** por EXPERIMENT.md y el runbook |
| [EXPERIMENT-RUNBOOK.md](EXPERIMENT-RUNBOOK.md) | Guía operativa para quien ejecute el experimento en la máquina de desarrollo |

## Evidencia de partida

ACM demuestra viabilidad, no superioridad. Su registro informa de 22.900 tokens AIF frente
a 12.700 de una interfaz convencional (+80,3 %) y 106.000 de db-ai directo, con n=1 por
variante. No existe todavía una línea base comparable frente a una CLI JSON competente; el
experimento de [EXPERIMENT.md](EXPERIMENT.md) es el filtro de inversión que debe producirla
antes de construir infraestructura.

## Trabajo en el repositorio

- [AGENTS.md](AGENTS.md) — mapa de dónde vive cada fuente canónica y qué leer antes de cada encargo.
- [CONTRIBUTING.md](CONTRIBUTING.md) — política de ramas, commits, verificación y autorizaciones.
- [CHANGELOG.md](CHANGELOG.md) — qué cambió en cada versión.
- [`llms.txt`](llms.txt) — catálogo enlazable para agentes.
- [`.tasks-ai/AGENTS.md`](.tasks-ai/AGENTS.md) — el tablero del proyecto y sus operaciones.
