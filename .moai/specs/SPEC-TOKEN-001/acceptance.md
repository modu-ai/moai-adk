---
id: SPEC-TOKEN-001
phase: acceptance
---

# Acceptance Criteria

## AC1: Skill Count Limit
- GIVEN any agent definition in .claude/agents/moai/
- WHEN the skills list in YAML frontmatter is counted
- THEN no agent has more than 4 skills

## AC2: No Language Skills in Frontmatter
- GIVEN any agent definition in .claude/agents/moai/
- WHEN searching for moai-lang-* in skills list
- THEN zero results are found

## AC3: Foundation-Claude Restricted to Builders
- GIVEN agent definitions other than builder-agent, builder-skill, builder-plugin
- WHEN searching for moai-foundation-claude in skills list
- THEN zero results are found

## AC4: JIT Language Detection Documented
- GIVEN the run.md workflow file
- WHEN searching for "JIT Language" or language detection section
- THEN a section exists documenting project language detection rules

## AC5: Build Succeeds
- GIVEN all template changes are complete
- WHEN running `make build`
- THEN the command succeeds with exit code 0

## AC6: Tests Pass
- GIVEN the rebuilt embedded templates
- WHEN running `go test ./...`
- THEN all tests pass

## AC7: All Agents Retain foundation-core
- GIVEN any agent with skills in frontmatter
- WHEN the skills list is checked
- THEN moai-foundation-core is present

## Verification Commands

```bash
# AC1: No agent has more than 4 skills
for f in internal/template/templates/.claude/agents/moai/*.md; do
  count=$(awk '/^skills:/{found=1; next} found && /^  - /{count++} found && !/^  - /{found=0} END{print count}' "$f")
  [ "$count" -gt 4 ] && echo "FAIL: $(basename $f) has $count skills"
done

# AC2: No language skills in frontmatter
grep -l "moai-lang-" internal/template/templates/.claude/agents/moai/*.md && echo "FAIL" || echo "PASS"

# AC3: foundation-claude only in builders
grep -l "moai-foundation-claude" internal/template/templates/.claude/agents/moai/*.md | grep -v builder && echo "FAIL" || echo "PASS"

# AC5-6: Build and test
make build && go test ./...
```
