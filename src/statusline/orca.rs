//! The Orca pass-through (SPEC-20261122-statusline-limits REQ-8, D-6, D-10): when the session
//! runs in an Orca pane, the stdin the status line received is handed to Orca's own status line
//! wrapper, so Orca still sees the usage windows although our command replaced its wrapper.
//!
//! The wrapper is not ours: it is started, never copied. It is started only when the Orca hook
//! variables are set, the stdin holds `rate_limits`, the payload fits a pipe, the file looks the
//! way Orca installs it (a regular file, no link, with the owner's execute bit, inside
//! `~/.orca/agent-hooks/`, the file and both folders owned by this user, none of them a link and
//! none writable by group or others) and no earlier wrapper of this user is still running.
//!
//! One at a time, for a bounded time: the wrapper holds a lock on a file in the status line's
//! private cache folder for as long as it (or anything it starts) runs. The lock is an `flock`
//! the render takes and hands to the wrapper as an open descriptor, so the kernel releases it
//! when the last holder ends: no pid file, nothing to clean up. A render that finds the lock
//! held starts nothing. The wrapper gets a `SIGALRM` after [`TIME_LIMIT_SECS`] (an alarm
//! survives `exec`), so a wrapper that hangs ends and the next render can start one again.
//!
//! The wrapper gets only the variables it reads (the hook variables, `CLAUDE_JOB_DIR`,
//! `CLAUDE_CONFIG_DIR`, `HOME`, `PATH`, `TMPDIR` and the locale), not the whole environment of
//! the session. Its output is discarded and the render never waits for it.

use super::state::Context;
use std::path::{Path, PathBuf};

/// The wrapper Orca installs, below HOME.
pub const SCRIPT: &str = ".orca/agent-hooks/claude-statusline.sh";
/// The three hook variables the wrapper needs.
const HOOK_VARS: [&str; 3] = [
    "ORCA_AGENT_HOOK_PORT",
    "ORCA_AGENT_HOOK_TOKEN",
    "ORCA_PANE_KEY",
];
/// The variable naming a file the wrapper reads the hook variables from instead.
const ENDPOINT_VAR: &str = "ORCA_AGENT_HOOK_ENDPOINT";
/// The variables the wrapper reads besides the hook variables (Orca's
/// `claude-statusline.sh`: the job folder makes it post nothing, the config folder, the
/// hook's environment name and version are posted) and what any shell script needs to run.
const PASSED_VARS: [&str; 7] = [
    "ORCA_AGENT_HOOK_ENV",
    "ORCA_AGENT_HOOK_VERSION",
    "CLAUDE_JOB_DIR",
    "CLAUDE_CONFIG_DIR",
    "HOME",
    "PATH",
    "TMPDIR",
];
/// The largest stdin passed on: a new pipe holds at least this much (16 KiB on macOS, 64 KiB
/// on Linux), so the whole payload is written before the render goes on and the write never
/// waits for the wrapper. A larger payload starts nothing (a cut payload would post nothing
/// useful).
const PAYLOAD_LIMIT: usize = 16 * 1024;
/// Seconds a wrapper may run before its `SIGALRM`; Orca's own posts give up after 1.5 s.
pub const TIME_LIMIT_SECS: u32 = 10;
/// The lock file in the cache folder that a running wrapper holds.
const LOCK_FILE: &str = "orca-wrapper.lock";

/// The wrapper to start for this render: `None` without the hook variables (all three, or the
/// endpoint file variable) or when [`SCRIPT`] is not a wrapper this user installed
/// ([`installed_by`]).
pub fn script(ctx: &Context) -> Option<PathBuf> {
    let hooked = HOOK_VARS.iter().all(|k| ctx.var(k).is_some()) || ctx.var(ENDPOINT_VAR).is_some();
    if !hooked {
        return None;
    }
    let home = ctx.home_dir()?;
    let path = home.join(SCRIPT);
    installed_by(&home, &path, effective_uid()?).then_some(path)
}

#[cfg(unix)]
fn effective_uid() -> Option<u32> {
    // SAFETY: geteuid has no preconditions and cannot fail.
    Some(unsafe { libc::geteuid() })
}

#[cfg(not(unix))]
fn effective_uid() -> Option<u32> {
    None
}

/// The wrapper and the two folders it sits in (`~/.orca`, `~/.orca/agent-hooks`), read without
/// following links: each must be of its kind (folder, regular file; a link is neither), owned by
/// `uid` and not writable by group or others, and the file must carry the owner's execute bit.
/// Anything else (another user's file, a link to elsewhere, a folder others may write into) is
/// not started: the status line would otherwise run what someone else put there.
#[cfg(unix)]
pub fn installed_by(home: &Path, path: &Path, uid: u32) -> bool {
    use std::os::unix::fs::MetadataExt;
    let ours = |p: &Path, folder: bool, mode_bits: u32| {
        std::fs::symlink_metadata(p).is_ok_and(|m| {
            (if folder { m.is_dir() } else { m.is_file() })
                && m.uid() == uid
                && m.mode() & 0o022 == 0
                && m.mode() & mode_bits == mode_bits
        })
    };
    let orca = home.join(".orca");
    ours(&orca, true, 0) && ours(&orca.join("agent-hooks"), true, 0) && ours(path, false, 0o100)
}

#[cfg(not(unix))]
pub fn installed_by(_home: &Path, _path: &Path, _uid: u32) -> bool {
    false
}

/// The environment the wrapper gets: the hook variables, [`PASSED_VARS`] and the locale
/// (`LANG`, `LC_*`) as this render has them, nothing else.
pub fn wrapper_env(ctx: &Context) -> Vec<(String, String)> {
    ctx.vars()
        .filter(|(k, _)| {
            HOOK_VARS.contains(&k.as_str())
                || k.as_str() == ENDPOINT_VAR
                || PASSED_VARS.contains(&k.as_str())
                || k.as_str() == "LANG"
                || k.starts_with("LC_")
        })
        .map(|(k, v)| (k.clone(), v.clone()))
        .collect()
}

/// Starts the wrapper with `stdin` when the payload has `rate_limits`, fits a pipe, [`script`]
/// names one and no earlier wrapper holds the lock. Detached in its own process group with the
/// lock and a time limit; no error reaches the line.
pub fn pass_through(stdin: &[u8], has_rate_limits: bool, ctx: &Context) {
    if !has_rate_limits || stdin.len() > PAYLOAD_LIMIT {
        return;
    }
    let Some(path) = script(ctx) else {
        return;
    };
    #[cfg(unix)]
    start(&path, stdin, ctx);
    // No system without Unix permissions gets here: `script` names no wrapper there.
    #[cfg(not(unix))]
    let _ = (path, stdin);
}

/// The lock file, opened and locked for this start: `None` when there is no private cache
/// folder, the file cannot be opened as a plain file of its own, or another wrapper holds it.
#[cfg(unix)]
fn take_lock(ctx: &Context) -> Option<std::fs::File> {
    use std::os::unix::fs::OpenOptionsExt;
    use std::os::unix::io::AsRawFd;
    let dir = super::state::cache_folder(ctx.var("TMPDIR"))?;
    if !super::state::private_folder(&dir) {
        return None;
    }
    let file = std::fs::OpenOptions::new()
        .create(true)
        .append(true)
        .mode(0o600)
        .custom_flags(libc::O_NOFOLLOW)
        .open(dir.join(LOCK_FILE))
        .ok()?;
    // SAFETY: flock on a descriptor this function owns; no memory is touched.
    let locked = unsafe { libc::flock(file.as_raw_fd(), libc::LOCK_EX | libc::LOCK_NB) } == 0;
    locked.then_some(file)
}

#[cfg(unix)]
fn start(path: &Path, stdin: &[u8], ctx: &Context) {
    use std::io::Write;
    use std::os::unix::io::AsRawFd;
    use std::os::unix::process::CommandExt;
    use std::process::{Command, Stdio};
    let Some(lock) = take_lock(ctx) else {
        return;
    };
    let fd = lock.as_raw_fd();
    let mut cmd = Command::new(path);
    cmd.env_clear()
        .envs(wrapper_env(ctx))
        .stdin(Stdio::piped())
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .process_group(0);
    if let Some(dir) = ctx.cwd.as_ref() {
        cmd.current_dir(dir);
    }
    // SAFETY: the closure runs in the child between fork and exec and calls only
    // async-signal-safe functions (fcntl, signal, alarm) on values copied in.
    unsafe {
        cmd.pre_exec(move || {
            // The lock descriptor stays open across exec, so the wrapper holds the lock.
            if libc::fcntl(fd, libc::F_SETFD, 0) == -1 {
                return Err(std::io::Error::last_os_error());
            }
            // The default action of SIGALRM ends the wrapper; an ignored one would not.
            libc::signal(libc::SIGALRM, libc::SIG_DFL);
            libc::alarm(TIME_LIMIT_SECS);
            Ok(())
        });
    }
    let Ok(mut child) = cmd.spawn() else {
        return;
    };
    // The render's copy closes here; the wrapper's copy keeps the lock until it ends.
    drop(lock);
    if let Some(mut pipe) = child.stdin.take() {
        let _ = pipe.write_all(stdin);
    }
    // The child is not waited for: the render returns at once and the wrapper ends on its own,
    // at the latest at its alarm.
}

#[cfg(test)]
mod tests {
    use super::*;

    /// The owner check compares with the uid it is given: the same tree is a wrapper of its
    /// owner and of nobody else.
    #[cfg(unix)]
    #[test]
    fn the_wrapper_counts_as_installed_only_by_its_owner() {
        use std::os::unix::fs::PermissionsExt;
        let tmp = tempfile::tempdir().unwrap();
        let home = tmp.path();
        let hooks = home.join(".orca/agent-hooks");
        std::fs::create_dir_all(&hooks).unwrap();
        for dir in [hooks.as_path(), &home.join(".orca")] {
            std::fs::set_permissions(dir, std::fs::Permissions::from_mode(0o755)).unwrap();
        }
        let path = home.join(SCRIPT);
        std::fs::write(&path, "#!/bin/sh\n").unwrap();
        std::fs::set_permissions(&path, std::fs::Permissions::from_mode(0o755)).unwrap();
        let me = effective_uid().unwrap();
        assert!(installed_by(home, &path, me));
        assert!(!installed_by(home, &path, me.wrapping_add(1)));
    }

    /// Only the hook variables, the variables the wrapper reads and the locale pass.
    #[test]
    fn the_wrapper_gets_only_the_variables_it_reads() {
        let ctx = Context::for_test(
            &[
                ("ORCA_AGENT_HOOK_TOKEN", "t"),
                ("ORCA_AGENT_HOOK_PORT", "1"),
                ("ORCA_PANE_KEY", "p"),
                ("ORCA_AGENT_HOOK_ENDPOINT", "/e"),
                ("CLAUDE_CONFIG_DIR", "/c"),
                ("HOME", "/h"),
                ("PATH", "/bin"),
                ("LC_ALL", "C"),
                ("LANG", "C"),
                ("GITHUB_TOKEN", "secret"),
                ("ANTHROPIC_API_KEY", "secret"),
                ("ORCA_OTHER", "x"),
            ],
            None,
        );
        let mut names: Vec<String> = wrapper_env(&ctx).into_iter().map(|(k, _)| k).collect();
        names.sort();
        assert_eq!(
            names,
            [
                "CLAUDE_CONFIG_DIR",
                "HOME",
                "LANG",
                "LC_ALL",
                "ORCA_AGENT_HOOK_ENDPOINT",
                "ORCA_AGENT_HOOK_PORT",
                "ORCA_AGENT_HOOK_TOKEN",
                "ORCA_PANE_KEY",
                "PATH"
            ]
        );
    }
}
