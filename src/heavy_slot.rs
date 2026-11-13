//! Issue #2677 (AC-1, AC-2): a machine-wide limit on concurrent heavy runs — the product port
//! of the repository tool `tools/heavy_check_slot.py` (#2491), with the same folder, file
//! names and lock, so the two take turns with each other.
//!
//! `heavy-slot-acquire --max N --name LANE [--dir DIR] -- COMMAND…` waits for one of N slots,
//! runs COMMAND and returns COMMAND's exit code (128 + the signal number when it ends on a
//! signal). A slot is a file `slot-<i>` with an exclusive `flock` (std `File::try_lock`); the
//! open file is handed to COMMAND, so the slot stays held while either the wrapper or COMMAND
//! is alive and the operating system frees it when both are gone, also after SIGKILL — no
//! stale slot is left to clean. A waiter holds `waiting/<ns>-<pid>`, and its place is its rank
//! among live waiters by name. While waiting, and when COMMAND fails, the load averages and the
//! free disk space of the working folder are printed (AC-2); where to stop is not decided
//! here (load and disk thresholds are a HOLD), only shown.
//!
//! A slot or queue path that is a symlink, FIFO, device, folder or hard link is refused before
//! anything is written: acquire exits 125 without running COMMAND (`status` lists it under
//! `refused`). The default folder is `${XDG_CACHE_HOME:-~/.cache}/moai-cli/heavy-slots`.
//! Unix only; elsewhere both operations refuse.

use serde_json::Value;
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

/// The wrapper's own refusal, apart from COMMAND's codes (like env(1) and timeout(1)).
pub const REFUSED: u8 = 125;
/// Usage errors (bad `--max`, no command, unsupported platform).
pub const USAGE: u8 = 2;

/// `${XDG_CACHE_HOME:-$HOME/.cache}/moai-cli/heavy-slots`, the Python tool's folder.
pub fn default_dir() -> Result<PathBuf> {
    let base = match std::env::var_os("XDG_CACHE_HOME").filter(|v| !v.is_empty()) {
        Some(x) => PathBuf::from(x),
        None => PathBuf::from(std::env::var_os("HOME").ok_or("HOME is not set")?).join(".cache"),
    };
    Ok(base.join("moai-cli").join("heavy-slots"))
}

/// `load 1.23/0.98/0.75; disk free 41.2 GiB at <folder>` — what is known of the machine now.
/// A value the system does not give is shown as `unknown`.
pub fn machine_line(folder: &Path) -> String {
    let load = sys::load()
        .map(|l| format!("{:.2}/{:.2}/{:.2}", l[0], l[1], l[2]))
        .unwrap_or_else(|| "unknown".into());
    let disk = sys::free_bytes(folder)
        .map(|b| format!("{:.1} GiB", b as f64 / (1u64 << 30) as f64))
        .unwrap_or_else(|| "unknown".into());
    format!("load {load}; disk free {disk} at {}", folder.display())
}

#[cfg(unix)]
pub use sys::{acquire, status};

#[cfg(not(unix))]
pub fn status(_folder: &Path) -> Result<Value> {
    Err("heavy-slot needs flock, which this platform does not provide".into())
}

#[cfg(not(unix))]
pub fn acquire(_folder: &Path, _max: usize, _name: &str, _command: &[String]) -> u8 {
    eprintln!("heavy-slot: flock is unavailable on this platform");
    USAGE
}

#[cfg(not(unix))]
mod sys {
    use std::path::Path;
    pub fn load() -> Option<[f64; 3]> {
        None
    }
    pub fn free_bytes(_folder: &Path) -> Option<u64> {
        None
    }
}

#[cfg(unix)]
mod sys {
    use super::{machine_line, Path, PathBuf, Result, Value, REFUSED, USAGE};
    use serde_json::json;
    use std::fs::{self, File, OpenOptions, TryLockError};
    use std::io::{Seek, SeekFrom, Write};
    use std::os::unix::ffi::OsStrExt;
    use std::os::unix::fs::{MetadataExt, OpenOptionsExt};
    use std::os::unix::io::AsRawFd;
    use std::os::unix::process::{CommandExt, ExitStatusExt};
    use std::time::{Duration, SystemTime, UNIX_EPOCH};

    const POLL: Duration = Duration::from_millis(200);
    const FORWARDED: [libc::c_int; 3] = [libc::SIGINT, libc::SIGTERM, libc::SIGHUP];

    pub fn load() -> Option<[f64; 3]> {
        let mut l = [0f64; 3];
        // SAFETY: getloadavg writes at most `nelem` doubles into the buffer it is given.
        let n = unsafe { libc::getloadavg(l.as_mut_ptr(), 3) };
        (n == 3).then_some(l)
    }

    pub fn free_bytes(folder: &Path) -> Option<u64> {
        let c = std::ffi::CString::new(folder.as_os_str().as_bytes()).ok()?;
        // SAFETY: statvfs fills the zeroed struct for a NUL-terminated path.
        let mut s: libc::statvfs = unsafe { std::mem::zeroed() };
        let rc = unsafe { libc::statvfs(c.as_ptr(), &mut s) };
        #[allow(clippy::unnecessary_cast)]
        (rc == 0).then(|| s.f_bavail as u64 * s.f_frsize as u64)
    }

    /// Why a path cannot be a slot or queue file.
    enum Open {
        NotPlain(PathBuf),
        Io(String),
    }

    /// Opens `path` without following a symlink and only as a regular file with one link;
    /// `O_NONBLOCK` keeps a FIFO from blocking the open itself (cleared after the checks).
    fn open_plain(path: &Path, create: bool, exclusive: bool) -> std::result::Result<File, Open> {
        let mut o = OpenOptions::new();
        o.read(true).write(true).mode(0o600);
        o.custom_flags(libc::O_NOFOLLOW | libc::O_NONBLOCK);
        if exclusive {
            o.create_new(true);
        } else if create {
            o.create(true);
        }
        let file = match o.open(path) {
            Ok(f) => f,
            Err(e) => {
                return Err(match fs::symlink_metadata(path) {
                    Ok(m) if !m.file_type().is_file() => Open::NotPlain(path.into()),
                    _ => Open::Io(format!("{}: {e}", path.display())),
                })
            }
        };
        let m = file
            .metadata()
            .map_err(|e| Open::Io(format!("{}: {e}", path.display())))?;
        if !m.file_type().is_file() || m.nlink() != 1 {
            return Err(Open::NotPlain(path.into()));
        }
        let fd = file.as_raw_fd();
        // SAFETY: plain fcntl flag calls on a descriptor this function owns.
        unsafe {
            let fl = libc::fcntl(fd, libc::F_GETFL);
            if fl >= 0 {
                libc::fcntl(fd, libc::F_SETFL, fl & !libc::O_NONBLOCK);
            }
        }
        Ok(file)
    }

    /// The file locked exclusively, or `None` when another process holds it.
    fn try_exclusive(
        path: &Path,
        exclusive_create: bool,
    ) -> std::result::Result<Option<File>, Open> {
        let f = open_plain(path, true, exclusive_create)?;
        match f.try_lock() {
            Ok(()) => Ok(Some(f)),
            Err(TryLockError::WouldBlock) => Ok(None),
            Err(TryLockError::Error(e)) => Err(Open::Io(format!("{}: {e}", path.display()))),
        }
    }

    /// Whether another process holds `path` (a shared try-lock that fails means held).
    fn is_held(path: &Path) -> std::result::Result<bool, Open> {
        let f = open_plain(path, false, false)?;
        match f.try_lock_shared() {
            Ok(()) => Ok(false),
            Err(TryLockError::WouldBlock) => Ok(true),
            Err(TryLockError::Error(e)) => Err(Open::Io(format!("{}: {e}", path.display()))),
        }
    }

    fn take_slot(folder: &Path, max: usize) -> std::result::Result<Option<(usize, File)>, Open> {
        for i in 0..max {
            if let Some(f) = try_exclusive(&folder.join(format!("slot-{i}")), false)? {
                return Ok(Some((i, f)));
            }
        }
        Ok(None)
    }

    /// 1-based place of `me` among live waiters (oldest first by name), and their count.
    fn queue_place(waiting: &Path, me: &str) -> (usize, usize) {
        let mut names: Vec<String> = fs::read_dir(waiting)
            .map(|r| {
                r.filter_map(|e| e.ok())
                    .map(|e| e.file_name().to_string_lossy().into_owned())
                    .collect()
            })
            .unwrap_or_default();
        names.sort();
        let live: Vec<&String> = names
            .iter()
            .filter(|n| n.as_str() == me || matches!(is_held(&waiting.join(n)), Ok(true)))
            .collect();
        let place = live
            .iter()
            .position(|n| n.as_str() == me)
            .map_or(1, |p| p + 1);
        (place, live.len().max(1))
    }

    fn now_ns() -> u128 {
        SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .map_or(0, |d| d.as_nanos())
    }

    fn refused(p: &Path) -> u8 {
        eprintln!(
            "heavy-slot: refused {}: not a plain file (symlink, FIFO, device, folder or hard link); nothing was written and COMMAND did not run — remove it or pass another --dir",
            p.display()
        );
        REFUSED
    }

    fn io(e: &str) -> u8 {
        eprintln!("heavy-slot: {e}; COMMAND did not run");
        REFUSED
    }

    /// Waits for a slot, runs `command` holding it, returns its exit code.
    pub fn acquire(folder: &Path, max: usize, name: &str, command: &[String]) -> u8 {
        if max < 1 || command.is_empty() {
            eprintln!("heavy-slot: acquire needs --max >= 1 and a command after --");
            return USAGE;
        }
        let waiting = folder.join("waiting");
        if let Err(e) = fs::create_dir_all(&waiting) {
            return io(&format!("creating {} failed ({e})", waiting.display()));
        }
        let cwd = std::env::current_dir().unwrap_or_else(|_| folder.to_path_buf());
        let (slot, file) = match take_slot(folder, max) {
            Err(Open::NotPlain(p)) => return refused(&p),
            Err(Open::Io(e)) => return io(&e),
            Ok(Some(t)) => t,
            Ok(None) => {
                let me = format!("{:020}-{}", now_ns(), std::process::id());
                let wait_path = waiting.join(&me);
                let wait_file = match try_exclusive(&wait_path, true) {
                    Err(Open::NotPlain(p)) => return refused(&p),
                    Err(Open::Io(e)) => return io(&e),
                    Ok(f) => f,
                };
                let mut last = None;
                let taken = loop {
                    let place = queue_place(&waiting, &me);
                    if last != Some(place) {
                        println!(
                            "heavy-slot: waiting {} of {} for {max} slot(s); {}",
                            place.0,
                            place.1,
                            machine_line(&cwd)
                        );
                        last = Some(place);
                    }
                    std::thread::sleep(POLL);
                    match take_slot(folder, max) {
                        Ok(Some(t)) => break Ok(t),
                        Ok(None) => continue,
                        Err(e) => break Err(e),
                    }
                };
                drop(wait_file);
                let _ = fs::remove_file(&wait_path);
                match taken {
                    Ok(t) => t,
                    Err(Open::NotPlain(p)) => return refused(&p),
                    Err(Open::Io(e)) => return io(&e),
                }
            }
        };
        run_holding(slot, file, name, command, &cwd)
    }

    fn run_holding(slot: usize, mut file: File, name: &str, command: &[String], cwd: &Path) -> u8 {
        // Forwarded signals are blocked before the command starts, so none lands in a gap
        // where the wrapper dies and leaves the command running; a waiter thread receives them
        // with sigwait and passes them on. The command unblocks them before it execs.
        // SAFETY: signal-set calls on a stack-local set.
        let set = unsafe {
            let mut s: libc::sigset_t = std::mem::zeroed();
            libc::sigemptyset(&mut s);
            for sig in FORWARDED {
                libc::sigaddset(&mut s, sig);
            }
            libc::pthread_sigmask(libc::SIG_BLOCK, &s, std::ptr::null_mut());
            s
        };
        let record = json!({
            "name": name,
            "command": command,
            "pid": std::process::id(),
            "started_unix": SystemTime::now().duration_since(UNIX_EPOCH).map_or(0, |d| d.as_secs()),
            "slot": slot,
        });
        let _ = file.set_len(0);
        let _ = file.seek(SeekFrom::Start(0));
        let _ = writeln!(file, "{record}");
        println!("heavy-slot: slot {slot} taken by {name}");
        // The command inherits the locked file: clear close-on-exec on it.
        let fd = file.as_raw_fd();
        // SAFETY: fcntl on a descriptor this function owns.
        unsafe {
            let fl = libc::fcntl(fd, libc::F_GETFD);
            if fl >= 0 {
                libc::fcntl(fd, libc::F_SETFD, fl & !libc::FD_CLOEXEC);
            }
        }
        let mut cmd = std::process::Command::new(&command[0]);
        cmd.args(&command[1..]);
        let unblock = set;
        // SAFETY: pthread_sigmask is async-signal-safe; nothing else runs between fork and exec.
        unsafe {
            cmd.pre_exec(move || {
                libc::pthread_sigmask(libc::SIG_UNBLOCK, &unblock, std::ptr::null_mut());
                Ok(())
            });
        }
        let mut child = match cmd.spawn() {
            Ok(c) => c,
            Err(e) => {
                eprintln!("heavy-slot: {name} could not start {} ({e})", command[0]);
                return 127;
            }
        };
        let pid = child.id() as libc::pid_t;
        std::thread::spawn(move || loop {
            let mut sig: libc::c_int = 0;
            // SAFETY: sigwait on the set blocked above; kill to the child's pid.
            if unsafe { libc::sigwait(&set, &mut sig) } == 0 {
                unsafe { libc::kill(pid, sig) };
            }
        });
        let code = match child.wait() {
            Ok(s) => s
                .code()
                .map(|c| c as u8)
                .or_else(|| s.signal().map(|n| (128 + n) as u8))
                .unwrap_or(1),
            Err(_) => 1,
        };
        drop(file);
        if code != 0 {
            eprintln!("heavy-slot: {name} exited {code}; {}", machine_line(cwd));
        }
        code
    }

    /// The slots held now (their records) and slot paths that are not plain files. Reads only.
    pub fn status(folder: &Path) -> Result<Value> {
        let (mut held, mut refused) = (Vec::new(), Vec::new());
        let mut names: Vec<String> = match fs::read_dir(folder) {
            Err(_) => vec![],
            Ok(r) => r
                .filter_map(|e| e.ok())
                .map(|e| e.file_name().to_string_lossy().into_owned())
                .filter(|n| n.starts_with("slot-"))
                .collect(),
        };
        names.sort();
        for n in names {
            let p = folder.join(&n);
            match is_held(&p) {
                Ok(true) => held.push(
                    fs::read(&p)
                        .ok()
                        .and_then(|b| serde_json::from_slice::<Value>(&b).ok())
                        .unwrap_or_else(|| json!({})),
                ),
                Ok(false) => {}
                Err(Open::NotPlain(_)) => refused.push(n),
                Err(Open::Io(e)) => return Err(e),
            }
        }
        Ok(json!({
            "folder": folder.display().to_string(),
            "held": held,
            "refused": refused,
            "machine": machine_line(folder),
            "writes": false,
        }))
    }
}
