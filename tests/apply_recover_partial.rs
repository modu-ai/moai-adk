//! Issue #2072: `apply-recover` that fails part way must say what it already reverted.
//! The receipts here are written by hand so one backup can be missing on purpose.
mod common;
use serde_json::{json, Value};
use std::{fs, path::Path, process::Command};
use tempfile::TempDir;

fn digest(bytes: &[u8]) -> String {
    moai_cli::digest(bytes)
}

fn recover(root: &Path, id: &str) -> (bool, Value) {
    let out = Command::new(crate::common::MOAI_BIN)
        .args([
            "apply-recover",
            root.to_str().unwrap(),
            "--receipt",
            id,
            "--permission-reference",
            "fixture-approved",
        ])
        .output()
        .unwrap();
    (
        out.status.success(),
        serde_json::from_slice(&out.stdout).unwrap(),
    )
}

/// A receipt `id` that replaced each `(path, before, after)` file; `after` is written as the
/// current content and `before` is stored as the backup unless the path is in `no_backup`.
fn replaced(root: &Path, id: &str, files: &[(&str, &[u8], &[u8])], no_backup: &[&str]) {
    let store = root.join(".moai-cli/receipts");
    let backups = store.join(format!("{id}.before"));
    fs::create_dir_all(&backups).unwrap();
    let mut entries = Vec::new();
    for (path, before, after) in files {
        fs::write(root.join(path), after).unwrap();
        if !no_backup.contains(path) {
            fs::write(backups.join(path), before).unwrap();
        }
        entries.push(json!({"path": path, "decision": "merge",
            "before_sha256": digest(before), "after_sha256": digest(after),
            "published": true, "error": null}));
    }
    let receipt = json!({"schema_version": 1, "request_id": id, "operation": "agents.apply",
        "clone_id": null, "workspace_id": null, "entries": entries, "status": "complete"});
    fs::write(
        store.join(format!("{id}.json")),
        serde_json::to_vec(&receipt).unwrap(),
    )
    .unwrap();
}

#[test]
fn recover_reports_partial_with_reverted_failed_and_skipped() {
    // Two replaced files; the second one's backup is gone. The first is reverted before
    // the second fails, so the failure must name what already changed (RED on aaa28393).
    let root = TempDir::new().unwrap();
    let r = root.path();
    replaced(
        r,
        "manual-two-1",
        &[
            ("first.txt", b"first before\n", b"first after\n"),
            ("second.txt", b"second before\n", b"second after\n"),
        ],
        &["second.txt"],
    );
    let (ok, v) = recover(r, "manual-two-1");
    assert!(!ok);
    assert_eq!(v["error"]["kind"], json!("partial"), "{v}");
    let msg = v["error"]["message"].as_str().unwrap();
    assert!(msg.contains("reverted [first.txt]"), "{msg}");
    assert!(msg.contains("failed at second.txt"), "{msg}");
    // The first file really was reverted; the second keeps its bytes.
    assert_eq!(fs::read(r.join("first.txt")).unwrap(), b"first before\n");
    assert_eq!(fs::read(r.join("second.txt")).unwrap(), b"second after\n");
}

#[test]
fn recover_partial_lists_skipped_and_not_attempted() {
    // Order: user-edited (skipped), reverted, failing, then one never reached.
    let root = TempDir::new().unwrap();
    let r = root.path();
    replaced(
        r,
        "manual-four-1",
        &[
            ("a.txt", b"a0\n", b"a1\n"),
            ("b.txt", b"b0\n", b"b1\n"),
            ("c.txt", b"c0\n", b"c1\n"),
            ("d.txt", b"d0\n", b"d1\n"),
        ],
        &["c.txt"],
    );
    fs::write(r.join("a.txt"), b"user edit\n").unwrap();
    let (ok, v) = recover(r, "manual-four-1");
    assert!(!ok);
    assert_eq!(v["error"]["kind"], json!("partial"), "{v}");
    let msg = v["error"]["message"].as_str().unwrap();
    assert!(msg.contains("reverted [b.txt]"), "{msg}");
    assert!(msg.contains("failed at c.txt"), "{msg}");
    assert!(
        msg.contains("skipped [a.txt: user_changed_since_apply]"),
        "{msg}"
    );
    assert!(msg.contains("not attempted [d.txt]"), "{msg}");
    assert_eq!(fs::read(r.join("a.txt")).unwrap(), b"user edit\n");
    assert_eq!(fs::read(r.join("b.txt")).unwrap(), b"b0\n");
    assert_eq!(fs::read(r.join("c.txt")).unwrap(), b"c1\n");
    assert_eq!(fs::read(r.join("d.txt")).unwrap(), b"d1\n");
}

#[test]
fn recover_failing_before_any_revert_keeps_its_plain_error() {
    // Nothing changed yet: the error stays the bare refusal it was, not `partial`.
    let root = TempDir::new().unwrap();
    let r = root.path();
    replaced(
        r,
        "manual-one-1",
        &[("only.txt", b"o0\n", b"o1\n")],
        &["only.txt"],
    );
    let (ok, v) = recover(r, "manual-one-1");
    assert!(!ok);
    assert_ne!(v["error"]["kind"], json!("partial"), "{v}");
    assert_eq!(fs::read(r.join("only.txt")).unwrap(), b"o1\n");
    // The success path is unchanged: with the backup back, it reverts as before.
    fs::write(
        r.join(".moai-cli/receipts/manual-one-1.before/only.txt"),
        b"o0\n",
    )
    .unwrap();
    let (ok, v) = recover(r, "manual-one-1");
    assert!(ok, "{v}");
    assert_eq!(v["result"]["state"], json!("reverted"));
    assert_eq!(v["result"]["reverted"], json!(["only.txt"]));
    assert_eq!(v["result"]["skipped"], json!([]));
}
