//! `go-hooks-revert` (#2877): undo one `go-hooks-apply`. The logic is `moai_cli::go_hooks`.

use std::path::PathBuf;

/// Undo a Go-era hook replacement
///
/// Undo one `go-hooks-apply` (#2877) while `.claude/settings.json` and its ownership receipt
/// still hold what the apply wrote: both get their earlier bytes back and the scripts move
/// back to `.claude/hooks/moai`. Deletes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The receipt id the apply printed (`go-hooks-…`).
    #[arg(long)]
    receipt: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
    /// The project (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "go_hooks.revert";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::go_hooks::revert(&args.project, &args.receipt, &args.permission_reference)
}
