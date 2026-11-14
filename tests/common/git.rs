//! The shared git fixture runner (#2352). Every test-side `git` invocation goes
//! through here so fixture repositories behave the same on every machine:
//!
//! - background maintenance is off (`gc.auto=0`, `maintenance.auto=false`) — a
//!   concurrent `git gc` can repack objects out from under a running assertion
//!   (#2082/#2092);
//! - the runner's global/system git config is ignored (`GIT_CONFIG_GLOBAL`/`SYSTEM`
//!   pointed at `/dev/null`), so user settings like `init.defaultBranch` cannot
//!   change a fixture's branch names;
//! - `LC_ALL=C` keeps git messages in English for stderr assertions.
//!
//! Identity is applied per fixture (the fixture layers add `-c user.name=…` args or
//! their own env), not here, because the test suites use two different identities.

use std::path::Path;
use std::process::{Command, Stdio};

/// Guard + cleanup args shared by every fixture git call.
fn base_args() -> [&'static str; 4] {
    ["-c", "gc.auto=0", "-c", "maintenance.auto=false"]
}

/// Starts a guarded git command without choosing its directory, identity or stdio.
/// Callers keep their existing command chain and output/assertion semantics.
pub fn command() -> Command {
    let mut command = Command::new("git");
    command
        .args(base_args())
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("LC_ALL", "C");
    command
}

/// The guarded result of one git call: success flag plus lossy text streams.
pub struct GitOut {
    pub success: bool,
    pub stdout: String,
    pub stderr: String,
}

/// Runs git with the fixture guards and returns the raw outcome. No assertion —
/// for call sites that inspect the outcome themselves (tolerant probes, clones
/// whose failure is reported with context).
#[allow(dead_code)] // adopted incrementally as test files convert to the helper
pub fn git_raw(dir: &Path, args: &[&str]) -> GitOut {
    let out = command()
        .args(args)
        .current_dir(dir)
        .stdin(Stdio::null())
        .output()
        .expect("spawn git");
    GitOut {
        success: out.status.success(),
        stdout: String::from_utf8_lossy(&out.stdout).to_string(),
        stderr: String::from_utf8_lossy(&out.stderr).to_string(),
    }
}

/// Runs git with the fixture guards and returns the trimmed stdout; panics with the
/// stderr when the call fails.
#[allow(dead_code)] // adopted incrementally as test files convert to the helper
pub fn git(dir: &Path, args: &[&str]) -> String {
    let out = git_raw(dir, args);
    assert!(out.success, "git {args:?} failed: {}", out.stderr);
    out.stdout.trim().to_string()
}

/// Runs git with the fixture guards and asserts success; for call sites that do not
/// need the output.
#[allow(dead_code)] // adopted incrementally as test files convert to the helper
pub fn git_ok(dir: &Path, args: &[&str]) {
    let out = git_raw(dir, args);
    assert!(out.success, "git {args:?} failed: {}", out.stderr);
}

/// Runs git with the fixture guards and one pinned commit date pair (`GIT_AUTHOR_DATE`
/// / `GIT_COMMITTER_DATE` set to `<date> +0000`); asserts success. Deterministic-hash
/// fixtures (#2187 archive family) use this so their object ids stay stable.
#[allow(dead_code)] // adopted incrementally as test files convert to the helper
pub fn git_pinned(dir: &Path, args: &[&str], date: Option<&str>) {
    let mut cmd = command();
    cmd.args(args).current_dir(dir).stdin(Stdio::null());
    if let Some(d) = date {
        cmd.env("GIT_AUTHOR_DATE", format!("{d} +0000"))
            .env("GIT_COMMITTER_DATE", format!("{d} +0000"));
    }
    let out = cmd.output().expect("spawn git");
    assert!(
        out.status.success(),
        "git {args:?} failed: {}",
        String::from_utf8_lossy(&out.stderr)
    );
}

/// Runs git in `dir` with the fixture identity and no signing, autocrlf, excludes file or
/// background maintenance; asserts success. Moved from `tests/product_cli.rs` unchanged
/// (#2834 R1) so the group test files share it.
#[allow(dead_code)] // only the product CLI group files use it
pub fn git_in(dir: &std::path::Path, args: &[&str]) {
    let out = std::process::Command::new("git")
        .args([
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
        ])
        .args(["-c", "commit.gpgsign=false", "-c", "core.autocrlf=false"])
        .args(["-c", "core.excludesFile=/dev/null"])
        .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
        .arg("-C")
        .arg(dir)
        .args(args)
        .output()
        .unwrap();
    assert!(out.status.success(), "git {args:?}: {:?}", out.stderr);
}
