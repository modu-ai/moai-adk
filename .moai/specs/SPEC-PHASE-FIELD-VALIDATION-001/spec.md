---
id: SPEC-PHASE-FIELD-VALIDATION-001
title: "phase 프론트매터 필드의 값-형태 검증과 오염 코퍼스 교정"
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

# SPEC-PHASE-FIELD-VALIDATION-001 — phase 필드 값-형태 검증

## HISTORY

### v0.2.0 (2026-08-02)

- plan-audit iter1 FAIL(0.75) 반영. 핵심 정정은 **강등 술어를 두 갈래 중 한 갈래만
  읽었던 것**이다. v0.1.0은 terminal-status 갈래만 검토하고 era 갈래를 보지 않았다.
  era 갈래는 `sync_commit_sha`가 없는 SPEC을 유산으로 분류하므로 **저작 시점의 SPEC은
  기본적으로 여기 걸린다**. §1.4 전면 재작성.
- 그 결과 v0.1.0의 "기존 `FrontmatterInvalid` 코드 재사용" 결정을 철회하고 전용 코드
  `FrontmatterPhaseInvalid`를 도입한다. 철회 근거는 실측이며 plan.md §A.5에 있다.
- `tier: M` 명시. REQ 19→16, AC 17→16으로 **통합**(삭제 아님).
- §1.2의 "18개 SPEC"을 실측값 **20개**로 정정. 파일 수 31/14/17과 레거시 11은 정확했다.
- 코퍼스 모집단을 "이 SPEC 포함 564개"로 통일(v0.1.0은 563 표기).

### v0.1.0 (2026-08-02)

- 최초 저작. 문제는 "phase 필드에 릴리스 타깃 대신 워크플로 단계 토큰이 들어간다"는
  단일 드리프트이며, 스키마가 이미 값의 의미를 정의하므로 새 값 체계는 필요 없다.
- 저작 중 실측으로 브리프의 두 전제를 정정: 오염 파일은 11개가 아니라 **31개**이고,
  린터는 `spec.md`만 읽으므로 형제 산출물은 **린트 불가시** 영역이다.

---

## 1. 배경 (Context)

### 1.1 스키마는 이미 값의 의미를 정의하고 있다

`.claude/rules/moai/development/spec-frontmatter-schema.md`는 `phase`를 두 곳에서
정의한다. 템플릿 블록은 `phase: "vX.Y.Z target"`을 제시하고, 필드 참조 표는
`phase | string | non-empty, typically release target | e.g. "v3.0.0"`으로 적는다.
즉 **릴리스 타깃 문자열**이 이 필드의 정의된 의미다.

전체 564개 `spec.md`(이 SPEC 포함)의 값 분포를 실측하면 이 정의가 사실상 지켜지고
있다. 555개가 버전 유사 문자열(`"v3.0.0"` 217개, `"v2.x - Legacy"` 95개,
`"v3.1.0 target"` 17개 등 40여 종)을 담고 있고, 워크플로 단계 토큰을 담은 것은 9개다.

### 1.2 오염의 실제 규모 — 31개 파일 / 20개 SPEC

`.moai/specs/` 전체 마크다운 산출물을 훑으면 워크플로 단계 토큰을 담은 파일은
**31개**이고, 이들이 걸쳐 있는 SPEC 디렉터리는 **20개**다.

| 구분 | SPEC 수 | 파일 수 | 린트 가시성 |
|---|---:|---:|---|
| `spec.md`가 오염된 SPEC (v3.0.1 이후 착수) | 9 | 9 | 가시 |
| 위 9개 중 2개 SPEC의 형제 산출물 | (위에 포함) | 5 | 불가시 |
| `spec.md`는 깨끗하고 형제 산출물만 오염된 레거시 SPEC | 11 | 17 | 불가시 |
| 합계 | **20** | **31** | — |

린트 불가시성은 추정이 아니라 코드 사실이다. 린터의 SPEC 발견 함수는
`SPEC-*/spec.md` 패턴만 수집하므로 `plan.md`·`acceptance.md`·`progress.md`·
`research.md`의 프론트매터는 어떤 규칙에도 도달하지 않는다. 따라서 형제 산출물의
교정은 린트를 통과시키기 위한 조치가 아니라 저작 일관성 조치이며, 동시에 형제
산출물에서는 가드가 **회귀할 수도 없다**.

레거시 11개 SPEC은 `spec.md` 자체는 깨끗하고 형제 산출물만 오염돼 있다. 오염이
"SPEC 단위"가 아니라 "작성 순간 단위"로 발생했음을 보여준다.

### 1.3 검증 게이트가 이 드리프트를 보지 못한다

프론트매터 스키마 규칙은 12개 필수 필드 각각에 대해 공백 여부만 확인한다.
비어 있지 않으면 통과하며 값의 형태는 어떤 규칙도 검사하지 않는다. 저장소 전체
린트 기준선은 62건(모두 warning), error 0건, 종료코드 0이다. 도구가 이 드리프트를
전혀 보지 못한다는 뜻이다.

### 1.4 강등 술어는 두 갈래이고, era 갈래가 지배적이다

이 절은 v0.1.0에서 가장 크게 틀렸던 부분이다.

린터에는 구조 규칙 ERROR를 advisory warning으로 내리는 강등 경로가 있다. 적용
조건은 **두 갈래의 OR**다 — SPEC 디렉터리가 grandfather era로 분류되거나, 또는
frontmatter가 terminal 상태(`completed`·`superseded`·`archived`·`rejected`)를 담거나.

v0.1.0은 오른쪽(terminal) 갈래만 검토하고 왼쪽(era) 갈래를 보지 않았다. 왼쪽 갈래가
실제로는 지배적이다.

era 분류기의 H-3 규칙은 `progress.md`에 `§E.2` 표식이 있고 `sync_commit_sha`가
없으면 **V3R5**로 분류하며, V3R5는 grandfather 보호 대상이다. H-2 규칙은
`§E.*` 표식이 아예 없으면 **V3R2-R4**로 분류하고 이것도 보호 대상이다.
`sync_commit_sha`는 sync 단계에서야 기록되므로, **plan 단계와 run 단계의 SPEC은
기본적으로 유산으로 분류된다** — 즉 예방이 필요한 바로 그 시점에 보호막이 켜진다.
명시적 `era:` 오버라이드를 단 SPEC만 이 기본값을 벗어난다.

이 SPEC 자신도 예외가 아니다. 자신의 `progress.md`가 `§E.2`를 담고
`sync_commit_sha`가 없으므로 H-3에 걸려 V3R5로 분류된다.

귀결은 두 가지다.

- **기존 `FrontmatterInvalid` 코드를 재사용하면 가드가 무력화된다.** 그 코드는 강등
  대상 집합에 등록돼 있어, 신규 저작 시점의 SPEC에서 error가 warning으로 내려간다.
  가드의 목적이 신규 드리프트 예방인데 예방 지점에서만 꺼지는 셈이다.
- **따라서 전용 코드가 필요하다.** 강등 대상 집합에 등록하지 않은 코드는 두 갈래
  어느 쪽에도 걸리지 않고 error로 남는다. 실측 근거는 plan.md §A.5에 있다.

### 1.5 근원은 저작 지시의 부재다

`phase: plan|run|sync`를 주입하는 템플릿·스킬·스캐폴드는 저장소 어디에도 없다
(배포 대상 표면 전수 검색 결과 0건). 한편 plan-phase 저작 에이전트 정의에서
`phase:`가 등장하는 유일한 줄은 `progress.md`의 `§E.2`~`§E.4` 절에 관한 서술로
프론트매터 필드와 무관하다. 즉 저작자에게 이 필드의 의미를 알려주는 문장이
존재하지 않으며, "phase = 지금 어느 워크플로 단계인가"라는 추론이 자연스럽게
발생한다. 오염된 SPEC 중 하나가 `spec.md`/`plan.md`/`acceptance.md`에 `plan`을,
`progress.md`에 `run`을 적은 사실이 이 추론이 일관되게 재현됐음을 보여준다.

---

## 2. 목적 (Purpose)

정의된 의미와 실제 기입 값 사이의 간극을 닫는다. 세 층에서 동시에 닫는다.
도구가 위반을 **탐지**하게 하고(M1), 저작자가 의미를 **알게** 하고(M2), 이미 발생한
오염을 **교정**한다(M3). 그리고 가드가 실제로 하중을 받는다는 것을 반증 가능한
방식으로 고정한다(M4).

M1의 탐지는 **저작 시점에 작동해야** 의미가 있다. §1.4가 보인 대로 저작 시점의
SPEC은 기본적으로 유산으로 분류되므로, 탐지가 강등 경로를 타면 목적을 달성하지
못한다. 전용 코드는 이 요구에서 도출된 것이지 임의 선택이 아니다.

이 SPEC은 새로운 값 체계를 정의하지 않는다. §1.1이 보인 대로 값 체계는 이미
존재하며, 이 SPEC이 추가하는 것은 그 체계에 대한 기계적 집행뿐이다.

### 2.1 기각된 두 프레이밍

- **기각 — "phase의 값 체계를 정의한다"**: 값 체계는 스키마 템플릿 블록과 필드
  참조 표에 이미 존재한다(§1.1). 정의할 것이 없다.
- **기각 — "SPEC 종료 시 phase 전이 소유자를 배정한다"**: `phase`는 릴리스 타깃
  필드이지 생명주기 단계 필드가 아니다. 전이 자체가 존재하지 않으므로 소유자를
  배정할 대상이 없다. "종료 시 phase 전이 소유자 부재"라는 종전 기술은 거짓
  전제 위에 서 있다.

---

## 3. 요구사항 (GEARS)

### 3.1 값-형태 검증

**REQ-PFV-001** — The SPEC frontmatter schema rule shall reject a `phase` value
that, after trimming surrounding whitespace and case-folding, exactly equals one
of the workflow-stage tokens `plan`, `run`, `sync`, or `mx`, and shall emit the
rejection under a dedicated finding code at error severity.

**REQ-PFV-002** — The dedicated finding code shall not be registered in the
era-demotion code set, so that the rejection is never downgraded to an advisory
warning.

**REQ-PFV-003** — The rule shall decide membership by exact equality against the
denied-token set, and shall not decide it by substring containment.

**REQ-PFV-004** — The rule shall not constrain the accepted shape of any `phase`
value outside the denied-token set, and shall emit no finding for such a value.

**REQ-PFV-005** — **When** a denied token is detected on a SPEC whose directory
classifies as grandfather era or whose frontmatter carries a terminal lifecycle
status, the rule shall still emit error severity.

### 3.2 저작 지시

**REQ-PFV-006** — The plan-phase authoring agent definition shall state that the
`phase` frontmatter field denotes the release target of the work, and shall name
the workflow-stage tokens that are rejected.

**REQ-PFV-007** — The distributed template mirror of that agent definition shall
carry an equivalent instruction expressed without any internal SPEC identifier,
internal source-file path, internal working date, or commit hash.

**REQ-PFV-008** — The authoring instruction shall state that the field is not a
lifecycle-stage field and therefore carries no status transition.

### 3.3 코퍼스 교정

**REQ-PFV-009** — Every `spec.md` whose `phase` value lies in the denied-token set
shall be corrected to the release target of the work it describes.

**REQ-PFV-010** — **Where** the dedicated finding code escapes era demotion, the
correction of REQ-PFV-009 shall be treated as mandatory for every affected
`spec.md` rather than for a subset, because each one otherwise emits an
un-demoted error.

**REQ-PFV-011** — The sibling plan-phase artifacts of an in-scope contaminated SPEC
shall be corrected to the same release-target value as that SPEC's `spec.md`.

**REQ-PFV-012** — The corrected value of an in-scope SPEC shall be derived from
observed git history rather than assumed, by determining whether the SPEC's
artifacts first landed before or after the most recent release tag.

### 3.4 회귀 가드

**REQ-PFV-013** — The regression test shall fail when the call site of the
value-shape check is removed, and shall also fail when the body of the value-shape
predicate is neutralized while its call site remains.

**REQ-PFV-014** — The regression test shall assert error severity against a fixture
whose directory classifies as grandfather era, so that the escape from demotion is
pinned as intended behavior rather than left as an unobserved side effect.

**REQ-PFV-015** — The test selector used to execute the regression test shall match
at least one test function, so that a zero-match selector cannot report success.

### 3.5 비회귀

**REQ-PFV-016** — The repository-wide error-severity finding count shall be zero
after all milestones land, and no SPEC that produced no error-severity finding
before the change shall produce one after it.

---

## 4. 범위 제외 (Exclusions)

### Out of Scope — phase 값의 허용 형태 정의

허용 목록(allowlist)을 도입해 `phase` 값이 특정 형태를 따르도록 강제하는 일은 이
SPEC의 범위 밖이다. 실측상 564개 중 310개가 엄격한 `vX.Y.Z` 형태를 벗어나며,
그중 301개는 정당한 유산 표기다. 이 SPEC은 부정 목록(denylist)만 도입한다.

- `phase` 값의 정규 형태 표준화
- 40여 종으로 흩어진 기존 표기의 일괄 통일

### Out of Scope — 레거시 SPEC 11개의 형제 산출물 17개

레거시 11개 SPEC의 형제 산출물 17개는 교정하지 않는다. 이들은 린트 불가시 영역에
있고(§1.2), 종료된 pre-v3 이력에 릴리스 타깃을 사후 배정할 근거가 없다.

- `SPEC-V3R2-*` / `SPEC-V3R3-*` 계열의 plan/acceptance/progress/research 프론트매터
- `spec-compact.md` 등 비표준 산출물의 프론트매터

### Out of Scope — 형제 산출물에 대한 린트 확장

린터가 `plan.md`·`acceptance.md`·`progress.md`까지 읽도록 발견 범위를 넓히는 일은
별개의 설계 결정이며 이 SPEC에서 다루지 않는다.

- SPEC 발견 함수의 glob 확장
- 형제 산출물용 프론트매터 스키마 정의

### Out of Scope — 강등 doctrine 자체의 변경

강등 대상 코드 집합과 terminal 상태 집합은 건드리지 않는다. 이 SPEC은 새 코드를
그 집합 **밖에** 두는 방식으로 목적을 달성하며, 기존 규칙의 강등 동작은 그대로다.

- 강등 대상 코드 집합에서 기존 코드 제거
- terminal 상태 집합 재정의
- era 분류 heuristic 수정

### Out of Scope — 인접 부채 3건

phase 드리프트와 인과 관계가 없으며, 포함하면 판정면이 흐려진다.

- CHANGELOG의 3.0.2 미출시 표기 항목
- 실제 홈 디렉터리에 의존하는 CLI 테스트 항목
- 네 번째 홈 디렉터리 판독기 항목

### Out of Scope — phase 전이 소유자 배정

§2.1에서 기각한 프레이밍. `phase`에는 전이가 존재하지 않는다.

- 상태 전이 소유권 표에 `phase` 행 추가
- 종료 시점 `phase` 갱신 훅

---

## 5. 참조

- plan.md — 마일스톤, 코드 선택의 실측 근거, 위험
- acceptance.md — 판정 명령과 저작 시점에 관측한 기준선
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — `phase` 필드 정의의 SSOT
