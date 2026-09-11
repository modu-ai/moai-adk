---
id: SPEC-SPECLINT-GATE-SIGNAL-001
title: "SPEC Lint 게이트 재신호화 — 상시 적색 판정(M1)과 신규 적색 분리 기제(M2)"
version: "0.1.0"
status: in-progress
created: 2026-09-07
updated: 2026-09-11
author: manager-spec
priority: P1
phase: "v3.1.5 target"
module: "internal/cli/spec_lint.go, internal/spec/lint.go, .github/workflows/spec-lint.yml"
lifecycle: spec-anchored
tags: "spec-lint, ci-gate, strict-mode, baseline-ratchet, warning-signal, re-baseline"
tier: M
dependencies: [SPEC-SPEC-LINT-BLIND-AXES-001]
---

# SPEC: SPEC Lint 게이트 재신호화 — 상시 적색 판정(M1)과 신규 적색 분리 기제(M2)

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-07 | manager-spec | 최초 작성 — 카드 t525. M1(판정) → M2((b) 수리 기제) → M3(t518 상호잠금·재기준 절차) → M4(부수: SpecsDirMissingSpecFile 2건) 순서 고정. 기준 측정 트리 `dd1439502` |

## 1. 문제 — 측정된 형태

CI 잡 `SPEC Lint`(`.github/workflows/spec-lint.yml`)는 develop push 에서
`go run ./cmd/moai spec lint --strict`(`.github/workflows/spec-lint.yml:58`)를 돌리고,
CLI 는 `report.HasErrors()`가 참이면 exit 1 을 낸다(`internal/cli/spec_lint.go:96-98`).
`--strict` 는 "treat warnings as errors" 플래그이고(`internal/cli/spec_lint.go:106`),
그 판정은 `internal/spec/lint.go:62` 의 한 줄이다:

```go
if r.Strict && f.Severity == SeverityWarning && !f.Advisory {
    return true
}
```

즉 **advisory 가 아닌 경고가 하나만 서 있어도 게이트는 무조건 적색**이다. 그리고 지금
advisory 가 아닌 경고가 수천 건 서 있다.

- CI run 34088825415(develop `d4162b368`, 2026-09-07): `0 error(s), 4344 warning(s)` →
  `Process completed with exit code 1`. **에러는 0이다.**
- `gh run list --branch develop --workflow "SPEC Lint" --limit 5` → 전부 failure.
  `d4162b368`(09-07 05:58) · `ace1c5440`(09-07 02:30) · `33fcb644b`(09-07 02:14) ·
  `615d18c1f`(09-06 05:47) · `25a3212a9`(09-03 10:50). **최소 나흘, 최소 5회 연속.**
- 작업 트리 `dd1439502`(plan 시점 origin/develop tip) 재측정:
  `moai spec lint` → `0 error(s), 4368 warning(s)`, `--strict` 없이는 rc=0.
  원문: `.moai/reports/t525/spec-lint-baseline-dd1439502.txt`.

적색의 원인이 경고뿐이라는 뜻이다. 그런데 경고 재고는 이미 진행 중이다 — 그래서 문제가
더 커진다.

### 1.1 수치는 흐른다 — 동결 금지의 근거

두 측정이 이미 다르다: CI `d4162b368` 에서 4,344, 트리 `dd1439502` 에서 4,368. 그 사이에
SPEC 이 잡혔을 뿐인데 수가 움직였다. 코퍼스 수(SPEC 디렉터 수 vs `SPEC-*/spec.md` 글롭 수)도
`SpecsDirMissingSpecFile` 2건(SPEC-V3R4-CC2X-ADOPT-001/002)만큼 어긋나며, SPEC 이 착지할
때마다 함께 +1 된다.

따라서 이 SPEC 은 **어떤 경고 수도, 어떤 코퍼스 수도 정수로 못박지 않는다.** 측정 시점의
그때-current develop 에서 재도출하는 것을 요구사항으로 강제한다(REQ-SLGS-001,
REQ-SLGS-003). 카드가 참조한 규칙별 분포(CoverageIncomplete 3,588 · ModalityMalformed 412 ·
MovingRefUnpinned 117 · StatusTransitionInvalid 102 · MissingExclusions 25 ·
FrontmatterInvalid 14 · InvalidREQID 6)는 트리 `0b1e27877` 시점의 **역사적 표본**이며,
M1 이 재도출하기 전까지는 아무 판정의 근거로 쓰지 않는다.

### 1.2 왜 "상시 적색"이 게이트를 죽이는가

게이트의 가치는 **변화에 반응하는 것**이다. 상시 적색에서는:

- 아무것도 안 해도 빨간 것이다 — 새로 망가뜨린 사람도, 아무것도 안 한 사람도 같은 빨강을 본다.
- 경고 4,344건 중 하나라도 새로 생겼는지는 요약 줄에서 알 수 없다(총계만 바뀐다).
- 결국 사람은 적색을 무시하고 지나가고, 게이트는 신호가 아니라 배경이 된다.

t518(SPEC-SPEC-LINT-BLIND-AXES-001, 아직 develop 미착지)은 이 경고 인구의 **판정 기준
자체**를 고치는 카드다. 그 M1 이 `REQEntry.Widened` 선례(`internal/spec/lint.go:712-717`,
t385)로 권고 확대(advisory widening)에 들어가면 — Widened 인 REQ 의 finding 은
`(SeverityWarning, Advisory: true)` 가 된다 — 경고 총수와 advisory/비-advisory 분포가
**통째로** 움직인다. 오늘 수치로 부채를 갚거나 임계를 못박으면 그 후에 전부 다시 세야 한다.

### 1.3 인과 주장 하나는 철회됐다

카드 초기판은 "CI 4,344"와 "t518 의 4,344"가 같은 수여서 인과로 읽혔으나, lane-5 실측으로
**철회됐다** — 서로 다른 단위를 다른 커밋에서 잰 세 측정이 우연히 같은 수를 공유한 것.
t518 먼저라는 **순서 판정은 유지**되는데 사유가 바뀌었다: §1.2 의 인구 이동이기 때문이다.
이 SPEC 은 그 순서를 요구사항으로 못박는다(REQ-SLGS-011).

### 1.4 판정 축 — (a) / (b) / (c)

카드가 세운 세 갈래 판정:

- **(a) 경고가 진짜 부채이고 게이트가 옳다** → 부채를 갚는 카드가 필요하다.
- **(b) 경고로 exit 1 을 내는 정책이 잘못됐다** → 임계 또는 기준선 분리가 필요하다.
- **(c) 경고 자체가 오탐이다** → t518 의 소관이다.

둘은 서로 배타적이지 않다. plan 시점의 구조 관측은 (b)를 가리킨다 — `--strict` +
비-advisory 경고 재고 ≥ 1 이면 영구 적색은 수학적 귀결이다. 그러나 이 SPEC 은 판정을
가정으로 두지 않고 **증거로 게이트된 마일스톤(M1)**으로 만든다. (c)는 t518 의 축이고,
(a)는 진짜지만 t518 착지 전까지 실행하지 않는다(REQ-SLGS-011).

## 2. 요구사항 (GEARS)

### M1 — 판정: 무엇이 적색을 만드는가

- **REQ-SLGS-001** (Ubiquitous): run 단계는 측정 시점의 그때-current `origin/develop`
  트리에서 SPEC lint 발견 분포를 **규칙별 · severity 별 · advisory 마킹별**로 재도출하고,
  명령 전문과 트리 SHA 를 증거로 `.moai/reports/t525/` 에 기록한다. 다른 트리·다른 날짜에서
  잰 수를 새 측정으로 귀속하는 것은 금지다.
- **REQ-SLGS-002** (Ubiquitous): run 단계는 (a)/(b)/(c) 각각에 대해 "성립/불성립/소관
  외" 판정을 내리고, 각 판정에 명령 + 관측 출력을 대응시켜 기록한다. 판정 없는 기제 구현은
  이 SPEC 의 성공이 아니다.
- **REQ-SLGS-003** (unwanted): 이 SPEC 은 경고 수, 코퍼스 수, 규칙별 분포를 교리·코드
  상수·AC 임계값에 **정수로 못박지 않는다.** 모든 수는 측정 시점 재도출 값이며, 문서가
  참조할 때는 "어느 트리에서 잰 값"까지 함께 쓴다.

### M2 — 수리 기제: 서 있는 재고는 초록으로, 새 적색은 빨강으로

- **REQ-SLGS-004** (state-driven): While 기준선 파일이 존재하면, the gate shall exit 1
  only on **error**-severity findings or a per-rule non-advisory increase over the
  baseline. 기준선 이하의 서 있는 재고는 통과하며 경고 총수는 출력에 그대로 보인다.
- **REQ-SLGS-005** (Ubiquitous): 기준선의 운반체는 **CLI 플래그 + 저장소에 체크인된
  파일**이다(예: `--baseline <path>`, 플래그 미지정 시 오늘의 동작과 동일). 새
  `.moai/config/sections/*.yaml` 을 만들지 않는다(§2.1).
- **REQ-SLGS-006** (event-driven): When 기준선 초과(비-advisory 경고의 규칙별 증가)가 감지되면 the gate shall
  exit 1 with the per-rule delta — **어느 규칙이 얼마나 컸는지**를 출력한다. 총계만 바뀐
  붉음은 신호가 아니다.
- **REQ-SLGS-007** (event-driven): When 발견이 감소하면, the gate shall pass and print
  the improvement — 통과하며 감소분을 출력한다. 이때 기준선 파일을 **함축적으로 고쳐 쓰지
  않는다** — 기준선 수축은 오직 명시적 재기준 절차(REQ-SLGS-008)로만 일어난다.
- **REQ-SLGS-008** (Ubiquitous): 재기준(re-baseline)은 명시적 명령·절차로만 가능하며, the
  re-baseline command shall require a non-empty `--reason "<text>"` — 사유가 없거나
  비어있으면 실행을 거절한다. 트리 SHA · 날짜 · 사유를 기록해 git 이력에서 감사 가능해야
  한다. 기록 없는 재기준은 기준선 조작이다.
- **REQ-SLGS-009** (Ubiquitous): error severity 발견은 기준선과 **무관하게** 항상
  exit 1 을 낸다. 기준선이 진짜 에러를 가리는 일은 없다(오늘의 `HasErrors()` error 경로
  보존).
- **REQ-SLGS-010** (Ubiquitous): `SPEC Lint` 워크플로는 새 기제로 게이트를 돌리고, 통과
  run 의 로그에도 경고 총수(및 기준선 대비 상태)가 남아 서 있는 재고를 계속 관측
  가능하게 한다.

### M3 — t518 상호잠금과 재기준

- **REQ-SLGS-011** (state-driven): While t518(SPEC-SPEC-LINT-BLIND-AXES-001)이 develop 미착지인 동안 this SPEC shall
  not execute the (a)-axis warning-debt payment —
  (a)축 경고 부채 상환, 곧 대량 SPEC 문서 수정이다. 기준선의 **최초 산출**은 진행할 수
  있으나, t518 이 가져올 인구 이동을 흡수하는
  **재기준 단계는 t518 착지로 게이트된다.** t518 이 착지하면(When) 이 SPEC 의 프런트매터에
  DAG 간선 `dependencies: [SPEC-SPEC-LINT-BLIND-AXES-001]` 를 추가한다 — plan 시점에
  넣으면 코퍼스에 없어 `MissingDependency` **error**(`internal/spec/lint.go:1083`)로 스스로
  적색을 만들므로, 간선은 M3 의 산출물이다.

### M4 — 부수: SpecsDirMissingSpecFile 2건

- **REQ-SLGS-012** (event-driven): When run 단계가 M4 에 들어가면, the run phase shall
  first record why SPEC-V3R4-CC2X-ADOPT-001/002 lack spec.md — 먼저 왜 그 상태인지 답을
  기록하고(plan 관측 — 둘은 다르다: 001 은 `spec_id`·`phase: research` frontmatter 의
  2026-05-12 리서치 우산 문서, 002 는 frontmatter 없이 2026-08-23 에
  `/harness:release-update` 가 만든 우산 문서다. M4 는 각 디렉터의 '왜'를 따로 기록한다),
  그 답에 따라
  spec.md 를 보태거나 디렉터를 `.moai/specs/` 밖으로 옮겨 발견 2건을 닫는다. "왜" 없이
  발견만 지우는 것은 금지다.

### 2.1 설계 결정 ① — 운반체는 플래그 + 체크인 파일이다 (config 섹션이 아니다)

`.moai/config/sections/` 에 새 섹션 파일을 두는 대안을 검토하고 기각했다. 근거 셋:

1. **Template-First 의무**(CLAUDE.local.md §2): config 섹션 파일은
   `internal/template/templates/.moai/config/sections/` 미러 + template-neutrality(C1-C8)
   검사를 잦는다. 그런데 기준선은 **moai-adk-go 개발 저장소의 상태**지 배포 사용자
   프로젝트의 설정이 아니다 — 16개 프로그래밍 언어 사용자 프로젝트에 비어 있는 기준선
   설정을 배포하는 것은 템플릿 중립성 교리와 정면충돌이다.
2. **update 소거 위험**(CLAUDE.local.md §2.3): `.moai/config` 뿌리는 `moai update` 가
   통째 삭제 후 재배포한다. config 안의 기준선 파일은 매 update 마다 지워지거나 템플릿판에
   덮어써진다.
3. **성질**: 기준선은 설정이 아니라 **잠금파일류의 가변 상태**다. 측정 결과물이고, git
   이력으로 감사되어야 하고, 의도적으로만 고쳐 쓴다. 설정 파일의 자리가 아니다.

파일 위치 후보는 plan.md §F M2 에서 확정한다. 구속 조건 하나: `.moai/config/` 아래는
피하고(1·2), 템플릿에 미러하지 않는다(기준선은 배포하지 않는 개발 저장소 상태).

### 2.2 설계 결정 ② — CI 에서는 `--strict` 를 `--baseline` 이 대체한다 (권고, kickoff 승인 대상)

비교한 모양 넷:

| 모양 | 내용 | 트레이드오프 |
|------|------|--------------|
| (i) `--strict` 폐기 | 적색 = error 뿐 | 경고 증가가 완전히 안 보인다 — 목적(REQ-SLGS-006) 위반. 기각 |
| (ii) 총량 임계 `--max-warnings N` | 간단 | 규칙 해상도 상실 — 한 규칙의 증가가 다른 규칙의 감소 뒤에 숨고, N 은 동결 금지(REQ-SLGS-003)와 충돌한다. 기각 |
| (iii) 규칙별 기준선 래칫 | 체크인 기준선 대비 증가분만 적색 | t518 인구 이동을 게이트된 재기준 하나로 흡수한다. **권고** |
| (iv) 하이브리드 | (iii) + `--strict` 병기 | escalation 정책이 두 플래그에 갈라져 의미가 중복된다. 기각(단순성) |

권고 구체형: CI 는 `go run ./cmd/moai spec lint --baseline .moai/<기준선 파일>` 로
바뀌고, `--strict` 는 그대로 남아 다른 용도(로컬 일회성 엄격 판정)에 쓰인다. 기존 사용자
동작 불변 — 새 플래그는 옵션이고 기본값은 오늘의 동작. 최종 선택은 kickoff 승인에서
운영자가 (i)-(iv) 중 정하며, 이 SPEC 은 (iii)을 권고한다.

### 2.3 설계 결정 ③ — 재기준은 명명된 명령이다

재기준은 "기준선 파일을 손으로 고친다"가 아니라 도구가 대행한다: 현재 트리에서 재측정한
분포로 기준선을 다시 쓰되, 커밋 메시지/기록에 트리 SHA · 사유를 남긴다. t518 착지 직후의
재기준(REQ-SLGS-011)이 이 절차의 첫 고객이다 — 권고 확대로 비-advisory 재고가 통째로
움직여도, 게이트된 재기준 한 번으로 흡수된다.

## 3. 인수 기준

`acceptance.md` 를 본다. AC-SLGS-001..012 가 REQ-SLGS-001..012 를 전부 덮으며, 각 AC 는
명령 + 기대 관측치 + 실패 모양을 갖는다. 핵심 판정은 AC-SLGS-005(합성 경고 주입 — 새
적색이 붉어지는지), AC-SLGS-002((b) 구조 입증의 mutation), AC-SLGS-006(서 있는 재고가
초록인지)이다.

## 4. 잔여 위험

- **기준선 조작**: 재기준이 쉬우면 무의미해진다. REQ-SLGS-008 의 기록 의무와 plan-audit /
  sync-audit 의 관심으로 완화한다. 재기준 커밋은 사유를 반드시 명명한다.
- **advisory 분류의 드리프트**: t518 이 advisory 경계를 넓히면 기준선의 비-advisory 키가
  통째로 움직인다 — 이것이 M3 재기준 게이트의 존재 이유다. 게이트를 잊으면 기준선은
  깨진 초록이 된다.
- **환경 의존 규칙**: `StatusGitConsistency` 류는 환경에 따라 수가 흔들린다. 기준선 키는
  advisory 마킹을 기준으로 삼아야 하며, 기제 구현 시 이 가정을 M1 인구통계로 검증한다.
- **(a) 부채의 지연**: 이 SPEC 은 부채를 안 갚는다. CoverageIncomplete 류의 재고는 그대로
  남고, 그 책임은 t518 이후의 별도 카드에 있다. 게이트가 신호를 내게 하는 것과 재고가
  줄는 것은 별개다 — 그리고 이 SPEC 의 성공 판정은 전자뿐이다.
- **프런트매터 필드명 분기**: 스키마 문서는 `depends_on` 을 싣지만 코드 바인딩은
  `dependencies`(`internal/spec/lint.go:500`)다. `depends_on` 로 쓰면 디코더가 조용히
  버린다. M3 의 간선 추가는 `dependencies:` 로 한다.

## 5. Scope and Out of Scope

### 5.1 In Scope

- SPEC lint 발견 분포의 재도출과 (a)/(b)/(c) 판정 기록(M1)
- 기준선 비교 기제: CLI 플래그, 체크인 기준선 파일, 증가분 적색, 감소 통과, 감사 가능한
  재기준(M2)
- `SPEC Lint` 워크플로 배선 변경(M3)
- SPEC-V3R4-CC2X-ADOPT-001/002 의 사유 규명과 발견 2건 종결(M4)
- t518 상호잠금과 게이트된 재기준 절차(M3)

### 5.2 Out of Scope

아래 각 주제는 이 SPEC 이 하지 않는 것이다.

### Out of Scope — 경고 부채 상환

- CoverageIncomplete / ModalityMalformed 등 서 있는 경고 재고의 감축 — (a)축이며
  t518 착지 후 별도 카드 소관이다. 이 SPEC 의 성공은 재고 감축이 아니라 **새 적색의
  식별**이다.
- 개별 SPEC 문서의 내용 수리.

### Out of Scope — 오탐 판정 기준 수리

- advisory 마킹 경계의 재설계, era demotion 정책 변경 — (c)축이며
  SPEC-SPEC-LINT-BLIND-AXES-001(t518)의 소관이다.

### Out of Scope — 배포 사용자 프로젝트 대상 변경

- `--strict`, `moai spec lint` 기본 동작의 파괴적 변경 — 새 기제는 옵션 플래그다.
- 템플릿 배포물에 기준선 파일이나 config 섹션을 싣는 것.

### Out of Scope — 구현 세부

- 기준선 파일 포맷(JSON 키 정렬 등), 내부 함수·타입 명명, 규칙 클래스 추상화 — run
  단계 설계 소관이다.
