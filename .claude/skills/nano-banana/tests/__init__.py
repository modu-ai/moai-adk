"""
Nano Banana Pro Skill 테스트 모듈

모든 테스트는 pytest 형식으로 작성되며,
! uv run -m pytest 명령으로 실행됩니다.
"""

import sys
from pathlib import Path

# 부모 디렉토리 추가 (modules import)
sys.path.insert(0, str(Path(__file__).parent.parent))
