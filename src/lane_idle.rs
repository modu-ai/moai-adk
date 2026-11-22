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

/// What one `lane-next` read says about the lane.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Read {
    /// No task, nothing to wait for, and GitHub answered in full: the lane has nothing to do.
    Empty,
    /// A task, something to wait for, or an open pull request of the lane's own: work in flight.
    Work,
    /// GitHub answered in part (`partial`), or the lane-wide end is unknown: the read says
    /// nothing about whether the lane has work, so it is not counted either way.
    Unreadable,
}

/// Sorts a `lane-next` result. Work wins over everything: a lane with its own open pull
/// request is not idle, even when the read was partial. That is `own_open_prs` (every own
/// pull request, a draft too) or `pr_skipped` reason `own_no_blocker` (no blocker now, CI or
/// audit pending). Only a full read with no work is empty.
pub fn classify(result: &Value) -> Read {
    let own_pr = result["own_open_prs"]
        .as_array()
        .is_some_and(|l| !l.is_empty())
        || result["pr_skipped"]
            .as_array()
            .is_some_and(|l| l.iter().any(|p| p["reason"] == json!("own_no_blocker")));
    if !result["next"].is_null()
        || result["waiting"].as_array().is_some_and(|w| !w.is_empty())
        || own_pr
    {
        return Read::Work;
    }
    // Two notes are always there in the state in which a lane goes idle, so they do not make a
    // read unreadable: no tip base is known (nothing is open), and the comment window of a busy
    // Issue (the newest comments were read). Any other note is a part GitHub did not answer.
    let partial = result["partial"].as_array().is_some_and(|p| {
        p.iter().any(|n| {
            n.as_str().is_none_or(|s| {
                s != crate::lane_next::TIP_BASE_UNKNOWN
                    && !crate::lane_next::is_comment_window_note(s)
            })
        })
    });
    if partial || !result["end_condition"]["met"].is_boolean() {
        return Read::Unreadable;
    }
    Read::Empty
}

/// What the reads of one `lane-next` call add to the count: `empty` reads in a row at the
/// end of the call, and whether a read with work came before them (`reset`, which starts the
/// count over from 0). An unreadable read changes neither.
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq)]
pub struct Tally {
    pub empty: u64,
    pub reset: bool,
}

impl Tally {
    pub fn add(&mut self, read: Read) {
        match read {
            Read::Empty => self.empty = self.empty.saturating_add(1),
            Read::Work => {
                self.empty = 0;
                self.reset = true;
            }
            Read::Unreadable => {}
        }
    }

    /// Whether the call counted anything (an unreadable call leaves the saved count alone).
    pub fn counted(&self) -> bool {
        self.empty > 0 || self.reset
    }

    fn to_json(self) -> Value {
        json!({"empty": self.empty, "reset": self.reset})
    }

    fn from_json(v: &Value) -> Option<Self> {
        Some(Self {
            empty: v["empty"].as_u64()?,
            reset: v["reset"].as_bool()?,
        })
    }
}

/// The limit and the count already saved, so a `--wait` call can stop at the Nth empty read
/// instead of at the end of the wait.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct IdleBudget {
    pub max: u64,
    pub already: u64,
}

impl IdleBudget {
    /// Whether `tally` brings the count to the limit.
    pub fn reached(&self, tally: &Tally) -> bool {
        let base = if tally.reset { 0 } else { self.already };
        self.max > 0 && tally.empty > 0 && base.saturating_add(tally.empty) >= self.max
    }
}

/// The budget a `lane-next` run counts against: the limit and the count `lane` already has for
/// `repo`. `None` when nothing is counted: `max_empty` 0, or no clone (`anchor` `None`).
pub fn budget(anchor: Option<&Path>, lane: &str, repo: &str, max_empty: u64) -> Option<IdleBudget> {
    let anchor = anchor.filter(|_| max_empty > 0)?;
    Some(IdleBudget {
        max: max_empty,
        already: load(&state_file(anchor, lane, repo)).consecutive_empty,
    })
}

/// Attaches what the reads of the call counted to the result, for [`apply`].
pub fn attach(result: &mut Value, tally: &Tally) {
    result["idle_read"] = tally.to_json();
}

/// The saved state after a call that counted `tally`, and whether the limit is reached.
/// `max_empty` 0, or a call that counted nothing, returns the state unchanged and never the
/// limit.
pub fn step(prev: &IdleState, tally: &Tally, max_empty: u64, now: &str) -> (IdleState, bool) {
    if max_empty == 0 || !tally.counted() {
        return (prev.clone(), false);
    }
    let base = if tally.reset {
        0
    } else {
        prev.consecutive_empty
    };
    let consecutive_empty = base.saturating_add(tally.empty);
    let state = IdleState {
        consecutive_empty,
        last: now.to_string(),
    };
    (state, tally.empty > 0 && consecutive_empty >= max_empty)
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
    // `run` counts each read of a `--wait` call and leaves the tally here; a single read
    // (a snapshot, or a call that did not wait) is sorted now.
    let counted = result.as_object_mut().and_then(|o| o.remove("idle_read"));
    if max_empty == 0 {
        return;
    }
    let Some(anchor) = anchor else {
        result["idle"] =
            json!({"counting": false, "reason": "not a git clone: the empty reads cannot be kept"});
        return;
    };
    let tally = counted
        .as_ref()
        .and_then(Tally::from_json)
        .unwrap_or_else(|| {
            let mut t = Tally::default();
            t.add(classify(result));
            t
        });
    let file = state_file(anchor, lane, repo);
    let prev = load(&file);
    if !tally.counted() {
        result["idle"] = json!({
            "counting": true,
            "consecutive_empty": prev.consecutive_empty,
            "max_empty": max_empty,
            "limit_reached": false,
            "not_counted": "the read was partial or its end is unknown (GitHub did not answer in full), so it is not an empty read",
        });
        return;
    }
    let (state, limit) = step(&prev, &tally, max_empty, now);
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
        let mut tally = Tally::default();
        tally.add(if empty { Read::Empty } else { Read::Work });
        step(prev, &tally, max, NOW)
    }

    /// A full read of GitHub with no task, nothing to wait for and the lane-wide end known.
    fn empty_read() -> Value {
        json!({"next": null, "waiting": [], "pr_skipped": [], "partial": [],
            "end_condition": {"met": false}})
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

    fn tally(reads: &[Read]) -> Tally {
        let mut t = Tally::default();
        for r in reads {
            t.add(*r);
        }
        t
    }

    /// A `--wait` call that read several times counts every empty read, and a read with work
    /// inside the call starts the count over from 0, before the empty reads that follow it.
    #[test]
    fn every_read_inside_a_wait_counts_from_the_last_work() {
        let fresh = IdleState::default();
        let (s, limit) = step(&fresh, &tally(&[Read::Empty; 4]), 6, NOW);
        assert_eq!((s.consecutive_empty, limit), (4, false));
        let (s, limit) = step(&s, &tally(&[Read::Empty; 4]), 6, NOW);
        assert_eq!((s.consecutive_empty, limit), (8, true));
        let (s, _) = step(
            &IdleState {
                consecutive_empty: u64::MAX,
                last: String::new(),
            },
            &tally(&[Read::Empty; 4]),
            6,
            NOW,
        );
        assert_eq!(s.consecutive_empty, u64::MAX, "the count saturates");
        // Five empty reads saved, then a call with work and two empty reads after it: 2, not 7.
        let five = IdleState {
            consecutive_empty: 5,
            last: String::new(),
        };
        let (s, limit) = step(
            &five,
            &tally(&[Read::Empty, Read::Work, Read::Empty, Read::Empty]),
            6,
            NOW,
        );
        assert_eq!((s.consecutive_empty, limit), (2, false));
        let (s, limit) = step(&five, &tally(&[Read::Empty, Read::Work]), 6, NOW);
        assert_eq!((s.consecutive_empty, limit), (0, false), "work last: 0");
    }

    /// A call whose reads were all unreadable counts nothing: the saved state is returned
    /// unchanged and the limit is never reached, however high the saved count already is.
    #[test]
    fn an_unreadable_call_leaves_the_count_alone() {
        let five = IdleState {
            consecutive_empty: 5,
            last: "before".into(),
        };
        let t = tally(&[Read::Unreadable; 3]);
        assert!(!t.counted());
        assert_eq!(step(&five, &t, 6, NOW), (five.clone(), false));
        let at_limit = IdleState {
            consecutive_empty: 9,
            last: "before".into(),
        };
        assert_eq!(step(&at_limit, &t, 6, NOW), (at_limit.clone(), false));
        // Unreadable reads between empty ones do not reset or add.
        let t = tally(&[Read::Empty, Read::Unreadable, Read::Empty]);
        assert_eq!(t.empty, 2);
        // The budget a `--wait` call stops on agrees with `step`.
        let b = IdleBudget { max: 6, already: 4 };
        assert!(!b.reached(&tally(&[Read::Empty])));
        assert!(b.reached(&tally(&[Read::Empty, Read::Empty])));
        assert!(
            !b.reached(&tally(&[Read::Unreadable; 5])),
            "unreadable never"
        );
        assert!(
            !b.reached(&tally(&[Read::Work, Read::Empty, Read::Empty])),
            "work restarts the count from 0"
        );
        assert!(!IdleBudget { max: 0, already: 0 }.reached(&tally(&[Read::Empty])));
        assert!(
            !IdleBudget { max: 6, already: 9 }.reached(&tally(&[Read::Unreadable])),
            "a count already past the limit is not reached again by a read that counted nothing"
        );
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
    fn a_read_with_work_in_flight_is_not_empty() {
        assert_eq!(classify(&empty_read()), Read::Empty);
        let with = |k: &str, v: Value| {
            let mut r = empty_read();
            r[k] = v;
            r
        };
        assert_eq!(
            classify(&with("waiting", json!([{"kind": "wait"}]))),
            Read::Work
        );
        assert_eq!(
            classify(&with("next", json!({"kind": "audit"}))),
            Read::Work
        );
        // An own pull request with no blocker (CI or an audit pending) is work in flight; a
        // foreign draft or a reserved audit that is not the lane's is not.
        assert_eq!(
            classify(&with(
                "pr_skipped",
                json!([{"pr": 10, "reason": "own_no_blocker"}])
            )),
            Read::Work
        );
        assert_eq!(
            classify(&with(
                "pr_skipped",
                json!([{"pr": 11, "reason": "draft"}, {"pr": 12, "reason": "audit_reserved"}])
            )),
            Read::Empty
        );
        // The lane's own draft pull request is work in flight too: `own_open_prs` lists it
        // (its `pr_skipped` reason is `draft`), an empty list is not work.
        assert_eq!(
            classify(&with("own_open_prs", json!([20]))),
            Read::Work,
            "an own draft PR"
        );
        let mut draft = with("own_open_prs", json!([20]));
        draft["pr_skipped"] = json!([{"pr": 20, "reason": "draft"}]);
        draft["partial"] = json!(["open pull requests: HTTP 502"]);
        assert_eq!(
            classify(&draft),
            Read::Work,
            "work wins over a partial read"
        );
        assert_eq!(classify(&with("own_open_prs", json!([]))), Read::Empty);
    }

    /// A read GitHub answered in part, or whose lane-wide end is unknown, says nothing about
    /// whether the lane has work: it is neither empty nor work. Work still wins over it.
    #[test]
    fn a_partial_or_unknown_read_is_unreadable() {
        let with = |k: &str, v: Value| {
            let mut r = empty_read();
            r[k] = v;
            r
        };
        assert_eq!(
            classify(&with("partial", json!(["open pull requests: HTTP 502"]))),
            Read::Unreadable
        );
        assert_eq!(
            classify(&with(
                "end_condition",
                json!({"met": null, "status": "partial"})
            )),
            Read::Unreadable
        );
        // "No tip base" is there whenever nothing is open, the very state of an idle lane: it
        // alone leaves the read empty; next to another note the read is unreadable again.
        let tip = crate::lane_next::TIP_BASE_UNKNOWN;
        assert_eq!(
            classify(&with("partial", json!([tip]))),
            Read::Empty,
            "a lane with nothing open can go idle"
        );
        assert_eq!(
            classify(&with("partial", json!([tip, "goal exclusions: x"]))),
            Read::Unreadable
        );
        // The comment window of a busy Issue is on every real read; it alone is not a failure,
        // and next to a real failure the read is unreadable.
        let window = "#1975: only the newest 100 comments read";
        assert_eq!(
            classify(&with(
                "partial",
                json!([window, "PR #3: only the newest 100 comments read", tip])
            )),
            Read::Empty
        );
        assert_eq!(
            classify(&with(
                "partial",
                json!([window, "open pull requests: HTTP 502"])
            )),
            Read::Unreadable
        );
        assert_eq!(
            classify(&with(
                "partial",
                json!(["#7: only the newest 50 comments read"])
            )),
            Read::Unreadable,
            "another window size is not the known note"
        );
        // Only the exact note counts: a real failure that happens to end with the phrase, or
        // a subject that is not an Issue or pull request number, stays unreadable.
        for bad in [
            "lists: HTTP 502 only the newest 100 comments read",
            "lists: only the newest 100 comments read",
            "#x: only the newest 100 comments read",
            "#12:only the newest 100 comments read",
            "#: only the newest 100 comments read",
            "PR : only the newest 100 comments read",
            "#12: only the newest 100 comments read, HTTP 502",
            "x #12: only the newest 100 comments read",
        ] {
            assert_eq!(
                classify(&with("partial", json!([bad]))),
                Read::Unreadable,
                "{bad}"
            );
        }
        for good in [
            "#1975: only the newest 100 comments read",
            "PR #3: only the newest 100 comments read",
        ] {
            assert_eq!(
                classify(&with("partial", json!([good]))),
                Read::Empty,
                "{good}"
            );
        }
        assert_eq!(classify(&json!({"next": null})), Read::Unreadable);
        assert_eq!(classify(&json!({})), Read::Unreadable);
        let mut r = with("partial", json!(["x"]));
        r["waiting"] = json!([{"kind": "wait"}]);
        assert_eq!(classify(&r), Read::Work, "work wins over a partial read");
        assert_eq!(
            classify(&with("end_condition", json!({"met": true}))),
            Read::Empty
        );
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
        std::os::unix::fs::symlink(&elsewhere, anchor.join(crate::fixed_state!())).unwrap();
        let mut v = empty_read();
        apply(&mut v, Some(&anchor), "agent-1", "o/r", 6, NOW);
        assert_eq!(v["idle"]["counting"], json!(false), "{v}");
        assert!(v.get("status").is_none());
        assert_eq!(
            std::fs::read_dir(&elsewhere).unwrap().count(),
            0,
            "nothing written through the link"
        );
    }

    /// Three partial reads never reach the limit and write nothing (no file, the saved count
    /// stays); three reads with an own pull request in flight never reach it either (work
    /// keeps the count at 0). The tally `run` attaches is used and removed from the output.
    #[test]
    fn apply_does_not_count_an_unreadable_read_or_work_in_flight() {
        let d = tempfile::tempdir().unwrap();
        let anchor = d.path();
        let file = state_file(anchor, "agent-1", "o/r");
        for _ in 0..3 {
            let mut v = empty_read();
            v["partial"] = json!(["open pull requests: HTTP 502"]);
            apply(&mut v, Some(anchor), "agent-1", "o/r", 3, NOW);
            assert_eq!(v["idle"]["consecutive_empty"], json!(0), "{v}");
            assert_eq!(v["idle"]["limit_reached"], json!(false));
            assert!(v["idle"]["not_counted"].is_string(), "{v}");
            assert!(v.get("status").is_none() && v["idle"].get("line").is_none());
            assert!(!file.exists(), "an unreadable read writes nothing");
        }
        for _ in 0..3 {
            let mut v = empty_read();
            v["pr_skipped"] = json!([{"pr": 10, "reason": "own_no_blocker"}]);
            apply(&mut v, Some(anchor), "agent-1", "o/r", 3, NOW);
            assert_eq!(v["idle"]["consecutive_empty"], json!(0), "{v}");
            assert!(v.get("status").is_none(), "own PR in flight: {v}");
        }
        // Nothing was saved above. A `--wait` call that counted three empty reads (the tally
        // `run` attaches) is the limit although the result itself is a single read.
        let mut v = empty_read();
        attach(&mut v, &tally(&[Read::Empty; 3]));
        apply(&mut v, Some(anchor), "agent-1", "o/r", 3, NOW);
        assert_eq!(v["status"], json!("idle_limit"), "{v}");
        assert_eq!(v["idle"]["consecutive_empty"], json!(3), "{v}");
        assert!(v.get("idle_read").is_none(), "the tally is not output");
    }

    /// The budget carries the saved count of this lane and repository only; no clone or
    /// `--max-empty 0` has no budget.
    #[test]
    fn the_budget_is_the_saved_count_of_this_lane_and_repository() {
        let d = tempfile::tempdir().unwrap();
        let anchor = d.path();
        let none = budget(Some(anchor), "agent-1", "o/r", 6).unwrap();
        assert_eq!((none.max, none.already), (6, 0), "nothing saved yet");
        let file = state_file(anchor, "agent-1", "o/r");
        let state = IdleState {
            consecutive_empty: 4,
            last: NOW.into(),
        };
        save(anchor, &file, &state).unwrap();
        let b = budget(Some(anchor), "agent-1", "o/r", 6).unwrap();
        assert_eq!((b.max, b.already), (6, 4));
        assert_eq!(
            budget(Some(anchor), "agent-2", "o/r", 6).unwrap().already,
            0
        );
        assert_eq!(
            budget(Some(anchor), "agent-1", "o/s", 6).unwrap().already,
            0
        );
        assert!(budget(Some(anchor), "agent-1", "o/r", 0).is_none());
        assert!(budget(None, "agent-1", "o/r", 6).is_none());
    }

    #[test]
    fn apply_counts_to_the_limit_and_writes_the_line() {
        let d = tempfile::tempdir().unwrap();
        let anchor = d.path();
        for n in 1..=3u64 {
            let mut v = empty_read();
            v["writes"] = json!(false);
            apply(&mut v, Some(anchor), "agent-1", "o/r", 3, NOW);
            assert_eq!(v["idle"]["consecutive_empty"], json!(n));
            assert_eq!(v["idle"]["limit_reached"], json!(n == 3), "read {n}");
            assert_eq!(v.get("status").is_some(), n == 3);
            assert_eq!(v["writes"], json!(true));
        }
        let mut v = empty_read();
        v["next"] = json!({"kind": "issue"});
        apply(&mut v, Some(anchor), "agent-1", "o/r", 3, NOW);
        assert_eq!(v["idle"]["consecutive_empty"], json!(0));
        let mut v = empty_read();
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
