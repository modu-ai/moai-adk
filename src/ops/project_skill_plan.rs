//! `project-skill-plan` (I5 #2242): read-only plan of project `moai` skill copies. Moved out of
//! `src/main.rs` unchanged (#2834 R2a).

use std::path::PathBuf;

/// Preview project copies of the moai skill
///
/// Plan project-unit `moai` skill copies read-only (I5 #2242, P04 HOLD):
/// typed request on stdin; judgment envelope on stdout; writes nothing.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "project_skill.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let mut request = super::super::input::<moai_cli::project_skills::ProjectSkillRequest>()?;
    request.project_root = args.project;
    serde_json::to_value(moai_cli::project_skills::project_skill_plan(&request)?)
        .map_err(|_| "encode failed".into())
}
