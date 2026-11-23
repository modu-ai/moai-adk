//! Hygiene lints for the test suite and recorded verification commands (#2352).
//!
//! - git fixture calls must go through `tests/common/git.rs`: a test file that
//!   spawns git outside that helper (and without the gc/maintenance guards in its
//!   wrapper) misses the guards that keep fixture repositories stable
//!   (#2082/#2092), and the scan points new tests at the shared helper.
//! - recorded `cargo test --test X …` verification commands must name an existing
//!   target and select at least one test — cargo exits 0 with "running 0 tests"
//!   for a filter that matches nothing, which has silently passed as verification
//!   more than once (survey 2.11, s01-scope-05).

mod common;

use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;

const MANIFEST: &str = env!("CARGO_MANIFEST_DIR");

fn tests_dir() -> PathBuf {
    Path::new(MANIFEST).join("tests")
}

fn rs_sources(dir: &Path) -> Vec<PathBuf> {
    let mut out = Vec::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(dir) = stack.pop() {
        let entries = match fs::read_dir(&dir) {
            Ok(entries) => entries,
            Err(_) => continue,
        };
        for entry in entries.flatten() {
            let path = entry.path();
            if path.is_dir() {
                stack.push(path);
            } else if path.extension().is_some_and(|e| e == "rs") {
                out.push(path);
            }
        }
    }
    out.sort();
    out
}

/// Files that carry their own git spawns must guard them: a wrapper that spawns git
/// without the fixture guards is exactly what this card routes into
/// `tests/common/git.rs`. This file is exempt — it quotes the pattern it scans for.
fn files_with_unguarded_git(dir: &Path) -> Vec<String> {
    let mut violations = Vec::new();
    for path in rs_sources(dir) {
        let rel = path
            .strip_prefix(dir)
            .unwrap()
            .to_string_lossy()
            .to_string();
        if rel.starts_with("common/") || rel == "ac_command_lint.rs" {
            continue;
        }
        let text =
            fs::read_to_string(&path).unwrap_or_else(|e| panic!("read {}: {e}", path.display()));
        let has_spawn = text.contains("Command::new(\"git\")");
        let guarded = text.contains("gc.auto=0") && text.contains("maintenance.auto=false");
        if has_spawn && !guarded {
            violations.push(format!("tests/{rel}"));
        }
    }
    violations
}

#[test]
fn git_fixture_calls_live_in_the_shared_helper() {
    let violations = files_with_unguarded_git(&tests_dir());
    assert!(
        violations.is_empty(),
        "test files spawning git without the gc/maintenance fixture guards — route \
         them through tests/common/git.rs or add the guards:\n{}",
        violations.join("\n")
    );
}

/// Behavioural check: the guards must reach git itself, not just appear in the
/// helper's text (its doc comment names them, so a text match alone survives a
/// mutation of the actual argument values).
#[test]
fn git_helper_applies_the_guards_to_every_call() {
    let dir = tempfile::TempDir::new().unwrap();
    common::git::git_ok(dir.path(), &["init", "-q"]);
    assert_eq!(
        common::git::git(dir.path(), &["config", "--get", "gc.auto"]),
        "0",
        "git() must run with gc.auto=0"
    );
    assert_eq!(
        common::git::git(dir.path(), &["config", "--get", "maintenance.auto"]),
        "false",
        "git() must run with maintenance.auto=false"
    );
    let raw = common::git::git_raw(dir.path(), &["config", "--get", "gc.auto"]);
    assert_eq!(raw.stdout.trim(), "0", "git_raw() must run with gc.auto=0");
    // git_pinned only asserts success; a value pattern makes `git config --get`
    // exit 1 unless the shared factory supplied exactly the expected guard.
    common::git::git_pinned(dir.path(), &["config", "--get", "gc.auto", "^0$"], None);
    common::git::git_pinned(
        dir.path(),
        &["config", "--get", "maintenance.auto", "^false$"],
        None,
    );
    let global = common::git::git_raw(dir.path(), &["config", "--global", "--list"]);
    assert!(
        global.stdout.trim().is_empty(),
        "GIT_CONFIG_GLOBAL must point at an empty config: {}",
        global.stdout
    );
}

/// Runs the recorded-AC-command lint with the given arguments and returns its exit code.
fn run_lint(args: &[&str]) -> std::process::Output {
    let script = Path::new(MANIFEST).join("tools/check_ac_commands.py");
    let mut cmd = Command::new("python3");
    cmd.arg("-I").arg(&script);
    for arg in args {
        cmd.arg(arg);
    }
    cmd.output().expect("spawn the AC command lint")
}

/// SPEC-20261122-ac-lint-no-cargo: the lint reads test names from the sources and starts no
/// cargo, so the nextest special case that made this binary run alone (#2661) has no reason.
#[test]
fn nextest_has_no_special_case_for_this_binary() {
    let config = fs::read_to_string(Path::new(MANIFEST).join(".config/nextest.toml"))
        .expect(".config/nextest.toml must exist");
    let special = config
        .split("[[profile.default.overrides]]")
        .skip(1)
        .any(|block| {
            let block = block.split("\n[").next().unwrap_or("");
            block.lines().any(|line| {
                line.split_once('=').is_some_and(|(k, v)| {
                    k.trim() == "filter" && v.contains("binary(ac_command_lint)")
                })
            })
        });
    assert!(
        !special,
        "no override for binary(ac_command_lint) may stay in .config/nextest.toml:\n{config}"
    );
}

/// A fake `cargo` first on PATH, which shadows the real one and records each start in the marker
/// file. Returns the folder that owns it (keep it alive), the marker and the PATH value to use.
#[cfg(unix)]
fn fake_cargo() -> (tempfile::TempDir, PathBuf, std::ffi::OsString) {
    use std::os::unix::fs::PermissionsExt;
    let bin = tempfile::TempDir::new().unwrap();
    let marker = bin.path().join("cargo-started");
    let fake = bin.path().join("cargo");
    fs::write(
        &fake,
        format!("#!/bin/sh\necho \"$@\" >> '{}'\nexit 1\n", marker.display()),
    )
    .unwrap();
    fs::set_permissions(&fake, fs::Permissions::from_mode(0o755)).unwrap();
    let path = std::env::var_os("PATH").unwrap_or_default();
    let path = std::env::join_paths(
        std::iter::once(bin.path().to_path_buf()).chain(std::env::split_paths(&path)),
    )
    .unwrap();
    (bin, marker, path)
}

/// Runs the lint with a fake `cargo` first on PATH, which shadows the real one and records each
/// start in `cargo-started`. Returns the output and whether cargo was started.
#[cfg(unix)]
fn run_lint_without_cargo(args: &[&str]) -> (std::process::Output, bool) {
    let (_bin, marker, path) = fake_cargo();
    let out = Command::new("python3")
        .arg("-I")
        .arg(Path::new(MANIFEST).join("tools/check_ac_commands.py"))
        .args(args)
        .env("PATH", path)
        .output()
        .expect("spawn the AC command lint");
    (out, marker.exists())
}

#[cfg(unix)]
fn text_of(out: &std::process::Output) -> String {
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

/// Runs the lint script unchanged under an interpreter audit hook. The hook writes the process
/// starts and the repository files opened to a report file, one line each (`spawn <event>` and
/// `open <path>`), so a test can bound the work the lint does by counting it, never by the clock.
/// The script itself is opened by the runner and is not counted.
#[cfg(unix)]
const AUDIT_RUNNER: &str = r#"
import os, runpy, sys
root, report, script = sys.argv[1:4]
# The separator keeps a sibling folder that shares the root's name as a prefix out of the count.
prefix = root.rstrip(os.sep) + os.sep
sys.argv = [script] + sys.argv[4:]
SPAWN = ("subprocess.Popen", "os.system", "os.exec", "os.posix_spawn", "os.spawn", "os.fork", "os.forkpty")
lines = []
def hook(event, args):
    if event == "open":
        # One event per file read: Python 3.9 reads through an opener, so `os.open` raises a second
        # `open` event for the same read, and that one has no mode string.
        path = args[0]
        if isinstance(path, int) or not isinstance(args[1], str):
            return
        path = os.fsdecode(path)
        if path.startswith(prefix) and path != script:
            lines.append("open " + path)
    elif event in SPAWN:
        lines.append("spawn " + event)
sys.addaudithook(hook)
code = 0
try:
    runpy.run_path(script, run_name="__main__")
except SystemExit as stop:
    code = stop.code if isinstance(stop.code, int) else (0 if stop.code is None else 1)
with open(report, "w") as out:
    out.write("\n".join(lines) + "\n")
raise SystemExit(code)
"#;

/// What one counted lint run did: the lint's own output, whether the fake cargo was started, the
/// repository files it opened (in order) and the process starts the audit hook saw.
#[cfg(unix)]
struct CountedRun {
    out: std::process::Output,
    cargo_started: bool,
    opened: Vec<String>,
    spawned: Vec<String>,
}

#[cfg(unix)]
fn run_lint_counted(script: &Path, root: &str) -> CountedRun {
    let (bin, marker, path) = fake_cargo();
    let report = bin.path().join("audit-report");
    let out = Command::new("python3")
        .arg("-I")
        .arg("-c")
        .arg(AUDIT_RUNNER)
        .arg(root)
        .arg(&report)
        .arg(script)
        .args(["--tree", root])
        .env("PATH", path)
        .output()
        .expect("spawn the AC command lint under the audit runner");
    let report = fs::read_to_string(&report).unwrap_or_default();
    let mut opened = Vec::new();
    let mut spawned = Vec::new();
    for line in report.lines() {
        match line.split_once(' ') {
            Some(("open", p)) => opened.push(p.to_string()),
            Some(("spawn", e)) => spawned.push(e.to_string()),
            _ => {}
        }
    }
    CountedRun {
        out,
        cargo_started: marker.exists(),
        opened,
        spawned,
    }
}

/// The recorded files the lint says it checked and the `--test` targets it listed, read from its
/// summary line `checked <n> recorded files, <m> --test targets listed`.
#[cfg(unix)]
fn checked_counts(text: &str) -> (usize, usize) {
    let line = text
        .lines()
        .find(|l| l.starts_with("checked "))
        .unwrap_or_else(|| panic!("the lint printed no summary line:\n{text}"));
    let words: Vec<&str> = line.split_whitespace().collect();
    let number = |i: usize| -> usize {
        words
            .get(i)
            .and_then(|w| w.parse().ok())
            .unwrap_or_else(|| panic!("unreadable summary line: {line}"))
    };
    (number(1), number(4))
}

/// Files opened more often than the lint's design allows, as `(path, opens)`. The lint reads each
/// recorded document once and each target root `tests/<name>.rs` once (one reading per target is
/// cached); a module file in a subfolder of `tests/` (`tests/common/*.rs`) is read once for every
/// target that declares it, so at most `targets` times.
#[cfg(unix)]
fn repeated_opens(opened: &[String], targets: usize) -> Vec<(String, usize)> {
    let mut counts = std::collections::BTreeMap::<&str, usize>::new();
    for path in opened {
        *counts.entry(path.as_str()).or_default() += 1;
    }
    let tests = format!("{MANIFEST}/tests/");
    counts
        .into_iter()
        .filter(|(path, n)| {
            let shared_module = path
                .strip_prefix(tests.as_str())
                .is_some_and(|rest| rest.contains('/'));
            *n > if shared_module { targets.max(1) } else { 1 }
        })
        .map(|(path, n)| (path.to_string(), n))
        .collect()
}

/// Files opened outside the three folders the lint scans (recorded documents, recorded reports and
/// the test sources), under `root`.
#[cfg(unix)]
fn outside_opens<'a>(opened: &'a [String], root: &str) -> Vec<&'a String> {
    let reports = format!("{}/reports", moai_cli::PROJECT_STATE_DIR);
    let allowed = ["docs", reports.as_str(), "tests"].map(|d| format!("{root}/{d}/"));
    opened
        .iter()
        .filter(|p| !allowed.iter().any(|a| p.starts_with(a.as_str())))
        .collect()
}

/// The lint is a pure reader that does a bounded amount of work, counted and not timed: it starts
/// no process of any kind (the 68-112 s cargo build of the earlier lint was one), it opens each
/// recorded file and each target root once and a shared test module once per target (see
/// `repeated_opens`), and it opens nothing outside the folders it scans. A clock limit failed at 5.18 s under machine load 200 and
/// could not tell a slow machine from a slow lint; a count does not move with load.
#[cfg(unix)]
#[test]
fn the_repository_lint_starts_no_process_and_reads_each_file_once() {
    let script = Path::new(MANIFEST).join("tools/check_ac_commands.py");
    let run = run_lint_counted(&script, MANIFEST);
    let text = text_of(&run.out);
    assert!(!run.cargo_started, "the lint must not start cargo:\n{text}");
    assert!(
        run.out.status.success(),
        "the recorded commands must pass without cargo:\n{text}"
    );
    assert!(
        run.spawned.is_empty(),
        "the lint started processes: {:?}\n{text}",
        run.spawned
    );
    let (files, targets) = checked_counts(&text);
    let repeated = repeated_opens(&run.opened, targets);
    assert!(
        repeated.is_empty(),
        "the lint opened files more often than its design allows: {repeated:?}"
    );
    let outside = outside_opens(&run.opened, MANIFEST);
    assert!(
        outside.is_empty(),
        "the lint opened files outside the three folders it scans: {outside:?}"
    );
    // A counter that saw nothing proves nothing: the lint read every recorded file it reports and
    // at least one test source for every target it listed.
    assert!(
        run.opened.len() >= files + targets,
        "the counter saw {} opens for {files} recorded files and {targets} targets:\n{text}",
        run.opened.len()
    );
}

/// The counter itself is checked by small stand-in scripts, so the bound above cannot go green
/// because the counter stopped counting: a script that starts a process, one that reads a recorded
/// file twice, one that reads outside the scanned folders, and one that reports checked files but
/// opens none each show up in exactly the count that the bound reads.
#[cfg(unix)]
#[test]
fn the_audit_counter_sees_a_process_a_second_read_a_stray_read_and_a_silent_run() {
    let tree = tempfile::TempDir::new().unwrap();
    let root = tree.path().to_str().unwrap().to_string();
    fs::create_dir_all(tree.path().join("docs")).unwrap();
    fs::write(tree.path().join(doc("a.md")), "x\n").unwrap();
    fs::write(tree.path().join("secret.txt"), "x\n").unwrap();
    let scripts = tempfile::TempDir::new().unwrap();
    let run_script = |name: &str, body: &str| {
        let path = scripts.path().join(name);
        fs::write(&path, body).unwrap();
        run_lint_counted(&path, &root)
    };

    let spawned = run_script(
        "spawn.py",
        "import subprocess\nsubprocess.run(['true'])\nprint('checked 0 recorded files, 0 --test targets listed')\n",
    );
    assert_eq!(
        spawned.spawned,
        vec!["subprocess.Popen"],
        "{:?}",
        spawned.out
    );
    assert!(spawned.out.status.success());

    let doc_path = format!("{root}/{}", doc("a.md"));
    let twice = run_script(
        "twice.py",
        &format!("open('{doc_path}').read()\nopen('{doc_path}').read()\n"),
    );
    assert_eq!(twice.opened, vec![doc_path.clone(), doc_path.clone()]);
    assert_eq!(
        repeated_opens(&twice.opened, 3),
        vec![(doc_path.clone(), 2)],
        "a recorded document read twice is over the bound however many targets there are"
    );
    assert!(twice.spawned.is_empty());

    let stray_path = format!("{root}/secret.txt");
    let stray = run_script("stray.py", &format!("open('{stray_path}').read()\n"));
    assert_eq!(outside_opens(&stray.opened, &root), vec![&stray_path]);

    // A folder next to the root whose name starts with the root's name is not under the root.
    let sibling = format!("{root}-sibling");
    fs::create_dir_all(&sibling).unwrap();
    let sibling_file = format!("{sibling}/note.txt");
    fs::write(&sibling_file, "x\n").unwrap();
    let beside = run_script("beside.py", &format!("open('{sibling_file}').read()\n"));
    fs::remove_dir_all(&sibling).unwrap();
    assert!(
        beside.opened.is_empty(),
        "a sibling folder with the root's name as a prefix is not counted: {:?}",
        beside.opened
    );

    let silent = run_script(
        "silent.py",
        "print('checked 3 recorded files, 1 --test targets listed')\n",
    );
    let (files, targets) = checked_counts(&text_of(&silent.out));
    assert_eq!((files, targets), (3, 1));
    assert!(
        silent.opened.len() < files + targets,
        "a run that opened nothing must fall below the count the bound requires"
    );

    // The shared modules of the test sources are read once per target, so the same file seen
    // three times is within the bound for three targets and over it for two.
    let shared = format!("{MANIFEST}/tests/common/mod.rs");
    let thrice = vec![shared.clone(), shared.clone(), shared.clone()];
    assert!(repeated_opens(&thrice, 3).is_empty());
    assert_eq!(repeated_opens(&thrice, 2), vec![(shared, 3)]);
    // A target root read for every command is over the bound at any target count.
    let root_file = format!("{MANIFEST}/tests/product_cli.rs");
    assert_eq!(
        repeated_opens(&[root_file.clone(), root_file.clone()], 6),
        vec![(root_file, 2)]
    );
}

/// `docs/<name>` inside a temporary tree. Built at run time so tests/ci_scope_test.py, which greps
/// quoted `docs/…` strings as repository documents, does not read these temporary paths as such.
#[cfg(unix)]
fn doc(name: &str) -> String {
    format!("{}/{name}", "docs")
}

/// Writes `files` (relative path, text) under a new temporary tree.
#[cfg(unix)]
fn tree(files: &[(&str, &str)]) -> tempfile::TempDir {
    let dir = tempfile::TempDir::new().unwrap();
    for (rel, text) in files {
        let path = dir.path().join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, text).unwrap();
    }
    dir
}

#[cfg(unix)]
const SHAPES: &str = r##"// a brace in a comment: {
/* a block comment with a brace } and a /* nested */ one */
const TEXT: &str = "not a module { mod fake { #[test] fn in_a_string() {} } }";
const RAW: &str = r#"#[test] fn in_a_raw_string() {}"#;
const OPEN: char = '{';
fn borrow<'a>(x: &'a str) -> &'a str { x }

#[test]
fn top_level() {}

mod inner {
    #[test]
    // a comment between the attributes
    #[ignore = "kept for the list"]
    fn nested_one() {
        let _ = "}";
    }

    mod deeper {
        #[tokio::test]
        async fn deepest() {}
    }
}

mod helpers;

#[path = "elsewhere/odd_name.rs"]
mod renamed;

#[test]
fn after_the_modules() {}
"##;

#[cfg(unix)]
#[test]
fn test_names_are_read_from_the_sources_with_their_module_paths() {
    let dir = tree(&[
        ("tests/shapes.rs", SHAPES),
        ("tests/helpers.rs", "#[test]\nfn from_a_module_file() {}\n"),
        (
            "tests/elsewhere/odd_name.rs",
            "#[test]\nfn from_a_path_attribute() {}\n",
        ),
        (
            doc("good.md").as_str(),
            "`cargo test --test shapes top_level`\n\
             `cargo test --test shapes inner::nested_one`\n\
             `cargo test --test shapes inner::deeper::deepest`\n\
             `cargo test --test shapes helpers::from_a_module_file`\n\
             `cargo test --test shapes renamed::from_a_path_attribute`\n\
             `cargo test --test shapes after_the_modules`\n",
        ),
        (
            doc("bad.md").as_str(),
            "`cargo test --test shapes in_a_string`\n\
             `cargo test --test shapes in_a_raw_string`\n\
             `cargo test --test shapes inner::after_the_modules`\n\
             `cargo test --test shapes deeper::nested_one`\n",
        ),
    ]);
    let root = dir.path().to_str().unwrap();
    let (out, cargo_started) = run_lint_without_cargo(&["--tree", root]);
    let text = text_of(&out);
    assert!(!cargo_started, "the lint must not start cargo:\n{text}");
    let violations: Vec<&str> = text
        .lines()
        .filter(|l| l.starts_with("VIOLATION:"))
        .collect();
    assert!(
        violations.iter().all(|l| l.contains(&doc("bad.md"))),
        "every good filter must select a read test:\n{text}"
    );
    for filter in [
        "in_a_string",
        "in_a_raw_string",
        "inner::after_the_modules",
        "deeper::nested_one",
    ] {
        assert!(
            violations
                .iter()
                .any(|l| l.contains(&format!("'{filter}'"))),
            "the filter {filter} names no test and must be a violation:\n{text}"
        );
    }
    assert_eq!(violations.len(), 4, "{text}");
    assert!(!out.status.success(), "{text}");
}

#[cfg(unix)]
#[test]
fn a_filter_on_macro_made_tests_says_the_names_cannot_be_read() {
    let dir = tree(&[
        (
            "tests/made.rs",
            "macro_rules! case {\n    ($name:ident) => {\n        #[test]\n        fn $name() {}\n    };\n}\n\
             case!(made_by_macro);\n",
        ),
        (doc("record.md").as_str(), "`cargo test --test made made_by_macro`\n"),
    ]);
    let root = dir.path().to_str().unwrap();
    let (out, cargo_started) = run_lint_without_cargo(&["--tree", root]);
    let text = text_of(&out);
    assert!(!cargo_started, "the lint must not start cargo:\n{text}");
    assert!(!out.status.success(), "{text}");
    assert!(
        text.lines().any(|l| l.starts_with("VIOLATION:")
            && l.contains(&doc("record.md"))
            && l.contains("macro")),
        "the violation must say the names are made by a macro:\n{text}"
    );
}

#[test]
fn recorded_ac_commands_name_existing_targets_and_select_tests() {
    let out = run_lint(&["--tree", MANIFEST]);
    let stdout = String::from_utf8_lossy(&out.stdout).to_string();
    let stderr = String::from_utf8_lossy(&out.stderr).to_string();
    assert!(
        out.status.success(),
        "recorded `cargo test --test …` commands must name existing targets and \
         select at least one test:\n{stdout}{stderr}"
    );
}

#[test]
fn lint_flags_a_zero_selection_command() {
    let dir = tempfile::TempDir::new().unwrap();
    let report = dir.path().join("zero-select-report.md");
    fs::write(
        &report,
        "verification: `cargo test --offline --locked --test product_cli \
         definitely_not_a_selected_test -- --nocapture`\n",
    )
    .unwrap();
    let file = report.to_str().unwrap();
    let out = run_lint(&["--file", file]);
    let combined = format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    assert!(
        !out.status.success(),
        "a `--test` command whose filters select nothing must be flagged:\n{combined}"
    );
    assert!(
        combined.contains(file),
        "violation must name the file: {combined}"
    );
}
