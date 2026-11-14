//! Open-Issue reads shared by the lane commands, and `queue-free` (#2675): the open cards a
//! lane may take, read only.
//!
//! The self-claim command `queue-next` (rc.2 A2 #2161) is retired (#2901): `lane-next` names
//! the next startable card (read only, [`crate::startable::judge`]) and the lane claims it with
//! `claim-acquire`. The remote stays the source of truth (user decision D-4); live GitHub runs
//! are not part of the tests (fixtures only).
use crate::claim::{parse_claim_line, ClaimComment};
use crate::forge::policy::{next_link, parse_utc, Clock};
use crate::forge::transport::Transport;
use crate::forge::Client;
use serde::Deserialize;
use serde_json::{json, Value};
use std::collections::BTreeSet;

type Result<T> = std::result::Result<T, String>;

const API: &str = "https://api.github.com";
/// Same safety bound as the claim comment pagination.
const MAX_PAGES: u32 = 50;

/// Issue numbers named on lines that start with `prefix` (`#123` tokens only).
pub(crate) fn refs_on(body: &str, prefix: &str) -> Vec<i64> {
    body.lines()
        .filter_map(|l| l.trim().strip_prefix(prefix))
        .flat_map(|rest| {
            rest.split('#')
                .skip(1)
                .filter_map(|t| {
                    let digits: String = t.chars().take_while(char::is_ascii_digit).collect();
                    digits.parse().ok()
                })
                .collect::<Vec<i64>>()
        })
        .collect()
}

pub(crate) fn get_json<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    url: &str,
) -> Result<(Value, Option<String>)> {
    let resp = client.get(url, None).map_err(String::from)?;
    let next = next_link(resp.header("link"));
    Ok((resp.body.unwrap_or(Value::Null), next))
}

/// Every open Issue object of the repository (GET only, `Link` pagination).
fn read_open_issues<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    owner: &str,
    repo: &str,
) -> Result<Vec<Value>> {
    let mut url = Some(format!(
        "{API}/repos/{owner}/{repo}/issues?state=open&per_page=100"
    ));
    let mut out = Vec::new();
    let mut pages = 0;
    while let Some(u) = url {
        pages += 1;
        if pages > MAX_PAGES {
            return Err("queue issue page limit exceeded (50)".into());
        }
        let (body, next) = get_json(client, &u)?;
        out.extend(body.as_array().cloned().unwrap_or_default());
        url = next;
    }
    Ok(out)
}

struct Handoff {
    comment_id: i64,
    target: i64,
    to: String,
}

/// Every page of an Issue's comments, its claim-family lines replayed in comment-id order: the
/// claims still open, the outstanding handoff, and whether the card ever carried a claim line,
/// so that a card whose claims were all released can be told apart from one nobody claimed
/// (#2675).
fn read_claim_history<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    owner: &str,
    repo: &str,
    number: i64,
) -> Result<(Vec<ClaimComment>, Option<Handoff>, bool)> {
    let mut url = Some(format!(
        "{API}/repos/{owner}/{repo}/issues/{number}/comments?per_page=100"
    ));
    let mut comments = Vec::new();
    let mut pages = 0;
    while let Some(u) = url {
        pages += 1;
        if pages > MAX_PAGES {
            return Err("queue claim comment page limit exceeded (50)".into());
        }
        let (body, next) = get_json(client, &u)?;
        if let Value::Array(page) = body {
            for c in page {
                let (Some(id), Some(text)) = (c["id"].as_i64(), c["body"].as_str()) else {
                    continue;
                };
                let mut kind = crate::claim_verify::classify_comment(text);
                if matches!(kind, crate::claim_verify::CommentKind::Malformed(_)) {
                    // Keep the queue's existing no-takeover boundary: a complete claim
                    // with an invalid lease still blocks until it is explicitly closed.
                    if let Some(line) = parse_claim_line(text) {
                        kind = crate::claim_verify::CommentKind::Claim {
                            session: line.session,
                            epoch: line.epoch,
                            lease_until: line.lease_until,
                        };
                    }
                }
                if matches!(
                    kind,
                    crate::claim_verify::CommentKind::Claim { .. }
                        | crate::claim_verify::CommentKind::Release { .. }
                        | crate::claim_verify::CommentKind::Handoff { .. }
                ) {
                    comments.push((id, kind));
                }
            }
        }
        url = next;
    }
    comments.sort_unstable_by_key(|(id, _)| *id);
    let mut claims = Vec::new();
    let mut known = BTreeSet::new();
    let mut handoffs: Vec<Handoff> = Vec::new();
    for (id, kind) in comments {
        match kind {
            crate::claim_verify::CommentKind::Claim {
                session,
                epoch,
                lease_until,
            } => {
                // Taking up a handoff consumes its reservation, even if that new claim
                // is subsequently released. Historical handoffs must not reserve forever.
                handoffs.retain(|h| h.to != session);
                known.insert(id);
                claims.push(ClaimComment {
                    comment_id: id,
                    session,
                    epoch,
                    lease_until,
                });
            }
            crate::claim_verify::CommentKind::Release { target } if known.contains(&target) => {
                claims.retain(|c| c.comment_id != target);
                handoffs.retain(|h| h.target != target);
            }
            crate::claim_verify::CommentKind::Handoff { target, to } if known.contains(&target) => {
                claims.retain(|c| c.comment_id != target);
                handoffs.retain(|h| h.target != target);
                handoffs.push(Handoff {
                    comment_id: id,
                    target,
                    to,
                });
            }
            _ => {}
        }
    }
    Ok((claims, handoffs.pop(), !known.is_empty()))
}

/// `queue-free` stdin (typed JSON, unknown keys rejected).
#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct FreeRequest {
    pub schema_version: u32,
    pub owner: String,
    pub repo: String,
}

fn valid_name(s: &str) -> bool {
    !s.is_empty()
        && s.len() <= 100
        && s.bytes()
            .all(|b| b.is_ascii_alphanumeric() || b"-_.".contains(&b))
        && s != "."
        && s != ".."
}

/// `queue-free` (#2675): the open cards a lane may take, read only. Every open Issue is
/// sorted into exactly one bucket — `held` (a label named `needs:decision`, `hold` or `<axis>:hold`), `epics`
/// (title `][EP-`), `blocked` (an open `Blocked-By:` Issue) — these three by the shared
/// [`crate::startable::judge`] (#2901) — `reserved` (an outstanding
/// handoff names its recipient), `claimed` (a live claim), or `free`. A free card says why:
/// `unclaimed` (no claim line ever), `released` (every claim released) or `expired` (only
/// expired claims remain; taking it over still follows the handoff rules). An empty result
/// is stated as `free_count: 0` and a `0건` summary, never as missing output; a failed read
/// is an error, never an empty list.
pub fn free<T: Transport, C: Clock>(
    req: &FreeRequest,
    client: &Client<'_, T, C>,
    clock: &C,
) -> Result<Value> {
    if req.schema_version != 1 {
        return Err("unsupported request schema_version".into());
    }
    if !valid_name(&req.owner) || !valid_name(&req.repo) {
        return Err("owner and repo must be GitHub names".into());
    }
    let issues = read_open_issues(client, &req.owner, &req.repo)?;
    let open: BTreeSet<u64> = issues.iter().filter_map(|i| i["number"].as_u64()).collect();
    let now = clock.now();
    let live = |c: &ClaimComment| parse_utc(&c.lease_until).is_some_and(|t| t > now);
    let (mut free, mut reserved, mut claimed) = (Vec::new(), Vec::new(), Vec::new());
    let (mut held, mut epics, mut blocked) = (Vec::new(), Vec::new(), Vec::new());
    let mut inspected = 0u64;
    for i in &issues {
        let Some(number) = i["number"].as_i64() else {
            continue;
        };
        if i.get("pull_request").is_some_and(|p| !p.is_null()) {
            continue;
        }
        inspected += 1;
        let title = i["title"].as_str().unwrap_or_default();
        let body = i["body"].as_str().unwrap_or_default();
        let labels: Vec<String> = i["labels"]
            .as_array()
            .map(|a| {
                a.iter()
                    .filter_map(|l| l["name"].as_str().map(str::to_string))
                    .collect()
            })
            .unwrap_or_default();
        // The REST list carries no sub-issue or blocking-Issue fields: those stay zero here.
        let card = crate::startable::Card {
            number: number as u64,
            title,
            labels: &labels,
            body,
            open_children: 0,
            blocked_by: &[],
        };
        match crate::startable::judge(&card, &open, &[]) {
            Some(crate::startable::Skip::Held(hold)) => {
                held.push(json!({"issue": number, "title": title, "labels": hold}));
                continue;
            }
            Some(crate::startable::Skip::Epic { .. }) => {
                epics.push(json!({"issue": number, "title": title}));
                continue;
            }
            Some(crate::startable::Skip::BlockedBy(open_blockers)) => {
                blocked.push(json!({"issue": number, "title": title, "blocked_by": open_blockers}));
                continue;
            }
            // No goal exclusions are passed, so `KeepSet` cannot occur.
            Some(crate::startable::Skip::KeepSet(_)) | None => {}
        }
        let (claims, handoff, ever) = read_claim_history(client, &req.owner, &req.repo, number)?;
        if let Some(h) = handoff {
            reserved.push(json!({"issue": number, "title": title, "to": h.to,
                                 "handoff_comment_id": h.comment_id}));
        } else if let Some(c) = claims
            .iter()
            .filter(|c| live(c))
            .min_by_key(|c| c.comment_id)
        {
            claimed.push(
                json!({"issue": number, "title": title, "claimed_by": c.session,
                                "comment_id": c.comment_id, "lease_until": c.lease_until}),
            );
        } else if let Some(c) = claims.iter().min_by_key(|c| c.comment_id) {
            free.push(json!({"issue": number, "title": title, "why": "expired",
                             "claimed_by": c.session, "comment_id": c.comment_id,
                             "lease_until": c.lease_until}));
        } else {
            let why = if ever { "released" } else { "unclaimed" };
            free.push(json!({"issue": number, "title": title, "why": why}));
        }
    }
    let summary = format!(
        "free {}건 (inspected {inspected}, claimed {}, reserved {}, held {}, epics {}, blocked {})",
        free.len(),
        claimed.len(),
        reserved.len(),
        held.len(),
        epics.len(),
        blocked.len()
    );
    Ok(json!({
        "schema_version": 1,
        "repository": format!("{}/{}", req.owner, req.repo),
        "inspected": inspected,
        "free_count": free.len(),
        "free": free,
        "reserved": reserved,
        "claimed": claimed,
        "held": held,
        "epics": epics,
        "blocked": blocked,
        "summary": summary,
        "writes": false,
    }))
}

/// `queue-free` with the system transport and clock; GET only, no journal.
pub fn free_cli<T: Transport, C: Clock>(
    req: FreeRequest,
    transport: &T,
    clock: &C,
) -> std::result::Result<Value, crate::contract::OpError> {
    let client = Client::new(transport, clock, Default::default(), Default::default());
    Ok(free(&req, &client, clock)?)
}

#[cfg(test)]
mod tests {
    use super::*;

    /// #2255: release and handoff close the source claim; a recipient's subsequent
    /// claim holds the card in its own right.
    #[test]
    fn released_claims_do_not_hold_the_card() {
        use crate::forge::transport::{Exchange, HttpRequest, HttpResponse, RecordedTransport};
        struct Fixed;
        impl Clock for Fixed {
            fn now(&self) -> i64 {
                0
            }
            fn sleep(&self, _: u64) {}
        }
        let lease = "2026-11-09T09:00:00Z";
        let page = |body: Value| Exchange {
            request: HttpRequest::new(
                "GET",
                format!("{API}/repos/acme/widgets/issues/7/comments?per_page=100"),
            ),
            response: HttpResponse::new(200).with_body(body),
        };
        let claim = |id: i64, s: &str| json!({"id": id, "body": format!("moai-claim: session={s} epoch=1 lease_until={lease}")});
        let t = RecordedTransport::new(vec![
            page(json!([
                claim(100, "a"),
                {"id": 150, "body": crate::claim_verify::render_release_line(100, "leader")},
            ])),
            page(json!([
                claim(100, "a"),
                {"id": 150, "body": crate::claim_verify::render_handoff_line(100, "b", "leader")},
                claim(200, "b"),
            ])),
        ]);
        let client = Client::new(&t, &Fixed, Default::default(), Default::default());
        assert!(read_claim_history(&client, "acme", "widgets", 7)
            .unwrap()
            .0
            .is_empty());
        let open = read_claim_history(&client, "acme", "widgets", 7).unwrap();
        assert_eq!(
            open.0.iter().map(|c| c.comment_id).collect::<Vec<_>>(),
            [200]
        );
    }
}
