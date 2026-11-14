//! `hooks-revert` (#2853): undoes the hook layer one confirmed update applied. Moved out of `src/main.rs` unchanged (#2834 R4b).

use std::path::PathBuf;

/// Undo the hook set one update applied
///
/// Undo the hook layer one confirmed `update` applied in PROJECT (#2853), from its receipt:
/// a host's files get their earlier bytes back only while all of them still hold what the
/// update wrote; a host changed since is skipped and named.
#[derive(clap::Args)]
pub struct Args {
    /// The project the update ran in.
    project: PathBuf,
    /// The receipt id the update printed under `hooks_result.receipt`.
    #[arg(long)]
    receipt: String,
    /// Caller-supplied permission reference (recorded; not an authorization verifier).
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "hooks.revert";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::update_hooks::revert(&args.project, &args.receipt, &args.permission_reference)
}
