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


def holder_names(folder):
    """Names of the held slots. The holder takes the slot's lock before it writes its record,
    so for a moment `status` lists a held slot with an empty record (#2599); that slot has no
    name yet and the caller keeps waiting instead of failing on a missing key."""
    return [h.get("name") for h in held(folder)]


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
        self.assertTrue(wait_for(lambda: holder_names(self.folder) == ["victim"]))
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
        started = Path(self.tmp.name) / "started"
        script = (f"import os,time\nopen({str(started)!r}, 'w').close()\n"
                  f"while not os.path.exists({str(marker)!r}): time.sleep(0.05)")
        wrapper = self.acquire(1, "orphan", sys.executable, "-c", script)
        # Wait for the command itself, not only the slot lock: the wrapper takes the lock,
        # writes its record and only then starts the command (#2599, audit 6838460410 B1).
        # Killing the wrapper before the command runs leaves nobody holding the slot.
        self.assertTrue(wait_for(started.exists))
        self.assertTrue(wait_for(lambda: len(held(self.folder)) == 1))
        os.kill(wrapper.pid, signal.SIGKILL)
        wrapper.wait(timeout=10)
        wrapper.stdout.close()  # the surviving command still holds the pipe's write end
        time.sleep(0.5)
        self.assertEqual(len(held(self.folder)), 1, "the command still holds the slot")
        marker.write_text("")
        self.assertTrue(wait_for(lambda: held(self.folder) == []), "freed when the command ends")

    def test_sigterm_to_the_wrapper_alone_reaches_the_command(self):
        """Pre-audit 6814850663 M5: a SIGTERM sent only to the wrapper is passed to the
        command, which ends; the wrapper returns 143 and the slot is free again."""
        wrapper = self.acquire(1, "term", "sleep", "60")
        # Wait for the holder's record, not only the slot lock: the wrapper installs its signal
        # handlers after taking the lock and before writing the record, so a SIGTERM in that
        # gap would kill the wrapper outright (#2599, audit 6838460410). Once the record is
        # there, a SIGTERM before the command starts is held and passed on (#2571 N1).
        self.assertTrue(wait_for(lambda: holder_names(self.folder) == ["term"]))
        os.kill(wrapper.pid, signal.SIGTERM)
        try:
            out = wrapper.communicate(timeout=10)[0]
        except subprocess.TimeoutExpired:
            wrapper.kill()
            wrapper.wait()
            wrapper.stdout.close()  # the unsignalled command still holds the pipe
            self.fail("the command did not receive the wrapper's SIGTERM; wrapper still waiting")
        self.assertEqual(wrapper.returncode, 128 + signal.SIGTERM, out)
        self.assertEqual(held(self.folder), [], "the slot is free once both are gone")

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
        self.assertTrue(wait_for(lambda: holder_names(self.folder) == ["lane-x"]))
        [record] = held(self.folder)
        self.assertEqual(record["name"], "lane-x")
        self.assertEqual(record["command"], ["sleep", "30"])
        self.assertEqual(record["slot"], 0)
        self.assertIsInstance(record["started_unix"], int)

    def refused_run(self, make_slot):
        """Plant a non-plain `slot-0`, run acquire --max 1, return (result, ran marker)."""
        self.folder.mkdir()
        make_slot(self.folder / "slot-0")
        ran = Path(self.tmp.name) / "ran"
        r = subprocess.run(
            tool(self.folder, "acquire", "--max", "1", "--name", "x", "--",
                 sys.executable, "-c", f"open({str(ran)!r}, 'w')"),
            capture_output=True, text=True, timeout=30)
        return r, ran

    def test_symlinked_slot_is_refused_and_its_target_untouched(self):
        """Audit B1 (PR #2571): a symlink at a slot path made the tool truncate and
        overwrite the link's target."""
        victim = Path(self.tmp.name) / "victim.txt"
        victim.write_bytes(b"precious\n")
        r, ran = self.refused_run(lambda slot: slot.symlink_to(victim))
        self.assertEqual(r.returncode, 125, r)
        self.assertEqual(victim.read_bytes(), b"precious\n")
        self.assertFalse(ran.exists(), "COMMAND must not run on a refusal")
        self.assertIn("refused", r.stderr)
        self.assertNotIn("Traceback", r.stderr)
        self.assertTrue((self.folder / "slot-0").is_symlink(), "the link is left as found")

    def test_dangling_symlinked_slot_creates_nothing_at_its_target(self):
        target = Path(self.tmp.name) / "created-elsewhere"
        r, ran = self.refused_run(lambda slot: slot.symlink_to(target))
        self.assertEqual(r.returncode, 125, r)
        self.assertFalse(target.exists())
        self.assertFalse(ran.exists())

    def test_hard_linked_slot_is_refused_and_the_other_name_untouched(self):
        victim = Path(self.tmp.name) / "victim.txt"
        victim.write_bytes(b"precious\n")
        r, ran = self.refused_run(lambda slot: os.link(victim, slot))
        self.assertEqual(r.returncode, 125, r)
        self.assertEqual(victim.read_bytes(), b"precious\n")
        self.assertFalse(ran.exists())

    def test_fifo_and_folder_slots_are_refused_without_a_traceback(self):
        """Audit N2: a FIFO slot ended in an unhandled OSError traceback."""
        for make in (os.mkfifo, os.mkdir):
            with self.subTest(make=make.__name__):
                self.folder = Path(self.tmp.name) / f"slots-{make.__name__}"
                r, ran = self.refused_run(make)
                self.assertEqual(r.returncode, 125, r)
                self.assertNotIn("Traceback", r.stderr)
                self.assertFalse(ran.exists())

    def test_status_lists_a_refused_slot_and_writes_nothing(self):
        victim = Path(self.tmp.name) / "victim.txt"
        victim.write_bytes(b"precious\n")
        self.folder.mkdir()
        (self.folder / "slot-0").symlink_to(victim)
        out = subprocess.run(tool(self.folder, "status"), capture_output=True, text=True, check=True)
        state = json.loads(out.stdout)
        self.assertEqual((state["held"], state["refused"]), ([], ["slot-0"]))
        self.assertEqual(victim.read_bytes(), b"precious\n")

    def test_a_symlink_among_the_waiters_is_skipped_never_written(self):
        victim = Path(self.tmp.name) / "victim.txt"
        victim.write_bytes(b"precious\n")
        marker = Path(self.tmp.name) / "release"
        script = f"import os,time\nwhile not os.path.exists({str(marker)!r}): time.sleep(0.05)"
        holder = self.acquire(1, "holder", sys.executable, "-c", script)
        self.assertTrue(wait_for(lambda: len(held(self.folder)) == 1))
        (self.folder / "waiting" / "00000000000000000000-1").symlink_to(victim)
        nxt = self.acquire(1, "next", "true")
        self.assertTrue(wait_for(lambda: len(list((self.folder / "waiting").iterdir())) == 2))
        marker.write_text("")
        out = nxt.communicate(timeout=30)[0]
        holder.communicate(timeout=30)
        self.assertEqual(nxt.returncode, 0, out)
        self.assertIn("waiting 1 of 1 for 1 slot(s)", out, "the link is not counted as a waiter")
        self.assertEqual(victim.read_bytes(), b"precious\n")

    def test_a_signal_before_the_command_starts_is_passed_on_once_it_does(self):
        """Audit N1: the held-signal path. SIGTERM reaches the wrapper inside Popen, before
        the command exists (no timing: the patched Popen sends it), and must still end the
        command; without the hand-over the command would sleep for an hour."""
        driver = (
            "import os, signal, subprocess, sys\n"
            f"sys.path.insert(0, {str(TOOL.parent)!r})\n"
            "import heavy_check_slot as h\n"
            "real = subprocess.Popen\n"
            "def popen(*a, **k):\n"
            "    os.kill(os.getpid(), signal.SIGTERM)\n"
            "    return real(*a, **k)\n"
            "h.subprocess.Popen = popen\n"
            f"sys.exit(h.main(['acquire', '--max', '1', '--name', 'early', '--dir', {str(self.folder)!r},"
            " '--', 'sleep', '3600']))\n")
        p = subprocess.Popen([sys.executable, "-I", "-c", driver], stdin=subprocess.DEVNULL,
                             stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True,
                             start_new_session=True)
        try:
            out = p.communicate(timeout=30)[0]
        except subprocess.TimeoutExpired:
            os.killpg(p.pid, signal.SIGKILL)
            out = p.communicate()[0]
            self.fail(f"the early SIGTERM never reached the command: {out!r}")
        self.assertEqual(p.returncode, 128 + signal.SIGTERM, out)
        self.assertEqual(held(self.folder), [])

    def test_usage_errors_exit_two(self):
        for args in (["--max", "0", "--name", "x", "--", "true"], ["--max", "1", "--name", "x", "--"]):
            r = subprocess.run(tool(self.folder, "acquire", *args), capture_output=True)
            self.assertEqual(r.returncode, 2, args)


if __name__ == "__main__":
    unittest.main()
