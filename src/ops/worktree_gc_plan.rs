//! `worktree-gc-plan` (WT-4 #2380): read-only removal plan for merged worktrees. Moved out of
//! `src/main.rs` unchanged (#2834 R2c), with the paced arguments #2879 added there.

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
}

pub const OPERATION: &str = "worktree.gc.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::worktree_gc::plan_paced(
        &args.project,
        &args.integration_ref,
        std::time::Duration::from_secs(args.recent_minutes.saturating_mul(60)),
        std::time::Duration::from_millis(args.budget_ms),
        args.now_unix,
        false,
        &moai_cli::worktree_gc::Pace {
            max_checked: args.max_checked,
            cursor: args.cursor,
            only: args.only,
        },
    )
}
