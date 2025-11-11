#!/usr/bin/env python3
"""Hook performance test script"""

import subprocess
import time
from pathlib import Path

def test_hook_performance():
    """Test hook execution performance"""

    # Test hooks
    hooks = [
        ".claude/hooks/alfred/pre_tool__tag_policy_validator.py",
        ".claude/hooks/alfred/pre_tool__auto_checkpoint.py",
    ]

    print("=== Hook Performance Test ===\n")

    for hook_path in hooks:
        hook_file = Path(hook_path)
        if not hook_file.exists():
            print(f"❌ {hook_path}: NOT FOUND")
            continue

        # Run hook 5 times and measure average
        times = []
        for i in range(5):
            start = time.time()
            result = subprocess.run(
                ["uv", "run", "python3", str(hook_file)],
                capture_output=True,
                text=True,
                timeout=10
            )
            elapsed = time.time() - start
            times.append(elapsed)

        avg_time = sum(times) / len(times)
        min_time = min(times)
        max_time = max(times)

        print(f"📊 {hook_file.name}")
        print(f"   평균: {avg_time:.3f}초")
        print(f"   최소: {min_time:.3f}초")
        print(f"   최대: {max_time:.3f}초")
        print()

if __name__ == "__main__":
    test_hook_performance()
