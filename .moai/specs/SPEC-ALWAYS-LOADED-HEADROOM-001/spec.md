---
id: SPEC-ALWAYS-LOADED-HEADROOM-001
title: "always-loaded 지시문 표면의 허용 제거 풀 A_adm 실측과 150,000자 런타임 한도 달성 가능성 판정"
version: "0.1.0"
status: draft
created: 2026-09-27
updated: 2026-09-27
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "CLAUDE.md, AGENTS.md, .claude/rules/moai, internal/template/templates"
lifecycle: spec-anchored
tags: "always-loaded, instruction-budget, admissible-pool, measurement, verdict, binding-clause-freeze, escalation"
tier: M
related_specs: [SPEC-ALWAYS-LOADED-DIET-002, SPEC-ALWAYS-LOADED-DIET-001]
---

# SPEC: always-loaded 지시문 표면의 허용 제거 풀 A_adm 실측과 150,000자 한도 달성 가능성 판정

## HISTORY

| 날짜 | 버전 | 변경 | 작성자 |
|---|---|---|---|
| 2026-09-27 | 0.1.0 | 최초 작성. 카드 t1226(Tier M · 클래스 C). `SPEC-ALWAYS-LOADED-DIET-002`(t1175)가 채무로 넘긴 두 미결 — 허용 풀 `A_adm` 미측정(`acceptance.md §AC-ALD2-001.3`)과 `P절` 재조정 (a)/(b) 미결(`design.md §4.0`) — 을 이 카드가 측정으로 닫고, 그 결과로 150,000자 런타임 한도의 달성 가능성을 판정한다. 기준선은 오케스트레이터가 이 트리(`7fe658815`)에서 잰 `199111 total` 이다. | manager-spec |

---

## §A. 배경

Claude Code 런타임은 always-loaded 지시문 파일 18개의 합계가 150,000자를 넘으면 매 세션 경고를 낸다. `SPEC-ALWAYS-LOADED-DIET-002`(t1175)는 합계를 기준선 246,943자에서 줄였으나 한도에는 닿지 못했고, 그 잔여를 **성격 미확정 채무**로 이 카드에 넘겼다. 성격이 확정되지 않은 이유는 하나다 — 구속 조항 동결(REQ-ALD2-002·003, 동결 해시 `d97b33d9…c6c3`, 170줄) 아래에서 실제로 더 들어낼 수 있는 양, 곧 허용 풀 `A_adm` 이 측정되지 않았다. t1175 가 확립한 것은 `A_adm ≤ 73,126`(공표값 계열)이라는 상한 하나와, 구조적 하한 F 가 세 값의 구간 `{170,528 · 171,695 · 172,863}` 이라는 서술뿐이다.

### 이 트리의 기준선 — 오케스트레이터 실측, 이 실행

- 기준 커밋: `7fe658815eb0d4110b9acadad56e5a85bee3ed3f` (`git rev-parse HEAD`, 깨끗한 트리). t1175 병합을 포함한 로컬 `develop` 이다.
- 명령: `SPEC-ALWAYS-LOADED-DIET-002/acceptance.md §AC-ALD2-001` 의 18경로 `wc -m` 블록 그대로(이 SPEC `acceptance.md §AC-ALH-001` 에 전문 재수록).
- 관측: `199111 total`. 150,000 에 대한 잔여 **49,111**.
- 동결 다중집합 sha256(같은 트리, 같은 실행에서 manager-spec 이 재실행): `d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3` — t1175 기준선과 동일.

카드 본문의 197,897(병합 전 트리 `3a48485af`)과 198,361(t1175 HEAD `8e50ef148`)은 **다른 트리의 값**이다. 이 카드는 그 둘을 폐기된 수치로 기록만 하고 어떤 산술에도 쓰지 않는다.

### 계량기가 두 벌이다 — 도그푸드 표면과 배포 표면

같은 18경로라도 이 저장소의 라이브 사본(도그푸드)과 `moai init` 이 사용자 프로젝트에 까는 사본은 다르다. 템플릿은 `AGENTS.md` 를 `AGENTS.md.tmpl` 로, `user.yaml`·`language.yaml` 을 `.tmpl` 로만 갖고 있어 렌더링 뒤에야 실물이 생긴다(원본 18개 대응 파일의 미렌더링 합계 203,611 — `research.md §1`). 카드 본문은 「상시 로드 **템플릿** rules 풀」을 말하고, t1184 는 격리된 `moai init` 트리에서 같은 경고(18파일 249.2k)를 관측했다. 사용자가 보는 경고는 배포 표면에서 난다. 이 SPEC 은 두 표면을 모두 재고, 판정을 표면별로 낸다(REQ-ALH-002).

---

## §B. 용어

| 기호 | 뜻 |
|---|---|
| `S_init` | 이 트리에서 빌드한 바이너리로 격리 디렉터리에 `moai init` 을 실행해 얻은 18파일 표면 — **사용자가 보는 표면, 한도 판정의 1차 대상** |
| `S_live` | 이 저장소 라이브 사본의 18파일 표면(도그푸드) — 기준선 199,111 |
| 후보 | always-loaded 표면에서 제거·이동·압축을 검토하는 절 또는 문단 하나 |
| M1 / M1′ / M2 | companion 재배치 / 비구속 산문 중복 제거 / 제자리 압축 (`SPEC-ALWAYS-LOADED-DIET-002 §B` 정의 승계) |
| 조건 1~4 | `SPEC-ALWAYS-LOADED-DIET-002/plan.md §C` M1 조건 — 1 구속 조항 줄 0, 2 역방향 인용 생존, 3 목적지 `paths:` 도달, 4 목적지 40,000자 수용량 |
| `A_adm` | 조건 1~4·REQ-ALD2-011·REQ-ALD2-013·동결 해시를 모두 지키며 제거 가능한 자수의 합 |
| `R` | 제거에 따라 stub 에 새로 들어가는 포인터 줄의 자수(재유입) |
| `T_min` | `현재 합계 − A_adm + R` — 동결 아래 달성 가능한 최저 합계 |

---

## §C. GEARS 요구사항

- **REQ-ALH-001** (Ubiquitous) — The 판정서 `.moai/reports/t1226/verdict.md` shall 첫 머리에 레인 백엔드 줄(`레인 백엔드: Claude Opus 5.5 (claude-opus-5-5)`), 측정 명령 전문, 기준 커밋 `7fe658815eb0d4110b9acadad56e5a85bee3ed3f`, 기준선 `199111 total` 과 잔여 49,111 을 고정해 적는다.
- **REQ-ALH-002** (Ubiquitous) — The 측정 shall `S_init` 과 `S_live` 두 표면을 각각 재고, 두 표면의 18파일 합계·후보 표·`T_min`·판정을 표면별로 따로 적는다. 150,000자 한도 판정의 1차 대상은 `S_init` 이다.
- **REQ-ALH-003** (Ubiquitous) — The 후보 표 shall always-loaded 표면의 모든 비구속 절과 문단을 한 행씩 담고, 행마다 파일·절 이름·자수·기제(M1 / M1′ / M2)·조건 1~4 각각의 판정과 증거 명령·허용/기각·기각 사유를 채운다.
- **REQ-ALH-004** (Where) — **Where** 후보의 파일이 `AGENTS.md` 인 경우, the 후보 표 shall 그 후보를 M2 로만 허용하고 M1 행으로 계상하지 않는다(REQ-ALD2-011 승계).
- **REQ-ALH-005** (Unwanted) — The 측정 shall not 구속 조항 줄(`[HARD]` / `MUST` / `shall ` 을 담은 줄)의 삭제·재작성·재배치를 요구하는 후보를 `A_adm` 에 계상한다 — 같은 의무가 path-scoped 룰에 있다는 이유의 중복 제거(REQ-ALD2-013)도 포함한다.
- **REQ-ALH-006** (When) — **When** 후보가 M2 로 분류될 때, the 측정 shall 그 후보의 허용 자수를 구속 조항 줄의 하드랩을 바꾸지 않는 실제 압축 시도로 재고, 시도하지 않은 후보의 값을 수율 외삽(예: 10.77%)으로 채우지 않는다. 시도하지 않은 후보는 자수 칸을 `미측정` 으로 두고 Gap 에 올린다.
- **REQ-ALH-007** (When) — **When** 측정이 `P절` 을 확정할 때, the 측정 shall `SPEC-ALWAYS-LOADED-DIET-002/design.md §4.0` 이 쓴 절 분할을 원 트리(`172ef22eb`)에서 재실행해 재조정 (a)·(b) 중 하나를 증거와 함께 고르고, F 를 단일값으로 적는다.
- **REQ-ALH-008** (Ubiquitous) — The 절 분할 스크립트와 후보 표를 만든 명령 shall `.moai/reports/t1226/` 아래에 커밋되어 재실행 가능하다 — `/tmp` 경로의 스크립트를 증거로 인용하지 않는다.
- **REQ-ALH-009** (Ubiquitous) — The 판정서 shall 표면마다 `T_min = 현재 합계 − A_adm + R` 을 Claim 으로 세우고, 세 항 각각에 명령과 관측 출력을 붙인다. `R` 은 제거 후보를 목적지별로 묶었을 때 stub 에 들어가는 포인터 줄 자수의 실측값이다.
- **REQ-ALH-010** (Ubiquitous) — The 판정서 shall 표면마다 판정 토큰 하나를 적는다 — `T_min < 150,000` 이면 `ACHIEVABLE`, `T_min ≥ 150,000` 이면 `STRUCTURALLY-INFEASIBLE-UNDER-FREEZE`, `미측정` 후보의 자수 합을 빼고 더한 두 `T_min` 이 150,000 을 사이에 두면 `UNDETERMINED`.
- **REQ-ALH-011** (When) — **When** `S_init` 의 판정이 `STRUCTURALLY-INFEASIBLE-UNDER-FREEZE` 일 때, the 판정서 shall 동결 해제 상신 절차를 담는다 — (a) 해제 대상 구속 조항 줄의 목록과 그 줄이 묶어 두는 자수, (b) 150,000 에 닿는 최소 해제 집합, (c) 위험(REQ-AMC-002 강등, `AGENTS.md` 자기충족성, 템플릿 중립성), (d) 결정권자는 운영자이며 리드가 `AskUserQuestion` 으로 상신한다는 명시.
- **REQ-ALH-012** (Unwanted) — The 레인 shall not 동결 해제 여부를 결정하거나 권고를 결정으로 적는다. 레인이 내는 것은 측정값과 선택지이며, 판정서의 해당 절은 `RECOMMEND:` 접두의 권고로만 끝난다.
- **REQ-ALH-013** (Unwanted) — The 이 카드의 구현 shall not 18개 계수 파일과 그 템플릿 미러의 내용을 수정한다. 실제 감축은 후속 카드의 일이다.
- **REQ-ALH-014** (Unwanted) — The 판정서 shall not 다른 트리·다른 시점에서 잰 수치를 이 트리의 측정값으로 쓴다. 197,897 과 198,361 은 폐기 수치로만 적는다.
- **REQ-ALH-015** (Ubiquitous) — The 판정서 shall 조건 4(목적지 40,000자 수용량)를 라이브·템플릿 두 트리 각각의 목적지 크기로 판정하고, 이미 40,000자를 넘은 파일을 목적지로 허용하지 않는다(REQ-ALD2-015 승계).

---

## §D. 범위 밖

### Out of Scope — 18개 계수 파일의 실제 감축

- 이 카드는 측정과 판정만 한다. `CLAUDE.md`·`AGENTS.md`·두 yaml·14개 룰 파일과 그 템플릿 미러를 고치지 않는다.
- 판정이 `ACHIEVABLE` 이면 그 감축은 후속 카드로 발행을 요청한다. 발행은 리드의 일이다.

### Out of Scope — 동결 해제의 결정과 집행

- 동결 해시 `d97b33d9…c6c3` 의 해제 여부는 운영자가 정한다. 이 카드는 상신 자료를 만들 뿐 결정하지 않고, 해시를 바꾸지 않는다.

### Out of Scope — 출력 스타일과 토큰 가드

- `.claude/output-styles/moai/moai.md` 와 `internal/config/token_budget_guard.go` 의 `AlwaysLoadedTokenBudget` 은 다른 계량기다. 이 카드의 어떤 AC 도 그것을 재지 않는다.

### Out of Scope — 파일당 40,000자 초과 4건의 수리

- 이미 한도를 넘은 path-scoped 룰 4개의 수리는 카드 t1180 소관이다. 이 카드는 그 파일들을 목적지 후보에서 제외하는 데만 쓴다.

---

## §E. 성공 기준

- 두 표면 모두에서 후보 표가 완결되고, `A_adm` 이 행 합으로 재현된다.
- `P절` 재조정이 증거와 함께 하나로 닫히고 F 가 단일값으로 적힌다.
- `S_init` 과 `S_live` 각각에 판정 토큰 하나가 붙는다.
- 18개 계수 파일과 그 템플릿 미러가 기준 커밋 대비 바이트 동일하다.

---

## §F. 교차참조

- `.moai/specs/SPEC-ALWAYS-LOADED-DIET-002/spec.md` — REQ-ALD2-001(개정 3), REQ-ALD2-002·003·011·013·015
- `.moai/specs/SPEC-ALWAYS-LOADED-DIET-002/acceptance.md` — §AC-ALD2-001(측정 블록), §AC-ALD2-001.1(F 구간), §AC-ALD2-001.3(`A_adm ≤ 73,126`, 편향 방향 셋), AC-ALD2-002(동결 해시)
- `.moai/specs/SPEC-ALWAYS-LOADED-DIET-002/plan.md §C` — M1 조건 1~4
- `.moai/specs/SPEC-ALWAYS-LOADED-DIET-002/design.md` — §1, §4.0, §4.3, §4.4
- `.moai/reports/t1175/verdict.md`, `sync-audit.md`, `sync-reaudit.md`
- `.claude/rules/moai/core/verification-claim-integrity.md` §2 — 귀속 요구
