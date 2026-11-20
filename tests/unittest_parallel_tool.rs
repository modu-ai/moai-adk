//! SPEC-20261120-parallel-python-tests: `tools/unittest_parallel.py` runs the Python test modules
//! of a folder in parallel, one `python3 -m unittest discover` process per module, and prints the
//! summary of a serial run. The tests build a folder of fixture modules in a temporary directory
//! and run the tool on it; no real HOME and no repository module runs here.

use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output};

const TOOL: &str = "tools/unittest_parallel.py";

fn tool() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR")).join(TOOL)
}

/// A fixture module `<name>.py` in `dir` with `body` as its source.
fn module(dir: &Path, name: &str, body: &str) {
    fs::write(dir.join(format!("{name}.py")), body).unwrap();
}

/// Runs the tool on `dir` with `args`; `home` is the HOME the caller has.
fn run(dir: &Path, home: &Path, args: &[&str], env: &[(&str, &str)]) -> Output {
    let mut cmd = Command::new("python3");
    cmd.arg("-I")
        .arg(tool())
        .args(["-s"])
        .arg(dir)
        .args(["-p", "*_test.py"])
        .args(args)
        .current_dir(dir)
        .env_clear()
        .env("PATH", std::env::var_os("PATH").unwrap())
        .env("HOME", home)
        .env("PYTHONDONTWRITEBYTECODE", "1")
        .env("CARGO_HOME", "/caller/cargo")
        .env("RUSTUP_HOME", "/caller/rustup");
    for (k, v) in env {
        cmd.env(k, v);
    }
    cmd.output().unwrap()
}

fn text(out: &Output) -> String {
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

const PASS: &str = "import unittest\nclass T(unittest.TestCase):\n    def test_a(self):\n        pass\n    def test_b(self):\n        pass\n";
const FAIL: &str = "import unittest\nclass T(unittest.TestCase):\n    def test_bad(self):\n        self.assertEqual(1, 2)\n    def test_boom(self):\n        raise RuntimeError('boom')\n    def test_fine(self):\n        pass\n";
const SKIP: &str = "import unittest\nclass T(unittest.TestCase):\n    @unittest.skip('not here')\n    def test_skipped(self):\n        pass\n    def test_ok(self):\n        pass\n";

fn fixture() -> (tempfile::TempDir, tempfile::TempDir) {
    (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap())
}

#[test]
fn the_tool_sums_the_modules_and_prints_ok() {
    let (dir, home) = fixture();
    module(dir.path(), "a_test", PASS);
    module(dir.path(), "b_test", SKIP);
    module(dir.path(), "notes", "raise SystemExit('never imported')\n");
    // A module with no case is not a failure (`NO TESTS RAN` exits 5 on new Pythons).
    module(dir.path(), "empty_test", "import unittest\n");
    let out = run(dir.path(), home.path(), &[], &[]);
    let all = text(&out);
    assert_eq!(out.status.code(), Some(0), "{all}");
    let lines: Vec<&str> = all.lines().filter(|l| !l.is_empty()).collect();
    let ran = lines.iter().rev().nth(1).copied().unwrap_or_default();
    assert!(
        ran.starts_with("Ran 4 tests in ") && ran.ends_with('s'),
        "summary line: {ran:?}\n{all}"
    );
    assert!(lines.last().unwrap().starts_with("OK"), "last line: {all}");
    assert!(all.contains("skipped=1"), "the skip count is kept: {all}");
}

#[test]
fn a_failing_module_fails_the_run_and_the_rest_still_run() {
    let (dir, home) = fixture();
    module(dir.path(), "a_test", PASS);
    module(dir.path(), "b_test", FAIL);
    module(dir.path(), "c_test", PASS);
    let out = run(dir.path(), home.path(), &[], &[]);
    let all = text(&out);
    assert_eq!(out.status.code(), Some(1), "{all}");
    assert!(
        all.contains("AssertionError: 1 != 2"),
        "failure detail: {all}"
    );
    assert!(all.contains("RuntimeError: boom"), "error detail: {all}");
    assert!(
        all.contains("Ran 7 tests in "),
        "the other modules ran: {all}"
    );
    let last = all.lines().rev().find(|l| !l.is_empty()).unwrap();
    assert_eq!(last, "FAILED (failures=1, errors=1)", "{all}");
    assert!(!all.contains("\nOK\n"), "{all}");
}

#[test]
fn a_module_that_cannot_be_imported_fails_the_run() {
    let (dir, home) = fixture();
    module(dir.path(), "a_test", PASS);
    module(
        dir.path(),
        "broken_test",
        "import module_that_does_not_exist\n",
    );
    let out = run(dir.path(), home.path(), &[], &[]);
    let all = text(&out);
    assert_eq!(out.status.code(), Some(1), "{all}");
    assert!(all.contains("module_that_does_not_exist"), "{all}");
    assert!(all.contains("FAILED ("), "{all}");
}

#[test]
fn verbose_keeps_the_per_test_line_shape() {
    let (dir, home) = fixture();
    module(dir.path(), "a_test", PASS);
    module(dir.path(), "b_test", SKIP);
    module(dir.path(), "c_test", FAIL);
    let out = run(dir.path(), home.path(), &["-v"], &[]);
    let all = text(&out);
    for line in [
        "test_a (a_test.T.test_a) ... ok",
        "test_b (a_test.T.test_b) ... ok",
        "test_ok (b_test.T.test_ok) ... ok",
        "test_skipped (b_test.T.test_skipped) ... skipped 'not here'",
        "test_bad (c_test.T.test_bad) ... FAIL",
        "test_boom (c_test.T.test_boom) ... ERROR",
    ] {
        assert!(all.lines().any(|l| l == line), "no line {line:?} in\n{all}");
    }
    // Each module's lines stay together: its three test lines are consecutive.
    let at: Vec<usize> = all
        .lines()
        .enumerate()
        .filter(|(_, l)| l.contains("(c_test.T.") && l.contains(" ... "))
        .map(|(i, _)| i)
        .collect();
    assert_eq!(at.len(), 3, "{all}");
    assert_eq!(at[2] - at[0], 2, "the module's lines are one block: {all}");
}

/// A module that records how many modules run at the same moment: it creates a marker in
/// `running/`, waits, counts the markers and leaves its count in `seen/<name>`.
fn overlap_module(dir: &Path, name: &str) {
    let body = format!(
        "import os, time, unittest\n\
         HERE = os.path.dirname(os.path.abspath(__file__))\n\
         class T(unittest.TestCase):\n\
         \x20   def test_overlap(self):\n\
         \x20       run = os.path.join(HERE, 'running'); seen = os.path.join(HERE, 'seen')\n\
         \x20       os.makedirs(run, exist_ok=True); os.makedirs(seen, exist_ok=True)\n\
         \x20       mine = os.path.join(run, '{name}')\n\
         \x20       open(mine, 'w').close()\n\
         \x20       time.sleep(1.5)\n\
         \x20       n = len(os.listdir(run))\n\
         \x20       open(os.path.join(seen, '{name}'), 'w').write(str(n))\n\
         \x20       time.sleep(0.3)\n\
         \x20       os.remove(mine)\n"
    );
    module(dir, name, &body);
}

fn most_seen(dir: &Path) -> usize {
    fs::read_dir(dir.join("seen"))
        .unwrap()
        .map(|e| {
            fs::read_to_string(e.unwrap().path())
                .unwrap()
                .parse::<usize>()
                .unwrap()
        })
        .max()
        .unwrap()
}

#[test]
fn the_job_count_bounds_the_overlap() {
    let names = ["o1_test", "o2_test", "o3_test"];
    // -j 1: one module at a time.
    let (dir, home) = fixture();
    for n in names {
        overlap_module(dir.path(), n);
    }
    let out = run(dir.path(), home.path(), &["-j", "1"], &[]);
    assert_eq!(out.status.code(), Some(0), "{}", text(&out));
    assert_eq!(most_seen(dir.path()), 1, "-j 1 overlapped");
    // -j 3: the three modules overlap.
    let (dir, home) = fixture();
    for n in names {
        overlap_module(dir.path(), n);
    }
    let out = run(dir.path(), home.path(), &["-j", "3"], &[]);
    assert_eq!(out.status.code(), Some(0), "{}", text(&out));
    assert!(most_seen(dir.path()) >= 2, "-j 3 did not overlap");
    // MOAI_UNITTEST_JOBS is read when -j is absent, and -j wins over it.
    let (dir, home) = fixture();
    for n in names {
        overlap_module(dir.path(), n);
    }
    let out = run(dir.path(), home.path(), &[], &[("MOAI_UNITTEST_JOBS", "1")]);
    assert_eq!(out.status.code(), Some(0), "{}", text(&out));
    assert_eq!(most_seen(dir.path()), 1, "MOAI_UNITTEST_JOBS=1 overlapped");
    let (dir, home) = fixture();
    for n in names {
        overlap_module(dir.path(), n);
    }
    let out = run(
        dir.path(),
        home.path(),
        &["-j", "3"],
        &[("MOAI_UNITTEST_JOBS", "1")],
    );
    assert_eq!(out.status.code(), Some(0), "{}", text(&out));
    assert!(
        most_seen(dir.path()) >= 2,
        "-j did not win over the variable"
    );
}

/// A module of more than 32 KiB with two test classes, `B` inheriting the test of `A`; `hold`
/// records how many cases run at the same moment, as `overlap_module` does.
fn big_module(dir: &Path, name: &str, extra: &str) {
    let body = format!(
        "import os, time, unittest\n\
         # {pad}\n\
         HERE = os.path.dirname(os.path.abspath(__file__))\n\
         def hold(tag):\n\
         \x20   run = os.path.join(HERE, 'running'); seen = os.path.join(HERE, 'seen')\n\
         \x20   os.makedirs(run, exist_ok=True); os.makedirs(seen, exist_ok=True)\n\
         \x20   mine = os.path.join(run, tag + str(os.getpid()))\n\
         \x20   open(mine, 'w').close()\n\
         \x20   time.sleep(1.5)\n\
         \x20   n = len(os.listdir(run))\n\
         \x20   with open(os.path.join(seen, tag + str(os.getpid())), 'w') as f:\n\
         \x20       f.write(str(n))\n\
         \x20   time.sleep(0.3)\n\
         \x20   os.remove(mine)\n\
         class A(unittest.TestCase):\n\
         \x20   def test_a(self):\n\
         \x20       hold('a')\n\
         class B(A):\n\
         \x20   def test_b(self):\n\
         \x20       hold('b')\n\
         {extra}",
        pad = "#".repeat(33_000)
    );
    module(dir, name, &body);
}

#[test]
fn a_big_module_is_split_by_class() {
    let (dir, home) = fixture();
    big_module(dir.path(), "big_test", "");
    module(dir.path(), "small_test", PASS);
    let out = run(dir.path(), home.path(), &["-v", "-j", "2"], &[]);
    let all = text(&out);
    assert_eq!(out.status.code(), Some(0), "{all}");
    // A has one case, B its own and the one it inherits; the split loses and repeats none.
    assert!(all.contains("Ran 5 tests in "), "{all}");
    for line in [
        "test_a (big_test.A.test_a) ... ok",
        "test_a (big_test.B.test_a) ... ok",
        "test_b (big_test.B.test_b) ... ok",
    ] {
        assert!(all.lines().any(|l| l == line), "no line {line:?} in\n{all}");
    }
    assert!(
        all.contains("modules: 2, processes: 3,"),
        "the big module runs as one process per class: {all}"
    );
    assert!(
        most_seen(dir.path()) >= 2,
        "the two classes did not overlap"
    );
}

#[test]
fn a_big_module_with_a_load_tests_hook_runs_whole() {
    let (dir, home) = fixture();
    big_module(
        dir.path(),
        "hook_test",
        // The hook adds a case that belongs to no class of the module, so a split by class
        // would not know where it runs.
        "def load_tests(loader, tests, pattern):\n    tests.addTest(unittest.FunctionTestCase(lambda: None))\n    return tests\n",
    );
    let out = run(dir.path(), home.path(), &["-j", "2"], &[]);
    let all = text(&out);
    assert_eq!(out.status.code(), Some(0), "{all}");
    assert!(all.contains("Ran 4 tests in "), "{all}");
    assert!(all.contains("modules: 1, processes: 1,"), "{all}");
}

#[test]
fn each_module_gets_its_own_home_and_tmpdir() {
    let (dir, home) = fixture();
    for n in ["h1_test", "h2_test"] {
        let body = format!(
            "import os, unittest\n\
             HERE = os.path.dirname(os.path.abspath(__file__))\n\
             class T(unittest.TestCase):\n\
             \x20   def test_env(self):\n\
             \x20       keys = ['HOME', 'TMPDIR', 'CARGO_HOME', 'RUSTUP_HOME']\n\
             \x20       with open(os.path.join(HERE, '{n}.env'), 'w') as f:\n\
             \x20           f.write('\\n'.join(k + '=' + os.environ.get(k, '') for k in keys))\n\
             \x20       self.assertTrue(os.path.isdir(os.environ['HOME']))\n\
             \x20       self.assertTrue(os.path.isdir(os.environ['TMPDIR']))\n"
        );
        module(dir.path(), n, &body);
    }
    let out = run(dir.path(), home.path(), &[], &[]);
    assert_eq!(out.status.code(), Some(0), "{}", text(&out));
    let read = |n: &str| -> Vec<String> {
        fs::read_to_string(dir.path().join(format!("{n}.env")))
            .unwrap()
            .lines()
            .map(str::to_string)
            .collect()
    };
    let (a, b) = (read("h1_test"), read("h2_test"));
    let caller_home = format!("HOME={}", home.path().display());
    assert_ne!(a[0], b[0], "the two modules share a HOME");
    assert_ne!(a[1], b[1], "the two modules share a TMPDIR");
    for m in [&a, &b] {
        assert_ne!(m[0], caller_home, "a module got the caller's HOME");
        assert_ne!(m[0], "HOME=", "HOME is empty");
        assert_eq!(m[2], "CARGO_HOME=/caller/cargo");
        assert_eq!(m[3], "RUSTUP_HOME=/caller/rustup");
    }
    // Without CARGO_HOME and RUSTUP_HOME the caller's own defaults are kept.
    let (dir, home) = fixture();
    module(
        dir.path(),
        "d_test",
        "import os, unittest\nclass T(unittest.TestCase):\n    def test_d(self):\n        open(os.path.join(os.path.dirname(os.path.abspath(__file__)), 'd.env'), 'w').write(os.environ.get('CARGO_HOME', '') + '|' + os.environ.get('RUSTUP_HOME', ''))\n",
    );
    let mut cmd = Command::new("python3");
    cmd.arg("-I")
        .arg(tool())
        .args(["-s"])
        .arg(dir.path())
        .args(["-p", "*_test.py"])
        .current_dir(dir.path())
        .env_clear()
        .env("PATH", std::env::var_os("PATH").unwrap())
        .env("HOME", home.path())
        .env("PYTHONDONTWRITEBYTECODE", "1");
    let out = cmd.output().unwrap();
    assert_eq!(out.status.code(), Some(0), "{}", text(&out));
    let seen = fs::read_to_string(dir.path().join("d.env")).unwrap();
    let h = home.path().display();
    assert_eq!(
        seen,
        format!("{h}/.cargo|{h}/.rustup"),
        "defaults of the caller"
    );
}

#[test]
fn no_matching_module_fails() {
    let (dir, home) = fixture();
    let out = run(dir.path(), home.path(), &[], &[]);
    let all = text(&out);
    assert_eq!(out.status.code(), Some(1), "{all}");
    assert!(all.contains("no test module matched"), "{all}");
}

#[test]
fn local_check_runs_the_tool() {
    let script =
        fs::read_to_string(PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("tools/local_check.sh"))
            .unwrap();
    let line = script
        .lines()
        .find(|l| l.contains("step python unittest"))
        .expect("local_check.sh has no unittest step");
    assert!(
        line.contains("python3 tools/unittest_parallel.py -v -s tests -p '*_test.py' 2>&1 | tee ci-unittest.log"),
        "{line}"
    );
    assert!(line.contains("PIPESTATUS[0]"), "{line}");
    assert!(
        !script.contains("python3 -m unittest discover -v -s tests"),
        "the serial command is still in local_check.sh"
    );
}
