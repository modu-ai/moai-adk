#!/usr/bin/env python3
"""
BaaS Provider Selection Tool

AI-powered provider selection based on project requirements.
Integrates with Context7 for latest provider documentation.
"""

import asyncio
import re
import sys
from typing import Any, Dict


class BaaSProviderSelector:
    """AI-powered BaaS provider selection engine."""

    def __init__(self):
        self.auth_providers = {
            "auth0": {
                "strengths": ["enterprise_sso", "b2b_saas", "50_connections"],
                "cost_tier": "enterprise",
                "use_cases": ["enterprise", "b2b", "large_teams"],
            },
            "clerk": {
                "strengths": ["webauthn", "modern_ui", "organizations"],
                "cost_tier": "mid_tier",
                "use_cases": ["modern_apps", "startups", "user_experience"],
            },
            "firebase_auth": {
                "strengths": ["google_integration", "mobile_first", "analytics"],
                "cost_tier": "pay_as_you_go",
                "use_cases": ["mobile_apps", "google_ecosystem", "startups"],
            },
        }

        self.database_providers = {
            "supabase": {
                "strengths": ["postgresql_16", "realtime", "rls"],
                "cost_tier": "mid_tier",
                "use_cases": ["complex_apps", "multi_tenant", "realtime_features"],
            },
            "neon": {
                "strengths": ["serverless", "auto_scaling", "branching"],
                "cost_tier": "pay_as_you_go",
                "use_cases": ["serverless_apps", "development", "cost_optimization"],
            },
            "convex": {
                "strengths": ["realtime_backend", "reactive", "optimistic_updates"],
                "cost_tier": "mid_tier",
                "use_cases": ["collaborative_apps", "realtime_features", "typescript_apps"],
            },
            "firestore": {
                "strengths": ["mobile_sync", "offline_support", "google_integration"],
                "cost_tier": "pay_as_you_go",
                "use_cases": ["mobile_apps", "offline_first", "google_ecosystem"],
            },
        }

        self.deployment_providers = {
            "vercel": {
                "strengths": ["edge_deployment", "nextjs_optimization", "global_cdn"],
                "cost_tier": "mid_tier",
                "use_cases": ["jamstack", "nextjs_apps", "global_performance"],
            },
            "railway": {
                "strengths": ["containers", "full_stack", "multi_region"],
                "cost_tier": "mid_tier",
                "use_cases": ["containers", "complex_backends", "multi_region"],
            },
        }

    def parse_budget(self, budget_str: str) -> int:
        """Parse budget string and return numeric value."""
        if not budget_str:
            return 1000  # Default budget

        # Extract numbers from budget string
        match = re.search(r"\$?(\d+)", budget_str)
        if match:
            return int(match.group(1))
        return 1000  # Default if no number found

    def analyze_requirements(self, requirements: Dict[str, Any]) -> Dict[str, float]:
        """Analyze project requirements and return weighted scores."""

        scores = {
            "enterprise_needs": 0.0,
            "realtime_required": 0.0,
            "mobile_first": 0.0,
            "cost_sensitivity": 0.0,
            "performance_critical": 0.0,
            "multi_tenant": 0.0,
            "development_speed": 0.0,
            "global_users": 0.0,
        }

        # Analyze requirements
        if "enterprise" in requirements.get("requirements", []):
            scores["enterprise_needs"] = 1.0

        if "real-time" in requirements.get("requirements", []):
            scores["realtime_required"] = 1.0

        if "mobile" in requirements.get("tech_stack", "").lower():
            scores["mobile_first"] = 0.8

        # Parse budget for cost sensitivity
        budget = self.parse_budget(requirements.get("budget", ""))
        if budget < 300:
            scores["cost_sensitivity"] = 1.0
        elif budget < 500:
            scores["cost_sensitivity"] = 0.5

        if "performance" in requirements.get("requirements", []):
            scores["performance_critical"] = 0.8

        if "scalable" in requirements.get("requirements", []):
            scores["multi_tenant"] = 0.6

        if requirements.get("team_size", 0) < 5:
            scores["development_speed"] = 0.8

        if "global" in requirements.get("requirements", []):
            scores["global_users"] = 1.0

        return scores

    def score_auth_provider(self, provider: str, scores: Dict[str, float]) -> float:
        """Score authentication provider against requirements."""
        score = 0.0

        # Enterprise needs scoring
        if scores["enterprise_needs"] > 0.5:
            if provider == "auth0":
                score += 3.0
            elif provider == "clerk":
                score += 1.0
            else:
                score += 0.5

        # Development speed scoring
        if scores["development_speed"] > 0.5:
            if provider == "clerk":
                score += 2.0
            elif provider == "firebase_auth":
                score += 1.5
            else:
                score += 1.0

        # Mobile first scoring
        if scores["mobile_first"] > 0.5:
            if provider == "firebase_auth":
                score += 2.5
            elif provider == "clerk":
                score += 1.5
            else:
                score += 1.0

        # Cost sensitivity scoring
        if scores["cost_sensitivity"] > 0.5:
            if provider in ["clerk", "firebase_auth"]:
                score += 1.5
            else:
                score += 0.5

        return score

    def score_database_provider(self, provider: str, scores: Dict[str, float]) -> float:
        """Score database provider against requirements."""
        score = 0.0

        # Real-time requirements scoring
        if scores["realtime_required"] > 0.5:
            if provider in ["supabase", "convex"]:
                score += 3.0
            elif provider == "firestore":
                score += 2.0
            else:
                score += 1.0

        # Cost sensitivity scoring
        if scores["cost_sensitivity"] > 0.5:
            if provider in ["neon", "firestore"]:
                score += 2.0
            else:
                score += 1.0

        # Multi-tenant scoring
        if scores["multi_tenant"] > 0.5:
            if provider == "supabase":
                score += 3.0
            elif provider in ["neon", "convex"]:
                score += 2.0
            else:
                score += 1.0

        # Development speed scoring
        if scores["development_speed"] > 0.5:
            if provider == "supabase":
                score += 2.0
            elif provider == "convex":
                score += 1.5
            else:
                score += 1.0

        return score

    def score_deployment_provider(self, provider: str, scores: Dict[str, float]) -> float:
        """Score deployment provider against requirements."""
        score = 0.0

        # Performance critical scoring
        if scores["performance_critical"] > 0.5:
            if provider == "vercel":
                score += 3.0
            else:
                score += 2.0

        # Global users scoring
        if scores["global_users"] > 0.5:
            if provider == "vercel":
                score += 3.0
            else:
                score += 2.0

        # Development speed scoring
        if scores["development_speed"] > 0.5:
            if provider == "vercel":
                score += 2.0
            else:
                score += 1.5

        # Complex backend scoring
        if scores["multi_tenant"] > 0.5:
            if provider == "railway":
                score += 2.5
            else:
                score += 1.5

        return score

    def select_providers(self, requirements: Dict[str, Any]) -> Dict[str, str]:
        """Select optimal providers based on requirements."""

        scores = self.analyze_requirements(requirements)

        # Score all providers
        auth_scores = {provider: self.score_auth_provider(provider, scores) for provider in self.auth_providers}

        db_scores = {provider: self.score_database_provider(provider, scores) for provider in self.database_providers}

        deploy_scores = {
            provider: self.score_deployment_provider(provider, scores) for provider in self.deployment_providers
        }

        # Select top providers
        recommendation = {
            "authentication": max(auth_scores, key=auth_scores.get),
            "database": max(db_scores, key=db_scores.get),
            "deployment": max(deploy_scores, key=deploy_scores.get),
        }

        return {
            "recommendation": recommendation,
            "scores": {"authentication": auth_scores, "database": db_scores, "deployment": deploy_scores},
            "requirement_analysis": scores,
        }


async def main():
    """Main function to run provider selection."""

    if len(sys.argv) < 2:
        print("Usage: python provider-selector.py <json_requirements>")
        print(
            'Example: python provider-selector.py \'{"requirements": ["scalable", "real-time"], "tech_stack": "Next.js", "team_size": 5, "budget": "$500/month"}\''
        )
        sys.exit(1)

    try:
        import json

        requirements = json.loads(sys.argv[1])

        selector = BaaSProviderSelector()
        result = selector.select_providers(requirements)

        print("=== BaaS Provider Recommendation ===")
        print(f"Authentication: {result['recommendation']['authentication']}")
        print(f"Database: {result['recommendation']['database']}")
        print(f"Deployment: {result['recommendation']['deployment']}")

        print("\n=== Detailed Scores ===")
        print("Authentication Scores:")
        for provider, score in result["scores"]["authentication"].items():
            print(f"  {provider}: {score:.1f}")

        print("\nDatabase Scores:")
        for provider, score in result["scores"]["database"].items():
            print(f"  {provider}: {score:.1f}")

        print("\nDeployment Scores:")
        for provider, score in result["scores"]["deployment"].items():
            print(f"  {provider}: {score:.1f}")

    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    asyncio.run(main())
