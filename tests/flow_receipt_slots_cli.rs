//! SPEC-20261124-flow-receipt-slots: the flow rule file's ownership digest (`flow_sha256` of the
//! rules state file) is one slot that the current kind `moai_flow` (`rules/moai-flow.md`) and the
//! earlier kind `flow` (`rules/moai/flow.md`) share. Reverting the earlier receipt after the
//! current apply must not take the slot away from the file that still owns it. Every test runs
//! against a temporary HOME only, never the real one.

use moai_cli::home_routing::{apply, plan, revert, FLOW_SOURCE, SOURCE};
use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};

const PERM: &str = "test-permission";
const NEW_FLOW: &str = ".claude/rules/moai-flow.md";
const OLD_FLOW: &str = ".claude/rules/moai/flow.md";
const EARLIER_ID: &str = "routing-1000-0";

const ROUTING_TEXT: &str = "## MoAI\n\nUse the skill ({invoke}).\n";
const FLOW_TEXT: &str =
    "# MoAI evidence and handoff\n\n- Make the smallest change that meets the criteria.\n";

/// The text of one package source, written here so no test reads a checked-in file.
fn source_text(rel: &str) -> String {
    match rel {
        r if r == SOURCE => ROUTING_TEXT.to_string(),
        r if r == FLOW_SOURCE => FLOW_TEXT.to_string(),
        other => panic!("no text for {other}"),
    }
}

fn temp_home() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::tempdir().unwrap();
    let path = fs::canonicalize(dir.path()).unwrap();
    (dir, path)
}

/// A package holding only what a plan reads: `SKILL.md`, the routing source and the flow source.
fn package(flow: &str) -> tempfile::TempDir {
    package_with(&source_text(SOURCE), flow)
}

/// The same package with the routing text chosen too.
fn package_with(routing: &str, flow: &str) -> tempfile::TempDir {
    let pkg = tempfile::tempdir().unwrap();
    fs::write(pkg.path().join("SKILL.md"), "package\n").unwrap();
    for (rel, text) in [(SOURCE, routing.to_string()), (FLOW_SOURCE, flow.to_string())] {
        let path = pkg.path().join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, text).unwrap();
    }
    pkg
}

fn write(path: &Path, bytes: &str) {
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, bytes).unwrap();
}

fn state_path(home: &Path) -> PathBuf {
    let claude = home.join(".claude");
    let digest = moai_cli::digest(claude.to_str().unwrap().as_bytes());
    let id = &digest.strip_prefix("sha256:").unwrap()[..12];
    moai_cli::home_root::path(home).join(format!("home-routing/rules-{id}.json"))
}

fn state_bytes(home: &Path) -> Vec<u8> {
    fs::read(state_path(home)).unwrap()
}

fn state_json(home: &Path) -> Value {
    serde_json::from_slice(&state_bytes(home)).unwrap()
}

fn flow_slot(home: &Path) -> Value {
    state_json(home)["flow_sha256"].clone()
}

/// What an earlier build left: `rules/moai/flow.md` holding `bytes`, its ownership digest in the
/// state file's flow slot, and a committed receipt with one entry of kind `flow`. The state file
/// is written the way the product writes it (pretty JSON and a newline).
fn earlier_layout(home: &Path, bytes: &str) {
    write(&home.join(OLD_FLOW), bytes);
    let digest = moai_cli::digest(bytes.as_bytes());
    let state = serde_json::json!({
        "schema_version": 1, "routing_sha256": null, "flow_sha256": digest,
    });
    let mut out = serde_json::to_vec_pretty(&state).unwrap();
    out.push(b'\n');
    let path = state_path(home);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, out).unwrap();
    let dir = moai_cli::home_root::path(home)
        .join("home-routing/receipts")
        .join(EARLIER_ID);
    fs::create_dir_all(&dir).unwrap();
    let receipt = serde_json::json!({
        "schema_version": 1,
        "id": EARLIER_ID,
        "plan_digest": "sha256:unused",
        "permission_reference": PERM,
        "status": "committed",
        "entries": [{
            "host": "claude",
            "kind": "flow",
            "config_dir": home.join(".claude").to_string_lossy(),
            "action": "create",
            "before_sha256": null,
            "after_sha256": digest,
            "baseline_before": null
        }]
    });
    fs::write(dir.join("receipt.json"), receipt.to_string()).unwrap();
}

/// The current apply of `pkg`; the receipt id.
fn current_apply(home: &Path, pkg: &Path) -> String {
    let p = plan(home, pkg).unwrap();
    let r = apply(home, pkg, &p.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "applied", "{r}");
    r["receipt"].as_str().unwrap().to_string()
}

fn flow_action(home: &Path, pkg: &Path) -> String {
    let p = plan(home, pkg).unwrap();
    p.hosts[0]
        .flow
        .as_ref()
        .expect("a flow rule")
        .action
        .clone()
}

#[test]
fn an_earlier_receipt_revert_keeps_the_new_files_hash() {
    // Equal text in both files (one digest) and a newer text in the new file (two digests).
    for same_text in [true, false] {
        let (_t, home) = temp_home();
        let flow = source_text(FLOW_SOURCE);
        let earlier = if same_text {
            flow.clone()
        } else {
            format!("{flow}the flow rule an earlier build wrote\n")
        };
        earlier_layout(&home, &earlier);
        let pkg = package(&flow);
        current_apply(&home, pkg.path());
        let new_digest = moai_cli::digest(fs::read(home.join(NEW_FLOW)).unwrap().as_slice());
        assert_eq!(flow_slot(&home), Value::String(new_digest.clone()));

        let v = revert(&home, EARLIER_ID, PERM).unwrap();
        assert_eq!(v["state"], "reverted", "{v}");
        assert!(!home.join(OLD_FLOW).exists(), "the earlier file goes: {v}");
        assert!(home.join(NEW_FLOW).is_file(), "the new file stays");
        assert_eq!(
            flow_slot(&home),
            Value::String(new_digest),
            "same_text={same_text}: the slot still holds the new file's digest"
        );

        // A package with newer flow text merges into the file the slot still owns.
        let newer = package(&format!("{flow}one more line\n"));
        assert_eq!(
            flow_action(&home, newer.path()),
            "merge",
            "same_text={same_text}"
        );
    }
}

#[test]
fn an_earlier_receipt_revert_without_the_new_file_clears_the_slot() {
    // The new file is absent: the slot goes back to the entry's baseline (null).
    let (_t, home) = temp_home();
    earlier_layout(&home, "the flow rule an earlier build wrote\n");
    let v = revert(&home, EARLIER_ID, PERM).unwrap();
    assert_eq!(v["state"], "reverted", "{v}");
    assert_eq!(flow_slot(&home), Value::Null);

    // The new file was edited after the apply: its digest is not the slot's, so the slot goes back
    // to the baseline as before.
    let (_t2, home) = temp_home();
    earlier_layout(&home, "the flow rule an earlier build wrote\n");
    let pkg = package(&source_text(FLOW_SOURCE));
    current_apply(&home, pkg.path());
    write(&home.join(NEW_FLOW), "the user wrote this\n");
    let v = revert(&home, EARLIER_ID, PERM).unwrap();
    assert_eq!(v["state"], "reverted", "{v}");
    assert_eq!(flow_slot(&home), Value::Null);
    assert_eq!(
        fs::read_to_string(home.join(NEW_FLOW)).unwrap(),
        "the user wrote this\n",
        "the user's file keeps every byte"
    );
}

#[test]
fn reverts_in_order_restore_the_state_file_byte_for_byte() {
    // The state a first apply and its revert leave, from the product itself.
    let (_r, reference) = temp_home();
    let pkg = package(&source_text(FLOW_SOURCE));
    let id = current_apply(&reference, pkg.path());
    revert(&reference, &id, PERM).unwrap();
    let empty = state_bytes(&reference);
    assert_eq!(
        serde_json::from_slice::<Value>(&empty).unwrap()["flow_sha256"],
        Value::Null
    );

    let (_t, home) = temp_home();
    let flow = source_text(FLOW_SOURCE);
    earlier_layout(&home, &flow);
    let before_current = state_bytes(&home);
    let id = current_apply(&home, pkg.path());
    // Newest first: the current receipt restores the state the earlier layout left, byte for byte.
    revert(&home, &id, PERM).unwrap();
    assert_eq!(state_bytes(&home), before_current);
    // Then the earlier receipt: the state of a home no apply ever touched.
    revert(&home, EARLIER_ID, PERM).unwrap();
    assert_eq!(state_bytes(&home), empty);
}

fn routing_slot(home: &Path) -> Value {
    state_json(home)["routing_sha256"].clone()
}

#[test]
fn a_routing_receipt_revert_restores_its_own_slot_while_the_flow_slot_is_held() {
    // Only the kind `flow` entry defers to the file that holds the flow slot. A receipt that
    // changed the routing text alone still gives `routing_sha256` back, though the current
    // `rules/moai-flow.md` holds the flow slot's digest.
    let (_t, home) = temp_home();
    let flow = source_text(FLOW_SOURCE);
    let pkg = package(&flow);
    current_apply(&home, pkg.path());
    let first_routing = routing_slot(&home);
    let held_flow = flow_slot(&home);
    assert!(first_routing.is_string() && held_flow.is_string());

    let newer = package_with("## MoAI\n\nUse the skill ({invoke}), newer text.\n", &flow);
    let second = current_apply(&home, newer.path());
    assert_ne!(routing_slot(&home), first_routing, "the apply moved the slot");
    assert_eq!(flow_slot(&home), held_flow, "the flow text did not change");

    let v = revert(&home, &second, PERM).unwrap();
    assert_eq!(v["state"], "reverted", "{v}");
    assert_eq!(routing_slot(&home), first_routing);
    assert_eq!(flow_slot(&home), held_flow);
}

#[cfg(unix)]
#[test]
fn a_linked_new_file_with_the_slots_bytes_does_not_hold_the_slot() {
    // `rules/moai-flow.md` is a link to a file holding the slot's exact bytes: it is not the plain
    // file the apply wrote, so the earlier receipt's revert gives the slot back (null).
    let (_t, home) = temp_home();
    let flow = source_text(FLOW_SOURCE);
    earlier_layout(&home, &flow);
    let outside = home.join("elsewhere/flow-copy.md");
    write(&outside, &flow);
    let link = home.join(NEW_FLOW);
    fs::create_dir_all(link.parent().unwrap()).unwrap();
    std::os::unix::fs::symlink(&outside, &link).unwrap();
    assert_eq!(
        flow_slot(&home),
        Value::String(moai_cli::digest(flow.as_bytes()))
    );

    let v = revert(&home, EARLIER_ID, PERM).unwrap();
    assert_eq!(v["state"], "reverted", "{v}");
    assert!(!home.join(OLD_FLOW).exists(), "the earlier file goes: {v}");
    assert_eq!(flow_slot(&home), Value::Null);
    assert!(
        fs::symlink_metadata(&link).unwrap().file_type().is_symlink(),
        "the link stays"
    );
    assert_eq!(fs::read_to_string(&outside).unwrap(), flow);
}

#[test]
fn a_skipped_earlier_entry_leaves_the_slot() {
    let (_t, home) = temp_home();
    let flow = source_text(FLOW_SOURCE);
    earlier_layout(&home, &flow);
    let pkg = package(&flow);
    current_apply(&home, pkg.path());
    let slot = flow_slot(&home);
    write(&home.join(OLD_FLOW), "the user edited the earlier file\n");
    let v = revert(&home, EARLIER_ID, PERM).unwrap();
    assert_eq!(v["skipped"][0]["reason"], "user_changed_since_apply", "{v}");
    assert_eq!(flow_slot(&home), slot);
    assert_eq!(
        fs::read_to_string(home.join(OLD_FLOW)).unwrap(),
        "the user edited the earlier file\n"
    );
}
