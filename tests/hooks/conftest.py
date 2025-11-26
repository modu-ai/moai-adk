#!/usr/bin/env python3
"""Pytest configuration for hooks tests

이 파일은 hooks 테스트가 실행되기 전에 sys.path를 설정합니다.
pytest의 assertion rewriting 메커니즘이 완료되기 전에 모듈을 찾을 수 있도록 합니다.
"""

import sys
from pathlib import Path

# Hook 디렉토리를 sys.path에 추가
HOOKS_DIR = Path(__file__).parent.parent.parent / ".claude" / "hooks"
MOAI_LIB_DIR = HOOKS_DIR / "moai" / "lib"
SHARED_DIR = HOOKS_DIR / "moai" / "shared"
UTILS_DIR = HOOKS_DIR / "moai" / "utils"

# sys.path에 추가 (앞에 추가하여 우선순위 높임)
# 이 코드는 conftest.py가 import될 때 즉시 실행됨
for dir_path in [MOAI_LIB_DIR, SHARED_DIR, UTILS_DIR, HOOKS_DIR]:
    if str(dir_path) not in sys.path:
        sys.path.insert(0, str(dir_path))
