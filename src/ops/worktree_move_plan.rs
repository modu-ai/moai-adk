//! `worktree-move-plan` (#2723): inventory of the linked worktrees and a preview of moving
//! them under one common root. Reads only; the logic is `moai_cli::worktree_move`.

use std::path::PathBuf;

/// Preview moving this clone's worktrees
///
/// Preview moving this clone's linked worktrees under `<root>/<project key>/<name>` (#2723).
/// Lists every worktree with `move`, `keep` (already there) or `blocked` and its reasons:
/// main worktree, missing, locked, submodules, uncommitted or untracked files, unpushed
/// commits, a stash of its branch, a process inside it, a held branch, or the hold check not
/// run (no `holds` observation). Writes nothing; the answer carries a `plan_digest`.
#[derive(clap::Args)]
pub struct Args {
    /// Any checkout of the repository (the primary one or a worktree).
    project: PathBuf,
    /// Absolute folder that holds one folder per project (default `$HOME/<home state
    /// folder>/worktrees`, #2723). Must lie outside every worktree.
    #[arg(long)]
    root: Option<PathBuf>,
    /// JSON file: `schema_version` 1 and optional `holds` {source, branches [{branch,
    /// reason: hold|handoff_pending}]}. Without `holds` every worktree is blocked by
    /// `hold_check_not_run`.
    #[arg(long)]
    observations: Option<PathBuf>,
}

pub const OPERATION: &str = "worktree.move.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let obs = moai_cli::worktree_move::read_observations(args.observations.as_deref())?;
    moai_cli::worktree_move::plan(&args.project, args.root.as_deref(), &obs)
}
