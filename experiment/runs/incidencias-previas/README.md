# Incidencias de infraestructura previas a la medición (2026-11-24)

Tres tentativas de la ejecución 1 (acm, completa, par 1) abortadas antes de
producir salida del modelo; la base quedó limpia, sin efectos de dominio.
No consumen la muestra: son reparación del instrumento.

1. `XDG` aislado sin runtime de Node: opencode intentó descargar Node (sin red).
2. Con caché real (`XDG_CACHE_HOME=/home/desarrollo/.cache`): «Package not
   found in workspace» — el shim `opencode` del PATH es el conmutador de
   vite-plus (vp), que bajo el entorno aislado no delega (sin vars OPENCODE_*).
3. Con `package.json` en el LAB: vp registró los flags como tareas propias
   («Task "--model" not found»). Binarios verificados: `vp` (Rust) primero en
   PATH; opencode real v2.2.6 en `~/.opencode/bin`; v2.4.1 en
   `vite-plus/packages/@opencode/cli/<hash>/bin/opencode` (la versión validada
   en tk-2). Resolución: invocación del binario 2.4.1 por ruta absoluta.
