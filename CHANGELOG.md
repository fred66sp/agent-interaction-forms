# Registro de cambios

Cambios con valor para quien usa el proyecto, por versión. El formato sigue las
recomendaciones de *Keep a Changelog* y el versionado es semántico. Lo que se anota aquí
al publicar; la historia de cada cambio, commit a commit, vive en Git.

## [Sin publicar]

### Protocolo AIF 0.2 (evolución autorizada; no es una publicación)

- SPECIFICATION v0.2.0 (**APROBADO** como evolución autorizada por el PO): única
  versión soportada, `"0.2"`; la validación puede ser progresiva o exhaustiva y el
  consumidor no debe suponer que `invalid` enumera todos los problemas detectables;
  `collected` conserva lo aportado en la petición sin certificar su validez;
  `invalid` es estructurado (`value`, `reason`, `allowed`, `schema`); las respuestas
  reparables (`input_needed` e `invalid`) no producen efecto de dominio; el dominio
  valida por completo antes de cualquier efecto; continuación sin estado; y `schema`
  admite fragmentos JSON Schema 2020-12 para las reglas que `type` y `allowed` no
  expresan. v0.1.1 queda como contrato histórico implementado. La autorización no
  aprueba una publicación.
- Biblioteca `aif` actualizada al contrato 0.2: `InvalidError` estructurado que el
  dominio devuelve antes de producir efectos y que el ejecutor recompone en
  `invalid` —el dominio no compone `collected` ni la respuesta—, validación
  exhaustiva por intercambio con precedencia determinista de `invalid`, omisión de
  valores sensibles en toda la respuesta, rechazo de versiones distintas de `"0.2"`
  sin procesar `input`, y caso de conformidad CJ-14 para la regla de dominio
  aplicada antes del efecto.
- README, AGENTS.md y llms.txt alineados con el versionado vigente 0.2, el fin de la
  congelación únicamente para esta evolución autorizada y el carácter histórico de
  v0.1.1.
- Descubrimiento completo: la respuesta sin `operation` incluye `request`, la
  plantilla de emisión `{"aif":"0.2","operation":"<id>","input":{}}`, además de
  `operations`; CJ-09 la exige y SPECIFICATION CL-05 la describe.

### Documentación

- docs: consolidación conceptual y documental de AIF según el encargo del PO —
  `docs/PRODUCT.md` v0.1 incorporado como fuente canónica de visión; README
  simplificado como puerta de entrada (producto antes que detalles técnicos);
  `docs/web` reorganizada en cuatro niveles (Comprender/Utilizar/Referencia
  técnica/Investigación) con diagramas; `AGENTS.md` y `llms.txt`
  alineados por referencia; tabla de fuentes de SPECIFICATION actualizada
  (contenido normativo intacto; contrato v0.1 congelado).
- docs: alinear `docs/PRD.md` (v0.1.6) con la visión aprobada en `docs/PRODUCT.md`
  v0.1 — el PRD separa base técnica implementada, visión del producto completo y
  capacidades pendientes de diseño (AIF CLI como producto y AIF Definition), sin
  duplicar la visión, sin alterar requisitos, decisiones D-01..D-09, puerta §7 ni
  secuencia, y deja señalada para el PO la decisión de encaje del alcance v0.1
  cerrado con las modalidades futuras.

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
