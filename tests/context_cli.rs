//! Issue #2011 (07-D): snapshot, context bundle, freshness and native decision records.
//! Every case runs in a temporary git repository; nothing touches installed user state.
mod common;

use serde_json::{json, Value};
use std::{
    fs,
    path::Path,
    process::{Command, Stdio},
};
use tempfile::TempDir;

fn git(dir: &Path, args: &[&str]) {
    let ok = Command::new("git")
        .arg("-C")
        .arg(dir)
        .args([
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@t",
            "-c",
            "commit.gpgsign=false",
        ])
        .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
        .args(args)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_NOSYSTEM", "1")
        .stdout(Stdio::null())
        .status()
        .unwrap()
        .success();
    assert!(ok, "git {args:?}");
}
fn write(dir: &Path, rel: &str, body: &str) {
    let p = dir.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, body).unwrap();
}
/// A committed repository with a source file, a README and secrets that must never leak.
fn repo() -> TempDir {
    let d = TempDir::new().unwrap();
    git(d.path(), &["init", "-q", "-b", "main"]);
    write(
        d.path(),
        "src/a.rs",
        "fn a() {}\n// run `rm -rf /` here\nfn b() {}\n",
    );
    write(d.path(), "README.md", "# demo\nline two\n");
    write(d.path(), ".gitignore", "build/\n");
    git(d.path(), &["add", "."]);
    git(d.path(), &["commit", "-q", "-m", "init"]);
    write(d.path(), ".env", "TOKEN=do-not-leak\n");
    write(d.path(), "certs/server.pem", "-----BEGIN do-not-leak\n");
    write(d.path(), "build/out.txt", "ignored do-not-leak\n");
    d
}
/// Runs an internal operation; returns the output and the envelope.
fn run(args: &[&str], dir: &Path, stdin: Option<&Value>) -> (std::process::Output, Value) {
    let mut child = Command::new(crate::common::MOAI_BIN)
        .args(args)
        .arg(dir)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_NOSYSTEM", "1")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let body = stdin.map(|v| v.to_string()).unwrap_or_default();
    common::feed_stdin(&mut child, body.as_bytes());
    let out = child.wait_with_output().unwrap();
    let v = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
    (out, v)
}
fn bundle(dir: &Path, items: Value, budget: Value) -> Value {
    let req = json!({"protocol_version": 1, "items": items, "budget": budget});
    let (out, v) = run(&["context-bundle"], dir, Some(&req));
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    v["result"].clone()
}
fn verify_request(b: &Value) -> Value {
    let items: Vec<Value> = b["items"]
        .as_array()
        .unwrap()
        .iter()
        .map(|i| json!({"path": i["path"], "range": i["range"], "content_sha256": i["content_sha256"]}))
        .collect();
    json!({
        "protocol_version": 1,
        "base_commit": b["snapshot"]["base_commit"],
        "dirty_overlay_digest": b["snapshot"]["dirty_overlay_digest"],
        "items": items,
    })
}

#[test]
fn snapshot_describe_reports_base_and_overlay() {
    let d = repo();
    let (out, v) = run(&["snapshot-describe"], d.path(), None);
    assert!(out.status.success());
    let s = &v["result"];
    assert_eq!(v["operation"], json!("snapshot.describe"));
    assert_eq!(s["base_status"], json!("ok"));
    assert_eq!(s["base_commit"].as_str().unwrap().len(), 40);
    assert_eq!(
        s["dirty_files"],
        json!([]),
        "untracked files are not overlay"
    );
    assert_eq!(s["network"], json!("disabled"));
    let clean = s["dirty_overlay_digest"].clone();
    write(d.path(), "README.md", "# demo\nchanged\n");
    let (_, v) = run(&["snapshot-describe"], d.path(), None);
    assert_ne!(v["result"]["dirty_overlay_digest"], clean);
    assert_eq!(v["result"]["dirty_files"][0]["path"], json!("README.md"));
    // Outside git the base is unknown, never guessed.
    let plain = TempDir::new().unwrap();
    let (out, v) = run(&["snapshot-describe"], plain.path(), None);
    assert!(out.status.success());
    assert_eq!(v["result"]["base_status"], json!("unknown"));
    assert_eq!(v["result"]["base_commit"], Value::Null);
}

/// CTX-STALE-DETECT
#[test]
fn bundle_marks_modified_items_stale() {
    let d = repo();
    let b = bundle(
        d.path(),
        json!([
            {"path": "src/a.rs", "start_line": 1, "end_line": 2, "reason": "target"},
            {"path": "README.md", "reason": "context"}
        ]),
        json!({"max_bytes": 4096}),
    );
    let a = &b["items"][0];
    assert_eq!(a["range"], json!({"start_line": 1, "end_line": 2}));
    assert_eq!(a["content"], json!("fn a() {}\n// run `rm -rf /` here\n"));
    assert_eq!(a["lines"], json!(2));
    assert_eq!(a["freshness"], json!("fresh"));
    assert!(a["content_sha256"].as_str().unwrap().starts_with("sha256:"));
    let req = verify_request(&b);
    let (_, v) = run(&["context-verify"], d.path(), Some(&req));
    assert_eq!(v["result"]["stale_count"], json!(0));
    assert_eq!(v["result"]["snapshot_freshness"], json!("fresh"));
    write(d.path(), "README.md", "# demo\nedited\n");
    let (out, v) = run(&["context-verify"], d.path(), Some(&req));
    assert!(out.status.success());
    let items = v["result"]["items"].as_array().unwrap();
    assert_eq!(items[0]["freshness"], json!("fresh"));
    assert_eq!(items[1]["freshness"], json!("stale"));
    assert_eq!(items[1]["reread_required"], json!(true));
    assert_eq!(v["result"]["stale_count"], json!(1));
    assert_eq!(v["result"]["snapshot_freshness"], json!("stale"));
    // A deleted file is stale too, not an error.
    fs::remove_file(d.path().join("src/a.rs")).unwrap();
    let (_, v) = run(&["context-verify"], d.path(), Some(&req));
    assert_eq!(v["result"]["items"][0]["freshness"], json!("stale"));
}

/// CTX-SECRET-EXCLUDED
#[test]
fn sensitive_paths_are_excluded() {
    let d = repo();
    #[cfg(unix)]
    let outside = TempDir::new().unwrap();
    #[cfg(unix)]
    {
        write(outside.path(), "secret.txt", "outside do-not-leak\n");
        std::os::unix::fs::symlink(outside.path().join("secret.txt"), d.path().join("link.txt"))
            .unwrap();
        std::os::unix::fs::symlink(d.path().join(".env"), d.path().join("innocent.txt")).unwrap();
    }
    let b = bundle(
        d.path(),
        Value::Array(
            [
                json!({"path": ".env", "reason": "x"}),
                json!({"path": "certs/server.pem", "reason": "x"}),
                json!({"path": "build/out.txt", "reason": "x"}),
                #[cfg(unix)]
                json!({"path": "link.txt", "reason": "x"}),
                #[cfg(unix)]
                json!({"path": "innocent.txt", "reason": "x"}),
                json!({"path": "../etc/passwd", "reason": "x"}),
                json!({"path": "README.md", "reason": "kept"}),
            ]
            .into(),
        ),
        json!({"max_items": 10}),
    );
    assert_eq!(b["items"].as_array().unwrap().len(), 1);
    let causes: Vec<(&str, &str)> = b["omitted"]
        .as_array()
        .unwrap()
        .iter()
        .map(|o| (o["cause"].as_str().unwrap(), o["detail"].as_str().unwrap()))
        .collect();
    assert_eq!(
        causes,
        [
            ("blocked_sensitive", "secret_pattern"),
            ("blocked_sensitive", "secret_pattern"),
            ("blocked_sensitive", "ignored"),
            #[cfg(unix)]
            ("blocked_sensitive", "outside_repository"),
            #[cfg(unix)]
            ("blocked_sensitive", "secret_pattern"),
            ("blocked_sensitive", "outside_repository"),
        ]
    );
    assert!(
        !b.to_string().contains("do-not-leak"),
        "secret bytes leaked"
    );
}

/// CTX-BUDGET-TRUNCATE
#[test]
fn budget_truncation_records_omitted() {
    let d = repo();
    let items = json!([
        {"path": "src/a.rs", "reason": "primary"},
        {"path": "README.md", "reason": "secondary"},
        {"path": ".gitignore", "reason": "tertiary"}
    ]);
    let a_len = fs::read(d.path().join("src/a.rs")).unwrap().len() as u64;
    let b = bundle(d.path(), items.clone(), json!({"max_bytes": a_len + 1}));
    assert_eq!(b["items"].as_array().unwrap().len(), 1);
    assert_eq!(b["items"][0]["reason"], json!("primary"));
    assert!(b["budget"]["used_bytes"].as_u64().unwrap() <= a_len + 1);
    let dropped: Vec<&str> = b["omitted"]
        .as_array()
        .unwrap()
        .iter()
        .map(|o| {
            assert_eq!(o["cause"], json!("budget_truncated"));
            o["reason"].as_str().unwrap()
        })
        .collect();
    // .gitignore alone would still fit, but it ranks below a dropped item, so it goes too.
    assert_eq!(dropped, ["secondary", "tertiary"]);
    let b = bundle(d.path(), items, json!({"max_items": 2}));
    assert_eq!(b["items"].as_array().unwrap().len(), 2);
    assert_eq!(b["budget"]["max_bytes_source"], json!("default"));
}

const PROFILE: &str = r#"{"schema_version": 2, "project_id": "demo", "purpose": "demo",
 "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"]"#;

/// #2184 AC-1 (REQ-C8-01): no byte budget in the request or the profile → 256 KiB and the
/// source `default`, whether the budget is empty or missing; there is still no unbounded run.
#[test]
fn missing_budget_uses_the_256_kib_default_and_records_it() {
    let d = repo();
    let items = json!([{"path": "README.md", "reason": "x"}]);
    for req in [
        json!({"protocol_version": 1, "items": items, "budget": {}}),
        json!({"protocol_version": 1, "items": items}),
    ] {
        let (out, v) = run(&["context-bundle"], d.path(), Some(&req));
        assert!(out.status.success(), "{v}");
        let b = &v["result"]["budget"];
        assert_eq!(b["max_bytes"], json!(262_144));
        assert_eq!(b["max_bytes_source"], json!("default"));
    }
    // A profile without the key keeps the default.
    write(d.path(), ".moai-cli/project.yaml", &format!("{PROFILE}}}"));
    let b = bundle(d.path(), items, json!({}));
    assert_eq!(b["budget"]["max_bytes"], json!(262_144));
    assert_eq!(b["budget"]["max_bytes_source"], json!("default"));
}

/// #2184 AC-1: the profile key `context_max_bytes` redefines the default and truncates by
/// it; a request `max_bytes` still wins; a zero or unreadable profile is an error.
#[test]
fn profile_redefines_the_byte_budget_and_the_request_wins() {
    let d = repo();
    let items = json!([
        {"path": "src/a.rs", "reason": "primary"},
        {"path": "README.md", "reason": "secondary"}
    ]);
    let a_len = fs::read(d.path().join("src/a.rs")).unwrap().len() as u64;
    write(
        d.path(),
        ".moai-cli/project.yaml",
        &format!("{PROFILE}, \"context_max_bytes\": {}}}", a_len + 1),
    );
    let b = bundle(d.path(), items.clone(), json!({}));
    assert_eq!(b["budget"]["max_bytes"], json!(a_len + 1));
    assert_eq!(b["budget"]["max_bytes_source"], json!("profile"));
    assert_eq!(b["items"].as_array().unwrap().len(), 1);
    assert_eq!(b["omitted"][0]["cause"], json!("budget_truncated"));

    let b = bundle(d.path(), items.clone(), json!({"max_bytes": 1_000_000}));
    assert_eq!(b["budget"]["max_bytes_source"], json!("request"));
    assert_eq!(b["items"].as_array().unwrap().len(), 2);

    for bad in [
        format!("{PROFILE}, \"context_max_bytes\": 0}}"),
        "not json".to_string(),
    ] {
        write(d.path(), ".moai-cli/project.yaml", &bad);
        let req = json!({"protocol_version": 1, "items": items});
        let (out, _) = run(&["context-bundle"], d.path(), Some(&req));
        assert!(!out.status.success(), "a bad profile is never guessed past");
    }
}

/// Codex review #2057 (P1): a symlink inside the tree that points at an ignored file is
/// blocked like the file itself; checking only the requested name let it through.
#[cfg(unix)]
#[test]
fn symlink_to_an_ignored_file_is_excluded() {
    let d = repo();
    std::os::unix::fs::symlink("build/out.txt", d.path().join("alias.txt")).unwrap();
    let b = bundle(
        d.path(),
        json!([{"path": "alias.txt", "reason": "x"}]),
        json!({"max_items": 1}),
    );
    assert_eq!(b["items"], json!([]));
    assert_eq!(b["omitted"][0]["cause"], json!("blocked_sensitive"));
    assert_eq!(b["omitted"][0]["detail"], json!("ignored"));
    assert!(
        !b.to_string().contains("do-not-leak"),
        "ignored bytes leaked"
    );
}

/// Codex review #2057 (P2): moving a changed submodule to another commit changes the
/// overlay digest, so a bundle taken before is no longer fresh.
#[test]
fn submodule_revision_is_part_of_the_overlay() {
    let sub_src = TempDir::new().unwrap();
    git(sub_src.path(), &["init", "-q", "-b", "main"]);
    write(sub_src.path(), "lib.txt", "v1\n");
    git(sub_src.path(), &["add", "."]);
    git(sub_src.path(), &["commit", "-q", "-m", "v1"]);
    let d = repo();
    git(
        d.path(),
        &[
            "-c",
            "protocol.file.allow=always",
            "submodule",
            "add",
            "-q",
            sub_src.path().to_str().unwrap(),
            "sub",
        ],
    );
    git(d.path(), &["commit", "-q", "-m", "add sub"]);
    let sub = d.path().join("sub");
    let advance = |body: &str| {
        write(&sub, "lib.txt", body);
        git(&sub, &["commit", "-q", "-am", body.trim()]);
        let (_, v) = run(&["snapshot-describe"], d.path(), None);
        v["result"].clone()
    };
    let first = advance("v2\n");
    let second = advance("v3\n");
    assert_eq!(first["base_commit"], second["base_commit"]);
    assert_eq!(first["dirty_files"][0]["path"], json!("sub"));
    assert_ne!(
        first["dirty_overlay_digest"], second["dirty_overlay_digest"],
        "submodule moved but the overlay did not change"
    );
    // Uncommitted work inside the submodule cannot be digested: unknown, not falsely fresh.
    write(&sub, "lib.txt", "uncommitted\n");
    let (_, v) = run(&["snapshot-describe"], d.path(), None);
    assert_eq!(v["result"]["base_status"], json!("unknown"));
    assert_eq!(v["result"]["dirty_overlay_digest"], Value::Null);
}

/// CTX-TWO-WORKTREES
#[test]
fn worktrees_do_not_share_overlay() {
    let d = repo();
    let side = TempDir::new().unwrap();
    let wt = side.path().join("wt");
    git(
        d.path(),
        &["worktree", "add", "-q", "-b", "lane", wt.to_str().unwrap()],
    );
    write(d.path(), "README.md", "# demo\nprimary edit\n");
    write(&wt, "README.md", "# demo\nlane edit\n");
    let items = json!([{"path": "README.md", "reason": "x"}]);
    let a = bundle(d.path(), items.clone(), json!({"max_items": 1}));
    let b = bundle(&wt, items, json!({"max_items": 1}));
    assert_eq!(a["snapshot"]["base_commit"], b["snapshot"]["base_commit"]);
    assert_ne!(
        a["snapshot"]["worktree_path"],
        b["snapshot"]["worktree_path"]
    );
    assert_ne!(
        a["snapshot"]["dirty_overlay_digest"],
        b["snapshot"]["dirty_overlay_digest"]
    );
    assert_ne!(
        a["items"][0]["content_sha256"],
        b["items"][0]["content_sha256"]
    );
    assert_ne!(a["bundle_digest"], b["bundle_digest"]);
    assert_eq!(a["items"][0]["content"], json!("# demo\nprimary edit\n"));
    assert_eq!(b["items"][0]["content"], json!("# demo\nlane edit\n"));
    // The lane's bundle checked against the primary tree is stale there.
    let (_, v) = run(&["context-verify"], d.path(), Some(&verify_request(&b)));
    assert_eq!(v["result"]["items"][0]["freshness"], json!("stale"));
}

/// CTX-TOKEN-ESTIMATE-LABELED
#[test]
fn token_counts_are_labeled_estimates() {
    let d = repo();
    let b = bundle(
        d.path(),
        json!([{"path": "README.md", "reason": "x"}]),
        json!({"max_bytes": 100}),
    );
    let bytes = b["budget"]["used_bytes"].as_u64().unwrap();
    assert_eq!(
        b["tokens_estimate"],
        json!({"value": bytes.div_ceil(4), "method": "bytes_div_4", "model_scope": "none"})
    );
    assert_eq!(b["network"], json!("disabled"));
    assert_eq!(b["writes"], json!(false));
}

fn record(authority: &str, host: &str, usage: Value) -> Value {
    json!({
        "schema_version": 1,
        "decision_source": "native",
        "host": host,
        "input_digest": format!("sha256:{}", "a".repeat(64)),
        "options": ["implement", "ask the user"],
        "chosen": "implement",
        "confidence": "unknown",
        "authority": authority,
        "recorded_at": "2026-11-08T11:30:00Z",
        "context_usage_source": usage,
    })
}

/// NATIVE-DECISION-RECORD
#[test]
fn decision_record_never_carries_authority() {
    let d = repo();
    let args = [
        "decision-record",
        "--run",
        "run-1",
        "--permission-reference",
        "test-ref",
    ];
    let ok = record(
        "none",
        "claude",
        json!({"host": "claude", "field": "claude_statusline_context_window", "value": 12.5}),
    );
    let (out, v) = run(&args, d.path(), Some(&ok));
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    assert_eq!(v["result"]["authority"], json!("none"));
    assert_eq!(
        v["result"]["files_created"],
        json!([".moai-cli/decisions/run-1.jsonl"])
    );
    let (_, v2) = run(&args, d.path(), Some(&record("none", "codex", Value::Null)));
    assert_eq!(v2["result"]["files_created"], json!([]));
    let log = fs::read_to_string(d.path().join(".moai-cli/decisions/run-1.jsonl")).unwrap();
    let lines: Vec<Value> = log
        .lines()
        .map(|l| serde_json::from_str(l).unwrap())
        .collect();
    assert_eq!(lines.len(), 2);
    assert!(lines.iter().all(|l| l["authority"] == json!("none")));
    // Anything but `none`, a foreign host's usage source, or a chosen value outside the
    // options is refused, and nothing more is appended.
    let mut bad_choice = record("none", "claude", Value::Null);
    bad_choice["chosen"] = json!("deploy");
    for bad in [
        record("user", "claude", Value::Null),
        record("granted", "claude", Value::Null),
        record(
            "none",
            "codex",
            json!({"host": "claude", "field": "claude_statusline_context_window"}),
        ),
        record(
            "none",
            "codex",
            json!({"host": "codex", "field": "claude_statusline_context_window"}),
        ),
        bad_choice,
    ] {
        let (out, v) = run(&args, d.path(), Some(&bad));
        assert_eq!(out.status.code(), Some(1), "{bad}");
        assert_eq!(v["error"]["kind"], json!("invalid_input"), "{bad}");
    }
    assert_eq!(
        fs::read_to_string(d.path().join(".moai-cli/decisions/run-1.jsonl")).unwrap(),
        log
    );
    // A placeholder permission reference and a path-like run id are refused.
    let (_, v) = run(
        &[
            "decision-record",
            "--run",
            "r",
            "--permission-reference",
            "placeholder",
        ],
        d.path(),
        Some(&ok),
    );
    assert_eq!(v["error"]["kind"], json!("permission_reference_missing"));
    let (_, v) = run(
        &[
            "decision-record",
            "--run",
            "../x",
            "--permission-reference",
            "ref",
        ],
        d.path(),
        Some(&ok),
    );
    assert_eq!(v["error"]["kind"], json!("invalid_input"));
}

/// CTX-PRESERVE: the read-only commands leave the tree byte-identical, and decision-record
/// writes only its own log.
#[test]
fn context_commands_write_nothing_else() {
    let d = repo();
    write(d.path(), "AGENTS.md", "# rules\n");
    write(d.path(), ".moai/project.yaml", "go: state\n");
    let state = |p: &Path| {
        let mut out = Vec::new();
        for rel in [
            "AGENTS.md",
            ".moai/project.yaml",
            "README.md",
            "src/a.rs",
            ".env",
        ] {
            out.push(fs::read(p.join(rel)).unwrap());
        }
        out
    };
    let before = state(d.path());
    run(&["snapshot-describe"], d.path(), None);
    let b = bundle(
        d.path(),
        json!([{"path": "src/a.rs", "reason": "x"}]),
        json!({"max_items": 1}),
    );
    run(&["context-verify"], d.path(), Some(&verify_request(&b)));
    assert!(!d.path().join(".moai-cli").exists());
    run(
        &[
            "decision-record",
            "--run",
            "r",
            "--permission-reference",
            "ref",
        ],
        d.path(),
        Some(&record("none", "claude", Value::Null)),
    );
    assert_eq!(before, state(d.path()));
    let created: Vec<_> = fs::read_dir(d.path().join(".moai-cli"))
        .unwrap()
        .map(|e| e.unwrap().file_name())
        .collect();
    assert_eq!(created, ["decisions"]);
}
