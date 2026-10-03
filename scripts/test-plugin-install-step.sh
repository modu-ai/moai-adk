#!/bin/sh
# Stub harness for the install scripts (SPEC-PLUGIN-MARKETPLACE-001, REQ-025, AC-018 (a), AC-025 (f), design.md section 6).
#
# Usage: test-plugin-install-step.sh [--negative-control-install-dir | --negative-control-decoy-missing] <path-to-moai>
#
# Scope of this file today (task R0, binding instruction BI-2 of the plan-audit): the environment scrub, the stubs, the
# decoy install roots and the four installer-* cases with their negative controls. The isolation-* and verb-* cases and
# the other negative-control flags belong to M3 and are not here yet.
#
# Why the decoys matter: without --install-dir, install.sh installs into `go env GOBIN`, else `$GOPATH/bin`, else
# `$HOME/.local/bin`. A stub `go` answers the first two with decoy directories inside the scratch. `$HOME/.local/bin`
# (a real directory) is reached only when BOTH decoys are absent, so the harness asserts that both exist, and aborts
# before the installer runs if one does not. Nothing here sets HOME.
set -eu

mode=normal
while [ $# -gt 0 ]; do
    case $1 in
        --negative-control-install-dir) mode=omit-flag; shift ;;
        --negative-control-decoy-missing) mode=decoy-missing; shift ;;
        -*) echo "usage: $0 [--negative-control-install-dir | --negative-control-decoy-missing] <path-to-moai>" >&2; exit 2 ;;
        *) break ;;
    esac
done
[ $# -eq 1 ] || { echo "usage: $0 [--negative-control-install-dir | --negative-control-decoy-missing] <path-to-moai>" >&2; exit 2; }

repo=$(cd "$(dirname "$0")/.." && pwd -P)
hashsh=$repo/scripts/protected-set-hash.sh
moai_bin=$1
case $moai_bin in /*) ;; *) moai_bin=$(pwd -P)/$moai_bin ;; esac
[ -x "$moai_bin" ] || { echo "test-plugin-install-step: not an executable file: $moai_bin" >&2; exit 2; }

S=$(mktemp -d)
S=$(cd "$S" && pwd -P)
trap 'rm -rf "$S"' EXIT
trap 'exit 1' INT TERM

# 1. Capture the protected real roots from the caller's own environment BEFORE any scrub, then take the before hash.
sh "$hashsh" --save-roots "$S/roots"
before=$(sh "$hashsh" --roots-file "$S/roots" --dump "$S/before.txt")

# 2. Plant poison in this script's own environment, so the scrub is judged on names it did not choose: a pin naming a
#    recorder, a canary directory, and one variable of each family whose name carries the process id.
mkdir -p "$S/canary"
printf '#!/bin/sh\necho "recorder $*" >> "%s/recorder.log"\n' "$S" > "$S/recorder"
chmod +x "$S/recorder"
export MOAI_CLAUDE_BIN="$S/recorder" MOAI_CANARY_DIR="$S/canary"
export "MOAI_POISON_$$=1" "CLAUDE_POISON_$$=1" "CODEX_POISON_$$=1"

# 3. Scrub by live enumeration of the three families (never a typed list of names).
scrubbed=0
for name in $(awk 'BEGIN { for (n in ENVIRON) if (n ~ /^(MOAI|CLAUDE|CODEX)_/) print n }'); do
    unset "$name"
    scrubbed=$((scrubbed + 1))
done
echo "scrub: enumerated and unset $scrubbed names"
left=$(awk 'BEGIN { for (n in ENVIRON) if (n ~ /^(MOAI|CLAUDE|CODEX)_/) print n }')
[ -z "$left" ] || { echo "ABORT: scrub left names behind: $left" >&2; exit 2; }

# 4. Set what the run needs: its own PATH, scratch homes and working directory, the stubs, the decoy roots.
shim=$S/shim
mkdir -p "$shim" "$S/tmp" "$S/work" "$S/claude-home" "$S/codex-home" "$S/moai-home"
PATH="$shim:/usr/bin:/bin"
TMPDIR=$S/tmp
CLAUDE_CONFIG_DIR=$S/claude-home
CODEX_HOME=$S/codex-home
MOAI_HOME=$S/moai-home
HARNESS_DECOY_GOBIN=$S/decoy-gobin
HARNESS_DECOY_GOPATH=$S/decoy-gopath
export PATH TMPDIR CLAUDE_CONFIG_DIR CODEX_HOME MOAI_HOME HARNESS_DECOY_GOBIN HARNESS_DECOY_GOPATH
mkdir -p "$HARNESS_DECOY_GOBIN" "$HARNESS_DECOY_GOPATH/bin"
cd "$S/work"

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

# --- helpers ---------------------------------------------------------------------------------------------------

mk_archive() { # <case> <kind>: the archive the stub curl serves; kind real = the binary under test
    d=$S/arc-$1
    mkdir -p "$d"
    case $2 in
        real) cp "$moai_bin" "$d/moai" ;;
        fail-verb) printf '#!/bin/sh\necho "moai $*" >> "${HARNESS_CALLS:-/dev/null}"\nexit 1\n' > "$d/moai" ;;
        old-binary) cat > "$d/moai" <<'EOF'
#!/bin/sh
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

# run_case <name> <archive-kind> <with|without> [VAR=value]: install with install.sh (flag with = --install-dir passed),
# then judge the three install-directory assertions of AC-018 (a) plus the case's own. Appends a row to results.tsv.
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
    if [ "$name" = installer-optout ]; then
        plugin_calls=$(awk '$2 == "plugin"' "$S/calls-$name.log" 2>/dev/null | wc -l | tr -d ' ')
        [ "$plugin_calls" -eq 0 ] || add_reason "plugin-calls-recorded($plugin_calls)"
    fi

    if [ -n "$reasons" ]; then
        status=FAIL
    elif [ "$name" = installer-calls-verb-by-installed-path ]; then
        # install.sh does not call the verb yet (M3): only the install-directory assertions ran for this case.
        status=PENDING
        reasons="install-directory assertions pass; the verb-call assertion lands with install.sh in M3"
    else
        status=PASS
    fi
    printf '%s\t%s\t%s\n' "$status" "$name" "$reasons" >> "$S/results.tsv"
}

run_installer_cases() { # <with|without>
    : > "$S/results.tsv"
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

finish() { # <exit status so far>: the closing real-roots comparison
    status=$1
    after=$(sh "$hashsh" --roots-file "$S/roots" --dump "$S/after.txt")
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

# --- modes -----------------------------------------------------------------------------------------------------

case $mode in
    normal)
        run_installer_cases with
        awk -F'\t' '{ if ($1 == "PASS") print "PASS " $2; else print $1 " " $2 ": " $3 }' "$S/results.tsv"
        pass=$(awk -F'\t' '$1 == "PASS"' "$S/results.tsv" | wc -l | tr -d ' ')
        fail=$(awk -F'\t' '$1 == "FAIL"' "$S/results.tsv" | wc -l | tr -d ' ')
        pending=$(awk -F'\t' '$1 == "PENDING"' "$S/results.tsv" | wc -l | tr -d ' ')
        echo "RESULT pass=$pass fail=$fail pending=$pending"
        [ "$fail" -eq 0 ] && finish 0
        finish 1
        ;;
    omit-flag)
        # The four installer cases without --install-dir: the install lands in a decoy, so all four must go red, each by the
        # decoy assertion (a red caused only by a missing path would not prove the landing), and the stub go must have been
        # consulted for both roots.
        run_installer_cases without
        ok=1
        tab=$(printf '\t')
        while IFS=$tab read -r status name reasons; do
            landed=no
            case $reasons in *decoy-holds-moai*) landed=yes ;; esac
            consulted=no
            if grep -qx 'env GOBIN' "$S/go-$name.log" && grep -qx 'env GOPATH' "$S/go-$name.log"; then consulted=yes; fi
            if [ "$status" = FAIL ] && [ "$landed" = yes ] && [ "$consulted" = yes ]; then
                echo "RED $name: $reasons"
            else
                echo "green $name ($status, decoy-landing=$landed, stub-go-consulted=$consulted): expected red by the decoy assertion"
                ok=0
            fi
        done < "$S/results.tsv"
        if [ "$ok" -eq 1 ]; then
            echo "RESULT negative-control-install-dir: red set and green set are exactly the expected ones"
            finish 0
        fi
        echo "RESULT negative-control-install-dir: NOT the expected red set"
        finish 1
        ;;
    decoy-missing)
        # Remove one decoy, run an installer case with the flag omitted, and require that it aborts before the installer
        # executes: proved by the recording wrapper (installer-invocations.log), not by a lack of output.
        ok=1
        for victim in "$HARNESS_DECOY_GOBIN" "$HARNESS_DECOY_GOPATH/bin"; do
            mkdir -p "$HARNESS_DECOY_GOBIN" "$HARNESS_DECOY_GOPATH/bin"
            rmdir "$victim"
            : > "$S/installer-invocations.log"
            : > "$S/results.tsv"
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
        if [ "$ok" -eq 1 ]; then
            echo "RESULT negative-control-decoy-missing: the installer never ran with a decoy missing"
            finish 0
        fi
        echo "RESULT negative-control-decoy-missing: NOT proven, a run did not abort with status 2 before the installer"
        finish 1
        ;;
esac
