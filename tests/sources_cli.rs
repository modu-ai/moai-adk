//! Extra 03-B edge coverage (Issue #1996): the symlink/oversize refusal and the
//! private-identifier search-term gate. The AC-named coverage lives in
//! tests/product_cli.rs because the Issue's verification commands filter there.

use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    process::{Command, Stdio},
};
use tempfile::TempDir;

fn cli(args: &[&str], input: &Value) -> std::process::Output {
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_NOSYSTEM", "1")
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

fn ok_result(out: &std::process::Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn sources_request(candidates: Value) -> Value {
    json!({"protocol_version": 1, "search_terms": ["code review"], "candidates": candidates})
}

#[cfg(unix)]
#[test]
fn skill_install_refuses_symlink_and_oversized_tree() {
    // REQ-03B-04: a symlink or an oversized file inside the source tree refuses the
    // plan — and therefore any install — with a typed detail and no digest.
    let root = TempDir::new().unwrap();
    let src = TempDir::new().unwrap();
    let linked = src.path().join("linked-skill");
    fs::create_dir_all(linked.join("references")).unwrap();
    fs::write(
        linked.join("SKILL.md"),
        b"---\nname: linked-skill\ndescription: fixture\nlicense: MIT\n---\nbody\n",
    )
    .unwrap();
    std::os::unix::fs::symlink("/etc/hosts", linked.join("references/leak")).unwrap();
    let oversized = src.path().join("big-skill");
    fs::create_dir_all(&oversized).unwrap();
    fs::write(
        oversized.join("SKILL.md"),
        b"---\nname: big-skill\ndescription: fixture\nlicense: MIT\n---\nbody\n",
    )
    .unwrap();
    let mut blob = fs::File::create(oversized.join("payload.bin")).unwrap();
    blob.write_all(&[b'x'; 70_000]).unwrap();
    drop(blob);
    let request = sources_request(json!([
        {"source_url": linked.to_str().unwrap(), "source_type": "local",
         "skill_names": ["linked-skill"], "host_targets": ["claude"], "scope": "user"},
        {"source_url": oversized.to_str().unwrap(), "source_type": "local",
         "skill_names": ["big-skill"], "host_targets": ["claude"], "scope": "user"}
    ]));
    let plan = ok_result(&cli(
        &["skill-source-plan", root.path().to_str().unwrap()],
        &request,
    ));
    let results = plan["results"].as_array().unwrap();
    assert_eq!(results[0]["decision"], json!("refused: source_unreadable"));
    assert_eq!(results[0]["tree_digest"], Value::Null);
    assert!(results[0]["detail"].as_str().unwrap().contains("symlink"));
    assert_eq!(results[1]["decision"], json!("refused: source_unreadable"));
    assert!(results[1]["detail"].as_str().unwrap().contains("bounded"));
    // A refused plan installs nothing and leaves no manifest.
    assert!(!root
        .path()
        .join(".moai-cli/manifests/sources.json")
        .exists());
}

#[test]
fn skill_source_rejects_private_identifier_queries() {
    // REQ-03B-08: secret-shaped or path-shaped search terms are refused without echo;
    // no manifest appears and no value reaches either stream.
    for term in [
        "/Users/example/private",
        "token=fixture-secret",
        "sk-fixture000000000000000000000000",
    ] {
        let root = TempDir::new().unwrap();
        let request = json!({"protocol_version": 1, "search_terms": [term], "candidates": []});
        let out = cli(
            &["skill-source-plan", root.path().to_str().unwrap()],
            &request,
        );
        assert!(!out.status.success(), "{term}");
        let failure: Value = serde_json::from_slice(&out.stdout).unwrap();
        assert_eq!(failure["error"]["kind"], json!("invalid_input"));
        assert!(!String::from_utf8_lossy(&out.stdout).contains(term));
        assert!(!String::from_utf8_lossy(&out.stderr).contains(term));
        assert!(!root.path().join(".moai-cli").exists());
    }
}
