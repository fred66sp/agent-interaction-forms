#!/usr/bin/env bash
# Lanza el servidor estático de la documentación web (docs/web) en 127.0.0.1:8891.
# Servicio local, solo para revisión en destino; pid y log en .run/. Se detiene con
# `./serve-web.sh stop`.
set -u
cd "$(dirname "$0")"
PORT=8891
PIDFILE=.run/serve-web.pid
LOG=.run/serve-web.log

if [ "${1:-}" = "stop" ]; then
  if [ -f "$PIDFILE" ]; then
    PID=$(cat "$PIDFILE")
    if ps -p "$PID" -o command= 2>/dev/null | grep -q 'http.server'; then
      kill "$PID" && rm -f "$PIDFILE" && echo "detenido"
    else
      rm -f "$PIDFILE" && echo "pidfile obsoleto (proceso $PID ya no es http.server); nada que detener"
    fi
    exit 0
  fi
  echo "no hay servidor registrado"
  exit 0
fi

if [ -f "$PIDFILE" ] && kill -0 "$(cat "$PIDFILE")" 2>/dev/null; then
  echo "ya activo en http://127.0.0.1:$PORT (pid $(cat "$PIDFILE"))"
  exit 0
fi

mkdir -p .run
python3 -m http.server "$PORT" --directory docs/web >>"$LOG" 2>&1 &
echo $! > "$PIDFILE"
sleep 1
if kill -0 "$(cat "$PIDFILE")" 2>/dev/null && curl -sf -o /dev/null "http://127.0.0.1:$PORT/index.html"; then
  echo "sirviendo docs/web en http://127.0.0.1:$PORT (pid $(cat "$PIDFILE"), log $LOG)"
else
  echo "falló el arranque; revisa $LOG"
  exit 1
fi
