"""tools/self_check.sh (card C-77): the machine items of the pre-PR self-check from one command.

The script runs the machine-checkable items of skills/moai/references/internal/self-check.md
in the working folder, keeps going past a failure, prints one `self-check: <item> ...` line per
item with its exit code and ends with one summary line. These tests pin the item list and its
commands (whole `--bin` runs, never filtered), the keep-going rule, the skip rule, the package
file bound at 65,536 / 65,537 bytes and the summary line. Real runs use stub `cargo` and
`python3` in a temporary git repository with a temporary HOME, so no cargo build runs here.
"""
import re
import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SCRIPT = ROOT / "tools" / "self_check.sh"
REFERENCE = ROOT / "skills" / "moai" / "references" / "internal" / "self-check.md"
PROFILE = ROOT / "src" / "profile.rs"

# The items in order with the exact command each runs (package-size is built in).
ITEMS = [
    ("bin-moai", "cargo test --offline --locked --bin moai"),
    ("bin-moai-cli", "cargo test --offline --locked --bin moai-cli"),
    ("clippy", "cargo clippy --offline --locked --all-targets -- -D warnings"),
    ("fmt", "cargo fmt --check"),
    ("git-env-scrub", "cargo test --offline --locked --test git_env_scrub"),
    ("home-isolation", "cargo test --offline --locked --test home_isolation_scan"),
    ("ac-command-lint", "cargo test --offline --locked --test ac_command_lint"),
    ("package-inspect",
     "cargo run --offline --locked --quiet --bin moai -- package-inspect skills/moai"),
    ("package-size", None),
    ("manifest", "python3 -I tools/source_manifest.py check"),
]
LABELS = [label for label, _ in ITEMS]
COMMANDS = [cmd for _, cmd in ITEMS if cmd]
ITEM = re.compile(r"^self-check: (?P<label>[a-z-]+) (?P<rest>.*)$")
LAST = re.compile(
    r"^self-check: head=(?P<head>[0-9a-f]{40}|unknown) result=(?P<result>pass|fail) "
    r"(?P<dirty>dirty=1 )?(?P<dry>dry-run=1 )?failed=(?P<failed>none|[a-z,-]+) "
    r"logs=(?P<logs>\S+)$")


class FakeRepo:
    """A temporary git repository with a copy of the script, the files each item needs and
    stub `cargo` / `python3` on PATH that record their arguments."""

    def __init__(self):
        self.tmp = Path(tempfile.mkdtemp(prefix="self-check-"))
        self.repo = self.tmp / "repo"
        (self.repo / "tools").mkdir(parents=True)
        shutil.copy2(SCRIPT, self.repo / "tools" / "self_check.sh")
        (self.repo / "tools" / "source_manifest.py").write_text("# stub\n")
        (self.repo / "Cargo.toml").write_text("[package]\nname = \"x\"\n")
        (self.repo / "tests").mkdir()
        for name in ("git_env_scrub", "home_isolation_scan", "ac_command_lint"):
            (self.repo / "tests" / f"{name}.rs").write_text("\n")
        (self.repo / "skills" / "moai").mkdir(parents=True)
        (self.repo / "skills" / "moai" / "SKILL.md").write_text("# moai\n")
        (self.repo / ".gitignore").write_text("target/\n")
        self.calls = self.tmp / "calls"
        self.bin = self.tmp / "bin"
        self.bin.mkdir()
        for name in ("cargo", "python3"):
            stub = self.bin / name
            stub.write_text(
                "#!/bin/sh\n"
                f"echo \"{name} $*\" >> '{self.calls}'\n"
                "echo \"stub output of {name} $*\"\n"
                "case \"$*\" in *\"${STUB_FAIL:-@none@}\"*) exit 3;; esac\n"
                "exit 0\n".replace("{name}", name))
            stub.chmod(0o755)
        self.git("init", "-q")
        self.git("add", "-A")
        self.git("commit", "-q", "-m", "one")

    def env(self, **extra):
        env = {"PATH": f"{self.bin}:/usr/bin:/bin", "HOME": str(self.tmp), "LC_ALL": "C",
               "GIT_CONFIG_NOSYSTEM": "1", "GIT_AUTHOR_NAME": "t",
               "GIT_AUTHOR_EMAIL": "t@example.invalid", "GIT_COMMITTER_NAME": "t",
               "GIT_COMMITTER_EMAIL": "t@example.invalid"}
        env.update(extra)
        return env

    def git(self, *args):
        return subprocess.run(["git", *args], cwd=self.repo, env=self.env(), check=True,
                              capture_output=True, text=True).stdout.strip()

    def run(self, *args, **extra):
        # Started from outside the repository: the script must find its own root.
        return subprocess.run(["bash", str(self.repo / "tools" / "self_check.sh"), *args],
                              cwd=self.tmp, env=self.env(**extra), capture_output=True,
                              text=True, timeout=60)

    def recorded(self):
        return self.calls.read_text().splitlines() if self.calls.exists() else []

    def close(self):
        shutil.rmtree(self.tmp, ignore_errors=True)


def items(stdout):
    """{label: rest} of the item lines, in order, and the last line."""
    lines = [line for line in stdout.splitlines() if line.strip()]
    found = []
    for line in lines[:-1]:
        m = ITEM.match(line)
        if m:
            found.append((m["label"], m["rest"]))
    return found, (lines[-1] if lines else "")


class SelfCheckTest(unittest.TestCase):
    def setUp(self):
        self.repo = FakeRepo()
        self.addCleanup(self.repo.close)

    def test_every_item_runs_in_order_with_one_line_and_the_run_passes(self):
        proc = self.repo.run()
        self.assertEqual(proc.returncode, 0, proc.stdout + proc.stderr)
        found, last = items(proc.stdout)
        self.assertEqual([label for label, _ in found], LABELS)
        for label, rest in found:
            self.assertTrue(rest.startswith("exit=0 "), f"{label}: {rest}")
        self.assertEqual(self.repo.recorded(), COMMANDS)
        m = LAST.match(last)
        self.assertIsNotNone(m, last)
        self.assertEqual(m["head"], self.repo.git("rev-parse", "HEAD"))
        self.assertEqual((m["result"], m["dirty"], m["dry"], m["failed"]),
                         ("pass", None, None, "none"))
        # Each item's output goes to its own log, not to the screen.
        self.assertNotIn("stub output", proc.stdout)
        log = self.repo.repo / "target" / "self-check" / "clippy.log"
        self.assertIn("stub output of cargo clippy", log.read_text())

    def test_whole_bin_runs_carry_no_test_name_filter(self):
        bins = [cmd for cmd in COMMANDS if " --bin moai" in cmd and cmd.startswith("cargo test")]
        self.assertEqual(bins, ["cargo test --offline --locked --bin moai",
                                "cargo test --offline --locked --bin moai-cli"])
        self.repo.run()
        recorded = self.repo.recorded()
        for cmd in bins:
            self.assertIn(cmd, recorded)

    def test_a_failing_item_is_reported_and_every_later_item_still_runs(self):
        proc = self.repo.run(STUB_FAIL="clippy")
        self.assertEqual(proc.returncode, 1, proc.stdout + proc.stderr)
        found, last = items(proc.stdout)
        self.assertEqual([label for label, _ in found], LABELS)
        rest = dict(found)
        self.assertTrue(rest["clippy"].startswith("exit=3 "), rest["clippy"])
        self.assertTrue(rest["manifest"].startswith("exit=0 "), rest["manifest"])
        self.assertEqual(self.repo.recorded(), COMMANDS)
        m = LAST.match(last)
        self.assertEqual((m["result"], m["failed"]), ("fail", "clippy"))

    def test_two_failures_are_both_named(self):
        proc = self.repo.run(STUB_FAIL="--bin moai")
        self.assertEqual(proc.returncode, 1)
        m = LAST.match(items(proc.stdout)[1])
        # `--bin moai` matches bin-moai, bin-moai-cli and package-inspect.
        self.assertEqual(m["failed"], "bin-moai,bin-moai-cli,package-inspect")

    def test_package_file_bound_passes_at_65536_and_fails_at_65537_bytes(self):
        big = self.repo.repo / "skills" / "moai" / "workflows" / "big.md"
        big.parent.mkdir()
        big.write_bytes(b"x" * 65536)
        proc = self.repo.run()
        self.assertEqual(proc.returncode, 0, proc.stdout)
        rest = dict(items(proc.stdout)[0])["package-size"]
        self.assertTrue(rest.startswith("exit=0 limit=65536 largest=65536 "), rest)
        self.assertIn(" over=0 ", rest)
        big.write_bytes(b"x" * 65537)
        proc = self.repo.run()
        self.assertEqual(proc.returncode, 1, proc.stdout)
        found, last = items(proc.stdout)
        rest = dict(found)["package-size"]
        self.assertTrue(rest.startswith("exit=1 limit=65536 largest=65537 "), rest)
        self.assertIn(" over=1 ", rest)
        log = self.repo.repo / "target" / "self-check" / "package-size.log"
        self.assertIn("skills/moai/workflows/big.md", log.read_text())
        self.assertEqual(LAST.match(last)["failed"], "package-size")

    def test_an_empty_package_folder_is_a_skip_not_a_pass(self):
        shutil.rmtree(self.repo.repo / "skills")
        proc = self.repo.run()
        self.assertEqual(proc.returncode, 0, proc.stdout)
        rest = dict(items(proc.stdout)[0])
        for label in ("package-inspect", "package-size"):
            self.assertEqual(rest[label], "skip (no skills/moai/SKILL.md)")
        self.assertNotIn("package-inspect", " ".join(self.repo.recorded()))

    def test_an_item_whose_input_is_missing_is_skipped_with_the_reason(self):
        (self.repo.repo / "tests" / "git_env_scrub.rs").unlink()
        (self.repo.repo / "tools" / "source_manifest.py").unlink()
        proc = self.repo.run()
        self.assertEqual(proc.returncode, 0, proc.stdout)
        rest = dict(items(proc.stdout)[0])
        self.assertEqual(rest["git-env-scrub"], "skip (no tests/git_env_scrub.rs)")
        self.assertEqual(rest["manifest"], "skip (no tools/source_manifest.py)")
        self.assertNotIn("--test git_env_scrub", " ".join(self.repo.recorded()))

    def test_without_cargo_toml_every_cargo_item_is_skipped(self):
        (self.repo.repo / "Cargo.toml").unlink()
        proc = self.repo.run()
        rest = dict(items(proc.stdout)[0])
        for label, cmd in ITEMS:
            if cmd and cmd.startswith("cargo "):
                self.assertEqual(rest[label], "skip (no Cargo.toml)", label)
        self.assertEqual(self.repo.recorded(), ["python3 -I tools/source_manifest.py check"])

    def test_dry_run_runs_nothing_and_never_passes(self):
        proc = self.repo.run(SELF_CHECK_DRY_RUN="1")
        self.assertEqual(proc.returncode, 1)
        found, last = items(proc.stdout)
        self.assertEqual([label for label, _ in found], LABELS)
        self.assertEqual(self.repo.recorded(), [])
        m = LAST.match(last)
        self.assertEqual((m["result"], m["dry"]), ("fail", "dry-run=1 "))

    def test_an_uncommitted_change_is_flagged_on_the_last_line(self):
        (self.repo.repo / "new.txt").write_text("x\n")
        proc = self.repo.run()
        m = LAST.match(items(proc.stdout)[1])
        self.assertEqual(m["dirty"], "dirty=1 ")
        # The folder was checked; the line says it is not the committed head alone.
        self.assertEqual(proc.returncode, 0)

    def test_an_argument_is_a_usage_error(self):
        proc = self.repo.run("--bin")
        self.assertEqual(proc.returncode, 2)
        self.assertEqual(self.repo.recorded(), [])
        self.assertIn("usage", proc.stderr)


class SelfCheckSourcesTest(unittest.TestCase):
    def test_the_bound_is_the_profile_limit(self):
        self.assertIn("pub const LIMIT: u64 = 65536;", PROFILE.read_text())
        self.assertRegex(SCRIPT.read_text(), r"(?m)^limit=65536 ")

    def test_the_reference_names_the_runner_and_every_item(self):
        text = REFERENCE.read_text()
        self.assertIn("`tools/self_check.sh`", text)
        for label in LABELS:
            self.assertIn(f"`{label}`", text, label)


if __name__ == "__main__":
    unittest.main()
