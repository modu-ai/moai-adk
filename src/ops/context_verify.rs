//! `context-verify`: re-checks a bundle's ranges against the files now; read-only. Moved out of
//! `src/main.rs` unchanged (#2834 R2b).

use std::path::PathBuf;

/// Re-check a bundle's ranges against the files
///
/// Read-only: re-check a bundle's ranges against the files now (typed request on stdin).
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "context.verify";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::context::verify(&args.project, &super::super::input()?)
}
