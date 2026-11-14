//! `moai leader-round` (Issue #2902 O3b, plan 6875163101): the pure part. A leader observes
//! each lane's terminal (Orca `terminal read --screen --json`) and writes what it saw into an
//! observations file; this module reads that file, finds the lane's `LANE-*` lines and picks
//! one action per lane. It runs no git and calls no forge: the `lane-safe-to-reset` verdict is
//! an argument, and the caller fills the `release` slot of a `clear_redispatch`.
//!
//! The turn, the input box and the context percentage are what the observer wrote; nothing
//! here guesses them from the screen (the `lane_status` principle).

use crate::lane_reset::{ClaimsObservation, CwdObservation};
use crate::lane_status::{PaneState, MAX_LANES};
use serde::{Deserialize, Serialize};

/// The observations file version this module reads.
pub const SCHEMA_VERSION: u32 = 1;
/// Oldest observation, in seconds, a decision may rest on (the `lane-safe-to-reset` limit).
pub const MAX_AGE_SECS: i64 = crate::lane_reset::SCREEN_MAX_AGE_SECS;
/// From this context use, a `LANE-REPORT` is a card boundary too.
pub const CLEAR_CONTEXT_PCT: u32 = 85;
/// Highest `context_used_pct` accepted.
pub const MAX_CONTEXT_PCT: u32 = 100;
/// Host bullets in front of an assistant line: Claude Code `⏺`, Codex `•` (decision Q2).
pub const BULLETS: [char; 2] = ['⏺', '•'];
/// The four lane markers, each counted only at the start of a line.
pub const MARKER_WORDS: [(&str, MarkerKind); 4] = [
    ("LANE-DONE", MarkerKind::Done),
    ("LANE-REPORT", MarkerKind::Report),
    ("LANE-BLOCKED", MarkerKind::Blocked),
    ("LANE-WAIT", MarkerKind::Wait),
];
/// A line after the last marker holding this text means a card was taken (decision Q3).
pub const CLAIM_TEXT: &str = "moai-claim:";
/// The line a `restart` sends.
pub const RESTART_LINE: &str = "순환 계속 — self-pick (1)부터";

/// The observations file (typed JSON, unknown keys refused).
#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ObservationFile {
    pub schema_version: u32,
    pub lanes: Vec<LaneObservation>,
}

/// One lane as the observer saw it.
#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct LaneObservation {
    pub lane: String,
    /// `claude`, `codex` or `omp` (`lane_duty::HOSTS`).
    pub host: String,
    /// `YYYY-MM-DDTHH:MM:SSZ`, when the terminal was read.
    pub read_at: String,
    /// Who observed, and how.
    pub source: String,
    pub terminal: Terminal,
    pub pane: PaneState,
    pub turn: Turn,
    pub input_box: InputBox,
    /// Context use as the host shows it; absent when the screen does not show it.
    #[serde(default)]
    pub context_used_pct: Option<u32>,
    pub wake_armed: bool,
    pub background_jobs: u64,
    /// The `lane-safe-to-reset` observation parts the forge and the terminal give.
    #[serde(default)]
    pub reset: ResetParts,
}

/// Orca `terminal read --screen --json`, as returned.
#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Terminal {
    pub handle: String,
    pub status: String,
    pub source: String,
    pub tail: String,
    pub limited: bool,
    pub truncated: bool,
    #[serde(default)]
    pub draft: Option<String>,
}

/// The `lane_reset::Observations` parts the observer supplies.
#[derive(Clone, Debug, Default, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ResetParts {
    #[serde(default)]
    pub claims: Option<ClaimsObservation>,
    #[serde(default)]
    pub session_cwd: Option<CwdObservation>,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum Turn {
    Running,
    Idle,
    Unknown,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum InputBox {
    Empty,
    Text,
    Unknown,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum MarkerKind {
    Done,
    Report,
    Blocked,
    Wait,
}

/// A counted `LANE-*` line: its index in the tail and its text without the bullet.
#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct Marker {
    pub kind: MarkerKind,
    pub line: usize,
    pub text: String,
}

#[derive(Clone, Debug, Default, PartialEq, Eq, Serialize)]
pub struct Markers {
    pub found: Vec<Marker>,
    /// Marker lines for another lane or with no lane word.
    pub ignored: Vec<String>,
    /// A `moai-claim:` line stands after the last counted marker.
    pub claim_after_last: bool,
}

/// The `lane-safe-to-reset` answer for the lane.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum ResetVerdict {
    Safe,
    NotSafe(Vec<String>),
    Unknown(Vec<String>),
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum Act {
    ClearRedispatch,
    Restart,
    Answer,
    Leave,
}

/// One lane's action.
#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct Action {
    pub lane: String,
    pub action: Act,
    pub reason: String,
    /// Lines to type into the lane's terminal, in order.
    pub send: Vec<String>,
    /// For `clear_redispatch`: the session-register release, filled by the caller.
    pub release: Option<String>,
    /// For `answer`: the `LANE-BLOCKED` line, as data (never sent back).
    pub data: Option<String>,
    pub findings: Vec<String>,
    pub notes: Vec<String>,
    pub ignored: Vec<String>,
}

/// Reads an observations file and checks its limits.
pub fn parse(text: &str) -> Result<ObservationFile, String> {
    let file: ObservationFile =
        serde_json::from_str(text).map_err(|e| format!("observations file is invalid: {e}"))?;
    if file.schema_version != SCHEMA_VERSION {
        return Err(format!(
            "observations file is invalid: schema_version {} (this moai reads {SCHEMA_VERSION})",
            file.schema_version
        ));
    }
    if file.lanes.len() > MAX_LANES {
        return Err(format!(
            "observations file is invalid: {} lanes (at most {MAX_LANES})",
            file.lanes.len()
        ));
    }
    let mut seen = std::collections::BTreeSet::new();
    for l in &file.lanes {
        if !crate::lane_registry::valid_name(&l.lane) {
            return Err(format!(
                "observations file is invalid: lane name `{}` (use 1 to {} of A-Z a-z 0-9 . _ -)",
                l.lane,
                crate::lane_registry::MAX_NAME_LEN
            ));
        }
        if !seen.insert(l.lane.as_str()) {
            return Err(format!(
                "observations file is invalid: lane `{}` is observed twice",
                l.lane
            ));
        }
        if !crate::lane_duty::HOSTS.iter().any(|(h, _)| *h == l.host) {
            return Err(format!(
                "observations file is invalid: lane `{}` host `{}` (use claude, codex or omp)",
                l.lane, l.host
            ));
        }
        if let Some(pct) = l.context_used_pct.filter(|p| *p > MAX_CONTEXT_PCT) {
            return Err(format!(
                "observations file is invalid: lane `{}` context_used_pct {pct} (at most {MAX_CONTEXT_PCT})",
                l.lane
            ));
        }
    }
    Ok(file)
}

/// The marker word that starts `line`, followed by whitespace or the line end, and the rest.
fn marker_word(line: &str) -> Option<(MarkerKind, &str)> {
    MARKER_WORDS.iter().find_map(|(word, kind)| {
        let rest = line.strip_prefix(word)?;
        (rest.is_empty() || rest.starts_with(char::is_whitespace)).then_some((*kind, rest))
    })
}

/// The lane's counted `LANE-*` lines in `tail`. A line counts when, without surrounding
/// whitespace and one host bullet, it starts with a marker word; quoted (`>`) lines and lines
/// inside ``` fences do not. The second word must be the lane (or `lane=<lane>`, one trailing
/// `:` allowed); a marker for another lane is listed under `ignored`.
pub fn markers(tail: &str, lane: &str) -> Markers {
    let mut out = Markers::default();
    let mut fenced = false;
    let mut last = None;
    let mut claim_lines = Vec::new();
    for (i, raw) in tail.lines().enumerate() {
        if raw.contains(CLAIM_TEXT) {
            claim_lines.push(i);
        }
        let mut line = raw.trim();
        if line.starts_with("```") {
            fenced = !fenced;
            continue;
        }
        if fenced {
            continue;
        }
        if let Some(rest) = line.strip_prefix(BULLETS) {
            line = rest.trim_start();
        }
        if line.starts_with('>') {
            continue;
        }
        let Some((kind, rest)) = marker_word(line) else {
            continue;
        };
        let second = rest
            .split_whitespace()
            .next()
            .map(|w| w.strip_suffix(':').unwrap_or(w));
        if second.is_some_and(|w| w == lane || w.strip_prefix("lane=") == Some(lane)) {
            let kind = if kind == MarkerKind::Report
                && rest.split_whitespace().any(|w| w == "LANE-DONE")
            {
                MarkerKind::Done
            } else {
                kind
            };
            out.found.push(Marker {
                kind,
                line: i,
                text: line.to_string(),
            });
            last = Some(i);
        } else {
            out.ignored.push(line.to_string());
        }
    }
    out.claim_after_last = last.is_some_and(|l| claim_lines.iter().any(|c| *c > l));
    out
}

/// The `until <time>` of a `LANE-WAIT` line, as Unix seconds.
fn wait_until(text: &str) -> Option<i64> {
    let mut words = text.split_whitespace();
    words.find(|w| *w == "until")?;
    crate::forge::policy::parse_utc(words.next()?)
}

fn pane_word(pane: PaneState) -> &'static str {
    match pane {
        PaneState::Working => "working",
        PaneState::PermissionPrompt => "permission_prompt",
        PaneState::InputPending => "input_pending",
        PaneState::Unreachable => "unreachable",
        PaneState::UsageLimit => "usage_limit",
    }
}

/// Why the observation cannot be relied on, if it cannot (row 0).
fn unusable(obs: &LaneObservation, now: i64) -> Option<String> {
    let Some(at) = crate::forge::policy::parse_utc(&obs.read_at) else {
        return Some(format!(
            "read_at `{}` is not YYYY-MM-DDTHH:MM:SSZ",
            obs.read_at
        ));
    };
    if at > now {
        return Some(format!("read_at `{}` is in the future", obs.read_at));
    }
    if now - at > MAX_AGE_SECS {
        return Some(format!(
            "observation is {} s old; the limit is {MAX_AGE_SECS} s",
            now - at
        ));
    }
    if obs.terminal.source != "screen" {
        return Some(format!(
            "terminal source is `{}`, not `screen`",
            obs.terminal.source
        ));
    }
    if obs.terminal.status != "running" {
        return Some(format!(
            "terminal status is `{}`, not `running`",
            obs.terminal.status
        ));
    }
    None
}

/// Rows 0–7 of the plan, first match wins; `now` is Unix seconds.
pub fn decide(obs: &LaneObservation, now: i64, reset: &ResetVerdict) -> Action {
    let found = markers(&obs.terminal.tail, &obs.lane);
    let mut a = Action {
        lane: obs.lane.clone(),
        action: Act::Leave,
        reason: String::new(),
        send: Vec::new(),
        release: None,
        data: None,
        findings: Vec::new(),
        notes: Vec::new(),
        ignored: found.ignored.clone(),
    };
    let leave = |mut a: Action, reason: &str| {
        a.action = Act::Leave;
        a.reason = reason.into();
        a
    };
    // 0. An observation the round cannot rely on.
    if let Some(why) = unusable(obs, now) {
        a.notes.push(why);
        return leave(a, "observation_unusable");
    }
    // 1. A pane the leader must not type into (a permission prompt is not answered).
    if matches!(
        obs.pane,
        PaneState::Unreachable | PaneState::PermissionPrompt | PaneState::UsageLimit
    ) {
        return leave(a, pane_word(obs.pane));
    }
    // 2. A turn that runs, or may.
    match obs.turn {
        Turn::Running => return leave(a, "turn_running"),
        Turn::Unknown => return leave(a, "turn_unknown"),
        Turn::Idle => {}
    }
    // 3. Something typed and not yet sent.
    if obs.input_box != InputBox::Empty || obs.pane == PaneState::InputPending {
        return leave(a, "input_pending");
    }
    // 4. The lane waits on purpose.
    if obs.wake_armed {
        return leave(a, "wake_armed");
    }
    let last = found.found.last();
    if let Some(m) = last.filter(|m| m.kind == MarkerKind::Wait) {
        match wait_until(&m.text) {
            Some(until) if until <= now => {}
            Some(_) => return leave(a, "lane_wait"),
            None => {
                a.notes
                    .push("LANE-WAIT has no `until <YYYY-MM-DDTHH:MM:SSZ>`".into());
                return leave(a, "lane_wait");
            }
        }
    }
    // 5. A question for the leader: its line is data, never sent back.
    if let Some(m) = last.filter(|m| m.kind == MarkerKind::Blocked) {
        a.action = Act::Answer;
        a.reason = "lane_blocked".into();
        a.data = Some(m.text.clone());
        return a;
    }
    // 6. A card boundary: clear and redispatch only on a `safe` reset verdict.
    let pct = match obs.context_used_pct {
        Some(p) => p,
        None => {
            a.notes.push(format!(
                "context_used_pct unknown: treated as below {CLEAR_CONTEXT_PCT} %"
            ));
            0
        }
    };
    let boundary = !found.claim_after_last
        && last.is_some_and(|m| {
            m.kind == MarkerKind::Done || (m.kind == MarkerKind::Report && pct >= CLEAR_CONTEXT_PCT)
        });
    if boundary {
        return match reset {
            ResetVerdict::Safe => match crate::lane_duty::line(&obs.lane, None, &obs.host) {
                Ok(line) => {
                    a.action = Act::ClearRedispatch;
                    a.reason = "card_boundary".into();
                    a.send = vec!["/clear".into(), line];
                    a
                }
                Err(e) => {
                    a.notes.push(e);
                    leave(a, "duty_line_invalid")
                }
            },
            ResetVerdict::NotSafe(findings) => {
                a.findings = findings.clone();
                leave(a, "reset_not_safe")
            }
            ResetVerdict::Unknown(unknowns) => {
                a.findings = unknowns.clone();
                leave(a, "reset_unknown")
            }
        };
    }
    // 7. Anything else: restart the loop.
    // O3b step 4: the end-condition check (`startable::end_condition`) and the
    // no-repeat rule come with O2 #2916; until then this always restarts.
    a.action = Act::Restart;
    a.reason = "continue".into();
    a.send = vec![RESTART_LINE.into()];
    a
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::forge::policy::parse_utc;

    const WORKING: &str = include_str!("../tests/fixtures/leader_round/working.json");
    const INPUT_PRESENT: &str = include_str!("../tests/fixtures/leader_round/input_present.json");
    const LANE_DONE: &str = include_str!("../tests/fixtures/leader_round/lane_done.json");
    const UNPUSHED_HEAD: &str = include_str!("../tests/fixtures/leader_round/unpushed_head.json");
    const CONTEXT_91: &str = include_str!("../tests/fixtures/leader_round/context_91.json");

    /// The fixtures' `read_at` plus 60 s.
    fn now() -> i64 {
        parse_utc("2026-11-14T09:20:00Z").unwrap() + 60
    }

    fn lane(fixture: &str) -> LaneObservation {
        let mut file = parse(fixture).expect("fixture parses");
        assert_eq!(file.lanes.len(), 1);
        file.lanes.remove(0)
    }

    fn safe(obs: &LaneObservation) -> Action {
        decide(obs, now(), &ResetVerdict::Safe)
    }

    fn duty() -> String {
        crate::lane_duty::line("agent-29", None, "claude").unwrap()
    }

    fn with_tail(tail: &str) -> LaneObservation {
        let mut obs = lane(LANE_DONE);
        obs.terminal.tail = tail.into();
        obs
    }

    // AC1: the five goldens.

    #[test]
    fn golden_working_is_left() {
        let a = safe(&lane(WORKING));
        assert_eq!((a.action, a.reason.as_str()), (Act::Leave, "turn_running"));
        assert!(a.send.is_empty());
    }

    #[test]
    fn golden_input_present_is_left() {
        let a = safe(&lane(INPUT_PRESENT));
        assert_eq!((a.action, a.reason.as_str()), (Act::Leave, "input_pending"));
        assert!(a.send.is_empty());
    }

    #[test]
    fn golden_lane_done_clears_and_redispatches() {
        let obs = lane(LANE_DONE);
        assert_eq!(obs.reset.claims.as_ref().unwrap().items[0].issue, 2902);
        let a = safe(&obs);
        assert_eq!(
            (a.action, a.reason.as_str()),
            (Act::ClearRedispatch, "card_boundary")
        );
        assert_eq!(a.send, vec!["/clear".to_string(), duty()]);
        assert_eq!(a.release, None);
    }

    #[test]
    fn golden_unpushed_head_is_left_with_the_finding() {
        let finding = "/work/moai-adk-wt-2902: 1 unpushed commit".to_string();
        let a = decide(
            &lane(UNPUSHED_HEAD),
            now(),
            &ResetVerdict::NotSafe(vec![finding.clone()]),
        );
        assert_eq!(
            (a.action, a.reason.as_str()),
            (Act::Leave, "reset_not_safe")
        );
        assert_eq!(a.findings, vec![finding]);
        assert!(a.send.is_empty());
    }

    #[test]
    fn golden_context_91_report_clears() {
        let a = safe(&lane(CONTEXT_91));
        assert_eq!(
            (a.action, a.reason.as_str()),
            (Act::ClearRedispatch, "card_boundary")
        );
        assert_eq!(a.send, vec!["/clear".to_string(), duty()]);
    }

    // AC2 / AC7: not_safe and unknown leave the lane with the reason.

    #[test]
    fn reset_unknown_leaves_the_lane() {
        let why = "no screen observation: background jobs unchecked".to_string();
        let a = decide(
            &lane(LANE_DONE),
            now(),
            &ResetVerdict::Unknown(vec![why.clone()]),
        );
        assert_eq!((a.action, a.reason.as_str()), (Act::Leave, "reset_unknown"));
        assert_eq!(a.findings, vec![why]);
        assert!(a.send.is_empty());
    }

    // AC4: a pane the leader must not type into.

    #[test]
    fn blocked_panes_are_left_and_prompts_unanswered() {
        for (pane, reason) in [
            (PaneState::PermissionPrompt, "permission_prompt"),
            (PaneState::Unreachable, "unreachable"),
            (PaneState::UsageLimit, "usage_limit"),
        ] {
            let mut obs = lane(LANE_DONE);
            obs.pane = pane;
            let a = safe(&obs);
            assert_eq!((a.action, a.reason.as_str()), (Act::Leave, reason));
            assert!(a.send.is_empty());
        }
    }

    #[test]
    fn unknown_turn_and_input_box_are_left() {
        let mut obs = lane(LANE_DONE);
        obs.turn = Turn::Unknown;
        assert_eq!(safe(&obs).reason, "turn_unknown");
        let mut obs = lane(LANE_DONE);
        obs.input_box = InputBox::Unknown;
        assert_eq!(safe(&obs).reason, "input_pending");
        let mut obs = lane(LANE_DONE);
        obs.pane = PaneState::InputPending;
        assert_eq!(safe(&obs).reason, "input_pending");
    }

    // AC5: the context threshold for a LANE-REPORT.

    #[test]
    fn report_clears_from_85_percent() {
        for (pct, act) in [
            (Some(85), Act::ClearRedispatch),
            (Some(91), Act::ClearRedispatch),
            (Some(84), Act::Restart),
            (None, Act::Restart),
        ] {
            let mut obs = lane(CONTEXT_91);
            obs.context_used_pct = pct;
            let a = safe(&obs);
            assert_eq!(a.action, act, "context {pct:?}");
            if pct.is_none() {
                assert!(
                    a.notes
                        .iter()
                        .any(|n| n.contains("context_used_pct unknown")),
                    "{:?}",
                    a.notes
                );
            }
        }
    }

    #[test]
    fn a_claim_after_the_last_marker_is_no_boundary() {
        let obs = with_tail(
            "⏺ LANE-DONE agent-29 #2902 O3a\n⏺ moai-claim: session=agent-29 epoch=1 lease_until=2026-11-14T16:00:00Z",
        );
        let a = safe(&obs);
        assert_eq!(a.action, Act::Restart);
    }

    // AC6: only the lane's own LANE-* lines at a line start count.

    #[test]
    fn markers_skip_quotes_fences_mid_line_and_other_lanes() {
        let tail = "> LANE-DONE agent-29 quoted\n\
                    ```\n\
                    LANE-DONE agent-29 fenced\n\
                    ```\n\
                    see LANE-DONE agent-29 mid-line\n\
                    ⏺ LANE-DONE agent-2 other lane\n\
                    LANE-DONE agent-290 longer lane\n\
                    LANE-DONE lane=agent-290 longer lane key\n\
                    LANE-DONEX agent-29 longer word\n\
                    LANE-DONE\n";
        let m = markers(tail, "agent-29");
        assert!(m.found.is_empty(), "{:?}", m.found);
        assert_eq!(
            m.ignored,
            vec![
                "LANE-DONE agent-2 other lane".to_string(),
                "LANE-DONE agent-290 longer lane".into(),
                "LANE-DONE lane=agent-290 longer lane key".into(),
                "LANE-DONE".into()
            ]
        );
        let a = safe(&with_tail(tail));
        assert_eq!(a.action, Act::Restart);
        assert_eq!(a.ignored.len(), 4);
    }

    #[test]
    fn markers_count_bullets_lane_keys_and_report_with_done() {
        let tail = "  ⏺ LANE-WAIT agent-29 usage-limit until 2026-11-14T10:00:00Z\n\
                    • LANE-BLOCKED lane=agent-29 decision needed\n\
                    LANE-REPORT agent-29: LANE-DONE agent-29 PR #2912 MERGED\n\
                    ⏺ LANE-REPORT agent-29 progress only";
        let m = markers(tail, "agent-29");
        let kinds: Vec<_> = m.found.iter().map(|k| k.kind).collect();
        assert_eq!(
            kinds,
            vec![
                MarkerKind::Wait,
                MarkerKind::Blocked,
                MarkerKind::Done,
                MarkerKind::Report
            ]
        );
        assert_eq!(
            m.found[1].text,
            "LANE-BLOCKED lane=agent-29 decision needed"
        );
        assert_eq!(m.found[3].line, 3);
        assert!(m.ignored.is_empty());
        assert!(!m.claim_after_last);
    }

    #[test]
    fn a_fence_opened_with_a_language_still_hides_its_lines() {
        let m = markers(
            "```text\nLANE-DONE agent-29 inside\n```\nLANE-DONE agent-29 outside",
            "agent-29",
        );
        assert_eq!(m.found.len(), 1);
        assert_eq!(m.found[0].text, "LANE-DONE agent-29 outside");
    }

    // AC8: the observation's age.

    #[test]
    fn observation_age_limit_is_600_seconds() {
        let at = parse_utc("2026-11-14T09:20:00Z").unwrap();
        let obs = lane(LANE_DONE);
        let ok = decide(&obs, at + 600, &ResetVerdict::Safe);
        assert_eq!(ok.action, Act::ClearRedispatch);
        let old = decide(&obs, at + 601, &ResetVerdict::Safe);
        assert_eq!(
            (old.action, old.reason.as_str()),
            (Act::Leave, "observation_unusable")
        );
        let future = decide(&obs, at - 1, &ResetVerdict::Safe);
        assert_eq!(future.reason, "observation_unusable");
    }

    #[test]
    fn unusable_terminal_reads_are_left() {
        let mut obs = lane(LANE_DONE);
        obs.terminal.source = "scrollback".into();
        assert_eq!(safe(&obs).reason, "observation_unusable");
        let mut obs = lane(LANE_DONE);
        obs.terminal.status = "exited".into();
        assert_eq!(safe(&obs).reason, "observation_unusable");
        let mut obs = lane(LANE_DONE);
        obs.read_at = "2026-11-14 09:20".into();
        assert_eq!(safe(&obs).reason, "observation_unusable");
    }

    // Row 4: a wake or a LANE-WAIT with time left.

    #[test]
    fn wakes_and_waits_with_time_left_are_left() {
        let mut obs = lane(LANE_DONE);
        obs.wake_armed = true;
        assert_eq!(safe(&obs).reason, "wake_armed");
        let wait =
            |until: &str| with_tail(&format!("⏺ LANE-WAIT agent-29 usage-limit until {until}"));
        let a = safe(&wait("2026-11-14T09:30:00Z"));
        assert_eq!((a.action, a.reason.as_str()), (Act::Leave, "lane_wait"));
        let a = safe(&wait("2026-11-14T09:00:00Z"));
        assert_eq!(a.action, Act::Restart);
        assert_eq!(a.send, vec![RESTART_LINE.to_string()]);
    }

    // AC9: input limits.

    fn file_with(lanes: usize, pct: u32) -> String {
        let one = |i: usize| {
            let mut v: serde_json::Value = serde_json::from_str(LANE_DONE).unwrap();
            let mut l = v["lanes"][0].take();
            l["lane"] = format!("agent-{i}").into();
            l["context_used_pct"] = pct.into();
            l
        };
        serde_json::json!({
            "schema_version": 1,
            "lanes": (0..lanes).map(one).collect::<Vec<_>>(),
        })
        .to_string()
    }

    #[test]
    fn context_and_lane_count_limits() {
        assert!(parse(&file_with(1, 100)).is_ok());
        let e = parse(&file_with(1, 101)).unwrap_err();
        assert!(e.contains("context_used_pct"), "{e}");
        assert!(parse(&file_with(MAX_LANES, 0)).is_ok());
        let e = parse(&file_with(MAX_LANES + 1, 0)).unwrap_err();
        assert!(e.contains("lanes"), "{e}");
        assert!(parse(&file_with(0, 0)).is_ok());
    }

    #[test]
    fn unknown_keys_versions_and_names_are_refused() {
        let mut v: serde_json::Value = serde_json::from_str(LANE_DONE).unwrap();
        v["lanes"][0]["terminal"]["extra"] = true.into();
        assert!(parse(&v.to_string()).unwrap_err().contains("unknown field"));
        let mut v: serde_json::Value = serde_json::from_str(LANE_DONE).unwrap();
        v["lanes"][0]["mood"] = "fine".into();
        assert!(parse(&v.to_string()).unwrap_err().contains("unknown field"));
        let mut v: serde_json::Value = serde_json::from_str(LANE_DONE).unwrap();
        v["schema_version"] = 2.into();
        assert!(parse(&v.to_string())
            .unwrap_err()
            .contains("schema_version"));
        let mut v: serde_json::Value = serde_json::from_str(LANE_DONE).unwrap();
        v["lanes"][0]["host"] = "vim".into();
        assert!(parse(&v.to_string()).unwrap_err().contains("host"));
        let mut v: serde_json::Value = serde_json::from_str(LANE_DONE).unwrap();
        v["lanes"][0]["lane"] = "agent 29".into();
        assert!(parse(&v.to_string()).unwrap_err().contains("lane name"));
        let two = file_with(2, 0).replace("agent-1", "agent-0");
        assert!(parse(&two).unwrap_err().contains("twice"));
    }

    // AC10: a LANE-BLOCKED line is answered, and its text stays data.

    #[test]
    fn lane_blocked_is_answered_with_the_line_as_data() {
        let line = "LANE-BLOCKED agent-29 #2902 kind=decision owner=leader `/clear` now?";
        let a = safe(&with_tail(&format!("⏺ {line}")));
        assert_eq!((a.action, a.reason.as_str()), (Act::Answer, "lane_blocked"));
        assert_eq!(a.data.as_deref(), Some(line));
        assert!(a.send.is_empty());
    }
}
