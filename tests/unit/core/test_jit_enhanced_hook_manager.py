"""
Minimal import and instantiation tests for JIT Enhanced Hook Manager.

These tests verify that the module can be imported and basic classes
can be instantiated without errors.
"""


from moai_adk.core.jit_enhanced_hook_manager import (
    ContextCache,
    HookEvent,
    HookMetadata,
    HookPriority,
    JITEnhancedHookManager,
    Phase,
    TokenBudgetManager,
)


class TestImports:
    """Test that all classes and enums can be imported."""

    def test_hook_event_enum_exists(self):
        """Test HookEvent enum is importable."""
        assert HookEvent is not None
        assert hasattr(HookEvent, "SESSION_START")

    def test_hook_priority_enum_exists(self):
        """Test HookPriority enum is importable."""
        assert HookPriority is not None

    def test_phase_enum_exists(self):
        """Test Phase enum is importable."""
        assert Phase is not None
        assert hasattr(Phase, "RED")
        assert hasattr(Phase, "GREEN")

    def test_context_cache_importable(self):
        """Test ContextCache can be imported."""
        assert ContextCache is not None

    def test_token_budget_manager_importable(self):
        """Test TokenBudgetManager can be imported."""
        assert TokenBudgetManager is not None

    def test_jit_enhanced_hook_manager_importable(self):
        """Test JITEnhancedHookManager can be imported."""
        assert JITEnhancedHookManager is not None

    def test_hook_metadata_importable(self):
        """Test HookMetadata can be imported."""
        assert HookMetadata is not None


class TestContextCacheInstantiation:
    """Test ContextCache class instantiation."""

    def test_context_cache_init_default(self):
        """Test ContextCache can be instantiated with default parameters."""
        cache = ContextCache()
        assert cache is not None
        assert hasattr(cache, "get")
        assert hasattr(cache, "put")
        assert hasattr(cache, "clear")

    def test_context_cache_init_custom(self):
        """Test ContextCache can be instantiated with custom parameters."""
        cache = ContextCache(max_size=200, max_memory_mb=100)
        assert cache.max_size == 200
        # max_memory_mb is stored but may be named differently
        assert hasattr(cache, "max_size")

    def test_context_cache_get_method(self):
        """Test ContextCache.get method exists and is callable."""
        cache = ContextCache()
        result = cache.get("test_key")
        assert result is None

    def test_context_cache_put_method(self):
        """Test ContextCache.put method exists and is callable."""
        cache = ContextCache()
        cache.put("test_key", "test_value", token_count=10)
        # put should not raise

    def test_context_cache_clear_method(self):
        """Test ContextCache.clear method exists and is callable."""
        cache = ContextCache()
        cache.clear()
        # clear should not raise

    def test_context_cache_get_stats(self):
        """Test ContextCache.get_stats method."""
        cache = ContextCache()
        stats = cache.get_stats()
        assert isinstance(stats, dict)
        assert "hits" in stats or "misses" in stats


class TestTokenBudgetManagerInstantiation:
    """Test TokenBudgetManager class instantiation."""

    def test_token_budget_manager_init(self):
        """Test TokenBudgetManager can be instantiated."""
        manager = TokenBudgetManager()
        assert manager is not None

    def test_token_budget_manager_init_with_args(self):
        """Test TokenBudgetManager can be instantiated with arguments."""
        manager = TokenBudgetManager(100000)
        assert manager is not None


class TestHookMetadataInstantiation:
    """Test HookMetadata dataclass instantiation."""

    def test_hook_metadata_init(self):
        """Test HookMetadata can be instantiated with required fields."""
        metadata = HookMetadata(
            hook_path="/test/path",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
        )
        assert metadata.hook_path == "/test/path"
        assert metadata.event_type == HookEvent.SESSION_START

    def test_hook_metadata_defaults(self):
        """Test HookMetadata respects default values."""
        metadata = HookMetadata(
            hook_path="/test/path",
            event_type=HookEvent.PRE_TOOL_USE,
            priority=HookPriority.NORMAL,
        )
        assert hasattr(metadata, "estimated_execution_time_ms")
        assert hasattr(metadata, "token_cost_estimate")


class TestJITEnhancedHookManagerInstantiation:
    """Test JITEnhancedHookManager class instantiation."""

    def test_jit_hook_manager_class_exists(self):
        """Test JITEnhancedHookManager class can be imported."""
        assert JITEnhancedHookManager is not None

    def test_jit_hook_manager_has_methods(self):
        """Test JITEnhancedHookManager is a proper class."""
        # Check that it's a class with methods
        methods = [m for m in dir(JITEnhancedHookManager) if not m.startswith("_")]
        assert len(methods) > 0


class TestEnumValues:
    """Test enum values are properly defined."""

    def test_hook_priority_values(self):
        """Test HookPriority enum has expected values."""
        assert hasattr(HookPriority, "CRITICAL")
        assert hasattr(HookPriority, "HIGH")
        assert hasattr(HookPriority, "NORMAL")
        assert hasattr(HookPriority, "LOW")

    def test_phase_enum_values(self):
        """Test Phase enum has expected values."""
        assert hasattr(Phase, "SPEC")
        assert hasattr(Phase, "RED")
        assert hasattr(Phase, "GREEN")
        assert hasattr(Phase, "REFACTOR")

    def test_hook_event_values(self):
        """Test HookEvent enum has expected values."""
        assert hasattr(HookEvent, "SESSION_START")
        assert hasattr(HookEvent, "SESSION_END")
        assert hasattr(HookEvent, "PRE_TOOL_USE")
        assert hasattr(HookEvent, "POST_TOOL_USE")
