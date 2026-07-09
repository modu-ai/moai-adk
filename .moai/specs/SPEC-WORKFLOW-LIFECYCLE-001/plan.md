# SPEC-WORKFLOW-LIFECYCLE-001 — Implementation Plan

> Tier: **L** (large) | doc-only | 대상 파일: live 4 + template mirror 4 (총 8 편집) + `make build`

## §A Context

SPEC-AUDIT-GATE-INTEGRITY-001 P1 백로그 3건을 단일 Tier L doc-only SPEC으로 완결한다. 공통 원리: **SPEC 워크플로 수명주기의 3축(시간/의존성/산물) 골격을 명시적 SSOT로 확정한다** — 확정은 doctrine 레이어에서 일어나고 Go 구현은 후속 SPEC으로 연기된다.

### A.1 편집 대상 파일 매트릭스

| 파일 (live) | 관련 REQ | Template mirror | mirror 경로 |
|-------------|----------|-----------------|-------------|
| `.claude/rules/moai/development/spec-frontmatter-schema.md` | WFL-001, 003, 008 | YES (run-phase 진입 시 재확인) | `internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md` |
| `.claude/rules/moai/workflow/spec-workflow.md` | WFL-005, 007, 010, 011 | YES (재확인) | `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` |
| `.claude/agents/moai/plan-auditor.md` | WFL-009 | YES (재확인) | `internal/template/templates/.claude/agents/moai/plan-auditor.md` |
| `.claude/agents/moai/manager-spec.md` | WFL-001, 002, 003 (amendment authoring 가이드 + `amendment_of:` Optional 필드 서술) | YES (재확인) | `internal/template/templates/.claude/agents/moai/manager-spec.md` |

Mirror 존재는 2026-07-09 `ls` 실측 전제이나 Template 트리는 live의 SUBSET이므로 run-phase 진입 시 재확인 의무 (§C Pre-flight).

### A.2 위험 평가

- **복잡도**: Tier L 5-artifact 세트 + 12 REQ / ~28 AC. doc-only이므로 Run phase는 8파일 편집 + `make build`로 종결.
- **회귀 위험**: LOW — doc-only이므로 Go 동작 무변경. 유일한 회귀는 template neutrality 위반(SPEC ID/REQ 토큰 유입)이며 AC-WFL-028이 가드.
- **병렬 세션 충돌 위험**: MEDIUM — 현재 working tree에 11 modified + 11 untracked 파일이 타 in-flight SPEC에 속함. pathspec 한정 커밋 필수 (feedback_shared_checkout_concurrent_commit_race).

## §B Known Issues / Risks

| # | 리스크 | 경감책 |
|---|--------|--------|
| B1 | 라인 번호 drift — spec.md §A.3의 실측 앵커(`audit.go:355-356`, `relatedness.go:177-180`, `plan-auditor.md:470-476`)는 병렬 세션 커밋으로 이동 가능 | run-phase 진입 시 content-token 앵커(`ValidStatuses`, `If status is already completed`, `## Input Contract`, `planArtifactNames`)로 재탐색. 라인 번호는 참고용 |
| B2 | Template Neutrality 위반 — mirror에 SPEC ID/REQ 토큰 유입 시 CI guard(`template-neutrality-check.yaml`) FAIL | M5에서 mirror별 neutrality grep을 검증 배치에 포함 (AC-WFL-028) |
| B3 | doctrine-vs-Go drift 완전 해결 아님 — REQ-WFL-010는 Go의 4-file을 정합으로 명문화하나 design/research가 여전히 hash에서 제외됨 (수동 skip 입력으로만 서술). 이는 본 SPEC이 유발한 것이 아니라 기존 drift의 정직한 서술 | AC-WFL-022로 4-file 정합 명문화를 검증; design/research를 hash에 넣는 Go 변경은 Out of Scope (후속 SPEC). plan-audit MP-7 Honest Gap Disclosure로 명시 |
| B4 | 공유 checkout 병렬 세션 race — working tree에 무관한 미커밋 변경 다수 (11 modified + 11 untracked, 타 in-flight SPEC) | pathspec 한정 커밋 (`git add <경로> && git commit -- <경로>`)만 사용. `git add -A` 절대 금지 (feedback_shared_checkout_concurrent_commit_race). `credentials.yml` [CRITICAL] 무접촉 |
| B5 | `amendment_of:` 필드가 기존 supersession 패턴과 충돌 가능 — `superseded_by` / `partially_superseded_by` Optional 필드가 이미 존재 | M1에서 세 필드의 역할 분리 명문화: `amendment_of:` (lineage link, 사후 개정), `superseded_by:` (wholesale replacement), `partially_superseded_by:` (partial supersession — 한 SPEC이 여러 후속에게 부분 대체됨). 상호 배타적이지 않으나 의미가 다름 |
| B6 | REQ-WFL-009 Tier-differentiated 입력 계약이 기존 M1 Context Isolation (plan-auditor가 spec.md만 본다는 가정)과 충돌해 보일 수 있음 | M3 편집 시 Context Isolation은 "author reasoning 무시"를 의미하고 artifact 파일 추가는 "Tier에 따른 입력 확장"이므로 양립 가능함을 명시. M1 원문엔 "artifact 입력" 제한이 없고 "reasoning context" 제한만 있음 |
| B7 | plan-auditor.md가 본 SPEC의 감사자다 — M3 편집 후 plan-audit 재실행 시 자기 정의가 바뀌어 있음 | plan-audit는 SPEC 산물을 심사할 뿐 자기 정의 파일과 무관. 단 run-phase에서 M3을 먼저 실행하면 이후 감사가 신규 Input Contract 기준으로 동작함을 인지 |

## §C Pre-flight (run-phase 진입 검증)

run-phase 첫 턴에 단일 병렬 배치로 실행:

```bash
# 1. 결함 baseline 재확인 (R1, R2, R3 gap이 여전히 존재하는가)
grep -c 'ValidStatuses' internal/spec/status.go                                  # 기대: ≥1 (amended 부재 확인 전제)
grep -c '"amended"' internal/spec/status.go                                       # 기대: 0 (새 상태 추가 안함 확인)
grep -c 'depends_on' internal/runtime/audit_cache.go internal/runtime/audit_gate.go internal/runtime/audit_report.go 2>/dev/null | awk -F: '{s+=$2} END {print s+0}'  # 기대: 0 (run-phase 강제 부재 확인)
grep -c 'design\.md\|research\.md' .claude/agents/moai/plan-auditor.md             # 기대: 0 (Input Contract에 언급 없음 확인)
# 1b. 신규 토큰 baseline (전부 0이어야 RED 상태 유지)
grep -c 'amendment_of' .claude/rules/moai/development/spec-frontmatter-schema.md  # 기대: 0
grep -c 'Depends_on Pre-flight' .claude/rules/moai/workflow/spec-workflow.md       # 기대: 0
grep -c 'Tier-differentiated input contract' .claude/agents/moai/plan-auditor.md   # 기대: 0
grep -c 'manual-skip judgment inputs' .claude/rules/moai/workflow/spec-workflow.md # 기대: 0
# 2. mirror 존재 재확인
ls internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md
ls internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md
ls internal/template/templates/.claude/agents/moai/plan-auditor.md
ls internal/template/templates/.claude/agents/moai/manager-spec.md
# 3. git 상태 — origin 발산 확인 (Pre-Spawn Sync Check)
git fetch origin main && git rev-list --count --left-right origin/main...HEAD
# 4. credentials.yml 무접촉 사전 서약 (CRITICAL — 타 in-flight SPEC scope)
ls -la .moai/config/astgrep-rules/security/credentials.yml 2>/dev/null && echo "CRITICAL: must not touch" || echo "not present"
```

어느 baseline이든 이미 수정되어 있으면 (병렬 세션 선점) 해당 REQ를 검증-후-SKIP 처리하고 blocker report로 보고.

## §D Constraints

- **doc-only** — Go 코드 무변경 (spec.md §C 참조). `make build`는 template 재컴파일 목적
- 편집 도구: `Edit` 우선 (Read-before-Edit 준수). `sed`/`awk` 금지
- 커밋: milestone별 pathspec 한정 커밋. 커밋 subject는 `fix(SPEC-WORKFLOW-LIFECYCLE-001): M{N} <요약>` 형식
- 시간 예측 금지 — 우선순위/순서만 명시 (agent-common-protocol.md § Time Estimation)
- 작업 디렉터리: 본 SPEC 산물 ONLY (.moai/specs/SPEC-WORKFLOW-LIFECYCLE-001/). 타 in-flight SPEC 디렉터리 무접촉

## §E Self-Verification (run-phase 완료 기준)

- E1: acceptance.md §D AC matrix 전 행 PASS/FAIL 판정 + 명령 verbatim 출력 인용 (verification-claim-integrity 5-section format)
- E2: `make build` exit 0 (template 편집 후)
- E3: template neutrality grep 0건 (AC-WFL-028)
- E4: `moai spec lint` 본 SPEC 무결 (ERROR-급만 판정; WARNING은 비계수 — vci §3.4 Gaps)
- E5: 편집 diff가 REQ 범위를 벗어나지 않음 (scope discipline — 무관 섹션 무접촉, B4)
- E6: `credentials.yml` 무접촉 (git diff 확인)
- E7: Blocker report (있는 경우) — 구조화된 형식, AskUserQuestion 호출 금지 (subagent boundary)

## §F Milestones

### M1 — delta-spec 수명주기 서술 (REQ-WFL-001..004)

우선순위: High. 대상: `.claude/rules/moai/development/spec-frontmatter-schema.md` + `.claude/agents/moai/manager-spec.md`

1. spec-frontmatter-schema.md § Optional Fields: `amendment_of:` 신규 Optional 필드 추가 (문자열, SPEC ID; self-referential for in-place 또는 parent ID for successor). `prior_completed_sha:` 필드는 HISTORY에 넣고 frontmatter에는 넣지 않음 (frontmatter 인플레이션 방지)
2. spec-frontmatter-schema.md § Status Enum: `completed → in-progress (amendment)` 전이 서술 추가 — `planned` legacy-optional 각주와 동일 패턴으로 "amendment 전이는 in-progress 재사용; 새 상태 없음" 명시
3. spec-frontmatter-schema.md § Status Transition Ownership Matrix: 신규 행 `completed → in-progress (amendment)` / owner: manager-spec / commit subject: `feat(SPEC-{ID}): in-place amendment <rationale>` 추가
4. manager-spec.md § SPEC Frontmatter Canonical Schema Optional fields: `amendment_of:` 서술 추가 + `## Amendments` HISTORY sub-section 작성 가이드 1줄 추가 (REQ-WFL-002)

### M2 — depends_on run-phase 집행 서술 (REQ-WFL-005..007)

우선순위: High. 대상: `.claude/rules/moai/workflow/spec-workflow.md`

1. spec-workflow.md § Phase 0.5 Plan Audit Gate: 신규 sub-section "Depends_on Pre-flight Check" 추가 (plan-auditor 호출 직전 단계; `depends_on:` Optional 필드 로드 → 각 dep status 조회 → unfulfilled 시 AskUserQuestion blocker)
2. 동일 § Phase 0.5: REQ-WFL-006 fulfillment 정의 서술 (status: completed만 충족; 나머지 7값은 미충족)
3. 동일 § Phase 0.5: REQ-WFL-007 blocker 3-option 서술 (wait / override `--ignore-deps` + `.moai/logs/depends-on-override.log` / abort). `--ignore-deps` flag와 log path 리터럴 토큰 명시

### M3 — Tier L 산물 집합 + plan-auditor 입력 계약 (REQ-WFL-008, 009)

우선순위: High. 대상: `.claude/rules/moai/development/spec-frontmatter-schema.md` + `.claude/agents/moai/plan-auditor.md`

1. spec-frontmatter-schema.md § Optional Fields `tier:` 항목: 5-file 집합 명시적 열거 ("Tier L = spec.md + plan.md + acceptance.md + design.md + research.md") + spec-workflow.md § SPEC Complexity Tier cross-reference
2. plan-auditor.md § Input Contract: Tier-differentiated 재작성 — "Tier L (또는 부재 시 Tier L backward-compat): plan-auditor는 design.md AND research.md를 추가로 읽는다 (리터럴 토큰 `Tier L: design.md + research.md are required inputs` + `Tier-differentiated input contract`); Tier M: primary trio (spec/plan/acceptance); Tier S: spec.md + plan.md"
3. plan-auditor.md § Input Contract: M1 Context Isolation과의 양립 명시 ("reasoning context 무시"는 artifact 파일 추가와 충돌하지 않는다)

### M4 — plan-artifact hash subject list 정합 (REQ-WFL-010, 011)

우선순위: High. 대상: `.claude/rules/moai/workflow/spec-workflow.md`

1. spec-workflow.md § Report Persistence skip-eligibility 항목: hash subject list를 4-file `{spec.md, plan.md, acceptance.md, tasks.md}`로 명시 (`internal/runtime/audit_cache.go` `planArtifactNames` verbatim 정합)
2. 동일 항목: tasks.md가 V3R4-era 잔재이며 V3R6 Tier L이 design/research로 대체했다는 사실 명시 + design.md/research.md는 `manual-skip judgment inputs` (리터럴 토큰)이라는 서술 추가
3. 동일 항목: REQ-WFL-011 amendment시 hash 변경 + cache invalidation 서술 (리터럴 토큰 `cache-invalidating event`)

### M5 — Template mirror 4건 + make build + 전체 검증 배치 (REQ-WFL-012)

우선순위: High (M1-M4 완료 후). 순서 의존: M1-M4 → M5

1. 4개 mirror에 동일 편집 적용 + Neutrality strip (SPEC-WORKFLOW-LIFECYCLE / REQ-WFL 토큰, 내부 날짜·감사 인용 제거)
2. `make build` 실행, exit 0 확인
3. acceptance.md §D 전체 AC 검증 배치를 단일 턴 병렬 실행 (file-redirect contract 준수)
4. milestone별 pathspec 커밋 정리 확인

## §G Anti-Patterns

- **AP-1**: `amendment_of:`를 새 상태 `amended`와 혼동하여 enum 확장 (REQ-WFL-001 명시 위반 — 본 SPEC은 in-progress 재전이로 해결)
- **AP-2**: depends_on pre-flight를 별도 Phase 0.6으로 신설 (단계 인플레이션 — REQ-WFL-005는 Phase 0.5 sub-step 확장)
- **AP-3**: mirror 편집 시 live 파일 내용을 그대로 복사 (Neutrality 위반 — SPEC ID/내부 날짜 유입)
- **AP-4**: `git add -A` 또는 디렉터리 광역 add (병렬 세션 미커밋 변경 휩쓸림, B4)
- **AP-5**: AC 검증을 "수정했으니 통과할 것" 추정으로 보고 (vci §1.1 위반 — 본 SPEC의 R3c가 정확히 이 병리의 교정이므로 자기 모순)
- **AP-6**: `credentials.yml` 접근 (CRITICAL — 타 in-flight SPEC scope, 무접촉 서약)
- **AP-7**: REQ-WFL-010의 hash subject list를 5-file로 "수정하여" doctrine-Go 일치시키기 (Go 변경은 Out of Scope — 본 SPEC은 현행 Go 정합을 명문화만 한다)
- **AP-8**: amendment_of/superseded_by/partially_superseded_by 세 필드를 동의어로 취급 (B5 — 의미가 다름)

## §H Cross-References

- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 (defect claim은 도구 검증 전까지 가설 — research.md가 이를 준수)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` (12-field SSOT, Status Enum, Status Transition Ownership Matrix, Optional Fields)
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier / § Phase 0.5 Plan Audit Gate / § Report Persistence
- `internal/spec/status.go` `ValidStatuses` (8값 enum — 무변경 참조 대상)
- `internal/spec/audit.go:355-356` (completed no-drift predicate — 무변경 참조 대상)
- `internal/spec/drift.go:121` (completed vs git predicate — 무변경 참조 대상)
- `internal/runtime/audit_cache.go:63-68` `planArtifactNames` (4-file hash subject — 무변경 참조 대상, 정합의 SSOT)
- `internal/bodp/relatedness.go:177-180` (depends_on BODP Signal A 소비 — 무변경 참조 대상)
- `.claude/agents/moai/plan-auditor.md` § Input Contract / § Tier-differentiated PASS threshold (M3 편집 대상)
- `.claude/agents/moai/manager-spec.md` § SPEC Frontmatter Canonical Schema (M1 편집 대상)
- `CLAUDE.local.md` §2.1 Template Content Neutrality + §15 16-언어 동등 + §25 Template Internal-Content Isolation
- 선행 SPEC: `.moai/specs/SPEC-AUDIT-GATE-INTEGRITY-001/{spec,plan,acceptance}.md` (P0 4건 완결 — 본 SPEC의 P1 3건 출처)
- 감사 provenance: 3-agent 병렬 감사 2026-07-09 (agent-definitions / workflow-doctrine / SDD 웹 리서치) + sync-audit 2026-07-09
