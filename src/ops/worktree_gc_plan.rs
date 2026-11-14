//! `worktree-gc-plan` (WT-4 #2380): read-only removal plan for merged worktrees. Moved out of
//! `src/main.rs` unchanged (#2834 R2c).

use std::path::PathBuf;

/// Preview removing or trimming merged worktrees
///
/// Read-only removal plan (WT-4, #2380): per linked worktree `remove`, `trim` (cargo
/// build output only) or `skip` with reasons, and a digest. Deletes nothing.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    /// Local ref of the integration branch, e.g. `origin/v2026.11.0`; a worktree whose
    /// HEAD it contains counts as merged. Nothing is fetched.
    #[arg(long)]
    integration_ref: String,
    /// A change newer than this many minutes keeps a worktree (design: 120).
    #[arg(long, default_value_t = moai_cli::worktree_gc::DEFAULT_RECENT_MINUTES)]
    recent_minutes: u64,
    /// Time budget in milliseconds; past it the remaining worktrees are `skip`.
    #[arg(long, default_value_t = moai_cli::worktree_gc::DEFAULT_BUDGET_MS)]
    budget_ms: u64,
    /// Test seam: the current time in Unix seconds instead of the clock.
    #[arg(long, hide = true)]
    now_unix: Option<u64>,
}

pub const OPERATION: &str = "worktree.gc.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::worktree_gc::plan(
        &args.project,
        &args.integration_ref,
        std::time::Duration::from_secs(args.recent_minutes.saturating_mul(60)),
        std::time::Duration::from_millis(args.budget_ms),
        args.now_unix,
    )
}
