//! `cleanup-revert-plan` (#2872 stage 1): preview moving one cleanup's archive back. Reads
//! only; the logic is `moai_cli::update_cleanup`.

use std::path::PathBuf;

/// Preview undoing one HOME cleanup
///
/// Read the receipt of the cleanup `--stamp` and list each item (#2872): `restore` (archived,
/// unchanged, original free), `done` (already back), `conflict` (something is at the original
/// path) or `missing` (archived copy gone or changed). Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// User home root.
    home: PathBuf,
    /// The archive folder name the apply printed (`YYYYMMDDTHHMMSSZ` or `….N`).
    #[arg(long)]
    stamp: String,
}

pub const OPERATION: &str = "cleanup.revert_plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::update_cleanup::revert_plan(&args.home, &args.stamp)
}
