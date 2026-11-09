"""Pi v1.5.0 pin spike harness (#2199): conformance, security, cancellation, auth.

Runs the pinned `@earendil-works/pi-coding-agent@1.5.0` install under an allowlisted
environment (no provider keys, isolated HOME and agent dir, PI_OFFLINE=1) against a local
OpenAI-compatible mock. No hosted model is called; no credential is issued or read.

Usage (from the repository root):
  python3 -I docs/spikes/pi-v1.5.0/measure_pi.py --node-bin <dir with node>=22.19> \
      --install <dir holding node_modules with pi 1.5.0> --work <scratch dir>
Writes <work>/results.json and prints it.
"""

import argparse
import http.server
import json
import queue
import shutil
import subprocess
import threading
import time
from pathlib import Path

ap = argparse.ArgumentParser()
ap.add_argument("--node-bin", required=True)
ap.add_argument("--install", required=True)
ap.add_argument("--work", required=True)
args = ap.parse_args()
NODE_BIN = Path(args.node_bin).resolve()
INSTALL = Path(args.install).resolve()
WORK = Path(args.work).resolve()
PKG = INSTALL / "node_modules/@earendil-works/pi-coding-agent"
CLI = PKG / "dist/bundle/cli.js"
HOME = WORK / "home"
AGENT_DIR = HOME / ".pi/agent"
PROJ = WORK / "proj"
for d in (HOME, PROJ):
    shutil.rmtree(d, ignore_errors=True)
AGENT_DIR.mkdir(parents=True)
PROJ.mkdir(parents=True)
R = {}


def rel(p):
    return str(Path(p)).replace(str(WORK), "<work>")


# ---------------------------------------------------------------- mock provider
class Mock:
    mode = "slow"  # slow | short | tool
    marker = ""
    requests = []
    disconnect_at = None


class Handler(http.server.BaseHTTPRequestHandler):
    def log_message(self, *a):
        pass

    def do_GET(self):
        Mock.requests.append({"method": "GET", "path": self.path})
        self.send_response(404)
        self.end_headers()

    def do_POST(self):
        body = json.loads(self.rfile.read(int(self.headers.get("content-length", 0))) or b"{}")
        auth = self.headers.get("authorization", "")
        Mock.requests.append({
            "method": "POST", "path": self.path, "mode": Mock.mode,
            "auth_bearer_matches_command_key": auth == "Bearer mock-key-from-command",
            "has_tool_result": any(m.get("role") == "tool" for m in body.get("messages", [])),
            "header_names": sorted(h.lower() for h in self.headers.keys()),
        })
        self.send_response(200)
        self.send_header("content-type", "text/event-stream")
        self.end_headers()

        def chunk(delta, finish=None):
            c = {"id": "m", "object": "chat.completion.chunk", "created": 0, "model": "mock-1",
                 "choices": [{"index": 0, "delta": delta, "finish_reason": finish}]}
            self.wfile.write(b"data: " + json.dumps(c).encode() + b"\n\n")
            self.wfile.flush()

        try:
            if Mock.mode == "tool" and not Mock.requests[-1]["has_tool_result"]:
                arg = json.dumps({"command": "touch " + Mock.marker})
                chunk({"role": "assistant", "tool_calls": [{"index": 0, "id": "call_1", "type": "function",
                                                            "function": {"name": "bash", "arguments": arg}}]})
                chunk({}, "tool_calls")
            elif Mock.mode == "slow":
                for _ in range(300):
                    chunk({"role": "assistant", "content": "tick "})
                    time.sleep(0.1)
                chunk({}, "stop")
            else:
                chunk({"role": "assistant", "content": "ok"})
                chunk({}, "stop")
            self.wfile.write(b"data: [DONE]\n\n")
            self.wfile.flush()
        except (BrokenPipeError, ConnectionResetError):
            Mock.disconnect_at = time.monotonic()


srv = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Handler)
threading.Thread(target=srv.serve_forever, daemon=True).start()
PORT = srv.server_address[1]
(AGENT_DIR / "models.json").write_text(json.dumps({"providers": {"mock": {
    "baseUrl": f"http://127.0.0.1:{PORT}/v1", "api": "openai-completions",
    "apiKey": "!printf mock-key-from-command",
    "models": [{"id": "mock-1", "cost": {"input": 0, "output": 0, "cacheRead": 0, "cacheWrite": 0}}],
}}}))

# ---------------------------------------------------------------- environment
ENV = {
    "PATH": f"{NODE_BIN}:/usr/bin:/bin",
    "HOME": str(HOME),
    "PI_CODING_AGENT_DIR": str(AGENT_DIR),
    "PI_OFFLINE": "1",
    "PI_SKIP_VERSION_CHECK": "1",
    "PI_TELEMETRY": "0",
    "LANG": "C.UTF-8",
    "SPIKE_CANARY": "canary-value-not-secret",
}
R["environment"] = {"node": subprocess.run([str(NODE_BIN / "node"), "--version"], capture_output=True,
                                           text=True).stdout.strip(),
                    "env_keys": sorted(ENV), "provider_key_vars_present": False,
                    "mock": "127.0.0.1:<ephemeral> openai-completions"}


def cli(*a, cwd=PROJ, timeout=60):
    t = time.monotonic()
    p = subprocess.run([str(NODE_BIN / "node"), str(CLI), *a], cwd=cwd, env=ENV,
                       capture_output=True, text=True, timeout=timeout)
    return p.returncode, p.stdout, p.stderr, round((time.monotonic() - t) * 1000)


class Rpc:
    def __init__(self, *extra, cwd=PROJ):
        self.p = subprocess.Popen([str(NODE_BIN / "node"), str(CLI), "--mode", "rpc", "--no-session", *extra],
                                  cwd=cwd, env=ENV, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                  stderr=subprocess.PIPE)
        self.q = queue.Queue()
        self.log = []
        threading.Thread(target=self._read, daemon=True).start()
        self.err = []
        threading.Thread(target=lambda: self.err.extend(self.p.stderr), daemon=True).start()

    def _read(self):
        for line in self.p.stdout:  # binary pipe: splits on LF only
            try:
                rec = json.loads(line)
            except ValueError:
                rec = {"type": "_unparsed"}
            rec["_t"] = time.monotonic()
            self.log.append(rec)
            self.q.put(rec)

    def send(self, cmd):
        self.p.stdin.write((json.dumps(cmd) if isinstance(cmd, dict) else cmd).encode() + b"\n")
        self.p.stdin.flush()
        return time.monotonic()

    def wait(self, pred, timeout=20):
        end = time.monotonic() + timeout
        while time.monotonic() < end:
            try:
                rec = self.q.get(timeout=max(0.01, end - time.monotonic()))
            except queue.Empty:
                break
            if pred(rec):
                return rec
        return None

    def call(self, cmd, timeout=20):
        t = self.send(cmd)
        rec = self.wait(lambda r: r.get("type") == "response" and r.get("id") == cmd.get("id"), timeout)
        return rec, (round((rec["_t"] - t) * 1000) if rec else None)

    def close(self, timeout=15):
        t = time.monotonic()
        self.p.stdin.close()
        try:
            code = self.p.wait(timeout)
        except subprocess.TimeoutExpired:
            self.p.kill()
            code = "killed_after_timeout"
        return code, round((time.monotonic() - t) * 1000)


def strip(rec):
    return {k: v for k, v in (rec or {}).items() if k != "_t"}


MOCK = ["--provider", "mock", "--model", "mock-1", "--no-approve"]

# ---------------------------------------------------------------- 1 conformance
c = {}
code, out, err, ms = cli("--version")
c["version"] = {"exit": code, "stdout": out.strip()}
code, out, err, ms = cli("--help")
flags = ["--mode", "--no-session", "--session-dir", "--provider", "--model", "--approve", "--no-approve"]
c["help_flags"] = {"exit": code, "missing": [f for f in flags if f not in out]}
docs = (PKG / "docs/rpc-commands.md").read_text()
doc_cmds = [ln[4:].strip() for ln in docs.splitlines() if ln.startswith("### ") and ln[4:].strip().isidentifier()]
r = Rpc(*MOCK)
st, ms = r.call({"id": "s1", "type": "get_state"})
c["get_state"] = {"success": st and st.get("success"), "id_echoed": bool(st), "ms": ms,
                  "data_keys": sorted((st or {}).get("data", {}).keys())}
r.send("{not json")
pr = r.wait(lambda x: x.get("type") == "response" and x.get("command") == "parse")
c["malformed_json"] = {"success": pr and pr.get("success"), "has_id": bool(pr and "id" in pr)}
uk, _ = r.call({"id": "u1", "type": "no_such_command"})
c["unknown_command"] = strip(uk)
c["doc_commands"] = {"count": len(doc_cmds)}
# Every documented command must at least be recognized (a non-"unknown" response). Commands
# that would mutate a session or run work are probed with an empty body only.
unknown = []
for i, name in enumerate(doc_cmds):
    if name in ("prompt", "steer", "follow_up", "bash", "compact", "export_html", "switch_session",
                "fork", "clone", "new_session"):
        continue
    rec, _ = r.call({"id": f"d{i}", "type": name}, timeout=10)
    if rec is None or "unknown" in str(rec.get("error", "")).lower():
        unknown.append(name)
c["doc_commands"]["probed_read_only"] = len(doc_cmds) - 10
c["doc_commands"]["unrecognized"] = unknown
c["stdin_close"] = dict(zip(("exit", "ms"), r.close()))
R["conformance"] = c

# ---------------------------------------------------------------- 2 cancellation
x = {}
r = Rpc(*MOCK)
ab, ms = r.call({"id": "a0", "type": "abort"})
x["idle_abort"] = {"success": ab and ab.get("success"), "ms": ms}
Mock.mode, Mock.disconnect_at = "slow", None
r.call({"id": "p1", "type": "prompt", "message": "stream slowly"})
first = r.wait(lambda q: q.get("type") == "message_update"
               and q.get("assistantMessageEvent", {}).get("type") == "text_delta")
t_abort = r.send({"id": "a1", "type": "abort"})
resp = r.wait(lambda q: q.get("id") == "a1")
settled = next((q for q in r.log if q.get("type") == "agent_settled"), None)
time.sleep(0.5)
st, _ = r.call({"id": "s2", "type": "get_state"})
x["mid_stream_abort"] = {
    "saw_text_delta_before_abort": bool(first),
    "abort_response_ms": round((resp["_t"] - t_abort) * 1000) if resp else None,
    "abort_success": resp and resp.get("success"),
    "agent_settled_aborted": settled and settled.get("aborted"),
    "provider_stream_closed_ms": round((Mock.disconnect_at - t_abort) * 1000) if Mock.disconnect_at else None,
    "state_after": {k: (st or {}).get("data", {}).get(k) for k in ("isStreaming", "isCompacting")},
}
Mock.mode = "short"
n0 = len(r.log)
r.call({"id": "p2", "type": "prompt", "message": "again"})
s2 = r.wait(lambda q: q.get("type") == "agent_settled", 20)
text = "".join(q.get("assistantMessageEvent", {}).get("delta", "") for q in r.log[n0:]
               if q.get("type") == "message_update")
x["prompt_after_abort"] = {"settled": bool(s2), "aborted": s2 and s2.get("aborted"), "text": text}
tag = f"{time.time_ns() % 100000}.25"
t_bash = r.send({"id": "b1", "type": "bash", "command": f"sleep {tag}"})
time.sleep(0.7)
ab2, ms2 = r.call({"id": "a2", "type": "abort_bash"})
bres = r.wait(lambda q: q.get("id") == "b1" and q.get("type") == "response", 10)
time.sleep(0.3)
left = subprocess.run(["pgrep", "-f", f"sleep {tag}"], capture_output=True, text=True).stdout.split()
x["abort_bash"] = {"abort_success": ab2 and ab2.get("success"), "abort_ms": ms2,
                   "bash_response_after_ms": round((bres["_t"] - t_bash) * 1000) if bres else None,
                   "bash_result_cancelled": (bres or {}).get("data", {}).get("cancelled"),
                   "sleep_processes_left": len(left)}
x["stdin_close"] = dict(zip(("exit", "ms"), r.close()))
R["cancellation"] = x

# ---------------------------------------------------------------- 3 security
s = {}
lock = json.loads((INSTALL / "package-lock.json").read_text())
s["install_scripts"] = sorted(k.split("node_modules/")[-1] for k, v in lock.get("packages", {}).items()
                              if v.get("hasInstallScript"))
s["package_count"] = len([k for k in lock.get("packages", {}) if k])
a = subprocess.run([str(NODE_BIN / "npm"), "audit", "--json"], cwd=INSTALL, env={**ENV, "HOME": str(HOME)},
                   capture_output=True, text=True)
try:
    s["npm_audit"] = {"exit": a.returncode,
                      "vulnerabilities": json.loads(a.stdout)["metadata"]["vulnerabilities"]}
except (ValueError, KeyError):
    s["npm_audit"] = {"exit": a.returncode, "unparsed": True}
# project trust: an extension in the working folder writes a marker when loaded
tp = WORK / "trustproj"
shutil.rmtree(tp, ignore_errors=True)
(tp / ".pi/extensions").mkdir(parents=True)
mark = tp / "extension-loaded.marker"
(tp / ".pi/extensions/probe.ts").write_text(
    "import { writeFileSync } from 'node:fs';\n"
    f"writeFileSync({json.dumps(str(mark))}, 'loaded');\n"
    "export default function () {}\n")
trust = {}
for label, extra in (("no_override", []), ("--no-approve", ["--no-approve"]), ("--approve", ["--approve"])):
    mark.unlink(missing_ok=True)
    r = Rpc("--provider", "mock", "--model", "mock-1", *extra, cwd=tp)
    rec, _ = r.call({"id": "t", "type": "get_state"})
    r.close()
    trust[label] = {"responded": bool(rec), "extension_ran": mark.exists()}
s["project_trust_rpc"] = trust
# a model tool call runs without an approval step
tool_mark = WORK / "tool-call.marker"
tool_mark.unlink(missing_ok=True)
Mock.mode, Mock.marker = "tool", str(tool_mark)
r = Rpc(*MOCK)
r.call({"id": "p3", "type": "prompt", "message": "run the tool"})
r.wait(lambda q: q.get("type") == "agent_settled", 30)
s["model_tool_call_without_approval"] = {
    "tool_started": any(q.get("type") == "tool_execution_start" for q in r.log),
    "approval_request_seen": any("extension_ui_request" == q.get("type") for q in r.log),
    "marker_created": tool_mark.exists()}
b, _ = r.call({"id": "e1", "type": "bash", "command": "printf '%s|%s|%s' \"$AI_AGENT\" \"$PI_CODING_AGENT\" "
                                                       "\"${SPIKE_CANARY:+inherited}\""})
s["bash_child_env"] = (b or {}).get("data", {}).get("output")
r.close()
R["security"] = s

# ---------------------------------------------------------------- 4 auth
au = {}
au["command_key_sent_as_bearer"] = any(q.get("auth_bearer_matches_command_key") for q in Mock.requests
                                       if q["method"] == "POST")
au["request_header_names"] = sorted({h for q in Mock.requests if q["method"] == "POST" for h in q["header_names"]})
au["non_post_requests_to_mock"] = [q["path"] for q in Mock.requests if q["method"] != "POST"]
au["agent_dir_files"] = {rel(p): oct(p.stat().st_mode & 0o777) for p in sorted(AGENT_DIR.rglob("*")) if p.is_file()}
au["auth_json_top_level_keys"] = sorted(json.loads((AGENT_DIR / "auth.json").read_text() or "{}")) \
    if (AGENT_DIR / "auth.json").exists() else None
# A fresh agent dir with no models.json and no credential at all.
AGENT2 = HOME / ".pi/agent-empty"
AGENT2.mkdir(parents=True)
ENV["PI_CODING_AGENT_DIR"] = str(AGENT2)
r = Rpc("--no-approve")
models, _ = r.call({"id": "m", "type": "get_available_models"})
au["available_models_without_credentials"] = sorted(
    f"{m.get('provider')}/{m.get('id')}" for m in (models or {}).get("data", {}).get("models", []))
st, _ = r.call({"id": "s3", "type": "get_state"})
au["default_model_without_credentials"] = (st or {}).get("data", {}).get("model")
sm, _ = r.call({"id": "sm", "type": "set_model", "provider": "anthropic", "modelId": "claude-sonnet-4-5"})
au["set_builtin_model_without_key"] = {"success": sm and sm.get("success"), "error": (sm or {}).get("error")}
n0 = len(Mock.requests)
pp, _ = r.call({"id": "p4", "type": "prompt", "message": "hi"})
r.wait(lambda q: q.get("type") == "agent_settled", 15)
au["prompt_without_key"] = {"response": strip(pp), "mock_requests": len(Mock.requests) - n0}
r.close()
au["empty_agent_dir_files"] = {rel(p): oct(p.stat().st_mode & 0o777) for p in sorted(AGENT2.rglob("*"))
                               if p.is_file()}
R["auth"] = au

srv.shutdown()
out = json.dumps(R, indent=2, default=str)
(WORK / "results.json").write_text(out)
print(out)
