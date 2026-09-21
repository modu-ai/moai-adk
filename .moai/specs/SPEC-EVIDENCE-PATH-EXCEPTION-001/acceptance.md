# SPEC-EVIDENCE-PATH-EXCEPTION-001 — 인수조건

## §A 판별식 규약 (모든 ignore 관련 AC에 선행 적용)

[HARD] ignore 동작을 주장하는 AC는 아래 둘 중 하나만 판별식으로 쓴다.

- **P1** — `git check-ignore --no-index -q <path>` : `rc=0` = 무시됨, `rc=1` = 무시되지 않음
- **P2** — `git status --porcelain --untracked-files=all <path>` : 출력 행 존재 = add 대상

[HARD] `git check-ignore -v`의 exit code를 판정에 쓰지 않는다 — negation 적중에도 0이다.
`-v`는 **「어느 규칙이 결정했는가」를 읽는 용도**로만 쓴다(AC-EPE-006이 그 용법이다).
[HARD] `--no-index`를 빼지 않는다 — 추적 파일이 스킵돼 `rc=1`이 「negation이 살렸다」로 오독된다.
[HARD] `git status`에 `--ignored`를 빼면 무시된 경로에 대해 **구조상 침묵한다.** 그 0행은
부재가 아니라 **미측정**이다(`spec.md` §1.0에서 실측).
[HARD] 모든 AC는 **실패할 수 있는 양성 대조**를 함께 명령한다. 양성 대조가 비어 나오면 그
AC는 PASS가 아니라 **계측 불가**로 기록한다.

프로브 경로는 실재하지 않아도 된다(P1은 경로 존재를 요구하지 않는다). 파일을 만들어 P2로
재는 경우, 그 파일은 `.moai/specs/SPEC-EVIDENCE-PATH-EXCEPTION-001/probe/` 아래 만들고 측정
직후 삭제한다.

[HARD] **픽스처 결과는 baseline-attribution이 될 수 없다**(REQ-EPE-008). `spec.md` §5의 표는
4줄짜리 픽스처에서 나왔고 이 트리 `.gitignore`의 규칙 순서(§1.5의 다섯 무리)를 재현하지 않는다
— 특히 픽스처에는 무리 D(`:294`)가 없다. 아래 어떤 AC도 그 표를 근거로 PASS하지 않으며, 모든
ignore AC의 baseline은 **이 트리의 실제 `.gitignore`에서, 이번 실행에서** 관측된 출력이다.

[HARD] **줄수·줄번호를 인용하는 AC는 트리를 명명한다.** 이 워크트리는 417줄, `main` 트리는
319줄이다.

[HARD] **드리프트하는 수치에 대해 정확-일치를 단언하지 않는다.** primary 체크아웃의 폭 수치
(305 / 155 / 19)와 docs-site 행 수는 다른 세션·다른 시점에 바뀐다. 아래 AC는 그 수치들을
**비-감소** 또는 **델타** 또는 **0으로의 도달**로만 잰다.

[HARD] **폭 의존 AC는 둘뿐이다**(`spec.md` §3.4): AC-EPE-002(결정 의존 — 폭이 곧 입력),
AC-EPE-013(귀결 기록형 — 어느 쪽도 요구하지 않는다). `REQ-EPE-010`은 **참조 의존**이며
결정하지 않으므로, 그것을 검증하는 AC-EPE-005는 깊이만 잰다. 나머지 전부는 폭이 (A)든 (B)든
같은 판정을 낸다.

---

## §B 수용 기준

### B.1 `.gitignore` 예외 — 되살아나는 것 (REQ-EPE-005 / 011)

> **`REQ-EPE-005`의 커버리지가 어디 있는가 (iteration 5, D22).** 그 요구사항은 두 절로 돼
> 있고 각각 다른 AC가 잰다: `shall`(판정서를 되살린다) → **`AC-EPE-001`**,
> `shall not`(카드 본문·로그·중간 산출물을 되살리지 않는다) → **`AC-EPE-003`**. 커버리지는
> iteration 4까지도 실질적이었고 없던 것은 **이름 붙이기**뿐이었다. 이제 두 AC가 명시한다.
>
> 측정 주의: `REQ-EPE-003` · `013` · `015`에 대한 리터럴 `grep -c 'REQ-EPE-0NN'`도 0을 내지만
> 그것은 부재가 아니다 — §B.4 · §B.5 · §B.6 제목이 `REQ-EPE-012 / 013` 같은 **축약형**으로
> 이름 붙이기 때문이다. 세 건은 계측 artifact이고, `REQ-EPE-005`만이 축약형으로도 부재였다.

**AC-EPE-001** — 판정서 파일이 무시되지 않는다 (REQ-EPE-005의 `shall` 절)
Given `.gitignore`에 M1의 예외가 적용된 상태에서,
When `git check-ignore --no-index -q .moai/reports/tZZZ/verdict.md` 를 실행하면,
Then `rc=1`(무시되지 않음)이다.
양성 대조: `git check-ignore --no-index -q .moai/reports/tZZZ/report.md` → `rc=0`.
(대조가 `rc=1`을 내면 예외가 과도하거나 블랭킷이 깨진 것 — 계측 불가가 아니라 FAIL.)
RED-now(이 트리, HEAD `116820f40`): 같은 프로브가 지금 `rc=0` — `-v` 출력이
`.gitignore:227:.moai/reports/*`. 착수 전에 초록이 아니다.

**AC-EPE-002** [폭 매개변수화 — `spec.md` §3.4의 결정 의존. **입력 확정: (B), 2026-09-21**] —
확정된 판정서 집합 전원이 무시되지 않는다
Given REQ-EPE-006의 폭이 운영자에 의해 (A) 또는 (B)로 확정된 상태에서 — **확정값은 (B)이므로
집합은 `{verdict.md}` 단독이고, 이 AC는 「평가 불가」를 벗어나 평가 가능하다** —,
When **그 확정 집합의** 각 파일명에 대해 `.moai/reports/tZZZ/<name>` 을 P1으로 재면,
Then **전원** `rc=1`이다.
- 독법 (A) 집합: `verdict.md`, `plan-audit.md`, **`plan-audit-iter<N>.md` 계열 전체**,
  `sync-audit.md`, `sync-audit-verdict.md`, `review-verdict.md`
- 독법 (B) 집합: `verdict.md`
양성 대조: 같은 디렉터리의 `evidence.md` → `rc=0`.
이 AC는 어느 독법도 선취하지 않는다 — 폭이 그 입력이며, 폭 없이는 **평가 불가**(FAIL 아님).

[NOT APPLICABLE — 2026-09-21 독법 (B) 확정] **아래 [HARD] 블록 전체는 (A)를 택했을 때에만
발화하는 구현 지시다.** (B)의 확정 집합은 `verdict.md` 단독이고 계열 글롭
(`!.moai/reports/*/plan-audit-iter*.md`)은 **존재하지 않으므로**, 겨눌 계열 구성원 자체가 없다
— 블록의 마지막 문단(D18 조건화)이 이미 그렇게 적어 두었고, 이제 그 조건이 판정됐다.
**블록을 지우지 않는 이유**: (A)를 택했다면 「구체 파일명 열거로는 다음 iteration 판정서가
되살아나지 않는다」는 함정이 실재했고, 그 함정이 검토됐다는 기록은 폭이 다시 논의될 때
필요하다. 아래는 **채택되지 않은 갈래의 설계 기록**으로 읽는다.

[HARD] **`plan-audit-iter<N>.md`는 계열이지 파일 하나가 아니다(iteration 3, D13).** `spec.md`
§3.2가 정의하는 것은 계열이며, iteration 2 판이 `plan-audit-iter1.md`를 구체 파일명으로 열거한
탓에 **문자 그대로 구현하면 `plan-audit-iter2.md`가 되살아나지 않았다** — 그 파일은 이 카드의
iter2 판정서이자 이 결함의 살아 있는 실사례다. 따라서 (A)를 택할 경우 구현은 파일명당 negation
한 줄이 아니라 **계열 글롭 한 줄**(`!.moai/reports/*/plan-audit-iter*.md`)이며, `design.md` §1.3의
「파일명당 negation 한 줄」은 이 계열에 대해 그렇게 읽는다.
**(A)를 택한 경우에만** P1 프로브는 계열의 실재 구성원 **둘 이상**을 겨눈다 — 최소
`plan-audit-iter1.md`와 `plan-audit-iter2.md`. 하나만 재면 계열이 아니라 그 한 파일만
검증된다. 독법 **(B)에서는 이 요구가 적용되지 않는다** — (B)의 확정 집합은 `verdict.md`
단독이고 계열 글롭은 존재하지 않으므로, 겨눌 구성원 자체가 없다(이 조건화는 iteration 4, D18
— 폭을 고르는 것이 아니라 어느 독법에서 요구가 발화하는지를 적는 것이다).

**AC-EPE-021** — 깊이 제한이 `.gitignore` 주석에 적혔다 (REQ-EPE-011)
Given M1 적용 후,
When `grep -n 'reports/lead' .gitignore` 를 실행하면,
Then 1행 이상이 나오고, 그 행(들)은 주석이며 **놓치는 형태를 이름 붙인다** —
`reports/lead/<batch>/verdict.md`가 한 단계 깊이 제한에 걸려 되살아나지 않는다는 서술.
양성 대조: `grep -c 'moai/reports' .gitignore` → 0보다 큼(파일을 제대로 읽고 있다).
RED-now(이 트리, HEAD `116820f40`): `grep -c 'reports/lead' .gitignore` → **0**, 양성 대조는
**23**. 계측기가 발화하는데 대상이 없으므로 착수 전 판정은 FAIL이다 — 공허하지 않다.
(이 AC는 **주석의 존재**를 재고, `AC-EPE-005`는 **깊이 동작**을 재며, `AC-EPE-010`은
`:107-108` 주석을 잰다 — 셋은 서로 다른 대상이다.)

### B.2 `.gitignore` 예외 — 무시된 채로 남아야 하는 것

**AC-EPE-003** — 비-판정서 산출물은 여전히 무시된다 (REQ-EPE-005의 `shall not` 절 — 폭 제한)
Given 예외 적용 상태에서,
When `report.md` · `evidence.md` · `pr-body.md` · `probe.log` 각각을
`.moai/reports/tZZZ/` 아래 경로로 P1에 넣으면,
Then **전원** `rc=0`이다.
양성 대조: 같은 디렉터리의 `verdict.md` → `rc=1`(계측기가 두 방향을 모두 낸다).

**AC-EPE-004** — `plan-audit/` 누수가 닫혀 있다 (REQ-EPE-007)
Given 예외 적용 상태에서,
When `git check-ignore -v --no-index .moai/reports/plan-audit/verdict.md` 를 실행하면,
Then `-v` 출력이 존재하고(= 어떤 규칙이 적중), 같은 경로에 대한 P1이 `rc=0`(무시됨)이다.
**추가로 `-v`가 이름 붙인 규칙의 줄번호를 기록한다** — 착수 전에는 `:294`가 결정하고 있으며
(`spec.md` §1.5 측정), 삽입 이후 어느 규칙이 결정하는지가 바뀔 수 있기 때문이다. 바뀌었다면
그 사실이 §E.2에 적혀야 PASS다.
양성 대조: `.moai/reports/tZZZ/verdict.md` → `rc=1`.
음성 대조(카브아웃이 살아 있는지): `.moai/reports/plan-audit/.gitkeep` → `rc=1`.

**AC-EPE-005** — 깊이 2 이상은 되살아나지 않는다 (REQ-EPE-010 / 범위 밖 선언과 일치)
Given 예외 적용 상태에서,
When `git check-ignore --no-index -q .moai/reports/tZZZ/sub/verdict.md` 를 실행하면,
Then `rc=0`이다.
양성 대조: `.moai/reports/tZZZ/verdict.md` → `rc=1`.
(이 AC는 **의도한 제한을 고정**한다. 되살아나면 폭이 선언보다 넓다는 뜻이므로 FAIL.
REQ-EPE-010은 깊이만 결정하므로 이 AC도 깊이만 잰다 — 파일명 집합은 AC-EPE-002의 몫이다.)

### B.3 제자리 측정 의무 (REQ-EPE-008)

**AC-EPE-006** [HARD] — 예외 효과가 **이 트리의 실제** `.gitignore`에서 재측정됐다
Given run 단계가 예외를 적용한 뒤, 작업 트리를 `git rev-parse --show-toplevel` 로 명명하고,
When `wc -l .gitignore` 와 `git check-ignore -v --no-index .moai/reports/tZZZ/verdict.md` 를
같은 턴에 실행하면,
Then 줄수는 그 트리의 착수 시점 줄수(이 워크트리 기준 417) ±(이 카드가 추가/삭제한 줄수)와
일치하고, `-v` 출력의 적중 규칙은 **이 카드가 삽입한 줄번호**를 가리킨다.
양성 대조: 같은 `-v` 프로브를 `report.md`에 돌려 블랭킷 규칙(`.moai/reports/*`)이 나오는지
확인 — 블랭킷이 여전히 살아 있음을 보인다.
[HARD] 이 AC의 baseline은 **이번 실행의 관측**이다. `spec.md` §5의 픽스처 결과를 인용해
이 AC를 PASS시킬 수 없다(REQ-EPE-008). 픽스처는 무엇을 잴지만 정했다.
[HARD] 줄수는 트리 귀속과 함께 적는다 — `main` 트리의 같은 파일은 319줄이다.

**AC-EPE-007** — 기존 추적 54건과 충돌하지 않는다 (REQ-EPE-009)
Given 예외 적용 전후에,
When `git ls-files .moai/reports | wc -l` 와 `git status --porcelain .moai/reports` 를 실행하면,
Then 추적 수가 감소하지 않고, `status`에 ` D`(삭제) 행이 없다.

### B.4 죽은 negation 철회 (REQ-EPE-012 / 013)

**AC-EPE-008** — `:109` negation이 사라졌다
Given M3 적용 후,
When `grep -n '!\.moai/reports/\*\*/\*\.log' .gitignore` 를 실행하면,
Then 출력이 0행이고 exit 1이다.
양성 대조: `grep -n '!\.moai/reports/t338/' .gitignore` → 1행 이상(다른 negation은 그대로).
RED-now(이 트리, HEAD `116820f40` — 이번 실행): `grep -c '!\.moai/reports/\*\*/\*\.log' .gitignore`
→ **1**, 양성 대조 `grep -c '!\.moai/reports/t338/' .gitignore` → **2**. 지울 대상이 실재하므로
공허하지 않다.

**AC-EPE-009** — 로그는 철회 후에도 무시된다 (동작 무변화)
Given M3 적용 후,
When `git check-ignore --no-index -q .moai/reports/tZZZ/probe.log` 를 실행하면,
Then `rc=0`이다 — 철회 전과 같다(죽은 규칙이었으므로 동작이 바뀌지 않아야 한다).
양성 대조: `.moai/reports/tZZZ/verdict.md` → `rc=1`.

**AC-EPE-010** — `:107-108` 주석이 정정됐다
Given M3 적용 후,
When 삭제 지점 주변 주석을 읽으면,
Then 주석은 (a) negation이 제거된 이유와 (b) 로그가 의도적으로 무시된다는 사실을 담고,
「cited evidence paths must resolve post-merge」처럼 **무언가를 고쳤다고 읽히는 문장**을
담지 않는다.
기계 점검: `grep -c 'must resolve post-merge' .gitignore` → `0`.
양성 대조: `grep -c 'moai/reports' .gitignore` → 0보다 큼(파일을 제대로 읽고 있다).
RED-now: `grep -c 'must resolve post-merge' .gitignore` → **1**(이 트리, HEAD `116820f40`).

### B.5 독트린 문언 (REQ-EPE-001 / 002 / 003)

**AC-EPE-011** — 넓은 추적 주장이 사라졌다
Given M2 적용 후,
When `grep -rlnE 'citation target is a (\*\*)?tracked(\*\*)? path' --include='*.md'
--include='*.toml' .` 를 저장소 루트에서 실행하면,
Then `.claude/rules/…/agent-common-protocol.md`, 같은 이름의 `reference.md`, 그리고 두
템플릿 미러 **4건에서 좁혀진 문언**이 읽히고, 어느 것도 「임의 산출물이 추적 경로로
반출된다」로 읽히지 않는다. `SPEC-HIERARCHICAL-TEAM-001/progress.md`는 **수정되지 않는다**.
양성 대조: 같은 정규식에 존재하지 않는 토큰을 넣어 `0`행을 확인.
주의: 이 SPEC 자신의 `spec.md` §1.3이 같은 문자열을 인용하므로 지금 이 명령은 **자기-적중을
포함한 6행**을 낸다. 자기-적중은 수정 표면이 아니며, 세는 대상에서 제외한다.

**AC-EPE-012** — 비-판정서 경로의 추적 주장이 사라졌다 (REQ-EPE-002)
Given M2 적용 후,
When 아래 **네** 프로브를 실행하면,
Then **전원 0**이다.

| # | 형태 | 프로브 | RED-now (이 트리, HEAD `116820f40`) |
|---|---|---|---|
| P-a | i | ``grep -c 'tracked path `.moai/reports/<card-id>/M<n>\.<AC-id>\.log`' .claude/agents/moai/manager-lead.md`` | **1** |
| P-b | i | `grep -c 'only the tracked path does' .claude/agents/moai/manager-lead.md` | **1** |
| P-c | i | `grep -c '(tracked; exported before citing' .claude/output-styles/moai/moai.md` | **2** |
| P-d | ii | `grep -c 'the deciding lines to the tracked' .claude/agents/moai/manager-lead.md` | **1** |

양성 대조: `grep -c 'moai/reports' .claude/agents/moai/manager-lead.md` → **5**(0보다 큼).
음성 대조: 같은 파일에 `grep -c 'ZZZNOPE'` → **0**(계측기가 아무거나 세지 않는다).
네 값 모두 이번 실행에서 관측했다.

[HARD] **P-d는 iteration 3이 추가했다(D12).** `manager-lead.md:63`은 `REQ-EPE-002`의 좌표
목록에 있었지만 **어떤 AC도 그 좌표에 닿지 않았다** — P-a·P-b·P-c는 `:152`·`:154`·`moai.md`를
겨누고, `AC-EPE-011`의 정규식은 `:63`의 문언에 맞지 않는다. `:63`이 추적으로 주석하는 것은
파일이 아니라 **디렉터리**(`.moai/reports/<card-id>/`)이고, 같은 형태를 `AC-EPE-022`는
docs-site에서 **12행의 위반**으로 센다. 저장소 쪽만 비워 두면 같은 결함이 한 층 아래에서
살아남으므로 프로브를 추가했다.
미러 주의: 같은 문장이 `internal/template/templates/.claude/agents/moai/manager-lead.md`에도
**1건** 있다(이번 실행 관측). 루트만 고치면 `AC-EPE-014`(미러 동기화)가 잡는다.

[HARD] **이 AC가 왜 이 세 프로브를 쓰는가 — iteration 2의 정정.** iter1 판은
`grep -n 'tracked' … | (M<n>-report.md를 이름 붙인 행 세기)` 였고, 그 수가 **착수 전부터 0**
이었다. `manager-lead.md`의 `tracked` 3행 중 `M<n>-report.md`를 담은 행이 하나도 없기 때문이다
(`:161`이 그 토큰을 담지만 `tracked`라는 낱말이 없다). 즉 공허한 인수조건이었고, 그 결과
`REQ-EPE-002`가 지목한 좌표의 수정 여부를 **어떤 AC도 검출하지 못했다**.

정정의 근거: `:161`의 접기 행은 `.moai/reports/<card-id>/M<n>-report.md`를 **해소 가능한 인용
경로로 제시**하고, 그것을 추적 주장으로 만드는 것은 바로 위 **`:154`의
"and only the tracked path does"** 문장이다(= P-b). 그래서 판정식은 그 문장을 직접 겨눈다.
P-a는 같은 파일 `:152`의 **직접** 추적 주장(비-판정서 `.log` 경로)이고, P-c는
`moai.md:354,561`의 같은 형태다. 셋 다 지금 비어 있지 않으므로 이 AC는 실패할 수 있다.

추가 기록 의무: run 단계는 세 프로브를 0으로 만든 **대체 문언**을 §E.2에 인용한다. 지우기만
하고 대체하지 않으면 「무엇이 참인가」가 사라지므로, 인용 없는 0은 PASS가 아니라 **미완**이다.

**AC-EPE-013** [귀결 기록형 — 어느 독법도 요구하지 않는다] — 생존 표면의 상태가 측정되고
선택된 폭의 귀결과 일치한다
Given M2 적용 후,
When `plan-auditor.md:601` 과 `sync-auditor.md:108` 이 지목하는 파일명(`plan-audit.md`,
`plan-audit-iter<N>.md`, `sync-audit.md`, `sync-audit-verdict*.md`)을 §A P1으로 재면,
Then 결과가 `spec.md` §3.0 귀결표와 일치한다:
- 독법 (A) → 전원 `rc=1`, 두 [HARD] 문장이 참이 됨
- 독법 (B) → 전원 `rc=0`, 두 [HARD] 문장이 **거짓인 채로 남고**, 그 잔존 결함이
  `progress.md` §E.3 Gaps에 명시 기록됨(기록되지 않으면 FAIL)
양성 대조: 같은 디렉터리의 `verdict.md` → `rc=1`(계측기가 두 방향을 모두 낸다).
이 AC는 폭 선택을 **평가하지 않는다** — 선택의 귀결이 알려졌는지만 평가한다.

[2026-09-21 — 발화하는 갈래가 정해졌다] 확정값이 (B)이므로 **위 두 줄 중 (A) 갈래는 해당
없음**이고, 이 AC는 (B) 갈래로 평가된다: 네 파일명 전원 `rc=0`, 두 [HARD] 문장이 거짓인 채로
남고, 그 잔존이 `progress.md` §E.3 Gaps에 명시 기록돼 있을 것. (A) 갈래를 지우지 않는 이유는
`spec.md` §3.0 귀결표와 이 AC가 **한 쌍**이기 때문이다 — 한쪽만 남기면 「(B)를 고르면 무엇을
잃는가」가 AC 층에서 사라진다.

### B.6 미러와 기계 방출 (REQ-EPE-014 / 015 / 016)

**AC-EPE-014** — 미러가 동기화됐다
Given M4 적용 후,
When 수정된 4개 루트 파일 각각에 대해 `internal/template/templates/` 대응 파일에서 같은
좁혀진 문언을 grep하면,
Then 루트에서 바뀐 문장이 미러에도 존재한다.
양성 대조: 존재하지 않는 토큰 grep → 0행.

**AC-EPE-015** — codex toml이 재생성으로만 바뀌었다
Given M4 적용 후,
When `make agents-emit-check` 를 실행하면,
Then exit 0 (커밋된 `.toml` 이 `.md` 방출 결과와 일치).

**AC-EPE-016** — 템플릿 중립성
Given M4 적용 후,
When **`internal/template/templates/` 아래** 변경 hunk에서 SPEC ID(`SPEC-`), REQ 토큰(`REQ-`),
내부 날짜, 커밋 SHA, `CLAUDE.local`, macOS 절대경로를 grep하면,
Then 전원 0행이다.
양성 대조: 같은 hunk에 `grep -c 'evidence'` → 0보다 큼(hunk를 제대로 겨누고 있다).
범위 주의: docs-site는 `internal/template/templates/` 뿌리 밖이므로 이 AC의 대상이 아니다
(REQ-EPE-016 본문 참조).

### B.7 선행 SPEC 충돌 기록 (REQ-EPE-004)

**AC-EPE-017** — 충돌이 양쪽 끝에서 발견 가능하다
Given M6 적용 후,
When `SPEC-EVIDENCE-CITATION-CANON-001/spec.md` 의 frontmatter와 HISTORY를 읽으면,
Then `partially_superseded_by: [SPEC-EVIDENCE-PATH-EXCEPTION-001]` 과
`related_specs`에 이 SPEC이 있고, HISTORY에 2026-09-20 행이 **추가**돼 있다.

**AC-EPE-018** — 완료 SPEC의 요구사항 본문이 보존됐다
Given M6 적용 후,
When `git diff -- .moai/specs/SPEC-EVIDENCE-CITATION-CANON-001/spec.md` 를 실행하면,
Then diff hunk가 frontmatter와 HISTORY 절에만 걸리고, `## 2. 요구사항` 이하 REQ-ECC-001 ~
REQ-ECC-011 본문 줄에 **-(삭제) 행이 0개**다.
기계 점검: `git diff -U0 -- <path> | grep '^-' | grep -c 'REQ-ECC'` → `0`.
양성 대조: `git diff -U0 -- <path> | grep -c '^+'` → 0보다 큼(diff가 비어 있지 않다).

### B.8 미해소 축의 처리

**AC-EPE-019** — 폭이 누구에 의해서도 암묵적으로 확정되지 않았다
Given run 단계 진입 시점에,
When REQ-EPE-006의 상태를 확인하면,
Then 다음 중 하나다: (a) 운영자가 (A) 또는 (B)를 명시했고 그 출처가 기록돼 있다,
또는 (b) 미해소이며 AC-EPE-002가 **평가 불가**로, AC-EPE-013이 **미측정**으로 기록돼 있다.
에이전트·리드가 「권고가 곧 결정」으로 진행한 흔적이 있으면 FAIL.
**plan 단계는 (b) 상태로 닫힐 수 있다** — 이 축의 미해소는 plan-phase 결함이 아니다.

**AC-EPE-020** — 폭 무관 AC들이 실제로 무관하다
Given run 단계에서,
When AC-EPE-001, 003~012, 014~018, 021~023을 평가하면,
Then 어느 것도 REQ-EPE-006의 값을 입력으로 요구하지 않는다 — `spec.md` §3.4의 전수 열거가
참임이 확인된다. 열거 밖의 AC가 폭에 의존함이 드러나면, 그 사실이 §3.4에 추가돼야 한다
(발견 자체는 FAIL이 아니고, **기록하지 않는 것**이 FAIL이다).
추가 점검(iteration 2): §3.4는 **결정 의존**과 **참조 의존**을 구분한다. `REQ-EPE-010`이
참조 의존으로 분류된 것이 맞는지 — 즉 폭이 (A)든 (B)든 그 요구사항의 문언과 AC-EPE-005의
판정이 바뀌지 않는지 — 를 함께 확인한다.

### B.9 docs-site 공개 문서 (REQ-EPE-017 / 018)

**AC-EPE-022** — docs-site에 비-판정서 추적 주장이 남지 않았다
Given M5 적용 후,
When 로케일 내성 인벤토리를 실행하면,

```
grep -rnE '(tracked|추적|追跡|跟踪|追踪)' --include='*.md' docs-site/content/ \
  | grep -E '\.moai/reports'
```

Then 남은 각 행은 **판정서 파일**을 이름 붙이며, 비-판정서 좌표(`REQ-EPE-002`의 두 형태 —
(i) 비-판정서 파일명, (ii) 파일명 없는 `<card-id>/` 디렉터리 단독)를 추적 경로로 서술하는 행이
**0행**이다.

기계 하위 점검 — **두 갈래로 나누고, 그 분할이 남김 없음을 잔차로 다시 잰다**(모두 같은
인벤토리 파이프에 이어 붙임).

[HARD] **정본 carrier는 아래 판정식 원장(fenced 블록)이고, 표는 그 id를 인용할 뿐 command를
싣지 않는다 (iteration 4, D15 — 범위는 바로 아래에서 한정한다).**

> **[범위 한정 — iteration 5, D20]** 이 [HARD]는 **§B.9의 판정식 원장 규약**이며, 이
> `acceptance.md` 전체에 소급되는 일반 금지가 **아니다.** 구체적으로 **§B.5 `AC-EPE-012`의
> P-a~P-d 셀은 이 규칙의 대상이 아니고, 옮겨지지도 않는다.**
>
> 판별식은 「표에 command가 있는가」가 아니라 **「그 command가 표 셀에서 훼손되는가」**다.
> 훼손의 기제는 하나뿐이다 — markdown 표 셀이 `|`와 백틱을 이스케이프해야 한다는 것. 따라서
> 셸 메타문자(특히 `|`)를 **필요로 하는** command만 이 규칙에 걸린다.
>
> [HARD] **P-a~P-d의 안전은 설계가 아니라 우연이다.** iteration 4 감사가 네 셀을 축자
> 실행해 `1 / 1 / 2 / 1`을 재측정했고 훼손이 없었다 — 그 넷이 `|`를 **한 글자도 쓰지 않는**
> 단순 `grep -c` 이기 때문이다. 우연히 안전한 것과 안전하도록 설계된 것은 다르므로, 다음
> 사람을 위해 그 사실을 여기 적고 **조건**을 남긴다:
>
> - **파이프·교대(`|`)·서브셸·리다이렉션을 담는 판정식은 표 셀에 두지 않는다.** 원장 블록에
>   두고 표는 id만 인용한다. §B.9의 `L-S1`~`L-R2`가 그 예다.
> - **셸 메타문자가 없는 단일 `grep -c '<리터럴>' <파일>` 형태는 표 셀에 둘 수 있다.**
>   §B.5의 P-a~P-d가 그 예다.
> - 어느 셀이든 **백틱을 담은 패턴**을 쓰게 되면 그 순간 원장으로 옮긴다 — 백틱은 두 번째
>   훼손 경로이고, §B.9의 원장이 바로 그것을 담는다.
>
> 이 한정이 없으면 위 [HARD]는 문언상 P-a~P-d를 위반으로 만든다. 그것이 **문언이 의도보다
> 넓은** 형태이며, 이 SPEC 자신이 `spec.md`에서 같은 결함을 고치고 있으므로 여기서도 문언을
> 좁힌다.

markdown 표 셀은 `|`와 백틱을 이스케이프해야 해서 셸
메타문자를 훼손한다 — `(log\|md)`는 ERE에서 교대가 아니라 **리터럴 `log|md`**가 되므로 S1이
아무것도 맞히지 못한다. 실측(이번 실행, 같은 파이프): 실행본 → `8`(exit 0), 표 셀을 그대로
복사한 이스케이프본 → `0`(exit 1). **이 훼손은 green 시점에 침묵한다** — 이 AC의 PASS 목표가
`0`이므로, 이스케이프본을 복사한 실행자는 `0`을 받고 그 값이 PASS 조건과 일치한다. 모순
신호가 정확히 필요한 순간에 사라진다. 따라서 수리는 **표에서 command를 빼는 것**이지 이스케이프를
다시 고르는 것이 아니다 — 후자는 같은 층에서의 세 번째 시도일 뿐이다
(`verification-completeness.md` §2.1의 evidence-ledger carrier 권고).

#### 판정식 원장 — 이 블록이 정본 실행본이다

아래 각 줄은 **원본 파일 바이트 그대로** 위 Given/When의 인벤토리 파이프 뒤에 이어 붙여
실행한다. 어떤 셀도, 어떤 표도 이 블록을 대신하지 않는다.

```
# L-S1 — 형태 i (비-판정서 파일명)
grep -cE '\.moai/reports/[^/ `]+/M[^ `]*\.(log|md)'

# L-S2 — 형태 ii (파일명 없는 `<card-id>/` 디렉터리 단독)
grep -cE '\.moai/reports/[^/ `]+/`'

# L-VD — 판정서 파일을 이름 붙인 행 (분류 관측용; PASS 조건 아님)
grep -cE 'verdict\.md|plan-audit.*\.md|sync-audit.*\.md'

# L-R1 — 잔차: S1·S2 어느 쪽에도 안 걸리면서 판정서도 이름 붙이지 않는 행
grep -vE '\.moai/reports/[^/ `]+/M[^ `]*\.(log|md)' | grep -vE '\.moai/reports/[^/ `]+/`' | grep -vE 'verdict\.md|plan-audit.*\.md|sync-audit.*\.md' | wc -l

# L-R2 — 중복: S1과 S2에 동시에 걸리는 행
grep -E '\.moai/reports/[^/ `]+/M[^ `]*\.(log|md)' | grep -cE '\.moai/reports/[^/ `]+/`'
```

| # | 형태 | 실행본 | 처분 | RED-now | PASS 조건 |
|---|---|---|---|---|---|
| S1 | i | 원장 `L-S1` | release-blocking | **8** | `= 0` |
| S2 | ii | 원장 `L-S2` | release-blocking | **12** | `= 0` |
| R1 | 잔차 | 원장 `L-R1` | regression-guard | **0** (착수 전 초록이 정상) | `= 0` |
| R2 | 중복 | 원장 `L-R2` | regression-guard | **0** (착수 전 초록이 정상) | `= 0` |

[HARD] **PASS 조건은 이 넷이며, 분할 항등식은 그중에 없다 (iteration 4, D16).** 종전 판은
`8 + 12 = 20`이라는 항등식을 「하위 점검이 인벤토리의 어느 부분도 놓치지 않는다」의 근거로
적었다. **그 추론은 성립하지 않는다** — 양쪽에 걸리는 행 1개와 어느 쪽에도 안 걸리는 행 1개가
함께 있어도 합은 똑같이 20이다. 더구나 green 시점 형태 `0 + 0 = 0`은 **동어반복**이다: S1과
S2가 0이면 인벤토리에 무엇이 남아 있든 합은 0이므로 아무것도 배제하지 못한다. 남김-없음의
근거는 항등식이 아니라 **직접 측정한 잔차(`L-R1`)와 중복(`L-R2`)**이고, 그 둘은 green 시점에
다시 재기 때문에 동어반복으로 퇴화하지 않는다.

`L-R1`·`L-R2`가 착수 전부터 0인 것은 공허함이 아니라 **regression-guard의 정의**다 — 재는
것은 「예외가 적용된 뒤에도 두 하위 점검이 인벤토리를 남김 없이·중복 없이 덮는가」이며,
`S1 = 0 AND S2 = 0`만 보면 **두 점검 어느 쪽도 분류하지 않는 위반 형태가 조용히 살아남는다**.
가상의 형태가 아니다: `.moai/reports/<card-id>/report.md`는 인벤토리에 걸리고, 선두 `M`이
없어 S1에 안 걸리고, 디렉터리 뒤 백틱이 없어 S2에 안 걸리고, 판정서도 아니다 —
`report.md`·`evidence.md`·`pr-body.md`는 `spec.md` §7이 **비-판정서로 명시 열거**하는 바로
그 이름들이다. `L-R1`이 그 행을 잡는다.

`L-R1` 변이 프로브 (이번 실행, 파일을 만들지 않고 문자열을 파이프에 넣었다):

```
$ printf '...the tracked path `.moai/reports/<card-id>/report.md` is cited' | (L-R1)
1        ← 위반 형태를 잡는다
$ printf '...the tracked path `.moai/reports/<card-id>/M<n>.AC-001.log`'    | (L-R1)
0        ← 정상 S1 행은 잡지 않는다
```

두 방향이 함께 발화하므로 `L-R1`은 전-적중 변이(모든 줄을 잔차로 읽음)와 무-적중 변이(아무
줄도 잔차로 읽지 않음) 양쪽을 배제한다.

RED-now (이 트리, HEAD `116820f40` — 모두 이번 실행의 관측):

```
$ grep -rnE '(tracked|추적|追跡|跟踪|追踪)' --include='*.md' docs-site/content/ \
    | grep -E '\.moai/reports' | wc -l
      20
$ (같은 파이프) | grep -cE '\.moai/reports/[^/ `]+/M[^ `]*\.(log|md)'
8
$ (같은 파이프) | grep -cE '\.moai/reports/[^/ `]+/`'
12
$ (같은 파이프) | (원장 L-R1)                                        ← 잔차
       0
$ (같은 파이프) | (원장 L-R2)                                        ← 중복
0
$ (같은 파이프) | grep -cE 'ZZZNOPE_NOT_A_TOKEN'                     ← 음성 대조
0
$ grep -rlE '(tracked|추적|追跡|跟踪|追踪)' --include='*.md' docs-site/content/ \
    | xargs grep -lE '\.moai/reports' | wc -l
      12
```

S1이 잡는 8행: `{en,ko,ja,zh}/advanced/manager-lead.md` 의 `:122`·`:123`(각 로케일 2행).
S2가 잡는 12행: 같은 네 로케일의 `agent-guide.md` 1행 + `token-budget.md` 2행.
행 목록은 `research.md` §5.4가 파일:줄 단위로 싣는다.

**판정서 파일을 이름 붙이는 행은 0행이다** — `| grep -cE 'verdict\.md|plan-audit.*\.md|sync-audit.*\.md'`
→ `0`(이번 실행). 즉 인벤토리 20행 전부가 위반이며, 정상 행은 하나도 없다.

양성 대조: `grep -rl 'moai/reports' docs-site/ | wc -l` → **41**(0보다 큼, 계측기 도달 확인).

[HARD] **iteration 3의 정정(D10).** iteration 2가 적은 하위 점검
`grep -cE '(<check>\.log|M<n>-report\.md|M<n>\.<AC-id>\.log)'`은 이 인벤토리에서 **0행**을
낸다 — 착수 전부터 초록이므로 **공허한 인수조건**이었다. 원인은 로케일이 플레이스홀더 자체를
번역한다는 것이다: 실제 문자열은 `M<milestone>.<AC-id>.log` / `M<마일스톤>…` / `M<マイルストーン>…`
/ `M<里程碑>…`이고, `M<n>` 리터럴은 어디에도 없다. `M<n>-report.md` 역시 본문에는 구체 예시
`M2-report.md`로 적혀 있으며, `<check>.log`는 docs-site에 아예 없다. 위 S1은 플레이스홀더 이름을
전제하지 않고 **`.moai/reports/<무언가>/` 뒤에 오는 `M…` 파일명**을 겨누어 이 함정을 피한다.
그리고 iteration 2가 RED-now로 적었던 **18은 어떤 명령의 출력도 아니었다** — `20 − 2`로 계산한
값이었다. 위 블록의 수치는 전부 명령을 다시 돌려 출력을 옮긴 것이다.

[HARD] **인벤토리 grep은 로케일 내성이어야 한다.** 영어 토큰만 겨눈 grep은 3파일 5행만
보고하며, 그것이 iter1이 이 표면 전체를 놓친 경로다. `ko`는 플레이스홀더 자체도 번역한다
(`.moai/reports/<카드-id>/`).
[HARD] 이 AC는 **정확 행 수 일치를 단언하지 않는다** — 재는 것은 위반 행의 **0으로의 도달**
이고, 20·8·12는 baseline 기록일 뿐이다. 편집이 인벤토리 행 수를 바꿔도 PASS 조건은 위 표의
네 줄 — `S1 = 0` · `S2 = 0` · `L-R1 = 0` · `L-R2 = 0` — 이며, 행 수를 세는 조항은 없다.

**AC-EPE-023** — 4-로케일 동기화가 지켜졌고 사이트가 빌드된다 (REQ-EPE-018)
Given M5 적용 후,
When 아래 둘을 실행하면,
Then 둘 다 통과한다.

1. **로케일 패리티** —
   `git diff --name-only -- docs-site/content | sed -E 's#^docs-site/content/[a-z]{2}/##' | sort | uniq -c`
   의 **모든 행의 count가 정확히 4**다(en/ko/ja/zh). 3 이하가 하나라도 있으면 FAIL.
   양성 대조: 같은 명령의 출력이 0행이 아님(diff가 비어 있지 않다 — 비면 **계측 불가**).
2. **빌드** — docs-site 빌드가 **warning 없이** 성공한다(저장소의 docs-site verify 레시피).

추가로, 비-영어 페이지의 수정 문장이 **그 로케일의 자연스러운 원어**인지 확인하고 그 확인을
§E.2에 기록한다 — 이 항목은 grep으로 재지 않으며, **기록되지 않으면 PASS가 아니다**.
영어 문장이 `ko`/`ja`/`zh` 본문에 그대로 들어간 것이 발견되면 FAIL.
RED-now: 착수 전 `git diff --name-only -- docs-site/content` 는 0행 — 이 AC는 변경이 생긴
뒤에만 평가 가능하며, 변경 없이 PASS로 기록하는 것은 **계측 불가**를 PASS로 읽는 것이다.

### B.10 AC 처분 분류 (iteration 3, D14)

`verification-completeness.md` §2는 인수조건 채택에 **두 셀 쌍**(RED-now + green path)을 요구하고,
§2.1은 착수 전 RED를 **재현할 수 없는** 기준을 **regression-guard로 분류하고 release-blocking
자격에서 내린다**고 정한다. iteration 2까지 이 처분이 **적혀 있지 않았다**. 아래 표가 23개
전량의 처분이다 — 새 의무를 만들지 않고, 각 AC가 이미 어느 부류인지를 명시할 뿐이다.

| 처분 | AC | 의미 |
|---|---|---|
| **release-blocking** (RED-now 보유) | 001, 008, 010, 012, 021, 022 | 착수 전 RED가 이 트리에서 관측됐고 green path가 밀리스톤에 배정돼 있다. 릴리스를 막을 자격이 있다. |
| **regression-guard** (보존 가드) | 003, 004, 005, 007, 009, 018 | **착수 전 초록인 것이 정상이다** — 재는 것은 「예외가 뚫린 뒤에도 이 성질이 유지되는가」다. RED-now가 없는 것은 결함이 아니라 이 부류의 정의이며, §2.1에 따라 **AC 층에서는** release-blocking 자격이 없다. **단, 하위 점검으로 실린 경우는 다르다 — 아래 단서 참조.** `AC-EPE-022`의 하위 점검 `L-R1`·`L-R2`가 그 경우다(§B.9). |
| **post-change-only** (변경 이후에만 평가 가능) | 006, 011, 014, 015, 016, 017, 023 | 대상 변경이 존재해야 평가가 성립한다(미러 동기화·재생성·중립성·선행 SPEC 화해·로케일 패리티). 착수 전에는 **계측 불가**이며, 변경 없이 PASS로 기록하는 것은 계측 불가를 PASS로 읽는 것이다. |
| **width-parameterized** (폭 미해소 시 닫히는 형태가 정해져 있음) | 002, 013 | `REQ-EPE-006`이 미해소면 002는 **평가 불가**, 013은 **미측정**으로 닫힌다(`AC-EPE-019`가 이 처분을 강제한다). plan 단계는 이 상태로 닫힐 수 있다. **[2026-09-21] 폭이 (B)로 확정됐으므로 이 대체 경로는 발화하지 않는다** — 002는 집합 `{verdict.md}`로 평가되고, 013은 (B) 갈래로 측정된다. 부류 자체는 그대로 둔다: 폭이 입력이라는 성질은 값이 정해져도 바뀌지 않는다. |
| **process-record** (절차 기록형) | 019, 020 | 기계 측정이 아니라 **기록의 존재**를 잰다 — 폭이 암묵 확정되지 않았다는 기록, §3.4 전수 열거가 참이라는 확인. 기록되지 않는 것이 FAIL이다. |

[HARD] **regression-guard 단서 — 하위 점검은 자기를 실은 AC의 처분을 따른다 (iteration 5, D21).**
위 행의 「release-blocking 자격 없음」은 **AC 하나가 통째로 그 부류일 때**의 규정이다. 하위
점검으로 실린 경우에는 그렇지 않다: `L-R1`·`L-R2`는 `AC-EPE-022`의 **PASS 조건 네 줄 중
둘**이고(§B.9), `AC-EPE-022`는 release-blocking이다. 따라서 **`S1 = 0` · `S2 = 0`이어도
`L-R1` 또는 `L-R2`가 0이 아니면 `AC-EPE-022`가 FAIL하고, 그것만으로 릴리스가 막힌다.**

이것은 결함이 아니라 **D16이 요구한 바로 그 동작**이다 — D16은 남김-없음의 근거를 분할
항등식(green 시점에 동어반복으로 퇴화한다)에서 **직접 측정한 잔차·중복**으로 옮겼고, 그
측정이 차단력을 갖지 못하면 옮긴 의미가 없다. 부류 이름이 말하는 것은 **RED-now의 모양**
(착수 전 초록이 정상)이지 **차단력**이 아니며, 이 두 축이 한 낱말에 겹쳐 있던 것이 D21이
지목한 빈 칸이다.

판별식: **처분 표의 행은 AC 단위로 읽는다. 하위 점검의 차단력은 그것을 실은 AC의 처분에서
나온다.**

6 + 6 + 7 + 2 + 2 = **23** — `acceptance.md`의 AC 전량과 일치한다(`grep -cE '^\*\*AC-EPE-[0-9]+\*\*'`
→ **23**, 이번 실행 관측). 어느 AC도 분류 없이 남지 않았다.

[HARD] **`AC-EPE-023`의 재분류 (iteration 4, D17).** iteration 3 판은 `AC-EPE-023`을
release-blocking에 넣었는데, 같은 AC의 RED-now 문장이 스스로 「변경이 생긴 뒤에만 평가
가능하며 … **계측 불가**」라고 선언한다 — release-blocking의 정의(「착수 전 RED가 **이
트리에서 관측**됐고」)와 정면으로 어긋난다. `verification-completeness.md` §2.1은 착수 전
RED를 재현할 수 없는 기준의 release-blocking 자격을 **박탈**하므로, 이 표가 §2.1을 명시하려고
만들어졌으면서 그 절을 어긴 자리였다. 고친 방향은 「자기 서술을 문언 그대로 따른다」이고 —
post-change-only의 정의가 `AC-EPE-023`의 자기 서술과 축자 일치한다 — 측정 불가를 측정된 것처럼
바꿔 쓰는 반대 방향이 아니다. 합계 **23은 불변**이며, 옮긴 것은 한 칸뿐이다.

---

## §C 엣지 케이스

- **추적 중인 판정서** — 54건 중 일부는 이미 인덱스에 있다. 이들에 대한 P1은 `--no-index`
  덕에 여전히 유효하지만, P2는 출력이 없다(이미 추적 중이므로 untracked 목록에 없다). 추적
  파일로 P2 기반 AC를 만들지 않는다.
- **`--ignored` 없는 `git status`의 침묵** — 무시된 경로는 보이지 않는다. 그 0행을 「문제
  없음」으로 읽는 것이 `spec.md` §1.0에서 실측된 오독 경로다.
- **worktree vs primary** — 폭 수치(§1.4)는 primary에서만 의미가 있다. 이 워크트리에서
  `find .moai/reports -type f | wc -l` 은 54를 내며 폭의 근거가 아니다.
- **규칙 순서** — 삽입 위치를 옮기면 AC-EPE-004(누수)와 AC-EPE-001이 동시에 뒤집힐 수 있다.
  순서를 바꾸면 B.1~B.2 전량을 재측정한다. 특히 `:294`(무리 D)가 후보 삽입 지점보다 **아래**에
  있으므로, 삽입 위치에 따라 `plan-audit/`의 결정 규칙이 바뀐다 — AC-EPE-004가 `-v` 줄번호를
  기록하게 한 이유다.
- **docs-site 인벤토리와 로케일** — 번역된 토큰을 놓치는 grep은 조용히 0을 낸다. AC-EPE-022의
  양성 대조(41파일)가 그 침묵을 가른다.

## §D Definition of Done

- [ ] B.1~B.9 전 AC PASS, 각 AC의 양성 대조가 비어 있지 않음
- [ ] 제자리 측정(AC-EPE-006)이 픽스처가 아닌 저장소 `.gitignore`에서 수행됨
- [ ] `.gitignore` 줄수 변화가 의도한 추가/삭제와 정확히 일치(우발적 손실 없음)
- [ ] `make agents-emit-check` exit 0, `make build` 성공
- [ ] docs-site 4-로케일 패리티 count == 4, warning-free 빌드
- [ ] `progress.md` §E.2에 5절 형식 증거 기록 — AC-EPE-012의 **대체 문언**과 AC-EPE-023의
      **로케일 확인**이 인용돼 있을 것
- [ ] 닫지 못한 것은 §E.3 Gaps에 명시(특히: 이 카드 자신의 증거 경로 반출 여부)
