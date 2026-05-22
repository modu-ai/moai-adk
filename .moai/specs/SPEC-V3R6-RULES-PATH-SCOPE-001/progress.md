---
spec_id: SPEC-V3R6-RULES-PATH-SCOPE-001
artifact: progress
created: 2026-05-22
updated: 2026-05-22
---

# Progress — SPEC-V3R6-RULES-PATH-SCOPE-001

run-phase COMPLETE 2026-05-22. manager-develop cycle_type=ddd orchestrator-direct (per V3R6 commit 25ee73039 agent worktree hook regression precedent).

## Pre-flight Baseline (M0)

| Item | Value | Verification |
|---|---|---|
| Branch | `main` | `git branch --show-current` |
| HEAD baseline | `c620d0a84c230ee37bf63790f81087e4482dbfa8` | `git rev-parse HEAD` |
| Build linux | exit 0 | `go build ./...` |
| Build windows | exit 0 | `GOOS=windows GOARCH=amd64 go build ./...` |
| 4 target rules frontmatter | absent (start with `# <Title>`) | `head -1 × 4` |
| 5 keep-always rules frontmatter | absent (start with `# <Title>`) | `head -1 × 5` |
| 4 template mirror existence | confirmed (zone 35969B / design 18757B / manager-develop 6934B / agent-teams 6119B) | `ls -la` |
| internal/rules/ | ABSENT | `ls internal/rules/ 2>/dev/null` |
| paths frontmatter precedent | 14+ files (agent-authoring, agent-patterns, spec-frontmatter-schema, etc.) | `grep -l '^paths:'` |
| Cross-SPEC retirement conflicts | 0 (only documentation references, not retire markers) | `grep -r 'Retired\|deprecation-marker'` |
| 4 target rules word count (pre-SPEC) | zone-registry 3453 / design 2472 / manager-develop 1136 / agent-teams 770 = 7831 | `wc -w` |
| 5 keep-always word count | 5679 | `wc -w` |
| Total 9-rule baseline | 13510 words | sum |

## Pre-existing Template Mirror Drift Baseline (§1.4.1)

| Rule | Local | Template | Drift type | Status |
|---|---|---|---|---|
| core/zone-registry.md | 35,969B | 35,969B | content differ (4 diff lines, byte-equal) | drift modulo — Group B |
| design/constitution.md | 18,757B | 18,757B | byte-identical | Group A strict |
| development/manager-develop-prompt-template.md | 8,250B | 6,934B | template lacks ~1,316B (Tier S Applicability section) | drift modulo — Group B |
| workflow/agent-teams-pattern.md | 6,119B | 6,119B | byte-identical | Group A strict |

## M1 — 4 Rule Local Frontmatter Prepend

Edit tool 4회. 각 파일에 5-line frontmatter (line 1 `---` / line 2 `description:` / line 3 `paths:` / line 4 `---` / line 5 blank / line 6+ original body) prepend.

| Rule | description | paths CSV |
|---|---|---|
| zone-registry | "Zone Registry — MoAI-ADK HARD 조항 SSOT. rules/agents/skills/specs 디렉토리 수정 시에만 로드." | `.claude/**,.moai/specs/**,.claude/rules/**` |
| design/constitution | "Design System Constitution — MoAI design pipeline FROZEN/EVOLVABLE zone, GAN Loop contract, brand integration. design/brand 작업 시에만 로드." | `.moai/design/**,.moai/specs/SPEC-*-DESIGN-*/**,.moai/project/brand/**,.claude/skills/moai/**/design*.md,.claude/skills/moai/**/brand*.md` |
| manager-develop-prompt-template | "manager-develop 위임 Prompt Template — Tier M/L SPEC run-phase 5-section 표준. SPEC 위임 작성 시에만 로드." | `.moai/specs/**,.claude/agents/moai/manager-develop.md,.claude/skills/moai/workflows/run.md` |
| agent-teams-pattern | "Agent Teams Pattern — 5+1+1 team composition (5 implementer + 1 tester + 1 reviewer). team mode 설정/활성 시에만 로드." | `.moai/config/sections/workflow.yaml,.claude/agents/moai/manager-strategy.md,.claude/skills/moai/team/**` |

## M2 — 4 Template Mirror Sync + make build

Edit tool 4회 (template mirrors byte-identical frontmatter prepend) + `make build` (catalog.yaml 갱신 + binary rebuild).

| Mirror | Edit Status | Body Source Preserved |
|---|---|---|
| internal/template/templates/.claude/rules/moai/core/zone-registry.md | prepend OK | yes |
| internal/template/templates/.claude/rules/moai/design/constitution.md | prepend OK | yes |
| internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md | prepend OK | yes (existing 6934B template body unchanged, baseline drift modulo preserved) |
| internal/template/templates/.claude/rules/moai/workflow/agent-teams-pattern.md | prepend OK | yes |
| make build | catalog.yaml updated (16963 bytes) + go build success | n/a |

## M3 — Doctor Simulation Report

`.moai/reports/rules-path-scope-simulation-2026-05-22.md` 작성 (markdown, 5 session scenarios × 4 path-scoped rules glob match analysis).

| Scenario | zone-registry | design | manager-develop | agent-teams | 5 keep-always | Expected (AC) |
|---|---|---|---|---|---|---|
| Go-only (`internal/cli/foo.go`) | ✗ | ✗ | ✗ | ✗ | ✓ all 5 | AC-RPS-010 PASS |
| SPEC-only (`.moai/specs/SPEC-XXX/spec.md`) | ✓ | ✗ | ✓ | ✗ | ✓ all 5 | AC-RPS-011 PASS |
| Design (`.moai/design/tokens.json`) | ✗ | ✓ | ✗ | ✗ | ✓ all 5 | AC-RPS-012 PASS |
| Team (`.moai/config/sections/workflow.yaml`) | ✗ | ✗ | ✗ | ✓ | ✓ all 5 | AC-RPS-013 PASS |
| General docs (`README.md`) | ✗ | ✗ | ✗ | ✗ | ✓ all 5 | EC-RPS-001 verified |

**Findings**: Trigger miss = 0, Spurious load = 0 → AC-RPS-014 PASS.

**Token saving estimate**: ~23,253 tokens (~57% reduction in pre-SPEC always-loaded rule footprint).

## M4 — Parallel Regression Verification Batch

Single-turn multi-Bash batch per agent-common-protocol §Parallel Execution.

| Check | Command | Result |
|---|---|---|
| Linux build | `go build ./...` | exit 0 (empty output) |
| Windows build | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 (empty output) |
| Template tests | `go test ./internal/template/...` | FAIL (5 pre-existing baseline failures: TestAllAgentsInCatalog, TestLateBranchTemplateMirror/spec-assembly, TestSkillsContainPlanAuditGateMarkers/solo_run, TestRuleTemplateMirrorDrift/{manager-develop-prompt-template,plan-auditor,spec-workflow}, TestRetirementCompletenessAssertion) — NEW failures: 0. Baseline matches `project_v3r6_template_mirror_drift_audit_2026_05_22` memory. |
| C-HRA-008 boundary | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/ internal/hook/ \| grep -v "_test.go" \| grep -v "^[^:]*:[0-9]*:[ \t]*//" \| wc -l` | 0 |
| Lint baseline | `golangci-lint run --timeout=2m` | 23 pre-existing baseline issues (errcheck 6 / ineffassign 1 / staticcheck 5 / unused 11) — NEW: 0 (본 SPEC 는 Go code 미수정) |
| Mirror sync (Group A) | `diff` for design+agent-teams | byte-identical (empty) |
| Mirror sync (Group B) | `diff <(head -5)` for zone+manager-develop | frontmatter-5-line PASS |
| Body byte preservation | `diff git HEAD baseline body vs tail -n +6 current` × 4 | PASS empty × 4 |
| Word count saving | 7,831 words ≥ 7,500 threshold | PASS |
| YAML parse | python yaml.safe_load × 4 | PASS (`['description', 'paths']` × 4) |
| 5 keep-always intact | head -1 × 5 | no `---` start (PASS × 5) |
| No Go code change | git status filtered to SPEC files | only `.md` + (embedded.go via make build, no manual edit) |
| internal/rules/ + internal/loader/ absent | `ls 2>/dev/null` | both absent |

## M5 — Implementation 완료 보고

| Action | Status |
|---|---|
| spec.md frontmatter status: draft → implemented | DONE |
| spec.md frontmatter version: 0.1.0 → 0.2.0 | DONE |
| spec.md HISTORY entry v0.2.0 추가 | DONE |
| progress.md 작성 | DONE (this file) |

## 18 ACs Self-Verification Matrix

| AC | Status | Verification |
|---|---|---|
| AC-RPS-001 zone-registry frontmatter | PASS | `head -4` shows `---` / description / paths / `---` |
| AC-RPS-002 design/constitution frontmatter | PASS | `head -4` matches CSV spec |
| AC-RPS-003 manager-develop-prompt frontmatter | PASS | `head -4` matches CSV spec |
| AC-RPS-004 agent-teams-pattern frontmatter | PASS | `head -4` matches CSV spec |
| AC-RPS-005 5 keep-always rules unchanged | PASS | all 5 start with `#` not `---` |
| AC-RPS-006 4 rule body byte preservation | PASS | `diff git HEAD body vs tail -n +6` empty × 4 |
| AC-RPS-007 4 template mirror sync (Group A+B) | PASS | Group A byte-identical; Group B frontmatter-5-line + drift not expanded |
| AC-RPS-008 word count saving ≥ 7500w | PASS | 7,831w saved (5 keep-always = 5,679w, was 13,510w with 9 always-loaded) |
| AC-RPS-009 YAML parse validity | PASS | yaml.safe_load × 4 returns `['description', 'paths']` |
| AC-RPS-010 Go-only scenario | PASS | report row `✗ \| ✗ \| ✗ \| ✗ \| ✓ all 5` |
| AC-RPS-011 SPEC-only scenario | PASS | report row `✓ \| ✗ \| ✓ \| ✗ \| ✓ all 5` |
| AC-RPS-012 Design scenario | PASS | design column = ✓ |
| AC-RPS-013 Team scenario | PASS | agent-teams-pattern column = ✓ |
| AC-RPS-014 Trigger miss/spurious load 0 | PASS | report § Findings both 0 |
| AC-RPS-015 No Go code change | PASS | only 4 local `.md` + 4 template `.md` modified; internal/rules + internal/loader absent |
| AC-RPS-016 Cross-platform build | PASS | linux + windows both exit 0 |
| AC-RPS-017 CI baseline preserved | PASS | template test failures all pre-existing (5 baseline list confirmed); lint 23 baseline issues unchanged |
| AC-RPS-018 C-HRA-008 boundary | PASS | grep returns 0 matches |

**Total: 18/18 ACs PASS (100%).**

## Definition of Done

- [x] M1 완료: 4 rule local frontmatter prepend (AC-RPS-001 ~ AC-RPS-004 PASS)
- [x] M2 완료: 4 template mirror byte-identical sync + `make build` (AC-RPS-007 PASS)
- [x] M3 완료: 5 session 시나리오 시뮬레이션 보고서 + 위반 0 (AC-RPS-010 ~ AC-RPS-014 PASS)
- [x] M4 완료: 회귀 테스트 (AC-RPS-006, AC-RPS-008, AC-RPS-009, AC-RPS-015 ~ AC-RPS-018 PASS)
- [x] M5 완료: spec.md `status: implemented`, version `0.2.0` + progress.md 작성
- [ ] sync-phase: Hybrid Trunk Tier M 의무 = feat branch + PR (별도 sync-phase 위임, manager-git)
- [ ] PR 머지 시 origin/main 정렬

## Files Modified (Final)

8 files modified by this SPEC (run-phase scope only, B8 working tree unrelated files preserved):

```
.claude/rules/moai/core/zone-registry.md                              | +5 (frontmatter)
.claude/rules/moai/design/constitution.md                             | +5 (frontmatter)
.claude/rules/moai/development/manager-develop-prompt-template.md     | +5 (frontmatter)
.claude/rules/moai/workflow/agent-teams-pattern.md                    | +5 (frontmatter)
internal/template/templates/.claude/rules/moai/core/zone-registry.md  | +5 (mirror)
internal/template/templates/.claude/rules/moai/design/constitution.md | +5 (mirror)
internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md | +5 (mirror)
internal/template/templates/.claude/rules/moai/workflow/agent-teams-pattern.md | +5 (mirror)
.moai/specs/SPEC-V3R6-RULES-PATH-SCOPE-001/spec.md                    | status/version/HISTORY
.moai/specs/SPEC-V3R6-RULES-PATH-SCOPE-001/progress.md                | NEW (this file)
.moai/reports/rules-path-scope-simulation-2026-05-22.md               | NEW (M3 deliverable)
```

(internal/template/embedded.go does NOT exist as a file — `internal/template/embed.go` uses `//go:embed templates/**` compile-time embedding; `make build` rebuilds the binary but creates no embedded.go artifact.)

## Token Economy Impact

| Metric | Pre-SPEC | Post-SPEC | Delta |
|---|---|---|---|
| Always-loaded rule words | 13,510 | 5,679 | −7,831 (−58%) |
| Always-loaded rule tokens (~3 t/w) | ~40,530 | ~17,037 | −23,493 (−58%) |
| Per Opus 4.7 1M window utilization | 4.05% | 1.70% | −2.35 percentage points |
| Per Sonnet/Opus standard 200K window | 20.27% | 8.52% | −11.75 percentage points |

In Go-only / general docs sessions (most common path), the full ~23.5K savings materializes per session start.

## Outstanding Items

- **sync-phase**: Hybrid Trunk Tier M = feat branch `feat/SPEC-V3R6-RULES-PATH-SCOPE-001` + auto PR (manager-git delegation, separate phase). Not part of this run-phase scope per delegation prompt.
- **Pre-existing CI baseline failures (5 tests)**: Out of scope per `project_v3r6_template_mirror_drift_audit_2026_05_22` memory — separate `SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001` candidate (deferred).
- **Pre-existing lint baseline (23 issues)**: Out of scope (Go code 0 change).
- **Template mirror baseline drift for manager-develop-prompt-template (+1316B Tier S Applicability)**: Documented §1.4.1, AC-RPS-007 Group B drift modulo PASS, separate sync SPEC candidate.

## Branch / Commit Status (post-run, pre-sync)

Branch not yet created. All edits remain in working tree on `main` (HEAD `c620d0a84`). Per delegation Section D constraint, branch creation + commit + push deferred to sync-phase. Working tree state for this SPEC:

```
M  .claude/rules/moai/core/zone-registry.md
M  .claude/rules/moai/design/constitution.md
M  .claude/rules/moai/development/manager-develop-prompt-template.md
M  .claude/rules/moai/workflow/agent-teams-pattern.md
M  internal/template/templates/.claude/rules/moai/core/zone-registry.md
M  internal/template/templates/.claude/rules/moai/design/constitution.md
M  internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md
M  internal/template/templates/.claude/rules/moai/workflow/agent-teams-pattern.md
M  .moai/specs/SPEC-V3R6-RULES-PATH-SCOPE-001/spec.md
?? .moai/specs/SPEC-V3R6-RULES-PATH-SCOPE-001/progress.md
?? .moai/reports/rules-path-scope-simulation-2026-05-22.md
```

(B8 unrelated working tree files preserved intact per delegation Section D PRESERVE list — no `git add .` or `git add -A` used.)

---

End of progress.md. SPEC v0.2.0 status: implemented.
