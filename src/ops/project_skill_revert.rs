//! `project-skill-revert` (I5 #2242): reverts one applied project-skill copy from its rollback
//! receipt. Moved out of `src/main.rs` unchanged (#2834 R2a).

use std::path::PathBuf;

/// Undo one project skill copy
///
/// Revert one applied project-skill copy from its rollback receipt
/// (I5 #2242): typed request on stdin.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "project_skill.revert";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let mut request =
        super::super::input::<moai_cli::project_skills::ProjectSkillRevertRequest>()?;
    request.project_root = args.project;
    serde_json::to_value(moai_cli::project_skills::project_skill_revert(&request)?)
        .map_err(|_| "encode failed".into())
}
