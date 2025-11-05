# @TEST:AUTH-004

# SPEC:@SPEC:AUTH-004

import unittest
from src.auth.example import example_function


class TestAuth(unittest.TestCase):
    """JWT 인증 시스템 테스트"""

    def test_example_function(self):
        """예제 함수 테스트"""
        result = example_function()
        self.assertIsNone(result)  # 예제 함수는 아무것도 반환하지 않음


if __name__ == '__main__':
    unittest.main()