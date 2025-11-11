#!/usr/bin/env python3
"""Test ConfigManager caching effectiveness"""

import time
import sys
from pathlib import Path

# Add module path
sys.path.insert(0, str(Path(__file__).parent))
sys.path.insert(0, str(Path(__file__).parent / ".claude" / "hooks" / "alfred"))

# Test 1: ConfigManager with SharedCache
print("=== ConfigManager 캐싱 테스트 ===\n")

try:
    from shared.core.config_manager import get_config_manager

    print("1. ConfigManager 로드 테스트 (10회)")
    start = time.time()
    cm = get_config_manager(Path(".moai/config.json"))

    for i in range(10):
        config = cm.load_config()

    elapsed = time.time() - start
    print(f"   10회 로드 시간: {elapsed:.3f}초")
    print(f"   평균 로드 시간: {elapsed/10:.3f}초\n")

except Exception as e:
    print(f"   ConfigManager 테스트 실패: {e}\n")

# Test 2: Direct JSON loading (no cache)
print("2. 직접 JSON 로드 테스트 (10회, 캐싱 없음)")
import json

start = time.time()
for i in range(10):
    with open(".moai/config.json", 'r') as f:
        config = json.load(f)

elapsed = time.time() - start
print(f"   10회 로드 시간: {elapsed:.3f}초")
print(f"   평균 로드 시간: {elapsed/10:.3f}초\n")

# Test 3: Module-level caching test
print("3. 모듈 레벨 캐싱 테스트 (pre_tool__tag_policy_validator.py)")
try:
    sys.path.insert(0, str(Path(__file__).parent / ".claude" / "hooks" / "alfred"))

    # Import hook module
    import pre_tool__tag_policy_validator as validator_hook

    # First call (should load from file)
    start = time.time()
    config1 = validator_hook.load_config()
    elapsed1 = time.time() - start

    # Second call (should use cached config)
    start = time.time()
    config2 = validator_hook.load_config()
    elapsed2 = time.time() - start

    print(f"   첫 번째 로드: {elapsed1:.3f}초")
    print(f"   두 번째 로드 (캐시): {elapsed2:.3f}초")
    print(f"   성능 향상: {elapsed1/elapsed2:.1f}배\n")

    # Test validator caching
    print("4. Validator 캐싱 테스트")
    start = time.time()
    v1 = validator_hook.create_policy_validator()
    elapsed1 = time.time() - start

    start = time.time()
    v2 = validator_hook.create_policy_validator()
    elapsed2 = time.time() - start

    print(f"   첫 번째 생성: {elapsed1:.3f}초")
    print(f"   두 번째 생성 (캐시): {elapsed2:.3f}초")
    print(f"   성능 향상: {elapsed1/elapsed2:.1f}배")
    print(f"   같은 객체: {v1 is v2}\n")

except Exception as e:
    print(f"   모듈 레벨 캐싱 테스트 실패: {e}\n")
    import traceback
    traceback.print_exc()

print("=== 테스트 완료 ===")
