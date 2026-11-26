//! `worktree-new` (SPEC-20261124-lane-worktree-command): a lane's worktree in the primary
//! checkout, made and read back. The logic is `moai_cli::worktree_new`.

use std::path::PathBuf;

/// Make a lane's worktree in the primary checkout and read it back
///
/// Adds `<primary checkout>/.claude/worktrees/<name>` on the new branch `--branch`, cut from the
/// integration branch (or from `--base`), then reads `git worktree list --porcelain` again and
/// prints the absolute path only when the list holds that path on the new branch at the base
/// commit and the folder exists. The primary checkout is read from the git common directory, so
/// a run from inside a linked worktree still makes the folder in the primary's place. The name is
/// one plain folder name (lower-case letters, digits and `-`, at most 64 characters), never a
/// path: a relative path, a nested place, a taken folder or branch, an unknown base and a link
/// in the place are refused and nothing is made. It only makes and reads back; it has no option
/// to force, delete, prune or move a worktree.
#[derive(clap::Args)]
pub struct Args {
    /// The folder name under `.claude/worktrees`, for example `lane-a`.
    name: String,
    /// The new branch the worktree starts on, for example `spec/<slug>-<yyyymmdd>`.
    #[arg(long)]
    branch: String,
    /// The commit the branch starts at (default: the integration branch's tip).
    #[arg(long)]
    base: Option<String>,
    /// The project folder (any checkout of the repository).
    #[arg(long, default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "worktree.new";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::worktree_new::make(
        &args.project,
        &args.name,
        &args.branch,
        args.base.as_deref(),
    )
}
