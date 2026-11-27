//! `project-update-leftovers-revert`: put archived leftovers back. The logic is
//! `moai_cli::project_update::leftovers_archive`.

use std::path::PathBuf;

/// Put archived leftovers back
///
/// Restores the items of one `project-update-leftovers-apply` from its receipt. The receipt is
/// not trusted: an item whose path is not one of the five leftover paths, is named twice or
/// passes through a link makes the whole revert report a `conflict` and move nothing. Otherwise
/// an item goes back only when its original place is empty and the archived copy's listing
/// digest (a lock's sha256) is the receipt's; any other item is left as it is and reported as
/// `conflict` or `missing`. That judgement is made from the disk, never from the receipt's
/// `moved` flags. Nothing is overwritten and nothing is deleted. The receipt is marked
/// `reverted` only when every item is restored.
#[derive(clap::Args)]
pub struct Args {
    /// The apply's UTC stamp (the `receipt` its result printed).
    #[arg(long)]
    receipt: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
    /// The project (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "project_update.leftovers_revert";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::project_update::leftovers_archive::revert(
        &args.project,
        &args.receipt,
        &args.permission_reference,
    )
}
