//! `index-query` (C6 #2182): read-only state of the incremental index. Moved out of
//! `src/main.rs` unchanged (#2834 R2c).

use std::path::PathBuf;

/// Query the code map index
///
/// Query the incremental index (C6 #2182): read-only — resolves the published
/// generation and reports one of fresh/stale/partial/unavailable with the detected
/// changes (additions, deletions, modifications, untracked new files, branch switch).
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "index.query";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    serde_json::to_value(moai_cli::index::index_query(
        &args.project,
        &super::super::input::<moai_cli::index::IndexQueryRequest>()?,
    )?)
    .map_err(|_| "encode failed".into())
}
