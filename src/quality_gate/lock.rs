//! The gate-run lock: `<project>/.moai/state/gate-run.lock`, one `flock` per run (the Go build's
//! `gate_lock.go` and `gate_lock_unix.go`). The kernel drops the lock when the process ends, so
//! a crashed run never blocks the next one; the owner record is only for the "held by pid"
//! message and is the same JSON the Go build writes.

use super::fmtutil::{go_duration, utc_rfc3339};
use std::fs::{File, OpenOptions};
use std::io::Write;
use std::os::unix::fs::OpenOptionsExt;
use std::os::unix::io::AsRawFd;
use std::path::{Path, PathBuf};
use std::time::{Duration, Instant, SystemTime};

#[derive(Debug)]
pub enum LockError {
    Held,
    Other(String),
}

pub struct GateRunLock {
    _file: File,
}

pub fn lock_path(project_dir: &Path) -> PathBuf {
    project_dir.join(".moai/state/gate-run.lock")
}

/// Takes the lock without waiting: `Held` when another open file description holds it.
pub fn acquire(project_dir: &Path) -> Result<GateRunLock, LockError> {
    let path = lock_path(project_dir);
    if let Some(dir) = path.parent() {
        std::fs::create_dir_all(dir).map_err(|e| {
            LockError::Other(format!("acquire gate-run lock: creating state dir: {e}"))
        })?;
    }
    let mut file = OpenOptions::new()
        .read(true)
        .write(true)
        .create(true)
        .truncate(false)
        .mode(0o644)
        // O_NOFOLLOW: a lock file that is a link is refused, so the lock never opens (and the
        // owner record never overwrites) a file the link points at (the Go build's NoFollow).
        .custom_flags(libc::O_CLOEXEC | libc::O_NOFOLLOW)
        .open(&path)
        .map_err(|e| {
            if e.raw_os_error() == Some(libc::ELOOP) {
                LockError::Other(format!(
                    "open gate-run lock {}: refusing a symbolic link",
                    path.display()
                ))
            } else {
                LockError::Other(format!("open gate-run lock {}: {e}", path.display()))
            }
        })?;
    if unsafe { libc::flock(file.as_raw_fd(), libc::LOCK_EX | libc::LOCK_NB) } != 0 {
        return Err(LockError::Held);
    }
    let record = format!(
        "{{\n  \"pid\": {},\n  \"created_at\": \"{}\"\n}}\n",
        std::process::id(),
        utc_rfc3339(SystemTime::now())
    );
    file.set_len(0)
        .and_then(|_| file.write_all(record.as_bytes()))
        .map_err(|e| {
            LockError::Other(format!(
                "record gate-run lock owner {}: {e}",
                path.display()
            ))
        })?;
    Ok(GateRunLock { _file: file })
}

/// The pid in the owner record, when the file holds a readable one.
fn holder_pid(project_dir: &Path) -> Option<i64> {
    let text = std::fs::read_to_string(lock_path(project_dir)).ok()?;
    let after = text.split("\"pid\"").nth(1)?;
    let digits: String = after
        .trim_start_matches(|c: char| c == ':' || c.is_whitespace())
        .chars()
        .take_while(char::is_ascii_digit)
        .collect();
    digits.parse::<i64>().ok().filter(|p| *p > 0)
}

pub struct Wait {
    pub lock: Option<GateRunLock>,
    /// A line to print after the wait ("running unserialized ...").
    pub line: String,
}

/// Waits up to `budget` for the lock (`waitForGateLock`): prints the holder once per pid; when
/// the budget runs out the run goes on without the lock and says so.
pub fn wait(project_dir: &Path, budget: Duration, err: &mut dyn Write) -> Wait {
    let deadline = Instant::now() + budget;
    let mut notified: Option<i64> = None;
    loop {
        match acquire(project_dir) {
            Ok(lock) => {
                return Wait {
                    lock: Some(lock),
                    line: String::new(),
                }
            }
            Err(LockError::Other(e)) => {
                return Wait {
                    lock: None,
                    line: format!("gate-run lock: unavailable ({e}) — running unserialized"),
                }
            }
            Err(LockError::Held) => {}
        }
        if let Some(pid) = holder_pid(project_dir) {
            if notified != Some(pid) {
                notified = Some(pid);
                let _ = writeln!(
                    err,
                    "gate-run lock: held by pid {pid} — waiting (budget {})",
                    go_duration(budget)
                );
            }
        }
        let left = deadline.saturating_duration_since(Instant::now());
        if left.is_zero() {
            return Wait {
                lock: None,
                line: format!(
                    "gate-run lock: wait budget {} expired — running unserialized",
                    go_duration(budget)
                ),
            };
        }
        // On unix a dead holder's flock is already gone, so the next acquire succeeds (the Go
        // build's stale-lock clearing is a no-op here too); the sleep keeps the retry calm.
        std::thread::sleep(left.min(Duration::from_millis(100)));
    }
}
