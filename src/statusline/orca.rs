//! The Orca pass-through (SPEC-20261122-statusline-limits REQ-8, D-6): when the session runs
//! in an Orca pane, the stdin the status line received is handed to Orca's own status line
//! wrapper, so Orca still sees the usage windows although our command replaced its wrapper.
//!
//! The wrapper is not ours: it is started, never copied. It is started only when it is an
//! executable file at the place Orca installs it, the Orca hook variables are set and the stdin
//! holds `rate_limits`; its output is discarded and the render never waits for it.

use super::state::Context;
use std::io::Write;
use std::path::PathBuf;
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
/// endpoint file variable) or without an executable file at [`SCRIPT`].
pub fn script(ctx: &Context) -> Option<PathBuf> {
    let hooked = HOOK_VARS.iter().all(|k| ctx.var(k).is_some()) || ctx.var(ENDPOINT_VAR).is_some();
    if !hooked {
        return None;
    }
    let path = ctx.home_dir()?.join(SCRIPT);
    executable_file(&path).then_some(path)
}

#[cfg(unix)]
fn executable_file(path: &std::path::Path) -> bool {
    use std::os::unix::fs::PermissionsExt;
    std::fs::metadata(path).is_ok_and(|m| m.is_file() && m.permissions().mode() & 0o111 != 0)
}

#[cfg(not(unix))]
fn executable_file(_path: &std::path::Path) -> bool {
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
