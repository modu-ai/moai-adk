# SPEC-CODEX-BODY-NEUTRALITY-001 — 수락 기준

## §A. 계수 함정 (판정 전 반드시 읽는다)

**[HARD] 모집단은 `.codex/agents/moai/*.toml` 본문뿐이다.** 같은 패턴을 `.claude/agents/moai/*.md` 전체 파일에 돌리면 `Task(Create|Update|List|Get)` 가 **4 가 아니라 48** 로 나온다. 차이는 `.md` 프론트매터의 `tools:` CSV 이고, emitter 는 프론트매터를 `developer_instructions` 에 싣지 않는다. `.md` 전수 grep 을 근거로 쓴 판정은 **거짓 계수**이며, 이 SPEC 에서 가장 쉽게 만들어지는 오류다.

**[HARD] 단위를 붙인다.** 합집합 패턴의 **발생 수는 84**, **서로 다른 줄 수는 81** 이다(한 줄이 두 토큰을 싣는 경우가 있다). `grep -o | wc -l` 과 `grep -c` 는 다른 값을 낸다. 모든 판정은 명령과 단위를 함께 기록한다.

**[HARD] 기준선의 `= 74` 는 덧셈 착오다.** `.moai/reports/t497/measurement.md` 의 파일별 값은 정확히 재현되지만 그 합은 84 다(`14+13+13+10+8+5+5+5+4+4+3`). 74 를 기대값으로 쓰는 판정은 실패한다.

**[HARD] 결속표 행 수의 셀렉터는 하나뿐이고 기대값은 3이다.**

```
sed -n '/^\*\*Capability bindings/,/^---$/p' AGENTS.md | grep -c '^| [a-z]'
```

`grep -c '^| ' AGENTS.md` 는 **쓰지 않는다.** 그 셀렉터는 (a) 결속표 밖의 `| ` 줄까지 세고 (b) 헤더 행(`| Capability …`)을 포함해 3행 표에 4를 낸다. 앞 라운드에서 같은 결과에 `4`·`3`·`5` 세 기대값이 적혔던 원인이 그 한 자리다. 위 셀렉터는 구역을 좁히고 `^| [a-z]` 로 헤더(대문자 C)와 구분자(`|---`)를 함께 제외해 **데이터 행만** 센다 — 착수 전 실측 **3**, 두 사본 동일.

---

## §B. 사전 고정 기준선 (이 트리, `ace1c5440`)

| 명령 | 기대 출력 |
|---|---|
| `find internal/template/templates/.codex/agents -name '*.toml' \| wc -l` | 11 |
| `grep -rhoE 'AskUserQuestion\|TaskCreate\|TaskUpdate\|TaskList\|TaskGet\|DesignSync\|Skill\(\|Agent\(' internal/template/templates/.codex/agents/moai/*.toml \| wc -l` | 84 |
| `grep -rhoE 'invoke Skill\(' ... \| wc -l` | 41 |
| `grep -rhoE 'AskUserQuestion' ... \| wc -l` | 4 |
| `grep -rhoE 'Task(Create\|Update\|List\|Get)' ... \| wc -l` | 4 |
| `grep -rhoE 'DesignSync' ... \| wc -l` | 6 |
| `grep -rlE 'AskUserQuestion\|Agent\(\|TaskCreate\|TaskUpdate\|TaskList\|TaskGet\|DesignSync' internal/template/templates/.claude/skills/ \| wc -l` | 77 |

---

## §C. Given-When-Then 판정 (AC 행렬)

### AC-CBN-001 — 분류표가 모집단 전수를 덮는다 (must-pass) · maps REQ-CBN-001, REQ-CBN-003, REQ-CBN-015

- **Given** M2 가 `.moai/reports/t497/body-classification.md` 를 산출했다,
- **When** 분류 행 수를 세고, 같은 실행에서 모집단을 재측정하고, 같은 실행에서 좌표 집합을 대조하면,
- **Then** 세 조건이 모두 성립한다. (a) 분류 행 수 = **84**(발생 단위). (b) 모집단 발생 수 = **84**. (c) 분류표 `file:line` 열의 정렬본이 `grep -rnoE '<§B 합집합 패턴>' internal/template/templates/.codex/agents/moai/*.toml | cut -d: -f1-2 | sort -u`(= **81** 개 서로 다른 줄)와 `diff` 에서 무출력·rc 0. **(c) 가 없으면 84행이 전부 같은 좌표를 가리켜도 통과한다** — (a)는 개수만, (c)는 대응을 본다. 세 값을 **같은 판정에서 함께** 기록한다.
- 착수 전 실측(RED): 분류표 파일이 없으므로 (a)(c) 모두 불성립. (b) 는 84.

### AC-CBN-002 — `AskUserQuestion` 3줄이 산문으로 판정돼 있다 (must-pass) · maps REQ-CBN-002, REQ-CBN-003, REQ-CBN-015

- **Given** 분류표가 있다,
- **When** `sync-auditor.toml:131` · `plan-auditor.toml:146` · `super-advisor.toml:62` **세 좌표**의 행을 읽으면,
- **Then** 세 행 모두 `verdict=prose` 이고, `subject` 는 각각 `prohibition` · `orchestrator` · `orchestrator` 다. `plan-auditor.toml:146` 은 한 줄에 두 발생을 실으므로 표에서 **2행**을 차지하고 두 행 모두 같은 판정을 갖는다. 각 행의 `rationale` 이 왜 이 에이전트의 행위가 아닌지를 한 문장으로 적는다.
- **[HARD] 단위.** `AskUserQuestion` 모집단은 **서로 다른 줄 3 / 발생 4** 다. 앞 라운드는 발생 4를 좌표 4로 읽어 `manager-develop.toml:64` 를 네 번째 좌표로 못박았으나, **그 파일의 `AskUserQuestion` 발생 수는 0**이다(이 트리 재측정: `grep -c 'AskUserQuestion' …/manager-develop.toml` → 0). `manager-develop.toml:64,65,66` 은 `Agent(` 축의 좌표이고 이 AC 의 대상이 아니다.
- 검증(같은 판정에서 함께): `grep -rn 'AskUserQuestion' internal/template/templates/.codex/agents/moai/*.toml | wc -l` → **3**(줄), `grep -rho 'AskUserQuestion' … | wc -l` → **4**(발생). 두 값이 이 셋과 어긋나면 모집단이 움직인 것이므로 판정을 멈춘다.
- `subject` 열은 **세 값**을 갖는다(`this-agent` · `orchestrator` · `prohibition`). `sync-auditor.toml:131` 은 금지 서술이라 앞의 두 값 중 어느 쪽도 아니며, 두 값만 허용하는 표는 이 행에서 거짓이 된다.

### AC-CBN-003 — 41줄이 손상되지 않았다 (must-pass · 전면 치환 차단) · maps REQ-CBN-011

- **Given** M3·M4 가 끝났다,
- **When** `grep -rhoE 'invoke Skill\(' internal/template/templates/.codex/agents/moai/*.toml | wc -l` 을 돌리면,
- **Then** 출력은 **41** 이다. 40 이하이면 REQ-CBN-011 위반이며 실패다.

### AC-CBN-004 — 산문 `AskUserQuestion` 4건이 손상되지 않았다 (must-pass · 전면 치환 차단) · maps REQ-CBN-010

- **Given** M3·M4 가 끝났다,
- **When** `grep -rhoE 'AskUserQuestion' internal/template/templates/.codex/agents/moai/*.toml | wc -l` 을 돌리면,
- **Then** 출력은 **4** 이다. 0 이면 일괄 치환이 일어난 것이며 실패다. **이 AC 는 일부러 "줄어들면 실패" 방향으로 쓰였다** — 중립화가 삭제로 통과하는 길을 막는다.

### AC-CBN-005 — Task\* 지시 3줄이 클래스 이름으로 바뀌었다 (must-pass) · maps REQ-CBN-009

- **Given** M3·M4 가 끝났다,
- **When** `grep -rhoE 'Task(Create|Update|List|Get)' internal/template/templates/.codex/agents/moai/*.toml | wc -l` 을 돌리면,
- **Then** 출력은 **0** 이고, 같은 판정에서 `grep -rn 'task-list' internal/template/templates/.codex/agents/moai/*.toml` 이 **3개 이상의 좌표**를 낸다. 두 번째 조건이 없으면 "그냥 지웠다"가 통과한다.

### AC-CBN-006 — 파생 기록이 존재하고 결속표와 맞물린다 (must-pass) · maps REQ-CBN-004, REQ-CBN-016

- **Given** M1 이 `.moai/reports/t497/capability-absence.md` 에 `tool_classes` 전수 판정을 남겼다,
- **When** 아래 네 명령을 **같은 판정에서 함께** 돌리면,
  ```
  (a) grep -cE '\|[[:space:]](absent|present)[[:space:]]\|[[:space:]]*$' .moai/reports/t497/capability-absence.md
  (b) grep -cE '\|[[:space:]]absent[[:space:]]\|[[:space:]]*$'           .moai/reports/t497/capability-absence.md
  (c) sed -n '/^\*\*Capability bindings/,/^---$/p' AGENTS.md | grep -c '^| [a-z]'
  (d) grep -c '현재 측정값 4행' .moai/specs/SPEC-CODEX-SKILL-NEUTRAL-001/spec.md
  ```
- **Then** (a) = **11**(`tool_classes` 값 집합 전수), (b) = **3**, (c) = **3**, **(b) == (c)**, (d) = **0**.
- **착수 전 실측(RED, 이 트리 `845dd65af`):** (a)(b) — 파일이 없어 `grep` rc=2, 무출력. (c) — 3. (d) — **1**. **(a)(b)(d) 세 조건이 오늘의 트리에서 성립하지 않으므로, 아무것도 하지 않으면 이 AC 는 통과할 수 없다.** 앞 라운드의 이 AC 는 `grep -c '^| ' AGENTS.md` → 4 하나뿐이라 **착수 전에 이미 초록**이었다.
- **왜 산출물을 직접 묶는가.** 앞 라운드의 `Given` 은 `capability-absence.md` 를 이름 붙였으나 `Then` 이 그 파일을 읽지 않아, 파일이 없어도 통과했다. (a)(b) 가 그 구멍을 닫는다.
- **셀렉터 주의.** (a)(b) 는 「verdict 열이 마지막 칸」이라는 표 모양에만 의존하고 헤더 대소문자에 기대지 않는다. `^| [a-z]` 로 세면 헤더가 `| class …` 처럼 소문자로 시작할 때 헤더까지 세어 12를 낸다. 대조군: (a) 의 셀렉터를 `AGENTS.md` 에 돌리면 **0**(이 트리 실측) — 다른 표를 우연히 세지 않는다.
- **행 0 추가는 정당한 결과다.** 이 AC 가 묻는 것은 행이 늘었는가가 아니라 **파생이 지켜졌는가**이며, (b) == (c) 가 그 판정이다.

### AC-CBN-007 — 두 `AGENTS.md` 사본의 표 구역이 바이트 동일하다 (must-pass) · maps REQ-CBN-006

- **Given** M1 이 결속표를 건드렸거나 건드리지 않았다,
- **When** `diff <(sed -n '/^\*\*Capability bindings/,/^---$/p' AGENTS.md) <(sed -n '/^\*\*Capability bindings/,/^---$/p' internal/template/templates/AGENTS.md)` 를 돌리면,
- **Then** 출력이 없고 종료코드가 0 이다. 한쪽만 고친 상태는 드리프트이며 실패다.

### AC-CBN-008 — 새 어휘가 생기지 않았다 (must-pass) · maps REQ-CBN-005

- **Given** 결속표가 변경됐다,
- **When** 표의 첫 칸 값 집합과 `agents-codex.yaml` `tool_classes` 의 값 집합을 대조하면,
- **Then** 표의 모든 능력 이름이 `tool_classes` 값 집합의 원소다. `reference-loader` 처럼 값 집합 밖의 이름이 하나라도 있으면 실패다.

### AC-CBN-009 — always-loaded 예산 가드가 통과한다 (must-pass) · maps REQ-CBN-007, REQ-CBN-015

- **Given** 결속표가 변경됐다,
- **When** `go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v` 를 돌리면,
- **Then** `PASS` 이고 로그의 `headroom` 이 양수다. 판정 기록에 **셀렉터와 매치 수**를 함께 적는다(0매치 셀렉터의 `no tests to run` 초록은 판정이 아니다). 착수 전 실측값: `budget 77600, headroom 3065`.

### AC-CBN-010 — 골든이 드리프트 없이 정합한다 (must-pass) · maps REQ-CBN-008, REQ-CBN-012

- **Given** M4 가 `AGENTEMIT_UPDATE=1` 로 재생성했다,
- **When** UPDATE 없이 `go test ./internal/template/agentemit/...` 를 돌리면,
- **Then** `PASS` 다. sha256 불일치가 하나라도 있으면 실패다.

### AC-CBN-011 — 변경 반경이 의도한 경로 안이다 (must-pass) · maps REQ-CBN-008, REQ-CBN-013

- **Given** 재생성이 끝났고 아직 커밋 전이다,
- **When** 변경 경로 집합에서 증거 두 디렉터리를 걸러낸 나머지를 `spec.md` §C.5 산출물 반경 11줄(정렬본)과 대조하면,
  ```
  git status --short | awk '{print $NF}' \
    | grep -v '^\.moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/' \
    | grep -v '^\.moai/reports/t497/' | sort   >   /tmp/t497-radius-actual.txt
  diff /tmp/t497-radius-actual.txt <(sort <<'EOF'
  ...spec.md §C.5 의 11줄을 그대로...
  EOF
  )
  ```
- **Then** `diff` 가 무출력이고 rc 0 이다 — **집합 동일성**. 걸러진 나머지 경로는 전부 `.moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/` 또는 `.moai/reports/t497/` 접두여야 한다. 어느 조건이든 어긋나면 **커밋 전에 보고**하고 판정을 멈춘다.
- **왜 개수가 아니라 집합인가.** 앞 라운드의 「변경 파일 수가 M2·M3 이 명시한 대상 수와 일치」는 비교할 대상 수가 어디에도 적혀 있지 않았고, 설령 있었더라도 **개수 비교는 한 파일이 빠지고 다른 파일이 들어온 경우를 통과시킨다.** 허용 목록이 `*.md` 글롭이었던 것도 같은 문제였다 — 11본 에이전트 전부가 반경 안이라, 다른 카드의 편집이 딸려 들어와도 걸리지 않았다. `spec.md` §C.5 는 그래서 **문자 그대로의 파일 집합**이다.
- **루트 사본은 반경 밖이다.** `spec.md` §C.5 에는 저장소 루트 `.claude/agents/moai/*.md` 가 없다(REQ-CBN-008). 그 경로가 `git status` 에 나타나면 편집이 잘못된 사본에 들어간 것이므로 실패다.

### AC-CBN-012 — 손댄 패키지만 검증했다 (must-pass) · maps REQ-CBN-014

- **Given** run-phase 검증 기록이 있다,
- **When** 실행한 명령 목록을 읽으면,
- **Then** `go test ./...` 형태의 전체 스위트 실행이 없다. `./internal/template/agentemit/...` 와 `./internal/config/`(예산 가드) 두 범위만 나타난다.

### AC-CBN-013 — `manager-lead` 의 지시 줄이 **전부** 능력 이름을 부른다 (must-pass) · maps REQ-CBN-009

- **Given** M2 분류표(`.moai/reports/t497/body-classification.md`)가 있고 M3·M4 가 끝났다,
- **When** 같은 판정에서 두 값을 재면,

  ```
  N = grep -cE '^\| [^|]*manager-lead\.toml \| [0-9]+ \| Agent\( \| directive \|' \
        .moai/reports/t497/body-classification.md
  M = grep -c 'subagent-spawn' \
        internal/template/templates/.codex/agents/moai/manager-lead.toml
  ```

- **Then** 세 조건이 모두 성립한다. (a) **N ≥ 4** 이고, §B.4 가 경계 표본으로 든 `37,57,59,193` 네 좌표가 그 directive 행 집합에 **전부** 들어 있다. (b) **M ≥ N** — 개정된 줄 수가 분류가 지시로 센 줄 수 이상이다. (c) 착수 전 실측 **M = 0** — 아무것도 하지 않으면 통과할 수 없다.
- **왜 상수 4 가 아니라 N 인가.** 상수 4 에 묶으면 이 파일의 나머지 6줄 중 하나가 지시로 판정돼도 판정은 초록으로 남는다 — 코덱스 하네스에 클로드 서브에이전트 스폰을 지시하는 줄이 그대로 살아 있는데도. 그것이 이 SPEC 이 막으려는 바로 그 실패다. N 에 묶으면 M2 가 지시로 센 줄이 하나라도 개정되지 않은 순간 (b) 가 깨진다. 모집단은 이 트리 실측 `grep -c 'Agent(' …/manager-lead.toml` → **10** 이고, M2 ③ 이 81줄 전수 좌표 대응을 `diff` 로 증명하므로 그 10줄 중 미분류로 남는 줄은 없다.
- **[HARD] 개정은 지시 줄을 병합하지 않는다.** 두 지시 줄을 한 줄로 합치면 M < N 이 되어 (b) 가 **거짓 RED** 를 낸다. 병합이 필요하다고 판단되면 그것은 설계 변경이므로 상신 대상이지, 판정을 느슨하게 할 사유가 아니다.
- **셀렉터는 M2 표기에 묶인다.** N 의 셀렉터는 분류표의 `token` 열이 합집합 패턴의 **리터럴 토큰**(`Agent(`)을 그대로 싣는다는 전제 위에 있다(`plan.md` §F M2 열 정의). 표기가 다르면 N 이 0 이 되어 (a) 가 먼저 깨지므로, 이 어긋남은 조용히 통과하지 않는다.
- **결속행 참조를 만들지 않는다.** `subagent-spawn` 은 능력 존재로 확정됐으므로 결속표에 행이 없다(`spec.md` §A.2 정정 2). M2 분류표가 directive 로 센 N 줄은 능력 이름을 부르고 코덱스 쪽 대체 행동을 본문에 직접 적는다.

### AC-CBN-014 — `manager-design` 사다리에 능력 부재 시 행동이 붙었다 (must-pass) · maps REQ-CBN-009

- **Given** M3·M4 가 끝났다,
- **When** `grep -oE '(^|[^/])design-sync' internal/template/templates/.codex/agents/moai/manager-design.toml | wc -l` 과 `grep -c 'default = DesignSync tool push' …/manager-design.toml` 을 같은 판정에서 돌리면,
- **Then** 앞은 **1 이상**(발생 단위, 착수 전 실측 **0**), 뒤는 **1 불변**(사다리 문면 무손상). 앞의 셀렉터가 `[^/]` 를 요구하는 이유: 이 파일의 기존 `design-sync` 4건은 **전부 `/design-sync` 슬래시 커맨드**이므로 맨 `grep -c 'design-sync'` 는 착수 전에도 4를 내어 공허하다.
- **`(^|` 갈래가 없으면 안 되는 이유.** `[^/]design-sync` 만 쓰면 토큰 **앞에 한 글자**를 요구하므로 줄 **머리**에 오는 `design-sync` 를 못 잡는다. 요구한 문단이 실재해도 0 이 나와 **거짓 RED** 가 된다 — 안전한 방향의 결함이지만(거짓 PASS 는 만들지 않는다) run-phase 왕복을 한 번 문다. 두 형태 모두 착수 전 실측 **0** 이므로 기대값은 바뀌지 않는다.

---

## §D. 경계 사례

- **행이 0개 추가되는 경우 — 이것이 예상 경로다.** `spec.md` §A.2 정정 2 가 부재 3건을 확정했고 그 3건은 이미 표에 있으므로, 결속표는 바뀌지 않는다. 그때 AC-CBN-006 은 (b)=3 · (c)=3 · (b)==(c) 로 통과하고, AC-CBN-003 은 41 불변으로 통과한다. **기대값은 3 하나뿐이다** — 앞 라운드가 같은 경로에 `4`(AC-CBN-006) · `3`(plan.md) · `5`(이 절) 세 값을 적었던 것이 §A 가 셀렉터를 하나로 못박은 이유다.
- **41줄의 덮개는 AC-CBN-003 이 아니라 별도 대상이다.** 덮개는 `AGENTS.md` 두 사본에 실리는 1문장이며, 그 존재는 `grep -c '\.agents/skills' AGENTS.md` → **1 이상**(착수 전 실측 **0**), 그리고 템플릿 사본에 대해 **같은 값**으로 확인한다. 앞 라운드의 `grep -c '.agents/skills'` 는 경로 인자가 없고 `.` 가 임의 문자라 **실행 자체가 되지 않았다**; 이 형태는 돌아가고, 기준선 0에서 출발하므로 공허하지 않다.
- **코덱스 프로브를 돌릴 수 없는 경우 — 해당 없음.** M1 은 프로브가 아니라 문면 대조이므로 이 경계 사례는 사라졌다. rationale 이 낡았을 가능성은 남지만(§A.2 정정 2 Residual-risk), 그 위험은 행을 **더하지 않는** 방향이라 보수적이다. 여전히 금지되는 것은 **"측정 못 했으니 없다고 본다"** 이며, 부재 주장의 근거는 언제나 rationale 원문 인용이다.
- **M2 분류 중 모집단이 84 에서 벗어나는 경우.** 다른 카드의 `.md` 변경이 흘러들어온 것이다. 판정을 멈추고 `git status --short` 로 귀속한다.
- **`manager-lead` 자기 스폰 지시 줄(M2 분류표의 directive 행, 모집단 N).** `subagent-spawn` 은 능력 존재로 확정돼 결속행이 없으므로, 존재하지 않는 행을 가리키는 참조는 만들지 않고 **본문에 능력 이름과 대체 행동을 직접 적는다.** 이것은 선택지가 아니라 유일한 처분이며, AC-CBN-013 이 그 결과를 양성으로 잡는다(착수 전 0). 앞 라운드의 이 항목은 「M1 이 행을 만들지 않았으면 건너뛴다」로 읽혀 REQ-CBN-009 의 한 부류가 아무 판정 없이 빠져나갈 수 있었다 — 그 허용을 제거한다.

---

## §E. 품질 게이트

- `go test ./internal/template/agentemit/...` → PASS
- `go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$'` → PASS
- `gofmt -l internal/` → 무출력(코드 변경이 있을 때만 해당; 이 카드는 기본적으로 문서·생성물 변경이다)
- 판정 기록은 5절 형식(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)으로 남긴다.

---

## §F. 완료 정의

- AC-CBN-001..014 전부 PASS, 각 판정에 명령·축자 출력·단위(발생 수 / 줄 수)가 붙어 있다.
- `.moai/reports/t497/body-classification.md` 와 `capability-absence.md` 가 존재하고 커밋돼 있다(증거는 재측정 **앞에** 커밋한다 — 그래야 트리 동일성 델타가 성립한다).
- `SPEC-CODEX-SKILL-NEUTRAL-001` REQ-CSN-003 의 「현재 측정값 4행」이 3행으로 정정되고 그 SPEC HISTORY 에 Amendments 1행이 남아 있다(AC-CBN-006 (d)).
- 미해결 마커 **0건** — `grep -rnE '\[NEEDS[[:space:]]CLARIFICATION' .moai/specs/SPEC-CODEX-BODY-NEUTRALITY-001/` 가 **무출력·rc 1**(이 실행에서 실행). 셀렉터 자신이 마커 모양이 아니므로 이 완료 정의 항목이 자기 자신을 세지 않는다 — 넓은 `'NEEDS'` 로 잡으면 이 줄이 걸려 완료 정의가 영원히 불만족이 된다. ① 은 철회, ② 는 `spec.md` §D 의 운영자 결정 기록으로 전환됐다(`plan.md` §E).
- M5(미러 스킬 77파일)는 **완료 정의에 들어가지 않는다** — 운영자가 2026-09-07 에 후속 카드 분리로 확정했다(`spec.md` §D · `plan.md` §E ②).
