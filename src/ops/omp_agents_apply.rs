//! `omp-agents-apply` (#2936 R9 S5d): write a previewed set of project `.omp/agents` files.
//! The logic is `moai_cli::omp_agents`.

use std::path::PathBuf;

/// Write the previewed project OMP agents
///
/// Write the `create` and `update` files of a previewed `omp-agents-plan` into PROJECT's
/// `.omp/agents` and record what was written. The digest must equal a fresh plan's; a file you
/// edited is kept, and nothing is deleted.
#[derive(clap::Args)]
pub struct Args {
    /// The project whose `.omp/agents` is written.
    project: PathBuf,
    /// Skill package folder, as given to the plan.
    package: PathBuf,
    /// Agent source folder, as given to the plan.
    agents: PathBuf,
    /// The `plan_digest` of the confirmed `omp-agents-plan`.
    #[arg(long)]
    plan_digest: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "omp_agents.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::omp_agents::apply(
        &args.project,
        &args.package,
        &args.agents,
        &args.plan_digest,
        &args.permission_reference,
    )
}
