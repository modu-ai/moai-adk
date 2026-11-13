//! Parity check fixtures (D6 #2192). New-file discipline: existing test files are
//! untouched. AC-1 exercises a fixture that plants all four key classes (missing /
//! extra / value_mismatch / identical) plus an optional doc set; AC-2 asserts the
//! report shape (envelope, counts by kind, per-key rows with set digests).

mod common;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;

fn plant(dir: &Path, rel: &str, body: &str) {
    let p = dir.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, body).unwrap();
}

fn run_parity(root: &TempDir, template: &str, generated: &str, doc: Option<&str>) -> Value {
    let mut args = vec![
        "parity-check".to_string(),
        root.path().join(template).to_str().unwrap().to_string(),
        root.path().join(generated).to_str().unwrap().to_string(),
    ];
    if let Some(d) = doc {
        args.push("--doc".to_string());
        args.push(root.path().join(d).to_str().unwrap().to_string());
    }
    let mut child = Command::new(EXE)
        .args(&args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    pipe.write_all(b"").unwrap();
    drop(pipe);
    let out = child.wait_with_output().unwrap();
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let envelope: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(envelope["protocol_version"], json!(1));
    assert_eq!(envelope["operation"], json!("parity.check"));
    assert_eq!(envelope["ok"], json!(true));
    envelope["result"].clone()
}

/// AC-1: a fixture planting every difference kind is detected and classified
/// (missing / extra / value_mismatch / identical / doc_only).
#[test]
fn parity_detects_planted_fixture_differences() {
    let root = TempDir::new().unwrap();
    plant(root.path(), "template/a.md", "a: template-only (missing)\n");
    plant(root.path(), "template/b.md", "b: v1\n");
    plant(root.path(), "template/c.md", "c: same bytes\n");
    plant(root.path(), "generated/b.md", "b: v2 (changed)\n");
    plant(root.path(), "generated/c.md", "c: same bytes\n");
    plant(root.path(), "generated/d.md", "d: generated-only (extra)\n");
    plant(root.path(), "docset/e.md", "e: doc-only\n");

    let result = run_parity(&root, "template", "generated", Some("docset"));
    assert_eq!(
        result["counts"],
        json!({"missing":1,"extra":1,"value_mismatch":1,"identical":1,"doc_only":1,"total":5})
    );
    let row_of = |path: &str| -> Value {
        result["files"]
            .as_array()
            .unwrap()
            .iter()
            .find(|f| f["path"] == json!(path))
            .expect("row for planted key")
            .clone()
    };
    assert_eq!(row_of("a.md")["kind"], json!("missing"));
    assert_eq!(row_of("a.md")["class"], json!("template-only"));
    assert_eq!(row_of("d.md")["kind"], json!("extra"));
    assert_eq!(row_of("d.md")["class"], json!("generated-only"));
    assert_eq!(row_of("b.md")["kind"], json!("value_mismatch"));
    assert_eq!(row_of("b.md")["class"], json!("differing"));
    assert_eq!(row_of("c.md")["kind"], json!("identical"));
    assert_eq!(row_of("e.md")["kind"], json!("doc_only"));
    // rows carry the digests they compared
    assert!(row_of("b.md")["template_sha256"].as_str().is_some());
    assert!(row_of("b.md")["generated_sha256"].as_str().is_some());
    assert_eq!(result["writes"], json!(false));
}

/// AC-2 (report shape): identical trees report all-identical with zero
/// differences, and the optional doc set only adds doc_only rows — the counts
/// block stays keyed by the five kinds.
#[test]
fn parity_report_shape_is_stable() {
    let root = TempDir::new().unwrap();
    for f in ["x.md", "y.md"] {
        let body = format!("{f}: shared\n");
        plant(root.path(), &format!("template/{f}"), &body);
        plant(root.path(), &format!("generated/{f}"), &body);
        plant(root.path(), &format!("docset/{f}"), &body);
    }
    let result = run_parity(&root, "template", "generated", Some("docset"));
    assert_eq!(
        result["counts"],
        json!({"missing":0,"extra":0,"value_mismatch":0,"identical":2,"doc_only":0,"total":2})
    );
    assert!(result["sets"]["template"].as_str().is_some());
    assert_eq!(result["writes"], json!(false));

    // Without the doc set the third digests are absent from the rows.
    let result = run_parity(&root, "template", "generated", None);
    assert_eq!(
        result["counts"],
        json!({"missing":0,"extra":0,"value_mismatch":0,"identical":2,"doc_only":0,"total":2})
    );
    for row in result["files"].as_array().unwrap() {
        assert!(row.get("doc_sha256").is_none() || row["doc_sha256"] == Value::Null);
    }
}
