//! `check affected` (SPEC-20261120-check-affected): from the paths a range changes, pick the
//! tests the change can break and the steps that check it. [`select`] and [`steps`] are pure
//! (changed paths, test file facts, the scan list text and the host name in; a selection and
//! a step list out); [`read_facts`] is the one place that reads git and the files.

use crate::project_ci::CheckEntry;
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
    /// The area crate that holds this module (`tests/<area>/<name>.rs`), `None` for a target of
    /// its own. A moved module keeps its old target name in `name`, so the selection does not
    /// change; only the step that runs it names the crate (SPEC-20261123-test-area-crates).
    pub area: Option<String>,
}

/// Which set a caller gets: the full set (today's, the bare command) or the precise lane set
/// (`--lane`), a subset of the full set for the same range (SPEC-20261123-check-precise-selection
/// REQ-1).
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq)]
pub enum Pick {
    #[default]
    Full,
    Lane,
    /// What one card's change reaches (`--card`, SPEC-20261124-card-select-batch-land): see
    /// [`select_card`]. A path whose effect is not clear makes [`choose`] answer the full set.
    Card,
}

impl Pick {
    /// The word of the answer's `selection` field.
    pub fn word(self) -> &'static str {
        match self {
            Pick::Full => "full",
            Pick::Lane => "lane",
            Pick::Card => "card",
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
    /// Old target name to area crate, for the selected names that are modules of an area crate.
    pub areas: std::collections::BTreeMap<String, String>,
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
    id: String,
    /// The test files its `tests` entries name (the text before `::` and before ` (`).
    test_files: Vec<String>,
    targets: BTreeSet<String>,
    /// `"always": true`: the landing requires the scan whenever the diff touches it. The full set
    /// adds its targets only then (SPEC-20261123-land-file-read-scans D-9); a scan that is not
    /// `always` joins the full set whatever changed.
    always: bool,
    rules: Vec<String>,
    scope: Option<Vec<(String, Vec<String>)>>,
}

impl ScanEntry {
    /// Whether `path` is in the scope: under a `prefix`, and with one of its `suffixes` when it
    /// names any (the rule of `scope_covers` in `tests/repo_scans_table.rs`).
    fn covers(&self, path: &str) -> bool {
        self.scope.iter().flatten().any(|(prefix, suffixes)| {
            path.starts_with(prefix.as_str())
                && (suffixes.is_empty() || suffixes.iter().any(|s| path.ends_with(s.as_str())))
        })
    }

    /// An `always` scan that names `file` as a rule file and holds `path` in its scope.
    fn always_covers(&self, file: &str, path: &str) -> bool {
        self.always && self.rules.iter().any(|r| r == file) && self.covers(path)
    }

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
        for check in crate::repo_scans::checks_of(scan).ok()? {
            let words: Vec<&str> = check.split_whitespace().collect();
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
            id: scan
                .get("id")
                .and_then(|i| i.as_str())
                .unwrap_or("?")
                .to_string(),
            test_files: strings(scan.get("tests"))
                .iter()
                .map(|t| {
                    let head = t.split("::").next().unwrap_or(t);
                    head.split(" (").next().unwrap_or(head).to_string()
                })
                .collect(),
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
    // The card set is made by `select_card`; asked here, it is the lane set.
    let pick = if pick == Pick::Card { Pick::Lane } else { pick };
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
            // A module of an area crate (`tests/<area>/<module>.rs`) is the old target `<module>`;
            // the crate's own `main.rs` is every module of that area.
            [dir, file]
                if file.ends_with(".rs")
                    && tests.iter().any(|t| t.area.as_deref() == Some(*dir)) =>
            {
                let name = stem(file);
                for t in tests.iter().filter(|t| {
                    t.area.as_deref() == Some(*dir) && (*file == "main.rs" || t.name == name)
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
    let areas = areas_of(tests, &rust);
    Selection { rust, scans, areas }
}

/// Old target name to area crate for the selected names that are modules of an area crate.
fn areas_of(
    tests: &[TestFile],
    rust: &BTreeSet<String>,
) -> std::collections::BTreeMap<String, String> {
    tests
        .iter()
        .filter(|t| rust.contains(&t.name))
        .filter_map(|t| Some((t.name.clone(), t.area.clone()?)))
        .collect()
}

/// How many test targets may carry the name word of a changed module before the word says nothing
/// (SPEC-20261124-card-select-batch-land REQ-2).
const COMMON_WORD_CAP: usize = 10;

/// Why `--card` answered the full set: the reason class of the first wide path and every wide path.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Wide {
    pub reason: &'static str,
    pub paths: Vec<String>,
}

/// What the card set decided (SPEC-20261124-card-select-batch-land).
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct CardPick {
    /// The card set; when `wide` is set, the caller uses the full set instead.
    pub selection: Selection,
    pub wide: Option<Wide>,
    /// The name words of changed modules that more than [`COMMON_WORD_CAP`] targets carry.
    pub ignored_words: Vec<String>,
}

/// The class of a path whose effect on the tests is not clear (REQ-3): the whole dependency graph,
/// the build and the common settings, the permission and hook guards and the ledger's storage
/// layer. Everything else under `src/ledger/` is ordinary code and takes the card set (D-1).
fn wide_class(path: &str) -> Option<&'static str> {
    let file = path.rsplit('/').next().unwrap_or(path);
    if matches!(file, "Cargo.toml" | "Cargo.lock") {
        return Some("dependencies");
    }
    if matches!(file, "build.rs" | "rust-toolchain" | "rust-toolchain.toml")
        || path.starts_with(".cargo/")
    {
        return Some("build settings");
    }
    if path == ".moai/project.yaml" {
        return Some("common settings");
    }
    if matches!(
        path,
        "src/ledger/schema.rs"
            | "src/ledger/migrate.rs"
            | "src/ledger/open.rs"
            | "src/ledger/recover.rs"
    ) || path.starts_with("migrations/")
        || path.contains("/migrations/")
    {
        return Some("ledger storage");
    }
    if (path.starts_with("src/hook_") && path.ends_with(".rs"))
        || matches!(
            path,
            "src/hooks.rs" | "src/scope.rs" | "src/scopes.rs" | "src/claude_settings_merge.rs"
        )
    {
        return Some("permission");
    }
    // The audit floor (`gate::TIER2_PREFIXES`, the paths a full audit covers), the git hooks and the
    // landing code: a change there can break a test that does not name its module.
    if crate::gate::TIER2_PREFIXES
        .iter()
        .any(|p| path.starts_with(p))
        || path.starts_with("install/git-hooks/")
        || path.starts_with("src/git_hooks/")
        || path.starts_with("src/card/land/")
    {
        return Some("audit floor");
    }
    None
}

/// The module a changed source path names: its leaf word and the path a test names it by
/// (`moai_cli::<leaf>` for a top-level file, `<dir>::<leaf>` below one). `mod.rs`, `lib.rs`,
/// `main.rs` and the binary's `src/ops/` name no module a test can call.
fn module_of(path: &str) -> Option<(String, Option<String>)> {
    let rest = path.strip_prefix("src/")?.strip_suffix(".rs")?;
    let parts: Vec<&str> = rest.split('/').collect();
    let leaf = *parts.last()?;
    if matches!(leaf, "mod" | "lib" | "main") {
        return None;
    }
    let leaf = leaf.replace('-', "_");
    // The word a test name carries: `<dir>_<leaf>` below a folder (`card_sweep` for
    // `src/card/sweep.rs`), the leaf alone for a top-level file.
    let (word, key) = match parts.as_slice() {
        ["ops", _] => (leaf, None),
        [_] => (leaf.clone(), Some(format!("moai_cli::{leaf}"))),
        [.., dir, _] => {
            let dir = dir.replace('-', "_");
            (format!("{dir}_{leaf}"), Some(format!("{dir}::{leaf}")))
        }
        [] => return None,
    };
    Some((word, key))
}

/// Whether `text` names `key` as a whole path: no identifier character before it or after it.
fn names_path(text: &str, key: &str) -> bool {
    let ident = |c: char| c.is_alphanumeric() || c == '_';
    let mut from = 0;
    while let Some(i) = text[from..].find(key) {
        let at = from + i;
        let before_ok = !text[..at].chars().next_back().is_some_and(ident);
        let after_ok = !text[at + key.len()..].chars().next().is_some_and(ident);
        if before_ok && after_ok {
            return true;
        }
        from = at + key.len();
    }
    false
}

/// Whether `word` is a whole run of `_` separated words of the target name.
fn name_has_word(name: &str, word: &str) -> bool {
    format!("_{name}_").contains(&format!("_{word}_"))
}

/// Pick the test targets one card's change reaches (SPEC-20261124-card-select-batch-land REQ-1 to
/// REQ-3): the tests the range changes or adds, the tests whose source names a changed module or
/// whose name carries its word, the targets of each scan a changed path touches, and the tests
/// that name a changed `skills/` or `docs/` path. A wide path (see [`wide_class`]) or a scan list
/// that cannot be read sets `wide`: the caller then uses the full set.
pub fn select_card(
    changed: &[String],
    tests: &[TestFile],
    scan_list: Option<&str>,
    host_os: &str,
) -> CardPick {
    let mut wide_paths: Vec<(&'static str, String)> = changed
        .iter()
        .filter_map(|p| wide_class(p).map(|c| (c, p.clone())))
        .collect();
    // The tests, fixtures and documents the range changes: the lane rules, no OS rule.
    let own: Vec<String> = changed
        .iter()
        .filter(|p| p.starts_with("tests/") || p.starts_with("skills/") || p.starts_with("docs/"))
        .cloned()
        .collect();
    let mut rust = select_pick(Pick::Lane, &own, tests, None, host_os).rust;
    let mut ignored = BTreeSet::new();
    for path in changed {
        let is_source = path.starts_with("src/") && path.ends_with(".rs");
        let Some((word, key)) = module_of(path) else {
            // `mod.rs`, `lib.rs` and `main.rs` name no module a test can call: nothing narrows
            // them, so the full set runs (a source path that picks no test is not a pass).
            if is_source {
                wide_paths.push(("no test reached", path.clone()));
            }
            continue;
        };
        // The targets that carry the module's name word or name its path in their source; a
        // module that more than the cap of them reach is a common one and says nothing.
        let reached: BTreeSet<&str> = tests
            .iter()
            .filter(|t| {
                name_has_word(&t.name, &word)
                    || key.as_ref().is_some_and(|k| names_path(&t.text, k))
            })
            .map(|t| t.name.as_str())
            .collect();
        if reached.len() > COMMON_WORD_CAP {
            // A word that says nothing is not a reason to run nothing: the full set runs.
            ignored.insert(word);
            wide_paths.push(("common name word", path.clone()));
        } else if reached.is_empty() && is_source {
            wide_paths.push(("no test reached", path.clone()));
        } else {
            rust.extend(reached.into_iter().map(String::from));
        }
    }
    // A path is named once, under the first reason it was found for.
    let mut seen = BTreeSet::new();
    wide_paths.retain(|(_, p)| seen.insert(p.clone()));
    let mut wide = wide_paths.first().map(|(reason, _)| Wide {
        reason,
        paths: wide_paths.iter().map(|(_, p)| p.clone()).collect(),
    });
    let mut scans = Scans::None;
    if let Some(text) = scan_list {
        match scan_entries(text) {
            Some(entries) => {
                // A scan's scope is what it reads, often the whole tree; the landing asks for
                // those receipts (SPEC-20261123-land-repo-scans) and the batch's wide check runs
                // them. The card set takes a scan only when one of its rule files changed (D-6).
                let picked: BTreeSet<String> = entries
                    .into_iter()
                    .filter(|e| changed.iter().any(|p| e.rules.iter().any(|r| r == p)))
                    .flat_map(|e| e.targets)
                    .collect();
                scans = Scans::Listed(picked.len());
                rust.extend(picked);
            }
            None => {
                scans = Scans::Unreadable;
                wide.get_or_insert(Wide {
                    reason: "scan table",
                    paths: vec!["tools/repo_scans.json".to_string()],
                });
            }
        }
    }
    let areas = areas_of(tests, &rust);
    CardPick {
        selection: Selection { rust, scans, areas },
        wide,
        ignored_words: ignored.into_iter().collect(),
    }
}

/// What a run uses for `pick`: the effective pick and its selection, and for `--card` the notes the
/// answer carries. A card pick with a wide path is the full set, as the bare command picks it, and
/// the summary line of such a run has no `selection=` word.
pub struct Chosen {
    pub pick: Pick,
    pub selection: Selection,
    /// Only for a `--card` request: the wide reason, if any, and the ignored name words.
    pub card: Option<(Option<Wide>, Vec<String>)>,
}

pub fn choose(
    pick: Pick,
    changed: &[String],
    tests: &[TestFile],
    scan_list: Option<&str>,
    host_os: &str,
) -> Chosen {
    if pick != Pick::Card {
        return Chosen {
            pick,
            selection: select_pick(pick, changed, tests, scan_list, host_os),
            card: None,
        };
    }
    let card = select_card(changed, tests, scan_list, host_os);
    if card.wide.is_some() {
        return Chosen {
            pick: Pick::Full,
            selection: select_pick(Pick::Full, changed, tests, scan_list, host_os),
            card: Some((card.wide, card.ignored_words)),
        };
    }
    Chosen {
        pick: Pick::Card,
        selection: card.selection,
        card: Some((None, card.ignored_words)),
    }
}

/// The answer fields a `--card` request adds: `card_wide` (`{reason, paths}` or null) and
/// `card_ignored_words`.
fn card_fields(chosen: &Chosen, answer: &mut serde_json::Value) {
    if let Some((wide, ignored)) = &chosen.card {
        answer["card_wide"] = match wide {
            Some(w) => serde_json::json!({"reason": w.reason, "paths": w.paths}),
            None => serde_json::Value::Null,
        };
        answer["card_ignored_words"] = serde_json::json!(ignored);
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

/// The earlier and the new state folder names, built at run time (`tests/state_folder_literal_scan.rs`
/// counts them the same way).
fn earlier_state_name() -> String {
    format!(".{}-{}", "moai", "cli")
}

fn new_state_name() -> String {
    format!(".{}", "moai")
}

fn is_word_char(c: char) -> bool {
    c.is_ascii_alphanumeric() || c == '_'
}

/// How often `name` stands in `text` as a whole path segment: not right after a word character
/// and not followed by a word character or `-` (the rule of `tests/state_folder_literal_scan.rs`).
fn count_state_name(text: &str, name: &str) -> usize {
    let mut n = 0;
    let mut from = 0;
    while let Some(at) = text[from..].find(name) {
        let start = from + at;
        let end = start + name.len();
        let before_ok = text[..start]
            .chars()
            .next_back()
            .is_none_or(|c| !is_word_char(c));
        let after_ok = text[end..]
            .chars()
            .next()
            .is_none_or(|c| !is_word_char(c) && c != '-');
        if before_ok && after_ok {
            n += 1;
        }
        from = end;
    }
    n
}

/// The quoted strings of one `key = ["a", "b"]` line's right side.
fn quoted(rest: &str) -> Vec<String> {
    rest.split('"')
        .enumerate()
        .filter(|(i, _)| i % 2 == 1)
        .map(|(_, s)| s.to_string())
        .collect()
}

/// `tests/areas.toml` read by lines (one array per line, as the table's own header asks): the
/// `[areas]` arrays in file order and the `[solo]` files.
fn parse_area_table(text: &str) -> (Vec<(String, Vec<String>)>, Vec<String>) {
    let mut areas = Vec::new();
    let mut solo = Vec::new();
    let mut section = "";
    for line in text.lines().map(str::trim) {
        if line.is_empty() || line.starts_with('#') {
            continue;
        }
        if line.starts_with('[') && !line.contains('=') {
            section = line.trim_matches(|c| c == '[' || c == ']').trim();
            continue;
        }
        let Some((key, rest)) = line.split_once('=') else {
            continue;
        };
        match section {
            "areas" => areas.push((key.trim().to_string(), quoted(rest))),
            "solo" if key.trim() == "files" => solo = quoted(rest),
            _ => {}
        }
    }
    (areas, solo)
}

/// The rows `path<TAB>earlier<TAB>new` of the state-folder allow list.
fn parse_allow_rows(text: &str) -> Vec<(String, usize, usize)> {
    text.lines()
        .filter(|l| !l.trim().is_empty() && !l.starts_with('#'))
        .filter_map(|l| {
            let cols: Vec<&str> = l.trim_end().split('\t').collect();
            match cols.as_slice() {
                [path, e, n] => Some((path.to_string(), e.parse().ok()?, n.parse().ok()?)),
                _ => None,
            }
        })
        .collect()
}

/// The lines that tell what a range's added top-level `tests/*.rs` files break in the two tables
/// that `tests/test_area_layout.rs` and `tests/state_folder_literal_scan.rs` read
/// (SPEC-20261124-new-test-file-precheck): the first word of a file name that no area lists, a
/// file of a moved area, a file named like an area crate, and a state folder name that the allow
/// list does not cover. `moved` are the areas that hold a `tests/<area>/main.rs`; a table that is
/// `None` is absent and its rules are skipped. Each line names the file, the table and what to
/// add. The tables are only read.
pub fn table_violations(
    added: &[String],
    tests: &[TestFile],
    areas: Option<&str>,
    allow: Option<&str>,
    moved: &[String],
) -> Vec<String> {
    let mut out = Vec::new();
    let stems: Vec<&str> = added
        .iter()
        .filter_map(|p| p.strip_prefix("tests/")?.strip_suffix(".rs"))
        .filter(|s| !s.contains('/'))
        .collect();
    if let Some(text) = areas {
        let (table, solo) = parse_area_table(text);
        let area_of = |word: &str| {
            table
                .iter()
                .find(|(_, words)| words.iter().any(|w| w == word))
                .map(|(name, _)| name.as_str())
        };
        for stem in &stems {
            let word = first_word(stem);
            match area_of(word) {
                None => out.push(format!(
                    "tests/{stem}.rs: the first word `{word}` is not in tests/areas.toml; add \"{word}\" to one array of [areas] ({})",
                    table
                        .iter()
                        .map(|(n, _)| n.as_str())
                        .collect::<Vec<_>>()
                        .join(", ")
                )),
                Some(area) => {
                    if moved.iter().any(|m| m == area) && !solo.iter().any(|s| s == stem) {
                        out.push(format!(
                            "tests/{stem}.rs: the area `{area}` is moved (tests/{area}/main.rs exists); put the file at tests/{area}/{stem}.rs and declare `mod {stem};` in tests/{area}/main.rs"
                        ));
                    }
                }
            }
            if table.iter().any(|(name, _)| name == stem) {
                out.push(format!(
                    "tests/{stem}.rs: the name is that of the area crate `{stem}` in tests/areas.toml; rename the file"
                ));
            }
        }
    }
    if let Some(text) = allow {
        let rows = parse_allow_rows(text);
        let (earlier, new) = (earlier_state_name(), new_state_name());
        // The added files whose text the allow list counts: `tests/<name>.rs` and a module of a
        // moved area, `tests/<area>/<name>.rs`, which the list names by that longer path.
        let mut files: Vec<(String, &str, Option<&str>)> = stems
            .iter()
            .map(|s| (format!("tests/{s}.rs"), *s, None))
            .collect();
        for p in added {
            let Some((area, name)) = p
                .strip_prefix("tests/")
                .and_then(|r| r.strip_suffix(".rs"))
                .and_then(|r| r.split_once('/'))
            else {
                continue;
            };
            if !name.contains('/') && name != "main" && moved.iter().any(|m| m == area) {
                files.push((p.clone(), name, Some(area)));
            }
        }
        for (path, stem, area) in &files {
            let Some(file) = tests
                .iter()
                .find(|t| t.name == *stem && t.area.as_deref() == *area)
            else {
                continue;
            };
            let have = (
                count_state_name(&file.text, &earlier),
                count_state_name(&file.text, &new),
            );
            if have == (0, 0) {
                continue;
            }
            let row = format!("{path}\t{}\t{}", have.0, have.1);
            match rows.iter().find(|(p, _, _)| p == path) {
                None => out.push(format!(
                    "{path}: spells a state folder name ({earlier} x{}, {new} x{}) and is not in tests/fixtures/state_folder_literals/allow.tsv; use the resolvers, or for a fixture that builds an earlier layout on purpose add the row `{row}`",
                    have.0, have.1
                )),
                Some((_, ae, an)) => {
                    for (name, count, allowed) in [(&earlier, have.0, *ae), (&new, have.1, *an)] {
                        if count > allowed {
                            out.push(format!(
                                "{path}: spells {name} {count} times, the baseline in tests/fixtures/state_folder_literals/allow.tsv allows {allowed}; the row would be `{row}`"
                            ));
                        }
                    }
                }
            }
        }
    }
    out
}

/// The lines that tell what a range's added top-level `tests/*.rs` files break in the repository
/// scan list that `tests/repo_scans_table.rs` keeps whole (SPEC-20261124-repo-scans-precheck): a
/// file that walks the tree, one that reads a file outside `tests/` with `include_str!` or
/// `include_bytes!`, one that opens a checked-in file by its path, and a rule file of an `always`
/// scan that no check of the scan runs. `scan_list` is the text of `tools/repo_scans.json`
/// (`None` is an absent list), `tracked` the files git tracks (`None` when they cannot be read).
/// The exemption lists are read from the text of `tests/repo_scans_table.rs` among `tests`. The
/// scan list and the exemption lists are only read.
pub fn scan_violations(
    added: &[String],
    tests: &[TestFile],
    scan_list: Option<&str>,
    tracked: Option<&[String]>,
) -> Vec<String> {
    let Some(scans) = scan_list.and_then(scan_entries) else {
        return Vec::new();
    };
    let stems: Vec<&str> = added
        .iter()
        .filter_map(|p| p.strip_prefix("tests/")?.strip_suffix(".rs"))
        .filter(|s| !s.contains('/'))
        .collect();
    let exempt_text = tests
        .iter()
        .find(|t| t.name == "repo_scans_table")
        .map_or("", |t| t.text.as_str());
    let walk_exempt = exempt_rows(exempt_text, "EXEMPT");
    let include_exempt = exempt_rows(exempt_text, "INCLUDE_STR_EXEMPT");
    let read_exempt = exempt_rows(exempt_text, "FILE_READ_EXEMPT");
    let tracked = tracked.map(Tracked::new);
    let mut out = Vec::new();
    for stem in &stems {
        let Some(test) = tests.iter().find(|t| t.name == *stem) else {
            continue;
        };
        let path = format!("tests/{stem}.rs");
        let (text, names_root) = (test.text.as_str(), test.text.contains("CARGO_MANIFEST_DIR"));
        if names_root
            && RUST_WALK.iter().any(|n| text.contains(n))
            && !scans
                .iter()
                .any(|s| s.rules.contains(&path) || s.test_files.contains(&path))
            && !walk_exempt.iter().any(|(f, _)| *f == path)
        {
            out.push(format!(
                "{path}: names the repository root and walks the tree; add an entry that names it as a rule file to tools/repo_scans.json, or, when it only walks temporary or named files, a row with the reason to EXEMPT in tests/repo_scans_table.rs"
            ));
        }
        for read in include_str_paths(&path, text) {
            if scans.iter().any(|s| s.always_covers(&path, &read))
                || include_exempt.iter().any(|(f, _)| *f == path)
            {
                continue;
            }
            out.push(format!(
                "{path}: reads {read} with include_str! or include_bytes! and no `always` scan names it as a rule file with that path in its scope; add an entry with \"always\": true and a check `--test {stem}` to tools/repo_scans.json, or a row with the reason to INCLUDE_STR_EXEMPT in tests/repo_scans_table.rs"
            ));
        }
        if !names_root {
            continue;
        }
        let Some(tracked) = &tracked else {
            out.push(format!(
                "{path}: names the repository root but `git ls-files` cannot be read, so the checked-in files it may read are not known; run `check affected` again in a git checkout"
            ));
            continue;
        };
        for read in file_read_paths(text, tracked) {
            if scans.iter().any(|s| s.always_covers(&path, &read))
                || read_exempt
                    .iter()
                    .any(|(f, paths)| *f == path && (paths.is_empty() || paths.contains(&read)))
            {
                continue;
            }
            out.push(format!(
                "{path}: opens {read} at run time and no `always` scan names it as a rule file with that path in its scope; add it to a `reads-` scan of tools/repo_scans.json with a check `--test {stem}`, or a row with the reason to FILE_READ_EXEMPT in tests/repo_scans_table.rs when the path is one of the temporary project it builds"
            ));
        }
    }
    for scan in scans.iter().filter(|s| s.always) {
        for rule in &scan.rules {
            let stem = rule
                .strip_prefix("tests/")
                .and_then(|r| r.strip_suffix(".rs"))
                .filter(|s| !s.contains('/'));
            if let Some(stem) = stem.filter(|s| stems.contains(s) && !scan.targets.contains(*s)) {
                out.push(format!(
                    "{}: {rule} is a rule file but no check runs --test {stem}; add `cargo test --offline --locked --test {stem}` to its checks in tools/repo_scans.json",
                    scan.id
                ));
            }
        }
    }
    out
}

/// What marks a Rust test file as walking the tree: it names the repository root and uses one of
/// these (`RUST_WALK` of `tests/repo_scans_table.rs`).
const RUST_WALK: &[&str] = &[
    "read_dir(",
    "ls-files",
    "package_digest(",
    "stage_package(",
    "package_agents(",
    "walk_files(",
    "package_files(",
];

/// The literals that start an interpreter: a file that names one and the repository root runs a
/// checked-in script, so its literal `tools` is the script folder.
const INTERPRETER_LITERALS: &[&str] = &["\"python3\"", "\"bash\"", "\"sh\""];

/// The files git tracks and every folder above them, relative to the root with `/`.
struct Tracked {
    files: BTreeSet<String>,
    dirs: BTreeSet<String>,
}

impl Tracked {
    fn new(paths: &[String]) -> Tracked {
        let (mut files, mut dirs) = (BTreeSet::new(), BTreeSet::new());
        for path in paths {
            for (at, _) in path.match_indices('/') {
                dirs.insert(path[..at].to_string());
            }
            files.insert(path.clone());
        }
        Tracked { files, dirs }
    }

    /// The tracked path a string literal names, if it names one (`repo_path` of
    /// `tests/repo_scans_table.rs`): no space or backslash, a placeholder folder in front and
    /// anything from a placeholder on cut away, outside `tests/`, `target/` and `.git/`. A file
    /// counts at any depth, a folder only with two segments or more.
    fn repo_path(&self, literal: &str) -> Option<String> {
        if literal.chars().any(|c| c.is_whitespace() || c == '\\') {
            return None;
        }
        let mut literal = literal;
        while let Some(rest) = literal
            .strip_prefix('{')
            .and_then(|r| r.split_once("}/"))
            .filter(|(name, _)| !name.contains(['{', '}', '/']))
            .map(|(_, rest)| rest)
        {
            literal = rest;
        }
        let cut = literal
            .find(['{', '<', '*', '$'])
            .map_or(literal, |at| &literal[..at]);
        let path = cut.trim_start_matches("./").trim_matches('/');
        let first = path.split('/').next().unwrap_or("");
        if path.is_empty() || path.contains("..") || ["tests", "target", ".git"].contains(&first) {
            return None;
        }
        let is_file = self.files.contains(path);
        let is_folder = path.contains('/') && self.dirs.contains(path);
        (is_file || is_folder).then(|| path.to_string())
    }
}

/// The string literals of Rust source, escapes as written. Line and block comments, raw strings
/// and the character literals `'"'` and `'\"'` are skipped so their quotes pair nothing.
fn string_literals(text: &str) -> Vec<String> {
    let b = text.as_bytes();
    let ident = |c: u8| c.is_ascii_alphanumeric() || c == b'_';
    let (mut out, mut i) = (Vec::new(), 0);
    while i < b.len() {
        match b[i] {
            b'/' if b.get(i + 1) == Some(&b'/') => {
                while i < b.len() && b[i] != b'\n' {
                    i += 1;
                }
            }
            b'/' if b.get(i + 1) == Some(&b'*') => {
                i += 2;
                while i + 1 < b.len() && !(b[i] == b'*' && b[i + 1] == b'/') {
                    i += 1;
                }
                i += 2;
            }
            b'\'' if b.get(i + 1) == Some(&b'"') && b.get(i + 2) == Some(&b'\'') => i += 3,
            b'\'' if b.get(i + 1..i + 4) == Some(b"\\\"'") => i += 4,
            b'r' if i == 0 || !ident(b[i - 1]) => {
                let hashes = b[i + 1..].iter().take_while(|c| **c == b'#').count();
                if b.get(i + 1 + hashes) != Some(&b'"') {
                    i += 1;
                    continue;
                }
                let mut close = vec![b'"'];
                close.extend(std::iter::repeat_n(b'#', hashes));
                let body = i + 2 + hashes;
                match b[body..].windows(close.len()).position(|w| w == &close[..]) {
                    Some(end) => i = body + end + close.len(),
                    None => break,
                }
            }
            b'"' => {
                let start = i + 1;
                i = start;
                while i < b.len() && b[i] != b'"' {
                    i += if b[i] == b'\\' { 2 } else { 1 };
                }
                if i >= b.len() {
                    break;
                }
                out.push(text[start..i].to_string());
                i += 1;
            }
            _ => i += 1,
        }
    }
    out
}

/// The repository paths a test file spells as string literals and so may read: a file that names
/// `CARGO_MANIFEST_DIR`, with the tracked paths of its literals, and the script folder `tools`
/// next to an interpreter literal (`file_read_paths` of `tests/repo_scans_table.rs`).
fn file_read_paths(text: &str, tracked: &Tracked) -> Vec<String> {
    if !text.contains("CARGO_MANIFEST_DIR") {
        return Vec::new();
    }
    let literals = string_literals(text);
    let mut out: Vec<String> = literals
        .iter()
        .filter_map(|literal| tracked.repo_path(literal))
        .collect();
    if INTERPRETER_LITERALS.iter().any(|i| text.contains(i))
        && literals.iter().any(|l| l == "tools")
        && tracked.dirs.contains("tools")
    {
        out.push("tools".to_string());
    }
    out.sort();
    out.dedup();
    out
}

/// The repository paths (relative to the root, with `/`) a Rust file at `file` reads with
/// `include_str!("<literal>")` or `include_bytes!("<literal>")`, for the literals that leave
/// `tests/` (`include_str_paths` of `tests/repo_scans_table.rs`).
fn include_str_paths(file: &str, text: &str) -> Vec<String> {
    const OPENS: [&str; 2] = ["include_str!(\"", "include_bytes!(\""];
    let folder = file.rsplit_once('/').map_or("", |(dir, _)| dir);
    let mut out = Vec::new();
    for open in OPENS {
        let mut rest = text;
        while let Some(at) = rest.find(open) {
            rest = &rest[at + open.len()..];
            let Some(end) = rest.find('"') else { break };
            let literal = &rest[..end];
            rest = &rest[end..];
            if !literal.contains("..") {
                continue;
            }
            let mut parts: Vec<&str> = folder.split('/').filter(|s| !s.is_empty()).collect();
            for segment in literal.split('/') {
                match segment {
                    ".." => {
                        parts.pop();
                    }
                    "." | "" => {}
                    other => parts.push(other),
                }
            }
            let path = parts.join("/");
            if !path.starts_with("tests/") {
                out.push(path);
            }
        }
    }
    out.sort();
    out.dedup();
    out
}

/// The rows of the exemption list `const <name>` in the text of `tests/repo_scans_table.rs`: each
/// row is `(first string, strings inside a nested list)`. A list of a row that names no path is
/// empty. Comments and the text of the strings are skipped, so a commented row is no row.
fn exempt_rows(text: &str, name: &str) -> Vec<(String, Vec<String>)> {
    let Some(at) = text.find(&format!("const {name}:")) else {
        return Vec::new();
    };
    let Some(open) = text[at..].find("= &[") else {
        return Vec::new();
    };
    let b = text.as_bytes();
    let mut i = at + open + "= &[".len();
    let (mut rows, mut paren, mut bracket) = (Vec::new(), 0usize, 0usize);
    let mut row: Vec<(String, usize)> = Vec::new();
    while i < b.len() {
        match b[i] {
            b'/' if b.get(i + 1) == Some(&b'/') => {
                while i < b.len() && b[i] != b'\n' {
                    i += 1;
                }
            }
            b'/' if b.get(i + 1) == Some(&b'*') => {
                i += 2;
                while i + 1 < b.len() && !(b[i] == b'*' && b[i + 1] == b'/') {
                    i += 1;
                }
                i += 2;
            }
            b'"' => {
                let start = i + 1;
                i = start;
                while i < b.len() && b[i] != b'"' {
                    i += if b[i] == b'\\' { 2 } else { 1 };
                }
                if paren > 0 {
                    row.push((text[start..i.min(b.len())].to_string(), bracket));
                }
                i += 1;
            }
            b'(' => {
                if paren == 0 {
                    row.clear();
                    bracket = 0;
                }
                paren += 1;
                i += 1;
            }
            b')' => {
                paren = paren.saturating_sub(1);
                if paren == 0 {
                    if let Some((file, _)) = row.iter().find(|(_, d)| *d == 0) {
                        let paths = row
                            .iter()
                            .filter(|(_, d)| *d > 0)
                            .map(|(s, _)| s.clone())
                            .collect();
                        rows.push((file.clone(), paths));
                    }
                    row.clear();
                }
                i += 1;
            }
            b'[' => {
                if paren > 0 {
                    bracket += 1;
                }
                i += 1;
            }
            b']' => {
                if paren == 0 {
                    break;
                }
                bracket = bracket.saturating_sub(1);
                i += 1;
            }
            _ => i += 1,
        }
    }
    rows
}

/// One planned step. The `heavy` steps run one after another under one heavy slot hold.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Step {
    pub label: String,
    pub program: String,
    pub args: Vec<String>,
    pub heavy: bool,
    pub env: Vec<(String, String)>,
    /// An in-process step: the verdict lines, decided when the plan is made. The run writes them
    /// to the step's log and exits 1 when there are any; it starts no program.
    pub report: Option<Vec<String>>,
}

/// How a repository is verified (SPEC-20261125-check-affected-non-rust).
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Stack {
    /// A `Cargo.toml` at the top folder or a `.rs` file at the head: the cargo steps, as always.
    Rust,
    /// No Rust code, and the profile declares these checks.
    Declared(Vec<CheckEntry>),
    /// No Rust code and no declared check can be read; the text says why.
    Undeclared(String),
}

/// What the answer says when nothing could be verified: the result is `not_run`, never `PASS`.
const NOT_RUN_NEXT: &str = "declare the project's check commands as `checks` (a list of {\"id\": \"test\", \"command\": \"<one-line command>\"}) in .moai/project.yaml, commit it, and run `moai check affected` again; the `/moai project` interview saves them. Until then nothing was verified.";

/// The answer's `not_run` object, `Null` unless the stack has no declared check.
fn not_run_note(stack: &Stack) -> serde_json::Value {
    match stack {
        Stack::Undeclared(reason) => serde_json::json!({"reason": reason, "next": NOT_RUN_NEXT}),
        _ => serde_json::Value::Null,
    }
}

/// Whether `head` tracks a `.rs` file. A git that cannot answer reads as `true`, so the Rust steps
/// stay as they were.
fn tracks_rust(root: &Path, head: &str) -> bool {
    git(root, &["ls-tree", "-r", "--name-only", "-z", head])
        .map(|out| out.split('\0').any(|p| p.ends_with(".rs")))
        .unwrap_or(true)
}

/// The stack of the checkout at `root`: a `Cargo.toml` or a tracked `.rs` file makes it Rust
/// (D-1); otherwise the profile's declared `checks`, or why there are none (D-4).
fn read_stack(root: &Path, head: &str) -> Stack {
    if root.join("Cargo.toml").is_file() || tracks_rust(root, head) {
        return Stack::Rust;
    }
    match crate::profile::read_profile(root) {
        Ok((profile, _)) => match profile.checks.filter(|c| !c.is_empty()) {
            Some(checks) => Stack::Declared(checks),
            None => Stack::Undeclared("the profile declares no `checks`".to_string()),
        },
        Err(e) if e == crate::profile::PROFILE_MISSING => {
            Stack::Undeclared("the project has no profile".to_string())
        }
        Err(e) => Stack::Undeclared(format!("the profile cannot be read ({e})")),
    }
}

/// The steps of a repository without Rust code: `git diff --check`, `bash -n` of each changed
/// script and one heavy `sh -c` step per declared check, in the declared order (REQ-2).
fn other_steps(checks: &[CheckEntry], present: &[String], range: &str) -> Vec<Step> {
    let mut out = vec![step(
        "diff_check",
        "git",
        &["diff", "--check", range],
        false,
    )];
    for p in present.iter().filter(|p| p.ends_with(".sh")) {
        out.push(step(
            &format!("bash_n:{p}"),
            "bash",
            &["-n", "--", p],
            false,
        ));
    }
    for c in checks {
        out.push(step(
            &format!("check:{}", c.id),
            "sh",
            &["-c", &c.command],
            true,
        ));
    }
    out
}

/// The steps of a range for the repository's stack: [`steps`] for a Rust repository, the declared
/// checks otherwise.
pub fn plan_steps(
    facts: &Facts,
    sel: &Selection,
    range: &str,
    runner: Runner,
    threads: Option<usize>,
) -> Vec<Step> {
    match &facts.stack {
        Stack::Rust => steps(sel, &facts.present, range, runner, threads),
        Stack::Declared(checks) => other_steps(checks, &facts.present, range),
        Stack::Undeclared(_) => other_steps(&[], &facts.present, range),
    }
}

fn step(label: &str, program: &str, args: &[&str], heavy: bool) -> Step {
    Step {
        label: label.to_string(),
        program: program.to_string(),
        args: args.iter().map(|s| s.to_string()).collect(),
        heavy,
        env: Vec::new(),
        report: None,
    }
}

/// The cheap step `test_tables` (SPEC-20261124-new-test-file-precheck): the added top-level
/// `tests/*.rs` files checked against the area table and the state-folder allow list. `None`
/// when the range adds no such file or the repository has neither table, so the plan, the log
/// and the summary of every other run stay as they were.
fn table_step(facts: &Facts) -> Option<Step> {
    let added = facts
        .added
        .iter()
        .filter(|p| {
            p.strip_prefix("tests/")
                .and_then(|r| r.strip_suffix(".rs"))
                .is_some_and(|s| !s.contains('/'))
        })
        .count();
    if added == 0
        || (facts.areas_text.is_none()
            && facts.allow_text.is_none()
            && facts.scan_list.is_none()
            && facts.unreadable.is_empty())
    {
        return None;
    }
    let mut lines = table_violations(
        &facts.added,
        &facts.tests,
        facts.areas_text.as_deref(),
        facts.allow_text.as_deref(),
        &facts.moved,
    );
    if let Some(list) = facts.scan_list.as_deref() {
        if scan_entries(list).is_none() {
            lines.push(
                "tools/repo_scans.json: the table exists but cannot be read (it is not JSON with a `scans` list whose entries each have a `checks` list); fix or restore it, it is not treated as absent"
                    .to_string(),
            );
        }
    }
    lines.extend(scan_violations(
        &facts.added,
        &facts.tests,
        facts.scan_list.as_deref(),
        facts.tracked.as_deref(),
    ));
    for (table, error) in &facts.unreadable {
        lines.push(format!(
            "{table}: the table exists but cannot be read ({error}); fix or restore it, it is not treated as absent"
        ));
    }
    let mut s = step(
        "test_tables",
        "(in-process)",
        &["test_tables", &format!("{added} added test file(s)")],
        false,
    );
    s.report = Some(lines);
    Some(s)
}

/// `steps` with the table step, when there is one, placed after the last cheap step and before
/// the first heavy one.
fn plan(table: &Option<Step>, mut planned: Vec<Step>) -> Vec<Step> {
    if let Some(t) = table {
        let at = planned.iter().take_while(|s| !s.heavy).count();
        planned.insert(at, t.clone());
    }
    planned
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
    // A selected name that is a module of an area crate runs through that crate, built once
    // (SPEC-20261123-test-area-crates REQ-5). Nextest narrows the crate to the selected modules
    // with one filterset; `cargo test` has no per-binary filter and runs the whole crate.
    let mut built: BTreeSet<&str> = BTreeSet::new();
    for t in &sel.rust {
        let target = sel.areas.get(t).unwrap_or(t);
        if built.insert(target.as_str()) {
            rest.push("--test".to_string());
            rest.push(target.clone());
        }
    }
    if runner == Runner::Nextest && !sel.areas.is_empty() {
        rest.push("-E".to_string());
        rest.push(area_filterset(sel));
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

/// The nextest filterset of a selection that holds modules of area crates: the library and the
/// binary tests, every target of its own, and in each area crate only the selected modules.
fn area_filterset(sel: &Selection) -> String {
    let mut parts = vec!["kind(lib)".to_string(), "kind(bin)".to_string()];
    let mut modules: std::collections::BTreeMap<&str, Vec<&str>> = Default::default();
    for t in &sel.rust {
        match sel.areas.get(t) {
            Some(area) => modules.entry(area).or_default().push(t),
            None => parts.push(format!("binary(={t})")),
        }
    }
    for (area, names) in modules {
        let tests: Vec<String> = names.iter().map(|m| format!("test(/^{m}::/)")).collect();
        parts.push(format!("(binary(={area}) & ({}))", tests.join(" | ")));
    }
    parts.join(" | ")
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
        report: None,
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
    /// The paths the range adds (`--diff-filter=A`).
    pub added: Vec<String>,
    /// `tests/areas.toml` and the state-folder allow list, as the checkout holds them.
    pub areas_text: Option<String>,
    pub allow_text: Option<String>,
    /// A table that exists but cannot be read (no `NotFound`): its path and the read error. Such a
    /// table is not an absent one (SPEC-20261124-new-test-file-precheck REQ-10).
    pub unreadable: Vec<(String, String)>,
    /// The files git tracks, read only when the range adds a top-level test file; `None` when
    /// they were not read or cannot be (SPEC-20261124-repo-scans-precheck D-4).
    pub tracked: Option<Vec<String>>,
    /// The test folders that hold a `main.rs` (the moved areas).
    pub moved: Vec<String>,
    /// How the repository is verified: the cargo steps, or the profile's declared checks.
    pub stack: Stack,
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
    let added: Vec<String> = git(
        dir,
        &[
            "diff",
            "--name-only",
            "--no-renames",
            "--diff-filter=A",
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
    let mut moved = Vec::new();
    let mut area_modules = Vec::new();
    if let Ok(entries) = std::fs::read_dir(dir.join("tests")) {
        for e in entries.flatten() {
            let path = e.path();
            if path.join("main.rs").is_file() {
                if let Some(name) = path.file_name().and_then(|f| f.to_str()) {
                    moved.push(name.to_string());
                    area_modules.extend(read_area_modules(&path, name));
                }
            }
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
                    area: None,
                });
            }
        }
    }
    tests.extend(area_modules);
    tests.sort_by(|a, b| a.name.cmp(&b.name));
    let mut unreadable = Vec::new();
    // The scan list is `None` when it cannot be read, as it always was for the selection; the
    // table step reads the reason from `unreadable`. A check that names a module of an area crate
    // (`--test card card_cli::`) is read as the old target `card_cli`.
    let scan_list = read_table(dir, "tools/repo_scans.json", &mut unreadable)
        .map(|text| fold_area_checks(&text, &moved));
    let adds_test_file = added.iter().any(|p| {
        p.strip_prefix("tests/")
            .and_then(|r| r.strip_suffix(".rs"))
            .is_some_and(|s| !s.contains('/'))
    });
    let tracked = if adds_test_file {
        git(dir, &["ls-files", "-z"]).ok().map(|out| {
            out.split('\0')
                .filter(|p| !p.is_empty())
                .map(String::from)
                .collect()
        })
    } else {
        None
    };
    let areas_text = read_table(dir, "tests/areas.toml", &mut unreadable);
    let allow_text = read_table(
        dir,
        "tests/fixtures/state_folder_literals/allow.tsv",
        &mut unreadable,
    );
    let stack = read_stack(dir, &head);
    Ok(Facts {
        root,
        base,
        head,
        changed,
        present,
        tests,
        scan_list,
        added,
        areas_text,
        allow_text,
        unreadable,
        tracked,
        moved,
        stack,
    })
}

/// The modules of the area crate `area` (`<path>/<module>.rs`, `main.rs` and folders left out) as
/// test files that keep their old target name and name their crate. A module reaches a shared
/// helper (a module its crate's `main.rs` declares that has no file of its own in the folder, as
/// `common` and `ready_receipt` are loaded by `#[path]`) by `use crate::<helper>;` or by a path
/// such as `crate::<helper>::X`; either counts as declaring the helper (the selection of a
/// `tests/common/` change reads `modules`).
fn read_area_modules(path: &Path, area: &str) -> Vec<TestFile> {
    let mut out = Vec::new();
    let Ok(entries) = std::fs::read_dir(path) else {
        return out;
    };
    let files: Vec<std::path::PathBuf> = entries.flatten().map(|e| e.path()).collect();
    let stem_of = |f: &std::path::PathBuf| {
        f.file_name()
            .and_then(|n| n.to_str())
            .and_then(|n| n.strip_suffix(".rs"))
            .map(String::from)
    };
    let own: BTreeSet<String> = files
        .iter()
        .filter(|f| f.is_file())
        .filter_map(stem_of)
        .collect();
    let helpers: Vec<String> = std::fs::read_to_string(path.join("main.rs"))
        .map(|main| {
            declared_mods(&main)
                .into_iter()
                .filter(|m| !own.contains(m))
                .collect()
        })
        .unwrap_or_default();
    for file in files {
        let Some(name) = file.file_name().and_then(|f| f.to_str()) else {
            continue;
        };
        if !file.is_file() || !name.ends_with(".rs") || name == "main.rs" {
            continue;
        }
        let text = std::fs::read_to_string(&file).unwrap_or_default();
        let mut modules = declared_mods(&text);
        for h in &helpers {
            if uses_crate_item(&text, h) && !modules.contains(h) {
                modules.push(h.clone());
            }
        }
        out.push(TestFile {
            name: stem(name).to_string(),
            modules,
            os: os_names(&text),
            text,
            area: Some(area.to_string()),
        });
    }
    out
}

/// Whether `text` names `crate::<item>` as a whole word: `use crate::common;`, `crate::common::X`
/// and `use crate::{common, x}` count; `crate::common_x` does not.
fn uses_crate_item(text: &str, item: &str) -> bool {
    let needle = format!("crate::{item}");
    let mut rest = text;
    while let Some(i) = rest.find(&needle) {
        let after = &rest[i + needle.len()..];
        if !after
            .chars()
            .next()
            .is_some_and(|c| c.is_alphanumeric() || c == '_')
        {
            return true;
        }
        rest = after;
    }
    // A grouped import: `use crate::{a, common};`.
    text.lines().any(|l| {
        l.trim().strip_prefix("use crate::{").is_some_and(|g| {
            g.trim_end_matches(';')
                .trim_end_matches('}')
                .split(',')
                .any(|w| w.trim() == item)
        })
    })
}

/// `text` with every `--test <area> <module>::<rest>` of a moved area read as `--test <module>`:
/// the scan list names the module through its crate, the selection names it by its old target.
/// A word after `--test <name>` that is no module path, and a name that is no moved area, stay.
fn fold_area_checks(text: &str, moved: &[String]) -> String {
    let mut out = String::with_capacity(text.len());
    let mut rest = text;
    while let Some(i) = rest.find("--test ") {
        let (head, tail) = rest.split_at(i + "--test ".len());
        out.push_str(head);
        let word = |s: &str| -> usize {
            s.find(|c: char| !(c.is_ascii_alphanumeric() || c == '_'))
                .unwrap_or(s.len())
        };
        let area = &tail[..word(tail)];
        let mut after = &tail[area.len()..];
        let mut modules: Vec<&str> = Vec::new();
        if moved.iter().any(|m| m == area) {
            // Every `<module>::<rest>` word that follows the crate name is one module filter.
            while let Some(mut a) = after.strip_prefix(' ') {
                // Cargo takes one filter before `--`; the others of the command follow it.
                if let Some(past) = a.strip_prefix("-- ") {
                    a = past;
                }
                let module = &a[..word(a)];
                if module.is_empty() || !a[module.len()..].starts_with("::") {
                    break;
                }
                modules.push(module);
                let skip = a
                    .find(|c: char| c.is_whitespace() || "\"',;)".contains(c))
                    .unwrap_or(a.len());
                after = &a[skip..];
            }
        }
        if modules.is_empty() {
            rest = tail;
        } else {
            out.push_str(&modules.join(" --test "));
            rest = after;
        }
    }
    out.push_str(rest);
    out
}

/// One table's text. A missing file is an absent table (`None`, no note); any other read error
/// (invalid UTF-8, a folder, no permission) is noted in `unreadable` with the table's path, so the
/// table step stops on it instead of reading it as absent.
fn read_table(dir: &Path, rel: &str, unreadable: &mut Vec<(String, String)>) -> Option<String> {
    match std::fs::read_to_string(dir.join(rel)) {
        Ok(text) => Some(text),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => None,
        Err(e) => {
            unreadable.push((rel.to_string(), e.to_string()));
            None
        }
    }
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
    let chosen = choose(
        pick,
        &facts.changed,
        &facts.tests,
        facts.scan_list.as_deref(),
        host_os(),
    );
    let pick = chosen.pick;
    let sel = chosen.selection.clone();
    let runner = runner_in(&facts.root);
    // `--list` reads nothing from the slot folder, so it names no cap.
    let planned = plan(
        &table_step(&facts),
        plan_steps(
            &facts,
            &sel,
            &format!("{}...{}", facts.base, facts.head),
            runner,
            None,
        ),
    );
    let rust = facts.stack == Stack::Rust;
    let mut answer = serde_json::json!({
        "action": "affected",
        "mode": "list",
        "selection": pick.word(),
        "range_note": range_note(&facts, integration_ref),
        "base": facts.base,
        "head": facts.head,
        "changed": facts.changed.len(),
        "host_os": host_os(),
        "runner": rust.then(|| runner.word()),
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
    });
    card_fields(&chosen, &mut answer);
    if let Stack::Undeclared(_) = &facts.stack {
        answer["not_run"] = not_run_note(&facts.stack);
    }
    Ok(answer)
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

/// What [`refresh_changed`] did: how many files got a new time and which paths it left alone.
#[derive(Debug, Default, PartialEq, Eq)]
struct Refresh {
    touched: usize,
    skipped: Vec<String>,
}

/// SPEC-20261124-check-affected-touch REQ-1, REQ-2: set the modification time of each of `paths`
/// (relative to `root`) to now, so cargo sees a file that was restored with its old time (`cp -p`)
/// as newer than the last build. Only a regular file whose real path lies inside `root` is
/// touched: a link, a folder, a path that runs through a linked folder and a file whose time
/// cannot be set are named in `skipped`. The bytes, the mode and the git state stay as they are.
fn refresh_changed(root: &Path, paths: &[String]) -> Refresh {
    let now = std::time::SystemTime::now();
    let inside = std::fs::canonicalize(root).ok();
    let mut out = Refresh::default();
    for p in paths {
        let full = root.join(p);
        let plain = std::fs::symlink_metadata(&full).is_ok_and(|m| m.is_file());
        let within = match (&inside, std::fs::canonicalize(&full)) {
            (Some(root), Ok(real)) => real.starts_with(root),
            _ => false,
        };
        let done = plain
            && within
            && std::fs::File::open(&full)
                .and_then(|f| f.set_modified(now))
                .is_ok();
        if done {
            out.touched += 1;
        } else {
            out.skipped.push(p.clone());
        }
    }
    out
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

/// An in-process step: its verdict lines go to the log and any line makes the exit code 1.
fn run_report(lines: &[String], log: &Path) -> i32 {
    let text = if lines.is_empty() {
        "no violation\n".to_string()
    } else {
        format!("{}\n", lines.join("\n"))
    };
    let _ = std::fs::write(log, text);
    i32::from(!lines.is_empty())
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
    // The tree is clean here, so only the file times change: a restored file is newer than the
    // last build (SPEC-20261124-check-affected-touch).
    let refresh = refresh_changed(dir, &facts.present);
    let threads = DEFAULT_TEST_THREADS;
    let chosen = choose(
        opts.pick,
        &facts.changed,
        &facts.tests,
        facts.scan_list.as_deref(),
        host_os(),
    );
    let pick = chosen.pick;
    let sel = chosen.selection.clone();
    let target = target_dir(&std::fs::canonicalize(dir).unwrap_or_else(|_| dir.to_path_buf()));
    let range3 = format!("{}...{}", facts.base, facts.head);
    // The cap goes onto the test step unless the caller set the runner's own knob; what the
    // summary shows is the caller's value then.
    let env = |k: &str| std::env::var_os(k);
    let cap_for = |r: Runner| caller_threads(r, env).is_none().then_some(threads);
    let mut shown = threads.to_string();
    // Planned with nextest until the slot is held: the runner is asked only then, so a refused
    // slot starts no cargo at all. The labels do not depend on the runner.
    let table = table_step(&facts);
    let rust = facts.stack == Stack::Rust;
    let mut planned = plan(
        &table,
        plan_steps(
            &facts,
            &sel,
            &range3,
            Runner::Nextest,
            cap_for(Runner::Nextest),
        ),
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
                    planned = plan(&table, plan_steps(&facts, &sel, &range3, r, cap_for(r)));
                    if let Some(theirs) = caller_threads(r, env) {
                        shown = theirs;
                    }
                    runner = rust.then_some(r);
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
        let code = match (&s.report, s.heavy, &held, refused) {
            (Some(lines), _, _, _) => run_report(lines, &log),
            (None, true, Some((h, slots)), _) => {
                let mut extra = vec![(crate::heavy_slot::SLOT_FD_ENV, h.env_value())];
                if let Some(jobs) = build_jobs(*slots) {
                    extra.push(("CARGO_BUILD_JOBS", jobs));
                }
                run_step(dir, s, &log, &extra)
            }
            (None, true, None, Some(code)) => {
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
    // Nothing verified a repository without Rust code that declares no check: that is `not_run`,
    // unless a step failed (REQ-3, REQ-4).
    let failed = results.iter().any(|(_, c)| *c != 0);
    let verdict = match (&facts.stack, failed) {
        (_, true) => Verdict::Fail,
        (Stack::Undeclared(_), false) => Verdict::NotRun,
        _ => Verdict::Pass,
    };
    let line = summary_line_for(
        &SummaryParts {
            base: &facts.base,
            head: &facts.head,
            targets: sel.rust.len(),
            scans: &sel.scans.word(),
            pick,
            threads: &shown,
            results: &results,
            stopped_at: stopped_at.as_deref(),
            pass: verdict == Verdict::Pass,
        },
        verdict,
    );
    let mut answer = serde_json::json!({
        "action": "affected",
        "mode": "run",
        "selection": pick.word(),
        "range_note": range_note(&facts, opts.integration_ref.as_deref()),
        "summary": line,
        "result": verdict.word(),
        "stopped_at": stopped_at,
        "runner": runner.map(Runner::word),
        "refresh": {"touched": refresh.touched, "skipped": refresh.skipped},
        "steps": records,
        "out": out.to_string_lossy(),
        "writes": true,
    });
    card_fields(&chosen, &mut answer);
    if verdict == Verdict::NotRun {
        answer["not_run"] = not_run_note(&facts.stack);
    }
    Ok(answer)
}

/// What the summary line of a finished run is made of.
pub struct SummaryParts<'a> {
    pub base: &'a str,
    pub head: &'a str,
    pub targets: usize,
    pub scans: &'a str,
    pub pick: Pick,
    pub threads: &'a str,
    pub results: &'a [(String, i32)],
    pub stopped_at: Option<&'a str>,
    pub pass: bool,
}

/// The one summary line of a finished run: the receipt that `card note` and `card land-batch`
/// read. A `--lane` or `--card` run says so right after `scans=`; the bare line has no such word
/// (SPEC-20261124-ready-receipt-followup REQ-6, REQ-9).
pub fn summary_line(p: &SummaryParts<'_>) -> String {
    summary_line_for(p, if p.pass { Verdict::Pass } else { Verdict::Fail })
}

/// How a run ended: every step passed, one failed, or nothing could be verified because the
/// project declares no check (SPEC-20261125-check-affected-non-rust).
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Verdict {
    Pass,
    Fail,
    NotRun,
}

impl Verdict {
    /// The word of the answer's `result` field.
    pub fn word(self) -> &'static str {
        match self {
            Verdict::Pass => "PASS",
            Verdict::Fail => "FAIL",
            Verdict::NotRun => "not_run",
        }
    }
}

/// [`summary_line`] with the verdict given: the last word is `result=<word>`.
pub fn summary_line_for(p: &SummaryParts<'_>, verdict: Verdict) -> String {
    let selection_word = match p.pick {
        Pick::Lane => " selection=lane",
        Pick::Card => " selection=card",
        Pick::Full => "",
    };
    let mut line = format!(
        "check-affected base={} head={} targets={} scans={}{selection_word} test-threads={}",
        short(p.base),
        short(p.head),
        p.targets,
        p.scans,
        p.threads
    );
    for (k, c) in summary_words(p.results) {
        line.push_str(&format!(" {k}={c}"));
    }
    if let Some(label) = p.stopped_at {
        line.push_str(&format!(" stopped_at={label}"));
    }
    line.push_str(&format!(" result={}", verdict.word()));
    line
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

    /// SPEC-20261125-check-affected-non-rust AC-7: the last word of the summary line is the
    /// verdict; `summary_line` keeps writing `PASS` and `FAIL` as before.
    #[test]
    fn summary_line_not_run_pass_and_fail() {
        let results = vec![("diff_check".to_string(), 0)];
        let parts = |pass: bool| SummaryParts {
            base: "aaaaaaaaaaaaaaaa",
            head: "bbbbbbbbbbbbbbbb",
            targets: 0,
            scans: "none",
            pick: Pick::Full,
            threads: "4",
            results: &results,
            stopped_at: None,
            pass,
        };
        let head = "check-affected base=aaaaaaaaaaaa head=bbbbbbbbbbbb targets=0 scans=none test-threads=4 diff_check=0";
        assert_eq!(
            summary_line_for(&parts(false), Verdict::NotRun),
            format!("{head} result=not_run")
        );
        assert_eq!(summary_line(&parts(true)), format!("{head} result=PASS"));
        assert_eq!(summary_line(&parts(false)), format!("{head} result=FAIL"));
        assert_eq!(
            summary_line_for(&parts(true), Verdict::Pass),
            summary_line(&parts(true))
        );
    }

    /// SPEC-20261125-check-affected-non-rust REQ-2: a declared check is one heavy `sh -c` step
    /// after the cheap ones, in the declared order; no declared check adds none.
    #[test]
    fn declared_checks_follow_the_cheap_steps_in_their_order() {
        let check = |id: &str, command: &str| CheckEntry {
            id: id.to_string(),
            command: command.to_string(),
        };
        let planned = other_steps(
            &[check("lint", "npm run lint"), check("test", "npm test")],
            &["bin/run.sh".to_string(), "src/a.ts".to_string()],
            "a...b",
        );
        let labels: Vec<&str> = planned.iter().map(|s| s.label.as_str()).collect();
        assert_eq!(
            labels,
            [
                "diff_check",
                "bash_n:bin/run.sh",
                "check:lint",
                "check:test"
            ]
        );
        assert_eq!(planned[2].program, "sh");
        assert_eq!(planned[2].args, ["-c", "npm run lint"]);
        assert!(planned[2].heavy && planned[3].heavy && !planned[0].heavy && !planned[1].heavy);
        assert_eq!(other_steps(&[], &[], "a...b").len(), 1);
    }

    const AREAS: &str = "# t\n[areas]\ncard = [\"card\"]\nspec = [\"spec\", \"check\"]\nlane = [\"lane\"]\n\n[solo]\nfiles = [\"check_solo\"]\n";

    fn file(name: &str, text: &str) -> TestFile {
        TestFile {
            name: name.to_string(),
            text: text.to_string(),
            ..Default::default()
        }
    }

    fn added(stems: &[&str]) -> Vec<String> {
        stems.iter().map(|s| format!("tests/{s}.rs")).collect()
    }

    #[test]
    fn table_violations_name_each_kind_of_miss() {
        let moved = vec!["lane".to_string()];
        let got = table_violations(
            &added(&["zzfresh_cli", "lane_x", "spec", "check_solo", "card_ok"]),
            &[],
            Some(AREAS),
            None,
            &moved,
        );
        assert_eq!(got.len(), 3, "{got:?}");
        assert!(got[0].starts_with("tests/zzfresh_cli.rs") && got[0].contains("\"zzfresh\""));
        assert!(got[0].contains("tests/areas.toml") && got[0].contains("card, spec, lane"));
        assert!(got[1].contains("moved") && got[1].contains("tests/lane/lane_x.rs"));
        assert!(got[2].contains("area crate `spec`"));
    }

    #[test]
    fn table_violations_pass_a_registered_file_and_skip_an_absent_table() {
        let text = format!("let p = \"{}/x\";", earlier_state_name());
        let tests = [file("zzfresh_cli", &text)];
        let names = added(&["zzfresh_cli"]);
        assert!(table_violations(&names, &tests, None, None, &[]).is_empty());
        assert!(table_violations(&added(&["card_ok"]), &[], Some(AREAS), None, &[]).is_empty());
        // Files in a folder, and paths that are no test file, are not top-level tests.
        let other = vec![
            "tests/lane/zz.rs".to_string(),
            "tests/zz.py".to_string(),
            "src/zz.rs".to_string(),
        ];
        assert!(table_violations(&other, &[], Some(AREAS), Some(""), &[]).is_empty());
    }

    #[test]
    fn table_violations_count_each_state_folder_name_against_its_row() {
        let both = format!(
            "\"{}/a\" \"{}/b\" \"{}/c\"",
            earlier_state_name(),
            new_state_name(),
            new_state_name()
        );
        let tests = [file("check_a", &both)];
        let names = added(&["check_a"]);
        let none = table_violations(&names, &tests, None, Some("# only a comment\n"), &[]);
        assert_eq!(none.len(), 1, "{none:?}");
        assert!(none[0].contains("`tests/check_a.rs\t1\t2`"), "{none:?}");
        let exact = "tests/check_a.rs\t1\t2\n";
        assert!(table_violations(&names, &tests, None, Some(exact), &[]).is_empty());
        let roomy = "tests/check_a.rs\t5\t5\n";
        assert!(table_violations(&names, &tests, None, Some(roomy), &[]).is_empty());
        let low = "tests/check_a.rs\t1\t1\n";
        let got = table_violations(&names, &tests, None, Some(low), &[]);
        assert_eq!(got.len(), 1, "{got:?}");
        assert!(
            got[0].contains(" 2 times") && got[0].contains("allows 1"),
            "{got:?}"
        );
        // The earlier name never counts as the new one, and program names count nothing.
        let text = format!("{}-v1 moai-cli x{}", earlier_state_name(), new_state_name());
        assert_eq!(count_state_name(&text, &new_state_name()), 0);
        assert_eq!(count_state_name(&text, &earlier_state_name()), 0);
    }

    /// Audit N1 of SPEC-20261123-test-area-crates: a module added to a moved area is a test file
    /// the state-folder allow list counts by its longer path, so the step reads it too.
    #[test]
    fn table_violations_read_a_module_added_to_a_moved_area_by_its_longer_path() {
        let text = format!("\"{}/a\" \"{}/b\"", earlier_state_name(), new_state_name());
        let mut module = file("lane_x", &text);
        module.area = Some("lane".to_string());
        let tests = [module];
        let added = vec!["tests/lane/lane_x.rs".to_string()];
        let moved = vec!["lane".to_string()];
        let none = table_violations(&added, &tests, None, Some("# none\n"), &moved);
        assert_eq!(none.len(), 1, "{none:?}");
        assert!(none[0].contains("`tests/lane/lane_x.rs\t1\t1`"), "{none:?}");
        let row = "tests/lane/lane_x.rs\t1\t1\n";
        assert!(table_violations(&added, &tests, None, Some(row), &moved).is_empty());
        // An area that is not moved, and the crate's own main.rs, are no module.
        let other = vec![
            "tests/lane/lane_x.rs".to_string(),
            "tests/lane/main.rs".to_string(),
        ];
        assert!(table_violations(&other, &tests, None, Some("# none\n"), &[]).is_empty());
        assert!(table_violations(&other[1..], &tests, None, Some("# none\n"), &moved).is_empty());
    }

    const MANIFEST: &str = env!("CARGO_MANIFEST_DIR");

    /// Audit B1 on the real files: every module of `tests/card/` that names `crate::common` in
    /// any form is read as a user of it, and a change of `tests/common/mod.rs` selects exactly
    /// those modules. The six that name it by a path alone (`close_check_cli`, `startable_shared`,
    /// `release_suite`, `ledger_forge_only_cli`, `close_step_2903`, `decision_file_cli`) are the
    /// ones the first version of the read missed.
    #[test]
    fn a_change_of_the_common_helper_selects_every_real_card_module_that_names_it() {
        let dir = Path::new(MANIFEST).join("tests/card");
        let tests = read_area_modules(&dir, "card");
        assert!(tests.len() > 100, "{} modules read", tests.len());
        let naming: BTreeSet<&str> = tests
            .iter()
            .filter(|t| t.text.contains("crate::common::") || t.text.contains("use crate::common;"))
            .map(|t| t.name.as_str())
            .collect();
        for six in [
            "close_check_cli",
            "startable_shared",
            "release_suite",
            "ledger_forge_only_cli",
            "close_step_2903",
            "decision_file_cli",
        ] {
            assert!(naming.contains(six), "{six} no longer names the helper");
        }
        let sel = select(&paths(&["tests/common/mod.rs"]), &tests, None, "linux");
        let picked: BTreeSet<&str> = sel.rust.iter().map(String::as_str).collect();
        let missed: Vec<&&str> = naming.difference(&picked).collect();
        assert!(missed.is_empty(), "not selected: {missed:?}");
    }

    /// SPEC-20261124-new-test-file-precheck REQ-8: the repository's own test files, all taken as
    /// added, pass the step: it cannot be stricter than the two repository tests, which pass on
    /// the same tree. A unit test, not a `tests/` file: a test file that walks the tree needs a
    /// row in `tools/repo_scans.json`.
    #[test]
    fn the_repository_tests_pass_the_table_step_as_added_files() {
        let root = Path::new(MANIFEST);
        let mut added = Vec::new();
        let mut tests = Vec::new();
        for entry in std::fs::read_dir(root.join("tests")).unwrap().flatten() {
            let path = entry.path();
            if !path.is_file() || path.extension().is_none_or(|e| e != "rs") {
                continue;
            }
            let name = path.file_stem().unwrap().to_string_lossy().to_string();
            added.push(format!("tests/{name}.rs"));
            tests.push(TestFile {
                text: std::fs::read_to_string(&path).unwrap(),
                name,
                ..Default::default()
            });
        }
        assert!(added.len() > 100, "the scan found {} files", added.len());
        let areas = std::fs::read_to_string(root.join("tests/areas.toml")).unwrap();
        let allow =
            std::fs::read_to_string(root.join("tests/fixtures/state_folder_literals/allow.tsv"))
                .unwrap();
        let moved: Vec<String> = std::fs::read_dir(root.join("tests"))
            .unwrap()
            .flatten()
            .filter(|e| e.path().join("main.rs").is_file())
            .map(|e| e.file_name().to_string_lossy().to_string())
            .collect();
        let found = table_violations(&added, &tests, Some(&areas), Some(&allow), &moved);
        assert!(found.is_empty(), "{}", found.join("\n"));
    }

    /// REQ-9: the guidance says where a new test file registers and that the step stops the run.
    #[test]
    fn the_guidance_says_where_a_new_test_file_registers() {
        let root = Path::new(MANIFEST);
        for rel in [
            "docs/design-review/appendix/operations/check.md",
            "docs/lane-rules.md",
        ] {
            let text = std::fs::read_to_string(root.join(rel)).unwrap();
            for word in [
                "test_tables",
                "tests/areas.toml",
                "allow.tsv",
                "tools/repo_scans.json",
                "EXEMPT",
            ] {
                assert!(text.contains(word), "{rel} does not name `{word}`");
            }
        }
    }

    const SCAN_LIST: &str = r#"{"scans":[
      {"id":"walk-listed","tests":["tests/zz_listed_walk.rs::t"],
       "checks":["cargo test --offline --locked --test zz_listed_walk"],
       "rules":["tests/zz_listed_walk.rs"],"scope":[{"prefix":"docs/"}]},
      {"id":"reads-docs","always":true,"tests":["tests/zz_ok_reader.rs::t"],
       "checks":["cargo test --offline --locked --test zz_ok_reader"],
       "rules":["tests/zz_ok_reader.rs"],"scope":[{"prefix":"docs/","suffixes":[".md"]}]},
      {"id":"reads-nocheck","always":true,"tests":["tests/zz_nocheck.rs::t"],
       "checks":["cargo test --offline --locked --test other"],
       "rules":["tests/zz_nocheck.rs"],"scope":[{"prefix":"docs/"}]}
    ]}"#;

    /// The three exemption lists as `tests/repo_scans_table.rs` spells them, with a comment line
    /// that looks like a row and a reason that holds escaped quotes.
    const EXEMPT_TEXT: &str = "\
// (\"tests/zz_comment.rs\", \"not a row\")
const EXEMPT: &[(&str, &str)] = &[
    // (\"tests/zz_comment.rs\", \"not a row either\"),
    (\"tests/zz_walk_exempt.rs\", \"walks a \\\"temporary\\\" folder\"),
];
const INCLUDE_STR_EXEMPT: &[(&str, &str)] = &[
    (\"tests/zz_inc_exempt.rs\", \"reads a fixture\"),
];
const FILE_READ_EXEMPT: &[(&str, &[&str], &str)] = &[
    (
        \"tests/zz_read_exempt.rs\",
        &[\"docs/a.md\", \"src/lib.rs\"],
        \"writes them into the temporary project\",
    ),
    (\"tests/zz_read_whole.rs\", &[], \"names paths of the project it builds\"),
];
";

    fn tracked() -> Vec<String> {
        paths(&["docs/a.md", "docs/a.txt", "src/lib.rs"])
    }

    /// A test file that names the repository root and carries `body`.
    fn rooted(body: &str) -> String {
        format!("const M: &str = env!(\"CARGO_MANIFEST_DIR\");\n{body}\n")
    }

    fn violations(
        files: &[(&str, String)],
        list: Option<&str>,
        tracked: Option<&[String]>,
    ) -> Vec<String> {
        let mut tests: Vec<TestFile> = files.iter().map(|(n, t)| file(n, t)).collect();
        tests.push(file("repo_scans_table", EXEMPT_TEXT));
        let names: Vec<&str> = files.iter().map(|(n, _)| *n).collect();
        scan_violations(&added(&names), &tests, list, tracked)
    }

    fn line_for<'a>(got: &'a [String], name: &str) -> &'a str {
        got.iter()
            .find(|l| l.starts_with(&format!("tests/{name}.rs")))
            .unwrap_or_else(|| panic!("no line for {name}: {got:?}"))
    }

    #[test]
    fn scan_violations_name_each_kind_of_miss() {
        let files = [
            (
                "zz_walk",
                rooted("fn t() { let _ = std::fs::read_dir(\".\"); }"),
            ),
            (
                "zz_inc",
                "const A: &str = include_str!(\"../src/zz_new_table.rs\");".to_string(),
            ),
            (
                "zz_incb",
                "const A: &[u8] = include_bytes!(\"../docs/a.md\");".to_string(),
            ),
            ("zz_read", rooted("fn t() { let _ = \"docs/a.md\"; }")),
            // The scan names this file, but its scope covers only `.md` files.
            ("zz_ok_reader", rooted("fn t() { let _ = \"docs/a.txt\"; }")),
            (
                "zz_listed_walk",
                rooted("fn t() { let _ = \"docs/a.md\"; }"),
            ),
            ("zz_plain", "fn t() {}".to_string()),
        ];
        let got = violations(&files, Some(SCAN_LIST), Some(&tracked()));
        let walk = line_for(&got, "zz_walk");
        assert!(
            walk.contains("tools/repo_scans.json") && walk.contains("EXEMPT"),
            "{walk}"
        );
        let inc = line_for(&got, "zz_inc");
        assert!(
            inc.contains("src/zz_new_table.rs") && inc.contains("INCLUDE_STR_EXEMPT"),
            "{inc}"
        );
        let incb = line_for(&got, "zz_incb");
        assert!(
            incb.contains("docs/a.md") && incb.contains("include_bytes!"),
            "{incb}"
        );
        let read = line_for(&got, "zz_read");
        assert!(
            read.contains("docs/a.md") && read.contains("FILE_READ_EXEMPT"),
            "{read}"
        );
        let suffix = line_for(&got, "zz_ok_reader");
        assert!(suffix.contains("docs/a.txt"), "{suffix}");
        // A scan without `always` claims no coverage, whatever it names and holds in its scope.
        let plain_scan = line_for(&got, "zz_listed_walk");
        assert!(plain_scan.contains("docs/a.md"), "{plain_scan}");
        // One line per miss and none for the plain file.
        assert_eq!(got.len(), 6, "{got:?}");
    }

    #[test]
    fn scan_violations_pass_a_registered_or_exempt_file_and_skip_what_is_absent() {
        let files = [
            ("zz_listed_walk", rooted("fn t() { let _ = std::fs::read_dir(\".\"); }")),
            ("zz_ok_reader", rooted("fn t() { let _ = \"docs/a.md\"; }")),
            ("zz_walk_exempt", rooted("fn t() { let _ = std::fs::read_dir(\".\"); }")),
            (
                "zz_inc_exempt",
                "const A: &str = include_str!(\"../src/zz_new_table.rs\");".to_string(),
            ),
            ("zz_read_exempt", rooted("fn t() { let _ = \"docs/a.md\"; let _ = \"src/lib.rs\"; }")),
            ("zz_read_whole", rooted("fn t() { let _ = \"docs/a.md\"; let _ = \"src/lib.rs\"; }")),
            // `tests/` is no file outside `tests/`, and a literal that is no tracked path is no read.
            (
                "zz_inside",
                rooted("const A: &str = include_str!(\"fixtures/x.txt\"); fn t() { let _ = \"docs/none.md\"; }"),
            ),
        ];
        let got = violations(&files, Some(SCAN_LIST), Some(&tracked()));
        assert!(got.is_empty(), "{got:?}");
        // An absent scan list is skipped as a whole (the repository has no such table).
        let bare = [(
            "zz_walk",
            rooted("fn t() { let _ = std::fs::read_dir(\".\"); }"),
        )];
        assert!(violations(&bare, None, Some(&tracked())).is_empty());
        // Folders, other suffixes and paths outside `tests/` are no top-level test files.
        let other = paths(&["tests/lane/zz.rs", "tests/zz.py", "src/zz.rs"]);
        let tests = [file(
            "zz",
            &rooted("fn t() { let _ = std::fs::read_dir(\".\"); }"),
        )];
        assert!(scan_violations(&other, &tests, Some(SCAN_LIST), Some(&tracked())).is_empty());
    }

    #[test]
    fn scan_violations_read_the_three_exemption_lists() {
        // Each list keeps out only its own rule: a walking file is not kept out by the reader
        // lists, a reader of a path the row does not name is not kept out, and a comment is no row.
        let files = [
            (
                "zz_comment",
                rooted("fn t() { let _ = std::fs::read_dir(\".\"); }"),
            ),
            (
                "zz_inc_exempt_wrong",
                "const A: &str = include_str!(\"../docs/a.md\");".to_string(),
            ),
            (
                "zz_read_exempt_more",
                rooted("fn t() { let _ = \"docs/a.md\"; }"),
            ),
            // Listed in `EXEMPT` only: reading is not walking.
            (
                "zz_walk_exempt",
                rooted("fn t() { let _ = \"docs/a.md\"; }"),
            ),
            // Listed in `FILE_READ_EXEMPT` only: walking is not reading.
            (
                "zz_read_exempt",
                rooted("fn t() { let _ = std::fs::read_dir(\".\"); }"),
            ),
        ];
        let got = violations(&files, Some(SCAN_LIST), Some(&tracked()));
        assert_eq!(got.len(), 5, "{got:?}");
        assert!(line_for(&got, "zz_comment").contains("EXEMPT"));
        // A row of `FILE_READ_EXEMPT` names its paths: `tests/zz_read_exempt.rs` is kept out for
        // `docs/a.md` and `src/lib.rs` only.
        let more = [(
            "zz_read_exempt",
            rooted("fn t() { let _ = \"docs/a.md\"; let _ = \"docs/a.txt\"; }"),
        )];
        let got = violations(&more, Some(SCAN_LIST), Some(&tracked()));
        assert_eq!(got.len(), 1, "{got:?}");
        assert!(got[0].contains("docs/a.txt"), "{got:?}");
        // No exemption text at all: every walking file is a violation.
        let tests = [file(
            "zz_walk_exempt",
            &rooted("fn t() { let _ = std::fs::read_dir(\".\"); }"),
        )];
        let got = scan_violations(
            &added(&["zz_walk_exempt"]),
            &tests,
            Some(SCAN_LIST),
            Some(&tracked()),
        );
        assert_eq!(got.len(), 1, "{got:?}");
    }

    #[test]
    fn scan_violations_name_a_rule_file_without_its_test_check() {
        let files = [
            ("zz_nocheck", rooted("fn t() { let _ = \"docs/a.md\"; }")),
            ("zz_ok_reader", "fn t() {}".to_string()),
        ];
        let got = violations(&files, Some(SCAN_LIST), Some(&tracked()));
        // The line starts with the scan, as the line of `tests/repo_scans_table.rs` does.
        let line = got
            .iter()
            .find(|l| l.starts_with("reads-nocheck: tests/zz_nocheck.rs"))
            .unwrap_or_else(|| panic!("no line for the scan: {got:?}"));
        assert!(line.contains("--test zz_nocheck"), "{line}");
        // The scan that names `zz_ok_reader` runs it: no line. `zz_nocheck` is a reader the scan
        // names, so it has no coverage line either, only the missing check.
        assert_eq!(got.len(), 1, "{got:?}");
    }

    #[test]
    fn scan_violations_say_when_the_tracked_files_are_unknown() {
        let files = [("zz_read", rooted("fn t() { let _ = \"docs/a.md\"; }"))];
        let got = violations(&files, Some(SCAN_LIST), None);
        assert_eq!(got.len(), 1, "{got:?}");
        assert!(
            got[0].contains("tests/zz_read.rs") && got[0].contains("git ls-files"),
            "{got:?}"
        );
        // Nothing that spells the root needs the tracked files.
        let plain = [("zz_plain", "fn t() {}".to_string())];
        assert!(violations(&plain, Some(SCAN_LIST), None).is_empty());
    }

    /// SPEC REQ-8: the repository's own test files, all taken as added, pass the scan step: it
    /// cannot be stricter than `tests/repo_scans_table.rs`, which passes on the same tree.
    #[test]
    fn the_repository_tests_pass_the_scan_step_as_added_files() {
        let root = Path::new(MANIFEST);
        let mut added = Vec::new();
        let mut tests = Vec::new();
        for entry in std::fs::read_dir(root.join("tests")).unwrap().flatten() {
            let path = entry.path();
            if !path.is_file() || path.extension().is_none_or(|e| e != "rs") {
                continue;
            }
            let name = path.file_stem().unwrap().to_string_lossy().to_string();
            added.push(format!("tests/{name}.rs"));
            tests.push(TestFile {
                text: std::fs::read_to_string(&path).unwrap(),
                name,
                ..Default::default()
            });
        }
        assert!(added.len() > 100, "the scan found {} files", added.len());
        let list = std::fs::read_to_string(root.join("tools/repo_scans.json")).unwrap();
        let listed = git(root, &["ls-files", "-z"]).expect("git ls-files");
        let tracked: Vec<String> = listed
            .split('\0')
            .filter(|p| !p.is_empty())
            .map(String::from)
            .collect();
        let found = scan_violations(&added, &tests, Some(&list), Some(&tracked));
        assert!(found.is_empty(), "{}", found.join("\n"));
    }

    fn rust(name: &str, modules: &[&str], os: &[&str]) -> TestFile {
        TestFile {
            name: name.to_string(),
            modules: modules.iter().map(|s| s.to_string()).collect(),
            os: os.iter().map(|s| s.to_string()).collect(),
            text: String::new(),
            area: None,
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

    /// C-456: the managed-block text of `src/agents.rs` is pinned by two tests whose names do not
    /// start with `agents`, so the area rule alone misses them (the red tip of 2026-11-24). The
    /// repository's own scan list picks them for that file, in the lane set and in the full set,
    /// and the lane set stays narrow for a change that cannot move the text.
    #[test]
    fn a_change_to_the_managed_block_source_selects_the_two_tests_that_pin_its_text() {
        let list = std::fs::read_to_string(Path::new(MANIFEST).join("tools/repo_scans.json"))
            .expect("the repository's scan list");
        let tests = [
            rust("agents_home_cli", &[], &[]),
            rust("language_setting", &[], &[]),
            rust("language_three_questions", &[], &[]),
            rust("card_cli", &[], &[]),
        ];
        for pick in [Pick::Lane, Pick::Full] {
            let sel = select_pick(
                pick,
                &paths(&["src/agents.rs"]),
                &tests,
                Some(&list),
                "linux",
            );
            for name in ["language_setting", "language_three_questions"] {
                assert!(sel.rust.contains(name), "{pick:?} missed {name}: {sel:?}");
            }
            assert!(sel.rust.contains("agents_home_cli"), "{pick:?}: {sel:?}");
        }
        let other = select_pick(
            Pick::Lane,
            &paths(&["src/card.rs"]),
            &tests,
            Some(&list),
            "linux",
        );
        assert!(!other.rust.contains("language_setting"), "{other:?}");
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

    /// A module of the area crate `card`, the way `read_area_modules` reads one.
    fn module(name: &str, text: &str) -> TestFile {
        TestFile {
            name: name.to_string(),
            modules: vec!["common".to_string()],
            text: text.to_string(),
            area: Some("card".to_string()),
            ..Default::default()
        }
    }

    /// SPEC-20261123-test-area-crates REQ-5: the old names of a moved area run through the one
    /// crate, and nextest narrows it to the selected modules while the other targets stay whole.
    #[test]
    fn area_modules_run_through_their_crate_with_a_module_filter() {
        let tests = [
            module("card_cli", ""),
            module("ledger_open", ""),
            module("ledger_schema", ""),
            rust("update_command_cli", &[], &[]),
        ];
        let changed = paths(&["src/ledger/open.rs"]);
        let sel = select(&changed, &tests, None, "linux");
        let want: BTreeSet<String> = ["ledger_open", "ledger_schema"].map(String::from).into();
        assert_eq!(sel.rust, want, "the selection stays by old name");
        let planned = steps(&sel, &changed, "a...b", Runner::Nextest, None);
        let test = planned.iter().find(|s| s.label == "test").unwrap();
        let targets: Vec<&str> = test
            .args
            .windows(2)
            .filter(|w| w[0] == "--test")
            .map(|w| w[1].as_str())
            .collect();
        assert_eq!(targets, ["card"], "{:?}", test.args);
        let at = test.args.iter().position(|a| a == "-E").unwrap();
        assert_eq!(
            test.args[at + 1],
            "kind(lib) | kind(bin) | (binary(=card) & (test(/^ledger_open::/) | test(/^ledger_schema::/)))"
        );
    }

    #[test]
    fn a_target_of_its_own_stays_whole_beside_an_area_crate() {
        let tests = [module("card_cli", ""), rust("card_solo", &[], &[])];
        let changed = paths(&["src/card/x.rs"]);
        let sel = select(&changed, &tests, None, "linux");
        let planned = steps(&sel, &changed, "a...b", Runner::Nextest, None);
        let test = planned.iter().find(|s| s.label == "test").unwrap();
        let at = test.args.iter().position(|a| a == "-E").unwrap();
        assert_eq!(
            test.args[at + 1],
            "kind(lib) | kind(bin) | binary(=card_solo) | (binary(=card) & (test(/^card_cli::/)))"
        );
        let runs: Vec<&str> = test
            .args
            .windows(2)
            .filter(|w| w[0] == "--test")
            .map(|w| w[1].as_str())
            .collect();
        assert_eq!(runs, ["card", "card_solo"], "{:?}", test.args);
    }

    #[test]
    fn libtest_has_no_per_binary_filter_and_runs_the_whole_crate() {
        let tests = [module("card_cli", "")];
        let changed = paths(&["src/card/x.rs"]);
        let sel = select(&changed, &tests, None, "linux");
        let planned = steps(&sel, &changed, "a...b", Runner::CargoTest, None);
        let test = planned.iter().find(|s| s.label == "test").unwrap();
        assert!(test.args.windows(2).any(|w| w == ["--test", "card"]));
        assert!(!test.args.iter().any(|a| a == "-E"), "{:?}", test.args);
    }

    #[test]
    fn a_selection_without_an_area_module_adds_no_filterset() {
        let tests = [rust("update_command_cli", &[], &[])];
        let changed = paths(&["src/update_models.rs"]);
        let sel = select(&changed, &tests, None, "linux");
        assert!(sel.areas.is_empty());
        let planned = steps(&sel, &changed, "a...b", Runner::Nextest, None);
        let test = planned.iter().find(|s| s.label == "test").unwrap();
        assert!(!test.args.iter().any(|a| a == "-E"), "{:?}", test.args);
    }

    #[test]
    fn a_changed_module_file_and_a_changed_crate_root_select_by_old_name() {
        let tests = [module("card_cli", ""), module("ledger_open", "")];
        let one = select(&paths(&["tests/card/card_cli.rs"]), &tests, None, "linux");
        assert_eq!(one.rust, ["card_cli"].map(String::from).into());
        assert_eq!(one.areas["card_cli"], "card");
        let all = select(&paths(&["tests/card/main.rs"]), &tests, None, "linux");
        assert_eq!(all.rust.len(), 2);
        let none = select(&paths(&["tests/card/notes.txt"]), &tests, None, "linux");
        assert!(none.rust.is_empty(), "{:?}", none.rust);
    }

    /// Audit B2: the card pick of a changed test module (`tests/card/<module>.rs`) names the
    /// area crate like the full and lane picks do, so the step builds `--test card`, not a
    /// target `card_cli` that no longer exists.
    #[test]
    fn the_card_pick_of_an_area_module_names_its_crate() {
        let tests = [module("card_cli", "fn a() {}"), rust("solo_cli", &[], &[])];
        let changed = paths(&["tests/card/card_cli.rs"]);
        let pick = select_card(&changed, &tests, None, "linux");
        assert!(pick.wide.is_none(), "{:?}", pick.wide);
        assert!(pick.selection.rust.contains("card_cli"));
        assert_eq!(pick.selection.areas["card_cli"], "card");
        let planned = steps(&pick.selection, &changed, "a..b", Runner::Nextest, None);
        let test = planned.iter().find(|s| s.label == "test").unwrap();
        assert!(
            test.args.windows(2).any(|w| w == ["--test", "card"]),
            "{:?}",
            test.args
        );
        assert!(
            !test.args.iter().any(|a| a == "card_cli"),
            "{:?}",
            test.args
        );
    }

    /// The `main.rs` of an area crate: two helpers loaded by `#[path]` and one module of its own.
    const AREA_MAIN: &str = "#[path = \"../common/mod.rs\"]\nmod common;\n#[path = \"../common/ready_receipt.rs\"]\nmod ready_receipt;\n\nmod card_cli;\n";

    /// Audit B1 of SPEC-20261123-test-area-crates: six moved modules reach the helper by a path
    /// (`crate::common::MOAI_BIN`) and carry no `use crate::common;` line. The selection of a
    /// `tests/common/` change reads the files as the checkout holds them, so each way of naming
    /// the helper is read here from real files, and the selection is made from what was read.
    #[test]
    fn a_helper_named_by_a_path_alone_is_read_from_the_file_and_selected() {
        let root = tempfile::tempdir().unwrap();
        let dir = root.path().join("card");
        std::fs::create_dir_all(&dir).unwrap();
        std::fs::write(dir.join("main.rs"), AREA_MAIN).unwrap();
        for (name, text) in [
            ("by_use", "use crate::common;\nfn a() {}\n"),
            ("by_path", "fn b() { let _ = crate::common::MOAI_BIN; }\n"),
            ("by_group", "use crate::{common, ready_receipt};\n"),
            (
                "by_other_helper",
                "fn c() { crate::ready_receipt::run(); }\n",
            ),
            // A longer name and a sibling module are no helper use.
            ("by_longer_name", "fn d() { crate::common_x::y(); }\n"),
            ("by_sibling", "fn e() { crate::card_cli::f(); }\n"),
            ("by_nothing", "fn g() {}\n"),
        ] {
            std::fs::write(dir.join(format!("{name}.rs")), text).unwrap();
        }
        let tests = read_area_modules(&dir, "card");
        let used = |name: &str, helper: &str| {
            tests
                .iter()
                .find(|t| t.name == name)
                .unwrap_or_else(|| panic!("{name} not read"))
                .modules
                .iter()
                .any(|m| m == helper)
        };
        assert!(used("by_use", "common"));
        assert!(used("by_path", "common"), "a path use is a helper use");
        assert!(used("by_group", "common") && used("by_group", "ready_receipt"));
        assert!(used("by_other_helper", "ready_receipt"));
        assert!(!used("by_other_helper", "common"));
        for none in ["by_longer_name", "by_sibling", "by_nothing"] {
            assert!(
                !used(none, "common") && !used(none, "ready_receipt"),
                "{none}"
            );
        }
        // The change of the helper reaches exactly the readers of it, in both selections.
        for pick in [Pick::Full, Pick::Lane] {
            let sel = select_pick(
                pick,
                &paths(&["tests/common/mod.rs"]),
                &tests,
                None,
                "linux",
            );
            for name in ["by_use", "by_path", "by_group"] {
                assert!(sel.rust.contains(name), "{pick:?} {name}: {:?}", sel.rust);
            }
            for name in ["by_longer_name", "by_sibling", "by_nothing"] {
                assert!(!sel.rust.contains(name), "{pick:?} {name}: {:?}", sel.rust);
            }
        }
    }

    #[test]
    fn a_helper_change_still_reaches_the_modules_that_use_it() {
        let tests = [module("card_cli", "use crate::common::git;")];
        let sel = select(&paths(&["tests/common/mod.rs"]), &tests, None, "linux");
        assert!(sel.rust.contains("card_cli"), "{:?}", sel.rust);
    }

    /// An old name in the scan list and a module filter through the crate are the same target.
    #[test]
    fn a_scan_check_through_an_area_crate_folds_to_the_old_name() {
        let moved = vec!["card".to_string()];
        let text = r#"{"checks":["cargo test --offline --locked --test card card_land_scans_cli::a","cargo test --test product_cli skill_package::","cargo test --test card_solo","cargo test --test lane lane_x::"]}"#;
        let got = fold_area_checks(text, &moved);
        assert_eq!(
            got,
            r#"{"checks":["cargo test --offline --locked --test card_land_scans_cli","cargo test --test product_cli skill_package::","cargo test --test card_solo","cargo test --test lane lane_x::"]}"#
        );
        assert_eq!(fold_area_checks(text, &[]), text);
        // Several module filters of one crate are several old targets.
        let many =
            r#"["cargo test --offline --locked --test card a_cli:: b_cli::x c_cli:: --test d"]"#;
        assert_eq!(
            fold_area_checks(many, &moved),
            r#"["cargo test --offline --locked --test a_cli --test b_cli --test c_cli --test d"]"#
        );
        // The checked-in form: one filter before `--`, the others after it.
        let dashed =
            r#"["cargo test --offline --locked --test card a_cli:: -- b_cli::x c_cli:: --test d"]"#;
        assert_eq!(
            fold_area_checks(dashed, &moved),
            r#"["cargo test --offline --locked --test a_cli --test b_cli --test c_cli --test d"]"#
        );
    }

    #[test]
    fn area_modules_are_read_with_their_helpers_and_old_names() {
        let root = tempfile::tempdir().unwrap();
        let dir = root.path().join("card");
        std::fs::create_dir_all(&dir).unwrap();
        std::fs::write(dir.join("main.rs"), AREA_MAIN).unwrap();
        std::fs::write(
            dir.join("card_cli.rs"),
            "use crate::common;\nuse crate::ready_receipt;\n#[cfg(target_os = \"linux\")]\nfn x() {}\n",
        )
        .unwrap();
        std::fs::write(dir.join("notes.txt"), "x").unwrap();
        let got = read_area_modules(&dir, "card");
        assert_eq!(got.len(), 1, "{got:?}");
        assert_eq!(got[0].name, "card_cli");
        assert_eq!(got[0].area.as_deref(), Some("card"));
        assert_eq!(got[0].modules, ["common", "ready_receipt"]);
        assert_eq!(got[0].os, ["linux"]);
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
