//! `labels-plan` (#2434): plans a project's label sync (GET only). Moved out of `src/main.rs` unchanged (#2834 R4b).

use std::path::PathBuf;

/// Preview syncing the project's GitHub labels
///
/// Plan a project's label sync from `.github/moai/labels.yml` (#2434 decision 40): typed
/// request `{schema_version, owner, repo}` on stdin; GET only, never deletes.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "labels.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    // Progress lines on a terminal's stderr, as `shows_progress` turned on for it in main.rs.
    moai_cli::progress::install_stderr();
    moai_cli::forge::labels::labels_plan_cli(
        &args.project,
        super::super::input::<moai_cli::forge::labels::LabelsRequest>()?,
        &moai_cli::forge::transport::GhTransport::from_env()?,
        &moai_cli::forge::policy::SystemClock,
    )
}
