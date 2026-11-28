//! `claude-md-revert`: undo one `claude-md-apply`. The logic is
//! `moai_cli::project_update::claude_md`.

use std::path::PathBuf;

/// Undo a replacement of the instruction files
///
/// Undo one `claude-md-apply` from its receipt, also one that failed or was stopped halfway.
/// A new file is removed only while it holds the bytes the apply wrote and its original can come
/// back. Every moved file returns from the backup folder with its bytes and mode, and only into
/// an empty place. The line the apply added to `.gitignore` is taken back (and a `.gitignore`
/// the apply created removed) only while that file holds what the apply left and
/// `CLAUDE.local.md` is absent or is the original again, as git showed it before the apply.
///
/// Whatever cannot be undone is left as it is and named under `conflicts` or `missing`, each with
/// its reason; the original of a conflict stays in the backup folder, and the same command runs
/// again after the cause is cleared. No link is followed. A receipt that no apply of this project
/// writes is refused with nothing written. Settings apply at once; running sessions and lanes read
/// changed instructions only after a restart.
#[derive(clap::Args)]
pub struct Args {
    /// The receipt id the apply answered (`YYYYMMDDTHHMMSSZ`).
    #[arg(long)]
    receipt: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
    /// The project (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "claude_md.revert";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::project_update::claude_md::revert(
        &args.project,
        &args.receipt,
        &args.permission_reference,
    )
}
