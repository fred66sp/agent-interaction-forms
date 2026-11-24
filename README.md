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
| Implementación de referencia | **MVP v0.1 implementado** en Go: biblioteca `aif`, fixtures CJ-01..CJ-13, CLI `aifdemo` y PoC aislada con ACM |

La secuencia acordada es RESEARCH → PRD → experimento → GO / NO-GO / INCONCLUSO →
Specification → PoC con fixtures; no hay etapa de documento SDD (decisión D-09 del
PO: el desarrollo sigue el flujo ODD del proyecto). No se inicia la Specification sin un GO
explícito del PO. El alcance v0.1 son tres entregables condicionados a esa puerta:
especificación mínima y perfil CLI, implementación de referencia con fixtures de
conformidad, y PoC aislada con ACM. No se modifica ACM de producción.

## Implementación de referencia (MVP v0.1)

Implementación en Go, sin dependencias externas (solo la biblioteca estándar):

- **Biblioteca [`aif`](aif.go)** — el núcleo del contrato (§§1–6): sobre de
  peticiones, los ocho resultados, descripción de campos, valores sensibles,
  continuación sin estado y límites de entrada (L-04: 256 KiB, profundidad 32).
  Un integrador declara sus operaciones y ejecuta con `Executor.Handle`;
  la lógica de negocio, la autorización y la persistencia quedan en el integrador.
- **CLI de demostración `cmd/aifdemo`** — una operación (`record_observation`)
  que ejercita el perfil CLI completo (§7): descubrimiento, información ausente,
  corrección y continuación, con códigos de salida normativos.
- **Fixtures de conformidad [`conformance/`](conformance)** — un fichero JSON por
  caso CJ-01..CJ-13 (§9), independiente del lenguaje, más el runner que los
  ejecuta contra la CLI.
- **PoC aislada con ACM [`poc/acm/`](poc/acm)** — integra la biblioteca con el
  adaptador real de ACM (copiado a un entorno efímero): ejecución directa,
  información ausente, corrección de inválidos y conservación de las garantías
  de ACM (deduplicación por contenido). Sin tocar ACM de producción.

### Ejemplo de uso de la biblioteca

```go
op := &aif.Operation{
    ID:          "record_observation",
    Description: "Registra una observación",
    Fields: []aif.Field{
        {Name: "title", Required: true, Type: "string",
            Schema: json.RawMessage(`{"maxLength":40}`)},
        {Name: "kind", Required: true,
            Allowed: []string{"learning", "decision"}},
    },
    Execute: func(input map[string]json.RawMessage) (json.RawMessage, error) {
        return json.Marshal(map[string]string{"id": "obs_0001"}) // dominio
    },
}
executor := aif.Executor{Operations: []*aif.Operation{op}}
resp := executor.Handle(rawRequest) // *Response con el resultado de §2/§3
```

### Pruebas y demostración

```bash
go test ./...            # biblioteca + fixtures de conformidad (100 %) + PoC
go run ./cmd/aifdemo     # lee una petición AIF de stdin, escribe la respuesta
bash poc/acm/run.sh      # PoC aislada de extremo a extremo (4 escenarios)
```

La CLI no es interactiva: la petición llega por stdin, la respuesta es un único
objeto JSON en stdout y los diagnósticos van en stderr. `SPECIFICATION.md`
permanece como única fuente normativa del contrato; este README no lo duplica.

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
