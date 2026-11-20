//! Issue #2677 (AC-1, AC-2): a machine-wide limit on concurrent heavy runs — the product port
//! of the repository tool `tools/heavy_check_slot.py` (#2491), with the same folder, file
//! names and lock, so the two take turns with each other.
//!
//! `heavy-slot-acquire [--max N] --name LANE [--dir DIR] -- COMMAND…` waits for one of N slots,
//! runs COMMAND and returns COMMAND's exit code (128 + the signal number when it ends on a
//! signal). A slot is a file `slot-<i>` with an exclusive `flock` (std `File::try_lock`); the
//! open file is handed to COMMAND, so the slot stays held while the wrapper, COMMAND or any
//! child or descendant of COMMAND that still has the file open is alive, and the operating
//! system frees it when every one of them is gone, also after SIGKILL — no slot file is left
//! to clean. A descendant that outlives COMMAND (a background process the command started)
//! therefore keeps the slot after the wrapper has ended; `status` then still lists the slot
//! with the wrapper's recorded pid, which is dead. Each `held` entry of `status` says so
//! (`alive`: true, false, or null when the record names no process); for a dead pid it also
//! lists the processes that have the slot file open (`holders`, `pid` and `comm`, from
//! `/proc` or `lsof`; `unavailable` with `holders` null is not "nobody"). That is the list of
//! processes with the file open, not a proof of who owns the lock. Ending such a process is
//! for its owner; a slot file is never edited or deleted by hand. The Python twin has no
//! `alive` or `holders`. A waiter holds `waiting/<ns>-<pid>`, and its place is its rank
//! among live waiters by name. While waiting, and when COMMAND fails, the load averages and the
//! free disk space of the working folder are printed (AC-2). A machine below `--min-free-gb`
//! free or above `--max-swap-pct` swap use starts nothing and exits 75 with the measured
//! values (leader's numbers 50 GB and 80 %, Issue #2677 comment 6850581277), judged before
//! queueing and again after a wait. On macOS (card C-88) memory is judged on the kernel's
//! memory pressure instead of swap, since macOS keeps pages swapped long after the pressure
//! drops (measured 2026-11-17: swap 99 % used at pressure level 2 with 37 % free): a run is
//! refused at `kern.memorystatus_vm_pressure_level` 4 (critical) or with
//! `kern.memorystatus_level` (the "System-wide memory free percentage" `memory_pressure`
//! prints) below `--min-mem-free-pct` (default 10); swap is then shown in the refusal for
//! information only. Linux keeps the swap rule and has no memory-free limit. A run that does start with under 100 GB free
//! ([`WARN_FREE_GB`]) first prints one warning line on stderr (#2930 item 6). A waiter whose parent process ended leaves the queue and
//! exits 125 without running COMMAND, so no orphan waits on for hours.
//!
//! A slot or queue path that is a symlink, FIFO, device, folder or hard link is refused before
//! anything is written: acquire exits 125 without running COMMAND (`status` lists it under
//! `refused`). The default folder is `${XDG_CACHE_HOME:-~/.cache}/moai-cli/heavy-slots`.
//! Unix only; elsewhere both operations refuse.
//!
//! #2930 R-2: the machine's limits live in one file beside the slots, `<folder>/limits.json`
//! (`max`, `min_free_gb`, `max_swap_pct`, `min_mem_free_pct`; defaults 4, 50, 80, 10). A
//! caller's flag wins only when it is stricter. A limits file that is not a plain file of at most 4096 bytes holding only
//! those keys, or that sets `max` 0, is refused: acquire exits 125 without running COMMAND.

use serde_json::Value;
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

/// The wrapper's own refusal, apart from COMMAND's codes (like env(1) and timeout(1)).
pub const REFUSED: u8 = 125;
/// Usage errors (bad `--max`, no command, unsupported platform).
pub const USAGE: u8 = 2;
/// The machine is too short of disk or memory to start a heavy run now: the
/// wrapper's own "try again later" code (sysexits `EX_TEMPFAIL`), apart from COMMAND's codes.
pub const BUSY: u8 = 75;

/// Where a new heavy run is not started (#2677, item 3 of the leader's measurement record,
/// comment 6850581277: "남은 공간 < 50 GB 또는 swap 사용 > 80% 면 새 빌드·시험을 시작하지 않고
/// 이유를 출력"). Only `limits.json` beside the slots can loosen them; a caller's flag only
/// tightens (#2930 R-2, [`effective`]).
#[derive(Clone, Copy, Debug)]
pub struct Limits {
    /// Free space (decimal GB, 10^9 bytes) the working folder's file system must keep.
    pub min_free_gb: u64,
    /// Highest swap use, in percent of the swap size, at which a run still starts. Linux
    /// only: macOS is judged on memory pressure (card C-88).
    pub max_swap_pct: u64,
    /// Lowest memory free percentage (`kern.memorystatus_level`) at which a run still starts.
    /// macOS only (card C-88).
    pub min_mem_free_pct: u64,
}

/// The leader's numbers (comment 6850581277); `min_mem_free_pct` 10 is card C-88's.
pub const DEFAULT_LIMITS: Limits = Limits {
    min_free_gb: 50,
    max_swap_pct: 80,
    min_mem_free_pct: 10,
};

/// `kern.memorystatus_vm_pressure_level` values: 1 normal, 2 warn, 4 critical (the
/// `DISPATCH_MEMORYPRESSURE_*` levels). A run is refused from this level on (card C-88).
pub const CRITICAL_PRESSURE_LEVEL: u64 = 4;

/// The slot count when neither the limits file nor the caller names one (#2930 R-2; the
/// 2026-11-20 cap of four heavy builds on one machine, so the announced cap holds without
/// a limits file too).
pub const DEFAULT_MAX: usize = 4;

/// The machine's limits, beside the slots: `<slot folder>/limits.json` (#2930 R-2, D-a).
pub const LIMITS_FILE: &str = "limits.json";

/// The largest limits file read; a bigger one is refused, not cut.
const LIMITS_FILE_MAX_BYTES: u64 = 4096;

/// `limits.json`: every key optional, any other key refused.
#[derive(Debug, Default, serde::Deserialize)]
#[serde(deny_unknown_fields)]
pub struct LimitsFile {
    pub max: Option<usize>,
    pub min_free_gb: Option<u64>,
    pub max_swap_pct: Option<u64>,
    pub min_mem_free_pct: Option<u64>,
}

/// The limits file in `folder`: `None` when there is none. A path that is not a plain file
/// (a symlink, even to a valid file, or a folder), a file over 4096 bytes, JSON that does not
/// parse, an unknown key or `max` 0 is an error — the caller runs nothing.
pub fn read_limits_file(folder: &Path) -> Result<Option<LimitsFile>> {
    use std::io::Read;
    let path = folder.join(LIMITS_FILE);
    let meta = match std::fs::symlink_metadata(&path) {
        Ok(m) => m,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(None),
        Err(e) => return Err(format!("{}: {e}", path.display())),
    };
    if !meta.file_type().is_file() {
        return Err(format!(
            "{}: not a plain file (symlink, FIFO, device or folder)",
            path.display()
        ));
    }
    let mut o = std::fs::OpenOptions::new();
    o.read(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        o.custom_flags(libc::O_NOFOLLOW | libc::O_NONBLOCK);
    }
    let file = o
        .open(&path)
        .map_err(|e| format!("{}: {e}", path.display()))?;
    let mut bytes = Vec::new();
    file.take(LIMITS_FILE_MAX_BYTES + 1)
        .read_to_end(&mut bytes)
        .map_err(|e| format!("{}: {e}", path.display()))?;
    if bytes.len() as u64 > LIMITS_FILE_MAX_BYTES {
        return Err(format!(
            "{}: larger than {LIMITS_FILE_MAX_BYTES} bytes",
            path.display()
        ));
    }
    // An object only: serde would also read a struct from a JSON array.
    let value: Value =
        serde_json::from_slice(&bytes).map_err(|e| format!("{}: {e}", path.display()))?;
    if !value.is_object() {
        return Err(format!("{}: not a JSON object", path.display()));
    }
    let file: LimitsFile =
        serde_json::from_value(value).map_err(|e| format!("{}: {e}", path.display()))?;
    if file.max == Some(0) {
        return Err(format!("{}: max must be at least 1", path.display()));
    }
    Ok(Some(file))
}

/// The limits a run uses: the stricter of the file (or the defaults) and the caller's flags —
/// a flag only tightens (#2930 D-b), so no caller turns a limit off. Also one note per caller
/// value that a stricter file or default overrode.
pub fn effective(
    file: Option<&LimitsFile>,
    caller_max: Option<usize>,
    caller_free: Option<u64>,
    caller_swap: Option<u64>,
    caller_mem_free: Option<u64>,
) -> (usize, Limits, Vec<String>) {
    let base_max = file.and_then(|f| f.max).unwrap_or(DEFAULT_MAX);
    let base_free = file
        .and_then(|f| f.min_free_gb)
        .unwrap_or(DEFAULT_LIMITS.min_free_gb);
    let base_swap = file
        .and_then(|f| f.max_swap_pct)
        .unwrap_or(DEFAULT_LIMITS.max_swap_pct);
    let base_mem_free = file
        .and_then(|f| f.min_mem_free_pct)
        .unwrap_or(DEFAULT_LIMITS.min_mem_free_pct);
    let source = if file.is_some() {
        LIMITS_FILE
    } else {
        "the default"
    };
    let mut notes = Vec::new();
    let mut note = |flag: &str, given: String, used: String| {
        notes.push(format!(
            "heavy-slot: {flag} {given} is looser than {source} ({used}); using {used}"
        ));
    };
    let max = match caller_max {
        Some(c) if c > base_max => {
            note("--max", c.to_string(), base_max.to_string());
            base_max
        }
        Some(c) => c,
        None => base_max,
    };
    let min_free_gb = match caller_free {
        Some(c) if c < base_free => {
            note("--min-free-gb", c.to_string(), base_free.to_string());
            base_free
        }
        Some(c) => c,
        None => base_free,
    };
    let max_swap_pct = match caller_swap {
        Some(c) if c > base_swap => {
            note("--max-swap-pct", c.to_string(), base_swap.to_string());
            base_swap
        }
        Some(c) => c,
        None => base_swap,
    };
    let min_mem_free_pct = match caller_mem_free {
        Some(c) if c < base_mem_free => {
            note(
                "--min-mem-free-pct",
                c.to_string(),
                base_mem_free.to_string(),
            );
            base_mem_free
        }
        Some(c) => c,
        None => base_mem_free,
    };
    (
        max,
        Limits {
            min_free_gb,
            max_swap_pct,
            min_mem_free_pct,
        },
        notes,
    )
}

/// The memory readings of a machine that reports memory pressure (macOS, card C-88); each part
/// is `None` when the system does not give it.
#[derive(Clone, Copy, Debug, Default)]
pub struct Memory {
    /// `kern.memorystatus_vm_pressure_level`: 1 normal, 2 warn, 4 critical.
    pub pressure_level: Option<u64>,
    /// `kern.memorystatus_level`: the free percentage `memory_pressure` prints.
    pub free_pct: Option<u64>,
}

/// What the limits are judged on: free bytes, swap percent, and the memory readings of a
/// machine that has them (`None` on Linux, which is judged on swap).
struct Machine {
    free: Option<u64>,
    swap: Option<u64>,
    memory: Option<Memory>,
}

/// The machine now. Debug builds take
/// `MOAI_CLI_TEST_MACHINE=free_gb=<n>,swap_pct=<n>[,pressure=<n>,mem_free_pct=<n>]` instead
/// (or `@<file>` holding that text, read on every call so a test can change the machine during
/// a wait); a `pressure` or `mem_free_pct` key makes it a machine with memory readings (macOS).
/// Release builds never read it.
fn measured(folder: &Path) -> Machine {
    #[cfg(debug_assertions)]
    if let Ok(mut s) = std::env::var("MOAI_CLI_TEST_MACHINE") {
        if let Some(file) = s.strip_prefix('@') {
            s = std::fs::read_to_string(file).unwrap_or_default();
        }
        let mut m = Machine {
            free: None,
            swap: None,
            memory: None,
        };
        for part in s.trim().split(',') {
            match part.split_once('=') {
                Some(("free_gb", v)) => m.free = v.parse::<u64>().ok().map(|g| g * 1_000_000_000),
                Some(("swap_pct", v)) => m.swap = v.parse().ok(),
                Some(("pressure", v)) => {
                    m.memory.get_or_insert_with(Memory::default).pressure_level = v.parse().ok()
                }
                Some(("mem_free_pct", v)) => {
                    m.memory.get_or_insert_with(Memory::default).free_pct = v.parse().ok()
                }
                _ => {}
            }
        }
        return m;
    }
    Machine {
        free: sys::free_bytes(folder),
        swap: sys::swap_pct(),
        memory: sys::memory(),
    }
}

/// Why a heavy run must not start now, or `None`. A value the system does not give is not a
/// reason to refuse (it is shown as `unknown` elsewhere). A machine with memory readings
/// (macOS) is judged on them and its swap is only shown; one without them (Linux) on swap.
pub fn over_limits(folder: &Path, limits: Limits) -> Option<String> {
    let Machine { free, swap, memory } = measured(folder);
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
    match memory {
        Some(mem) => {
            if let Some(l) = mem.pressure_level {
                if l >= CRITICAL_PRESSURE_LEVEL {
                    why.push(format!(
                        "memory pressure level {l} is at the critical level {CRITICAL_PRESSURE_LEVEL}"
                    ));
                }
            }
            if let Some(p) = mem.free_pct {
                if p < limits.min_mem_free_pct {
                    why.push(format!(
                        "memory free {p}% is below {}%",
                        limits.min_mem_free_pct
                    ));
                }
            }
            if let (false, Some(p)) = (why.is_empty(), swap) {
                why.push(format!("swap use {p}% (shown only; not a limit on macOS)"));
            }
        }
        None => {
            if let Some(p) = swap {
                if p > limits.max_swap_pct {
                    why.push(format!("swap use {p}% is above {}%", limits.max_swap_pct));
                }
            }
        }
    }
    (!why.is_empty()).then(|| why.join("; "))
}

/// #2930 item 6: free space (decimal GB, like the floor) under which a run that does start
/// still gets one warning line on stderr. The refusal is the floor (`min_free_gb`, default
/// 50); this mark only tells the lane the disk is getting short before the refusal comes.
pub const WARN_FREE_GB: u64 = 100;

/// The warning for a run that starts with `free` bytes left, or `None` when the system gives
/// no value or there are [`WARN_FREE_GB`] GB or more. `floor_gb` is the effective floor.
pub fn low_space_line(free: Option<u64>, floor_gb: u64, folder: &Path) -> Option<String> {
    let b = free.filter(|&b| b < WARN_FREE_GB.saturating_mul(1_000_000_000))?;
    Some(format!(
        "free disk space is under {WARN_FREE_GB} GB (disk free {:.1} GB at {}); the floor is {floor_gb} GB",
        b as f64 / 1e9,
        folder.display()
    ))
}

/// Where a wrapped command runs, for [`build_folder_note`]: its working folder, the worktree
/// that holds it, every worktree of the repository (real paths) and this worktree's build
/// folder `<HOME>/.moai/build/<name>` (card C-82).
#[derive(Debug, Clone)]
pub struct Here {
    pub cwd: PathBuf,
    pub worktree: PathBuf,
    pub worktrees: Vec<PathBuf>,
    pub folder: PathBuf,
}

impl Here {
    /// `None` outside a git worktree or without an absolute HOME: nothing to compare with.
    /// `home` and `userprofile` are the `HOME` and `USERPROFILE` values (the shared rule of
    /// `lane_env::home_dir`: USERPROFILE stands in on Windows when HOME is unset).
    pub fn of(
        cwd: &Path,
        home: Option<&std::ffi::OsStr>,
        userprofile: Option<&std::ffi::OsStr>,
    ) -> Option<Here> {
        let owner = crate::lane_env::owner(cwd).ok()?;
        let home = crate::lane_env::home_dir(home, userprofile, cfg!(windows)).ok()?;
        Some(Here {
            cwd: cwd.to_path_buf(),
            worktrees: crate::lane_env::worktrees(cwd).ok()?,
            folder: crate::lane_env::build_folder(&home, &owner.name),
            worktree: owner.worktree,
        })
    }
}

/// The wrapped command is `cargo` (by its file name, so a full path counts).
pub fn wraps_cargo(command: &[String]) -> bool {
    command
        .first()
        .and_then(|p| Path::new(p).file_name())
        .is_some_and(|n| n == "cargo")
}

/// `path` with its longest existing part made real, the rest appended as written.
fn settled(path: &Path) -> PathBuf {
    let mut rest = Vec::new();
    let mut at = path;
    loop {
        if let Ok(mut real) = std::fs::canonicalize(at) {
            real.extend(rest.iter().rev());
            return real;
        }
        match (at.file_name(), at.parent()) {
            (Some(n), Some(parent)) => {
                rest.push(n);
                at = parent;
            }
            _ => return path.to_path_buf(),
        }
    }
}

/// #2879 D3, revised for card C-82 (user 2026-11-17 08:22Z): when the wrapped command is `cargo`
/// and its `CARGO_TARGET_DIR` is shared with another worktree — the former per-lane
/// `moai-rc-target-*` folder, another worktree's folder under `~/.moai/build/`, a folder whose
/// owner record names another worktree — or lies inside another worktree, a one-line warning for
/// stderr. A warning, not a refusal. `None` when the name is no lane name, `here` is unknown, the
/// value is unset (cargo then builds in this worktree's own `target/`), or the folder is this
/// worktree's own build folder or inside this worktree.
pub fn build_folder_note(
    name: &str,
    command: &[String],
    target_dir: Option<&std::ffi::OsStr>,
    here: Option<&Here>,
) -> Option<String> {
    if !wraps_cargo(command) || crate::lane_env::check_lane(name).is_err() {
        return None;
    }
    let here = here?;
    let t = settled(&here.cwd.join(target_dir.filter(|t| !t.is_empty())?));
    let own = settled(&here.folder);
    let owner = crate::lane_env::read_record(&t)
        .ok()
        .flatten()
        .and_then(|r| r["worktree"].as_str().map(PathBuf::from));
    // C-82 slot audit N3: a folder whose record names this worktree is its own, whatever its
    // name (for example made before a rename); only another worktree's folder is warned about.
    let own_by_record = owner.as_ref().is_some_and(|w| *w == here.worktree);
    let recorded = owner.filter(|w| *w != here.worktree);
    let holder = here
        .worktrees
        .iter()
        .filter(|w| t.starts_with(w))
        .max_by_key(|w| w.components().count())
        .filter(|w| **w != here.worktree);
    let why = if t
        .file_name()
        .is_some_and(|n| n.to_string_lossy().starts_with("moai-rc-target-"))
    {
        "the per-lane folder every worktree of the lane shares (#2841)".to_string()
    } else if let Some(w) = recorded {
        format!(
            "a build folder whose record names the worktree {}",
            w.display()
        )
    } else if t != own && t.parent() == own.parent() && !own_by_record {
        "another worktree's build folder".to_string()
    } else if let Some(w) = holder {
        format!("inside the worktree {}", w.display())
    } else {
        return None;
    };
    Some(format!(
        "heavy-slot: CARGO_TARGET_DIR {} is {why}; this worktree's build folder is {} (internal/build-folder, card C-82): run `e=$(moai lane-env --lane {name}) && eval \"$e\"` first",
        t.display(),
        here.folder.display()
    ))
}

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
    pub fn memory() -> Option<super::Memory> {
        None
    }
}

#[cfg(unix)]
mod sys {
    use super::{
        low_space_line, machine_line, measured, over_limits, Limits, Path, PathBuf, Result, Value,
        BUSY, REFUSED, USAGE,
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

    /// macOS memory readings (card C-88), both by `sysctlbyname`, no child process:
    /// `kern.memorystatus_vm_pressure_level` and `kern.memorystatus_level`, the "System-wide
    /// memory free percentage" that `memory_pressure` prints (on 2026-11-17 the pressure level
    /// read 2, and `kern.memorystatus_level` and `memory_pressure -Q` both read 37).
    #[cfg(target_os = "macos")]
    pub fn memory() -> Option<super::Memory> {
        fn int(name: &std::ffi::CStr) -> Option<u64> {
            let mut v: libc::c_int = 0;
            let mut len = std::mem::size_of::<libc::c_int>();
            // SAFETY: sysctlbyname writes at most `len` bytes into the int it is given.
            let rc = unsafe {
                libc::sysctlbyname(
                    name.as_ptr(),
                    (&mut v as *mut libc::c_int).cast(),
                    &mut len,
                    std::ptr::null_mut(),
                    0,
                )
            };
            if rc != 0 || len != std::mem::size_of::<libc::c_int>() {
                return None;
            }
            u64::try_from(v).ok()
        }
        Some(super::Memory {
            pressure_level: int(c"kern.memorystatus_vm_pressure_level"),
            free_pct: int(c"kern.memorystatus_level"),
        })
    }

    /// Linux is judged on swap (card C-88 leaves it unchanged).
    #[cfg(not(target_os = "macos"))]
    pub fn memory() -> Option<super::Memory> {
        None
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
        // #2930 item 6: the run starts, but the disk is getting short — stderr only, so
        // stdout and the exit code stay COMMAND's own.
        if let Some(line) = low_space_line(measured(&cwd).free, limits.min_free_gb, &cwd) {
            eprintln!("heavy-slot: {line}");
        }
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

    /// Whether the process a slot record names is alive (#3033): `Some(true)` or `Some(false)`
    /// from `kill(pid, 0)`, `None` when the record has no usable pid — missing, not an integer,
    /// 0 or negative (those would signal a process group), or above `i32::MAX`. A live pid is
    /// not proof that it holds the lock: the lock belongs to the open file, and the pid may
    /// have been reused by an unrelated process.
    pub(super) fn alive_of(record: &Value) -> Option<bool> {
        let pid = record.get("pid")?.as_i64()?;
        if !(1..=i64::from(i32::MAX)).contains(&pid) {
            return None;
        }
        crate::home_state::pid_alive(pid)
    }

    /// One `held` entry: the record as read (any JSON object, `{}` when the file holds anything
    /// else) plus `alive`, true/false/null as [`alive_of`] says. When the recorded pid is dead
    /// (`alive` false) the slot is held by someone else, so the entry also lists the processes
    /// that have the slot file open (`holders`, with `holders_source` `proc` or `lsof`), or
    /// `holders` null with `holders_source` `unavailable` when neither source could be read —
    /// "unavailable" is not "nobody".
    fn held_entry(bytes: Option<Vec<u8>>, slot: &Path) -> Value {
        held_entry_with(bytes, slot, super::holders::find)
    }

    /// [`held_entry`] with the holder lookup passed in, so a test can reach the branch where
    /// neither `/proc` nor `lsof` can be read (`find` answers `None`).
    pub(super) fn held_entry_with<F>(bytes: Option<Vec<u8>>, slot: &Path, find: F) -> Value
    where
        F: FnOnce(&Path) -> Option<(Vec<super::holders::Holder>, &'static str)>,
    {
        let mut rec = bytes
            .and_then(|b| serde_json::from_slice::<Value>(&b).ok())
            .filter(Value::is_object)
            .unwrap_or_else(|| json!({}));
        let alive = alive_of(&rec);
        rec["alive"] = json!(alive);
        if alive == Some(false) {
            match find(slot) {
                Some((found, source)) => {
                    rec["holders"] = json!(found
                        .iter()
                        .map(|h| json!({"pid": h.pid, "comm": h.comm}))
                        .collect::<Vec<_>>());
                    rec["holders_source"] = json!(source);
                }
                None => {
                    rec["holders"] = Value::Null;
                    rec["holders_source"] = json!("unavailable");
                }
            }
        }
        rec
    }

    /// The slots held now (their records, each with `alive`) and slot paths that are not plain
    /// files. Reads only.
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
                Ok(true) => held.push(held_entry(fs::read(&p).ok(), &p)),
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

/// #3033 D2: which processes have a slot file open. This answers "who has this file open",
/// not "who owns the lock": a `flock` belongs to an open file description, and no portable
/// read says which of the processes sharing one it is. A slot whose recorded pid is dead is
/// held through a descendant that inherited the file; the open files show it.
#[cfg(unix)]
mod holders {
    use std::fs;
    use std::io::Read;
    use std::os::unix::fs::MetadataExt;
    use std::path::Path;
    use std::process::{Command, Stdio};
    use std::time::{Duration, Instant};

    /// `lsof` that has not answered by now is given up on (the answer is then "unavailable").
    const LSOF_TIMEOUT: Duration = Duration::from_secs(5);
    /// Longest process name kept; names are cut and control characters replaced.
    const MAX_COMM: usize = 64;

    /// A process that has the slot file open. No argv: it can carry secrets.
    #[derive(Debug, PartialEq, Eq)]
    pub struct Holder {
        pub pid: i64,
        pub comm: String,
    }

    fn clean_comm(s: &str) -> String {
        s.trim_end_matches(['\n', '\r'])
            .chars()
            .map(|c| if c.is_control() { '?' } else { c })
            .take(MAX_COMM)
            .collect()
    }

    /// The processes of `lsof -F pc` output: a `p<pid>` line starts a process, its `c<name>` line
    /// names it. A process with an unreadable pid is dropped with its fields; a process listed
    /// twice counts once.
    pub fn parse_lsof(out: &str) -> Vec<Holder> {
        let mut found: Vec<Holder> = Vec::new();
        let mut current: Option<usize> = None;
        for line in out.lines() {
            if let Some(p) = line.strip_prefix('p') {
                current = match p.parse::<i64>() {
                    Ok(pid) if pid > 0 && p.bytes().all(|b| b.is_ascii_digit()) => {
                        match found.iter().position(|h| h.pid == pid) {
                            Some(i) => Some(i),
                            None => {
                                found.push(Holder {
                                    pid,
                                    comm: String::new(),
                                });
                                Some(found.len() - 1)
                            }
                        }
                    }
                    _ => None,
                };
            } else if let Some(c) = line.strip_prefix('c') {
                if let Some(i) = current {
                    if found[i].comm.is_empty() {
                        found[i].comm = clean_comm(c);
                    }
                }
            }
        }
        found
    }

    /// The processes under `proc_root` (a `/proc`) with an open file that is the same file as
    /// `slot` — the same device and inode, so a hard link or another path to it counts and a
    /// path string is never compared. `None` when `slot` cannot be read. A process whose `fd`
    /// folder this user cannot read is not seen.
    pub fn scan_proc(proc_root: &Path, slot: &Path) -> Option<Vec<Holder>> {
        let m = fs::metadata(slot).ok()?;
        let want = (m.dev(), m.ino());
        let mut out = Vec::new();
        for e in fs::read_dir(proc_root).ok()?.flatten() {
            let name = e.file_name();
            let Some(name) = name.to_str() else { continue };
            if name.is_empty() || !name.bytes().all(|b| b.is_ascii_digit()) {
                continue;
            }
            let Ok(pid) = name.parse::<i64>() else {
                continue;
            };
            let Ok(fds) = fs::read_dir(e.path().join("fd")) else {
                continue;
            };
            let hit = fds
                .flatten()
                .any(|fd| fs::metadata(fd.path()).is_ok_and(|m| (m.dev(), m.ino()) == want));
            if hit {
                let comm = fs::read_to_string(e.path().join("comm")).unwrap_or_default();
                out.push(Holder {
                    pid,
                    comm: clean_comm(&comm),
                });
            }
        }
        out.sort_by_key(|h| h.pid);
        Some(out)
    }

    /// `lsof -w -F pc -- <slot>`: exit 0 or 1 is an answer (1 with no output means no process has
    /// the file open); another code, a failure to start or no answer within `timeout` is `None`.
    /// `lsof` is the program and its leading arguments (`["lsof"]`; a test passes `["sh",
    /// "<script>"]` so no freshly written file is executed).
    pub fn run_lsof(lsof: &[&str], slot: &Path, timeout: Duration) -> Option<Vec<Holder>> {
        let (program, lead) = lsof.split_first()?;
        let mut child = Command::new(program)
            .args(lead)
            .args(["-w", "-F", "pc", "--"])
            .arg(slot)
            .env("LC_ALL", "C")
            .stdin(Stdio::null())
            .stdout(Stdio::piped())
            .stderr(Stdio::null())
            .spawn()
            .ok()?;
        let start = Instant::now();
        let status = loop {
            match child.try_wait() {
                Ok(Some(s)) => break s,
                Ok(None) if start.elapsed() < timeout => {
                    std::thread::sleep(Duration::from_millis(10))
                }
                _ => {
                    let _ = child.kill();
                    let _ = child.wait();
                    return None;
                }
            }
        };
        if !matches!(status.code(), Some(0 | 1)) {
            return None;
        }
        let mut text = String::new();
        child.stdout.take()?.read_to_string(&mut text).ok()?;
        Some(parse_lsof(&text))
    }

    /// The holders of `slot` and where they were read: `/proc` when `proc_root` has one
    /// (`proc`), else `lsof` (`lsof`). The calling process is not listed. `None` when neither
    /// source can be read.
    pub fn find_with(
        proc_root: &Path,
        lsof: &[&str],
        slot: &Path,
        timeout: Duration,
    ) -> Option<(Vec<Holder>, &'static str)> {
        let me = i64::from(std::process::id());
        let (mut found, source) = if proc_root.join("self/fd").exists() {
            (scan_proc(proc_root, slot)?, "proc")
        } else {
            (run_lsof(lsof, slot, timeout)?, "lsof")
        };
        found.retain(|h| h.pid != me);
        Some((found, source))
    }

    pub fn find(slot: &Path) -> Option<(Vec<Holder>, &'static str)> {
        find_with(Path::new("/proc"), &["lsof"], slot, LSOF_TIMEOUT)
    }
}

#[cfg(all(test, unix))]
mod holders_tests {
    use super::holders::{find_with, parse_lsof, run_lsof, scan_proc, Holder};
    use std::fs;
    use std::path::Path;
    use std::time::Duration;

    fn h(pid: i64, comm: &str) -> Holder {
        Holder {
            pid,
            comm: comm.into(),
        }
    }

    #[test]
    fn lsof_output_is_parsed() {
        let out = "p123\ncsleep\np456\ncsh\n";
        assert_eq!(parse_lsof(out), vec![h(123, "sleep"), h(456, "sh")]);
        // A process with no name line, a name line before any process, and the same pid twice.
        assert_eq!(
            parse_lsof("cstray\np7\np8\ncx\np7\ncy\n"),
            vec![h(7, "y"), h(8, "x")]
        );
        // An unreadable pid drops that process and its name, not the next one.
        assert_eq!(
            parse_lsof("pabc\ncbad\np-3\ncbad\np0\ncbad\np+5\ncbad\np9\ncok\n"),
            vec![h(9, "ok")]
        );
        assert!(parse_lsof("").is_empty());
        assert!(parse_lsof("f3\nn/some/file\n").is_empty());
        // A name is cut to 64 characters and control characters are replaced.
        let long = format!("p1\nc{}\n", "x".repeat(200));
        assert_eq!(parse_lsof(&long)[0].comm.len(), 64);
        assert_eq!(parse_lsof("p1\ncab\u{1b}[31mc\n")[0].comm, "ab?[31mc");
    }

    /// A fake `/proc`: the processes that have the slot file open are found by device and inode —
    /// also through a hard link under another name — and nothing else is.
    #[test]
    fn holders_from_proc_match_by_inode() {
        let d = tempfile::tempdir().unwrap();
        let slot = d.path().join("slot-0");
        fs::write(&slot, "{}").unwrap();
        let alias = d.path().join("alias");
        fs::hard_link(&slot, &alias).unwrap();
        let other = d.path().join("other");
        fs::write(&other, "x").unwrap();
        let proc = d.path().join("proc");
        let mk = |pid: &str, fds: &[(&str, &Path)], comm: Option<&str>| {
            let fd = proc.join(pid).join("fd");
            fs::create_dir_all(&fd).unwrap();
            for (n, target) in fds {
                std::os::unix::fs::symlink(target, fd.join(n)).unwrap();
            }
            if let Some(c) = comm {
                fs::write(proc.join(pid).join("comm"), c).unwrap();
            }
        };
        mk("self", &[("0", &slot)], Some("me\n")); // not a number: skipped
        mk("100", &[("3", &alias)], Some("sleep\n")); // by hard link
        mk("200", &[("0", &other)], Some("noise\n")); // another file
        mk("300", &[], Some("nofd\n")); // no open file
        mk("400", &[("5", &slot), ("6", &slot)], None); // twice open, no name
        mk("12x", &[("3", &slot)], Some("bad\n")); // not a number: skipped
        fs::create_dir_all(proc.join("500")).unwrap(); // no fd folder
        let got = scan_proc(&proc, &slot).unwrap();
        assert_eq!(got, vec![h(100, "sleep"), h(400, "")]);
        // A slot that cannot be read gives no answer.
        assert!(scan_proc(&proc, &d.path().join("missing")).is_none());
    }

    #[test]
    fn find_prefers_proc_and_leaves_out_the_caller() {
        let d = tempfile::tempdir().unwrap();
        let slot = d.path().join("slot-0");
        fs::write(&slot, "{}").unwrap();
        let proc = d.path().join("proc");
        let me = std::process::id().to_string();
        for pid in [me.as_str(), "77"] {
            let fd = proc.join(pid).join("fd");
            fs::create_dir_all(&fd).unwrap();
            std::os::unix::fs::symlink(&slot, fd.join("3")).unwrap();
        }
        fs::create_dir_all(proc.join("self/fd")).unwrap();
        let (found, source) =
            find_with(&proc, &["/nonexistent/lsof"], &slot, Duration::from_secs(1)).unwrap();
        assert_eq!(source, "proc");
        assert_eq!(found, vec![h(77, "")]);
    }

    /// A stub `lsof` script, run as `sh <script>` (a file just written is never executed, which
    /// would race with a fork in another test thread: ETXTBSY).
    fn stub(dir: &Path, body: &str) -> String {
        let p = dir.join("lsof.sh");
        fs::write(&p, format!("{body}\n")).unwrap();
        p.display().to_string()
    }

    /// `lsof` exit 0 and 1 are answers; any other exit, a missing program and a program that
    /// does not answer in time are "unavailable" (`None`), never an empty list.
    #[test]
    fn lsof_exit_codes_and_silence_are_told_apart() {
        let d = tempfile::tempdir().unwrap();
        let slot = d.path().join("slot-0");
        fs::write(&slot, "{}").unwrap();
        let t = Duration::from_millis(300);
        let run = |body: &str| {
            let script = stub(d.path(), body);
            run_lsof(&["sh", &script], &slot, t)
        };
        assert_eq!(run("printf 'p5\\ncfoo\\n'"), Some(vec![h(5, "foo")]));
        assert_eq!(run("exit 1"), Some(vec![]));
        assert_eq!(
            run("printf 'p6\\ncbar\\n'; exit 1"),
            Some(vec![h(6, "bar")])
        );
        assert_eq!(run("printf 'p5\\ncfoo\\n'; exit 2"), None);
        assert_eq!(run("exec sleep 30"), None);
        assert_eq!(run_lsof(&["/nonexistent/lsof"], &slot, t), None);
        assert_eq!(run_lsof(&[], &slot, t), None);
        // No /proc and no lsof: unavailable.
        let noproc = d.path().join("noproc");
        assert!(find_with(&noproc, &["/nonexistent/lsof"], &slot, t).is_none());
    }
}

/// #3033 D3: the four places that describe the slot say the same thing — a child or descendant
/// that outlives the command keeps the slot, status shows who, and no one edits a slot file.
#[cfg(test)]
mod docs_tests {
    #[test]
    fn the_four_descriptions_name_the_leftover_child_and_the_hands_off_rule() {
        let docs = [
            ("src/heavy_slot.rs", include_str!("heavy_slot.rs")),
            (
                "skills/moai/workflows/verify.md",
                include_str!("../skills/moai/workflows/verify.md"),
            ),
            (
                "docs/rust-product-path.md",
                include_str!("../docs/rust-product-path.md"),
            ),
            (
                "tools/heavy_check_slot.py",
                include_str!("../tools/heavy_check_slot.py"),
            ),
        ];
        for (name, text) in docs {
            assert!(text.contains("descendant"), "{name}: leftover descendant");
            assert!(
                text.contains("never edited or deleted by hand"),
                "{name}: hands off the slot file"
            );
            assert!(text.contains("`alive`"), "{name}: names alive");
        }
        // Only the Rust status has `holders`; the Python tool says so instead of listing it.
        for (name, text) in docs {
            assert!(text.contains("`holders`"), "{name}: names holders");
        }
    }
}

#[cfg(all(test, unix))]
mod alive_tests {
    use super::holders::Holder;
    use super::sys::{alive_of, held_entry_with};
    use serde_json::{json, Value};
    use std::path::Path;

    /// A pid that was alive a moment ago and has been reaped.
    fn dead_pid() -> i64 {
        let mut child = std::process::Command::new("true").spawn().unwrap();
        let pid = i64::from(child.id());
        child.wait().unwrap();
        pid
    }

    fn entry_for(
        record: &Value,
        find: impl FnOnce(&Path) -> Option<(Vec<Holder>, &'static str)>,
    ) -> Value {
        held_entry_with(
            Some(record.to_string().into_bytes()),
            Path::new("slot-0"),
            find,
        )
    }

    /// #3033 D1: `alive` follows `kill(pid, 0)` for a pid in 1..=i32::MAX and is null for
    /// anything the record cannot name a process by.
    #[test]
    fn the_recorded_pid_decides_alive() {
        let me = i64::from(std::process::id());
        let mut child = std::process::Command::new("true").spawn().unwrap();
        let reaped = i64::from(child.id());
        child.wait().unwrap();
        let cases: Vec<(Value, Option<bool>)> = vec![
            (json!({"pid": me}), Some(true)),
            (json!({"pid": reaped}), Some(false)),
            (json!({"pid": 1}), Some(true)),
            (json!({"pid": i32::MAX}), Some(false)),
            (json!({"pid": 2_147_483_648_i64}), None),
            (json!({"pid": u64::MAX}), None),
            (json!({"pid": 0}), None),
            (json!({"pid": -1}), None),
            (json!({"pid": i64::MIN}), None),
            (json!({"pid": "1"}), None),
            (json!({"pid": 1.5}), None),
            (json!({"pid": null}), None),
            (json!({"pid": [1]}), None),
            (json!({"name": "x"}), None),
            (json!({}), None),
            (json!([1]), None),
            (json!(7), None),
        ];
        for (record, want) in cases {
            assert_eq!(alive_of(&record), want, "{record}");
        }
    }

    /// #3033 B1: a slot whose recorded pid is dead is held through someone else. When neither
    /// `/proc` nor `lsof` can be read the entry says `holders` null and `holders_source`
    /// `unavailable` — never an empty list, which would read as "nobody holds it" and invite a
    /// hand edit of the slot file. The same entry with a source that answered "nobody" is an
    /// empty list from that source, and with an answer it lists the processes.
    #[test]
    fn unreadable_holders_are_unavailable_never_an_empty_list() {
        let record = json!({"pid": dead_pid(), "name": "cargo"});
        let e = entry_for(&record, |_| None);
        assert_eq!(e["alive"], json!(false), "{e}");
        assert!(e["holders"].is_null(), "unavailable is not nobody: {e}");
        assert_eq!(e["holders_source"], json!("unavailable"), "{e}");
        assert_eq!(e["name"], json!("cargo"), "the record is kept: {e}");

        let e = entry_for(&record, |_| Some((vec![], "proc")));
        assert_eq!(e["holders"], json!([]), "{e}");
        assert_eq!(e["holders_source"], json!("proc"), "{e}");

        let e = entry_for(&record, |_| {
            Some((
                vec![Holder {
                    pid: 41,
                    comm: "sleep".into(),
                }],
                "lsof",
            ))
        });
        assert_eq!(e["holders"], json!([{"pid": 41, "comm": "sleep"}]), "{e}");
        assert_eq!(e["holders_source"], json!("lsof"), "{e}");
    }

    /// Only a recorded pid that is dead asks where the file is open: a live pid, a record with no
    /// usable pid and bytes that are not a JSON object never do, and their entries carry neither
    /// `holders` nor `holders_source`.
    #[test]
    fn only_a_dead_pid_asks_for_holders() {
        let me = i64::from(std::process::id());
        let asked = |_: &Path| -> Option<(Vec<Holder>, &'static str)> {
            panic!("the holders were asked for")
        };
        for record in [json!({"pid": me}), json!({"pid": "x"}), json!({})] {
            let e = entry_for(&record, asked);
            assert!(e.get("holders").is_none(), "{record}: {e}");
            assert!(e.get("holders_source").is_none(), "{record}: {e}");
        }
        for bytes in [None, Some(b"not json".to_vec()), Some(b"[1]".to_vec())] {
            let e = held_entry_with(bytes, Path::new("slot-0"), asked);
            assert_eq!(e["alive"], Value::Null, "{e}");
            assert!(e.get("holders").is_none(), "{e}");
        }
    }
}

#[cfg(test)]
mod limits_tests {
    use super::{effective, LimitsFile, DEFAULT_MAX};

    fn file(max: Option<usize>, free: Option<u64>, swap: Option<u64>) -> LimitsFile {
        LimitsFile {
            max,
            min_free_gb: free,
            max_swap_pct: swap,
            min_mem_free_pct: None,
        }
    }

    #[test]
    fn no_file_and_no_flags_use_the_defaults() {
        let (max, l, notes) = effective(None, None, None, None, None);
        assert_eq!((max, l.min_free_gb, l.max_swap_pct), (DEFAULT_MAX, 50, 80));
        assert_eq!(l.min_mem_free_pct, 10);
        assert_eq!(DEFAULT_MAX, 4);
        assert!(notes.is_empty());
    }

    #[test]
    fn flags_only_tighten() {
        // Looser flags lose to the defaults, each with a note.
        let (max, l, notes) = effective(None, Some(9), Some(0), Some(100), None);
        assert_eq!((max, l.min_free_gb, l.max_swap_pct), (DEFAULT_MAX, 50, 80));
        assert_eq!(notes.len(), 3, "{notes:?}");
        assert!(notes[1].contains("--min-free-gb 0 is looser than the default (50)"));
        // Stricter flags win, silently; equal flags too.
        let (max, l, notes) = effective(None, Some(1), Some(51), Some(79), None);
        assert_eq!((max, l.min_free_gb, l.max_swap_pct), (1, 51, 79));
        assert!(notes.is_empty());
        let (max, l, notes) = effective(None, Some(3), Some(50), Some(80), None);
        assert_eq!((max, l.min_free_gb, l.max_swap_pct), (3, 50, 80));
        assert!(notes.is_empty());
        // The file replaces the defaults, also in the looser direction.
        let f = file(Some(1), Some(10), Some(95));
        let (max, l, notes) = effective(Some(&f), None, None, None, None);
        assert_eq!((max, l.min_free_gb, l.max_swap_pct), (1, 10, 95));
        assert!(notes.is_empty());
        let (max, l, notes) = effective(Some(&f), Some(2), Some(9), Some(96), None);
        assert_eq!((max, l.min_free_gb, l.max_swap_pct), (1, 10, 95));
        assert!(notes.iter().all(|n| n.contains("limits.json")), "{notes:?}");
        assert_eq!(notes.len(), 3);
    }

    /// Card C-88: the memory free limit only tightens too — 9 loses to the default 10 with a
    /// note, 10 and 11 are used silently, and the file's value replaces the default both ways.
    #[test]
    fn the_memory_free_flag_only_tightens() {
        let (_, l, notes) = effective(None, None, None, None, Some(9));
        assert_eq!(l.min_mem_free_pct, 10);
        assert_eq!(
            notes,
            ["heavy-slot: --min-mem-free-pct 9 is looser than the default (10); using 10"]
        );
        for c in [10, 11] {
            let (_, l, notes) = effective(None, None, None, None, Some(c));
            assert_eq!(l.min_mem_free_pct, c);
            assert!(notes.is_empty(), "{notes:?}");
        }
        let mut f = file(None, None, None);
        f.min_mem_free_pct = Some(0);
        let (_, l, notes) = effective(Some(&f), None, None, None, None);
        assert_eq!(l.min_mem_free_pct, 0);
        assert!(notes.is_empty());
        f.min_mem_free_pct = Some(20);
        let (_, l, notes) = effective(Some(&f), None, None, None, Some(19));
        assert_eq!(l.min_mem_free_pct, 20);
        assert!(
            notes[0].contains("is looser than limits.json (20)"),
            "{notes:?}"
        );
    }

    #[test]
    fn extreme_values_are_capped_by_the_stricter_side() {
        let (max, l, _) = effective(
            None,
            Some(usize::MAX),
            Some(u64::MAX),
            Some(u64::MAX),
            Some(u64::MAX),
        );
        assert_eq!(
            (max, l.min_free_gb, l.max_swap_pct),
            (DEFAULT_MAX, u64::MAX, 80)
        );
        assert_eq!(l.min_mem_free_pct, u64::MAX);
        let f = file(Some(usize::MAX), Some(u64::MAX), Some(u64::MAX));
        let (max, l, _) = effective(Some(&f), Some(2), Some(0), Some(100), None);
        assert_eq!((max, l.min_free_gb, l.max_swap_pct), (2, u64::MAX, 100));
    }
}

/// Card C-88: the real macOS readings come back, in their ranges (the pressure level is one of
/// the dispatch levels, the free percentage at most 100).
#[cfg(all(test, target_os = "macos"))]
mod memory_tests {
    #[test]
    fn macos_gives_a_pressure_level_and_a_free_percentage() {
        let m = super::sys::memory().expect("macOS has memory readings");
        let level = m
            .pressure_level
            .expect("kern.memorystatus_vm_pressure_level");
        assert!([1, 2, 4].contains(&level), "{level}");
        let free = m.free_pct.expect("kern.memorystatus_level");
        assert!(free <= 100, "{free}");
    }
}

#[cfg(test)]
mod build_folder_tests {
    use super::{build_folder_note, Here};
    use std::ffi::OsStr;
    use std::path::PathBuf;

    fn cmd(v: &[&str]) -> Vec<String> {
        v.iter().map(|s| s.to_string()).collect()
    }

    /// Paths under a root that does not exist, so they stay as written.
    const R: &str = "/nonexistent-moai-c82";

    fn here() -> Here {
        Here {
            cwd: PathBuf::from(format!("{R}/r/repo/.claude/worktrees/one")),
            worktree: PathBuf::from(format!("{R}/r/repo/.claude/worktrees/one")),
            worktrees: vec![
                PathBuf::from(format!("{R}/r/repo")),
                PathBuf::from(format!("{R}/r/repo/.claude/worktrees/one")),
                PathBuf::from(format!("{R}/r/two")),
            ],
            folder: PathBuf::from(format!("{R}/h/.moai/build/repo-one")),
        }
    }

    fn note(target: Option<&str>) -> Option<String> {
        let cargo = cmd(&["cargo", "test", "--offline"]);
        build_folder_note("agent-7", &cargo, target.map(OsStr::new), Some(&here()))
    }

    #[test]
    fn own_folders_say_nothing() {
        for t in [
            None,
            Some(""),
            Some("target"),
            Some("target/../target"),
            Some(&*format!("{R}/h/.moai/build/repo-one")),
            Some(&*format!("{R}/r/repo/.claude/worktrees/one/target")),
            Some(&*format!("{R}/elsewhere/build")),
        ] {
            assert_eq!(note(t), None, "{t:?}");
        }
    }

    #[test]
    fn a_shared_or_foreign_folder_is_named() {
        for (t, says) in [
            (
                format!("{R}/tmp/moai-rc-target-agent-7"),
                "the per-lane folder",
            ),
            (
                format!("{R}/tmp/moai-rc-target-x-a2846"),
                "the per-lane folder",
            ),
            (
                format!("{R}/h/.moai/build/repo-two"),
                "another worktree's build folder",
            ),
            (
                format!("{R}/r/repo/target"),
                "inside the worktree /nonexistent-moai-c82/r/repo;",
            ),
            (
                format!("{R}/r/two/target"),
                "inside the worktree /nonexistent-moai-c82/r/two;",
            ),
        ] {
            let n = note(Some(&t)).unwrap_or_else(|| panic!("{t}"));
            assert!(n.contains(says), "{n}");
            assert!(
                n.starts_with(&format!("heavy-slot: CARGO_TARGET_DIR {t} is ")),
                "{n}"
            );
            assert!(
                n.contains(&format!("build folder is {R}/h/.moai/build/repo-one")),
                "{n}"
            );
            assert!(n.contains("moai lane-env --lane agent-7"), "{n}");
        }
        // A full path to cargo counts as cargo too.
        let full = cmd(&["/usr/local/bin/cargo", "build"]);
        let t = format!("{R}/r/two/target");
        assert!(build_folder_note("agent-7", &full, Some(OsStr::new(&t)), Some(&here())).is_some());
    }

    #[test]
    fn a_record_of_another_worktree_is_named() {
        let d = tempfile::TempDir::new().unwrap();
        let dir = std::fs::canonicalize(d.path()).unwrap();
        let rec = |w: &str| {
            std::fs::write(
                dir.join(".moai-build.json"),
                format!(r#"{{"schema_version": 1, "worktree": "{w}"}}"#),
            )
            .unwrap()
        };
        let t = dir.to_str().unwrap();
        rec(&format!("{R}/r/two"));
        assert!(note(Some(t)).unwrap().contains(
            "a build folder whose record names the worktree /nonexistent-moai-c82/r/two;"
        ));
        rec(&format!("{R}/r/repo/.claude/worktrees/one"));
        assert_eq!(note(Some(t)), None);
    }

    /// C-82 slot audit N3: a sibling build folder under another name is this worktree's own
    /// when its record says so (no warning); without that record it is another worktree's.
    #[test]
    fn a_sibling_folder_recorded_for_this_worktree_is_its_own() {
        let d = tempfile::TempDir::new().unwrap();
        let root = std::fs::canonicalize(d.path()).unwrap();
        let build = root.join("h/.moai/build");
        let old = build.join("repo-old-name");
        std::fs::create_dir_all(&old).unwrap();
        let mut h = here();
        h.folder = build.join("repo-one");
        let cargo = cmd(&["cargo", "test"]);
        let t = Some(old.as_os_str());
        let n = build_folder_note("agent-7", &cargo, t, Some(&h)).unwrap();
        assert!(n.contains("another worktree's build folder"), "{n}");
        std::fs::write(
            old.join(".moai-build.json"),
            format!(
                r#"{{"schema_version": 1, "worktree": "{}"}}"#,
                h.worktree.display()
            ),
        )
        .unwrap();
        assert_eq!(build_folder_note("agent-7", &cargo, t, Some(&h)), None);
    }

    #[test]
    fn other_commands_non_lane_names_and_unknown_places_are_left_alone() {
        let t = Some(OsStr::new("/nonexistent-moai-c82/r/two/target"));
        let h = here();
        let py = cmd(&["python3", "-m", "unittest"]);
        assert_eq!(build_folder_note("agent-7", &py, t, Some(&h)), None);
        let cargo = cmd(&["cargo", "test"]);
        assert_eq!(build_folder_note("a b", &cargo, t, Some(&h)), None);
        assert_eq!(build_folder_note("agent-7", &[], t, Some(&h)), None);
        assert_eq!(build_folder_note("agent-7", &cargo, t, None), None);
    }
}

#[cfg(test)]
mod low_space_tests {
    use super::{low_space_line, WARN_FREE_GB};
    use std::path::Path;

    const GB: u64 = 1_000_000_000;

    #[test]
    fn the_warning_mark_is_100_gb() {
        assert_eq!(WARN_FREE_GB, 100);
    }

    /// #2930 item 6: under 100 GB (decimal, like the floor) one line; 100 GB and up, none.
    #[test]
    fn the_edges_are_just_under_and_exactly_100_gb() {
        let f = Path::new("/work");
        for (free, warns) in [
            (Some(99 * GB), true),
            (Some(100 * GB - 1), true),
            (Some(100 * GB), false),
            (Some(101 * GB), false),
            (Some(0), true),
            (None, false),
        ] {
            assert_eq!(low_space_line(free, 50, f).is_some(), warns, "{free:?}");
        }
    }

    #[test]
    fn the_line_names_the_100_gb_mark_the_measured_space_and_the_floor() {
        let line = low_space_line(Some(87_300_000_000), 50, Path::new("/work")).unwrap();
        assert_eq!(
            line,
            "free disk space is under 100 GB (disk free 87.3 GB at /work); the floor is 50 GB"
        );
        assert!(low_space_line(Some(60 * GB), 61, Path::new("/w"))
            .unwrap()
            .ends_with("the floor is 61 GB"));
    }
}
