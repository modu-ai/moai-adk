#!/usr/bin/env bash
# REQ-CSN-009 sweep: enumerate the surfaces that decide the working directory of
# a process reading skill content, and record which support the project-root
# premise and which do not.
set -u

echo "### S1. every explicit working-directory assignment in the launcher family"
grep -rn '\.Dir = \|Dir:' internal/cli/codex_launcher.go internal/cli/claude_launcher.go 2>/dev/null

echo
echo "### S2. the codex launcher's root resolution and its degradation branch"
sed -n '238,255p' internal/cli/codex_launcher.go

echo
echo "### S3. every .Dir assignment across the CLI package (is any other spawn path relevant?)"
grep -rln '\.Dir = ' internal/cli/ | sort

echo
echo "### S4. does anything in the repo force cwd for a directly-launched codex session?"
grep -rn 'CODEX_\|codex exec\|codex resume' internal/cli/*.go | grep -i 'dir\|chdir\|cwd' || echo "(no cwd-forcing match)"

echo
echo "### S5. skills tree presence relative to THIS process's cwd (the premise, exercised)"
pwd
test -d .claude/skills && echo ".claude/skills resolves from cwd: yes" || echo ".claude/skills resolves from cwd: NO"

echo
echo "### S6. spec.md §A.7 carries the three AC-CSN-013 items"
grep -n 'codex_launcher.go:245-250\|강등\|묶는 장치가 없다' .moai/specs/SPEC-CODEX-SKILL-NEUTRAL-001/spec.md
