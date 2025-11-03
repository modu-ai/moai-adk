#!/usr/bin/env bash
# validate-spec.sh - SPEC validation helper for MoAI-ADK
# Usage: ./validate-spec.sh .moai/specs/SPEC-AUTH-001

set -e

SPEC_DIR="$1"

if [ -z "$SPEC_DIR" ]; then
  echo "Usage: $0 <SPEC_DIR>"
  echo "Example: $0 .moai/specs/SPEC-AUTH-001"
  exit 1
fi

if [ ! -d "$SPEC_DIR" ]; then
  echo "Error: Directory $SPEC_DIR does not exist"
  exit 1
fi

if [ ! -f "$SPEC_DIR/spec.md" ]; then
  echo "Error: spec.md not found in $SPEC_DIR"
  exit 1
fi

echo "=========================================="
echo "Validating SPEC: $SPEC_DIR"
echo "=========================================="
echo ""

# Check required fields
echo -n "Required fields (7)... "
REQUIRED_COUNT=$(rg "^(id|version|status|created|updated|author|priority):" "$SPEC_DIR/spec.md" | wc -l | tr -d ' ')
if [ "$REQUIRED_COUNT" -eq 7 ]; then
  echo "✅ ($REQUIRED_COUNT/7)"
else
  echo "❌ ($REQUIRED_COUNT/7)"
  echo "  Missing fields. Expected: id, version, status, created, updated, author, priority"
fi

# Check author format
echo -n "Author format (@Handle)... "
if rg "^author: @[A-Z]" "$SPEC_DIR/spec.md" > /dev/null 2>&1; then
  echo "✅"
else
  echo "❌"
  echo "  Expected format: author: @YourHandle"
fi

# Check version format
echo -n "Version format (0.x.y)... "
if rg "^version: 0\.\d+\.\d+" "$SPEC_DIR/spec.md" > /dev/null 2>&1; then
  echo "✅"
else
  echo "❌"
  echo "  Expected format: version: 0.x.y (e.g., 0.0.1, 0.1.0)"
fi

# Check HISTORY section
echo -n "HISTORY section... "
if rg "^## HISTORY" "$SPEC_DIR/spec.md" > /dev/null 2>&1; then
  echo "✅"
else
  echo "❌"
  echo "  Missing HISTORY section"
fi

# Check TAG block
echo -n "TAG block (@SPEC:ID)... "
if rg "^# @SPEC:" "$SPEC_DIR/spec.md" > /dev/null 2>&1; then
  echo "✅"
else
  echo "❌"
  echo "  Expected format: # @SPEC:YOUR-ID: Title"
fi

# Check Environment section
echo -n "Environment section... "
if rg "^## Environment" "$SPEC_DIR/spec.md" > /dev/null 2>&1; then
  echo "✅"
else
  echo "⚠️  (optional but recommended)"
fi

# Check Assumptions section
echo -n "Assumptions section... "
if rg "^## Assumptions" "$SPEC_DIR/spec.md" > /dev/null 2>&1; then
  echo "✅"
else
  echo "⚠️  (optional but recommended)"
fi

# Check Requirements section
echo -n "Requirements section... "
if rg "^## Requirements" "$SPEC_DIR/spec.md" > /dev/null 2>&1; then
  echo "✅"
else
  echo "❌"
  echo "  Missing Requirements section"
fi

# Check duplicate IDs
SPEC_ID=$(basename "$SPEC_DIR" | sed 's/SPEC-//')
echo -n "Duplicate ID check ($SPEC_ID)... "
DUPLICATE_COUNT=$(rg "@SPEC:$SPEC_ID" -n .moai/specs/ 2>/dev/null | wc -l | tr -d ' ')
if [ "$DUPLICATE_COUNT" -eq 1 ]; then
  echo "✅ (1 occurrence)"
else
  echo "❌ (found $DUPLICATE_COUNT occurrences)"
  if [ "$DUPLICATE_COUNT" -gt 1 ]; then
    echo "  Duplicate SPEC IDs detected:"
    rg "@SPEC:$SPEC_ID" -n .moai/specs/
  fi
fi

# Check EARS patterns
echo ""
echo "EARS Pattern Usage:"
echo -n "  Ubiquitous (UR-XXX)... "
if rg "^\*\*UR-\d+" "$SPEC_DIR/spec.md" > /dev/null 2>&1; then
  COUNT=$(rg "^\*\*UR-\d+" "$SPEC_DIR/spec.md" | wc -l | tr -d ' ')
  echo "✅ ($COUNT found)"
else
  echo "⚠️  (none found)"
fi

echo -n "  Event-driven (ER-XXX)... "
if rg "^\*\*ER-\d+" "$SPEC_DIR/spec.md" > /dev/null 2>&1; then
  COUNT=$(rg "^\*\*ER-\d+" "$SPEC_DIR/spec.md" | wc -l | tr -d ' ')
  echo "✅ ($COUNT found)"
else
  echo "⚠️  (none found)"
fi

echo -n "  State-driven (SR-XXX)... "
if rg "^\*\*SR-\d+" "$SPEC_DIR/spec.md" > /dev/null 2>&1; then
  COUNT=$(rg "^\*\*SR-\d+" "$SPEC_DIR/spec.md" | wc -l | tr -d ' ')
  echo "✅ ($COUNT found)"
else
  echo "⚠️  (none found)"
fi

echo -n "  Optional Features (OF-XXX)... "
if rg "^\*\*OF-\d+" "$SPEC_DIR/spec.md" > /dev/null 2>&1; then
  COUNT=$(rg "^\*\*OF-\d+" "$SPEC_DIR/spec.md" | wc -l | tr -d ' ')
  echo "✅ ($COUNT found)"
else
  echo "⚠️  (none found)"
fi

echo -n "  Unwanted Behaviors (UB-XXX)... "
if rg "^\*\*UB-\d+" "$SPEC_DIR/spec.md" > /dev/null 2>&1; then
  COUNT=$(rg "^\*\*UB-\d+" "$SPEC_DIR/spec.md" | wc -l | tr -d ' ')
  echo "✅ ($COUNT found)"
else
  echo "⚠️  (none found)"
fi

echo ""
echo "=========================================="
echo "Validation complete!"
echo "=========================================="
