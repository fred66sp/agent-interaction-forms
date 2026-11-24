# Registro de cambios

Cambios con valor para quien usa el proyecto, por versión. El formato sigue las
recomendaciones de *Keep a Changelog* y el versionado es semántico. Lo que se anota aquí
al publicar; la historia de cada cambio, commit a commit, vive en Git.

## [Sin publicar]

### Añadido

- Documentación base del proyecto: PRD v0.1.4, RESEARCH v0.1.2 y protocolo de experimento
  v0.1.6, con la puerta GO/NO-GO en el PRD §7 y la guía operativa en EXPERIMENT-RUNBOOK.
- Estructura de entrada del repositorio: README con el estado del proyecto, AGENTS.md con
  el mapa de fuentes canónicas, CONTRIBUTING.md con la política Git y llms.txt como
  catálogo enlazable para agentes.
- Tablero tasks-ai en `.tasks-ai/`: instrucciones y contrato de etiquetas versionados; la
  base, el binario y la extensión quedan fuera de Git.
- Memoria de continuidad ACM montada sobre `.agent-memory/memory.db` con el alias
  `agent-interaction-forms` y el adaptador del ejemplo `agent-continuity`.
