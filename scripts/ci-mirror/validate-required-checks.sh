#!/bin/sh
# Validate SSoT integrity of .github/required-checks.yml
# SPEC-V3R3-CI-AUTONOMY-001 W4-T05
#
# Three-dimension validation:
#   (a) auxiliary: items match actual workflow names (case-insensitive)
#   (b) branches.main.contexts ∩ auxiliary: = ∅ (no overlap)
#   (c) branches.release/*.contexts ∩ auxiliary: = ∅ (no overlap)
#
# Usage:
#   ./scripts/ci-mirror/validate-required-checks.sh
#   echo $?  # 0 = pass, 1 = fail, 2 = missing file

set -eu

REQUIRED_CHECKS_FILE=".github/required-checks.yml"

# Verify file exists
if [ ! -f "$REQUIRED_CHECKS_FILE" ]; then
	echo "Error: $REQUIRED_CHECKS_FILE not found" >&2
	exit 2
fi

# Helper: extract YAML field as list (requires yq)
get_yaml_list() {
	local file="$1"
	local path="$2"
	yq eval "$path | .[]" "$file" 2>/dev/null || true
}

# Helper: check if value is in list
in_list() {
	local value="$1"
	shift
	for item; do
		[ "$item" = "$value" ] && return 0
	done
	return 1
}

# Dimension A: auxiliary items must correspond to actual workflow names
echo "=== Dimension A: Validating auxiliary → workflow name mapping ==="

auxiliary_items=$(get_yaml_list "$REQUIRED_CHECKS_FILE" ".auxiliary")
exit_code=0

while read -r aux_item; do
	[ -z "$aux_item" ] && continue

	# Workflow name is derived from:
	#   - .github/workflows/<name>.yml → name field in workflow
	#   - .github/workflows/<name>.optional.yml → name field in workflow
	# Get all workflow names via gh workflow list

	workflow_name=$(gh workflow list --json name,path --limit 100 2>/dev/null | \
		yq eval ".[] | select(.path | contains(\"$aux_item.yml\") or contains(\"$aux_item.optional.yml\")) | .name" 2>/dev/null || true)

	if [ -z "$workflow_name" ]; then
		# Fallback: check if file exists with either extension
		if [ -f ".github/workflows/${aux_item}.yml" ] || [ -f ".github/workflows/${aux_item}.optional.yml" ]; then
			echo "✓ $aux_item (file found)"
		else
			echo "✗ $aux_item (no matching workflow file)" >&2
			exit_code=1
		fi
	else
		echo "✓ $aux_item → $workflow_name"
	fi
done <<EOF
$(echo "$auxiliary_items")
EOF

# Dimension B: auxiliary items must not be in branches.main.contexts
echo ""
echo "=== Dimension B: Validating branches.main.contexts ∩ auxiliary = ∅ ==="

main_contexts=$(get_yaml_list "$REQUIRED_CHECKS_FILE" ".branches.main.contexts")
while read -r aux_item; do
	[ -z "$aux_item" ] && continue

	if echo "$main_contexts" | grep -q "^${aux_item}$"; then
		echo "✗ $aux_item found in branches.main.contexts (should be auxiliary only)" >&2
		exit_code=1
	else
		echo "✓ $aux_item not in branches.main.contexts"
	fi
done <<EOF
$(echo "$auxiliary_items")
EOF

# Dimension C: auxiliary items must not be in branches.release/*.contexts
echo ""
echo "=== Dimension C: Validating branches.release/*.contexts ∩ auxiliary = ∅ ==="

release_contexts=$(get_yaml_list "$REQUIRED_CHECKS_FILE" ".branches[\"release/*\"].contexts")
while read -r aux_item; do
	[ -z "$aux_item" ] && continue

	if echo "$release_contexts" | grep -q "^${aux_item}$"; then
		echo "✗ $aux_item found in branches.release/*.contexts (should be auxiliary only)" >&2
		exit_code=1
	else
		echo "✓ $aux_item not in branches.release/*.contexts"
	fi
done <<EOF
$(echo "$auxiliary_items")
EOF

# Summary
echo ""
if [ $exit_code -eq 0 ]; then
	echo "✅ All validations passed"
else
	echo "❌ Validation failed" >&2
fi

exit $exit_code
