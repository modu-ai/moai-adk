"""SPEC-20261119-ci-smoke-step (card C-101): the `check` job of .github/workflows/ci.yml runs the
smoke harness (tools/smoke/run.sh) for the install and update areas after the check set (the
step that runs tools/local_check.sh, whatever it is called: C-20 may split it into several).

The workflow is read as text, not through a YAML library: the test runs wherever the other
unittest files run, and the shape it pins (a step's place, its condition, its commands) is plain
lines. `problems(text)` names each missing part; the real workflow must have none.
"""

import re
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
CI = ROOT / ".github" / "workflows" / "ci.yml"
STEP_NAME = "Smoke harness (install, update)"
STEP_LINE = f"      - name: {STEP_NAME}"


def job_region(text, job):
    """The text of one top-level job: from `  <job>:` to the next two-space key."""
    start = re.search(rf"(?m)^  {re.escape(job)}:[ \t]*\n", text)
    if start is None:
        return ""
    rest = text[start.end():]
    nxt = re.search(r"(?m)^  [A-Za-z0-9_-]+:[ \t]*(#.*)?$", rest)
    return rest[: nxt.start()] if nxt else rest


def step_block(region):
    """The lines of the smoke step: from its `- name:` line to the next step or job key."""
    lines = region.splitlines()
    start = next((i for i, line in enumerate(lines) if line.rstrip() == STEP_LINE), None)
    if start is None:
        return None
    end = len(lines)
    for i in range(start + 1, len(lines)):
        line = lines[i]
        if line.startswith("      - ") or (line.strip() and not line.startswith("   ")):
            end = i
            break
    return "\n".join(lines[start:end])


def problems(text):
    """What the workflow text lacks of the smoke step, one phrase per part; [] when whole."""
    check = job_region(text, "check")
    block = step_block(check)
    if block is None:
        return ["smoke step missing"] if STEP_LINE not in text else ["smoke step is not in the check job"]
    found = []
    # "After the check set" is read by order, not by one step's name: the last step before the
    # smoke step that is named `Check set…` or that runs tools/local_check.sh (a later change
    # may split the check set into several steps; the smoke step must follow all of them).
    own = check.find(STEP_LINE)
    before = check[:own]
    check_set = max(before.rfind("      - name: Check set"), before.rfind("tools/local_check.sh"))
    after = check[own + len(STEP_LINE):]
    if check_set < 0 or "tools/local_check.sh" in after or "      - name: Check set" in after:
        found.append("smoke step is not after the check set")
    if "if: ${{ env.CI_SCOPE == 'full' }}" not in block:
        found.append("smoke step lacks the full-scope condition")
    build = block.find("cargo build --offline --locked --bin moai-cli")
    harness = block.find("tools/smoke/run.sh")
    if build < 0 or harness < 0 or build > harness:
        found.append("smoke step lacks the build of moai-cli before the harness")
    if "${CARGO_TARGET_DIR:-target}/debug/moai-cli" not in block:
        found.append("smoke step lacks the debug binary path")
    # `--bin` is also a word of the cargo build line, so read the harness call itself.
    if 'tools/smoke/run.sh "$area" --bin "$bin"' not in block:
        found.append("smoke step lacks --bin")
    if not re.search(r"for area in install update; do", block) or 'tools/smoke/run.sh "$area"' not in block:
        found.append("smoke step lacks the areas install and update")
    if "$GITHUB_STEP_SUMMARY" not in block or "tail -n 1" not in block or "$code" not in block:
        found.append("smoke step lacks the summary line")
    if "|| code=$?" not in block or "failed=1" not in block or 'exit "$failed"' not in block:
        found.append("smoke step lacks the failing exit")
    # A step that may fail without failing the job keeps the exit above from reaching the PR.
    if re.search(r"(?m)^\s*continue-on-error:", block):
        found.append("smoke step continues on error")
    return found


GOOD = """\
jobs:
  check:
    name: ci
    steps:
      - name: Check set
        shell: bash
        run: |
          bash tools/local_check.sh
      # C-101: the smoke harness on the binary this job built.
      - name: Smoke harness (install, update)
        if: ${{ env.CI_SCOPE == 'full' }}
        shell: bash
        run: |
          set -u
          cargo build --offline --locked --bin moai-cli
          bin="${CARGO_TARGET_DIR:-target}/debug/moai-cli"
          failed=0
          for area in install update; do
            code=0
            tools/smoke/run.sh "$area" --bin "$bin" > "ci-smoke-$area.log" 2>&1 || code=$?
            cat "ci-smoke-$area.log"
            echo "- smoke $area: exit $code" "$(tail -n 1 "ci-smoke-$area.log")" >> "$GITHUB_STEP_SUMMARY"
            [ "$code" -eq 0 ] || failed=1
          done
          exit "$failed"
  windows-build:
    runs-on: windows-2025
    steps:
      - name: Build
        run: cargo build
"""


class ShapeTest(unittest.TestCase):
    """The reading names each missing part; whole text names none."""

    def test_a_whole_step_has_no_problem(self):
        self.assertEqual(problems(GOOD), [])

    def broken(self, old, new):
        self.assertEqual(GOOD.count(old), 1, old)
        return problems(GOOD.replace(old, new))

    def test_a_workflow_without_the_step_says_so_and_nothing_else(self):
        text = GOOD.replace(STEP_LINE, "      - name: Another step")
        self.assertEqual(problems(text), ["smoke step missing"])

    def test_the_step_in_another_job_is_not_the_check_job_step(self):
        head, tail = GOOD.split("  windows-build:\n")
        step = head[head.index("      # C-101") :]
        moved = head.replace(step, "") + "  windows-build:\n" + tail + step
        self.assertEqual(problems(moved), ["smoke step is not in the check job"])

    def test_the_step_before_check_set_is_refused(self):
        step_start = GOOD.index("      # C-101")
        step_end = GOOD.index("  windows-build:")
        step = GOOD[step_start:step_end]
        before = GOOD.replace(step, "").replace(
            "      - name: Check set", step + "      - name: Check set"
        )
        self.assertEqual(problems(before), ["smoke step is not after the check set"])

    def test_the_place_is_read_by_order_not_by_one_name(self):
        """C-20 may split the check set into several steps: the smoke step must follow the one
        that runs tools/local_check.sh whatever it is called, and none may run after it."""
        split = GOOD.replace(
            "      - name: Check set\n        shell: bash\n        run: |\n          bash tools/local_check.sh\n",
            "      - name: Bridge one\n        run: |\n          echo prepare\n"
            "      - name: Bridge two\n        shell: bash\n        run: |\n          bash tools/local_check.sh\n",
        )
        self.assertNotEqual(split, GOOD)
        self.assertEqual(problems(split), [])
        late = split.replace(
            "  windows-build:\n",
            "      - name: Late bridge\n        run: |\n          bash tools/local_check.sh\n  windows-build:\n",
        )
        self.assertEqual(problems(late), ["smoke step is not after the check set"])
        none = GOOD.replace("          bash tools/local_check.sh\n", "          echo nothing\n").replace(
            "- name: Check set", "- name: Something else"
        )
        self.assertNotEqual(none, GOOD)
        self.assertEqual(problems(none), ["smoke step is not after the check set"])

    def test_continue_on_error_is_refused(self):
        self.assertEqual(
            self.broken("        run: |\n          set -u\n", "        continue-on-error: true\n        run: |\n          set -u\n"),
            ["smoke step continues on error"],
        )

    def test_each_missing_part_is_named(self):
        cases = {
            "smoke step lacks the full-scope condition": (
                "        if: ${{ env.CI_SCOPE == 'full' }}\n", ""),
            "smoke step lacks the build of moai-cli before the harness": (
                "          cargo build --offline --locked --bin moai-cli\n", ""),
            "smoke step lacks the debug binary path": (
                'bin="${CARGO_TARGET_DIR:-target}/debug/moai-cli"', 'bin="target/release/moai-cli"'),
            "smoke step lacks --bin": (' --bin "$bin"', ""),
            "smoke step lacks the areas install and update": (
                "for area in install update; do", "for area in install; do"),
            "smoke step lacks the summary line": (
                '>> "$GITHUB_STEP_SUMMARY"', ">> smoke.txt"),
            "smoke step lacks the failing exit": (' || code=$?', ""),
        }
        for message, (old, new) in cases.items():
            with self.subTest(message=message):
                self.assertIn(message, self.broken(old, new))

    def test_the_build_after_the_harness_is_refused(self):
        build = "          cargo build --offline --locked --bin moai-cli\n"
        text = GOOD.replace(build, "").replace("          exit \"$failed\"\n", "          exit \"$failed\"\n" + build)
        self.assertIn("smoke step lacks the build of moai-cli before the harness", problems(text))

    def test_failing_exit_needs_the_exit_and_the_flag(self):
        self.assertIn("smoke step lacks the failing exit", self.broken('exit "$failed"\n', ""))
        self.assertIn("smoke step lacks the failing exit", self.broken("failed=1", "failed=0"))


class RealWorkflowTest(unittest.TestCase):
    def test_the_check_job_runs_the_smoke_harness_for_install_and_update(self):
        text = CI.read_text(encoding="utf-8")
        self.assertEqual(problems(text), [])

    def test_the_smoke_step_comes_after_the_dependency_cache(self):
        """The step reuses the build the cache and the check set left; github_assets_test pins
        the other steps by name, so only this one order is read here."""
        check = job_region(CI.read_text(encoding="utf-8"), "check")
        names = re.findall(r"(?m)^      - name: (.+)$", check)
        self.assertIn("Dependency cache", names)
        self.assertIn(STEP_NAME, names)
        self.assertLess(names.index("Dependency cache"), names.index(STEP_NAME))


def run_body(text):
    """The shell text of the smoke step's `run: |` block, dedented as the runner sees it."""
    block = step_block(job_region(text, "check"))
    if block is None:
        raise AssertionError("smoke step missing")
    lines = block.splitlines()
    start = next(i for i, line in enumerate(lines) if line == "        run: |") + 1
    body = []
    for line in lines[start:]:
        if line.strip() and not line.startswith("          "):
            break
        body.append(line[10:])
    return "\n".join(body) + "\n"


FAKE_RUN_SH = """\
#!/bin/bash
echo "$*" >> "$CALLS"
case " $FAIL_AREAS " in
  *" $1 "*) echo "smoke: $1 fail steps=2 (fake)"; exit 1 ;;
esac
echo "smoke: $1 pass steps=4 (fake)"
exit 0
"""


def run_step(text, fail_areas=""):
    """Run the step's shell with a fake `tools/smoke/run.sh` and a fake `cargo` in a temporary
    folder and a temporary HOME, under the runner's shell flags. Returns (exit, calls, summary)."""
    import os
    import subprocess
    import tempfile

    with tempfile.TemporaryDirectory() as tmp:
        root = Path(tmp)
        (root / "tools" / "smoke").mkdir(parents=True)
        (root / "bin").mkdir()
        script = root / "tools" / "smoke" / "run.sh"
        script.write_text(FAKE_RUN_SH, encoding="utf-8")
        cargo = root / "bin" / "cargo"
        cargo.write_text("#!/bin/sh\necho \"cargo $*\" >> \"$CALLS\"\nexit 0\n", encoding="utf-8")
        for path in (script, cargo):
            path.chmod(0o755)
        env = {
            "PATH": f"{root / 'bin'}:{os.environ.get('PATH', '/usr/bin:/bin')}",
            "HOME": str(root / "home"),
            "CALLS": str(root / "calls.txt"),
            "FAIL_AREAS": fail_areas,
            "GITHUB_STEP_SUMMARY": str(root / "summary.md"),
        }
        done = subprocess.run(
            ["bash", "--noprofile", "--norc", "-eo", "pipefail", "-c", run_body(text)],
            cwd=root, env=env, capture_output=True, text=True, timeout=60,
        )
        calls = (root / "calls.txt").read_text(encoding="utf-8").splitlines() if (root / "calls.txt").exists() else []
        summary = (root / "summary.md").read_text(encoding="utf-8").splitlines() if (root / "summary.md").exists() else []
        return done.returncode, calls, summary


class BehaviorTest(unittest.TestCase):
    """REQ-3 to REQ-5 by running the step, not by reading it: both areas run whatever the first
    one does, the job's exit is 1 when any area failed, one summary line per area."""

    @classmethod
    def setUpClass(cls):
        cls.text = CI.read_text(encoding="utf-8")

    def check(self, fail_areas, exit_code):
        code, calls, summary = run_step(self.text, fail_areas)
        self.assertEqual(code, exit_code, (calls, summary))
        self.assertEqual(calls[0], "cargo build --offline --locked --bin moai-cli")
        # Both areas ran, in order, each on the debug binary of this job.
        self.assertEqual(
            calls[1:],
            ["install --bin target/debug/moai-cli", "update --bin target/debug/moai-cli"],
        )
        self.assertEqual(len(summary), 2, summary)
        for area, line in zip(("install", "update"), summary):
            wanted = 1 if area in fail_areas.split() else 0
            self.assertIn(f"- smoke {area}: exit {wanted}", line)
            self.assertIn(f"smoke: {area} {'fail' if wanted else 'pass'} steps=", line)

    def test_both_pass_exits_zero(self):
        self.check("", 0)

    def test_a_failing_first_area_still_runs_the_second_and_exits_one(self):
        self.check("install", 1)

    def test_a_failing_second_area_exits_one(self):
        self.check("update", 1)

    def test_both_failing_exits_one(self):
        self.check("install update", 1)

    def test_a_break_exit_or_continue_on_error_would_be_seen(self):
        """The reading above cannot tell `|| failed=1` from `|| { failed=1; break; }`; the run can."""
        broken = self.text.replace('[ "$code" -eq 0 ] || failed=1', '[ "$code" -eq 0 ] || { failed=1; break; }')
        self.assertNotEqual(broken, self.text)
        code, calls, _ = run_step(broken, "install")
        self.assertEqual(code, 1)
        self.assertEqual(calls, ["cargo build --offline --locked --bin moai-cli", "install --bin target/debug/moai-cli"])


if __name__ == "__main__":
    unittest.main()
