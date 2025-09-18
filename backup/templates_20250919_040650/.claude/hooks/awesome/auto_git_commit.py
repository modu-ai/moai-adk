#!/usr/bin/env python3
"""
Intelligent git commit hook for Claude Code
Automatically creates meaningful commit messages based on file changes
"""

import os
import sys
import subprocess
from pathlib import Path

def run_command(cmd, check=False):
    """Run shell command and return result"""
    try:
        result = subprocess.run(
            cmd,
            shell=True,
            capture_output=True,
            text=True,
            timeout=5
        )
        if check and result.returncode != 0:
            return None
        return result.stdout.strip() if result.stdout else ""
    except:
        return None

def is_git_repo():
    """Check if current directory is a git repository"""
    result = run_command("git rev-parse --git-dir")
    return result is not None

def get_change_stats(file_path):
    """Get change statistics for a file"""
    # Add file to staging
    run_command(f"git add '{file_path}'")

    # Get number of lines changed
    stats = run_command(f"git diff --cached --numstat '{file_path}'")
    if not stats:
        return 0

    parts = stats.split()
    if len(parts) >= 2:
        added = int(parts[0]) if parts[0].isdigit() else 0
        deleted = int(parts[1]) if parts[1].isdigit() else 0
        return added + deleted
    return 0

def generate_commit_message(file_path, is_new_file=False):
    """Generate intelligent commit message based on file changes"""
    filename = Path(file_path).name
    file_ext = Path(file_path).suffix
    parent_dir = Path(file_path).parent.name

    if is_new_file:
        # Messages for new files
        file_types = {
            '.md': f"📝 Add documentation: {filename}",
            '.py': f"🐍 Add Python module: {filename}",
            '.js': f"📜 Add JavaScript: {filename}",
            '.ts': f"📘 Add TypeScript: {filename}",
            '.jsx': f"⚛️ Add React component: {filename}",
            '.tsx': f"⚛️ Add React TypeScript component: {filename}",
            '.css': f"🎨 Add styles: {filename}",
            '.json': f"📋 Add configuration: {filename}",
            '.yaml': f"⚙️ Add configuration: {filename}",
            '.yml': f"⚙️ Add configuration: {filename}",
            '.sh': f"🔧 Add script: {filename}",
            '.test': f"🧪 Add test: {filename}",
            '.spec': f"🧪 Add test: {filename}",
        }

        for ext, msg in file_types.items():
            if ext in filename:
                return msg

        return f"✨ Add new file: {filename}"
    else:
        # Messages for modified files
        changed_lines = get_change_stats(file_path)

        if changed_lines == 0:
            return None

        # Determine change size
        if changed_lines < 10:
            size = "minor"
            emoji = "🔧"
        elif changed_lines < 50:
            size = "moderate"
            emoji = "📝"
        else:
            size = "major"
            emoji = "🚀"

        # Special handling for specific file types
        if file_ext == '.md':
            return f"📝 Update documentation: {filename}"
        elif file_ext in ['.test', '.spec'] or 'test' in filename:
            return f"🧪 Update tests: {filename}"
        elif filename == 'package.json':
            return f"📦 Update dependencies"
        elif filename in ['requirements.txt', 'Pipfile', 'pyproject.toml']:
            return f"📦 Update Python dependencies"
        elif file_ext in ['.yml', '.yaml'] and 'workflow' in str(Path(file_path)):
            return f"👷 Update CI/CD workflow: {filename}"
        elif parent_dir == 'hooks':
            return f"🪝 Update hook: {filename}"
        elif parent_dir == 'agents':
            return f"🤖 Update agent: {filename}"
        elif parent_dir == 'commands':
            return f"⚡ Update command: {filename}"

        return f"{emoji} Update {filename} ({size}: {changed_lines} lines)"

def commit_file(file_path, message):
    """Create git commit with the given message"""
    try:
        # Commit the file
        result = subprocess.run(
            ['git', 'commit', '-m', message, file_path],
            capture_output=True,
            text=True,
            timeout=5
        )
        return result.returncode == 0
    except:
        return False

def main():
    file_path = os.environ.get('CLAUDE_TOOL_FILE_PATH')
    tool_name = os.environ.get('CLAUDE_TOOL_NAME')

    if not file_path:
        return 0  # Silent exit if no file path

    if not is_git_repo():
        return 0  # Not a git repo, skip

    is_new_file = (tool_name == 'Write')
    message = generate_commit_message(file_path, is_new_file)

    if message:
        if commit_file(file_path, message):
            print(f"📌 Auto-committed: {message}")
        else:
            # Silent fail - don't block the workflow
            pass

    return 0

if __name__ == "__main__":
    sys.exit(main())