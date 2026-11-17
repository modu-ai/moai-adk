//! SPEC-20261117-audit-gate AC-1/AC-2: `audit-gate` decides from a recorded input on stdin (no
//! network) whether a PR head has the audit its tier needs, and `--require-pass` turns anything
//! but `pass` into a failed run, the CI job's exit code.

mod common;

use serde_json::{json, Value};
use std::process::{Command, Output, Stdio};

const H: &str = "1111111111111111111111111111111111111111";
const OLD: &str = "2222222222222222222222222222222222222222";
const B: &str = "3333333333333333333333333333333333333333";

/// A documentation path of the fixture, built at run time so the docs-path scan of
/// `tools/ci_scope.py` does not read it as a repository document.
fn doc() -> String {
    format!("{}/n.md", "docs")
}

fn run(input: &Value, extra: &[&str]) -> Output {
    let home = tempfile::TempDir::new().unwrap();
    let mut child = Command::new(crate::common::MOAI_BIN)
        .arg("audit-gate")
        .args(extra)
        .env("HOME", home.path())
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, input.to_string().as_bytes());
    child.wait_with_output().unwrap()
}

fn decide(input: &Value) -> Value {
    let out = run(input, &[]);
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn verdict(id: u64, head: &str, word: &str, lane: &str) -> Value {
    json!({"id": id, "author_association": "MEMBER",
           "body": format!("AUDIT #9 (x) head {head} base {B} — verdict: {word}\n\n감사: {lane}\n")})
}

fn input(tier: u8, files: &[String], sessions: &[&str], comments: Vec<Value>) -> Value {
    let messages: Vec<String> = sessions
        .iter()
        .map(|s| format!("feat: x\n\nMoai-Spec: SPEC-1\nMoai-Session: {s}\n"))
        .collect();
    json!({"schema_version": 1, "pr": 9, "head": H,
           "body": format!("audit-tier: {tier} reason\n\nbody"),
           "files": files, "commit_messages": messages, "comments": comments})
}

#[test]
fn each_tier_reads_the_verdicts_it_needs() {
    let docs = vec![doc()];
    let src = vec!["src/a.rs".to_string()];
    // Tier 0: no verdict is a pass; a posted FAIL stops it until the same auditor passes.
    assert_eq!(decide(&input(0, &docs, &[], vec![]))["state"], "pass");
    let fail = verdict(5, OLD, "FAIL", "agent-7");
    let r = decide(&input(0, &docs, &[], vec![fail.clone()]));
    assert_eq!(
        (r["state"].clone(), r["rule"].clone()),
        (json!("fail"), json!("open-fail"))
    );
    let cleared = vec![fail.clone(), verdict(6, H, "PASS", "agent-7")];
    assert_eq!(decide(&input(0, &docs, &[], cleared))["state"], "pass");
    let other = vec![fail, verdict(6, H, "PASS", "agent-8")];
    assert_eq!(decide(&input(0, &docs, &[], other))["state"], "fail");
    // Audit B2: a PASS posted before the FAIL (ids 200 then 300, same auditor) clears nothing.
    let earlier = vec![
        verdict(200, H, "PASS", "agent-7"),
        verdict(300, H, "FAIL", "agent-7"),
    ];
    let r = decide(&input(0, &docs, &[], earlier));
    assert_eq!(
        (r["state"].clone(), r["rule"].clone()),
        (json!("fail"), json!("open-fail")),
        "{r}"
    );

    // Tier 1 and 2: a PASS on exactly this head by a lane that wrote no commit.
    for tier in [1, 2] {
        let pass = vec![verdict(5, H, "PASS", "agent-7")];
        let r = decide(&input(tier, &src, &["agent-1"], pass.clone()));
        assert_eq!(r["state"], "pass", "{r}");
        assert_eq!(r["tier"], json!(tier));
        assert_eq!(r["cites_run_checked"], json!(false));
        let r = decide(&input(tier, &src, &["agent-7"], pass.clone()));
        assert_eq!(r["state"], "missing", "{r}");
        let r = decide(&input(tier, &src, &[], pass));
        assert_eq!(r["rule"], "authors", "{r}");
        let r = decide(&input(
            tier,
            &src,
            &["agent-1"],
            vec![verdict(5, OLD, "PASS", "agent-7")],
        ));
        assert_eq!(r["rule"], "pass-on-head", "{r}");
    }
    // A `.github/` path is tier 2 whatever the body says.
    let r = decide(&input(0, &[".github/workflows/x.yml".into()], &[], vec![]));
    assert_eq!(
        (r["state"].clone(), r["tier"].clone()),
        (json!("missing"), json!(2))
    );
}

#[test]
fn a_comment_by_a_non_member_is_not_read() {
    let mut forged = verdict(5, H, "PASS", "agent-7");
    forged["author_association"] = json!("NONE");
    let r = decide(&input(1, &["src/a.rs".into()], &["agent-1"], vec![forged]));
    assert_eq!(r["state"], "missing", "{r}");
    let mut forged_fail = verdict(5, H, "FAIL", "agent-7");
    forged_fail["author_association"] = json!("CONTRIBUTOR");
    assert_eq!(
        decide(&input(0, &[doc()], &[], vec![forged_fail]))["state"],
        "pass"
    );
}

#[test]
fn require_pass_fails_the_run_unless_it_passes() {
    let src = vec!["src/a.rs".to_string()];
    let ok = run(
        &input(
            1,
            &src,
            &["agent-1"],
            vec![verdict(5, H, "PASS", "agent-7")],
        ),
        &["--require-pass"],
    );
    assert_eq!(
        ok.status.code(),
        Some(0),
        "{}",
        String::from_utf8_lossy(&ok.stderr)
    );
    let missing = run(&input(1, &src, &["agent-1"], vec![]), &["--require-pass"]);
    assert_eq!(missing.status.code(), Some(1));
    let v: Value = serde_json::from_slice(&missing.stdout).unwrap();
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(v.to_string().contains("audit gate missing"), "{v}");
    // A malformed input is refused, not passed.
    let bad = run(&json!({"schema_version": 1}), &["--require-pass"]);
    assert_eq!(bad.status.code(), Some(1));
    let r = decide(&json!({"schema_version": 2, "pr": 9, "head": H, "body": "", "comments": []}));
    assert_eq!(r["state"], "fail", "{r}");
}
