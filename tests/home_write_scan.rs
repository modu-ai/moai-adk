//! The places under `src/` that resolve a HOME state folder, and what each of them does about
//! the one write check (`home_write::check`), are listed below. The comparison reports every
//! difference. By default (observe) it only reports; `MOAI_HOME_WRITE_CHECK=enforce` fails the
//! run of the checked-in sources on a difference.
//!
//! A row names one function and how many resolving uses it holds. `Entry` and `BestEffort` rows
//! must call the check (or one of the older refusals) in the function body; `Under` rows name the
//! entry row whose call covers them; `Read`, `WayOut` and `Resolver` rows say why they pass
//! without it; `Unknown` rows are functions nobody has read yet and must be gone before the
//! switch to enforce.
//!
//! The detector reads source text: comments, string contents and `#[cfg(test)]` modules that
//! start in the first column are skipped. It sees only module-qualified calls, so a call made
//! inside the module that defines the function, a function brought in under another name and a
//! HOME held in a variable whose name has no `home` in it are not seen.

use std::collections::{BTreeMap, BTreeSet};
use std::fs;
use std::io::Write;
use std::path::Path;

const MODE_VARIABLE: &str = "MOAI_HOME_WRITE_CHECK";
const PREFIX: &str = "home-write-check observe:";
/// Fewer resolving uses than this means the scan reads fewer files than it should.
const FLOOR: usize = 60;

/// What a function does about the write check.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum Class {
    /// A write entry: its body calls the check.
    Entry,
    /// Covered by the entry row named in the note (`file::function`).
    Under,
    /// Reads only; the note says what.
    Read,
    /// A clean-up command that must work on a refused HOME; the note says why and where it writes.
    WayOut,
    /// Defines or forwards a root; the note says which.
    Resolver,
    /// A write that must not fail its caller: the body calls the check and skips the write.
    BestEffort,
    /// Not read yet.
    Unknown,
}

/// A function, how many resolving uses it holds, its class and a note.
type Row = (&'static str, &'static str, usize, Class, &'static str);

/// The calls that resolve a HOME state folder: the module that must qualify the call and the
/// function name.
const ROOT_CALLS: [(&str, &str); 14] = [
    ("home_root", "path"),
    ("home_root", "dir"),
    ("home_workdir", "root"),
    ("home_workdir", "projects_root"),
    ("home_workdir", "reports_root"),
    ("home_workdir", "reports_dir"),
    ("home_workdir", "worktrees_root"),
    ("home_workdir", "service_dir"),
    ("home_layout", "rust_path"),
    ("layout", "default_home_root"),
    ("layout", "switched_dir"),
    ("layout", "switched_dir_in"),
    ("layout", "switched_dir_checked"),
    ("project_home", "resolve"),
];

/// Constants that name a state folder; joined onto a value whose name holds `home` they count.
const FOLDER_NAMES: [&str; 2] = ["MOAI_DIR", "HOME_STATE_DIR"];

/// Calls that stand for the check besides `home_write::check`: the two older refusals.
const OLDER_GATES: [&str; 3] = ["check_ready", "check_home_ready", "backup_root"];

#[derive(Debug, PartialEq)]
enum Token {
    Word(String),
    Text(String),
    Mark(char),
}

/// The tokens of Rust source with comments dropped; the flag says the token starts its line.
fn tokens(source: &str) -> Vec<(Token, bool)> {
    let s: Vec<char> = source.chars().collect();
    let at = |i: usize| s.get(i).copied().unwrap_or('\0');
    let word = |c: char| c.is_alphanumeric() || c == '_';
    let mut out = Vec::new();
    let mut i = 0;
    while i < s.len() {
        let c = s[i];
        let first = i == 0 || s[i - 1] == '\n';
        if c.is_whitespace() {
            i += 1;
        } else if c == '/' && at(i + 1) == '/' {
            while i < s.len() && s[i] != '\n' {
                i += 1;
            }
        } else if c == '/' && at(i + 1) == '*' {
            let mut depth = 0usize;
            while i < s.len() {
                if s[i] == '/' && at(i + 1) == '*' {
                    depth += 1;
                    i += 2;
                } else if s[i] == '*' && at(i + 1) == '/' {
                    depth -= 1;
                    i += 2;
                    if depth == 0 {
                        break;
                    }
                } else {
                    i += 1;
                }
            }
        } else if c == '"' {
            let start = i + 1;
            i = start;
            while i < s.len() && s[i] != '"' {
                i += if s[i] == '\\' { 2 } else { 1 };
            }
            let end = i.min(s.len());
            out.push((Token::Text(s[start..end].iter().collect()), first));
            i += 1;
        } else if c == '\'' {
            if at(i + 1) == '\\' {
                // An escaped character literal.
                i += 3;
                while i < s.len() && s[i] != '\'' {
                    i += 1;
                }
                i += 1;
            } else if at(i + 2) == '\'' {
                i += 3;
            } else {
                // A lifetime or a label.
                i += 1;
            }
        } else if word(c) {
            let start = i;
            while i < s.len() && word(s[i]) {
                i += 1;
            }
            let name: String = s[start..i].iter().collect();
            let hashes = s[i..].iter().take_while(|&&c| c == '#').count();
            if matches!(name.as_str(), "r" | "br" | "cr") && at(i + hashes) == '"' {
                // A raw string ends at a quote followed by as many `#` as opened it.
                let from = i + hashes + 1;
                let mut j = from;
                while j < s.len() && !(s[j] == '"' && (1..=hashes).all(|k| at(j + k) == '#')) {
                    j += 1;
                }
                let end = j.min(s.len());
                out.push((Token::Text(s[from..end].iter().collect()), first));
                i = j + 1 + hashes;
            } else {
                out.push((Token::Word(name), first));
            }
        } else {
            out.push((Token::Mark(c), first));
            i += 1;
        }
    }
    out
}

/// Where the rows are matched with the source: file and enclosing function.
type Key = (String, String);

/// What one source file holds: the resolving uses by function, the functions that call a gate,
/// and the bare imports of a resolving function.
#[derive(Default)]
struct Found {
    uses: BTreeMap<Key, usize>,
    gated: BTreeSet<Key>,
    bare: Vec<String>,
}

fn scan(file: &str, source: &str, found: &mut Found) {
    let list = tokens(source);
    let word = |i: usize| match list.get(i) {
        Some((Token::Word(w), _)) => w.as_str(),
        _ => "",
    };
    let mark = |i: usize, c: char| matches!(list.get(i), Some((Token::Mark(m), _)) if *m == c);
    // `module::` stands right before index `i`.
    let qualified = |i: usize, module: &str| {
        i >= 3 && word(i - 3) == module && mark(i - 2, ':') && mark(i - 1, ':')
    };
    // A `cfg` predicate that holds in test builds only: `test`, or `all(…)` naming it directly.
    let only_in_tests = |from: usize| {
        if word(from) == "test" {
            return mark(from + 1, ')');
        }
        if word(from) != "all" {
            return false;
        }
        let mut depth = 0i32;
        for j in from + 1..list.len() {
            if mark(j, '(') {
                depth += 1;
            } else if mark(j, ')') {
                depth -= 1;
            } else if depth == 1 && word(j) == "test" {
                return true;
            }
            if depth == 0 {
                break;
            }
        }
        false
    };
    let mut braces = 0usize;
    let mut open: Vec<(String, usize)> = Vec::new();
    let mut pending: Option<(String, i32)> = None;
    let mut in_use = false;
    let mut i = 0;
    while i < list.len() {
        let test_only = list[i].1
            && mark(i, '#')
            && mark(i + 1, '[')
            && word(i + 2) == "cfg"
            && mark(i + 3, '(')
            && only_in_tests(i + 4);
        if test_only {
            let mut j = i;
            while mark(j, '#') {
                let mut depth = 0;
                j += 1;
                loop {
                    if mark(j, '[') {
                        depth += 1;
                    } else if mark(j, ']') {
                        depth -= 1;
                    }
                    j += 1;
                    if depth == 0 || j >= list.len() {
                        break;
                    }
                }
            }
            if word(j) == "pub" {
                j += 1;
                if mark(j, '(') {
                    while j < list.len() && !mark(j, ')') {
                        j += 1;
                    }
                    j += 1;
                }
            }
            if word(j) == "mod" && mark(j + 2, '{') {
                let mut depth = 0usize;
                j += 2;
                while j < list.len() {
                    if mark(j, '{') {
                        depth += 1;
                    } else if mark(j, '}') {
                        depth -= 1;
                        if depth == 0 {
                            break;
                        }
                    }
                    j += 1;
                }
                i = j + 1;
                continue;
            }
        }
        let here = || {
            let function = open
                .last()
                .map_or_else(|| "<file>".to_string(), |(name, _)| name.clone());
            (file.to_string(), function)
        };
        if word(i) == "fn" && !word(i + 1).is_empty() {
            pending = Some((word(i + 1).to_string(), 0));
            i += 2;
            continue;
        }
        if word(i) == "use" {
            in_use = true;
        }
        if mark(i, ';') {
            in_use = false;
        }
        // A bare import of a resolving function: `use …::home_root::{path}`.
        if in_use {
            for (module, name) in ROOT_CALLS {
                if word(i) == module && mark(i + 1, ':') && mark(i + 2, ':') {
                    let group = mark(i + 3, '{');
                    let single = word(i + 3) == name && !mark(i + 4, '(');
                    let in_group = group && {
                        let mut j = i + 4;
                        let mut hit = false;
                        while j < list.len() && !mark(j, '}') {
                            hit |= word(j) == name;
                            j += 1;
                        }
                        hit
                    };
                    if single || in_group {
                        found.bare.push(format!("{file}: {module}::{name}"));
                    }
                }
            }
        }
        if mark(i + 1, '(') && !in_use {
            for (module, name) in ROOT_CALLS {
                if word(i) == name && qualified(i, module) {
                    *found.uses.entry(here()).or_insert(0) += 1;
                }
            }
            if (word(i) == "check" && qualified(i, "home_write")) || OLDER_GATES.contains(&word(i))
            {
                found.gated.insert(here());
            }
            // A state folder constant joined onto a value whose name holds `home`.
            if word(i) == "join"
                && i >= 2
                && mark(i - 1, '.')
                && word(i - 2).to_lowercase().contains("home")
            {
                let mut depth = 0i32;
                let mut hit = false;
                let mut j = i + 1;
                while j < list.len() {
                    if mark(j, '(') {
                        depth += 1;
                    } else if mark(j, ')') {
                        depth -= 1;
                        if depth == 0 {
                            break;
                        }
                    } else {
                        hit |= FOLDER_NAMES.contains(&word(j));
                    }
                    j += 1;
                }
                if hit {
                    *found.uses.entry(here()).or_insert(0) += 1;
                }
            }
        }
        if let Some((Token::Mark(c), _)) = list.get(i) {
            let in_signature = pending.as_ref().map(|(_, depth)| *depth);
            match (*c, in_signature) {
                ('(' | '[', Some(_)) => pending.as_mut().unwrap().1 += 1,
                (')' | ']', Some(_)) => pending.as_mut().unwrap().1 -= 1,
                (';', Some(0)) => pending = None,
                ('{', _) => {
                    braces += 1;
                    if in_signature == Some(0) {
                        open.push((pending.take().unwrap().0, braces));
                    }
                }
                ('}', _) => {
                    braces = braces.saturating_sub(1);
                    while open.last().is_some_and(|(_, depth)| *depth > braces) {
                        open.pop();
                    }
                }
                _ => {}
            }
        }
        i += 1;
    }
}

/// One line for every difference between the rows and the source.
fn compare(files: &[(String, String)], rows: &[Row], floor: usize) -> Vec<String> {
    let mut found = Found::default();
    for (file, source) in files {
        scan(file, source, &mut found);
    }
    let mut lines = Vec::new();
    let total: usize = found.uses.values().sum();
    if total < floor {
        lines.push(format!(
            "the scan found {total} resolving uses, fewer than the floor of {floor}"
        ));
    }
    let mut listed: BTreeMap<Key, usize> = BTreeMap::new();
    for (file, function, count, _, _) in rows {
        *listed
            .entry((file.to_string(), function.to_string()))
            .or_insert(0) += count;
    }
    for (key, count) in &found.uses {
        for _ in listed.get(key).copied().unwrap_or(0)..*count {
            lines.push(format!("unlisted HOME root use in {} fn {}", key.0, key.1));
        }
    }
    for (key, count) in &listed {
        for _ in found.uses.get(key).copied().unwrap_or(0)..*count {
            lines.push(format!(
                "listed HOME root use not found in {} fn {}",
                key.0, key.1
            ));
        }
    }
    for entry in &found.bare {
        lines.push(format!("bare import of a HOME root function in {entry}"));
    }
    let entries: BTreeSet<String> = rows
        .iter()
        .filter(|r| r.3 == Class::Entry)
        .map(|r| format!("{}::{}", r.0, r.1))
        .collect();
    for (file, function, _, class, note) in rows {
        let key = (file.to_string(), function.to_string());
        let present = found.uses.contains_key(&key);
        match class {
            Class::Entry | Class::BestEffort if present && !found.gated.contains(&key) => {
                lines.push(format!("{file} fn {function}: no call of the write check"));
            }
            Class::Under if !entries.contains(*note) => {
                lines.push(format!(
                    "{file} fn {function}: under `{note}`, which is no entry row"
                ));
            }
            Class::Read | Class::WayOut | Class::Resolver if note.is_empty() => {
                lines.push(format!("{file} fn {function}: no reason given"));
            }
            Class::Unknown => lines.push(format!("{file} fn {function}: not read yet")),
            _ => {}
        }
    }
    lines
}

/// `enforce` fails with the lines; any other mode writes them to `out` and passes.
fn judge(lines: &[String], mode: Option<&str>, out: &mut dyn Write) -> Result<(), String> {
    if mode == Some("enforce") {
        return if lines.is_empty() {
            Ok(())
        } else {
            Err(lines.join("\n"))
        };
    }
    for line in lines {
        writeln!(out, "{PREFIX} {line}").map_err(|e| e.to_string())?;
    }
    Ok(())
}

fn rust_files(root: &Path, dir: &Path, out: &mut Vec<(String, String)>) {
    let mut entries: Vec<_> = fs::read_dir(dir)
        .unwrap()
        .map(|e| e.unwrap().path())
        .collect();
    entries.sort();
    for path in entries {
        if path.is_dir() {
            rust_files(root, &path, out);
        } else if path.extension().is_some_and(|e| e == "rs") {
            let name = path
                .strip_prefix(root)
                .unwrap()
                .to_string_lossy()
                .replace('\\', "/");
            out.push((name, fs::read_to_string(&path).unwrap()));
        }
    }
}

#[test]
fn every_home_root_use_in_product_code_is_classified() {
    let root = Path::new(env!("CARGO_MANIFEST_DIR"));
    let mut files = Vec::new();
    rust_files(root, &root.join("src"), &mut files);
    let lines = compare(&files, ROWS, FLOOR);
    let mode = std::env::var(MODE_VARIABLE).ok();
    // Written to the process's stderr, not through `eprintln!`, so a passing run shows it.
    if let Err(differences) = judge(&lines, mode.as_deref(), &mut std::io::stderr().lock()) {
        panic!("HOME root uses differ from the rows in tests/home_write_scan.rs:\n{differences}");
    }
}

fn planted(file: &str, source: &str) -> Vec<(String, String)> {
    vec![(file.to_string(), source.to_string())]
}

/// A function that reads only, as a row.
fn read_row(function: &'static str, count: usize) -> Row {
    ("src/demo.rs", function, count, Class::Read, "reads only")
}

#[test]
fn an_unlisted_root_use_is_reported_and_enforce_fails() {
    let source = r##"
//! home_root::path(h) in a module comment
/* home_workdir::root(h) /* nested */ home_root::dir(h) */
pub fn listed<'a>(h: &'a Path, x: &'a [u8; 2]) -> Option<PathBuf> {
    let quote = '"';
    let raw = r#"home_root::path(h)"#;
    let text = "home_root::path(h)";
    let other = other_module::path(h);
    Some(crate::home_root::path(h)) // home_root::dir(h)
}
trait T {
    fn declared(&self) -> [u8; 4];
}
#[cfg(test)]
#[allow(dead_code)]
mod tests {
    fn helper(h: &Path) {
        home_root::path(h);
    }
}
#[cfg(all(test, unix))]
mod unix_tests {
    fn helper(h: &Path) {
        home_workdir::projects_root(h);
    }
}
"##;
    let rows = [read_row("listed", 1)];
    let equal = compare(&planted("src/demo.rs", source), &rows, 1);
    assert!(equal.is_empty(), "{equal:?}");

    let more = format!(
        "{source}\nfn added(h: &Path) {{\n    let a = || home_workdir::reports_dir(\n        h, p,\n    );\n}}\n"
    );
    let lines = compare(&planted("src/demo.rs", &more), &rows, 1);
    assert_eq!(lines, ["unlisted HOME root use in src/demo.rs fn added"]);

    // A module that is also built outside tests is product source.
    let built = format!(
        "{source}#[cfg(all(not(test), unix))]\nmod real {{ fn built(h: &Path) {{ home_root::dir(h); }} }}\n"
    );
    let lines = compare(&planted("src/demo.rs", &built), &rows, 1);
    assert_eq!(lines, ["unlisted HOME root use in src/demo.rs fn built"]);

    let gone = compare(&planted("src/demo.rs", "fn listed() {}\n"), &rows, 0);
    assert_eq!(
        gone,
        ["listed HOME root use not found in src/demo.rs fn listed"]
    );

    let join = "fn joined(home: &Path) { let p = home.join(crate::home_root::MOAI_DIR); \
                let q = base.join(MOAI_DIR); let r = home.join(\"x\"); }\n";
    let lines = compare(&planted("src/demo.rs", join), &[], 0);
    assert_eq!(lines, ["unlisted HOME root use in src/demo.rs fn joined"]);

    let bare = "use crate::home_root::{dir, path};\nuse crate::layout::switched_dir;\nuse crate::home_root::MOAI_DIR;\nfn f() {}\n";
    let lines = compare(&planted("src/demo.rs", bare), &[], 0);
    assert_eq!(
        lines,
        [
            "bare import of a HOME root function in src/demo.rs: home_root::path",
            "bare import of a HOME root function in src/demo.rs: home_root::dir",
            "bare import of a HOME root function in src/demo.rs: layout::switched_dir",
        ]
    );

    let few = compare(&planted("src/demo.rs", "fn f() {}\n"), &[], 3);
    assert_eq!(
        few,
        ["the scan found 0 resolving uses, fewer than the floor of 3"]
    );

    let mut out = Vec::new();
    let failure = judge(&lines, Some("enforce"), &mut out).unwrap_err();
    assert!(failure.contains(&lines[0]), "{failure}");
}

#[test]
fn an_entry_row_without_the_check_is_reported_and_the_other_classes_hold_their_rules() {
    let without = "fn write(h: &Path) { let d = home_root::path(h); }\n";
    let with = "fn write(h: &Path) { home_write::check(h)?; let d = home_root::path(h); }\n";
    let older = "fn write(h: &Path) { let d = home_root::path(h); let b = backup_root(h); }\n";
    let wrong_module = "fn write(h: &Path) { other::check(h)?; let d = home_root::path(h); }\n";
    let row = |class, note| [("src/demo.rs", "write", 1, class, note)];
    for class in [Class::Entry, Class::BestEffort] {
        let rows = row(class, "pending");
        let lines = compare(&planted("src/demo.rs", without), &rows, 1);
        assert_eq!(
            lines,
            ["src/demo.rs fn write: no call of the write check"],
            "{class:?}"
        );
        let lines = compare(&planted("src/demo.rs", wrong_module), &rows, 1);
        assert_eq!(lines.len(), 1, "{class:?}: {lines:?}");
        for good in [with, older] {
            let lines = compare(&planted("src/demo.rs", good), &rows, 1);
            assert!(lines.is_empty(), "{class:?}: {lines:?}");
        }
    }

    let under = [("src/demo.rs", "write", 1, Class::Under, "src/other.rs::run")];
    let lines = compare(&planted("src/demo.rs", without), &under, 1);
    assert_eq!(
        lines,
        ["src/demo.rs fn write: under `src/other.rs::run`, which is no entry row"]
    );
    let both = [
        ("src/demo.rs", "write", 1, Class::Under, "src/other.rs::run"),
        ("src/other.rs", "run", 1, Class::Entry, "pending"),
    ];
    let files = vec![
        ("src/demo.rs".to_string(), without.to_string()),
        (
            "src/other.rs".to_string(),
            "fn run(h: &Path) { home_write::check(h)?; home_workdir::root(h); }\n".to_string(),
        ),
    ];
    let lines = compare(&files, &both, 2);
    assert!(lines.is_empty(), "{lines:?}");

    for class in [Class::Read, Class::WayOut, Class::Resolver] {
        let lines = compare(&planted("src/demo.rs", without), &row(class, ""), 1);
        assert_eq!(
            lines,
            ["src/demo.rs fn write: no reason given"],
            "{class:?}"
        );
        let lines = compare(&planted("src/demo.rs", without), &row(class, "why"), 1);
        assert!(lines.is_empty(), "{class:?}: {lines:?}");
    }

    let lines = compare(
        &planted("src/demo.rs", without),
        &row(Class::Unknown, ""),
        1,
    );
    assert_eq!(lines, ["src/demo.rs fn write: not read yet"]);
}

/// A planted difference: one use the empty rows do not hold.
fn one_difference() -> Vec<String> {
    let source = "fn added(h: &Path) { home_root::path(h); }\n";
    compare(&planted("src/demo.rs", source), &[], 0)
}

#[test]
fn enforce_mode_fails_on_a_difference() {
    let lines = one_difference();
    assert_eq!(lines.len(), 1, "{lines:?}");
    let mut out = Vec::new();
    let failure = judge(&lines, Some("enforce"), &mut out).unwrap_err();
    assert!(failure.contains(&lines[0]), "{failure}");
    assert_eq!(judge(&[], Some("enforce"), &mut out), Ok(()));
}

#[test]
fn observe_mode_reports_and_passes() {
    let mut lines = one_difference();
    lines.push("a second line".to_string());
    for mode in [None, Some(""), Some("observe")] {
        let mut out = Vec::new();
        assert_eq!(judge(&lines, mode, &mut out), Ok(()), "{mode:?}");
        let shown = String::from_utf8(out).unwrap();
        let expected: Vec<String> = lines.iter().map(|l| format!("{PREFIX} {l}")).collect();
        assert_eq!(shown.lines().collect::<Vec<_>>(), expected, "{mode:?}");
    }
}

/// The rows of the checked-in sources; the next slices fill in the `Unknown` ones as they read
/// them and give each `Entry` its runtime check in `tests/home_write_cli.rs`.
const ROWS: &[Row] = &[
    ("src/agent_install.rs", "read_state", 1, Class::Unknown, "not read yet"),
    ("src/agent_install.rs", "state_root", 1, Class::Resolver, "returns the agents state root"),
    ("src/apply.rs", "receipts_dir", 1, Class::Resolver, "returns the project receipts folder through the layout"),
    ("src/bin_names.rs", "versions_dir_of", 1, Class::Unknown, "not read yet"),
    ("src/build_number.rs", "place", 1, Class::Unknown, "not read yet"),
    ("src/card/command.rs", "export", 1, Class::Unknown, "not read yet"),
    ("src/card/finish.rs", "report_path", 1, Class::Unknown, "not read yet"),
    ("src/claim.rs", "journal_root_in", 1, Class::Resolver, "returns the claim journal folder"),
    ("src/edit_diagnostics.rs", "lock_dir", 1, Class::BestEffort, "the edit lock must not fail the host call"),
    ("src/forge/outbox.rs", "outbox_dir", 1, Class::Resolver, "returns the outbox folder through the layout"),
    ("src/gate_schedule.rs", "plan", 1, Class::Read, "schedule preview; writes nothing"),
    ("src/gate_schedule.rs", "schedule_dir", 1, Class::Resolver, "returns the schedule folder"),
    ("src/go_settings_clean.rs", "launcher_path", 1, Class::Unknown, "not read yet"),
    ("src/go_settings_clean.rs", "launcher_shown", 2, Class::Unknown, "not read yet"),
    ("src/home_install_apply.rs", "move_items", 1, Class::WayOut, "exempt until the move command is deleted; writes inside the state folder"),
    ("src/home_install_apply.rs", "new_target", 1, Class::WayOut, "exempt until the move command is deleted; names the new folder"),
    ("src/home_install_apply.rs", "of", 1, Class::WayOut, "exempt until the move command is deleted; names a path"),
    ("src/home_install.rs", "alias_table", 1, Class::Unknown, "not read yet"),
    ("src/home_install.rs", "committed_aliases", 1, Class::Unknown, "not read yet"),
    ("src/home_install.rs", "old_root", 1, Class::Unknown, "not read yet"),
    ("src/home_install.rs", "plan", 2, Class::Read, "move preview; writes nothing"),
    ("src/home_install.rs", "registration_blocked", 1, Class::Read, "answers whether the ledger may register; writes nothing"),
    ("src/home_layout.rs", "root_in", 1, Class::Resolver, "pure layout root of a HOME"),
    ("src/home_retire.rs", "apply_checked", 1, Class::WayOut, "go clean-up apply: renames Go items out of the state folder into a dated folder beside HOME, so it writes outside as well as inside; stops itself on interrupted and refuse"),
    ("src/home_retire.rs", "plan", 1, Class::Read, "go clean-up preview; writes nothing"),
    ("src/home_retire.rs", "revert", 1, Class::WayOut, "go clean-up revert: renames the items back; same stops"),
    ("src/home_root.rs", "adopt_new", 2, Class::Resolver, "defines the HOME resolver; creates the folder only for a new HOME and checks that itself"),
    ("src/home_root.rs", "backup_root", 1, Class::Resolver, "older refusal of the four generations; its text stays"),
    ("src/home_root.rs", "is_new_home", 1, Class::Resolver, "defines the HOME resolver; reads only"),
    ("src/home_root.rs", "path", 3, Class::Resolver, "defines the HOME resolver; reads only"),
    ("src/home_routing.rs", "new_receipt_dir", 1, Class::Resolver, "makes the receipt folder for the apply entries"),
    ("src/home_routing.rs", "read_state", 1, Class::Unknown, "not read yet"),
    ("src/home_routing.rs", "revert_in", 1, Class::Entry, "restores routing state under HOME"),
    ("src/home_routing.rs", "state_root", 1, Class::Resolver, "returns the routing state root"),
    ("src/home_routing/rules.rs", "read_rules_state", 1, Class::Unknown, "not read yet"),
    ("src/home_state_retire.rs", "apply", 1, Class::Entry, "archive apply writes inside the state folder with no generation check today"),
    ("src/home_state_retire.rs", "earlier", 1, Class::Unknown, "not read yet"),
    ("src/home_state_retire.rs", "names", 1, Class::Unknown, "not read yet"),
    ("src/home_state_retire.rs", "plan", 1, Class::Read, "archive preview; writes nothing"),
    ("src/home_state_retire.rs", "revert", 2, Class::Entry, "archive revert renames inside the state folder with no generation check today"),
    ("src/home_state.rs", "apply", 1, Class::WayOut, "moves the Go state folder aside and makes a new one; checks the generation itself (go only)"),
    ("src/home_state.rs", "plan", 1, Class::Read, "home-state preview; writes nothing"),
    ("src/home_state.rs", "plan_selective", 1, Class::Read, "home-state preview; writes nothing"),
    ("src/home_state.rs", "revert", 1, Class::WayOut, "puts the Go folder back; checks the generation itself (rust only)"),
    ("src/home_style.rs", "new_receipt_dir", 1, Class::Resolver, "makes the receipt folder for the apply entries"),
    ("src/home_style.rs", "read_state", 1, Class::Unknown, "not read yet"),
    ("src/home_style.rs", "revert_in", 1, Class::Entry, "restores style state under HOME"),
    ("src/home_style.rs", "state_root", 1, Class::Resolver, "returns the style state root"),
    ("src/hook_config.rs", "load", 1, Class::Read, "loads the hook configuration"),
    ("src/hooks.rs", "home_area_dir", 2, Class::BestEffort, "hook receipts must not fail the host call"),
    ("src/hooks.rs", "project_root", 3, Class::Unknown, "not read yet"),
    ("src/identity.rs", "identity_in", 1, Class::Unknown, "not read yet"),
    ("src/install.rs", "inspect", 1, Class::Read, "install-inspect; writes nothing"),
    ("src/install.rs", "installations_dir", 1, Class::Resolver, "returns the installations folder"),
    ("src/install.rs", "read_receipts", 1, Class::Read, "reads receipts"),
    ("src/key_vault.rs", "state_dir", 1, Class::Resolver, "returns the vault root; the key commands get their check at the command"),
    ("src/language.rs", "home_conversation", 1, Class::Unknown, "not read yet"),
    ("src/language.rs", "set_home_mode", 1, Class::Entry, "writes config/language.yaml under HOME"),
    ("src/ledger/folders.rs", "components", 2, Class::Unknown, "not read yet"),
    ("src/ledger/folders.rs", "modes", 1, Class::Unknown, "not read yet"),
    ("src/ledger/location.rs", "project_dir", 1, Class::Resolver, "returns the ledger folder; the ledger writers check the generation themselves"),
    ("src/optimize.rs", "plan", 1, Class::Unknown, "not read yet"),
    ("src/project_home.rs", "resolve", 1, Class::Unknown, "not read yet"),
    ("src/recovery.rs", "backup_manifest", 1, Class::Unknown, "not read yet"),
    ("src/recovery.rs", "known_good", 1, Class::Unknown, "not read yet"),
    ("src/reports_move.rs", "plan", 3, Class::Read, "reports move preview; writes nothing"),
    ("src/session.rs", "user_launcher", 1, Class::Unknown, "not read yet"),
    ("src/session/machine_names.rs", "folder", 1, Class::Resolver, "returns the session names folder"),
    ("src/star_ask.rs", "dismissal_file", 1, Class::Resolver, "returns the marker path"),
    ("src/star_ask.rs", "file", 1, Class::Resolver, "returns a state file path"),
    ("src/star_ask.rs", "store", 1, Class::Entry, "stores the report consent under HOME; no product caller today"),
    ("src/star_ask.rs", "write_state_marker", 2, Class::Entry, "writes the dismissal marker under HOME"),
    ("src/statusline/state.rs", "home_state", 1, Class::Unknown, "not read yet"),
    ("src/tidy/mod.rs", "open_moai", 1, Class::Unknown, "not read yet"),
    ("src/tidy/mod.rs", "scan", 1, Class::Unknown, "not read yet"),
    ("src/update_cleanup.rs", "archive_base", 1, Class::Resolver, "returns the clean-up archive folder; the clean-up commands get the check"),
    ("src/update_cleanup.rs", "go_manifest", 1, Class::Unknown, "not read yet"),
    ("src/update_cleanup.rs", "in_use", 1, Class::Unknown, "not read yet"),
    ("src/update_cleanup.rs", "load_receipt", 1, Class::Unknown, "not read yet"),
    ("src/update_cleanup.rs", "moai_items", 1, Class::Unknown, "not read yet"),
    ("src/update_cleanup.rs", "plan_json", 1, Class::Unknown, "not read yet"),
    ("src/update_models.rs", "revert", 1, Class::Entry, "restores the model settings and receipts under HOME"),
    ("src/update_models.rs", "settings_path", 1, Class::Resolver, "returns the settings path"),
    ("src/update.rs", "cross_filesystem", 1, Class::Unknown, "not read yet"),
    ("src/update.rs", "execution_link_version", 1, Class::Unknown, "not read yet"),
    ("src/update.rs", "host_receipt_path", 1, Class::Resolver, "update helper; apply_update_in gets the check"),
    ("src/update.rs", "known_good_copies", 1, Class::Unknown, "not read yet"),
    ("src/update.rs", "refuse_backup_roots", 1, Class::Unknown, "not read yet"),
    ("src/update.rs", "resolve_host_receipt", 1, Class::Resolver, "update helper; reads"),
    ("src/update/extra_skill.rs", "known_good_copies", 1, Class::Unknown, "not read yet"),
    ("src/update/extra_skill.rs", "receipt_path", 1, Class::Unknown, "not read yet"),
    ("src/update/extra_skill.rs", "resolve_receipt", 1, Class::Unknown, "not read yet"),
    ("src/update/run_receipt.rs", "write_record", 1, Class::Unknown, "not read yet"),
    ("src/update/switch.rs", "link_dir", 1, Class::Resolver, "returns the link folder; reads"),
    ("src/worktree_gc.rs", "work_roots", 1, Class::Unknown, "not read yet"),
    ("src/worktree_move.rs", "plan", 1, Class::Unknown, "not read yet"),
    ("src/worktree_move.rs", "refuse_linked_move_root", 1, Class::Unknown, "not read yet"),
];
