//! `moai-cli repo-scans`, the judge of `tools/repo_scans.json` (#2681, #2693, #2708).
//!
//! Each case runs the built binary: it reads the change sets of a pull request and of the tip
//! (on stdin, or from git with `--refs`) and says whether the merge needs a re-merge and which
//! checks to run. The judge always exits 0 with a report, except an unknown option, which is a
//! usage error (exit 2) with no report. Every run uses a temporary HOME. Moved here from the
//! Python module `repo_scans_test.py` (SPEC-20261120-python-binary-tests-to-rust); the six table
//! cases of that module (`RepoScansTableTest`, which read `tools/repo_scans.json` and scan the
//! repository's test files) stay in Python because they test the table, not the binary.

mod common;
use common::git::git_in;
use common::pty::{env_with, piped_with, Env, Seen};
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
};
use tempfile::TempDir;

const ROOT: &str = env!("CARGO_MANIFEST_DIR");
const INCIDENT_CHECK: &str =
    "cargo test --offline --locked --lib permission_reference_rule_has_one_copy";

// The incident (#2681 body): #2642 added a walk of src/**/*.rs, #2632 added a file it rejects.
const PR_2642: [&str; 20] = [
    "src/agents.rs",
    "src/apply.rs",
    "src/archive.rs",
    "src/claude.rs",
    "src/contract.rs",
    "src/decision.rs",
    "src/decision_log.rs",
    "src/forge/audit_status.rs",
    "src/forge/report.rs",
    "src/generation.rs",
    "src/github_templates.rs",
    "src/hooks.rs",
    "src/key_vault.rs",
    "src/layout.rs",
    "src/message.rs",
    "src/migrate.rs",
    "src/profile.rs",
    "src/recovery.rs",
    "src/update.rs",
    "src/worktree_gc.rs",
];
const PR_2632: [&str; 9] = [
    "docs/design-review/appendix/operations/agents_home_apply.md",
    "docs/design-review/appendix/operations/agents_home_plan.md",
    "docs/design-review/appendix/operations/agents_home_revert.md",
    "src/agent_install.rs",
    "src/lib.rs",
    "src/ops/agents_home_apply.rs",
    "src/ops/agents_home_plan.rs",
    "src/ops/agents_home_revert.rs",
    "tests/agents_home_cli.rs",
];

fn table() -> PathBuf {
    Path::new(ROOT).join("tools").join("repo_scans.json")
}

/// The scans of the checked-in table.
fn load() -> Vec<Value> {
    let text = fs::read_to_string(table()).unwrap();
    serde_json::from_str::<Value>(&text).unwrap()["scans"]
        .as_array()
        .unwrap()
        .clone()
}

struct Home(TempDir);

impl Home {
    fn new() -> Home {
        Home(TempDir::new().unwrap())
    }

    fn env(&self, extra: &[(&str, &str)]) -> Env {
        env_with(
            &[
                ("PATH", "/usr/bin:/bin"),
                ("HOME", self.0.path().to_str().unwrap()),
                ("LANG", "en_US.UTF-8"),
                ("GIT_CONFIG_GLOBAL", "/dev/null"),
                ("GIT_CONFIG_NOSYSTEM", "1"),
                ("GIT_AUTHOR_DATE", "2026-11-13T00:00:00Z"),
                ("GIT_COMMITTER_DATE", "2026-11-13T00:00:00Z"),
            ],
            extra,
        )
    }
}

/// The raw run of `moai-cli repo-scans --repo <cwd> --table <table> <args>`.
fn run_rust(home: &Home, args: &[&str], stdin: &str, cwd: &Path, table: &Path) -> Seen {
    let mut all = vec![
        "repo-scans",
        "--repo",
        cwd.to_str().unwrap(),
        "--table",
        table.to_str().unwrap(),
    ];
    all.extend_from_slice(args);
    piped_with(common::MOAI_CLI_BIN, cwd, &home.env(&[]), &all, Some(stdin))
}

/// The report (the envelope's `result`); the judge always exits 0.
fn judge_run(home: &Home, args: &[&str], stdin: &str, cwd: &Path, table: &Path) -> Value {
    let rs = run_rust(home, args, stdin, cwd, table);
    assert_eq!(
        rs.code,
        Some(0),
        "repo-scans: {:?} {:?}",
        rs.stdout,
        rs.stderr
    );
    let envelope: Value = serde_json::from_str(&rs.stdout).unwrap();
    assert_eq!(envelope["ok"], true, "exit 0 without ok: {:?}", rs.stdout);
    envelope["result"].clone()
}

fn judge_root(home: &Home, stdin: &Value) -> Value {
    judge_run(home, &[], &stdin.to_string(), Path::new(ROOT), &table())
}

fn judge(home: &Home, pr: &[&str], tip: &[&str]) -> (bool, Vec<Value>) {
    let report = judge_root(home, &json!({"pr": pr, "tip": tip}));
    (
        report["remerge"].as_bool().unwrap(),
        report["hits"].as_array().unwrap().clone(),
    )
}

fn ids(hits: &[Value]) -> Vec<String> {
    hits.iter()
        .map(|h| h["id"].as_str().unwrap().to_string())
        .collect()
}

const NOTHING: (bool, Vec<Value>) = (false, Vec::new());

// ---- the judge ----

#[test]
fn the_2642_2632_incident_requires_a_remerge_either_way_round() {
    let home = Home::new();
    for (pr, tip, rules_side) in [
        (&PR_2642[..], &PR_2632[..], "pr"),
        (&PR_2632[..], &PR_2642[..], "tip"),
    ] {
        let (remerge, hits) = judge(&home, pr, tip);
        assert!(remerge);
        let hit: Vec<&Value> = hits
            .iter()
            .filter(|h| h["id"] == "permission-reference-one-copy")
            .collect();
        assert_eq!(hit.len(), 1, "{hits:?}");
        assert_eq!(hit[0]["rules_side"], rules_side);
        assert_eq!(hit[0]["rules"], json!(["src/contract.rs"]));
        assert!(hit[0]["scope"]
            .as_array()
            .unwrap()
            .contains(&json!("src/agent_install.rs")));
    }
}

#[test]
fn scope_on_both_sides_without_a_rule_change_needs_no_remerge() {
    let home = Home::new();
    assert_eq!(
        judge(&home, &["src/agent_install.rs"], &["src/update_x.rs"]),
        NOTHING
    );
}

#[test]
fn a_rule_change_alone_needs_no_remerge() {
    let home = Home::new();
    assert_eq!(judge(&home, &["src/contract.rs"], &["README.md"]), NOTHING);
    assert_eq!(judge(&home, &["src/contract.rs"], &[]), NOTHING);
}

#[test]
fn scope_boundaries() {
    let home = Home::new();
    let permission = |tip: &str| -> bool {
        let (_, hits) = judge(&home, &["src/contract.rs"], &[tip]);
        ids(&hits).contains(&"permission-reference-one-copy".to_string())
    };
    // suffix filter: a non-.rs file under src/ is outside the permission scan
    assert!(!permission("src/notes.md"));
    // prefix is a path prefix: "srcx/" is not "src/"
    assert!(!permission("srcx/a.rs"));
    // one level deeper still counts
    assert!(permission("src/forge/new.rs"));
    // the prefix anchors at the start: "src/" inside a path is not "src/" (#2708 mutant M2)
    assert!(!permission("lib/src/a.rs"));
}

#[test]
fn hit_paths_are_sorted() {
    // input order does not reach the report (#2708 mutant M9)
    let home = Home::new();
    let (_, hits) = judge(
        &home,
        &["tools/repo_scans.json", "tests/repo_scans_test.py"],
        &["tests/zz_test.py", "src/b.rs", "src/a.rs"],
    );
    let hit = hits.iter().find(|h| h["id"] == "repo-scans-table").unwrap();
    assert_eq!(
        hit["rules"],
        json!(["tests/repo_scans_test.py", "tools/repo_scans.json"])
    );
    assert_eq!(
        hit["scope"],
        json!(["src/a.rs", "src/b.rs", "tests/zz_test.py"])
    );
}

#[test]
fn an_expected_set_file_counts_as_a_rule() {
    let home = Home::new();
    let (remerge, hits) = judge(
        &home,
        &["skills/moai/references/source-manifest.json"],
        &["skills/moai/workflows/plan.md"],
    );
    assert!(remerge);
    assert!(ids(&hits).contains(&"skill-package-files".to_string()));
}

#[test]
fn empty_input_needs_no_remerge() {
    let home = Home::new();
    assert_eq!(judge(&home, &[], &[]), NOTHING);
    assert_eq!(judge(&home, &[""], &[""]), NOTHING);
}

#[test]
fn checks_are_the_hit_scans_in_table_order_without_duplicates() {
    let home = Home::new();
    let report = judge_root(&home, &json!({"pr": PR_2642, "tip": PR_2632}));
    let checks: Vec<String> = report["checks"]
        .as_array()
        .unwrap()
        .iter()
        .map(|c| c.as_str().unwrap().to_string())
        .collect();
    assert!(checks.iter().any(|c| c == INCIDENT_CHECK), "{checks:?}");
    let mut seen = checks.clone();
    seen.sort();
    seen.dedup();
    assert_eq!(checks.len(), seen.len(), "duplicate check in {checks:?}");
    let order: Vec<String> = load()
        .iter()
        .flat_map(|s| s["checks"].as_array().unwrap().clone())
        .map(|c| c.as_str().unwrap().to_string())
        .collect();
    let place = |c: &String| order.iter().position(|o| o == c).unwrap();
    let mut sorted = checks.clone();
    sorted.sort_by_key(place);
    assert_eq!(checks, sorted);
    // two hits of one scan (both directions) give its checks once
    let both = judge_root(
        &home,
        &json!({"pr": ["tests/repo_scans_test.py"], "tip": ["tests/repo_scans_test.py"]}),
    );
    let table_hits = both["hits"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|h| h["id"] == "repo-scans-table")
        .count();
    assert_eq!(table_hits, 2, "{}", both["hits"]);
    let once = both["checks"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|c| *c == "python3 -m unittest tests.repo_scans_test")
        .count();
    assert_eq!(once, 1, "{}", both["checks"]);
    // no hit, no check
    assert_eq!(
        judge_root(&home, &json!({"pr": ["README.md"], "tip": []}))["checks"],
        json!([])
    );
}

// ---- the command ----

#[test]
fn incident_through_stdin() {
    let home = Home::new();
    let report = judge_root(&home, &json!({"pr": PR_2642, "tip": PR_2632}));
    assert_eq!(report["remerge"], true);
    assert!(ids(report["hits"].as_array().unwrap())
        .contains(&"permission-reference-one-copy".to_string()));
}

#[test]
fn disjoint_change_sets_pass() {
    let home = Home::new();
    assert_eq!(
        judge_root(&home, &json!({"pr": ["README.md"], "tip": ["notes/x.txt"]})),
        json!({"remerge": false, "hits": [], "checks": [], "reasons": []})
    );
}

#[test]
fn refs_read_the_change_sets_from_git() {
    // the same commit on both sides: nothing changed since the merge-base
    let home = Home::new();
    let report = judge_run(
        &home,
        &["--refs", "HEAD", "HEAD"],
        "",
        Path::new(ROOT),
        &table(),
    );
    assert_eq!(
        report,
        json!({"remerge": false, "hits": [], "checks": [], "reasons": []})
    );
}

#[test]
fn bad_input_fails_safe_with_every_check() {
    let home = Home::new();
    let mut every: Vec<Value> = Vec::new();
    for scan in load() {
        for check in scan["checks"].as_array().unwrap() {
            if !every.contains(check) {
                every.push(check.clone());
            }
        }
    }
    let cases: Vec<(String, Vec<&str>)> = vec![
        (String::new(), vec![]),
        ("not json".to_string(), vec![]),
        (json!({"pr": "src/a.rs", "tip": []}).to_string(), vec![]),
        (json!({"pr": []}).to_string(), vec![]),
        (String::new(), vec!["--refs", "HEAD"]),
        (String::new(), vec!["--refs", "no-such-ref-2681", "HEAD"]),
        (String::new(), vec!["--refs", "HEAD", "HEAD", "HEAD"]),
    ];
    for (stdin, args) in cases {
        let report = judge_run(&home, &args, &stdin, Path::new(ROOT), &table());
        let case = format!("{stdin:?} {args:?}");
        assert_eq!(report["remerge"], true, "{case}");
        assert_eq!(report["checks"], json!(every), "{case}");
        assert!(
            report["reasons"][0]
                .as_str()
                .unwrap()
                .starts_with("judge internal error"),
            "{case}: {}",
            report["reasons"]
        );
    }
}

#[test]
fn an_unknown_option_is_a_usage_error() {
    // an unknown option, or a ref that looks like one, is a usage error: no report at all
    let home = Home::new();
    for args in [vec!["--bogus"], vec!["--refs", "-x", "HEAD"]] {
        let rs = run_rust(&home, &args, "", Path::new(ROOT), &table());
        let envelope: Value = serde_json::from_str(&rs.stdout).unwrap();
        assert_eq!(rs.code, Some(2), "{args:?}");
        assert_eq!(envelope["ok"], false, "{args:?}");
        assert_eq!(envelope["error"]["kind"], "usage", "{args:?}");
        assert!(envelope.get("result").is_none(), "{args:?}");
    }
}

#[test]
fn two_scans_sharing_a_command() {
    // both scans hit; "y" once, at its first place
    let home = Home::new();
    let dir = TempDir::new().unwrap();
    let scope = json!([{"prefix": "s/"}]);
    let table = dir.path().join("repo_scans.json");
    fs::write(
        &table,
        json!({"scans": [
            {"id": "a", "rules": ["r"], "scope": scope, "checks": ["x", "y"]},
            {"id": "b", "rules": ["r"], "scope": scope, "checks": ["y", "z"]}]})
        .to_string(),
    )
    .unwrap();
    let report = judge_run(
        &home,
        &[],
        &json!({"pr": ["s/1"], "tip": ["r"]}).to_string(),
        dir.path(),
        &table,
    );
    assert_eq!(report["checks"], json!(["x", "y", "z"]));
    let hits: Vec<(String, String)> = report["hits"]
        .as_array()
        .unwrap()
        .iter()
        .map(|h| {
            (
                h["id"].as_str().unwrap().to_string(),
                h["rules_side"].as_str().unwrap().to_string(),
            )
        })
        .collect();
    assert_eq!(
        hits,
        [
            ("a".to_string(), "tip".to_string()),
            ("b".to_string(), "tip".to_string())
        ]
    );
}

// ---- --refs reads the table from both refs ----

/// #2693 audit B1: with `--refs` the table comes from both refs, not from the checkout that runs
/// the judge, so a scan added by the pull request itself is judged.
struct Refs {
    home: Home,
    repo: PathBuf,
    entry: Value,
    without: Vec<Value>,
}

impl Refs {
    fn new() -> Refs {
        let home = Home::new();
        let repo = fs::canonicalize(home.0.path()).unwrap().join("repo");
        fs::create_dir_all(&repo).unwrap();
        let full = load();
        let entry = full
            .iter()
            .find(|s| s["id"] == "permission-reference-one-copy")
            .unwrap()
            .clone();
        let without: Vec<Value> = full
            .iter()
            .filter(|s| s["id"] != "permission-reference-one-copy")
            .cloned()
            .collect();
        let fx = Refs {
            home,
            repo,
            entry,
            without,
        };
        git_in(&fx.repo, &["init", "-q", "-b", "base"]);
        fx.write("src/lib.rs", "");
        fx.write_table(&fx.without);
        fx.commit("base");
        // tip side = #2632: a new file in src/ (the scope)
        git_in(&fx.repo, &["checkout", "-q", "-b", "tip"]);
        fx.write("src/agent_install.rs", "// a copy of the rule\n");
        fx.commit("tip");
        // pull request side = #2642: the scanning test and its table entry (a branch named
        // "head" would be HEAD on a case-insensitive disk)
        git_in(&fx.repo, &["checkout", "-q", "-b", "pull", "base"]);
        fx.write("src/contract.rs", "// the walking test\n");
        let mut with = fx.without.clone();
        with.push(fx.entry.clone());
        fx.write_table(&with);
        fx.commit("head");
        fx
    }

    fn write(&self, rel: &str, text: &str) {
        let path = self.repo.join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, text).unwrap();
    }

    fn write_table(&self, scans: &[Value]) {
        self.write(
            "tools/repo_scans.json",
            &(serde_json::to_string_pretty(&json!({ "scans": scans })).unwrap() + "\n"),
        );
    }

    fn commit(&self, message: &str) {
        git_in(&self.repo, &["add", "-A"]);
        git_in(&self.repo, &["commit", "-q", "-m", message]);
    }

    /// Judge with the checkout's own table as the fallback, unless `table` is given.
    fn judge(&self, tip: &str, head: &str, table: Option<&Path>) -> Value {
        let own = self.repo.join("tools").join("repo_scans.json");
        judge_run(
            &self.home,
            &["--refs", tip, head],
            "",
            &self.repo,
            table.unwrap_or(&own),
        )
    }
}

fn has_check(report: &Value, check: &str) -> bool {
    report["checks"]
        .as_array()
        .unwrap()
        .iter()
        .any(|c| c == check)
}

#[test]
fn a_scan_the_pull_request_adds_is_judged_from_any_checkout() {
    let fx = Refs::new();
    for checkout in ["tip", "pull", "base"] {
        git_in(&fx.repo, &["checkout", "-q", checkout]);
        let report = fx.judge("tip", "pull", None);
        assert_eq!(report["remerge"], true, "{checkout}");
        assert!(
            ids(report["hits"].as_array().unwrap())
                .contains(&"permission-reference-one-copy".to_string()),
            "{checkout}"
        );
        assert!(has_check(&report, INCIDENT_CHECK), "{checkout}");
    }
}

#[test]
fn a_scan_only_the_tip_has_is_judged_too() {
    // reverse roles: the tip carries the scan, the pull request adds the file
    let fx = Refs::new();
    let report = fx.judge("pull", "tip", None);
    assert!(has_check(&report, INCIDENT_CHECK));
}

#[test]
fn refs_without_any_table_fall_back_to_the_given_table() {
    let fx = Refs::new();
    git_in(&fx.repo, &["checkout", "-q", "-b", "old", "base"]);
    git_in(&fx.repo, &["rm", "-q", "tools/repo_scans.json"]);
    fx.write("src/contract.rs", "// rule\n");
    fx.commit("old head without a table");
    git_in(&fx.repo, &["checkout", "-q", "-b", "old-tip", "base"]);
    git_in(&fx.repo, &["rm", "-q", "tools/repo_scans.json"]);
    fx.write("src/agent_install.rs", "// copy\n");
    fx.commit("old tip without a table");
    // the gate's own (current) table has the entry
    let report = fx.judge("old-tip", "old", Some(&table()));
    assert!(has_check(&report, INCIDENT_CHECK));
}

#[test]
fn a_broken_table_at_a_ref_fails_safe() {
    let fx = Refs::new();
    git_in(&fx.repo, &["checkout", "-q", "pull"]);
    fx.write("tools/repo_scans.json", "{ not json");
    fx.commit("broken table");
    let report = fx.judge("tip", "pull", Some(&table()));
    assert_eq!(report["remerge"], true);
    assert!(report["reasons"][0]
        .as_str()
        .unwrap()
        .starts_with("judge internal error"));
    assert!(has_check(&report, INCIDENT_CHECK));
}

#[test]
fn no_readable_table_at_all_says_hold() {
    // N1: no table to fall back to and bad input: nothing to run, hold the merge
    let fx = Refs::new();
    let missing = fx.home.0.path().join("lone").join("repo_scans.json");
    let report = judge_run(&fx.home, &[], "not json", &fx.repo, &missing);
    assert_eq!(report["remerge"], true);
    assert_eq!(report["checks"], json!([]));
    let reasons = report["reasons"].as_array().unwrap();
    assert!(
        reasons
            .last()
            .unwrap()
            .as_str()
            .unwrap()
            .ends_with("hold the merge"),
        "{reasons:?}"
    );
}
