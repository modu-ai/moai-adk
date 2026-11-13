//! Usage-limit record (#2668 S1): what the Claude Code status line last saw of the claude.ai
//! usage windows, kept so a lane — which never sees the status line — can renew its claim
//! leases before a limit stops it.
//!
//! Claude Code gives the status line command `rate_limits.five_hour` and `.seven_day`, each
//! with `used_percentage` (0–100) and `resets_at` ("Unix epoch seconds"), only for claude.ai
//! subscriptions and only after the session's first response; a window may be absent, and
//! Claude Code drops a window once its `resets_at` passes
//! (https://code.claude.com/docs/en/statusline, read 2026-11-13). Once a limit is hit the
//! model can do nothing, and the `StopFailure` hook carries no reset time
//! (https://code.claude.com/docs/en/hooks, read 2026-11-13), so the record is written on every
//! render and read before the limit.
//!
//! `record` writes `$HOME/.moai-cli/state/usage-limits.json` atomically; `evaluate` reads it
//! and writes nothing. Resuming after the reset is Claude Code's own automatic continue
//! (`autoContinueAtUsageLimit`), not done here.

use crate::forge::policy::format_utc;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::path::{Path, PathBuf};

pub const SCHEMA_VERSION: u32 = 1;
/// `near` from this used percentage (proposal in #2668 comment 6848078534; one constant).
pub const NEAR_PERCENT: f64 = 90.0;
/// Seconds added to `resets_at` for the lease a lane renews to.
pub const MARGIN_SECS: i64 = 15 * 60;
const FILE: &str = "usage-limits.json";
/// The two windows Claude Code names; any other key is ignored when reading.
pub const WINDOWS: [&str; 2] = ["five_hour", "seven_day"];

#[derive(Clone, Copy, Debug, PartialEq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Window {
    pub used_percentage: f64,
    pub resets_at: i64,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Record {
    pub schema_version: u32,
    /// Unix seconds of the render that wrote the record.
    pub observed_at: i64,
    pub windows: BTreeMap<String, Window>,
}

pub fn record_path(home: &Path) -> PathBuf {
    home.join(crate::HOME_STATE_DIR).join("state").join(FILE)
}

/// A folder that must be a real folder when present: a symlink here could send the write
/// anywhere.
fn plain_dir_or_absent(path: &Path) -> Result<(), String> {
    match std::fs::symlink_metadata(path) {
        Ok(m) if m.file_type().is_symlink() => {
            Err(format!("{} is a symlink; refused", path.display()))
        }
        Ok(m) if !m.is_dir() => Err(format!("{} is not a folder", path.display())),
        Ok(_) => Ok(()),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(()),
        Err(_) => Err(format!("{} is unreadable", path.display())),
    }
}

/// The windows worth keeping: both fields present, a finite percentage in 0–100, and a reset
/// still in the future (Claude Code drops a passed window too).
pub fn windows_from(
    pairs: &[(&str, Option<f64>, Option<i64>)],
    now: i64,
) -> BTreeMap<String, Window> {
    pairs
        .iter()
        .filter_map(|(name, used, resets)| {
            let (used, resets) = ((*used)?, (*resets)?);
            (WINDOWS.contains(name)
                && used.is_finite()
                && (0.0..=100.0).contains(&used)
                && resets > now)
                .then(|| {
                    (
                        name.to_string(),
                        Window {
                            used_percentage: used,
                            resets_at: resets,
                        },
                    )
                })
        })
        .collect()
}

/// Write the record when there is at least one window; an absent `rate_limits` (API key,
/// before the first response) leaves the previous record alone, since that says nothing new.
/// Returns whether a record was written.
pub fn record(home: &Path, windows: BTreeMap<String, Window>, now: i64) -> Result<bool, String> {
    if windows.is_empty() {
        return Ok(false);
    }
    let base = home.join(crate::HOME_STATE_DIR);
    let dir = base.join("state");
    plain_dir_or_absent(&base)?;
    plain_dir_or_absent(&dir)?;
    std::fs::create_dir_all(&dir).map_err(|_| format!("{} cannot be created", dir.display()))?;
    let path = dir.join(FILE);
    if let Ok(m) = std::fs::symlink_metadata(&path) {
        if !m.is_file() {
            return Err(format!("{} is not a regular file; refused", path.display()));
        }
    }
    let rec = Record {
        schema_version: SCHEMA_VERSION,
        observed_at: now,
        windows,
    };
    let mut bytes = serde_json::to_vec(&rec).map_err(|_| "encode failed")?;
    bytes.push(b'\n');
    // A process-unique temporary name, renamed over the record: a reader sees the old or the
    // new record, never half of one, even with several status lines rendering at once.
    let tmp = dir.join(format!(".{FILE}.{}.tmp", std::process::id()));
    std::fs::write(&tmp, &bytes).map_err(|_| format!("{} cannot be written", tmp.display()))?;
    std::fs::rename(&tmp, &path).map_err(|_| {
        let _ = std::fs::remove_file(&tmp);
        format!("{} cannot be replaced", path.display())
    })?;
    Ok(true)
}

/// Read the record and judge it at `now`. Writes nothing. `state`: `unknown` (no record, an
/// unreadable one, or no window still open), `reached` (a window at 100), `near` (a window at
/// [`NEAR_PERCENT`] or more) or `ok`. For `near` and `reached`, `renew_until` is the latest
/// such window's reset plus [`MARGIN_SECS`]: the lease a lane renews its claims to.
pub fn evaluate(home: &Path, now: i64) -> Value {
    let path = record_path(home);
    let unknown = |reason: &str| {
        json!({
            "state": "unknown",
            "reason": reason,
            "record": path.to_string_lossy(),
            "windows": {},
        })
    };
    let base = home.join(crate::HOME_STATE_DIR);
    if plain_dir_or_absent(&base).is_err() || plain_dir_or_absent(&base.join("state")).is_err() {
        return unknown("state folder is not a plain folder");
    }
    let bytes = match std::fs::symlink_metadata(&path) {
        Ok(m) if m.is_file() => match std::fs::read(&path) {
            Ok(b) => b,
            Err(_) => return unknown("record unreadable"),
        },
        Ok(_) => return unknown("record is not a regular file"),
        Err(_) => return unknown("no record"),
    };
    let rec: Record = match serde_json::from_slice(&bytes) {
        Ok(r) => r,
        Err(_) => return unknown("record unreadable"),
    };
    if rec.schema_version != SCHEMA_VERSION {
        return unknown("record has an unknown schema");
    }
    let open: BTreeMap<&String, &Window> = rec
        .windows
        .iter()
        .filter(|(k, w)| WINDOWS.contains(&k.as_str()) && w.resets_at > now)
        .collect();
    if open.is_empty() {
        return unknown("no open window");
    }
    let at_or_above = |pct: f64| {
        open.iter()
            .filter(move |(_, w)| w.used_percentage >= pct)
            .map(|(k, w)| (k.as_str(), w.resets_at))
            .max_by_key(|(_, r)| *r)
    };
    let (state, hit) = match (at_or_above(100.0), at_or_above(NEAR_PERCENT)) {
        (Some(h), _) => ("reached", Some(h)),
        (None, Some(h)) => ("near", Some(h)),
        (None, None) => ("ok", None),
    };
    let windows: BTreeMap<&String, Value> = open
        .iter()
        .map(|(k, w)| {
            (
                *k,
                json!({
                    "used_percentage": w.used_percentage,
                    "resets_at": format_utc(w.resets_at),
                }),
            )
        })
        .collect();
    let mut out = json!({
        "state": state,
        "record": path.to_string_lossy(),
        "observed_at": format_utc(rec.observed_at),
        "near_percent": NEAR_PERCENT,
        "windows": windows,
    });
    if let Some((window, resets)) = hit {
        out["window"] = json!(window);
        out["resets_at"] = json!(format_utc(resets));
        out["renew_until"] = json!(format_utc(resets + MARGIN_SECS));
    }
    out
}

#[cfg(test)]
mod tests {
    use super::*;
    use tempfile::TempDir;

    const NOW: i64 = 1_800_000_000;

    fn win(used: f64, resets: i64) -> Window {
        Window {
            used_percentage: used,
            resets_at: resets,
        }
    }
    fn put(home: &Path, five: Option<Window>, seven: Option<Window>) {
        let mut w = BTreeMap::new();
        if let Some(f) = five {
            w.insert("five_hour".to_string(), f);
        }
        if let Some(s) = seven {
            w.insert("seven_day".to_string(), s);
        }
        record(home, w, NOW).unwrap();
    }

    #[test]
    fn near_starts_at_the_threshold_and_reached_at_one_hundred() {
        let home = TempDir::new().unwrap();
        let reset = NOW + 3600;
        for (used, want) in [
            (0.0, "ok"),
            (89.9, "ok"),
            (90.0, "near"),
            (99.9, "near"),
            (100.0, "reached"),
        ] {
            put(home.path(), Some(win(used, reset)), None);
            let v = evaluate(home.path(), NOW);
            assert_eq!(v["state"], want, "{used}");
            if want == "ok" {
                assert!(v.get("renew_until").is_none(), "{used}");
            } else {
                assert_eq!(v["window"], "five_hour");
                assert_eq!(v["resets_at"], format_utc(reset));
                assert_eq!(v["renew_until"], format_utc(reset + MARGIN_SECS));
            }
        }
    }

    #[test]
    fn renew_until_follows_the_latest_window_at_or_above_the_threshold() {
        let home = TempDir::new().unwrap();
        // Both near: the later reset (seven_day) decides the lease.
        put(
            home.path(),
            Some(win(95.0, NOW + 3600)),
            Some(win(91.0, NOW + 86_400)),
        );
        let v = evaluate(home.path(), NOW);
        assert_eq!(v["state"], "near");
        assert_eq!(v["window"], "seven_day");
        assert_eq!(v["renew_until"], format_utc(NOW + 86_400 + MARGIN_SECS));
        // Only five_hour near: seven_day below the threshold does not stretch the lease.
        put(
            home.path(),
            Some(win(95.0, NOW + 3600)),
            Some(win(50.0, NOW + 86_400)),
        );
        let v = evaluate(home.path(), NOW);
        assert_eq!(v["window"], "five_hour");
        assert_eq!(v["renew_until"], format_utc(NOW + 3600 + MARGIN_SECS));
    }

    #[test]
    fn a_passed_window_is_not_open_and_no_record_is_unknown() {
        let home = TempDir::new().unwrap();
        assert_eq!(evaluate(home.path(), NOW)["state"], "unknown");
        put(home.path(), Some(win(100.0, NOW + 1)), None);
        assert_eq!(evaluate(home.path(), NOW)["state"], "reached");
        // At and after the reset the window is closed: N/N+1 edge on resets_at.
        assert_eq!(evaluate(home.path(), NOW + 1)["state"], "unknown");
        assert_eq!(evaluate(home.path(), NOW + 2)["state"], "unknown");
    }

    #[test]
    fn windows_from_keeps_only_complete_open_known_windows() {
        let got = windows_from(
            &[
                ("five_hour", Some(42.0), Some(NOW + 10)),
                ("seven_day", Some(12.0), None),
                ("spend_limit", Some(99.0), Some(NOW + 10)),
            ],
            NOW,
        );
        assert_eq!(got.keys().collect::<Vec<_>>(), vec!["five_hour"]);
        for bad in [-0.1, 100.1, f64::NAN] {
            assert!(windows_from(&[("five_hour", Some(bad), Some(NOW + 10))], NOW).is_empty());
        }
        assert!(windows_from(&[("five_hour", Some(1.0), Some(NOW))], NOW).is_empty());
    }

    #[test]
    fn empty_windows_keep_the_previous_record() {
        let home = TempDir::new().unwrap();
        put(home.path(), Some(win(95.0, NOW + 3600)), None);
        assert!(!record(home.path(), BTreeMap::new(), NOW + 5).unwrap());
        assert_eq!(evaluate(home.path(), NOW)["state"], "near");
    }

    #[cfg(unix)]
    #[test]
    fn a_symlinked_state_folder_is_refused_and_reads_as_unknown() {
        let home = TempDir::new().unwrap();
        let elsewhere = TempDir::new().unwrap();
        std::fs::create_dir_all(home.path().join(crate::HOME_STATE_DIR)).unwrap();
        std::os::unix::fs::symlink(
            elsewhere.path(),
            home.path().join(crate::HOME_STATE_DIR).join("state"),
        )
        .unwrap();
        let mut w = BTreeMap::new();
        w.insert("five_hour".to_string(), win(95.0, NOW + 60));
        assert!(record(home.path(), w, NOW).is_err());
        assert_eq!(std::fs::read_dir(elsewhere.path()).unwrap().count(), 0);
        assert_eq!(evaluate(home.path(), NOW)["state"], "unknown");
    }

    #[test]
    fn an_unknown_schema_or_garbage_reads_as_unknown() {
        let home = TempDir::new().unwrap();
        let path = record_path(home.path());
        std::fs::create_dir_all(path.parent().unwrap()).unwrap();
        for text in [
            "not json",
            r#"{"schema_version":2,"observed_at":1,"windows":{}}"#,
        ] {
            std::fs::write(&path, text).unwrap();
            assert_eq!(evaluate(home.path(), NOW)["state"], "unknown", "{text}");
        }
    }
}
