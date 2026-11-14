//! #2859 (leader decisions D1–D4): a pull request stacked on another open one does not bring
//! that PR's unaudited head into the tip. `gate_collect::collect` reads the open PRs below
//! (GitHub through a recorded transport, GET only; git from a temporary repository) and G12
//! stops unless the PR below has a PASS that holds every commit this PR brings in, by an
//! auditor who wrote none of them. A PR below that already merged is part of the tip (D2).

mod common;

use common::git::{git, git_ok};
use moai_cli::forge::transport::{Exchange, HttpRequest, HttpResponse, RecordedTransport};
use moai_cli::gate::{self, Snapshot};
use moai_cli::gate_collect::{collect, Target};
use serde_json::{json, Value};
use std::fs;
use std::path::Path;
use tempfile::TempDir;

const API: &str = "https://api.github.com/repos/o/r";
/// A PR body the Self-Audit checker passes (the four measured lines).
const PASS_BODY: &str = "Refs #2859\nMoai-Claim: 9#555\n\n## Self-Audit\n\n\
- mutation: claim-verify guard mutated (claim filter removed) -> check FAILED, measured\n\
- clone: check set from a fresh clone, not the /tmp symlink path\n\
- host: api.github.com exercised via fake gh executable, no network\n\
- receipt: receipt digest re-read and matched before PASS recorded\n\
- plan: PLAN comment 6872621175\n\
- red: gate_check failed first without the plan line\n";

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

/// Base on `main`; PR #7 (branch `a`) adds `a1`, or `a1` then `a2` when `moved`; PR #9 (branch
/// `b`) adds `b1` on top of PR #7's last commit. With `merged`, `main` has merged PR #7.
struct Stack {
    dir: TempDir,
    base: String,
    a1: String,
    a2: Option<String>,
    b1: String,
    tip: String,
}

fn stack(moved: bool, merged: bool) -> Stack {
    let dir = TempDir::new().unwrap();
    let d = dir.path();
    git_ok(d, &["init", "-q", "-b", "main"]);
    let base = commit(d, &[("src/base.rs", "base\n")], "base");
    git_ok(d, &["checkout", "-q", "-b", "a"]);
    let a1 = commit(d, &[("src/a.rs", "a1\n")], "a1\n\nMoai-Claim: 7#777");
    let a2 = moved.then(|| commit(d, &[("src/a.rs", "a2\n")], "a2\n\nMoai-Claim: 7#777"));
    git_ok(d, &["checkout", "-q", "-b", "b"]);
    let b1 = commit(d, &[("src/b.rs", "b1\n")], "b1\n\nMoai-Claim: 9#555");
    git_ok(d, &["checkout", "-q", "main"]);
    if merged {
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
                "Merge PR 7",
                "a",
            ],
        );
    }
    let tip = git(d, &["rev-parse", "main"]).trim().to_string();
    Stack {
        dir,
        base,
        a1,
        a2,
        b1,
        tip,
    }
}

fn ok(body: Value) -> HttpResponse {
    HttpResponse::new(200).with_body(body)
}

fn get(url: String, body: Value) -> Exchange {
    Exchange {
        request: HttpRequest::new("GET", url),
        response: ok(body),
    }
}

fn claim_comment(id: u64, session: &str) -> Exchange {
    get(
        format!("{API}/issues/comments/{id}"),
        json!({"id": id, "body": format!("moai-claim: session={session} epoch=1\n\nmore")}),
    )
}

fn verdict(pr: u64, id: u64, head: &str, base: &str, v: &str, lane: &str) -> Value {
    json!({"id": id, "body": format!("AUDIT #{pr} (x) head {head} base {base} — verdict: {v}\n\n감사: {lane} (예약 {}).\n", id - 1)})
}

/// The `local-check:` pass line for `head` (#2906 S5b, G5).
fn local_check(head: &str) -> Value {
    json!({"id": 250, "body": format!("local-check: head={head} result=pass cmd=tools/local_check.sh at=2026-11-14T12:00:00Z\n")})
}

/// The reads `collect` makes for PR #9 with PR #7 open below it: PR #9 (passing on its own),
/// the open list, PR #7's commits, the claim its carried commits name, PR #7's comments, the
/// green tip. `pr7_comments` is PR #7's verdict history.
fn reads(s: &Stack, pr7_open: bool, pr7_comments: Value) -> RecordedTransport {
    let mut ex = vec![
        get(
            format!("{API}/pulls/9"),
            json!({"head": {"sha": s.b1}, "body": PASS_BODY}),
        ),
        get(
            format!("{API}/issues/9/comments?per_page=100"),
            json!([
                {"id": 100, "body": format!("moai-audit-claim: lane=agent-10 head={}\n", s.b1)},
                verdict(9, 200, &s.b1, &s.base, "PASS", "agent-10"),
                local_check(&s.b1),
            ]),
        ),
        get(
            format!("{API}/commits/{}/check-runs?per_page=100", s.b1),
            json!({"check_runs": [{"name": "ci", "conclusion": "success"}]}),
        ),
        claim_comment(555, "agent-26"),
    ];
    let carried = !pr7_open || s.tip == s.base;
    if s.tip == s.base {
        // PR #7's commits are not on the tip: PR #9's own trailers name its claim too.
        ex.push(claim_comment(777, "agent-31"));
    }
    let open = if pr7_open {
        json!([{"number": 9}, {"number": 7}])
    } else {
        json!([{"number": 9}])
    };
    ex.push(get(format!("{API}/pulls?state=open&per_page=100"), open));
    if pr7_open {
        let mut commits = vec![json!({"sha": s.a1})];
        if let Some(a2) = &s.a2 {
            commits.push(json!({"sha": a2}));
        }
        ex.push(get(
            format!("{API}/pulls/7/commits?per_page=100"),
            Value::Array(commits),
        ));
        if carried {
            ex.push(claim_comment(777, "agent-31"));
            ex.push(get(
                format!("{API}/issues/7/comments?per_page=100"),
                pr7_comments,
            ));
        }
    }
    ex.push(get(
        format!("{API}/commits/{}/check-runs?filter=all&per_page=100", s.tip),
        json!({"check_runs": [{"name": "ci", "conclusion": "success"}]}),
    ));
    RecordedTransport::new(ex)
}

fn target(s: &Stack) -> Target<'_> {
    Target {
        dir: s.dir.path(),
        owner: "o",
        repo: "r",
        pr: 9,
        tip: "main",
        review_confirmed_prs: Vec::new(),
    }
}

fn snapshot(s: &Stack, pr7_open: bool, pr7_comments: Value) -> Snapshot {
    let t = reads(s, pr7_open, pr7_comments);
    let snap = collect(&target(s), &t).unwrap();
    assert_eq!(t.remaining(), 0, "every recorded read was made");
    snap
}

fn g12_stop(snap: &Snapshot) -> String {
    let stop = gate::check(snap).unwrap_err();
    assert_eq!(stop.rule, "G12", "{stop:?}");
    stop.reason
}

/// AC-1: PR #9 stacked on PR #7 whose head has no PASS stops at G12, naming PR #7 and the
/// commit it would bring in (AC-2: the reason the leader loop prints).
#[test]
fn a_stacked_pr_whose_pr_below_has_no_pass_stops() {
    let s = stack(false, false);
    let snap = snapshot(&s, true, json!([]));
    let below = snap.stacked_below.as_ref().unwrap();
    assert_eq!(below.len(), 1);
    assert_eq!(below[0].pr, 7);
    assert_eq!(below[0].brought, vec![s.a1.clone()]);
    assert_eq!(below[0].authors, vec!["agent-31".to_string()]);
    let why = g12_stop(&snap);
    assert!(why.contains("PR #7 below carries"), "{why}");
    assert!(why.contains(&s.a1[..10]), "{why}");
    assert!(why.contains("no AUDIT verdict"), "{why}");

    // A FAIL is not a PASS either.
    let snap = snapshot(
        &s,
        true,
        json!([verdict(7, 301, &s.a1, &s.base, "FAIL", "agent-30")]),
    );
    assert!(g12_stop(&snap).contains("is FAIL"));
}

/// AC-1: a PASS on the head PR #9 carries, by an auditor who wrote none of it, passes.
#[test]
fn a_pass_on_the_carried_head_passes() {
    let s = stack(false, false);
    let snap = snapshot(
        &s,
        true,
        json!([verdict(7, 301, &s.a1, &s.base, "PASS", "agent-30")]),
    );
    let below = snap.stacked_below.as_ref().unwrap();
    assert!(below[0].covered, "{below:?}");
    assert!(gate::check(&snap).is_ok(), "{:?}", gate::check(&snap));
}

/// AC-1 / D2: a PR below that already merged is part of the tip — not open and nothing of it
/// brought in — so PR #9 passes without its verdict being read.
#[test]
fn a_pr_below_that_merged_passes() {
    let s = stack(false, true);
    let snap = snapshot(&s, false, json!([]));
    assert!(snap.stacked_below.as_ref().unwrap().is_empty());
    assert!(gate::check(&snap).is_ok(), "{:?}", gate::check(&snap));

    // Still listed open (GitHub lags), its commits are on the tip: nothing is brought in.
    let snap = snapshot(&s, true, json!([]));
    assert!(snap.stacked_below.as_ref().unwrap().is_empty());
}

/// AC-1: PR #7 has a PASS only on its earlier head `a1`, and PR #9 carries its newer head `a2`
/// — the PASS does not hold `a2`, so PR #9 stops. A PASS on `a2` (newer than everything
/// carried) passes.
#[test]
fn a_pass_on_an_older_head_of_the_pr_below_does_not_cover_a_newer_one() {
    let s = stack(true, false);
    let a2 = s.a2.clone().unwrap();
    let snap = snapshot(
        &s,
        true,
        json!([verdict(7, 301, &s.a1, &s.base, "PASS", "agent-30")]),
    );
    let below = snap.stacked_below.as_ref().unwrap();
    assert_eq!(
        below[0].brought,
        vec![a2.clone()],
        "only the newest carried commit"
    );
    assert!(!below[0].covered);
    let why = g12_stop(&snap);
    assert!(
        why.contains("does not hold the commits this PR brings in"),
        "{why}"
    );
    assert!(why.contains(&a2[..10]), "{why}");

    let snap = snapshot(
        &s,
        true,
        json!([verdict(7, 301, &a2, &s.base, "PASS", "agent-30")]),
    );
    assert!(gate::check(&snap).is_ok(), "{:?}", gate::check(&snap));
}

/// D1: the PR below's PASS follows G9 — a PASS by a lane that wrote the carried commits stops.
#[test]
fn a_pass_below_by_its_own_author_stops() {
    let s = stack(false, false);
    let snap = snapshot(
        &s,
        true,
        json!([verdict(7, 301, &s.a1, &s.base, "PASS", "agent-31")]),
    );
    let why = g12_stop(&snap);
    assert!(
        why.contains("its PASS is by agent-31, and agent-31 wrote commits of it"),
        "{why}"
    );
}

/// A PASS on a head the gate's repository does not have cannot be shown to hold the carried
/// commits: unknown is never a pass.
#[test]
fn a_pass_on_a_head_that_is_not_local_does_not_cover() {
    let s = stack(false, false);
    let missing = "e".repeat(40);
    let snap = snapshot(
        &s,
        true,
        json!([verdict(7, 301, &missing, &s.base, "PASS", "agent-30")]),
    );
    assert!(!snap.stacked_below.as_ref().unwrap()[0].covered);
    assert!(g12_stop(&snap).contains("does not hold"));
}

/// A snapshot without the field (an older collector, or one typed by hand) stops at G12.
#[test]
fn a_snapshot_that_did_not_read_the_prs_below_stops() {
    let s = stack(false, false);
    let mut snap = snapshot(
        &s,
        true,
        json!([verdict(7, 301, &s.a1, &s.base, "PASS", "agent-30")]),
    );
    snap.stacked_below = None;
    assert!(g12_stop(&snap).contains("were not read"));
}

/// The PR below's commit list is read over at most three pages (GitHub lists 250 commits of a
/// pull request): three pages, the last without `next`, are all read and a carried commit on
/// the third is found; a third page that points to a fourth stops before the fourth request.
#[test]
fn the_pr_below_commit_pages_stop_at_three() {
    let s = stack(false, false);
    let first = format!("{API}/pulls/7/commits?per_page=100");
    let page_url = |n: usize| {
        if n == 1 {
            first.clone()
        } else {
            format!("{API}/pr7-commit-pages/{n}")
        }
    };
    let pages = |more: bool| -> Vec<Exchange> {
        (1..=3)
            .map(|n| {
                let shas = if n == 3 {
                    json!([{"sha": s.a1}])
                } else {
                    json!([{"sha": format!("{n}").repeat(40)}])
                };
                let mut resp = ok(shas);
                if n < 3 || more {
                    resp = resp.with_header("link", format!("<{}>; rel=\"next\"", page_url(n + 1)));
                }
                Exchange {
                    request: HttpRequest::new("GET", page_url(n)),
                    response: resp,
                }
            })
            .collect()
    };
    let head_reads = || -> Vec<Exchange> {
        vec![
            get(
                format!("{API}/pulls/9"),
                json!({"head": {"sha": s.b1}, "body": PASS_BODY}),
            ),
            get(
                format!("{API}/issues/9/comments?per_page=100"),
                json!([
                    {"id": 100, "body": format!("moai-audit-claim: lane=agent-10 head={}\n", s.b1)},
                    verdict(9, 200, &s.b1, &s.base, "PASS", "agent-10"),
                    local_check(&s.b1),
                ]),
            ),
            get(
                format!("{API}/commits/{}/check-runs?per_page=100", s.b1),
                json!({"check_runs": [{"name": "ci", "conclusion": "success"}]}),
            ),
            claim_comment(555, "agent-26"),
            claim_comment(777, "agent-31"),
            get(
                format!("{API}/pulls?state=open&per_page=100"),
                json!([{"number": 9}, {"number": 7}]),
            ),
        ]
    };

    // N = 3 pages: read, and the carried commit on the last page is found.
    let mut ex = head_reads();
    ex.extend(pages(false));
    ex.push(claim_comment(777, "agent-31"));
    ex.push(get(
        format!("{API}/issues/7/comments?per_page=100"),
        json!([verdict(7, 301, &s.a1, &s.base, "PASS", "agent-30")]),
    ));
    ex.push(get(
        format!("{API}/commits/{}/check-runs?filter=all&per_page=100", s.tip),
        json!({"check_runs": [{"name": "ci", "conclusion": "success"}]}),
    ));
    let t = RecordedTransport::new(ex);
    let snap = collect(&target(&s), &t).unwrap();
    assert_eq!(t.remaining(), 0);
    assert_eq!(
        snap.stacked_below.as_ref().unwrap()[0].brought,
        vec![s.a1.clone()]
    );

    // N + 1: the third page points to a fourth; nothing past the third is requested.
    let mut ex = head_reads();
    ex.extend(pages(true));
    let t = RecordedTransport::new(ex);
    let err = collect(&target(&s), &t).unwrap_err();
    assert!(err.contains("more than 3 pages"), "{err}");
    assert_eq!(t.remaining(), 0, "the fourth page is never requested");
}
