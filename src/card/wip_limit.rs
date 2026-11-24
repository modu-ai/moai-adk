//! The limit of cards in work and in review (SPEC-20261125-landing-moves-checkout REQ-1 to
//! REQ-4, user decision 2026-11-24T21:52:39Z): the `active` and `review` cards may number at
//! most `wip_limit` (profile key `work_flow.wip_limit`, default [`DEFAULT_PER_LANE`]) times the
//! live registered lanes, at least one lane. At the limit a session helps with a check or a
//! landing instead of taking a card, so `card take` and `card add` refuse and `card next` names
//! the cards that wait for an audit. `wip_limit` 0 turns the limit off.

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

/// The live registered lanes: entries whose lease has not ended and that are no leader entry; at
/// least 1, also when the registry is empty or unreadable.
pub fn live_lanes(project: &Path) -> usize {
    let now = now_secs();
    lane_registry::session_entries(project)
        .map(|entries| {
            entries
                .iter()
                .filter(|e| e.lease_until > now && e.mode.as_deref() != Some("leader"))
                .count()
        })
        .unwrap_or(0)
        .max(1)
}

/// A card in work or in review.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct InWork {
    pub card_id: i64,
    pub state: String,
    pub title: String,
}

/// Every `active` and `review` card, oldest first.
pub fn in_work(conn: &Connection) -> Result<Vec<InWork>, String> {
    let mut stmt = conn
        .prepare(
            "SELECT card_id, state, title FROM card \
             WHERE state IN ('active', 'review') ORDER BY card_id",
        )
        .map_err(|e| e.to_string())?;
    let rows = stmt
        .query_map([], |r| {
            Ok(InWork {
                card_id: r.get(0)?,
                state: r.get(1)?,
                title: r.get(2)?,
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
    pub count: usize,
    pub cards: Vec<InWork>,
}

impl Reading {
    pub fn at_limit(&self) -> bool {
        self.per_lane != 0 && self.count >= self.limit
    }

    pub fn json(&self) -> Value {
        json!({
            "per_lane": self.per_lane, "lanes": self.lanes, "limit": self.limit,
            "count": self.count, "at_limit": self.at_limit(),
        })
    }
}

/// Reads the limit and the cards in work.
pub fn read(project: &Path, conn: &Connection) -> Result<Reading, String> {
    let per_lane = per_lane(project);
    let lanes = live_lanes(project);
    let cards = in_work(conn)?;
    Ok(Reading {
        per_lane,
        lanes,
        limit: (per_lane as usize).saturating_mul(lanes),
        count: cards.len(),
        cards,
    })
}

/// The refusal of `action` (`take` or `add`) at the limit, `None` below it or with the limit off.
/// The text names the cards that wait, so the refused session knows what to help with.
pub fn refusal(project: &Path, conn: &Connection, action: &str) -> Result<Option<String>, String> {
    let reading = read(project, conn)?;
    if !reading.at_limit() {
        return Ok(None);
    }
    let waiting: Vec<String> = reading
        .cards
        .iter()
        .map(|c| format!("C-{} ({}, {})", c.card_id, c.state, c.title))
        .collect();
    Ok(Some(format!(
        "card conflict in wip_limit: card {action} refused: the work-in-progress limit is reached: \
         {} cards are in work or in review and the limit is {} ({} per lane, {} lane(s)); help with \
         a check or a landing first: {}",
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
