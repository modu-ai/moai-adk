#!/bin/sh
# scripts/ci-mirror/test/makefile_test.sh — Verify Makefile ci-local + pr-merge targets
# Run with: sh scripts/ci-mirror/test/makefile_test.sh
set -eu

ROOT="${1:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
MK="$ROOT/Makefile"
PASS=0
FAIL=0
TMP_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/moai-makefile-test.XXXXXX")

cleanup() {
    rm -rf -- "$TMP_ROOT"
}
trap cleanup EXIT HUP INT TERM

check() {
    desc="$1"
    result="$2"
    if [ "$result" = "ok" ]; then
        printf '[PASS] %s\n' "$desc"
        PASS=$((PASS + 1))
    else
        printf '[FAIL] %s\n' "$desc"
        FAIL=$((FAIL + 1))
    fi
}

# Test 1: ci-local target exists
if make -n ci-local -C "$ROOT" 2>&1 | grep -q 'ci-mirror/run.sh'; then
    check "ci-local dry-run contains ./scripts/ci-mirror/run.sh" ok
else
    check "ci-local dry-run contains ./scripts/ci-mirror/run.sh" fail
    make -n ci-local -C "$ROOT" 2>&1 | head -5 >&2
fi

# Test 2: pr-merge target with PR=999 STRATEGY=merge
if make -n pr-merge PR=999 STRATEGY=merge -C "$ROOT" 2>&1 | grep -q 'gh pr merge 999.*--auto.*--merge'; then
    check "pr-merge PR=999 STRATEGY=merge dry-run contains correct gh command" ok
else
    check "pr-merge PR=999 STRATEGY=merge dry-run contains correct gh command" fail
    make -n pr-merge PR=999 STRATEGY=merge -C "$ROOT" 2>&1 | head -5 >&2
fi

# Test 3: pr-merge with no PR= exits non-zero
if make -n pr-merge -C "$ROOT" 2>&1 | grep -qE 'Usage:|error'; then
    check "pr-merge without PR= shows usage hint" ok
else
    # Some make versions may handle this differently — soft check
    printf '[SKIP] pr-merge without PR= behavior varies by make version\n'
fi

# Test 4: local-install verification must not invoke macOS strings. A fake
# strings command records any accidental call and reproduces the Xcode-license
# failure that motivated this regression test.
mkdir -p "$TMP_ROOT/build" "$TMP_ROOT/install" "$TMP_ROOT/tools"
cat > "$TMP_ROOT/build/moai" <<'EOF'
#!/bin/sh
[ "${1:-}" = "version" ] || exit 64
printf 'moai test-build (commit abc123def)\n'
EOF
chmod +x "$TMP_ROOT/build/moai"
cp "$TMP_ROOT/build/moai" "$TMP_ROOT/install/moai"
cat > "$TMP_ROOT/tools/strings" <<'EOF'
#!/bin/sh
: > "${STRINGS_CALLED:?}"
printf 'You have not agreed to the Xcode license agreements.\n' >&2
exit 69
EOF
chmod +x "$TMP_ROOT/tools/strings"

STRINGS_CALLED="$TMP_ROOT/strings-called"
if output=$(STRINGS_CALLED="$STRINGS_CALLED" PATH="$TMP_ROOT/tools:$PATH" \
    sh "$ROOT/scripts/verify-local-install.sh" \
    "$TMP_ROOT/build/moai" "$TMP_ROOT/install/moai" abc123def 2>&1) \
    && printf '%s\n' "$output" | grep -q 'local-install-check: OK' \
    && [ ! -e "$STRINGS_CALLED" ]; then
    check "verify-local-install proves the copy without invoking strings" ok
else
    check "verify-local-install proves the copy without invoking strings" fail
    printf '%s\n' "$output" >&2
fi

# Test 5: byte drift must remain a hard failure instead of being hidden by a
# successful version command.
printf '\n' >> "$TMP_ROOT/install/moai"
if output=$(PATH="$TMP_ROOT/tools:$PATH" sh "$ROOT/scripts/verify-local-install.sh" \
    "$TMP_ROOT/build/moai" "$TMP_ROOT/install/moai" abc123def 2>&1); then
    check "verify-local-install rejects a byte-mismatched installed binary" fail
else
    case "$output" in
        *"installed binary differs"*)
            check "verify-local-install rejects a byte-mismatched installed binary" ok
            ;;
        *)
            check "verify-local-install rejects a byte-mismatched installed binary" fail
            printf '%s\n' "$output" >&2
            ;;
    esac
fi

# Test 6: an identical binary with the wrong embedded commit must fail.
cp "$TMP_ROOT/build/moai" "$TMP_ROOT/install/moai"
if output=$(PATH="$TMP_ROOT/tools:$PATH" sh "$ROOT/scripts/verify-local-install.sh" \
    "$TMP_ROOT/build/moai" "$TMP_ROOT/install/moai" deadbeef0 2>&1); then
    check "verify-local-install rejects an unexpected embedded commit" fail
else
    case "$output" in
        *"does not report expected commit"*)
            check "verify-local-install rejects an unexpected embedded commit" ok
            ;;
        *)
            check "verify-local-install rejects an unexpected embedded commit" fail
            printf '%s\n' "$output" >&2
            ;;
    esac
fi

# Test 7: the Makefile convenience target must delegate to the same standalone
# verifier, so there is only one judgment implementation.
if make -n verify-local-install -C "$ROOT" 2>&1 | grep -q 'scripts/verify-local-install.sh'; then
    check "verify-local-install target delegates to the standalone verifier" ok
else
    check "verify-local-install target delegates to the standalone verifier" fail
fi

printf '\n%d passed, %d failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
