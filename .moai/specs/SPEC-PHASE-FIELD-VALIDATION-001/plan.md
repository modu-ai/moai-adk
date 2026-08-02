---
id: SPEC-PHASE-FIELD-VALIDATION-001
title: "phase 프론트매터 필드의 값-형태 검증과 오염 코퍼스 교정 — 구현 계획"
version: "0.2.0"
status: in-progress
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

# 구현 계획

가장 되돌리기 어려운 결정을 앞에 둔다. §A의 코드 선택과 §B의 교정 범위는 한 번
정하면 코퍼스 전체에 각인되므로 먼저 검토받아야 하고, §F의 마일스톤 순서는
기계적이다.

---

## A. 되돌리기 어려운 결정 1 — 판정 술어와 finding 코드

### A.1 부정 목록 후보의 실측 비교

저작 시점에 세 후보를 564개 `spec.md` 전수에 대해 실제로 돌렸다. 수치는 추정이
아니라 관측이다.

| 후보 | 적중 | 오탐 | 판정 |
|---|---:|---:|---|
| 엄격 허용목록 `^v[0-9]+\.[0-9]+\.[0-9]+$` | 9 | **301** | 기각 |
| 부정목록 — 부분 문자열 포함(대소문자 무시) | 9 | **8** | 기각 |
| 부정목록 — 정확 일치(트림 + 소문자화) | 9 | **0** | **채택** |

### A.2 허용목록을 기각하는 이유

`phase`는 스키마에서 "typically release target"으로 기술된다. `typically`는 형태를
강제하지 않겠다는 서술이며, 코퍼스가 그 서술을 그대로 반영한다. 실측된 40여 종의
값에는 `"Phase 4 - Multi-Language LSP"` 같은 산문, `"maintainer-tooling"`,
`"docs-site maintenance"`, `"vNext"`, `"patch-release"`, `"v2.x - Legacy"`가 섞여
있다. 엄격 허용목록은 이 중 301건을 위반으로 만든다. 9건의 진짜 드리프트를 잡기
위해 301건의 정당한 유산 표기를 위반으로 재분류하는 것은 신호 대 잡음이 1:33이며,
가드를 즉시 무시 대상으로 만든다.

### A.3 부분 문자열 대신 정확 일치를 쓰는 이유

부분 문자열 판정은 8건을 오탐한다. 전부 `Run`이 다른 단어 안에 들어간 경우다.

```
v3.0.0 — Phase 2 — Runtime Hardening        (5건)
v3.0.0 — Phase 7 — Agent Runtime Robustness (1건)
v3.0.0 R2 — Runtime Protocol Migration      (1건)
v3.0.0 R3 — Phase A — Runtime Safety Net    (1건)
```

`Runtime`이 `run`을 포함한다는 이유로 정당한 릴리스 타깃 기술을 위반으로 만드는
것은 명백한 설계 결함이다. 트림 후 소문자화한 값 전체가 부정 토큰과 **같을 때만**
위반으로 판정한다.

### A.4 부정 토큰 집합

`plan`, `run`, `sync`, `mx` 네 개.

- 앞의 셋은 3단계 생명주기의 단계명이며, 실제 오염값 전부가 이 셋에 속한다
  (`plan` 8건 + `sync` 1건).
- `mx`는 은퇴한 4번째 단계명이다. 코퍼스에 현재 0건이지만, 은퇴 전 관행을 기억하는
  저작자가 기입할 수 있는 값이므로 선제적으로 포함한다. 정확 일치이므로 포함해도
  오탐은 0건이다(실측 확인).

인용부호 처리는 불필요하다. 규칙은 YAML 디코딩 **이후**의 값을 보므로 `phase: plan`과
`phase: "plan"`은 디코더 단계에서 이미 같은 문자열로 수렴한다.

### A.5 전용 코드 `FrontmatterPhaseInvalid` — v0.1.0 결정의 철회와 그 근거

v0.1.0은 기존 `FrontmatterInvalid` 코드를 재사용하기로 했다. 그 결정을 **철회**한다.

**철회 이유.** 강등 술어는 `isGrandfatheredSpecDir(<디렉터리>) || terminalStatus[<상태>]`
두 갈래의 OR인데, v0.1.0은 오른쪽만 보았다. 왼쪽(era) 갈래는 `sync_commit_sha`가 없는
SPEC을 유산으로 분류하므로 **plan/run 단계의 SPEC이 기본적으로 걸린다**. 기존 코드는
강등 대상 집합에 등록돼 있으므로, 재사용하면 신규 저작 시점에 error가 warning으로
내려간다 — 가드가 존재 이유인 지점에서만 꺼진다.

**v0.1.0이 신규 코드를 기각했던 이유는 성립하지 않는다.** 당시 근거는 "강등 보호 밖에
있어 종료된 이력에 하드 error를 낸다"였다. 확인해 보면 남는 모집단이 없다.

- 정확 일치로 걸리는 `spec.md`는 정확히 9건이고, 9건 전부가 M3 교정 대상이다
  (§B.2). 교정 후 잔여는 **0건**이다.
- 형제 산출물 22건은 SPEC 발견 함수가 `SPEC-*/spec.md`만 수집하므로 규칙에 도달조차
  하지 않는다.

즉 소급 집행의 대상이 애초에 존재하지 않는다.

**전용 코드가 실제로 강등을 벗어나는지 실측.** 강등 함수는 `severity == error && 강등대상코드[code]`
와 `severity == warning` 두 case의 switch이므로, 강등 대상이 아닌 error 코드는 두 case
어디에도 걸리지 않고 그대로 통과한다. 코드 독해만으로 끝내지 않고 실행으로 확인했다.

이 SPEC 자신의 디렉터리(H-3 → V3R5 → 유산)를 복사해, 강등 대상 코드와 비대상 코드를
동시에 유발하는 두 위반을 심고 한 번의 린트로 관측했다.

```
$ cp -R .moai/specs/SPEC-PHASE-FIELD-VALIDATION-001 <임시>/
$ perl -0pi -e 's/^phase: "v3\.0\.2"$/phase: ""/m' <임시>/spec.md     # 강등 대상 코드 유발
$ printf '\n- REQ-PFV-001-001: The proof requirement shall exist.\n' >> <임시>/spec.md  # 비대상 코드 유발
$ moai spec lint <임시>/spec.md --json | jq -c '.[] | {code,severity,advisory}'
{"code":"CoverageIncomplete","severity":"error","advisory":null}
{"code":"FrontmatterInvalid","severity":"warning","advisory":true}
```

같은 유산 디렉터리, 같은 실행에서 강등 대상 코드는 warning으로 내려갔고 비대상 코드는
**error로 살아남았다**. 전용 코드 방식이 작동한다는 직접 증거다.

**결론.** 신규 코드 `FrontmatterPhaseInvalid`를 도입하고, 강등 대상 집합에 등록하지
않는다. 등록하지 않는 것이 이 SPEC의 핵심 설계 결정이므로 REQ-PFV-002로 명시하고
AC-PFV-003으로 기계 판정한다.

**강등 후 위반 분포.** 전용 코드에서는 9건 전부가 un-demoted error다. terminal 상태
8건과 draft 1건의 구분은 이 코드 아래에서 **판정에 영향을 주지 않는다** — 두 갈래
모두 우회하기 때문이다.

### A.6 이 결정의 즉각적 귀결 — M3이 선택이 아닌 이유

전용 코드에서는 오염된 9개 `spec.md` 전부가 강등 없이 error를 낸다. 따라서 M1만
단독으로 랜딩하면 저장소 린트가 **error 9건**으로 깨진다. v0.1.0은 이 수를 1건으로
적었는데, 그것은 기존 코드 재사용을 전제한 계산이었고 그 전제가 철회됐다.

M3의 9건 교정은 정합성 조치가 아니라 **빌드 유지 조건**이며, 이 사실이 §G의
"M3 먼저 또는 동시 랜딩" 제약을 약화시키는 게 아니라 **강화**한다.

참고로 `SPEC-UPDATE-YAML-PRESERVE-001`이 v0.1.0에서 "유일하게 강등되지 않는 파일"로
지목됐던 것은 결과적으로 맞았으나 이유가 틀렸다. 그 파일이 강등을 벗어난 실제 이유는
`status: draft`가 아니라 **명시적 `era:` 오버라이드를 달고 있어 H-override로 V3R6으로
분류**되기 때문이다. `draft`라는 상태 자체는 era 갈래를 우회시키지 못한다.
어느 쪽이든 전용 코드에서는 9건 모두가 동등하게 필수 교정 대상이므로, 이 파일에
특별한 지위는 없다.

---

## B. 되돌리기 어려운 결정 2 — M3 교정 범위

### B.1 v0.1.0 집계의 정정

v0.1.0은 "18개 SPEC / 31개 파일"로 적었다. 파일 수 31은 정확했으나 SPEC 수는
실측 **20개**다. 오염된 `spec.md`를 가진 SPEC 9개 + 형제 산출물만 오염된 레거시
SPEC 11개 = 20개다. 브리프의 "8개 SPEC / 11개 파일"도 최근 SPEC 2개의 형제 산출물만
열거한 부분 집계였다.

### B.2 교정선

| 집합 | 파일 수 | 교정 | 근거 |
|---|---:|---|---|
| `spec.md` (post-v3.0.1 SPEC 9개) | 9 | **함** | 린트 가시. 전용 코드에서 9건 전부 un-demoted error(§A.6) |
| in-scope SPEC 2개의 형제 산출물 | 5 | **함** | 같은 SPEC 내부 값 불일치 제거 |
| 레거시 SPEC 11개의 형제 산출물 | 17 | **안 함** | 린트 불가시 + 종료 이력 소급 |
| 합계 | 31 | 14건 교정 | — |

레거시 17건을 남기는 이유: 이들은 종료된 pre-v3 SPEC의 산출물이고, 린터가 읽지
않으며, 릴리스 타깃을 사후 배정할 근거 자료가 없다. 값을 지어내는 것보다 틀린 값을
그대로 두는 편이 정직하다. `spec.md`에서 §A의 술어로 차단되므로 새 드리프트는 더
생기지 않는다.

### B.3 형제 산출물 교정의 성격 명시

형제 산출물 5건의 교정은 **린트를 통과시키는 조치가 아니다**. 린터의 SPEC 발견
함수는 `SPEC-*/spec.md`만 수집한다. 따라서 이 5건에 대해서는 어떤 AC도 "린트가
탐지한다"를 주장할 수 없고, 판정은 파일 내용 직접 확인으로만 가능하다. 이 비대칭을
acceptance.md AC-PFV-009가 명시적으로 반영한다.

---

## C. 되돌리기 어려운 결정 3 — M3 타깃 값

### C.1 값 도출 방법

가정하지 않고 git 이력에서 도출했다. 최신 태그는 `v3.0.1`(2026-07-24 01:37 +0900)이며
v3.0.2는 미출시다. 각 SPEC 디렉터리의 **최초** 커밋 시각을 태그와 비교했다.

| SPEC | 최초 커밋 | 태그 대비 | 타깃 |
|---|---|---|---|
| SPEC-WORKTREE-BRANCH-GUARD-001 | 2026-07-27 19:50 | post | `"v3.0.2"` |
| SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 | 2026-07-30 11:25 | post | `"v3.0.2"` |
| SPEC-UPDATE-YAML-PRESERVE-001 | 2026-07-30 15:43 | post | `"v3.0.2"` |
| SPEC-ENVKEY-ANTHROPIC-SSOT-001 | 2026-07-30 19:57 | post | `"v3.0.2"` |
| SPEC-UPDATE-REINSTALL-LOOP-002 | 2026-07-31 06:32 | post | `"v3.0.2"` |
| SPEC-REF-SEO-ABSORB-001 | 2026-08-01 13:22 | post | `"v3.0.2"` |
| SPEC-CI-LOOP-DEVONLY-001 | 2026-08-01 19:12 | post | `"v3.0.2"` |
| SPEC-UPDATE-GUARD-EFFICACY-001 | 2026-08-01 19:43 | post | `"v3.0.2"` |
| SPEC-PIPELINE-FANOUT-ACTIVATION-001 | 2026-08-02 04:30 | post | `"v3.0.2"` |

아홉 건 모두 태그 이후에 착수됐다. 타깃은 균일하게 `"v3.0.2"`이며, 이는 기존
8개 SPEC이 이미 `phase: "v3.0.2"`를 쓰고 있는 관행과도 일치한다.

형제 산출물 5건은 각자의 `spec.md`와 같은 값을 받는다(REQ-PFV-011).

### C.2 이 도출이 틀릴 수 있는 지점

v3.0.2가 실제로 잘려 나갈 때 이 아홉 건 중 일부가 다음 마이너로 밀릴 수 있다. 그
경우 값은 사후에 틀려지지만, 이는 릴리스 계획 변경이지 이 SPEC의 결함이 아니다.
`phase`는 "타깃"이지 "출시 사실"이 아니다.

---

## D. 제약

- **템플릿 중립성**: `internal/template/templates/` 아래 미러 파일에는 내부 SPEC ID,
  내부 소스 경로, 내부 날짜, 커밋 해시를 쓸 수 없다. 실측 결과 이 미러는 byte-identical이
  아니라 **sanitized pair**다(로컬본의 내부 소스 경로와 내부 SPEC ID가 미러에서는
  일반 서술로 치환돼 있다). M2는 이 관행을 따른다.
- **강등 doctrine 불변**: 이 SPEC은 강등 대상 코드 집합도, terminal 상태 집합도,
  era heuristic도 바꾸지 않는다. 새 코드를 집합 **밖에** 두는 방식으로만 목적을
  달성한다.
- **발견 범위 불변**: SPEC 발견 함수의 glob은 건드리지 않는다(범위 제외).
- **`--strict` 상호작용**: strict 모드는 non-advisory warning을 error로 승격시키는
  경로다. 신규 finding은 이미 error severity이므로 strict 여부와 무관하게 동일하게
  집계된다. 즉 이 SPEC은 strict 동작을 바꾸지 않는다.
- **REQ 표기 관행**: 이 저장소의 최근 SPEC은 `**REQ-XXX-NNN** — text` 형태를 쓴다.
  린터의 REQ 파서는 `- REQ-ABC-001-001: text` 형태만 인식하므로 이 표기는 파싱되지
  않고, 따라서 커버리지 규칙과 modality 규칙이 적용되지 않는다. 이 SPEC도 코퍼스
  관행을 따르며, 이 사실을 숨기지 않고 여기 기록한다. (REQ↔AC 커버리지는
  acceptance.md §H의 대조표로 수동 보증한다.)
- **증거 파일 경로**: 판정 산출물은 `/tmp`가 아니라 `.moai/state/verify/`에 남긴다.
  OS가 `/tmp`를 비우면 인용한 증거 경로가 감사 시점에 해소되지 않기 때문이다.

---

## E. 자체 검증

랜딩 전 다음을 실행해 관측한다.

1. `moai spec lint --json` 종료코드 0, error-severity 0건
2. `go test ./internal/spec/ -run '^TestPhaseValueShape' -list '.*' -count=1` — 1개
   이상 매칭 확인 후 동일 선택자로 PASS 관측
3. 가드 무력화 왕복 2회(호출부 제거 / 술어 본문 무력화) 각각 FAIL 관측 후 복원
4. `go build ./...` + `go vet ./internal/spec/` 성공
5. 템플릿 중립성 가드 테스트 통과

---

## F. 마일스톤

### M1 — 값-형태 검증 도입

`internal/spec/lint.go`의 프론트매터 스키마 규칙에 부정 토큰 집합과 정확 일치
판정을 추가한다. 기존 필수 필드 공백 검사 **이후**에 배치해, 값이 비어 있는 경우와
값이 부정 토큰인 경우가 각각 하나의 finding만 내도록 한다.

- 부정 토큰 집합을 패키지 수준 변수로 선언(테스트에서 참조 가능해야 함)
- 판정: `strings.ToLower(strings.TrimSpace(fm.Phase))`가 집합에 속하는지
- finding: **신규 코드 `FrontmatterPhaseInvalid`**, `SeverityError`, 위반 값을
  메시지에 인용
- 신규 코드를 강등 대상 집합에 **등록하지 않는다** (REQ-PFV-002 — 등록 여부가
  이 SPEC의 핵심이므로 AC-PFV-003이 기계 판정)

### M2 — 저작 지시 추가

`.claude/agents/moai/manager-spec.md`의 프론트매터 스키마 절에 `phase` 필드 설명을
추가하고, `internal/template/templates/.claude/agents/moai/manager-spec.md`에 중립화된
동등 문장을 미러한다.

- 로컬본: 필드 의미 + 부정 토큰 + 생명주기 필드가 아니라는 서술
- 미러본: 같은 의미를 내부 식별자·경로·날짜 없이 일반 서술로
- 두 파일의 편집 위치는 동일 절이어야 한다(미러 구조 일관성)

### M3 — 오염 코퍼스 교정 (14건)

`spec.md` 9건 + in-scope 형제 산출물 5건을 `phase: "v3.0.2"`로 교정한다.
전용 코드에서는 9건 전부가 필수이므로 부분 교정은 성립하지 않는다(§A.6).

- 편집 단위는 프론트매터 `phase:` 한 줄
- 본문·다른 필드는 건드리지 않는다
- 레거시 17건은 손대지 않는다

### M4 — 회귀 가드

`internal/spec/`에 반증 가능한 회귀 테스트를 추가한다. 테스트 함수명은
**`TestPhaseValueShape` 접두사로 고정**한다(선택자 앵커; AC-PFV-014가 이 접두사로
매칭 수를 확인한다).

- 유산 era 픽스처(`§E.2` 있고 `sync_commit_sha` 없음)에 부정 토큰 → **error** 관측
  (REQ-PFV-014 — 강등 탈출을 의도된 동작으로 고정)
- terminal 상태 픽스처에 부정 토큰 → 역시 **error** 관측 (두 갈래 모두 우회 확인)
- 정당한 유산 값(`"v3.0.0 — Phase 2 — Runtime Hardening"` 등 §A.3의 네 값) → finding
  없음 관측 (오탐 8건이 재발하지 않음을 고정)
- 빈 값 → 기존 필수 필드 finding만, 중복 없음

M4를 마지막에 두는 이유는 M1의 술어 시그니처가 M4 테스트의 무력화 지점을 결정하기
때문이다. M1이 확정되기 전에 무력화 왕복을 설계할 수 없다.

#### M4 설계 제약 두 가지 (run-phase 관측에서 도출)

**(a) 두 왕복의 실패 출력은 서로 구별되지 않는다.** 호출부 제거와 술어 본문 무력화는
둘 다 "finding이 사라진다"는 동일한 결과를 만들므로 실패 출력이 바이트 단위로 같다
(실측: 양쪽 모두 1776 bytes). 따라서 **출력만으로는 어느 왕복을 돌렸는지 증명할 수
없다** — 같은 왕복을 두 번 돌리고 둘 다 했다고 보고하는 것이 출력상 구별되지 않는다.
구별 증거는 왕복 (b) 실행 **직전에** 호출부가 아직 남아 있음을 확인하는 grep이다.
이 grep이 (b)를 (a)와 분리하는 유일한 관측이므로, 왕복 (b)는 grep 선행이 필수다.

**(b) 부정 단언 테스트는 두 왕복을 모두 통과한다.** `_LegitimateValuesAccepted`와
`_EmptyPhaseEmitsOnlyRequiredFieldFinding`은 finding의 **부재**를 단언하므로, 술어를
무력화해도 여전히 통과한다. 하중을 받는 것은 긍정 단언 3건
(`_WorkflowStageTokenRejected`, `_GrandfatheredEraBypassesDemotion`,
`_TerminalStatusBypassesDemotion`)뿐이다. 부정 단언만으로 구성된 스위트는 두 왕복을
모두 통과하면서 아무것도 증명하지 못한다 — M4의 반증 가능성은 긍정 단언이 있을 때만
성립한다.

---

## G. 위험

| 위험 | 성격 | 완화 |
|---|---|---|
| M1 단독 랜딩 시 빌드 파손 (**error 9건**) | 확실(§A.6에서 실측 도출) | M1과 M3을 같은 PR에 넣거나 M3을 먼저 랜딩. 전용 코드 채택으로 파손 규모가 1건→9건으로 커졌으므로 이 제약은 v0.1.0보다 강해졌다 |
| 신규 코드가 예상과 달리 강등된다 | 낮음 — §A.5에서 동등 경로를 실행으로 관측 | AC-PFV-002/003이 픽스처와 집합 등록 여부를 각각 판정 |
| 형제 산출물 교정을 "린트 통과"로 오인 | 판정 오류 | acceptance.md가 해당 AC의 판정을 파일 내용 직접 확인으로 한정 |
| 레거시 17건 미교정이 지적될 가능성 | 범위 | spec.md §4에 근거와 함께 명시적 제외. AC-PFV-010이 잔여를 17로 양방향 고정 |
| v3.0.2 릴리스 계획 변경 | 외부 | §C.2에 기록. `phase`는 타깃이지 출시 사실이 아님 |

---

## H. 상호 참조

- spec.md §1.4 — 강등 술어 두 갈래와 era 갈래의 지배성
- spec.md §4 — 제외 항목과 근거
- acceptance.md §A — 판정 명령의 공허 통과 방지 규약
