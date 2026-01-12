# LSP Protocol Tests - RED Phase
"""Tests for LSP JSON-RPC 2.0 protocol implementation."""

import asyncio
import json

import pytest

from moai_adk.lsp.protocol import (
    ContentLengthError,
    JsonRpcError,
    JsonRpcNotification,
    JsonRpcRequest,
    JsonRpcResponse,
    LSPProtocol,
    ProtocolError,
)


class TestJsonRpcMessage:
    """Tests for JSON-RPC message types."""

    def test_create_request(self):
        """Request should have jsonrpc, id, method, and params."""
        request = JsonRpcRequest(id=1, method="initialize", params={"rootUri": "file:///project"})
        assert request.id == 1
        assert request.method == "initialize"
        assert request.params == {"rootUri": "file:///project"}

    def test_request_to_json(self):
        """Request should serialize to JSON-RPC format."""
        request = JsonRpcRequest(id=42, method="textDocument/hover", params={"uri": "file:///test.py"})
        data = request.to_json()
        parsed = json.loads(data)
        assert parsed["jsonrpc"] == "2.0"
        assert parsed["id"] == 42
        assert parsed["method"] == "textDocument/hover"
        assert parsed["params"] == {"uri": "file:///test.py"}

    def test_create_notification(self):
        """Notification should have method and params but no id."""
        notif = JsonRpcNotification(method="textDocument/didOpen", params={"uri": "file:///test.py"})
        assert notif.method == "textDocument/didOpen"
        assert notif.params == {"uri": "file:///test.py"}

    def test_notification_to_json(self):
        """Notification should serialize without id."""
        notif = JsonRpcNotification(method="exit", params=None)
        data = notif.to_json()
        parsed = json.loads(data)
        assert parsed["jsonrpc"] == "2.0"
        assert parsed["method"] == "exit"
        assert "id" not in parsed

    def test_create_response_success(self):
        """Response should have id and result for success."""
        response = JsonRpcResponse(id=1, result={"capabilities": {}})
        assert response.id == 1
        assert response.result == {"capabilities": {}}
        assert response.error is None

    def test_create_response_error(self):
        """Response should have id and error for failure."""
        response = JsonRpcResponse(id=1, error=JsonRpcError(code=-32600, message="Invalid Request"))
        assert response.id == 1
        assert response.result is None
        assert response.error.code == -32600

    def test_parse_response_from_json(self):
        """Response should be parseable from JSON."""
        data = '{"jsonrpc": "2.0", "id": 1, "result": {"foo": "bar"}}'
        response = JsonRpcResponse.from_json(data)
        assert response.id == 1
        assert response.result == {"foo": "bar"}

    def test_parse_error_response_from_json(self):
        """Error response should be parseable from JSON."""
        data = '{"jsonrpc": "2.0", "id": 1, "error": {"code": -32600, "message": "Invalid"}}'
        response = JsonRpcResponse.from_json(data)
        assert response.id == 1
        assert response.error is not None
        assert response.error.code == -32600


class TestJsonRpcError:
    """Tests for JSON-RPC error codes."""

    def test_standard_error_codes(self):
        """JSON-RPC standard error codes should be defined."""
        assert JsonRpcError.PARSE_ERROR == -32700
        assert JsonRpcError.INVALID_REQUEST == -32600
        assert JsonRpcError.METHOD_NOT_FOUND == -32601
        assert JsonRpcError.INVALID_PARAMS == -32602
        assert JsonRpcError.INTERNAL_ERROR == -32603

    def test_error_to_dict(self):
        """Error should be convertible to dict."""
        error = JsonRpcError(code=-32600, message="Invalid Request", data={"detail": "missing id"})
        d = error.to_dict()
        assert d["code"] == -32600
        assert d["message"] == "Invalid Request"
        assert d["data"] == {"detail": "missing id"}


class TestLSPProtocol:
    """Tests for LSP protocol message encoding/decoding."""

    def test_encode_message(self):
        """Protocol should encode message with Content-Length header."""
        protocol = LSPProtocol()
        request = JsonRpcRequest(id=1, method="initialize", params={})
        encoded = protocol.encode_message(request)

        assert b"Content-Length:" in encoded
        assert b"\r\n\r\n" in encoded
        # Body should be after the header
        header, body = encoded.split(b"\r\n\r\n", 1)
        assert json.loads(body.decode("utf-8"))["method"] == "initialize"

    def test_encode_message_content_length(self):
        """Content-Length should match actual body size."""
        protocol = LSPProtocol()
        request = JsonRpcRequest(id=1, method="test", params={"key": "value"})
        encoded = protocol.encode_message(request)

        header, body = encoded.split(b"\r\n\r\n", 1)
        # Extract Content-Length value
        header_str = header.decode("utf-8")
        for line in header_str.split("\r\n"):
            if line.startswith("Content-Length:"):
                length = int(line.split(":")[1].strip())
                assert length == len(body)
                break

    def test_decode_message(self):
        """Protocol should decode message from bytes."""
        protocol = LSPProtocol()
        body = '{"jsonrpc": "2.0", "id": 1, "result": {"data": "test"}}'
        content = f"Content-Length: {len(body)}\r\n\r\n{body}".encode("utf-8")

        result = protocol.decode_message(content)
        assert result["id"] == 1
        assert result["result"]["data"] == "test"

    def test_decode_message_invalid_content_length(self):
        """Protocol should raise error for invalid Content-Length."""
        protocol = LSPProtocol()
        content = b"Content-Length: abc\r\n\r\n{}"

        with pytest.raises(ContentLengthError):
            protocol.decode_message(content)

    def test_decode_message_missing_header(self):
        """Protocol should raise error for missing Content-Length."""
        protocol = LSPProtocol()
        content = b'{"jsonrpc": "2.0"}'

        with pytest.raises(ProtocolError):
            protocol.decode_message(content)

    def test_generate_request_id(self):
        """Protocol should generate unique request IDs."""
        protocol = LSPProtocol()
        id1 = protocol.generate_id()
        id2 = protocol.generate_id()
        id3 = protocol.generate_id()

        assert id1 != id2 != id3
        assert isinstance(id1, int)


@pytest.mark.asyncio
class TestLSPProtocolAsync:
    """Async tests for LSP protocol stream handling."""

    async def test_read_message_from_stream(self):
        """Protocol should read complete message from stream."""
        protocol = LSPProtocol()

        # Create a mock stream with a complete message
        body = '{"jsonrpc": "2.0", "id": 1, "result": {}}'
        message = f"Content-Length: {len(body)}\r\n\r\n{body}".encode("utf-8")

        reader = asyncio.StreamReader()
        reader.feed_data(message)
        reader.feed_eof()

        result = await protocol.read_message(reader)
        assert result["id"] == 1

    async def test_read_message_partial_header(self):
        """Protocol should handle partial header reads."""
        protocol = LSPProtocol()

        body = '{"jsonrpc": "2.0", "id": 2, "result": {"key": "val"}}'
        message = f"Content-Length: {len(body)}\r\n\r\n{body}".encode("utf-8")

        reader = asyncio.StreamReader()
        # Feed data in chunks
        reader.feed_data(message[:10])
        reader.feed_data(message[10:])
        reader.feed_eof()

        result = await protocol.read_message(reader)
        assert result["id"] == 2

    async def test_write_message_to_stream(self):
        """Protocol should write message to stream."""
        protocol = LSPProtocol()

        # Use mock writer
        written_data = []

        class MockWriter:
            def write(self, data):
                written_data.append(data)

            async def drain(self):
                pass

        request = JsonRpcRequest(id=1, method="shutdown", params=None)
        await protocol.write_message(MockWriter(), request)

        assert len(written_data) == 1
        assert b"Content-Length:" in written_data[0]

    async def test_pending_requests(self):
        """Protocol should track pending requests."""
        protocol = LSPProtocol()

        # Add pending request
        future = asyncio.Future()
        protocol.add_pending_request(1, future)

        assert protocol.has_pending_request(1)
        assert not protocol.has_pending_request(2)

        # Complete request
        protocol.complete_request(1, {"result": "done"})
        assert future.done()
        assert future.result() == {"result": "done"}

    async def test_pending_request_error(self):
        """Protocol should handle request errors."""
        protocol = LSPProtocol()

        future = asyncio.Future()
        protocol.add_pending_request(1, future)

        error = JsonRpcError(code=-32600, message="Invalid")
        protocol.fail_request(1, error)

        assert future.done()
        with pytest.raises(Exception) as exc_info:
            future.result()
        assert "Invalid" in str(exc_info.value)
