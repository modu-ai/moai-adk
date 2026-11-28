//! SPEC-20261124-home-env-readers: the places under `src/` that read `HOME` from the process
//! environment, and the places outside a command edge that call a function reading it for them,
//! are listed below. The comparison reports every difference. By default (observe) it only
//! reports; `MOAI_HOME_READERS=enforce` fails the run of the checked-in sources on a difference.
//!
//! The detector reads source text: comments, string contents and `#[cfg(test)]` modules that
//! start in the first column are skipped. It does not see a bare `home()` call brought in under
//! another name, a variable name held in a constant, or a child process such as git reading HOME.

use std::collections::BTreeMap;
use std::fs;
use std::io::Write;
use std::path::Path;

const MODE_VARIABLE: &str = "MOAI_HOME_READERS";
const PREFIX: &str = "home-env-readers observe:";

/// A direct read: the file, its enclosing function and how many reads that function holds.
type Read = (&'static str, &'static str, usize);
/// A call site of an adapter: the adapter, the file, the enclosing function and the count.
type Call = (&'static str, &'static str, &'static str, usize);

/// Direct reads at a command edge.
const EDGE_READS: &[Read] = &[
    ("src/hook_entry.rs", "env", 1),
    ("src/ledger/command.rs", "resolve_home", 1),
    ("src/main.rs", "main", 1),
    ("src/main.rs", "run_internal_classified", 5),
    ("src/ops/baseline_inventory.rs", "run", 1),
    ("src/ops/bin_names_apply.rs", "run", 1),
    ("src/ops/bin_names_plan.rs", "run", 1),
    ("src/ops/go_hooks_apply.rs", "run", 1),
    ("src/ops/go_hooks_plan.rs", "run", 1),
    ("src/ops/go_settings_apply.rs", "run", 1),
    ("src/ops/go_settings_plan.rs", "run", 1),
    ("src/ops/home_archive_apply.rs", "run", 1),
    ("src/ops/home_archive_plan.rs", "run", 1),
    ("src/ops/home_archive_revert.rs", "run", 1),
    ("src/ops/home_install_plan.rs", "run", 1),
    ("src/ops/home_retire_apply.rs", "run", 1),
    ("src/ops/home_retire_plan.rs", "run", 1),
    ("src/ops/home_retire_revert.rs", "run", 1),
    ("src/ops/home_state_apply.rs", "run", 1),
    ("src/ops/home_state_plan.rs", "run", 1),
    ("src/ops/home_state_revert.rs", "run", 1),
    ("src/ops/host_wake.rs", "run", 1),
    ("src/ops/identity.rs", "run", 1),
    ("src/ops/install_inspect.rs", "run", 1),
    ("src/ops/lane_next.rs", "run", 1),
    ("src/ops/star_ask.rs", "run", 1),
    ("src/statusline/mod.rs", "run", 1),
];

/// Direct reads below the command edge.
const INNER_READS: &[Read] = &[
    ("src/agents.rs", "default_global_excludes", 1),
    ("src/agents.rs", "tilde", 1),
    ("src/claude.rs", "doctor", 1),
    ("src/claude.rs", "refuse_user_scope", 1),
    ("src/edit_diagnostics.rs", "lock_dir", 1),
    ("src/heavy_slot.rs", "cache_dir", 1),
    ("src/home_workdir.rs", "home", 1),
    ("src/hook_config.rs", "load", 1),
    ("src/hook_engine.rs", "bypass_boundary_deny", 1),
    ("src/hook_engine.rs", "session_check", 1),
    ("src/hooks.rs", "env_home", 1),
    ("src/key_vault.rs", "home", 1),
    ("src/layout.rs", "environment_home", 1),
    ("src/migrate.rs", "open_cli_home", 1),
    ("src/project_update.rs", "go_settings_step", 1),
    ("src/project_update.rs", "settings_step", 1),
    ("src/update_tux.rs", "view_for", 1),
];

/// Adapter calls outside a command edge.
const ADAPTER_CALLS: &[Call] = &[
    ("claim::journal_root", "src/claim.rs", "acquire_cli", 1),
    ("claim::journal_root", "src/claim.rs", "release_cli", 1),
    ("claim::journal_root", "src/claim.rs", "show_cli", 1),
    ("claim::journal_root", "src/handoff.rs", "open_journal", 1),
    ("claim::journal_root", "src/ledger_recover.rs", "gather", 1),
    (
        "claim::journal_root",
        "src/worktree.rs",
        "active_records",
        1,
    ),
    ("claim::journal_root", "src/worktree.rs", "load_record", 1),
    (
        "home_workdir::home",
        "src/project_update.rs",
        "reports_step",
        1,
    ),
    ("home_workdir::home", "src/worktree_gc.rs", "work_roots", 1),
    (
        "home_workdir::home",
        "src/worktree_move.rs",
        "default_root",
        1,
    ),
    ("home_workdir::home", "src/worktree_move.rs", "old_root", 1),
    ("home_workdir::home", "src/worktree_move.rs", "plan", 1),
    (
        "home_workdir::home",
        "src/worktree_move.rs",
        "refuse_linked_move_root",
        1,
    ),
    (
        "hooks::env_home",
        "src/hook_engine.rs",
        "ac_verify_recorded",
        1,
    ),
    (
        "hooks::env_home",
        "src/hook_engine.rs",
        "consecutive_blocks",
        1,
    ),
    ("hooks::env_home", "src/hooks.rs", "append_receipt_line", 1),
    ("hooks::env_home", "src/hooks.rs", "receipt_lines", 1),
    ("key_vault::home", "src/host_home.rs", "resolve", 1),
    (
        "key_vault::home",
        "src/launcher_init.rs",
        "lock_file_is_ours",
        1,
    ),
    (
        "key_vault::home",
        "src/launcher_init.rs",
        "missing_skill",
        1,
    ),
    (
        "key_vault::home",
        "src/launcher_init.rs",
        "prepare_project",
        1,
    ),
    ("key_vault::home", "src/ops/key_set.rs", "prompt", 1),
    ("key_vault::home", "src/session.rs", "glm_token", 1),
    ("key_vault::home", "src/session.rs", "user_launcher", 1),
    ("layout::default_home_root", "src/layout.rs", "apply_cli", 1),
    ("layout::default_home_root", "src/layout.rs", "plan_cli", 1),
    (
        "layout::environment_home",
        "src/apply.rs",
        "identity_fields",
        1,
    ),
    ("layout::environment_home", "src/audit.rs", "observe", 1),
    (
        "layout::environment_home",
        "src/claim.rs",
        "journal_root",
        1,
    ),
    (
        "layout::environment_home",
        "src/key_vault.rs",
        "scope_here",
        1,
    ),
    (
        "layout::environment_home",
        "src/launcher_init.rs",
        "prepare_project",
        1,
    ),
    (
        "layout::environment_home",
        "src/layout.rs",
        "default_home_root",
        1,
    ),
    (
        "layout::environment_home",
        "src/layout.rs",
        "switched_dir",
        1,
    ),
    (
        "layout::environment_home",
        "src/layout.rs",
        "switched_dir_checked",
        1,
    ),
    ("layout::environment_home", "src/session.rs", "glm_token", 1),
    (
        "layout::environment_home",
        "src/worktree.rs",
        "open_locked_journal",
        1,
    ),
    (
        "layout::switched_dir_checked",
        "src/apply.rs",
        "receipts_dir",
        1,
    ),
    (
        "layout::switched_dir_checked",
        "src/forge/outbox.rs",
        "outbox_dir",
        1,
    ),
    (
        "worktree::open_locked_journal",
        "src/claim_link.rs",
        "link_cli",
        1,
    ),
    (
        "worktree::open_locked_journal",
        "src/pr_link.rs",
        "confirm_and_record",
        1,
    ),
    (
        "worktree::open_locked_journal",
        "src/pr_link.rs",
        "render_and_record",
        1,
    ),
    (
        "worktree::open_locked_journal",
        "src/scope.rs",
        "check_and_record",
        1,
    ),
    (
        "worktree::open_locked_journal",
        "src/worktree.rs",
        "apply_and_record",
        1,
    ),
];

/// Functions that read HOME for their caller: the name in reports, the module that must qualify
/// the call (empty: any spelling) and the function name.
const ADAPTERS: [(&str, &str, &str); 9] = [
    ("key_vault::home", "key_vault", "home"),
    ("home_workdir::home", "home_workdir", "home"),
    ("hooks::env_home", "", "env_home"),
    ("layout::environment_home", "", "environment_home"),
    ("layout::default_home_root", "", "default_home_root"),
    ("layout::switched_dir", "", "switched_dir"),
    ("layout::switched_dir_checked", "", "switched_dir_checked"),
    ("claim::journal_root", "", "journal_root"),
    ("worktree::open_locked_journal", "", "open_locked_journal"),
];

/// Command edges outside `src/main.rs` and the `run` functions under `src/ops/`.
const EDGE_FUNCTIONS: [(&str, &str); 3] = [
    ("src/ledger/command.rs", "resolve_home"),
    ("src/hook_entry.rs", "env"),
    ("src/statusline/mod.rs", "run"),
];

fn is_edge(file: &str, function: &str) -> bool {
    file == "src/main.rs"
        || (file.starts_with("src/ops/") && function == "run")
        || EDGE_FUNCTIONS.contains(&(file, function))
}

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

/// One read or adapter call found in a source file; `adapter` is `None` for a direct read.
struct Site {
    adapter: Option<&'static str>,
    file: String,
    function: String,
}

fn sites(file: &str, source: &str) -> Vec<Site> {
    let list = tokens(source);
    let word = |i: usize| match list.get(i) {
        Some((Token::Word(w), _)) => w.as_str(),
        _ => "",
    };
    let mark = |i: usize, c: char| matches!(list.get(i), Some((Token::Mark(m), _)) if *m == c);
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
    let mut found = Vec::new();
    let mut braces = 0usize;
    // The open functions with the brace depth of their body.
    let mut open: Vec<(String, usize)> = Vec::new();
    // A function whose body has not opened yet, with the bracket depth inside its signature.
    let mut pending: Option<(String, i32)> = None;
    let mut i = 0;
    while i < list.len() {
        // A test-only `cfg` attribute in the first column, further attributes, then an inline
        // module.
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
            open.last()
                .map_or_else(|| "<file>".to_string(), |(name, _)| name.clone())
        };
        if word(i) == "fn" && !word(i + 1).is_empty() {
            pending = Some((word(i + 1).to_string(), 0));
            i += 2;
            continue;
        }
        if word(i) == "env"
            && mark(i + 1, ':')
            && mark(i + 2, ':')
            && matches!(word(i + 3), "var" | "var_os")
            && mark(i + 4, '(')
            && matches!(list.get(i + 5), Some((Token::Text(t), _)) if t == "HOME")
        {
            found.push(Site {
                adapter: None,
                file: file.to_string(),
                function: here(),
            });
        }
        if mark(i + 1, '(') {
            for (label, module, name) in ADAPTERS {
                let qualified = module.is_empty()
                    || (i >= 3 && word(i - 3) == module && mark(i - 2, ':') && mark(i - 1, ':'));
                if word(i) == name && qualified {
                    found.push(Site {
                        adapter: Some(label),
                        file: file.to_string(),
                        function: here(),
                    });
                }
            }
        }
        if let Some((Token::Mark(c), _)) = list.get(i) {
            let in_signature = pending.as_ref().map(|(_, depth)| *depth);
            match (*c, in_signature) {
                ('(' | '[', Some(_)) => pending.as_mut().unwrap().1 += 1,
                (')' | ']', Some(_)) => pending.as_mut().unwrap().1 -= 1,
                // A declaration without a body.
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
    found
}

/// Counted places, keyed by what was found there (a class or an adapter), file and function.
type Counts = BTreeMap<(String, String, String), usize>;

/// One entry per place of `more` that `less` does not cover.
fn surplus(more: &Counts, less: &Counts) -> Vec<(String, String, String)> {
    let mut out = Vec::new();
    for (key, count) in more {
        for _ in less.get(key).copied().unwrap_or(0)..*count {
            out.push(key.clone());
        }
    }
    out
}

fn add(counts: &mut Counts, what: &str, file: &str, function: &str, count: usize) {
    *counts
        .entry((what.to_string(), file.to_string(), function.to_string()))
        .or_insert(0) += count;
}

/// One line for every direct read and every adapter call that differs from the lists.
fn compare(
    files: &[(String, String)],
    edge: &[Read],
    inner: &[Read],
    calls: &[Call],
) -> Vec<String> {
    let (mut listed_reads, mut listed_calls) = (Counts::new(), Counts::new());
    let (mut found_reads, mut found_calls) = (Counts::new(), Counts::new());
    for (class, rows) in [("edge", edge), ("inner", inner)] {
        for (file, function, count) in rows {
            add(&mut listed_reads, class, file, function, *count);
        }
    }
    for (adapter, file, function, count) in calls {
        add(&mut listed_calls, adapter, file, function, *count);
    }
    for (file, source) in files {
        for site in sites(file, source) {
            let at_edge = is_edge(&site.file, &site.function);
            match site.adapter {
                None => {
                    let class = if at_edge { "edge" } else { "inner" };
                    add(&mut found_reads, class, &site.file, &site.function, 1);
                }
                Some(adapter) if !at_edge => {
                    add(&mut found_calls, adapter, &site.file, &site.function, 1);
                }
                Some(_) => {}
            }
        }
    }
    let mut lines = Vec::new();
    for (class, file, function) in surplus(&found_reads, &listed_reads) {
        lines.push(format!(
            "unlisted {class} HOME read in {file} fn {function}"
        ));
    }
    for (class, file, function) in surplus(&listed_reads, &found_reads) {
        lines.push(format!(
            "listed {class} HOME read not found in {file} fn {function}"
        ));
    }
    let mut gone = surplus(&listed_calls, &found_calls);
    for (adapter, file, function) in surplus(&found_calls, &listed_calls) {
        // A call that left one place and appeared in another is one moved call.
        match gone.iter().position(|(name, _, _)| *name == adapter) {
            Some(at) => {
                let (_, from_file, from_function) = gone.remove(at);
                lines.push(format!(
                    "call of {adapter} moved from {from_file} fn {from_function} to {file} fn {function}"
                ));
            }
            None => lines.push(format!(
                "unlisted call of {adapter} in {file} fn {function}"
            )),
        }
    }
    for (adapter, file, function) in gone {
        lines.push(format!(
            "listed call of {adapter} not found in {file} fn {function}"
        ));
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
fn the_checked_in_sources_match_the_list() {
    let root = Path::new(env!("CARGO_MANIFEST_DIR"));
    let mut files = Vec::new();
    rust_files(root, &root.join("src"), &mut files);
    let lines = compare(&files, EDGE_READS, INNER_READS, ADAPTER_CALLS);
    let mode = std::env::var(MODE_VARIABLE).ok();
    // Written to the process's stderr, not through `eprintln!`, so a passing run shows it.
    if let Err(differences) = judge(&lines, mode.as_deref(), &mut std::io::stderr().lock()) {
        panic!("HOME readers differ from the list in tests/home_env_readers.rs:\n{differences}");
    }
}

fn planted(file: &str, source: &str) -> Vec<(String, String)> {
    vec![(file.to_string(), source.to_string())]
}

/// AC-17 (REQ-13): equal sources report nothing, an added read names its file and function, a
/// changed class is reported, and comments, strings and test modules are not counted.
#[test]
fn an_added_read_is_reported_with_its_file_and_function() {
    let source = r##"
//! env::var("HOME") in a module comment
/* env::var_os("HOME") /* nested */ env::var("HOME") */
pub fn listed<'a>(x: &'a [u8; 2]) -> Option<String> {
    let quote = '"';
    let raw = r#"env::var("HOME")"#;
    let text = "env::var(\"HOME\")";
    let other = std::env::var("CLAUDE_CONFIG_DIR");
    std::env::var("HOME").ok() // env::var("HOME")
}
trait T {
    fn declared(&self) -> [u8; 4];
}
#[cfg(test)]
#[allow(dead_code)]
mod tests {
    fn helper() {
        std::env::var_os("HOME");
    }
}
#[cfg(all(test, unix))]
mod unix_tests {
    fn helper() {
        std::env::var("HOME");
        default_home_root();
    }
}
"##;
    let inner: &[Read] = &[("src/demo.rs", "listed", 1)];
    let equal = compare(&planted("src/demo.rs", source), &[], inner, &[]);
    assert!(equal.is_empty(), "{equal:?}");

    let more = format!(
        "{source}\nfn added() {{\n    let a = || std::env::var_os(\n        \"HOME\",\n    );\n}}\n"
    );
    let lines = compare(&planted("src/demo.rs", &more), &[], inner, &[]);
    assert_eq!(lines, ["unlisted inner HOME read in src/demo.rs fn added"]);

    // A module that is also built outside tests is product source.
    let built = format!(
        "{source}#[cfg(all(not(test), unix))]\nmod real {{ fn built() {{ std::env::var(\"HOME\"); }} }}\n"
    );
    let lines = compare(&planted("src/demo.rs", &built), &[], inner, &[]);
    assert_eq!(lines, ["unlisted inner HOME read in src/demo.rs fn built"]);

    let wrong_class = compare(&planted("src/demo.rs", source), inner, &[], &[]);
    assert_eq!(
        wrong_class,
        [
            "unlisted inner HOME read in src/demo.rs fn listed",
            "listed edge HOME read not found in src/demo.rs fn listed",
        ]
    );

    let edge = "pub fn run() { std::env::var_os(\"HOME\"); }\n";
    let listed: &[Read] = &[("src/ops/demo.rs", "run", 1)];
    let lines = compare(&planted("src/ops/demo.rs", edge), listed, &[], &[]);
    assert!(lines.is_empty(), "{lines:?}");
}

/// AC-18 (REQ-14): an added, removed or moved adapter call is reported; definitions, the `_in`
/// partners, other modules' `home` and calls at a command edge are not.
#[test]
fn a_changed_adapter_call_is_reported() {
    let source = r#"
pub fn journal_root(p: &str) {}
fn first() {
    claim::journal_root("p");
    journal_root_in("p", None);
    layout::switched_dir_in("p", "claims", None);
    key_vault::home();
    other::home();
}
fn second() {}
"#;
    let calls: &[Call] = &[
        ("claim::journal_root", "src/demo.rs", "first", 1),
        ("key_vault::home", "src/demo.rs", "first", 1),
    ];
    let run = |text: &str| compare(&planted("src/demo.rs", text), &[], &[], calls);
    assert!(run(source).is_empty(), "{:?}", run(source));

    let added = source.replace("fn second() {}", "fn second() { switched_dir(\"p\"); }");
    assert_eq!(
        run(&added),
        ["unlisted call of layout::switched_dir in src/demo.rs fn second"]
    );

    let removed = source.replace("    key_vault::home();\n", "");
    assert_eq!(
        run(&removed),
        ["listed call of key_vault::home not found in src/demo.rs fn first"]
    );

    let moved = removed.replace("fn second() {}", "fn second() { key_vault::home(); }");
    assert_eq!(
        run(&moved),
        ["call of key_vault::home moved from src/demo.rs fn first to src/demo.rs fn second"]
    );

    let at_edge = "fn main() { key_vault::home(); open_locked_journal(p); }\n";
    let lines = compare(&planted("src/main.rs", at_edge), &[], &[], &[]);
    assert!(lines.is_empty(), "{lines:?}");
}

/// A planted difference: one read the empty lists do not hold.
fn one_difference() -> Vec<String> {
    let source = "fn added() { std::env::var(\"HOME\"); }\n";
    compare(&planted("src/demo.rs", source), &[], &[], &[])
}

/// AC-22 (REQ-17)
#[test]
fn enforce_mode_fails_on_a_difference() {
    let lines = one_difference();
    assert_eq!(lines.len(), 1, "{lines:?}");
    let mut out = Vec::new();
    let failure = judge(&lines, Some("enforce"), &mut out).unwrap_err();
    assert!(failure.contains(&lines[0]), "{failure}");
    assert_eq!(judge(&[], Some("enforce"), &mut out), Ok(()));
}

/// AC-23 (REQ-18)
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
