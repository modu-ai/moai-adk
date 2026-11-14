//! `snapshot-describe` (07-D #2011): a project's HEAD commit and a digest of its tracked changes;
//! read-only. Moved out of `src/main.rs` unchanged (#2834 R2b).

use std::path::PathBuf;

/// Describe a project's HEAD and tracked changes
///
/// Read-only: PROJECT's HEAD commit and a digest of its tracked changes (07-D #2011).
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "snapshot.describe";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::context::snapshot(&args.project)
}
