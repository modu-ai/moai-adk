//! `audit-prep` (SPEC-20261124-audit-prep): an audit takes a fixed audit slot for a card, with the
//! lane's build, and builds the named test targets once. The logic is `moai_cli::audit_prep`.

use std::path::PathBuf;

/// Prepare the audit slot of a card at a head and build it once
///
/// Takes the audit slot the card already holds, else the first one without a live holder, with the
/// card as holder, switches it detached to `--head`, replaces its `target/` with a clone of the
/// lane's `target/` when `--lane` names a worktree at that head with no tracked change, and runs
/// one incremental `cargo test --no-run` of up to two `--test` targets (or the library) in a heavy
/// slot, without `RUSTC_WRAPPER`. The answer lists the units Cargo compiled and the ones it found
/// fresh. A failed build keeps the slot held; `audit-done` finishes it.
#[derive(clap::Args)]
pub struct Args {
    /// The card the audit is for, like `C-1`; it is the holder of the slot.
    card: String,
    /// The audited head, as a full 40-character commit id.
    #[arg(long)]
    head: String,
    /// The lane's worktree whose `target/` is inherited (at the head, no tracked change).
    #[arg(long)]
    lane: Option<PathBuf>,
    /// A test target to build (repeat, at most two); without one the library is built.
    #[arg(long = "test")]
    tests: Vec<String>,
    /// A process whose end ends the hold before the lease does.
    #[arg(long)]
    pid: Option<i64>,
    /// Seconds the hold lasts without an `audit-done` (1 to 86400).
    #[arg(long, default_value_t = moai_cli::audit_slot::DEFAULT_LEASE_SECS)]
    lease_secs: u64,
    /// The project folder (any checkout of the repository).
    #[arg(long, default_value = ".")]
    project: PathBuf,
    /// Linked worktrees allowed before a new slot is refused.
    #[arg(long, default_value_t = moai_cli::worktree_slot::MAX_WORKTREES)]
    max_worktrees: usize,
    /// Free share of the disk, in percent, below which a new slot is refused.
    #[arg(long, default_value_t = moai_cli::worktree_slot::MIN_FREE_PERCENT)]
    min_free_percent: u64,
    /// The heavy slot folder (default: the machine's).
    #[arg(long)]
    slot_dir: Option<PathBuf>,
    /// The most heavy builds at once.
    #[arg(long, default_value_t = 4)]
    max: u32,
    /// The free disk space, in GB, a heavy build needs.
    #[arg(long, default_value_t = 50)]
    min_free_gb: u32,
}

pub const OPERATION: &str = "audit.prep";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::audit_prep::prep(&moai_cli::audit_prep::Prep {
        project: &args.project,
        card: &args.card,
        head: &args.head,
        lane: args.lane.as_deref(),
        tests: &args.tests,
        pid: args.pid,
        lease_secs: args.lease_secs,
        caps: moai_cli::worktree_slot::Caps {
            max_worktrees: args.max_worktrees,
            min_free_percent: args.min_free_percent,
        },
        slot_dir: args.slot_dir,
        max: args.max,
        min_free_gb: args.min_free_gb,
    })
}
