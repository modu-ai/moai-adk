//! Read-only clone/workspace identity and state-root resolver (Issue #1988, 01-A).
//!
//! Boundaries come from git itself, never from whether `.git` is a directory or a gitfile:
//! `git rev-parse` reports the common dir (clone boundary) and the git dir (workspace
//! boundary). Nothing is written. The Go generation's `.moai/`, `~/.moai/` and `MOAI_HOME`
//! are never returned as a state root.
//!
//! Provisional decisions (leader RC1, decisions log 18:1x, pending P05 and P12): git runs as a
//! fixed-argv subprocess, ids are sha256 of canonical paths, the home root is fixed at
//! `~/.moai-cli`, and the output keeps the current contract (no envelope, exit 0/1).
use crate::{digest, read_profile};
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

type Result<T> = std::result::Result<T, String>;

/// Upper bound on what one rev-parse call may print; four absolute paths fit easily.
const GIT_OUTPUT_LIMIT: usize = 64 * 1024;
const ID_DERIVATION: &str = "sha256-canonical-path-v0";

/// Caller variables that would redirect git to another repository, inject configuration
/// (`core.bare`, `core.worktree`, ...) or change how it discovers one (Issue #2033).
const GIT_ENV_REMOVED: &[&str] = &[
    "GIT_DIR",
    "GIT_WORK_TREE",
    "GIT_COMMON_DIR",
    "GIT_INDEX_FILE",
    "GIT_NAMESPACE",
    "GIT_CONFIG_PARAMETERS",
    "GIT_CONFIG_COUNT",
    "GIT_CEILING_DIRECTORIES",
    "GIT_DISCOVERY_ACROSS_FILESYSTEM",
];
/// Under `LC_ALL=C` git starts its "no repository here" message with this text.
pub(crate) const NOT_A_REPOSITORY: &str = "not a git repository";

/// The trace2 targets set to `0` on every git child. Removing the variables is not enough: git's
/// own configuration (`trace2.normalTarget` and its siblings) would switch the trace on again and
/// write it to standard error (SPEC-20261122-test-defects-session-reclaim REQ-15).
const GIT_TRACE2_OFF: &[&str] = &["GIT_TRACE2", "GIT_TRACE2_EVENT", "GIT_TRACE2_PERF"];

/// Removes every variable in [`GIT_ENV_REMOVED`], the numbered
/// `GIT_CONFIG_KEY_<n>`/`GIT_CONFIG_VALUE_<n>` pairs and every trace variable (a name starting
/// with `GIT_TRACE`, see [`removed_by_prefix`]) from `cmd`, switches the three trace2 targets off
/// ([`GIT_TRACE2_OFF`]), and fixes the locale so failures can be classified. Shared by every git
/// subprocess (`plan`, Issue #2043).
pub(crate) fn scrub_git_env(cmd: &mut Command) {
    for key in GIT_ENV_REMOVED {
        cmd.env_remove(key);
    }
    // GIT_CONFIG_KEY_<n>/GIT_CONFIG_VALUE_<n> are numbered and the trace variables are many, so
    // they are matched by prefix.
    for (key, _) in std::env::vars_os() {
        if removed_by_prefix(&key.to_string_lossy()) {
            cmd.env_remove(&key);
        }
    }
    for key in GIT_TRACE2_OFF {
        cmd.env(key, "0");
    }
    cmd.env("LC_ALL", "C");
}

/// True for a variable [`scrub_git_env`] removes by its name prefix, not by the fixed list: the
/// numbered config pairs, and every trace variable (`GIT_TRACE`, `GIT_TRACE_*`, `GIT_TRACE2`,
/// `GIT_TRACE2_*`), because a trace writes to git's standard error, which a caller that reads
/// that stream as "git had a problem" cannot tell from a warning
/// (SPEC-20261122-test-defects-session-reclaim REQ-11).
fn removed_by_prefix(name: &str) -> bool {
    name.starts_with("GIT_CONFIG_KEY_")
        || name.starts_with("GIT_CONFIG_VALUE_")
        || name.starts_with("GIT_TRACE")
}

enum Git {
    Unavailable,
    /// git ran and failed; carries a path-free summary of its message.
    Failed(String),
    Lines(Vec<String>),
}

/// First line of git's stderr without its `fatal:`/`error:` prefix, cut before the first
/// quote or path separator so no repository path reaches the caller.
pub(crate) fn failure_summary(stderr: &[u8]) -> String {
    let text = String::from_utf8_lossy(stderr);
    let line = text.lines().next().unwrap_or("").trim();
    let line = line
        .strip_prefix("fatal:")
        .or_else(|| line.strip_prefix("error:"))
        .unwrap_or(line)
        .trim();
    let cut = line.find(['\'', '"', '/', '\\']).unwrap_or(line.len());
    let summary: String = line[..cut].trim_end().chars().take(160).collect();
    if summary.is_empty() {
        "git rev-parse failed".into()
    } else {
        summary
    }
}

/// Runs `git -C <dir> rev-parse <args>` with a fixed argv. Variables that would redirect or
/// reconfigure git are removed and the locale is fixed. stderr is read only to classify a
/// failure; what reaches the caller is `failure_summary`, never a path.
fn rev_parse(dir: &Path, args: &[&str]) -> Git {
    let mut cmd = Command::new("git");
    cmd.arg("--no-optional-locks")
        .arg("-C")
        .arg(dir)
        .arg("rev-parse")
        .args(args);
    scrub_git_env(&mut cmd);
    let out = cmd
        .env("GIT_OPTIONAL_LOCKS", "0")
        .stdin(Stdio::null())
        .stderr(Stdio::piped())
        .output();
    let Ok(out) = out else {
        return Git::Unavailable;
    };
    if !out.status.success() {
        return Git::Failed(failure_summary(&out.stderr));
    }
    if out.stdout.len() > GIT_OUTPUT_LIMIT {
        return Git::Failed("git output exceeded the size limit".into());
    }
    match String::from_utf8(out.stdout) {
        Ok(text) => Git::Lines(text.lines().map(str::to_string).collect()),
        Err(_) => Git::Failed("git output was not UTF-8".into()),
    }
}

/// `no_git` status only when git said so; any other git failure is an error whose message
/// (`git failed: <summary>`) the contract classifies as `git_failed` (Issue #2033).
fn failed(summary: &str, home: Value) -> Result<Value> {
    if summary.starts_with(NOT_A_REPOSITORY) {
        Ok(empty("no_git", "not inside a git repository", home))
    } else {
        Err(format!("git failed: {summary}"))
    }
}

fn canonical(raw: &str) -> Result<PathBuf> {
    fs::canonicalize(raw).map_err(|_| "git reported a path that cannot be resolved".into())
}
fn text(p: &Path) -> Result<String> {
    p.to_str()
        .map(str::to_string)
        .ok_or_else(|| "non UTF-8 path".into())
}
fn refuse_symlink(path: &Path, what: &str) -> Result<()> {
    match fs::symlink_metadata(path) {
        Ok(m) if m.file_type().is_symlink() => Err(format!("{what} is a symlink; refused")),
        _ => Ok(()),
    }
}

// @MX:DEBT Ids are sha256 of the canonical absolute path (leader RC1 provisional, HOLD-1 A).
// @MX:CEILING Moving or renaming the repository changes every id; two clones cannot share one.
// @MX:UPGRADE Re-evaluate when 06-A #2005 consumes workspace identity in the claim journal.
fn path_id(p: &Path) -> Result<String> {
    Ok(digest(text(p)?.as_bytes()))
}

fn empty(status: &str, reason: &str, home: Value) -> Value {
    json!({
        "identity_status": status,
        "reason": reason,
        "workspace_kind": null,
        "clone_anchor": null,
        "workspace_git_dir": null,
        "work_tree": null,
        "clone_id": null,
        "workspace_id": null,
        "project_id": null,
        "profile_status": null,
        "profile_note": null,
        "project_state_root": null,
        "home_state_root": home,
        "id_derivation": ID_DERIVATION,
        "writes": false,
    })
}

/// The process-edge reader of the environment HOME (SPEC-20261124-identity-home-arg): the one
/// place in this module that reads it. A command's edge uses this; code that is not about the
/// HOME calls [`identity_in`] with `None`, and code that is hands it the HOME it was given.
///
/// Resolves the identity of `path`. Never writes. Returns `Err` only for an unusable input
/// path, a refused symlink or a git failure (`git failed: ...`); "not a repository" and "git
/// missing" are typed statuses. An unset or empty HOME counts as no HOME.
pub fn identity(path: &Path) -> Result<Value> {
    let home = std::env::var_os("HOME")
        .filter(|h| !h.is_empty())
        .map(PathBuf::from);
    identity_in(path, home.as_deref())
}

/// [`identity`] with the HOME given instead of read from the environment. `home_state_root` is
/// the home state root of `home`, or `null` for `None`, which opens no file under any HOME; no
/// other field depends on it. Reads no environment variable.
pub fn identity_in(path: &Path, home: Option<&Path>) -> Result<Value> {
    let meta = fs::symlink_metadata(path).map_err(|_| "path unavailable")?;
    if meta.file_type().is_symlink() {
        return Err("path is a symlink; refused".into());
    }
    if !meta.is_dir() {
        return Err("path must be a directory".into());
    }
    let home = match home {
        Some(h) => {
            let root = crate::home_root::path(h);
            refuse_symlink(&root, "home state root")?;
            json!(text(&root)?)
        }
        None => Value::Null,
    };
    let lines = match rev_parse(
        path,
        &[
            "--path-format=absolute",
            "--git-common-dir",
            "--git-dir",
            "--is-bare-repository",
            "--is-inside-work-tree",
        ],
    ) {
        Git::Unavailable => return Ok(empty("git_unavailable", "git was not found on PATH", home)),
        Git::Failed(summary) => return failed(&summary, home),
        Git::Lines(l) if l.len() == 4 => l,
        Git::Lines(_) => return Err("unexpected git rev-parse output".into()),
    };
    let common = canonical(&lines[0])?;
    let git_dir = canonical(&lines[1])?;
    let bare = lines[2] == "true";
    let in_work_tree = lines[3] == "true";
    let work_tree = if in_work_tree {
        match rev_parse(path, &["--path-format=absolute", "--show-toplevel"]) {
            Git::Lines(l) if l.len() == 1 => Some(canonical(&l[0])?),
            Git::Unavailable => {
                return Ok(empty("git_unavailable", "git was not found on PATH", home))
            }
            Git::Failed(summary) => return failed(&summary, home),
            Git::Lines(_) => return Err("git did not report a work tree".into()),
        }
    } else {
        None
    };
    let kind = match (bare, &work_tree) {
        (true, _) => "bare",
        (false, None) => "git_dir",
        (false, Some(_)) if common == git_dir => "primary",
        (false, Some(_)) => "linked",
    };
    let (project_root, project_id, profile_status, profile_note) = match &work_tree {
        Some(top) => {
            let root = crate::state_root::path(top, "");
            refuse_symlink(&root, "project state root")?;
            // The status stays a fixed word (the delta audit's N-c); the file that was actually
            // read and the reason travel beside it in `profile_note` (the delta audit's N2).
            let (id, status, note) = match crate::profile::chosen_profile(top) {
                Err(e) => (
                    Value::Null,
                    "unreadable",
                    Some(format!("{} ({e})", crate::profile::read_rel(top))),
                ),
                Ok(None) => (Value::Null, "missing", None),
                Ok(Some(_)) => match read_profile(top) {
                    Ok((p, _)) => (json!(p.project_id), "present", None),
                    Err(e) => (
                        Value::Null,
                        "unreadable",
                        Some(format!("{} ({e})", crate::profile::read_rel(top))),
                    ),
                },
            };
            (json!(text(&root)?), id, json!(status), note)
        }
        None => (Value::Null, Value::Null, Value::Null, None),
    };
    Ok(json!({
        "identity_status": "ok",
        "reason": null,
        "workspace_kind": kind,
        "clone_anchor": text(&common)?,
        "workspace_git_dir": text(&git_dir)?,
        "work_tree": work_tree.as_deref().map(text).transpose()?,
        "clone_id": path_id(&common)?,
        "workspace_id": path_id(&git_dir)?,
        "project_id": project_id,
        "profile_status": profile_status,
        "profile_note": profile_note,
        "project_state_root": project_root,
        "home_state_root": home,
        "id_derivation": ID_DERIVATION,
        "writes": false,
    }))
}

#[cfg(test)]
mod scrub_tests {
    use super::{removed_by_prefix, scrub_git_env};
    use std::{ffi::OsStr, process::Command};

    /// SPEC-20261122-test-defects-session-reclaim AC-12 (REQ-15, REQ-16): the three trace2 targets are
    /// set to `0` on the child, not left unset, because git's own configuration
    /// (`trace2.normalTarget` and its siblings) would otherwise turn the trace on again.
    #[test]
    fn the_trace2_targets_are_switched_off_for_the_child() {
        let mut cmd = Command::new("git");
        scrub_git_env(&mut cmd);
        let envs: Vec<_> = cmd.get_envs().collect();
        for name in ["GIT_TRACE2", "GIT_TRACE2_EVENT", "GIT_TRACE2_PERF"] {
            assert!(
                envs.contains(&(OsStr::new(name), Some(OsStr::new("0")))),
                "{name} should be 0, got {envs:?}"
            );
        }
    }

    /// SPEC-20261122-test-defects-session-reclaim AC-10 (REQ-11): every trace variable git
    /// documents (`GIT_TRACE`, `GIT_TRACE_*`, `GIT_TRACE2`, `GIT_TRACE2_*`) is removed from a
    /// git child, as are the numbered config pairs; an unrelated variable stays.
    #[test]
    fn git_trace_variables_are_removed_by_prefix() {
        for name in [
            "GIT_TRACE",
            "GIT_TRACE2",
            "GIT_TRACE2_EVENT",
            "GIT_TRACE2_PERF",
            "GIT_TRACE_PERFORMANCE",
            "GIT_TRACE_SETUP",
            "GIT_TRACE_CURL",
            "GIT_CONFIG_KEY_0",
            "GIT_CONFIG_VALUE_12",
        ] {
            assert!(removed_by_prefix(name), "{name} should be removed");
        }
        for name in [
            "GIT_TRACKER",
            "GIT_AUTHOR_NAME",
            "TRACE",
            "PATH",
            "GIT_CONFIG_GLOBAL",
        ] {
            assert!(!removed_by_prefix(name), "{name} should stay");
        }
    }
}

/// SPEC-20261124-identity-home-arg: `identity_in` takes the HOME as an argument and `identity` is
/// the one process-edge reader of the environment HOME. A test that must see the environment HOME
/// run as another value re-runs itself in a child process, so no test changes this process's
/// environment and none reads the real HOME.
#[cfg(all(test, unix))]
mod home_arg_tests {
    use super::{identity, identity_in};
    use std::{
        ffi::OsString,
        path::{Path, PathBuf},
        process::Command,
    };

    /// Set in the child process; names the work folder the child resolves identities for.
    const CHILD_WORK: &str = "MOAI_IDENTITY_TEST_WORK";
    /// Set in the child process of the equality test; the HOME `identity` must have read, absent
    /// when the child's HOME is unset or empty.
    const CHILD_EXPECT: &str = "MOAI_IDENTITY_TEST_EXPECT_HOME";

    /// The libtest name of `test` in this module, as `--exact` wants it (no crate prefix).
    fn libtest_name(test: &str) -> String {
        let module = module_path!();
        let inside = module.split_once("::").map_or(module, |(_, rest)| rest);
        format!("{inside}::{test}")
    }

    /// Re-runs `test` of this module alone in a child process with `HOME` set to `home` (removed
    /// when `None`), the work folder in `CHILD_WORK` and `extra` added. Fails unless the child
    /// really ran that one test and it passed: a name that matched nothing prints `0 passed`.
    fn run_in_child(test: &str, home: Option<&OsString>, work: &Path, extra: &[(&str, &Path)]) {
        let mut cmd = Command::new(std::env::current_exe().unwrap());
        cmd.args([
            "--exact",
            &libtest_name(test),
            "--nocapture",
            "--test-threads=1",
        ])
        .env(CHILD_WORK, work);
        match home {
            Some(h) => cmd.env("HOME", h),
            None => cmd.env_remove("HOME"),
        };
        for (k, v) in extra {
            cmd.env(k, v);
        }
        let out = cmd.output().unwrap();
        let text = format!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
        assert!(out.status.success(), "child failed:\n{text}");
        assert!(
            text.contains("1 passed; 0 failed"),
            "child did not run exactly {test}:\n{text}"
        );
    }

    fn child_work() -> Option<PathBuf> {
        std::env::var_os(CHILD_WORK).map(PathBuf::from)
    }

    /// A home whose state folder is a symlink to a folder beside it: the refusal case.
    fn symlinked_home(dir: &Path) -> PathBuf {
        let home = dir.join("home");
        std::fs::create_dir_all(home.join("elsewhere")).unwrap();
        std::os::unix::fs::symlink(home.join("elsewhere"), home.join(crate::HOME_STATE_DIR))
            .unwrap();
        home
    }

    /// AC-1 (REQ-1, REQ-3): the HOME that is passed decides `home_state_root`, whatever the
    /// environment HOME is.
    #[test]
    fn identity_in_uses_the_home_it_is_given() {
        let work = tempfile::tempdir().unwrap();
        let home = tempfile::tempdir().unwrap();
        let v = identity_in(work.path(), Some(home.path())).unwrap();
        let root = v["home_state_root"].as_str().expect("a home state root");
        assert_eq!(
            Path::new(root),
            crate::home_root::path(home.path()),
            "home_state_root is the state root of the given HOME"
        );
        assert!(Path::new(root).starts_with(home.path()), "{root}");
    }

    /// AC-2 (REQ-2), child side: the environment HOME holds a symlinked state folder, so a read of
    /// it would be refused. `identity` reads it and is refused; `identity_in` with no HOME does
    /// not read it and answers `home_state_root: null`.
    #[test]
    fn identity_in_without_a_home_ignores_the_environment_home() {
        let Some(work) = child_work() else {
            let dir = tempfile::tempdir().unwrap();
            let home = symlinked_home(dir.path());
            let work = dir.path().join("work");
            std::fs::create_dir(&work).unwrap();
            run_in_child(
                "identity_in_without_a_home_ignores_the_environment_home",
                Some(&home.clone().into_os_string()),
                &work,
                &[],
            );
            return;
        };
        let refused = identity(&work).expect_err("identity reads the environment HOME");
        assert_eq!(refused, "home state root is a symlink; refused");
        let v = identity_in(&work, None).expect("no HOME, no refusal");
        assert!(v["home_state_root"].is_null(), "{v}");
    }

    /// AC-3 (REQ-4): a symlinked state folder of the given HOME is refused as before.
    #[test]
    fn identity_in_refuses_a_symlinked_home_state_root() {
        let dir = tempfile::tempdir().unwrap();
        let home = symlinked_home(dir.path());
        let work = dir.path().join("work");
        std::fs::create_dir(&work).unwrap();
        let refused = identity_in(&work, Some(&home)).expect_err("a symlinked state root");
        assert_eq!(refused, "home state root is a symlink; refused");
    }

    /// AC-4 (REQ-5), child side: `identity` answers what `identity_in` answers for the HOME the
    /// parent put in the environment (`CHILD_EXPECT`), or for no HOME when it did not.
    #[test]
    fn identity_equals_identity_in_with_the_environment_home() {
        let Some(work) = child_work() else {
            let dir = tempfile::tempdir().unwrap();
            let work = dir.path().join("work");
            std::fs::create_dir(&work).unwrap();
            let home = dir.path().join("home");
            std::fs::create_dir(&home).unwrap();
            let test = "identity_equals_identity_in_with_the_environment_home";
            run_in_child(
                test,
                Some(&home.clone().into_os_string()),
                &work,
                &[(CHILD_EXPECT, &home)],
            );
            run_in_child(test, Some(&OsString::new()), &work, &[]);
            run_in_child(test, None, &work, &[]);
            return;
        };
        let expect = std::env::var_os(CHILD_EXPECT).map(PathBuf::from);
        assert_eq!(
            identity(&work).unwrap(),
            identity_in(&work, expect.as_deref()).unwrap()
        );
    }
}
