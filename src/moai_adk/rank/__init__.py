"""MoAI Rank CLI Module

This module provides CLI commands for interacting with the MoAI Rank service
at rank.mo.ai.kr. It includes:

- API client with HMAC-SHA256 request signing
- OAuth flow for GitHub authentication
- Secure credential storage
- Session hook for automatic token usage submission

Commands:
- moai rank register: Register via GitHub OAuth
- moai rank status: Show user's current rank and stats
- moai rank leaderboard: Display the leaderboard
"""

from moai_adk.rank.auth import OAuthHandler
from moai_adk.rank.client import RankClient
from moai_adk.rank.config import RankConfig

__all__ = ["RankClient", "RankConfig", "OAuthHandler"]
