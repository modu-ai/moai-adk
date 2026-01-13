"""Additional coverage tests for LSP protocol module.

Tests for lines not covered by existing tests.
"""

import asyncio

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


class TestJsonRpcRequestToDict:
    """Test JsonRpcRequest.to_dict method."""

    def test_to_dict_with_params(self):
        """Should convert request with params to dictionary."""
        request = JsonRpcRequest(id=1, method="test", params={"key": "value"})
        result = request.to_dict()

        assert result["jsonrpc"] == "2.0"
        assert result["id"] == 1
        assert result["method"] == "test"
        assert result["params"] == {"key": "value"}

    def test_to_dict_without_params(self):
        """Should convert request without params to dictionary."""
        request = JsonRpcRequest(id=2, method="test", params=None)
        result = request.to_dict()

        assert result["jsonrpc"] == "2.0"
        assert result["id"] == 2
        assert result["method"] == "test"
        assert "params" not in result


class TestJsonRpcNotification:
    """Test JsonRpcNotification class."""

    def test_to_json_with_params(self):
        """Should serialize notification with params."""
        notification = JsonRpcNotification(method="notify", params={"data": "test"})
        json_str = notification.to_json()

        # Parse and verify structure
        import json

        parsed = json.loads(json_str)
        assert parsed["method"] == "notify"
        assert parsed["params"] == {"data": "test"}

    def test_to_json_without_params(self):
        """Should serialize notification without params."""
        notification = JsonRpcNotification(method="notify", params=None)
        json_str = notification.to_json()

        # Parse and verify structure
        import json

        parsed = json.loads(json_str)
        assert parsed["method"] == "notify"
        assert "params" not in parsed

    def test_to_dict_with_params(self):
        """Should convert notification with params to dict."""
        notification = JsonRpcNotification(method="notify", params=[1, 2, 3])
        result = notification.to_dict()

        assert result["jsonrpc"] == "2.0"
        assert result["method"] == "notify"
        assert result["params"] == [1, 2, 3]

    def test_to_dict_without_params(self):
        """Should convert notification without params to dict."""
        notification = JsonRpcNotification(method="notify", params=None)
        result = notification.to_dict()

        assert result["jsonrpc"] == "2.0"
        assert result["method"] == "notify"
        assert "params" not in result


class TestJsonRpcResponse:
    """Test JsonRpcResponse class."""

    def test_to_json_with_error(self):
        """Should serialize response with error."""
        error = JsonRpcError(code=-32600, message="Invalid request")
        response = JsonRpcResponse(id=1, error=error)

        json_str = response.to_json()

        # Parse and verify structure
        import json

        parsed = json.loads(json_str)
        assert "error" in parsed
        assert parsed["error"]["code"] == -32600
        assert parsed["error"]["message"] == "Invalid request"

    def test_to_json_with_result(self):
        """Should serialize response with result."""
        response = JsonRpcResponse(id=1, result={"status": "success"})

        json_str = response.to_json()

        # Parse and verify structure
        import json

        parsed = json.loads(json_str)
        assert parsed["result"] == {"status": "success"}
        assert "error" not in parsed


class TestLSPProtocolDecodeErrors:
    """Test LSPProtocol error handling in decode_message."""

    def test_decode_message_missing_content_length(self):
        """Should raise ProtocolError when Content-Length is missing."""
        protocol = LSPProtocol()
        data = b"Content-Type: application/json\r\n\r\n{}"

        with pytest.raises(ProtocolError, match="Missing Content-Length"):
            protocol.decode_message(data)

    def test_decode_message_invalid_content_length_value(self):
        """Should raise ContentLengthError for invalid Content-Length."""
        protocol = LSPProtocol()
        data = b"Content-Length: invalid\r\n\r\n{}"

        with pytest.raises(ContentLengthError, match="Invalid Content-Length"):
            protocol.decode_message(data)


class TestLSPProtocolReadErrors:
    """Test LSPProtocol error handling in read_message."""

    @pytest.mark.asyncio
    async def test_read_message_connection_closed(self):
        """Should raise ProtocolError when connection closes during headers."""
        protocol = LSPProtocol()

        # Create a mock reader that returns empty bytes (connection closed)
        class MockReader:
            async def readline(self):
                return b""

        with pytest.raises(ProtocolError, match="Connection closed"):
            await protocol.read_message(MockReader())

    @pytest.mark.asyncio
    async def test_read_message_missing_content_length(self):
        """Should raise ProtocolError when Content-Length is missing."""
        protocol = LSPProtocol()

        class MockReader:
            def __init__(self):
                self.headers_read = False

            async def readline(self):
                if not self.headers_read:
                    self.headers_read = True
                    return b"Content-Type: application/json\r\n\r\n"
                return b""

            async def readexactly(self, n):
                return b"{}"

        with pytest.raises(ProtocolError, match="Missing Content-Length"):
            await protocol.read_message(MockReader())

    @pytest.mark.asyncio
    async def test_read_message_invalid_content_length(self):
        """Should raise ContentLengthError for invalid Content-Length."""
        protocol = LSPProtocol()

        class MockReader:
            def __init__(self):
                self.headers_read = False

            async def readline(self):
                if not self.headers_read:
                    self.headers_read = True
                    return b"Content-Length: abc\r\n\r\n"
                return b""

            async def readexactly(self, n):
                return b"{}"

        with pytest.raises(ContentLengthError, match="Invalid Content-Length"):
            await protocol.read_message(MockReader())


class TestLSPProtocolPendingRequests:
    """Test LSPProtocol pending request management."""

    @pytest.mark.asyncio
    async def test_add_and_check_pending_request(self):
        """Should add and check pending requests."""
        protocol = LSPProtocol()
        future = asyncio.Future()

        protocol.add_pending_request(1, future)

        assert protocol.has_pending_request(1)

    @pytest.mark.asyncio
    async def test_complete_request(self):
        """Should complete pending request with result."""
        protocol = LSPProtocol()
        future = asyncio.Future()

        protocol.add_pending_request(1, future)
        protocol.complete_request(1, "test_result")

        assert future.result() == "test_result"
        assert not protocol.has_pending_request(1)

    @pytest.mark.asyncio
    async def test_complete_already_done_request(self):
        """Should handle completing already done request gracefully."""
        protocol = LSPProtocol()
        future = asyncio.Future()
        future.set_result("already_done")

        protocol.add_pending_request(1, future)
        # Should not raise exception
        protocol.complete_request(1, "new_result")

        assert future.result() == "already_done"

    @pytest.mark.asyncio
    async def test_fail_request(self):
        """Should fail pending request with error."""
        protocol = LSPProtocol()
        future = asyncio.Future()

        error = JsonRpcError(code=-32600, message="Invalid request")
        protocol.add_pending_request(1, future)
        protocol.fail_request(1, error)

        with pytest.raises(Exception, match="Invalid request"):
            future.result()

        assert not protocol.has_pending_request(1)

    @pytest.mark.asyncio
    async def test_fail_already_done_request(self):
        """Should handle failing already done request gracefully."""
        protocol = LSPProtocol()
        future = asyncio.Future()
        future.set_result("already_done")

        error = JsonRpcError(code=-32600, message="Invalid request")
        protocol.add_pending_request(1, future)
        # Should not raise exception
        protocol.fail_request(1, error)

        # Future should still have the original result
        assert future.result() == "already_done"
