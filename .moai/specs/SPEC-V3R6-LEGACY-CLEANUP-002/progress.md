# SPEC-V3R6-LEGACY-CLEANUP-002 — Progress Tracking

## Status Matrix

| Phase | Status | Commit | Date |
|-------|--------|--------|------|
| plan | DONE | `0f813831b` (spec + plan plan-phase) | 2026-05-23 |
| run M1 | DONE | (M1 commit pending) | 2026-05-23 |
| sync | TBD | — | — |
| mx | TBD (skip candidate — no Go code in scope) | — | — |

## Run-phase Evidence (M1 — Template Mirror Cascade)

### Pre-flight (HEAD = `0f813831b`)

- 5 target files exist (sizes captured: 19050 / 9725 / 8242 / 10903 / 15078 bytes)
- Plan-phase diff scope clean (only `.moai/specs/SPEC-V3R6-LEGACY-CLEANUP-002/{spec,plan}.md`)
- Lint baseline: `0 issues`
- Test baseline: `RULE_TEMPLATE_MIRROR_DRIFT` reports drift on `spec-assembly.md` (OUT OF SCOPE; pre-existing, will persist post-M1)
- L44 origin sync: `0 0` (ahead/behind both zero)
- Agency-keyword PRE count in `internal/template/templates/.claude/`: 5 files

### M1 Operations

5 `cp -p` operations (verbatim file copy):

1. `.claude/rules/moai/design/constitution.md` → `internal/template/templates/.claude/rules/moai/design/constitution.md`
2. `.claude/skills/moai-domain-brand-design/SKILL.md` → `internal/template/templates/.claude/skills/moai-domain-brand-design/SKILL.md`
3. `.claude/skills/moai-domain-copywriting/SKILL.md` → `internal/template/templates/.claude/skills/moai-domain-copywriting/SKILL.md`
4. `.claude/skills/moai-workflow-gan-loop/SKILL.md` → `internal/template/templates/.claude/skills/moai-workflow-gan-loop/SKILL.md`
5. `.claude/skills/moai/workflows/design.md` → `internal/template/templates/.claude/skills/moai/workflows/design.md`

### 5 AC Binary Verification Matrix

| AC | Status | Verification | Source SHA-256 | Target SHA-256 (post-M1) | Match? |
|----|--------|--------------|----------------|--------------------------|--------|
| AC-LCL2-001 | PASS | `shasum -a 256` constitution.md | `aa45b0255eed2f5278087078272b57fab991d000db7859631372866ca45d0670` | `aa45b0255eed2f5278087078272b57fab991d000db7859631372866ca45d0670` | YES |
| AC-LCL2-002 | PASS | `shasum -a 256` brand-design SKILL.md | `5185d309df1fd0ef9265b87f9952207ed297428601446c1397360c1386577421` | `5185d309df1fd0ef9265b87f9952207ed297428601446c1397360c1386577421` | YES |
| AC-LCL2-003 | PASS | `shasum -a 256` copywriting SKILL.md | `e00607b138afe71fab38dfb43026ea1d3794c51d946cc266b0711eb2d28ec1af` | `e00607b138afe71fab38dfb43026ea1d3794c51d946cc266b0711eb2d28ec1af` | YES |
| AC-LCL2-004 | PASS | `shasum -a 256` gan-loop SKILL.md | `e57bff5ccc8cf403fb96d9e0967e6f93c7d9f56146765a245909f3d1d3a2fde5` | `e57bff5ccc8cf403fb96d9e0967e6f93c7d9f56146765a245909f3d1d3a2fde5` | YES |
| AC-LCL2-005 | PASS | `shasum -a 256` design.md workflow | `fadac136e27bebf0f591044a7dcb88e288a01ea470b003a2c03027b18779c32c` | `fadac136e27bebf0f591044a7dcb88e288a01ea470b003a2c03027b18779c32c` | YES |

### PRESERVE Verification

- Source `.claude/` files NOT touched: `git diff --name-only -- '.claude/'` → empty (0 source modifications)
- Ambient files NOT staged: `usage-log.jsonl M`, `observations.yaml ??`, `v3.0-redesign-2026-05-23.md ??` all excluded
- No Go code modified: `.go` files in diff = 0
- Sibling SPEC dirs untouched

### Quality Gate

| Check | Result |
|-------|--------|
| go vet | not applicable (no Go file change scope) |
| go test ./internal/template/... | pre-existing drift on `spec-assembly.md` persists (OUT OF SCOPE — caused by a separate drift unrelated to LCL-002 5 files); NO NEW failures introduced by M1 |
| golangci-lint | baseline `0 issues` maintained (NEW 0) |
| Cross-platform build | not invoked (no Go file change scope) |

### Known Pre-existing State (NOT caused by this SPEC)

`internal/template/rule_template_mirror_test.go` reports `RULE_TEMPLATE_MIRROR_DRIFT` for `spec-assembly.md` (28423 bytes source vs 25939 bytes mirror). This file is OUT OF SCOPE for SPEC-V3R6-LEGACY-CLEANUP-002 (LCL-002 scope = exactly 5 files cleaned by LCL-001). The drift originates from a different upstream change and would require a separate SPEC to resolve.

## Sync-phase Evidence

TBD — pending `/moai sync SPEC-V3R6-LEGACY-CLEANUP-002`.

## MX-phase Evidence

Skip candidate — SPEC scope is 5 `.md` files only, no Go code in scope. `/moai mx` would have nothing to annotate.
