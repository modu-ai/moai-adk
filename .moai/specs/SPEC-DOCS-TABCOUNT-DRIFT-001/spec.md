---
id: SPEC-DOCS-TABCOUNT-DRIFT-001
title: 설정 탭 수·이름이 문서마다 따로 세어지는 드리프트 차단
version: "0.3.1"
status: draft
created: 2026-09-12
updated: 2026-09-13
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: "README.{md,ko,ja,zh}, docs-site/content/{ko,en,ja,zh}/{cli-reference/web.md,advanced/moai-web-console.md}, internal/web"
lifecycle: spec-anchored
tags: "docs, docs-site, readme, i18n, drift, web-console, guard, t530"
tier: M
related_specs:
  - SPEC-WEB-CODEX-PANEL-001
  - SPEC-PRECOMMIT-GATE-SCOPE-001
---

# SPEC-DOCS-TABCOUNT-DRIFT-001 — 설정 탭 수·이름 드리프트 차단

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.3.1 | 2026-09-13 | manager-spec | plan-audit iter3 (PASS 0.8375) 의 유일한 blocking 소견 **N1** 수리 — 좁은 수리이지 새 라운드가 아니다. **무엇이 비어 있었나**: `REQ-TCD-005` 는 가드가 "숫자와 낱말 모두" 를 매치하라고 명시하는데, 그 조항의 **낱말 클래스 폭**을 구속하는 검증 항목이 없었다 — `AC-TCD-012` 의 When 이 실행하는 것은 셸 파이프라인이라 **셸**을 구속했고, `AC-TCD-006 V4` 는 변이가 `fourteen` 한 건이었다. **무엇이 이제 구속하나**: `AC-TCD-012` 에 [HARD] And 절을 더해 **Go 가드 자신**이 낱말 축 적중 집합을 `word-axis hit <경로>: <구절>` 로 찍게 하고, M1 직후·M2 이전 트리에서 그 출력이 **정확히 6줄**이며 A2·A4·B4·C1·C2·C4 와 같음을 단정한다 — 즉 `nine`(en)·`九`(zh)·`fourteen`(en)·`열네`(ko)·`十四`(zh) 다섯 토큰이 모두 한 줄 이상을 낸다. 보강으로 `AC-TCD-006` 의 변이를 4종 → **6종**으로 넓혔다(V5 ko `열네`, V6 zh `九个`). **죽는 변이체**: 낱말 클래스를 **`{fourteen}` 하나로만 구현한 가드** — V1·V2·V3·V4 와 셸 Then 절을 전부 통과하면서 `fourteen` 자리 2줄만 찍어 6줄 단정에서 FAIL 하고, V5·V6 에서도 FAIL 한다. 그 가드는 열거된 낱말 표기 6자리 중 5자리를 놓치며, 그것이 이 카드가 지난 청소의 실패 원인으로 지목한 부류다(§1.3 첫 함정). 부수: `plan.md §A.5`·M1 을 같은 내용으로 훑었고(가드가 집합을 찍는 것과 **M2 이전에** 재야 한다는 순서 전제), REQ/AC/RG 수는 9/12/3 그대로다 — 새 AC 를 만들지 않고 기존 둘을 조였다. N5(트레일러 귀속)는 리드 판정으로 이력을 고치지 않고 `progress.md §E.1` 에 설명 기록으로 닫았다. |
| 0.3.0 | 2026-09-12 | manager-spec | plan-audit iter2 (FAIL 0.775) 수리 — 네 blocking 중 **둘은 0.2.0 이 만든 것을 되돌린 것**이다. **D1(critical)**: `AC-TCD-010` 이 리터럴 base + `docs-site/content` 통째 pathspec 이라, 이 저장소가 [HARD] 로 요구하는 develop 흡수·병합 트리 재측정 절차가 **그 자체로** 기준을 붉게 만들었다(실측: `1d150a27d..develop` 40커밋·155파일, 위반 상대 경로 4개 전부 남의 것). pathspec 을 대상 12파일로 좁히고 왼쪽 끝을 읽는 시점 `git merge-base develop HEAD` 로 바꿨다; `RG-TCD-001` 도 같은 끝점으로(D11). **D2(major)**: 낱말 표기 6자리에 **두 번째 층이 없었다** — `unfolds fourteen tabs` → `fourteen settings tabs` 한 낱말이면 리터럴 0 / 숫자 스윕 0 으로 인수 기준이 초록이 된다(한국어 `열네 개 탭` → `탭 열네 개` 도 동일). 가드 N1 의 수사 축을 **낱말까지** 열고, `AC-TCD-012`(낱말 축 적중 집합) + `AC-TCD-006 V4`(변이) 를 신설. **D3(major)**: 0.2.0 이 "낱말 축은 열 수 없다 — 실측 확인" 이라 적었으나 근거 3자리가 전부 `:149` codex 문단이었고 이미 채택한 허용 규칙이 덮는 줄이었다 — 전제가 측정에 의해 뒷받침되지 않았다(`verification-claim-integrity.md` §1.1 surface 4). 허용 규칙 + 서수 접두(`第`) 제외를 적용해 재측정: 원시 8행 → 7 → **6행, 전부 열거 자리, 허용되지 않은 오탐 0** ⇒ 축을 열고 §7 잔여 위험을 "오탐" 이 아니라 **수사 클래스 유지 비용**으로 고쳐 적었다. **D4(major)**: 0.2.0 의 정직성 정정이 `spec.md`·`progress.md` 에만 닿고 열거 산출물의 머리 문장(`:14`)을 놓쳐, 같은 파일 안에서 `:14` 와 `:167` 이 서로 모순이었다 — 전 아티팩트를 훑어 고치고 0.1.0 HISTORY 행에도 정정 주석을 달았다. 부수: D5 열거 완전성 추출을 A/B/C/D 표 행으로 한정(B군 표 삭제 변이가 통과하던 구멍 — B군 경로 축약을 전체 경로로 되돌림), D6 `swept` 카운터를 **읽기 성공 후 증가**로 규정, D7 `allowed` 단위를 `rules 1` + `lines 16` 으로 분리하고 허용 규칙의 **파일 범위 필수**를 단정(범위 없는 `codex` 규칙은 README 4자리를 조용히 면제 — 실측 6행→3행), D8 존재하지 않는 AC id 인용 2건 정정, D9 `§1.1` 자기 모순 해소, D10 `REQ-TCD-005` 의 행 번호를 구절 식별로 교체. |
| 0.2.0 | 2026-09-12 | manager-spec | plan-audit iter1 (FAIL 0.66) 수리. D1: 인수 판정을 창 기반 정규식에서 **정확 리터럴 집합**으로 옮겨 영어 `The nine settings tabs`(수사-명사 간격 10글자)를 구조적으로 포함 — base 실측 A군 4/4, 전체 20/20 적중. D2: 화이트리스트 **안**의 정당한 계수(`ja/advanced/moai-web-console.md:149`, codex 근거 문장)를 발견해 열거 §7-B 허용 목록으로 승격하고, AC 기대값을 "스윕 0건" 에서 "열거 집합 0매치" 로 바꿔 도달 불가능(red-forever)을 제거. D3: 변이 지점을 파일 부류 × 로케일 3종으로 넓히고 서브테스트 PASS 줄을 리터럴로 단정 + 읽은 파일 수 == 12 AC 신설. D4: 패리티 AC 에 README 4본 포함·고정 base `1d150a27d`·대조군(변경 0 ⇒ 측정 불가 = FAIL). D5: 스크린샷·hugo 를 **회귀 가드(RG)** 로 재분류. D6: D군 산문 4자리를 D9~D12 행으로 승격(열거 28 → **32**행). D7: `plan.md`·`acceptance.md` 의 `status:` 제거. D8: "서로 독립인 세 측정" → 원천 1 + 미러 1 + 일치 테스트 1 로 하향. D9/D10/D11: 낱말 축을 정규식이 아니라 리터럴로 닫는 근거 기록, 오탐 자리를 구절까지 고정, 열거 최신성에 검증 명령 부여. MP-7: 해소 마커 토큰 제거. |
| 0.1.0 | 2026-09-12 | manager-spec | 초기 draft (카드 t530). **[0.2.0 정정: "서로 독립인 세 측정" 은 과장이었다 — 원천 1(구현) + 미러 1(테스트 리터럴) + 둘의 일치 테스트 1 이다. 아래 문구는 당시 기록으로 보존한다.]** base `1d150a27d` 에서 탭 수 14를 서로 독립인 세 측정으로 재도출하고, 손으로 적힌 자리 20곳(A군 4 오류 + B군 8 + C군 8)과 이름 드리프트 8곳(D군)을 전수 열거했다. 전수 목록은 `.moai/reports/t530/tab-count-sites.md`. 해법의 축은 "14로 고치기"가 아니라 "수를 없애고, 남는 자리는 기계가 지키게 하기". 스크린샷 재생성은 근거를 적어 별건으로 분리(§5). |

---

## 1. Goal

`moai web` 설정 화면의 탭 수와 탭 이름은 코드(`internal/web/schemaform.go` 의 `consoleTabs()`)가 정하는 **하나의 사실**인데,
지금은 12개 문서에서 20자리에 걸쳐 **각각 따로 세어져** 손으로 적혀 있다. 그 결과 4자리는 이미 틀렸고(9라고 적혀 있다),
탭 **이름**도 한 칸 어긋나 있다. 본 SPEC 은 개별 숫자를 고치는 것이 목표가 아니다 — **같은 사실이 여러 문서에서
독립적으로 세어지는 구조 자체를 없애는 것**이 목표다.

세 갈래로 처리한다.

1. **수를 지운다.** 탭을 세는 대신 이름을 부르거나 "아래 목록" 으로 가리킬 수 있는 자리는 수를 제거한다. 지워진 수는 다시 어긋날 수 없다.
2. **남는 자리는 기계가 지킨다.** 수 또는 이름을 남겨야 하는 자리에는 `consoleTabs()` 를 읽어 문서와 대조하는 가드 테스트를 둔다.
3. **4로케일을 한 변경으로 묶는다.** 한 로케일만 고치는 것은 새 드리프트를 만드는 일이다.

### 1.1 배경 — 실측

base `1d150a27d`, 워크트리 `.claude/worktrees/t530` 에서 다시 쟀다. 카드가 들고 온 숫자를 그대로 옮기지 않았다.

**정확히 말하면 원천 1(구현) + 미러 1(테스트 리터럴) + 둘의 일치를 확인하는 테스트 1 이다 — "서로 독립인 세 측정" 이 아니다.**
측정 3은 측정 1과 2를 비교하는 테스트 그 자체이고(`tab_layout_test.go` 가 `len(tabs) != len(wantTabOrder)` 를 단정한다),
측정 2의 `wantTabOrder` 는 `consoleTabs()` 를 미러하도록 손으로 적은 리터럴이라 독립성이 약하다.
실질은 **원천 1 + 미러 1 + 둘의 일치 테스트 1** 이다. 근거로 충분하되, 독립성을 부풀려 적지 않는다 —
같은 사실을 여러 곳에서 따로 센 것처럼 보이게 하는 과장이야말로 이 카드가 다루는 결함이다.

| 측정 | 성격 | 명령 | 관측 |
|---|---|---|---|
| 1 | 원천(구현) | `sed -n '30,90p' internal/web/schemaform.go \| grep -c 'LabelKey:'` | `14` |
| 2 | 미러(테스트 리터럴) | `grep -n -A 6 'var wantTabOrder' internal/web/tab_layout_test.go` | id 14개 |
| 3 | 1과 2의 일치 확인 | `go test ./internal/web/ -run 'TestConsoleTabsOrder'` | `ok … 0.794s` |

**14 는 이 base 에서의 실측치이지 상수가 아니다.** 탭을 하나 추가하는 다른 카드가 값을 바꾸며,
지난 드리프트(11 → 14)가 바로 그렇게 생겼다. 그래서 본 문서도 14를 정답으로 박지 않고,
독자가 **다시 재는 방법**을 §1.1 의 표로 남긴다.

손으로 적힌 자리의 전수 목록 — 파일·행·로케일·숫자/낱말 표기 — 은 `.moai/reports/t530/tab-count-sites.md` 에 있다.
요약: A군 4자리(9라고 적힘, 전부 `cli-reference/web.md:53`), B군 8자리(14, `advanced/moai-web-console.md`),
C군 8자리(14, README 4본 × 2) — 수 자리 합계 **20**; D군 **12자리**(탭 **이름** 드리프트: 번호 목록 8 + 산문 4).
수 20자리 중 **6자리는 낱말 표기**(A2 `The nine settings tabs`, A4 `设置九个标签页`, B4 `unfolds fourteen tabs`,
C1 `fourteen tabs`, C2 `열네 개 탭`, C4 `十四个标签页`)이며, 이 부류가 §1.3 의 함정이 집중되는 자리다.

### 1.2 핵심 발견 — 이름 드리프트는 수 드리프트에 관한 증거다

D군(탭 **이름**이 한 칸 어긋나 있다)은 수 드리프트와 나란히 놓인 두 번째 결함이 **아니다.**
수 드리프트가 어떻게 닫혀야 하는지를 말해 주는 **증거**다.

논증은 세 걸음이다.

1. 자리별 처분을 끝까지 해 보니 **20자리가 전부 removable 이었다**(`plan.md §B`). 수를 지울 수 없는 자리는 없다.
2. 수를 지우면 그 자리의 정보는 사라지지 않고 **이름이 대신 진다** — README:414 는 이미 이름 14개를 적고 있고,
   `advanced/moai-web-console.md:128` 바로 아래에도 이름 목록이 있다. 이름은 지울 수 있는 정보가 아니다.
3. 그러므로 수만 고치는 수정은 결함을 **닫지 않고 이름 쪽으로 옮긴다.** 그리고 D군은 그 이동이
   **이미 한 번 일어났음을** 보여 준다 — 수(14)는 맞는데 이름은 어긋나 있는 상태가 지금 트리의 실물이다.

이것이 이 카드가 "숫자 하나 고치기" 가 아닌 이유이고, 판정이 N0(리터럴 집합) · N1(수사 스윕) · N2(이름 == 렌더 라벨)
**세 층**이어야 하는 이유다(`plan.md §A.3`). 수 축만 두면 오늘의 결함은 지우면서 내일의 자리를 열어 둔 채 닫는 셈이 된다.

### 1.3 세 가지 함정 — 정규식 하나로는 판별되지 않는다

- **숫자와 낱말이 섞여 있다.** `nine`, `fourteen`, `十四`, `九个`, `열네 개` 처럼 로케일마다 수사가 낱말로도 적힌다.
  숫자만 긁는 청소나 가드는 절반을 조용히 놓친다 — 지난 청소가 살아남은 경로가 이것으로 보인다.
- **낱말 사이에 수식어가 낀다.** 영어 A군 자리는 `The nine settings tabs` 로, 수사와 탭 명사 사이가 10글자다.
  인접 창을 좁게 잡으면 이 자리를 **구조적으로 못 본다** — 좁히는 방향은 오탐을 줄이는 대신 어순 수식어가 낀
  영어를 통째로 잃는다.
- **넓히면 화이트리스트 안에서 오탐이 난다.** 정당한 비-설정탭 계수가 대상 파일 안에 실재한다:
  `docs-site/content/ja/advanced/moai-web-console.md:149` 의 `2 つのタブを行き来し` 는 codex 패널의 존재 이유를
  설명하며 **감사 탭과 MCP 탭 둘**을 센다. 파일 화이트리스트도 수-명사 인접도 이 줄을 걸러 내지 못한다 —
  **명시 허용 목록**이 필요하다. 그 허용 목록이 있고 나면 낱말 축을 여는 비용은 사라진다(아래 표 각주).

**결론 — 정규식 하나에 "무엇이 설정 탭 계수인가" 를 단독으로 맡기지 않는다.** 본 SPEC 은 판정을 두 층으로 나눈다.

| 층 | 무엇을 판정하나 | 수단 | 오탐 가능성 |
|---|---|---|---|
| 인수(이 카드가 일을 했는가) | 열거된 20자리가 사라졌는가 | 열거에서 뽑은 **정확 리터럴 집합** (`.moai/reports/t530/count-literals.txt`) | 없음 — 실측한 문자열만 본다 |
| 가드(앞으로 어긋나지 않는가) | 새 수·이름이 들어왔는가 | **숫자 + 낱말** 수사 인접 스윕 + 서수 접두 제외 + **명시 허용 목록**(파일 범위 포함) + 이름 대조(N2) | 실측 0 — 허용 목록 적용 후 남은 적중 6행이 전부 열거된 자리다 |

인수 판정이 정규식에서 분리되므로, 가드의 창을 넓히거나 좁히는 조정이 이 카드의 합격 여부를 흔들지 않는다.

**두 층은 서로를 덮어야 하고, 낱말 표기 자리에서도 그렇다.** 리터럴 층만으로는 재작성 한 번에 빗나간다 —
`unfolds fourteen tabs` → `unfolds fourteen settings tabs`, 또는 한국어의 `열네 개 탭` → `탭 열네 개`. 둘 다 악의가 아니라
평범한 문장 손질이며, 그 순간 손으로 적힌 수가 남은 채 인수 기준이 초록이 된다. 그래서 가드의 수사 축은
숫자와 낱말을 **모두** 본다. 낱말 축을 여는 대가로 늘어나는 오탐은 실측 결과 **0** 이다(`plan.md §A.4` 의 단계별 표).

---

## 2. Requirements (GEARS)

### REQ-TCD-001 — 열거 산출물 (Ubiquitous)

The card artifact set shall contain an exhaustive enumeration of every location where the settings tab
count or the settings tab name list is written by hand, naming for each: file path, line number, locale,
and whether the value appears as a digit or as a spelled-out word. The enumeration shall live at
`.moai/reports/t530/tab-count-sites.md` as a durable artifact, not merely as a step someone performed.

### REQ-TCD-002 — 틀린 수의 제거 (Ubiquitous)

All four `docs-site/content/<locale>/cli-reference/web.md` line-53 rows — ko, **en**, ja, zh — shall no
longer state a settings tab count. The `/settings` row shall describe the endpoint without counting the
tabs. Verification shall be per-locale against the enumerated literal of each site, never a window-based
regex sweep: the en site reads `The nine settings tabs`, whose numeral and tab noun are ten characters
apart, so a window narrow enough to suppress false positives cannot see it.

### REQ-TCD-003 — 수를 없앨 수 있는 자리는 없앤다 (Ubiquitous)

Where a page can name the tabs or point at an adjacent list instead of counting them, the document shall
name or point rather than count. Each of the 20 enumerated sites shall be recorded in `plan.md` as either
`removable` or `must-stay`, with the reason stated per site.

### REQ-TCD-004 — 남는 수와 이름은 기계가 지킨다 (Where 절)

Where a settings tab count or tab name list remains written in a document after REQ-TCD-003, the repository
shall carry a guard test that reads `consoleTabs()` and fails when the documented value diverges from it.

### REQ-TCD-005 — 가드의 판별력 (Event-driven)

When the guard scans a document, it shall restrict its scan to an explicit 12-file list, shall match **both
digit and spelled-out numerals** only where adjacent to a **word-bounded** tab noun
(`\btabs?\b` / `탭` / `タブ` / `标签页`), shall exclude a numeral immediately preceded by the ordinal prefix
`第`, and shall exempt only the entries of an explicit allowlist, each carrying a stated reason and a
**file scope**.

Negative cases are named by phrase, never by line number — this card edits these files, and `develop` has
already moved README line numbers. The guard shall not flag: the SVG-forms sentence containing
`measured nine forms` (README), the skills count `Eleven ref skills` (README), the agent count
`Eleven of the twelve are agents` (README), the substring inside `stays selectable` (README — the reason the
tab noun is word-bounded), the ordinal phrase `第三方 LLM` followed by `标签页` (zh console prose — the reason
`第` is excluded), or the plugin-manager sentence `4 タブを持つプラグインマネージャー` (outside the file list).

The allowlist shall carry exactly one rule — the codex-panel rationale lines, identified as lines containing
the token `codex` **within `advanced/moai-web-console.md`** — which count the Audit and MCP tabs and are out
of scope per §4. The file scope is load-bearing: an unscoped `codex` rule also exempts the four README tab-name
list lines, silently removing four enumerated sites from the guard.

### REQ-TCD-006 — 이름 정본 대조 (Ubiquitous)

The documented tab name list shall equal the labels the console actually renders
(`internal/web/assets/i18n.js`, with `schemaform.go` `Baseline` as fallback), in the order `consoleTabs()`
returns them. The guard shall assert this equality for the enumerated list sites.

### REQ-TCD-007 — 4로케일 동시성 (State-driven)

While any one of the ko / en / ja / zh copies of a page is changed, the other three copies of that page
shall be changed in the same commit. A locale left behind is itself a new drift.

### REQ-TCD-008 — 수를 되살리지 않는다 (Unwanted)

The change shall not reintroduce a hand-written settings tab count anywhere the guard does not cover, and
shall not "fix" a site by writing `14` where the number can be removed instead.

### REQ-TCD-009 — 스크린샷 분리 (Ubiquitous)

The card shall not regenerate `assets/images/moai-web-settings.png`. The decision and its reasons shall be
recorded in this document (§5), and the missing regeneration procedure shall be raised as a separate card
request rather than folded in silently.

---

## 3. Success criteria

- `docs-site` 4로케일 `cli-reference/web.md` 에 탭 수가 없다.
- 남은 수·이름 자리를 `consoleTabs()` 와 대조하는 가드 테스트가 존재하고 통과한다.
- 가드를 일부러 깨뜨리면(문서의 수/이름을 한 글자 바꾸면) 실패한다 — 공허한 초록이 아님을 보인다.
- 오탐 목록 3자리에 대해 가드가 침묵한다.
- 12개 문서의 로케일 패리티가 유지된다.
- `hugo` 빌드가 경고 없이 통과한다.

기계적으로 확인 가능한 형태(명령 + 기대 출력)는 `acceptance.md` 에 있다.

---

## 4. 경계 — t509 와의 관계

카드 t509 의 반경은 웹 콘솔의 **codex 패널**이었고, `cli-reference` 는 그 밖이었다. 그 경계는 옳았고,
t530 이 존재하는 이유가 그것이다. 본 SPEC 은 그 경계를 다시 다투지 않으며 codex 패널로 넓히지 않는다.
`advanced/moai-web-console.md:137` 의 `codex 설정 12개` 는 codex 패널 내부 사실이므로 **본 SPEC 범위 밖**이다.

---

## 5. 스크린샷 범위 판단 (결정 기록)

**결정: 본 카드에 넣지 않는다. 별건 카드 2장으로 분리한다.**

관측 사실:

- `assets/images/moai-web-settings.png` 는 README 4본의 411행에서 참조되며 11탭 시절 이미지다.
- t509 가 alt 텍스트에서 수를 제거했으나(`.moai/reports/oss-docs-v311/drafts/README.ko.md:372` 의 `10개 설정 탭` → 현행 alt 는 수 없음) 이미지 자체는 그대로다.
- 재생성 절차는 저장소 어디에도 기록되어 있지 않다(`moai-web-settings` 를 참조하는 문서 어디에도 캡처 방법이 없다).

근거:

1. **산출물 종류가 다르다.** 나머지는 텍스트이고 명령 한 줄로 검증되지만, 이미지는 콘솔을 띄우고 사람이 캡처해야 하며
   기계적 AC 를 붙일 수 없다. 한 카드에 섞으면 본 카드의 acceptance 가 산문 단정으로 내려앉는다.
2. **절차의 부재가 별개의 결함이다.** 절차가 없으면 이번에 한 번 다시 찍어도 다음에 또 같은 자리에서 낡는다.
   고쳐야 할 것은 이미지가 아니라 절차의 부재이며, 그것은 자기 카드를 가질 값어치가 있다.
3. **가시성.** 조용히 접으면 "이미지는 여전히 낡았다" 는 사실이 카드 마감과 함께 사라진다.

후속 카드 제안(리드가 큐에 올릴 것):

- `[t530-후속-1]` `moai web` 설정 화면 스크린샷 재생성 절차 기록 — 어떤 창 크기·로케일·프로파일로 찍는지, 어디에 두는지.
- `[t530-후속-2]` 위 절차로 `assets/images/moai-web-settings.png` 재촬영(현행 14탭).

---

## 6. Out of Scope

본 SPEC 은 문서의 사실 정합성만 다룬다. 다음은 **명시적으로 범위 밖**이다.

### Out of Scope — 코드 동작

- `consoleTabs()` 에 탭을 더하거나 빼지 않는다. 탭 구성은 이 카드가 건드리는 대상이 아니라 **읽는 대상**이다.
- 콘솔 UI·CSS·라우팅을 수정하지 않는다.

### Out of Scope — codex 패널 (t509 반경)

- `advanced/moai-web-console.md:137` 의 codex 설정 개수(12) 는 다루지 않는다.
- codex 탭의 읽기 전용 성격을 설명하는 문단을 재작성하지 않는다.

### Out of Scope — 스크린샷

- `assets/images/moai-web-settings.png` 를 다시 찍지 않는다(§5 근거).
- 스크린샷 재생성 절차 문서를 이 카드에서 쓰지 않는다 — 후속 카드 소관.

### Out of Scope — 다른 수의 드리프트

- 스킬 수(`Eleven ref skills`), 에이전트 수(12), statusline 키 수(16), 칸반 용어 수(9) 등 같은 형태의 다른 손 계수는
  본 카드에서 손대지 않는다. 같은 병이지만 반경을 넓히면 이 카드가 전수 문서 감사로 변한다.

### Out of Scope — 문서 전반 재작성

- 대상 12파일의 문장 구조·어조를 재작성하지 않는다. 수·이름과 그 수를 담고 있던 최소 구절만 고친다.
- 템플릿 미러(`internal/template/templates/**`) 는 대상이 아니다 — README 와 docs-site 는 템플릿에 미러되지 않는다.

---

## 7. 잔여 위험

- **화이트리스트 밖 방향** — 가드의 대상 파일이 명시 12파일 목록이므로, **새 문서가 탭 수를 새로 적으면 잡히지 않는다.**
  오탐을 없애기 위해 치른 값이며, 대안(전 문서 스캔)은 `plugins.md:100` 류를 계속 잡는다.
- **화이트리스트 안 방향 — 이쪽이 이 카드를 실제로 막았던 방향이다.** 대상 파일 안에도 정당한 비-설정탭 계수가 있다:
  `…/advanced/moai-web-console.md:149` 의 codex 패널 근거 문장이 감사 탭·MCP 탭 **둘**을 센다. 파일 화이트리스트도
  수-명사 인접도 이 줄을 거르지 못한다 — **두 장치가 동시에 실패한다.** 그래서 가드는 명시 허용 목록을 들고,
  인수 판정은 정규식이 아니라 리터럴 집합으로 한다(§1.3 두 층 표). 남는 값: 허용된 줄에 언젠가 진짜 설정 탭 계수가
  섞여 들어오면 가드가 침묵한다. 허용 항목이 4줄뿐이고 각 줄에 이유가 붙어 있어 사람이 검토할 수 있다는 것이 그 대가다.
- **낱말 축은 열렸고, 남는 것은 유지 비용이다.** 이전 판은 "낱말 축은 열 수 없다 — 실측으로 확인했다" 고 적었으나
  그 측정은 결론을 뒷받침하지 않았다: 근거로 든 세 자리(`탭 두 곳`·`two tabs`·`两个标签页`)가 전부 `:149` 의
  codex 근거 문단이고, 이미 채택한 `codex` 허용 규칙이 덮는 줄이었다. 허용 규칙과 서수 접두(`第`) 제외를 적용해
  다시 재니 **허용되지 않은 오탐은 0**(남은 6행 전부가 열거된 자리)이어서, 축을 열었다(`plan.md §A.4`).
  남는 위험은 오탐이 아니라 **로케일별 수사 클래스의 유지 비용**이다 — 수사 목록은 손으로 적힌 집합이므로,
  새 언어나 새 표기(`seventeen`, `열일곱`, `十七` 밖의 형태)가 들어오면 클래스를 손봐야 하고, 그 손봄을 잊으면
  그 표기만 조용히 빠진다. 그 경우에도 탭 **수**가 바뀌면 N2(이름 대조)가 붉어지므로, 노후화는 이름 축에서 드러난다.
- **허용 규칙 1개가 면제하는 줄은 16개다.** 규칙 수와 면제 표면은 다른 값이다(실측: `grep -ci codex` → 4로케일 각 4행).
  허용된 줄 안에 언젠가 진짜 설정 탭 계수가 섞여 들어오면 가드가 침묵한다. 가드가 규칙 수와 면제 줄 수를
  **둘 다 보고**하므로(AC-TCD-007) 그 표면이 커지는 것은 보이지만, 커진 표면 안의 내용까지 보지는 못한다.
- D군 목록 밖 산문 4자리(`…:165` 의 `3rd Party LLM 탭`)는 가드가 보지 않는다. 같은 변경에서 손으로 고친다.
- **`3rd Party LLM` 이 의도한 이름이었을 가능성.** 본 카드는 문서를 렌더 라벨(`GLM Settings` 계열)에 맞추기로 결정했다
  (`plan.md §C`, 2026-09-12). 만약 `3rd Party LLM` 쪽이 본래 의도한 이름이었다면, 그것을 바로잡는 일은
  `internal/web/assets/i18n.js` 의 4로케일 값(`:229`, `:1096`, `:1852`, `:2608`)을 고치는 **콘솔 변경**이며
  본 카드 범위 밖이다(§6 Out of Scope — 코드 동작). 그때는 문서가 아니라 코드를 고치는 별건 카드를 세우고,
  본 카드가 넣은 N2 가드가 그 변경을 자동으로 잡아 문서 쪽 12자리를 함께 고치라고 알려 준다.
