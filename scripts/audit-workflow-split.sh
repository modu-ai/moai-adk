#!/usr/bin/env bash
# audit-workflow-split.sh — Cross-reference audit for SPEC-V3R4-WORKFLOW-SPLIT-001
#
# Greps all `Read workflows/.../...md` references inside
# .claude/skills/moai/workflows/**/*.md, verifies each target file exists,
# returns exit code 0 (PASS) or 1 (FAIL) with line-by-line broken-link report.
#
# Usage: bash scripts/audit-workflow-split.sh [--strict]
#   --strict: exit 1 on ANY issue (same as default; reserved for future use)
#
# Output format on FAIL:
#   BROKEN REF: <file>:<line> → <target>
# Output final line on PASS:
#   ✓ All references valid (N checked, 0 broken)

set -euo pipefail

# Determine project root (script must be in scripts/ under project root)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

WORKFLOWS_DIR="${PROJECT_ROOT}/.claude/skills/moai/workflows"
CHECKED=0
BROKEN=0

# Collect all .md files in the workflows directory (recursive)
if [ ! -d "${WORKFLOWS_DIR}" ]; then
    echo "WARN: workflows directory not found: ${WORKFLOWS_DIR}"
    echo "✓ All references valid (0 checked, 0 broken)"
    exit 0
fi

# Find all .md files in workflows/
while IFS= read -r -d '' mdfile; do
    # Extract lines that match reference patterns:
    # Pattern 1: `Read workflows/{name}/{sub}.md`
    # Pattern 2: `Read .claude/skills/moai/workflows/{name}/{sub}.md`
    while IFS= read -r match; do
        lineno="$(echo "${match}" | cut -d: -f1)"
        linetext="$(echo "${match}" | cut -d: -f2-)"

        # Extract the referenced path from patterns
        # Pattern 1: Read workflows/foo/bar.md
        target=""
        if echo "${linetext}" | grep -qE "Read workflows/[A-Za-z0-9_/-]+\.md"; then
            target="$(echo "${linetext}" | sed -nE 's/.*Read (workflows\/[A-Za-z0-9_/.-]+\.md).*/\1/p' | head -1)"
            if [ -n "${target}" ]; then
                resolved="${WORKFLOWS_DIR}/../${target}"
                resolved="$(cd "${WORKFLOWS_DIR}/.." && pwd)/${target}"
                CHECKED=$((CHECKED + 1))
                if [ ! -f "${resolved}" ]; then
                    echo "BROKEN REF: ${mdfile#${PROJECT_ROOT}/}:${lineno} → ${target}"
                    BROKEN=$((BROKEN + 1))
                fi
            fi
        fi

        # Pattern 2: Read .claude/skills/moai/workflows/foo/bar.md
        if echo "${linetext}" | grep -qE "Read \.claude/skills/moai/workflows/[A-Za-z0-9_/-]+\.md"; then
            target2="$(echo "${linetext}" | sed -nE 's/.*Read (\.claude\/skills\/moai\/workflows\/[A-Za-z0-9_/.-]+\.md).*/\1/p' | head -1)"
            if [ -n "${target2}" ]; then
                resolved2="${PROJECT_ROOT}/${target2}"
                CHECKED=$((CHECKED + 1))
                if [ ! -f "${resolved2}" ]; then
                    echo "BROKEN REF: ${mdfile#${PROJECT_ROOT}/}:${lineno} → ${target2}"
                    BROKEN=$((BROKEN + 1))
                fi
            fi
        fi
    done < <(grep -n "Read workflows/\|Read \.claude/skills/moai/workflows/" "${mdfile}" 2>/dev/null || true)
done < <(find "${WORKFLOWS_DIR}" -name "*.md" -print0 2>/dev/null)

if [ "${BROKEN}" -gt 0 ]; then
    echo "✗ Cross-reference audit FAILED (${CHECKED} checked, ${BROKEN} broken)"
    exit 1
fi

echo "✓ All references valid (${CHECKED} checked, 0 broken)"
exit 0
