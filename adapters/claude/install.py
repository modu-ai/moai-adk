#!/usr/bin/env python3
"""Provisional project-only Claude installer with conservative settings merge."""
import argparse
import hashlib
import json
import os
import pathlib
import stat
import sys
import tempfile

import settings_update

HERE = pathlib.Path(__file__).resolve().parent


def unique_object(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError("duplicate JSON key in settings")
        result[key] = value
    return result


def check_path(root, path):
    for part in [path, *path.parents]:
        if part == root.parent:
            break
        if part.is_symlink():
            raise ValueError(f"symlink conflict: {part}")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("project", type=pathlib.Path)
    parser.add_argument("--style", choices=["MoAI", "MoAI-Easy", "MoAI-Learn"], default="MoAI-Easy")
    parser.add_argument("--recovery-plan", action="store_true")
    args = parser.parse_args()
    root = args.project.resolve(strict=True)
    if not root.is_dir():
        raise ValueError("project must be a directory")
    if args.recovery_plan:
        parse = lambda raw: json.loads(raw, object_pairs_hook=unique_object, parse_constant=lambda value: (_ for _ in ()).throw(ValueError("invalid JSON constant")))
        print(json.dumps(settings_update.recovery_plan(root, parse)))
        return
    command = 'python3 "${CLAUDE_PROJECT_DIR}/.moai-cli/statusline.py"'
    wanted = {"outputStyle": args.style, "statusLine": {"type": "command", "command": command}}
    settings_path = root / ".claude/settings.json"
    files = {".moai-cli/statusline.py": (HERE / "statusline.py").read_bytes()}
    for name in ("moai.md", "moai-easy.md", "moai-learn.md"):
        files[".claude/output-styles/" + name] = (HERE / "output-styles" / name).read_bytes()
    for name in ("LICENSE", "NOTICE.md"):
        files[".moai-cli/licenses/claude-output-styles/" + name] = (HERE / "output-styles" / name).read_bytes()
    # Check every asset and the settings before creating or replacing anything.
    for rel, data in files.items():
        path = root / rel
        check_path(root, path)
        if path.exists() and (not path.is_file() or path.read_bytes() != data):
            raise ValueError(f"existing file conflict: {path}")
    check_path(root, settings_path)
    def parse(raw):
        return json.loads(raw, object_pairs_hook=unique_object, parse_constant=lambda value: (_ for _ in ()).throw(ValueError("invalid JSON constant")))
    plan = settings_update.prepare(root, wanted, parse)
    for rel, data in files.items():
        path = root / rel
        path.parent.mkdir(parents=True, exist_ok=True)
        check_path(root, path)
        try:
            with path.open("xb") as output:
                output.write(data)
        except FileExistsError:
            if path.is_symlink() or not path.is_file() or path.read_bytes() != data:
                raise ValueError(f"concurrent conflict: {path}")
    settings_update.apply(plan)
    print("project preview files present; host loading not verified")


if __name__ == "__main__":
    try:
        main()
    except (OSError, ValueError) as exc:
        print(str(exc), file=sys.stderr)
        sys.exit(1)
