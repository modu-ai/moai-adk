//! `home-style-plan` (#2875): preview of the MoAI-Easy output style in HOME; writes nothing.

use std::path::PathBuf;

/// Preview the MoAI-Easy output style in HOME
///
/// Preview the MoAI-Easy output style in HOME (#2875): Claude Code's style file and default
/// `outputStyle`, and Codex's managed block, each create, merge, append, set, unchanged,
/// conflict, skipped or blocked, and the hosts not written with the reason. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// User home root whose `.claude/` and `.codex/AGENTS.md` are inspected.
    home: PathBuf,
    /// Skill package folder carrying `references/home/output-style.md`.
    package: PathBuf,
}

pub const OPERATION: &str = "home.style_plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    serde_json::to_value(moai_cli::home_style::plan(&args.home, &args.package)?)
        .map_err(|_| "encode failed".into())
}
