//! SPEC-20261127-review-term-rename (slice 1b, first change): `src/review_names.rs` lists the
//! verdict words and markers the merge gate accepts, so a change to it changes what the gate
//! trusts. It must sit where the other files that judge a pull request sit: in the tier-2
//! prefixes of the gate and in the protected list `check-guard` reads. (The third list, the floor
//! list of the audit contract, is held equal to the code's prefixes by `next_flow_skill_scan`.)

use moai_cli::gate::{effective_tier, tier_of, Snapshot, TIER2_PREFIXES};
use serde_json::{json, Value};

const HEAD: &str = "0123456789abcdef0123456789abcdef01234567";
const FILE: &str = "src/review_names.rs";

fn snapshot_of(path: &str) -> Snapshot {
    serde_json::from_value(json!({
        "schema_version": 1, "pr": 1, "head": HEAD, "comments": [], "checks": [],
        "pr_files": [path], "tip_files_since_verdict_base": [],
        "merge_clean": true, "body_check_ok": true, "body_tier": 0
    }))
    .expect("a minimal snapshot")
}

#[test]
fn a_change_of_the_accepted_verdict_words_is_tier_2_under_a_tier_0_body() {
    assert_eq!(tier_of(Some(0), &[FILE.to_string()]), 2, "tier_of");
    assert_eq!(effective_tier(&snapshot_of(FILE)), 2, "effective_tier");
    // The same name outside the code folders is not raised: a note about the file is docs.
    let note = format!("docs/{FILE}");
    assert!(tier_of(Some(0), &[note]) < 2);
}

#[test]
fn the_gate_prefixes_and_the_protected_list_both_name_the_file() {
    assert!(
        TIER2_PREFIXES.iter().any(|p| FILE.starts_with(p)),
        "no tier-2 prefix covers {FILE}"
    );
    let path = std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join("tools/protected_checks.json");
    let text = std::fs::read_to_string(&path).expect("the protected list is readable");
    let list: Value = serde_json::from_str(&text).expect("the protected list is JSON");
    let entries: Vec<&str> = list["protected"]
        .as_array()
        .expect("`protected` is an array")
        .iter()
        .filter_map(Value::as_str)
        .collect();
    assert!(
        entries.contains(&FILE),
        "tools/protected_checks.json does not name {FILE}: {entries:?}"
    );
}
