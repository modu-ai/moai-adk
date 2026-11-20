//! `check affected` (SPEC-20261120-check-affected): from the paths a range changes, pick the
//! tests the change can break and the steps that check it. [`select`] and [`steps`] are pure
//! (changed paths, test file facts, the scan list text and the host name in; a selection and
//! a step list out); [`read_facts`] is the one place that reads git and the files.

use std::collections::BTreeSet;
use std::path::Path;
use std::process::Command;

type Result<T> = std::result::Result<T, String>;

/// One file of `tests/` as the caller read it: a Rust test target or a Python test module.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct TestFile {
    /// The target name (`tests/<name>.rs`) or the module name (`tests/<name>.py`).
    pub name: String,
    pub python: bool,
    /// `mod <x>;` lines of a Rust target: the shared helper folders it declares.
    pub modules: Vec<String>,
    /// The operating systems named in `cfg(target_os = "<os>")`.
    pub os: Vec<String>,
}

/// What the scan list gave: no file, a readable one (with how many targets it named), or one
/// that could not be read (every test is then selected).
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Scans {
    None,
    Listed(usize),
    Unreadable,
}

impl Scans {
    pub fn word(&self) -> String {
        match self {
            Scans::None => "none".to_string(),
            Scans::Listed(n) => format!("listed:{n}"),
            Scans::Unreadable => "unreadable".to_string(),
        }
    }
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Selection {
    pub rust: BTreeSet<String>,
    pub python: BTreeSet<String>,
    pub scans: Scans,
}

/// The first word of an area name: `update_models` → `update`.
fn first_word(area: &str) -> &str {
    area.split('_').next().unwrap_or(area)
}

fn stem(file: &str) -> &str {
    file.rsplit_once('.').map_or(file, |(s, _)| s)
}

/// The area of a changed source path: `src/<x>.rs` → `x`, `src/<dir>/…` → `dir`, and
/// `src/ops/<y>.rs` → `y` (one operation is one area, not all of `ops`).
pub fn area(path: &str) -> Option<String> {
    let rest = path.strip_prefix("src/")?;
    let parts: Vec<&str> = rest.split('/').collect();
    let name = match parts.as_slice() {
        ["ops", file] => stem(file),
        [file] => stem(file),
        [dir, ..] => dir,
        [] => return None,
    };
    Some(name.replace('-', "_"))
}

fn same_first_word(test: &str, word: &str) -> bool {
    test == word || test.strip_prefix(word).is_some_and(|r| r.starts_with('_'))
}

/// The targets and modules a scan list (`tools/repo_scans.json`) names: `--test <name>` of a
/// `cargo test` check and each `tests.<name>` of a `unittest` check. `None` for text that is
/// not a scan list.
fn scan_targets(text: &str) -> Option<(BTreeSet<String>, BTreeSet<String>)> {
    let value: serde_json::Value = serde_json::from_str(text).ok()?;
    let scans = value.get("scans")?.as_array()?;
    let mut rust = BTreeSet::new();
    let mut python = BTreeSet::new();
    for scan in scans {
        for check in scan.get("checks")?.as_array()? {
            let words: Vec<&str> = check.as_str()?.split_whitespace().collect();
            for (i, w) in words.iter().enumerate() {
                if *w == "--test" {
                    if let Some(n) = words.get(i + 1) {
                        rust.insert(n.to_string());
                    }
                }
                if let Some(m) = w.strip_prefix("tests.") {
                    python.insert(m.to_string());
                }
            }
        }
    }
    Some((rust, python))
}

/// Pick the Rust targets and Python modules a change can break (REQ-1 to REQ-6).
pub fn select(
    changed: &[String],
    tests: &[TestFile],
    scan_list: Option<&str>,
    host_os: &str,
) -> Selection {
    let mut rust = BTreeSet::new();
    let mut python = BTreeSet::new();
    let mut scans = Scans::None;
    let rust_names: BTreeSet<&str> = tests
        .iter()
        .filter(|t| !t.python)
        .map(|t| t.name.as_str())
        .collect();
    let python_names: BTreeSet<&str> = tests
        .iter()
        .filter(|t| t.python)
        .map(|t| t.name.as_str())
        .collect();
    for path in changed {
        if let Some(a) = area(path) {
            let word = first_word(&a);
            for t in tests.iter().filter(|t| !t.python) {
                if same_first_word(&t.name, word) {
                    rust.insert(t.name.clone());
                }
            }
            for t in tests.iter().filter(|t| t.python) {
                if t.name.ends_with("_test") && same_first_word(&t.name, word) {
                    python.insert(t.name.clone());
                }
            }
            continue;
        }
        let Some(rest) = path.strip_prefix("tests/") else {
            continue;
        };
        let parts: Vec<&str> = rest.split('/').collect();
        match parts.as_slice() {
            [file] => {
                let name = stem(file);
                if file.ends_with(".rs") && rust_names.contains(name) {
                    rust.insert(name.to_string());
                }
                if file.ends_with(".py") && python_names.contains(name) {
                    python.insert(name.to_string());
                }
                if file.ends_with(".rs") {
                    for t in tests.iter().filter(|t| t.modules.iter().any(|m| m == name)) {
                        rust.insert(t.name.clone());
                    }
                }
            }
            ["fixtures", dir, ..] => {
                let dir = dir.replace('-', "_");
                for t in tests.iter() {
                    if same_first_word(&t.name, &dir) {
                        if t.python {
                            python.insert(t.name.clone());
                        } else {
                            rust.insert(t.name.clone());
                        }
                    }
                }
            }
            [dir, ..] => {
                for t in tests.iter().filter(|t| t.modules.iter().any(|m| m == dir)) {
                    rust.insert(t.name.clone());
                }
            }
            [] => {}
        }
    }
    match scan_list {
        None => {}
        Some(text) => match scan_targets(text) {
            Some((r, p)) => {
                scans = Scans::Listed(r.len() + p.len());
                rust.extend(r);
                python.extend(p);
            }
            None => {
                scans = Scans::Unreadable;
                rust.extend(rust_names.iter().map(|s| s.to_string()));
                python.extend(python_names.iter().map(|s| s.to_string()));
            }
        },
    }
    for t in tests.iter().filter(|t| !t.python) {
        if t.os.iter().any(|o| o == host_os) {
            rust.insert(t.name.clone());
        }
    }
    Selection {
        rust,
        python,
        scans,
    }
}

/// The helper folders a Rust test file declares: `mod <x>;` lines, also with attributes before
/// them, a `pub` or `pub(crate)` and a comment after them.
pub fn declared_mods(text: &str) -> Vec<String> {
    text.lines()
        .filter_map(|l| {
            let mut t = l.trim();
            while t.starts_with("#[") {
                t = t[t.find(']')? + 1..].trim_start();
            }
            if t.starts_with("pub(") {
                t = t[t.find(')')? + 1..].trim_start();
            }
            let t = t.strip_prefix("pub ").unwrap_or(t);
            let rest = t.strip_prefix("mod ")?;
            let (name, after) = rest.split_once(';')?;
            let after = after.trim();
            let name = name.trim();
            let plain = !name.is_empty() && name.chars().all(|c| c.is_alphanumeric() || c == '_');
            (plain && (after.is_empty() || after.starts_with("//"))).then(|| name.to_string())
        })
        .collect()
}

/// The operating systems a file names in `target_os = "<os>"`.
pub fn os_names(text: &str) -> Vec<String> {
    let mut found = Vec::new();
    let mut rest = text;
    while let Some(i) = rest.find("target_os") {
        rest = &rest[i + "target_os".len()..];
        let t = rest.trim_start();
        let Some(t) = t.strip_prefix('=') else {
            continue;
        };
        let Some(t) = t.trim_start().strip_prefix('"') else {
            continue;
        };
        if let Some(end) = t.find('"') {
            let name = &t[..end];
            if !found.iter().any(|f| f == name) {
                found.push(name.to_string());
            }
        }
    }
    found
}

/// The host's operating system in `cfg(target_os)` spelling.
pub fn host_os() -> &'static str {
    std::env::consts::OS
}

/// One planned step. `heavy` steps run under one heavy slot hold each.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Step {
    pub label: String,
    pub program: String,
    pub args: Vec<String>,
    pub heavy: bool,
    pub env: Vec<(String, String)>,
}

fn step(label: &str, program: &str, args: &[&str], heavy: bool) -> Step {
    Step {
        label: label.to_string(),
        program: program.to_string(),
        args: args.iter().map(|s| s.to_string()).collect(),
        heavy,
        env: Vec::new(),
    }
}

/// The steps of a selection, in the order the summary names them (REQ-4, REQ-7). `present` are
/// the changed paths that exist in the checkout.
pub fn steps(sel: &Selection, present: &[String], range: &str, target_dir: &str) -> Vec<Step> {
    let mut out = vec![
        step("diff_check", "git", &["diff", "--check", range], false),
        step("fmt", "cargo", &["fmt", "--all", "--", "--check"], false),
    ];
    let ops: Vec<&String> = present
        .iter()
        .filter(|p| p.starts_with("src/ops/") && p.ends_with(".rs"))
        .collect();
    if !ops.is_empty() {
        let mut s = step(
            "ops_rustfmt",
            "rustfmt",
            &["--edition", "2021", "--check"],
            false,
        );
        s.args.extend(ops.into_iter().cloned());
        out.push(s);
    }
    for p in present.iter().filter(|p| p.ends_with(".sh")) {
        out.push(step(
            &format!("bash_n:{p}"),
            "bash",
            &["-n", "--", p],
            false,
        ));
    }
    out.push(cargo_step(
        "clippy",
        &["clippy"],
        &["--all-targets", "--", "-D", "warnings"],
    ));
    out.push(cargo_step(
        "test_lib",
        &["test"],
        &["--no-fail-fast", "--lib"],
    ));
    out.push(cargo_step(
        "test_bins",
        &["test"],
        &["--no-fail-fast", "--bins"],
    ));
    if !sel.rust.is_empty() {
        let mut rest = vec!["--no-fail-fast".to_string()];
        for t in &sel.rust {
            rest.push("--test".to_string());
            rest.push(t.clone());
        }
        let refs: Vec<&str> = rest.iter().map(String::as_str).collect();
        out.push(cargo_step("test_targets", &["test"], &refs));
    }
    if !sel.python.is_empty() {
        out.push(cargo_step("build_cli", &["build"], &["--bin", "moai-cli"]));
        let mut py = step("python", "python3", &["-m", "unittest"], false);
        py.args
            .extend(sel.python.iter().map(|m| format!("tests.{m}")));
        py.env
            .push(("CARGO_TARGET_DIR".to_string(), target_dir.to_string()));
        py.env.push((
            "MOAI_CLI_BIN".to_string(),
            format!("{target_dir}/debug/moai-cli"),
        ));
        out.push(py);
    }
    out
}

/// A cargo step: `cargo <sub> --locked --offline <rest>`.
fn cargo_step(label: &str, sub: &[&str], rest: &[&str]) -> Step {
    let mut args: Vec<String> = sub.iter().map(|s| s.to_string()).collect();
    args.extend(["--locked".to_string(), "--offline".to_string()]);
    args.extend(rest.iter().map(|s| s.to_string()));
    Step {
        label: label.to_string(),
        program: "cargo".to_string(),
        args,
        heavy: true,
        env: Vec::new(),
    }
}

/// What a range gives: the commits, the changed paths and the files of `tests/`.
pub struct Facts {
    /// The repository's top folder (`--dir` may be a folder inside it).
    pub root: std::path::PathBuf,
    pub base: String,
    pub head: String,
    pub changed: Vec<String>,
    pub present: Vec<String>,
    pub tests: Vec<TestFile>,
    pub scan_list: Option<String>,
}

fn git(dir: &Path, args: &[&str]) -> Result<String> {
    let mut cmd = Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let o = cmd
        .args(args)
        .current_dir(dir)
        .output()
        .map_err(|e| format!("cannot run git: {e}"))?;
    if !o.status.success() {
        return Err(format!(
            "git {} failed: {}",
            args.first().copied().unwrap_or(""),
            String::from_utf8_lossy(&o.stderr).trim()
        ));
    }
    Ok(String::from_utf8_lossy(&o.stdout).into_owned())
}

fn commit(dir: &Path, rev: &str) -> Result<String> {
    if rev.is_empty() || rev.starts_with('-') {
        return Err(format!("check affected refused: `{rev}` is not a revision"));
    }
    git(
        dir,
        &[
            "rev-parse",
            "--verify",
            "--quiet",
            &format!("{rev}^{{commit}}"),
        ],
    )
    .map(|s| s.trim().to_string())
    .map_err(|_| format!("check affected refused: the revision `{rev}` cannot be read"))
}

/// Split `<base>..<head>`; a three-dot range or a missing side is refused.
pub fn split_range(range: &str) -> Result<(&str, &str)> {
    match range.split_once("..") {
        Some((b, h)) if !b.is_empty() && !h.is_empty() && !h.starts_with('.') => Ok((b, h)),
        _ => Err(format!(
            "check affected refused: `{range}` is not `<base>..<head>`"
        )),
    }
}

/// Read the range and the project's test files. Reads git and the files only.
pub fn read_facts(dir: &Path, range: &str) -> Result<Facts> {
    let (b, h) = split_range(range)?;
    let base = commit(dir, b)?;
    let head = commit(dir, h)?;
    let root = std::path::PathBuf::from(git(dir, &["rev-parse", "--show-toplevel"])?.trim());
    let dir = root.as_path();
    // `-z` keeps a non-ASCII name unquoted, `--no-renames` lists the old path of a move too.
    let changed: Vec<String> = git(
        dir,
        &[
            "diff",
            "--name-only",
            "--no-renames",
            "-z",
            &format!("{base}...{head}"),
        ],
    )?
    .split('\0')
    .filter(|l| !l.is_empty())
    .map(String::from)
    .collect();
    let present = changed
        .iter()
        .filter(|p| dir.join(p).is_file())
        .cloned()
        .collect();
    let mut tests = Vec::new();
    if let Ok(entries) = std::fs::read_dir(dir.join("tests")) {
        for e in entries.flatten() {
            let path = e.path();
            if !path.is_file() {
                continue;
            }
            let Some(file) = path.file_name().and_then(|f| f.to_str()) else {
                continue;
            };
            let name = stem(file).to_string();
            if file.ends_with(".rs") {
                let text = std::fs::read_to_string(&path).unwrap_or_default();
                tests.push(TestFile {
                    name,
                    python: false,
                    modules: declared_mods(&text),
                    os: os_names(&text),
                });
            } else if file.ends_with(".py") && name.ends_with("_test") {
                tests.push(TestFile {
                    name,
                    python: true,
                    ..TestFile::default()
                });
            }
        }
    }
    tests.sort_by(|a, b| (a.python, &a.name).cmp(&(b.python, &b.name)));
    let scan_list = std::fs::read_to_string(dir.join("tools/repo_scans.json")).ok();
    Ok(Facts {
        root,
        base,
        head,
        changed,
        present,
        tests,
        scan_list,
    })
}

/// The folder cargo builds into: the caller's `CARGO_TARGET_DIR` (a relative one counts from the
/// repository), else `<root>/target`. The Python step and the step logs use the same folder as
/// the build step, so the binary the tests look for is the one that was built.
fn target_dir(root: &Path) -> std::path::PathBuf {
    match std::env::var_os("CARGO_TARGET_DIR") {
        Some(t) if !t.is_empty() => root.join(t),
        _ => root.join("target"),
    }
}

fn short(sha: &str) -> &str {
    sha.get(..12).unwrap_or(sha)
}

/// `check affected <range> --list`: the selection and the steps, nothing run or written.
pub fn list(dir: &Path, range: &str) -> Result<serde_json::Value> {
    let facts = read_facts(dir, range)?;
    let sel = select(
        &facts.changed,
        &facts.tests,
        facts.scan_list.as_deref(),
        host_os(),
    );
    let target = target_dir(&facts.root).to_string_lossy().into_owned();
    let planned = steps(
        &sel,
        &facts.present,
        &format!("{}...{}", facts.base, facts.head),
        &target,
    );
    Ok(serde_json::json!({
        "action": "affected",
        "mode": "list",
        "base": facts.base,
        "head": facts.head,
        "changed": facts.changed.len(),
        "host_os": host_os(),
        "scans": sel.scans.word(),
        "rust_targets": sel.rust,
        "python_modules": sel.python,
        "steps": planned.iter().map(|s| {
            let mut words = vec![s.program.clone()];
            words.extend(s.args.iter().cloned());
            serde_json::json!({"label": s.label, "heavy": s.heavy, "command": words})
        }).collect::<Vec<_>>(),
        "summary_prefix": format!(
            "check-affected base={} head={} targets={} py={} scans={}",
            short(&facts.base), short(&facts.head), sel.rust.len(), sel.python.len(), sel.scans.word()
        ),
        "writes": false,
    }))
}

/// How `run` holds the heavy slot of a cargo step.
pub struct RunOptions {
    pub slot_dir: Option<std::path::PathBuf>,
    pub max: u32,
    pub min_free_gb: u32,
    /// Where each step's log goes (default: `<dir>/target/check-affected-<head12>`).
    pub out: Option<std::path::PathBuf>,
}

/// The exit status of a finished process: its code, or 128 + the signal.
fn exit_of(status: std::process::ExitStatus) -> i32 {
    if let Some(c) = status.code() {
        return c;
    }
    #[cfg(unix)]
    {
        use std::os::unix::process::ExitStatusExt;
        if let Some(sig) = status.signal() {
            return 128 + sig;
        }
    }
    1
}

/// Run one step in `dir`; its output goes to `log`. A program that cannot start is 127.
fn run_step(dir: &Path, s: &Step, opts: &RunOptions, log: &Path) -> i32 {
    let Ok(file) = std::fs::File::create(log) else {
        return 127;
    };
    let Ok(err) = file.try_clone() else {
        return 127;
    };
    let mut cmd = if s.heavy {
        let Ok(exe) = std::env::current_exe() else {
            return 127;
        };
        let mut c = Command::new(exe);
        c.arg("heavy-slot-acquire")
            .args(["--max", &opts.max.to_string()])
            .args(["--min-free-gb", &opts.min_free_gb.to_string()]);
        if let Some(d) = &opts.slot_dir {
            c.arg("--dir").arg(d);
        }
        c.args(["--name", "check-affected", "--", &s.program])
            .args(&s.args);
        c
    } else {
        let mut c = Command::new(&s.program);
        if s.program == "git" {
            crate::identity::scrub_git_env(&mut c);
        }
        c.args(&s.args);
        c
    };
    cmd.current_dir(dir)
        .envs(s.env.iter().map(|(k, v)| (k, v)))
        .stdin(std::process::Stdio::null())
        .stdout(file)
        .stderr(err);
    match cmd.status() {
        Ok(status) => exit_of(status),
        Err(_) => 127,
    }
}

/// `key=exit` words in step order: one key per label up to its `:`, the first non-zero exit
/// of a repeated label (REQ-11).
fn summary_words(results: &[(String, i32)]) -> Vec<(String, i32)> {
    let mut words: Vec<(String, i32)> = Vec::new();
    for (label, code) in results {
        let key = label.split(':').next().unwrap_or(label).to_string();
        match words.iter_mut().find(|(k, _)| *k == key) {
            Some((_, c)) => {
                if *c == 0 {
                    *c = *code;
                }
            }
            None => words.push((key, *code)),
        }
    }
    words
}

/// `check affected <range>`: refuse unless the checkout is the range's head and clean, then run
/// every step (never stopping at the first failure) and give one summary line (REQ-8 to REQ-11).
pub fn run(dir: &Path, range: &str, opts: &RunOptions) -> Result<serde_json::Value> {
    let facts = read_facts(dir, range)?;
    let dir = facts.root.as_path();
    let at = git(dir, &["rev-parse", "HEAD"])?.trim().to_string();
    if at != facts.head {
        return Err(format!(
            "check affected refused: the checkout is at {} but the range ends at {}; check out the head first",
            short(&at),
            short(&facts.head)
        ));
    }
    let dirty = git(dir, &["status", "--porcelain", "--untracked-files=no"])?;
    if !dirty.trim().is_empty() {
        return Err(
            "check affected refused: a tracked file is modified; commit or restore it first"
                .to_string(),
        );
    }
    let sel = select(
        &facts.changed,
        &facts.tests,
        facts.scan_list.as_deref(),
        host_os(),
    );
    let target = target_dir(&std::fs::canonicalize(dir).unwrap_or_else(|_| dir.to_path_buf()));
    let planned = steps(
        &sel,
        &facts.present,
        &format!("{}...{}", facts.base, facts.head),
        &target.to_string_lossy(),
    );
    let out = opts
        .out
        .clone()
        .unwrap_or_else(|| target.join(format!("check-affected-{}", short(&facts.head))));
    std::fs::create_dir_all(&out).map_err(|e| {
        format!(
            "check affected refused: cannot create {}: {e}",
            out.display()
        )
    })?;
    let mut results: Vec<(String, i32)> = Vec::new();
    let mut records = Vec::new();
    for (i, s) in planned.iter().enumerate() {
        let log = out.join(format!(
            "{:02}-{}.log",
            i + 1,
            s.label.replace([':', '/'], "_")
        ));
        let code = run_step(dir, s, opts, &log);
        results.push((s.label.clone(), code));
        records.push(serde_json::json!({
            "label": s.label, "exit": code, "log": log.to_string_lossy(),
        }));
    }
    let pass = results.iter().all(|(_, c)| *c == 0);
    let mut line = format!(
        "check-affected base={} head={} targets={} py={} scans={}",
        short(&facts.base),
        short(&facts.head),
        sel.rust.len(),
        sel.python.len(),
        sel.scans.word()
    );
    for (k, c) in summary_words(&results) {
        line.push_str(&format!(" {k}={c}"));
    }
    line.push_str(if pass { " result=PASS" } else { " result=FAIL" });
    Ok(serde_json::json!({
        "action": "affected",
        "mode": "run",
        "summary": line,
        "result": if pass { "PASS" } else { "FAIL" },
        "steps": records,
        "out": out.to_string_lossy(),
        "writes": true,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn a_repeated_label_keeps_its_first_failure() {
        let r = vec![
            ("bash_n:a.sh".to_string(), 0),
            ("bash_n:b.sh".to_string(), 2),
            ("bash_n:c.sh".to_string(), 3),
            ("fmt".to_string(), 0),
        ];
        assert_eq!(
            summary_words(&r),
            [("bash_n".to_string(), 2), ("fmt".to_string(), 0)]
        );
    }

    fn rust(name: &str, modules: &[&str], os: &[&str]) -> TestFile {
        TestFile {
            name: name.to_string(),
            python: false,
            modules: modules.iter().map(|s| s.to_string()).collect(),
            os: os.iter().map(|s| s.to_string()).collect(),
        }
    }

    fn py(name: &str) -> TestFile {
        TestFile {
            name: name.to_string(),
            python: true,
            ..TestFile::default()
        }
    }

    fn paths(p: &[&str]) -> Vec<String> {
        p.iter().map(|s| s.to_string()).collect()
    }

    const SCANS: &str = r#"{"about":"x","scans":[
        {"id":"a","checks":["cargo test --offline --locked --test home_isolation_scan",
                            "python3 -m unittest tests.repo_scans_test"]},
        {"id":"b","checks":["cargo test --offline --locked --lib state_dir"]}]}"#;

    #[test]
    fn an_area_is_the_file_the_folder_or_the_operation() {
        assert_eq!(
            area("src/update_models.rs").as_deref(),
            Some("update_models")
        );
        assert_eq!(area("src/card/inbox.rs").as_deref(), Some("card"));
        assert_eq!(
            area("src/ops/reports_move_plan.rs").as_deref(),
            Some("reports_move_plan")
        );
        assert_eq!(area("src/ops/check-tree.rs").as_deref(), Some("check_tree"));
        assert_eq!(area("tests/a.rs"), None);
        assert_eq!(area("Cargo.toml"), None);
    }

    #[test]
    fn the_first_word_matches_whole_words_only() {
        assert!(same_first_word("update", "update"));
        assert!(same_first_word("update_cli", "update"));
        assert!(!same_first_word("updates", "update"));
        assert!(!same_first_word("go", "gogo"));
    }

    #[test]
    fn the_misses_of_the_integration_day_are_selected() {
        let tests = [
            rust("update_command_cli", &[], &[]),
            rust("update_tux_golden", &[], &[]),
            rust("bundle_body_cli", &[], &[]),
            rust("home_isolation_scan", &[], &[]),
            py("update_tux_tty_test"),
            py("bundle_ops_test"),
            py("repo_scans_test"),
        ];
        let sel = select(
            &paths(&["src/update_models.rs"]),
            &tests,
            Some(SCANS),
            "macos",
        );
        assert!(sel.rust.contains("update_command_cli"));
        assert!(sel.rust.contains("update_tux_golden"));
        assert!(sel.rust.contains("home_isolation_scan"));
        assert!(!sel.rust.contains("bundle_body_cli"));
        assert!(sel.python.contains("update_tux_tty_test"));
        assert!(sel.python.contains("repo_scans_test"));
        assert!(!sel.python.contains("bundle_ops_test"));
        assert_eq!(sel.scans, Scans::Listed(2));
    }

    #[test]
    fn a_changed_test_a_fixture_and_a_helper_select_their_targets() {
        let tests = [
            rust("help_layout", &[], &[]),
            rust("card_cli", &["common"], &[]),
            rust("bundle_cli", &["common"], &[]),
            rust("solo_cli", &[], &[]),
            py("help_golden_test"),
        ];
        let by = |p: &[&str]| select(&paths(p), &tests, None, "linux");
        assert_eq!(
            by(&["tests/solo_cli.rs"]).rust,
            ["solo_cli".to_string()].into()
        );
        assert_eq!(
            by(&["tests/fixtures/help/all.txt"]).rust,
            ["help_layout".to_string()].into()
        );
        assert_eq!(
            by(&["tests/fixtures/help/all.txt"]).python,
            ["help_golden_test".to_string()].into()
        );
        let helper = by(&["tests/common/mod.rs"]).rust;
        assert!(helper.contains("card_cli") && helper.contains("bundle_cli"));
        assert!(!helper.contains("solo_cli"));
        assert!(by(&["README.md"]).rust.is_empty());
    }

    #[test]
    fn the_host_os_targets_are_selected_whatever_changed() {
        let tests = [
            rust("mac_only", &[], &["macos"]),
            rust("lin_only", &[], &["linux"]),
        ];
        let sel = select(&paths(&["README.md"]), &tests, None, "macos");
        assert_eq!(sel.rust, ["mac_only".to_string()].into());
    }

    #[test]
    fn a_list_that_cannot_be_read_selects_every_test() {
        let tests = [rust("a_cli", &[], &[]), py("b_test")];
        let sel = select(&paths(&["README.md"]), &tests, Some("{not json"), "linux");
        assert_eq!(sel.scans, Scans::Unreadable);
        assert!(sel.rust.contains("a_cli") && sel.python.contains("b_test"));
        let none = select(&paths(&["README.md"]), &tests, None, "linux");
        assert_eq!(none.scans, Scans::None);
        assert!(none.rust.is_empty());
    }

    #[test]
    fn modules_and_operating_systems_are_read_from_the_text() {
        assert_eq!(
            declared_mods("mod common;\n  pub mod fx;\n// mod no;\nfn f() {}"),
            ["common", "fx"]
        );
        assert_eq!(
            os_names("#[cfg(target_os = \"macos\")]\n#[cfg(all(unix, target_os=\"linux\"))]\n#[cfg(target_os = \"macos\")]"),
            ["macos", "linux"]
        );
        assert!(os_names("let target_os = 1;").is_empty());
        // A comment after the line, an attribute before it and a restricted `pub` still declare it.
        assert_eq!(
            declared_mods(
                "mod a; // shared\n#[macro_use] mod b;\npub(crate) mod c;\n#[path = \"x.rs\"]\nmod d;\nmod e { }\nlet mod_x = 1;"
            ),
            ["a", "b", "c", "d"]
        );
    }

    #[test]
    fn the_steps_always_hold_lib_bins_and_the_fixed_checks() {
        let sel = select(&paths(&["README.md"]), &[], None, "linux");
        let planned = steps(&sel, &[], "a...b", "/t");
        let labels: Vec<&str> = planned.iter().map(|s| s.label.as_str()).collect();
        assert_eq!(
            labels,
            ["diff_check", "fmt", "clippy", "test_lib", "test_bins"]
        );
        assert!(planned
            .iter()
            .filter(|s| s.program == "cargo" && s.label != "fmt")
            .all(|s| s.heavy));
    }

    #[test]
    fn changed_operation_files_and_scripts_get_their_own_checks() {
        let sel = select(
            &paths(&["src/ops/x.rs", "tools/a.sh"]),
            &[py("x_test")],
            None,
            "linux",
        );
        let present = paths(&["src/ops/x.rs", "tools/a.sh"]);
        let planned = steps(&sel, &present, "a...b", "/t");
        let ops = planned.iter().find(|s| s.label == "ops_rustfmt").unwrap();
        assert_eq!(ops.args, ["--edition", "2021", "--check", "src/ops/x.rs"]);
        assert!(planned.iter().any(|s| s.label == "bash_n:tools/a.sh"));
        let python = planned.iter().find(|s| s.label == "python").unwrap();
        assert_eq!(python.args, ["-m", "unittest", "tests.x_test"]);
        assert!(planned.iter().any(|s| s.label == "build_cli"));
    }

    #[test]
    fn a_range_needs_both_sides_and_two_dots() {
        assert_eq!(split_range("a..b").unwrap(), ("a", "b"));
        assert!(split_range("a...b").is_err());
        assert!(split_range("a..").is_err());
        assert!(split_range("..b").is_err());
        assert!(split_range("a").is_err());
    }
}
