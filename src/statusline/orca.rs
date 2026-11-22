//! The Orca pass-through (SPEC-20261122-statusline-limits REQ-8, D-6): when the session runs
//! in an Orca pane, the stdin the status line received is handed to Orca's own status line
//! wrapper, so Orca still sees the usage windows although our command replaced its wrapper.
//!
//! The wrapper is not ours: it is started, never copied. It is started only when the Orca hook
//! variables are set, the stdin holds `rate_limits` and the file looks the way Orca installs it:
//! a regular file (no link) with the owner's execute bit, inside `~/.orca/agent-hooks/`, the
//! file and both folders owned by this user, none of them a link and none writable by group or
//! others. Its output is discarded and the render never waits for it (no timeout either: it is
//! not killed, it runs detached and ends on its own).

use super::state::Context;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

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
/// Stdin up to this size is written before the render goes on: a new pipe holds at least this
/// much (16 KiB on macOS, 64 KiB on Linux), so the write never waits for the wrapper. A larger
/// payload is written from a thread; the status line may exit before it finishes, so such a
/// payload can reach the wrapper cut short (the wrapper then posts nothing useful).
const SYNC_WRITE_LIMIT: usize = 16 * 1024;

/// The wrapper to start for this render: `None` without the hook variables (all three, or the
/// endpoint file variable) or when [`SCRIPT`] is not a wrapper this user installed
/// ([`installed_by_user`]).
pub fn script(ctx: &Context) -> Option<PathBuf> {
    let hooked = HOOK_VARS.iter().all(|k| ctx.var(k).is_some()) || ctx.var(ENDPOINT_VAR).is_some();
    if !hooked {
        return None;
    }
    let home = ctx.home_dir()?;
    let path = home.join(SCRIPT);
    installed_by_user(&home, &path).then_some(path)
}

/// The wrapper and the two folders it sits in (`~/.orca`, `~/.orca/agent-hooks`), read without
/// following links: each must be of its kind (folder, regular file; a link is neither), owned by
/// this user and not writable by group or others, and the file must carry the owner's execute
/// bit. Anything else (another user's file, a link to elsewhere, a folder others may write
/// into) is not started: the status line would otherwise run what someone else put there.
#[cfg(unix)]
fn installed_by_user(home: &Path, path: &Path) -> bool {
    use std::os::unix::fs::MetadataExt;
    // SAFETY: geteuid has no preconditions and cannot fail.
    let me = unsafe { libc::geteuid() };
    let ours = |p: &Path, folder: bool, mode_bits: u32| {
        std::fs::symlink_metadata(p).is_ok_and(|m| {
            (if folder { m.is_dir() } else { m.is_file() })
                && m.uid() == me
                && m.mode() & 0o022 == 0
                && m.mode() & mode_bits == mode_bits
        })
    };
    let orca = home.join(".orca");
    ours(&orca, true, 0) && ours(&orca.join("agent-hooks"), true, 0) && ours(path, false, 0o100)
}

#[cfg(not(unix))]
fn installed_by_user(_home: &Path, _path: &Path) -> bool {
    false
}

/// Starts the wrapper with `stdin` when the payload has `rate_limits` and [`script`] names
/// one. Detached in its own process group, environment as the render's, output discarded; no
/// error reaches the line.
pub fn pass_through(stdin: &[u8], has_rate_limits: bool, ctx: &Context) {
    if !has_rate_limits {
        return;
    }
    let Some(path) = script(ctx) else {
        return;
    };
    let mut cmd = Command::new(&path);
    cmd.env_clear()
        .envs(ctx.vars())
        .stdin(Stdio::piped())
        .stdout(Stdio::null())
        .stderr(Stdio::null());
    if let Some(dir) = ctx.cwd.as_ref() {
        cmd.current_dir(dir);
    }
    #[cfg(unix)]
    std::os::unix::process::CommandExt::process_group(&mut cmd, 0);
    let Ok(mut child) = cmd.spawn() else {
        return;
    };
    if let Some(mut pipe) = child.stdin.take() {
        if stdin.len() <= SYNC_WRITE_LIMIT {
            let _ = pipe.write_all(stdin);
        } else {
            let bytes = stdin.to_vec();
            std::thread::spawn(move || {
                let _ = pipe.write_all(&bytes);
            });
        }
    }
    // The child is not waited for: the render returns at once and the wrapper ends on its own.
}
