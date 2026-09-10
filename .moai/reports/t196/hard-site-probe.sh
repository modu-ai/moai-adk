#!/usr/bin/env bash
# Observe that a replaced HARD-site path actually resolves, and that the
# empty-expansion form a codex session would build does not.
set -u

echo "### pwd (the cwd premise the root-relative form rests on)"
pwd
echo
echo "### node --version"
node --version
echo
printf '<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10"><rect width="10" height="10"/></svg>\n' \
  > .moai/reports/t196/probe.svg

echo "### CONTROL — the path an empty \${CLAUDE_SKILL_DIR} expansion produces"
echo '### command: node /scripts/check-svg.mjs .moai/reports/t196/probe.svg'
node /scripts/check-svg.mjs .moai/reports/t196/probe.svg 2>&1 | head -4
echo "exit=${PIPESTATUS[0]}"
echo
echo "### REPLACED — verbatim from SKILL.md:235"
echo '### command: node .claude/skills/moai-domain-svg-infographic/scripts/check-svg.mjs .moai/reports/t196/probe.svg'
node .claude/skills/moai-domain-svg-infographic/scripts/check-svg.mjs .moai/reports/t196/probe.svg 2>&1 | head -12
echo "exit=${PIPESTATUS[0]}"
