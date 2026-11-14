//! `omp-agents-plan` (#2936 R9 S5d): preview of the project `.omp/agents` files; writes nothing.
//! The logic is `moai_cli::omp_agents`.

use std::path::PathBuf;

/// Preview the project OMP agents
///
/// Preview writing the base MoAI agents into PROJECT's `.omp/agents` (#2936 R9): per file create,
/// update, unchanged, preserved (your edit), conflict or remove_blocked. Other agent files there
/// are not read. A linked or non-folder `.omp` blocks the project. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The project whose `.omp/agents` is inspected.
    project: PathBuf,
    /// Skill package folder whose `references/` the agents' `reads` lists name.
    package: PathBuf,
    /// Agent source folder outside the skill (`agents/` beside `skills/`, #2721).
    agents: PathBuf,
}

pub const OPERATION: &str = "omp_agents.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::omp_agents::plan(&args.project, &args.package, &args.agents)
}
