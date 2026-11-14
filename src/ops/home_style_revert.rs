//! `home-style-revert` (#2875): undo one output style apply.

use std::path::PathBuf;

/// Undo one output style apply
///
/// Undo one output style apply (#2875) for the targets that still hold what it wrote; a target
/// changed since is skipped and reported.
#[derive(clap::Args)]
pub struct Args {
    /// User home root.
    home: PathBuf,
    /// The receipt id a `home-style-apply` returned (`style-…`).
    #[arg(long)]
    receipt: String,
    /// Caller-supplied permission reference (recorded; not an authorization verifier).
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "home.style_revert";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::home_style::revert(&args.home, &args.receipt, &args.permission_reference)
}
