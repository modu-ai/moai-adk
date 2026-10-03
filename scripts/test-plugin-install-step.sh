#!/bin/sh
# Stub harness for the moai plugin install step, the verb and the install scripts (SPEC-PLUGIN-MARKETPLACE-001, REQ-025,
# AC-018 (a), AC-019 (c), AC-025, design.md section 6).
#
# Usage: test-plugin-install-step.sh [FLAG] <path-to-moai>
#   no flag                          the twelve default cases; exit 0 only when none fails
#   --negative-control               the scrub disabled and a working directory inside a project whose llm.yaml pins the
#                                    recorder: the isolation cases that depend on them must go red (AC-025 (b))
#   --negative-control-cwd           the scrub working, the working directory inside that project (AC-025 (c))
#   --typed-list-mutant              the scrub replaced by a typed list of two names (AC-025 (c))
#   --negative-control-install-dir   the four installer cases without --install-dir (AC-025 (f))
#   --negative-control-decoy-missing one decoy root removed: the installer must never run
# Each negative control exits 0 only when exactly the expected cases go red and the others stay green.
#
# Why the decoys matter: without --install-dir, install.sh installs into `go env GOBIN`, else `$GOPATH/bin`, else
# `$HOME/.local/bin`. A stub `go` answers the first two with decoy directories inside the scratch. `$HOME/.local/bin`
# (a real directory) is reached only when BOTH decoys are absent, so the harness asserts that both exist, and aborts
# before the installer runs if one does not. Nothing here sets HOME and nothing here runs pwsh.
set -eu

usage="usage: $0 [--negative-control | --negative-control-cwd | --typed-list-mutant | --negative-control-install-dir | --negative-control-decoy-missing] <path-to-moai>"
mode=normal
while [ $# -gt 0 ]; do
    case $1 in
        --negative-control) mode=neg-scrub; shift ;;
        --negative-control-cwd) mode=neg-cwd; shift ;;
        --typed-list-mutant) mode=typed-list; shift ;;
        --negative-control-install-dir) mode=omit-flag; shift ;;
        --negative-control-decoy-missing) mode=decoy-missing; shift ;;
        -*) echo "$usage" >&2; exit 2 ;;
        *) break ;;
    esac
done
[ $# -eq 1 ] || { echo "$usage" >&2; exit 2; }

repo=$(cd "$(dirname "$0")/.." && pwd -P)
hashsh=$repo/scripts/protected-set-hash.sh
moai_bin=$1
case $moai_bin in /*) ;; *) moai_bin=$(pwd -P)/$moai_bin ;; esac
[ -x "$moai_bin" ] || { echo "test-plugin-install-step: not an executable file: $moai_bin" >&2; exit 2; }

S=$(mktemp -d)
S=$(cd "$S" && pwd -P)
trap 'rm -rf "$S"' EXIT
trap 'exit 1' INT TERM

sha256() {
    if command -v shasum >/dev/null 2>&1; then shasum -a 256; else sha256sum; fi | awk '{print $1}'
}

# 1. Capture the protected real roots from the caller's own environment BEFORE any scrub, then take the before hash.
sh "$hashsh" --save-roots "$S/roots"
before=$(sh "$hashsh" --roots-file "$S/roots" --dump "$S/before.txt")

# 2. Plant poison in this script's own environment, so the scrub is judged on names it did not choose: a pin naming a
#    recorder, a canary directory that stands in for a real profile, and one variable of each family whose name carries the
#    process id (no typed list can name those). The recorder logs its run and, when it can see the canary, grows it.
poison_pid=$$
mkdir -p "$S/canary/plugins/data"
: > "$S/recorder.log"
printf '#!/bin/sh\necho "recorder $*" >> "%s/recorder.log"\n[ -z "${MOAI_CANARY_DIR:-}" ] || mkdir -p "$MOAI_CANARY_DIR/leaked/dir"\n' "$S" > "$S/recorder"
chmod +x "$S/recorder"
export MOAI_CLAUDE_BIN="$S/recorder" MOAI_CANARY_DIR="$S/canary"
export "MOAI_POISON_$poison_pid=1" "CLAUDE_POISON_$poison_pid=1" "CODEX_POISON_$poison_pid=1"

family_names() { awk 'BEGIN { for (n in ENVIRON) if (n ~ /^(MOAI|CLAUDE|CODEX)_/) print n }'; }

# 3. Scrub by live enumeration of the three families (never a typed list of names). The two mutant modes break it on purpose.
case $mode in
    neg-scrub)
        echo "scrub: DISABLED (negative control)"
        ;;
    typed-list)
        unset MOAI_CLAUDE_BIN MOAI_CANARY_DIR
        echo "scrub: typed list of 2 names (mutant)"
        ;;
    *)
        scrubbed=0
        for name in $(family_names); do
            unset "$name"
            scrubbed=$((scrubbed + 1))
        done
        echo "scrub: enumerated and unset $scrubbed names"
        left=$(family_names)
        [ -z "$left" ] || { echo "ABORT: scrub left names behind: $left" >&2; exit 2; }
        ;;
esac

# 4. Set what the run needs: its own PATH, scratch homes and working directory, the stubs, the decoy roots. The two
#    negative controls that start inside a project put the pin there (llm.claude_bin names the recorder).
shim=$S/shim
mkdir -p "$shim" "$S/empty-shim" "$S/tmp" "$S/work" "$S/claude-home" "$S/codex-home" "$S/moai-home"
PATH="$shim:/usr/bin:/bin"
TMPDIR=$S/tmp
CLAUDE_CONFIG_DIR=$S/claude-home
CODEX_HOME=$S/codex-home
MOAI_HOME=$S/moai-home
HARNESS_DECOY_GOBIN=$S/decoy-gobin
HARNESS_DECOY_GOPATH=$S/decoy-gopath
export PATH TMPDIR CLAUDE_CONFIG_DIR CODEX_HOME MOAI_HOME HARNESS_DECOY_GOBIN HARNESS_DECOY_GOPATH
mkdir -p "$HARNESS_DECOY_GOBIN" "$HARNESS_DECOY_GOPATH/bin"
case $mode in
    neg-scrub|neg-cwd)
        mkdir -p "$S/proj/.moai/config/sections"
        printf 'llm:\n  claude_bin: %s\n' "$S/recorder" > "$S/proj/.moai/config/sections/llm.yaml"
        cd "$S/proj"
        ;;
    *) cd "$S/work" ;;
esac

for tool in claude codex; do
    cat > "$shim/$tool" <<'EOF'
#!/bin/sh
echo "$(basename "$0") $*" >> "${HARNESS_CALLS:-/dev/null}"
EOF
done
cat > "$shim/go" <<'EOF'
#!/bin/sh
[ -z "${HARNESS_GO_LOG:-}" ] || echo "$*" >> "$HARNESS_GO_LOG"
case "$*" in
    "env GOBIN") echo "$HARNESS_DECOY_GOBIN" ;;
    "env GOPATH") echo "$HARNESS_DECOY_GOPATH" ;;
    *) echo "stub go: refusing: $*" >&2; exit 1 ;;
esac
EOF
cat > "$shim/curl" <<'EOF'
#!/bin/sh
out=""; url=""
while [ $# -gt 0 ]; do
    case $1 in -o) out=$2; shift 2 ;; http*) url=$1; shift ;; *) shift ;; esac
done
case $url in *.tar.gz) cp "$HARNESS_ARCHIVE" "$out" ;; *) exit 22 ;; esac
EOF
chmod +x "$shim/claude" "$shim/codex" "$shim/go" "$shim/curl"

canary_entries() { find "$S/canary" -print | LC_ALL=C sort; }
canary_before=$(canary_entries | sha256)
canary_before_n=$(canary_entries | wc -l | tr -d ' ')

# --- helpers ---------------------------------------------------------------------------------------------------

mk_archive() { # <case> <kind>: the archive the stub curl serves; kind real = the binary under test
    d=$S/arc-$1
    mkdir -p "$d"
    case $2 in
        real) cp "$moai_bin" "$d/moai" ;;
        fail-verb) printf '#!/bin/sh\necho "moai $*" >> "${HARNESS_CALLS:-/dev/null}"\nexit 1\n' > "$d/moai" ;;
        old-binary) cat > "$d/moai" <<'EOF'
#!/bin/sh
echo "moai $*" >> "${HARNESS_CALLS:-/dev/null}"
echo 'Unknown command "plugin" for "moai".' >&2
exit 1
EOF
            ;;
    esac
    chmod +x "$d/moai"
    tar -czf "$S/arc-$1.tar.gz" -C "$d" moai
}

add_reason() { reasons="$reasons${reasons:+; }$1"; }

abort() { echo "ABORT: $1 (the installer was not run)" >&2; exit 2; }

# finish_case <name>: record PASS or FAIL for the reasons collected so far.
finish_case() {
    if [ -n "$reasons" ]; then st=FAIL; else st=PASS; fi
    printf '%s\t%s\t%s\n' "$st" "$1" "$reasons" >> "$S/results.tsv"
}

# count_calls <log> <field-2-word>: recorded calls whose first argument is that word (`plugin`), so a tool invoked for
# another reason does not pollute a case.
count_calls() { awk -v w="$2" '$2 == w' "$1" 2>/dev/null | wc -l | tr -d ' '; }

# The four vectors the step runs through the real binary (the stub tools record them): both tools, add before install.
assert_four_vectors() { # <calls log>
    n=$(count_calls "$1" plugin)
    [ "$n" -eq 4 ] || add_reason "plugin-calls=$n(want 4)"
    for want in "claude plugin marketplace add modu-ai/moai-adk" "claude plugin install moai@moai-adk" \
        "codex plugin marketplace add modu-ai/moai-adk" "codex plugin add moai@moai-adk"; do
        grep -qxF "$want" "$1" 2>/dev/null || add_reason "missing[$want]"
    done
}

# run_verb <name> [VAR=value ...]: run the real verb from the harness working directory; output in out-<name>.txt and the
# recorded stub calls in calls-<name>.log; the exit status lands in verb_rc.
run_verb() {
    vname=$1; shift
    : > "$S/calls-$vname.log"
    verb_rc=0
    env HARNESS_CALLS="$S/calls-$vname.log" "$@" "$moai_bin" plugin install > "$S/out-$vname.txt" 2>&1 || verb_rc=$?
}

# BI-2: the decoys are what keeps a flag-less install.sh away from the real $GOBIN, $GOPATH/bin and $HOME/.local/bin, so
# they are asserted, not assumed, immediately before every installer run: both exist, both lie under the scratch root
# (resolved), `go` is the stub, and the stub answers `go env GOBIN` / `go env GOPATH` with exactly those directories.
preflight_decoys() {
    for decoy in "$HARNESS_DECOY_GOBIN" "$HARNESS_DECOY_GOPATH/bin"; do
        [ -d "$decoy" ] || abort "decoy directory is missing: $decoy"
        resolved=$(cd "$decoy" && pwd -P)
        case $resolved in "$S"/*) ;; *) abort "decoy directory is outside the scratch root: $resolved" ;; esac
    done
    [ "$(command -v go)" = "$shim/go" ] || abort "go does not resolve to the stub: $(command -v go)"
    [ "$(go env GOBIN)" = "$HARNESS_DECOY_GOBIN" ] || abort "the stub go does not answer GOBIN with the decoy"
    [ "$(go env GOPATH)" = "$HARNESS_DECOY_GOPATH" ] || abort "the stub go does not answer GOPATH with the decoy"
}

# --- the five isolation cases (AC-025) ---------------------------------------------------------------------------

# env-scrubbed, cwd-has-no-project and resolves-to-stubs judge the environment the product cases run in.
isolation_static() {
    # a name of the three families left in the environment, other than the three scratch homes this harness sets itself
    reasons=""
    leftover=$(family_names | grep -v -x -E 'CLAUDE_CONFIG_DIR|CODEX_HOME|MOAI_HOME' || true)
    if [ -n "$leftover" ]; then
        add_reason "$(printf '%s\n' "$leftover" | wc -l | tr -d ' ') name(s) left: $(printf '%s\n' "$leftover" | head -3 | tr '\n' ' ')"
    fi
    finish_case isolation-env-scrubbed

    # no .moai in the working directory or any ancestor (a pin outranks a PATH shim, and an environment scrub cannot remove it)
    reasons=""
    dir=$(pwd -P)
    while :; do
        if [ -e "$dir/.moai" ]; then add_reason "project-found($dir/.moai)"; break; fi
        [ "$dir" != / ] || break
        dir=$(dirname "$dir")
    done
    finish_case isolation-cwd-has-no-project

    reasons=""
    for tool in claude codex go curl; do
        [ "$(command -v "$tool")" = "$shim/$tool" ] || add_reason "$tool-resolves-to-$(command -v "$tool" || echo nothing)"
    done
    finish_case isolation-resolves-to-stubs
}

# poisoned-pin-never-executed: the real verb runs once; the recorder behind the pin must not, and the stub claude must
# (the positive control: the verb did reach a tool, and it was the stub).
isolation_pin() {
    run_verb isolation-pin
    reasons=""
    recorder_runs=$(wc -l < "$S/recorder.log" | tr -d ' ')
    stub_claude=$(awk '$1 == "claude" && $2 == "plugin"' "$S/calls-isolation-pin.log" | wc -l | tr -d ' ')
    [ "$recorder_runs" -eq 0 ] || add_reason "recorder-executed($recorder_runs)"
    [ "$stub_claude" -ge 1 ] || add_reason "stub-claude-never-called"
    finish_case isolation-poisoned-pin-never-executed
}

# real-home-unchanged: the canary (the stand-in for a real profile) and the protected real roots, after everything ran.
isolation_real_home() {
    reasons=""
    canary_after=$(canary_entries | sha256)
    canary_after_n=$(canary_entries | wc -l | tr -d ' ')
    [ "$canary_before" = "$canary_after" ] || add_reason "canary-changed(entries $canary_before_n -> $canary_after_n)"
    after=$(sh "$hashsh" --roots-file "$S/roots" --dump "$S/after.txt")
    [ "$before" = "$after" ] || add_reason "real-roots-changed"
    finish_case isolation-real-home-unchanged
}

# --- the verb cases (AC-019 (c)) ----------------------------------------------------------------------------------

verb_cases() {
    run_verb verb-install-all-tools
    reasons=""
    [ "$verb_rc" -eq 0 ] || add_reason "exit-$verb_rc"
    assert_four_vectors "$S/calls-verb-install-all-tools.log"
    finish_case verb-install-all-tools

    # no tool on PATH: the precondition is shown first, so a claude that resolves outside the shims cannot hide behind a pass
    run_verb verb-no-tools-exit-0 PATH="$S/empty-shim:/usr/bin:/bin"
    reasons=""
    found=$(env PATH="$S/empty-shim:/usr/bin:/bin" sh -c 'command -v claude; command -v codex' 2>/dev/null || true)
    [ -z "$found" ] || add_reason "a tool resolves outside the shims: $found"
    [ "$verb_rc" -eq 0 ] || add_reason "exit-$verb_rc"
    n=$(count_calls "$S/calls-verb-no-tools-exit-0.log" plugin)
    [ "$n" -eq 0 ] || add_reason "plugin-calls=$n(want 0)"
    skips=$(grep -c 'not found on PATH' "$S/out-verb-no-tools-exit-0.txt" || true)
    [ "$skips" -eq 2 ] || add_reason "skip-lines=$skips(want 2)"
    finish_case verb-no-tools-exit-0

    run_verb verb-optout-zero-calls MOAI_SKIP_PLUGIN_INSTALL=1
    reasons=""
    [ "$verb_rc" -eq 0 ] || add_reason "exit-$verb_rc"
    n=$(count_calls "$S/calls-verb-optout-zero-calls.log" plugin)
    [ "$n" -eq 0 ] || add_reason "plugin-calls=$n(want 0)"
    finish_case verb-optout-zero-calls
}

# --- the installer cases (AC-018 (a), AC-025 (f)) -----------------------------------------------------------------

# run_case <name> <archive-kind> <with|without> [VAR=value]: install with install.sh (flag with = --install-dir passed),
# then judge the three install-directory assertions of AC-018 (a) plus the case's own.
run_case() {
    name=$1; kind=$2; flag=$3; extra=${4:-}
    inst=$S/inst-$name/bin
    mkdir -p "$S/inst-$name"
    mk_archive "$name" "$kind"
    preflight_decoys
    echo "$name" >> "$S/installer-invocations.log" # recording wrapper: every installer execution passes this line
    rc=0
    if [ "$flag" = with ]; then
        env HARNESS_ARCHIVE="$S/arc-$name.tar.gz" HARNESS_CALLS="$S/calls-$name.log" HARNESS_GO_LOG="$S/go-$name.log" $extra \
            bash "$repo/install.sh" --version 9.9.9 --install-dir "$inst" > "$S/out-$name.txt" 2>&1 || rc=$?
    else
        env HARNESS_ARCHIVE="$S/arc-$name.tar.gz" HARNESS_CALLS="$S/calls-$name.log" HARNESS_GO_LOG="$S/go-$name.log" $extra \
            bash "$repo/install.sh" --version 9.9.9 > "$S/out-$name.txt" 2>&1 || rc=$?
    fi

    reasons=""
    { [ "$rc" -eq 0 ] && grep -q 'Installation complete!' "$S/out-$name.txt"; } || add_reason "installer-did-not-complete(rc=$rc)"
    # (1) go resolves to the harness stub
    [ "$(command -v go)" = "$shim/go" ] || add_reason "go-not-stub($(command -v go))"
    # (2) the installed binary exists and lies under this case's own directory (both sides resolved: macOS /var is a link)
    if [ -f "$inst/moai" ] && [ ! -L "$inst/moai" ]; then
        casedir=$(cd "$S/inst-$name" && pwd -P)
        resolved=$(cd "$inst" && pwd -P)/moai
        case $resolved in "$casedir"/*) ;; *) add_reason "installed-path-not-under-case-dir($resolved)" ;; esac
    else
        add_reason "installed-path-not-under-case-dir(no moai at $inst)"
    fi
    # (3) neither decoy default root holds a moai (install.sh installs into <GOPATH>/bin, not <GOPATH>)
    for held in "$HARNESS_DECOY_GOBIN/moai" "$HARNESS_DECOY_GOPATH/bin/moai"; do
        [ ! -e "$held" ] || add_reason "decoy-holds-moai($held)"
    done
    case $name in
        installer-calls-verb-by-installed-path)
            # the install directory is not on PATH, so a bare `moai` records nothing; the four vectors prove the installed path
            case ":$PATH:" in *":$inst:"*) add_reason "install-dir-is-on-PATH" ;; esac
            assert_four_vectors "$S/calls-$name.log"
            ;;
        installer-optout)
            n=$(count_calls "$S/calls-$name.log" plugin)
            [ "$n" -eq 0 ] || add_reason "plugin-calls-recorded($n)"
            ;;
        installer-set-e-guard|installer-old-binary-unknown-verb)
            # the verb was called (a script that never calls it would pass "exit 0" vacuously) and it failed
            n=$(awk '$1 == "moai" && $2 == "plugin" && $3 == "install"' "$S/calls-$name.log" 2>/dev/null | wc -l | tr -d ' ')
            [ "$n" -eq 1 ] || add_reason "verb-calls=$n(want 1)"
            if [ "$name" = installer-old-binary-unknown-verb ]; then
                grep -q 'Unknown command "plugin" for "moai"' "$S/out-$name.txt" || add_reason "unknown-verb-text-not-seen"
            fi
            ;;
    esac
    finish_case "$name"
}

run_installer_cases() { # <with|without>
    run_case installer-calls-verb-by-installed-path real "$1"
    [ "$mode" != omit-flag ] || clean_decoys
    run_case installer-optout real "$1" MOAI_SKIP_PLUGIN_INSTALL=1
    [ "$mode" != omit-flag ] || clean_decoys
    run_case installer-set-e-guard fail-verb "$1"
    [ "$mode" != omit-flag ] || clean_decoys
    run_case installer-old-binary-unknown-verb old-binary "$1"
    [ "$mode" != omit-flag ] || clean_decoys
}

clean_decoys() { # only the omit-flag control lets a binary land in a decoy on purpose; clear it between cases
    rm -f "$HARNESS_DECOY_GOBIN/moai" "$HARNESS_DECOY_GOPATH/bin/moai"
}

# --- judging ---------------------------------------------------------------------------------------------------

status_of() { awk -F'\t' -v n="$1" '$2 == n { print $1 }' "$S/results.tsv"; }
reason_of() { awk -F'\t' -v n="$1" '$2 == n { print $3 }' "$S/results.tsv"; }

finish() { # <exit status so far>: the closing real-roots comparison
    status=$1
    if [ "$before" = "$after" ]; then
        echo "LEAK=0 (real roots unchanged: $after)"
    else
        echo "LEAK=1 (real roots changed)"
        echo "before: $before"
        echo "after:  $after"
        diff "$S/before.txt" "$S/after.txt" || true
        status=1
    fi
    exit "$status"
}

# judge_negative <label> <red names> <green names>: the broken run is right only when exactly the red names failed and
# the green names passed. An omitted --install-dir adds a stricter test for its red names: the install must have landed
# in a decoy and the stub go must have been consulted for both roots.
judge_negative() {
    label=$1; red=$2; green=$3
    ok=1
    for name in $red; do
        st=$(status_of "$name")
        reasons=$(reason_of "$name")
        extra_ok=1
        if [ "$mode" = omit-flag ]; then
            landed=no; consulted=no
            case $reasons in *decoy-holds-moai*) landed=yes ;; esac
            if grep -qx 'env GOBIN' "$S/go-$name.log" 2>/dev/null && grep -qx 'env GOPATH' "$S/go-$name.log" 2>/dev/null; then consulted=yes; fi
            { [ "$landed" = yes ] && [ "$consulted" = yes ]; } || extra_ok=0
        fi
        if [ "$st" = FAIL ] && [ "$extra_ok" -eq 1 ]; then
            echo "RED $name: $reasons"
        else
            echo "green $name ($st): expected red${reasons:+ ($reasons)}"
            ok=0
        fi
    done
    for name in $green; do
        st=$(status_of "$name")
        if [ "$st" = PASS ]; then
            echo "green $name"
        else
            echo "RED $name: expected green ($(reason_of "$name"))"
            ok=0
        fi
    done
    echo "recorder runs: $(wc -l < "$S/recorder.log" | tr -d ' '); canary entries: $canary_before_n -> $canary_after_n"
    if [ "$ok" -eq 1 ]; then
        echo "RESULT $label: red set and green set are exactly the expected ones"
        finish 0
    fi
    echo "RESULT $label: NOT the expected red set"
    finish 1
}

iso_names="isolation-env-scrubbed isolation-cwd-has-no-project isolation-resolves-to-stubs isolation-poisoned-pin-never-executed isolation-real-home-unchanged"
verb_names="verb-install-all-tools verb-no-tools-exit-0 verb-optout-zero-calls"
inst_names="installer-calls-verb-by-installed-path installer-optout installer-set-e-guard installer-old-binary-unknown-verb"

# --- modes -----------------------------------------------------------------------------------------------------

: > "$S/results.tsv"
case $mode in
    normal)
        isolation_static
        isolation_pin
        verb_cases
        run_installer_cases with
        isolation_real_home
        pass=0; fail=0
        for name in $iso_names $verb_names $inst_names; do
            if [ "$(status_of "$name")" = PASS ]; then
                echo "PASS $name"
                pass=$((pass + 1))
            else
                echo "FAIL $name: $(reason_of "$name")"
                fail=$((fail + 1))
            fi
        done
        echo "RESULT pass=$pass fail=$fail"
        [ "$fail" -eq 0 ] && finish 0
        finish 1
        ;;
    neg-scrub)
        isolation_static; isolation_pin; isolation_real_home
        judge_negative negative-control \
            "isolation-env-scrubbed isolation-cwd-has-no-project isolation-poisoned-pin-never-executed isolation-real-home-unchanged" \
            "isolation-resolves-to-stubs"
        ;;
    neg-cwd)
        isolation_static; isolation_pin; isolation_real_home
        judge_negative negative-control-cwd \
            "isolation-cwd-has-no-project isolation-poisoned-pin-never-executed" \
            "isolation-env-scrubbed isolation-resolves-to-stubs isolation-real-home-unchanged"
        ;;
    typed-list)
        isolation_static; isolation_pin; isolation_real_home
        judge_negative typed-list-mutant \
            "isolation-env-scrubbed" \
            "isolation-cwd-has-no-project isolation-resolves-to-stubs isolation-poisoned-pin-never-executed isolation-real-home-unchanged"
        ;;
    omit-flag)
        # The four installer cases without --install-dir: the install lands in a decoy, so all four must go red, each by the
        # decoy assertion (a red caused only by a missing path would not prove the landing), and the five isolation cases
        # must stay green.
        isolation_static; isolation_pin
        run_installer_cases without
        isolation_real_home
        judge_negative negative-control-install-dir "$inst_names" "$iso_names"
        ;;
    decoy-missing)
        # Remove one decoy, run an installer case with the flag omitted, and require that it aborts before the installer
        # executes: proved by the recording wrapper (installer-invocations.log), not by a lack of output.
        ok=1
        for victim in "$HARNESS_DECOY_GOBIN" "$HARNESS_DECOY_GOPATH/bin"; do
            mkdir -p "$HARNESS_DECOY_GOBIN" "$HARNESS_DECOY_GOPATH/bin"
            rmdir "$victim"
            : > "$S/installer-invocations.log"
            out_rc=0
            out=$( (run_case installer-optout real without MOAI_SKIP_PLUGIN_INSTALL=1) 2>&1 ) || out_rc=$?
            runs=$(wc -l < "$S/installer-invocations.log" | tr -d ' ')
            case $out in *"ABORT:"*) aborted=yes ;; *) aborted=no ;; esac
            # status 2 is abort's own: a different failure that merely stopped the run would not prove the guard fired
            if [ "$out_rc" -eq 2 ] && [ "$aborted" = yes ] && [ "$runs" -eq 0 ]; then
                echo "PASS decoy-missing-aborts-before-installer (removed ${victim#$S/}: rc=$out_rc, installer runs=$runs)"
            else
                echo "FAIL decoy-missing-aborts-before-installer (removed ${victim#$S/}: rc=$out_rc, aborted=$aborted, installer runs=$runs)"
                ok=0
            fi
            clean_decoys
        done
        after=$(sh "$hashsh" --roots-file "$S/roots" --dump "$S/after.txt")
        if [ "$ok" -eq 1 ]; then
            echo "RESULT negative-control-decoy-missing: the installer never ran with a decoy missing"
            finish 0
        fi
        echo "RESULT negative-control-decoy-missing: NOT proven, a run did not abort with status 2 before the installer"
        finish 1
        ;;
esac
