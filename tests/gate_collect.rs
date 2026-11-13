//! Issue #2674 S1b: `gate_collect::collect` reads one pull request into a gate snapshot —
//! GitHub through a recorded transport (GET only), git from a temporary repository. Each case
//! checks the field the gate decides from: changed files on both sides, merge result, known
//! verdict base, kept documentation lines, manifest match, the Self-Audit checker and a head
//! that was never fetched.

mod common;

use common::git::{git, git_ok};
use moai_cli::forge::transport::{Exchange, HttpRequest, HttpResponse, RecordedTransport};
use moai_cli::gate;
use moai_cli::gate_collect::{self, collect, lines_kept, manifest_matches, Target};
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
            ("docs/a.md", "one\n\ntwo\n"),
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

/// The tip's push CI read (every attempt) for `sha`, answering `runs`.
fn tip_runs(sha: &str, runs: Value) -> Exchange {
    Exchange {
        request: HttpRequest::new(
            "GET",
            format!("{API}/commits/{sha}/check-runs?filter=all&per_page=100"),
        ),
        response: ok(json!({ "check_runs": runs })),
    }
}

/// The reads `collect` makes, with a green tip push CI on `tip`.
fn exchanges(head: &str, tip: &str, comments: Vec<Value>, body: &str) -> RecordedTransport {
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
        tip_runs(tip, json!([{"name": "ci", "conclusion": "success"}])),
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
        &r.c,
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
    let t = exchanges(
        &r.b,
        &r.c,
        vec![claim(&r.b), verdict(&r.b, &r.a)],
        "no section",
    );
    assert!(!collect(&target(&r), &t).unwrap().body_check_ok);
    // A verdict base the repository does not have.
    let unknown = "abababababababababababababababababababab";
    let t = exchanges(
        &r.b,
        &r.c,
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
        &[("docs/a.md", "one\nPR line\n\ntwo\n")],
        &[("docs/a.md", "one\n\ntwo\ntip line\n")],
    );
    let t = exchanges(
        &r.b,
        &r.c,
        vec![claim(&r.b), verdict(&r.b, &r.a)],
        "SELF-AUDIT-OK",
    );
    let s = collect(&target(&r), &t).unwrap();
    assert!(s.merge_clean);
    assert_eq!(s.docs_lines_kept, vec!["docs/a.md".to_string()]);
    assert!(gate::check(&s).is_ok(), "{:?}", gate::check(&s));
    // Both sides change the same line of code: a conflict.
    let r = repo(&[("src/b.rs", "pr\n")], &[("src/b.rs", "tip\n")]);
    let t = exchanges(
        &r.b,
        &r.c,
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
                .with_header("link", format!("<{page2}>; rel=\"next\"")),
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
        tip_runs(&r.c, json!([{"name": "ci", "conclusion": "success"}])),
    ]);
    let s = collect(&target(&r), &t).unwrap();
    assert_eq!(t.remaining(), 0);
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
            ("skills/moai/references/routes.json", "x\n"),
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
        &[("docs/a.md", "one\nPR line\n\ntwo\n")],
        &[("docs/a.md", "one\n\ntwo\ntip line\n")],
    );
    let d = r._dir.path();
    let mb = git(d, &["merge-base", &r.b, &r.c]).trim().to_string();
    let sides = [(mb.as_str(), r.b.as_str()), (r.a.as_str(), r.c.as_str())];
    let good = commit(
        d,
        &[("docs/a.md", "one\nPR line\n\ntwo\ntip line\n")],
        "both",
    );
    assert!(lines_kept(d, &format!("{good}^{{tree}}"), "docs/a.md", &sides).unwrap());
    let lost = commit(
        d,
        &[("docs/a.md", "one\n\ntwo\ntip line\n")],
        "lost the PR line",
    );
    assert!(!lines_kept(d, &format!("{lost}^{{tree}}"), "docs/a.md", &sides).unwrap());
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
        &r.c,
        vec![claim(&r.b), verdict(&r.b, &r.a)],
        "SELF-AUDIT-OK",
    );
    let s = collect(&target(&r), &t).unwrap();
    assert!(!s.body_check_ok);
    assert_eq!(gate::check(&s).unwrap_err().rule, "G4");
}

/// A verdict whose base is an earlier PR head (not on the tip) must not count the PR's own
/// changes as tip changes: the tip side is measured from where that base meets the tip.
#[test]
fn a_verdict_base_off_the_tip_is_not_a_false_overlap() {
    let r = repo(&[("src/a.rs", "a2\n")], &[("src/b.rs", "b2\n")]);
    let d = r._dir.path();
    // The PR got a second commit after the verdict; the verdict's base is the first PR head.
    git_ok(d, &["checkout", "-q", "pr"]);
    let b2 = commit(d, &[("src/c.rs", "c\n")], "pr 2");
    git_ok(d, &["checkout", "-q", "main"]);
    let t = exchanges(
        &b2,
        &r.c,
        vec![claim(&b2), verdict(&b2, &r.b)],
        "SELF-AUDIT-OK",
    );
    let s = collect(&target(&r), &t).unwrap();
    assert_eq!(s.tip_files_since_verdict_base, vec!["src/b.rs".to_string()]);
    assert!(gate::check(&s).is_ok(), "{:?}", gate::check(&s));
}

// ---- #2777 G8: the tip's push CI ----

fn run(conclusion: Option<&str>, id: u64) -> Value {
    json!({
        "name": "ci",
        "conclusion": conclusion,
        "details_url": format!("https://github.com/o/r/actions/runs/{id}/job/1"),
    })
}

/// The PR reads of `exchanges` followed by the given tip reads instead of one green tip.
fn with_tip(head: &str, body: &str, tips: Vec<Exchange>, base: &str) -> RecordedTransport {
    let mut ex = vec![
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/pulls/9")),
            response: ok(json!({"head": {"sha": head}, "body": body})),
        },
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/issues/9/comments?per_page=100")),
            response: ok(json!([claim(head), verdict(head, base)])),
        },
        Exchange {
            request: HttpRequest::new(
                "GET",
                format!("{API}/commits/{head}/check-runs?per_page=100"),
            ),
            response: ok(json!({"check_runs": [{"name": "ci", "conclusion": "success"}]})),
        },
    ];
    ex.extend(tips);
    RecordedTransport::new(ex)
}

fn g8(s: &gate::Snapshot) -> Option<String> {
    gate::check(s)
        .err()
        .filter(|e| e.rule == "G8")
        .map(|e| e.reason)
}

/// AC-3: a red tip stops an ordinary PR and lets a repair through; a green tip passes. A
/// failure whose re-run is still running stays red (the last completed result), and a
/// failure a re-run turned green is flaky — not a stop, as in flow-gate-verdict.
#[test]
fn a_red_tip_stops_an_ordinary_pr_and_passes_its_repair() {
    let r = repo(&[("src/a.rs", "a2\n")], &[("src/b.rs", "b2\n")]);
    let red = || tip_runs(&r.c, json!([run(Some("failure"), 77)]));

    let t = with_tip(&r.b, "SELF-AUDIT-OK", vec![red()], &r.a);
    let s = collect(&target(&r), &t).unwrap();
    assert_eq!(t.remaining(), 0);
    let ci = s.tip_ci.clone().unwrap();
    assert_eq!(
        (ci.state.as_str(), ci.decided_by.as_deref()),
        ("red", Some(r.c.as_str()))
    );
    assert!(!s.tip_repair);
    let reason = g8(&s).expect("G8 stops an ordinary PR on a red tip");
    assert!(reason.contains("run 77"), "{reason}");

    let body = "SELF-AUDIT-OK\nmoai-tip-repair: fixes run 77\n";
    let s = collect(&target(&r), &with_tip(&r.b, body, vec![red()], &r.a)).unwrap();
    assert!(s.tip_repair);
    assert!(gate::check(&s).is_ok(), "{:?}", gate::check(&s));

    // A completed failure with its re-run still running: red, the walk does not go back.
    let t = with_tip(
        &r.b,
        "SELF-AUDIT-OK",
        vec![tip_runs(
            &r.c,
            json!([run(Some("failure"), 77), run(None, 78)]),
        )],
        &r.a,
    );
    let s = collect(&target(&r), &t).unwrap();
    assert_eq!(t.remaining(), 0);
    assert!(g8(&s).is_some());

    // Failure then success on the same tip: flaky, not a stop.
    let t = with_tip(
        &r.b,
        "SELF-AUDIT-OK",
        vec![tip_runs(
            &r.c,
            json!([run(Some("failure"), 77), run(Some("success"), 78)]),
        )],
        &r.a,
    );
    let s = collect(&target(&r), &t).unwrap();
    assert_eq!(s.tip_ci.as_ref().unwrap().state, "flaky");
    assert!(gate::check(&s).is_ok(), "{:?}", gate::check(&s));
}

/// AC-1/AC-3: a running or only-cancelled tip is judged by the earlier tip commit's completed
/// CI — red there stops, green there passes; cancelled on every commit read is unknown.
#[test]
fn running_or_cancelled_tip_ci_is_judged_by_the_last_completed_one() {
    let r = repo(&[("src/a.rs", "a2\n")], &[("src/b.rs", "b2\n")]);
    // `main` is a → c: the tip c and its first parent a.
    for (tip_runs_now, parent, want, decided) in [
        (
            json!([run(None, 90)]),
            json!([run(Some("failure"), 80)]),
            "red",
            Some(&r.a),
        ),
        (
            json!([run(None, 90)]),
            json!([run(Some("success"), 80)]),
            "green",
            Some(&r.a),
        ),
        (
            json!([run(Some("cancelled"), 90), run(Some("cancelled"), 91)]),
            json!([run(Some("failure"), 80)]),
            "red",
            Some(&r.a),
        ),
        (
            json!([run(Some("cancelled"), 90)]),
            json!([run(Some("success"), 80)]),
            "green",
            Some(&r.a),
        ),
        (
            json!([]),
            json!([run(Some("success"), 80)]),
            "green",
            Some(&r.a),
        ),
        (
            json!([run(Some("cancelled"), 90)]),
            json!([run(Some("cancelled"), 80)]),
            "unknown",
            None,
        ),
    ] {
        let t = with_tip(
            &r.b,
            "SELF-AUDIT-OK",
            vec![
                tip_runs(&r.c, tip_runs_now.clone()),
                tip_runs(&r.a, parent.clone()),
            ],
            &r.a,
        );
        let s = collect(&target(&r), &t).unwrap();
        assert_eq!(t.remaining(), 0, "{tip_runs_now} / {parent}");
        let ci = s.tip_ci.clone().unwrap();
        assert_eq!(ci.state, want, "{tip_runs_now} / {parent}");
        assert_eq!(ci.decided_by.as_ref(), decided, "{tip_runs_now} / {parent}");
        assert_eq!(ci.tip, r.c);
        assert_eq!(
            g8(&s).is_some(),
            want != "green",
            "{tip_runs_now} / {parent}"
        );
    }
}

/// A tip read that fails is unknown and stops, never a pass; so does a repair on it.
#[test]
fn an_unreadable_tip_ci_is_unknown_and_stops_even_a_repair() {
    let r = repo(&[("src/a.rs", "a2\n")], &[("src/b.rs", "b2\n")]);
    let failing = Exchange {
        request: HttpRequest::new(
            "GET",
            format!("{API}/commits/{}/check-runs?filter=all&per_page=100", r.c),
        ),
        response: HttpResponse::new(502),
    };
    let body = "SELF-AUDIT-OK\nmoai-tip-repair: x\n";
    let t = with_tip(&r.b, body, vec![failing], &r.a);
    let s = collect(&target(&r), &t).unwrap();
    assert_eq!(t.remaining(), 0);
    let ci = s.tip_ci.clone().unwrap();
    assert_eq!(ci.state, "unknown");
    assert!(ci.detail.contains("HTTP 502"), "{}", ci.detail);
    assert!(g8(&s).is_some());
}

/// The walk reads at most `TIP_WALK` (10) tip commits: completed CI on the 10th decides; with
/// none in 10 the 11th is never read and the result is unknown.
#[test]
fn the_tip_walk_stops_at_its_limit() {
    let r = repo(&[("src/a.rs", "a2\n")], &[("src/b.rs", "b2\n")]);
    let d = r._dir.path();
    // Nine more commits on main: c is now the 10th from the tip, a the 11th.
    let mut chain = vec![r.c.clone()];
    for i in 0..9 {
        chain.push(commit(d, &[("src/n.rs", &format!("{i}\n"))], "tip n"));
    }
    chain.reverse(); // tip first
    assert_eq!(gate_collect::TIP_WALK, 10);
    assert_eq!(chain.len(), 10);
    let mut tips: Vec<Exchange> = chain[..9]
        .iter()
        .map(|sha| tip_runs(sha, json!([run(None, 1)])))
        .collect();
    tips.push(tip_runs(&chain[9], json!([run(Some("failure"), 10)])));
    let t = with_tip(&r.b, "SELF-AUDIT-OK", tips, &r.a);
    let s = collect(&target(&r), &t).unwrap();
    assert_eq!(t.remaining(), 0);
    let ci = s.tip_ci.unwrap();
    assert_eq!(
        (ci.state.as_str(), ci.decided_by),
        ("red", Some(r.c.clone()))
    );

    let tips: Vec<Exchange> = chain
        .iter()
        .map(|sha| tip_runs(sha, json!([run(None, 1)])))
        .chain([tip_runs(&r.a, json!([run(Some("success"), 11)]))])
        .collect();
    let t = with_tip(&r.b, "SELF-AUDIT-OK", tips, &r.a);
    let s = collect(&target(&r), &t).unwrap();
    assert_eq!(t.remaining(), 1, "the 11th commit is not read");
    let ci = s.tip_ci.unwrap();
    assert_eq!(ci.state, "unknown");
    assert!(ci.detail.contains("last 10 tip commit"), "{}", ci.detail);
}
