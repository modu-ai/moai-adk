"""Tests for tools/smoke (SPEC-20261119-smoke-harness): the smoke harness the lanes and the
auditors share. It boots the built binary in an empty temporary HOME and a temporary git
project and prints a step trace.

These tests run the real harness with the real built binary (`target/debug/moai-cli`, or
`$CARGO_TARGET_DIR/debug/moai-cli`) and with small fake binaries for the failure and isolation
cases. The caller's HOME in every test is a decoy folder with sentinel files: a harness that
read or wrote the caller's HOME would change the decoy or leak its path.
"""

import hashlib
import os
import re
import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
RUN = REPO_ROOT / "tools" / "smoke" / "run.sh"

# The steps each area must trace as passed (SPEC REQ-6 to REQ-9).
STEPS = {
    "install": ["stage-package", "version", "install-inspect", "home-untouched"],
    "update": ["stage-package", "preview", "confirm", "rerun"],
    "project": ["detect", "plan", "apply"],
    "launcher": ["dry-run", "host-in-pty"],
}


def built_binary() -> Path:
    target = os.environ.get("CARGO_TARGET_DIR")
    binary = (Path(target) if target else REPO_ROOT / "target") / "debug" / "moai-cli"
    if not binary.is_file():
        raise AssertionError(f"moai-cli not built at {binary}; run cargo build first")
    return binary


def snapshot(folder: Path) -> dict:
    """Every entry under `folder` (files by content hash, folders by name): what a read-only
    caller leaves unchanged."""
    seen = {}
    for path in sorted(folder.rglob("*")):
        rel = str(path.relative_to(folder))
        if path.is_file():
            seen[rel] = hashlib.sha256(path.read_bytes()).hexdigest()
        else:
            seen[rel] = "dir"
    return seen


class Decoy:
    """A folder standing for the caller's real HOME, with sentinel files."""

    def __init__(self, test: unittest.TestCase):
        self.path = Path(tempfile.mkdtemp(prefix="smoke-decoy-home-")).resolve()
        test.addCleanup(shutil.rmtree, self.path, True)
        (self.path / ".claude").mkdir()
        (self.path / ".claude" / "settings.json").write_text('{"decoy":true}\n')
        (self.path / "sentinel.txt").write_text("do not touch\n")
        self.before = snapshot(self.path)

    def assert_untouched(self, test: unittest.TestCase):
        test.assertEqual(self.before, snapshot(self.path), "the caller's HOME changed")


def fake_binary(test: unittest.TestCase, body: str) -> Path:
    """An executable shell script standing for the binary under test."""
    folder = Path(tempfile.mkdtemp(prefix="smoke-fake-bin-"))
    test.addCleanup(shutil.rmtree, folder, True)
    script = folder / "moai-cli"
    script.write_text("#!/bin/sh\n" + body)
    script.chmod(0o755)
    return script


def smoke(args, home: Path, extra_env=None) -> subprocess.CompletedProcess:
    """`tools/smoke/run.sh` with a caller environment that has only PATH and the decoy HOME
    (plus `extra_env`), as a lane's shell would."""
    env = {"PATH": os.environ["PATH"], "HOME": str(home)}
    env.update(extra_env or {})
    return subprocess.run(
        ["bash", str(RUN), *args],
        env=env,
        stdin=subprocess.DEVNULL,
        capture_output=True,
        text=True,
        timeout=300,
    )


def last_line(result: subprocess.CompletedProcess) -> str:
    lines = [l for l in result.stdout.splitlines() if l.strip()]
    return lines[-1] if lines else ""


class AreasTest(unittest.TestCase):
    """AC-1: each area passes with the built binary and traces its steps."""

    def run_area(self, area: str) -> subprocess.CompletedProcess:
        decoy = Decoy(self)
        result = smoke([area, "--bin", str(built_binary())], decoy.path)
        decoy.assert_untouched(self)
        return result

    def check_area(self, area: str):
        result = self.run_area(area)
        out = result.stdout
        self.assertEqual(result.returncode, 0, out + result.stderr)
        self.assertRegex(last_line(result), rf"^smoke: {area} pass steps=\d+", out)
        for step in STEPS[area]:
            self.assertRegex(out, rf"(?m)^\[ok\] +{re.escape(step)} \(\d+\.\d+s\)", out)
        self.assertNotIn("[FAIL]", out)

    def test_install(self):
        self.check_area("install")

    def test_update(self):
        self.check_area("update")

    def test_project(self):
        self.check_area("project")

    def test_launcher(self):
        self.check_area("launcher")

    def test_all_runs_every_area_in_order(self):
        result = self.run_area("all")
        out = result.stdout
        self.assertEqual(result.returncode, 0, out + result.stderr)
        total = sum(len(steps) for steps in STEPS.values())
        self.assertRegex(last_line(result), rf"^smoke: all pass steps={total}\b", out)
        positions = [out.index(f"area {area}") for area in STEPS]
        self.assertEqual(positions, sorted(positions), out)


class FailureTest(unittest.TestCase):
    """AC-2: a failing step names itself and shows its command, exit code and output."""

    def test_a_binary_that_fails_stops_at_its_first_step(self):
        decoy = Decoy(self)
        binary = fake_binary(self, 'echo "boom: $*" >&2\nexit 3\n')
        result = smoke(["install", "--bin", str(binary)], decoy.path)
        out = result.stdout
        self.assertEqual(result.returncode, 1, out + result.stderr)
        self.assertRegex(out, r"(?m)^\[ok\] +stage-package ")
        self.assertRegex(out, r"(?m)^\[FAIL\] +version ")
        self.assertRegex(out, r"(?m)^  command: .*moai-cli.* --version")
        self.assertRegex(out, r"(?m)^  exit: 3$")
        self.assertIn("boom: --version", out)
        self.assertNotRegex(out, r"(?m)^\[ok\] +install-inspect")
        self.assertRegex(last_line(result), r"^smoke: install fail steps=2 failed=version\b")
        decoy.assert_untouched(self)

    def test_an_answer_that_is_not_the_expected_json_names_the_step_and_shows_it(self):
        decoy = Decoy(self)
        binary = fake_binary(
            self,
            'case "$1" in --version) echo "moai-cli 0.0.0 (rust)";; *) echo "not json at all";; esac\n',
        )
        result = smoke(["install", "--bin", str(binary)], decoy.path)
        out = result.stdout
        self.assertEqual(result.returncode, 1, out + result.stderr)
        self.assertRegex(out, r"(?m)^\[ok\] +version ")
        self.assertRegex(out, r"(?m)^\[FAIL\] +install-inspect ")
        self.assertIn("not json at all", out)
        self.assertRegex(last_line(result), r"failed=install-inspect\b")


class IsolationTest(unittest.TestCase):
    """AC-3: the caller's HOME and environment never reach a step."""

    def test_no_step_sees_the_callers_home_or_environment(self):
        decoy = Decoy(self)
        log = Path(tempfile.mkdtemp(prefix="smoke-log-")) / "calls.log"
        self.addCleanup(shutil.rmtree, log.parent, True)
        binary = fake_binary(
            self,
            f'{{ echo "HOME=$HOME"; echo "PWD=$PWD"; env; }} >> "{log}"\n'
            'case "$1" in --version) echo "moai-cli 0.0.0 (rust)";; *) exit 9;; esac\n',
        )
        result = smoke(
            ["install", "--bin", str(binary)],
            decoy.path,
            {"PROBE_CALLER_VALUE": "leak-me-1", "CLAUDE_CODE_SESSION_ID": "leak-me-2"},
        )
        self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
        text = log.read_text()
        self.assertNotIn("leak-me-1", text)
        self.assertNotIn("leak-me-2", text)
        self.assertNotIn(str(decoy.path), text)
        homes = re.findall(r"(?m)^HOME=(.*)$", text)
        self.assertTrue(homes, text)
        for home in homes:
            real = Path(home).resolve()
            self.assertNotEqual(real, decoy.path)
            self.assertTrue(real.name == "home", home)
        decoy.assert_untouched(self)

    def test_a_temp_base_inside_the_callers_home_is_refused_before_anything_is_created(self):
        decoy = Decoy(self)
        base = decoy.path / "tmp"
        base.mkdir()
        before = snapshot(decoy.path)
        binary = fake_binary(self, "exit 0\n")
        result = smoke(
            ["install", "--bin", str(binary)], decoy.path, {"TMPDIR": str(base)}
        )
        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        self.assertIn("real home", result.stdout + result.stderr)
        self.assertEqual(before, snapshot(decoy.path), "something was created inside HOME")

    def test_there_is_no_option_that_points_a_step_at_a_home(self):
        decoy = Decoy(self)
        binary = fake_binary(self, "exit 0\n")
        result = smoke(
            ["install", "--bin", str(binary), "--home", str(decoy.path)], decoy.path
        )
        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        decoy.assert_untouched(self)


class UsageTest(unittest.TestCase):
    """AC-4: usage errors exit 2; the temporary root goes away unless `--keep`."""

    def test_an_unknown_area_is_a_usage_error(self):
        decoy = Decoy(self)
        result = smoke(["nope", "--bin", str(fake_binary(self, "exit 0\n"))], decoy.path)
        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        self.assertIn("unknown area", result.stdout + result.stderr)

    def test_a_missing_or_non_executable_binary_is_a_usage_error(self):
        decoy = Decoy(self)
        missing = smoke(["install", "--bin", "/no/such/moai-cli"], decoy.path)
        self.assertEqual(missing.returncode, 2, missing.stdout + missing.stderr)
        self.assertIn("binary", missing.stdout + missing.stderr)
        plain = Path(tempfile.mkdtemp(prefix="smoke-plain-")) / "moai-cli"
        self.addCleanup(shutil.rmtree, plain.parent, True)
        plain.write_text("not executable\n")
        plain.chmod(0o644)
        result = smoke(["install", "--bin", str(plain)], decoy.path)
        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        self.assertIn("executable", result.stdout + result.stderr)

    def root_of(self, result: subprocess.CompletedProcess) -> Path:
        first = result.stdout.splitlines()[0]
        self.assertTrue(first.startswith("smoke: start "), first)
        return Path(re.search(r"root=(\S+)", first).group(1))

    def test_the_temporary_root_is_removed_and_kept_with_keep(self):
        decoy = Decoy(self)
        binary = fake_binary(self, "exit 3\n")
        gone = smoke(["install", "--bin", str(binary)], decoy.path)
        self.assertFalse(self.root_of(gone).exists(), gone.stdout)
        kept = smoke(["install", "--bin", str(binary), "--keep"], decoy.path)
        root = self.root_of(kept)
        self.addCleanup(shutil.rmtree, root, True)
        self.assertTrue(root.is_dir(), kept.stdout)
        self.assertTrue((root / "home").is_dir(), kept.stdout)
        decoy.assert_untouched(self)


if __name__ == "__main__":
    unittest.main()
