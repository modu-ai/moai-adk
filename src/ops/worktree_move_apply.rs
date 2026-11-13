//! `worktree-move-apply` (#2723): move the previewed worktrees, only on an identical fresh
//! preview. The logic is `moai_cli::worktree_move`.

use std::path::PathBuf;

/// Move the worktrees a `worktree-move-plan` listed as `move` (#2723), only when the plan
/// recomputed from the same inputs has the same `plan_digest`. Each move re-checks processes
/// and open files, runs `git worktree move` (never `--force`), then checks git lists it at the
/// new path with the same HEAD and branch (`git worktree repair` once if not). Each moved
/// worktree comes with the command that moves it back.
#[derive(clap::Args)]
pub struct Args {
    /// Any checkout of the repository (the primary one or a worktree).
    project: PathBuf,
    /// The same `--root` the preview read (default as in the preview).
    #[arg(long)]
    root: Option<PathBuf>,
    /// The same observations file the preview read.
    #[arg(long)]
    observations: Option<PathBuf>,
    /// The preview's `plan_digest`.
    #[arg(long)]
    plan_digest: String,
    /// The caller's reference for the user's approval of these moves.
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "worktree.move.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let obs = moai_cli::worktree_move::read_observations(args.observations.as_deref())?;
    moai_cli::worktree_move::apply(
        &args.project,
        args.root.as_deref(),
        &obs,
        &args.plan_digest,
        &args.permission_reference,
    )
}
