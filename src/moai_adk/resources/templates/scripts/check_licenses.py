#!/usr/bin/env python3
# @TASK:LICENSE-CHECK-011
"""
MoAI-ADK License Compliance Checker v0.1.12 - Modularized Entry Point
프로젝트 의존성 라이선스 검사 및 호환성 검증

This script is now a thin wrapper around the modularized check_licenses package.
All core functionality has been moved to separate modules following TRUST principles.
"""

from .check_licenses import main


if __name__ == "__main__":
    main()