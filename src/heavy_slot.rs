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
//! free disk space of the working folder are printed (AC-2). A machine below `--min-free-gb`
//! free or above `--max-swap-pct` swap use starts nothing and exits 75 with the measured
//! values (leader's numbers 50 GB and 80 %, Issue #2677 comment 6850581277), judged before
//! queueing and again after a wait. A waiter whose parent process ended leaves the queue and
//! exits 125 without running COMMAND, so no orphan waits on for hours.
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
/// The machine is too short of disk or too deep in swap to start a heavy run now: the
/// wrapper's own "try again later" code (sysexits `EX_TEMPFAIL`), apart from COMMAND's codes.
pub const BUSY: u8 = 75;

/// Where a new heavy run is not started (#2677, item 3 of the leader's measurement record,
/// comment 6850581277: "남은 공간 < 50 GB 또는 swap 사용 > 80% 면 새 빌드·시험을 시작하지 않고
/// 이유를 출력"). `--min-free-gb 0 --max-swap-pct 100` turns both off.
#[derive(Clone, Copy, Debug)]
pub struct Limits {
    /// Free space (decimal GB, 10^9 bytes) the working folder's file system must keep.
    pub min_free_gb: u64,
    /// Highest swap use, in percent of the swap size, at which a run still starts.
    pub max_swap_pct: u64,
}

/// The leader's numbers (comment 6850581277).
pub const DEFAULT_LIMITS: Limits = Limits {
    min_free_gb: 50,
    max_swap_pct: 80,
};

/// The free bytes and swap percent the limits are judged on. Debug builds take
/// `MOAI_CLI_TEST_MACHINE=free_gb=<n>,swap_pct=<n>` instead (or `@<file>` holding that text,
/// read on every call so a test can change the machine during a wait); release builds never
/// read it.
fn measured(folder: &Path) -> (Option<u64>, Option<u64>) {
    #[cfg(debug_assertions)]
    if let Ok(mut s) = std::env::var("MOAI_CLI_TEST_MACHINE") {
        if let Some(file) = s.strip_prefix('@') {
            s = std::fs::read_to_string(file).unwrap_or_default();
        }
        let mut free = None;
        let mut swap = None;
        for part in s.trim().split(',') {
            match part.split_once('=') {
                Some(("free_gb", v)) => free = v.parse::<u64>().ok().map(|g| g * 1_000_000_000),
                Some(("swap_pct", v)) => swap = v.parse().ok(),
                _ => {}
            }
        }
        return (free, swap);
    }
    (sys::free_bytes(folder), sys::swap_pct())
}

/// Why a heavy run must not start now, or `None`. A value the system does not give is not a
/// reason to refuse (it is shown as `unknown` elsewhere).
pub fn over_limits(folder: &Path, limits: Limits) -> Option<String> {
    let (free, swap) = measured(folder);
    let mut why = Vec::new();
    if let Some(b) = free {
        if b < limits.min_free_gb.saturating_mul(1_000_000_000) {
            why.push(format!(
                "disk free {:.1} GB at {} is below {} GB",
                b as f64 / 1e9,
                folder.display(),
                limits.min_free_gb
            ));
        }
    }
    if let Some(p) = swap {
        if p > limits.max_swap_pct {
            why.push(format!("swap use {p}% is above {}%", limits.max_swap_pct));
        }
    }
    (!why.is_empty()).then(|| why.join("; "))
}

/// `${XDG_CACHE_HOME:-$HOME/.cache}/moai-cli/heavy-slots`, the Python tool's folder.
/// #2879 D3: one build folder per lane. When the wrapped command is `cargo` and the lane's
/// `CARGO_TARGET_DIR` is not the folder `moai lane-env --lane <name>` names (unset counts: cargo
/// then builds into the worktree's own `target/`), a one-line warning for stderr. A warning, not
/// a refusal: CI runners and lanes not yet on `lane-env` keep working while the line tells them
/// how to switch. `None` when the name is no lane name or the folder is already the lane's.
pub fn build_folder_note(
    name: &str,
    command: &[String],
    target_dir: Option<&std::ffi::OsStr>,
    base: &Path,
) -> Option<String> {
    let program = command.first().map(Path::new)?.file_name()?;
    if program != "cargo" || crate::lane_env::check_lane(name).is_err() {
        return None;
    }
    let want = base.join(format!("moai-rc-target-{name}"));
    if target_dir.is_some_and(|t| Path::new(t) == want) {
        return None;
    }
    let now = target_dir.map_or_else(
        || "unset (the worktree's own target/)".to_string(),
        |t| Path::new(t).display().to_string(),
    );
    Some(format!(
        "heavy-slot: CARGO_TARGET_DIR is {now}, not this lane's one build folder {} (#2879); run `e=$(moai lane-env --lane {name}) && eval \"$e\"` first",
        want.display()
    ))
}

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
pub fn acquire(
    _folder: &Path,
    _max: usize,
    _name: &str,
    _command: &[String],
    _limits: Limits,
) -> u8 {
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
    pub fn swap_pct() -> Option<u64> {
        None
    }
}

#[cfg(unix)]
mod sys {
    use super::{
        machine_line, over_limits, Limits, Path, PathBuf, Result, Value, BUSY, REFUSED, USAGE,
    };
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

    /// Swap in use, in percent of the swap size (0 when there is no swap). macOS: the
    /// `vm.swapusage` sysctl; Linux: `SwapTotal`/`SwapFree` of `/proc/meminfo`.
    #[cfg(target_os = "macos")]
    pub fn swap_pct() -> Option<u64> {
        // SAFETY: sysctlbyname fills at most `len` bytes of the zeroed struct.
        let mut u: libc::xsw_usage = unsafe { std::mem::zeroed() };
        let mut len = std::mem::size_of::<libc::xsw_usage>();
        let rc = unsafe {
            libc::sysctlbyname(
                c"vm.swapusage".as_ptr(),
                (&mut u as *mut libc::xsw_usage).cast(),
                &mut len,
                std::ptr::null_mut(),
                0,
            )
        };
        if rc != 0 {
            return None;
        }
        Some(if u.xsu_total == 0 {
            0
        } else {
            u.xsu_used.saturating_mul(100) / u.xsu_total
        })
    }

    #[cfg(not(target_os = "macos"))]
    pub fn swap_pct() -> Option<u64> {
        let text = fs::read_to_string("/proc/meminfo").ok()?;
        let field = |name: &str| -> Option<u64> {
            text.lines()
                .find_map(|l| l.strip_prefix(name))
                .and_then(|r| r.trim().trim_end_matches("kB").trim().parse().ok())
        };
        let (total, free) = (field("SwapTotal:")?, field("SwapFree:")?);
        Some(if total == 0 {
            0
        } else {
            total.saturating_sub(free).saturating_mul(100) / total
        })
    }

    fn busy(name: &str, why: &str, cwd: &Path) -> u8 {
        eprintln!(
            "heavy-slot: {name} not started — {why}; {}; COMMAND did not run, try again later",
            machine_line(cwd)
        );
        BUSY
    }

    fn parent() -> libc::pid_t {
        // SAFETY: getppid has no failure mode.
        unsafe { libc::getppid() }
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

    /// Waits for a slot, runs `command` holding it, returns its exit code. A machine over
    /// `limits` starts nothing ([`BUSY`]); a waiter whose parent ended leaves the queue without
    /// running anything ([`REFUSED`]) — nobody is left to take its result (#2677 items 3, 4).
    pub fn acquire(
        folder: &Path,
        max: usize,
        name: &str,
        command: &[String],
        limits: Limits,
    ) -> u8 {
        if max < 1 || command.is_empty() {
            eprintln!("heavy-slot: acquire needs --max >= 1 and a command after --");
            return USAGE;
        }
        let cwd = std::env::current_dir().unwrap_or_else(|_| folder.to_path_buf());
        if let Some(why) = over_limits(&cwd, limits) {
            return busy(name, &why, &cwd);
        }
        let ppid = parent();
        let waiting = folder.join("waiting");
        if let Err(e) = fs::create_dir_all(&waiting) {
            return io(&format!("creating {} failed ({e})", waiting.display()));
        }
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
                    if parent() != ppid {
                        break Ok(None);
                    }
                    match take_slot(folder, max) {
                        Ok(Some(t)) => break Ok(Some(t)),
                        Ok(None) => continue,
                        Err(e) => break Err(e),
                    }
                };
                drop(wait_file);
                let _ = fs::remove_file(&wait_path);
                match taken {
                    Ok(Some(t)) => {
                        // The machine may have changed during the wait: judge it again
                        // before starting; dropping the slot file frees the slot.
                        if let Some(why) = over_limits(&cwd, limits) {
                            return busy(name, &why, &cwd);
                        }
                        t
                    }
                    Ok(None) => {
                        eprintln!(
                            "heavy-slot: {name} left the queue: the process that started it (pid {ppid}) has ended; COMMAND did not run"
                        );
                        return REFUSED;
                    }
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

#[cfg(test)]
mod build_folder_tests {
    use super::build_folder_note;
    use std::ffi::OsStr;
    use std::path::Path;

    fn cmd(v: &[&str]) -> Vec<String> {
        v.iter().map(|s| s.to_string()).collect()
    }

    #[test]
    fn a_cargo_run_off_the_lane_folder_is_named() {
        let base = Path::new("/tmp");
        let cargo = cmd(&["cargo", "test", "--offline"]);
        // The lane's own folder: nothing to say.
        let own = OsStr::new("/tmp/moai-rc-target-agent-7");
        assert_eq!(build_folder_note("agent-7", &cargo, Some(own), base), None);
        // Another folder, a per-card folder, or none at all: one line naming the right one.
        for t in [
            Some("/tmp/moai-rc-target-agent-7-a2846"),
            Some("/tmp/x"),
            None,
        ] {
            let n = build_folder_note("agent-7", &cargo, t.map(OsStr::new), base).unwrap();
            assert!(n.contains("/tmp/moai-rc-target-agent-7 (#2879)"), "{n}");
            assert!(n.contains("moai lane-env --lane agent-7"), "{n}");
        }
        // A full path to cargo counts as cargo too.
        let full = cmd(&["/usr/local/bin/cargo", "build"]);
        assert!(build_folder_note("agent-7", &full, None, base).is_some());
    }

    #[test]
    fn other_commands_and_non_lane_names_are_left_alone() {
        let base = Path::new("/tmp");
        assert_eq!(
            build_folder_note("agent-7", &cmd(&["python3", "-m", "unittest"]), None, base),
            None
        );
        assert_eq!(
            build_folder_note("a b", &cmd(&["cargo", "test"]), None, base),
            None
        );
        assert_eq!(build_folder_note("agent-7", &[], None, base), None);
    }
}
