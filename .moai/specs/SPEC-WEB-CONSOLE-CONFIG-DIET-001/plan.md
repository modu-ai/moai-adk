# plan.md — SPEC-WEB-CONSOLE-CONFIG-DIET-001

> Web Console Config 과다 노출 정리 — Dead-Config 다이어트. 구현 계획, 마일스톤, 그룹별 결정 옵션.
> 시간 추정 금지 — 우선순위/의존 순서로만 표기.

## §A. Context

SPEC-WEB-CONSOLE-011 M2가 콘솔을 약 163 필드로 확장했고, `seamSectionFields()`가 각 yaml 섹션의 모든
키를 편집 가능하게 렌더한다. 5-에이전트 감사 + 본 SPEC의 재-grep(research.md)로 dead-config 그룹을
확정했다. 목표는 **노출 축소(diet)** — 렌더 SSOT(`internal/settings/schema.go` /
`internal/settings/schema_sections.go`)에서 dead 필드를 필터/강등/제거한다. **배선(wiring)은 범위 밖.**

## §B. Known Issues / 주의점

- **B1. N3 오삭제 위험 (최상위)**: git-strategy 다수 키는 skill-body가 yaml-direct read로 소비한다
  (`internal/config/types.go:53`, `:71`). 이들을 삭제하면 skill 프롬프트 동작이 깨진다. git-strategy는
  **제거 대상이 아니라 "중복 3-mode 노출 축소" 대상**이다 (REQ-CD-021).
- **B2. read-only vs 은닉의 트레이드오프**: N2(미배선 게이트)는 향후 배선 시 부활할 수 있다. 완전 제거보다
  read-only 강등(값은 보이되 편집 불가 + no-effect 주석)이 사용자 투명성에 유리하나, 필드 수가 많아
  콘솔이 장황해질 수 있다. 그룹별 결정(§F)에서 판단.
- **B3. era 보호**: 본 SPEC은 V3R6 3-phase(plan→run→sync) 대상 신규 SPEC이다. 기존 grandfather SPEC과
  무관.
- **B4. team.auto_selection "consumed" 주장 미확인**: 감사의 "consumed but not exposed" 주장이 재-grep으로
  확인되지 않았다 (research.md §R.5). REQ-CD-050(SHOULD)의 해당 키 노출은 소비자 재확인을 전제로 한다.
- **B5. 정합성 결함의 수정 방향 선택**: REQ-CD-040(harness prune)·042(db auto_sync)는 "코드를 config에
  맞춤" vs "표시/문서를 코드에 맞춤" 2방향이 있다. §F.4에서 각각 권장안 제시.

## §C. Pre-flight (구현 착수 전 재확인 명령)

run-phase 진입 시 아래를 **재실행**하여 앵커 drift를 재검증한다 (실측 앵커 재검증 의무):

```bash
ROOT=/Users/goos/.moai/worktrees/moai-adk-go/preview-wc011   # 또는 landing 후 main
# 1. research 사강 재확인
grep -rn "\.Research\." internal/ pkg/ --include="*.go" | grep -v "_test.go"        # expect: empty
ls internal/research/ 2>&1                                                          # expect: No such file
# 2. workflow_agents 사강
grep -rn "\.WorkflowAgents\b" internal/ pkg/ --include="*.go" | grep -v "_test.go"  # expect: empty
# 3. TrustGate 미배선
grep -rn "NewTrustGate" internal/ --include="*.go" | grep -v "_test.go" | grep -v "core/quality/"  # expect: empty
# 4. harness prune 하드코딩
grep -n "defaultRetentionDays" internal/harness/observer.go                          # expect: = 30
# 5. role_profiles.model dispatch 미배선
grep -rn "LoadRoleProfiles" internal/ cmd/ --include="*.go" | grep -v "_test.go" | grep -v "team_spawn.go:"  # expect: empty
# 6. 현재 렌더 필드 카운트 baseline
go test ./internal/settings/... -run . -count=1                                      # baseline green
```

## §D. Constraints

- [HARD] 변경 레버는 렌더 SSOT 2파일 + 결함 수정 시 `internal/harness/observer.go`,
  `internal/config/types.go`(주석), 세션 명명 경로. templ/TUI에 개별 필터 추가 금지 (REQ-CD-002).
- [HARD] N3(skill-body 소비) 키 삭제 금지 (REQ-CD-003).
- [HARD] `role_profiles.*.isolation`(라이브, `moai workflow lint` 소비) 유지 (REQ-CD-034).
- [HARD] 라이브 유지 키 보존: quality `development_mode`/`test_coverage_target`, ralph
  `lint_as_instruction`/`warn_as_instruction`, observability `enabled`/`slow_hook_threshold_ms`,
  git-strategy `mode`/`hooks.pre_push`/`automation.auto_*`/`main_branch`, git-convention 5키 전부.
- run-phase 종료 시 두 표면 테스트 green + zero lint errors.

## §E. Self-Verification

acceptance.md §D의 AC-CD-0xx가 검증 SSOT. 각 마일스톤 종료 시 해당 AC의 grep/go-test를 실행하여
PASS를 관측(verification-claim-integrity: 실행 후 출력 관측).

## §F. Milestones (우선순위 순)

의존 그래프: M0 → M1 → M2 → M3 → M4 (M5는 선택, M1~M4 비차단). 각 마일스톤은 독립 커밋.

### M0 — 결정 확정 + 스키마 테스트 하네스 (선행)

- 각 그룹의 처리 방식(은닉 / read-only / 제거)을 §F.5 결정 매트릭스로 확정 (오케스트레이터/사용자 승인 필요
  항목은 착수 승인 게이트에서).
- 스키마 카운트/부재 검증용 테스트 스캐폴드 추가 (`schema_test.go`): 특정 섹션 키가 `allFields()` 결과에
  없음을 assert 하는 헬퍼.

### M1 — Tier 1 무효과 제거 (최우선, 신뢰 훼손)

REQ-CD-010/011/012. 가장 위험한 그룹(사용자가 바꿔도 무효과)을 먼저 제거.

- **research (12키)**: `AllSections()`에서 `SectionResearch` 제거 + `seamSectionFields()`의 research 블록
  (schema_sections.go:283–294) 제거 + `SchemaSectionIDs()`에서 제거.
- **db 인터뷰 3키**: `seamSectionFields()`의 db 블록(schema_sections.go:314–316) 제거 또는 read-only 강등.
- **security 스칼라 3키**: security 블록(schema_sections.go:309–311) 제거 또는 read-only 강등.

### M2 — Tier 2 대량 사강

REQ-CD-020/021/022/023.

- **quality 무효과**: `qualityExtraFields()`(schema_sections.go:181–203)에서 TrustGate 의존 무효과 키
  강등. `development_mode`/`test_coverage_target` 유지.
- **git-strategy 중복 3-mode**: `gitStrategyFields()`(schema_sections.go:131–157)에서 비선택 mode 프로파일의
  중복 노출 축소 — **키 삭제 아님**(N3). mode-조건부 렌더로 전환하거나 비활성 프로파일을 collapsed/read-only.
- **workflow_agents 14키**: `agentSettingsFields()`(schema_sections.go:352–358)의 `workflow_agents` 루프
  제거 또는 read-only.
- **ralph 중첩**: `seamSectionFields()`의 ralph `ast_grep.*`/`loop.*`/`lsp.*`/`loop.completion.*`
  (schema_sections.go:265–280) 은닉/강등. `lint_as_instruction`/`warn_as_instruction` 유지.

### M3 — Tier 3 부분 사강

REQ-CD-030/031/032/033/034.

- **workflow dead accessors 12키**: workflow 블록의 `auto_clear.*`/`token_budget.*`/`loop_prevention.*`/
  `worktree.auto_merge`(schema_sections.go:225–242 일부) 강등.
- **harness 6키**: `escalation.*`/`mode_defaults.*`/`auto_detection.enabled`(schema_sections.go:250–256 일부) 강등.
- **observability 5키**: `trace_dir`/`report_dir`/`max_file_size_mb`/`retention_days`/`hook_metrics.output_path`
  (schema_sections.go:301–305) 강등. `enabled`/`slow_hook_threshold_ms` 유지.
- **llm 4키**: `performance_tier`/`claude_models.{high,medium,low}`(llmFields, schema_sections.go:162–172) 강등.
- **role_profiles model/mode/effort**: `agentSettingsFields()`(schema_sections.go:344–350)의 model/effort/mode
  seam 필드 강등. **isolation 유지**.

### M4 — 정합성 결함 수정

REQ-CD-040/041/042/043. §F.4 결정 반영.

- harness prune, security @MX 주석, db auto_sync 서술, session_name_pattern.

### M5 — (선택, SHOULD) 과소 노출 라이브 키 표면화

REQ-CD-050. **소비자 재확인 전제**. `team.auto_selection.min_*`는 production read 확인 후에만 노출;
확인 실패 시 스킵. `llm.glm.base_url`도 동일.

### §F.4 정합성 결함 수정 방향 (권장안 — 확정은 착수 승인 시)

| 결함 | 옵션 A | 옵션 B | 권장 |
|------|--------|--------|------|
| REQ-CD-040 harness prune 90/30 | prune이 `learning.log_retention_days` config를 읽도록 배선 | 표시/문서를 30으로 정정 | **B** (다이어트 원칙 = 배선 아님; 저위험). 단 A가 사용자 기대에 더 부합 → 착수 승인에서 확인 |
| REQ-CD-041 security @MX stale | 주석을 실측(0 refs)에 맞게 정정/강등 | (해당 없음) | **A** |
| REQ-CD-042 db auto_sync 드리프트 | 산문 서술을 nested 형태로 정정 | 렌더를 스칼라로 | **A** (실제 yaml이 nested; 신뢰도 낮아 doc-level) |
| REQ-CD-043 session_name_pattern | 세션 명명이 config를 읽도록 배선 | config 키를 read-only 강등 | **B** (배선은 범위 밖; round-trip 전용 필드는 편집 불가로) |

### §F.5 그룹별 다이어트 결정 매트릭스 (은닉 / read-only / 제거)

> **PROPOSE — 은밀히 확정하지 않음.** 아래는 근거 있는 권장이며, 착수 승인 게이트에서 사용자가 그룹별로
> override 가능. 기본 규칙: **N1(진짜 사강) → 제거 또는 은닉; N2(미배선) → read-only(향후 배선 대비 yaml
> 키 보존); N3(skill-body 소비) → 유지, 중복만 축소.**

| 그룹 | 결(nuance) | 권장 처리 | 근거 |
|------|-----------|----------|------|
| research 12 | N1 (패키지 삭제) | **은닉**(콘솔). struct 완전삭제는 out of scope(§D) | 부활 가능성 0; 그러나 struct 삭제는 별도 큰 결정 |
| db 인터뷰 3 | N1 (채움소스 폐지) | **은닉** | `/moai db init` 폐지, 소비자 0 |
| security 스칼라 3 | N1/N2 (하드코딩·미배선) | **read-only + no-effect 주석** | 보안 키라 "숨김"보다 투명 표시가 안전 |
| quality 무효과 | N2 (TrustGate 미배선) | **read-only** | 향후 TrustGate 배선 시 부활 대비 |
| git-strategy 3-mode | N3 (skill 소비) | **유지 + mode-조건부 렌더**(중복 축소) | 삭제 금지; 비선택 mode만 collapsed |
| workflow_agents 14 | N1/N2 | **은닉** | JS 런타임 미소비 + typed reader test-only |
| ralph 중첩 | N1 (엔진 미소비) | **은닉** | engine.go가 flat 플래그만 읽음 |
| workflow dead 12 | N1 (accessor 호출 0) | **은닉** | dead-code accessor |
| harness 6 | N2 (test-only manager) | **read-only** | 향후 escalation 배선 대비 |
| observability 5 | N1 (하드코딩 미러) | **은닉** | 상수 미러, 부활 불가 |
| llm 4 | N1 (dispatch 미소비) | **은닉** | 모델은 settings.json에서 해석 |
| role_profiles model/mode/effort | N2 (dispatch 미배선) | **read-only** (isolation은 유지) | 향후 dispatch 배선 대비 |

## §G. Anti-Patterns

- **AP-1**: git-strategy N3 키를 사강으로 오인해 삭제 → skill 프롬프트 파손. (B1)
- **AP-2**: `role_profiles.*.isolation`을 model/mode와 함께 뭉뚱그려 강등 → `moai workflow lint` 파손.
- **AP-3**: `ResearchConfig` struct + yaml 키를 이 SPEC에서 무단 완전삭제 → out of scope 위반(§D).
- **AP-4**: 결함 수정을 "배선"으로 확대해석 → 다이어트 범위 초과 (REQ-CD-040/043은 read-only/문서정정이 기본).
- **AP-5**: 감사의 "consumed" 주장을 재검증 없이 라이브로 신뢰 → REQ-CD-050의 team.auto_selection 오노출.
- **AP-6**: templ/TUI에 개별 하드코딩 필터 추가 (렌더 SSOT 우회) → REQ-CD-002 위반.

## §H. Cross-References

- spec.md §B (REQ-CD-0xx), §D (Out of Scope).
- acceptance.md §D (AC-CD-0xx 검증 술어).
- research.md (file:line 증거).
- `internal/settings/schema.go`, `internal/settings/schema_sections.go`.
- `internal/harness/observer.go`(REQ-CD-040), `internal/config/types.go:429`(REQ-CD-041).
