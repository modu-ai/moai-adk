//! Stalled-claim re-inspection and explicit handoff (06-B #2006).
//!
//! 06-A records who owns a task; this module answers what happened when that owner
//! stopped. Everything here runs after a human decision (the leader's explicit
//! `claim-handoff` request IS the approval): nothing decides "the writer is dead" from a
//! timeout, a heartbeat or silence — the lease has no numbers by user decision (P08 (a)),
//! so expiry and takeover are always named, never automatic.
//!
//! - [`claim_handoff`] reads the record, probes the previous owner's liveness **only** for
//!   a `local-process` owner on the same host (REQ-06B-02), refuses an alive or
//!   indeterminate owner without `--force-explicit` + a reason (REQ-06B-03), preserves the
//!   previous worktree as evidence with read-only git only (REQ-06B-04), and closes the
//!   record as `handed_off` naming `to_epoch` (REQ-06B-05). The epoch+1 record itself is
//!   written by the next `claim-acquire` (the journal hands out max+1, so it is exactly
//!   `to_epoch`) — the closure fences everything below it.
//! - [`effect_intend`] / [`effect_confirm`] record an external write before and after it is
//!   sent (REQ-06B-06), reusing the 07-B outbox operation-id and `moai:op` marker formats.
//!   An intent below the record's current epoch is rejected as `stale_epoch` and kept only
//!   as a diagnostic (REQ-06B-08).
//! - [`effects_reconcile`] queries the remote with the stable marker for every `uncertain`
//!   effect and moves it to `confirmed` (found) or `not_performed` (confirmed absent); it
//!   never re-creates anything (REQ-06B-07). A rate-limited probe changes no state until
//!   retry-after (REQ-06B-09).
//!
//! Results are classified `passed|failed|blocked|uncertain|not_run` in the output `verdict`
//! field; anything not probed is never reported as success (REQ-06B-10).
use crate::claim::{is_active, is_key, ClaimRecord, Journal, JournalEvent, TaskRef};
use crate::forge::outbox::{op_id, op_marker, parse_op_markers};
use crate::forge::policy::{format_utc, next_link, Clock};
use crate::forge::transport::Transport;
use crate::forge::{Client, ForgeError};
use crate::{digest, identity};
use serde_json::{json, Value};
use std::{
    path::Path,
    process::{Command, Stdio},
};

type Result<T> = std::result::Result<T, String>;

/// The state a handoff closes the previous record with (REQ-06B-05). Not an active state:
/// the closure releases the claim and the next acquire writes the new epoch.
pub const HANDOFF_STATE: &str = "handed_off";

/// Marker prefix written into external artifacts so a reconcile can find them again
/// (07-B convention: the full marker is `<!-- moai:op <operation>:<project>:<local>:<n> -->`).
pub const MARKER_PREFIX: &str = "moai:op ";

fn now_of<C: Clock>(clock: &C) -> String {
    format_utc(clock.now())
}

fn current_record<'a>(events: &'a [JournalEvent], key: &str) -> Option<&'a ClaimRecord> {
    events
        .iter()
        .rev()
        .find(|e| e.record.task_key == key)
        .map(|e| &e.record)
}

fn next_epoch(events: &[JournalEvent], key: &str) -> u64 {
    events
        .iter()
        .filter(|e| e.record.task_key == key)
        .map(|e| e.record.epoch)
        .max()
        .unwrap_or(0)
        + 1
}

// ---- liveness (REQ-06B-02) ----

/// Probes one recorded `local-process` owner. `ps -p <pid> -o lstart=` answers both
/// existence and start time under `LC_ALL=C`; a differing start time means the pid was
/// reused by another process. No `kill`, no crate, no reading of any other host.
pub fn probe_liveness(pid: Option<u32>, expected_start: Option<&str>) -> (String, Value) {
    let Some(pid) = pid else {
        return (
            "indeterminate".into(),
            json!({"reason": "no pid recorded", "pid": Value::Null}),
        );
    };
    let out = Command::new("ps")
        .args(["-p", &pid.to_string(), "-o", "lstart="])
        .env("LC_ALL", "C")
        .stdin(Stdio::null())
        .stderr(Stdio::piped())
        .output();
    let Ok(out) = out else {
        return (
            "indeterminate".into(),
            json!({"pid": pid, "reason": "ps unavailable"}),
        );
    };
    if !out.status.success() {
        return (
            "dead".into(),
            json!({"pid": pid, "reason": "no such process"}),
        );
    }
    let start = String::from_utf8_lossy(&out.stdout).trim().to_string();
    if start.is_empty() {
        return (
            "indeterminate".into(),
            json!({"pid": pid, "reason": "start time unavailable"}),
        );
    }
    match expected_start {
        Some(expected) if expected == start => {
            ("alive".into(), json!({"pid": pid, "start": start}))
        }
        Some(expected) => (
            "dead".into(),
            json!({"pid": pid, "observed_start": start, "recorded_start": expected, "reason": "pid reused"}),
        ),
        None => (
            "indeterminate".into(),
            json!({"pid": pid, "observed_start": start, "reason": "no recorded start time"}),
        ),
    }
}

// ---- worktree evidence (REQ-06B-04) ----

/// Read-only git evidence of the previous worktree. Only `status`, `rev-parse` and
/// `stash list` run — no `reset`, no `checkout --`, no `stash`, no `clean`, no
/// `worktree remove` (REQ-06B-04; the checklist reviews this).
fn collect_evidence(worktree_path: &str) -> Value {
    let dir = Path::new(worktree_path);
    if !dir.is_dir() {
        return json!({
            "status": "unavailable",
            "note": "the previous worktree path is missing; dirty changes cannot be inspected and the new owner starts fresh",
        });
    }
    let git_read = |args: &[&str]| -> String {
        let mut cmd = Command::new("git");
        cmd.arg("--no-optional-locks").arg("-C").arg(dir).args(args);
        identity::scrub_git_env(&mut cmd);
        let out = cmd
            .env("GIT_OPTIONAL_LOCKS", "0")
            .stdin(Stdio::null())
            .stderr(Stdio::piped())
            .output();
        match out {
            Ok(o) if o.status.success() => String::from_utf8_lossy(&o.stdout).into_owned(),
            _ => String::new(),
        }
    };
    json!({
        "status": "collected",
        "status_v2_z": git_read(&["status", "--porcelain=v2", "-z"]),
        "head": git_read(&["rev-parse", "HEAD"]).trim().to_string(),
        "stash_list": git_read(&["stash", "list"]),
    })
}

// ---- handoff (REQ-06B-01..05) ----

/// The leader's explicit handoff. Every refusal and every probe result is recorded; the
/// only state-changing refusals are `stale_request` (an error, nothing written) and the
/// `handoff_pending` marking for an alive or indeterminate owner.
#[allow(clippy::too_many_arguments)]
pub fn claim_handoff<C: Clock>(
    journal: &Journal,
    key: &str,
    from_epoch: u64,
    reason: &str,
    force_explicit: bool,
    host_id: Option<&str>,
    clock: &C,
) -> Result<Value> {
    if !is_key(key) {
        return Err("task must be a 16 hex character claim key".into());
    }
    if reason.trim().is_empty() || reason.len() > 256 || reason.chars().any(char::is_control) {
        return Err("reason must be a bounded string without control characters".into());
    }
    let events = journal.read_events()?;
    let Some(current) = current_record(&events, key) else {
        return Ok(json!({
            "outcome": "unclaimed", "verdict": "not_run", "record": Value::Null,
            "liveness": Value::Null, "notes": ["no record exists for this task key"],
        }));
    };
    // REQ-06B-01: the request must name the epoch it saw — checked before anything else,
    // a closed record still has a current epoch and a late request is a stale one.
    if from_epoch != current.epoch {
        return Err(format!(
            "conflict in claim handoff: stale_request {from_epoch}; current record epoch is {}",
            current.epoch
        ));
    }
    if !is_active(&current.state) {
        return Ok(json!({
            "outcome": "unclaimed", "verdict": "not_run", "record": json!(current),
            "liveness": Value::Null,
            "notes": [format!("the record is not in an active state ({})", current.state)],
        }));
    }
    // REQ-06B-02: probe only a local-process owner on the same host.
    let probeable = current.owner.kind == "local-process"
        && current.owner.host_id.is_some()
        && current.owner.host_id.as_deref() == host_id;
    let (liveness, liveness_evidence) = if probeable {
        let (result, evidence) =
            probe_liveness(current.owner.pid, current.owner.process_start.as_deref());
        (result, Some(evidence))
    } else {
        ("not_applicable".to_string(), None)
    };
    let now = now_of(clock);
    let mut effects = current.effects.clone();
    effects.push(json!({
        "kind": "liveness", "result": liveness, "evidence": liveness_evidence, "at": now,
    }));
    // REQ-06B-03: an alive or indeterminate owner needs the explicit flag; silence,
    // timeouts and heartbeats never decide a takeover.
    if (liveness == "alive" || liveness == "indeterminate") && !force_explicit {
        let mut pending = current.clone();
        pending.state = "handoff_pending".into();
        pending.effects = effects;
        pending.last_observed_at = now.clone();
        journal.commit("handoff_pending", &pending)?;
        return Ok(json!({
            "outcome": "handoff_pending", "verdict": "blocked", "record": json!(pending),
            "liveness": liveness,
            "notes": ["an alive or indeterminate owner needs --force-explicit with a reason; no automatic takeover (REQ-06B-03)"],
        }));
    }
    // REQ-06B-04: preserve the previous worktree as evidence; read-only git only.
    let evidence = collect_evidence(&current.worktree.path);
    effects.push(json!({"kind": "worktree_evidence", "evidence": evidence, "at": now}));
    // REQ-06B-05: the worktree lock is an intent only — executing it is the caller's
    // approval scope, never this command.
    effects.push(json!({
        "kind": "worktree_lock_intent",
        "command": format!("git worktree lock --reason \"handoff from epoch {}\"", current.epoch),
        "state": "intended",
        "note": "recorded as an intent; never executed here",
        "at": now,
    }));
    let to_epoch = next_epoch(&events, key);
    let mut closed = current.clone();
    closed.state = HANDOFF_STATE.into();
    closed.last_observed_at = now.clone();
    closed.effects = effects;
    closed.effects.push(json!({
        "kind": "handoff", "to_epoch": to_epoch, "reason": reason, "at": now,
    }));
    journal.commit("handoff", &closed)?;
    Ok(json!({
        "outcome": "handed_off", "verdict": "passed", "record": json!(closed),
        "liveness": liveness, "to_epoch": to_epoch,
        "notes": ["the next claim-acquire writes the to_epoch record; publish intents below to_epoch are fenced (REQ-06B-08)"],
    }))
}

// ---- effects: intend / confirm (REQ-06B-06, REQ-06B-08) ----

/// One publish effect in the record's `effects[]`: `{"kind": <publish kind>,
/// "operation_id", "intent_digest", "state": intended|confirmed|uncertain|not_performed,
/// "epoch", "marker", "response_id"?, "observed_state"?}`. Bookkeeping entries
/// (liveness, worktree_evidence, worktree_lock_intent, stale_result, handoff) carry no
/// operation_id and are never touched here.
fn publish_effect<'a>(record: &'a ClaimRecord, operation_id: &str) -> Option<(usize, &'a Value)> {
    record
        .effects
        .iter()
        .enumerate()
        .find(|(_, e)| e["operation_id"].as_str() == Some(operation_id))
}

#[allow(clippy::too_many_arguments)]
pub fn effect_intend<C: Clock>(
    journal: &Journal,
    key: &str,
    epoch: u64,
    operation: &str,
    project_id: &str,
    local_id: &str,
    attempt: u32,
    kind: &str,
    payload: &[u8],
    clock: &C,
) -> Result<Value> {
    if !is_key(key) {
        return Err("task must be a 16 hex character claim key".into());
    }
    if !bounded_kind(kind) {
        return Err("kind must be a bounded string without control characters".into());
    }
    // The 07-B formats: both fail on non-portable parts.
    let operation_id = op_id(operation, project_id, local_id, attempt).map_err(String::from)?;
    let marker = op_marker(operation, project_id, local_id, attempt);
    let events = journal.read_events()?;
    let Some(current) = current_record(&events, key) else {
        return Err("unknown task key; no claim record exists".into());
    };
    // REQ-06B-08: the intent names the epoch it believes it is publishing under.
    if epoch < current.epoch {
        let now = now_of(clock);
        let mut diag = current.clone();
        diag.last_observed_at = now.clone();
        diag.effects.push(json!({
            "kind": "stale_result",
            "request": {"operation": operation, "project_id": project_id, "local_id": local_id,
                        "attempt": attempt, "publish_kind": kind, "epoch": epoch},
            "at": now,
        }));
        journal.commit("stale_result", &diag)?;
        return Err(format!(
            "conflict in claim journal: stale_epoch {epoch}; current record epoch is {}",
            current.epoch
        ));
    }
    if !is_active(&current.state) {
        return Err(format!(
            "conflict in claim journal: the record is not active ({});
             publish intents attach to an active epoch",
            current.state
        ));
    }
    // draft-ietf-httpapi-idempotency-key semantics, borrowed locally (never sent to
    // GitHub): the same key with the same payload is the same intent, a different payload
    // under the same key is a conflict.
    let intent_digest = digest(payload);
    if let Some((_, existing)) = publish_effect(current, &operation_id) {
        if existing["intent_digest"] == json!(intent_digest) {
            return Ok(json!({
                "outcome": "intended", "verdict": "passed",
                "operation_id": operation_id, "marker": existing["marker"],
                "intent_digest": intent_digest, "epoch": epoch, "state": existing["state"],
                "notes": ["the same key and payload were already intended; nothing changed"],
            }));
        }
        return Err(format!(
            "conflict in claim journal: operation_id {operation_id} is already recorded with a different payload"
        ));
    }
    let now = now_of(clock);
    let mut record = current.clone();
    record.last_observed_at = now.clone();
    record.effects.push(json!({
        "kind": kind,
        "operation_id": operation_id,
        "intent_digest": intent_digest,
        "state": "intended",
        "epoch": epoch,
        "marker": marker,
        "at": now,
    }));
    journal.commit("effect_intend", &record)?;
    Ok(json!({
        "outcome": "intended", "verdict": "passed",
        "operation_id": operation_id, "marker": marker,
        "intent_digest": intent_digest, "epoch": epoch, "state": "intended",
    }))
}

fn bounded_kind(s: &str) -> bool {
    !s.is_empty() && s.len() <= 64 && !s.chars().any(char::is_control)
}

/// Confirms an intended effect after the external write returned, or marks it `uncertain`
/// when the write was interrupted (REQ-06B-06). One confirm per intent.
pub fn effect_confirm<C: Clock>(
    journal: &Journal,
    key: &str,
    operation_id: &str,
    response_id: &str,
    observed_state: &str,
    clock: &C,
) -> Result<Value> {
    if !is_key(key) {
        return Err("task must be a 16 hex character claim key".into());
    }
    if response_id.len() > 256 || response_id.chars().any(char::is_control) {
        return Err("response_id must be a bounded string without control characters".into());
    }
    if !bounded_kind(observed_state) {
        return Err("observed_state must be a bounded string without control characters".into());
    }
    let events = journal.read_events()?;
    let Some(current) = current_record(&events, key) else {
        return Err("unknown task key; no claim record exists".into());
    };
    let Some((index, effect)) = publish_effect(current, operation_id) else {
        return Err(format!(
            "unknown operation_id {operation_id}; nothing was intended under it"
        ));
    };
    if effect["state"] != json!("intended") {
        return Err(format!(
            "conflict in claim journal: operation_id {operation_id} is already {}",
            effect["state"]
        ));
    }
    let now = now_of(clock);
    let mut record = current.clone();
    record.last_observed_at = now.clone();
    let mut updated = effect.clone();
    let outcome = if observed_state == "uncertain" {
        updated["state"] = json!("uncertain");
        "uncertain"
    } else {
        updated["state"] = json!("confirmed");
        updated["response_id"] = json!(response_id);
        updated["observed_state"] = json!(observed_state);
        "confirmed"
    };
    updated["at"] = json!(now);
    record.effects[index] = updated;
    journal.commit("effect_confirm", &record)?;
    Ok(json!({
        "outcome": outcome, "verdict": "passed",
        "operation_id": operation_id, "state": record.effects[index]["state"],
        "response_id": if outcome == "confirmed" { json!(response_id) } else { Value::Null },
    }))
}

// ---- reconcile (REQ-06B-07, REQ-06B-09, REQ-06B-10) ----

/// What one probe said about one uncertain effect.
#[derive(Debug, PartialEq, Eq)]
enum Probe {
    Found {
        remote_id: String,
    },
    Absent,
    /// Rate limited; `retry_after_s` carries what the probe reported.
    Limited {
        retry_after_s: Option<u64>,
    },
    /// The probe could not answer; the effect stays uncertain.
    Failed {
        why: String,
    },
}

/// Reads every comment of the issue (GET only) once — the built-in probe matches all
/// uncertain markers against this single pass, minimizing re-reads (GitHub best
/// practices). `--remote-probe` replaces it in tests and for non-comment kinds.
fn fetch_comments<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    t: &TaskRef,
) -> std::result::Result<Vec<(i64, String)>, ForgeError> {
    let mut url = Some(format!(
        "https://api.github.com/repos/{}/{}/issues/{}/comments?per_page=100",
        t.owner, t.repo, t.issue_number
    ));
    let mut comments = Vec::new();
    let mut pages = 0;
    while let Some(u) = url {
        pages += 1;
        if pages > 50 {
            return Err(ForgeError {
                kind: "invalid_input",
                message: "reconcile comment page limit exceeded (50)".into(),
                retryable: false,
            });
        }
        let resp = client.get(&u, None)?;
        let arr = resp
            .body
            .as_ref()
            .and_then(Value::as_array)
            .cloned()
            .unwrap_or_default();
        for c in &arr {
            if let (Some(id), Some(body)) = (c["id"].as_i64(), c["body"].as_str()) {
                comments.push((id, body.to_string()));
            }
        }
        url = resp.header("link").and_then(|l| next_link(Some(l)));
    }
    Ok(comments)
}

fn run_probe_command(cmd: &str, marker: &str) -> Probe {
    let out = Command::new(cmd)
        .arg(marker)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .output();
    let Ok(out) = out else {
        return Probe::Failed {
            why: format!("probe {cmd} cannot be started"),
        };
    };
    let stdout = String::from_utf8_lossy(&out.stdout).trim().to_string();
    match out.status.code() {
        Some(0) => match serde_json::from_str::<Value>(&stdout) {
            Ok(mut v) => match v["remote_id"].take() {
                Value::String(s) if !s.is_empty() => Probe::Found { remote_id: s },
                Value::Number(n) => Probe::Found {
                    remote_id: n.to_string(),
                },
                _ => Probe::Failed {
                    why: "exit 0 without a remote_id".into(),
                },
            },
            Err(_) => Probe::Failed {
                why: "exit 0 with a non-JSON body".into(),
            },
        },
        Some(4) => Probe::Absent,
        Some(5) => Probe::Limited {
            retry_after_s: stdout.parse().ok(),
        },
        code => Probe::Failed {
            why: format!("probe exited {code:?}"),
        },
    }
}

/// Reconciles every `uncertain` publish effect against the remote. Reads only — the
/// function has no write path at all, so nothing is ever re-created (REQ-06B-07). The
/// journal is written once, only when at least one effect moved; a rate-limited probe
/// leaves every state unchanged (REQ-06B-09).
pub fn effects_reconcile<T: Transport, C: Clock>(
    journal: &Journal,
    key: &str,
    remote_probe: Option<&str>,
    transport: &T,
    clock: &C,
) -> Result<Value> {
    if !is_key(key) {
        return Err("task must be a 16 hex character claim key".into());
    }
    let events = journal.read_events()?;
    let Some(current) = current_record(&events, key).cloned() else {
        return Err("unknown task key; no claim record exists".into());
    };
    let uncertain: Vec<usize> = current
        .effects
        .iter()
        .enumerate()
        .filter(|(_, e)| e["state"] == json!("uncertain") && e["operation_id"].is_string())
        .map(|(i, _)| i)
        .collect();
    if uncertain.is_empty() {
        return Ok(json!({
            "outcome": "reconciled", "verdict": "not_run", "transitions": [],
            "notes": ["no uncertain effects to reconcile; nothing was probed and nothing written"],
        }));
    }
    let mut record = current.clone();
    let mut transitions = Vec::new();
    let mut retry_after = Vec::new();
    let mut blocked = false;
    let mut failed = 0;
    // The built-in probe fetches the comments once for every uncertain effect.
    let comments = if remote_probe.is_none() {
        let client = Client::new(transport, clock, Default::default(), Default::default());
        match fetch_comments(&client, &record.task_ref) {
            Ok(c) => Ok(c),
            Err(e) if e.kind == "blocked" => {
                blocked = true;
                Err(e)
            }
            Err(e) => Err(e),
        }
    } else {
        Ok(Vec::new())
    };
    for i in uncertain {
        let effect = current.effects[i].clone();
        let marker = effect["marker"].as_str().unwrap_or_default().to_string();
        let operation_id = effect["operation_id"]
            .as_str()
            .unwrap_or_default()
            .to_string();
        let verdict = match remote_probe {
            Some(cmd) => run_probe_command(cmd, &marker),
            None => match &comments {
                Err(e) if e.kind == "blocked" => Probe::Limited {
                    retry_after_s: None,
                },
                Err(e) => Probe::Failed {
                    why: e.message.clone(),
                },
                Ok(comments) => match comments.iter().find(|(_, body)| body.contains(&marker)) {
                    Some((id, _)) => Probe::Found {
                        remote_id: id.to_string(),
                    },
                    None => Probe::Absent,
                },
            },
        };
        match verdict {
            Probe::Found { remote_id } => {
                let mut updated = effect.clone();
                updated["state"] = json!("confirmed");
                updated["response_id"] = json!(remote_id.clone());
                updated["observed_state"] = json!("marker found on the remote");
                updated["at"] = json!(now_of(clock));
                record.effects[i] = updated.clone();
                transitions.push(json!({
                    "operation_id": operation_id, "from": "uncertain",
                    "to": "confirmed", "response_id": remote_id,
                }));
            }
            Probe::Absent => {
                let mut updated = effect.clone();
                updated["state"] = json!("not_performed");
                updated["observed_state"] = json!("marker confirmed absent");
                updated["at"] = json!(now_of(clock));
                record.effects[i] = updated.clone();
                transitions.push(json!({
                    "operation_id": operation_id, "from": "uncertain", "to": "not_performed",
                }));
            }
            Probe::Limited { retry_after_s } => {
                blocked = true;
                if let Some(s) = retry_after_s {
                    retry_after.push(s);
                }
                transitions.push(json!({
                    "operation_id": operation_id, "from": "uncertain", "to": "uncertain",
                    "why": "rate limited; state unchanged until retry-after (REQ-06B-09)",
                }));
            }
            Probe::Failed { why } => {
                failed += 1;
                transitions.push(json!({
                    "operation_id": operation_id, "from": "uncertain", "to": "uncertain",
                    "why": format!("probe failed: {why}"),
                }));
            }
        }
    }
    // One event only when something actually moved; a second identical run writes nothing.
    let moved = transitions.iter().any(|t| t["to"] != json!("uncertain"));
    if moved {
        record.last_observed_at = now_of(clock);
        journal.commit("effects_reconcile", &record)?;
    }
    let verdict = if blocked {
        "blocked"
    } else if failed > 0 {
        "uncertain"
    } else {
        "passed"
    };
    let mut notes = Vec::new();
    if blocked {
        notes.push(
            "a probe was rate limited; effects keep their state until retry-after (REQ-06B-09)"
                .into(),
        );
    }
    if !retry_after.is_empty() {
        notes.push(format!(
            "retry_after_s reported by the probe: {:?}",
            retry_after
        ));
    }
    Ok(json!({
        "outcome": "reconciled", "verdict": verdict,
        "transitions": transitions, "notes": notes,
    }))
}

/// The `moai:op` markers found in a remote body, re-exported for callers that parse
/// comment bodies directly (the parse itself is 07-B's `parse_op_markers`).
pub fn markers_in(body: &str) -> Vec<crate::forge::outbox::OpMarker> {
    parse_op_markers(body)
}

// ---- CLI entries (`main` wraps the results in the 01-C envelope) ----

fn open_journal(project: &Path) -> Result<Journal> {
    let root = crate::claim::journal_root(project)?;
    let journal = Journal::open(&root)?;
    if !journal.try_lock()? {
        return Err("conflict in claim journal: another process holds the journal lock; the waiting policy is the caller's (REQ-06A-05)".into());
    }
    Ok(journal)
}

pub fn claim_handoff_cli<C: Clock>(
    project: &Path,
    key: &str,
    from_epoch: u64,
    reason: &str,
    force_explicit: bool,
    host_id: Option<&str>,
    clock: &C,
) -> Result<Value> {
    let journal = open_journal(project)?;
    claim_handoff(
        &journal,
        key,
        from_epoch,
        reason,
        force_explicit,
        host_id,
        clock,
    )
}

#[allow(clippy::too_many_arguments)]
pub fn effect_intend_cli<C: Clock>(
    project: &Path,
    key: &str,
    epoch: u64,
    operation: &str,
    project_id: &str,
    local_id: &str,
    attempt: u32,
    kind: &str,
    payload: &[u8],
    clock: &C,
) -> Result<Value> {
    let journal = open_journal(project)?;
    effect_intend(
        &journal, key, epoch, operation, project_id, local_id, attempt, kind, payload, clock,
    )
}

pub fn effect_confirm_cli<C: Clock>(
    project: &Path,
    key: &str,
    operation_id: &str,
    response_id: &str,
    observed_state: &str,
    clock: &C,
) -> Result<Value> {
    let journal = open_journal(project)?;
    effect_confirm(
        &journal,
        key,
        operation_id,
        response_id,
        observed_state,
        clock,
    )
}

pub fn effects_reconcile_cli<T: Transport, C: Clock>(
    project: &Path,
    key: &str,
    remote_probe: Option<&str>,
    transport: &T,
    clock: &C,
) -> Result<Value> {
    let journal = open_journal(project)?;
    effects_reconcile(&journal, key, remote_probe, transport, clock)
}
