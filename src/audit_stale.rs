//! `audit-claim-stale` (#3015 C7b, N2): audit reservations that stayed on closed or merged PRs.
//!
//! A reservation (`moai-audit-claim: lane=… head=…`) only matters on an open PR's current head.
//! When the PR is merged or closed without a verdict for that head (a merge outside the gate),
//! the line stays and nobody sees it: `lane-next` and `lane-status` read open PRs only, so the
//! lane that held it keeps an audit worktree and a task nothing asks for any more. This lists
//! those PRs from the recent closed ones. It writes nothing and invents no release line: a
//! reservation on a closed PR has no effect under the rule of [`crate::audit_claim`], and a prose
//! release is not read there.

use crate::audit_claim::{holder_of, reservations_on, same_lane};
use crate::forge::policy::{format_utc, parse_utc, Clock};
use crate::forge::transport::Transport;
use crate::forge::Client;
use crate::lane_status::{pages_with_evidence, read_comments, verdict_on, Comment};
use serde_json::{json, Value};

type Result<T> = std::result::Result<T, String>;

/// Closed pull requests one call looks at (one page, newest update first).
pub const PAGE: usize = 30;
/// Longest look-back in hours.
pub const MAX_SINCE_HOURS: i64 = 720;
/// Look-back when `--since-hours` is not given.
pub const DEFAULT_SINCE_HOURS: i64 = 48;

/// The row for one closed PR: the reservations on `head` that no verdict for `head` answers.
pub fn stale_row(
    pr: u64,
    merged: bool,
    closed_at: &str,
    head: &str,
    comments: &[Comment],
) -> Option<Value> {
    let reservations = reservations_on(comments, head);
    if reservations.is_empty() || verdict_on(comments, head).is_some() {
        return None;
    }
    let holder = holder_of(comments, head);
    Some(json!({
        "pr": pr,
        "merged": merged,
        "closed_at": closed_at,
        "head": head,
        "holder": holder.as_ref().map(|h| h.lane.clone()),
        "reservations": reservations.iter()
            .map(|(lane, id)| json!({"lane": lane, "comment_id": id}))
            .collect::<Vec<_>>(),
        "verdict_seen": false,
    }))
}

/// Lists the closed PRs of the last `since_hours` whose final head still carries a reservation
/// without a verdict. `lane` keeps only those its holder or a reserver matches by lane name.
pub fn list<T: Transport, C: Clock>(
    owner: &str,
    repo: &str,
    lane: Option<&str>,
    since_hours: i64,
    now: i64,
    transport: &T,
    clock: &C,
) -> Result<Value> {
    if !crate::lane_loop::repo_ok(&format!("{owner}/{repo}"))
        || [owner, repo].iter().any(|p| p.chars().all(|c| c == '.'))
    {
        return Err(format!("invalid repository `{owner}/{repo}`"));
    }
    if !(1..=MAX_SINCE_HOURS).contains(&since_hours) {
        return Err(format!(
            "invalid --since-hours {since_hours}: expected 1 to {MAX_SINCE_HOURS}"
        ));
    }
    let client = Client::new(transport, clock, Default::default(), Default::default());
    let url = format!(
        "https://api.github.com/repos/{owner}/{repo}/pulls?state=closed&sort=updated&direction=desc&per_page={PAGE}"
    );
    let resp = client
        .get(&url, None)
        .map_err(|e| format!("the closed pull requests could not be read: {}", e.message))?;
    let Some(Value::Array(list)) = resp.body else {
        return Err("the closed pull request list was not a list".into());
    };
    let since = now - since_hours * 3600;
    let mut partial = Vec::new();
    let (mut rows, mut unread) = (Vec::new(), Vec::new());
    // Control lines written by an account outside the organization do not count (#3052).
    let mut ignored = crate::comment_authority::Dropped::default();
    let mut looked = 0u64;
    for pr in &list {
        let (Some(n), Some(closed_at)) = (pr["number"].as_u64(), pr["closed_at"].as_str()) else {
            continue;
        };
        if parse_utc(closed_at).is_none_or(|t| t < since) {
            continue;
        }
        let Some(head) = pr["head"]["sha"].as_str() else {
            continue;
        };
        looked += 1;
        let url =
            format!("https://api.github.com/repos/{owner}/{repo}/issues/{n}/comments?per_page=100");
        match pages_with_evidence(&client, url, None) {
            Ok((raw, true)) => {
                let (comments, dropped) = read_comments(raw);
                ignored.merge(dropped);
                let Some(mut row) =
                    stale_row(n, !pr["merged_at"].is_null(), closed_at, head, &comments)
                else {
                    continue;
                };
                if let Some(want) = lane {
                    let hit = reservations_on(&comments, head)
                        .iter()
                        .any(|(l, _)| same_lane(l, want))
                        || row["holder"].as_str().is_some_and(|h| same_lane(h, want));
                    if !hit {
                        continue;
                    }
                }
                row["guidance"] = json!(
                    "the reservation has no effect on a closed pull request: no audit is owed, and the lane's audit worktree is for `worktree-gc-plan`"
                );
                rows.push(row);
            }
            _ => unread.push(n),
        }
    }
    // A full page whose last entry is still inside the window may hide older closed PRs.
    if list.len() >= PAGE {
        let last_in = list
            .last()
            .and_then(|p| p["closed_at"].as_str())
            .and_then(parse_utc)
            .is_some_and(|t| t >= since);
        if last_in {
            partial.push(format!(
                "the page of {PAGE} closed pull requests ends inside the last {since_hours} hours; older ones were not read"
            ));
        }
    }
    rows.sort_by_key(|r| std::cmp::Reverse(r["pr"].as_u64().unwrap_or(0)));
    let mut out = json!({
        "now": format_utc(now),
        "since": format_utc(since),
        "pull_requests_read": looked,
        "stale": rows,
        "unread": unread,
        "partial": partial,
    });
    if !ignored.is_empty() {
        out["ignored_control_lines"] = ignored.to_json();
    }
    Ok(out)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::forge::transport::{HttpRequest, HttpResponse};
    use std::cell::RefCell;
    use std::collections::BTreeMap;

    const NOW: i64 = 1_800_000_000;
    const H1: &str = "1111111111111111111111111111111111111111";
    const H2: &str = "2222222222222222222222222222222222222222";

    fn c(id: i64, body: &str) -> Comment {
        Comment {
            id,
            created_at: String::new(),
            body: body.into(),
        }
    }

    fn reserve(id: i64, lane: &str, head: &str) -> Comment {
        c(id, &format!("moai-audit-claim: lane={lane} head={head}"))
    }

    fn verdict(id: i64, pr: u64, head: &str, v: &str) -> Comment {
        c(
            id,
            &format!("AUDIT #{pr} (x) head {head} base {H2} — verdict: {v}"),
        )
    }

    #[test]
    fn a_reservation_without_a_verdict_for_its_head_is_stale() {
        let row = stale_row(
            7,
            true,
            "2026-11-15T01:00:00Z",
            H1,
            &[reserve(10, "agent-1", H1)],
        )
        .unwrap();
        assert_eq!(row["pr"], 7);
        assert_eq!(row["merged"], true);
        assert_eq!(row["holder"], "agent-1");
        assert_eq!(row["reservations"][0]["comment_id"], 10);
        assert_eq!(row["verdict_seen"], false);
    }

    #[test]
    fn a_verdict_for_the_head_a_missing_reservation_or_another_head_is_not() {
        let r = reserve(10, "agent-1", H1);
        for done in ["PASS", "FAIL"] {
            assert!(stale_row(7, false, "t", H1, &[r.clone(), verdict(11, 7, H1, done)]).is_none());
        }
        assert!(stale_row(7, false, "t", H1, &[c(1, "text")]).is_none());
        assert!(stale_row(7, false, "t", H1, &[reserve(10, "agent-1", H2)]).is_none());
        // A verdict on another head does not answer this head's reservation.
        assert!(stale_row(7, false, "t", H1, &[r, verdict(11, 7, H2, "PASS")]).is_some());
    }

    #[test]
    fn the_holder_follows_the_lowest_reservation_and_a_leader_handoff() {
        let both = [reserve(10, "agent-1", H1), reserve(12, "agent-2", H1)];
        let row = stale_row(7, false, "t", H1, &both).unwrap();
        assert_eq!(row["holder"], "agent-1");
        assert_eq!(row["reservations"].as_array().unwrap().len(), 2);
        let moved = [
            reserve(10, "agent-1", H1),
            c(13, "moai-claim-handoff: comment=10 to=agent-3 by=leader"),
        ];
        let row = stale_row(7, false, "t", H1, &moved).unwrap();
        assert_eq!(row["holder"], "agent-3");
    }

    struct Forge {
        pulls: Value,
        comments: BTreeMap<u64, Option<Value>>,
        seen: RefCell<Vec<String>>,
    }
    impl Transport for Forge {
        fn send(&self, req: &HttpRequest) -> std::result::Result<HttpResponse, String> {
            self.seen.borrow_mut().push(req.url.clone());
            if req.url.contains("/pulls?") {
                return Ok(HttpResponse::new(200).with_body(self.pulls.clone()));
            }
            let n: u64 = req
                .url
                .split("/issues/")
                .nth(1)
                .and_then(|r| r.split('/').next())
                .and_then(|n| n.parse().ok())
                .unwrap();
            match self.comments.get(&n) {
                Some(Some(b)) => Ok(HttpResponse::new(200).with_body(b.clone())),
                _ => Ok(HttpResponse::new(200).with_body(json!({"message": "boom"}))),
            }
        }
    }

    struct C0;
    impl Clock for C0 {
        fn now(&self) -> i64 {
            0
        }
        fn sleep(&self, _: u64) {}
    }

    fn pr(n: u64, closed_ago_h: i64, head: &str, merged: bool) -> Value {
        json!({
            "number": n,
            "closed_at": format_utc(NOW - closed_ago_h * 3600),
            "merged_at": if merged { json!("2026-11-15T00:00:00Z") } else { Value::Null },
            "head": {"sha": head},
        })
    }

    fn thread(cs: &[Comment]) -> Option<Value> {
        Some(Value::Array(
            cs.iter()
                .map(|c| json!({"id": c.id, "body": c.body, "author_association": "MEMBER"}))
                .collect(),
        ))
    }

    fn forge(pulls: Vec<Value>, comments: Vec<(u64, Option<Value>)>) -> Forge {
        Forge {
            pulls: Value::Array(pulls),
            comments: comments.into_iter().collect(),
            seen: Default::default(),
        }
    }

    #[test]
    fn only_closed_prs_inside_the_window_are_read_and_listed_newest_first() {
        let f = forge(
            // The list is not in PR-number order: the result is sorted, newest number first.
            vec![
                pr(20, 47, H1, false),
                pr(10, 49, H1, true),
                pr(30, 1, H1, true),
            ],
            vec![
                (30, thread(&[reserve(1, "agent-1", H1)])),
                (20, thread(&[reserve(2, "agent-2", H1)])),
                (10, thread(&[reserve(3, "agent-3", H1)])),
            ],
        );
        let v = list("o", "r", None, 48, NOW, &f, &C0).unwrap();
        let rows = v["stale"].as_array().unwrap();
        assert_eq!(rows.len(), 2, "{v}");
        assert_eq!(rows[0]["pr"], 30);
        assert_eq!(rows[1]["pr"], 20);
        // `merged` comes from the pull request's `merged_at`.
        assert_eq!(rows[0]["merged"], true, "{v}");
        assert_eq!(rows[1]["merged"], false, "{v}");
        assert_eq!(v["pull_requests_read"], 2);
        assert!(!f.seen.borrow().iter().any(|u| u.contains("/issues/10/")));
        assert!(rows[0]["guidance"].as_str().unwrap().contains("no effect"));
    }

    #[test]
    fn the_window_edge_is_inclusive() {
        for (ago, want) in [(47, 1), (48, 1), (49, 0)] {
            let f = forge(
                vec![pr(5, ago, H1, true)],
                vec![(5, thread(&[reserve(1, "agent-1", H1)]))],
            );
            let v = list("o", "r", None, 48, NOW, &f, &C0).unwrap();
            assert_eq!(v["stale"].as_array().unwrap().len(), want, "{ago} h");
        }
    }

    #[test]
    fn the_lane_filter_keeps_what_the_lane_reserved_or_holds() {
        let f = forge(
            vec![pr(30, 1, H1, true), pr(20, 2, H1, true)],
            vec![
                (30, thread(&[reserve(1, "agent-1", H1)])),
                (
                    20,
                    thread(&[
                        reserve(2, "agent-2", H1),
                        c(3, "moai-claim-handoff: comment=2 to=agent-1 by=leader"),
                    ]),
                ),
            ],
        );
        let mine = list("o", "r", Some("agent-1"), 48, NOW, &f, &C0).unwrap();
        assert_eq!(mine["stale"].as_array().unwrap().len(), 2);
        let other = list("o", "r", Some("agent-9"), 48, NOW, &f, &C0).unwrap();
        assert!(other["stale"].as_array().unwrap().is_empty());
        // A lane name with a restart tag is the same lane (the rule of audit-claim).
        let tagged = list("o", "r", Some("agent-1-h2"), 48, NOW, &f, &C0).unwrap();
        assert_eq!(tagged["stale"].as_array().unwrap().len(), 2);
        // ... also for a reserver that is not the holder: agent-2 reserved #20 and handed it
        // to agent-1, so only the reservation matches `agent-2-h2`.
        let reserver = list("o", "r", Some("agent-2-h2"), 48, NOW, &f, &C0).unwrap();
        let rows = reserver["stale"].as_array().unwrap();
        assert_eq!(rows.len(), 1, "{reserver}");
        assert_eq!(rows[0]["pr"], 20);
    }

    #[test]
    fn a_control_line_from_outside_the_organization_counts_for_nothing_and_is_reported() {
        let as_json = |id: i64, body: &str, who: &str| json!({"id": id, "body": body, "author_association": who});
        let reservation = format!("moai-audit-claim: lane=agent-1 head={H1}");
        let verdict_text = format!("AUDIT #30 (x) head {H1} base {H2} — verdict: PASS");
        // #30: a member reserved and an outsider posted a verdict: the verdict does not count.
        // #20: only an outsider reserved: there is no reservation at all.
        let f = forge(
            vec![pr(30, 1, H1, true), pr(20, 2, H1, true)],
            vec![
                (
                    30,
                    Some(json!([
                        as_json(1, &reservation, "MEMBER"),
                        as_json(2, &verdict_text, "NONE")
                    ])),
                ),
                (20, Some(json!([as_json(3, &reservation, "CONTRIBUTOR")]))),
            ],
        );
        let v = list("o", "r", None, 48, NOW, &f, &C0).unwrap();
        let rows = v["stale"].as_array().unwrap();
        assert_eq!(rows.len(), 1, "{v}");
        assert_eq!(rows[0]["pr"], 30);
        assert_eq!(rows[0]["verdict_seen"], false);
        let ignored = &v["ignored_control_lines"];
        assert_eq!(ignored["count"], 2, "{v}");
        let kinds: Vec<_> = ignored["items"]
            .as_array()
            .unwrap()
            .iter()
            .map(|i| {
                (
                    i["comment_id"].as_i64().unwrap(),
                    i["kind"].as_str().unwrap(),
                )
            })
            .collect();
        assert!(kinds.contains(&(2, "audit_verdict")), "{v}");
        assert!(kinds.contains(&(3, "audit_claim")), "{v}");
        // Nothing dropped: the key is absent.
        let clean = forge(
            vec![pr(30, 1, H1, true)],
            vec![(30, thread(&[reserve(1, "agent-1", H1)]))],
        );
        let v = list("o", "r", None, 48, NOW, &clean, &C0).unwrap();
        assert!(v.get("ignored_control_lines").is_none(), "{v}");
    }

    #[test]
    fn a_pull_request_that_cannot_be_read_is_named_never_taken_for_clean() {
        let f = forge(
            vec![pr(30, 1, H1, true), pr(20, 2, H1, true)],
            vec![(30, thread(&[reserve(1, "agent-1", H1)])), (20, None)],
        );
        let v = list("o", "r", None, 48, NOW, &f, &C0).unwrap();
        assert_eq!(v["unread"], json!([20]), "{v}");
        assert_eq!(v["pull_requests_read"], 2);
        assert_eq!(v["stale"].as_array().unwrap().len(), 1);
    }

    #[test]
    fn a_full_page_that_ends_inside_the_window_is_partial() {
        let page = |n: usize, last_ago: i64| -> Vec<Value> {
            (0..n)
                .map(|i| {
                    let ago = if i + 1 == n { last_ago } else { 1 };
                    pr(100 + i as u64, ago, H1, true)
                })
                .collect()
        };
        let quiet = |n: usize, last_ago: i64| {
            let comments = (0..n as u64).map(|i| (100 + i, thread(&[]))).collect();
            forge(page(n, last_ago), comments)
        };
        let v = list("o", "r", None, 48, NOW, &quiet(30, 10), &C0).unwrap();
        assert_eq!(v["partial"].as_array().unwrap().len(), 1, "{v}");
        let v = list("o", "r", None, 48, NOW, &quiet(30, 60), &C0).unwrap();
        assert!(v["partial"].as_array().unwrap().is_empty(), "{v}");
        let v = list("o", "r", None, 48, NOW, &quiet(29, 10), &C0).unwrap();
        assert!(v["partial"].as_array().unwrap().is_empty(), "{v}");
    }

    #[test]
    fn bad_input_is_refused_before_a_request() {
        let f = forge(vec![], vec![]);
        for (o, r) in [("o", ".."), ("..", "r"), ("o/x", "r"), ("o", "r?x")] {
            assert!(list(o, r, None, 48, NOW, &f, &C0).is_err(), "{o}/{r}");
        }
        for hours in [0, -1, 721] {
            assert!(
                list("o", "r", None, hours, NOW, &f, &C0).is_err(),
                "{hours}"
            );
        }
        assert!(list("o", "r", None, 720, NOW, &f, &C0).is_ok());
        assert!(list("o", "r", None, 1, NOW, &f, &C0).is_ok());
        let seen = f.seen.borrow();
        assert_eq!(seen.len(), 2, "only the two good calls ask: {seen:?}");
    }

    #[test]
    fn the_closed_list_is_one_get_for_the_newest_updates() {
        let f = forge(vec![], vec![]);
        list("o", "r", None, 48, NOW, &f, &C0).unwrap();
        assert_eq!(
            f.seen.borrow().as_slice(),
            ["https://api.github.com/repos/o/r/pulls?state=closed&sort=updated&direction=desc&per_page=30"]
        );
    }
}
