//! `build-number` (SPEC-20261125-install-build-number): hands out the next install build number
//! and records it, so a person never picks a `MOAI_LOCAL_BUILD` by hand.
//!
//! The record is `<home>/.moai/projects/<project id>/build-number.log`, one line per number
//! handed out: `<number>\t<UTC stamp>\t<session id or ->`. It is only ever appended to, under an
//! exclusive file lock, and a number is printed only after its line is on disk, so two runs never
//! see the same number and a number is never given twice (HeadVer: the build counter only goes up
//! and is never reused, also when the build that asked for the number failed).
//!
//! The next number is one above the larger of two things: the last recorded number and the
//! highest build number of an installed version folder (`<home>/.moai/versions/<head>.<yearweek>.
//! <n>[-dev]`) of the project's head. The second makes the first run continue from what is
//! installed. A record that is not whole numbers, strictly increasing and ending in a line break
//! is refused and left as it is. The folder of the record is created by project registration and
//! never by this module; a link in its path is refused.

use crate::headver;
use serde_json::{json, Value};
use std::fs;
use std::io::Write as _;
use std::path::{Path, PathBuf};
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};

type Result<T> = std::result::Result<T, String>;

/// The record's file name inside the project's state folder.
pub const LOG_FILE: &str = "build-number.log";
/// The lock file beside it; the lock lives as long as the process holds the open file.
const LOCK_FILE: &str = "build-number.lock";
/// How long a run waits for another run's lock before it refuses.
const LOCK_WAIT: Duration = Duration::from_secs(10);

/// One line of the record.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Entry {
    pub build: u32,
    pub at: String,
    pub session: String,
}

/// The whole record text as entries. Every line must be `<number>\t<UTC stamp>\t<session>`, the
/// numbers strictly increasing from 1, and the text must end with a line break (a torn last line
/// is refused). An empty text is an empty record.
pub fn parse_log(text: &str) -> std::result::Result<Vec<Entry>, String> {
    if !text.is_empty() && !text.ends_with('\n') {
        return Err("its last line is cut off (no line break at the end)".to_string());
    }
    let mut entries: Vec<Entry> = Vec::new();
    for (i, line) in text.lines().enumerate() {
        let n = i + 1;
        let mut parts = line.split('\t');
        let (Some(number), Some(at), Some(session), None) =
            (parts.next(), parts.next(), parts.next(), parts.next())
        else {
            return Err(format!("line {n} is not `<number> <UTC time> <session>`"));
        };
        let build = crate::local_build::number_of("build number", Some(number))
            .ok()
            .flatten()
            .ok_or_else(|| format!("line {n} has no build number from 1 in plain digits"))?;
        if !crate::state_root::utc_stamp(at) {
            return Err(format!("line {n} has no UTC time like 20261125T093000Z"));
        }
        if session.is_empty() || session.chars().any(char::is_whitespace) {
            return Err(format!("line {n} has no session id (or `-`)"));
        }
        if entries.last().is_some_and(|last| build <= last.build) {
            return Err(format!("line {n} does not go above the number before it"));
        }
        entries.push(Entry {
            build,
            at: at.to_string(),
            session: session.to_string(),
        });
    }
    Ok(entries)
}

/// The highest build number among folder names of the form `<head>.<yearweek>.<n>` with an
/// optional `-dev`, for this `head` only; 0 when none. Anything else in the list is ignored.
pub fn installed_floor<'a>(head: u32, names: impl IntoIterator<Item = &'a str>) -> u32 {
    names
        .into_iter()
        .filter_map(|name| {
            let name = name.strip_suffix(headver::DEV_SUFFIX).unwrap_or(name);
            let mut parts = name.split('.');
            let (Some(h), Some(yearweek), Some(build), None) =
                (parts.next(), parts.next(), parts.next(), parts.next())
            else {
                return None;
            };
            let plain = |s: &str| {
                !s.is_empty()
                    && s.bytes().all(|b| b.is_ascii_digit())
                    && (s == "0" || !s.starts_with('0'))
            };
            if !(plain(h)
                && yearweek.len() == 4
                && yearweek.bytes().all(|b| b.is_ascii_digit())
                && plain(build))
            {
                return None;
            }
            (h.parse::<u32>().ok()? == head).then(|| build.parse::<u32>().ok())?
        })
        .max()
        .unwrap_or(0)
}

/// The number after `last` and `floor`, or why there is none.
pub fn after(last: u32, floor: u32) -> Result<u32> {
    last.max(floor).checked_add(1).ok_or_else(|| {
        format!(
            "the build numbers are used up: {} is the largest a build can have",
            crate::local_build::MAX
        )
    })
}

/// `YYYYMMDDTHHMMSSZ` for a time in seconds since 1970-01-01 UTC.
fn stamp(epoch_secs: i64) -> String {
    let days = epoch_secs.div_euclid(86_400);
    let secs = epoch_secs.rem_euclid(86_400);
    let (y, m, d) = headver::civil_from_days(days);
    format!(
        "{y:04}{m:02}{d:02}T{:02}{:02}{:02}Z",
        secs / 3600,
        secs % 3600 / 60,
        secs % 60
    )
}

fn now_secs() -> i64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map_or(0, |d| d.as_secs() as i64)
}

/// Where everything is for one project: its head, its state folder and the record's path.
struct Place {
    head: u32,
    dir: PathBuf,
    log: PathBuf,
    versions: PathBuf,
}

fn place(project: &Path, home: &Path) -> Result<Place> {
    let id = crate::ledger::command::project_id_of(project)?;
    crate::ledger::folders::refuse_links(home, &id).map_err(|e| e.to_string())?;
    let dir = crate::ledger::location::project_dir(home, &id).map_err(|e| e.to_string())?;
    match fs::symlink_metadata(&dir) {
        Ok(m) if m.is_dir() => {}
        _ => {
            return Err(format!(
                "build-number refused: the project's state folder {} does not exist; register \
                 the project first (ledger-register-plan); nothing was written",
                dir.display()
            ))
        }
    }
    let head_text = fs::read_to_string(project.join(headver::HEAD_FILE)).map_err(|_| {
        format!(
            "build-number refused: {} is missing in {}; the project does not use HeadVer",
            headver::HEAD_FILE,
            project.display()
        )
    })?;
    let head = headver::parse_head(&head_text)?;
    Ok(Place {
        head,
        log: dir.join(LOG_FILE),
        dir,
        versions: crate::home_workdir::root(home).join("versions"),
    })
}

fn read_log(path: &Path) -> Result<Vec<Entry>> {
    let text = match fs::read_to_string(path) {
        Ok(t) => t,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(Vec::new()),
        Err(e) => return Err(format!("{LOG_FILE} could not be read: {e}")),
    };
    parse_log(&text).map_err(|why| {
        format!("{LOG_FILE} is refused: {why}; nothing was recorded (the file is left as it is)")
    })
}

fn installed(place: &Place) -> Result<u32> {
    let names = match fs::read_dir(&place.versions) {
        Ok(rd) => rd
            .map(|e| e.map(|e| e.file_name().to_string_lossy().into_owned()))
            .collect::<std::io::Result<Vec<String>>>()
            .map_err(|e| format!("the installed version folders could not be listed: {e}"))?,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Vec::new(),
        Err(e) => {
            return Err(format!(
                "the installed version folders could not be listed: {e}"
            ))
        }
    };
    Ok(installed_floor(
        place.head,
        names.iter().map(String::as_str),
    ))
}

/// Takes the exclusive lock of the record, waiting at most [`LOCK_WAIT`]. The lock is released
/// when the returned file is dropped, also when the process ends in any way.
fn lock(dir: &Path) -> Result<fs::File> {
    let file = fs::OpenOptions::new()
        .create(true)
        .truncate(false)
        .write(true)
        .open(dir.join(LOCK_FILE))
        .map_err(|e| format!("the build-number lock could not be opened: {e}"))?;
    let deadline = Instant::now() + LOCK_WAIT;
    loop {
        match file.try_lock() {
            Ok(()) => return Ok(file),
            Err(fs::TryLockError::WouldBlock) if Instant::now() < deadline => {
                std::thread::sleep(Duration::from_millis(20));
            }
            Err(fs::TryLockError::WouldBlock) => {
                return Err(format!(
                    "another build-number run held the lock for {} seconds; nothing was recorded",
                    LOCK_WAIT.as_secs()
                ))
            }
            Err(fs::TryLockError::Error(e)) => {
                return Err(format!("the build-number lock could not be taken: {e}"))
            }
        }
    }
}

/// `build-number next`: records the next install build number and returns it.
pub fn next(project: &Path, home: &Path, session: Option<&str>) -> Result<Value> {
    next_at(project, home, session, now_secs())
}

/// [`next`] at a given time (seconds since 1970-01-01 UTC), for tests.
pub fn next_at(
    project: &Path,
    home: &Path,
    session: Option<&str>,
    epoch_secs: i64,
) -> Result<Value> {
    let place = place(project, home)?;
    let _lock = lock(&place.dir)?;
    // The record is read under the lock, so the number below is above every line another run
    // wrote before this one got the lock.
    let entries = read_log(&place.log)?;
    let last = entries.last().map_or(0, |e| e.build);
    let floor = installed(&place)?;
    let build = after(last, floor)?;
    let session = session
        .filter(|s| !s.is_empty() && !s.chars().any(char::is_whitespace))
        .unwrap_or("-");
    let line = format!("{build}\t{}\t{session}\n", stamp(epoch_secs));
    append(&place.log, &line)?;
    let (y, m, d) = headver::civil_from_days(epoch_secs.div_euclid(86_400));
    Ok(json!({
        "build": build,
        "env": format!("{}={build}", crate::local_build::ENV),
        "version": headver::dev_version(place.head, headver::yearweek(y, m, d), build),
        "recorded": true,
        "previous": last,
        "installed_floor": floor,
    }))
}

/// Appends one line and forces it to disk before the caller prints the number.
fn append(path: &Path, line: &str) -> Result<()> {
    let mut options = fs::OpenOptions::new();
    options.create(true).append(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        options.mode(0o600);
    }
    let mut file = options
        .open(path)
        .map_err(|e| format!("{LOG_FILE} could not be opened: {e}; nothing was recorded"))?;
    file.write_all(line.as_bytes())
        .and_then(|()| file.sync_all())
        .map_err(|e| format!("{LOG_FILE} could not be written: {e}; the number was not given out"))
}

/// `build-number show`: the last recorded number and the one the next run would give. Writes
/// nothing, and takes no lock.
pub fn show(project: &Path, home: &Path) -> Result<Value> {
    let place = place(project, home)?;
    let entries = read_log(&place.log)?;
    let last = entries.last();
    let floor = installed(&place)?;
    let next = after(last.map_or(0, |e| e.build), floor)?;
    Ok(json!({
        "last": last.map(|e| json!({"build": e.build, "at": e.at, "session": e.session})),
        "next": next,
        "installed_floor": floor,
        "recorded": entries.len(),
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn a_record_parses_into_entries_and_an_empty_one_is_empty() {
        assert_eq!(parse_log("").unwrap(), []);
        let text = "1\t20261125T000000Z\t-\n7\t20261125T000100Z\tabc-123\n";
        let entries = parse_log(text).unwrap();
        assert_eq!(entries.len(), 2);
        assert_eq!(entries[1].build, 7);
        assert_eq!(entries[1].session, "abc-123");
    }

    #[test]
    fn a_broken_record_is_refused_with_its_line() {
        for bad in [
            "1\t20261125T000000Z\t-",
            "x\t20261125T000000Z\t-\n",
            "1\t20261125T000000Z\n",
            "1\t20261125T000000Z\t-\textra\n",
            "1\tnot-a-time\t-\n",
            "2\t20261125T000000Z\t-\n2\t20261125T000001Z\t-\n",
        ] {
            assert!(parse_log(bad).is_err(), "{bad:?}");
        }
        assert!(parse_log("1\t20261125T000000Z\t-\nx\n")
            .unwrap_err()
            .contains("line 2"));
    }

    #[test]
    fn the_floor_is_the_highest_build_of_the_same_head() {
        let names = [
            "4.2648.36-dev",
            "4.2640.7",
            "5.2648.99-dev",
            "4.2648.40-rc",
            "4.2648.040",
            "4.648.90",
            "latest",
            "4.2649.0",
        ];
        assert_eq!(installed_floor(4, names), 36);
        assert_eq!(installed_floor(5, names), 99);
        assert_eq!(installed_floor(6, names), 0);
        assert_eq!(installed_floor(4, []), 0);
    }

    #[test]
    fn the_next_number_is_above_both_and_stops_at_the_largest() {
        assert_eq!(after(0, 0).unwrap(), 1);
        assert_eq!(after(5, 36).unwrap(), 37);
        assert_eq!(after(40, 36).unwrap(), 41);
        assert!(after(u32::MAX, 0).is_err());
        assert_eq!(after(u32::MAX - 1, 0).unwrap(), u32::MAX);
    }

    #[test]
    fn a_time_is_written_as_a_utc_stamp() {
        assert_eq!(stamp(0), "19700101T000000Z");
        assert_eq!(stamp(1_795_606_084), "20261125T112804Z");
        assert!(crate::state_root::utc_stamp(&stamp(now_secs())));
    }
}
