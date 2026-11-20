//! The proof that lets `card reclaim` take a reservation before its lease ends
//! (SPEC-20261120-reclaim-dead-holder). The user decision of 2026-11-20: a reservation whose
//! holder session is proven gone is reclaimed even while its lease runs, and only then.
//!
//! Three proofs, every one read-only, every one answering "proven" or a named reason:
//! 1. the liveness lookup of the holder session answers `launcher-exited`
//!    ([`crate::lane_registry::liveness_of_session`]; `alive` and `unknown` both refuse);
//! 2. the recorded worktree is gone, or clean (nothing uncommitted, nothing unpushed): the same
//!    gate every takeover passes ([`super::dispatch::worktree_gate`]); a reservation that records
//!    no worktree proves nothing about the work and is refused;
//! 3. no running process has the worktree, or a folder below it, as its working folder.
//!
//! Anything that cannot be read, a platform without a probe included, is "unproven": the lease
//! rule stays as it was. The proof writes nothing; the ledger transaction that uses it re-checks
//! the holder and the generation ([`crate::ledger::reservation::reclaim_early`]).

use super::dispatch::worktree_gate;
use crate::ledger::{self, reservation};
use serde_json::{json, Value};
use std::path::Path;

/// Files a tool rewrites by itself on its next run, so one left in a worktree is no work
/// (SPEC-20261120-reclaim-dead-holder REQ-12): the statusline's GitHub counts, under either
/// state folder name. A name is listed here or it is work; nothing is matched by pattern.
const TOOL_CACHE_FILES: &[&str] = &[
    ".moai-cli/state/github/counts.json",
    ".moai/state/github/counts.json",
];

/// True when one `git status --porcelain=v1 --untracked-files=all` line is an untracked file
/// from [`TOOL_CACHE_FILES`]. A changed tracked file, an untracked file that is not listed and
/// a line that is quoted or unreadable are never a cache: they stay work.
pub(super) fn is_tool_cache_entry(line: &str) -> bool {
    line.strip_prefix("?? ")
        .is_some_and(|path| TOOL_CACHE_FILES.contains(&path))
}

/// A complete proof, about one holder.
#[derive(Debug)]
pub struct Proof {
    /// The holder session the proof is about; the ledger transaction refuses another one.
    pub holder: String,
    /// One short line for the reclaim reason.
    pub text: String,
    /// The same facts for the event payload.
    pub detail: Value,
}

/// What [`assess`] found.
#[derive(Debug)]
pub enum Early {
    /// The ordinary rules apply: the lease has ended, the caller is the holder, there is no
    /// held reservation, or the ledger cannot be read here (the ordinary path refuses it).
    NotNeeded,
    /// The lease is live and every proof holds.
    Proven(Proof),
    /// The lease is live and this proof is missing (the first one found).
    Unproven(String),
}

/// Looks at the card's current reservation and, when it is held by another session with a lease
/// that has not ended, collects the three proofs. Reads only.
pub fn assess(project: &Path, home: &Path, card_id: i64, to_holder: &str) -> Early {
    let Ok(id) = crate::ledger::command::project_id_of(project) else {
        return Early::NotNeeded;
    };
    let Ok(ro) = ledger::open_read_only(home, &id) else {
        return Early::NotNeeded;
    };
    let Ok(Some(cur)) = reservation::current(ro.conn(), card_id) else {
        return Early::NotNeeded;
    };
    if cur.holder == to_holder || cur.status != "held" {
        return Early::NotNeeded;
    }
    let Ok(now) = ro
        .conn()
        .query_row("SELECT unixepoch()", [], |r| r.get::<_, i64>(0))
    else {
        return Early::NotNeeded;
    };
    if reservation::lease_ended(cur.lease_until, now) {
        return Early::NotNeeded;
    }
    match prove(project, &cur) {
        Ok(proof) => Early::Proven(proof),
        Err(why) => Early::Unproven(why),
    }
}

/// The three proofs for one live reservation, in the order the refusal names them.
fn prove(project: &Path, cur: &reservation::Reservation) -> Result<Proof, String> {
    let Some(worktree) = cur.worktree.as_deref().filter(|w| !w.is_empty()) else {
        return Err(
            "the reservation records no worktree, so nothing shows that no work is \
                    left there"
                .to_string(),
        );
    };
    let (live, why) = crate::lane_registry::liveness_of_session(project, &cur.holder);
    let name = why["name"].as_str().unwrap_or("").to_string();
    match live {
        crate::lane_registry::Liveness::LauncherExited => {}
        crate::lane_registry::Liveness::Alive => {
            return Err("the holder session's launcher is still running".to_string());
        }
        crate::lane_registry::Liveness::Unknown => {
            let reason = why["reason"].as_str().unwrap_or("no reason given");
            return Err(format!(
                "the holder session's liveness is unknown, never a proof: {reason}"
            ));
        }
    }
    // The clean-or-gone gate every takeover passes: a path that exists with uncommitted or
    // unpushed work, or that cannot be told apart from "gone", is refused.
    worktree_gate(Some(worktree)).map_err(|e| e.replace("card conflict in ", ""))?;
    let path = Path::new(worktree);
    let gone = !path.try_exists().unwrap_or(true);
    let shape = if gone { "gone" } else { "clean" };
    if !gone {
        let here = processes_in(path)?;
        if here > 0 {
            return Err(format!(
                "{here} running process(es) have the worktree as their working folder"
            ));
        }
    }
    Ok(Proof {
        holder: cur.holder.clone(),
        text: format!(
            "early reclaim: launcher exited (session {name}); worktree {shape}; no process in it"
        ),
        detail: json!({
            "liveness": "launcher-exited",
            "session_name": name,
            "worktree": shape,
            "processes_in_worktree": 0,
        }),
    })
}

/// How many running processes have `dir`, or a folder below it, as their working folder.
/// `Err` when the platform has no probe or the probe cannot answer: that is "unproven", never
/// "none". Only whole path components match, so `/w/tree-2` is not inside `/w/tree`.
pub fn processes_in(dir: &Path) -> Result<usize, String> {
    let dir = std::fs::canonicalize(dir)
        .map_err(|e| format!("the worktree cannot be resolved for the process check: {e}"))?;
    let cwds = working_folders()?;
    Ok(cwds.iter().filter(|c| c.starts_with(&dir)).count())
}

/// The working folder of every process this user can see.
#[cfg(target_os = "macos")]
fn working_folders() -> Result<Vec<std::path::PathBuf>, String> {
    // `lsof` has a fixed home on macOS; a child without a PATH must still find it.
    let out = std::process::Command::new("/usr/sbin/lsof")
        .args(["-a", "-d", "cwd", "-Fn"])
        .stdin(std::process::Stdio::null())
        .output()
        .map_err(|e| format!("the process check could not run lsof: {e}"))?;
    let text = String::from_utf8_lossy(&out.stdout);
    // `-Fn` prints one `n<path>` line for every cwd it could read; a process it may not read
    // (another user's) is skipped by lsof itself and makes it exit 1, so the exit code alone
    // is no answer. An empty answer cannot be right: this very process has a folder.
    let folders: Vec<std::path::PathBuf> = text
        .lines()
        .filter_map(|l| l.strip_prefix('n'))
        .map(std::path::PathBuf::from)
        .collect();
    if folders.is_empty() {
        return Err("the process check got no answer from lsof".to_string());
    }
    Ok(folders)
}

/// The working folder of every process of this user, read from `/proc`. A process of another
/// user cannot hold an agent's worktree and may not be readable; one of this user that cannot
/// be read makes the answer unproven.
#[cfg(target_os = "linux")]
fn working_folders() -> Result<Vec<std::path::PathBuf>, String> {
    use std::os::unix::fs::MetadataExt;
    let me = std::fs::metadata("/proc/self")
        .map_err(|e| format!("the process check cannot read /proc: {e}"))?
        .uid();
    let mut folders = Vec::new();
    for entry in std::fs::read_dir("/proc")
        .map_err(|e| format!("the process check cannot read /proc: {e}"))?
        .flatten()
    {
        let name = entry.file_name();
        if !name.to_string_lossy().bytes().all(|b| b.is_ascii_digit()) {
            continue;
        }
        let Ok(meta) = entry.metadata() else {
            continue; // the process ended while the list was read
        };
        if meta.uid() != me {
            continue;
        }
        match std::fs::read_link(entry.path().join("cwd")) {
            Ok(p) => folders.push(p),
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => {}
            Err(e) => {
                return Err(format!(
                    "the process check cannot read the working folder of process {}: {e}",
                    name.to_string_lossy()
                ))
            }
        }
    }
    Ok(folders)
}

#[cfg(not(any(target_os = "macos", target_os = "linux")))]
fn working_folders() -> Result<Vec<std::path::PathBuf>, String> {
    Err("this platform has no process check, so no early reclaim".to_string())
}

#[cfg(all(test, unix))]
mod tests {
    use super::*;
    use std::process::{Child, Command, Stdio};

    fn sleeper(dir: &Path) -> Child {
        Command::new("sleep")
            .arg("30")
            .current_dir(dir)
            .stdin(Stdio::null())
            .spawn()
            .expect("start sleep")
    }

    /// AC-4 (REQ-4): a process inside the folder is counted, one in a subfolder too, and a
    /// sibling whose name only starts like the folder is not.
    #[test]
    fn processes_in_counts_whole_components_only() {
        let base = tempfile::tempdir().unwrap();
        let tree = base.path().join("tree");
        let sub = tree.join("sub");
        let sibling = base.path().join("tree-2");
        for d in [&sub, &sibling] {
            std::fs::create_dir_all(d).unwrap();
        }
        assert_eq!(processes_in(&tree).unwrap(), 0, "nobody is in it yet");
        let mut in_sibling = sleeper(&sibling);
        assert_eq!(
            processes_in(&tree).unwrap(),
            0,
            "a name that only starts alike is not inside"
        );
        let mut in_tree = sleeper(&tree);
        let mut in_sub = sleeper(&sub);
        assert_eq!(
            processes_in(&tree).unwrap(),
            2,
            "the folder and a folder below it count"
        );
        for c in [&mut in_tree, &mut in_sub, &mut in_sibling] {
            let _ = c.kill();
            let _ = c.wait();
        }
        assert_eq!(processes_in(&tree).unwrap(), 0, "after they end");
    }

    #[test]
    fn processes_in_a_missing_folder_is_unproven_not_zero() {
        let base = tempfile::tempdir().unwrap();
        let gone = base.path().join("gone");
        assert!(
            processes_in(&gone).is_err(),
            "a folder that cannot be resolved gives no answer"
        );
    }
}

#[cfg(test)]
mod cache_tests {
    use super::is_tool_cache_entry;

    /// AC-10 (REQ-12): only an untracked, listed cache file is a cache.
    #[test]
    fn only_an_untracked_listed_file_is_a_cache() {
        assert!(is_tool_cache_entry("?? .moai-cli/state/github/counts.json"));
        assert!(is_tool_cache_entry("?? .moai/state/github/counts.json"));
        // A changed tracked file of the same name is work.
        assert!(!is_tool_cache_entry(
            " M .moai-cli/state/github/counts.json"
        ));
        assert!(!is_tool_cache_entry(
            "A  .moai-cli/state/github/counts.json"
        ));
        // An unlisted name, a folder and a path that only ends alike are work.
        assert!(!is_tool_cache_entry("?? .moai-cli/state/github/other.json"));
        assert!(!is_tool_cache_entry("?? .moai-cli/state/"));
        assert!(!is_tool_cache_entry(
            "?? sub/.moai-cli/state/github/counts.json"
        ));
        assert!(!is_tool_cache_entry(
            "?? .moai-cli/state/github/counts.json.bak"
        ));
        assert!(!is_tool_cache_entry(""));
    }
}
