"""Offline checks for tools/host_live_harness.py (I8 #2245).

No real host is spawned. Fake `claude` / `codex` executables stand in for the hosts: the
"good" fakes follow the documented loading rules (AGENTS.md unless a CLAUDE.md variant
exists; AGENTS.override.md before AGENTS.md; project skills; project SessionStart hooks),
the "bad" fakes ignore every instruction file, so each verdict branch is exercised both
ways. These tests prove the harness logic, not any host behaviour.
"""
import json
import os
import pathlib
import secrets
import shutil
import subprocess
import sys
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parent.parent
HARNESS = ROOT / "tools/host_live_harness.py"
CHECKER = ROOT / "tools/check_support_matrix.py"
REQUIRED = ["ac", "os", "arch", "host", "host_exact_version", "mode", "skill_digest",
            "moai_cli_sha", "result", "ran_at"]
SCRIPTED = {"AC-I8C-1", "AC-I8C-2", "AC-I8C-3", "AC-I8C-4", "AC-I8X-1", "AC-I8X-2", "AC-I8X-3"}
NEVER_RUN = {"AC-I8C-5", "AC-I8X-4", "AC-I8X-5"}
# A per-run random stand-in for a credential value; the fakes echo it back so the
# redaction path is exercised. It is not a real credential.
FAKE_CREDENTIAL = "fake-" + secrets.token_hex(12)

FAKE_CLAUDE = r'''
import json, os, pathlib, subprocess, sys
cfg = json.loads((pathlib.Path(sys.argv[0]).parent / "fake.json").read_text())
with open(cfg["log"], "a") as fh:
    fh.write("claude " + " ".join(sys.argv[1:]) + "\n")
if sys.argv[1:] == ["--version"]:
    print("9.9.9 (Fake Claude)"); sys.exit(0)
assert os.environ["HOME"] != cfg["real_home"], "HOME not isolated"
assert os.environ["CLAUDE_CONFIG_DIR"].startswith(os.environ["HOME"]), "config dir not isolated"
mode, cwd = cfg["mode"], pathlib.Path.cwd()
prompt = sys.argv[sys.argv.index("-p") + 1]
text = "nothing"
if mode == "good":
    settings = cwd / ".claude/settings.json"
    if settings.is_file():
        for group in json.loads(settings.read_text()).get("hooks", {}).get("SessionStart", []):
            for hook in group["hooks"]:
                subprocess.run(["sh", "-c", hook["command"]], check=True)
    if prompt.startswith("/"):
        body = (cwd / ".claude/skills" / prompt[1:] / "SKILL.md").read_text()
        text = body.split("reply with exactly: ")[1].strip()
    elif (cwd / "AGENTS.md").is_file() and not any(
            (cwd / n).exists() for n in ("CLAUDE.md", ".claude/CLAUDE.md", "CLAUDE.local.md")):
        for line in (cwd / "AGENTS.md").read_text().splitlines():
            if line.startswith("MOAI_PROBE_MARKER="):
                text = line.split("=", 1)[1]
    text += " key=" + os.environ.get("ANTHROPIC_API_KEY", "")
print(json.dumps({"type": "result", "is_error": False, "result": text}))
'''

FAKE_CODEX = r'''
import json, os, pathlib, sys
cfg = json.loads((pathlib.Path(sys.argv[0]).parent / "fake.json").read_text())
with open(cfg["log"], "a") as fh:
    fh.write("codex " + " ".join(sys.argv[1:]) + "\n")
if sys.argv[1:] == ["--version"]:
    print("codex-cli 9.9.9"); sys.exit(0)
assert os.environ["HOME"] != cfg["real_home"], "HOME not isolated"
assert os.environ["CODEX_HOME"].startswith(os.environ["HOME"]), "CODEX_HOME not isolated"
mode, cwd = cfg["mode"], pathlib.Path.cwd()
args = sys.argv[1:]
out, prompt = args[args.index("-o") + 1], args[-1]
text = "nothing"
if mode == "good":
    if prompt.startswith("$"):
        body = (cwd / ".agents/skills" / prompt[1:] / "SKILL.md").read_text()
        text = body.split("reply with exactly: ")[1].strip()
    else:
        for name in ("AGENTS.override.md", "AGENTS.md"):
            if (cwd / name).is_file():
                for line in (cwd / name).read_text().splitlines():
                    if line.startswith("MOAI_PROBE_MARKER="):
                        text = line.split("=", 1)[1]
                break
pathlib.Path(out).write_text(text)
print(json.dumps({"type": "thread.started"}))
if mode != "no-turn":
    print(json.dumps({"type": "turn.completed"}))
'''


class HarnessTest(unittest.TestCase):
    def setUp(self):
        self.tmp = pathlib.Path(tempfile.mkdtemp(prefix="i8-test-"))
        self.out = self.tmp / "receipts"
        self.log = self.tmp / "fake.log"
        self.bins = {}
        for name, body in (("claude", FAKE_CLAUDE), ("codex", FAKE_CODEX)):
            path = self.tmp / f"fake-{name}"
            path.write_text(f"#!{sys.executable}\n{body}")
            path.chmod(0o755)
            self.bins[name] = path

    def tearDown(self):
        shutil.rmtree(self.tmp, ignore_errors=True)

    def run_harness(self, *extra, mode="good", creds=True, bins=True):
        env = {k: v for k, v in os.environ.items() if k not in ("ANTHROPIC_API_KEY", "CODEX_API_KEY")}
        # The harness passes the hosts a minimal environment, so the fakes read their mode
        # from a file beside them rather than from the environment.
        (self.tmp / "fake.json").write_text(json.dumps(
            {"mode": mode, "log": str(self.log), "real_home": os.environ.get("HOME", "")}))
        if creds:
            env.update(ANTHROPIC_API_KEY=FAKE_CREDENTIAL, CODEX_API_KEY=FAKE_CREDENTIAL)
        argv = [sys.executable, "-I", str(HARNESS), "--out", str(self.out), "--repo", str(ROOT)]
        if bins:
            argv += ["--bin", f"claude={self.bins['claude']}", "--bin", f"codex={self.bins['codex']}"]
        else:
            argv += ["--bin", f"claude={self.tmp / 'missing-claude'}", "--bin", f"codex={self.tmp / 'missing-codex'}"]
        return subprocess.run(argv + list(extra), capture_output=True, text=True, env=env, cwd=ROOT)

    def receipts(self):
        lines = (self.out / "receipts.jsonl").read_text(encoding="utf-8").splitlines()
        return {r["ac"]: r for r in map(json.loads, lines)}

    def test_dry_run_spawns_nothing_and_records_not_run(self):
        done = self.run_harness()
        self.assertEqual(done.returncode, 0, done.stderr)
        receipts = self.receipts()
        self.assertEqual(len(receipts), 10)
        for r in receipts.values():
            for field in REQUIRED:
                self.assertIn(field, r)
            self.assertEqual(r["result"], "not_run")
            self.assertIsNone(r["host_exact_version"])
        for ac in SCRIPTED:
            self.assertIn("dry-run: not executed:", receipts[ac]["detail"])
            self.assertNotIn(str(self.tmp), receipts[ac]["detail"], "absolute path in a receipt")
            self.assertNotIn(str(ROOT), receipts[ac]["detail"], "absolute path in a receipt")
        self.assertFalse(self.log.exists(), "a host was spawned in dry-run")
        self.assertIn("결정 대기", receipts["AC-I8C-5"]["detail"])
        self.assertIn("manual only", receipts["AC-I8X-4"]["detail"])

    def test_live_good_hosts_pass_every_scripted_item(self):
        done = self.run_harness("--live")
        self.assertEqual(done.returncode, 0, done.stderr)
        receipts = self.receipts()
        for ac in SCRIPTED:
            self.assertEqual(receipts[ac]["result"], "passed", (ac, receipts[ac]["detail"]))
        for ac in NEVER_RUN:
            self.assertEqual(receipts[ac]["result"], "not_run", ac)
        self.assertEqual(receipts["AC-I8C-1"]["host_exact_version"], "9.9.9 (Fake Claude)")
        self.assertTrue(receipts["AC-I8C-1"]["detail"].startswith("AC-I8C-1.out ("), receipts["AC-I8C-1"]["detail"])
        self.assertEqual(self.log.read_text().count("--version"), 2, "version probed once per host")

    def test_live_bad_hosts_fail_every_scripted_item_except_the_negative_control(self):
        self.run_harness("--live", mode="bad")
        receipts = self.receipts()
        # AC-I8C-2 expects the AGENTS.md marker to be ABSENT; a host that reads nothing passes it.
        for ac in SCRIPTED - {"AC-I8C-2"}:
            self.assertEqual(receipts[ac]["result"], "failed", (ac, receipts[ac]["detail"]))
        self.assertEqual(receipts["AC-I8C-2"]["result"], "passed")

    def test_codex_without_turn_completed_fails(self):
        self.run_harness("--live", "--only", "AC-I8X-1", mode="no-turn")
        receipt = self.receipts()["AC-I8X-1"]
        self.assertEqual(receipt["result"], "failed")
        self.assertIn("turn.completed", receipt["detail"])

    def test_missing_credentials_or_binary_is_not_run(self):
        self.run_harness("--live", "--only", "AC-I8C-1", "--only", "AC-I8X-1", creds=False)
        for r in self.receipts().values():
            self.assertEqual(r["result"], "not_run")
            self.assertIn("not set; not executed", r["detail"])
        self.assertFalse(self.log.exists())
        (self.out / "receipts.jsonl").unlink()
        self.run_harness("--live", "--only", "AC-I8C-1", bins=False)
        receipt = self.receipts()["AC-I8C-1"]
        self.assertEqual(receipt["result"], "not_run")
        self.assertIn("binary not found", receipt["detail"])

    def test_credential_value_never_kept(self):
        self.run_harness("--live")
        for path in self.out.iterdir():
            self.assertNotIn(FAKE_CREDENTIAL, path.read_text(encoding="utf-8"), path.name)
        self.assertIn("[REDACTED:ANTHROPIC_API_KEY]", (self.out / "AC-I8C-1.out").read_text())

    def test_rows_are_accepted_by_the_support_matrix_checker(self):
        for mode in ("--live", None):
            shutil.rmtree(self.out, ignore_errors=True)
            self.run_harness(*([mode] if mode else []))
            receipts_dir = self.tmp / "e2e-receipts"
            shutil.rmtree(receipts_dir, ignore_errors=True)
            shutil.copytree(ROOT / "docs/e2e-receipts", receipts_dir)
            with (receipts_dir / "receipts.jsonl").open("a", encoding="utf-8") as fh:
                fh.write((self.out / "receipts.jsonl").read_text(encoding="utf-8"))
            matrix = self.tmp / "support-matrix.md"
            matrix.write_text((ROOT / "docs/support-matrix.md").read_text(encoding="utf-8") + "\n"
                              + (self.out / "rows.md").read_text(encoding="utf-8"), encoding="utf-8")
            done = subprocess.run([sys.executable, "-I", str(CHECKER), str(matrix), str(receipts_dir)],
                                  capture_output=True, text=True)
            self.assertEqual(done.returncode, 0, (mode, done.stdout, done.stderr))
            rows = (self.out / "rows.md").read_text(encoding="utf-8")
            if mode:
                self.assertEqual(rows.count("| passed | receipts.jsonl |"), len(SCRIPTED))
            else:
                self.assertNotIn("| passed |", rows)

    def test_usage_errors_exit_2_and_bad_repo_exits_1(self):
        def code(*argv, cwd=ROOT):
            return subprocess.run([sys.executable, "-I", str(HARNESS), *argv], capture_output=True, cwd=cwd).returncode
        self.assertEqual(code(), 2)
        self.assertEqual(code("--out", str(self.out), "--only", "AC-NOPE-1"), 2)
        self.assertEqual(code("--out", str(self.out), "--bin", "vim=/x"), 2)
        self.assertEqual(code("--out", str(self.out), "--repo", str(self.tmp)), 1)


if __name__ == "__main__":
    unittest.main()
