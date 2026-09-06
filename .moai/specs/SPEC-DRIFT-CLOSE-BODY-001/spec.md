---
id: SPEC-DRIFT-CLOSE-BODY-001
title: "본문 선언 close 인식 — subject가 못 담은 close로 생기는 drift 오탐 차단"
version: "0.4.0"
status: completed
created: 2026-09-03
updated: 2026-09-05
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/spec"
lifecycle: spec-anchored
tags: "drift, lifecycle, close-convention, false-positive, internal-spec"
tier: S
amendment_of: SPEC-DRIFT-CLOSE-BODY-001
---

# SPEC-DRIFT-CLOSE-BODY-001 — 본문 선언 close 인식

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-03 | manager-spec | 최초 작성 (Tier S). 카드 t410. 근거는 `.moai/reports/t410/discovery.md`(R1~R4) — 카드 문면이 아니라 재현 |
| 0.2.0 | 2026-09-03 | manager-spec | **초판 술어 2건이 실측으로 기각됐다.** ① 후보 게이트를 subject `closeInfixMatch`로 잡은 것 — 실측 close 커밋 6개 중 4개만 통과하고 **확정 대상 `e979a4d13`(`mx-phase close` ≠ `mx-phase audit-ready`)이 탈락**해 AC-DCB-004가 구조적으로 실패했다(§5.4 신설, REQ-DCB-003 재작성). ② 본문 술어를 "줄 선두 SPEC-ID"로만 잡은 것 — squash 하위 subject 모양(전체 close 커밋의 다수)을 통째로 놓치고, 72열 접힘으로 생긴 연속 줄(`51d18d3fe`)을 오탐한다. 두 모양(A `<ID>:` / B conventional-commit subject) + `:` 요구로 정정하고 실측 10줄 픽스처를 §5.3에 못박음. 조사의 TIGHT 15에서 오탐 3건 추가 확인(`fb8aff006`×2 · `80dea9684`) — §6에 반영 |
| 0.3.0 | 2026-09-03 | manager-spec | **plan-audit(t410, PASS-WITH-DEBT 0.80) 차단 결함 3건 + 문서 내 모순 2건 상환.** **D1** AC-DCB-007 1항의 `^[-+].*^status:`가 두 grep 방언에서 각각 과대매칭/공허 — 이 머신 실측(BSD grep, 픽스처 6줄)에서 산문 1줄을 포함한 3건을 매치했고 정정 술어 `^[-+]status:`는 실제 frontmatter 변경 2줄만 매치했다. 술어 교체 + RED 실측 의무(임시 편집 → 기대 `2` → 복귀) + 빈-diff 0 거부(기준 SHA 명시)를 명문화. **D2** 모양 B의 스캔 의미 미정의 — `7beda68a5` 본문 실측 재현(`SPEC-WORKTREE-ENTRY-STRATEGY-001` 명명 줄 8개 중 `completed`는 117번 1개뿐)으로 "첫 모양-B 줄 반환" 뮤턴트가 기존 10줄 픽스처를 통과함을 확인. REQ-DCB-003에 `ClassifyPRTitle == completed` 자격 + 전수 훑기를 명시하고, 픽스처를 순서 있는 11·12 쌍으로 확장(10줄 → 12줄), AC-DCB-002 뮤테이션 의무에 3번째 뮤턴트 추가. **D3** REQ-DCB-005 대응 AC 부재 — §5.1의 "구조적 보장" 논거가 모양 B 도입으로 무효화됨을 §5.1에 기록하고, AC를 새로 만들지 않고 AC-DCB-003에 (d) 케이스로 병합(Tier S AC 상한 8 유지, 7개 그대로). **D5** §5 후보표 행 B의 `closeInfixMatch` 근거와 §5.2의 낡은 포인터를 §5.4 기각 결과에 맞춰 정정. **D7** AC-DCB-002 Given에 fallback 도달 전제(frontmatter `completed` + 1차 워크 비-terminal 비-`completed`) 추가. **부채로 남김**: D4(REQ-DCB-002가 `inMemImpliedStatus` 오류 경로까지 주장 — 효과는 무해, REQ 2개+§4 동시 수정이 필요해 run-phase 원장 처리), D6(plan.md Tier 파일 모집단·초과 시 동작). 아울러 frontmatter `version:`이 0.1.0에 머물러 HISTORY(0.2.0)와 어긋나 있던 것을 함께 바로잡는다 |
| 0.4.0 | 2026-09-05 | manager-spec | **in-place amendment (카드 t484).** t410 sync-audit F1·F5 + plan-audit D4·D6 상환 — 넷 모두 트리 `a825183dd`에서 재측정되어 생존함을 확인(판정서 `.moai/reports/t484/verdict.md`). D4: REQ-DCB-002 문면을 "상태를 **반환했고**"로 좁히고 오류 경로를 §4에 신설. F5: §5.2에 모양 A 꼬리 무제약 잔여(반대 방향)를 기록. F1-②: REQ-DCB-007·AC-DCB-007을 manager-spec 재위임 경유로 개정. D6: plan.md §A에 파일 수 모집단 정의와 범위 안 초과 시 동작을 추가. 사유·범위·간극 조정은 아래 `## Amendments` |

## Amendments

**0.4.0 (2026-09-05, 카드 t484) — in-place amendment**

- **직전 완료 버전**: 0.3.0 (`status: completed`)
- **prior_completed_sha**: `c1a389036` — progress.md §E.4 `sync_commit_sha`에서 검증했다(해당 파일은 읽기 전용 입력이다)
- **사유 (rationale)**: 카드 t484에서 t410 sync-audit F1·F5와 plan-audit D4·D6를 상환한다. 네 항목 모두 문서 층 결함이며 트리 `a825183dd`에서 전부 재측정되어 여전히 생존함이 확인됐다(판정서: `.moai/reports/t484/verdict.md`). D4·D6을 당시 run-phase가 보류한 사유는 `spec.md`/`plan.md` 본문이 manager-develop에게 금지된 표면이었기 때문이고, 이 amendment가 그때 약속된 manager-spec 재위임 경로다. F1의 지시 결함(AC가 run-phase에게 다른 SPEC 편집을 직접 지시)도 같은 경로로 개정한다.
- **범위 (scope)**: spec.md REQ-DCB-002, REQ-DCB-007, AC-DCB-007, §4(오류 경로 신설), §5.2(반대 방향 잔여); plan.md §A(모집단 정의·초과 시 동작).
- **간극 조정 기록**: HISTORY 0.3.0 D1 행은 "REQ 2개+§4"라 적었고 원장 스케치(`run-evidence.md:480`)는 REQ-DCB-002+§4만 명시했다. 이 amendment가 확정한 편집 대상 REQ는 **REQ-DCB-002와 REQ-DCB-007** 두 개다 — 0.3.0 시점에는 F1-②(REQ-DCB-007 개정)가 알려지지 않았다. REQ-DCB-006은 문면 수정이 불필요했다: 오류 경로가 그 요구(동작 보존)의 결과라는 사실은 §4의 새 항목이 교차 참조로 담는다.

## 1. 배경

### 1.1 결함

`moai spec drift`는 실제로 닫힌 SPEC을 `in-progress`로 추론해 DRIFT로 보고한다. 확정된 사례 하나:

```
SPEC-V3R6-SESSION-HANDOFF-AUTO-001  completed  in-progress  DRIFT
```

이 SPEC은 `e979a4d13`에서 실제로 닫혔다. 그 커밋의 frontmatter 변경(`status: implemented → completed`, `updated: 2026-06-02`)이 커밋 날짜와 정확히 일치하므로, frontmatter가 앞서 나간 것이 아니라 **판정기가 close를 못 본 것**이다.

### 1.2 기제 — 원인은 하나다

판정기의 2단 구조가 원인을 만든다(`internal/spec/drift_index.go` `inMemImpliedStatus`).

- **1단(후보 창)**: 커밋의 **전체 메시지**(subject + body)가 SPEC-ID를 포함하면 후보다. `git log --grep`의 재현이다.
- **2단(재필터)**: 후보의 **subject**만 보고 `shouldSkipCommitTitle` → `commitMatchesSPECID`(정확 토큰) → `ClassifyPRTitle`을 건다. 첫 유의미 분류가 승리한다.

진짜 close 커밋 `e979a4d13`의 subject는 `chore(SPEC group C): Mx-phase close (...)`다. 완전한 SPEC-ID 토큰이 없으므로 `ExtractSPECIDs(subject)`가 빈 집합을 내고 2단에서 탈락한다. 워커는 계속 내려가 같은 SPEC의 더 오래된 `docs(...)` 커밋을 만나 거기서 `in-progress`를 집는다.

축자 추적(`.moai/reports/t410/r1-walker-trace.log`):

```
SKIP  e979a4d13  [subject lacks id]  chore(SPEC group C): Mx-phase close (...)   extracted=[]
ADOPT 97a36b5a2  [status=in-progress]  docs(SPEC-V3R6-SESSION-HANDOFF-AUTO-001): /moai mx Step C
```

### 1.3 앞선 카드의 진단 절반이 반증됐다

t382(`SPEC-ERA-H3-NARROWING-001`)는 원인을 둘로 적었다. 재현은 그중 하나를 **반증**했다.

| t382가 적은 원인 | 재현 결과 |
|---|---|
| ① `--grep`이 본문을 매치해 **다른 SPEC의 plan 커밋**이 최신으로 채택된다 | **거짓.** 해당 커밋 `7cffb9717`은 2단 subject 필터에서 정상적으로 걸러진다(`extracted=[SPEC-HANDOFF-AUTORESUME-001]`). 본문 매치는 후보 창을 넓힐 뿐이고 본문 전용 매치는 다음 필터가 전부 떨어뜨린다 |
| ② 결합 범위 close subject에 완전한 ID가 없어 안 보인다 | **참이고, 유일한 작동 원인이다** |

이 구분이 수리 공간을 바꾼다. ①이 참이었다면 "본문을 안 보게 하면" 되지만, 실제로는 **본문을 조건부로 더 읽는 것**이 수리 방향이다.

`SPEC-ERA-H3-NARROWING-001`은 `completed`이고 t382의 착지는 유효하다. 이 SPEC은 그것을 소급 재판정하지 않는다 — 원인 서술 정정은 해당 SPEC의 HISTORY 한 줄로만 남긴다(REQ-DCB-007).

### 1.4 위반 subject 3종과, 규약의 실효성

관측된 형태:

| 형태 | 예 |
|---|---|
| 명시적 결합 범위 | `chore(SPEC group C): Mx-phase close` · `docs(specs): batch sync-phase close — 5 B-grade SPECs` |
| 범위 없는 산문 subject | `Close out 2 SPECs with 3-phase lifecycle completion (doc-only)` |
| 단일 범위인데 형제를 함께 닫음 | `docs(SPEC-INTERNAL-TEST-001): sync-phase artifacts + 3-phase close` (본문에서 `-002`도 닫음) |

`.claude/rules/moai/development/spec-frontmatter-schema.md` § Close-subject full-ID mandate가 이 형태를 이미 **금지한다**. 금지가 있는데도 이력에 남아 있다는 것은, 그 규약이 기계적으로 강제되지 않으며 **이미 착지한 이력은 규약으로 되돌릴 수 없다**는 뜻이다. 판정기는 규약이 지켜졌다고 가정할 수 없다.

### 1.5 무게중심

이 SPEC의 무게중심은 **판정기가 close 증거를 읽는 방식**이다. 개별 drift 행의 수리도, close 규약의 강제도 아니다. 행은 결과로 바뀌고, 규약 강제는 §4에서 범위 밖으로 둔다.

## 2. 요구사항 (GEARS)

**REQ-DCB-001** — drift 판정기는, close 커밋의 **본문**에 close가 선언되고 그 커밋의 subject가 대상 SPEC-ID를 담지 못한 경우에도, 해당 SPEC의 git 함의 상태를 `completed`로 추론해야 한다(shall).

**REQ-DCB-002** — While frontmatter가 `completed`이고 1차 워크가 상태를 **반환했고** 그 값이 `completed`도 terminal 상태도 아닌 동안, 판정기는 본문 선언 close 조회를 수행해야 한다. 그 밖의 상태 조합에서는 조회하지 않는다(FALLBACK-ONLY). 오류를 반환한 갈래는 "반환하지 않은" 경우다 — 그곳에서는 조회가 발화하지 않는다(§4의 오류 경로 항목, 0.4.0 amendment).

**REQ-DCB-003** — Where 후보 커밋의 subject가 대상 SPEC-ID를 담지 않을 때에 한해, 판정기는 그 커밋의 본문에서 §5.3이 정의한 두 모양(A: `<full-ID>:` 줄 선두 / B: conventional-commit subject 줄)의 줄만 close 선언으로 취급해야 한다. 모양 B는 기존 필터 체인이 그 줄에 대해 `ClassifyPRTitle == completed`를 낼 때에**만** close 선언이며, 그 밖의 분류(`implemented`·`in-progress`·`draft`·`skip`·`unknown`)는 close 선언이 아니다. subject 쪽 close 신호를 후보 게이트로 쓸 경우 §5.4의 세 조건을 만족해야 하며, `closeInfixMatch` 단독 사용은 금지한다(측정으로 기각 — 확정 대상 `e979a4d13`을 떨어뜨린다). 그리고 조회는 자격 있는 줄을 만날 때까지 **본문 전체를 훑어야 하며**, 자격 없는 줄에서 판정을 종료해서는 안 된다 — 자격 있는 줄이 하나라도 있으면 close 선언이고 하나도 없으면 무판정이다(§5.3).

**REQ-DCB-004** — 판정기는 SPEC-ID를 **언급만 하는** 본문 줄로 drift를 해제해서는 안 된다(shall not). 최소한 `depends_on:` / `related:` 형태의 선행 키가 붙은 줄은 close 선언이 아니다.

**REQ-DCB-005** — 판정기는 어떤 커밋의 **본문**으로부터 `completed` 이외의 상태를 채택해서는 안 된다(shall not). 본문 조회의 출력은 `completed` 또는 무판정 둘 중 하나다.

**REQ-DCB-006** — 판정기의 1차 워크(2단 subject 재필터)와 기존 combined-scope prefix fallback의 동작은 변경되지 않아야 한다(shall). 본문 조회는 오직 추가 경로다.

**REQ-DCB-007** — When run-phase가 drift 표의 어떤 행의 판정을 바꾸면, 구현은 그 행을 **행별로** close 커밋 SHA와 함께 원장에 기록해야 한다. 그리고 t382의 원인 서술 정정은 `SPEC-ERA-H3-NARROWING-001`의 HISTORY 한 줄로만 남기고, 그 SPEC의 `status`와 AC 판정은 건드리지 않는다. 그 정정 편집 자체는 run-phase가 직접 수행하지 않는다(0.4.0) — run-phase는 blocker를 보고하고 오케스트레이터가 manager-spec에 재위임하며, 다른 SPEC의 본문(HISTORY 행과 `version:`/`updated:` 포함)은 manager-spec이 편집한다.

## 3. 인수 기준 (Tier S — 인라인)

모든 판정은 이 워크트리에서 빌드한 바이너리로 수행한다(`go build -o /tmp/moai-t410 ./cmd/moai`). 설치본을 쓰지 않는다 — 도구 출처 귀속(`verification-claim-integrity.md` §2.2). 코퍼스 측정은 전부 `--no-cache`로 수행하고, **잰 트리 SHA와 잰 브랜치를 함께 적는다**.

---

**AC-DCB-001 — 본문 선언 close가 인식된다 (단위, RED→GREEN)**

- **Given** 커밋 색인 픽스처에 subject `chore(SPEC group C): Mx-phase close (status implemented→completed)`, 본문에 `- SPEC-FIX-ALPHA-001: Mx verdict EVALUATE-PASS` 줄을 갖는 커밋과, 같은 SPEC의 더 오래된 `docs(SPEC-FIX-ALPHA-001): ...` 커밋이 있고, frontmatter가 `completed`일 때
- **When** `go test ./internal/spec/ -run TestDriftCloseBody -count=1` 를 실행하면
- **Then** rc 0. 수리 **이전**에는 같은 명령이 rc 1로 실패해야 한다(RED를 실측하고 축자 출력을 원장에 남긴다).
- **공허 방지**: 같은 테스트가 **먼저** 1차 워크가 실제로 `in-progress`를 낸다는 것을 단언한다. 이 단언이 없으면 픽스처가 애초에 결함을 재현하지 않아도 통과한다.

---

**AC-DCB-002 — 언급은 close가 아니다 (단위, 실측된 반례)**

- **Given** 실측 커밋 `a83934d55`를 본뜬 픽스처 — subject `feat(SPEC-HIER-BETA-001): ... (Tier M, 3-phase close)`(close-infix **있음**), 본문에 `depends_on: SPEC-DEP-GAMMA-001 (completed)` 줄. 그리고 이 케이스가 fallback에 **도달하도록**, `SPEC-DEP-GAMMA-001`의 frontmatter는 `completed`이고 1차 워크는 그 SPEC에 대해 비-terminal 비-`completed` 상태를 낸다(REQ-DCB-002의 발화 조건 — 이 전제가 없으면 케이스가 게이트 앞에서 멈춰 어떤 뮤턴트도 죽이지 못한다)
- **When** 같은 테스트 명령을 실행하면
- **Then** `SPEC-DEP-GAMMA-001`의 drift는 **해제되지 않는다**.
- **판정 모집단은 §5.3의 실측 12줄 전부다.** 4줄(선언 2 + 언급 2)만 넣고 통과시키지 않는다 — 12줄은 서로 다른 함정을 하나씩 담고 있다(줄바꿈 접힘 8, `:` 부재 10, 목록 표지 뒤 다른 키 9, 비-close 모양 B 선행 11).
- **공허 방지 (뮤테이션 의무 — 뮤턴트 3종)**: 아래 셋을 각각 실행하고 축자 출력과 rc를 원장에 남긴다. 하나라도 죽지 않으면 이 AC는 아무것도 주장하지 않는 것이므로 술어 또는 픽스처가 잘못된 것이다.
  1. 술어를 `strings.Contains(body, specID)`로 바꾼 뮤턴트 — **이 케이스(5번 줄)가 반드시 실패**해야 한다.
  2. 모양 A의 `:` 요구를 제거한 뮤턴트 — **8번 줄이 반드시 실패**해야 한다.
  3. 첫 모양-B 줄에서 반환하는 뮤턴트(전수 훑기 대신 조기 종료) — **11·12번 쌍이 반드시 실패**해야 한다. 이 뮤턴트는 1~10번만으로는 죽지 않는다(§5.3).

---

**AC-DCB-003 — FALLBACK-ONLY 게이트와 출력 제약이 유지된다 (단위)**

입력 게이트(언제 발화하는가)와 출력 게이트(무엇을 낼 수 있는가)는 같은 좁힘의 두 면이므로 한 AC로 판정한다.

- **Given (입력)** 본문 선언 close 증거가 존재하는 동일 픽스처에서 frontmatter가 (a) `in-progress` (b) `superseded` (c) `draft` 인 세 경우
- **Given (출력, REQ-DCB-005)** (d) frontmatter가 `completed`이고 1차 워크가 비-terminal 비-`completed`를 냈으며, 본문에는 대상 SPEC의 **비-close 모양-B 줄만** 있는 경우 — 입력은 §5.3의 11번 줄만 담고 12번 줄은 담지 않는다(11·12 쌍에서 12를 뺀 것)
- **When** 같은 테스트 명령을 실행하면
- **Then** (a)(b)(c)는 git 함의 상태가 1차 워크의 결과 그대로이며 `completed`로 덮이지 않는다 — 본문 조회는 frontmatter가 `completed`일 때만 발화한다. (d)도 git 함의 상태가 1차 워크의 값 그대로다 — 본문의 모양-B 줄이 `implemented`를 내더라도 그 값이 채택되지 않는다. 즉 본문 조회의 출력은 `completed` 또는 무판정 둘뿐이다.
- **(d)가 필요한 이유**: §5.1은 REQ-DCB-005를 "fallback의 출력이 `completed` 하나뿐"이라는 **구조적** 보장으로 논증했다. 그 논거는 §5.3이 모양 B를 도입하기 전에는 옳았다 — 지금은 본문 줄이 `ClassifyPRTitle`에 먹여지고 그 함수는 `implemented`·`in-progress`·`draft`를 반환한다(`transitions.go:132-177`). 보장은 이제 "구현이 비-`completed` 결과를 버린다"에 의존하므로, 구조적 주장이 아니라 측정으로 확인해야 한다.
- **공허 방지**: (d)의 1차 워크가 `completed`를 내면 이 케이스는 아무것도 구별하지 못한다 — 1차 워크가 실제로 비-`completed`를 냈음을 (a)와 같은 방식으로 **먼저 단언**한다.

---

**AC-DCB-004 — 확정된 그 행이 실제로 사라진다 (코퍼스)**

- **Given** 이 워크트리, 판정기가 보는 브랜치(`cachedMainBranch()` → `main`)
- **When** 수리 전후로 각각 `/tmp/moai-t410 spec drift --no-cache | grep SPEC-V3R6-SESSION-HANDOFF-AUTO-001` 을 실행하면
- **Then** 이전에는 `... completed  in-progress  DRIFT`, 이후에는 그 행이 DRIFT가 아니다.
- **공허 방지**: "이후에 행이 없다"로 판정하지 않는다 — 행이 존재하면서 DRIFT가 아님을 본다. 그리고 **이전 상태를 인용하지 않고 이 트리에서 재측정**한다. 양쪽 출력의 트리 SHA를 원장에 적는다.

---

**AC-DCB-005 — 표 전체에서 판정이 바뀐 행이 전수 정당화된다 (코퍼스)**

- **Given** 동일 트리·동일 브랜치에서 수리 전 표와 수리 후 표를 각각 파일로 받아 둔 상태(`spec drift --no-cache`)
- **When** 두 표를 SPEC-ID 기준으로 대조하면
- **Then** 다음 셋을 모두 만족한다.
  1. **비-DRIFT → DRIFT로 바뀐 행이 0건이다.** (회귀 없음)
  2. DRIFT → 해제로 바뀐 **모든** 행에 대해, 그 행을 해제시킨 close 커밋 SHA와 그 커밋의 subject·본문 해당 줄을 원장에 **행별로** 적는다.
  3. 각 해제 행의 본문 줄이 언급이 아니라 close 선언임을 사람이 판단해 적는다.
- **수를 미리 못박지 않는다**: 조사 단계의 LOOSE 33 / TIGHT 15는 프로브의 경계값이고, TIGHT는 알려진 오탐(`SPEC-AUTONOMY-TIERS-001`)을 최소 1건 포함한 **상한**이며 하한은 확정돼 있지 않다. 실제 수는 이 AC의 **결과로** 정해진다. 어떤 특정 수치를 기대값으로 삼는 판정은 이 AC를 위반한다.
- **공허 방지**: 해제 행이 0건이면 이 AC는 실패다(AC-DCB-004가 최소 1건을 보장하므로, 0건은 대조가 잘못된 것이다).

---

**AC-DCB-006 — 기존 가드가 전부 살아 있다 (회귀)**

- **When** `go test ./internal/spec/... -count=1` 을 실행하면
- **Then** rc 0. 특히 다음 가드가 통과한다: `drift_chore_skip_test.go`(AC-LSCSK-003), `drift_specid_grep_test.go`(LSGF-001 단어경계), `drift_combined_scope_test.go`(기존 prefix fallback 3-gate), `drift_characterization_test.go`·`drift_seam_test.go`(record 동등성).
- **공허 방지**: 실행 목록에 위 파일들이 실제로 포함됐음을 `-v` 출력의 테스트 이름으로 확인한다. 셀렉터 0매치는 초록이 아니다.

---

**AC-DCB-007 — t382 정정이 HISTORY로만 남는다 (문서, 기계 판정)**

- **When** 아래 명령을 실행하면

  ```
  git diff -U0 -- .moai/specs/SPEC-ERA-H3-NARROWING-001/spec.md | grep -cE '^[-+]status:'
  ```

- **수행 채널 (0.4.0 amendment)**: 이 정정 편집은 manager-spec이 오케스트레이터 재위임을 거쳐 수행한다. run-phase는 blocker를 보고할 뿐 다른 SPEC의 본문을 직접 편집하지 않는다 — 아래 세 판정은 그렇게 수행된 편집의 결과에 그대로 적용된다.
- **Then** 다음을 모두 만족한다.
  1. 위 명령의 출력이 **`0`**이다 — `status: completed`가 그대로다.
  2. 변경은 HISTORY 표의 새 행 1개와 `version:`/`updated:` 두 필드에 국한된다.
  3. 새 HISTORY 행이 `.moai/reports/t410/r1-walker-trace.log` 경로를 근거로 인용한다(`grep -c 'r1-walker-trace' ...` ≥ 1).
- **술어에 `^`는 줄 선두 하나뿐이다.** 패턴 중간에 `^`를 두면(`^[-+].*^status:`) 판정이 grep 방언에 따라 갈린다 — BSD grep은 중간 `^`를 빈 문자열로 소거해 본문 어디서든 `status:`를 언급하는 산문 줄까지 매치하고(과대매칭), GNU grep ERE에서는 만족 불가능한 앵커라 항상 0건을 내 무조건 통과한다(공허). 어느 쪽도 "변경된 `status:` 프론트매터 줄이 없다"가 아니다. 위 술어는 두 방언에서 같은 판정을 내며, `+++`/`---` diff 헤더와 "status" 문자열을 담은 HISTORY 표 행(`| ... |`)을 매치하지 않는다.
- **공허 방지 ① — RED 실측 의무**: 초록을 보고하기 전에 붉은 것을 본다. `SPEC-ERA-H3-NARROWING-001/spec.md`의 `status: completed`를 일시적으로 `status: implemented`로 바꾸고 같은 명령을 실행해 출력이 **`2`**(`-status:` 한 줄 + `+status:` 한 줄)임을 확인한 뒤 되돌린다. 세 실행(초록 → 붉음 → 되돌린 초록)의 축자 출력을 원장에 남긴다. RED를 본 적 없는 초록은 아무것도 주장하지 않는다.
- **공허 방지 ② — 빈 diff로 나온 0을 인정하지 않는다**: 판정 시점에 t382 편집이 이미 커밋됐다면 워킹트리가 깨끗해서 0이 나온다. 그 경우 `git diff -U0 <편집 직전 SHA> -- <경로> | grep -cE '^[-+]status:'` 로 재측정하고, 잰 기준 SHA를 원장에 적는다. 같은 diff에 대해 `grep -cE '^[-+]'` 가 0이 아님(= 대조 대상이 실재함)을 함께 보인다.
- 근거: 이 정정은 `status:`를 바꾸지 않으므로 amendment 절차 대상이 아니며, 수행 주체는 manager-spec이다(`spec-frontmatter-schema.md` § Non-transition frontmatter corrections — 소유자 `manager-spec`). 0.4.0 amendment는 지시 문면을 이 근거의 인용 규약과 일치시켰다 — 초판은 run-phase에게 직접 편집을 지시함으로써 자기 근거와 어긋났다(t410 sync-audit F1).

## 4. 범위 밖

### Out of Scope — 기존 close 커밋의 백필

조사 후보 C(SPEC별 full-ID + close-infix를 갖는 합성 커밋을 쌓기)는 채택하지 않는다. 판정기를 안 고치므로 같은 형태가 다시 들어오면 같은 오탐이 재발하고, 합성 커밋이 이력에 영구히 남으며, 작업량이 확정되지 않은 건수(§3 AC-DCB-005)에 비례한다.

- 새 `chore(SPEC-XXX-NNN): ... close` 백필 커밋을 만들지 않는다
- 기존 close 커밋의 메시지를 재작성하지 않는다

### Out of Scope — `inMemImpliedStatus` 오류 경로

`inMemImpliedStatus`가 오류를 내면 `DetectDrift`는 ① 블록(본문 선언 close 조회)에 도달하기 전에 `continue`한다. 따라서 오류 갈래에서는 본문 조회가 발화하지 않는다 — REQ-DCB-002(0.4.0)의 "상태를 **반환했고**" 문면이 이 갈래를 요구 범위에서 제외한다. 이는 REQ-DCB-006(1차 워크 동작 보존)의 결과이며 방향은 보수적이다: 오류 창은 놓침(무판정)이지 거짓 해제가 아니다. 코드를 바꾸지 않는다 — 이 부채(plan-audit D4)의 수리는 요구 문면을 좁힌 이 amendment 자체다.

- 오류 갈래에서 조회가 발화하도록 `drift.go`의 `continue` 배치를 바꾸지 않는다
- 오류 케이스에 대한 새 테스트를 추가하지 않는다

### Out of Scope — close 규약의 기계적 강제

조사 후보 D(향후 위반을 훅/린트로 막기)는 별도 소관이다. 이 SPEC이 판정기를 관대하게 만드는 것과 규약이 그 형태를 여전히 금지하는 것은 **양립한다** — 판정기가 읽을 수 있게 되는 것이 그 subject 형태를 합법화하지는 않는다. 이 긴장은 의도된 선택이며 누락이 아니다.

- close subject를 검사하는 훅·린트 규칙을 만들지 않는다
- `spec-frontmatter-schema.md` § Close-subject full-ID mandate의 문면을 완화하지 않는다

### Out of Scope — drift 판정 기준 브랜치

판정기는 `cachedMainBranch()` → `main`을 본다. 이 저장소의 작업은 `develop`에 착지하므로 표의 대부분은 "close가 아직 main에 없음"이며 이 카드의 결함이 아니다(§6). 기준 브랜치를 바꾸는 것은 별개 축이다.

- `cachedMainBranch()`의 결정 규칙을 바꾸지 않는다
- develop 기준 판정 옵션을 추가하지 않는다

### Out of Scope — `ClassifyPRTitle`의 분류표

`docs(...)`를 `in-progress`로 분류하는 것이 그 자체로 옳은지는 판정하지 않는다. 이 행에는 무관하다 — `docs`를 빈 상태로 만들어도 워커는 더 내려가 `test(...)`를 만나며 `completed`는 여전히 안 나온다.

- `transitions.go`의 prefix→status 표를 바꾸지 않는다

### Out of Scope — `SPEC-ERA-H3-NARROWING-001`의 재판정

t382의 착지는 유효하다. AC 판정, 상태, 본문 결론을 소급해 뒤집지 않는다. 원인 서술 정정은 HISTORY 한 줄이다(AC-DCB-007).

## 5. 설계 선택 — 왜 후보 B인가

조사가 후보 넷을 남겼다(`discovery.md` § 수리 후보). 판정:

| 후보 | 판정 | 근거 |
|---|---|---|
| A. 워커가 본문에서 ID를 추출하도록 **무조건** 확대 | **기각** | `e979a4d13`이 보이게 되는 대신 `7cffb9717`(다른 SPEC의 plan 커밋)도 보이게 되어, t382가 **상상만 했던** 오탐을 진짜로 만든다. LSGF-001이 막으려던 형태 그 자체다 |
| **B. close 선언 커밋에 한해 본문 조회** | **채택** | A의 위험을 열지 않으면서 `e979a4d13`을 살린다. 코드에 이미 FALLBACK-ONLY 조회 골격(`inMemCombinedScopeClose`)이 있어 같은 자리에 축을 하나 더 놓는 일이다. (초판은 `closeInfixMatch` 재사용도 근거로 들었으나 **§5.4에서 측정으로 기각됐다** — 그 술어는 확정 대상 `e979a4d13`을 떨어뜨린다. 채택 근거는 골격 재사용뿐이다) |
| C. 백필 커밋 | 기각 | §4 |
| D. 규약의 기계적 강제 | 범위 밖 | §4 |

### 5.1 B가 새 기제가 아니라 기존 기제의 확장인 이유

`internal/spec/drift.go`(①번 주석)와 `drift_index.go` `inMemCombinedScopeClose`에 **이미 같은 모양의 fallback이 착지해 있다**. 그것은 `chore(SPEC-CCSYNC): ... 3-phase close (CLAUDEMD + TOOLCAT)` 형태 — subject가 SPEC-ID에서 파생된 **scope-prefix**를 명명하는 경우 — 를 3-gate(FALLBACK-ONLY + close-infix + distinguishing-segment 단어경계)로 처리한다.

`chore(SPEC group C)`가 안 걸리는 이유는 명확하다: gate (a)가 `chore(spec-<prefix>)`를 요구하는데 `SPEC group C`는 어떤 SPEC-ID에서도 파생되지 않는 **임의 문자열**이기 때문이다. 즉 이 SPEC이 추가하는 것은 같은 자리에 놓이는 **세 번째 판정 축**이지, 새 골격이 아니다.

기존 combined-scope fallback에서는 이 배치가 REQ-DCB-005를 **구조적으로** 보장했다 — 그 경로의 출력이 `completed` 하나뿐이라 본문에서 `draft`나 `implemented`를 주워 올 자리가 없었다. **모양 B가 이 성질을 바꾼다.** 모양 B는 본문 줄을 `ClassifyPRTitle`에 그대로 먹이고 그 함수는 `implemented`·`in-progress`·`draft`도 반환하므로, 보장은 이제 구조가 아니라 "구현이 비-`completed` 결과를 버린다"는 동작에 의존한다. 그래서 REQ-DCB-005는 논증이 아니라 측정으로 지켜야 하며, AC-DCB-003 (d)가 그 측정이다.

### 5.2 이 선택이 지불하는 대가 — 정직하게

fallback은 frontmatter가 `completed`일 때만 발화하고 출력도 `completed`뿐이다. 따라서 **판정기가 frontmatter와 불일치할 여지가 그만큼 줄어든다**. 반대 방향 오류(frontmatter가 거짓 `completed`인데 본문 언급 한 줄로 무죄 방면되는 것)가 이 수리의 실질 위험이며, **REQ-DCB-003 / §5.3의 두 모양 술어**(모양 A의 줄 선두 + `:` 요구, 모양 B의 `ClassifyPRTitle == completed` 자격)와 AC-DCB-002의 뮤테이션 의무 3종이 그것을 겨눈다. REQ-DCB-004는 그 술어가 반드시 떨어뜨려야 할 최소 집합(`depends_on:` / `related:` 선행 키)을 못박는다.

모양 A 쪽에는 술어가 걷지 못하는 잔여가 하나 더 있다(0.4.0 — 카드 t484, 판정서 F5). 모양 A(`<ID>:` 줄 선두)는 **콜론 뒤 본문을 전혀 제약하지 않는다**. 그래서 subject에 close가 있는 커밋의 본문에 비-close 성격의 `<ID>:` 줄이 하나 있으면 그 SPEC이 해제될 수 있다. t410 감사의 코퍼스 전수 스캔은 이 형태를 18줄 찾았고 실제 오해제는 0건이었다 — 오늘의 0은 close 커밋 본문 작성 관행의 함수이지 술어의 성질이 아니다. `drift_index.go`의 코드 주석이 같은 내용을 이미 정직하게 적고 있다.

기존 combined-scope fallback이 이미 같은 성질을 갖고 있으므로 이 SPEC이 **새로운** 약점을 여는 것은 아니다. 다만 그 약점이 닿는 행 수를 늘린다. 이 사실을 기록해 두는 이유는, 나중에 술어를 넓히자는 제안이 왔을 때 그것이 무엇을 무르게 하는지 읽히게 하기 위해서다.

### 5.3 본문 줄 술어 — 실측 12줄로 그은 경계

가르는 성질은 "같은 줄에 `completed`가 있는가"가 **아니다** — 아래 두 줄 다 갖고 있다.

```
- SPEC-V3R6-SESSION-HANDOFF-AUTO-001: Mx verdict EVALUATE-PASS   ← close 선언 (e979a4d13)
depends_on: SPEC-AUTONOMY-TIERS-001 (completed)                  ← 언급     (a83934d55)
```

실측된 close 선언은 **두 모양**이다. 하나만 구현하면 나머지가 통째로 빠진다.

| 모양 | 예 | 판정 |
|---|---|---|
| **A. verdict 목록** | `- SPEC-V3R6-PROMPT-CACHE-001: Mx verdict EVALUATE-PASS` | 선행 목록 표지를 걷어낸 뒤 줄이 완전한 SPEC-ID로 시작하고 **바로 뒤가 `:`** |
| **B. squash 하위 subject** | `* docs(SPEC-GLM-KEY-INPUT-001): sync-phase artifacts — 3-phase close` | 표지를 걷어낸 줄이 그 자체로 conventional-commit subject — 기존 필터 체인(`shouldSkipCommitTitle` → `commitMatchesSPECID` → `ClassifyPRTitle`)에 그대로 먹이고, **그 체인이 `completed`를 낼 때에만** 선언 |

모양 B가 압도적으로 흔하다. squash 병합이 개별 커밋 subject를 본문에 옮겨 놓기 때문이고, 그 줄들은 **subject였다면 1차 워크가 정상적으로 채택했을** 줄이다. 그래서 B는 새 텍스트 술어를 만들지 않고 착지·검증된 기존 체인을 재사용한다 — 이 SPEC이 여는 판정 표면을 최소화하는 길이다.

**모양 B의 스캔 의미 — 전수 훑기이고, 자격은 `completed` 하나다.** 하나의 본문이 대상 SPEC-ID를 명명하는 모양-B 줄을 여럿 담는 것이 정상이며, 그중 대부분은 close가 아니다. 실측:

```
$ git show -s --format=%b 7beda68a5 | grep -n 'SPEC-WORKTREE-ENTRY-STRATEGY-001'
1:  * fix(...): M1 web auto-toggles default OFF (...)            → implemented   (close 아님)
15: * feat(...): M3a launcher L2 absolute-path resolver (...)     → implemented   (close 아님)
43: * docs(...): populate progress.md §E.2/§E.3 (...)             → in-progress   (close 아님)
53: * docs(...): M2-M6 doc-alignment (...)                        → in-progress   (close 아님)
108:* chore(...): backfill Round 2 run_commit_sha 1201680b3 (...)  → skip          (close 아님)
117:* docs(...): sync-phase artifacts — 3-phase close (...)        → completed  ★ 유일한 close
123:* chore(...): backfill sync_commit_sha 14027ffec (...)         → skip          (close 아님)
143: (산문) Close the F1 residual from SPEC-WORKTREE-ENTRY-...     → 모양 B 아님   (close 아님)
```

`ClassifyPRTitle`은 close-infix를 prefix 루프보다 **먼저** 본다(`transitions.go:149`). 따라서 117번만 `completed`를 내고 나머지 일곱 줄은 전부 비-close다. **첫 모양-B 줄에서 반환하는 구현은 이 SPEC의 진짜 close를 놓친다** — 그리고 `SPEC-WORKTREE-ENTRY-STRATEGY-001`은 조사의 TIGHT 목록에 들어 있으므로 가상의 경우가 아니다. 그래서 REQ-DCB-003은 전수 훑기를 요구하고, 픽스처 11번이 그 요구를 붉게 만드는 입력이다.

**`:` 요구는 장식이 아니다.** 실측 반례가 있다.

```
SPEC-INTERNAL-SECURITY-001 f3193bac8 / SPEC-HANDOFF-GOALFIX-001    ← 언급 (51d18d3fe)
```

앞선 줄의 괄호 안 열거가 72열에서 접혀 생긴 연속 줄이다. 완전한 SPEC-ID로 **시작**하지만 close 선언이 아니다. `:` 요구가 이것을 떨어뜨린다. 즉 "줄 선두"만으로는 부족하고, 줄바꿈 위치라는 서식 우연에 판정이 흔들린다.

#### 필수 픽스처 — 실측 12줄

run-phase는 아래 12줄을 픽스처로 못박고, 술어가 12/12를 맞추는지 본다. 이 목록이 이 SPEC의 술어 판정 기준이다.

| # | 줄 (발췌) | 출처 | 기대 |
|---|---|---|---|
| 1 | `- SPEC-V3R6-SESSION-HANDOFF-AUTO-001: Mx verdict EVALUATE-PASS` | e979a4d13 | 선언 (A) |
| 2 | `- SPEC-V3R6-PROMPT-CACHE-001: Mx verdict EVALUATE-PASS, 10 @MX:ANCHOR` | e979a4d13 | 선언 (A) |
| 3 | `* docs(SPEC-GLM-KEY-INPUT-001): sync-phase artifacts — 3-phase close` | 2f449e189 | 선언 (B) |
| 4 | `* chore(SPEC-V3R6-CODERABBIT-ADOPTION-001): sync-phase artifacts — 3-phase close (4be491a0b)` | 7beda68a5 | 선언 (B) |
| 5 | `depends_on: SPEC-AUTONOMY-TIERS-001 (completed)` | a83934d55 | 언급 |
| 6 | `... Depends on SPEC-WORKTREE-BRANCH-GUARD-001 and SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 (both completed).` | fb8aff006 | 언급 |
| 7 | `... 3 residual_debt items owned by follow-up SPEC-INTERNAL-TEST-002)` | 80dea9684 | 언급 |
| 8 | `SPEC-INTERNAL-SECURITY-001 f3193bac8 / SPEC-HANDOFF-GOALFIX-001` | 51d18d3fe | 언급 (줄바꿈 함정) |
| 9 | `- CHANGELOG [Unreleased]: Added entry for SPEC-GLM-KEY-INPUT-001` | 2f449e189 | 언급 |
| 10 | `- SPEC-GLM-KEY-INPUT-001         → d06771f07` | 2f449e189 | 언급 (`:` 없음) |
| 11 | `* fix(SPEC-WORKTREE-ENTRY-STRATEGY-001): M1 web auto-toggles default OFF (AutoCleanup+AutoMerge true→false)` | 7beda68a5 | 언급 (모양 B이나 `implemented` — 비-close) |
| 12 | `* docs(SPEC-WORKTREE-ENTRY-STRATEGY-001): sync-phase artifacts — 3-phase close (CHANGELOG + completed transition)` | 7beda68a5 | 선언 (B) |

5~8은 **측정으로 확인된 오탐**이다. 조사의 TIGHT 15에 5·6·7이 들어 있었다 — 즉 TIGHT는 "조치 대상"이 아니라 상한이다(§6).

**11·12는 순서가 있는 한 쌍이며, 같은 픽스처 커밋의 본문에 11이 12보다 앞서 놓여야 한다.** 둘 다 같은 SPEC-ID를 명명하는 모양 B이고 11은 비-close, 12는 close다(출처 `7beda68a5`의 실제 본문 순서 그대로 — 1번 줄과 117번 줄). 이 쌍이 전수 훑기 요구(REQ-DCB-003)를 붉게 만드는 유일한 입력이다: 첫 모양-B 줄에서 반환하는 구현은 11에서 `implemented`를 집고 멈춰 12를 못 본다. 1~10번만으로는 3·4번이 둘 다 자격을 갖춘 모양 B라서 그 뮤턴트가 10/10을 통과한다.

술어는 후보이지 확정이 아니다. AC-DCB-005의 전수 대조에서 이 12줄 밖의 반례가 나오면 술어를 고치거나(범위 안) 그 행을 미해제로 남긴다(허용). **술어를 넓혀 반례를 삼키는 것은 금지한다** — 그것이 A의 위험을 뒷문으로 들이는 경로다.

### 5.4 후보 창 게이트 — `closeInfixMatch`로 좁히면 대상이 빠진다

초판은 후보 커밋을 subject의 `closeInfixMatch`로 걸렀다. **측정해 보니 그 게이트가 이 카드의 확정 대상을 떨어뜨린다.**

`closeInfixMatch`가 인정하는 문자열은 셋뿐이다(`transitions.go:80-82`): `3-phase close` · `4-phase close` · `mx-phase audit-ready`. 실측 close 커밋 6개 중 **4개만** 통과한다.

```
7beda68a5|close out 2 specs with 3-phase lifecycle completion (doc-only)   ← 탈락
e979a4d13|chore(spec group c): mx-phase close (...)                        ← 탈락 ★ 이 카드의 확정 대상
```

`e979a4d13`은 `mx-phase close`이지 `mx-phase audit-ready`가 아니다. 즉 subject `closeInfixMatch`를 후보 게이트로 쓰면 **AC-DCB-004가 구조적으로 실패한다.**

따라서 후보 창 게이트는 run-phase가 정하되, 다음을 지킨다.

1. **판별의 무게는 본문 줄 술어(§5.3)가 진다.** subject 게이트는 후보를 줄이는 값싼 필터일 뿐 판별자가 아니다.
2. **subject가 specID를 담으면 후보에서 제외한다** — 그 경우는 1차 워크 소관이다(REQ-DCB-006).
3. subject 쪽 close 신호를 쓴다면 위 6개 커밋 전부를 통과시켜야 하며, 그 술어와 통과 여부를 원장에 실측으로 남긴다. `closeInfixMatch` 재사용은 **측정으로 기각됐다** — 재사용하려면 먼저 그 상수 집합을 넓혀야 하고, 그것은 close 규약 매처를 건드리는 별개 변경이라 이 SPEC 범위 밖이다(§4).

## 6. 측정 환경 주의 — 203을 오독하지 말 것

판정기는 `main`을 본다. 이 저장소의 작업은 `develop`에 착지하고 release PR로만 `main`에 도달하므로 두 브랜치가 크게 벌어져 있다(조사 시점 `main..develop` 1,362 커밋 — 이 수는 **그 시점의 참조값**이며, 판정에 쓰려면 `git rev-list --count main..develop`으로 다시 잰다).

따라서:

- drift 표 203행 대부분은 "close가 아직 main에 없음"이고 **이 카드의 결함이 아니다**. 203을 이 SPEC의 모집단으로 삼으면 안 된다.
- 이 SPEC의 대상은 "close가 main에 있는데도 못 보는" 부분집합뿐이다.
- 그 결과, **이 수리의 효과는 표의 총계에서 거의 보이지 않는다.** 총계 변화로 성패를 판정하지 않는다 — 판정은 AC-DCB-004(확정된 행)와 AC-DCB-005(행별 대조)로만 한다.
- 코퍼스 측정을 인용할 때는 **잰 브랜치와 잰 트리 SHA를 함께 적고**, 앞선 측정의 수를 그대로 옮기지 않는다.

### 6.1 TIGHT 15는 조치 대상 목록이 아니다 — 오탐 4건이 측정됐다

조사의 프로브(`blast-radius.sh`)는 "본문에서 ID를 담은 줄이 `close|completed|Mx verdict`도 담는가"로 TIGHT를 갈랐다. 그 키워드는 **두 부류가 함께 갖는다**(§5.3). 그래서 TIGHT 15는 상한이고, 그중 넷을 본문 대조로 오탐 확인했다.

| 행 | 커밋 | 본문 줄 | 실제 |
|---|---|---|---|
| `SPEC-AUTONOMY-TIERS-001` | a83934d55 | `depends_on: ... (completed)` | 의존성 언급 |
| `SPEC-WORKTREE-BRANCH-GUARD-001` | fb8aff006 | `Depends on ... (both completed)` | 의존성 언급 |
| `SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001` | fb8aff006 | 같은 줄 | 의존성 언급 |
| `SPEC-INTERNAL-TEST-002` | 80dea9684 | `residual_debt items owned by follow-up SPEC-INTERNAL-TEST-002` | **후속** SPEC 지목 — close의 정반대 |

남는 11행은 6개 커밋 중 셋(`e979a4d13`·`2f449e189`·`7beda68a5`)에서 나오며 **미판정**이다. 마지막 행은 조사 §1.4가 "단일 범위인데 형제를 함께 닫음" 형태의 근거로 든 바로 그 커밋이다 — 본문은 형제를 닫는 것이 아니라 **부채를 넘긴다**고 적고 있다. 그 형태가 존재하지 않는다는 뜻은 아니고, **인용된 사례로는 입증되지 않았다**는 뜻이다.

하한도 여전히 미확정이다. LOOSE-only 18건 중 표본 4건을 본문 대조했고(83610e03e·d21b76ebb·51d18d3fe·3169f06d0) 넷 다 언급이었다 — 표본이지 전수가 아니다.

## 7. 참조

- 조사 원장: `.moai/reports/t410/discovery.md`, `r1-walker-trace.log`, `r2-blast-radius.log`, `r3-drift-row.log`, `blast-radius.sh`
- 구현 대상: `internal/spec/drift.go`(`getGitImpliedStatus`, `commitMatchesSPECID`, `shouldSkipCommitTitle`, `combinedScopeCloseMatches`), `internal/spec/drift_index.go`(`inMemImpliedStatus`, `inMemCombinedScopeClose`), `internal/spec/transitions.go`(`closeInfixMatch`)
- 규약: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Close-subject full-ID mandate, § Non-transition frontmatter corrections
- 판정 규율: `.claude/rules/moai/core/verification-claim-integrity.md` §1(공허한 주장), §2(baseline 귀속), §2.2(도구 출처 귀속)
- 선행 SPEC(재판정 대상 아님): `SPEC-ERA-H3-NARROWING-001`, `SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001`, `SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001`, `SPEC-SESSIONSTART-PERF-001`
