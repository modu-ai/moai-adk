#!/bin/zsh
# SPEC-CODEX-GHOST-SKILLS-MEASURE-001 (card t533) — layer-1 measurement, READ-ONLY.
#
# [HARD] This script writes nothing under ~/.codex. It contains no write verb and
# no write-enabling flag; AC-CGM-001 asserts that mechanically by grepping this
# file for the flag literal, so the literal must not appear even inside a comment.
# [HARD] Layer 2 (`moai clean --codex-skills`) is NOT invoked here, not even in
# its dry-run default form.
#
# Instrument discipline encoded below (each line is a recorded incident, not style):
#   - /bin/ls, never bare `ls`: the interactive alias on this machine is long-format
#     and silently pollutes any count built on it.
#   - `find -maxdepth 1 -mindepth 1` for directory enumeration, never the glob pair
#     `DIR/* DIR/.[!.]*`: the glob pair cannot match a name beginning with `..`
#     (measured hole: 72 real entries vs 70 seen). Under zsh it is worse — a nomatch
#     on either glob aborts the whole command line, yielding an EMPTY selector that
#     an absence-as-PASS check reads as "no change".
#   - `LC_ALL=C grep -a` on ~/.zsh_history: the default locale makes grep give up
#     silently on that Non-ISO extended-ASCII file (no output, rc=1) — indistinguishable
#     from an empty corpus.
#   - exit status is read immediately after an unpiped command; never after a pipeline.
#
# Usage: zsh .moai/reports/t533/measure.sh

set -u
CODEX="$HOME/.codex"
CFG="$CODEX/config.toml"

hdr() { print -r -- ""; print -r -- "=== $* ==="; }

hdr "coordinates"
print -r -- "worktree: $(pwd)"
print -r -- "HEAD:     $(git rev-parse HEAD)"
print -r -- "tree:     $(git rev-parse 'HEAD^{tree}')"

hdr "AC-CGM-001 — ~/.codex snapshot selectors (N6-corrected: find, not the glob pair)"
print -r -- "entries (/bin/ls -A):            $(/bin/ls -A "$CODEX" | wc -l | tr -d ' ')"
print -r -- "mtimes  (find -maxdepth 1):      $(find "$CODEX" -maxdepth 1 -mindepth 1 -exec stat -f '%N %m' {} + | wc -l | tr -d ' ')"
print -r -- "mtimes  (legacy glob pair):      $(stat -f '%N %m' "$CODEX"/* "$CODEX"/.[!.]* 2>/dev/null | wc -l | tr -d ' ')  <- fewer: the measured hole"
print -r -- "config.toml mtime:               $(stat -f '%m' "$CFG")"
print -r -- "t533 residue in ~/.codex:        $(/bin/ls -A "$CODEX" | grep -c 't533')"

hdr "AC-CGM-002 — ghost count, two independent counters"
print -r -- "grep: $(grep -c '^\[\[skills\.config\]\]' "$CFG")"
print -r -- "awk:  $(awk '/^\[\[skills\.config\]\]/{n++} END{print n}' "$CFG")"

hdr "AC-CGM-003 — the counting trap (whole-file count is NOT a block count)"
print -r -- "whole-file 'enabled = false': $(grep -c 'enabled = false' "$CFG")"
print -r -- "in-block   'enabled = false': $(awk '/^\[\[skills\.config\]\]/{inb=1;next} /^\[/{inb=0} inb&&/enabled = false/{n++} END{print n+0}' "$CFG")"

hdr "AC-CGM-004 — ghost paths absent, root alive"
grep -o '/Users/goos/.codex/skills/[^"]*' "$CFG" | head -5 | while IFS= read -r p; do
  if test -e "$p"; then print -r -- "EXISTS  $p"; else print -r -- "MISSING $p"; fi
done
print -r -- "root listing: $(/bin/ls -A "$CODEX/skills" | tr '\n' ' ')"

hdr "AC-CGM-005 — prune never wrote; discriminator is the PRODUCER's filename format"
out=$(/bin/ls -A "$CODEX" | grep -E 'config\.toml\.bak-[0-9]{8}T[0-9]{6}Z$')
print -r -- "strict (T...Z format): rc=$? len=${#out} out=[$out]"
print -r -- "loose control, same tool:  $(/bin/ls -A "$CODEX" | grep -c 'config\.toml\.bak-')"
print -r -- "loose control, DIFFERENT tool (find): $(find "$CODEX" -maxdepth 1 -name 'config.toml.bak-*' | wc -l | tr -d ' ')"

hdr "AC-CGM-006 — timeline is not bracketed: every snapshot holds the same count"
for f in "$CODEX/config.toml.bak-20260822-022202" "$CODEX/config.toml.bak-20260901-133347" "$CODEX/config.toml.bak" "$CFG"; do
  if test -e "$f"; then printf '%-40s %s\n' "${f:t}" "$(grep -c '^\[\[skills\.config\]\]' "$f")"
  else printf '%-40s ABSENT\n' "${f:t}"; fi
done

hdr "AC-CGM-007 — the reserializer is not moai"
print -r -- "non-test Go files naming skills.config:"
grep -rln 'skills\.config' internal/ --include='*.go' | grep -v _test.go | sort
out=$(grep -n 'os.WriteFile\|os.Create\|atomicWrite\|\.Encode(\|WriteString' internal/codexwiring/skills.go)
print -r -- "write-verb probe in codexwiring/skills.go: rc=$? len=${#out}"
print -r -- "control, same file (grep -c 'func '):      $(grep -c 'func ' internal/codexwiring/skills.go)"
print -r -- "control, DIFFERENT tool (awk):             $(awk '/func /{n++} END{print n+0}' internal/codexwiring/skills.go)"
print -r -- "path= list, 2026-08-22 vs now — ordered diff:"
grep '^path = ' "$CODEX/config.toml.bak-20260822-022202" > /tmp/t533_p_old.txt
grep '^path = ' "$CFG" > /tmp/t533_p_new.txt
diff /tmp/t533_p_old.txt /tmp/t533_p_new.txt
print -r -- "                            sorted (set) diff:"
sort /tmp/t533_p_old.txt > /tmp/t533_ps_old.txt; sort /tmp/t533_p_new.txt > /tmp/t533_ps_new.txt
diff /tmp/t533_ps_old.txt /tmp/t533_ps_new.txt
print -r -- "(ordered diff non-empty + sorted diff empty  ==  reordering only)"

hdr "AC-CGM-008 — judgeCodexSkillEntry never consults 'enabled'"
print -r -- "function starts: $(grep -n 'func judgeCodexSkillEntry' internal/cli/codex_skills_prune.go)"
print -r -- "function ends:   $(awk 'NR>=59 && /^}/{print NR; exit}' internal/cli/codex_skills_prune.go)"
out=$(sed -n '59,110p' internal/cli/codex_skills_prune.go | grep -c 'Enabled\|enabled')
print -r -- "probe  'enabled' in body: rc=$? len=${#out} out=[$out]"
out=$(sed -n '59,110p' internal/cli/codex_skills_prune.go | grep -c 'Path')
print -r -- "control 'Path'  in body:  rc=$? len=${#out} out=[$out]"
print -r -- "needle-viability control (same instrument, parser file): $(grep -c 'Enabled\|enabled' internal/codexwiring/skills.go)"

hdr "AC-CGM-009 — the invocation-history gap: instrument SCOPE, not instrument absence"
out=$(grep -rc 'codex-skills' "$HOME/.moai/logs/" 2>/dev/null | grep -v ':0$')
print -r -- "~/.moai/logs probe: len=${#out}  (len=0 means NO OUTPUT, which is not a zero)"
print -r -- "~/.moai/logs control (files naming moai): $(grep -rl 'moai' "$HOME/.moai/logs/" 2>/dev/null | wc -l | tr -d ' ')"
print -r -- "that corpus is: $(/bin/ls -A "$HOME/.moai/logs/" | tr '\n' ' ')"
# Instrument-identity note (run-phase correction 6): the silent-give-up on this file
# is NOT a locale/encoding property of grep. It is the Claude Code agent shell's `grep`
# SHELL FUNCTION, which routes to ugrep with -I (skip binary files) and --ignore-files.
# That function is not exported to a child shell, so the line below calls the real
# /usr/bin/grep and DOES print a count. Both instruments are recorded on purpose.
out=$(grep -c 'moai' "$HOME/.zsh_history" 2>&1)
print -r -- "whatever 'grep' resolves to here, -c 'moai': rc=$? len=${#out} out=[$out]"
out=$(/usr/bin/grep -c 'moai' "$HOME/.zsh_history" 2>&1)
print -r -- "REAL BINARY /usr/bin/grep -c 'moai':         rc=$? len=${#out} out=[$out]"
out=$(/usr/bin/grep -c 'codex-skills' "$HOME/.zsh_history" 2>&1)
print -r -- "REAL BINARY /usr/bin/grep -c 'codex-skills': rc=$? len=${#out} out=[$out]"
out=$(/usr/bin/grep -c 'moai clean' "$HOME/.zsh_history" 2>&1)
print -r -- "REAL BINARY /usr/bin/grep -c 'moai clean':   rc=$? len=${#out} out=[$out]"
print -r -- "different tool proves the corpus is not empty: $(head -1 "$HOME/.zsh_history")"
out=$(LC_ALL=C grep -ac 'moai' "$HOME/.zsh_history")
print -r -- "WORKING instrument LC_ALL=C grep -ac 'moai':         rc=$? len=${#out} out=[$out]"
out=$(LC_ALL=C grep -ac 'codex-skills' "$HOME/.zsh_history")
print -r -- "WORKING instrument LC_ALL=C grep -ac 'codex-skills': rc=$? len=${#out} out=[$out]"
out=$(LC_ALL=C grep -ac 'moai clean' "$HOME/.zsh_history")
print -r -- "WORKING instrument LC_ALL=C grep -ac 'moai clean':   rc=$? len=${#out} out=[$out]"
print -r -- "scope limit 3 (measured truncation ceiling): $(grep -n 'SAVEHIST' /etc/zshrc)"
print -r -- "                        entries in file now: $(LC_ALL=C grep -ac '^: [0-9]' "$HOME/.zsh_history")"
print -r -- "scope limit 4 (other terminals ARE visible): $(LC_ALL=C grep -ao 'lane-[0-9]*' "$HOME/.zsh_history" | sort -u | wc -l | tr -d ' ') distinct lane names"

hdr "done — this script wrote nothing under ~/.codex"
print -r -- "config.toml mtime (unchanged expected): $(stat -f '%m' "$CFG")"
