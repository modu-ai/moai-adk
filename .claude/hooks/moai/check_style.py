#!/usr/bin/env python3
"""
Claude Code PostToolUse 훅: 파일별 린트/포맷 실행기

입력: stdin JSON (docs/cc-docs/hooks.md의 PostToolUse 입력 스키마)
동작: Write/Edit/MultiEdit로 생성/수정된 파일 경로를 감지하여
      언어별로 가벼운 포맷/린트 도구를 가능한 범위에서 실행.

설계 원칙
- 빠름: 단일 파일 기준으로만 동작, 무거운 전역 스캔 금지
- 안전: 도구 미설치 시 건너뛰고 안내
- 온건: 가능하면 자동 포맷, 린트는 요약만 출력
- 비차단: PostToolUse는 차단하지 않고 피드백만 제공

지원 대상(도구가 있을 때만 실행)
- Python: ruff --fix | flake8, black, isort, py_compile
- JS/TS/JSX/TSX/MD/CSS/HTML/JSON/YAML: prettier, eslint, jq, yamllint
- Go: gofmt -w
- Rust: rustfmt
- Shell: shfmt -w, shellcheck
- C/C++: clang-format -i

확장: 기타 언어는 명령어 존재 시만 실행 (예: php -l, rubocop 등)
"""

from __future__ import annotations

import json
import os
import shlex
import sys
import subprocess
from pathlib import Path
from typing import List, Tuple, Optional


def _which(cmd: str) -> Optional[str]:
    from shutil import which

    return which(cmd)


def _run(cmd: List[str], cwd: Optional[Path] = None, timeout_sec: int = 20) -> Tuple[int, str, str]:
    try:
        proc = subprocess.run(
            cmd,
            cwd=str(cwd) if cwd else None,
            capture_output=True,
            text=True,
            timeout=timeout_sec,
        )
        return proc.returncode, proc.stdout.strip(), proc.stderr.strip()
    except subprocess.TimeoutExpired:
        return 124, "", f"timeout: {shlex.join(cmd)}"
    except Exception as e:  # 실행 실패는 차단하지 않음
        return 1, "", f"error: {e}"


def _is_code_tool_event(tool_name: str) -> bool:
    return tool_name in {"Write", "Edit", "MultiEdit"}


def _ext(path: str) -> str:
    p = path.lower()
    # .d.ts 같은 복합 확장자 우선 처리
    if p.endswith(".d.ts"):
        return "d.ts"
    if p.endswith(".d.tsx"):
        return "d.tsx"
    return Path(p).suffix.lstrip(".")


def _python_tools(file_path: str) -> List[List[str]]:
    tools: List[List[str]] = []
    if _which("python3"):
        tools.append([sys.executable or "python3", "-m", "py_compile", file_path])
    if _which("ruff"):
        tools.append(["ruff", "check", "--fix", file_path])
    elif _which("flake8"):
        tools.append(["flake8", file_path])
    if _which("black"):
        tools.append(["black", file_path])
    if _which("isort"):
        tools.append(["isort", file_path])
    return tools


def _js_ts_tools(file_path: str) -> List[List[str]]:
    tools: List[List[str]] = []
    if _which("npx") and _which("node"):
        # prettier 먼저 적용
        tools.append(["npx", "--yes", "prettier", "--write", file_path])
        # eslint는 존재할 때만 (프로젝트에 설정 필요)
        if _which("eslint") or True:
            tools.append(["npx", "--yes", "eslint", "--fix", file_path])
    return tools


def _web_text_tools(file_path: str) -> List[List[str]]:
    tools: List[List[str]] = []
    if _which("npx") and _which("node"):
        tools.append(["npx", "--yes", "prettier", "--write", file_path])
    return tools


def _json_yaml_tools(file_path: str) -> List[List[str]]:
    tools: List[List[str]] = []
    # JSON 검증/정렬
    if file_path.lower().endswith(".json") and _which("jq"):
        # 포맷팅: jq -S . <file> > tmp && mv
        tools.append(["bash", "-lc", f"tmp=$(mktemp) && jq -S . {shlex.quote(file_path)} > $tmp && mv $tmp {shlex.quote(file_path)}"])
    # YAML 검증
    if file_path.lower().endswith(('.yml', '.yaml')) and _which("yamllint"):
        tools.append(["yamllint", "-s", file_path])
    return tools


def _go_tools(file_path: str) -> List[List[str]]:
    tools: List[List[str]] = []
    if _which("gofmt"):
        tools.append(["gofmt", "-w", file_path])
    return tools


def _rust_tools(file_path: str) -> List[List[str]]:
    tools: List[List[str]] = []
    if _which("rustfmt"):
        tools.append(["rustfmt", file_path])
    return tools


def _shell_tools(file_path: str) -> List[List[str]]:
    tools: List[List[str]] = []
    if _which("shfmt"):
        tools.append(["shfmt", "-w", file_path])
    if _which("shellcheck"):
        tools.append(["shellcheck", file_path])
    return tools


def _clang_tools(file_path: str) -> List[List[str]]:
    tools: List[List[str]] = []
    if _which("clang-format"):
        tools.append(["clang-format", "-i", file_path])
    return tools


def _misc_lang_tools(file_path: str) -> List[List[str]]:
    tools: List[List[str]] = []
    p = file_path.lower()
    if p.endswith(".php") and _which("php"):
        tools.append(["php", "-l", file_path])
    if p.endswith(".rb") and _which("rubocop"):
        tools.append(["rubocop", "-A", file_path])
    if p.endswith(".kt") and _which("ktlint"):
        tools.append(["ktlint", "-F", file_path])
    if p.endswith(".scala") and _which("scalafmt"):
        tools.append(["scalafmt", file_path])
    if p.endswith(".cs") and _which("dotnet"):
        # 단일 파일 포맷은 제한적이므로 시도만
        tools.append(["dotnet", "format", file_path])
    if p.endswith(".swift") and _which("swiftformat"):
        tools.append(["swiftformat", file_path])
    if p.endswith(".dart") and _which("dart"):
        tools.append(["dart", "format", file_path])
    return tools


def build_tool_commands(file_path: str) -> List[List[str]]:
    ext = _ext(file_path)
    cmds: List[List[str]] = []
    if ext in {"py"}:
        cmds += _python_tools(file_path)
    elif ext in {"js", "jsx", "ts", "tsx", "d.ts", "d.tsx"}:
        cmds += _js_ts_tools(file_path)
    elif ext in {"md", "mdx", "css", "scss", "sass", "less", "html"}:
        cmds += _web_text_tools(file_path)
    elif ext in {"json", "yml", "yaml"}:
        cmds += _json_yaml_tools(file_path)
    elif ext in {"go"}:
        cmds += _go_tools(file_path)
    elif ext in {"rs"}:
        cmds += _rust_tools(file_path)
    elif ext in {"sh"}:
        cmds += _shell_tools(file_path)
    elif ext in {"c", "cpp", "cc", "cxx", "h", "hpp"}:
        cmds += _clang_tools(file_path)
    # 기타 언어(선택적)
    cmds += _misc_lang_tools(file_path)
    return cmds


def main() -> int:
    try:
        data = json.load(sys.stdin)
    except json.JSONDecodeError as e:
        print(f"hook input JSON parse error: {e}")
        return 0  # 비차단

    tool_name = data.get("tool_name", "")
    tool_input = data.get("tool_input", {}) or {}
    file_path = tool_input.get("file_path") or tool_input.get("filePath")

    if not _is_code_tool_event(tool_name):
        return 0
    if not file_path or not isinstance(file_path, str):
        return 0

    abs_path = Path(file_path).resolve()
    if not abs_path.exists() or abs_path.is_dir():
        return 0

    cmds = build_tool_commands(str(abs_path))
    if not cmds:
        print(f"no style tools matched for: {abs_path.name}")
        return 0

    project_dir = os.environ.get("CLAUDE_PROJECT_DIR")
    cwd = Path(project_dir) if project_dir else abs_path.parent

    print(f"style hooks: {abs_path.name}")
    for cmd in cmds:
        code, out, err = _run(cmd, cwd=cwd)
        icon = "✓" if code == 0 else ("⚠️" if code in {124} else "✗")
        print(f" {icon} {' '.join(map(shlex.quote, cmd))}")
        if out:
            print(out)
        if err:
            print(err)

    return 0


if __name__ == "__main__":
    sys.exit(main())
