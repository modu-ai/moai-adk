#!/usr/bin/env python3
# @TASK:TAG-VALIDATE-011
"""
MoAI-ADK Tag System Validator v0.1.12 - Modularized Entry Point
16-Core @TAG 무결성 검사 및 추적성 매트릭스 검증

This script is now a thin wrapper around the modularized validate_tags package.
All core functionality has been moved to separate modules following TRUST principles.

⚠️  NOTE: 이 스크립트는 SQLite 전용입니다. JSON 호환성은 완전히 제거되었습니다.
"""

from .validate_tags import main


if __name__ == "__main__":
    main()