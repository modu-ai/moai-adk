//! `claude-doctor`: read-only check that a Claude Code session in PROJECT reads AGENTS.md.
//! Moved out of `src/main.rs` unchanged (#2834 R3).

use std::path::PathBuf;

/// Check that Claude Code reads AGENTS.md
///
/// Read-only Claude Code check: whether a session started in PROJECT reads AGENTS.md.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "claude.doctor";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::claude::doctor(&args.project)
}
