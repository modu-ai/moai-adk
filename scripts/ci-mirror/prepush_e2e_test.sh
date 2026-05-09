#!/bin/sh
# scripts/ci-mirror/prepush_e2e_test.sh — End-to-end test for the pre-push hook
#
# NOTE: This test requires a working git configuration and is intended for
# manual verification only. It is NOT run in CI (requires git user config).
#
# Usage: sh scripts/ci-mirror/prepush_e2e_test.sh
#
# What it tests:
#   1. Create a temp git repo with git config.
#   2. Install the pre-push hook via moai binary (or copy directly).
#   3. Add a commit.
#   4. Attempt git push --dry-run.
#   5. Assert hook ran and produced expected output.
#
# Documented but not executed in automated CI.
set -eu

ROOT="${1:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"

printf '[prepush-e2e] This is a manual E2E test. Run interactively after moai install.\n'
printf '[prepush-e2e] Steps:\n'
printf '  1. cd /tmp && git init prepush-test && cd prepush-test\n'
printf '  2. git config user.email "test@example.com" && git config user.name "Test"\n'
printf '  3. cp %s/internal/template/templates/.git_hooks/pre-push .git/hooks/pre-push\n' "$ROOT"
printf '  4. chmod +x .git/hooks/pre-push\n'
printf '  5. echo "hello" > README.md && git add . && git commit -m "test: initial"\n'
printf '  6. git remote add origin https://github.com/example/repo.git\n'
printf '  7. SKIP_MOAI_PREPUSH=1 git push --dry-run (should bypass + exit 0)\n'
printf '  8. git push --dry-run (should run make ci-local)\n'
printf '  9. Check .moai/logs/prepush-bypass.log for invocation record\n'
printf '[prepush-e2e] See scripts/ci-mirror/prepush_e2e_test.sh for details.\n'
exit 0
