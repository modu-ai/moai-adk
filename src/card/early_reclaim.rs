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

/// The statusline's GitHub counts file, relative to a state folder. A tool rewrites it by itself
/// on its next run, so one left in a worktree is no work (SPEC-20261120-reclaim-dead-holder
/// REQ-12). A name is listed or it is work; nothing is matched by pattern.
const GITHUB_COUNTS: &str = "state/github/counts.json";

/// The earlier (Go) generation's cache of the same counts, which the statusline still reads.
const GO_GITHUB_COUNTS: &str = ".moai/state/github/counts.json";

/// True when one `git status --porcelain=v1 --untracked-files=all` line of the worktree `root`
/// is an untracked tool cache file: the counts file under the state folder `root` uses now, or
/// the earlier generation's. A changed tracked file, an untracked file that is not listed and a
/// line that is quoted or unreadable are never a cache: they stay work.
pub(super) fn is_tool_cache_entry(root: &Path, line: &str) -> bool {
    let current = crate::state_root::rel(root, GITHUB_COUNTS);
    line.strip_prefix("?? ")
        .is_some_and(|path| path == current || path == GO_GITHUB_COUNTS)
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
    match prove(project, &cur, now) {
        Ok(proof) => Early::Proven(proof),
        Err(why) => Early::Unproven(why),
    }
}

/// The three proofs for one live reservation, in the order the refusal names them. `now` is the
/// ledger's clock, the one the reservation lease is read against.
fn prove(project: &Path, cur: &reservation::Reservation, now: i64) -> Result<Proof, String> {
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
    // The lookup above answers `launcher-exited` for every session that has no `.live` file:
    // one registered with `session-register` alone (an Orca pane, a plain `claude`) never had a
    // launcher, and is not gone for that. The launcher counts only when its lock file exists
    // (and nobody holds it, the answer above) and the registry lease of the session has ended
    // (SPEC-20261120-reclaim-dead-holder REQ-1; audit of the early reclaim, B1).
    let trace =
        crate::lane_registry::launcher_trace_of_session(project, &cur.holder).map_err(|e| {
            format!("the holder session's launcher record cannot be read, never a proof: {e}")
        })?;
    if !trace.live_file {
        return Err(format!(
            "the holder session {name} has no launcher lock file (.live), so nothing shows that a \
             launcher ever ran for it; a session registered with session-register alone is never \
             proven gone"
        ));
    }
    if trace.lease_until > now {
        return Err(format!(
            "the registry lease of the holder session {name} has not ended ({} s left)",
            trace.lease_until - now
        ));
    }
    // The clean-or-gone gate every takeover passes: a path that exists with uncommitted or
    // unpushed work, or that cannot be told apart from "gone", is refused.
    worktree_gate(Some(worktree)).map_err(|e| e.replace("card conflict in ", ""))?;
    let path = Path::new(worktree);
    let gone = !path.try_exists().unwrap_or(true);
    let shape = if gone { "gone" } else { "clean" };
    if !gone {
        // A live lease makes a missed draft cost more than an expired one does, so the early
        // path also counts what the takeover gate leaves out (REQ-3).
        remaining_work(path)?;
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
            "live_file": true,
            "registry_lease_ended": true,
            "worktree": shape,
            "processes_in_worktree": 0,
        }),
    })
}

/// Ignored entries of a worktree that a tool rebuilds by itself, so one left behind is no draft:
/// the log folder, the cargo output folder and a Python bytecode folder. Every other ignored
/// entry is a file somebody chose to keep out of git, so it counts as work. `entry` is a folder
/// entry of `git status` (it ends with a slash) or a single file below one; a folder name counts
/// as a whole component, so `my__pycache__/` and `src/target/` are not on the list.
fn is_regenerable_ignored(entry: &str) -> bool {
    entry.starts_with(".moai/logs/")
        || entry.starts_with("target/")
        || entry.starts_with("__pycache__/")
        || entry.contains("/__pycache__/")
}

/// What the early reclaim counts as work beyond the takeover gate (REQ-3): a stash entry of the
/// worktree's own branch, unpushed commits on a local branch nobody has checked out (the
/// checked-out ones are other worktrees' own work, and HEAD's is the gate's), and an ignored
/// entry that is not on the regenerable list. `Err` names the first one found, or what could not
/// be read: unreadable is unproven, never "none".
fn remaining_work(path: &Path) -> Result<(), String> {
    let unreadable = |what: &str| format!("git could not read {what} of the worktree");
    let git = |args: &[&str], what: &str| {
        super::dispatch::git_output(path, args).ok_or_else(|| unreadable(what))
    };
    let head = git(&["rev-parse", "--abbrev-ref", "HEAD"], "its branch")?;
    let head = head.trim();
    let branch = if head == "HEAD" { "(no branch)" } else { head };
    let stashes = git(&["stash", "list", "--format=%gs"], "the stash list")?
        .lines()
        .filter(|l| {
            l.strip_prefix("WIP on ")
                .or_else(|| l.strip_prefix("On "))
                .is_some_and(|rest| {
                    rest.strip_prefix(branch)
                        .is_some_and(|r| r.starts_with(':'))
                })
        })
        .count();
    if stashes > 0 {
        return Err(format!(
            "the worktree holds {stashes} stash entr(ies) of branch {branch}, which a reclaim \
             would orphan"
        ));
    }
    let checked_out: Vec<String> = git(&["worktree", "list", "--porcelain"], "the worktree list")?
        .lines()
        .filter_map(|l| l.strip_prefix("branch refs/heads/"))
        .map(str::to_string)
        .collect();
    let branches = git(
        &["for-each-ref", "refs/heads", "--format=%(refname:short)"],
        "the branches",
    )?;
    for local in branches
        .lines()
        .filter(|b| !checked_out.iter().any(|c| c == b))
    {
        let ahead = git(
            &["rev-list", "--count", local, "--not", "--remotes"],
            "a branch",
        )?
        .trim()
        .parse::<u64>()
        .map_err(|_| unreadable("a branch"))?;
        if ahead > 0 {
            return Err(format!(
                "the local branch {local} holds {ahead} unpushed commit(s) and no worktree has it \
                 checked out, which a reclaim would orphan"
            ));
        }
    }
    let ignored = git(
        &["status", "--porcelain=v1", "--ignored=traditional"],
        "the ignored files",
    )?;
    for entry in ignored.lines().filter_map(|l| l.strip_prefix("!! ")) {
        if is_regenerable_ignored(entry) {
            continue;
        }
        // git folds a folder that holds nothing but ignored files into one entry (`.moai/` when
        // only its log folder is ignored); look inside it, file by file, before calling it work.
        let draft = if entry.ends_with('/') {
            let files = git(
                &[
                    "ls-files",
                    "--others",
                    "--ignored",
                    "--exclude-standard",
                    "--",
                    entry,
                ],
                "the ignored files",
            )?;
            files
                .lines()
                .find(|f| !is_regenerable_ignored(f))
                .map(str::to_string)
        } else {
            Some(entry.to_string())
        };
        if let Some(draft) = draft {
            return Err(format!(
                "the worktree holds an ignored entry that may be a draft ({draft}), which a \
                 reclaim would orphan"
            ));
        }
    }
    Ok(())
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

/// How long the process check waits for `lsof`: a wedged `lsof` (it can stall on a hung network
/// mount) must make the answer "unproven", never hang `card reclaim`.
#[cfg(target_os = "macos")]
const LSOF_LIMIT: std::time::Duration = std::time::Duration::from_secs(20);

/// The standard output of `cmd`, or `Err` when it cannot start or has not ended within `limit`
/// (it is killed then). The output is read on a thread of its own so a long answer never fills
/// the pipe and looks like a hang; the exit code is not looked at, as `lsof`'s is no answer.
#[cfg(any(target_os = "macos", test))]
fn stdout_within(
    mut cmd: std::process::Command,
    limit: std::time::Duration,
) -> Result<Vec<u8>, String> {
    use std::io::Read;
    let mut child = cmd
        .stdin(std::process::Stdio::null())
        .stdout(std::process::Stdio::piped())
        .stderr(std::process::Stdio::null())
        .spawn()
        .map_err(|e| e.to_string())?;
    let mut pipe = child.stdout.take().ok_or("no output pipe")?;
    let reader = std::thread::spawn(move || {
        let mut bytes = Vec::new();
        let _ = pipe.read_to_end(&mut bytes);
        bytes
    });
    let deadline = std::time::Instant::now() + limit;
    loop {
        match child.try_wait() {
            Ok(Some(_)) => break,
            Ok(None) if std::time::Instant::now() < deadline => {
                std::thread::sleep(std::time::Duration::from_millis(20));
            }
            Ok(None) => {
                let _ = child.kill();
                let _ = child.wait();
                let _ = reader.join();
                return Err(format!("it did not end within {} s", limit.as_secs_f32()));
            }
            Err(e) => return Err(e.to_string()),
        }
    }
    reader
        .join()
        .map_err(|_| "its output could not be read".to_string())
}

/// The working folder of every process this user can see.
#[cfg(target_os = "macos")]
fn working_folders() -> Result<Vec<std::path::PathBuf>, String> {
    // `lsof` has a fixed home on macOS; a child without a PATH must still find it.
    let mut cmd = std::process::Command::new("/usr/sbin/lsof");
    cmd.args(["-a", "-d", "cwd", "-Fn"]);
    let stdout = stdout_within(cmd, LSOF_LIMIT)
        .map_err(|e| format!("the process check could not run lsof: {e}"))?;
    let text = String::from_utf8_lossy(&stdout);
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

    /// N3 of the audit (REQ-4): a command that does not end within the limit is killed and the
    /// answer is an error (unproven), a command that ends in time gives its whole output, even
    /// when it is longer than a pipe holds.
    #[test]
    fn stdout_within_kills_a_command_that_overruns() {
        use std::time::{Duration, Instant};
        let started = Instant::now();
        let mut slow = Command::new("sleep");
        slow.arg("30");
        let err = stdout_within(slow, Duration::from_millis(300)).unwrap_err();
        assert!(err.contains("did not end within"), "{err}");
        assert!(
            started.elapsed() < Duration::from_secs(10),
            "the command was not stopped at the limit"
        );
        let mut echo = Command::new("echo");
        echo.arg("hi");
        assert_eq!(
            stdout_within(echo, Duration::from_secs(20)).unwrap(),
            b"hi\n"
        );
        let mut long = Command::new("head");
        long.args(["-c", "300000", "/dev/zero"]);
        assert_eq!(
            stdout_within(long, Duration::from_secs(20)).unwrap().len(),
            300_000,
            "a long answer filled the pipe"
        );
        let missing = Command::new("/nonexistent/moai-probe");
        assert!(stdout_within(missing, Duration::from_secs(1)).is_err());
    }

    /// REQ-3: only the three named regenerable entries are not drafts.
    #[test]
    fn only_listed_ignored_entries_are_regenerable() {
        for ok in [
            ".moai/logs/",
            "target/",
            "tests/__pycache__/",
            "__pycache__/",
            ".moai/logs/a.log",
            "target/debug/moai",
            "tools/__pycache__/x.pyc",
        ] {
            assert!(is_regenerable_ignored(ok), "{ok}");
        }
        for work in [
            "scratch/",
            "draft.md",
            "target",
            ".moai/",
            ".moai/other.txt",
            ".moai/logs",
            "src/target/",
            "my__pycache__/",
            "notes/__pycache__",
        ] {
            assert!(!is_regenerable_ignored(work), "{work}");
        }
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

    /// AC-10 (REQ-12): only an untracked, listed cache file is a cache. The folder name the
    /// worktree uses comes from its own state root, here a worktree with no marker.
    #[test]
    fn only_an_untracked_listed_file_is_a_cache() {
        let tmp = tempfile::tempdir().unwrap();
        let root = tmp.path();
        let current = crate::state_root::rel(root, "state/github/counts.json");
        let is = |line: &str| is_tool_cache_entry(root, line);
        assert!(is(&format!("?? {current}")));
        assert!(is("?? .moai/state/github/counts.json"));
        // A changed tracked file of the same name is work.
        assert!(!is(&format!(" M {current}")));
        assert!(!is(&format!("A  {current}")));
        // An unlisted name, a folder and a path that only ends alike are work.
        assert!(!is(&format!("?? {}", current.replace("counts", "other"))));
        assert!(!is(&format!(
            "?? {}/",
            crate::state_root::rel(root, "state")
        )));
        assert!(!is(&format!("?? sub/{current}")));
        assert!(!is(&format!("?? {current}.bak")));
        assert!(!is(""));
    }
}
