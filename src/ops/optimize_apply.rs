//! `optimize-apply` (SPEC-20261120-worktree-cap-and-cleanup): removes the worktrees a caller
//! picked from the `cleanup` candidates of `optimize-plan`. The logic is `moai_cli::worktree_cap`.

use std::path::PathBuf;

/// Remove picked cleanup candidates
///
/// Remove the worktrees you picked from `optimize-plan`'s `cleanup.candidates` (`--pick
/// PATH@item_digest`, repeatable). Each pick is judged again just before it goes and removed with
/// `git worktree remove` (no `--force`), the deepest first; a changed, locked or claimed worktree,
/// or one whose digest differs, is refused and the next pick goes on. Branches are kept. A record
/// with the restore commands is written under the git folder; the space freed is the change of
/// the free bytes.
#[derive(clap::Args)]
pub struct Args {
    /// Any checkout of the repository.
    project: PathBuf,
    /// Local ref of the integration branch, as given to `optimize-plan`.
    #[arg(long)]
    integration_ref: String,
    /// A change newer than this many minutes keeps a worktree (at least 60).
    #[arg(long, default_value_t = moai_cli::worktree_gc::DEFAULT_RECENT_MINUTES)]
    recent_minutes: u64,
    /// Time budget in milliseconds for each check.
    #[arg(long, default_value_t = moai_cli::worktree_gc::DEFAULT_BUDGET_MS)]
    budget_ms: u64,
    /// A worktree to remove: its `path` and `item_digest` from the preview, joined by `@`.
    #[arg(long = "pick", value_name = "PATH@DIGEST")]
    picks: Vec<String>,
    /// The caller's reference to the user's approval; not an authorization verifier.
    #[arg(long)]
    permission_reference: String,
    /// Test seam: the current time in Unix seconds instead of the clock.
    #[arg(long, hide = true)]
    now_unix: Option<u64>,
}

pub const OPERATION: &str = "optimize.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = moai_cli::hooks::env_home();
    moai_cli::worktree_cap::apply(
        &args.project,
        &args.integration_ref,
        std::time::Duration::from_secs(args.recent_minutes.saturating_mul(60)),
        std::time::Duration::from_millis(args.budget_ms),
        args.now_unix,
        home.as_deref(),
        &args.picks,
        &args.permission_reference,
    )
}
