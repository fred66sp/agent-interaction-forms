# AIF — Agent Interaction Forms

*Herramientas que saben interactuar con agentes de inteligencia artificial.*

## Qué es AIF

AIF permite dotar a una herramienta de software de una interfaz de interacción
autodescriptiva, determinista y orientada a agentes: la herramienta comunica de
forma estructurada qué necesita, qué admite y cómo continuar, en lugar de obligar
al agente a descubrirlo. Cubre tres situaciones: **información completa** (se
ejecuta), **incompleta** (formulario contextual) e **incorrecta** (se identifican
los errores, se conservan los valores válidos no sensibles y se indica cómo seguir).
No es un agente, un orquestador, un gestor de workflows ni un sistema de formularios
gráficos, ni sustituye la lógica de negocio; el dominio conserva la autoridad sobre
sus operaciones, permisos y reglas.

## El problema

Las herramientas tradicionales están diseñadas para personas. Cuando un agente las
usa, consulta documentación, interpreta errores, descubre parámetros, reconstruye
peticiones y encadena invocaciones: incertidumbre, contexto y esfuerzo extra. AIF
propone que la herramienta, no el agente, describa sus necesidades de interacción.

## Ejemplo sencillo

Un agente pide crear una tarea titulada «Revisar documentación». La herramienta
necesita también la prioridad y, en lugar de un error genérico, le comunica el
título recibido, la información pendiente (prioridad) y los valores admitidos
(baja, media y alta). El agente la completa y repite la operación; con todos los
datos desde el principio, se habría ejecutado en una invocación.

**Principio fundamental: una necesidad, una invocación siempre que la información
disponible sea suficiente.**

## Dos modalidades

- **Integración nativa:** la biblioteca AIF se incorpora al código de la aplicación,
  que gana las capacidades de interacción sin ejecutable intermediario. Por ejemplo,
  Tasks-AI podría integrarla en su binario Go conservando su lógica de negocio.
- **Integración externa:** un ejecutable AIF actúa de intermediario entre el agente
  y una herramienta existente descrita por un fichero de definición. **Pendiente de
  diseño; no disponible.**

Ambas modalidades comparten núcleo y contrato de interacción.

## Estado actual

| Componente | Estado |
|---|---|
| AIF Core | Implementación inicial disponible en Go. |
| AIF Specification | Contrato v0.1 definido y aprobado. |
| CLI de demostración | Disponible. |
| Pruebas de conformidad | Disponibles. |
| Prueba de concepto con ACM | Realizada. |
| AIF CLI como producto y AIF Definition | **Pendientes de diseño; no disponibles** (la CLI de demostración `aifdemo` existe, pero no es la capa externa del producto). |

La base técnica es funcional, pero no es el producto completo ni hay beneficios
cuantitativos demostrados: el experimento previo quedó **INCONCLUSO**
([EXPERIMENT-RESULTS.md](docs/experiment/EXPERIMENT-RESULTS.md)).

## Acceso a documentación

| Documento | Qué contiene |
|---|---|
| [PRODUCT.md](docs/PRODUCT.md) | Visión de producto aprobada por el PO (v0.1) |
| [PRD.md](docs/PRD.md) | Requisitos, alcance, criterios de la PoC y puerta GO/NO-GO (§7) |
| [SPECIFICATION.md](docs/SPECIFICATION.md) | Contrato normativo v0.1.1: modelo, resultados y perfil CLI |
| [RESEARCH.md](docs/RESEARCH.md) | Investigación previa, evidencia propia y hueco diferencial |
| [EXPERIMENT.md](docs/experiment/EXPERIMENT.md) | Protocolo aprobado: diseño, métricas y plazo |
| [EXPERIMENT-RESULTS.md](docs/experiment/EXPERIMENT-RESULTS.md) | Resultados y lectura; **INCONCLUSO** |
| [EXPERIMENT-RUNBOOK.md](docs/experiment/EXPERIMENT-RUNBOOK.md) · [PLAN](docs/experiment/EXPERIMENT-PLAN.md) | Guía operativa · plan histórico superado |
| [docs/web](docs/web/index.html) | Guía navegable: contrato, mensajes, integración y experimento |

También: [AGENTS.md](AGENTS.md), [CONTRIBUTING.md](CONTRIBUTING.md),
[CHANGELOG.md](CHANGELOG.md) y [`llms.txt`](llms.txt).

## Implementación de referencia (MVP v0.1)

Go sin dependencias externas (los §§ remiten a [`docs/SPECIFICATION.md`](docs/SPECIFICATION.md)): [`aif`](aif.go) implementa el núcleo (§§1–6) y
`Executor.Handle`; [`cmd/aifdemo`](cmd/aifdemo) ejercita el perfil CLI (§7);
[`conformance/`](conformance) reúne CJ-01..CJ-13 (§9) con su runner; y
[`poc/acm/`](poc/acm) es la PoC aislada con ACM (el proyecto que motivó AIF: su
adaptador de memoria, montado sobre `db-ai`, fue el primer integrador real). Contrato normativo y detalle:
[docs/SPECIFICATION.md](docs/SPECIFICATION.md) y [docs/web](docs/web/index.html).

### Ejemplo de uso de la biblioteca

```go
op := &aif.Operation{
    ID: "record_observation", Description: "Registra una observación",
    Fields: []aif.Field{
        {Name: "title", Required: true, Type: "string", Schema: json.RawMessage(`{"maxLength":40}`)},
        {Name: "kind", Required: true, Allowed: []string{"learning", "decision"}},
    },
    Execute: func(in map[string]json.RawMessage) (json.RawMessage, error) {
        return json.Marshal(map[string]string{"id": "obs_0001"}) // dominio
    },
}
resp := (&aif.Executor{Operations: []*aif.Operation{op}}).Handle(rawRequest)
```

### Pruebas y demostración

```bash
go test ./...            # biblioteca + fixtures de conformidad (100 %) + PoC
go run ./cmd/aifdemo     # lee una petición AIF de stdin, escribe la respuesta
bash poc/acm/run.sh      # PoC aislada de extremo a extremo (4 escenarios)
```

La CLI no es interactiva: petición por stdin, respuesta JSON única en stdout y
diagnósticos en stderr.
