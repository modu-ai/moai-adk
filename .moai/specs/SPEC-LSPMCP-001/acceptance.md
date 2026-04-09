# SPEC-LSPMCP-001 Acceptance Criteria

## AC-LSPMCP-001: CLI subcommand exists
Given moai binary is built
When `moai mcp lsp --help` is run
Then help text shows LSP MCP server description

## AC-LSPMCP-002: MCP stdio server starts
Given gopls is installed
When `moai mcp lsp` is started with stdin JSON-RPC
Then server responds to `initialize` and `tools/list` requests
And 6 tools are listed

## AC-LSPMCP-003: goto_definition works
Given a Go project with gopls available
When `tools/call` with `goto_definition` for a known symbol
Then response contains the definition file path and position

## AC-LSPMCP-004: find_references works
Given a Go project with gopls available
When `tools/call` with `find_references` for an exported function
Then response contains all reference locations

## AC-LSPMCP-005: Auto-registration in .mcp.json
Given `moai init` is run on a new project
When `.mcp.json` is generated
Then it contains `"moai-lsp"` server entry with command "moai" args ["mcp", "lsp"]

## AC-LSPMCP-006: Missing server error
Given no Python language server is installed
When `diagnostics` is called for a .py file
Then response contains error with pyright installation instructions

## AC-LSPMCP-007: DiagnosticsProvider integration
Given MCP LSP bridge is wired in deps.go
When `diagnosticsCollector.Collect()` is called
Then it uses the LSP bridge (not nil) with CLI fallback

## AC-LSPMCP-008: Unit tests pass
Given all new code in internal/mcp/
When `go test ./internal/mcp/...` runs
Then all tests pass with >= 85% coverage
