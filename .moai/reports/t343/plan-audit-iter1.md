# SPEC Review Report: SPEC-RED-NOW-THRESHOLD-001

Iteration: 1/2 (Tier M ceiling = 2, `harness.plan_audit_tier_ceilings`)
Verdict: **FAIL**
Overall Score: **0.75**
Tier: **M** (`spec.md:14` `tier: M`) · PASS threshold: **0.80** (`spec-workflow.md` § SPEC Complexity Tier)
감사 트리: `WT-red-now-threshold@a6bbbf82b` (`git rev-parse --short HEAD` → `a6bbbf82b`)

Reasoning context ignored per M1 Context Isolation. 감사 입력은 Tier M 아티팩트 3종(`spec.md`,
`plan.md`, `acceptance.md`) + `progress.md`이며, 카드 배차문의 사전 측정 5건은 근거로 채택하지
않고 전부 이 트리에서 직접 재실행했다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -n "^\*\*REQ-RNT-" spec.md` → 12행, `REQ-RNT-001`
  ~ `REQ-RNT-012`, 3자리 zero-padding 일관, 결번 0, 중복 0. (`spec.md:133,135,137,139,141,143,145,147,149,151,153,155`)
- **[PASS] MP-2 GEARS/EARS format compliance** — **요구 계층(`REQ-XXX`) 기준으로 판정**. 12개 전부
  GEARS 패턴 또는 그 합성절에 대응한다: Where(001·003·006), Ubiquitous(002·004·007·008·009·012),
  Event-driven(005·010), Unwanted(002·009·011·012의 `shall not`). Given-When-Then은 검증 계층
  (`acceptance.md` AC 셀)에만 존재하므로 여기서 감점하지 않았다(M3 § Scope).
- **[PASS] MP-3 YAML frontmatter validity** — 정본 12필드 전부 존재(`spec.md:2-13`),
  `tier: M`은 선택 필드. `grep -c "created_at:\|updated_at:\|labels:\|spec_id:" spec.md` → `0`
  (거부되는 snake_case 별칭 없음). `phase: "v3.1.4 target"`는 금지된 라이프사이클 토큰(plan/run/
  sync/mx) 아님.
- **[PASS] MP-4 language neutrality** — 배포 템플릿에 들어가는 것은 `plan-auditor.md` 미러의 MP-8
  절뿐이고, Go 테스트와 fixture는 REQ-RNT-012가 명시적으로 배포 트리 밖에 둔다. 템플릿 반입분에
  언어별 도구명이 없다. Go 관련 토큰(`internal/spec/`, `make build`)은 전부 비배포 범위.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — 참조 3건.
  `SPEC-TODO-SQLITE-001` → `grep -n "^status:" …/spec.md` → `5:status: completed`
  (retired/superseded/archived 아님 → D7-4 미발화).
  `SPEC-TODO-LANDING-STATE-001` → `.moai/specs/`에 부재(`ls` exit 1) → D7-5 SHOULD, 그러나
  `spec.md:91-102`와 `plan.md:167-171`이 "unlanded, `.claude/worktrees/t331/`에만 존재, 파일
  의존 없음"을 명시 → 해소됨.
  `SPEC-AUTH-001` → 부재이나 템플릿 미러의 illustrative placeholder를 인용한 것(`spec.md:153`)
  → 해소됨. **BLOCKING 없음.**
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` → `0`. D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-RED-NOW-THRESHOLD-001/`
  → 매치 없음(exit 1). `research.md`는 Tier M이라 부재. **다만 D5 참조** — `plan.md` §F M1이
  마커 없이 미해결 결정 2건을 들고 있다.

---

## MP-8 (this SPEC's own proposal) applied to itself

SPEC이 자기 적용을 주장하므로, 제안된 MP-8을 이 SPEC의 `acceptance.md`에 그대로 걸었다.

### L2 재실행 — release-blocking 10건 전부

명령 8종(AC-004/006 공유, AC-009a/b 공유)을 이 워크트리 루트에서 재실행했다. 배차문이 이미
확인했다는 5건도 포함해 **전부 직접 재실행**했다.

| AC | 재실행한 명령 | 관측 출력 | exit | 셀 주장과 일치 |
|----|---------------|-----------|------|----------------|
| AC-RNT-001 | `grep -c "RED-now cell content" .claude/rules/moai/development/verification-completeness.md` | `0` | 1 | 일치 |
| AC-RNT-003 | `grep -c "regression-guard" .claude/rules/moai/development/verification-completeness.md` | `0` | 1 | 일치 |
| AC-RNT-004 | `grep -c "MP-8" .claude/agents/moai/plan-auditor.md` | `0` | 1 | 일치 |
| AC-RNT-005 | `ls internal/spec/red_now_cell_test.go` | `ls: internal/spec/red_now_cell_test.go: No such file or directory` | 1 | 일치 |
| AC-RNT-006 | `grep -c "MP-8" .claude/agents/moai/plan-auditor.md` | `0` | 1 | 일치 |
| AC-RNT-007 | `grep -c "AC-6:" .claude/agents/moai/plan-auditor.md` | `0` | 1 | 일치 |
| AC-RNT-008 | `grep -c "MOAI-REDNOW-BEGIN" .claude/agents/moai/plan-auditor.md` | `0` | 1 | 일치 |
| AC-RNT-009a | `ls internal/spec/testdata/red_now/` | `ls: …: No such file or directory` | 1 | 일치 |
| AC-RNT-009b | `ls internal/spec/testdata/red_now/` | 동일 | 1 | 일치 |
| AC-RNT-010 | `grep -c "MOAI-REDNOW-BEGIN" internal/template/templates/.claude/agents/moai/plan-auditor.md` | `0` | 1 | 일치 |

**10/10 재현. release-blocking 셀에 관해서는 SPEC의 중심 주장이 자기 증거로 성립한다.**

### 부수 검증 — §A.2의 load-bearing 관측

SPEC의 결정적 근거(`spec.md:66-71`)도 재실행했다:

```
$ go test -count=1 ./internal/kanban -run TestMigrationParity
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.449s   (exit 0)
```

존재하지 않는 테스트를 겨눈 `-run` 셀렉터가 exit 0 + `ok`를 낸다는 관측은 이 트리에서 재현된다
(소요시간만 0.216s→0.449s로 다르며, 이는 주장 대상이 아니다). `grep -c "Red via missing test."
.moai/specs/SPEC-TODO-SQLITE-001/acceptance.md` → `7` — `plan.md` §B의 dispatch 수정(4→7)도 정확하다.

### 강등 정직성 — regression-guard 3건

MP-8은 release-blocking만 구속하므로 이 3건은 MP-8 밖이다. 그래서 별도로 검사했고, **여기서
두 개의 결함이 나왔다**(D1, D2).

| AC | 강등이 정직한가 | 판정 |
|----|-----------------|------|
| AC-RNT-002 | 뒤집을 RED이 실제로 없다(보존 대상 속성). 강등 자체는 정직 | **명령이 공허(D1)** |
| AC-RNT-011 | 전수 count-zero 형태는 red-at-arrival-and-forever(446·464 확인)이므로 좁힌 것이 옳다 | **좁히기는 타당, 표현이 부정확(D3)** |
| AC-RNT-012 | 오늘 green인 이유가 "아직 산출물이 없어서"임을 셀이 스스로 밝힌다. 정직 | **명령이 복사-실행 불가(D2)** |

**AC-RNT-011의 좁히기 검증** — `grep -nE "SPEC-[A-Z]+-[0-9]{3}" internal/template/templates/.claude/agents/moai/plan-auditor.md`
→ `446:…SPEC-AUTH-001/…`, `464:…SPEC-AUTH-001/…` (exit 0). 두 placeholder는 실재하며, 이 작업이
건드리지 않는다. 전수 범위 기준은 rule §2의 impossible / wrong-reason-red 방향에 정확히 해당하므로
좁힌 판단은 **타당하다**. 형제 release-blocking 10건은 전부 이 작업의 M1~M4가 뒤집을 수 있는
토큰 카운트/파일 부재 기준이므로 **같은 처리를 필요로 하는 형제는 없다**.

### MP-8 자기적용 결론

MP-8을 문언 그대로 걸면 이 SPEC은 **통과한다**. `acceptance.md` §D.5의 자기적용 주장은
**문언상 참이다** — 그러나 통과하는 이유가, 실제로 결함이 있는 두 셀(AC-RNT-002·AC-RNT-012)이
MP-8이 선언한 범위(release-blocking) 밖에 있기 때문이다. 자기적용이 성립하되 공허하다. 이것이
이번 감사의 머리 소견이며 D4로 기록한다.

---

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | 대부분 단일 해석. 다만 `spec.md:133` REQ-RNT-001의 "the command"가 정의되지 않았고(`plan.md:94-97`이 열린 결정으로 명시), `spec.md:153` REQ-RNT-011이 트리 핀 없는 줄번호를 요구문 안에 박았다. |
| Completeness | 0.75 | 0.75 | 필수 섹션·frontmatter 12필드·`### Out of Scope` H3 5개(`spec.md:176,181,187,194,199`) 전부 존재. `acceptance.md` §D.3 continued-firing이 rule §1.3이 묻는 질문(비실행 감지)에 답하지 않고 §1.2/§2(계약 드리프트)에만 답한다 — MP-8 쪽은 아예 공백. |
| Testability | 0.50 | 0.50 | 13개 전부 Given-When-Then, weasel word 0(`grep -niE "appropriate\|adequate\|reasonable\|proper\|…" acceptance.md` → 매치 1건은 "property" 안의 "proper" 오탐). 그러나 5개의 green path가 명령이 판정하지 못하는 산문 접속절을 달고 있고(D5), 명령 2개가 실제로 결함(D1·D2). |
| Traceability | 1.00 | 1.0 | `acceptance.md:37-52` 매핑표. REQ 12 ↔ AC 13(`grep -c "^| \*\*AC-RNT-"` → `13`, release-blocking `grep -c "release-blocking |"` → `10`). 고아 AC 0, 미커버 REQ 0. Tier M 예산(16/16) 이내. |

**Aggregate = (0.75 + 0.75 + 0.50 + 1.00) / 4 = 0.75 < 0.80 (Tier M threshold)**

---

## Defects Found

**D1. AC-RNT-002의 RED 셀 명령이 공허한 green이다** — `acceptance.md:22` — Severity: **critical** — Class: **blocking**

셀에 적힌 명령을 원문 그대로(`\|` 포함) 실행하면 ERE에서 `\|`는 교대(alternation)가 아니라
**리터럴 파이프 문자**다. 즉 `tense|mood|counterfactual|future?sense` 라는 한 덩어리 문자열을
찾으며, 판별어가 실제로 등장해도 절대 매치하지 않는다. 뮤테이션으로 발산을 관측했다:

```
# 판별어 'mood'를 심은 사본에 대해
$ grep -cE "tense\|mood\|counterfactual\|future.sense" <mutant>   → 0    ← 원문 형태: 거짓 green
$ grep -cE "tense|mood|counterfactual|future.sense"   <mutant>   → 1    ← 올바른 교대: 참 red
```

오늘 `0:0`이 나오는 것은 판별어가 정말로 없기 때문이지만(올바른 형태로도 `0:0` 확인), **명령
자체는 그 사실을 검사하지 못한다**. `verification-completeness.md` §1.1이 정의하는 report-not-verdict —
붉어질 수 없는 검사가 통과하는 형태 — 이며, §5 Audit-verification이 요구하는 "두 형태를 직접
돌려 발산을 관측하라"가 정확히 이 결함을 잡았다. RED 셀의 진실성을 강제하겠다는 SPEC의
`acceptance.md` 안에서 발생했다는 점에서 자기모순이다.

**Required fix**: 명령을 표 셀 밖(코드펜스)으로 빼거나, 파이프를 쓰지 않는 형태로 다시 쓴다.
예: `grep -c -e tense -e mood -e counterfactual -e "future.sense" <file1> <file2>`. 그리고 이
criterion을 채택하기 전에 뮤테이션(판별어 1개 심기)으로 red를 한 번 관측하고 그 출력을 셀에 적는다.

---

**D2. AC-RNT-012의 명령이 원문 그대로는 실행되지 않는다** — `acceptance.md:33` — Severity: **major** — Class: **blocking**

`\| wc -l`을 셸에서 원문 그대로 실행하면 `\|`가 파이프 연산자가 아니라 리터럴 인자가 되어
`grep`이 `|`, `wc`, `-l`을 파일 경로로 읽는다. 관측:

```
$ grep -rl "red_now" internal/template/templates/ \| wc -l
ugrep: warning: |: No such file or directory
ugrep: warning: wc: No such file or directory
(exit 2, 출력 없음)
```

셀은 `→ 0`을 주장하지만 원문 명령은 `0`을 내지 않는다. 백슬래시를 사람이 조용히 떼어낸 뒤에야
`0`이 나온다(그 형태로 확인함: `0`, exit 0). `acceptance.md:8-11`이 "the command, its verbatim
output"을 요구하는데, **어느 사본이 verbatim인지**(raw 파일인가, GFM 렌더 결과인가) 정의되어
있지 않다. 이 리포지토리에는 GFM 표 셀의 파이프 이스케이프와 `grep` 교대 의미가 상호작용해
공허한 GREEN 1건과 부당한 FAIL 1건을 만든 기록이 있으며, D1·D2가 그 두 방향을 각각 재현했다.

**Required fix**: 표 셀 안에 셸 파이프를 담지 않는다. 명령을 셀 밖 코드펜스로 옮기고 셀은 그
앵커만 참조하게 하거나, 파이프 없는 형태(`grep -rlc` 대신 `grep -rl … ; echo $?` 등)로 다시
쓴다. 아울러 REQ-RNT-001에 "verbatim은 raw 파일 기준"임을 명시한다.

---

**D3. AC-RNT-011의 RED 셀이 존재하지 않는 대상을 "보존 중"이라고 서술한다** — `acceptance.md:32` — Severity: **minor** — Class: **blocking**

셀은 "Neutrality holds today and is being preserved"라고 적는다. 그러나 MP-8 절은 아직 존재하지
않으므로(`grep -c "MOAI-REDNOW-BEGIN" .claude/agents/moai/plan-auditor.md` → `0`) 보존되고 있는
속성이 없다. 이것은 AC-RNT-012와 정확히 같은 형태의 forward-looking guard인데, AC-RNT-012의
셀은 그 사실을 스스로 밝힌다("It is green today only because the artifact does not exist yet").
형제 셀 두 개가 같은 구조를 서로 다른 정직도로 서술한다. 정직한 쪽은 AC-RNT-012이다.

**Required fix**: AC-RNT-011의 셀 문구를 AC-RNT-012의 형태로 맞춘다 — "아직 대상이 없어서 green이며,
그래서 release gate가 아니다".

---

**D4. MP-8의 범위(release-blocking 한정)가 명령 검증 자체의 탈출구가 된다** — `spec.md:139` (REQ-RNT-004) — Severity: **major** — Class: **blocking**

REQ-RNT-003은 재실행 불가한 RED를 가진 criterion을 regression-guard로 **강등**한다. 강등은
release-blocking 자격만 잃게 하고, MP-8은 release-blocking에만 걸리므로, **강등된 criterion의
명령은 그 뒤로 아무도 검증하지 않는다**. 이 SPEC 자신이 그 증거다: 결함 있는 명령 두 개(D1·D2)가
정확히 regression-guard 셀에 있고, MP-8은 문언 그대로 걸어도 둘 다 놓친다. §D.5의 자기적용
주장이 참이면서 공허해지는 원인이 여기다.

강등 자체는 정직했다(위 표 참조). 문제는 강등의 **결과**다: RED 의무는 면제되어야 마땅하지만
**명령 문법의 유효성 검사까지 면제될 이유는 없다**.

**Required fix**: L1(구조 검사)의 적용 범위를 release-blocking에 한정하지 말고, RED/green 셀에
명령이 적혀 있으면 class와 무관하게 그 명령이 단일 호출로 실행 가능한지(=파이프/이스케이프
파손 없음)를 검사하도록 REQ-RNT-008의 form contract에 한 줄을 추가한다. MP-8(L2 재실행)의
범위는 release-blocking 그대로 두어도 된다 — 바꿔야 하는 것은 L1이다.

---

**D5. green path 5건이 명령이 판정하지 못하는 산문 접속절에 기대고 있다 (token-insertion mutant)** — `acceptance.md:21,23,24,26,27` — Severity: **major** — Class: **blocking**

다음 5개는 green path가 `grep -c … ≥ 1` **AND** 산문 조건의 형태다. `grep ≥ 1`은 파일 어디에나
토큰을 써 넣으면 만족되며 — 주석이든 목차든 — 행위를 전혀 바꾸지 않는다. rule §5
Audit-verification: "grep을 만족시키는 가장 값싼 수정은 결함을 고치는 수정이 아니다".

| AC | 기계적 절반 | 토큰 삽입만으로 만족되는가 | 나머지 절반(명령이 판정 못 함) |
|----|-------------|---------------------------|-------------------------------|
| AC-RNT-001 | `grep -c "RED-now cell content"` ≥1 | **예** | "the clause enumerates all three elements" |
| AC-RNT-003 | `grep -c "regression-guard"` ≥1 | **예** | "the clause states demotion, not pass" |
| AC-RNT-004 | `grep -c "MP-8"` ≥1 | **예** | "the MP-8 body names re-execution against the current tree" |
| AC-RNT-006 | `grep -c "MP-8"` ≥1 | **예** | "contains the `N/A` branch and the stated-reason obligation" |
| AC-RNT-007 | `grep -c "AC-6:"` ≥1 | **예** | "a `- [PASS/FAIL/N/A] MP-8` row exists" |

AC-RNT-007이 가장 얕다: `AC-6:` 라는 여섯 글자를 아무 데나 쓰면 기계적 절반이 통과한다.
`acceptance.md` §D.2의 mutant probe는 AC-RNT-008 **하나에만** 돌렸고, 실제로 얕은 다섯 개에는
돌리지 않았다. 나머지 5건(AC-005·008·009a·009b·010·012)은 green path가 실제 테스트 실행/파일
생성/스팬 동등 비교라 이 mutant에 면역이다.

**Required fix**: 다섯 개의 산문 접속절을 M3 Go 테스트가 판정하는 기계적 술어로 옮긴다 — MP-8
스팬을 추출한 뒤 세 요소 토큰(command / output / SHA), `N/A` 분기, `severity=critical`,
점수 독립 문구, 보고 행 패턴의 존재를 스팬 **안에서** 단언한다. 스팬 추출 seam은 이미
REQ-RNT-008이 만들고 있으므로 새 기구가 필요 없다.

---

**D6. MP-8이 저자 통제 하의 임의 명령을 실행하도록 지시하는데, 읽기전용 제약도 실패 처리도 없다** — `spec.md:139` (REQ-RNT-004), `spec.md:116-118` — Severity: **major** — Class: **blocking**

`spec.md:118`은 "plan-auditor already carries `Bash` in its `tools:` frontmatter, so no capability
is added"라고 적는다. 도구에 관해서는 참이다(확인: `plan-auditor.md:8`의 `tools:`에 `Bash` 존재).
그러나 **위험면에 관해서는 거짓이다**. 오늘 auditor가 돌리는 Bash는 auditor 자신이 작성한 감사
배치이고, MP-8이 돌리라는 것은 **SPEC 저자가 표 셀에 써 넣은 임의 문자열**이다. 이 둘은 같은
도구를 쓰지만 신뢰 경계가 반대다.

MP-8은 이 중 어느 것도 정하지 않는다: (a) 명령이 읽기전용이어야 하는가, (b) 명령이 실패·행·
파괴적으로 동작할 때 auditor가 무엇을 하는가, (c) 재실행이 이 리포지토리의 [HARD] 규율과
충돌할 때 — 예컨대 셀이 `go test ./...`를 인용하면 CLAUDE.local.md §4 / `gitflow-lane-protocol.md`
§8의 "로컬 전체 스위트 금지"와 정면 충돌한다 — 무엇이 우선하는가.

`plan.md` §G는 `-run <TestName>`을 **RED 증거로 쓰는 것**을 금지하지만, 셀이 테스트 명령을
인용하는 것 자체는 막지 않는다. 그러면 MP-8이 그것을 돌려야 한다.

**부수 판정 — 배차문의 공격 지점 1에 대한 답**: plan-auditor의 frontmatter `description`이 적은
`NOT for: … running tests`는 **라우팅 절**이다 — "테스트를 돌려달라는 의도를 이 에이전트로
보내지 말라"는 뜻이지, 감사 중 명령 실행을 금지하는 능력 제약이 아니다. 근거: 같은 파일이
§Verification Execution Mandate와 Group A~D 검증 배치에서 `Bash`·`awk`·`git`·`grep` 실행을
명시적으로 지시하고, `mcp__moai__spec_audit`을 호출한다. 나 자신이 이 감사에서 `go test -count=1
./internal/kanban -run TestMigrationParity`를 돌렸고 그것이 역할 위반이라고 보지 않는다.

따라서 **MP-8은 원칙적으로 나에게 판정 가능하다** — 실제로 이번에 10/10 판정했다. 결함은
"할 수 없다"가 아니라 **"일반적인 경우에 무엇을 해야 하는지 쓰여 있지 않다"** 이다. 외교적으로
말하지 않겠다: 지금 문언대로면 MP-8은 신뢰할 수 없는 입력을 실행하라는 지시이고, 그 사실이
SPEC 어디에도 — 요구에도, 제외에도, anti-pattern에도 — 적혀 있지 않다.

**Required fix**: REQ-RNT-001의 "the command"를 **단일 호출로 끝나는 읽기전용 셸 invocation**으로
한정하고, 그 조건을 만족하지 않는 인용은 REQ-RNT-003의 강등 경로로 보낸다(auditor에게 실행
의무를 지우지 않는다). MP-8 본문에는 실행 실패/타임아웃 시의 처분(= 강등, pass 아님)을 명시한다.

---

**D7. §D.3 continued-firing이 rule §1.3이 묻는 질문에 답하지 않는다** — `acceptance.md:76-91` — Severity: **major** — Class: **blocking**

§D.3이 드는 근거 세 가지 — (1) 테스트가 계약을 restate하지 않고 extract한다, (2) 비교 전에
exactly-one-non-empty를 단언한다, (3) AC-RNT-010이 두 운반체의 스팬을 비교한다 — 는 전부
**계약이 바뀌었을 때**를 다룬다. rule §1.3이 묻는 것은 다른 질문이다: **"검사가 아예 돌지 않게
되었을 때 독자가 어떻게 아는가."** §1.3은 그 구분을 못박는다 — "absent execution, not suppressed
failure … nothing failed, and there was nothing there to fail."

두 산출물 각각에 §1.3의 시험("내일 이게 멈추면 내가 보는 것 중 무엇이 달라지는가")을 걸면:

- **Go 테스트**: 파일이 지워지거나 빌드 태그가 붙거나 셀렉터가 어긋나면 — 달라지는 것 **없음**.
  CI는 여전히 초록이다.
- **MP-8**: 더 나쁘다. MP-8은 마크다운 에이전트 파일 안의 지시문이다. 나중에 plan-auditor가
  재작성되며 MP-8 행이 빠지거나, auditor가 그냥 돌리지 않으면, 보고서는 MP-1~MP-7으로 여전히
  PASS를 낸다. 유일한 liveness 신호는 보고서의 `- [PASS/FAIL/N/A] MP-8` 행인데, **그 행의 부재를
  검사하는 것이 아무것도 없다**. 비실행과 성공이 구별되지 않는다 — §1.3이 정의하는 결함 그 자체.

즉 이 SPEC은 §1.1(관측된 실패)과 §1.2(WHEN/INPUT/reachability)까지는 자기 적용을 해냈지만
§1.3에서는 하지 않았다. 그리고 §1.3을 커버하는 REQ도 AC도 없다(traceability 표 12/13행 전수 확인).

**Required fix**: MP-8의 continued-firing 답을 정하고 요구로 승격한다. 가장 값싼 형태는 D5의
수정과 같은 기구를 재사용하는 것 — M3 Go 테스트가 `plan-auditor.md`에서 MP-8 스팬 **과** 보고
템플릿의 MP-8 행을 둘 다 단언하게 하면, MP-8이 문서에서 사라지는 순간 CI가 붉어진다. Go 테스트
자신의 비실행 감지는 별도 문제이며, 그렇다면 §D.3에 "이 축은 커버하지 않는다"고 명시적으로
적는 것이 지금처럼 답한 것처럼 보이는 것보다 낫다.

---

**D8. REQ-RNT-011이 트리 핀 없는 줄번호를, 이 작업이 곧 바꿀 파일에 대해 요구문 안에 박았다** — `spec.md:153` — Severity: **minor** — Class: **blocking**

"the mirror already carries two illustrative `SPEC-AUTH-001` placeholders at **lines 446 and 464**"
— 관측상 정확하다(`grep -nE "SPEC-[A-Z]+-[0-9]{3}" internal/template/templates/.claude/agents/moai/plan-auditor.md`
→ `446:`, `464:`). 그러나 두 문제가 있다:

1. **트리 핀 부재.** rule §4 [HARD]는 불변 단언이 측정 트리 SHA를 못박기를 요구한다.
   `acceptance.md:32`의 같은 인용은 "measured on `a6bbbf82b`"를 달았는데 `spec.md:153`은 달지
   않았다. 두 사본이 갈렸다.
2. **자기 무효화.** M4는 **바로 이 파일에** MP-8 절을 삽입한다. 삽입 지점이 446행보다 위라면
   두 줄번호는 SPEC이 닫히기도 전에 틀린다. 이 SPEC이 닫힐 시점에 이 요구문은 거짓을 담고 있다.

**Required fix**: 줄번호를 요구문에서 빼고 "미러가 이미 담고 있는 illustrative placeholder"라는
성질만 남긴다. 정확한 위치가 필요하면 `acceptance.md`의 증거 셀에 SHA 핀과 함께 두고, 요구문은
그것을 참조한다.

---

**D9. `module: spec-audit`가 스키마의 path-like 요건에 어긋난다** — `spec.md:11` — Severity: **minor** — Class: **optional**

`spec-frontmatter-schema.md` § Field Reference는 `module`을 "non-empty, **path-like**"로 정의하고
예시로 `internal/auth`를 든다. `spec-audit`은 경로가 아니다. `FrontmatterSchemaRule`은 비어있음만
검사하므로 lint는 통과하며(`mcp__moai__spec_audit` → drift finding은 `EraAutoDetected` INFO 1건뿐),
실제 편집 대상은 `.claude/rules/…`와 `.claude/agents/…`와 `internal/spec/`이다.

**Required fix**: `module: "internal/spec"` 로 바꾸거나, 실제 주 편집 표면을 반영하는 경로로 바꾼다.

---

**D10. `plan.md` §F M1의 미해결 결정 2건이 `[NEEDS CLARIFICATION]` 마커 없이 산문으로만 있다** — `plan.md:94-102` — Severity: **minor** — Class: **optional**

plan.md는 두 결정을 "surfaced for review rather than settled here", "asks review to confirm"이라고
정직하게 밝힌다. 그러나 이 결정들을 kickoff 게이트 전에 `AskUserQuestion`으로 라우팅하는 기구는
`[NEEDS CLARIFICATION: <topic>]` 마커이고, 그것이 쓰이지 않았다. MP-7은 기계적으로 통과한다
(grep 매치 없음) — 즉 **MP-7 자신이 §1.3 결함을 갖고 있다: 마커를 안 쓰면 게이트가 발화하지
않으며, 그 비발화는 성공과 구별되지 않는다.** 이것은 이 SPEC의 결함이라기보다 auditor 쪽 관측이라
optional로 분류한다.

**Required fix**: 두 결정에 마커를 달아 kickoff 전 해소를 강제하거나, 아래 §Recommendation의
판단을 채택해 결정을 닫는다.

---

## Recommendation

**Verdict: FAIL.** 두 축이 각각 독립적으로 FAIL을 만든다.

1. **집계 0.75 < Tier M 임계 0.80.** Testability 0.50이 지배적이며, 그 근거는 D1·D2·D5다.
2. **blocking 결함 8건**(D1~D8), 그 중 critical 1건(D1)이 SPEC 자신의 acceptance table 안에 있다.

MP-1~MP-7 firewall은 전부 통과했고, MP-8을 문언대로 자기 적용해도 10/10 통과한다. **즉 이 FAIL은
must-pass 위반이 아니라 rubric과 blocking 결함에서 나온다.** MP-8이 잡지 못한다는 사실 자체가
D4의 근거이며, 그것이 이번 감사에서 가장 값진 발견이다.

### 수정 순서 (되돌리기 쉬운 것부터 — plan.md §F의 정렬 원칙을 그대로 따름)

1. **D1** — AC-RNT-002 명령 재작성 + 뮤테이션으로 red 1회 관측 후 그 출력을 셀에 기록. (critical)
2. **D2** — AC-RNT-012 명령을 표 셀 밖으로. 아울러 REQ-RNT-001에 "verbatim = raw 파일 기준" 명시.
3. **D4** — REQ-RNT-008의 form contract에 "class 무관, 명령이 적혀 있으면 단일 호출 실행 가능성
   검사" 한 줄 추가. L1의 범위만 넓히고 MP-8(L2)의 범위는 건드리지 않는다.
4. **D6** — REQ-RNT-001에 읽기전용·단일호출 한정을 넣고, 미충족 인용은 REQ-RNT-003 강등 경로로.
5. **D5** — 다섯 개 green path의 산문 접속절을 M3 테스트의 스팬 내 술어로 이관.
6. **D7** — MP-8의 continued-firing 답을 요구로 승격(테스트가 보고 템플릿의 MP-8 행까지 단언).
7. **D3, D8, D9, D10** — 문구·필드 정정.

### 열린 결정 2건에 대한 판단 (`plan.md` §F M1)

**(a) 무엇이 "the command"인가.** — **읽기전용이며 단일 호출로 끝나는 셸 invocation으로 한정하라.**
plan.md는 비-명령(CI job, 수동 관측)을 강등으로 보내는 "deliberately permissive landing"을
제안하는데, 처분이 관대한 것과 **형식이 느슨한 것**은 다르다. MP-8이 그 문자열을 실행하기 때문에
형식은 오히려 조여야 한다(D6). 구체적으로: 파이프·리다이렉션·`&&`·서브셸을 배제하면 D1·D2가
구조적으로 재발할 수 없고, 부작용 있는 명령이 auditor 손에 들어오지 않으며, 스팬 추출도 단순해진다.
배제된 인용은 오류가 아니라 regression-guard 강등 — plan.md의 관대한 착지는 여기서 그대로 유효하다.

**(b) 문서 수준 트리 핀이 셀별 핀을 대신할 수 있는가.** — **가능하다. 다만 plan.md가 제안한
조건은 약하다.** 제안된 조건은 "the cell does not itself contradict it"인데, 침묵하는 셀은
모순하지 않으며 **침묵이 일반적인 경우**다. 즉 이 조건은 실질적으로 아무것도 배제하지 않는다.
더 강하면서 여전히 값싼 형태:

- 문서 수준 핀은 **자기 핀을 갖지 않은 모든 셀을 구속한다**(암묵이 아니라 명시적 승계).
- 핀은 **브랜치 이름이 아니라 SHA**여야 한다(rule §4 — moving ref는 측정과 판독 사이에 스스로를
  거짓으로 만든다).
- 셀이 자기 핀을 가지면 그것이 우선하며, 문서 핀과 다를 때는 셀 핀이 이긴다.

주목할 점: **이 SPEC의 `acceptance.md`는 이미 자신이 제안하는 것보다 엄격하게 행동한다** —
헤더(`acceptance.md:8-11`)가 `a6bbbf82b`를 못박고, 각 셀이 "on `a6bbbf82b`"를 다시 적는다.
요구문은 관행보다 낮게 쓰지 말고 관행에 맞춰 쓰는 것이 옳다. 반례는 `spec.md:153`(D8)뿐이며,
그것이 정확히 이 규칙이 막아야 할 형태다.

### 다음 반복

Tier M 상한은 2회다. iter2는 위 D1~D8의 델타에 한정하고, 각 항목에 대해 해소 여부를 회귀
검사한다. iter2에서 FAIL이면 상한에 도달하므로 PASS-with-debt / 범위 축소 / 사용자 override의
3택으로 운영자에게 에스컬레이션한다.

---

## Gaps — 이 감사에서 관측하지 **않은** 것

- **`.moai/reports/t343/red-now-premeasurement.md`를 열지 않았다.** 존재는 확인했으나(8,578바이트,
  `ls`), 그 내용은 저자 측 추론 산출물이므로 M1 Context Isolation에 따라 감사 입력에서 제외했다.
  그 안의 주장은 이 보고서의 어떤 판정도 뒷받침하지 않는다.
- **`SPEC-TODO-SQLITE-001`의 나머지 6개 "Red via missing test." 셀은 개별 재실행하지 않았다.**
  개수(7)만 측정했다. SPEC 자신이 그 SPEC을 범위 밖으로 선언했다(`spec.md:184-185`).
- **706개 SPEC 디렉터리 전수 sweep을 하지 않았다.** plan.md §A가 이 gap을 이미 기록하고 있으며,
  이번 감사도 그것을 메우지 않았다.
- **`0.94` 감사 점수를 재도출하지 않았다.** SPEC 자신이 "quoted from the card dispatch, not
  re-derived"라고 밝히며(`spec.md:76-80`) 어떤 절도 그 값에 기대지 않는다 — 이 정직한 표기는
  이 SPEC의 강점이며 감점 사유가 아니다.
- **`.moai/reports/t350/discovery.md`**는 이 워크트리에서 해소되지 않는다(plan.md:158-161이
  이미 그렇게 기록). 재확인하지 않았다.
- **M2/M3/M4가 실제로 산출할 코드는 존재하지 않는다.** plan-phase 감사이므로 미작성 산출물에
  대한 판정은 하지 않았다.

## Residual risk

- D1의 뮤테이션은 판별어 `mood` **하나**로 관측했다. 나머지 세 판별어에 대해 개별 발산을 관측하지
  않았으나, ERE의 `\|` 의미는 패턴 전체에 균일하게 적용되므로 결론은 바뀌지 않는다.
- `grep`은 이 환경에서 `ugrep`으로 해석된다(D2의 경고 문구가 `ugrep:` 접두를 달았다). GNU grep에서
  같은 명령이 다른 종료 코드를 낼 여지가 있다 — 다만 `\|`의 리터럴 해석은 POSIX ERE 공통이므로
  D1의 공허성 판정은 구현에 의존하지 않는다. D2의 exit code(2)는 구현 의존일 수 있다.
- 집계 점수는 4개 차원의 산술평균이다. 조화평균을 쓰면(Testability 0.50이 더 강하게 지배) 더 낮게
  나오며, 어느 쪽이든 0.80을 넘지 않는다.
