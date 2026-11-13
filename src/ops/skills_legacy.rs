//! `skills-legacy` (#2592 Codex parity): legacy `moai-*` skills beside `moai`; writes nothing.

use std::path::PathBuf;

/// List legacy `moai-*` skill folders (every `moai-` entry except `moai` itself) under HOME's
/// host skill folders — `.agents/skills`, `.claude/skills`, `.codex/skills`,
/// `.omp/agent/skills`, `.cursor/skills` — with `count` and a `summary` line that says `0건`
/// when there are none, plus the Codex `[[skills.config]]` entry that turns each Codex-read one
/// off. Links are reported, never followed. Reads only; moves, deletes and edits nothing.
#[derive(clap::Args)]
pub struct Args {
    /// User home root.
    home: PathBuf,
}

pub const OPERATION: &str = "skills.legacy";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    Ok(moai_cli::skills_legacy::check(&args.home))
}
