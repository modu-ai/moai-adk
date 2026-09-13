# t624 H01 재현 — develop 기준 트리

- card: t624
- 측정 트리: `.claude/worktrees/t624` HEAD `d5dc42959` (부모 `d060e0d13` + 로컬 develop `7beba0342` 흡수)
- 대상 훅: `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` (로컬 `.claude/hooks/moai/` 사본과 `diff -q` exit 0, 바이트 동일)
- 픽스처: 세션 스크래치의 격리 git 저장소 `t624-fx` (`go.mod` + `main.go`), 커밋 제목은 모두 `docs(fx): sync-phase close…`
- 환경: 같은 호출 안에서 `unset MOAI_SYNC_GATE_BLOCKING MOAI_AUTONOMY_TIER`, `CLAUDE_PROJECT_DIR=<픽스처>`, stdin `/dev/null`

## 관측

| 호출 | HEAD | 코드 상태 | exit | stdout | 로그 줄 |
|---|---|---|---|---|---|
| call1 | `0a9aa6d` | `undefinedSymbol()` — 빌드 실패 | 0 | 237B block JSON | `decision=block go vet=1 go build=1` |
| call2 | `0a9aa6d` (동일) | 동일 | 0 | **0B** | **추가 없음** |
| ctrlA | `124e484` | `anotherUndefined()` — 빌드 실패 | 0 | 237B block JSON | `decision=block` |
| ctrlB1 | `c08b6e4` | 빈 `main` — 성공 | 0 | 0B | `decision=allow` |
| ctrlB2 | `c08b6e4` (동일) | 동일 | 0 | 0B | 추가 없음 |

call1 뒤 `.moai/state/sync-quality-gate.last` = `0a9aa6d8b3b1fc5864dc17a4accf93e22bfea69c`.

call1 stdout 전문:

```
{"hookSpecificOutput":{"hookEventName":"Stop","decision":"block","reason":"go vet failed"},"systemMessage":"sync-phase quality gate BLOCKED: go vet failed (go vet=1 go build=1 deps_modified=1). Detail: .moai/logs/sync-quality-gate.log"}
```

## 판독

- 결함은 **같은 HEAD 에서 실패한 결과를 다음 호출에 다시 전달하지 않는 것** 하나다. call2 는 검사도, 로그도, 출력도 없이 끝났다.
- 원인 줄: 훅 176–186줄이 검사 **전에** HEAD SHA 만 기록하고, 179–182줄이 같은 SHA 를 보면 결과와 무관하게 `exit 0` 한다.
- 대조군 A: 입력(HEAD)이 바뀌면 다시 검사해 차단한다 — 재실행 조건 자체는 정상.
- 대조군 B: 성공 뒤 같은 HEAD 재호출이 조용한 것은 정상 동작이다.

## 미관측

- 실제 Claude Code Stop 이벤트에서 call2 에 해당하는 턴이 어떻게 해석되는지(차단 상한·`stop_hook_active` 상호작용)는 이 픽스처로 재지 않았다.
- 검사가 60초 타임아웃에 걸려 중단된 경우(기록만 남고 결과가 없는 상태)는 재현하지 않았다.
