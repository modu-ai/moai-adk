---
id: SPEC-DOCS-TABCOUNT-DRIFT-001
title: 인수 기준 — 설정 탭 수·이름 드리프트 차단
version: "0.2.0"
created: 2026-09-12
updated: 2026-09-12
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

**문서 수준 pin: 모든 RED-now 관측은 base `1d150a27d` 에서 쟀다.** 개별 AC 가 따로 pin 하지 않으면 이 값이 적용된다
(`verification-completeness.md §2.1`). 브랜치 이름을 기준으로 잰 수치는 하나도 없다 — 이 저장소는 리드가 `develop` 을
일괄 push 하므로 카드 수명 중 움직이는 ref 는 반드시 전진한다.

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
| AC-TCD-005 | REQ-TCD-004, 005 | 가드가 읽은 파일 수 == 12 | 가드 부재 | M1+M5 |
| AC-TCD-006 | REQ-TCD-004 | 변이 3종(파일 부류 × 로케일) | 가드 부재 | M5 |
| AC-TCD-007 | REQ-TCD-005 | 가드 N1 허용 목록 정확 일치 | 가드 부재 | M1+M5 |
| AC-TCD-008 | REQ-TCD-005 | 오탐 4자리 침묵 | 가드 부재 | M5 |
| AC-TCD-009 | REQ-TCD-006 | 이름 배열 == 렌더 라벨(순서 포함) | 가드 부재 | M4+M5 |
| AC-TCD-010 | REQ-TCD-007 | 4로케일 패리티(README 포함) | 변경 0 ⇒ 측정 불가 = FAIL | M2~M4 |
| AC-TCD-011 | REQ-TCD-001 | 열거표 경로 실재·완전성 | — (M6 에서 재확인) | M6 |

| RG | 대응 REQ | 무엇을 지키나 | 도착 시점 실측 |
|---|---|---|---|
| RG-TCD-001 | REQ-TCD-009 | 스크린샷 미변경 | 변경 파일 없음 |
| RG-TCD-002 | **없음(의도적)** — 요구사항이 아니라 빌드 위생 | docs-site 빌드 경고 0 | 무경고 |
| RG-TCD-003 | REQ-TCD-001 | 열거 산출물이 줄어들지 않음 | 32행 |

REQ 커버리지: REQ-TCD-001~009 전부 최소 1개의 AC 또는 RG 에 대응한다(001→AC-011·RG-003, 002→AC-001,
003→AC-002, 004→AC-004·005·006, 005→AC-007·008, 006→AC-003·009, 007→AC-010, 008→AC-002, 009→RG-001).
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
**왜 필요한가**: `README.md` 한 파일만 읽는 가드도 AC-TCD-004·006·008·009 를 전부 통과한다(변이 지점이 README 에 있으면
FAIL→PASS 도 재현한다). 읽은 집합의 크기를 판정 전에 세지 않으면 **아무것도 읽지 않은 초록과 전부 읽은 초록이 구분되지 않는다.**

### AC-TCD-006 — 변이 3종: 파일 부류 × 로케일

**RED-now 셀**: 가드 부재. 붉은 이유: 가드가 무엇을 잡는지 보인 적이 없다.

**Given** 가드가 GREEN 인 트리에서,
**When** 아래 세 변이를 **각각 따로** 넣고 가드를 돌린 뒤 되돌리면,

| 변이 | 대상 | 조작 |
|---|---|---|
| V1 | `README.ko.md` | 탭 이름 하나를 `Audit` → `Audits` |
| V2 | `docs-site/content/en/advanced/moai-web-console.md` | 탭 이름 하나를 `Report` → `Reports` |
| V3 | `docs-site/content/zh/cli-reference/web.md` | `14 个标签页` 를 되살림 |

**Then** 세 변이가 각각 `--- FAIL` 을 내고, 실패 메시지가 **어느 파일의 무엇**이 어긋났는지 적시하며, 되돌린 뒤에는 모두 `ok` 다.
**And** 여섯 출력(FAIL 3 + PASS 3)을 `progress.md §E.2` 에 축어로 남긴다.
V1·V2 는 서로 다른 파일 부류(README / docs-site)이면서 서로 다른 로케일(ko / en)이다 — 한 부류만 읽는 가드를 걸러낸다.
V3 는 N1 숫자 스윕이 살아 있는지 본다.

### AC-TCD-007 — N1 허용 목록이 정확히 일치한다

**RED-now 셀**: 가드 부재. 붉은 이유: 허용 목록을 들고 판정하는 주체가 없다.

**Given** M5 가 끝난 트리에서,
**When** `go test ./internal/web/ -run 'TestDocsTabContract/allowlist' -v 2>&1 | grep -E '^--- (PASS|FAIL)|allowed [0-9]+'` 를 실행하면,
**Then** `--- PASS: TestDocsTabContract/allowlist` 가 있고 `allowed 1` 을 포함한다.
허용 항목은 `advanced/moai-web-console.md` 안에서 `codex` 토큰을 포함하는 줄이며(줄 번호가 아니라 내용으로 식별 —
편집이 줄 번호를 움직인다), 이유는 `tab-count-sites.md §7-B` 에 적혀 있다.
**And** 허용 목록이 0항목이거나 2항목 이상이면 FAIL 이다 — 목록이 조용히 자라면 가드가 무력해진다.

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

### AC-TCD-010 — 4로케일 패리티 (README 포함, 고정 base)

**RED-now 셀**

- 명령: `git diff --name-only 1d150a27d -- README.md README.ko.md README.ja.md README.zh.md docs-site/content`
- 실측(base `1d150a27d`): 출력 없음 → 변경 파일 **0개**.
- 붉은 이유: 대조군이 0 이므로 **"패리티 유지" 가 아니라 "측정 불가"** 이고, 이 AC 는 그 상태를 FAIL 로 판정한다.
  이전 판은 같은 상태를 통과로 읽었다(0 × 4 == 0).

**Given** M2~M4 가 끝난 트리에서,
**When**
```bash
git diff --name-only 1d150a27d -- README.md README.ko.md README.ja.md README.zh.md docs-site/content > /tmp/t530-changed.txt
wc -l < /tmp/t530-changed.txt
grep -c '^README' /tmp/t530-changed.txt
sed -E 's#^docs-site/content/[a-z]+/##' /tmp/t530-changed.txt | grep -v '^README' | sort | uniq -c
```
를 실행하면,
**Then** 첫 수가 **1 이상**이고(0 이면 측정 불가 = FAIL), 둘째 수가 `4` 이며(README 4본 전부),
셋째 출력의 모든 상대 경로가 정확히 `4` 회씩 나온다.
기준은 고정 SHA `1d150a27d` 다 — `origin/develop...` 같은 움직이는 ref 를 쓰지 않는다. 리드의 일괄 push 로
`origin/develop` 이 전진하면 남의 변경이 분모에 섞여 판정이 스스로 뒤집힌다.

### AC-TCD-011 — 열거 산출물이 최신이다 (경로 축)

**Given** M6 에서,
**When**
```bash
grep -oE '`(README(\.(ko|ja|zh))?\.md|docs-site/content/[a-z]+/(cli-reference|advanced)/[a-z-]+\.md)`' \
  .moai/reports/t530/tab-count-sites.md | tr -d '`' | sort -u > /tmp/t530-paths.txt
wc -l < /tmp/t530-paths.txt
while read -r f; do test -f "$f" || echo "MISSING $f"; done < /tmp/t530-paths.txt
```
를 실행하면,
**Then** 첫 수가 `12` 이고, 둘째 명령의 출력이 없다(모든 경로 실재).
**행 번호는 검증 대상이 아니다** — 이 카드의 편집이 행을 움직이므로, 열거표의 `행` 열은 base `1d150a27d` 에서의
위치라고 산출물이 스스로 밝힌다(`tab-count-sites.md § 재현 명령` 머리말). 검증 가능한 축(경로·완전성)만 단정한다.

---

## RG — 회귀 가드 (이미 초록, 깨뜨리지 않았음을 본다)

### RG-TCD-001 — 스크린샷 미변경

**Given** 카드가 마감될 때,
**When** `git diff --quiet 1d150a27d -- assets/images/` 를 실행하면,
**Then** 종료코드가 `0` 이다(무변경). 종료코드가 판정을 운반한다 — `echo "rc=$?"` 를 뒤에 붙이지 않는다
(그 형태는 `git diff` 의 성공 여부를 찍을 뿐 무매치를 뜻하지 않는다).
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

1. AC-TCD-001 ~ 011 전부 PASS, RG-TCD-001 ~ 003 전부 유지. 각 명령의 출력이 `progress.md §E.2` 에 축어로 남는다.
2. AC 각각의 **RED-now 출력과 green 출력이 모두** 기록된다 — 초록만 기록된 항목은 미완이다.
3. `go test ./internal/web/...` 전체 통과 — 새 가드가 기존 테스트를 깨지 않았다.
4. D군 12자리가 렌더 라벨(`GLM Settings` 계열)로 정합되었고, N2 가드가 그 일치를 단정한다(§C 결정, 2026-09-12).
5. 후속 카드 2건(스크린샷 절차·재촬영)이 리드에게 카드 요청으로 전달됐다 — 이 카드에서 만들지 않는다.
6. frontmatter `status` 는 `spec.md` 한 곳에서만 전이된다(본 에이전트는 `draft` 까지만 쓴다).
   `plan.md`·`acceptance.md` 는 status 축에서 무상태다.
