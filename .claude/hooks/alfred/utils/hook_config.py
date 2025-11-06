#!/usr/bin/env python3
# Hook Configuration Utilities

"""
Hook 관련 설정 및 유틸리티 함수들
- Hook timeout 설정 로드
- Graceful degradation 설정 확인
"""

import json
import logging
from pathlib import Path

logger = logging.getLogger(__name__)


def load_hook_timeout() -> int:
    """
    .moai/config.json에서 Hook timeout 설정 로드

    Returns:
        int: timeout 값 (밀리초), 설정이 없으면 기본값 5000 반환
    """
    try:
        config_path = Path(".moai/config.json")
        if not config_path.exists():
            return 5000  # 기본값

        with open(config_path, 'r', encoding='utf-8') as f:
            config = json.load(f)

        # hooks 섹션에서 timeout_ms 값 가져오기
        hooks_config = config.get("hooks", {})
        timeout_ms = hooks_config.get("timeout_ms", 5000)

        return int(timeout_ms)
    except (json.JSONDecodeError, FileNotFoundError, KeyError, ValueError):
        logger.warning("Failed to load hook timeout from config, using default 5000ms")
        return 5000


def get_graceful_degradation() -> bool:
    """
    .moai/config.json에서 graceful_degradation 설정 로드

    Returns:
        bool: graceful_degradation 설정값, 설정이 없으면 기본값 True 반환
    """
    try:
        config_path = Path(".moai/config.json")
        if not config_path.exists():
            return True  # 기본값

        with open(config_path, 'r', encoding='utf-8') as f:
            config = json.load(f)

        # hooks 섹션에서 graceful_degradation 값 가져오기
        hooks_config = config.get("hooks", {})
        return hooks_config.get("graceful_degradation", True)
    except (json.JSONDecodeError, FileNotFoundError, KeyError):
        logger.warning("Failed to load graceful_degradation from config, using default True")
        return True


def get_hook_execution_config() -> dict:
    """
    Hook 실행 관련 모든 설정 로드

    Returns:
        dict: Hook 설정 딕셔너리
    """
    try:
        config_path = Path(".moai/config.json")
        if not config_path.exists():
            return {
                "timeout_ms": 5000,
                "graceful_degradation": True
            }

        with open(config_path, 'r', encoding='utf-8') as f:
            config = json.load(f)

        hooks_config = config.get("hooks", {})

        return {
            "timeout_ms": hooks_config.get("timeout_ms", 5000),
            "graceful_degradation": hooks_config.get("graceful_degradation", True)
        }
    except (json.JSONDecodeError, FileNotFoundError):
        logger.warning("Failed to load hook config, using defaults")
        return {
            "timeout_ms": 5000,
            "graceful_degradation": True
        }
