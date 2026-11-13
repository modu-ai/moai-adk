//! #2173 (B6): `forge-audit-status` posts the `moai/audit` commit status. WHEN an audit-only
//! App token is present (`MOAI_AUDIT_APP_TOKEN`) it is handed to gh as `GH_TOKEN` and the status
//! is marked `source=app`; otherwise the shared account posts it with a warning (REQ-B6-01).
//! The context stays `moai/audit` (the ruleset's required context). No token ever reaches
//! stdout, stderr, argv or disk. Driven against a stub `gh` (no network).
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

const SHARED: &str = "ghp_moaiSharedSentinelNotAToken000";
const APP: &str = "ghs_moaiAppSentinelNotAToken0000000";
const SHA: &str = "0123456789abcdef0123456789abcdef01234567";

/// Logs one line per call: the args and which token gh saw (a label, never the value).
fn stub_gh(dir: &Path) -> PathBuf {
    let log = dir.join("gh-calls.log");
    let script = format!(
        "#!/bin/sh\n\
         if [ \"$GH_TOKEN\" = '{APP}' ]; then t=app; elif [ \"$GH_TOKEN\" = '{SHARED}' ]; then t=default; else t=none; fi\n\
         printf 'token=%s %s\\n' \"$t\" \"$*\" >> '{log}'\n\
         body=$(cat)\n\
         printf '%s\\n' \"$body\" >> '{log}.bodies'\n\
         case \"$*\" in\n\
           */commits/*/statuses*) printf '%s' 'HTTP/2.0 200 OK\r\n\r\n[{{\"context\":\"moai/audit\",\"state\":\"success\",\"description\":\"seen\"}}]';;\n\
           */statuses/*) printf '%s' 'HTTP/2.0 201 Created\r\n\r\n{{\"id\":1,\"context\":\"moai/audit\",\"state\":\"success\"}}';;\n\
           *) printf '%s' 'HTTP/2.0 500 Unexpected\r\n\r\n{{}}'; exit 1;;\n\
         esac\n",
        log = log.display()
    );
    let path = dir.join("gh-stub");
    common::write_executable(&path, &script);
    path
}

fn post(
    dir: &Path,
    app_token: Option<&str>,
    permission: &str,
    input: &Value,
) -> (std::process::Output, Value) {
    let project = dir.join("project");
    fs::create_dir_all(&project).unwrap();
    let mut cmd = Command::new(crate::common::MOAI_BIN);
    cmd.arg("forge-audit-status")
        .arg(&project)
        .arg("--permission-reference")
        .arg(permission)
        .env("MOAI_FORGE_GH_BIN", stub_gh(dir))
        .env("GH_TOKEN", SHARED)
        .env_remove("MOAI_AUDIT_APP_TOKEN")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    if let Some(t) = app_token {
        cmd.env("MOAI_AUDIT_APP_TOKEN", t);
    }
    let mut child = cmd.spawn().unwrap();
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
    json!({"schema_version": 1, "owner": "acme", "repo": "widgets", "sha": SHA,
           "state": "success",
           "target_url": "https://github.com/acme/widgets/pull/9#issuecomment-1",
           "description": "AUDIT PASS"})
}

fn log(dir: &Path) -> String {
    fs::read_to_string(dir.join("gh-calls.log")).unwrap_or_default()
}

/// AC-2: no token value anywhere the program writes.
fn assert_no_secret(dir: &Path, out: &std::process::Output) {
    let mut seen = String::new();
    seen.push_str(&String::from_utf8_lossy(&out.stdout));
    seen.push_str(&String::from_utf8_lossy(&out.stderr));
    seen.push_str(&log(dir));
    seen.push_str(&fs::read_to_string(dir.join("gh-calls.log.bodies")).unwrap_or_default());
    for secret in [SHARED, APP] {
        assert!(!seen.contains(secret), "a token value leaked");
    }
    let project_files: Vec<_> = fs::read_dir(dir.join("project")).unwrap().collect();
    assert!(
        project_files.is_empty(),
        "nothing is written to the project"
    );
}

#[test]
fn audit_status_posts_with_the_app_token_when_present() {
    let tmp = tempfile::tempdir().unwrap();
    let (out, v) = post(tmp.path(), Some(APP), "operator-ok-2173", &request());
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    assert_eq!(v["operation"], json!("forge.audit_status"));
    let r = &v["result"];
    assert_eq!(r["context"], json!("moai/audit"));
    assert_eq!(r["token_source"], json!("app"));
    assert_eq!(r["token_variable"], json!("MOAI_AUDIT_APP_TOKEN"));
    assert_eq!(r["warnings"], json!([]));
    assert_eq!(r["observed"]["state"], json!("success"));
    let calls = log(tmp.path());
    let lines: Vec<&str> = calls.lines().collect();
    assert_eq!(lines.len(), 2, "one POST and one read-back GET: {calls}");
    assert!(lines[0].starts_with("token=app ") && lines[0].contains("-X POST"));
    assert!(lines[0].contains(&format!("repos/acme/widgets/statuses/{SHA}")));
    assert!(lines[1].starts_with("token=app ") && lines[1].contains("-X GET"));
    let bodies = fs::read_to_string(tmp.path().join("gh-calls.log.bodies")).unwrap();
    let sent: Value = serde_json::from_str(bodies.lines().next().unwrap()).unwrap();
    assert_eq!(sent["context"], json!("moai/audit"));
    assert_eq!(sent["state"], json!("success"));
    assert!(sent["description"].as_str().unwrap().contains("source=app"));
    assert_no_secret(tmp.path(), &out);
}

#[test]
fn audit_status_warns_and_uses_the_shared_account_without_an_app_token() {
    let tmp = tempfile::tempdir().unwrap();
    let (out, v) = post(tmp.path(), None, "operator-ok-2173", &request());
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let r = &v["result"];
    assert_eq!(r["context"], json!("moai/audit"));
    assert_eq!(r["token_source"], json!("shared_account"));
    let warnings = r["warnings"].as_array().unwrap();
    assert!(warnings
        .iter()
        .any(|w| w.as_str().unwrap().contains("not pinned")));
    let calls = log(tmp.path());
    assert!(
        calls.lines().all(|l| l.starts_with("token=default ")),
        "{calls}"
    );
    let bodies = fs::read_to_string(tmp.path().join("gh-calls.log.bodies")).unwrap();
    let sent: Value = serde_json::from_str(bodies.lines().next().unwrap()).unwrap();
    assert!(sent["description"]
        .as_str()
        .unwrap()
        .contains("source=shared-account"));
    assert_no_secret(tmp.path(), &out);
}

#[test]
fn audit_status_refuses_bad_input_without_requests() {
    let base = request();
    let with = |k: &str, v: Value| {
        let mut r = base.clone();
        r[k] = v;
        r
    };
    let cases = [
        ("placeholder", base.clone()),
        ("ok", with("sha", json!("abc"))),
        ("ok", with("state", json!("passed"))),
        ("ok", with("target_url", json!("https://example.com/x"))),
        ("ok", with("description", json!("x".repeat(141)))),
        ("ok", with("extra", json!(true))),
    ];
    for (permission, input) in cases {
        let tmp = tempfile::tempdir().unwrap();
        let (out, v) = post(tmp.path(), Some(APP), permission, &input);
        assert!(!out.status.success(), "{input}");
        assert_eq!(v["ok"], json!(false), "{input}");
        assert_eq!(v["operation"], json!("forge.audit_status"), "{input}");
        assert!(
            log(tmp.path()).is_empty(),
            "no request on bad input: {input}"
        );
        assert_no_secret(tmp.path(), &out);
    }
}

/// Each enforced limit is pinned at its edge, so moving the bound either way fails a test:
/// description 140 characters including the source marker (" · source=app" is 13, " ·
/// source=shared-account" is 24), target_url 512 bytes.
#[test]
fn audit_status_limits_hold_exactly_at_their_edges() {
    let base = request();
    let with = |k: &str, v: Value| {
        let mut r = base.clone();
        r[k] = v;
        r
    };
    let url = |len: usize| {
        let head = "https://github.com/";
        json!(format!("{head}{}", "a".repeat(len - head.len())))
    };
    // (app token?, input, accepted?)
    let cases = [
        (true, with("description", json!("d".repeat(127))), true),
        (true, with("description", json!("d".repeat(128))), false),
        (false, with("description", json!("d".repeat(116))), true),
        (false, with("description", json!("d".repeat(117))), false),
        (true, with("target_url", url(512)), true),
        (true, with("target_url", url(513)), false),
    ];
    for (app, input, accepted) in cases {
        let tmp = tempfile::tempdir().unwrap();
        let token = if app { Some(APP) } else { None };
        let (out, v) = post(tmp.path(), token, "operator-ok-2173", &input);
        assert_eq!(v["ok"], json!(accepted), "{input}");
        assert_eq!(out.status.success(), accepted, "{input}");
        if accepted {
            let sent = fs::read_to_string(tmp.path().join("gh-calls.log.bodies")).unwrap();
            let sent: Value = serde_json::from_str(sent.lines().next().unwrap()).unwrap();
            if input["description"].as_str().unwrap().starts_with('d') {
                assert_eq!(sent["description"].as_str().unwrap().chars().count(), 140);
            }
        } else {
            assert!(
                log(tmp.path()).is_empty(),
                "no request past a limit: {input}"
            );
        }
        assert_no_secret(tmp.path(), &out);
    }
}
