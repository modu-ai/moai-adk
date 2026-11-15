//! When the card ledger is backed up and what the backup is called (#3070 K2, requirement 8).
//!
//! The rules, as pure functions of the clock and of names:
//! - a project has **one** backup file. A new backup replaces it; nothing is ever added next to it;
//! - a write may make a new backup at most once per UTC calendar day: the first write of a day
//!   that finds the backup older than that day (or no backup) makes one, every later write of the
//!   same day does not;
//! - the new backup is written under a temporary name beside the old one and moved over it, so
//!   the old backup stays whole until the new one is complete.
//!
//! Taking the backup itself (SQLite's own backup method, never a copy of the open file) belongs
//! to the ledger code of #3068 and #3070; this module opens nothing and reads no clock.
//!
//! Own choices, easy to change: the day is the UTC calendar day, not a rolling 24 hours, so three
//! simulated days of writes give exactly three backups in turn and one file at the end (AC7); the
//! backup sits in the ledger's folder under the ledger's name plus `.bak`.

use crate::forge::policy::format_utc;
use std::ffi::OsString;
use std::path::{Path, PathBuf};

const DAY_SECS: i64 = 86_400;
const BACKUP_SUFFIX: &str = ".bak";
const TEMP_SUFFIX: &str = ".bak.tmp";

/// What a write should do about the backup.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Plan {
    /// Make a backup now; `replaces` says one already exists and will be replaced.
    Create { replaces: bool },
    /// The backup is from today (or from a later time than the clock says); leave it.
    Keep,
}

/// What is known about the existing backup file.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct BackupInfo {
    /// Seconds since the Unix epoch when it was made.
    pub made_at: i64,
    pub bytes: u64,
}

fn day(secs: i64) -> i64 {
    secs.div_euclid(DAY_SECS)
}

/// Decides, from the existing backup and the clock, whether this write makes a new backup.
/// A backup dated in the future (a clock that went back) counts as today's and is kept.
pub fn plan(existing: Option<BackupInfo>, now: i64) -> Plan {
    match existing {
        None => Plan::Create { replaces: false },
        Some(info) if day(now) > day(info.made_at) => Plan::Create { replaces: true },
        Some(_) => Plan::Keep,
    }
}

fn with_suffix(ledger: &Path, suffix: &str) -> PathBuf {
    let mut name: OsString = ledger.as_os_str().to_owned();
    name.push(suffix);
    PathBuf::from(name)
}

/// The one backup file of a ledger: the ledger's path plus `.bak`.
pub fn backup_path(ledger: &Path) -> PathBuf {
    with_suffix(ledger, BACKUP_SUFFIX)
}

/// Where a new backup is written before it replaces the old one.
pub fn temp_path(ledger: &Path) -> PathBuf {
    with_suffix(ledger, TEMP_SUFFIX)
}

/// The line `card backup` shows: when the backup was made and how big it is.
pub fn describe(existing: Option<BackupInfo>) -> String {
    match existing {
        None => "no backup yet".to_string(),
        Some(info) => format!(
            "backup made {} · {} bytes",
            format_utc(info.made_at),
            info.bytes
        ),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    // 2026-11-15T00:00:00Z.
    const DAY0: i64 = 1_794_700_800;

    fn info(made_at: i64) -> Option<BackupInfo> {
        Some(BackupInfo { made_at, bytes: 1 })
    }

    #[test]
    fn the_day_constants_are_what_the_tests_assume() {
        assert_eq!(format_utc(DAY0), "2026-11-15T00:00:00Z");
    }

    #[test]
    fn no_backup_means_make_one_without_replacing() {
        assert_eq!(plan(None, DAY0), Plan::Create { replaces: false });
        assert_eq!(plan(None, 0), Plan::Create { replaces: false });
    }

    #[test]
    fn a_backup_from_earlier_today_is_kept_up_to_the_last_second_of_the_day() {
        let made = DAY0 + 3_600;
        assert_eq!(plan(info(made), made), Plan::Keep);
        assert_eq!(plan(info(made), DAY0 + 86_399), Plan::Keep);
        assert_eq!(plan(info(DAY0), DAY0 + 86_399), Plan::Keep);
    }

    #[test]
    fn the_first_second_of_the_next_day_makes_a_new_backup_that_replaces_the_old() {
        assert_eq!(
            plan(info(DAY0 + 86_399), DAY0 + 86_400),
            Plan::Create { replaces: true }
        );
        assert_eq!(
            plan(info(DAY0), DAY0 + 86_400),
            Plan::Create { replaces: true }
        );
        assert_eq!(
            plan(info(DAY0), DAY0 + 10 * 86_400),
            Plan::Create { replaces: true }
        );
    }

    #[test]
    fn less_than_twenty_four_hours_across_midnight_still_makes_a_new_backup() {
        // The day is the calendar day: 23:59:59 and 00:00:01 are two seconds apart and two days.
        assert_eq!(
            plan(info(DAY0 + 86_399), DAY0 + 86_401),
            Plan::Create { replaces: true }
        );
    }

    #[test]
    fn a_backup_from_the_future_is_kept() {
        assert_eq!(plan(info(DAY0 + 5 * 86_400), DAY0), Plan::Keep);
        assert_eq!(plan(info(DAY0 + 86_400), DAY0 + 86_399), Plan::Keep);
    }

    #[test]
    fn dates_before_the_epoch_use_the_same_calendar_days() {
        // -1 is 1969-12-31T23:59:59Z; 0 is the next day.
        assert_eq!(plan(info(-1), 0), Plan::Create { replaces: true });
        assert_eq!(plan(info(-86_400), -1), Plan::Keep);
    }

    #[test]
    fn three_days_of_writes_make_three_backups_in_turn_and_never_more_than_one_file() {
        // Ten writes a day for three days; the "file" is a single Option replaced in place.
        let mut file: Option<BackupInfo> = None;
        let mut made = 0;
        for d in 0..3 {
            for k in 0..10 {
                let now = DAY0 + d * 86_400 + k * 8_000;
                if let Plan::Create { replaces } = plan(file, now) {
                    assert_eq!(replaces, file.is_some());
                    file = info(now);
                    made += 1;
                }
            }
        }
        assert_eq!(made, 3);
        assert!(file.is_some());
    }

    #[test]
    fn the_backup_and_temporary_names_sit_beside_the_ledger() {
        let ledger = Path::new("/h/.moai/projects/p1/cards.sqlite");
        assert_eq!(
            backup_path(ledger),
            PathBuf::from("/h/.moai/projects/p1/cards.sqlite.bak")
        );
        assert_eq!(
            temp_path(ledger),
            PathBuf::from("/h/.moai/projects/p1/cards.sqlite.bak.tmp")
        );
        assert_ne!(backup_path(ledger), temp_path(ledger));
        assert_ne!(backup_path(ledger), ledger);
        // A name with spaces or non-ASCII letters keeps its parent and its own name.
        let odd = Path::new("a b/원장.db");
        assert_eq!(backup_path(odd), PathBuf::from("a b/원장.db.bak"));
    }

    #[test]
    fn the_listing_line_shows_the_time_and_size_or_says_there_is_none() {
        assert_eq!(describe(None), "no backup yet");
        assert_eq!(
            describe(Some(BackupInfo {
                made_at: DAY0 + 3_661,
                bytes: 12_288
            })),
            "backup made 2026-11-15T01:01:01Z · 12288 bytes"
        );
        assert_eq!(
            describe(Some(BackupInfo {
                made_at: 0,
                bytes: 0
            })),
            "backup made 1970-01-01T00:00:00Z · 0 bytes"
        );
    }
}
