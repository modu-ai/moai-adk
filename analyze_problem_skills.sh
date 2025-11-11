#!/bin/bash

SKILL_DIR="src/moai_adk/templates/.claude/skills"
ANALYSIS_FILE="detailed_skill_analysis.md"

echo "# Detailed Skill Analysis Report" > "$ANALYSIS_FILE"
echo "Generated: $(date)" >> "$ANALYSIS_FILE"
echo "" >> "$ANALYSIS_FILE"

# Function to analyze skill structure
analyze_skill() {
    local skill_dir="$1"
    local skill_name=$(basename "$skill_dir")
    
    echo "## $skill_name" >> "$ANALYSIS_FILE"
    echo "" >> "$ANALYSIS_FILE"
    
    # File structure
    echo "### File Structure" >> "$ANALYSIS_FILE"
    echo '```' >> "$ANALYSIS_FILE"
    ls -la "$skill_dir" >> "$ANALYSIS_FILE"
    echo '```' >> "$ANALYSIS_FILE"
    echo "" >> "$ANALYSIS_FILE"
    
    # Check SKILL.md if exists
    local skill_file="$skill_dir/SKILL.md"
    if [[ -f "$skill_file" ]]; then
        echo "### SKILL.md Preview (first 20 lines)" >> "$ANALYSIS_FILE"
        echo '```' >> "$ANALYSIS_FILE"
        head -20 "$skill_file" >> "$ANALYSIS_FILE"
        echo '```' >> "$ANALYSIS_FILE"
        echo "" >> "$ANALYSIS_FILE"
        
        # Check file size
        local file_size=$(wc -l < "$skill_file")
        echo "**File size:** $file_size lines" >> "$ANALYSIS_FILE"
        echo "" >> "$ANALYSIS_FILE"
    else
        echo "❌ **SKILL.md file missing**" >> "$ANALYSIS_FILE"
        echo "" >> "$ANALYSIS_FILE"
    fi
    
    echo "---" >> "$ANALYSIS_FILE"
    echo "" >> "$ANALYSIS_FILE"
}

# Analyze major problem categories
echo "## Critical Issues (Missing SKILL.md)" >> "$ANALYSIS_FILE"
echo "" >> "$ANALYSIS_FILE"
analyze_skill "$SKILL_DIR/moai-document-processing"

echo "## Metadata Issues (Sample)" >> "$ANALYSIS_FILE"
echo "" >> "$ANALYSIS_FILE"

# Sample skills with metadata issues
problem_skills=(
    "moai-alfred-agent-guide"
    "moai-alfred-code-reviewer"
    "moai-baas-foundation"
    "moai-foundation-trust"
)

for skill in "${problem_skills[@]}"; do
    analyze_skill "$SKILL_DIR/$skill"
done

echo "## Missing Supporting Files (Sample)" >> "$ANALYSIS_FILE"
echo "" >> "$ANALYSIS_FILE"

# Sample skills with missing files
missing_files_skills=(
    "moai-cc-agents"
    "moai-alfred-workflow"
    "moai-artifacts-builder"
)

for skill in "${missing_files_skills[@]}"; do
    analyze_skill "$SKILL_DIR/$skill"
done

echo "Analysis complete. Report saved to $ANALYSIS_FILE" > /dev/stderr
