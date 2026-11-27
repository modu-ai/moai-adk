//! `worktree-slot` (SPEC-20261123-worktree-slots slice 1): the fixed worktree slots, their
//! warming from the seed and the caps. The logic is `moai_cli::worktree_slot`.

use std::path::PathBuf;

/// Fixed worktree slots, warmed from the seed
///
/// `plan` lists the slots `.claude/worktrees/lane-01..10`, `audit-1..3` and `seed` with their
/// state and the caps, and writes nothing. `ensure` adds a missing slot detached at a commit
/// and, when it has no `target/` while the seed has one, copies the seed's `target/` into it
/// (an APFS clone on macOS); a present slot is left as it is. An added slot leaves the folder
/// listed in the local exclude file unless git already ignores it. A slot is refused past the
/// worktree cap or under the free-disk floor.
#[derive(clap::Args)]
pub struct Args {
    #[command(subcommand)]
    action: Action,
    /// The project folder (any checkout of the repository).
    #[arg(long, global = true, default_value = ".")]
    project: PathBuf,
    /// Linked worktrees allowed before a slot is refused.
    #[arg(long, global = true, default_value_t = moai_cli::worktree_slot::MAX_WORKTREES)]
    max_worktrees: usize,
    /// Free share of the disk, in percent, below which a slot is refused.
    #[arg(long, global = true, default_value_t = moai_cli::worktree_slot::MIN_FREE_PERCENT)]
    min_free_percent: u64,
}

#[derive(clap::Subcommand)]
enum Action {
    /// The slots, their state and the caps; writes nothing.
    Plan,
    /// Add a missing slot detached at a commit and warm it from the seed.
    Ensure {
        /// `lane-01` to `lane-10`, `audit-1` to `audit-3` or `seed`.
        slot: String,
        /// The commit a new slot starts at.
        #[arg(long)]
        at: String,
    },
}

pub const OPERATION: &str = "worktree.slot";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let caps = moai_cli::worktree_slot::Caps {
        max_worktrees: args.max_worktrees,
        min_free_percent: args.min_free_percent,
    };
    match args.action {
        Action::Plan => moai_cli::worktree_slot::plan(&args.project, caps),
        Action::Ensure { slot, at } => {
            moai_cli::worktree_slot::ensure(&args.project, &slot, &at, caps)
        }
    }
}
