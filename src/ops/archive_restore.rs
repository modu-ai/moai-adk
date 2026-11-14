//! `archive-restore` (D2 #2188): restores one archived SPEC after verifying its manifest. Moved
//! out of `src/main.rs` unchanged (#2834 R2a).

use std::path::PathBuf;

/// Restore one archived SPEC
///
/// Restore one archived SPEC to `.moai-cli/specs/<id>` after verifying its manifest
/// (D2 #2188).
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    #[arg(long)]
    spec: String,
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "archive.restore";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::archive::restore(&args.project, &args.spec, &args.permission_reference)
}
