//! SPEC-20261119-agents-conflict-choices (card C-57, slice 2a): a CLAUDE.md conflict comes with
//! choices the session can run after the user's pick. The pick travels in the request as
//! `picks`; a request without picks plans, digests and applies as before.
//!
//! The conflict is made the way `product_cli.rs` makes it: apply once, replace CLAUDE.md with a
//! human text that has no markers, and plan again with another language so the rendered block
//! differs from the baseline.

mod common;

use common::cli::{create, ok_result, profile};

use serde_json::{json, Value};
use std::fs;
use std::io::Write;
use std::process::{Command, Stdio};
use tempfile::TempDir;

/// Runs `moai ARGS` with `input` (JSON) on stdin. The child's HOME is a temporary folder, so no
/// command of this file can read or write the real one (tests/home_isolation_scan.rs).
fn cli(args: &[&str], input: &Value) -> std::process::Output {
    let home = TempDir::new().unwrap();
    let mut child = Command::new(common::MOAI_BIN)
        .args(args)
        .env("HOME", home.path())
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    match child
        .stdin
        .take()
        .unwrap()
        .write_all(input.to_string().as_bytes())
    {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    child.wait_with_output().unwrap()
}

const HUMAN: &str = "# demo\r\n\r\nMy own rules, written by hand.\r\nNo final newline";

fn request(language: &str, local_file: &str, picks: &[&str]) -> Value {
    let mut r = json!({
        "protocol_version": 1,
        "language": language,
        "workflow_reference": "skills/moai/workflows/",
        "local_file": local_file,
    });
    if !picks.is_empty() {
        r["picks"] = json!(picks);
    }
    r
}

fn plan(root: &TempDir, request: &Value) -> std::process::Output {
    cli(&["agents-plan", root.path().to_str().unwrap()], request)
}

fn plan_ok(root: &TempDir, request: &Value) -> Value {
    let out = plan(root, request);
    assert!(
        out.status.success(),
        "{:?} {:?}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    ok_result(&out)
}

fn apply_with(root: &TempDir, plan: &Value, digest: &str) -> std::process::Output {
    cli(
        &[
            "agents-apply",
            root.path().to_str().unwrap(),
            "--plan-digest",
            digest,
            "--permission-reference",
            "fixture-approved",
        ],
        plan,
    )
}

fn apply(root: &TempDir, plan: &Value) -> std::process::Output {
    apply_with(root, plan, plan["plan_digest"].as_str().unwrap())
}

fn apply_ok(root: &TempDir, plan: &Value) -> Value {
    let out = apply(root, plan);
    assert!(
        out.status.success(),
        "{:?} {:?}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    ok_result(&out)
}

/// A project whose generation was applied once in English with `local_file` `none`.
fn applied_project() -> TempDir {
    let root = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    create(&root, &profile(None));
    let first = plan_ok(&root, &request("en", "none", &[]));
    apply_ok(&root, &first);
    root
}

/// The project above with CLAUDE.md replaced by a human text without markers.
fn conflict_project() -> TempDir {
    let root = applied_project();
    fs::write(root.path().join("CLAUDE.md"), HUMAN).unwrap();
    root
}

fn row<'a>(plan: &'a Value, path: &str) -> Option<&'a Value> {
    plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == path)
}

fn action(plan: &Value, path: &str) -> String {
    row(plan, path).unwrap_or_else(|| panic!("no row for {path} in {plan}"))["action"]
        .as_str()
        .unwrap()
        .to_string()
}

fn agents_md(root: &TempDir) -> String {
    fs::read_to_string(root.path().join("CLAUDE.md")).unwrap()
}

// ---- picks (AC-1) ---------------------------------------------------------------------------

#[test]
fn picks_an_unknown_id_is_refused_as_invalid_input_and_writes_nothing() {
    let root = conflict_project();
    let out = plan(&root, &request("ko", "none", &["no-such-pick"]));
    assert!(!out.status.success());
    let body: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(body["ok"], json!(false), "{body}");
    assert_eq!(body["error"]["kind"], json!("invalid_input"), "{body}");
    assert!(
        body["error"]["message"].as_str().unwrap().contains("picks"),
        "{body}"
    );
    assert_eq!(agents_md(&root), HUMAN);
}

#[test]
fn picks_absent_or_empty_plan_digest_and_apply_as_before() {
    let root = conflict_project();
    let plain = plan_ok(&root, &request("ko", "none", &[]));
    // An explicit empty list is the same request: it is not serialized, so the digest is equal.
    let mut with_empty = request("ko", "none", &[]);
    with_empty["picks"] = json!([]);
    let empty = plan_ok(&root, &with_empty);
    assert_eq!(plain["plan_digest"], empty["plan_digest"]);
    // Without a pick the conflict still stops the apply and nothing is written.
    assert_eq!(action(&plain, "CLAUDE.md"), "conflict");
    let refused = apply(&root, &plain);
    assert!(!refused.status.success());
    assert!(String::from_utf8_lossy(&refused.stderr).contains("conflict"));
    assert_eq!(agents_md(&root), HUMAN);
}

// ---- append-managed (AC-2) ------------------------------------------------------------------

#[test]
fn append_pick_turns_the_missing_block_conflict_into_an_append_without_writing() {
    let root = conflict_project();
    let plain = plan_ok(&root, &request("ko", "none", &[]));
    let picked = plan_ok(&root, &request("ko", "none", &["append-managed"]));
    assert_eq!(action(&plain, "CLAUDE.md"), "conflict");
    assert_eq!(action(&picked, "CLAUDE.md"), "append");
    assert_ne!(plain["plan_digest"], picked["plan_digest"]);
    assert_eq!(agents_md(&root), HUMAN, "a plan writes nothing");
}

#[test]
fn append_apply_keeps_every_earlier_byte_and_adds_one_marker_pair() {
    let root = conflict_project();
    let picked = plan_ok(&root, &request("ko", "none", &["append-managed"]));
    apply_ok(&root, &picked);
    let text = agents_md(&root);
    assert!(text.starts_with(HUMAN), "{text:?}");
    assert_eq!(text.matches("moai:managed:begin").count(), 1, "{text:?}");
    assert_eq!(text.matches("moai:managed:end").count(), 1, "{text:?}");
    assert!(text.len() > HUMAN.len());
    // The same request now plans unchanged: the baseline was recorded for the appended block.
    let after = plan_ok(&root, &request("ko", "none", &["append-managed"]));
    assert_eq!(action(&after, "CLAUDE.md"), "unchanged");
}

#[test]
fn append_apply_with_the_digest_of_the_plain_plan_is_refused_as_stale() {
    let root = conflict_project();
    let plain = plan_ok(&root, &request("ko", "none", &[]));
    let picked = plan_ok(&root, &request("ko", "none", &["append-managed"]));
    let out = apply_with(&root, &picked, plain["plan_digest"].as_str().unwrap());
    assert!(!out.status.success());
    assert!(
        String::from_utf8_lossy(&out.stderr).contains("stale"),
        "{:?}",
        String::from_utf8_lossy(&out.stderr)
    );
    assert_eq!(agents_md(&root), HUMAN);
}

#[test]
fn append_pick_leaves_an_edited_block_and_malformed_markers_as_conflict() {
    // An edited block: the block is present, so it is not the missing-block conflict.
    let root = applied_project();
    let generated = agents_md(&root);
    let edited = generated.replace("Checks:", "CHECKS:");
    assert_ne!(edited, generated);
    fs::write(root.path().join("CLAUDE.md"), &edited).unwrap();
    let plan = plan_ok(&root, &request("ko", "none", &["append-managed"]));
    assert_eq!(action(&plan, "CLAUDE.md"), "conflict");
    assert_eq!(agents_md(&root), edited);

    // Malformed markers: a begin marker with no end.
    let broken = "<!-- moai:managed:begin id=moai-workflow -->\nno end\n";
    fs::write(root.path().join("CLAUDE.md"), broken).unwrap();
    let plan = plan_ok(&root, &request("ko", "none", &["append-managed"]));
    assert_eq!(action(&plan, "CLAUDE.md"), "conflict");
    assert_eq!(agents_md(&root), broken);
}

#[test]
fn append_pick_does_not_recreate_a_deleted_file() {
    // The pick appends to a file that is there. A deleted CLAUDE.md whose baseline no longer
    // matches stays a conflict: creating the file is a different choice than the user made.
    let root = applied_project();
    fs::remove_file(root.path().join("CLAUDE.md")).unwrap();
    let plain = plan_ok(&root, &request("ko", "none", &[]));
    let picked = plan_ok(&root, &request("ko", "none", &["append-managed"]));
    assert_eq!(action(&plain, "CLAUDE.md"), "conflict");
    assert_eq!(action(&picked, "CLAUDE.md"), "conflict");
    assert!(!root.path().join("CLAUDE.md").exists());
}

#[test]
fn append_receipt_recovers_the_bytes_the_file_held_before() {
    let root = conflict_project();
    let picked = plan_ok(&root, &request("ko", "none", &["append-managed"]));
    let applied = apply_ok(&root, &picked);
    let receipt = applied["receipt"].as_str().unwrap().to_string();
    assert_ne!(agents_md(&root), HUMAN);
    let recovered = cli(
        &[
            "apply-recover",
            root.path().to_str().unwrap(),
            "--receipt",
            &receipt,
            "--permission-reference",
            "fixture-approved",
        ],
        &json!({}),
    );
    assert!(
        recovered.status.success(),
        "{:?} {:?}",
        String::from_utf8_lossy(&recovered.stdout),
        String::from_utf8_lossy(&recovered.stderr)
    );
    assert_eq!(agents_md(&root), HUMAN);
}

// ---- skip-conflicts (AC-3) ------------------------------------------------------------------

/// The conflict project asked for a personal file that is ignored and picked it: CLAUDE.local.md
/// and the `.gitignore` block are planned beside the conflicting CLAUDE.md.
fn skip_request(picks: &[&str]) -> Value {
    let mut all = vec!["personal-file"];
    all.extend_from_slice(picks);
    request("ko", "ignored", &all)
}

#[test]
fn skip_pick_marks_only_the_conflicting_file_and_the_rest_still_applies() {
    let root = conflict_project();
    let plain = plan_ok(&root, &skip_request(&[]));
    let picked = plan_ok(&root, &skip_request(&["skip-conflicts"]));
    assert_eq!(action(&plain, "CLAUDE.md"), "conflict");
    assert_eq!(action(&picked, "CLAUDE.md"), "skip");
    for path in ["CLAUDE.local.md", ".gitignore"] {
        assert_eq!(action(&picked, path), action(&plain, path), "{path}");
        assert_ne!(action(&picked, path), "conflict", "{path}");
    }

    apply_ok(&root, &picked);
    assert_eq!(agents_md(&root), HUMAN, "the skipped file keeps every byte");
    assert!(root.path().join("CLAUDE.local.md").exists());
    let ignore = fs::read_to_string(root.path().join(".gitignore")).unwrap();
    assert!(ignore.contains("CLAUDE.local.md"), "{ignore:?}");
}

#[test]
fn skip_records_no_baseline_so_the_plain_plan_conflicts_again() {
    let root = conflict_project();
    let picked = plan_ok(&root, &skip_request(&["skip-conflicts"]));
    apply_ok(&root, &picked);
    let plain = plan_ok(&root, &skip_request(&[]));
    assert_eq!(action(&plain, "CLAUDE.md"), "conflict");
    let again = plan_ok(&root, &skip_request(&["skip-conflicts"]));
    assert_eq!(action(&again, "CLAUDE.md"), "skip");
    // A skipped file adds nothing to the manifest: the manifest row is `unchanged` too.
    assert_eq!(
        action(&again, ".moai-cli/manifests/generation.json"),
        "unchanged"
    );
    for path in ["CLAUDE.local.md", ".gitignore"] {
        assert_eq!(action(&again, path), "unchanged", "{path}");
    }
}

// ---- the earlier claude-import pick is gone from the writer ---------------------------------

/// `CLAUDE.md` is moai's file now, so it shadows nothing: the plan carries no `claude_shadow`,
/// and the earlier pick `claude-import` is refused like any unknown id.
#[test]
fn claude_import_is_not_a_pick_and_no_shadow_is_reported() {
    let root = applied_project();
    fs::write(root.path().join("CLAUDE.md"), "# mine\n").unwrap();
    let plain = plan_ok(&root, &request("en", "none", &[]));
    assert!(plain.get("claude_shadow").is_none(), "{plain}");
    let out = plan(&root, &request("en", "none", &["claude-import"]));
    assert!(!out.status.success());
    let body: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(body["error"]["kind"], json!("invalid_input"), "{body}");
    assert_eq!(
        fs::read_to_string(root.path().join("CLAUDE.md")).unwrap(),
        "# mine\n"
    );
}

/// What `apply-recover` does with a receipt the earlier pick `claude-import` left: a receipt
/// records paths and digests, never a pick id, so its `CLAUDE.md` row is read like any other row.
/// The file goes back when it still holds the bytes the apply wrote, and is skipped, named, when
/// it changed since. The receipt of the removed pick is neither refused by name nor ignored.
#[test]
fn an_old_claude_import_receipt_reverts_like_any_other_row() {
    for edited in [false, true] {
        let root = TempDir::new().unwrap();
        let before = "# mine\n";
        let written = "# mine\n@AGENTS.md\n";
        let current = if edited {
            "# mine\n@AGENTS.md\nmore\n"
        } else {
            written
        };
        fs::write(root.path().join("CLAUDE.md"), current).unwrap();
        let store = moai_cli::state_root::path(root.path(), "receipts");
        let id = "agents-apply-0123456789abcdef";
        fs::create_dir_all(store.join(format!("{id}.before"))).unwrap();
        fs::write(store.join(format!("{id}.before/CLAUDE.md")), before).unwrap();
        let receipt = json!({
            "schema_version": 1, "request_id": id, "operation": "agents.apply",
            "clone_id": null, "workspace_id": null, "status": "complete",
            "entries": [{
                "path": "CLAUDE.md", "decision": "append",
                "before_sha256": moai_cli::digest(before.as_bytes()),
                "after_sha256": moai_cli::digest(written.as_bytes()),
                "published": true, "error": null,
            }],
        });
        fs::write(store.join(format!("{id}.json")), receipt.to_string()).unwrap();

        let out = cli(
            &[
                "apply-recover",
                root.path().to_str().unwrap(),
                "--receipt",
                id,
                "--permission-reference",
                "fixture-approved",
            ],
            &json!({}),
        );
        let body: Value = serde_json::from_slice(&out.stdout).unwrap();
        assert_eq!(body["ok"], json!(true), "{body}");
        let text = fs::read_to_string(root.path().join("CLAUDE.md")).unwrap();
        if edited {
            assert_eq!(text, current, "a changed file is left as it is");
            assert_eq!(body["result"]["state"], "nothing_to_revert", "{body}");
            assert_eq!(
                body["result"]["skipped"],
                json!([{"path": "CLAUDE.md", "reason": "user_changed_since_apply"}]),
                "{body}"
            );
        } else {
            assert_eq!(text, before, "{body}");
            assert_eq!(body["result"]["reverted"], json!(["CLAUDE.md"]), "{body}");
        }
    }
}
