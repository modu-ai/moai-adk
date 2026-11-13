//! `usage-limit` (#2668 S1): judge the usage record the status line keeps; writes nothing.

use std::path::PathBuf;

/// Read `<HOME>/.moai-cli/state/usage-limits.json`, which the Claude Code status line writes
/// on each render, and report `ok`, `near` (a window at 90% or more), `reached` or `unknown`,
/// with `renew_until` (reset + 15 minutes) for near and reached. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// User home root holding `.moai-cli/state/usage-limits.json`.
    home: PathBuf,
}

pub const OPERATION: &str = "usage.limit";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let now = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map_err(|_| "clock before 1970")?
        .as_secs() as i64;
    Ok(moai_cli::usage_limit::evaluate(&args.home, now))
}
