//! Audit reservation: `audit-claim` (rc.2 #2403 item 1; user decision on #2403 comment
//! 6784520362).
//!
//! One call reserves the independent audit of one exact PR head for one lane, by the rule the
//! lanes already follow by hand (#1975 comment 6773009778): the `moai-audit-claim: lane=<lane>
//! head=<40hex>` line with the lowest comment id on the PR wins that head. The user decided
//! that this first reservation wins and that the leader changes the holder only with a handoff
//! line; the only such line read here is `moai-claim-handoff: comment=<reservation id>
//! to=<lane> by=leader` posted on the PR. A prose release is not read.
//!
//! Order of checks, all from GitHub GET reads before any write:
//! 1. the PR is open and `head` is its current head (otherwise nothing is written);
//! 2. the lane is not an author: no `Moai-Claim: <issue>#<id>` trailer in the PR's commits
//!    names a claim comment of that session (a trailer whose claim comment cannot be read
//!    refuses the call, so an unreadable author is never taken for a non-author);
//! 3. the lane already holds a line on that head → its id, no write;
//! 4. another lane holds the head → `held_by`, no write;
//! 5. otherwise one comment POST, then one re-read; the lowest id on the head decides
//!    `reserved` or `lost`. A POST or re-read whose result is unknown gives `uncertain`.
//!
//! Not done here: deleting or editing any comment (a losing line stays; it is inert under the
//! rule), the `moai/audit` status, assignment, verdicts, leader dispatch.
use crate::claim_verify::{classify_comment, CommentKind};
use crate::forge::policy::Clock;
use crate::forge::transport::{HttpRequest, Transport};
use crate::forge::Client;
use crate::lane_status::{claim_ids_in, one, pages, read_comments, reservation_in, Comment};
use serde::Deserialize;
use serde_json::{json, Value};

type Result<T> = std::result::Result<T, String>;

const API: &str = "https://api.github.com";
/// The GitHub list-commits-on-a-PR endpoint returns at most 250 commits; at that count the
/// author check may have missed some, so it refuses instead of guessing.
const MAX_PR_COMMITS: usize = 250;
/// How authors are matched (leader answer, #2403 comment 6784934128; normalisation #2499):
/// lane names are compared by [`same_lane`], so `pr-1-hB` and `pr-1` are one lane.
pub const AUTHOR_RULE: &str = "exact match of lane with a claim session= name after the \
                               lane name rule (NAME_RULE); names outside the rule stay byte-exact";
/// The lane name rule (#2499), from the names lanes actually wrote (2026-11-1x): a name of the
/// form `<lowercase letters>-<digits>-<fragment>` (fragment: ASCII letters and digits, for a
/// session id, a restart tag or an Issue number) belongs to lane `<letters>-<digits>`, and
/// `<letters>-<digits>` is its own lane. Any other name (`lane-t74`, a UUID, a second fragment)
/// is not normalised and matches byte-exact only.
pub const NAME_RULE: &str = "<letters>-<digits>-<fragment> is lane <letters>-<digits>; \
                             other names match byte-exact";

/// The lane a name belongs to under [`NAME_RULE`], or `None` when the rule does not apply.
pub fn lane_key(name: &str) -> Option<&str> {
    let mut parts = name.splitn(3, '-');
    let (letters, digits) = (parts.next()?, parts.next()?);
    let fragment = parts.next();
    let base_ok = !letters.is_empty()
        && letters.bytes().all(|b| b.is_ascii_lowercase())
        && !digits.is_empty()
        && digits.bytes().all(|b| b.is_ascii_digit());
    let fragment_ok =
        fragment.is_none_or(|f| !f.is_empty() && f.bytes().all(|b| b.is_ascii_alphanumeric()));
    (base_ok && fragment_ok).then(|| &name[..letters.len() + 1 + digits.len()])
}

/// Whether two lane names are the same lane: byte-equal, or equal under [`NAME_RULE`].
pub fn same_lane(a: &str, b: &str) -> bool {
    a == b || matches!((lane_key(a), lane_key(b)), (Some(x), Some(y)) if x == y)
}

/// `audit-claim` stdin (typed JSON, unknown keys rejected).
#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct AuditClaimRequest {
    pub schema_version: u32,
    pub owner: String,
    pub repo: String,
    pub pr: i64,
    pub head: String,
    pub lane: String,
}

fn refused(message: impl Into<String>) -> String {
    message.into()
}

fn plain_name(s: &str) -> bool {
    !s.is_empty()
        && s.len() <= 100
        && s.bytes()
            .all(|b| b.is_ascii_alphanumeric() || b"-_.".contains(&b))
        && s != "."
        && s != ".."
}

fn hex40(s: &str) -> bool {
    s.len() == 40 && s.bytes().all(|b| matches!(b, b'0'..=b'9' | b'a'..=b'f'))
}

fn validate(req: &AuditClaimRequest) -> Result<()> {
    if req.schema_version != 1 {
        return Err(refused("unsupported request schema_version"));
    }
    if !plain_name(&req.owner) || !plain_name(&req.repo) {
        return Err(refused(
            "audit claim refused: owner and repo must be plain names",
        ));
    }
    if req.pr <= 0 {
        return Err(refused("audit claim refused: pr must be a positive number"));
    }
    if !hex40(&req.head) {
        return Err(refused(
            "audit claim refused: head must be 40 lowercase hex",
        ));
    }
    if !plain_name(&req.lane) {
        return Err(refused(
            "audit claim refused: lane must be a plain name (letters, digits, - _ .)",
        ));
    }
    Ok(())
}

/// The reservation line this tool posts; `reservation_in` reads it back.
pub fn render_line(lane: &str, head: &str) -> String {
    format!("moai-audit-claim: lane={lane} head={head}")
}

/// Who holds `head` on these PR comments: the lowest-id reservation for that head, then the
/// last leader handoff line that names that reservation's comment id.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Holder {
    pub lane: String,
    /// The comment that gives the hold: the reservation, or the handoff that moved it.
    pub comment_id: i64,
    pub reservation_id: i64,
    pub via_handoff: bool,
}

/// Reservation lines on `head`, as (lane, comment id), lowest id first.
pub fn reservations_on(comments: &[Comment], head: &str) -> Vec<(String, i64)> {
    let mut out: Vec<(String, i64)> = comments
        .iter()
        .filter_map(|c| {
            let (lane, h) = reservation_in(&c.body)?;
            (h == head).then_some((lane, c.id))
        })
        .collect();
    out.sort_by_key(|(_, id)| *id);
    out
}

fn by_leader(body: &str) -> bool {
    body.lines().map(str::trim).any(|l| {
        l.starts_with("moai-claim-handoff:") && l.split_whitespace().any(|t| t == "by=leader")
    })
}

/// The holder of `head` (user decision 6784520362): lowest reservation id wins; a later
/// `moai-claim-handoff: comment=<that id> to=<lane> by=leader` on the PR moves it, and the
/// last such line counts.
pub fn holder_of(comments: &[Comment], head: &str) -> Option<Holder> {
    let (lane, rid) = reservations_on(comments, head).into_iter().next()?;
    let handoff = comments
        .iter()
        .filter(|c| c.id > rid && by_leader(&c.body))
        .filter_map(|c| match classify_comment(&c.body) {
            CommentKind::Handoff { target, to } if target == rid => Some((to, c.id)),
            _ => None,
        })
        .max_by_key(|(_, id)| *id);
    Some(match handoff {
        Some((to, hid)) => Holder {
            lane: to,
            comment_id: hid,
            reservation_id: rid,
            via_handoff: true,
        },
        None => Holder {
            lane,
            comment_id: rid,
            reservation_id: rid,
            via_handoff: false,
        },
    })
}

fn holder_json(h: &Option<Holder>) -> Value {
    match h {
        Some(h) => json!({
            "lane": h.lane,
            "comment_id": h.comment_id,
            "reservation_id": h.reservation_id,
            "via": if h.via_handoff { "handoff" } else { "reservation" },
        }),
        None => Value::Null,
    }
}

fn list_json(r: &[(String, i64)]) -> Value {
    Value::Array(
        r.iter()
            .map(|(lane, id)| json!({"lane": lane, "comment_id": id}))
            .collect(),
    )
}

fn transport_err(e: impl Into<String>) -> String {
    e.into()
}

/// Reads (GET only) the PR, refuses a stale head or an author lane, and returns the PR's
/// conversation comments with the authors found.
fn precheck<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    req: &AuditClaimRequest,
) -> Result<(Vec<Comment>, Vec<Value>)> {
    let repo = format!("{API}/repos/{}/{}", req.owner, req.repo);
    let pr = one(client, &format!("{repo}/pulls/{}", req.pr)).map_err(transport_err)?;
    if pr["state"].as_str() != Some("open") {
        return Err(format!("audit claim conflict: PR #{} is not open", req.pr));
    }
    let current = pr["head"]["sha"].as_str().unwrap_or_default();
    if current != req.head {
        return Err(format!(
            "audit claim conflict: head is not the current head of PR #{} ({current}); \
                 re-read the PR and reserve its current head",
            req.pr
        ));
    }
    let Some(issue) = crate::leader::refs_issue(pr["body"].as_str().unwrap_or_default()) else {
        return Err(format!(
            "audit claim blocked: PR #{} names no Refs #<issue>, so its authors cannot be read",
            req.pr
        ));
    };
    let commits = pages(
        client,
        format!("{repo}/pulls/{}/commits?per_page=100", req.pr),
    )
    .map_err(transport_err)?;
    if commits.len() >= MAX_PR_COMMITS {
        return Err(format!(
            "audit claim blocked: PR #{} lists {} commits, the API limit; authors may be missing",
            req.pr,
            commits.len()
        ));
    }
    let mut claim_ids: Vec<i64> = Vec::new();
    for c in &commits {
        for id in claim_ids_in(c["commit"]["message"].as_str().unwrap_or_default(), issue) {
            if !claim_ids.contains(&id) {
                claim_ids.push(id);
            }
        }
    }
    let mut authors = Vec::new();
    if !claim_ids.is_empty() {
        let issue_comments = read_comments(
            pages(
                client,
                format!("{repo}/issues/{issue}/comments?per_page=100"),
            )
            .map_err(transport_err)?,
        );
        for id in &claim_ids {
            let session = issue_comments.iter().find(|c| c.id == *id).and_then(|c| {
                match classify_comment(&c.body) {
                    CommentKind::Claim { session, .. } => Some(session),
                    _ => None,
                }
            });
            let Some(session) = session else {
                return Err(format!(
                    "audit claim blocked: claim comment {id} named by a Moai-Claim trailer is \
                         not a claim line on Issue #{issue}; authors cannot be ruled out"
                ));
            };
            if same_lane(&session, &req.lane) {
                return Err(format!(
                    "audit claim conflict: lane {} is an author of PR #{} (claim comment {id}); \
                         an author or repairer never audits its own PR",
                    req.lane, req.pr
                ));
            }
            authors.push(json!({"session": session, "claim_comment_id": id}));
        }
    }
    let comments = read_comments(
        pages(
            client,
            format!("{repo}/issues/{}/comments?per_page=100", req.pr),
        )
        .map_err(transport_err)?,
    );
    Ok((comments, authors))
}

/// Reserve `req.head` of `req.pr` for `req.lane`. At most one comment POST.
pub fn reserve<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    req: &AuditClaimRequest,
) -> Result<Value> {
    validate(req)?;
    let (comments, authors) = precheck(client, req)?;
    let repo = format!("{API}/repos/{}/{}", req.owner, req.repo);
    let base = |outcome: &str, posted: Value, own: Option<i64>, comments: &[Comment]| {
        let holder = holder_of(comments, &req.head);
        json!({
            "outcome": outcome,
            "pr": req.pr,
            "head": req.head,
            "lane": req.lane,
            "posted": posted,
            "comment_id": own,
            "holder": holder_json(&holder),
            "reservations": list_json(&reservations_on(comments, &req.head)),
            "authors": authors,
            "writes": match posted.as_bool() {
                Some(false) => "none",
                Some(true) => "one comment",
                None => "unknown",
            },
            "rule": "lowest reservation comment id on the head wins; a leader handoff line on the PR moves it",
            "author_rule": AUTHOR_RULE,
        })
    };
    let holder = holder_of(&comments, &req.head);
    let own = reservations_on(&comments, &req.head)
        .into_iter()
        .find(|(lane, _)| same_lane(lane, &req.lane))
        .map(|(_, id)| id);
    match (&holder, own) {
        (Some(h), _) if same_lane(&h.lane, &req.lane) => {
            let id = own.unwrap_or(h.comment_id);
            return Ok(base("reserved", json!(false), Some(id), &comments));
        }
        (Some(_), Some(id)) => return Ok(base("lost", json!(false), Some(id), &comments)),
        (Some(_), None) => return Ok(base("held_by", json!(false), None, &comments)),
        (None, _) => {}
    }
    let post = HttpRequest::new("POST", format!("{repo}/issues/{}/comments", req.pr))
        .with_body(json!({"body": render_line(&req.lane, &req.head)}));
    let posted_id = match client.send_with_backoff(&post) {
        Ok(resp) if (200..300).contains(&resp.status) => {
            resp.body.as_ref().and_then(|b| b["id"].as_i64())
        }
        Ok(resp) if resp.status == 404 || resp.status == 403 || resp.status == 422 => {
            // GitHub refused the write; nothing was created.
            return Err(transport_err(format!(
                "audit claim blocked: the comment POST returned HTTP {}; nothing was posted",
                resp.status
            )));
        }
        // Any other answer, or no answer: the comment may exist.
        _ => None,
    };
    let Some(posted_id) = posted_id else {
        let mut out = base("uncertain", json!("unknown"), None, &comments);
        out["note"] = json!("the POST result is unknown; re-run audit-claim (it reads first and posts nothing if the line exists)");
        return Ok(out);
    };
    let reread = pages(
        client,
        format!("{repo}/issues/{}/comments?per_page=100", req.pr),
    )
    .map(read_comments);
    let after = match reread {
        Ok(c) if c.iter().any(|c| c.id == posted_id) => c,
        Ok(_) | Err(_) => {
            let mut out = base("uncertain", json!(true), Some(posted_id), &comments);
            out["note"] = json!(format!(
                "comment {posted_id} was posted but the re-read does not show it; the winner \
                 is unresolved — re-run audit-claim"
            ));
            return Ok(out);
        }
    };
    let won = holder_of(&after, &req.head).is_some_and(|h| same_lane(&h.lane, &req.lane));
    let outcome = if won { "reserved" } else { "lost" };
    Ok(base(outcome, json!(true), Some(posted_id), &after))
}

/// CLI: validate, read, reserve. The outcome field decides: only `reserved` lets the lane
/// start the audit.
pub fn reserve_cli<T: Transport, C: Clock>(
    req: AuditClaimRequest,
    transport: &T,
    clock: &C,
) -> Result<Value> {
    validate(&req)?;
    let client = Client::new(transport, clock, Default::default(), Default::default());
    reserve(&client, &req)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn c(id: i64, body: &str) -> Comment {
        Comment {
            id,
            created_at: String::new(),
            body: body.into(),
        }
    }

    fn h(ch: char) -> String {
        std::iter::repeat_n(ch, 40).collect()
    }

    #[test]
    fn lowest_id_on_the_head_wins_and_other_heads_are_separate() {
        let (a, b) = (h('a'), h('b'));
        let comments = vec![
            c(5, &render_line("lane-y", &a)),
            c(3, &render_line("lane-x", &b)),
            c(4, &render_line("lane-z", &a)),
        ];
        let got = holder_of(&comments, &a).unwrap();
        assert_eq!((got.lane.as_str(), got.comment_id), ("lane-z", 4));
        let got = holder_of(&comments, &b).unwrap();
        assert_eq!((got.lane.as_str(), got.comment_id), ("lane-x", 3));
        assert!(holder_of(&comments, &h('c')).is_none());
    }

    #[test]
    fn a_leader_handoff_naming_the_reservation_moves_it() {
        let a = h('a');
        let comments = vec![
            c(10, &render_line("lane-x", &a)),
            c(11, &render_line("lane-y", &a)),
            c(12, "moai-claim-handoff: comment=11 to=lane-q by=leader"),
            c(15, "moai-claim-handoff: comment=10 to=lane-y by=lane-x"),
            c(14, "moai-claim-handoff: comment=10 to=lane-z by=leader"),
        ];
        let got = holder_of(&comments, &a).unwrap();
        assert_eq!(
            got,
            Holder {
                lane: "lane-z".into(),
                comment_id: 14,
                reservation_id: 10,
                via_handoff: true
            }
        );
    }

    #[test]
    fn validate_refuses_bad_fields() {
        let ok = AuditClaimRequest {
            schema_version: 1,
            owner: "acme".into(),
            repo: "widgets".into(),
            pr: 7,
            head: h('a'),
            lane: "agent-1".into(),
        };
        assert!(validate(&ok).is_ok());
        for bad in [
            AuditClaimRequest {
                head: h('A'),
                ..ok.clone()
            },
            AuditClaimRequest {
                head: "a".repeat(39),
                ..ok.clone()
            },
            AuditClaimRequest {
                lane: "a b".into(),
                ..ok.clone()
            },
            AuditClaimRequest {
                lane: String::new(),
                ..ok.clone()
            },
            AuditClaimRequest {
                pr: 0,
                ..ok.clone()
            },
            AuditClaimRequest {
                owner: "a/b".into(),
                ..ok.clone()
            },
            AuditClaimRequest {
                schema_version: 2,
                ..ok.clone()
            },
        ] {
            assert!(validate(&bad).is_err());
        }
    }
}
