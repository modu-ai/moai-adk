"""
Comprehensive test suite for backend.py module.

Tests cover API design validation, microservice architecture patterns,
async/await patterns, authentication and authorization, error handling,
performance optimization, and metrics collection with 90%+ coverage goal.

Module: src/moai_adk/foundation/backend.py
Classes: 7 main classes, 2 data classes
Lines: 998 total
"""

import asyncio
from unittest.mock import patch

import pytest

from src.moai_adk.foundation.backend import (
    APIDesignValidator,
    AsyncPatternAdvisor,
    AuthenticationManager,
    BackendMetricsCollector,
    ErrorHandlingStrategy,
    ErrorLog,
    HTTPMethod,
    MicroserviceArchitect,
    PerformanceOptimizer,
    RequestMetric,
)

# ============================================================================
# HTTPMethod Enum Tests
# ============================================================================


class TestHTTPMethod:
    """Test HTTPMethod enum."""

    def test_http_method_get(self):
        """Test GET HTTP method."""
        assert HTTPMethod.GET.value == "GET"

    def test_http_method_post(self):
        """Test POST HTTP method."""
        assert HTTPMethod.POST.value == "POST"

    def test_http_method_put(self):
        """Test PUT HTTP method."""
        assert HTTPMethod.PUT.value == "PUT"

    def test_http_method_patch(self):
        """Test PATCH HTTP method."""
        assert HTTPMethod.PATCH.value == "PATCH"

    def test_http_method_delete(self):
        """Test DELETE HTTP method."""
        assert HTTPMethod.DELETE.value == "DELETE"

    def test_http_method_head(self):
        """Test HEAD HTTP method."""
        assert HTTPMethod.HEAD.value == "HEAD"

    def test_http_method_options(self):
        """Test OPTIONS HTTP method."""
        assert HTTPMethod.OPTIONS.value == "OPTIONS"

    def test_all_http_methods_enum(self):
        """Test all HTTP methods are present in enum."""
        methods = {method.value for method in HTTPMethod}
        expected = {"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
        assert methods == expected


# ============================================================================
# APIDesignValidator Tests
# ============================================================================


class TestAPIDesignValidator:
    """Test APIDesignValidator class."""

    @pytest.fixture
    def validator(self):
        """Create validator instance."""
        return APIDesignValidator()

    def test_validator_initialization(self, validator):
        """Test validator initializes with empty endpoints."""
        assert validator.validated_endpoints == {}
        assert len(validator.VALID_HTTP_METHODS) == 7
        assert len(validator.VALID_STATUS_CODES) == 23

    # ========================================================================
    # REST Endpoint Validation Tests
    # ========================================================================

    def test_validate_rest_endpoint_valid_get(self, validator):
        """Test valid GET endpoint."""
        result = validator.validate_rest_endpoint({"method": "GET", "path": "/api/v1/users", "status_code": 200})
        assert result["valid"] is True
        assert result["method"] == "GET"
        assert result["status_code"] == 200
        assert result["errors"] is None

    def test_validate_rest_endpoint_valid_post(self, validator):
        """Test valid POST endpoint returns 201."""
        result = validator.validate_rest_endpoint({"method": "POST", "path": "/api/v1/users", "status_code": 201})
        assert result["valid"] is True
        assert result["method"] == "POST"
        assert result["status_code"] == 201
        assert result["errors"] is None

    def test_validate_rest_endpoint_valid_delete(self, validator):
        """Test valid DELETE endpoint returns 204."""
        result = validator.validate_rest_endpoint({"method": "DELETE", "path": "/api/v1/users/123", "status_code": 204})
        assert result["valid"] is True
        assert result["method"] == "DELETE"
        assert result["status_code"] == 204

    def test_validate_rest_endpoint_valid_put(self, validator):
        """Test valid PUT endpoint returns 200."""
        result = validator.validate_rest_endpoint({"method": "PUT", "path": "/api/v1/users/123", "status_code": 200})
        assert result["valid"] is True
        assert result["method"] == "PUT"

    def test_validate_rest_endpoint_valid_patch(self, validator):
        """Test valid PATCH endpoint returns 200."""
        result = validator.validate_rest_endpoint({"method": "PATCH", "path": "/api/v1/users/123", "status_code": 200})
        assert result["valid"] is True
        assert result["method"] == "PATCH"

    def test_validate_rest_endpoint_invalid_http_method(self, validator):
        """Test invalid HTTP method."""
        result = validator.validate_rest_endpoint({"method": "INVALID", "path": "/api/v1/users", "status_code": 200})
        assert result["valid"] is False
        assert "Invalid HTTP method" in result["errors"][0]

    def test_validate_rest_endpoint_invalid_path_no_slash(self, validator):
        """Test path without leading slash."""
        result = validator.validate_rest_endpoint({"method": "GET", "path": "api/v1/users", "status_code": 200})
        assert result["valid"] is False
        assert "must start with '/'" in result["errors"][0]

    def test_validate_rest_endpoint_invalid_path_empty(self, validator):
        """Test empty path."""
        result = validator.validate_rest_endpoint({"method": "GET", "path": "", "status_code": 200})
        assert result["valid"] is False
        assert "must start with '/'" in result["errors"][0]

    def test_validate_rest_endpoint_invalid_status_code(self, validator):
        """Test invalid status code."""
        result = validator.validate_rest_endpoint({"method": "GET", "path": "/api/v1/users", "status_code": 999})
        assert result["valid"] is False
        assert "Invalid status code" in result["errors"][0]

    def test_validate_rest_endpoint_post_returns_200_not_201(self, validator):
        """Test POST endpoint with wrong status code."""
        result = validator.validate_rest_endpoint({"method": "POST", "path": "/api/v1/users", "status_code": 200})
        assert result["valid"] is False
        assert "not allowed for POST" in result["errors"][0]

    def test_validate_rest_endpoint_delete_returns_200_not_204(self, validator):
        """Test DELETE endpoint with wrong status code."""
        result = validator.validate_rest_endpoint({"method": "DELETE", "path": "/api/v1/users/123", "status_code": 200})
        assert result["valid"] is False
        assert "not allowed for DELETE" in result["errors"][0]

    def test_validate_rest_endpoint_get_returns_206_partial_content(self, validator):
        """Test GET endpoint with 206 Partial Content."""
        result = validator.validate_rest_endpoint(
            {"method": "GET", "path": "/api/v1/files/download", "status_code": 206}
        )
        assert result["valid"] is True

    def test_validate_rest_endpoint_lowercase_method(self, validator):
        """Test lowercase HTTP method is normalized."""
        result = validator.validate_rest_endpoint({"method": "get", "path": "/api/v1/users", "status_code": 200})
        assert result["valid"] is True
        assert result["method"] == "GET"

    def test_validate_rest_endpoint_multiple_errors(self, validator):
        """Test endpoint with multiple validation errors."""
        result = validator.validate_rest_endpoint({"method": "INVALID", "path": "no-slash", "status_code": 999})
        assert result["valid"] is False
        assert len(result["errors"]) >= 3

    # ========================================================================
    # Versioning Strategy Tests
    # ========================================================================

    def test_get_versioning_strategy_url(self, validator):
        """Test URL versioning strategy."""
        strategy = validator.get_versioning_strategy("url")
        assert strategy["strategy"] == "url"
        assert strategy["pattern"] == "/api/v{version}/"
        assert "example" in strategy
        assert "pros" in strategy
        assert "cons" in strategy

    def test_get_versioning_strategy_header(self, validator):
        """Test header versioning strategy."""
        strategy = validator.get_versioning_strategy("header")
        assert strategy["strategy"] == "header"
        assert strategy["header_name"] == "API-Version"
        assert "example" in strategy

    def test_get_versioning_strategy_content_type(self, validator):
        """Test content-type versioning strategy."""
        strategy = validator.get_versioning_strategy("content-type")
        assert strategy["strategy"] == "content-type"
        assert "content_type" in strategy
        assert "vnd.api+json" in strategy["content_type"]

    def test_get_versioning_strategy_unknown_defaults_to_url(self, validator):
        """Test unknown strategy defaults to URL."""
        strategy = validator.get_versioning_strategy("unknown")
        assert strategy["strategy"] == "url"

    # ========================================================================
    # Error Response Standardization Tests
    # ========================================================================

    def test_standardize_error_response_validation_error(self, validator):
        """Test standardize error response for validation error."""
        error = {
            "type": "ValidationError",
            "message": "Invalid input",
            "status_code": 400,
            "details": {"field": "email"},
            "path": "/api/v1/users",
        }
        response = validator.standardize_error_response(error)

        assert response["type"] == "ValidationError"
        assert response["message"] == "Invalid input"
        assert response["status_code"] == 400
        assert response["details"]["field"] == "email"
        assert response["path"] == "/api/v1/users"
        assert "timestamp" in response
        assert "trace_id" in response

    def test_standardize_error_response_missing_fields(self, validator):
        """Test error response with missing optional fields."""
        error = {}
        response = validator.standardize_error_response(error)

        assert response["type"] == "Error"
        assert response["message"] == "An error occurred"
        assert response["status_code"] == 500
        assert response["details"] == {}
        assert response["path"] == ""

    def test_standardize_error_response_with_trace_id(self, validator):
        """Test trace_id is generated for error response."""
        error = {"type": "ServerError", "message": "Internal server error"}
        response = validator.standardize_error_response(error)

        assert len(response["trace_id"]) == 36  # UUID4 length
        assert "-" in response["trace_id"]

    def test_standardize_error_response_timestamp_format(self, validator):
        """Test timestamp is ISO 8601 format."""
        error = {"type": "Error"}
        response = validator.standardize_error_response(error)

        assert response["timestamp"].endswith("Z")
        assert "T" in response["timestamp"]


# ============================================================================
# MicroserviceArchitect Tests
# ============================================================================


class TestMicroserviceArchitect:
    """Test MicroserviceArchitect class."""

    @pytest.fixture
    def architect(self):
        """Create architect instance."""
        return MicroserviceArchitect()

    def test_architect_initialization(self, architect):
        """Test architect initializes with empty services."""
        assert architect.services == {}
        assert architect.communication_matrix == {}
        assert len(architect.COMMUNICATION_PATTERNS) == 3
        assert len(architect.SERVICE_DISCOVERY_BACKENDS) == 3

    # ========================================================================
    # Service Boundary Validation Tests
    # ========================================================================

    def test_validate_service_boundary_valid(self, architect):
        """Test valid service boundary."""
        service = {"name": "user-service", "domain": "auth", "endpoints": ["/api/v1/users"]}
        result = architect.validate_service_boundary(service)

        assert result["valid"] is True
        assert result["name"] == "user-service"
        assert result["domain"] == "auth"
        assert result["errors"] is None

    def test_validate_service_boundary_invalid_name_no_hyphen(self, architect):
        """Test service name without hyphen."""
        service = {"name": "userservice", "domain": "auth", "endpoints": ["/api/v1/users"]}
        result = architect.validate_service_boundary(service)

        assert result["valid"] is False
        assert "domain-service" in result["errors"][0]

    def test_validate_service_boundary_invalid_name_empty(self, architect):
        """Test empty service name."""
        service = {"name": "", "domain": "auth", "endpoints": ["/api/v1/users"]}
        result = architect.validate_service_boundary(service)

        assert result["valid"] is False

    def test_validate_service_boundary_missing_domain(self, architect):
        """Test service without domain."""
        service = {"name": "user-service", "endpoints": ["/api/v1/users"]}
        result = architect.validate_service_boundary(service)

        assert result["valid"] is False
        assert "domain" in result["errors"][0].lower()

    def test_validate_service_boundary_missing_endpoints(self, architect):
        """Test service without endpoints."""
        service = {"name": "user-service", "domain": "auth", "endpoints": []}
        result = architect.validate_service_boundary(service)

        assert result["valid"] is False
        assert "endpoint" in result["errors"][0].lower()

    def test_validate_service_boundary_stores_service(self, architect):
        """Test validated service is stored."""
        service = {"name": "user-service", "domain": "auth", "endpoints": ["/api/v1/users"]}
        architect.validate_service_boundary(service)

        assert "user-service" in architect.services
        assert architect.services["user-service"] == service

    # ========================================================================
    # Communication Pattern Tests
    # ========================================================================

    def test_get_communication_pattern_rest(self, architect):
        """Test REST communication pattern."""
        pattern = architect.get_communication_pattern(
            "rest", {"source": "api-gateway", "target": "user-service", "operation": "query"}
        )

        assert pattern["pattern"] == "rest"
        assert pattern["protocol"] == "HTTP/REST"
        assert pattern["async"] is False
        assert pattern["source"] == "api-gateway"
        assert pattern["target"] == "user-service"

    def test_get_communication_pattern_async(self, architect):
        """Test async communication pattern."""
        pattern = architect.get_communication_pattern("async", {})

        assert pattern["pattern"] == "async"
        assert pattern["async"] is True
        assert pattern["protocol"] in ["RabbitMQ", "Kafka", "AWS SQS"]

    def test_get_communication_pattern_grpc(self, architect):
        """Test gRPC communication pattern."""
        pattern = architect.get_communication_pattern("grpc", {})

        assert pattern["pattern"] == "grpc"
        assert pattern["async"] is True

    def test_get_communication_pattern_unknown_defaults_to_rest(self, architect):
        """Test unknown pattern defaults to REST."""
        pattern = architect.get_communication_pattern("unknown", {})

        assert pattern["pattern"] == "rest"

    # ========================================================================
    # Service Discovery Configuration Tests
    # ========================================================================

    def test_configure_service_discovery_consul(self, architect):
        """Test Consul service discovery configuration."""
        config = architect.configure_service_discovery(
            "consul", {"consul_host": "consul.local", "consul_port": 8500, "health_check_interval": 5}
        )

        assert config["registry"] == "consul"
        assert config["host"] == "consul.local"
        assert config["port"] == 8500
        assert config["health_check_enabled"] is True
        assert config["auto_deregister"] is True

    def test_configure_service_discovery_eureka(self, architect):
        """Test Eureka service discovery configuration."""
        config = architect.configure_service_discovery("eureka", {"consul_host": "eureka.local", "consul_port": 8761})

        assert config["registry"] == "eureka"
        assert config["health_check_enabled"] is True
        assert config["auto_deregister"] is False

    def test_configure_service_discovery_etcd(self, architect):
        """Test etcd service discovery configuration."""
        config = architect.configure_service_discovery("etcd", {})

        assert config["registry"] == "etcd"
        assert config["auto_deregister"] is True

    def test_configure_service_discovery_default_host_port(self, architect):
        """Test default host and port values."""
        config = architect.configure_service_discovery("consul", {})

        assert config["host"] == "localhost"
        assert config["port"] == 8500

    def test_configure_service_discovery_deregister_after(self, architect):
        """Test deregister_critical_service_after configuration."""
        config = architect.configure_service_discovery("consul", {"deregister_critical_service_after": "60s"})

        assert config["deregister_after"] == "60s"


# ============================================================================
# AsyncPatternAdvisor Tests
# ============================================================================


class TestAsyncPatternAdvisor:
    """Test AsyncPatternAdvisor class."""

    @pytest.fixture
    def advisor(self):
        """Create advisor instance."""
        return AsyncPatternAdvisor()

    def test_advisor_initialization(self, advisor):
        """Test advisor initializes with empty operations."""
        assert advisor.async_operations == []

    # ========================================================================
    # Concurrent Execution Tests
    # ========================================================================

    @pytest.mark.asyncio
    async def test_execute_concurrent_single_operation(self, advisor):
        """Test concurrent execution of single operation."""

        async def dummy_op():
            return "result"

        results = await advisor.execute_concurrent([dummy_op])
        assert results == ["result"]

    @pytest.mark.asyncio
    async def test_execute_concurrent_multiple_operations(self, advisor):
        """Test concurrent execution of multiple operations."""

        async def op1():
            return "result1"

        async def op2():
            return "result2"

        async def op3():
            return "result3"

        results = await advisor.execute_concurrent([op1, op2, op3])
        assert len(results) == 3
        assert "result1" in results
        assert "result2" in results
        assert "result3" in results

    @pytest.mark.asyncio
    async def test_execute_concurrent_with_timeout_success(self, advisor):
        """Test concurrent execution completes within timeout."""

        async def quick_op():
            await asyncio.sleep(0.01)
            return "done"

        results = await advisor.execute_concurrent([quick_op], timeout=1.0)
        assert results == ["done"]

    @pytest.mark.asyncio
    async def test_execute_concurrent_with_timeout_exceeds(self, advisor):
        """Test concurrent execution raises timeout error."""

        async def slow_op():
            await asyncio.sleep(1.0)
            return "done"

        with pytest.raises(asyncio.TimeoutError):
            await advisor.execute_concurrent([slow_op], timeout=0.1)

    @pytest.mark.asyncio
    async def test_execute_concurrent_timeout_cancels_tasks(self, advisor):
        """Test concurrent execution cancels remaining tasks on timeout."""

        async def slow_op():
            try:
                await asyncio.sleep(2.0)
            except asyncio.CancelledError:
                # Task was properly cancelled
                raise

        with pytest.raises(asyncio.TimeoutError):
            await advisor.execute_concurrent([slow_op, slow_op], timeout=0.05)

    @pytest.mark.asyncio
    async def test_execute_concurrent_without_timeout(self, advisor):
        """Test concurrent execution without timeout."""

        async def op():
            await asyncio.sleep(0.01)
            return "result"

        results = await advisor.execute_concurrent([op])
        assert results == ["result"]

    # ========================================================================
    # Timeout Execution Tests
    # ========================================================================

    @pytest.mark.asyncio
    async def test_with_timeout_success(self, advisor):
        """Test operation completes within timeout."""

        async def quick_coro():
            await asyncio.sleep(0.01)
            return "success"

        result = await advisor.with_timeout(quick_coro(), 1.0)
        assert result == "success"

    @pytest.mark.asyncio
    async def test_with_timeout_exceeds(self, advisor):
        """Test operation exceeds timeout."""

        async def slow_coro():
            await asyncio.sleep(1.0)
            return "done"

        with pytest.raises(asyncio.TimeoutError):
            await advisor.with_timeout(slow_coro(), 0.1)

    # ========================================================================
    # Async Retry Decorator Tests
    # ========================================================================

    @pytest.mark.asyncio
    async def test_async_retry_success_first_attempt(self, advisor):
        """Test async retry succeeds on first attempt."""
        call_count = 0

        @advisor.async_retry(max_attempts=3, backoff_factor=1.0)
        async def successful_op():
            nonlocal call_count
            call_count += 1
            return "success"

        result = await successful_op()
        assert result == "success"
        assert call_count == 1

    @pytest.mark.asyncio
    async def test_async_retry_succeeds_after_retries(self, advisor):
        """Test async retry succeeds after failed attempts."""
        call_count = 0

        @advisor.async_retry(max_attempts=3, backoff_factor=1.0)
        async def failing_then_success():
            nonlocal call_count
            call_count += 1
            if call_count < 3:
                raise ValueError("Temporary failure")
            return "success"

        result = await failing_then_success()
        assert result == "success"
        assert call_count == 3

    @pytest.mark.asyncio
    async def test_async_retry_exhausts_attempts(self, advisor):
        """Test async retry raises after exhausting attempts."""
        call_count = 0

        @advisor.async_retry(max_attempts=3, backoff_factor=1.0)
        async def always_failing():
            nonlocal call_count
            call_count += 1
            raise ValueError("Always fails")

        with pytest.raises(ValueError):
            await always_failing()

        assert call_count == 3

    @pytest.mark.asyncio
    async def test_async_retry_with_custom_backoff(self, advisor):
        """Test async retry with custom backoff factor."""

        @advisor.async_retry(max_attempts=2, backoff_factor=2.0)
        async def should_work_second_time():
            await asyncio.sleep(0.01)
            return "done"

        result = await should_work_second_time()
        assert result == "done"


# ============================================================================
# AuthenticationManager Tests
# ============================================================================


class TestAuthenticationManager:
    """Test AuthenticationManager class."""

    @pytest.fixture
    def auth(self):
        """Create auth manager instance."""
        return AuthenticationManager(secret_key="test-secret-key")

    def test_auth_initialization(self, auth):
        """Test authentication manager initialization."""
        assert auth.secret_key == "test-secret-key"
        assert auth.algorithms == ["HS256"]
        assert auth.oauth_codes == {}

    # ========================================================================
    # JWT Token Generation Tests
    # ========================================================================

    def test_generate_jwt_token_basic(self, auth):
        """Test JWT token generation."""
        token = auth.generate_jwt_token({"sub": "user@example.com"})

        assert isinstance(token, str)
        assert token.count(".") == 2

    def test_generate_jwt_token_contains_parts(self, auth):
        """Test JWT token has three parts."""
        token = auth.generate_jwt_token({"sub": "user@example.com"})
        header, payload, signature = token.split(".")

        assert len(header) > 0
        assert len(payload) > 0
        assert len(signature) > 0

    def test_generate_jwt_token_custom_expires_in(self, auth):
        """Test JWT token with custom expiration."""
        token = auth.generate_jwt_token({"sub": "user@example.com"}, expires_in_hours=24)

        assert isinstance(token, str)
        assert token.count(".") == 2

    def test_generate_jwt_token_with_multiple_claims(self, auth):
        """Test JWT token with multiple claims."""
        claims = {"sub": "user@example.com", "name": "John Doe", "email": "john@example.com"}
        token = auth.generate_jwt_token(claims)

        assert isinstance(token, str)

    # ========================================================================
    # JWT Token Validation Tests
    # ========================================================================

    def test_validate_jwt_token_valid(self, auth):
        """Test validate valid JWT token."""
        token = auth.generate_jwt_token({"sub": "user@example.com"}, expires_in_hours=1)
        payload = auth.validate_jwt_token(token)

        assert payload["sub"] == "user@example.com"
        assert "exp" in payload
        assert "iat" in payload

    def test_validate_jwt_token_invalid_format(self, auth):
        """Test validate token with invalid format."""
        with pytest.raises(ValueError) as exc:
            auth.validate_jwt_token("invalid.token")

        assert "Invalid token" in str(exc.value)

    def test_validate_jwt_token_expired(self, auth):
        """Test validate expired token."""
        token = auth.generate_jwt_token({"sub": "user@example.com"}, expires_in_hours=-1)

        with pytest.raises(ValueError) as exc:
            auth.validate_jwt_token(token)

        assert "expired" in str(exc.value).lower()

    def test_validate_jwt_token_malformed(self, auth):
        """Test validate malformed token."""
        with pytest.raises(ValueError):
            auth.validate_jwt_token("not.a.token.with.dots")

    # ========================================================================
    # OAuth2 Authorization Code Tests
    # ========================================================================

    def test_generate_oauth_auth_code(self, auth):
        """Test generate OAuth2 authorization code."""
        params = {"client_id": "client123", "redirect_uri": "https://example.com/callback", "state": "state123"}
        result = auth.generate_oauth_auth_code(params)

        assert "code" in result
        assert result["expires_in"] == 600
        assert result["state"] == "state123"

    def test_generate_oauth_auth_code_stores_code(self, auth):
        """Test generated OAuth2 code is stored."""
        params = {"client_id": "client123"}
        result = auth.generate_oauth_auth_code(params)

        code = result["code"]
        assert code in auth.oauth_codes
        assert auth.oauth_codes[code]["params"]["client_id"] == "client123"

    def test_generate_oauth_auth_code_multiple_codes(self, auth):
        """Test multiple OAuth2 codes can be generated."""
        params1 = {"client_id": "client1"}
        params2 = {"client_id": "client2"}

        result1 = auth.generate_oauth_auth_code(params1)
        result2 = auth.generate_oauth_auth_code(params2)

        assert result1["code"] != result2["code"]
        assert len(auth.oauth_codes) == 2

    # ========================================================================
    # Permission Checking Tests
    # ========================================================================

    def test_has_permission_granted(self, auth):
        """Test user has required permission."""
        user = {"id": "user1", "permissions": ["read:posts", "write:posts"]}

        assert auth.has_permission(user, "read:posts") is True
        assert auth.has_permission(user, "write:posts") is True

    def test_has_permission_denied(self, auth):
        """Test user lacks required permission."""
        user = {"id": "user1", "permissions": ["read:posts"]}

        assert auth.has_permission(user, "write:posts") is False
        assert auth.has_permission(user, "delete:posts") is False

    def test_has_permission_no_permissions(self, auth):
        """Test user with no permissions."""
        user = {"id": "user1"}

        assert auth.has_permission(user, "read:posts") is False

    def test_has_permission_empty_permissions(self, auth):
        """Test user with empty permissions list."""
        user = {"id": "user1", "permissions": []}

        assert auth.has_permission(user, "read:posts") is False


# ============================================================================
# ErrorLog Dataclass Tests
# ============================================================================


class TestErrorLog:
    """Test ErrorLog dataclass."""

    def test_error_log_initialization(self):
        """Test ErrorLog creation."""
        log = ErrorLog(
            level="ERROR",
            message="Something went wrong",
            timestamp="2025-01-01T00:00:00Z",
            trace_id="trace-123",
            context={"user_id": "user1"},
        )

        assert log.level == "ERROR"
        assert log.message == "Something went wrong"
        assert log.trace_id == "trace-123"
        assert log.context["user_id"] == "user1"

    def test_error_log_different_levels(self):
        """Test ErrorLog with different log levels."""
        for level in ["INFO", "WARNING", "ERROR", "CRITICAL"]:
            log = ErrorLog(
                level=level, message="Message", timestamp="2025-01-01T00:00:00Z", trace_id="trace-123", context={}
            )
            assert log.level == level


# ============================================================================
# ErrorHandlingStrategy Tests
# ============================================================================


class TestErrorHandlingStrategy:
    """Test ErrorHandlingStrategy class."""

    @pytest.fixture
    def handler(self):
        """Create error handler instance."""
        return ErrorHandlingStrategy()

    def test_handler_initialization(self, handler):
        """Test error handler initializes."""
        assert handler.error_handlers == {}
        assert handler.logs == []

    # ========================================================================
    # Error Handling Tests
    # ========================================================================

    def test_handle_error_validation_error(self, handler):
        """Test handle validation error."""
        error = {
            "type": "ValidationError",
            "message": "Invalid email format",
            "status_code": 400,
            "details": {"field": "email"},
            "path": "/api/v1/users",
        }

        result = handler.handle_error(error)

        assert result["type"] == "ValidationError"
        assert result["message"] == "Invalid email format"
        assert result["status_code"] == 400
        assert result["details"]["field"] == "email"
        assert "trace_id" in result
        assert "timestamp" in result

    def test_handle_error_server_error(self, handler):
        """Test handle server error."""
        error = {"type": "InternalServerError", "message": "Database connection failed", "status_code": 500}

        result = handler.handle_error(error)

        assert result["type"] == "InternalServerError"
        assert result["status_code"] == 500

    def test_handle_error_missing_fields(self, handler):
        """Test handle error with missing optional fields."""
        error = {}

        result = handler.handle_error(error)

        assert result["type"] == "Error"
        assert result["message"] == "An error occurred"
        assert result["status_code"] == 500
        assert result["details"] == {}

    def test_handle_error_trace_id_unique(self, handler):
        """Test each error gets unique trace_id."""
        error = {"type": "Error"}

        result1 = handler.handle_error(error)
        result2 = handler.handle_error(error)

        assert result1["trace_id"] != result2["trace_id"]

    # ========================================================================
    # Logging with Context Tests
    # ========================================================================

    def test_log_with_context_info(self, handler):
        """Test log with INFO level."""
        result = handler.log_with_context("INFO", "User logged in", {"user_id": "user1"})

        assert result["level"] == "INFO"
        assert result["message"] == "User logged in"
        assert result["context"]["user_id"] == "user1"
        assert "trace_id" in result

    def test_log_with_context_error(self, handler):
        """Test log with ERROR level."""
        result = handler.log_with_context("ERROR", "Payment failed", {"order_id": "order123"})

        assert result["level"] == "ERROR"
        assert result["message"] == "Payment failed"
        assert len(handler.logs) == 1

    def test_log_with_context_no_context(self, handler):
        """Test log without context."""
        result = handler.log_with_context("WARNING", "Low memory")

        assert result["context"] == {}
        assert "trace_id" in result

    def test_log_with_context_stores_log(self, handler):
        """Test logs are stored."""
        handler.log_with_context("INFO", "Message 1", {})
        handler.log_with_context("ERROR", "Message 2", {})

        assert len(handler.logs) == 2
        assert handler.logs[0].message == "Message 1"
        assert handler.logs[1].message == "Message 2"

    def test_log_with_context_different_levels(self, handler):
        """Test logging with different levels."""
        for level in ["INFO", "WARNING", "ERROR"]:
            handler.log_with_context(level, f"{level} message")

        assert len(handler.logs) == 3
        assert handler.logs[0].level == "INFO"
        assert handler.logs[1].level == "WARNING"
        assert handler.logs[2].level == "ERROR"

    @patch("src.moai_adk.foundation.backend.logger")
    def test_log_with_context_calls_logger(self, mock_logger, handler):
        """Test log_with_context calls logger."""
        handler.log_with_context("INFO", "Test message")

        mock_logger.info.assert_called_once()
        assert "Test message" in str(mock_logger.info.call_args)


# ============================================================================
# PerformanceOptimizer Tests
# ============================================================================


class TestPerformanceOptimizer:
    """Test PerformanceOptimizer class."""

    @pytest.fixture
    def optimizer(self):
        """Create optimizer instance."""
        return PerformanceOptimizer()

    def test_optimizer_initialization(self, optimizer):
        """Test optimizer initializes."""
        assert optimizer.cache_configs == {}
        assert optimizer.rate_limits == {}

    # ========================================================================
    # Cache Configuration Tests
    # ========================================================================

    def test_configure_cache_redis(self, optimizer):
        """Test Redis cache configuration."""
        config = optimizer.configure_cache(backend="redis", ttl=3600)

        assert config["backend"] == "redis"
        assert config["ttl"] == 3600
        assert config["enabled"] is True

    def test_configure_cache_memcached(self, optimizer):
        """Test Memcached cache configuration."""
        config = optimizer.configure_cache(backend="memcached", ttl=1800)

        assert config["backend"] == "memcached"
        assert config["ttl"] == 1800

    def test_configure_cache_memory(self, optimizer):
        """Test in-memory cache configuration."""
        config = optimizer.configure_cache(backend="memory", ttl=600)

        assert config["backend"] == "memory"

    def test_configure_cache_with_key_pattern(self, optimizer):
        """Test cache configuration with key pattern."""
        config = optimizer.configure_cache(backend="redis", key_pattern="user:{user_id}:posts")

        assert config["key_pattern"] == "user:{user_id}:posts"

    def test_configure_cache_with_invalidation_triggers(self, optimizer):
        """Test cache configuration with invalidation triggers."""
        triggers = ["user_updated", "post_created"]
        config = optimizer.configure_cache(backend="redis", invalidation_triggers=triggers)

        assert config["invalidation_triggers"] == triggers

    def test_configure_cache_default_ttl(self, optimizer):
        """Test cache uses default TTL."""
        config = optimizer.configure_cache(backend="redis")

        assert config["ttl"] == 3600

    def test_configure_cache_default_triggers(self, optimizer):
        """Test cache uses default empty triggers."""
        config = optimizer.configure_cache(backend="redis")

        assert config["invalidation_triggers"] == []

    # ========================================================================
    # Rate Limit Configuration Tests
    # ========================================================================

    def test_configure_rate_limit_basic(self, optimizer):
        """Test basic rate limit configuration."""
        config = optimizer.configure_rate_limit(requests_per_minute=100, requests_per_hour=5000, burst_size=20)

        assert config["requests_per_minute"] == 100
        assert config["requests_per_hour"] == 5000
        assert config["burst_size"] == 20
        assert config["enabled"] is True

    def test_configure_rate_limit_custom_strategy(self, optimizer):
        """Test rate limit with custom strategy."""
        config = optimizer.configure_rate_limit(strategy="sliding_window", requests_per_minute=150)

        assert config["strategy"] == "sliding_window"

    def test_configure_rate_limit_token_bucket(self, optimizer):
        """Test token bucket rate limiting."""
        config = optimizer.configure_rate_limit(strategy="token_bucket", requests_per_minute=200, burst_size=50)

        assert config["strategy"] == "token_bucket"
        assert config["burst_size"] == 50

    def test_configure_rate_limit_default_values(self, optimizer):
        """Test rate limit default configuration."""
        config = optimizer.configure_rate_limit()

        assert config["requests_per_minute"] == 100
        assert config["requests_per_hour"] == 5000
        assert config["burst_size"] == 20
        assert config["strategy"] == "token_bucket"

    # ========================================================================
    # Query Optimization Tips Tests
    # ========================================================================

    def test_get_query_optimization_tips_select(self, optimizer):
        """Test SELECT query optimization tips."""
        tips = optimizer.get_query_optimization_tips("SELECT")

        assert len(tips) >= 3
        assert any("indexes" in tip.lower() for tip in tips)
        assert any("columns" in tip.lower() for tip in tips)

    def test_get_query_optimization_tips_join(self, optimizer):
        """Test JOIN query optimization tips."""
        tips = optimizer.get_query_optimization_tips("JOIN")

        assert len(tips) >= 3
        assert any("JOIN columns" in tip or "join columns" in tip.lower() for tip in tips)

    def test_get_query_optimization_tips_update(self, optimizer):
        """Test UPDATE query optimization tips."""
        tips = optimizer.get_query_optimization_tips("UPDATE")

        assert len(tips) >= 1
        assert any("batch" in tip.lower() for tip in tips)

    def test_get_query_optimization_tips_unknown(self, optimizer):
        """Test unknown query type returns empty tips."""
        tips = optimizer.get_query_optimization_tips("UNKNOWN")

        assert tips == []


# ============================================================================
# RequestMetric Dataclass Tests
# ============================================================================


class TestRequestMetric:
    """Test RequestMetric dataclass."""

    def test_request_metric_initialization(self):
        """Test RequestMetric creation."""
        metric = RequestMetric(
            path="/api/v1/users",
            method="GET",
            status_code=200,
            duration_ms=45.5,
            response_size_bytes=1024,
            timestamp="2025-01-01T00:00:00Z",
        )

        assert metric.path == "/api/v1/users"
        assert metric.method == "GET"
        assert metric.status_code == 200
        assert metric.duration_ms == 45.5
        assert metric.response_size_bytes == 1024

    def test_request_metric_different_methods(self):
        """Test RequestMetric with different HTTP methods."""
        for method in ["GET", "POST", "PUT", "DELETE"]:
            metric = RequestMetric(
                path="/test",
                method=method,
                status_code=200,
                duration_ms=10.0,
                response_size_bytes=100,
                timestamp="2025-01-01T00:00:00Z",
            )
            assert metric.method == method


# ============================================================================
# BackendMetricsCollector Tests
# ============================================================================


class TestBackendMetricsCollector:
    """Test BackendMetricsCollector class."""

    @pytest.fixture
    def collector(self):
        """Create metrics collector instance."""
        return BackendMetricsCollector()

    def test_collector_initialization(self, collector):
        """Test collector initializes."""
        assert collector.metrics == []
        assert collector.error_counts == {}

    # ========================================================================
    # Request Metrics Recording Tests
    # ========================================================================

    def test_record_request_metrics_basic(self, collector):
        """Test record basic request metrics."""
        result = collector.record_request_metrics(path="/api/v1/users", method="GET", status_code=200, duration_ms=45.5)

        assert result["path"] == "/api/v1/users"
        assert result["method"] == "GET"
        assert result["status_code"] == 200
        assert result["duration_ms"] == 45.5
        assert "timestamp" in result

    def test_record_request_metrics_with_response_size(self, collector):
        """Test record metrics with response size."""
        result = collector.record_request_metrics(
            path="/api/v1/users", method="GET", status_code=200, duration_ms=50.0, response_size_bytes=2048
        )

        assert result["response_size_bytes"] == 2048

    def test_record_request_metrics_stores_metric(self, collector):
        """Test recorded metric is stored."""
        collector.record_request_metrics(path="/api/v1/users", method="GET", status_code=200, duration_ms=30.0)

        assert len(collector.metrics) == 1
        assert collector.metrics[0].path == "/api/v1/users"

    def test_record_request_metrics_multiple_requests(self, collector):
        """Test recording multiple request metrics."""
        collector.record_request_metrics("/api/v1/users", "GET", 200, 30.0)
        collector.record_request_metrics("/api/v1/users", "POST", 201, 50.0)
        collector.record_request_metrics("/api/v1/users/1", "GET", 200, 25.0)

        assert len(collector.metrics) == 3

    def test_record_request_metrics_tracks_errors(self, collector):
        """Test error metrics are tracked."""
        collector.record_request_metrics(path="/api/v1/users", method="GET", status_code=404, duration_ms=20.0)

        assert "/api/v1/users:404" in collector.error_counts
        assert collector.error_counts["/api/v1/users:404"] == 1

    def test_record_request_metrics_multiple_errors_same_endpoint(self, collector):
        """Test multiple errors on same endpoint are counted."""
        collector.record_request_metrics("/api/v1/users", "GET", 404, 20.0)
        collector.record_request_metrics("/api/v1/users", "GET", 404, 20.0)

        assert collector.error_counts["/api/v1/users:404"] == 2

    def test_record_request_metrics_success_not_tracked(self, collector):
        """Test success responses not tracked in error_counts."""
        collector.record_request_metrics("/api/v1/users", "GET", 200, 30.0)

        assert len(collector.error_counts) == 0

    def test_record_request_metrics_various_error_codes(self, collector):
        """Test various error status codes."""
        collector.record_request_metrics("/api/v1/users", "POST", 400, 20.0)
        collector.record_request_metrics("/api/v1/users", "DELETE", 403, 20.0)
        collector.record_request_metrics("/api/v1/users", "GET", 500, 20.0)

        assert len(collector.error_counts) == 3

    # ========================================================================
    # Error Rate Calculation Tests
    # ========================================================================

    def test_get_error_rate_no_metrics(self, collector):
        """Test error rate with no metrics."""
        rate = collector.get_error_rate()

        assert rate == 0.0

    def test_get_error_rate_all_success(self, collector):
        """Test error rate when all requests succeed."""
        for _ in range(10):
            collector.record_request_metrics("/api/v1/users", "GET", 200, 30.0)

        rate = collector.get_error_rate()
        assert rate == 0.0

    def test_get_error_rate_all_errors(self, collector):
        """Test error rate when all requests fail."""
        for _ in range(5):
            collector.record_request_metrics("/api/v1/users", "GET", 500, 30.0)

        rate = collector.get_error_rate()
        assert rate == 1.0

    def test_get_error_rate_mixed(self, collector):
        """Test error rate with mixed success and error."""
        # 70 successes, 30 errors
        for _ in range(70):
            collector.record_request_metrics("/api/v1/users", "GET", 200, 30.0)
        for _ in range(30):
            collector.record_request_metrics("/api/v1/users", "GET", 500, 30.0)

        rate = collector.get_error_rate()
        assert rate == 0.3

    def test_get_error_rate_specific_path(self, collector):
        """Test error rate for specific path."""
        # Path 1: 5 successes, 5 errors (50% error rate)
        for _ in range(5):
            collector.record_request_metrics("/api/v1/users", "GET", 200, 30.0)
        for _ in range(5):
            collector.record_request_metrics("/api/v1/users", "GET", 500, 30.0)

        # Path 2: 10 successes, 0 errors
        for _ in range(10):
            collector.record_request_metrics("/api/v1/posts", "GET", 200, 30.0)

        path1_rate = collector.get_error_rate("/api/v1/users")
        path2_rate = collector.get_error_rate("/api/v1/posts")

        assert path1_rate == 0.5
        assert path2_rate == 0.0

    def test_get_error_rate_nonexistent_path(self, collector):
        """Test error rate for nonexistent path."""
        collector.record_request_metrics("/api/v1/users", "GET", 200, 30.0)

        rate = collector.get_error_rate("/api/v1/nonexistent")
        assert rate == 0.0

    # ========================================================================
    # Service Health Status Tests
    # ========================================================================

    def test_get_service_health_no_metrics(self, collector):
        """Test service health with no metrics."""
        health = collector.get_service_health()

        assert "status" in health
        assert "metrics" in health
        assert health["status"] == "healthy"

    def test_get_service_health_healthy(self, collector):
        """Test service health when error rate < 1%."""
        # 999 successes, 1 error (0.1% error rate < 1%)
        for _ in range(999):
            collector.record_request_metrics("/api/v1/users", "GET", 200, 30.0)
        collector.record_request_metrics("/api/v1/users", "GET", 500, 30.0)

        health = collector.get_service_health()
        assert health["status"] == "healthy"

    def test_get_service_health_degraded(self, collector):
        """Test service health when error rate 1-5%."""
        # 96 successes, 4 errors (4% error rate is between 1% and 5%)
        for _ in range(96):
            collector.record_request_metrics("/api/v1/users", "GET", 200, 30.0)
        for _ in range(4):
            collector.record_request_metrics("/api/v1/users", "GET", 500, 30.0)

        health = collector.get_service_health()
        assert health["status"] == "degraded"

    def test_get_service_health_unhealthy(self, collector):
        """Test service health when error rate > 5%."""
        # 50 successes, 50 errors (50% error rate)
        for _ in range(50):
            collector.record_request_metrics("/api/v1/users", "GET", 200, 30.0)
        for _ in range(50):
            collector.record_request_metrics("/api/v1/users", "GET", 500, 30.0)

        health = collector.get_service_health()
        assert health["status"] == "unhealthy"

    def test_get_service_health_includes_metrics(self, collector):
        """Test service health includes detailed metrics."""
        collector.record_request_metrics("/api/v1/users", "GET", 200, 50.0)
        collector.record_request_metrics("/api/v1/users", "POST", 201, 100.0)

        health = collector.get_service_health()

        assert health["metrics"]["total_requests"] == 2
        assert health["metrics"]["error_rate"] == 0.0
        assert "avg_duration_ms" in health["metrics"]
        assert "uptime_seconds" in health["metrics"]

    def test_get_service_health_timestamp(self, collector):
        """Test service health includes timestamp."""
        health = collector.get_service_health()

        assert "timestamp" in health
        assert health["timestamp"].endswith("Z")

    def test_get_service_health_error_summary(self, collector):
        """Test service health includes error summary."""
        collector.record_request_metrics("/api/v1/users", "GET", 404, 20.0)
        collector.record_request_metrics("/api/v1/posts", "GET", 500, 20.0)

        health = collector.get_service_health()

        assert "error_summary" in health
        assert len(health["error_summary"]) == 2

    # ========================================================================
    # Metrics Summary Tests
    # ========================================================================

    def test_get_metrics_summary_empty(self, collector):
        """Test metrics summary with no data."""
        summary = collector.get_metrics_summary()

        assert summary["total_requests"] == 0
        assert "health" in summary
        assert "error_rate" in summary

    def test_get_metrics_summary_with_data(self, collector):
        """Test metrics summary with data."""
        collector.record_request_metrics("/api/v1/users", "GET", 200, 30.0)
        collector.record_request_metrics("/api/v1/users", "POST", 201, 50.0)
        collector.record_request_metrics("/api/v1/posts", "GET", 200, 40.0)

        summary = collector.get_metrics_summary()

        assert summary["total_requests"] == 3
        assert summary["error_rate"] == 0.0
        assert summary["endpoints"] == 2

    def test_get_metrics_summary_includes_health(self, collector):
        """Test metrics summary includes health information."""
        collector.record_request_metrics("/api/v1/users", "GET", 200, 30.0)

        summary = collector.get_metrics_summary()

        assert "health" in summary
        assert "status" in summary["health"]

    def test_get_metrics_summary_endpoints_count(self, collector):
        """Test unique endpoints count in summary."""
        collector.record_request_metrics("/api/v1/users", "GET", 200, 30.0)
        collector.record_request_metrics("/api/v1/users", "POST", 201, 50.0)
        collector.record_request_metrics("/api/v1/posts", "GET", 200, 40.0)
        collector.record_request_metrics("/api/v1/posts", "PUT", 200, 60.0)

        summary = collector.get_metrics_summary()

        # Only 2 unique paths
        assert summary["endpoints"] == 2


# ============================================================================
# Integration Tests
# ============================================================================


class TestBackendIntegration:
    """Integration tests for backend components."""

    def test_api_validation_and_error_handling(self):
        """Test API validation with error handling."""
        validator = APIDesignValidator()
        handler = ErrorHandlingStrategy()

        # Validate invalid endpoint
        result = validator.validate_rest_endpoint({"method": "POST", "path": "/api/v1/users", "status_code": 200})

        # Handle validation error
        if not result["valid"]:
            error_response = handler.handle_error(
                {
                    "type": "ValidationError",
                    "message": f"Endpoint validation failed: {result['errors']}",
                    "status_code": 400,
                }
            )

            assert error_response["type"] == "ValidationError"
            assert error_response["status_code"] == 400

    def test_microservice_with_auth(self):
        """Test microservice with authentication."""
        architect = MicroserviceArchitect()
        auth = AuthenticationManager(secret_key="test-secret")

        # Validate service
        service = {"name": "user-service", "domain": "auth", "endpoints": ["/api/v1/users"]}
        result = architect.validate_service_boundary(service)
        assert result["valid"] is True

        # Generate token for service communication
        token = auth.generate_jwt_token({"sub": "user-service", "aud": "api-gateway"})

        payload = auth.validate_jwt_token(token)
        assert payload["sub"] == "user-service"

    def test_metrics_with_error_tracking(self):
        """Test metrics collection with error tracking."""
        collector = BackendMetricsCollector()

        # Simulate requests
        collector.record_request_metrics("/api/v1/users", "GET", 200, 30.0)
        collector.record_request_metrics("/api/v1/users", "GET", 200, 35.0)
        collector.record_request_metrics("/api/v1/users", "GET", 404, 25.0)

        # Check metrics
        health = collector.get_service_health()
        assert health["metrics"]["total_requests"] == 3
        assert health["metrics"]["error_rate"] > 0

    @pytest.mark.asyncio
    async def test_async_with_auth_tokens(self):
        """Test async operations with authentication."""
        advisor = AsyncPatternAdvisor()
        auth = AuthenticationManager(secret_key="test-secret")

        async def generate_token():
            return auth.generate_jwt_token({"sub": "user1"})

        async def validate_token(token):
            return auth.validate_jwt_token(token)

        # Execute async operations
        tokens = await advisor.execute_concurrent([generate_token, generate_token])
        assert len(tokens) == 2

        # Validate tokens
        for token in tokens:
            payload = auth.validate_jwt_token(token)
            assert payload["sub"] == "user1"
