//! `audit-slot` (SPEC-20261123-audit-slot): an audit takes a fixed audit slot, switches it to
//! the audited head with the lane's build, and leaves it clean. The logic is
//! `moai_cli::audit_slot`.

use std::path::PathBuf;

/// Audit slots that inherit the lane's build
///
/// `prep` takes `.claude/worktrees/audit-1..3` (or `auto`, the first free one) for an audited
/// head as `--holder`: it adds a missing slot, refuses one another live holder has or that holds
/// changes, records the holder, switches it detached to the head, replaces its `target/` with a
/// clone of the lane's `target/` when `--lane` names a worktree at that head with no tracked
/// change (else it keeps the slot's build or warms it from the seed) and touches the sources so
/// the workspace crate compiles once from them. `done` reverts what an audit changed in the
/// slot with `--discard`, keeps its `target/` and releases the slot.
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
        /// `audit-1`, `audit-2`, `audit-3`, or `auto` for the first slot without a live holder.
        slot: String,
        /// The audited head, as a full 40-character commit id.
        #[arg(long)]
        head: String,
        /// The lane's worktree whose `target/` is inherited (at the head, no tracked change).
        #[arg(long)]
        lane: Option<PathBuf>,
        /// Who takes the slot (a lane or auditor name); another holder is refused while it lives.
        #[arg(long)]
        holder: String,
        /// A process whose end ends the hold before the lease does.
        #[arg(long)]
        pid: Option<i64>,
        /// Seconds the hold lasts without a `done` (1 to 86400).
        #[arg(long, default_value_t = moai_cli::audit_slot::DEFAULT_LEASE_SECS)]
        lease_secs: u64,
    },
    /// Leave an audit slot clean for the next audit and release it; `target/` stays.
    Done {
        /// `audit-1`, `audit-2` or `audit-3`.
        slot: String,
        /// Revert tracked changes and remove untracked files outside `target/`.
        #[arg(long)]
        discard: bool,
        /// Who finishes it: the holder, or anyone once the holder is dead.
        #[arg(long)]
        holder: String,
    },
}

pub const OPERATION: &str = "audit.slot";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let caps = moai_cli::worktree_slot::Caps {
        max_worktrees: args.max_worktrees,
        min_free_percent: args.min_free_percent,
    };
    match args.action {
        Action::Prep {
            slot,
            head,
            lane,
            holder,
            pid,
            lease_secs,
        } => {
            let who = moai_cli::audit_slot::Who {
                holder: &holder,
                pid,
                lease_secs,
            };
            moai_cli::audit_slot::prep(&args.project, &slot, &head, lane.as_deref(), caps, &who)
        }
        Action::Done {
            slot,
            discard,
            holder,
        } => {
            let who = moai_cli::audit_slot::Who {
                holder: &holder,
                pid: None,
                lease_secs: moai_cli::audit_slot::DEFAULT_LEASE_SECS,
            };
            moai_cli::audit_slot::done(&args.project, &slot, discard, &who)
        }
    }
}
