//! `worktree-move-plan` (#2723): inventory of the linked worktrees and a preview of moving
//! them under one common root. Reads only; the logic is `moai_cli::worktree_move`.

use std::path::PathBuf;

/// Preview moving this clone's worktrees
///
/// Preview moving this clone's linked worktrees into `<root>/<project>-<name>` (#2723, #2835),
/// the place `worktree-plan` creates new ones in. Lists every worktree with its `location`,
/// `move`, `keep` (already in the root, in the earlier `~/.moai-cli/worktrees/<key>/` root or
/// in the Go generation's `~/.moai/worktrees/projects/`) or `blocked` and its reasons:
/// main worktree, missing, locked, submodules, uncommitted or untracked files, unpushed
/// commits, a stash of its branch, a process inside it, a held branch, or the hold check not
/// run (no `holds` observation). Writes nothing; the answer carries a `plan_digest`.
#[derive(clap::Args)]
pub struct Args {
    /// Any checkout of the repository (the primary one or a worktree).
    project: PathBuf,
    /// Absolute folder the worktrees move into (default `<HOME>/.moai/worktrees`, #2835).
    /// Must lie outside every worktree.
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
