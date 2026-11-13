//! `home-routing-plan` (#2638): preview of the MoAI routing section in the global instruction
//! files; writes nothing.

use std::path::PathBuf;

/// Preview the MoAI routing section in HOME's global instruction files (#2638): per host
/// create, append, merge, unchanged or conflict, and the hosts not written with the reason.
/// Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// User home root whose `.claude/CLAUDE.md` and `.codex/AGENTS.md` are inspected.
    home: PathBuf,
    /// Skill package folder carrying `references/home/routing.md`.
    package: PathBuf,
}

pub const OPERATION: &str = "home.routing_plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    serde_json::to_value(moai_cli::home_routing::plan(&args.home, &args.package)?)
        .map_err(|_| "encode failed".into())
}
