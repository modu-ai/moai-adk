//! SPEC-20261120-protected-paths-one-list: the protected check list that `check-guard` reads
//! (`tools/protected_checks.json`) and the paths the merge gate raises to audit tier 2
//! (`gate::TIER2_PREFIXES`) are two lists. This test pins one direction of their agreement:
//! every entry of the protected list must come out as effective tier 2 even when the pull
//! request body declares tier 0, so the two lists cannot drift apart again unnoticed.
//!
//! The test reads the real list of the checkout, calls the library's `effective_tier` and
//! `tier_of` directly, and writes nothing.

use moai_cli::gate::{effective_tier, tier_of, Snapshot};
use serde_json::{json, Value};
use std::path::PathBuf;

/// The files that judge a pull request but sat in neither list before the SPEC.
/// `tools/review_checklist.json` is listed before the file exists: the list is kept in one
/// place, and the pull request that creates the file does not touch the gate.
const JUDGING_FILES: [&str; 5] = [
    "tools/repo_scans.json",
    "tools/self_check.sh",
    "tools/review_checklist.json",
    "skills/moai/workflows/audit.md",
    "AGENTS.md",
];

const HEAD: &str = "0123456789abcdef0123456789abcdef01234567";

fn protected_list_path() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .join("tools")
        .join("protected_checks.json")
}

/// The `protected` entries of the real list, exactly as `tools/check_guard.py` loads them.
fn protected_entries() -> Vec<String> {
    let path = protected_list_path();
    let text = std::fs::read_to_string(&path).unwrap_or_else(|e| panic!("{}: {e}", path.display()));
    let value: Value = serde_json::from_str(&text).expect("the protected list is JSON");
    let entries = value["protected"]
        .as_array()
        .expect("`protected` is an array")
        .iter()
        .map(|e| e.as_str().expect("an entry is a string").to_string())
        .collect::<Vec<_>>();
    assert!(entries.iter().all(|e| !e.is_empty()), "an entry is empty");
    entries
}

/// `*` matches any run of characters, `/` included (fnmatch's rule for a whole-path glob).
fn wildcard(pattern: &str, text: &str) -> bool {
    match pattern.split_once('*') {
        None => pattern == text,
        Some((head, rest)) => {
            let Some(after) = text.strip_prefix(head) else {
                return false;
            };
            (0..=after.len())
                .filter(|i| after.is_char_boundary(*i))
                .any(|i| wildcard(rest, &after[i..]))
        }
    }
}

/// `tools/check_guard.py` `entry_matches`: `/**` is a folder prefix (checked first), an entry
/// with `*` is a whole-path glob, every other entry is an exact name.
fn entry_matches(entry: &str, path: &str) -> bool {
    if let Some(prefix) = entry.strip_suffix("**").filter(|p| p.ends_with('/')) {
        return path.starts_with(prefix);
    }
    if entry.contains('*') {
        return wildcard(entry, path);
    }
    path == entry
}

/// Paths the entry matches, one per way the entry can match. An entry with `?` or `[` is an
/// fnmatch class this test cannot sample, so it is an error and never a skip.
fn samples(entry: &str) -> Result<Vec<String>, String> {
    if entry.contains('?') || entry.contains('[') {
        return Err(format!(
            "entry `{entry}` holds `?` or `[`: the test cannot sample it"
        ));
    }
    let paths: Vec<String> =
        if let Some(prefix) = entry.strip_suffix("**").filter(|p| p.ends_with('/')) {
            vec![format!("{prefix}x"), format!("{prefix}a/b/x")]
        } else if entry.contains('*') {
            ["", "x", "a/b"]
                .iter()
                .map(|filler| entry.replace('*', filler))
                .collect()
        } else {
            vec![entry.to_string()]
        };
    for path in &paths {
        if !entry_matches(entry, path) {
            return Err(format!(
                "sample `{path}` does not match its own entry `{entry}`"
            ));
        }
    }
    Ok(paths)
}

/// A snapshot whose body declares tier 0 and whose only changed file is `path`: only the path
/// rule can raise it.
fn snapshot_of(path: &str) -> Snapshot {
    serde_json::from_value(json!({
        "schema_version": 1,
        "pr": 1,
        "head": HEAD,
        "comments": [],
        "checks": [],
        "pr_files": [path],
        "tip_files_since_verdict_base": [],
        "merge_clean": true,
        "body_check_ok": true,
        "body_tier": 0
    }))
    .expect("a minimal snapshot")
}

/// The tier the gate judges `path` by, under a tier-0 body. `effective_tier` (gate-check) and
/// `tier_of` (audit-gate) must agree: 2 only when both say 2.
fn judged(path: &str) -> u8 {
    let from_snapshot = effective_tier(&snapshot_of(path));
    let from_files = tier_of(Some(0), &[path.to_string()]);
    assert_eq!(
        from_snapshot, from_files,
        "effective_tier and tier_of disagree on `{path}`"
    );
    from_snapshot
}

/// Every (entry, sample path) pair of `entries` that `tier` does not judge as 2, or the first
/// entry the test cannot sample.
fn uncovered(
    entries: &[String],
    tier: impl Fn(&str) -> u8,
) -> Result<Vec<(String, String)>, String> {
    let mut gaps = Vec::new();
    for entry in entries {
        for path in samples(entry)? {
            if tier(&path) != 2 {
                gaps.push((entry.clone(), path));
            }
        }
    }
    Ok(gaps)
}

/// AC-1 (REQ-1, REQ-5): the real list and the gate's path rule agree.
#[test]
fn every_protected_entry_is_judged_tier_2() {
    let entries = protected_entries();
    assert!(!entries.is_empty(), "the protected list is empty");
    let gaps = uncovered(&entries, judged).expect("every real entry can be sampled");
    assert!(
        gaps.is_empty(),
        "check-guard protects these paths but the gate does not judge them tier 2 \
         (add them to gate::TIER2_PREFIXES): {gaps:?}"
    );
}

/// AC-2 (REQ-2): the files that judge a pull request are protected by their exact names.
#[test]
fn the_judging_files_are_in_the_protected_list() {
    let entries = protected_entries();
    for file in JUDGING_FILES {
        assert!(
            entries.iter().any(|e| e == file),
            "`{file}` judges pull requests but is not in tools/protected_checks.json: {entries:?}"
        );
    }
}

/// AC-3 (REQ-3): a gap is named by its entry and its path; a covered list has none.
#[test]
fn a_gap_between_the_two_lists_is_named() {
    let covered = vec![
        "src/gate*.rs".to_string(),
        ".github/workflows/**".to_string(),
    ];
    assert_eq!(uncovered(&covered, judged), Ok(vec![]));

    let exact = vec![
        "src/gate*.rs".to_string(),
        "tools/not_a_tier2_path.txt".to_string(),
    ];
    assert_eq!(
        uncovered(&exact, judged),
        Ok(vec![(
            "tools/not_a_tier2_path.txt".to_string(),
            "tools/not_a_tier2_path.txt".to_string()
        )])
    );

    // A glob names every sample the gate leaves below tier 2, under its own entry.
    let glob = vec!["tools/x*.txt".to_string()];
    let gaps = uncovered(&glob, judged).expect("a glob entry can be sampled");
    assert_eq!(gaps.len(), 3, "{gaps:?}");
    assert!(
        gaps.iter().all(|(entry, _)| entry == "tools/x*.txt"),
        "{gaps:?}"
    );

    // The judge is injectable: a gate that raised nothing leaves every entry a gap.
    let all = uncovered(&covered, |_| 1).expect("sampled");
    assert!(
        all.iter().any(|(e, _)| e == ".github/workflows/**"),
        "{all:?}"
    );
}

/// AC-4 (REQ-4): an entry with an fnmatch class the test cannot sample fails by name.
#[test]
fn an_entry_the_test_cannot_sample_is_reported() {
    for entry in ["src/gate?.rs", "src/[gate].rs"] {
        let err = uncovered(&[entry.to_string()], judged).expect_err(entry);
        assert!(err.contains(entry), "{err}");
    }
}

/// AC-5 (REQ-1): a sample matches its own entry by `check_guard.py`'s rule, and a path the entry
/// does not match (the same name under `docs/`, which is tier-0 safe) stays below tier 2.
#[test]
fn a_sample_path_matches_its_own_entry_and_a_neighbour_is_not_raised() {
    for entry in protected_entries() {
        for path in samples(&entry).expect("sampled") {
            assert!(
                entry_matches(&entry, &path),
                "`{path}` must match `{entry}`"
            );
        }
        let neighbour = format!("docs/{}", entry.replace('*', "x").trim_end_matches("/**"));
        let neighbour = if entry.ends_with("/**") {
            format!("{neighbour}/y")
        } else {
            neighbour
        };
        assert!(
            !entry_matches(&entry, &neighbour),
            "`{neighbour}` must not match `{entry}`"
        );
        assert!(
            judged(&neighbour) < 2,
            "`{neighbour}` is outside `{entry}` and must not be raised to tier 2"
        );
    }
}
