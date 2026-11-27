//! Cases for `tools/repo_scans.json` (#2681) that read the table and the repository's test files
//! and run no binary. The judge `moai-cli repo-scans` (#2708) is tested through the built binary
//! in `tests/repo_scans_cli.rs`.
//!
//! The table lists tests whose input is the checked-in tree. `every_walking_test_file_is_listed_or_exempt`
//! keeps it complete: a test file that names the repository root and walks a directory must be in
//! the table or in `EXEMPT` with a reason. The heuristic sees direct walks only; a scan hidden
//! inside product code (an op that reads skills/moai through package_digest, spec_links::extract
//! running `git ls-files`) is listed by hand and kept by the rule-file existence check.
//!
//! Moved here from the Python module `repo_scans_test.py` (SPEC-20261123-rust-scans-before-ready).
//!
//! The commands of a scan are built by `moai_cli::repo_scans::checks_of` (a `test_targets` array
//! and the `checks` list, SPEC-20261124-repo-scans-targets-per-line); the one case that starts a
//! program, `two_branches_adding_different_targets_merge_without_conflict`, runs git in a
//! temporary repository.

mod common;

use common::git::{git_ok, git_raw};
use moai_cli::repo_scans::{checks_of, is_target_name};
use serde_json::{json, Value};
use std::{
    collections::{BTreeMap, BTreeSet},
    fs,
    path::{Path, PathBuf},
    sync::OnceLock,
};

const ROOT: &str = env!("CARGO_MANIFEST_DIR");

/// Files the heuristic flags that only walk temporary directories or named files.
const EXEMPT: &[(&str, &str)] = &[
    (
        "tests/agents_flat_ownership.rs",
        "read_dir of temporary homes (the install tree before and after); reads the named skills/moai package files and no other checked-in tree",
    ),
    ("tests/atlas_query_cli.rs", "walks the TempDir tree it built from named atlas fixtures"),
    (
        "tests/audit_prep_cli.rs",
        "read_dir lists the logs audit-prep left in a slot of its temporary repository; CARGO_MANIFEST_DIR is only text of the test crate it writes there",
    ),
    (
        "tests/base_agents_discovery_record.rs",
        "read_dir copies the named skills/moai and agents folders into a temporary home and lists the profile folder it built; reads no other checked-in tree",
    ),
    (
        "tests/base_agents_nested.rs",
        "read_dir of temporary homes (the install tree before and after); reads the named src and package files",
    ),
    ("tests/ci_install_ripgrep_test.py", "rglob over the test's own temporary runner folder"),
    (
        "tests/codex_home_update_removed.rs",
        "read_dir of temporary homes (before/after trees of the Codex and OMP folders) and of src/update, whose files the `reads-src` scan names",
    ),
    ("tests/card/decision_append_cli.rs", "read_dir of a TempDir docs folder"),
    (
        "tests/e2e_host_fixture.rs",
        "clones into a temp dir, then checks named route and workflow files",
    ),
    (
        "tests/e2e_temp_home_cli.rs",
        "read_dir copies the named skills/moai and agents folders into a temporary version folder and lists the temporary home it built; reads no other checked-in tree",
    ),
    ("tests/card/forge_cli.rs", "recursive walks of TempDir only"),
    ("tests/card/github_cli.rs", "hashes temp dirs; templates are named include_str! files"),
    (
        "tests/home_style_cli.rs",
        "read_dir of the temporary home it built (before/after trees of the config and state folders); reads no checked-in tree",
    ),
    (
        "tests/home_style_ownership_cli.rs",
        "read_dir of the temporary home it built (before/after trees); reads the named template and fixtures",
    ),
    (
        "tests/hook_bash_boundary_cli.rs",
        "read_dir of the temporary world it built (before/after trees); reads the named src/hook_bash_boundary.rs",
    ),
    (
        "tests/install_script_test.py",
        "os.walk of the test's temporary HOME and link folders; runs the named install/install.sh and install.ps1",
    ),
    (
        "tests/launcher_settings_receipt.rs",
        "read_dir of the temp home's receipts folder; reads the named operations document models_revert.md",
    ),
    (
        "tests/local_check_test.py",
        "os.walk of the test's own temporary clone (a before/after snapshot of its working folder); reads the named tools/local_check.sh and .github/workflows/ci.yml",
    ),
    ("tests/mutate_test.py", "ls-files and scandir on temporary clones"),
    (
        "tests/project_flow_closing.rs",
        "reads only the named skills/moai/workflows/project.md; `git ls-files --error-unmatch` runs in TempDir repos",
    ),
    (
        "tests/project_lsp_plugins_cli.rs",
        "read_dir of the TempDir project and PATH folders it built (before/after trees); reads the named core-rules.md and project.md",
    ),
    (
        "tests/retired_hosts_layer.rs",
        "read_dir of temporary homes (before/after trees of a host folder); reads the named src files of the backup move and the skills/moai and agents packages",
    ),
    ("tests/status_line_cli.rs", "read_dir of a TempDir settings-backups folder"),
    (
        "tests/statusline_golden.rs",
        "read_dir lists the golden files of its own fixture folder to check that none is orphaned; every run and write is in a TempDir",
    ),
    (
        "tests/ui_korean_terminal_cli.rs",
        "read_dir of the TempDir project it set up; reads the named golden screens under tests/fixtures/ui",
    ),
    (
        "tests/update_route.rs",
        "walks its TempDir project; reads the named update.md and routes.json",
    ),
    ("tests/worktree_report.rs", "read_dir of temporary worktrees"),
];

/// What marks a Rust file as walking the tree: it names the repository root and uses one of these.
const RUST_WALK: &[&str] = &[
    "read_dir(",
    "ls-files",
    "package_digest(",
    "stage_package(",
    "package_agents(",
    "walk_files(",
    "package_files(",
];
const PY_ROOT: &[&str] = &["__file__", "REPO_ROOT"];
const PY_WALK: &[&str] = &[
    ".rglob(",
    ".glob(",
    "os.walk(",
    "ls-files",
    "\"grep\", \"-r",
];

fn root() -> &'static Path {
    Path::new(ROOT)
}

fn read_lossy(path: &Path) -> String {
    String::from_utf8_lossy(&fs::read(path).unwrap_or_else(|e| panic!("{}: {e}", path.display())))
        .into_owned()
}

fn load() -> Vec<Value> {
    let text = read_lossy(&root().join("tools/repo_scans.json"));
    let doc: Value = serde_json::from_str(&text).expect("tools/repo_scans.json is JSON");
    doc["scans"]
        .as_array()
        .expect("`scans` is an array")
        .clone()
}

fn strings(value: &Value) -> Vec<String> {
    value
        .as_array()
        .map(|a| {
            a.iter()
                .filter_map(|v| v.as_str().map(str::to_string))
                .collect()
        })
        .unwrap_or_default()
}

/// The commands of one scan as the readers see them: the command built from `test_targets` first,
/// then the `checks` list (SPEC-20261124-repo-scans-targets-per-line D-4: one function builds
/// them for the judge, `check affected` and this file).
fn built_checks(scan: &Value) -> Vec<String> {
    checks_of(scan).unwrap_or_else(|e| panic!("{}: {e}", id(scan)))
}

/// Every file below `dir` (recursive), as paths relative to the repository root with `/`. A
/// `.claude/worktrees` folder is not entered: it holds other checkouts, not this tree
/// (SPEC-20261122-worktree-place D-13).
fn walk(dir: &Path, out: &mut Vec<PathBuf>) {
    let Ok(entries) = fs::read_dir(dir) else {
        return;
    };
    for entry in entries.flatten() {
        let path = entry.path();
        // A link is never followed: the walk reads the checked-in tree, not what it points to.
        match entry.file_type() {
            Ok(t) if t.is_dir() && path.ends_with(".claude/worktrees") => {}
            Ok(t) if t.is_dir() => walk(&path, out),
            Ok(t) if t.is_file() => out.push(path),
            _ => {}
        }
    }
}

/// A test source the scans read: a top-level `tests/*.rs` file, or a module file of an area crate
/// (`tests/<area>/*.rs` where `tests/<area>/main.rs` exists, SPEC-20261123-test-area-crates).
fn is_test_source(path: &Path) -> bool {
    let tests = root().join("tests");
    match path.parent() {
        Some(parent) if parent == tests => true,
        Some(parent) => parent.parent() == Some(&tests) && parent.join("main.rs").is_file(),
        None => false,
    }
}

fn relative(path: &Path) -> String {
    path.strip_prefix(root())
        .unwrap()
        .to_string_lossy()
        .replace('\\', "/")
}

/// The files `tests/*.rs`, `src/**/*.rs` and `tests/*_test.py` that name the root and walk it.
fn flagged() -> BTreeSet<String> {
    let mut files = Vec::new();
    walk(&root().join("src"), &mut files);
    let mut tests = Vec::new();
    walk(&root().join("tests"), &mut tests);
    files.extend(tests.into_iter().filter(|p| is_test_source(p)));
    let mut out = BTreeSet::new();
    for path in files {
        let name = path.file_name().unwrap().to_string_lossy().to_string();
        let rust = name.ends_with(".rs");
        let py = name.ends_with("_test.py") && path.parent() == Some(&root().join("tests"));
        if !(rust || py) {
            continue;
        }
        let text = read_lossy(&path);
        let hit = if rust {
            text.contains("CARGO_MANIFEST_DIR") && RUST_WALK.iter().any(|n| text.contains(n))
        } else {
            PY_ROOT.iter().any(|n| text.contains(n)) && PY_WALK.iter().any(|n| text.contains(n))
        };
        if hit {
            out.insert(relative(&path));
        }
    }
    out
}

/// The test or rule file a `tests` entry names: the text before `::` and before ` (`.
fn test_file(entry: &str) -> &str {
    let head = entry.split("::").next().unwrap();
    head.split(" (").next().unwrap()
}

fn named_files(scans: &[Value]) -> BTreeSet<String> {
    let mut names = BTreeSet::new();
    for scan in scans {
        names.extend(strings(&scan["rules"]));
        for test in strings(&scan["tests"]) {
            names.insert(test_file(&test).to_string());
        }
    }
    names
}

fn id(scan: &Value) -> String {
    scan["id"].as_str().unwrap_or("?").to_string()
}

#[test]
fn entries_are_well_formed() {
    let scans = load();
    for scan in &scans {
        assert!(
            scan["id"].as_str().is_some_and(|s| !s.is_empty()),
            "an entry has no id: {scan}"
        );
    }
    let ids: Vec<String> = scans.iter().map(id).collect();
    let unique: BTreeSet<&String> = ids.iter().collect();
    assert_eq!(ids.len(), unique.len(), "duplicate id");
    for scan in &scans {
        for key in ["rules", "scope", "tests"] {
            assert!(
                scan[key].as_array().is_some_and(|a| !a.is_empty()),
                "{}: `{key}` is missing or empty",
                id(scan)
            );
        }
        for item in scan["scope"].as_array().unwrap() {
            assert!(
                item["prefix"].is_string(),
                "{}: a scope item has no prefix",
                id(scan)
            );
        }
    }
}

#[test]
fn every_named_file_exists() {
    let missing: Vec<String> = named_files(&load())
        .into_iter()
        .filter(|n| !root().join(n).is_file())
        .collect();
    assert!(
        missing.is_empty(),
        "rule or test file named in tools/repo_scans.json is gone: {missing:?}"
    );
}

#[test]
fn each_scanning_test_file_is_one_of_its_rule_files() {
    // a change to the test itself is a rule change of its scan
    for scan in load() {
        let rules = strings(&scan["rules"]);
        for test in strings(&scan["tests"]) {
            assert!(
                rules.iter().any(|r| r == test_file(&test)),
                "{}: {test}",
                id(&scan)
            );
        }
    }
}

#[test]
fn every_walking_test_file_is_listed_or_exempt() {
    let named = named_files(&load());
    let exempt: BTreeSet<&str> = EXEMPT.iter().map(|(p, _)| *p).collect();
    let unlisted: Vec<String> = flagged()
        .into_iter()
        .filter(|p| !named.contains(p) && !exempt.contains(p.as_str()))
        .collect();
    assert!(
        unlisted.is_empty(),
        "a test walks the repository tree: add it to tools/repo_scans.json with its rule files and scope, \
         or to EXEMPT with the reason it only walks temporary or named files: {unlisted:?}"
    );
}

/// The function names a source text declares: `fn <name>` where the name is `[a-z_0-9]+`.
fn fn_names(text: &str) -> Vec<String> {
    let bytes = text.as_bytes();
    let word = |b: u8| b.is_ascii_alphanumeric() || b == b'_';
    let mut names = Vec::new();
    let mut from = 0;
    while let Some(at) = text[from..].find("fn ") {
        let start = from + at;
        from = start + 3;
        if start > 0 && word(bytes[start - 1]) {
            continue;
        }
        let name: String = text[start + 3..]
            .chars()
            .take_while(|c| c.is_ascii_lowercase() || c.is_ascii_digit() || *c == '_')
            .collect();
        if !name.is_empty() {
            names.push(name);
        }
    }
    names
}

/// Whether `mod <name>` appears as a whole word.
fn declares_mod(text: &str, name: &str) -> bool {
    let needle = format!("mod {name}");
    let word = |c: char| c.is_ascii_alphanumeric() || c == '_';
    text.match_indices(&needle).any(|(at, _)| {
        let before = text[..at].chars().next_back();
        let after = text[at + needle.len()..].chars().next();
        !before.is_some_and(word) && !after.is_some_and(word)
    })
}

/// Static: the gate runs these on the merge result, so a check that selects nothing would pass
/// silently (cargo exits 0 on "running 0 tests"). No cargo is spawned here (#2661).
#[test]
fn every_check_names_a_real_target_and_test() {
    for scan in load() {
        assert_checks_name_real_targets(&scan);
    }
    // SPEC-20261124-repo-scans-targets-per-line REQ-6: a name in `test_targets` is a target like
    // the one in a listed `--test`, so it needs its `tests/<name>.rs`.
    let named = |names: Value| json!({"id": "zz", "test_targets": names});
    assert_checks_name_real_targets(&named(json!(["repo_scans_table"])));
    let missing = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        assert_checks_name_real_targets(&named(json!(["zz_no_such_test_file"])))
    }));
    assert!(
        missing.is_err(),
        "a `test_targets` name with no tests/<name>.rs must fail the target check"
    );
}

/// The judgement of `every_check_names_a_real_target_and_test` for one scan.
fn assert_checks_name_real_targets(scan: &Value) {
    let cargo = "cargo test --offline --locked ";
    let python = "python3 -m unittest ";
    {
        let checks = built_checks(scan);
        assert!(!checks.is_empty(), "{}", id(scan));
        for command in checks {
            if let Some(modules) = command.strip_prefix(python) {
                for module in modules.split_whitespace() {
                    let file = root().join(format!("{}.py", module.replace('.', "/")));
                    assert!(file.is_file(), "{command}");
                }
                continue;
            }
            let Some(rest) = command.strip_prefix(cargo) else {
                panic!("{command}: a check is a cargo test or a unittest command");
            };
            let words: Vec<&str> = rest.split_whitespace().collect();
            let (mut sources, mut filters, mut after_sep, mut i) =
                (Vec::<String>::new(), Vec::<&str>::new(), false, 0);
            while i < words.len() {
                if after_sep {
                    filters.push(words[i]);
                    i += 1;
                } else if words[i] == "--test" {
                    let path = root().join("tests").join(format!("{}.rs", words[i + 1]));
                    if path.is_file() {
                        sources.push(read_lossy(&path));
                    } else {
                        // An area crate: `tests/<area>/main.rs` declares the modules and each
                        // module file holds its tests.
                        let dir = root().join("tests").join(words[i + 1]);
                        assert!(dir.join("main.rs").is_file(), "{command}");
                        let mut all = Vec::new();
                        walk(&dir, &mut all);
                        sources.extend(
                            all.iter()
                                .filter(|p| p.extension().is_some_and(|e| e == "rs"))
                                .map(|p| read_lossy(p)),
                        );
                    }
                    i += 2;
                } else if words[i] == "--lib" {
                    let mut all = Vec::new();
                    walk(&root().join("src"), &mut all);
                    sources.extend(
                        all.iter()
                            .filter(|p| {
                                p.extension().is_some_and(|e| e == "rs")
                                    && p.file_name().is_some_and(|n| n != "main.rs")
                            })
                            .map(|p| read_lossy(p)),
                    );
                    i += 1;
                } else if words[i..].starts_with(&["--bin", "moai-cli"]) {
                    sources.push(read_lossy(&root().join("src/main.rs")));
                    i += 2;
                } else if words[i] == "--" {
                    after_sep = true;
                    i += 1;
                } else {
                    // cargo takes one test name before `--`; more go after it to libtest
                    assert!(!words[i].starts_with('-'), "{command}");
                    assert!(filters.is_empty(), "{command}: put several names after --");
                    filters.push(words[i]);
                    i += 1;
                }
            }
            assert!(!sources.is_empty(), "{command}");
            let text = sources.join("\n");
            let fns = fn_names(&text);
            for f in filters {
                if let Some(module) = f.strip_suffix("::") {
                    assert!(declares_mod(&text, module), "{command}");
                } else {
                    assert!(fns.iter().any(|name| name.contains(f)), "{command}: {f}");
                }
            }
        }
    }
}

#[test]
fn exemptions_are_not_stale() {
    let flagged = flagged();
    let stale: Vec<&str> = EXEMPT
        .iter()
        .map(|(p, _)| *p)
        .filter(|p| !flagged.contains(*p))
        .collect();
    assert!(
        stale.is_empty(),
        "EXEMPT names a file the heuristic no longer flags: {stale:?}"
    );
}

/// The exemption list carries a reason for each file, and none twice.
#[test]
fn every_exemption_has_a_reason_and_appears_once() {
    let mut seen = BTreeMap::new();
    for (path, reason) in EXEMPT {
        assert!(!reason.trim().is_empty(), "{path}: no reason");
        assert!(seen.insert(*path, reason).is_none(), "{path} appears twice");
    }
}

// ---- the `always` flag and the `include_str!` readers (SPEC-20261123-land-repo-scans) ----

/// Integration test files that read a repository file outside `tests/` with `include_str!` or
/// `include_bytes!` and need no scan, each with its reason. None today.
const INCLUDE_STR_EXEMPT: &[(&str, &str)] = &[];

fn include_str_exempt(file: &str, _path: &str) -> bool {
    INCLUDE_STR_EXEMPT.iter().any(|(p, _)| *p == file)
}

/// The repository paths (relative to the root, with `/`) a Rust file at `file` reads with
/// `include_str!("<literal>")` or `include_bytes!("<literal>")` (SPEC-20261123-land-scan-more-readers
/// D-1), for the literals that leave `tests/`. A literal without `..` and a path that stays inside
/// `tests/` are not repository reads of this kind.
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

/// Whether `path` is in the scope of `scan`: a prefix and, when the item names any, a suffix.
fn scope_covers(scan: &Value, path: &str) -> bool {
    scan["scope"].as_array().into_iter().flatten().any(|item| {
        let prefix = item["prefix"].as_str().unwrap_or("");
        let suffixes = strings(&item["suffixes"]);
        path.starts_with(prefix)
            && (suffixes.is_empty() || suffixes.iter().any(|s| path.ends_with(s.as_str())))
    })
}

/// Whether a scan names `file` as one of its rule files. The tests of a `reads-` scan that open the
/// scope files are its rule files, as the user approved (SPEC-20261123-land-file-read-scans D-6):
/// a change to one of them selects the whole group.
fn names_file(scan: &Value, file: &str) -> bool {
    strings(&scan["rules"]).iter().any(|r| r == file)
}

/// The `file reads path` lines of the readers that no `always` scan covers: the file is named by an
/// `always` scan (`names_file`) whose scope holds the path it reads. `exempt(file, path)` is
/// true for a pair kept out (`INCLUDE_STR_EXEMPT` or `FILE_READ_EXEMPT`, each with its reason).
fn uncovered_readers(
    readers: &BTreeMap<String, Vec<String>>,
    scans: &[Value],
    exempt: &dyn Fn(&str, &str) -> bool,
) -> Vec<String> {
    let mut out = Vec::new();
    for (file, paths) in readers {
        for path in paths {
            if exempt(file, path) {
                continue;
            }
            let covered = scans.iter().any(|scan| {
                scan["always"] == Value::Bool(true)
                    && names_file(scan, file)
                    && scope_covers(scan, path)
            });
            if !covered {
                out.push(format!("{file} reads {path}"));
            }
        }
    }
    out
}

/// Every top-level `tests/*.rs` file with the repository paths it reads by `include_str!` or
/// `include_bytes!`.
fn include_str_readers() -> BTreeMap<String, Vec<String>> {
    let mut files = Vec::new();
    walk(&root().join("tests"), &mut files);
    let mut out = BTreeMap::new();
    for path in files {
        if !is_test_source(&path) || path.extension().is_none_or(|e| e != "rs") {
            continue;
        }
        let rel = relative(&path);
        let paths = include_str_paths(&rel, &read_lossy(&path));
        if !paths.is_empty() {
            out.insert(rel, paths);
        }
    }
    out
}

#[test]
fn always_marks_the_two_scans_that_read_the_whole_test_tree() {
    let scans = load();
    for want in ["test-area-layout", "clean-env-only"] {
        let scan = scans
            .iter()
            .find(|s| id(s) == want)
            .unwrap_or_else(|| panic!("no scan {want}"));
        assert_eq!(scan["always"], Value::Bool(true), "{want}");
    }
}

#[test]
fn every_include_str_reader_of_a_repository_file_has_an_always_scan() {
    let readers = include_str_readers();
    assert!(
        readers.contains_key("tests/green_test_lock.rs"),
        "the reader scan finds no include_str! reader at all: {readers:?}"
    );
    let uncovered = uncovered_readers(&readers, &load(), &include_str_exempt);
    assert!(uncovered.is_empty(), "{INCLUDE_READER_HINT}: {uncovered:?}");
}

/// What the failure of `every_include_str_reader_of_a_repository_file_has_an_always_scan` tells.
const INCLUDE_READER_HINT: &str =
    "a test reads a repository file with include_str! or include_bytes! and no `always` scan names it as a rule \
     file with that file in its scope: add an entry to tools/repo_scans.json, or an \
     INCLUDE_STR_EXEMPT line with the reason";

#[test]
fn the_include_str_reader_texts_name_include_bytes() {
    assert!(
        INCLUDE_READER_HINT.contains("include_bytes!"),
        "{INCLUDE_READER_HINT}"
    );
}

#[test]
fn the_include_str_check_sees_a_reader_with_no_scan_and_skips_what_it_should() {
    let text = "const A: &str = include_str!(\"../src/zz_new_table.rs\");\n\
                const B: &str = include_str!(\"fixtures/x.txt\");\n\
                const C: &str = include_str!(\"../tests/fixtures/y.txt\");\n\
                const D: &str = include_str!(concat!(env!(\"OUT_DIR\"), \"/z.rs\"));\n";
    assert_eq!(
        include_str_paths("tests/zz_reader.rs", text),
        ["src/zz_new_table.rs"]
    );
    let readers = BTreeMap::from([(
        "tests/zz_reader.rs".to_string(),
        include_str_paths("tests/zz_reader.rs", text),
    )]);
    assert_eq!(
        uncovered_readers(&readers, &load(), &include_str_exempt),
        ["tests/zz_reader.rs reads src/zz_new_table.rs"]
    );
}

#[test]
fn the_green_test_lock_reader_is_covered_by_its_source_table() {
    let scans = load();
    let scan = scans
        .iter()
        .find(|s| {
            strings(&s["rules"])
                .iter()
                .any(|r| r == "tests/green_test_lock.rs")
        })
        .expect("an entry that names tests/green_test_lock.rs");
    assert_eq!(scan["always"], Value::Bool(true), "{}", id(scan));
    assert!(scope_covers(scan, "src/project_detect.rs"), "{}", id(scan));
}

// ---- readers that open a checked-in file at run time (SPEC-20261123-land-file-read-scans) ----

/// What a reader spells that is no read of the checkout: `(file, paths, reason)`. The paths are
/// names the test writes into the temporary project it builds (or lists as text); an empty list
/// keeps the whole file out. The reason says how that was read from the file.
///
/// Only the paths that change often are listed: `src/lib.rs`, `src/main.rs`, `README.md` and the
/// `.moai/specs` folder. A fixture name left in a scope only adds a run (D-2).
const FILE_READ_EXEMPT: &[(&str, &[&str], &str)] = &[
    (
        "tests/atlas_query_cli.rs",
        &["src/lib.rs"],
        "writes src/lib.rs into the temporary project (`write(root, ..)`) and names it in the query",
    ),
    (
        "tests/audit_admission.rs",
        &["README.md"],
        "`fs::write` into its TempDir repository, then `git add README.md` there",
    ),
    (
        "tests/audit_prep_cli.rs",
        &[".gitignore", "Cargo.toml", "src/lib.rs"],
        "writes them into the temporary project and names them in the build it runs",
    ),
    (
        "tests/boundary_repro.rs",
        &["src/main.rs"],
        "src/main.rs is an argument of `overlap(..)`, a path text compared with another, never opened",
    ),
    (
        "tests/bundle_body_cli.rs",
        &[".moai/specs"],
        "`put(..)` writes `.moai/specs/<id>/spec.md` into the fixture repository it builds",
    ),
    (
        "tests/bundle_build_cli.rs",
        &[".moai/specs"],
        "`put(..)` writes `.moai/specs/<id>/spec.md` into the fixture repository it builds",
    ),
    (
        "tests/bundle_verify_cli.rs",
        &[".moai/specs"],
        "`put(..)` and `fs::write` put `.moai/specs/<id>/spec.md` into the fixture worktree it builds",
    ),
    (
        "tests/card/card_finish_cli.rs",
        &[".moai/specs"],
        "joins `.moai/specs` onto the temporary project it builds",
    ),
    (
        "tests/card/card_line_lease_cli.rs",
        &["src/lib.rs"],
        "a path leased in its temporary ledger, never opened",
    ),
    (
        "tests/card/release_branch_version.rs",
        &["HEADVER"],
        "writes `HEADVER` into the temporary repository it builds and plans a release there",
    ),
    (
        "tests/check_affected_modes_cli.rs",
        &["README.md"],
        "`put(p, \"README.md\", ..)` into the fixture repository it builds",
    ),
    (
        "tests/code_atlas_reference.rs",
        &["src/lib.rs"],
        "`fs::write` into the temporary project it builds",
    ),
    (
        "tests/freshness_fixture.rs",
        &["src/lib.rs"],
        "a file body in the map of fixture file contents",
    ),
    (
        "tests/gate_collect.rs",
        &["src/lib.rs"],
        "committed into the fixture repository (`commit(d, &[(\"src/lib.rs\", ..)])`)",
    ),
    (
        "tests/generation_request_record.rs",
        &["CLAUDE.md", "README.md"],
        "names of files in the temporary project it builds (a planted link target, the launch text)",
    ),
    (
        "tests/gitignore_block_cli.rs",
        &[
            ".moai-cli/manifests",
            ".moai-cli/manifests/generation.json",
            ".moai-cli/manifests/profile-baseline.json",
            ".moai-cli/project.yaml",
            ".moai/project.yaml",
            ".moai/specs",
            "CLAUDE.md",
            "skills/moai/workflows",
        ],
        "names of files in the temporary projects it builds; its one read of the checkout is the source scan, covered by the reads-gitignore-block-cli scan",
    ),
    (
        "tests/ops_table_gen_cli.rs",
        &["src/main.rs", "src/lib.rs"],
        "the shared file names it looks for in the research table text",
    ),
    (
        "tests/product_cli.rs",
        &[".moai/specs"],
        "`repo.join(\".moai/specs\").join(arg).is_dir()` (product_cli.rs:2897) checks that a SPEC id named in its adoption records still has a folder and reads no content; only removing or renaming a SPEC folder breaks it, which a card does not do, while every card edits this folder",
    ),
    (
        "tests/repo_scans_cli.rs",
        &["README.md", "src/lib.rs"],
        "changed-path arguments of the judge and fixture writes (`fx.write(..)`)",
    ),
    (
        "tests/repo_scans_table.rs",
        &[".moai/specs", "HEADVER"],
        "a name in the fixture tests of this file, and in the exemption row of the release test",
    ),
    (
        "tests/review_check_cli.rs",
        &["src/lib.rs"],
        "`write(&repo, ..)` into the fixture repository it builds",
    ),
    (
        "tests/run_verify.rs",
        &["src/lib.rs"],
        "copied into its TempDir project (`fs::copy(.., dir.path().join(..))`)",
    ),
    (
        "tests/spec_approval_digest.rs",
        &[".moai/specs"],
        "`World::write_spec` puts the SPEC file under the specs folder of the temporary project it builds",
    ),
    (
        "tests/spec_branch_name_first.rs",
        &["src/main.rs", ".moai/specs"],
        "paths a fixture tool call writes under its temporary root",
    ),
    (
        "tests/spec_cli.rs",
        &["README.md", ".moai/specs"],
        "README.md is written into the specs folder of its temporary project; `.moai/specs` is joined onto that project",
    ),
    (
        "tests/update_route.rs",
        &[".moai/specs"],
        "writes `.moai/specs/SPEC-1/spec.md` under its temporary root",
    ),
    (
        "tests/worktree_cli.rs",
        &["src/lib.rs"],
        "path names in the overlap lists",
    ),
    (
        "tests/worktree_place_cli.rs",
        &[".gitignore", "README.md"],
        "both are written into the temporary repositories it builds (`fs::write`, then `git add` there) and looked for in the worktrees made from them",
    ),
];

fn file_read_exempt(file: &str, path: &str) -> bool {
    FILE_READ_EXEMPT
        .iter()
        .any(|(f, paths, _)| *f == file && (paths.is_empty() || paths.contains(&path)))
}

/// The literals that start an interpreter. A file that names one and `CARGO_MANIFEST_DIR` runs a
/// checked-in script, so its literal `tools` is the script folder (SPEC-20261123-land-scan-more-readers D-3).
const INTERPRETER_LITERALS: &[&str] = &["\"python3\"", "\"bash\"", "\"sh\""];

/// The files git tracks and every folder above them, relative to the root with `/`.
struct Tracked {
    files: BTreeSet<String>,
    dirs: BTreeSet<String>,
}

impl Tracked {
    fn from_paths<'a>(paths: impl IntoIterator<Item = &'a str>) -> Self {
        let (mut files, mut dirs) = (BTreeSet::new(), BTreeSet::new());
        for path in paths {
            for (at, _) in path.match_indices('/') {
                dirs.insert(path[..at].to_string());
            }
            files.insert(path.to_string());
        }
        Tracked { files, dirs }
    }

    /// `git ls-files` of the checkout; a checkout without git falls back to the files on disk
    /// (more paths than git tracks, which only widens what a reader may spell).
    fn of_repo() -> &'static Tracked {
        static TRACKED: OnceLock<Tracked> = OnceLock::new();
        TRACKED.get_or_init(|| {
            let out = common::git::git_raw(root(), &["ls-files", "-z"]);
            if out.success {
                Tracked::from_paths(out.stdout.split('\0').filter(|p| !p.is_empty()))
            } else {
                let mut files = Vec::new();
                walk(root(), &mut files);
                let rel: Vec<String> = files.iter().map(|p| relative(p)).collect();
                Tracked::from_paths(
                    rel.iter()
                        .map(String::as_str)
                        .filter(|p| !p.starts_with("target/") && !p.starts_with(".git/")),
                )
            }
        })
    }
}

impl Tracked {
    /// The tracked path a string literal names, if it names one: no space or backslash, cut
    /// before a placeholder (`{`, `<`, `*`, `$`), outside `tests/`, `target/` and `.git/`. A file
    /// counts at any depth, a folder only with two segments or more (`docs` alone is the name
    /// of a folder in the temporary projects the tests build).
    fn repo_path(&self, literal: &str) -> Option<String> {
        if literal.chars().any(|c| c.is_whitespace() || c == '\\') {
            return None;
        }
        // a placeholder folder in front (`{}/skills/..`, `{manifest_dir}/Cargo.toml`) is the root
        // the test formats in; it is no part of the path (D-8)
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
/// `CARGO_MANIFEST_DIR`, with the tracked paths of its literals, whether it opens them itself or
/// hands them to a binary or a script (SPEC-20261123-land-scan-more-readers D-2). A path built from
/// pieces is not seen (SPEC-20261123-land-file-read-scans D-2).
fn file_read_paths(text: &str, tracked: &Tracked) -> Vec<String> {
    if !text.contains("CARGO_MANIFEST_DIR") {
        return Vec::new();
    }
    let literals = string_literals(text);
    let mut out: Vec<String> = literals
        .iter()
        .filter_map(|literal| tracked.repo_path(literal))
        .collect();
    // D-3: next to an interpreter the literal `tools` is the folder of the checked-in scripts
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

/// Every top-level `tests/*.rs` file with the repository paths it may read at run time, before
/// `FILE_READ_EXEMPT` is applied.
fn file_readers() -> BTreeMap<String, Vec<String>> {
    let mut files = Vec::new();
    walk(&root().join("tests"), &mut files);
    let mut out = BTreeMap::new();
    for path in files {
        if !is_test_source(&path) || path.extension().is_none_or(|e| e != "rs") {
            continue;
        }
        let paths = file_read_paths(&read_lossy(&path), Tracked::of_repo());
        if !paths.is_empty() {
            out.insert(relative(&path), paths);
        }
    }
    out
}

/// The `always` scans that name a top-level test file as a rule file and run no check with
/// `--test <stem>` for it. A scan without `always` claims no coverage and is not held to this.
fn scans_not_running_readers(scans: &[Value]) -> Vec<String> {
    let mut out = Vec::new();
    for scan in scans.iter().filter(|s| s["always"] == Value::Bool(true)) {
        let checks = built_checks(scan);
        let run: BTreeSet<&str> = checks
            .iter()
            .flat_map(|c| {
                let words: Vec<&str> = c.split_whitespace().collect();
                words
                    .windows(2)
                    .filter(|w| w[0] == "--test")
                    .map(|w| w[1])
                    .collect::<Vec<_>>()
            })
            .collect();
        for rule in strings(&scan["rules"]) {
            let stem = rule
                .strip_prefix("tests/")
                .and_then(|r| r.strip_suffix(".rs"))
                .filter(|s| !s.contains('/'));
            if let Some(stem) = stem {
                if !run.contains(stem) {
                    out.push(format!(
                        "{}: {rule} is a rule file but no check runs --test {stem}",
                        id(scan)
                    ));
                }
            }
        }
    }
    out
}

#[test]
fn file_read_paths_finds_what_a_reader_spells_and_skips_the_rest() {
    let tracked = Tracked::from_paths([
        "skills/moai/SKILL.md",
        "skills/moai/workflows/work.md",
        "docs/a.md",
        "src/lib.rs",
        "src/ops/x.rs",
        "src/raw.rs",
        "tests/fixtures/x",
    ]);
    let reader = "const R: &str = env!(\"CARGO_MANIFEST_DIR\");\n\
        fn a() { std::fs::read_to_string(Path::new(R).join(\"skills/moai/SKILL.md\")).unwrap(); }\n\
        fn b() { std::fs::read_dir(Path::new(R).join(\"skills/moai/workflows\")).unwrap(); }\n\
        fn c() {\n\
            let _ = \"docs\";\n\
            let _ = \"docs/a.md\";\n\
            let _ = \"src/ops/{name}.rs\";\n\
            let _ = \"tests/fixtures/x\";\n\
            let _ = \"docs/a.md and more\";\n\
            let _ = \"/tmp/x/y\";\n\
            let _ = concat!(env!(\"CARGO_MANIFEST_DIR\"), \"/src/lib.rs\");\n\
        }\n\
        // \"skills/moai/commented.md\"\n\
        fn d() { let _ = r##\"src/raw.rs \"quoted\" \"##; let _q = '\"'; }\n";
    assert_eq!(
        file_read_paths(reader, &tracked),
        [
            "docs/a.md",
            "skills/moai/SKILL.md",
            "skills/moai/workflows",
            "src/lib.rs",
            "src/ops"
        ]
    );
    // no manifest dir: not a reader
    let no_root = reader.replace("CARGO_MANIFEST_DIR", "OTHER");
    assert!(file_read_paths(&no_root, &tracked).is_empty());
}

/// SPEC-20261123-land-scan-more-readers D-2: a file that names the manifest dir and spells a
/// repository path is a reader whether or not it calls a read function; it may hand the path to a
/// binary or a script.
#[test]
fn file_read_paths_counts_a_file_with_no_read_call() {
    let tracked = Tracked::from_paths(["skills/moai/SKILL.md", "tools/x.py"]);
    let hands_a_path_over =
        "const P: &str = concat!(env!(\"CARGO_MANIFEST_DIR\"), \"/skills/moai\");\n\
        fn run() { Command::new(BIN).arg(P).arg(\"tools/x.py\").output().unwrap(); }\n";
    assert_eq!(
        file_read_paths(hands_a_path_over, &tracked),
        ["skills/moai", "tools/x.py"]
    );
    let no_root = hands_a_path_over.replace("CARGO_MANIFEST_DIR", "OTHER");
    assert!(file_read_paths(&no_root, &tracked).is_empty());
    let no_path = "fn f() { let _ = env!(\"CARGO_MANIFEST_DIR\"); let _ = \"notes/y.txt\"; }\n";
    assert!(file_read_paths(no_path, &tracked).is_empty());
}

/// D-3: `tools` is the script folder next to an interpreter literal and a temporary project name
/// everywhere else.
#[test]
fn file_read_paths_counts_the_tools_folder_next_to_an_interpreter() {
    let tracked = Tracked::from_paths(["tools/zz_script.py", "src/lib.rs"]);
    let runs_a_script = "fn f() { Command::new(\"python3\").arg(Path::new(env!(\"CARGO_MANIFEST_DIR\")).join(\"tools\")); }\n";
    assert_eq!(file_read_paths(runs_a_script, &tracked), ["tools"]);
    for (what, text) in [
        ("shell", runs_a_script.replace("\"python3\"", "\"bash\"")),
        ("sh", runs_a_script.replace("\"python3\"", "\"sh\"")),
    ] {
        assert_eq!(file_read_paths(&text, &tracked), ["tools"], "{what}");
    }
    // no interpreter literal, no manifest dir, or no tracked `tools` folder: not a script run
    let no_interpreter = runs_a_script.replace("\"python3\"", "\"git\"");
    assert!(file_read_paths(&no_interpreter, &tracked).is_empty());
    let no_root = runs_a_script.replace("CARGO_MANIFEST_DIR", "OTHER");
    assert!(file_read_paths(&no_root, &tracked).is_empty());
    let no_tools_folder = Tracked::from_paths(["src/lib.rs"]);
    assert!(file_read_paths(runs_a_script, &no_tools_folder).is_empty());
}

/// D-1: `include_bytes!` leaves the test tree the same way `include_str!` does.
#[test]
fn include_bytes_is_read_like_include_str() {
    let text = "const A: &[u8] = include_bytes!(\"../install/zz.sh\");\n\
                const B: &[u8] = include_bytes!(\"fixtures/x.bin\");\n\
                const C: &[u8] = include_bytes!(\"../tests/fixtures/y.bin\");\n\
                const D: &str = include_str!(\"../src/zz.rs\");\n";
    assert_eq!(
        include_str_paths("tests/zz_reader.rs", text),
        ["install/zz.sh", "src/zz.rs"]
    );
}

/// D-4 and the live tree: the five readers of the card and `e2e_temp_home_cli` are found.
#[test]
fn the_more_readers_in_the_live_tree_are_found() {
    let reads = file_readers();
    // Tails, not whole paths: this file is a reader itself, and a whole tracked path spelled here
    // would need an exemption of its own (and a state-folder literal moves the literal baseline).
    for (file, tail) in [
        ("tests/atlas_three_arm.rs", "tools"),
        ("tests/ui_summary_terminal_cli.rs", "skills/moai"),
        ("tests/e2e_temp_home_cli.rs", "skills/moai"),
    ] {
        let paths = reads
            .get(file)
            .unwrap_or_else(|| panic!("the reader scan does not find {file}"));
        assert!(paths.iter().any(|p| p.ends_with(tail)), "{file}: {paths:?}");
    }
    let bytes = include_str_readers();
    for (file, tail) in [
        ("tests/update_switch_layer.rs", "/preview/moai.v1"),
        ("tests/update_switch_layer.rs", "/preview/moai.v2"),
        ("tests/update_switch_cli.rs", "/preview/moai.v1"),
        ("tests/update_switch_cli.rs", "/preview/moai.v2"),
    ] {
        let paths = bytes
            .get(file)
            .unwrap_or_else(|| panic!("the include scan does not find {file}"));
        assert!(paths.iter().any(|p| p.ends_with(tail)), "{file}: {paths:?}");
    }
}

/// REQ-6: each planted shape with no scan is named with its path, and an `always` scan that names
/// the file with the path in its scope clears it.
#[test]
fn a_planted_shape_without_a_scan_is_named_and_a_scan_clears_it() {
    let tracked = Tracked::from_paths(["install/zz.sh", "skills/moai/SKILL.md", "tools/zz.py"]);
    let bytes = "const A: &[u8] = include_bytes!(\"../install/zz.sh\");\n";
    let no_read =
        "const P: &str = concat!(env!(\"CARGO_MANIFEST_DIR\"), \"/skills/moai\");\nfn r() { run(P); }\n";
    let script = "fn r() { Command::new(\"python3\").arg(Path::new(env!(\"CARGO_MANIFEST_DIR\")).join(\"tools\")); }\n";
    let readers = BTreeMap::from([
        (
            "tests/zz_bytes.rs".to_string(),
            include_str_paths("tests/zz_bytes.rs", bytes),
        ),
        (
            "tests/zz_no_read.rs".to_string(),
            file_read_paths(no_read, &tracked),
        ),
        (
            "tests/zz_script.rs".to_string(),
            file_read_paths(script, &tracked),
        ),
    ]);
    let none = |_: &str, _: &str| false;
    assert_eq!(
        uncovered_readers(&readers, &[], &none),
        [
            "tests/zz_bytes.rs reads install/zz.sh",
            "tests/zz_no_read.rs reads skills/moai",
            "tests/zz_script.rs reads tools"
        ]
    );
    let scan = json!({"id": "reads-zz", "always": true,
        "rules": ["tests/zz_bytes.rs", "tests/zz_no_read.rs", "tests/zz_script.rs"],
        "scope": [{"prefix": "install/"}, {"prefix": "skills/moai"}, {"prefix": "tools"}]});
    assert!(uncovered_readers(&readers, &[scan], &none).is_empty());
}

/// SPEC-20261123-land-file-read-scans D-8: a test that formats the root into the path
/// (`format!("{}/skills/moai/{rel}", env!("CARGO_MANIFEST_DIR"))`, `"{manifest_dir}/Cargo.toml"`)
/// spells a repository path behind a placeholder folder; the placeholder folder is not part of it.
#[test]
fn file_read_paths_reads_through_a_placeholder_folder_in_front_of_the_path() {
    let tracked = Tracked::from_paths([
        "Cargo.toml",
        "skills/moai/SKILL.md",
        "src/lib.rs",
        "src/ops/x.rs",
    ]);
    let reader = "fn skill_text(rel: &str) -> String {\n\
        std::fs::read_to_string(format!(\"{}/skills/moai/{rel}\", env!(\"CARGO_MANIFEST_DIR\"))).unwrap()\n\
        }\n\
        fn manifest() -> String {\n\
            let manifest_dir = env!(\"CARGO_MANIFEST_DIR\");\n\
            std::fs::read_to_string(format!(\"{manifest_dir}/Cargo.toml\")).unwrap()\n\
        }\n\
        fn deeper() {\n\
            let _ = format!(\"{root}/{sub}/src/lib.rs\");\n\
            let _ = format!(\"{}/src/ops/{name}.rs\", 1);\n\
            let _ = \"{}\";\n\
            let _ = \"{x}/\";\n\
            let _ = \"{x}.rs\";\n\
            let _ = \"{x}/tests/fixtures/y\";\n\
            let _ = \"{x}/../Cargo.toml\";\n\
        }\n";
    assert_eq!(
        file_read_paths(reader, &tracked),
        ["Cargo.toml", "skills/moai", "src/lib.rs", "src/ops"]
    );
}

#[test]
fn file_read_scan_finds_the_readers_that_format_the_root_into_the_path() {
    let readers = file_readers();
    for (file, want) in [
        ("tests/project_decide_model.rs", "skills/moai"),
        ("tests/project_flow_closing.rs", "skills/moai"),
        ("tests/profile_debug_info.rs", "Cargo.toml"),
    ] {
        let paths = readers
            .get(file)
            .unwrap_or_else(|| panic!("the reader scan does not find {file}"));
        assert!(paths.iter().any(|p| p == want), "{file}: {paths:?}");
    }
}

/// SPEC-20261123-land-file-read-scans D-6 (revision 3): the readers of a `reads-` scan are rule
/// files of the scan, as the user approved; no key of their own holds them.
#[test]
fn a_scan_keeps_its_readers_in_rules_and_no_scan_has_a_readers_key() {
    for scan in load() {
        assert!(
            scan.get("readers").is_none(),
            "{}: name the readers in `rules`, not in a `readers` key",
            id(&scan)
        );
    }
    let scans = load();
    let reads: Vec<&Value> = scans
        .iter()
        .filter(|s| id(s).starts_with("reads-"))
        .collect();
    assert!(reads.len() >= 7, "{} reads- scans", reads.len());
    for scan in reads {
        let rules = strings(&scan["rules"]);
        assert!(
            rules.len() > 2,
            "{}: the readers are its rule files: {rules:?}",
            id(scan)
        );
        assert!(
            rules.iter().any(|r| r == "tests/repo_scans_table.rs"),
            "{}: the test that keeps the scan whole is a rule file too",
            id(scan)
        );
    }
}

#[test]
fn file_read_scan_finds_mode_workflow_steps_in_the_live_tree() {
    let readers = file_readers();
    let paths = readers
        .get("tests/mode_workflow_steps.rs")
        .unwrap_or_else(|| panic!("the reader scan does not find tests/mode_workflow_steps.rs"));
    assert!(
        paths.iter().any(|p| p == "skills/moai/workflows"),
        "{paths:?}"
    );
}

#[test]
fn an_uncovered_file_reader_is_named_and_a_covering_always_scan_clears_it() {
    let readers = BTreeMap::from([(
        "tests/zz_reader.rs".to_string(),
        vec!["skills/moai/SKILL.md".to_string()],
    )]);
    assert_eq!(
        uncovered_readers(&readers, &[], &|_, _| false),
        ["tests/zz_reader.rs reads skills/moai/SKILL.md"]
    );
    let scan = |extra: Value| {
        let mut s = json!({"id": "reads-zz", "always": true, "rules": ["tests/zz_reader.rs"],
                           "scope": [{"prefix": "skills/moai"}]});
        for (k, v) in extra.as_object().unwrap() {
            s[k.as_str()] = v.clone();
        }
        s
    };
    let none = |_: &str, _: &str| false;
    assert!(uncovered_readers(&readers, &[scan(json!({}))], &none).is_empty());
    // without `always`, with another rule file or with a scope elsewhere it stays uncovered
    for broken in [
        json!({"always": false}),
        json!({"rules": ["tests/other.rs"]}),
        json!({"scope": [{"prefix": "docs/"}]}),
    ] {
        assert_eq!(
            uncovered_readers(&readers, &[scan(broken)], &none).len(),
            1,
            "a scan that does not cover the reader must not clear it"
        );
    }
    // an exemption keeps out the named path of the named file and nothing else
    let one = |f: &str, p: &str| f == "tests/zz_reader.rs" && p == "skills/moai/SKILL.md";
    assert!(uncovered_readers(&readers, &[], &one).is_empty());
    let other_path = |f: &str, p: &str| f == "tests/zz_reader.rs" && p == "docs/a.md";
    assert_eq!(uncovered_readers(&readers, &[], &other_path).len(), 1);
}

#[test]
fn every_file_reader_has_an_always_scan_or_an_exemption() {
    let readers = file_readers();
    assert!(
        readers.len() > 50,
        "the reader scan finds only {} readers: {readers:?}",
        readers.len()
    );
    let uncovered = uncovered_readers(&readers, &load(), &file_read_exempt);
    assert!(
        uncovered.is_empty(),
        "a test reads a checked-in file at run time and no `always` scan names it as a rule file \
         with that path in its scope: add it to a `reads-` scan of tools/repo_scans.json, or to \
         FILE_READ_EXEMPT with the reason it only names paths of the temporary project it builds: \
         {uncovered:?}"
    );
}

#[test]
fn every_always_scan_runs_the_readers_it_names() {
    assert_eq!(scans_not_running_readers(&load()), Vec::<String>::new());
    // SPEC-20261124-repo-scans-targets-per-line REQ-6: a reader named in `test_targets` is run.
    let scan = |targets: Value| {
        json!({"id": "reads-zz", "always": true, "rules": ["tests/zz_reader.rs"],
               "test_targets": targets, "scope": [{"prefix": "docs/"}]})
    };
    assert!(scans_not_running_readers(&[scan(json!(["other", "zz_reader"]))]).is_empty());
    assert_eq!(
        scans_not_running_readers(&[scan(json!(["other"]))]),
        ["reads-zz: tests/zz_reader.rs is a rule file but no check runs --test zz_reader"]
    );
}

#[test]
fn a_scan_that_names_a_reader_but_never_runs_the_readers_is_named() {
    let scan = |checks: Value, always: bool| {
        json!({"id": "reads-zz", "always": always,
               "rules": ["tests/zz_reader.rs", "tests/areas.toml", "tests/common/mod.rs"],
               "checks": checks, "scope": [{"prefix": "docs/"}]})
    };
    let other = json!(["cargo test --offline --locked --test other"]);
    assert_eq!(
        scans_not_running_readers(&[scan(other.clone(), true)]),
        ["reads-zz: tests/zz_reader.rs is a rule file but no check runs --test zz_reader"]
    );
    let runs = json!(["cargo test --offline --locked --test other --test zz_reader"]);
    assert!(scans_not_running_readers(&[scan(runs, true)]).is_empty());
    // a scan without `always` claims no coverage, so it is not held to this
    assert!(scans_not_running_readers(&[scan(other.clone(), false)]).is_empty());
}

#[test]
fn file_read_exempt_keeps_out_the_listed_paths_of_the_listed_file_only() {
    let (file, paths, _) = FILE_READ_EXEMPT
        .iter()
        .find(|(_, paths, _)| !paths.is_empty())
        .expect("an exemption with a path list");
    for path in *paths {
        assert!(file_read_exempt(file, path), "{file} {path}");
        assert!(!file_read_exempt("tests/zz_other.rs", path), "{path}");
    }
    assert!(!file_read_exempt(file, "zz/not-listed.md"), "{file}");
}

#[test]
fn file_read_exemptions_are_readers_with_a_reason_once() {
    let readers = file_readers();
    let mut seen = BTreeSet::new();
    for (file, paths, reason) in FILE_READ_EXEMPT {
        assert!(!reason.trim().is_empty(), "{file}: no reason");
        assert!(seen.insert(*file), "{file} appears twice");
        let spelled = readers.get(*file).unwrap_or_else(|| {
            panic!("{file}: FILE_READ_EXEMPT names a file the reader scan no longer finds")
        });
        for path in *paths {
            assert!(
                spelled.iter().any(|p| p == path),
                "{file}: FILE_READ_EXEMPT names {path}, which the file no longer spells"
            );
        }
    }
}

/// SPEC-20261124-repo-scans-precheck REQ-10: a `FILE_READ_EXEMPT` pair that an `always` scan
/// already covers keeps nothing out. The exemption list may only hold what a scan does not cover.
#[test]
fn every_file_read_exemption_is_needed() {
    let readers = file_readers();
    let scans = load();
    let mut idle = Vec::new();
    for (file, paths, _) in FILE_READ_EXEMPT {
        let spelled = readers.get(*file).cloned().unwrap_or_default();
        for path in &spelled {
            if !paths.is_empty() && !paths.contains(&path.as_str()) {
                continue;
            }
            let one = BTreeMap::from([(file.to_string(), vec![path.clone()])]);
            if uncovered_readers(&one, &scans, &|_, _| false).is_empty() {
                idle.push(format!("{file} {path}"));
            }
        }
    }
    assert!(
        idle.is_empty(),
        "FILE_READ_EXEMPT keeps out a path that an `always` scan already covers; drop the pair: \
         {idle:?}"
    );
}

/// SPEC-20261124-repo-scans-precheck REQ-12: `tools` is the script folder only when a test file
/// spells an interpreter literal, the literal `tools` and a tracked `tools` folder together; each
/// of the three conditions, taken away alone, gives no folder.
#[test]
fn file_read_paths_adds_the_tools_folder_only_with_all_three_conditions() {
    let tracked = Tracked::from_paths(["tools/zz_script.py", "src/lib.rs"]);
    let all_three = "fn f() { Command::new(\"python3\").arg(Path::new(env!(\"CARGO_MANIFEST_DIR\")).join(\"tools\")); }\n";
    assert_eq!(file_read_paths(all_three, &tracked), ["tools"]);
    let no_interpreter = all_three.replace("\"python3\"", "\"git\"");
    assert!(file_read_paths(&no_interpreter, &tracked).is_empty());
    let no_literal = all_three.replace("\"tools\"", "\"toolz\"");
    assert!(file_read_paths(&no_literal, &tracked).is_empty());
    let no_folder = Tracked::from_paths(["src/lib.rs"]);
    assert!(file_read_paths(all_three, &no_folder).is_empty());
}

// ---- `test_targets` (SPEC-20261124-repo-scans-targets-per-line) ----

/// `"name"` or `"name",` with a target name inside: one quoted name and nothing else.
fn one_quoted_target(line: &str) -> Option<String> {
    let inner = line
        .strip_suffix(',')
        .unwrap_or(line)
        .strip_prefix('"')?
        .strip_suffix('"')?;
    let name = !inner.is_empty()
        && inner
            .chars()
            .all(|c| c.is_ascii_lowercase() || c.is_ascii_digit() || c == '_');
    name.then(|| inner.to_string())
}

/// REQ-5: the problems of the `test_targets` arrays in the text of a table. An array opens with
/// `[` alone after its key, holds one quoted name per line, and its names are strictly ascending
/// in byte order (so no name is there twice); the text, not the parsed value, is read, because
/// the lines are what two branches merge.
fn test_targets_violations(text: &str) -> Vec<String> {
    let mut out = Vec::new();
    let mut scan = "?".to_string();
    let mut open: Option<Vec<String>> = None;
    for line in text.lines().map(str::trim) {
        match open.as_mut() {
            None => {
                if let Some(rest) = line.strip_prefix("\"id\": \"") {
                    scan = rest.trim_end_matches(',').trim_end_matches('"').to_string();
                } else if line.starts_with("\"test_targets\":") {
                    if line == "\"test_targets\": [" {
                        open = Some(Vec::new());
                    } else {
                        out.push(format!(
                            "{scan}: `test_targets` opens with `[` alone on its line and holds \
                             one name per line"
                        ));
                    }
                }
            }
            Some(names) => {
                if line == "]" || line == "]," {
                    for pair in names.windows(2) {
                        if pair[1] <= pair[0] {
                            out.push(format!(
                                "{scan}: `{}` does not come after `{}` (sorted, each name once)",
                                pair[1], pair[0]
                            ));
                        }
                    }
                    open = None;
                } else if let Some(name) = one_quoted_target(line) {
                    names.push(name);
                } else {
                    out.push(format!(
                        "{scan}: `{line}` is not one quoted `[a-z0-9_]+` name on a line of its own"
                    ));
                }
            }
        }
    }
    out
}

#[test]
fn test_targets_are_sorted_unique_and_one_per_line() {
    let table = |array: &str| {
        format!(
            "{{\n  \"scans\": [\n    {{\n      \"id\": \"reads-zz\",\n      \"test_targets\": {array}\n    }}\n  ]\n}}\n"
        )
    };
    assert!(
        test_targets_violations(&table("[\n        \"a_cli\",\n        \"b_cli\"\n      ]"))
            .is_empty()
    );
    for (what, array, name) in [
        ("unsorted", "[\n \"b_cli\",\n \"a_cli\"\n ]", "a_cli"),
        ("repeated", "[\n \"a_cli\",\n \"a_cli\"\n ]", "a_cli"),
        ("two on a line", "[\n \"a_cli\", \"b_cli\"\n ]", "a_cli"),
        ("one-line array", "[\"a_cli\"]", "test_targets"),
        ("a bad character", "[\n \"a-cli\"\n ]", "a-cli"),
    ] {
        let found = test_targets_violations(&table(array));
        assert!(!found.is_empty(), "{what}: not found");
        assert!(
            found
                .iter()
                .all(|f| f.starts_with("reads-zz: ") && f.contains(name)),
            "{what}: {found:?}"
        );
    }
    assert_eq!(
        test_targets_violations(&read_lossy(&root().join("tools/repo_scans.json"))),
        Vec::<String>::new()
    );
}

/// The table text with one scan whose `test_targets` hold `names`, one name per line.
fn table_with_targets(names: &[&str]) -> String {
    let lines: Vec<String> = names.iter().map(|n| format!("        \"{n}\"")).collect();
    format!(
        "{{\n  \"scans\": [\n    {{\n      \"id\": \"reads-zz\",\n      \"test_targets\": [\n{}\n      ]\n    }}\n  ]\n}}\n",
        lines.join(",\n")
    )
}

/// The table text of the earlier form: one scan whose one command lists `names` after `--test`.
fn table_with_command(names: &[&str]) -> String {
    let tests: String = names.iter().map(|n| format!(" --test {n}")).collect();
    format!(
        "{{\n  \"scans\": [\n    {{\n      \"id\": \"reads-zz\",\n      \"checks\": [\n        \"cargo test --offline --locked{tests}\"\n      ]\n    }}\n  ]\n}}\n"
    )
}

/// Commits `base` to `main` of a new temporary repository, `one` and `two` on branches of their
/// own, and merges `two` into `one`: whether the merge succeeded and the table text afterwards.
fn merge_two(base: &str, one: &str, two: &str) -> (bool, String) {
    let dir = tempfile::tempdir().unwrap();
    let p = dir.path();
    let who = [
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
    ];
    let commit = |text: &str, message: &str| {
        fs::write(p.join("repo_scans.json"), text).unwrap();
        git_ok(p, &["add", "repo_scans.json"]);
        let mut args = who.to_vec();
        args.extend(["commit", "-q", "-m", message]);
        git_ok(p, &args);
    };
    git_ok(p, &["init", "-q", "-b", "main"]);
    commit(base, "base");
    git_ok(p, &["checkout", "-q", "-b", "one"]);
    commit(one, "one");
    git_ok(p, &["checkout", "-q", "main"]);
    git_ok(p, &["checkout", "-q", "-b", "two"]);
    commit(two, "two");
    git_ok(p, &["checkout", "-q", "one"]);
    let mut args = who.to_vec();
    args.extend(["merge", "--no-edit", "two"]);
    let merged = git_raw(p, &args).success;
    (
        merged,
        fs::read_to_string(p.join("repo_scans.json")).unwrap(),
    )
}

/// REQ-7: two branches that each add a target to the same scan, at places with a name between
/// them, merge without a conflict and keep both; the same two additions to the one command of
/// the earlier form conflict.
#[test]
fn two_branches_adding_different_targets_merge_without_conflict() {
    let base = ["a_cli", "c_cli", "e_cli", "g_cli", "i_cli"];
    let plus = |name: &'static str| {
        let mut names = base.to_vec();
        names.push(name);
        names.sort();
        names
    };
    let (merged, text) = merge_two(
        &table_with_targets(&base),
        &table_with_targets(&plus("b_cli")),
        &table_with_targets(&plus("h_cli")),
    );
    assert!(
        merged,
        "the merge of two separate additions conflicted:\n{text}"
    );
    let all = [
        "a_cli", "b_cli", "c_cli", "e_cli", "g_cli", "h_cli", "i_cli",
    ];
    assert_eq!(text, table_with_targets(&all));
    assert!(test_targets_violations(&text).is_empty());
    // the earlier form: both additions edit the one command line
    let (merged, _) = merge_two(
        &table_with_command(&base),
        &table_with_command(&plus("b_cli")),
        &table_with_command(&plus("h_cli")),
    );
    assert!(!merged, "the one-line form is the case that conflicts");
}

// ---- the conversion (SPEC-20261125-repo-scans-convert-test-targets) ----

/// REQ-3: a `checks` command that is one plain `cargo test --offline --locked --test <t> …` list,
/// the form that belongs in `test_targets`: the words of `^cargo test --offline --locked( --test
/// [a-z0-9_]+)+$`, each `--test` followed by one target name, single spaces, nothing else.
fn is_plain_target_command(command: &str) -> bool {
    let Some(rest) = command.strip_prefix("cargo test --offline --locked") else {
        return false;
    };
    let mut words = rest.split(' ');
    if words.next() != Some("") {
        return false;
    }
    let mut pairs = 0;
    loop {
        match (words.next(), words.next()) {
            (None, _) => break,
            (Some("--test"), Some(name)) if is_target_name(name) => pairs += 1,
            _ => return false,
        }
    }
    pairs > 0
}

/// REQ-3: one line per plain `--test` list in the `checks` of `scans`: the scan id, then the
/// command.
fn plain_target_commands(scans: &[Value]) -> Vec<String> {
    let mut out = Vec::new();
    for scan in scans {
        for command in strings(scan.get("checks").unwrap_or(&Value::Null)) {
            if is_plain_target_command(&command) {
                out.push(format!("{}: {command}", id(scan)));
            }
        }
    }
    out
}

/// REQ-3 (AC-2): the checked-in table keeps no plain `--test` list in `checks`: the names go to
/// the scan's sorted `test_targets` array, one per line, so that two branches that add a test do
/// not edit the same line. A command with a filter, a `--lib`, a `--` or another tool stays in
/// `checks`.
#[test]
fn no_check_is_a_plain_target_list() {
    let plain = "cargo test --offline --locked --test a_cli --test b_cli";
    let scans = [
        json!({"id": "one", "checks": [plain, "cargo test --offline --locked --lib state_dir"]}),
        json!({"id": "two", "test_targets": ["d_cli"], "checks": [
            "cargo test --offline --locked --test c_cli card_x:: -- y::",
            "cargo test --offline --locked --test c_cli",
            "python3 -m unittest tests.x_test"]}),
        json!({"id": "three", "test_targets": ["e_cli"]}),
    ];
    assert_eq!(
        plain_target_commands(&scans),
        [
            format!("one: {plain}"),
            "two: cargo test --offline --locked --test c_cli".to_string()
        ]
    );
    // the edges of the shape: each of these stays a listed command
    for kept in [
        "cargo test --offline --locked",
        "cargo test --offline --locked --test",
        "cargo test --offline --locked --test a_cli ",
        "cargo test --offline --locked --test  a_cli",
        "cargo test --offline --locked --test A_cli",
        "cargo test --offline --locked --test a-cli",
        "cargo test --offline --locked --test a_cli --lib b",
        "cargo test --offline --locked --test a_cli b::",
        "cargo test --offline --locked --test a_cli -- b::",
        "cargo test --offline --locked --lib a",
        "cargo test --locked --offline --test a_cli",
        "cargo test --offline --locked --test a_cli && true",
    ] {
        assert!(!is_plain_target_command(kept), "{kept}");
    }
    assert!(is_plain_target_command(
        "cargo test --offline --locked --test a"
    ));
    let found = plain_target_commands(&load());
    let shown: Vec<String> = found
        .iter()
        .map(|line| line.chars().take(110).collect::<String>() + "...")
        .collect();
    assert!(
        found.is_empty(),
        "tools/repo_scans.json keeps {} plain `--test` list(s) in `checks`; move the names into \
         the scan's sorted `test_targets` array (one name per line), which is the form two \
         branches merge without a conflict:\n{}",
        found.len(),
        shown.join("\n")
    );
}

/// The line numbers of the names of the `test_targets` array of scan `scan` in the table text, as
/// `(first name line, closing bracket line)`; or the reason the scan has none.
fn test_targets_span(lines: &[&str], scan: &str) -> Result<(usize, usize), String> {
    let marker = format!("\"id\": \"{scan}\"");
    let at = lines
        .iter()
        .position(|l| l.trim().trim_end_matches(',') == marker)
        .ok_or_else(|| format!("the table has no scan {scan}"))?;
    for (i, line) in lines.iter().enumerate().skip(at + 1) {
        let line = line.trim();
        if line.starts_with("\"id\":") {
            break;
        }
        if line == "\"test_targets\": [" {
            let close = lines[i + 1..]
                .iter()
                .position(|l| matches!(l.trim(), "]" | "],"))
                .ok_or_else(|| format!("scan {scan}: the `test_targets` array is never closed"))?;
            return Ok((i + 1, i + 1 + close));
        }
    }
    Err(format!(
        "scan {scan} has no `test_targets` array: the table is not converted yet"
    ))
}

/// The names of the lines `first..close` of the table text.
fn target_names(lines: &[&str], first: usize, close: usize) -> Vec<String> {
    lines[first..close]
        .iter()
        .map(|l| {
            one_quoted_target(l.trim())
                .unwrap_or_else(|| panic!("`{l}` is not one quoted target name on a line"))
        })
        .collect()
}

/// The table text with `name` added to the `test_targets` of scan `scan`, in its sorted place,
/// one name per line; the rest of the text is kept byte for byte.
fn with_added_target(text: &str, scan: &str, name: &str) -> Result<String, String> {
    let lines: Vec<&str> = text.lines().collect();
    let (first, close) = test_targets_span(&lines, scan)?;
    let mut names = target_names(&lines, first, close);
    names.push(name.to_string());
    names.sort();
    let indent: String = lines[first]
        .chars()
        .take_while(|c| c.is_whitespace())
        .collect();
    let body: Vec<String> = names.iter().map(|n| format!("{indent}\"{n}\"")).collect();
    let mut out: Vec<String> = lines[..first].iter().map(|l| l.to_string()).collect();
    out.push(body.join(",\n"));
    out.extend(lines[close..].iter().map(|l| l.to_string()));
    Ok(out.join("\n") + "\n")
}

/// REQ-5 (AC-4): two branches of the checked-in table that each add one target name to the
/// `test_targets` of `reads-skills`, one near the start of the array and one near its end, merge
/// without a conflict and keep both; the merged text still has one sorted name per line.
#[test]
fn two_branches_adding_targets_to_the_checked_in_table_merge_without_conflict() {
    let scan = "reads-skills";
    let base = read_lossy(&root().join("tools/repo_scans.json"));
    let lines: Vec<&str> = base.lines().collect();
    let span = test_targets_span(&lines, scan);
    assert!(
        span.is_ok(),
        "the checked-in table cannot show the merge: {}; the conversion moves the plain `--test` \
         list of {scan} into `test_targets`",
        span.as_ref().err().map(String::as_str).unwrap_or("")
    );
    let (first, close) = span.unwrap();
    let names = target_names(&lines, first, close);
    assert!(
        names.len() >= 6,
        "{scan} has {} names: too few for two additions that are not neighbours",
        names.len()
    );
    // a name that sorts right after the first one, and one that sorts after the last one
    let near_start = format!("{}_a", names[0]);
    let near_end = format!("{}_z", names[names.len() - 1]);
    assert!(
        names[0] < near_start && near_start < names[1] && !names.contains(&near_start),
        "{near_start} does not sort between {} and {}",
        names[0],
        names[1]
    );
    assert!(!names.contains(&near_end), "{near_end} is a name already");
    let one = with_added_target(&base, scan, &near_start).unwrap();
    let two = with_added_target(&base, scan, &near_end).unwrap();
    assert_ne!(one, base);
    assert_ne!(two, base);
    let (merged, text) = merge_two(&base, &one, &two);
    assert!(
        merged,
        "the merge of two additions to {scan} (after {} and after {}) conflicted",
        names[0],
        names[names.len() - 1]
    );
    for name in [&near_start, &near_end] {
        assert!(
            text.contains(&format!("\"{name}\"")),
            "the merged table lost {name}"
        );
    }
    assert_eq!(
        test_targets_violations(&text),
        Vec::<String>::new(),
        "the merged table"
    );
}
