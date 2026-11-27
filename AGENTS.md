# AGENTS.md

Este repositorio alberga AIF: una base técnica funcional (biblioteca Go y contrato
v0.2) dentro de una visión de producto aprobada en [`docs/PRODUCT.md`](docs/PRODUCT.md).
La evolución v0.2 está autorizada por el PO y no constituye una publicación; v0.1.1
queda como contrato histórico implementado. Este fichero dice
dónde vive cada fuente canónica y qué leer antes de cada encargo; no guarda contenido,
solo enlaza.

## Visión de producto y estado (resumen, no fuente)

- AIF es un producto con dos modalidades complementarias que comparten núcleo y contrato:
  **integración nativa** (biblioteca Go reutilizable dentro del ejecutable de la
  aplicación) e **integración externa** (ejecutable AIF como capa sobre herramientas
  existentes, con definiciones declarativas, pendiente de diseño). La visión completa
  vive en [`docs/PRODUCT.md`](docs/PRODUCT.md); este fichero no la duplica.
- La congelación documental se levanta únicamente para la evolución autorizada v0.2
  (protocolo y biblioteca); esa autorización no equivale a una publicación ni aprueba
  el producto completo.
- Jerarquía de fuentes: visión → [`PRODUCT.md`](docs/PRODUCT.md); requisitos →
  [`PRD.md`](docs/PRD.md); contrato normativo → [`SPECIFICATION.md`](docs/SPECIFICATION.md);
  investigación → [`RESEARCH.md`](docs/RESEARCH.md); evidencia → [`docs/experiment/`](docs/experiment).

## Qué se busca → dónde vive

| Qué se busca | Dónde vive |
|---|---|
| La visión de producto aprobada (qué es AIF, problema, modalidades, componentes, estado) | [`PRODUCT.md`](docs/PRODUCT.md) — fuente canónica de visión, aprobada por el PO |
| Requisitos y alcance | [`PRD.md`](docs/PRD.md) |
| Qué es AIF, alcance y requisitos | [`PRD.md`](docs/PRD.md) y [`README.md`](README.md) |
| Investigación previa: antecedentes, evidencia propia y hueco diferencial | [`RESEARCH.md`](docs/RESEARCH.md) |
| La puerta GO/NO-GO y sus umbrales | [`PRD.md`](docs/PRD.md) §7 |
| El procedimiento aprobado del experimento: diseño, métricas e interpretación | [`EXPERIMENT.md`](docs/experiment/EXPERIMENT.md) |
| El plan de preparación original, histórico y superado | [`EXPERIMENT-PLAN.md`](docs/experiment/EXPERIMENT-PLAN.md) |
| Cómo ejecutar el experimento en la máquina de desarrollo | [`EXPERIMENT-RUNBOOK.md`](docs/experiment/EXPERIMENT-RUNBOOK.md) |
| Qué cambió en cada versión | [`CHANGELOG.md`](CHANGELOG.md) |
| Ramas, commits, verificación, integración y operaciones con autorización | [`CONTRIBUTING.md`](CONTRIBUTING.md) |
| Catálogo enlazable para agentes | [`llms.txt`](llms.txt) |
| Trabajo del repositorio: crear, mover, bloquear y cerrar tareas | [`.tasks-ai/AGENTS.md`](.tasks-ai/AGENTS.md) |
| Etiquetas: tipos, estructura, ámbito y catálogo | [`.tasks-ai/CONTRATO_ETIQUETAS.md`](.tasks-ai/CONTRATO_ETIQUETAS.md) → «2. Asignación correcta a una tarea» |
| Memoria de sesión: decisiones, aprendizajes, checkpoints y preferencias | Montaje db-ai `agent-interaction-forms` (rw) sobre `.agent-memory/memory.db`; se abre con `acm start`, que recupera el contexto de la sesión anterior en una invocación, y `acm aif` es la interfaz autodescriptiva del resto de operaciones: el envelope de descubrimiento (`{"aif":"0.2"}`) por stdin devuelve su lista, sin esquemas que reproducir aquí; lo recuperado es contexto, no hecho: se contrasta con el repo antes de actuar sobre él |

## Encargo → qué leer antes → con qué se cierra

| Encargo | Qué leer antes | Con qué se cierra |
|---|---|---|
| Cambiar requisitos o alcance | [`PRD.md`](docs/PRD.md) completo | El PRD actualizado antes del commit y el CHANGELOG si hay valor de usuario |
| Cambiar la puerta o el protocolo del experimento | [`PRD.md`](docs/PRD.md) §7 completo y [`EXPERIMENT.md`](docs/experiment/EXPERIMENT.md) completo | La puerta en el PRD y el protocolo en EXPERIMENT alineados, y el PO informado del cambio de umbrales |
| Ejecutar el experimento | [`PRD.md`](docs/PRD.md) §7 y [`EXPERIMENT-RUNBOOK.md`](docs/experiment/EXPERIMENT-RUNBOOK.md) completos | El registro por ejecución del protocolo, la decisión GO/NO-GO/INCONCLUSO y los datos publicados dentro del plazo de [`EXPERIMENT.md`](docs/experiment/EXPERIMENT.md) §5 |
| Crear o etiquetar una tarea | [`.tasks-ai/CONTRATO_ETIQUETAS.md`](.tasks-ai/CONTRATO_ETIQUETAS.md) sección 2 | La tarea nace con su tipo y su estructural (`task` o `epic`); el catálogo vivo, en `tasks-ai tags` |
| Preparar la Specification o la PoC | [`PRD.md`](docs/PRD.md) §3–§6 y el resultado del experimento | Solo tras un GO explícito; el PRD actualizado con lo que el experimento haya corregido |
| Cambiar la guía o la prosa | [`README.md`](README.md) y [`llms.txt`](llms.txt) completos | Nada que la prosa diga contradice a PRD, RESEARCH o EXPERIMENT |

## Autoridad

Este mapa no es fuente de nada: enlaza. Si una ruta o una afirmación de este archivo
contradice a la fuente que enlaza, manda la fuente y se corrige aquí. Fuente canónica
de la visión del producto: [`docs/PRODUCT.md`](docs/PRODUCT.md), aprobada por el PO.
Los contratos vigentes son [`PRD.md`](docs/PRD.md) y
[`SPECIFICATION.md`](docs/SPECIFICATION.md) v0.2 (evolución autorizada por el PO;
historia de versiones en [`CHANGELOG.md`](CHANGELOG.md)), junto con
[`EXPERIMENT.md`](docs/experiment/EXPERIMENT.md) dentro de la puerta (registro
INCONCLUSO conservado como evidencia histórica); si discrepan en requisitos,
prevalece el PRD; en visión, PRODUCT. La congelación se levanta únicamente para la
evolución autorizada v0.2, que no equivale a una publicación.
