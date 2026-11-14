//! #2838 (leader decisions D1–D2): `claim-verify` also takes its request as flags,
//! `--issue N --comment ID --lane NAME [--repo OWNER/REPO] [--allowed-author LOGIN]…`.
//! AC-1: the flag form and the stdin JSON form give the same verdict, envelope and exit for the
//! same claim comments. AC-2: a missing flag is a usage error that names what is missing.
//! The binary runs against a stub `gh` (no live GitHub call) in a temporary repository.
#![cfg(unix)]
mod common;

use common::CleanEnv;
use serde_json::json;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

struct Repo {
    _dir: tempfile::TempDir,
    root: PathBuf,
    home: PathBuf,
    gh: PathBuf,
    gh_log: PathBuf,
}

fn live_lease() -> String {
    let now = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs() as i64;
    moai_cli::forge::policy::format_utc(now + 3600)
}

fn git(root: &Path, home: &Path, args: &[&str]) {
    let out = common::git::command()
        .env("HOME", home)
        .arg("-C")
        .arg(root)
        .args(args)
        .output()
        .unwrap();
    assert!(out.status.success(), "git {args:?}: {out:?}");
}

/// A repository whose `origin` is `remote` (none when `None`), and a `gh` stub that answers
/// every call with `STUB_JSON` and appends its arguments to a log.
fn repo(remote: Option<&str>) -> Repo {
    let dir = tempfile::TempDir::new().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let root = base.join("project");
    let home = base.join("home");
    fs::create_dir_all(&root).unwrap();
    fs::create_dir_all(&home).unwrap();
    git(&root, &home, &["init", "-q"]);
    if let Some(url) = remote {
        git(&root, &home, &["remote", "add", "origin", url]);
    }
    let gh = base.join("gh-stub");
    let gh_log = base.join("gh.log");
    common::write_executable(
        &gh,
        "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$GH_LOG\"\nprintf 'HTTP/2.0 200 OK\\r\\n\\r\\n%s' \"$STUB_JSON\"\n",
    );
    Repo {
        _dir: dir,
        root,
        home,
        gh,
        gh_log,
    }
}

fn raw(id: i64, login: &str, session: &str, lease: &str) -> serde_json::Value {
    json!({
        "id": id,
        "user": {"login": login},
        "body": format!("moai-claim: session={session} epoch=1 lease_until={lease}"),
    })
}

/// Three claim lines on Issue 7: a forged one from another author first (id 50), then the
/// lanes `a` (100) and `b` (200), all live.
fn comments() -> String {
    let lease = live_lease();
    json!([
        raw(50, "mallory", "a", &lease),
        raw(100, "GoosLab", "a", &lease),
        raw(200, "GoosLab", "b", &lease),
    ])
    .to_string()
}

fn run(r: &Repo, args: &[&str], stdin: &str, body: &str) -> Output {
    let mut child = Command::new(common::MOAI_BIN)
        .clean_env()
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("HOME", &r.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("MOAI_FORGE_GH_BIN", &r.gh)
        .env("GH_LOG", &r.gh_log)
        .env("STUB_JSON", body)
        .arg("claim-verify")
        .arg(&r.root)
        .args(args)
        .current_dir(&r.root)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, stdin.as_bytes());
    child.wait_with_output().unwrap()
}

fn gh_calls(r: &Repo) -> String {
    let calls = fs::read_to_string(&r.gh_log).unwrap_or_default();
    fs::remove_file(&r.gh_log).ok();
    calls
}

fn text(bytes: &[u8]) -> String {
    String::from_utf8_lossy(bytes).into_owned()
}

/// AC-1: owner, not owner, and the author filter on and off — each case run in both forms
/// gives byte-identical stdout and stderr, the same exit, and one read of the same Issue.
#[test]
fn flags_and_stdin_give_the_same_verdict() {
    let r = repo(Some("https://github.com/acme/widgets.git"));
    let body = comments();
    // (session, comment id, allowed authors, expected exit)
    let cases: [(&str, i64, &[&str], i32); 4] = [
        ("a", 100, &["GoosLab"], 0),
        ("b", 200, &["GoosLab"], 1),
        // Without the filter the forged line 50 holds the Issue, so lane `a`'s 100 does not.
        ("a", 100, &[], 1),
        ("a", 50, &[], 0),
    ];
    for (session, id, authors, exit) in cases {
        let request = json!({
            "schema_version": 1,
            "owner": "acme",
            "repo": "widgets",
            "issue_number": 7,
            "session": session,
            "comment_id": id,
            "allowed_authors": authors,
        })
        .to_string();
        let from_stdin = run(&r, &[], &request, &body);
        let stdin_calls = gh_calls(&r);
        let id_text = id.to_string();
        let mut args = vec!["--issue", "7", "--comment", &id_text, "--lane", session];
        for a in authors {
            args.extend(["--allowed-author", a]);
        }
        // Stdin is not read in the flag form: unrelated bytes there change nothing.
        let from_flags = run(&r, &args, "not json", &body);
        let flag_calls = gh_calls(&r);
        let case = format!("{session} {id} {authors:?}");
        assert_eq!(
            from_stdin.status.code(),
            Some(exit),
            "{case}: {from_stdin:?}"
        );
        assert_eq!(
            from_flags.status.code(),
            Some(exit),
            "{case}: {from_flags:?}"
        );
        assert_eq!(text(&from_flags.stdout), text(&from_stdin.stdout), "{case}");
        assert_eq!(text(&from_flags.stderr), text(&from_stdin.stderr), "{case}");
        assert!(
            stdin_calls.contains("repos/acme/widgets/issues/7/comments"),
            "{stdin_calls}"
        );
        assert_eq!(flag_calls, stdin_calls, "{case}");
    }
}

/// `--repo` names the repository instead of the origin remote, and with neither a GitHub
/// origin nor `--repo` the error says to pass `--repo`.
#[test]
fn repo_flag_and_origin_remote() {
    let body = comments();
    let flags = ["--issue", "7", "--comment", "100", "--lane", "a"];

    let ssh = repo(Some("git@github.com:acme/widgets.git"));
    assert_eq!(run(&ssh, &flags, "", &body).status.code(), Some(1));
    assert!(gh_calls(&ssh).contains("repos/acme/widgets/issues/7/comments"));

    let mut with_repo = flags.to_vec();
    with_repo.extend(["--repo", "other-org/thing.rs"]);
    run(&ssh, &with_repo, "", &body);
    assert!(gh_calls(&ssh).contains("repos/other-org/thing.rs/issues/7/comments"));

    let none = repo(None);
    let out = run(&none, &flags, "", &body);
    assert_ne!(out.status.code(), Some(0));
    assert!(
        text(&out.stderr).contains("pass --repo OWNER/REPO"),
        "{out:?}"
    );
    assert_eq!(gh_calls(&none), "", "no GitHub read without a repository");
    let out = run(&none, &with_repo, "", &body);
    assert!(
        gh_calls(&none).contains("repos/other-org/thing.rs/issues/7/comments"),
        "{out:?}"
    );

    // An empty session name is refused by the shared request check, before any read.
    let out = run(
        &ssh,
        &["--issue", "7", "--comment", "100", "--lane", ""],
        "",
        &body,
    );
    assert_eq!(out.status.code(), Some(1), "{out:?}");
    assert!(
        text(&out.stderr).contains("session are required"),
        "{out:?}"
    );
    assert_eq!(gh_calls(&ssh), "", "empty --lane: no GitHub read");

    for bad in ["acme", "acme/widgets/extra", "/widgets", "acme/"] {
        let mut args = flags.to_vec();
        args.extend(["--repo", bad]);
        let out = run(&none, &args, "", &body);
        assert_ne!(out.status.code(), Some(0), "{bad}");
        let err = text(&out.stderr);
        assert!(
            err.contains("expected OWNER/REPO") || err.contains("owner, repo, issue_number"),
            "{bad}: {err}"
        );
        assert_eq!(gh_calls(&none), "", "{bad}: no GitHub read");
    }
}

/// AC-2: every incomplete or conflicting flag set is a usage error that names the missing or
/// conflicting flag, before any GitHub read: exit 2, except with `--hook`, where a usage error
/// exits 1 so a broken hook line never blocks the host's tool call (exit 2 blocks).
#[test]
fn a_missing_flag_is_a_usage_error_that_names_it() {
    let r = repo(Some("https://github.com/acme/widgets.git"));
    let body = comments();
    let cases: [(&[&str], &[&str], i32); 8] = [
        (&["--issue", "7"], &["--comment", "--lane"], 2),
        (&["--issue", "7", "--comment", "100"], &["--lane"], 2),
        (&["--issue", "7", "--lane", "a"], &["--comment"], 2),
        (&["--comment", "100", "--lane", "a"], &["--issue"], 2),
        (&["--repo", "acme/widgets"], &["--issue"], 2),
        (&["--allowed-author", "GoosLab"], &["--issue"], 2),
        (
            &["--issue", "0", "--comment", "100", "--lane", "a"],
            &["--issue"],
            2,
        ),
        (
            &["--hook", "--issue", "7", "--comment", "100", "--lane", "a"],
            &["--hook", "--issue"],
            1,
        ),
    ];
    for (args, named, exit) in cases {
        let out = run(&r, args, "", &body);
        let err = text(&out.stderr);
        assert_eq!(out.status.code(), Some(exit), "{args:?}: {err}");
        assert!(err.contains("error:"), "{args:?}: {err}");
        for flag in named {
            assert!(err.contains(flag), "{args:?} must name {flag}: {err}");
        }
        assert_eq!(gh_calls(&r), "", "{args:?}: no GitHub read");
    }
}

/// D2: the lane workflow shows the flag form first and keeps the stdin form.
#[test]
fn mode_agent_shows_the_flag_form() {
    let doc = fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/workflows/mode-agent.md"),
    )
    .unwrap();
    for phrase in [
        "`moai claim-verify . --issue <n> --comment <claim id> --lane <session> --allowed-author <login>`",
        "with the same typed request on stdin",
        "both forms give the same verdict, and a missing flag is a usage error that names it (Issue #2838)",
    ] {
        assert!(doc.contains(phrase), "mode-agent.md must say `{phrase}`");
    }
}
