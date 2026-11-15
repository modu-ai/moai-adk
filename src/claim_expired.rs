//! `claim-expired` (#3015 C7a, N19): the open claims whose lease has ended, in one read.
//!
//! A leader that wants to know which lane has gone quiet had to name the Issues one by one. This
//! searches the open threads that carry a `moai-claim:` line, reads each thread's comments to
//! the end and asks [`crate::claim_verify::evaluate`] (the rule of A9) who holds the Issue. A
//! holder whose lease ended (`lease_until <= now`, the end second itself is over) is listed.
//! Nothing is written and nothing is taken over: a takeover still needs the three conditions of
//! `claim-timeout-check` and a handoff line.

use crate::claim_verify::{classify_comment, evaluate, CommentKind, IssueComment, Verifier};
use crate::forge::policy::Clock;
use crate::forge::transport::Transport;
use crate::forge::Client;
use serde_json::{json, Value};
use std::collections::BTreeSet;

type Result<T> = std::result::Result<T, String>;

/// Threads one search page holds (GitHub's per-page limit).
const SEARCH_PAGE: u64 = 100;
/// Most threads whose comments one call reads; more are named in `partial`.
pub const MAX_THREADS: usize = 60;

/// One expired claim.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Expired {
    pub issue: u64,
    pub claim_comment: i64,
    pub session: String,
    pub epoch: u64,
    pub lease_until: i64,
}

/// The expired claims of one thread: the holder the A9 rule names, when its lease has ended.
pub fn expired_in(issue: u64, comments: &[IssueComment], now: i64) -> Option<Expired> {
    let verdict = evaluate(
        comments,
        &Verifier {
            session: "",
            comment_id: None,
            allowed_authors: &[],
            now,
        },
    );
    let holder = verdict["holder"].as_i64()?;
    let CommentKind::Claim {
        session,
        epoch,
        lease_until,
    } = comments
        .iter()
        .find(|c| c.id == holder)
        .map(|c| classify_comment(&c.body))?
    else {
        return None;
    };
    // `holder_live` is the verdict's own reading of `lease_until > now`.
    let lease_until = crate::forge::policy::parse_utc(&lease_until)?;
    if verdict["holder_live"].as_bool() != Some(false) || lease_until > now {
        return None;
    }
    Some(Expired {
        issue,
        claim_comment: holder,
        session,
        epoch,
        lease_until,
    })
}

/// Reads the open threads with a claim line and lists the expired holders, longest-expired
/// first. `unread` names the threads whose comments could not be read to the end: they are
/// never taken for "no expired claim".
pub fn list<T: Transport, C: Clock>(
    owner: &str,
    repo: &str,
    now: i64,
    transport: &T,
    clock: &C,
) -> Result<Value> {
    if !crate::lane_loop::repo_ok(&format!("{owner}/{repo}"))
        || [owner, repo].iter().any(|p| p.chars().all(|c| c == '.'))
    {
        return Err(format!("invalid repository `{owner}/{repo}`"));
    }
    let client = Client::new(transport, clock, Default::default(), Default::default());
    let query = format!("repo:{owner}/{repo} is:open in:comments \"moai-claim:\"");
    let url = format!(
        "https://api.github.com/search/issues?q={}&per_page={SEARCH_PAGE}",
        crate::lane_reset::encode_query(&query)
    );
    let found = client
        .get(&url, None)
        .map_err(|e| format!("the search for claims failed: {}", e.message))?;
    let body = found.body.unwrap_or(Value::Null);
    let items = body["items"]
        .as_array()
        .ok_or("the search for claims answered without a list of items")?;
    let shown = items.len() as u64;
    let mut partial = Vec::new();
    match body["total_count"].as_u64() {
        Some(total) if total > shown => partial.push(format!(
            "the search found {total} threads and returned {shown}"
        )),
        None if shown >= SEARCH_PAGE => partial.push(format!(
            "the search returned a full page of {shown} threads and did not say how many it found"
        )),
        _ => {}
    }
    if body["incomplete_results"].as_bool() == Some(true) {
        partial.push("GitHub's search flagged its result as incomplete".to_string());
    }
    let numbers: BTreeSet<u64> = items.iter().filter_map(|i| i["number"].as_u64()).collect();
    if numbers.len() > MAX_THREADS {
        partial.push(format!(
            "{} threads have a claim line; the first {MAX_THREADS} were read",
            numbers.len()
        ));
    }
    let (mut expired, mut unread) = (Vec::new(), Vec::new());
    let mut read = 0u64;
    for n in numbers.into_iter().take(MAX_THREADS) {
        let url =
            format!("https://api.github.com/repos/{owner}/{repo}/issues/{n}/comments?per_page=100");
        match crate::lane_status::pages_with_evidence(&client, url, None) {
            Ok((raw, true)) => {
                read += 1;
                let comments: Vec<IssueComment> = raw
                    .iter()
                    .filter_map(|c| {
                        Some(IssueComment {
                            id: c["id"].as_i64()?,
                            author: c["user"]["login"].as_str().unwrap_or("").to_string(),
                            body: c["body"].as_str().unwrap_or("").to_string(),
                        })
                    })
                    .collect();
                expired.extend(expired_in(n, &comments, now));
            }
            _ => unread.push(n),
        }
    }
    expired.sort_by_key(|e| (e.lease_until, e.issue));
    Ok(json!({
        "now": crate::forge::policy::format_utc(now),
        "threads_read": read,
        "expired": expired.iter().map(|e| json!({
            "issue": e.issue,
            "claim_comment": e.claim_comment,
            "session": e.session,
            "epoch": e.epoch,
            "lease_until": crate::forge::policy::format_utc(e.lease_until),
            "expired_secs": now - e.lease_until,
        })).collect::<Vec<_>>(),
        "unread": unread,
        "partial": partial,
        "takeover": "not done here: it needs `claim-timeout-check` and a handoff line",
    }))
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::forge::transport::{HttpRequest, HttpResponse};
    use std::cell::RefCell;
    use std::collections::BTreeMap;

    const NOW: i64 = 1_800_000_000;

    fn at(secs: i64) -> String {
        crate::forge::policy::format_utc(secs)
    }

    fn claim(id: i64, session: &str, epoch: u64, lease: i64) -> IssueComment {
        IssueComment {
            id,
            author: "GoosLab".into(),
            body: format!(
                "moai-claim: session={session} epoch={epoch} lease_until={}",
                at(lease)
            ),
        }
    }

    fn text(id: i64, body: &str) -> IssueComment {
        IssueComment {
            id,
            author: "GoosLab".into(),
            body: body.into(),
        }
    }

    #[test]
    fn a_holder_is_expired_from_the_second_its_lease_ends() {
        for (lease, want) in [(NOW + 1, false), (NOW, true), (NOW - 1, true)] {
            let found = expired_in(7, &[claim(10, "agent-1", 1, lease)], NOW);
            assert_eq!(found.is_some(), want, "lease {lease}");
        }
        let e = expired_in(7, &[claim(10, "agent-1", 1, NOW - 50)], NOW).unwrap();
        assert_eq!(
            e,
            Expired {
                issue: 7,
                claim_comment: 10,
                session: "agent-1".into(),
                epoch: 1,
                lease_until: NOW - 50
            }
        );
    }

    #[test]
    fn a_released_or_handed_off_claim_is_not_listed_and_a_renewed_one_is_judged_by_its_new_lease() {
        let old = claim(10, "agent-1", 1, NOW - 500);
        let released = text(11, "moai-claim-release: comment=10");
        assert!(expired_in(7, &[old.clone(), released], NOW).is_none());
        let handoff = text(11, "moai-claim-handoff: comment=10 to=agent-2 by=holder");
        assert!(expired_in(7, &[old.clone(), handoff], NOW).is_none());
        // A later epoch of the same session replaces the ended one.
        let fresh = claim(12, "agent-1", 2, NOW + 600);
        assert!(expired_in(7, &[old, fresh], NOW).is_none());
    }

    #[test]
    fn a_thread_without_a_claim_lists_nothing() {
        assert!(expired_in(7, &[text(1, "just a comment")], NOW).is_none());
        assert!(expired_in(7, &[], NOW).is_none());
    }

    struct Forge {
        search: Value,
        threads: BTreeMap<u64, Option<Value>>,
        seen: RefCell<Vec<String>>,
    }
    impl Transport for Forge {
        fn send(&self, req: &HttpRequest) -> std::result::Result<HttpResponse, String> {
            self.seen
                .borrow_mut()
                .push(format!("{} {}", req.method, req.url));
            if req.url.contains("/search/issues") {
                return Ok(HttpResponse::new(200).with_body(self.search.clone()));
            }
            let n: u64 = req
                .url
                .split("/issues/")
                .nth(1)
                .and_then(|r| r.split('/').next())
                .and_then(|n| n.parse().ok())
                .unwrap();
            match self.threads.get(&n) {
                Some(Some(body)) => Ok(HttpResponse::new(200).with_body(body.clone())),
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

    fn comment(c: &IssueComment) -> Value {
        json!({"id": c.id, "body": c.body, "user": {"login": c.author}})
    }

    fn thread(cs: &[IssueComment]) -> Option<Value> {
        Some(Value::Array(cs.iter().map(comment).collect()))
    }

    fn forge(total: u64, numbers: &[u64], threads: Vec<(u64, Option<Value>)>) -> Forge {
        Forge {
            search: json!({
                "total_count": total,
                "incomplete_results": false,
                "items": numbers.iter().map(|n| json!({"number": n})).collect::<Vec<_>>(),
            }),
            threads: threads.into_iter().collect(),
            seen: Default::default(),
        }
    }

    #[test]
    fn the_list_is_sorted_longest_expired_first_and_reads_only() {
        let f = forge(
            3,
            &[3, 5, 9],
            vec![
                (3, thread(&[claim(30, "agent-3", 1, NOW - 100)])),
                (5, thread(&[claim(50, "agent-5", 1, NOW + 900)])),
                (9, thread(&[claim(90, "agent-9", 1, NOW - 4000)])),
            ],
        );
        let v = list("o", "r", NOW, &f, &C0).unwrap();
        let rows = v["expired"].as_array().unwrap();
        assert_eq!(rows.len(), 2, "{v}");
        assert_eq!(rows[0]["issue"], 9);
        assert_eq!(rows[0]["expired_secs"], 4000);
        assert_eq!(rows[0]["session"], "agent-9");
        assert_eq!(rows[0]["epoch"], 1);
        assert_eq!(rows[0]["claim_comment"], 90);
        assert_eq!(rows[0]["lease_until"], at(NOW - 4000));
        assert_eq!(rows[1]["issue"], 3);
        assert_eq!(v["threads_read"], 3);
        assert_eq!(v["unread"], json!([]));
        assert_eq!(v["partial"], json!([]));
        assert!(f.seen.borrow().iter().all(|r| r.starts_with("GET ")));
    }

    #[test]
    fn a_thread_that_cannot_be_read_is_named_and_never_taken_for_clean() {
        let f = forge(
            2,
            &[3, 4],
            vec![
                (3, thread(&[claim(30, "agent-3", 1, NOW - 100)])),
                (4, None),
            ],
        );
        let v = list("o", "r", NOW, &f, &C0).unwrap();
        assert_eq!(v["unread"], json!([4]), "{v}");
        assert_eq!(v["threads_read"], 1);
        assert_eq!(v["expired"].as_array().unwrap().len(), 1);
    }

    /// A forge whose answers fail: one fixed reply for the search and one for every thread.
    struct Failing {
        search: std::result::Result<HttpResponse, String>,
        thread: std::result::Result<HttpResponse, String>,
    }
    impl Transport for Failing {
        fn send(&self, req: &HttpRequest) -> std::result::Result<HttpResponse, String> {
            if req.url.contains("/search/issues") {
                self.search.clone()
            } else {
                self.thread.clone()
            }
        }
    }

    fn one_found() -> std::result::Result<HttpResponse, String> {
        Ok(HttpResponse::new(200).with_body(
            json!({"total_count": 1, "incomplete_results": false, "items": [{"number": 4}]}),
        ))
    }

    #[test]
    fn a_search_that_fails_is_an_error_and_never_an_empty_list() {
        let ok_thread = Ok(HttpResponse::new(200).with_body(json!([])));
        let cases: Vec<(&str, std::result::Result<HttpResponse, String>)> = vec![
            (
                "500",
                Ok(HttpResponse::new(500).with_body(json!({"message": "boom"}))),
            ),
            (
                "403",
                Ok(HttpResponse::new(403).with_body(json!({"message": "rate limit"}))),
            ),
            ("transport", Err("connection refused".to_string())),
            ("no body", Ok(HttpResponse::new(200))),
            (
                "no items",
                Ok(HttpResponse::new(200).with_body(json!({"total_count": 0}))),
            ),
            (
                "items not a list",
                Ok(HttpResponse::new(200).with_body(json!({"items": "x"}))),
            ),
        ];
        for (name, search) in cases {
            let f = Failing {
                search,
                thread: ok_thread.clone(),
            };
            assert!(list("o", "r", NOW, &f, &C0).is_err(), "{name}");
        }
    }

    #[test]
    fn a_thread_request_that_fails_is_named_in_unread() {
        let cases: Vec<(&str, std::result::Result<HttpResponse, String>)> = vec![
            (
                "404",
                Ok(HttpResponse::new(404).with_body(json!({"message": "Not Found"}))),
            ),
            (
                "500",
                Ok(HttpResponse::new(500).with_body(json!({"message": "boom"}))),
            ),
            ("transport", Err("connection reset".to_string())),
        ];
        for (name, thread) in cases {
            let f = Failing {
                search: one_found(),
                thread,
            };
            let v = list("o", "r", NOW, &f, &C0).unwrap();
            assert_eq!(v["unread"], json!([4]), "{name}: {v}");
            assert_eq!(v["threads_read"], 0, "{name}: {v}");
            assert_eq!(v["expired"], json!([]), "{name}: {v}");
        }
    }

    #[test]
    fn a_cut_search_is_partial() {
        let f = forge(
            31,
            &[3],
            vec![(3, thread(&[claim(30, "agent-3", 1, NOW - 100)]))],
        );
        let v = list("o", "r", NOW, &f, &C0).unwrap();
        assert_eq!(
            v["partial"][0], "the search found 31 threads and returned 1",
            "{v}"
        );
        // Exactly what was found: not partial.
        let f = forge(1, &[3], vec![(3, thread(&[]))]);
        assert_eq!(list("o", "r", NOW, &f, &C0).unwrap()["partial"], json!([]));
        // A flag from GitHub is said too.
        let mut f = forge(1, &[3], vec![(3, thread(&[]))]);
        f.search["incomplete_results"] = json!(true);
        assert_eq!(
            list("o", "r", NOW, &f, &C0).unwrap()["partial"][0],
            "GitHub's search flagged its result as incomplete"
        );
    }

    #[test]
    fn a_repository_that_would_change_the_url_is_refused_before_a_request() {
        let f = forge(0, &[], vec![]);
        for (o, r) in [("o", ".."), ("..", "r"), ("o/x", "r"), ("o", "r?x")] {
            assert!(list(o, r, NOW, &f, &C0).is_err(), "{o}/{r}");
        }
        assert!(f.seen.borrow().is_empty());
    }

    #[test]
    fn one_call_reads_at_most_sixty_threads_and_says_so() {
        let nums = |n: u64| (1..=n).collect::<Vec<_>>();
        let quiet = |n: u64| {
            forge(
                n,
                &nums(n),
                (1..=n).map(|i| (i, thread(&[]))).collect::<Vec<_>>(),
            )
        };
        let v = list("o", "r", NOW, &quiet(60), &C0).unwrap();
        assert_eq!(v["threads_read"], 60);
        assert_eq!(v["partial"], json!([]), "{v}");
        let v = list("o", "r", NOW, &quiet(61), &C0).unwrap();
        assert_eq!(v["threads_read"], 60);
        assert_eq!(
            v["partial"][0], "61 threads have a claim line; the first 60 were read",
            "{v}"
        );
    }

    #[test]
    fn the_search_asks_for_open_threads_with_a_claim_line() {
        let f = forge(0, &[], vec![]);
        list("o", "r", NOW, &f, &C0).unwrap();
        let seen = f.seen.borrow();
        assert_eq!(
            seen.as_slice(),
            ["GET https://api.github.com/search/issues?q=repo%3Ao%2Fr%20is%3Aopen%20in%3Acomments%20%22moai-claim%3A%22&per_page=100"]
        );
    }
}
