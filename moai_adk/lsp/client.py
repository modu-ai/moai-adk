# MoAI LSP Client
"""LSP client interface for MoAI-ADK.

This module provides the main LSP client interface for getting diagnostics,
finding references, renaming symbols, and other LSP operations.
"""

from __future__ import annotations

import asyncio
import logging
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

from moai_adk.lsp.models import (
    Diagnostic,
    DiagnosticSeverity,
    DocumentSymbol,
    HoverInfo,
    Location,
    Position,
    Range,
    SymbolKind,
    TextEdit,
    WorkspaceEdit,
)
from moai_adk.lsp.protocol import (
    JsonRpcNotification,
    JsonRpcRequest,
    LSPProtocol,
)
from moai_adk.lsp.server_manager import LSPServer, LSPServerManager

logger = logging.getLogger(__name__)


class LSPClientError(Exception):
    """Error raised by the LSP client."""

    pass


class LSPTimeoutError(LSPClientError):
    """Error raised when an LSP request times out."""

    pass


class LSPServerNotInitializedError(LSPClientError):
    """Error raised when attempting operations before initialization."""

    pass


@dataclass
class LanguageSession:
    """Tracks the state of an LSP session for a specific language.

    Attributes:
        language: The language identifier.
        server: The LSP server instance.
        protocol: The protocol handler for this session.
        initialized: Whether the server has been initialized.
        capabilities: Server capabilities received during initialization.
        open_documents: Set of currently open document URIs.
        diagnostics_cache: Cached diagnostics received from the server.
    """

    language: str
    server: LSPServer
    protocol: LSPProtocol = field(default_factory=LSPProtocol)
    initialized: bool = False
    capabilities: dict[str, Any] = field(default_factory=dict)
    open_documents: set[str] = field(default_factory=set)
    diagnostics_cache: dict[str, list[dict[str, Any]]] = field(default_factory=dict)
    _response_reader_task: asyncio.Task[None] | None = None


class MoAILSPClient:
    """LSP client for MoAI-ADK.

    Provides a high-level interface for LSP operations including:
    - Getting diagnostics for files
    - Finding references to symbols
    - Renaming symbols across the project
    - Getting hover information
    - Going to definition
    - Getting document symbols

    Attributes:
        project_root: The root directory of the project.
        server_manager: Manager for LSP server processes.
        request_timeout: Timeout for LSP requests in seconds.
    """

    # Default timeout for LSP requests
    DEFAULT_TIMEOUT: float = 30.0

    def __init__(
        self,
        project_root: str | Path,
        request_timeout: float = DEFAULT_TIMEOUT,
    ) -> None:
        """Initialize the LSP client.

        Args:
            project_root: Path to the project root directory.
            request_timeout: Timeout for LSP requests in seconds.
        """
        self.project_root = Path(project_root)
        self.server_manager = LSPServerManager()
        self.request_timeout = request_timeout
        self._sessions: dict[str, LanguageSession] = {}
        self._load_config()

    def _load_config(self) -> None:
        """Load LSP configuration from .lsp.json file."""
        config_path = self.project_root / ".lsp.json"
        if config_path.exists():
            self.server_manager.load_config(config_path)

    # ==========================================================================
    # Public API
    # ==========================================================================

    async def get_diagnostics(self, file_path: str) -> list[Diagnostic]:
        """Get diagnostics for a file.

        Args:
            file_path: Path to the file.

        Returns:
            List of Diagnostic objects.
        """
        raw_diagnostics = await self._request_diagnostics(file_path)
        return [self._parse_diagnostic(d) for d in raw_diagnostics]

    async def find_references(self, file_path: str, position: Position) -> list[Location]:
        """Find all references to the symbol at position.

        Args:
            file_path: Path to the file.
            position: Position of the symbol.

        Returns:
            List of Location objects.
        """
        raw_refs = await self._request_references(file_path, position)
        return [self._parse_location(r) for r in raw_refs]

    async def rename_symbol(self, file_path: str, position: Position, new_name: str) -> WorkspaceEdit:
        """Rename the symbol at position.

        Args:
            file_path: Path to the file.
            position: Position of the symbol.
            new_name: New name for the symbol.

        Returns:
            WorkspaceEdit with all changes.
        """
        raw_edit = await self._request_rename(file_path, position, new_name)
        return self._parse_workspace_edit(raw_edit)

    async def get_hover_info(self, file_path: str, position: Position) -> HoverInfo | None:
        """Get hover information for position.

        Args:
            file_path: Path to the file.
            position: Position to get hover info for.

        Returns:
            HoverInfo or None if not available.
        """
        raw_hover = await self._request_hover(file_path, position)
        if raw_hover is None:
            return None
        return self._parse_hover_info(raw_hover)

    async def go_to_definition(self, file_path: str, position: Position) -> list[Location]:
        """Go to the definition of a symbol at position.

        Args:
            file_path: Path to the file.
            position: Position of the symbol.

        Returns:
            List of Location objects representing definitions.
        """
        raw_definitions = await self._request_definition(file_path, position)
        # Handle both single location and array of locations
        if isinstance(raw_definitions, dict):
            return [self._parse_location(raw_definitions)]
        return [self._parse_location(loc) for loc in raw_definitions]

    async def get_document_symbols(self, file_path: str) -> list[DocumentSymbol]:
        """Get all symbols in a document.

        Args:
            file_path: Path to the file.

        Returns:
            List of DocumentSymbol objects.
        """
        raw_symbols = await self._request_document_symbols(file_path)
        return [self._parse_document_symbol(s) for s in raw_symbols]

    def get_language_for_file(self, file_path: str) -> str | None:
        """Get the language for a file.

        Args:
            file_path: Path to the file.

        Returns:
            Language identifier or None.
        """
        return self.server_manager.get_language_for_file(file_path)

    async def ensure_server_running(self, language: str) -> None:
        """Ensure an LSP server is running and initialized for a language.

        Args:
            language: Language identifier.
        """
        await self._get_or_create_session(language)

    async def cleanup(self) -> None:
        """Clean up by stopping all servers and cancelling tasks."""
        # Cancel all response reader tasks
        for session in self._sessions.values():
            if session._response_reader_task:
                session._response_reader_task.cancel()
                try:
                    await session._response_reader_task
                except asyncio.CancelledError:
                    pass

        # Stop all servers
        await self.server_manager.stop_all_servers()
        self._sessions.clear()

    # ==========================================================================
    # Session Management
    # ==========================================================================

    async def _get_or_create_session(self, language: str) -> LanguageSession:
        """Get or create a session for a language.

        Args:
            language: Language identifier.

        Returns:
            LanguageSession for the language.

        Raises:
            LSPClientError: If server fails to start or initialize.
        """
        if language in self._sessions:
            session = self._sessions[language]
            if session.server.is_running() and session.initialized:
                return session

        # Start the server
        server = await self.server_manager.start_server(language)
        session = LanguageSession(language=language, server=server)
        self._sessions[language] = session

        # Start the response reader task
        session._response_reader_task = asyncio.create_task(self._response_reader(session))

        # Initialize the server
        await self._initialize_server(session)

        return session

    async def _initialize_server(self, session: LanguageSession) -> None:
        """Initialize an LSP server with the initialize handshake.

        Args:
            session: The language session to initialize.

        Raises:
            LSPClientError: If initialization fails.
        """
        if session.initialized:
            return

        # Send initialize request
        init_params = {
            "processId": None,
            "rootUri": self._file_to_uri(str(self.project_root)),
            "rootPath": str(self.project_root),
            "capabilities": {
                "textDocument": {
                    "synchronization": {
                        "dynamicRegistration": False,
                        "willSave": False,
                        "willSaveWaitUntil": False,
                        "didSave": True,
                    },
                    "completion": {
                        "dynamicRegistration": False,
                        "completionItem": {
                            "snippetSupport": False,
                            "documentationFormat": ["plaintext", "markdown"],
                        },
                    },
                    "hover": {
                        "dynamicRegistration": False,
                        "contentFormat": ["plaintext", "markdown"],
                    },
                    "definition": {"dynamicRegistration": False},
                    "references": {"dynamicRegistration": False},
                    "documentSymbol": {
                        "dynamicRegistration": False,
                        "hierarchicalDocumentSymbolSupport": True,
                    },
                    "rename": {"dynamicRegistration": False},
                    "publishDiagnostics": {"relatedInformation": True},
                },
                "workspace": {
                    "workspaceFolders": True,
                    "applyEdit": True,
                },
            },
            "workspaceFolders": [
                {
                    "uri": self._file_to_uri(str(self.project_root)),
                    "name": self.project_root.name,
                }
            ],
        }

        result = await self._send_request(session, "initialize", init_params)
        if result is None:
            raise LSPClientError(f"Failed to initialize {session.language} server")

        session.capabilities = result.get("capabilities", {})

        # Send initialized notification
        await self._send_notification(session, "initialized", {})
        session.initialized = True
        logger.info(f"LSP server initialized for {session.language}")

    # ==========================================================================
    # Request/Response Communication
    # ==========================================================================

    async def _send_request(
        self,
        session: LanguageSession,
        method: str,
        params: dict[str, Any] | None = None,
    ) -> Any:
        """Send a request to the LSP server and wait for response.

        Args:
            session: The language session.
            method: The LSP method name.
            params: Optional parameters for the request.

        Returns:
            The result from the server.

        Raises:
            LSPTimeoutError: If the request times out.
            LSPClientError: If the server returns an error.
        """
        if not session.server.is_running():
            raise LSPClientError(f"Server for {session.language} is not running")

        process = session.server.process
        if process is None or process.stdin is None:
            raise LSPClientError(f"Server process not available for {session.language}")

        # Create request
        request_id = session.protocol.generate_id()
        request = JsonRpcRequest(id=request_id, method=method, params=params)

        # Create future for response
        future: asyncio.Future[Any] = asyncio.get_event_loop().create_future()
        session.protocol.add_pending_request(request_id, future)

        # Send request
        try:
            await session.protocol.write_message(process.stdin, request)
        except Exception as e:
            session.protocol._pending_requests.pop(request_id, None)
            raise LSPClientError(f"Failed to send request: {e}") from e

        # Wait for response
        try:
            result = await asyncio.wait_for(future, timeout=self.request_timeout)
            return result
        except asyncio.TimeoutError:
            session.protocol._pending_requests.pop(request_id, None)
            raise LSPTimeoutError(f"Request {method} timed out after {self.request_timeout}s")

    async def _send_notification(
        self,
        session: LanguageSession,
        method: str,
        params: dict[str, Any] | None = None,
    ) -> None:
        """Send a notification to the LSP server (no response expected).

        Args:
            session: The language session.
            method: The LSP method name.
            params: Optional parameters for the notification.
        """
        if not session.server.is_running():
            return

        process = session.server.process
        if process is None or process.stdin is None:
            return

        notification = JsonRpcNotification(method=method, params=params)
        try:
            await session.protocol.write_message(process.stdin, notification)
        except Exception as e:
            logger.warning(f"Failed to send notification {method}: {e}")

    async def _response_reader(self, session: LanguageSession) -> None:
        """Background task to read responses from the LSP server.

        Args:
            session: The language session.
        """
        process = session.server.process
        if process is None or process.stdout is None:
            return

        reader = process.stdout
        try:
            while session.server.is_running():
                try:
                    message = await session.protocol.read_message(reader)
                    await self._handle_message(session, message)
                except Exception as e:
                    if session.server.is_running():
                        logger.warning(f"Error reading LSP message: {e}")
                    break
        except asyncio.CancelledError:
            pass

    async def _handle_message(self, session: LanguageSession, message: dict[str, Any]) -> None:
        """Handle a message received from the LSP server.

        Args:
            session: The language session.
            message: The parsed JSON-RPC message.
        """
        # Check if it's a response (has id but no method)
        if "id" in message and "method" not in message:
            request_id = message["id"]
            if session.protocol.has_pending_request(request_id):
                if "error" in message:
                    error = message["error"]
                    logger.warning(f"LSP error: {error.get('message', 'Unknown error')}")
                    session.protocol.complete_request(request_id, None)
                else:
                    session.protocol.complete_request(request_id, message.get("result"))
            return

        # Check if it's a notification (has method but no id)
        method = message.get("method", "")
        params = message.get("params", {})

        if method == "textDocument/publishDiagnostics":
            # Cache diagnostics for the file
            uri = params.get("uri", "")
            diagnostics = params.get("diagnostics", [])
            session.diagnostics_cache[uri] = diagnostics
            logger.debug(f"Received {len(diagnostics)} diagnostics for {uri}")

        elif method == "window/logMessage":
            msg_type = params.get("type", 4)
            msg = params.get("message", "")
            if msg_type == 1:  # Error
                logger.error(f"LSP [{session.language}]: {msg}")
            elif msg_type == 2:  # Warning
                logger.warning(f"LSP [{session.language}]: {msg}")
            else:
                logger.debug(f"LSP [{session.language}]: {msg}")

    # ==========================================================================
    # Document Management
    # ==========================================================================

    async def _open_document(self, session: LanguageSession, file_path: str) -> None:
        """Open a document in the LSP server.

        Args:
            session: The language session.
            file_path: Path to the file.
        """
        uri = self._file_to_uri(file_path)
        if uri in session.open_documents:
            return

        # Read file contents
        try:
            with open(file_path, "r", encoding="utf-8", errors="replace") as f:
                text = f.read()
        except Exception as e:
            logger.warning(f"Failed to read file {file_path}: {e}")
            return

        # Determine language ID
        language_id = self._get_language_id(file_path)

        params = {
            "textDocument": {
                "uri": uri,
                "languageId": language_id,
                "version": 1,
                "text": text,
            }
        }

        await self._send_notification(session, "textDocument/didOpen", params)
        session.open_documents.add(uri)

    async def _close_document(self, session: LanguageSession, file_path: str) -> None:
        """Close a document in the LSP server.

        Args:
            session: The language session.
            file_path: Path to the file.
        """
        uri = self._file_to_uri(file_path)
        if uri not in session.open_documents:
            return

        params = {"textDocument": {"uri": uri}}
        await self._send_notification(session, "textDocument/didClose", params)
        session.open_documents.discard(uri)

    def _get_language_id(self, file_path: str) -> str:
        """Get the language ID for a file.

        Args:
            file_path: Path to the file.

        Returns:
            Language ID string.
        """
        ext = Path(file_path).suffix.lower()
        language_map = {
            ".py": "python",
            ".pyi": "python",
            ".ts": "typescript",
            ".tsx": "typescriptreact",
            ".js": "javascript",
            ".jsx": "javascriptreact",
            ".go": "go",
            ".rs": "rust",
            ".java": "java",
            ".kt": "kotlin",
            ".swift": "swift",
            ".cs": "csharp",
            ".c": "c",
            ".cpp": "cpp",
            ".h": "c",
            ".hpp": "cpp",
            ".rb": "ruby",
            ".php": "php",
            ".ex": "elixir",
            ".exs": "elixir",
            ".scala": "scala",
            ".r": "r",
            ".dart": "dart",
        }
        return language_map.get(ext, "plaintext")

    # ==========================================================================
    # Internal Request Methods
    # ==========================================================================

    async def _request_diagnostics(self, file_path: str) -> list[dict[str, Any]]:
        """Request diagnostics from the LSP server.

        Diagnostics are push-based in LSP (via textDocument/publishDiagnostics).
        This method opens the document and waits briefly for diagnostics.

        Args:
            file_path: Path to the file.

        Returns:
            Raw diagnostic data from server.
        """
        language = self.get_language_for_file(file_path)
        if language is None:
            return []

        try:
            session = await self._get_or_create_session(language)
        except Exception as e:
            logger.warning(f"Failed to get session for {language}: {e}")
            return []

        uri = self._file_to_uri(file_path)

        # Open the document to trigger diagnostics
        await self._open_document(session, file_path)

        # Wait for diagnostics to arrive (they come via notification)
        # Give the server time to process
        for _ in range(10):  # Wait up to 1 second
            if uri in session.diagnostics_cache:
                return session.diagnostics_cache[uri]
            await asyncio.sleep(0.1)

        return session.diagnostics_cache.get(uri, [])

    async def _request_references(self, file_path: str, position: Position) -> list[dict[str, Any]]:
        """Request references from the LSP server.

        Args:
            file_path: Path to the file.
            position: Position of the symbol.

        Returns:
            Raw reference data from server.
        """
        language = self.get_language_for_file(file_path)
        if language is None:
            return []

        try:
            session = await self._get_or_create_session(language)
        except Exception as e:
            logger.warning(f"Failed to get session for {language}: {e}")
            return []

        # Ensure document is open
        await self._open_document(session, file_path)

        params = {
            "textDocument": {"uri": self._file_to_uri(file_path)},
            "position": {"line": position.line, "character": position.character},
            "context": {"includeDeclaration": True},
        }

        try:
            result = await self._send_request(session, "textDocument/references", params)
            return result if result else []
        except LSPTimeoutError:
            logger.warning(f"References request timed out for {file_path}")
            return []
        except LSPClientError as e:
            logger.warning(f"References request failed: {e}")
            return []

    async def _request_rename(self, file_path: str, position: Position, new_name: str) -> dict[str, Any]:
        """Request rename from the LSP server.

        Args:
            file_path: Path to the file.
            position: Position of the symbol.
            new_name: New name for the symbol.

        Returns:
            Raw workspace edit data from server.
        """
        language = self.get_language_for_file(file_path)
        if language is None:
            return {"changes": {}}

        try:
            session = await self._get_or_create_session(language)
        except Exception as e:
            logger.warning(f"Failed to get session for {language}: {e}")
            return {"changes": {}}

        # Ensure document is open
        await self._open_document(session, file_path)

        params = {
            "textDocument": {"uri": self._file_to_uri(file_path)},
            "position": {"line": position.line, "character": position.character},
            "newName": new_name,
        }

        try:
            result = await self._send_request(session, "textDocument/rename", params)
            return result if result else {"changes": {}}
        except LSPTimeoutError:
            logger.warning(f"Rename request timed out for {file_path}")
            return {"changes": {}}
        except LSPClientError as e:
            logger.warning(f"Rename request failed: {e}")
            return {"changes": {}}

    async def _request_hover(self, file_path: str, position: Position) -> dict[str, Any] | None:
        """Request hover info from the LSP server.

        Args:
            file_path: Path to the file.
            position: Position to get hover info for.

        Returns:
            Raw hover data from server or None.
        """
        language = self.get_language_for_file(file_path)
        if language is None:
            return None

        try:
            session = await self._get_or_create_session(language)
        except Exception as e:
            logger.warning(f"Failed to get session for {language}: {e}")
            return None

        # Ensure document is open
        await self._open_document(session, file_path)

        params = {
            "textDocument": {"uri": self._file_to_uri(file_path)},
            "position": {"line": position.line, "character": position.character},
        }

        try:
            result = await self._send_request(session, "textDocument/hover", params)
            return result
        except LSPTimeoutError:
            logger.warning(f"Hover request timed out for {file_path}")
            return None
        except LSPClientError as e:
            logger.warning(f"Hover request failed: {e}")
            return None

    async def _request_definition(self, file_path: str, position: Position) -> list[dict[str, Any]]:
        """Request definition from the LSP server.

        Args:
            file_path: Path to the file.
            position: Position of the symbol.

        Returns:
            Raw definition data from server.
        """
        language = self.get_language_for_file(file_path)
        if language is None:
            return []

        try:
            session = await self._get_or_create_session(language)
        except Exception as e:
            logger.warning(f"Failed to get session for {language}: {e}")
            return []

        # Ensure document is open
        await self._open_document(session, file_path)

        params = {
            "textDocument": {"uri": self._file_to_uri(file_path)},
            "position": {"line": position.line, "character": position.character},
        }

        try:
            result = await self._send_request(session, "textDocument/definition", params)
            if result is None:
                return []
            # LSP can return Location, Location[], or LocationLink[]
            if isinstance(result, dict):
                return [result]
            return result if result else []
        except LSPTimeoutError:
            logger.warning(f"Definition request timed out for {file_path}")
            return []
        except LSPClientError as e:
            logger.warning(f"Definition request failed: {e}")
            return []

    async def _request_document_symbols(self, file_path: str) -> list[dict[str, Any]]:
        """Request document symbols from the LSP server.

        Args:
            file_path: Path to the file.

        Returns:
            Raw document symbol data from server.
        """
        language = self.get_language_for_file(file_path)
        if language is None:
            return []

        try:
            session = await self._get_or_create_session(language)
        except Exception as e:
            logger.warning(f"Failed to get session for {language}: {e}")
            return []

        # Ensure document is open
        await self._open_document(session, file_path)

        params = {"textDocument": {"uri": self._file_to_uri(file_path)}}

        try:
            result = await self._send_request(session, "textDocument/documentSymbol", params)
            return result if result else []
        except LSPTimeoutError:
            logger.warning(f"Document symbols request timed out for {file_path}")
            return []
        except LSPClientError as e:
            logger.warning(f"Document symbols request failed: {e}")
            return []

    # ==========================================================================
    # Helper Methods
    # ==========================================================================

    def _file_to_uri(self, file_path: str) -> str:
        """Convert a file path to a file:// URI.

        Args:
            file_path: Path to the file.

        Returns:
            File URI string.
        """
        if not file_path.startswith("/"):
            file_path = "/" + file_path
        return f"file://{file_path}"

    def _uri_to_file(self, uri: str) -> str:
        """Convert a file:// URI to a file path.

        Args:
            uri: File URI string.

        Returns:
            File path string.
        """
        if uri.startswith("file://"):
            return uri[7:]
        return uri

    def _parse_diagnostic(self, raw: dict[str, Any]) -> Diagnostic:
        """Parse a Diagnostic from raw LSP data.

        Args:
            raw: Raw diagnostic dictionary from LSP.

        Returns:
            Diagnostic instance.
        """
        range_data = raw["range"]
        severity_value = raw.get("severity", DiagnosticSeverity.ERROR)
        return Diagnostic(
            range=self._parse_range(range_data),
            severity=DiagnosticSeverity(severity_value),
            code=raw.get("code"),
            source=raw.get("source", ""),
            message=raw.get("message", ""),
        )

    def _parse_location(self, raw: dict[str, Any]) -> Location:
        """Parse a Location from raw LSP data.

        Args:
            raw: Raw location dictionary from LSP.

        Returns:
            Location instance.
        """
        # Handle LocationLink format
        if "targetUri" in raw:
            return Location(
                uri=raw["targetUri"],
                range=self._parse_range(raw.get("targetRange", raw.get("targetSelectionRange", {}))),
            )
        return Location(
            uri=raw["uri"],
            range=self._parse_range(raw["range"]),
        )

    def _parse_range(self, raw: dict[str, Any]) -> Range:
        """Parse a Range from raw LSP data.

        Args:
            raw: Raw range dictionary from LSP.

        Returns:
            Range instance.
        """
        return Range(
            start=Position(
                line=raw["start"]["line"],
                character=raw["start"]["character"],
            ),
            end=Position(
                line=raw["end"]["line"],
                character=raw["end"]["character"],
            ),
        )

    def _parse_workspace_edit(self, raw: dict[str, Any]) -> WorkspaceEdit:
        """Parse a WorkspaceEdit from raw LSP data.

        Args:
            raw: Raw workspace edit dictionary from LSP.

        Returns:
            WorkspaceEdit instance.
        """
        changes: dict[str, list[TextEdit]] = {}

        raw_changes = raw.get("changes", {})
        for uri, edits in raw_changes.items():
            changes[uri] = [
                TextEdit(
                    range=self._parse_range(edit["range"]),
                    new_text=edit["newText"],
                )
                for edit in edits
            ]

        return WorkspaceEdit(changes=changes)

    def _parse_hover_info(self, raw: dict[str, Any]) -> HoverInfo:
        """Parse HoverInfo from raw LSP data.

        Args:
            raw: Raw hover dictionary from LSP.

        Returns:
            HoverInfo instance.
        """
        contents = raw.get("contents", "")

        # Handle MarkupContent format
        if isinstance(contents, dict):
            contents = contents.get("value", "")
        # Handle array format
        elif isinstance(contents, list):
            contents = "\n".join(c.get("value", str(c)) if isinstance(c, dict) else str(c) for c in contents)

        range_data = raw.get("range")
        hover_range = self._parse_range(range_data) if range_data else None

        return HoverInfo(contents=contents, range=hover_range)

    def _parse_document_symbol(self, raw: dict[str, Any]) -> DocumentSymbol:
        """Parse a DocumentSymbol from raw LSP data.

        Args:
            raw: Raw document symbol dictionary from LSP.

        Returns:
            DocumentSymbol instance.
        """
        # Handle SymbolInformation format (flat)
        if "location" in raw:
            location = raw["location"]
            return DocumentSymbol(
                name=raw["name"],
                kind=SymbolKind(raw.get("kind", SymbolKind.VARIABLE)),
                range=self._parse_range(location["range"]),
                selection_range=self._parse_range(location["range"]),
                detail=raw.get("containerName"),
                children=[],
            )

        # Handle DocumentSymbol format (hierarchical)
        children = [self._parse_document_symbol(c) for c in raw.get("children", [])]
        return DocumentSymbol(
            name=raw["name"],
            kind=SymbolKind(raw.get("kind", SymbolKind.VARIABLE)),
            range=self._parse_range(raw["range"]),
            selection_range=self._parse_range(raw["selectionRange"]),
            detail=raw.get("detail"),
            children=children,
        )
