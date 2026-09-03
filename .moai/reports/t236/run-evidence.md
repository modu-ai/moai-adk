# t236 run evidence — RED / GREEN / Gates

측정 트리: 워크트리 `.claude/worktrees/t236`, 브랜치 `WT-project-dir-stale`.
RED 구간은 구현 커밋 전, 트리 HEAD `65196a5a7` (origin/develop tip) + 테스트 파일만
추가한 상태에서 측정. GREEN/Gates 구간은 구현 후 동일 워크트리, 동일 세션.
모든 go test/vet/build 호출은 단일 compound `unset MOAI_KANBAN MOAI_KANBAN_ID
MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && <cmd>`
형태로 실행.

## RED (pre-fix)

### Stage A-1 — internal/hook (branch missing)

명령: `unset … && go test -C <worktree> ./internal/hook/ -run 'TestPostToolWorktreeMove' -count=1`
exit code: **1**

```
--- FAIL: TestPostToolWorktreeMove_StampsEnvFile (0.00s)
    post_tool_worktree_test.go:40: expected a systemMessage about the spawn-frozen MCP server root, got &{Continue:<nil> StopReason: SystemMessage: SuppressOutput:false Decision: Reason: HookSpecificOutput:0x684003b964d0 UpdatedInput: Retry:false ExitCode:0 WorktreePath: Data:[123 34 115 101 115 115 105 111 110 95 105 100 34 58 34 115 101 115 115 45 119 116 34 44 34 116 111 111 108 95 110 97 109 101 34 58 34 69 110 116 101 114 87 111 114 107 116 114 101 101 34 125]}
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.436s
FAIL
```

Red 이유: `EnterWorktree` PostToolUse가 정상 경로로 흘러 SystemMessage가 비어 있음 —
worktree-move 분기 부재 (잘못된 이유 아님: 파라미터 무시가 아니라 분기 부재가
테스트의 명시적 대상).

### Stage A-2 — internal/cli behavioral (param ignored / provenance absent)

명령: `unset … && go test -C <worktree> ./internal/cli/ -run 'TestSpecProgress_ResponseCarriesRootProvenance|TestVerifySnapshot_HonorsProjectRoot|TestVerifyTrend_HonorsProjectRoot' -count=1`
exit code: **1**

```
--- FAIL: TestSpecProgress_ResponseCarriesRootProvenance (0.00s)
    mcp_project_root_test.go:368: spec_progress response lacks param provenance; got={"content":[{"type":"text","text":"spec_progress: ok"}],"structuredContent":{"count":1,"specs":[…SPEC-PROBEPROV-908…]}}
    mcp_project_root_test.go:376: spec_progress fallback resolution carried no warning; got={"content":[…],"structuredContent":{"count":0,"specs":[]}}
--- FAIL: TestVerifySnapshot_HonorsProjectRoot (0.00s)
    mcp_project_root_test.go:414: snapshot missing under the named project_root — the parameter was ignored and the record landed on the fallback tree
--- FAIL: TestVerifyTrend_HonorsProjectRoot (0.00s)
    mcp_project_root_test.go:463: verify_trend response lacks param provenance; got={…}
    mcp_project_root_test.go:468: verify_trend without project_root saw a check that exists only in the probe tree; got={…}
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.938s
FAIL
```

Red 이유 (각):
- spec_progress/verify_trend: 응답에 `_root` 프리비넌스 없음 (live gap L2).
- verify_snapshot: `project_root` 무시 — record가 fallback 트리에 착지 (live gap L3).
- verify_trend no-param: fallback에 record가 있어 보임 → 파라미터가 리다이렉트하지
  않았다는 직접 증거.

### Stage B — resolver source test (symbol absent, compile red)

명령: `unset … && go test -C <worktree> ./internal/cli/ -run 'TestResolveToolProjectRootWithSource_ParamAndFallback' -count=1`
exit code: **1**

```
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/mcp_project_root_test.go:480:37: undefined: resolveToolProjectRootWithSource
internal/cli/mcp_project_root_test.go:487:10: undefined: rootProvenanceMap
internal/cli/mcp_project_root_test.go:495:47: undefined: resolveToolProjectRootWithSource
internal/cli/mcp_project_root_test.go:502:18: undefined: rootProvenanceMap
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
```

## GREEN (post-fix)

### internal/hook (full package)

명령: `unset … && go test -C <worktree> ./internal/hook/... -count=1`
exit code: **0**

```
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	24.224s
ok  	github.com/modu-ai/moai-adk/internal/hook/security	12.501s
ok  	github.com/modu-ai/moai-adk/internal/hook/testutil	1.010s
ok  	github.com/modu-ai/moai-adk/internal/hook/trace	2.296s
```

### internal/cli (full package tree, 17 ok)

명령: `unset … && go test -C <worktree> ./internal/cli/... -count=1`
exit code: **0** (`grep -c "^ok"` → 17)

```
ok  	github.com/modu-ai/moai-adk/internal/cli/update/report	4.482s
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	8.608s
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	9.081s
```

### internal/template (full package tree — matcher·mirror-parity 표면)

명령: `unset … && go test -C <worktree> ./internal/template/... -count=1`
exit code: **0**

```
ok  	github.com/modu-ai/moai-adk/internal/template	23.708s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.834s
?   	github.com/modu-ai/moai-adk/internal/template/scripts	[no test files]
```

### Doc-parity guard (기계 판독, M3)

명령: `unset … && go test -C <worktree> ./internal/cli/ -run 'TestProjectRootDocMatchesServer' -count=1 -v`

```
--- PASS: TestProjectRootDocMatchesServer (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.967s
```

이 판정은 라이브 서버 스키마(`toolsDeclaringProjectRoot`)가 project_root를 선언한
12개 도구 목록과 문서 명단이 정렬일 때만 통과 — 문서 나열이 코드보다 1개라도
앞서거나 뒤지면 빨간다.

## Gates

### go vet (touched packages)

명령: `unset … && go vet -C <worktree> ./internal/hook/... ./internal/cli/...`
exit code: **0**, 출력 없음.

### make build (embedded FS regen)

명령: `unset … && make -C <worktree> build`
exit code: **0**

```
catalog.yaml updated successfully (12899 bytes)
go build -ldflags "-s -w -X …Commit=65196a5a7 …" -o bin/moai ./cmd/moai
```

### 검증 중 라이브 PostToolUse 발화 (미측정 — 의도된 Gap)

본 세션은 훅 설정이 세션 시작 스냅샷에 고정되어 신규 matcher
(`EnterWorktree|ExitWorktree`)를 라이브로 검증 불가. 착지 후 신규 세션에서의 첫
systemMessage 관측이 잔여 확인 항목 (재현 증거 L1의 Gaps와 동일 지점).
