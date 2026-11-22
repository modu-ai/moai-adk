//! `audit-slot` (SPEC-20261123-audit-slot): an audit takes a fixed audit slot, switches it to
//! the audited head with the lane's build, and leaves it clean. The logic is
//! `moai_cli::audit_slot`.

use std::path::PathBuf;

/// Audit slots that inherit the lane's build
///
/// `prep` takes `.claude/worktrees/audit-1..3` for an audited head: it adds a missing slot,
/// refuses one that holds changes, switches it detached to the head and replaces its `target/`
/// with a clone of the lane's `target/` when `--lane` names a worktree at that head with no
/// tracked change (else it keeps the slot's build or warms it from the seed). `done` reverts
/// what an audit changed in the slot with `--discard` and keeps its `target/`.
#[derive(clap::Args)]
pub struct Args {
    #[command(subcommand)]
    action: Action,
    /// The project folder (any checkout of the repository).
    #[arg(long, global = true, default_value = ".")]
    project: PathBuf,
    /// Linked worktrees allowed before a new slot is refused.
    #[arg(long, global = true, default_value_t = moai_cli::worktree_slot::MAX_WORKTREES)]
    max_worktrees: usize,
    /// Free share of the disk, in percent, below which a new slot is refused.
    #[arg(long, global = true, default_value_t = moai_cli::worktree_slot::MIN_FREE_PERCENT)]
    min_free_percent: u64,
}

#[derive(clap::Subcommand)]
enum Action {
    /// Take an audit slot for a head and place a build in it.
    Prep {
        /// `audit-1`, `audit-2` or `audit-3`.
        slot: String,
        /// The audited head.
        #[arg(long)]
        head: String,
        /// The lane's worktree whose `target/` is inherited (at the head, no tracked change).
        #[arg(long)]
        lane: Option<PathBuf>,
    },
    /// Leave an audit slot clean for the next audit; `target/` stays.
    Done {
        /// `audit-1`, `audit-2` or `audit-3`.
        slot: String,
        /// Revert tracked changes and remove untracked files outside `target/`.
        #[arg(long)]
        discard: bool,
    },
}

pub const OPERATION: &str = "audit.slot";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let caps = moai_cli::worktree_slot::Caps {
        max_worktrees: args.max_worktrees,
        min_free_percent: args.min_free_percent,
    };
    match args.action {
        Action::Prep { slot, head, lane } => {
            moai_cli::audit_slot::prep(&args.project, &slot, &head, lane.as_deref(), caps)
        }
        Action::Done { slot, discard } => {
            moai_cli::audit_slot::done(&args.project, &slot, discard)
        }
    }
}
