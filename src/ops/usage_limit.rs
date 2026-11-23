//! `usage-limit` (#2668 S1): judge the usage record the status line keeps; writes nothing.

use std::path::PathBuf;

/// Report how close the usage limits are
///
/// Read the usage record the Claude Code status line keeps for HOME in its cache folder under
/// `TMPDIR` (this process's `TMPDIR`, which must be the status line's), written when a window
/// changes, and report `ok`, `near` (a window at 90% or more), `reached` or `unknown`, with
/// `renew_until` (reset + 15 minutes) for near and reached. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// User home root whose status line keeps the record.
    home: PathBuf,
}

pub const OPERATION: &str = "usage.limit";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let now = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map_err(|_| "clock before 1970")?
        .as_secs() as i64;
    let tmpdir = std::env::var("TMPDIR").ok();
    Ok(moai_cli::usage_limit::evaluate(
        &args.home,
        tmpdir.as_deref(),
        now,
    ))
}
