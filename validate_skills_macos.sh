#!/bin/bash

SKILL_DIR="src/moai_adk/templates/.claude/skills"
REPORT_FILE="skill_validation_report_macos.md"

# Initialize counters
MISSING_SKILL_MD=0
INVALID_YAML=0
MISSING_METADATA=0
MISSING_EXAMPLES=0
MISSING_REFERENCE=0
VALID_SKILLS=0

# Initialize report
echo "# Skill Validation Report - $(date)" > "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
echo "## Validation Results" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

# Function to validate a skill
validate_skill() {
    local skill_dir="$1"
    local skill_name=$(basename "$skill_dir")
    local skill_file="$skill_dir/SKILL.md"
    
    # Check if SKILL.md exists
    if [[ ! -f "$skill_file" ]]; then
        echo "❌ $skill_name: SKILL.md file missing" >> "$REPORT_FILE"
        MISSING_SKILL_MD=$((MISSING_SKILL_MD + 1))
        return 1
    fi
    
    # Check YAML frontmatter
    local yaml_start=0
    local yaml_end=0
    local has_name=0
    local has_version=0
    local has_status=0
    local has_description=0
    
    # Read file line by line
    while IFS= read -r line; do
        if [[ "$line" == "---" ]]; then
            if [[ $yaml_start -eq 0 ]]; then
                yaml_start=1
            elif [[ $yaml_end -eq 0 ]]; then
                yaml_end=1
                break
            fi
        elif [[ $yaml_start -eq 1 && $yaml_end -eq 0 ]]; then
            [[ "$line" =~ ^name: ]] && has_name=1
            [[ "$line" =~ ^version: ]] && has_version=1
            [[ "$line" =~ ^status: ]] && has_status=1
            [[ "$line" =~ ^description: ]] && has_description=1
        fi
    done < "$skill_file"
    
    # Check required metadata
    local missing_fields=""
    if [[ $has_name -eq 0 ]]; then missing_fields="$missing_fields name"; fi
    if [[ $has_version -eq 0 ]]; then missing_fields="$missing_fields version"; fi
    if [[ $has_status -eq 0 ]]; then missing_fields="$missing_fields status"; fi
    if [[ $has_description -eq 0 ]]; then missing_fields="$missing_fields description"; fi
    
    if [[ -n "$missing_fields" ]]; then
        echo "⚠️  $skill_name: Missing metadata fields:$missing_fields" >> "$REPORT_FILE"
        MISSING_METADATA=$((MISSING_METADATA + 1))
        return 1
    fi
    
    # Check for additional files
    local missing_files=""
    [[ ! -f "$skill_dir/examples.md" ]] && missing_files="$missing_files examples.md"
    [[ ! -f "$skill_dir/reference.md" ]] && missing_files="$missing_files reference.md"
    
    if [[ -n "$missing_files" ]]; then
        echo "⚠️  $skill_name: Missing files:$missing_files" >> "$REPORT_FILE"
        [[ "$missing_files" =~ "examples.md" ]] && MISSING_EXAMPLES=$((MISSING_EXAMPLES + 1))
        [[ "$missing_files" =~ "reference.md" ]] && MISSING_REFERENCE=$((MISSING_REFERENCE + 1))
        return 1
    fi
    
    echo "✅ $skill_name: Valid" >> "$REPORT_FILE"
    VALID_SKILLS=$((VALID_SKILLS + 1))
    return 0
}

# Validate all skills
echo "Validating all skills..." > /dev/stderr

for skill_dir in "$SKILL_DIR"/*; do
    if [[ -d "$skill_dir" ]]; then
        validate_skill "$skill_dir"
    fi
done

# Calculate totals
TOTAL_SKILLS=$(find "$SKILL_DIR" -maxdepth 1 -type d | wc -l)

# Add summary to report
echo "" >> "$REPORT_FILE"
echo "## Statistics Summary" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
echo "- Total Skills: $TOTAL_SKILLS" >> "$REPORT_FILE"
echo "- Valid Skills: $VALID_SKILLS" >> "$REPORT_FILE"
echo "- Missing SKILL.md: $MISSING_SKILL_MD" >> "$REPORT_FILE"
echo "- Missing Metadata: $MISSING_METADATA" >> "$REPORT_FILE"
echo "- Missing examples.md: $MISSING_EXAMPLES" >> "$REPORT_FILE"
echo "- Missing reference.md: $MISSING_REFERENCE" >> "$REPORT_FILE"

# Calculate success rate
if [[ $TOTAL_SKILLS -gt 0 ]]; then
    SUCCESS_RATE=$(( (VALID_SKILLS * 100) / TOTAL_SKILLS ))
    echo "" >> "$REPORT_FILE"
    echo "**Validation Success Rate: $SUCCESS_RATE%**" >> "$REPORT_FILE"
fi

echo "Validation complete. Report saved to $REPORT_FILE" > /dev/stderr
