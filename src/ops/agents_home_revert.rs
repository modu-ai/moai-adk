//! `agents-home-revert` (#2592 S2): undo one base agent install.

use std::path::PathBuf;

/// Undo one base agent install
///
/// Undo one base agent install (#2592) for the files that still hold the bytes it wrote: a
/// created file is removed, an updated one gets its before bytes back; a changed file is skipped.
#[derive(clap::Args)]
pub struct Args {
    /// User home root.
    home: PathBuf,
    /// The receipt id an `agents-home-apply` returned.
    #[arg(long)]
    receipt: String,
    /// Caller-supplied permission reference (recorded; not an authorization verifier).
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "agents.home_revert";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::agent_install::revert(&args.home, &args.receipt, &args.permission_reference)
}
