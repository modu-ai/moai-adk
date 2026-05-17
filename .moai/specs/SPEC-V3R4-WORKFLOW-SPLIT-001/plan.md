# Implementation Plan — SPEC-V3R4-WORKFLOW-SPLIT-001

## Plan Summary

**SPEC**: SPEC-V3R4-WORKFLOW-SPLIT-001
**Strategy**: 4-Wave sequential PR (lessons #9 wave-split)
**Harness**: thorough (plan-auditor PASS + evaluator-active cross-validation 필수)
**Base branch**: `main` HEAD `7a118e6b2`
**Working branch**: `plan/SPEC-V3R4-WORKFLOW-SPLIT-001` (current)
**Estimated PR count**: 1 plan-PR + 4 wave-PRs = 5 PRs

각 Wave는 독립 sub-PR로 admin squash merge. main에 머지된 직후 다음 Wave 진입. 단, Wave 1 (run.md) 머지 후에는 lesson learned + REQ-WFSP-005c regression test 결과를 반영하여 후속 Wave 작업 패턴 보정 권장.

---

## Wave 0 — Preparation (선행 작업, plan-PR 머지 후 즉시)

### Goal

Wave 1-4 공통 인프라 (cross-ref validation, template mirror scaffold, audit test) 를 사전 구축하여 Wave별 PR이 audit-clean 상태로 시작하도록 보장.

### Tasks

- **T0.1 — Cross-reference audit script 작성**
  - 파일: `scripts/audit-workflow-split.sh`
  - 기능: 모든 sub-skill `.md` 파일에서 `Read /path/to/...`, `${CLAUDE_SKILL_DIR}/...`, 상대경로 reference를 grep으로 추출 → 실제 파일 존재 여부 검증 → 0건이면 PASS
  - 외부 의존: bash + grep만 (CI 환경 호환)

- **T0.2 — LOC ceiling assertion Go test 작성**
  - 파일: `internal/skills/workflow_split_test.go`
  - 테스트 케이스:
    - `TestSubSkillLOCCeiling`: workflows/{name}/*.md 모두 ≤500 LOC
    - `TestEntryRouterLOCCeiling`: workflows/{run,sync,project,plan}.md 모두 ≤200 LOC
    - `TestTemplateMirrorParity`: local과 template 디렉토리 파일 목록 비교
  - 실행: `go test ./internal/skills/...`

- **T0.3 — Template mirror dir scaffold**
  - 디렉토리 생성: `internal/template/templates/.claude/skills/moai/workflows/{run,sync,project,plan}/`
  - 빈 placeholder는 만들지 않음 (Wave별 sub-skill과 함께 동시 생성)

- **T0.4 — SKILL.md Intent Router byte-for-byte 검증 baseline**
  - `git rev-parse HEAD:.claude/skills/moai/SKILL.md` hash 기록 → Wave별 PR 종료 시 동일 hash 검증

- **T0.5 — Slash command regression baseline 캡처 메커니즘 확립** (plan-auditor M1 mitigation)
  - 목표: AC-WFSP-004 (slash command regression 0) 검증 경로 확정. `--dry-run` flag 부재 위험 사전 해소.
  - Step 1 — `--dry-run` flag 존재 여부 검증: `grep -rE "--dry-run|dryRun" .claude/commands/{plan,run,sync,project}.md` 및 `internal/cli/{plan,run,sync,project}.go` (또는 동등 entry). 4개 슬래시 커맨드 각각 dry-run 지원 여부 확정.
  - Step 2 — `--dry-run` 미지원 시 trace mode 설계: 각 workflow entry skill 본문 진입 직전에 `MOAI_TRACE_PHASES=1` env 가드로 phase 실행 시작/종료를 stderr에 1줄씩 기록 (예: `[trace] /moai run Phase 0.5 enter`). 변경은 entry router 4개 파일에만 국한 (~10 LOC 추가).
  - Step 3 — Baseline trace 캡처: 4개 슬래시 커맨드를 main HEAD `7a118e6b2` 기준으로 1회씩 실행 (trivial scope: e.g., `/moai plan --help` 또는 `--dry-run`이 있으면 사용, 없으면 `MOAI_TRACE_PHASES=1` env로 실행). trace를 `.moai/reports/baseline-trace/<command>-main-7a118e6b2.txt`에 저장 → Wave별 PR 종료 시 동일 명령으로 trace 재생산 + diff 0건 검증.
  - Acceptance: Wave 1 진입 전 4개 슬래시 커맨드 baseline trace 파일 4개 (`.moai/reports/baseline-trace/{plan,run,sync,project}-main-7a118e6b2.txt`) 생성 + Wave별 verification fixture로 사용.

### Acceptance

- T0.1 script가 현재 main HEAD에서 0건 broken link 보고 (현재 split 전 baseline)
- T0.2 test가 현재 main HEAD에서 4개 workflow 모두 ceiling 위반 (1073/1203/1076/932 LOC) 보고 — 의도된 baseline failure (split 후 PASS로 전환됨을 검증할 fixture)
- T0.3 디렉토리 4개 생성, `git status` 에서 untracked dir로 인식
- T0.5 baseline trace 4개 파일 생성 (4개 슬래시 커맨드 main HEAD `7a118e6b2` 기준)

### Risks (Wave 0)

- audit-script가 false-positive를 일으킬 경우 후속 Wave가 잘못 차단 → 사전에 다양한 reference 패턴 (markdown link, code-fence path, plain text)을 test fixture에 포함

---

## Wave 1 — run.md split (3 sub-skills + entry router)

### Goal

`workflows/run.md` 1073 LOC → 3 phase-scoped sub-skills + ≤200 LOC entry router. Template mirror 동시 진행.

### Sub-skill Decomposition

| Sub-skill | LOC est. | Source lines | Content |
|-----------|----------|--------------|---------|
| `run/context-loading.md` | ~210 | 1-214 | Purpose / Scope / Input / Mode Dispatch / UltraThink / Harness Level / Context Loading / Worktree Path Rules |
| `run/phase-execution.md` | ~500-745 | 215-958 | Phase Sequence / Audit Run N of total / Task Decomposition. **If estimated >500 LOC after verification, further split into `run/phase-execution.md` + `run/task-decomposition.md`** |
| `run/mode-orchestration.md` | ~120 | 959-1073 | Execution Mode Gate / Team Mode Routing / Context Propagation / Completion / Test Scenarios / Custom Harness Extension |

### Tasks

- **T1.1 — Actual LOC verification**: `awk 'NR==215,NR==958' run.md | wc -l` 등 실제 phase 콘텐츠 LOC 측정. Estimate와 실제값 차이 5% 이내면 그대로 진행, 그 외에는 boundary 재조정
- **T1.2 — 3 sub-skill 파일 생성** (`workflows/run/{context-loading,phase-execution,mode-orchestration}.md`)
  - frontmatter: `user-invocable: false`, `name: moai-workflow-run-{sub}`, `description`, `metadata.parent: moai-workflow-run`
  - 본문: source 줄 범위 그대로 cut+paste, 단 내부 H2 → H1 승격, anchor 깨짐 수정
- **T1.3 — Entry router refactor**: `workflows/run.md` 를 ≤200 LOC thin router로 재작성
  - frontmatter 보존 (기존 metadata, user-invocable: true)
  - 본문 구조:
    ```
    ## Phase Sequence (Router)
    Phase 0: Context Loading → Read workflows/run/context-loading.md
    Phase 1-N: Phase Execution → Read workflows/run/phase-execution.md
    Mode Routing: → Read workflows/run/mode-orchestration.md
    ```
  - Phase별 실행 prelude (각 phase에서 sub-skill을 `Read`로 로드) 포함
- **T1.4 — Template mirror**: `internal/template/templates/.claude/skills/moai/workflows/run.md` + `run/*.md` 4개 파일 1:1 복사
- **T1.5 — make build**: `make build` 실행 → `embedded.go` 재생성 → `git status` 에서 변경 확인
- **T1.6 — Audit & test**: T0.1 script + T0.2 Go test 실행, 모두 PASS 확인
- **T1.7 — Regression smoke test**: `/moai run SPEC-DUMMY` invocation (dry-run, agent spawn skipped) phase trace 비교
- **T1.8 — PR 생성**: feat/SPEC-V3R4-WORKFLOW-SPLIT-001-wave-1 branch, base=main, auto-merge SQUASH

### Acceptance

- AC-WFSP-001 (모든 sub-skill ≤500 LOC) — phase-execution.md가 745 LOC 추정이면 추가 split 필수
- AC-WFSP-002 (entry router ≤200 LOC)
- AC-WFSP-005a (`moai spec lint --strict` clean)
- AC-WFSP-005c (regression smoke test PASS)
- AC-WFSP-005d (LOC ceiling test PASS)

### Risks (Wave 1)

| Risk | 영향도 | 완화책 |
|------|--------|--------|
| `phase-execution.md` 745 LOC가 ≤500 LOC 한계 초과 | High | T1.1에서 사전 검증, 초과 시 sub-split (`task-decomposition.md` 분리) — 2 sub-skill → 4 sub-skill로 증가 가능 |
| 내부 anchor link 깨짐 (`#phase-3-1`) | Medium | T1.2에서 grep으로 모든 anchor 추출 → cross-ref audit |
| Template mirror 누락 → 사용자 프로젝트 회귀 | High | T1.4 후 `find` 비교 검증, T1.6 audit이 parity test 포함 |
| make build 후 embedded.go 변경 누락 commit | Medium | T1.5 후 `git diff --stat` 확인 (must include `internal/template/embedded.go`) |

---

## Wave 2 — sync.md split (3 sub-skills + entry router)

### Goal

`workflows/sync.md` 1203 LOC → 3 sub-skills + ≤200 LOC entry router. Wave 1 lesson learned 반영.

### Sub-skill Decomposition

| Sub-skill | LOC est. | Source lines | Content |
|-----------|----------|--------------|---------|
| `sync/quality-gates.md` | ~540 | 1-538 | Purpose / Scope / Input / Mode Flag / Supported Modes+Flags / Context Loading / Phase 0~0.7 inc. HUMAN GATEs. **If exceeds 500 LOC, split Phase 0~0.4 / Phase 0.5~0.7** |
| `sync/doc-execution.md` | ~205 | 539-743 | Phase 1 Analysis and Planning + HUMAN GATE Documentation Scope + Phase 2 Execute |
| `sync/delivery.md` | ~460 | 744-1203 | Phase 3 Git Operations / SPEC Reference / Context / Affected Areas / Local CI Mirror / Team Mode / Graceful Exit / Completion / Test Scenarios / Custom Harness |

### Tasks

T2.1 ~ T2.8 — Wave 1 T1.1 ~ T1.8 패턴 동일 적용. `sync/quality-gates.md` 540 LOC 추정이 borderline이므로 T2.1 LOC verification 결과에 따라 split 결정.

### Acceptance

Wave 1과 동일 + AC-WFSP-005c (sync regression smoke test).

### Risks (Wave 2)

| Risk | 영향도 | 완화책 |
|------|--------|--------|
| `quality-gates.md` 540 LOC borderline | High | T2.1에서 정확한 측정, 초과 시 sub-split. HUMAN GATE는 한 sub-skill 안에 묶는 것이 가독성 우선 |
| sync 워크플로우는 Phase 0 HUMAN GATE 3개 + Phase 1 GATE 1개 + Phase 3 stash GATE 1개 = total 5개 — split으로 GATE 흐름 단절 위험 | Critical | entry router에 "Phase 0 GATEs covered in quality-gates.md §HUMAN GATE 1-3" 같은 GATE map을 명시적으로 기록 |

---

## Wave 3 — project.md split (4 sub-skills + entry router)

### Goal

`workflows/project.md` 1076 LOC → 4 sub-skills + ≤200 LOC entry router.

### Sub-skill Decomposition

| Sub-skill | LOC est. | Source lines | Content |
|-----------|----------|--------------|---------|
| `project/mode-detection.md` | ~200 | 1-197 | Mode Flag Compatibility / Scope Boundary / Flag --from-brain / Phase 0 / Phase 0.3 |
| `project/codebase-analysis.md` | ~115 | 198-310 | Phase 1 / Phase 1.5 + 3 Rounds / Phase 2 User Confirmation |
| `project/doc-generation.md` | ~470 | 311-776 | Phase 3 / 3.1 / 3.3 / 3.5 / 3.7 / 4.1a / 4 Completion |
| `project/meta-harness.md` | ~300 | 777-1076 | Phase 5 Socratic / Phase 6 meta-harness invocation |

### Tasks

T3.1 ~ T3.8 — Wave 1 패턴 동일. 4 sub-skill로 가장 많지만 모두 ≤500 LOC 안전 마진. `doc-generation.md` 470 LOC가 가장 큰 single sub-skill.

### Acceptance

Wave 1과 동일 + project regression smoke test (`/moai project --explain`).

### Risks (Wave 3)

| Risk | 영향도 | 완화책 |
|------|--------|--------|
| `doc-generation.md` 4-1a `_brain` ingestion phase가 다른 sub-skill (`mode-detection.md` `--from-brain` flag) 와 cross-ref | Medium | entry router에 명시적 sub-skill map 표기 + cross-ref audit (T0.1) 검증 |
| Phase 1.5의 3-round AskUserQuestion이 codebase-analysis.md 안에 모두 포함되어야 함 (split 시 round-1/2/3 분리 금지) | Medium | T3.1 LOC 측정 시 round 콘텐츠가 한 파일에 머무는지 확인 |

---

## Wave 4 — plan.md split (3 sub-skills + entry router)

### Goal

`workflows/plan.md` 932 LOC → 3 sub-skills + ≤200 LOC entry router. **본 SPEC의 자기-참조 (self-referential) Wave** — 본 SPEC을 작성한 plan workflow 자체를 변형하므로, run-phase 진입 후 본 Wave 진입 시 작업자는 변경 영향을 신중히 검토.

### Sub-skill Decomposition

| Sub-skill | LOC est. | Source lines | Content |
|-----------|----------|--------------|---------|
| `plan/context-discovery.md` | ~140 | 1-141 | Purpose / Scope / Input / Supported Flags / Mode Flag Compatibility / Context Loading / Brain Context Auto-Detection |
| `plan/clarity-interview.md` | ~475 | 142-615 | Phase Sequence Phases 1A/0.3/0.3.1/0.4/0.5/1.25/1B + Annotation Cycle + Decision Point 1 + Round subsections + Clarity Score |
| `plan/spec-assembly.md` | ~315 | 616-932 | Phase 1.5 / Phase 2 / Phase 2.3 / Phase 2.5 / Phase 3 / Phase 3.5 / Phase 3.6 / Decision Points 2-3.5 / Team Mode Routing / Completion / Test Scenarios / Custom Harness |

### Tasks

T4.1 ~ T4.8 — Wave 1 패턴 동일. `clarity-interview.md` 475 LOC borderline 주의.

### Acceptance

Wave 1과 동일 + plan regression smoke test (`/moai plan "dummy feature"` dry-run, 단 실제 SPEC 작성 stop).

### Risks (Wave 4)

| Risk | 영향도 | 완화책 |
|------|--------|--------|
| 본 Wave 작업 중 plan workflow 자체 변경 → 현재 진행 중인 다른 SPEC plan-phase 영향 | High | Wave 4를 가장 마지막에 배치. 모든 active plan SPEC이 머지된 후 진입. PR 머지 시점에 다른 plan-PR open 0건 확인 |
| `clarity-interview.md` 475 LOC + Annotation Cycle + Decision Point 1 묶음이 ≤500 borderline | Medium | T4.1 LOC verification, 초과 시 `plan/annotation-cycle.md` 별도 분리 (4 sub-skill로 증가) |
| SKILL.md Intent Router (`"plan"` 키) byte-for-byte 무변경 검증 | Critical | T0.4 baseline hash 비교 |

---

## Cross-Wave Common Tasks

### spec-compact.md auto-generation (optional, plan-PR 시)

plan workflow §spec-compact.md Auto-Generation 규칙에 따라, plan-auditor가 본 SPEC을 검토한 후 spec-compact.md를 자동 생성할 수 있음. 수동 생성 불필요.

### Plan-auditor PASS gate

본 plan-PR은 harness=thorough이므로 plan-auditor 독립 검토 PASS 필수. R1 verdict PASS 또는 R2 PASS (one revision) 후 PR 머지.

### Evaluator-active cross-validation

thorough harness triggers `cross_validate_with_evaluator_active`. plan-auditor와 evaluator-active 두 독립 검토가 모두 PASS여야 plan-PR 머지 가능.

---

## Technology Approach

### Core constraints

- **Pure markdown reorganization**: 코드 변경 0건. 단 audit/test 인프라는 Go test (`internal/skills/workflow_split_test.go`) 작성
- **No external tool dependency**: bash + Go stdlib만 사용 (CI 호환)
- **Template-First rule (CLAUDE.local.md §2)**: 모든 변경은 `internal/template/templates/` 에 먼저 적용 → `make build` → local copy 자동 동기화

### Build flow per Wave

```
1. Edit workflows/{name}/*.md (sub-skill) + workflows/{name}.md (entry router)
2. Mirror to internal/template/templates/.claude/skills/moai/workflows/...
3. make build → regenerates internal/template/embedded.go
4. go test ./internal/skills/... (LOC ceiling + cross-ref audit)
5. bash scripts/audit-workflow-split.sh (final validation)
6. moai spec lint --strict
7. git add → commit → push → PR
```

---

## MX Plan (high fan_in functions tracking)

본 SPEC은 markdown reorganization이므로 Go code @MX tag 변경 minimal. 단 다음 함수는 sub-skill 로딩 mechanism과 관련하여 @MX:ANCHOR 검토 대상:

- `internal/template/templates.go` (embedded.go 생성 entry point)
  - 후보: @MX:NOTE — sub-skill 디렉토리 trees가 embed FS에 포함되는지 sanity check
- `internal/skills/loader.go` (if exists — skill loading entry)
  - 후보: @MX:ANCHOR — invariant "sub-skill은 entry router 경유로만 로드"
- `internal/cmd/moai_init.go` (moai init 명령 entry)
  - 후보: @MX:NOTE — Wave 4 이후 `moai init` 시 workflow/{name}/*.md 디렉토리 배포 확인

Run-phase 진입 시 위 함수 fan_in 확인 후 정식 @MX tag 추가. Plan-phase에서는 후보만 식별.

---

## Risk Analysis (Cross-Wave)

| Risk ID | Risk | 영향도 | 완화 |
|---------|------|--------|------|
| R1 | Cross-reference breakage in MEMORY/lessons | High | T0.1 audit script + entry router 경로 보존 (REQ-WFSP-003a/d) |
| R2 | Template sync 누락으로 사용자 프로젝트 회귀 | Critical | Wave별 T*.4 mirror + T*.5 make build + audit (REQ-WFSP-004) |
| R3 | Sub-skill LOC overflow (future maintenance 시) | Medium | T0.2 ceiling test가 CI에서 영구 enforce, 미래 PR이 위반 시 자동 block |
| R4 | `moai init` 사용자 프로젝트 regression | Critical | Wave별 PR 머지 후 사후 `moai init /tmp/test-X` 실행 + sub-skill 배포 확인 |
| R5 | SKILL.md Intent Router 의도치 않은 변경 | Critical | T0.4 baseline hash 비교, Wave별 PR 마지막에 검증 |
| R6 | Phase content boundary 잘못 인지 → sub-skill에 phase 누락 | High | Wave별 T*.1 LOC verification + diff 비교 (전체 source LOC = sum(sub-skill LOC) + delta) |
| R7 | Wave 4 작업 중 plan workflow 자체 변경이 현재 active plan-PR 충돌 | High | Wave 4를 마지막 배치 + open plan-PR 0건 확인 |

---

## Mitigation Summary

- 모든 Wave는 audit-test (T0.1, T0.2) 가 자동으로 PASS/FAIL 검증
- 모든 Wave PR은 plan → run → sync 정상 lifecycle 경유 (hot patch 금지)
- main HEAD를 매 Wave 종료 시점에 갱신 후 다음 Wave fork (stacked PR 금지 — base 충돌 방지)
- Wave 1 후 lesson learned가 후속 Wave 작업 패턴에 반영됨 (특히 LOC borderline 판정 기준)

---

## Implementation Sequence

```
plan-PR (본 PR)
   ├─ plan-auditor R1 → PASS or R2
   ├─ evaluator-active cross-validate → PASS
   └─ admin squash merge → main 진입

Wave 0 (preparation infra)
   └─ T0.1~T0.4 → PR → admin squash → main

Wave 1 (run.md split, 3+1 sub-skills)
   └─ T1.1~T1.8 → PR → CI green → admin squash → main

Wave 2 (sync.md split, 3+1 sub-skills)
   └─ T2.1~T2.8 → PR → CI green → admin squash → main

Wave 3 (project.md split, 4+1 sub-skills)
   └─ T3.1~T3.8 → PR → CI green → admin squash → main

Wave 4 (plan.md split, 3+1 sub-skills) [self-referential, LAST]
   └─ T4.1~T4.8 → PR → CI green → admin squash → main

Post-completion verification
   ├─ moai init /tmp/test-post-split 실행 → sub-skill 배포 확인
   ├─ /moai plan|run|sync|project 각 1회 invocation regression
   └─ Total LOC reduction 측정 (4284 → ~800 router + 13 sub-skill total)
```

---

## Out of Scope (deferred to follow-up SPECs)

- **F-015 docs drift chore** (hooks-system.md UserPromptExpansion/PostToolBatch/mcp_tool 5 events missing) — 별도 chore PR
- **SPEC-V3R4-LLM-REVIEW-CI-001** (Claude/GLM/Gemini/Codex CLI not found in CI runner) — 별도 SPEC follow-up
- **SPEC-V3R4-WORKFLOW-SPLIT-001-DOCS-FOLLOWUP** (조건부 생성) — 만약 run-phase에서 docs-site reference 발견 시
- **Sub-skill 추가 분할 (5+ split)** — 본 SPEC은 13 sub-skill 까지 분할. 만약 사용 추이상 추가 분할 필요 시 별도 SPEC

---

## References

- spec.md § Goals/EARS Requirements
- design.md § Architectural Decisions
- acceptance.md § Quality Gates
- scenarios.md § Wave-by-Wave Test Plans
- `.moai/research/workflow-audit-2026-05-16.md` Bundle F
- lessons #9 wave-split
- CLAUDE.local.md §2 Template-First Rule, §18 Git Workflow
