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
//! (https://code.claude.com/docs/en/hooks, read 2026-11-13), so the record is kept by every
//! render that sees the windows and read before the limit.
//!
//! Where it lives (SPEC-20261122-statusline-limits REQ-7, D-5): in the status line's cache
//! folder (`$TMPDIR/moai-statusline/`, `/tmp/moai-statusline-<uid>/` when `TMPDIR` is unset or
//! empty, none when it is relative: [`crate::statusline::state::cache_folder`]), one file per
//! HOME named `usage-<first 12 hex digits of the SHA-256 of the canonical HOME path>.record`.
//! Only a system with Unix permissions has the private folder the record needs; elsewhere
//! (Windows) no record is written and every read is `unknown` (the SPEC's D-7). The suffix is
//! not `.json` on purpose: the session cache files of the same folder are `<session id>.json`
//! and its clean-up removes old `.json` files, so the record is neither a session's name nor
//! cleaned away however long it stays unchanged. Lane profiles
//! run with their own HOME and account, so the HOME the status line runs under and the `home` a
//! reader is given name the same file; a reader looks in its own `TMPDIR`, which must therefore be
//! the writer's (on macOS it is the user's own `TMPDIR`). Nothing is written or read under HOME;
//! the earlier record in the retired HOME state folder is neither written nor read, so before
//! the first render after the move a lane sees no record, the `unknown` state as before any
//! render.
//!
//! [`record`] writes only when a percentage or a reset time differs from the record on disk (or
//! the file is missing or unreadable), through a 0600 temporary file renamed over the record in
//! a folder [`crate::statusline::state::private_folder`] accepts (0700, the user's own, never a
//! link); an unchanged render leaves the file and its modified time alone, so `observed_at` is
//! the render that last changed it. [`evaluate`] reads it and writes nothing. Resuming after the
//! reset is Claude Code's own automatic continue (`autoContinueAtUsageLimit`), not done here.

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
/// The record's file name is this, the HOME digest and [`FILE_SUFFIX`].
const FILE_PREFIX: &str = "usage-";
/// Not `.json`: see the module note (the session cache's names and clean-up).
const FILE_SUFFIX: &str = ".record";
/// Hex digits of the HOME digest in the file name.
const HOME_DIGITS: usize = 12;
/// The two windows Claude Code names; any other key is ignored when reading.
pub const WINDOWS: [&str; 2] = ["five_hour", "seven_day"];
/// The furthest a window's reset may lie ahead of the render and still be kept: the longest
/// window is seven days, so 31 days is generous, and an absurd `resets_at` (one that would
/// overflow `resets_at + MARGIN_SECS` or the date printing) is never recorded or judged.
pub const MAX_RESET_AHEAD_SECS: i64 = 31 * 86_400;

/// A reset after `now` and no further ahead than [`MAX_RESET_AHEAD_SECS`].
fn reset_in_range(resets_at: i64, now: i64) -> bool {
    resets_at > now && resets_at <= now.saturating_add(MAX_RESET_AHEAD_SECS)
}

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

/// The record of `home` in the status line's cache folder for the `TMPDIR` value `tmpdir` (see
/// the module text). `None` for a relative `TMPDIR` or a `home` without a canonical path (one
/// that does not exist). Makes nothing.
pub fn record_path(home: &Path, tmpdir: Option<&str>) -> Option<PathBuf> {
    use sha2::{Digest, Sha256};
    let folder = crate::statusline::state::cache_folder(tmpdir)?;
    let canonical = std::fs::canonicalize(home).ok()?;
    let hex = format!(
        "{:x}",
        Sha256::digest(canonical.as_os_str().as_encoded_bytes())
    );
    Some(folder.join(format!("{FILE_PREFIX}{}{FILE_SUFFIX}", &hex[..HOME_DIGITS])))
}

/// The record at `path`, or why there is none: missing, not a regular file (a link is not
/// followed), unreadable or of another schema.
fn read_record(path: &Path) -> Result<Record, &'static str> {
    let bytes = match std::fs::symlink_metadata(path) {
        Ok(m) if m.is_file() => std::fs::read(path).map_err(|_| "record unreadable")?,
        Ok(_) => return Err("record is not a regular file"),
        Err(_) => return Err("no record"),
    };
    let rec: Record = serde_json::from_slice(&bytes).map_err(|_| "record unreadable")?;
    if rec.schema_version != SCHEMA_VERSION {
        return Err("record has an unknown schema");
    }
    Ok(rec)
}

/// The windows worth keeping: both fields present, a finite percentage in 0–100, and a reset
/// still in the future (Claude Code drops a passed window too) but not past
/// [`MAX_RESET_AHEAD_SECS`].
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
                && reset_in_range(resets, now))
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

/// Write the record of `home` in the cache folder for `tmpdir` when there is at least one window
/// and a percentage or a reset time differs from the record on disk; an absent `rate_limits`
/// (API key, before the first response) leaves the previous record alone, since that says
/// nothing new. Returns whether a record was written; `Err` when there is no cache folder this
/// user alone may enter or the file cannot be written, and then nothing changed.
pub fn record(
    home: &Path,
    tmpdir: Option<&str>,
    windows: BTreeMap<String, Window>,
    now: i64,
) -> Result<bool, String> {
    use std::io::Write;
    if windows.is_empty() {
        return Ok(false);
    }
    let path = record_path(home, tmpdir).ok_or("no cache folder for this TMPDIR and HOME")?;
    let dir = path.parent().ok_or("no cache folder")?;
    if !crate::statusline::state::private_folder(dir) {
        return Err(format!(
            "{} is not a private folder; refused",
            dir.display()
        ));
    }
    if read_record(&path).is_ok_and(|on_disk| on_disk.windows == windows) {
        return Ok(false);
    }
    // A link at the record's name is replaced, not followed: the rename below swaps the folder
    // entry itself, and the folder is this user's alone. A folder there is refused.
    if std::fs::symlink_metadata(&path).is_ok_and(|m| m.is_dir()) {
        return Err(format!("{} is a folder; refused", path.display()));
    }
    let rec = Record {
        schema_version: SCHEMA_VERSION,
        observed_at: now,
        windows,
    };
    let mut bytes = serde_json::to_vec(&rec).map_err(|_| "encode failed")?;
    bytes.push(b'\n');
    // A process-unique temporary name made new (never through a link), renamed over the record:
    // a reader sees the old or the new record, never half of one, even with several status lines
    // rendering at once.
    let name = path.file_name().map(|n| n.to_string_lossy().into_owned());
    let tmp = dir.join(format!(
        ".{}.{}.tmp",
        name.unwrap_or_default(),
        std::process::id()
    ));
    let mut options = std::fs::OpenOptions::new();
    options.write(true).create_new(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        options.mode(0o600);
    }
    let written = options
        .open(&tmp)
        .and_then(|mut f| f.write_all(&bytes))
        .and_then(|()| std::fs::rename(&tmp, &path));
    if written.is_err() {
        let _ = std::fs::remove_file(&tmp);
        return Err(format!("{} cannot be written", path.display()));
    }
    Ok(true)
}

/// Read the record of `home` in the cache folder for `tmpdir` and judge it at `now`. Writes
/// nothing. `state`: `unknown` (no cache folder, no record, an unreadable one, or no window
/// still open), `reached` (a window at 100), `near` (a window at [`NEAR_PERCENT`] or more) or
/// `ok`. For `near` and `reached`, `renew_until` is the latest such window's reset plus
/// [`MARGIN_SECS`]: the lease a lane renews its claims to.
pub fn evaluate(home: &Path, tmpdir: Option<&str>, now: i64) -> Value {
    let path = record_path(home, tmpdir);
    let shown = path.as_ref().map(|p| p.to_string_lossy().into_owned());
    let unknown = |reason: &str| {
        json!({
            "state": "unknown",
            "reason": reason,
            "record": shown,
            "windows": {},
        })
    };
    let Some(path) = path.as_deref() else {
        return unknown("no record: no cache folder for this TMPDIR and HOME");
    };
    let dir = path.parent().unwrap_or(path);
    if std::fs::symlink_metadata(dir).is_err() {
        return unknown("no record");
    }
    if !crate::statusline::state::is_private_folder(dir) {
        return unknown("cache folder is not a private folder");
    }
    let rec = match read_record(path) {
        Ok(rec) => rec,
        Err(reason) => return unknown(reason),
    };
    let open: BTreeMap<&String, &Window> = rec
        .windows
        .iter()
        .filter(|(k, w)| WINDOWS.contains(&k.as_str()) && reset_in_range(w.resets_at, now))
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
        "record": shown,
        "observed_at": format_utc(rec.observed_at),
        "near_percent": NEAR_PERCENT,
        "windows": windows,
    });
    if let Some((window, resets)) = hit {
        out["window"] = json!(window);
        out["resets_at"] = json!(format_utc(resets));
        out["renew_until"] = json!(format_utc(resets.saturating_add(MARGIN_SECS)));
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
    /// A temporary HOME and a temporary `TMPDIR` for it.
    struct Place {
        home: TempDir,
        tmp: TempDir,
    }
    impl Place {
        fn new() -> Place {
            Place {
                home: TempDir::new().unwrap(),
                tmp: TempDir::new().unwrap(),
            }
        }
        fn path(&self) -> &Path {
            self.home.path()
        }
        fn tmpdir(&self) -> Option<&str> {
            self.tmp.path().to_str()
        }
        fn record(&self, w: BTreeMap<String, Window>) -> Result<bool, String> {
            record(self.path(), self.tmpdir(), w, NOW)
        }
        fn evaluate(&self, now: i64) -> Value {
            evaluate(self.path(), self.tmpdir(), now)
        }
        fn file(&self) -> PathBuf {
            record_path(self.path(), self.tmpdir()).unwrap()
        }
    }
    fn windows(five: Option<Window>, seven: Option<Window>) -> BTreeMap<String, Window> {
        let mut w = BTreeMap::new();
        if let Some(f) = five {
            w.insert("five_hour".to_string(), f);
        }
        if let Some(s) = seven {
            w.insert("seven_day".to_string(), s);
        }
        w
    }
    fn put(home: &Place, five: Option<Window>, seven: Option<Window>) {
        home.record(windows(five, seven)).unwrap();
    }

    #[test]
    fn near_starts_at_the_threshold_and_reached_at_one_hundred() {
        let home = Place::new();
        let reset = NOW + 3600;
        for (used, want) in [
            (0.0, "ok"),
            (89.9, "ok"),
            (90.0, "near"),
            (99.9, "near"),
            (100.0, "reached"),
        ] {
            put(&home, Some(win(used, reset)), None);
            let v = home.evaluate(NOW);
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
        let home = Place::new();
        // Both near: the later reset (seven_day) decides the lease.
        put(
            &home,
            Some(win(95.0, NOW + 3600)),
            Some(win(91.0, NOW + 86_400)),
        );
        let v = home.evaluate(NOW);
        assert_eq!(v["state"], "near");
        assert_eq!(v["window"], "seven_day");
        assert_eq!(v["renew_until"], format_utc(NOW + 86_400 + MARGIN_SECS));
        // Only five_hour near: seven_day below the threshold does not stretch the lease.
        put(
            &home,
            Some(win(95.0, NOW + 3600)),
            Some(win(50.0, NOW + 86_400)),
        );
        let v = home.evaluate(NOW);
        assert_eq!(v["window"], "five_hour");
        assert_eq!(v["renew_until"], format_utc(NOW + 3600 + MARGIN_SECS));
    }

    #[test]
    fn a_passed_window_is_not_open_and_no_record_is_unknown() {
        let home = Place::new();
        assert_eq!(home.evaluate(NOW)["state"], "unknown");
        assert_eq!(home.evaluate(NOW)["reason"], "no record");
        put(&home, Some(win(100.0, NOW + 1)), None);
        assert_eq!(home.evaluate(NOW)["state"], "reached");
        // At and after the reset the window is closed: N/N+1 edge on resets_at.
        assert_eq!(home.evaluate(NOW + 1)["state"], "unknown");
        assert_eq!(home.evaluate(NOW + 2)["state"], "unknown");
    }

    /// REQ-7: only a changed percentage or reset time, or a missing or unreadable record, is
    /// written; the same windows leave the file and its modified time alone.
    #[test]
    fn the_record_is_written_only_when_a_window_changed() {
        let home = Place::new();
        let first = windows(Some(win(42.0, NOW + 3600)), None);
        assert_eq!(home.record(first.clone()), Ok(true));
        let old = std::time::UNIX_EPOCH + std::time::Duration::from_secs(1_000_000_000);
        std::fs::File::options()
            .write(true)
            .open(home.file())
            .unwrap()
            .set_modified(old)
            .unwrap();
        assert_eq!(home.record(first.clone()), Ok(false));
        assert_eq!(
            std::fs::metadata(home.file()).unwrap().modified().unwrap(),
            old
        );
        // A later render of the same windows: observed_at stays the render that changed it.
        assert_eq!(
            record(home.path(), home.tmpdir(), first.clone(), NOW + 60),
            Ok(false)
        );
        assert_eq!(
            home.evaluate(NOW)["observed_at"],
            format_utc(NOW),
            "unchanged"
        );
        for changed in [
            windows(Some(win(43.0, NOW + 3600)), None),
            windows(Some(win(43.0, NOW + 3601)), None),
            windows(Some(win(43.0, NOW + 3601)), Some(win(1.0, NOW + 86_400))),
        ] {
            assert_eq!(home.record(changed.clone()), Ok(true), "{changed:?}");
            assert_eq!(home.record(changed), Ok(false));
        }
        std::fs::write(home.file(), "not json").unwrap();
        assert_eq!(
            home.record(first),
            Ok(true),
            "an unreadable record is replaced"
        );
        assert_eq!(home.evaluate(NOW)["state"], "ok");
    }

    /// The file rule: `usage-` and 12 hex digits of the canonical HOME path's SHA-256, in the
    /// cache folder of `TMPDIR`; another HOME has another file; no canonical HOME or a relative
    /// `TMPDIR` names no file, and then nothing is written and the reader says `unknown`.
    #[test]
    fn the_record_file_is_named_by_the_canonical_home_in_the_cache_folder() {
        let home = Place::new();
        let file = home.file();
        assert_eq!(
            file.parent(),
            crate::statusline::state::cache_folder(home.tmpdir()).as_deref()
        );
        let name = file.file_name().unwrap().to_str().unwrap();
        let hex = name
            .strip_prefix("usage-")
            .and_then(|r| r.strip_suffix(".record"))
            .unwrap();
        assert_eq!(hex.len(), 12, "{name}");
        assert!(hex.bytes().all(|b| matches!(b, b'0'..=b'9' | b'a'..=b'f')));
        let other = TempDir::new().unwrap();
        assert_ne!(record_path(other.path(), home.tmpdir()), Some(file.clone()));
        // The same HOME through another spelling names the same file.
        let dotted = home.path().join(".");
        assert_eq!(record_path(&dotted, home.tmpdir()), Some(file.clone()));
        #[cfg(unix)]
        {
            let link = other.path().join("home-link");
            std::os::unix::fs::symlink(home.path(), &link).unwrap();
            assert_eq!(record_path(&link, home.tmpdir()), Some(file));
        }
        let gone = home.path().join("missing");
        assert_eq!(record_path(&gone, home.tmpdir()), None);
        for relative in ["rel", "./rel"] {
            assert_eq!(record_path(home.path(), Some(relative)), None);
            let w = windows(Some(win(95.0, NOW + 60)), None);
            assert!(record(home.path(), Some(relative), w, NOW).is_err());
            assert_eq!(
                evaluate(home.path(), Some(relative), NOW)["state"],
                "unknown"
            );
        }
        assert_eq!(std::fs::read_dir(home.path()).unwrap().count(), 0);
    }

    /// The cache folder must be a private folder of this user: a link in its place is refused
    /// for writing and reads as unknown, and nothing reaches the link's target.
    #[cfg(unix)]
    #[test]
    fn a_linked_cache_folder_is_refused_and_reads_as_unknown() {
        let home = Place::new();
        let elsewhere = TempDir::new().unwrap();
        let folder = crate::statusline::state::cache_folder(home.tmpdir()).unwrap();
        std::os::unix::fs::symlink(elsewhere.path(), &folder).unwrap();
        let w = windows(Some(win(95.0, NOW + 60)), None);
        assert!(home.record(w).is_err());
        assert_eq!(std::fs::read_dir(elsewhere.path()).unwrap().count(), 0);
        assert_eq!(home.evaluate(NOW)["state"], "unknown");
        assert_eq!(
            home.evaluate(NOW)["reason"],
            "cache folder is not a private folder"
        );
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
        let home = Place::new();
        put(&home, Some(win(95.0, NOW + 3600)), None);
        assert!(!record(home.path(), home.tmpdir(), BTreeMap::new(), NOW + 5).unwrap());
        assert_eq!(home.evaluate(NOW)["state"], "near");
    }

    /// D-5: the retired HOME state folder plays no part, even as a link to elsewhere: the
    /// record goes to the cache folder and nothing reaches the link's target.
    #[cfg(unix)]
    #[test]
    fn a_symlinked_home_state_folder_is_never_followed() {
        let home = Place::new();
        let elsewhere = TempDir::new().unwrap();
        std::fs::create_dir_all(home.path().join(crate::HOME_STATE_DIR)).unwrap();
        std::os::unix::fs::symlink(
            elsewhere.path(),
            home.path().join(crate::HOME_STATE_DIR).join("state"),
        )
        .unwrap();
        let w = windows(Some(win(95.0, NOW + 60)), None);
        assert_eq!(home.record(w), Ok(true));
        assert_eq!(std::fs::read_dir(elsewhere.path()).unwrap().count(), 0);
        assert_eq!(home.evaluate(NOW)["state"], "near");
    }

    #[test]
    fn an_unknown_schema_or_garbage_reads_as_unknown() {
        let home = Place::new();
        put(&home, Some(win(95.0, NOW + 60)), None);
        let path = home.file();
        for text in [
            "not json",
            r#"{"schema_version":2,"observed_at":1,"windows":{}}"#,
        ] {
            std::fs::write(&path, text).unwrap();
            assert_eq!(home.evaluate(NOW)["state"], "unknown", "{text}");
        }
    }

    /// The reader refuses a cache folder that others may enter: with group or other rights on
    /// it the record reads as unknown although the record itself is fine.
    #[cfg(unix)]
    #[test]
    fn a_cache_folder_others_may_enter_reads_as_unknown() {
        use std::os::unix::fs::PermissionsExt;
        let home = Place::new();
        put(&home, Some(win(95.0, NOW + 60)), None);
        let folder = home.file().parent().unwrap().to_path_buf();
        assert_eq!(home.evaluate(NOW)["state"], "near");
        for mode in [0o750, 0o705] {
            std::fs::set_permissions(&folder, std::fs::Permissions::from_mode(mode)).unwrap();
            let v = home.evaluate(NOW);
            assert_eq!(v["state"], "unknown", "{mode:o}");
            assert_eq!(
                v["reason"], "cache folder is not a private folder",
                "{mode:o}"
            );
        }
    }

    /// A reset further ahead than the longest window allows is neither recorded nor judged,
    /// and an absurd one never overflows (a plain `resets_at + MARGIN_SECS` panicked in a debug
    /// build); the edge itself is kept.
    #[test]
    fn a_reset_too_far_ahead_is_left_out_and_never_overflows() {
        let edge = NOW + MAX_RESET_AHEAD_SECS;
        for (resets, kept) in [(edge, true), (edge + 1, false), (i64::MAX, false)] {
            let w = windows_from(&[("five_hour", Some(95.0), Some(resets))], NOW);
            assert_eq!(w.len(), usize::from(kept), "{resets}");
        }
        let home = Place::new();
        put(&home, Some(win(95.0, NOW + 60)), None);
        std::fs::write(
            home.file(),
            format!(
                r#"{{"schema_version":1,"observed_at":1,"windows":{{"five_hour":{{"used_percentage":95.0,"resets_at":{}}}}}}}"#,
                i64::MAX
            ),
        )
        .unwrap();
        let v = home.evaluate(NOW);
        assert_eq!(v["state"], "unknown", "{v}");
        assert_eq!(v["reason"], "no open window", "{v}");
    }

    /// A link at the record's name is replaced by the record (the folder is this user's alone)
    /// and its target is left as it was; a folder at that name is refused.
    #[cfg(unix)]
    #[test]
    fn a_link_at_the_record_name_is_replaced_and_a_folder_is_refused() {
        let home = Place::new();
        put(&home, Some(win(50.0, NOW + 60)), None);
        let path = home.file();
        let target = home.tmp.path().join("target");
        std::fs::write(&target, "kept").unwrap();
        std::fs::remove_file(&path).unwrap();
        std::os::unix::fs::symlink(&target, &path).unwrap();
        assert_eq!(
            home.record(windows(Some(win(95.0, NOW + 60)), None)),
            Ok(true)
        );
        assert!(std::fs::symlink_metadata(&path).unwrap().is_file());
        assert_eq!(std::fs::read_to_string(&target).unwrap(), "kept");
        assert_eq!(home.evaluate(NOW)["state"], "near");
        std::fs::remove_file(&path).unwrap();
        std::fs::create_dir(&path).unwrap();
        assert!(home
            .record(windows(Some(win(96.0, NOW + 60)), None))
            .is_err());
    }
}
