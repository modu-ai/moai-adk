//! Outbox and operation markers (07-B #2009; outbox and marker rules taken over from 07-C #2010).
//!
//! GitHub's Issues API has no idempotency key, so an external write is at-least-once at best.
//! Each write therefore leaves a durable record *before* it is sent and a marker in the
//! remote body, and a later run reconciles by reading the remote back:
//!
//! - `<project>/.moai-cli/forge/outbox/<op-id>.json` holds one operation with three phases,
//!   always written in this order: `intent` (before the request), `response` (what came back,
//!   if anything), `observed` (what a fresh GET confirmed). Files are replaced by temporary
//!   file + rename.
//! - States: `proposed` → `publishing` → `confirmed` | `publish_unknown` | `blocked`, and
//!   `conflict` when the remote disagrees with the intent. `publish_unknown` never goes back to
//!   `publishing` by itself: only a reconcile that finds the marker moves it to `confirmed`.
//! - Markers are HTML comments, invisible in rendered Markdown:
//!   `<!-- moai:op <operation>:<project_id>:<local_id>:<attempt> -->` and
//!   `<!-- moai:report run=<run_id> -->` (07-C report comments). `moai-op` is accepted when
//!   parsing, for the spelling 07-C's design also uses.
use super::ForgeError;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::{
    fs::{self, OpenOptions},
    io::Write,
    path::{Path, PathBuf},
};

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum OutboxState {
    Proposed,
    Publishing,
    Confirmed,
    PublishUnknown,
    Conflict,
    Blocked,
}

impl OutboxState {
    /// The only transitions an outbox file may make.
    pub fn may_become(self, next: OutboxState) -> bool {
        use OutboxState::*;
        matches!(
            (self, next),
            (Proposed, Publishing)
                | (Proposed, Blocked)
                | (Publishing, Confirmed)
                | (Publishing, PublishUnknown)
                | (Publishing, Blocked)
                | (Publishing, Conflict)
                // Reconcile only: a marker found on the remote confirms an unknown publish.
                | (PublishUnknown, Confirmed)
                | (PublishUnknown, Conflict)
        )
    }
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Phase {
    pub at: String,
    #[serde(default)]
    pub detail: Value,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct OutboxEntry {
    pub schema_version: u32,
    pub op_id: String,
    pub marker: String,
    pub state: OutboxState,
    pub intent: Option<Phase>,
    pub response: Option<Phase>,
    pub observed: Option<Phase>,
}

impl OutboxEntry {
    pub fn new(op_id: &str, marker: &str) -> Self {
        OutboxEntry {
            schema_version: 1,
            op_id: op_id.into(),
            marker: marker.into(),
            state: OutboxState::Proposed,
            intent: None,
            response: None,
            observed: None,
        }
    }

    /// Moves to `next`, recording `phase` in the slot that state implies. Refuses illegal
    /// transitions and any phase out of the intent → response → observed order.
    pub fn advance(&mut self, next: OutboxState, phase: Phase) -> Result<(), ForgeError> {
        if !self.state.may_become(next) {
            return Err(ForgeError::of("internal", "outbox state machine violation"));
        }
        match next {
            OutboxState::Publishing => self.intent = Some(phase),
            OutboxState::PublishUnknown | OutboxState::Blocked if self.intent.is_some() => {
                self.response = Some(phase)
            }
            OutboxState::Blocked => self.intent = Some(phase),
            OutboxState::Confirmed | OutboxState::Conflict => {
                if self.intent.is_none() {
                    return Err(ForgeError::of("internal", "outbox state machine violation"));
                }
                self.observed = Some(phase)
            }
            _ => {}
        }
        self.state = next;
        Ok(())
    }

    /// Records what came back while still `publishing` (a 2xx before it is confirmed).
    pub fn record_response(&mut self, phase: Phase) -> Result<(), ForgeError> {
        if self.state != OutboxState::Publishing {
            return Err(ForgeError::of("internal", "outbox state machine violation"));
        }
        self.response = Some(phase);
        Ok(())
    }
}

fn valid_part(s: &str) -> bool {
    !s.is_empty()
        && s.len() <= 64
        && s.bytes()
            .all(|b| b.is_ascii_alphanumeric() || b"-_.".contains(&b))
        && s != "."
        && s != ".."
}

/// `<!-- moai:op <operation>:<project_id>:<local_id>:<attempt> -->`.
pub fn op_marker(operation: &str, project_id: &str, local_id: &str, attempt: u32) -> String {
    format!("<!-- moai:op {operation}:{project_id}:{local_id}:{attempt} -->")
}

/// `<!-- moai:report run=<run_id> -->`.
pub fn report_marker(run_id: &str) -> String {
    format!("<!-- moai:report run={run_id} -->")
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct OpMarker {
    pub operation: String,
    pub project_id: String,
    pub local_id: String,
    pub attempt: u32,
}

/// Every operation marker in `text` (`moai:op` or `moai-op`); malformed ones are skipped.
pub fn parse_op_markers(text: &str) -> Vec<OpMarker> {
    let mut out = Vec::new();
    let mut rest = text;
    while let Some(start) = rest.find("<!--") {
        let after = &rest[start + 4..];
        let Some(end) = after.find("-->") else { break };
        let inner = after[..end].trim();
        rest = &after[end + 3..];
        let Some(body) = inner
            .strip_prefix("moai:op ")
            .or_else(|| inner.strip_prefix("moai-op "))
        else {
            continue;
        };
        let parts: Vec<&str> = body.trim().split(':').collect();
        if let [op, project, local, attempt] = parts[..] {
            if [op, project, local].iter().all(|p| valid_part(p)) {
                if let Ok(attempt) = attempt.parse() {
                    out.push(OpMarker {
                        operation: op.into(),
                        project_id: project.into(),
                        local_id: local.into(),
                        attempt,
                    });
                }
            }
        }
    }
    out
}

/// `<operation>-<project_id>-<local_id>-<attempt>`; every part is a portable name.
pub fn op_id(
    operation: &str,
    project_id: &str,
    local_id: &str,
    attempt: u32,
) -> Result<String, ForgeError> {
    if ![operation, project_id, local_id]
        .iter()
        .all(|p| valid_part(p))
    {
        return Err(ForgeError::invalid(
            "invalid operation id: parts must be portable names",
        ));
    }
    Ok(format!("{operation}-{project_id}-{local_id}-{attempt}"))
}

fn outbox_dir(project: &Path, create: bool) -> Result<PathBuf, ForgeError> {
    // #2175: the P09 opt-in home layout switches the outbox store to
    // ~/.moai-cli/projects/<id>/forge/outbox once layout.json is in place.
    if let Some(home) = crate::layout::switched_dir(project, "forge/outbox") {
        if create {
            std::fs::create_dir_all(&home)
                .map_err(|_| ForgeError::io("outbox directory unavailable"))?;
        }
        return Ok(home);
    }
    if !create {
        return Ok(crate::state_root::path(project, "forge/outbox"));
    }
    let mut dir = project.to_path_buf();
    for part in [crate::state_root::dir(project), "forge", "outbox"] {
        dir.push(part);
        match fs::symlink_metadata(&dir) {
            Ok(m) if m.file_type().is_symlink() || !m.is_dir() => {
                return Err(ForgeError::invalid(
                    "invalid outbox directory: a symlink or not a directory; refused",
                ))
            }
            Ok(_) => {}
            Err(_) => {
                fs::create_dir(&dir).map_err(|_| ForgeError::io("outbox directory unavailable"))?
            }
        }
    }
    Ok(dir)
}

pub fn load(project: &Path, op_id: &str) -> Result<Option<OutboxEntry>, ForgeError> {
    let path = outbox_dir(project, false)?.join(format!("{op_id}.json"));
    match fs::read(&path) {
        Ok(bytes) => serde_json::from_slice(&bytes)
            .map(Some)
            .map_err(|_| ForgeError::io("outbox entry unreadable")),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(None),
        Err(_) => Err(ForgeError::io("outbox entry unreadable")),
    }
}

/// Temporary file + rename in the outbox directory.
pub fn save(project: &Path, entry: &OutboxEntry) -> Result<(), ForgeError> {
    let dir = outbox_dir(project, true)?;
    let path = dir.join(format!("{}.json", entry.op_id));
    let tmp = dir.join(format!(".{}.json.tmp-{}", entry.op_id, std::process::id()));
    let bytes = serde_json::to_vec_pretty(entry)
        .map_err(|_| ForgeError::of("internal", "encode failed"))?;
    let mut f = OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&tmp)
        .map_err(|_| ForgeError::io("outbox write failed"))?;
    f.write_all(&bytes)
        .and_then(|_| f.sync_all())
        .map_err(|_| ForgeError::io("outbox write failed"))?;
    fs::rename(&tmp, &path).map_err(|_| {
        let _ = fs::remove_file(&tmp);
        ForgeError::io("outbox replace failed")
    })
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    fn phase() -> Phase {
        Phase {
            at: "2026-11-08T00:00:00Z".into(),
            detail: json!({}),
        }
    }

    #[test]
    fn transitions_follow_intent_response_observed() {
        let mut e = OutboxEntry::new("promote-p-t1-1", &op_marker("promote", "p", "t1", 1));
        assert!(
            e.advance(OutboxState::Confirmed, phase()).is_err(),
            "no observed before intent"
        );
        e.advance(OutboxState::Publishing, phase()).unwrap();
        e.advance(OutboxState::PublishUnknown, phase()).unwrap();
        assert!(e.response.is_some() && e.observed.is_none());
        assert!(
            e.advance(OutboxState::Publishing, phase()).is_err(),
            "no silent retry"
        );
        e.advance(OutboxState::Confirmed, phase()).unwrap();
        assert!(e.observed.is_some());
        assert!(e.advance(OutboxState::Publishing, phase()).is_err());
    }

    #[test]
    fn markers_round_trip_and_reject_junk() {
        let m = op_marker("promote", "demo", "T-1", 2);
        assert_eq!(m, "<!-- moai:op promote:demo:T-1:2 -->");
        let text = format!("body\n\n{m}\n<!-- moai-op promote:demo:T-2:1 -->\n<!-- moai:op a:b -->\n<!-- moai:op x:../y:z:1 -->");
        let found = parse_op_markers(&text);
        assert_eq!(found.len(), 2);
        assert_eq!((found[0].local_id.as_str(), found[0].attempt), ("T-1", 2));
        assert_eq!(found[1].local_id, "T-2");
        assert_eq!(report_marker("r1"), "<!-- moai:report run=r1 -->");
        assert!(op_id("promote", "demo", "../x", 1).is_err());
    }
}
