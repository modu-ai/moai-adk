"""Pinning tests for the load-independent child budget of mutate_test (#2502).

The runner child's real progress is observable: every completed fake-cargo
invocation appends one JSON line to the mutation log. The budget therefore
resets on progress instead of expiring on wall-clock alone, keeps a hard hang
ceiling, and caps the number of invocations. These tests pin the mechanism and
its constants on a simulated clock with a fake child — no real waiting, so the
results are deterministic even on an overloaded machine — and fail when the
budget is loosened or removed (#2466 audit supplement 6811398145: five false
subprocess.TimeoutExpired under load 167-218 with zero assertion failures).
"""

import unittest

import mutate_test


class FakeClock:
    """Clock whose sleep() advances simulated time and runs scheduled events."""

    def __init__(self, schedule, child):
        self.now = 1000.0
        self.schedule = sorted(schedule)  # (sim_seconds_from_start, action)
        self.child = child

    def __call__(self):
        return self.now

    def sleep(self, seconds):
        self.now += seconds
        while self.schedule and self.schedule[0][0] <= self.now - 1000.0:
            _, action = self.schedule.pop(0)
            action()


class FakeChild:
    """Child that exits at a simulated time and never touches real processes."""

    def __init__(self, exit_at):
        self.exit_at = exit_at
        self.clock = None
        self.returncode = None
        self.killed = False

    def attach(self, clock):
        self.clock = clock
        return self

    def poll(self):
        if self.returncode is None and self.clock.now - 1000.0 >= self.exit_at:
            self.returncode = 0
        return self.returncode

    def wait(self):
        return self.returncode

    def kill(self):
        self.killed = True
        self.returncode = -9


class ChildBudgetPinTest(unittest.TestCase):
    def run_budget(self, *, schedule, exit_at, step, ceiling, budget):
        log_actions = [mutate_test.__dict__]  # placeholder, replaced below
        del log_actions
        import tempfile
        from pathlib import Path
        temp = tempfile.TemporaryDirectory(prefix="mutate-budget-pin-")
        self.addCleanup(temp.cleanup)
        log = Path(temp.name) / "progress.jsonl"

        def append_line():
            with log.open("a") as handle:
                handle.write('{"step": true}\n')

        child = FakeChild(exit_at)
        clock = FakeClock([(when, append_line) for when in schedule], child)
        child.attach(clock)
        out = mutate_test.run_with_progress_budget(
            ["fake-child"], cwd=temp.name, env={}, log=log,
            step_seconds=step, ceiling_seconds=ceiling, budget=budget,
            clock=clock, sleep=clock.sleep,
            spawn=lambda *args, **kwargs: child)
        return out, child

    def test_an_exited_child_that_overspent_the_budget_is_refused(self):
        # The child appends four entries and exits between polls; the exit
        # branch must still refuse it instead of returning a CompletedProcess.
        with self.assertRaises(mutate_test.BudgetExceeded) as caught:
            self.run_budget(schedule=[0.05, 0.1, 0.15, 0.19], exit_at=0.05,
                            step=0.5, ceiling=30.0, budget=3)
        self.assertIn("invocation budget", caught.exception.reason)

    def test_a_shared_log_baseline_does_not_consume_this_childs_budget(self):
        # MutationCLI keeps one log across invokes: earlier entries are the
        # baseline, and only this child's own three entries count.
        import tempfile
        from pathlib import Path
        temp = tempfile.TemporaryDirectory(prefix="mutate-budget-shared-")
        self.addCleanup(temp.cleanup)
        log = Path(temp.name) / "progress.jsonl"
        log.write_text('{"old": true}\n' * 5)

        def append_line():
            with log.open("a") as handle:
                handle.write('{"step": true}\n')

        child = FakeChild(0.9)
        clock = FakeClock([(0.3, append_line), (0.6, append_line), (0.8, append_line)],
                          child)
        child.attach(clock)
        out = mutate_test.run_with_progress_budget(
            ["fake-child"], cwd=temp.name, env={}, log=log,
            step_seconds=0.4, ceiling_seconds=30.0, budget=3,
            clock=clock, sleep=clock.sleep,
            spawn=lambda *args, **kwargs: child)
        self.assertEqual(out.returncode, 0)
        self.assertEqual(len(log.read_text().splitlines()), 8)

    def test_pinned_constants(self):
        # Touching the budget or the hang guard breaks this pin first (#2502 AC-2).
        self.assertEqual(mutate_test.PROGRESS_LOG_LINES_BUDGET, 64)
        self.assertEqual(mutate_test.STEP_DEADLINE_SECONDS, 120)
        self.assertEqual(mutate_test.HARD_CEILING_SECONDS, 600)

    def test_progress_inside_each_window_resets_the_deadline(self):
        # A line lands inside every 0.4s window; without the reset the very
        # first window would kill the child as silent.
        out, child = self.run_budget(
            schedule=[0.3, 0.9], exit_at=1.5, step=0.4, ceiling=30.0, budget=64)
        self.assertFalse(child.killed)
        self.assertEqual(out.returncode, 0)

    def test_a_child_without_progress_fails_with_the_no_progress_reason(self):
        with self.assertRaises(mutate_test.BudgetExceeded) as caught:
            self.run_budget(schedule=[], exit_at=30.0, step=0.4,
                            ceiling=30.0, budget=64)
        self.assertIn("no progress", caught.exception.reason)
        self.assertTrue(caught.exception.args[0])

    def test_progress_forever_stops_at_the_invocation_budget(self):
        with self.assertRaises(mutate_test.BudgetExceeded) as caught:
            self.run_budget(schedule=[0.1, 0.2, 0.3, 0.4], exit_at=30.0,
                            step=0.5, ceiling=30.0, budget=3)
        self.assertIn("invocation budget", caught.exception.reason)

    def test_slow_but_progressing_work_hits_the_hard_ceiling_not_the_deadline(self):
        with self.assertRaises(mutate_test.BudgetExceeded) as caught:
            self.run_budget(schedule=[0.4, 0.8, 1.2, 1.6], exit_at=30.0,
                            step=0.5, ceiling=1.2, budget=64)
        self.assertIn("hang ceiling", caught.exception.reason)


if __name__ == "__main__":
    unittest.main()
