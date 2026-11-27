//! `worktree-move-plan` (#2723): inventory of the linked worktrees and a preview of moving
//! them under one common root. Reads only; the logic is `moai_cli::worktree_move`.

use std::path::PathBuf;

/// Preview moving this clone's worktrees
///
/// Preview moving this clone's linked worktrees into `<root>/<project>-<name>` (#2723, #2835);
/// without `--root` the root is the primary checkout's `.claude/worktrees`, where
/// `worktree-plan` creates new ones. Lists every worktree with its `location`, `move`, `keep`
/// (already in the root, in the earlier `~/.moai-cli/worktrees/<key>/` root or in the Go
/// generation's `~/.moai/worktrees/projects/`; without `--root` also anything in the project's
/// `.claude/worktrees` or the home worktree folder, and a worktree holding one of those places)
/// or `blocked` and its reasons: main worktree, missing, locked, submodules, uncommitted or
/// untracked files, unpushed commits, a stash of its branch, a process inside it, a held
/// branch, or the hold check not run (no `holds` observation). A linked `.claude` or
/// `.claude/worktrees` in the way of the place is answered as `would_refuse`. Writes nothing;
/// the answer carries a `plan_digest`.
#[derive(clap::Args)]
pub struct Args {
    /// Any checkout of the repository (the primary one or a worktree).
    project: PathBuf,
    /// Absolute folder the worktrees move into (default: the primary checkout's
    /// `.claude/worktrees`; a call without it needs a readable HOME). Must lie outside every
    /// Cargo build folder (#2930 R-3) and outside every worktree, the primary checkout's
    /// `.claude/worktrees` itself excepted.
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
