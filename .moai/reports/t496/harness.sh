#!/bin/bash
# t496 firing campaign harness — SPEC-CODEX-EVENT-COVERAGE-001 M2.
# CODEX_HOME-isolated (t504/t83 shape): auth via symlink (no copy), minimal
# config, home-level hooks.json (t83 round3 H4: project-level does not fire).
# Real-home zero-write evidence uses a CONTROL window: snapshot C (t-20s),
# snapshot P (campaign start), snapshot O (campaign end). Any delta(P,O) that
# also appears in delta(C,P) is pre-existing live-home churn, not ours.
# Campaign artifacts live entirely under /tmp (mktemp EXP_ROOT).
set -u

EXP_ROOT="$(mktemp -d /tmp/t496-campaign-XXXXXXXX)"
mkdir -p "$EXP_ROOT/home" "$EXP_ROOT/proj" "$EXP_ROOT/captures" "$EXP_ROOT/runs" "$EXP_ROOT/homecheck"

codex --version > "$EXP_ROOT/version.txt" 2>&1

# ---- CODEX_HOME fixtures ----------------------------------------------------
ln -s "$HOME/.codex/auth.json" "$EXP_ROOT/home/auth.json"
printf '# t496 campaign home (isolated)\nmodel = "gpt-6-astra"\n' > "$EXP_ROOT/home/config.toml"

snap_home() {
  local out="$1"
  ( cd "$HOME/.codex" && find . -type f -print | sort ) > "$out.list"
  ( cd "$HOME/.codex" && find . -type f -print0 | xargs -0 stat -f '%m %N' | sort -k2 ) > "$out.mtimes"
}

snap_home "$EXP_ROOT/homecheck/C"
sleep 20
snap_home "$EXP_ROOT/homecheck/P"

# ---- hooks.json — one logger per campaign event -----------------------------
{
  printf '{\n  "hooks": {\n'
  for e in PreCompact PostCompact PermissionRequest SubagentStart SubagentStop Interrupt SessionStart SessionEnd Stop; do
    printf '    "%s": [ { "hooks": [ { "type": "command", "command": "/bin/sh -c '"'"'cat >> %s/captures/%s.jsonl'"'"'", "timeout": 10 } ] } ],\n' "$e" "$EXP_ROOT" "$e"
  done | sed '$s/,$//'
  printf '  }\n}\n'
} > "$EXP_ROOT/home/hooks.json"

# ---- run helper -------------------------------------------------------------
run_exec() {
  name="$1"; shift
  ( cd "$EXP_ROOT/proj" && CODEX_HOME="$EXP_ROOT/home" timeout 300 codex exec --json \
      --skip-git-repo-check --dangerously-bypass-hook-trust "$@" ) \
    > "$EXP_ROOT/runs/$name.jsonl" 2> "$EXP_ROOT/runs/$name.err" < /dev/null
  echo $? > "$EXP_ROOT/runs/$name.rc"
}

# R-P0: CODEX_HOME support verification (SessionStart logger must land here).
run_exec p0 "Reply with exactly: T496P0OK"

echo "EXP_ROOT=$EXP_ROOT"
echo "--- p0 rc=$(cat "$EXP_ROOT/runs/p0.rc") ---"
echo "--- captures after p0 ---"
ls -la "$EXP_ROOT/captures/"
echo "--- p0 last output ---"
tail -3 "$EXP_ROOT/runs/p0.jsonl"
echo "--- p0 stderr head ---"
head -5 "$EXP_ROOT/runs/p0.err"
