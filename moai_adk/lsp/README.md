# MoAI-ADK LSP Module

Language Server Protocol (LSP) client implementation for MoAI-ADK.

## Overview

This module provides a complete LSP client implementation for integrating language intelligence features into MoAI-ADK workflows. It enables real-time diagnostics, code navigation, and refactoring capabilities across multiple programming languages.

## Architecture

```
lsp/
├── __init__.py          # Public API exports
├── client.py            # High-level LSP client interface
├── models.py            # LSP data models (Position, Range, Diagnostic, etc.)
├── protocol.py          # JSON-RPC 2.0 protocol implementation
├── server_manager.py    # LSP server lifecycle management
└── README.md            # This file
```

## Quick Start

### Basic Usage

```python
from moai_adk.lsp import MoAILSPClient, Position

async def main():
    # Initialize client with project root
    client = MoAILSPClient("/path/to/project")

    try:
        # Get diagnostics for a file
        diagnostics = await client.get_diagnostics("src/main.py")
        for diag in diagnostics:
            print(f"{diag.severity.name}: {diag.message} at line {diag.range.start.line}")

        # Find references to a symbol
        refs = await client.find_references("src/main.py", Position(line=10, character=5))
        print(f"Found {len(refs)} references")

        # Go to definition
        definitions = await client.go_to_definition("src/main.py", Position(line=10, character=5))
        for loc in definitions:
            print(f"Definition at {loc.uri}:{loc.range.start.line}")
    finally:
        await client.cleanup()
```

### Configuration

Create a `.lsp.json` file in your project root:

```json
{
  "python": {
    "command": "pyright-langserver",
    "args": ["--stdio"],
    "extensionToLanguage": {
      ".py": "python",
      ".pyi": "python"
    }
  },
  "typescript": {
    "command": "typescript-language-server",
    "args": ["--stdio"],
    "extensionToLanguage": {
      ".ts": "typescript",
      ".tsx": "typescriptreact",
      ".js": "javascript",
      ".jsx": "javascriptreact"
    }
  }
}
```

## API Reference

### MoAILSPClient

The main client interface for LSP operations.

#### Constructor

```python
MoAILSPClient(
    project_root: str | Path,
    request_timeout: float = 30.0
)
```

#### Methods

| Method | Description | Returns |
|--------|-------------|---------|
| `get_diagnostics(file_path)` | Get diagnostics (errors, warnings) for a file | `list[Diagnostic]` |
| `find_references(file_path, position)` | Find all references to a symbol | `list[Location]` |
| `rename_symbol(file_path, position, new_name)` | Rename a symbol across the project | `WorkspaceEdit` |
| `get_hover_info(file_path, position)` | Get hover information for a symbol | `HoverInfo \| None` |
| `go_to_definition(file_path, position)` | Navigate to symbol definition | `list[Location]` |
| `get_document_symbols(file_path)` | Get all symbols in a document | `list[DocumentSymbol]` |
| `get_language_for_file(file_path)` | Get language identifier for a file | `str \| None` |
| `ensure_server_running(language)` | Start server for a language if needed | `None` |
| `cleanup()` | Stop all servers and clean up resources | `None` |

### Data Models

#### Position

```python
@dataclass
class Position:
    line: int       # Zero-based line number
    character: int  # Zero-based character offset
```

#### Range

```python
@dataclass
class Range:
    start: Position
    end: Position

    def contains(self, position: Position) -> bool: ...
    def is_single_line(self) -> bool: ...
```

#### Location

```python
@dataclass
class Location:
    uri: str       # File URI (file:///path/to/file)
    range: Range   # Range within the file
```

#### Diagnostic

```python
@dataclass
class Diagnostic:
    range: Range
    message: str
    severity: DiagnosticSeverity
    source: str | None = None
    code: str | int | None = None

    def is_error(self) -> bool: ...
```

#### DiagnosticSeverity

```python
class DiagnosticSeverity(IntEnum):
    ERROR = 1
    WARNING = 2
    INFORMATION = 3
    HINT = 4
```

#### DocumentSymbol

```python
@dataclass
class DocumentSymbol:
    name: str
    kind: SymbolKind
    range: Range
    selection_range: Range
    children: list[DocumentSymbol]
```

#### WorkspaceEdit

```python
@dataclass
class WorkspaceEdit:
    changes: dict[str, list[TextEdit]]

    def file_count(self) -> int: ...
```

### Exceptions

| Exception | Description |
|-----------|-------------|
| `LSPClientError` | Base exception for LSP client errors |
| `LSPTimeoutError` | Request timeout exceeded |
| `LSPServerNotInitializedError` | Server not initialized before operation |
| `ServerNotFoundError` | Language server config not found |
| `ServerStartError` | Failed to start language server |
| `ProtocolError` | JSON-RPC protocol error |
| `ContentLengthError` | Invalid Content-Length header |

## Supported Language Servers

| Language | Recommended Server | Package |
|----------|-------------------|---------|
| Python | Pyright | `pyright` |
| TypeScript/JS | TypeScript Language Server | `typescript-language-server` |
| Rust | rust-analyzer | `rust-analyzer` |
| Go | gopls | `golang.org/x/tools/gopls` |
| Java | Eclipse JDT LS | `eclipse.jdt.ls` |
| C/C++ | clangd | `clangd` |

## Testing

Run the LSP module tests:

```bash
# Run all LSP tests
uv run pytest tests/lsp/ -v

# Run specific test file
uv run pytest tests/lsp/test_client.py -v

# Run with coverage
uv run pytest tests/lsp/ --cov=src/moai_adk/lsp
```

### Test Coverage

- `test_client.py` - Client creation, diagnostics, references, rename, hover, definition, symbols
- `test_models.py` - Position, Range, Location, Diagnostic, TextEdit, WorkspaceEdit, HoverInfo
- `test_protocol.py` - JSON-RPC messages, encoding/decoding, async streams
- `test_server_manager.py` - Server config, lifecycle, language detection

## Integration with MoAI-ADK

### With Ralph Engine (Autonomous Loop)

```python
from moai_adk.lsp import MoAILSPClient
from moai_adk.core.ralph_engine import RalphEngine

async def run_with_lsp():
    client = MoAILSPClient("/project")
    engine = RalphEngine(lsp_client=client)

    # Ralph uses LSP for real-time error detection
    await engine.run_autonomous_loop()
```

### With Expert Agents

The LSP module integrates with expert agents for enhanced code intelligence:

- **expert-debug**: Uses diagnostics for error analysis
- **expert-refactoring**: Uses rename and references for safe refactoring
- **manager-quality**: Uses diagnostics for quality validation

## Protocol Details

### JSON-RPC 2.0

The module implements the JSON-RPC 2.0 specification for LSP communication:

```python
# Request
{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "textDocument/definition",
    "params": { ... }
}

# Response
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": { ... }
}
```

### Message Framing

LSP uses Content-Length headers for message framing:

```
Content-Length: 52\r\n
\r\n
{"jsonrpc":"2.0","id":1,"method":"initialize",...}
```

## Version History

- **v1.3.9** - Current version with full LSP 3.17 support
- Added DocumentSymbol and SymbolKind support
- Improved async server management
- Enhanced error handling with typed exceptions

## License

MIT License - Part of MoAI-ADK project
