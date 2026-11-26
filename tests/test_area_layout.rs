//! Layout rules of the integration test areas (SPEC-20261123-test-area-crates).
//!
//! `tests/areas.toml` maps the first word of every test file name to one of eight areas. An area
//! is moved once `tests/<area>/main.rs` exists; its files then live as modules of that one crate.
//! Each scan takes the repository root, so a planted temporary tree can prove every failure and the
//! real tree proves the repository passes. With no area moved yet, the scans pass on the real tree.

use std::collections::{BTreeMap, BTreeSet};
use std::fs;
use std::path::Path;

const MANIFEST: &str = env!("CARGO_MANIFEST_DIR");

const AREAS: [&str; 8] = [
    "card", "update", "home", "project", "lane", "screen", "spec", "repo",
];

struct Table {
    areas: BTreeMap<String, Vec<String>>,
    solo: Vec<String>,
}

fn strings(array: &toml_edit::Array, what: &str) -> Result<Vec<String>, String> {
    array
        .iter()
        .map(|value| {
            value
                .as_str()
                .map(str::to_string)
                .ok_or_else(|| format!("{what}: an entry is not a string"))
        })
        .collect()
}

fn parse_table(text: &str) -> Result<Table, String> {
    let doc = text
        .parse::<toml_edit::DocumentMut>()
        .map_err(|e| format!("areas.toml is not valid TOML: {e}"))?;
    for line in text.lines().map(str::trim) {
        let skip = line.is_empty() || line.starts_with('#') || line.starts_with('[');
        if !(skip || line.contains(" = [") && line.ends_with(']')) {
            return Err(format!("areas.toml: keep each array on one line: `{line}`"));
        }
    }
    let mut areas = BTreeMap::new();
    let table = doc["areas"]
        .as_table()
        .ok_or("areas.toml has no [areas] table")?;
    for (name, item) in table.iter() {
        let array = item
            .as_array()
            .ok_or_else(|| format!("area {name} is not an array"))?;
        areas.insert(name.to_string(), strings(array, name)?);
    }
    let solo = doc["solo"]["files"]
        .as_array()
        .ok_or("areas.toml has no [solo] files array")?;
    Ok(Table {
        areas,
        solo: strings(solo, "solo files")?,
    })
}

fn load_table(root: &Path) -> Result<Table, String> {
    let path = root.join("tests/areas.toml");
    let text = fs::read_to_string(&path).map_err(|e| format!("read {}: {e}", path.display()))?;
    parse_table(&text)
}

fn first_word(stem: &str) -> &str {
    stem.split('_').next().unwrap_or(stem)
}

fn word_areas(table: &Table) -> BTreeMap<&str, &str> {
    let mut map = BTreeMap::new();
    for (area, words) in &table.areas {
        for word in words {
            map.entry(word.as_str()).or_insert(area.as_str());
        }
    }
    map
}

/// The stems of the `.rs` files directly inside `dir`, sorted.
fn rs_stems(dir: &Path) -> Vec<String> {
    let mut out = Vec::new();
    if let Ok(entries) = fs::read_dir(dir) {
        for entry in entries.flatten() {
            let path = entry.path();
            if path.is_file() && path.extension().is_some_and(|e| e == "rs") {
                if let Some(stem) = path.file_stem() {
                    out.push(stem.to_string_lossy().to_string());
                }
            }
        }
    }
    out.sort();
    out
}

fn top_level_tests(root: &Path) -> Vec<String> {
    rs_stems(&root.join("tests"))
}

/// The module files of an area crate: every `.rs` file in `tests/<area>/` except `main.rs`.
fn area_modules(root: &Path, area: &str) -> Vec<String> {
    rs_stems(&root.join("tests").join(area))
        .into_iter()
        .filter(|stem| stem != "main")
        .collect()
}

fn has_main(root: &Path, dir: &str) -> bool {
    root.join("tests").join(dir).join("main.rs").is_file()
}

/// Folders of `tests/` that hold a `main.rs`: crates cargo discovers.
fn crate_dirs(root: &Path) -> Vec<String> {
    let mut out = Vec::new();
    if let Ok(entries) = fs::read_dir(root.join("tests")) {
        for entry in entries.flatten() {
            let name = entry.file_name().to_string_lossy().to_string();
            if entry.path().is_dir() && has_main(root, &name) {
                out.push(name);
            }
        }
    }
    out.sort();
    out
}

fn moved_areas(root: &Path, table: &Table) -> Vec<String> {
    table
        .areas
        .keys()
        .filter(|area| has_main(root, area))
        .cloned()
        .collect()
}

/// The modules an area's `main.rs` declares (`mod x;`, with or without attributes or `pub`),
/// `common` left out.
fn declared_modules(root: &Path, area: &str) -> Vec<String> {
    let path = root.join("tests").join(area).join("main.rs");
    let text = fs::read_to_string(path).unwrap_or_default();
    let mut out = Vec::new();
    // A module under `#[path = ...]` lives outside the area folder (the shared helpers): it needs
    // no file here, like `common`.
    let mut path_attr = false;
    for line in text.lines() {
        let mut line = line.trim();
        while line.starts_with("#[") {
            path_attr |= line.starts_with("#[path");
            match line.find("] ") {
                Some(i) => line = line[i + 2..].trim_start(),
                None => break,
            }
        }
        let line = line.strip_prefix("pub ").unwrap_or(line);
        if let Some(name) = line
            .strip_prefix("mod ")
            .and_then(|rest| rest.strip_suffix(';'))
        {
            if name != "common" && !path_attr {
                out.push(name.trim().to_string());
            }
            path_attr = false;
        }
    }
    out
}

/// A file of the table's first-word rule: an unmapped word, a stray crate, a file named like an
/// area crate, or a top-level file of a moved area that is not a solo target.
fn area_table_violations(root: &Path, table: &Table) -> Vec<String> {
    let words = word_areas(table);
    let mut out = Vec::new();
    let moved = moved_areas(root, table);
    for stem in top_level_tests(root) {
        let word = first_word(&stem);
        match words.get(word) {
            None => out.push(format!(
                "tests/{stem}.rs: the first word `{word}` is not in tests/areas.toml"
            )),
            Some(area) => {
                let solo = table.solo.contains(&stem);
                if moved.iter().any(|m| m == area) && !solo {
                    out.push(format!(
                        "tests/{stem}.rs belongs to the moved area `{area}`: it is a module of tests/{area}/main.rs"
                    ));
                }
            }
        }
        if table.areas.contains_key(&stem) {
            out.push(format!(
                "tests/{stem}.rs has the name of the area crate `{stem}`"
            ));
        }
    }
    for dir in crate_dirs(root) {
        if !table.areas.contains_key(&dir) {
            out.push(format!(
                "tests/{dir}/main.rs: `{dir}` is not an area in tests/areas.toml"
            ));
        }
    }
    for area in &moved {
        for stem in area_modules(root, area) {
            let word = first_word(&stem);
            match words.get(word) {
                None => out.push(format!(
                    "tests/{area}/{stem}.rs: the first word `{word}` is not in tests/areas.toml"
                )),
                Some(mapped) if mapped != area => out.push(format!(
                    "tests/{area}/{stem}.rs: its word `{word}` belongs to the area `{mapped}`"
                )),
                Some(_) => {}
            }
        }
    }
    out
}

fn duplicate_words(table: &Table) -> Vec<String> {
    let mut seen: BTreeMap<&str, &str> = BTreeMap::new();
    let mut out = Vec::new();
    for (area, words) in &table.areas {
        for word in words {
            if let Some(first) = seen.insert(word, area) {
                out.push(format!("`{word}` is listed in `{first}` and `{area}`"));
            }
        }
    }
    out
}

/// The module files and `main.rs` of every moved area agree.
fn module_violations(root: &Path, table: &Table) -> Vec<String> {
    let mut out = Vec::new();
    for area in moved_areas(root, table) {
        let files: BTreeSet<String> = area_modules(root, &area).into_iter().collect();
        let declared = declared_modules(root, &area);
        let mut seen = BTreeSet::new();
        for name in &declared {
            if !seen.insert(name.clone()) {
                out.push(format!("tests/{area}/main.rs declares `{name}` twice"));
            } else if !files.contains(name) {
                out.push(format!(
                    "tests/{area}/main.rs declares `{name}` but tests/{area}/{name}.rs is missing"
                ));
            }
        }
        for name in &files {
            if !seen.contains(name) {
                out.push(format!(
                    "tests/{area}/{name}.rs is not declared in tests/{area}/main.rs"
                ));
            }
            if table.solo.contains(name) {
                out.push(format!(
                    "tests/{area}/{name}.rs is a solo target: it stays at tests/{name}.rs"
                ));
            }
        }
    }
    out
}

/// The three calls that change process-global state. Spelled in pieces so this file does not
/// contain them.
fn global_state_calls() -> [String; 3] {
    [
        ["set_", "var("].concat(),
        ["remove_", "var("].concat(),
        ["set_current_", "dir("].concat(),
    ]
}

fn solo_violations(root: &Path, table: &Table) -> Vec<String> {
    let mut out = Vec::new();
    let top = top_level_tests(root);
    for name in &table.solo {
        if !top.contains(name) {
            out.push(format!(
                "solo target tests/{name}.rs is missing: it stays a test crate of its own"
            ));
        }
    }
    let calls = global_state_calls();
    for area in moved_areas(root, table) {
        for stem in area_modules(root, &area) {
            let path = root.join(format!("tests/{area}/{stem}.rs"));
            let text = fs::read_to_string(&path).unwrap_or_default();
            for (n, line) in text.lines().enumerate() {
                if line.trim_start().starts_with("//") {
                    continue;
                }
                if let Some(call) = calls.iter().find(|call| line.contains(call.as_str())) {
                    out.push(format!(
                        "tests/{area}/{stem}.rs:{}: `{call}` changes process-global state; the file stays a solo target",
                        n + 1
                    ));
                }
            }
        }
    }
    out
}

/// The `--test <name> [filter]` commands inside a text, as (name, first word after the name).
fn test_commands(text: &str) -> Vec<(String, Option<String>)> {
    let mut out = Vec::new();
    let mut rest = text;
    while let Some(i) = rest.find("--test ") {
        rest = &rest[i + "--test ".len()..];
        let name: String = rest
            .chars()
            .take_while(|c| c.is_ascii_alphanumeric() || *c == '_')
            .collect();
        let after = rest[name.len()..].trim_start_matches(' ');
        let filter: String = after
            .chars()
            .take_while(|c| !c.is_whitespace() && !"\"',\\;)".contains(*c))
            .collect();
        let filter = (!filter.is_empty() && !filter.starts_with('-')).then_some(filter);
        out.push((name, filter));
    }
    out
}

/// Every `tests/<name>.rs` or `tests/<dir>/<name>.rs` path written in a text.
fn test_paths(text: &str) -> Vec<String> {
    let mut out = Vec::new();
    let mut rest = text;
    while let Some(i) = rest.find("tests/") {
        let tail = &rest[i..];
        let end = tail
            .find(|c: char| !(c.is_ascii_alphanumeric() || "_/.-".contains(c)))
            .unwrap_or(tail.len());
        let word = &tail[..end];
        if word.ends_with(".rs") {
            out.push(word.to_string());
        }
        rest = &tail["tests/".len()..];
    }
    out.sort();
    out.dedup();
    out
}

/// The files that list test targets and test paths by name.
const LIST_FILES: [&str; 3] = [
    "tools/repo_scans.json",
    "tools/self_check.sh",
    "tools/ci_scope.py",
];

fn scan_list_violations(root: &Path, table: &Table) -> Vec<String> {
    let mut out = Vec::new();
    let top = top_level_tests(root);
    for file in LIST_FILES {
        let Ok(text) = fs::read_to_string(root.join(file)) else {
            continue;
        };
        for (name, filter) in test_commands(&text) {
            if top.contains(&name) && !table.areas.contains_key(&name) {
                continue;
            }
            if table.areas.contains_key(&name) && has_main(root, &name) {
                let modules = declared_modules(root, &name);
                match filter.as_deref().and_then(|f| f.split_once("::")) {
                    None => out.push(format!(
                        "{file}: `--test {name}` is an area crate: name a module filter `<module>::`"
                    )),
                    Some((module, _)) if !modules.iter().any(|m| m == module) => out.push(format!(
                        "{file}: `--test {name} {module}::` names no module of tests/{name}/main.rs"
                    )),
                    Some(_) => {}
                }
            } else {
                out.push(format!(
                    "{file}: `--test {name}` names no test target (no tests/{name}.rs, no tests/{name}/main.rs)"
                ));
            }
        }
        for path in test_paths(&text) {
            if !root.join(&path).is_file() {
                out.push(format!("{file}: `{path}` does not exist"));
            }
        }
    }
    out
}

fn root() -> &'static Path {
    Path::new(MANIFEST)
}

fn real_table() -> Table {
    load_table(root()).unwrap_or_else(|e| panic!("{e}"))
}

fn assert_clean(what: &str, violations: Vec<String>) {
    assert!(
        violations.is_empty(),
        "{what}:\n  {}",
        violations.join("\n  ")
    );
}

/// A temporary tree with the given files (path relative to the root, content).
fn plant(files: &[(&str, &str)]) -> tempfile::TempDir {
    let dir = tempfile::TempDir::new().unwrap();
    for (rel, content) in files {
        let path = dir.path().join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, content).unwrap();
    }
    dir
}

const TABLE: &str = "[areas]\n\
card = [\"card\"]\n\
update = [\"update\"]\n\
home = [\"home\"]\n\
project = [\"project\"]\n\
lane = [\"lane\"]\n\
screen = [\"screen\"]\n\
spec = [\"spec\"]\n\
repo = [\"repo\", \"git\"]\n\
[solo]\n\
files = [\"git_solo\"]\n";

fn planted_table(extra: &[(&str, &str)]) -> tempfile::TempDir {
    let mut files = vec![("tests/areas.toml", TABLE)];
    files.extend_from_slice(extra);
    plant(&files)
}

#[test]
fn layout_table_is_valid_one_line_toml() {
    let text = fs::read_to_string(root().join("tests/areas.toml")).unwrap();
    parse_table(&text).unwrap_or_else(|e| panic!("{e}"));
}

#[test]
fn layout_table_has_exactly_the_eight_areas() {
    let table = real_table();
    let names: Vec<&str> = table.areas.keys().map(String::as_str).collect();
    let mut expected = AREAS.to_vec();
    expected.sort();
    assert_eq!(names, expected);
}

#[test]
fn layout_no_word_is_listed_in_two_areas() {
    assert_clean("duplicate words", duplicate_words(&real_table()));
    let table = parse_table("[areas]\na = [\"x\"]\nb = [\"x\"]\n[solo]\nfiles = []\n").unwrap();
    assert_eq!(duplicate_words(&table).len(), 1);
}

#[test]
fn layout_a_wrapped_array_is_refused() {
    let wrapped = "[areas]\ncard = [\n  \"card\",\n]\n[solo]\nfiles = []\n";
    assert!(parse_table(wrapped).is_err());
}

#[test]
fn layout_moved_areas_agree_with_their_main_rs() {
    let table = real_table();
    assert_clean("area modules", module_violations(root(), &table));
}

#[test]
fn layout_a_planted_area_must_declare_each_module_once() {
    let ok = planted_table(&[
        (
            "tests/repo/main.rs",
            "mod common;\n#[path = \"../common/shared.rs\"]\nmod shared;\nmod git_a;\n#[cfg(unix)]\nmod git_b;\n",
        ),
        ("tests/repo/git_a.rs", ""),
        ("tests/repo/git_b.rs", ""),
    ]);
    let table = load_table(ok.path()).unwrap();
    assert_clean("a complete area", module_violations(ok.path(), &table));

    let bad = planted_table(&[
        (
            "tests/repo/main.rs",
            "mod git_a;\nmod git_a;\nmod git_gone;\n",
        ),
        ("tests/repo/git_a.rs", ""),
        ("tests/repo/git_late.rs", ""),
    ]);
    let violations = module_violations(bad.path(), &table);
    assert_eq!(violations.len(), 3, "{violations:?}");
    assert!(violations.iter().any(|v| v.contains("`git_a` twice")));
    assert!(violations
        .iter()
        .any(|v| v.contains("git_gone.rs is missing")));
    assert!(violations
        .iter()
        .any(|v| v.contains("git_late.rs is not declared")));
}

#[test]
fn area_table_maps_every_real_test_file() {
    assert_clean(
        "test files outside the area table",
        area_table_violations(root(), &real_table()),
    );
}

#[test]
fn area_table_rejects_a_planted_unmapped_word() {
    let tree = planted_table(&[("tests/zebra_cli.rs", "")]);
    let table = load_table(tree.path()).unwrap();
    let violations = area_table_violations(tree.path(), &table);
    assert_eq!(violations.len(), 1, "{violations:?}");
    assert!(violations[0].contains("`zebra` is not in tests/areas.toml"));
}

#[test]
fn area_table_rejects_a_planted_file_of_a_moved_area() {
    let tree = planted_table(&[
        ("tests/repo/main.rs", ""),
        ("tests/git_x.rs", ""),
        ("tests/git_solo.rs", ""),
    ]);
    let table = load_table(tree.path()).unwrap();
    let violations = area_table_violations(tree.path(), &table);
    assert_eq!(violations.len(), 1, "{violations:?}");
    assert!(violations[0].contains("tests/git_x.rs belongs to the moved area `repo`"));
}

#[test]
fn area_table_rejects_a_stray_crate_and_a_module_of_another_area() {
    let tree = planted_table(&[
        ("tests/stray/main.rs", ""),
        ("tests/repo/main.rs", ""),
        ("tests/repo/card_misplaced.rs", ""),
        ("tests/repo/nope_x.rs", ""),
    ]);
    let table = load_table(tree.path()).unwrap();
    let violations = area_table_violations(tree.path(), &table);
    assert_eq!(violations.len(), 3, "{violations:?}");
    assert!(violations
        .iter()
        .any(|v| v.contains("`stray` is not an area")));
    assert!(violations
        .iter()
        .any(|v| v.contains("belongs to the area `card`")));
    assert!(violations.iter().any(|v| v.contains("`nope` is not in")));
}

#[test]
fn area_table_rejects_a_test_file_named_like_an_area_crate() {
    let tree = planted_table(&[("tests/repo.rs", "")]);
    let table = load_table(tree.path()).unwrap();
    let violations = area_table_violations(tree.path(), &table);
    assert_eq!(violations.len(), 1, "{violations:?}");
    assert!(violations[0].contains("has the name of the area crate"));
}

#[test]
fn scan_list_names_only_existing_targets_and_paths() {
    assert_clean(
        "target lists naming a missing target or path",
        scan_list_violations(root(), &real_table()),
    );
}

#[test]
fn scan_list_rejects_a_planted_missing_target_and_path() {
    let tree = planted_table(&[
        ("tests/git_here.rs", ""),
        (
            "tools/self_check.sh",
            "cargo test --test git_here\ncargo test --test git_gone x\n# tests/git_here.rs tests/git_lost.rs\n",
        ),
    ]);
    let table = load_table(tree.path()).unwrap();
    let violations = scan_list_violations(tree.path(), &table);
    assert_eq!(violations.len(), 2, "{violations:?}");
    assert!(violations
        .iter()
        .any(|v| v.contains("`--test git_gone` names no test target")));
    assert!(violations
        .iter()
        .any(|v| v.contains("`tests/git_lost.rs` does not exist")));
}

#[test]
fn scan_list_wants_a_module_filter_on_a_moved_area() {
    let tree = planted_table(&[
        ("tests/repo/main.rs", "mod git_a;\n"),
        ("tests/repo/git_a.rs", ""),
        (
            "tools/repo_scans.json",
            "\"cargo test --test repo git_a::x\"\n\"cargo test --test repo\"\n\"cargo test --test repo git_b::\"\n",
        ),
    ]);
    let table = load_table(tree.path()).unwrap();
    let violations = scan_list_violations(tree.path(), &table);
    assert_eq!(violations.len(), 2, "{violations:?}");
    assert!(violations
        .iter()
        .any(|v| v.contains("name a module filter")));
    assert!(violations.iter().any(|v| v.contains("names no module")));
}

#[test]
fn solo_files_stay_top_level_targets() {
    let table = real_table();
    assert_eq!(
        table.solo,
        [
            "card_cleanup_outbox_cli",
            "commit_identity_guard_2878",
            "update_switch_layer",
            "worktree_place_cli"
        ]
    );
    assert_clean("solo targets", solo_violations(root(), &table));
}

#[test]
fn solo_rejects_a_missing_solo_file_and_a_global_state_call_in_a_module() {
    let call = ["std::env::set_", "var(\"A\", \"b\");"].concat();
    let comment = format!("// {}\n", ["remove_", "var("].concat());
    let tree = planted_table(&[
        ("tests/repo/main.rs", "mod git_a;\nmod git_b;\n"),
        ("tests/repo/git_a.rs", &format!("fn f() {{ {call} }}\n")),
        ("tests/repo/git_b.rs", &comment),
    ]);
    let table = load_table(tree.path()).unwrap();
    let violations = solo_violations(tree.path(), &table);
    assert_eq!(violations.len(), 2, "{violations:?}");
    assert!(violations
        .iter()
        .any(|v| v.contains("tests/git_solo.rs is missing")));
    assert!(violations
        .iter()
        .any(|v| v.contains("tests/repo/git_a.rs:1")));
}

/// Runs `tools/ci_scope.py` on the given changed paths and returns its JSON answer.
fn ci_scope(paths: &[&str]) -> serde_json::Value {
    use std::io::Write;
    let mut child = std::process::Command::new("python3")
        .arg("-I")
        .arg(root().join("tools/ci_scope.py"))
        .stdin(std::process::Stdio::piped())
        .stdout(std::process::Stdio::piped())
        .spawn()
        .expect("spawn tools/ci_scope.py");
    child
        .stdin
        .take()
        .unwrap()
        .write_all(paths.join("\n").as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    assert!(out.status.success(), "ci_scope.py failed");
    serde_json::from_slice(&out.stdout).expect("ci_scope.py prints JSON")
}

/// A platform-sensitive test file keeps selecting the macOS leg after it moves into an area
/// crate: the platform tables are keyed by the old `tests/<name>.rs` path.
#[test]
fn scan_list_ci_scope_looks_a_moved_test_up_by_its_old_path() {
    for moved in [
        "tests/home/exe_link_cli.rs",
        "tests/lane/launcher_auto_init.rs",
        "tests/screen/statusline_numbers_only.rs",
        "tests/update/update_hooks_cli.rs",
    ] {
        let answer = ci_scope(&[moved]);
        assert_eq!(answer["macos_needed"], true, "{moved}: {answer}");
    }
    let plain = ci_scope(&["tests/card/card_wait_cli.rs"]);
    assert_eq!(plain["macos_needed"], false, "{plain}");
}

/// An area crate root declares the area's modules, so a change to it cannot be judged as
/// platform-free.
#[test]
fn scan_list_ci_scope_treats_an_area_crate_root_as_platform_sensitive() {
    let answer = ci_scope(&["tests/repo/main.rs"]);
    assert_eq!(answer["macos_needed"], true, "{answer}");
}

/// The area names inside `tools/ci_scope.py` are the area names of the table.
#[test]
fn scan_list_ci_scope_area_dirs_are_the_table_areas() {
    let text = fs::read_to_string(root().join("tools/ci_scope.py")).unwrap();
    let line = text
        .lines()
        .find(|l| l.starts_with("AREA_DIRS = ("))
        .expect("tools/ci_scope.py names AREA_DIRS");
    let mut listed: Vec<&str> = line.split('"').skip(1).step_by(2).collect();
    listed.sort();
    let mut expected = AREAS.to_vec();
    expected.sort();
    assert_eq!(listed, expected);
}
