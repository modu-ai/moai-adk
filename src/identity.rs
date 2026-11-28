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
    cell::RefCell,
    collections::HashMap,
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
///
/// With a `home` that is not `process_home` (the process's own HOME, given by the caller; none
/// counts as another) the child runs with that HOME and without `XDG_CONFIG_HOME` and
/// `GIT_CONFIG_GLOBAL`, so git reads that HOME's `.gitconfig` and `.config/git/config` and neither
/// the environment's files nor the file the environment names. A `home` equal to `process_home`,
/// and no `home` at all, leave the child the process's own environment, so a user's
/// `XDG_CONFIG_HOME` and `GIT_CONFIG_GLOBAL` keep working (SPEC-20261124-home-env-readers D-8;
/// SPEC-20261124-identity-home-arg REQ-32..34; git(1), FILES and ENVIRONMENT). No environment
/// variable named HOME is read here.
fn rev_parse(dir: &Path, args: &[&str], home: Option<&Path>, process_home: Option<&Path>) -> Git {
    let mut cmd = Command::new("git");
    cmd.arg("--no-optional-locks")
        .arg("-C")
        .arg(dir)
        .arg("rev-parse")
        .args(args);
    scrub_git_env(&mut cmd);
    if let Some(home) = home.filter(|h| !h.as_os_str().is_empty()) {
        if process_home != Some(home) {
            cmd.env("HOME", home)
                .env_remove("XDG_CONFIG_HOME")
                .env_remove("GIT_CONFIG_GLOBAL");
        }
    }
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
pub(crate) fn refuse_symlink(path: &Path, what: &str) -> Result<()> {
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

/// The `ok` answers kept inside one [`with_memo`] scope, by the path asked, the HOME given and the
/// process HOME given with it (the last decides the environment of the git children).
type Memo = HashMap<(PathBuf, Option<PathBuf>, Option<PathBuf>), Value>;

thread_local! {
    static MEMO: RefCell<Option<Memo>> = const { RefCell::new(None) };
}

/// Runs `f` with one memory of identity answers (SPEC-20261126-card-next-speed REQ-10, REQ-11):
/// inside it, a later [`identity_in`] for the same path and HOME takes the first `ok` answer
/// instead of asking git again. The checks of the given path and of the home state root still run
/// on every call, nothing that is refused or not `ok` is kept, and the memory ends with `f`. Only
/// for a command that reads: a change `f` makes to the repository or its profile is not seen by a
/// later call. A scope inside a scope uses the outer memory. This thread only.
pub fn with_memo<T>(f: impl FnOnce() -> T) -> T {
    /// Forgets the answers when the outer scope ends, by return or by unwinding.
    struct End;
    impl Drop for End {
        fn drop(&mut self) {
            MEMO.with(|m| *m.borrow_mut() = None);
        }
    }
    if MEMO.with(|m| m.borrow().is_some()) {
        return f();
    }
    MEMO.with(|m| *m.borrow_mut() = Some(Memo::new()));
    let _end = End;
    f()
}

/// Resolves the identity of `path` against the HOME given (SPEC-20261124-identity-home-arg: the
/// HOME is an argument, and every caller reads or passes its own). Never writes. Returns `Err`
/// only for an unusable input path, a refused symlink or a git failure (`git failed: ...`); "not a
/// repository" and "git missing" are typed statuses.
///
/// `home_state_root` is the home state root of `home`, or `null` for `None` (and for an empty
/// HOME, which counts as none), for which no home state root is resolved; no other field depends
/// on it. `home_state_root` is found from no environment variable. The `git` children read the
/// global git configuration of `home` when one is given and, with `None`, keep the process's own
/// lookup, so the environment HOME's configuration reaches them only through the caller that
/// passes no HOME (SPEC-20261124-home-env-readers D-8).
pub fn identity_in(path: &Path, home: Option<&Path>) -> Result<Value> {
    // Slice 2c-1 of SPEC-20261124-identity-home-arg: this function still reads the environment HOME
    // for the comparison, so its answers stay what they were; slice 2c-2 gives `None` here once
    // every caller that gives a HOME has moved to `identity_in_from`.
    identity_in_from(path, home, crate::layout::environment_home().as_deref())
}

/// [`identity_in`] with the process's own HOME given as an argument (SPEC-20261124-identity-home-arg
/// D-17). It reads no environment variable named HOME: `process_home` is the HOME the process runs
/// with, as the caller's command edge read it, and it alone decides the environment of the git
/// children. A `home` equal to `process_home` leaves them the process's own environment (the user's
/// `XDG_CONFIG_HOME` and `GIT_CONFIG_GLOBAL` keep working); a `home` that differs, or no
/// `process_home`, makes `home` the only source of the global git configuration.
pub fn identity_in_from(
    path: &Path,
    home: Option<&Path>,
    process_home: Option<&Path>,
) -> Result<Value> {
    let meta = fs::symlink_metadata(path).map_err(|_| "path unavailable")?;
    if meta.file_type().is_symlink() {
        return Err("path is a symlink; refused".into());
    }
    if !meta.is_dir() {
        return Err("path must be a directory".into());
    }
    // An empty HOME is no HOME: joined with the state folder name it would name a relative path.
    let home = home.filter(|h| !h.as_os_str().is_empty());
    let given = home;
    let home = match home {
        Some(h) => {
            let root = crate::home_root::path(h);
            refuse_symlink(&root, "home state root")?;
            json!(text(&root)?)
        }
        None => Value::Null,
    };
    let key = (
        path.to_path_buf(),
        given.map(Path::to_path_buf),
        process_home.map(Path::to_path_buf),
    );
    if let Some(known) = MEMO.with(|m| m.borrow().as_ref().and_then(|m| m.get(&key).cloned())) {
        return Ok(known);
    }
    let answer = ask_git(path, given, process_home, home)?;
    if answer["identity_status"] == "ok" {
        MEMO.with(|m| {
            if let Some(m) = m.borrow_mut().as_mut() {
                m.insert(key, answer.clone());
            }
        });
    }
    Ok(answer)
}

/// The part of [`identity_in_from`] that asks git, after the checks of the given path and HOME.
/// `given` is the HOME of the caller, `process_home` the process HOME it gave with it and `home`
/// the `home_state_root` field made from `given`.
fn ask_git(
    path: &Path,
    given: Option<&Path>,
    process_home: Option<&Path>,
    home: Value,
) -> Result<Value> {
    let lines = match rev_parse(
        path,
        &[
            "--path-format=absolute",
            "--git-common-dir",
            "--git-dir",
            "--is-bare-repository",
            "--is-inside-work-tree",
        ],
        given,
        process_home,
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
        match rev_parse(
            path,
            &["--path-format=absolute", "--show-toplevel"],
            given,
            process_home,
        ) {
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

/// SPEC-20261126-card-next-speed AC-6: what one [`with_memo`] scope reuses and what it asks again.
/// Whether git was asked again shows in the answer: the repository is broken (or made) between the
/// two calls. No call here is given the environment HOME.
#[cfg(all(test, unix))]
mod tests {
    use super::{identity_in, linked_home_test::symlinked_home, scrub_git_env, with_memo};
    use std::{fs, path::Path, process::Command};

    fn init(dir: &Path) {
        let mut cmd = Command::new("git");
        cmd.arg("-C").arg(dir).args(["init", "-q"]);
        scrub_git_env(&mut cmd);
        assert!(cmd.status().unwrap().success());
    }

    /// Replaces `.git` with a file that names no git folder: git then fails for `dir`, wherever
    /// the temporary folder lies.
    fn break_repository(dir: &Path) {
        fs::rename(dir.join(".git"), dir.join("was-git")).unwrap();
        fs::write(dir.join(".git"), "gitdir: nowhere\n").unwrap();
    }

    fn is_ok(dir: &Path, home: Option<&Path>) -> bool {
        identity_in(dir, home).is_ok_and(|a| a["identity_status"] == "ok")
    }

    #[test]
    fn reuse_of_an_ok_answer_lasts_until_the_scope_ends() {
        let t = tempfile::tempdir().unwrap();
        init(t.path());
        with_memo(|| {
            let first = identity_in(t.path(), None).unwrap();
            assert_eq!(first["identity_status"], "ok");
            break_repository(t.path());
            assert_eq!(identity_in(t.path(), None).unwrap(), first);
            // A scope inside the scope uses the same memory and does not end it.
            with_memo(|| assert!(is_ok(t.path(), None)));
            assert!(is_ok(t.path(), None));
        });
        assert!(!is_ok(t.path(), None), "asked again after the scope");
    }

    #[test]
    fn reuse_needs_a_scope() {
        let t = tempfile::tempdir().unwrap();
        init(t.path());
        assert!(is_ok(t.path(), None));
        break_repository(t.path());
        assert!(!is_ok(t.path(), None));
    }

    #[test]
    fn reuse_is_only_for_the_same_path_and_home() {
        let t = tempfile::tempdir().unwrap();
        let home = tempfile::tempdir().unwrap();
        let (one, other) = (t.path().join("one"), t.path().join("other"));
        for dir in [&one, &other] {
            fs::create_dir(dir).unwrap();
            init(dir);
        }
        with_memo(|| {
            assert!(is_ok(&one, None));
            break_repository(&one);
            break_repository(&other);
            assert!(!is_ok(&other, None), "another path is asked");
            assert!(!is_ok(&one, Some(home.path())), "another HOME is asked");
            assert!(is_ok(&one, None));
        });
    }

    #[test]
    fn reuse_keeps_nothing_that_is_not_ok() {
        let t = tempfile::tempdir().unwrap();
        fs::write(t.path().join(".git"), "gitdir: nowhere\n").unwrap();
        with_memo(|| {
            assert!(!is_ok(t.path(), None));
            fs::remove_file(t.path().join(".git")).unwrap();
            init(t.path());
            assert!(is_ok(t.path(), None), "asked again");
        });
    }

    #[test]
    fn reuse_does_not_skip_the_checks_of_the_path_and_the_home() {
        let t = tempfile::tempdir().unwrap();
        let home = symlinked_home(t.path());
        let link = crate::home_root::path(&home);
        let repo = t.path().join("repo");
        fs::create_dir(&repo).unwrap();
        init(&repo);
        with_memo(|| {
            fs::remove_file(&link).unwrap();
            assert!(is_ok(&repo, Some(&home)));
            // The home state root turns into a link: refused, whatever was answered before.
            std::os::unix::fs::symlink(home.join("elsewhere"), &link).unwrap();
            assert!(identity_in(&repo, Some(&home)).is_err());
            // The path turns into a link.
            assert!(is_ok(&repo, None));
            let moved = t.path().join("moved");
            fs::rename(&repo, &moved).unwrap();
            std::os::unix::fs::symlink(&moved, &repo).unwrap();
            assert!(identity_in(&repo, None).is_err());
        });
    }
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

/// SPEC-20261124-identity-home-arg slice 2a: the shared fixture of a test that proves code which
/// asks `identity_in(path, None)` no longer touches the home state of the environment HOME. The
/// test re-runs itself in a child process whose HOME holds a symlinked home state folder, the one
/// state `identity_in` refuses, so no test changes this process's environment and none reads the
/// real HOME. The child first checks that `identity_in` still refuses that HOME when it is handed
/// over (the refusal where the home state is read stays), then runs the test body.
#[cfg(all(test, unix))]
pub(crate) mod linked_home_test {
    use std::{
        path::{Path, PathBuf},
        process::Command,
    };

    /// Set in the child process; names the project folder the body works on.
    const CHILD_PROJECT: &str = "MOAI_LINKED_HOME_TEST_PROJECT";

    /// The folders a body works on, all beside each other in one temporary folder.
    pub(crate) struct Folders {
        /// A clone with one commit on `main`.
        pub(crate) project: PathBuf,
        /// A linked worktree of `project`.
        pub(crate) linked: PathBuf,
        /// An empty folder for the body's outputs.
        pub(crate) scratch: PathBuf,
    }

    fn folders_of(project: &Path) -> Folders {
        let base = project.parent().unwrap();
        Folders {
            project: project.to_path_buf(),
            linked: base.join("linked"),
            scratch: base.join("scratch"),
        }
    }

    fn git(dir: &Path, args: &[&str]) {
        let mut cmd = Command::new("git");
        cmd.args([
            "-c",
            "user.name=test",
            "-c",
            "user.email=test@example.invalid",
        ])
        .args(args)
        .current_dir(dir);
        super::scrub_git_env(&mut cmd);
        let out = cmd
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_NOSYSTEM", "1")
            .output()
            .unwrap();
        assert!(
            out.status.success(),
            "git {args:?} failed: {}",
            String::from_utf8_lossy(&out.stderr)
        );
    }

    /// A home whose state folder is a symlink to a folder beside it: the refusal case. A Go
    /// generation marker under `.moai` makes the earlier state folder the HOME's state root, and
    /// that is where the link goes; the state root is read back to prove the link is on it.
    pub(crate) fn symlinked_home(dir: &Path) -> PathBuf {
        let home = dir.join("home");
        let go = home.join(crate::home_root::MOAI_DIR);
        std::fs::create_dir_all(&go).unwrap();
        std::fs::write(
            go.join("user-assets.json"),
            br#"{"schema_version": 1, "files": {}}"#,
        )
        .unwrap();
        std::fs::create_dir(home.join("elsewhere")).unwrap();
        let link = crate::home_root::path(&home);
        std::os::unix::fs::symlink(home.join("elsewhere"), &link).unwrap();
        assert_eq!(
            crate::home_root::path(&home),
            link,
            "the link is the state root"
        );
        assert!(std::fs::symlink_metadata(&link).unwrap().is_symlink());
        home
    }

    /// Runs `body` in a child of this test binary whose HOME holds a symlinked home state folder.
    /// `module` is the caller's `module_path!()` and `test` its own function name. The parent
    /// builds the folders and fails unless the child ran exactly that one test and it passed: a
    /// name that matched nothing prints `0 passed`.
    pub(crate) fn run(module: &str, test: &str, body: impl FnOnce(&Folders)) {
        if let Some(project) = std::env::var_os(CHILD_PROJECT) {
            let folders = folders_of(Path::new(&project));
            let env_home = crate::layout::environment_home();
            assert_eq!(
                super::identity_in(&folders.project, env_home.as_deref()).unwrap_err(),
                "home state root is a symlink; refused",
                "where the home state is read, the environment HOME is still refused"
            );
            body(&folders);
            return;
        }
        let dir = tempfile::tempdir().unwrap();
        let base = std::fs::canonicalize(dir.path()).unwrap();
        let project = base.join("project");
        std::fs::create_dir(&project).unwrap();
        git(&project, &["init", "-q", "-b", "main"]);
        git(&project, &["commit", "-q", "--allow-empty", "-m", "base"]);
        let folders = folders_of(&project);
        git(
            &project,
            &[
                "worktree",
                "add",
                "-q",
                "-b",
                "linked-branch",
                folders.linked.to_str().unwrap(),
            ],
        );
        std::fs::create_dir(&folders.scratch).unwrap();
        let home = symlinked_home(&base);
        let inside = module.split_once("::").map_or(module, |(_, rest)| rest);
        let out = Command::new(std::env::current_exe().unwrap())
            .args([
                "--exact",
                &format!("{inside}::{test}"),
                "--nocapture",
                "--test-threads=1",
            ])
            .env("HOME", &home)
            .env(CHILD_PROJECT, &project)
            .output()
            .unwrap();
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
}

/// SPEC-20261124-identity-home-arg: `identity_in` takes the HOME as an argument and reads no
/// environment variable for it. A test that must see the environment HOME run as another value
/// re-runs itself in a child process, so no test changes this process's environment and none
/// reads the real HOME.
#[cfg(all(test, unix))]
mod home_arg_tests {
    use super::identity_in;
    use super::linked_home_test::symlinked_home;
    use std::{
        ffi::OsString,
        path::{Path, PathBuf},
        process::Command,
    };

    /// Set in the child process; names the work folder the child resolves identities for.
    const CHILD_WORK: &str = "MOAI_IDENTITY_TEST_WORK";

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
    /// it would be refused. `identity_in` with that HOME handed over is refused; `identity_in`
    /// with no HOME does not read it and answers `home_state_root: null`.
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
        let env_home = crate::layout::environment_home();
        let refused =
            identity_in(&work, env_home.as_deref()).expect_err("the environment HOME is refused");
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

    /// SPEC-20261124-identity-home-arg AC-12 (REQ-10): an empty HOME is no HOME. Joined with the
    /// state folder name it would name a home state root relative to the current directory.
    #[test]
    fn home_arg_rest_identity_in_treats_an_empty_home_as_none() {
        let work = tempfile::tempdir().unwrap();
        let v = identity_in(work.path(), Some(Path::new(""))).unwrap();
        assert!(v["home_state_root"].is_null(), "{v}");
    }
}

/// SPEC-20261124-home-env-readers (slice 2, D-8): the `git` children of `identity_in` read the
/// global git configuration of the HOME they are given, and with no HOME they keep the process's
/// own lookup. The probe is `safe.directory`: git refuses a repository that another user owns
/// unless the global configuration trusts it. The refusal is forced with git's own test switch,
/// which only the child test process carries. The test re-runs itself in a child process, so no
/// test changes this process's environment and none reads the real HOME.
#[cfg(all(test, unix))]
mod git_config_tests {
    use super::{identity_in, identity_in_from};
    use crate::layout::home_arg_tests::{child_var, init_repo, run_in_child_of};
    use std::{
        ffi::OsString,
        fs,
        path::{Path, PathBuf},
    };

    /// Set in the child process; the repository the child resolves the identity of.
    const CHILD_WORK: &str = "MOAI_IDENTITY_GIT_TEST_WORK";
    /// Set in the child process; a HOME whose `.gitconfig` trusts every repository.
    const CHILD_TRUSTING: &str = "MOAI_IDENTITY_GIT_TEST_TRUSTING_HOME";
    /// Set in the child process; a HOME without any git configuration.
    const CHILD_PLAIN: &str = "MOAI_IDENTITY_GIT_TEST_PLAIN_HOME";

    fn refused(result: crate::Result<serde_json::Value>) -> String {
        let message = result.expect_err("git must refuse the repository");
        assert!(message.contains("dubious ownership"), "{message}");
        message
    }

    /// AC-9 (REQ-10), child side: the repository counts as owned by another user. The answer of
    /// `identity_in(.., Some(home))` follows the `.gitconfig` of `home` whatever the environment
    /// HOME holds; `identity_in(.., None)` follows the environment HOME, as a command's edge
    /// that gives no HOME always did.
    #[test]
    fn identity_in_git_config_follows_the_home_it_is_given() {
        let test = "identity_in_git_config_follows_the_home_it_is_given";
        let (Some(work), Some(trusting), Some(plain)) = (
            child_var(CHILD_WORK),
            child_var(CHILD_TRUSTING),
            child_var(CHILD_PLAIN),
        ) else {
            let dir = tempfile::tempdir().unwrap();
            let work = dir.path().join("work");
            fs::create_dir(&work).unwrap();
            init_repo(&work);
            let trusting = dir.path().join("trusting");
            fs::create_dir(&trusting).unwrap();
            fs::write(trusting.join(".gitconfig"), "[safe]\n\tdirectory = *\n").unwrap();
            let plain = dir.path().join("plain");
            fs::create_dir(&plain).unwrap();
            let xdg = dir.path().join("xdg");
            fs::create_dir(&xdg).unwrap();
            for env_home in [&trusting, &plain] {
                run_in_child_of(
                    module_path!(),
                    test,
                    Some(&OsString::from(env_home.as_os_str())),
                    &[
                        (CHILD_WORK, &work),
                        (CHILD_TRUSTING, &trusting),
                        (CHILD_PLAIN, &plain),
                        ("GIT_TEST_ASSUME_DIFFERENT_OWNER", Path::new("1")),
                        ("GIT_CONFIG_NOSYSTEM", Path::new("1")),
                        ("XDG_CONFIG_HOME", &xdg),
                    ],
                );
            }
            return;
        };
        let environment_trusts =
            std::env::var_os("HOME").map(PathBuf::from).as_deref() == Some(trusting.as_path());
        assert_eq!(
            identity_in(&work, Some(&trusting)).unwrap()["identity_status"],
            "ok"
        );
        refused(identity_in(&work, Some(&plain)));
        if environment_trusts {
            assert_eq!(identity_in(&work, None).unwrap()["identity_status"], "ok");
        } else {
            refused(identity_in(&work, None));
        }
    }

    /// The global git configuration that trusts every repository, in the file `git` reads.
    fn trusting_config(path: &Path) {
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, "[safe]\n\tdirectory = *\n").unwrap();
    }

    /// The variables every child of the tests below carries: git's own switch that makes the
    /// repository count as owned by another user, and no system configuration.
    const OWNER_SWITCHES: [(&str, &str); 2] = [
        ("GIT_TEST_ASSUME_DIFFERENT_OWNER", "1"),
        ("GIT_CONFIG_NOSYSTEM", "1"),
    ];

    /// AC-14 (REQ-10), child side: a given HOME that keeps its trust only in git's second global
    /// file, `<home>/.config/git/config`, is followed like a `.gitconfig` is; the child's own HOME
    /// trusts nothing.
    #[test]
    fn identity_in_reads_the_xdg_git_config_of_the_home_it_is_given() {
        let test = "identity_in_reads_the_xdg_git_config_of_the_home_it_is_given";
        let (Some(work), Some(xdg_only), Some(plain)) = (
            child_var("MOAI_IDENTITY_XDG_TEST_WORK"),
            child_var("MOAI_IDENTITY_XDG_TEST_XDG_HOME"),
            child_var("MOAI_IDENTITY_XDG_TEST_PLAIN_HOME"),
        ) else {
            let dir = tempfile::tempdir().unwrap();
            let work = dir.path().join("work");
            fs::create_dir(&work).unwrap();
            init_repo(&work);
            let xdg_only = dir.path().join("xdg_only");
            trusting_config(&xdg_only.join(".config/git/config"));
            let plain = dir.path().join("plain");
            fs::create_dir(&plain).unwrap();
            let mut extra: Vec<(&str, &Path)> = vec![
                ("MOAI_IDENTITY_XDG_TEST_WORK", &work),
                ("MOAI_IDENTITY_XDG_TEST_XDG_HOME", &xdg_only),
                ("MOAI_IDENTITY_XDG_TEST_PLAIN_HOME", &plain),
            ];
            extra.extend(OWNER_SWITCHES.iter().map(|(k, v)| (*k, Path::new(*v))));
            run_in_child_of(
                module_path!(),
                test,
                Some(&OsString::from(plain.as_os_str())),
                &extra,
            );
            return;
        };
        assert_eq!(
            identity_in(&work, Some(&xdg_only)).unwrap()["identity_status"],
            "ok"
        );
        refused(identity_in(&work, Some(&plain)));
    }

    /// AC-15 (REQ-10), child side: `identity_in` given the environment HOME runs git with the
    /// environment as it is. Trust that
    /// the environment keeps in `$XDG_CONFIG_HOME/git/config`, or in the file `GIT_CONFIG_GLOBAL`
    /// names, counts; with neither the repository is refused (the control).
    #[test]
    fn identity_keeps_the_git_config_of_the_environment() {
        let test = "identity_keeps_the_git_config_of_the_environment";
        let (Some(work), Some(case)) = (
            child_var("MOAI_IDENTITY_ENV_TEST_WORK"),
            child_var("MOAI_IDENTITY_ENV_TEST_CASE"),
        ) else {
            let dir = tempfile::tempdir().unwrap();
            let work = dir.path().join("work");
            fs::create_dir(&work).unwrap();
            init_repo(&work);
            let env_home = dir.path().join("env_home");
            fs::create_dir(&env_home).unwrap();
            let xdg = dir.path().join("xdg");
            trusting_config(&xdg.join("git/config"));
            let global = dir.path().join("trusting.gitconfig");
            trusting_config(&global);
            for (case, extra_var) in [
                ("xdg", Some(("XDG_CONFIG_HOME", &xdg))),
                ("global", Some(("GIT_CONFIG_GLOBAL", &global))),
                ("none", None),
            ] {
                let case_name = PathBuf::from(case);
                let mut extra: Vec<(&str, &Path)> = vec![
                    ("MOAI_IDENTITY_ENV_TEST_WORK", &work),
                    ("MOAI_IDENTITY_ENV_TEST_CASE", &case_name),
                ];
                extra.extend(OWNER_SWITCHES.iter().map(|(k, v)| (*k, Path::new(*v))));
                extra.extend(extra_var.map(|(k, v)| (k, v.as_path())));
                run_in_child_of(
                    module_path!(),
                    test,
                    Some(&OsString::from(env_home.as_os_str())),
                    &extra,
                );
            }
            return;
        };
        // The environment HOME handed over, as a command's edge hands it.
        let env_home = crate::layout::environment_home();
        if case == Path::new("none") {
            refused(identity_in(&work, env_home.as_deref()));
        } else {
            assert_eq!(
                identity_in(&work, env_home.as_deref()).unwrap()["identity_status"],
                "ok",
                "{case:?}"
            );
            // The same HOME given by hand is the environment HOME: nothing is replaced.
            let home = PathBuf::from(std::env::var_os("HOME").unwrap());
            assert_eq!(
                identity_in(&work, Some(&home)).unwrap()["identity_status"],
                "ok",
                "{case:?}"
            );
        }
    }

    /// AC-16 (REQ-10), child side: the environment trusts the repository through every route
    /// (`.gitconfig`, `$XDG_CONFIG_HOME/git/config`, `GIT_CONFIG_GLOBAL`), and a HOME that is
    /// given and is another folder still decides alone: none of them reaches its git child.
    #[test]
    fn identity_in_leaves_the_environment_git_config_out_for_another_home() {
        let test = "identity_in_leaves_the_environment_git_config_out_for_another_home";
        let (Some(work), Some(plain)) = (
            child_var("MOAI_IDENTITY_OTHER_TEST_WORK"),
            child_var("MOAI_IDENTITY_OTHER_TEST_PLAIN_HOME"),
        ) else {
            let dir = tempfile::tempdir().unwrap();
            let work = dir.path().join("work");
            fs::create_dir(&work).unwrap();
            init_repo(&work);
            let env_home = dir.path().join("env_home");
            trusting_config(&env_home.join(".gitconfig"));
            let xdg = dir.path().join("xdg");
            trusting_config(&xdg.join("git/config"));
            let global = dir.path().join("trusting.gitconfig");
            trusting_config(&global);
            let plain = dir.path().join("plain");
            fs::create_dir(&plain).unwrap();
            let mut extra: Vec<(&str, &Path)> = vec![
                ("MOAI_IDENTITY_OTHER_TEST_WORK", &work),
                ("MOAI_IDENTITY_OTHER_TEST_PLAIN_HOME", &plain),
                ("XDG_CONFIG_HOME", &xdg),
                ("GIT_CONFIG_GLOBAL", &global),
            ];
            extra.extend(OWNER_SWITCHES.iter().map(|(k, v)| (*k, Path::new(*v))));
            run_in_child_of(
                module_path!(),
                test,
                Some(&OsString::from(env_home.as_os_str())),
                &extra,
            );
            return;
        };
        // The control: the environment really trusts the repository.
        let env_home = crate::layout::environment_home();
        assert_eq!(
            identity_in(&work, env_home.as_deref()).unwrap()["identity_status"],
            "ok"
        );
        refused(identity_in(&work, Some(&plain)));
    }

    /// The child of the `identity_in_from` tests (SPEC-20261124-identity-home-arg REQ-32..34). Its
    /// HOME is a folder without any git configuration, one environment route (`case`: `xdg` or
    /// `global`) trusts every repository, and git counts the repository as owned by another user.
    /// The parent runs the child once per route and gets `None`; the child gets the repository,
    /// a second HOME without configuration and the route.
    fn from_child(test: &str) -> Option<(PathBuf, PathBuf, PathBuf)> {
        const WORK: &str = "MOAI_IDENTITY_FROM_TEST_WORK";
        const OTHER: &str = "MOAI_IDENTITY_FROM_TEST_OTHER_HOME";
        const CASE: &str = "MOAI_IDENTITY_FROM_TEST_CASE";
        if let (Some(work), Some(other), Some(case)) =
            (child_var(WORK), child_var(OTHER), child_var(CASE))
        {
            return Some((work, other, case));
        }
        let dir = tempfile::tempdir().unwrap();
        let work = dir.path().join("work");
        fs::create_dir(&work).unwrap();
        init_repo(&work);
        let env_home = dir.path().join("env_home");
        fs::create_dir(&env_home).unwrap();
        let other = dir.path().join("other");
        fs::create_dir(&other).unwrap();
        let xdg = dir.path().join("xdg");
        trusting_config(&xdg.join("git/config"));
        let global = dir.path().join("trusting.gitconfig");
        trusting_config(&global);
        for (case, route) in [
            ("xdg", ("XDG_CONFIG_HOME", &xdg)),
            ("global", ("GIT_CONFIG_GLOBAL", &global)),
        ] {
            let case_name = PathBuf::from(case);
            let mut extra: Vec<(&str, &Path)> = vec![
                (WORK, &work),
                (OTHER, &other),
                (CASE, &case_name),
                (route.0, route.1.as_path()),
            ];
            extra.extend(OWNER_SWITCHES.iter().map(|(k, v)| (*k, Path::new(*v))));
            run_in_child_of(
                module_path!(),
                test,
                Some(&OsString::from(env_home.as_os_str())),
                &extra,
            );
        }
        None
    }

    /// AC-25 (REQ-33), child side: a `home` equal to the `process_home` given leaves the git child
    /// the process's own environment, so the trust that `XDG_CONFIG_HOME` or `GIT_CONFIG_GLOBAL`
    /// names counts.
    #[test]
    fn identity_in_from_keeps_the_process_environment_for_the_process_home() {
        let test = "identity_in_from_keeps_the_process_environment_for_the_process_home";
        let Some((work, _other, case)) = from_child(test) else {
            return;
        };
        let home = PathBuf::from(std::env::var_os("HOME").unwrap());
        assert_eq!(
            identity_in_from(&work, Some(&home), Some(&home)).unwrap()["identity_status"],
            "ok",
            "{case:?}"
        );
    }

    /// AC-25 (REQ-32, REQ-34), child side: a `home` that is not the `process_home` given decides
    /// alone, whatever the environment HOME holds: another folder, and the environment HOME
    /// itself when the `process_home` handed over is another folder (no HOME variable is read).
    #[test]
    fn identity_in_from_uses_only_the_given_home_for_another_one() {
        let test = "identity_in_from_uses_only_the_given_home_for_another_one";
        let Some((work, other, case)) = from_child(test) else {
            return;
        };
        let home = PathBuf::from(std::env::var_os("HOME").unwrap());
        refused(identity_in_from(&work, Some(&other), Some(&home)));
        refused(identity_in_from(&work, Some(&home), Some(&other)));
        // The control: the same route trusts the repository when the HOMEs are the same.
        assert_eq!(
            identity_in_from(&work, Some(&home), Some(&home)).unwrap()["identity_status"],
            "ok",
            "{case:?}"
        );
    }

    /// AC-25 (REQ-34), child side: no `process_home` means the `home` given is another HOME.
    #[test]
    fn identity_in_from_without_a_process_home_treats_the_home_as_another_one() {
        let test = "identity_in_from_without_a_process_home_treats_the_home_as_another_one";
        let Some((work, _other, case)) = from_child(test) else {
            return;
        };
        let home = PathBuf::from(std::env::var_os("HOME").unwrap());
        refused(identity_in_from(&work, Some(&home), None));
        assert_eq!(
            identity_in_from(&work, Some(&home), Some(&home)).unwrap()["identity_status"],
            "ok",
            "{case:?}"
        );
    }

    /// AC-27 (REQ-35), child side: `identity_in` answers as `identity_in_from` with no
    /// `process_home`, so a `home` equal to the environment HOME is another HOME to it and its git
    /// child does not keep the environment's `XDG_CONFIG_HOME` or `GIT_CONFIG_GLOBAL`.
    #[test]
    fn identity_in_treats_the_home_it_is_given_as_another_home() {
        let test = "identity_in_treats_the_home_it_is_given_as_another_home";
        let Some((work, _other, case)) = from_child(test) else {
            return;
        };
        let home = PathBuf::from(std::env::var_os("HOME").unwrap());
        refused(identity_in(&work, Some(&home)));
        // The control: the same route trusts the repository when the process HOME is given.
        assert_eq!(
            identity_in_from(&work, Some(&home), Some(&home)).unwrap()["identity_status"],
            "ok",
            "{case:?}"
        );
    }
}
