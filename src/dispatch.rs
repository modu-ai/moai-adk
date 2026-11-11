//! Dispatch and decision observations over the S1 durable message protocol (#2390 S2).
//! This module never grants a claim, publishes a verdict, or changes an audit gate.
//! Session IDs remain declarations. A fresh GET snapshot is not a cross-host CAS.
use crate::{
    contract::{Kind, OpError},
    forge::{policy::Clock, transport::Transport},
    message::{self, Message, Publication},
};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{
    collections::{BTreeMap, BTreeSet},
    path::Path,
};

type Result<T> = std::result::Result<T, OpError>;
#[derive(Clone, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct RecordRequest {
    pub message: Message,
}
#[derive(Clone, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct AckRequest {
    pub dispatch: Message,
    pub decision: Message,
}
#[derive(Clone, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct CheckRequest {
    pub dispatch: Message,
}
#[derive(Clone, Debug, PartialEq, Eq, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
struct Previous {
    report_id: String,
    payload_digest: String,
    comment_id: u64,
}
#[derive(Clone, Debug, PartialEq, Eq, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
struct AuditTarget {
    pr: u64,
    head: String,
}
#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
struct Dispatch {
    dispatch_id: String,
    revision: u64,
    previous: Option<Previous>,
    scope: Vec<String>,
    completion: Vec<String>,
    audit: Option<AuditTarget>,
}
#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
struct Decision {
    dispatch_report_id: String,
    dispatch_payload_digest: String,
    disposition: String,
    reason: String,
}
fn error(kind: Kind, text: &str) -> OpError {
    OpError::new(kind, text)
}
fn conflict() -> OpError {
    error(
        Kind::Conflict,
        "dispatch identity, revision or decision conflict",
    )
}
fn literal(s: &str) -> bool {
    !s.trim().is_empty() && !s.chars().any(char::is_control)
}
fn dispatch(m: &mut Message) -> Result<(Dispatch, String)> {
    let digest = message::validate(m)?;
    let p: Dispatch = serde_json::from_value(m.payload.clone())
        .map_err(|_| error(Kind::InvalidInput, "invalid dispatch payload"))?;
    if m.event != "dispatch"
        || p.revision == 0
        || m.task.revision != p.revision.to_string()
        || !literal(&p.dispatch_id)
        || p.scope.is_empty()
        || p.completion.is_empty()
        || p.scope.iter().chain(&p.completion).any(|s| !literal(s))
        || (p.revision == 1) != p.previous.is_none()
    {
        return Err(error(
            Kind::InvalidInput,
            "invalid dispatch revision, scope or completion",
        ));
    }
    if let Some(a) = &p.audit {
        if a.pr == 0
            || a.head.len() != 40
            || !a
                .head
                .bytes()
                .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
            || m.task.task_id != format!("audit:#{}@{}", a.pr, a.head)
        {
            return Err(error(Kind::InvalidInput, "invalid dispatch audit task key"));
        }
    } else if m.task.task_id.starts_with("audit:") {
        return Err(error(
            Kind::InvalidInput,
            "audit task requires exact PR and head",
        ));
    }
    Ok((p, digest))
}
fn decision(d: &mut Message, m: &Message, digest: &str) -> Result<Decision> {
    message::validate(d)?;
    let p: Decision = serde_json::from_value(d.payload.clone())
        .map_err(|_| error(Kind::InvalidInput, "invalid dispatch decision payload"))?;
    if d.event != "dispatch-decision"
        || d.report_id == m.report_id
        || d.project_id != m.project_id
        || d.task != m.task
        || d.sender != m.receiver
        || d.receiver != m.sender
        || d.owner_epoch != m.owner_epoch
        || d.leader_epoch != m.leader_epoch
        || d.candidate_digest != m.candidate_digest
        || d.approval_refs != m.approval_refs
        || d.hold_refs != m.hold_refs
        || p.dispatch_report_id != m.report_id
        || p.dispatch_payload_digest != digest
        || !["accepted", "deferred", "rejected"].contains(&p.disposition.as_str())
        || !literal(&p.reason)
    {
        return Err(conflict());
    }
    Ok(p)
}
struct Assignment {
    publication: Publication,
    payload: Dispatch,
    received: bool,
    decision: Option<Publication>,
}
struct Snapshot {
    latest_timestamp: Option<i64>,
    assignments: BTreeMap<u64, Assignment>,
}
fn previous(a: &Assignment) -> Previous {
    Previous {
        report_id: a.publication.message.report_id.clone(),
        payload_digest: a.publication.payload_digest.clone(),
        comment_id: a.publication.comment_id,
    }
}
fn follows(p: &Dispatch, m: &Message, a: &Assignment) -> bool {
    p.previous.as_ref() == Some(&previous(a))
        && p.dispatch_id == a.payload.dispatch_id
        && p.audit == a.payload.audit
        && m.sender == a.publication.message.sender
        && m.leader_epoch == a.publication.message.leader_epoch
        && m.owner_epoch >= a.publication.message.owner_epoch
}
fn snapshot<T: Transport, C: Clock>(anchor: &Message, t: &T, c: &C) -> Result<Snapshot> {
    let publications = message::dispatch_publications(anchor, t, c)?;
    let mut latest_timestamp: Option<i64> = None;
    let mut records = BTreeSet::new();
    let mut acknowledged = BTreeSet::new();
    let mut report_ids = BTreeMap::new();
    let mut assignments: BTreeMap<u64, Assignment> = BTreeMap::new();
    let mut receipts = Vec::new();
    let mut decisions = Vec::new();
    let mut ids = BTreeMap::new();
    for mut p in publications {
        // One server id must not describe two packets, even on overlapping pages.
        let identity = (
            serde_json::to_vec(&p.message).map_err(|_| conflict())?,
            p.ack,
            p.created_at,
        );
        if let Some(old) = ids.insert(p.comment_id, identity.clone()) {
            if old != identity {
                return Err(conflict());
            }
        }
        // S1 identity is shared by every record/ACK, before all kind/context filters.
        if p.message.report_id == anchor.report_id && p.message != *anchor {
            return Err(conflict());
        }
        if let Some(old) = report_ids.insert(p.message.report_id.clone(), p.message.clone()) {
            if old != p.message {
                return Err(conflict());
            }
        }
        if p.ack {
            acknowledged.insert(p.message.report_id.clone());
        } else {
            records.insert(p.message.report_id.clone());
        }
        if p.message.project_id != anchor.project_id
            || p.message.task.task_id != anchor.task.task_id
        {
            continue;
        }
        if p.message.event != "dispatch" && p.message.event != "dispatch-decision" {
            continue;
        }
        let mut task = p.message.task.clone();
        task.revision = anchor.task.revision.clone();
        if task != anchor.task {
            return Err(conflict());
        }
        latest_timestamp = Some(latest_timestamp.map_or(p.created_at, |old| old.max(p.created_at)));
        if p.message.event == "dispatch-decision" {
            if !p.ack {
                decisions.push(p);
            }
            continue;
        }
        let (payload, digest) = dispatch(&mut p.message)?;
        if digest != p.payload_digest {
            return Err(conflict());
        }
        if p.ack {
            receipts.push(p);
            continue;
        }
        if let Some(old) = assignments.get_mut(&payload.revision) {
            if old.publication.message != p.message {
                return Err(conflict());
            }
            if p.comment_id < old.publication.comment_id {
                old.publication = p;
            }
        } else {
            assignments.insert(
                payload.revision,
                Assignment {
                    publication: p,
                    payload,
                    received: false,
                    decision: None,
                },
            );
        }
    }
    if !acknowledged.is_subset(&records) {
        return Err(conflict());
    }
    let mut prior: Option<(u64, &Assignment)> = None;
    for (&rev, a) in &assignments {
        match prior {
            None if rev == 1 => {}
            Some((oldrev, old))
                if oldrev.checked_add(1) == Some(rev)
                    && follows(&a.payload, &a.publication.message, old) => {}
            _ => return Err(conflict()),
        }
        prior = Some((rev, a));
    }
    for receipt in receipts {
        let rev = receipt
            .message
            .task
            .revision
            .parse::<u64>()
            .map_err(|_| conflict())?;
        let a = assignments.get_mut(&rev).ok_or_else(conflict)?;
        if receipt.message != a.publication.message || receipt.created_at < a.publication.created_at
        {
            return Err(conflict());
        }
        a.received = true;
    }
    for mut d in decisions {
        let rev = d
            .message
            .task
            .revision
            .parse::<u64>()
            .map_err(|_| conflict())?;
        let a = assignments.get_mut(&rev).ok_or_else(conflict)?;
        decision(
            &mut d.message,
            &a.publication.message,
            &a.publication.payload_digest,
        )?;
        if !a.received || d.created_at < a.publication.created_at {
            return Err(conflict());
        }
        if let Some(old) = &a.decision {
            if old.message != d.message {
                return Err(conflict());
            }
        }
        if a.decision
            .as_ref()
            .is_none_or(|old| d.comment_id < old.comment_id)
        {
            a.decision = Some(d);
        }
    }
    Ok(Snapshot {
        assignments,
        latest_timestamp,
    })
}
fn view(s: &Snapshot, anchor: &Message, now: i64) -> Result<Value> {
    let (rev, current) = s
        .assignments
        .last_key_value()
        .ok_or_else(|| error(Kind::Blocked, "dispatch publication not observed"))?;
    let requested = anchor
        .task
        .revision
        .parse::<u64>()
        .map_err(|_| conflict())?;
    let a = s.assignments.get(&requested).ok_or_else(conflict)?;
    if a.publication.message != *anchor {
        return Err(conflict());
    }
    let elapsed = now
        .checked_sub(current.publication.created_at)
        .filter(|n| *n >= 0 && s.latest_timestamp.is_some_and(|at| at <= now));
    let alert = match elapsed {
        None => "unknown",
        Some(_) if current.received => "received",
        Some(n) if n >= 300 => "unacknowledged",
        Some(_) => "waiting",
    };
    let disposition = current
        .decision
        .as_ref()
        .map(|p| p.message.payload["disposition"].clone())
        .unwrap_or(Value::Null);
    Ok(
        json!({"dispatch_id":current.payload.dispatch_id,"requested_revision":requested,"current_revision":rev,
        "current_report_id":current.publication.message.report_id,"current_payload_digest":current.publication.payload_digest,
        "record_comment_id":current.publication.comment_id,"queued":true,"received_ack":current.received,
        "disposition":disposition,"decision_comment_id":current.decision.as_ref().map(|p|p.comment_id),
        "superseded":requested!=*rev,"observation_current":requested==*rev && elapsed.is_some(),"ack_alert":alert,"elapsed_secs":elapsed,
        "observed_at":now,"claim_authorized":false,"ownership_transferred":false,"audit_gate_integrated":false,"cross_host_atomic":false,
        "business_completion":false,"c3":"observation_only_gate_integration_unfulfilled"}),
    )
}
pub fn check_cli<T: Transport, C: Clock>(mut req: CheckRequest, t: &T, c: &C) -> Result<Value> {
    dispatch(&mut req.dispatch)?;
    view(&snapshot(&req.dispatch, t, c)?, &req.dispatch, c.now())
}
pub fn record_cli<T: Transport, C: Clock>(
    root: &Path,
    mut req: RecordRequest,
    permission: &str,
    t: &T,
    c: &C,
) -> Result<Value> {
    message::permission_gate(permission)?;
    let (p, _) = dispatch(&mut req.message)?;
    let s = snapshot(&req.message, t, c)?;
    if let Some(a) = s.assignments.get(&p.revision) {
        if a.publication.message != req.message {
            return Err(conflict());
        }
        // An old packet can be observed again, but never republished by S2.
        message::record_cli(
            root,
            message::RecordRequest {
                message: req.message.clone(),
            },
            permission,
            t,
            c,
        )?;
        return check_cli(
            CheckRequest {
                dispatch: req.message,
            },
            t,
            c,
        );
    }
    match s.assignments.last_key_value() {
        None if p.revision == 1 => {}
        Some((rev, a))
            if rev.checked_add(1) == Some(p.revision) && follows(&p, &req.message, a) => {}
        _ => return Err(conflict()),
    }
    message::record_cli(
        root,
        message::RecordRequest {
            message: req.message.clone(),
        },
        permission,
        t,
        c,
    )?;
    check_cli(
        CheckRequest {
            dispatch: req.message,
        },
        t,
        c,
    )
}
pub fn ack_cli<T: Transport, C: Clock>(
    root: &Path,
    mut req: AckRequest,
    permission: &str,
    t: &T,
    c: &C,
) -> Result<Value> {
    message::permission_gate(permission)?;
    let (_, digest) = dispatch(&mut req.dispatch)?;
    decision(&mut req.decision, &req.dispatch, &digest)?;
    let s = snapshot(&req.dispatch, t, c)?;
    let v = view(&s, &req.dispatch, c.now())?;
    if v["observation_current"] != true {
        return Err(error(
            Kind::Blocked,
            "dispatch superseded or time unknown; decision blocked",
        ));
    }
    let a = s.assignments.last_key_value().unwrap().1;
    if a.decision
        .as_ref()
        .is_some_and(|p| p.message != req.decision)
    {
        return Err(conflict());
    }
    message::ack_cli(
        root,
        message::AckRequest {
            message: req.dispatch.clone(),
            payload_digest: digest,
        },
        permission,
        t,
        c,
    )?;
    // Receipt publication can take time; re-observe revision before the decision POST.
    let v = check_cli(
        CheckRequest {
            dispatch: req.dispatch.clone(),
        },
        t,
        c,
    )?;
    if v["observation_current"] != true {
        return Err(error(
            Kind::Blocked,
            "dispatch superseded before decision; blocked",
        ));
    }
    message::record_cli(
        root,
        message::RecordRequest {
            message: req.decision,
        },
        permission,
        t,
        c,
    )?;
    check_cli(
        CheckRequest {
            dispatch: req.dispatch,
        },
        t,
        c,
    )
}
