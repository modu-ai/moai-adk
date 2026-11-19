//! `models-revert` (card C-47 slice 5): undo the launcher settings one `update` saved.

use std::path::PathBuf;

/// Undo the model settings one update saved
///
/// Put the launcher settings file back as the receipt of one `moai update` model save kept it:
/// the text before the save, byte for byte, or no file when there was none. This happens only
/// while the file is still the one that save wrote; a file changed since is left alone and named.
#[derive(clap::Args)]
pub struct Args {
    /// User home root.
    home: PathBuf,
    /// The receipt id the update printed under `models.receipt` (`models-…`).
    #[arg(long)]
    receipt: String,
    /// Caller-supplied permission reference (recorded; not an authorization verifier).
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "models.revert";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::update_models::revert(&args.home, &args.receipt, &args.permission_reference)
}
