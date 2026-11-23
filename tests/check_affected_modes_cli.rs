//! SPEC-20261123-check-precise-selection: `check affected --list` picks a precise set with
//! `--lane` and the full set without it. Every case builds a temporary git repository and reads
//! the command's answer; the real HOME, the real tools and the machine's heavy slots are not used.
mod common;
use common::git::{git, git_ok};
use serde_json::Value;
use std::collections::BTreeSet;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;

fn put(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

/// A repository holding `files`, one commit on top of it that writes `changed`, and the range
/// between them.
struct Repo {
    dir: tempfile::TempDir,
    range: String,
}

fn commit_all(p: &Path, msg: &str) -> String {
    git_ok(p, &["add", "-A"]);
    git_ok(p, &["commit", "-q", "-m", msg]);
    git(p, &["rev-parse", "HEAD"]).trim().to_string()
}

fn init(p: &Path) {
    git_ok(p, &["init", "-q", "-b", "main"]);
    git_ok(p, &["config", "user.email", "t@example.invalid"]);
    git_ok(p, &["config", "user.name", "t"]);
}

impl Repo {
    fn new(files: &[(&str, &str)], changed: &[(&str, &str)]) -> Repo {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        init(p);
        put(p, "README.md", "x\n");
        for (rel, text) in files {
            put(p, rel, text);
        }
        let base = commit_all(p, "base");
        for (rel, text) in changed {
            put(p, rel, text);
        }
        let head = commit_all(p, "change");
        Repo {
            dir,
            range: format!("{base}..{head}"),
        }
    }

    fn root(&self) -> PathBuf {
        self.dir.path().to_path_buf()
    }
}

/// The `--list` answer for `range` in `root`; a refused command fails the test with its text.
fn list_in(root: &Path, range: &str, extra: &[&str]) -> Value {
    let out = Command::new(BIN)
        .args(["check", "affected", range, "--list", "--dir"])
        .arg(root)
        .args(extra)
        .env_remove("MOAI_HOME")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let text = String::from_utf8_lossy(&out.stdout).to_string();
    let err = String::from_utf8_lossy(&out.stderr).to_string();
    let v: Value = serde_json::from_str(text.lines().next().unwrap_or(""))
        .unwrap_or_else(|_| panic!("`check affected {extra:?}` gave no answer: {text} / {err}"));
    assert_eq!(
        v["ok"],
        Value::Bool(true),
        "`check affected {extra:?}` refused: {v}"
    );
    v["result"].clone()
}

fn list(repo: &Repo, extra: &[&str]) -> Value {
    list_in(&repo.root(), &repo.range, extra)
}

fn rust(v: &Value) -> BTreeSet<String> {
    v["rust_targets"]
        .as_array()
        .unwrap_or_else(|| panic!("no rust_targets in {v}"))
        .iter()
        .map(|s| s.as_str().unwrap().to_string())
        .collect()
}

fn lane_and_full(repo: &Repo) -> (BTreeSet<String>, BTreeSet<String>) {
    let lane = rust(&list(repo, &["--lane"]));
    let full = rust(&list(repo, &[]));
    assert!(
        lane.is_subset(&full),
        "the full set must hold every lane target: lane {lane:?}, full {full:?}"
    );
    (lane, full)
}

const PLAIN: &str = "fn main() {}\n";

/// Three scans: one scoped to `src/update/`, one scoped to the docs folder with a rule file
/// under `tools/`, one without a scope.
const SCAN_LIST: &str = r#"{"scans": [
  {"id": "update-scan", "rules": ["tests/update_switch_layer.rs"], "scope": [{"prefix": "src/update/", "suffixes": [".rs"]}], "checks": ["cargo test --offline --locked --test update_switch_layer"]},
  {"id": "docs-scan", "rules": ["tools/doc_walk.sh"], "scope": [{"prefix": "docs/", "suffixes": [".md"]}], "checks": ["cargo test --offline --locked --test doc_references"]},
  {"id": "bare-scan", "checks": ["cargo test --offline --locked --test other_plain"]}
]}"#;

fn scan_files() -> Vec<(&'static str, &'static str)> {
    vec![
        ("tools/repo_scans.json", SCAN_LIST),
        ("tools/doc_walk.sh", "#\n"),
        ("src/update/switch.rs", "// a\n"),
        ("src/card.rs", "// a\n"),
        ("tests/update_switch_layer.rs", PLAIN),
        ("tests/doc_references.rs", PLAIN),
        ("tests/other_plain.rs", PLAIN),
    ]
}

// AC-1: a scan whose scope misses the diff, and a scan without a scope, leave the lane set;
// the full set keeps every scan target.
#[test]
fn scan_scope_decides_the_lane_and_the_full_set_keeps_every_scan() {
    let repo = Repo::new(&scan_files(), &[("src/update/switch.rs", "// b\n")]);
    let (lane, full) = lane_and_full(&repo);
    assert!(
        lane.contains("update_switch_layer"),
        "the scan whose scope holds src/update/switch.rs stays: {lane:?}"
    );
    assert!(
        !lane.contains("doc_references"),
        "the docs scan's scope misses the diff: {lane:?}"
    );
    assert!(
        !lane.contains("other_plain"),
        "a scan without a scope leaves the lane: {lane:?}"
    );
    for want in ["update_switch_layer", "doc_references", "other_plain"] {
        assert!(full.contains(want), "the full set keeps {want}: {full:?}");
    }
}

// AC-1: a changed rule file keeps its scan in the lane although its scope misses the diff.
#[test]
fn scan_rule_file_keeps_its_scan_in_the_lane() {
    let repo = Repo::new(&scan_files(), &[("tools/doc_walk.sh", "# b\n")]);
    let (lane, _) = lane_and_full(&repo);
    assert!(
        lane.contains("doc_references"),
        "the docs scan's rule file changed: {lane:?}"
    );
    assert!(
        !lane.contains("update_switch_layer"),
        "the update scan is neither in scope nor in its rules: {lane:?}"
    );
}

fn common_files() -> Vec<(&'static str, &'static str)> {
    vec![
        ("tests/common/mod.rs", "pub mod git;\npub mod home;\n"),
        ("tests/common/git.rs", "pub fn g() {}\n"),
        ("tests/common/home.rs", "pub fn h() {}\n"),
        (
            "tests/alpha_cli.rs",
            "mod common;\nuse common::git::g;\nfn main() { g() }\n",
        ),
        (
            "tests/beta_home.rs",
            "mod common;\nfn main() { crate::common::home::h() }\n",
        ),
        ("tests/gamma_plain.rs", PLAIN),
    ]
}

// AC-2: a change to one helper sub-module selects only the tests that name it in the lane;
// the full set keeps every consumer.
#[test]
fn common_sub_module_change_selects_only_its_users_in_the_lane() {
    let repo = Repo::new(
        &common_files(),
        &[("tests/common/home.rs", "pub fn h() {}\n// b\n")],
    );
    let (lane, full) = lane_and_full(&repo);
    assert!(lane.contains("beta_home"), "the home user stays: {lane:?}");
    assert!(
        !lane.contains("alpha_cli"),
        "alpha_cli names only common::git: {lane:?}"
    );
    for want in ["alpha_cli", "beta_home"] {
        assert!(full.contains(want), "the full set keeps {want}: {full:?}");
    }
    assert!(
        !full.contains("gamma_plain"),
        "gamma_plain uses no helper: {full:?}"
    );
}

// AC-2: a change to the helper's mod.rs keeps every consumer in the lane.
#[test]
fn common_mod_change_keeps_every_consumer_in_the_lane() {
    let repo = Repo::new(
        &common_files(),
        &[("tests/common/mod.rs", "pub mod git;\npub mod home;\n// b\n")],
    );
    let (lane, _) = lane_and_full(&repo);
    for want in ["alpha_cli", "beta_home"] {
        assert!(lane.contains(want), "mod.rs reaches {want}: {lane:?}");
    }
}

fn path_import_files() -> Vec<(&'static str, &'static str)> {
    vec![
        ("tests/common/mod.rs", "pub mod git;\n"),
        ("tests/common/git.rs", "pub fn g() {}\n"),
        ("tests/common/spawn_cap.rs", "pub fn s() {}\n"),
        ("tests/common/deep/x.rs", "pub fn d() {}\n"),
        (
            "tests/alpha_cli.rs",
            "mod common;\nuse common::git::g;\nfn main() { g() }\n",
        ),
        (
            "tests/path_user.rs",
            "#[path = \"common/spawn_cap.rs\"]\nmod spawn_cap;\nfn main() { spawn_cap::s() }\n",
        ),
        (
            "tests/renamed_user.rs",
            "#[path = \"common/spawn_cap.rs\"]\nmod cap;\nfn main() { cap::s() }\n",
        ),
        (
            "tests/stem_user.rs",
            "mod spawn_cap; // helper\nfn main() { spawn_cap::s() }\n",
        ),
        (
            "tests/deep_user.rs",
            "#[path = \"common/deep/x.rs\"]\nmod x;\nfn main() { x::d() }\n",
        ),
        ("tests/gamma_plain.rs", PLAIN),
    ]
}

// AC-9: a helper that a test imports by `#[path = "common/<file>"]` (or a module named like the
// file) reaches that test in both selections; the lane adds no other consumer of `tests/common`.
#[test]
fn common_path_import_keeps_its_users_in_both_selections() {
    let repo = Repo::new(
        &path_import_files(),
        &[("tests/common/spawn_cap.rs", "pub fn s() {}\n// b\n")],
    );
    let (lane, full) = lane_and_full(&repo);
    for (label, set) in [("lane", &lane), ("full", &full)] {
        for want in ["path_user", "renamed_user", "stem_user"] {
            assert!(set.contains(want), "{label} keeps {want}: {set:?}");
        }
        for not in ["gamma_plain", "deep_user"] {
            assert!(!set.contains(not), "{label} leaves {not} out: {set:?}");
        }
    }
    assert!(
        !lane.contains("alpha_cli"),
        "the lane leaves a `mod common` consumer that does not name the file: {lane:?}"
    );
    assert!(
        full.contains("alpha_cli"),
        "the full set keeps every `mod common` consumer: {full:?}"
    );
}

// AC-10: a deeper helper path keeps every `mod common` consumer in the lane, adds the tests that
// import that very file by path, and leaves the other path importers out.
#[test]
fn common_deeper_path_keeps_every_consumer_and_its_path_users() {
    let repo = Repo::new(
        &path_import_files(),
        &[("tests/common/deep/x.rs", "pub fn d() {}\n// b\n")],
    );
    let (lane, full) = lane_and_full(&repo);
    for (label, set) in [("lane", &lane), ("full", &full)] {
        for want in ["alpha_cli", "deep_user"] {
            assert!(set.contains(want), "{label} keeps {want}: {set:?}");
        }
        for not in ["path_user", "renamed_user", "stem_user", "gamma_plain"] {
            assert!(!set.contains(not), "{label} leaves {not} out: {set:?}");
        }
    }
}

// AC-3: a test tagged with the host OS joins the lane only when its first word names a changed
// area; the full set keeps it.
#[test]
fn os_tagged_test_joins_the_lane_by_area_only() {
    let tagged = format!(
        "#[cfg(target_os = \"{}\")]\nfn main() {{}}\n",
        std::env::consts::OS
    );
    let files = vec![
        ("src/update/switch.rs", "// a\n"),
        ("src/statusline.rs", "// a\n"),
        ("tests/statusline_host.rs", tagged.as_str()),
        ("tests/update_host.rs", tagged.as_str()),
    ];
    let repo = Repo::new(&files, &[("src/update/switch.rs", "// b\n")]);
    let (lane, full) = lane_and_full(&repo);
    assert!(
        lane.contains("update_host"),
        "update_host names the changed area: {lane:?}"
    );
    assert!(
        !lane.contains("statusline_host"),
        "statusline is not a changed area: {lane:?}"
    );
    assert!(
        full.contains("statusline_host"),
        "the full set keeps every host-tagged test: {full:?}"
    );
}

// AC-4: the answer names the selection.
#[test]
fn selection_field_names_lane_or_full() {
    let repo = Repo::new(&scan_files(), &[("src/card.rs", "// b\n")]);
    assert_eq!(list(&repo, &["--lane"])["selection"], "lane");
    assert_eq!(list(&repo, &[])["selection"], "full");
}

fn text_files() -> Vec<(&'static str, &'static str)> {
    vec![
        ("skills/moai/workflows/x.md", "a\n"),
        ("docs/a.md", "a\n"),
        (
            "tests/workflow_words.rs",
            "fn main() { let _ = \"skills/moai/workflows/x.md\"; }\n",
        ),
        (
            "tests/page_words.rs",
            "fn main() { let _ = \"docs/a.md\"; }\n",
        ),
        ("tests/other_words.rs", PLAIN),
    ]
}

// AC-5: a changed skill or doc file selects the tests whose source names it, in both selections.
#[test]
fn text_change_selects_the_tests_that_name_the_file() {
    let repo = Repo::new(&text_files(), &[("skills/moai/workflows/x.md", "b\n")]);
    let (lane, full) = lane_and_full(&repo);
    for set in [&lane, &full] {
        assert!(
            set.contains("workflow_words"),
            "workflow_words names the changed skill file: {set:?}"
        );
        assert!(
            !set.contains("page_words") && !set.contains("other_words"),
            "only the test naming the file joins: {set:?}"
        );
    }
    let repo = Repo::new(&text_files(), &[("docs/a.md", "b\n")]);
    let (lane, _) = lane_and_full(&repo);
    assert!(
        lane.contains("page_words"),
        "page_words names the changed doc: {lane:?}"
    );
}

/// A branch forked at `fork`, the integration branch moved by one commit, and that move merged
/// into the branch: `(repo root, fork, tip, merge)`.
fn merged_tip_repo() -> (tempfile::TempDir, String, String, String) {
    let dir = tempfile::tempdir().unwrap();
    let p = dir.path();
    init(p);
    put(p, "README.md", "x\n");
    put(p, "src/card.rs", "// a\n");
    let fork = commit_all(p, "fork");
    git_ok(p, &["switch", "-q", "-c", "feature"]);
    put(p, "src/card.rs", "// b\n");
    commit_all(p, "feature");
    git_ok(p, &["switch", "-q", "main"]);
    put(p, "src/tip_one.rs", "// t\n");
    put(p, "src/tip_two.rs", "// t\n");
    let tip = commit_all(p, "tip");
    git_ok(p, &["switch", "-q", "feature"]);
    git_ok(p, &["merge", "-q", "--no-edit", "main"]);
    let merge = git(p, &["rev-parse", "HEAD"]).trim().to_string();
    (dir, fork, tip, merge)
}

// AC-6: a range from the old fork point after the tip was merged in carries a range note; a
// range from the merge base carries none.
#[test]
fn range_from_an_old_fork_point_carries_a_note() {
    let (dir, fork, tip, merge) = merged_tip_repo();
    let wide = list_in(
        dir.path(),
        &format!("{fork}..{merge}"),
        &["--integration-ref", "main"],
    );
    let note = &wide["range_note"];
    assert!(note.is_object(), "a widened range is noted: {wide}");
    assert_eq!(note["merge_base"], Value::String(tip.clone()), "{note}");
    assert_eq!(note["extra_changed"], 2, "the tip's two files: {note}");
    let narrow = list_in(
        dir.path(),
        &format!("{tip}..{merge}"),
        &["--integration-ref", "main"],
    );
    assert!(
        narrow.get("range_note").is_none_or(Value::is_null),
        "a range from the merge base is not noted: {narrow}"
    );
}

// AC-7: the operations page describes the two selections and the range note.
#[test]
fn page_describes_the_two_selections() {
    let page = fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("docs/design-review/appendix/operations/check.md"),
    )
    .unwrap();
    for word in [
        "--lane",
        "selection",
        "range_note",
        "names that path",
        "path-import rule",
    ] {
        assert!(page.contains(word), "check.md does not say `{word}`");
    }
}
