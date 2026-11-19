//! `reports-move-plan` (C-126): which untracked report files an earlier generation left in the
//! repository would move to the user-level reports folder. Reads only; the logic is
//! `moai_cli::reports_move`.

use std::path::PathBuf;

/// Preview moving the earlier report files out of the repository
///
/// Preview moving the files git does not track under `.moai-cli/reports` and `.moai/reports`
/// of the project to `<HOME>/.moai/reports/<project>` with the same relative paths. A file git
/// tracks stays where it is and is only counted; a destination name that already exists is kept
/// and counted as skipped. The project must be a git repository, otherwise the plan is `blocked`
/// and plans no move. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The project (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "reports_move.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = moai_cli::home_workdir::home()?;
    let plan = moai_cli::reports_move::plan(&args.project, &home)?;
    serde_json::to_value(plan).map_err(|e| e.to_string())
}
