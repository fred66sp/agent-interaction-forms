#!/usr/bin/env bash
# PoC: AIF v0.2 over a real ACM operation (§9) with a throwaway memory.
#
# Copies an untouched snapshot of the ACM example (acm.py, content-hash.py,
# init.sh, schema.sql) into a temporary directory, initialises an isolated
# db-ai mount there (ACM_HOME, XDG_CONFIG_HOME and ACM_ALIAS all point inside
# the temp dir, so the user's db-ai configuration and memories are not read or
# written), builds the acmpoc integrator and runs four scenarios:
#   S1 direct execution          -> executed, one row
#   S2 missing required data     -> input_needed, no row
#   S3 invalid data, correction  -> invalid then executed, one new row
#   S4 same content again        -> ACM deduplicates, no second row
# Each scenario prints the request, the response, the exit code, the stored row
# and a parseable "VEREDICT Sn: PASS|FAIL" line. Exit 0 only if all four pass.
#
# Deterministic and non-interactive: no network, no LLM, no real database.
set -uo pipefail

root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P) || exit 1
snapshot=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P) || exit 1

die() { printf 'acmpoc: %s\n' "$*" >&2; exit 1; }

tmp=$(mktemp -d) || die "no se pudo crear el directorio temporal"
trap 'rm -rf "$tmp"' EXIT

export PATH="/home/desarrollo/.local/bin:$PATH"
export XDG_CONFIG_HOME="$tmp/xdg-config"
export ACM_HOME="$tmp"
export ACM_ALIAS=acmpoc
export OPENCODE_SESSION_ID=acmpoc
export ACM_POC_TOOL="$tmp/tools/acm.py"

db="$tmp/.agent-memory/memory.db"
bin="$tmp/acmpoc"
req1='{"aif":"0.2","operation":"learning","input":{"title":"El parser exige UTC","body":"Las fechas sin zona se interpretan como UTC; el test lo fija.","topic":"parser-fechas","confidence":"confirmed"}}'
req2='{"aif":"0.2","operation":"learning","input":{"topic":"parser-fechas"}}'
req3='{"aif":"0.2","operation":"learning","input":{"title":"Confianza sin corregir","body":"La confianza se corrige tras el rechazo.","confidence":"certain"}}'
req4='{"aif":"0.2","operation":"learning","input":{"title":"Confianza sin corregir","body":"La confianza se corrige tras el rechazo.","confidence":"inference"}}'

mkdir -p "$tmp/tools" || die "no se pudo preparar $tmp"
cp "$snapshot/init.sh" "$snapshot/schema.sql" "$tmp/" || die "no se pudo copiar la instantánea de init.sh/schema.sql"
cp "$snapshot/tools/acm.py" "$snapshot/tools/content-hash.py" "$tmp/tools/" \
  || die "no se pudo copiar la instantánea de acm.py/content-hash.py"

printf 'compilando acmpoc\n'
( cd "$root" && go build -o "$bin" ./poc/acm/cmd/acmpoc ) || die "go build falló"
cd "$tmp" || die "no se pudo entrar en $tmp"

printf '\n== memoria efímera aislada: %s (alias %s)\n' "$tmp" "$ACM_ALIAS"
bash init.sh || die "init.sh falló"
python3 "$ACM_POC_TOOL" start --goal "PoC AIF sobre ACM" >/dev/null || die "acm start falló"
printf 'base: %s\n' "$db"

count_rows() { sqlite3 "$db" 'SELECT count(*) FROM observations;'; }
count_titled() { sqlite3 "$db" "SELECT count(*) FROM observations WHERE title='$1';"; }
row_of() { sqlite3 "$db" "SELECT 'fila: '||id||' | '||title||' | '||confidence||' | dup='||duplicate_count FROM observations WHERE title='$1';"; }
show() { printf '%s\n' "$1"; printf 'exit=%d\n' "$2"; }
verdict() {
  if [ "$2" -eq 1 ]; then
    printf 'VEREDICT %s: PASS\n' "$1"
  else
    printf 'VEREDICT %s: FAIL\n' "$1"
    failed=1
  fi
}

failed=0

printf '\n== S1: ejecución directa con datos completos\ncomando: %s\n' "$req1"
out=$(printf '%s' "$req1" | "$bin"); code=$?
show "$out" "$code"
ok=0
if [ "$code" -eq 0 ] &&
  printf '%s' "$out" | grep -Fq '{"aif":"0.2","operation":"learning","outcome":{"ok":true,"result":"created"},"result":"executed"}' &&
  [ "$(count_rows)" -eq 1 ] && [ "$(count_titled 'El parser exige UTC')" -eq 1 ]; then
  ok=1
fi
row_of 'El parser exige UTC'
verdict S1 "$ok"

printf '\n== S2: información ausente (sin title ni body)\ncomando: %s\n' "$req2"
out=$(printf '%s' "$req2" | "$bin"); code=$?
show "$out" "$code"
ok=0
if [ "$code" -eq 2 ] &&
  printf '%s' "$out" | grep -Fq '"result":"input_needed"' &&
  printf '%s' "$out" | grep -Fq '"missing":["title","body"]' &&
  ! printf '%s' "$out" | grep -Fq '"invalid"' &&
  [ "$(count_rows)" -eq 1 ]; then
  ok=1
fi
printf 'filas: %s (sin cambios)\n' "$(count_rows)"
verdict S2 "$ok"

printf '\n== S3: dato inválido y corrección\ncomando: %s\n' "$req3"
out=$(printf '%s' "$req3" | "$bin"); code=$?
show "$out" "$code"
ok=0
if [ "$code" -eq 2 ] &&
  printf '%s' "$out" | grep -Fq '"result":"invalid"' &&
  printf '%s' "$out" | grep -Fq '"reason":"valor no permitido"' &&
  printf '%s' "$out" | grep -Fq '"allowed":["confirmed","inference","hypothesis","proposal","unknown"]' &&
  printf '%s' "$out" | grep -Fq '"fields"' &&
  [ "$(count_rows)" -eq 1 ]; then
  ok=1
fi
printf 'filas: %s (sin cambios)\n' "$(count_rows)"
printf 'comando: %s\n' "$req4"
out=$(printf '%s' "$req4" | "$bin"); code=$?
show "$out" "$code"
if [ "$code" -ne 0 ] ||
  ! printf '%s' "$out" | grep -Fq '"outcome":{"ok":true,"result":"created"}' ||
  [ "$(count_rows)" -ne 2 ] || [ "$(count_titled 'Confianza sin corregir')" -ne 1 ]; then
  ok=0
fi
row_of 'Confianza sin corregir'
verdict S3 "$ok"

printf '\n== S4: mismo contenido otra vez (garantías de ACM)\ncomando: %s\n' "$req1"
out=$(printf '%s' "$req1" | "$bin"); code=$?
show "$out" "$code"
ok=0
if [ "$code" -eq 0 ] &&
  printf '%s' "$out" | grep -Fq '{"aif":"0.2","operation":"learning","outcome":{"ok":true,"result":"consolidated"},"result":"executed"}' &&
  [ "$(count_rows)" -eq 2 ] &&
  [ "$(sqlite3 "$db" "SELECT duplicate_count FROM observations WHERE title='El parser exige UTC';")" -eq 2 ]; then
  ok=1
fi
row_of 'El parser exige UTC'
printf 'filas: %s (sin segunda fila)\n' "$(count_rows)"
verdict S4 "$ok"

if [ "$failed" -eq 0 ]; then
  printf '\nRESULT: 4/4 PASS\n'
else
  printf '\nRESULT: FAIL\n'
fi
exit "$failed"
