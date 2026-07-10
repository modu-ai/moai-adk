#!/bin/bash
# Hook: sync-phase-quality-gate
# Purpose: Fast sync-phase quality gate (compile/vet + dependency manifest audit)
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
# (SPEC-OBSERVE-HYGIENE-001 REQ-OBH-004, D3=Promote). A failing vet/build emits
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
# Once-per-commit: a given sync commit is gated at most ONCE. The gated HEAD SHA is
# recorded in .moai/state/sync-quality-gate.last and the hook short-circuits on any
# later turn whose HEAD is unchanged, so the gate does not re-run every turn-end.
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

# Once-per-commit sentinel: gate a given sync commit at most ONCE. Without this the
# Stop hook re-fires on every subsequent turn-end while HEAD is still the sync commit
# (the last-commit-subject trigger stays matched until a newer non-sync commit lands),
# re-running the toolchain each turn. Record the gated HEAD SHA and short-circuit when
# it is unchanged. The SHA is recorded BEFORE the checks run, so a slow/killed run
# still counts as gated and cannot re-trigger a per-turn re-run.
HEAD_SHA=$(git rev-parse HEAD 2>/dev/null || echo "")
STATE_DIR="${CLAUDE_PROJECT_DIR:-$PWD}/.moai/state"
SENTINEL_FILE="$STATE_DIR/sync-quality-gate.last"
if [ -n "$HEAD_SHA" ] && [ -f "$SENTINEL_FILE" ] && [ "$(cat "$SENTINEL_FILE" 2>/dev/null)" = "$HEAD_SHA" ]; then
    # This commit was already gated in a prior turn — silent pass, no re-run.
    exit 0
fi
if [ -n "$HEAD_SHA" ]; then
    mkdir -p "$STATE_DIR"
    echo "$HEAD_SHA" > "$SENTINEL_FILE"
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

# Dependency manifest audit: flag if a dependency manifest was modified in the
# sync-phase commit (unexpected for a docs sync). Informational only — it does NOT
# drive the block decision. Language-specific manifest set.
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
# D3=Promote (SPEC-OBSERVE-HYGIENE-001 REQ-OBH-004): vet/build block by DEFAULT.
# MOAI_SYNC_GATE_BLOCKING is the opt-OUT — set to 0/off/false/advisory/no to
# downgrade a failing vet/build to a non-blocking warning. Default (unset) and
# the legacy =1 value both select blocking. tests/coverage are NOT run here.
case "${MOAI_SYNC_GATE_BLOCKING:-1}" in
    0|off|false|advisory|no) MODE="advisory" ;;
    *) MODE="blocking" ;;
esac

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
if [ "$DECISION" = "block" ]; then
    if [ "$MODE" = "blocking" ]; then
        printf '{"hookSpecificOutput":{"hookEventName":"Stop","decision":"block","reason":"%s"},"systemMessage":"sync-phase quality gate BLOCKED: %s (%s=%s %s=%s deps_modified=%s). Detail: .moai/logs/sync-quality-gate.log"}\n' \
            "$BLOCKED_REASON" "$BLOCKED_REASON" "$C1_LABEL" "$C1_EXIT" "$C2_LABEL" "$C2_EXIT" "$DEPS_MODIFIED"
    else
        printf '{"systemMessage":"sync-phase quality gate WARNING (advisory, not blocking): %s (%s=%s %s=%s deps_modified=%s). Heavy lint/tests run in CI. Detail: .moai/logs/sync-quality-gate.log"}\n' \
            "$BLOCKED_REASON" "$C1_LABEL" "$C1_EXIT" "$C2_LABEL" "$C2_EXIT" "$DEPS_MODIFIED"
    fi
fi

mkdir -p "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs"
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) [sync-phase-quality-gate] language=$GATE_LANG mode=$MODE decision=$DECISION $C1_LABEL=$C1_EXIT $C2_LABEL=$C2_EXIT deps_modified=$DEPS_MODIFIED head=$HEAD_SHA" \
    >> "${CLAUDE_PROJECT_DIR:-$PWD}/.moai/logs/sync-quality-gate.log"

# The hook always exits 0. In blocking mode the {"decision":"block"} stdout JSON
# above is the blocking channel — exiting 2 here would make Claude Code discard
# that stdout JSON and surface only stderr (which the wrappers redirect away),
# silently losing the block reason.
exit 0
