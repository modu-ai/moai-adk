"""
Comprehensive TDD tests for backend.py module.
Tests cover all 7 classes.
"""

import asyncio
from datetime import UTC, datetime

import pytest

from moai_adk.foundation.backend import (
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
# Test APIDesignValidator
# ============================================================================


class TestAPIDesignValidator:
    """Test suite for APIDesignValidator class."""

    def test_initialization(self):
        """Test validator initialization."""
        validator = APIDesignValidator()
        assert validator.validated_endpoints == {}
        assert validator.VALID_HTTP_METHODS == {m.value for m in HTTPMethod}
        assert 200 in validator.VALID_STATUS_CODES
        assert 404 in validator.VALID_STATUS_CODES

    def test_validate_rest_endpoint_valid(self, sample_endpoint):
        """Test validation of valid REST endpoint."""
        validator = APIDesignValidator()
        result = validator.validate_rest_endpoint(sample_endpoint)

        assert result["valid"] is True
        assert result["method"] == "GET"
        assert result["path"] == "/api/v1/users"
        assert result["status_code"] == 200
        assert result["errors"] is None

    @pytest.mark.parametrize("method,status_code", [
        ("GET", 200),
        ("POST", 201),
        ("PUT", 200),
        ("PATCH", 200),
        ("DELETE", 204),
    ])
    def test_validate_rest_endpoint_valid_methods(self, method, status_code):
        """Test validation with valid method-status code combinations."""
        validator = APIDesignValidator()
        endpoint = {"method": method, "path": "/api/v1/resource", "status_code": status_code}
        result = validator.validate_rest_endpoint(endpoint)

        assert result["valid"] is True
        assert result["method"] == method
        assert result["status_code"] == status_code

    def test_validate_rest_endpoint_invalid_method(self):
        """Test validation with invalid HTTP method."""
        validator = APIDesignValidator()
        endpoint = {"method": "INVALID", "path": "/api/v1/resource", "status_code": 200}
        result = validator.validate_rest_endpoint(endpoint)

        assert result["valid"] is False
        assert "Invalid HTTP method" in result["errors"][0]

    def test_validate_rest_endpoint_invalid_path(self):
        """Test validation with invalid path format."""
        validator = APIDesignValidator()
        endpoint = {"method": "GET", "path": "invalid-path", "status_code": 200}
        result = validator.validate_rest_endpoint(endpoint)

        assert result["valid"] is False
        assert "Path must start with" in result["errors"][0]

    def test_validate_rest_endpoint_invalid_status_code(self):
        """Test validation with invalid status code."""
        validator = APIDesignValidator()
        endpoint = {"method": "GET", "path": "/api/v1/resource", "status_code": 999}
        result = validator.validate_rest_endpoint(endpoint)

        assert result["valid"] is False
        assert "Invalid status code" in result["errors"][0]

    def test_validate_rest_endpoint_wrong_status_for_method(self):
        """Test validation with mismatched method and status code."""
        validator = APIDesignValidator()
        endpoint = {"method": "POST", "path": "/api/v1/resource", "status_code": 200}
        result = validator.validate_rest_endpoint(endpoint)

        assert result["valid"] is False
        assert "not allowed for POST" in result["errors"][0]

    def test_get_versioning_strategy_url(self):
        """Test URL versioning strategy."""
        validator = APIDesignValidator()
        strategy = validator.get_versioning_strategy("url")

        assert strategy["strategy"] == "url"
        assert "/api/v{version}/" in strategy["pattern"]
        assert "SEO friendly" in strategy["pros"]

    def test_get_versioning_strategy_header(self):
        """Test header versioning strategy."""
        validator = APIDesignValidator()
        strategy = validator.get_versioning_strategy("header")

        assert strategy["strategy"] == "header"
        assert strategy["header_name"] == "API-Version"
        assert "Clean URLs" in strategy["pros"]

    def test_get_versioning_strategy_content_type(self):
        """Test content-type versioning strategy."""
        validator = APIDesignValidator()
        strategy = validator.get_versioning_strategy("content-type")

        assert strategy["strategy"] == "content-type"
        assert "application/vnd.api+json" in strategy["content_type"]
        assert "RESTful" in strategy["pros"]

    def test_get_versioning_strategy_default(self):
        """Test default versioning strategy."""
        validator = APIDesignValidator()
        strategy = validator.get_versioning_strategy("invalid")

        # Should return URL strategy as default
        assert strategy["strategy"] == "url"

    def test_standardize_error_response(self):
        """Test error response standardization."""
        validator = APIDesignValidator()
        error = {
            "type": "ValidationError",
            "message": "Invalid input",
            "status_code": 400,
            "details": {"field": "email"}
        }

        result = validator.standardize_error_response(error)

        assert result["type"] == "ValidationError"
        assert result["message"] == "Invalid input"
        assert result["status_code"] == 400
        assert result["details"] == {"field": "email"}
        assert "trace_id" in result
        assert "timestamp" in result

    def test_standardize_error_response_defaults(self):
        """Test error response with default values."""
        validator = APIDesignValidator()
        error = {}

        result = validator.standardize_error_response(error)

        assert result["type"] == "Error"
        assert result["message"] == "An error occurred"
        assert result["status_code"] == 500
        assert result["details"] == {}
        assert result["path"] == ""


# ============================================================================
# Test MicroserviceArchitect
# ============================================================================


class TestMicroserviceArchitect:
    """Test suite for MicroserviceArchitect class."""

    def test_initialization(self):
        """Test architect initialization."""
        architect = MicroserviceArchitect()
        assert architect.services == {}
        assert architect.communication_matrix == {}
        assert "rest" in architect.COMMUNICATION_PATTERNS

    def test_validate_service_boundary_valid(self, sample_service):
        """Test validation of valid service boundary."""
        architect = MicroserviceArchitect()
        result = architect.validate_service_boundary(sample_service)

        assert result["valid"] is True
        assert result["name"] == "user-service"
        assert result["domain"] == "auth"
        assert len(result["endpoints"]) == 2
        assert result["errors"] is None

    def test_validate_service_boundary_invalid_name(self):
        """Test validation with invalid service name."""
        architect = MicroserviceArchitect()
        service = {"name": "invalid", "domain": "auth", "endpoints": ["/path"]}
        result = architect.validate_service_boundary(service)

        assert result["valid"] is False
        assert "Service name should follow pattern" in result["errors"][0]

    def test_validate_service_boundary_missing_domain(self):
        """Test validation without domain."""
        architect = MicroserviceArchitect()
        service = {"name": "user-service", "endpoints": ["/path"]}
        result = architect.validate_service_boundary(service)

        assert result["valid"] is False
        assert "domain/bounded context" in result["errors"][0]

    def test_validate_service_boundary_no_endpoints(self):
        """Test validation without endpoints."""
        architect = MicroserviceArchitect()
        service = {"name": "user-service", "domain": "auth", "endpoints": []}
        result = architect.validate_service_boundary(service)

        assert result["valid"] is False
        assert "at least one endpoint" in result["errors"][0]

    def test_get_communication_pattern_rest(self):
        """Test REST communication pattern."""
        architect = MicroserviceArchitect()
        context = {"source": "service-a", "target": "service-b", "operation": "query"}

        result = architect.get_communication_pattern("rest", context)

        assert result["pattern"] == "rest"
        assert result["protocol"] == "HTTP/REST"
        assert result["async"] is False
        assert result["source"] == "service-a"

    def test_get_communication_pattern_async(self):
        """Test async communication pattern."""
        architect = MicroserviceArchitect()
        context = {"source": "service-a", "target": "service-b"}

        result = architect.get_communication_pattern("async", context)

        assert result["pattern"] == "async"
        assert result["async"] is True
        assert "RabbitMQ" in result["protocol"]

    def test_get_communication_pattern_grpc(self):
        """Test gRPC communication pattern."""
        architect = MicroserviceArchitect()
        context = {"source": "service-a", "target": "service-b"}

        result = architect.get_communication_pattern("grpc", context)

        assert result["pattern"] == "grpc"
        assert result["protocol"] == "gRPC"
        assert result["async"] is True

    def test_configure_service_discovery_consul(self):
        """Test Consul service discovery configuration."""
        architect = MicroserviceArchitect()
        config = {"consul_host": "localhost", "consul_port": 8500}

        result = architect.configure_service_discovery("consul", config)

        assert result["registry"] == "consul"
        assert result["host"] == "localhost"
        assert result["port"] == 8500
        assert result["health_check_enabled"] is True
        assert result["auto_deregister"] is True

    def test_configure_service_discovery_eureka(self):
        """Test Eureka service discovery configuration."""
        architect = MicroserviceArchitect()
        config = {}

        result = architect.configure_service_discovery("eureka", config)

        assert result["registry"] == "eureka"
        assert result["health_check_enabled"] is True
        assert result["auto_deregister"] is False

    def test_configure_service_discovery_etcd(self):
        """Test etcd service discovery configuration."""
        architect = MicroserviceArchitect()
        config = {"health_check_interval": 15}

        result = architect.configure_service_discovery("etcd", config)

        assert result["registry"] == "etcd"
        assert result["health_check_enabled"] is True
        assert result["health_check_interval"] == 15


# ============================================================================
# Test AsyncPatternAdvisor
# ============================================================================


class TestAsyncPatternAdvisor:
    """Test suite for AsyncPatternAdvisor class."""

    def test_initialization(self):
        """Test advisor initialization."""
        advisor = AsyncPatternAdvisor()
        assert advisor.async_operations == []

    @pytest.mark.asyncio
    async def test_execute_concurrent_basic(self):
        """Test basic concurrent execution."""
        advisor = AsyncPatternAdvisor()

        async def operation1():
            return "result1"

        async def operation2():
            return "result2"

        results = await advisor.execute_concurrent([operation1, operation2])

        assert len(results) == 2
        assert "result1" in results
        assert "result2" in results

    @pytest.mark.asyncio
    async def test_execute_concurrent_with_timeout(self):
        """Test concurrent execution with timeout."""
        advisor = AsyncPatternAdvisor()

        async def quick_operation():
            return "quick"

        results = await advisor.execute_concurrent([quick_operation], timeout=5.0)

        assert results == ["quick"]

    @pytest.mark.asyncio
    async def test_execute_concurrent_timeout_error(self):
        """Test timeout handling in concurrent execution."""
        advisor = AsyncPatternAdvisor()

        async def slow_operation():
            await asyncio.sleep(10)
            return "slow"

        with pytest.raises(asyncio.TimeoutError):
            await advisor.execute_concurrent([slow_operation], timeout=0.5)

    @pytest.mark.asyncio
    async def test_with_timeout_success(self):
        """Test timeout wrapper with successful execution."""
        advisor = AsyncPatternAdvisor()

        async def quick_coro():
            return "done"

        result = await advisor.with_timeout(quick_coro(), 5.0)

        assert result == "done"

    @pytest.mark.asyncio
    async def test_with_timeout_failure(self):
        """Test timeout wrapper with timeout exceeded."""
        advisor = AsyncPatternAdvisor()

        async def slow_coro():
            await asyncio.sleep(10)
            return "done"

        with pytest.raises(asyncio.TimeoutError):
            await advisor.with_timeout(slow_coro(), 0.5)

    @pytest.mark.asyncio
    async def test_async_retry_success(self):
        """Test async retry decorator with success."""
        advisor = AsyncPatternAdvisor()

        attempt_count = 0

        @advisor.async_retry(max_attempts=3, backoff_factor=1.0)
        async def flaky_function():
            nonlocal attempt_count
            attempt_count += 1
            if attempt_count < 3:
                raise ValueError("Temporary failure")
            return "success"

        result = await flaky_function()

        assert result == "success"
        assert attempt_count == 3

    @pytest.mark.asyncio
    async def test_async_retry_exhausted(self):
        """Test async retry decorator with exhausted attempts."""
        advisor = AsyncPatternAdvisor()

        @advisor.async_retry(max_attempts=2, backoff_factor=0.1)
        async def failing_function():
            raise ValueError("Permanent failure")

        with pytest.raises(ValueError):
            await failing_function()

    def test_async_retry_parameters(self):
        """Test async retry with custom parameters."""
        advisor = AsyncPatternAdvisor()

        @advisor.async_retry(max_attempts=5, backoff_factor=3.0)
        async def dummy_function():
            return "result"

        assert callable(dummy_function)


# ============================================================================
# Test AuthenticationManager
# ============================================================================


class TestAuthenticationManager:
    """Test suite for AuthenticationManager class."""

    def test_initialization(self):
        """Test manager initialization with default secret."""
        auth = AuthenticationManager()
        assert auth.secret_key == "default-secret-key"
        assert auth.algorithms == ["HS256"]
        assert auth.oauth_codes == {}

    def test_initialization_custom_secret(self):
        """Test manager initialization with custom secret."""
        auth = AuthenticationManager(secret_key="custom-secret")
        assert auth.secret_key == "custom-secret"

    def test_generate_jwt_token_basic(self):
        """Test basic JWT token generation."""
        auth = AuthenticationManager()
        data = {"sub": "user@example.com", "name": "Test User"}

        token = auth.generate_jwt_token(data)

        assert isinstance(token, str)
        assert len(token.split(".")) == 3  # header.payload.signature

    def test_generate_jwt_token_with_expiration(self):
        """Test JWT token generation with custom expiration."""
        auth = AuthenticationManager()
        data = {"sub": "user@example.com"}

        token = auth.generate_jwt_token(data, expires_in_hours=24)

        assert isinstance(token, str)
        assert len(token) > 0

    def test_validate_jwt_token_valid(self):
        """Test validation of valid JWT token."""
        auth = AuthenticationManager()
        data = {"sub": "user@example.com", "user_id": "123"}

        token = auth.generate_jwt_token(data)
        payload = auth.validate_jwt_token(token)

        assert payload["sub"] == "user@example.com"
        assert payload["user_id"] == "123"
        assert "iat" in payload
        assert "exp" in payload

    def test_validate_jwt_token_expired(self):
        """Test validation of expired JWT token."""
        auth = AuthenticationManager()

        # Create a token with negative expiration (already expired)
        import base64
        import hashlib
        import hmac
        import json

        header = {"alg": "HS256", "typ": "JWT"}
        payload = {
            "sub": "user@example.com",
            "iat": int(datetime.now(UTC).timestamp()) - 10000,
            "exp": int(datetime.now(UTC).timestamp()) - 5000
        }

        header_encoded = base64.urlsafe_b64encode(json.dumps(header).encode()).decode().rstrip("=")
        payload_encoded = base64.urlsafe_b64encode(json.dumps(payload).encode()).decode().rstrip("=")

        message = f"{header_encoded}.{payload_encoded}"
        signature = hmac.new(auth.secret_key.encode(), message.encode(), hashlib.sha256).digest()
        signature_encoded = base64.urlsafe_b64encode(signature).decode().rstrip("=")

        token = f"{message}.{signature_encoded}"

        with pytest.raises(ValueError, match="Token expired"):
            auth.validate_jwt_token(token)

    def test_validate_jwt_token_invalid_format(self):
        """Test validation of invalid JWT token format."""
        auth = AuthenticationManager()

        with pytest.raises(ValueError, match="Invalid token"):
            auth.validate_jwt_token("invalid.token.format")

    def test_generate_oauth_auth_code(self):
        """Test OAuth authorization code generation."""
        auth = AuthenticationManager()
        params = {"client_id": "test_client", "redirect_uri": "https://example.com", "state": "random_state"}

        result = auth.generate_oauth_auth_code(params)

        assert "code" in result
        assert result["expires_in"] == 600
        assert result["state"] == "random_state"
        assert result["code"] in auth.oauth_codes

    def test_has_permission_true(self, sample_user_permissions):
        """Test permission check with valid permission."""
        auth = AuthenticationManager()

        result = auth.has_permission(sample_user_permissions, "read:users")

        assert result is True

    def test_has_permission_false(self, sample_user_permissions):
        """Test permission check with invalid permission."""
        auth = AuthenticationManager()

        result = auth.has_permission(sample_user_permissions, "delete:everything")

        assert result is False

    def test_has_permission_empty_permissions(self):
        """Test permission check with no permissions."""
        auth = AuthenticationManager()
        user = {"id": "user-123", "permissions": []}

        result = auth.has_permission(user, "read:users")

        assert result is False


# ============================================================================
# Test ErrorHandlingStrategy
# ============================================================================


class TestErrorHandlingStrategy:
    """Test suite for ErrorHandlingStrategy class."""

    def test_initialization(self):
        """Test strategy initialization."""
        handler = ErrorHandlingStrategy()
        assert handler.error_handlers == {}
        assert handler.logs == []

    def test_handle_error_basic(self):
        """Test basic error handling."""
        handler = ErrorHandlingStrategy()
        error = {
            "type": "ValidationError",
            "message": "Invalid email format",
            "status_code": 400
        }

        result = handler.handle_error(error)

        assert result["type"] == "ValidationError"
        assert result["message"] == "Invalid email format"
        assert result["status_code"] == 400
        assert "timestamp" in result
        assert "trace_id" in result

    def test_handle_error_with_details(self):
        """Test error handling with details."""
        handler = ErrorHandlingStrategy()
        error = {
            "type": "DatabaseError",
            "message": "Connection failed",
            "status_code": 500,
            "details": {"host": "localhost", "port": 5432}
        }

        result = handler.handle_error(error)

        assert result["details"] == {"host": "localhost", "port": 5432}

    def test_handle_error_defaults(self):
        """Test error handling with default values."""
        handler = ErrorHandlingStrategy()
        error = {}

        result = handler.handle_error(error)

        assert result["type"] == "Error"
        assert result["message"] == "An error occurred"
        assert result["status_code"] == 500
        assert result["details"] == {}

    def test_log_with_context_basic(self):
        """Test basic logging with context."""
        handler = ErrorHandlingStrategy()
        log_entry = handler.log_with_context("INFO", "Test message")

        assert log_entry["level"] == "INFO"
        assert log_entry["message"] == "Test message"
        assert "timestamp" in log_entry
        assert "trace_id" in log_entry
        assert log_entry["context"] == {}

    def test_log_with_context_with_context(self):
        """Test logging with additional context."""
        handler = ErrorHandlingStrategy()
        context = {"user_id": "123", "request_id": "abc"}
        log_entry = handler.log_with_context("ERROR", "Request failed", context)

        assert log_entry["context"] == context
        assert log_entry["level"] == "ERROR"

    def test_log_with_context_storage(self):
        """Test that logs are stored in handler."""
        handler = ErrorHandlingStrategy()
        handler.log_with_context("INFO", "Test 1")
        handler.log_with_context("WARNING", "Test 2")

        assert len(handler.logs) == 2
        assert handler.logs[0].message == "Test 1"
        assert handler.logs[1].level == "WARNING"


# ============================================================================
# Test PerformanceOptimizer
# ============================================================================


class TestPerformanceOptimizer:
    """Test suite for PerformanceOptimizer class."""

    def test_initialization(self):
        """Test optimizer initialization."""
        optimizer = PerformanceOptimizer()
        assert optimizer.cache_configs == {}
        assert optimizer.rate_limits == {}

    def test_configure_cache_basic(self):
        """Test basic cache configuration."""
        optimizer = PerformanceOptimizer()
        config = optimizer.configure_cache()

        assert config["backend"] == "redis"
        assert config["ttl"] == 3600
        assert config["enabled"] is True
        assert config["invalidation_triggers"] == []

    def test_configure_cache_custom(self):
        """Test cache configuration with custom parameters."""
        optimizer = PerformanceOptimizer()
        config = optimizer.configure_cache(
            backend="memcached",
            ttl=7200,
            key_pattern="user:*",
            invalidation_triggers=["user_updated"]
        )

        assert config["backend"] == "memcached"
        assert config["ttl"] == 7200
        assert config["key_pattern"] == "user:*"
        assert "user_updated" in config["invalidation_triggers"]

    def test_configure_rate_limit_basic(self):
        """Test basic rate limiting configuration."""
        optimizer = PerformanceOptimizer()
        config = optimizer.configure_rate_limit()

        assert config["requests_per_minute"] == 100
        assert config["requests_per_hour"] == 5000
        assert config["burst_size"] == 20
        assert config["strategy"] == "token_bucket"
        assert config["enabled"] is True

    def test_configure_rate_limit_custom(self):
        """Test rate limiting with custom parameters."""
        optimizer = PerformanceOptimizer()
        config = optimizer.configure_rate_limit(
            requests_per_minute=200,
            requests_per_hour=10000,
            burst_size=50,
            strategy="sliding_window"
        )

        assert config["requests_per_minute"] == 200
        assert config["requests_per_hour"] == 10000
        assert config["burst_size"] == 50
        assert config["strategy"] == "sliding_window"

    def test_get_query_optimization_tips_select(self):
        """Test query optimization tips for SELECT."""
        optimizer = PerformanceOptimizer()
        tips = optimizer.get_query_optimization_tips("SELECT")

        assert len(tips) > 0
        assert any("indexes" in tip.lower() for tip in tips)

    def test_get_query_optimization_tips_join(self):
        """Test query optimization tips for JOIN."""
        optimizer = PerformanceOptimizer()
        tips = optimizer.get_query_optimization_tips("JOIN")

        assert len(tips) > 0
        assert any("indexed" in tip.lower() for tip in tips)

    def test_get_query_optimization_tips_update(self):
        """Test query optimization tips for UPDATE."""
        optimizer = PerformanceOptimizer()
        tips = optimizer.get_query_optimization_tips("UPDATE")

        assert len(tips) > 0
        assert any("batch" in tip.lower() for tip in tips)

    def test_get_query_optimization_tips_unknown(self):
        """Test query optimization tips for unknown query type."""
        optimizer = PerformanceOptimizer()
        tips = optimizer.get_query_optimization_tips("UNKNOWN")

        assert tips == []


# ============================================================================
# Test BackendMetricsCollector
# ============================================================================


class TestBackendMetricsCollector:
    """Test suite for BackendMetricsCollector class."""

    def test_initialization(self):
        """Test collector initialization."""
        collector = BackendMetricsCollector()
        assert collector.metrics == []
        assert collector.error_counts == {}

    def test_record_request_metrics_basic(self):
        """Test basic request metrics recording."""
        collector = BackendMetricsCollector()
        metric = collector.record_request_metrics(
            path="/api/v1/users",
            method="GET",
            status_code=200,
            duration_ms=45.5
        )

        assert metric["path"] == "/api/v1/users"
        assert metric["method"] == "GET"
        assert metric["status_code"] == 200
        assert metric["duration_ms"] == 45.5
        assert metric["response_size_bytes"] == 0
        assert "timestamp" in metric

    def test_record_request_metrics_with_size(self):
        """Test request metrics with response size."""
        collector = BackendMetricsCollector()
        metric = collector.record_request_metrics(
            path="/api/v1/users",
            method="POST",
            status_code=201,
            duration_ms=120.3,
            response_size_bytes=1024
        )

        assert metric["response_size_bytes"] == 1024

    def test_record_request_metrics_error_tracking(self):
        """Test error tracking in metrics."""
        collector = BackendMetricsCollector()
        collector.record_request_metrics("/api/v1/error", "GET", 500, 100)
        collector.record_request_metrics("/api/v1/error", "GET", 500, 100)

        assert "/api/v1/error:500" in collector.error_counts
        assert collector.error_counts["/api/v1/error:500"] == 2

    def test_get_error_rate_empty(self):
        """Test error rate with no metrics."""
        collector = BackendMetricsCollector()
        error_rate = collector.get_error_rate()

        assert error_rate == 0.0

    def test_get_error_rate_all_success(self):
        """Test error rate with all successful requests."""
        collector = BackendMetricsCollector()
        collector.record_request_metrics("/api/v1/users", "GET", 200, 50)
        collector.record_request_metrics("/api/v1/users", "GET", 200, 50)

        error_rate = collector.get_error_rate()

        assert error_rate == 0.0

    def test_get_error_rate_with_errors(self):
        """Test error rate with mixed success and errors."""
        collector = BackendMetricsCollector()
        collector.record_request_metrics("/api/v1/users", "GET", 200, 50)
        collector.record_request_metrics("/api/v1/users", "GET", 500, 50)
        collector.record_request_metrics("/api/v1/users", "GET", 404, 50)

        error_rate = collector.get_error_rate()

        assert error_rate == 2 / 3

    def test_get_error_rate_filtered_by_path(self):
        """Test error rate filtered by specific path."""
        collector = BackendMetricsCollector()
        collector.record_request_metrics("/api/v1/users", "GET", 200, 50)
        collector.record_request_metrics("/api/v1/posts", "GET", 500, 50)

        error_rate = collector.get_error_rate(path="/api/v1/posts")

        assert error_rate == 1.0

    def test_get_service_health_healthy(self):
        """Test service health status when healthy."""
        collector = BackendMetricsCollector()
        collector.record_request_metrics("/api/v1/users", "GET", 200, 50)
        collector.record_request_metrics("/api/v1/users", "GET", 200, 50)

        health = collector.get_service_health()

        assert health["status"] == "healthy"
        assert health["metrics"]["total_requests"] == 2
        assert health["metrics"]["error_rate"] == 0.0

    def test_get_service_health_degraded(self):
        """Test service health status when degraded."""
        collector = BackendMetricsCollector()
        for i in range(100):
            status = 500 if i < 4 else 200  # 4% error rate
            collector.record_request_metrics("/api/v1/users", "GET", status, 50)

        health = collector.get_service_health()

        assert health["status"] == "degraded"

    def test_get_service_health_unhealthy(self):
        """Test service health status when unhealthy."""
        collector = BackendMetricsCollector()
        for i in range(100):
            status = 500 if i < 10 else 200  # 10% error rate
            collector.record_request_metrics("/api/v1/users", "GET", status, 50)

        health = collector.get_service_health()

        assert health["status"] == "unhealthy"

    def test_get_metrics_summary(self):
        """Test comprehensive metrics summary."""
        collector = BackendMetricsCollector()
        collector.record_request_metrics("/api/v1/users", "GET", 200, 50)
        collector.record_request_metrics("/api/v1/posts", "GET", 200, 60)

        summary = collector.get_metrics_summary()

        assert summary["total_requests"] == 2
        assert summary["endpoints"] == 2
        assert "health" in summary
        assert "error_rate" in summary


# ============================================================================
# Test Data Classes
# ============================================================================


class TestDataClasses:
    """Test suite for data classes."""

    def test_error_log_creation(self):
        """Test ErrorLog dataclass creation."""
        error_log = ErrorLog(
            level="ERROR",
            message="Test error",
            timestamp="2025-01-13T10:00:00Z",
            trace_id="trace-123",
            context={"user_id": "123"}
        )

        assert error_log.level == "ERROR"
        assert error_log.message == "Test error"
        assert error_log.context == {"user_id": "123"}

    def test_request_metric_creation(self):
        """Test RequestMetric dataclass creation."""
        metric = RequestMetric(
            path="/api/v1/test",
            method="GET",
            status_code=200,
            duration_ms=45.5,
            response_size_bytes=1024,
            timestamp="2025-01-13T10:00:00Z"
        )

        assert metric.path == "/api/v1/test"
        assert metric.duration_ms == 45.5
        assert metric.response_size_bytes == 1024
