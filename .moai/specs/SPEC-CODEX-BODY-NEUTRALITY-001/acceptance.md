# SPEC-CODEX-BODY-NEUTRALITY-001 — 수락 기준

## §A. 계수 함정 (판정 전 반드시 읽는다)

**[HARD] 모집단은 `.codex/agents/moai/*.toml` 본문뿐이다.** 같은 패턴을 `.claude/agents/moai/*.md` 전체 파일에 돌리면 `Task(Create|Update|List|Get)` 가 **4 가 아니라 48** 로 나온다. 차이는 `.md` 프론트매터의 `tools:` CSV 이고, emitter 는 프론트매터를 `developer_instructions` 에 싣지 않는다. `.md` 전수 grep 을 근거로 쓴 판정은 **거짓 계수**이며, 이 SPEC 에서 가장 쉽게 만들어지는 오류다.

**[HARD] 단위를 붙인다.** 합집합 패턴의 **발생 수는 84**, **서로 다른 줄 수는 81** 이다(한 줄이 두 토큰을 싣는 경우가 있다). `grep -o | wc -l` 과 `grep -c` 는 다른 값을 낸다. 모든 판정은 명령과 단위를 함께 기록한다.

**[HARD] 기준선의 `= 74` 는 덧셈 착오다.** `.moai/reports/t497/measurement.md` 의 파일별 값은 정확히 재현되지만 그 합은 84 다(`14+13+13+10+8+5+5+5+4+4+3`). 74 를 기대값으로 쓰는 판정은 실패한다.

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
- **When** 분류 행 수를 세고 같은 실행에서 모집단을 재측정하면,
- **Then** 분류 행 수와 모집단 발생 수가 **둘 다 84** 로 같다. 두 값을 **같은 판정에서 함께** 기록한다 — 한쪽만 재면 모집단이 바뀐 경우를 못 잡는다.

### AC-CBN-002 — 오케스트레이터 주어 줄 4건이 산문으로 판정돼 있다 (must-pass) · maps REQ-CBN-002, REQ-CBN-003

- **Given** 분류표가 있다,
- **When** `plan-auditor.toml:146` · `super-advisor.toml:62` · `sync-auditor.toml:131` · `manager-develop.toml:64` 네 좌표의 행을 읽으면,
- **Then** 네 행 모두 `verdict=prose` 이고, 앞의 두 행은 `subject=orchestrator` 를 싣는다. 각 행의 `rationale` 이 왜 이 에이전트의 행위가 아닌지를 한 문장으로 적는다.

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

### AC-CBN-006 — 결속표 행 수가 실측된 부재 수와 일치한다 (must-pass) · maps REQ-CBN-004

- **Given** M1 이 `.moai/reports/t497/capability-absence.md` 에 클래스별 판정을 남겼다,
- **When** `grep -c '^| ' AGENTS.md` 를 돌리면,
- **Then** 출력은 `4 + (absent 로 판정된 클래스 수)` 다. 셈의 근거: 이 셀렉터는 `^| ` (파이프+공백)이라 헤더 1행과 본문 3행만 잡고 **구분자 행 `|---|---|---|` 은 잡지 않는다** — 착수 전 실측 `grep -c '^| ' AGENTS.md` → **4**. `absent` 가 0건이면 출력은 **4** 이고, 이것도 통과다 — 행 추가가 아니라 **파생이 지켜졌는가**가 판정 대상이다.

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
- **When** `git status --short` 와 `git diff --stat` 을 돌리면,
- **Then** 모든 변경 경로가 `plan.md` §F M4 의 허용 목록 안이고, 변경 파일 수가 M2·M3 이 명시한 대상 수와 일치한다. 목록 밖 경로가 있으면 **커밋 전에 보고**하고 판정을 멈춘다.

### AC-CBN-012 — 손댄 패키지만 검증했다 (must-pass) · maps REQ-CBN-014

- **Given** run-phase 검증 기록이 있다,
- **When** 실행한 명령 목록을 읽으면,
- **Then** `go test ./...` 형태의 전체 스위트 실행이 없다. `./internal/template/agentemit/...` 와 `./internal/config/`(예산 가드) 두 범위만 나타난다.

---

## §D. 경계 사례

- **부재가 하나도 실측되지 않는 경우.** AC-CBN-006 은 출력 5 로 통과하고, AC-CBN-003 은 41 불변으로 통과한다. 이때 41줄의 덮개는 본문 1문장이며, 그 문장의 존재를 `grep -c '.agents/skills'` 로 확인한다.
- **측정 자체가 불가능한 경우(중첩 codex 프로브 불가).** `capability-absence.md` 에 `unmeasurable` 과 그 근거를 적고 행을 추가하지 않는다. **"측정 못 했으니 없다고 본다"는 금지** — 부재는 관측돼야 한다.
- **M2 분류 중 모집단이 84 에서 벗어나는 경우.** 다른 카드의 `.md` 변경이 흘러들어온 것이다. 판정을 멈추고 `git status` 로 귀속한다.
- **manager-lead 자기 스폰 줄에 결속 참조를 붙일 때 M1 이 행을 만들지 않은 경우.** 존재하지 않는 행을 가리키는 참조는 만들지 않는다 — 본문에 대체 행동을 직접 적는다.

---

## §E. 품질 게이트

- `go test ./internal/template/agentemit/...` → PASS
- `go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$'` → PASS
- `gofmt -l internal/` → 무출력(코드 변경이 있을 때만 해당; 이 카드는 기본적으로 문서·생성물 변경이다)
- 판정 기록은 5절 형식(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)으로 남긴다.

---

## §F. 완료 정의

- AC-CBN-001..012 전부 PASS, 각 판정에 명령과 축자 출력이 붙어 있다.
- `.moai/reports/t497/body-classification.md` 와 `capability-absence.md` 가 존재하고 커밋돼 있다(증거는 재측정 **앞에** 커밋한다 — 그래야 트리 동일성 델타가 성립한다).
- `[NEEDS CLARIFICATION]` 마커 2건이 Implementation Kickoff Approval 전에 해소돼 있다.
- M5(미러 스킬 77파일)는 운영자가 착수를 지시하지 않는 한 완료 정의에 들어가지 않는다.
