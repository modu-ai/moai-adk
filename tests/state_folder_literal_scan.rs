//! SPEC-20261122-state-folder-literal-scan: tests do not spell the state folder name beyond a
//! counted baseline.
//!
//! A fresh project or HOME finds its state folder through the resolvers (`state_root`,
//! `home_root`, or the Python helper that mirrors them). A literal folder name is right only
//! where a test builds an earlier layout on purpose (an old-layout fixture, a Go-generation
//! folder, a migration). A text scan cannot tell those lines from a fresh-project expectation,
//! so it ratchets per file: `tests/fixtures/state_folder_literals/allow.tsv` lists every test
//! file allowed to spell either name and how many times. A file outside the list that spells
//! one, or a listed file that spells one more often than its count, fails. Counts may fall
//! without touching the list.
//!
//! `MOAI_STATE_LITERALS_WRITE=1 cargo test --test state_folder_literal_scan` writes the list
//! from the current counts (for a real new fixture, or after test files move).
//!
//! This file spells neither name: both are built at run time, so it needs no row of its own.

mod common;

use common::home::TempHome;
use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};

const MANIFEST: &str = env!("CARGO_MANIFEST_DIR");
const WRITE_SWITCH: &str = "MOAI_STATE_LITERALS_WRITE";

/// The earlier state folder name (the new name plus `-cli`).
fn earlier() -> String {
    format!(".{}-{}", "moai", "cli")
}

/// The new state folder name.
fn new_name() -> String {
    format!(".{}", "moai")
}

/// Per-file counts: (earlier name, new name).
type Counts = BTreeMap<String, (usize, usize)>;

fn is_word(c: char) -> bool {
    c.is_ascii_alphanumeric() || c == '_'
}

/// How often `name` stands in `text` as a whole path segment: not right after a word character
/// (so `mo.ai.kr` or `x.moai` do not count) and not followed by a word character or `-` (so the
/// new name never matches inside the earlier one, nor `.moai-cli-v1` style text).
fn count_segment(text: &str, name: &str) -> usize {
    let mut n = 0;
    let mut from = 0;
    while let Some(at) = text[from..].find(name) {
        let start = from + at;
        let end = start + name.len();
        let before_ok = text[..start]
            .chars()
            .next_back()
            .is_none_or(|c| !is_word(c));
        let after_ok = text[end..]
            .chars()
            .next()
            .is_none_or(|c| !is_word(c) && c != '-');
        if before_ok && after_ok {
            n += 1;
        }
        from = end;
    }
    n
}

fn count_text(text: &str) -> (usize, usize) {
    (
        count_segment(text, &earlier()),
        count_segment(text, &new_name()),
    )
}

/// Every `.rs` and `.py` file under `<root>/tests` outside `tests/fixtures`, by relative path.
fn scan(root: &Path) -> Counts {
    let mut out = Counts::new();
    let tests = root.join("tests");
    let mut stack = vec![tests.clone()];
    while let Some(dir) = stack.pop() {
        let Ok(entries) = fs::read_dir(&dir) else {
            continue;
        };
        for entry in entries.flatten() {
            let path = entry.path();
            let Ok(kind) = entry.file_type() else {
                continue;
            };
            if kind.is_dir() {
                if path != tests.join("fixtures") {
                    stack.push(path);
                }
                continue;
            }
            if !kind.is_file() {
                continue;
            }
            let ext = path.extension().and_then(|e| e.to_str());
            if !matches!(ext, Some("rs") | Some("py")) {
                continue;
            }
            let Ok(text) = fs::read_to_string(&path) else {
                continue;
            };
            let counts = count_text(&text);
            if counts != (0, 0) {
                let rel = path
                    .strip_prefix(root)
                    .unwrap()
                    .to_string_lossy()
                    .replace('\\', "/");
                out.insert(rel, counts);
            }
        }
    }
    out
}

fn allow_path(root: &Path) -> PathBuf {
    root.join("tests/fixtures/state_folder_literals/allow.tsv")
}

/// `path<TAB>earlier<TAB>new` rows; `#` lines are comments.
fn parse_allow(text: &str) -> Counts {
    let mut out = Counts::new();
    for (i, line) in text.lines().enumerate() {
        let line = line.trim_end();
        if line.is_empty() || line.starts_with('#') {
            continue;
        }
        let cols: Vec<&str> = line.split('\t').collect();
        assert_eq!(
            cols.len(),
            3,
            "allow.tsv line {}: want path, earlier, new",
            i + 1
        );
        let num = |s: &str| {
            s.parse::<usize>()
                .unwrap_or_else(|_| panic!("allow.tsv line {}: {s:?} is no count", i + 1))
        };
        out.insert(cols[0].to_string(), (num(cols[1]), num(cols[2])));
    }
    out
}

const ALLOW_HEADER: &str = "\
# Test files allowed to spell a state folder name, and how many times (SPEC-20261122-state-folder-literal-scan).
# Columns: path, earlier folder count, new folder count. A fresh project or HOME uses the resolvers
# (state_root, home_root, or a Python helper that mirrors them); a literal is for an old-layout
# fixture, a Go-generation folder or a migration. Raise a count only for such a fixture, in the
# change that adds it; counts may fall without editing this file.
# Rewrite from the tree: MOAI_STATE_LITERALS_WRITE=1 cargo test --test state_folder_literal_scan
";

fn render_allow(counts: &Counts) -> String {
    let mut out = String::from(ALLOW_HEADER);
    for (path, (e, n)) in counts {
        out.push_str(&format!("{path}\t{e}\t{n}\n"));
    }
    out
}

/// The violations of `found` against `allow`, one line each.
fn judge(found: &Counts, allow: &Counts) -> Vec<String> {
    let names = [earlier(), new_name()];
    let mut out = Vec::new();
    for (path, &(e, n)) in found {
        let Some(&(ae, an)) = allow.get(path) else {
            out.push(format!(
                "{path}: spells a state folder name ({} x{e}, {} x{n}) and is not in \
                 tests/fixtures/state_folder_literals/allow.tsv",
                names[0], names[1]
            ));
            continue;
        };
        for (name, have, allowed) in [(&names[0], e, ae), (&names[1], n, an)] {
            if have > allowed {
                out.push(format!(
                    "{path}: spells {name} {have} times, the baseline allows {allowed}"
                ));
            }
        }
    }
    out
}

const HOW_TO_FIX: &str = "\
A test that needs a HOME's state folder asks TempHome::state_dir, and one that builds an earlier layout
asks TempHome::make_earlier_state_dir (both in tests/common/home.rs); neither spells a folder name.
Otherwise a fresh project or HOME finds its state folder through the resolvers: moai_cli::state_root::{dir, path, rel}
or moai_cli::home_root::{dir, path} in Rust, a helper that mirrors them in Python. Only a fixture that
builds an earlier layout on purpose (an old-layout fixture, a Go-generation folder, a migration) spells the
name; for such a fixture raise its count in tests/fixtures/state_folder_literals/allow.tsv in the same change.";

#[test]
fn the_repository_tests_stay_within_the_state_folder_baseline() {
    let root = Path::new(MANIFEST);
    let found = scan(root);
    if std::env::var(WRITE_SWITCH).as_deref() == Ok("1") {
        let path = allow_path(root);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(&path, render_allow(&found)).unwrap();
        return;
    }
    let allow = parse_allow(&fs::read_to_string(allow_path(root)).unwrap_or_default());
    let violations = judge(&found, &allow);
    assert!(
        violations.is_empty(),
        "{} test file(s) spell a state folder name beyond the baseline:\n{}\n\n{HOW_TO_FIX}",
        violations.len(),
        violations.join("\n")
    );
}

/// SPEC-20261121-moai-name slice 3: the test seams are `MOAI_TEST_<NAME>`; the earlier prefix
/// (`MOAI_`, `CLI_`, `TEST_` joined) stays nowhere in `src/` or `tests/`. It is joined at run
/// time so this file does not spell it either. History documents are outside the scan.
fn earlier_seam_prefix() -> String {
    format!("MOAI_{}{}", "CLI_", "TEST_")
}

/// `path:line` of every line of `text` that spells the earlier seam prefix.
fn seam_lines(path: &str, text: &str) -> Vec<String> {
    let prefix = earlier_seam_prefix();
    text.lines()
        .enumerate()
        .filter(|(_, line)| line.contains(&prefix))
        .map(|(i, _)| format!("{path}:{}", i + 1))
        .collect()
}

/// The hits under `<root>/src` (`.rs`), `<root>/tests` (`.rs`, `.py`, outside `tests/fixtures`)
/// and `<root>/tools` (`.py`; SPEC-20261124-test-seams-release REQ-3), sorted by path.
fn seam_scan(root: &Path) -> Vec<String> {
    let mut out = Vec::new();
    for (top, exts) in [
        ("src", &["rs"][..]),
        ("tests", &["rs", "py"][..]),
        ("tools", &["py"][..]),
    ] {
        let base = root.join(top);
        let mut stack = vec![base.clone()];
        while let Some(dir) = stack.pop() {
            let Ok(entries) = fs::read_dir(&dir) else {
                continue;
            };
            for entry in entries.flatten() {
                let path = entry.path();
                let Ok(kind) = entry.file_type() else {
                    continue;
                };
                if kind.is_dir() {
                    if path != base.join("fixtures") {
                        stack.push(path);
                    }
                    continue;
                }
                let ext = path.extension().and_then(|e| e.to_str());
                if !kind.is_file() || !ext.is_some_and(|e| exts.contains(&e)) {
                    continue;
                }
                let Ok(text) = fs::read_to_string(&path) else {
                    continue;
                };
                let rel = path
                    .strip_prefix(root)
                    .unwrap()
                    .to_string_lossy()
                    .replace('\\', "/");
                out.extend(seam_lines(&rel, &text));
            }
        }
    }
    out.sort();
    out
}

#[test]
fn no_source_or_test_spells_the_earlier_test_seam_prefix() {
    let hits = seam_scan(Path::new(MANIFEST));
    assert!(
        hits.is_empty(),
        "{} line(s) spell the earlier test seam prefix; a test seam is MOAI_TEST_<NAME>:\n{}",
        hits.len(),
        hits.join("\n")
    );
}

#[test]
fn the_seam_scan_finds_a_planted_line_and_skips_what_it_should() {
    let dir = tempfile::TempDir::new().unwrap();
    let root = dir.path();
    let p = earlier_seam_prefix();
    let files = [
        (
            "src/a.rs",
            format!("fn a() {{}}\nconst V: &str = \"{p}FAULT\";\n"),
        ),
        ("src/deep/b.rs", format!("// {p}MACHINE in a comment\n")),
        (
            "tests/c.py",
            format!("os.environ[\"{p}MACHINE\"] = \"x\"\n"),
        ),
        ("tests/fixtures/d.rs", format!("{p}FAULT\n")),
        (
            "tests/e.rs",
            "const V: &str = \"MOAI_TEST_FAULT\";\n".to_string(),
        ),
        ("tests/notes.md", format!("{p}FAULT\n")),
        (
            "tools/x.py",
            format!("import os\nos.environ[\"{p}MACHINE\"] = \"x\"\n"),
        ),
        ("tools/x.rs", format!("{p}FAULT\n")),
        ("tools/notes.md", format!("{p}FAULT\n")),
    ];
    for (rel, text) in &files {
        let path = root.join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, text).unwrap();
    }
    assert_eq!(
        seam_scan(root),
        [
            "src/a.rs:2",
            "src/deep/b.rs:1",
            "tests/c.py:1",
            "tools/x.py:2"
        ],
        "fixtures, the new name and other file kinds are not hits; tools/ is read for .py only"
    );
}

/// SPEC-20261124-test-seams-release: the test-only seams that a release binary once read (the
/// `agent-N` numbering seam went with the numbering, SPEC-20261124-launcher-retire-role-flags).
/// Each is `(source file, the constant that names the variable)`; the variable is read exactly
/// once in that file, and the read sits directly under `#[cfg(debug_assertions)]`.
const GATED_SEAMS: [(&str, &str); 3] = [
    ("src/lane_registry.rs", "TEST_HOLD_VAR"),
    ("src/launcher_init.rs", "INIT_TEST_HOLD_VAR"),
    ("src/launcher_init.rs", "INIT_TEST_FAIL_VAR"),
];

/// How `text` reads the variable named by the constant `name`: the 1-based line of every read
/// (`env::var(NAME)` or `env::var_os(NAME)`), and whether its line directly follows a
/// `#[cfg(debug_assertions)]` line (comment lines between the two do not count).
fn seam_reads(text: &str, name: &str) -> Vec<(usize, bool)> {
    let reads = [format!("env::var({name})"), format!("env::var_os({name})")];
    let lines: Vec<&str> = text.lines().collect();
    let mut out = Vec::new();
    for (i, line) in lines.iter().enumerate() {
        if !reads.iter().any(|r| line.contains(r.as_str())) {
            continue;
        }
        let above = lines[..i]
            .iter()
            .rev()
            .map(|l| l.trim())
            .find(|l| !l.starts_with("//"));
        out.push((i + 1, above == Some("#[cfg(debug_assertions)]")));
    }
    out
}

#[test]
fn the_test_seams_are_read_only_in_debug_builds() {
    let mut problems = Vec::new();
    for (file, name) in GATED_SEAMS {
        let text = fs::read_to_string(Path::new(MANIFEST).join(file)).unwrap();
        let reads = seam_reads(&text, name);
        if reads.len() != 1 {
            problems.push(format!(
                "{file}: {name} is read {} times, not once",
                reads.len()
            ));
        }
        for (line, gated) in reads {
            if !gated {
                problems.push(format!(
                    "{file}:{line}: the read of {name} is not directly under #[cfg(debug_assertions)]"
                ));
            }
        }
    }
    assert!(
        problems.is_empty(),
        "a release build must not read a test-only seam:\n{}",
        problems.join("\n")
    );
}

#[test]
fn the_gate_scan_tells_a_gated_read_from_a_bare_one() {
    let text = "fn a() {\n    #[cfg(debug_assertions)]\n    if let Some(ms) = std::env::var(HOLD)\n        .ok()\n    {}\n    if std::env::var_os(FAIL).is_some() {}\n    #[cfg(debug_assertions)]\n    // a comment between\n    let _ = std::env::var(OTHER);\n    #[cfg(not(debug_assertions))]\n    let _ = std::env::var(NEG);\n}\n";
    assert_eq!(seam_reads(text, "HOLD"), [(3, true)]);
    assert_eq!(seam_reads(text, "FAIL"), [(6, false)]);
    assert_eq!(seam_reads(text, "OTHER"), [(9, true)]);
    assert_eq!(
        seam_reads(text, "NEG"),
        [(11, false)],
        "a cfg(not(debug_assertions)) read is not gated"
    );
    assert!(seam_reads(text, "NONE").is_empty());
}

/// An old-layout fixture text, as a listed file holds it: it builds the earlier folder first.
fn fixture_text() -> String {
    format!(
        "fn earlier_layout(root: &Path) {{\n    fs::create_dir_all(root.join(\"{e}/receipts\")).unwrap();\n    \
         fs::write(root.join(\"{e}/receipts/a.json\"), \"{{}}\").unwrap();\n}}\n",
        e = earlier()
    )
}

fn one(path: &str, text: &str) -> Counts {
    let counts = count_text(text);
    let mut out = Counts::new();
    if counts != (0, 0) {
        out.insert(path.to_string(), counts);
    }
    out
}

#[test]
fn a_new_fresh_project_line_in_a_listed_file_fails() {
    let path = "tests/old_layout_cli.rs";
    let allow = one(path, &fixture_text());
    assert_eq!(allow[path], (2, 0));
    let grown = format!(
        "{}\n#[test]\nfn fresh_project_receipt() {{\n    assert!(root.join(\"{}/receipts\").is_dir());\n}}\n",
        fixture_text(),
        earlier()
    );
    let violations = judge(&one(path, &grown), &allow);
    assert_eq!(violations.len(), 1, "{violations:?}");
    let v = &violations[0];
    assert!(
        v.contains(path)
            && v.contains(&earlier())
            && v.contains(" 3 times")
            && v.contains("allows 2"),
        "the message names the file, the name and both counts: {v}"
    );
}

#[test]
fn a_listed_fixture_passes_unchanged_and_with_fewer_literals() {
    let path = "tests/old_layout_cli.rs";
    let allow = one(path, &fixture_text());
    assert!(judge(&one(path, &fixture_text()), &allow).is_empty());
    let fewer = fixture_text().replacen(&earlier(), "{state}", 1);
    assert_eq!(count_text(&fewer), (1, 0));
    assert!(judge(&one(path, &fewer), &allow).is_empty());
}

#[test]
fn an_unlisted_python_file_that_spells_the_new_name_fails() {
    let path = "tests/install_like_test.py";
    let text = format!(
        "def state(self, *parts):\n    return os.path.join(self.home, \"{}\", *parts)\n",
        new_name()
    );
    let violations = judge(&one(path, &text), &Counts::new());
    assert_eq!(violations.len(), 1, "{violations:?}");
    assert!(violations[0].contains(path) && violations[0].contains("not in"));
}

#[test]
fn program_names_tags_and_hosts_count_nothing() {
    let text = "Command::new(\"moai-cli\"); let tag = \"moai-cli-v1.2.3\"; \
                let url = \"https://mo.ai.kr\"; let dir = \"x.moai\"; let p = \"target/debug/moai-cli\";";
    assert_eq!(count_text(text), (0, 0), "{text}");
    let both = format!(
        "\"{e}/a\" '{n}' ~/{n}/x {e}\n",
        e = earlier(),
        n = new_name()
    );
    assert_eq!(count_text(&both), (2, 2), "{both}");
    let earlier_with_dash = format!("{}-v1", earlier());
    assert_eq!(count_text(&earlier_with_dash), (0, 0));
}

#[test]
fn the_write_switch_form_is_sorted_and_judges_clean() {
    let dir = tempfile::TempDir::new().unwrap();
    let root = dir.path();
    let files = [
        ("tests/b_cli.rs", fixture_text()),
        (
            "tests/a_test.py",
            format!("HOME_DIR = \"{}\"\n", new_name()),
        ),
        ("tests/fixtures/data.rs", fixture_text()),
        ("tests/clean.rs", "fn main() {}\n".to_string()),
    ];
    for (rel, text) in &files {
        let path = root.join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, text).unwrap();
    }
    let found = scan(root);
    assert_eq!(
        found.keys().cloned().collect::<Vec<_>>(),
        ["tests/a_test.py", "tests/b_cli.rs"],
        "fixtures/ and clean files are not listed"
    );
    let rendered = render_allow(&found);
    let rows: Vec<&str> = rendered.lines().filter(|l| !l.starts_with('#')).collect();
    assert_eq!(rows, ["tests/a_test.py\t0\t1", "tests/b_cli.rs\t2\t0"]);
    assert!(judge(&found, &parse_allow(&rendered)).is_empty());
}

/// SPEC-20261125-state-folder-test-helpers REQ-1, REQ-2: the two `TempHome` helpers give the
/// folder the resolvers name, so a test never spells the folder itself.
#[test]
fn the_state_folder_helpers_agree_with_the_resolver() {
    let fresh = TempHome::new();
    let dir = fresh.state_dir();
    assert_eq!(dir, fresh.path().join(moai_cli::home_root::MOAI_DIR));
    assert_eq!(dir, moai_cli::home_root::path(fresh.path()));
    assert!(dir.is_dir(), "{dir:?} is created");

    let old = TempHome::new();
    let earlier_dir = old.make_earlier_state_dir();
    assert_eq!(earlier_dir, old.path().join(moai_cli::HOME_STATE_DIR));
    assert!(earlier_dir.is_dir(), "{earlier_dir:?} is created");
    assert_eq!(
        moai_cli::home_root::dir(old.path()),
        moai_cli::HOME_STATE_DIR,
        "the resolver names the earlier folder for a HOME built that way"
    );
}

/// SPEC-20261125-state-folder-test-helpers REQ-1, REQ-2: a test that uses the helpers spells no
/// name, so it needs no `allow.tsv` row; the same text with a literal is flagged.
#[test]
fn a_test_that_uses_the_helpers_spells_no_state_folder_name() {
    let path = "tests/helper_user_cli.rs";
    let text = "#[test]\nfn finds_the_state() {\n    let home = TempHome::new();\n    \
                let state = home.state_dir();\n    let old = TempHome::new();\n    \
                let earlier = old.make_earlier_state_dir();\n    assert!(state.is_dir() && earlier.is_dir());\n}\n";
    assert_eq!(count_text(text), (0, 0), "{text}");
    assert!(judge(&one(path, text), &Counts::new()).is_empty());

    let spelled = format!(
        "{text}fn literal(home: &TempHome) {{ let _ = home.path().join(\"{}\"); }}\n",
        earlier()
    );
    assert_eq!(count_text(&spelled), (1, 0), "{spelled}");
    let violations = judge(&one(path, &spelled), &Counts::new());
    assert_eq!(violations.len(), 1, "{violations:?}");
    assert!(violations[0].contains(path), "{violations:?}");
}

/// SPEC-20261125-state-folder-test-helpers REQ-3: the rejection tells a test author the helpers
/// by name, before the resolvers and the `allow.tsv` row rule.
#[test]
fn the_scan_rejection_names_the_helpers() {
    for helper in ["TempHome::state_dir", "TempHome::make_earlier_state_dir"] {
        assert!(
            HOW_TO_FIX.contains(helper),
            "{helper} is missing:\n{HOW_TO_FIX}"
        );
    }
    let helper_at = HOW_TO_FIX.find("TempHome::state_dir").unwrap();
    let resolver_at = HOW_TO_FIX.find("moai_cli::state_root").unwrap();
    assert!(
        helper_at < resolver_at,
        "the helpers come before the resolvers"
    );
}
