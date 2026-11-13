//! 06-C (#2007) worktree per claim, write_scope conflicts and PR body rendering.
//! Every test builds its own git repository in a temporary directory, runs the real
//! `moai-cli` binary, and never touches the developer's repository, its common dir or $HOME.
//! No `gh` is run: `pr-body-render` only renders text (real GitHub writes: not_run).
//! The file is new per the lane dispatch; the AC commands name `product_cli`, these tests
//! carry the same function names in `worktree_cli`.
use moai_cli::forge::outbox::parse_op_markers;
use moai_cli::pr_link::closing_keywords;
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Output},
};

const ISSUE: i64 = 7;
const KEY_A: &str = "aaaaaaaaaaaaaaaa";
const KEY_B: &str = "bbbbbbbbbbbbbbbb";
const COMMENT_ID: i64 = 6_721_450_649;

fn git(dir: &Path, args: &[&str]) -> String {
    let out = Command::new("git")
        // #2092 (as #2082/#2088): background maintenance off for every fixture git call.
        .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
        .args([
            "-c",
            "user.email=test@example.invalid",
            "-c",
            "user.name=test",
        ])
        .args(args)
        .current_dir(dir)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "git {args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    String::from_utf8(out.stdout).unwrap().trim().to_string()
}

/// A primary repository with one commit; returns its HEAD.
fn make_repo(dir: &Path) -> String {
    fs::create_dir_all(dir).unwrap();
    git(dir, &["init", "-q", "-b", "main"]);
    fs::write(dir.join("README.md"), "x\n").unwrap();
    // As in the real repository (decision 16:5x→17:0x): nested lane worktrees are ignored.
    fs::write(dir.join(".gitignore"), "/.moai-cli/worktrees/\n").unwrap();
    git(dir, &["add", "README.md", ".gitignore"]);
    git(dir, &["commit", "-q", "-m", "init"]);
    git(dir, &["rev-parse", "HEAD"])
}

fn record(key: &str, logical: &str, base: &str, scope: &[&str], state: &str) -> Value {
    json!({
        "schema_version": 1,
        "task_ref": {
            "provider": "github", "host": "github.com", "owner": "acme", "repo": "widgets",
            "repository_id": 101, "issue_id": 7001, "issue_number": ISSUE
        },
        "task_key": key,
        "logical_key": logical,
        "owner": {"kind": "session", "id": "agent-1-0dd806"},
        "epoch": 1,
        "base_revision": base,
        "worktree": {"common_dir": "/x/.git", "git_dir": "/x/.git", "path": "/x", "head": base},
        "write_scope": scope,
        "spec_ref": {"id": "SPEC-RC1-06-C", "revision": "proposed", "digest": "sha256:abc"},
        "comment_id": COMMENT_ID,
        "claim_token": format!("Moai-Claim: {ISSUE}#{COMMENT_ID}"),
        "lease_until": "2026-11-08T15:25:00Z",
        "state": state,
        "effects": [],
        "last_observed_at": "2026-11-08T12:00:00Z"
    })
}

fn comment(epoch: u64) -> Value {
    json!({
        "id": COMMENT_ID,
        "body": format!("> relay\n\nmoai-claim: session=agent-1-0dd806 epoch={epoch} lease_until=2026-11-08T15:25:00Z")
    })
}

/// Writes the record and comment files; returns the `--record ... --claim-comment ...` args.
fn record_args(dir: &Path, rec: &Value, comment: &Value) -> Vec<String> {
    let r = dir.join("record.json");
    let c = dir.join("comment.json");
    fs::write(&r, rec.to_string()).unwrap();
    fs::write(&c, comment.to_string()).unwrap();
    vec![
        "--record".into(),
        r.to_string_lossy().into(),
        "--claim-comment".into(),
        c.to_string_lossy().into(),
    ]
}

/// Seeds the clone's claim journal (06-A layout) with the given records.
fn seed_journal(repo: &Path, records: &[Value]) {
    let dir = repo.join(".git").join(".moai-cli").join("claims");
    fs::create_dir_all(&dir).unwrap();
    let lines: String = records
        .iter()
        .map(|r| format!("{}\n", json!({"event": "claimed", "record": r})))
        .collect();
    fs::write(dir.join("journal.log"), lines).unwrap();
}

fn run(cwd: &Path, args: &[String]) -> Output {
    Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
        .current_dir(cwd)
        .env("HOME", cwd.join("home"))
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("GIT_AUTHOR_NAME", "test")
        .env("GIT_AUTHOR_EMAIL", "test@example.invalid")
        .env("GIT_COMMITTER_NAME", "test")
        .env("GIT_COMMITTER_EMAIL", "test@example.invalid")
        .output()
        .unwrap()
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON: {e}\n{}\n{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn args(words: &[&str], extra: &[String]) -> Vec<String> {
    words
        .iter()
        .map(|s| s.to_string())
        .chain(extra.iter().cloned())
        .collect()
}

struct Fixture {
    _tmp: tempfile::TempDir,
    repo: PathBuf,
    base: String,
    source: Vec<String>,
}

fn fixture() -> Fixture {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("primary");
    let base = make_repo(&repo);
    let rec = record(KEY_A, "06-C", &base, &["src/worktree.rs"], "running");
    let source = record_args(tmp.path(), &rec, &comment(1));
    Fixture {
        _tmp: tmp,
        repo,
        base,
        source,
    }
}

fn plan(f: &Fixture, prefix: &str) -> Value {
    let p = project(f);
    let out = run(
        &f.repo,
        &args(&["worktree-plan", &p, "--branch-prefix", prefix], &f.source),
    );
    let env = envelope(&out);
    assert_eq!(env["ok"], true, "{env}");
    env["result"].clone()
}

fn project(f: &Fixture) -> String {
    f.repo.to_string_lossy().into()
}

fn apply(f: &Fixture, prefix: &str, digest: &str) -> Value {
    let p = project(f);
    let out = run(
        &f.repo,
        &args(
            &[
                "worktree-apply",
                &p,
                "--branch-prefix",
                prefix,
                "--plan-digest",
                digest,
                // The host's free space must not decide these tests (WT-5 floor).
                "--disk-floor-bytes",
                "0",
            ],
            &f.source,
        ),
    );
    envelope(&out)
}

// ---- worktree ----

/// #2336 AC-1: the branch `worktree-plan` names reads back, through the claim guard's
/// parser, as the record's Issue — for logical keys led by digits, by letters, and ending in
/// digits. The guard reads `issue/` lane branches only, so the round trip uses that prefix;
/// another prefix is reported as not a lane branch (the prefix itself stays a HOLD).
#[test]
fn worktree_plan_branch_round_trips_through_the_claim_guard() {
    use moai_cli::claim_verify::{issue_from_branch, lane_branch, LaneBranch};
    for logical in ["06-C", "A9", "rc2-a9", "D9-2", "7-B"] {
        let tmp = tempfile::tempdir().unwrap();
        let repo = tmp.path().join("primary");
        let base = make_repo(&repo);
        let rec = record(KEY_A, logical, &base, &["src/worktree.rs"], "running");
        let source = record_args(tmp.path(), &rec, &comment(1));
        let p: String = repo.to_string_lossy().into();
        for (prefix, want) in [
            ("issue", LaneBranch::Issue(ISSUE)),
            ("lane", LaneBranch::NotLane),
        ] {
            let out = run(
                &repo,
                &args(&["worktree-plan", &p, "--branch-prefix", prefix], &source),
            );
            let env = envelope(&out);
            assert_eq!(env["ok"], true, "{env}");
            let branch = env["result"]["branch"].as_str().unwrap().to_string();
            assert_eq!(branch, format!("{prefix}/{ISSUE}-{logical}-e1"));
            assert_eq!(lane_branch(&branch), want, "{branch}");
        }
        let issue_branch = format!("issue/{ISSUE}-{logical}-e1");
        assert_eq!(issue_from_branch(&issue_branch), Some(ISSUE));
    }
}

#[test]
fn worktree_plan_idempotent() {
    let f = fixture();
    let p = project(&f);
    let a = args(&["worktree-plan", &p, "--branch-prefix", "lane"], &f.source);
    let first = run(&f.repo, &a);
    let second = run(&f.repo, &a);
    assert!(first.status.success());
    assert_eq!(first.stdout, second.stdout, "plan output is byte-identical");
    let r = envelope(&first)["result"].clone();
    assert_eq!(r["branch"], "lane/7-06-C-e1");
    assert_eq!(r["writes"], false);
    assert_eq!(r["base_sha"], f.base);
    assert!(r["path"]
        .as_str()
        .unwrap()
        .ends_with(".moai-cli/worktrees/7-06-C"));
    // Planning created nothing.
    let refs = git(&f.repo, &["for-each-ref", "refs/heads"]);
    assert!(!refs.contains("lane/"), "{refs}");
    assert_eq!(
        git(&f.repo, &["worktree", "list", "--porcelain"])
            .matches("worktree ")
            .count(),
        1
    );
}

#[test]
fn worktree_plan_requires_prefix_without_default() {
    let f = fixture();
    let p = project(&f);
    let out = run(&f.repo, &args(&["worktree-plan", &p], &f.source));
    let env = envelope(&out);
    assert_eq!(env["ok"], false);
    assert_eq!(env["error"]["kind"], "usage");
    let bad = run(
        &f.repo,
        &args(&["worktree-plan", &p, "--branch-prefix", "../x"], &f.source),
    );
    assert_eq!(envelope(&bad)["ok"], false);
}

#[test]
fn worktree_identity_recorded() {
    let f = fixture();
    let plan = plan(&f, "lane");
    let env = apply(&f, "lane", plan["plan_digest"].as_str().unwrap());
    assert_eq!(env["ok"], true, "{env}");
    let r = &env["result"];
    assert_eq!(r["decision"], "created");
    let path = PathBuf::from(r["path"].as_str().unwrap());
    let measured = |flag: &str| git(&path, &["rev-parse", "--path-format=absolute", flag]);
    let id = &r["identity"];
    assert_eq!(id["common_dir"], measured("--git-common-dir"));
    assert_eq!(id["git_dir"], measured("--git-dir"));
    assert_eq!(id["path"], measured("--show-toplevel"));
    assert_eq!(id["head"], git(&path, &["rev-parse", "HEAD"]));
    assert_eq!(id["head"], f.base);
    assert_eq!(r["record_update_request"]["worktree"], *id);
    assert_eq!(r["record_update_request"]["task_key"], KEY_A);
    // Created locked, with the claim as the reason.
    let list = git(&f.repo, &["worktree", "list", "--porcelain"]);
    assert!(
        list.contains(&format!("locked moai claim {KEY_A} e1")),
        "{list}"
    );
}

#[test]
fn worktree_one_writer_per_task() {
    let f = fixture();
    let plan = plan(&f, "lane");
    let digest = plan["plan_digest"].as_str().unwrap();
    assert_eq!(apply(&f, "lane", digest)["result"]["decision"], "created");
    let again = apply(&f, "lane", digest);
    assert_eq!(again["ok"], true);
    assert_eq!(again["result"]["decision"], "conflict");
    assert_eq!(again["result"]["reason"], "branch_in_use");
    assert_eq!(again["result"]["existing"].as_array().unwrap().len(), 1);
    // Another branch name for the same task still meets the task's locked worktree.
    let other = run(
        &f.repo,
        &args(
            &[
                "worktree-plan",
                &project(&f),
                "--branch-prefix",
                "other",
                "--path",
                "elsewhere",
            ],
            &f.source,
        ),
    );
    let other = envelope(&other)["result"].clone();
    let p = project(&f);
    let out = run(
        &f.repo,
        &args(
            &[
                "worktree-apply",
                &p,
                "--branch-prefix",
                "other",
                "--path",
                "elsewhere",
                "--plan-digest",
                other["plan_digest"].as_str().unwrap(),
                "--disk-floor-bytes",
                "0",
            ],
            &f.source,
        ),
    );
    let env = envelope(&out);
    assert_eq!(env["result"]["decision"], "conflict", "{env}");
    assert_eq!(env["result"]["reason"], "task_has_worktree");
    assert!(!f.repo.join("elsewhere").exists());
}

#[test]
fn worktree_refuses_checked_out_branch() {
    let f = fixture();
    let other = f.repo.parent().unwrap().join("someone-else");
    git(
        &f.repo,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "lane/7-06-C-e1",
            other.to_str().unwrap(),
            &f.base,
        ],
    );
    let plan = plan(&f, "lane");
    let env = apply(&f, "lane", plan["plan_digest"].as_str().unwrap());
    assert_eq!(env["ok"], true, "{env}");
    let r = &env["result"];
    assert_eq!(r["decision"], "conflict");
    assert_eq!(r["reason"], "branch_in_use");
    let existing = r["existing"][0]["path"].as_str().unwrap();
    assert!(existing.ends_with("someone-else"), "{existing}");
    assert_eq!(r["record_update_request"], Value::Null);
    assert!(!PathBuf::from(plan["path"].as_str().unwrap()).exists());
    // A stale digest is refused before anything runs.
    let stale = apply(&f, "lane", "sha256:0");
    assert_eq!(stale["ok"], false);
    assert_eq!(stale["error"]["kind"], "stale_plan");
}

#[test]
fn primary_checkout_untouched() {
    let f = fixture();
    fs::write(f.repo.join("dirty.txt"), "user work\n").unwrap();
    fs::write(f.repo.join("README.md"), "edited\n").unwrap();
    let snapshot = |r: &Path| {
        (
            git(r, &["rev-parse", "HEAD"]),
            git(r, &["symbolic-ref", "HEAD"]),
            git(r, &["status", "--porcelain"]),
            fs::read(r.join(".git").join("index")).unwrap(),
            fs::read(r.join("dirty.txt")).unwrap(),
        )
    };
    let before = snapshot(&f.repo);
    let plan = plan(&f, "lane");
    let env = apply(&f, "lane", plan["plan_digest"].as_str().unwrap());
    assert_eq!(env["result"]["decision"], "created", "{env}");
    assert_eq!(before, snapshot(&f.repo));
}

#[test]
fn worktree_source_has_no_primary_mutating_commands() {
    let src =
        fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join("src/worktree.rs")).unwrap();
    for word in ["checkout", "switch", "reset", "stash", "--force"] {
        assert!(!src.contains(word), "src/worktree.rs mentions {word}");
    }
}

// ---- record input ----

#[test]
fn record_file_must_match_claim_comment() {
    let f = fixture();
    let tmp = tempfile::tempdir().unwrap();
    let rec = record(KEY_A, "06-C", &f.base, &["src/a.rs"], "running");
    let p = project(&f);
    let mismatched = record_args(tmp.path(), &rec, &comment(2));
    let env = envelope(&run(
        &f.repo,
        &args(
            &["worktree-plan", &p, "--branch-prefix", "lane"],
            &mismatched,
        ),
    ));
    assert_eq!(env["ok"], false);
    assert_eq!(env["error"]["kind"], "conflict");
    assert!(env["error"]["message"].as_str().unwrap().contains("epoch"));
    let mut other_id = comment(1);
    other_id["id"] = json!(1);
    let env = envelope(&run(
        &f.repo,
        &args(
            &["worktree-plan", &p, "--branch-prefix", "lane"],
            &record_args(tmp.path(), &rec, &other_id),
        ),
    ));
    assert!(env["error"]["message"]
        .as_str()
        .unwrap()
        .contains("comment_id"));
    // --record without the comment is refused.
    let no_comment = vec!["--record".to_string(), mismatched[1].clone()];
    let env = envelope(&run(
        &f.repo,
        &args(
            &["worktree-plan", &p, "--branch-prefix", "lane"],
            &no_comment,
        ),
    ));
    assert_eq!(env["ok"], false);
    assert!(env["error"]["message"]
        .as_str()
        .unwrap()
        .contains("--claim-comment"));
}

// ---- scope ----

fn scope_check(repo: &Path, extra: &[&str]) -> Value {
    let p = repo.to_string_lossy().to_string();
    let mut a = vec!["scope-check".to_string(), p, "--task".into(), KEY_A.into()];
    a.extend(extra.iter().map(|s| s.to_string()));
    let env = envelope(&run(repo, &a));
    assert_eq!(env["ok"], true, "{env}");
    env["result"].clone()
}

#[test]
fn scope_conflict_blocks() {
    let tmp = tempfile::tempdir().unwrap();
    let r = tmp.path().join("primary");
    let base = make_repo(&r);
    seed_journal(
        &r,
        &[
            record(
                KEY_A,
                "06-C",
                &base,
                &["src/lib.rs", "docs/a.md"],
                "running",
            ),
            record(KEY_B, "05-B", &base, &["src/**"], "running"),
        ],
    );
    let res = scope_check(&r, &[]);
    assert_eq!(res["decision"], "blocked");
    let c = &res["conflicts"][0]["scope_conflict"];
    assert_eq!(c["other_task"], KEY_B);
    assert_eq!(c["overlap"][0]["mine"], "src/lib.rs");
    assert_eq!(c["overlap"].as_array().unwrap().len(), 1);
    assert_eq!(res["record_update_request"]["state"], "blocked");
    assert_eq!(res["record_update_request"]["reason"], "scope_conflict");
}

#[test]
fn scope_conflict_allowed_with_order() {
    let tmp = tempfile::tempdir().unwrap();
    let r = tmp.path().join("primary");
    let base = make_repo(&r);
    seed_journal(
        &r,
        &[
            record(KEY_A, "06-C", &base, &["src/lib.rs"], "running"),
            record(KEY_B, "05-B", &base, &["src/lib.rs"], "running"),
        ],
    );
    let res = scope_check(&r, &["--integration-order", KEY_B]);
    assert_eq!(res["decision"], "allowed_with_order");
    assert_eq!(res["conflicts"][0]["allowed_by_integration_order"], true);
    assert_eq!(res["record_update_request"]["integration_order"], KEY_B);
    // Naming an unrelated task does not lift the block.
    assert_eq!(
        scope_check(&r, &["--integration-order", "x"])["decision"],
        "blocked"
    );
}

#[test]
fn shared_interface_paths_default() {
    let tmp = tempfile::tempdir().unwrap();
    let r = tmp.path().join("primary");
    let base = make_repo(&r);
    seed_journal(
        &r,
        &[
            record(KEY_A, "06-C", &base, &["Cargo.lock", "src/a.rs"], "running"),
            record(KEY_B, "02-B", &base, &["Cargo.lock", "src/b.rs"], "running"),
            record(
                "cccccccccccccccc",
                "01-B",
                &base,
                &["src/a.rs"],
                "cancelled",
            ),
        ],
    );
    let res = scope_check(&r, &[]);
    assert_eq!(res["decision"], "blocked");
    assert_eq!(res["conflicts"].as_array().unwrap().len(), 1, "{res}");
    let o = &res["conflicts"][0]["scope_conflict"]["overlap"][0];
    assert_eq!(o["mine"], "Cargo.lock");
    assert_eq!(o["shared_interface"], true);
    let shared: Vec<&str> = res["shared_interface_paths"]
        .as_array()
        .unwrap()
        .iter()
        .map(|v| v.as_str().unwrap())
        .collect();
    assert_eq!(
        shared,
        [
            "Cargo.lock",
            "Cargo.toml",
            ".github/**",
            ".moai-cli/manifests/**",
            ".moai/manifests/**",
            "AGENTS.md"
        ]
    );
}

// ---- PR body ----

fn receipts_file(dir: &Path, base: &str, command: &str) -> String {
    let check = |cmd: &str, result: &str, exit: Value, passed: Value| {
        json!({
            "command_ref": 0, "command": cmd, "argv": cmd.split(' ').collect::<Vec<_>>(),
            "result": result, "exit_code": exit, "timed_out": false, "cancelled": false,
            "spawn_error": null, "started_unix_ms": null, "finished_unix_ms": null,
            "duration_ms": null, "stdout_sha256": null, "stderr_sha256": null,
            "failure_class": null, "test_filters": [], "tests_passed": passed,
            "tests_failed": if passed.is_null() { Value::Null } else { json!(0) },
            "output_complete": null
        })
    };
    let receipt = json!({
        "producer": {"kind": "runner-receipt"},
        "task_id": "06-C", "attempt_id": "a1", "spec_ref": "SPEC-RC1-06-C@proposed",
        "spec_digest": "sha256:abc", "plan_digest": "sha256:def", "base_revision": base,
        "tree_digest": "sha256:000", "policy_digest": null,
        "toolchain": {"rustc": null, "cargo": null},
        "method": {"change_strategy": "tdd", "domain_modeling": "not-required", "legacy_mode": null},
        "phase": null, "required_evidence": [],
        "checks": [
            check(command, "passed", json!(0), json!(12)),
            check("cargo fmt --check", "not_run", Value::Null, Value::Null)
        ],
        "oracle_changes": [], "red_confirmed": null, "result": "passed"
    });
    let path = dir.join("receipts.json");
    fs::write(&path, json!([receipt]).to_string()).unwrap();
    path.to_string_lossy().into()
}

fn render(f: &Fixture, receipts: &str, extra: &[&str]) -> Output {
    let p = project(f);
    let mut a = args(
        &[
            "pr-body-render",
            &p,
            "--receipts",
            receipts,
            "--base-sha",
            &f.base,
            "--base-branch",
            "v2026.11.0",
            "--default-branch",
            "main",
        ],
        &f.source,
    );
    a.extend(extra.iter().map(|s| s.to_string()));
    run(&f.repo, &a)
}

#[test]
fn pr_body_has_required_fields() {
    let f = fixture();
    let tmp = tempfile::tempdir().unwrap();
    let receipts = receipts_file(tmp.path(), &f.base, "cargo test --offline --locked");
    fs::create_dir_all(f.repo.join("src")).unwrap();
    fs::write(f.repo.join("src/new.rs"), "// new\n").unwrap();
    git(&f.repo, &["add", "src/new.rs"]);
    git(&f.repo, &["commit", "-q", "-m", "change"]);
    let env = envelope(&render(
        &f,
        &receipts,
        &[
            "--not-run",
            "live gh (fake only)",
            "--gap",
            "branch prefix is a HOLD",
        ],
    ));
    assert_eq!(env["ok"], true, "{env}");
    let r = &env["result"];
    let body = r["body"].as_str().unwrap();
    let head = git(&f.repo, &["rev-parse", "HEAD"]);
    for needle in [
        "Refs #7",
        "https://github.com/acme/widgets/issues/7",
        "`06-C`",
        "`SPEC-RC1-06-C` · revision `proposed` · digest `sha256:abc`",
        &format!("`v2026.11.0` @ `{}`", f.base),
        &format!("candidate head: `{head}`"),
        "epoch: 1",
        &format!("`Moai-Claim: 7#{COMMENT_ID}`"),
        "| `cargo test --offline --locked` | 0 | passed · 12 passed / 0 failed (receipt a1) |",
        "- not_run: `cargo fmt --check` (receipt a1)",
        "- not_run: live gh (fake only)",
        "## 독립 review",
        "## 변경",
        "| `src/new.rs` | A |",
        "- branch prefix is a HOLD\n\n<!-- moai-op",
    ] {
        assert!(body.contains(needle), "missing {needle:?} in\n{body}");
    }
    let markers = parse_op_markers(body);
    assert_eq!(markers.len(), 1, "{body}");
    assert_eq!(markers[0].operation, "pr_create");
    assert_eq!(markers[0].local_id, KEY_A);
    assert!(body.trim_end().ends_with(&format!(
        "<!-- moai-op {} -->",
        r["operation_id"].as_str().unwrap()
    )));
    assert_eq!(r["record_update_request"]["state"], "pr_ready");
    assert!(r["reconcile"].as_str().unwrap().starts_with("not_run"));
}

#[test]
fn pr_body_no_closing_keyword() {
    let f = fixture();
    let tmp = tempfile::tempdir().unwrap();
    let receipts = receipts_file(tmp.path(), &f.base, "cargo test --offline --locked");
    let env = envelope(&render(&f, &receipts, &["--allow-close"]));
    assert_eq!(env["ok"], true, "{env}");
    let body = env["result"]["body"].as_str().unwrap();
    assert!(closing_keywords(body).is_empty(), "{body}");
    assert!(!body.to_lowercase().contains("closes #"));
    let warning = env["result"]["warnings"][0].as_str().unwrap();
    assert!(warning.contains("default branch (main)"), "{warning}");
    // A body that would carry a keyword is refused, not rendered.
    let bad = receipts_file(tmp.path(), &f.base, "cargo clippy --fix");
    let env = envelope(&render(&f, &bad, &[]));
    assert_eq!(env["ok"], false);
    assert_eq!(env["error"]["kind"], "conflict");
    // A short SHA is refused.
    let p = project(&f);
    let short = run(
        &f.repo,
        &args(
            &[
                "pr-body-render",
                &p,
                "--receipts",
                &receipts,
                "--base-sha",
                &f.base[..12],
                "--base-branch",
                "v2026.11.0",
                "--default-branch",
                "main",
            ],
            &f.source,
        ),
    );
    assert_eq!(envelope(&short)["ok"], false);
}

#[test]
fn pr_body_idempotent() {
    let f = fixture();
    let tmp = tempfile::tempdir().unwrap();
    let receipts = receipts_file(tmp.path(), &f.base, "cargo test --offline --locked");
    let a = render(&f, &receipts, &[]);
    let b = render(&f, &receipts, &[]);
    assert!(a.status.success());
    assert_eq!(a.stdout, b.stdout);
    let empty = tmp.path().join("empty.json");
    fs::write(&empty, "[]").unwrap();
    let env = envelope(&render(&f, empty.to_str().unwrap(), &[]));
    assert!(env["result"]["body"]
        .as_str()
        .unwrap()
        .contains("| (receipt 없음) | not_run |"));
}

// ---- WT-5 (#2380): free-space floor and project-total warning ----

fn apply_with(f: &Fixture, digest: &str, extra: &[&str]) -> Value {
    let p = project(f);
    let mut words = vec![
        "worktree-apply",
        p.as_str(),
        "--branch-prefix",
        "lane",
        "--plan-digest",
        digest,
    ];
    words.extend_from_slice(extra);
    envelope(&run(&f.repo, &args(&words, &f.source)))
}

#[test]
fn below_the_free_space_floor_no_worktree_is_created() {
    let f = fixture();
    let pl = plan(&f, "lane");
    let digest = pl["plan_digest"].as_str().unwrap();
    // No disk has u64::MAX bytes free: the floor refuses deterministically.
    let env = apply_with(&f, digest, &["--disk-floor-bytes", &u64::MAX.to_string()]);
    assert_eq!(env["ok"], true, "{env}");
    let r = &env["result"];
    assert_eq!(r["decision"], "conflict", "{r}");
    assert_eq!(r["reason"], "disk_free_below_floor");
    assert_eq!(r["existing"][0]["floor_bytes"], u64::MAX);
    assert!(r["existing"][0]["free_bytes"].as_u64().unwrap() > 0, "{r}");
    assert!(!Path::new(pl["path"].as_str().unwrap()).exists());
    let branch = pl["branch"].as_str().unwrap();
    assert_eq!(git(&f.repo, &["branch", "--list", branch]), "");
}

#[test]
fn a_created_worktree_reports_capacity_and_the_warning_line() {
    let f = fixture();
    let pl = plan(&f, "lane");
    let env = apply_with(
        &f,
        pl["plan_digest"].as_str().unwrap(),
        &["--disk-floor-bytes", "0", "--warn-total-bytes", "0"],
    );
    let r = &env["result"];
    assert_eq!(r["decision"], "created", "{env}");
    let c = &r["capacity"];
    assert_eq!(c["floor_bytes"], 0);
    assert_eq!(c["warn_bytes"], 0);
    assert!(c["free_bytes"].as_u64().unwrap() > 0, "{c}");
    assert!(c["project_bytes"].as_u64().unwrap() > 0, "{c}");
    assert_eq!(c["warning"], true, "{c}");
    assert_eq!(c["override"], true, "{c}");

    let g = fixture();
    let pl = plan(&g, "lane");
    let env = apply_with(
        &g,
        pl["plan_digest"].as_str().unwrap(),
        &[
            "--disk-floor-bytes",
            "0",
            "--warn-total-bytes",
            &u64::MAX.to_string(),
        ],
    );
    assert_eq!(env["result"]["decision"], "created", "{env}");
    assert_eq!(env["result"]["capacity"]["warning"], false);
}

/// Pre-audit FAIL 6775775969 B2: without the hidden flags the real call path uses the
/// decided limits (30 GiB floor, 40 GiB warning line) and says it is not an override,
/// whether this host's free space lets the worktree be created or not.
#[test]
fn without_flags_the_decided_limits_apply() {
    let f = fixture();
    let pl = plan(&f, "lane");
    let env = apply_with(&f, pl["plan_digest"].as_str().unwrap(), &[]);
    assert_eq!(env["ok"], true, "{env}");
    let r = &env["result"];
    let limits = match r["decision"].as_str() {
        Some("created") => r["capacity"].clone(),
        Some("conflict") => {
            assert_eq!(r["reason"], "disk_free_below_floor", "{r}");
            r["existing"][0].clone()
        }
        other => panic!("unexpected decision {other:?}: {r}"),
    };
    assert_eq!(limits["floor_bytes"], 32_212_254_720u64, "{limits}");
    assert_eq!(limits["warn_bytes"], 42_949_672_960u64, "{limits}");
    assert_eq!(limits["override"], false, "{limits}");
}
