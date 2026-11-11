//! `moai-cli search` (rc.2 C7 #2183) on temporary git projects: index generation or tracked
//! files as the file source, citations, refusals, and no writes.
mod common;

use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Command, Stdio},
};

fn git(dir: &Path, args: &[&str]) {
    let ok = common::git::command()
        .args(["-c", "user.name=t", "-c", "user.email=t@example.invalid"])
        .args(args)
        .current_dir(dir)
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .status()
        .unwrap()
        .success();
    assert!(ok, "git {args:?}");
}

fn project() -> tempfile::TempDir {
    let dir = tempfile::TempDir::new().unwrap();
    git(dir.path(), &["init", "-q"]);
    fs::write(dir.path().join("a.txt"), "alpha needle\nbeta\n").unwrap();
    fs::write(dir.path().join("b.txt"), "no match\n").unwrap();
    git(dir.path(), &["add", "-A"]);
    git(dir.path(), &["commit", "-q", "-m", "base"]);
    dir
}

fn run(dir: &Path, op: &str, input: &Value) -> (i32, Value) {
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .arg(op)
        .arg(dir)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut stdin = child.stdin.take().unwrap();
    match stdin.write_all(input.to_string().as_bytes()) {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    drop(stdin);
    let out = child.wait_with_output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
    (out.status.code().unwrap_or(-1), v)
}

fn result(v: &Value) -> &Value {
    v.get("result").unwrap_or(v)
}

#[test]
fn without_a_generation_the_tracked_files_are_searched_and_coverage_says_so() {
    let p = project();
    let (code, v) = run(
        p.path(),
        "search",
        &json!({"protocol_version": 1, "query": {"literals": ["needle"]}}),
    );
    assert_eq!(code, 0, "{v}");
    let r = result(&v);
    assert_eq!(r["coverage"]["file_source"], "tracked_files_fallback");
    assert_eq!(r["coverage"]["generation_id"], Value::Null);
    assert!(r["coverage"]["graph"]
        .as_str()
        .unwrap()
        .starts_with("unavailable"));
    assert_eq!(r["freshness"], "unavailable");
    let hits = r["results"].as_array().unwrap();
    assert_eq!(hits.len(), 1);
    assert_eq!(hits[0]["path"], "a.txt");
    assert_eq!(hits[0]["start_line"], 1);
    assert_eq!(hits[0]["grade"], "lexical");
    assert_eq!(hits[0]["state"], "fresh");
    assert_eq!(r["writes"], false);
    assert!(!p.path().join(".moai-cli").exists(), "search wrote nothing");
}

#[test]
fn with_a_generation_its_file_list_is_the_source() {
    let p = project();
    let (code, _) = run(p.path(), "index-build", &json!({"protocol_version": 1}));
    assert_eq!(code, 0);
    let before = fs::read_dir(p.path().join(".moai-cli/index"))
        .unwrap()
        .count();
    let (code, v) = run(
        p.path(),
        "search",
        &json!({"protocol_version": 1, "query": {"literals": ["needle"], "paths": ["b.txt"]}}),
    );
    assert_eq!(code, 0, "{v}");
    let r = result(&v);
    assert_eq!(r["coverage"]["file_source"], "index_generation");
    assert!(r["coverage"]["generation_id"].is_string());
    assert_eq!(r["freshness"], "fresh");
    let hits = r["results"].as_array().unwrap();
    assert_eq!(hits[0]["path"], "b.txt");
    assert_eq!(hits[0]["grade"], "exact");
    assert_eq!(hits[1]["path"], "a.txt");
    assert_eq!(
        fs::read_dir(p.path().join(".moai-cli/index"))
            .unwrap()
            .count(),
        before,
        "search wrote nothing into the index"
    );
}

#[test]
fn bad_requests_are_refused_with_the_bound_named() {
    let p = project();
    for (query, limits, needle) in [
        (json!({}), json!({}), "at least one path or literal"),
        (
            json!({"paths": ["../x"]}),
            json!({}),
            "relative inside the repository",
        ),
        (json!({"literals": [""]}), json!({}), "non-empty"),
        (json!({"literals": ["a\nb"]}), json!({}), "one line"),
        (
            json!({"literals": ["x".repeat(513)]}),
            json!({}),
            "512 bytes",
        ),
        (
            json!({"literals": ["a"]}),
            json!({"max_files": 0}),
            "max_files",
        ),
        (
            json!({"literals": ["a"]}),
            json!({"max_files": 2001}),
            "max_files",
        ),
        (
            json!({"literals": ["a"]}),
            json!({"max_spans_per_file": 1001}),
            "max_spans_per_file",
        ),
    ] {
        let (code, v) = run(
            p.path(),
            "search",
            &json!({"protocol_version": 1, "query": query, "limits": limits}),
        );
        assert_ne!(code, 0, "{query} {limits} -> {v}");
        assert!(v.to_string().contains(needle), "{needle}: {v}");
    }
    // The edge values themselves are accepted: 512 bytes, max_files 2000, spans 1000.
    for (query, limits) in [
        (json!({"literals": ["x".repeat(512)]}), json!({})),
        (
            json!({"literals": ["a"]}),
            json!({"max_files": 2000, "max_spans_per_file": 1000}),
        ),
        (
            json!({"literals": ["a"]}),
            json!({"max_files": 1, "max_spans_per_file": 1}),
        ),
    ] {
        let (code, v) = run(
            p.path(),
            "search",
            &json!({"protocol_version": 1, "query": query, "limits": limits}),
        );
        assert_eq!(code, 0, "{query} {limits} -> {v}");
    }
    let (code, _) = run(
        p.path(),
        "search",
        &json!({"protocol_version": 2, "query": {"literals": ["a"]}}),
    );
    assert_ne!(code, 0);
}

/// Pre-audit B2: paths + literals together are capped at 64 terms — 64 is accepted, 65 refused.
#[test]
fn the_query_term_limit_is_64() {
    let p = project();
    let terms = |n: usize| -> Vec<String> { (0..n).map(|i| format!("t{i}")).collect() };
    let (code, v) = run(
        p.path(),
        "search",
        &json!({"protocol_version": 1, "query": {"literals": terms(63), "paths": ["a.txt"]}}),
    );
    assert_eq!(code, 0, "64 terms: {v}");
    let (code, v) = run(
        p.path(),
        "search",
        &json!({"protocol_version": 1, "query": {"literals": terms(64), "paths": ["a.txt"]}}),
    );
    assert_ne!(code, 0, "65 terms: {v}");
    assert!(v.to_string().contains("too many query terms"), "{v}");
}
