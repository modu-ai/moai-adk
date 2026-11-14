//! `identity`: read-only clone/workspace identity and state roots of a path. Moved out of
//! `src/main.rs` unchanged (#2834 R3).

use std::path::PathBuf;

/// Show where a path sits in its clone
///
/// Read-only clone/workspace identity and state roots of a path (typed JSON on stdout).
#[derive(clap::Args)]
pub struct Args {
    path: PathBuf,
}

pub const OPERATION: &str = "identity";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::identity::identity(&args.path)
}
