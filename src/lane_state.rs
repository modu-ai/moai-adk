//! The state of a lane, told from what is on record (#3015 C3b; user instruction 2026-11-15,
//! #3005 comment 6890270975).
//!
//! A quiet screen is not an idle lane. A lane whose cards wait for a check or an audit, a lane
//! that stopped with a resume note, and a lane that stopped with nothing saved all look the
//! same on a screen, and reading them as one "everything stopped" sends the wrong work to the
//! wrong place. This module sorts a lane into one of four states, or says it is working or
//! that it cannot tell, from five facts: the lane's open claims and their resume notes (the
//! durable hand-off note), the open pull requests those claims lead to, the work left unsaved in
//! the lane's worktrees, the background jobs a fresh screen observation shows (a running
//! writer), and whether the lane's session lease still runs. It reads facts it is given and
//! writes nothing.
//!
//! | `state`     | meaning                              | decided by                                  |
//! |-------------|--------------------------------------|---------------------------------------------|
//! | `working`   | a writer is running                  | background jobs > 0                         |
//! | `idle`      | nothing is owed                      | no open claim, nothing unsaved              |
//! | `waiting`   | done; waits for a check or an audit  | every claim has an open pull request, nothing unsaved |
//! | `resumable` | stopped, and can be taken up again   | a claim without a pull request (or unsaved work), each with a resume note |
//! | `stopped`   | stopped, nothing says where to go on | a claim without a pull request or resume note, or unsaved work with no claim |
//! | `unknown`   | a fact is missing                    | any fact that was not read; never guessed   |

use crate::forge::transport::Transport;
use crate::lane_reset::OpenClaim;
use serde_json::{json, Value};

/// An open pull request that one of the lane's open claims leads to.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct OpenPr {
    pub number: u64,
    /// The claimed Issue it belongs to.
    pub claim_issue: u64,
}

/// What is known about the lane; `None` is "not read" and is never taken as "none".
pub struct Facts<'a> {
    pub claims: Option<&'a [OpenClaim]>,
    /// The read of the claims was cut (the search found more threads than it returned): a claim
    /// that is not in `claims` may still be open.
    pub claims_partial: bool,
    pub open_prs: Option<&'a [OpenPr]>,
    /// The read of the open pull requests was cut (more than one page): a claim without a
    /// pull request in it may still have one.
    pub prs_partial: bool,
    /// Uncommitted, unpushed or stashed work in the worktrees that belong to the lane.
    pub unsaved: Option<bool>,
    /// Background jobs on the lane's screen, from an observation fresh enough to rely on.
    pub background_jobs: Option<u64>,
    /// The lane's session registry entry: `Some(true)` while its lease runs.
    pub session_live: Option<bool>,
}

/// The open pull request of a claimed Issue: one whose head branch is `issue/<n>-…` or whose
/// body carries the claim's own `Moai-Claim: <n>#<claim comment>` line.
pub fn pr_for_claim(claim: &OpenClaim, branch: &str, body: &str) -> bool {
    branch.starts_with(&format!("issue/{}-", claim.issue))
        || body
            .lines()
            .any(|l| l.trim() == format!("Moai-Claim: {}#{}", claim.issue, claim.claim_comment))
}

/// Most open pull requests the one page holds (GitHub's per-page limit); a full page may hide
/// more.
const PR_PAGE: usize = 100;

/// The open pull requests that lead to the lane's claims, and whether the read was cut.
#[derive(Clone, Debug, Default)]
pub struct PrsRead {
    pub prs: Vec<OpenPr>,
    pub partial: bool,
}

/// Reads the repository's open pull requests (GET, one page) and keeps those that lead to one
/// of `claims`. No claim, no request.
pub fn read_open_prs<T: Transport, C: crate::forge::policy::Clock>(
    owner: &str,
    repo: &str,
    claims: &[OpenClaim],
    transport: &T,
    clock: &C,
) -> Result<PrsRead, String> {
    if claims.is_empty() {
        return Ok(PrsRead::default());
    }
    if !crate::lane_loop::repo_ok(&format!("{owner}/{repo}"))
        || [owner, repo].iter().any(|p| p.chars().all(|c| c == '.'))
    {
        return Err(format!("invalid repository `{owner}/{repo}`"));
    }
    let client =
        crate::forge::Client::new(transport, clock, Default::default(), Default::default());
    let url =
        format!("https://api.github.com/repos/{owner}/{repo}/pulls?state=open&per_page={PR_PAGE}");
    let found = client
        .get(&url, None)
        .map_err(|e| format!("the open pull requests could not be read: {}", e.message))?;
    let body = found.body.unwrap_or(Value::Null);
    let items = body
        .as_array()
        .ok_or("the open pull requests were answered without a list")?;
    let mut prs = Vec::new();
    for item in items {
        let Some(number) = item["number"].as_u64() else {
            continue;
        };
        let branch = item["head"]["ref"].as_str().unwrap_or("");
        let text = item["body"].as_str().unwrap_or("");
        if let Some(claim) = claims.iter().find(|c| pr_for_claim(c, branch, text)) {
            prs.push(OpenPr {
                number,
                claim_issue: claim.issue,
            });
        }
    }
    Ok(PrsRead {
        prs,
        partial: items.len() >= PR_PAGE,
    })
}

/// The state, with the facts it rests on.
pub fn classify(f: &Facts) -> Value {
    let claims = f.claims.unwrap_or(&[]);
    let prs = f.open_prs.unwrap_or(&[]);
    let pr_of = |issue: u64| {
        prs.iter()
            .find(|p| p.claim_issue == issue)
            .map(|p| p.number)
    };
    let session = match f.session_live {
        Some(true) => "live",
        Some(false) => "ended",
        None => "unknown",
    };
    let evidence = json!({
        "open_claims": claims.iter().map(|c| json!({
            "issue": c.issue,
            "claim_comment": c.claim_comment,
            "resume_note": c.resume_note,
            "open_pr": pr_of(c.issue),
        })).collect::<Vec<_>>(),
        "unsaved_work": f.unsaved,
        "background_jobs": f.background_jobs,
        "session": session,
    });
    let answer = |state: &str, reasons: Vec<String>| json!({"state": state, "reasons": reasons, "evidence": evidence});

    // A running writer decides at once, whatever else is unread.
    if f.background_jobs.is_some_and(|n| n > 0) {
        return answer(
            "working",
            vec![format!(
                "{} background job(s) are running on the lane's screen",
                f.background_jobs.unwrap_or(0)
            )],
        );
    }

    let mut missing = Vec::new();
    if f.claims.is_none() {
        missing.push("the lane's open claims were not read".to_string());
    }
    if f.claims.is_some() && f.claims_partial {
        missing.push(
            "the lane's open claims were read only in part: a claim that is not listed may be open"
                .to_string(),
        );
    }
    if f.unsaved.is_none() {
        missing.push("the lane's worktrees could not be checked for unsaved work".to_string());
    }
    if f.background_jobs.is_none() {
        missing
            .push("no fresh screen observation: a running writer cannot be ruled out".to_string());
    }
    let needs_prs = claims.iter().any(|c| pr_of(c.issue).is_none());
    if f.claims.is_some() && !claims.is_empty() && f.open_prs.is_none() {
        missing.push("the open pull requests were not read".to_string());
    } else if f.prs_partial && needs_prs {
        missing.push(
            "the open pull requests were read only in part: a claim without one may have one"
                .to_string(),
        );
    }
    if !missing.is_empty() {
        return answer("unknown", missing);
    }

    let unsaved = f.unsaved == Some(true);
    if claims.is_empty() {
        return if unsaved {
            answer(
                "stopped",
                vec!["unsaved work in the lane's worktrees and no open claim to say where it belongs".into()],
            )
        } else {
            answer("idle", vec!["no open claim, no open pull request of a claim, nothing unsaved, no running writer".into()])
        };
    }
    // The claims that still need a hand: no pull request carries them, or work is unsaved.
    let open: Vec<&OpenClaim> = claims
        .iter()
        .filter(|c| unsaved || pr_of(c.issue).is_none())
        .collect();
    if open.is_empty() {
        return answer(
            "waiting",
            vec!["every open claim has an open pull request and nothing is unsaved".into()],
        );
    }
    let lacking: Vec<String> = open
        .iter()
        .filter(|c| c.resume_note.is_none())
        .map(|c| format!("#{}", c.issue))
        .collect();
    if lacking.is_empty() {
        answer(
            "resumable",
            vec!["every claim that still needs a hand has a resume note".into()],
        )
    } else {
        answer(
            "stopped",
            vec![format!(
                "claim(s) {} need a hand and have no resume note or pull request",
                lacking.join(", ")
            )],
        )
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn claim(issue: u64, note: Option<u64>) -> OpenClaim {
        OpenClaim {
            issue,
            claim_comment: issue * 10,
            resume_note: note,
        }
    }

    fn facts<'a>(claims: &'a [OpenClaim], prs: &'a [OpenPr]) -> Facts<'a> {
        Facts {
            claims: Some(claims),
            claims_partial: false,
            open_prs: Some(prs),
            prs_partial: false,
            unsaved: Some(false),
            background_jobs: Some(0),
            session_live: Some(true),
        }
    }

    fn state(f: &Facts) -> String {
        classify(f)["state"].as_str().unwrap().to_string()
    }

    fn pr(number: u64, issue: u64) -> OpenPr {
        OpenPr {
            number,
            claim_issue: issue,
        }
    }

    #[test]
    fn the_session_lease_is_shown_as_evidence_and_never_decides() {
        for (live, word) in [
            (Some(true), "live"),
            (Some(false), "ended"),
            (None, "unknown"),
        ] {
            let mut f = facts(&[], &[]);
            f.session_live = live;
            let v = classify(&f);
            assert_eq!(v["evidence"]["session"], word, "{v}");
            assert_eq!(
                v["state"], "idle",
                "the lease does not change the state: {v}"
            );
        }
    }

    #[test]
    fn nothing_owed_is_idle() {
        assert_eq!(state(&facts(&[], &[])), "idle");
    }

    #[test]
    fn every_claim_with_a_pull_request_and_nothing_unsaved_is_waiting() {
        let c = [claim(7, None), claim(8, Some(81))];
        let p = [pr(70, 7), pr(80, 8)];
        let v = classify(&facts(&c, &p));
        assert_eq!(v["state"], "waiting", "{v}");
        assert_eq!(v["evidence"]["open_claims"][0]["open_pr"], 70);
    }

    #[test]
    fn a_claim_without_a_pull_request_but_with_a_note_is_resumable() {
        let c = [claim(7, Some(71))];
        assert_eq!(state(&facts(&c, &[])), "resumable");
        // The session being over or live does not change it: the note is what resumes it.
        for live in [Some(false), Some(true), None] {
            let mut f = facts(&c, &[]);
            f.session_live = live;
            assert_eq!(state(&f), "resumable");
        }
    }

    #[test]
    fn a_claim_without_a_pull_request_or_a_note_is_stopped() {
        let c = [claim(7, None)];
        let v = classify(&facts(&c, &[]));
        assert_eq!(v["state"], "stopped", "{v}");
        assert!(v["reasons"][0].as_str().unwrap().contains("#7"), "{v}");
    }

    #[test]
    fn one_claim_waiting_and_one_without_a_note_is_stopped_not_waiting() {
        let c = [claim(7, None), claim(8, None)];
        let p = [pr(70, 7)];
        let v = classify(&facts(&c, &p));
        assert_eq!(v["state"], "stopped", "{v}");
        let reason = v["reasons"][0].as_str().unwrap();
        assert!(reason.contains("#8") && !reason.contains("#7"), "{v}");
    }

    #[test]
    fn unsaved_work_makes_a_waiting_lane_need_a_hand() {
        let c = [claim(7, Some(71))];
        let p = [pr(70, 7)];
        let mut f = facts(&c, &p);
        f.unsaved = Some(true);
        assert_eq!(state(&f), "resumable", "a note says where to go on");
        let c = [claim(7, None)];
        let mut f = facts(&c, &p);
        f.unsaved = Some(true);
        assert_eq!(state(&f), "stopped");
    }

    #[test]
    fn unsaved_work_with_no_claim_is_stopped() {
        let mut f = facts(&[], &[]);
        f.unsaved = Some(true);
        assert_eq!(state(&f), "stopped");
    }

    #[test]
    fn a_running_writer_is_working_whatever_else_is_unread() {
        let mut f = facts(&[], &[]);
        f.background_jobs = Some(1);
        f.claims = None;
        f.unsaved = None;
        assert_eq!(state(&f), "working");
        f.background_jobs = Some(0);
        assert_eq!(
            state(&f),
            "unknown",
            "with the jobs at 0 the missing facts show"
        );
    }

    #[test]
    fn a_fact_that_was_not_read_is_unknown_and_named() {
        let c = [claim(7, Some(71))];
        let p = [pr(70, 7)];
        type Break = fn(&mut Facts);
        let cases: [(&str, Break); 5] = [
            ("claims", |f| f.claims = None),
            ("unsaved", |f| f.unsaved = None),
            ("screen", |f| f.background_jobs = None),
            ("prs", |f| f.open_prs = None),
            ("partial", |f| {
                f.prs_partial = true;
                f.open_prs = Some(&[]);
            }),
        ];
        for (name, break_it) in cases {
            let mut f = facts(&c, &p);
            break_it(&mut f);
            let v = classify(&f);
            assert_eq!(v["state"], "unknown", "{name}: {v}");
            assert!(!v["reasons"].as_array().unwrap().is_empty(), "{name}");
        }
        // No claims at all needs no pull request read; and a cut read does not matter when
        // every claim already has its pull request.
        let mut f = facts(&[], &[]);
        f.open_prs = None;
        assert_eq!(state(&f), "idle");
        let mut f = facts(&c, &p);
        f.prs_partial = true;
        assert_eq!(state(&f), "waiting");
    }

    #[test]
    fn a_pull_request_belongs_to_a_claim_by_branch_or_by_its_claim_line() {
        let c = claim(7, None);
        assert!(pr_for_claim(&c, "issue/7-something-agent7", ""));
        assert!(!pr_for_claim(&c, "issue/70-something", ""));
        assert!(!pr_for_claim(&c, "issue/17-x", ""));
        assert!(pr_for_claim(&c, "other", "text\nMoai-Claim: 7#70\n"));
        assert!(!pr_for_claim(&c, "other", "Moai-Claim: 7#71"));
        assert!(!pr_for_claim(&c, "other", "Moai-Claim: 8#70"));
        // Only a line that is exactly the claim line counts: a longer id, a line that quotes it
        // in other text and text before it do not.
        assert!(pr_for_claim(&c, "other", "  Moai-Claim: 7#70  \n"));
        assert!(!pr_for_claim(&c, "other", "Moai-Claim: 7#700"));
        assert!(!pr_for_claim(
            &c,
            "other",
            "see Moai-Claim: 7#70 for the claim"
        ));
        assert!(!pr_for_claim(&c, "other", "> Moai-Claim: 7#70"));
        assert!(!pr_for_claim(&c, "other", "Moai-Claim: 7#70 (old)"));
    }

    #[test]
    fn claims_read_only_in_part_are_unknown_whatever_the_listed_ones_say() {
        // Every listed claim has a pull request, and none is listed at all: both would read as
        // `waiting` and `idle` if the cut were ignored.
        let claims = [claim(7, None)];
        let prs = [pr(41, 7)];
        let mut f = facts(&claims, &prs);
        assert_eq!(state(&f), "waiting");
        f.claims_partial = true;
        let v = classify(&f);
        assert_eq!(v["state"], "unknown");
        assert!(v["reasons"][0]
            .as_str()
            .unwrap()
            .contains("read only in part"));
        let mut none = facts(&[], &[]);
        assert_eq!(state(&none), "idle");
        none.claims_partial = true;
        assert_eq!(state(&none), "unknown");
        // Claims that were not read at all stay `unknown` without the partial reason twice.
        let mut unread = facts(&[], &[]);
        unread.claims = None;
        unread.claims_partial = true;
        let v = classify(&unread);
        assert_eq!(v["state"], "unknown");
        assert_eq!(v["reasons"].as_array().unwrap().len(), 1);
    }

    struct Stub(Vec<Value>, std::cell::RefCell<Vec<String>>);
    impl Transport for Stub {
        fn send(
            &self,
            req: &crate::forge::transport::HttpRequest,
        ) -> Result<crate::forge::transport::HttpResponse, String> {
            self.1
                .borrow_mut()
                .push(format!("{} {}", req.method, req.url));
            Ok(crate::forge::transport::HttpResponse::new(200)
                .with_body(Value::Array(self.0.clone())))
        }
    }

    struct Clock0;
    impl crate::forge::policy::Clock for Clock0 {
        fn now(&self) -> i64 {
            0
        }
        fn sleep(&self, _: u64) {}
    }

    #[test]
    fn open_pull_requests_are_read_with_one_get_and_matched_to_claims() {
        let claims = [claim(7, None), claim(8, None)];
        let list = vec![
            json!({"number": 70, "head": {"ref": "issue/7-x"}, "body": ""}),
            json!({"number": 80, "head": {"ref": "other"}, "body": "Moai-Claim: 8#80"}),
            json!({"number": 90, "head": {"ref": "issue/9-y"}, "body": "Moai-Claim: 9#90"}),
        ];
        let stub = Stub(list, Default::default());
        let read = read_open_prs("o", "r", &claims, &stub, &Clock0).unwrap();
        assert_eq!(read.prs, vec![pr(70, 7), pr(80, 8)]);
        assert!(!read.partial);
        assert_eq!(
            stub.1.borrow().as_slice(),
            ["GET https://api.github.com/repos/o/r/pulls?state=open&per_page=100"]
        );
    }

    #[test]
    fn no_claim_asks_nothing_and_a_full_page_is_partial() {
        let stub = Stub(vec![], Default::default());
        let read = read_open_prs("o", "r", &[], &stub, &Clock0).unwrap();
        assert!(read.prs.is_empty() && stub.1.borrow().is_empty());
        let claims = [claim(7, None)];
        let page = |n: usize| -> Vec<Value> {
            (0..n)
                .map(|i| json!({"number": 1000 + i, "head": {"ref": "x"}, "body": ""}))
                .collect()
        };
        let full = Stub(page(100), Default::default());
        assert!(
            read_open_prs("o", "r", &claims, &full, &Clock0)
                .unwrap()
                .partial
        );
        let short = Stub(page(99), Default::default());
        assert!(
            !read_open_prs("o", "r", &claims, &short, &Clock0)
                .unwrap()
                .partial
        );
    }

    #[test]
    fn a_repository_that_would_change_the_url_is_refused_before_a_request() {
        let claims = [claim(7, None)];
        for (o, r) in [("o", ".."), ("..", "r"), ("o/x", "r"), ("o", "r?x")] {
            let stub = Stub(vec![], Default::default());
            assert!(
                read_open_prs(o, r, &claims, &stub, &Clock0).is_err(),
                "{o}/{r}"
            );
            assert!(stub.1.borrow().is_empty(), "{o}/{r}");
        }
    }
}
