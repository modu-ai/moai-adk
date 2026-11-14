//! `archive-plan` (D1 #2187): which SPECs can be archived; writes nothing. Moved out of
//! `src/main.rs` unchanged (#2834 R2a).

use std::path::PathBuf;

/// Preview which SPECs can be archived
///
/// Dry-run (D1 #2187, REQ-D1-01): which SPEC directories under
/// `.moai-cli/specs/` are archive-eligible and why the rest are excluded.
/// Writes nothing. `--grace-days` overrides the profile's inactivity window.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    #[arg(long)]
    grace_days: Option<u64>,
}

pub const OPERATION: &str = "archive.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    serde_json::to_value(moai_cli::archive::plan(
        &args.project,
        std::time::SystemTime::now(),
        args.grace_days,
    )?)
    .map_err(|_| "encode failed".into())
}
