# SPEC-V3R5-WORKFLOW-OPT-001 — Vision Document

**Source**: W3 HARNESS-AUTONOMY-001 운영 후 메타-분석 (2026-05-20)
**Status**: Research artifact — manager-spec 위임 source of truth
**Parent lessons**: `~/.claude/projects/.../memory/feedback_w3_metaanalysis_lessons.md`

---

## 1. Background

W3 HARNESS-AUTONOMY-001 run-phase 측정 결과 wall-time **91분** (목표 30분 대비 **+200%**). 4 결함 패턴이 manager-develop 재위임 3회 + orchestrator 직접 fix 1회를 유발.

### W3 결함 timeline (90분 → ideal 30분의 60분 격차 원인)

| 결함 | 직접 손실 | 재위임 cycle | 검증 wait |
|------|----------|--------------|-----------|
| spec-lint h3 heading 누락 | 4분 | 0 | 3분 |
| AC-HRA-009 V3R4 retirement 충돌 | 14분 | 1 (manager-develop 2차) | 10분 |
| Windows syscall.Flock build tag 누락 | 10분 | 0 (orchestrator 직접) | 5분 |
| observer.go path resolution bug | 5분 (cleanup) | 0 | 0 |
| **합계** | **33분 직접 + 18분 wait** | **+1 cycle** | |

### 측정된 wall-time 분포 (총 91분 = 5460s)

- **위임 작업 실시간**: 46분 (50.5%) — manager-develop 1차 31분 + 2차 14분 + orchestrator 직접 fix 10분
- **검증/대기/결정 시간**: 45분 (49.5%) — orchestrator 검증 직렬 + CI 대기 + AskUserQuestion 결정 + 사용자 응답

→ **위임 대비 검증/대기가 1:1 비율 = 병목**

---

## 2. Goals (8 Layer 전체 완료)

### 2.1 정량 목표

| 지표 | 현재 (W3) | 목표 |
|------|----------|------|
| Run-phase wall-time | 91분 | ≤ 30분 |
| manager-develop 위임 횟수 (1-pass 성공률) | 3회 (33%) | ≤ 1회 (≥ 80%) |
| CI 직렬 대기 시간 | 15분 | ≤ 3분 |
| 검증 직렬 시간 | 10분 | ≤ 3분 |

### 2.2 정성 목표

- 알려진 결함 (cross-platform syscall, cross-SPEC 충돌, spec-lint 규약)이 위임 prompt에 자동 주입
- 독립 패키지 implementation은 병렬 가능
- CI 대기 동안 orchestrator는 다른 작업 진행 가능
- plan-auditor가 cross-SPEC + cross-platform discipline을 사전 검증

---

## 3. Architecture — 4 영역 분류

| 영역 | Layers | 산출물 유형 | 위험도 | 의존성 |
|------|--------|------------|--------|--------|
| **R (rule-only)** | A (잔여), C, D, E, H | `.claude/rules/moai/` markdown + template 동기화 | Low | 독립 |
| **C (config)** | B | `.moai/config/sections/workflow.yaml` + experimental flag | Medium | R 권장 |
| **G (Go code — capture)** | F | `internal/harness/capture/` 확장 (defect pattern detection) | High | R, C |
| **A (agent prompt — auditor)** | G | `.claude/agents/plan-auditor.md` + `.moai/config/evaluator-profiles/` | Medium-High | R |

### Milestone 권장 구조 (manager-spec 입력)

- **M1**: R 영역 통합 — A 잔여 (template 동기화) + C + D + E + H 5개 rule 단일 milestone
- **M2**: B Config — workflow.yaml + experimental flag + role_profiles definitions
- **M3**: F Code — `internal/harness/capture/` defect pattern detection 확장
- **M4**: G Agent — plan-auditor D7 (cross-SPEC) + D8 (cross-platform) dimension 추가
- **M5**: Integration — 본 SPEC 자체 적용으로 wall-time 측정 (자체-validation)
- **M6**: Documentation + lessons memory archive

---

## 4. Detailed Layer Plans

### Layer A — 위임 Prompt 품질 향상 (잔여 작업)

**이미 적용**: `.claude/rules/moai/development/manager-develop-prompt-template.md` 생성됨 (5-section 구조 + 8 known issues B1-B8).

**잔여 작업**:
- [HARD] Template 동기화: `internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md` 복사
- [HARD] `make build` 실행하여 embedded files 재생성
- Cross-reference: `.claude/skills/moai/workflows/run.md` § Phase 1 에서 본 template 참조 추가
- evaluator-active 위임 prompt에도 동일 5-section 구조 적용 (별도 변형 template 또는 본 template 재사용)

### Layer B — 병렬 위임 (Agent Teams)

**범위**:
- `.moai/config/sections/workflow.yaml`: `team.enabled: true` + 5 role_profiles 정의 (capture-impl / tier-impl / safety-impl / throttle-impl / seeds-impl + tester + reviewer)
- `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` 환경변수 설정 doc
- `.claude/rules/moai/workflow/agent-teams-pattern.md` (신규): 5-teammate 패턴 표준화

**Acceptance**:
- W4 또는 후속 SPEC에서 5 independent package 병렬 implementation 시도 가능
- TeamCreate → 5 implementer teammates + 1 tester + 1 reviewer

**위험**: Agent Teams 실험적 기능. Fallback to solo mode 필요.

### Layer C — Background CI / Concurrent Operations

**범위**:
- `.claude/rules/moai/workflow/ci-watch-protocol.md` 확장: `gh pr checks --watch` + `run_in_background: true` 표준화
- `polling pattern (sleep + check)` 사용 금지 명시 (anti-pattern)
- orchestrator self-discipline rule: CI 대기 중 다른 작업 동시 진행

**Acceptance**:
- 위임 후 CI 대기 시간이 다른 productive work로 채워짐 (idle time 제거)
- AskUserQuestion 응답 대기 시간을 다른 작업으로 활용

### Layer D — 검증 병렬화

**범위**:
- `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution 강화
- orchestrator self-discipline: 모든 read-only verification은 단일 turn multi-Bash
- 신규 rule: `verification-batch-pattern.md` (검증 항목 그룹화)

**Acceptance**:
- 7 검증 항목 (test/coverage/grep/sentinel/CLI/benchmark/lint)을 1 turn에 병렬 실행
- evaluator-active 자동 spawning (manager-develop 결과와 동시 검증)

### Layer E — Plan-Run Pipeline (선행 시작)

**범위**:
- `.claude/rules/moai/workflow/spec-workflow.md` § Phase Transitions 수정
- Plan PR review/CI 중에도 run-phase 시작 허용 (rebase 인지)
- Plan Audit Gate 정책: high-confidence PASS는 재실행 skip 표준화

**Acceptance**:
- W4 시작 시 W3 sync PR이 아직 머지 안 돼도 W4 plan 작성 시작 가능
- plan-auditor PASS ≥ 0.90 시 run-phase Plan Audit Gate skip

### Layer F — Lessons 자동 capture + 주입

**범위 (코드 변경)**:
- `internal/harness/capture/` 에 defect pattern detection 추가
  - input.CWD path validation
  - cross-platform syscall 사용 detection
  - cross-SPEC reference 검증
- 발견된 defect를 lessons memory에 자동 append
- manager-develop 위임 prompt 생성 시 lessons memory에서 keyword 매칭으로 자동 prepend

**Acceptance**:
- W4 진입 시 본 W3 결과 4 결함이 자동으로 위임 prompt에 주입됨
- 새 결함 발견 시 lessons memory에 자동 entry 추가 (orchestrator 수동 작성 불요)

**위험**: SubagentStop hook 부하 증가 가능성. 비동기 처리 필수.

### Layer G — plan-auditor D7/D8 dimension 강화

**범위 (agent prompt 변경)**:
- `.claude/agents/plan-auditor.md` 확장
- 신규 D7: Cross-SPEC Reconciliation
  - SPEC 본문에서 참조한 다른 SPEC (V3R4 등)의 현재 status 확인
  - retired/superseded 충돌 자동 검출
- 신규 D8: Cross-Platform Discipline
  - syscall 사용 시 build tag 명시 의무 검증
  - file path resolution 패턴 검증 ($CLAUDE_PROJECT_DIR 사용 권장)
- `.moai/config/evaluator-profiles/` 에 D7/D8 dimension 가중치 추가

**Acceptance**:
- W3 같은 케이스 재발 시 plan-auditor가 iter 1에서 BLOCKING으로 잡음
- spec-lint h3 규약, build tag 누락 등을 plan-phase에서 사전 검출

### Layer H — Tool 최적화

**범위 (rule)**:
- `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution 강화
- `gh pr checks --json` + `jq` 조합 권장 패턴 추가
- ToolSearch preload pattern 표준화 (세션 시작 + per-turn)
- `git log --format=...` 한 줄 명령 통합 패턴

**Acceptance**:
- 모든 orchestrator agent들이 단일 message multi-Bash 활용
- ToolSearch 호출 횟수 최소화

---

## 5. Draft Acceptance Criteria (EARS format candidates)

manager-spec가 다음을 EARS format로 변환:

### Ubiquitous (always-true)
- **REQ-WO-001**: 모든 manager-develop 위임 prompt는 `manager-develop-prompt-template.md` 5-section 구조 준수
- **REQ-WO-002**: 모든 read-only verification은 단일 turn multi-Bash로 실행
- **REQ-WO-003**: CI 대기 중 orchestrator는 idle 상태 금지 (다른 작업 또는 background watch)

### State-driven
- **REQ-WO-010 (When)**: agent_teams flag enabled, independent packages 5+ 일 때, manager-develop 5-teammate 병렬 spawn 가능
- **REQ-WO-011 (When)**: plan-auditor verdict PASS ≥ 0.90 시, run-phase Plan Audit Gate 재실행 skip
- **REQ-WO-012 (When)**: 새 defect 발견 시, capture 패키지가 lessons memory에 자동 entry 추가

### Event-driven
- **REQ-WO-020 (When defect detected by manager-develop)**: lessons memory에 자동 append (Layer F)
- **REQ-WO-021 (When plan-PR opened)**: plan-auditor가 D7 cross-SPEC 자동 스캔 (Layer G)
- **REQ-WO-022 (When SPEC body references syscall package)**: plan-auditor D8가 build tag 의무 검증

### Optional
- **REQ-WO-030 (Where high-stakes SPEC)**: evaluator-active 자동 병렬 spawning

### Binary AC candidates
- **AC-WO-001**: 본 SPEC 자체 W3 wall-time 91분 → ≤ 30분 (자체 dogfooding)
- **AC-WO-002**: manager-develop 위임 prompt에 8 known issues B1-B8 자동 포함 (template 적용 확인)
- **AC-WO-003**: `internal/template/templates/` 미러 + `make build` 통과
- **AC-WO-004**: plan-auditor가 cross-SPEC 충돌 (V3R4 retirement 시뮬레이션) 자동 검출
- **AC-WO-005**: capture 패키지 defect pattern detection 단위 테스트 ≥ 90% coverage

---

## 6. Risks

| Risk | 확률 | 영향 | 완화 |
|------|------|------|------|
| R1: Agent Teams 실험적 기능 instability | Medium | High | Solo mode fallback 의무화 + 단계적 적용 (W4부터 시범) |
| R2: lessons 자동 capture의 false positive | Medium | Medium | Heuristic-only (LLM 미사용), 신뢰도 threshold 설정 |
| R3: plan-auditor D7/D8 over-strict | Low | Medium | iter 1 결과를 user review (기존 패턴 유지) |
| R4: 본 SPEC 자체가 W4 PROJECT-MEGA-001과 scope 중복 | Low | Low | W4는 harness self-improvement (model layer), 본 SPEC은 orchestrator workflow (meta layer). 분리 가능 |
| R5: template 동기화 누락 | Low | High | CI guard test (template diff vs runtime rule) |
| R6: Background CI watch가 main session blocking | Low | Medium | `run_in_background: true` + notification 패턴 검증 |

---

## 7. Out of Scope (EXCL)

- **EXCL-WO-001**: W4 PROJECT-MEGA-001의 harness self-improvement 기능 (별도 SPEC)
- **EXCL-WO-002**: PR #1024 (W3 run) merge 자체 (사용자 자연 머지 영역)
- **EXCL-WO-003**: `internal/harness/observer.go` path resolution bug fix (별도 SPEC — `SPEC-V3R5-OBSERVER-PATH-001` 가칭)
- **EXCL-WO-004**: Lint baseline 2 issues cleanup (W1/W2 잔재, 별도 SPEC — `SPEC-V3R5-LINT-DEBT-001` 가칭)
- **EXCL-WO-005**: GitHub webhook 활용 (advanced CI integration, follow-up)
- **EXCL-WO-006**: 다른 영역 메타-분석 (예: SPEC-creation phase 자체 최적화)

---

## 8. Cross-SPEC Reconciliation

본 SPEC가 영향 받는 영역과 충돌 가능 SPEC:

| 영역 | 관련 SPEC | 영향 |
|------|----------|------|
| `internal/harness/capture/` | W3 SPEC-V3R5-HARNESS-AUTONOMY-001 | Layer F가 capture 패키지 확장 — W3 PRESERVED 약속과 충돌 가능. 새 file 추가 (defect detection) 형태로 회피 가능 |
| `.claude/agents/plan-auditor.md` | SPEC-V3R4-PLAN-AUDIT-001 (가칭, 존재 여부 확인 필요) | Layer G가 dimension 추가. 기존 dimension은 PRESERVE |
| `.moai/config/sections/workflow.yaml` | SPEC-V3R4-WF-003, WF-004 (mode dispatch) | Layer B가 team config 추가. 기존 mode dispatch는 PRESERVE |
| `.claude/rules/moai/workflow/spec-workflow.md` | SPEC-V3R2-SPC-001 (EARS hierarchical AC) | Layer E가 Phase Transitions 수정. AC schema는 무관 |

본 SPEC는 기존 SPEC들의 retirement 또는 reversal 없이 add-only 형태로 진행 가능.

---

## 9. Pre-flight Checklist (manager-spec가 plan.md 작성 시 검증)

- [ ] 8 Layer 모두 REQ로 변환 (rule/config/code-F/code-G 4 영역)
- [ ] 6 milestones 명시 (M1 rule통합 / M2 config / M3 capture / M4 plan-auditor / M5 self-validation / M6 docs)
- [ ] Brownfield strategy: PRESERVE 대상 파일 enumeration (capture.go, plan-auditor.md 기존 dimensions)
- [ ] EXCL-WO-001..006 explicit list
- [ ] R1-R6 risk mitigation 매핑
- [ ] Cross-SPEC reconciliation (§8) 본문 반영
- [ ] `## 4. Out of Scope` h2 + `### 4.1 Exclusion List` h3 sub-section (W3 학습)
- [ ] Frontmatter 12-field canonical schema

---

**Vision document created**: 2026-05-20
**Author**: GOOS Kim (via MoAI orchestrator + W3 meta-analysis)
**Next step**: manager-spec 위임 → SPEC artifacts (spec.md/plan.md/acceptance.md/spec-compact.md/progress.md) 생성
