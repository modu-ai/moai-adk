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
/// A PR body that `pr-self-audit-check` passes (the four measured lines, #2279; #2708 made the
/// gate's body check the binary's own).
const PASS_BODY: &str = "Refs #2200\nMoai-Claim: 9#555\n\n## Self-Audit\n\n\
- mutation: claim-verify guard mutated (claim filter removed) -> check FAILED, measured\n\
- clone: check set from a fresh clone, not the /tmp symlink path\n\
- host: api.github.com exercised via fake gh executable, no network\n\
- receipt: receipt digest re-read and matched before PASS recorded\n";

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

/// A repository with base A, a PR branch B and a tip C. No checker file: the body check is
/// the binary's own (#2708), so a passing body is [`PASS_BODY`].
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
    let a = commit(
        d,
        &[
            ("docs/a.md", "one\n\ntwo\n"),
            ("src/a.rs", "a\n"),
            ("src/b.rs", "b\n"),
        ],
        "base",
    );
    git_ok(d, &["checkout", "-q", "-b", "pr"]);
    let b = commit(d, pr_files, "pr\n\nMoai-Claim: 9#555");
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
        claim_comment(555, "moai-claim: session=agent-26 epoch=1"),
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
        PASS_BODY,
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
        PASS_BODY,
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
        PASS_BODY,
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
        PASS_BODY,
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
            response: ok(json!({"head": {"sha": r.b}, "body": PASS_BODY})),
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
        claim_comment(555, "moai-claim: session=agent-26 epoch=1"),
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

/// A manifest entry with no file fails (mutation C5); a body without a measured Self-Audit
/// section never reads as passing (mutation C7; since #2708 the check is the binary's own
/// `pr-self-audit-check`, so the old stub marker alone is such a body).
#[test]
fn an_entry_without_a_file_and_a_body_without_the_section_fail() {
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
    let t = exchanges(
        &r.b,
        &r.c,
        vec![claim(&r.b), verdict(&r.b, &r.a)],
        "body SELF-AUDIT-OK",
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
    let b2 = commit(d, &[("src/c.rs", "c\n")], "pr 2\n\nMoai-Claim: 9#555");
    git_ok(d, &["checkout", "-q", "main"]);
    let t = exchanges(&b2, &r.c, vec![claim(&b2), verdict(&b2, &r.b)], PASS_BODY);
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
        claim_comment(555, "moai-claim: session=agent-26 epoch=1"),
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

    let t = with_tip(&r.b, PASS_BODY, vec![red()], &r.a);
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

    let body = &format!("{PASS_BODY}moai-tip-repair: fixes run 77\n");
    let s = collect(&target(&r), &with_tip(&r.b, body, vec![red()], &r.a)).unwrap();
    assert!(s.tip_repair);
    assert!(gate::check(&s).is_ok(), "{:?}", gate::check(&s));

    // A completed failure with its re-run still running: red, the walk does not go back.
    let t = with_tip(
        &r.b,
        PASS_BODY,
        vec![tip_runs(
            &r.c,
            json!([run(Some("failure"), 77), run(None, 78)]),
        )],
        &r.a,
    );
    let s = collect(&target(&r), &t).unwrap();
    assert_eq!(t.remaining(), 0);
    assert_eq!(
        t.sent.borrow().len(),
        5,
        "PR, comments, checks, claim, tip: the earlier tip commit is not read"
    );
    let ci = s.tip_ci.clone().unwrap();
    assert_eq!(
        (ci.state.as_str(), ci.decided_by.as_deref()),
        ("red", Some(r.c.as_str()))
    );
    assert!(g8(&s).is_some());

    // Failure then success on the same tip: flaky, not a stop.
    let t = with_tip(
        &r.b,
        PASS_BODY,
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
            PASS_BODY,
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
    let body = &format!("{PASS_BODY}moai-tip-repair: x\n");
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
    let t = with_tip(&r.b, PASS_BODY, tips, &r.a);
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
    let t = with_tip(&r.b, PASS_BODY, tips, &r.a);
    let s = collect(&target(&r), &t).unwrap();
    assert_eq!(t.remaining(), 1, "the 11th commit is not read");
    let ci = s.tip_ci.unwrap();
    assert_eq!(ci.state, "unknown");
    assert!(ci.detail.contains("last 10 tip commit"), "{}", ci.detail);
}

/// Audit 6863284751 B2: the tip's check runs are read over at most 10 pages. Ten pages, the
/// last without `next`, are all read (a failure on page 10 decides red); a tenth page that
/// points to an 11th stops before the 11th request, and the result is `unknown`, which stops
/// a repair too.
#[test]
fn the_tip_check_run_pages_stop_at_ten() {
    let r = repo(&[("src/a.rs", "a2\n")], &[("src/b.rs", "b2\n")]);
    let first = format!("{API}/commits/{}/check-runs?filter=all&per_page=100", r.c);
    let page_url = |n: usize| {
        if n == 1 {
            first.clone()
        } else {
            format!("{API}/tip-pages/{n}")
        }
    };
    // `pages` pages; the last carries `next` to page `pages + 1` when `more`.
    let pages = |count: usize, more: bool, last_runs: Value| -> Vec<Exchange> {
        (1..=count)
            .map(|n| {
                let runs = if n == count {
                    last_runs.clone()
                } else {
                    json!([run(Some("success"), n as u64)])
                };
                let mut resp = ok(json!({ "check_runs": runs }));
                if n < count || more {
                    resp = resp.with_header("link", format!("<{}>; rel=\"next\"", page_url(n + 1)));
                }
                Exchange {
                    request: HttpRequest::new("GET", page_url(n)),
                    response: resp,
                }
            })
            .collect()
    };

    let t = with_tip(
        &r.b,
        PASS_BODY,
        pages(10, false, json!([run(Some("failure"), 10)])),
        &r.a,
    );
    let s = collect(&target(&r), &t).unwrap();
    assert_eq!(t.remaining(), 0, "all ten pages are read");
    // Success on pages 1–9 and a failure on page 10 on the same commit: flaky, which proves
    // page 10 was judged together with the others.
    assert_eq!(s.tip_ci.as_ref().unwrap().state, "flaky");

    let mut tips = pages(10, true, json!([run(Some("success"), 10)]));
    tips.push(Exchange {
        request: HttpRequest::new("GET", page_url(11)),
        response: ok(json!({ "check_runs": [run(Some("success"), 11)] })),
    });
    let body = &format!("{PASS_BODY}moai-tip-repair: x\n");
    let t = with_tip(&r.b, body, tips, &r.a);
    let s = collect(&target(&r), &t).unwrap();
    assert_eq!(t.remaining(), 1, "the 11th page is never requested");
    let ci = s.tip_ci.clone().unwrap();
    assert_eq!(ci.state, "unknown");
    assert!(
        ci.detail.contains("more than 10 check run pages"),
        "{}",
        ci.detail
    );
    assert!(s.tip_repair);
    assert!(g8(&s).is_some(), "unknown stops a repair too");
}

/// Audit 6863284751 B3: the walk follows first parents only. On a tip that merged a PR, the
/// merged PR branch's head carries that PR's own `ci`; it must never decide the tip. Here the
/// tip (a merge) is still running, its first parent (the previous tip) is red and the merged
/// branch head — committed later, so a date-ordered walk would read it first — is green: the
/// state is red, decided by the first parent, and the branch head is never read.
#[test]
fn the_walk_follows_first_parents_not_a_merged_branch() {
    let r = repo(&[("src/a.rs", "a2\n")], &[("src/b.rs", "b2\n")]);
    let d = r._dir.path();
    git_ok(d, &["checkout", "-q", "-b", "merged-pr", &r.a]);
    fs::write(d.join("src/m.rs"), "m\n").unwrap();
    git_ok(d, &["add", "-A"]);
    let out = common::git::command()
        .current_dir(d)
        .env("GIT_COMMITTER_DATE", "2030-01-01T00:00:00Z")
        .env("GIT_AUTHOR_DATE", "2030-01-01T00:00:00Z")
        .args([
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@example.invalid",
            "commit",
            "-q",
            "-m",
            "merged pr head",
        ])
        .output()
        .unwrap();
    assert!(out.status.success());
    let branch_head = git(d, &["rev-parse", "HEAD"]).trim().to_string();
    git_ok(d, &["checkout", "-q", "main"]);
    git_ok(
        d,
        &[
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@example.invalid",
            "merge",
            "-q",
            "--no-ff",
            "-m",
            "Merge merged-pr",
            "merged-pr",
        ],
    );
    let tip = git(d, &["rev-parse", "HEAD"]).trim().to_string();
    // Without --first-parent, git's date order puts the 2030 branch head right after the tip.
    let all = git(d, &["rev-list", "--max-count=2", &tip]);
    assert_eq!(all.lines().nth(1), Some(branch_head.as_str()));

    let t = with_tip(
        &r.b,
        PASS_BODY,
        vec![
            tip_runs(&tip, json!([run(None, 90)])),
            tip_runs(&r.c, json!([run(Some("failure"), 80)])),
        ],
        &r.a,
    );
    let s = collect(&target(&r), &t).unwrap();
    assert_eq!(t.remaining(), 0);
    let ci = s.tip_ci.clone().unwrap();
    assert_eq!(
        (ci.state.as_str(), ci.decided_by.as_deref()),
        ("red", Some(r.c.as_str()))
    );
    assert!(
        !t.sent.borrow().iter().any(|q| q.url.contains(&branch_head)),
        "the merged branch head is never read"
    );
    assert!(g8(&s).is_some());
}

fn claim_comment(id: u64, first_line: &str) -> Exchange {
    Exchange {
        request: HttpRequest::new("GET", format!("{API}/issues/comments/{id}")),
        response: ok(json!({"id": id, "body": format!("{first_line}\n\nmore text")})),
    }
}

/// A PR whose own commit carries `Moai-Claim: 7#555`, then a merge of the tip whose message
/// names another claim (merges are skipped, so 999 is never read).
fn repo_with_trailers(trailer: &str) -> Repo {
    let r = repo(&[("src/a.rs", "a2\n")], &[("src/b.rs", "b2\n")]);
    let d = r._dir.path();
    git_ok(d, &["checkout", "-q", "pr"]);
    commit(d, &[("src/a.rs", "a3\n")], &format!("fix\n\n{trailer}"));
    git_ok(
        d,
        &[
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@example.invalid",
            "merge",
            "-q",
            "--no-ff",
            "-m",
            "Merge main\n\nMoai-Claim: 7#999",
            "main",
        ],
    );
    let head = git(d, &["rev-parse", "HEAD"]).trim().to_string();
    git_ok(d, &["checkout", "-q", "main"]);
    Repo {
        _dir: r._dir,
        a: r.a,
        b: head,
        c: r.c,
    }
}

/// #2674 R1: the authors are the sessions of the claims the PR's own commits name.
#[test]
fn the_authors_are_the_sessions_of_the_claims_in_the_trailers() {
    let r = repo_with_trailers("Moai-Claim: 7#555");
    let ex = RecordedTransport::new(vec![
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/pulls/9")),
            response: ok(json!({"head": {"sha": r.b}, "body": PASS_BODY})),
        },
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/issues/9/comments?per_page=100")),
            response: ok(json!([claim(&r.b), verdict(&r.b, &r.a)])),
        },
        Exchange {
            request: HttpRequest::new(
                "GET",
                format!("{API}/commits/{}/check-runs?per_page=100", r.b),
            ),
            response: ok(json!({"check_runs": [{"name": "ci", "conclusion": "success"}]})),
        },
        claim_comment(
            555,
            "moai-claim: session=agent-26 epoch=1 lease_until=2026-11-13T19:00:00Z",
        ),
        tip_runs(&r.c, json!([{"name": "ci", "conclusion": "success"}])),
    ]);
    let s = collect(&target(&r), &ex).unwrap();
    assert_eq!(s.pr_authors, Some(vec!["agent-26".to_string()]));
    assert_eq!(ex.remaining(), 0, "the merge's claim 999 is never read");
    // The gate then stops a verdict by that author.
    let mut by_author = s.clone();
    by_author.comments = vec![gate::Comment {
        id: 200,
        body: format!(
            "AUDIT #9 (x) head {} base {} — verdict: PASS\n\n감사: agent-26.\n",
            r.b, r.a
        ),
    }];
    assert_eq!(gate::check(&by_author).unwrap_err().rule, "G9");
}

#[test]
fn a_trailer_or_claim_line_that_cannot_be_read_stops_the_collection() {
    for (trailer, comment_line, needle) in [
        ("Moai-Claim: not-a-claim", None, "does not read"),
        (
            "Moai-Claim: 7#555",
            Some("just a note"),
            "no `moai-claim: session=` line",
        ),
    ] {
        let r = repo_with_trailers(trailer);
        let mut ex = vec![
            Exchange {
                request: HttpRequest::new("GET", format!("{API}/pulls/9")),
                response: ok(json!({"head": {"sha": r.b}, "body": PASS_BODY})),
            },
            Exchange {
                request: HttpRequest::new("GET", format!("{API}/issues/9/comments?per_page=100")),
                response: ok(json!([claim(&r.b), verdict(&r.b, &r.a)])),
            },
            Exchange {
                request: HttpRequest::new(
                    "GET",
                    format!("{API}/commits/{}/check-runs?per_page=100", r.b),
                ),
                response: ok(json!({"check_runs": []})),
            },
        ];
        if let Some(line) = comment_line {
            ex.push(claim_comment(555, line));
        }
        let t = RecordedTransport::new(ex);
        let err = collect(&target(&r), &t).unwrap_err();
        assert!(err.contains(needle), "{trailer}: {err}");
    }
}

/// A PR with `n` commits, commit `i` carrying `Moai-Claim: 7#<ids[i]>`.
fn repo_with_claim_ids(ids: &[u64]) -> Repo {
    let r = repo(&[("src/a.rs", "a2\n")], &[("src/b.rs", "b2\n")]);
    let d = r._dir.path();
    git_ok(d, &["checkout", "-q", "pr"]);
    for (i, id) in ids.iter().enumerate() {
        commit(
            d,
            &[("src/a.rs", &format!("a{i}\n"))],
            &format!("step {i}\n\nMoai-Claim: 7#{id}"),
        );
    }
    let head = git(d, &["rev-parse", "HEAD"]).trim().to_string();
    git_ok(d, &["checkout", "-q", "main"]);
    Repo {
        _dir: r._dir,
        a: r.a,
        b: head,
        c: r.c,
    }
}

fn three_reads(r: &Repo) -> Vec<Exchange> {
    vec![
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/pulls/9")),
            response: ok(json!({"head": {"sha": r.b}, "body": PASS_BODY})),
        },
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/issues/9/comments?per_page=100")),
            response: ok(json!([claim(&r.b), verdict(&r.b, &r.a)])),
        },
        Exchange {
            request: HttpRequest::new(
                "GET",
                format!("{API}/commits/{}/check-runs?per_page=100", r.b),
            ),
            response: ok(json!({"check_runs": []})),
        },
    ]
}

/// Audit 6863303949 B1: at most 20 distinct claims are read — 20 pass with 20 claim reads,
/// 21 stop the collection before any claim read.
#[test]
fn twenty_claims_are_read_and_twenty_one_are_refused() {
    // The fixture's first PR commit names claim 555, so 19 more make 20.
    let ids: Vec<u64> = (1..=19).map(|i| 1000 + i).collect();
    let r = repo_with_claim_ids(&ids);
    let mut ex = three_reads(&r);
    // `git log` lists the newest commit first, so the claims are read in reverse order.
    for (i, id) in ids.iter().enumerate().rev() {
        ex.push(claim_comment(
            *id,
            &format!("moai-claim: session=agent-{i} epoch=1"),
        ));
    }
    ex.push(claim_comment(555, "moai-claim: session=agent-26 epoch=1"));
    ex.push(tip_runs(
        &r.c,
        json!([{"name": "ci", "conclusion": "success"}]),
    ));
    let t = RecordedTransport::new(ex);
    let s = collect(&target(&r), &t).unwrap();
    assert_eq!(s.pr_authors.as_ref().unwrap().len(), 20);
    assert_eq!(t.remaining(), 0);

    let ids: Vec<u64> = (1..=20).map(|i| 1000 + i).collect();
    let r = repo_with_claim_ids(&ids);
    let t = RecordedTransport::new(three_reads(&r));
    let err = collect(&target(&r), &t).unwrap_err();
    assert!(err.contains("more than 20 claims"), "{err}");
    assert_eq!(t.sent.borrow().len(), 3, "no claim comment is read");
}

/// Audit 6863303949 N1: several commits naming one claim read it once.
#[test]
fn one_claim_named_by_several_commits_is_read_once() {
    let r = repo_with_claim_ids(&[555, 555, 555]);
    let mut ex = three_reads(&r);
    ex.push(claim_comment(555, "moai-claim: session=agent-26 epoch=2"));
    ex.push(tip_runs(
        &r.c,
        json!([{"name": "ci", "conclusion": "success"}]),
    ));
    let t = RecordedTransport::new(ex);
    let s = collect(&target(&r), &t).unwrap();
    assert_eq!(s.pr_authors, Some(vec!["agent-26".to_string()]));
    assert_eq!(t.remaining(), 0);
    assert_eq!(t.sent.borrow().len(), 5, "three reads, one claim, the tip");
}

/// Audit 6864692054 B1: G10 and G11 carry over `flow-gate-verdict`'s claim-line checks. A PR
/// commit with no `Moai-Claim` trailer is listed (a claim line followed by another paragraph is
/// not a trailer); a merge commit is a record and is not; the body's claim line is read.
#[test]
fn commits_without_a_claim_trailer_and_the_body_claim_are_collected() {
    let r = repo(&[("src/a.rs", "a2\n")], &[("src/b.rs", "b2\n")]);
    let d = r._dir.path();
    git_ok(d, &["checkout", "-q", "pr"]);
    let bare = commit(d, &[("src/c.rs", "c\n")], "no claim");
    let buried = commit(
        d,
        &[("src/d.rs", "d\n")],
        "claim not in the trailer block\n\nMoai-Claim: 9#555\n\nsigned off below",
    );
    git_ok(
        d,
        &[
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@example.invalid",
            "merge",
            "-q",
            "--no-ff",
            "-m",
            "Merge main",
            "main",
        ],
    );
    let head = git(d, &["rev-parse", "HEAD"]).trim().to_string();
    git_ok(d, &["checkout", "-q", "main"]);
    let t = exchanges(
        &head,
        &r.c,
        vec![claim(&head), verdict(&head, &r.a)],
        &PASS_BODY.replace("Moai-Claim: 9#555\n", ""),
    );
    let s = collect(&target(&r), &t).unwrap();
    assert_eq!(t.remaining(), 0);
    let mut missing = s.commits_without_claim.clone().unwrap();
    missing.sort();
    let mut want = vec![bare, buried];
    want.sort();
    assert_eq!(
        missing, want,
        "the merge commit and the claimed commit are not listed"
    );
    assert!(!s.body_claim);
    assert_eq!(gate::check(&s).unwrap_err().rule, "G10");

    // The same PR with the two commits repaired: only the body is left, and G11 stops it.
    let mut fixed = s.clone();
    fixed.commits_without_claim = Some(vec![]);
    assert_eq!(gate::check(&fixed).unwrap_err().rule, "G11");
    fixed.body_claim = true;
    assert!(gate::check(&fixed).is_ok(), "{:?}", gate::check(&fixed));
}

// ---- #2833: the mechanical delta ----

const EIGHT: &str = "1\n2\n3\n4\n5\n6\n7\n8\n";

fn git_as(dir: &Path, args: &[&str]) {
    let mut all = vec!["-c", "user.name=t", "-c", "user.email=t@example.invalid"];
    all.extend_from_slice(args);
    git_ok(dir, &all);
}

/// Base A; the PR's verdict head H changes line 1 of src/a.rs and adds src/p.rs; the tip C
/// changes line 8 of src/a.rs (the same file, other lines) and src/b.rs. The work tree is left
/// on the PR branch at H.
struct Delta {
    dir: TempDir,
    a: String,
    h: String,
    c: String,
}

fn delta_repo() -> Delta {
    let dir = TempDir::new().unwrap();
    let d = dir.path();
    git_ok(d, &["init", "-q", "-b", "main"]);
    let a = commit(
        d,
        &[
            ("src/a.rs", EIGHT),
            ("src/b.rs", "b\n"),
            ("src/c.rs", "c\n"),
        ],
        "base",
    );
    git_ok(d, &["checkout", "-q", "-b", "pr"]);
    let h = commit(
        d,
        &[
            ("src/a.rs", "one\n2\n3\n4\n5\n6\n7\n8\n"),
            ("src/p.rs", "p\n"),
        ],
        "pr\n\nMoai-Claim: 9#555",
    );
    git_ok(d, &["checkout", "-q", "main"]);
    let c = commit(
        d,
        &[
            ("src/a.rs", "1\n2\n3\n4\n5\n6\n7\neight\n"),
            ("src/b.rs", "b2\n"),
        ],
        "tip",
    );
    git_ok(d, &["checkout", "-q", "pr"]);
    Delta { dir, a, h, c }
}

impl Delta {
    fn d(&self) -> &Path {
        self.dir.path()
    }
    fn head(&self) -> String {
        git(self.d(), &["rev-parse", "HEAD"]).trim().to_string()
    }
    /// The leader loop's update-branch: merge the tip into the PR branch.
    fn merge_tip(&self) -> String {
        git_as(
            self.d(),
            &["merge", "-q", "--no-ff", "-m", "Merge main", "main"],
        );
        self.head()
    }
    /// A merge of the tip whose commit also carries `edit`.
    fn merge_tip_with(&self, edit: &[(&str, &str)]) -> String {
        git_as(self.d(), &["merge", "-q", "--no-ff", "--no-commit", "main"]);
        commit(self.d(), edit, "Merge main")
    }
    fn delta(&self, to: &str) -> gate::AutoDelta {
        gate_collect::mechanical_delta(self.d(), &self.h, to, "main").unwrap()
    }
}

/// AC-1 (accepted): a merge of only the tip — even with the tip changing other lines of a PR
/// file — keeps every PR file's patch-id.
#[test]
fn a_tip_only_merge_is_a_mechanical_delta() {
    let r = delta_repo();
    let n = r.merge_tip();
    let d = r.delta(&n);
    assert!(d.ok, "{}", d.detail);
    let paths: Vec<&str> = d.files.iter().map(|f| f.path.as_str()).collect();
    assert_eq!(paths, ["src/a.rs", "src/p.rs"]);
    for f in &d.files {
        assert!(
            !f.patch_id_from.is_empty() && f.patch_id_from == f.patch_id_to,
            "{f:?}"
        );
    }
    assert_eq!((d.from.as_str(), d.to.as_str()), (r.h.as_str(), n.as_str()));
    let _ = (&r.a, &r.c);
}

/// AC-1 (refused): the PR's author added a commit after the verdict head.
#[test]
fn an_author_commit_after_the_verdict_is_not_a_mechanical_delta() {
    let r = delta_repo();
    r.merge_tip();
    let n = commit(r.d(), &[("src/p.rs", "p2\n")], "more\n\nMoai-Claim: 9#555");
    let d = r.delta(&n);
    assert!(!d.ok);
    assert!(
        d.detail
            .contains("1 commit(s) after the verdict head are not merges"),
        "{}",
        d.detail
    );
    assert!(d.detail.contains(&n[..10]), "{}", d.detail);
    assert!(d.files.is_empty());
}

/// AC-1 (refused): the merge changed a PR line in an overlapping file.
#[test]
fn a_merge_that_changes_a_pr_line_is_not_a_mechanical_delta() {
    let r = delta_repo();
    let n = r.merge_tip_with(&[("src/a.rs", "ONE\n2\n3\n4\n5\n6\n7\neight\n")]);
    let d = r.delta(&n);
    assert!(!d.ok);
    assert!(
        d.detail.contains("src/a.rs: the PR's change differs"),
        "{}",
        d.detail
    );
}

/// AC-1 (refused): the merge changed a file outside the PR's file set (and outside the tip's).
#[test]
fn a_merge_that_changes_a_file_outside_the_pr_is_not_a_mechanical_delta() {
    let r = delta_repo();
    let n = r.merge_tip_with(&[("src/c.rs", "c-evil\n")]);
    let d = r.delta(&n);
    assert!(!d.ok);
    assert!(
        d.detail
            .contains("only on the current head: [\"src/c.rs\"]"),
        "{}",
        d.detail
    );
}

/// A merge that changes a PR binary file: its patch keeps the bytes (`--binary`), so the change
/// differs; without them both sides would read "Binary files differ" and look the same.
#[test]
fn a_merge_that_changes_a_pr_binary_file_is_not_a_mechanical_delta() {
    let r = delta_repo();
    commit(
        r.d(),
        &[("img.bin", "\u{0}\u{1}\u{2}bin-one\u{0}")],
        "image\n\nMoai-Claim: 9#555",
    );
    let h = r.head();
    let n = r.merge_tip_with(&[("img.bin", "\u{0}\u{1}\u{2}bin-two\u{0}")]);
    let d = gate_collect::mechanical_delta(r.d(), &h, &n, "main").unwrap();
    assert!(!d.ok);
    assert!(
        d.detail.contains("img.bin: the PR's change differs"),
        "{}",
        d.detail
    );
    // The same history with the binary left alone is a mechanical delta.
    let r = delta_repo();
    commit(
        r.d(),
        &[("img.bin", "\u{0}\u{1}\u{2}bin-one\u{0}")],
        "image\n\nMoai-Claim: 9#555",
    );
    let h = r.head();
    let n = r.merge_tip();
    let d = gate_collect::mechanical_delta(r.d(), &h, &n, "main").unwrap();
    assert!(d.ok, "{}", d.detail);
}

/// AC-2: a head the verdict head is not an ancestor of (the change re-made on the tip).
#[test]
fn a_head_that_does_not_descend_from_the_verdict_head_is_refused() {
    let r = delta_repo();
    git_ok(r.d(), &["checkout", "-q", "-b", "redo", "main"]);
    let n = commit(
        r.d(),
        &[
            ("src/a.rs", "one\n2\n3\n4\n5\n6\n7\neight\n"),
            ("src/p.rs", "p\n"),
        ],
        "redo\n\nMoai-Claim: 9#555",
    );
    let d = r.delta(&n);
    assert!(!d.ok);
    assert!(
        d.detail.contains("is not an ancestor of the current head"),
        "{}",
        d.detail
    );
    // An unknown verdict head is refused too, not an error.
    let d = gate_collect::mechanical_delta(r.d(), &"e".repeat(40), &n, "main").unwrap();
    assert!(
        !d.ok && d.detail.contains("not a local commit"),
        "{}",
        d.detail
    );
}

/// The file limit's boundary: as many files as the limit pass, one more is refused.
#[test]
fn the_delta_compares_at_most_its_file_limit() {
    let r = delta_repo();
    let n = r.merge_tip();
    let at = gate_collect::mechanical_delta_limited(r.d(), &r.h, &n, "main", 2).unwrap();
    assert!(at.ok, "{}", at.detail);
    let over = gate_collect::mechanical_delta_limited(r.d(), &r.h, &n, "main", 1).unwrap();
    assert!(!over.ok);
    assert!(
        over.detail.contains("2 PR files, more than the 1"),
        "{}",
        over.detail
    );
    assert_eq!(gate_collect::MAX_DELTA_FILES, 200);
}

/// End to end: `collect` measures the delta for a PASS on the verdict head, counts the tip's
/// changes from where the current head meets it (the overlap the merge resolved is not an
/// overlap any more), and the gate passes the current head.
#[test]
fn collect_carries_a_pass_across_an_update_branch_merge() {
    let r = delta_repo();
    let n = r.merge_tip();
    let target = Target {
        dir: r.d(),
        owner: "o",
        repo: "r",
        pr: 9,
        tip: "main",
        review_confirmed_prs: vec![],
    };
    let t = exchanges(&n, &r.c, vec![claim(&r.h), verdict(&r.h, &r.a)], PASS_BODY);
    let s = collect(&target, &t).unwrap();
    assert_eq!(t.remaining(), 0);
    let d = s.auto_delta.clone().expect("a delta is measured");
    assert!(d.ok, "{}", d.detail);
    assert!(
        s.tip_files_since_verdict_base.is_empty(),
        "{:?}",
        s.tip_files_since_verdict_base
    );
    assert!(gate::check(&s).is_ok(), "{:?}", gate::check(&s));
    assert!(gate::accepted_auto_delta(&s).is_some());
    // `gate-merge` and `gate-check` ask for the accepted delta directly: a FAIL on the verdict
    // head is never one.
    let mut failed = s.clone();
    for c in failed.comments.iter_mut() {
        c.body = c.body.replace("verdict: PASS", "verdict: FAIL");
    }
    assert!(gate::accepted_auto_delta(&failed).is_none());

    // The same head with a verdict on itself measures no delta.
    let t = exchanges(&n, &r.c, vec![claim(&n), verdict(&n, &r.a)], PASS_BODY);
    let s = collect(&target, &t).unwrap();
    assert!(s.auto_delta.is_none());
}

// ---- #2833 D4 / AC-3: the verdict first line ----

fn pr_head(head: &str) -> RecordedTransport {
    RecordedTransport::new(vec![Exchange {
        request: HttpRequest::new("GET", format!("{API}/pulls/9")),
        response: ok(json!({"head": {"sha": head}})),
    }])
}

fn delta_target(r: &Delta) -> Target<'_> {
    Target {
        dir: r.d(),
        owner: "o",
        repo: "r",
        pr: 9,
        tip: "main",
        review_confirmed_prs: vec![],
    }
}

/// AC-3: the line names the head the auditor examined, read back through the gate's own
/// parser, while that head is still the PR's head; the base is the merge base unless given.
#[test]
fn the_verdict_line_names_the_audited_head_and_reads_back() {
    let r = delta_repo();
    let target = delta_target(&r);
    let t = pr_head(&r.h);
    let v = gate_collect::verdict_line_for(&target, &r.h, "S3", "PASS", None, &t).unwrap();
    assert_eq!(t.remaining(), 0);
    let line = v["line"].as_str().unwrap();
    let (h, b, verdict) = gate::parse_verdict(line).expect("the gate reads the line");
    assert_eq!((h.as_str(), verdict.as_str()), (r.h.as_str(), "PASS"));
    assert_eq!(b, r.a, "the merge base of the PR head and the tip");
    assert_eq!(v["base_from"], json!("merge-base"));
    assert!(v.get("auto_delta").is_none());
    assert!(line.starts_with("AUDIT #9 (S3) head "), "{line}");

    // A given base (a delta re-audit names the earlier head) is used as it is.
    let v = gate_collect::verdict_line_for(&target, &r.h, "re", "FAIL", Some(&r.c), &pr_head(&r.h))
        .unwrap();
    let (_, b, verdict) = gate::parse_verdict(v["line"].as_str().unwrap()).unwrap();
    assert_eq!((b.as_str(), verdict.as_str()), (r.c.as_str(), "FAIL"));
    assert_eq!(v["base_from"], json!("given"));

    // `--head` must be a full commit id.
    let err =
        gate_collect::verdict_line_for(&target, &r.h[..10], "x", "PASS", None, &pr_head(&r.h))
            .unwrap_err();
    assert!(err.contains("--head must be the audited head"), "{err}");
    // Refused before any read: the PR is not even fetched.
    let t = pr_head(&r.h);
    let _ = gate_collect::verdict_line_for(&target, "abc", "x", "PASS", None, &t);
    assert_eq!(t.remaining(), 1, "no request before --head is checked");
}

/// Audit 6869947940 B1: a head that moved during the audit never gets a line for a head nobody
/// examined. An author commit (or any move that is not a mechanical delta) is refused with the
/// delta to audit; an update-branch merge of the tip gives the line for the audited head, which
/// the gate carries (AUTO-DELTA) — a PASS only.
#[test]
fn a_head_that_moved_during_the_audit_is_never_named_unaudited() {
    let r = delta_repo();
    let target = delta_target(&r);
    let n = r.merge_tip();
    let v =
        gate_collect::verdict_line_for(&target, &r.h, "S3", "PASS", None, &pr_head(&n)).unwrap();
    let (h, _, _) = gate::parse_verdict(v["line"].as_str().unwrap()).unwrap();
    assert_eq!(h, r.h, "the line names the audited head, not the merge");
    assert_eq!(v["current_head"], json!(n));
    assert_eq!(v["auto_delta"], json!(true));
    let v =
        gate_collect::verdict_line_for(&target, &r.h, "S3", "FAIL", None, &pr_head(&n)).unwrap();
    assert_eq!(v["auto_delta"], json!(false), "a FAIL is never carried");

    // The author added a commit after the audited head: refused, nothing printed.
    let m = commit(r.d(), &[("src/p.rs", "p2\n")], "more\n\nMoai-Claim: 9#555");
    let err = gate_collect::verdict_line_for(&target, &r.h, "S3", "PASS", None, &pr_head(&m))
        .unwrap_err();
    assert!(
        err.contains(&format!("now {m}, not the audited head {}", r.h)),
        "{err}"
    );
    assert!(err.contains("not a mechanical delta"), "{err}");
    assert!(
        err.contains(&format!("audit the delta {}..{m}", r.h)),
        "{err}"
    );

    // A new head that was never fetched cannot be checked: refused, not guessed.
    let missing = "f".repeat(40);
    let err = gate_collect::verdict_line_for(&target, &r.h, "x", "PASS", None, &pr_head(&missing))
        .unwrap_err();
    assert!(err.contains("fetch refs/pull/9/head"), "{err}");
    assert!(err.contains("no line printed"), "{err}");
}

#[test]
fn a_verdict_line_the_gate_would_not_read_is_refused() {
    let h = "a".repeat(40);
    let b = "b".repeat(40);
    assert!(gate::format_verdict_line(9, "t", &h, &b, "PASS").is_ok());
    for (title, head, base, verdict) in [
        ("t", "A".repeat(40), b.clone(), "PASS"),
        ("t", "a".repeat(39), b.clone(), "PASS"),
        ("t", h.clone(), "b".repeat(41), "PASS"),
        ("t", h.clone(), b.clone(), "Pass"),
        ("t", h.clone(), b.clone(), "PASS WITH NOTES"),
        ("two\nlines", h.clone(), b.clone(), "PASS"),
        ("  ", h.clone(), b.clone(), "PASS"),
    ] {
        assert!(
            gate::format_verdict_line(9, title, &head, &base, verdict).is_err(),
            "{title:?} {head} {base} {verdict}"
        );
    }
    // A title that itself says " head " still reads back: the gate takes the last one.
    let line = gate::format_verdict_line(9, "the head moved", &h, &b, "FAIL").unwrap();
    assert_eq!(gate::parse_verdict(&line), Some((h, b, "FAIL".to_string())));
}

/// AC-3 (document and gate agree): the example verdict line in the AUDIT workflow is exactly
/// what the helper prints for the same values.
#[test]
fn the_audit_workflow_example_is_the_helpers_line() {
    let text = fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/workflows/audit.md"),
    )
    .unwrap();
    let example = text
        .lines()
        .map(str::trim)
        .find(|l| l.starts_with("AUDIT #2786 "))
        .expect("the example line");
    let (h, b, v) = gate::parse_verdict(example).expect("the gate reads the example");
    let built = gate::format_verdict_line(2786, "#2708 4/4 repo-scans", &h, &b, &v).unwrap();
    assert_eq!(built, example);
    assert!(text.contains(
        "moai audit-verdict-line --pr <n> --head <audited 40-hex> --verdict PASS|FAIL --title <scope>"
    ));
    assert!(text.contains("a verdict never names a head nobody audited"));
}

/// #2860 rule 3: the collector resolves a short verdict base through git. A 12-digit prefix of
/// a real commit becomes its full SHA and the gate passes; a prefix git does not know leaves it
/// unresolved and G1 stops.
#[test]
fn a_short_verdict_base_is_resolved_to_one_commit() {
    let r = repo(&[("src/a.rs", "a2\n")], &[("src/b.rs", "b2\n")]);
    let t = exchanges(
        &r.b,
        &r.c,
        vec![claim(&r.b), verdict(&r.b, &r.a[..12])],
        PASS_BODY,
    );
    let s = collect(&target(&r), &t).unwrap();
    assert_eq!(s.verdict_base_full.as_deref(), Some(r.a.as_str()));
    assert!(s.verdict_base_known);
    assert!(gate::check(&s).is_ok(), "{:?}", gate::check(&s).err());

    let t = exchanges(
        &r.b,
        &r.c,
        vec![claim(&r.b), verdict(&r.b, "abababababab")],
        PASS_BODY,
    );
    let s = collect(&target(&r), &t).unwrap();
    assert_eq!(s.verdict_base_full, None);
    assert_eq!(gate::check(&s).unwrap_err().rule, "G1");
}

/// A repository whose base already has `src/lib.rs`; the PR and the tip each rewrite it.
fn repo_lib(pr_lib: &str, tip_lib: &str) -> Repo {
    let dir = TempDir::new().unwrap();
    let d = dir.path();
    git_ok(d, &["init", "-q", "-b", "main"]);
    let base = "pub mod a;\n\npub mod m;\n\npub mod z;\n";
    let a = commit(d, &[("src/lib.rs", base), ("src/a.rs", "a\n")], "base");
    git_ok(d, &["checkout", "-q", "-b", "pr"]);
    let b = commit(
        d,
        &[("src/lib.rs", pr_lib), ("src/new.rs", "n\n")],
        "pr\n\nMoai-Claim: 9#555",
    );
    git_ok(d, &["checkout", "-q", "main"]);
    let c = commit(d, &[("src/lib.rs", tip_lib)], "tip");
    Repo { _dir: dir, a, b, c }
}

/// #2860 rule 2 (shell gate_merge.sh): when the PR only adds module lines to `src/lib.rs` and
/// the merged tree keeps them, the tip's change to that file is no code overlap. Any other PR
/// change there (a non-module line, a removed line) stays an overlap and stops at G6.
#[test]
fn a_lib_rs_overlap_of_added_module_lines_is_allowed() {
    let tip = "pub mod a;\n\npub mod m;\n\npub mod z;\npub mod zz;\n";
    let check = |pr_lib: &str| {
        let r = repo_lib(pr_lib, tip);
        let t = exchanges(
            &r.b,
            &r.c,
            vec![claim(&r.b), verdict(&r.b, &r.a)],
            PASS_BODY,
        );
        let s = collect(&target(&r), &t).unwrap();
        assert!(s.merge_clean, "{pr_lib}");
        assert!(s
            .tip_files_since_verdict_base
            .contains(&"src/lib.rs".to_string()));
        (s.lib_mod_only_kept, gate::check(&s).err().map(|e| e.rule))
    };
    // Two added module lines (`pub mod` and `mod`), nothing else: allowed.
    assert_eq!(
        check("pub mod a;\npub mod b;\n\npub mod m;\nmod n;\n\npub mod z;\n"),
        (true, None)
    );
    // An added line that is not a module declaration.
    assert_eq!(
        check("pub mod a;\npub use a::x;\n\npub mod m;\n\npub mod z;\n"),
        (false, Some("G6"))
    );
    // A removed module line.
    assert_eq!(check("pub mod a;\n\n\npub mod z;\n"), (false, Some("G6")));
    // A module name with an upper-case letter is not the declaration form the rule allows.
    assert_eq!(
        check("pub mod a;\npub mod Bad;\n\npub mod m;\n\npub mod z;\n"),
        (false, Some("G6"))
    );
}

fn job(name: &str, conclusion: Option<&str>, id: u64) -> Value {
    json!({
        "name": name,
        "conclusion": conclusion,
        "details_url": format!("https://github.com/o/r/actions/runs/{id}/job/1"),
    })
}

/// #2860 (leader, #2856 N10; #2554 D8(a), leader tool green_tip.py): G8 judges the tip's jobs,
/// not the aggregate `ci`. `main` is a → c; c's CI still runs, so a decides.
/// - a's macOS job was cancelled by the newer push and the aggregate reads failure, but the
///   same macOS job succeeded on c: green (the aggregate alone would have read red).
/// - the same without a newer macOS success: a stays undecided and the walk ends unknown.
/// - a job that really failed: red.
/// - a cancelled job that is not macOS is never excused: unknown.
#[test]
fn g8_reads_the_jobs_and_excuses_a_macos_cancel_a_newer_run_passed() {
    let r = repo(&[("src/a.rs", "a2\n")], &[("src/b.rs", "b2\n")]);
    let linux = "ci / linux-x86_64";
    let mac = "ci / macos-arm64";
    for (tip_now, parent, want, decided) in [
        (
            json!([job(linux, None, 90), job(mac, Some("success"), 90)]),
            json!([
                job("ci", Some("failure"), 80),
                job(linux, Some("success"), 80),
                job(mac, Some("cancelled"), 80)
            ]),
            "green",
            Some(&r.a),
        ),
        (
            json!([job(linux, None, 90), job(mac, None, 90)]),
            json!([
                job("ci", Some("failure"), 80),
                job(linux, Some("success"), 80),
                job(mac, Some("cancelled"), 80)
            ]),
            "unknown",
            None,
        ),
        (
            json!([job(linux, None, 90), job(mac, Some("success"), 90)]),
            json!([
                job("ci", Some("failure"), 80),
                job(linux, Some("failure"), 80),
                job(mac, Some("cancelled"), 80)
            ]),
            "red",
            Some(&r.a),
        ),
        (
            json!([
                job(linux, None, 90),
                job("ci / windows-x86_64", Some("success"), 90)
            ]),
            json!([
                job("ci", Some("failure"), 80),
                job(linux, Some("success"), 80),
                job("ci / windows-x86_64", Some("cancelled"), 80)
            ]),
            "unknown",
            None,
        ),
    ] {
        let t = with_tip(
            &r.b,
            PASS_BODY,
            vec![
                tip_runs(&r.c, tip_now.clone()),
                tip_runs(&r.a, parent.clone()),
            ],
            &r.a,
        );
        let s = collect(&target(&r), &t).unwrap();
        assert_eq!(t.remaining(), 0, "{tip_now} / {parent}");
        let ci = s.tip_ci.clone().unwrap();
        assert_eq!(ci.state, want, "{tip_now} / {parent}");
        assert_eq!(ci.decided_by.as_ref(), decided, "{tip_now} / {parent}");
        assert_eq!(g8(&s).is_some(), want != "green", "{tip_now} / {parent}");
    }
}

/// D8(a)'s pure judgment, shared with `lane-next` (#2856): the aggregate and cancelled jobs are
/// not judged; a real job failure is red; nothing left is pending.
#[test]
fn the_d8a_judgment_ignores_the_aggregate_and_cancelled_jobs() {
    use moai_cli::gate_collect::d8a_judgment;
    let linux = "ci / linux-x86_64";
    let mac = "ci / macos-arm64";
    let green = [
        job("ci", Some("failure"), 1),
        job(linux, Some("success"), 1),
        job(mac, Some("cancelled"), 1),
    ];
    assert_eq!(d8a_judgment(&green)["state"], json!("green"));
    let red = [job(linux, Some("failure"), 1), job(mac, Some("success"), 1)];
    assert_eq!(d8a_judgment(&red)["state"], json!("red"));
    let nothing = [
        job("ci", Some("failure"), 1),
        job(mac, Some("cancelled"), 1),
    ];
    assert_eq!(d8a_judgment(&nothing)["state"], json!("pending"));
}
