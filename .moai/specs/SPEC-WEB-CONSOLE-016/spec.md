---
id: SPEC-WEB-CONSOLE-016
title: "moai web console — partial-apply failure reporting: disk-truth re-render and banner precision"
version: "0.1.0"
status: draft
created: 2026-09-20
updated: 2026-09-20
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/web
lifecycle: spec-anchored
tags: web-console, handleSave, partial-apply, failure-reporting, banner, observability
era: V3R6
tier: M
related_specs: [SPEC-WEB-CONSOLE-007, SPEC-WEB-CONSOLE-011, SPEC-GLM-KEY-INPUT-001]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-20 | manager-spec | 최초 draft. 근거는 카드 t1049 의 1차 측정(`.moai/reports/t1049/findings.md`, 계측기 `internal/web/partial_apply_repro_test.go`, 커밋 `a52ce60f0`) — 측정 트리 `WT-partial-apply` @ `d8304b49a`(로컬 develop 과 동일). 스코프는 리드 판정(2026-09-20)에 따라 **실패 보고 표면 2축**으로 한정하고, 트랜잭션 축은 §F 에서 명시적으로 제외한다. |

---

## §1 Context & Motivation

`handleSave`(`internal/web/handlers.go:350`)의 영속화 구간은 **465-549행, 고정 순서 7단계**다: (1) `writePreferences` :465, (2) `recordLastProfile` :478(advisory), (3) `syncToProject` :482, (4) `writeProjectConfig` :494, (5) `writeProjectNestedConfig` :503, (6) `applySchemaEdits` :513, (7) `applyPerfTierEdits` :523, (8) `patchAgentFM` :533, (9) `glmcred.Save` :546. 각 단계는 오류에서 즉시 반환하므로 **N 단계가 실패하면 1..N-1 은 이미 디스크에 들어가 있다.** 주입 가능한 6단계 전부에서 이 잔존 상태가 관측됐다.

이 SPEC 이 다루는 것은 **그 잔존 상태 자체가 아니라, 그 상태를 사용자에게 어떻게 알리는가**이다. 측정된 두 결함:

1. **실패 페이지는 칸마다 출처가 다르고, 그 구분이 화면에 없다.** `renderErrorPage`(선언 :600)는 `projectView`(선언 :589)를 거치는데, 그 경로에서 **어떤 표면은 디스크값을, 어떤 표면은 제출값을 렌더한다** — 그리고 어느 칸이 어느 쪽인지 표시하는 것이 아무것도 없다. 사용자는 페이지 안에서 자기 입력과 저장된 값을 구별할 수단이 없다(페이지를 새로 고치면 드러난다 — 화면 안에 단서가 없다는 뜻이지, 영영 알 수 없다는 뜻이 아니다). 6개 실패 사례 전부에서 폼이 제출한 `user_name` 을 그대로 되비췄고, **1단계가 실패해 아무것도 쓰이지 않은 경우에도 그랬다.** 화면은 원하는 상태와 구별되지 않으며, 진실은 페이지를 새로 고쳐야만 드러난다. 성공 경로는 다르다 — `successProjectView`(선언 :564)는 nested config 를 디스크에서 다시 읽는다. 즉 **디스크 재독 seam 은 이미 존재하고, 실패 경로만 그것을 쓰지 않는다.**

   표면별로 정밀하게: 실패 경로에서도 10섹션 스키마 값은 `applySchemaCurrentBestEffort`(:593)가 디스크에서 best-effort 로 읽는다 — 이쪽은 이미 디스크 진실이다. 제출값이 되비치는 것은 **프로필 preferences 와 두 스칼라**(`devMode`/`convention`)이며, t1049 가 실제로 관측한 것도 그 축(`user_name`)이다. nested 표면은 실패 경로에서 `applyNestedCurrent` 도 `applyNestedForm` 도 거치지 않으므로 성공·거부 경로와 다르게 렌더되는데, **그 렌더 결과는 재지 않았다**(§6-4).
2. **배너가 아는 것보다 넓게 주장한다.** 3-6단계는 `profile preferences saved, but <X> failed`, 8-9단계는 `settings saved, but <X> failed` 로 갈린다. 뒤쪽 문구는 핸들러가 확인한 범위를 넘어선다.

### §1.1 왜 이것이 "존중해야 할 설계"가 아니라 "반복으로 굳은 형태"인가

원자성 경계는 **검증 단계에 명시적으로 그어져 있다**(`:456`, atomic reject, REQ-WC-008 — 검증 실패 시 어느 층도 쓰이지 않는다). 영속화 구간 안에는 그런 경계가 없다.

여섯 개 영속화 경계 중 **선언된 근거를 가진 것은 하나뿐**이다 — `:483-486` 의 "Advisory D1" 주석과 그것을 지키는 기존 테스트 `TestSaveSyncFailureSurfacesReadableError`. **나머지 다섯 경계는 같은 형태를 따르되 기록된 결정이 없다.** 이 1/6 비율이 본 SPEC 의 전제다: 하나는 의도된 설계이므로 존중하고, 다섯은 결정 없이 굳은 형태이므로 재검토 대상이다. 다섯을 "설계"로 읽으면 고칠 이유가 사라지고, 하나를 "굳은 형태"로 읽으면 명시된 결정을 지운다.

### §1.2 측정 귀속

본 문서의 모든 file:line 과 수치는 트리 `WT-partial-apply` @ `d8304b49a` 에서 t1049 레인이 이번 실행으로 측정한 값이다. 카드 본문이 인용한 좌표(`handlers.go:464-548`)는 t1043 트리(`0bf27ea69`) 기준이며 본 SPEC 은 그것을 옮겨 쓰지 않았다. 증거: `.moai/reports/t1049/findings.md`, `.moai/reports/t1049/repro.log`, `.moai/reports/t1049/web_pkg.log`.

---

## §2 Requirements (GEARS)

### §2.1 축 (4) — 실패 페이지가 디스크의 실제 상태를 보여준다

**REQ-WC16-001 (When):** When a persistence step of `handleSave` fails, the console shall re-render the form from values re-read from disk, not from the values submitted in the failing request.

**REQ-WC16-002 (Ubiquitous):** The partial-apply failure page shall present, per persistence step, whether that step landed or did not land — so the reader distinguishes the applied prefix from the unapplied remainder without refreshing the page.

**REQ-WC16-003 (When):** When the first persistence step fails, the failure page shall state that no submitted value landed, and shall not present any submitted value as a current value.

**REQ-WC16-004 (Where):** Where a re-read of the on-disk value itself fails, the console shall render that value as unverified rather than falling back to the submitted value — a degraded read shall never be presented as disk truth.

**REQ-WC16-005 (Unwanted):** The failure page and its banner shall not contain any submitted value of a secret-bearing field; a failure involving the GLM credential surface shall name the layer and the target file only.

### §2.2 축 (5) — 배너는 아는 것만 말한다

**REQ-WC16-006 (Ubiquitous):** Every persistence-failure banner shall assert exactly the set of steps the handler observed as completed, and no wider set.

**REQ-WC16-007 (Unwanted):** A persistence-failure banner shall not use a whole-save success phrase (the `settings saved` form at `handlers.go:533` and `:546`) when only a proper subset of the persistence steps completed.

**REQ-WC16-008 (When):** When the failing step is the first persistence step, the banner shall state that nothing was saved.

### §2.3 검증 범위와 재사용

**REQ-WC16-009 (Ubiquitous):** The delivered verification shall cover the failure-reporting behaviour at **all seven** persistence steps, including `applyPerfTierEdits` (step 7) and `glmcred.Save` (step 9), whose failure behaviour was NOT measured at plan time — their absence from the t1049 instrument is a measured gap (no injection seam: both are package-level functions), never an absence of the hazard.

**REQ-WC16-010 (Ubiquitous):** The run phase shall extend the existing injection harness `recordingSeams` (`internal/web/partial_apply_repro_test.go`) rather than replace it; sibling card t1051 builds on the same harness, so a rewrite that breaks its seam list is a regression.

**REQ-WC16-011 (While):** While the "Advisory D1" decision at `handlers.go:483-486` stands, the run phase shall preserve the behaviour its guard test `TestSaveSyncFailureSurfacesReadableError` asserts — this is the one persistence boundary carrying a declared rationale, and precision work on the banner shall not silently revoke it.

---

## §3 Acceptance Criteria

전량은 `acceptance.md` (AC-WC16-001 .. AC-WC16-011). 각 AC 는 명령으로 반증 가능하며, **현재 코드에서 무엇 때문에 실패하는지**를 함께 기록한다.

---

## §4 Constraints (HARD)

- **HARD-1** 영속화 단계의 **순서·개수·존재를 바꾸지 않는다.** 본 SPEC 은 보고 표면만 만진다.
- **HARD-2** 롤백·트랜잭션·보상 쓰기를 도입하지 않는다(§F.1).
- **HARD-3** 비밀값(GLM API key)은 어떤 보고 표면에도 나타나지 않는다 — 층 이름과 파일 경로만(REQ-WC16-005, 형제 카드 t1051 과 공유하는 경계).
- **HARD-4** `recordingSeams` 의 seam 목록과 시그니처를 보존한다(REQ-WC16-010).
- **HARD-5** 검증 실패 경로(`:456` atomic reject)는 무변경 — 그 경계는 이미 명시적으로 그어져 있다.
- **HARD-6** 실패 배너 문자열은 현재 Go 리터럴이며 `root.templ:151` 에서 그대로 렌더된다(이 트리 실측 — i18n 키를 거치지 않는다). 본 SPEC 은 그 관례를 바꾸지 않는다; 배너를 i18n 표면으로 옮기는 것은 별건이다.

---

## §5 Measured Inventory

영속화 7단계와 관측된 잔존 상태(주입 가능한 6단계, t1049 실측):

| 실패 지점 | 이미 디스크에 들어간 단계 | 현재 배너 문구 |
|---|---|---|
| 1 `writePreferences` :465 | (없음) | `could not save profile preferences: …` |
| 2 `recordLastProfile` :478 | — (advisory, 오류를 삼킨다; `selected != "default"` 가드) | — |
| 3 `syncToProject` :482 | 1 | `profile preferences saved, but …` |
| 4 `writeProjectConfig` :494 | 1, 3 | `profile preferences saved, but …` |
| 5 `writeProjectNestedConfig` :503 | 1, 3, 4 | `profile preferences saved, but …` |
| 6 `applySchemaEdits` :513 | 1, 3, 4, 5 | `profile preferences saved, but …` |
| 7 `applyPerfTierEdits` :523 | **미측정**(seam 없음) | `profile preferences saved, but …` |
| 8 `patchAgentFM` :533 | 1, 3, 4, 5, 6 | `settings saved, but …` ← 과잉 주장 |
| 9 `glmcred.Save` :546 | **미측정**(seam 없음) | `settings saved, but …` ← 과잉 주장 |

---

## §6 Gaps carried forward — 미측정이지 부재가 아니다

이 세 항목은 plan-phase 에서 **재지 못했다.** 어느 것도 "그런 위험이 없다"는 뜻이 아니다.

1. **7·9단계의 실패 거동.** `applyPerfTierEdits`, `glmcred.Save` 는 패키지 함수라 주입 seam 이 없어 재현하지 못했다. 파일시스템 수준 탐침(대상 경로를 쓰기 불가로 만들기)이 필요하며, run-phase 의 일이다(REQ-WC16-009).
2. **각 층의 자체 원자성.** 한 층(profile store / config manager / yamlpatch) 안에서 부분 쓰기가 일어나는지는 재지 않았다. 이 값은 **트랜잭션 축의 비용 산정에 필요한 선행 입력**이며, 그 카드의 첫 작업이다(§F.1).
3. **계측 대리값.** 계측기는 디스크 바이트가 아니라 seam 호출을 관측한다. 실제 구현이 물려 있을 때 호출 = 쓰기 시도이므로 충실한 대리값이지만, **동일하지는 않다.**
4. **실패 경로의 nested 렌더 결과.** `renderErrorPage → projectView` 는 `applyNestedCurrent`(성공 경로)도 `applyNestedForm`(거부 경로)도 호출하지 않는다는 것까지는 코드에서 확인했으나, **그 결과 화면에 무엇이 보이는지는 재지 않았다.** 빈 값인지 기본값인지, 그것이 사용자에게 어떻게 읽히는지는 run-phase 측정 대상이다.
5. **실제 브라우저 표시.** t1049 의 관측은 렌더된 HTML 바이트 수준이며, 브라우저에서 무엇이 보이는지는 재지 않았다.

---

## §F Exclusions (What NOT to Build)

본 절은 이 SPEC 이 **만들지 않는 것**을 열거한다. 아래 항목은 전부 out of scope 이며, 삭제 대상이 아니라 후속 판단의 입력이다.

### Out of Scope — 트랜잭션 축(7단계 원자화 / 3층 롤백)

- 일곱 개 영속화 단계를 원자적으로 만드는 일, 그리고 세 영속화 층(profile store / config manager / yamlpatch)을 가로지르는 롤백은 본 SPEC 밖이다.
- **왜 별건인가**: 두 축은 서로 독립이다. (4)는 **트랜잭션 없이 고칠 수 있고**, 반대로 트랜잭션을 도입해도 (4)가 자동으로 고쳐지지 않는다 — 제출값을 되비추는 것은 렌더 경로의 성질이지 쓰기 경로의 성질이 아니다. 두 축을 합치면 처방이 커지고, 큰 쪽이 작고 확실한 쪽을 인질로 잡는다.
- SPEC-WEB-CONSOLE t1043 이 이미 그 크기를 자기 스코프 밖으로 판정했다. 리드가 **별도 카드로 발행해 운영자 판단에 올린다**; 본 SPEC 은 그것을 선취하지 않는다.
- 그 카드의 **첫 작업은 §6-2(층별 자체 원자성 측정)**다 — 측정 없이 비용을 말할 수 없다.

### Out of Scope — 영속화 순서·구성의 변경

- 단계의 순서 바꾸기, 합치기, 나누기, 제거하기.
- `recordLastProfile`(2단계)의 advisory 성질 변경 — 그 침묵은 `:483-486` 에 선언된 결정이다.

### Out of Scope — 저장 실패의 로깅·관측 가능성 표면

- 서버 측 로그 라인, 구조화 이벤트, 진단 덤프는 형제 카드 **t1051** 의 스코프다. 본 SPEC 은 **사용자에게 보이는 페이지와 배너**만 다룬다.
- 두 카드가 같은 `handleSave` 와 같은 주입 하네스(`recordingSeams`)를 만지므로, 본 SPEC 은 그 하네스를 보존할 의무만 진다(REQ-WC16-010).

### Out of Scope — 배너의 i18n 이전

- 실패 배너는 현재 Go 리터럴이며 i18n 키를 거치지 않는다(이 트리 실측). 문구를 정밀화하되, i18n 표면으로 옮기지 않는다.

### Out of Scope — 새 config 섹션·필드·검증기

- 편집 가능한 필드 집합은 무변경. 본 SPEC 은 필드를 더하지도 빼지도 않는다.

### Out of Scope — CSRF/token 인프라

- `app.go` 의 REQ-WC-009 금지 조항은 그대로다. 본 SPEC 은 쓰기 안전 모델을 건드리지 않는다.

---

## §7 Cross-References

- 측정 기록: `.moai/reports/t1049/findings.md` (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)
- 계측기: `internal/web/partial_apply_repro_test.go` (커밋 `a52ce60f0`)
- 선행 결정: `handlers.go:483-486` "Advisory D1" + `TestSaveSyncFailureSurfacesReadableError`
- 원자성 선례: `handlers.go:456` 검증 단계 atomic reject (REQ-WC-008)
- 형제 카드: t1051(저장 실패 관측 가능성), 트랜잭션 축 카드(미발행, 운영자 판단 대기)
