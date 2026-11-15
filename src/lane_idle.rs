//! #3005 P8 (user decision, Issue #3005 comment 6887207942): a lane that watches for work
//! with `lane-next` stops after N empty reads in a row. The host calls `lane-next` again and
//! again, each call a new process, so the count lives in the clone-shared state folder
//! (`<clone state>/lane-idle/<lane>__<owner>__<repo>.json`), one file per lane and repository.
//!
//! An empty read is a read with no task and nothing the lane waits for; a read that finds a
//! task, or only something to wait for (its pull request on CI, an audit, a decision), sets
//! the count back to 0, so a lane never goes idle while its own work is in flight.
//!
//! The count is advisory: a state folder that cannot be written (not a git clone, a link in
//! the way) leaves the count off and the lane keeps going; it never turns an error into a stop.

use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::path::{Path, PathBuf};

/// Empty reads in a row after which `lane-next` answers `idle_limit` (`--max-empty`).
pub const DEFAULT_MAX_EMPTY: u64 = 6;
/// The largest `--max-empty` accepted.
pub const MAX_MAX_EMPTY: u64 = 1000;
const STATE_FILE_LIMIT: u64 = 4096;

/// What one lane has seen so far.
#[derive(Clone, Debug, Default, PartialEq, Eq, Serialize, Deserialize)]
pub struct IdleState {
    /// Empty reads since the last read that found something.
    #[serde(default)]
    pub consecutive_empty: u64,
    /// When the last read happened (UTC), for the `LANE-IDLE` line.
    #[serde(default)]
    pub last: String,
}

/// The limit a run uses: `--max-empty` when given, else [`DEFAULT_MAX_EMPTY`] on a GitHub
/// read and 0 on a `--snapshot` read (a measurement stays read-only); above [`MAX_MAX_EMPTY`]
/// is refused.
pub fn resolve_max(arg: Option<u64>, snapshot: bool) -> Result<u64, String> {
    let max = match (arg, snapshot) {
        (Some(n), _) => n,
        (None, false) => DEFAULT_MAX_EMPTY,
        (None, true) => 0,
    };
    if max > MAX_MAX_EMPTY {
        return Err(format!("--max-empty accepts 0 to {MAX_MAX_EMPTY}"));
    }
    Ok(max)
}

/// A read with no task and nothing the lane waits for.
pub fn is_empty_read(result: &Value) -> bool {
    result["next"].is_null() && result["waiting"].as_array().is_none_or(Vec::is_empty)
}

/// The state after `reads` reads that were all empty (`empty`) or the last of which found
/// something, and whether the limit is reached. `max_empty` 0 does not count: the state is
/// returned unchanged and the limit is never reached.
pub fn step(
    prev: &IdleState,
    empty: bool,
    reads: u64,
    max_empty: u64,
    now: &str,
) -> (IdleState, bool) {
    if max_empty == 0 {
        return (prev.clone(), false);
    }
    let consecutive_empty = if empty {
        prev.consecutive_empty.saturating_add(reads.max(1))
    } else {
        0
    };
    let state = IdleState {
        consecutive_empty,
        last: now.to_string(),
    };
    (state, empty && consecutive_empty >= max_empty)
}

/// `lane` and `part` with every character outside `A-Za-z0-9._-` replaced by `_`.
fn safe(part: &str) -> String {
    part.chars()
        .map(|c| {
            if c.is_ascii_alphanumeric() || matches!(c, '.' | '_' | '-') {
                c
            } else {
                '_'
            }
        })
        .collect()
}

/// The state file of `lane` for `repo` (`OWNER/REPO`) under the clone state folder `anchor`.
pub fn state_file(anchor: &Path, lane: &str, repo: &str) -> PathBuf {
    let (owner, name) = repo.split_once('/').unwrap_or(("", repo));
    crate::clone_root::path(anchor)
        .join("lane-idle")
        .join(format!(
            "{}__{}__{}.json",
            safe(lane),
            safe(owner),
            safe(name)
        ))
}

/// The saved state; a missing, oversized, linked or broken file is a fresh state.
pub fn load(file: &Path) -> IdleState {
    match std::fs::symlink_metadata(file) {
        Ok(m) if m.is_file() && m.len() <= STATE_FILE_LIMIT => std::fs::read(file)
            .ok()
            .and_then(|b| serde_json::from_slice(&b).ok())
            .unwrap_or_default(),
        _ => IdleState::default(),
    }
}

fn real_dir(p: &Path) -> bool {
    std::fs::symlink_metadata(p).is_ok_and(|m| m.is_dir())
}

/// Writes the state under the clone state folder, creating the two folders one level at a
/// time and never through a link.
pub fn save(anchor: &Path, file: &Path, state: &IdleState) -> Result<(), String> {
    let root = crate::clone_root::path(anchor);
    let dir = file.parent().ok_or("state file has no folder")?;
    for folder in [root.as_path(), dir] {
        if std::fs::symlink_metadata(folder).is_err() {
            std::fs::create_dir(folder).map_err(|e| format!("{}: {e}", folder.display()))?;
        }
        if !real_dir(folder) {
            return Err(format!("{} is not a plain folder", folder.display()));
        }
    }
    if std::fs::symlink_metadata(file).is_ok_and(|m| !m.is_file()) {
        return Err(format!("{} is not a plain file", file.display()));
    }
    let bytes = serde_json::to_vec(state).map_err(|e| e.to_string())?;
    std::fs::write(file, bytes).map_err(|e| format!("{}: {e}", file.display()))
}

/// The one line a lane prints (and posts on its coordination Issue) when `lane-next` answers
/// `idle_limit`. The leader's round reads it as an idle lane.
pub fn idle_line(lane: &str, checks: u64, last: &str) -> String {
    format!(
        "LANE-IDLE lane={lane} checks={checks} last={last} restart=enter MODE-AGENT again as lane {lane} (`/moai --mode-agent` in Claude Code, `$moai --mode-agent` in Codex CLI)"
    )
}

/// Counts this read for `lane` and adds `idle` to the `lane-next` result (and `status:
/// "idle_limit"` when the limit is reached). `max_empty` 0 counts nothing and adds nothing.
/// `anchor` is the clone's common git folder; `None` (not a clone) leaves the count off.
pub fn apply(
    result: &mut Value,
    anchor: Option<&Path>,
    lane: &str,
    repo: &str,
    max_empty: u64,
    now: &str,
) {
    if max_empty == 0 {
        return;
    }
    let Some(anchor) = anchor else {
        result["idle"] =
            json!({"counting": false, "reason": "not a git clone: the empty reads cannot be kept"});
        return;
    };
    let file = state_file(anchor, lane, repo);
    let prev = load(&file);
    let reads = result["rounds"].as_u64().unwrap_or(1);
    let (state, limit) = step(&prev, is_empty_read(result), reads, max_empty, now);
    if let Err(e) = save(anchor, &file, &state) {
        result["idle"] =
            json!({"counting": false, "reason": format!("the empty reads cannot be kept: {e}")});
        return;
    }
    result["writes"] = json!(true);
    result["idle"] = json!({
        "counting": true,
        "consecutive_empty": state.consecutive_empty,
        "max_empty": max_empty,
        "limit_reached": limit,
    });
    if limit {
        result["status"] = json!("idle_limit");
        result["idle"]["line"] = json!(idle_line(lane, state.consecutive_empty, now));
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    const NOW: &str = "2026-11-15T02:00:00Z";

    fn run(prev: &IdleState, empty: bool, max: u64) -> (IdleState, bool) {
        step(prev, empty, 1, max, NOW)
    }

    /// The boundary: reads 1 to N-1 keep going, read N is the limit, a read that finds
    /// something sets the count back to 0, and the read after a reset starts over.
    #[test]
    fn the_limit_is_the_nth_empty_read_in_a_row() {
        let mut s = IdleState::default();
        for n in 1..=5 {
            let (next, limit) = run(&s, true, 6);
            assert_eq!(next.consecutive_empty, n);
            assert!(!limit, "read {n} of 6 is not the limit");
            s = next;
        }
        let (s, limit) = run(&s, true, 6);
        assert_eq!((s.consecutive_empty, limit), (6, true), "read 6 of 6");
        let (s, limit) = run(&s, true, 6);
        assert_eq!(
            (s.consecutive_empty, limit),
            (7, true),
            "it stays at the limit"
        );
        let (s, limit) = run(&s, false, 6);
        assert_eq!(
            (s.consecutive_empty, limit),
            (0, false),
            "work resets the count"
        );
        let (s, limit) = run(&s, true, 6);
        assert_eq!(
            (s.consecutive_empty, limit),
            (1, false),
            "and counting starts over"
        );
    }

    #[test]
    fn max_empty_zero_and_one_are_the_two_edges() {
        let fresh = IdleState::default();
        let (s, limit) = run(&fresh, true, 0);
        assert_eq!(
            (s.clone(), limit),
            (fresh.clone(), false),
            "0 counts nothing"
        );
        let (s, limit) = run(&fresh, true, 1);
        assert_eq!(
            (s.consecutive_empty, limit),
            (1, true),
            "1 stops at the first empty read"
        );
        let (_, limit) = run(&fresh, false, 1);
        assert!(!limit, "a read with work is never the limit");
    }

    /// A `--wait` call that read several times and found nothing counts every read.
    #[test]
    fn every_read_inside_a_wait_counts() {
        let fresh = IdleState::default();
        let (s, limit) = step(&fresh, true, 4, 6, NOW);
        assert_eq!((s.consecutive_empty, limit), (4, false));
        let (s, limit) = step(&s, true, 4, 6, NOW);
        assert_eq!((s.consecutive_empty, limit), (8, true));
        let (s, _) = step(
            &IdleState {
                consecutive_empty: u64::MAX,
                last: String::new(),
            },
            true,
            4,
            6,
            NOW,
        );
        assert_eq!(s.consecutive_empty, u64::MAX, "the count saturates");
    }

    /// The default is 6 on a GitHub read, a snapshot read counts only when asked, and 1000 is
    /// the largest value (1001 is refused).
    #[test]
    fn the_default_is_six_on_a_github_read_and_zero_on_a_snapshot() {
        assert_eq!(resolve_max(None, false), Ok(6));
        assert_eq!(resolve_max(None, true), Ok(0));
        assert_eq!(resolve_max(Some(3), true), Ok(3));
        assert_eq!(resolve_max(Some(0), false), Ok(0));
        assert_eq!(resolve_max(Some(1000), false), Ok(1000));
        assert!(resolve_max(Some(1001), false).is_err());
        assert!(resolve_max(Some(1001), true).is_err());
    }

    #[test]
    fn a_read_with_something_to_wait_for_is_not_empty() {
        assert!(is_empty_read(&json!({"next": null, "waiting": []})));
        assert!(is_empty_read(&json!({"next": null})));
        assert!(!is_empty_read(
            &json!({"next": null, "waiting": [{"kind": "wait"}]})
        ));
        assert!(!is_empty_read(
            &json!({"next": {"kind": "audit"}, "waiting": []})
        ));
    }

    #[test]
    fn two_lanes_or_two_repositories_never_share_a_count() {
        let a = Path::new("/c");
        let f = |l: &str, r: &str| state_file(a, l, r);
        assert_ne!(f("agent-1", "o/r"), f("agent-2", "o/r"));
        assert_ne!(f("agent-1", "o/r"), f("agent-1", "o/s"));
        assert_ne!(f("agent-1", "o/r"), f("agent-1", "p/r"));
        let file = f("../x", "o/../r");
        assert_eq!(
            file.parent().unwrap(),
            crate::clone_root::path(a).join("lane-idle")
        );
        assert!(!file.file_name().unwrap().to_string_lossy().contains('/'));
    }

    #[test]
    fn a_missing_oversized_or_broken_file_is_a_fresh_count() {
        let d = tempfile::tempdir().unwrap();
        let f = d.path().join("s.json");
        assert_eq!(load(&f), IdleState::default());
        std::fs::write(&f, b"not json").unwrap();
        assert_eq!(load(&f), IdleState::default());
        let padding = " ".repeat(5000);
        std::fs::write(
            &f,
            format!(r#"{{"consecutive_empty":3,"last":"{padding}"}}"#),
        )
        .unwrap();
        assert_eq!(
            load(&f),
            IdleState::default(),
            "a valid file past the size limit is not read"
        );
        std::fs::write(&f, vec![b' '; 5000]).unwrap();
        assert_eq!(load(&f), IdleState::default());
        std::fs::write(&f, br#"{"consecutive_empty":3,"last":"x"}"#).unwrap();
        assert_eq!(load(&f).consecutive_empty, 3);
        std::fs::write(&f, br#"{"consecutive_empty":3,"extra":1}"#).unwrap();
        assert_eq!(load(&f).consecutive_empty, 3, "an unknown field is ignored");
    }

    #[test]
    fn the_line_names_the_lane_the_count_the_time_and_how_to_restart() {
        let line = idle_line("agent-22", 6, NOW);
        assert!(
            line.starts_with("LANE-IDLE lane=agent-22 checks=6 last=2026-11-15T02:00:00Z restart=")
        );
        assert!(line.contains("MODE-AGENT") && !line.contains('\n'));
    }

    #[cfg(unix)]
    #[test]
    fn a_link_in_the_way_leaves_the_count_off() {
        let d = tempfile::tempdir().unwrap();
        let anchor = d.path().join("git");
        std::fs::create_dir(&anchor).unwrap();
        let elsewhere = d.path().join("elsewhere");
        std::fs::create_dir(&elsewhere).unwrap();
        std::os::unix::fs::symlink(&elsewhere, crate::clone_root::path(&anchor)).unwrap();
        let mut v = json!({"next": null, "waiting": []});
        apply(&mut v, Some(&anchor), "agent-1", "o/r", 6, NOW);
        assert_eq!(v["idle"]["counting"], json!(false), "{v}");
        assert!(v.get("status").is_none());
        assert_eq!(
            std::fs::read_dir(&elsewhere).unwrap().count(),
            0,
            "nothing written through the link"
        );
    }

    #[test]
    fn apply_counts_to_the_limit_and_writes_the_line() {
        let d = tempfile::tempdir().unwrap();
        let anchor = d.path();
        for n in 1..=3u64 {
            let mut v = json!({"next": null, "waiting": [], "writes": false});
            apply(&mut v, Some(anchor), "agent-1", "o/r", 3, NOW);
            assert_eq!(v["idle"]["consecutive_empty"], json!(n));
            assert_eq!(v["idle"]["limit_reached"], json!(n == 3), "read {n}");
            assert_eq!(v.get("status").is_some(), n == 3);
            assert_eq!(v["writes"], json!(true));
        }
        let mut v = json!({"next": {"kind": "issue"}, "waiting": []});
        apply(&mut v, Some(anchor), "agent-1", "o/r", 3, NOW);
        assert_eq!(v["idle"]["consecutive_empty"], json!(0));
        let mut v = json!({"next": null});
        apply(&mut v, None, "agent-1", "o/r", 3, NOW);
        assert_eq!(v["idle"]["counting"], json!(false));
        assert!(v.get("status").is_none(), "no clone never stops the lane");
        let mut v = json!({"next": null, "writes": false});
        apply(&mut v, Some(anchor), "agent-1", "o/r", 0, NOW);
        assert!(
            v.get("idle").is_none() && v["writes"] == json!(false),
            "0 adds nothing"
        );
    }
}
