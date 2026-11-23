//! The audit contract (`skills/moai/references/audit.md`) keeps the scratch-folder rule: each audit
//! writes in a folder of its own (SPEC-20261124-bundle-local-tip, AC-4). It lives in its own file
//! because it reads the checked-in `audit.md`, and `tests/bundle_check_cli.rs` stays a test of
//! temporary projects only.

use std::fs;
use std::path::Path;

/// AC-4: the audit contract keeps one line saying each audit has a scratch folder of its own.
#[test]
fn the_audit_contract_gives_each_audit_its_own_scratch_folder() {
    let path = Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/references/audit.md");
    let text = fs::read_to_string(&path).unwrap();
    let line = text
        .lines()
        .find(|l| l.contains("one folder per audit"))
        .expect("audit.md names one scratch folder per audit");
    assert!(
        line.contains("another audit's folder") && line.contains("never"),
        "the line forbids reading and writing another audit's folder: {line}"
    );
}
