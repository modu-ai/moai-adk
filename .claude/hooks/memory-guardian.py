"""
Memory Guardian Hook - Prevents memory leaks at session start
Automatically runs memory cleanup when Claude sessions start
"""

import os
import subprocess
import sys
from pathlib import Path

def main():
    """Run memory cleanup at session start"""
    try:
        # Run MCP memory manager cleanup
        script_path = Path(__file__).parent.parent / "scripts" / "mcp-memory-manager.py"
        if script_path.exists():
            result = subprocess.run([sys.executable, str(script_path), "cleanup"], 
                                  capture_output=True, text=True)
            if result.stdout:
                print(f"🧹 Memory Guardian: {result.stdout}")
    except Exception as e:
        print(f"⚠️ Memory Guardian error: {e}")

if __name__ == "__main__":
    main()
