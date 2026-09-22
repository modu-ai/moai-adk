# t220 — handle-stop-goal.sh가 goal 평가기를 매 turn 여러 번 돌리던 문제

card: t220 · branch: `WT-stop-goal-double-exec` · base: `origin/main@62eeda07e`

## Claim

1. 래퍼가 goal 평가기를 **매 turn 3번** 발화했다(리드 보고는 2번 — 3계층 전부 존재하면 3번이고, `$HOME/.local/bin/moai`까지 있는 환경이 그렇다).
2. 원인은 `printf … | exec <bin>` — `exec`은 파이프라인 **우변 서브셸**을 대체할 뿐 래퍼를 대체하지 않아, 래퍼가 살아남아 다음 계층 `if`로 떨어진다.
3. 배포된 3개 사본 전부 동일 패턴이었고, 3개 모두 고쳤다.
4. 범위 내에 **제3의 `exec` 변종은 없다**.

## Evidence

### 재현 (수정 전 트리, RED)

```
$ go test ./internal/hook/ -run TestStopGoalWrapper -count=1
--- FAIL: TestStopGoalWrapperFiresEvaluatorExactlyOnce/.claude/hooks/moai/handle-stop-goal.sh
    invoked the goal evaluator 3 time(s) [tier1-PATH tier2-HOME-GO-BIN tier3-HOME-LOCAL-BIN]; want exactly 1
--- FAIL: …/internal/template/templates/.claude/hooks/moai/handle-stop-goal.sh
    invoked the goal evaluator 3 time(s) [tier1-PATH tier2-HOME-GO-BIN tier3-HOME-LOCAL-BIN]; want exactly 1
--- FAIL: …/internal/template/templates/.claude/hooks/moai/handle-stop-goal.sh.tmpl
    invoked the goal evaluator 3 time(s) [tier1-PATH tier2-HOME-GO-BIN tier3-HOME-LOCAL-BIN]; want exactly 1
FAIL	github.com/modu-ai/moai-adk/internal/hook	10.945s
```

실행 기반 계수다 — 문자열 grep이 아니라 3개 계층에 각각 계수 스텁을 놓고 래퍼를 실제로 돌려 호출 라벨을 기록했다.

**기존 테스트가 왜 못 잡았는가**: `TestAC001_GoalAbsentSkipsMoaiBinary`는 이미 "정확히 1"을 단언하지만 PATH 계층에만 스텁을 두고 `$HOME`는 빈 임시 디렉터리로 둔다 — 폴스루가 찾을 게 없어 관측되지 않았다. 계층을 **전부 채우는 것**이 이 결함을 관측 가능하게 만드는 조건이다.

### 수정 후 (GREEN)

```
$ go test ./internal/hook/ -run 'TestStopGoal|TestAC001' -count=1
ok  	github.com/modu-ai/moai-adk/internal/hook	3.102s

$ go test ./internal/template/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	62.560s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.968s

$ go vet ./internal/hook/ ./internal/template/   # 무출력
$ bash -n .claude/hooks/moai/handle-stop-goal.sh # syntax ok
$ go build ./...                                  # build ok
```

### 제3 변종 스윕

```
$ git grep -ln 'printf.*| *exec' -- '*.sh' '*.tmpl'
.claude/hooks/moai/handle-stop-goal.sh
internal/template/templates/.claude/hooks/moai/handle-stop-goal.sh
internal/template/templates/.claude/hooks/moai/handle-stop-goal.sh.tmpl

$ git grep -nE '\(\s*exec |exec [^|]*\|\||exec [^|]*&&|\| *exec ' -- '.claude/hooks' 'internal/template/templates/.claude/hooks'
→ 위 3개 파일의 해당 라인 + sync-phase-quality-gate.sh의 `find … -exec`(셸 exec 아님) 뿐

$ grep -rlE 'exec ' .claude/hooks/moai/ | xargs grep -lE '^\s*[a-z_]+\(\)'
.claude/hooks/moai/sync-phase-quality-gate.sh    # 함수 내부 exec은 find -exec 뿐
```

나머지 래퍼 31개는 bare `exec moai hook <event>` — 셸을 통째로 대체하므로 폴스루가 불가능하다. stdin을 잡는 형제 래퍼들(`chain-event`, `handle-codex-review-gate`, `handle-multi-review-gate`)은 이미 `MOAI_BIN` 선해결 + 단일 파이프 형태였다 — 이번 수정이 그 하우스 스타일에 합류한 것이지 새 패턴을 만든 게 아니다.

**즉, 제3의 변종은 없다.** 근거는 위 3개 명령의 출력이며, 훑은 범위는 `.claude/hooks/**` 와 `internal/template/templates/.claude/hooks/**` 다.

## Baseline-attribution

전부 이 트리(`.claude/worktrees/t220`), base `62eeda07e`, 이번 실행에서 측정. RED는 수정 **전**, GREEN은 수정 **후** 같은 트리.

## 변경 내용

- `.claude/hooks/moai/handle-stop-goal.sh` + 템플릿 `.sh` + 템플릿 `.sh.tmpl` (3개 사본, 바이트 동일 유지) — 3계층을 **먼저 해결**한 뒤 `printf '%s' "$INPUT" | "$MOAI_BIN" hook stop-goal` 한 번. 왜 계층별 `exec`이면 안 되는지 주석으로 고정.
- `internal/template/renderer.go` — `$MOAI_BIN`을 `claudeCodePassthroughTokens`에 추가. 렌더러가 미확장 셸 변수를 차단하는데, 이 목록엔 이미 같은 부류(래퍼 셸 지역변수: `$INPUT`, `$SESSION_ID`, `$STATE_FILE`, `$PROJECT_ROOT`)가 등록돼 있다. 누락 시 `TestLLMPanelTemplate_Integration`이 `unexpanded dynamic token detected: found "$MOAI_BIN"`로 실패한다 — 실제로 한 번 걸렸고 그 게이트가 옳게 작동한 것이다.
- `internal/hook/stop_goal_single_exec_test.go` (신규) — 계수 스텁 기반 단일 발화 단언 + 3개 사본 바이트 동일성 잠금(Template-First: 배포되는 건 `.tmpl`이라 한쪽만 고치면 다음 `moai update`가 되돌린다).

동작 계약 보존: 계층 우선순위(PATH → `$HOME/go/bin` → `$HOME/.local/bin`), 미발견 시 조용한 `exit 0`, stderr 로그 리다이렉트, 항상 exit 0(차단 결정은 stdout JSON이 나른다 — 평가기 자체가 non-zero를 반환하지 않는다) 전부 그대로다.

## Gaps

- 실제 Claude Code 세션에서 `turns_used` 증가가 turn당 1로 돌아오는지는 **관측하지 않았다**. 기계적 근거는 래퍼 실행 계수(3→1)이고, `turns_used` 이중 증가는 그 계수로부터의 추론이다.
- 리드가 보고한 지연 수치(76ms → 31ms)는 재측정하지 않았다.
- 전체 스위트는 로컬에서 돌리지 않았다(레인 규율). 영향 패키지는 `internal/hook/...`, `internal/template/...`.

## Residual-risk

- `$MOAI_BIN` 패스스루 등록은 렌더러의 미확장 토큰 검사를 그 이름에 한해 완화한다. 같은 이름의 진짜 미확장 템플릿 변수가 생기면 잡히지 않는다 — 목록의 기존 7개 래퍼 지역변수와 동일한 성격의 위험이다.
- 3개 사본 동일성은 새 테스트가 잠그지만, 네 번째 사본이 생기면 `stopGoalWrapperCopies()`에 손으로 추가해야 한다.
