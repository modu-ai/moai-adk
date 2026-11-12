"""#2014 agent-8 N3/N4: the AC-08C-4 verdict of tests/e2e/host-codex.sh, driven by a
fake `codex` placed first on PATH. No real host and no real credential: the API key is
a per-run random stand-in the script only checks for presence.

The verdict passes only when ALL of these hold, and each case below breaks exactly one
of them, so weakening any verdict line flips a case and fails the test:

- exit 0 (`route-rc1`);
- a `turn.completed` event (`route-nocomplete`) and no `turn.failed` (`route-failed`)
  or `error` (`route-error`) event;
- a non-empty last-message file (`route-nolast`);
- `command_execution` items naming the skill's references/routes.json and
  workflows/setup.md (`plain`, `route-one`, and `route-message-only`, where the paths
  appear only in an agent message).

The event and item shapes the fake emits follow the Codex non-interactive docs' sample
lines. That the real host reads the skill files through such items is an UNVERIFIED
assumption (see the script comment); these cases pin the script's logic, not the host.

Every codex call (the version probe and `exec`) must see HOME and CODEX_HOME inside the
script's temporary work dir (N4), and `exec` must read stdin from /dev/null.
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
CODEX_SH = os.path.join(REPO, "tests", "e2e", "host-codex.sh")
FAKE_CREDENTIAL = "fake-" + secrets.token_hex(12)

FAKE = r'''
import json, os, pathlib, sys
cfg = json.loads((pathlib.Path(sys.argv[0]).parent / "fake.json").read_text())
args = sys.argv[1:]
with open(cfg["log"], "a") as fh:
    fh.write("call " + " ".join(args) + "\n")
    fh.write("home=%s codex_home=%s\n" % (os.environ.get("HOME"), os.environ.get("CODEX_HOME")))
if args == ["--version"]:
    print("codex-cli 9.9.9-fake"); sys.exit(0)
assert args[0] == "exec", args
with open(cfg["log"], "a") as fh:
    fh.write("stdin=devnull\n" if os.path.samestat(os.fstat(0), os.stat(os.devnull)) else "stdin=other\n")
last = args[args.index("-o") + 1]
mode = cfg["mode"]
def emit(obj):
    print(json.dumps(obj))
def cmd(i, path):
    item = {"id": "item_%d" % i, "type": "command_execution",
            "command": "bash -lc 'cat %s'" % path, "status": "in_progress"}
    emit({"type": "item.started", "item": item})
    done = dict(item, status="completed")
    emit({"type": "item.completed", "item": done})
skill = ".agents/skills/moai"
emit({"type": "thread.started", "thread_id": "t-1"})
emit({"type": "turn.started"})
if mode == "route-message-only":
    emit({"type": "item.completed", "item": {"id": "item_1", "type": "agent_message",
          "text": "read %s/references/routes.json and %s/workflows/setup.md" % (skill, skill)}})
elif mode != "plain":
    cmd(1, skill + "/references/routes.json")
    if mode != "route-one":
        cmd(2, skill + "/workflows/setup.md")
if mode == "route-error":
    emit({"type": "error", "message": "stream interrupted"})
if mode == "route-failed":
    emit({"type": "turn.failed", "error": {"message": "failed"}})
emit({"type": "item.completed", "item": {"id": "item_9", "type": "agent_message",
      "text": "SETUP preview"}})
if mode != "route-nocomplete":
    emit({"type": "turn.completed", "usage": {"input_tokens": 1, "output_tokens": 1}})
pathlib.Path(last).write_text("" if mode == "route-nolast" else "SETUP preview\n")
sys.exit(1 if mode == "route-rc1" else 0)
'''


@unittest.skipUnless(
    sys.platform == "darwin" and shutil.which("jq") and os.path.exists(CODEX_SH),
    "needs macOS, jq and the script; the scripts are macOS-only operator tools (P12)",
)
class HostCodexVerdictTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.mkdtemp(prefix="t2-codex-verdict-")
        self.bin = os.path.join(self.tmp, "bin")
        os.mkdir(self.bin)
        fake = os.path.join(self.bin, "codex")
        with open(fake, "w", encoding="utf-8") as fh:
            fh.write(f"#!{sys.executable}\n{FAKE}")
        os.chmod(fake, 0o755)
        self.outside = os.path.join(self.tmp, "outside")
        os.mkdir(self.outside)

    def tearDown(self):
        shutil.rmtree(self.tmp, ignore_errors=True)

    def run_case(self, mode):
        receipts = os.path.join(self.outside, "r")
        log = os.path.join(self.tmp, "fake.log")
        with open(os.path.join(self.bin, "fake.json"), "w", encoding="utf-8") as fh:
            json.dump({"mode": mode, "log": log}, fh)
        env = dict(os.environ, PATH=self.bin + os.pathsep + os.environ["PATH"],
                   CODEX_API_KEY=FAKE_CREDENTIAL)
        # Relative receipt dir, run from outside the repository; stdin is a pipe, so
        # the fake sees /dev/null only through the script's own redirect.
        out = subprocess.run(["bash", CODEX_SH, "r"], cwd=self.outside, env=env,
                             input="", capture_output=True, text=True, timeout=120)
        self.assertEqual(out.returncode, 0, out.stderr)
        with open(os.path.join(receipts, "receipts.jsonl"), encoding="utf-8") as fh:
            rows = [json.loads(line) for line in fh if line.strip()]
        self.assertEqual([r["ac"] for r in rows], ["AC-08C-4"])
        row = rows[0]
        self.assertNotIn(FAKE_CREDENTIAL, json.dumps(row))
        self.assertEqual(row["host_exact_version"], "codex-cli 9.9.9-fake")
        cited = row["detail"].split(" (exit")[0]
        self.assertTrue(cited.startswith(os.path.realpath(receipts)), row["detail"])
        self.assertTrue(os.path.isfile(cited), row["detail"])
        with open(log, encoding="utf-8") as fh:
            lines = fh.read().splitlines()
        calls = [line for line in lines if line.startswith("call ")]
        self.assertEqual(len(calls), 2, lines)
        # N4: both calls run with HOME and CODEX_HOME in the script's work dir.
        homes = [line for line in lines if line.startswith("home=")]
        self.assertEqual(len(homes), 2, lines)
        for line in homes:
            home = line.split(" ")[0][len("home="):]
            codex_home = line.split(" ")[1][len("codex_home="):]
            self.assertRegex(home, r"^/tmp/e2e-codex-[^/]+/home$", line)
            self.assertEqual(codex_home, home + "/.codex", line)
        self.assertEqual(lines.count("stdin=devnull"), 1, lines)
        return row["result"]

    def test_route_commands_on_a_clean_completed_run_pass(self):
        self.assertEqual(self.run_case("route"), "passed")

    def test_plain_answer_without_route_commands_fails(self):
        self.assertEqual(self.run_case("plain"), "failed")

    def test_only_one_route_file_fails(self):
        self.assertEqual(self.run_case("route-one"), "failed")

    def test_paths_named_only_in_a_message_fail(self):
        self.assertEqual(self.run_case("route-message-only"), "failed")

    def test_nonzero_exit_fails_despite_route_commands(self):
        self.assertEqual(self.run_case("route-rc1"), "failed")

    def test_missing_turn_completed_fails(self):
        self.assertEqual(self.run_case("route-nocomplete"), "failed")

    def test_turn_failed_event_fails(self):
        self.assertEqual(self.run_case("route-failed"), "failed")

    def test_error_event_fails(self):
        self.assertEqual(self.run_case("route-error"), "failed")

    def test_empty_last_message_fails(self):
        self.assertEqual(self.run_case("route-nolast"), "failed")


if __name__ == "__main__":
    unittest.main()
