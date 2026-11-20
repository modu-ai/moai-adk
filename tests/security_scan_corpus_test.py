"""tools/security_scan_corpus.py (SPEC-20261120-security-scan-notify REQ-12): the report of a small
fixture repository with the real built binary (`target/debug/moai-cli`, or
`$CARGO_TARGET_DIR/debug/moai-cli`), the skip reasons, the untouched repository and the usage
exits. The numbers of a real corpus are reported by the tool, never asserted."""

import json
import os
import pathlib
import subprocess
import tempfile
import unittest

REPO_ROOT = pathlib.Path(__file__).resolve().parents[1]
TOOL = REPO_ROOT / "tools/security_scan_corpus.py"


def built_binary():
    target = os.environ.get("CARGO_TARGET_DIR")
    binary = (pathlib.Path(target) if target else REPO_ROOT / "target") / "debug" / "moai-cli"
    if not binary.exists():
        raise AssertionError(f"moai-cli not built at {binary}; run cargo build first")
    return binary


def git(repo, *args):
    env = dict(os.environ, GIT_CONFIG_GLOBAL="/dev/null", GIT_CONFIG_SYSTEM="/dev/null")
    subprocess.run(["git", "-C", str(repo), *args], check=True, capture_output=True, env=env)


def run_tool(*args):
    return subprocess.run(
        ["python3", "-I", str(TOOL), *args], capture_output=True, text=True
    )


class CorpusToolTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.repo = pathlib.Path(self.tmp.name) / "repo"
        self.repo.mkdir()
        git(self.repo, "init", "-q", "-b", "main")
        (self.repo / "a.py").write_text("result = eval(text)\n")
        (self.repo / "b.py").write_text("os.system(cmd)\nvalue = eval(other)\n")
        (self.repo / "c.md").write_text("The eval function is discussed here.\n")
        (self.repo / "bin.dat").write_bytes(b"eval(x)\0\1\2")
        (self.repo / "big.txt").write_text("eval(x)\n" + "a" * (1024 * 1024))
        (self.repo / "latin.txt").write_bytes(b"eval(x) \xe9\n")
        git(self.repo, "add", "-A")
        git(
            self.repo,
            "-c", "user.name=t", "-c", "user.email=t@example.invalid",
            "commit", "-q", "-m", "fixture",
        )

    def test_report_counts_classes_files_and_skips_without_touching_the_repo(self):
        run = run_tool("--bin", str(built_binary()), "--repo", str(self.repo), "--jobs", "2")
        self.assertEqual(run.returncode, 0, run.stderr)
        report = json.loads(run.stdout)
        self.assertEqual(report["files_fed"], 3)
        self.assertEqual(report["files_with_findings"], 2)
        self.assertEqual(
            report["files_skipped"], {"binary": 1, "not utf-8": 1, "over 1 MiB": 1}
        )
        self.assertEqual(
            report["files_per_class"], {"code-injection-eval": 2, "command-injection": 1}
        )
        self.assertEqual(
            report["top_files_per_class"]["code-injection-eval"],
            [
                {"path": "b.py", "findings_on_file": 2},
                {"path": "a.py", "findings_on_file": 1},
            ],
        )
        # The repository is as it was: nothing written, no state folder made.
        status = subprocess.run(
            ["git", "-C", str(self.repo), "status", "--porcelain", "--ignored"],
            capture_output=True, text=True,
        )
        self.assertEqual(status.stdout, "")

    def test_limit_and_top_bound_the_report(self):
        run = run_tool(
            "--bin", str(built_binary()), "--repo", str(self.repo),
            "--jobs", "1", "--top", "1", "--limit", "2",
        )
        self.assertEqual(run.returncode, 0, run.stderr)
        report = json.loads(run.stdout)
        # The first two tracked files in git order are a.py and b.py.
        self.assertEqual(report["files_fed"], 2)
        self.assertEqual(
            report["top_files_per_class"]["code-injection-eval"],
            [{"path": "b.py", "findings_on_file": 2}],
        )

    def test_usage_and_unusable_inputs(self):
        self.assertEqual(run_tool().returncode, 2)
        self.assertEqual(run_tool("--bin", "x", "--repo", "y", "--jobs", "0").returncode, 2)
        missing = run_tool("--bin", "/nonexistent/moai", "--repo", str(self.repo))
        self.assertEqual(missing.returncode, 1)
        not_repo = run_tool("--bin", str(built_binary()), "--repo", self.tmp.name)
        self.assertEqual(not_repo.returncode, 1)


if __name__ == "__main__":
    unittest.main()
