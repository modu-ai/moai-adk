//! `archive-search` (D2 #2188): read-only search over the archive index and archived text.
//! Moved out of `src/main.rs` unchanged (#2834 R2a).

use std::path::PathBuf;

/// Search the archived SPECs
///
/// Read-only, non-vector search over the archive index and archived text (D2 #2188).
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    #[arg(long)]
    query: String,
}

pub const OPERATION: &str = "archive.search";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::archive::search(&args.project, &args.query)
}
