"""tools/heavy_check_slot.py (Issue #2491): real processes, a temporary slot folder only."""
import json
import os
import signal
import subprocess
import sys
import tempfile
import time
import unittest
from pathlib import Path

TOOL = Path(__file__).resolve().parent.parent / "tools" / "heavy_check_slot.py"
UNIX = os.name == "posix"

# Each job appends "start <t>" then "end <t>" to its own file around a sleep.
JOB = "import sys,time;p=sys.argv[1];open(p,'a').write(f'start {time.time()}\\n');time.sleep(float(sys.argv[2]));open(p,'a').write(f'end {time.time()}\\n')"


def tool(folder, *args):
    return [sys.executable, "-I", str(TOOL), *args[:1], "--dir", str(folder), *args[1:]]


def wait_for(predicate, seconds=10.0):
    deadline = time.monotonic() + seconds
    while time.monotonic() < deadline:
        if predicate():
            return True
        time.sleep(0.05)
    return False


def held(folder):
    out = subprocess.run(tool(folder, "status"), capture_output=True, text=True, check=True)
    return json.loads(out.stdout)["held"]


@unittest.skipUnless(UNIX, "flock slots are unix only")
class HeavyCheckSlotTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.folder = Path(self.tmp.name) / "slots"
        self.addCleanup(self.tmp.cleanup)

    def acquire(self, maximum, name, *command, **kw):
        return subprocess.Popen(
            tool(self.folder, "acquire", "--max", str(maximum), "--name", name, "--", *command),
            stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, **kw)

    def test_ac1_max_two_runs_two_and_the_third_waits_for_one_to_end(self):
        logs = [Path(self.tmp.name) / f"job{i}" for i in range(3)]
        procs = [self.acquire(2, f"lane{i}", sys.executable, "-c", JOB, str(log), "1.5")
                 for i, log in enumerate(logs)]
        outs = [p.communicate(timeout=30)[0] for p in procs]
        self.assertEqual([p.returncode for p in procs], [0, 0, 0], outs)
        spans = []
        for log in logs:
            lines = dict(line.split() for line in log.read_text().splitlines())
            spans.append((float(lines["start"]), float(lines["end"])))
        spans.sort()
        # At no moment were three jobs running: the last start is after the first end.
        self.assertGreaterEqual(spans[2][0], min(spans[0][1], spans[1][1]), spans)
        self.assertTrue(any("waiting 1 of 1 for 2 slot(s)" in o for o in outs), outs)

    def test_ac2_sigkill_of_the_holder_frees_the_slot_at_once(self):
        holder = self.acquire(1, "victim", "sleep", "60", start_new_session=True)
        self.assertTrue(wait_for(lambda: [h["name"] for h in held(self.folder)] == ["victim"]))
        os.killpg(holder.pid, signal.SIGKILL)
        holder.communicate(timeout=10)
        nxt = self.acquire(1, "next", "true")
        try:
            out = nxt.communicate(timeout=10)[0]
        except subprocess.TimeoutExpired:
            nxt.kill()
            out = nxt.communicate()[0]
            self.fail(f"the killed holder's slot was not freed; next caller still waiting: {out!r}")
        self.assertEqual(nxt.returncode, 0, out)
        # "At once": the next caller took the slot on its first try, never queued.
        self.assertNotIn("waiting", out)
        self.assertIn("slot 0 taken by next", out)
        self.assertEqual(held(self.folder), [])

    def test_slot_stays_held_while_the_command_outlives_a_killed_wrapper(self):
        marker = Path(self.tmp.name) / "release"
        script = f"import os,time\nwhile not os.path.exists({str(marker)!r}): time.sleep(0.05)"
        wrapper = self.acquire(1, "orphan", sys.executable, "-c", script)
        self.assertTrue(wait_for(lambda: len(held(self.folder)) == 1))
        os.kill(wrapper.pid, signal.SIGKILL)
        wrapper.wait(timeout=10)
        wrapper.stdout.close()  # the surviving command still holds the pipe's write end
        time.sleep(0.5)
        self.assertEqual(len(held(self.folder)), 1, "the command still holds the slot")
        marker.write_text("")
        self.assertTrue(wait_for(lambda: held(self.folder) == []), "freed when the command ends")

    def test_ac3_exit_codes_are_passed_through(self):
        for command, want in [
            (["true"], 0),
            (["false"], 1),
            ([sys.executable, "-c", "import sys; sys.exit(7)"], 7),
            (["sh", "-c", "kill -TERM $$"], 128 + signal.SIGTERM),
            (["sh", "-c", "kill -KILL $$"], 128 + signal.SIGKILL),
        ]:
            p = self.acquire(1, "codes", *command)
            out = p.communicate(timeout=10)[0]
            self.assertEqual(p.returncode, want, (command, out))
        self.assertEqual(held(self.folder), [])

    def test_status_names_the_holder_and_reads_only(self):
        self.assertEqual(held(self.folder), [], "a missing folder is empty, not an error")
        self.assertFalse(self.folder.exists(), "status creates nothing")
        p = self.acquire(1, "lane-x", "sleep", "30", start_new_session=True)
        self.addCleanup(lambda: (os.killpg(p.pid, signal.SIGKILL), p.communicate(timeout=10)))
        self.assertTrue(wait_for(lambda: len(held(self.folder)) == 1))
        [record] = held(self.folder)
        self.assertEqual(record["name"], "lane-x")
        self.assertEqual(record["command"], ["sleep", "30"])
        self.assertEqual(record["slot"], 0)
        self.assertIsInstance(record["started_unix"], int)

    def test_usage_errors_exit_two(self):
        for args in (["--max", "0", "--name", "x", "--", "true"], ["--max", "1", "--name", "x", "--"]):
            r = subprocess.run(tool(self.folder, "acquire", *args), capture_output=True)
            self.assertEqual(r.returncode, 2, args)


if __name__ == "__main__":
    unittest.main()
