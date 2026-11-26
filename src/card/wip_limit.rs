//! The limit of cards in work and in review (SPEC-20261125-landing-moves-checkout REQ-1 to
//! REQ-4, user decision 2026-11-24T21:52:39Z): the `active` and `review` cards may number at
//! most `wip_limit` (profile key `work_flow.wip_limit`, default [`DEFAULT_PER_LANE`]) times the
//! live registered lanes. Only a card whose reservation of the current generation has a lease that
//! has not ended is counted (SPEC-20261125-wip-limit-live-leases, user decision 2026-11-25); the
//! others are listed apart in a refusal and in `card next`. At the limit a session helps with a
//! check or a landing instead of
//! taking a card, so `card take` and `card add` refuse and `card next` names the cards that wait
//! for an audit. `wip_limit` 0 turns the limit off, and so does a project with no registry
//! folder or no live lane (SPEC D-3). A lane is live while its launcher holds the `<name>.live`
//! lock or its registration lease runs (SPEC-20261125-wip-limit-live-lock-lanes); a registry
//! folder that cannot be read, or whose items none can be read
//! (SPEC-20261126-wip-limit-all-unreadable), does not turn the limit off: the limit is then
//! reached.

use crate::lane_registry;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::path::Path;

/// The cards a lane may have in work or in review when the profile names no number.
pub const DEFAULT_PER_LANE: u32 = 2;

/// The actor whose adds the limit never refuses: the user's own.
pub const USER: &str = "user";

fn now_secs() -> i64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map_or(0, |d| d.as_secs() as i64)
}

/// The multiple per lane: the profile's `work_flow.wip_limit`, [`DEFAULT_PER_LANE`] when it names
/// none or the profile cannot be read (a profile problem never stops a take).
pub fn per_lane(project: &Path) -> u32 {
    crate::profile::read_profile(project)
        .ok()
        .and_then(|(profile, _)| profile.work_flow.and_then(|w| w.wip_limit))
        .unwrap_or(DEFAULT_PER_LANE)
}

/// The live lanes of the session registry and what could not be read of it.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct Lanes {
    /// The live lanes; 0 when the registry does not exist, which turns the limit off.
    pub live: usize,
    /// The names whose entry or lock could not be read; each is left out of `live`.
    pub unreadable: Vec<String>,
    /// Why the registry folder could not be read as a whole, or why none of its items could be
    /// read (SPEC-20261126-wip-limit-all-unreadable REQ-1); the limit is then not off but
    /// reached, so that nothing is taken past a limit that cannot be checked
    /// (SPEC-20261125-wip-limit-live-lock-lanes REQ-4).
    pub registry_error: Option<String>,
}

/// The live lanes (SPEC-20261125-wip-limit-live-lock-lanes REQ-1 to REQ-5): every lane whose
/// launcher holds the `<name>.live` lock or whose registry entry has a lease that has not ended,
/// each lane once, the leaders left out. A name with a held lock and no readable entry counts
/// only when it is named like a lane of this project (`agent-…` or `<project name>-agent-…`): a
/// lock alone does not say what the session is. This is the rule of `session-register
/// --lane-check`, so the limit and the leader's message check agree.
///
/// A lane is counted once whatever its files are (SPEC-20261125-session-registry-session-key
/// REQ-6): the two name forms `agent-<name>` and `<project name>-agent-<name>` are one name, a
/// held lock makes the entries of its name one lane, and entries that carry one host session id
/// are one lane.
///
/// A registry that holds items of which none can be read (no entry, no held lock) answers with a
/// `registry_error` naming them, so the limit is reached and not off
/// (SPEC-20261126-wip-limit-all-unreadable REQ-1).
pub fn live_lanes(project: &Path) -> Lanes {
    let scan = match lane_registry::scan_registry(project) {
        Ok(Some(scan)) => scan,
        Ok(None) => return Lanes::default(),
        Err(why) => {
            return Lanes {
                registry_error: Some(why),
                ..Lanes::default()
            }
        }
    };
    let now = now_secs();
    // One project name for both: each `session::*_name` call reads the git identity again, and a
    // git process costs far more than the registry scan. The forms mirror
    // `session::leader_session_name` and `session::agent_prefix`; a unit test pins them.
    let project_name = crate::session::project_name(project);
    let leader_name = format!("{project_name}-leader");
    let project_prefix = format!("{project_name}-agent-");
    let lane_prefixes = ["agent-".to_string(), project_prefix.clone()];
    // The one name of a lane: the project-prefixed form reads as the short form.
    let canon = |name: &str| match name.strip_prefix(project_prefix.as_str()) {
        Some(rest) => format!("agent-{rest}"),
        None => name.to_string(),
    };
    let locked: std::collections::BTreeSet<String> = scan.locked.iter().map(|n| canon(n)).collect();
    // `(true, session id)` for an entry that carries one, `(false, name)` for a lane known by its
    // name (an entry without a session, or a name whose lock is held).
    let mut live: std::collections::BTreeSet<(bool, String)> = std::collections::BTreeSet::new();
    for entry in &scan.entries {
        let leader = entry.mode.as_deref() == Some("leader") || entry.name == leader_name;
        let name = canon(&entry.name);
        let by_lock = locked.contains(&name);
        if leader || !(entry.lease_until > now || by_lock) {
            continue;
        }
        live.insert(match (&entry.session, by_lock) {
            (Some(session), false) => (true, session.clone()),
            _ => (false, name),
        });
    }
    for name in &scan.locked {
        let name_canon = canon(name);
        let has_entry = scan.entries.iter().any(|e| canon(&e.name) == name_canon);
        let lane_like = lane_prefixes.iter().any(|p| name.starts_with(p.as_str()));
        if !has_entry && *name != leader_name && lane_like {
            live.insert((false, name_canon));
        }
    }
    let unreadable: Vec<String> = scan.unreadable.iter().cloned().collect();
    // A registry whose every item is unreadable is no registry that says "no lane": the limit
    // cannot be checked, so it is reached (SPEC-20261126-wip-limit-all-unreadable REQ-1).
    let registry_error =
        (scan.entries.is_empty() && scan.locked.is_empty() && !unreadable.is_empty()).then(|| {
            format!(
                "the sessions folder holds {} item(s) and none can be read: {}",
                unreadable.len(),
                unreadable.join(", ")
            )
        });
    Lanes {
        live: live.len(),
        unreadable,
        registry_error,
    }
}

/// How the reservation of a card's current generation stands (SPEC-20261125-wip-limit-live-leases
/// REQ-1): only a live lease makes a card live work that the limit counts.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Lease {
    /// Held, or held for its review, with a lease that has not ended.
    Live,
    /// A reservation whose lease ended at this time (Unix seconds).
    Ended(i64),
    /// No reservation of the current generation, or one that was released or reclaimed.
    NoReservation,
}

impl Lease {
    /// The word of a refusal and of the `card next` answer.
    pub fn reason(self) -> &'static str {
        match self {
            Lease::Live => "live lease",
            Lease::Ended(_) => "lease ended",
            Lease::NoReservation => "no reservation",
        }
    }
}

/// A card in work or in review.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct InWork {
    pub card_id: i64,
    pub state: String,
    pub title: String,
    pub lease: Lease,
}

impl InWork {
    /// The text of a card left out of the count: its name, state, title and why it is left out.
    fn left_out_text(&self) -> String {
        let why = match self.lease {
            Lease::Ended(at) => format!("lease ended {}", crate::forge::policy::format_utc(at)),
            other => other.reason().to_string(),
        };
        format!(
            "C-{} ({}, {}, {})",
            self.card_id, self.state, self.title, why
        )
    }
}

/// Every `active` and `review` card, oldest first, with the state of its current reservation read
/// against the ledger's own clock (the clock the other lease checks use).
pub fn in_work(conn: &Connection) -> Result<Vec<InWork>, String> {
    let mut stmt = conn
        .prepare(
            "SELECT c.card_id, c.state, c.title, r.status, r.lease_until, unixepoch() \
             FROM card c \
             LEFT JOIN reservation r ON r.card_id = c.card_id AND r.generation = c.generation \
             WHERE c.state IN ('active', 'review') ORDER BY c.card_id",
        )
        .map_err(|e| e.to_string())?;
    let rows = stmt
        .query_map([], |r| {
            let status: Option<String> = r.get(3)?;
            let lease_until: Option<i64> = r.get(4)?;
            let now: i64 = r.get(5)?;
            let lease = match (status.as_deref(), lease_until) {
                (Some("held" | "held_for_review"), Some(until)) if until > now => Lease::Live,
                (Some("held" | "held_for_review" | "expired_candidate"), Some(until)) => {
                    Lease::Ended(until)
                }
                _ => Lease::NoReservation,
            };
            Ok(InWork {
                card_id: r.get(0)?,
                state: r.get(1)?,
                title: r.get(2)?,
                lease,
            })
        })
        .map_err(|e| e.to_string())?;
    rows.collect::<Result<_, _>>().map_err(|e| e.to_string())
}

/// The limit now: the numbers and whether the cards in work reach it.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Reading {
    pub per_lane: u32,
    pub lanes: usize,
    /// `per_lane` times `lanes`; 0 when the limit is off.
    pub limit: usize,
    /// The cards counted: in work or in review with a live lease.
    pub count: usize,
    pub cards: Vec<InWork>,
    /// The cards in work or in review that the count leaves out: an ended lease, or no
    /// reservation (REQ-3).
    pub not_counted: Vec<InWork>,
    /// The names whose registry entry or lock could not be read (REQ-5).
    pub unreadable: Vec<String>,
    /// Why the registry folder could not be read as a whole (REQ-4); the limit is then reached.
    pub registry_error: Option<String>,
}

impl Reading {
    /// Reached: the cards counted fill `per_lane` times the live lanes, or the registry folder
    /// cannot be read, so the limit cannot be checked. Off with `wip_limit` 0, with no registry
    /// folder and with no live lane (REQ-8).
    pub fn at_limit(&self) -> bool {
        self.per_lane != 0
            && (self.registry_error.is_some() || (self.lanes != 0 && self.count >= self.limit))
    }

    pub fn json(&self) -> Value {
        let left_out: Vec<Value> = self
            .not_counted
            .iter()
            .map(|c| {
                let until = match c.lease {
                    Lease::Ended(at) => json!(at),
                    _ => Value::Null,
                };
                json!({
                    "name": format!("C-{}", c.card_id), "card_id": c.card_id, "state": c.state,
                    "reason": c.lease.reason(), "lease_until": until,
                })
            })
            .collect();
        json!({
            "per_lane": self.per_lane, "lanes": self.lanes, "limit": self.limit,
            "count": self.count, "at_limit": self.at_limit(), "not_counted": left_out,
            "registry_error": self.registry_error, "unreadable": self.unreadable,
        })
    }
}

/// Reads the limit and the cards in work: the cards with a live lease are counted, the others are
/// listed apart (SPEC-20261125-wip-limit-live-leases REQ-1 and REQ-3).
pub fn read(project: &Path, conn: &Connection) -> Result<Reading, String> {
    let per_lane = per_lane(project);
    let found = live_lanes(project);
    let lanes = found.live;
    let (cards, not_counted): (Vec<InWork>, Vec<InWork>) = in_work(conn)?
        .into_iter()
        .partition(|c| c.lease == Lease::Live);
    Ok(Reading {
        per_lane,
        lanes,
        limit: (per_lane as usize).saturating_mul(lanes),
        count: cards.len(),
        cards,
        not_counted,
        unreadable: found.unreadable,
        registry_error: found.registry_error,
    })
}

/// The refusal of `action` (`take` or `add`) at the limit, `None` below it or with the limit off.
/// The text names the cards that wait, so the refused session knows what to help with.
pub fn refusal(project: &Path, conn: &Connection, action: &str) -> Result<Option<String>, String> {
    let reading = read(project, conn)?;
    if !reading.at_limit() {
        return Ok(None);
    }
    // A registry folder that cannot be read is no reason to take past the limit: say why and stop
    // (REQ-4). `wip_limit` 0 is the way past it.
    if let Some(why) = &reading.registry_error {
        return Ok(Some(format!(
            "card conflict in wip_limit: card {action} refused: the session registry cannot be \
             read, so the work-in-progress limit cannot be checked: {why}; fix the registry \
             folder, or set work_flow.wip_limit to 0 in the project profile to turn the limit off"
        )));
    }
    let waiting: Vec<String> = reading
        .cards
        .iter()
        .map(|c| format!("C-{} ({}, {})", c.card_id, c.state, c.title))
        .collect();
    // The cards the count leaves out, apart from the counted ones (REQ-3, REQ-4).
    let left_out = if reading.not_counted.is_empty() {
        String::new()
    } else {
        let texts: Vec<String> = reading
            .not_counted
            .iter()
            .map(InWork::left_out_text)
            .collect();
        format!("; not counted (no live lease): {}", texts.join("; "))
    };
    // The registry names that could not be read and so are left out of the lane count (REQ-5).
    let left_out = if reading.unreadable.is_empty() {
        left_out
    } else {
        format!(
            "{left_out}; unreadable (left out of the lane count): {}",
            reading.unreadable.join(", ")
        )
    };
    Ok(Some(format!(
        "card conflict in wip_limit: card {action} refused: the work-in-progress limit is reached: \
         {} cards are in work or in review and the limit is {} ({} per lane, {} lane(s)); help with \
         a check or a landing first: {}{left_out}",
        reading.count,
        reading.limit,
        reading.per_lane,
        reading.lanes,
        waiting.join("; ")
    )))
}

/// A card in `review` that no verdict names and that `holder` did not hold: the checks a session
/// can take up at the limit.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct AuditNeeded {
    pub card_id: i64,
    pub title: String,
}

/// The `review` cards without a `card.verdict` event whose current reservation is not held by
/// `holder`, oldest first.
pub fn audit_needed(conn: &Connection, holder: Option<&str>) -> Result<Vec<AuditNeeded>, String> {
    let mut stmt = conn
        .prepare(
            "SELECT c.card_id, c.title FROM card c \
             WHERE c.state = 'review' \
               AND NOT EXISTS (SELECT 1 FROM event e \
                               WHERE e.card_id = c.card_id AND e.kind = 'card.verdict') \
               AND NOT EXISTS (SELECT 1 FROM reservation r \
                               WHERE r.card_id = c.card_id AND r.generation = c.generation \
                                 AND r.holder = ?1) \
             ORDER BY c.card_id",
        )
        .map_err(|e| e.to_string())?;
    let rows = stmt
        .query_map([holder.unwrap_or("")], |r| {
            Ok(AuditNeeded {
                card_id: r.get(0)?,
                title: r.get(1)?,
            })
        })
        .map_err(|e| e.to_string())?;
    rows.collect::<Result<_, _>>().map_err(|e| e.to_string())
}
