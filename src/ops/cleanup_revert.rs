//! `cleanup-revert` (#2872 stage 1): move one cleanup's archive back. The logic is
//! `moai_cli::update_cleanup`.

use std::path::PathBuf;

/// Undo one HOME cleanup
///
/// Move every archived item of the cleanup `--stamp` back to its original path (#2872). Any
/// `conflict` or `missing` item refuses the whole revert before moving anything; items already
/// back are skipped, so a rerun is safe. Deletes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// User home root.
    home: PathBuf,
    /// The archive folder name the apply printed (`YYYYMMDDTHHMMSSZ` or `….N`).
    #[arg(long)]
    stamp: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "cleanup.revert";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::update_cleanup::revert(&args.home, &args.stamp, &args.permission_reference)
}
