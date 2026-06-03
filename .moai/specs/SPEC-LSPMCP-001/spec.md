---
id: SPEC-LSPMCP-001
version: "1.0.0"
status: superseded
superseded_by: SPEC-LSPMCP-RETIRE-001
created: "2026-04-07"
updated: 2026-06-03
author: GOOS
priority: P2
issue_number: 0
title: "LSP MCP Bridge (Go 바이너리 내장)"
phase: "v2.x - Legacy"
module: "lsp"
lifecycle: completed
tags: "legacy"
---

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-04-07 | 1.0.0 | Initial draft |

---

# SPEC-LSPMCP-001: LSP MCP Bridge (Go 바이너리 내장)

## Overview

`moai mcp lsp` 서브커맨드로 MCP stdio 서버를 시작하여, Claude Code 에이전트가 LSP 기반 시맨틱 코드 인텔리전스 도구(goto definition, find references, hover 등)를 사용할 수 있게 함. 기존 `internal/lsp/` 패키지의 nil DiagnosticsProvider 슬롯을 채움.

## Motivation

현재 에이전트는 Grep/Glob 텍스트 기반 탐색만 가능. `internal/lsp/models.go`에 `DiagnosticsProvider` 인터페이스가 정의되어 있으나 `deps.go` line 94에서 nil로 전달됨. CLI fallback(`go vet`, `ruff`, `tsc`)은 진단만 제공하고 시맨틱 탐색은 불가. 별도 MCP 서버가 아닌 Go 바이너리에 내장하여 단일 배포 유지.

## Requirements (EARS Format)

### REQ-LSPMCP-001 (Ubiquitous)
The system shall provide a `moai mcp lsp` CLI subcommand that starts an MCP stdio server implementing JSON-RPC 2.0 protocol.

### REQ-LSPMCP-002 (Ubiquitous)
The MCP server shall expose 6 tools: `goto_definition`, `find_references`, `hover`, `document_symbols`, `diagnostics`, `rename`.

### REQ-LSPMCP-003 (Ubiquitous)
The MCP server shall auto-detect and connect to language servers (gopls, typescript-language-server, pyright) based on project indicator files.

### REQ-LSPMCP-004 (Event-Driven)
When `moai init` or `moai update` is run, the system shall register `moai-lsp` in `.mcp.json` with `{"command": "moai", "args": ["mcp", "lsp"]}`.

### REQ-LSPMCP-005 (State-Driven)
When no language server is available for the requested file type, the MCP tool shall return a clear error message with installation instructions.

### REQ-LSPMCP-006 (Ubiquitous)
The MCP server shall manage language server lifecycle (start on first request, shutdown on MCP connection close).

### REQ-LSPMCP-007 (Ubiquitous)
The system shall satisfy the existing `lsp.DiagnosticsProvider` interface from `internal/lsp/models.go`, replacing the current nil slot in `deps.go`.

## Affected Files

### New Files
- `internal/mcp/server.go` - MCP stdio server (JSON-RPC 2.0)
- `internal/mcp/handler.go` - MCP tool dispatch
- `internal/mcp/tools.go` - Tool definitions (6 tools)
- `internal/mcp/lsp/bridge.go` - LSP client manager
- `internal/mcp/lsp/languages.go` - Language server auto-detection
- `internal/mcp/lsp/transport.go` - LSP stdio transport
- `internal/mcp/server_test.go`
- `internal/mcp/lsp/bridge_test.go`
- `internal/mcp/lsp/languages_test.go`
- `internal/cli/mcp.go` - Cobra subcommand

### Modified Files
- `internal/template/templates/.mcp.json.tmpl` - Add moai-lsp server
- `internal/cli/deps.go` - Wire MCP LSP bridge as DiagnosticsProvider
- `go.mod` - Add JSON-RPC 2.0 dependency

## Technical Design

### CLI Pattern (follows existing Cobra pattern)

```go
// internal/cli/mcp.go
var mcpCmd = &cobra.Command{Use: "mcp", Short: "MCP server commands", GroupID: "tools"}
var mcpLSPCmd = &cobra.Command{Use: "lsp", Short: "Start LSP MCP stdio server", RunE: runMCPLSP}
func init() { mcpCmd.AddCommand(mcpLSPCmd); rootCmd.AddCommand(mcpCmd) }
```

### Architecture

```
Claude Code <--stdio JSON-RPC 2.0--> moai mcp lsp <--LSP stdio--> gopls/tsserver/pyright
```

### Language Server Detection

| Language | Indicator | Server | Install |
|----------|-----------|--------|---------|
| Go | go.mod | gopls | go install golang.org/x/tools/gopls@latest |
| TypeScript | tsconfig.json, package.json | typescript-language-server --stdio | npm i -g typescript-language-server |
| Python | pyproject.toml, setup.py | pyright --stdio | pip install pyright |

### DiagnosticsProvider Wiring

```go
// deps.go: Replace nil with MCP bridge
lspProvider := mcplsp.NewBridge(projectDir)
diagnosticsCollector := lsphook.NewDiagnosticsCollector(lspProvider, fallbackDiags)
```

## Dependencies

- SPEC-LSP-001 (Completed): Uses existing `internal/lsp/` types and interfaces


### Out of Scope

- N/A (legacy SPEC)

## Non-Goals

- Custom language server implementation
- All languages (Phase 1: Go, TypeScript, Python)
- Real-time LSP notifications (request-response only)
- IDE-level features (completion, signature help)
