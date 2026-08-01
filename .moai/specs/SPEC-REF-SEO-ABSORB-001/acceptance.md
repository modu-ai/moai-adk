# Acceptance — SPEC-REF-SEO-ABSORB-001

모든 항목은 명령·기대 출력·실패 형태를 함께 적는다. 0매치가 성공 조건인 선택자는 반드시 짝이 되는 존재 확인 또는 self-trip을 동반한다.

가드 클래스를 언급하는 항목은 **강제 파일을 반드시 함께 지목한다** — 리포에는 서로 대응하지 않는 3개의 독립된 C1-C8 클래스 번호 체계가 존재하므로 한정 없는 "C1 위반 없음"은 모호하다.

---

## §A. 판정 위생 규칙

### A.1 선택자 공허 방지

- `go test -run <pattern>`은 0개 매칭 시 exit 0이다. 테스트 실행을 주장하는 항목은 `-v`와 함께 실행하고 **기대 테스트 함수명이 출력에 등장함을 확인**한다.
- grep 0매치가 성공 조건인 항목은 짝이 되는 존재 확인(양성 대조)을 함께 실행한다.
- 이 셸에서 `ls`는 `ls -la` 별칭이다. 파일 열거는 `git ls-files` 또는 `find`를 쓴다.

### A.2 경로 규약

명령은 리포 루트에서 실행한다. 원문 경로는 `~/.agents/skills/higgsfield-websites/references/seo.md`이며 스크립트 내부에서 `os.path.expanduser`로 해석한다. 원문은 `sha256 = c088f089f365fae621dac90db77b56abdc47463becb2dbd182bd0cd04de98ee7`로 고정되어 있다(spec.md §E).

**판정 대상 트리는 템플릿 트리다.** 배포되는 것은 `internal/template/templates/.claude/skills/moai-ref-seo/SKILL.md`이므로 §B 클린룸 판정과 §C 콘텐츠 판정 모두 이 경로를 좌변으로 쓴다. 로컬 미러 `.claude/skills/moai-ref-seo/SKILL.md`는 AC-SEO-011b가 템플릿 트리와의 동일성을 별도로 판정한다. (v0.1.0은 클린룸 판정만 로컬 미러를 봐서, 로컬만 클린하고 템플릿에 원문 잔재가 남아도 통과하는 구멍이 있었다.)

### A.3 브랜치 전제

`origin/main...HEAD` 범위를 쓰는 항목(AC-SEO-021b)은 **run-phase feature 브랜치에서 실행한다.** `main`에서 origin과 동기(`0 0`)인 상태로 실행하면 두 값 모두 0이 나와 판정이 무의미하다.

---

## §B. 클린룸 프로토콜 (형제 SPEC 재사용 대상)

### AC-SEO-010 — 출처 라이선스 부재 (배포 디렉터리 한정)

Given 원문이 제3자 배포 디렉터리 안에 있고
When 그 배포 디렉터리에 한정해 라이선스 파일을 열거하면
Then 결과가 0건이고, 동일 명령을 상위 스킬 루트에 적용하면 0이 아니어야 한다(명령 자체가 무력하지 않음을 증명).

```bash
find ~/.agents/skills/higgsfield-websites \( -iname 'LICENSE*' -o -iname 'NOTICE*' -o -iname 'COPYING*' \) | wc -l
find ~/.agents/skills \( -iname 'LICENSE*' -o -iname 'NOTICE*' -o -iname 'COPYING*' \) | wc -l
```

기대: 첫 명령 `0`, 둘째 명령 `1` 이상 (plan-phase 실측: 각각 `0`, `1` — `notion-cli/LICENSE.md`).
실패: 첫 명령이 0이 아니면 라이선스가 존재하므로 클린룸 전제가 무너진다. 둘째 명령이 0이면 `find` 자체가 동작하지 않은 것이므로 첫 결과는 증거가 아니다.

### AC-SEO-011 — n-gram 중첩 임계값

Given 산출된 SKILL.md와 원문이 존재하고
When 아래 정규화 후 단어 8-gram 집합 교집합을 계산하면
Then `shared_8grams=0`이다.

```bash
python3 - <<'PY'
import re, os
def norm(p):
    t = open(os.path.expanduser(p), encoding='utf-8').read().lower()
    t = re.sub(r'\A---.*?^---\s*$', '', t, flags=re.S | re.M)   # frontmatter 제거
    return re.sub(r'[^a-z0-9]+', ' ', t).split()                # 소문자 + 영숫자 외 전부 공백
N = 8
a = norm('internal/template/templates/.claude/skills/moai-ref-seo/SKILL.md')
b = norm('~/.agents/skills/higgsfield-websites/references/seo.md')
ga = {' '.join(a[i:i+N]) for i in range(len(a)-N+1)}
gb = {' '.join(b[i:i+N]) for i in range(len(b)-N+1)}
sh = sorted(ga & gb)
print("shared_8grams=%d" % len(sh))
for s in sh: print("  HIT:", s)
PY
```

기대: `shared_8grams=0`.
실패: 1 이상이면 각 HIT 행이 출력된다. 해당 문단을 재작성하고 재실행한다. **임계값을 올려 통과시키는 것은 REQ-SEO-014 위반이다.**

**HIT가 순수 기술 용어 연쇄인 경우의 통제된 출구**: 도메인 필수 어휘가 우연히 8단어 일치할 수 있다(예: canonical URL 규칙의 정형 문구). 이 경우에도 임계값 조정은 금지이며, 절차는 다음과 같다 — (1) 먼저 재작성을 시도한다, (2) 재작성해도 의미가 깨져 불가피하면 HIT 전문과 재작성 시도 내역을 `progress.md` §E.2에 기록하고 사람 승인을 받는다, (3) 승인 없이 통과 선언하지 않는다. 예외를 봉인하는 것보다 통제된 출구를 두는 편이 규칙 준수율이 높다.

대조 baseline (plan-phase 실측, 원문과 무관한 기존 ref 스킬):

| 비교 | 실측값 |
|---|---|
| `moai-ref-ui-polish` vs 원문 | `shared_8grams=0` |
| `moai-ref-api-patterns` vs 원문 | `shared_8grams=0` |

### AC-SEO-011b — 로컬 미러와 템플릿 트리 동일성

Given 클린룸 판정과 콘텐츠 판정이 템플릿 트리를 대상으로 하고
When 로컬 미러와 비교하면
Then 두 파일이 바이트 동일하다.

```bash
diff internal/template/templates/.claude/skills/moai-ref-seo/SKILL.md \
     .claude/skills/moai-ref-seo/SKILL.md && echo "IDENTICAL"
shasum -a 256 internal/template/templates/.claude/skills/moai-ref-seo/SKILL.md \
              .claude/skills/moai-ref-seo/SKILL.md
```

기대: `diff` 무출력 + `IDENTICAL` 출력, 두 sha256 값이 동일.
실패: 차이가 있으면 §B·§C 판정이 배포되지 않는 파일에 대한 것이었다는 뜻이다. 템플릿 트리를 정본으로 삼아 로컬 미러를 다시 복사하고 §B·§C를 재실행한다.

### AC-SEO-012 — self-trip 시연 (검사가 공허하지 않음을 증명) + 원문 고정

Given AC-SEO-011과 동일한 명령에서 좌변을 원문 자신으로 바꾸고
When 고정된 원문 digest를 먼저 단언한 뒤 실행하면
Then digest가 일치하고, 공유 8-gram이 임계값을 크게 초과해 FAIL 해야 한다.

```bash
python3 - <<'PY'
import re, os, hashlib
PINNED = 'c088f089f365fae621dac90db77b56abdc47463becb2dbd182bd0cd04de98ee7'
S = '~/.agents/skills/higgsfield-websites/references/seo.md'
raw = open(os.path.expanduser(S), 'rb').read()
digest = hashlib.sha256(raw).hexdigest()
print("source_sha256=%s" % digest)
print("source_pin_match=%s" % (digest == PINNED))
def norm(p):
    t = open(os.path.expanduser(p), encoding='utf-8').read().lower()
    t = re.sub(r'\A---.*?^---\s*$', '', t, flags=re.S | re.M)
    return re.sub(r'[^a-z0-9]+', ' ', t).split()
N = 8
a = norm(S); b = norm(S)
ga = {' '.join(a[i:i+N]) for i in range(len(a)-N+1)}
gb = {' '.join(b[i:i+N]) for i in range(len(b)-N+1)}
print("selftrip_shared_8grams=%d" % len(ga & gb))
PY
```

기대: `source_pin_match=True` **그리고** `selftrip_shared_8grams=4600`.

판정 규칙 — 두 출력을 분리해서 읽는다.

| `source_pin_match` | `selftrip_shared_8grams` | 판정 |
|---|---|---|
| `True` | `4600` | PASS |
| `True` | `1000` 초과 `4600` 아님 | 조사 필요 — 정규화 코드가 바뀌었는지 확인 |
| `True` | `1000` 이하 | **FAIL** — 정규화가 텍스트를 삼켰다. AC-SEO-011의 `0`은 증거가 아니며 그 결과를 무효 처리한다 |
| `False` | (무관) | **baseline 무효** — 제3자 원문이 갱신되었다. 산출물의 결함이 아니므로 산출물을 고치지 말고, spec.md §E의 digest와 `4600` 기대값을 재측정해 갱신한 뒤 다시 판정한다 |

마지막 행이 이 AC의 존재 이유다: 고정이 없으면 원문 갱신이 산출물 결함으로 오독된다.

### AC-SEO-013 — 최장 공통 부분문자열 상한

Given 정규화된 두 문자열에서
When 최장 공통 부분문자열 길이를 구하면
Then 40자 이하다.

```bash
python3 - <<'PY'
import re, os
from difflib import SequenceMatcher
def norm(p):
    t = open(os.path.expanduser(p), encoding='utf-8').read().lower()
    t = re.sub(r'\A---.*?^---\s*$', '', t, flags=re.S | re.M)
    return re.sub(r'[^a-z0-9]+', ' ', t).strip()
a = norm('internal/template/templates/.claude/skills/moai-ref-seo/SKILL.md')
b = norm('~/.agents/skills/higgsfield-websites/references/seo.md')
m = SequenceMatcher(None, a, b, autojunk=False).find_longest_match(0, len(a), 0, len(b))
print("lcs_chars=%d" % m.size)
print("lcs_text=%r" % a[m.a:m.a+m.size])
PY
```

기대: `lcs_chars` ≤ 40.
실패: 40 초과 시 `lcs_text`가 출력하는 구절을 재작성한다.

대조 baseline (plan-phase 실측): `moai-ref-ui-polish` 23자(`'nts is the most common '`), `moai-ref-api-patterns` 17자(`' error responses '`).

**임계값 근거 (v0.1.0의 60에는 근거가 없었다)**: 원문과 무관한 기존 ref 스킬 2건의 실측 최대가 23자다. 40은 그 1.7배로, 무관한 문서 사이에서 자연 발생하는 우연 일치 폭을 여유 있게 덮으면서 의도적 복제를 잡는 위치다. 60은 정규화 후 영어 기준 약 9-10 단어여서 8-gram 공유가 0이면 사실상 도달 불가였고, AC-SEO-011이 통과하는 한 실패할 수 없어 독립 판정력이 없었다.

**이 AC의 역할**: AC-SEO-011보다 촘촘한 그물이 아니라, AC-SEO-011이 HIT를 냈을 때 **위반 구절의 위치를 특정하는 진단 도구**로도 쓴다(`lcs_text`가 그 구절을 출력한다).

### AC-SEO-014 — 출처 없는 수치 3건 미재현

Given 원문의 무출처 수치 주장 3건이 있고
When 산출물에서 해당 수치 형태를 검색하면
Then 매칭이 0이다. 양성 대조로 원문에서 동일 패턴을 검색하면 1 이상이어야 한다.

```bash
PAT='80%|100ms|100-300ms|300ms'
grep -nE "$PAT" internal/template/templates/.claude/skills/moai-ref-seo/SKILL.md | wc -l
grep -nE "$PAT" ~/.agents/skills/higgsfield-websites/references/seo.md | wc -l
```

기대: 첫 명령 `0`, 둘째 명령 `1` 이상 (plan-phase 실측: 원문 `3`).
실패: 첫 명령이 0이 아니면 해당 행을 제거하거나 검증 가능한 출처를 붙인다. 둘째 명령이 0이면 패턴이 잘못된 것이므로 첫 결과는 증거가 아니다.

### AC-SEO-015 — 구조 발산 판정 (8-gram·LCS가 잡지 못하는 실패 양식)

Given 8-gram과 LCS가 둘 다 **어휘 연속성** 지표이고
When 원문의 섹션 순서와 표 열 구성을 그대로 둔 채 각 셀을 동의어로 치환하면
Then 두 검사를 모두 통과하면서도 파생물이 된다 — 따라서 이 실패 양식은 사람 판독으로만 잡힌다.

research.md §A.5가 원문의 구조 골격(topic-major 배열, 섹션당 근거 산문 → 하위 섹션 → 레퍼런스 표 → 말미 5항목 pitfalls)을 상세히 기록했기 때문에, 이 실패 양식은 **더 쉽게 발생한다**. REQ-SEO-015가 선언한 "형제 SPEC이 재도출 없이 채택 가능"은 프로토콜이 가장 도달하기 쉬운 실패 양식을 다룰 때에만 참이다.

판독 보조 — 두 문서의 구조를 나란히 출력한다. **각 블록은 출력 행 수를 함께 세어, 0행이 "대응 없음"으로 오독되는 것을 막는다**(§A.1 선택자 공허 방지).

```bash
ART=internal/template/templates/.claude/skills/moai-ref-seo/SKILL.md
SRC=~/.agents/skills/higgsfield-websites/references/seo.md

echo "=== 산출물 H2 순서 ==="
grep -n '^## ' "$ART";                          echo "count=$(grep -c '^## ' "$ART")"
echo "=== 원문 최상위 섹션 순서 ==="
grep -nE '^#{1,2} ' "$SRC";                     echo "count=$(grep -cE '^#{1,2} ' "$SRC")"
echo "=== 산출물 표 헤더 행 ==="
grep -n '^|' "$ART" | grep -B1 -- '---' | grep -v -- '---' | grep '^[0-9]'
echo "count=$(grep -n '^|' "$ART" | grep -B1 -- '---' | grep -v -- '---' | grep -c '^[0-9]')"
echo "=== 원문 표 헤더 행 ==="
grep -n '^|' "$SRC" | grep -B1 -- '---' | grep -v -- '---' | grep '^[0-9]'
echo "count=$(grep -n '^|' "$SRC" | grep -B1 -- '---' | grep -v -- '---' | grep -c '^[0-9]')"
```

> **v0.2.0 명령 실행 불가 정정 (N1)**: v0.2.0의 표 헤더 명령은 `grep -- '---' -B1` 형태였다. `--`가 옵션 파싱을 종료시키므로 `-B1`이 **파일명으로 해석**되어 `grep: -B1: No such file or directory` + exit 1 + 출력 0행이 된다. 원문·산출물 두 명령이 같은 형태였으므로 **표 열 구성 판정의 양쪽 근거가 모두 비어 있었다.** 판독자는 질문 2에 근거 없이 답하고 그 답이 `structural_divergence: PASS`로 기록되는 상태였다 — §A.1과 plan.md §G가 명시적으로 금지한 실패 양식이다. `-B1`을 `--` 앞으로 옮겨 정정했다.

**원문 측 baseline (plan-phase 실측, 명령이 살아 있음을 고정)**:

```
$ grep -nE '^#{1,2} ' "$SRC"          →  count=7
1:# SEO
3:## Meta tags & OG
148:## Technical SEO
362:## Schema markup
531:## Entity SEO
653:## GEO / content
756:## Audit

$ grep -n '^|' "$SRC" | grep -B1 -- '---' | grep -v -- '---' | grep '^[0-9]'   →  count=10
83:| Page type | `robots` value |
95:| Intake field | Meta target |
368:| Site type | Schema types to apply |
425:| Field | Value |
435:| Field | Value |
444:| Field | Value |
456:| Field | Value |
467:| Field | Value |
477:| Field | Value |
541:| Field          | Example                                      | Required |
```

**전제 검사 (판정 전 반드시 확인)**: 원문 두 `count`는 각각 `7`·`10`이어야 한다. 어느 하나라도 `0`이면 명령이 다시 깨진 것이므로 **판정을 진행하지 말고 명령을 고친다.** 산출물 두 `count`가 `0`이면 그것은 "구조가 대응하지 않는다"가 아니라 **산출물이 아직 없거나 H2/표가 없다**는 뜻이며, 역시 판정 불가다. 네 `count`가 모두 1 이상일 때에만 아래 질문에 답한다.

기대: 위 4개 출력을 인용한 뒤, 판독자가 다음 두 질문에 답한 **판정 기록**을 `progress.md` §E.2에 남긴다.

1. 산출물의 섹션 순서가 원문의 섹션 순서와 1:1 대응하는가? (대응하면 위반)
2. 산출물의 표 열 구성이 원문 대응 표의 열 구성과 1:1 대응하는가? (대응하면 위반)

판정 기록 형식: `structural_divergence: PASS|FAIL` + 각 질문에 대한 2-3문장 근거. 근거 없는 `PASS` 한 줄은 판정으로 인정하지 않는다.

실패: 어느 한 질문이라도 "대응한다"이면 해당 섹션을 MoAI 자체 조직 원리(도메인 H2 4-10개, 표 중심)로 재배열하고 재판독한다. **완전 기계화는 어렵다 — 이 AC는 판정 부재를 명시적 사람 판독으로 대체하는 것이지, 기계 판정을 흉내내지 않는다.**

---

## §C. 스킬 콘텐츠

### AC-SEO-001 — 파일 존재와 단일 파일 구성

Given 저작이 끝났을 때
When 스킬 디렉터리를 열거하면
Then 양쪽 트리에 `SKILL.md` 하나씩만 존재한다.

```bash
find internal/template/templates/.claude/skills/moai-ref-seo .claude/skills/moai-ref-seo -type f | sort
```

기대: 정확히 2행, 둘 다 `SKILL.md`로 끝난다.
실패: 3행 이상이면 `references/`나 `modules/`가 생긴 것이다(REQ-SEO-001 위반).

### AC-SEO-002 — 분량과 구조

```bash
wc -l internal/template/templates/.claude/skills/moai-ref-seo/SKILL.md
grep -c '^## ' internal/template/templates/.claude/skills/moai-ref-seo/SKILL.md
python3 - <<'PY'
import re
t = open('internal/template/templates/.claude/skills/moai-ref-seo/SKILL.md', encoding='utf-8').read()
t = re.sub(r'\A---.*?^---\s*$', '', t, flags=re.S | re.M)   # frontmatter 제거
print("body_h1=%d" % len(re.findall(r'^# ', t, re.M)))
PY
```

기대: 줄 수 150-220(파일 전체, frontmatter 포함), H2 개수 **7-14**, `body_h1=1`.

> **v0.3.1 판정 명령 공허 정정 (N12)**: 이 자리의 H1 판정은 `grep -c '^# '`를 파일 전체에 적용했다. frontmatter의 `progressive_disclosure:` 블록에 `# MoAI Extension: Progressive Disclosure` 주석 행이 있어 **그 행이 H1로 계수된다.** 템플릿 트리의 `moai-ref-*` 10건 전수 실측:
>
> ```
> secops H1=2 / ui-polish H1=2 / api-patterns H1=2 / react-patterns H1=2 / owasp-checklist H1=2
> llm-security H1=2 / supply-chain H1=2 / testing-pyramid H1=2 / git-workflow H1=2 / seo H1=2
> ```
>
> 전 10건이 frontmatter에 `progressive_disclosure:`를 갖는다. 따라서 **어떤 ref 스킬도 "정확히 1"을 만족할 수 없고**, 선례 9건이 9/9 실패한다 — 산출물이 아니라 판정 명령의 결함이다(§A.1이 금지하는 실패 양식이며, REQ-SEO-015가 이 프로토콜을 형제 SPEC 재사용 대상으로 선언하므로 방치하면 결함이 전파된다). frontmatter를 제거하고 세면 10건 전부 `H1=1`이고 H2는 불변이다:
>
> ```
> api-patterns H1=1 H2=11 / git-workflow H1=1 H2=16 / llm-security H1=1 H2=14
> owasp-checklist H1=1 H2=11 / react-patterns H1=1 H2=11 / secops H1=1 H2=9
> seo H1=1 H2=12 / supply-chain H1=1 H2=15 / testing-pyramid H1=1 H2=13 / ui-polish H1=1 H2=13
> ```
>
> 정정은 **기준을 완화하지 않는다** — AC가 처음부터 의도한 "본문 H1"을 실제로 재게 만들 뿐이며 임계값 `1`은 그대로다. frontmatter 제거 관용구는 AC-SEO-011/012/013이 이미 쓰는 `re.sub(r'\A---.*?^---\s*$', '', t, flags=re.S | re.M)`를 그대로 재사용해 파일 내 일관성을 유지한다. `wc -l`과 `grep -c '^## '`는 이 결함의 영향을 받지 않으므로 명령·기대 범위 모두 불변이다(산출물 실측: 207줄 / H2 12 — 둘 다 범위 내).

H2 상한 산식: 도메인 4-10 + 불변 3 + 선택적 `## Target Agents` 0-1 = **7-14**. `## Target Agents`는 ref 스킬의 표준 선택 구성요소다(research.md §B.3). v0.1.0의 상한 13은 이 선택 섹션을 셈에 넣지 않아, 선택 섹션을 쓰면서 도메인 H2를 상한까지 채우면 hard-fail하는 값이었다.

**선례 실측 (plan-phase, `grep -c '^## '`)** — v0.2.0은 이 자리에 "`moai-ref-ui-polish`가 H2 11개 + 불변 3 = 14"라고 적었으나 실측은 **13**이다(N4 정정):

```
secops 9 / api-patterns 11 / owasp-checklist 11 / react-patterns 11
testing-pyramid 13 / ui-polish 13 / llm-security 14 / supply-chain 15 / git-workflow 16
```

`moai-ref-ui-polish` H2 13개 = 명명 10(Target Agents / Core Philosophy / Geometry and Alignment / Elevation and Structure / Motion / Typography / Imagery / Interaction / Icons / Review Modes) + 불변 3. research.md §B.4의 "H2 11개" 기재가 1건 과다 계상이었고 v0.2.0이 그것을 물려받았다.

**상한 14는 유지한다.** 근거는 선례 최대값이 아니라 REQ-SEO-002가 고정한 도메인 H2 4-10에서 독립적으로 도출되기 때문이다. 실측 최대는 `git-workflow` 16이지만, 그 스킬들이 도메인 H2를 10개 넘게 쓴다는 사실은 **이 스킬의 상한을 올릴 근거가 아니다** — REQ-SEO-002가 이 스킬을 4-10으로 제한하기로 이미 결정했다.

실패: 범위를 벗어나면 압축 또는 분할한다.

### AC-SEO-003 — 불변 3종과 evolvable-zone id

```bash
grep -n 'moai:evolvable-start id="rationalizations"\|moai:evolvable-start id="red-flags"\|moai:evolvable-start id="verification"' \
  internal/template/templates/.claude/skills/moai-ref-seo/SKILL.md
grep -n '^## Common Rationalizations\|^## Red Flags\|^## Verification' \
  internal/template/templates/.claude/skills/moai-ref-seo/SKILL.md
```

기대: 각 명령 3행. 둘째 명령의 행 번호가 Common Rationalizations < Red Flags < Verification 순서로 증가한다.
실패: id 오타나 순서 역전은 기존 ref 스킬 9건과의 불일치를 만든다.

### AC-SEO-004 — frontmatter 길이 상한

```bash
python3 - <<'PY'
import re
t = open('internal/template/templates/.claude/skills/moai-ref-seo/SKILL.md', encoding='utf-8').read()
fm = re.match(r'\A---\n(.*?)^---\s*$', t, re.S | re.M).group(1)
d = re.search(r'^description:\s*>\n((?:[ \t]+.*\n)+)', fm, re.M).group(1)
w = re.search(r'^when_to_use:\s*>\n((?:[ \t]+.*\n)+)', fm, re.M).group(1)
n = len(' '.join(d.split())) + len(' '.join(w.split()))
print("desc_plus_when_chars=%d" % n)
PY
```

기대: 1536 미만.
실패: 초과 시 두 필드를 줄인다. 정규식이 매칭에 실패해 예외가 나면 frontmatter 형식(folded scalar `>`)이 표준을 벗어난 것이다.

### AC-SEO-005 — description 3악장

```bash
python3 - <<'PY'
import re
t=open('internal/template/templates/.claude/skills/moai-ref-seo/SKILL.md',encoding='utf-8').read()
d=' '.join(re.search(r'^description:\s*>\n((?:[ \t]+.*\n)+)',t,re.M).group(1).split())
print("ends_reference_phrase=", bool(re.search(r'reference\b', d, re.I)))
print("has_amplify=", 'Agent-extending skill that amplifies' in d)
print("has_not_for=", 'NOT for:' in d)
nf = d.split('NOT for:', 1)[1] if 'NOT for:' in d else ''
print("not_for_names_geo=", bool(re.search(r'generative[- ]engine', nf, re.I)))
print("not_for_names_a11y=", all(re.search(p, nf, re.I) for p in
      [r'keyboard', r'focus', r'form.{0,12}label|label.{0,12}form']))
PY
```

기대: 다섯 값 모두 `True`.

**정규식 확장 근거 (N5 정정) — 기존 9건 전수 양성 대조**: v0.2.0의 `reference[:,]`는 `reference` 뒤에 콜론이나 쉼표를 강제했다. REQ-SEO-005는 "a scope noun-phrase ending in `reference`"만 요구하며 뒤 구두점을 규정하지 않는다. 기존 9건에 적용한 실측:

| 정규식 | 통과 | 판정 |
|---|---|---|
| `reference[:,]` (v0.2.0) | **4 / 9** — llm-security · secops · supply-chain · ui-polish | 관례를 따르는 5건을 오탐 탈락시킨다 |
| `reference\b` (채택) | **9 / 9** | research.md §B.2의 "9건 전수 `reference`로 종료" 구조 주장과 일치 |

탈락하던 5건(api-patterns / git-workflow / owasp-checklist / react-patterns / testing-pyramid)은 전부 `reference`로 끝나는 범위 명사구를 갖고 있고 뒤따르는 문자가 공백이거나 마침표일 뿐이다(예: `… and input validation reference for backend development.`). 즉 결함은 선례가 아니라 정규식에 있었다. 같은 실측에서 `has_amplify`·`has_not_for`는 **9/9 True**였으므로 이 두 토큰은 처음부터 유효한 판별자다.

**`not_for_names_geo` / `not_for_names_a11y` 신설 근거**: REQ-SEO-005는 v0.2.0에서 `NOT for:` 절이 GEO 제외와 접근성 조작성 3종 위임을 **명시하도록** 확장되었으나, AC는 `NOT for:` 문자열의 존재만 보고 내용을 보지 않았다. plan.md §B.3·§B.4 결정의 유일한 산출물이 그 절의 내용이므로, 존재만 판정하면 두 결정 모두 무판정으로 남는다. 두 토큰이 그 공백을 닫는다.

**두 신설 토큰의 비공허성 (plan-phase 실측)**: 산출물이 아직 없으므로 스크립트를 합성 입력 2건에 돌려 판별력을 확인했다. 조건을 갖춘 `description`(`NOT for:` 절이 keyboard / focus / form label + generative-engine을 명명)은 5개 값 전부 `True`, `NOT for:` 절은 있으나 그 결정들을 담지 않은 입력은 앞 3개 `True` + **뒤 2개 `False`**. 즉 새 토큰은 `NOT for:` 존재만으로 통과하지 않는다.

```
--- conforming ---            ends=True amplify=True not_for=True geo=True  a11y=True
--- non-conforming ---        ends=True amplify=True not_for=True geo=False a11y=False
```

실패: `has_amplify`가 False면 불변 역할 문장이 없다. `has_not_for`가 False면 부정 범위절이 없다. `not_for_names_geo`/`not_for_names_a11y`가 False면 절은 있으나 §B.3·§B.4 결정이 본문에 도달하지 않은 것이다 — 절을 채운다.

### AC-SEO-006 — 내구성 높은 개념 커버리지 (12종)

REQ-SEO-006의 개념 집합 12종 각각에 판정 토큰을 둔다. 뒤 4종은 plan.md §B.3 결정으로 편입된 SEO 인과 접근성 항목이다.

```bash
TARGET=internal/template/templates/.claude/skills/moai-ref-seo/SKILL.md
while IFS='|' read -r label pat; do
  printf '%-20s %s\n' "$label" "$(grep -icE "$pat" "$TARGET")"
done <<'EOF'
canonical|canonical
robots.txt|robots\.txt
sitemap.xml|sitemap\.xml
json-ld|json-ld
meta-description|meta description
title|\btitle\b
entity|\bentit(y|ies)\b
header-chokepoint|redirect
heading-structure|\bh1\b|heading level
image-alt|\balt\b.*(attribute|text)|alt=
anchor-text|anchor text
fragment-target|fragment
EOF
```

기대: 12개 토큰 전부 1 이상.

**양성 대조 (패턴이 실제로 매칭하는가 — plan-phase 실측)**: 동일 루프를 원문에 적용한 결과. 12개 전부 1 이상이므로 패턴 자체는 유효하다.

| 토큰 | 원문 | 토큰 | 원문 |
|---|---|---|---|
| canonical | 14 | header-chokepoint(`redirect`) | 4 |
| robots.txt | 4 | heading-structure | 4 |
| sitemap.xml | 3 | image-alt | 3 |
| json-ld | 9 | anchor-text | 3 |
| meta-description | 2 | fragment-target | 8 |
| title | 17 | entity | 16 |

**음성 대조 (공허하지 않은가 — plan-phase 실측)**: SEO와 무관한 기존 ref 스킬 3건에 동일 루프를 적용했다.

| 대조 스킬 | 12토큰 중 1 이상인 개수 | 비고 |
|---|---|---|
| `moai-ref-ui-polish` | 0 | 전 토큰 0 |
| `moai-ref-api-patterns` | 0 | 전 토큰 0 |
| `moai-ref-git-workflow` | 1 | `\btitle\b`만 1 |

v0.1.0의 토큰 집합에는 공허 토큰이 있었다 — 맨 `description`은 frontmatter의 `description:` 필드 때문에 어떤 스킬에서든 항상 1 이상이었고(ui-polish 1, git-workflow 9), 맨 `header`도 무관 스킬에서 매칭됐다(api-patterns 3, owasp-checklist 11). 각각 `meta description`과 `redirect`로 좁혀 음성 대조를 통과시켰다. 남은 최약 토큰은 `\btitle\b`(git-workflow 1)이며, 그 사실을 여기 기록해 둔다.

실패: 0인 토큰은 해당 개념이 누락된 것이다. 단, 이 검사는 **존재 확인일 뿐 서술 품질을 보증하지 않는다** — 각 토큰이 실제 규칙 행을 갖는지는 사람이 본문을 읽어 확인한다(§E DoD).

### AC-SEO-008 — 휘발성 수치 미탑재

```bash
grep -nE '\b(60|150|160|200|50)[ ]?(characters|chars|자|words|단어)|changefreq|priority.*0\.[0-9]' \
  internal/template/templates/.claude/skills/moai-ref-seo/SKILL.md | wc -l
grep -ncE 'changefreq' ~/.agents/skills/higgsfield-websites/references/seo.md
```

기대: 첫 명령 `0`, 둘째 명령 `1` 이상(패턴 유효성 양성 대조).
실패: 첫 명령이 0이 아니면 해당 행을 결정 규칙 서술로 바꾼다.

### AC-SEO-009 — 플랫폼 종속 마켓플레이스·크레딧 흐름 미포함

Given 원문에 1st-party 마켓플레이스 리스팅 카드 동기화 흐름과 유료 커버 비디오 크레딧 비용 규칙이 있고
When 산출물에서 해당 어휘를 검색하면
Then 매칭이 0이다. 양성 대조로 원문에서 동일 패턴을 검색하면 1 이상이어야 한다.

```bash
P='marketplace|listing card|credit cost|cover video|sidecar'
grep -inE "$P" internal/template/templates/.claude/skills/moai-ref-seo/SKILL.md | wc -l
grep -icE "$P" ~/.agents/skills/higgsfield-websites/references/seo.md
```

기대: 첫 명령 `0`, 둘째 명령 `1` 이상.

대조 baseline (plan-phase 실측): 원문 `7`. 음성 대조로 무관한 기존 ref 스킬 4건(`ui-polish` / `api-patterns` / `git-workflow` / `secops`) 전부 `0` — 패턴이 일반 어휘를 잡지 않음을 확인했다.

실패: 첫 명령이 0이 아니면 플랫폼 종속 흐름이 유입된 것이므로 해당 문단을 제거한다(REQ-SEO-009 위반). 둘째 명령이 0이면 패턴이 잘못된 것이므로 첫 결과는 증거가 아니다.

---

## §D. 배포 가드

### AC-SEO-020 — 템플릿 leak 가드

강제 파일: `internal/template/internal_content_leak_test.go`

```bash
go test ./internal/template/ -run 'TestTemplateNoInternalContentLeak|TestSkillBodyLeakClassRecurrenceBackstop' -count=1 -v 2>&1 | tail -30
```

기대: `--- PASS: TestTemplateNoInternalContentLeak` 및 `--- PASS: TestSkillBodyLeakClassRecurrenceBackstop` 두 행이 **출력에 실제로 등장**하고 `ok`로 끝난다.
실패: 두 PASS 행 중 하나라도 없으면 `-run` 선택자가 아무것도 매칭하지 않은 것이며(exit 0이어도 무효), 테스트가 실행되지 않은 것이다. FAIL이면 위반 파일·클래스가 출력되므로 스킬 본문을 고친다 — **가드 면제 목록에 추가하지 않는다**(REQ-SEO-027).

### AC-SEO-025 — 언어 중립성 가드

강제 파일 (실측 확인):

| 파일 | 테스트 함수 | 무엇을 강제하는가 |
|---|---|---|
| `internal/template/lang_boundary_audit_test.go:190` | `TestLanguageNeutrality` | 언어 우위 정규식 (`langNames` 교대 목록 199행, sentinel `LANG_NEUTRALITY_VIOLATION`) — 실제 언어 중립성 가드 |
| `internal/template/lang_boundary_audit_test.go:132` | `TestSkillBodyNoLangReference` | 스킬 본문의 `moai-lang-*` 토큰 금지 (sentinel `DEAD_LANG_SKILL_REFERENCE`) |
| `internal/template/template_neutrality_audit_test.go:335` | `TestTemplateNeutralityAudit` | 템플릿 중립성 감사 |
| `internal/template/template_neutrality_audit_test.go:399` | `TestTemplateNeutralityAuditC8Preserve` | `GOOS=` 토큰 보유 파일이 정확히 2개임을 단언 |

> **v0.1.0 오지정 정정 (MF-1)**: v0.1.0은 강제 파일을 `internal_content_leak_test.go`로 지목하고 `-run` 선택자를 `TestTemplateNeutralityAudit|TestTemplateNeutralityAuditC8Preserve`로 두었다. 언어 우위 정규식은 그 파일에 없고, 그것을 담은 `TestLanguageNeutrality`는 위 선택자에 매칭되지 않는다. 그대로 실행하면 두 `--- PASS:` 행은 나오지만 **언어 중립성은 한 번도 검사되지 않은 채 통과한다.** plan.md §F가 1순위 위험으로 지목한 항목이 정확히 이것이므로, 이 정정이 없으면 최상위 위험이 무검증 상태로 남는다.

```bash
go test ./internal/template/ \
  -run 'TestLanguageNeutrality|TestSkillBodyNoLangReference|TestTemplateNeutralityAudit|TestTemplateNeutralityAuditC8Preserve' \
  -count=1 -v 2>&1 | grep -E '^(--- PASS|--- FAIL|ok|FAIL)'
grep -rl 'GOOS=' internal/template/templates/ | wc -l
```

기대: 최상위 `--- PASS:` 행 4개가 **모두** 출력에 등장하고 `ok`로 끝난다. `GOOS=` 보유 템플릿 파일 수 `2`.

변경 전 baseline (plan-phase 실측, `TestLanguageNeutrality`는 하위 테스트가 많아 최상위 행만 필터):

```
--- PASS: TestSkillBodyNoLangReference (0.04s)
--- PASS: TestTemplateNeutralityAuditC8Preserve (0.09s)
--- PASS: TestTemplateNeutralityAudit (0.00s)
--- PASS: TestLanguageNeutrality (0.60s)
ok  	github.com/modu-ai/moai-adk/internal/template	1.025s
$ grep -rl 'GOOS=' internal/template/templates/ | wc -l
       2
```

실패: 4개 PASS 행 중 하나라도 없으면 `-run` 선택자가 그 테스트를 매칭하지 않은 것이며(exit 0이어도 무효) 검사가 실행되지 않았다. `TestLanguageNeutrality` FAIL은 본문에 언어 우위 문장이 들어갔다는 뜻이다 — 교대 목록에 맨 토큰 `go`와 `r`이 포함되므로 "Go is the default" 같은 우발적 문장이 걸린다. 프레임워크·언어 지목을 프로토콜·출력 레이어 서술로 되돌린다. `GOOS=` 카운트가 3 이상이면 신규 스킬이 토큰을 추가한 것이므로 제거한다.

### AC-SEO-020b — catalog 정합

강제 파일: `internal/template/catalog_tier_audit_test.go`, `internal/template/catalog_loader_test.go`, `internal/template/embed_catalog_test.go`

```bash
go test ./internal/template/ -run 'TestAllSkillsInCatalog|TestCatalogNoDuplicateEntries|TestManifestHashFormat|TestCatalogTierValid|TestLoadCatalog|TestLoadEmbeddedCatalog_Success|TestEmbeddedMoaiSkillNames' -count=1 -v 2>&1 | tail -40
```

기대: 위 7개 테스트 각각의 `--- PASS:` 행이 출력에 등장.
실패: PASS 행이 누락된 테스트는 실행되지 않은 것이다. `TestManifestHashFormat` FAIL은 해시를 수기로 썼다는 신호이며 `make build` 재실행으로 해소한다.

### AC-SEO-020c — tier 값 유효성과 워크플로 커버리지

강제 파일: `internal/template/catalog_tier_audit_test.go`

Given catalog `tier` 값이 `optional-pack:frontend`로 확정되었고(plan.md §B.1)
When tier 검증기와 워크플로 커버리지 단언을 실행하면
Then 둘 다 통과하고, optional-pack 소속이 별도 선언을 요구하지 않음이 확인된다.

```bash
grep -n 'tier: optional-pack:frontend' internal/template/catalog.yaml | grep -c .
grep -n -A1 'name: moai-ref-seo' internal/template/catalog.yaml
go test ./internal/template/ -run 'TestCatalogTierValid|TestWorkflowTriggerCoverage' -count=1 -v 2>&1 \
  | grep -E '^(--- PASS|--- FAIL|ok|FAIL)'
grep -rn 'required-skills' .claude/skills/moai/workflows/ internal/template/templates/.claude/skills/moai/workflows/ | wc -l
```

기대: 신규 엔트리의 `tier`가 `optional-pack:frontend`, `--- PASS: TestCatalogTierValid` 및 `--- PASS: TestWorkflowTriggerCoverage` 두 행이 출력에 등장, `required-skills` 매칭 수 `0`.

근거 baseline (plan-phase 실측):

```
$ grep -n 'tierPattern' internal/template/catalog_tier_audit_test.go
300:	tierPattern := regexp.MustCompile(`^(core|optional-pack:[a-z][a-z0-9-]{1,30}|harness-generated)$`)

$ grep -rn 'required-skills' .claude/skills/moai/workflows/ internal/template/templates/.claude/skills/moai/workflows/
(출력 없음 — 현재 어떤 워크플로도 required-skills를 선언하지 않는다)
```

`WORKFLOW_UNCOVERED` 단언(같은 파일 509행)은 워크플로 frontmatter의 `metadata.required-skills`에 선언된 스킬만 검사하고, 그 검사의 `knownSkills` 집합은 optional-pack 스킬도 bare name으로 등록한다. 따라서 **optional-pack 소속이 별도 선언 표면을 요구하지 않는다.**

실패: 마지막 명령이 0이 아니면 워크플로가 `required-skills`를 선언하기 시작한 것이므로, `moai-ref-seo`가 그 목록에 필요한지 M4에서 재판단한다(이 AC는 위 전제가 계속 참인지 감시하는 역할이다). `TestCatalogTierValid` FAIL은 tier 문자열 오타다.

### AC-SEO-021 — 카운트 상수 증가와 지상 진실 일치

```bash
grep -n 'expectedSkillCount = ' internal/template/catalog_tier_audit_test.go
grep -n 'expectedTotal = '      internal/template/catalog_loader_test.go
grep -n 'wantTotal = '          internal/template/embed_catalog_test.go
git ls-files 'internal/template/templates/.claude/skills/*/SKILL.md' | wc -l
grep -c '^\s*- name: ' internal/template/catalog.yaml
```

기대: `expectedSkillCount = 32`, `expectedTotal = 42`, `wantTotal = 42`, 디스크 스킬 수 `32`, catalog 엔트리 수 `42`.
변경 전 baseline(plan-phase 실측): 31 / 41 / 41 / 31 / 41.
실패: 상수와 실측이 어긋나면 `TestAllSkillsInCatalog` 또는 `TestLoadCatalog`가 FAIL한다.

### AC-SEO-021b — 마크다운과 Go 상수가 동일 변경에 존재

Given CI path-filter가 마크다운 전용 변경을 docs-only로 분류해 race 테스트 잡을 스킵하고
When 변경 집합을 점검하면
Then `.go` 파일과 `.md` 파일이 함께 포함되어 있다.

**전제 (§A.3)**: 이 항목은 **run-phase feature 브랜치에서 실행한다.** 브랜치를 먼저 확인해 전제를 증명한다.

```bash
git branch --show-current            # main이면 아래 판정은 무효
git rev-list --count origin/main..HEAD   # 0이면 비교 범위가 비어 있어 판정 무효
git diff --name-only origin/main...HEAD | grep -c '\.go$'
git diff --name-only origin/main...HEAD | grep -c '\.md$'
```

기대: 브랜치가 `main`이 아니고, 커밋 수가 1 이상이며, `.go`·`.md` 카운트가 둘 다 1 이상.
실패: `main`에서 origin과 동기(`0 0`)인 상태로 실행하면 두 카운트가 모두 0이 나오는데, 이것은 위반이 아니라 **판정 자체가 성립하지 않은 것**이다. `.go` 카운트가 0인데 브랜치·커밋 전제가 만족된 경우에만 진짜 실패이며, 이때는 가드 잡이 스킵되어 이 SPEC의 모든 §D 항목이 CI에서 실제로 실행되지 않는다.

### AC-SEO-022 — 등록 표면 6종

Given `moai-ref-secops`가 완전한 등록 선례이고 `moai-ref-ui-polish`가 불완전 선례일 때
When 신규 스킬의 등장 파일 집합을 선례와 비교하면
Then 신규 스킬이 secops와 같은 표면 집합을 덮는다.

```bash
for S in moai-ref-secops moai-ref-seo; do
  echo "=== $S ==="
  git grep -ln "$S" -- ':!*/skills/moai-ref-*/SKILL.md' ':!.moai/specs' ':!.moai/reports' | sort
done
```

기대: `moai-ref-seo`의 목록이 아래 **secops 선례 실측 집합 15건**과 동일한 표면을 덮는다.

선례 baseline (plan-phase 실측, 위 명령의 `moai-ref-secops` 절 출력 전문):

```
.claude/skills/moai/workflows/review.md
.claude/skills/moai/workflows/sync/quality-gates-quality.md
.moai/config/sections/delegation.yaml
docs-site/content/en/advanced/skill-guide.md
docs-site/content/ja/advanced/skill-guide.md
docs-site/content/ko/advanced/skill-guide.md
docs-site/content/zh/advanced/skill-guide.md
internal/template/catalog_loader_test.go
internal/template/catalog_tier_audit_test.go
internal/template/catalog.yaml
internal/template/embed_catalog_test.go
internal/template/skills_manifest_test.go
internal/template/templates/.claude/skills/moai/workflows/review.md
internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-quality.md
internal/template/templates/.moai/config/sections/delegation.yaml
```

> **v0.1.0 기대 집합 역전 정정 (MF-2)**: v0.1.0의 기대 목록은 `.claude/rules/moai/workflow/skill-routing.md`를 **포함**하고 워크플로 본문 4건을 **누락**했다 — 실측과 정반대다. `moai-ref-secops`는 skill-routing.md에 등장하지 않는다:
>
> ```
> $ grep -c 'moai-ref-secops' .claude/rules/moai/workflow/skill-routing.md
> 0
> ```
>
> 그 파일의 3개 스킬 예시(`moai-ref-api-patterns` / `moai-ref-owasp-checklist` / `moai-ref-react-patterns`)는 고정 예시이지 등록 표면이 아니다. research.md §D.3의 비교표는 정확했고("워크플로 본문 있음"), 오류는 research → spec/acceptance 이관 과정에서 발생했다.
>
> **skill-routing.md 판정: 표면에서 제외한다.** 예시 추가를 별도로 원한다면 "secops와 동일 표면"이라는 근거는 쓸 수 없으므로 별개 요구사항으로 분리해야 한다. 본 SPEC은 그렇게 하지 않는다.

실패: 누락 파일은 그 표면에서 스킬이 inert하다는 뜻이다. 특히 워크플로 본문 4건을 빠뜨리면 plan.md §G의 안티패턴("등록 표면을 절반만 채우고 완료 선언하기 — `moai-ref-ui-polish` 선례가 그 결과다")을 이 SPEC이 스스로 반복하게 된다.

**`skills_manifest_test.go` 항목의 가치에 대한 정직한 단서**: 이 스팟체크 목록은 특정 회귀를 고정하기 위한 핀이며, 매니페스트는 이제 파생 집합(`EmbeddedMoaiSkillNames()`가 디렉터리를 walk)이다. 신규 스킬을 여기 추가해도 새 가드가 생기지는 않는다. secops 표면 등가를 위해 채우되, 이를 "가드 확보"로 계상하지 않는다.

### AC-SEO-023 — docs-site 4-로케일 동기

```bash
for L in en ja zh ko; do
  printf '%s\t%s\n' "$L" "$(grep -c 'moai-ref-seo' docs-site/content/$L/advanced/skill-guide.md)"
done
```

기대: 4개 로케일 전부 1 이상.
실패: 0인 로케일은 표 행 누락이다(`moai-ref-ui-polish`가 en·ja·zh에서 실제로 이 상태다 — 반복하지 않는다).

### AC-SEO-024 — 전체 템플릿 패키지 회귀

```bash
go test ./internal/template/... -count=1 2>&1 | tail -5
```

기대: `ok  	github.com/modu-ai/moai-adk/internal/template` (FAIL 0).
실패: 개별 테스트명이 출력되므로 §D의 해당 항목으로 되돌아간다.

---

## §D.1 추적성 — REQ 23 → AC 24 전수 매핑

> **v0.2.0 매핑 오기재 정정 (N2)**: v0.2.0의 이 표는 "예외 3건 외 나머지 20건은 동일 번호 AC가 판정한다"고 단언했다. **거짓이다.** §B.2 프로토콜 블록에서 AC 번호가 REQ 번호와 한 칸 어긋나며(010→011, 011→012, 012→013), 기계 판정 AC가 없는 REQ는 3건이 아니라 **4건**(010 / 014 / 026 / 027)이다. v0.2.0은 그중 2건(010 / 014)을 누락했다. 아래 표는 **각 AC 본문이 실제로 무엇을 검사하는가**를 기준으로 다시 작성한 것이며, 번호 규칙에서 연역한 것이 아니다.
>
> **번호 오프셋을 재부여로 해소하지 않는 이유**: §B.2에서 REQ↔AC 관계는 **다대다**다. `REQ-SEO-011`(라이선스 확인 + 해시 고정)은 두 AC가 절반씩 판정하고, `REQ-SEO-012`(중첩 검사)는 `AC-SEO-011`+`AC-SEO-013`이 함께 판정한다. 어떤 번호 부여로도 1:1을 만들 수 없으므로, 재부여는 불일치를 없애지 못하고 위치만 옮긴다. 따라서 **번호는 그대로 두고 이 표가 대응을 명시 관리한다**(spec.md §B 번호 관례 문단과 동일한 결론).

### 전수 매핑

| REQ | 판정 수단 | 대응 형태 |
|---|---|---|
| REQ-SEO-001 단일 SKILL.md, 양 트리 | AC-SEO-001 (+ AC-SEO-011b 동일성) | 동일 번호 + 보강 |
| REQ-SEO-002 H1 1 / 도메인 H2 4-10 / 150-220줄 | AC-SEO-002 | 동일 번호 |
| REQ-SEO-003 불변 3종 종료 + zone id | AC-SEO-003 | 동일 번호 |
| REQ-SEO-004 frontmatter 필드 + 1536자 | AC-SEO-004 | 동일 번호 |
| REQ-SEO-005 description 3악장 + GEO·접근성 제외 명시 | AC-SEO-005 | 동일 번호 |
| REQ-SEO-006 내구성 개념 12종 | AC-SEO-006 | 동일 번호 |
| REQ-SEO-007 무출처 수치 미재현 | **AC-SEO-014** | **번호 불일치** |
| REQ-SEO-008 휘발성 수치 미탑재 | AC-SEO-008 | 동일 번호 |
| REQ-SEO-009 마켓플레이스·크레딧 제외 | AC-SEO-009 | 동일 번호 |
| REQ-SEO-010 프로토콜 재사용 가능 형태 기록 | **기계 판정 AC 없음** → §E 사람 판독 | **공백 (문서화)** |
| REQ-SEO-011 provenance 배포 디렉터리 한정 + sha256 고정 | **AC-SEO-010**(라이선스 절반) + **AC-SEO-012**(digest 절반) | **번호 오프셋 + 다대일** |
| REQ-SEO-012 중첩 검사 명령·정규화·임계값·기대출력 | **AC-SEO-011** (+ **AC-SEO-013** 보강) | **번호 오프셋 + 다대일** |
| REQ-SEO-013 self-trip FAIL 관측 + digest 선단언 | **AC-SEO-012** | **번호 오프셋** |
| REQ-SEO-014 초과 시 재작성(임계값 조정 금지) | **기계 판정 AC 없음** → §E 사람 판독 + AC-SEO-011 통제된 출구 절차 | **공백 (프로세스)** |
| REQ-SEO-015 구조 재사용 허용 + 구조 발산 판정 기록 | AC-SEO-015 | 동일 번호 |
| REQ-SEO-020 catalog 5키 + tier 값 | AC-SEO-020b + AC-SEO-020c | 동일 번호 계열 |
| REQ-SEO-021 Go 상수 3개 증가 + 동일 변경 | AC-SEO-021 + AC-SEO-021b | 동일 번호 계열 |
| REQ-SEO-022 secops와 동일 표면 집합 | AC-SEO-022 | 동일 번호 |
| REQ-SEO-023 docs-site 4로케일 | AC-SEO-023 | 동일 번호 |
| REQ-SEO-024 금지 내부 토큰 미포함 | **AC-SEO-020** | **번호 불일치** |
| REQ-SEO-025 언어 우위 금지 + `GOOS=` 금지 | AC-SEO-025 | 동일 번호 |
| REQ-SEO-026 프레임워크 동일 깊이 나열 | **기계 판정 AC 없음** → §E 사람 판독 | **공백 (판독)** |
| REQ-SEO-027 가드 지적 시 면제 추가 금지 | **기계 판정 AC 없음** → §E 사람 판독 + AC-SEO-020/025 실패 절 | **공백 (프로세스)** |

### 집계

- **기계 판정 AC가 없는 REQ: 4건** — REQ-SEO-010 / 014 / 026 / 027. **4건 전부** §E DoD에 REQ 번호를 명시한 사람 판독 체크박스를 갖는다. 판정 부재가 아니라 **판정 수단이 사람**이다. (v0.1.0 5건 → v0.3.0 4건. 009는 AC-SEO-009 신설로, 015는 AC-SEO-015 신설로 해소되었고, 014가 이 표를 다시 세우는 과정에서 새로 드러났다. v0.2.0 보고가 주장한 "5 → 2"는 §D.1이 2건만 문서화한 데서 나온 착시이며 **거짓**이다.)
- **번호가 어긋나는 대응: 5건** — REQ-007↔AC-014, REQ-011↔AC-010+012, REQ-012↔AC-011+013, REQ-013↔AC-012, REQ-024↔AC-020.
- **존재하지 않는 REQ를 판정하는 AC: 0건.**
- **REQ 없이 존재하는 AC: 1건** — AC-SEO-024(전체 템플릿 패키지 회귀). 특정 REQ가 아닌 일반 회귀 방지이며 정당한 추가다.
- **보강 AC 3건** — AC-SEO-011b(로컬·템플릿 동일성) / AC-SEO-013(LCS) / AC-SEO-020c(tier 유효성)는 각각 REQ-SEO-001 / 012 / 020을 보강한다.

형제 SPEC 2건은 이 표를 **자신의 실제 매핑으로 다시 채운다.** 번호 관례만 복사하고 이 표를 비워두면 v0.2.0과 같은 거짓 1:1 주장이 재생산된다.

---

## §E. Definition of Done

### 기계 판정

- [ ] §B 클린룸 7항목(AC-SEO-010/011/011b/012/013/014/015) 통과 — 특히 AC-SEO-012가 `source_pin_match=True`와 self-trip FAIL 관측을 **둘 다** 보였다
- [ ] §C 콘텐츠 8항목 통과 (AC-SEO-001/002/003/004/005/006/008/009)
- [ ] §D 배포 가드 9항목 통과 (AC-SEO-020/020b/020c/021/021b/022/023/024/025), 각 `--- PASS:` 행을 출력에서 실제로 인용
- [ ] AC-SEO-025에서 **`TestLanguageNeutrality`의 `--- PASS:` 행이 출력에 실제로 등장**했다 — 이 SPEC의 1순위 위험이 검증되었다는 유일한 증거다

### 사람 판독 (기계 판정이 대신할 수 없는 항목)

- [ ] **구조 발산 판정** — AC-SEO-015의 `structural_divergence: PASS` 판정과 2-3문장 근거가 `progress.md` §E.2에 기록되었다 (REQ-SEO-015)
- [ ] **개념 서술 품질** — AC-SEO-006의 12개 토큰 각각이 단순 등장이 아니라 실제 규칙 행을 갖는지 본문 판독으로 확인했다 (REQ-SEO-006)
- [ ] **프레임워크 동일 깊이 나열** — 본문이 특정 프레임워크를 지목한 경우 복수를 동일 깊이로 나열했거나 아예 나열하지 않았고, 서술이 프로토콜·출력 레이어에 고정되어 있다 (REQ-SEO-026 — 판정 AC 없음, 이 체크가 유일한 판정 수단)
- [ ] **프로토콜 재사용 가능성** — 형제 SPEC이 재도출 없이 채택할 수 있는 형태로 spec.md §B.2 + §E 원문 고정 표에 명령·정규화·임계값·기대 출력·self-trip·구조 발산 절차가 전부 적혀 있다 (REQ-SEO-010/015)
- [ ] **임계값 조정 금지 준수** — 중첩 검사 HIT가 있었다면 재작성으로 해소했고 임계값을 올리지 않았다. 기술 용어 연쇄 예외를 쓴 경우 HIT 전문·재작성 시도·사람 승인이 `progress.md` §E.2에 기록되었다 (REQ-SEO-014)
- [ ] **가드 면제 미추가** — 가드가 신규 스킬을 지적했을 때 면제 목록에 추가하지 않고 본문을 고쳤다 (REQ-SEO-027)

### 결정 게이트

- [x] plan.md §B의 열린 결정 5건이 전부 결정 기록으로 치환됨 (v0.2.0 완료 — plan.md에 잔여 clarification 마커 0건)
