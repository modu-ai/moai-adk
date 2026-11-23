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
    ("tests/atlas_query_cli.rs", "walks the TempDir tree it built from named atlas fixtures"),
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
        "read_dir of temporary homes (before/after trees of the Codex and OMP folders); reads no checked-in tree",
    ),
    ("tests/decision_append_cli.rs", "read_dir of a TempDir docs folder"),
    (
        "tests/e2e_host_fixture.rs",
        "clones into a temp dir, then checks named route and workflow files",
    ),
    ("tests/forge_cli.rs", "recursive walks of TempDir only"),
    ("tests/github_cli.rs", "hashes temp dirs; templates are named include_str! files"),
    (
        "tests/heavy_slot_cli.rs",
        "read_dir of a TempDir slot folder; reads the named tools/heavy_check_slot.py",
    ),
    (
        "tests/home_style_cli.rs",
        "read_dir of the temporary home it built (before/after trees of the config and state folders); reads no checked-in tree",
    ),
    (
        "tests/home_style_ownership_cli.rs",
        "read_dir of the temporary home it built (before/after trees); reads the named template and fixtures",
    ),
    (
        "tests/home_update_notice_follow_ups.rs",
        "read_dir copies the named skills/moai and agents folders into a temporary home; reads named fixtures",
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

/// Every file below `dir` (recursive), as paths relative to the repository root with `/`.
fn walk(dir: &Path, out: &mut Vec<PathBuf>) {
    let Ok(entries) = fs::read_dir(dir) else {
        return;
    };
    for entry in entries.flatten() {
        let path = entry.path();
        // A link is never followed: the walk reads the checked-in tree, not what it points to.
        match entry.file_type() {
            Ok(t) if t.is_dir() => walk(&path, out),
            Ok(t) if t.is_file() => out.push(path),
            _ => {}
        }
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
    files.extend(
        tests
            .into_iter()
            .filter(|p| p.parent() == Some(&root().join("tests"))),
    );
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
    let cargo = "cargo test --offline --locked ";
    let python = "python3 -m unittest ";
    for scan in load() {
        let checks = strings(&scan["checks"]);
        assert!(!checks.is_empty(), "{}", id(&scan));
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
                    assert!(path.is_file(), "{command}");
                    sources.push(read_lossy(&path));
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

/// Integration test files that read a repository file outside `tests/` with `include_str!` and
/// need no scan, each with its reason. None today.
const INCLUDE_STR_EXEMPT: &[(&str, &str)] = &[];

/// The repository paths (relative to the root, with `/`) a Rust file at `file` reads with
/// `include_str!("<literal>")`, for the literals that leave `tests/`. A literal without `..` and a
/// path that stays inside `tests/` are not repository reads of this kind.
fn include_str_paths(file: &str, text: &str) -> Vec<String> {
    const OPEN: &str = "include_str!(\"";
    let folder = file.rsplit_once('/').map_or("", |(dir, _)| dir);
    let mut out = Vec::new();
    let mut rest = text;
    while let Some(at) = rest.find(OPEN) {
        rest = &rest[at + OPEN.len()..];
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

/// The `file reads path` lines of the readers that no `always` scan covers: the file is one of the
/// rule files of an `always` scan whose scope holds the path it reads. `exempt` lists the files
/// kept out, each with its reason (`INCLUDE_STR_EXEMPT` or `FILE_READ_EXEMPT`).
fn uncovered_readers(
    readers: &BTreeMap<String, Vec<String>>,
    scans: &[Value],
    exempt: &[(&str, &str)],
) -> Vec<String> {
    let exempt: BTreeSet<&str> = exempt.iter().map(|(p, _)| *p).collect();
    let mut out = Vec::new();
    for (file, paths) in readers {
        if exempt.contains(file.as_str()) {
            continue;
        }
        for path in paths {
            let covered = scans.iter().any(|scan| {
                scan["always"] == Value::Bool(true)
                    && strings(&scan["rules"]).iter().any(|r| r == file)
                    && scope_covers(scan, path)
            });
            if !covered {
                out.push(format!("{file} reads {path}"));
            }
        }
    }
    out
}

/// Every top-level `tests/*.rs` file with the repository paths it reads by `include_str!`.
fn include_str_readers() -> BTreeMap<String, Vec<String>> {
    let mut files = Vec::new();
    walk(&root().join("tests"), &mut files);
    let mut out = BTreeMap::new();
    for path in files {
        let top = path.parent() == Some(&root().join("tests"));
        if !top || path.extension().is_none_or(|e| e != "rs") {
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
    let uncovered = uncovered_readers(&readers, &load(), INCLUDE_STR_EXEMPT);
    assert!(
        uncovered.is_empty(),
        "a test reads a repository file with include_str! and no `always` scan names it as a rule \
         file with that file in its scope: add an entry to tools/repo_scans.json, or an \
         INCLUDE_STR_EXEMPT line with the reason: {uncovered:?}"
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
        uncovered_readers(&readers, &load(), INCLUDE_STR_EXEMPT),
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

/// Integration test files that name `CARGO_MANIFEST_DIR`, read, and spell tracked paths only as
/// names inside the temporary projects they build, each with its reason. None yet.
const FILE_READ_EXEMPT: &[(&str, &str)] = &[];

/// The calls that open a file or a folder while the test runs.
const FILE_READ_CALLS: &[&str] = &["read_to_string(", "fs::read(", "read_dir(", "File::open("];

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
            let mut git = std::process::Command::new("git");
            git.arg("-C").arg(root()).args(["ls-files", "-z"]);
            for key in ["GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_COMMON_DIR"] {
                git.env_remove(key);
            }
            match git.output() {
                Ok(out) if out.status.success() => {
                    let text = String::from_utf8_lossy(&out.stdout).into_owned();
                    Tracked::from_paths(text.split('\0').filter(|p| !p.is_empty()))
                }
                _ => {
                    let mut files = Vec::new();
                    walk(root(), &mut files);
                    let rel: Vec<String> = files.iter().map(|p| relative(p)).collect();
                    Tracked::from_paths(
                        rel.iter()
                            .map(String::as_str)
                            .filter(|p| !p.starts_with("target/") && !p.starts_with(".git/")),
                    )
                }
            }
        })
    }
}

/// The repository paths a test file spells as string literals and so may read (RED stub).
fn file_read_paths(_text: &str, _tracked: &Tracked) -> Vec<String> {
    Vec::new()
}

/// Every top-level `tests/*.rs` file with the repository paths it may read at run time.
fn file_readers() -> BTreeMap<String, Vec<String>> {
    let mut files = Vec::new();
    walk(&root().join("tests"), &mut files);
    let mut out = BTreeMap::new();
    for path in files {
        let top = path.parent() == Some(&root().join("tests"));
        if !top || path.extension().is_none_or(|e| e != "rs") {
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
/// `--test <stem>` for it (RED stub).
fn scans_not_running_readers(_scans: &[Value]) -> Vec<String> {
    Vec::new()
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
    // no manifest dir, or no read call: not a reader
    let no_root = reader.replace("CARGO_MANIFEST_DIR", "OTHER");
    assert!(file_read_paths(&no_root, &tracked).is_empty());
    let no_read = reader
        .replace("read_to_string(", "x(")
        .replace("read_dir(", "y(");
    assert!(file_read_paths(&no_read, &tracked).is_empty());
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
        uncovered_readers(&readers, &[], FILE_READ_EXEMPT),
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
    assert!(uncovered_readers(&readers, &[scan(json!({}))], &[]).is_empty());
    // without `always`, with another rule file or with a scope elsewhere it stays uncovered
    for broken in [
        json!({"always": false}),
        json!({"rules": ["tests/other.rs"]}),
        json!({"scope": [{"prefix": "docs/"}]}),
    ] {
        assert_eq!(
            uncovered_readers(&readers, &[scan(broken)], &[]).len(),
            1,
            "a scan that does not cover the reader must not clear it"
        );
    }
    let exempt = [("tests/zz_reader.rs", "fixture names only")];
    assert!(uncovered_readers(&readers, &[], &exempt).is_empty());
}

#[test]
fn every_file_reader_has_an_always_scan_or_an_exemption() {
    let readers = file_readers();
    assert!(
        readers.len() > 50,
        "the reader scan finds only {} readers: {readers:?}",
        readers.len()
    );
    let uncovered = uncovered_readers(&readers, &load(), FILE_READ_EXEMPT);
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
    assert!(scans_not_running_readers(&[scan(other, false)]).is_empty());
}

#[test]
fn file_read_exemptions_are_readers_with_a_reason_once() {
    let readers = file_readers();
    let mut seen = BTreeSet::new();
    for (path, reason) in FILE_READ_EXEMPT {
        assert!(!reason.trim().is_empty(), "{path}: no reason");
        assert!(seen.insert(*path), "{path} appears twice");
        assert!(
            readers.contains_key(*path),
            "{path}: FILE_READ_EXEMPT names a file the reader scan no longer finds"
        );
    }
}
