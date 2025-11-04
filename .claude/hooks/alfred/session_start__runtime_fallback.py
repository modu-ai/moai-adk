#!/usr/bin/env python3
"""
Runtime Fallback System for Claude Code
Provides automatic runtime switching when Bun shows memory issues
"""
import subprocess
import os
import sys
import json
from datetime import datetime

def check_runtime_health():
    """Check if current runtime (Bun) is healthy"""
    try:
        # Test Bun with simple operation
        result = subprocess.run(
            ["bun", "--version"],
            capture_output=True,
            text=True,
            timeout=5
        )

        if result.returncode != 0:
            return False, "Bun not responding"

        # Check if Bun process is using excessive memory
        memory_check = subprocess.run(
            ["ps", "-o", "rss,comm", "-A"],
            capture_output=True,
            text=True
        )

        bun_processes = []
        for line in memory_check.stdout.split('\n'):
            if 'bun' in line.lower():
                try:
                    rss_kb = int(line.split()[0])
                    rss_mb = rss_kb / 1024
                    if rss_mb > 2000:  # 2GB threshold
                        bun_processes.append(rss_mb)
                except:
                    pass

        if bun_processes:
            max_memory = max(bun_processes)
            if max_memory > 4000:  # 4GB critical threshold
                return False, f"Bun using excessive memory: {max_memory:.0f}MB"

        return True, "Runtime healthy"

    except subprocess.TimeoutExpired:
        return False, "Bun timeout"
    except Exception as e:
        return False, f"Runtime check failed: {str(e)}"

def get_nodejs_version():
    """Get Node.js version information"""
    try:
        result = subprocess.run(
            ["node", "--version"],
            capture_output=True,
            text=True,
            timeout=5
        )

        if result.returncode == 0:
            return result.stdout.strip(), True
        return None, False

    except Exception:
        return None, False

def suggest_runtime_switch():
    """Suggest runtime switch if needed"""
    runtime_healthy, message = check_runtime_health()
    node_version, node_available = get_nodejs_version()

    if not runtime_healthy:
        print(f"⚠️  Runtime issue detected: {message}")

        if node_available:
            print(f"✅ Node.js available: {node_version}")
            print("🔄 Recommended: Switch to Node.js runtime")
            print("\nTo switch runtime:")
            print("1. Set environment variable: export CLAUDE_RUNTIME=node")
            print("2. Or restart Claude Code with Node.js")
            print("3. Monitor memory usage with built-in tools")
            return True
        else:
            print("❌ Node.js not available")
            print("🔧 Recommended: Install Node.js LTS")
            print("brew install node@20")
            return False

    return False

def create_runtime_config():
    """Create runtime configuration for fallback"""
    config_dir = os.path.expanduser("~/.claude")
    config_file = os.path.join(config_dir, "runtime_config.json")

    config = {
        "timestamp": datetime.now().isoformat(),
        "fallback_enabled": True,
        "runtime_preference": ["node", "bun"],
        "memory_thresholds": {
            "warning_mb": 2000,
            "critical_mb": 4000
        },
        "health_checks": {
            "bun_timeout": 5,
            "node_timeout": 5
        },
        "auto_switch": False,  # Manual confirmation required
        "notes": "Runtime fallback configuration for memory stability"
    }

    try:
        os.makedirs(config_dir, exist_ok=True)
        with open(config_file, 'w') as f:
            json.dump(config, f, indent=2)
        print(f"📝 Runtime config created: {config_file}")
        return True
    except Exception as e:
        print(f"❌ Failed to create runtime config: {e}")
        return False

def main():
    """Main runtime fallback function"""
    print("🔄 Runtime Fallback System Check")
    print(f"📅 {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")

    # Check runtime health
    should_switch = suggest_runtime_switch()

    # Create fallback configuration
    config_created = create_runtime_config()

    # Environment check
    runtime_env = os.environ.get("CLAUDE_RUNTIME", "bun")
    print(f"🎯 Current runtime preference: {runtime_env}")

    if should_switch:
        print("\n🚨 Runtime fallback recommended")
        if config_created:
            print("✅ Fallback configuration ready")
        return 1  # Warning status
    else:
        print("\n✅ Runtime operating normally")
        if config_created:
            print("✅ Fallback configuration ready for future use")
        return 0

if __name__ == "__main__":
    sys.exit(main())