//! `project-update-plan` (#2936 S1): one read-only preview of the project update, in the
//! three phases of the user's UPDATE flow. The logic is `moai_cli::project_update`.

use std::path::PathBuf;

/// Preview the whole project update in one view
///
/// Preview the project update (#2936): the backup phase (the earlier generation's state moves
/// aside), the install phase (profile, generated files, Claude Code settings, GitHub
/// templates) and the merge phase (a later slice). Each step shows its state, reason, preview,
/// plan digest and the exact apply command with its stdin; the only placeholder is
/// `<permission-reference>`. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// Project folder to preview.
    project: PathBuf,
}

pub const OPERATION: &str = "project.update_plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::project_update::plan(&args.project)
}
