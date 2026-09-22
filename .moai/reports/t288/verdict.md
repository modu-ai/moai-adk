# t288 — goal 조건 타입 오분류 (GitHub #1660)

- 카드: t288 (Class B — plan 생략, 재현 우선)
- 워크트리: `.claude/worktrees/t288`, 브랜치 `WT-goal-arm-cond-type`
- 기준: `origin/develop` `1e5199b88` (`git merge-base --is-ancestor origin/develop HEAD` → 참)

---

## 1. 카드 전제 반증 — MCP 래퍼 축이 아니다

카드와 이슈 #1660 은 원인을 "`goal_arm` MCP 래퍼가 CLI 파서를 거치지 않는다"로 지목했습니다. **거짓입니다.**

- `internal/cli/mcp_server.go:641` — `cond := parseCondition(conditionText) // same classifier the CLI uses`
- 그 줄이 들어온 커밋: `93b7adf84` (2026-08-06, #1378). 이슈 보고 환경은 v3.1.3-rc.1(2026-08-24 빌드)이므로 **보고 시점에 이미 분류기를 타고 있었습니다.**

즉 CLI 경로와 MCP 경로는 같은 함수를 쓰며, **둘 다 똑같이 오분류합니다.** 실제 결함은 그 분류기의 판별자에 있습니다.

## 2. 실제 원인 — 판별자가 canonical 산문을 놓친다

`internal/cli/goal.go` `parseCondition` 은 소문자 리터럴 토큰 **`transcript`** 하나만 보고 model 조건을 갈랐습니다.

그런데 orchestrator 가 실제로 무장하는 canonical `ac_converge` 조건(`.claude/skills/moai/workflows/run.md:148-157`)에는 그 단어가 없습니다 — `surfaced in **the conversation**` 이라고 씁니다. 그래서 문단 전체가 mechanical 로 분류되어 `cmd` 에 통째로 들어갔습니다.

REQ-GLE-032(`.moai/specs/SPEC-GOAL-ENGINE-001/spec.md:357`)의 규정 문구는 *"a natural-language claim that references **the conversation transcript**"* — 두 지시어를 함께 말합니다. 구현이 그중 한 단어만 받은 것이 누락 지점입니다.

## 3. 재현 (Claim / Evidence)

**Claim A** — canonical 산문이 mechanical 로 분류된다.

```
$ go test ./internal/cli/ -run 'TestParseCondition|TestRunmdAcConverge' -count=1
--- FAIL: TestParseCondition_CanonicalAcConvergeIsModel (0.00s)
    goal_condition_classify_test.go:40: canonical ac_converge prose classified as "mechanical"
      (cmd="Every blocking acceptance criterion in .moai/specs/SPEC-XXX/acceptance.md has its
       PASS evidence surfaced in the conversation (test output, build exit 0, or explicit
       AC-id: PASS line); AND `go test ./...` exit 0 is surfaced; AND no test file outside
       the SPEC scope was modified (surfaced via git status). Stop when all hold."), want "model"
--- FAIL: TestParseCondition_TranscriptReferentsAreModel (0.00s)
    goal_condition_classify_test.go:97: "all AC rows show PASS in the conversation" classified as "mechanical", want model
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.040s
```

**Claim B** — mechanical 경로에 들어가면 exit 2 로 매 턴엔드가 막힌다.

```
$ sh .moai/reports/t288/prose-as-shell.sh; echo "exit=$?"
prose-as-shell.sh: line 1: syntax error near unexpected token `('
exit=2
```

이슈 본문의 관측치 `mechanical condition failed: cmd "Every blocking ..." exited 2 (want 0)` 와 종료 코드가 일치합니다. 셸 구문 오류라서 실행 상황과 무관하게 항상 2 — 수렴 가능성이 0 인 차단입니다.

**격리 준수**: 재현은 단위 테스트와 독립 셸 스크립트로만 했습니다. 실세션 `goal arm` 은 하지 않았고, 따라서 해제할 goal 도 없습니다 (`goal_arm` 은 오케스트레이터 전용 도구 — 그 경계도 건드리지 않았습니다).

## 4. 수리

`internal/cli/goal.go` — 판별자를 REQ-GLE-032 가 명시한 두 지시어로 확장. 휴리스틱이 아니라 **명시 토큰 목록**을 유지합니다(요구사항이 "EXPLICIT rule, NOT an implicit heuristic" 을 못박고 있음).

```go
var modelConditionReferents = []string{"transcript", "conversation"}
```

CLI arm 경로와 MCP `goal_arm` 이 이 함수 하나를 공유하므로 **두 경로가 함께 수리됩니다.**

## 5. 회귀 (GREEN)

```
$ go test ./internal/cli/ -run 'TestParseCondition|TestRunmdAcConverge|Goal|goal' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	1.002s

$ go test ./internal/cli/... ./internal/goal/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	331.449s
ok  	github.com/modu-ai/moai-adk/internal/cli/agentlint	4.720s
ok  	github.com/modu-ai/moai-adk/internal/cli/harness	9.598s
ok  	github.com/modu-ai/moai-adk/internal/cli/pr	1.551s
ok  	github.com/modu-ai/moai-adk/internal/cli/preference	6.240s
ok  	github.com/modu-ai/moai-adk/internal/cli/printer	0.616s
ok  	github.com/modu-ai/moai-adk/internal/cli/specid	4.990s
ok  	github.com/modu-ai/moai-adk/internal/cli/taskledger	6.011s
ok  	github.com/modu-ai/moai-adk/internal/cli/uikit	3.978s
ok  	github.com/modu-ai/moai-adk/internal/cli/update	2.469s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	1.309s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/deploy	3.278s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	0.797s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/plan	5.638s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/report	3.546s
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	7.106s
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	9.616s
ok  	github.com/modu-ai/moai-adk/internal/goal	5.823s

$ go vet ./internal/cli/          → 출력 없음, exit 0
$ golangci-lint run ./internal/cli/...  → 0 issues.
```

**공허한 초록 아님**: 같은 트리에서 수리 전 §3 이 FAIL 했고 수리 후 PASS 합니다 — 판별자 한 줄이 실제 판정 축입니다.

가드 방향 양쪽: `TestParseCondition_ShellCommandsStayMechanical` 이 실행 가능한 명령(`go test ./internal/cli/...`, 후행 `exits N` 포함)이 model 로 끌려가지 않음을 고정합니다. `TestRunmdAcConvergeProseMatchesFixture` 는 run.md 원문이 드리프트하면 테스트가 침묵하지 않고 실패하게 합니다.

## 6. 두 번째 축 — 별개 결함, 이 카드에서 닫지 않음

리드가 지목한 저장 위치 축(메모리 `goal-arm-mcp-worktree-split`, t267)은 **조건 타입 축과 별개**입니다. 근거:

- `handleGoalArm` 은 `resolveProjectDir()`(`internal/cli/session.go:264` — `CLAUDE_PROJECT_DIR` → cwd)로 root 를 잡고, stop-goal 훅은 `resolveHookProjectRoot()`(`internal/cli/hook.go:681` — 같은 순서)로 잡습니다. **함수 자체는 같은 규칙**입니다.
- 분열은 규칙이 아니라 **프로세스 출처**에서 옵니다. MCP 서버는 primary 체크아웃에서 뜬 장수 프로세스라 cwd/env 가 primary 를 가리키고, 훅은 워크트리 cwd 에서 실행됩니다.

즉 조건 타입 오분류를 고쳐도 워크트리 분열은 그대로 남고, 그 반대도 마찬가지입니다. **범위를 넓히지 않았습니다** — 별도 카드 소관으로 보고합니다.

## 6b. 통합 창 재측정 (병합 트리)

창 지명 후 lane-9 단독으로 `origin/develop` `f94326b4c` 를 흡수했습니다. 카드 기준(`1e5199b88`)에서 develop 이 다섯 번 움직인 뒤라 흡수량이 30커밋입니다.

- 창 안 재측정: `git rev-list --count --left-right origin/develop...HEAD` → `30 2` (흡수 전) → `0 3` (흡수 후, 병합 커밋 `23054f9f0`)
- 충돌 1건: `CHANGELOG.md` — 양쪽 모두 `### Fixed` 최상단에 추가. **양쪽 보존**(t288 항목 먼저, t286/t369 항목 이어서). 다른 파일 충돌 0.
- **판정 도구 출처를 병합 트리로 고정**: `make build` 재실행 (`BuildID=v3.1.2-896-g23054f9f0`). PATH 설치본으로는 재지 않았습니다. 빌드 후 워킹트리 dirty 0.

병합 트리 측정값:

```
$ go test ./internal/cli/... ./internal/goal/... -count=1 -timeout 900s   → exit 0, 18/18 ok
  internal/cli 292.989s  (흡수 전 331.449s)
  전문: .moai/reports/t288/merged-tree-tests.txt
$ go vet ./internal/cli/            → 출력 없음
$ ./bin/moai spec lint              → 0 error(s), 1091 warning(s)   [로컬 라벨; CI 는 -19 = 1072 (t371 소관)]
```

SPEC Lint error 0 은 리드가 준 기대값(lane-4·lane-6·lane-8 병합 트리 일치값)과 같습니다.

**Graph Freshness 분모 구별**: 이 카드 자체 기여분은 `origin/develop...HEAD` 직접 측정으로 **5 파일**(Go 2, CHANGELOG 1, 증거 2)입니다. 흡수 병합 후 도구가 내는 `contribution:` 줄은 first parent 기준이라 **흡수분**을 세므로 이 수와 더하거나 같게 읽으면 안 됩니다. codemaps 재생성은 하지 않았습니다(배치 끝 일괄).

## 7. Gaps / 잔여 위험

- **Gaps**: CI 미판독(이 카드는 아직 push 전). darwin 로컬 단일 플랫폼에서만 측정 — windows/linux 미관측. `moai` 바이너리 재빌드/재설치 후 실경로 arm 은 하지 않음(격리 의무 준수).
- **잔여 위험**: 두 지시어 중 어느 것도 쓰지 않은 산문 조건은 여전히 mechanical 로 갑니다. 판별자가 명시 토큰 방식인 한 남는 구조적 한계이고, REQ-GLE-032 가 그 방식을 규정하고 있어 이 카드에서는 바꾸지 않았습니다. 넓히려면 요구사항 개정이 선행돼야 합니다.
- **잔여 위험**: `conversation` 이라는 단어를 포함하는 실제 셸 명령은 model 로 오분류됩니다(예: `grep conversation log.txt`). 확률은 낮지만 0 은 아니며, `exits N` 후행절을 붙여도 회피되지 않습니다.
