#!/bin/bash
set -euo pipefail

# validate-readme.sh - Validate README.md structure and content
# Usage: ./validate-readme.sh [README_FILE]

README_FILE="${1:-README.md}"

if [ ! -f "$README_FILE" ]; then
    echo "❌ ERROR: README file not found: $README_FILE" >&2
    exit 1
fi

echo "🔍 Validating README.md structure..."

# Check file size (< 500 KiB per GitHub limit)
FILE_SIZE=$(wc -c < "$README_FILE")
MAX_SIZE=512000

if [ "$FILE_SIZE" -gt "$MAX_SIZE" ]; then
    echo "⚠️  WARNING: README exceeds 500 KiB GitHub limit (${FILE_SIZE} bytes)" >&2
    echo "   Recommendation: Move detailed content to docs/ directory" >&2
fi

# Check for required sections
REQUIRED_SECTIONS=(
    "^# "           # Title (H1)
    "^## "          # At least one H2 section
    "[Ii]nstall"    # Installation instructions
)

MISSING_SECTIONS=()
for SECTION in "${REQUIRED_SECTIONS[@]}"; do
    if ! grep -qE "$SECTION" "$README_FILE"; then
        MISSING_SECTIONS+=("$SECTION")
    fi
done

if [ ${#MISSING_SECTIONS[@]} -gt 0 ]; then
    echo "❌ ERROR: Missing required sections:" >&2
    for SECTION in "${MISSING_SECTIONS[@]}"; do
        echo "   - $SECTION" >&2
    done
    exit 1
fi

# Check for relative links (warn on absolute localhost paths)
if grep -q "](http://localhost" "$README_FILE"; then
    echo "⚠️  WARNING: Found localhost links (use relative paths)" >&2
fi

# Check for code blocks without language identifiers
UNLABELED_BLOCKS=$(grep -c '^```$' "$README_FILE" || true)
if [ "$UNLABELED_BLOCKS" -gt 0 ]; then
    echo "⚠️  WARNING: Found $UNLABELED_BLOCKS code blocks without language identifiers" >&2
    echo "   Recommendation: Add language (e.g., \`\`\`bash, \`\`\`python)" >&2
fi

# Validate markdown syntax (if markdownlint available)
if command -v markdownlint &> /dev/null; then
    if markdownlint "$README_FILE" 2>&1 | grep -q "error"; then
        echo "⚠️  WARNING: Markdown linting errors detected" >&2
    fi
fi

# Check for badges (recommended)
if ! grep -q "!\[.*\](.*/badge\.svg)" "$README_FILE"; then
    echo "💡 TIP: Consider adding badges (version, license, CI status)" >&2
fi

# Summary
echo ""
echo "✅ README validation complete"
echo "   File size: $(numfmt --to=iec-i --suffix=B "$FILE_SIZE" 2>/dev/null || echo "${FILE_SIZE} bytes")"
echo "   Required sections: ✓"
echo "   Structure: ✓"

exit 0
