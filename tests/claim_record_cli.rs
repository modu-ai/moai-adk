//! 06-C (#2007) remainder: the record writes through 06-B's `Journal::commit`.
//! `worktree-apply`, `scope-check` and `pr-body-render` given `--task` (a record in the local
//! journal) commit the worktree identity, `blocked(reason=scope_conflict)` and `pr_ready` with
//! the 06-B `pr_create` effect intent; given `--record` they write nothing. Every test builds
//! its own repository and journal in a temporary directory; no network, no `gh`.
use moai_cli::claim::JournalEvent;
use moai_cli::forge::outbox::parse_op_markers;
mod common;

use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Output},
};

const KEY_A: &str = "aaaaaaaaaaaaaaaa";
const KEY_B: &str = "bbbbbbbbbbbbbbbb";

fn git(dir: &Path, args: &[&str]) -> String {
    let full = [
        ["-c", "user.name=t", "-c", "user.email=t@example.invalid"].as_slice(),
        args,
    ]
    .concat();
    common::git::git(dir, &full)
}

fn make_repo(dir: &Path) -> String {
    fs::create_dir_all(dir).unwrap();
    git(dir, &["init", "-q", "-b", "main"]);
    fs::write(dir.join(".gitignore"), "/.moai-cli/worktrees/\n").unwrap();
    git(dir, &["add", ".gitignore"]);
    git(dir, &["commit", "-q", "-m", "init"]);
    git(dir, &["rev-parse", "HEAD"])
}

fn record(key: &str, logical: &str, base: &str, scope: &[&str]) -> Value {
    json!({
        "schema_version": 1,
        "task_ref": {"provider": "github", "host": "github.com", "owner": "acme", "repo": "widgets",
                     "repository_id": 101, "issue_id": 7001, "issue_number": 7},
        "task_key": key, "logical_key": logical,
        "owner": {"kind": "session", "id": "agent-1-0dd806"},
        "epoch": 1, "base_revision": base,
        "worktree": {"common_dir": "/x/.git", "git_dir": "/x/.git", "path": "/x", "head": base},
        "write_scope": scope,
        "spec_ref": {"id": "SPEC-RC1-06-C", "revision": "proposed"},
        "comment_id": 6_721_450_649i64, "claim_token": "Moai-Claim: 7#6721450649",
        "lease_until": "2026-11-08T15:25:00Z", "state": "running", "effects": [],
        "last_observed_at": "2026-11-08T12:00:00Z"
    })
}

fn journal_dir(repo: &Path) -> PathBuf {
    repo.join(".git").join(".moai-cli").join("claims")
}

fn seed(repo: &Path, records: &[Value]) {
    let dir = journal_dir(repo);
    fs::create_dir_all(&dir).unwrap();
    let lines: String = records
        .iter()
        .map(|r| format!("{}\n", json!({"event": "claimed", "record": r})))
        .collect();
    fs::write(dir.join("journal.log"), lines).unwrap();
}

fn events(repo: &Path) -> Vec<JournalEvent> {
    fs::read_to_string(journal_dir(repo).join("journal.log"))
        .unwrap()
        .lines()
        .map(|l| serde_json::from_str(l).unwrap())
        .collect()
}

fn latest(repo: &Path, key: &str) -> Value {
    let ev = events(repo);
    json!(
        ev.iter()
            .rev()
            .find(|e| e.record.task_key == key)
            .unwrap()
            .record
    )
}

fn run(cwd: &Path, args: &[&str], env: &[(&str, &str)]) -> Output {
    let mut c = Command::new(crate::common::MOAI_BIN);
    c.args(args)
        .current_dir(cwd)
        .env("HOME", cwd.join("home"))
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null");
    for (k, v) in env {
        c.env(k, v);
    }
    c.output().unwrap()
}

fn ok(out: &Output) -> Value {
    let env: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON: {e}\n{}\n{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(env["ok"], true, "{env}");
    env["result"].clone()
}

fn render(repo: &Path, base: &str, source: &[&str]) -> Value {
    let receipts = repo.parent().unwrap().join("receipts.json");
    fs::write(&receipts, "[]").unwrap();
    let p = repo.to_string_lossy().to_string();
    let r = receipts.to_string_lossy().to_string();
    let mut a = vec![
        "pr-body-render",
        &p,
        "--receipts",
        &r,
        "--base-sha",
        base,
        "--base-branch",
        "v2026.11.0",
        "--default-branch",
        "main",
    ];
    a.extend_from_slice(source);
    ok(&run(repo, &a, &[]))
}

#[test]
fn worktree_apply_records_identity_in_journal() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("primary");
    let base = make_repo(&repo);
    seed(&repo, &[record(KEY_A, "06-C", &base, &["src/a.rs"])]);
    let p = repo.to_string_lossy().to_string();
    let plan = ok(&run(
        &repo,
        &[
            "worktree-plan",
            &p,
            "--task",
            KEY_A,
            "--branch-prefix",
            "lane",
        ],
        &[],
    ));
    let res = ok(&run(
        &repo,
        &[
            "worktree-apply",
            &p,
            "--task",
            KEY_A,
            "--branch-prefix",
            "lane",
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            // The host's free space must not decide this test (WT-5 floor).
            "--disk-floor-bytes",
            "0",
        ],
        &[],
    ));
    assert_eq!(res["decision"], "created");
    assert_eq!(res["record_update"]["applied"], true);
    assert_eq!(res["record_update"]["event"], "worktree_recorded");
    let rec = latest(&repo, KEY_A);
    assert_eq!(rec["worktree"], res["identity"]);
    assert_eq!(
        rec["state"], "running",
        "06-A's state is not changed by recording identity"
    );
    assert_eq!(events(&repo).last().unwrap().event, "worktree_recorded");
}

#[test]
fn scope_check_records_blocked_once() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("primary");
    let base = make_repo(&repo);
    seed(
        &repo,
        &[
            record(KEY_A, "06-C", &base, &["src/lib.rs"]),
            record(KEY_B, "05-B", &base, &["src/**"]),
        ],
    );
    let p = repo.to_string_lossy().to_string();
    let first = ok(&run(&repo, &["scope-check", &p, "--task", KEY_A], &[]));
    assert_eq!(first["decision"], "blocked");
    assert_eq!(first["record_update"]["applied"], true);
    let rec = latest(&repo, KEY_A);
    assert_eq!(rec["state"], "blocked");
    let effect = rec["effects"].as_array().unwrap().last().unwrap().clone();
    assert_eq!(effect["kind"], "scope_conflict");
    assert_eq!(
        effect["conflicts"][0]["scope_conflict"]["other_task"],
        KEY_B
    );
    let count = events(&repo).len();
    let second = ok(&run(&repo, &["scope-check", &p, "--task", KEY_A], &[]));
    assert_eq!(second["record_update"]["applied"], false);
    assert_eq!(events(&repo).len(), count, "a repeated check adds no event");
}

#[test]
fn scope_check_records_integration_order() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("primary");
    let base = make_repo(&repo);
    seed(
        &repo,
        &[
            record(KEY_A, "06-C", &base, &["Cargo.lock"]),
            record(KEY_B, "02-B", &base, &["Cargo.lock"]),
        ],
    );
    let p = repo.to_string_lossy().to_string();
    let res = ok(&run(
        &repo,
        &[
            "scope-check",
            &p,
            "--task",
            KEY_A,
            "--integration-order",
            KEY_B,
        ],
        &[],
    ));
    assert_eq!(res["decision"], "allowed_with_order");
    let rec = latest(&repo, KEY_A);
    assert_eq!(rec["state"], "running");
    assert_eq!(
        rec["effects"].as_array().unwrap().last().unwrap(),
        &json!({"kind": "integration_order", "order": KEY_B})
    );
}

#[test]
fn pr_body_render_records_intent_only() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("primary");
    let base = make_repo(&repo);
    seed(&repo, &[record(KEY_A, "06-C", &base, &["src/a.rs"])]);
    let res = render(&repo, &base, &["--task", KEY_A]);
    assert_eq!(res["effect_intent"]["state"], "intended");
    assert_eq!(res["record_update"]["applied"], true);
    assert_eq!(res["record_update"]["event"], "effect_intend");
    let rec = latest(&repo, KEY_A);
    assert_eq!(
        rec["state"], "running",
        "rendering is not opening; pr_ready waits for the ACK"
    );
    let intents: Vec<&Value> = rec["effects"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|e| e["kind"] == "pr_create")
        .collect();
    assert_eq!(intents.len(), 1);
    assert_eq!(
        intents[0]["operation_id"],
        format!("pr_create-101-{KEY_A}-1")
    );
    // The body marker (`moai-op`, r33) and the recorded 06-B marker name one operation.
    let body = res["body"].as_str().unwrap();
    let in_body = parse_op_markers(body);
    let recorded = parse_op_markers(intents[0]["marker"].as_str().unwrap());
    assert_eq!(in_body, recorded);
    assert!(body.contains(&format!("<!-- moai-op pr_create:101:{KEY_A}:1 -->")));
    // A re-render (a new commit changes the body, not the intent) adds nothing.
    git(&repo, &["commit", "-q", "--allow-empty", "-m", "more"]);
    let count = events(&repo).len();
    let again = render(&repo, &base, &["--task", KEY_A]);
    assert_eq!(again["record_update"]["applied"], false);
    assert_eq!(again["effect_intent"]["state"], "intended");
    assert_eq!(events(&repo).len(), count);
}

fn pr_confirm(repo: &Path, op: &str, response: &str, observed: &str) -> Value {
    let p = repo.to_string_lossy().to_string();
    ok(&run(
        repo,
        &[
            "pr-confirm",
            &p,
            "--task",
            KEY_A,
            "--operation-id",
            op,
            "--response-id",
            response,
            "--observed-state",
            observed,
        ],
        &[],
    ))
}

#[test]
fn pr_ready_is_recorded_at_the_pr_ack() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("primary");
    let base = make_repo(&repo);
    seed(&repo, &[record(KEY_A, "06-C", &base, &["src/a.rs"])]);
    let res = render(&repo, &base, &["--task", KEY_A]);
    let op = res["effect_intent"]["operation_id"]
        .as_str()
        .unwrap()
        .to_string();
    let ack = pr_confirm(&repo, &op, "2098", "opened");
    assert_eq!(ack["effect_state"], "confirmed");
    assert_eq!(ack["effect_confirm"]["response_id"], "2098");
    assert_eq!(ack["record_update"]["applied"], true);
    let rec = latest(&repo, KEY_A);
    assert_eq!(rec["state"], "pr_ready");
    let effect = rec["effects"]
        .as_array()
        .unwrap()
        .iter()
        .find(|e| e["kind"] == "pr_create")
        .unwrap()
        .clone();
    assert_eq!(effect["state"], "confirmed");
    // A second ACK changes nothing and goes no further than pr_ready.
    let count = events(&repo).len();
    let again = pr_confirm(&repo, &op, "2098", "opened");
    assert_eq!(again["record_update"]["applied"], false);
    assert_eq!(events(&repo).len(), count);
    assert_eq!(latest(&repo, KEY_A)["state"], "pr_ready");
}

#[test]
fn pr_ready_waits_when_the_ack_is_uncertain() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("primary");
    let base = make_repo(&repo);
    seed(&repo, &[record(KEY_A, "06-C", &base, &["src/a.rs"])]);
    let res = render(&repo, &base, &["--task", KEY_A]);
    let op = res["effect_intent"]["operation_id"]
        .as_str()
        .unwrap()
        .to_string();
    let lost = pr_confirm(&repo, &op, "", "uncertain");
    assert_eq!(lost["effect_state"], "uncertain");
    assert_eq!(lost["record_update"]["applied"], false);
    assert_eq!(latest(&repo, KEY_A)["state"], "running");
    // Unknown operation ids are refused.
    let p = repo.to_string_lossy().to_string();
    let out = run(
        &repo,
        &[
            "pr-confirm",
            &p,
            "--task",
            KEY_A,
            "--operation-id",
            "pr_create-1-x-1",
            "--observed-state",
            "opened",
        ],
        &[],
    );
    let env: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(env["ok"], false);
}

fn pr_create_effects(repo: &Path) -> Vec<Value> {
    latest(repo, KEY_A)["effects"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|e| e["kind"] == "pr_create")
        .cloned()
        .collect()
}

/// #2329: a second PR under the same claim epoch, rendered with `--pr-seq 2`, gets its own
/// operation and marker, and its ACK is recorded next to the first PR's. Before the fix both
/// PRs shared `pr_create:101:<key>:1` and the second ACK was taken as the first's.
#[test]
fn second_pr_in_the_same_epoch_gets_its_own_operation() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("primary");
    let base = make_repo(&repo);
    seed(&repo, &[record(KEY_A, "06-C", &base, &["src/a.rs"])]);
    let first = render(&repo, &base, &["--task", KEY_A]);
    let op1 = first["effect_intent"]["operation_id"]
        .as_str()
        .unwrap()
        .to_string();
    // The first PR's marker is the one written before the sequence existed.
    assert!(first["body"]
        .as_str()
        .unwrap()
        .contains(&format!("<!-- moai-op pr_create:101:{KEY_A}:1 -->")));
    assert_eq!(
        pr_confirm(&repo, &op1, "2324", "opened")["effect_state"],
        "confirmed"
    );

    git(
        &repo,
        &["commit", "-q", "--allow-empty", "-m", "second change"],
    );
    let second = render(&repo, &base, &["--task", KEY_A, "--pr-seq", "2"]);
    let op2 = second["effect_intent"]["operation_id"]
        .as_str()
        .unwrap()
        .to_string();
    assert_ne!(op1, op2);
    assert_eq!(op2, format!("pr_create-101-{KEY_A}-2-1"));
    assert_eq!(second["effect_intent"]["state"], "intended");
    assert_eq!(second["record_update"]["applied"], true);
    let body2 = second["body"].as_str().unwrap();
    assert!(body2.contains(&format!("<!-- moai-op pr_create:101:{KEY_A}-2:1 -->")));
    let (m1, m2) = (
        parse_op_markers(first["body"].as_str().unwrap()),
        parse_op_markers(body2),
    );
    assert_eq!((m1.len(), m2.len()), (1, 1));
    assert_ne!(m1[0], m2[0], "two PRs, two operation identities");

    let ack2 = pr_confirm(&repo, &op2, "2328", "opened");
    assert_eq!(ack2["effect_state"], "confirmed");
    assert_eq!(ack2["effect_confirm"]["response_id"], "2328");
    let responses: Vec<Value> = pr_create_effects(&repo)
        .iter()
        .map(|e| e["response_id"].clone())
        .collect();
    assert_eq!(responses, vec![json!("2324"), json!("2328")]);
}

/// #2329: an ACK for a different PR on an operation already confirmed is refused as a
/// conflict instead of being taken silently, and the journal is not written.
#[test]
fn a_different_pr_ack_on_a_confirmed_operation_is_a_conflict() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("primary");
    let base = make_repo(&repo);
    seed(&repo, &[record(KEY_A, "06-C", &base, &["src/a.rs"])]);
    let first = render(&repo, &base, &["--task", KEY_A]);
    let op = first["effect_intent"]["operation_id"]
        .as_str()
        .unwrap()
        .to_string();
    pr_confirm(&repo, &op, "2324", "opened");
    let count = events(&repo).len();
    let p = repo.to_string_lossy().to_string();
    let out = run(
        &repo,
        &[
            "pr-confirm",
            &p,
            "--task",
            KEY_A,
            "--operation-id",
            &op,
            "--response-id",
            "2328",
            "--observed-state",
            "opened",
        ],
        &[],
    );
    let env: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(env["ok"], false, "{env}");
    assert!(env.to_string().contains("--pr-seq"), "{env}");
    assert_eq!(events(&repo).len(), count);
    assert_eq!(pr_create_effects(&repo)[0]["response_id"], "2324");
    // The same PR's ACK again is still accepted as is.
    assert_eq!(
        pr_confirm(&repo, &op, "2324", "opened")["effect_state"],
        "confirmed"
    );
}

#[test]
fn reconcile_finds_the_pr_by_its_body_marker() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("primary");
    let base = make_repo(&repo);
    seed(&repo, &[record(KEY_A, "06-C", &base, &["src/a.rs"])]);
    let res = render(&repo, &base, &["--task", KEY_A]);
    let body_path = tmp.path().join("pr-body.md");
    fs::write(&body_path, res["body"].as_str().unwrap()).unwrap();
    let op = res["effect_intent"]["operation_id"]
        .as_str()
        .unwrap()
        .to_string();
    let p = repo.to_string_lossy().to_string();
    // The create was sent and its answer lost.
    ok(&run(
        &repo,
        &[
            "effect-confirm",
            &p,
            "--task",
            KEY_A,
            "--operation-id",
            &op,
            "--observed-state",
            "uncertain",
        ],
        &[],
    ));
    // A PR-body probe: the recorded marker's operation, looked up as the body's marker.
    let probe = tmp.path().join("probe.sh");
    common::write_executable(
        &probe,
        "#!/bin/sh\nid=$(printf '%s' \"$1\" | sed -e 's/^<!-- moai:op //' -e 's/ -->$//')\n\
         if grep -qF \"<!-- moai-op $id -->\" \"$PR_BODY\"; then echo '{\"remote_id\":\"2080\"}'; exit 0; fi\nexit 4\n",
    );
    let probe_s = probe.to_string_lossy().to_string();
    let body_s = body_path.to_string_lossy().to_string();
    let rec = ok(&run(
        &repo,
        &[
            "effects-reconcile",
            &p,
            "--task",
            KEY_A,
            "--remote-probe",
            &probe_s,
        ],
        &[("PR_BODY", &body_s)],
    ));
    assert_eq!(rec["transitions"][0]["to"], "confirmed", "{rec}");
    assert_eq!(rec["transitions"][0]["response_id"], "2080");
    let effect = latest(&repo, KEY_A)["effects"]
        .as_array()
        .unwrap()
        .iter()
        .find(|e| e["kind"] == "pr_create")
        .unwrap()
        .clone();
    assert_eq!(effect["state"], "confirmed");
    // The reconciled ACK is taken by pr-confirm as is, and only then is the record pr_ready.
    assert_eq!(latest(&repo, KEY_A)["state"], "running");
    let ack = pr_confirm(&repo, &op, "2080", "opened");
    assert_eq!(ack["effect_confirm"], Value::Null);
    assert_eq!(ack["record_update"]["applied"], true);
    assert_eq!(latest(&repo, KEY_A)["state"], "pr_ready");
    // Without the PR, the same probe reports the effect as not performed.
    fs::write(&body_path, "no marker\n").unwrap();
    seed(&repo, &[record(KEY_A, "06-C", &base, &["src/a.rs"])]);
    let res = render(&repo, &base, &["--task", KEY_A]);
    let op = res["effect_intent"]["operation_id"]
        .as_str()
        .unwrap()
        .to_string();
    ok(&run(
        &repo,
        &[
            "effect-confirm",
            &p,
            "--task",
            KEY_A,
            "--operation-id",
            &op,
            "--observed-state",
            "uncertain",
        ],
        &[],
    ));
    let rec = ok(&run(
        &repo,
        &[
            "effects-reconcile",
            &p,
            "--task",
            KEY_A,
            "--remote-probe",
            &probe_s,
        ],
        &[("PR_BODY", &body_s)],
    ));
    assert_eq!(rec["transitions"][0]["to"], "not_performed", "{rec}");
}

#[test]
fn record_file_mode_writes_no_journal() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("primary");
    let base = make_repo(&repo);
    let rec_path = tmp.path().join("record.json");
    let comment_path = tmp.path().join("comment.json");
    fs::write(
        &rec_path,
        record(KEY_A, "06-C", &base, &["src/a.rs"]).to_string(),
    )
    .unwrap();
    fs::write(
        &comment_path,
        json!({"id": 6_721_450_649i64,
               "body": "moai-claim: session=agent-1-0dd806 epoch=1 lease_until=2026-11-08T15:25:00Z"})
        .to_string(),
    )
    .unwrap();
    let r = rec_path.to_string_lossy().to_string();
    let c = comment_path.to_string_lossy().to_string();
    let res = render(&repo, &base, &["--record", &r, "--claim-comment", &c]);
    assert_eq!(res["record_update"]["applied"], false);
    assert_eq!(res["effect_intent"], Value::Null);
    assert!(!journal_dir(&repo).join("journal.log").exists());
}
