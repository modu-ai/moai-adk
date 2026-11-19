"""SPEC-20261119-ci-smoke-step (card C-101): the `check` job of .github/workflows/ci.yml runs the
smoke harness (tools/smoke/run.sh) for the install and update areas after the `Check set` step.

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
    check_set = check.find("      - name: Check set")
    if check_set < 0 or check_set > check.find(STEP_LINE):
        found.append("smoke step is not after Check set")
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
        self.assertEqual(problems(before), ["smoke step is not after Check set"])

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

    def test_the_other_steps_of_the_check_job_are_still_there_in_order(self):
        """github_assets_test finds these by name; the new step adds one and moves none."""
        check = job_region(CI.read_text(encoding="utf-8"), "check")
        names = re.findall(r"(?m)^      - name: (.+)$", check)
        for earlier, later in (("Dependency cache", "Check set"), ("Check set", STEP_NAME)):
            self.assertIn(earlier, names)
            self.assertIn(later, names)
            self.assertLess(names.index(earlier), names.index(later))


if __name__ == "__main__":
    unittest.main()
