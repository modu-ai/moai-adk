# t338 — AC 개수 자가검사 과다 계상: 계획 전 실측

- card: t338
- worktree: `.claude/worktrees/t338` (branch `WT-ac-count-sweep`)
- base: `da03d9188` — reflog 증거 `worktree-t338@{0}: branch: Created from origin/develop`
- 측정 트리: 위 base, 이 카드의 변경 0인 상태

모든 수치는 이 트리에서 이 실행으로 측정했다.

---

## 1. 자가검사가 정의된 곳 (처방이 걸릴 지점)

두 곳에 같은 명령이 있다.

- `.claude/agents/moai/manager-docs.md:81` — `### B12 CHANGELOG emission discipline (mandatory self-test before commit)`, 그 안 `:89` 에 스윕 명령
- `.claude/rules/moai/development/manager-develop-prompt-template.md:127` — `**B12. Sync-phase CHANGELOG emission discipline (manager-docs only)**`

스윕 명령 원문(`manager-docs.md:89`):

```bash
grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' .moai/specs/<SPEC-ID>/acceptance.md | sort -u | wc -l
```

**이미 한 종류의 공허성은 다루고 있다.** 같은 절이 "**A count of 0 is a RED flag, not a pass.** `0 == 0` is a vacuous comparison" 이라고 적는다. 즉 저작자는 *과소* 방향(아무것도 안 세는 선택자)을 의식했고, *과다* 방향은 의식하지 않았다. t241 이 다루는 것과 이 카드가 다루는 것이 정확히 그 두 방향이다.

주석도 형태 다양성을 인정한다 — AC 는 `### AC-RIL2-001 — …` 헤딩으로도, `| AC-HYG-001-001 |` 표 셀로도, `| AC-DFF-01 |` 두 자리 행으로도 나타난다. 스윕이 마크업이 아니라 토큰에 앵커한 이유이고, **동시에 폐기 각주 안의 토큰과 유효 기준을 구분할 수 없게 된 이유**이기도 하다.

## 2. 영향권의 크기

```console
$ ls -d .moai/specs/*/ | wc -l
     690
$ ls .moai/specs/*/acceptance.md | wc -l
     601
$ grep -rl "폐기\|retired\|RETIRED\|superseded\|SUPERSEDED" .moai/specs/*/acceptance.md | wc -l
     105
```

601개 `acceptance.md` 중 105건(17%)이 폐기 계열 표지를 단다. 다만 이건 느슨한 매칭이라 상한이지 실제 과다 계상 건수가 아니다.

## 3. 실제 과다 계상 — 검출기로 실측

검출기: `.moai/reports/t338/overcount-detector.sh` · 전체 출력: `.moai/reports/t338/overcount-scan.txt`

판정 규칙 — 어떤 AC-ID 의 **모든** 등장 행이 폐기 표지를 달고 있으면, 그 ID 는 스윕이 세지만 유효 기준이 아니다.

```console
$ bash .moai/reports/t338/overcount-detector.sh
---
acceptance.md scanned : 601
files over-counted    : 18
phantom AC ids        : 29
```

**601건 중 18건이 과다 계상되고, 유령 ID 는 29개다.** 그리고 이 18건 전부에서 자가검사는 지금까지 통과해 왔다 — 과다 계상은 실패를 내지 않고 그저 틀린 수를 낸다.

이 카드를 낳은 SPEC 자신이 목록 첫 줄에 있다:

```
OVERCOUNT .moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/acceptance.md  AC-AEL-008
```

근거 행(`acceptance.md:165`):

```
> **폐기 — 종전 AC-AEL-008 (update 분기 되읽기).** REQ-AEL-008 과 함께 폐기했다. … 수락 기준은 001-007 총 7건이다.
```

스윕은 이 파일에서 `AC-AEL-001` … `AC-AEL-008` 여덟 개를 뽑는다. 유효 기준은 일곱이다.

### 이 수치는 하한이다

검출기는 "해당 ID 의 모든 등장 행에 표지가 있을 것"을 요구한다. 폐기 기록이 여러 행에 걸쳐 있고 그중 한 행에 표지가 없으면 놓친다. 따라서 18/29 는 **하한**이며 참값은 이보다 크다. 정확한 참값을 내는 것이 이 카드가 정할 판별 방법의 몫이다.

### 측정 중 저지른 실수 하나 (기록해 둔다)

첫 실행에서 출력을 `| tail -30` 으로 잘라 읽었고, 알파벳 순으로 앞에 있던 `SPEC-AGENT-EMIT-LINEAGE-001` 행이 잘려 나갔다. 그 결과 "이 카드를 낳은 SPEC 이 검출기에 안 잡힌다 — 검출기의 한계다"라고 잠시 오판했다. `head` 로 다시 읽어 정정했다. **큐든 스캔 결과든 잘라 읽으면 없는 결론이 생긴다.**

## 4. 이 카드가 정해야 할 것 (실측이 좁혀 준 범위)

카드 본문이 "문면 규약인지 파서인지"를 묻는다. 실측이 그 선택에 두 가지 제약을 준다.

1. **마크업 형태에 앵커할 수 없다.** 헤딩·표 셀·두 자리 행이 공존한다(스윕 주석이 스스로 인정). 헤딩만 세는 계수기는 표 기반 SPEC 에서 과소 계상한다 — 지금 문제의 반대 방향으로 틀린다.
2. **행 단위 표지 매칭으로는 부족하다.** §3 의 검출기가 바로 그 방식이고, 스스로 하한임을 인정한다.

## 5. 완료 조건 (카드가 못 박은 것)

새 계수기는 **폐기 기준이 있는 SPEC 과 없는 SPEC 양쪽에서 돌려 두 값이 갈리는 것을 관측하기 전까지 미완성**이다. 폐기 이력이 없는 SPEC 에서만 맞는 계수기는 현행과 구별되지 않는다.

첫 번째 시험대는 `SPEC-AGENT-EMIT-LINEAGE-001`(스윕 8 / 참 7)이고, 대조군은 §3 스캔에서 과다 계상되지 않은 583건 중 아무거나면 된다.

## Gaps — 관측하지 않은 것

- 참값(유효 기준 수)을 SPEC 별로 독립 산출하지 않았다. 18건은 검출기가 잡은 것이지 전수 대조가 아니다.
- 스윕이 **과소** 계상하는 경우가 있는지는 재지 않았다(이 카드 범위 밖이나, 새 계수기는 양방향을 다 봐야 한다).
- 두 정의 지점(`manager-docs.md`, `manager-develop-prompt-template.md`) 사이에 문면 차이가 있는지 대조하지 않았다.
- `spec.md` 쪽의 REQ 개수 스윕에도 같은 형태가 있는지 보지 않았다.

## Residual risk

- 검출기의 표지 목록(`폐기|retired|RETIRED|superseded|SUPERSEDED|철회|withdrawn`)은 관측된 표현에서 낸 것이지 전수 어휘 조사가 아니다. 다른 표현으로 폐기를 적은 SPEC 은 검출기가 놓친다.

---

## 6. §3 정정 — 내 수치는 하한이 아니라 혼합이었다

SPEC 저작자가 §3 을 반박했고, 재현해 보니 맞다. **정정한다.**

§3 에서 나는 18/29 를 "하한"이라고 적었다. 검출기가 놓치는 경우(여러 행에 걸친 폐기 기록)만 생각했기 때문이다. 그런데 검출기는 **과다 검출도 한다** — 폐기를 *주제로 다루는* SPEC 의 살아 있는 기준을, 그 행에 "retired" 같은 낱말이 있다는 이유로 유령으로 분류한다.

```console
$ grep -nE "AC-RA-02([^0-9-]|$)" .moai/specs/SPEC-V3R3-RETIRED-AGENT-001/acceptance.md
46:### AC-RA-02: manager-tdd.md retired stub has all 5 standardized fields (REQ-RA-002)
```

이건 폐기 각주가 아니라 **살아 있는 기준의 헤딩**이고, 제목에 "retired stub" 이 들어 있을 뿐이다. 같은 형태가 7건(`AC-RA-02/07/11/14/17`, `AC-CMR-002`, `AC-LSPMCP-RETIRE-007`), 파일 단위로는 3건이 통째 오탐이다.

**따라서 폐기 축의 실제 과다 계상은 15개 파일 / 22개 식별자**이며, 여기에 미검출분이 더해진다. 내 검출기는 **양방향으로 틀린다.**

이 정정이 설계 결정을 좌우한다: **자연어 어휘를 읽는 판별기는 양방향으로 불건전하다.** 낱말은 그 식별자의 생사를 말해 주지 않는다 — 문장이 폐기를 *말하는지*와 그 기준이 *폐기됐는지*는 다른 문제다. 판별기가 산문이 우연히 만들어낼 수 없는 예약 토큰에 걸려야 하는 이유가 여기 있다.

## 7. 카드가 이름 붙이지 않은 두 번째 축 — 그리고 이쪽이 더 크다

저작자가 새로 잰 것이고, 재현했다.

```console
$ grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/acceptance.md | sort -u | wc -l
      18
$ grep -cE '^#{2,4} AC-ACD-' .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/acceptance.md
       8
```

**방금 만들어진 SPEC 에서 스윕이 18, 실제 기준은 8이다.** 나머지 10개는 폐기와 무관한 **인용된 외부 식별자**다 — 다른 SPEC 의 기준을 근거로 언급한 것.

유병률 상한(`.moai/reports/t338/multidomain-scan.sh`, 이 트리에서 재현):

```console
acceptance.md scanned      : 602
files with >=2 AC prefixes : 156
```

156은 상한이다 — `AC-APO-*` 와 `AC-DCP-*` 를 함께 쓰는 정당한 다도메인 SPEC 도 같은 모양으로 읽힌다. 그래도 폐기 축의 15와 견주면 **한 자릿수가 다르다.**

두 축은 같은 형태다 — "이 파일의 살아 있는 기준이 아닌 식별자 등장". 그래서 판별 규약은 예약 토큰 *집합*(`[RETIRED]` / `[REF]`)으로 자연히 확장되고, 계수 규칙 자체는 그대로다.

(602 대 §3 의 601: 이 SPEC 의 `acceptance.md` 가 트리에 착지한 뒤 스캔했기 때문이다. 다른 파일 변동 없음.)
