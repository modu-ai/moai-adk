//! `worktree-gc-apply` (WT-9 #2498): acts on a `worktree-gc-plan` after a digest match. Moved
//! out of `src/main.rs` unchanged (#2834 R2c).

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
        &args.plan_digest,
        &args.permission_reference,
    )
}
