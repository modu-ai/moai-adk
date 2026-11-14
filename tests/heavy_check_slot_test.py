"""tools/heavy_check_slot.py (Issue #2491): real processes, a temporary slot folder only."""
import json
import re
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


ROOMY = "free_gb=1000,swap_pct=0"


def roomy_machine(test):
    """Every slot test runs on an injected roomy machine (#2829): the real disk of a busy
    runner or lane machine must not turn a slot test into a 75 'machine busy' exit."""
    before = os.environ.get("MOAI_CLI_TEST_MACHINE")
    os.environ["MOAI_CLI_TEST_MACHINE"] = ROOMY

    def restore():
        if before is None:
            os.environ.pop("MOAI_CLI_TEST_MACHINE", None)
        else:
            os.environ["MOAI_CLI_TEST_MACHINE"] = before

    test.addCleanup(restore)


@unittest.skipUnless(UNIX, "flock slots are unix only")
class HeavyCheckSlotTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.folder = Path(self.tmp.name) / "slots"
        self.addCleanup(self.tmp.cleanup)
        roomy_machine(self)

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


ROOT = Path(__file__).resolve().parent.parent
REASON = re.compile(r"not started — (.*); load ")


def padded(n):
    """A valid limits file of exactly `n` bytes."""
    head = '{"max":2}'
    return head + " " * (n - len(head))


# Limits files both tools refuse with 125 (#2930 R-2); the symlink and folder cases plant
# something other than text and are run on their own.
BAD_LIMITS_FILES = [
    ("max 0", '{"max":0}'),
    ("unknown key", '{"max":2,"min_free":1}'),
    ("invalid JSON", "{max:2"),
    ("not an object", "[2,50,80]"),
    ("negative", '{"min_free_gb":-1}'),
    ("not a whole number", '{"max_swap_pct":"80"}'),
    ("4097 bytes", padded(4097)),
]


def rust_bin():
    """The debug `moai-cli` the cargo steps built (debug builds read MOAI_CLI_TEST_MACHINE)."""
    target = os.environ.get("CARGO_TARGET_DIR")
    base = Path(target) if target else ROOT / "target"
    return base / "debug" / "moai-cli"


@unittest.skipUnless(UNIX, "flock slots are unix only")
class MachineLimitTest(unittest.TestCase):
    """#2829: free disk and swap limits, the same rule and words as src/heavy_slot.rs."""

    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(os.path.realpath(self.tmp.name))
        self.folder = self.root / "slots"
        self.marker = self.root / "ran"

    def run_python(self, machine, *extra, name="lane"):
        env = dict(os.environ, MOAI_CLI_TEST_MACHINE=machine)
        args = tool(self.folder, "acquire", "--max", "1", "--name", name, *extra,
                    "--", sys.executable, "-c", f"open({str(self.marker)!r}, 'w')")
        return subprocess.run(args, capture_output=True, text=True, env=env, cwd=self.root,
                              timeout=30)

    def run_rust(self, machine, *extra, name="lane"):
        env = dict(os.environ, MOAI_CLI_TEST_MACHINE=machine)
        args = [str(rust_bin()), "heavy-slot-acquire", "--max", "1", "--name", name,
                "--dir", str(self.folder), *extra,
                "--", sys.executable, "-c", f"open({str(self.marker)!r}, 'w')"]
        return subprocess.run(args, capture_output=True, text=True, env=env, cwd=self.root,
                              timeout=30)

    def ran(self):
        found = self.marker.exists()
        if found:
            self.marker.unlink()
        return found

    def test_ac1_free_disk_boundary_49_50_51_gb(self):
        r = self.run_python("free_gb=49,swap_pct=0")
        self.assertEqual(r.returncode, 75, r.stderr)
        self.assertFalse(self.ran(), "COMMAND must not run below the limit")
        self.assertIn(f"disk free 49.0 GB at {self.root} is below 50 GB", r.stderr)
        self.assertIn("COMMAND did not run, try again later", r.stderr)
        self.assertEqual(held(self.folder) if self.folder.exists() else [], [])
        for free in (50, 51):
            r = self.run_python(f"free_gb={free},swap_pct=0")
            self.assertEqual(r.returncode, 0, (free, r.stderr))
            self.assertTrue(self.ran(), free)

    def test_swap_boundary_80_81_and_both_reasons(self):
        r = self.run_python("free_gb=1000,swap_pct=80")
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertTrue(self.ran())
        r = self.run_python("free_gb=1000,swap_pct=81")
        self.assertEqual(r.returncode, 75, r.stderr)
        self.assertIn("swap use 81% is above 80%", r.stderr)
        self.assertFalse(self.ran())
        r = self.run_python("free_gb=10,swap_pct=95")
        self.assertEqual(r.returncode, 75)
        self.assertRegex(r.stderr, r"is below 50 GB; swap use 95% is above 80%")

    def limits(self, body):
        """Write `<slots>/limits.json` (#2930 R-2), or remove it for None."""
        path = self.folder / "limits.json"
        if body is None:
            if path.is_file() and not path.is_symlink():
                path.unlink()
            return
        self.folder.mkdir(exist_ok=True)
        path.write_text(body)

    def test_limits_can_be_set_or_turned_off_and_unknown_values_do_not_refuse(self):
        r = self.run_python("free_gb=60,swap_pct=0", "--min-free-gb", "61")
        self.assertEqual(r.returncode, 75, r.stderr)
        self.assertIn("is below 61 GB", r.stderr)
        # #2930 R-2: the old flags alone no longer turn the limits off; the file does.
        r = self.run_python("free_gb=1,swap_pct=0", "--min-free-gb", "0", "--max-swap-pct", "100")
        self.assertEqual(r.returncode, 75, r.stderr)
        self.assertIn("is below 50 GB", r.stderr)
        self.assertFalse(self.ran())
        self.limits('{"min_free_gb":0,"max_swap_pct":100}')
        r = self.run_python("free_gb=0,swap_pct=100")
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertTrue(self.ran())
        self.limits(None)
        r = self.run_python("nothing=1")
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertTrue(self.ran())

    def test_d2_the_machine_is_judged_again_after_a_wait(self):
        machine = self.root / "machine"
        machine.write_text(ROOMY)
        env = dict(os.environ, MOAI_CLI_TEST_MACHINE=f"@{machine}")
        hold = subprocess.Popen(
            tool(self.folder, "acquire", "--max", "1", "--name", "holder", "--",
                 sys.executable, "-c", "import time; time.sleep(30)"),
            env=env, cwd=self.root, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
            start_new_session=True)
        self.addCleanup(lambda: hold.poll() is None and os.killpg(hold.pid, signal.SIGKILL))
        self.assertTrue(wait_for(lambda: "holder" in holder_names(self.folder)))
        waiter = subprocess.Popen(
            tool(self.folder, "acquire", "--max", "1", "--name", "waiter", "--",
                 sys.executable, "-c", f"open({str(self.marker)!r}, 'w')"),
            env=env, cwd=self.root, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
        waiting = self.folder / "waiting"
        self.assertTrue(wait_for(lambda: waiting.is_dir() and any(waiting.iterdir())))
        machine.write_text("free_gb=5,swap_pct=0")
        os.killpg(hold.pid, signal.SIGKILL)
        out, err = waiter.communicate(timeout=30)
        self.assertEqual(waiter.returncode, 75, (out, err))
        self.assertIn("disk free 5.0 GB", err)
        self.assertFalse(self.ran(), "COMMAND must not run when the wait ate the space")
        self.assertTrue(wait_for(lambda: held(self.folder) == []), "the slot is freed")

    def test_ac2_python_and_rust_judge_the_same_inputs_the_same_way(self):
        exe = rust_bin()
        if not exe.exists():
            raise AssertionError(f"moai-cli not built at {exe}; run cargo build first")
        cases = [
            ("free_gb=49,swap_pct=0", (), None),
            ("free_gb=50,swap_pct=0", (), None),
            ("free_gb=51,swap_pct=0", (), None),
            ("free_gb=1000,swap_pct=80", (), None),
            ("free_gb=1000,swap_pct=81", (), None),
            ("free_gb=10,swap_pct=95", (), None),
            ("free_gb=60,swap_pct=0", ("--min-free-gb", "61"), None),
            ("free_gb=0,swap_pct=100", ("--min-free-gb", "0", "--max-swap-pct", "100"), None),
            ("free_gb=0,swap_pct=100", (), '{"min_free_gb":0,"max_swap_pct":100}'),
            ("free_gb=49,swap_pct=0", ("--min-free-gb", "0"), '{"min_free_gb":50}'),
            ("free_gb=50,swap_pct=0", ("--min-free-gb", "0"), '{"min_free_gb":50}'),
            ("free_gb=59,swap_pct=0", (), '{"min_free_gb":60}'),
            ("free_gb=60,swap_pct=0", (), '{"min_free_gb":60}'),
            ("free_gb=60,swap_pct=0", ("--min-free-gb", "61"), '{"min_free_gb":50}'),
            ("free_gb=61,swap_pct=0", ("--min-free-gb", "61"), '{"min_free_gb":50}'),
            ("free_gb=1000,swap_pct=81", ("--max-swap-pct", "100"), '{"max_swap_pct":80}'),
            ("free_gb=1000,swap_pct=80", ("--max-swap-pct", "100"), '{"max_swap_pct":80}'),
            ("nothing=1", (), None),
        ] + [(ROOMY, (), body) for _, body in BAD_LIMITS_FILES]
        for machine, extra, body in cases:
            self.limits(body)
            py = self.run_python(machine, *extra)
            py_ran = self.ran()
            rs = self.run_rust(machine, *extra)
            rs_ran = self.ran()
            self.assertEqual(py.returncode, rs.returncode, (machine, extra, py.stderr, rs.stderr))
            self.assertEqual(py_ran, rs_ran, (machine, extra))
            py_why, rs_why = REASON.search(py.stderr), REASON.search(rs.stderr)
            self.assertEqual(bool(py_why), bool(rs_why), (machine, py.stderr, rs.stderr))
            if py_why:
                self.assertEqual(py_why.group(1), rs_why.group(1), machine)

    def test_a_limits_file_wins_when_stricter(self):
        """#2930 R-2: each pair is the edge — the first value refused, then the first that
        starts."""
        for body, extra, refused, starts, says in [
            ('{"min_free_gb":50}', ("--min-free-gb", "0"), 49, 50, "below 50 GB"),
            ('{"min_free_gb":60}', (), 59, 60, "below 60 GB"),
            ('{"min_free_gb":50}', ("--min-free-gb", "61"), 60, 61, "below 61 GB"),
        ]:
            self.limits(body)
            r = self.run_python(f"free_gb={refused},swap_pct=0", *extra)
            self.assertEqual(r.returncode, 75, (body, extra, r.stderr))
            self.assertIn(says, r.stderr)
            self.assertFalse(self.ran())
            r = self.run_python(f"free_gb={starts},swap_pct=0", *extra)
            self.assertEqual(r.returncode, 0, (body, extra, r.stderr))
            self.assertTrue(self.ran())
        self.limits('{"max_swap_pct":80}')
        r = self.run_python("free_gb=1000,swap_pct=81", "--max-swap-pct", "100")
        self.assertEqual(r.returncode, 75, r.stderr)
        self.assertIn("swap use 81% is above 80%", r.stderr)
        self.assertIn("--max-swap-pct 100 is looser than limits.json (80)", r.stderr)
        r = self.run_python("free_gb=1000,swap_pct=80", "--max-swap-pct", "100")
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertTrue(self.ran())

    def test_a_max_in_the_file_beats_a_larger_caller_max(self):
        self.limits('{"max":1}')
        env = dict(os.environ, MOAI_CLI_TEST_MACHINE=ROOMY)
        hold = subprocess.Popen(
            tool(self.folder, "acquire", "--max", "2", "--name", "holder", "--",
                 sys.executable, "-c", "import time; time.sleep(30)"),
            env=env, cwd=self.root, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
            start_new_session=True)
        self.addCleanup(lambda: hold.poll() is None and os.killpg(hold.pid, signal.SIGKILL))
        self.assertTrue(wait_for(lambda: "holder" in holder_names(self.folder)))
        waiter = subprocess.Popen(
            tool(self.folder, "acquire", "--max", "2", "--name", "waiter", "--",
                 sys.executable, "-c", f"open({str(self.marker)!r}, 'w')"),
            env=env, cwd=self.root, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
        waiting = self.folder / "waiting"
        self.assertTrue(wait_for(lambda: waiting.is_dir() and any(waiting.iterdir())))
        self.assertFalse(self.marker.exists(), "the waiter ran while the only slot was held")
        self.assertFalse((self.folder / "slot-1").exists(), "slot 1 was created")
        os.killpg(hold.pid, signal.SIGKILL)
        out, err = waiter.communicate(timeout=30)
        self.assertEqual(waiter.returncode, 0, (out, err))
        self.assertIn("waiting 1 of 1 for 1 slot(s)", out)
        self.assertTrue(self.ran())

    def test_a_bad_limits_file_runs_nothing(self):
        valid = self.root / "valid.json"
        valid.write_text('{"min_free_gb":0}')
        path = self.folder / "limits.json"
        plants = [(what, lambda b=body: path.write_text(b)) for what, body in BAD_LIMITS_FILES]
        plants += [("symlink to a valid file", lambda: path.symlink_to(valid)),
                   ("directory", lambda: path.mkdir())]
        for what, plant in plants:
            with self.subTest(what):
                with tempfile.TemporaryDirectory() as d:
                    self.folder = Path(os.path.realpath(d)) / "slots"
                    self.folder.mkdir()
                    path = self.folder / "limits.json"
                    plant()
                    r = self.run_python(ROOMY)
                    self.assertEqual(r.returncode, 125, (what, r.stderr))
                    self.assertIn("heavy-check-slot: refused", r.stderr)
                    self.assertIn("COMMAND did not run", r.stderr)
                    self.assertNotIn("Traceback", r.stderr)
                    self.assertFalse(self.ran(), what)
                    self.assertFalse((self.folder / "slot-0").exists(), what)
                    self.assertFalse((self.folder / "waiting").exists(), what)
        self.folder = self.root / "slots"
        self.limits(padded(4096))
        r = self.run_python(ROOMY)
        self.assertEqual(r.returncode, 0, ("4096 bytes", r.stderr))
        self.assertTrue(self.ran())


if __name__ == "__main__":
    unittest.main()
