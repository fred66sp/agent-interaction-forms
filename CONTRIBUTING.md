# Flujo de trabajo Git

Política de este repositorio para ramas, commits, verificación, integración y CHANGELOG.
Prevalece sobre cualquier modelo genérico.

## Modelo

- `main` es la única rama de larga vida; contiene lo que está listo para usarse.
- No hay `develop`, `release/*` ni `hotfix/*`: introducirlos exige cambiar esta política.
- El commit se hace antes de cerrar la tarea en el tablero.

## Ramas

Se crean desde `main` y viven lo mínimo: nombres en minúsculas, palabras separadas con
guiones.

| Patrón | Uso |
|---|---|
| `feature/<slug>` | Funcionalidad nueva |
| `fix/<slug>` | Corrección de un defecto |
| `docs/<slug>` | Documentación |
| `refactor/<slug>` | Reorganización sin cambio de comportamiento |
| `chore/<slug>` | Mantenimiento, herramientas, configuración |

Una tarea pequeña puede ir directamente a `main`. Es pequeña cuando el diff se entiende
leyéndolo, se deshace revirtiendo el commit y no toca datos almacenados, contratos
públicos, seguridad, concurrencia ni código ejecutable en varios módulos, ni mezcla
cambios ajenos. Ante la duda, rama.

Crear una rama local no integra ni publica nada.

## Commits

- Conventional Commits, en inglés: `feat`, `fix`, `docs`, `chore`, `refactor`, `test`,
  `style`, `perf`, `ci`, `build`, `revert`.
- Descripción breve, en imperativo y sin punto final.
- Un commit representa una unidad coherente. Varios solo cuando separan cambios
  independientes o hacen la revisión más clara.
- El cuerpo nombra la tarea que sostiene el commit con `Refs: tk-<n>`, una por tarea. Es
  la única atadura entre tarea y commit; el tablero no guarda hashes:

```bash
git log --format='%h %s' --grep='Refs: tk-1'
```

- Todo cambio de código lleva en el cuerpo una línea de riesgo, `Risk: item N (motivo)` o
  `Risk: none`. Los ítems: 1 datos o efectos irreversibles; 2 seguridad; 3 contratos que
  otros consumen; 4 concurrencia; 5 entrega o entorno; 6 nada que detecte una regresión en
  lo que cambia. Un riesgo alto añade una verificación independiente: otro agente lee el
  criterio de la tarea y lo prueba contra el resultado real.

- Ningún commit con cabos sueltos: lo que se descubre al prepararlo se arregla o se
  pregunta antes, no en un commit de corrección inmediata.
- Sin firmas de IA ni líneas `Co-Authored-By` salvo petición expresa.
- Ningún commit se crea sin confirmación explícita del PO: cubre lo presentado, y un
  cambio de alcance pide otra.

## Verificación

La verificación es proporcional al cambio y se contrasta con el criterio del PRD: los
criterios de aceptación de la PoC ([`PRD.md`](docs/PRD.md) §6) y la puerta del experimento
([`PRD.md`](docs/PRD.md) §7). Mientras el repositorio no tenga código, la comprobación es
documental: cada enlace relativo resuelve y ninguna afirmación contradice a PRD, RESEARCH
o EXPERIMENT. Toda verificación no ejecutada se declara.

Cuando exista código ejecutable, su batería se fija aquí antes de la primera entrega; no
se improvisa en el commit. La PoC no se declara correcta sin que sus fixtures pasen al
100 % según [PRD.md](docs/PRD.md) §6.

## Integración

- Se integra en `main` con merge no fast-forward: `git merge --no-ff <rama>`.
- Sin `squash` ni `rebase` sobre ramas ya publicadas.
- Antes: la tarea terminada, las comprobaciones correctas, ningún cambio ajeno que se
  pierda y la rama correcta.
- Después: se comprueba estado, rama y que el commit está en `main`.
- Una rama con trabajo no integrado no se borra sin autorización.

## CHANGELOG

- Se anota lo que tiene valor para quien usa el proyecto, al publicar una versión, bajo
  `## [Sin publicar]`, con las secciones `Añadido`, `Modificado` y `Corregido`.
- No se actualiza por cada commit ni por cambios internos de prosa.
- Un cambio que altere requisitos, alcance o la puerta del experimento actualiza el
  CHANGELOG antes del commit.

## Versiones y etiquetas

- Este repositorio no tiene versión publicada ni proceso de release. Las versiones de
  PRD, RESEARCH y EXPERIMENT viven en la cabecera de cada documento y no son etiquetas de
  Git.
- Cuando exista algo publicable, el versionado será semántico: `PATCH` para una corrección
  compatible, `MINOR` para funcionalidad compatible y `MAJOR` para un cambio incompatible.
- La etiqueta será anotada, con prefijo `v`, y marcará el commit integrado.
- Crear una etiqueta, preparar una release o publicar exige autorización explícita; y
  antes hay que resolver la decisión D-07 del PRD (licencia, nombre, gobernanza y
  publicación).

## Operaciones que exigen autorización

- Fusionar en `main`.
- `push`, `fetch` para integrar y cualquier otra operación remota.
- Crear una etiqueta o publicar una versión.
- `reset`, `clean`, `rebase`, `filter-repo` u otra operación que elimine o reescriba
  trabajo.
- Sobrescribir, descartar u ocultar cambios de la persona.
- Borrar una rama con trabajo no integrado.

Una autorización cubre lo que nombra y no se hereda a otra operación.

## Principio

Git aporta seguridad y trazabilidad sin convertirse en una carga. El orden de prioridad:
no perder trabajo, entregar algo verificable, mantener un historial comprensible, pedir
autorización para lo persistente y no añadir ramas, commits ni comprobaciones que no
aporten.
