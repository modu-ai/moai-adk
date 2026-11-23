//! `check affected` (SPEC-20261120-check-affected): from the paths a range changes, pick the
//! tests the change can break and the steps that check it. [`select`] and [`steps`] are pure
//! (changed paths, test file facts, the scan list text and the host name in; a selection and
//! a step list out); [`read_facts`] is the one place that reads git and the files.

use std::collections::BTreeSet;
use std::path::Path;
use std::process::Command;

type Result<T> = std::result::Result<T, String>;

/// One Rust test target of `tests/` as the caller read it.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct TestFile {
    /// The target name (`tests/<name>.rs`).
    pub name: String,
    /// `mod <x>;` lines of a Rust target: the shared helper folders it declares.
    pub modules: Vec<String>,
    /// The operating systems named in `cfg(target_os = "<os>")`.
    pub os: Vec<String>,
    /// The file's source text: the lane rule for `tests/common/<sub>.rs` and the text-pin rule
    /// read it (SPEC-20261123-check-precise-selection REQ-3, REQ-6).
    pub text: String,
}

/// Which set a caller gets: the full set (today's, the bare command) or the precise lane set
/// (`--lane`), a subset of the full set for the same range (SPEC-20261123-check-precise-selection
/// REQ-1).
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq)]
pub enum Pick {
    #[default]
    Full,
    Lane,
}

impl Pick {
    /// The word of the answer's `selection` field.
    pub fn word(self) -> &'static str {
        match self {
            Pick::Full => "full",
            Pick::Lane => "lane",
        }
    }
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

/// One scan of `tools/repo_scans.json`: its cargo targets, its rule files and its scope. A scope
/// that is absent, empty or not a list of `{prefix, suffixes?}` objects is `None`.
struct ScanEntry {
    targets: BTreeSet<String>,
    /// `"always": true`: the landing requires the scan whenever the diff touches it. The full set
    /// adds its targets only then (SPEC-20261123-land-file-read-scans D-9); a scan that is not
    /// `always` joins the full set whatever changed.
    always: bool,
    rules: Vec<String>,
    scope: Option<Vec<(String, Vec<String>)>>,
}

impl ScanEntry {
    /// REQ-2: a changed path is in the scope (under a `prefix`, and with one of its `suffixes`
    /// when it names any) or is one of the rule files.
    fn touched_by(&self, changed: &[String]) -> bool {
        changed.iter().any(|p| {
            self.rules.iter().any(|r| r == p)
                || self.scope.iter().flatten().any(|(prefix, suffixes)| {
                    p.starts_with(prefix.as_str())
                        && (suffixes.is_empty() || suffixes.iter().any(|s| p.ends_with(s.as_str())))
                })
        })
    }
}

/// The scans of a scan list (`tools/repo_scans.json`). A scan's targets are the `--test <name>` of
/// its `cargo test` checks; a `unittest` check names no target: no Python module is selected
/// (SPEC-20261123-rust-scans-before-ready REQ-3). `None` for text that is not a scan list.
fn scan_entries(text: &str) -> Option<Vec<ScanEntry>> {
    let value: serde_json::Value = serde_json::from_str(text).ok()?;
    let scans = value.get("scans")?.as_array()?;
    let strings = |v: Option<&serde_json::Value>| -> Vec<String> {
        v.and_then(|v| v.as_array())
            .map(|a| {
                a.iter()
                    .filter_map(|s| s.as_str().map(String::from))
                    .collect()
            })
            .unwrap_or_default()
    };
    let mut out = Vec::new();
    for scan in scans {
        let mut targets = BTreeSet::new();
        for check in scan.get("checks")?.as_array()? {
            let words: Vec<&str> = check.as_str()?.split_whitespace().collect();
            for (i, w) in words.iter().enumerate() {
                if *w == "--test" {
                    if let Some(n) = words.get(i + 1) {
                        targets.insert(n.to_string());
                    }
                }
            }
        }
        let scope: Vec<(String, Vec<String>)> = scan
            .get("scope")
            .and_then(|s| s.as_array())
            .map(|a| {
                a.iter()
                    .filter_map(|e| {
                        let prefix = e.get("prefix")?.as_str()?.to_string();
                        Some((prefix, strings(e.get("suffixes"))))
                    })
                    .collect()
            })
            .unwrap_or_default();
        out.push(ScanEntry {
            targets,
            always: scan.get("always").and_then(|a| a.as_bool()) == Some(true),
            rules: strings(scan.get("rules")),
            scope: (!scope.is_empty()).then_some(scope),
        });
    }
    Some(out)
}

/// Whether `text` names the helper sub-module `common::<sub>`: `common::<sub>…` or `<sub>` as an
/// item of a `common::{…}` group, with whole identifiers only.
fn names_common_sub(text: &str, sub: &str) -> bool {
    let ident = |c: char| c.is_alphanumeric() || c == '_';
    let lead = |s: &str| -> String { s.trim_start().chars().take_while(|c| ident(*c)).collect() };
    let mut rest = text;
    while let Some(i) = rest.find("common::") {
        let before_ok = !rest[..i].chars().next_back().is_some_and(ident);
        let after = &rest[i + "common::".len()..];
        rest = after;
        if !before_ok {
            continue;
        }
        let Some(group) = after.strip_prefix('{') else {
            if lead(after) == sub {
                return true;
            }
            continue;
        };
        let mut depth = 0usize;
        let mut item = String::new();
        for c in group.chars() {
            match c {
                '{' => {
                    depth += 1;
                    item.push(c);
                }
                '}' if depth == 0 => break,
                '}' => {
                    depth -= 1;
                    item.push(c);
                }
                ',' if depth == 0 => {
                    if lead(&item) == sub {
                        return true;
                    }
                    item.clear();
                }
                _ => item.push(c),
            }
        }
        if lead(&item) == sub {
            return true;
        }
    }
    false
}

/// Pick the full set of Rust test targets a change can break (REQ-1 to REQ-6 of the earlier SPEC).
/// A Python test module is never picked: the Python tests that remain run in CI only.
pub fn select(
    changed: &[String],
    tests: &[TestFile],
    scan_list: Option<&str>,
    host_os: &str,
) -> Selection {
    select_pick(Pick::Full, changed, tests, scan_list, host_os)
}

/// Pick the Rust test targets of `pick` (SPEC-20261123-check-precise-selection). The lane set
/// narrows three rules of the full set: a scan joins only when a changed path is in its scope or
/// is one of its rule files (REQ-2); a change to `tests/common/<sub>.rs` selects only the
/// consumers that name `common::<sub>` (REQ-3); an OS-tagged test joins only through the area
/// rule, not for being tagged with the host (REQ-4). The text-pin rule (REQ-6) and the helper
/// path-import rule (REQ-9) add to both. An `always` scan joins even the full set only when a
/// changed path touches it (SPEC-20261123-land-file-read-scans D-9).
pub fn select_pick(
    pick: Pick,
    changed: &[String],
    tests: &[TestFile],
    scan_list: Option<&str>,
    host_os: &str,
) -> Selection {
    let mut rust = BTreeSet::new();
    let mut scans = Scans::None;
    let rust_names: BTreeSet<&str> = tests.iter().map(|t| t.name.as_str()).collect();
    for path in changed {
        if path.starts_with("skills/") || path.starts_with("docs/") {
            for t in tests.iter().filter(|t| t.text.contains(path.as_str())) {
                rust.insert(t.name.clone());
            }
        }
        if let Some(a) = area(path) {
            let word = first_word(&a);
            for t in tests {
                if same_first_word(&t.name, word) {
                    rust.insert(t.name.clone());
                }
            }
            continue;
        }
        let Some(rest) = path.strip_prefix("tests/") else {
            continue;
        };
        // REQ-9: a test that imports a helper file by path (`#[path = "common/<file>"]`), or
        // declares a module named like it, uses it without `mod common;`; both selections keep it.
        if let Some(under) = rest.strip_prefix("common/") {
            let literal = format!("common/{under}");
            let direct = (!under.contains('/') && under != "mod.rs" && under.ends_with(".rs"))
                .then(|| stem(under));
            for t in tests.iter().filter(|t| {
                t.text.contains(literal.as_str())
                    || direct.is_some_and(|s| t.modules.iter().any(|m| m == s))
            }) {
                rust.insert(t.name.clone());
            }
        }
        let parts: Vec<&str> = rest.split('/').collect();
        match parts.as_slice() {
            [file] => {
                let name = stem(file);
                if file.ends_with(".rs") && rust_names.contains(name) {
                    rust.insert(name.to_string());
                }
                if file.ends_with(".rs") {
                    for t in tests.iter().filter(|t| t.modules.iter().any(|m| m == name)) {
                        rust.insert(t.name.clone());
                    }
                }
            }
            ["fixtures", dir, ..] => {
                let dir = dir.replace('-', "_");
                for t in tests {
                    if same_first_word(&t.name, &dir) {
                        rust.insert(t.name.clone());
                    }
                }
            }
            // REQ-3: one sub-module of the helper folder reaches only the tests that name it;
            // `mod.rs` and anything deeper keep every consumer.
            ["common", file]
                if pick == Pick::Lane && *file != "mod.rs" && file.ends_with(".rs") =>
            {
                let sub = stem(file);
                for t in tests.iter().filter(|t| {
                    t.modules.iter().any(|m| m == "common") && names_common_sub(&t.text, sub)
                }) {
                    rust.insert(t.name.clone());
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
        Some(text) => match scan_entries(text) {
            Some(entries) => {
                let picked: BTreeSet<String> = entries
                    .into_iter()
                    .filter(|e| (pick == Pick::Full && !e.always) || e.touched_by(changed))
                    .flat_map(|e| e.targets)
                    .collect();
                scans = Scans::Listed(picked.len());
                rust.extend(picked);
            }
            None => {
                scans = Scans::Unreadable;
                rust.extend(rust_names.iter().map(|s| s.to_string()));
            }
        },
    }
    // REQ-4: the lane keeps an OS-tagged test only by the area rule above.
    if pick == Pick::Full {
        for t in tests {
            if t.os.iter().any(|o| o == host_os) {
                rust.insert(t.name.clone());
            }
        }
    }
    Selection { rust, scans }
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

/// What runs the tests: `cargo nextest run` when `cargo nextest --version` succeeds, else one
/// `cargo test` (SPEC-20261122-check-affected-one-slot REQ-3).
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Runner {
    Nextest,
    CargoTest,
}

impl Runner {
    pub fn word(self) -> &'static str {
        match self {
            Runner::Nextest => "nextest",
            Runner::CargoTest => "cargo-test",
        }
    }

    /// The cargo subcommand of the test step.
    fn sub(self) -> &'static [&'static str] {
        match self {
            Runner::Nextest => &["nextest", "run"],
            Runner::CargoTest => &["test"],
        }
    }
}

/// `cargo nextest --version` in `dir`: nextest when it exits 0, `cargo test` otherwise.
pub fn runner_in(dir: &Path) -> Runner {
    let ok = Command::new("cargo")
        .args(["nextest", "--version"])
        .current_dir(dir)
        .stdin(std::process::Stdio::null())
        .stdout(std::process::Stdio::null())
        .stderr(std::process::Stdio::null())
        .status()
        .is_ok_and(|s| s.success());
    if ok {
        Runner::Nextest
    } else {
        Runner::CargoTest
    }
}

/// One planned step. The `heavy` steps run one after another under one heavy slot hold.
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
/// the changed paths that exist in the checkout; `runner` runs the one `test` step. `threads` is
/// the test-thread cap of that step and of no other (SPEC-20261120-test-parallelism-cap REQ-1):
/// `-j <n>` right after `run` for nextest, `RUST_TEST_THREADS=<n>` in the step env for libtest;
/// `None` adds nothing (the caller set the knob, or nothing was asked).
pub fn steps(
    sel: &Selection,
    present: &[String],
    range: &str,
    runner: Runner,
    threads: Option<usize>,
) -> Vec<Step> {
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
    // The library, binary and selected target tests in one call (SPEC-20261122-check-affected-
    // one-slot REQ-3): one build graph and, with nextest, the binaries run side by side. The
    // binary tests are `--bin moai-cli` alone: `--bins` would also build and run the harness of
    // every other binary, even one whose `[[bin]]` says `test = false` (D-5).
    let mut rest = vec![
        "--no-fail-fast".to_string(),
        "--lib".to_string(),
        "--bin".to_string(),
        "moai-cli".to_string(),
    ];
    for t in &sel.rust {
        rest.push("--test".to_string());
        rest.push(t.clone());
    }
    let refs: Vec<&str> = rest.iter().map(String::as_str).collect();
    let mut test = cargo_step("test", runner.sub(), &refs);
    if let Some(n) = threads {
        match runner {
            Runner::Nextest => {
                // After `nextest run`, before the flags the step was built with.
                test.args.splice(2..2, ["-j".to_string(), n.to_string()]);
            }
            Runner::CargoTest => test
                .env
                .push(("RUST_TEST_THREADS".to_string(), n.to_string())),
        }
    }
    out.push(test);
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

/// What a range gives: the commits, the changed paths and the Rust test files of `tests/`.
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

/// Read the range and the project's Rust test files. Reads git and the files only.
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
                    modules: declared_mods(&text),
                    os: os_names(&text),
                    text,
                });
            }
        }
    }
    tests.sort_by(|a, b| a.name.cmp(&b.name));
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
/// repository), else `<root>/target`. The step logs go below the same folder cargo builds into.
fn target_dir(root: &Path) -> std::path::PathBuf {
    match std::env::var_os("CARGO_TARGET_DIR") {
        Some(t) if !t.is_empty() => root.join(t),
        _ => root.join("target"),
    }
}

fn short(sha: &str) -> &str {
    sha.get(..12).unwrap_or(sha)
}

/// REQ-7: when the range's base is a strict ancestor of the merge base of its head and
/// `integration_ref`, the range re-tests the integration branch's own changes. The note names that
/// merge base and how many changed paths the range has beyond the merge base's own range. It warns
/// and refuses nothing (D-4); a ref that cannot be read gives no note.
fn range_note(facts: &Facts, integration_ref: Option<&str>) -> serde_json::Value {
    let note = || -> Result<Option<serde_json::Value>> {
        let Some(integ) = integration_ref else {
            return Ok(None);
        };
        let dir = facts.root.as_path();
        let integ = commit(dir, integ)?;
        let mb = git(dir, &["merge-base", &facts.head, &integ])?
            .trim()
            .to_string();
        if mb == facts.base || git(dir, &["merge-base", "--is-ancestor", &facts.base, &mb]).is_err()
        {
            return Ok(None);
        }
        let own: BTreeSet<String> = git(
            dir,
            &[
                "diff",
                "--name-only",
                "--no-renames",
                "-z",
                &format!("{mb}...{}", facts.head),
            ],
        )?
        .split('\0')
        .filter(|l| !l.is_empty())
        .map(String::from)
        .collect();
        let extra = facts.changed.iter().filter(|p| !own.contains(*p)).count();
        Ok(Some(serde_json::json!({
            "merge_base": mb,
            "integration_ref": integration_ref,
            "extra_changed": extra,
        })))
    };
    note().ok().flatten().unwrap_or(serde_json::Value::Null)
}

/// `check affected <range> --list`: the selection and the steps, nothing run or written.
pub fn list(
    dir: &Path,
    range: &str,
    pick: Pick,
    integration_ref: Option<&str>,
) -> Result<serde_json::Value> {
    let facts = read_facts(dir, range)?;
    let sel = select_pick(
        pick,
        &facts.changed,
        &facts.tests,
        facts.scan_list.as_deref(),
        host_os(),
    );
    let runner = runner_in(&facts.root);
    // `--list` reads nothing from the slot folder, so it names no cap.
    let planned = steps(
        &sel,
        &facts.present,
        &format!("{}...{}", facts.base, facts.head),
        runner,
        None,
    );
    Ok(serde_json::json!({
        "action": "affected",
        "mode": "list",
        "selection": pick.word(),
        "range_note": range_note(&facts, integration_ref),
        "base": facts.base,
        "head": facts.head,
        "changed": facts.changed.len(),
        "host_os": host_os(),
        "runner": runner.word(),
        "scans": sel.scans.word(),
        "rust_targets": sel.rust,
        "steps": planned.iter().map(|s| {
            let mut words = vec![s.program.clone()];
            words.extend(s.args.iter().cloned());
            serde_json::json!({"label": s.label, "heavy": s.heavy, "command": words})
        }).collect::<Vec<_>>(),
        "summary_prefix": format!(
            "check-affected base={} head={} targets={} scans={}",
            short(&facts.base), short(&facts.head), sel.rust.len(), sel.scans.word()
        ),
        "writes": false,
    }))
}

/// How `run` holds the one heavy slot of its cargo steps.
pub struct RunOptions {
    pub slot_dir: Option<std::path::PathBuf>,
    pub max: u32,
    pub min_free_gb: u32,
    /// Where each step's log goes (default: `<dir>/target/check-affected-<head12>`).
    pub out: Option<std::path::PathBuf>,
    /// The full set (the bare command) or the precise lane set (`--lane`).
    pub pick: Pick,
    /// The integration branch the range note compares the base with; none gives no note.
    pub integration_ref: Option<String>,
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

/// Run one step in `dir` with `extra` added to its environment; its output goes to `log`. A
/// program that cannot start is 127.
fn run_step(dir: &Path, s: &Step, log: &Path, extra: &[(&str, std::ffi::OsString)]) -> i32 {
    let Ok(file) = std::fs::File::create(log) else {
        return 127;
    };
    let Ok(err) = file.try_clone() else {
        return 127;
    };
    let mut cmd = Command::new(&s.program);
    if s.program == "git" {
        crate::identity::scrub_git_env(&mut cmd);
    }
    cmd.args(&s.args)
        .current_dir(dir)
        .envs(s.env.iter().map(|(k, v)| (k, v)))
        .envs(extra.iter().map(|(k, v)| (k, v)))
        .stdin(std::process::Stdio::null())
        .stdout(file)
        .stderr(err);
    match cmd.status() {
        Ok(status) => exit_of(status),
        Err(_) => 127,
    }
}

/// The `CARGO_BUILD_JOBS` of a cargo step when the caller set none: the logical CPUs shared out
/// over the slots, at least 2 (SPEC-20261122-check-affected-one-slot REQ-6).
pub fn default_jobs(cpus: usize, slots: usize) -> usize {
    (cpus / slots.max(1)).max(2)
}

/// The jobs value to add to the heavy steps' environment: `None` when the caller set
/// `CARGO_BUILD_JOBS` (theirs passes through unchanged).
fn build_jobs(slots: usize) -> Option<std::ffi::OsString> {
    if std::env::var_os("CARGO_BUILD_JOBS").is_some() {
        return None;
    }
    let cpus = std::thread::available_parallelism().map_or(1, |n| n.get());
    Some(default_jobs(cpus, slots).to_string().into())
}

/// The test-thread cap of the `test` step (SPEC-20261120-test-parallelism-cap REQ-1). The slots'
/// `limits.json` does not carry it yet: `heavy_slot::LimitsFile` refuses unknown keys, so the key
/// and its refusal of a bad value wait for the slice that adds the field there.
pub const DEFAULT_TEST_THREADS: usize = 4;

/// The value the caller already set for the test step's thread knob, if any: nextest's
/// `NEXTEST_TEST_THREADS` or libtest's `RUST_TEST_THREADS`, `get` being the environment lookup.
/// A set knob is kept as it is and no cap is injected over it (REQ-1).
fn caller_threads(
    runner: Runner,
    get: impl Fn(&str) -> Option<std::ffi::OsString>,
) -> Option<String> {
    let key = match runner {
        Runner::Nextest => "NEXTEST_TEST_THREADS",
        Runner::CargoTest => "RUST_TEST_THREADS",
    };
    get(key)
        .map(|v| v.to_string_lossy().into_owned())
        .filter(|v| !v.is_empty())
}

/// The slot folder of a run: the caller's `--slot-dir`, else the machine's.
fn slot_folder(opts: &RunOptions) -> Result<std::path::PathBuf> {
    opts.slot_dir
        .clone()
        .map(Ok)
        .unwrap_or_else(crate::heavy_slot::default_dir)
}

/// Takes the one heavy slot of a run, with its effective slot count. The waiting lines and any
/// refusal reason go to `<out>/heavy-slot.log` — stdout is the JSON answer. `Err` is the slot's
/// exit code: 125 refused, 75 machine too busy, 2 no usable slot folder or `--max 0`.
fn hold_slot(
    opts: &RunOptions,
    out: &Path,
) -> std::result::Result<(crate::heavy_slot::Held, usize), i32> {
    use std::io::Write;
    let mut log: Box<dyn Write> = match std::fs::File::create(out.join("heavy-slot.log")) {
        Ok(f) => Box::new(f),
        Err(_) => Box::new(std::io::sink()),
    };
    let folder = match slot_folder(opts) {
        Ok(d) => d,
        Err(e) => {
            let _ = writeln!(log, "heavy-slot: {e}");
            return Err(i32::from(crate::heavy_slot::USAGE));
        }
    };
    let (max, limits, notes) = match crate::heavy_slot::resolve(
        &folder,
        Some(opts.max as usize),
        Some(u64::from(opts.min_free_gb)),
        None,
        None,
    ) {
        Ok(r) => r,
        Err(e) => {
            let _ = writeln!(
                log,
                "heavy-slot: refused {e}; nothing was written and no cargo step ran"
            );
            return Err(i32::from(crate::heavy_slot::REFUSED));
        }
    };
    for n in notes {
        let _ = writeln!(log, "{n}");
    }
    if max < 1 {
        let _ = writeln!(log, "heavy-slot: --max 0 leaves no slot; no cargo step ran");
        return Err(i32::from(crate::heavy_slot::USAGE));
    }
    let name = "check-affected";
    // The slot record names this run's own command line, as `heavy-slot-acquire` names COMMAND.
    let command: Vec<String> = std::env::args_os()
        .map(|a| a.to_string_lossy().into_owned())
        .collect();
    let held = crate::heavy_slot::hold(&folder, max, name, &command, limits, &mut log)
        .map_err(i32::from)?;
    let _ = writeln!(log, "{}", held.taken_line(name));
    Ok((held, max))
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
/// every cheap step, then every heavy step unless a cheap one failed (the run stops before the
/// heavy steps and the summary says `stopped_at=<step>`), and give one summary line (REQ-8 to
/// REQ-11 of the earlier SPEC, changed by SPEC-20261122-check-light-stop). The heavy steps run
/// under one heavy slot, taken before the first and let go after the last; a
/// slot that is not obtained gives its exit code to every heavy step and none starts
/// (SPEC-20261122-check-affected-one-slot REQ-1, REQ-2).
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
    let threads = DEFAULT_TEST_THREADS;
    let sel = select_pick(
        opts.pick,
        &facts.changed,
        &facts.tests,
        facts.scan_list.as_deref(),
        host_os(),
    );
    let target = target_dir(&std::fs::canonicalize(dir).unwrap_or_else(|_| dir.to_path_buf()));
    let range3 = format!("{}...{}", facts.base, facts.head);
    // The cap goes onto the test step unless the caller set the runner's own knob; what the
    // summary shows is the caller's value then.
    let env = |k: &str| std::env::var_os(k);
    let cap_for = |r: Runner| caller_threads(r, env).is_none().then_some(threads);
    let mut shown = threads.to_string();
    // Planned with nextest until the slot is held: the runner is asked only then, so a refused
    // slot starts no cargo at all. The labels do not depend on the runner.
    let mut planned = steps(
        &sel,
        &facts.present,
        &range3,
        Runner::Nextest,
        cap_for(Runner::Nextest),
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
    // The cheap steps are the ones planned before the first heavy step. They all run; when one
    // failed, the run stops before the first heavy step (SPEC-20261122-check-light-stop).
    let cheap = planned.iter().take_while(|s| !s.heavy).count();
    let mut stopped_at: Option<String> = None;
    let mut tried = false;
    let mut held: Option<(crate::heavy_slot::Held, usize)> = None;
    let mut refused: Option<i32> = None;
    let mut runner: Option<Runner> = None;
    for i in 0..planned.len() {
        if i == cheap {
            if let Some((label, _)) = results.iter().find(|(_, c)| *c != 0) {
                stopped_at = Some(label.split(':').next().unwrap_or(label).to_string());
                break;
            }
        }
        if planned[i].heavy && !tried {
            tried = true;
            match hold_slot(opts, &out) {
                Ok(h) => {
                    let r = runner_in(dir);
                    planned = steps(&sel, &facts.present, &range3, r, cap_for(r));
                    if let Some(theirs) = caller_threads(r, env) {
                        shown = theirs;
                    }
                    runner = Some(r);
                    held = Some(h);
                }
                Err(code) => refused = Some(code),
            }
        }
        // Every step after the first heavy one is heavy, so the slot is let go once the loop
        // ends (D-2: after the last cargo step).
        let s = &planned[i];
        let log = out.join(format!(
            "{:02}-{}.log",
            i + 1,
            s.label.replace([':', '/'], "_")
        ));
        let code = match (s.heavy, &held, refused) {
            (true, Some((h, slots)), _) => {
                let mut extra = vec![(crate::heavy_slot::SLOT_FD_ENV, h.env_value())];
                if let Some(jobs) = build_jobs(*slots) {
                    extra.push(("CARGO_BUILD_JOBS", jobs));
                }
                run_step(dir, s, &log, &extra)
            }
            (true, None, Some(code)) => {
                let _ = std::fs::write(
                    &log,
                    format!("not run: the heavy slot was not obtained (exit {code}); see heavy-slot.log\n"),
                );
                code
            }
            _ => run_step(dir, s, &log, &[]),
        };
        let mut words = vec![s.program.clone()];
        words.extend(s.args.iter().cloned());
        results.push((s.label.clone(), code));
        records.push(serde_json::json!({
            "label": s.label, "exit": code, "log": log.to_string_lossy(), "command": words,
        }));
    }
    drop(held);
    let pass = results.iter().all(|(_, c)| *c == 0);
    // A `--lane` run says so right after `scans=`; the bare line has no such word
    // (SPEC-20261124-ready-receipt-followup REQ-6, REQ-9).
    let lane_word = if opts.pick == Pick::Lane {
        " selection=lane"
    } else {
        ""
    };
    let mut line = format!(
        "check-affected base={} head={} targets={} scans={}{lane_word} test-threads={}",
        short(&facts.base),
        short(&facts.head),
        sel.rust.len(),
        sel.scans.word(),
        shown
    );
    for (k, c) in summary_words(&results) {
        line.push_str(&format!(" {k}={c}"));
    }
    if let Some(label) = &stopped_at {
        line.push_str(&format!(" stopped_at={label}"));
    }
    line.push_str(if pass { " result=PASS" } else { " result=FAIL" });
    Ok(serde_json::json!({
        "action": "affected",
        "mode": "run",
        "selection": opts.pick.word(),
        "range_note": range_note(&facts, opts.integration_ref.as_deref()),
        "summary": line,
        "result": if pass { "PASS" } else { "FAIL" },
        "stopped_at": stopped_at,
        "runner": runner.map(Runner::word),
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
            modules: modules.iter().map(|s| s.to_string()).collect(),
            os: os.iter().map(|s| s.to_string()).collect(),
            text: String::new(),
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
        // The `unittest` check of the scan list names no target: only the one `--test` counts.
        assert_eq!(sel.scans, Scans::Listed(1));
    }

    /// SPEC-20261123-rust-scans-before-ready REQ-3: a Python test module is neither a test file
    /// of the project nor a target, whatever changed: a changed module, a changed fixture and a
    /// source file whose area has a module of the same name select nothing of it.
    #[test]
    fn no_python_module_is_selected_or_read_as_a_test() {
        let tests = [rust("update_cli", &[], &[])];
        let by = |p: &[&str]| select(&paths(p), &tests, Some(SCANS), "linux");
        let changed = by(&[
            "tests/update_tux_tty_test.py",
            "tests/repo_scans_test.py",
            "tools/source_manifest.py",
        ]);
        assert_eq!(
            changed.rust,
            ["home_isolation_scan".to_string()].into(),
            "only the scan list's own cargo target is selected: {changed:?}"
        );
        // A `unittest` word of the scan list is not a target either.
        let named: BTreeSet<String> = scan_entries(SCANS)
            .unwrap()
            .into_iter()
            .flat_map(|e| e.targets)
            .collect();
        assert_eq!(named, ["home_isolation_scan".to_string()].into());
    }

    #[test]
    fn a_changed_test_a_fixture_and_a_helper_select_their_targets() {
        let tests = [
            rust("help_layout", &[], &[]),
            rust("card_cli", &["common"], &[]),
            rust("bundle_cli", &["common"], &[]),
            rust("solo_cli", &[], &[]),
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
        let tests = [rust("a_cli", &[], &[]), rust("b_cli", &[], &[])];
        let sel = select(&paths(&["README.md"]), &tests, Some("{not json"), "linux");
        assert_eq!(sel.scans, Scans::Unreadable);
        assert!(sel.rust.contains("a_cli") && sel.rust.contains("b_cli"));
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

    /// SPEC-20261122-check-affected-one-slot REQ-3: the library, binary and selected target tests
    /// are one `test` step, run by nextest or by one `cargo test`.
    #[test]
    fn the_steps_always_hold_lib_bins_and_the_fixed_checks() {
        let sel = select(&paths(&["README.md"]), &[], None, "linux");
        let planned = steps(&sel, &[], "a...b", Runner::Nextest, None);
        let labels: Vec<&str> = planned.iter().map(|s| s.label.as_str()).collect();
        assert_eq!(labels, ["diff_check", "fmt", "clippy", "test"]);
        assert!(planned
            .iter()
            .filter(|s| s.program == "cargo" && s.label != "fmt")
            .all(|s| s.heavy));
        let test = |runner: Runner, sel: &Selection| -> Vec<String> {
            steps(sel, &[], "a...b", runner, None)
                .into_iter()
                .find(|s| s.label == "test")
                .unwrap()
                .args
        };
        assert_eq!(
            test(Runner::Nextest, &sel),
            [
                "nextest",
                "run",
                "--locked",
                "--offline",
                "--no-fail-fast",
                "--lib",
                "--bin",
                "moai-cli"
            ]
        );
        let targets = select(
            &paths(&["src/update_models.rs"]),
            &[rust("update_cli", &[], &[]), rust("update_x", &[], &[])],
            None,
            "linux",
        );
        assert_eq!(
            test(Runner::CargoTest, &targets),
            [
                "test",
                "--locked",
                "--offline",
                "--no-fail-fast",
                "--lib",
                "--bin",
                "moai-cli",
                "--test",
                "update_cli",
                "--test",
                "update_x"
            ]
        );
        let planned = steps(&targets, &[], "a...b", Runner::CargoTest, None);
        let labels: Vec<&str> = planned.iter().map(|s| s.label.as_str()).collect();
        assert_eq!(labels, ["diff_check", "fmt", "clippy", "test"]);
        // D-5: no step selects every binary's harness.
        for runner in [Runner::Nextest, Runner::CargoTest] {
            for sel in [&sel, &targets] {
                assert!(!test(runner, sel).iter().any(|a| a == "--bins"));
            }
        }
    }

    /// SPEC-20261120-test-parallelism-cap REQ-1: the cap goes onto the `test` step only —
    /// `-j <n>` right after `run` for nextest, `RUST_TEST_THREADS` for libtest — and nowhere else.
    #[test]
    fn the_thread_cap_goes_onto_the_test_step_alone() {
        let sel = select(&paths(&["README.md"]), &[], None, "linux");
        let planned = steps(&sel, &[], "a...b", Runner::Nextest, Some(3));
        for s in &planned {
            let capped = s.args.iter().any(|a| a == "-j");
            assert_eq!(capped, s.label == "test", "{}: {:?}", s.label, s.args);
            assert!(s.env.is_empty(), "{}: {:?}", s.label, s.env);
        }
        let test = planned.iter().find(|s| s.label == "test").unwrap();
        assert_eq!(
            test.args[..5],
            ["nextest", "run", "-j", "3", "--locked"],
            "{:?}",
            test.args
        );

        let planned = steps(&sel, &[], "a...b", Runner::CargoTest, Some(3));
        for s in &planned {
            let capped = s
                .env
                .contains(&("RUST_TEST_THREADS".to_string(), "3".to_string()));
            assert_eq!(capped, s.label == "test", "{}: {:?}", s.label, s.env);
            assert!(
                !s.args.iter().any(|a| a == "-j"),
                "{}: {:?}",
                s.label,
                s.args
            );
        }
        // `None` (the caller's own knob, or no cap asked) adds nothing, to either runner.
        for runner in [Runner::Nextest, Runner::CargoTest] {
            for s in steps(&sel, &[], "a...b", runner, None) {
                assert!(s.env.is_empty() && !s.args.iter().any(|a| a == "-j"));
            }
        }
    }

    /// SPEC-20261123-check-precise-selection REQ-3: a sub-module is named whole, by a path or as
    /// an item of a group, never by a longer name or another path's tail.
    #[test]
    fn a_common_sub_module_is_named_by_whole_identifiers_only() {
        assert!(names_common_sub("use common::git::g;", "git"));
        assert!(names_common_sub("crate::common::home::h()", "home"));
        assert!(names_common_sub(
            "use common::{cli, git::{a, b}, home};",
            "home"
        ));
        assert!(names_common_sub("use common::{cli,\n    git::g};", "git"));
        assert!(!names_common_sub("use common::home_x::h;", "home"));
        assert!(!names_common_sub("use common::{git::{home, b}};", "home"));
        assert!(!names_common_sub("use mycommon::home::h;", "home"));
        assert!(!names_common_sub("let common = home::h();", "home"));
    }

    fn pinned(name: &str, text: &str, modules: &[&str], os: &[&str]) -> TestFile {
        TestFile {
            text: text.to_string(),
            ..rust(name, modules, os)
        }
    }

    /// REQ-2/REQ-4/REQ-6 beside the command's own fixtures: the lane set of a library call is a
    /// subset of the full set, and the text pin joins both.
    #[test]
    fn the_lane_set_is_inside_the_full_set_and_the_text_pin_joins_both() {
        let tests = [
            pinned("page_cli", "\"docs/a.md\"", &[], &[]),
            pinned("mac_only", "", &[], &[host_os()]),
            pinned("scan_a", "", &[], &[]),
        ];
        let list = r#"{"scans":[{"checks":["cargo test --test scan_a"]}]}"#;
        let changed = paths(&["docs/a.md"]);
        let full = select_pick(Pick::Full, &changed, &tests, Some(list), host_os());
        let lane = select_pick(Pick::Lane, &changed, &tests, Some(list), host_os());
        assert!(lane.rust.is_subset(&full.rust), "{lane:?} {full:?}");
        assert_eq!(lane.rust, ["page_cli".to_string()].into());
        assert_eq!(
            full.rust,
            ["mac_only", "page_cli", "scan_a"].map(String::from).into()
        );
        assert_eq!(select(&changed, &tests, Some(list), host_os()), full);
    }

    /// REQ-1: the cap is 4 until the slots' `limits.json` can carry a value of its own.
    #[test]
    fn the_default_cap_is_four() {
        assert_eq!(DEFAULT_TEST_THREADS, 4);
    }

    /// REQ-1: the caller's own knob is kept — nextest's for nextest, libtest's for `cargo test`,
    /// and an empty value counts as not set.
    #[test]
    fn a_caller_set_knob_is_seen_for_its_own_runner_only() {
        let env = |k: &str, v: &str| {
            let (k, v) = (k.to_string(), std::ffi::OsString::from(v));
            move |q: &str| (q == k).then(|| v.clone())
        };
        assert_eq!(
            caller_threads(Runner::CargoTest, env("RUST_TEST_THREADS", "8")),
            Some("8".to_string())
        );
        assert_eq!(
            caller_threads(Runner::Nextest, env("NEXTEST_TEST_THREADS", "num-cpus")),
            Some("num-cpus".to_string())
        );
        assert_eq!(
            caller_threads(Runner::Nextest, env("RUST_TEST_THREADS", "8")),
            None
        );
        assert_eq!(
            caller_threads(Runner::CargoTest, env("NEXTEST_TEST_THREADS", "8")),
            None
        );
        assert_eq!(
            caller_threads(Runner::CargoTest, env("RUST_TEST_THREADS", "")),
            None
        );
        assert_eq!(caller_threads(Runner::CargoTest, |_| None), None);
    }

    /// REQ-6: the logical CPUs shared out over the slots, never under 2.
    #[test]
    fn the_default_jobs_share_the_cpus_over_the_slots_and_never_go_under_two() {
        for (cpus, slots, jobs) in [
            (16, 3, 5),
            (16, 4, 4),
            (6, 3, 2),
            (5, 3, 2),
            (9, 3, 3),
            (8, 3, 2),
            (1, 1, 2),
            (2, 1, 2),
            (3, 1, 3),
            (16, 0, 16),
        ] {
            assert_eq!(default_jobs(cpus, slots), jobs, "{cpus} / {slots}");
        }
    }

    #[test]
    fn changed_operation_files_and_scripts_get_their_own_checks() {
        let sel = select(
            &paths(&["src/ops/x.rs", "tools/a.sh"]),
            &[rust("x_cli", &[], &[])],
            None,
            "linux",
        );
        let present = paths(&["src/ops/x.rs", "tools/a.sh"]);
        let planned = steps(&sel, &present, "a...b", Runner::Nextest, None);
        let ops = planned.iter().find(|s| s.label == "ops_rustfmt").unwrap();
        assert_eq!(ops.args, ["--edition", "2021", "--check", "src/ops/x.rs"]);
        assert!(planned.iter().any(|s| s.label == "bash_n:tools/a.sh"));
    }

    /// SPEC-20261123-rust-scans-before-ready REQ-3: no step runs Python and none builds the
    /// binary apart from the one cargo test call, whatever the selection holds.
    #[test]
    fn no_step_runs_python_or_builds_the_binary_by_itself() {
        let tests = [rust("update_command_cli", &[], &[])];
        for changed in [
            paths(&["src/update_models.rs"]),
            paths(&["src/ops/x.rs"]),
            paths(&["tests/update_tux_tty_test.py", "tools/mutate.py"]),
        ] {
            let sel = select(&changed, &tests, Some(SCANS), "linux");
            let planned = steps(&sel, &changed, "a...b", Runner::Nextest, None);
            for s in &planned {
                assert!(
                    !s.program.contains("python") && !s.args.iter().any(|a| a.contains("unittest")),
                    "{s:?} runs Python"
                );
                assert!(s.label != "python" && s.label != "build_cli", "{s:?}");
            }
        }
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
