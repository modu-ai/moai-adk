"""Model allocation based on pricing plan.

Assigns appropriate models (opus/sonnet/haiku) to agents based on
the user's selected service and pricing plan.
"""

from typing import TypedDict


class ModelAllocation(TypedDict):
    """Model allocation for different agent types."""

    # High-complexity agents (reasoning, architecture)
    plan: str
    expert_security: str
    expert_refactoring: str
    manager_strategy: str

    # Medium-complexity agents (implementation)
    expert_backend: str
    expert_frontend: str
    expert_database: str
    manager_tdd: str
    manager_spec: str
    manager_docs: str
    manager_quality: str

    # Low-complexity agents (exploration, quick tasks)
    explore: str
    expert_debug: str
    general_purpose: str


# Claude subscription model allocations by pricing plan
CLAUDE_SUBSCRIPTION_ALLOCATIONS: dict[str, ModelAllocation] = {
    # Pro ($20/mo) - sonnet-focused, limited opus
    "pro": {
        # High-complexity: sonnet (opus limited)
        "plan": "sonnet",
        "expert_security": "sonnet",
        "expert_refactoring": "sonnet",
        "manager_strategy": "sonnet",
        # Medium-complexity: sonnet
        "expert_backend": "sonnet",
        "expert_frontend": "sonnet",
        "expert_database": "sonnet",
        "manager_tdd": "sonnet",
        "manager_spec": "sonnet",
        "manager_docs": "sonnet",
        "manager_quality": "sonnet",
        # Low-complexity: haiku for speed
        "explore": "haiku",
        "expert_debug": "haiku",
        "general_purpose": "sonnet",
    },
    # Max5 ($100/mo) - opus for complex, sonnet for standard
    "max5": {
        # High-complexity: opus
        "plan": "opus",
        "expert_security": "opus",
        "expert_refactoring": "opus",
        "manager_strategy": "opus",
        # Medium-complexity: sonnet
        "expert_backend": "sonnet",
        "expert_frontend": "sonnet",
        "expert_database": "sonnet",
        "manager_tdd": "sonnet",
        "manager_spec": "sonnet",
        "manager_docs": "sonnet",
        "manager_quality": "sonnet",
        # Low-complexity: haiku for speed
        "explore": "haiku",
        "expert_debug": "haiku",
        "general_purpose": "sonnet",
    },
    # Max20 ($200/mo) - opus freely available
    "max20": {
        # High-complexity: opus
        "plan": "opus",
        "expert_security": "opus",
        "expert_refactoring": "opus",
        "manager_strategy": "opus",
        # Medium-complexity: opus or sonnet based on preference
        "expert_backend": "opus",
        "expert_frontend": "opus",
        "expert_database": "opus",
        "manager_tdd": "opus",
        "manager_spec": "opus",
        "manager_docs": "sonnet",  # Docs don't need opus
        "manager_quality": "opus",
        # Low-complexity: haiku for speed
        "explore": "haiku",
        "expert_debug": "sonnet",
        "general_purpose": "opus",
    },
}

# Claude API model allocations (same as subscription)
CLAUDE_API_ALLOCATIONS = CLAUDE_SUBSCRIPTION_ALLOCATIONS.copy()

# GLM CodePlan model allocations
GLM_ALLOCATIONS: dict[str, ModelAllocation] = {
    # Basic - equivalent to sonnet
    "basic": {
        "plan": "glm-basic",
        "expert_security": "glm-basic",
        "expert_refactoring": "glm-basic",
        "manager_strategy": "glm-basic",
        "expert_backend": "glm-basic",
        "expert_frontend": "glm-basic",
        "expert_database": "glm-basic",
        "manager_tdd": "glm-basic",
        "manager_spec": "glm-basic",
        "manager_docs": "glm-basic",
        "manager_quality": "glm-basic",
        "explore": "glm-basic",
        "expert_debug": "glm-basic",
        "general_purpose": "glm-basic",
    },
    # Pro - equivalent to sonnet~opus
    "glm_pro": {
        "plan": "glm-pro",
        "expert_security": "glm-pro",
        "expert_refactoring": "glm-pro",
        "manager_strategy": "glm-pro",
        "expert_backend": "glm-pro",
        "expert_frontend": "glm-pro",
        "expert_database": "glm-pro",
        "manager_tdd": "glm-pro",
        "manager_spec": "glm-pro",
        "manager_docs": "glm-basic",
        "manager_quality": "glm-pro",
        "explore": "glm-basic",
        "expert_debug": "glm-basic",
        "general_purpose": "glm-pro",
    },
    # Enterprise - equivalent to opus
    "enterprise": {
        "plan": "glm-enterprise",
        "expert_security": "glm-enterprise",
        "expert_refactoring": "glm-enterprise",
        "manager_strategy": "glm-enterprise",
        "expert_backend": "glm-enterprise",
        "expert_frontend": "glm-enterprise",
        "expert_database": "glm-enterprise",
        "manager_tdd": "glm-enterprise",
        "manager_spec": "glm-enterprise",
        "manager_docs": "glm-pro",
        "manager_quality": "glm-enterprise",
        "explore": "glm-basic",
        "expert_debug": "glm-pro",
        "general_purpose": "glm-enterprise",
    },
}


def get_model_allocation(
    service_type: str,
    pricing_plan: str | None = None,
) -> ModelAllocation:
    """Get model allocation based on service type and pricing plan.

    Args:
        service_type: Service type (claude_subscription, claude_api, glm, hybrid)
        pricing_plan: Pricing plan (pro, max5, max20, basic, glm_pro, enterprise)

    Returns:
        ModelAllocation dictionary mapping agent types to model names
    """
    if service_type == "claude_subscription":
        plan = pricing_plan or "pro"
        return CLAUDE_SUBSCRIPTION_ALLOCATIONS.get(plan, CLAUDE_SUBSCRIPTION_ALLOCATIONS["pro"])

    if service_type == "claude_api":
        plan = pricing_plan or "pro"
        return CLAUDE_API_ALLOCATIONS.get(plan, CLAUDE_API_ALLOCATIONS["pro"])

    if service_type == "glm":
        plan = pricing_plan or "basic"
        return GLM_ALLOCATIONS.get(plan, GLM_ALLOCATIONS["basic"])

    if service_type == "hybrid":
        # Hybrid uses Claude allocation with GLM fallback for simple tasks
        plan = pricing_plan or "pro"
        base = CLAUDE_SUBSCRIPTION_ALLOCATIONS.get(plan, CLAUDE_SUBSCRIPTION_ALLOCATIONS["pro"])
        # Override exploration agents to use GLM for cost optimization
        return {
            **base,
            "explore": "glm-basic",
            "expert_debug": "glm-basic",
        }

    # Default fallback
    return CLAUDE_SUBSCRIPTION_ALLOCATIONS["pro"]


def get_agent_model(
    agent_name: str,
    service_type: str,
    pricing_plan: str | None = None,
) -> str:
    """Get the model for a specific agent.

    Args:
        agent_name: Agent name (e.g., "Plan", "expert-backend", "Explore")
        service_type: Service type
        pricing_plan: Pricing plan

    Returns:
        Model name (opus, sonnet, haiku, or glm-*)
    """
    allocation = get_model_allocation(service_type, pricing_plan)

    # Normalize agent name to key format
    key = agent_name.lower().replace("-", "_")

    # Try direct lookup
    value = allocation.get(key)
    if isinstance(value, str):
        return value

    # Try category-based lookup
    if key.startswith("expert_"):
        return allocation.get("expert_backend", "sonnet")
    if key.startswith("manager_"):
        return allocation.get("manager_tdd", "sonnet")
    if key.startswith("builder_"):
        return allocation.get("general_purpose", "sonnet")

    # Default to sonnet
    return "sonnet"


__all__ = [
    "ModelAllocation",
    "get_model_allocation",
    "get_agent_model",
    "CLAUDE_SUBSCRIPTION_ALLOCATIONS",
    "CLAUDE_API_ALLOCATIONS",
    "GLM_ALLOCATIONS",
]
