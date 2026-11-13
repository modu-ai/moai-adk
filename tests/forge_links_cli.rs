//! #2172 (B5): `forge-links-probe` reads whether the sub-issue and issue-dependency APIs answer
//! for one Issue, and plans the links per feature: API write intents when the API answers,
//! `Parent:` / `Blocked-By:` body lines otherwise (REQ-B5-01). It sends GET only; nothing is
//! written. The CLI is driven against a stub `gh` (no network). Numbers come from
//! `tests/fixtures/github/{sub-issues,dependencies}.json`: #1991 is a child of #1978 and is
//! blocked by #1985 and #1987.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

const SENTINEL: &str = "ghp_moaiTestSentinelNotAToken000";

/// `subs` and `deps` are the HTTP statuses the stub answers for the two probe URLs.
fn stub_gh(dir: &Path, subs: u16, deps: u16) -> PathBuf {
    let answer = |status: u16| match status {
        200 => "HTTP/2.0 200 OK\r\n\r\n[]".to_string(),
        s => format!("HTTP/2.0 {s} Nope\r\n\r\n{{\"message\":\"Not Found\"}}"),
    };
    let exit = |status: u16| if status == 200 { 0 } else { 1 };
    let log = dir.join("gh-calls.log");
    let script = format!(
        "#!/bin/sh\nprintf '%s\\n' \"$*\" >> '{log}'\nfor a in \"$@\"; do last=\"$a\"; done\n\
         case \"$last\" in\n\
           *sub_issues*) printf '%s' '{s}'; exit {se};;\n\
           *dependencies/blocked_by*) printf '%s' '{d}'; exit {de};;\n\
           *) printf '%s' 'HTTP/2.0 500 Unexpected\r\n\r\n{{}}'; exit 1;;\n\
         esac\n",
        log = log.display(),
        s = answer(subs),
        se = exit(subs),
        d = answer(deps),
        de = exit(deps),
    );
    let path = dir.join("gh-stub");
    common::write_executable(&path, &script);
    path
}

fn probe(project: &Path, gh: &Path, input: &Value) -> (std::process::Output, Value) {
    let mut child = Command::new(crate::common::MOAI_BIN)
        .arg("forge-links-probe")
        .arg(project)
        .env("MOAI_FORGE_GH_BIN", gh)
        .env("GH_TOKEN", SENTINEL)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut stdin = child.stdin.take().unwrap();
    match std::io::Write::write_all(&mut stdin, input.to_string().as_bytes()) {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    drop(stdin);
    let out = child.wait_with_output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
    (out, v)
}

fn request() -> Value {
    json!({"schema_version": 1, "owner": "acme", "repo": "widgets", "issue_number": 1991,
           "parent": 1978, "blocked_by": [1985, 1987]})
}

/// Every gh call was a GET.
fn calls(dir: &Path) -> Vec<String> {
    fs::read_to_string(dir.join("gh-calls.log"))
        .unwrap_or_default()
        .lines()
        .map(str::to_string)
        .collect()
}

fn assert_get_only(dir: &Path) {
    let calls = calls(dir);
    assert_eq!(calls.len(), 2, "one probe GET per feature: {calls:?}");
    for c in &calls {
        assert!(c.contains("-X GET"), "only GET is sent: {c}");
        assert!(c.contains("per_page=1"), "probe reads one item: {c}");
    }
}

fn run(subs: u16, deps: u16) -> (tempfile::TempDir, std::process::Output, Value) {
    let tmp = tempfile::tempdir().unwrap();
    let project = tmp.path().join("project");
    fs::create_dir(&project).unwrap();
    let gh = stub_gh(tmp.path(), subs, deps);
    let (out, v) = probe(&project, &gh, &request());
    let stdout = String::from_utf8_lossy(&out.stdout);
    let stderr = String::from_utf8_lossy(&out.stderr);
    assert!(!stdout.contains(SENTINEL) && !stderr.contains(SENTINEL));
    (tmp, out, v)
}

#[test]
fn links_probe_uses_the_api_when_both_features_answer() {
    let (tmp, out, v) = run(200, 200);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    assert_eq!(v["operation"], json!("forge.links_probe"));
    assert_eq!(v["ok"], json!(true));
    let r = &v["result"];
    for feature in ["sub_issues", "dependencies"] {
        assert_eq!(r[feature]["status"], json!("available"), "{feature}");
        assert_eq!(r[feature]["http_status"], json!(200), "{feature}");
        assert_eq!(r[feature]["mode"], json!("api"), "{feature}");
        assert_eq!(r[feature]["body_lines"], json!([]), "{feature}");
    }
    let sub = r["sub_issues"]["api_write_intents"].as_array().unwrap();
    assert_eq!(sub.len(), 1);
    assert_eq!(sub[0]["method"], json!("POST"));
    assert_eq!(
        sub[0]["path"],
        json!("repos/acme/widgets/issues/1978/sub_issues")
    );
    let dep = r["dependencies"]["api_write_intents"].as_array().unwrap();
    assert_eq!(dep.len(), 2);
    assert_eq!(
        dep[0]["path"],
        json!("repos/acme/widgets/issues/1991/dependencies/blocked_by")
    );
    assert_eq!(r["executed"], json!(false));
    assert_get_only(tmp.path());
}

#[test]
fn links_probe_falls_back_to_body_links_and_plans_no_api_write() {
    // REQ-B5-01: a failed probe uses the body fallback and never attempts an API write.
    let (tmp, out, v) = run(404, 410);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let r = &v["result"];
    assert_eq!(r["sub_issues"]["status"], json!("unavailable"));
    assert_eq!(r["sub_issues"]["http_status"], json!(404));
    assert_eq!(r["dependencies"]["status"], json!("unavailable"));
    assert_eq!(r["dependencies"]["http_status"], json!(410));
    for feature in ["sub_issues", "dependencies"] {
        assert_eq!(r[feature]["mode"], json!("body_fallback"), "{feature}");
        assert_eq!(r[feature]["api_write_intents"], json!([]), "{feature}");
    }
    assert_eq!(r["sub_issues"]["body_lines"], json!(["Parent: #1978"]));
    assert_eq!(
        r["dependencies"]["body_lines"],
        json!(["Blocked-By: #1985, #1987"])
    );
    assert_get_only(tmp.path());
}

#[test]
fn links_probe_decides_each_feature_on_its_own() {
    let (tmp, out, v) = run(200, 404);
    assert!(out.status.success());
    let r = &v["result"];
    assert_eq!(r["sub_issues"]["mode"], json!("api"));
    assert_eq!(r["dependencies"]["mode"], json!("body_fallback"));
    assert_eq!(r["dependencies"]["api_write_intents"], json!([]));
    assert_get_only(tmp.path());
}

#[test]
fn links_probe_treats_a_permission_refusal_as_blocked_and_falls_back() {
    let (tmp, out, v) = run(403, 200);
    assert!(out.status.success());
    let r = &v["result"];
    assert_eq!(r["sub_issues"]["status"], json!("blocked"));
    assert_eq!(r["sub_issues"]["mode"], json!("body_fallback"));
    assert_eq!(r["sub_issues"]["api_write_intents"], json!([]));
    assert_eq!(r["dependencies"]["mode"], json!("api"));
    assert_get_only(tmp.path());
}

#[test]
fn links_probe_rejects_bad_input_without_requests() {
    let tmp = tempfile::tempdir().unwrap();
    let gh = stub_gh(tmp.path(), 200, 200);
    for input in [
        json!({"schema_version": 1, "owner": "acme", "repo": "widgets", "issue_number": 0}),
        json!({"schema_version": 1, "owner": "../x", "repo": "widgets", "issue_number": 7}),
        json!({"schema_version": 1, "owner": "acme", "repo": "widgets", "issue_number": 7,
               "blocked_by": [7]}),
        json!({"schema_version": 1, "owner": "acme", "repo": "widgets", "issue_number": 7,
               "extra": true}),
        json!({"schema_version": 1, "owner": "acme", "repo": "widgets", "issue_number": 1,
               "blocked_by": (2..=52).collect::<Vec<i64>>()}),
        json!({"schema_version": 1, "owner": "acme", "repo": "widgets", "issue_number": 7,
               "blocked_by": [5, 5]}),
    ] {
        let (out, v) = probe(tmp.path(), &gh, &input);
        assert!(!out.status.success(), "{input}");
        assert_eq!(v["ok"], json!(false), "{input}");
        assert_eq!(v["operation"], json!("forge.links_probe"), "{input}");
        // Every rejection above is an input error: the envelope kind must say so
        // (REQ audit #2220 follow-up — the repeated-blocker message used to fall
        // through to `internal`).
        assert_eq!(v["error"]["kind"], json!("invalid_input"), "{input}");
    }
    assert!(calls(tmp.path()).is_empty(), "no request on bad input");
}

#[test]
fn links_probe_accepts_exactly_fifty_blockers() {
    let tmp = tempfile::tempdir().unwrap();
    let project = tmp.path().join("project");
    fs::create_dir(&project).unwrap();
    let gh = stub_gh(tmp.path(), 200, 200);
    let input = json!({"schema_version": 1, "owner": "acme", "repo": "widgets",
                       "issue_number": 1, "blocked_by": (2..=51).collect::<Vec<i64>>()});
    let (out, v) = probe(&project, &gh, &input);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    assert_eq!(v["ok"], json!(true));
    assert_get_only(tmp.path());
}
