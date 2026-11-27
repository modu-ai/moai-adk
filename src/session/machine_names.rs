//! SPEC-20261127-launcher-session-names: names of launcher sessions that are unique on the whole
//! machine. A launcher that names a session (`-l`, `-a`, or a `--name` beside one of them) first
//! takes the exclusive lock of `<HOME>/.moai/session-names/<name>.live`, then the clone's own
//! `<name>.live` lock, and holds both while the host runs. The OS drops a lock when the launcher
//! or its host ends, however it ends, so no stale entry stays behind. The body of a lock file
//! (name, project, process id, time) is information for the refusal message only; the lock decides.
//!
//! A probe (`--dry-run`) asks for a shared lock and lets go at once: it creates no folder, no file
//! and no lock. The lock files stay (R-5 of the SPEC): at most the names of the launcher's list
//! plus one leader file per project name.
use std::fs::{self, File, OpenOptions};
use std::io::{Seek, SeekFrom, Write};
use std::path::{Path, PathBuf};

/// The folder of the machine-wide name locks under the home state folder.
const FOLDER: &str = "session-names";

/// Why a name was not given.
#[derive(Debug)]
pub enum Refusal {
    /// Another launcher holds the name (or its lock cannot be judged); the project it recorded,
    /// when that could be read.
    Held { project: Option<String> },
    /// The folder or the lock file cannot be used: the reason.
    Unusable(String),
}

/// `<HOME>/.moai/session-names`.
pub fn folder(home: &Path) -> PathBuf {
    crate::home_workdir::root(home).join(FOLDER)
}

/// Whether the folder exists; a link or a thing that is not a folder is refused.
fn folder_exists(dir: &Path) -> Result<bool, String> {
    match fs::symlink_metadata(dir) {
        Ok(m) if m.file_type().is_symlink() => Err(format!(
            "session names folder {} is a symlink; refused",
            dir.display()
        )),
        Ok(m) if !m.is_dir() => Err(format!(
            "session names folder {} is not a plain folder; refused",
            dir.display()
        )),
        Ok(_) => Ok(true),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(false),
        Err(e) => Err(format!(
            "session names folder {} cannot be read: {e}",
            dir.display()
        )),
    }
}

/// The project a lock file says holds it: the `project=` line, reduced to the registry's
/// characters so that nothing but a plain name reaches a message.
fn holder_project(path: &Path) -> Option<String> {
    let text = fs::read_to_string(path).ok()?;
    let value = text
        .lines()
        .find_map(|l| l.strip_prefix("project="))?
        .chars()
        .filter(|c| c.is_ascii_alphanumeric() || matches!(c, '.' | '_' | '-'))
        .take(64)
        .collect::<String>();
    (!value.is_empty()).then_some(value)
}

/// A machine-wide lock held until it is dropped.
#[derive(Debug)]
pub struct Claim {
    _file: File,
}

/// Takes the exclusive machine-wide lock of `name` for the launcher of `project` (REQ-4, REQ-6,
/// REQ-7). The folder and the file are made when absent; a link, or anything that is not a plain
/// folder or file, is refused before it is opened.
pub fn claim(home: &Path, name: &str, project: &str) -> Result<Claim, Refusal> {
    if !crate::lane_registry::valid_name(name) {
        return Err(Refusal::Unusable(
            "session name is invalid: use 1 to 64 of A-Z a-z 0-9 . _ -".into(),
        ));
    }
    let dir = folder(home);
    folder_exists(&dir).map_err(Refusal::Unusable)?;
    fs::create_dir_all(&dir).map_err(|e| {
        Refusal::Unusable(format!(
            "session names folder {} cannot be made: {e}",
            dir.display()
        ))
    })?;
    let path = dir.join(format!("{name}.live"));
    super::refuse_unless_regular(&path).map_err(Refusal::Unusable)?;
    let unusable =
        || Refusal::Unusable(format!("session name lock {} unavailable", path.display()));
    let file = OpenOptions::new()
        .read(true)
        .write(true)
        .create(true)
        .truncate(false)
        .open(&path)
        .map_err(|_| unusable())?;
    match file.try_lock() {
        Ok(()) => {}
        Err(fs::TryLockError::WouldBlock) => {
            return Err(Refusal::Held {
                project: holder_project(&path),
            })
        }
        Err(fs::TryLockError::Error(_)) => return Err(unusable()),
    }
    // REQ-6: information only. A failed write leaves the lock held and the launch goes on.
    let secs = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map_or(0, |d| d.as_secs());
    let body = format!(
        "name={name}\nproject={project}\npid={}\ntime={secs}\n",
        std::process::id()
    );
    let mut f = &file;
    let _ = file.set_len(0);
    let _ = f.seek(SeekFrom::Start(0));
    let _ = f.write_all(body.as_bytes());
    Ok(Claim { _file: file })
}

/// Whether `name` could be claimed now, without making anything (REQ-8). A lock file it cannot
/// judge answers `Held`; a folder it cannot use answers `Unusable`.
pub fn probe(home: &Path, name: &str) -> Result<(), Refusal> {
    let dir = folder(home);
    if !folder_exists(&dir).map_err(Refusal::Unusable)? {
        return Ok(());
    }
    let path = dir.join(format!("{name}.live"));
    let taken = || Refusal::Held { project: None };
    if super::refuse_unless_regular(&path).is_err() {
        return Err(taken());
    }
    let file = match OpenOptions::new().read(true).open(&path) {
        Ok(f) => f,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(()),
        Err(_) => return Err(taken()),
    };
    match file.try_lock_shared() {
        Ok(()) => Ok(()),
        Err(fs::TryLockError::WouldBlock) => Err(Refusal::Held {
            project: holder_project(&path),
        }),
        Err(fs::TryLockError::Error(_)) => Err(taken()),
    }
}

/// The refusal text of a name another launcher holds (REQ-5).
pub fn held_message(name: &str, project: Option<&str>) -> String {
    let who = project.map_or_else(
        || "another session".to_string(),
        |p| format!("a session of project {p}"),
    );
    format!(
        "{name} is already held by {who} on this machine: stop that session, or name this one with `--name <name>`"
    )
}

/// The text of a launcher list used up machine-wide (REQ-2).
pub fn used_up_message(count: usize) -> String {
    format!(
        "every name of the launcher's list ({count} names) is taken on this machine: name this session with `--name <name>`"
    )
}

/// The name a launch will use and the names to go on to when the first is taken in the meantime.
#[derive(Debug)]
pub struct Plan {
    pub name: String,
    pub alternatives: Vec<String>,
}

fn refusal_text(name: &str, why: Refusal) -> String {
    match why {
        Refusal::Held { project } => held_message(name, project.as_deref()),
        Refusal::Unusable(reason) => reason,
    }
}

/// A name that is fixed (`-l`, or `--name` beside a role flag): refused when another launcher
/// holds it. A name the registry rejects is left to the launcher's own check (REQ-5, D-3).
pub fn plan_fixed(home: &Path, name: &str) -> Result<Plan, String> {
    if crate::lane_registry::valid_name(name) {
        probe(home, name).map_err(|why| refusal_text(name, why))?;
    }
    Ok(Plan {
        name: name.to_string(),
        alternatives: Vec::new(),
    })
}

/// A lane name chosen from `candidates` (in the launcher's seeded order): the first that no
/// launcher of the machine holds; the ones after it are kept for a race at the claim (REQ-1, REQ-2).
pub fn plan_lane(home: &Path, candidates: Vec<String>) -> Result<Plan, String> {
    let count = candidates.len();
    let mut free: Vec<String> = Vec::new();
    for name in candidates {
        match probe(home, &name) {
            Ok(()) => free.push(name),
            Err(Refusal::Held { .. }) => {}
            Err(Refusal::Unusable(reason)) => return Err(reason),
        }
    }
    if free.is_empty() {
        return Err(used_up_message(count));
    }
    let name = free.remove(0);
    Ok(Plan {
        name,
        alternatives: free,
    })
}

/// What a launch holds while its host runs: the machine-wide claim, then the clone's lock.
#[derive(Debug)]
pub struct Held {
    _machine: Option<Claim>,
    _clone: Option<File>,
}

/// Takes the locks of a launch, machine-wide first when `machine` is set (a role flag) and then the
/// clone's own (REQ-4). The clone's lock is let go together with the machine-wide one when the
/// second cannot be taken. A name that is taken in the meantime goes on to the next of
/// `alternatives`; the last name's refusal is the answer. Outside a git repository there is no
/// sessions folder and only the machine-wide claim is made. Returns the name that was claimed.
pub fn hold(
    here: &Path,
    home: Option<&Path>,
    name: &str,
    alternatives: &[String],
    machine: bool,
) -> Result<(String, Held), String> {
    let project = machine.then(|| super::project_name(here));
    let mut names: Vec<&str> = vec![name];
    names.extend(alternatives.iter().map(String::as_str));
    let last = names.len() - 1;
    for (i, cand) in names.into_iter().enumerate() {
        let fail = |why: String| format!("cannot start {cand}: {why}");
        let machine_claim = match (&project, home) {
            (Some(project), Some(home)) => match claim(home, cand, project) {
                Ok(c) => Some(c),
                Err(Refusal::Held { .. }) if i < last => continue,
                Err(why) => return Err(fail(refusal_text(cand, why))),
            },
            (Some(_), None) => return Err(fail("the home folder cannot be resolved".into())),
            (None, _) => None,
        };
        let clone_lock = match super::agent_seq_dir(here) {
            Ok(dir) => match super::try_lock_named_live(&dir, cand) {
                Ok(f) => Some(f),
                Err(super::LiveLockError::Held(_)) if i < last => continue,
                Err(super::LiveLockError::Held(why) | super::LiveLockError::Other(why)) => {
                    return Err(fail(why))
                }
            },
            // No sessions folder (not a git repository): no clone lock, the lookup stays unknown.
            Err(_) => None,
        };
        return Ok((
            cand.to_string(),
            Held {
                _machine: machine_claim,
                _clone: clone_lock,
            },
        ));
    }
    Err("no session name to start".into())
}
