//! `agents-merge-revert` (#2936 R10): undo one `agents-merge-apply`. The logic is
//! `moai_cli::agents_local_merge`.

use std::path::PathBuf;

/// Undo an AGENTS.local.md merge
///
/// Undo one `agents-merge-apply` (#2936 R10) while `AGENTS.md` still holds the bytes the
/// apply wrote: it gets its earlier bytes back. The source file and its backup are not touched.
#[derive(clap::Args)]
pub struct Args {
    /// The receipt id the apply printed (`agents-local-merge-…`).
    #[arg(long)]
    receipt: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
    /// The project (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "agents_merge.revert";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::agents_local_merge::revert(&args.project, &args.receipt, &args.permission_reference)
}
