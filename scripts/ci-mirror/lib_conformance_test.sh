#!/bin/sh
# scripts/ci-mirror/lib_conformance_test.sh — Structural conformance tests for lib/*.sh
# Verifies all language modules follow the required pattern.
# Run with: sh scripts/ci-mirror/lib_conformance_test.sh
set -eu

ROOT="${1:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
LIB_DIR="$ROOT/scripts/ci-mirror/lib"
PASS=0
FAIL=0

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

# Find all language modules (exclude _common.sh which is a helper, not a language module).
for f in "$LIB_DIR"/*.sh; do
    [ -f "$f" ] || continue
    name="$(basename "$f")"
    [ "$name" = "_common.sh" ] && continue

    # Conformance check 1: starts with POSIX shebang #!/bin/sh
    if head -1 "$f" | grep -q '^#!/bin/sh'; then
        check "$name: starts with #!/bin/sh" ok
    else
        check "$name: starts with #!/bin/sh" fail
    fi

    # Conformance check 2: contains command -v for toolchain check
    if grep -q 'command -v' "$f"; then
        check "$name: contains 'command -v'" ok
    else
        check "$name: contains 'command -v'" fail
    fi

    # Conformance check 3: contains exit 0 (silent skip path)
    if grep -q 'exit 0' "$f"; then
        check "$name: contains 'exit 0' (silent skip path)" ok
    else
        check "$name: contains 'exit 0' (silent skip path)" fail
    fi

    # Conformance check 4: does NOT contain bash-isms ([[ or ${var//])
    if grep -qE '\[\[|'"\$\{[^}]*//"'}' "$f"; then
        check "$name: NO bash-isms detected" fail
    else
        check "$name: NO bash-isms detected" ok
    fi
done

printf '\n%d passed, %d failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
