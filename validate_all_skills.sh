#!/bin/bash

SKILL_DIR="src/moai_adk/templates/.claude/skills"
REPORT_FILE="skill_validation_report.md"
TEMP_FILE="temp_validation.txt"

# Initialize report
echo "# Skill Validation Report - $(date)" > "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
echo "## Summary Statistics" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

# Count total skills
TOTAL_SKILLS=$(find "$SKILL_DIR" -maxdepth 1 -type d | wc -l)
echo "- Total Skills: $TOTAL_SKILLS" >> "$REPORT_FILE"

# Validation categories
declare -A stats=(
    ["missing_skill_md"]=0
    ["invalid_yaml"]=0
    ["missing_metadata"]=0
    ["missing_examples"]=0
    ["missing_reference"]=0
    ["valid_skills"]=0
)

echo "" >> "$REPORT_FILE"
echo "## Validation Results" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

# Function to validate YAML frontmatter
validate_yaml() {
    local file="$1"
    local skill_name="$2"
    
    # Check if file exists
    if [[ ! -f "$file" ]]; then
        echo "❌ $skill_name: SKILL.md file missing"
        ((stats["missing_skill_md"]++))
        return 1
    fi
    
    # Extract YAML frontmatter
    local yaml_content=$(sed -n '/^---$/,/^---$/p' "$file" | sed '1d;$d')
    
    # Check required fields
    local required_fields=("name" "version" "status" "description")
    local missing_fields=()
    
    for field in "${required_fields[@]}"; do
        if ! echo "$yaml_content" | grep -q "^$field:"; then
            missing_fields+=("$field")
        fi
    done
    
    if [[ ${#missing_fields[@]} -gt 0 ]]; then
        echo "⚠️  $skill_name: Missing metadata fields: ${missing_fields[*]}"
        ((stats["missing_metadata"]++))
        return 1
    fi
    
    # Check for additional files
    local skill_dir=$(dirname "$file")
    local missing_files=()
    
    [[ ! -f "$skill_dir/examples.md" ]] && missing_files+=("examples.md")
    [[ ! -f "$skill_dir/reference.md" ]] && missing_files+=("reference.md")
    
    if [[ ${#missing_files[@]} -gt 0 ]]; then
        echo "⚠️  $skill_name: Missing files: ${missing_files[*]}"
        [[ " ${missing_files[*]} " =~ " examples.md " ]] && ((stats["missing_examples"]++))
        [[ " ${missing_files[*]} " =~ " reference.md " ]] && ((stats["missing_reference"]++))
        return 1
    fi
    
    echo "✅ $skill_name: Valid"
    ((stats["valid_skills"]++))
    return 0
}

# Validate all skills
for skill_dir in "$SKILL_DIR"/*; do
    if [[ -d "$skill_dir" ]]; then
        skill_name=$(basename "$skill_dir")
        skill_file="$skill_dir/SKILL.md"
        validate_yaml "$skill_file" "$skill_name" >> "$REPORT_FILE"
    fi
done

# Add statistics summary
echo "" >> "$REPORT_FILE"
echo "## Statistics Summary" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"
for category in "${!stats[@]}"; do
    echo "- $category: ${stats[$category]}" >> "$REPORT_FILE"
done

# Calculate percentage
VALID_PERCENT=$(( (stats["valid_skills"] * 100) / TOTAL_SKILLS ))
echo "" >> "$REPORT_FILE"
echo "**Validation Success Rate: $VALID_PERCENT%**" >> "$REPORT_FILE"

echo "Validation complete. Report saved to $REPORT_FILE"
