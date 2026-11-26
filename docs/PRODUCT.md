# AIF — Agent Interaction Forms

**Visión de producto · v0.1 · Aprobada por el PO**

*Herramientas que saben interactuar con agentes de inteligencia artificial.*

## 1. Qué es AIF

AIF es una solución que permite dotar a las herramientas de software de una interfaz de interacción autodescriptiva, determinista y orientada a agentes de inteligencia artificial.

Su finalidad es que un agente pueda utilizar una herramienta sin tener que descubrir constantemente cómo funciona, qué información necesita o cómo corregir una petición incorrecta.

La propia herramienta se lo comunica de forma estructurada.

## 2. El problema

Las herramientas tradicionales están diseñadas principalmente para ser utilizadas por personas.

Cuando un agente necesita utilizarlas, frecuentemente tiene que consultar documentación, interpretar mensajes de error, descubrir parámetros, reconstruir peticiones y realizar invocaciones adicionales.

Esto introduce incertidumbre, consume contexto y aumenta el esfuerzo necesario para completar una operación.

AIF propone que sea la herramienta, y no el agente, quien tenga la responsabilidad de describir sus necesidades de interacción.

## 3. Cómo funciona

El funcionamiento se basa en tres situaciones habituales:

- **Información completa:** la herramienta ejecuta directamente la operación.
- **Información incompleta:** la herramienta devuelve un formulario contextual indicando qué información necesita.
- **Información incorrecta:** la herramienta identifica los errores, conserva los datos válidos no sensibles e indica cómo continuar.

El formulario no es una interfaz gráfica. Es una respuesta estructurada que el agente puede interpretar directamente.

### Ejemplo

Un agente solicita crear una tarea titulada «Revisar documentación».

La herramienta necesita conocer también su prioridad. En lugar de devolver un error genérico, comunica al agente:

- Título recibido: Revisar documentación.
- Información pendiente: prioridad.
- Valores admitidos: baja, media y alta.

El agente completa la información y vuelve a solicitar la operación.

Si hubiera proporcionado todos los datos correctamente desde el principio, la operación se habría ejecutado en una única invocación.

**Principio fundamental: una necesidad, una invocación siempre que la información disponible sea suficiente.**

## 4. Modalidades de utilización

AIF contempla dos modalidades complementarias.

### Integración nativa

La biblioteca AIF se incorpora directamente al código fuente de una aplicación.

La aplicación adquiere las capacidades de interacción de AIF sin necesitar un ejecutable intermediario.

Por ejemplo, Tasks-AI podría incorporar la biblioteca dentro de su propio ejecutable Go, conservando íntegramente su lógica de negocio.

### Integración externa

Un ejecutable AIF actúa como intermediario entre el agente y una herramienta existente.

Un archivo de definición describe las operaciones disponibles, la información necesaria, sus restricciones y cómo invocar la herramienta.

Esto permite adaptar herramientas sin modificar su código fuente, siempre que dispongan de una interfaz invocable y sus resultados puedan interpretarse de forma fiable.

Ambas modalidades comparten el mismo núcleo y el mismo contrato de interacción.

## 5. Componentes del producto

| Componente | Finalidad |
|---|---|
| AIF Core | Biblioteca reutilizable que proporciona las capacidades de interacción. |
| AIF CLI | Ejecutable que permite utilizar AIF como capa externa sobre herramientas existentes. |
| AIF Definition | Formato declarativo para describir operaciones e integrar herramientas. |
| AIF Specification | Contrato común de comunicación entre agentes y herramientas. |

Estos componentes forman parte de un único producto. No son proyectos independientes.

## 6. Principios de diseño

- **Simplicidad:** incorporar AIF debe requerir el menor esfuerzo posible.
- **Agent-first:** la interacción está diseñada prioritariamente para agentes.
- **Autodescripción:** la herramienta comunica lo que necesita en cada situación.
- **Determinismo:** las reglas de interacción no dependen de un modelo de inteligencia artificial.
- **Eficiencia:** evitar invocaciones, consultas y procesamiento innecesarios.
- **Independencia:** no depender de un proveedor, modelo o plataforma de agentes.
- **Responsabilidad:** la herramienta conserva la autoridad sobre sus operaciones, permisos y reglas de negocio.
- **Reutilización:** ambas modalidades comparten las mismas capacidades fundamentales.

## 7. Qué no es AIF

AIF no es un agente de inteligencia artificial, un orquestador, un gestor de workflows ni un sistema de formularios gráficos.

Tampoco sustituye la lógica de negocio de las aplicaciones.

Su responsabilidad se limita a facilitar la interacción entre agentes y herramientas.

## 8. Estado actual

El proyecto dispone de una primera base técnica funcional:

| Componente | Estado |
|---|---|
| AIF Core | Implementación inicial disponible en Go. |
| AIF Specification | Contrato v0.1 definido y aprobado. |
| CLI de demostración | Disponible. |
| Pruebas de conformidad | Disponibles. |
| Prueba de concepto con ACM | Realizada. |
| AIF CLI como producto | Pendiente de diseño. |
| AIF Definition | Pendiente de diseño. |

La base técnica existente permite continuar hacia el producto descrito en este documento, pero todavía no constituye el producto completo.

Los beneficios cuantitativos en eficiencia y fiabilidad frente a otras interfaces todavía no han quedado demostrados experimentalmente.

## 9. Objetivo del producto

Conseguir que cualquier herramienta compatible pueda ofrecer una experiencia de interacción preparada para agentes de inteligencia artificial, mediante una integración nativa o una configuración declarativa, con el mínimo esfuerzo de adopción.

El éxito de AIF no debe medirse por la sofisticación de su arquitectura, sino por la facilidad con la que permite utilizar herramientas de forma fiable.

---

**AIF: la herramienta conoce sus necesidades y el agente sabe cómo continuar.**
