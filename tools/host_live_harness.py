#!/usr/bin/env python3
"""I8 live-session harness (Issue #2245): AGENTS.md loading, skill exposure and hook
firing on Claude Code and Codex, measured in isolated temporary projects.

Usage:
  python3 -I tools/host_live_harness.py --out DIR [--live] [--repo ROOT]
                                        [--bin claude=PATH] [--bin codex=PATH] [--only AC]

Without --live (the default, a dry run) no host is spawned — not even `--version` — and
every item is written as not_run with the command it would have run. With --live an item
runs only when its host binary is found AND its credential variable is set
(ANTHROPIC_API_KEY for Claude Code, CODEX_API_KEY for Codex); otherwise it is not_run.
Only the variable NAME is ever recorded; a value that shows up in raw host output is
replaced by `[REDACTED:<NAME>]` before the file is kept.

Isolation: each item gets a fresh temporary project and a fresh temporary HOME
(HOME, CLAUDE_CONFIG_DIR, CODEX_HOME all point inside it), so the operator's
~/.claude, ~/.codex, ~/.agents and ~/.moai-cli are neither read for configuration nor
written. Nothing is written outside DIR and the temporary tree, which is removed.

Output (REQ-I8-01): DIR/receipts.jsonl (schemas/e2e-receipt.schema.json fields, appended),
DIR/<AC>.out / .err (raw host output), DIR/rows.md (docs/support-matrix.md rows: passed
items in the supported-row shape, the rest in the not_run-row shape; see render_rows).
rows.md is also printed.

Exit codes: 0 = receipts written, 1 = unreadable input or I/O failure, 2 = usage.
"""

import argparse
import hashlib
import json
import os
import platform
import secrets
import shlex
import shutil
import subprocess
import sys
import tempfile
from datetime import datetime, timezone
from pathlib import Path

HOSTS = {
    "claude": {"receipt": "claude-code", "label": "Claude Code", "credential": "ANTHROPIC_API_KEY"},
    "codex": {"receipt": "codex", "label": "Codex", "credential": "CODEX_API_KEY"},
}
PROMPT = "Reply with the value of MOAI_PROBE_MARKER from your project instructions, and nothing else."
TIMEOUT_SECS = 600
PENDING = "결정 대기: RC1 MoAI hook 0개(P04 = A), 양 호스트 hook 생성은 별도 결정(Codex 계약 C4)"
CODEX_HOOK_MANUAL = ("manual only: Codex runs a project hook only after hook trust review (/hooks); "
                     "the per-run trust bypass flag is a security bypass and is not used")

# (ac, host, kind, capability text). kind decides fixture, command and verdict.
ITEMS = [
    ("AC-I8C-1", "claude", "agents", "AGENTS.md 로드 (CLAUDE.md 없음)"),
    ("AC-I8C-2", "claude", "agents_shadowed", "CLAUDE.md가 있으면 AGENTS.md를 읽지 않음"),
    ("AC-I8C-3", "claude", "skill", "project skill 노출·명시 호출 (/moai-probe)"),
    ("AC-I8C-4", "claude", "hook", "project SessionStart hook 발화 (host 메커니즘)"),
    ("AC-I8C-5", "claude", "pending", "MoAI hook 발화"),
    ("AC-I8X-1", "codex", "agents", "AGENTS.md 로드"),
    ("AC-I8X-2", "codex", "agents_override", "AGENTS.override.md가 AGENTS.md를 대체"),
    ("AC-I8X-3", "codex", "skill", "project skill 노출·명시 호출 ($moai-probe)"),
    ("AC-I8X-4", "codex", "manual", "project hook 발화 (host 메커니즘)"),
    ("AC-I8X-5", "codex", "pending", "MoAI hook 발화"),
]


def now() -> str:
    return datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")


def sha256_file(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def write(path: Path, text: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text, encoding="utf-8")


def agents_text(marker: str) -> str:
    return f"# Probe\n\nMOAI_PROBE_MARKER={marker}\n\nWhen asked for MOAI_PROBE_MARKER, reply with its value only.\n"


def probe_skill(marker: str) -> str:
    return ("---\nname: moai-probe\ndescription: I8 harness probe skill. Use only when invoked by name.\n---\n\n"
            f"When this skill is invoked, reply with exactly: {marker}\n")


def build_fixture(kind: str, host: str, project: Path, repo: Path, markers: dict) -> None:
    """Writes the item's temporary project. markers gets the expected/forbidden values."""
    skills = ".claude/skills" if host == "claude" else ".agents/skills"
    shutil.copytree(repo / "skills/moai", project / skills / "moai")
    if kind in ("agents", "agents_shadowed", "agents_override"):
        markers["expect"] = "mp-" + secrets.token_hex(8)
        write(project / "AGENTS.md", agents_text(markers["expect"]))
    if kind == "agents_shadowed":
        markers["forbid"], markers["expect"] = markers["expect"], None
        write(project / "CLAUDE.md", "# Project notes\n\nNo imports here.\n")
    if kind == "agents_override":
        markers["forbid"] = markers["expect"]
        markers["expect"] = "mp-" + secrets.token_hex(8)
        write(project / "AGENTS.override.md", agents_text(markers["expect"]))
    if kind == "skill":
        markers["expect"] = "mp-" + secrets.token_hex(8)
        write(project / skills / "moai-probe/SKILL.md", probe_skill(markers["expect"]))
    if kind == "hook":
        markers["expect"] = "mp-" + secrets.token_hex(8)
        target = project / "hook-fired.txt"
        command = f"printf %s {shlex.quote(markers['expect'])} > {shlex.quote(str(target))}"
        hooks = {"hooks": {"SessionStart": [{"hooks": [{"type": "command", "command": command}]}]}}
        write(project / ".claude/settings.json", json.dumps(hooks, indent=2) + "\n")


def command(kind: str, host: str, binary: str, out: Path, ac: str) -> list:
    if host == "claude":
        prompt = "/moai-probe" if kind == "skill" else ("Reply OK." if kind == "hook" else PROMPT)
        return [binary, "-p", prompt, "--output-format", "json"]
    prompt = "$moai-probe" if kind == "skill" else PROMPT
    return [binary, "exec", "--json", "--ephemeral", "--skip-git-repo-check", "--sandbox", "read-only",
            "-o", str(out / f"{ac}.last.md"), prompt]


def claude_result(stdout: str):
    try:
        data = json.loads(stdout)
    except json.JSONDecodeError:
        return None
    if not isinstance(data, dict) or data.get("is_error") is True:
        return None
    result = data.get("result")
    return result if isinstance(result, str) else None


def jsonl_types(stdout: str) -> set:
    types = set()
    for line in stdout.splitlines():
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            continue
        if isinstance(event, dict):
            types.add(event.get("type"))
    return types


def verdict(kind: str, host: str, stdout: str, project: Path, out: Path, ac: str, markers: dict):
    """(result, reason). Structural only: a random marker present or absent, never prose."""
    if host == "claude":
        text = claude_result(stdout)
        if text is None:
            return "failed", "stdout is not a JSON result object with a string .result"
    else:
        if "turn.completed" not in jsonl_types(stdout):
            return "failed", "no turn.completed event in the JSONL stream"
        last = out / f"{ac}.last.md"
        text = last.read_text(encoding="utf-8") if last.is_file() else ""
    if kind == "hook":
        fired = project / "hook-fired.txt"
        got = fired.read_text(encoding="utf-8") if fired.is_file() else ""
        return ("passed", "hook wrote the marker") if got == markers["expect"] else ("failed", "hook marker file absent or wrong")
    if markers.get("forbid") and markers["forbid"] in text:
        return "failed", "the shadowed instruction file's marker appeared"
    if markers.get("expect") and markers["expect"] not in text:
        return "failed", "expected marker absent from the reply"
    return "passed", "marker check held"


def isolated_env(home: Path, credential: str) -> dict:
    env = {k: os.environ[k] for k in ("PATH", "LANG", "LC_ALL", "TMPDIR") if k in os.environ}
    env.update(HOME=str(home), CLAUDE_CONFIG_DIR=str(home / ".claude"), CODEX_HOME=str(home / ".codex"))
    if credential in os.environ:
        env[credential] = os.environ[credential]
    return env


def redact(path: Path, credential: str) -> None:
    value = os.environ.get(credential)
    if value and path.is_file():
        text = path.read_text(encoding="utf-8", errors="replace")
        if value in text:
            path.write_text(text.replace(value, f"[REDACTED:{credential}]"), encoding="utf-8")


def resolve_bin(host: str, overrides: dict):
    """An executable path, or None. An override that is not executable counts as absent."""
    return shutil.which(overrides[host]) if host in overrides else shutil.which(host)


def host_version(binary: str, env: dict):
    try:
        done = subprocess.run([binary, "--version"], capture_output=True, text=True, timeout=60, env=env)
    except (OSError, subprocess.TimeoutExpired):
        return None
    lines = done.stdout.strip().splitlines()
    return lines[0] if lines else None


def run_item(item, args, overrides: dict, skill_digest: str, cli_sha: str, versions: dict) -> dict:
    ac, host, kind, _ = item
    meta = HOSTS[host]
    out = args.out
    receipt = {"ac": ac, "os": platform.system(), "arch": platform.machine(), "host": meta["receipt"],
               "host_exact_version": None, "mode": "local", "skill_digest": skill_digest,
               "moai_cli_sha": cli_sha, "probe": "I8", "kind": kind}
    binary = resolve_bin(host, overrides)
    if kind == "pending":
        return {**receipt, "result": "not_run", "ran_at": now(), "detail": PENDING}
    if kind == "manual":
        return {**receipt, "result": "not_run", "ran_at": now(), "detail": CODEX_HOOK_MANUAL}
    # Bare host name and an <out> placeholder: a receipt may later be committed, so it
    # carries no absolute machine path.
    argv_preview = " ".join(shlex.quote(a) for a in command(kind, host, host, Path("<out>"), ac))
    if not args.live:
        return {**receipt, "result": "not_run", "ran_at": now(), "detail": f"dry-run: not executed: {argv_preview}"}
    if not binary:
        return {**receipt, "result": "not_run", "ran_at": now(), "detail": f"{host} binary not found; not executed"}
    if not os.environ.get(meta["credential"]):
        return {**receipt, "result": "not_run", "ran_at": now(),
                "detail": f"credential {meta['credential']} not set; not executed"}
    with tempfile.TemporaryDirectory(prefix=f"moai-i8-{ac}-") as tmp:
        home, project, markers = Path(tmp) / "home", Path(tmp) / "project", {}
        home.mkdir()
        project.mkdir()
        env = isolated_env(home, meta["credential"])
        if host not in versions:
            versions[host] = host_version(binary, env)
        receipt["host_exact_version"] = versions[host]
        if host == "codex":  # Codex walks AGENTS.md from the project root (usually the Git root)
            subprocess.run(["git", "init", "-q"], cwd=project, env=env, capture_output=True)
        build_fixture(kind, host, project, args.repo, markers)
        stdout_path, stderr_path = out / f"{ac}.out", out / f"{ac}.err"
        try:
            done = subprocess.run(command(kind, host, binary, out, ac), cwd=project, env=env,
                                  capture_output=True, text=True, timeout=TIMEOUT_SECS)
            stdout_path.write_text(done.stdout, encoding="utf-8")
            stderr_path.write_text(done.stderr, encoding="utf-8")
            if done.returncode != 0:
                result, reason = "failed", f"exit {done.returncode}"
            else:
                result, reason = verdict(kind, host, done.stdout, project, out, ac, markers)
        except subprocess.TimeoutExpired:
            result, reason = "failed", f"timed out after {TIMEOUT_SECS}s"
        for path in (stdout_path, stderr_path, out / f"{ac}.last.md"):
            redact(path, meta["credential"])
    return {**receipt, "result": result, "ran_at": now(), "detail": f"{stdout_path.name} ({reason})"}


def render_rows(receipts: list) -> str:
    """docs/support-matrix.md rows. A passed receipt becomes a supported row citing
    receipts.jsonl; every other result becomes a not_run row with its reason."""
    caps = {ac: (host, kind, cap) for ac, host, kind, cap in ITEMS}
    supported, others = [], []
    for r in receipts:
        host, kind, cap = caps[r["ac"]]
        label = HOSTS[host]["label"]
        osarch = f"{r['os']}/{r['arch']}"
        if r["result"] == "passed":
            discovery = "✓" if kind in ("agents", "agents_shadowed", "agents_override", "skill") else "—"
            invoke = "✓" if kind == "skill" else "—"
            supported.append(f"| {osarch} | {label} | {r['host_exact_version'] or 'unknown'} | {r['mode']} | `{r['skill_digest'][:8]}…` "
                             f"| {r['moai_cli_sha'] if r['moai_cli_sha'] == 'absent' else r['moai_cli_sha'][:8] + '…'} "
                             f"| {r['ac']} | {cap} | {discovery} | {invoke} | — | — | — | — | passed | receipts.jsonl |")
        else:
            reason = str(r.get("detail") or "").replace("|", "\\|")
            others.append(f"| {osarch} | {label} | {r['mode']} | {cap} ({r['ac']}) | {r['result']} | {reason} |")
    head_s = ("| OS/arch (Rust tier) | host | host exact version | mode | skill digest (sha256) | moai-cli SHA (sha256) "
              "| AC | capability | discovery | 명시 호출 | SETUP | UPDATE | 권한 거부 | restore | 결과 | receipt |\n"
              "| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |")
    head_o = "| OS/arch | host | mode | 항목 | 결과 | 사유 |\n| --- | --- | --- | --- | --- | --- |"
    parts = []
    if supported:
        parts.append("### I8 지원 행 (passed receipt)\n\n" + head_s + "\n" + "\n".join(supported))
    if others:
        parts.append("### I8 not_run·failed 행\n\n" + head_o + "\n" + "\n".join(others))
    return "\n\n".join(parts) + "\n"


def parse_args(argv: list):
    parser = argparse.ArgumentParser(description="I8 live-session harness (#2245)")
    parser.add_argument("--out", required=True, type=Path, help="receipt directory (created)")
    parser.add_argument("--live", action="store_true", help="spawn hosts (operator only)")
    parser.add_argument("--repo", type=Path, default=Path.cwd(), help="repository root (default: cwd)")
    parser.add_argument("--bin", action="append", default=[], metavar="HOST=PATH", help="host binary override")
    parser.add_argument("--only", action="append", default=[], metavar="AC", help="run only these items")
    args = parser.parse_args(argv)
    overrides = {}
    for spec in args.bin:
        host, sep, path = spec.partition("=")
        if not sep or host not in HOSTS or not path:
            parser.error(f"--bin expects claude=PATH or codex=PATH, got {spec!r}")
        overrides[host] = path
    known = {item[0] for item in ITEMS}
    unknown = [ac for ac in args.only if ac not in known]
    if unknown:
        parser.error(f"unknown item(s): {', '.join(unknown)}")
    return args, overrides


def main(argv: list) -> int:
    try:
        args, overrides = parse_args(argv)
    except SystemExit as exc:
        return 2 if exc.code else 0
    skill_md = args.repo / "skills/moai/SKILL.md"
    if not skill_md.is_file():
        print(f"not a repository root (skills/moai/SKILL.md missing): {args.repo}", file=sys.stderr)
        return 1
    try:
        args.out.mkdir(parents=True, exist_ok=True)
        args.out = args.out.resolve()
        cli = shutil.which("moai-cli")
        cli_sha = sha256_file(Path(cli)) if cli else "absent"
        skill_digest, versions, receipts = sha256_file(skill_md), {}, []
        for item in ITEMS:
            if args.only and item[0] not in args.only:
                continue
            receipts.append(run_item(item, args, overrides, skill_digest, cli_sha, versions))
        with (args.out / "receipts.jsonl").open("a", encoding="utf-8") as fh:
            for r in receipts:
                fh.write(json.dumps(r, ensure_ascii=False) + "\n")
        rows = render_rows(receipts)
        write(args.out / "rows.md", rows)
    except OSError as exc:
        print(f"I/O failure: {exc}", file=sys.stderr)
        return 1
    print(rows, end="")
    counts = {k: sum(r["result"] == k for r in receipts) for k in ("passed", "failed", "not_run")}
    print(f"mode={'live' if args.live else 'dry-run'} items={len(receipts)} " +
          " ".join(f"{k}={v}" for k, v in counts.items()), file=sys.stderr)
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
