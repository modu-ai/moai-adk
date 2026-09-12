#!/usr/bin/env python3
"""Bounded MoAI product-seam probes. Never persist raw terminal/API output."""
import argparse
import hashlib
import json
import os
import pathlib
import pty
import re
import select
import signal
import struct
import subprocess
import termios
import fcntl
import time

ANSI = re.compile(r"\x1b(?:\[[0-?]*[ -/]*[@-~]|\][^\x07]*(?:\x07|\x1b\\))")
MODEL = re.compile(r"\b(?:gpt-[a-zA-Z0-9.-]+|glm-[a-zA-Z0-9.-]+|claude-[a-zA-Z0-9.-]+)\b")
ERRORS = ["401 Unauthorized", "400 Bad Request", "API Error", "not logged in", "Unknown command", "gateway resume", "permission denied"]


def capture(argv, cwd, seconds, picker=False, interactive_marker=None):
    env = os.environ.copy()
    # Do not inherit a parent Claude session; authentication stays in product stores.
    for key in list(env):
        if key.startswith(("ANTHROPIC_", "CLAUDE_CODE_", "MOAI_SESSION", "MOAI_KANBAN", "MOAI_FACTORY")) or key in ("CLAUDECODE", "CLAUDE_SESSION_ID", "TMUX", "TMUX_PANE"):
            env.pop(key, None)
    env.update(TERM="xterm-256color", CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC="1")
    master, slave = pty.openpty()
    fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack("HHHH", 50, 180, 0, 0))
    proc = None
    output = bytearray()
    sent = False
    selected = False
    prompt_sent = False
    reply_seen_at = None
    api_key_accepted = False
    cleanup_errors = []
    started = time.monotonic()
    timeout = False
    try:
        proc = subprocess.Popen(argv, cwd=cwd, env=env, stdin=slave, stdout=slave, stderr=slave, start_new_session=True)
        os.close(slave)
        slave = -1
        while time.monotonic() - started < seconds:
            ready, _, _ = select.select([master], [], [], 0.2)
            if ready:
                try:
                    chunk = os.read(master, 65536)
                except OSError:
                    break
                if not chunk:
                    break
                output.extend(chunk)
                if len(output) > 8 * 1024 * 1024:
                    raise RuntimeError("bounded output limit exceeded")
            screen_compact = re.sub(r"\s", "", ANSI.sub("", output.decode("utf-8", "replace")))
            if picker and not api_key_accepted and "DoyouwanttousethisAPIkey?" in screen_compact:
                os.write(master, b"\x1b[A\r")
                api_key_accepted = True
            if picker and not sent and time.monotonic() - started > 7:
                os.write(master, b"/model\r")
                sent = True
            if picker and sent and not selected and time.monotonic() - started > 12:
                os.write(master, b"\x1b[As")
                selected = True
            if interactive_marker and selected and not prompt_sent and "forthissessiononly" in screen_compact:
                os.write(master, ("Reply with exactly " + interactive_marker + ".\r").encode())
                prompt_sent = True
            if interactive_marker and prompt_sent and ("⏺" + interactive_marker) in screen_compact:
                if reply_seen_at is None:
                    reply_seen_at = time.monotonic()
                elif time.monotonic() - reply_seen_at > 6:
                    break
            if proc.poll() is not None:
                break
        else:
            timeout = True
    finally:
        if slave != -1:
            os.close(slave)
        if proc is not None:
            # Kill the whole group even if launcher itself has already exited.
            try:
                os.killpg(proc.pid, signal.SIGTERM)
            except ProcessLookupError:
                pass
            except PermissionError:
                cleanup_errors.append("signal_permission_denied")
            try:
                proc.wait(timeout=3)
            except subprocess.TimeoutExpired:
                pass
            try:
                os.killpg(proc.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
            except PermissionError:
                cleanup_errors.append("signal_permission_denied")
            proc.wait(timeout=3)
        os.close(master)
    return ANSI.sub("", output.decode("utf-8", "replace")), {
        "exit_code": proc.returncode, "elapsed_seconds": round(time.monotonic() - started, 2),
        "timeout": timeout, "group_cleanup_attempted": True, "picker_command_sent": sent, "selection_sent": selected, "interactive_prompt_sent": prompt_sent, "test_family_api_key_accepted": api_key_accepted, "cleanup_errors": cleanup_errors,
    }


def summarize(text, meta, marker, picker):
    meta["prompt_categories"] = [x for x in ["Select login method", "Choose the text style", "trust this folder", "trust the files", "Choose the theme", "Welcome to Claude Code", "Do you want to use this API key?", "Bypass Permissions mode", "No, exit"] if x.lower().replace(" ", "") in text.lower().replace(" ", "")]
    meta["diagnostic_categories"] = re.findall(r"T649_DIAG[^\r\n]{0,240}", text)
    meta["error_categories"] = [x for x in ERRORS if x.lower() in text.lower()]
    meta["observed_model_ids"] = sorted(set(MODEL.findall(text)))
    normalized = text.replace(r"\r", "\n").replace("\r", "\n")
    compact = re.sub(r"[ \t]", "", normalized)
    meta["picker_visible"] = "Selectmodel" in compact
    # Only numbered picker rows; startup warnings and status text are excluded.
    rows = [line for line in compact.splitlines() if re.match(r"^[❯> ]?\d+\.", line)]
    meta["picker_rows"] = [{"default": "Default" in line, "model_ids": MODEL.findall(" ".join(re.findall(r"Custommodel\(([^)]+)\)", line))) if "Custommodel(" in line else MODEL.findall(line)} for line in rows]
    selection = re.search(r"Setmodelto([^\n]+?)forthissessiononly", compact)
    meta["selected_model_ids"] = MODEL.findall(selection.group(1)) if selection else []
    meta["selected_default"] = bool(selection and "default" in selection.group(1).lower())
    meta["session_only_selection_confirmed"] = bool(re.search(r"Setmodelto[^\n]+forthissessiononly", compact))
    events = []
    for line in text.splitlines():
        try:
            value = json.loads(line)
        except ValueError:
            continue
        if isinstance(value, dict):
            events.append(value)
    results = [v for v in events if v.get("type") == "result"]
    tool_names = []
    assistant_models = []
    for event in events:
        message = event.get("message", {})
        if not isinstance(message, dict):
            continue
        if event.get("type") == "assistant" and isinstance(message.get("model"), str):
            assistant_models.append(message["model"])
        for block in message.get("content", []):
            if isinstance(block, dict) and block.get("type") == "tool_use":
                tool_names.append(block.get("name"))
    meta["assistant_models"] = sorted(set(assistant_models))
    meta["tool_names"] = tool_names
    meta["result_event_count"] = len(results)
    meta["marker_in_final_result"] = any(marker in str(v.get("result", "")) for v in results)
    meta["session_id"] = next((v.get("session_id") for v in reversed(results) if v.get("session_id")), None)
    meta["result_subtypes"] = [v.get("subtype") for v in results]
    meta["success"] = (meta["picker_visible"] and not meta["error_categories"]) if picker else (
        meta["exit_code"] == 0 and meta["marker_in_final_result"] and not meta["error_categories"]
        and all(not v.get("is_error", False) for v in results))
    return meta


def main():
    p = argparse.ArgumentParser()
    p.add_argument("--binary", required=True, type=pathlib.Path)
    p.add_argument("--cwd", required=True, type=pathlib.Path)
    p.add_argument("--output", required=True, type=pathlib.Path)
    p.add_argument("--mode", choices=["cc", "glm", "gpt"], default="gpt")
    p.add_argument("--journey", choices=["picker", "answer", "tool", "resume"], required=True)
    p.add_argument("--model", default="gpt-6-astra")
    p.add_argument("--expected-models", nargs="+", help="Exact permitted custom picker model IDs")
    p.add_argument("--profile", help="Existing MoAI profile name; no profile is created or changed by the harness")
    p.add_argument("--resume-session")
    p.add_argument("--marker", default="T649_GATEWAY_OK")
    p.add_argument("--seconds", type=int, default=90)
    p.add_argument("--allow-live", action="store_true", help="Explicitly permit inference for answer/tool/resume")
    a = p.parse_args()
    if not a.binary.is_file() or not a.cwd.is_dir():
        p.error("binary and cwd must exist")
    if not 10 <= a.seconds <= 300:
        p.error("seconds must be 10..300")
    if not re.fullmatch(r"[A-Z0-9_]{1,80}", a.marker):
        p.error("marker must be synthetic uppercase alphanumeric")
    if a.journey != "picker" and not a.allow_live:
        p.error("live journey requires --allow-live")
    if a.journey == "resume" and not a.resume_session:
        p.error("resume requires --resume-session")
    argv = [str(a.binary.resolve()), a.mode]
    if a.profile:
        argv += ["-p", a.profile]
    argv += ["--permission-mode", "default", "--settings", '{"disableAllHooks":true}', "--model", a.model, "--strict-mcp-config", "--mcp-config", '{"mcpServers":{}}']
    if a.journey != "picker":
        argv += ["--print", "--output-format", "stream-json", "--verbose"]
        prompt = "Reply with exactly " + a.marker + "."
        if a.journey == "tool":
            fixture = a.cwd / "t649-fixture.txt"
            if fixture.exists():
                p.error("refusing to overwrite existing fixture")
            fixture.write_text(a.marker + "\n", encoding="utf-8")
            argv += ["--allowedTools", "Read"]
            prompt = "Use the Read tool to read t649-fixture.txt in the working directory, then reply with its exact content."
        if a.journey == "resume":
            argv += ["--resume", a.resume_session]
            prompt = "Repeat exactly the uppercase marker from your preceding response."
        argv += ["--", prompt]
    try:
        text, meta = capture(argv, str(a.cwd.resolve()), a.seconds, a.journey == "picker")
    finally:
        if a.journey == "tool":
            fixture.unlink(missing_ok=True)
    result = summarize(text, meta, a.marker, a.journey == "picker")
    result.update(journey=a.journey, mode=a.mode, requested_model=a.model,
                  binary_sha256=hashlib.sha256(a.binary.read_bytes()).hexdigest())
    if a.journey == "picker":
        actual = {m for row in result["picker_rows"] if not row["default"] for m in row["model_ids"]}
        expected = set(a.expected_models or [])
        result["picker_exact_allowlist"] = bool(expected) and actual == expected
        result["success"] = result["success"] and result["picker_exact_allowlist"] and result["session_only_selection_confirmed"]
    if a.journey == "tool":
        result["success"] = result["success"] and "Read" in result["tool_names"]
    a.output.parent.mkdir(parents=True, exist_ok=True)
    a.output.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(result, indent=2))
    raise SystemExit(0 if result["success"] else 1)


if __name__ == "__main__":
    main()
