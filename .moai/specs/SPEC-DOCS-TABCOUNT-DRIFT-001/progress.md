# SPEC-DOCS-TABCOUNT-DRIFT-001 — progress

카드 t530 / 워크트리 `.claude/worktrees/t530` / 브랜치 `WT-web-tab-docs` / base `1d150a27d`

## §E.1 Plan-phase Audit-Ready Signal

- 산출물: `spec.md`, `plan.md`, `acceptance.md`, 열거 산출물 `.moai/reports/t530/tab-count-sites.md`,
  그리고 인수 판정의 패턴 원본 `.moai/reports/t530/count-literals.txt`(16줄).
- 기준값 14는 이 트리에서 재도출했다. **원천 1(구현) + 미러 1(테스트 리터럴) + 둘의 일치 테스트 1** 이며,
  "서로 독립인 세 측정" 이 아니다(`spec.md §1.1`).
- 미결 없음: llm 탭 이름은 2026-09-12 에 "문서를 코드에 맞춘다" 로 해소됨 — 근거는 `plan.md §C`.
- 범위 판단 1건: 스크린샷 재생성은 별건으로 분리, 근거는 `spec.md §5`.

### plan-audit iter1 (FAIL 0.66) 수리 라운드 — 재측정 기록

감사 보고: `.moai/reports/t530/plan-audit.md`. 아래 수치는 **수리 시점에 다시 잰 값**이다(base `1d150a27d`).

| 측정 | 명령 | 관측 |
|---|---|---|
| A군 리터럴 적중 | `head -4 count-literals.txt > /tmp/t530-a.txt; grep -rnF -f /tmp/t530-a.txt <4파일> \| wc -l` | `4` (en 포함 — 종전 창 기반 정규식은 3) |
| 전체 리터럴 적중 | `grep -rnF -f count-literals.txt <12파일> \| wc -l` | `20` (집합이 열거 20자리와 동일) |
| D군 이름 적중 | `grep -rn '3rd Party LLM\|서드파티 LLM\|サードパーティ LLM\|第三方 LLM' <8파일> \| wc -l` | `12` |
| 가드 층 숫자 스윕(낱말 경계 포함) | `grep -rnE '[0-9]+[^.。]{0,12}(\btabs?\b\|탭\|タブ\|标签页)' <12파일>` | `15`행 = 열거 14 + 허용 1(`ja/advanced/moai-web-console.md:149`) |
| 낱말 경계 없는 스윕의 오탐 | 같은 스윕에서 `\b` 제거 | `README.md:698` (`selec**tab**le`) 추가 적중 |
| 가드 선택자(현재) | `go test ./internal/web/ -run 'TestDocsTabContract' -v` | `testing: warning: no tests to run` / `PASS` / `ok … [no tests to run]`, 종료코드 `0` |
| 패리티 대조군 | `git diff --name-only 1d150a27d -- <README 4본> docs-site/content \| wc -l` | `0` → 측정 불가 = AC-TCD-010 FAIL |
| 스크린샷 | `git diff --quiet 1d150a27d -- assets/images/` | 종료코드 `0` (무변경) |
| 열거 행 수 | `grep -c '^| [ABCD][0-9]' tab-count-sites.md` | `32` (28 → 32, D군 산문 4자리 승격) |
| 열거 경로 완전성 | 표에서 뽑은 고유 문서 경로 수 + 실재 확인 | `12`, 누락 없음 |

**Gap(수리 라운드에서 관측하지 못한 것)** — 조용히 통과시키지 않는다:

- **가드 자체의 동작**: `internal/web/docs_tab_contract_test.go` 는 아직 없다. AC-TCD-004~009 는 실행으로 판정하지 못했고,
  RED-now 셀은 "가드 부재" 로만 기록했다. 변이체 A(README 만 읽는 가드)·B(서브테스트 없음)가 실제로 막히는지는
  가드 작성 후 M5 에서 확인해야 결론이 확정된다.
- **hugo 무경고의 인과**: 수정 후에도 무경고인지는 수정이 없어 확인 불가(RG-TCD-002).
- **넓힌 정규식의 새 미열거 적중**: 낱말 축을 열면 `ko:149 탭 두 곳`·`en:149 two tabs`·`zh:149 两个标签页` 가 모두 적중함을
  실측했다. 이 때문에 낱말 축은 정규식에 넣지 않기로 했고(`plan.md §A.4`), 그 결정의 대가는 `spec.md §7` 에 적었다.

### plan-audit iter2 (FAIL 0.775) 수리 라운드 — 재측정 기록

감사 보고: `.moai/reports/t530/plan-audit-iter2.md`(iter1 보고 `plan-audit.md` 는 보존). HEAD `28c1ea062`.
왼쪽 끝은 `git merge-base develop HEAD` = `1d150a27d4c5cdeedb37df19b7a4a025e5dd2c09`(측정 시점 값).

| 측정 | 명령 | 관측 |
|---|---|---|
| **낱말 축 원시 스윕** | 로케일별 수사 클래스 + 낱말 경계 탭 명사, 대상 12파일 | **8행** |
| − 서수 접두(`第`) 제외 | 같은 스윕에서 `第[一二三四五六七八九十]` 치환 후 재적용 | **7행** (`zh/advanced/…:165` 의 `第三方` 빠짐 — 단독 검증도 `0`) |
| − `codex` 허용 규칙(**파일 범위 포함**) | `grep -vE '^docs-site/content/[a-z]+/advanced/moai-web-console\.md:[0-9]+:.*codex'` | **6행 = A2·A4·B4·C1·C2·C4**, 허용되지 않은 오탐 **0** |
| 같은 규칙을 **파일 범위 없이** 적용 | `grep -v 'codex'` | **3행** — README 3자리가 함께 면제된다(파일 범위가 필수인 이유) |
| 허용 규칙의 면제 표면 | `grep -ci codex` × 4 console 파일 | 각 **4**, 합계 **16줄** (규칙 수 1 과 다른 값) |
| merge-base | `git merge-base develop HEAD` | `1d150a27d4c5…` |
| 좁힌 pathspec, 카드 기여 | `git diff --name-only <merge-base> -- <12파일>` | **0** → 측정 불가 = AC-TCD-010 FAIL (붉은 이유가 "이 카드가 아직 안 고쳤다" 로 바뀜) |
| 좁힌 pathspec, develop 쪽 churn | `git diff --name-only <merge-base> develop -- <12파일>` | **8파일** — 흡수 전에는 보이고 흡수 후 merge-base 전진으로 사라지는 값 |
| 넓은 pathspec 대비(이전 판의 결함) | `git diff --name-only 1d150a27d develop -- <README 4본> docs-site/content` | **155파일**, `1d150a27d..develop` **40커밋** |
| 스크린샷 | `git diff --quiet <merge-base> -- assets/images/` | 종료코드 `0` (무변경) |
| 열거 행/경로(같은 영역) | `grep '^| [ABCD][0-9]'` → 행 수 / 그 행에서 뽑은 경로 수 | **32행 / 12경로**, 누락 0 |
| 구조 불변식 | REQ / AC / RG / `### Out of Scope —` 계수 | **9 / 12 / 3 / 5**, `status:` 는 `spec.md` 에만, clarification 토큰 0 |

**Gap(이 라운드에서 관측하지 못한 것)**:

- **가드는 여전히 없다.** AC-TCD-004~009·012 의 가드 쪽 단정은 명세 수준 논증이며, 낱말 축·서수 제외·파일 범위
  허용 규칙은 **shell 재현**으로만 확인했다. Go 구현이 같은 집합을 내는지는 M1 에서 확인해야 결론이 확정된다.
- **병합 후 거동은 시뮬레이션이다.** develop 을 실제로 흡수하지 않았다(감사 트리를 변형하지 않는다는 같은 이유).
  "흡수 후 merge-base 가 전진해 남의 커밋이 범위에서 빠진다" 는 git 의 성질이지 이 트리에서 관측한 사실이 아니다.
- **hugo 는 이번에도 실행하지 않았다**(RG-TCD-002).
- **낱말 수사 클래스의 완전성은 미확인.** 오늘 0 오탐을 낸 클래스가 앞으로 들어올 표기까지 덮는다는 보장은 없다 —
  `spec.md §7` 의 유지 비용 항목이 이 미확인을 가리킨다.

### plan-audit iter3 (PASS 0.8375) — 판정과 인계되는 부채

감사 보고: `.moai/reports/t530/plan-audit-iter3.md`. 점수 추이 0.66 → 0.775 → **0.8375**(Tier M 임계 0.80),
must-pass 실패 0. D1~D4 전부 REPAIRED 이며, 그중 셋은 감사가 문서 독해가 아니라 **실행**으로 확인했다 —
재현 명령 5 축어 실행 6행, 스크래치 트리에서 카드 처분대로 편집 후 6 → 0(green path 실재),
그 초록 트리에 V4 변이 주입 시 1행 적중(iter2 가 통과시킨 변이체가 이제 잡힌다).

**[부채 N1 — run-phase M1 의 입구 작업] 가드의 낱말 *클래스* 를 구속하는 AC 가 없다.**

`REQ-TCD-005` 는 "숫자와 낱말 모두" 를 명시하지만, 그것을 기계적으로 구속하는 자리가 없다:
`AC-TCD-012` 의 When 은 셸 파이프라인이어서 **셸**을 구속하고 Go 가드를 구속하지 않으며,
`AC-TCD-006 V4` 는 변이가 `fourteen` 한 건이다. 실측으로도 acceptance.md 가 이름을 대는 수사 낱말은
`nine` / `fourteen` / `열네` 뿐이고 `十四`·`九个`·`아홉` 은 어디에도 없다.

결과: 낱말 클래스를 `{fourteen}` 으로만 구현한 가드가 전 기준을 통과하면서 6자리 중 5자리를 놓친다 —
**이 카드가 지난 청소의 실패 원인으로 지목한 바로 그 부류**다.

네 번째 감사 반복은 돌리지 않는다(상한 3 도달, 임계는 이미 넘었다). PASS-with-debt / 범위 축소 /
운영자 override 는 임계 **미달** 시의 선택지라 여기 해당하지 않는다. 고칠 것은 `acceptance.md` 한 문장이고
그것이 구속하는 대상이 M1 이 지금 쓰려는 가드 파일 자신이므로, M1 의 첫 작업으로 라우팅한다.
**그때까지 이 축의 방어는 사람 독해뿐이다** — 그 사실을 여기 남기는 것이 조건이다.

> **[2026-09-13, 0.3.1] 이 부채는 닫혔다.** M1 입구를 기다리지 않고 지금 수리했다 — 아래 「N1 수리 (0.3.1)」 절.
> 위 문단은 당시 판단의 기록으로 보존한다.

기타(optional, 차단 아님): N2 `§2.1` RED-now 4요소 중 축어 stdout·종료코드가 AC-004 외 11개 셀에서 누락
(감사가 셀을 전부 재실행해 재현했고, DoD 2번이 run-phase 에 그 기록을 이미 요구한다) · N3 수사 토큰에 낱말
경계가 없어 `세부`·`네트워크`·`열기`·`phone` 이 적중(오늘 오탐 실측 0, 미래 노출) · N4 재현 명령 5 의 2단계
정규식이 1단계의 부분집합이 아니라 `twenty tabs` 가 소실 · N5 D12 트레일러가 전이를 운반하지 않는 커밋에
붙어 린트가 여전히 `OwnershipTransitionUnmeasured` INFO 를 낸다(`c205eeeff` 는 이미 착지, 부채로 기록).

### N1 수리 (0.3.1) — 낱말 클래스가 이제 가드에 걸린다

부채로 인계했던 N1 을 M1 입구가 아니라 **지금** 닫았다. 고친 것은 인수 층 두 자리이며, 가드는 여전히 없다.

**언제·왜 그때인가** — 2026-09-13, run-phase M1 이 아니라 **plan-phase 수리 중**에 닫았다. 이유는
`acceptance.md` 가 이미 소유 에이전트(manager-spec) 아래 열려 있었다는 것이다. 같은 편집을 M1 에서 하려면
그 맥락을 다시 열어야 하고, 그 비용이 지금 쓰는 비용보다 크다.

조기 종결이 **편해서**가 아니라 **안전해서** 가능했던 근거 둘:

- **좁히는 변경이지 푸는 변경이 아니다.** 기존 기준 둘을 조였고, 새 AC 를 만들지 않았으며, Tier M 예산도 그대로다
  (REQ 9 / AC 12 / RG 3). 통과 범위를 **넓히는** 사후 편집이었다면 판단이 달랐다.
- **plan-audit PASS(0.8375)는 수리 *이전* 아티팩트에 대해 잰 값이다.** 이 절은 그 판정 뒤에 들어간 변경이므로,
  PASS 가 이 문장들을 보고 내려진 것이 아니다. sync-auditor 는 갱신된 `acceptance.md` 를 자기 baseline 으로 삼아
  다시 검증한다.

| 자리 | 변경 |
|---|---|
| `AC-TCD-012` | [HARD] And 절 신설 — 가드가 `word-axis hit <경로>: <구절>` 로 적중 집합을 찍고, **M1 직후·M2 이전** 트리에서 그 출력이 정확히 6줄이며 A2·A4·B4·C1·C2·C4 와 같다. 다섯 토큰(`nine` en · `九` zh · `fourteen` en · `열네` ko · `十四` zh)이 각각 한 줄 이상. 5줄 이하 또는 토큰 하나라도 0줄이면 FAIL |
| `AC-TCD-006` | 변이 4종 → **6종** (V5 ko `열네`, V6 zh `设置九个标签页`). V6 은 V3(숫자 `14 个`)과 다른 축이다 |
| `plan.md §A.5`, `§D M1` | 같은 내용으로 훑음 — 가드가 집합을 찍는 것, 그리고 M2 가 문서를 고치기 **전에** 재야 한다는 순서 전제 |

죽는 변이체는 **낱말 클래스를 `{fourteen}` 하나로만 구현한 가드**다. 그 가드는 V1·V2·V3·V4 와
`AC-TCD-012` 의 셸 Then 절(파이프라인 실행)을 전부 통과하지만, 가드 축 절에서 `fourteen` 자리 2줄만 찍어 FAIL 하고
V5·V6 에서도 FAIL 한다. 종전에는 그 가드가 **모든 기준을 통과하면서** 열거된 낱말 표기 6자리 중 5자리를 놓칠 수 있었다.

두 셀 규율(`verification-completeness.md` §2):

- **RED-now (가드 축)**: 붉은 이유는 **`internal/web/docs_tab_contract_test.go` 의 부재**다. 그 명령을 실행해 본 적이 없으므로
  적중 수를 수치로 적지 않는다 — 선택자가 아무것도 고르지 않아도 `ok` 와 종료코드 0 이 나오는 트리라(AC-TCD-004),
  거기서 얻은 0 은 "적중이 없다" 가 아니라 "잰 적이 없다" 다.
- **green path**: **M1**. 가드가 그 출력을 내기 시작하는 자리이며, 문서 편집(M2~M4)과 독립이다.

구조 불변식 재확인: REQ **9** / AC **12** / RG **3** (Tier M 상한 16, 독립 적용) · `### Out of Scope —` **5** ·
`status:` 는 `spec.md` 한 곳 · clarification 토큰 **0**. 새 AC 를 만들지 않고 기존 둘을 조인 결과다.

**N1 을 그대로 두었다면** — 가드를 쓰는 사람이 `REQ-TCD-005` 본문을 읽지 않고 V4 만 보고 클래스를 좁혀도
카드가 초록으로 마감됐다. 그 방어가 사람 독해뿐이던 상태가 이 수리로 기계 단정으로 바뀌었다.

### N5 처분 (리드 판정) — 이력을 고치지 않는다

`(none) → draft` 전이를 운반하는 커밋은 `c205eeeff` 인데 `Authored-By-Agent:` 트레일러는 그 커밋이 아니라
`8336f0ac7`(`version` 만 바꾼 커밋)에 붙었다. 그래서 `moai spec lint` 가 이 SPEC 에 대해
`OwnershipTransitionUnmeasured` INFO 를 계속 낸다.

**리드 판정: 이력을 다시 쓰지 않는다.** `c205eeeff` 는 이미 착지했고 소견은 비차단 INFO 다
(`spec-frontmatter-schema.md` § OwnershipTransitionRule — `--strict` 도 INFO 를 에러로 올리지 않는다).
전이 자체가 사라진 것이 아니라 **귀속이 측정되지 않은** 상태이며, 3단계 마감(plan→run→sync) 아래에서
이 SPEC 의 소유권 전이를 실제로 운반하는 다음 자리는 sync 커밋이다 — 그 커밋이 트레일러를 제 커밋에 달면
그 전이는 귀속된다. 여기 남는 INFO 는 그때까지의 잡음이고, 이 기록이 그 잡음의 정체다.

**린트 축 실측 (레인, 2026-09-13) — 수리 라운드가 미측정으로 남긴 축을 닫는다.**
`manager-spec` 은 240초 bound 에서 `moai spec lint` 가 timeout(exit 124) 해 이 축을 Gap 으로 남겼다.
더 넉넉한 bound 로 다시 돌려 완주시켰다(`exit=0`).

| 측정 | 관측 |
|---|---|
| 이 SPEC 의 소견 전량 | **1건** — `INFO OwnershipTransitionUnmeasured`(`(none) → draft`). 즉 N5 그 자체이고 그 외에는 없다 |
| `FrontmatterInvalid` / `MissingExclusions` | 이 SPEC 에 **0건** |
| 저장소 전체 `OwnershipTransitionUnmeasured` | **203건** |

203 이라는 값이 리드 판정을 뒷받침한다 — 이 소견은 **이 카드가 만든 결함이 아니라 저장소 전반의 상시 잡음**이고,
한 SPEC 의 이력을 고쳐서 줄일 수 있는 종류가 아니다. 트레일러 규약이 도입되기 전에 착지한 전이가 대다수다.
이 카드가 할 수 있는 일은 **앞으로의 전이를 귀속시키는 것**뿐이며, 그 자리가 위에 적은 sync 커밋이다.

## §E.2 Run-phase Evidence

### M1 — 가드 작성 (실패하는 상태로)

산출물: `internal/web/docs_tab_contract_test.go` (신규 1파일). 기존 Go 소스는 건드리지 않았다 —
`consoleTabs()` 와 `wantTabOrder` 는 이 카드의 **읽는 대상**이지 고치는 대상이 아니다(`spec.md §6`).

측정 트리: `.claude/worktrees/t530`, 브랜치 `WT-web-tab-docs`, M1 착수 시 HEAD `0516119e3`.
아래 모든 수치는 그 트리에서 이번 실행으로 관측했다.

#### RED 관측 (E8) — 가드는 현재 문서에 대해 붉다

`go test ./internal/web/ -run 'TestDocsTabContract' -v` 축어(발췌):

```
--- FAIL: TestDocsTabContract (0.14s)
    --- FAIL: TestDocsTabContract/literals (0.00s)
    --- FAIL: TestDocsTabContract/allowlist (0.13s)
    --- FAIL: TestDocsTabContract/names (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/web	0.763s
```

세 층이 각각 무엇을 잡았는지:

| 층 | 관측 |
|---|---|
| `literals` | `swept 12 files against 16 enumerated literals` · 열거 리터럴 **18**건 잔존 |
| `allowlist` | `allowed rules 1` · `allowed lines 16` · 미허용 적중 **20**건(숫자 14 + 낱말 6) |
| `names` | 8자리 전부 `tab 4 name = "3rd Party LLM", console renders "GLM Settings"` |

**`literals` 18 vs 열거 20 — 차이의 정체(추정 아님, 대조 결과).** 셸 재현 명령 2는 `grep -rnF` 로 **행**을
세어 20을 내고, 가드는 (파일, 리터럴) 쌍을 센다. 같은 파일 안에서 같은 리터럴이 두 행에 있는 두 자리
(`ko/advanced/…` 의 `14개 탭` :25·:128, `zh/advanced/…` 의 `14 个标签页` :25·:128)가 각각 1건으로 접힌다.
20 − 2 = 18. 가드의 단정은 "0건" 이므로 이 접힘은 판정에 영향이 없고, 20이라는 수치는 AC-TCD-002 의
셸 축이 계속 소유한다.

#### AC-TCD-012 가드 축 — 낱말 축 적중 집합 (M2 이전 트리에서 잰 값)

`go test ./internal/web/ -run 'TestDocsTabContract/allowlist' -v 2>&1 | grep '^word-axis hit '` 축어:

```
word-axis hit README.md: fourteen tabs
word-axis hit README.ko.md: 열네 개 탭
word-axis hit README.zh.md: 十四个标签页
word-axis hit docs-site/content/en/cli-reference/web.md: nine settings tabs
word-axis hit docs-site/content/zh/cli-reference/web.md: 九个标签页
word-axis hit docs-site/content/en/advanced/moai-web-console.md: fourteen tabs
```

행 수 **6**. 열거와의 대조 — 6줄이 A2·A4·B4·C1·C2·C4 와 **같다**:

| 줄 | 열거 자리 | 토큰 | 로케일 |
|---|---|---|---|
| `en/cli-reference/web.md: nine settings tabs` | A2 | `nine` | en |
| `zh/cli-reference/web.md: 九个标签页` | A4 | `九` | zh |
| `en/advanced/moai-web-console.md: fourteen tabs` | B4 | `fourteen` | en |
| `README.md: fourteen tabs` | C1 | `fourteen` | en |
| `README.ko.md: 열네 개 탭` | C2 | `열네` | ko |
| `README.zh.md: 十四个标签页` | C4 | `十四` | zh |

토큰별 줄 수 실측: `nine 1` · `九 1` · `fourteen 2` · `열네 1` · `十四 1` — 다섯 토큰 모두 1줄 이상.
ja 줄이 없는 것은 누락이 아니라 실측이다(ja 4자리는 전부 숫자 표기).

**출력이 `t.Logf` 가 아니라 `fmt.Printf` 인 이유**: AC-TCD-012 가 `grep '^word-axis hit '` 로 **0열에 앵커**해
세는데, `t.Logf` 는 `-v` 에서 들여쓰기와 `file:line` 접두를 붙여 그 앵커에 걸리지 않는다. 나머지 토큰
(`swept` / `allowed` / `extracted`)은 앵커 없는 grep 이라 `t.Logf` 로 둔다.

#### 변이 1 — 낱말 클래스를 `{fourteen}` 하나로 좁힌 가드는 죽는다

`wordNumeralRe` 의 로케일 클래스를 `fourteen` 단독으로 바꾼 뒤 같은 명령:

```
word-axis hit README.md: fourteen tabs
word-axis hit docs-site/content/en/advanced/moai-web-console.md: fourteen tabs
```

행 수 **2**. `nine` · `九` · `열네` · `十四` 가 각각 **0줄**이다. AC-TCD-012 의 6줄 단정과 다섯 토큰 단정
양쪽에서 FAIL 한다. 변이 후 원본 정규식으로 되돌렸고, 되돌린 뒤 다시 재어 위 6줄·다섯 토큰이 그대로임을
확인했다(위 표가 그 재측정값이다).

이것이 plan-audit iter3 의 부채 N1 이 지목한 변이체이며, 명세 텍스트가 아니라 **이 구현**에 대해 죽는 것을
보인 것이 M1 의 몫이었다.

#### 변이 2 — 읽지 못한 파일은 skip 이 아니라 FAIL 이다 (AC-TCD-005)

대상 12파일 중 하나를 존재하지 않는 경로로 바꾸고 `literals` 를 돌렸다:

```
docs_tab_contract_test.go:127: docs tab guard: cannot read docs-site/content/zh/cli-reference/web-MUTANT-UNREADABLE.md: open ../../docs-site/content/zh/cli-reference/web-MUTANT-UNREADABLE.md: no such file or directory (a guard that cannot read its target is not passing, it is blind)
--- FAIL: TestDocsTabContract/literals (0.00s)
```

`SKIP` 없음, `swept` 줄 없음 — 카운터가 **읽기 성공 뒤에** 증가하므로 실패한 읽기는 수에 들지 않고,
`t.Fatal` 이 판정 전 이탈을 막는다. 변이 후 원본 경로로 되돌렸다.

#### 기존 테스트 무회귀 (DoD 3)

`go test ./internal/web/ -v` 의 실패 줄 전량:

```
    --- FAIL: TestDocsTabContract/allowlist (0.07s)
    --- FAIL: TestDocsTabContract/literals (0.00s)
    --- FAIL: TestDocsTabContract/names (0.00s)
--- FAIL: TestDocsTabContract (0.08s)
```

실패는 새 가드 4줄뿐이다 — 기존 테스트는 하나도 깨지지 않았고, 이 FAIL 은 M1 이 의도한 RED 다.

#### 빌드 · 정적 검사

| 명령 | 관측 |
|---|---|
| `go build ./internal/web/` | exit `0` |
| `go vet ./internal/web/` | 출력 없음, exit `0` |
| `gofmt -l internal/web/docs_tab_contract_test.go` | 출력 없음 |

#### Gap — M1 이 관측하지 못한 것

- **GREEN 을 보지 못했다.** 세 층 모두 아직 붉다. 초록 전이는 M2~M4 가 문서를 고친 뒤 M5 에서 관측된다.
  이 M1 은 RED 만 근거로 가진다.
- **AC-TCD-006 V1~V6 미실행.** 변이 6종은 **가드가 초록인 트리**에서 FAIL→복구를 보이는 시험이라 M5 소관이다.
  M1 에서 돌린 두 변이는 성격이 다르다 — 가드 구현 자신(낱말 클래스)과 읽기 실패 경로를 겨눈 것이다.
- **`allowed lines 16` 은 이 트리의 실측이자 baseline 상수다.** 문서 편집이 `codex` 줄 수를 바꾸면 이 단정이
  붉어진다. M2~M4 의 처분 대상에 `:149` 문단이 없으므로 바뀌지 않을 전망이지만, 전망은 관측이 아니다.
- **hugo 미실행**(RG-TCD-002) — M6 소관.

#### 잔여 위험 — 가드가 보지 않는 자리 (M1 에서 발견, 범위 밖)

`names` 층은 로케일 문서의 `현지어(English)` 항목에서 **괄호 안 English 만** 대조하고 앞의 현지어는 대조하지
않는다. 그 자리를 대조하도록 만들면 **이 카드가 고칠 수 없는 붉음**이 생기기 때문이다 — 실측:
`docs-site/content/zh/advanced/moai-web-console.md:130` 은 1번 탭을 `用户信息（Identity）` 로 적는데
`internal/web/assets/i18n.js:2708` 의 zh 라벨은 `身份` 다. 이것은 D군(`3rd Party LLM`)과 **다른 자리의 이름
드리프트**이고, 본 카드의 M4 는 D군 12자리만 고치므로 그 축을 열면 green path 가 이 카드 밖을 지나간다
(`verification-completeness.md` §2 가 실격시키는 모양). 따라서 가드는 그 축을 **열지 않고, 열지 않았다는
사실을 여기 적는다.**

괄호가 없는 항목(현지어 단독: ko `피드백`·`품질 게이트`, ja `フィードバック`·`品質ゲート`,
zh `反馈`·`质量门禁`)은 i18n.js 의 해당 로케일 라벨과 대조하며, 여섯 자리 모두 현재 일치한다.

**후속 카드 요청(리드에게)**: zh 탭 이름 현지어 축의 드리프트 — `用户信息` vs i18n `身份`.
다른 로케일·다른 탭에도 같은 형태가 있는지는 재지 않았다(가드가 그 축을 열지 않으므로 측정 자체가 없다).

## §E.3 Run-phase Audit-Ready Signal

_<pending — M1 만 완료. 이 신호는 run-phase 전체(M1~M5)가 닫힐 때 채운다.>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
