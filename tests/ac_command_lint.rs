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

/// The paths under `tests/` this file has, before and after its area (`spec`, first word `ac`
/// in `tests/areas.toml`) is moved into a crate.
const SELF_PATHS: [&str; 2] = ["ac_command_lint.rs", "spec/ac_command_lint.rs"];

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
            .replace('\\', "/");
        // This file is exempt at its two exact places: today at the top level and, once its area
        // is moved, as a module of the `spec` crate. A same-named file anywhere else is not.
        if rel.starts_with("common/") || SELF_PATHS.contains(&rel.as_str()) {
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

/// After an area move this file lives at `tests/spec/ac_command_lint.rs` and keeps its
/// exemption; any other file with an unguarded git spawn is still reported, a file of the same
/// name in another folder included.
#[test]
fn area_module_of_this_file_keeps_the_git_scan_exemption() {
    let dir = tempfile::TempDir::new().unwrap();
    let spawn = "fn f() { std::process::Command::new(\"git\"); }\n";
    for rel in [
        "ac_command_lint.rs",
        "spec/ac_command_lint.rs",
        "spec/other_cli.rs",
        "repo/ac_command_lint.rs",
        "spec/deep/ac_command_lint.rs",
    ] {
        let file = dir.path().join(rel);
        fs::create_dir_all(file.parent().unwrap()).unwrap();
        fs::write(file, spawn).unwrap();
    }
    assert_eq!(
        files_with_unguarded_git(dir.path()),
        vec![
            "tests/repo/ac_command_lint.rs".to_string(),
            "tests/spec/deep/ac_command_lint.rs".to_string(),
            "tests/spec/other_cli.rs".to_string(),
        ]
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

fn read_text(path: &Path) -> std::io::Result<String> {
    fs::read(path).map(|bytes| String::from_utf8_lossy(&bytes).into_owned())
}

/// The lint's verdict and its text (summary line, then one `VIOLATION:` line each).
struct Run {
    ok: bool,
    text: String,
}

/// Runs the recorded-AC-command lint as `--tree <root>` or `--file <file>` (the file's targets
/// are resolved under the repository root).
fn run_lint(args: &[&str]) -> Run {
    let mut read = read_text;
    let report = match args {
        ["--tree", root] => lint::lint_tree(Path::new(root), &mut read),
        ["--file", file] => lint::lint_file(Path::new(MANIFEST), Path::new(file), &mut read),
        _ => panic!("usage: --tree <root> | --file <file>, got {args:?}"),
    };
    Run {
        ok: report.ok(),
        text: report.text(),
    }
}

/// One lint run over `root`, with the files it opened in order.
fn lint_counted(root: &str) -> (lint::Report, Vec<String>) {
    let mut opened = Vec::new();
    let report = {
        let mut read = |path: &Path| {
            opened.push(path.to_string_lossy().into_owned());
            read_text(path)
        };
        lint::lint_tree(Path::new(root), &mut read)
    };
    (report, opened)
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

/// Files opened more often than the lint's design allows, as `(path, opens)`. The lint reads each
/// recorded document once and each target root `tests/<name>.rs` once (one reading per target is
/// cached); a module file in a subfolder of `tests/` (`tests/common/*.rs`) is read once for every
/// target that declares it, so at most `targets` times.
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
fn outside_opens<'a>(opened: &'a [String], root: &str) -> Vec<&'a String> {
    let reports = format!("{}/reports", moai_cli::PROJECT_STATE_DIR);
    let allowed = ["docs", reports.as_str(), "tests"].map(|d| format!("{root}/{d}/"));
    opened
        .iter()
        .filter(|p| !allowed.iter().any(|a| p.starts_with(a.as_str())))
        .collect()
}

/// The lint does a bounded amount of work, counted and not timed: it opens each recorded file and
/// each target root once and a shared test module once per target (see `repeated_opens`), and it
/// opens nothing outside the folders it scans. A clock limit failed at 5.18 s under machine load
/// 200 and could not tell a slow machine from a slow lint; a count does not move with load.
#[test]
fn the_repository_lint_reads_each_file_once_and_only_the_scanned_folders() {
    let (report, opened) = lint_counted(MANIFEST);
    assert!(
        report.ok(),
        "the recorded commands must pass without cargo:\n{}",
        report.text()
    );
    let repeated = repeated_opens(&opened, report.targets);
    assert!(
        repeated.is_empty(),
        "the lint opened files more often than its design allows: {repeated:?}"
    );
    let outside = outside_opens(&opened, MANIFEST);
    assert!(
        outside.is_empty(),
        "the lint opened files outside the three folders it scans: {outside:?}"
    );
    // A counter that saw nothing proves nothing: the lint read every recorded file it reports and
    // at least one test source for every target it listed.
    assert!(
        opened.len() >= report.files + report.targets,
        "the counter saw {} opens for {} recorded files and {} targets:\n{}",
        opened.len(),
        report.files,
        report.targets,
        report.text()
    );
}

/// The lint module starts no process of any kind (the 68-112 s cargo build of the earlier lint
/// was one) and opens a file only through the reader it is given, which is what the counted run
/// above relies on. Read from this file's text between its two marker comments.
#[test]
fn the_lint_module_starts_no_process_and_opens_files_only_through_its_reader() {
    let text = include_str!("ac_command_lint.rs");
    let begin = text
        .find(concat!("// lint", ":begin"))
        .expect("the begin marker of the lint module");
    let end = text
        .find(concat!("// lint", ":end"))
        .expect("the end marker of the lint module");
    assert!(begin < end, "the markers are out of order");
    let span = &text[begin..end];
    for needle in [
        concat!("std::", "process"),
        concat!("Comm", "and"),
        concat!("File", "::open"),
        concat!("fs::", "read("),
        concat!("fs::", "read_to_string"),
        concat!("fs::", "write"),
    ] {
        assert!(
            !span.contains(needle),
            "the lint module must not use `{needle}`: it starts no process and reads only \
             through its reader"
        );
    }
}

/// The counter itself is checked on a small tree, so the bound above cannot go green because the
/// counter stopped counting: one run lists exactly the recorded file and the target root it read
/// (a command named twice still reads the target once), and the two bounds report a second read,
/// a stray read and a shared module read too often.
#[test]
fn the_read_counter_sees_each_read_and_the_bounds_catch_a_second_read_or_a_stray_one() {
    let dir = tree(&[
        ("tests/t.rs", "#[test]\nfn f() {}\n"),
        (
            doc("a.md").as_str(),
            "`cargo test --test t f`\n`cargo test --test t f`\n",
        ),
    ]);
    let root = dir.path().to_str().unwrap().to_string();
    let (report, opened) = lint_counted(&root);
    assert!(report.ok(), "{}", report.text());
    assert_eq!((report.files, report.targets), (1, 1));
    let doc_path = format!("{root}/{}", doc("a.md"));
    assert_eq!(opened, vec![doc_path.clone(), format!("{root}/tests/t.rs")]);
    assert!(repeated_opens(&opened, report.targets).is_empty());
    assert!(outside_opens(&opened, &root).is_empty());

    assert_eq!(
        repeated_opens(&[doc_path.clone(), doc_path.clone()], 3),
        vec![(doc_path.clone(), 2)],
        "a recorded document read twice is over the bound however many targets there are"
    );
    let stray_path = format!("{root}/secret.txt");
    let stray = vec![stray_path.clone()];
    assert_eq!(outside_opens(&stray, &root), vec![&stray_path]);
    // A folder next to the root whose name starts with the root's name is not one of the three.
    let beside = vec![format!("{root}-sibling/docs/note.md")];
    assert_eq!(outside_opens(&beside, &root).len(), 1);

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
fn doc(name: &str) -> String {
    format!("{}/{name}", "docs")
}

/// Writes `files` (relative path, text) under a new temporary tree.
fn tree(files: &[(&str, &str)]) -> tempfile::TempDir {
    let dir = tempfile::TempDir::new().unwrap();
    for (rel, text) in files {
        let path = dir.path().join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, text).unwrap();
    }
    dir
}

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
    let run = run_lint(&["--tree", root]);
    let text = &run.text;
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
    assert!(!run.ok, "{text}");
}

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
    let run = run_lint(&["--tree", root]);
    assert!(!run.ok, "{}", run.text);
    assert!(
        run.text.lines().any(|l| l.starts_with("VIOLATION:")
            && l.contains(&doc("record.md"))
            && l.contains("macro")),
        "the violation must say the names are made by a macro:\n{}",
        run.text
    );
}

/// A repository tree with one moved area `repo`: the crate root declares the module `git_a`,
/// which holds the test `finds_it`; `command` is the recorded line of `docs/record.md`. Returns
/// the lint's result on that tree (SPEC-20261123-test-area-crates REQ-6, REQ-12).
fn lint_area_tree(command: &str) -> Run {
    let dir = tempfile::TempDir::new().unwrap();
    let files = [
        ("tests/repo/main.rs", "mod git_a;\n"),
        ("tests/repo/git_a.rs", "#[test]\nfn finds_it() {}\n"),
        ("tests/plain.rs", "#[test]\nfn plain_test() {}\n"),
    ];
    for (rel, text) in files {
        let path = dir.path().join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, text).unwrap();
    }
    fs::create_dir_all(dir.path().join("docs")).unwrap();
    fs::write(
        dir.path().join("docs/record.md"),
        format!("verification: `{command}`\n"),
    )
    .unwrap();
    run_lint(&["--tree", dir.path().to_str().unwrap()])
}

/// A closed SPEC keeps `--test <old target>`: the old name is a module of an area crate and the
/// recorded filter still selects its test.
#[test]
fn area_old_target_name_resolves_to_its_module() {
    let run = lint_area_tree("cargo test --offline --locked --test git_a finds_it");
    assert!(run.ok, "{}", run.text);
}

#[test]
fn area_crate_name_with_a_module_filter_passes() {
    let run = lint_area_tree("cargo test --offline --locked --test repo git_a::finds_it");
    assert!(run.ok, "{}", run.text);
}

#[test]
fn area_filter_that_selects_nothing_fails() {
    let run = lint_area_tree("cargo test --offline --locked --test repo git_zzz::");
    assert!(!run.ok, "{}", run.text);
    assert!(run.text.contains("select 0 tests in repo"), "{}", run.text);
}

#[test]
fn area_unknown_target_fails_even_beside_moved_areas() {
    let run = lint_area_tree("cargo test --offline --locked --test never_moved some_test");
    assert!(!run.ok, "{}", run.text);
    assert!(
        run.text
            .contains("target tests/never_moved.rs does not exist"),
        "{}",
        run.text
    );
}

#[test]
fn recorded_ac_commands_name_existing_targets_and_select_tests() {
    let run = run_lint(&["--tree", MANIFEST]);
    assert!(
        run.ok,
        "recorded `cargo test --test …` commands must name existing targets and \
         select at least one test:\n{}",
        run.text
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
    let run = run_lint(&["--file", file]);
    assert!(
        !run.ok,
        "a `--test` command whose filters select nothing must be flagged:\n{}",
        run.text
    );
    assert!(
        run.text.contains(file),
        "violation must name the file: {}",
        run.text
    );
}
// lint:begin
/// The recorded-AC-command lint (#2352), ported from the removed `tools/check_ac_commands.py`.
/// A `cargo test --test X [filters…]` line in the recorded verification commands must name an
/// existing test target and its filters must select at least one test: cargo exits 0 with
/// "running 0 tests" for a filter that matches nothing, so such a command has passed as
/// verification before (survey 2.11).
///
/// The scanned files are `docs/**/*.md` and the recorded reports of the project state folder.
/// A candidate command is a `cargo test` line inside a fenced code block or a backtick code
/// span; env prefixes are stripped. The target must have `tests/<name>.rs` (or be a module of an
/// area crate), and its filters must select at least one test, checked against the test names
/// read from the sources as cargo prints them (`module::name`, one reading per target, cached).
/// Names a `macro_rules!` body makes cannot be read; a filter only such a name would satisfy is
/// reported as such. Placeholder filters (`<name>`) are skipped as templates.
///
/// The module starts no process and opens a file only through the `read` closure it is given,
/// so a test can count the reads (SPEC-20261122-ac-lint-no-cargo).
mod lint {
    use std::collections::{BTreeSet, HashMap, HashSet};
    use std::io;
    use std::path::{Path, PathBuf};

    /// Opens one file as text. The lint reads files only through this.
    pub type Read<'a> = &'a mut dyn FnMut(&Path) -> io::Result<String>;

    pub struct Report {
        pub files: usize,
        pub targets: usize,
        pub violations: Vec<String>,
    }

    impl Report {
        pub fn ok(&self) -> bool {
            self.violations.is_empty()
        }

        /// The summary line, then one `VIOLATION:` line each.
        pub fn text(&self) -> String {
            let mut out = format!(
                "checked {} recorded files, {} --test targets listed\n",
                self.files, self.targets
            );
            for v in &self.violations {
                out.push_str("VIOLATION: ");
                out.push_str(v);
                out.push('\n');
            }
            out
        }
    }

    fn is_word(c: char) -> bool {
        c.is_alphanumeric() || c == '_'
    }

    fn is_ident_start(c: char) -> bool {
        c.is_ascii_alphabetic() || c == '_'
    }

    fn is_ident(c: char) -> bool {
        c.is_ascii_alphanumeric() || c == '_'
    }

    fn starts(s: &[char], i: usize, pat: &str) -> bool {
        let mut k = i;
        for p in pat.chars() {
            if s.get(k) != Some(&p) {
                return false;
            }
            k += 1;
        }
        true
    }

    fn skip_ws(s: &[char], mut i: usize) -> usize {
        while s.get(i).is_some_and(|c| c.is_whitespace()) {
            i += 1;
        }
        i
    }

    /// The end of an identifier that starts at `i`, or `None`.
    fn ident_end(s: &[char], i: usize) -> Option<usize> {
        if !s.get(i).is_some_and(|&c| is_ident_start(c)) {
            return None;
        }
        let mut e = i + 1;
        while s.get(e).is_some_and(|&c| is_ident(c)) {
            e += 1;
        }
        Some(e)
    }

    fn find_seq(s: &[char], needle: &[char], from: usize) -> Option<usize> {
        if needle.is_empty() || s.len() < needle.len() {
            return None;
        }
        (from..=s.len() - needle.len()).find(|&p| s[p..p + needle.len()] == *needle)
    }

    fn slice(s: &[char], a: usize, b: usize) -> String {
        let b = b.min(s.len());
        if a >= b {
            String::new()
        } else {
            s[a..b].iter().collect()
        }
    }

    // ---- recorded files and their commands -------------------------------------------------

    fn markdown_under(dir: &Path, out: &mut Vec<PathBuf>) {
        let Ok(entries) = std::fs::read_dir(dir) else {
            return;
        };
        for entry in entries.flatten() {
            let path = entry.path();
            if path.is_dir() {
                markdown_under(&path, out);
            } else if path.extension().is_some_and(|e| e == "md") {
                out.push(path);
            }
        }
    }

    fn markdown_in(dir: &Path) -> Vec<PathBuf> {
        let mut out: Vec<PathBuf> = std::fs::read_dir(dir)
            .into_iter()
            .flatten()
            .flatten()
            .map(|e| e.path())
            .filter(|p| p.extension().is_some_and(|e| e == "md"))
            .collect();
        out.sort();
        out
    }

    /// The recorded files under `root`. Frozen RC1 records stay out: the closed-issue mirrors
    /// (numbers below 2153, the rc.2 split) and the rc1-issue-design HTML export; a historical
    /// record is not edited to match a renamed test. Live records (bodies 2153 and up, root
    /// reports, docs) stay guarded.
    pub fn scan_files(root: &Path) -> Vec<PathBuf> {
        let mut docs = Vec::new();
        markdown_under(&root.join("docs"), &mut docs);
        docs.sort();
        let reports = root.join(moai_cli::PROJECT_STATE_DIR).join("reports");
        let mut files = docs;
        files.extend(markdown_in(&reports));
        for path in markdown_in(&reports.join("issue-bodies")) {
            let name = path
                .file_name()
                .map(|n| n.to_string_lossy().into_owned())
                .unwrap_or_default();
            let digits: String = name.chars().take_while(|c| c.is_ascii_digit()).collect();
            let after = name[digits.len()..].starts_with('-');
            if after && digits.parse::<u64>().is_ok_and(|n| n >= 2153) {
                files.push(path);
            }
        }
        files
    }

    /// `cargo test` command strings: fenced-block lines and backtick code spans.
    fn candidates(text: &str) -> Vec<String> {
        let mut out = Vec::new();
        let mut in_fence = false;
        for line in text.lines() {
            if line.trim().starts_with("```") {
                in_fence = !in_fence;
                continue;
            }
            if in_fence {
                out.push(line.to_string());
                continue;
            }
            let chars: Vec<char> = line.chars().collect();
            let mut p = 0;
            while let Some(open) = chars[p..].iter().position(|&c| c == '`').map(|q| p + q) {
                let Some(close) = chars[open + 1..]
                    .iter()
                    .position(|&c| c == '`')
                    .map(|q| open + 1 + q)
                else {
                    break;
                };
                if close == open + 1 {
                    p = close;
                } else {
                    out.push(slice(&chars, open + 1, close));
                    p = close + 1;
                }
            }
        }
        out
    }

    /// The line without leading `NAME=value ` env prefixes.
    fn strip_env_prefix(line: &str) -> &str {
        let mut rest = line;
        loop {
            let chars: Vec<char> = rest.chars().collect();
            let Some(&first) = chars.first() else {
                return rest;
            };
            if !(first.is_ascii_uppercase() || first == '_') {
                return rest;
            }
            let mut k = 1;
            while chars
                .get(k)
                .is_some_and(|c| c.is_ascii_uppercase() || c.is_ascii_digit() || *c == '_')
            {
                k += 1;
            }
            if chars.get(k) != Some(&'=') {
                return rest;
            }
            let value_start = k + 1;
            let mut v = value_start;
            while chars.get(v).is_some_and(|c| !c.is_whitespace()) {
                v += 1;
            }
            let after = skip_ws(&chars, v);
            if v == value_start || after == v {
                return rest;
            }
            let byte: usize = chars[..after].iter().map(|c| c.len_utf8()).sum();
            rest = &rest[byte..];
        }
    }

    /// Shell-like tokenization good enough for a recorded cargo command.
    fn tokens_of(candidate: &str) -> Vec<String> {
        let mut line = candidate.trim();
        loop {
            let stripped = strip_env_prefix(line);
            if stripped == line {
                break;
            }
            line = stripped;
        }
        let line = if line.contains('`') {
            line.trim_matches('`')
        } else {
            line
        };
        line.split_whitespace().map(str::to_string).collect()
    }

    /// Positional filter words after `--test <name>` up to the next flag token.
    fn filters_after_target(tokens: &[String]) -> Vec<String> {
        let Some(idx) = tokens.iter().position(|t| t == "--test") else {
            return Vec::new();
        };
        tokens
            .iter()
            .skip(idx + 2)
            .take_while(|t| !t.starts_with('-'))
            .cloned()
            .collect()
    }

    /// Whether `cargo\s+test\b` occurs in `text`.
    fn mentions_cargo_test(text: &str) -> bool {
        let chars: Vec<char> = text.chars().collect();
        let mut from = 0;
        while let Some(p) = find_seq(&chars, &['c', 'a', 'r', 'g', 'o'], from) {
            let k = skip_ws(&chars, p + 5);
            if k > p + 5
                && starts(&chars, k, "test")
                && !chars.get(k + 4).is_some_and(|&c| is_word(c))
            {
                return true;
            }
            from = p + 1;
        }
        false
    }

    /// The name after the first `--test\s+` in `text`.
    fn test_target_name(text: &str) -> Option<String> {
        let chars: Vec<char> = text.chars().collect();
        let mut from = 0;
        while let Some(p) = find_seq(&chars, &['-', '-', 't', 'e', 's', 't'], from) {
            let k = skip_ws(&chars, p + 6);
            if k > p + 6 {
                let mut e = k;
                while chars
                    .get(e)
                    .is_some_and(|c| c.is_ascii_alphanumeric() || *c == '_' || *c == '-')
                {
                    e += 1;
                }
                if e > k {
                    return Some(slice(&chars, k, e));
                }
            }
            from = p + 1;
        }
        None
    }

    /// The text of an HTML export with its tags blanked.
    fn strip_tags(text: &str) -> String {
        let chars: Vec<char> = text.chars().collect();
        let mut out = String::new();
        let mut i = 0;
        while i < chars.len() {
            if chars[i] == '<' {
                let mut j = i + 1;
                while j < chars.len() && chars[j] != '>' {
                    j += 1;
                }
                if j < chars.len() && j > i + 1 {
                    out.push(' ');
                    i = j + 1;
                    continue;
                }
            }
            out.push(chars[i]);
            i += 1;
        }
        out
    }

    // ---- test targets and the test names read from their sources ---------------------------

    /// (file to read, folder its `mod x;` lines resolve in) of the test target `name`, or `None`.
    ///
    /// A target is a top-level `tests/<name>.rs`, an area crate `tests/<name>/main.rs`, or the
    /// old name of a file moved into an area (`tests/<area>/<name>.rs`, a module of that crate):
    /// closed SPECs keep `--test <old name>` and it still resolves (SPEC-20261123-test-area-crates
    /// REQ-6). Found by where the files are, so no table is read. Anything else is `None`.
    fn resolve_target(root: &Path, name: &str) -> Option<(PathBuf, PathBuf)> {
        let tests_dir = root.join("tests");
        let top = tests_dir.join(format!("{name}.rs"));
        if top.is_file() {
            return Some((top, tests_dir));
        }
        let crate_main = tests_dir.join(name).join("main.rs");
        if crate_main.is_file() {
            return Some((crate_main, tests_dir.join(name)));
        }
        if name != "main" && tests_dir.is_dir() {
            let mut areas: Vec<PathBuf> = std::fs::read_dir(&tests_dir)
                .into_iter()
                .flatten()
                .flatten()
                .map(|e| e.path())
                .filter(|p| p.is_dir())
                .collect();
            areas.sort();
            for area in areas {
                let module = area.join(format!("{name}.rs"));
                if area.join("main.rs").is_file() && module.is_file() {
                    return Some((module, area.join(name)));
                }
            }
        }
        None
    }

    /// The source with comments, string, raw-string and char literals blanked (string contents
    /// kept only inside attributes, where `#[path = "…"]` needs them). Lengths are not kept;
    /// only the token order matters to the reader.
    fn code_only(src: &str) -> Vec<char> {
        let s: Vec<char> = src.chars().collect();
        let n = s.len();
        let mut out: Vec<char> = Vec::with_capacity(n);
        let mut i = 0;
        let mut attr_depth = 0usize;
        while i < n {
            let c = s[i];
            if starts(&s, i, "//") {
                i = s[i..].iter().position(|&x| x == '\n').map_or(n, |p| i + p);
                continue;
            }
            if starts(&s, i, "/*") {
                let (mut depth, mut k) = (1usize, i + 2);
                while k < n && depth > 0 {
                    if starts(&s, k, "/*") {
                        depth += 1;
                        k += 2;
                    } else if starts(&s, k, "*/") {
                        depth -= 1;
                        k += 2;
                    } else {
                        k += 1;
                    }
                }
                out.push(' ');
                i = k;
                continue;
            }
            if let Some((hashes, open_len)) = raw_string_open(&s, i) {
                if i == 0 || !is_word(s[i - 1]) {
                    let mut close = vec!['"'];
                    close.extend(std::iter::repeat('#').take(hashes));
                    let end = find_seq(&s, &close, i + open_len).map_or(n, |p| p + close.len());
                    if attr_depth > 0 {
                        out.extend_from_slice(&s[i..end]);
                    } else {
                        out.extend(['"', '"']);
                    }
                    i = end;
                    continue;
                }
            }
            if c == '"' {
                let mut j = i + 1;
                while j < n && s[j] != '"' {
                    j += if s[j] == '\\' { 2 } else { 1 };
                }
                if attr_depth > 0 {
                    out.extend_from_slice(&s[i..(j + 1).min(n)]);
                } else {
                    out.extend(['"', '"']);
                }
                i = j + 1;
                continue;
            }
            if c == '\'' {
                if let Some(len) = char_literal_len(&s, i) {
                    out.extend(['\'', ' ', '\'']);
                    i += len;
                    continue;
                }
            }
            if starts(&s, i, "#[") || starts(&s, i, "#![") {
                attr_depth += 1;
            } else if c == ']' && attr_depth > 0 {
                attr_depth -= 1;
            }
            out.push(c);
            i += 1;
        }
        out
    }

    /// `b?r#*"` at `i`: (number of `#`, length of the opening).
    fn raw_string_open(s: &[char], i: usize) -> Option<(usize, usize)> {
        let mut k = i;
        if s.get(k) == Some(&'b') {
            k += 1;
        }
        if s.get(k) != Some(&'r') {
            return None;
        }
        k += 1;
        let mut hashes = 0;
        while s.get(k) == Some(&'#') {
            hashes += 1;
            k += 1;
        }
        if s.get(k) == Some(&'"') {
            Some((hashes, k + 1 - i))
        } else {
            None
        }
    }

    /// The length of a char literal at `i` (`'a'`, `'\n'`, `'\x41'`, `'\u{1F600}'`), or `None`
    /// for a lifetime.
    fn char_literal_len(s: &[char], i: usize) -> Option<usize> {
        let k = i + 1;
        let c = *s.get(k)?;
        if c == '\\' {
            let e = k + 1;
            let q = *s.get(e)?;
            if q == 'x'
                && s.get(e + 1).is_some_and(|c| c.is_ascii_hexdigit())
                && s.get(e + 2).is_some_and(|c| c.is_ascii_hexdigit())
                && s.get(e + 3) == Some(&'\'')
            {
                return Some(e + 4 - i);
            }
            if q == 'u' && s.get(e + 1) == Some(&'{') {
                let mut h = e + 2;
                while s.get(h).is_some_and(|c| c.is_ascii_hexdigit()) {
                    h += 1;
                }
                let digits = h - (e + 2);
                if (1..=6).contains(&digits)
                    && s.get(h) == Some(&'}')
                    && s.get(h + 1) == Some(&'\'')
                {
                    return Some(h + 2 - i);
                }
            }
            if q != '\n' && s.get(e + 1) == Some(&'\'') {
                return Some(e + 2 - i);
            }
            return None;
        }
        if c != '\'' && c != '\n' && s.get(k + 1) == Some(&'\'') {
            return Some(3);
        }
        None
    }

    /// The file of `mod name;` declared in a file whose child modules live under `base`.
    fn module_file(base: &Path, name: &str) -> Option<PathBuf> {
        [
            base.join(format!("{name}.rs")),
            base.join(name).join("mod.rs"),
        ]
        .into_iter()
        .find(|c| c.is_file())
    }

    /// The end of an optional `pub` / `pub(…)` prefix and its whitespace at `i` (`i` itself
    /// when there is none); `None` when `pub` is there without the whitespace after it.
    fn after_pub(code: &[char], i: usize) -> Option<usize> {
        if !starts(code, i, "pub") {
            return Some(i);
        }
        let mut k = i + 3;
        if code.get(k) == Some(&'(') {
            let close = code[k + 1..].iter().position(|&c| c == ')')?;
            k = k + 1 + close + 1;
        }
        let w = skip_ws(code, k);
        (w > k).then_some(w)
    }

    /// `[pub …] mod NAME {` or `;` at `i`: (name, the `{` or `;`, end).
    fn match_mod(code: &[char], i: usize) -> Option<(String, char, usize)> {
        let mut k = after_pub(code, i)?;
        if !starts(code, k, "mod") {
            return None;
        }
        k += 3;
        let w = skip_ws(code, k);
        if w == k {
            return None;
        }
        let e = ident_end(code, w)?;
        let at = skip_ws(code, e);
        match code.get(at) {
            Some(&c) if c == '{' || c == ';' => Some((slice(code, w, e), c, at + 1)),
            _ => None,
        }
    }

    /// `[pub …] [const] [async] [unsafe] fn NAME` at `i`: (name, end).
    fn match_fn(code: &[char], i: usize) -> Option<(String, usize)> {
        let mut k = after_pub(code, i)?;
        for kw in ["const", "async", "unsafe"] {
            if starts(code, k, kw) {
                let w = skip_ws(code, k + kw.len());
                if w > k + kw.len() {
                    k = w;
                }
            }
        }
        if !starts(code, k, "fn") {
            return None;
        }
        k += 2;
        let w = skip_ws(code, k);
        if w == k {
            return None;
        }
        let e = ident_end(code, w)?;
        Some((slice(code, w, e), e))
    }

    /// `IDENT (\s* :: \s* IDENT)*` at the start of `body`: its end.
    fn path_head_end(body: &[char]) -> Option<usize> {
        let mut e = ident_end(body, 0)?;
        loop {
            let k = skip_ws(body, e);
            if starts(body, k, "::") {
                if let Some(e2) = ident_end(body, skip_ws(body, k + 2)) {
                    e = e2;
                    continue;
                }
            }
            return Some(e);
        }
    }

    /// `path\s*=\s*"([^"]+)"` at the start of `body`.
    fn path_attr(body: &[char]) -> Option<String> {
        if !starts(body, 0, "path") {
            return None;
        }
        let k = skip_ws(body, 4);
        if body.get(k) != Some(&'=') {
            return None;
        }
        let q = skip_ws(body, k + 1);
        if body.get(q) != Some(&'"') {
            return None;
        }
        let close = body[q + 1..].iter().position(|&c| c == '"')?;
        (close > 0).then(|| slice(body, q + 1, q + 1 + close))
    }

    /// Whether a `macro_rules!` body in `code` holds a test attribute.
    fn macro_makes_tests(code: &[char]) -> bool {
        let mut from = 0;
        let marker: Vec<char> = "macro_rules!".chars().collect();
        while let Some(p) = find_seq(code, &marker, from) {
            from = p + 1;
            let name = skip_ws(code, p + marker.len());
            let Some(e) = ident_end(code, name) else {
                continue;
            };
            let open = skip_ws(code, e);
            if code.get(open) != Some(&'{') {
                continue;
            }
            let begin = open + 1;
            let (mut depth, mut j) = (1i32, begin);
            while j < code.len() && depth != 0 {
                match code[j] {
                    '{' => depth += 1,
                    '}' => depth -= 1,
                    _ => {}
                }
                j += 1;
            }
            if has_test_attribute(&code[begin..j]) {
                return true;
            }
            from = begin;
        }
        false
    }

    /// `#[ (IDENT::)* test ]` or `#[ (IDENT::)* test(` anywhere in `body`.
    fn has_test_attribute(body: &[char]) -> bool {
        let mut from = 0;
        while let Some(p) = find_seq(body, &['#', '['], from) {
            from = p + 1;
            let mut k = skip_ws(body, p + 2);
            loop {
                let Some(e) = ident_end(body, k) else {
                    break;
                };
                let after = skip_ws(body, e);
                if starts(body, after, "::") {
                    k = skip_ws(body, after + 2);
                    continue;
                }
                if slice(body, k, e) == "test" && matches!(body.get(after), Some(']') | Some('(')) {
                    return true;
                }
                break;
            }
        }
        false
    }

    /// Adds `module::name` for every test function in `file` (and the module files it declares)
    /// to `out`. Returns whether a `macro_rules!` body there holds a test attribute.
    fn read_tests(
        file: &Path,
        base: &Path,
        prefix: &str,
        out: &mut BTreeSet<String>,
        seen: &mut HashSet<PathBuf>,
        read: Read,
    ) -> bool {
        if seen.contains(file) || !file.is_file() {
            return false;
        }
        seen.insert(file.to_path_buf());
        let code = code_only(&read(file).unwrap_or_default());
        let n = code.len();
        let mut macro_tests = macro_makes_tests(&code);
        let mut stack: Vec<(String, i64)> = Vec::new();
        let mut depth: i64 = 0;
        let mut pending_test = false;
        let mut pending_path: Option<String> = None;
        let mut i = 0;
        let join = |parts: &[&str]| -> String {
            parts
                .iter()
                .filter(|p| !p.is_empty())
                .copied()
                .collect::<Vec<_>>()
                .join("::")
        };
        while i < n {
            let c = code[i];
            if c == '#' && code.get(i + 1) == Some(&'[') {
                let (mut depth_a, mut j) = (1i32, i + 2);
                while j < n && depth_a != 0 {
                    match code[j] {
                        '[' => depth_a += 1,
                        ']' => depth_a -= 1,
                        _ => {}
                    }
                    j += 1;
                }
                let body: Vec<char> = slice(&code, i + 2, j.saturating_sub(1))
                    .trim()
                    .chars()
                    .collect();
                if let Some(e) = path_head_end(&body) {
                    let head = slice(&body, 0, e);
                    let rest: String = slice(&body, e, body.len()).trim().to_string();
                    let last = head.rsplit("::").next().unwrap_or("").trim();
                    if last == "test" && (rest.is_empty() || rest.starts_with('(')) {
                        pending_test = true;
                    }
                }
                if let Some(p) = path_attr(&body) {
                    pending_path = Some(p);
                }
                i = j;
                continue;
            }
            if c == '{' {
                depth += 1;
                i += 1;
                continue;
            }
            if c == '}' {
                depth -= 1;
                if stack.last().is_some_and(|s| s.1 == depth) {
                    stack.pop();
                }
                i += 1;
                continue;
            }
            if c.is_alphabetic() || c == '_' {
                let at_word_start = i == 0 || !is_word(code[i - 1]);
                if let Some((name, kind, end)) = match_mod(&code, i).filter(|_| at_word_start) {
                    if kind == '{' {
                        stack.push((name, depth));
                        depth += 1;
                    } else {
                        let inner: Vec<&str> = stack.iter().map(|s| s.0.as_str()).collect();
                        let mut sub_base = base.to_path_buf();
                        for part in &inner {
                            sub_base.push(part);
                        }
                        let target = match &pending_path {
                            Some(p) => {
                                let from = if inner.is_empty() {
                                    file.parent().unwrap_or(base).to_path_buf()
                                } else {
                                    sub_base.clone()
                                };
                                Some(from.join(p))
                            }
                            None => module_file(&sub_base, &name),
                        };
                        if let Some(target) = target.filter(|t| t.is_file()) {
                            let parent = target.parent().unwrap_or(base).to_path_buf();
                            let child_base = if target.file_name().is_some_and(|f| f == "mod.rs")
                                || pending_path.is_some()
                            {
                                parent
                            } else {
                                parent.join(target.file_stem().unwrap_or_default())
                            };
                            let mut parts = vec![prefix];
                            parts.extend(inner.iter().copied());
                            parts.push(name.as_str());
                            let mod_prefix = join(&parts);
                            macro_tests |=
                                read_tests(&target, &child_base, &mod_prefix, out, seen, read);
                        }
                    }
                    pending_test = false;
                    pending_path = None;
                    i = end;
                    continue;
                }
                if let Some((name, end)) = match_fn(&code, i).filter(|_| at_word_start) {
                    if pending_test {
                        let mut parts = vec![prefix];
                        parts.extend(stack.iter().map(|s| s.0.as_str()));
                        parts.push(name.as_str());
                        out.insert(join(&parts));
                    }
                    pending_test = false;
                    pending_path = None;
                    i = end;
                    continue;
                }
                let mut j = i;
                while j < n && is_word(code[j]) {
                    j += 1;
                }
                pending_test = false;
                pending_path = None;
                i = j;
                continue;
            }
            if c == ';' {
                pending_test = false;
                pending_path = None;
            }
            i += 1;
        }
        macro_tests
    }

    type Listed = (BTreeSet<String>, bool);

    /// (test names, whether a `macro_rules!` body makes tests) of `tests/<name>.rs`, read from
    /// the sources: no cargo is started. A moved file's names keep their old form (no module
    /// prefix), so an old recorded filter selects what it always did.
    fn list_tests(
        root: &Path,
        name: &str,
        cache: &mut HashMap<String, Listed>,
        read: Read,
    ) -> Listed {
        if !cache.contains_key(name) {
            let mut tests = BTreeSet::new();
            let macro_tests = match resolve_target(root, name) {
                Some((source, base)) => {
                    read_tests(&source, &base, "", &mut tests, &mut HashSet::new(), read)
                }
                None => false,
            };
            cache.insert(name.to_string(), (tests, macro_tests));
        }
        cache[name].clone()
    }

    /// cargo's filter rule: a test is selected when ANY positional filter is a substring of the
    /// full test path; no filters means "select everything".
    fn selects_any(filters: &[String], tests: &BTreeSet<String>) -> bool {
        filters.is_empty()
            || filters
                .iter()
                .any(|f| tests.iter().any(|t| t.contains(f.as_str())))
    }

    /// A filter list as the earlier tool printed it: `['a', 'b']`.
    fn listed(filters: &[String]) -> String {
        let items: Vec<String> = filters.iter().map(|f| format!("'{f}'")).collect();
        format!("[{}]", items.join(", "))
    }

    fn clip(candidate: &str) -> String {
        candidate.trim().chars().take(170).collect()
    }

    fn violations_in(
        path: &Path,
        text: &str,
        root: &Path,
        cache: &mut HashMap<String, Listed>,
        read: Read,
    ) -> Vec<String> {
        let text = if path.extension().is_some_and(|e| e == "html") {
            strip_tags(text)
        } else {
            text.to_string()
        };
        let mut found = Vec::new();
        let where_ = path.display();
        for candidate in candidates(&text) {
            if !mentions_cargo_test(&candidate) || !candidate.contains("--test") {
                continue;
            }
            let tokens = tokens_of(&candidate);
            let Some(name) = test_target_name(&tokens.join(" ")) else {
                continue;
            };
            if resolve_target(root, &name).is_none() {
                found.push(format!(
                    "{where_}: target tests/{name}.rs does not exist :: {}",
                    clip(&candidate)
                ));
                continue;
            }
            let filters: Vec<String> = filters_after_target(&tokens)
                .into_iter()
                .filter(|f| !f.starts_with('<'))
                .collect();
            let (tests, macro_tests) = list_tests(root, &name, cache, read);
            if macro_tests && !selects_any(&filters, &tests) {
                found.push(format!(
                    "{where_}: filters {} select 0 written tests in {name}; its other tests are \
                     made by a macro_rules! body, whose names cannot be read without cargo: name \
                     a written test :: {}",
                    listed(&filters),
                    clip(&candidate)
                ));
                continue;
            }
            if tests.is_empty() {
                found.push(format!(
                    "{where_}: target {name} lists no tests :: {}",
                    clip(&candidate)
                ));
                continue;
            }
            if !selects_any(&filters, &tests) {
                found.push(format!(
                    "{where_}: filters {} select 0 tests in {name} ({} available) :: {}",
                    listed(&filters),
                    tests.len(),
                    clip(&candidate)
                ));
            }
        }
        found
    }

    fn lint_files(root: &Path, files: Vec<PathBuf>, read: Read) -> Report {
        let mut cache: HashMap<String, Listed> = HashMap::new();
        let mut violations = Vec::new();
        for path in &files {
            match read(path) {
                Ok(text) => violations.extend(violations_in(path, &text, root, &mut cache, read)),
                Err(e) => violations.push(format!("{}: unreadable ({e})", path.display())),
            }
        }
        Report {
            files: files.len(),
            targets: cache.len(),
            violations,
        }
    }

    /// The whole recorded set under `root` (`--tree`).
    pub fn lint_tree(root: &Path, read: Read) -> Report {
        lint_files(root, scan_files(root), read)
    }

    /// One recorded file, its targets resolved under `root` (`--file`).
    pub fn lint_file(root: &Path, file: &Path, read: Read) -> Report {
        lint_files(root, vec![file.to_path_buf()], read)
    }
}
// lint:end
