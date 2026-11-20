//! `work-flow-plan` (SPEC-20261120-mflow-integration-branch-choice): the first-setup
//! integration-branch question evidence. One word, one action; the logic is
//! `moai_cli::work_flow`.

use std::path::PathBuf;

/// Reads only. Print the integration-branch question evidence
///
/// Reads only. Prints the repository's branch candidates with a `canonical` flag each
/// (a remote pull-request candidate ref such as `refs/remotes/origin/pr/<n>` is
/// `canonical: false` with a candidate label and is never recommended), one recommended
/// branch with its reason — the `work_flow.integration_branch` recorded in the project
/// profile on a re-run, else the remote's default branch (`origin/HEAD`), else the current
/// branch —, the recorded work-flow state, and, when the recorded target is missing or
/// occupied by another worktree, a `missing-target` or `occupied-worktree` state with the
/// safe choices. It creates no branch, no worktree, no file and no ref, and it runs no
/// network command.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder (its profile holds the recorded work flow).
    #[arg(long, default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "work_flow_plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::work_flow::plan(&args.project)
}
