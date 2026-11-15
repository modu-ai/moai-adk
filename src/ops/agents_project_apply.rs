//! `agents-project-apply` (#2936 R9 S5d): write a previewed set of base agent files into a
//! project's second host folder (`.omp/agents`).
//! The logic is `moai_cli::omp_agents`.

use std::path::PathBuf;

/// Write the previewed base agents
///
/// Write the `create` and `update` files of a previewed `agents-project-plan` into the project's
/// agent folder and record what was written. The digest must equal a fresh plan's; a file you
/// edited is kept, and nothing is deleted.
#[derive(clap::Args)]
pub struct Args {
    /// The project whose agent folder is written.
    project: PathBuf,
    /// Skill package folder, as given to the plan.
    package: PathBuf,
    /// Agent source folder, as given to the plan.
    agents: PathBuf,
    /// The `plan_digest` of the confirmed `agents-project-plan`.
    #[arg(long)]
    plan_digest: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "agents_project.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::omp_agents::apply(
        &args.project,
        &args.package,
        &args.agents,
        &args.plan_digest,
        &args.permission_reference,
    )
}
