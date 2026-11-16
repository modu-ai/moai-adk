//! `go-settings-revert` (#2936 L2a): undo one `go-settings-apply`. The logic is
//! `moai_cli::go_settings_clean`.

use std::path::PathBuf;

/// Undo a Go-era settings clean-up
///
/// Undo one `go-settings-apply` (#2936 L2a) while `.claude/settings.json` still holds what the
/// apply wrote: the file gets its earlier bytes back. Deletes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The receipt id the apply printed (`go-settings-…`).
    #[arg(long)]
    receipt: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
    /// The project (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "go_settings.revert";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::go_settings_clean::revert(&args.project, &args.receipt, &args.permission_reference)
}
