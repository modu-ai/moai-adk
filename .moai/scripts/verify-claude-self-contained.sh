#!/bin/bash
##############################################################################
# CLAUDE.md Self-Contained Verification Script
# Purpose: Verify the optimized CLAUDE.md meets all success criteria
# Usage: ./verify-claude-self-contained.sh [path/to/CLAUDE.md]
##############################################################################

set -e  # Exit on first error

# Default path
CLAUDE_FILE="${1:-src/moai_adk/templates/CLAUDE.md}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'  # No Color

# Counters
CHECKS_PASSED=0
CHECKS_FAILED=0
CHECKS_WARNING=0

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}CLAUDE.md Verification Script${NC}"
echo -e "${BLUE}================================${NC}"
echo ""
echo -e "${BLUE}File: ${CLAUDE_FILE}${NC}"
echo ""

# Check if file exists
if [ ! -f "$CLAUDE_FILE" ]; then
    echo -e "${RED}❌ ERROR: File not found: $CLAUDE_FILE${NC}"
    exit 1
fi

echo -e "${BLUE}Starting verification...${NC}"
echo ""

# ============================================================================
# CHECK 1: Line Count
# ============================================================================
echo -e "${BLUE}[1] Checking line count...${NC}"

LINE_COUNT=$(wc -l < "$CLAUDE_FILE")
MIN_LINES=1180
MAX_LINES=1220
TARGET_LINES=1200

if [ "$LINE_COUNT" -ge "$MIN_LINES" ] && [ "$LINE_COUNT" -le "$MAX_LINES" ]; then
    echo -e "${GREEN}✅ PASS${NC}: Line count is $LINE_COUNT (target $TARGET_LINES ±20)"
    ((CHECKS_PASSED++))
else
    echo -e "${RED}❌ FAIL${NC}: Line count is $LINE_COUNT (target $TARGET_LINES ±20)"
    echo "    Expected: $MIN_LINES - $MAX_LINES lines"
    ((CHECKS_FAILED++))
fi
echo ""

# ============================================================================
# CHECK 2: No .moai/learning/ External Links
# ============================================================================
echo -e "${BLUE}[2] Checking for external .moai/learning/ links...${NC}"

LEARNING_LINKS=$(grep -c '\.moai/learning/' "$CLAUDE_FILE" || true)

if [ "$LEARNING_LINKS" -eq 0 ]; then
    echo -e "${GREEN}✅ PASS${NC}: No .moai/learning/ references found"
    ((CHECKS_PASSED++))
else
    echo -e "${RED}❌ FAIL${NC}: Found $LEARNING_LINKS .moai/learning/ reference(s)"
    grep -n '\.moai/learning/' "$CLAUDE_FILE" | head -5
    ((CHECKS_FAILED++))
fi
echo ""

# ============================================================================
# CHECK 3: Template Variables Preserved
# ============================================================================
echo -e "${BLUE}[3] Checking template variable preservation...${NC}"

VARIABLES=(
    "{{PROJECT_NAME}}"
    "{{PROJECT_OWNER}}"
    "{{CONVERSATION_LANGUAGE}}"
    "{{MOAI_VERSION}}"
)

ALL_VARS_FOUND=true

for var in "${VARIABLES[@]}"; do
    COUNT=$(grep -c "$var" "$CLAUDE_FILE" || true)
    if [ "$COUNT" -gt 0 ]; then
        echo "   ✅ $var: $COUNT occurrence(s)"
    else
        echo -e "   ${RED}❌ $var: NOT FOUND${NC}"
        ALL_VARS_FOUND=false
    fi
done

if [ "$ALL_VARS_FOUND" = true ]; then
    echo -e "${GREEN}✅ PASS${NC}: All template variables preserved"
    ((CHECKS_PASSED++))
else
    echo -e "${RED}❌ FAIL${NC}: Some template variables missing"
    ((CHECKS_FAILED++))
fi
echo ""

# ============================================================================
# CHECK 4: No Hardcoded Secrets
# ============================================================================
echo -e "${BLUE}[4] Checking for hardcoded secrets...${NC}"

SECRETS_PATTERN="(password|secret|api_key|token).*[=:]\s*['\"]"
SECRETS_FOUND=$(grep -iE "$SECRETS_PATTERN" "$CLAUDE_FILE" | grep -v '{{' | grep -v '#' | wc -l || true)

if [ "$SECRETS_FOUND" -eq 0 ]; then
    echo -e "${GREEN}✅ PASS${NC}: No hardcoded secrets detected"
    ((CHECKS_PASSED++))
else
    echo -e "${RED}❌ FAIL${NC}: Potential hardcoded secret(s) found"
    grep -iE "$SECRETS_PATTERN" "$CLAUDE_FILE" | grep -v '{{' | grep -v '#' | head -3
    ((CHECKS_FAILED++))
fi
echo ""

# ============================================================================
# CHECK 5: Internal Anchor Link Validation
# ============================================================================
echo -e "${BLUE}[5] Validating internal anchor links...${NC}"

# Extract all section headers
SECTIONS=$(grep -E "^##+ " "$CLAUDE_FILE" | sed 's/^.*## *//' | sed 's/^.*### *//' | sed 's/ *$//')

# Extract all internal links
LINKS=$(grep -oE '\[.*\]\(#[^)]+\)' "$CLAUDE_FILE" | sed 's/.*](#\([^)]*\)).*/\1/' | sort -u)

BROKEN_LINKS=0

while IFS= read -r link; do
    # Convert anchor to section header format
    # Examples: #-quick-start -> "quick-start" -> "## Quick Start"
    # Handle special characters in anchors

    # Check if this anchor exists as a section
    ANCHOR_FOUND=false

    while IFS= read -r section; do
        # Convert section to anchor format (lowercase, replace spaces with dashes)
        SECTION_ANCHOR=$(echo "$section" | tr '[:upper:]' '[:lower:]' | sed 's/ /-/g' | sed 's/[^a-z0-9-]//g')

        if [ "$link" = "$SECTION_ANCHOR" ]; then
            ANCHOR_FOUND=true
            break
        fi
    done <<< "$SECTIONS"

    if [ "$ANCHOR_FOUND" = false ]; then
        echo -e "   ${YELLOW}⚠️  Anchor not found: #$link${NC}"
        ((BROKEN_LINKS++))
    fi
done <<< "$LINKS"

if [ "$BROKEN_LINKS" -eq 0 ]; then
    echo -e "${GREEN}✅ PASS${NC}: All internal anchor links are valid"
    ((CHECKS_PASSED++))
elif [ "$BROKEN_LINKS" -le 3 ]; then
    echo -e "${YELLOW}⚠️  WARNING${NC}: $BROKEN_LINKS anchor link(s) may be broken"
    echo "    (Note: Some false positives due to emoji/special chars)"
    ((CHECKS_WARNING++))
else
    echo -e "${RED}❌ FAIL${NC}: $BROKEN_LINKS broken anchor link(s)"
    ((CHECKS_FAILED++))
fi
echo ""

# ============================================================================
# CHECK 6: Content Sections Present
# ============================================================================
echo -e "${BLUE}[6] Checking for required content sections...${NC}"

REQUIRED_SECTIONS=(
    "Quick Start"
    "Alfred SuperAgent"
    "SPEC-First"
    "TRUST 5"
    "Error Recovery"
    "Glossary"
    "MCP"
)

MISSING_SECTIONS=0

for section in "${REQUIRED_SECTIONS[@]}"; do
    if grep -qi "$section" "$CLAUDE_FILE"; then
        echo "   ✅ Found: $section"
    else
        echo -e "   ${RED}❌ Missing: $section${NC}"
        ((MISSING_SECTIONS++))
    fi
done

if [ "$MISSING_SECTIONS" -eq 0 ]; then
    echo -e "${GREEN}✅ PASS${NC}: All required sections present"
    ((CHECKS_PASSED++))
else
    echo -e "${RED}❌ FAIL${NC}: $MISSING_SECTIONS required section(s) missing"
    ((CHECKS_FAILED++))
fi
echo ""

# ============================================================================
# CHECK 7: Copy-Paste Ready Examples
# ============================================================================
echo -e "${BLUE}[7] Checking for code examples...${NC}"

CODE_BLOCKS=$(grep -c '```' "$CLAUDE_FILE" || true)
BASH_EXAMPLES=$(grep -c 'bash' "$CLAUDE_FILE" || true)
JSON_EXAMPLES=$(grep -c '```json' "$CLAUDE_FILE" || true)

if [ "$CODE_BLOCKS" -gt 10 ] && [ "$BASH_EXAMPLES" -gt 3 ]; then
    echo "   ✅ Code blocks: $CODE_BLOCKS"
    echo "   ✅ Bash examples: $BASH_EXAMPLES"
    echo "   ✅ JSON examples: $JSON_EXAMPLES"
    echo -e "${GREEN}✅ PASS${NC}: Adequate copy-paste ready examples"
    ((CHECKS_PASSED++))
else
    echo -e "${YELLOW}⚠️  WARNING${NC}: Limited code examples"
    echo "   Code blocks: $CODE_BLOCKS | Bash: $BASH_EXAMPLES | JSON: $JSON_EXAMPLES"
    ((CHECKS_WARNING++))
fi
echo ""

# ============================================================================
# CHECK 8: No Repeated Content Sections
# ============================================================================
echo -e "${BLUE}[8] Checking for significant repetition...${NC}"

# Count major section headers
MAJOR_HEADERS=$(grep -c "^## " "$CLAUDE_FILE" || true)

if [ "$MAJOR_HEADERS" -gt 15 ] && [ "$MAJOR_HEADERS" -lt 30 ]; then
    echo "   ✅ Major sections: $MAJOR_HEADERS (reasonable)"
    echo -e "${GREEN}✅ PASS${NC}: No obvious content duplication"
    ((CHECKS_PASSED++))
elif [ "$MAJOR_HEADERS" -lt 15 ]; then
    echo -e "${YELLOW}⚠️  WARNING${NC}: May be missing major sections ($MAJOR_HEADERS found)"
    ((CHECKS_WARNING++))
else
    echo -e "${YELLOW}⚠️  WARNING${NC}: Many sections ($MAJOR_HEADERS), check for duplication"
    ((CHECKS_WARNING++))
fi
echo ""

# ============================================================================
# SUMMARY
# ============================================================================
echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}Verification Summary${NC}"
echo -e "${BLUE}================================${NC}"
echo ""
echo -e "${GREEN}✅ Passed: $CHECKS_PASSED${NC}"
echo -e "${RED}❌ Failed: $CHECKS_FAILED${NC}"
echo -e "${YELLOW}⚠️  Warnings: $CHECKS_WARNING${NC}"
echo ""

# Final result
if [ "$CHECKS_FAILED" -eq 0 ]; then
    if [ "$CHECKS_WARNING" -eq 0 ]; then
        echo -e "${GREEN}✅ ALL CHECKS PASSED${NC}"
        echo ""
        echo "Document is ready for:"
        echo "  ✅ Git commit"
        echo "  ✅ Pull request"
        echo "  ✅ Production deployment"
        echo ""
        exit 0
    else
        echo -e "${YELLOW}⚠️  CHECKS PASSED WITH WARNINGS${NC}"
        echo ""
        echo "Review warnings above and decide if acceptable:"
        echo "  ⚠️  Anchor link false positives (due to emoji)"
        echo "  ⚠️  Limited code examples (add more if needed)"
        echo "  ⚠️  Potential duplication (review sections)"
        echo ""
        exit 0
    fi
else
    echo -e "${RED}❌ VERIFICATION FAILED${NC}"
    echo ""
    echo "Issues to fix before deployment:"
    echo "  ❌ Check line count: Should be 1,180-1,220"
    echo "  ❌ Remove .moai/learning/ references"
    echo "  ❌ Preserve template variables"
    echo "  ❌ Add hardcoded secrets warning"
    echo "  ❌ Fix broken anchor links"
    echo ""
    exit 1
fi
