//! `project-ci-plan` (SPEC-20261122-project-ci-design REQ-6): previews the project's CI
//! workflow files, read-only. The logic is `moai_cli::project_ci_plan`.

use std::path::PathBuf;

/// Preview the project's CI workflow files
///
/// Read-only. Typed request on stdin (`protocol_version`, optional `runs_per_month`, `language`,
/// `remove`). Plans the CI workflow files from the answers and declared checks saved in the
/// project profile and compares each with the file on disk and with its recorded baseline: a
/// file is created, merged, unchanged or a conflict that is kept. Prints the monthly cost
/// estimate and the plan digest. A file MoAI wrote earlier that the answers no longer produce is
/// shown as stale and removed only when its path is in `remove` and it was not changed since.
/// It writes no file.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder (its profile holds the saved CI answers and declared checks).
    project: PathBuf,
}

pub const OPERATION: &str = "project.ci_plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    serde_json::to_value(moai_cli::project_ci_plan::ci_plan(
        &args.project,
        super::super::input::<moai_cli::project_ci_plan::CiRequest>()?,
    )?)
    .map_err(|_| "encode failed".into())
}
