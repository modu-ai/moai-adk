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

### M2 — A군 4자리: 틀린 수(9) 제거

측정 트리: `.claude/worktrees/t530`, 브랜치 `WT-web-tab-docs`, M2 직전 HEAD `d85ac3e9e`.

#### 처분 — 수를 지운다(고쳐 쓰지 않는다)

| 자리 | 파일 `:53` | 이전 | 이후 |
|---|---|---|---|
| A1 | `docs-site/content/ko/cli-reference/web.md` | `설정 9개 탭.` | `설정 탭 화면.` |
| A2 | `docs-site/content/en/cli-reference/web.md` | `The nine settings tabs.` | `The settings tabs.` |
| A3 | `docs-site/content/ja/cli-reference/web.md` | `設定 9 タブ。` | `設定タブ画面。` |
| A4 | `docs-site/content/zh/cli-reference/web.md` | `设置九个标签页。` | `设置标签页界面。` |

`?tab=` / `?profile=` 설명은 네 자리 모두 그대로 두었다 — 그 부분이 이 표 행의 실제 정보다.

#### RED (M2 이전, HEAD `d85ac3e9e`) — AC-TCD-001 축어

```
$ head -4 .moai/reports/t530/count-literals.txt > /tmp/t530-a.txt
$ wc -l < /tmp/t530-a.txt
       4
$ grep -rnF -f /tmp/t530-a.txt docs-site/content/{ko,en,ja,zh}/cli-reference/web.md | wc -l
       4
```

리터럴 파일이 4줄임을 함께 쟀다 — 빈 패턴 파일도 `0` 을 내므로 `0` 만으로는 통과 근거가 되지 않는다(AC-TCD-001).

#### GREEN (M2 이후) — AC-TCD-001 축어

```
$ head -4 .moai/reports/t530/count-literals.txt > /tmp/t530-a.txt
$ wc -l < /tmp/t530-a.txt
       4
$ grep -rnF -f /tmp/t530-a.txt docs-site/content/{ko,en,ja,zh}/cli-reference/web.md | wc -l
       0
```

#### 가드 상태 — 세 층 모두 여전히 RED (예정된 부분 진척)

```
$ go test ./internal/web/ -run 'TestDocsTabContract' -v
    docs_tab_contract_test.go:136: swept 12 files against 16 enumerated literals
    docs_tab_contract_test.go:145: 14 enumerated literal(s) still present; the run phase removes the count, it does not rewrite it
    docs_tab_contract_test.go:281: allowed rules 1
    docs_tab_contract_test.go:282: allowed lines 16
    docs_tab_contract_test.go:303: numeral sweep found 16 unallowed hit(s) (4 on the word axis); every one is a count that can drift
--- FAIL: TestDocsTabContract (0.08s)
    --- FAIL: TestDocsTabContract/literals (0.00s)
    --- FAIL: TestDocsTabContract/allowlist (0.07s)
    --- FAIL: TestDocsTabContract/names (0.00s)
```

M1 baseline(리터럴 18 / 스윕 20 · 낱말 6) 대비 리터럴 18→14, 스윕 20→16, 낱말 축 6→4.
낱말 축에서 빠진 둘은 A2(`nine settings tabs`)·A4(`九个标签页`) 이며, 이는 A군 처분과 정확히 대응한다.
`allowed rules 1` / `allowed lines 16` 은 불변 — A군 편집이 `codex` 허용 면을 건드리지 않았다는 관측이다.

#### Gap — M2 가 관측하지 못한 것

- 세 층 모두 아직 붉다. B·C군(M3)과 D군(M4)이 남아 있으므로 예정된 상태다.
- `names` 층은 A군과 무관하다(cli-reference 4본은 이름 목록을 담지 않는다) — M2 는 그 층에 어떤 영향도 주지 않았고, 실제로 출력이 변하지 않았다.

### M3 — B군 8자리 + C군 8자리: 남은 수 제거

측정 트리: `.claude/worktrees/t530`, 브랜치 `WT-web-tab-docs`, M3 직전 HEAD `b34b7b4e6`(M2 커밋).

#### 처분 — B군 8자리 (`advanced/moai-web-console.md`, 4로케일 × 2자리)

| 자리 | 로케일 | 이전 | 이후 |
|---|---|---|---|
| B1 | ko `:25` | `… 프로젝트 섹션 편집 (14개 탭)` | `… 프로젝트 섹션 편집` |
| B3 | en `:25` | `Profile preferences and project sections (14 tabs)` | `Profile preferences and project sections` |
| B5 | ja `:25` | `… プロジェクトセクションの編集（14 タブ）` | `… プロジェクトセクションの編集` |
| B7 | zh `:25` | `… 项目章节的编辑（14 个标签页）` | `… 项目章节的编辑` |
| B2 | ko `:128` | `그 아래로 14개 탭이 세로 목록으로 펼쳐집니다.` | `그 아래로 다음 탭들이 세로 목록으로 펼쳐집니다.` |
| B4 | en `:128` | `unfolds fourteen tabs below it as a vertical list.` | `unfolds the tabs below as a vertical list.` |
| B6 | ja `:128` | `その下に 14 のタブが縦のリストとして開きます。` | `その下に次のタブが縦のリストとして開きます。` |
| B8 | zh `:128` | `其下会展开 14 个标签页的纵向列表。` | `其下会以纵向列表展开以下标签页。` |

`:128` 네 자리는 수를 지우고 **바로 아래 번호 목록을 가리키는 말**로 바꿨다 — 이름 14개가 그 자리에 이미 있으므로
수는 중복 정보였다(plan.md §B).

#### 처분 — C군 8자리 (README 4본 × 2자리)

| 자리 | 파일 | 이전 | 이후 |
|---|---|---|---|
| C1 | `README.md:414` | `splits into fourteen tabs: Identity, …` | `splits into these tabs: Identity, …` |
| C2 | `README.ko.md:414` | `… Quality Gate 열네 개 탭으로 나뉜다.` | `… Quality Gate 탭으로 나뉜다.` |
| C3 | `README.ja.md:414` | `… Quality Gate の 14 タブに分かれる。` | `… Quality Gate のタブに分かれる。` |
| C4 | `README.zh.md:414` | `设置画面分成十四个标签页：Identity、…` | `设置画面分成以下标签页：Identity、…` |
| C5 | `README.md:750` | `… · Todo), 14-tab settings` | `… · Todo), settings tabs` |
| C6 | `README.ko.md:750` | `… · Todo), 14-탭 설정` | `… · Todo), 설정 탭` |
| C7 | `README.ja.md:750` | `… · Todo)、14 タブ設定` | `… · Todo)、設定タブ` |
| C8 | `README.zh.md:750` | `… · Todo）、14 标签页设置` | `… · Todo）、设置标签页` |

`:414` 네 자리는 **이름 목록을 그대로 두고 수만** 뺐다 — 목록이 정보를 운반하고, 수는 그 목록을 세면 나온다.

#### GREEN — AC-TCD-002 축어 (열거 20자리 리터럴 전부)

```
$ wc -l < .moai/reports/t530/count-literals.txt
      16
$ grep -rnF -f .moai/reports/t530/count-literals.txt \
    README.md README.ko.md README.ja.md README.zh.md \
    docs-site/content/{ko,en,ja,zh}/cli-reference/web.md \
    docs-site/content/{ko,en,ja,zh}/advanced/moai-web-console.md | wc -l
       0
```

패턴 파일이 16줄임을 함께 쟀다 — 빈 패턴 파일의 공허한 `0` 을 차단하는 대조군이다(AC-TCD-002).

#### GREEN — AC-TCD-012 셸 축 축어 (`tab-count-sites.md § 재현 명령 5`)

```
$ grep -rniE '(one|two|…|twenty|하나|…|네|[一二三四五六七八九十])[^.。]{0,12}(\btabs?\b|탭|タブ|标签页)' \
    <대상 12파일> \
  | sed 's/第[一二三四五六七八九十]/第X/g' \
  | grep -iE '(nine|fourteen|열네|十四|九|…)[^.。]{0,12}(\btabs?\b|탭|タブ|标签页)' \
  | grep -vE '^docs-site/content/[a-z]+/advanced/moai-web-console\.md:[0-9]+:.*codex'
(출력 없음)
```

M2 이전 6행 → M3 이후 **0행**. 허용 규칙의 `grep -vE` 에 파일 경로가 포함된 형태를 그대로 썼다.

#### GREEN — AC-TCD-012 가드 축 축어 (낱말 축 적중 집합)

```
$ go test -count=1 ./internal/web/ -run 'TestDocsTabContract/allowlist' -v 2>&1 | grep '^word-axis hit '
(출력 없음)
$ go test -count=1 ./internal/web/ -run 'TestDocsTabContract/allowlist' -v 2>&1 | grep -c '^word-axis hit '
0
```

**이 0 은 M1 의 6줄과 짝을 이루어야만 의미를 가진다.** M1 이 같은 명령으로 6줄
(A2 `nine settings tabs` · A4 `九个标签页` · B4 `fourteen tabs` · C1 `fourteen tabs` · C2 `열네 개 탭` · C4 `十四个标签页`)을
찍은 기록이 위 M1 절에 축어로 남아 있고, M2 가 그중 둘(A2·A4)을, M3 가 나머지 넷(B4·C1·C2·C4)을 없앴다.
M1 기록이 없었다면 이 `0` 은 "적중이 없다" 와 "가드가 낱말 축을 구현하지 않았다" 를 구분하지 못한다 —
AC-TCD-012 가 측정 순서를 판정의 전제로 못 박은 이유가 이것이다.

#### 가드 상태 — 3층 중 2층 GREEN

```
$ go test ./internal/web/ -run 'TestDocsTabContract' -v
    docs_tab_contract_test.go:136: swept 12 files against 16 enumerated literals
    docs_tab_contract_test.go:281: allowed rules 1
    docs_tab_contract_test.go:282: allowed lines 16
--- FAIL: TestDocsTabContract (0.08s)
    --- PASS: TestDocsTabContract/literals (0.00s)
    --- PASS: TestDocsTabContract/allowlist (0.07s)
    --- FAIL: TestDocsTabContract/names (0.00s)
```

`numeral sweep found N unallowed hit(s)` 줄과 `enumerated literal(s) still present` 줄이 **사라졌다** —
두 층이 적중 0 으로 통과했다는 뜻이다. `swept 12 files` 는 그대로이며, 읽기 자체는 계속 일어나고 있음을 보인다.
`names` 층은 D군 12자리가 남아 있어 붉고, M4 가 뒤집는다.

#### Gap — M3 가 관측하지 못한 것

- **`names` 층 GREEN 미관측.** M4 소관이다.
- **AC-TCD-006 변이 6종 미실행.** 가드가 완전히 초록인 트리에서 도는 시험이라 M5 소관이다.
- **hugo 빌드(RG-TCD-002) 미실행.** M6 소관 — 이 커밋의 문구 교체가 마크다운 구조를 바꾸지 않았지만, 바꾸지 않았다는 것은 재지 않았다는 말과 다르다.
- **AC-TCD-010 패리티 미측정.** M4 까지 끝난 트리에서 한 번에 재는 것이 맞다(merge-base 기준).

### M4 — D군 12자리: 탭 이름을 렌더 라벨에 맞춘다

측정 트리: `.claude/worktrees/t530`, 브랜치 `WT-web-tab-docs`, M4 직전 HEAD `34ef57f79`(M3 커밋).
결정 근거: `plan.md §C` — 갈래 1(문서를 코드에 맞춘다), 2026-09-12. 한시 허용 목록도 후속 카드 주석도 두지 않았다.

#### 처분 — 12자리, 4로케일 한 커밋 (REQ-TCD-007)

| 자리 | 파일 | 이전 | 이후 |
|---|---|---|---|
| D1~D4 | `README{,.ko,.ja,.zh}.md:414` | `3rd Party LLM` | `GLM Settings` |
| D5 | `ko/advanced/moai-web-console.md:133` | `**서드파티 LLM(3rd Party LLM)**` | `**GLM 설정(GLM Settings)**` |
| D6 | `en/advanced/moai-web-console.md:133` | `**3rd Party LLM**` | `**GLM Settings**` |
| D7 | `ja/advanced/moai-web-console.md:133` | `**サードパーティ LLM（3rd Party LLM）**` | `**GLM設定（GLM Settings）**` |
| D8 | `zh/advanced/moai-web-console.md:133` | `**第三方 LLM（3rd Party LLM）**` | `**GLM设置（GLM Settings）**` |
| D9 | `ko/advanced/moai-web-console.md:165` | `서드파티 LLM 탭에는 …` | `GLM 설정 탭에는 …` |
| D10 | `en/advanced/moai-web-console.md:165` | `The 3rd Party LLM tab carries …` | `The GLM Settings tab carries …` |
| D11 | `ja/advanced/moai-web-console.md:165` | `サードパーティ LLM タブには …` | `GLM設定タブには …` |
| D12 | `zh/advanced/moai-web-console.md:165` | `第三方 LLM 标签页带有 …` | `GLM设置标签页带有 …` |

D9~D12 는 번호 목록 밖 산문이라 **N2 가드가 보지 않는 자리**다 — 열거표가 그 넷을 표 행으로 올려둔 덕에
사람이 같은 변경에서 함께 고칠 수 있었다. 가드가 초록이어도 이 넷은 가드가 지켜 주지 않는다는 사실을 여기 남긴다.
zh D12 의 `第三方` 가 사라지면서 서수 접두 제외 규칙이 겨누던 유일한 오탐 자리도 함께 없어졌다 — 규칙은 그대로 두었다
(규칙은 미래의 재등장을 막는 장치이지 오늘의 적중을 세는 장치가 아니다).

#### GREEN — AC-TCD-003 축어

```
$ grep -rn '3rd Party LLM\|서드파티 LLM\|サードパーティ LLM\|第三方 LLM' \
    README.md README.ko.md README.ja.md README.zh.md \
    docs-site/content/{ko,en,ja,zh}/advanced/moai-web-console.md | wc -l
       0
```

대조군(지우기만 하고 대체 라벨을 넣지 않은 상태를 배제) — 파일별 정본 라벨 실재 확인:

```
$ grep -c 'GLM Settings' README.md README.ko.md README.ja.md README.zh.md docs-site/content/en/advanced/moai-web-console.md
README.ko.md:1
README.zh.md:1
README.ja.md:1
README.md:1
docs-site/content/en/advanced/moai-web-console.md:2
$ grep -c 'GLM 설정' docs-site/content/ko/advanced/moai-web-console.md
2
$ grep -c 'GLM設定' docs-site/content/ja/advanced/moai-web-console.md
2
$ grep -c 'GLM设置' docs-site/content/zh/advanced/moai-web-console.md
2
```

README 4본 각 1행(번호 목록) + 콘솔 4본 각 2행(번호 목록 + `:165` 산문) = **합계 12행**, D군 12자리와 같다.

#### GREEN — AC-TCD-004 / 005 / 007 / 009: 가드 3층 전부 초록

```
$ go test -count=1 ./internal/web/ -run 'TestDocsTabContract' -v
    docs_tab_contract_test.go:136: swept 12 files against 16 enumerated literals
    docs_tab_contract_test.go:281: allowed rules 1
    docs_tab_contract_test.go:282: allowed lines 16
    docs_tab_contract_test.go:357: README.md: extracted 14 tab names
    docs_tab_contract_test.go:357: README.ko.md: extracted 14 tab names
    docs_tab_contract_test.go:357: README.ja.md: extracted 14 tab names
    docs_tab_contract_test.go:357: README.zh.md: extracted 14 tab names
    docs_tab_contract_test.go:357: docs-site/content/ko/advanced/moai-web-console.md: extracted 14 tab names
    docs_tab_contract_test.go:357: docs-site/content/en/advanced/moai-web-console.md: extracted 14 tab names
    docs_tab_contract_test.go:357: docs-site/content/ja/advanced/moai-web-console.md: extracted 14 tab names
    docs_tab_contract_test.go:357: docs-site/content/zh/advanced/moai-web-console.md: extracted 14 tab names
--- PASS: TestDocsTabContract (0.08s)
    --- PASS: TestDocsTabContract/literals (0.00s)
    --- PASS: TestDocsTabContract/allowlist (0.07s)
    --- PASS: TestDocsTabContract/names (0.00s)
```

세 서브테스트 이름이 리터럴로 모두 있고 `--- FAIL` 도 `no tests to run` 도 없다(AC-TCD-004).
`swept 12 files`(AC-TCD-005) · `allowed rules 1` + `allowed lines 16`(AC-TCD-007) · 8파일 각 `extracted 14`(AC-TCD-009).

#### GREEN — AC-TCD-010 4로케일 패리티 (읽는 시점 merge-base)

```
$ git merge-base develop HEAD
1d150a27d4c5cdeedb37df19b7a4a025e5dd2c09
$ git diff --name-only 1d150a27d4c5cdeedb37df19b7a4a025e5dd2c09 -- <대상 12파일>
README.ja.md
README.ko.md
README.md
README.zh.md
docs-site/content/en/advanced/moai-web-console.md
docs-site/content/en/cli-reference/web.md
docs-site/content/ja/advanced/moai-web-console.md
docs-site/content/ja/cli-reference/web.md
docs-site/content/ko/advanced/moai-web-console.md
docs-site/content/ko/cli-reference/web.md
docs-site/content/zh/advanced/moai-web-console.md
docs-site/content/zh/cli-reference/web.md
```

변경 파일 **12**(0 이 아니므로 측정이 성립), README **4**본 전부, 로케일 상대 경로
`advanced/moai-web-console.md` **4회** · `cli-reference/web.md` **4회** — 네 로케일이 빠짐없이 같은 자리에서 움직였다.
`develop` 흡수 전이라 merge-base 가 분기점(`1d150a27d`)에 머물러 있고, 그 값을 여기 리터럴로 **기록**하되
판정식에는 읽는 시점 `git merge-base` 를 쓴다(acceptance.md §D 두 축 표).

#### 유지 — RG-TCD-001 (스크린샷 미변경)

```
$ git diff --quiet 1d150a27d4c5cdeedb37df19b7a4a025e5dd2c09 -- assets/images/
$ echo $?
0
```

#### 유지 — 기존 테스트 미파손 (§D.2 항목 3)

```
$ go test -count=1 ./internal/web/
ok  	github.com/modu-ai/moai-adk/internal/web	14.957s
```

#### Gap — M4 가 관측하지 못한 것

- **AC-TCD-006 변이 6종 미실행** — M5 소관. 지금 가드는 초록이지만 **무엇을 잡는지는 아직 보인 적이 없다**;
  M1~M4 의 RED→GREEN 전이는 집합 축의 증거이고, 변이는 축 자체의 증거다. 둘은 대체재가 아니다.
- **AC-TCD-008 오탐 4자리 침묵 미확인** — M5 소관.
- **AC-TCD-011 / RG-TCD-002 / RG-TCD-003 미측정** — M6 소관(열거표 계수, hugo 빌드).
- **범위 밖 이름 드리프트 8자리 유지** — Identity(zh)·Git&Worktree(zh)·Codex(ko/ja/zh)·Cross-Session(ko/ja/zh)의
  현지어 축 드리프트는 M4 가 건드리지 않았다. 가드의 `names` 층이 그 축을 열지 않으므로(위 M1 잔여 위험 절)
  초록이 그 여덟 자리의 정합을 뜻하지 않는다. 별도 카드 소관이다.

### M5 — 변이 6종 + 오탐 4자리: 가드의 축이 살아 있음을 보인다

측정 트리: `.claude/worktrees/t530`, 브랜치 `WT-web-tab-docs`, M5 착수 시 HEAD `e275edbee`(M4 커밋), 작업 트리 clean.
아래 모든 출력은 그 트리에서 이번 실행으로 관측했다. 변이는 **한 번에 하나만** 넣었다 — 둘을 겹치면 귀속이 성립하지 않는다.

#### 왜 M5 가 형식이 아닌가 — 낱말 축의 0은 두 가지를 뜻할 수 있다

M2~M4 가 낱말 표기 6자리를 없앤 뒤, `grep -c '^word-axis hit '` 는 **0** 을 낸다. 그런데 낱말 클래스가
아무것도 매치하지 않는 스테일한 가드도 **똑같이 0** 을 낸다 — 두 판독이 바이트 동일하다.
`verification-completeness.md` §1.3 이 이름 붙인 「비실행이 성공과 구분되지 않는 검사」가 정확히 이 모양이다.
V4·V5·V6 이 그 둘을 가르는 유일한 장치이며, 그래서 M5 는 표의 세 행이 아니라 이 마일스톤의 목적이다.
같은 논리가 `literals`·`names` 층에도 걸린다 — 적중 집합이 빈 초록은 속을 도려낸 검사가 내는 초록과 같다.

#### 베이스라인 — 변이 전 트리는 초록

```
$ go test -count=1 ./internal/web/ -run 'TestDocsTabContract' -v
--- PASS: TestDocsTabContract (0.08s)
    --- PASS: TestDocsTabContract/literals (0.00s)
    --- PASS: TestDocsTabContract/allowlist (0.07s)
    --- PASS: TestDocsTabContract/names (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/web	0.584s
```

#### AC-TCD-006 — 변이 6종 (각각 따로 주입 → FAIL 관측 → 되돌림 → `ok` 재확인)

| 변이 | 대상 | 조작 | 붉어진 서브테스트 | 가드가 찍은 축어 | 되돌림 |
|---|---|---|---|---|---|
| V1 | `README.ko.md:414` | `·Audit·` → `·Audits·` | `names` | `README.ko.md: tab 7 name = "Audits", console renders "Audit"` | `ok` |
| V2 | `en/advanced/moai-web-console.md:139` | `10. **Report**` → `10. **Reports**` | `names` | `docs-site/content/en/advanced/moai-web-console.md: tab 10 name = "Reports", console renders "Report"` | `ok` |
| V3 | `zh/cli-reference/web.md:53` | `设置标签页界面。` → `设置 14 个标签页。` | `literals` + `allowlist` | `digit axis: docs-site/content/zh/cli-reference/web.md:53 … "14 个标签页"` / `numeral sweep found 1 unallowed hit(s) (0 on the word axis)` | `ok` |
| V4 | `en/advanced/moai-web-console.md:128` | `unfolds the tabs below …` → `unfolds fourteen settings tabs below it …` | `allowlist` **단독** | `word-axis hit docs-site/content/en/advanced/moai-web-console.md: fourteen settings tabs` | `ok` |
| V5 | `README.ko.md:414` | `Quality Gate 탭으로` → `Quality Gate 열네 개 탭으로` | `literals` + `allowlist` | `word-axis hit README.ko.md: 열네 개 탭` | `ok` |
| V6 | `zh/cli-reference/web.md:53` | `设置标签页界面。` → `设置九个标签页。` | `literals` + `allowlist` | `word-axis hit docs-site/content/zh/cli-reference/web.md: 九个标签页` | `ok` |

**여섯 변이 전부 FAIL 을 냈다. 잡히지 않은 변이는 없다.**

**V4 가 이 표에서 가장 무거운 행이다.** 그 변이에서 `literals` 는 **PASS** 했다 —
`fourteen settings tabs` 는 열거 리터럴 `fourteen tabs` 가 아니므로 리터럴 층에 보이지 않는다.
낱말 축 하나가 단독으로 잡았고, 축이 없었다면 손으로 적힌 수가 살아 있는 채 AC-TCD-002 가 초록이 됐다.
축어:

```
word-axis hit docs-site/content/en/advanced/moai-web-console.md: fourteen settings tabs
    docs_tab_contract_test.go:300: word axis: …:128 writes a settings tab count by hand: "fourteen settings tabs"
--- FAIL: TestDocsTabContract (0.09s)
    --- PASS: TestDocsTabContract/literals (0.00s)
    --- FAIL: TestDocsTabContract/allowlist (0.08s)
    --- PASS: TestDocsTabContract/names (0.00s)
```

**V3 와 V6 은 같은 파일·같은 줄을 겨누지만 다른 축이다.** V3 는 숫자 `14 个` 이고 가드는
`digit axis` 로 보고하며 `(0 on the word axis)` 를 함께 찍었다. V6 은 낱말 `九` 이고 `word-axis hit` 줄을 냈다.
한 변이가 두 축을 동시에 증명하지 않는다는 것이 이 두 행의 존재 이유다.

**토큰별 자기 줄 확인 (V4·V5·V6 개별 단정)** — 세 변이가 각각 **자기 토큰의** `word-axis hit` 줄을 냈다:

| 변이 | 토큰 | 로케일 | 그 변이에서 관측된 `word-axis hit` 줄 |
|---|---|---|---|
| V4 | `fourteen` | en | `word-axis hit docs-site/content/en/advanced/moai-web-console.md: fourteen settings tabs` |
| V5 | `열네` | ko | `word-axis hit README.ko.md: 열네 개 탭` |
| V6 | `九` | zh | `word-axis hit docs-site/content/zh/cli-reference/web.md: 九个标签页` |

en 토큰 하나가 셋을 대신 운반한 것이 아니다 — ko·zh 변이는 en 문서를 건드리지 않았고, 각 줄의 파일 경로가 그 사실이다.
낱말 클래스를 `{fourteen}` 하나로 좁힌 가드라면 V5·V6 에서 0줄을 찍고 통과했을 것이며, 그것이 plan-audit iter3
부채 N1 이 지목한 변이체다(M1 이 구현 측에서, 여기가 문서 측에서 각각 죽인다).

#### AC-TCD-008 — 오탐 4자리에 대해 가드가 침묵한다

네 자리가 실재함을 먼저 확인하고(구절까지 고정), 가드를 돌렸다:

```
$ grep -c 'measured nine forms' README.md                                    → 1
$ grep -c 'Eleven ref skills' README.md                                      → 1
$ grep -c 'stays selectable' README.md                                       → 1
$ grep -c '4 タブ' docs-site/content/ja/claude-code/extensibility/plugins.md  → 1
$ go test -count=1 ./internal/web/ -run 'TestDocsTabContract'
ok  	github.com/modu-ai/moai-adk/internal/web	0.584s
```

침묵의 **직접** 근거 — 가드의 `-v` 출력 전량에서 네 자리를 가리키는 문자열을 셌다:

```
$ go test -count=1 ./internal/web/ -run 'TestDocsTabContract' -v > /tmp/t530-m5-base.txt
$ grep -cE 'selectable|measured nine forms|Eleven ref skills|plugins\.md' /tmp/t530-m5-base.txt
0
```

네 자리가 각각 다른 장치로 닫힌다(`tab-count-sites.md §7-A`): `stays selectable` 은 낱말 경계 `\btabs?\b`,
`measured nine forms` 와 `Eleven ref skills` 는 수-명사 인접(탭 명사 부재), `4 タブ` 는 파일 목록(대상 12파일 밖).
`ok` 하나만으로는 "가드가 이 자리를 보고 통과시켰다" 와 "가드가 이 자리를 애초에 읽지 않았다" 가 구분되지 않으므로,
적중 0을 출력 검색으로 함께 쟀다.

#### 되돌림 — 작업 트리는 깨끗하다

여섯 변이를 모두 되돌린 뒤:

```
$ git status --short
(출력 없음)
$ go test -count=1 ./internal/web/ -run 'TestDocsTabContract' -v
--- PASS: TestDocsTabContract (0.11s)
    --- PASS: TestDocsTabContract/literals (0.00s)
    --- PASS: TestDocsTabContract/allowlist (0.10s)
    --- PASS: TestDocsTabContract/names (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/web	0.704s
```

`no tests to run` 토큰 없음, `--- FAIL` 없음 — AC-TCD-004 의 세 줄 단정이 이 트리에서도 성립한다.
변이는 하나도 남지 않았다(위 `git status --short` 가 그 관측이며, 이 M5 커밋이 담는 것은 `progress.md` 뿐이다).

#### 무회귀 · 빌드 · 정적 검사

| 명령 | 관측 |
|---|---|
| `go test -count=1 ./internal/web/` | `ok  github.com/modu-ai/moai-adk/internal/web  18.561s` |
| `go build ./internal/web/` | exit `0` |
| `GOOS=windows GOARCH=amd64 go build ./internal/web/` | exit `0` |
| `gofmt -l internal/web/` | 출력 없음 |
| `go vet ./internal/web/` | 출력 없음, exit `0` |

#### Gap — M5 가 관측하지 못한 것

- **가드가 잡지 못하는 변이가 없다고 단정하지 않는다.** 이 여섯은 acceptance.md 가 지정한 여섯이고,
  **여섯 자리 × 세 층**에 대한 증거다. 열거되지 않은 표기(새 로케일, 새 수사 낱말)를 이 마일스톤은 재지 않았다 —
  낱말 클래스의 유지 비용은 `spec.md §7` 이 이미 소유한다.
- **D9~D12 산문 4자리는 여전히 가드 밖이다**(M4 절에 기록). V1·V2 는 번호 목록·이름 배열 축의 변이이며
  산문 축을 겨누지 않았다.
- **`names` 층의 현지어 축은 열지 않았다**(M1 잔여 위험 절). 범위 밖 이름 드리프트 8자리는 별도 카드 소관이다.
- **AC-TCD-011 / RG-TCD-002 / RG-TCD-003 미측정** — M6 소관(열거표 계수, hugo 빌드).
- **머신 부하.** M5 실행 구간의 `uptime` 1분 평균은 4.85 → 17.48 로 상승했다. 측정은 `./internal/web/` 로
  한정했고 `go test ./...` 는 돌리지 않았으나, 이 수치들의 벽시계 시간은 부하의 함수이지 코드의 함수가 아니다
  — 판정은 PASS/FAIL 이고 소요 시간을 근거로 쓰지 않는다.

### M6 — 빌드·패리티 검증 (Low)

M1~M5 가 남긴 세 자리(AC-TCD-011 / RG-TCD-002 / RG-TCD-003)를 재고, plan §M6 이 요구한 패리티 두 축을 잰다.
측정 트리는 `dd66526ff`(M5 커밋), 워킹 트리는 측정 시작·종료 모두 `git status --short` 무출력이다.

#### AC-TCD-011 — 열거 산출물이 최신이다 (경로 축)

acceptance.md §AC-TCD-011 의 When 블록을 축어로 실행:

```bash
grep '^| [ABCD][0-9]' .moai/reports/t530/tab-count-sites.md > /tmp/t530-rows.txt
wc -l < /tmp/t530-rows.txt
grep -oE '`[^`]+\.md`' /tmp/t530-rows.txt | tr -d '`' | sort -u > /tmp/t530-paths.txt
wc -l < /tmp/t530-paths.txt
while read -r f; do test -f "$f" || echo "MISSING $f"; done < /tmp/t530-paths.txt
```

```
rows=      32
paths=      12
--- missing ---
--- end (empty=all exist) ---
```

행 32 · 경로 12 · `MISSING` 0건 — Then 의 세 조건이 모두 성립한다. **PASS.**
`MISSING` 줄이 하나도 없다는 것은 12경로가 전부 이 트리에 실재한다는 관측이며, 빈 출력을 통과로 읽지 않도록
종료 표식(`--- end (empty=all exist) ---`)을 함께 찍어 명령이 실제로 끝까지 돌았음을 남긴다.

#### RG-TCD-003 — 열거 산출물이 줄어들지 않는다

```bash
grep -c '^| [ABCD][0-9]' .moai/reports/t530/tab-count-sites.md
```

```
32
```

기준은 `32 이상`이고 실측 32 — **유지.** AC-TCD-011 의 행 수와 같은 영역·같은 셀렉터에서 나온 값이므로
두 수치는 독립 관측이 아니라 같은 관측의 두 인용이다. 그렇게 적는다.

#### RG-TCD-002 — docs-site 빌드 경고 0

빌드 산출물이 트리를 더럽히지 않는지부터 확인했다 — `git check-ignore -v docs-site/public` 가
`docs-site/.gitignore:2:public/` 를 반환하고, 루트 `.gitignore:341-343` 이 `public/` · `resources/_gen/` ·
`.hugo_build.lock` 를 덮는다. 그 뒤 실행:

```bash
cd docs-site && hugo --gc --minify > /tmp/t530-hugo.log 2>&1
```

```
exit=0
warn/error count=0
hugo v0.160.1+extended+withdeploy darwin/arm64

              │ KO  │ EN  │ JA  │ ZH
──────────────┼─────┼─────┼─────┼─────
 Pages        │ 187 │ 185 │ 185 │ 185
 Non-page     │  12 │  12 │  12 │  12
 files        │     │     │     │
 Static files │ 265 │ 265 │ 265 │ 265
 Aliases      │   6 │   5 │   5 │   5

Total in 2706 ms
```

`grep -ciE 'warn|error'` 가 `0`, exit `0` — **유지.** 빌드 직후 `git status --short` 도 무출력이라
산출물 유출도 없다. 이 카드가 경고를 **만들지 않았음**을 본 것이지, 빌드가 원래 초록이었다는 사실을
이 카드의 공으로 적지 않는다.

#### 로케일 존재·섹션 수 패리티

카드가 만진 docs-site 8본은 두 문서 × 4로케일이다. 로케일마다 파일 존재와 heading 계수를 함께 잰다:

| 문서 | ko | en | ja | zh |
|---|---|---|---|---|
| `cli-reference/web.md` | 7 / h2 7 / h3 0 | 7 / h2 7 / h3 0 | 7 / h2 7 / h3 0 | 7 / h2 7 / h3 0 |
| `advanced/moai-web-console.md` | 18 / h2 13 / h3 4 | 18 / h2 13 / h3 4 | 18 / h2 13 / h3 4 | 18 / h2 13 / h3 4 |

8본 전부 `exists=yes`, 문서별로 4로케일의 전체 heading 수·h2 수·h3 수가 모두 일치한다. **PASS.**

#### README 4본 heading 패리티

| 파일 | 전체 heading | h2 | h3 |
|---|---|---|---|
| `README.md` | 80 | 12 | 63 |
| `README.ko.md` | 80 | 12 | 63 |
| `README.ja.md` | 80 | 12 | 63 |
| `README.zh.md` | 80 | 12 | 63 |

네 본이 세 계수 모두 동일하다. **PASS.**

#### 무회귀

| 명령 | 관측 |
|---|---|
| `go test -count=1 ./internal/web/ -run 'TestDocsTabContract'` | `ok  github.com/modu-ai/moai-adk/internal/web  0.688s` |
| `go test -count=1 ./internal/web/` | `ok  github.com/modu-ai/moai-adk/internal/web  15.114s` |

#### Gap — M6 이 관측하지 못한 것

- **heading 계수는 구조 패리티이지 내용 패리티가 아니다.** 네 로케일의 heading **수**가 같음을 쟀을 뿐,
  같은 자리에 같은 뜻의 제목이 있는지는 재지 않았다. 계수가 같으면서 순서가 어긋난 상태는 이 측정이 통과시킨다.
- **hugo 무경고는 이 로컬 hugo 버전(0.160.1)의 관측이다.** Vercel 빌드가 쓰는 버전과 같다는 것은 재지 않았다.
- **docs-site Vercel 바인딩 미검증** — 이 카드의 docs-site 변경이 develop 착지 시 프리뷰/프로덕션 배포에
  어떻게 반응하는지는 재지 않았다(CLAUDE.local.md §4.1 의 미검증 항목). 리드 판단 사항으로 남긴다.
- **M5 의 Gap 3건은 그대로 남는다** — D9~D12 산문 4자리 가드 밖, `names` 층 현지어 축 미개방,
  열거되지 않은 새 수사 표기. M6 은 그 어느 것도 닫지 않았다.
- **후속 카드 2건 미발행** — plan §M6 의 둘째 항목(스크린샷 절차·재촬영)은 카드 요청으로 리드에게 올린다.
  이 카드가 만들지 않는다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-13
run_commit_sha: dd66526ff              # M5 커밋. M6 커밋은 자기 해시를 인용할 수 없어 sync 커밋이 backfill 한다
run_status: complete                    # M1~M6 종료
ac_pass_count: 12                       # AC-TCD-001~012 전부
ac_fail_count: 0
ac_pending_count: 0                     # AC-TCD-011 은 M6 에서 측정 완료 (행 32 / 경로 12 / MISSING 0)
rg_maintained: 3                        # RG-TCD-001 스크린샷 미변경 / 002 hugo 경고 0 / 003 열거 32행
preserve_list_post_run_count: 0         # PRESERVE 목록 침범 0 — 기존 Go 소스(schemaform.go / tab_layout_test.go),
                                        # 스크린샷, 범위 밖 로케일 이름 8자리 전부 미변경
l44_pre_commit_fetch: "git fetch origin develop → rc=0 (M5 커밋 직전)"
l44_post_push_fetch: "N/A — 레인은 push 하지 않는다 (CLAUDE.local.md §4.1, 리드 일괄 push)"
new_warnings_or_lints_introduced: 0     # go vet 무출력, gofmt 무출력
cross_platform_build:
  darwin: "go build ./internal/web/ → exit 0"
  windows: "GOOS=windows GOARCH=amd64 go build ./internal/web/ → exit 0"
total_run_phase_files: 14               # 신규 1 (internal/web/docs_tab_contract_test.go)
                                        # + 문서 12 (README 4본 + docs-site 8본)
                                        # + progress.md
m1_to_mN_commit_strategy: "마일스톤당 1커밋, 5커밋 (M1 가드 / M2 A군 / M3 B·C군 / M4 D군 / M5 변이 증거).
  Conventional Commits, 모든 커밋 메시지에 카드 id t530. WT 브랜치 push 없음 — 로컬 병합 SHA 를 리드에게 보고한다."
```

**§E.3 의 근거는 이 문서 §E.2 의 M1~M5 절이다.** 각 AC 의 RED-now 출력과 GREEN 출력이 축어로 그 안에 있으며
(`§D.2` 항목 2), 이 YAML 은 그 관측의 요약이지 별개의 측정이 아니다.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
