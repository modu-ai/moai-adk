//! `project-skill-apply` (I5 #2242): copies the staged `moai` skill package into the project.
//! Moved out of `src/main.rs` unchanged (#2834 R2a).

use std::path::PathBuf;

/// Copy the moai skill into the project
///
/// Copy the staged `moai` skill package into the project host folders
/// (I5 #2242): the exact request on stdin; requires `opt_in` and writes
/// only under the caller-supplied project root.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "project_skill.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let mut request = super::super::input::<moai_cli::project_skills::ProjectSkillRequest>()?;
    request.project_root = args.project;
    serde_json::to_value(moai_cli::project_skills::project_skill_apply(&request)?)
        .map_err(|_| "encode failed".into())
}
