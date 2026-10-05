---
id: SPEC-WORKTREE-EXIT-RETURN-001
title: "ExitWorktree 복귀 지점 계약 명시 — 연쇄 진입·워크트리 출발 세션에서의 복귀 지점을 독트린에 못 박는다"
version: "0.3.0"
status: draft
created: 2026-09-19
updated: 2026-09-20
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tier: S
tags: "worktree, exitworktree, enterworktree, doctrine, contract, return-point, documentation-gap"
---

# SPEC-WORKTREE-EXIT-RETURN-001 — ExitWorktree 복귀 지점 계약 명시

## HISTORY

| Version | Date       | Author       | Change |
|---------|------------|--------------|--------|
| 0.1.0   | 2026-09-19 | manager-spec | 최초 초안. 카드 t965. 근거는 `.moai/reports/t965/observations.md` 의 관측 1~4 이며, 이 SPEC 은 새 측정을 만들지 않는다. |
| 0.3.0   | 2026-09-20 | manager-spec | plan-audit iter2(FAIL 0.80) 반영. D1′ — 0.2.0 의 비대칭 결정이 `rule_template_mirror_test.go` 의 바이트 동일성 허용목록과 충돌해 「SPEC 문언을 만족시키는 트리 상태가 존재하지 않는」 상태를 만들었다. **Path A 채택**: 두 사본을 바이트 동일하게 두고 양쪽 모두 런타임 버전 인라인만으로 귀속한다(§1.5 전면 재작성, REQ-WXR-005·008 개정). D9 분할은 초과분으로 선언. D2′ 는 경로 선택으로 닫힘. D11·D12 판별식을 AC-WXR-001·005 에 추가(AC 수 8 유지). 코드 변경 0. |
| 0.2.0   | 2026-09-20 | manager-spec | plan-audit iter1(FAIL 0.69) 반영. D1 귀속 표면 비대칭 결정(§1.5) — 미러에 날짜·내부 경로·SPEC ID 를 싣지 않고 런타임 버전으로 귀속한다. D3 §1.4 버전 Gap 을 근거 파일 현행 문언으로 교체. D4 REQ-001 전칭 술어 제거 + 표본 명시 + 미재현 제보와의 공존 명시. D8 §1.2 분류 단정 철회. AC 7→8(순서 검증 AC 추가), 각 AC 에 커버 REQ 명시. |

## 1. Overview

### 1.1 문제 — 무엇이 비어 있나

카드 t965 의 물음은 「복귀 지점이 틀렸는가」가 아니라 **「`A → EnterWorktree B → EnterWorktree C
→ ExitWorktree` 에서 복귀가 B 인가 A 인가」**였다. 조사(`.moai/reports/t965/observations.md`)의
답은 **A 쪽이되, 카드가 상정한 A 가 아니다**:

- 관측 1 (단일 진입, 대조군): A → Enter(C) → Exit ⇒ **A 로 복귀**. 문서 서술과 일치.
- 관측 2 (연쇄): A → Enter(B) → Enter(C) → Exit ⇒ **A 로 복귀. B 가 아니다.** A 와 B 는 서로
  다른 경로 문자열이므로 이 픽스처는 공허하지 않다.
- 관측 3: **한 응답 안에서** `Your work is preserved at …` 는 C 를, `Session is now back in …`
  은 A 를 지목한다. 읽는 사람이 복귀 지점을 오독하는 모양이 여기서 나온다.
- 관측 4 (세션 이력 1회): 워크트리 **안에서 출발한** 세션이 첫 `EnterWorktree` 뒤 Exit 했을 때
  복귀 지점은 출발 디렉터리가 아니라 **primary checkout** 이었다.

즉 **관측된 네 경우 전부에서** 복귀 지점은 primary checkout 이었다 — 연쇄 깊이 1·2 (관측 1·2)
와, 비-primary 출발 디렉터리 **단 한 사례**(관측 4)에서. 조사 파일 자신이 여기서 "…고정되는
것으로 **보인다**"로 유보했고(`observations.md` § 관측 4), 이 SPEC 은 그 유보를 유지한다 —
표본 2점·1점은 두 축에 대한 **전칭 주장의 근거가 되지 못한다**(REQ-WXR-001). 카드가 상정한
이분법(B 인가 A 인가)에는 없던 **세 번째 답**이다.

### 1.2 이 카드가 고칠 수 있는 것은 계약(문서) 층이다

**분류를 단정하지 않는다.** 관측이 지지하는 읽기는 둘이며, 이 SPEC 은 둘 중 하나를 고르지
않는다:

- **읽기 (가) — 계약(문서) 층의 공백.** 복귀 동작 자체는 관측 1~4 에서 일관되고 예측 가능하다.
  결함은 그 동작을 서술하는 문장이 어느 트리를 가리키는지 말하지 않는다는 데 있다.
- **읽기 (나) — 상위 런타임이 자기 설명을 위반한다.** 아래에 인용한 상위 설명 문장 (2) 는 관측 4
  에서 **거짓으로 읽힌다**. 일관성은 상위 계약과의 합치를 입증하지 않는다 — 문서가 말하는 바를
  일관되게 어기는 런타임도 일관적이다. 이 읽기에서 고칠 표면은 런타임이다.

[HARD] **읽기 (나) 를 기각하지 않는다.** 이 SPEC 이 (나) 를 다루지 않는 이유는 그것이 틀려서가
아니라 **그 표면이 우리 산출물이 아니어서**다(§1.3, §4). (나) 는 §4 에 이름을 달아 범위 밖으로
치워 두며, 제보 후보로 남는다.

아래 공백 서술은 읽기 (가) 의 근거이며, (나) 를 배제하지 않는다:

- 상위 도구 설명 문장 (1): "return the session to the **original working directory**" — A 쪽을
  시사한다.
- 상위 도구 설명 문장 (2): "Restores the session's working directory to **where it was before
  EnterWorktree**" — **어느 `EnterWorktree`** 인지 말하지 않는다. 연쇄에서는 B 로도 읽히고,
  관측 4 의 경우(워크트리 출발 세션)에는 **거짓으로 읽힌다**.
- 같은 도구군의 `EnterWorktree` 설명은 **정리 대상**에 대해서만 명시적이다("only the new one is
  tracked for exit-time cleanup"). 정리 대상은 한 문장으로 못 박아 두고 **복귀 지점은 같은
  수준으로 못 박지 않았다** — 이것이 공백의 정확한 위치다.
- 이 저장소 자신의 독트린(`.claude/rules/moai/workflow/worktree-integration.md`
  § `EnterWorktree` / `ExitWorktree` Tools)도 같은 공백을 물려받았다: "`ExitWorktree` returns to
  the **originating checkout**" 한 문장뿐이고, 「originating」이 세션 출발 디렉터리를 뜻하는지
  primary checkout 을 뜻하는지 구분하지 않는다. 관측 4 는 이 두 해석을 **갈라 놓는다**.

### 1.3 우리가 고칠 수 있는 표면은 하나뿐이다

[HARD] **상위 도구 설명은 이 저장소의 산출물이 아니다.** `EnterWorktree` / `ExitWorktree` 는
Claude Code 런타임 도구이고 그 설명 문자열은 우리가 편집할 수 없다. 따라서 이 SPEC 은 **우리가
바꿀 수 없는 표면에 대고 요구사항을 쓰지 않는다.** 실제 인도물은:

1. `.claude/rules/moai/workflow/worktree-integration.md` — 이 저장소의 워크트리 독트린 (+ §2.0
   Template-First 규율에 따른 템플릿 미러).
2. 그 서술을 관측에 귀속시키는 인수조건, 그리고 증거의 가장 약한 고리를 닫는 재측정 인수조건.

상위 도구 설명 문장 (2) 가 **관측 4 에서 거짓으로 읽힌다**는 사실 자체는 **관측으로 기록하고
제보 후보로 남길 수는 있으나**, 이 SPEC 의 완료 조건이 될 수 없다(우리 손 밖이다). 이것을
「모호성」으로 낮춰 부르지 않는다 — 모호성은 여러 읽기를 허용한다는 뜻이고, 관측 4 에서 문장 (2)
는 읽기 하나가 **거짓**이다.

### 1.4 관측의 귀속과 한계 (그대로 운반한다 — 격상하지 않는다)

아래는 조사 파일의 Gaps 를 **그대로** 옮긴 것이다. 이 SPEC 은 이것들을 findings 로 격상하지
않으며, 격상하는 요구사항을 쓰지도 않는다.

- **원 사례는 재현되지 않았다.** 다른 레인이 다른 **워크트리**로 복귀했다는 최초 제보와, 이
  조사의 모든 관측(primary checkout 복귀)이 서로 다르다. 그 세션의 출발 디렉터리는 복구
  불가능하므로 불일치는 **열린 채로 남는다**.
- **관측 4 는 세션 이력에서 얻은 1회 관측**이며 의도적으로 재실행하지 않았다.
- `EnterWorktree(name:)` 가 워크트리 안에서 거부되는지는 **시험하지 않았다**.
- **런타임 버전은 관측 *도중*에 재지 않았다.** 리드 지시로 **사후 기록**됐고(조사 파일 § 측정
  환경: Claude Code `2.1.278`, macOS `27.0` / `Darwin 27.0.0`), 따라서 이 항목은 닫힌 것이
  아니라 **사후 귀속으로 약화된 채 남는다**. **관측 4 는 그 귀속조차 공유하지 않는다** — 같은
  세션 이력에서 나왔으나 그 시점에 버전을 재지 않았고, 위 값이 그때도 참이었다는 것은 세션
  연속성에 기댄 추론이지 측정이 아니다. 「미기록」과 「사후 기록되어 약화됨」은 다른 상태이며,
  이 SPEC 은 뒤쪽을 운반한다.
- 관측은 **도구 응답 문자열과 환경 갱신 보고**에 의존하며, Exit 직후의 `pwd` 판독이 아니다.

### 1.5 귀속 표면은 두 사본에서 동일하다 — 바이트 파리티 + 런타임 버전 인라인

이 절은 REQ-WXR-005 · REQ-WXR-008 · AC-WXR-005 · AC-WXR-007 의 형태를 결정한다.

**문제.** 「관측에 귀속시킨다」를 소박하게 이행하면 날짜(`2026-09-19`)와 기록 경로
(`.moai/reports/t965/…`)가 독트린 본문에 들어간다. 대상 파일은 템플릿 미러로 그대로 배포되므로
그 두 요소는 모든 사용자 프로젝트에 실린다. 그런데 이 파일 주위에는 가드가 **셋** 있고, 셋이
함께 덫을 이룬다. 아래 표는 세 상태를 각각 심어 보고 관측한 결과다:

| 심은 상태 | 발화하는 가드 | 티어 |
|---|---|---|
| 두 사본의 **비대칭**(미러에만 날짜·내부 경로·SPEC ID 부재) | `RULE_TEMPLATE_MIRROR_DRIFT` — FAIL | — |
| 파리티 유지 + **양쪽에 SPEC ID** | `C1-spec-id-prefix` — FAIL | **default** |
| 파리티 유지 + **양쪽에 날짜 리터럴** | `S1-internal-date` — FAIL | **strict** |

**첫 행이 이 개정의 핵심이다.** 대상 파일 `.claude/rules/moai/workflow/worktree-integration.md`
는 `internal/template/rule_template_mirror_test.go` 의 `workflowOptMirroredPaths` **바이트 동일성
허용목록에 명시 등재**돼 있다(등재 줄은 그 파일에서 대상 경로 문자열로 찾는다 — 측정 시점에는
57행이었으나 행 번호는 움직이는 좌표이므로 심볼로 인용한다). 단언부는
`bytes.Equal(srcContent, mirrorContent)` 이며, 실패 시 센티널 `RULE_TEMPLATE_MIRROR_DRIFT` 를
낸다. 등재 전제는 지금도 살아 있다 — 두 사본의 sha256 이 현재 동일하다. 즉 **비대칭은 이 파일에
한해 기계적으로 금지돼 있다.**

이 SPEC 의 이전 판(0.2.0)은 비대칭을 결정으로 채택했다. 그 결정은 누출 가드는 피하지만 파리티
가드를 적색으로 만든다 — 완료 시점에 CI 가 적색이 되는 조건을 **제거한 것이 아니라 다른 가드로
옮긴 것**이다.

**결정 — 두 사본을 바이트 동일하게 두고, 양쪽 모두 런타임 버전 인라인만으로 귀속한다.**

| | 로컬 사본 (`.claude/rules/…`) | 템플릿 미러 (`internal/template/templates/…`) |
|---|---|---|
| 규범 문장 | 동일 | **동일**(REQ-WXR-008) |
| 바이트 | — | **동일** — 허용목록 등재가 요구한다 |
| 측정 날짜 | **없음** | **없음** — `S1-internal-date`(strict) |
| 저장소 내부 증거 경로 | **없음** | **없음** — 배포 독자에게 해소되지 않는다 |
| SPEC ID | **없음** | **없음** — `C1-spec-id-prefix`(default) |
| 런타임 버전 | **있음 — 인라인** | **있음 — 인라인. 이것이 두 사본 공통의 유일한 귀속 표면이다** |

**왜 런타임 버전이 충분한 귀속 표면인가.** 이 주장이 낡는 축은 날짜가 아니라 **런타임 버전**이다
— 「어느 빌드에서 이 동작이 관측됐는가」가 다음 사람이 「구현이 바뀌었다」와 「측정이 틀렸다」를
가를 수 있는 좌표이고, 그것은 사용자 프로젝트의 독자도 자기 `claude --version` 과 대조해 쓸 수
있다. 날짜는 같은 일을 하지 못한다(같은 날 서로 다른 빌드가 돌 수 있다). 0.2.0 에서 이 논거는
미러 한쪽에만 적용됐으나, Path A 에서는 **두 사본 모두**에 적용되므로 오히려 강해진다 — 저장소
기여자에게도 날짜보다 버전이 나은 좌표라는 사실은 사본이 어느 쪽이든 변하지 않는다.

**전례(측정).** 배포되는 미러가 런타임 버전만으로 관측을 귀속시키면서 바이트 파리티를 유지하는
형태는 이미 이 트리에 있다: `.claude/rules/moai/workflow/kanban-dispatch.md` 와 그 미러다.
두 사본의 sha256 이 동일하고(`63ff0953…`), 미러는 "Measured on Claude Code 2.1.275" /
"Measured on Claude Code 2.1.276, inside a worktree session" 을 싣고 있으며, 그 파일의 날짜
리터럴은 **0건**이고 default·strict 두 티어 아래에서 모두 초록이다. 주목할 점: **그 파일은
허용목록에 등재돼 있지 않은데도 바이트 동일하다** — 파리티가 등재의 부산물이 아니라 이 형태가
자연히 도달하는 상태라는 뜻이다. 이 SPEC 은 새 형식을 발명하지 않고 그 형태를 따른다.

**귀속 기록의 인용은 SPEC·progress 층이 맡는다(0.2.0 의 D9 분할은 이 결정으로 초과분이 된다).**
0.2.0 은 「로컬 사본은 추적되는 증거 경로를 인용하고 미러는 인용하지 않는다」는 분할을 채택했다.
그 분할의 *논리*는 지금도 옳다 — 두 사본은 서로 다른 독자를 가지며, 해소되지 않는 경로를
인용하는 주장은 `verification-claim-integrity.md` §2 의 미귀속 주장이다. **초과분이 된 이유는
논리가 틀려서가 아니라 전제가 사라져서다**: 바이트 파리티 아래에서 「로컬 사본에만 있는 문장」은
존재할 수 없고, 로컬 사본에 실은 저장소 내부 경로는 그대로 미러로 배포된다. 따라서 **독트린
본문(두 사본 모두)은 어떤 증거 경로도 인용하지 않는다.**

그렇다고 반출 의무가 사라지지는 않는다. 조사 기록과 M1 재측정 기록은 여전히 추적되는 경로
`.moai/specs/SPEC-WORKTREE-EXIT-RETURN-001/evidence/` 아래로 **반출되고**, 그 경로를 인용하는
것은 이 SPEC 자신과 `progress.md` 다 — 배포되지 않는 층이므로 경로도 SPEC ID 도 날짜도 실을 수
있다. 「인용 전 반출」 의무(`agent-common-protocol.md` § Parallel Execution)는 인용이 일어나는
층에서 그대로 이행된다. 이동한 것은 **귀속 앵커의 위치**이지 귀속 자체가 아니다.

**D2′ — 상시 가드 소실 위험은 경로 선택으로 닫혔다(편집으로 닫힌 것이 아니다).** 비대칭을
유지하는 길(Path B)은 대상 파일을 `workflowOptMirroredPaths` 에서 **빼고**
`sanitized_pair_parity_test.go` 의 `sanitizedPairPaths` 에 넣는 §25 sanitized-pair 이전을
요구한다. 그 이전을 절반만 이행하면 — 허용목록에서 빼고 등재를 빠뜨리면 — 이 파일을 묶는
**상시 가드가 하나도 남지 않는다**. 이 SPEC 의 AC 는 이 카드의 run/sync 회차에 한 번 도는 문서
검사이지 이후 편집을 잡는 CI 가드가 아니므로, 그 상태는 조용하다. Path A 는 허용목록을 건드리지
않으므로 **바이트 동일성 가드가 그대로 살아 있고**, 이 위험은 발생하지 않는다. 기록해 두는
이유는 두 가지다: (가) 이것은 편집으로 닫힌 결함이 아니라 **결정으로 닫힌 결함**이므로 문언
어디에도 대응 수정이 없고, (나) 나중에 누군가 비대칭을 다시 채택하려 하면 **등재를 같은 커밋에서
함께 해야 한다**는 조건이 여기 남아 있어야 한다.

**Path B 를 기각한 이유(두 가지, 각각 독립적으로 충분하다).** 첫째, 두 편집 대상이 모두
`internal/template/*_test.go` 이므로 이 SPEC 자신의 [HARD] 「코드 변경 없음」(§4, `plan.md` §D,
DoD)과 정면으로 충돌한다. 둘째, 위 D2′ 가 보인 대로 최소 이행이 상시 가드를 0 으로 만든다 —
결함을 제거하는 대신 옮기는 수정이 **연속 두 번** 되는 모양이다.

## 2. Requirements (GEARS)

`<subject>` 는 일반화된 주어를 쓴다. 대상 문서는 `.claude/rules/moai/workflow/worktree-integration.md`
(및 그 템플릿 미러)이며, 아래에서 **the doctrine** 으로 부른다.

- **REQ-WXR-001** (Ubiquitous) — The doctrine shall state the observed `ExitWorktree` return point
  as the **primary checkout**, and shall carry the sample the statement rests on **in the sentence
  itself**: chain depths 1 and 2 (observations 1 and 2), and a **single** observed non-primary
  launch-directory case (observation 4, one session-history observation, not re-runnable). The
  doctrine shall **not** use a universal predicate — `independent of`, `always`, `regardless of` —
  for either axis. The doctrine shall further state, in the same place, that this sample does **not**
  close the unreproduced cross-worktree return report that REQ-WXR-006 keeps open: the observations
  differ from that report, and differing from a report is not the same as explaining it.

- **REQ-WXR-002** (Event-driven) — **When** a reader consults the doctrine's `EnterWorktree` /
  `ExitWorktree` section for the chained case (`A → Enter B → Enter C → Exit`), the doctrine shall
  name the return point unambiguously as A (the primary checkout) and shall state explicitly that
  it is **not** B.

- **REQ-WXR-003** (Event-driven) — **When** the doctrine describes the `ExitWorktree` response, it
  shall state that the response names **two different trees in one message** — the tree just left
  (`Your work is preserved at …`) and the return point (`Session is now back in …`) — and that the
  first is not the return point.

- **REQ-WXR-004** (Ubiquitous) — The doctrine shall reconcile its own sentence "`ExitWorktree`
  returns to the originating checkout" with the observation that a session **launched inside a
  worktree** returned to the primary checkout rather than to its launch directory, so that
  "originating" carries exactly one reading.

- **REQ-WXR-005** (Where — capability gate) — **Where** a return-point statement rests on runtime
  behaviour this repository does not produce — whether the ground is an **observation** of that
  behaviour or the **upstream tool description** this repository cannot edit — the doctrine shall
  present the statement as a **measured observation carrying its attribution**, never as a quotation
  of a tool contract and never with guarantee wording (`always` / `guaranteed` / 「항상」). The
  attribution surface is **identical in both copies**, per REQ-WXR-008: each copy carries the
  **runtime version inline** and nothing else — no measurement-date literal, no repository-internal
  evidence path, no SPEC ID. The tracked evidence-record path is cited by this SPEC and its
  `progress.md`, never by the doctrine (§1.5).

- **REQ-WXR-006** (Unwanted) — The doctrine shall not assert that the originating cross-worktree
  return report was reproduced, and shall not assert a cause for the discrepancy; it shall record
  the discrepancy as open and unreproduced.

- **REQ-WXR-007** (While — state-driven) — **While** any observation the return-point statement
  rests on has **not** been covered by a direct post-exit working-directory read, the doctrine shall
  carry a residual-risk note naming those observations and stating that they rest on tool-response
  strings and environment-update notices rather than on a read. The re-measurement of AC-WXR-004
  covers the chained-entry axis **only**; it does not cover observation 4 (the launch-directory
  axis), which is not re-runnable, so the note survives that re-measurement in narrowed form rather
  than being deleted by it.

- **REQ-WXR-008** (Ubiquitous) — The template mirror of the doctrine file shall be **byte-identical**
  to the local copy, in the same commit, per the repository's Template-First rule. The parity is not
  a stylistic preference: this file is enrolled in the `workflowOptMirroredPaths` byte-identity
  allowlist of `internal/template/rule_template_mirror_test.go`, whose assertion is
  `bytes.Equal(srcContent, mirrorContent)` and whose failure sentinel is
  `RULE_TEMPLATE_MIRROR_DRIFT`, so **no divergence between the two copies is permitted** — not in
  the normative layer, and not in the attribution layer. Consequently the doctrine shall carry, in
  **both** copies alike, **no** measurement-date literal, **no** repository-internal evidence path,
  and **no** SPEC ID, and shall attribute every return-point statement by **runtime version inline**
  (§1.5). Parity binds the normative layer in full — the return point, the chained case and its
  explicit `not B`, the two-tree response warning, the corrected `originating` wording, the
  non-guarantee hedge, the residual-risk note, and the open unreproduced report — and, being byte
  parity, it binds every other character of the file as well.

## 3. Acceptance Criteria

Tier S 이지만 카드가 세 파일을 명시했으므로 인수조건 본문은 `acceptance.md` 에 둔다(8개, Tier S
상한 이내). 각 AC 는 자기가 덮는 `REQ-WXR-XXX` 를 머리에 명시 인용한다. 요지:

- AC-WXR-001~003 — 문서 층의 서술을 고정한다.
- **AC-WXR-004 — 증거의 가장 약한 고리**(응답 문자열 의존)를 닫는 재측정. Exit 직후 작업
  디렉터리를 **직접 판독**하고 **런타임 버전을 기록**한다. 이 AC 는 판독값이 A 일 것을 요구하지
  않는다.
- AC-WXR-005 — 귀속. **회차별로 분리**하고(관측 회차 ↔ 재측정 회차), 두 사본 공통의 귀속 표면이
  **런타임 버전 인라인**임을 판정한다. 증거 경로의 추적 여부는 독트린이 아니라 **SPEC·progress
  층의 인용**에 대해 판정한다(§1.5).
- AC-WXR-006 — 미재현 불일치가 열린 채로 남아 있음.
- AC-WXR-007 — 미러 파리티. **두 사본의 바이트 동일성**과 **양쪽 공통 부재 검사**(날짜·내부
  경로·SPEC ID)를 판정하고, 이 파일을 지배하는 가드 **셋 모두**(바이트 파리티 + 중립성 default +
  중립성 strict)를 완료 **이전에** 실행해 초록임을 관측한다(§1.5).
- AC-WXR-008 — M1 기록 커밋이 M2 독트린 커밋의 **조상**임을 커밋 그래프로 판정한다
  (`verification-claim-integrity.md` §2.3).

## 4. Exclusions

이 SPEC 이 만들지 않는 것 — out of scope 목록이다.

### Out of Scope — 상위 런타임 도구 표면

- `EnterWorktree` / `ExitWorktree` 도구 설명 문자열의 수정. 우리 저장소의 산출물이 아니며,
  요구사항을 걸 수 없다.
- **읽기 (나) 의 추궁 — 「상위 런타임이 자기 설명(문장 (2))을 위반한다」.** §1.2 가 이 읽기를
  이름 달아 두었고, 관측 4 가 그것을 지지한다. **기각이 아니라 범위 밖이다**: 고칠 표면이 상위
  런타임이며 우리 손 밖이다. 이 SPEC 은 이 읽기를 부정하는 문장을 쓰지 않으며, 별건 제보
  후보로 남긴다.
- 복귀 지점 자체의 동작 변경(어느 트리로 복귀할지). 관측된 동작은 일관되며, 이 카드는 그것을
  바꾸자고 하지 않는다.

### Out of Scope — 미측정 영역

- 원 제보(다른 워크트리로의 복귀)의 원인 규명. 그 세션의 출발 디렉터리가 복구 불가능하므로 이
  SPEC 에서 닫지 않는다.
- `EnterWorktree(name:)` 가 워크트리 안에서 거부되는지의 판정. 조사에서 시험되지 않았고, 이
  SPEC 은 그에 대한 서술을 만들지 않는다.
- 복귀 지점의 **버전별 귀속**(어느 런타임 버전부터 이 동작인가). AC-WXR-004 는 측정 시점 버전
  **한 점**만 기록하며, 버전 구간을 주장하지 않는다.

### Out of Scope — 코드 층

- Go 코드, 훅, CLI 변경 없음. 이 카드는 계약 명시 카드이며 기능 카드가 아니다.
- **템플릿 중립성 가드의 완화.** 미러에 측정 날짜를 싣기 위해
  `internal/template/internal_content_leak_test.go` 의 `dateAllowlist` 에 항목을 더하는 길은
  **검토 후 채택하지 않았다**(§1.5). 그 길은 코드 층 변경이 되고, 배포되는 문서에 내부 날짜를
  들이기 위해 가드를 여는 거래가 된다. 미러는 런타임 버전으로 귀속하면 충분하므로 가드를 건드릴
  이유가 없다.
- **바이트 동일성 허용목록에서의 이탈(§25 sanitized-pair 이전).** 대상 파일을
  `internal/template/rule_template_mirror_test.go` 의 `workflowOptMirroredPaths` 에서 빼고
  `internal/template/sanitized_pair_parity_test.go` 의 `sanitizedPairPaths` 에 등재해 두 사본의
  비대칭을 허용하는 길(Path B)은 **검토 후 기각했다**(§1.5). 두 편집 모두 `internal/` 아래의
  테스트 파일이고, 최소 이행(허용목록 제거만)은 이 파일을 묶는 상시 가드를 0 으로 만든다.
- 워크트리 생성·폐기·잠금 수명 규율의 변경. 독트린의 다른 절은 건드리지 않는다.
