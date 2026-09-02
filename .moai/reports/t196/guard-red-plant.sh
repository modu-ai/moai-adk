#!/usr/bin/env bash
# AC-CSN-009: observe the guard going RED against a deliberately-degraded subject,
# then returning GREEN once the plant is reverted. Census recorded on both sides.
set -u
TARGET=internal/template/templates/.claude/skills/moai/SKILL.md
CENSUS='grep -rc CLAUDE_SKILL_DIR'

census() {
  grep -rn 'CLAUDE_SKILL_DIR' internal/template/templates/.claude/skills | wc -l | tr -d ' '
}

echo "### 0. census BEFORE planting"
census

SHA_BEFORE=$(shasum -a 256 "$TARGET" | cut -d' ' -f1)
cp "$TARGET" /tmp/t196-guard-plant.bak
printf '\nFor detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/plan.md\n' >> "$TARGET"

echo
echo "### 1. census AFTER planting one line into $TARGET"
census
echo
echo "### 2. guard run against the degraded subject"
go test ./internal/template/ -run TestSkillTreeHasNoClaudeSkillDirToken -v 2>&1 | tail -8
echo "guard-exit=${PIPESTATUS[0]}"

cp /tmp/t196-guard-plant.bak "$TARGET"
rm -f /tmp/t196-guard-plant.bak

echo
echo "### 3. census AFTER reverting the plant"
census
echo
echo "### 4. guard run against the restored subject"
go test ./internal/template/ -run TestSkillTreeHasNoClaudeSkillDirToken -v 2>&1 | tail -6
echo "guard-exit=${PIPESTATUS[0]}"
echo
echo "### 5. target restored byte-for-byte (sha256 before planting vs after revert)"
echo "before=$SHA_BEFORE"
echo "after =$(shasum -a 256 "$TARGET" | cut -d' ' -f1)"
