//! `ledger-recover-plan` and `ledger-recover-apply` (Issue #3068 P6): the command layer of the
//! ledger recovery. The library (`ledger::recover`) never starts a process or reads outside the
//! ledger; this module gathers the facts it needs and passes them in as plain data:
//!
//! - the claim journal (`claim::Journal`, the last record of every claim; its lease judged with
//!   the claim module's own rule, [`crate::claim::is_live`]);
//! - the worktree of every active reservation, found on this machine with read-only `git` (the
//!   environment scrubbed like every other git spawn), behind [`WorktreeProbe`] so tests can
//!   inject facts;
//! - an optional observation file (JSON, the shape of [`Observations`]): remote facts, and any
//!   journal facts the operator wants to add. A `worktrees` entry in the file counts only for a
//!   path that does not exist on this machine (a worktree that lives on another one).
//!
//! The plan is read-only. The apply recomputes the plan from the same inputs and acts only when
//! its digest equals the one the caller was shown and a permission reference is given.

use crate::ledger::recover::{self, JournalFact, Observations, WorktreeFact};
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::path::{Path, PathBuf};
use std::process::Command;

type Result<T> = std::result::Result<T, String>;

/// The observation file is read whole; a larger one is refused.
pub const MAX_OBSERVATION_FILE_BYTES: u64 = 1 << 20;

/// Finds what is true of a worktree path on this machine.
pub trait WorktreeProbe {
    fn probe(&self, path: &str) -> WorktreeFact;
}

/// The real probe: the folder is a git worktree root and `git` names its branch.
pub struct GitProbe;

fn git(dir: &Path, args: &[&str]) -> Option<String> {
    let mut cmd = Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .arg("--no-optional-locks")
        .arg("-C")
        .arg(dir)
        .args(args)
        .output()
        .ok()?;
    out.status
        .success()
        .then(|| String::from_utf8_lossy(&out.stdout).into_owned())
}

impl WorktreeProbe for GitProbe {
    fn probe(&self, path: &str) -> WorktreeFact {
        let missing = WorktreeFact {
            path: path.to_string(),
            exists: false,
            branch: None,
            clean: false,
        };
        let dir = Path::new(path);
        // A relative path means nothing without a working folder: it is not observed.
        if !dir.is_absolute() || !dir.is_dir() {
            return missing;
        }
        // The folder itself must be the worktree's top, not a folder inside some repository.
        let top = git(dir, &["rev-parse", "--show-toplevel"]).map(|t| PathBuf::from(t.trim()));
        let same = match (
            top.and_then(|t| std::fs::canonicalize(t).ok()),
            std::fs::canonicalize(dir),
        ) {
            (Some(a), Ok(b)) => a == b,
            _ => false,
        };
        if !same {
            return missing;
        }
        let branch = git(dir, &["symbolic-ref", "--short", "-q", "HEAD"])
            .map(|b| b.trim().to_string())
            .filter(|b| !b.is_empty());
        let clean = git(
            dir,
            &["status", "--porcelain=v1", "--untracked-files=normal"],
        )
        .is_some_and(|t| t.lines().all(str::is_empty));
        WorktreeFact {
            path: path.to_string(),
            exists: true,
            branch,
            clean,
        }
    }
}

/// Reads an observation file (JSON in the shape of [`Observations`]; unknown keys refused).
pub fn read_observations(path: &Path) -> Result<Observations> {
    let meta =
        std::fs::metadata(path).map_err(|_| "observation file cannot be read".to_string())?;
    if !meta.is_file() || meta.len() > MAX_OBSERVATION_FILE_BYTES {
        return Err(format!(
            "observation file must be a regular file of at most {MAX_OBSERVATION_FILE_BYTES} bytes"
        ));
    }
    let text = std::fs::read_to_string(path)
        .map_err(|_| "observation file cannot be read as UTF-8 text".to_string())?;
    let value: Value =
        serde_json::from_str(&text).map_err(|e| format!("observation file is not valid: {e}"))?;
    if !value.is_object() {
        return Err("observation file is not valid: the top level must be a JSON object".into());
    }
    serde_json::from_value(value).map_err(|e| format!("observation file is not valid: {e}"))
}

/// The last record of every claim in the journal at `root`, as facts: holder = owner id,
/// generation = epoch, live = the claim is still held and its lease has not ended at `now`.
pub fn journal_facts(root: &Path, now: i64) -> Result<Vec<JournalFact>> {
    let journal = crate::claim::Journal::open_read(root)?;
    let mut last: BTreeMap<String, crate::claim::ClaimRecord> = BTreeMap::new();
    for event in journal.read_events()? {
        last.insert(event.record.task_key.clone(), event.record);
    }
    Ok(last
        .into_values()
        .map(|r| JournalFact {
            holder: r.owner.id.clone(),
            generation: i64::try_from(r.epoch).unwrap_or(i64::MAX),
            live: crate::claim::is_active(&r.state)
                && r.lease_until
                    .as_deref()
                    .is_some_and(|l| crate::claim::is_live(l, now)),
        })
        .collect())
}

/// What the plan and the apply both read.
pub struct Inputs<'a> {
    pub project: &'a Path,
    pub home: &'a Path,
    pub source: &'a Path,
    /// A claim journal folder; the project's own when `None`.
    pub journal: Option<&'a Path>,
    pub observations: Option<&'a Path>,
    /// Unix seconds the journal leases are judged against.
    pub now: i64,
}

struct Gathered {
    project_id: String,
    observations: Observations,
    report: Value,
}

fn gather(inputs: &Inputs<'_>, probe: &dyn WorktreeProbe) -> Result<Gathered> {
    let project_id = crate::ledger::command::project_id_of(inputs.project)?;
    let from_file = match inputs.observations {
        Some(path) => read_observations(path)?,
        None => Observations::default(),
    };

    let mut journal = from_file.journal.clone();
    let (journal_root, journal_error) = match inputs.journal {
        Some(root) => (Some(root.to_path_buf()), None),
        None => match crate::claim::journal_root(inputs.project) {
            Ok(root) => (Some(root), None),
            Err(e) => (None, Some(e)),
        },
    };
    let mut journal_records = 0usize;
    let mut journal_error = journal_error;
    if let Some(root) = &journal_root {
        match journal_facts(root, inputs.now) {
            Ok(facts) => {
                journal_records = facts.len();
                journal.extend(facts);
            }
            Err(e) => journal_error = Some(e),
        }
    }

    let reservations =
        recover::queries(inputs.home, &project_id, inputs.source).map_err(|e| e.to_string())?;
    let mut worktrees: Vec<WorktreeFact> = Vec::new();
    let mut seen: Vec<&str> = Vec::new();
    for path in reservations.iter().filter_map(|r| r.worktree.as_deref()) {
        if seen.contains(&path) {
            continue;
        }
        seen.push(path);
        let mut fact = probe.probe(path);
        if !fact.exists {
            if let Some(given) = from_file.worktrees.iter().find(|w| w.path == path) {
                fact = given.clone();
            }
        }
        worktrees.push(fact);
    }
    let observations = Observations {
        journal,
        worktrees,
        remotes: from_file.remotes.clone(),
    };
    let report = json!({
        "journal": {
            "root": journal_root.map(|r| r.display().to_string()),
            "records": journal_records,
            "from_file": from_file.journal.len(),
            "error": journal_error,
        },
        "worktrees": observations.worktrees,
        "remotes": observations.remotes.len(),
        "now": inputs.now,
    });
    Ok(Gathered {
        project_id,
        observations,
        report,
    })
}

/// `ledger-recover-plan`: what recovering from `source` would build. Writes nothing.
pub fn plan(inputs: &Inputs<'_>, probe: &dyn WorktreeProbe) -> Result<Value> {
    let g = gather(inputs, probe)?;
    let plan = recover::plan(inputs.home, &g.project_id, inputs.source, &g.observations)
        .map_err(|e| e.to_string())?;
    let mut out = serde_json::to_value(&plan).map_err(|e| e.to_string())?;
    out["observed"] = g.report;
    out["apply"] = json!(
        "ledger-recover-apply with the same project, home, source and observations, this plan_digest and a permission reference, only after the user approves"
    );
    out["never"] = json!(
        "nothing is deleted or overwritten: the damaged file is kept under a damaged-<UTC> name and leftover register temp files are only reported"
    );
    out["network"] = json!("disabled");
    out["writes"] = json!(false);
    Ok(out)
}

/// `ledger-recover-apply`: recovers when the freshly computed plan has the digest the caller
/// holds. The ledger at the real name stays as it is when it is already this plan's result
/// (`unchanged`) or a usable ledger of another history (refused).
pub fn apply(
    inputs: &Inputs<'_>,
    plan_digest: &str,
    permission_reference: &str,
    probe: &dyn WorktreeProbe,
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("permission reference required".into());
    }
    let g = gather(inputs, probe)?;
    let result = recover::apply(
        inputs.home,
        &g.project_id,
        inputs.source,
        &g.observations,
        plan_digest,
    )
    .map_err(|e| e.to_string())?;
    let mut out = serde_json::to_value(&result).map_err(|e| e.to_string())?;
    out["project_id"] = json!(g.project_id);
    out["observed"] = g.report;
    out["network"] = json!("disabled");
    Ok(out)
}
