//! `lane-resume` (SPEC-20261120-lane-limit-resume): a usage-limit stall ends at the reset
//! instant its screen names; this sends the lane one resume line once that instant has passed.
//!
//! The screen reads, the classification and the reset instant are `lane-screen`'s (`observe`,
//! `reset_of`, `scan`); the text and handle checks are `lane-send`'s. The send itself goes
//! through the same two-read guard: the screen is read again just before the send, and a
//! screen that stopped being a usage limit sends nothing — a confirmation prompt never
//! receives text. One run sends at most one line per stalled lane and keeps no state: a lane
//! that woke and repainted shows no stall to the next run.

use serde_json::{json, Value};

use moai_cli::forge::policy::{format_utc, parse_utc};
use moai_cli::lane_status::lane_screen::{self as screen, Reset, State};
use moai_cli::lane_status::lane_send::{check_handle, check_text, read_checked, Orca, ProcessOrca};

/// Resume a stalled lane at its usage limit's reset instant
///
/// Reads the rendered screens of the Orca terminals, keeps the lanes a usage limit stalls and
/// names each with its reset instant in UTC. A lane whose reset instant has passed gets
/// `--text` sent once through `orca terminal send`; its screen is read again just before the
/// send, and one that now shows working, idle, a confirmation prompt or anything else sends
/// nothing and is named with its new state. A lane whose reset instant is still in the future
/// is named as waiting; with `--wait` the command sleeps once to the earliest such instant (at
/// most 8 hours) and reads the lanes again before sending. A lane whose reset instant cannot
/// be told, or that cannot be read at all, is skipped with its reason. The text runs the
/// checks of `lane-send` before any `orca` call: not empty, no leading `-`, 2000 characters at
/// most, no control or bidirectional character. Exit is 0 only when at least one line was
/// sent. With `--terminal` only that terminal is considered; without it, every terminal of
/// `orca terminal list`, read one at a time like `lane-screen --scan`. Runs `orca`
/// (`MOAI_ORCA_BIN` overrides the program, `MOAI_ORCA_TIMEOUT_SECS` the 15 second limit per
/// call). Keeps no state: at most one line per stalled lane per run.
#[derive(clap::Args)]
pub struct Args {
    /// Resume only this Orca terminal handle, such as `term_123`; without it, the whole scan.
    #[arg(long, allow_hyphen_values = true)]
    terminal: Option<String>,
    /// The one resume line to send, such as the lane's own wake line.
    #[arg(long, allow_hyphen_values = true)]
    text: String,
    /// Send Enter after the text.
    #[arg(long)]
    enter: bool,
    /// Sleep once to the earliest future reset instant and read the lanes again, instead of
    /// reporting the instant and sending nothing.
    #[arg(long)]
    wait: bool,
    /// The time to take as now, `YYYY-MM-DDTHH:MM:SSZ` (default: the clock).
    #[arg(long)]
    now: Option<String>,
}

pub const OPERATION: &str = "lane.resume";

/// The longest single sleep `--wait` may take: the limit window is five hours, and a mistyped
/// future date must not park the process for a day.
const MAX_WAIT_SECS: i64 = 8 * 3600;

fn clock(now: Option<&str>) -> Result<i64, String> {
    match now {
        Some(s) => Ok(parse_utc(s).ok_or("--now must be YYYY-MM-DDTHH:MM:SSZ")?),
        None => Ok(std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .map_err(|_| "the clock is before 1970")?
            .as_secs() as i64),
    }
}

/// One lane the run looked at. `state` is the `lane-screen` state word; `reset` is the reset
/// instant in Unix seconds, `Some` only when the screen told it in UTC.
#[derive(Clone)]
struct Lane {
    handle: String,
    state: String,
    reason: String,
    evidence: Option<String>,
    reset: Option<i64>,
    unknown: Option<String>,
}

impl Lane {
    fn stalled(&self) -> bool {
        self.state == State::UsageLimit.as_str()
    }

    /// The answer row of a lane that is not a resume target.
    fn entry(&self, action: &str) -> Value {
        let mut v = json!({
            "terminal": self.handle,
            "action": action,
            "state": self.state,
            "reason": self.reason,
            "evidence": self.evidence,
        });
        if let Some(t) = self.reset {
            v["reset"] = json!({"utc": format_utc(t), "source": "screen"});
        }
        if let Some(why) = &self.unknown {
            v["reset"] = json!({"utc": null, "unknown": true, "why": why});
        }
        v
    }

    /// The one phrase that names what holds this lane back.
    fn part(&self, action: &str) -> String {
        match action {
            "waiting" => format!(
                "{} waits; the limit resets at {}",
                self.handle,
                format_utc(self.reset.unwrap_or(0))
            ),
            "skipped" => format!(
                "{} is skipped; {}{}",
                self.handle,
                self.reason,
                self.unknown
                    .as_deref()
                    .map(|w| format!("; the reset time is unknown ({w})"))
                    .unwrap_or_default()
            ),
            "refused" => format!(
                "{} is refused; its screen is {} ({})",
                self.handle, self.state, self.reason
            ),
            _ => format!(
                "{} need not be resumed; its screen is {} ({})",
                self.handle, self.state, self.reason
            ),
        }
    }
}

/// The lanes of one classification pass: what is due, what waits, what cannot be told, and
/// the lanes that are not resume targets at all.
struct Pass {
    due: Vec<Lane>,
    waiting: Vec<Lane>,
    unknowable: Vec<Lane>,
    refused: Vec<Lane>,
    recovered: Vec<Lane>,
    skipped: Vec<Lane>,
}

/// Sorts the lanes of a look into the buckets of [`Pass`], at `now` (Unix seconds).
fn classify(lanes: &[Lane], now: i64) -> Pass {
    let mut pass = Pass {
        due: Vec::new(),
        waiting: Vec::new(),
        unknowable: Vec::new(),
        refused: Vec::new(),
        recovered: Vec::new(),
        skipped: Vec::new(),
    };
    for lane in lanes {
        match lane.state.as_str() {
            s if s == State::UsageLimit.as_str() => match lane.reset {
                Some(t) if t <= now => pass.due.push(lane.clone()),
                Some(_) => pass.waiting.push(lane.clone()),
                None => pass.unknowable.push(lane.clone()),
            },
            s if s == State::Unreachable.as_str() => pass.skipped.push(lane.clone()),
            s if s == State::Working.as_str() || s == State::Idle.as_str() => {
                pass.recovered.push(lane.clone())
            }
            _ => pass.refused.push(lane.clone()),
        }
    }
    pass
}

pub fn run(args: Args) -> Result<Value, String> {
    let now = clock(args.now.as_deref())?;
    let orca = ProcessOrca::from_env();
    let mut real_sleep = |secs: i64| {
        if secs > 0 {
            std::thread::sleep(std::time::Duration::from_secs(secs as u64));
        }
    };
    resume(
        &orca,
        args.terminal.as_deref(),
        &args.text,
        args.enter,
        args.wait,
        now,
        &mut real_sleep,
    )
}

/// Reads and classifies one terminal's screen, the way `lane-send` does: anything that is not
/// an Orca answer is an unreachable lane.
fn look(orca: &dyn Orca, terminal: &str) -> Result<(screen::Screen, screen::Seen), String> {
    let out = read_checked(orca, terminal)
        .map_err(|d| format!("lane resume refused: lane unreachable: {d}"))?;
    let (s, seen) = screen::observe(&out);
    if s.source != "orca:screen" {
        return Err(
            "lane resume refused: lane unreachable: the answer of `orca terminal read` is not an Orca answer"
                .into(),
        );
    }
    Ok((s, seen))
}

fn lane_from_look(terminal: &str, s: &screen::Screen, seen: &screen::Seen, now: i64) -> Lane {
    let (reset, unknown) = match screen::reset_of(&s.rows, now) {
        Some(Reset::Utc(t)) => (Some(t), None),
        Some(Reset::Unknown { why, .. }) => (None, Some(why.to_string())),
        None => (None, Some("no reset row".to_string())),
    };
    Lane {
        handle: terminal.to_string(),
        state: seen.state.as_str().to_string(),
        reason: seen.reason.to_string(),
        evidence: seen.evidence.clone(),
        reset,
        unknown,
    }
}

/// A lane that could not be read (again) is an unreachable lane with the detail as its reason.
fn unread_lane(handle: &str, detail: String) -> Lane {
    Lane {
        handle: handle.to_string(),
        state: State::Unreachable.as_str().to_string(),
        reason: detail,
        evidence: None,
        reset: None,
        unknown: None,
    }
}

/// Every lane of a scan that matters here: the stalled, the unreadable and the ones a text
/// must never reach (a prompt, an unknown screen). Healthy lanes (`working`, `idle`) stay out
/// — the scan read them and they are simply not stalls.
fn scan_lanes(orca: &dyn Orca, now: i64) -> Result<Vec<Lane>, String> {
    let list = orca.list().map_err(|d| screen::list_error(&d))?;
    let answer = screen::scan(&list, &mut |h| read_checked(orca, h), now)?;
    let mut lanes = Vec::new();
    for row in answer["terminals"].as_array().into_iter().flatten() {
        let Some(handle) = row["handle"].as_str() else {
            continue;
        };
        let state = row["state"].as_str().unwrap_or(State::Unknown.as_str());
        if state == State::Working.as_str() || state == State::Idle.as_str() {
            continue;
        }
        let utc = row["reset"]["utc"].as_str().and_then(parse_utc);
        let unknown = if utc.is_none() {
            Some(
                row["reset"]["why"]
                    .as_str()
                    .map(String::from)
                    .unwrap_or_else(|| "no reset row".to_string()),
            )
        } else {
            None
        };
        lanes.push(Lane {
            handle: handle.to_string(),
            state: state.to_string(),
            reason: row["reason"].as_str().unwrap_or("").to_string(),
            evidence: row["evidence"].as_str().map(String::from),
            reset: utc,
            unknown,
        });
    }
    Ok(lanes)
}

/// The answer of `orca terminal send --json`, read the way `lane-send` reads it: only an
/// object whose `ok` is `true` counts as sent. `ok: false` and anything else is a failure that
/// says the text may have reached the lane.
fn send_answer(answer: &str) -> Result<Value, String> {
    let value: Value = serde_json::from_str(answer).unwrap_or(Value::Null);
    match value.get("ok") {
        Some(Value::Bool(true)) => Ok(value),
        Some(Value::Bool(false)) => {
            let line = value
                .pointer("/error/message")
                .or_else(|| value.get("error"))
                .and_then(Value::as_str)
                .and_then(|t| t.lines().map(str::trim).find(|l| !l.is_empty()))
                .map(|l| {
                    l.chars()
                        .filter(|c| !c.is_control())
                        .take(160)
                        .collect::<String>()
                })
                .unwrap_or_default();
            Err(format!(
                "lane resume failed: `orca terminal send` answered ok:false {line}; look at the lane before sending again"
            ))
        }
        _ => Err(
            "lane resume failed: the answer of `orca terminal send` is not an Orca answer (no `ok: true`); the text may have reached the lane"
                .into(),
        ),
    }
}

/// The whole flow. `sleep` is the waiter of `--wait`; a test passes one that records.
fn resume(
    orca: &dyn Orca,
    terminal: Option<&str>,
    text: &str,
    enter: bool,
    wait: bool,
    now: i64,
    sleep: &mut dyn FnMut(i64),
) -> Result<Value, String> {
    // The text is checked before any `orca` call is made.
    check_text(text).map_err(|d| {
        format!(
            "lane resume refused: {}",
            d.trim_start_matches("lane send refused: ")
        )
    })?;
    let mut now_eff = now;

    let lanes = match terminal {
        Some(t) => {
            check_handle(t).map_err(|d| {
                format!(
                    "lane resume refused: {}",
                    d.trim_start_matches("lane send refused: ")
                )
            })?;
            let (s, seen) = look(orca, t)?;
            vec![lane_from_look(t, &s, &seen, now_eff)]
        }
        None => scan_lanes(orca, now_eff)?,
    };
    let mut pass = classify(&lanes, now_eff);

    // `--wait`: one sleep to the earliest future instant, then every stalled lane is read
    // again and classified once more. A reset beyond the cap sleeps for no time and stays
    // waiting.
    if wait {
        if let Some(earliest) = pass.waiting.iter().filter_map(|l| l.reset).min() {
            let gap = earliest - now_eff;
            if gap > 0 && gap <= MAX_WAIT_SECS {
                sleep(gap);
                now_eff += gap;
                let mut reread: Vec<Lane> = Vec::new();
                for lane in lanes.iter().filter(|l| l.stalled()) {
                    reread.push(match look(orca, &lane.handle) {
                        Ok((s, seen)) => lane_from_look(&lane.handle, &s, &seen, now_eff),
                        Err(d) => unread_lane(&lane.handle, d),
                    });
                }
                // Only the stalled lanes were read again: the lanes that were never resume
                // targets keep their first verdicts.
                let old = std::mem::replace(&mut pass, classify(&reread, now_eff));
                pass.skipped.extend(old.skipped);
                pass.refused.extend(old.refused);
                pass.recovered.extend(old.recovered);
            }
        }
    }

    // Each due lane gets its pre-send look and, when the limit row still holds, one line.
    let mut lines: Vec<Value> = Vec::new();
    for lane in std::mem::take(&mut pass.due) {
        let (_, seen2) = match look(orca, &lane.handle) {
            Ok(pair) => pair,
            Err(d) => {
                pass.refused.push(unread_lane(&lane.handle, d));
                continue;
            }
        };
        match seen2.state {
            State::UsageLimit => {
                let answer = orca
                    .send(&lane.handle, text, enter)
                    .map_err(|d| format!("lane resume failed: {d}"))?;
                let orca_answer = send_answer(&answer)?;
                let instant = lane.reset.unwrap_or(now_eff);
                lines.push(json!({
                    "terminal": lane.handle,
                    "action": "sent",
                    "state": seen2.state.as_str(),
                    "evidence": seen2.evidence,
                    "reset": {"utc": format_utc(instant), "source": "screen"},
                    "enter": enter,
                    "orca": orca_answer,
                }));
            }
            State::Working | State::Idle => pass.recovered.push(Lane {
                handle: lane.handle.clone(),
                state: seen2.state.as_str().to_string(),
                reason: seen2.reason.to_string(),
                evidence: seen2.evidence.clone(),
                reset: None,
                unknown: None,
            }),
            _ => pass.refused.push(Lane {
                handle: lane.handle.clone(),
                state: seen2.state.as_str().to_string(),
                reason: seen2.reason.to_string(),
                evidence: seen2.evidence.clone(),
                reset: None,
                unknown: None,
            }),
        }
    }

    if lines.is_empty() {
        // Nothing was sent: every lane is named with what holds it back, and the exit is not 0.
        let mut parts: Vec<String> = Vec::new();
        for (lane, action) in pass
            .waiting
            .iter()
            .map(|l| (l, "waiting"))
            .chain(pass.unknowable.iter().map(|l| (l, "skipped")))
            .chain(pass.skipped.iter().map(|l| (l, "skipped")))
            .chain(pass.refused.iter().map(|l| (l, "refused")))
            .chain(pass.recovered.iter().map(|l| (l, "recovered")))
        {
            parts.push(lane.part(action));
        }
        return Err(format!(
            "lane resume waiting: nothing was sent{}",
            if parts.is_empty() {
                String::new()
            } else {
                format!(" ({})", parts.join("; "))
            }
        ));
    }

    Ok(json!({
        "sent": true,
        "lines": lines,
        "waiting": pass.waiting.iter().map(|l| l.entry("waiting")).collect::<Vec<_>>(),
        "refused": pass.refused.iter().map(|l| l.entry("refused")).collect::<Vec<_>>(),
        "recovered": pass.recovered.iter().map(|l| l.entry("recovered")).collect::<Vec<_>>(),
        "skipped": pass.skipped.iter().map(|l| l.entry("skipped")).collect::<Vec<_>>(),
    }))
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::cell::{Cell, RefCell};

    const NOW: i64 = 1_795_161_600; // 2026-11-20T08:00:00Z (date -u -j -f verified)
    const HANDLE: &str = "term_lane9";

    /// An Orca `terminal read --screen --json` answer holding `rows`.
    fn read_answer(rows: &[&str]) -> String {
        json!({
            "id": "fixture", "ok": true,
            "result": {"terminal": {
                "handle": HANDLE, "status": "running", "tail": rows,
                "truncated": false, "limited": false,
                "returnedLineCount": rows.len(), "source": "screen",
            }}
        })
        .to_string()
    }

    fn limit_row(instant: &str) -> String {
        read_answer(&[
            "Working (5m · esc to interrupt)",
            "",
            &format!("Usage limit reached for 5 hour. Your limit will reset at {instant}"),
            "",
        ])
    }

    /// A fake `Orca`: the N-th read is `reads[N-1]` (the last repeats), every send is recorded.
    struct FakeOrca {
        reads: Vec<String>,
        sent: RefCell<Vec<(String, String, bool)>>,
        send_answer: Option<&'static str>,
        reads_made: Cell<usize>,
    }

    impl FakeOrca {
        fn one(read: String) -> FakeOrca {
            FakeOrca {
                reads: vec![read],
                sent: RefCell::new(Vec::new()),
                send_answer: None,
                reads_made: Cell::new(0),
            }
        }

        fn sent_now(&self) -> Vec<(String, String, bool)> {
            self.sent.borrow().clone()
        }
    }

    impl Orca for FakeOrca {
        fn read(&self, _terminal: &str) -> Result<String, String> {
            let n = self.reads_made.get();
            self.reads_made.set(n + 1);
            let idx = n.min(self.reads.len() - 1);
            Ok(self.reads[idx].clone())
        }

        fn send(&self, terminal: &str, text: &str, enter: bool) -> Result<String, String> {
            self.sent
                .borrow_mut()
                .push((terminal.to_string(), text.to_string(), enter));
            match self.send_answer {
                Some(answer) => Ok(answer.to_string()),
                None => {
                    Ok(json!({"id":"x","ok":true,"result":{"send":{"accepted":true}}}).to_string())
                }
            }
        }

        fn list(&self) -> Result<String, String> {
            Ok(r#"{"ok":true,"result":{"terminals":[]}}"#.to_string())
        }
    }

    fn resume_one(
        orca: &FakeOrca,
        do_wait: bool,
        now: i64,
        slept: &mut Vec<i64>,
    ) -> Result<Value, String> {
        let mut sleep = |secs: i64| slept.push(secs);
        resume(
            orca,
            Some(HANDLE),
            "wake line",
            false,
            do_wait,
            now,
            &mut sleep,
        )
    }

    /// `--wait` sleeps exactly once, to the reset instant, and then sends once.
    #[test]
    fn wait_sleeps_once_to_the_reset_and_sends_once() {
        // 2026-11-20 18:23 (UTC+8) is 10:23Z, two hours twenty-three minutes after NOW.
        let orca = FakeOrca::one(limit_row("2026-11-20 18:23 (UTC+8)"));
        let mut slept = Vec::new();
        let out = resume_one(&orca, true, NOW, &mut slept);
        assert!(out.is_ok(), "{out:?}");
        assert_eq!(slept, vec![2 * 3600 + 23 * 60], "one sleep, to the instant");
        assert_eq!(orca.sent_now().len(), 1, "one line, after the sleep");
    }

    /// A reset beyond the cap sleeps for no time and sends nothing.
    #[test]
    fn the_wait_cap_blocks_the_sleep() {
        let orca = FakeOrca::one(limit_row("2026-11-21 18:23 (UTC+8)"));
        let mut slept = Vec::new();
        let out = resume_one(&orca, true, NOW, &mut slept);
        assert!(out.is_err(), "{out:?}");
        assert!(slept.is_empty(), "no sleep past the cap");
        assert!(orca.sent_now().is_empty(), "nothing was sent");
        assert!(out.unwrap_err().contains("2026-11-21T10:23:00Z"));
    }

    /// An `ok: false` send answer is never reported sent.
    #[test]
    fn an_ok_false_send_answer_is_not_reported_sent() {
        let mut orca = FakeOrca::one(limit_row("2026-11-20 16:23 (UTC+8)"));
        orca.send_answer = Some(r#"{"id":"x","ok":false,"error":{"message":"no such terminal"}}"#);
        let mut slept = Vec::new();
        let out = resume_one(&orca, false, NOW + 3 * 3600, &mut slept);
        assert!(out.is_err(), "{out:?}");
        let message = out.unwrap_err();
        assert!(message.contains("ok:false"), "{message}");
        assert_eq!(orca.sent_now().len(), 1, "the call was made, not reported");
    }

    /// The pre-send look refuses a lane that woke while the run was deciding.
    #[test]
    fn a_lane_that_woke_before_the_send_gets_nothing() {
        let idle = read_answer(&["Idle — waiting for work", ""]);
        let orca = FakeOrca {
            reads: vec![limit_row("2026-11-20 16:23 (UTC+8)"), idle],
            sent: RefCell::new(Vec::new()),
            send_answer: None,
            reads_made: Cell::new(0),
        };
        let mut slept = Vec::new();
        let out = resume_one(&orca, false, NOW + 3 * 3600, &mut slept);
        assert!(out.is_err(), "{out:?}");
        assert!(out.unwrap_err().contains("idle"), "the new state is named");
        assert!(orca.sent_now().is_empty(), "nothing was sent");
    }

    /// The state words the flow matches on are the ones `lane-screen` names.
    #[test]
    fn the_state_words_pair_with_lane_screen() {
        assert_eq!(State::UsageLimit.as_str(), "usage_limit");
        assert_eq!(State::Unreachable.as_str(), "unreachable");
        assert_eq!(State::PermissionPrompt.as_str(), "permission_prompt");
    }
}
