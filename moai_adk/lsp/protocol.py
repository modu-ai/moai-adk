# LSP Protocol Implementation
"""JSON-RPC 2.0 protocol implementation for LSP communication.

This module implements the JSON-RPC 2.0 protocol used by the Language Server
Protocol (LSP) for communication between client and server.
"""

from __future__ import annotations

import asyncio
import json
from dataclasses import dataclass
from typing import Any


class ProtocolError(Exception):
    """Base exception for protocol errors."""

    pass


class ContentLengthError(ProtocolError):
    """Error raised when Content-Length is invalid or missing."""

    pass


@dataclass
class JsonRpcError:
    """JSON-RPC 2.0 error object.

    Attributes:
        code: Error code (integer).
        message: Error message.
        data: Optional additional error data.
    """

    # Standard JSON-RPC error codes
    PARSE_ERROR = -32700
    INVALID_REQUEST = -32600
    METHOD_NOT_FOUND = -32601
    INVALID_PARAMS = -32602
    INTERNAL_ERROR = -32603

    code: int
    message: str
    data: Any | None = None

    def to_dict(self) -> dict[str, Any]:
        """Convert error to dictionary.

        Returns:
            Dictionary representation of the error.
        """
        result = {"code": self.code, "message": self.message}
        if self.data is not None:
            result["data"] = self.data
        return result

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> JsonRpcError:
        """Create error from dictionary.

        Args:
            data: Dictionary containing error fields.

        Returns:
            JsonRpcError instance.
        """
        return cls(
            code=data["code"],
            message=data["message"],
            data=data.get("data"),
        )


@dataclass
class JsonRpcRequest:
    """JSON-RPC 2.0 request message.

    A request message to describe a request between client and server.

    Attributes:
        id: The request id (used to match response).
        method: The method to be invoked.
        params: The method's params.
    """

    id: int | str
    method: str
    params: dict[str, Any] | list[Any] | None = None

    def to_json(self) -> str:
        """Serialize request to JSON string.

        Returns:
            JSON string representation.
        """
        data: dict[str, Any] = {
            "jsonrpc": "2.0",
            "id": self.id,
            "method": self.method,
        }
        if self.params is not None:
            data["params"] = self.params
        return json.dumps(data)

    def to_dict(self) -> dict[str, Any]:
        """Convert to dictionary.

        Returns:
            Dictionary representation.
        """
        data: dict[str, Any] = {
            "jsonrpc": "2.0",
            "id": self.id,
            "method": self.method,
        }
        if self.params is not None:
            data["params"] = self.params
        return data


@dataclass
class JsonRpcNotification:
    """JSON-RPC 2.0 notification message.

    A notification message has no id and expects no response.

    Attributes:
        method: The method to be invoked.
        params: The method's params.
    """

    method: str
    params: dict[str, Any] | list[Any] | None = None

    def to_json(self) -> str:
        """Serialize notification to JSON string.

        Returns:
            JSON string representation.
        """
        data: dict[str, Any] = {
            "jsonrpc": "2.0",
            "method": self.method,
        }
        if self.params is not None:
            data["params"] = self.params
        return json.dumps(data)

    def to_dict(self) -> dict[str, Any]:
        """Convert to dictionary.

        Returns:
            Dictionary representation.
        """
        data: dict[str, Any] = {
            "jsonrpc": "2.0",
            "method": self.method,
        }
        if self.params is not None:
            data["params"] = self.params
        return data


@dataclass
class JsonRpcResponse:
    """JSON-RPC 2.0 response message.

    A response message sent as a result of a request.

    Attributes:
        id: The request id.
        result: The result of a request (mutually exclusive with error).
        error: The error in case of failure (mutually exclusive with result).
    """

    id: int | str
    result: Any | None = None
    error: JsonRpcError | None = None

    def to_json(self) -> str:
        """Serialize response to JSON string.

        Returns:
            JSON string representation.
        """
        data: dict[str, Any] = {
            "jsonrpc": "2.0",
            "id": self.id,
        }
        if self.error is not None:
            data["error"] = self.error.to_dict()
        else:
            data["result"] = self.result
        return json.dumps(data)

    @classmethod
    def from_json(cls, data: str) -> JsonRpcResponse:
        """Parse response from JSON string.

        Args:
            data: JSON string to parse.

        Returns:
            JsonRpcResponse instance.
        """
        parsed = json.loads(data)
        return cls.from_dict(parsed)

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> JsonRpcResponse:
        """Create response from dictionary.

        Args:
            data: Dictionary containing response fields.

        Returns:
            JsonRpcResponse instance.
        """
        error = None
        if "error" in data:
            error = JsonRpcError.from_dict(data["error"])

        return cls(
            id=data["id"],
            result=data.get("result"),
            error=error,
        )


class LSPProtocol:
    """LSP JSON-RPC 2.0 protocol handler.

    Handles encoding and decoding of LSP messages using the JSON-RPC 2.0
    protocol with Content-Length headers.
    """

    def __init__(self) -> None:
        """Initialize the protocol handler."""
        self._request_id: int = 0
        self._pending_requests: dict[int | str, asyncio.Future[Any]] = {}

    def generate_id(self) -> int:
        """Generate a unique request ID.

        Returns:
            Unique integer ID.
        """
        self._request_id += 1
        return self._request_id

    def encode_message(self, message: JsonRpcRequest | JsonRpcNotification | JsonRpcResponse) -> bytes:
        """Encode a message with Content-Length header.

        Args:
            message: The message to encode.

        Returns:
            Encoded bytes with header and body.
        """
        body = message.to_json().encode("utf-8")
        header = f"Content-Length: {len(body)}\r\n\r\n".encode("utf-8")
        return header + body

    def decode_message(self, data: bytes) -> dict[str, Any]:
        """Decode a message from bytes.

        Args:
            data: The raw bytes to decode.

        Returns:
            Parsed JSON dictionary.

        Raises:
            ProtocolError: If the message format is invalid.
            ContentLengthError: If Content-Length is invalid.
        """
        # Split header and body
        try:
            header_end = data.index(b"\r\n\r\n")
        except ValueError:
            raise ProtocolError("Missing header separator")

        header_bytes = data[:header_end]
        body_bytes = data[header_end + 4 :]

        # Parse Content-Length
        content_length = None
        for line in header_bytes.decode("utf-8").split("\r\n"):
            if line.startswith("Content-Length:"):
                try:
                    content_length = int(line.split(":")[1].strip())
                except ValueError:
                    raise ContentLengthError(f"Invalid Content-Length value: {line}")
                break

        if content_length is None:
            raise ProtocolError("Missing Content-Length header")

        # Parse body
        body_str = body_bytes[:content_length].decode("utf-8")
        return json.loads(body_str)

    async def read_message(self, reader: asyncio.StreamReader) -> dict[str, Any]:
        """Read a complete message from a stream.

        Args:
            reader: The stream reader.

        Returns:
            Parsed JSON dictionary.

        Raises:
            ProtocolError: If the message format is invalid.
            ContentLengthError: If Content-Length is invalid.
        """
        # Read headers until we find \r\n\r\n
        headers = b""
        while True:
            line = await reader.readline()
            if not line:
                raise ProtocolError("Connection closed while reading headers")
            headers += line
            if headers.endswith(b"\r\n\r\n"):
                break

        # Parse Content-Length
        content_length: int | None = None
        headers_text: str = headers.decode("utf-8")
        for header_line in headers_text.split("\r\n"):
            if header_line.startswith("Content-Length:"):
                try:
                    content_length = int(header_line.split(":")[1].strip())
                except ValueError:
                    raise ContentLengthError(f"Invalid Content-Length: {header_line}")
                break

        if content_length is None:
            raise ProtocolError("Missing Content-Length header")

        # Read body
        body = await reader.readexactly(content_length)
        return json.loads(body.decode("utf-8"))

    async def write_message(
        self,
        writer: Any,  # asyncio.StreamWriter or mock
        message: JsonRpcRequest | JsonRpcNotification | JsonRpcResponse,
    ) -> None:
        """Write a message to a stream.

        Args:
            writer: The stream writer.
            message: The message to write.
        """
        encoded = self.encode_message(message)
        writer.write(encoded)
        await writer.drain()

    def add_pending_request(self, request_id: int | str, future: asyncio.Future[Any]) -> None:
        """Add a pending request to track.

        Args:
            request_id: The request ID.
            future: The future to complete when response arrives.
        """
        self._pending_requests[request_id] = future

    def has_pending_request(self, request_id: int | str) -> bool:
        """Check if a request is pending.

        Args:
            request_id: The request ID to check.

        Returns:
            True if the request is pending.
        """
        return request_id in self._pending_requests

    def complete_request(self, request_id: int | str, result: Any) -> None:
        """Complete a pending request with a result.

        Args:
            request_id: The request ID.
            result: The result to set.
        """
        if request_id in self._pending_requests:
            future = self._pending_requests.pop(request_id)
            if not future.done():
                future.set_result(result)

    def fail_request(self, request_id: int | str, error: JsonRpcError) -> None:
        """Fail a pending request with an error.

        Args:
            request_id: The request ID.
            error: The error to set.
        """
        if request_id in self._pending_requests:
            future = self._pending_requests.pop(request_id)
            if not future.done():
                future.set_exception(Exception(error.message))
