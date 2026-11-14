//! `keys-revert` (#2871 I2 S3): undo the keys layer one `update` applied.

use std::path::PathBuf;

/// Undo the keys one update imported
///
/// Undo the keys layer one confirmed `update` applied (#2871): a vault entry it imported is
/// removed only while it still holds what the update wrote; an entry changed since is skipped
/// and named. The original key file is never touched.
#[derive(clap::Args)]
pub struct Args {
    /// User home root.
    home: PathBuf,
    /// The receipt id the update printed under `keys_result.receipt` (`keys-…`).
    #[arg(long)]
    receipt: String,
    /// Caller-supplied permission reference (recorded; not an authorization verifier).
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "keys.revert";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::update_keys::revert(&args.home, &args.receipt, &args.permission_reference)
}
