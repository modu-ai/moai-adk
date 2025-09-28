#!/usr/bin/env python3
# @TASK:COVERAGE-CHECK-011
"""
MoAI-ADK Test Coverage Checker v0.1.12 - Modularized Entry Point
테스트 커버리지 측정 및 임계값 검증

This script is now a thin wrapper around the modularized check_coverage package.
All core functionality has been moved to separate modules following TRUST principles.
"""

from .check_coverage import main


if __name__ == "__main__":
    main()