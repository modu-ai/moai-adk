//! SPEC-20261124-desktop-install-agents: the flow rule sits directly under `rules/` of the
//! default Claude folder (`rules/moai-flow.md`), an earlier `rules/moai/flow.md` is left as it
//! is and named by the preview, and a receipt reverts the path its kind names. Every test runs
//! against a temporary HOME only, never the real one.

use moai_cli::home_routing::{apply, plan, revert, Plan};
use std::fs;
use std::path::{Path, PathBuf};

const PERM: &str = "test-permission";
const NEW_FLOW: &str = ".claude/rules/moai-flow.md";
const OLD_FLOW: &str = ".claude/rules/moai/flow.md";

fn package() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
}

fn temp_home() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::tempdir().unwrap();
    let path = fs::canonicalize(dir.path()).unwrap();
    (dir, path)
}

fn packaged_flow() -> String {
    fs::read_to_string(package().join("references/home/flow.md")).unwrap()
}

fn flow_row(p: &Plan) -> &moai_cli::agents::FilePlan {
    p.hosts[0].flow.as_ref().expect("the package has a flow rule")
}

fn write(path: &Path, bytes: &str) {
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, bytes).unwrap();
}

#[test]
fn flow_rule_is_written_directly_under_rules() {
    let (_t, home) = temp_home();
    let p = plan(&home, &package()).unwrap();
    let row = flow_row(&p);
    assert_eq!(row.action, "create");
    assert!(
        Path::new(&row.path).ends_with("rules/moai-flow.md"),
        "the plan names {}",
        row.path
    );
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "applied", "{r}");
    assert_eq!(fs::read_to_string(home.join(NEW_FLOW)).unwrap(), packaged_flow());
    assert!(
        !home.join(".claude/rules/moai").exists(),
        "no rules/moai folder is made"
    );
    // The routing file stays where it was.
    assert!(home.join(".claude/rules/moai.md").is_file());
}

#[test]
fn earlier_flow_file_is_left_in_place() {
    let (_t, home) = temp_home();
    let old = home.join(OLD_FLOW);
    write(&old, "an earlier build wrote this\n");
    let before = fs::metadata(&old).unwrap().modified().unwrap();
    let p = plan(&home, &package()).unwrap();
    let row = flow_row(&p);
    assert_eq!(row.action, "create", "the new file is still planned");
    let detail = row.detail.as_deref().unwrap_or("");
    assert!(
        detail.contains("rules/moai/flow.md") && detail.contains("left in place"),
        "the preview names the earlier file: {detail:?}"
    );
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "applied", "{r}");
    assert_eq!(
        fs::read_to_string(&old).unwrap(),
        "an earlier build wrote this\n"
    );
    assert_eq!(fs::metadata(&old).unwrap().modified().unwrap(), before);
    assert_eq!(fs::read_to_string(home.join(NEW_FLOW)).unwrap(), packaged_flow());
    let again = plan(&home, &package()).unwrap();
    assert_eq!(flow_row(&again).action, "unchanged");
    assert_eq!(again.hosts[0].section.as_ref().unwrap().action, "unchanged");
    let r = apply(&home, &package(), &again.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "unchanged");
}

#[test]
fn revert_follows_the_kind() {
    // A receipt of the new apply reverts the new path.
    let (_t, home) = temp_home();
    let p = plan(&home, &package()).unwrap();
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    let id = r["receipt"].as_str().unwrap().to_string();
    assert!(home.join(NEW_FLOW).is_file());
    let r = revert(&home, &id, PERM).unwrap();
    assert_eq!(r["state"], "reverted", "{r}");
    assert!(!home.join(NEW_FLOW).exists(), "the new file goes");

    // An entry of kind `flow`, as the earlier layout wrote it, reverts `rules/moai/flow.md`
    // and never touches `rules/moai-flow.md`.
    let (_t2, home) = temp_home();
    let bytes = "the flow rule an earlier build wrote\n";
    write(&home.join(OLD_FLOW), bytes);
    write(&home.join(NEW_FLOW), bytes);
    let claude = home.join(".claude");
    let id = "routing-1000-0";
    let dir = moai_cli::home_root::path(&home)
        .join("home-routing/receipts")
        .join(id);
    fs::create_dir_all(&dir).unwrap();
    let receipt = serde_json::json!({
        "schema_version": 1,
        "id": id,
        "plan_digest": "sha256:unused",
        "permission_reference": PERM,
        "status": "committed",
        "entries": [{
            "host": "claude",
            "kind": "flow",
            "config_dir": claude.to_string_lossy(),
            "action": "create",
            "before_sha256": null,
            "after_sha256": moai_cli::digest(bytes.as_bytes()),
            "baseline_before": null
        }]
    });
    fs::write(dir.join("receipt.json"), receipt.to_string()).unwrap();
    let r = revert(&home, id, PERM).unwrap();
    assert_eq!(r["state"], "reverted", "{r}");
    assert!(
        !home.join(OLD_FLOW).exists(),
        "the earlier path is what an entry of kind flow reverts: {r}"
    );
    assert_eq!(
        fs::read_to_string(home.join(NEW_FLOW)).unwrap(),
        bytes,
        "the new path is not touched by an earlier receipt"
    );
}
