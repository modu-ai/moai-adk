//! Issue #2677 (AC-1, AC-2): a machine-wide limit on concurrent heavy runs — the product port
//! of the repository tool `tools/heavy_check_slot.py` (#2491), which was removed with the
//! repository's Python tools; the folder, file names and lock stay the ones it used.
//!
//! `heavy-slot-acquire [--max N] --name LANE [--dir DIR] -- COMMAND…` waits for one of N slots,
//! runs COMMAND and returns COMMAND's exit code (128 + the signal number when it ends on a
//! signal). A slot is a file `slot-<i>` with an exclusive `flock` (std `File::try_lock`); the
//! open file is handed to COMMAND, so a nested caller in COMMAND can reuse the slot. The
//! holder unlocks the slot when the holder ends (SPEC-20261125-heavy-slot-release-on-exit):
//! `flock(LOCK_UN)` on the holder's own open file description frees the slot for every
//! descriptor that shares it, so a child or descendant of COMMAND that outlives it (a
//! background process, a test child) and still has the file open no longer keeps the slot. A
//! caller that only reuses an inherited slot never unlocks the outer holder's slot. A holder
//! killed by SIGKILL cannot unlock: the slot then stays held until the wrapper's command and
//! every descendant that has the file open are gone, and the operating system frees it — no
//! slot file is left to clean. `status` then lists the slot with the wrapper's recorded pid,
//! which is dead. Each `held` entry of `status` says so (`alive`: true, false, or null when
//! the record names no process); for a dead pid it also lists the processes that have the slot
//! file open (`holders`, `pid` and `comm`, from `/proc` or `lsof`; `unavailable` with
//! `holders` null is not "nobody"), carries one `warning` line naming the slot, the dead pid
//! and those processes, and the same line is in the top-level `warnings` array (empty when no
//! slot has one). That is the list of processes with the file open, not a proof of who owns
//! the lock. Ending such a process is for its owner; a slot file is never edited or deleted
//! by hand. The Python twin has no `alive` or `holders`. A waiter holds `waiting/<ns>-<pid>-<max>` (its own `--max` last), and
//! its place is its rank among live waiters by name. Waiters start in arrival order
//! (SPEC-20261122-check-affected-one-slot REQ-4): every caller first creates and locks its own
//! waiting file. On each look it lists the free slots of its own `0..max`, lowest first, and
//! gives away one of them to every live (locked) waiter that arrived earlier and can use one of
//! those left — the lowest it can use, which is the slot that waiter takes itself; a waiting file
//! without a `<max>` (the Python twin, an older build) can use any slot. The caller takes a slot
//! that is still left, or waits when none is: so a caller never passes an earlier waiter that
//! can use the same slot, and never waits for a free slot of its own that no earlier waiter can
//! use (a narrower waiter queued ahead keeps no wider caller off a slot it cannot take itself).
//! One caller queues at the front: [`hold_ordered`] with `front` writes the stamp of its waiting
//! file in microseconds instead of nanoseconds, so the name sorts ahead of every other waiting
//! file and the caller starts before them (SPEC-20261126-leader-slot-front; `check affected
//! --full` run by the registered leader session). Front callers keep their arrival order among
//! themselves; nothing else about the order, the slots or the limits changes.
//! An unlocked waiting file is a dead waiter and is not counted. A caller that asks for a slot
//! first removes the waiting files of callers that are gone: plain files named like a waiter's
//! own whose pid is not alive and which no process holds locked (SPEC-20261123-heavy-slot-leftover
//! REQ-3); `status` only counts them (`stale_waiting`). The `waiting` folder must be a real
//! folder of the slot folder: it is opened once without following a symlink, and every look at
//! it and every removal in it goes through that descriptor (`openat`, `unlinkat`), so a link in
//! its place is refused (exit 125, nothing created, listed or removed behind it) and `status`
//! counts zero (REQ-5). Every command the wrapper starts gets
//! `CARGO_BUILD_JOBS` = the logical CPUs shared out over the `--max` slots, at least 2, unless
//! the caller or an outer holder already set a non-empty one; `status` prints that cap
//! (`build_jobs_cap`) for the slot count of `limits.json` (or the default), the most slots any
//! caller can use, so it is the lowest cap a caller gets: one that asks for fewer slots with a
//! smaller `--max` gets the larger share of the CPUs. A
//! process started under a held slot finds `MOAI_HEAVY_SLOT_FD=<fd>:<slot path>`
//! ([`SLOT_FD_ENV`]); asking for a slot in the same folder, it reuses that slot without queueing
//! only when the path names `slot-<digits>` in that folder, the descriptor is open on that very
//! file, someone holds the slot, and an exclusive non-blocking `flock` on the descriptor succeeds
//! — a `flock` belongs to the open file description, so that last step succeeds only for the
//! holder's own description handed down to this process, never for the slot file opened anew
//! (REQ-5). While waiting, and in the one exit line every started COMMAND gets on stderr
//! (`heavy-slot: <name> exited <code>; …`, 0 included), the load averages and the
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
//! `refused`). The default folder is `${XDG_CACHE_HOME:-~/.cache}/moai/heavy-slots`; builds from
//! before the name change used `…/moai-cli/heavy-slots`, whose `limits.json` is read while the
//! default folder has none and whose held slots `status` shows (`earlier_folder`).
//! Unix only; elsewhere both operations refuse.
//!
//! #2930 R-2: the machine's limits live in one file beside the slots, `<folder>/limits.json`
//! (`max`, `min_free_gb`, `max_swap_pct`, `min_mem_free_pct`; defaults 3, 50, 80, 10). A
//! caller's flag wins only when it is stricter. A limits file that is not a plain file of at most 4096 bytes holding only
//! those keys, or that sets `max` 0, is refused: acquire exits 125 without running COMMAND.
//!
//! The CPU rule (SPEC-20261126-heavy-slot-cpu-free): a caller that could take a slot waits while
//! the CPU free percentage is below `min_cpu_free_pct` of `limits.json` (default 0, which
//! switches the rule off; a value above 100, a negative number or a text is refused like `max`
//! 0). An earlier build refuses a file that holds this key, so write it only once every build
//! that reads the folder knows it. The rule is
//! a wait, not a refusal like low disk or memory (exit 75). The percentage is the idle share of
//! the CPU ticks between two readings at least 100 ms apart, so it is the load of now and not an
//! average since boot; a machine that gives no reading never makes a caller wait, and a slot
//! inherited from an outer holder is not judged again. The caller prints
//! `waiting for CPU headroom; cpu free <n>% is below <limit>% (<cores> cores)` when the wait
//! starts and again every 30 seconds while it lasts. The wait has a cap of 300 seconds: after
//! that COMMAND starts anyway, and the slot line says that the cap ended the wait. A caller that
//! waited ends its slot line and its exit line with `waited <n>s for CPU headroom` (the seconds
//! spent held back by the CPU rule, rounded up); a caller that did not wait gets no waiting
//! text. Whether the rule is on or off, the slot line of a slot this caller took names the
//! reading when the machine gives one, before any wait text: `cpu free <n>% (limit <m>%)` or
//! `cpu free <n>% (rule off)`.
//! `status` reports `cpu_free_pct` (`null` when the machine gives no reading), `cpu_cores`,
//! `min_cpu_free_pct` and `heavy_runs` (every held slot, one held by a dead pid included).

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
    /// Lowest CPU free percentage at which a caller takes a slot at once; below it the caller
    /// waits (at most [`CPU_WAIT_CAP`]). 0 switches the rule off.
    pub min_cpu_free_pct: u64,
}

/// The leader's numbers (comment 6850581277); `min_mem_free_pct` 10 is card C-88's. The CPU
/// rule is off until a limits file names a value (SPEC-20261126-heavy-slot-cpu-free D-7).
pub const DEFAULT_LIMITS: Limits = Limits {
    min_free_gb: 50,
    max_swap_pct: 80,
    min_mem_free_pct: 10,
    min_cpu_free_pct: 0,
};

/// The longest a caller waits for CPU headroom; after it COMMAND starts anyway.
pub const CPU_WAIT_CAP: std::time::Duration = std::time::Duration::from_secs(300);

/// `kern.memorystatus_vm_pressure_level` values: 1 normal, 2 warn, 4 critical (the
/// `DISPATCH_MEMORYPRESSURE_*` levels). A run is refused from this level on (card C-88).
pub const CRITICAL_PRESSURE_LEVEL: u64 = 4;

/// The slot count when neither the limits file nor the caller names one (#2930 R-2).
pub const DEFAULT_MAX: usize = 3;

/// The machine's limits, beside the slots: `<slot folder>/limits.json` (#2930 R-2, D-a).
pub const LIMITS_FILE: &str = "limits.json";

/// The variable a command run under a held slot finds: `<fd>:<slot path>`, the inherited open
/// slot file and where it is. A nested caller in the same folder reuses that slot (REQ-5).
pub const SLOT_FD_ENV: &str = "MOAI_HEAVY_SLOT_FD";

/// The `CARGO_BUILD_JOBS` every command started under a slot gets when the caller set none: the
/// logical CPUs shared out over the `slots`, at least 2 (SPEC-20261123-heavy-slot-leftover REQ-1;
/// the rule of `check affected`, so both give the same number).
pub fn build_jobs_cap(slots: usize) -> usize {
    let cpus = std::thread::available_parallelism().map_or(1, |n| n.get());
    crate::check_affected::default_jobs(cpus, slots)
}

/// The arrival stamp a waiting file starts with (SPEC-20261126-leader-slot-front REQ-3): the time
/// in nanoseconds, or in microseconds for the front place of a confirmed leader's wide check. A
/// waiting file keeps its three-part name `<20 digits>-<pid>-<max>` and is ordered as text, so a
/// microsecond stamp (16 digits, zero padded to 20) sorts ahead of every nanosecond one, in a
/// build without this change too; two front stamps keep their arrival order.
#[cfg(unix)]
fn waiting_stamp(front: bool, ns: u128) -> u128 {
    if front {
        ns / 1_000
    } else {
        ns
    }
}

/// The pid a waiting file names (`<digits>-<pid>` or `<digits>-<pid>-<max>`); `None` for any
/// other name, a pid outside `1..=i32::MAX`, or empty parts. A name that is not a waiter's own
/// is never cleaned up.
#[cfg(unix)]
fn waiter_pid(name: &str) -> Option<i64> {
    let parts: Vec<&str> = name.split('-').collect();
    if !(2..=3).contains(&parts.len())
        || parts
            .iter()
            .any(|p| p.is_empty() || !p.bytes().all(|b| b.is_ascii_digit()))
    {
        return None;
    }
    parts[1]
        .parse::<i64>()
        .ok()
        .filter(|p| (1..=i64::from(i32::MAX)).contains(p))
}

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
    pub min_cpu_free_pct: Option<u64>,
}

/// The limits file in `folder`: `None` when there is none. A path that is not a plain file
/// (a symlink, even to a valid file, or a folder), a file over 4096 bytes, JSON that does not
/// parse, an unknown key or `max` 0 is an error — the caller runs nothing. For the default
/// folder without a file of its own, the earlier folder's file is read by the same rules.
pub fn read_limits_file(folder: &Path) -> Result<Option<LimitsFile>> {
    match read_limits_in(folder)? {
        Some(f) => Ok(Some(f)),
        None => match earlier_dir(folder) {
            Some(earlier) => read_limits_in(&earlier),
            None => Ok(None),
        },
    }
}

fn read_limits_in(folder: &Path) -> Result<Option<LimitsFile>> {
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
    // Read from the value before serde sees it, so the refusal names the key: a negative
    // number, a text or a fraction would otherwise only say "expected u64".
    // `null` means the key is not set, as for every other key.
    if let Some(v) = value.get("min_cpu_free_pct").filter(|v| !v.is_null()) {
        if v.as_u64().is_none_or(|p| p > 100) {
            return Err(format!(
                "{}: min_cpu_free_pct must be a whole number from 0 to 100 (0 switches the rule off), not {v}",
                path.display()
            ));
        }
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
    // No caller flag for the CPU rule: the machine's file is the one place that sets it.
    let min_cpu_free_pct = file
        .and_then(|f| f.min_cpu_free_pct)
        .unwrap_or(DEFAULT_LIMITS.min_cpu_free_pct);
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
            min_cpu_free_pct,
        },
        notes,
    )
}

/// The limits `heavy-slot-acquire` uses in `folder`: its limits file read and the caller's
/// values applied by [`effective`]. An unusable limits file is an error — the caller runs nothing
/// and exits [`REFUSED`].
pub fn resolve(
    folder: &Path,
    caller_max: Option<usize>,
    caller_free: Option<u64>,
    caller_swap: Option<u64>,
    caller_mem_free: Option<u64>,
) -> Result<(usize, Limits, Vec<String>)> {
    let file = read_limits_file(folder)?;
    Ok(effective(
        file.as_ref(),
        caller_max,
        caller_free,
        caller_swap,
        caller_mem_free,
    ))
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
    /// The injected CPU free percentage (`cpu_free_pct`); never set from the real machine, whose
    /// CPU is read by [`CpuGauge`] over two readings.
    #[cfg_attr(not(unix), allow(dead_code))]
    cpu_free_pct: Option<u64>,
    /// The injected CPU count (`cpu_cores`).
    #[cfg_attr(not(unix), allow(dead_code))]
    cpu_cores: Option<u64>,
}

/// The machine a debug build is told to have: `MOAI_TEST_MACHINE`, or `None` when it is unset
/// and in every release build, which never reads it. The value is
/// `free_gb=<n>,swap_pct=<n>[,pressure=<n>,mem_free_pct=<n>][,cpu_free_pct=<0..100>,cpu_cores=<n>]`
/// (or `@<file>` holding that text, read on every call so a test can change the machine during
/// a wait); a `pressure` or `mem_free_pct` key makes it a machine with memory readings (macOS). A
/// machine that is injected without a `cpu_free_pct` key gives no CPU reading at all.
fn injected_machine() -> Option<Machine> {
    #[cfg(debug_assertions)]
    if let Ok(mut s) = std::env::var("MOAI_TEST_MACHINE") {
        if let Some(file) = s.strip_prefix('@') {
            s = std::fs::read_to_string(file).unwrap_or_default();
        }
        let mut m = Machine {
            free: None,
            swap: None,
            memory: None,
            cpu_free_pct: None,
            cpu_cores: None,
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
                Some(("cpu_free_pct", v)) => m.cpu_free_pct = v.parse().ok(),
                Some(("cpu_cores", v)) => m.cpu_cores = v.parse().ok(),
                _ => {}
            }
        }
        return Some(m);
    }
    None
}

/// The machine now: the injected one in a debug build that has one ([`injected_machine`]),
/// otherwise the system's readings.
fn measured(folder: &Path) -> Machine {
    if let Some(m) = injected_machine() {
        return m;
    }
    Machine {
        free: sys::free_bytes(folder),
        swap: sys::swap_pct(),
        memory: sys::memory(),
        cpu_free_pct: None,
        cpu_cores: None,
    }
}

/// One CPU tick reading as a pair `(idle, total)`, wrapping at 32 bits: the macOS counters
/// are 32 bits wide and wrap, and a difference over a short pause is far below 2^31, so the
/// wrapped difference is the true one.
#[cfg(unix)]
type Ticks = (u32, u32);

/// The share of the CPU that was idle between two tick readings, in whole percent (rounded
/// down), or `None` when the readings give no usable difference: no tick passed, or the total
/// went backwards (a difference above 2^31 cannot be a wrapped short pause).
#[cfg(unix)]
fn cpu_free_pct(prev: Ticks, next: Ticks) -> Option<u64> {
    let total = next.1.wrapping_sub(prev.1);
    let idle = next.0.wrapping_sub(prev.0);
    if total == 0 || total > i32::MAX as u32 {
        return None;
    }
    Some(u64::from(idle.min(total)) * 100 / u64::from(total))
}

/// The `(idle, total)` ticks of the first `cpu ` line of a `/proc/stat` text: idle counts the
/// idle and iowait fields, total the first eight fields (user, nice, system, idle, iowait, irq,
/// softirq, steal; guest time is already inside user). `None` when there is no such line or a
/// field is not a number. Cut to 32 bits like [`Ticks`].
#[cfg(all(unix, any(target_os = "linux", test)))]
fn parse_proc_stat(text: &str) -> Option<Ticks> {
    let line = text.lines().find(|l| l.starts_with("cpu "))?;
    let fields = line
        .split_whitespace()
        .skip(1)
        .take(8)
        .map(|f| f.parse::<u64>().ok())
        .collect::<Option<Vec<u64>>>()?;
    if fields.len() < 4 {
        return None;
    }
    let idle = fields[3].wrapping_add(fields.get(4).copied().unwrap_or(0));
    let total = fields.iter().fold(0u64, |a, f| a.wrapping_add(*f));
    Some((idle as u32, total as u32))
}

/// What a caller does at one look when it could take a slot: [`CpuStep::Wait`] for CPU
/// headroom or [`CpuStep::Run`] (take the slot).
#[cfg(unix)]
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum CpuStep {
    Run,
    Wait,
}

/// Wait while the CPU is `low` and the time already spent held back is under [`CPU_WAIT_CAP`];
/// at the cap the caller starts anyway.
#[cfg(unix)]
fn cpu_wait_decision(low: bool, waited: std::time::Duration) -> CpuStep {
    if low && waited < CPU_WAIT_CAP {
        CpuStep::Wait
    } else {
        CpuStep::Run
    }
}

/// The CPU readings a caller has taken for the CPU rule: `Some((free, cores))` of a debug build
/// that was given a machine ([`injected_machine`]; `free` is `None` without a `cpu_free_pct`
/// key), `None` when the real machine is read.
#[cfg(unix)]
fn cpu_injected() -> Option<(Option<u64>, Option<u64>)> {
    injected_machine().map(|m| (m.cpu_free_pct, m.cpu_cores))
}

/// The logical CPUs, the same count `build_jobs_cap` shares out.
#[cfg(unix)]
fn logical_cpus() -> u64 {
    std::thread::available_parallelism().map_or(1, |n| n.get()) as u64
}

/// The CPU count the CPU rule and the status report: the injected one, else the logical CPUs.
#[cfg(unix)]
fn cpu_cores() -> u64 {
    cpu_injected()
        .and_then(|(_, cores)| cores)
        .unwrap_or_else(logical_cpus)
}

/// The shortest time between the two tick readings a CPU percentage is made of.
#[cfg(unix)]
const CPU_SAMPLE_GAP: std::time::Duration = std::time::Duration::from_millis(100);

/// How long the first look reads again while the system hands out the same ticks, before the
/// reading counts as one the machine does not give.
#[cfg(unix)]
const CPU_FIRST_SAMPLE_BOUND: std::time::Duration = std::time::Duration::from_secs(2);

/// Reads the CPU free percentage turn by turn: every reading is the idle share of the ticks
/// since the last reading that differed, at least [`CPU_SAMPLE_GAP`] apart. macOS hands the
/// same counters to calls that follow each other closely, so an unchanged reading is no new
/// sample: the last percentage holds. A debug build that was given a machine reads that machine
/// instead (and takes no tick reading).
#[cfg(unix)]
struct CpuGauge {
    last: Option<(Ticks, std::time::Instant)>,
    pct: Option<u64>,
}

#[cfg(unix)]
impl CpuGauge {
    /// A gauge with its first tick reading taken now (none for an injected machine, or when the
    /// system gives none).
    fn new() -> CpuGauge {
        let last = if cpu_injected().is_some() {
            None
        } else {
            sys::cpu_ticks().map(|t| (t, std::time::Instant::now()))
        };
        CpuGauge { last, pct: None }
    }

    /// The CPU free percentage now; `None` when the machine gives no reading (never a reason
    /// to wait). The look waits out the rest of [`CPU_SAMPLE_GAP`] when the reading before it
    /// is younger. Until there is a first percentage it reads again, for at most
    /// [`CPU_FIRST_SAMPLE_BOUND`].
    fn read(&mut self) -> Option<u64> {
        if let Some((free, _)) = cpu_injected() {
            return free;
        }
        if self.last.is_none() {
            self.take(sys::cpu_ticks()?, std::time::Instant::now());
        }
        let start = std::time::Instant::now();
        loop {
            let age = self.last.map_or(CPU_SAMPLE_GAP, |(_, at)| at.elapsed());
            if age < CPU_SAMPLE_GAP {
                std::thread::sleep(CPU_SAMPLE_GAP - age);
            }
            let pct = self.take(sys::cpu_ticks()?, std::time::Instant::now());
            if pct.is_some() || start.elapsed() >= CPU_FIRST_SAMPLE_BOUND {
                return pct;
            }
            std::thread::sleep(CPU_SAMPLE_GAP);
        }
    }

    /// Adds the tick reading `now`, taken at `at`, and answers the percentage that holds after
    /// it. A reading equal to the one before it leaves the base and the percentage as they are;
    /// one that gives no usable difference keeps the percentage and becomes the base.
    fn take(&mut self, now: Ticks, at: std::time::Instant) -> Option<u64> {
        let Some((before, _)) = self.last else {
            self.last = Some((now, at));
            return self.pct;
        };
        if now != before {
            self.last = Some((now, at));
            self.pct = cpu_free_pct(before, now).or(self.pct);
        }
        self.pct
    }
}

/// Why a heavy run must not start now, or `None`. A value the system does not give is not a
/// reason to refuse (it is shown as `unknown` elsewhere). A machine with memory readings
/// (macOS) is judged on them and its swap is only shown; one without them (Linux) on swap.
pub fn over_limits(folder: &Path, limits: Limits) -> Option<String> {
    // The CPU readings are not part of these limits: low CPU is a wait ([`sys::hold`]), not a
    // refusal.
    let Machine {
        free, swap, memory, ..
    } = measured(folder);
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
/// that holds it, every worktree of the repository (real paths) and the retired home build root
/// `<HOME>/.moai/build` when the home is known, both with its links resolved (`home_build`) and
/// as it is written (`home_build_named`). Nothing is read from or made under that root.
#[derive(Debug, Clone)]
pub struct Here {
    pub cwd: PathBuf,
    pub worktree: PathBuf,
    pub worktrees: Vec<PathBuf>,
    pub home_build: Option<PathBuf>,
    pub home_build_named: Option<PathBuf>,
}

impl Here {
    /// `None` outside a git worktree: nothing to compare with. `home` and `userprofile` are the
    /// `HOME` and `USERPROFILE` values (the shared rule of `lane_env::home_dir`: USERPROFILE
    /// stands in on Windows when HOME is unset); without a usable home only the retired home
    /// build root is not recognised.
    pub fn of(
        cwd: &Path,
        home: Option<&std::ffi::OsStr>,
        userprofile: Option<&std::ffi::OsStr>,
    ) -> Option<Here> {
        let named = crate::lane_env::home_dir(home, userprofile, cfg!(windows))
            .ok()
            .map(|h| crate::hook_config::normalize(&h.join(".moai").join("build")));
        Some(Here {
            cwd: cwd.to_path_buf(),
            worktree: crate::lane_env::worktree_root(cwd).ok()?,
            worktrees: crate::lane_env::worktrees(cwd).ok()?,
            home_build: named.as_deref().map(settled),
            home_build_named: named,
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

/// #2879 D3, revised for the build rule of 2026-11-22 (a worktree builds in its own `target/`,
/// `CARGO_TARGET_DIR` unset): when the wrapped command is `cargo` and its `CARGO_TARGET_DIR`
/// names the retired per-lane `moai-rc-target-*` folder, a folder under the retired home build
/// root `<HOME>/.moai/build` (this worktree's own old folder included) or a folder inside another
/// worktree, a one-line warning for stderr. A warning, not a refusal; nothing is made anywhere.
/// `None` when the name is no lane name, `here` is unknown, the value is unset or empty, or the
/// folder is inside this worktree or in no place named above.
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
    // The value as written (`.` and `..` applied, no link followed) and with its links
    // resolved: a link inside the retired home build root may lead out of it, and a link
    // above it may lead into it, so the root is looked for in both.
    let written = here.cwd.join(target_dir.filter(|t| !t.is_empty())?);
    let named = crate::hook_config::normalize(&written);
    let t = settled(&written);
    let holder = here
        .worktrees
        .iter()
        .filter(|w| t.starts_with(w))
        .max_by_key(|w| w.components().count())
        .filter(|w| **w != here.worktree);
    let mut shown = &t;
    let why = if t
        .file_name()
        .is_some_and(|n| n.to_string_lossy().starts_with("moai-rc-target-"))
    {
        "the retired per-lane folder every worktree of the lane shared".to_string()
    } else if let Some(root) = here.home_build.as_ref().filter(|r| t.starts_with(r)) {
        format!("inside the retired home build root {}", root.display())
    } else if let Some(root) = here
        .home_build_named
        .as_ref()
        .filter(|r| named.starts_with(r))
    {
        shown = &named;
        format!("inside the retired home build root {}", root.display())
    } else if let Some(w) = holder {
        format!("inside the worktree {}", w.display())
    } else {
        return None;
    };
    Some(format!(
        "heavy-slot: CARGO_TARGET_DIR {} is {why}; this worktree builds in its own target/ ({}): unset CARGO_TARGET_DIR (`moai lane-env --lane {name}` prints that)",
        shown.display(),
        here.worktree.join("target").display()
    ))
}

/// `${XDG_CACHE_HOME:-$HOME/.cache}`, the user's cache folder.
fn cache_dir() -> Result<PathBuf> {
    Ok(
        match std::env::var_os("XDG_CACHE_HOME").filter(|v| !v.is_empty()) {
            Some(x) => PathBuf::from(x),
            None => {
                PathBuf::from(std::env::var_os("HOME").ok_or("HOME is not set")?).join(".cache")
            }
        },
    )
}

/// `${XDG_CACHE_HOME:-$HOME/.cache}/moai/heavy-slots`, the folder the removed Python tool used.
pub fn default_dir() -> Result<PathBuf> {
    Ok(cache_dir()?.join("moai").join("heavy-slots"))
}

/// The folder of the builds from before the name change, `<cache>/moai-cli/heavy-slots`, when
/// `folder` is the default folder: only then does the earlier folder matter. It is read, never
/// written, moved or made (SPEC-20261121-moai-name D-10, D-11).
pub fn earlier_dir(folder: &Path) -> Option<PathBuf> {
    if default_dir().ok().as_deref() != Some(folder) {
        return None;
    }
    Some(cache_dir().ok()?.join("moai-cli").join("heavy-slots"))
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
pub use sys::{acquire, hold, hold_ordered, status, Held};

#[cfg(not(unix))]
pub fn status(_folder: &Path) -> Result<Value> {
    Err("heavy-slot needs flock, which this platform does not provide".into())
}

/// A held slot; never made on this platform.
#[cfg(not(unix))]
pub struct Held {
    _never: (),
}

#[cfg(not(unix))]
impl Held {
    pub fn slot(&self) -> usize {
        0
    }
    pub fn inherited(&self) -> bool {
        false
    }
    pub fn cpu_waited_secs(&self) -> Option<u64> {
        None
    }
    pub fn env_value(&self) -> std::ffi::OsString {
        std::ffi::OsString::new()
    }
    pub fn taken_line(&self, _name: &str) -> String {
        String::new()
    }
}

#[cfg(not(unix))]
pub fn hold(
    _folder: &Path,
    _max: usize,
    _name: &str,
    _command: &[String],
    _limits: Limits,
    _progress: &mut dyn std::io::Write,
) -> std::result::Result<Held, u8> {
    eprintln!("heavy-slot: flock is unavailable on this platform");
    Err(USAGE)
}

#[cfg(not(unix))]
pub fn hold_ordered(
    _folder: &Path,
    _max: usize,
    _name: &str,
    _command: &[String],
    _limits: Limits,
    _progress: &mut dyn std::io::Write,
    _front: bool,
) -> std::result::Result<Held, u8> {
    eprintln!("heavy-slot: flock is unavailable on this platform");
    Err(USAGE)
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
        cpu_cores, cpu_wait_decision, logical_cpus, low_space_line, machine_line, measured,
        over_limits, CpuGauge, CpuStep, Limits, Path, PathBuf, Result, Value, BUSY, CPU_WAIT_CAP,
        DEFAULT_LIMITS, REFUSED, USAGE,
    };
    use serde_json::json;
    use std::fs::{self, File, OpenOptions, TryLockError};
    use std::io::{Seek, SeekFrom, Write};
    use std::os::unix::ffi::OsStrExt;
    use std::os::unix::fs::{MetadataExt, OpenOptionsExt};
    use std::os::unix::io::AsRawFd;
    use std::os::unix::process::{CommandExt, ExitStatusExt};
    use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};

    const POLL: Duration = Duration::from_millis(200);
    /// How often the line "waiting for CPU headroom" is printed again while the wait lasts.
    const CPU_SAY_EVERY: Duration = Duration::from_secs(30);
    const FORWARDED: [libc::c_int; 3] = [libc::SIGINT, libc::SIGTERM, libc::SIGHUP];

    pub fn load() -> Option<[f64; 3]> {
        let mut l = [0f64; 3];
        // SAFETY: getloadavg writes at most `nelem` doubles into the buffer it is given.
        let n = unsafe { libc::getloadavg(l.as_mut_ptr(), 3) };
        (n == 3).then_some(l)
    }

    /// The `host_port` of this process: `mach_host_self` is deprecated in the locked `libc`
    /// (its note points to the `mach2` crate, which this crate does not depend on), and it is
    /// the call that names the host for `host_statistics`. It is made once; every call adds a
    /// reference to the port.
    #[cfg(target_os = "macos")]
    fn host_port() -> libc::mach_port_t {
        static HOST: std::sync::OnceLock<libc::mach_port_t> = std::sync::OnceLock::new();
        *HOST.get_or_init(|| {
            // SAFETY: mach_host_self takes no argument and returns the host port name.
            #[allow(deprecated)]
            unsafe {
                libc::mach_host_self()
            }
        })
    }

    /// The CPU ticks since boot as `(idle, total)`, or `None` when the system gives none. macOS:
    /// `host_statistics` with `HOST_CPU_LOAD_INFO`, four 32-bit counters (user, system, idle,
    /// nice) that wrap, so the total wraps too. One reading is only an average since boot; the
    /// free percentage is made of two ([`super::CpuGauge`]).
    #[cfg(target_os = "macos")]
    pub fn cpu_ticks() -> Option<super::Ticks> {
        // SAFETY: host_statistics fills at most `count` integers of the zeroed struct, and
        // `count` is the size the flavour asks for.
        let mut info: libc::host_cpu_load_info = unsafe { std::mem::zeroed() };
        let mut count: libc::mach_msg_type_number_t = libc::HOST_CPU_LOAD_INFO_COUNT;
        let rc = unsafe {
            libc::host_statistics(
                host_port(),
                libc::HOST_CPU_LOAD_INFO,
                (&mut info as *mut libc::host_cpu_load_info).cast(),
                &mut count,
            )
        };
        if rc != 0 {
            return None;
        }
        let idle = info.cpu_ticks[libc::CPU_STATE_IDLE as usize];
        let total = info.cpu_ticks.iter().fold(0u32, |a, t| a.wrapping_add(*t));
        Some((idle, total))
    }

    /// Linux: the first `cpu ` line of `/proc/stat`.
    #[cfg(target_os = "linux")]
    pub fn cpu_ticks() -> Option<super::Ticks> {
        super::parse_proc_stat(&fs::read_to_string("/proc/stat").ok()?)
    }

    /// Other Unix systems give no CPU reading, so the CPU rule never makes a caller wait.
    #[cfg(not(any(target_os = "macos", target_os = "linux")))]
    pub fn cpu_ticks() -> Option<super::Ticks> {
        None
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
        /// The waiting path is not a real folder of the slot folder (a symlink, a file).
        NotFolder(PathBuf),
        Io(String),
    }

    /// The `waiting` folder, opened once without following a symlink. Every look at it and every
    /// removal in it goes through this descriptor (`openat`, `unlinkat`), so a link put in the
    /// place of the folder afterwards cannot lead any of them to another folder.
    struct WaitDir {
        dir: File,
        path: PathBuf,
    }

    impl WaitDir {
        /// The folder at `path`, which must be a real folder; with `create` a missing one is made
        /// (mode 0700). A symlink — to a folder or nowhere — or a file there is
        /// [`Open::NotFolder`]; nothing behind a link is created.
        fn open(path: &Path, create: bool) -> std::result::Result<WaitDir, Open> {
            if create {
                use std::os::unix::fs::DirBuilderExt;
                if let Err(e) = fs::DirBuilder::new().mode(0o700).create(path) {
                    if e.kind() != std::io::ErrorKind::AlreadyExists {
                        return Err(Open::Io(format!(
                            "creating {} failed ({e})",
                            path.display()
                        )));
                    }
                }
            }
            let dir = OpenOptions::new()
                .read(true)
                .custom_flags(libc::O_DIRECTORY | libc::O_NOFOLLOW | libc::O_CLOEXEC)
                .open(path)
                .map_err(|e| match fs::symlink_metadata(path) {
                    Ok(m) if !m.is_dir() => Open::NotFolder(path.into()),
                    _ => Open::Io(format!("{}: {e}", path.display())),
                })?;
            Ok(WaitDir {
                dir,
                path: path.into(),
            })
        }

        fn display(&self, name: &str) -> PathBuf {
            self.path.join(name)
        }

        /// The names in the folder, read through the descriptor.
        fn names(&self) -> Vec<String> {
            // SAFETY: fdopendir takes ownership of the duplicate; closedir releases it. The
            // duplicate shares the offset of the original, so the stream is rewound first.
            unsafe {
                let fd = libc::dup(self.dir.as_raw_fd());
                if fd < 0 {
                    return Vec::new();
                }
                let d = libc::fdopendir(fd);
                if d.is_null() {
                    libc::close(fd);
                    return Vec::new();
                }
                libc::rewinddir(d);
                let mut out = Vec::new();
                loop {
                    let e = libc::readdir(d);
                    if e.is_null() {
                        break;
                    }
                    let n = std::ffi::CStr::from_ptr((*e).d_name.as_ptr());
                    let n = n.to_string_lossy().into_owned();
                    if n != "." && n != ".." {
                        out.push(n);
                    }
                }
                libc::closedir(d);
                out
            }
        }

        /// `name` as the name of one entry of this folder: no slash, no NUL, not `.` or `..`.
        fn entry(&self, name: &str) -> Option<std::ffi::CString> {
            if name.is_empty() || name == "." || name == ".." || name.contains('/') {
                return None;
            }
            std::ffi::CString::new(name).ok()
        }

        /// Whether the entry `name` is gone from this folder (a missing name, not any other
        /// failure), looked up through the descriptor without following a link.
        fn is_missing(&self, name: &str) -> bool {
            let Some(c) = self.entry(name) else {
                return false;
            };
            // SAFETY: a zeroed stat the call fills in, on a descriptor this struct owns.
            let r = unsafe {
                let mut st: libc::stat = std::mem::zeroed();
                libc::fstatat(
                    self.dir.as_raw_fd(),
                    c.as_ptr(),
                    &mut st,
                    libc::AT_SYMLINK_NOFOLLOW,
                )
            };
            r != 0 && std::io::Error::last_os_error().kind() == std::io::ErrorKind::NotFound
        }

        /// Removes the entry `name` of this folder. `unlinkat` without `AT_REMOVEDIR` removes a
        /// symlink itself and never what it points to, and never a folder.
        fn remove(&self, name: &str) -> std::io::Result<()> {
            let c = self
                .entry(name)
                .ok_or_else(|| std::io::Error::from(std::io::ErrorKind::InvalidInput))?;
            // SAFETY: unlinkat on a descriptor this struct owns and a NUL-terminated name.
            let r = unsafe { libc::unlinkat(self.dir.as_raw_fd(), c.as_ptr(), 0) };
            if r == 0 {
                Ok(())
            } else {
                Err(std::io::Error::last_os_error())
            }
        }
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
        plain_checked(file, path)
    }

    /// [`open_plain`] for the entry `name` of the waiting folder, through its descriptor: the
    /// name is looked up in the folder that was opened, never along a path.
    fn open_plain_in(
        w: &WaitDir,
        name: &str,
        create: bool,
        exclusive: bool,
    ) -> std::result::Result<File, Open> {
        let path = w.display(name);
        let Some(c) = w.entry(name) else {
            return Err(Open::Io(format!(
                "{}: not a name of one entry",
                path.display()
            )));
        };
        let mut flags = libc::O_RDWR | libc::O_NOFOLLOW | libc::O_NONBLOCK | libc::O_CLOEXEC;
        if exclusive {
            flags |= libc::O_CREAT | libc::O_EXCL;
        } else if create {
            flags |= libc::O_CREAT;
        }
        // SAFETY: openat on the descriptor `w` owns and a NUL-terminated name.
        let fd =
            unsafe { libc::openat(w.dir.as_raw_fd(), c.as_ptr(), flags, 0o600 as libc::c_uint) };
        if fd < 0 {
            let e = std::io::Error::last_os_error();
            // SAFETY: a zeroed stat the call fills in; a link is not followed.
            let not_regular = unsafe {
                let mut st: libc::stat = std::mem::zeroed();
                libc::fstatat(
                    w.dir.as_raw_fd(),
                    c.as_ptr(),
                    &mut st,
                    libc::AT_SYMLINK_NOFOLLOW,
                ) == 0
                    && st.st_mode & libc::S_IFMT != libc::S_IFREG
            };
            return Err(if not_regular {
                Open::NotPlain(path)
            } else {
                Open::Io(format!("{}: {e}", path.display()))
            });
        }
        // SAFETY: a descriptor just returned by openat and owned by nobody else.
        let file = unsafe { std::os::unix::io::FromRawFd::from_raw_fd(fd) };
        plain_checked(file, &path)
    }

    /// The opened `file` when it is a regular file with one link, switched to blocking mode.
    fn plain_checked(file: File, path: &Path) -> std::result::Result<File, Open> {
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
        held_by_other(open_plain(path, false, false)?, path)
    }

    /// [`is_held`] for the entry `name` of the waiting folder.
    fn is_held_in(w: &WaitDir, name: &str) -> std::result::Result<bool, Open> {
        held_by_other(open_plain_in(w, name, false, false)?, &w.display(name))
    }

    fn held_by_other(f: File, path: &Path) -> std::result::Result<bool, Open> {
        match f.try_lock_shared() {
            Ok(()) => Ok(false),
            Err(TryLockError::WouldBlock) => Ok(true),
            Err(TryLockError::Error(e)) => Err(Open::Io(format!("{}: {e}", path.display()))),
        }
    }

    /// The first of `slots` this caller can lock, with its index.
    fn take_one(
        folder: &Path,
        slots: &[usize],
    ) -> std::result::Result<Option<(usize, File)>, Open> {
        for &i in slots {
            if let Some(f) = try_exclusive(&folder.join(format!("slot-{i}")), false)? {
                return Ok(Some((i, f)));
            }
        }
        Ok(None)
    }

    /// The slots of `0..max` nobody holds now, lowest first: a slot file that does not exist yet
    /// counts as free. Reads only (a shared try-lock that is let go at once).
    fn free_slots(folder: &Path, max: usize) -> std::result::Result<Vec<usize>, Open> {
        let mut free = Vec::new();
        for i in 0..max {
            let path = folder.join(format!("slot-{i}"));
            match fs::symlink_metadata(&path) {
                Err(e) if e.kind() == std::io::ErrorKind::NotFound => free.push(i),
                _ => {
                    if !is_held(&path)? {
                        free.push(i);
                    }
                }
            }
        }
        Ok(free)
    }

    /// The live waiters (their waiting file is locked) and `me`, oldest first by name. An
    /// unlocked file is a waiter that died: it is not counted (REQ-4). Read once per look.
    fn live_waiters(waiting: &WaitDir, me: &str) -> Vec<String> {
        let mut names = waiting.names();
        names.sort();
        names.retain(|n| n.as_str() == me || matches!(is_held_in(waiting, n), Ok(true)));
        names
    }

    /// The names of the waiting files of callers that are gone: a plain file (opened without
    /// following a link, with one link) named like a waiter's own file, whose pid is not alive
    /// and which no process holds locked. `me` (the caller's own file) is never one. A pid that
    /// cannot be judged, a name of any other shape and a file that cannot be opened as a plain
    /// file are left out, so a mistake keeps a file rather than removing a live waiter's
    /// (SPEC-20261123-heavy-slot-leftover REQ-3). Every look goes through the folder's
    /// descriptor, never along its path.
    fn stale_waiters(waiting: &WaitDir, me: &str) -> Vec<String> {
        let mut found: Vec<String> = waiting
            .names()
            .into_iter()
            .filter(|name| {
                name != me
                    && super::waiter_pid(name)
                        .is_some_and(|pid| crate::home_state::pid_alive(pid) == Some(false))
                    && matches!(is_held_in(waiting, name), Ok(false))
            })
            .collect();
        found.sort();
        found
    }

    /// 1-based place of `me` among `live` (oldest first by name), and their count.
    fn queue_place(live: &[String], me: &str) -> (usize, usize) {
        let place = live
            .iter()
            .position(|n| n.as_str() == me)
            .map_or(1, |p| p + 1);
        (place, live.len().max(1))
    }

    /// The `--max` a waiting file names last (`<ns>-<pid>-<max>`); `None` for a name without one
    /// (the Python twin, an older build), a waiter that can use any slot.
    pub(super) fn waiter_max(name: &str) -> Option<usize> {
        let mut parts = name.split('-');
        let (_ns, _pid, max) = (parts.next()?, parts.next()?, parts.next()?);
        if parts.next().is_some() || max.is_empty() || !max.bytes().all(|b| b.is_ascii_digit()) {
            return None;
        }
        max.parse().ok()
    }

    /// The slots `me` may take now, lowest first: its `free` slots that remain once every live
    /// waiter of `live` that arrived before it has been given, in arrival order, the lowest
    /// remaining slot it can use (below its own `--max`) — the slot it takes itself. Empty: `me`
    /// waits (REQ-4).
    pub(super) fn my_slots(mut free: Vec<usize>, live: &[String], me: &str) -> Vec<usize> {
        for w in live.iter().filter(|n| n.as_str() < me) {
            if free.is_empty() {
                break;
            }
            let can = waiter_max(w).unwrap_or(usize::MAX);
            if let Some(i) = free.iter().position(|&s| s < can) {
                free.remove(i);
            }
        }
        free
    }

    /// A slot held by this process or inherited from the one that started it. Dropping it lets
    /// this process's hold go; the operating system frees the slot when no process has the file
    /// open any more.
    pub struct Held {
        slot: usize,
        path: PathBuf,
        /// `None` for an inherited slot: the descriptor belongs to the process that took it.
        file: Option<File>,
        fd: libc::c_int,
        /// The CPU free percentage read when this caller took the slot and the limit it was
        /// held against; `None` when the machine gave no reading or the slot is inherited.
        cpu_seen: Option<(u64, u64)>,
        /// The wait for CPU headroom this caller went through; `None` when it did not wait.
        cpu_wait: Option<CpuWaited>,
    }

    /// A wait for CPU headroom that happened: the whole seconds spent held back (rounded up, at
    /// least 1) and whether the cap of [`CPU_WAIT_CAP`] ended it.
    struct CpuWaited {
        secs: u64,
        capped: bool,
    }

    impl Held {
        pub fn slot(&self) -> usize {
            self.slot
        }

        /// The seconds this caller waited for CPU headroom before it took the slot; `None` when
        /// it did not wait.
        pub fn cpu_waited_secs(&self) -> Option<u64> {
            self.cpu_wait.as_ref().map(|w| w.secs)
        }

        /// Whether the slot came from the process that started this one (REQ-5).
        pub fn inherited(&self) -> bool {
            self.file.is_none()
        }

        /// The line a caller prints once it has the slot: `heavy-slot: slot <i> taken by <name>`,
        /// or `heavy-slot: <name> reuses slot <i> it inherited`. A CPU reading follows as
        /// `; cpu free <n>% (limit <m>%)`, or `(rule off)` for a limit of 0. A caller that
        /// waited for CPU headroom ends the line with `; waited <n>s for CPU headroom`, with
        /// `; cpu wait ended by the 300 s cap after <n>s` before that when the cap ended it.
        pub fn taken_line(&self, name: &str) -> String {
            let mut line = if self.inherited() {
                format!("heavy-slot: {name} reuses slot {} it inherited", self.slot)
            } else {
                format!("heavy-slot: slot {} taken by {name}", self.slot)
            };
            if let Some((free, limit)) = self.cpu_seen {
                let rule = match limit {
                    0 => "rule off".to_string(),
                    l => format!("limit {l}%"),
                };
                line.push_str(&format!("; cpu free {free}% ({rule})"));
            }
            if let Some(w) = &self.cpu_wait {
                if w.capped {
                    line.push_str(&format!(
                        "; cpu wait ended by the {} s cap after {}s",
                        CPU_WAIT_CAP.as_secs(),
                        w.secs
                    ));
                }
                line.push_str(&format!("; waited {}s for CPU headroom", w.secs));
            }
            line
        }

        /// The [`super::SLOT_FD_ENV`] value for a child: `<fd>:<slot path>`.
        pub fn env_value(&self) -> std::ffi::OsString {
            let mut v = std::ffi::OsString::from(format!("{}:", self.fd));
            v.push(self.path.as_os_str());
            v
        }

        /// Overwrites the slot record that `status` reads; an inherited slot keeps its owner's.
        fn write_record(&mut self, name: &str, command: &[String]) {
            let Some(file) = self.file.as_mut() else {
                return;
            };
            let record = json!({
                "name": name,
                "command": command,
                "pid": std::process::id(),
                "started_unix": SystemTime::now().duration_since(UNIX_EPOCH).map_or(0, |d| d.as_secs()),
                "slot": self.slot,
            });
            let _ = file.set_len(0);
            let _ = file.seek(SeekFrom::Start(0));
            let _ = writeln!(file, "{record}");
        }
    }

    /// A holder that took the slot gives it back when it ends: `flock(LOCK_UN)` on its own open
    /// file description unlocks it for every descriptor that shares that description, so a
    /// descendant that outlives the command and still has the descriptor open (a test child, a
    /// background process) no longer keeps the slot. A holder killed by SIGKILL never gets here;
    /// the slot then stays held until every process with the file is gone. An inherited slot
    /// (`file` is `None`) belongs to the outer holder and is never unlocked here.
    impl Drop for Held {
        fn drop(&mut self) {
            if self.file.is_some() {
                // SAFETY: flock(2) on the descriptor this value owns; it is closed after this.
                unsafe { libc::flock(self.fd, libc::LOCK_UN) };
            }
        }
    }

    /// The slot named by [`super::SLOT_FD_ENV`], reused only when all of these hold; anything
    /// else is `None`, and the caller queues as before:
    /// 1. the path names `slot-<digits>` in `folder` itself (the same folder by device and inode,
    ///    so another spelling of its path counts) and that slot file is a plain file;
    /// 2. the descriptor is open on that very file (same device and inode);
    /// 3. someone holds the slot now: a shared try-lock on a fresh open of it fails;
    /// 4. an exclusive non-blocking `flock` on the descriptor succeeds. A `flock` belongs to the
    ///    open file description, so with the slot held (3) this succeeds only when the descriptor
    ///    shares the holder's description — the one the process that took the slot handed down —
    ///    and fails for the slot file opened anew, by this process or any other.
    ///
    /// Step 3 comes first so the check never takes a slot that nobody held. Should the holder let
    /// the slot go between 3 and 4, step 4 takes it on the descriptor's own description, and the
    /// slot still has one holder.
    fn inherited(folder: &Path) -> Option<Held> {
        let value = std::env::var_os(super::SLOT_FD_ENV)?;
        let bytes = value.as_bytes();
        let colon = bytes.iter().position(|&b| b == b':')?;
        let fd: libc::c_int = std::str::from_utf8(&bytes[..colon]).ok()?.parse().ok()?;
        if fd < 0 {
            return None;
        }
        let named = Path::new(std::ffi::OsStr::from_bytes(&bytes[colon + 1..]));
        let file_name = named.file_name()?.to_str()?;
        let digits = file_name.strip_prefix("slot-")?;
        if digits.is_empty() || !digits.bytes().all(|b| b.is_ascii_digit()) {
            return None;
        }
        let slot: usize = digits.parse().ok()?;
        let (theirs, ours) = (
            fs::metadata(named.parent()?).ok()?,
            fs::metadata(folder).ok()?,
        );
        if (theirs.dev(), theirs.ino()) != (ours.dev(), ours.ino()) {
            return None;
        }
        let here = folder.join(file_name);
        let m = fs::symlink_metadata(&here).ok()?;
        if !m.file_type().is_file() {
            return None;
        }
        // SAFETY: fstat fills the zeroed struct for a descriptor number; a closed one fails.
        let mut st: libc::stat = unsafe { std::mem::zeroed() };
        if unsafe { libc::fstat(fd, &mut st) } != 0 {
            return None;
        }
        #[allow(clippy::unnecessary_cast)]
        let same = (st.st_dev as u64, st.st_ino as u64) == (m.dev(), m.ino());
        if !same || !matches!(is_held(&here), Ok(true)) {
            return None;
        }
        // SAFETY: flock(2) on a descriptor number; a closed one fails with EBADF.
        if unsafe { libc::flock(fd, libc::LOCK_EX | libc::LOCK_NB) } != 0 {
            return None;
        }
        Some(Held {
            slot,
            path: here,
            file: None,
            fd,
            cpu_seen: None,
            cpu_wait: None,
        })
    }

    /// The descriptor stays open in the children this process starts.
    fn keep_for_children(fd: libc::c_int) {
        // SAFETY: fcntl on a descriptor the caller owns.
        unsafe {
            let fl = libc::fcntl(fd, libc::F_GETFD);
            if fl >= 0 {
                libc::fcntl(fd, libc::F_SETFD, fl & !libc::FD_CLOEXEC);
            }
        }
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

    fn refused_folder(p: &Path) -> u8 {
        eprintln!(
            "heavy-slot: refused {}: the waiting folder is not a real folder (a symlink or a file); nothing was written or removed and COMMAND did not run — remove it or pass another --dir",
            p.display()
        );
        REFUSED
    }

    fn io(e: &str) -> u8 {
        eprintln!("heavy-slot: {e}; COMMAND did not run");
        REFUSED
    }

    /// How many times a caller tries the lock of its own fresh waiting file, and the pause
    /// between tries: another waiter's look (a shared try-lock let go at once) can hold it for a
    /// moment, which is no reason to give up.
    const OWN_LOCK_TRIES: u32 = 50;
    const OWN_LOCK_PAUSE: Duration = Duration::from_millis(10);

    /// The exclusive lock of this caller's own fresh waiting file, tried up to `tries` times
    /// `pause` apart while another process holds it for a moment; still held after that, it is
    /// an error.
    pub(super) fn lock_own(
        file: &File,
        path: &Path,
        tries: u32,
        pause: Duration,
    ) -> std::result::Result<(), String> {
        for i in 0..tries {
            match file.try_lock() {
                Ok(()) => return Ok(()),
                Err(TryLockError::WouldBlock) => {
                    if i + 1 < tries {
                        std::thread::sleep(pause);
                    }
                }
                Err(TryLockError::Error(e)) => return Err(format!("{}: {e}", path.display())),
            }
        }
        Err(format!("{}: held by another process", path.display()))
    }

    /// Waits for a slot and returns it held, with its close-on-exec flag cleared so the
    /// children of this process keep it, and its record naming `name` and `command`. A slot
    /// inherited in [`super::SLOT_FD_ENV`] from the process that started this one is reused
    /// without queueing (REQ-5). Otherwise the caller queues in arrival order (REQ-4, the module
    /// doc); the waiting lines go to `progress`, and the caller announces the slot it got
    /// ([`Held::taken_line`]). A machine over `limits` gets [`BUSY`]; a waiter whose parent
    /// ended leaves the queue with [`REFUSED`] — nobody is left to take its result (#2677 items
    /// 3, 4). A caller that could take a slot while the CPU free percentage is below
    /// `limits.min_cpu_free_pct` keeps its place in the queue and waits (the module doc's CPU
    /// rule), at most [`CPU_WAIT_CAP`].
    pub fn hold(
        folder: &Path,
        max: usize,
        name: &str,
        command: &[String],
        limits: Limits,
        progress: &mut dyn Write,
    ) -> std::result::Result<Held, u8> {
        hold_ordered(folder, max, name, command, limits, progress, false)
    }

    /// [`hold`] with the place the caller queues at: `front` writes the arrival stamp of the
    /// waiting file in microseconds ([`super::waiting_stamp`]), so the file sorts ahead of every
    /// waiting file stamped in nanoseconds and the caller starts before them
    /// (SPEC-20261126-leader-slot-front REQ-1 to REQ-3). Callers whose `front` is true keep their
    /// arrival order among themselves. Nothing else differs: the slot count, the limits, the
    /// refusals and the CPU rule are the same for both places (REQ-6).
    pub fn hold_ordered(
        folder: &Path,
        max: usize,
        name: &str,
        command: &[String],
        limits: Limits,
        progress: &mut dyn Write,
        front: bool,
    ) -> std::result::Result<Held, u8> {
        if max < 1 {
            eprintln!("heavy-slot: acquire needs --max >= 1 and a command after --");
            return Err(USAGE);
        }
        if let Some(held) = inherited(folder) {
            return Ok(held);
        }
        // The first CPU reading of the CPU rule; a limit of 0 switches the rule off and takes none.
        let mut gauge = (limits.min_cpu_free_pct > 0).then(CpuGauge::new);
        let cwd = std::env::current_dir().unwrap_or_else(|_| folder.to_path_buf());
        if let Some(why) = over_limits(&cwd, limits) {
            return Err(busy(name, &why, &cwd));
        }
        let ppid = parent();
        // The slot folder may be new; `waiting` inside it must be a real folder, opened once
        // without following a link (SPEC-20261123-heavy-slot-leftover REQ-5): every look at it
        // and every removal in it goes through that descriptor.
        if let Err(e) = fs::create_dir_all(folder) {
            return Err(io(&format!("creating {} failed ({e})", folder.display())));
        }
        let waiting = match WaitDir::open(&folder.join("waiting"), true) {
            Ok(w) => w,
            Err(Open::Io(e)) => return Err(io(&e)),
            Err(Open::NotFolder(p)) => return Err(refused_folder(&p)),
            Err(Open::NotPlain(p)) => return Err(refused(&p)),
        };
        // Every caller queues first, so one that arrived earlier is never passed (REQ-4); the
        // name ends with this caller's `--max`, so the others know which slots it can use.
        let me = format!(
            "{:020}-{}-{max}",
            super::waiting_stamp(front, now_ns()),
            std::process::id()
        );
        let wait_path = waiting.display(&me);
        let queue = || -> std::result::Result<File, Open> {
            let f = open_plain_in(&waiting, &me, true, true)?;
            lock_own(&f, &wait_path, OWN_LOCK_TRIES, OWN_LOCK_PAUSE).map_err(Open::Io)?;
            Ok(f)
        };
        let mut wait_file = match queue() {
            Err(Open::NotPlain(p) | Open::NotFolder(p)) => return Err(refused(&p)),
            Err(Open::Io(e)) => return Err(io(&e)),
            Ok(f) => f,
        };
        // The files of callers that died while waiting are removed once per call; this caller's
        // own file is locked and never among them.
        for stale in stale_waiters(&waiting, &me) {
            let _ = waiting.remove(&stale);
        }
        let mut last = None;
        let mut waited = false;
        // The CPU rule: the time spent held back by it (closed intervals, and the one that is
        // open since `low_since`), when its line was last printed, and whether the cap ended it.
        let mut cpu_waited = Duration::ZERO;
        let mut low_since: Option<Instant> = None;
        let mut cpu_said: Option<Instant> = None;
        let mut cpu_capped = false;
        // The reading of the latest look, for the slot line.
        let mut cpu_free;
        let taken = loop {
            // Read at every look, so a reading is always the last pause's and not an older one;
            // before the queue is read, since the first reading may wait out the sample gap.
            cpu_free = gauge.as_mut().and_then(CpuGauge::read);
            // One read of the queue per look serves both the turn and the place.
            let live = live_waiters(&waiting, &me);
            let mut cpu_hold = false;
            match free_slots(folder, max).map(|free| my_slots(free, &live, &me)) {
                Err(e) => break Err(e),
                Ok(mine) => {
                    // The rule only holds back a caller that could take a slot now: time spent
                    // behind other callers is the queue's wait, not the CPU's.
                    let low = cpu_free.filter(|&f| !mine.is_empty() && f < limits.min_cpu_free_pct);
                    let now = Instant::now();
                    match low {
                        Some(f) => {
                            let since = *low_since.get_or_insert(now);
                            match cpu_wait_decision(true, cpu_waited + now.duration_since(since)) {
                                CpuStep::Wait => {
                                    cpu_hold = true;
                                    if cpu_said
                                        .is_none_or(|t| now.duration_since(t) >= CPU_SAY_EVERY)
                                    {
                                        let _ = writeln!(
                                            progress,
                                            "heavy-slot: waiting for CPU headroom; cpu free {f}% is below {}% ({} cores)",
                                            limits.min_cpu_free_pct,
                                            cpu_cores()
                                        );
                                        cpu_said = Some(now);
                                    }
                                }
                                CpuStep::Run => cpu_capped = true,
                            }
                        }
                        None => {
                            if let Some(since) = low_since.take() {
                                cpu_waited += now.duration_since(since);
                            }
                            cpu_said = None;
                        }
                    }
                    if !cpu_hold {
                        match take_one(folder, &mine) {
                            Ok(Some(t)) => break Ok(Some(t)),
                            Ok(None) => {}
                            Err(e) => break Err(e),
                        }
                    }
                }
            }
            if !cpu_hold {
                let place = queue_place(&live, &me);
                if last != Some(place) {
                    let _ = writeln!(
                        progress,
                        "heavy-slot: waiting {} of {} for {max} slot(s); {}",
                        place.0,
                        place.1,
                        machine_line(&cwd)
                    );
                    last = Some(place);
                }
            }
            waited = true;
            std::thread::sleep(POLL);
            if parent() != ppid {
                break Ok(None);
            }
            // A waiting file that vanished is made again under the same name: the place
            // stays the one of the arrival.
            if waiting.is_missing(&me) {
                // The assignment lets the old, unlinked file go.
                match queue() {
                    Ok(f) => wait_file = f,
                    Err(e) => break Err(e),
                }
            }
        };
        drop(wait_file);
        let _ = waiting.remove(&me);
        let (slot, file) = match taken {
            Ok(Some(t)) => {
                // The machine may have changed during the wait: judge it again before
                // starting; dropping the slot file frees the slot.
                if waited {
                    if let Some(why) = over_limits(&cwd, limits) {
                        return Err(busy(name, &why, &cwd));
                    }
                }
                t
            }
            Ok(None) => {
                eprintln!(
                    "heavy-slot: {name} left the queue: the process that started it (pid {ppid}) has ended; COMMAND did not run"
                );
                return Err(REFUSED);
            }
            Err(Open::NotPlain(p) | Open::NotFolder(p)) => return Err(refused(&p)),
            Err(Open::Io(e)) => return Err(io(&e)),
        };
        // #2930 item 6: the run starts, but the disk is getting short — stderr only, so
        // stdout and the exit code stay COMMAND's own.
        if let Some(line) = low_space_line(measured(&cwd).free, limits.min_free_gb, &cwd) {
            eprintln!("heavy-slot: {line}");
        }
        let fd = file.as_raw_fd();
        keep_for_children(fd);
        // An absolute path, so a nested caller in another working folder finds the same folder.
        let path = std::env::current_dir()
            .map_or_else(|_| folder.to_path_buf(), |c| c.join(folder))
            .join(format!("slot-{slot}"));
        // The interval the loop left open (the cap ended it) counts too; whole seconds, rounded
        // up, so a caller that was held back at all says at least 1.
        if let Some(since) = low_since.take() {
            cpu_waited += since.elapsed();
        }
        let cpu_wait = (!cpu_waited.is_zero()).then(|| CpuWaited {
            secs: cpu_waited.as_secs() + u64::from(cpu_waited.subsec_nanos() > 0),
            capped: cpu_capped,
        });
        // With the rule off no look read the CPU: one reading now, for the slot line.
        if gauge.is_none() {
            cpu_free = CpuGauge::new().read();
        }
        let mut held = Held {
            slot,
            path,
            file: Some(file),
            fd,
            cpu_seen: cpu_free.map(|f| (f, limits.min_cpu_free_pct)),
            cpu_wait,
        };
        held.write_record(name, command);
        Ok(held)
    }

    /// Waits for a slot ([`hold`], waiting lines on stdout), runs `command` holding it, returns
    /// its exit code or the refusal of [`hold`]. The line that names the slot taken and the exit
    /// line go to stderr, so stdout carries nothing of the wrapper once COMMAND runs.
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
        let held = match hold(folder, max, name, command, limits, &mut std::io::stdout()) {
            Ok(h) => h,
            Err(code) => return code,
        };
        eprintln!("{}", held.taken_line(name));
        let cwd = std::env::current_dir().unwrap_or_else(|_| folder.to_path_buf());
        run_holding(held, name, command, &cwd, max)
    }

    fn run_holding(held: Held, name: &str, command: &[String], cwd: &Path, max: usize) -> u8 {
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
        // The command inherits the locked file ([`hold`] cleared its close-on-exec flag) and
        // learns where it is, so a nested caller in the same folder reuses the slot (REQ-5).
        let mut cmd = std::process::Command::new(&command[0]);
        cmd.args(&command[1..])
            .env(super::SLOT_FD_ENV, held.env_value());
        // A value the caller (or an outer holder) already set stays; otherwise every command
        // under a slot gets the cap, so the slots do not each spawn one rustc per CPU.
        // An empty value is no value: cargo cannot read it, so it gets the cap too.
        if std::env::var_os("CARGO_BUILD_JOBS").is_none_or(|v| v.is_empty()) {
            cmd.env("CARGO_BUILD_JOBS", super::build_jobs_cap(max).to_string());
        }
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
        let cpu_waited = held.cpu_waited_secs();
        drop(held);
        // Every started command ends with one exit line, 0 included, so a shell waiting for
        // that line ends on success too (SPEC-20261122-heavy-slot-exit-line). A caller that
        // waited for CPU headroom ends it with the seconds it waited.
        let waited = cpu_waited
            .map(|s| format!("; waited {s}s for CPU headroom"))
            .unwrap_or_default();
        eprintln!(
            "heavy-slot: {name} exited {code}; {}{waited}",
            machine_line(cwd)
        );
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
            let open_by = match find(slot) {
                Some((found, source)) => {
                    rec["holders"] = json!(found
                        .iter()
                        .map(|h| json!({"pid": h.pid, "comm": h.comm}))
                        .collect::<Vec<_>>());
                    rec["holders_source"] = json!(source);
                    if found.is_empty() {
                        "no process found with the file open".to_string()
                    } else {
                        found
                            .iter()
                            .map(|h| format!("{} {}", h.pid, h.comm))
                            .collect::<Vec<_>>()
                            .join(", ")
                    }
                }
                None => {
                    rec["holders"] = Value::Null;
                    rec["holders_source"] = json!("unavailable");
                    "holders unavailable".to_string()
                }
            };
            // One line a person reads: the slot, the recorded pid that is gone, and who has the
            // slot file open now (pid and name only, never argv).
            let number = slot
                .file_name()
                .and_then(|n| n.to_str())
                .and_then(|n| n.strip_prefix("slot-"))
                .unwrap_or("?");
            let pid = rec.get("pid").and_then(Value::as_i64).unwrap_or(0);
            rec["warning"] = json!(format!(
                "slot {number} is held by a process other than its recorded holder (pid {pid} is not running); open by: {open_by}"
            ));
        }
        rec
    }

    /// The slots held now (their records, each with `alive`) and slot paths that are not plain
    /// files, with the CPU readings of the CPU rule (`cpu_free_pct`, `null` when the machine
    /// gives none, `cpu_cores`, `min_cpu_free_pct`) and `heavy_runs`, the number of held slots.
    /// Reads only; the CPU free percentage is taken over two readings 100 ms apart, once.
    pub fn status(folder: &Path) -> Result<Value> {
        let free = CpuGauge::new().read();
        status_with(folder, free, cpu_cores())
    }

    fn status_with(folder: &Path, cpu_free: Option<u64>, cores: u64) -> Result<Value> {
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
                Err(Open::NotPlain(_) | Open::NotFolder(_)) => refused.push(n),
                Err(Open::Io(e)) => return Err(e),
            }
        }
        // The dead waiting files, counted through the folder's own descriptor; a `waiting` that
        // is a link or a file, or missing, counts zero and nothing behind a link is read.
        let stale_waiting = WaitDir::open(&folder.join("waiting"), false)
            .map_or(0, |w| stale_waiters(&w, "").len());
        // The cap a command under a slot gets, from the slot count of `limits.json` (or the
        // default); `null` when that file cannot be used, since acquire refuses then too.
        let (build_jobs_cap, min_cpu_free_pct) = match super::read_limits_file(folder) {
            Ok(file) => {
                let slots = file
                    .as_ref()
                    .and_then(|f| f.max)
                    .unwrap_or(super::DEFAULT_MAX);
                let min_cpu = file
                    .and_then(|f| f.min_cpu_free_pct)
                    .unwrap_or(DEFAULT_LIMITS.min_cpu_free_pct);
                (
                    json!({
                        "slots": slots,
                        "cpus": logical_cpus(),
                        "jobs": super::build_jobs_cap(slots),
                    }),
                    json!(min_cpu),
                )
            }
            Err(_) => (Value::Null, Value::Null),
        };
        // The warning line of every held slot that has one, so a reader finds them in one place.
        let warnings: Vec<Value> = held
            .iter()
            .filter_map(|h| h.get("warning").cloned())
            .collect();
        let mut answer = json!({
            "folder": folder.display().to_string(),
            "heavy_runs": held.len(),
            "held": held,
            "warnings": warnings,
            "refused": refused,
            "build_jobs_cap": build_jobs_cap,
            "cpu_free_pct": cpu_free,
            "cpu_cores": cores,
            "min_cpu_free_pct": min_cpu_free_pct,
            "stale_waiting": stale_waiting,
            "machine": machine_line(folder),
            "writes": false,
        });
        // The default folder also shows what an older build still holds in the earlier one
        // (read only); an earlier folder that cannot be read answers `null`.
        if let Some(earlier) = super::earlier_dir(folder) {
            if fs::symlink_metadata(&earlier).is_ok_and(|m| m.is_dir()) {
                answer["earlier_folder"] =
                    status_with(&earlier, cpu_free, cores).unwrap_or(Value::Null);
            }
        }
        Ok(answer)
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

/// #3033 D3: the three places that describe the slot say the same thing — a child or descendant
/// that outlives the command keeps the slot, status shows who, and no one edits a slot file.
/// (A fourth place, the Python tool's own text, went with the tool.)
#[cfg(test)]
mod docs_tests {
    #[test]
    fn the_three_descriptions_name_the_leftover_child_and_the_hands_off_rule() {
        let docs = [
            ("src/heavy_slot.rs", include_str!("heavy_slot.rs")),
            (
                // The slot rules moved from `workflows/run.md` into this reference with
                // SPEC-20261125-run-sync-procedure-receipt; RUN step 1 points to it.
                "skills/moai/references/internal/build-slots.md",
                include_str!("../skills/moai/references/internal/build-slots.md"),
            ),
            (
                "docs/rust-product-path.md",
                include_str!("../docs/rust-product-path.md"),
            ),
        ];
        for (name, text) in docs {
            assert!(text.contains("descendant"), "{name}: leftover descendant");
            assert!(
                text.contains("never edited or deleted by hand"),
                "{name}: hands off the slot file"
            );
            assert!(text.contains("`alive`"), "{name}: names alive");
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

/// SPEC-20261122-check-affected-one-slot REQ-4: which free slots a caller may take, with waiters
/// of mixed `--max` ahead of it.
#[cfg(all(test, unix))]
mod queue_tests {
    use super::sys::{lock_own, my_slots, waiter_max};
    use super::waiting_stamp;
    use std::time::{Duration, Instant};

    fn names(v: &[&str]) -> Vec<String> {
        v.iter().map(|s| s.to_string()).collect()
    }

    /// SPEC-20261126-leader-slot-front REQ-1 to REQ-3: the front stamp is the time in
    /// microseconds inside the same three-part name, so it sorts ahead of every nanosecond name
    /// (an older build included), keeps its `--max`, and two front stamps keep their arrival order.
    #[test]
    fn a_front_stamp_sorts_ahead_of_every_nanosecond_name_and_keeps_the_name_shape() {
        let ns: u128 = 1_795_747_522_123_456_789;
        assert_eq!(waiting_stamp(false, ns), ns);
        assert_eq!(waiting_stamp(true, ns), 1_795_747_522_123_456);
        let lane = format!("{ns:020}-7-1");
        let front = format!("{:020}-42-3", waiting_stamp(true, ns + 5_000_000));
        let earlier = format!("{:020}-43-2", waiting_stamp(true, ns));
        assert!(front.as_str() < lane.as_str(), "{front} {lane}");
        assert!(earlier.as_str() < front.as_str(), "{earlier} {front}");
        assert_eq!(waiter_max(&front), Some(3));
        // The lane waits behind a front waiter that wants the one free slot, as it does behind
        // any earlier waiter; the front waiter itself is not held back by the lane.
        assert_eq!(
            my_slots(vec![0], &names(&[&front, &lane]), &lane),
            Vec::<usize>::new()
        );
        assert_eq!(my_slots(vec![0], &names(&[&front, &lane]), &front), vec![0]);
    }

    #[test]
    fn a_waiting_name_gives_its_max_or_none() {
        assert_eq!(waiter_max("00000000000000000001-42-3"), Some(3));
        assert_eq!(waiter_max("00000000000000000001-42-0"), Some(0));
        for n in [
            "00000000000000000001-42",
            "00000000000000000001-42-",
            "00000000000000000001-42-x",
            "00000000000000000001-42-3-4",
            "00000000000000000001-42-+3",
            "",
        ] {
            assert_eq!(waiter_max(n), None, "{n}");
        }
    }

    #[test]
    fn earlier_waiters_take_only_the_slots_they_can_use() {
        let me = "00000000000000000005-9-2";
        // Nobody ahead: every free slot is the caller's.
        assert_eq!(my_slots(vec![0, 1], &names(&[me]), me), vec![0, 1]);
        // A narrower waiter ahead (`--max 1`) cannot use slot 1: the caller keeps it.
        let narrow = "00000000000000000001-7-1";
        assert_eq!(my_slots(vec![1], &names(&[narrow, me]), me), vec![1]);
        // With both free, slot 0 is left to the narrower waiter and the caller gets slot 1.
        assert_eq!(my_slots(vec![0, 1], &names(&[narrow, me]), me), vec![1]);
        // A waiter as wide as the caller, or one whose name gives no max, takes the lowest.
        for ahead in ["00000000000000000001-7-2", "00000000000000000001-7"] {
            assert_eq!(
                my_slots(vec![1], &names(&[ahead, me]), me),
                Vec::<usize>::new()
            );
            assert_eq!(my_slots(vec![0, 1], &names(&[ahead, me]), me), vec![1]);
        }
        // Waiters that arrived later never count.
        let later = "00000000000000000009-7-2";
        assert_eq!(my_slots(vec![0], &names(&[me, later]), me), vec![0]);
        // Two narrow waiters ahead want one slot between them: the first gets it, the second
        // takes nothing from the caller's slot 1.
        let narrow2 = "00000000000000000002-8-1";
        assert_eq!(
            my_slots(vec![0, 1], &names(&[narrow, narrow2, me]), me),
            vec![1]
        );
        // A waiter that can use no slot (`--max 0`) takes none.
        let zero = "00000000000000000001-7-0";
        assert_eq!(my_slots(vec![0], &names(&[zero, me]), me), vec![0]);
    }

    /// N2 of the audit: the lock of a caller's own fresh waiting file is tried again while
    /// another process holds it for a moment, and given up only after the last try.
    #[test]
    fn the_own_waiting_lock_is_tried_again_and_then_given_up() {
        let d = tempfile::tempdir().unwrap();
        let path = d.path().join("w");
        let mine = std::fs::File::create(&path).unwrap();
        let probe = std::fs::File::open(&path).unwrap();
        probe.lock_shared().unwrap();
        let start = Instant::now();
        let e = lock_own(&mine, &path, 3, Duration::from_millis(20)).unwrap_err();
        assert!(e.ends_with("held by another process"), "{e}");
        assert!(start.elapsed() >= Duration::from_millis(40), "three tries");
        // A probe that lets go during the tries does not stop the lock.
        let t = std::thread::spawn(move || {
            std::thread::sleep(Duration::from_millis(60));
            drop(probe);
        });
        lock_own(&mine, &path, 50, Duration::from_millis(10)).unwrap();
        t.join().unwrap();
        // One try with nobody in the way is enough.
        let other = d.path().join("v");
        let f = std::fs::File::create(&other).unwrap();
        lock_own(&f, &other, 1, Duration::from_millis(10)).unwrap();
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
            min_cpu_free_pct: None,
        }
    }

    #[test]
    fn no_file_and_no_flags_use_the_defaults() {
        let (max, l, notes) = effective(None, None, None, None, None);
        assert_eq!((max, l.min_free_gb, l.max_swap_pct), (DEFAULT_MAX, 50, 80));
        assert_eq!(l.min_mem_free_pct, 10);
        assert_eq!(DEFAULT_MAX, 3);
        assert!(notes.is_empty());
    }

    #[test]
    fn flags_only_tighten() {
        // Looser flags lose to the defaults, each with a note.
        let (max, l, notes) = effective(None, Some(9), Some(0), Some(100), None);
        assert_eq!((max, l.min_free_gb, l.max_swap_pct), (3, 50, 80));
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
        assert_eq!((max, l.min_free_gb, l.max_swap_pct), (3, u64::MAX, 80));
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
            home_build: Some(PathBuf::from(format!("{R}/h/.moai/build"))),
            home_build_named: Some(PathBuf::from(format!("{R}/h/.moai/build"))),
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
            Some(&*format!("{R}/r/repo/.claude/worktrees/one/target")),
            Some(&*format!("{R}/elsewhere/build")),
        ] {
            assert_eq!(note(t), None, "{t:?}");
        }
    }

    /// N1 of the audit of SPEC-20261123-lane-env-own-target: the retired home build root is
    /// looked for in the value as written as well as in the value with its links resolved. A
    /// root whose links lead elsewhere (`real`) still names the folder written under `<HOME>`,
    /// and the resolved root still names a folder written by its real path.
    #[test]
    fn the_home_build_root_is_matched_as_written_and_as_resolved() {
        let mut h = here();
        h.home_build = Some(PathBuf::from(format!("{R}/real/moai/build")));
        let cargo = cmd(&["cargo", "test"]);
        let note = |t: &str| build_folder_note("agent-7", &cargo, Some(OsStr::new(t)), Some(&h));
        // As written: under `<HOME>/.moai/build`, though the resolved root is elsewhere.
        let written = format!("{R}/h/.moai/build/repo-two");
        let n = note(&written).expect("a value written under the home build root");
        assert!(
            n.starts_with(&format!("heavy-slot: CARGO_TARGET_DIR {written} is ")),
            "{n}"
        );
        assert!(
            n.contains("inside the retired home build root /nonexistent-moai-c82/h/.moai/build;"),
            "{n}"
        );
        // As resolved: written by the real path of the root.
        let real = format!("{R}/real/moai/build/repo-two");
        let n = note(&real).expect("a value written by the resolved root");
        assert!(n.contains("/nonexistent-moai-c82/real/moai/build;"), "{n}");
    }

    #[test]
    fn a_retired_or_foreign_folder_is_named() {
        for (t, says) in [
            (
                format!("{R}/tmp/moai-rc-target-agent-7"),
                "the retired per-lane folder",
            ),
            (
                format!("{R}/tmp/moai-rc-target-x-a2846"),
                "the retired per-lane folder",
            ),
            (
                format!("{R}/h/.moai/build/repo-two"),
                "inside the retired home build root /nonexistent-moai-c82/h/.moai/build;",
            ),
            (
                format!("{R}/h/.moai/build/repo-one"),
                "inside the retired home build root",
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
                n.contains(&format!(
                    "builds in its own target/ ({R}/r/repo/.claude/worktrees/one/target)"
                )),
                "{n}"
            );
            assert!(n.contains("unset CARGO_TARGET_DIR"), "{n}");
            assert!(n.contains("moai lane-env --lane agent-7"), "{n}");
        }
        // Without a known home the retired home build root is not recognised.
        let mut no_home = here();
        no_home.home_build = None;
        no_home.home_build_named = None;
        let cargo = cmd(&["cargo", "test"]);
        let t = format!("{R}/h/.moai/build/repo-two");
        assert_eq!(
            build_folder_note("agent-7", &cargo, Some(OsStr::new(&t)), Some(&no_home)),
            None
        );
        // A full path to cargo counts as cargo too.
        let full = cmd(&["/usr/local/bin/cargo", "build"]);
        let t = format!("{R}/r/two/target");
        assert!(build_folder_note("agent-7", &full, Some(OsStr::new(&t)), Some(&here())).is_some());
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

/// The CPU rule's pure parts: the free percentage from two tick readings (a wrapping difference
/// included), the `/proc/stat` text, the limit's default and the decision after a wait.
#[cfg(all(test, unix))]
mod cpu_tests {
    use super::{
        cpu_free_pct, cpu_wait_decision, effective, parse_proc_stat, CpuGauge, CpuStep, LimitsFile,
        CPU_WAIT_CAP, DEFAULT_LIMITS,
    };
    use std::time::{Duration, Instant};

    #[test]
    fn a_look_with_no_new_tick_keeps_the_last_percentage() {
        let at = Instant::now();
        let mut g = CpuGauge {
            last: None,
            pct: None,
        };
        assert_eq!(g.take((0, 0), at), None);
        // The same counters again before any percentage: still none.
        assert_eq!(g.take((0, 0), at), None);
        assert_eq!(g.take((50, 100), at), Some(50));
        // A repeated answer of the system is no new sample.
        assert_eq!(g.take((50, 100), at), Some(50));
        // The next difference is measured from the last reading that changed.
        assert_eq!(g.take((140, 200), at), Some(90));
        // Counters that went backwards give no difference: the percentage holds, and the next
        // one is measured from them.
        assert_eq!(g.take((10, 100), at), Some(90));
        assert_eq!(g.take((30, 200), at), Some(20));
    }

    #[test]
    fn the_free_percentage_is_the_idle_share_of_the_ticks_between_two_readings() {
        assert_eq!(cpu_free_pct((10, 100), (60, 200)), Some(50));
        assert_eq!(cpu_free_pct((10, 100), (110, 200)), Some(100));
        assert_eq!(cpu_free_pct((10, 100), (10, 200)), Some(0));
        // Whole percent, rounded down: 19.9 is 19, so a limit of 20 still waits.
        assert_eq!(cpu_free_pct((0, 0), (199, 1000)), Some(19));
        assert_eq!(cpu_free_pct((0, 0), (200, 1000)), Some(20));
    }

    #[test]
    fn no_passed_tick_and_an_idle_count_above_the_total_give_no_nonsense() {
        assert_eq!(cpu_free_pct((5, 50), (5, 50)), None);
        // More idle ticks than ticks cannot be above 100.
        assert_eq!(cpu_free_pct((0, 0), (300, 100)), Some(100));
    }

    #[test]
    fn a_wrapping_tick_difference_is_read_right() {
        let prev = (u32::MAX - 20, u32::MAX - 30);
        let next = (prev.0.wrapping_add(50), prev.1.wrapping_add(100));
        assert!(next.1 < prev.1 && next.0 < prev.0, "the readings wrapped");
        assert_eq!(cpu_free_pct(prev, next), Some(50));
        // Only the total wrapped: the idle difference is still its own.
        let prev = (1_000, u32::MAX - 49);
        let next = (1_040, prev.1.wrapping_add(100));
        assert!(next.1 < prev.1);
        assert_eq!(cpu_free_pct(prev, next), Some(40));
    }

    #[test]
    fn a_total_that_went_backwards_is_no_reading() {
        // The wrapped difference would be above 2^31: no short pause is that long.
        assert_eq!(cpu_free_pct((0, 1_000), (0, 900)), None);
    }

    const PROC_STAT: &str = "\
cpu  100 20 30 400 50 6 7 8 90 10
cpu0 50 10 15 200 25 3 3 4 45 5
intr 12345
";

    #[test]
    fn proc_stat_counts_idle_and_iowait_as_idle_and_leaves_guest_time_out() {
        // total = user nice system idle iowait irq softirq steal = 100+20+30+400+50+6+7+8
        assert_eq!(parse_proc_stat(PROC_STAT), Some((450, 621)));
    }

    #[test]
    fn proc_stat_without_a_cpu_line_or_with_a_bad_field_is_no_reading() {
        assert_eq!(parse_proc_stat(""), None);
        assert_eq!(parse_proc_stat("cpu0 1 2 3 4\nintr 1\n"), None);
        assert_eq!(parse_proc_stat("cpu  1 2 3\n"), None);
        assert_eq!(parse_proc_stat("cpu  1 2 x 4 5\n"), None);
        // An old kernel gives four fields: idle without iowait.
        assert_eq!(parse_proc_stat("cpu  1 2 3 4\n"), Some((4, 10)));
    }

    #[test]
    fn below_the_limit_under_the_cap_waits_and_at_the_cap_the_caller_starts() {
        assert_eq!(cpu_wait_decision(true, Duration::ZERO), CpuStep::Wait);
        let just_under = CPU_WAIT_CAP - Duration::from_millis(1);
        assert_eq!(cpu_wait_decision(true, just_under), CpuStep::Wait);
        assert_eq!(cpu_wait_decision(true, CPU_WAIT_CAP), CpuStep::Run);
        assert_eq!(
            cpu_wait_decision(true, CPU_WAIT_CAP + Duration::from_secs(1)),
            CpuStep::Run
        );
        // Not low: never a wait, at any waited time.
        assert_eq!(cpu_wait_decision(false, Duration::ZERO), CpuStep::Run);
        assert_eq!(cpu_wait_decision(false, CPU_WAIT_CAP), CpuStep::Run);
        assert_eq!(CPU_WAIT_CAP, Duration::from_secs(300));
    }

    #[test]
    fn the_limit_is_0_without_a_file_and_the_files_value_with_one() {
        let (_, l, notes) = effective(None, None, None, None, None);
        assert_eq!(l.min_cpu_free_pct, 0);
        assert_eq!(DEFAULT_LIMITS.min_cpu_free_pct, 0);
        assert!(notes.is_empty());
        let f = LimitsFile {
            min_cpu_free_pct: Some(30),
            ..LimitsFile::default()
        };
        let limit = |f: &LimitsFile| {
            effective(Some(f), None, None, None, None)
                .1
                .min_cpu_free_pct
        };
        assert_eq!(limit(&f), 30);
        let f = LimitsFile {
            min_cpu_free_pct: Some(0),
            ..LimitsFile::default()
        };
        assert_eq!(limit(&f), 0);
        // A file without the key keeps the default.
        assert_eq!(limit(&LimitsFile::default()), 0);
    }
}
