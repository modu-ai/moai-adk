# t550 plan-phase 감사 — SPEC-GATE-OXLINT-DETECT-001

감사자: plan-auditor (반복 1회차) · 감사 트리: `.claude/worktrees/t550`, HEAD `d060e0d13`
대상: `.moai/specs/SPEC-GATE-OXLINT-DETECT-001/` (spec.md · plan.md · acceptance.md · progress.md)
Tier: M (PASS 임계 0.80, `spec-workflow.md` § SPEC Complexity Tier)

**최종 판정(3회차): PASS** — blocking 등급 결함 0건. **F1 · F4 · F5 전부 실측으로 닫혔다.**
잔존은 advisory 3건(F2 · F3 · F6)이며 어떤 의무도 깎지 않는다. 상세는 §R6.

감사 대상 리비전(판단 이전에 직접 측정, 14:14:15):
`acceptance.md` 14:12:59 (size 12074) · `spec.md` 14:13:22 (11077) · `plan.md` 14:06:25 (7581) ·
`progress.md` 13:53:59 (987). 이 판정은 이 사본에 대한 것이다.

> 이력: 1회차 FAIL(F1 — REQ-002 부정 절 무대응) → 2회차 FAIL(F4 — F1 수정이 남긴 probe 개수 불일치)
> → 3회차 PASS. 점수 0.875 → 0.9625 → 0.9875, 회귀 없음(반복 3/3, STOP 조건 미해당).

M1 문맥 격리: 작성자의 추론 맥락은 전달받지 않았고 참조하지도 않았다. 판단 근거는 SPEC 아티팩트 4종,
`.moai/reports/t550/baseline.md`, 그리고 이 감사에서 직접 실행한 명령의 출력뿐이다.

### 감사한 리비전 (E0) — **[1회차 한정 · §R0 이 대체함]**

> **[SUPERSEDED by §R0]** 이 절의 mtime 13:55 과 "아래 모든 판정은 이 리비전에 대한 것"이라는 문장은
> **1회차 감사에만** 적용된다. F1 수정이 14:06 에 `acceptance.md`·`plan.md` 에 들어온 뒤의 판정은 §R0 이
> 별도 mtime 과 함께 기록한다. 이 절만 읽고 보고서 전체의 리비전을 판단하면 안 된다 — 그 오독 가능성은
> §R0 을 덧붙일 때 이 절의 범위 문장을 묶어 두지 않은 내 결함이며, 이 표지가 그 수정이다.

리드로부터 "스폰 이후 아티팩트가 바뀌었다"는 통지를 받고 **가정하지 않고 재측정했다.**

```
$ stat -f '%Sm %N' -t '%Y-%m-%d %H:%M:%S' spec.md plan.md acceptance.md progress.md
2026-09-10 13:55:05 spec.md
2026-09-10 13:55:21 plan.md
2026-09-10 13:55:14 acceptance.md
2026-09-10 13:53:59 progress.md
```

이 mtime 은 감사 착수 시점의 `ls -la` 가 보인 값과 **동일**하다(spec/plan/acceptance 13:55, progress 13:53).
즉 파일은 내가 읽은 뒤로 움직이지 않았다 — **나는 개정본을 읽었다.** 리드가 지목한 네 변경이 내가 읽은
사본에 모두 들어 있음을 항목별로 확인했다:

```
$ /usr/bin/grep -c "^| \(eslintprj\|biomeprj\|oxlintprj\|ox2\|ox3\|ox4\|bareprj\)" spec.md
7
$ /usr/bin/grep -n "considered and rejected" spec.md
164:  **considered and rejected by the lead** (§3), not left pending, and is therefore out of
$ /usr/bin/grep -n -i "open decision\|still open\|pending decision\|awaits" plan.md spec.md acceptance.md
spec.md:25:  config-file gating only; the former "Open decision" section became §3 Decision. (b) The
spec.md:165:  scope for this SPEC. Nothing here awaits a decision. Should the rejection ever be
```

- §1 매트릭스 7행 — 확인(4행 아님).
- §5 devDependencies 축 "considered and rejected" — 확인(pending 아님).
- AC-003 의 per-filename RED 근거 + ox3/ox4 이중 임무 — 확인(§D.3, E5-a 에 인용).
- plan.md §D M2 의 잔존 "open decision" 문구 — plan.md 적중 **0건**. spec.md 의 두 적중은 개정 **후** 형태다:
  25행은 HISTORY 가 전환 자체를 기록한 것이고, 165행은 `Nothing here awaits a decision` — 종결 선언이지
  미결 표시가 아니다.

따라서 **1회차 판정**은 이 리비전(mtime 13:55)에 대한 것이며, 재작업이 필요한 축은 없었다. 다만 리드가
지목한 축 3의 접두사 충돌 항목은 아래 E5-a 에서 **새 AC 를 명시적으로 반영해 다시 적었다**.

이 시점의 사본에는 `set-EQUAL` 이 아직 없었다 — F1 수정은 14:06 에 들어온다(§R0). 당시 "F1 은 개정본에
대해서도 그대로 선다"는 진술은 **그 시점의 현행 사본**에 대한 것으로 참이었고, §R1 이 그것을 대체한다.

---

## §R 재감사 (2회차) — F1 축 한정

리드 지시대로 **F1 축만** 재감사했다. 1회차에 통과시킨 축 2·3·4·5 는 재실행하지 않았다 — 수정이 그 축들을
건드리지 않았음을 아래 R0 의 변경 범위로 확인했다.

### R6. 3회차 재감사 — F4 계열 + F5, 그리고 최종 판정 **PASS**

#### R6.0 리비전 (판단 이전에 먼저 측정)

```
$ stat -f '%N mtime=%Sm size=%z' -t '%Y-%m-%d %H:%M:%S' acceptance.md spec.md plan.md progress.md
acceptance.md mtime=2026-09-10 14:12:59 size=12074
spec.md       mtime=2026-09-10 14:13:22 size=11077
plan.md       mtime=2026-09-10 14:06:25 size=7581
progress.md   mtime=2026-09-10 13:53:59 size=987
$ date '+%Y-%m-%d %H:%M:%S'
2026-09-10 14:14:15
```

리드의 측정값과 **일치**한다(불일치 없음). 이번 회차에 움직인 것은 `acceptance.md`(F4 계열)와
`spec.md`(F5 가드, 9901 → 11077, +1176B)이며 `plan.md`·`progress.md` 는 불변이다. 아래 판정은 이 사본에
대한 것이다.

#### R6.1 F4 계열 — 문구가 아니라 **요구(개수)** 로 훑었다 → **닫힘**

`Both` 라는 문구가 아니라 "이 문서가 세는 모든 것"을 축으로 잡고 전 아티팩트를 훑었다: probe 수 · AC 수 ·
fixture 수 · 통지 수 · lint 항목 수. 수사(numeral)와 수량 표현을 한 번에 긁었다:

```
$ /usr/bin/grep -n -i -E "\b(both|one|two|three|four|five|six|seven|eight|nine|ten|single|pair|all [a-z]+ )\b" spec.md plan.md acceptance.md progress.md
   (60행 적중 — 전수 분류 결과는 아래 표)
```

| 세는 대상 | 정답 | 적중 위치 | 판정 |
|---|---|---|---|
| probe 수 | 3 | acceptance:19(AC-008 행)·167·174, plan:81, spec:34 | **전부 3** ✓ |
| AC 수 | 8 | acceptance:172 `All eight ACs`, 매트릭스 행 8개(`grep -c "^| AC-00"` → 8) | ✓ |
| oxlint 파일명 수 | 4 | acceptance:14·15·52·57·59·61·63·69·82·83(`len(configFiles) == 4`)·157, plan:30·39·58·82·108 | **전부 4** ✓ |
| 변경 후 통지 수 | 3 | acceptance:18·127, plan:77 | ✓ |
| baseline 통지/항목 수 | 2 | acceptance:128 `up from the two`, :129 `both entries skipped`, spec:45 `exactly two entries`, plan:9 | **baseline 서술로서 정확** ✓ |
| baseline fixture 수 | 7 | spec:50 `seven fixtures` (매트릭스 7행) | ✓ |

`2` 를 말하는 자리들은 전부 **변경 전 트리를 서술하는 baseline 문장**이며 스테일이 아니다 — 이 구분을
하지 않으면 정상 문장을 결함으로 오인한다.

```
$ /usr/bin/grep -n "Both mutant" acceptance.md plan.md spec.md; echo "(exit=$?)"
(exit=1)
```

**"제3의 인스턴스가 살아 있는가" — 없다.** 문구가 아니라 대상으로 훑어 probe 를 언급하는 **모든 줄**을
전수 확인했다(15행):

```
$ /usr/bin/grep -n -i "mutant\|probe" spec.md plan.md acceptance.md progress.md
spec.md:33·34            (HISTORY — 수정 사실 서술)
plan.md:41·81·107·108    (81=all three ✓, 107·108=M1/M2 개별 지칭, 41=M3 지칭)
acceptance.md:19         all three ✓   ← 형제, 이번에 수정됨
acceptance.md:92·146·151·154·157  (M3 지칭 / 절 제목 / M1·M2·M3 각 항목)
acceptance.md:167        All three ✓
acceptance.md:174        All three (M1, M2, M3) ✓
acceptance.md:148        총칭 단수 + AC 범위 절 → F6(별개 계열, optional)
```

probe 개수를 **주장하는** 자리는 5곳(acceptance:19·167·174, plan:81, spec:34)이고 **전부 3**이다.
나머지는 개별 mutant 지칭이거나 절 제목이라 개수 주장이 아니다. 스테일했던 것은 정확히 둘
(§D.10 = 내가 이름 붙인 것, AC-008 행 = 형제)이었고 **둘 다 닫혔다. 셋째는 없다.**

§D.10 은 `All three mutant probes of §D.9 (M1, M2, M3)` 로 바뀌었고, **M3 비선택 조항**이 붙었다 —
`M3 is not optional: dropping it closes the card on an asserted §D.4 fix rather than a demonstrated one`.
내가 F4 에서 지적한 "종결 게이트만 2를 유지" 경로가 사라졌다. **F4 닫힘.**

#### R6.2 내 F4 지적의 결함 — 형제를 찾고도 등급을 낮췄다

리드 지적대로 F4 에는 형제가 있었다: §D 매트릭스 AC-008 행의 `the mutant probe of §D.9`(단수) — 같은
pre-M3 개수다. 나는 그것을 **찾았지만** F4 의 Required fix 에서 `필수는 아니다` 로 분류했다("열거 목록이
뒤따르므로 오독 위험이 낮아"). 그 판단이 틀렸다: AC-008 행은 **AC 매트릭스**, 즉 독자가 AC 단위로 조회할 때
가장 먼저 닿는 자리이고, 뒤따르는 열거는 §D.9 에 있지 그 행에 있지 않다.

교훈은 개수 하나가 아니라 **범주**다 — 개수 수정을 "내가 이름 붙인 인스턴스"에서 멈추면 계열이 살아남는다.
1회차에 나는 이미 "문구가 아니라 요구로 훑어야 한다"를 F1 에 적용했으면서, F4 에서는 인스턴스 단위로
멈췄다. 이번 회차에 위 표처럼 **세는 대상 전체**를 축으로 다시 훑은 것이 그 교정이다.

#### R6.3 잔존 단수 표현 2건 — 개수가 아니라 총칭이다 (의무 손실 없음)

```
$ /usr/bin/grep -n -i "a mutant\|the mutant probe\|probe of §D.9" acceptance.md plan.md spec.md
acceptance.md:148:Before adopting AC-001/AC-003, the run phase writes a mutant that would satisfy a shallower
acceptance.md:167:All three probes are run and their observed output recorded; a mutant that passes means the
```

둘 다 **총칭 단수**다(기법을 서술하는 "어떤 mutant 하나"), 개수 주장이 아니다. 148행 뒤에는 M1·M2·M3
열거가 곧바로 오고, 167행은 "통과하는 mutant 가 하나라도 있으면"의 뜻이다. 의무를 깎는 자리가 없다.

다만 148행의 **범위 절**은 pre-M3 다: `Before adopting AC-001/AC-003` — M3 이 지키는 것은 AC-004 인데 그
열거에 AC-004 가 없다. → **F6 (optional)**. 의무는 §D.10 과 각 mutant 자신의 문장이 지므로
(§D.10: `the ACs they are supposed to fail`, M3: `must fail AC-004`) 아무것도 탈락하지 않는다.

#### R6.4 F5 가드의 위치 — **독자가 실제로 마주치는 자리다** → 닫힘

리드의 질문은 "건너뛸 절에 있는가"였다. 아니다. 가드는 REQ-002 **안에**, 네 파일명 목록 **바로 아래**
blockquote 로 붙어 있다:

```
$ sed -n 86,99p spec.md
- **REQ-002 (When):** ... The recognized filenames are exactly those oxlint itself discovers, per
  <https://oxc.rs/docs/guide/usage/linter/config.html> (accessed 2026-09-10):
  `.oxlintrc.json`, `.oxlintrc.jsonc`, `oxlint.config.ts`, `oxlint.config.mts`. No other
  filename is recognized.

  > **Reference-set guard (documentation, not a criterion).** Set equality measures whether
  > the implementation matches these four names; it cannot measure whether the four names are
  > still right, and that rests entirely on the citation above as read on its access date.
  > Equality sharpens the stake rather than softening it: should oxlint upstream add a fifth
  > discovery filename, AC-004 will actively **reject** it. If that happens, it is a REQ-002
  > amendment — update this list and the citation's access date — and never a test failure to
  > route around by loosening AC-004.
```

목록을 고치려는 독자는 목록을 읽어야 하고, 가드는 그 목록과 **같은 요구 안에서 바로 다음 줄**이다. 얻을 수
있는 최선의 위치다. 내용도 셋을 다 갖췄다: (a) 접근 날짜 `(accessed 2026-09-10)`, (b) 동등 검사가 이 위험을
**키운다**는 성질의 명시, (c) 상류 변경 시의 처분 규칙 — `never a test failure to route around by loosening
AC-004`. 내가 F5 에서 요구한 "미검증 가정이 아니라 기록된 판정으로 만들 것"이 그대로 충족됐다. **F5 닫힘.**

부수 관측(결함 아님): `plan.md:30` 의 §B.2 는 같은 URL 을 접근 날짜 없이 인용한다. 규범 목록의 SSOT 는
REQ-002 이고 §B.2 는 근거 반향이므로 의무 충돌은 없다.

#### R6.5 3회차 필수통과 재확인 (직접 측정)

```
$ moai spec lint SPEC-GATE-OXLINT-DETECT-001 >/dev/null 2>&1; echo "target exit=$?"
target exit=0
$ moai spec lint SPEC-NOPE-999 >/dev/null 2>&1; echo "control exit=$?"
control exit=3
$ /usr/bin/grep -c "^| AC-00" acceptance.md
8
$ /usr/bin/grep -n "^version:\|^updated:\|^status:" spec.md
4:version: "0.1.0"
5:status: draft
7:updated: 2026-09-10
```

`spec.md` 가 이번에 개정됐으므로 frontmatter 를 재확인했다 — `updated: 2026-09-10` 로 개정일과 일치하고,
`status: draft` 는 plan-phase 로서 정합하다. 필수통과 7항 불변.

#### R6.6 최종 판정

**blocking 등급 결함 0건 → PASS.** 남은 것은 전부 advisory(optional)다: **F2**(린터 병존 조합 미규정),
**F3**(템플릿 `javascript.md` 린터 목록), **F6**(§D.9 도입부 범위 절이 AC-004 를 안 셈). 셋 다 어떤 의무도
깎지 않으며, M6 에 따라 optional 결함 목록의 길이는 그 자체로 FAIL 근거가 되지 않는다.

**F1 · F4 · F5 는 전부 실측으로 닫혔다.** 이 SPEC 은 run-phase 진입 가능하다.

---

### R0'. 리비전 귀속에 대한 정정 요구와 그 처리 (2→3회차 통지)

리드가 "네 개정 판정은 pre-fix 사본에 대한 것"이라며 재측정을 요구했다. **가정하지 않고 다시 쟀고, 그
전제는 성립하지 않는다.** 근거를 그대로 남긴다:

```
$ stat -f '%N mtime=%Sm size=%z' -t '%Y-%m-%d %H:%M:%S' acceptance.md plan.md spec.md progress.md
acceptance.md mtime=2026-09-10 14:06:09 size=11940
plan.md       mtime=2026-09-10 14:06:25 size=7581
spec.md       mtime=2026-09-10 13:55:05 size=9901
progress.md   mtime=2026-09-10 13:53:59 size=987
$ date '+%Y-%m-%d %H:%M:%S'
2026-09-10 14:12:34
$ /usr/bin/grep -c "set-EQUAL" acceptance.md
1
```

리드가 잰 값과 **바이트 단위로 같다**(size 포함). 네 번째 리비전은 없다.

그리고 이 보고서는 **이미 그 사본을 인용하고 있었다**:

```
$ /usr/bin/grep -n "14:06:09\|14:06:25\|set-EQUAL\|F1 닫힘\|CLOSED" plan-audit.md
63:2026-09-10 14:06:25 plan.md
64:2026-09-10 14:06:09 acceptance.md
77:  `optional` is true; and its `configFiles` is **set-EQUAL** to exactly the four names of
90:1회차 F1 이 지적한 ... 경로가 사라졌다. **F1 닫힘.**
474:**F1. [해소됨 — 2회차 검증] ... → CLOSED**
```

즉 §R0 은 post-fix mtime 을 직접 재서 인용했고, §R1 은 `set-EQUAL` 절을 **원문 그대로 인용**했으며,
결함 목록은 F1 을 `CLOSED` 로 바꿔 놓았다. 개정 판정은 post-fix 사본에 대한 것이다.

**다만 리드의 우려에는 참인 알맹이가 있고, 그것은 내 결함이다.** 13:55 을 말하는 E0 절과 14:06 을 말하는
§R0 절이 한 문서 안에 나란히 있는데, E0 의 "아래 **모든** 판정은 이 리비전에 대한 것"이라는 문장을 §R0 을
덧붙일 때 **1회차로 묶어 두지 않았다.** E0 만 읽은 독자는 보고서 전체가 pre-fix 사본에 대한 것이라고 읽게
된다 — 리드가 정확히 그렇게 읽었다. 이 카드가 다루는 "주장이 측정을 앞지른다"의 문서판이며, E0 에
`[SUPERSEDED by §R0]` 표지를 붙여 수정했다.

#### 정정 — 내가 위에서 편 변호는 틀렸다 (리드 지적 수용)

나는 2회차의 "F1 은 개정본에 대해서도 그대로 선다"를 **"측정 시점에는 참이었다"** 로 변호했다. 그 변호는
성립하지 않는다. 실제 시각을 놓고 보면:

| 시각 | 사건 |
|---|---|
| ~14:05 | 내가 `stat` 실행 → 13:55 관측 (당시 디스크와 일치) |
| **14:06:09** | **F1 수정이 `acceptance.md` 에 착지** |
| 14:07:11 | 내가 "F1 은 개정본에 대해서도 그대로 선다"를 **발신** |

측정과 발신 사이 약 2분, 그리고 그 사이에 디스크가 움직였다. **주장의 유효성은 측정 시각이 아니라 발신
시각에 판정된다** — 발신 순간 디스크는 이미 그 주장을 지지하지 않았다. "측정할 때는 맞았다"는 변호는
이 판정 기준을 측정 시각으로 몰래 옮기는 것이고, 그렇게 하면 어떤 낡은 주장도 방어된다.

이것이 이 카드가 다루는 주제 자체다. HEAD 판독이 낡듯 **파일 판독도 측정과 단언 사이에서 낡는다.** 다른
행위자가 같은 트리를 쓰는 동안에는 특히 그렇다. 올바른 규율은 "측정할 때 맞았는가"가 아니라
**"단언 직전에 다시 쟀는가"** 이며, 나는 재지 않았다.

따라서 이 절의 결함 목록을 정정한다 — 결함은 둘이다:
1. **E0 의 범위 문장을 §R0 추가 시점에 닫지 않은 것**(문서 위생) — 위에서 수정함.
2. **측정과 단언 사이의 재측정 누락**(판정 위생) — 리드 지적이 옳고, 내 최초 변호가 틀렸다.

리드가 이번 3회차 지시에서 "판단 이전에 먼저 `stat` 을 돌리고 관측값을 진술하라"고 못박은 이유가 이것이며,
§R6.0 은 그 순서를 지켰다.

리드가 유지하라고 한 ox3/ox4 구분(eslint 항목 대 oxlint 항목, 방향도 대상도 다름)은 E5-a 의 [HARD] 단락과
F1 이력 항목에 그대로 보존돼 있다.

### R0. 감사한 리비전 (직접 관측)

```
$ stat -f '%Sm %N' -t '%Y-%m-%d %H:%M:%S' spec.md plan.md acceptance.md progress.md
2026-09-10 13:55:05 spec.md
2026-09-10 14:06:25 plan.md
2026-09-10 14:06:09 acceptance.md
2026-09-10 13:53:59 progress.md
```

`acceptance.md` 14:06:09, `plan.md` 14:06:25 — 1회차 감사(13:55) 이후 움직인 것은 이 둘뿐이다. `spec.md` 와
`progress.md` 는 13:55:05 / 13:53:59 로 **불변**이다. 즉 수정은 acceptance/plan 두 파일에 국한됐고,
spec.md 를 근거로 삼았던 축 4(범위 정직성)와 코드를 근거로 삼았던 축 2·3·5 는 영향권 밖이다.

### R1. 새 §D.4 는 성장 방향을 실제로 닫는가 — **닫는다** (선언이 아니다)

```
$ sed -n '/## §D.4/,/## §D.5/p' acceptance.md
- **Then** it exists; its `binary` is `npx`; `strings.Join(args, " ")` is `oxlint`;
  `optional` is true; and its `configFiles` is **set-EQUAL** to exactly the four names of
  §D.3 — `len(configFiles) == 4` **and** equal membership in **both** directions. The
  assertion reports each direction separately and distinguishably: every name of §D.3 absent
  from `configFiles` is reported as `missing config file name: <name>`, and every entry of
  `configFiles` outside §D.3 is reported as `unexpected config file name: <name>`. A missing
  name and an unexpected name therefore produce different failures, never one merged count.
```

"닫혔다고 주장"과 "닫는 기계"를 가르는 것은 **실패가 관측 가능한 형태로 특정됐는가**이다. 여기엔 셋이 다
있다: (a) 기수 조건 `len == 4`, (b) 양방향 소속 동등, (c) 두 방향이 **구별되는 메시지**로 보고됨. 특히 (c)
가 없으면 실패 1건이 어느 방향인지 알 수 없어 M3 이 판별력을 잃는다 — 그 요구가 명시돼 있다.

`configFiles` 에 `oxlint.config.js` 를 더한 구현은 (a) 에서 5 ≠ 4, (b) 의 unexpected 쪽에서 각각 걸린다.
1회차 F1 이 지적한 "포함 검사는 목록이 길어지는 방향을 못 잡는다"는 경로가 사라졌다. **F1 닫힘.**

`§D.4` 의 Boundary 절도 이에 맞춰 확장됐다 — `a superset declaration passes every behavioural row while
failing only this one`. 선언층/행동층 분업이 유지된 채 성장 방향만 추가됐다.

### R2. M3 은 판별하는 probe 인가 — **그렇다** (열거로 확인, 가정 아님)

M3 의 주장은 "모든 행동 AC 는 초록으로 남고 AC-004 만 red 가 된다"이다. 이 주장은 **어떤 fixture 도
`oxlint.config.js` 를 담지 않는다**에 전적으로 의존하므로, SPEC 의 fixture 전수를 열거해 확인했다:

```
$ /usr/bin/grep -n -A2 "^- \*\*Given\*\*" acceptance.md | /usr/bin/grep -i "oxlint.config.js"
(no output)
$ echo "exit=1"   # 적중 0건
```

Given 절 전수(§D.1 `.oxlintrc.json` / §D.2 동일 / §D.3 네 이름 각 1개 / §D.5 `eslint.config.js` /
§D.6 `biome.json` / §D.7 설정 없음) 중 `oxlint.config.js` 를 담은 것은 없다. 따라서 목록에 그 이름을 더해도
`anyConfigFileExists` 의 판정이 바뀌는 fixture 가 **0개**이고, 행동층은 전부 불변이다.

유일한 근접 사례는 §D.5 의 `eslint.config.js` 인데, `oxlint.config.js` 와는 별개 파일명이고 판정은
`filepath.Join` + stat 정확 일치(E3)이므로 우연 적중이 없다. 두 이름이 모두 `.config.js` 로 끝나는 탓에
눈으로는 겹쳐 보이지만 기계에는 겹치지 않는다 — 이 지점을 확인하지 않았다면 M3 의 판별력 주장은 미검증으로
남았을 것이다.

→ M3 은 **한 축만 흔드는** probe 다. 여러 AC 를 동시에 넘어뜨려 어느 것이 원인인지 말해주지 않는 종류가
아니다. §D.9 가 그 성질을 자기 문장으로도 명시한다("If it does not fail AC-004, the equality check was not
actually written as equality").

### R3. plan.md 정합 — 잔존 문구는 리드 말대로 **부정 참조**다

```
$ /usr/bin/grep -n -i "membership\|containment" plan.md
plan.md:37:recognized", so the check is set **equality**, never containment: AC-003 exercises each name
plan.md:40:Containment would have constrained only the shrinking direction, leaving a fifth name free
plan.md:80:   distinguishable messages — `acceptance.md §D.4`), not a per-name membership check.
```

세 적중 전부 "containment 가 **아니다**"를 말하는 부정 참조이며, 검사를 포함으로 서술한 절은 없다.
80행은 `not a per-name membership check` — 리드가 지목한 그대로 읽힌다(가정하지 않고 확인함).
§D M1 step 6 도 `Run all three mutant probes` 로 갱신됐고 M3 의 판별 성질을 그 자리에 적었다.

### R4. 어느 방향이 아직 무방비인가 — **하나 있다**(F5, optional)

포함도 동등도 덮지 않는 제3의 방향은 **기준 집합 자체의 정확성**이다.

동등 검사는 구현을 §D.3 의 네 이름에 **고정**한다 — 즉 "구현이 기준을 지키는가"를 잰다. 그 기준 네 개가
oxlint 의 실제 탐색 목록과 같은가는 재지 않는다. 그 진리값은 오로지 인용
<https://oxc.rs/docs/guide/usage/linter/config.html> 에 얹혀 있고, REQ-002·§D.3·§B.2·§D.4·§D.9 어디에도
그 인용을 재검증하거나 날짜를 박거나 재확인 계기를 두는 절이 없다.

역설적이지만 **동등 검사가 이 위험을 키운다**: 상류가 다섯 번째 파일명을 추가하면, AC-004 는 그 이름을
추가하려는 모든 시도를 `unexpected` 로 **능동적으로 거부**한다. 닫힌 규범의 올바른 성질이지만, 그 대가로
SPEC 개정 없이는 상류를 따라갈 수 없다. 테스트로는 닫을 수 없는 종류의 위험이다 → optional (F5).

### R5. F1 수정이 남긴 개수 불일치 — **F4 (blocking)**

probe 가 2개에서 3개로 늘었는데, **종결 게이트만 2를 유지하고 있다.**

```
$ sed -n '/## §D.10/,$p' acceptance.md
- All eight ACs pass, each with the command and its verbatim output recorded in ...
- Both mutant probes of §D.9 observed failing the ACs they are supposed to fail.
```

같은 파일 §D.9 의 마무리 문장은 `All three probes are run and their observed output recorded` 이고,
plan.md §D M1 step 6 도 `Run all three mutant probes` 다. 세 자리 중 **§D.10 한 곳만** 2를 말한다.
상세는 아래 F4.

---

## Claim (주장)

1. 필수통과 7항(MP-1..MP-7) 전부 통과 또는 N/A — 3회차에 `spec.md` 가 개정됐으므로 frontmatter 와
   `moai spec lint` 를 재측정해 재확인했다(§R6.5).
2. 요청받은 5개 축 **전부 통과**(공허한 통과 위험 · control 적정성 · 회귀 표면 · 범위 정직성 ·
   Template-First).
3. 총점 0.9875 (0.875 → 0.9625 → 0.9875, 회귀 없음). Tier M 임계 0.80 초과.
4. **F1 · F4 · F5 전부 닫힘** — 각각 §R1·R2 / §R6.1 / §R6.4 에서 실측 확인.
5. **blocking 0건 → 최종 판정 PASS.** 잔존 F2 · F3 · F6 은 전부 advisory 이며, M6 에 따라 optional 결함
   목록의 길이는 그 자체로 FAIL 근거가 되지 않는다.
6. 이 감사가 확인한 것은 AC 와 probe 의 **설계**이지 그 **실행 관측**이 아니다 — probe 3개는 아직 돌지
   않았고 테스트 파일은 아직 없다(Gaps 참조). 그 관측은 run-phase 의 몫이다.

---

## Evidence (증거 — 실행 명령과 그 출력)

### E1. 감사 트리와 패키지 초록 baseline

```
$ git rev-parse --short HEAD && go test ./internal/hook/quality/...
d060e0d13
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	21.784s
```

SPEC 이 선언한 baseline SHA(`d060e0d13`)와 감사 트리 HEAD 가 동일하다. 이 SPEC 의 모든 RED-now 칸은
이 SHA 에 고정돼 있고, 그 SHA 는 실제로 이 트리다 — 이월 수치가 아니다.

### E2. 필수통과 판정

```
$ moai spec lint SPEC-GATE-OXLINT-DETECT-001 >/dev/null 2>&1; echo "target exit=$?"
target exit=0
$ moai spec lint SPEC-NOPE-999 >/dev/null 2>&1; echo "nonexistent exit=$?"
nonexistent exit=3
```

인자가 실제로 소비된다(없는 ID 는 3). 따라서 exit 0 은 공허한 통과가 아니다.

```
$ /usr/bin/grep -Eo 'SPEC(-[A-Z][A-Z0-9]+)+-[0-9]+' spec.md plan.md acceptance.md progress.md | sort -u
acceptance.md:SPEC-GATE-OXLINT-DETECT-001
plan.md:SPEC-GATE-OXLINT-DETECT-001
progress.md:SPEC-GATE-OXLINT-DETECT-001
spec.md:SPEC-GATE-OXLINT-DETECT-001
$ /usr/bin/grep -c syscall spec.md plan.md acceptance.md
spec.md:0
plan.md:0
acceptance.md:0
$ /usr/bin/grep -rn '\[NEEDS CLARIFICATION' plan.md progress.md acceptance.md spec.md; echo "exit=$?"
exit=1
```

- **MP-1 REQ 번호 정합**: PASS. REQ-001..REQ-007, 결번·중복 없음, 3자리 제로패딩 일관 (spec.md §2).
- **MP-2 GEARS 형식**: PASS (**요구층 `REQ-XXX` 기준 판정**). 7개 전부 Ubiquitous / When / Where 중 하나로
  명시 표기되고 패턴을 만족한다. acceptance.md 의 Given-When-Then 항목은 검증층 `AC-XXX` 이므로 이 축에서
  채점하지 않았다(Group 4 에서 별도 채점).
  - 경계 관측: REQ-004 의 `Where none of oxlint's configuration files is present` 는 설정파일 존재 여부라는
    정적 설정 게이트이므로 GEARS `Where` 로 유효하다. `When` 으로도 읽을 수 있으나 오류가 아니다.
- **MP-3 frontmatter**: PASS. 정규 12필드 전부 존재(`id` `title` `version` `status` `created` `updated`
  `author` `priority` `phase` `module` `lifecycle` `tags`) + `tier` `issue_number`. snake_case alias 없음.
  강제층(`moai spec lint`) exit 0.
  - 관측(이 카드 소관 아님): 스키마 문서 `spec-frontmatter-schema.md` 의 `id` 정규식은 단일 도메인 세그먼트
    `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$` 인데, 실제 강제 구현 `internal/spec/lint.go:1131` 은
    `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` 로 다중 세그먼트를 허용한다. `SPEC-GATE-OXLINT-DETECT-001` 은 구현을
    통과한다. progress.md §E.1 이 기록한 정규식은 문서가 아니라 **구현** 쪽과 일치하므로 이 SPEC 의 결함이
    아니다. 문서↔구현 괴리는 선재 사안이다.
- **MP-4 언어 중립성**: N/A. 이 SPEC 은 `toolchains` 표의 Node 항목 하나에 국한되고, §5 가 다른 툴체인을
  명시적으로 범위 밖에 둔다. 16개 프로그래밍 언어 중 어느 하나를 PRIMARY 로 승격하지 않는다.
- **MP-5 D7 교차-SPEC**: N/A → 자동 통과. 본문이 참조하는 외부 SPEC ID 가 0건(자기 자신뿐)이므로 폐기·대체
  상태를 조회할 대상이 없다.
- **MP-6 D8 크로스플랫폼**: PASS. `syscall` 문자열 0건이므로 D8-4 에 따라 자동 통과. 별건으로 Windows
  분기는 acceptance.md §D.8 이 명시 처리했다(실행 계열 테스트는 `runtime.GOOS == "windows"` 에서 skip,
  AC-004 는 전 플랫폼 실행).
- **MP-7 clarification gate**: PASS. `[NEEDS CLARIFICATION` 0건(exit 1).

### E3. 축 1 — 공허한 통과 위험 (요청된 최중요 축)

**검출 경로 매칭 방식을 먼저 실측했다.** 설정파일 판정은 정확 일치 stat 이다:

```
$ sed -n 1268,1280p internal/hook/quality/gate.go
func (g *QualityGate) anyConfigFileExists(configFiles []string) bool {
	dir := g.stepDir("QualityGate.anyConfigFileExists")
	if dir == "" {
		return false
	}
	for _, cf := range configFiles {
		if fileExists(filepath.Join(dir, cf)) {
			return true
		}
	}
	return false
}
```

이 사실이 두 가지를 결정한다.

- **네 파일명 요구는 실제로 네 번의 측정을 강제한다.** 항목에서 이름 하나가 빠지면 그 이름을 가진 fixture 는
  `anyConfigFileExists` 가 false → `markSkipped` → `rec.outcome == outcomeSkipped` 로 해당 subtest 만
  red 가 된다. 한 subtest 가 우연히 네 이름을 덮을 경로가 없다. 여기에 §D.3 의 swept-set 절
  (`oxlint_config/<filename>` 4건 이름 보고, 3건이면 부분 스윕)이 겹친다 → **이 축은 통과**.
- **반대 방향은 어떤 AC 로도 red 가 되지 않는다.** → F1 (아래).

**AC 별 red 가능성 실측 판정**:

| AC | 지금 red 인가 | 근거 |
|---|---|---|
| AC-001 | RED-now | 트리에 oxlint 항목이 없어 `recordFor("oxlint")` 가 nil (baseline `oxlintprj` 0 steps) |
| AC-002 | RED-now | 스텝이 선택되지 않아 non-zero 종료에 도달하지 못함 |
| AC-003 | RED-now ×4 | baseline `oxlintprj`/`ox2`/`ox3`/`ox4` 각각 0 steps |
| AC-004 | RED-now | Node `lintSteps` 에 `oxlint` 원소 부재 (gate.go:211-226) |
| AC-005 | **RED-now** | 오늘 `oxlint` 행 자체가 없으므로 "oxlint 행이 skipped" 를 단언할 수 없음 |
| AC-006 | **RED-now** | 동일 |
| AC-007 | **RED-now** | 오늘 config-absent 통지는 2건, AC 는 3건을 요구 |

AC-005..007 은 "변화 없음"만 단언하는 장식용 대조군이 **아니다**. 세 행 모두 자체 RED-now 칸을 갖는다.
여기에 §D.9 mutant M1(빈 `configFiles`)이 정확히 이 셋을 red 로 만들도록 설계돼 있고, §D.9 는 그 실행과
관측 기록을 의무화한다. **control 은 하중을 받는다.**

### E4. 축 2 — control 적정성 / 검출 경로

경고된 실패 형태("코드에는 추가됐으나 검출 경로에 닿지 않음")는 방어된다.

- AC-001·002·003·005·006·007 은 전부 `g.Run(ctx)` 를 돌린 뒤 `g.summary.recordFor(...)` 를 읽는다.
  그 행을 채우는 것은 선택 경로 자체다:

```
$ sed -n 1080,1082p internal/hook/quality/gate.go
	if len(step.configFiles) > 0 && !g.anyConfigFileExists(step.configFiles) {
		g.summary.markSkipped(step.name, fmt.Sprintf(reasonConfigFilesAbsentFmt, strings.Join(step.configFiles, ", ")))
		return true, ""
	}
```

  즉 `outcome`/`command` 는 struct 필드 존재가 아니라 **실행 결과**다.
- AC-004 만이 선언 검사인데, acceptance.md §D.4 가 그 경계를 스스로 명시한다("this AC asserts the
  declaration; §D.1–§D.3 assert the behaviour. Neither substitutes for the other"). 요청받은 실패 형태를
  SPEC 이 이름 붙여 격리해 두었다 → **이 축은 통과**.

### E5. 축 3 — 회귀 표면 (요청된 3개 하위 질문 전부 실측)

**(a) eslint 설정 목록과 oxlint 설정 목록의 충돌 여부 — 충돌 없음.**

```
$ sed -n 211,226p internal/hook/quality/gate.go
		lintSteps: []gateStep{{
			name: "eslint", binary: "npx", args: []string{"eslint", "."}, optional: true,
			configFiles: []string{
				"eslint.config.js", "eslint.config.mjs", "eslint.config.cjs",
				"eslint.config.ts", "eslint.config.mts", "eslint.config.cts",
				".eslintrc.js", ".eslintrc.cjs", ".eslintrc.yaml", ".eslintrc.yml", ".eslintrc.json", ".eslintrc",
			},
		}, {
			... (biome comment)
			name: "biome", binary: "npx", args: []string{"biome", "check", "."}, optional: true,
			configFiles: []string{"biome.json", "biome.jsonc"},
		}},
```

`eslint.config.ts` 와 `oxlint.config.ts` 는 접두사만 다른 **별개 파일명**이고, E3 이 보인 대로 판정은
`filepath.Join` + stat 정확 일치다. 접두사 부분일치 경로가 존재하지 않는다. baseline 의 `ox3`/`ox4` 행이
같은 사실을 실측으로도 보였다. SPEC 의 주장은 참이다.

여기에 개정본이 **살아 있는 대조군**을 하나 더 붙였다(acceptance.md §D.3):

```
- **Second finding carried by `ox3` / `ox4`:** the eslint entry must still skip on
  `oxlint.config.ts` and `oxlint.config.mts`. Its own list carries `eslint.config.ts` /
  `eslint.config.mts`, which differ only in prefix, so these two cases double as a
  mis-claim control and each asserts an `eslint` row of `outcomeSkipped`.
```

이제 이 축은 baseline 실측 + 코드 경로 분석 + **AC 대조군** 셋으로 받쳐진다. 축 3의 접두사 충돌 질문은
**닫혔다.**

[HARD] **다만 이 대조군은 F1 을 닫지 않는다 — 방향도 대상도 다르다.** 이 임무가 겨누는 것은
**eslint 항목**이 oxlint 접두사 이름을 잘못 주장하지 않는가이고, F1 이 겨누는 것은 **oxlint 항목 자신의**
`configFiles` 가 oxlint 가 읽지 않는 이름까지 담지 않는가이다. `oxlint.config.js` 를 oxlint 목록에 추가한
구현에서 ox3/ox4 케이스는 여전히 초록이다(그 fixture 들은 `oxlint.config.ts`/`.mts` 를 담고 있고,
eslint 행은 그대로 skipped 다). 두 지적을 같은 것으로 읽으면 F1 이 조용히 사라진다.

**(b) 두 린터 설정을 동시에 가진 프로젝트 — 두 스텝이 돈다. 그리고 SPEC 은 이 경우를 규정하지 않는다.**
`gate.go:501` 의 lint 루프는 `tc.lintSteps` 전체를 순회하므로, `biome.json` 과 `.oxlintrc.json` 을 모두
가진 프로젝트는 두 스텝을 실행한다. 이는 오늘 eslint+biome 에서 이미 참인 선재 동작이고 이 변경이 바꾸지
않는다. 다만 REQ-005 는 한 방향 배타성만 말한다("an eslint project never invokes oxlint") — 병존 조합은
어느 REQ 에도, 어느 AC 에도 없다. → **F2 (optional)**.

**(c) 스텝 순서 / 요약 행 수 단언 — 어긋나는 지점 없음(실측).**

```
$ sed -n 500,507p internal/hook/quality/gate.go
	for _, step := range tc.lintSteps {
		ok, out := g.executeStep(ctx, step, g.config.LintTimeout)
		if !ok {
			return false, g.withSummary(out)
		}
```

lint 루프는 첫 실패에서 즉시 반환(fail-fast)한다. oxlint 는 목록 **끝에** 추가되므로 eslint/biome
프로젝트에서 먼저 이름 불리는 스텝은 변하지 않는다.

행 수 단언 탐색(빈 결과 자체가 결론이므로 명령과 함께 남긴다):

```
$ /usr/bin/grep -rn "records" --include="*_test.go" internal/hook/quality/
internal/hook/quality/gate_empty_source_test.go:203:// TestRuffStep_DeclaresNoSourceExts records why ruff has no counterpart:
internal/hook/quality/gate_step_termination_test.go:43:... records
internal/hook/quality/gate_step_git_env_test.go:65:... records
$ /usr/bin/grep -rn "configured):" --include="*_test.go" internal/
(no output)
```

세 적중은 전부 산문 주석의 영어 단어 "records" 이고 요약 레코드 수 단언이 아니다. 헤더
`"%s (%d configured):"` (gate_summary.go:225)의 N 은 계산값이며 어떤 테스트도 그 문자열을 고정하지 않는다.
`TestSummaryIsCompleteAndVariesWithWhatExecuted` 가 라벨을 열거하지만 그 목록은 **Go 툴체인**
(`go vet`/typecheck/`golangci-lint`/`go test`)이라 Node 항목 추가와 무관하다.

```
$ /usr/bin/grep -rn "npx biome\|\"biome\"" --include="*.go" internal/ pkg/ cmd/ | /usr/bin/grep -v "internal/hook/quality"
(no output)
```

패키지 밖에서 Node 린터 집합을 읽는 Go 코드가 없다. → **회귀 표면 축은 통과.**

### E6. 축 4 — 범위 정직성

- §3(결정) · §3.1(카드 전제 정정) · §5(범위 밖) · plan.md §B.1 · §G 를 대조했고, 결정이 아직 열려 있는 것처럼
  읽히는 절은 **없다**. HISTORY 가 전환 자체를 기록한다("the former 'Open decision' section became §3
  Decision") — 은폐가 아니라 이력이다.
- spec.md §1 의 7행 측정표를 `baseline.md` 와 행 단위로 대조했고 전부 일치한다. 이월 수치 없음.
- 인용된 커밋도 실측 확인했다:

```
$ git log -1 --format='%H%n%s' 09561fd93
09561fd93b015c4cbad799b1bb05b1b813ce5c08
fix(hook): stop silent config-gated lint skip in moai gate (card t233)
$ git merge-base --is-ancestor 09561fd93 HEAD && echo "ancestor=yes"
ancestor=yes
```

  커밋 본문이 issue #1631 과 biome 항목 추가를 직접 명시한다. §3.1 의 카드-전제 정정은 근거를 갖췄다.
- 미귀속 주장은 **찾지 못했다.** REQ-007 이 이 규율을 SPEC 안에 요구로 박아 넣었고, 아티팩트 자신이 그것을
  지킨다.

### E7. 축 5 — Template-First 주장 검증

plan.md §E / spec.md §4 의 "적용되지 않는다" 주장을 경로로 직접 확인했다:

```
$ find internal/template/templates -name "gate.go"
(no output)
$ /usr/bin/grep -rn "oxlint" internal/template/templates/
(no output)
$ ls -l internal/hook/quality/gate.go
-rw-r--r--@ 1 goos  staff  61500 Sep 10 13:40 internal/hook/quality/gate.go
```

**주장은 참이다.** `gate.go` 는 템플릿 트리 밖이며 미러 대상이 없다. `make build` 의무도 따라 나오지 않는다.

부수 관측(위반 아님):

```
$ /usr/bin/grep -rln "biome" internal/template/templates/
internal/template/templates/.claude/rules/moai/languages/javascript.md
$ /usr/bin/grep -n "Linting" internal/template/templates/.claude/rules/moai/languages/javascript.md
20:- Linting: ESLint 9 flat config, Biome
```

이 파일은 게이트 표의 미러가 아니라 사용자용 생태계 안내이므로 Template-First 의무를 발생시키지 않는다.
다만 게이트가 인식하는 린터 집합보다 한 개 뒤처지게 된다 → **F3 (optional)**.

---

## Baseline-attribution (baseline 귀속)

이 보고서의 모든 수치·인용은 다음에서 이 감사 실행 중에 직접 측정했다.

- 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t550`, HEAD `d060e0d13` (E1 의 `git rev-parse` 출력)
- 패키지 초록: `go test ./internal/hook/quality/...` → `ok ... 21.784s` (E1)
- 소스 인용: 위 트리의 `internal/hook/quality/gate.go`, `sed -n` 로 지정 구간 발췌 (E3·E4·E5)
- SPEC 판정: `moai spec lint` exit 0 / 없는 ID exit 3 (E2)
- 커밋 주장: `git log` + `git merge-base --is-ancestor` (E6)

SPEC 이 선언한 baseline SHA 와 감사 트리 HEAD 가 같으므로, SPEC 의 RED-now 칸은 이 감사가 본 트리와 같은
트리에 고정돼 있다. 다른 트리·다른 시점의 수치를 재사용한 곳은 이 보고서에 없다.

---

## Defects Found (구조화 결함 목록)

**F6. §D.9 도입부의 범위 절이 AC-004 를 세지 않는다 — acceptance.md:148 — Severity: minor — Class: optional**

`Before adopting AC-001/AC-003, the run phase writes a mutant...` — probe 연습이 지키는 AC 를 열거하는데
M3 이 지키는 **AC-004** 가 그 열거에 없다(pre-M3 범위 절). 의무는 §D.10(`the ACs they are supposed to
fail`)과 M3 자신의 문장(`must fail AC-004`)이 지므로 **아무 의무도 탈락하지 않는다** → optional.

**Required fix:** `Before adopting AC-001/AC-003/AC-004` 로 한 토큰 추가. 생략해도 무방하다.

---

**F4. [해소됨 — 3회차 검증] §D.10 종결 게이트가 probe 3개를 2개로 센다 — acceptance.md:174 — Severity: minor — Class: blocking → CLOSED**

> **3회차 판정: 계열 전체가 닫혔다.** §D.10 은 `All three mutant probes of §D.9 (M1, M2, M3)` + M3 비선택
> 조항으로, AC-008 매트릭스 행은 `all three mutant probes` 로 개정됐다. `Both mutant` 적중 0건.
> 개수 축 전수 스윕(probe·AC·파일명·통지·항목) 결과 스테일 개수 0건 — §R6.1 표. 잔존 단수 2건은
> 총칭 표현이지 개수 주장이 아니다(§R6.3).
>
> **이 결함에 대한 내 지적 자체에 결함이 있었다:** 형제(AC-008 행)를 찾고도 `필수는 아니다` 로 등급을
> 낮췄다. 개수 수정을 이름 붙인 인스턴스에서 멈추면 계열이 살아남는다 — 경위와 교정은 §R6.2.

아래는 2회차 원문(이력 보존).

F1 수정이 §D.9 에 M3 을 추가하면서 probe 는 2 → 3 이 됐다. 그런데 종결 체크리스트만 옛 수를 유지한다:

- `acceptance.md:174` (§D.10 Definition of Done): **`Both mutant probes of §D.9`**
- `acceptance.md:167` (§D.9 마무리): `All three probes are run`
- `plan.md:79` (§D M1 step 6): `Run all three mutant probes`

세 자리 중 한 자리만 2다. **하필 그 한 자리가 종결 게이트다.** §D.10 을 문자 그대로 읽고 M1·M2 만 돌린
run-phase 는 DoD 를 충족한 것으로 닫히는데, 그때 빠지는 M3 은 **F1 수정을 실증하는 바로 그 probe** 다.
§D.9 가 스스로 적은 문장이 이 위험을 정확히 서술한다 — `without it the §D.4 fix is asserted rather than
demonstrated`. 즉 이 개수 하나가 어긋난 상태로 닫히면, F1 의 수정은 실증되지 않은 채 실린다.

위험 폭은 좁다: 실행 지시(§D.9, plan.md §D M1 step 6)는 두 자리 모두 3을 말하므로, milestone 을 따라
일하는 run-phase 는 세 개를 돌린다. 물리는 것은 종결 게이트를 plan 과 §D.9 에서 떼어 단독으로 읽을 때다.
그럼에도 blocking 으로 분류하는 이유는 두 가지다 — (1) 문서가 스스로 명시한 기준(probe 수)이 내부에서
충돌한다, (2) 충돌하는 쪽이 하필 종결 판정면이다.

**Required fix:** `acceptance.md:174` 의 `Both mutant probes` → `All three mutant probes` (한 단어).
아울러 §D.9 의 도입 문장 `the run phase writes a mutant`(단수)와 §D 매트릭스 AC-008 행의
`the mutant probe of §D.9`(단수)도 같은 손질에 포함하면 단수/복수 표현이 세 probe 체제와 일관된다 — 이 둘은
열거 목록이 뒤따르므로 오독 위험이 낮아 **필수는 아니다**.

---

**F1. [해소됨 — 2회차 검증] REQ-002 의 부정 절이 어떤 AC 로도 red 가 되지 않는다 — acceptance.md §D.4 — Severity: major — Class: blocking → CLOSED**

> **2회차 판정: 닫혔다.** §D.4 가 `len(configFiles) == 4` + 양방향 소속 동등 + 방향별 구별 메시지로 개정됐고,
> §D.9 에 성장 방향 대조군 M3 이 추가됐으며, M3 이 AC-004 하나만 흔든다는 성질을 fixture 전수 열거로
> 확인했다(§R1·R2). 목록을 위에서 묶는 단언이 없다는 1회차의 근거는 더 이상 성립하지 않는다.
> 아래 1회차 기록은 이력으로 보존한다.

REQ-002 는 규범으로 못박는다: *"No other filename is recognized."* 그런데 그 절을 겨누는 AC 가 없다.
AC-004 는 **포함(membership)** 검사다 — "its `configFiles` contains all four names of §D.3 — asserted as a
set membership check per name". 포함 검사는 목록이 **더 길어지는** 방향을 잡지 못한다. 따라서 run-phase 가
`configFiles` 에 `oxlint.config.js` 나 `.oxlintrc.yaml`(oxlint 가 읽지 않는 이름)을 추가해도
AC-001 ~ AC-008 이 **전부 초록**으로 통과한다.

이것이 장식적 지적이 아닌 이유는 SPEC 자신이 그 위험에 이미 이름을 붙였기 때문이다 — plan.md §B.2:
*"Adding a name that oxlint does not read would make the gate run oxlint on a project oxlint itself would
ignore; omitting one leaves a real oxlint project uncovered."* 두 방향을 나란히 적어 두고 **누락 방향에만**
대조군(AC-003 subtest 4개, mutant M2)을 붙였다. 과잉인식 방향은 무방비다.

같은 SPEC 이 §D.3 에서 스스로 세운 기준 — *"a four-name claim resting on one measured name would leave
three asserted-but-unmeasured"* — 을 부정 방향에 적용하지 않은 비대칭이기도 하다.

**Required fix (택일, 어느 쪽도 한 줄 수준):**
(a) AC-004 의 집합 검사를 **동등성**으로 승격 — `len(oxlint.configFiles) == 4` 이고 정렬한 집합이 §D.3 의
네 이름과 정확히 같을 것. 여분의 이름이 그 이름과 함께 실패 메시지에 뜨도록 한다. 또는
(b) AC-009 신설 — `oxlint.config.js`(oxlint 미인식, 형태만 유사) 하나만 가진 fixture 에서 `oxlint` 행이
`outcomeSkipped` 임을 단언. 아울러 §D.9 에 mutant M3(목록에 미인식 이름 1개 추가)를 넣어 그 AC 가 실제로
red 가 되는지 관측 기록을 남긴다.

(a) 가 더 싸고 §D.4 의 성격(선언 고정)과 결이 같다. (b) 는 행동층까지 덮으므로 더 강하다.

**[1회차 이력] 개정본 반영 재확인.** 당시 리드가 지목한 AC-003 의 ox3/ox4 이중 임무가 이 결함을 닫는지
확인했고 닫지 않았다 — 그 임무는 **eslint 항목**의 과잉주장을 겨누고, F1 은 **oxlint 항목 자신의** 목록을
겨눈다(E5-a 의 [HARD] 단락). 그 구분은 지금도 유효하며, F1 을 실제로 닫은 것은 **그 뒤의 §D.4 동등성
개정과 M3 추가**다(§R1·R2).

---

**F5. [해소됨 — 3회차 검증] 기준 집합 자체의 정확성이 어느 방향으로도 방어되지 않는다 — spec.md §2 (REQ-002) — Severity: minor — Class: optional → CLOSED**

> **3회차 판정: 닫혔다.** REQ-002 의 인용에 `(accessed 2026-09-10)` 가 붙었고, 파일명 목록 **바로 아래**
> blockquote 로 Reference-set guard 가 들어갔다 — 동등 검사가 위험을 키운다는 성질, 상류 5번째 파일명은
> REQ-002 개정 사안이라는 처분 규칙, 그리고 `never a test failure to route around by loosening AC-004`.
> 위치도 요구대로다: 목록을 고치려는 독자가 반드시 지나는 자리다(§R6.4). 미검증 가정이 **기록된 판정**이
> 됐다.

아래는 2회차 원문(이력 보존).

리드의 세 번째 질문("포함도 동등도 덮지 않는 제3의 방향이 있는가")에 대한 답이다. **있다.**

동등 검사는 *구현이 기준을 지키는가*를 잰다. *그 기준 네 개가 옳은가*는 재지 않는다. 네 이름의 진리값은
인용 <https://oxc.rs/docs/guide/usage/linter/config.html> 하나에 얹혀 있고, 그 인용을 재검증하거나 조회
날짜를 박거나 재확인 계기를 두는 절이 REQ-002·§D.3·§B.2·§D.4·§D.9 어디에도 없다.

동등 검사가 이 위험을 **키운다**는 점이 이 항목의 요지다. 상류가 다섯 번째 탐색 파일명을 추가하면 AC-004 는
그것을 더하려는 시도를 `unexpected` 로 능동 거부한다 — 닫힌 규범의 올바른 성질이되, SPEC 개정 없이는
상류를 따라갈 수 없다는 뜻이기도 하다. 테스트로 닫을 수 있는 종류가 아니므로 optional 로 둔다.

**Required fix (제안, 리드 재량):** §D.3 또는 REQ-002 의 URL 인용에 조회 날짜를 붙이고, "이 목록이 상류에서
바뀌면 그것은 테스트 수정이 아니라 REQ-002 개정 사안" 한 문장을 §5 또는 §D.4 에 남긴다. 그러면 이 의존이
미검증 가정이 아니라 **기록된 판정**이 된다.

---

**F2. 두 린터 설정을 동시에 가진 Node 프로젝트의 동작이 규정되지 않았다 — spec.md §2 (REQ-005) — Severity: minor — Class: optional**

REQ-005 는 한 방향만 말한다("an eslint project never invokes oxlint and an oxlint project never invokes
eslint or biome"). `biome.json` 과 `.oxlintrc.json` 을 **둘 다** 가진 프로젝트는 실측상 두 스텝을 실행하는데
(E5-b), 이 조합을 규정한 REQ 도 AC 도 없다.

선재 동작이며 이 변경이 바꾸지 않으므로 회귀는 아니다. optional 로 분류하는 이유가 그것이다.

**Required fix:** REQ-005 에 한 문장을 덧붙이거나(설정을 여러 개 가진 프로젝트는 해당하는 스텝을 모두
실행하며, 이는 기존 eslint+biome 동작과 동일하다), §5 Out of Scope 에 "린터 병존 조합" 한 줄을 넣어
의도적 미규정임을 명시한다. 어느 쪽이든 판정이지 누락이 아니게 된다.

---

**F3. 템플릿 생태계 안내 문서가 게이트보다 한 린터 뒤처지게 된다 — internal/template/templates/.claude/rules/moai/languages/javascript.md:20 — Severity: minor — Class: optional**

`- Linting: ESLint 9 flat config, Biome`. 이 파일은 게이트 `toolchains` 표의 미러가 **아니므로**
Template-First 위반이 아니다(E7 에서 확인). 그러나 이 변경 이후 게이트는 oxlint 를 인식하는데 사용자에게
배포되는 안내는 그렇지 않다고 읽힌다.

**Required fix:** 이 카드에서 다룰지 여부는 리드 판단. 다룬다면 해당 줄에 oxlint 를 추가하고, spec.md §4 에
"이 파일 한 줄은 미러 의무가 아니라 문서 정합 차원에서 함께 고친다"를 명시해 plan.md §E 의 "템플릿 미러 없음"
주장과 모순으로 읽히지 않게 한다. 다루지 않는다면 §5 에 한 줄로 범위 밖 선언을 남긴다.

---

## Category Scores (0.0-1.0, rubric-anchored)

| 차원 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 1.00 | 1.0 | REQ-001..007 각각 단일 해석. 지시대명사 모호 없음. 규범 문장에 "should"/"may" 없음(weasel 스캔 exit 1). REQ-002 의 파일명 네 개는 열거로 확정 |
| Completeness | 0.95 | 1.0 하단 | 필수 절·frontmatter 전부 존재, `### Out of Scope — <topic>` H3 4개 각각 `-` 항목 보유. **REQ-002 부정 절 커버됨(F1), probe 개수 계열 정합(F4), 기준집합 의존이 기록된 판정으로 전환(F5).** 잔여 감점: 린터 병존 조합 미규정(F2), §D.9 범위 절(F6) — 둘 다 optional |
| Testability | 1.00 | 1.0 | 8개 AC 전부 이진 판정 가능, weasel word 0건. AC-004 가 양방향 동등 + 방향별 구별 메시지로 성장 방향까지 관측 가능. 테스트로 닫을 수 없는 기준집합 축은 F5 가드로 **문서화된 경계**가 됐으므로 미검증 갭이 아니다 |
| Traceability | 1.00 | 1.0 | REQ-001→AC-001 / 002→AC-001·003·004 / 003→AC-002 / 004→AC-007 / 005→AC-005·006 / 006→AC-007 / 007→AC-008. 미커버 REQ 0건, 존재하지 않는 REQ 를 가리키는 AC 0건 |

**총점 0.9875** (1회차 0.875 → 2회차 0.9625 → 3회차 0.9875, Tier M 임계 0.80 대비 큰 폭 초과).
점수가 한 번도 회귀하지 않았으므로 STOP 에스컬레이션 조건(점수 하락)에 해당하지 않는다. 반복 3/3 —
ceiling 에 닿았으나 **PASS 로 닫히므로** 사용자 에스컬레이션(PASS-with-debt / 범위 축소 / 상한 연장)은
필요 없다.

weasel 스캔 실측:

```
$ /usr/bin/grep -n -i -E "appropriate|adequate|reasonable|proper|as needed|if possible" spec.md acceptance.md plan.md; echo "exit=$?"
exit=1
```

---

## Gaps (미검증 — 보지 않은 것)

이 절은 비어 있지 않다. 아래는 이 감사가 **관측하지 않은** 것들이며, 통과로 읽어서는 안 된다.

- **작성될 테스트 파일의 실제 RED 는 보지 않았다.** `gate_oxlint_lint_test.go` 는 아직 존재하지 않는다.
  AC-001..003·005..007 의 RED-now 는 baseline 측정(0 lint steps)과 소스 부재로부터의 **연역**이며, 테스트를
  돌려 red 를 본 것이 아니다. 그 관측은 plan.md §D M1-1 이 run-phase 의무로 부과한다 — 이 감사는 그 의무가
  **적혀 있음**만 확인했다.
- **mutant probe(§D.9 M1·M2·M3)를 직접 돌리지 않았다.** control 이 red 가 될 수 있다는 판정(E3)과 M3 이
  AC-004 만 흔든다는 판정(§R2)은 둘 다 **문서 열거 + 코드 경로 분석**(`anyConfigFileExists` 정확 일치,
  `markSkipped` 경로, fixture Given 절 전수)에 근거한 것이며, mutant 를 심어 관측한 결과가 아니다.
  M3 이 실제로 AC-004 를 red 로 만드는지는 run-phase 가 관측할 몫이다 — 이 감사는 그 probe 가 **판별
  가능한 형태로 설계됐음**만 확인했다.
- **실제 `npx oxlint` 는 실행하지 않았다.** oxlint 의 설정 탐색 순서를 실행으로 확인하지 않았고, SPEC 이
  인용한 <https://oxc.rs/docs/guide/usage/linter/config.html> 를 이 감사가 fetch 하지도 않았다 — 네 파일명의
  정확성은 **인용 신뢰**이지 이 감사의 관측이 아니다.
- **전체 스위트를 돌리지 않았다.** `go test ./internal/hook/quality/...` 하나만 실행했다(CLAUDE.local.md §4).
  다른 패키지가 Node 린터 집합에 의존하는지는 grep(E5-c)으로만 확인했고 실행으로 확인하지 않았다.
- **Windows/Linux 에서 아무것도 재지 않았다.** darwin 단일 플랫폼 관측이다.
- **`golangci-lint` / `go vet` 을 돌리지 않았다.** acceptance.md §D.10 이 요구하지만 run-phase 의무다.
- **plan.md 의 우선순위 라벨(High/Medium) 타당성은 판단하지 않았다.** 감사 대상 축이 아니다.
- **cross-model 백엔드(codex/GLM) 2차 의견을 부르지 않았다.** 이 프로젝트의 `audit_model` 을 조회하지
  않았고, 단일(claude) 감사로 판정했다.
- **개정 이전 사본과 diff 를 뜨지 않았다.** 그 사본은 커밋되지 않아 내가 볼 수 없다. E0 은 mtime 이
  내가 읽은 시점 이후로 움직이지 않았음과 리드가 지목한 네 변경이 내 사본에 들어 있음을 보인 것이지,
  "그 외에는 아무것도 바뀌지 않았다"를 세운 것이 아니다. 리드가 열거하지 않은 변경이 13:55 이전에
  있었다면 이 감사는 그것을 개정 전후로 구분하지 못한 채 개정본으로 채점했다 — 판정 자체는 개정본에
  대한 것이므로 유효하지만, 무엇이 언제 바뀌었는지는 내 관측이 아니다.

---

## Residual-risk (잔여 위험 — 관측한 것에도 불구하고 여전히 틀릴 수 있는 것)

- **F1 을 (a) 방식으로만 고치면 선언층만 잠긴다.** 집합 동등성 검사는 `toolchains` 표를 고정할 뿐,
  선택기가 그 목록을 어떻게 읽는지는 여전히 §D.1-§D.3 의 행동층에 맡겨진다. 지금 구현(정확 일치)에서는
  둘이 어긋날 수 없지만, 매칭 방식이 훗날 접두사/glob 로 바뀌면 (a) 는 그 변화를 잡지 못한다.
- **AC-007 의 "세 개의 통지" 는 통지 개수를 세는 단언이다.** 다른 카드가 Node 툴체인에 네 번째 lint 항목을
  추가하면 이 AC 는 그 카드에서 red 가 된다 — 이 SPEC 의 결함은 아니지만, 이 행이 다른 작업의 마찰 지점이
  될 수 있다는 사실은 남는다.
- **zero-config oxlint 갭은 닫히지 않는다.** SPEC 이 이를 알고 닫지 않기로 한 결정(§3)과 명시적 범위 밖
  선언(§5)을 갖췄으므로 감사 결함이 아니다. 다만 issue #1631 의 신고자가 zero-config 사용자였다면 이 SPEC
  을 다 구현해도 신고 증상이 남는다 — 그 모집단은 **누구도 측정한 적이 없고**(§3 이 그렇게 적었다), 이
  감사도 측정하지 않았다.
- **fail-fast 상호작용.** lint 루프가 첫 실패에서 반환하므로, eslint 와 oxlint 를 모두 가진 프로젝트에서
  eslint 가 red 면 oxlint 는 아예 실행되지 않는다. F2 가 규정 공백으로 남아 있는 한 이 조합의 기대 동작은
  문서화된 근거 없이 구현 순서에 의존한다.

---

## Recommendation (권고) — 3회차 최종

**PASS.** blocking 0건. run-phase 진입 가능하다.

**Advisory 3건 — 전부 리드 재량, 어느 것도 게이트가 아니다:**

- **F2** — 두 린터 설정 병존 조합이 어느 REQ/AC 에도 없다. 선재 동작이라 회귀는 아니다. REQ-005 에 한 문장,
  또는 §5 에 범위 밖 한 줄.
- **F3** — `internal/template/templates/.claude/rules/moai/languages/javascript.md:20` 의 린터 목록이 게이트보다
  한 개 뒤처지게 된다. Template-First 위반은 아니다(미러가 아님).
- **F6** — `acceptance.md:148` 의 `Before adopting AC-001/AC-003` 에 AC-004 추가. 한 토큰.

**run-phase 가 반드시 지고 가야 할 것 — 이 감사가 확인한 것은 "설계"이지 "관측"이 아니다:**

1. AC-001/002/003/005/006/007 을 **pre-edit 트리에서 red 로 관측**하고 출력을 기록한다(plan.md §D M1-1).
2. **mutant 세 개를 전부 돌린다.** 특히 M3 — 그것이 통과하면 §D.4 는 동등성으로 쓰이지 않은 것이고,
   F1 수정은 실증되지 않은 채 실린다. §D.10 이 이제 이것을 비선택으로 못박았다.
3. AC-003 의 subtest 4건이 **이름으로** 보고되는지 확인한다(3건이면 부분 스윕).

**이 카드에서 나온 감사자 교훈(내 결함):** 개수 수정을 내가 이름 붙인 인스턴스에서 멈추면 계열이 살아남는다
(§R6.2). 1회차에 F1 을 "문구가 아니라 요구로" 훑어 잡아 놓고, F4 에서는 인스턴스 단위로 멈춰 형제를 찾고도
등급을 낮췄다. 3회차의 개수 축 전수 스윕(§R6.1)이 그 교정이다.

---

## Recommendation (권고) — 2회차 [이력]

**F1 은 제대로 닫혔다.** 요청받은 세 질문에 각각: (1) 새 §D.4 는 성장 방향을 **닫는다** — 기수 조건·양방향
동등·방향별 구별 메시지 셋이 다 있어 실패가 관측 가능한 형태로 특정된다(§R1). (2) M3 은 **판별한다** —
fixture Given 절 전수 열거로 `oxlint.config.js` 를 담은 fixture 가 0개임을 확인했으므로 행동층은 전부
불변이고 AC-004 만 흔들린다(§R2). (3) 무방비한 제3 방향은 **하나 있다** — 기준 집합 자체의 정확성(F5,
optional, 테스트로 닫을 수 없는 종류).

**FAIL 은 단 한 건, 그것도 F1 수정이 남긴 개수 불일치다.** F1 재개가 아니다.

1. **F4 (blocking) — 한 단어.** `acceptance.md:174` 의 `Both mutant probes` → `All three mutant probes`.
   §D.9 와 plan.md §D M1 step 6 은 이미 3을 말하므로, 종결 게이트만 맞추면 세 자리가 일치한다.
2. **F5 (optional).** §D.3 / REQ-002 의 URL 인용에 조회 날짜 + "상류 변경은 REQ-002 개정 사안" 한 문장.
   리드 재량.
3. **F2 · F3 (optional).** 1회차와 동일, 변동 없음.

**리드가 F4 를 즉시 덮어써도 근거는 충분하다.** 필수통과 7항 전부 통과, 총점 0.9625, 실행 지시 두 자리
(§D.9 · plan.md §D M1 step 6)가 모두 3을 말하므로 milestone 을 따라 일하는 run-phase 는 세 probe 를 돌린다.
F4 가 물리는 것은 종결 게이트를 단독으로 읽을 때뿐이다. 그럼에도 내가 blocking 으로 올리는 이유는 1회차에
F1 을 올린 것과 **같은 규칙**이기 때문이다 — 문서가 스스로 명시한 기준이 내부에서 충돌하고, 충돌하는 쪽이
판정면이다. 규칙을 이번만 느슨하게 적용하면 1회차 판정의 근거도 함께 약해진다.

**재감사 범위:** F4 한 줄에 한정한다. 그것이 닫히면 남는 F2·F3·F5 는 전부 optional 등급이므로 그 자체로는
FAIL 을 만들지 않는다 — 즉 다음 회차는 PASS 가 된다.
