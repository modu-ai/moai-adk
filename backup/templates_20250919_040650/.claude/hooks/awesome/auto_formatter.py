#!/usr/bin/env python3
"""
Smart code formatter hook for Claude Code
Automatically formats code based on file type using appropriate formatters
"""

import os
import sys
import subprocess
from pathlib import Path

def format_file(file_path):
    """Format file based on its extension"""
    path = Path(file_path)
    ext = path.suffix.lower()

    formatters = {
        # JavaScript/TypeScript/Web
        ('.js', '.jsx', '.ts', '.tsx', '.json', '.css', '.html'):
            ['npx', 'prettier', '--write', file_path],

        # Python
        ('.py',):
            ['black', file_path],

        # Go
        ('.go',):
            ['gofmt', '-w', file_path],

        # Rust
        ('.rs',):
            ['rustfmt', file_path],

        # PHP
        ('.php',):
            ['php-cs-fixer', 'fix', file_path],
    }

    for extensions, command in formatters.items():
        if ext in extensions:
            try:
                result = subprocess.run(
                    command,
                    capture_output=True,
                    text=True,
                    timeout=10
                )
                if result.returncode == 0:
                    print(f"✅ Formatted {path.name} with {command[0]}")
                return result.returncode
            except FileNotFoundError:
                print(f"⚠️ {command[0]} not found, skipping formatting", file=sys.stderr)
                return 0
            except subprocess.TimeoutExpired:
                print(f"⚠️ Formatter timeout for {path.name}", file=sys.stderr)
                return 0
            except Exception as e:
                print(f"⚠️ Error formatting {path.name}: {e}", file=sys.stderr)
                return 0

    return 0

def main():
    file_path = os.environ.get('CLAUDE_TOOL_FILE_PATH')

    if not file_path:
        # Skip silently if no file path provided (e.g., MultiEdit operations)
        return 0

    if not os.path.exists(file_path):
        print(f"File not found: {file_path}", file=sys.stderr)
        return 1

    return format_file(file_path)

if __name__ == "__main__":
    sys.exit(main())