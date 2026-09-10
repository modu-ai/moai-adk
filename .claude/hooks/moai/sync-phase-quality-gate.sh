#!/bin/bash
# Hook: sync-phase-quality-gate
# Purpose: Fast sync-phase quality gate (compile/vet checks + dependency manifest-change observation)
# Trigger: Stop event when the current session's HEAD is a sync-phase commit
#
# Scope: the hook runs ONLY fast structural checks (compile/vet) that finish well
# within the Stop timeout. Heavy lint (golangci-lint) and the full test suite are
# deliberately NOT run here — they cannot finish within a turn-end Stop timeout and
# belong in CI. Each language's fast check lives inside its own case branch; absent
# tools are skipped gracefully; projects with no recognized language marker pass
# the gate silently.
#
# Behavior: BLOCKING by DEFAULT for the vet/build deterministic checks
# (observability hygiene policy, D3=Promote). A failing vet/build emits
# {"decision":"block", ...} on stdout + exit 0 — which blocks the turn.
# MOAI_SYNC_GATE_BLOCKING is the opt-OUT: set it to 0/off/false/advisory to
# downgrade a failing check to a non-blocking {"systemMessage": ...} warning.
# (The legacy opt-in semantics MOAI_SYNC_GATE_BLOCKING=1 are accepted but now
# redundant — blocking is the default.) tests/coverage are NOT run by this gate
# (advisory regardless — heavy checks belong in CI). This split matters because,
# per Claude Code Stop-hook semantics, stdout JSON is honored only on exit 0 (on
# exit 2 stdout is discarded and only stderr is surfaced) — the "decision" field
# is the blocking channel, so an advisory run must never emit that field.
# The runtime-recovery §4 carve-out (recovery turns SHOULD defer) is preserved:
# this script does not parse stopReason, so the carve-out remains documentation-
# only at this layer (per runtime-recovery-doctrine.md §4).
#
# Outcome record: for the HEAD it gates, the hook keeps one line
# "<head-sha> <outcome>" in .moai/state/sync-quality-gate.last. <outcome> is
# exactly one of running, pass, fail. The record reads "running" before any check
# starts, then "fail" if a check failed (whether the mode blocked or only advised)
# or "pass" otherwise. On a later turn with the same HEAD:
#   - pass: no checks, empty stdout.
#   - fail: no checks. The failing run's exact stdout, its kind (block or
#     advisory), and the failed-check exit codes are kept in
#     .moai/state/sync-quality-gate.payload. A stored block is re-delivered
#     byte-identical while the mode resolved on that turn is blocking; a stored
#     block under an advisory resolution, or a stored advisory message, stays
#     silent (the advisory warning is written once, by the run that checked).
#   - stop_hook_active: when stdin carries "stop_hook_active": true, a stored
#     block is not re-delivered on that turn and no state changes, so the next
#     turn without the flag re-delivers it. The flag never suppresses the output
#     of a run that executes the checks.
#   - running: a run did not finish. While the record is at most
#     SYNC_GATE_STALE_WINDOW seconds old, no checks run and a non-blocking notice
#     is emitted. An older record gets ONE re-run for that HEAD, recorded in
#     .moai/state/sync-quality-gate.retry; once that re-run is used, later turns
#     emit a non-blocking notice instead of re-running.
#   - no record, a record for another HEAD, or an empty, unreadable, legacy
#     (bare SHA), or malformed record: the checks run.
# Every state write goes through a temporary file renamed into place, and a
# failing run writes its payload before its "fail" record.
#
# Forcing a re-gate: delete .moai/state/sync-quality-gate.last (or .moai/state as
# a whole). There is no flag or environment variable for retrying.
#
# Manual smoke test:
#   echo '{}' | bash .claude/hooks/moai/sync-phase-quality-gate.sh
# Expected: empty stdout (silent pass) on skip/allow; on a blocking vet/build
# failure (DEFAULT) a Stop JSON {"decision":"block","reason":...,"systemMessage":...};
# set MOAI_SYNC_GATE_BLOCKING=0 to downgrade to an advisory {"systemMessage":...}
# warning. The per-check detail is written to .moai/logs/sync-quality-gate.log, not
# stdout (Stop JSON-schema rejects unknown fields and non-{approve,block} decision
# values).
#
# Unit-test the detector directly (bypasses the sync-phase git gate):
#   source .claude/hooks/moai/sync-phase-quality-gate.sh && detect_language "$dir"

set -e

# --- detect_language: directly-invocable, side-effect-free language detector ---
# Echoes a single language token (go|node|python|rust) or empty string when no
# recognized marker is present. Marker priority follows the language matrix order.
# This function MUST remain source-able so it can be unit-tested without first
# passing the sync-phase-commit git gate below.
detect_language() {
    root="${1:-.}"
    # Marker priority follows the language matrix order (16 supported languages)
    if [ -f "$root/go.mod" ]; then
        echo "go"
    elif [ -f "$root/pyproject.toml" ] || [ -f "$root/requirements.txt" ]; then
        echo "python"
    elif [ -f "$root/package.json" ]; then
        echo "node"
    elif [ -f "$root/Cargo.toml" ]; then
        echo "rust"
    elif [ -f "$root/pom.xml" ] || [ -f "$root/build.gradle" ] || [ -f "$root/build.gradle.kts" ]; then
        echo "java"
    elif [ -f "$root/Gemfile" ]; then
        echo "ruby"
    elif [ -f "$root/composer.json" ]; then
        echo "php"
    elif [ -f "$root/mix.exs" ]; then
        echo "elixir"
    elif [ -f "$root/CMakeLists.txt" ] || [ -f "$root/Makefile" ]; then
        echo "cpp"
    elif [ -f "$root/build.sbt" ] || [ -f "$root/pom.xml" ]; then
        echo "scala"
    elif [ -f "$root/DESCRIPTION" ] || [ -f "$root/renv.lock" ]; then
        echo "r"
    elif [ -f "$root/pubspec.yaml" ]; then
        echo "flutter"
    elif [ -f "$root/Package.swift" ]; then
        echo "swift"
    elif [ -d "$root/.vs" ] || find "$root" -maxdepth 1 -name '*.csproj' -print -quit 2>/dev/null | grep -q .; then
        echo "csharp"
    else
        echo ""
    fi
}

# --- code_delta_pattern: per-language source-file extension regex ---
# Used to detect whether the sync-phase commit touched code files; a 0-code-file
# delta means a docs/markdown-only sync and the gate skips.
code_delta_pattern() {
    case "$1" in
        go)       echo '\.go$' ;;
        python)   echo '\.py$' ;;
        node)     echo '\.(js|ts|jsx|tsx|mjs|cjs)$' ;;
        rust)     echo '\.rs$' ;;
        java)     echo '\.java$' ;;
        kotlin)   echo '\.kt|\.kts$' ;;
        csharp)   echo '\.cs$' ;;
        ruby)     echo '\.rb$' ;;
        php)      echo '\.php$' ;;
        elixir)   echo '\.ex$|\.exs$' ;;
        cpp)      echo '\.(cpp|cc|cxx|h|hpp|hxx)$' ;;
        scala)    echo '\.scala$' ;;
        r)        echo '\.r$|\.R$' ;;
        flutter)  echo '\.dart$' ;;
        swift)    echo '\.swift$' ;;
        *)        echo '' ;;
    esac
}

# Opt-out flag
if [ "$1" = "--skip-hook" ]; then
    echo "{\"skipped\": true, \"reason\": \"--skip-hook flag\"}" >&2
    mkdir -p "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs"
    echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [sync-phase-quality-gate] skipped via --skip-hook" \
        >> "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs/hook-skip.log"
    exit 0
fi

# When sourced for unit testing, stop here so detect_language is available
# without running the gate. Detection: BASH_SOURCE[0] != $0 means the file was
# sourced, not executed directly.
case "${BASH_SOURCE[0]}" in
    "$0") ;;            # executed directly — continue running the gate
    *) return 0 2>/dev/null || true ;;  # sourced — expose functions, do not run
esac

# Detect sync-phase via last commit subject
LAST_COMMIT_SUBJECT=$(git log -1 --format='%s' 2>/dev/null || echo "")
case "$LAST_COMMIT_SUBJECT" in
    *"docs("*"): sync-phase"*|*"chore("*"): sync-phase"*|*"docs: sync"*|*"chore: sync"*)
        ;;
    *)
        # Not a sync-phase commit — gate not applicable, silent pass.
        # stdout intentionally empty: Stop decision must be "approve" | "block"
        # (not "skip"); unknown fields also fail Claude Code JSON-schema validation.
        exit 0
        ;;
esac

# Resolve project root and detect language from canonical markers.
# GATE_LANG (not LANG): LANG is the reserved POSIX locale variable — assigning
# the detected language to it would change the locale of every child tool.
PROJECT_ROOT="${CLAUDE_PROJECT_DIR:-$PWD}"
GATE_LANG=$(detect_language "$PROJECT_ROOT")

# Silent pass when no recognized language marker is present (docs-only projects, etc.)
if [ -z "$GATE_LANG" ]; then
    # stdout intentionally empty (Stop schema: decision must be approve|block, not "skip").
    exit 0
fi

# Detect code-file changes in HEAD commit; skip if 0 code-file delta (markdown-only sync).
# On an initial commit HEAD~1 does not exist, so diff against the empty tree instead.
DELTA_PATTERN=$(code_delta_pattern "$GATE_LANG")
if git rev-parse --verify -q HEAD~1 >/dev/null 2>&1; then
    DIFF_RANGE="HEAD~1..HEAD"
else
    DIFF_RANGE=$(git hash-object -t tree /dev/null)  # empty-tree SHA — initial commit
fi
# grep -c is wrapped so its no-match exit (1) under `set -e` does not abort; the
# result is normalized to a single integer (avoids a "0\n0" double-emit).
CODE_DELTA=$(git diff --name-only "$DIFF_RANGE" 2>/dev/null | grep -cE "$DELTA_PATTERN" || true)
CODE_DELTA=${CODE_DELTA:-0}
if [ "$CODE_DELTA" -eq 0 ]; then
    # stdout intentionally empty (Stop schema: decision must be approve|block, not "skip").
    exit 0
fi

# --- Outcome record and auxiliary state (described in the header) ---
# This runs only after the early exits above, so a non-sync HEAD, an unrecognized
# project, or a docs-only delta never reads or writes any state.
HEAD_SHA=$(git rev-parse HEAD 2>/dev/null || echo "")
STATE_DIR="${CLAUDE_PROJECT_DIR:-$PWD}/.moai/state"
RECORD_FILE="$STATE_DIR/sync-quality-gate.last"
PAYLOAD_FILE="$STATE_DIR/sync-quality-gate.payload"
RETRY_FILE="$STATE_DIR/sync-quality-gate.retry"
GATE_LOG_DIR="${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs"
# Stale window, in seconds, for a "running" record. It must equal the timeout the
# settings template registers for this hook's Stop entry: a run older than that
# timeout was killed by the runtime and will never write its outcome.
SYNC_GATE_STALE_WINDOW=60

# write_state_file <path>: copy stdin into <path> through a temporary file in the
# state dir renamed into place, so a reader never sees a half-written file.
# Failures are swallowed: the gate exits 0 on every path, and a missing write
# leaves a state that later re-gates or notifies, never a silent pass.
write_state_file() {
    wsf_tmp="$STATE_DIR/.${1##*/}.tmp.$$"
    if cat > "$wsf_tmp" 2>/dev/null && mv -f "$wsf_tmp" "$1" 2>/dev/null; then
        return 0
    fi
    rm -f "$wsf_tmp" 2>/dev/null || true
    return 0
}

# log_gate_event <fields>: one audit line in the gate log; failures are ignored.
log_gate_event() {
    mkdir -p "$GATE_LOG_DIR" 2>/dev/null || true
    echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [sync-phase-quality-gate] language=$GATE_LANG $1 head=$HEAD_SHA" \
        >> "$GATE_LOG_DIR/sync-quality-gate.log" 2>/dev/null || true
}

# emit_gate_notice <text>: a non-blocking notice. It carries only a systemMessage,
# so repeated notices never count toward the runtime Stop-hook block cap.
emit_gate_notice() {
    printf '{"systemMessage":"%s"}\n' "$1"
}

# stop_hook_active_set: succeeds when stdin carries "stop_hook_active": true in
# object-key position. jq-free: the key's opening quote must not follow a
# backslash (an escaped literal inside a JSON string value does not count), and
# any run of spaces or tabs may separate the key, the colon, and the value. A
# nested key is not told apart from a top-level one. A terminal stdin is not read.
stop_hook_active_set() {
    if [ -t 0 ]; then
        return 1
    fi
    shas_stdin=$(cat 2>/dev/null || true)
    shas_tab=$(printf '\t')
    printf '%s\n' "$shas_stdin" | grep -Eq "(^|[^\\\\])\"stop_hook_active\"[ ${shas_tab}]*:[ ${shas_tab}]*true"
}

# record_age_seconds: prints the record file's age in seconds, or nothing when the
# mtime cannot be read; the caller then treats the record as stale, which runs
# the checks rather than staying silent.
record_age_seconds() {
    ras_mtime=$(stat -c %Y "$RECORD_FILE" 2>/dev/null || stat -f %m "$RECORD_FILE" 2>/dev/null || true)
    case "$ras_mtime" in
        ''|*[!0-9]*) return 0 ;;
    esac
    ras_now=$(date +%s 2>/dev/null || true)
    case "$ras_now" in
        ''|*[!0-9]*) return 0 ;;
    esac
    echo $((ras_now - ras_mtime))
}

# resolve_gate_mode: sets MODE (blocking|advisory) from MOAI_SYNC_GATE_BLOCKING,
# MOAI_AUTONOMY_TIER, DECISION, C1_EXIT, and C2_EXIT. A check run and a
# re-delivery both call it, so a stored failure is re-delivered only under the
# same rules that decide a fresh run.
resolve_gate_mode() {
    # D3=Promote (observability hygiene policy): vet/build block by DEFAULT.
    # MOAI_SYNC_GATE_BLOCKING is the opt-OUT — set to 0/off/false/advisory/no to
    # downgrade a failing vet/build to a non-blocking warning. Default (unset) and
    # the legacy =1 value both select blocking. tests/coverage are NOT run here.
    case "${MOAI_SYNC_GATE_BLOCKING:-1}" in
        0|off|false|advisory|no) MODE="advisory" ;;
        *) MODE="blocking" ;;
    esac

    # Stop-chain trim guard: tier-aware mode override. Read
    # $MOAI_AUTONOMY_TIER at the shell layer (no moai binary — the token is an
    # env-key per OQ-1/REQ-003 so shell can read it directly). The tier relaxes
    # ONLY the advisory-vs-blocking MODE of this gate, never the deny/ask denylist
    # (that lives in pre_tool.go and is tier-invariant per REQ-007).
    #   - fully-autonomous: advisory only (systemMessage, no decision:block).
    #   - automatic:        build-only-block — a C2 (build) failure still blocks,
    #                        but C1 (vet/lint) failures become advisory.
    #   - semi-auto/unset:  current MODE (no change — backward compat, AC-007).
    AUTONOMY_TIER=$(printf '%s' "${MOAI_AUTONOMY_TIER:-}" | tr '[:upper:]' '[:lower:]')
    case "$AUTONOMY_TIER" in
        fully-autonomous)
            MODE="advisory"
            ;;
        automatic)
            # Build (C2) failure still blocks; vet/lint (C1) failure → advisory.
            if [ "$DECISION" = "block" ] && [ "$C1_EXIT" -ne 0 ] && [ "$C2_EXIT" -eq 0 ]; then
                MODE="advisory"
            fi
            ;;
        *)
            # semi-auto / unset / unrecognized → MODE unchanged (AC-007 backward compat).
            ;;
    esac
}

RERUN_OF_RUNNING=0
if [ -n "$HEAD_SHA" ]; then
    RECORD_CONTENT=""
    if [ -f "$RECORD_FILE" ]; then
        RECORD_CONTENT=$(cat "$RECORD_FILE" 2>/dev/null || echo "")
    fi
    case "$RECORD_CONTENT" in
        "$HEAD_SHA pass")
            # This HEAD already passed the gate: silent, no re-run.
            exit 0
            ;;
        "$HEAD_SHA fail")
            PAYLOAD_HEADER=""
            if [ -f "$PAYLOAD_FILE" ]; then
                PAYLOAD_HEADER=$(head -n 1 "$PAYLOAD_FILE" 2>/dev/null || echo "")
            fi
            P_SHA=""; P_KIND=""; P_C1=""; P_C2=""; P_EXTRA=""
            read -r P_SHA P_KIND P_C1 P_C2 P_EXTRA <<< "$PAYLOAD_HEADER" || true
            PAYLOAD_VALID=0
            if [ "$P_SHA" = "$HEAD_SHA" ] && [ -z "$P_EXTRA" ]; then
                case "$P_KIND" in
                    block|advisory) PAYLOAD_VALID=1 ;;
                esac
                case "$P_C1" in ''|*[!0-9]*) PAYLOAD_VALID=0 ;; esac
                case "$P_C2" in ''|*[!0-9]*) PAYLOAD_VALID=0 ;; esac
            fi
            if [ "$PAYLOAD_VALID" = "1" ]; then
                if [ "$P_KIND" = "advisory" ]; then
                    # The advisory warning was written once, by the run that checked.
                    exit 0
                fi
                DECISION="block"
                C1_EXIT="$P_C1"
                C2_EXIT="$P_C2"
                resolve_gate_mode
                if [ "$MODE" != "blocking" ]; then
                    # An advisory resolution never re-delivers a stored block.
                    exit 0
                fi
                if stop_hook_active_set; then
                    log_gate_event "mode=$MODE decision=redelivery-deferred stop_hook_active=true"
                    exit 0
                fi
                tail -n +2 "$PAYLOAD_FILE" 2>/dev/null || true
                log_gate_event "mode=$MODE decision=block-redelivered"
                exit 0
            fi
            # A "fail" record without a usable payload is an unknown outcome: re-gate.
            ;;
        "$HEAD_SHA running")
            RECORD_AGE=$(record_age_seconds)
            if [ -n "$RECORD_AGE" ] && [ "$RECORD_AGE" -le "$SYNC_GATE_STALE_WINDOW" ]; then
                emit_gate_notice "sync-phase quality gate: the previous gate run for this HEAD has not completed yet, so no checks ran this turn. To force a new gate run, delete .moai/state/sync-quality-gate.last."
                log_gate_event "decision=running-notice age=$RECORD_AGE"
                exit 0
            fi
            if [ "$(cat "$RETRY_FILE" 2>/dev/null || echo "")" = "$HEAD_SHA" ]; then
                emit_gate_notice "sync-phase quality gate: the gate run for this HEAD has not completed and its one stale re-run was already used, so no checks ran this turn. Delete .moai/state/sync-quality-gate.last to force a new gate run."
                log_gate_event "decision=retry-exhausted-notice"
                exit 0
            fi
            RERUN_OF_RUNNING=1
            ;;
    esac

    # This invocation runs the checks: invalidate the payload, set or clear the
    # retry marker, and record "running" before any check starts.
    mkdir -p "$STATE_DIR" 2>/dev/null || true
    rm -f "$PAYLOAD_FILE" 2>/dev/null || true
    if [ "$RERUN_OF_RUNNING" = "1" ]; then
        printf '%s\n' "$HEAD_SHA" | write_state_file "$RETRY_FILE"
    else
        rm -f "$RETRY_FILE" 2>/dev/null || true
    fi
    printf '%s running\n' "$HEAD_SHA" | write_state_file "$RECORD_FILE"
fi

# Per-check result scratch dir.
# GATE_TMPDIR (not TMPDIR): TMPDIR is the reserved POSIX temp-dir variable —
# exporting/assigning it would redirect every child tool's temp files here.
GATE_TMPDIR=$(mktemp -d)
trap "rm -rf $GATE_TMPDIR" EXIT

# Default per-check results: 0 = pass/skipped, used when a step does not run for the
# detected language. command -v guards every tool invocation so an absent toolchain
# is skipped gracefully (exit 0, recorded as skipped) rather than failing.
echo "0" > "$GATE_TMPDIR/c1.exit"; echo "not run for $GATE_LANG" > "$GATE_TMPDIR/c1.log"
echo "0" > "$GATE_TMPDIR/c2.exit"; echo "not run for $GATE_LANG" > "$GATE_TMPDIR/c2.log"

# run_step <tool> <result-prefix> <command...>: run only if the tool is on PATH,
# otherwise record exit 0 and log a graceful skip. The `&& rc=0 || rc=$?` idiom
# captures the tool's exit code without letting `set -e` abort the hook when a
# tool legitimately fails (the failure is recorded and drives the decision).
run_step() {
    tool="$1"; prefix="$2"; shift 2
    if command -v "$tool" >/dev/null 2>&1; then
        local rc=0
        "$@" > "$GATE_TMPDIR/$prefix.log" 2>&1 && rc=0 || rc=$?
        echo "$rc" > "$GATE_TMPDIR/$prefix.exit"
    else
        echo "0" > "$GATE_TMPDIR/$prefix.exit"
        echo "skipped: $tool absent" > "$GATE_TMPDIR/$prefix.log"
    fi
}

# Fast structural checks only. Two slots per language: c1 (vet/lint) + c2 (build).
# Heavy lint (golangci-lint) and the full test suite are intentionally NOT run here —
# they cannot finish within the Stop timeout and belong in CI.
C1_LABEL="(none)"
C2_LABEL="(none)"

case "$GATE_LANG" in
    go)
        C1_LABEL="go vet"; C2_LABEL="go build"
        run_step go c1 go vet ./...
        run_step go c2 go build ./...
        ;;
    python)
        C1_LABEL="ruff"
        run_step ruff c1 ruff check .
        ;;
    node)
        C1_LABEL="eslint"
        run_step eslint c1 eslint .
        ;;
    rust)
        C1_LABEL="cargo check"
        run_step cargo c1 cargo check
        ;;
    java)
        C1_LABEL="javac compile check"
        # Simple compile check: find .java files and attempt compilation
        run_step javac c1 sh -c 'find . -name "*.java" -exec javac -cp "$(find . -name "*.jar" -printf "{}:")" {} + 2>&1 | head -20' || true
        ;;
    kotlin)
        C1_LABEL="kotlinc"
        run_step kotlinc c1 sh -c 'find . -name "*.kt" -exec kotlinc -cp "$(find . -name "*.jar" -printf "{}:")" {} + 2>&1 | head -20' || true
        ;;
    csharp)
        C1_LABEL="dotnet build"
        run_step dotnet c1 dotnet build --no-restore 2>&1 | head -30 || true
        ;;
    ruby)
        C1_LABEL="ruby syntax"
        run_step ruby c1 sh -c 'find . -name "*.rb" -exec ruby -c {} \; 2>&1' || true
        ;;
    php)
        C1_LABEL="php syntax"
        run_step php c1 sh -c 'find . -name "*.php" -exec php -l {} \; 2>&1' || true
        ;;
    elixir)
        C1_LABEL="mix compile"
        run_step mix c1 mix compile --no-start 2>&1 | head -20 || true
        ;;
    cpp)
        C1_LABEL="g++ syntax check"
        run_step g++ c1 sh -c 'find . -name "*.cpp" -o -name "*.cc" -exec g++ -fsyntax-only -std=c++17 {} \; 2>&1' || true
        ;;
    scala)
        C1_LABEL="scalac"
        run_step scalac c1 sh -c 'find . -name "*.scala" -exec scalac -cp "$(find . -name "*.jar" -printf "{}:")" {} + 2>&1 | head -20' || true
        ;;
    r)
        C1_LABEL="R syntax"
        run_step R c1 sh -c 'find . -name "*.R" -o -name "*.r" | head -5 | while read f; do Rscript -e "parse(\"$f\")" 2>&1; done' || true
        ;;
    flutter)
        C1_LABEL="dart analyze"
        run_step dart c1 dart analyze 2>&1 | head -30 || true
        ;;
    swift)
        C1_LABEL="swift build"
        run_step swift c1 swift build 2>&1 | head -30 || true
        ;;
esac

# Dependency manifest-change observation: set DEPS_MODIFIED=1 when a dependency
# manifest of the detected language changed in the HEAD commit (unexpected for a
# docs sync). Informational only — it does NOT drive the block decision and it is
# not a vulnerability scan. Language-specific manifest set.
DEPS_MANIFESTS=""
case "$GATE_LANG" in
    go)       DEPS_MANIFESTS="go.mod go.sum" ;;
    python)   DEPS_MANIFESTS="pyproject.toml requirements.txt poetry.lock" ;;
    node)     DEPS_MANIFESTS="package.json package-lock.json yarn.lock pnpm-lock.yaml" ;;
    rust)     DEPS_MANIFESTS="Cargo.toml Cargo.lock" ;;
    java)     DEPS_MANIFESTS="pom.xml build.gradle build.gradle.kts gradle.properties" ;;
    kotlin)   DEPS_MANIFESTS="pom.xml build.gradle.kts gradle.properties" ;;
    csharp)   DEPS_MANIFESTS="*.csproj packages.lock.json" ;;
    ruby)     DEPS_MANIFESTS="Gemfile Gemfile.lock" ;;
    php)      DEPS_MANIFESTS="composer.json composer.lock" ;;
    elixir)   DEPS_MANIFESTS="mix.exs mix.lock" ;;
    cpp)      DEPS_MANIFESTS="CMakeLists.txt Makefile" ;;
    scala)    DEPS_MANIFESTS="build.sbt pom.xml build.scala" ;;
    r)        DEPS_MANIFESTS="DESCRIPTION renv.lock .Rprofile" ;;
    flutter)  DEPS_MANIFESTS="pubspec.yaml pubspec.lock" ;;
    swift)    DEPS_MANIFESTS="Package.swift Package.resolved" ;;
esac
# Reuse the initial-commit-safe DIFF_RANGE computed above (HEAD~1..HEAD would
# fail on an initial commit; DIFF_RANGE already falls back to the empty tree).
git diff "$DIFF_RANGE" -- $DEPS_MANIFESTS > "$GATE_TMPDIR/deps.diff" 2>&1 || true
DEPS_MODIFIED=0
if [ -s "$GATE_TMPDIR/deps.diff" ]; then
    DEPS_MODIFIED=1
fi

C1_EXIT=$(cat "$GATE_TMPDIR/c1.exit")
C2_EXIT=$(cat "$GATE_TMPDIR/c2.exit")

# Decision
DECISION="allow"
BLOCKED_REASON=""
if [ "$C1_EXIT" -ne 0 ]; then
    DECISION="block"
    BLOCKED_REASON="$C1_LABEL failed"
elif [ "$C2_EXIT" -ne 0 ]; then
    DECISION="block"
    BLOCKED_REASON="$C2_LABEL failed"
fi

# Resolve the mode once (set -e safe) for both stdout and the audit log.
resolve_gate_mode

# Emit a Stop-schema-compliant response.
#
# Blocking (DEFAULT): a failing vet/build emits {"hookSpecificOutput":
# {"hookEventName":"Stop","decision":"block","reason":"..."},"systemMessage":"..."}
# on stdout — this blocks the turn. The decision/reason ride inside a
# hookSpecificOutput object carrying hookEventName:"Stop" per the official Stop
# hook contract (a bare top-level "decision" field is non-compliant for Stop).
# The hook still exits 0: per Claude Code hook semantics, stdout JSON is honored
# only on exit 0 (on exit 2 stdout is discarded and only stderr is surfaced).
#
# Advisory (opt-out, MOAI_SYNC_GATE_BLOCKING=0/off/false/advisory): a failing
# check emits ONLY {"systemMessage": ...} — a non-blocking warning. The
# nested "decision":"block" stdout field is the blocking channel (honored on
# exit 0), so the advisory path MUST NOT emit it.
#
# On allow, stdout is intentionally empty (silent pass); the audit log records detail.
# The response is composed into a scratch file first so the exact bytes can be
# stored for re-delivery before they are written to stdout.
GATE_OUTPUT_FILE="$GATE_TMPDIR/stdout"
: > "$GATE_OUTPUT_FILE"
PAYLOAD_KIND=""
if [ "$DECISION" = "block" ]; then
    if [ "$MODE" = "blocking" ]; then
        PAYLOAD_KIND="block"
        printf '{"hookSpecificOutput":{"hookEventName":"Stop","decision":"block","reason":"%s"},"systemMessage":"sync-phase quality gate BLOCKED: %s (%s=%s %s=%s deps_modified=%s). Detail: .moai/logs/sync-quality-gate.log"}\n' \
            "$BLOCKED_REASON" "$BLOCKED_REASON" "$C1_LABEL" "$C1_EXIT" "$C2_LABEL" "$C2_EXIT" "$DEPS_MODIFIED" > "$GATE_OUTPUT_FILE"
    else
        PAYLOAD_KIND="advisory"
        printf '{"systemMessage":"sync-phase quality gate WARNING (advisory, not blocking): %s (%s=%s %s=%s deps_modified=%s). Heavy lint/tests run in CI. Detail: .moai/logs/sync-quality-gate.log"}\n' \
            "$BLOCKED_REASON" "$C1_LABEL" "$C1_EXIT" "$C2_LABEL" "$C2_EXIT" "$DEPS_MODIFIED" > "$GATE_OUTPUT_FILE"
    fi
fi

# Record the outcome before writing stdout. A failing run writes its payload
# first and its "fail" record second, so a crash between the two leaves a state
# that re-gates or notifies on a later turn rather than one that passes silently.
if [ -n "$HEAD_SHA" ]; then
    if [ -n "$PAYLOAD_KIND" ]; then
        { printf '%s %s %s %s\n' "$HEAD_SHA" "$PAYLOAD_KIND" "$C1_EXIT" "$C2_EXIT"; cat "$GATE_OUTPUT_FILE"; } | write_state_file "$PAYLOAD_FILE"
        printf '%s fail\n' "$HEAD_SHA" | write_state_file "$RECORD_FILE"
    else
        printf '%s pass\n' "$HEAD_SHA" | write_state_file "$RECORD_FILE"
    fi
fi

cat "$GATE_OUTPUT_FILE"

mkdir -p "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs"
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [sync-phase-quality-gate] language=$GATE_LANG mode=$MODE decision=$DECISION $C1_LABEL=$C1_EXIT $C2_LABEL=$C2_EXIT deps_modified=$DEPS_MODIFIED head=$HEAD_SHA" \
    >> "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs/sync-quality-gate.log"

# The hook always exits 0. In blocking mode the {"decision":"block"} stdout JSON
# above is the blocking channel — exiting 2 here would make Claude Code discard
# that stdout JSON and surface only stderr (which the wrappers redirect away),
# silently losing the block reason.
exit 0
