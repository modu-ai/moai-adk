//! `branch-gc-plan` (SPEC-20261123-branch-gc): read-only plan of the local branches that are
//! merged and unused; the logic is `moai_cli::branch_gc`.

use std::path::PathBuf;

/// Preview deleting the merged, unused local branches
///
/// Read-only plan (SPEC-20261123-branch-gc): every local branch with the action `delete`
/// (merged into the integration ref, no worktree has it, no active card reservation names it,
/// not a preserved ref) or `keep` with its reasons, and a digest. Deletes nothing and fetches
/// nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The repository (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
    /// Local ref of the integration branch, e.g. `v2026.11.0` or `origin/v2026.11.0`; a branch
    /// whose tip it contains counts as merged. Nothing is fetched.
    #[arg(long)]
    integration_ref: String,
}

pub const OPERATION: &str = "branch.gc.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::branch_gc::plan(&args.project, &args.integration_ref, None)
}
