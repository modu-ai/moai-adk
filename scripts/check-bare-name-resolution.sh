#!/bin/sh
# check-bare-name-resolution.sh — SPEC-INIT-SHRINK-001 REQ-008 + REQ-020.
#
# Measures the three bare-name resolution questions of REQ-008 under scratch
# config homes, driving the real tool runtimes through the t1434 routes:
#
#   claude-bare-skill    Skill("<bare>") resolves to the namespaced plugin
#                        skill (claude -p --plugin-dir, no scaffold copies)
#   claude-command-body  a plugin command body's bare Skill("moai") resolves
#                        (the same session shape, command-driven)
#   codex-naming         Codex lists the plugin-borne component names
#                        (codex plugin add + codex debug prompt-input render)
#
# Usage:
#   sh scripts/check-bare-name-resolution.sh <fixture-dir>
#       The full measurement. Real runtimes; the only sanctioned network
#       surface of this SPEC's verification (local fixture via --plugin-dir,
#       no marketplace network).
#   sh scripts/check-bare-name-resolution.sh --self-check [fixture-dir]
#       Harness-shape cases only: isolation scrub + negative control,
#       protected-set hash stability, output line shape. No real runtime is
#       started; safe under `go test`.
#   sh scripts/check-bare-name-resolution.sh --build-fixture <dir>
#       Writes the minimal probe plugin (one skill, the moai router skill,
#       one command whose body calls Skill("moai")) into <dir>.
#
# Output: one `PASS <case>` / `FAIL <case> <reason>` line per case, then
# `RESULT pass=<n> fail=<m>`. Exit 0 iff fail=0.
#
# Hermeticity (REQ-020): the scrub is live-enumerated (never a hand list);
# every case runs under a scratch CLAUDE_CONFIG_DIR / CODEX_HOME and a
# scratch working directory; the run is bracketed by an equal before/after
# protected-set hash that includes directory entries (the t1434
# LEAK-FINDING) AND the CONTENT of every protected file (card t1438 review
# finding 5: a name-only hash reads green through a same-name rewrite). No
# command in this script assigns HOME — the worktree guard refuses it and
# moving it into a script file is not a way around it.

set -u

PASS_COUNT=0
FAIL_COUNT=0

pass() { PASS_COUNT=$((PASS_COUNT + 1)); echo "PASS $1"; }
fail() { FAIL_COUNT=$((FAIL_COUNT + 1)); echo "FAIL $1 ${2:-unspecified}"; }

sha256_stream() {
    if command -v shasum >/dev/null 2>&1; then
        shasum -a 256
    else
        sha256sum
    fi
}

# env_var_names — the live environment's variable names, one per line.
env_var_names() {
    env | sed -n 's/^\([A-Za-z_][A-Za-z0-9_]*\)=.*/\1/p'
}

# foreign_env_names — the MOAI_/CLAUDE_/CODEX_ variable names currently set.
foreign_env_names() {
    env_var_names | grep -E '^(MOAI|CLAUDE|CODEX)_' || true
}

# scrub_and_isolate <scratch-root> — unsets every MOAI_/CLAUDE_/CODEX_
# variable, then points CLAUDE_CONFIG_DIR and CODEX_HOME at scratch homes.
# When MOAI_RESOLUTION_GATE_PLANT_LEAK is set the scrub is SKIPPED and the
# function returns 1 with the planted variables left in place: the negative
# control the self-check drives, proving the scrub is load-bearing.
scrub_and_isolate() {
    scratch="$1"
    if [ -n "${MOAI_RESOLUTION_GATE_PLANT_LEAK:-}" ]; then
        return 1
    fi
    for name in $(foreign_env_names); do
        # The two home variables are reset (not unset) below; everything else
        # in the three families goes.
        case "$name" in
            CLAUDE_CONFIG_DIR | CODEX_HOME) ;;
            *) unset "$name" 2>/dev/null || true ;;
        esac
    done
    CLAUDE_CONFIG_DIR="$scratch/claude-home"
    CODEX_HOME="$scratch/codex-home"
    export CLAUDE_CONFIG_DIR CODEX_HOME
    return 0
}

# protected_set_hash — a hash over the protected set of the real homes the
# run must not touch: the sorted directory and file ENTRIES plus the CONTENT
# of every regular file. Directory entries are included (t1434 LEAK-FINDING:
# content-only manifests cannot see a created directory); file contents are
# hashed because a name-only hash is vacuous against a same-name rewrite
# (card t1438 review finding 5 — the pre-fix form hashed only the pruned
# ambient paths, an empty-input hash). Read-only.
#
# The AMBIENT subtrees documented by the t1434 verdict (the account-synced
# plugins directory that changes by itself, the live-session trees a
# concurrently running Claude session writes, and the telemetry churn) are
# PRUNED: they are not writes this run could cause, and hashing them makes
# the gate flake on ambient motion rather than detect a leak. Measured on
# this class (card t1438 review finding 5 verification): the codex runtime's
# SQLite WAL/SHM sidecars rewrite continuously (logs_2.sqlite-shm flipped
# between two reads 4s apart), so the live sqlite stores, their -wal/-shm
# sidecars, and the live command-history jsonl files are pruned by name —
# a run's own leak shape (config, skills, plugin stores, a created
# directory) stays inside the hashed set.
#
# protected_set_hash_at <home-root> is the parameterized form the self-check
# negative control drives against a scratch fake home (the real home is
# never modified); protected_set_hash reads the real HOME.
protected_set_hash_at() {
    _home="$1"
    # Pass 1 — the structure: every non-pruned directory and file ENTRY.
    _structure=$(
        {
            if [ -d "$_home/.claude" ]; then
                find "$_home/.claude" -maxdepth 4 \
                    \( -name synced -o -name projects -o -name statsig -o -name todos \
                        -o -name shell-snapshots -o -name logs -o -name session-env \
                        -o -name '*.sqlite' -o -name '*-wal' -o -name '*-shm' \
                        -o -name history.jsonl -o -name timeline.jsonl \) -prune \
                    -o -print 2>/dev/null
            fi
            if [ -d "$_home/.codex" ]; then
                find "$_home/.codex" -maxdepth 4 \
                    \( -name log -o -name sessions -o -name archived_sessions \
                        -o -name sqlite \
                        -o -name '*.sqlite' -o -name '*-wal' -o -name '*-shm' \
                        -o -name history.jsonl -o -name session_index.jsonl \
                        -o -name transcription-history.jsonl \
                        -o -name models_cache.json \) -prune \
                    -o -print 2>/dev/null
            fi
        } | LC_ALL=C sort | sha256_stream | cut -d' ' -f1
    )
    # Pass 2 — the contents: every non-pruned regular file, hashed by bytes
    # (shasum/sha256sum accept the paths as operands). The /dev/null operand
    # keeps the xargs target from reading stdin on an empty file list (GNU
    # xargs runs its command once with no operands; the constant line is
    # deterministic). Symlinks are type l and never hashed or followed.
    _contents=$(
        {
            if [ -d "$_home/.claude" ]; then
                find "$_home/.claude" -maxdepth 4 \
                    \( -name synced -o -name projects -o -name statsig -o -name todos \
                        -o -name shell-snapshots -o -name logs -o -name session-env \
                        -o -name '*.sqlite' -o -name '*-wal' -o -name '*-shm' \
                        -o -name history.jsonl -o -name timeline.jsonl \) -prune \
                    -o -type f -print0 2>/dev/null
            fi
            if [ -d "$_home/.codex" ]; then
                find "$_home/.codex" -maxdepth 4 \
                    \( -name log -o -name sessions -o -name archived_sessions \
                        -o -name sqlite \
                        -o -name '*.sqlite' -o -name '*-wal' -o -name '*-shm' \
                        -o -name history.jsonl -o -name session_index.jsonl \
                        -o -name transcription-history.jsonl \
                        -o -name models_cache.json \) -prune \
                    -o -type f -print0 2>/dev/null
            fi
        } | xargs -0 sh -c '
                if command -v shasum >/dev/null 2>&1; then
                    shasum -a 256 /dev/null "$@"
                else
                    sha256sum /dev/null "$@"
                fi
            ' sh 2>/dev/null | LC_ALL=C sort | sha256_stream | cut -d' ' -f1
    )
    printf '%s\n%s\n' "$_structure" "$_contents" | sha256_stream | cut -d' ' -f1
}

protected_set_hash() {
    protected_set_hash_at "${HOME:?}"
}

# wait_for_file <path> <pid> <timeout-seconds> — polls until the path exists
# or the pid exits or the timeout fires. The bounded-wait kill is registered
# before the wait, so the child never outlives the loop.
wait_for_file() {
    _path="$1" _pid="$2" _budget="$3"
    _elapsed=0
    while [ "$_elapsed" -lt "$_budget" ]; do
        if [ -f "$_path" ]; then
            kill "$_pid" 2>/dev/null
            wait "$_pid" 2>/dev/null
            return 0
        fi
        if ! kill -0 "$_pid" 2>/dev/null; then
            wait "$_pid" 2>/dev/null
            return 0
        fi
        sleep 2
        _elapsed=$((_elapsed + 2))
    done
    kill "$_pid" 2>/dev/null
    wait "$_pid" 2>/dev/null
    return 1
}

# build_fixture <dir> — the minimal probe plugin, laid out for BOTH tools:
#
#   <dir>/market/.claude-plugin/marketplace.json   local marketplace wrapper
#   <dir>/market/plugin/...                        the plugin (source "./plugin")
#
# Claude consumes <dir>/market/plugin directly (--plugin-dir); Codex
# registers <dir>/market as a local marketplace and installs the plugin from
# it (codex plugin add requires PLUGIN@MARKETPLACE — a bare path is
# rejected). The plugin carries one skill (writes a marker when invoked),
# the moai router skill (the Skill("moai") target), and one command whose
# body calls Skill("moai").
build_fixture() {
    dir="$1"
    plugin="$dir/market/plugin"
    mkdir -p "$plugin/.claude-plugin" \
        "$dir/market/.claude-plugin" \
        "$plugin/skills/moai-resolution-probe" \
        "$plugin/skills/moai" \
        "$plugin/commands" || return 1
    cat >"$dir/market/.claude-plugin/marketplace.json" <<'EOF'
{
  "name": "resolution-market",
  "owner": {
    "name": "t1438-probe"
  },
  "plugins": [
    {
      "name": "resolution-probe",
      "source": "./plugin",
      "description": "SPEC-INIT-SHRINK-001 REQ-008 probe fixture (local only, never published)"
    }
  ]
}
EOF
    cat >"$plugin/.claude-plugin/plugin.json" <<'EOF'
{
  "name": "resolution-probe",
  "version": "0.0.1",
  "description": "SPEC-INIT-SHRINK-001 REQ-008 bare-name resolution probe fixture (local only, never published)"
}
EOF
    cat >"$plugin/skills/moai-resolution-probe/SKILL.md" <<'EOF'
---
name: moai-resolution-probe
description: Writes the marker file at MOAI_RESOLUTION_MARKER when invoked. REQ-008 probe fixture.
---

When this skill runs, immediately create the file whose path is in the
MOAI_RESOLUTION_MARKER environment variable, containing exactly the single
word `resolved`. Do nothing else.
EOF
    cat >"$plugin/skills/moai/SKILL.md" <<'EOF'
---
name: moai
description: Writes the marker file at MOAI_RESOLUTION_MARKER2 when invoked. REQ-008 probe fixture standing in for the moai router skill.
---

When this skill runs, immediately create the file whose path is in the
MOAI_RESOLUTION_MARKER2 environment variable, containing exactly the single
word `resolved-body`. Do nothing else.
EOF
    cat >"$plugin/commands/resolution-probe.md" <<'EOF'
---
description: REQ-008 probe command whose body resolves a bare skill name.
---

Continue by invoking the skill named `moai` with the Skill tool, passing no
arguments. Then stop.
EOF
    return 0
}

# self_check <fixture-dir> — the harness-shape cases. No real runtime runs.
self_check() {
    fixture="$1"
    scratch=$(mktemp -d "${TMPDIR:-/tmp}/t1438-selfcheck.XXXXXX")
    mkdir -p "$scratch/claude-home" "$scratch/codex-home"

    # Case isolation-scrub: plant foreign variables from all three families,
    # scrub, and assert none survives and both homes point under the scratch
    # root. (Explicit export + unset: a `VAR=x func` prefix has unspecified
    # persistence across shells.)
    MOAI_TEST_LEAK=1
    CLAUDE_TEST_LEAK=1
    CODEX_TEST_LEAK=1
    export MOAI_TEST_LEAK CLAUDE_TEST_LEAK CODEX_TEST_LEAK
    scrub_and_isolate "$scratch"
    scrub_rc=$?
    unset MOAI_TEST_LEAK CLAUDE_TEST_LEAK CODEX_TEST_LEAK || true
    leftover=$(foreign_env_names | grep -v -E '^(CLAUDE_CONFIG_DIR|CODEX_HOME)$' | grep -c . || true)
    homes_ok=0
    case "$CLAUDE_CONFIG_DIR" in "$scratch"/*) case "$CODEX_HOME" in "$scratch"/*) homes_ok=1 ;; esac ;; esac
    if [ "$scrub_rc" -eq 0 ] && [ "$leftover" -eq 0 ] && [ "$homes_ok" -eq 1 ]; then
        pass isolation-scrub
    else
        fail isolation-scrub "rc=$scrub_rc leftover=$leftover homes_ok=$homes_ok"
    fi

    # Case isolation-scrub-negative-control: with the scrub disabled the
    # gate MUST report the unscrubbed environment — the control fails the
    # case iff the scrub is skipped while a leak is planted, proving the
    # scrub is load-bearing (not vacuous).
    MOAI_RESOLUTION_GATE_PLANT_LEAK=1 scrub_and_isolate "$scratch"
    plant_rc=$?
    planted=$(foreign_env_names | grep -c . || true)
    if [ "$plant_rc" -ne 0 ] && [ "$planted" -gt 0 ]; then
        pass isolation-scrub-negative-control
    else
        fail isolation-scrub-negative-control "rc=$plant_rc planted=$planted"
    fi
    unset MOAI_RESOLUTION_GATE_PLANT_LEAK || true

    # The FAKE home: the deterministic tree the hash cases drive. The real
    # home carries ambient churn (live-session and telemetry writes) that no
    # prune list can fully exclude, so STABILITY is asserted on the fake
    # home; the real-home read is only asserted non-empty (the function
    # works against the live tree).
    fake_home="$scratch/fake-home"
    mkdir -p "$fake_home/.claude/skills/demo" "$fake_home/.codex"
    printf 'original bytes\n' >"$fake_home/.claude/skills/demo/SKILL.md"
    printf '{}\n' >"$fake_home/.codex/config.toml"

    # Case protected-set-hash: the hash is deterministic across back-to-back
    # reads of the same tree, and the real-home read is non-empty (a vacuous
    # hash would make the before/after bracket meaningless).
    before=$(protected_set_hash_at "$fake_home")
    after=$(protected_set_hash_at "$fake_home")
    real_hash=$(protected_set_hash)
    if [ -n "$before" ] && [ "$before" = "$after" ] && [ -n "$real_hash" ]; then
        pass protected-set-hash
    else
        fail protected-set-hash "empty=$([ -z "$before" ] && echo yes || echo no) stable=$([ "$before" = "$after" ] && echo yes || echo no) real_empty=$([ -z "$real_hash" ] && echo yes || echo no)"
    fi

    # Case protected-set-hash-negative-control: a REAL tamper on the fake
    # home must flip the hash — the content axis makes the check
    # non-vacuous (card t1438 review finding 5: the pre-fix form hashed only
    # the pruned ambient paths, so a rewritten protected file read as
    # unchanged). The real home is never touched.
    tamper_before=$(protected_set_hash_at "$fake_home")
    printf 'tampered bytes\n' >"$fake_home/.claude/skills/demo/SKILL.md"
    tamper_after=$(protected_set_hash_at "$fake_home")
    if [ -n "$tamper_before" ] && [ "$tamper_before" != "$tamper_after" ]; then
        pass protected-set-hash-negative-control
    else
        fail protected-set-hash-negative-control \
            "caught=$([ "$tamper_before" != "$tamper_after" ] && echo yes || echo no) empty=$([ -z "$tamper_before" ] && echo yes || echo no)"
    fi

    # Case shape-lines: the PASS/FAIL emitters produce the documented line
    # shapes (captured in a subshell so the counters here are untouched).
    shapes=$(
        PASS_COUNT=0
        FAIL_COUNT=0
        pass shape-probe
        fail shape-probe 2>/dev/null
    )
    echo "$shapes" | grep -q '^PASS shape-probe$' &&
        echo "$shapes" | grep -q '^FAIL shape-probe' &&
        pass shape-lines || fail shape-lines "unexpected emission shape"

    # The fixture directory is not exercised by the shape cases, but the
    # caller must pass a real one so a typo'd invocation fails loudly here
    # rather than silently in the full mode.
    if [ -d "$fixture" ]; then
        pass fixture-present
    else
        fail fixture-present "not a directory: $fixture"
    fi

    rm -rf "$scratch"
}

# run_measurement <fixture-dir> — the three REQ-008 questions against the
# real runtimes, bracketed by the protected-set hash.
run_measurement() {
    fixture="$1"
    scratch=$(mktemp -d "${TMPDIR:-/tmp}/t1438-resolution.XXXXXX")
    mkdir -p "$scratch/claude-home" "$scratch/codex-home" "$scratch/work"

    if ! scrub_and_isolate "$scratch"; then
        fail isolation-scrub "negative-control hook set in a full run"
        echo "RESULT pass=$PASS_COUNT fail=$FAIL_COUNT"
        return 1
    fi
    pass isolation-scrub

    before_hash=$(protected_set_hash)

    marker_skill="$scratch/work/marker-skill.txt"
    marker_body="$scratch/work/marker-body.txt"
    rm -f "$marker_skill" "$marker_body"
    MOAI_RESOLUTION_MARKER="$marker_skill"
    MOAI_RESOLUTION_MARKER2="$marker_body"
    export MOAI_RESOLUTION_MARKER MOAI_RESOLUTION_MARKER2

    # The plugin lives inside the fixture's market/ wrapper (see
    # build_fixture): Claude loads it by path, Codex installs it from the
    # local marketplace.
    plugin_dir="$fixture/market/plugin"
    market_dir="$fixture/market"
    if [ ! -f "$plugin_dir/.claude-plugin/plugin.json" ] || [ ! -f "$market_dir/.claude-plugin/marketplace.json" ]; then
        fail fixture-present "fixture missing market/plugin layout (build with --build-fixture)"
        echo "RESULT pass=$PASS_COUNT fail=$FAIL_COUNT"
        return 1
    fi
    pass fixture-present

    # Q1 — claude-bare-skill: a scratch-working-directory session (no
    # scaffold copies anywhere in the tree) asked to invoke the fixture
    # skill by its bare name.
    (
        cd "$scratch/work" || exit 1
        claude -p "Invoke the skill named moai-resolution-probe (bare name, no namespace prefix) with the Skill tool." \
            --plugin-dir "$plugin_dir" --no-session-persistence --max-turns 16 --allowedTools "Write,Bash,Edit" \
            >"$scratch/claude-skill.log" 2>&1
    ) &
    skill_pid=$!
    if wait_for_file "$marker_skill" "$skill_pid" 420 && [ "$(cat "$marker_skill" 2>/dev/null)" = "resolved" ]; then
        pass claude-bare-skill
    else
        fail claude-bare-skill "marker not written (log kept: $scratch/claude-skill.log)"
    fi

    # Q2 — claude-command-body: the fixture command's body calls
    # Skill("moai"); its marker proves the bare name resolved from inside a
    # plugin command body.
    (
        cd "$scratch/work" || exit 1
        claude -p "/resolution-probe" \
            --plugin-dir "$plugin_dir" --no-session-persistence --max-turns 16 --allowedTools "Write,Bash,Edit" \
            >"$scratch/claude-command.log" 2>&1
    ) &
    body_pid=$!
    if wait_for_file "$marker_body" "$body_pid" 420 && [ "$(cat "$marker_body" 2>/dev/null)" = "resolved-body" ]; then
        pass claude-command-body
    else
        fail claude-command-body "marker not written (log kept: $scratch/claude-command.log)"
    fi

    # Q3 — codex-naming: the fixture registered under the scratch CODEX_HOME
    # through its LOCAL MARKETPLACE (codex plugin add requires
    # PLUGIN@MARKETPLACE; a bare path is rejected), then the model-visible
    # prompt render scanned for the plugin-borne component names.
    (
        cd "$scratch/work" || exit 1
        codex plugin marketplace add "$market_dir" >"$scratch/codex-market-add.log" 2>&1 || exit 1
        codex plugin add "resolution-probe@resolution-market" >"$scratch/codex-add.log" 2>&1 || exit 1
        codex debug prompt-input >"$scratch/codex-render.json" 2>"$scratch/codex-render.err"
    )
    codex_rc=$?
    if [ "$codex_rc" -eq 0 ] && grep -q "moai-resolution-probe" "$scratch/codex-render.json" 2>/dev/null; then
        pass codex-naming
    else
        fail codex-naming "render did not list the plugin-borne names (rc=$codex_rc; see $scratch/codex-add.log)"
    fi

    after_hash=$(protected_set_hash)
    if [ "$before_hash" = "$after_hash" ]; then
        pass hermeticity-protected-set
    else
        fail hermeticity-protected-set "the protected set changed during the run"
    fi

    # Keep the scratch (its logs are the deciding evidence for any FAIL
    # case) when the verdict is not clean; clean it only on a full pass.
    if [ "$FAIL_COUNT" -eq 0 ]; then
        rm -rf "$scratch"
    else
        echo "note: scratch evidence kept at $scratch"
    fi
}

MODE="${1:-}"
FIXTURE_ARG="${2:-}"

case "$MODE" in
    --self-check)
        fixture="${FIXTURE_ARG:-$(mktemp -d "${TMPDIR:-/tmp}/t1438-selfcheck-fixture.XXXXXX")}"
        self_check "$fixture"
        ;;
    --build-fixture)
        if [ -z "$FIXTURE_ARG" ]; then
            echo "FAIL usage: --build-fixture requires a target directory"
            echo "RESULT pass=0 fail=1"
            exit 1
        fi
        if build_fixture "$FIXTURE_ARG"; then
            pass build-fixture
        else
            fail build-fixture "could not write the fixture"
            exit 1
        fi
        ;;
    *)
        # The plain measurement form: the fixture directory is the FIRST
        # argument (`sh check-bare-name-resolution.sh <fixture-dir>`). A
        # second argument is accepted and wins, for symmetry with the
        # two-argument modes above.
        fixture="$MODE"
        if [ -n "$FIXTURE_ARG" ]; then
            fixture="$FIXTURE_ARG"
        fi
        if [ -z "$fixture" ] || [ ! -d "$fixture" ]; then
            echo "FAIL usage: pass a fixture directory (build one with --build-fixture), or use --self-check"
            echo "RESULT pass=0 fail=1"
            exit 1
        fi
        run_measurement "$fixture"
        ;;
esac

echo "RESULT pass=$PASS_COUNT fail=$FAIL_COUNT"
[ "$FAIL_COUNT" -eq 0 ]
