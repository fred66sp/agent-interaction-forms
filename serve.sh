#!/usr/bin/env bash
# Lanza la superficie web del tablero tasks-ai en 127.0.0.1:8890 (settings.json http).
# Servicio local, solo para revisión interactiva; pid y log en .run/. Se detiene con
# `./serve.sh stop`.
set -u
cd "$(dirname "$0")"
PIDFILE=.run/serve.pid
LOG=.run/serve.log

if [ "${1:-}" = "stop" ]; then
  if [ -f "$PIDFILE" ]; then
    PID=$(cat "$PIDFILE")
    if ps -p "$PID" -o command= 2>/dev/null | grep -q 'tasks-ai'; then
      kill "$PID" && rm -f "$PIDFILE" && echo "detenido"
    else
      rm -f "$PIDFILE" && echo "pidfile obsoleto (proceso $PID ya no es tasks-ai); nada que detener"
    fi
    exit 0
  fi
  echo "no hay servidor registrado"
  exit 0
fi

if [ -f "$PIDFILE" ] && kill -0 "$(cat "$PIDFILE")" 2>/dev/null; then
  echo "ya activo (pid $(cat "$PIDFILE"))"
  exit 0
fi

mkdir -p .run
./.tasks-ai/bin/tasks-ai serve >>"$LOG" 2>&1 &
echo $! > "$PIDFILE"
sleep 1
if kill -0 "$(cat "$PIDFILE")" 2>/dev/null && curl -sf -o /dev/null "http://127.0.0.1:8890/"; then
  echo "tablero en http://127.0.0.1:8890 (pid $(cat "$PIDFILE"), log $LOG)"
else
  echo "falló el arranque; revisa $LOG"
  exit 1
fi
