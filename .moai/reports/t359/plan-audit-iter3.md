# SPEC Review Report: SPEC-TODO-LANDING-EVIDENCE-001 — iteration 3 (최종)

Card: **t359** · Iteration: **3/3 (마지막 허용 회차)** · Tier **L** (PASS 임계 **0.85**)
Tree: worktree `.claude/worktrees/t359`, branch `WT-landing-evidence`, HEAD `c0cfb2520`, 작업 트리 clean
Subject: SPEC **v0.3.0** (`c0cfb2520`), iter-2 부채 E1-E7 종결 커밋
Delta: `git diff 8a1ae5b70..c0cfb2520` — 5개 SPEC 아티팩트만 변경, 소스 파일 0건
Auditor: plan-auditor (Claude 단독; cross-model 백엔드 미조회)

**Verdict: PASS-WITH-DEBT**
**Overall Score: 0.92** (Tier L 임계 0.85) — 궤적 **0.80 → 0.89 → 0.92**, 회귀 없음, STOP 신호 없음
**E1-E7 전부 종결** (E1·E2·E3 blocking 3건 포함) · 델타가 심은 새 결함 **3건 (전부 minor)** + optional 1건
**Implementation Kickoff Approval 게이트로 진행하기에 적합함** — 잔여 부채는 아래 §부채 이관에 명시

Reasoning context ignored per M1 Context Isolation. 저를 부른 에이전트가 넘긴 E4 측정치는 판정 근거로
쓰지 않고 가설로만 다뤘습니다 — 아래 §E4 검증에서 직접 재현했고, **한쪽은 확인, 한쪽은 반증**했습니다.

Retry Loop Contract에 따라 이번 회차는 **열거된 델타에 한정한 확인 감사(confirming delta re-audit)**입니다.
구조적 must-pass는 수치가 바뀔 수 있으므로 전수 재측정했습니다.

---

## Must-Pass Results (7/7 통과)

- **[PASS] MP-1 REQ 번호 일관성.** `grep -o 'REQ-TLE-[0-9]*' spec.md | sort -u` → `REQ-TLE-001`…`021`,
  21개, 순차·중복 없음·3자리 패딩 균일. `grep -c '^- \*\*REQ-TLE-'` = **21** — 모든 id가 정의부를 가짐.
  델타는 요구사항을 추가하지도 삭제하지도 않았습니다.
- **[PASS] MP-2 GEARS 준수 — 요구사항 계층에 한정.** spec.md diff의 hunk는 frontmatter / HISTORY /
  §A.3b 출력 블록 / §G 2곳뿐이며, **§C의 REQ 본문은 한 줄도 바뀌지 않았습니다**. 신규 2건은 iter-2에서
  검증된 그대로: REQ-TLE-020(`spec.md:463`)은 `when` 두 절이 이어진 Event-driven 복합형,
  REQ-TLE-021(`:499`)은 Ubiquitous. `should`/`may`/`IF-THEN` 없음.
  이 판정은 **요구사항 계층(`REQ-TLE-*`)에 대해 내린 것**이며, `AC-TLE-*`의 Given-When-Then은
  검증 계층의 정규 형식이므로 여기서 감점하지 않았습니다(M3 §Scope).
- **[PASS] MP-3 YAML frontmatter.** 12개 정본 필드 전부 존재·타입 정상. `version: "0.3.0"` 인용된
  semver, `status: draft`, `created`/`updated` ISO, `tier: L`.
  `grep -nE '^(created_at|updated_at|labels|spec_id):' spec.md` → `rc=1` — 거부 별칭 없음.
- **[N/A] MP-4 언어 중립성.** 단일 언어(Go + 마크다운 독트린 2본) SPEC. 다국어 도구 주장 없음 → 자동 통과.
- **[PASS] MP-5 D7 교차-SPEC 조정.** 본문에서 추출한 6개 참조 SPEC 전부 `.moai/specs/`에 실존하고,
  측정된 status는 `in-progress` ×2(`SPEC-KANBAN-QUEUE-PR-SYNC-001`, `SPEC-KANBAN-TODO-CLI-001`),
  `completed` ×4(`SPEC-TODO-ANALYSIS-001`, `SPEC-TODO-ARCHIVE-QUERY-001`,
  `SPEC-TODO-DESTRUCTIVE-GUARD-001`, `SPEC-TODO-LANDING-STATE-001`).
  `retired`/`superseded`/`archived` 없음 → D7 BLOCKING 없음.
- **[PASS] MP-6 D8 크로스플랫폼.** 5개 아티팩트 전부 `grep -c 'syscall'` = **0** → 자동 PASS.
- **[PASS] MP-7 clarification 게이트.** `grep -rn '\[NEEDS CLARIFICATION'` 아티팩트 전역 → `rc=1`.

7개 모두 통과. 델타는 firewall을 건드리지 않았습니다.

---

## 1. E1-E7 회귀 점검 — 겉만 닫혔는지 실제로 닫혔는지

iter-2의 잔여 위험 경고("E1의 수정은 안 했는데 한 것처럼 보이기 쉬울 만큼 작다")를 전제로,
**문구 교체가 아니라 실패 가능성이 실제로 생겼는지**를 기준으로 판정했습니다.

| # | iter-2 분류 | 판정 | 근거 |
|---|---|---|---|
| **E1** | major / blocking | **진짜 종결** (잔여 F1) | 아래 §1.1 |
| **E2** | major / blocking | **절반은 진짜, 절반은 공허** (F2) | 아래 §1.2 |
| **E3** | minor / blocking | **진짜 종결** (잔여 F4) | 아래 §1.3 |
| **E4** | minor / non-blocking | **한 결함을 다른 결함으로 교체** (F3) | 아래 §2 |
| **E5** | minor / non-blocking | **깨끗이 종결** | `plan.md` §B가 1,2,3,4,5 순으로 재번호됨. diff에서 git-subprocess 항목이 5→4, `--json` 항목이 4→5로 교차 확인 |
| **E6** | minor / non-blocking | **깨끗이 종결, 요구된 것보다 강하게** | 아래 §1.4 |
| **E7** | minor / non-blocking | **깨끗이 종결** | 아래 §1.5 |

### 1.1 E1 — 실패할 수 없던 절이 실패할 수 있게 됐는가: 예

**Claim.** 귀속 경계 절은 이제 올바른 변수(카드 **id**)를 변화시키며, iter-2가 경고한 "문구만 바꾼 종결"이
아닙니다. 다만 명시된 RED이 무조건 성립하지는 않습니다(F1).

**Evidence.** `acceptance.md` AC-TLE-020 신규 본문:

> **And the attribution boundary holds**: the card id reaches neither check. Asserted by recording the
> **same `--sha` against two different card ids** — two cards created in the same fixture queue, both
> records attempted with an identical SHA — and observing an identical accept/reject outcome and an
> identical stderr classification for both. Run for the accepting condition (a) and for at least one
> refusing condition (b or c).

iter-2가 잔여 위험으로 지목한 두 가지 — ① 픽스처가 **서로 다른 카드 id 두 개**를 쓰는가, ② 두 호출이
**같은 `--sha`**를 넘기는가 — 가 문장 안에 명시돼 있고, 여기에 iter-2가 요구하지 않았던 것까지 더했습니다:
승인 분기와 거절 분기 **양쪽에서** 불변성을 주장하도록 한 것입니다. `plan.md` M3 항목 6에도 같은 문장이
반영돼 구현자가 읽는 자리에 놓였습니다.

근거로 인용된 코드 좌표도 직접 확인했습니다:

```
$ sed -n '96,110p' internal/kanban/prlink_landed.go
func LandedGrepArgs(ref, cardID string) ([]string, error) {
	...
	return []string{
		"log", ref,
		LandedRegexpEngineFlag,
		`--grep=\b` + cardID + `\b`,
		"--oneline",
	}, nil
}

$ sed -n '148,152p' internal/cli/todo.go   # 17개 verb 등록
	cmd.AddCommand(newTodoAddCmd(), newTodoListCmd(), newTodoDoneCmd(), newTodoUndoneCmd(), newTodoNextCmd(),
		newTodoUnpickCmd(), newTodoEditCmd(), newTodoMoveCmd(),
		newTodoDropCmd(), newTodoUndropCmd(),
		newTodoAnalyzeCmd(), newTodoRelateCmd(), newTodoUnrelateCmd(), newTodoWhyCmd(),
		newTodoPRCmd(), newTodoExportJSONCmd(), newTodoHistoryCmd())
```

술어가 카드 **id**를 키로 삼는다는 것, 등록된 어떤 verb도 id를 바꾸지 않는다는 것 둘 다 사실입니다.
(부기: iter-2는 verb를 16개로 셌으나 실제 등록은 **17개**입니다 — `unpick`이 빠져 있었습니다. SPEC은
개수를 인용하지 않고 좌표만 인용하므로 SPEC의 결함은 아닙니다.)

**Baseline-attribution.** HEAD `c0cfb2520`, worktree `.claude/worktrees/t359`, 이번 실행에서 측정.

### 1.2 E2 — 접합절 (c): 앞 절은 진짜 검출기, 뒤 절은 근거 없는 주장

**Claim.** 절 (c)의 **첫 번째 연언**("frozen DDL 문자열이 live `backlogDDL` 상수와 바이트 단위로 같다")은
E2가 요구한 그대로의 독립 RED을 가집니다. **두 번째 연언**("frozen switch의 accepted-version 집합이
live의 것과 같다")은 **기계적 실행 형태가 명시돼 있지 않고, 가장 값싼 구현으로는 명명한 드리프트를
검출하지 못합니다.** 이것이 F2입니다.

**Evidence.** 앞 절: `backlogDDL`은 같은 패키지의 const이므로 패키지 내부 테스트에서 직접 비교할 수 있고
(`internal/kanban/backlog_sqlite.go:102`), 명시된 RED도 정확합니다 — "live const를 편집하고 frozen 사본은
두면 (a)(b)는 green이고 (c)만 fail". 실행 가능하고 실패 가능합니다.

뒤 절의 대상인 live switch를 읽으면:

```
$ sed -n '286,298p' internal/kanban/backlog_sqlite.go
	switch version {
	case "":
		... INSERT INTO meta(...) VALUES (?, ?) ... backlogSchemaVersion ...
	case backlogSchemaVersion:
		// current layout
	default:
		return fmt.Errorf("... unsupported schema_version %q (want %q): %w",
			e.dbPath, version, backlogSchemaVersion, ErrBacklogCorrupt)
	}
```

"accepted-version 집합"은 **자료구조가 아니라 제어 흐름**입니다. 테스트가 live 코드에서 이 집합을
꺼낼 수단은 없습니다. 실행 가능한 가장 값싼 해석은 `backlogSchemaVersion` 상수를 frozen 사본의 상수와
비교하는 것인데, 그러면 나중에 누군가 live switch에 `case "2":`를 **추가**해도 상수는 그대로이므로
아무것도 트립하지 않습니다 — 이 연언이 막겠다고 이름 붙인 바로 그 드리프트를 놓칩니다.

결정적 방증: 이 기준의 **RED 목록에 switch 절에 대한 RED이 없습니다.** 세 개의 RED은 각각 버전 범프,
`NOT NULL` 기본값 누락, `backlogDDL` 편집이며, switch 구조 변경을 red시키는 돌연변이는 하나도 없습니다.
저자 자신이 이 연언의 실패 조건을 적지 못한 것입니다.

**Baseline-attribution.** HEAD `c0cfb2520`; `backlog_sqlite.go`는 이번 델타에서 수정되지 않았음
(diffstat에 소스 파일 0건).

### 1.3 E3 — cannot-be-run 분기가 기준을 얻었는가: 예, 그리고 실행 가능한 형태로

**Claim.** 케이스 (d)가 추가돼 REQ-TLE-020의 미검증 분기가 닫혔고, 이 SPEC이 지정한 실행 형태는 이
코드베이스에서 실제로 구현 가능합니다.

**Evidence.** AC-TLE-020 Given: "(d) the checks **cannot be run** — exercised twice, once with `git`
absent from the resolved command runner and once with a `--ref` that resolves to no ref."

"resolved command runner"가 실재하는 이음매인지 확인했습니다:

```
$ sed -n '57,65p' internal/cli/todo_pr.go
var todoRunCommand kanban.CommandRunner = func(name string, args ...string) (string, error) { ... }

$ grep -rn 'todoRunCommand' internal/cli/ | head -3
internal/cli/todo_undone_test.go:32:	original := todoRunCommand
internal/cli/todo_pr_test.go:55:	prev := todoRunCommand
internal/cli/todo_pr.go:48:// todoRunCommand is the process seam every subprocess in the todo surface
```

패키지 수준 변수이고 기존 테스트 3곳이 이미 교체해 쓰고 있으므로, "git이 없는" 조건은 PATH를 건드리지
않고 테스트 안에서 만들 수 있습니다. 기준이 요구하는 픽스처가 실현 가능합니다.

또한 이 SPEC은 `todo pr`의 fail-open과 정반대인 이유를 명시했습니다 — "읽기는 답할 수 없어도 관대하게
남고, **쓰기**는 검증할 수 없으면 거부한다". 구현자가 이웃 파일의 습관을 옮겨오는 것을 막는 정확한 배치이고,
`plan.md` M3 항목 6에도 같은 대조가 실려 있습니다.

### 1.4 E6 — 요구된 것보다 강하게 닫힘

iter-2는 "AC-TLE-021이 스스로 렌더링해 세도록 명시하고 `(7)`을 주석으로 강등하라"고 했습니다. v0.3.0은
그대로 이행하면서 **왜** 그런지까지 적었습니다 — "a test cannot read another test's runtime value …
comparing prose against a hard-coded 7 would make this a doc-consistency check rather than the
prose-versus-behaviour tie REQ-TLE-021 asks for."

의심 하나를 직접 확인했다가 기각했습니다: 자체 렌더링한 행의 필드 수가 픽스처 카드 상태에 따라 달라지면
이 기준이 불안정해지는데, AC-TLE-015이 "**every row** splits into exactly 7 tab-separated fields;
… field 6 holds the evidence (**empty** for the cards without it)"라고 고정폭을 못박고 있어
어떤 픽스처에서도 값이 정해집니다. **결함 아님** — 확인 후 기각한 항목으로 남깁니다.

### 1.5 E7 — "nothing flags it"이 철회되고, 완화책이 설계 결정으로 승격됨

§G가 v0.2.0의 과장을 명시적으로 철회하고("**A mitigation exists and was considered — v0.2.0's
'nothing flags it' overstated the absence and is withdrawn**"), 커밋 subject 렌더링이 왜
비-귀속적인지, 왜 §D의 read-path 비용 배제와 충돌하지 않는 유일한 형태가 **record-time 포착**인지까지
적었습니다. 그리고 이것을 부채로 접지 않고 `progress.md`의 **Kickoff 게이트 결정 항목 3번**으로
올렸습니다 — 기록 형태를 바꾸는 일이므로 결함 수정이 아니라 설계 결정이라는 판단이며, 저도 동의합니다.

---

## 2. E4 검증 — 넘겨받은 가설을 직접 재현: 한쪽 확인, 한쪽 반증

**Claim (2가지로 분리).**
**(A) 인용된 원인 설명은 거짓이다 — 확인됨.**
**(B) 붙여넣어진 순서가 6회 실행에서 재현되지 않았다 — 반증됨. 2/6회에서 정확히 재현됐다.**

**Evidence.** v0.3.0이 `research.md` §R.10.1에 새로 붙인 문장:

> (Full output, verbatim including paths. `grep -m1` emits in the order it resolves the files, which
> is not the argument order given above — …)

이 셸의 `grep`이 무엇인지부터:

```
$ type grep
grep is a shell function from /Users/goos/.moai/claude-profiles/moai-adk/shell-snapshots/snapshot-zsh-1788301811836-olj9z9.sh
```

같은 인자로 `/usr/bin/grep`을 6회:

```
$ for i in 1..6; do /usr/bin/grep -m1 '^status:' <QUEUE-PR-SYNC> <TODO-CLI> <ANALYSIS> <LANDING-STATE>; done
--- 6회 모두 동일 ---
.moai/specs/SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md:status: in-progress
.moai/specs/SPEC-KANBAN-TODO-CLI-001/spec.md:status: in-progress
.moai/specs/SPEC-TODO-ANALYSIS-001/spec.md:status: completed
.moai/specs/SPEC-TODO-LANDING-STATE-001/spec.md:status: completed
```

**6/6 전부 인자 순서 그대로, 결정적입니다.** 따라서 "`grep -m1`은 파일을 해결하는 순서로 내보내며 그것은
인자 순서가 아니다"라는 SPEC의 서술은 `grep -m1`이 갖지 않은 성질을 그것에 귀속시킨 것입니다.
**(A) 확인.**

같은 인자로 셸 함수 `grep`을 6회 돌리면 **세 가지 서로 다른 순서**가 나옵니다:

- run 1, 3 → TODO-CLI, QUEUE-PR-SYNC, ANALYSIS, LANDING-STATE  ← **붙여넣어진 블록과 일치**
- run 2, 5, 6 → QUEUE-PR-SYNC, TODO-CLI, ANALYSIS, LANDING-STATE
- run 4 → TODO-CLI, ANALYSIS, QUEUE-PR-SYNC, LANDING-STATE

즉 붙여넣어진 블록은 **이 셸에서 실제로 나올 수 있는 출력이며 6회 중 2회 재현됐습니다**. "6회 중 한 번도
재현되지 않았다"는 넘겨받은 주장은 제 측정으로 **반증**됩니다. verbatim 주장 자체는 무너지지 않았습니다.
무너진 것은 **원인 설명**이고, 부수적으로 **재현성**입니다 — 병렬 grep 래퍼의 출력 순서가 비결정적이라
같은 명령을 다시 돌린 독자는 대개 다른 순서를 봅니다.

**쌍둥이 블록.** `spec.md` §A.3b(`:127-136`)도 같은 전체 경로로 재붙여넣기됐으나 **공개 문장이 없습니다.**
거짓 서술이 한 곳에만 있다는 뜻이라 손해는 작지만, 두 표면의 주석이 갈립니다.

**분류.** 실질(네 개의 status 값)은 다툼이 없고 제 측정과 일치합니다. 결함은 문서가 스스로 선언한
증거 규율("Where a claim rests on a command … the command and its verbatim output are recorded")
안에서 **검증되지 않은 도구 동작 주장을 사실로 단언한 것**입니다. **F3, minor, blocking** — 한 문장 수정.

**Baseline-attribution.** 전부 HEAD `c0cfb2520`, worktree `.claude/worktrees/t359`, 이번 실행.

---

## 3. 카드가 부과한 4개 제약 — 언급이 아니라 이행 여부

### D1 — 저장 방식이 `REQ-TODO-013`의 실제 텍스트 위에 놓였는가: 예

원문을 직접 읽었습니다:

```
$ sed -n '59p' .moai/specs/SPEC-KANBAN-TODO-CLI-001/spec.md
- **REQ-TODO-013** (Ubiquitous) The backlog store shall preserve the existing version-1 record shape
  — `{"version":1,"items":[{"id","text","added_at","spec_id","state"}]}` with
  `state ∈ {queued, picked, dropped}` — changing it only additively (the high-water mark, per REQ-TODO-009).
```

SPEC §A.3은 이 텍스트를 verbatim으로 싣고 "It does not freeze the field set. It constrains the
*manner* of change to additive, and names its own precedent for one."이라고 읽습니다 — 텍스트에 충실한
독해입니다. **결론만 옮긴 것이 아니라 논거를 확인했습니다**: SPEC은 카드가 화해시키라고 지시한
`SPEC-TODO-ANALYSIS-001`의 기록도 verbatim으로 싣고, 그 기록이 같은 독해("같은 REQ가 additively 변경을
허용한다 — `last_seq`가 그 선례")를 하고 있음을 보여 **카드의 전제 후반부를 반증**했습니다. 선례도
트리에 실재합니다(`internal/kanban/backlog_store.go:187-193`의 `LastSeq` top-level 필드).

한 가지 해석 단계가 명시되지 않은 채 수행됩니다: REQ-TODO-013은 **JSON 레코드 모양**으로 서술돼 있고
이 SPEC은 **SQLite 컬럼**을 추가합니다. "항목당 필드 집합의 가법적 확장"이라는 표현-중립적 독해가
필요한데, SPEC은 그 다리를 명시적으로 놓지 않습니다. 다만 카드 자신이 `ADD COLUMN`이라는 형태로 질문을
던졌으므로 운영자가 이미 그 매핑을 받아들인 셈입니다 — **결함으로 올리지 않고 잔여 위험으로 기록**합니다.

### D2 — REQ-1.10 뒤집기가 아니라 실제로 성립하는 구분인가: 예

인용문을 원본과 대조했습니다(`SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md:251-255`, verbatim 일치 확인).
SPEC §B.3의 처분: "**REQ-1.10 stands, unamended and unreversed.** No requirement in this SPEC weakens
it, and the resolver's behaviour is unchanged."

구분이 재라벨링인지 실제인지는 **무엇이 입력인가**로 갈립니다. resolver의 술어는 카드 id를 입력으로 받아
커밋을 고릅니다(`LandedGrepArgs(ref, cardID)` — §1.1의 측정). 운영자가 타이핑한 SHA를 검증하는 두 명령은
카드 id를 입력으로 받지 않습니다. REQ-1.10이 금하는 것은 "어느 커밋이 카드를 전달했는지 **기계가 추론해
주장하는 것**"이고, 저장되는 것은 운영자의 의도입니다 — 두 명제의 주어가 다릅니다. **구분은 성립합니다.**

그리고 SPEC은 이 구분이 공짜가 아님을 §G에 적습니다: 도달 가능하지만 틀린 SHA는 기계가 구별할 수 없고,
"telling them apart is exactly the card-to-commit attribution REQ-1.10 forbids the machine to attempt".
자기 논거의 대가를 스스로 계산한 서술이며, §D에 "Out of Scope — reversing or amending REQ-1.10"
H3까지 두어 후속 독자가 되돌리지 못하게 막았습니다.

### D3 [HARD] — 착지 필드는 증거이지 자동 상태 전이가 아닌가: 예, 기계적으로

§B.2: "**The verb records; it never transitions.** … REQ-TLE-008 makes it mechanical rather than
aspirational: the verb writes the `landing` column and touches no other column of any row."
§D에 전용 H3("Out of Scope — automatic state transitions") 3개 불릿, 그리고 AC-TLE-008이
전체 큐 불변성 + 심어진 돌연변이로 검증합니다(§E: REQ-TLE-008 → AC-TLE-008 → M3). 이행됨.

### D4 [HARD] — 네 번째 `state` 값 없음: 예, 트리의 근거와 함께

§B.1 "Why not a fourth `state` value"와 §D "Out of Scope — a fourth state value" 양쪽에 있고,
근거로 든 코드 주석을 원본에서 확인했습니다:

```
$ sed -n '95,97p' internal/kanban/backlog_sqlite.go
// pair of tables rather than a fourth `state` value. SQLite cannot ALTER a
// CHECK constraint, so admitting a fourth state would need a table rebuild on
// every operator queue in the field.
```

인용이 정확하고 REQ-TLE-004가 CHECK 불변을 요구사항으로 못박습니다. 이행됨.

---

## 4. 델타가 심은 새 결함 — 재발 3형태 전수 훑기

iter-2가 지목한 세 형태(실패할 수 없는 절 / 독립 RED 없는 단언 / 미검증 when-분기)로 신규 텍스트 전부를
훑었습니다. 결과: **실패할 수 없는 절 1건(F2), 조건부로만 성립하는 RED 1건(F1), 미검증 하위분기 1건(F4)**.

**F1 — AC-TLE-020 경계 절의 RED이 픽스처 성질에 조건부로만 성립.**
기준은 무조건적으로 적습니다: "**Feed the card token into either check** → the two-card outcomes
diverge and the boundary clause fails". 실제로는 픽스처 히스토리가 두 카드 id와 어떤 관계인지에 따라
갈립니다. 카드 토큰을 흘리는 구현을 세 경우로 나누면 — 히스토리가 두 id를 **모두 언급하지 않으면**
두 호출 모두 거절되어 결과가 **일치**하고 경계 절은 통과합니다(누출은 조건 (a)의 "exits 0 and stores"
쪽에서 잡힙니다). 두 id를 **모두 언급하면** 누출이 세 절 전부를 통과합니다. **정확히 하나만 언급할 때만**
경계 절이 발화합니다. Given은 이 성질을 못박지 않습니다.
기준 전체로 보면 현실적 픽스처(카드 id를 언급하는 커밋을 일부러 만들지 않는 경우)에서 누출은 (a)에서
잡히므로 **E1이 닫은 "전혀 실패할 수 없음"과는 다른, 훨씬 좁은 잔여**입니다.
**minor / blocking** — 수정은 Given 한 절: 픽스처 히스토리가 두 카드 id 중 **정확히 하나만** 언급하도록
구성한다고 못박고, 명시된 RED을 그 전제에 매답니다.

**F2 — AC-TLE-018 절 (c)의 두 번째 연언은 기계적 형태가 없고 RED도 없다.** §1.2의 측정 참조.
**minor / blocking** — 수정은 둘 중 하나: ① 행동적 형태를 명시한다 — 여러 버전(`""`, `"1"`, `"2"`,
쓰레기 값)으로 스탬프된 DB에 live `ensureSchema`를 돌려 accept/stamp/reject 분할이 frozen switch의
같은 입력에 대한 분할과 일치함을 주장 — 하거나, ② 두 번째 연언을 삭제하고 실제 RED을 가진
바이트 동일성 주장만 남긴다. **②가 값싸고 정직합니다.**

**F3 — `research.md` §R.10.1의 원인 서술이 거짓.** §2의 측정 참조.
**minor / blocking** — 수정은 한 문장: "`grep -m1` emits in the order it resolves the files"를
"이 셸의 `grep`은 병렬 grep을 감싼 셸 함수이며 출력 순서가 비결정적이다(`/usr/bin/grep`은 인자 순서대로
결정적으로 내보낸다). 아래는 그 중 한 번의 캡처다"로 교체. `spec.md` §A.3b 쌍둥이 블록도 같은 주석을
달거나, 양쪽 모두 주석 없이 두어 표면을 일치시킵니다.

**F4 — AC-TLE-020 케이스 (d)의 두 하위 조건이 stderr 한 바구니로 뭉침.** REQ-TLE-020은 "naming which
check failed"를 요구하는데, (d)는 "unrunnable for (d)"라는 단일 분류만 주장합니다. 그런데 두 하위
조건은 서로 다릅니다: git 부재는 **두 검사 모두** 실행 불가, 해결 불가 ref는 존재 검사는 통과하고
**도달성 검사만** 실행 불가입니다. (b)·(c)와는 구별되지만 (d) 내부는 구별되지 않습니다.
**minor / optional** — M6에 따라 운영자 재량으로 남깁니다. 수정한다면 (d)의 stderr가 실행 불가인 검사를
지목하도록 한 절 추가.

---

## Category Scores (rubric-anchored)

| 차원 | 점수 | Rubric band | 근거 | Δ (iter-2) |
|---|---|---|---|---|
| Clarity | 0.90 | 0.75 밴드를 크게 상회 (1.0은 어디에도 모호함이 없을 것을 요구) | E5(번호), E6(모호한 비교 대상) 둘 다 깨끗이 종결. §B.3의 provenance 논거와 §1.2의 (c) 도입 설명은 정확하고 자기 과장을 스스로 철회함. 잔여: F3(거짓 원인 서술 — 종류로는 이전 누락보다 나쁨), F2의 "accepted-version set"이라는 미정의 표현 | 0.00 |
| Completeness | 0.93 | 0.75 밴드를 크게 상회 | 21/21, 5개 아티팩트, `### Out of Scope —` H3 6개 전부 불릿 보유, §G가 forward drift와 E7 완화책 검토 기록을 추가하고 기존 잔여를 하나도 버리지 않음. E3의 미검증 분기 종결. 잔여: F4 | +0.03 |
| Testability | 0.90 | 0.75 밴드 상회 | E1(실패 불가 절)·E2(검출기 아닌 커버리지 주장)·E3(when 절 절반) 3건 모두 종결. AC-TLE-018 (c) 앞 절과 AC-TLE-020 (d)는 실제 RED과 실현 가능한 이음매를 가짐(측정 확인). 잔여: F2(RED 없는 연언), F1(조건부 RED) | +0.05 |
| Traceability | 0.95 | 0.75 밴드를 크게 상회 | §E 21↔21 양방향, 고아 AC 0건·무커버 REQ 0건, 마일스톤 셀 전부 채워짐(019는 a/b/c 표기). REQ-TLE-020의 cannot-be-run 분기가 (d)로 추적됨 — iter-2의 유일한 추적 잔여가 닫힘 | +0.05 |

**Aggregate (조화평균)**: 4 / (1/0.90 + 1/0.93 + 1/0.90 + 1/0.95) = 4 / 4.35012 = **0.9195 → 0.92**
(산술평균 0.92)

**0.92 ≥ 0.85** — Tier L 임계 충족.
**궤적 0.80 → 0.89 → 0.92.** 회귀 없음 → LEAN STOP-on-regression 조항 미발화, 범위 축소 질문 불필요.

---

## Defects Found (iteration 3)

**F1** — `.moai/specs/SPEC-TODO-LANDING-EVIDENCE-001/acceptance.md` AC-TLE-020 경계 절 + 그 RED —
"카드 토큰을 검사에 흘리면 두 카드의 결과가 갈리고 경계 절이 실패한다"는 RED이 무조건적으로 서술돼
있으나, 실제 발화는 픽스처 히스토리가 두 카드 id 중 정확히 하나만 언급할 때에만 성립함. 둘 다 언급하지
않으면 두 호출 모두 거절되어 결과가 일치하고(누출은 조건 (a)에서 잡힘), 둘 다 언급하면 누출이 세 절을
모두 통과함 — Severity: **minor** — Class: **blocking** — Required fix: Given에 "픽스처 ref의 히스토리는
두 카드 id 중 정확히 하나만 언급한다"를 못박고, RED 문장을 그 전제에 매단다.

**F2** — `acceptance.md` AC-TLE-018 절 (c) 두 번째 연언 ("the frozen switch's accepted-version set
equals the live one") — live의 accepted-version 집합은 자료구조가 아니라 `backlog_sqlite.go:286-298`의
제어 흐름이라 테스트가 꺼낼 수 없음. 실행 가능한 가장 값싼 해석(`backlogSchemaVersion` 상수 비교)은
live switch에 `case` 하나가 추가되는 드리프트를 검출하지 못함. 이 기준의 RED 목록에 switch 절을
red시키는 돌연변이가 하나도 없다는 사실이 이를 방증함. (첫 번째 연언 — DDL 바이트 동일성 — 은
정상이며 명시된 독립 RED을 가짐) — Severity: **minor** — Class: **blocking** — Required fix: 두 번째
연언을 삭제하고 DDL 바이트 동일성만 남기거나(권장, 값쌈), 행동적 형태를 명시한다 — 여러 스탬프 버전으로
live `ensureSchema`를 구동해 accept/stamp/reject 분할이 frozen switch의 것과 일치함을 주장.

**F3** — `research.md` §R.10.1 (신규 공개 문장) — "`grep -m1` emits in the order it resolves the
files, which is not the argument order given above"는 `grep -m1`이 갖지 않은 성질을 사실로 단언함.
`/usr/bin/grep -m1`은 6/6회 인자 순서대로 결정적으로 출력함(§2 측정). 실제 원인은 이 셸의 `grep`이
병렬 grep을 감싼 셸 함수이고 그 출력 순서가 비결정적인 것(6회에 서로 다른 순서 3종). 붙여넣어진 출력
자체는 이 셸에서 6회 중 2회 재현되므로 verbatim 주장은 무너지지 않음 — 무너진 것은 원인 서술과
재현성임. `spec.md` §A.3b 쌍둥이 블록은 전체 경로로 고쳐졌으나 공개 문장이 없어 두 표면의 주석이 갈림
— Severity: **minor** — Class: **blocking** — Required fix: 원인 문장을 셸 함수/병렬 grep으로 교체하고
"아래는 비결정적 출력의 한 캡처"임을 밝힌다. 쌍둥이 블록의 주석 유무를 둘 중 하나로 통일한다.

**F4** — `acceptance.md` AC-TLE-020 케이스 (d) — REQ-TLE-020은 "naming which check failed"를
요구하는데 (d)는 두 하위 조건(git 부재 = 두 검사 모두 불가 / 해결 불가 ref = 도달성 검사만 불가)을
"unrunnable" 한 바구니로 분류함. (b)·(c)와는 구별되나 (d) 내부는 구별되지 않음 — Severity: **minor**
— Class: **optional** — Required fix: 없음(운영자 재량). 고친다면 (d)의 stderr가 실행 불가인 검사를
지목하도록 한 절 추가.

---

## Recommendation — 최종 회차 처분

**PASS-WITH-DEBT, 0.92 (Tier L 임계 0.85).** 7개 must-pass 전부 통과, iter-2의 E1-E7 전부 종결,
점수 회귀 없음(0.80 → 0.89 → 0.92), 정체(stagnation) 신호 없음 — 세 회차에 걸쳐 변하지 않고 남은
결함은 하나도 없습니다.

**이 SPEC은 Implementation Kickoff Approval 게이트로 진행하기에 적합합니다.**

판단 근거를 명시합니다. 이번 델타는 문서만 고친 커밋이면서, 감사가 지적한 세 가지 중 두 가지를
**주장을 철회하는 방식**으로 닫았습니다 — v0.2.0의 "두 절은 독립적으로 실패한다"와 "nothing flags it"을
각각 과장이라고 스스로 적고 물렀습니다. 잘못을 부드럽게 만들지 않고 취소한 문서는 드물고, 이는 run-phase에서
기준을 읽을 사람에게 유리한 성질입니다.

남은 세 blocking 항목(F1·F2·F3)은 **전부 minor이고, 전부 한 절 또는 한 문장 수정이며, 어떤 결정도
다시 열지 않습니다.** 세 가지 모두 must-pass 실패가 아니고, 합쳐도 0.92를 임계 아래로 끌어내리지 않습니다.

### 부채 이관 — 무엇을, 어디에 기록해 run-phase로 넘기는가

이번이 마지막 허용 회차이므로 확인 감사는 더 없습니다. 다음 둘 중 하나를 권합니다.

**권장 — 지금 문서만 고치고 넘어간다.** F1·F2·F3은 SPEC 텍스트 수정이지 구현 작업이 아니며, 합쳐서
한 문단 분량입니다. 감사 회차를 새로 열지 않고 v0.3.1 문서 수정으로 처리한 뒤 kickoff 게이트로 갑니다.
이 경로를 택하면 run-phase로 넘어가는 부채는 없습니다.

**대안 — 그대로 진행하고 마일스톤에 매단다.** 고치지 않는다면 아래를 `progress.md` §E.1의
"Open for the Implementation Kickoff Approval gate" 아래에 **run-phase 의무**로 명시해야 합니다.
현재 `progress.md`는 F1-F4를 담고 있지 않으므로, 이 기록이 없으면 부채는 이 보고서에만 남습니다.

| 항목 | 언제까지 | 어디서 읽히는가 |
|---|---|---|
| **F1** — AC-TLE-020 픽스처 전제 못박기 | **M3 진입 전** (AC-TLE-020은 M3 소관) | `acceptance.md` AC-TLE-020 Given |
| **F2** — AC-TLE-018 (c) 둘째 연언 삭제 또는 행동적 형태 명시 | **M5 진입 전** (AC-TLE-018은 M5 소관) | `acceptance.md` AC-TLE-018 Then |
| **F3** — §R.10.1 원인 문장 교체 | 아무 때나, kickoff 전 권장 | `research.md` §R.10.1 |
| **F4** (optional) | 운영자 재량 | — |

### 게이트에서 운영자가 판단할 항목 (결함 아님, `progress.md`에 이미 기록됨)

1. 저장 형태 — 하나의 JSON 보유 컬럼 대 네 개의 스칼라 컬럼 (§B.1).
2. 일곱 번째 컬럼 계약 변경 (§G, half A에서 상속).
3. **v0.3.0 신규** — 커밋 subject를 record-time에 포착할 것인가 (§G, E7). 기록에 일곱 번째 사실을
   더하고 REQ-TLE-016의 셀 형식을 바꾸므로 설계 결정.

**PASS-WITH-DEBT 판정은 Implementation Kickoff Approval 게이트를 건너뛸 허가가 아닙니다.**
그 게이트는 점수와 무관하며, 이 판정은 그 게이트의 입력이지 대체물이 아닙니다.

---

## Gaps — 이 감사가 **관측하지 않은** 것

빈 목록이 아무것도 놓치지 않았다는 주장이 되지 않도록 명시합니다.

- **run-phase 동작은 하나도 검증하지 않았습니다.** `moai todo landed`, `internal/kanban/landing_evidence.go`,
  확장된 guard는 아직 존재하지 않습니다. 이들에 대한 모든 판단은 **기준(criterion)에 대한 판단**입니다.
- **테스트를 한 건도 실행하지 않았습니다.** 이번 회차는 문서 감사이고 구현이 없으므로 `go test`를
  돌리지 않았습니다(`go test ./...`는 이 머신에서 금지). iter-2가 실행한 `TestAudit2_` 프로브를
  재실행하지 않았습니다 — 019a/b/c의 독립성은 **iter-2의 측정을 인용**한 것이며 제가 재측정한 것이
  아닙니다. 소스 트리는 그 이후 움직이지 않았으나(델타에 소스 파일 0건), 인용은 인용입니다.
- **`git rev-parse --verify` / `git merge-base --is-ancestor`를 픽스처에 대해 실행하지 않았습니다.**
  두 명령이 카드 id를 입력으로 받지 않는다는 것은 문서화된 의미론과 argv 구성(`LandedGrepArgs` 측정)에서
  판단했습니다.
- **AC-TLE-018 절 (c)를 실제 테스트로 작성해 보지 않았습니다.** F2는 live switch의 소스를 읽고
  "집합을 꺼낼 수단이 없다"고 판단한 것이며, 제가 못 찾은 리플렉션/AST 기반 구현이 존재할 가능성은
  배제하지 못합니다 — 다만 그런 구현이라면 기준이 그것을 명시해야 한다는 지적은 그대로 유효합니다.
- **v0.1.0/v0.2.0에서 변경 없이 이월된 약 30개 `file:line` 핀을 재검증하지 않았습니다.** 이번 회차에서는
  델타가 인용하거나 제 판정이 의존한 핀만 측정했습니다: `prlink_landed.go:96-110`, `todo.go:148-152`,
  `todo_pr.go:48-65`, `backlog_sqlite.go:93-99`·`:102`·`:274-300`,
  `SPEC-KANBAN-TODO-CLI-001/spec.md:59`, `SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md:251-255`.
- **cross-model 백엔드를 조회하지 않았습니다.** `mcp__moai__audit_multi` / `codex_audit` / `glm_audit`
  미호출 — Claude 단독 판정입니다.
- **`design.md`를 §2 언급 수준으로만 참조했고 전문 감사하지 않았습니다.** 이번 델타가
  `design.md`를 변경하지 않았기 때문이며(diffstat 5개 파일에 미포함), Tier L 입력 계약상 읽어야 하는
  아티팩트를 델타 범위로 축소한 것은 확인 감사의 성질에 따른 것입니다.

### 감사 중 트리에 남긴 흔적 (보고 의무)

제 세션이 `cd`로 SPEC 디렉터리에 들어간 동안 statusline이 그 자리에 중첩 `.moai/state/` 3개 파일을
만들었습니다(`context-usage/3ddb34bb-….json` — 제 세션 id). gitignore 대상이라 `git status`에는
나타나지 않았고, **삭제해 원상복구했으며** 이후 `git status --short`는 빈 출력입니다.

## Residual risk

- **F2를 제가 잘못 읽었을 가능성.** "accepted-version set"을 저자가 행동적 비교로 의도했다면
  결함이 아니라 서술 부족입니다. 어느 쪽이든 기준 텍스트가 실행 형태를 말하지 않는다는 지적은 남습니다.
- **F1의 수정도 F1과 같은 함정을 가집니다.** 픽스처 전제를 한 절 추가하는 일은 작아서, 문장은
  들어갔는데 실제 픽스처가 두 카드 id 중 하나만 언급하도록 구성되지 않을 수 있습니다. 확인 감사가
  더 없으므로 이것을 잡을 지점은 **run-phase의 RED 관측**뿐입니다 — M3에서 누출 돌연변이를 심고
  경계 절이 실제로 red되는 것을 보기 전까지, F1은 닫힌 것으로 취급해서는 안 됩니다.
- **세 회차 동안 같은 형태가 계속 재발했습니다** — 실패할 수 없는 절이 iter-1(D-군), iter-2(E1),
  iter-3(F2)에 걸쳐 새 텍스트에서 매번 다시 나타났습니다. 매번 더 좁아지고는 있으나(전면 공허 →
  절반 공허 → 연언 하나), 위험은 개별 기준이 아니라 **새 기준을 쓰는 방식**에 있습니다. run-phase는
  새로 작성되는 어떤 단언에 대해서도 "이 문장을 red시키는 돌연변이를 하나 적어라"를 먼저 요구하는 것이
  좋습니다 — F2는 그 질문을 던졌으면 저자 스스로 발견했을 결함입니다.
- **iter-2의 019a/b/c 측정은 인용이며 소관 SPEC이 다릅니다.** guard는
  `SPEC-TODO-ARCHIVE-QUERY-001` 소유이고, SPEC 자신이 AC-TLE-019 안에 decay note를 두어
  run-phase 재측정을 요구합니다. 그 배치는 옳고, 인용을 재측정으로 승격시키면 안 됩니다.
- **이번 회차의 모든 측정은 HEAD `c0cfb2520`에서 취해졌습니다.** develop이 움직이면
  교차 SPEC status와 코드 핀은 부패합니다.
