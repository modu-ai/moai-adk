---
id: SPEC-DOCS-TABCOUNT-DRIFT-001
title: 인수 기준 — 설정 탭 수·이름 드리프트 차단
version: "0.3.1"
created: 2026-09-12
updated: 2026-09-13
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: "README.{md,ko,ja,zh}, docs-site/content/**, internal/web"
lifecycle: spec-anchored
tags: "docs, guard, acceptance, t530"
---

# 인수 기준

모든 항목은 **명령 + 기대 출력**이다. 산문 단정은 인수 근거가 아니다.
모든 명령은 워크트리 루트(`.claude/worktrees/t530`)에서 실행한다.

**문서 수준 pin: 모든 RED-now 관측은 `git merge-base develop HEAD` = `1d150a27d` 에서 쟀다.**
개별 AC 가 따로 pin 하지 않으면 이 값이 적용된다(`verification-completeness.md §2.1`).

**두 축을 구분한다 — "고정 SHA" 와 "움직이는 ref" 의 이분법으로는 이 카드를 잴 수 없다.**

| 무엇을 단정하나 | 왼쪽 끝 | 이유 |
|---|---|---|
| **관측 기록**(RED-now 셀의 수치) | 잰 시점의 SHA 를 적는다 | 기록은 움직이면 안 된다 |
| **이 카드가 무엇을 바꿨는가**(AC-TCD-010, RG-TCD-001) | 읽는 시점 `git merge-base develop HEAD` | 리터럴 핀은 develop 흡수 후 남의 커밋을 분모에 넣는다 |

`origin/develop...` 같은 **브랜치 이름 그대로**를 쓴 판정식은 하나도 없다(움직이는 ref 는 남의 push 로 전진한다).
merge-base 는 브랜치 이름이 아니라 **흡수 지점**이며, 이 저장소의 [HARD] 절차(`gitflow-lane-protocol.md` §8)가
범위 판정식에 요구하는 바로 그 왼쪽 끝이다 — 흡수 전에는 분기점, 흡수 후에는 흡수한 develop 커밋에 머문다.

## §D.0 두 부류 — 인수 기준과 회귀 가드를 섞지 않는다

| 부류 | 뜻 | 판정 |
|---|---|---|
| **AC (인수 기준)** | 이 카드의 작업이 **붉은 것을 초록으로 바꾼다** | 각 항목에 RED-now 관측(명령 + 현재 출력)과 green path(어느 M 이 뒤집나)가 함께 있다 |
| **RG (회귀 가드)** | 도착 시점에 **이미 초록**이며, 이 카드가 **깨뜨리지 않았음**을 본다 | RED-now 가 없다. 통과는 "작업했다" 가 아니라 "망가뜨리지 않았다" 는 뜻이다 |

RG 를 AC 로 세면 합격이 부풀려진다 — 손대지 않은 트리에서 이미 통과하기 때문이다. 그래서 분리한다.

## §D AC 매트릭스

| AC | 검증 대상 REQ | 무엇을 쓸어담나 | RED-now (base `1d150a27d`) | green path |
|---|---|---|---|---|
| AC-TCD-001 | REQ-TCD-002 | A군 4자리 리터럴 | 4행 적중 | M2 |
| AC-TCD-002 | REQ-TCD-003, 008 | 열거 20자리 리터럴 전부 | 20행 적중 | M2+M3 |
| AC-TCD-003 | REQ-TCD-006 | D군 12자리 이름 | 12행 적중 | M4 |
| AC-TCD-004 | REQ-TCD-004 | 가드 3 서브테스트 존재·통과 | `[no tests to run]` | M1+M5 |
| AC-TCD-005 | REQ-TCD-004, 005 | 가드가 **읽어 낸** 파일 수 == 12 | 가드 부재 | M1+M5 |
| AC-TCD-006 | REQ-TCD-004, 005 | 변이 6종(파일 부류 × 로케일 × 수사 표기 en/ko/zh) | 가드 부재 | M5 |
| AC-TCD-007 | REQ-TCD-005 | 허용 **규칙** 1 + 면제 **줄** 16, 파일 범위 필수 | 가드 부재 | M1+M5 |
| AC-TCD-008 | REQ-TCD-005 | 오탐 4자리 침묵 | 가드 부재 | M5 |
| AC-TCD-009 | REQ-TCD-006 | 이름 배열 == 렌더 라벨(순서 포함) | 가드 부재 | M4+M5 |
| AC-TCD-010 | REQ-TCD-007 | 4로케일 패리티(README 포함, 12파일, merge-base) | 변경 0 ⇒ 측정 불가 = FAIL | M2~M4 |
| AC-TCD-011 | REQ-TCD-001 | 열거표 행 32 + 경로 12, 같은 영역에서 | 32 / 12 (M6 에서 재확인) | M6 |
| AC-TCD-012 | REQ-TCD-003, 005, 008 | 낱말 표기 6자리 — 셸 축 0행 **+ 가드 축 적중 집합 == 6자리** | 셸 6행 적중 / 가드 부재(미측정) | M1(가드 축) + M2+M3(셸 축) |

| RG | 대응 REQ | 무엇을 지키나 | 도착 시점 실측 |
|---|---|---|---|
| RG-TCD-001 | REQ-TCD-009 | 스크린샷 미변경 | 변경 파일 없음 |
| RG-TCD-002 | **없음(의도적)** — 요구사항이 아니라 빌드 위생 | docs-site 빌드 경고 0 | 무경고 |
| RG-TCD-003 | REQ-TCD-001 | 열거 산출물이 줄어들지 않음 | 32행 |

REQ 커버리지: REQ-TCD-001~009 전부 최소 1개의 AC 또는 RG 에 대응한다(001→AC-011·RG-003, 002→AC-001,
003→AC-002·012, 004→AC-004·005·006, 005→AC-007·008·012, 006→AC-003·009, 007→AC-010, 008→AC-002·012, 009→RG-001).
RG-TCD-002 만 대응 REQ 가 없고, 그 사실을 `—` 로 흐리지 않고 **의도적 부재**로 적었다 — 고아 항목이 아니라
요구사항 밖의 위생 검사다.

---

## AC — 인수 기준 (붉은 것을 초록으로)

### AC-TCD-001 — A군: 틀린 수 4자리, 영어 포함

**RED-now 셀**

- 명령: `grep -rncF -f /dev/stdin docs-site/content/ko/cli-reference/web.md` — 실제 실행 형태는 아래 Given/When 의 리터럴 파일 방식을 쓴다.
- 실측(base `1d150a27d`): A군 리터럴 4개(`count-literals.txt` 의 첫 4줄)로 4파일을 훑어 **4행** 적중 — ko·**en**·ja·zh 각 1행.
- 붉은 이유: 네 자리 모두 아직 `9` 를 적고 있다. **en 이 포함된다는 점이 이 AC 의 존재 이유다** — 창 기반 정규식은
  `The nine settings tabs`(수사-명사 간격 10글자)를 못 보고 통과한다.

**Given** M2 가 끝난 트리에서,
**When**
```bash
head -4 .moai/reports/t530/count-literals.txt > /tmp/t530-a.txt
grep -rnF -f /tmp/t530-a.txt docs-site/content/{ko,en,ja,zh}/cli-reference/web.md | wc -l
```
를 실행하면,
**Then** 출력이 `0` 이다.
**And** 같은 명령이 M2 이전에는 `4` 를 출력해야 한다 — 그 출력을 `progress.md §E.2` 에 축어로 남긴다.
`0` 만 보고 통과시키지 않는다: 리터럴 파일이 비어 있어도 `0` 이 나오므로, 파일이 4줄임을 `wc -l` 로 함께 센다.

### AC-TCD-002 — 열거 20자리 전부

**RED-now 셀**

- 명령: 아래 When 과 동일.
- 실측(base `1d150a27d`): **20행** 적중. 그 20행의 집합은 `tab-count-sites.md` A/B/C 표 20행과 **동일**하다(경로·행 대조 완료).
- 붉은 이유: 20자리가 아직 전부 살아 있다.
- 왜 스윕이 아니라 리터럴인가: 창 기반 정규식 스윕은 이 트리에서 20행을 내지만 **집합이 다르다** — en A군 자리를 놓치고,
  대신 열거 밖의 `ja/advanced/moai-web-console.md:149`(codex 근거 문장, 범위 밖)를 잡는다. 그 스윕으로 `0` 을 요구하면
  **범위 밖 문장을 고쳐야만 통과하는** 도달 불가능한 기준이 된다. 리터럴 집합은 실측한 문자열만 보므로 그 함정이 없다.

**Given** M3 가 끝난 트리에서,
**When**
```bash
grep -rnF -f .moai/reports/t530/count-literals.txt \
  README.md README.ko.md README.ja.md README.zh.md \
  docs-site/content/{ko,en,ja,zh}/cli-reference/web.md \
  docs-site/content/{ko,en,ja,zh}/advanced/moai-web-console.md | wc -l
```
를 실행하면,
**Then** 출력이 `0` 이다.
**And** 리터럴 파일이 16줄임을 `wc -l .moai/reports/t530/count-literals.txt` 로 함께 확인한다(빈 패턴 파일의 공허한 `0` 차단).

### AC-TCD-003 — D군: 탭 이름 12자리

**RED-now 셀**

- 명령: 아래 When 과 동일.
- 실측(base `1d150a27d`): **12행** 적중(목록 8 + 산문 4).
- 붉은 이유: 문서 12자리가 `3rd Party LLM` 계열로 적혀 있는데 콘솔은 `GLM Settings` 계열을 렌더한다.

**Given** M4 가 끝난 트리에서,
**When**
```bash
grep -rn '3rd Party LLM\|서드파티 LLM\|サードパーティ LLM\|第三方 LLM' \
  README.md README.ko.md README.ja.md README.zh.md \
  docs-site/content/{ko,en,ja,zh}/advanced/moai-web-console.md | wc -l
```
를 실행하면,
**Then** 출력이 `0` 이다.
**And** 로케일별 정본 라벨이 실재함을 확인한다 — `GLM Settings`(en) / `GLM 설정`(ko) / `GLM設定`(ja) / `GLM设置`(zh).
자리 수가 파일마다 다르므로 파일별로 센다: 각 로케일의 `advanced/moai-web-console.md` 에서 **2행 이상**(번호 목록 + `:165` 산문),
각 로케일의 README 에서 **1행 이상**(번호 목록). 합계 **12행 이상**.
지우기만 하고 대체 라벨을 넣지 않은 상태를 통과시키지 않기 위한 대조군이다.
현재 실측(base `1d150a27d`): `grep -c 'GLM Settings'` → `README.md:0`, `en/advanced/moai-web-console.md:0` — 아직 하나도 없다.

### AC-TCD-004 — 가드 3 서브테스트가 실재하고 통과한다

**RED-now 셀**

- 명령: `go test ./internal/web/ -run 'TestDocsTabContract' -v`
- 실측 출력(base `1d150a27d`, 축어):
  ```
  testing: warning: no tests to run
  PASS
  ok  	github.com/modu-ai/moai-adk/internal/web	0.689s [no tests to run]
  ```
  종료코드 `0`.
- 붉은 이유: 가드 파일이 아직 없다. **이 출력이 이 AC 의 존재 이유다** — 선택자가 아무것도 고르지 않아도 `ok` 와
  종료코드 0 이 나온다. 따라서 "`ok` 이면 통과" 라고 쓴 기준은 가드가 없어도 통과한다.

**Given** M5 가 끝난 트리에서,
**When** `go test ./internal/web/ -run 'TestDocsTabContract' -v 2>&1 | grep -E '^(--- PASS|--- FAIL)|no tests to run'` 를 실행하면,
**Then** 출력에 아래 **세 줄이 리터럴로 모두** 있다.
```
--- PASS: TestDocsTabContract/literals
--- PASS: TestDocsTabContract/allowlist
--- PASS: TestDocsTabContract/names
```
**And** `--- FAIL` 이 없고, `no tests to run` 이 **없다**.
부모 테스트의 `--- PASS: TestDocsTabContract` 한 줄만으로는 통과가 아니다 — 서브테스트를 만들지 않아도 부모는 통과한다.

### AC-TCD-005 — 가드가 읽은 파일 수를 스스로 센다

**RED-now 셀**: 가드 부재(AC-TCD-004 의 출력과 동일한 근거). 붉은 이유: 읽은 집합의 크기를 세는 주체가 없다.

**Given** M5 가 끝난 트리에서,
**When** `go test ./internal/web/ -run 'TestDocsTabContract/literals' -v 2>&1 | grep -E 'swept [0-9]+ files'` 를 실행하면,
**Then** 출력이 `swept 12 files` 를 포함한다.

[HARD] **카운터는 읽기에 성공한 뒤에 증가한다.** 선언된 파일 목록의 길이를 찍는 구현
(`t.Logf("swept %d files", len(targetFiles))`)은 한 파일도 열지 않고 이 AC 를 만족시킨다 — 이 AC 가 막으려는
바로 그 혼동이다. 요구되는 형태는 파일별 읽기가 성공한 **직후** `n++` 이고, 읽기 실패는 `t.Fatal` 이다
(`t.Skip` 금지 — §D.1 과 같은 규율). 따라서 `swept 12 files` 는 "12개를 읽었다" 는 관측이지 "12개를 읽을 예정이다"
라는 선언이 아니다.
**왜 필요한가**: `README.md` 한 파일만 읽는 가드도 AC-TCD-004·006·008·009 를 전부 통과할 수 있다. 읽은 집합의 크기를
판정 전에 세지 않으면 **아무것도 읽지 않은 초록과 전부 읽은 초록이 구분되지 않는다.**

### AC-TCD-006 — 변이 6종: 파일 부류 × 로케일 × 수사 표기

**RED-now 셀**: 가드 부재. 붉은 이유: 가드가 무엇을 잡는지 보인 적이 없다.

**Given** 가드가 GREEN 인 트리에서,
**When** 아래 여섯 변이를 **각각 따로** 넣고 가드를 돌린 뒤 되돌리면,

| 변이 | 대상 | 조작 | 무엇을 증명하나 |
|---|---|---|---|
| V1 | `README.ko.md` | 탭 이름 하나를 `Audit` → `Audits` | README 부류 + ko 로케일을 실제로 읽는다 |
| V2 | `docs-site/content/en/advanced/moai-web-console.md` | 탭 이름 하나를 `Report` → `Reports` | docs-site 부류 + en 로케일을 실제로 읽는다 |
| V3 | `docs-site/content/zh/cli-reference/web.md` | `14 个标签页` 를 되살림 | N1 **숫자** 축이 살아 있다 |
| V4 | `docs-site/content/en/advanced/moai-web-console.md` | `fourteen tabs` → `fourteen settings tabs` | N1 낱말 축의 **en 토큰(`fourteen`)** 이 살아 있다 |
| V5 | `README.ko.md` | C2 의 `열네 개 탭` 을 되살림 | N1 낱말 축의 **ko 토큰(`열네`)** 이 살아 있다 |
| V6 | `docs-site/content/zh/cli-reference/web.md` | A4 의 `设置九个标签页` 를 되살림 | N1 낱말 축의 **zh 토큰(`九`)** 이 살아 있다 — V3 의 숫자 축과 다른 축이다 |

**Then** 여섯 변이가 각각 `--- FAIL` 을 내고, 실패 메시지가 **어느 파일의 무엇**이 어긋났는지 적시하며, 되돌린 뒤에는 모두 `ok` 다.
**And** 열두 출력(FAIL 6 + PASS 6)을 `progress.md §E.2` 에 축어로 남긴다.

V4~V6 이 셋인 이유는 로케일마다 수사 토큰이 다르기 때문이다. 낱말 축을 `{fourteen}` 하나로만 구현한 가드는
V4 는 잡지만 V5·V6 을 놓친다 — 그 변이체를 집합 단정으로 죽이는 것이 AC-TCD-012 의 가드 축 절이고,
이 표는 같은 구속을 변이 쪽에서 한 번 더 건다. 두 자리에 거는 이유는 `REQ-TCD-005` 의
"숫자와 낱말 모두" 조항이 검증 층에서 한 곳에만 걸려 있으면, 그 한 곳의 표현이 흔들릴 때 조항이 통째로 빈다는 것이다.

V4~V6 이 이 AC 의 핵심이고, 그중 V4 는 감사가 실제로 성공시킨 변이체다. 측정은 이렇다 —
원문 `…unfolds fourteen tabs below it…` 은 리터럴 층 1적중 / 숫자 스윕 0적중이고,
한 낱말을 끼운 `…unfolds fourteen **settings** tabs…` 은 **리터럴 0 / 숫자 스윕 0** 이다.
낱말 축이 없으면 손으로 적힌 수가 살아 있는 채로 AC-TCD-002 가 초록이 된다. 한국어의
`열네 개 탭` → `탭 열네 개` 도 같은 결과를 내며, 이쪽은 악의가 아니라 평범한 재작성이다.

### AC-TCD-007 — N1 허용 목록이 정확히 일치한다

**RED-now 셀**: 가드 부재. 붉은 이유: 허용 목록을 들고 판정하는 주체가 없다.

**Given** M5 가 끝난 트리에서,
**When** `go test ./internal/web/ -run 'TestDocsTabContract/allowlist' -v 2>&1 | grep -E '^--- (PASS|FAIL)|allowed [0-9]+'` 를 실행하면,
**Then** `--- PASS: TestDocsTabContract/allowlist` 가 있고, **두 수치를 모두** 포함한다:

| 출력 토큰 | 단위 | 기대값 | 왜 두 개인가 |
|---|---|---|---|
| `allowed rules 1` | 허용 **규칙** 수 | `1` | 규칙이 조용히 늘면 가드가 무력해진다 |
| `allowed lines 16` | 그 규칙이 실제로 면제한 **줄** 수 | base 실측 `16` | 규칙 수는 그대로인데 면제 표면만 커지는 변화를 잡는다 |

단위를 토큰에 박는 이유: 이전 판은 `allowed 1` 만 요구했는데, 구현자가 "규칙 1개" 로도 "면제된 줄 1개" 로도 읽을 수 있었다.
오늘은 숫자 스윕 적중이 `ja:149` 하나뿐이라 **두 해석이 모두 1** 이어서 모호함이 보이지 않지만, 나중에 `ko:149` 가
숫자 표기로 바뀌면 한쪽은 1, 다른 쪽은 2가 되어 근거 없는 FAIL 이 난다.

실측 근거(base `1d150a27d`): `grep -ci codex docs-site/content/{ko,en,ja,zh}/advanced/moai-web-console.md` → 각 `4`, 합계 **16**.
규칙은 하나지만 그 하나가 16줄을 숫자·낱말 스윕에서 면제한다 — `spec.md §7` 의 잔여 위험이 가리키는 표면이 이 수치다.

허용 규칙은 `advanced/moai-web-console.md` **안에서** `codex` 토큰을 포함하는 줄이며(줄 번호가 아니라 내용으로 식별 —
편집이 줄 번호를 움직인다), 이유는 `tab-count-sites.md §7-B` 에 적혀 있다.
**And** 규칙 수가 0이거나 2 이상이면 FAIL 이다.
**And** 규칙에 **파일 범위가 없으면 FAIL** 이다 — 범위 없는 `codex` 규칙은 README 4본의 이름 목록 줄까지 면제해
열거된 C1~C4 를 가드에서 지운다(실측: 낱말 축 적중 6행 → 3행).

### AC-TCD-008 — 오탐 4자리에 대해 가드가 침묵한다

**RED-now 셀**: 가드 부재.

**Given** 가드가 GREEN 인 트리에서,
**When** 아래 네 자리가 실재함을 확인하고 가드를 다시 돌리면,
```bash
grep -c 'measured nine forms' README.md                                    # 1 이어야 한다
grep -c 'Eleven ref skills' README.md                                      # 1
grep -c 'stays selectable' README.md                                       # 1  (selec*tab*le — 낱말 경계 검사)
grep -c '4 タブ' docs-site/content/ja/claude-code/extensibility/plugins.md  # 1
go test ./internal/web/ -run 'TestDocsTabContract'                         # ok
```
**Then** 네 grep 이 모두 `1` 을 내고 가드는 `ok` 다.
구절까지 고정한다 — `grep -c 'nine' README.md >= 1` 식으로 느슨하게 쓰면 지켜야 할 자리가 사라지고 엉뚱한 곳에
`nine` 이 생겨도 통과한다.

### AC-TCD-009 — 이름 배열이 렌더 라벨과 순서까지 같다

**RED-now 셀**: 가드 부재 + D군 12자리가 어긋나 있음(AC-TCD-003 의 12행).

**Given** M4·M5 가 끝난 트리에서,
**When** `go test ./internal/web/ -run 'TestDocsTabContract/names' -v 2>&1 | grep -E '^--- (PASS|FAIL)|extracted [0-9]+'` 를 실행하면,
**Then** `--- PASS: TestDocsTabContract/names` 가 있고, 추출 개수가 `extracted 14` 로 보고된다.
**And** 추출 개수가 `len(consoleTabs())` 와 다르면 그 자체로 FAIL 이다 — 0개를 뽑고 "비교할 것이 없다" 며 통과하는 경로를 막는다.
비교는 집합이 아니라 **순서까지** 같은지를 본다.

### AC-TCD-010 — 4로케일 패리티 (README 포함, 읽는 시점 merge-base, 대상 12파일만)

**RED-now 셀**

- 명령: 아래 When 의 첫 두 줄(`CARD_BASE` 산출 + 12파일 pathspec diff).
- 실측(수리 시점): `git merge-base develop HEAD` → `1d150a27d4c5cdeedb37df19b7a4a025e5dd2c09`,
  그 base 에서 대상 12파일 변경 **0개**.
- 붉은 이유: 대조군이 0 이므로 **"패리티 유지" 가 아니라 "측정 불가"** 이고, 이 AC 는 그 상태를 FAIL 로 판정한다.
  이 0 은 **이 카드가 아직 문서를 고치지 않았기 때문**이며, 남의 변경 때문이 아니다 — 아래 두 정정이 그것을 보장한다.

**두 가지 정정 (이전 판이 red-forever 였던 이유)**

1. **왼쪽 끝을 리터럴 SHA 로 고정하지 않는다.** 이 저장소의 [HARD] 절차는 병합 전에 `develop` 을 흡수하고
   **병합 트리에서 재측정**하도록 정한다(`gitflow-lane-protocol.md` §8, `CLAUDE.local.md` §4.1). 흡수하는 순간
   리터럴 핀 범위에 남의 커밋이 들어오므로, **절차를 지키는 것이 이 AC 를 붉게 만든다** — 그 붉음은 이 카드가
   고칠 수 없고, green path 가 "누군가 무관한 파일을 고친다" 를 지나가므로 `verification-completeness.md` §2 가
   실격시키는 모양이다. 실측(수리 시점): `1d150a27d..develop` 40커밋, 넓은 pathspec 으로 155파일.
   따라서 왼쪽 끝은 **읽는 시점에** `git merge-base develop HEAD` 로 다시 구한다. 흡수 후에는 merge-base 가
   흡수한 develop 커밋이 되어 범위에 이 카드의 작업만 남는다.
2. **pathspec 을 대상 12파일로 좁힌다.** `docs-site/content` 통째는 이 카드가 손대지 않는 문서까지 분모에 넣는다.
   실측(수리 시점): 좁힌 pathspec 으로도 `CARD_BASE..develop` 에 8파일이 있다 — 흡수 전에는 보이고, 흡수 후에는
   merge-base 가 전진해 사라지는 값이다. 이 대비가 (1)의 정정이 필요한 이유를 그대로 보여 준다.

**Given** M2~M4 가 끝난 트리에서,
**When**
```bash
git merge-base develop HEAD   # 읽는 시점에 재구해 값을 <CARD_BASE>로 기록한다
git diff --name-only <CARD_BASE> -- README.md README.ko.md README.ja.md README.zh.md \
  docs-site/content/ko/cli-reference/web.md docs-site/content/en/cli-reference/web.md \
  docs-site/content/ja/cli-reference/web.md docs-site/content/zh/cli-reference/web.md \
  docs-site/content/ko/advanced/moai-web-console.md docs-site/content/en/advanced/moai-web-console.md \
  docs-site/content/ja/advanced/moai-web-console.md docs-site/content/zh/advanced/moai-web-console.md \
  > /tmp/t530-changed.txt
wc -l < /tmp/t530-changed.txt
grep -c '^README' /tmp/t530-changed.txt
sed -E 's#^docs-site/content/[a-z]+/##' /tmp/t530-changed.txt | grep -v '^README' | sort | uniq -c
```
를 실행하면,
**Then** 첫 수가 **1 이상**이고(0 이면 측정 불가 = FAIL), 둘째 수가 `4` 이며(README 4본 전부),
셋째 출력의 모든 상대 경로가 정확히 `4` 회씩 나온다.
**병합 뒤에는 이 판정식을 쓰지 않는다** — 병합 후 merge-base 는 카드 tip 자신이 되어 범위가 비고 공허하게 통과한다
(`gitflow-lane-protocol.md` §8 의 한계 조항). 병합 후 근거는 병합 트리와 카드 브랜치 트리의 동일성으로 대신한다.

### AC-TCD-011 — 열거 산출물이 최신이다 (경로 축)

**Given** M6 에서,
**When**
```bash
grep '^| [ABCD][0-9]' .moai/reports/t530/tab-count-sites.md > /tmp/t530-rows.txt
wc -l < /tmp/t530-rows.txt
grep -oE '`[^`]+\.md`' /tmp/t530-rows.txt | tr -d '`' | sort -u > /tmp/t530-paths.txt
wc -l < /tmp/t530-paths.txt
while read -r f; do test -f "$f" || echo "MISSING $f"; done < /tmp/t530-paths.txt
```
를 실행하면,
**Then** 행 수가 `32` 이고, 경로 수가 `12` 이며, 셋째 명령의 출력이 없다(모든 경로 실재).

[HARD] **추출 범위는 A/B/C/D 표 행(`^| [ABCD][0-9]`)에 한정하고, 행 수와 경로 수를 같은 영역에서 함께 센다.**
이전 판은 파일 어디서든 백틱 경로를 긁었고, 그 결과 **B군 표 8행을 통째로 지워도 통과했다** —
B군 행이 경로를 `.../ko/advanced/...` 로 줄여 쓰고 있어 계수에 기여하지 않았고, 같은 네 경로가 D군·§7-B 표에서
따로 잡혔기 때문이다(실측: B군 삭제 변이에서 경로 12 유지, 행 32 → 24). 두 정정이 이 구멍을 닫는다:
B군 행의 경로를 **전체 경로로 되돌렸고**(이 수리에서 수행, 실측 재확인 경로 12), 추출 범위를 표 행으로 좁혀
행 수와 경로 수가 **같은 영역**을 보게 했다. 이제 표 하나가 사라지면 두 수치가 함께 떨어진다.
부수 효과로 §7-A 의 글로브 행(`docs-site/content/*/…`)이 계수에서 빠진다 — 그 행들은 `A~D` 행이 아니므로
범위 밖이고, 경로가 `*` 를 포함해 실재 검사 대상도 아니다.
**행 번호는 검증 대상이 아니다** — 이 카드의 편집이 행을 움직이므로, 열거표의 `행` 열은 base `1d150a27d` 에서의
위치라고 산출물이 스스로 밝힌다(`tab-count-sites.md § 재현 명령` 머리말). 검증 가능한 축(경로·완전성)만 단정한다.

### AC-TCD-012 — 낱말 표기 6자리에 두 번째 층이 있다 (셸 축 + 가드 축)

리터럴 층(AC-TCD-002)만으로는 낱말 표기 자리가 재작성 한 번에 빠져나간다. 이 AC 는 그 자리에 **가드의 낱말 축**이
실제로 걸리는지를 본다 — AC-TCD-006 V4~V6 이 변이로 증명한다면, 이쪽은 축 자체의 적중 집합을 단정한다.

이 AC 는 **두 대상**을 구속한다. 셸 파이프라인만 구속하면 `REQ-TCD-005` 의 "숫자와 낱말 모두" 조항이
가드 쪽에서 비는데, 그 조항이 겨누는 것은 셸이 아니라 가드다.

| 축 | 구속 대상 | 판정 |
|---|---|---|
| 셸 축 | `tab-count-sites.md § 재현 명령 5` 파이프라인 | M2~M3 후 0행 |
| **가드 축** | `internal/web/docs_tab_contract_test.go` 의 Go 가드 | M1 후·M2 전 트리에서 적중 집합이 열거 6자리와 **같다** |

**RED-now 셀 (셸 축)**

- 명령: `tab-count-sites.md § 재현 명령 5` (낱말 수사 스윕 + 서수 접두 제외 + 파일 범위를 가진 `codex` 허용 규칙).
- 실측(수리 시점, base = 그때의 merge-base `1d150a27d`): **6행** — `docs-site/content/en/cli-reference/web.md`,
  `docs-site/content/zh/cli-reference/web.md`, `docs-site/content/en/advanced/moai-web-console.md`,
  `README.md`, `README.ko.md`, `README.zh.md` (= A2, A4, B4, C1, C2, C4).
- 붉은 이유: 낱말 표기 6자리가 아직 살아 있다.
- 단계별 근거: 원시 스윕 **8행** → 서수 접두(`第`) 제외 후 **7행**(`zh:165` 의 `第三方` 빠짐) →
  `codex` 허용 규칙 적용 후 **6행**. **허용되지 않은 오탐 0.**

**RED-now 셀 (가드 축)**

- 명령: `go test ./internal/web/ -run 'TestDocsTabContract/allowlist' -v 2>&1 | grep -c '^word-axis hit '`
- 붉은 이유: **`internal/web/docs_tab_contract_test.go` 가 아직 없다.** 실행해 본 적 없는 명령이며,
  0 이라는 수치를 여기 적지 않는다 — 선택자가 아무것도 고르지 않아도 `ok` 와 종료코드 0 이 나오므로(AC-TCD-004)
  그 0 은 "적중이 없다" 가 아니라 "잰 적이 없다" 는 뜻이고, 재지 않은 값을 관측처럼 적는 것이 이 카드가 막는 부류다.
  가드 파일의 부재는 `ls internal/web/docs_tab_contract_test.go` 로 지금 확인되는 조건이다.
- green path: **M1** — 가드가 이 출력을 내기 시작한다. M2 이전 트리에서 재므로 M2~M4 의 문서 편집과 독립이다.

**Given** M2~M3 이 끝난 트리에서,
**When** `tab-count-sites.md § 재현 명령 5` 를 그대로 실행하면,
**Then** 출력이 없다(0행).
**And** 같은 명령이 M2 이전에는 6행을 내야 하며, 그 출력을 `progress.md §E.2` 에 축어로 남긴다.
**And** 허용 규칙의 `grep -vE` 패턴에 **파일 경로가 포함**되어 있어야 한다 — `grep -v 'codex'` 로 줄이면 README 3행이
함께 사라져 6행이 3행으로 보이고, 그 상태로 M2 를 마치면 README 3자리가 검증되지 않은 채 통과한다(실측).

**And [HARD] 가드 자신이 낱말 축 적중 집합을 찍고, 그 집합이 열거 6자리와 같다.**
M1 이 끝나고 M2 가 시작되기 전의 트리에서

```bash
go test ./internal/web/ -run 'TestDocsTabContract/allowlist' -v 2>&1 | grep '^word-axis hit '
```

를 실행하면, 출력이 **정확히 6줄**이고 각 줄이 `word-axis hit <파일경로>: <적중 구절>` 형태이며,
6줄의 파일·구절이 A2·A4·B4·C1·C2·C4 와 **같다** — 즉 다음 다섯 수사 토큰이 모두 한 줄 이상을 낸다:

| 토큰 | 로케일 | 자리 |
|---|---|---|
| `nine` | en | A2 `The nine settings tabs` |
| `九`(`九个`) | zh | A4 `设置九个标签页` |
| `fourteen` | en | B4 `unfolds fourteen tabs`, C1 `fourteen tabs` |
| `열네` | ko | C2 `열네 개 탭` |
| `十四` | zh | C4 `十四个标签页` |

**5줄 이하이거나 위 다섯 토큰 중 하나라도 0줄이면 FAIL 이다.** 6줄이라도 집합이 다르면 FAIL 이다.

**이 절이 죽이는 변이체 — 낱말 클래스를 `{fourteen}` 하나로만 구현한 가드.** 그 가드는
AC-TCD-006 의 V1·V2(이름) · V3(숫자) · V4(`fourteen`) 를 전부 통과하고 셸 파이프라인을 실행하는 위 Then 절도 통과하지만,
여기서는 `fourteen` 자리 2줄만 찍어 **6줄이 아니므로 FAIL 한다**. 그 가드는 `nine`·`九`·`열네`·`十四` 를 놓쳐
열거된 낱말 표기 6자리 중 5자리를 가드 밖에 남기며, 그것이 이 카드가 지난 청소의 실패 원인으로 지목한 부류다
(`spec.md §1.3` 첫 함정). 셸 파이프라인만으로는 이 변이체를 잡지 못한다 — 파이프라인은 셸을 구속하지
`REQ-TCD-005` 가 말하는 **가드**를 구속하지 않기 때문이다.

**ja 에 열거된 낱말 표기 자리는 없다** — ja 4자리는 전부 숫자 표기이며(`tab-count-sites.md` A/B/C 표),
따라서 위 집합에 ja 줄이 없는 것은 누락이 아니라 실측이다. ja 수사 낱말은 이 AC 가 구속하지 않는다.

---

## RG — 회귀 가드 (이미 초록, 깨뜨리지 않았음을 본다)

### RG-TCD-001 — 스크린샷 미변경

**Given** 카드가 마감될 때,
**When** `git merge-base develop HEAD` 을 기록한 뒤 `git diff --quiet <그 값> -- assets/images/` 를 실행하면,
**Then** 종료코드가 `0` 이다(무변경). 종료코드가 판정을 운반한다 — `echo "rc=$?"` 를 뒤에 붙이지 않는다
(그 형태는 `git diff` 의 성공 여부를 찍을 뿐 무매치를 뜻하지 않는다).
왼쪽 끝은 AC-TCD-010 과 같은 이유로 **읽는 시점 merge-base** 다 — 리터럴 핀을 쓰면 develop 이 언젠가
`assets/images/` 를 건드린 뒤 이 카드가 흡수하는 순간, 남의 이미지 변경이 t530 의 것으로 보고된다.
오늘은 노출이 없다(실측: `git diff --name-only 1d150a27d develop -- assets/images` 빈 출력; 위 명령 종료코드 `0`) —
지금 고치는 것은 구조이지 증상이 아니다.
**And** 결정 기록이 남아 있다: `grep -c '## 5. 스크린샷 범위 판단' .moai/specs/SPEC-DOCS-TABCOUNT-DRIFT-001/spec.md` → `1`.
도착 시점 실측: 변경 파일 없음 — 이미 초록이므로 인수 기준이 아니다.

### RG-TCD-002 — docs-site 빌드 경고 0

**Given** 문서 수정이 끝난 트리에서,
**When** `cd docs-site && hugo --gc --minify 2>&1 | grep -ciE 'warn|error'` 를 실행하면,
**Then** 출력이 `0` 이다.
도착 시점 실측: 무경고(감사가 확인). 이미 초록이므로 인수 기준이 아니다 — 이 카드가 경고를 **만들지 않았음**을 본다.

### RG-TCD-003 — 열거 산출물이 줄어들지 않는다

**Given** 카드가 마감될 때,
**When** `grep -c '^| [ABCD][0-9]' .moai/reports/t530/tab-count-sites.md` 를 실행하면,
**Then** 출력이 `32` 이상이다.
도착 시점 실측: **32**(A군 4 + B군 8 + C군 8 + D군 12). D군 산문 4자리를 표 행으로 올린 뒤의 값이다 —
산문으로만 두면 문단을 통째로 지워도 이 수가 변하지 않아 완전성을 검증하지 못했다.

---

## §D.1 엣지 케이스

- **가드가 문서를 못 읽는 경우**: 파일 부재·경로 오류는 `t.Skip` 이 아니라 `t.Fatal` 이어야 한다. skip 은 판정 전 이탈이고, 초록으로 보인다.
- **빈 패턴 파일**: `count-literals.txt` 가 비면 `grep -f` 는 무매치로 `0` 을 낸다. AC-TCD-001·002 는 파일 줄 수를 함께 센다.
- **빈 선택자**: `-run` 이 없는 이름을 고르면 `no tests to run` 과 함께 `ok` 가 나온다(실측). AC-TCD-004 는 그 토큰의 부재를 단정한다.
- **허용 목록의 성장**: 허용 항목이 2 이상이면 AC-TCD-007 이 FAIL 한다. 목록이 조용히 자라 가드를 무력화하는 경로를 닫는다.
- **로케일별 수사 표기**: 낱말 축은 정규식으로 닫지 않는다(`두 곳`·`two tabs`·`两个标签页` 같은 정당한 계수까지 걸린다 — 실측).
  낱말은 리터럴 집합이 닫고, 탭 수 변화는 N2(이름 대조)가 잡는다. 새로 쓰인 낱말 계수는 남는 구멍이며 `spec.md §7` 에 있다.

## §D.2 Definition of Done

1. AC-TCD-001 ~ 012 전부 PASS, RG-TCD-001 ~ 003 전부 유지. 각 명령의 출력이 `progress.md §E.2` 에 축어로 남는다.
2. AC 각각의 **RED-now 출력과 green 출력이 모두** 기록된다 — 초록만 기록된 항목은 미완이다.
3. `go test ./internal/web/...` 전체 통과 — 새 가드가 기존 테스트를 깨지 않았다.
4. D군 12자리가 렌더 라벨(`GLM Settings` 계열)로 정합되었고, N2 가드가 그 일치를 단정한다(§C 결정, 2026-09-12).
5. 후속 카드 2건(스크린샷 절차·재촬영)이 리드에게 카드 요청으로 전달됐다 — 이 카드에서 만들지 않는다.
6. frontmatter `status` 는 `spec.md` 한 곳에서만 전이된다(본 에이전트는 `draft` 까지만 쓴다).
   `plan.md`·`acceptance.md` 는 status 축에서 무상태다.
