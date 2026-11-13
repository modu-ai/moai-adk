//! `moai-cli wait` (#2676): one run that polls GitHub until a condition holds, then ends.
//!
//! Every poll goes through [`Client::get`] (conditional GET with the ETag of the previous
//! answer) and [`Client::send_with_backoff`] (retry-after, rate-limit reset, secondary limit),
//! so an unchanged poll is a 304. GitHub: "Making a conditional request does not count
//! against your primary rate limit if a `304` response is returned and the request was made
//! while correctly authorized" (docs.github.com/en/rest/using-the-rest-api/
//! best-practices-for-using-the-rest-api, read 2026-11-13); `gh api` always sends its
//! authorization. Requests are serial, and the wait between polls is at least the
//! `x-poll-interval` GitHub sends. The clock is injected, so tests never sleep.

use super::policy::{next_link, Clock};
use super::transport::{HttpResponse, Transport};
use super::{valid_name, CachedPage, Client, Config, ForgeError};
use serde_json::{json, Value};
use std::collections::BTreeMap;

/// Longest wait one run accepts; a longer watch is several runs.
pub const MAX_TIMEOUT_S: u64 = 3600;
/// Shortest pause between polls.
pub const MIN_INTERVAL_S: u64 = 10;
/// Comment text kept per comment for prefix matching (the output carries the first line only).
const COMMENT_TEXT_LIMIT: usize = 4000;
/// First-line bytes a matched comment reports.
const FIRST_LINE_LIMIT: usize = 200;
/// Longest `--prefix`, in characters (not bytes).
const PREFIX_LIMIT: usize = 200;

/// What to wait for.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Condition {
    /// Every workflow run of the head commit has completed.
    Ci { head: String },
    /// A comment newer than `after` (by id); with `prefix`, one with a line starting with it.
    Comment {
        issue: u64,
        after: u64,
        prefix: Option<String>,
    },
    /// The pull request's head is no longer `head`, or the pull request is no longer open.
    Pr { pr: u64, head: String },
}

impl Condition {
    fn name(&self) -> &'static str {
        match self {
            Condition::Ci { .. } => "ci",
            Condition::Comment { .. } => "comment",
            Condition::Pr { .. } => "pr",
        }
    }
}

#[derive(Clone, Debug)]
pub struct WaitRequest {
    pub owner: String,
    pub repo: String,
    pub condition: Condition,
    pub timeout_s: u64,
    pub interval_s: u64,
}

fn hex40(s: &str) -> bool {
    s.len() == 40
        && s.bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}

fn validate(req: &WaitRequest) -> Result<(), ForgeError> {
    if !valid_name(&req.owner) || !valid_name(&req.repo) {
        return Err(ForgeError::invalid("owner and repo must be GitHub names"));
    }
    if req.timeout_s == 0 || req.timeout_s > MAX_TIMEOUT_S {
        return Err(ForgeError::invalid(format!(
            "wait timeout must be 1..={MAX_TIMEOUT_S} seconds"
        )));
    }
    if req.interval_s < MIN_INTERVAL_S {
        return Err(ForgeError::invalid(format!(
            "wait interval must be at least {MIN_INTERVAL_S} seconds"
        )));
    }
    match &req.condition {
        Condition::Ci { head } | Condition::Pr { head, .. } if !hex40(head) => {
            Err(ForgeError::invalid("wait head must be 40 lowercase hex"))
        }
        Condition::Comment { issue: 0, .. } | Condition::Pr { pr: 0, .. } => Err(
            ForgeError::invalid("wait issue or pull request number must be positive"),
        ),
        Condition::Comment {
            prefix: Some(p), ..
        } if p.is_empty()
            || p.chars().count() > PREFIX_LIMIT
            || p.chars().any(char::is_control) =>
        {
            Err(ForgeError::invalid(
                "wait prefix must be 1..=200 characters without control characters",
            ))
        }
        _ => Ok(()),
    }
}

/// The value a response projects to, from the cache on a 304.
fn projected<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    url: &str,
    resp: &HttpResponse,
    project: &dyn Fn(&HttpResponse) -> Value,
) -> Value {
    client
        .used
        .borrow()
        .get(url)
        .map(|c| c.value.clone())
        .unwrap_or_else(|| project(resp))
}

fn project_runs(resp: &HttpResponse) -> Value {
    let runs = resp
        .body
        .as_ref()
        .and_then(|b| b["workflow_runs"].as_array().cloned())
        .unwrap_or_default();
    Value::Array(
        runs.iter()
            .map(|r| {
                json!({"id": r["id"], "name": r["name"], "status": r["status"],
                       "conclusion": r["conclusion"]})
            })
            .collect(),
    )
}

fn project_pr(resp: &HttpResponse) -> Value {
    let b = resp.body.clone().unwrap_or(Value::Null);
    json!({"head": b["head"]["sha"], "state": b["state"], "merged": b["merged"]})
}

fn project_comments(resp: &HttpResponse) -> Value {
    let items: Vec<Value> = resp
        .body
        .as_ref()
        .and_then(Value::as_array)
        .map(|a| {
            a.iter()
                .map(|c| {
                    let text: String = c["body"]
                        .as_str()
                        .unwrap_or("")
                        .chars()
                        .take(COMMENT_TEXT_LIMIT)
                        .collect();
                    json!({"id": c["id"], "author": c["user"]["login"], "text": text})
                })
                .collect()
        })
        .unwrap_or_default();
    json!({"items": items, "next": next_link(resp.header("link"))})
}

fn first_line(text: &str) -> String {
    let line = text.lines().next().unwrap_or("");
    let mut end = line.len().min(FIRST_LINE_LIMIT);
    while !line.is_char_boundary(end) {
        end -= 1;
    }
    line[..end].to_string()
}

/// One poll: `Some(value)` when the condition holds, and whether every response was a 304.
fn check<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    base: &str,
    condition: &Condition,
) -> Result<(Option<Value>, bool), ForgeError> {
    match condition {
        Condition::Ci { head } => {
            let url = format!("{base}/actions/runs?head_sha={head}&per_page=100");
            let resp = client.get(&url, Some(&project_runs))?;
            let runs = projected(client, &url, &resp, &project_runs);
            let list = runs.as_array().cloned().unwrap_or_default();
            let done = !list.is_empty() && list.iter().all(|r| r["status"] == "completed");
            let ok = |r: &Value| {
                matches!(
                    r["conclusion"].as_str(),
                    Some("success" | "skipped" | "neutral")
                )
            };
            let value = done.then(|| {
                let conclusion = if list.iter().all(ok) {
                    "success"
                } else {
                    "failure"
                };
                json!({"head": head, "conclusion": conclusion, "runs": runs})
            });
            Ok((value, resp.status == 304))
        }
        Condition::Pr { pr, head } => {
            let url = format!("{base}/pulls/{pr}");
            let resp = client.get(&url, Some(&project_pr))?;
            let now = projected(client, &url, &resp, &project_pr);
            let moved = now["head"].as_str() != Some(head.as_str()) || now["state"] != "open";
            Ok((moved.then_some(now), resp.status == 304))
        }
        Condition::Comment {
            issue,
            after,
            prefix,
        } => {
            let url = format!("{base}/issues/{issue}/comments?per_page=100");
            let (items, all_cached) = client.pages(url, &project_comments)?;
            let hit = items.iter().find(|c| {
                c["id"].as_u64().is_some_and(|id| id > *after)
                    && prefix.as_ref().is_none_or(|p| {
                        c["text"]
                            .as_str()
                            .unwrap_or("")
                            .lines()
                            .any(|l| l.starts_with(p.as_str()))
                    })
            });
            let value = hit.map(|c| {
                json!({"id": c["id"], "author": c["author"],
                       "first_line": first_line(c["text"].as_str().unwrap_or(""))})
            });
            Ok((value, all_cached))
        }
    }
}

/// Polls until the condition holds or `timeout_s` passes. A timeout is an answer
/// (`matched: false`), not an error; authentication or rate-limit failures stay errors.
pub fn wait<T: Transport, C: Clock>(
    req: &WaitRequest,
    transport: &T,
    clock: &C,
) -> Result<Value, ForgeError> {
    validate(req)?;
    let base = format!("https://api.github.com/repos/{}/{}", req.owner, req.repo);
    let start = clock.now();
    let mut cache: BTreeMap<String, CachedPage> = BTreeMap::new();
    let (mut polls, mut not_modified) = (0u64, 0u64);
    let mut rate = Value::Null;
    loop {
        let client = Client::new(transport, clock, Config::default(), cache.clone());
        let (found, all_cached) = check(&client, &base, &req.condition)?;
        polls += 1;
        if all_cached {
            not_modified += 1;
        }
        if let Some(r) = client.last_rate.borrow().clone() {
            rate = r;
        }
        let hint = *client.poll_interval_header.borrow();
        cache.extend(client.used.into_inner());
        let elapsed = (clock.now() - start).max(0) as u64;
        let answer = |matched: bool, value: Value| {
            json!({
                "matched": matched,
                "condition": req.condition.name(),
                "value": value,
                "polls": polls,
                "not_modified": not_modified,
                "rate_limit": rate,
                "elapsed_secs": elapsed,
            })
        };
        if let Some(value) = found {
            return Ok(answer(true, value));
        }
        if elapsed >= req.timeout_s {
            return Ok(answer(false, Value::Null));
        }
        let pause = req
            .interval_s
            .max(hint.unwrap_or(0))
            .min(req.timeout_s - elapsed);
        clock.sleep(pause);
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::forge::transport::{Exchange, HttpRequest, RecordedTransport};
    use std::cell::{Cell, RefCell};

    struct FakeClock {
        now: Cell<i64>,
        slept: RefCell<Vec<u64>>,
    }
    impl Clock for FakeClock {
        fn now(&self) -> i64 {
            self.now.get()
        }
        fn sleep(&self, secs: u64) {
            self.slept.borrow_mut().push(secs);
            self.now.set(self.now.get() + secs as i64);
        }
    }
    fn clock() -> FakeClock {
        FakeClock {
            now: Cell::new(1_000),
            slept: RefCell::new(Vec::new()),
        }
    }

    const HEAD: &str = "0123456789abcdef0123456789abcdef01234567";
    const RUNS: &str = "https://api.github.com/repos/o/r/actions/runs?head_sha=0123456789abcdef0123456789abcdef01234567&per_page=100";

    fn req(condition: Condition, timeout_s: u64) -> WaitRequest {
        WaitRequest {
            owner: "o".into(),
            repo: "r".into(),
            condition,
            timeout_s,
            interval_s: 30,
        }
    }
    fn get(url: &str, etag: Option<&str>) -> HttpRequest {
        let r = HttpRequest::new("GET", url);
        match etag {
            Some(e) => r.with_header("If-None-Match", e),
            None => r,
        }
    }
    fn ok(body: Value, etag: &str, remaining: &str) -> HttpResponse {
        HttpResponse::new(200)
            .with_header("etag", etag)
            .with_header("x-ratelimit-limit", "5000")
            .with_header("x-ratelimit-remaining", remaining)
            .with_body(body)
    }
    fn not_modified(etag: &str, remaining: &str) -> HttpResponse {
        HttpResponse::new(304)
            .with_header("etag", etag)
            .with_header("x-ratelimit-limit", "5000")
            .with_header("x-ratelimit-remaining", remaining)
    }
    fn runs(status: &str, conclusion: Value) -> Value {
        json!({"workflow_runs": [{"id": 7, "name": "ci", "status": status, "conclusion": conclusion}]})
    }

    /// AC-1/2/3: the second poll asks with the first ETag and is a 304 (the remaining count
    /// does not drop), the third sees the run completed and the run ends at once.
    #[test]
    fn unchanged_polls_are_304s_and_a_true_condition_ends_the_run() {
        let t = RecordedTransport::new(vec![
            Exchange {
                request: get(RUNS, None),
                response: ok(runs("in_progress", Value::Null), "\"a\"", "4990"),
            },
            Exchange {
                request: get(RUNS, Some("\"a\"")),
                response: not_modified("\"a\"", "4990"),
            },
            Exchange {
                request: get(RUNS, Some("\"a\"")),
                response: ok(runs("completed", json!("success")), "\"b\"", "4989"),
            },
        ]);
        let c = clock();
        let out = wait(&req(Condition::Ci { head: HEAD.into() }, 600), &t, &c).unwrap();
        assert_eq!(out["matched"], true);
        assert_eq!(out["value"]["conclusion"], "success");
        assert_eq!(out["polls"], 3);
        assert_eq!(out["not_modified"], 1);
        assert_eq!(out["rate_limit"]["remaining"], 4989);
        assert_eq!(*c.slept.borrow(), vec![30, 30]);
        assert_eq!(t.remaining(), 0, "no request after the condition held");
    }

    /// A failed run completes the wait with `failure`; a timeout answers `matched: false`
    /// without sleeping past the deadline.
    #[test]
    fn failure_is_an_answer_and_timeout_is_matched_false() {
        let t = RecordedTransport::new(vec![Exchange {
            request: get(RUNS, None),
            response: ok(runs("completed", json!("failure")), "\"a\"", "10"),
        }]);
        let out = wait(&req(Condition::Ci { head: HEAD.into() }, 60), &t, &clock()).unwrap();
        assert_eq!(out["value"]["conclusion"], "failure");

        let t = RecordedTransport::new(vec![
            Exchange {
                request: get(RUNS, None),
                response: ok(runs("queued", Value::Null), "\"a\"", "10"),
            },
            Exchange {
                request: get(RUNS, Some("\"a\"")),
                response: not_modified("\"a\"", "10"),
            },
            Exchange {
                request: get(RUNS, Some("\"a\"")),
                response: not_modified("\"a\"", "10"),
            },
        ]);
        let c = clock();
        let out = wait(&req(Condition::Ci { head: HEAD.into() }, 45), &t, &c).unwrap();
        assert_eq!(out["matched"], false);
        assert_eq!(out["value"], Value::Null);
        assert_eq!(
            *c.slept.borrow(),
            vec![30, 15],
            "the last pause stops at the deadline"
        );
        assert_eq!(out["elapsed_secs"], 45);
    }

    /// With several workflow runs on the head, one still running keeps the wait going even
    /// when another has completed (mutation W6).
    #[test]
    fn ci_waits_until_every_run_has_completed() {
        let two = |second: &str| {
            json!({"workflow_runs": [
                {"id": 1, "name": "ci", "status": "completed", "conclusion": "success"},
                {"id": 2, "name": "docs", "status": second, "conclusion": if second == "completed" { json!("success") } else { Value::Null }}
            ]})
        };
        let t = RecordedTransport::new(vec![
            Exchange {
                request: get(RUNS, None),
                response: ok(two("in_progress"), "\"a\"", "10"),
            },
            Exchange {
                request: get(RUNS, Some("\"a\"")),
                response: ok(two("completed"), "\"b\"", "9"),
            },
        ]);
        let out = wait(&req(Condition::Ci { head: HEAD.into() }, 600), &t, &clock()).unwrap();
        assert_eq!(out["polls"], 2, "one run still in progress is not done");
        assert_eq!(out["value"]["runs"].as_array().unwrap().len(), 2);
    }

    /// GitHub's `x-poll-interval` lengthens the pause; it never shortens it.
    #[test]
    fn x_poll_interval_is_a_floor() {
        let t = RecordedTransport::new(vec![
            Exchange {
                request: get(RUNS, None),
                response: ok(runs("queued", Value::Null), "\"a\"", "10")
                    .with_header("x-poll-interval", "60"),
            },
            Exchange {
                request: get(RUNS, Some("\"a\"")),
                response: ok(runs("completed", json!("success")), "\"b\"", "9"),
            },
        ]);
        let c = clock();
        wait(&req(Condition::Ci { head: HEAD.into() }, 600), &t, &c).unwrap();
        assert_eq!(*c.slept.borrow(), vec![60]);
    }

    /// A comment counts only when its id is past `after` and, with a prefix, when one of its
    /// lines starts with it — a quoted mention mid-line does not.
    #[test]
    fn comment_needs_a_newer_id_and_a_line_starting_with_the_prefix() {
        let url = "https://api.github.com/repos/o/r/issues/5/comments?per_page=100";
        let body = json!([
            {"id": 10, "user": {"login": "a"}, "body": "moai-claim-handoff: comment=1 to=x by=leader"},
            {"id": 20, "user": {"login": "b"}, "body": "quoting `moai-claim-handoff: comment=2` mid-line"},
            {"id": 30, "user": {"login": "c"}, "body": "note\nmoai-claim-handoff: comment=3 to=agent-10 by=leader\nmore"}
        ]);
        let t = RecordedTransport::new(vec![Exchange {
            request: get(url, None),
            response: ok(body, "\"c\"", "100"),
        }]);
        let cond = Condition::Comment {
            issue: 5,
            after: 10,
            prefix: Some("moai-claim-handoff:".into()),
        };
        let out = wait(&req(cond, 60), &t, &clock()).unwrap();
        assert_eq!(out["value"]["id"], 30);
        assert_eq!(out["value"]["first_line"], "note");
    }

    /// `pr`: a moved head or a closed pull request ends the wait.
    #[test]
    fn pr_ends_on_a_new_head_or_a_closed_state() {
        let url = "https://api.github.com/repos/o/r/pulls/9";
        let pr =
            |sha: &str, state: &str| json!({"head": {"sha": sha}, "state": state, "merged": false});
        let other = "1111111111111111111111111111111111111111";
        let t = RecordedTransport::new(vec![
            Exchange {
                request: get(url, None),
                response: ok(pr(HEAD, "open"), "\"p\"", "50"),
            },
            Exchange {
                request: get(url, Some("\"p\"")),
                response: ok(pr(other, "open"), "\"q\"", "49"),
            },
        ]);
        let out = wait(
            &req(
                Condition::Pr {
                    pr: 9,
                    head: HEAD.into(),
                },
                600,
            ),
            &t,
            &clock(),
        )
        .unwrap();
        assert_eq!(out["value"]["head"], other);
        let t = RecordedTransport::new(vec![Exchange {
            request: get(url, None),
            response: ok(pr(HEAD, "closed"), "\"p\"", "50"),
        }]);
        let out = wait(
            &req(
                Condition::Pr {
                    pr: 9,
                    head: HEAD.into(),
                },
                600,
            ),
            &t,
            &clock(),
        )
        .unwrap();
        assert_eq!(out["value"]["state"], "closed");
    }

    /// Bad input is refused before any request.
    #[test]
    fn invalid_requests_send_nothing() {
        let t = RecordedTransport::new(vec![]);
        let bad = [
            req(Condition::Ci { head: "abc".into() }, 60),
            req(Condition::Ci { head: HEAD.into() }, 0),
            req(Condition::Ci { head: HEAD.into() }, MAX_TIMEOUT_S + 1),
            WaitRequest {
                interval_s: MIN_INTERVAL_S - 1,
                ..req(Condition::Ci { head: HEAD.into() }, 60)
            },
            req(
                Condition::Comment {
                    issue: 0,
                    after: 0,
                    prefix: None,
                },
                60,
            ),
            req(
                Condition::Comment {
                    issue: 1,
                    after: 0,
                    prefix: Some(String::new()),
                },
                60,
            ),
            WaitRequest {
                owner: "bad/owner".into(),
                ..req(
                    Condition::Pr {
                        pr: 1,
                        head: HEAD.into(),
                    },
                    60,
                )
            },
        ];
        for r in bad {
            assert!(wait(&r, &t, &clock()).is_err(), "{:?}", r.condition);
        }
        assert!(t.sent.borrow().is_empty());
        assert_eq!(MAX_TIMEOUT_S, 3600);
        assert_eq!(PREFIX_LIMIT, 200);
        assert_eq!(MIN_INTERVAL_S, 10);
        let edge = req(Condition::Ci { head: HEAD.into() }, MAX_TIMEOUT_S);
        assert!(validate(&edge).is_ok());
        let edge = WaitRequest {
            interval_s: MIN_INTERVAL_S,
            ..edge
        };
        assert!(validate(&edge).is_ok());
    }

    /// Right after a push the head may have no workflow run yet: an empty list is not
    /// "every run completed" and the wait goes on (audit 6849844828 B1, mutation A5).
    #[test]
    fn ci_with_no_run_yet_keeps_waiting() {
        let t = RecordedTransport::new(vec![
            Exchange {
                request: get(RUNS, None),
                response: ok(json!({"workflow_runs": []}), "\"a\"", "10"),
            },
            Exchange {
                request: get(RUNS, Some("\"a\"")),
                response: ok(runs("completed", json!("success")), "\"b\"", "9"),
            },
        ]);
        let out = wait(&req(Condition::Ci { head: HEAD.into() }, 600), &t, &clock()).unwrap();
        assert_eq!(out["polls"], 2, "an empty run list is not done");
        assert_eq!(out["value"]["conclusion"], "success");
    }

    /// `skipped` and `neutral` runs (this repository skips `windows-build` on pull requests)
    /// do not turn the answer into `failure` (audit 6849844828 B2, mutation A6).
    #[test]
    fn skipped_and_neutral_runs_count_as_success() {
        let body = json!({"workflow_runs": [
            {"id": 1, "name": "ci", "status": "completed", "conclusion": "success"},
            {"id": 2, "name": "windows-build", "status": "completed", "conclusion": "skipped"},
            {"id": 3, "name": "lint", "status": "completed", "conclusion": "neutral"}
        ]});
        let t = RecordedTransport::new(vec![Exchange {
            request: get(RUNS, None),
            response: ok(body, "\"a\"", "10"),
        }]);
        let out = wait(&req(Condition::Ci { head: HEAD.into() }, 60), &t, &clock()).unwrap();
        assert_eq!(out["value"]["conclusion"], "success");
    }

    /// `--prefix` counts characters: 200 Hangul syllables (600 bytes) pass, 201 are refused
    /// (audit 6849844828 B3, mutation A1).
    #[test]
    fn prefix_limit_is_200_characters_on_both_sides() {
        let with = |n: usize| {
            req(
                Condition::Comment {
                    issue: 1,
                    after: 0,
                    prefix: Some("가".repeat(n)),
                },
                60,
            )
        };
        assert!(validate(&with(PREFIX_LIMIT)).is_ok());
        assert!(validate(&with(PREFIX_LIMIT + 1)).is_err());
    }

    /// The reported first line is cut at 200 bytes, backing off to a character boundary
    /// instead of panicking inside a multi-byte character (audit 6849844828 B3, A2/A4).
    #[test]
    fn first_line_is_cut_at_200_bytes_on_a_character_boundary() {
        assert_eq!(FIRST_LINE_LIMIT, 200);
        assert_eq!(first_line(&"a".repeat(300)), "a".repeat(200));
        assert_eq!(first_line(&"a".repeat(200)), "a".repeat(200));
        // "가" is 3 bytes at 199..202, straddling the 200-byte cut.
        let straddle = format!("{}가나\nsecond", "a".repeat(199));
        assert_eq!(first_line(&straddle), "a".repeat(199));
        assert_eq!(first_line("short\nsecond"), "short");
    }
}
