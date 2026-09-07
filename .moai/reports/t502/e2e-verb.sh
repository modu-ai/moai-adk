#!/bin/bash
# t502 AC-CSD-003 — end-to-end: the REAL verb writes the config, and the
# instrument reads that config verbatim (neither --entry-path nor --enabled),
# measuring marker 1 -> 0.
#
# usage: bash .moai/reports/t502/e2e-verb.sh <mirror-mode: symlink|copy>
set -u
MODE="${1:-copy}"
HERE="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(mktemp -d /tmp/t502-e2e-XXXXXXXX)"
BIN="$ROOT/moai"
SKILL=t502probe

echo "== codex version =="
codex --version

echo "== build the verb under test =="
go build -o "$BIN" ./cmd/moai || exit 1
"$BIN" skills disable --help > "$ROOT/help.txt" 2>&1
echo "help captured: $(wc -l < "$ROOT/help.txt") lines"

echo "== fixture ($MODE) =="
bash "$HERE/probe.sh" fixture --root "$ROOT/lab" --mirror "$MODE" --skill "$SKILL" > "$ROOT/fx.txt" || exit 1
cat "$ROOT/fx.txt"
PROJECT="$(grep '^PROJECT=' "$ROOT/fx.txt" | cut -d= -f2-)"
LITERAL_PATH="$(grep '^LITERAL_PATH=' "$ROOT/fx.txt" | cut -d= -f2-)"

CH="$ROOT/codex-home"
mkdir -p "$CH"
printf '# t502 e2e: a user config carrying no skills.config\n' > "$CH/config.toml"
chmod 0644 "$CH/config.toml"
echo "pre-run config sha256: $(shasum -a 256 "$CH/config.toml" | cut -d' ' -f1)"
echo "pre-run config mode:   $(stat -f '%Lp' "$CH/config.toml")"

echo "== probe BEFORE the verb (expect exposed) =="
bash "$HERE/probe.sh" probe --codex-home "$CH" --project "$PROJECT" --skill "$SKILL" --expect exposed
BEFORE=$?
echo "before rc=$BEFORE"

echo "== the verb =="
( cd "$PROJECT" && CODEX_HOME="$CH" "$BIN" skills disable "$SKILL" --codex ) 2>&1 | sed 's/^/[dry-run] /'
echo "dry-run left sha256: $(shasum -a 256 "$CH/config.toml" | cut -d' ' -f1)"
( cd "$PROJECT" && CODEX_HOME="$CH" "$BIN" skills disable "$SKILL" --codex --force ) 2>&1 | sed 's/^/[force] /'
echo "verb exit: $?"

echo "== what the verb wrote =="
cat "$CH/config.toml"
echo "post-run config mode: $(stat -f '%Lp' "$CH/config.toml")"
echo "expected entry path:  $LITERAL_PATH"
ls "$CH" | sed 's/^/  /'

echo "== probe AFTER the verb (expect gated) =="
bash "$HERE/probe.sh" probe --codex-home "$CH" --project "$PROJECT" --skill "$SKILL" --expect gated
AFTER=$?
echo "after rc=$AFTER"

echo "== negative control on the counter (AFTER capture) =="
# The counter's ability to return NON-zero is shown by the BEFORE probe line
# above (marker=1) — same instrument, same counter, same skill. Here we only
# show that an absent token counts 0, so the AFTER 0 is a gate and not a
# broken counter.
CAP="$CH/.probe-capture/out"
echo "real marker count (after, expect 0):   $(grep -c T502PROBE_MARKER "$CAP")"
echo "absent-token count (expect 0 always):  $(grep -c ZZZNOTAMARKER "$CAP")"
echo "name count (after, expect 0):          $(grep -c "$SKILL" "$CAP")"

echo "== idempotence: re-run --force must not add a second entry =="
( cd "$PROJECT" && CODEX_HOME="$CH" "$BIN" skills disable "$SKILL" --codex --force ) 2>&1 | sed 's/^/[force-2] /'
echo "entries declaring the path: $(grep -c "^path = \"$LITERAL_PATH\"$" "$CH/config.toml")"

echo "ROOT=$ROOT"
if [ "$BEFORE" -eq 0 ] && [ "$AFTER" -eq 0 ]; then echo "E2E PASS"; else echo "E2E FAIL"; exit 1; fi
