//! `agents-project-plan` (#2936 R9 S5d): preview of the base agent files in a project's second
//! host folder (`.omp/agents`); writes nothing.
//! The logic is `moai_cli::omp_agents`.

use std::path::PathBuf;

/// Preview the base agents in a project
///
/// Preview writing the base MoAI agents into a project's own agent folder (#2936 R9): per file
/// create, update, unchanged, preserved (your edit), conflict or remove_blocked. Other agent files
/// there are not read. A linked or non-folder parent blocks the project. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The project whose agent folder is inspected.
    project: PathBuf,
    /// Skill package folder whose `references/` the agents' `reads` lists name.
    package: PathBuf,
    /// Agent source folder outside the skill (`agents/` beside `skills/`, #2721).
    agents: PathBuf,
}

pub const OPERATION: &str = "agents_project.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::omp_agents::plan(&args.project, &args.package, &args.agents)
}
