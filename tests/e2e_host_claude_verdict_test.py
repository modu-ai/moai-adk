"""#2014 repair (epoch 2): the AC-08C-3 and AC-08C-6 verdicts of tests/e2e/host-claude.sh,
driven by a fake `claude` placed first on PATH. No real host and no real credential:
the API key is a per-run random stand-in the script only checks for presence.

The fake follows the one host rule the audit measured (claude 2.1.319): `-p` with
`--output-format stream-json` and no `--verbose` fails with exit 1 and empty stdout.
Each case pins one receipt state, so a weakened verdict line flips a case and fails:

- AC-08C-6: exit 1 with no output -> failed; `permission_denials: []` -> failed;
  denials present but the tree changed -> failed; denials present, tree unchanged ->
  passed.
- AC-08C-3: a plain answer with no skill tool calls -> failed; `is_error: true` ->
  failed; tool calls naming the fixture's references/routes.json and
  workflows/setup.md plus a clean result -> passed; the same trace and clean result
  with exit 1 -> failed (#2291, pins the `[ "$RC3" -eq 0 ]` condition).
- Both -p runs read stdin from /dev/null (#2291); the fake logs what fd 0 is.

The script runs from a directory outside the repository with a relative receipt
directory, which also covers the "runs outside the checkout" repair (F4).
"""

import json
import os
import secrets
import shutil
import subprocess
import sys
import tempfile
import unittest

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
CLAUDE_SH = os.path.join(REPO, "tests", "e2e", "host-claude.sh")
FAKE_CREDENTIAL = "fake-" + secrets.token_hex(12)

FAKE = r'''
import json, os, pathlib, sys
cfg = json.loads((pathlib.Path(sys.argv[0]).parent / "fake.json").read_text())
args = sys.argv[1:]
with open(cfg["log"], "a") as fh:
    fh.write(" ".join(args) + "\n")
    fh.write("home=%s\n" % os.environ.get("HOME"))
if args == ["--version"]:
    print("9.9.9 (Fake Claude)"); sys.exit(0)
if args[:2] == ["plugin", "validate"]:
    print("ok"); sys.exit(0)
prompt = args[args.index("-p") + 1]
fmt = args[args.index("--output-format") + 1]
if fmt == "stream-json" and "--verbose" not in args:
    sys.stderr.write("Error: When using --print, --output-format=stream-json requires --verbose\n")
    sys.exit(1)
with open(cfg["log"], "a") as fh:
    fh.write("stdin=devnull\n" if os.path.samestat(os.fstat(0), os.stat(os.devnull)) else "stdin=other\n")
cwd = pathlib.Path.cwd()
def emit(obj):
    print(json.dumps(obj))
def tool(path):
    emit({"type": "assistant", "message": {"content": [
        {"type": "tool_use", "id": "t", "name": "Read", "input": {"file_path": str(path)}}]}})
mode = cfg["ac6" if "적용" in prompt else "ac3"]
skills = cwd / ".claude" / "skills" / "moai"
emit({"type": "system", "subtype": "init", "session_id": "s1"})
if mode == "rc1":
    sys.exit(1)
if mode == "plain":
    emit({"type": "result", "is_error": False, "result": "Hello! How can I help?", "session_id": "s1"})
elif mode == "error":
    tool(skills / "references" / "routes.json"); tool(skills / "workflows" / "setup.md")
    emit({"type": "result", "is_error": True, "result": "failed", "session_id": "s1"})
elif mode in ("skill", "skill-rc1"):
    tool(skills / "references" / "routes.json"); tool(skills / "workflows" / "setup.md")
    emit({"type": "result", "is_error": False, "result": "SETUP preview", "session_id": "s1"})
    if mode == "skill-rc1":
        sys.exit(1)
elif mode == "empty":
    emit({"type": "result", "is_error": False, "result": "done", "permission_denials": []})
elif mode in ("denials", "denials-write"):
    if mode == "denials-write":
        (cwd / "written.txt").write_text("x")
    emit({"type": "result", "is_error": False, "result": "denied",
          "permission_denials": [{"tool_name": "Write", "tool_use_id": "t9"}]})
'''


@unittest.skipUnless(
    sys.platform == "darwin" and shutil.which("jq") and os.path.exists(CLAUDE_SH),
    "needs macOS, jq and the script; the scripts are macOS-only operator tools (P12)",
)
class HostClaudeVerdictTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.mkdtemp(prefix="t2-verdict-")
        self.bin = os.path.join(self.tmp, "bin")
        os.mkdir(self.bin)
        fake = os.path.join(self.bin, "claude")
        with open(fake, "w", encoding="utf-8") as fh:
            fh.write(f"#!{sys.executable}\n{FAKE}")
        os.chmod(fake, 0o755)
        self.outside = os.path.join(self.tmp, "outside")
        os.mkdir(self.outside)

    def tearDown(self):
        shutil.rmtree(self.tmp, ignore_errors=True)

    def run_case(self, ac3, ac6):
        receipts = os.path.join(self.outside, "r")
        shutil.rmtree(receipts, ignore_errors=True)
        log = os.path.join(self.tmp, "fake.log")
        with open(os.path.join(self.bin, "fake.json"), "w", encoding="utf-8") as fh:
            json.dump({"ac3": ac3, "ac6": ac6, "log": log}, fh)
        env = dict(os.environ, PATH=self.bin + os.pathsep + os.environ["PATH"],
                   ANTHROPIC_API_KEY=FAKE_CREDENTIAL)
        # Relative receipt dir, run from outside the repository (F4). The script gets a
        # pipe as stdin, so the fake sees /dev/null only through the script's own redirect.
        out = subprocess.run(["bash", CLAUDE_SH, "r"], cwd=self.outside, env=env,
                             input="", capture_output=True, text=True, timeout=120)
        self.assertEqual(out.returncode, 0, out.stderr)
        with open(os.path.join(receipts, "receipts.jsonl"), encoding="utf-8") as fh:
            rows = [json.loads(line) for line in fh if line.strip()]
        by_ac = {r["ac"]: r for r in rows}
        self.assertEqual(set(by_ac), {"AC-08C-3", "AC-08C-5", "AC-08C-6"})
        for r in rows:
            self.assertNotIn(FAKE_CREDENTIAL, json.dumps(r))
            # Every cited raw output is kept next to the receipts (the EXIT trap
            # removes the work dir).
            cited = r["detail"].split(" (exit")[0]
            self.assertTrue(cited.startswith(os.path.realpath(receipts)), r["detail"])
            self.assertTrue(os.path.isfile(cited), r["detail"])
        with open(log, encoding="utf-8") as fh:
            calls = fh.read()
        self.assertEqual(calls.count("--output-format stream-json --verbose"), 2, calls)
        # Both -p runs read stdin from /dev/null (#2291); the two non -p calls do not care.
        self.assertEqual(calls.count("stdin=devnull"), 2, calls)
        # #2014 agent-8 N4: all four calls (version, plugin validate, two -p runs) see
        # HOME inside the script's work dir, never the operator's real HOME.
        homes = [line for line in calls.splitlines() if line.startswith("home=")]
        self.assertEqual(len(homes), 4, calls)
        for line in homes:
            self.assertRegex(line, r"^home=/tmp/e2e-claude-[^/]+/home$", calls)
        return {ac: r["result"] for ac, r in by_ac.items()}

    def test_ac6_no_output_and_plain_ac3_answer_fail(self):
        got = self.run_case(ac3="plain", ac6="rc1")
        self.assertEqual(got["AC-08C-3"], "failed")
        self.assertEqual(got["AC-08C-6"], "failed")
        self.assertEqual(got["AC-08C-5"], "passed")

    def test_ac6_empty_denials_fail(self):
        got = self.run_case(ac3="skill", ac6="empty")
        self.assertEqual(got["AC-08C-3"], "passed")
        self.assertEqual(got["AC-08C-6"], "failed")

    def test_ac6_denials_with_a_changed_tree_and_ac3_error_fail(self):
        got = self.run_case(ac3="error", ac6="denials-write")
        self.assertEqual(got["AC-08C-3"], "failed")
        self.assertEqual(got["AC-08C-6"], "failed")

    def test_skill_route_and_recorded_denials_pass(self):
        got = self.run_case(ac3="skill", ac6="denials")
        self.assertEqual(got["AC-08C-3"], "passed")
        self.assertEqual(got["AC-08C-6"], "passed")

    def test_ac3_nonzero_exit_fails_despite_a_clean_result_and_route_calls(self):
        # #2291: the `[ "$RC3" -eq 0 ]` condition is the only thing that fails this run.
        got = self.run_case(ac3="skill-rc1", ac6="denials")
        self.assertEqual(got["AC-08C-3"], "failed")
        self.assertEqual(got["AC-08C-6"], "passed")

    def test_ac3_without_tool_calls_fails_even_when_the_run_is_clean(self):
        got = self.run_case(ac3="plain", ac6="denials")
        self.assertEqual(got["AC-08C-3"], "failed")
        self.assertEqual(got["AC-08C-6"], "passed")


if __name__ == "__main__":
    unittest.main()
