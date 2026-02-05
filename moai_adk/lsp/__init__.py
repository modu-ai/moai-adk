# LSP Module
"""MoAI-ADK LSP (Language Server Protocol) client implementation.

This module provides:
- LSP data models (Position, Range, Diagnostic, etc.)
- JSON-RPC 2.0 protocol implementation
- LSP server lifecycle management
- LSP client interface for diagnostics, references, and more
"""

from moai_adk.lsp.client import (
    LanguageSession,
    LSPClientError,
    LSPServerNotInitializedError,
    LSPTimeoutError,
    MoAILSPClient,
)
from moai_adk.lsp.models import (
    Diagnostic,
    DiagnosticSeverity,
    DocumentSymbol,
    HoverInfo,
    Location,
    Position,
    Range,
    SymbolKind,
    TextDocumentIdentifier,
    TextDocumentPositionParams,
    TextEdit,
    WorkspaceEdit,
)
from moai_adk.lsp.protocol import (
    ContentLengthError,
    JsonRpcError,
    JsonRpcNotification,
    JsonRpcRequest,
    JsonRpcResponse,
    LSPProtocol,
    ProtocolError,
)
from moai_adk.lsp.server_manager import (
    LSPServer,
    LSPServerConfig,
    LSPServerManager,
    ServerNotFoundError,
    ServerStartError,
)

__all__ = [
    # Client
    "MoAILSPClient",
    "LSPClientError",
    "LSPTimeoutError",
    "LSPServerNotInitializedError",
    "LanguageSession",
    # Models
    "Position",
    "Range",
    "Location",
    "DiagnosticSeverity",
    "Diagnostic",
    "TextDocumentIdentifier",
    "TextDocumentPositionParams",
    "TextEdit",
    "WorkspaceEdit",
    "HoverInfo",
    "DocumentSymbol",
    "SymbolKind",
    # Protocol
    "LSPProtocol",
    "JsonRpcRequest",
    "JsonRpcResponse",
    "JsonRpcNotification",
    "JsonRpcError",
    "ProtocolError",
    "ContentLengthError",
    # Server Manager
    "LSPServerManager",
    "LSPServer",
    "LSPServerConfig",
    "ServerNotFoundError",
    "ServerStartError",
]
