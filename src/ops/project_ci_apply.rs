//! `project-ci-apply` (SPEC-20261122-project-ci-design REQ-7): publishes an approved CI plan.
//! The logic is `moai_cli::project_ci_plan`.

use std::path::PathBuf;

/// Write previewed CI workflow files
///
/// Apply an approved CI plan: the exact plan on stdin, `--plan-digest` and
/// `--permission-reference`. A different digest, a plan the project has outgrown since the
/// preview, or a conflict refuses the whole plan before anything is written. Only the planned
/// files, the removals the user picked and the CI manifest are touched.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    #[arg(long)]
    plan_digest: String,
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "project.ci_apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::project_ci_plan::ci_apply(
        &args.project,
        super::super::plan_input::<moai_cli::project_ci_plan::CiPlan>()?,
        &args.plan_digest,
        &args.permission_reference,
    )
}
