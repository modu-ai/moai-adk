//! `project-update-leftovers-apply`: archive leftovers of an earlier generation into a backup.
//! The logic is `moai_cli::project_update::leftovers_archive`.

use std::path::PathBuf;

/// Archive leftovers of an earlier generation into a backup
///
/// Moves each chosen leftover that `project-update-plan` listed with an `action` (the earlier
/// state folder, a folder an earlier revert set aside, the Go backups folder, a Go update lock
/// whose process is not running) into `<state folder>/backups/<UTC>/leftovers/<same path>` by one
/// rename (a folder) or one link and unlink (a lock), and deletes nothing. The `plan_digest` must
/// be the one a fresh preview prints now: a leftover that changed, a lock whose process started,
/// an item the plan does not list, an absolute path, a `..`, a link, a reference the contract
/// refuses and a backup area on another filesystem all stop the command before it writes
/// anything. A receipt `leftovers-receipt.json` is saved as `pending` and read back before the
/// first move, and marked `committed` only after every moved item reads back. The result counts
/// the moved files git tracks (a checkout brings them back unless the deletions are committed)
/// and the secret-shaped file names inside moved folders, which are never opened.
/// `project-update-leftovers-revert` puts the items back.
#[derive(clap::Args)]
pub struct Args {
    /// The `plan_digest` of the leftovers in the confirmed `project-update-plan`.
    #[arg(long)]
    plan_digest: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
    /// A leftover path exactly as the plan prints it (a folder ends in `/`); repeat for more.
    #[arg(long = "item", required = true)]
    item: Vec<String>,
    /// The project (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "project_update.leftovers_apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::project_update::leftovers_archive::apply(
        &args.project,
        &args.plan_digest,
        &args.permission_reference,
        &args.item,
    )
}
