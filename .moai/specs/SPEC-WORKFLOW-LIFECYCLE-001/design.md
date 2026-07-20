# SPEC-WORKFLOW-LIFECYCLE-001 — Design Decisions

> Tier L design artifact. 세 결함(R1 delta-spec / R2 depends_on / R3 Tier L 산물+입력계약) 각각에 대한 설계 결정, 기각 대안, 그리고 codebase reality와의 정합 근거를 서술한다. 모든 설계는 doc-only 원칙 하에 Go 동작 무변경을 전제로 한다.

## §A Design Decisions

### A.1 R1 — delta-spec 수명주기 (completed 사후 개정 전이)

#### A.1.1 채택 설계 — `completed → in-progress` 재전이 + `amendment_of:` Optional 필드 + HISTORY `## Amendments` sub-section

**상태 기계(state machine)**:

```
(none) → draft → in-progress → implemented → completed
                                  ↑                ↓
                                  └── amendment ───┘ (REQ-WFL-001)
                                          │
                                          ├── in-place: amendment_of: <self-ID>
                                          └── successor: amendment_of: <parent-ID> (신규 SPEC)

completed → superseded (기존 — wholesale replacement)
completed → archived (기존 — administrative cleanup)
```

**핵심 설계 원칙 3가지**:

1. **새 상태 값을 추가하지 않는다 (`amended` 도입 기각)**. `ValidStatuses`는 현재 8값(`internal/spec/status.go` 실측). 9번째 값을 추가하면:
   - `internal/spec/era.go` era 분류 로직 재검토 필요
   - `internal/spec/audit.go` completed-no-drift predicate (`audit.go:355-356`) 재검토 필요
   - lint `OwnershipTransitionRule` 확장 필요
   - 기존 88개 pre-v3 SPEC들의 frontmatter와의 호환성 검토 필요
   
   본 SPEC은 **기존 `in-progress` 상태를 재사용**한다 — `in-progress`는 이미 "활발 작업 중"을 의미하므로 사후 개정 시맥락에 자연스럽게 부합. 구분자는 frontmatter의 `amendment_of:` 필드와 HISTORY의 `## Amendments` 행으로 제공.

2. **`amendment_of:` 필드는 선언적 lineage link**. 두 패턴을 모두 커버:
   - **In-place amendment** (값 = self-ID): `SPEC-X-001`이 자기 자신을 개정 — `amendment_of: SPEC-X-001`. 원본 산물을 그 자리에서 고친다.
   - **Successor amendment** (값 = parent-ID): `SPEC-X-002`가 `SPEC-X-001`을 개정 — `amendment_of: SPEC-X-001`. 원본은 `superseded`로 전이(기존 패턴)하거나 그대로 두고 새 SPEC이 lineage를 이어받는다.
   
   이 두 패턴은 OpenSpec의 "propose" 단계에 해당한다. "apply"는 run→sync 재착수, "archive"는 HISTORY의 prior_completed_sha 보존.

3. **HISTORY `## Amendments` sub-section이 증거 저장소**. frontmatter는 스키마 인플레이션을 피하기 위해 최소 정보(lineage link)만 보유하고, 상세한 개정 정보(prior version, prior SHA, rationale, scope)는 HISTORY의 `## Amendments` 행에 보관. 이는 기존 HISTORY 테이블의 자연스러운 확장이며, 새로운 데이터 구조를 도입하지 않는다.

#### A.1.2 기각 대안

- **(a) 새 `amended` 상태 추가**: 기각 — 위에 서술한 호환성 비용. enum 인플레이션의 실익이 없다 (in-progress 재전이로 동일 의미 표현 가능).
- **(b) completed → completed 직접 전이 (version bump만)**: 기각 — `completed → completed`는 "아무 일도 일어나지 않았다"와 "개정이 일어났다"를 구분할 수 없다. 감사/era 분류가 전이를 인식하지 못한다. 또한 drift.go의 completed predicate와 충돌 (`frontmatterStatus == "completed" && gitStatus != "completed"`가 개정 도중 발화).
- **(c) 오직 `superseded_by` 패턴만 사용**: 기각 — `superseded_by`는 wholesale replacement를 전제한다. 사후 개정(부분 수정, 내용 보강, 버그 픽스)은 replacement가 아니다. 또한 `superseded` SPEC은 감사 대상에서 제외되므로(era_final: true) 개정된 내용이 감사 게이트를 통과할 수 없다.

#### A.1.3 audit/era 상호작용

`internal/spec/audit.go:355-356` "If status is already completed, no drift"는 completed 상태에서만 발화. 개정 도중 status는 `in-progress`이므로 이 predicate는 발화하지 않는다 — 개정 도중 SPEC은 정상적으로 drift detection 대상이 된다. era 분류 (`internal/spec/era.go`)는 `§E.2` 마커와 frontmatter 기반 — 개정이 진행 중이어도 `era: V3R6` 필드는 보존되므로 era 분류는 안정적. **Go 변경 불필요**, doctrine 서술만으로 충분 (REQ-WFL-004).

### A.2 R2 — depends_on run-phase 집행 (Phase 0.5 pre-flight 차단)

#### A.2.1 채택 설계 — Phase 0.5 sub-step "Depends_on Pre-flight Check" 확장

**평가 모델(fulfillment evaluation model)**:

```
Phase 0.5 (기존 Plan Audit Gate)
  ├── [신규] sub-step 0: Depends_on Pre-flight Check
  │     1. load <SPEC-ID>/spec.md frontmatter
  │     2. parse depends_on: [SPEC-Y-001, SPEC-Z-002, ...]
  │     3. for each dep:
  │          load <dep-ID>/spec.md frontmatter
  │          read status:
  │          if status != "completed":
  │            → unfulfilled, add to blocker list
  │     4. if blocker list non-empty:
  │          → AskUserQuestion (wait / override / abort)
  │     5. else: proceed to sub-step 1 (plan-auditor)
  └── [기존] sub-step 1: plan-auditor (M1 Context Isolation, Tier-differentiated per REQ-WFL-009)
```

**핵심 설계 원칙 3가지**:

1. **fulfillment = `status: completed` 단일 조건** (엄격한 정의). 8개 상태 중 오직 `completed`만 "충족"으로 간주. `implemented`도 충족이 아님 — `implemented`는 "run은 끝났으나 sync(문서화, CHANGELOG, 최종 승인)가 끝나지 않은" 상태이므로 의존성이 완전히 충족되었다고 볼 수 없다. 부분 점수, "거의 완료" 해석, 점수 기반 bypass 모두 금지 (REQ-WFL-006).

2. **Phase 0.5의 첫 sub-step으로 확장 (0.6 신설 아님)**. 별도 Phase를 두면:
   - 단계 인플레이션 (0.5, 0.6, 1.0 ...)
   - 우회 경로 복잡도 (0.6을 건너뛰는 예외 경로 정의 필요)
   - skip-eligibility 4-condition predicate와의 상호작용 복잡도 (0.5 skip이 0.6에도 적용되는가?)
   
   Phase 0.5 sub-step 0으로 두면 기존 skip-eligibility는 그대로 적용 (skip이 결정되면 pre-flight도 skip). 단 skip 시에도 depends_on이 변경되지 않았음이 전제 (frontmatter가 hash subject이므로 변경 시 hash 변경 → skip 불가).

3. **blocker는 AskUserQuestion 3-option** (subagent가 아니라 orchestrator가 방출). manager-develop은 pre-flight를 직접 실행하지 않는다 — pre-flight는 orchestrator-side 단계. 실패 시 orchestrator가 AskUserQuestion으로 wait/override/abort를 제시. `--ignore-deps` override는 감사 로그(`.moai/logs/depends-on-override.log`)에 미충족 dep ID + rationale을 기록 (단순 flag가 아니라 감사 가능한 결정).

#### A.2.2 기각 대안

- **(a) BODP Signal A 확장**: 기각 — BODP(`internal/bodp/relatedness.go`)는 branch-origin 결정("이 SPEC에 브랜치를 만들까?")이지 run 진입 허가("이 SPEC을 실행할까?")가 아니다. 의미가 다르다. BODP는 이미 depends_on을 Signal A 코드 의존성 heuristic으로 사용 중이며 확장 시 역할 혼란.
- **(b) Phase 0.6 신설**: 기각 — 위에 서술한 단계 인플레이션 비용.
- **(c) lint rule로 강제**: 기각 — `moai spec lint`는 SPEC 산물 무결성을 검사하는 정적 분석 도구. depends_on은 runtime 의존성(live 상태 조회 필요)이지 정적 속성이 아니다. dep의 status는 동적이므로 lint가 캐시할 수 없다.
- **(d) 사이클 탐지 + 위상 정렬**: 기각 — 본 SPEC 범위 밖 (Out of Scope 명시). 사이클이 존재하면 pre-flight는 단순히 "모든 dep이 completed인가?"를 각각 조회만 한다. 순회 시 무한 루프 위험은 후속 SPEC이 topological sort로 해결.

#### A.2.3 구현 범위

**doc-only**: spec-workflow.md § Phase 0.5에 sub-step 서술 추가. orchestrator의 실제 구현(Go 코드 또는 Skill 로직)은 Out of Scope. doctrine이 명확해지면 후속 SPEC이 Go/Skill 레이어를 배선.

### A.3 R3 — Tier L 산물 집합 + plan-auditor 입력 계약

#### A.3.1 채택 설계 — 3면 SSOT 동기화 + Go 정합 명문화

**3면(three surfaces)**:

1. **`spec-frontmatter-schema.md` § Optional Fields `tier:` 항목**: 5-file 집합을 명시적으로 열거. 현재 이 항목은 "S/M/L LEAN tier"라고만 되어 있고 산물 목록이 없다 — 독자가 spec-workflow.md § SPEC Complexity Tier로 교차 참조해야 한다. 본 SPEC이 산물 목록을 직접 임베드 (cross-reference와 함께) 하므로 SSOT가 자체 완결된다.

2. **`plan-auditor.md` § Input Contract를 Tier-differentiated로 재작성**: 현재 (plan-auditor.md:470-476) "spec.md primary + MAY read acceptance/plan"으로 되어 있어 Tier L의 design.md/research.md가 입력 계약에서 누락. 이것이 "계약이 암묵적/미명시"인 근원. 본 SPEC이 Tier-differentiated 계약을 명문:
   - **Tier L**: 5-file 전부 읽기 (spec + plan + acceptance + design + research)
   - **Tier M**: primary trio (spec + plan + acceptance)
   - **Tier S**: spec + plan (AC inline)
   
   계약은 "MUST read" (MAY 아님)로 명세 — Tier L 감사에서 design/research를 읽지 않으면 MP-7 Honest Gap Disclosure 위반.

3. **`spec-workflow.md` § Report Persistence hash subject list를 Go 정합으로 명문화**: 현재 doctrine은 "spec.md / plan.md / acceptance.md / research.md / design.md" (5-file)로 서술하고 Go는 `{acceptance, plan, spec, tasks}` (4-file)를 해싱 — drift가 존재. 본 SPEC은 **Go 구현을 정합의 SSOT로 명문화**:
   - hash subject list = 4-file `{spec.md, plan.md, acceptance.md, tasks.md}` (`internal/runtime/audit_cache.go` `planArtifactNames` verbatim)
   - `tasks.md`는 V3R4-era 잔재 (V3R6 Tier L이 design/research로 대체했으나 hash에는 남아 있음 — backward compat)
   - `design.md` / `research.md`는 `manual-skip judgment inputs` — 기계 hash 대상이 아니라 orchestrator의 수동 skip 판단 입력 (기존 language를 리터럴 토큰으로 승격)
   - **Go 변경은 Out of Scope** — 본 SPEC은 현행 Go 정합을 정직하게 명문화. hash를 5-file로 확장하려면 별도 후속 SPEC이 Go를 변경.

#### A.3.2 기각 대안

- **(a) Go의 `planArtifactNames`를 5-file로 확장**: 기각 — doc-only 원칙 위반. 또한 기존 cached verdict들이 모두 재계산되어야 하는 비용. 본 SPEC은 drift를 "정직한 서술"로 드러내고 Go 변경은 별도 SPEC으로 연기.
- **(b) `tasks.md`를 doctrine에서 제거**: 기각 — 일부 V3R4-era grandfathered SPEC은 여전히 tasks.md를 가질 수 있으며 (이것이 planArtifactNames에 포함된 이유), 제거하면 backward compat이 깨짐.
- **(c) Tier L 입력 계약을 "MAY read"로 유지**: 기각 — "MAY"는 plan-auditor가 design/research를 읽지 않아도 PASS를 줄 수 있음을 의미. Tier L의 design.md/research.md가 입력으로 의미가 있으려면 "MUST read"여야 한다.

#### A.3.3 Context Isolation과의 양립

plan-auditor의 **M1 Context Isolation** ("reasoning context 무시")은 artifact 파일 입력을 제한하지 않는다 — "context"는 "author의 사고 과정/대화 이력/초안"을 의미하며, "파일"은 SPEC 산물이다. Tier L 감사가 design.md/research.md를 읽는 것은 Context Isolation 위반이 아니다. M3 편집 시 이 양립을 명시적으로 서술 (B6 경감).

### A.4 R1 + R3 교차 — amendment시 plan-artifact hash 변경

REQ-WFL-011는 R1과 R3의 교차 사례: SPEC이 in-place 개정(`spec.md` 수정)되면 plan-artifact hash가 변경(spec.md가 hash subject이므로) → cached skip verdict가 invalidate → 다음 `/moai run`에서 Phase 0.5 plan-audit가 재실행. 이는 자연스러운 결과이지만 doctrine에 명시적으로 서술되어야 orchestrator가 amendment를 cache-invalidation 트리거로 인식 (리터럴 토큰 `cache-invalidating event`).

## §B Alternatives Considered Summary

| # | 대안 | 상태 | 기각 이유 |
|---|------|------|-----------|
| 1 | R1: 새 `amended` status 추가 | 기각 | enum 인플레이션 + era/audit 호환 비용 |
| 2 | R1: completed → completed 직접 전이 | 기각 | drift.go predicate 충돌 + 전이 인식 불가 |
| 3 | R1: 오직 `superseded_by`만 사용 | 기각 | 사후 개정(부분 수정) 시나리오 미커버 |
| 4 | R2: BODP Signal A 확장 | 기각 | 의미 불일치 (branch-origin vs run 허가) |
| 5 | R2: Phase 0.6 신설 | 기각 | 단계 인플레이션 + skip 정책 복잡도 |
| 6 | R2: lint rule 강제 | 기각 | runtime 의존성 (정적 속성 아님) |
| 7 | R2: 사이클 탐지 + 위상 정렬 | 기각 | 본 SPEC 범위 밖 (Out of Scope) |
| 8 | R3: Go planArtifactNames 5-file 확장 | 기각 | doc-only 원칙 위반 + cached verdict 재계산 비용 |
| 9 | R3: tasks.md doctrine 제거 | 기각 | V3R4 grandfathered SPEC backward compat 훼손 |
| 10 | R3: Tier L 입력 "MAY read" 유지 | 기각 | 입력 계약의 의미 약화 |

## §C Cross-References

- `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12-field SSOT, Status Enum, Status Transition Ownership Matrix, Optional Fields (M1, M3 편집 대상)
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier / § Phase 0.5 Plan Audit Gate / § Report Persistence (M2, M3, M4 편집 대상)
- `.claude/agents/moai/plan-auditor.md` § Input Contract / § M1 Context Isolation / § Tier-differentiated PASS threshold (M3 편집 대상)
- `.claude/agents/moai/manager-spec.md` § SPEC Frontmatter Canonical Schema (M1 편집 대상 — amendment_of 서술 + HISTORY 가이드)
- `internal/spec/status.go` `ValidStatuses` — 8값 enum, 무변경 참조 대상
- `internal/spec/audit.go:355-356` — completed no-drift predicate, 무변경 참조 대상
- `internal/spec/drift.go:121` — completed vs git predicate, 무변경 참조 대상
- `internal/runtime/audit_cache.go:63-68` `planArtifactNames` — 4-file hash subject SSOT, 무변경 참조 대상 (Go 정합의 기준)
- `internal/bodp/relatedness.go:177-180` — depends_on Signal A 소비, 무변경 참조 대상
- OpenSpec propose→apply→archive 패턴 — R1 설계 영감 (외부 SDD 표준)
- GitHub Spec Kit cross-artifact consistency — R2 설계 영감 ("gate must actually gate")
- Kiro 3-artifact 모델 — R3 대조군 (MoAI Tier L 5-artifact 확장의 기원)
- 선행 SPEC: `.moai/specs/SPEC-AUDIT-GATE-INTEGRITY-001/` (P0 완결 — 본 SPEC P1 3건 출처)
