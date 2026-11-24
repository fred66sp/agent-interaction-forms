# AIF — especificación del contrato de interacción, versión 0.1

| Metadato | Valor |
|---|---|
| ID | AIF-SPEC |
| Versión | 0.1.0 |
| Fecha | 2026-11-24 |
| Estado | **BORRADOR** — pendiente de aprobación del PO |
| Fuentes | [PRD.md](PRD.md) v0.1.4 · [RESEARCH.md](RESEARCH.md) v0.1.2 · [EXPERIMENT-RESULTS.md](EXPERIMENT-RESULTS.md) v1.0.0 |

## 0. Alcance, naturaleza y vocabulario normativo

AIF es un contrato de interacción entre una herramienta (el **ejecutor**) y un agente
de IA (el **consumidor**). Describe cómo el ejecutor comunica qué información necesita,
qué recibió correctamente, qué falta, qué es incorrecta y por qué, cómo corregirla,
cómo continuar y cuál fue el resultado real. AIF no es un framework de agentes: no
orquesta, no almacena, no autoriza y no ejecuta nada por sí mismo.

La especificación es **independiente del lenguaje, del modelo y del transporte**. La
única representación de intercambio es JSON (RFC 8259). La materialización primera es
el perfil CLI (§7); ningún perfil amplía el contrato del núcleo (§§1–6).

Los ejemplos (§8) ilustran el contrato y no introducen nada más allá de los §§2–6.
La trazabilidad con los requisitos del PRD está en §11.

Marcado. Las secciones §§1–7 y 9–10 son **normativas**; §8 (ejemplos) es normativo en
tanto los ejemplos ilustran exactamente este texto, y §§11–12 son **informativas**. Las
palabras clave normativas se escriben en mayúsculas, según RFC 2119: **DEBE** (obligación),
**NO DEBE** (prohibición), **PUEDE** (permiso), **MUST**/**MAY** como alias en inglés.

## 1. Modelo de interacción

La interacción es una secuencia de intercambios **sin estado para el contrato**: cada
petición debe contener todo lo que el ejecutor necesite; el ejecutor revalida siempre
(RF-08). El contrato no define sesiones, identificadores de formulario, expiración ni
correlación entre llamadas (§10).

Una **petición** es un objeto JSON con estos miembros:

| Miembro | Condición | Contenido |
|---|---|---|
| `aif` | DEBE | Versión del contrato que usa el consumidor. Literal `"0.1"` en esta especificación. |
| `operation` | PUEDE faltar solo para descubrimiento (§7, CL-05) | Identificador de la operación, textual. |
| `input` | PUEDE faltar | Objeto con los datos que el consumidor aporta, campo a campo. |

Una **respuesta a una operación** es un objeto JSON con estos miembros:

| Miembro | Condición | Contenido |
|---|---|---|
| `aif` | DEBE | Versión del contrato empleada para esta respuesta. |
| `operation` | DEBE | El identificador de la operación solicitada; salvo en `unsupported_version` (§3.5), donde no hay operación activa que reclamar. |
| `result` | DEBE | Uno de los resultados de §2. |

Cada resultado añade los miembros propios de §3. Los miembros no definidos aquí están
permitidos en las respuestas (extensiones del ejecutor) y el consumidor DEBE ignorar
los que no reconozca; el ejecutor NO DEBE colocar fuera de los miembros definidos
información de la que dependa decidir el siguiente paso. Los miembros desconocidos de
`input` hacen la petición inválida (§3.4): el ejecutor los rechaza de forma explícita
en lugar de ignorarlos.

El detector de versión es el único mecanismo de compatibilidad: el ejecutor que no
soporte la versión pedida responde `unsupported_version` (§3.5) y NO DEBE procesar
`input`. El ejecutor responde siempre con una versión que sigue produciendo
respuestas interpretables para el consumidor y la declara en `aif`.

## 2. Resultados posibles

| `result` | Significado | Comportamiento |
|---|---|---|
| `executed` | La operación se llevó a cabo. | ejecución directa |
| `input_needed` | Faltan datos obligatorios; nada se ejecutó. | información ausente |
| `invalid` | Hay datos incorrectos; nada se ejecutó. | información inválida |
| `denied` | El dominio niega la operación; nada se ejecutó. | operación denegada |
| `failed` | La ejecución no llegó a completarse o su estado es desconocido. | fallo de ejecución |
| `unsupported_version` | La versión del contrato no es soportada. | compatibilidad (§1) |
| `unknown_operation` | El identificador no corresponde a ninguna operación del ejecutor. | compatibilidad (§1) |

Reglas invariantes:

- R-01. El ejecutor DEBE fijar `result` según la tabla anterior y no DEBE declarar
  más de un sentido por respuesta: una respuesta afirma `executed` solo cuando la
  operación completó su efecto de dominio; en `input_needed`, `invalid` y `denied`
  **no** se produjo efecto de dominio alguno (RNF-05) y así debe poder comprobarse.
- R-02. La clasificación es determinista: ante idénticos `operation` e `input`,
  mismas reglas y mismo contexto de dominio, el ejecutor devuelve el mismo `result`
  y la misma descripción de campos (RNF-01). El determinismo no alcanza a valores de
  dominio como identificadores generados o relojes (RESEARCH §8).
- R-03. Únicamente el `result` y los miembros que §3 exige son fiables para decidir
  cómo continuar (`collected` incluido cuando la respuesta lo trae); ninguna prosa
  del ejecutor es parte del contrato ni sustituye a `result`. La interpretación de
  prosa libre es un error de conformidad (RF-06).

## 3. Contenido de cada resultado

### 3.1 `executed`

DEBE incluir `outcome`, cuyo contenido es territorio del dominio (identificador del
resultado, datos de retorno, mensaje). AIF no define su estructura interna: el
consumidor la trata como datos opacos del dominio o la usa según la documentación de
la operación. Continuar tras `executed` corresponde al dominio, no a AIF.

### 3.2 `input_needed` y `invalid` — campos

Ambos resultados describen los campos de la operación con el mismo vocabulario:

- `missing` — DEBE en `input_needed`, NO DEBE presente en `invalid`: matriz de
  nombres de campos obligatorios cuyo valor no llegó o llegó vacío.
- `invalid` — DEBE en `invalid`, NO DEBE presente en `input_needed`: objeto campo →
  descriptor del problema.
- `fields` — DEBE en `input_needed` y `invalid`: objeto campo → descriptor (§4) de,
  al menos, cada campo de `missing` y de `invalid`, con su obligatoriedad y sus
  restricciones; PUEDE describir campos adicionales (RF-03, CA-02).
- `collected` — PUEDE: objeto campo → valor; conserva los valores válidos **no
  sensibles** que el consumidor aportó en esta misma petición (RF-05, RNF-07).

El descriptor de un campo inválido en `invalid` DEBE listar:

| Miembro | Condición |
|---|---|
| `value` | DEBE: el valor recibido, salvo que sea sensible (§6). |
| `reason` | DEBE: por qué es inválido, en texto determinista derivado de la regla comprobada. |
| `allowed` | DEBE si el campo tiene un conjunto cerrado de valores: los valores permitidos. |

`missing` e `invalid` son excluyentes por campo: un campo no figura en ambos (RF-04,
distinguir ausencia de invalidez). Un campo opcional con valor inválido figura en
`invalid`; nada obliga al ejecutor a validarlo, pero, si lo comprueba y falla, DEBE
reportarlo igual que un obligatorio.

`reason` y las reglas que citan valores permitidos son repos de comprobación
mecánica: no es prosa interpretativa. La causa sigue siendo la regla del ejecutor,
no un análisis del signo del texto (RESEARCH, lección 3 del experimento).

### 3.3 `denied`

DEBE incluir `reason`: texto del dominio que explica la denegación. NO DEBE incluir
`missing`, `invalid`, `fields` ni `collected`: una denegación no es un trámite
reparable y AIF no concede permisos ni da instrucciones para eludirlos (RF-11).
Un motivo tan genérico que no explique nada cumple la letra y fracasa en el
propósito (RF-07): el ejecutor DEBE poder dar alguna razón, aun sin revelar
detalles sensibles.

### 3.4 `failed`

DEBE incluir `message`: qué falló y qué estado dejó el dominio si el ejecutor puede
conocerlo (completado parcial, desconocido, no iniciado). PUEDE incluir `code`, que
pertenece al dominio. El ejecutor NO DEBE reportar `executed` si el efecto no se
completó (RF-07); dudar es informar, no mentir.

Un error de interacción nunca llega como `failed`: entrada ausente es
`input_needed`, entrada inválida es `invalid` (§3.2). `failed` reserva su sentido
para errores del dominio o de infraestructura en la ejecución. Un `input` con
miembros desconocidos para la operación es `invalid`, no `failed`. Si esa misma
petición combina miembros desconocidos con campos ausentes o inválidos, el rechazo
estructural precede: el `result` es `invalid` por los miembros desconocidos y el
ejecutor PUEDE informar además de los campos.

### 3.5 `unsupported_version` y `unknown_operation`

`unsupported_version` DEBE incluir `supported`: matriz de las versiones del contrato
que el ejecutor soporta, y NO DEBE procesar `input`; su respuesta prescinde de
`operation` (§1), porque el contrato que podría dar sentido a ese identificador no
está activo. Es también la respuesta ante una entrada que el ejecutor no puede
interpretar como petición conforme —no es JSON válido, no es un objeto, carece de
`aif` o lleva claves duplicadas—: el contrato no queda establecido y por ello la
petición no se procesa; esta política es el rechazo explícito que RESEARCH §8 exige
para v0.1, y el ejecutor que quiera un diagnóstico más fino lo documenta fuera del
contrato. `unknown_operation` DEBE incluir `known_operations`: matriz de los
identificadores de operación que el ejecutor atendería, y DEBE hacer eco de
`operation` con el identificador solicitado. Ambos describen un problema del
contrato, no del dominio: sin efecto de dominio y sin campo alguno que reparar.

## 4. Descripción de campos

El conjunto mínimo para describir un campo de una operación:

| Miembro | Clase | Sentido |
|---|---|---|
| `required` | imprescindible | `true` si el ejecutor no puede ejecutar sin este valor. |
| `value` | contextual | Valor recibido del consumidor para este campo en esta petición. |
| `reason` | contextual | Por qué el valor es inválido; solo dentro de `invalid`. |
| `type` | opcional | Uno de `string`, `number`, `boolean`, `object`, `array`. |
| `allowed` | opcional | Matriz de valores permitidos, cuando el conjunto es cerrado. |
| `description` | opcional | Indicación contextual en texto, sin sustituir a reglas mecánicas (§2, R-03). |

Reglas: los miembros `required`, `type` y `description` describen el mismo campo con
la misma semántica en cualquier respuesta en que aparezcan, y una única fuente de
verdad por operación dentro del ejecutor (DRY). Ninguna otra propiedad pertenece al
núcleo de v0.1: restricciones finas (patrones, rangos, dependencias entre campos)
quedan en manos del dominio, que es quien valida (§6); si el ejecutor quiere
publicarlas, lo hace en `description` o como extensión tolerada (§1). La ausencia
de `type` o `allowed` no autoriza a aceptar cualquier valor: es una limitación de la
descripción, no de la validación del dominio.

## 5. Continuación

La continuación es implícita al resultado y explícita en su contenido:

- C-01. Tras `input_needed` o `invalid`, el consumo siguiente válido es reenviar la
  misma operación (§1) con un `input` completo: los valores de `collected` (§3.2),
  los valores que el consumidor añada y una corrección para cada campo de `invalid`.
- C-02. El consumidor puede tratar `collected` como copia fiable de lo que ya
  aportó; todo lo demás de la respuesta (§4 `fields`, `description`) es informativo
  para que el consumidor elija valores, no datos que reenviar.
- C-03. El ejecutor NO DEBE asumir estado previo: revalida el `input` completo en
  cada petición (RF-08). El `result` de la ronda anterior no da ningún derecho.
- C-04. Tras `denied`, `failed`, `unsupported_version` o `unknown_operation`, AIF no
  define continuación: quien decide el paso siguiente es el consumidor con la
  información de §3.3/§3.4/§3.5 y la documentación del dominio.
- C-05. Ninguna continuación implica garantía de idempotencia, autenticidad ni
  ejecución única (PRD RF-14: no prometer garantías de reintento inexistentes; el
  mecanismo idempotente de RESEARCH P05 queda diferido a v0.2): reenviar es una
  petición nueva, con el efecto que el dominio determine. El consumidor que necesite
  esas garantías debe obtenerlas del dominio, no de AIF.

El camino directo (RF-02) es la regla constitutiva de la interacción: un consumidor
que ya dispone de todos los datos invoca la operación y recibe `executed` sin
ningún intercambio previo impuesto por AIF.

## 6. Límites de responsabilidad

AIF describe la interacción; el ejecutor y su dominio retienen toda la autoridad:

| Quién | Retiene |
|---|---|
| Ejecutor | Reglas de negocio, validación efectiva de los datos, autorización y permisos, ejecución, persistencia, integridad del dominio, transacciones y cualquier garantía de idempotencia que exista. |
| AIF | Únicamente el vocabulario de §1–§5 para que esa autoridad se comunique. |

Consecuencias normativas:

- L-01. Los datos de una respuesta AIF no son autorización ni evidencia de ejecución;
  toda continuación es revalidada por el ejecutor (RF-08, RF-11, RNF-06).
- L-02. Entrada inválida o denegada no produce efecto de dominio (RNF-05) y el
  resultado debe permitir comprobarlo (R-01).
- L-03. El ejecutor NO DEBE devolver en la respuesta valores sensibles, sea cual
  sea el miembro: ni en `collected`, ni en `value`, ni en `reason`, ni en `message`,
  ni en `outcome`, ni en `description` (RNF-07). Sensibilidad la declara el dominio;
  el ejecutor PUEDE omitir cualquier valor por sensibilidad u otra prudencia, y la
  omisión no significa invalidez.
- L-04. El ejecutor DEBE documentar los límites de tamaño y profundidad de su
  `input` (RNF-03): la especificación no fija números, exige que se declaren.
- L-05. El contenido de `outcome`, `reason` y `message` es territorio del dominio;
  AIF solo exige su presencia y su honestidad (§3).

## 7. Perfil CLI

Primera materialización del contrato. Normativa para quien declare conformidad AIF
sobre una herramienta de línea de órdenes; el núcleo (§§1–6) no cambia.

- CL-01. No interactivo: el ejecutor NO DEBE abrir diálogo con el humano, bloquear
  la entrada esperándolo ni formular preguntas (RF-12, RESEARCH §8).
- CL-02. La respuesta completa es exactamente un objeto JSON del §1, con salto de
  línea final, en `stdout`. Ningún texto acompaña al JSON en `stdout`; los
  diagnósticos, ayudas y avisos van en `stderr` en prosa (RF-12).
- CL-03. La entrada es un objeto JSON del §1: PUEDE llegar por argumentos, por
  `stdin` o por fichero, a criterio del ejecutor; si llega por `stdin`, es un único
  objeto JSON con la forma de petición y nada más.
- CL-04. Códigos de salida:

| Código | `result` |
|---|---|
| `0` | `executed` y respuestas de descubrimiento (CL-05) |
| `2` | `input_needed`, `invalid` |
| `3` | `unknown_operation`, `unsupported_version` |
| `4` | `denied` |
| `1` | `failed` |

  El código solo depende del `result`, de forma determinista (RNF-01); `stderr` no
  altera el código.
- CL-05. Descubrimiento: invocado sin `operation`, el ejecutor responde con un
  objeto JSON `{"aif":"0.1","operations":[{"id":..., "description": ...}]}` con una
  entrada por operación y su indicación de una línea, y código 0. Esta respuesta no
  es una respuesta de operación: no lleva `result` ni `operation`.
- CL-06. Continuación: reenviar la petición completa (§5) — en CLI, la nueva
  invocación con el `input` íntegro corregido. El ejecutor no reserva estado entre
  invocaciones para el contrato.

## 8. Ejemplos normativos

Operación de ejemplo: `record_observation`, con campos `title` (cadena, obligatorio),
`kind` (enum obligatorio: `learning`, `decision`, `discovery`, `bugfix`) y
`confidence` (enum obligatorio: `hypothesis`, `confirmed`). Los valores de dominio
(`outcome`) son ilustrativos. «←» marca el paso que el ejemplo ilustra.

### 8.1 Ejecución directa (§5)

```json
{"aif":"0.1","operation":"record_observation",
 "input":{"title":"Perfil CLI decidido","kind":"decision","confidence":"confirmed"}}
```

```json
{"aif":"0.1","operation":"record_observation","result":"executed",
 "outcome":{"id":"obs_0142","recorded_at":"2026-11-24T19:41:00Z"}}
```

### 8.2 Campo obligatorio ausente (§3.2, `input_needed`)

```json
{"aif":"0.1","operation":"record_observation",
 "input":{"kind":"learning"}}                                  ← faltan title y confidence
```

```json
{"aif":"0.1","operation":"record_observation","result":"input_needed",
 "missing":["title","confidence"],
 "fields":{"title":{"required":true,"type":"string",
                   "description":"Título breve de la observación"},
           "confidence":{"required":true,"type":"string",
                         "allowed":["hypothesis","confirmed"]}},
 "collected":{"kind":"learning"}}
```

### 8.3 Continuación tras completar el campo (§5, C-01)

```json
{"aif":"0.1","operation":"record_observation",
 "input":{"title":"Perfil CLI decidido","kind":"learning","confidence":"hypothesis"}}
```

```json
{"aif":"0.1","operation":"record_observation","result":"executed",
 "outcome":{"id":"obs_0143","recorded_at":"2026-11-24T19:42:30Z"}}
```

### 8.4 Valor inválido de una enumeración (§3.2, `invalid`)

```json
{"aif":"0.1","operation":"record_observation",
 "input":{"title":"Perfil CLI decidido","kind":"Learning","confidence":"confirmed"}}
```

```json
{"aif":"0.1","operation":"record_observation","result":"invalid",
 "invalid":{"kind":{"value":"Learning",
                    "reason":"valor no permitido",
                    "allowed":["learning","decision","discovery","bugfix"]}},
 "collected":{"title":"Perfil CLI decidido","confidence":"confirmed"}}
```

### 8.5 Corrección conservando los valores válidos (§3.2, §5)

El consumidor reenvía `collected` (§5, C-02) con el campo corregido:

```json
{"aif":"0.1","operation":"record_observation",
 "input":{"title":"Perfil CLI decidido","kind":"decision","confidence":"confirmed"}}
```

```json
{"aif":"0.1","operation":"record_observation","result":"executed",
 "outcome":{"id":"obs_0144","recorded_at":"2026-11-24T19:44:00Z"}}
```

### 8.6 Operación denegada (§3.3)

```json
{"aif":"0.1","operation":"record_observation",
 "input":{"title":"Sin permiso","kind":"decision","confidence":"confirmed"}}
```

```json
{"aif":"0.1","operation":"record_observation","result":"denied",
 "reason":"el consumidor no tiene permiso de escritura sobre el dominio"}
```

### 8.7 Fallo de ejecución (§3.4)

```json
{"aif":"0.1","operation":"record_observation",
 "input":{"title":"Fallo simulado","kind":"discovery","confidence":"hypothesis"}}
```

```json
{"aif":"0.1","operation":"record_observation","result":"failed",
 "message":"el almacenamiento rechazó la escritura; estado desconocido, la observación no queda registrada",
 "code":"storage_unavailable"}
```

### 8.8 Versión incompatible (§3.5)

```json
{"aif":"1.0","operation":"record_observation","input":{"title":"X","kind":"learning","confidence":"confirmed"}}
```

```json
{"aif":"0.1","result":"unsupported_version","supported":["0.1"]}
```

Nótese que un `unsupported_version` no reclama la operación solicitada: la respuesta
lleva `aif` y `supported`, y prescinde de `operation` porque el contrato que podría
darle sentido no está activo (§1, §3.5).

## 9. Conformidad

Una implementación es conforme con AIF v0.1 cuando cumple todo lo normativo de
§§1–9: las reglas R-01–R-03 (§2), el contenido exigible por resultado (§3), el
vocabulario de campos (§4), la continuación (§5), los límites de responsabilidad
(§6) y, si declara conformidad de perfil CLI, §7. Los casos de conformidad son una
matriz de pruebas deterministas, ejecutables sin LLM ni red (RF-13, RNF-02), cada
uno con: petición, precondición de dominio y salidas afirmadas (`result`, miembros
obligatorios ausentes/presentes, código de salida del perfil CLI y ausencia de
efecto de dominio verificada en el estado previo y posterior).

| Caso | Petición | Precondición de dominio | Afirma | PRD |
|---|---|---|---|---|
| CJ-01 | §8.1 | acceso concedido | `executed`, `outcome` presente, salida 0, efecto de dominio exactamente el esperado (una escritura) | CA-01, RF-01/02 |
| CJ-02 | §8.2 | acceso concedido | `input_needed`, `missing` lista exacta, sin `invalid`, `fields` describe obligatoriedad y restricciones de los campos de `missing`, `collected` conserva no sensibles, salida 2, cero efecto de dominio | CA-02, RF-03/04 |
| CJ-03 | §8.4 | acceso concedido | `invalid` con `reason` citando regla comprobada y `allowed` íntegro; sin `missing`; un reenvío §8.5 produce `executed`; salida 2 luego 0 | CA-03, RF-04/05/06 |
| CJ-04 | §8.6 | permiso retirado | `denied` con `reason`, sin miembros de reparación, salida 4, cero efecto de dominio; ídem con `input` manipulado para simular permisos → sigue `denied` | CA-04, RF-07/11 |
| CJ-05 | §8.1 repetida, mismo dominio | acceso concedido | mismo `result` y mismos miembros de decisión; el JSON de `stdout` interpreta sin entrada interactiva | CA-05, RNF-01 |
| CJ-06a | §8.7 | fallo forzado del dominio | `failed`, `message` presente, salida 1, nunca `executed` | CA-06, RF-07 |
| CJ-06b | §8.8 | — | `unsupported_version` con `supported`, `input` no procesado, salida 3 | CA-06, RNF-08 |
| CJ-07 | §8.2 con un valor sensible en `input` | acceso concedido | ese valor NO DEBE aparecer en ninguna parte de la respuesta (RNF-07); sigue `input_needed` | RNF-07 |
| CJ-08 | `input` con miembro desconocido | acceso concedido | `invalid` con `reason` identificando el miembro, salida 2, cero efecto de dominio | §3, RNF-05 |
| CJ-09 | invocación sin `operation` (perfil CLI) | — | respuesta de descubrimiento, código 0, cada operación con `id` | CL-05, RF-12 |
| CJ-10 | `operation` desconocida para el ejecutor | — | `unknown_operation`, `known_operations` presente, eco de `operation`, salida 3, cero efecto de dominio | §3.5, RF-01 |
| CJ-11 | entrada no interpretable (no JSON, sin `aif`, claves duplicadas) | — | `unsupported_version` con `supported`, `input` no procesado, salida 3, cero efecto de dominio | §3.5, RNF-08 |

Los fixtures de conformidad se derivarán de esta tabla como ficheros independientes
del lenguaje (petición, precondición, aserciones), al pasar a la fase de
implementación; definirlos aquí sería diseñar el formato antes del SDD, que está
fuera del alcance de este documento.

## 10. Fuera del contrato de v0.1

No pertenece a AIF v0.1 y su presencia no es requisito de conformidad: sesiones,
persistencia de formularios, identificadores de formulario, expiración, revisiones,
idempotencia como garantía del núcleo, reintentos, control de concurrencia,
orquestación de agentes, servidores MCP, perfiles HTTP, interfaces gráficas,
generadores de código, SDK, bibliotecas. Las garantías que un ejecutor ofrezca por
su cuenta (como las de ACM) permanecen en el ejecutor y no se generalizan ni niegan.

## 11. Trazabilidad con el PRD

| Requisito del PRD | Dónde se satisface |
|---|---|
| RF-01 | §1, §2, §3.5 |
| RF-02 | §5, §8.1 |
| RF-03 | §3.2 (`fields` DEBE), §8.2 |
| RF-04 | §3.2, §8.2/§8.4 |
| RF-05 | §3.2, §8.4/§8.5 |
| RF-06 | §2 (R-03), §5, §7 (CL-05) |
| RF-07 | §2, §3.3, §3.4 |
| RF-08 | §1, §5 (C-03), §6 (L-01) |
| RF-11 | §3.3, §6 |
| RF-12 | §7 |
| RF-13 | §9 |
| RNF-01/02 | §2 (R-02), §9 |
| RNF-03 | §6 (L-04) |
| RNF-05 | §2 (R-01), §6 (L-02) |
| RNF-06 | §6 (L-01) |
| RNF-07 | §6 (L-03), §3.2, CJ-07 |
| RNF-08 | §3.5, §8.8, CJ-06b |
| RNF-10 | §4, §6 |
| CA-01..CA-06 | §9 (matriz CJ) |

## 12. Decisiones de este borrador (informativo)

1. Un único sobre JSON uniforme para todos los resultados (§1) en lugar de formatos
   distintos por estado: una sola forma que aprender, `result` como discriminante.
2. Siete resultados explícitos en lugar de códigos de error anidados: cada uno con
   contenido exigible y comprobable (§3); `unsupported_version` lo exige RNF-08 y
   `unknown_operation` es la lectura contractual de RF-01.
3. `missing` y `invalid` como miembros excluyentes (§3.2) frente a un único mapa con
   estados: separa ausencia de invalidez (RF-04) sin introducir una propiedad de estado.
4. Continuación stateless (§5) frente a identificadores de formulario: conserva el
   alcance del PRD (sin sesiones, D-05) y hace la revalidación explicable (RF-08).
5. Metadatos de campo mínimos con jerarquía imprescindible/contextual/opcional (§4):
   `type` y `allowed` opcionales; describir más es territorio del dominio o extensión.
6. Códigos de salida del perfil CLI fijados en función del `result` (§7, CL-04);
   los valores 0/2/3/4 coinciden con los ya ejercitados por el experimento para
   resultados equivalentes, y `failed`→1 es una decisión nueva sin antecedente
   ejercitado (el experimento usaba 9 para el error interno del ejecutor).
7. Descubrimiento como respuesta aparte del sobre (§7, CL-05): evita un octavo
   `result` y una operación fantasma.
8. Miembros desconocidos de `input` rechazados de forma explícita (§1, §3.4, CJ-08):
   el silencio ante nombres errados ocultaría errores del consumidor y rompería el
   determinismo de la decisión.
9. Cadena de autorización: el PRD §7 condicionaba Specification a un GO explícito;
   el PO clasificó el experimento como INCONCLUSO (opción B), decidió continuar
   asumiendo la incertidumbre residual y autorizó el inicio de Specification. Este
   borrador nace de esa autorización y queda a la espera de su aprobación.
10. Entrada no interpretable elevada a `unsupported_version` (§3.5, CJ-11): reutiliza
   un resultado existente en lugar de crear un octavo y cumple el rechazo explícito
   que RESEARCH §8 pide para JSON ambiguo y claves duplicadas.
