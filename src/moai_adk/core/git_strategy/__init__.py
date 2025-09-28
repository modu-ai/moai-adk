"""
Git Strategy Module

개인/팀 모드에 따른 Git 작업 전략을 구현하는 모듈입니다.

@DESIGN:GIT-STRATEGY-MODULE-001 - 모듈화된 Git 전략 패키지
@TRUST:MODULAR - 각 전략이 독립적인 파일로 분리됨

Usage:
    from moai_adk.core.git_strategy import GitStrategyBase, PersonalGitStrategy, TeamGitStrategy
"""

from .base import GitStrategyBase
from .personal_strategy import PersonalGitStrategy
from .team_strategy import TeamGitStrategy

__all__ = [
    "GitStrategyBase",
    "PersonalGitStrategy",
    "TeamGitStrategy",
]