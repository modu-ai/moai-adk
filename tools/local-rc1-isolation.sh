#!/bin/bash
# #2262 local RC1 kit — isolation rehearsal. Everything happens under a fresh
# temporary HOME and a fresh temporary project; the real HOME is never passed.
# Usage: bash tools/local-rc1-isolation.sh <moai-cli binary> <repo root (for skills/moai)> <out dir>
# Project leg order (#2291): project -> generation (creates .moai-cli/workflows/<id>/ itself,
# #2274) -> agents with the same workflow_reference. No WF_REF workaround is needed or
# accepted any more. Results: docs/local-rc1-kit.md §2.
set -u
BIN="$1"; REPO="$2"; OUT="$3"
mkdir -p "$OUT"
H="$(mktemp -d /tmp/moai-kit-home.XXXXXX)"
P="$(mktemp -d /tmp/moai-kit-proj.XXXXXX)"
echo "HOME=$H PROJECT=$P" > "$OUT/paths.txt"
export HOME="$H" GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1

# --- seed what a real machine already has (must survive untouched) ---
mkdir -p "$H/.claude/skills/moai" "$H/.agents/skills/moai" "$H/.claude/skills/other" \
         "$H/.moai" "$H/.codex" "$P/.moai" "$P/.claude"
printf '# pre-existing moai skill (Go generation)\n' > "$H/.claude/skills/moai/SKILL.md"
printf '# pre-existing agents moai skill\n' > "$H/.agents/skills/moai/SKILL.md"
printf '# unrelated user skill\n' > "$H/.claude/skills/other/SKILL.md"
printf 'go state\n' > "$H/.moai/state"
printf '# personal codex AGENTS\n' > "$H/.codex/AGENTS.md"
printf 'go project state\n' > "$P/.moai/config"
printf '{"user":true}\n' > "$P/.claude/settings.json"
printf '# My project rules\n\nkeep this line\n' > "$P/AGENTS.md"
printf '# personal override\n' > "$P/AGENTS.override.md"

snap() { (cd "$1" && find . -type f ! -path './.git/*' -print0 | sort -z | xargs -0 shasum -a 256); }
snap "$H" > "$OUT/home-before.sha256"; snap "$P" > "$OUT/proj-before.sha256"

run() { local name="$1"; shift; "$BIN" "$@" > "$OUT/$name.json" 2> "$OUT/$name.err"; echo "$name exit=$?"; }

# --- project leg: project-plan -> project-apply (SETUP, profile) ---
printf '{"schema_version":2,"project_id":"kit-demo","purpose":"Try moai-cli rc.1 locally","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"],"interview_revision":"kit-1"}' > "$OUT/project-request.json"
run project-plan project-plan "$P" < "$OUT/project-request.json"
python3 -I -c 'import json,sys; v=json.load(open(sys.argv[1])); json.dump(v["result"],open(sys.argv[2],"w"))' "$OUT/project-plan.json" "$OUT/project-plan.result.json" 2>/dev/null
PD="$(python3 -I -c 'import json,sys; print(json.load(open(sys.argv[1]))["plan_digest"])' "$OUT/project-plan.result.json" 2>/dev/null)"
run project-apply project-apply "$P" --plan-digest "$PD" --permission-reference kit-isolation < "$OUT/project-plan.result.json"

# --- project leg: generation-plan -> generation-apply (custom workflow + AGENTS.md block) ---
WORKFLOW_REF=".moai-cli/workflows/kit-demo/"
printf '{"protocol_version":1,"language":"en","workflow_reference":"%s","local_file":"none","workflow_id":"kit-demo","native_roles":[],"skill_projections":[]}' "$WORKFLOW_REF" > "$OUT/generation-request.json"
run generation-plan generation-plan "$P" < "$OUT/generation-request.json"
python3 -I -c 'import json,sys; v=json.load(open(sys.argv[1])); json.dump(v["result"],open(sys.argv[2],"w"))' "$OUT/generation-plan.json" "$OUT/generation-plan.result.json" 2>/dev/null
GD="$(python3 -I -c 'import json,sys; print(json.load(open(sys.argv[1]))["plan_digest"])' "$OUT/generation-plan.result.json" 2>/dev/null)"
run generation-apply generation-apply "$P" --plan-digest "$GD" --permission-reference kit-isolation < "$OUT/generation-plan.result.json"

# --- project leg: agents-plan -> agents-apply (re-sync the block, same workflow_reference) ---
printf '{"protocol_version":1,"language":"en","workflow_reference":"%s","local_file":"none"}' "$WORKFLOW_REF" > "$OUT/agents-request.json"
run agents-plan agents-plan "$P" < "$OUT/agents-request.json"
python3 -I -c 'import json,sys; v=json.load(open(sys.argv[1])); json.dump(v["result"],open(sys.argv[2],"w"))' "$OUT/agents-plan.json" "$OUT/agents-plan.result.json" 2>/dev/null
AD="$(python3 -I -c 'import json,sys; print(json.load(open(sys.argv[1]))["plan_digest"])' "$OUT/agents-plan.result.json" 2>/dev/null)"
run agents-apply agents-apply "$P" --plan-digest "$AD" --permission-reference kit-isolation < "$OUT/agents-plan.result.json"

# --- home leg: update-plan -> update-apply -> recovery-plan ---
run update-plan update-plan "$H" "$REPO/skills/moai" < /dev/null
python3 -I -c 'import json,sys; v=json.load(open(sys.argv[1])); json.dump(v["result"],open(sys.argv[2],"w"))' "$OUT/update-plan.json" "$OUT/update-plan.result.json" 2>/dev/null
UD="$(python3 -I -c 'import json,sys; print(json.load(open(sys.argv[1]))["plan_digest"])' "$OUT/update-plan.result.json" 2>/dev/null)"
run update-apply update-apply "$H" "$REPO/skills/moai" --plan-digest "$UD" --permission-reference kit-isolation < "$OUT/update-plan.result.json"
run recovery-plan recovery-plan "$H" < /dev/null

snap "$H" > "$OUT/home-after.sha256"; snap "$P" > "$OUT/proj-after.sha256"
echo "--- home files changed or removed (pre-existing only):"
comm -23 <(sort "$OUT/home-before.sha256") <(sort "$OUT/home-after.sha256")
echo "--- project files changed or removed (pre-existing only):"
comm -23 <(sort "$OUT/proj-before.sha256") <(sort "$OUT/proj-after.sha256")
echo "--- new home files:"; comm -13 <(sort "$OUT/home-before.sha256") <(sort "$OUT/home-after.sha256") | awk '{print $2}'
echo "--- new project files:"; comm -13 <(sort "$OUT/proj-before.sha256") <(sort "$OUT/proj-after.sha256") | awk '{print $2}'
