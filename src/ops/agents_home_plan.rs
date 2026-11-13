//! `agents-home-plan` (#2592 S2): preview of the base agent install; writes nothing.

use std::path::PathBuf;

/// Preview installing the base MoAI agents into HOME's host agent folders (#2592): per file
/// create, update, unchanged, preserved (your edit), conflict or remove_blocked. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// User home root whose `.claude/agents` and `.codex/agents` are inspected.
    home: PathBuf,
    /// Skill package folder carrying `references/agents/`.
    package: PathBuf,
}

pub const OPERATION: &str = "agents.home_plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    serde_json::to_value(moai_cli::agent_install::plan(&args.home, &args.package)?)
        .map_err(|_| "encode failed".into())
}
