"""
Git Strategy Classes

개인/팀 모드에 따른 Git 작업 전략을 구현합니다.

@DESIGN:GIT-STRATEGY-001 - Strategy 패턴으로 Git 워크플로우 관리
@PERF:BRANCH-FAST - 브랜치 작업 최적화 (빠른 전환)
@SEC:GIT-MED - Git 작업 보안 강화

TRUST 원칙 적용:
- 기존 557 LOC → 모듈화로 분해
- 각 전략 클래스 별도 파일 분리
- 하위 호환성 유지를 위한 re-export
"""

# Backward compatibility imports
from .git_strategy.base import GitStrategyBase
from .git_strategy.personal_strategy import PersonalGitStrategy
from .git_strategy.team_strategy import TeamGitStrategy

__all__ = [
    "GitStrategyBase",
    "PersonalGitStrategy",
    "TeamGitStrategy",
]