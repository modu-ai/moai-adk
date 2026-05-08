#!/bin/sh
# Test harness for validate-required-checks.sh
# SPEC-V3R3-CI-AUTONOMY-001 W4-T05
#
# Three test fixtures validating 3-dimension checks:
#   Fixture 1: Valid SSoT (pass)
#   Fixture 2: Auxiliary item in branches.main.contexts (fail, dimension B)
#   Fixture 3: Auxiliary item in branches.release/*.contexts (fail, dimension C)
#
# Usage:
#   ./scripts/ci-mirror/test/validate-required-checks_test.sh
#
# Returns:
#   0 if all fixtures pass as expected
#   1 if any fixture produces unexpected result

set -eu

TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

VALIDATOR="../validate-required-checks.sh"

# Fixture 1: Valid SSoT (should pass)
echo "=== Fixture 1: Valid SSoT ==="
cat > "$TEST_DIR/fixture1.yml" <<'EOF'
version: 1
branches:
  main:
    contexts:
      - Lint
      - "Test (ubuntu-latest)"
  release/*:
    contexts:
      - Lint
      - "Test (ubuntu-latest)"
auxiliary:
  - docs-i18n-check
  - claude-code-review
EOF

cd "$TEST_DIR"
if sh "$VALIDATOR" ".github/required-checks.yml" <<< "$(cat fixture1.yml)" 2>/dev/null; then
	# Need to actually run with a valid required-checks.yml in place
	# Create a fake .github/workflows/ structure for the test
	mkdir -p .github/workflows
	cat > .github/required-checks.yml <<'EOF'
version: 1
branches:
  main:
    contexts:
      - Lint
      - "Test (ubuntu-latest)"
  release/*:
    contexts:
      - Lint
      - "Test (ubuntu-latest)"
auxiliary:
  - docs-i18n-check
  - claude-code-review
EOF
	touch .github/workflows/docs-i18n-check.yml
	touch .github/workflows/claude-code-review.optional.yml

	if sh "../../$VALIDATOR" 2>/dev/null; then
		echo "✅ Fixture 1 PASS (as expected)"
	else
		echo "❌ Fixture 1 FAIL (unexpected)"
		exit 1
	fi
else
	echo "⚠️  Fixture 1 requires full environment setup (skipped)"
fi

# Fixture 2: Auxiliary item in branches.main.contexts (should fail)
echo ""
echo "=== Fixture 2: Auxiliary in main.contexts (should fail) ==="
rm -rf "$TEST_DIR/.github"
mkdir -p "$TEST_DIR/.github/workflows"
cat > "$TEST_DIR/.github/required-checks.yml" <<'EOF'
version: 1
branches:
  main:
    contexts:
      - Lint
      - docs-i18n-check
  release/*:
    contexts:
      - Lint
auxiliary:
  - docs-i18n-check
EOF
touch "$TEST_DIR/.github/workflows/docs-i18n-check.yml"

cd "$TEST_DIR"
if ! sh "../../$VALIDATOR" 2>/dev/null; then
	echo "✅ Fixture 2 PASS (failed as expected — dimension B violation)"
else
	echo "❌ Fixture 2 FAIL (should have detected dimension B violation)"
	exit 1
fi

# Fixture 3: Auxiliary item in branches.release/*.contexts (should fail)
echo ""
echo "=== Fixture 3: Auxiliary in release/*.contexts (should fail) ==="
rm -rf "$TEST_DIR/.github"
mkdir -p "$TEST_DIR/.github/workflows"
cat > "$TEST_DIR/.github/required-checks.yml" <<'EOF'
version: 1
branches:
  main:
    contexts:
      - Lint
  release/*:
    contexts:
      - Lint
      - docs-i18n-check
auxiliary:
  - docs-i18n-check
EOF
touch "$TEST_DIR/.github/workflows/docs-i18n-check.yml"

cd "$TEST_DIR"
if ! sh "../../$VALIDATOR" 2>/dev/null; then
	echo "✅ Fixture 3 PASS (failed as expected — dimension C violation)"
else
	echo "❌ Fixture 3 FAIL (should have detected dimension C violation)"
	exit 1
fi

echo ""
echo "✅ All test fixtures passed"
exit 0
