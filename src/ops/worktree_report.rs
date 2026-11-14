//! `worktree-report` (WT-1 #2380): read-only space report of worktrees, build folders and
//! temporary `moai-*` folders. Moved out of `src/main.rs` unchanged (#2834 R3).

use std::path::PathBuf;

/// Report the space worktrees use
///
/// Read-only space report of the repository's worktrees, their build folders and the
/// temporary `moai-*` folders (WT-1, #2380). Facts only; removal candidates are not
/// evaluated. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    /// Folder whose `moai-*` children are classified (default `/tmp`).
    #[arg(long)]
    tmp_root: Option<PathBuf>,
    /// Time budget in milliseconds for the size walk; past it a class is `partial`.
    #[arg(long, default_value_t = moai_cli::worktree_report::DEFAULT_BUDGET_MS)]
    budget_ms: u64,
}

pub const OPERATION: &str = "worktree.report";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::worktree_report::report(
        &args.project,
        &args.tmp_root.unwrap_or_else(|| PathBuf::from("/tmp")),
        std::time::Duration::from_millis(args.budget_ms),
    )
}
