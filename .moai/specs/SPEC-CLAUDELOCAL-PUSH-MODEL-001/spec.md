---
id: SPEC-CLAUDELOCAL-PUSH-MODEL-001
title: "CLAUDE.local.md push-model 정본화 — 미커밋 사본의 정본 참칭 차단과 흡수 대상 정정"
version: "0.1.1"
status: in-progress
created: 2026-09-08
updated: 2026-09-08
amendment_of: SPEC-CLAUDELOCAL-PUSH-MODEL-001
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "CLAUDE.local.md"
lifecycle: spec-anchored
tags: "claude-local, push-model, gitflow, doctrine, t531"
tier: S
---

# SPEC-CLAUDELOCAL-PUSH-MODEL-001 — CLAUDE.local.md push-model 정본화

## HISTORY

| 날짜 | 판 | 변경 | 근거 |
|---|---|---|---|
| 2026-09-08 | 0.1.0 | 최초 작성 (plan-phase). 카드 t531 의 전제가 재측정으로 **정정된 뒤** 작성됐다 — 아래 §A 참조 | `.moai/reports/t531/premise-remeasure.md` |
| 2026-09-08 | 0.1.1 | 제자리 개정 개시. `completed → in-progress`. 아래 §Amendments 참조 | `.moai/reports/t531/verdict.md` |

## Amendments

### 개정 1 — sync-audit FAIL 두 건의 산출물 수리 (2026-09-08)

| 항목 | 값 |
|---|---|
| 직전 completed 판 | `0.1.0` |
| 직전 completed SHA | `b7344d957` (close 를 실은 sync 커밋) |
| 개정 종류 | 제자리 개정 (in-place). 후속 SPEC 이 아니므로 `amendment_of` 가 자기 자신을 가리킨다 |
| 범위 | **본문 변경**은 `CLAUDE.local.md` §2.3 · §4.1 두 절뿐이다. 여기에 더해 재close 는 아래 「재close 가 함께 현행화하는 형제 산출물」이 열거하는 것을 같은 시점에 갱신한다 |

**근거.** sync-audit 이 FAIL 로 돌아왔다(`.moai/reports/t531/verdict.md`). AC-CLPM-001..008 은
감사자가 살아 있는 대조군과 함께 전부 재실행해 통과했다. 결함은 **어떤 인수 기준도 보지 않는
자리**, 즉 산출물 `CLAUDE.local.md` 의 본문에 있다.

- **F1 — 이 카드가 자기 파일 안에 만든 자기모순.** §2.3 은 `moai update` 실행 전 추적 파일
  수정이 0이어야 한다는 `[HARD]` 전제를 걸어 두었는데, 이 카드가 새로 넣은 §0.4 는 primary
  체크아웃에서 ` M CLAUDE.local.md` 가 **설계상 상시**임을 선언하고 그 파일에 `git restore` 를
  금지한다. 그 전제는 primary 체크아웃에서 성립할 수 없고, §2.3 을 문자 그대로 따르는 독자는
  §0.4 가 금지하는 바로 그 되돌림 — REQ-CLPM-004 가 막으려고 존재하는 행위 — 로 떠밀린다.
- **F2 — 한 방향만 고친 정정.** 고쳐 쓴 §4.1 은 로컬 `develop` 이 `origin/develop` 보다
  **앞설** 수 있다는 것만 설명한다. 반대 방향 — 다른 레인이 착지하면 로컬 develop 이
  **뒤처진다**, 그래서 낡은 베이스 위에서 작업하게 된다 — 은 형제 독트린
  `.claude/rules/local/gitflow-lane-protocol.md` §11 이 판단 기준과 갱신 경로까지 담아 다루는데,
  §4.1 은 그 거울상 위험을 그대로 남겼다. 실측: `CLAUDE.local.md` 에서 `뒤처` 0건, `앞설` 1건.

**범위 한계.** 요구사항 변경 없음, 인수 기준 변경 없음. 여덟 개 AC 는 적힌 그대로 유지되며
감사자가 이미 재검증했다. 개정이 고치는 것은 **전달된 본문**이지 계약이 아니다.

**재close 가 함께 현행화하는 형제 산출물.** 본문 변경은 위 두 절에서 끝나지만, 재close 는
아래 둘을 같은 시점에 현행화한다. 재close 가 건드릴 산출물을 범위에서 빼 두면 낡은 사본 위에서
닫히는 것을 부르는 셈이고, F1 이 바로 그 모양 — 문자 그대로 읽으면 카드가 막으려던 자리로
떠미는 규칙 — 의 결함이었다.

1. **`CHANGELOG.md`** — 이 카드의 항목 두 자리. (a) 닫힘 상태 문장이 아직
   `in-progress → completed` 를 주장하는데, 이 개정이 그 close 를 되돌렸다. (b) 흡수 논거가
   로컬 `develop` 이 `origin/develop` 보다 **앞설** 수 있다는 한 방향만 적고 있는데, 수리된
   §4.1 은 양방향을 준다. 둘 다 수리된 본문에 맞춰 고쳐 쓴다.
2. **`progress.md` §E.4** — 개정 이후 재측정을 **§E.4 에 새 날짜 블록으로 덧붙인다**(추가이지
   수정이 아니다). [HARD] **§E.3 은 적힌 그대로 둔다.** 그 `ac_matrix` 는
   `head_at_measurement: a833b5658` 을 달고 있는, 귀속이 성립한 run-phase 기록이며 `arm1: 2`
   는 그 커밋에서 참이었다. 덮어쓰면 귀속된 baseline 을 파괴해 결함 하나를 다른 결함으로
   바꾸는 것이다. 개정 이후 수치(`arm1`, `AC-CLPM-007` 의 대조군 값)는 §E.3 을 건드리지 않고
   §E.4 안에만 들어간다.

---

## §A 배경 — 카드의 전제는 틀렸고, 정정된 것을 쓴다

카드 t531 은 「`CLAUDE.local.md` 사본이 둘(main / develop)이고 서로 반대 지시를 한다」고 서술했다.
**그것은 실제 트리에 있는 것이 아니다.** 이 SPEC 의 모든 수치는
`.moai/reports/t531/premise-remeasure.md`(이 워크트리, base `bce6d7e08`, 2026-09-08 측정)에서 온다.
여기서 재유도하지 않고 그 파일을 인용한다.

측정된 것은 **세 변종**이다:

| # | 좌표 | §4.1 제목 | `push origin develop` 출현 | 모델 |
|---|---|---|---|---|
| 1 | `origin/main` = `main` = `7ad9f8534`, **커밋된 블롭** | `§4.1 로컬 통합 레인 (develop)` | **0** | **폐기된 제3의 모델** — 「develop 은 원격에 올리지 않는다」, 카드별 main PR |
| 2 | primary 체크아웃 **워킹트리, 미커밋** | `§4.1 GitFlow 통합 체인 (develop)` | **3** | 레인이 창 안에서 develop 을 push |
| 3 | `origin/develop` 블롭 | `§4.1 GitFlow 통합 체인 (develop)` | **2** (둘 다 리드 일괄 절차 안) | **리드 일괄 push** — 레인은 push 하지 않음 |

리드가 별도로 잰 `.claude/worktrees/develop/CLAUDE.local.md` 는 2건으로 블롭과 동일했다.
거기에 제4의 변종은 없었다.

### §A.1 축 1 — 「main 대 develop 갈라짐」이 아니다

피해를 낸 지시가 사는 곳은 **변종 2**, 즉 git 이 추적하지만 **어느 브랜치에도 커밋된 적 없는
워킹 사본**이다. `main` 의 커밋본에는 그 문장이 아예 없다(0건, 제목부터 다르고 내용은 폐기 모델).

따라서 결함의 이름은 「스테일한 브랜치를 인용했다」가 아니라
**「이력에 없는 워킹 사본이 정본 노릇을 했다」**이며, 그 아래에 한 층이 더 있다 —
**`main` 의 커밋본 자체가 폐기된 모델**이다.

2026-09-07 에 당시 리드가 레인 12곳에 레인-push 절차를 배포하면서 「main 사본」을 인용했다고
적었는데, 실제로 인용된 것은 거의 확실히 변종 2 다. **[추론]** — 텍스트 일치에서 나온 추론이고,
그 배차 자체를 관측한 것이 아니다(§G-4).

### §A.2 축 2 — 갈라짐이 아니라 두 사본 공통의 결함

`git merge origin/develop` 흡수 단계는 **변종 2 와 변종 3 에 동일하게** 있다(diff 로 확인:
양쪽이 같은 줄을 낸다). **사본을 일치시켜도 이 결함은 남는다.**

리드 일괄 push 모델에서 로컬 `develop` 은 `origin/develop` 보다 앞설 수 있다 — 레인들이 로컬
병합을 마쳤고 리드가 아직 push 하지 않은 구간이 존재한다. 그 구간에서 `origin/develop` 을
흡수하면 **다른 레인의 착지분이 빠진 베이스에서 재측정**하게 된다.

수리 형태는 「두 사본의 화해」가 아니라 **「정본 한 곳의 내용 정정」**이다.

### §A.3 집행된 처분 (운영자 결재, 기록만 한다)

운영자 결재로 리드가 primary 워킹본을 develop 정본으로 교체했다. 이 세션에서 독립 검증했다
(증거 파일 § 추가 2026-09-08):

- 백업: `.moai/reports/t531/CLAUDE.local.md.primary-uncommitted-backup-20260908`
  (**PRIMARY 체크아웃**에 있다. 이 워크트리가 아니다). 40545 바이트, **660줄**,
  `push origin develop` **3건**(변종 2 의 지문),
  sha256 `23f8427589b739c705c8b17def0409c187af92ce5f27ee8bb87037a8376c9a7a`.
- 교체 후: primary 워킹본 **734줄**, `push origin develop` **2건**,
  `git show origin/develop:CLAUDE.local.md` 와 `diff -q` rc=0 — **바이트 동일**.
  **변종 2 는 소멸했다.**
- 교체는 워킹 사본이 지웠던 절들을 되살렸다(develop 대비 116줄 삭제 / 42줄 추가였다):
  `§2.0 에이전트 정의는 사본이 셋이고…`, `Command-to-Skill Publication`,
  `인박스 유용성 범위`, git-flow 재적용 블록, statusline 폴백 체인 주석.

이 SPEC 은 이 처분을 **집행된 사실로 기록**하고 재발 방지에 집중한다. 처분을 다시 다투지 않는다.

---

## §B 요구 (GEARS)

> Tier S — 요구 상한 8, AC 상한 8.

**REQ-CLPM-001 (Ubiquitous)** — `CLAUDE.local.md` 의 정본은 **레인이 분기하는 트리의 사본**이어야
한다. 현재 그 트리는 `develop` 이다. 문서는 이 선택 규칙을 명시적 문장으로 담아야 한다 —
날짜나 「나중에 전달된 쪽」이 아니라 **분기 트리**가 판별식이다.

**REQ-CLPM-002 (Ubiquitous)** — 레인 창 절차의 흡수 대상은 **로컬 `develop`** 이어야 한다.
`git merge origin/develop` 은 리드 일괄 push 모델에서 다른 레인의 착지분을 빠뜨린 베이스를
만들므로, 정본에서 정정되어야 한다.

**REQ-CLPM-003 (Where)** — primary 체크아웃이 `main` 에 체크아웃돼 있는 동안,
`git status` 의 ` M CLAUDE.local.md` 는 **의도된 상태**로 문서에 기록되어야 한다.
`main` 의 커밋본이 폐기 모델(변종 1)이므로, 워킹본이 develop 판인 이상 main 대비로는 영원히
modified 로 읽힌다.

**REQ-CLPM-004 (When, event-detected)** — 누군가 그 ` M ` 표식을 「또 갈라졌다」로 읽고
`git restore CLAUDE.local.md` 류로 정리하려는 상황이 감지되면, 문서는 **그 행위가 폐기 모델로의
회귀임**을 그 자리에서 읽히도록 경고해야 한다. 각주가 아니라 요구 수준의 문장이어야 한다.

**REQ-CLPM-005 (Ubiquitous)** — 문서는 `main` 의 커밋본이 **폐기된 제3의 모델**이라는 사실을
기록해야 한다. 이것이 없으면 다음 사람이 main 사본을 정본으로 인용한다 — 2026-09-07 에 일어난
일의 재현이다.

**REQ-CLPM-006 (Ubiquitous)** — **이력에 커밋된 적 없는 워킹 사본은 정본으로 인용될 수 없다.**
문서는 이 일반 규칙을 담아야 한다. 축 1 의 결함이 특정 파일의 사고가 아니라 인용 규율의
구멍이기 때문이다.

**REQ-CLPM-007 (Ubiquitous)** — 이 SPEC 은 **문서만** 고친다. `.go` 파일을 건드리지 않는다.

---

## §C 수용 기준

Tier S 이지만 리드 지시에 따라 `acceptance.md` 를 별도 파일로 둔다.
AC 는 `acceptance.md` 를 정본으로 한다.

---

## §D 제약

- **문서 전용.** `.go` 파일 0 변경. 이 SPEC 은 Go 코드를 만들지 않는다.
- **정정은 이 워크트리의 `CLAUDE.local.md` 한 곳에서만.** 이 트리의 base 가 `origin/develop`
  tip 이므로, 여기서 고치고 develop 으로 병합하는 것이 「정본 한 곳」 경로다.
- **범위 제한 검증은 왼쪽 끝을 읽는 시점에 재유도한다**:
  `CARD_BASE=$(git merge-base origin/develop HEAD)`.
  리터럴 SHA(`bce6d7e08`)는 날짜 붙은 앵커일 뿐 범위의 왼쪽 끝이 아니다.
- **부재 주장은 `/usr/bin/grep`** 으로 잰다 — 이 셸의 `grep` 은 조용히 건너뛰는 ugrep 래퍼다.
- **부재 형태 AC 는 판별 대조군을 동반한다.** 죽은 필터는 빈 프로브를 공짜로 만족시킨다.
- **0 대조군은 「변경 없음」이 아니라 「측정 불가」로 보고한다.**
- 이 브랜치의 커밋 메시지는 `t531` 을 운반한다. 커밋·push·run-phase 진입은 하지 않는다.
- 시간 예측 금지. 사용자 대면 산출물은 한국어, 식별자·경로·명령·플래그는 원문 그대로.

---

## §E 재발 방지에 대한 판단

기계 가드는 **범위 밖**이다(§F). 이유를 여기에 남긴다 — 다음 사람이 「왜 가드를 안 만들었나」를
다시 묻지 않도록.

브랜치 대 브랜치 diff 검사는 **변종 2 를 원리상 볼 수 없다.** 가해자가 커밋된 블롭이 아니라
워킹 사본이었기 때문이다. 실효 있는 형태는 「세션에 로드되는 파일 각각이 추적 상태에서 깨끗한가」
쪽인데, 그 설계는 **재보지 않은 전제** 위에 선다 — 워크트리가 40개 이상이고 각각 워킹 사본을
가지며, 어느 좌표를 검사할지가 그 전수(§G-1)에 달렸다.

전수를 재기 전에 가드를 설계하면 손으로 열거한 판별식이 자기 결함을 재생산한다.

---

## §F 범위 밖

### Out of Scope — 기계 가드 (재발 방지 자동화)

- 세션 로드 파일의 추적-상태 청결성 검사기 설계·구현. **후속 카드로 발행한다** — 이 SPEC 에서
  설계하지 않는다.
- 브랜치 간 `CLAUDE.local.md` 자동 diff 검사. 변종 2 를 볼 수 없으므로 형태 자체가 틀렸다.
- 워크트리 40+ 전수 조사(census). 가드 설계의 선행 전제이고, 그 자체가 별개 작업이다.

### Out of Scope — 백업 파일의 내용

- `CLAUDE.local.md.primary-uncommitted-backup-20260908` 을 **분석하지 않는다.** diff 하지 않고,
  삭제된 116줄이 무엇이었는지 요약하지 않고, 그중 무엇을 되살리자고 제안하지 않는다.
- 백업은 **처분을 되돌릴 수 있게** 보존된다. 그것이 전부다. 경로와 sha256 만 기록한다.

### Out of Scope — 이미 집행된 처분의 재심

- primary 워킹본을 develop 정본으로 교체한 결정. 운영자 결재로 집행됐고 §A.3 에 기록됐다.
  이 SPEC 은 그것을 다시 다투지 않는다.

### Out of Scope — Go 코드

- `internal/**`, `pkg/**`, `cmd/**` 의 어떤 변경도. 이 SPEC 은 문서 수리다.

### Out of Scope — 2026-09-07 사건의 재현

- 레인 12곳 배포와 lane-3 · lane-4 정지의 재현·사후분석. 카드 본문의 서술이고 측정된 바 없다(§G-3).

---

## §G 미검증 간극 (열어 둔 채로 둔다)

1. **워크트리 40+ 전수를 재지 않았다.** 「세 변종」은 **읽은 세 좌표에 대한 진술**이지 열거가
   아니다. 다른 워크트리들이 각자의 워킹 사본을 갖고 있다. 네 번째 이상의 변종이 있을 수 있다.
2. **미커밋 편집본의 저자와 시점을 모른다.** 커밋되지 않았으므로 git 이력에 기록이 없다.
3. **2026-09-07 사건(레인 12곳, lane-3 · lane-4 정지)은 재현되지 않았다.** 카드 본문에서 읽은
   서술이며 여기서 측정한 것이 아니다.
4. **「변종 2 가 당시 인용된 것」은 추론이다.** 텍스트 일치에서 나왔고, 그 배차를 관측한 것이
   아니다. **[추론]** 표지를 §A.1 에 달아 두었다.

---

## §H 교차 참조

- `.moai/reports/t531/premise-remeasure.md` — 이 SPEC 이 인용하는 모든 수치의 원천
- `CLAUDE.local.md` §4.1 — 수리 대상 절
- `.claude/rules/local/gitflow-lane-protocol.md` — 레인 운영 규칙(정정 후 정합성 확인 대상)
