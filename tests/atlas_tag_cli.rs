//! Dual-read tag scanner fixtures (C2 #2178). New-file discipline: existing test
//! files are untouched. AC-1 pins grammar v1 (≥5 valid, ≥5 invalid); AC-2 scans a
//! 15-file `@MX:` baseline fixture and asserts the files are byte-identical after
//! the scan (read-only guarantee, REQ-C2-02) with every node read (REQ-C2-01).

use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");

fn scan(path: &Path) -> Value {
    let out = Command::new(EXE)
        .args(["atlas-tag-scan", path.to_str().unwrap()])
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let envelope: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(envelope["protocol_version"], json!(1));
    assert_eq!(envelope["operation"], json!("atlas_tag.scan"));
    assert_eq!(envelope["ok"], json!(true));
    envelope["result"].clone()
}

fn plant(dir: &Path, rel: &str, body: &str) {
    let p = dir.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, body).unwrap();
}

fn file_bytes(path: &PathBuf) -> Vec<u8> {
    fs::read(path).unwrap()
}

/// AC-1: grammar v1 — at least five valid and five invalid spellings, with the
/// two historical forms (space form from top §9, colon form from Core Brain)
/// both parsing into the same node shape.
#[test]
fn atlas_tag_grammar_accepts_and_rejects() {
    let root = TempDir::new().unwrap();

    // Valid spellings (six ≥ five): space form, colon-args form, kind:sub form,
    // kind with no args, MX namespace, marker mid-line after a comment prefix.
    let valid = [
        "@MOAI:REQ SPEC-X/REQ-007\n",      // space form (top §9)
        "@MOAI:SPEC: SPEC-PAYMENT-001\n",  // colon-args form (Core Brain)
        "@MOAI:SPEC:REQ-PAYMENT-003\n",    // kind:sub, colon then args
        "@MOAI:POLICY\n",                  // kind with no args
        "@MX:NOTE: [AUTO] context line\n", // MX namespace with args
        "// @MX:CEILING: < 10k entries\n", // MX marker after comment prefix
    ];
    for (idx, body) in valid.iter().enumerate() {
        let dir = root.path().join(format!("valid-{idx}"));
        plant(&dir, "tags.md", body);
        let result = scan(&dir);
        assert_eq!(result["invalid_count"], json!(0), "{body}");
        assert_eq!(result["node_count"], json!(1), "{body}");
    }
    // The two historical spellings normalize to the same kind with the same args.
    let dir = root.path().join("unify");
    plant(&dir, "space.md", "@MOAI:REQ SPEC-X/REQ-007\n");
    plant(&dir, "colon.md", "@MOAI:SPEC: SPEC-PAYMENT-001\n");
    let result = scan(&dir);
    let node_of = |file: &str| -> Value {
        result["files"]
            .as_array()
            .unwrap()
            .iter()
            .find(|f| f["path"].as_str().unwrap().ends_with(file))
            .expect("file report")["nodes"][0]
            .clone()
    };
    assert_eq!(node_of("space.md")["kind"], json!("REQ"));
    assert_eq!(node_of("space.md")["args"], json!("SPEC-X/REQ-007"));
    assert_eq!(node_of("colon.md")["kind"], json!("SPEC"));
    assert_eq!(node_of("colon.md")["args"], json!("SPEC-PAYMENT-001"));
    let colon = scan(&dir);
    let _ = colon;

    // Invalid spellings (six ≥ five): each lands in `invalid` with a reason.
    let invalid = [
        "@MOAI:note lowercase kind\n",   // lowercase kind
        "@MOAI:\n",                      // empty kind
        "@MOAI:::args without kind\n",   // empty kind segment
        "@MOAI:REQ-Thing /path\n",       // junk directly after the kind run
        "@MX:note: lowercase mx kind\n", // MX kind lowercase
        "@MOAI:SPEC::\n",                // empty sub segment between colons
    ];
    for (idx, body) in invalid.iter().enumerate() {
        let dir = root.path().join(format!("invalid-{idx}"));
        plant(&dir, "tags.md", body);
        let result = scan(&dir);
        assert_eq!(result["invalid_count"], json!(1), "{body}");
        assert_eq!(result["node_count"], json!(0), "{body}");
        let reason = result["files"][0]["invalid"][0]["reason"]
            .as_str()
            .unwrap()
            .to_string();
        assert!(!reason.is_empty(), "{body}");
    }
    assert_eq!(scan(root.path())["parser_version"], json!(1));
}

/// AC-2 + REQ-C2-01/02: a 15-file `@MX:` baseline fixture is scanned in full and
/// stays byte-identical; a mixed-namespace file reads both namespaces and the
/// duplicate is diagnosed only (nothing rewritten, no auto-migration).
#[test]
fn dual_read_fixture_is_read_only_and_diagnoses_duplicates() {
    let root = TempDir::new().unwrap();
    let baseline: Vec<String> = (1..=15)
        .map(|i| {
            format!(
                "# file {i}\n\n// @MX:NOTE: [AUTO] note {i}\n// @MX:SPEC: SPEC-{i:03}\n// @MX:CEILING: < {i}0 entries\n"
            )
        })
        .collect();
    let mut hashes: Vec<Vec<u8>> = Vec::new();
    for (idx, body) in baseline.iter().enumerate() {
        let dir = root.path().join(format!("baseline/f{idx:02}"));
        plant(&dir, "tags.md", body);
        hashes.push(file_bytes(&dir.join("tags.md")));
    }

    // Mixed namespaces: the same kind+sub spelled in both namespaces is a
    // duplicate — diagnosed, not rewritten.
    let mixed = root.path().join("mixed");
    plant(
        &mixed,
        "tags.md",
        "@MOAI:SPEC: SPEC-PAYMENT-001\n@MX:SPEC: SPEC-PAYMENT-001\n",
    );

    let result = scan(root.path());
    assert_eq!(result["file_count"], json!(16));
    assert_eq!(result["node_count"], json!(47)); // 45 baseline + 2 mixed
    assert_eq!(result["invalid_count"], json!(0));
    assert_eq!(result["duplicates"].as_array().unwrap().len(), 1);
    let dup = &result["duplicates"][0];
    assert_eq!(dup["key"], json!("SPEC|"));
    assert_eq!(
        dup["namespaces"].as_array().unwrap().len(),
        2,
        "both namespaces diagnosed"
    );
    assert_eq!(dup["lines"].as_array().unwrap().len(), 2);
    assert_eq!(result["duplicate_diagnosis_only"], json!(true));
    assert_eq!(result["writes"], json!(false));
    assert_eq!(result["modified"], json!(0));

    // Byte-identical fixture (REQ-C2-02).
    for (idx, _body) in baseline.iter().enumerate() {
        let path = root.path().join(format!("baseline/f{idx:02}/tags.md"));
        assert_eq!(file_bytes(&path), hashes[idx], "file {idx} unchanged");
    }
    assert_eq!(
        fs::read_to_string(mixed.join("tags.md")).unwrap(),
        "@MOAI:SPEC: SPEC-PAYMENT-001\n@MX:SPEC: SPEC-PAYMENT-001\n",
        "mixed file untouched — no auto-migration"
    );
}
