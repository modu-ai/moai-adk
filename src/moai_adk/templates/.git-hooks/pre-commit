#!/bin/bash
# @CODE:DOC-TAG-004 | Component 1: Pre-commit hook for TAG validation
#
# This hook validates TAG annotations in staged files before commit.
# It checks:
# - TAG format (@DOC:DOMAIN-TYPE-NNN)
# - Duplicate TAG detection
# - Orphan TAG detection (warnings only)
#
# Exit codes:
#   0 - Validation passed
#   1 - Validation failed (duplicates or format errors)

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get repository root
REPO_ROOT=$(git rev-parse --show-toplevel)

# Check if Python module is available
if ! python3 -c "import moai_adk.core.tags.pre_commit_validator" 2>/dev/null; then
    echo -e "${YELLOW}Warning: moai_adk TAG validator not found.${NC}"
    echo "Skipping TAG validation. Install moai_adk to enable validation."
    exit 0
fi

# Get staged files
STAGED_FILES=$(git diff --cached --name-only --diff-filter=ACM)

if [ -z "$STAGED_FILES" ]; then
    echo -e "${GREEN}No staged files to validate.${NC}"
    exit 0
fi

echo "🔍 Validating TAG annotations in staged files..."

# Run TAG validation
# Pass staged files as arguments to validator
python3 -m moai_adk.core.tags.pre_commit_validator \
    --files $STAGED_FILES

VALIDATION_RESULT=$?

# Check result
if [ $VALIDATION_RESULT -eq 0 ]; then
    echo -e "${GREEN}✓ TAG validation passed.${NC}"
    exit 0
else
    echo -e "${RED}✗ TAG validation failed.${NC}"
    echo ""
    echo "Commit blocked due to TAG validation errors."
    echo ""
    echo "To fix:"
    echo "  1. Fix duplicate TAGs or format errors shown above"
    echo "  2. Stage your changes with 'git add'"
    echo "  3. Try committing again"
    echo ""
    echo "To skip this validation (not recommended):"
    echo "  git commit --no-verify"
    exit 1
fi
