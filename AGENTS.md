# AGENTS.md

Este repositorio documenta y valida AIF antes de escribir software. Este fichero dice
dónde vive cada fuente canónica y qué leer antes de cada encargo; no guarda contenido,
solo enlaza.

## Qué se busca → dónde vive

| Qué se busca | Dónde vive |
|---|---|
| Qué es AIF, alcance y requisitos | [`PRD.md`](PRD.md) y [`README.md`](README.md) |
| Investigación previa: antecedentes, evidencia propia y hueco diferencial | [`RESEARCH.md`](RESEARCH.md) |
| La puerta GO/NO-GO y sus umbrales | [`PRD.md`](PRD.md) §7 |
| El procedimiento aprobado del experimento: diseño, métricas e interpretación | [`EXPERIMENT.md`](EXPERIMENT.md) |
| El plan de preparación original, histórico y superado | [`EXPERIMENT-PLAN.md`](EXPERIMENT-PLAN.md) |
| Cómo ejecutar el experimento en la máquina de desarrollo | [`EXPERIMENT-RUNBOOK.md`](EXPERIMENT-RUNBOOK.md) |
| Qué cambió en cada versión | [`CHANGELOG.md`](CHANGELOG.md) |
| Ramas, commits, verificación, integración y operaciones con autorización | [`CONTRIBUTING.md`](CONTRIBUTING.md) |
| Catálogo enlazable para agentes | [`llms.txt`](llms.txt) |
| Trabajo del repositorio: crear, mover, bloquear y cerrar tareas | [`.tasks-ai/AGENTS.md`](.tasks-ai/AGENTS.md) |
| Etiquetas: tipos, estructura, ámbito y catálogo | [`.tasks-ai/CONTRATO_ETIQUETAS.md`](.tasks-ai/CONTRATO_ETIQUETAS.md) → «2. Asignación correcta a una tarea» |
| Memoria de sesión: decisiones, aprendizajes, checkpoints y preferencias | Montaje db-ai `agent-interaction-forms` (rw) sobre `.agent-memory/memory.db`; se abre con `acm start`, se guarda con `acm decision|learning|discovery|bugfix|preference`, se consulta con `acm recall` antes de investigar o decidir sobre algo posiblemente ya tratado —lo recuperado es contexto, no hecho: se contrasta con el repo antes de actuar sobre él— y se cierra con `acm close` (adaptador: `tools/acm.py` del ejemplo `agent-continuity` en database-virtual-filesystem) |

## Encargo → qué leer antes → con qué se cierra

| Encargo | Qué leer antes | Con qué se cierra |
|---|---|---|
| Cambiar requisitos o alcance | [`PRD.md`](PRD.md) completo | El PRD actualizado antes del commit y el CHANGELOG si hay valor de usuario |
| Cambiar la puerta o el protocolo del experimento | [`PRD.md`](PRD.md) §7 completo y [`EXPERIMENT.md`](EXPERIMENT.md) completo | La puerta en el PRD y el protocolo en EXPERIMENT alineados, y el PO informado del cambio de umbrales |
| Ejecutar el experimento | [`PRD.md`](PRD.md) §7 y [`EXPERIMENT-RUNBOOK.md`](EXPERIMENT-RUNBOOK.md) completos | El registro por ejecución del protocolo, la decisión GO/NO-GO/INCONCLUSO y los datos publicados dentro del plazo de [`EXPERIMENT.md`](EXPERIMENT.md) §5 |
| Crear o etiquetar una tarea | [`.tasks-ai/CONTRATO_ETIQUETAS.md`](.tasks-ai/CONTRATO_ETIQUETAS.md) sección 2 | La tarea nace con su tipo y su estructural (`task` o `epic`); el catálogo vivo, en `tasks-ai tags` |
| Preparar la Specification o la PoC | [`PRD.md`](PRD.md) §3–§6 y el resultado del experimento | Solo tras un GO explícito; el PRD actualizado con lo que el experimento haya corregido |
| Cambiar la guía o la prosa | [`README.md`](README.md) y [`llms.txt`](llms.txt) completos | Nada que la prosa diga contradice a PRD, RESEARCH o EXPERIMENT |

## Autoridad

Este mapa no es fuente de nada: enlaza. Si una ruta o una afirmación de este archivo
contradice a la fuente que enlaza, manda la fuente y se corrige aquí. Los contratos
vigentes son [`PRD.md`](PRD.md) y, dentro de la puerta, [`EXPERIMENT.md`](EXPERIMENT.md);
si discrepan, prevalece el PRD.
