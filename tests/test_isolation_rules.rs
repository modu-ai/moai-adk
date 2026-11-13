//! Issue #2666: tests run as parallel processes (cargo-nextest), so no test may share a
//! fixed temporary path with another process (AC 2), and the VERIFY workflow states how a
//! flaky test is classified and where it is recorded (AC 3).

use std::fs;
use std::path::Path;

const ROOT: &str = env!("CARGO_MANIFEST_DIR");

/// Every `temp_dir().join("<name>")` or `CARGO_TARGET_TMPDIR` + `.join("<name>")` with a fixed
/// name in `text`, as (line number, name).
fn fixed_temp_joins(text: &str) -> Vec<(usize, String)> {
    let mut hits = Vec::new();
    for (i, line) in text.lines().enumerate() {
        for base in ["temp_dir().join(\"", "CARGO_TARGET_TMPDIR\")).join(\""] {
            let mut rest = line;
            while let Some(at) = rest.find(base) {
                let after = &rest[at + base.len()..];
                let name: String = after.chars().take_while(|c| *c != '"').collect();
                hits.push((i + 1, name));
                rest = after;
            }
        }
    }
    hits
}

#[test]
fn no_test_shares_a_fixed_temporary_path() {
    let mut offenders = Vec::new();
    let mut allowed = 0;
    for entry in fs::read_dir(Path::new(ROOT).join("tests")).unwrap() {
        let path = entry.unwrap().path();
        if path.extension().and_then(|e| e.to_str()) != Some("rs")
            || path.file_name().and_then(|n| n.to_str()) == Some("test_isolation_rules.rs")
        {
            continue;
        }
        let text = fs::read_to_string(&path).unwrap();
        for (line, name) in fixed_temp_joins(&text) {
            if name.contains("no-such") {
                allowed += 1;
            } else {
                offenders.push(format!("{}:{line} {name}", path.display()));
            }
        }
    }
    assert!(offenders.is_empty(), "fixed temporary paths: {offenders:?}");
    // The two paths that must stay absent are still found, so the scan itself is live.
    assert!(allowed >= 2, "{allowed}");
}

#[test]
fn the_scan_finds_both_forms_and_every_occurrence() {
    let sample = concat!(
        "let a = std::env::temp_dir().join(\"shared\");\n",
        "let b = Path::new(env!(\"CARGO_TARGET_TMPDIR\")).join(\"x\"); let c = std::env::temp_dir().join(\"y\");\n",
        "let d = tempfile::TempDir::new().unwrap();\n",
    );
    assert_eq!(
        fixed_temp_joins(sample),
        vec![
            (1, "shared".to_string()),
            (2, "y".to_string()),
            (2, "x".to_string())
        ]
    );
}

#[test]
fn verify_workflow_states_the_flaky_rules() {
    let text = fs::read_to_string(Path::new(ROOT).join("skills/moai/workflows/verify.md")).unwrap();
    for rule in [
        "passes when the same head is run again is flaky, not fixed (Issue #2666)",
        "the same `head_sha`, the same job, one failure and one later success",
        "one open Issue per test whose title starts with `[flake]`",
        "as a comment on that Issue instead of opening another",
        "A rerun that passes is never cited as the pass of a change that touches the failing test",
        "a test never shares a fixed path with another process",
        "with `no-such` in the name",
        "such a test runs alone under cargo-nextest",
    ] {
        assert!(text.contains(rule), "verify.md lacks: {rule}");
    }
}
