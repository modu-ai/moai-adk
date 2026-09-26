# t1250 — 같은 제공자의 MCP 서버 두 벌 활성 경고 (doctor)

card: t1250 · class B (run → sync) · branch WT-mcp-dup-warn · base develop c630de892

## 1. 재현 (수리 전, 2026-09-26)

측정 환경: primary checkout 세션, profile `CLAUDE_CONFIG_DIR=/Users/goos/.moai/claude-profiles/moai-adk`.

| 관측 | 명령 | 출력 |
|---|---|---|
| 프로젝트 `.mcp.json` 에 `context7` 선언 | `jq '.mcpServers|keys' .mcp.json` | `context7`, `chrome-devtools`, `playwright`, `moai` |
| claude.ai Context7 커넥터 연결 이력 | `jq '.claudeAiMcpEverConnected' $CLAUDE_CONFIG_DIR/.claude.json` 에서 context7 필터 | `["claude.ai Context7"]` |
| 이 프로젝트에서 어느 쪽도 끄지 않음 | 해당 프로젝트 항목의 `disabledMcpServers` 에서 context7 필터 | 0건 (항목 자체가 null 인 profile 도 있음) |
| 이 세션에 도구 두 벌 적재 | 세션의 지연 도구 목록 | `mcp__context7__query-docs`/`resolve-library-id` 와 `mcp__claude_ai_Context7__query-docs`/`resolve-library-id` 가 함께 나열 |
| 현행 doctor 는 못 잡음 | `moai doctor --check "MCP Scope Duplicates"` | `Pass 1 Warn 0 Fail 0` |

원인: 현행 `checkMCPScopeDuplicates` (`internal/cli/doctor.go:666`) 는 프로젝트 `.mcp.json` 과 `~/.claude/.mcp.json` 두 파일의 **같은 이름**만 비교한다. claude.ai 커넥터 이름은 `claude.ai Context7` 이고 도구 접두사도 `mcp__claude_ai_Context7__` 이므로, 이름 비교로도 Claude Code 의 자동 중복 억제로도 걸리지 않는다.

## 2. 수리

커밋 `2c3b1d951` (manager-develop, TDD). 새 advisory 점검 **MCP Provider Duplicates** — `internal/cli/doctor_mcp_provider.go`, `doctor.go` 에서 `MCP Scope Duplicates` 바로 뒤에 등록.

- 로컬 서버(프로젝트 `.mcp.json`, `~/.claude/.mcp.json`, 상태 파일의 사용자 범위 `mcpServers`)와 `claudeAiMcpEverConnected` 커넥터를 정규화 이름(소문자, `[a-z0-9]`)으로 비교
- 상태 파일은 `$CLAUDE_CONFIG_DIR/.claude.json`, 없으면 `~/.claude.json`
- 프로젝트의 `disabledMcpServers` 에 어느 쪽이든 있으면 제외
- 파일 없음·파싱 실패 → ok (fail-open, FAIL 없음)
- 부수 수정: 점검 이름 허용 목록(`binary_lag_test.go`), 골든 캡처에서 `CLAUDE_CONFIG_DIR` 비움, 골든 3개 재생성(내용 변화는 새 행 1개와 Pass 21→22, 나머지는 열 폭 재배치)
- 템플릿·`.mcp.json` 미변경(코드에 특정 제공자 이름 없음 — 중립). 안내 문구는 경고 Detail 에 있다.

sync: `CHANGELOG.md` [Unreleased] › Added 항목 1개. doctor 점검을 이름으로 나열하는 사용자 문서가 없어 docs-site 변경은 없다(`grep -rl "MCP Scope Duplicates" docs-site README*.md internal/template/templates` → 0건).

## 3. 검증 (이 트리, 2026-09-26)

| 항목 | 명령 | 결과 |
|---|---|---|
| 수리 후 발화(①) | `go run ./cmd/moai doctor --check "MCP Provider Duplicates" --verbose` | `warn … context7 (.mcp.json) + claude.ai Context7`, `Pass 0 Warn 1 Fail 0` (`doctor-e2e.txt`) — 수리 전 `Pass 1 Warn 0` 대비 |
| 양성 대조 | 임시 `CLAUDE_CONFIG_DIR` (`claude.ai Context7`, 비활성 없음) | warn (`control-pos.txt`) |
| 음성: 커넥터 없음 | 커넥터 `claude.ai Notion` 만 | `ok … no overlap` (`control-noconn.txt`) |
| 음성: 로컬명 비활성 | `disabledMcpServers: ["context7"]` | ok (`control-dis-local.txt`) |
| 음성: 커넥터명 비활성 | `disabledMcpServers: ["claude.ai Context7"]` | ok (`control-dis-conn.txt`) |
| 대상 테스트 | `go test ./internal/cli/ -run 'MCPProvider\|MCPScope\|ResolveClaudeStatePath\|BinaryLag_DoctorCheckNameSet\|DoctorGolden' -count=1` | `ok … 0.868s` (`targeted.txt`) |
| vet / gofmt | `go vet ./internal/cli/` · `gofmt -l internal/cli/` | exit 0 · 출력 없음 |
| 패키지 전체 | `go test ./internal/cli/ -count=1 -timeout 15m` | **판정 없음**: `--- FAIL` 0건, 15분 타임아웃(실행 중 테스트 `TestTodoSweepSelectorMatchesFamily`, load avg ~23). 그 테스트 단독 → `ok 0.788s`. 이전 에이전트 전체 실행도 10분 타임아웃(다른 무관 테스트, 단독 PASS). 전체 판정은 develop push 후 CI 몫 |

잔여 위험: `claudeAiMcpEverConnected` 는 연결 이력이라 계정에서 지운 커넥터도 경고될 수 있다. 이름 매칭이라 다른 이름의 같은 제공자는 놓친다. 프로젝트 로컬 `projects[..].mcpServers` 와 플러그인 서버는 보지 않는다. 음성 대조의 비활성 경우 메시지는 커넥터 없음과 같은 `no overlap` 이라 둘을 출력으로는 구분할 수 없다.
