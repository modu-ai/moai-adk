//! `project-skill-apply` (I5 #2242): copies the staged `moai` skill package into the project.
//! Moved out of `src/main.rs` unchanged (#2834 R2a).

use std::path::PathBuf;

/// Copy the moai skill into the project
///
/// Copy the staged `moai` skill package into the project host folders
/// (I5 #2242): the exact request on stdin; requires `opt_in` and writes
/// only under the caller-supplied project root. With `--plan-digest` and
/// `--permission-reference` (both or neither) it applies only the plan
/// `project-skill-plan` printed for this project and package (#2936 R9 S5c).
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    /// The `plan_digest` of the confirmed `project-skill-plan`.
    #[arg(long, requires = "permission_reference")]
    plan_digest: Option<String>,
    /// The caller's reference for the user's approval.
    #[arg(long, requires = "plan_digest")]
    permission_reference: Option<String>,
}

pub const OPERATION: &str = "project_skill.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let mut request = super::super::input::<moai_cli::project_skills::ProjectSkillRequest>()?;
    request.project_root = args.project;
    let receipt = match (&args.plan_digest, &args.permission_reference) {
        (Some(digest), Some(permission)) => {
            moai_cli::project_skills::project_skill_apply_confirmed(&request, digest, permission)?
        }
        _ => moai_cli::project_skills::project_skill_apply(&request)?,
    };
    serde_json::to_value(receipt).map_err(|_| "encode failed".into())
}
