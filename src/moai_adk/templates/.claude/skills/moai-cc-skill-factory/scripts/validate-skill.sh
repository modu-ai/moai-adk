#!/bin/bash
set -euo pipefail

# Skill Validation Script
# Validates Skill packages against MoAI-ADK quality standards
# Usage: ./validate-skill.sh /path/to/skill-name

SKILL_PATH="${1:-.}"
SCORE=0
MAX_SCORE=85
WARNINGS=0

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((SCORE+=1))
}

log_fail() {
    echo -e "${RED}✗${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
    ((WARNINGS+=1))
}

log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# Header
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Skill Validation: $(basename "$SKILL_PATH")${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo

# === METADATA COMPLETENESS (8 points) ===
echo -e "${BLUE}[Metadata Completeness]${NC}"

if [[ ! -f "$SKILL_PATH/SKILL.md" ]]; then
    log_fail "SKILL.md not found"
    exit 1
fi

# Check name field
if grep -q "^name:" "$SKILL_PATH/SKILL.md"; then
    NAME=$(grep "^name:" "$SKILL_PATH/SKILL.md" | sed 's/^name: *"\(.*\)"/\1/')
    NAME_LEN=${#NAME}
    if [[ $NAME_LEN -le 64 ]]; then
        log_pass "name field present and within limit ($NAME_LEN chars)"
        ((SCORE+=2))
    else
        log_fail "name field exceeds 64 characters ($NAME_LEN chars)"
    fi
else
    log_fail "name field missing"
fi

# Check description field
if grep -q "^description:" "$SKILL_PATH/SKILL.md"; then
    log_pass "description field present"
    ((SCORE+=2))
    DESC=$(grep "^description:" "$SKILL_PATH/SKILL.md" | sed 's/^description: *"\(.*\)"/\1/')
    DESC_LEN=${#DESC}
    if [[ $DESC_LEN -le 1024 ]]; then
        log_pass "description within limit ($DESC_LEN chars)"
        ((SCORE+=1))
    else
        log_fail "description exceeds 1024 characters"
    fi
else
    log_fail "description field missing"
fi

# Check allowed-tools field
if grep -q "^allowed-tools:" "$SKILL_PATH/SKILL.md"; then
    log_pass "allowed-tools field present"
    ((SCORE+=2))
else
    log_warn "allowed-tools field missing (recommended)"
fi

# Check YAML validity
if head -n 10 "$SKILL_PATH/SKILL.md" | grep -q "^---$"; then
    log_pass "YAML frontmatter starts with ---"
    ((SCORE+=1))
else
    log_fail "YAML frontmatter missing ---"
fi

echo

# === CONTENT QUALITY (16 points) ===
echo -e "${BLUE}[Content Quality]${NC}"

# Check file size
LINE_COUNT=$(wc -l < "$SKILL_PATH/SKILL.md")
if [[ $LINE_COUNT -le 500 ]]; then
    log_pass "SKILL.md within size limit ($LINE_COUNT lines)"
    ((SCORE+=2))
else
    log_fail "SKILL.md exceeds 500 lines ($LINE_COUNT lines)"
fi

# Check for examples
if grep -q "^##.*[Ee]xample" "$SKILL_PATH/SKILL.md"; then
    EXAMPLE_COUNT=$(grep -c "^###" "$SKILL_PATH/SKILL.md" || echo 0)
    if [[ $EXAMPLE_COUNT -ge 1 ]]; then
        log_pass "Examples found ($EXAMPLE_COUNT examples)"
        ((SCORE+=2))
    else
        log_warn "Few examples found"
    fi
else
    log_warn "No examples section detected"
fi

# Check for relative paths
if grep -q "\[.*\](.*)" "$SKILL_PATH/SKILL.md"; then
    if grep -q "\[.*\](\.\./\|http\|/\|~" "$SKILL_PATH/SKILL.md"; then
        log_warn "Some links may use absolute or external paths"
        ((WARNINGS+=1))
    else
        log_pass "Links use relative paths"
        ((SCORE+=2))
    fi
else
    log_pass "No external links detected"
    ((SCORE+=2))
fi

# Check for Windows paths
if grep -q "\\\\" "$SKILL_PATH/SKILL.md"; then
    log_fail "Windows-style paths detected (use forward slashes)"
else
    log_pass "No Windows-style paths"
    ((SCORE+=1))
fi

# Check for glossary/terminology
if grep -q "[Gg]lossary\|[Tt]ermino" "$SKILL_PATH/SKILL.md"; then
    log_pass "Glossary or terminology section present"
    ((SCORE+=1))
else
    log_warn "No glossary detected"
fi

# Check for time-sensitive data
if grep -q "20[0-9][0-9]-[0-9][0-9]-[0-9][0-9]\|today\|current date" "$SKILL_PATH/SKILL.md"; then
    log_fail "Time-sensitive information detected"
else
    log_pass "No time-sensitive data"
    ((SCORE+=1))
fi

echo

# === FILE STRUCTURE (6 points) ===
echo -e "${BLUE}[File Structure]${NC}"

# Check directory depth
MAX_DEPTH=$(find "$SKILL_PATH" -type d | sed "s|$SKILL_PATH||" | grep -v "^$" | awk -F/ '{print NF}' | sort -rn | head -1)
if [[ -z $MAX_DEPTH ]]; then
    MAX_DEPTH=1
fi

if [[ $MAX_DEPTH -le 2 ]]; then
    log_pass "Directory structure flat (max depth: $MAX_DEPTH)"
    ((SCORE+=2))
else
    log_fail "Directory structure too nested (max depth: $MAX_DEPTH)"
fi

# Check for common subdirs
if [[ -d "$SKILL_PATH/scripts" ]]; then
    SCRIPT_COUNT=$(find "$SKILL_PATH/scripts" -type f | wc -l)
    log_pass "scripts/ directory present ($SCRIPT_COUNT files)"
    ((SCORE+=1))
fi

if [[ -d "$SKILL_PATH/templates" ]]; then
    TEMPLATE_COUNT=$(find "$SKILL_PATH/templates" -type f | wc -l)
    log_pass "templates/ directory present ($TEMPLATE_COUNT files)"
    ((SCORE+=1))
fi

# Check for supporting docs
if [[ -f "$SKILL_PATH/reference.md" ]]; then
    log_pass "reference.md found"
    ((SCORE+=1))
fi

if [[ -f "$SKILL_PATH/examples.md" ]]; then
    log_pass "examples.md found"
    ((SCORE+=1))
fi

echo

# === MARKDOWN VALIDITY (3 points) ===
echo -e "${BLUE}[Markdown Validity]${NC}"

# Check for proper headers
if grep -q "^#" "$SKILL_PATH/SKILL.md"; then
    log_pass "Headers detected"
    ((SCORE+=1))
else
    log_warn "No headers found"
fi

# Check for code blocks
if grep -q "^\`\`\`" "$SKILL_PATH/SKILL.md"; then
    log_pass "Code blocks present"
    ((SCORE+=1))
else
    log_warn "No code blocks found"
fi

# Check for lists
if grep -q "^- \|^  - " "$SKILL_PATH/SKILL.md"; then
    log_pass "Lists present"
    ((SCORE+=1))
else
    log_warn "No lists detected"
fi

echo

# === SCRIPT VALIDITY (if present) ===
if [[ -d "$SKILL_PATH/scripts" ]]; then
    echo -e "${BLUE}[Script Validation]${NC}"

    for script in "$SKILL_PATH"/scripts/*; do
        if [[ -f "$script" && ! "$script" =~ \.gitkeep$ ]]; then
            SCRIPT_NAME=$(basename "$script")

            # Check shebang
            if head -n 1 "$script" | grep -q "^#!/"; then
                log_pass "$SCRIPT_NAME has shebang"
                ((SCORE+=1))
            else
                log_warn "$SCRIPT_NAME missing shebang"
            fi

            # Check for error handling (Bash scripts)
            if [[ "$script" =~ \.sh$ ]]; then
                if grep -q "set -euo pipefail\|set -e" "$script"; then
                    log_pass "$SCRIPT_NAME has error handling"
                    ((SCORE+=1))
                else
                    log_warn "$SCRIPT_NAME lacks error handling"
                fi
            fi
        fi
    done
    echo
fi

# === SECURITY CHECK ===
echo -e "${BLUE}[Security Check]${NC}"

# Check for credentials
if grep -q "password\|api_key\|secret\|token.*=" "$SKILL_PATH/SKILL.md" -i; then
    log_warn "Potential credentials detected in SKILL.md"
    ((WARNINGS+=1))
else
    log_pass "No obvious credentials"
    ((SCORE+=1))
fi

# Check for no email
if grep -q "[a-z0-9._%+-]*@[a-z0-9.-]*\.[a-z]" "$SKILL_PATH/SKILL.md"; then
    log_warn "Email addresses found (privacy concern?)"
    ((WARNINGS+=1))
else
    log_pass "No email addresses"
    ((SCORE+=1))
fi

echo

# === SUMMARY ===
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "Score: ${BLUE}$SCORE${NC}/$MAX_SCORE"
if [[ $WARNINGS -gt 0 ]]; then
    echo -e "Warnings: ${YELLOW}$WARNINGS${NC}"
fi
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Determine pass/fail
if [[ $SCORE -ge 70 ]]; then
    echo -e "${GREEN}✅ PASS${NC}"
    if [[ $SCORE -ge 85 ]]; then
        echo -e "${GREEN}Ready for publication${NC}"
        exit 0
    else
        echo -e "${YELLOW}Minor issues detected; review warnings${NC}"
        exit 0
    fi
else
    echo -e "${RED}❌ FAIL${NC}"
    echo -e "${RED}Needs significant improvements${NC}"
    exit 1
fi
