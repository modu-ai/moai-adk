//! `atlas-tag-scan` (C2 #2178): read-only scan of `@MOAI:` / `@MX:` tag markers. Moved out of
//! `src/main.rs` unchanged (#2834 R2c).

use std::path::PathBuf;

/// Scan for MOAI and MX tag markers
///
/// Read-only dual-read scan of `@MOAI:` / `@MX:` tag markers (C2 #2178): both
/// namespaces normalize into one node shape, duplicates are diagnosed only,
/// nothing is written. PATH is a file or a directory.
#[derive(clap::Args)]
pub struct Args {
    path: PathBuf,
}

pub const OPERATION: &str = "atlas_tag.scan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::atlas_tag::scan_path(&args.path)
}
