"""Comprehensive tests for model_allocator module.

This module tests model allocation based on pricing plans and service types.
"""

from moai_adk.core.model_allocator import (
    CLAUDE_API_ALLOCATIONS,
    CLAUDE_SUBSCRIPTION_ALLOCATIONS,
    GLM_ALLOCATIONS,
    ModelAllocation,
    get_agent_model,
    get_model_allocation,
)


class TestGetModelAllocation:
    """Test get_model_allocation function."""

    def test_claude_subscription_pro_plan(self) -> None:
        """Test Claude subscription with Pro plan."""
        result = get_model_allocation("claude_subscription", "pro")

        assert result["plan"] == "sonnet"
        assert result["expert_security"] == "sonnet"
        assert result["expert_backend"] == "sonnet"
        assert result["explore"] == "haiku"

    def test_claude_subscription_max5_plan(self) -> None:
        """Test Claude subscription with Max5 plan."""
        result = get_model_allocation("claude_subscription", "max5")

        assert result["plan"] == "opus"
        assert result["expert_security"] == "opus"
        assert result["expert_backend"] == "sonnet"
        assert result["explore"] == "haiku"

    def test_claude_subscription_max20_plan(self) -> None:
        """Test Claude subscription with Max20 plan."""
        result = get_model_allocation("claude_subscription", "max20")

        assert result["plan"] == "opus"
        assert result["expert_security"] == "opus"
        assert result["expert_backend"] == "opus"
        assert result["manager_docs"] == "sonnet"
        assert result["explore"] == "haiku"

    def test_claude_subscription_default_plan(self) -> None:
        """Test Claude subscription with default plan (None)."""
        result = get_model_allocation("claude_subscription", None)

        # Should default to pro plan
        assert result["plan"] == "sonnet"

    def test_claude_subscription_invalid_plan(self) -> None:
        """Test Claude subscription with invalid plan falls back to pro."""
        result = get_model_allocation("claude_subscription", "invalid_plan")

        # Should fall back to pro plan
        assert result["plan"] == "sonnet"

    def test_claude_api_pro_plan(self) -> None:
        """Test Claude API with Pro plan."""
        result = get_model_allocation("claude_api", "pro")

        assert result["plan"] == "sonnet"
        assert result["explore"] == "haiku"

    def test_claude_api_default_plan(self) -> None:
        """Test Claude API with default plan."""
        result = get_model_allocation("claude_api", None)

        assert result["plan"] == "sonnet"

    def test_glm_basic_plan(self) -> None:
        """Test GLM with basic plan."""
        result = get_model_allocation("glm", "basic")

        assert result["plan"] == "glm-basic"
        assert result["expert_backend"] == "glm-basic"
        assert result["explore"] == "glm-basic"

    def test_glm_pro_plan(self) -> None:
        """Test GLM with pro plan."""
        result = get_model_allocation("glm", "glm_pro")

        assert result["plan"] == "glm-pro"
        assert result["expert_backend"] == "glm-pro"
        assert result["manager_docs"] == "glm-basic"
        assert result["explore"] == "glm-basic"

    def test_glm_enterprise_plan(self) -> None:
        """Test GLM with enterprise plan."""
        result = get_model_allocation("glm", "enterprise")

        assert result["plan"] == "glm-enterprise"
        assert result["expert_backend"] == "glm-enterprise"
        assert result["manager_docs"] == "glm-pro"
        assert result["explore"] == "glm-basic"
        assert result["expert_debug"] == "glm-pro"

    def test_glm_default_plan(self) -> None:
        """Test GLM with default plan."""
        result = get_model_allocation("glm", None)

        assert result["plan"] == "glm-basic"

    def test_hybrid_service_type(self) -> None:
        """Test hybrid service type."""
        result = get_model_allocation("hybrid", "pro")

        assert result["plan"] == "sonnet"
        # Hybrid should override explore and expert_debug to use GLM
        assert result["explore"] == "glm-basic"
        assert result["expert_debug"] == "glm-basic"

    def test_hybrid_default_plan(self) -> None:
        """Test hybrid with default plan."""
        result = get_model_allocation("hybrid", None)

        assert result["plan"] == "sonnet"
        assert result["explore"] == "glm-basic"

    def test_unknown_service_type_fallback(self) -> None:
        """Test unknown service type falls back to Claude subscription pro."""
        result = get_model_allocation("unknown_service", None)

        # Should fall back to Claude subscription pro plan
        assert result["plan"] == "sonnet"
        assert result["explore"] == "haiku"


class TestGetAgentModel:
    """Test get_agent_model function."""

    def test_plan_agent_claude_subscription_pro(self) -> None:
        """Test Plan agent with Claude subscription Pro plan."""
        model = get_agent_model("plan", "claude_subscription", "pro")

        assert model == "sonnet"

    def test_explore_agent_claude_subscription(self) -> None:
        """Test explore agent with Claude subscription."""
        model = get_agent_model("explore", "claude_subscription", "pro")

        assert model == "haiku"

    def test_expert_security_agent_max5(self) -> None:
        """Test expert_security agent with Max5 plan."""
        model = get_agent_model("expert_security", "claude_subscription", "max5")

        assert model == "opus"

    def test_expert_backend_agent_max20(self) -> None:
        """Test expert_backend agent with Max20 plan."""
        model = get_agent_model("expert_backend", "claude_subscription", "max20")

        assert model == "opus"

    def test_manager_docs_agent_max20(self) -> None:
        """Test manager_docs agent with Max20 plan."""
        model = get_agent_model("manager_docs", "claude_subscription", "max20")

        assert model == "sonnet"

    def test_agent_name_with_hyphen(self) -> None:
        """Test agent name with hyphen is normalized."""
        model = get_agent_model("expert-backend", "claude_subscription", "pro")

        assert model == "sonnet"

    def test_agent_name_with_mixed_case(self) -> None:
        """Test agent name with mixed case is normalized."""
        model = get_agent_model("Expert-Backend", "claude_subscription", "pro")

        assert model == "sonnet"

    def test_glm_service_basic_plan(self) -> None:
        """Test GLM service with basic plan."""
        model = get_agent_model("plan", "glm", "basic")

        assert model == "glm-basic"

    def test_glm_service_enterprise_plan(self) -> None:
        """Test GLM service with enterprise plan."""
        model = get_agent_model("expert_backend", "glm", "enterprise")

        assert model == "glm-enterprise"

    def test_hybrid_service_uses_glm_for_explore(self) -> None:
        """Test hybrid service uses GLM for explore agent."""
        model = get_agent_model("explore", "hybrid", "pro")

        assert model == "glm-basic"

    def test_hybrid_service_uses_glm_for_debug(self) -> None:
        """Test hybrid service uses GLM for expert_debug agent."""
        model = get_agent_model("expert_debug", "hybrid", "pro")

        assert model == "glm-basic"

    def test_unknown_agent_category_expert(self) -> None:
        """Test unknown agent starting with expert_ gets fallback."""
        model = get_agent_model("expert_unknown", "claude_subscription", "pro")

        assert model == "sonnet"

    def test_unknown_agent_category_manager(self) -> None:
        """Test unknown agent starting with manager_ gets fallback."""
        model = get_agent_model("manager_unknown", "claude_subscription", "pro")

        assert model == "sonnet"

    def test_unknown_agent_category_builder(self) -> None:
        """Test unknown agent starting with builder_ gets fallback."""
        model = get_agent_model("builder_unknown", "claude_subscription", "pro")

        assert model == "sonnet"

    def test_unknown_agent_default_sonnet(self) -> None:
        """Test completely unknown agent defaults to sonnet."""
        model = get_agent_model("unknown_agent", "claude_subscription", "pro")

        assert model == "sonnet"

    def test_default_pricing_plan_none(self) -> None:
        """Test with None pricing plan."""
        model = get_agent_model("plan", "claude_subscription", None)

        assert model == "sonnet"


class TestModelAllocationConstants:
    """Test model allocation constants."""

    def test_claude_subscription_pro_has_all_keys(self) -> None:
        """Test Pro allocation has all required keys."""
        pro = CLAUDE_SUBSCRIPTION_ALLOCATIONS["pro"]

        required_keys = [
            "plan",
            "expert_security",
            "expert_refactoring",
            "manager_strategy",
            "expert_backend",
            "expert_frontend",
            "expert_database",
            "manager_ddd",
            "manager_spec",
            "manager_docs",
            "manager_quality",
            "explore",
            "expert_debug",
            "general_purpose",
        ]

        for key in required_keys:
            assert key in pro

    def test_claude_api_is_copy_of_subscription(self) -> None:
        """Test CLAUDE_API_ALLOCATIONS is a copy of subscription allocations."""
        # They should have the same keys
        assert set(CLAUDE_API_ALLOCATIONS.keys()) == set(CLAUDE_SUBSCRIPTION_ALLOCATIONS.keys())

    def test_glm_allocations_has_basic_plan(self) -> None:
        """Test GLM allocations has basic plan."""
        assert "basic" in GLM_ALLOCATIONS

        basic = GLM_ALLOCATIONS["basic"]
        assert basic["plan"] == "glm-basic"

    def test_glm_allocations_has_pro_plan(self) -> None:
        """Test GLM allocations has pro plan."""
        assert "glm_pro" in GLM_ALLOCATIONS

        pro = GLM_ALLOCATIONS["glm_pro"]
        assert pro["plan"] == "glm-pro"

    def test_glm_allocations_has_enterprise_plan(self) -> None:
        """Test GLM allocations has enterprise plan."""
        assert "enterprise" in GLM_ALLOCATIONS

        enterprise = GLM_ALLOCATIONS["enterprise"]
        assert enterprise["plan"] == "glm-enterprise"

    def test_max5_uses_opus_for_complex_agents(self) -> None:
        """Test Max5 uses opus for complex agents."""
        max5 = CLAUDE_SUBSCRIPTION_ALLOCATIONS["max5"]

        assert max5["plan"] == "opus"
        assert max5["expert_security"] == "opus"
        assert max5["manager_strategy"] == "opus"

    def test_max20_uses_opus_more_widely(self) -> None:
        """Test Max20 uses opus more widely than Max5."""
        max20 = CLAUDE_SUBSCRIPTION_ALLOCATIONS["max20"]

        # More agents use opus in max20
        assert max20["expert_backend"] == "opus"
        assert max20["expert_frontend"] == "opus"
        assert max20["expert_database"] == "opus"
        assert max20["manager_ddd"] == "opus"
        assert max20["manager_spec"] == "opus"
        assert max20["manager_quality"] == "opus"


class TestModelAllocationEdgeCases:
    """Test edge cases for model allocation."""

    def test_empty_plan_name(self) -> None:
        """Test with empty plan name."""
        result = get_model_allocation("claude_subscription", "")

        # Empty string should trigger fallback
        assert result["plan"] == "sonnet"

    def test_case_sensitivity_plan_name(self) -> None:
        """Test plan name is case-sensitive."""
        # "Pro" vs "pro" - different keys
        result = get_model_allocation("claude_subscription", "Pro")

        # Should fall back to pro (lowercase)
        assert result["plan"] == "sonnet"

    def test_all_agent_types_have_allocation(self) -> None:
        """Test all agent types defined in ModelAllocation are allocated."""
        allocation = get_model_allocation("claude_subscription", "pro")

        # Check that every key in ModelAllocation has a value
        for key in ModelAllocation.__annotations__:
            assert key in allocation
            assert isinstance(allocation[key], str)
            assert len(allocation[key]) > 0

    def test_hybrid_preserves_claudel_allocation_for_most(self) -> None:
        """Test hybrid service preserves Claude allocation for most agents."""
        hybrid = get_model_allocation("hybrid", "pro")
        claude = get_model_allocation("claude_subscription", "pro")

        # Most agents should be the same
        assert hybrid["plan"] == claude["plan"]
        assert hybrid["expert_backend"] == claude["expert_backend"]
        assert hybrid["manager_ddd"] == claude["manager_ddd"]

        # Except explore and expert_debug which are overridden
        assert hybrid["explore"] != claude["explore"]
        assert hybrid["expert_debug"] != claude["expert_debug"]
