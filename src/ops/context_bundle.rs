//! `context-bundle`: requested file ranges with digests and a budget; read-only. Moved out of
//! `src/main.rs` unchanged (#2834 R2b).

use std::path::PathBuf;

/// Bundle file ranges with digests
///
/// Read-only: requested file ranges with digests and a budget (typed request on stdin).
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "context.bundle";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::context::bundle(&args.project, &super::super::input()?)
}
