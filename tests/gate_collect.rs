//! Issue #2674 S1b: `gate_collect::collect` reads one pull request into a gate snapshot —
//! GitHub through a recorded transport (GET only), git from a temporary repository. Each case
//! checks the field the gate decides from: changed files on both sides, merge result, known
//! verdict base, kept documentation lines, manifest match, the Self-Audit checker and a head
//! that was never fetched.

mod common;

use common::git::{git, git_ok};
use moai_cli::forge::transport::{Exchange, HttpRequest, HttpResponse, RecordedTransport};
use moai_cli::gate;
use moai_cli::gate_collect::{collect, lines_kept, manifest_matches, Target};
use serde_json::{json, Value};
use std::fs;
use std::path::Path;
use tempfile::TempDir;

const API: &str = "https://api.github.com/repos/o/r";

fn commit(dir: &Path, files: &[(&str, &str)], msg: &str) -> String {
    for (path, text) in files {
        let p = dir.join(path);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, text).unwrap();
    }
    git_ok(dir, &["add", "-A"]);
    git_ok(
        dir,
        &[
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@example.invalid",
            "commit",
            "-q",
            "-m",
            msg,
        ],
    );
    git(dir, &["rev-parse", "HEAD"]).trim().to_string()
}

/// A repository with base A, a PR branch B and a tip C. The checker stub passes a body that
/// holds `SELF-AUDIT-OK`.
struct Repo {
    _dir: TempDir,
    a: String,
    b: String,
    c: String,
}

fn repo(pr_files: &[(&str, &str)], tip_files: &[(&str, &str)]) -> Repo {
    let dir = TempDir::new().unwrap();
    let d = dir.path();
    git_ok(d, &["init", "-q", "-b", "main"]);
    let checker = "import sys\nsys.exit(0 if 'SELF-AUDIT-OK' in open(sys.argv[2]).read() else 1)\n";
    let a = commit(
        d,
        &[
            ("tools/check_pr_self_audit.py", checker),
            ("docs/n.md", "one\n\ntwo\n"),
            ("src/a.rs", "a\n"),
            ("src/b.rs", "b\n"),
        ],
        "base",
    );
    git_ok(d, &["checkout", "-q", "-b", "pr"]);
    let b = commit(d, pr_files, "pr");
    git_ok(d, &["checkout", "-q", "main"]);
    let c = commit(d, tip_files, "tip");
    Repo { _dir: dir, a, b, c }
}

fn ok(body: Value) -> HttpResponse {
    HttpResponse::new(200).with_body(body)
}

fn exchanges(head: &str, comments: Vec<Value>, body: &str) -> RecordedTransport {
    RecordedTransport::new(vec![
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/pulls/9")),
            response: ok(json!({"head": {"sha": head}, "body": body})),
        },
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/issues/9/comments?per_page=100")),
            response: ok(Value::Array(comments)),
        },
        Exchange {
            request: HttpRequest::new(
                "GET",
                format!("{API}/commits/{head}/check-runs?per_page=100"),
            ),
            response: ok(json!({"check_runs": [
                {"name": "ci", "conclusion": "success"},
                {"name": "ci / linux-x86_64", "conclusion": "success"},
                {"name": "windows-build", "conclusion": "skipped"}
            ]})),
        },
    ])
}

fn verdict(head: &str, base: &str) -> Value {
    json!({"id": 200, "body": format!("AUDIT #9 (x) head {head} base {base} — verdict: PASS\n\n감사: agent-10 (예약 100).\n")})
}

fn claim(head: &str) -> Value {
    json!({"id": 100, "body": format!("moai-audit-claim: lane=agent-10 head={head}\n")})
}

fn target(r: &Repo) -> Target<'_> {
    Target {
        dir: r._dir.path(),
        owner: "o",
        repo: "r",
        pr: 9,
        tip: "main",
        review_confirmed_prs: vec![],
    }
}

#[test]
fn a_clean_pull_request_collects_into_a_passing_snapshot() {
    let r = repo(&[("src/a.rs", "a2\n")], &[("src/b.rs", "b2\n")]);
    let t = exchanges(
        &r.b,
        vec![claim(&r.b), verdict(&r.b, &r.a)],
        "body SELF-AUDIT-OK",
    );
    let s = collect(&target(&r), &t).unwrap();
    assert_eq!(t.remaining(), 0);
    assert_eq!(s.head, r.b);
    assert_eq!(s.pr_files, vec!["src/a.rs".to_string()]);
    assert_eq!(s.tip_files_since_verdict_base, vec!["src/b.rs".to_string()]);
    assert!(s.merge_clean && s.verdict_base_known && s.body_check_ok);
    assert_eq!(s.checks.len(), 3);
    assert!(gate::check(&s).is_ok(), "{:?}", gate::check(&s));
    let _ = &r.c;
}

#[test]
fn the_checker_unknown_base_and_missing_head_are_read_not_assumed() {
    let r = repo(&[("src/a.rs", "a2\n")], &[("src/b.rs", "b2\n")]);
    // The body fails the checker.
    let t = exchanges(&r.b, vec![claim(&r.b), verdict(&r.b, &r.a)], "no section");
    assert!(!collect(&target(&r), &t).unwrap().body_check_ok);
    // A verdict base the repository does not have.
    let unknown = "abababababababababababababababababababab";
    let t = exchanges(
        &r.b,
        vec![claim(&r.b), verdict(&r.b, unknown)],
        "SELF-AUDIT-OK",
    );
    let s = collect(&target(&r), &t).unwrap();
    assert!(!s.verdict_base_known);
    assert_eq!(gate::check(&s).unwrap_err().rule, "G1");
    // A head that was never fetched is an error naming what to fetch.
    let missing = "cdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcd";
    let t = RecordedTransport::new(vec![
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/pulls/9")),
            response: ok(json!({"head": {"sha": missing}, "body": ""})),
        },
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/issues/9/comments?per_page=100")),
            response: ok(json!([])),
        },
        Exchange {
            request: HttpRequest::new(
                "GET",
                format!("{API}/commits/{missing}/check-runs?per_page=100"),
            ),
            response: ok(json!({"check_runs": []})),
        },
    ]);
    let err = collect(&target(&r), &t).unwrap_err();
    assert!(err.contains("fetch refs/pull/9/head"), "{err}");
}

#[test]
fn overlap_and_conflict_are_measured_from_the_merged_tree() {
    // Both sides add different lines to the same doc: the merge keeps both.
    let r = repo(
        &[("docs/n.md", "one\nPR line\n\ntwo\n")],
        &[("docs/n.md", "one\n\ntwo\ntip line\n")],
    );
    let t = exchanges(
        &r.b,
        vec![claim(&r.b), verdict(&r.b, &r.a)],
        "SELF-AUDIT-OK",
    );
    let s = collect(&target(&r), &t).unwrap();
    assert!(s.merge_clean);
    assert_eq!(s.docs_lines_kept, vec!["docs/n.md".to_string()]);
    assert!(gate::check(&s).is_ok(), "{:?}", gate::check(&s));
    // Both sides change the same line of code: a conflict.
    let r = repo(&[("src/b.rs", "pr\n")], &[("src/b.rs", "tip\n")]);
    let t = exchanges(
        &r.b,
        vec![claim(&r.b), verdict(&r.b, &r.a)],
        "SELF-AUDIT-OK",
    );
    let s = collect(&target(&r), &t).unwrap();
    assert!(!s.merge_clean);
    assert_eq!(gate::check(&s).unwrap_err().rule, "G6");
}

#[test]
fn comment_pages_are_followed() {
    let r = repo(&[("src/a.rs", "a2\n")], &[("src/b.rs", "b2\n")]);
    let page2 = format!("{API}/issues/9/comments?per_page=100&page=2");
    let t = RecordedTransport::new(vec![
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/pulls/9")),
            response: ok(json!({"head": {"sha": r.b}, "body": "SELF-AUDIT-OK"})),
        },
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/issues/9/comments?per_page=100")),
            response: ok(json!([claim(&r.b)]))
                .with_header("link", &format!("<{page2}>; rel=\"next\"")),
        },
        Exchange {
            request: HttpRequest::new("GET", page2),
            response: ok(json!([verdict(&r.b, &r.a)])),
        },
        Exchange {
            request: HttpRequest::new(
                "GET",
                format!("{API}/commits/{}/check-runs?per_page=100", r.b),
            ),
            response: ok(json!({"check_runs": [{"name": "ci", "conclusion": "success"}]})),
        },
    ]);
    let s = collect(&target(&r), &t).unwrap();
    assert_eq!(s.comments.len(), 2);
    assert!(s.verdict_base_known);
}

#[test]
fn the_manifest_check_needs_every_file_and_hash() {
    let dir = TempDir::new().unwrap();
    let d = dir.path();
    git_ok(d, &["init", "-q", "-b", "main"]);
    let sha = |s: &str| {
        use sha2::{Digest, Sha256};
        format!("{:x}", Sha256::digest(s.as_bytes()))
    };
    let manifest = |entries: Value| json!({"schema_version": 1, "files": entries}).to_string();
    commit(
        d,
        &[
            ("skills/moai/SKILL.md", "skill\n"),
            (
                "skills/moai/references/source-manifest.json",
                &manifest(json!({"SKILL.md": {"sha256": sha("skill\n")}})),
            ),
        ],
        "ok",
    );
    assert!(manifest_matches(d, "HEAD^{tree}").unwrap());
    commit(d, &[("skills/moai/SKILL.md", "changed\n")], "stale hash");
    assert!(!manifest_matches(d, "HEAD^{tree}").unwrap());
    commit(
        d,
        &[
            ("skills/moai/SKILL.md", "skill\n"),
            ("skills/moai/extra.md", "x\n"),
        ],
        "file without entry",
    );
    assert!(!manifest_matches(d, "HEAD^{tree}").unwrap());
}

/// A merged documentation file that lost a line one side added is not kept; one holding both
/// sides' lines is (mutation C4). The merged tree here is built by hand to lose the line.
#[test]
fn a_doc_that_lost_an_added_line_is_not_kept() {
    let r = repo(
        &[("docs/n.md", "one\nPR line\n\ntwo\n")],
        &[("docs/n.md", "one\n\ntwo\ntip line\n")],
    );
    let d = r._dir.path();
    let mb = git(d, &["merge-base", &r.b, &r.c]).trim().to_string();
    let sides = [(mb.as_str(), r.b.as_str()), (r.a.as_str(), r.c.as_str())];
    let good = commit(
        d,
        &[("docs/n.md", "one\nPR line\n\ntwo\ntip line\n")],
        "both",
    );
    assert!(lines_kept(d, &format!("{good}^{{tree}}"), "docs/n.md", &sides).unwrap());
    let lost = commit(
        d,
        &[("docs/n.md", "one\n\ntwo\ntip line\n")],
        "lost the PR line",
    );
    assert!(!lines_kept(d, &format!("{lost}^{{tree}}"), "docs/n.md", &sides).unwrap());
}

/// A manifest entry with no file fails (mutation C5); a repository without the Self-Audit
/// checker never reads as a passing body (mutation C7).
#[test]
fn an_entry_without_a_file_and_a_missing_checker_fail() {
    let dir = TempDir::new().unwrap();
    let d = dir.path();
    git_ok(d, &["init", "-q", "-b", "main"]);
    let sha = {
        use sha2::{Digest, Sha256};
        format!("{:x}", Sha256::digest(b"skill\n"))
    };
    let m = json!({"schema_version": 1, "files": {"SKILL.md": {"sha256": sha}, "gone.md": {"sha256": sha}}}).to_string();
    commit(
        d,
        &[
            ("skills/moai/SKILL.md", "skill\n"),
            ("skills/moai/references/source-manifest.json", &m),
        ],
        "entry without file",
    );
    assert!(!manifest_matches(d, "HEAD^{tree}").unwrap());

    let r = repo(&[("src/a.rs", "a2\n")], &[("src/b.rs", "b2\n")]);
    fs::remove_file(r._dir.path().join("tools/check_pr_self_audit.py")).unwrap();
    let t = exchanges(
        &r.b,
        vec![claim(&r.b), verdict(&r.b, &r.a)],
        "SELF-AUDIT-OK",
    );
    let s = collect(&target(&r), &t).unwrap();
    assert!(!s.body_check_ok);
    assert_eq!(gate::check(&s).unwrap_err().rule, "G4");
}
