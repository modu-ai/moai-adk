//! `leader-run` (SPEC-20261121-leader-run-lease): one supervising leader run per project. The
//! lease is a file in the project's state folder; a file lock held for the length of one call
//! lets only one call decide at a time, so two ticks that start together give one `taken` and
//! one `skipped`. The logic lives here, not in the library, because nothing else reads it.

use serde_json::{json, Value};
use std::fs::{self, OpenOptions};
use std::io::Write;
use std::path::{Path, PathBuf};

/// Folder of the lease under the project's state folder.
const STATE_FOLDER: &str = "state";
/// The lease, and the file whose lock lets one call decide at a time.
const LEASE_FILE: &str = "leader-run.lease";
const GUARD_FILE: &str = "leader-run.guard";
/// Seconds a lease runs when `--lease-secs` is not given, and the limits of the setting.
const DEFAULT_LEASE_SECS: u64 = 600;
const MIN_LEASE_SECS: u64 = 60;
const MAX_LEASE_SECS: u64 = 86_400;
/// Longest session name, in characters.
const MAX_SESSION_CHARS: usize = 80;

/// Take, end or read the supervision lease of a project
///
/// `moai leader-run <action> <project>` keeps one supervising leader run per project: the leader
/// calls `begin` first in each tick and runs the round only when the answer is `taken`; a second
/// tick gets `skipped` with the holder, its age and the lease end and changes nothing; a lease
/// that has ended (a restart, a leader that never ended) is taken over; `begin` with no ready
/// card or a HOLD takes no lease and answers `idle` with what wakes the run; `end` releases the
/// lease of the named session only. `status` reads only. The lease is the file
/// `leader-run.lease` in the state folder of the project (written with its lock file
/// `leader-run.guard`); nothing else is written, and no setting, approval or permission is
/// touched. `--now` is the clock in Unix seconds for tests; the system clock when absent.
#[derive(clap::Args)]
pub struct Args {
    #[command(subcommand)]
    action: Action,
}

#[derive(clap::Subcommand)]
enum Action {
    /// Writes one lease file. Start a tick: `taken`, `skipped` or `idle`
    ///
    /// Writes one lease file. `taken` when no unexpired lease of another session is held and at
    /// least one card is ready and no HOLD stands (the same session renews its own unexpired
    /// lease, `renewed: true`; an ended lease is taken over and `replaced` names it); `skipped`
    /// with `holder`, `age_secs` and `lease_until` when another session's lease runs, nothing
    /// changed; `idle` with `wake_on` (`card_ready` or `hold_lifted`) when `--ready` is 0 or
    /// `--hold` is given, no lease taken or touched.
    Begin {
        /// Any checkout of the project (its state folder holds the lease).
        project: PathBuf,
        /// The leader's session name, 1 to 80 characters, no control character.
        #[arg(long)]
        session: String,
        /// How many cards are ready now, as the leader counted them.
        #[arg(long)]
        ready: u32,
        /// A HOLD stands (money, legal, external contract, secret or real data).
        #[arg(long)]
        hold: bool,
        /// Seconds the lease runs (60 to 86400).
        #[arg(long, default_value_t = DEFAULT_LEASE_SECS)]
        lease_secs: u64,
        /// The clock, Unix seconds (tests); the system clock when absent.
        #[arg(long)]
        now: Option<u64>,
    },
    /// Writes (removes) the lease file. End a tick: release the lease of this session
    ///
    /// Writes (removes) the lease file. Releases the lease only when the named session holds it
    /// (`released: true`); another session's lease is refused and kept; no lease is not an error
    /// (`released: false`).
    End {
        /// Any checkout of the project.
        project: PathBuf,
        /// The session that holds the lease.
        #[arg(long)]
        session: String,
    },
    /// Reads only. Show the lease, if any
    ///
    /// Reads only. Prints `lease` (`session`, `started`, `lease_until`, `age_secs`, `ended`) or
    /// null; writes nothing, not even the lock file.
    Status {
        /// Any checkout of the project.
        project: PathBuf,
        /// The clock, Unix seconds (tests); the system clock when absent.
        #[arg(long)]
        now: Option<u64>,
    },
}

pub const OPERATION: &str = "leader.run";

#[derive(Clone)]
struct Lease {
    session: String,
    started: u64,
    lease_until: u64,
}

impl Lease {
    fn to_json(&self) -> Value {
        json!({"session": self.session, "started": self.started, "lease_until": self.lease_until})
    }
}

pub fn run(args: Args) -> Result<Value, String> {
    match args.action {
        Action::Begin {
            project,
            session,
            ready,
            hold,
            lease_secs,
            now,
        } => begin(&project, &session, ready, hold, lease_secs, clock(now)?),
        Action::End { project, session } => end(&project, &session),
        Action::Status { project, now } => status(&project, clock(now)?),
    }
}

fn clock(now: Option<u64>) -> Result<u64, String> {
    match now {
        Some(n) => Ok(n),
        None => std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .map(|d| d.as_secs())
            .map_err(|_| "the system clock is before 1970".to_string()),
    }
}

fn check_session(session: &str) -> Result<(), String> {
    let chars = session.chars().count();
    if chars == 0 || chars > MAX_SESSION_CHARS {
        return Err(format!(
            "--session must be 1 to {MAX_SESSION_CHARS} characters, not {chars}"
        ));
    }
    if session.chars().any(char::is_control) {
        return Err("--session must be one line without a control character".into());
    }
    Ok(())
}

fn state_dir(project: &Path) -> Result<PathBuf, String> {
    if !project.is_dir() {
        return Err("<project> must be the folder path of an existing project".into());
    }
    // The project's state folder is asked of `state_root`, never spelled here.
    Ok(moai_cli::state_root::path(project, STATE_FOLDER))
}

/// The lease in `dir`, `None` when there is no file; a file that is not a lease is refused and
/// left as it is.
fn read_lease(dir: &Path) -> Result<Option<Lease>, String> {
    let path = dir.join(LEASE_FILE);
    let text = match fs::read_to_string(&path) {
        Ok(t) => t,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(None),
        Err(e) => return Err(format!("could not read {LEASE_FILE}: {e}")),
    };
    let invalid = || {
        format!(
            "{LEASE_FILE} is invalid (not a lease), left as it is: {}",
            path.display()
        )
    };
    let v: Value = serde_json::from_str(&text).map_err(|_| invalid())?;
    let session = v["session"].as_str().ok_or_else(invalid)?.to_string();
    let started = v["started"].as_u64().ok_or_else(invalid)?;
    let lease_until = v["lease_until"].as_u64().ok_or_else(invalid)?;
    Ok(Some(Lease {
        session,
        started,
        lease_until,
    }))
}

/// Writes the lease whole: a temporary file in the same folder, then a rename, so a reader
/// never sees half a lease.
fn write_lease(dir: &Path, lease: &Lease) -> Result<(), String> {
    let tmp = dir.join(format!("{LEASE_FILE}.tmp-{}", std::process::id()));
    let mut f = OpenOptions::new()
        .write(true)
        .create(true)
        .truncate(true)
        .open(&tmp)
        .map_err(|e| format!("could not write the lease: {e}"))?;
    f.write_all(format!("{}\n", lease.to_json()).as_bytes())
        .and_then(|()| f.sync_all())
        .map_err(|e| format!("could not write the lease: {e}"))?;
    drop(f);
    fs::rename(&tmp, dir.join(LEASE_FILE)).map_err(|e| {
        let _ = fs::remove_file(&tmp);
        format!("could not write the lease: {e}")
    })
}

/// The folder (created when absent) and the held lock of one deciding call.
fn lock(dir: &Path) -> Result<fs::File, String> {
    fs::create_dir_all(dir).map_err(|e| format!("could not create the state folder: {e}"))?;
    let guard = OpenOptions::new()
        .read(true)
        .write(true)
        .create(true)
        .truncate(false)
        .open(dir.join(GUARD_FILE))
        .map_err(|e| format!("could not open the lock file: {e}"))?;
    guard
        .lock()
        .map_err(|e| format!("could not take the lock: {e}"))?;
    Ok(guard)
}

fn begin(
    project: &Path,
    session: &str,
    ready: u32,
    hold: bool,
    lease_secs: u64,
    now: u64,
) -> Result<Value, String> {
    check_session(session)?;
    if !(MIN_LEASE_SECS..=MAX_LEASE_SECS).contains(&lease_secs) {
        return Err(format!(
            "--lease-secs must be {MIN_LEASE_SECS} to {MAX_LEASE_SECS}, not {lease_secs}"
        ));
    }
    let dir = state_dir(project)?;
    if hold || ready == 0 {
        // Nothing to supervise: no lease taken, none touched, and the caller is told what
        // wakes the run (a HOLD wins when both hold).
        let wake_on = if hold { "hold_lifted" } else { "card_ready" };
        return Ok(json!({
            "action": "begin", "result": "idle", "wake_on": wake_on,
            "ready": ready, "hold": hold, "writes": false,
        }));
    }
    let _guard = lock(&dir)?;
    let existing = read_lease(&dir)?;
    let mut replaced = Value::Null;
    let mut started = now;
    let mut renewed = false;
    if let Some(old) = existing {
        if now <= old.lease_until {
            if old.session != session {
                return Ok(json!({
                    "action": "begin", "result": "skipped", "holder": old.session,
                    "age_secs": now.saturating_sub(old.started),
                    "started": old.started, "lease_until": old.lease_until, "writes": false,
                }));
            }
            renewed = true;
            started = old.started;
        } else {
            replaced = old.to_json();
        }
    }
    let lease = Lease {
        session: session.to_string(),
        started,
        lease_until: now.saturating_add(lease_secs),
    };
    write_lease(&dir, &lease)?;
    Ok(json!({
        "action": "begin", "result": "taken", "session": lease.session, "started": lease.started,
        "lease_until": lease.lease_until, "renewed": renewed, "replaced": replaced,
        "writes": true,
    }))
}

fn end(project: &Path, session: &str) -> Result<Value, String> {
    check_session(session)?;
    let dir = state_dir(project)?;
    if !dir.join(LEASE_FILE).exists() {
        return Ok(
            json!({"action": "end", "released": false, "reason": "no lease", "writes": false}),
        );
    }
    let _guard = lock(&dir)?;
    match read_lease(&dir)? {
        None => Ok(json!({"action": "end", "released": false, "reason": "no lease", "writes": false})),
        Some(l) if l.session != session => Err(format!(
            "the session that ends the lease must be its holder {}, not {session}; the lease is kept",
            l.session
        )),
        Some(_) => {
            fs::remove_file(dir.join(LEASE_FILE))
                .map_err(|e| format!("could not remove the lease: {e}"))?;
            Ok(json!({"action": "end", "released": true, "writes": true}))
        }
    }
}

fn status(project: &Path, now: u64) -> Result<Value, String> {
    let dir = state_dir(project)?;
    let lease = read_lease(&dir)?.map(|l| {
        let mut v = l.to_json();
        v["age_secs"] = json!(now.saturating_sub(l.started));
        v["ended"] = json!(now > l.lease_until);
        v
    });
    Ok(json!({"action": "status", "lease": lease, "writes": false}))
}
