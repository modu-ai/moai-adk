# Progress — SPEC-V3R6-SEQ-THINKING-RETIRE-001

## SHOULD-FIX 3 (plan-auditor iter 1)

- **D1 — AC reference fix**: `plan.md:L128` says "AC-STR-003 (skill body literal check)" but the body-literal check is owned by AC-STR-005 per acceptance.md §2. Plan.md text retained as authored; this progress.md records the canonical AC mapping: M4's body verification is AC-STR-005 (positive ultrathink + negative sequential-thinking). AC-STR-003 covers configuration files (settings.json + skill body for moai-foundation-thinking). Both pass at M4/M5 boundary.
- **D2 — Inventory count reconciliation**: spec.md §6 totals to 50, plan.md M2 task list yields 28 + 10 + 2 + 7 + 4 = 51 files. Actual M1 baseline confirms canonical count: **49 files** (14 agents × 2 = 28, 5 skills × 2 + 1 thinking skill × 2 = 12, 1 rule × 2 = 2, 4 config files (`.mcp.json` + `.mcp.json.tmpl` + `.claude/settings.json` + `settings.json.tmpl`), 1 CLAUDE.local.md, 2 new files (audit test + allow-list) = 49). CLAUDE.md template mirror has zero sequential-thinking references (no template mirror edit needed). `settings.local.json` is runtime-managed (skipped per CLAUDE.local.md §2 [HARD]).
- **D3 — AC-STR-002 verification strengthening**: acceptance.md:L94 lists `go build ./cmd/moai` as a template-build proxy. M7 verification batch adds `go test ./internal/template/...` (full package test) which exercises template rendering via `TestEmbeddedTemplates` family.

## M1 — Inventory and Baseline Capture

**Commit**: (pending — M1 is local artifact creation only)
**Baseline file**: `.moai/specs/SPEC-V3R6-SEQ-THINKING-RETIRE-001/baseline.txt` (109 lines)
**Test baseline**: `.moai/specs/SPEC-V3R6-SEQ-THINKING-RETIRE-001/baseline-test.txt` (captured pre-existing failures)
**Lint baseline**: `.moai/specs/SPEC-V3R6-SEQ-THINKING-RETIRE-001/baseline-lint.txt`

### Pre-existing failures (NOT attributable to this SPEC)

`TestRuleTemplateMirrorDrift` 3 sub-failures:
- `.claude/rules/moai/development/manager-develop-prompt-template.md` (source 8496 vs mirror 7180 bytes) — pre-existing drift
- `.claude/rules/moai/workflow/spec-workflow.md` (source 29363 vs mirror 26709 bytes) — pre-existing drift
- `.claude/agents/meta/plan-auditor.md` (source 21042 vs mirror 18778 bytes) — pre-existing drift

`TestSkillsContainPlanAuditGateMarkers` — `run.md` missing 5 Plan Audit Gate markers (pre-existing)

`golangci-lint` 2 unused entities:
- `internal/cli/init_layout.go:61 renderInitNextSteps` (pre-existing unused)
- `internal/cli/wizard/review.go:15 reviewChoiceProceed` (pre-existing unused)

### Catalog location correction (vs plan.md)

Plan.md M7 references `.claude/skills/catalog.yaml` but the actual file is at `internal/template/catalog.yaml` (293 lines). Hash regeneration tool is `internal/template/scripts/gen-catalog-hashes.go`. M7 will use the correct path.

## M1 Inventory Summary

| Category | Local | Template | Total |
|----------|-------|----------|-------|
| Agents (14) | 14 | 14 | 28 |
| Skills (5 + 1 thinking) | 6 | 6 | 12 |
| Rules (settings-management.md) | 1 | 1 | 2 |
| .mcp.json + template | 1 | 1 | 2 |
| settings.json + template | 1 | 1 | 2 |
| CLAUDE.local.md (no template mirror) | 1 | 0 | 1 |
| CLAUDE.md (no sequential-thinking refs found) | 0 | 0 | 0 (but retirement note will be added in M6) |
| New files (audit test + allow-list) | n/a | n/a | 2 |
| **Total touched files** | | | **49** |

`internal/template/templates/CLAUDE.md` contains zero `sequential-thinking` references already — no edit needed in template mirror for CLAUDE.md.

Local `CLAUDE.md` is not templated (per CLAUDE.local.md §2 — local file). M6 adds the retirement note here only.

## Milestones — Status (initial)

- [ ] M1 — Inventory + baseline ✓ (this section)
- [ ] M2 — Agent frontmatter cleanup (28 files)
- [ ] M3 — Skill frontmatter + body cleanup (10 files, 5 skills × 2)
- [ ] M4 — moai-foundation-thinking redesign + BC marker (2 files)
- [ ] M5 — Settings + .mcp.json + rules cleanup (~7 files)
- [ ] M6 — CI guard + docs (4 files)
- [ ] M7 — Validation + catalog regeneration

## Commit SHA list (pending)

| Milestone | SHA | Subject |
|-----------|-----|---------|
| M1 | (pending) | chore(SPEC-V3R6-SEQ-THINKING-RETIRE-001): M1 inventory + baseline |
| M2 | (pending) | refactor(agents): M2 — 14 agent frontmatter tools 정리 |
| M3 | (pending) | refactor(skills): M3 — 5 skill frontmatter+body 정리 |
| M4 | (pending) | refactor(moai-foundation-thinking): M4 — ultrathink-centric 재설계 + BC marker |
| M5 | (pending) | chore(config): M5 — .mcp.json + settings.json + settings-management.md 정리 |
| M6 | (pending) | feat(template): M6 — SEQ_THINKING_REINTRODUCED CI guard + allow-list + docs sync |
| M7 | (pending) | chore(SPEC-V3R6-SEQ-THINKING-RETIRE-001): M7 — mark implemented v0.2.0 + catalog hash refresh |
