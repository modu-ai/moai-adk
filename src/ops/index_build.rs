//! `index-build` (C6 #2182): builds an incremental-index generation. Moved out of `src/main.rs` unchanged (#2834 R4b).

use std::path::PathBuf;

/// Build the code map index
///
/// Build an incremental-index generation (C6 #2182): scan the tracked files, stage an
/// immutable manifest directory, and publish it with an atomic pointer swap. A failed
/// build keeps the previous generation and marks partial (REQ-C6-01).
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "index.build";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    // Stage lines on a terminal's stderr, as `shows_progress` turned on for it in main.rs.
    moai_cli::progress::install_stderr();
    serde_json::to_value(moai_cli::index::index_build(
        &args.project,
        &super::super::input::<moai_cli::index::IndexBuildRequest>()?,
    )?)
    .map_err(|_| "encode failed".into())
}
