# Registro de cambios

Cambios con valor para quien usa el proyecto, por versión. El formato sigue las
recomendaciones de *Keep a Changelog* y el versionado es semántico. Lo que se anota aquí
al publicar; la historia de cada cambio, commit a commit, vive en Git.

## [Sin publicar]

### Añadido

- Implementación de referencia del MVP v0.1 en Go (D-03), sin dependencias externas:
  biblioteca `aif` que materializa el contrato de SPECIFICATION v0.1.1 (sobre
  `aif`/`operation`/`input`, los ocho resultados con su contenido exigible,
  precedencia de `invalid`, conservación de valores válidos no sensibles con
  omisión de los sensibles, continuación sin estado con revalidación total,
  `malformed_request` con diagnóstico sintáctico determinista incluidas claves
  duplicadas, descripción de campos con `schema` JSON Schema 2020-12 representado
  sin motor propio), CLI de demostración `cmd/aifdemo` conforme al perfil CLI
  (códigos de salida 0/1/2/3/4, descubrimiento, diagnósticos separados),
  fixtures de conformidad CJ-01..CJ-13 independientes del lenguaje con runner
  que verifica resultados, códigos, miembros presentes y ausentes, ausencia de
  efectos de dominio y determinismo (14/14 PASS), y PoC aislada de integración
  con ACM (`poc/acm`) con base efímera y copia intacta del adaptador: 4/4
  escenarios PASS sin modificar ACM de producción.
- SPECIFICATION v0.1.1 (**APROBADO** por el PO el 2026-11-24): el
  contrato de interacción AIF v0.1 — un sobre JSON con ocho resultados exigibles
  (`executed`, `input_needed`, `invalid`, `denied`, `failed`,
  `unsupported_version`, `malformed_request`, `unknown_operation`), respuesta única
  con los informes de campos ausentes e inválidos y precedencia determinista,
  descripción mínima de campos con `schema` JSON Schema 2020-12, continuación sin
  estado, límites de responsabilidad entre interacción y dominio, perfil CLI con
  códigos de salida fijos y once ejemplos normativos, con trece casos de
  conformidad deterministas (CJ-01..CJ-13) trazables a los CA del PRD. Versión
  corregida sobre la revisión del PO de v0.1.0 (ocho hallazgos H-01..H-08).
- PRD v0.1.5: decisión D-09 del PO — la secuencia de desarrollo no incluye una etapa
  de documento SDD; el proyecto desarrolla con su metodología ODD (tablero Tasks-AI,
  Orchestrator, ACM), sin documento sustitutivo ni metodología paralela.
- EXPERIMENT-RESULTS v1.0.0: ejecución del experimento de viabilidad (30/30
  mediciones) e informe de resultados con clasificación **INCONCLUSO** (PRD §7,
  regla 2: medición no comparable por la incidencia del fixture de usuario,
  documentada en `experiment/runs/INCIDENCIA-FIXTURE.md`, opción B del PO). La
  decisión de producto de continuar con AIF asumiendo la incertidumbre residual
  no equivale a un GO experimental; Specification queda pendiente de encargo
  propio. Incluye datos brutos y auditados en `experiment/runs/` y el cálculo
  determinista de métricas en `experiment/metrics.py`.
- Documentación base del proyecto: PRD v0.1.4, RESEARCH v0.1.2 y protocolo de experimento
  v0.1.6, con la puerta GO/NO-GO en el PRD §7 y la guía operativa en EXPERIMENT-RUNBOOK.
- Estructura de entrada del repositorio: README con el estado del proyecto, AGENTS.md con
  el mapa de fuentes canónicas, CONTRIBUTING.md con la política Git y llms.txt como
  catálogo enlazable para agentes.
- Tablero tasks-ai en `.tasks-ai/`: instrucciones y contrato de etiquetas versionados; la
  base, el binario y la extensión quedan fuera de Git.
- Memoria de continuidad ACM montada sobre `.agent-memory/memory.db` con el alias
  `agent-interaction-forms` y el adaptador del ejemplo `agent-continuity`.
