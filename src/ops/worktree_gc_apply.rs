//! `worktree-gc-apply` (WT-9 #2498): acts on a `worktree-gc-plan` after a digest match. Moved
//! out of `src/main.rs` unchanged (#2834 R2c), with the paced arguments #2879 added there.

use std::path::PathBuf;

/// Remove or trim the previewed worktrees
///
/// Act on a `worktree-gc-plan` (WT-9, #2498): recomputes the plan with the same inputs,
/// refuses before any write unless its digest equals `--plan-digest`, then trims the
/// listed cargo output and removes the listed worktrees (no `--force`; branches kept).
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    #[arg(long)]
    integration_ref: String,
    #[arg(long, default_value_t = moai_cli::worktree_gc::DEFAULT_RECENT_MINUTES)]
    recent_minutes: u64,
    #[arg(long, default_value_t = moai_cli::worktree_gc::DEFAULT_BUDGET_MS)]
    budget_ms: u64,
    /// Test seam: the current time in Unix seconds instead of the clock.
    #[arg(long, hide = true)]
    now_unix: Option<u64>,
    /// Examine at most this many worktrees in one run; the rest are `budget_spent` and
    /// `next_cursor` says where the next run goes on (#2879).
    #[arg(long)]
    max_checked: Option<usize>,
    /// Start after this candidate `path` of an earlier plan (its `next_cursor`).
    #[arg(long)]
    cursor: Option<String>,
    /// Plan only this worktree (repeatable): a lane cleans its own card worktrees (#2879).
    #[arg(long = "worktree", value_name = "PATH")]
    only: Vec<PathBuf>,
    #[arg(long)]
    plan_digest: String,
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "worktree.gc.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::worktree_gc::apply(
        &args.project,
        &args.integration_ref,
        std::time::Duration::from_secs(args.recent_minutes.saturating_mul(60)),
        std::time::Duration::from_millis(args.budget_ms),
        args.now_unix,
        &moai_cli::worktree_gc::Pace {
            max_checked: args.max_checked,
            cursor: args.cursor,
            only: args.only,
        },
        &args.plan_digest,
        &args.permission_reference,
    )
}
