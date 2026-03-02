# Acceptance Criteria: SPEC-CONTEXT-CLEANUP-001

## Verification Checklist

### DELETE Verification
- [ ] `context.md` does not exist in `.claude/skills/moai/workflows/`
- [ ] `context.md` does not exist in `internal/template/templates/.claude/skills/moai/workflows/`
- [ ] `context.yaml` does not exist in `.moai/config/sections/`
- [ ] `context.yaml` does not exist in `internal/template/templates/.moai/config/sections/`

### ROUTING Verification
- [ ] SKILL.md contains no `context`, `ctx`, or `memory` alias
- [ ] SKILL.md contains no context workflow reference

### CLAUDE.md Verification
- [ ] No "Context Search Protocol" section exists
- [ ] Section numbering is consistent (no gaps)

### COMMIT FORMAT Verification
- [ ] manager-git.md contains no `## Context` format instructions
- [ ] sync.md contains no Step 3.1.1 (Context Memory Generation)

### CROSS-REFERENCE Verification
- [ ] `grep -ri "context memory" --include="*.md" --include="*.yaml"` returns zero hits in skills/workflows/agents/rules (excluding SPEC, CHANGELOG)
- [ ] `grep -ri "moai-workflow-context" --include="*.md"` returns zero hits
- [ ] `grep -ri "/moai context\|/moai memory\|/moai ctx" --include="*.md"` returns zero hits (excluding SPEC, CHANGELOG)
- [ ] `grep -ri "Context Search Protocol" --include="*.md"` returns zero hits (excluding SPEC, CHANGELOG)

### ENHANCEMENT Verification
- [ ] plan.md workflow includes Implementation Log section guidance
- [ ] run.md workflow includes Implementation Log recording guidance
- [ ] moai-memory.md rule references SPEC documents as primary context

### PRESERVATION Verification
- [ ] moai-foundation-context SKILL.md still exists and functions
- [ ] 8 agents still reference moai-foundation-context in skills
- [ ] Session boundary git tags are still defined in sync.md
- [ ] Conventional commit format (type(scope): description) is preserved

### BUILD Verification
- [ ] `make build` succeeds
- [ ] `go test ./...` passes
- [ ] `go vet ./...` passes

### CONSISTENCY Verification
- [ ] Every modified template file has matching local file
- [ ] Every modified local file has matching template file (except README, CHANGELOG)
