# ADHOC-TOOL-ENABLE-CODEX-20260910 검증 보고서

## Claim

- `moai tool enable codex`가 프로젝트 Codex wiring을 추가하거나 갱신하는 정식 명령으로 등록됐다.
- `--dry-run`은 `.codex/hooks.json`, `.codex/config.toml`, `.moai/state/codex-wiring.json` 작업을 예고하고 파일을 쓰지 않는다.
- `moai update --add-codex`는 제거하지 않고 deprecated 호환 alias로 남았으며 새 명령을 안내한다.
- 변경 영향 범위의 Go 테스트와 lint가 통과했다.

## Evidence

### RED

Command:

```text
go test ./internal/cli -run 'TestToolEnableCodex_|TestUpdateAddCodex_IsDeprecatedAlias|TestInitAddCodexGuidancePrefersToolCommand' -count=1
```

Output:

```text
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/tool_enable_codex_test.go:14:9: undefined: newToolCmd
internal/cli/tool_enable_codex_test.go:42:12: undefined: runToolEnableCodexAt
internal/cli/tool_enable_codex_test.go:60:12: undefined: runToolEnableCodexAt
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
```

### GREEN 및 회귀 테스트

Command:

```text
go test ./internal/cli -run 'TestToolEnableCodex_|TestUpdateAddCodex_|TestInitAddCodexGuidance' -count=1
```

Output:

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	0.973s
```

Command:

```text
go test ./internal/cli ./internal/codexwiring -count=1 -cover
```

Output:

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	507.366s	coverage: 81.4% of statements
ok  	github.com/modu-ai/moai-adk/internal/codexwiring	0.731s	coverage: 89.6% of statements
```

### CLI 런타임

Command:

```text
go run ./cmd/moai tool enable codex --project-root /Users/goos/MoAI/moai-adk-go/.claude/worktrees/tool-enable-codex --dry-run
```

Output:

```text
Dry-run moai tool enable codex wiring plan (nothing written):
  - create-or-refresh .codex/hooks.json (merged hook render, whitelist-gated)
  - create-or-refresh .codex/config.toml ([mcp_servers.moai] + [tui].status_line, create-if-absent merge)
  - create-or-refresh .moai/state/codex-wiring.json (trust sidecar, sha256 of the generated content)
  - run without --dry-run to apply
```

Command:

```text
go run ./cmd/moai update --add-codex --check
```

Output:

```text
Flag --add-codex has been deprecated, use `moai tool enable codex` instead

   ERROR

  --Check and --add-codex are mutually exclusive (--check is informational; --add-codex mutates project wiring).

exit status 1
```

### 정적 검사

Command:

```text
golangci-lint run ./internal/cli/... ./internal/codexwiring/...
```

Output:

```text
0 issues.
```

Command:

```text
git diff --check
```

Output: 없음, exit 0.

## Baseline-attribution

이 작업은 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/tool-enable-codex`의 `WT-tool-enable-codex` 브랜치에서 수행했다. 시작 시 HEAD는 `84fa4ece4`였고 `git rev-list --count --left-right origin/develop...HEAD` 출력은 `0 0`이었다.

## Gaps

- 저장소 전체 `go test ./...`는 실행하지 않았다. 저장소 계약에 따라 변경 영향 범위인 `internal/cli`와 `internal/codexwiring`만 검증했다.
- 실제 사용자 프로젝트를 대상으로 한 non-dry-run 프로세스 실행은 하지 않았다. 같은 wiring helper의 임시 디렉터리 생성 동작은 단위 테스트에서 검증했다.
- 원격 CI와 PR 검증은 수행하지 않았다. 요청 범위는 로컬 `develop` 병합까지다.

## Residual-risk

- deprecated 플래그는 pflag 정책에 따라 help에서 숨겨질 수 있으나 기존 호출은 계속 파싱된다.
- 추후 `tool disable/status`를 추가할 때 현재 `tool` 명령 트리의 의미와 출력 규약을 별도로 정의해야 한다.
