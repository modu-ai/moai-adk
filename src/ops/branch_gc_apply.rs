//! `branch-gc-apply` (SPEC-20261123-branch-gc): deletes the planned local branches with
//! `git branch -d` after a digest match; the logic is `moai_cli::branch_gc`.

use std::path::PathBuf;

/// Delete the previewed local branches
///
/// Act on a `branch-gc-plan` (SPEC-20261123-branch-gc): recomputes the plan with the same
/// input, refuses before any write unless its digest equals `--plan-digest`, then runs
/// `git branch -d` (never `-D`) for each `delete` branch. Remote branches are not touched.
/// Deleting is the user's pick: run it only after they confirm.
#[derive(clap::Args)]
pub struct Args {
    /// The repository (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
    /// Local ref of the integration branch, the same value the plan was made with.
    #[arg(long)]
    integration_ref: String,
    /// The `plan_digest` the preview printed.
    #[arg(long)]
    plan_digest: String,
    /// The reference to the user's confirmation.
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "branch.gc.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::branch_gc::apply(
        &args.project,
        &args.integration_ref,
        &args.plan_digest,
        &args.permission_reference,
        None,
        &|_| {},
    )
}
