#!/bin/bash

echo "🔄 Syncing Week 1 Complete Skills to Templates..."
echo "=========================================="

SKILLS_DIR=".claude/skills"
TEMPLATES_DIR="src/moai_adk/templates/.claude/skills"

# 10 완료된 스킬들
SKILLS=(
  "moai-lang-ruby"
  "moai-lang-php"
  "moai-lang-scala"
  "moai-lang-cpp"
  "moai-lang-kotlin"
  "moai-lang-html-css"
  "moai-lang-rust"
  "moai-domain-frontend"
  "moai-domain-figma"
  "moai-domain-monitoring"
)

for skill in "${SKILLS[@]}"; do
  echo "📦 Syncing: $skill"
  
  # Create template directory
  mkdir -p "$TEMPLATES_DIR/$skill/modules"
  
  # Copy all files
  cp "$SKILLS_DIR/$skill/SKILL.md" "$TEMPLATES_DIR/$skill/" 2>/dev/null
  cp "$SKILLS_DIR/$skill/examples.md" "$TEMPLATES_DIR/$skill/" 2>/dev/null
  cp "$SKILLS_DIR/$skill/reference.md" "$TEMPLATES_DIR/$skill/" 2>/dev/null
  
  # Copy modules
  if [ -d "$SKILLS_DIR/$skill/modules" ]; then
    cp "$SKILLS_DIR/$skill/modules"/*.md "$TEMPLATES_DIR/$skill/modules/" 2>/dev/null
  fi
  
  echo "  ✓ $skill synced"
done

echo ""
echo "✅ All Week 1 skills synced to templates!"
echo ""
echo "📊 Sync Summary:"
for skill in "${SKILLS[@]}"; do
  file_count=$(find "$TEMPLATES_DIR/$skill" -type f -name "*.md" | wc -l)
  echo "  🔗 $skill: $file_count files"
done
