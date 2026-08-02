---
id: SPEC-PHASE-FIELD-VALIDATION-001
title: "phase 프론트매터 필드의 값-형태 검증과 오염 코퍼스 교정 — 진행"
version: "0.2.0"
status: completed
created: 2026-08-02
updated: 2026-08-02
author: Goos Kim
priority: P2
phase: "v3.0.2"
module: "internal/spec, .claude/agents/moai"
lifecycle: spec-anchored
tier: M
tags: "spec-lint, frontmatter, phase, drift, authoring-guard"
---

# 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** (frontmatter `tier: M` 명시 — spec.md / plan.md / acceptance.md + progress.md)
- REQ **16**건 (Tier M 상한 16), AC **16**건 (상한 16), 미커버 REQ 0건
  (acceptance.md §H 대조표)
- 마일스톤 4개: M1 값-형태 검증 / M2 저작 지시 / M3 코퍼스 교정 14건 / M4 회귀 가드
- 신규 finding 코드: `FrontmatterPhaseInvalid` — 강등 대상 집합에 **미등록**
  (설계의 핵심; AC-PFV-003이 기계 판정)

### 저작 시점 실측 기준선

- `spec.md` 전수: 564개(이 SPEC 포함), 부정 토큰 정확 일치 위반 **9**건
- 전체 산출물 부정 토큰 위반 **31**건, 걸친 SPEC **20**개
  (spec.md 오염 9 + 형제만 오염 레거시 11)
- 부분 문자열 판정 시 추가 오탐 **8**건 → 정확 일치 채택 근거
- 엄격 semver 허용목록 시 오탐 **301**건 (310 − 9) → 허용목록 기각 근거
- `moai spec lint --json`: 62건, error **0**건, exit 0
  (`MissingExclusions` 24 / `StatusGitConsistency` 16 / `FrontmatterInvalid` 14 /
  `LegacyEARSKeyword` 7 / `OwnershipTransitionInvalid` 1)
- 이 SPEC의 `spec.md` 단독 린트: `[]` (finding 0건)
- 템플릿 미러 중립성 grep baseline: **9건**, 줄 `68 69 104 134 135 149 150 151 202`
  (전부 교육용 예시 식별자 — M2가 도입하는 것이 아님)

### 강등 경로 실측 (iter1 FAIL의 근원)

유산 분류 디렉터리 사본에 강등 대상 코드와 비대상 코드를 동시에 심고 한 번의
린트로 관측:

```
{"code":"CoverageIncomplete","severity":"error","advisory":null}      # 비대상 → 살아남음
{"code":"FrontmatterInvalid","severity":"warning","advisory":true}    # 대상  → 강등됨
```

전용 코드 방식이 작동한다는 직접 증거. 상세는 plan.md §A.5.

### 상태

- 미해결 결정: 없음. `[NEEDS CLARIFICATION]` 마커 0건.
- 열린 위험: plan.md §G 5건. 최상위는 "M1 단독 랜딩 시 error 9건 발생"
  (전용 코드 채택으로 v0.1.0의 1건에서 확대 — M3 선행 또는 동시 랜딩 필수).
- plan-audit iter1: FAIL 0.75 → 본 v0.2.0에서 D1~D10 반영.

## §E.2 Run-phase Evidence

랜딩 순서는 M3 → M1 → M2 → M4다. M3을 먼저 둔 이유는 §A.6에 있다 — M1이 도입하는
코드는 강등되지 않는 error이므로, 오염된 9건을 먼저 교정하지 않으면 M1 단독 랜딩이
저장소 린트를 error 9건으로 깨뜨린다.

### M3 — 오염 코퍼스 교정 (14건)

- 교정 대상 14건 = `spec.md` 9건 + in-scope 형제 산출물 5건. 범위는 acceptance.md
  §D의 판정 명령으로 직접 도출했고 plan.md §B.2와 일치했다.
- 값 분포: 제거된 12건이 `plan`, 1건이 `run`(`SPEC-ENVKEY-ANTHROPIC-SSOT-001/progress.md`),
  1건이 `sync`(`SPEC-REF-SEO-ABSORB-001/spec.md`) → 14건 모두 `"v3.0.2"`. `git diff -U0`
  의 +/- 본문 줄은 전부 `phase:` 한 줄이었고 그 외 줄 변경은 0건이다.
- 잔여: `spec.md` 정확 일치 **0건**(교정 전 9건), 전체 산출물 **정확히 17건**(교정 전 31건).
  17건은 모두 허용된 레거시 접두사(`SPEC-V3R2-*` 11 / `V3R3-BRAIN-001` / `V3R6-LINK-FIX-001`
  / `GLM-MCP-001` / `CI-MULTI-LLM-001` / `TOKEN-001`) 소속이다.
- 타깃 값 `v3.0.2`는 가정이 아니라 git에서 재도출했다. 태그 `v3.0.1`은
  `2026-07-24 01:37:05 +0900`이고, 아홉 SPEC의 최초 커밋은 `2026-07-27 19:50`부터
  `2026-08-02 04:30` 사이로 전부 태그 이후다.
- 부분 문자열 오탐 8건(`Runtime` 안의 `Run`)은 그대로 유지됐다 — 교정 전후 모두 8.

### M1 — 값-형태 검증 도입

- `internal/spec/lint.go` **+37 / −0**. 기존 줄 수정 없이 두 가지만 추가했다:
  패키지 수준 `phaseWorkflowStageTokens` 맵(`plan`/`run`/`sync`/`mx`)과
  `FrontmatterSchemaRule.Check`의 값-형태 분기(`FrontmatterPhaseInvalid`, `SeverityError`).
- 분기는 필수 필드 공백 검사 **이후**에 배치했다. 빈 값은 기존 필수 필드 finding만 내고
  중복 finding을 만들지 않는다.
- 핵심 설계 결정은 **하지 않은 일**이다: `FrontmatterPhaseInvalid`를 `eraDemotableCodes`에
  등록하지 않았다. 미등록 코드는 `applyEraDemotion`의 두 case를 모두 통과해 유산 SPEC에서도
  error로 살아남는다. 저작 시점에 대부분의 진행 중 SPEC이 유산으로 분류되므로, 등록했다면
  가드가 보호하려던 바로 그 대상에서 advisory warning으로 강등됐을 것이다.
- 술어를 정확 일치 denylist로 택한 근거는 563개 코퍼스 실측이다: 엄격 semver 허용목록은
  오탐 301건, 부분 문자열 포함은 8건, 정확 일치 denylist는 **0건**.
- RED 증거(구현 전 실행):

  ```
  --- FAIL: TestPhaseValueShape_WorkflowStageTokenRejected/plan
      lint_phase_test.go:64: phase "plan": expected exactly 1 FrontmatterPhaseInvalid finding, got 0: []
  ```

  **정직한 단서**: 4개 테스트 중 진짜 RED는 1개뿐이었다. 나머지 3개(정당한 값 통과 /
  빈 값 중복 없음 / 강등 집합 미등록)는 규칙이 없는 상태에서 공허하게 통과한다 —
  회귀 가드이지 RED 증거가 아니다.

### M2 — 저작 지시 추가

- 로컬 `.claude/agents/moai/manager-spec.md` 147행 + 템플릿 미러 동일 위치에 문단 1개씩
  추가(각 +2). 삽입 위치는 `#### [HARD] SPEC Frontmatter Canonical Schema` 절의 SSOT
  상호참조 문단 바로 뒤다. 스키마를 재정의하지 않고 SSOT를 가리킨다.
- 두 사본의 삽입 문단은 byte-identical이다. 내부 SPEC ID·REQ 토큰·날짜·SHA·내부 경로를
  쓰지 않았기 때문에 미러 중립화가 따로 필요 없었다.
- 중립성 델타 **0**: 패턴 매치 건수 9 → 9, 매치된 내용 토큰 multiset이 `diff`로 동일.
  줄 번호만 삽입분만큼 밀렸다(`149 150 151 202` → `151 152 153 204`).
- **catalog.yaml 해시 재생성 cascade**: `manager-spec.md` 편집이 `internal/template/catalog.yaml`
  의 SHA256을 무효화해 `TestManifestHashFormat`이 FAIL했다. `gen-catalog-hashes.go --all`로
  1줄 재생성해 해소. 반직관적인 점 — 이 해시의 source는 템플릿 미러가 아니라 **로컬
  `.claude/` 사본**이다(실패 메시지의 `source=.claude/agents/moai/manager-spec.md`).
  즉 로컬만 고쳐도 해시는 깨진다.
- 이 FAIL은 AC-PFV-013의 판정 명령(`-run 'Leak|Neutral'`)이 잡지 못한다. 전체 패키지를
  돌려서 발견했다 — 선택자 통과가 패키지 green의 증거가 아니라는 사례다.

### M4 — 회귀 가드 (반증 가능성)

- 픽스처 2건 추가: 유산 era 갈래와 terminal 상태 갈래 각각에서 부정 토큰이 error로
  살아남는지 확인한다. 두 테스트 모두 **같은 리포트 안에서** 강등 대상 통제군
  (`MissingExclusions`)이 warning/advisory로 내려갔는지를 먼저 단언한다 — 통제군이
  강등되지 않으면 그 픽스처는 애초에 강등 우회를 시험할 수 없으므로 실패시킨다.
- plan.md §F M4의 나머지 두 항목(정당한 유산 값 4종 / 빈 값 중복 없음)은 M1이 이미
  커버하고 있어 중복 작성하지 않았다.
- 왕복 (a) 호출부 제거 — FAIL 원문:

  ```
  --- FAIL: TestPhaseValueShape_WorkflowStageTokenRejected/plan
      lint_phase_test.go:80: phase "plan": expected exactly 1 FrontmatterPhaseInvalid finding, got 0: []
  --- FAIL: TestPhaseValueShape_TerminalStatusBypassesDemotion
      lint_phase_test.go:230: terminal status: expected exactly 1 FrontmatterPhaseInvalid finding, got 0: []
  --- FAIL: TestPhaseValueShape_GrandfatheredEraBypassesDemotion
      lint_phase_test.go:222: grandfather era: expected exactly 1 FrontmatterPhaseInvalid finding, got 0: []
  FAIL	github.com/modu-ai/moai-adk/internal/spec	0.478s
  ```

- 왕복 (b) 술어 본문 무력화(호출부 유지, 토큰 맵을 비움) — FAIL 원문:

  ```
  --- FAIL: TestPhaseValueShape_WorkflowStageTokenRejected/plan
      lint_phase_test.go:80: phase "plan": expected exactly 1 FrontmatterPhaseInvalid finding, got 0: []
  --- FAIL: TestPhaseValueShape_TerminalStatusBypassesDemotion
      lint_phase_test.go:230: terminal status: expected exactly 1 FrontmatterPhaseInvalid finding, got 0: []
  --- FAIL: TestPhaseValueShape_GrandfatheredEraBypassesDemotion
      lint_phase_test.go:222: grandfather era: expected exactly 1 FrontmatterPhaseInvalid finding, got 0: []
  FAIL	github.com/modu-ai/moai-adk/internal/spec	0.454s
  ```

  (b)가 하중을 증명한다. 호출부가 그대로 실행되는데도 실패하므로, 테스트가 붙잡고 있는
  것은 도달 가능성이 아니라 술어의 **내용**이다.
- 복원 증명: 두 왕복 각각 이후 `internal/spec/lint.go`의 md5가
  `289521f4f58bae8feb75ede14b09a184`로 M1 상태와 일치했고, `git diff --stat`이 +37로
  돌아왔으며, 선택자 테스트가 다시 `ok`였다.
- **설계상 관측 2건**:
  1. 두 왕복의 FAIL 출력이 byte-identical(둘 다 1776 bytes)이다. 출력만으로는 (a)와 (b)를
     구별할 수 없고, (b) 직전에 호출부가 살아 있음을 grep으로 확인한 기록만이 두 왕복을
     구분한다. 이 grep이 없으면 (a)를 두 번 돌리고 둘 다 했다고 주장할 수 있다.
  2. 부정 단언 테스트(정당한 값 0건 / 빈 값 중복 없음)는 두 왕복을 **모두 통과**한다 —
     술어를 없애도 "finding이 없음"은 여전히 참이기 때문이다. 하중을 지는 것은 긍정
     단언 3건뿐이다. 부정 단언만으로 구성된 스위트였다면 두 왕복을 모두 통과하고
     아무것도 증명하지 못했을 것이다.

### 선택자 결함 2건 (M4에서 문서화)

판정 명령 자체의 결함 2건을 발견해 `internal/spec/lint_phase_test.go` 상단 주석에
기록했다. `acceptance.md` 본문 수정은 본 에이전트의 소유 경계를 벗어나므로 하지 않았다.

1. `go test -run '^X' -list '.*'`는 `-run` 선택자를 무시하고 패키지 전체를 나열한다.
   AC-PFV-014의 `-run` 인자는 무효이고, 실제로 거르는 것은 뒤따르는 `grep -c`다.
2. `-run 'Leak|Neutral'`은 `TestManifestHashFormat`에 도달하지 않는다. AC-PFV-013의
   세 번째 명령은 중립성은 보지만 템플릿 편집 무결성은 보지 못한다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-02
run_commit_sha: d320795bc
run_status: audit-ready

ac_pass_count: 16
ac_fail_count: 0
ac_matrix: "AC-PFV-001~016 전부 PASS. 단 2건은 판정력이 약하다 — 아래 ac_weak 참조"
ac_weak:
  AC-PFV-006: "구조적 공허. 통과 조건이 '새 error 파일 없음'인데 before/after error 파일
    목록이 둘 다 빈 목록이라 comm -13이 무조건 무출력이다. 저장소 error가 0인 한 실패할 수
    없다 — 즉 이 SPEC이 유지하려는 바로 그 상태에서만 판정력을 잃는다"
  AC-PFV-014: "-list가 -run을 덮으므로 -run 인자가 무효. grep -c가 실제 필터라 AC는
    작동하지만, 명령의 형태를 선택자 검증으로 읽으면 오독이다"

verification:
  spec_lint: "moai spec lint --json → exit 0, 총 62건, error 0건,
    MissingExclusions 24 / StatusGitConsistency 16 / FrontmatterInvalid 14 /
    LegacyEARSKeyword 7 / OwnershipTransitionInvalid 1 (M3 이전 baseline과 코드별 건수 동일)"
  go_test_all: "go test ./... -count=1 → exit 0, ok 105 패키지, no test files 3, FAIL 0줄"
  go_test_spec: "go test ./internal/spec/... -count=1 → ok"
  golangci_lint: "golangci-lint run ./internal/spec/... ./internal/template/... → 0 issues"
  go_build: "go build ./... → exit 0"
  go_vet: "go vet ./... (저장소 전체) → exit 0"
  gofmt: "gofmt -l internal/spec/lint.go internal/spec/lint_phase_test.go → 무출력"
  template_guard: "go test ./internal/template/... -count=1 → ok (catalog 해시 재생성 후)"
  planted_violation: "유산 분류 사본에 부정 토큰 + 강등 대상 위반을 함께 심고 1회 린트 →
    FrontmatterInvalid=warning/advisory:true, FrontmatterPhaseInvalid=error/advisory:null, exit 1"

evidence_dir: .moai/state/verify/pfv/
new_warnings_or_lints_introduced: 0
total_run_phase_files: 18
  # 14(M3 phase 교정) + lint.go + lint_phase_test.go(신규) + manager-spec.md 2사본
  # + catalog.yaml — d320795bc는 이 SPEC 자신의 plan-phase 산출물 4건도 함께 담는다
m1_to_mN_commit_strategy: "단일 커밋(d320795bc)에 M3+M1+M2+M4 일괄 랜딩. M1 단독 랜딩이
  저장소를 깨뜨리므로(§A.6) 분할 커밋이 성립하지 않는다"

carried_debt:
  - "AC-PFV-006은 위 ac_weak대로 구조적 공허 — 통과를 판정력의 증거로 읽지 말 것"
  - "린터의 SPEC 발견 함수는 SPEC-*/spec.md만 수집한다. 교정한 형제 5건과 남긴 레거시
    17건은 기계적 강제 범위 밖이고, 새 plan.md/progress.md가 부정 토큰으로 저작되는 것을
    막는 것은 M2의 지시뿐이다 — 지시는 강제가 아니다"
  - ".github/workflows/template-neutrality-check.yaml은 실행하지 못했다(러너 필요).
    Go 측 가드는 전체 패키지로 통과했으나 워크플로 자체의 판정은 미관측이다"
  - "커버리지 미측정(go test -cover 미실행), -race 미실행. 변경이 동기 맵 조회뿐이라
    레이스 가능성은 낮으나 관측하지 않았다"
  - "선택자 결함 2건은 문서화만 했다. acceptance.md 본문의 판정 명령은 그대로이므로,
    AC-PFV-013의 세 번째 명령을 단독 실행하면 여전히 red 패키지에서 PASS를 낸다.
    수정은 manager-spec 재위임 사안"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-02
sync_commit_sha: 460c529b2
sync_status: audit-ready

b12_self_test_a: "grep -c 'SPEC-PHASE-FIELD-VALIDATION-001' CHANGELOG.md → 0 (방출 전 실행, 사전 중복 없음)"
b12_self_test_b: "AC 개수: grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l → 16 (AC-PFV-001..016, 0이 아님을 직접 확인). CHANGELOG 항목은 16/16 판정으로 서술"
b12_self_test_c: "CHANGELOG 항목이 지목한 파일 5개 전부 ls 로 실재 확인 (internal/spec/lint.go / internal/spec/lint_phase_test.go / .claude/agents/moai/manager-spec.md / internal/template/templates/.claude/agents/moai/manager-spec.md / internal/template/catalog.yaml)"

changelog_entry_position: "[Unreleased] > ### Added 의 첫 항목 (SPEC-REF-SEO-ABSORB-001 항목 바로 앞)"

frontmatter_status_transitions:
  spec.md: "in-progress → implemented → completed (단일 sync 커밋 병합 전이; updated: 2026-08-02 — 이미 해당 날짜라 값 변화 없음)"
  plan.md: "in-progress → implemented → completed (동일)"
  acceptance.md: "in-progress → implemented → completed (동일)"
  progress.md: "in-progress → implemented → completed (동일)"

canary_compliance_check:
  applicable: true
  reason: "이 SPEC은 전방위 정책을 정의한다 — 모든 SPEC 산출물의 `phase:` 값은 릴리스 대상 버전 문자열이어야 하며 워크플로 단계 토큰이 아니다. 그 정책을 자기 자신에게 적용한 결과가 아래 self-check 이다."
  self_check: "이 SPEC 4개 산출물의 phase 값 = \"v3.0.2\" (4/4). 새 가드가 자기 자신을 통과한다."
  observed: "moai spec lint --json 에서 FrontmatterPhaseInvalid 0건"

user_facing_surface_judgment:
  readme: "변경 없음"
  docs_site: "변경 없음"
  reason: "변경은 SPEC 저작 파이프라인 내부에 한정된다 — 린트 finding 코드 하나와 저작 에이전트 지시문. 사용자가 쓰는 CLI 플래그·출력·설정 키가 하나도 바뀌지 않았다. 다만 `moai spec lint` 를 돌리는 사용자에게는 새 error 코드가 보일 수 있으며, 그 사실은 위 CHANGELOG 항목이 담당한다."

residual_risk:
  - "AC-PFV-006 은 저장소가 error 0건 상태인 동안 실패할 수 없다 — 형제 AC 보다 약한 검사다 (run-phase 에서 이미 부채로 기록)."
  - "판정은 정확 일치 denylist 다. `plan` / `run` / `sync` 이외의 새로운 오염 토큰이 등장하면 잡지 못한다. 허용목록으로 뒤집으면 301건 오탐이 되므로 의도적 선택이다."
```

**sync 시점 회귀 관측**

| 검사 | 명령 | 관측 |
|---|---|---|
| 저장소 spec lint | `./bin/moai spec lint --json` | exit 0, 62건 / error 0건 (`MissingExclusions` 24 / `StatusGitConsistency` 16 / `FrontmatterInvalid` 14 / `LegacyEARSKeyword` 7 / `OwnershipTransitionInvalid` 1) — run-phase 기준선과 동일 |
| 4개 산출물 status | `grep -n '^status:' {spec,plan,acceptance,progress}.md` | 4/4 `completed` |

**미검증 (Gaps)**

- `§E.2 Run-phase Evidence` / `§E.3 Run-phase Audit-Ready Signal` 은 `_<pending run-phase>_` 로 남아 있다. 두 절의 소유자는 manager-develop 이며 sync-phase 소유 범위 밖이라 채우지 않았다. run-phase 증거는 커밋 `d320795bc` 의 메시지 본문에 있다.
- `sync_commit_sha` 는 sync 커밋 시점에 자기 해시를 참조할 수 없어 placeholder 로 기록한 뒤 이 후속 커밋에서 백필했다. 백필 값 `460c529b2` 는 `git rev-parse --short 460c529b2` 로 해소를 확인하고 `git show --stat` 으로 대상 파일 5개(CHANGELOG.md + 산출물 4개)를 관측한 뒤 기록했다.
