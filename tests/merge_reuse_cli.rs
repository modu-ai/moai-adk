//! SPEC-20261120-merge-reuse: `moai check reuse` decides, from git alone, whether the evidence of
//! an audited merge candidate still holds after the local integration branch moved. Temporary
//! git repositories only: the real tools, the real HOME and the machine's slots are never used.
mod common;
use common::git::{git, git_ok};
use serde_json::Value;
use std::fs;
use std::path::Path;
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;

const SCANS: &str = r#"{"about":"fixture","scans":[
  {"id":"home","checks":["cargo test --offline --locked --test home_isolation_scan"]}]}"#;

const SCANS_WITH_OTHER: &str = r#"{"about":"fixture","scans":[
  {"id":"home","checks":["cargo test --offline --locked --test home_isolation_scan"]},
  {"id":"other","checks":["cargo test --offline --locked --test other_cli"]}]}"#;

/// Ten lines, so that two edits far apart merge without a conflict.
const MODELS: &str = "// 1\n// 2\n// 3\n// 4\n// 5\n// 6\n// 7\n// 8\n// 9\n// 10\n";

fn put(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

struct Fixture {
    dir: tempfile::TempDir,
    base: String,
}

impl Fixture {
    fn new() -> Fixture {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        git_ok(p, &["init", "-q", "-b", "main"]);
        git_ok(p, &["config", "user.email", "t@example.invalid"]);
        git_ok(p, &["config", "user.name", "t"]);
        put(p, "README.md", "x\n");
        put(p, "src/update_models.rs", MODELS);
        put(p, "src/ops/alpha.rs", "// a\n");
        put(p, "src/other.rs", "// o\n");
        put(p, "tests/update_command_cli.rs", "fn main() {}\n");
        put(p, "tests/other_cli.rs", "fn main() {}\n");
        put(p, "tests/home_isolation_scan.rs", "fn main() {}\n");
        put(p, "tests/update_tux_tty_test.py", "#\n");
        put(p, "tools/repo_scans.json", SCANS);
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "base"]);
        let base = git(p, &["rev-parse", "HEAD"]);
        Fixture { dir, base }
    }

    fn root(&self) -> &Path {
        self.dir.path()
    }

    /// A commit on a new branch off `from` that writes `files`; the checkout goes back to `main`.
    fn branch(&self, name: &str, from: &str, files: &[(&str, &str)]) -> String {
        let p = self.root();
        git_ok(p, &["checkout", "-q", "-b", name, from]);
        for (rel, text) in files {
            put(p, rel, text);
        }
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", name]);
        let sha = git(p, &["rev-parse", "HEAD"]);
        git_ok(p, &["checkout", "-q", "main"]);
        sha
    }

    /// A commit on a new branch off `from` that deletes `paths`; the checkout goes back to `main`.
    fn branch_without(&self, name: &str, from: &str, paths: &[&str]) -> String {
        let p = self.root();
        git_ok(p, &["checkout", "-q", "-b", name, from]);
        for rel in paths {
            git_ok(p, &["rm", "-q", rel]);
        }
        git_ok(p, &["commit", "-q", "-m", name]);
        let sha = git(p, &["rev-parse", "HEAD"]);
        git_ok(p, &["checkout", "-q", "main"]);
        sha
    }

    /// The candidate: edits line 1 of the models file.
    fn candidate(&self) -> String {
        let edited = MODELS.replacen("// 1\n", "// 1 edited\n", 1);
        self.branch("cand", &self.base, &[("src/update_models.rs", &edited)])
    }
}

fn reuse_args(base: &str, head: &str, tip: &str) -> Vec<String> {
    [
        "check", "reuse", "--base", base, "--head", head, "--tip", tip,
    ]
    .map(String::from)
    .to_vec()
}

fn run(fx: &Fixture, args: &[String]) -> (bool, Value, String) {
    let out = Command::new(BIN)
        .args(args)
        .arg("--dir")
        .arg(fx.root())
        .env_remove("MOAI_HOME")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let text = String::from_utf8_lossy(&out.stdout).to_string();
    let v: Value = serde_json::from_str(&text)
        .unwrap_or_else(|_| panic!("{text} / {}", String::from_utf8_lossy(&out.stderr)));
    let ok = v["ok"] == Value::Bool(true);
    let body = if ok {
        v["result"].clone()
    } else {
        v["error"].clone()
    };
    (ok, body, text)
}

fn decide(fx: &Fixture, base: &str, head: &str, tip: &str) -> Value {
    let (ok, body, text) = run(fx, &reuse_args(base, head, tip));
    assert!(ok, "{text}");
    body
}

fn names(v: &Value) -> Vec<String> {
    v.as_array()
        .unwrap()
        .iter()
        .map(|s| s.as_str().unwrap().to_string())
        .collect()
}

#[test]
fn decides_fast_forward_when_the_tip_is_still_the_base() {
    let fx = Fixture::new();
    let head = fx.candidate();
    let a = decide(&fx, &fx.base, &head, &fx.base);
    assert_eq!(a["decision"], "fast_forward", "{a}");
    assert_eq!(names(&a["rerun"]["rust"]), Vec::<String>::new());
    assert_eq!(a["base"], fx.base);
    assert_eq!(a["head"], head);
}

#[test]
fn decides_already_merged_when_the_head_is_inside_the_tip() {
    let fx = Fixture::new();
    let head = fx.candidate();
    // The tip moved on and contains the candidate.
    let tip = fx.branch("tip", &head, &[("src/other.rs", "// o2\n")]);
    let a = decide(&fx, &fx.base, &head, &tip);
    assert_eq!(a["decision"], "already_merged", "{a}");
}

#[test]
fn decides_reuse_after_an_unrelated_base_move() {
    let fx = Fixture::new();
    let head = fx.candidate();
    // Another merge changed a source file of another area: no shared file, same selection.
    let tip = fx.branch("tip", &fx.base, &[("src/other.rs", "// o2\n")]);
    let a = decide(&fx, &fx.base, &head, &tip);
    assert_eq!(a["decision"], "reuse", "{a}");
    let reasons = names(&a["reasons"]);
    for want in [
        "base_is_ancestor_of_tip",
        "merge_is_clean",
        "no_shared_path",
        "selection_is_equal",
    ] {
        assert!(reasons.contains(&want.to_string()), "{want}: {reasons:?}");
    }
    assert_eq!(names(&a["rerun"]["rust"]), Vec::<String>::new(), "{a}");
    assert_eq!(names(&a["rerun"]["steps"]), Vec::<String>::new(), "{a}");
    // The candidate's selection names what its evidence covers.
    assert_eq!(
        names(&a["selection"]["new"]["rust"]),
        ["home_isolation_scan", "update_command_cli"]
    );
}

#[test]
fn decides_reuse_when_the_tip_only_removed_a_target_the_candidate_ran() {
    let fx = Fixture::new();
    let head = fx.candidate();
    // The tip deleted a test target of the candidate's area: the new selection is smaller, so
    // every check the new selection names was already run.
    let tip = fx.branch_without("tip", &fx.base, &["tests/update_command_cli.rs"]);
    let a = decide(&fx, &fx.base, &head, &tip);
    assert_eq!(a["decision"], "reuse", "{a}");
    assert!(
        names(&a["reasons"]).contains(&"selection_is_covered".to_string()),
        "{a}"
    );
    assert_eq!(
        names(&a["selection"]["old"]["rust"]),
        ["home_isolation_scan", "update_command_cli"]
    );
    assert_eq!(
        names(&a["selection"]["new"]["rust"]),
        ["home_isolation_scan"]
    );
    assert_eq!(names(&a["rerun"]["rust"]), Vec::<String>::new(), "{a}");
}

#[test]
fn reruns_a_new_python_module_of_the_candidates_area() {
    let fx = Fixture::new();
    let head = fx.candidate();
    let tip = fx.branch("tip", &fx.base, &[("tests/update_new_test.py", "#\n")]);
    let a = decide(&fx, &fx.base, &head, &tip);
    assert_eq!(a["decision"], "rerun", "{a}");
    assert_eq!(names(&a["rerun"]["python"]), ["update_new_test"], "{a}");
    assert_eq!(names(&a["rerun"]["rust"]), Vec::<String>::new(), "{a}");
}

#[test]
fn reruns_a_new_test_target_of_the_candidates_area() {
    let fx = Fixture::new();
    let head = fx.candidate();
    let tip = fx.branch(
        "tip",
        &fx.base,
        &[("tests/update_extra_cli.rs", "fn main() {}\n")],
    );
    let a = decide(&fx, &fx.base, &head, &tip);
    assert_eq!(a["decision"], "rerun", "{a}");
    assert_eq!(names(&a["rerun"]["rust"]), ["update_extra_cli"], "{a}");
    assert_eq!(names(&a["rerun"]["python"]), Vec::<String>::new(), "{a}");
    assert_eq!(names(&a["rerun"]["steps"]), Vec::<String>::new(), "{a}");
}

#[test]
fn reruns_a_changed_test_of_the_candidates_area() {
    let fx = Fixture::new();
    let head = fx.candidate();
    let tip = fx.branch(
        "tip",
        &fx.base,
        &[("tests/update_command_cli.rs", "fn main() { /* more */ }\n")],
    );
    let a = decide(&fx, &fx.base, &head, &tip);
    assert_eq!(a["decision"], "rerun", "{a}");
    assert_eq!(names(&a["rerun"]["rust"]), ["update_command_cli"], "{a}");
}

#[test]
fn reruns_the_new_targets_of_a_changed_scan_list() {
    let fx = Fixture::new();
    let head = fx.candidate();
    let tip = fx.branch(
        "tip",
        &fx.base,
        &[("tools/repo_scans.json", SCANS_WITH_OTHER)],
    );
    let a = decide(&fx, &fx.base, &head, &tip);
    assert_eq!(a["decision"], "rerun", "{a}");
    assert_eq!(names(&a["rerun"]["rust"]), ["other_cli"], "{a}");
    assert_eq!(a["selection"]["old"]["scans"], "listed:1");
    assert_eq!(a["selection"]["new"]["scans"], "listed:2");
}

#[test]
fn reruns_the_whole_change_when_the_base_move_shares_a_file() {
    let fx = Fixture::new();
    let head = fx.candidate();
    // The same file, another place: the merge is clean, the file is shared.
    let moved = MODELS.replacen("// 9\n", "// 9 moved\n", 1);
    let tip = fx.branch("tip", &fx.base, &[("src/update_models.rs", &moved)]);
    let a = decide(&fx, &fx.base, &head, &tip);
    assert_eq!(a["decision"], "rerun", "{a}");
    assert_eq!(a["overlap"], serde_json::json!(["src/update_models.rs"]));
    assert_eq!(
        names(&a["rerun"]["steps"]),
        ["diff_check", "fmt", "clippy", "test_lib", "test_bins"],
        "{a}"
    );
    assert_eq!(names(&a["rerun"]["rust"]), ["update_command_cli"], "{a}");
}

#[test]
fn reruns_all_when_the_merge_conflicts() {
    let fx = Fixture::new();
    let head = fx.candidate();
    let clash = MODELS.replacen("// 1\n", "// 1 other edit\n", 1);
    let tip = fx.branch("tip", &fx.base, &[("src/update_models.rs", &clash)]);
    let a = decide(&fx, &fx.base, &head, &tip);
    assert_eq!(a["decision"], "rerun_all", "{a}");
    assert!(
        names(&a["reasons"]).contains(&"merge_conflict".to_string()),
        "{a}"
    );
}

#[test]
fn reruns_all_when_the_base_is_not_an_ancestor_of_the_tip() {
    let fx = Fixture::new();
    let head = fx.candidate();
    // A tip with no shared history (a rewritten branch).
    let p = fx.root();
    git_ok(p, &["checkout", "-q", "--orphan", "rewritten"]);
    git_ok(p, &["rm", "-rfq", "."]);
    put(p, "README.md", "rewritten\n");
    git_ok(p, &["add", "-A"]);
    git_ok(p, &["commit", "-q", "-m", "rewritten"]);
    let tip = git(p, &["rev-parse", "HEAD"]);
    git_ok(p, &["checkout", "-q", "main"]);
    let a = decide(&fx, &fx.base, &head, &tip);
    assert_eq!(a["decision"], "rerun_all", "{a}");
    assert!(
        names(&a["reasons"]).contains(&"base_not_ancestor_of_tip".to_string()),
        "{a}"
    );
}

#[test]
fn answers_without_moving_a_ref_a_file_or_the_index() {
    let fx = Fixture::new();
    let head = fx.candidate();
    let tip = fx.branch("tip", &fx.base, &[("src/other.rs", "// o2\n")]);
    let p = fx.root();
    let snapshot = |p: &Path| {
        (
            git(p, &["for-each-ref"]),
            git(p, &["rev-parse", "HEAD"]),
            git(p, &["status", "--porcelain=v1", "--untracked-files=all"]),
            git(p, &["diff", "--cached", "--stat"]),
            fs::read_to_string(p.join("src/update_models.rs")).unwrap(),
        )
    };
    let before = snapshot(p);
    for _ in 0..2 {
        decide(&fx, &fx.base, &head, &tip);
    }
    assert_eq!(snapshot(p), before);
}

#[test]
fn answers_for_a_merge_commit_like_for_its_two_parents() {
    let fx = Fixture::new();
    let head = fx.candidate();
    let p = fx.root();
    git_ok(p, &["checkout", "-q", "-b", "m", &fx.base]);
    git_ok(p, &["merge", "-q", "--no-ff", "-m", "merge cand", &head]);
    let merge = git(p, &["rev-parse", "HEAD"]);
    git_ok(p, &["checkout", "-q", "main"]);
    let tip = fx.branch("tip", &fx.base, &[("src/other.rs", "// o2\n")]);
    let by_parents = decide(&fx, &fx.base, &head, &tip);
    let args: Vec<String> = [
        "check",
        "reuse",
        "--merge",
        merge.as_str(),
        "--tip",
        tip.as_str(),
    ]
    .map(String::from)
    .to_vec();
    let (ok, by_merge, text) = run(&fx, &args);
    assert!(ok, "{text}");
    assert_eq!(by_merge, by_parents);
    // A commit that is not a merge of two parents is refused.
    let args: Vec<String> = [
        "check",
        "reuse",
        "--merge",
        head.as_str(),
        "--tip",
        tip.as_str(),
    ]
    .map(String::from)
    .to_vec();
    let (ok, body, _) = run(&fx, &args);
    assert!(!ok, "{body}");
}

#[test]
fn answers_a_refusal_for_a_missing_or_dashed_argument() {
    let fx = Fixture::new();
    let head = fx.candidate();
    // The same command line with valid arguments works, so the refusals below are about the
    // arguments and not about a word that does not exist.
    assert_eq!(
        decide(&fx, &fx.base, &head, &fx.base)["decision"],
        "fast_forward"
    );
    for args in [
        vec!["check", "reuse"],
        vec!["check", "reuse", "--base", &fx.base, "--head", &head],
        vec![
            "check", "reuse", "--base", "-x", "--head", &head, "--tip", &fx.base,
        ],
        vec![
            "check",
            "reuse",
            "--base",
            &fx.base,
            "--head",
            &head,
            "--tip",
            "no-such-ref",
        ],
    ] {
        let args: Vec<String> = args.into_iter().map(String::from).collect();
        let out = Command::new(BIN)
            .args(&args)
            .arg("--dir")
            .arg(fx.root())
            .env_remove("MOAI_HOME")
            .stdin(Stdio::null())
            .output()
            .unwrap();
        assert!(!out.status.success(), "{args:?}");
        let text = String::from_utf8_lossy(&out.stdout);
        assert!(!text.contains("\"decision\""), "{args:?}: {text}");
    }
}

#[test]
fn answers_in_a_fixed_order() {
    let fx = Fixture::new();
    let head = fx.candidate();
    let tip = fx.branch(
        "tip",
        &fx.base,
        &[
            ("tests/update_b_cli.rs", "fn main() {}\n"),
            ("tests/update_a_cli.rs", "fn main() {}\n"),
            ("tests/update_tux2_test.py", "#\n"),
        ],
    );
    let (_, first, first_text) = run(&fx, &reuse_args(&fx.base, &head, &tip));
    let (_, _, second_text) = run(&fx, &reuse_args(&fx.base, &head, &tip));
    assert_eq!(first_text, second_text);
    assert_eq!(first["decision"], "rerun");
    assert_eq!(
        names(&first["rerun"]["rust"]),
        ["update_a_cli", "update_b_cli"]
    );
    assert_eq!(names(&first["rerun"]["python"]), ["update_tux2_test"]);
}
