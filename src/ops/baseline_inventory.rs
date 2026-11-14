//! `baseline-inventory`: read-only baseline facts of a checkout. Moved out of `src/main.rs`
//! unchanged (#2834 R3).

use std::path::PathBuf;

/// List a checkout's baseline facts
///
/// Read-only baseline inventory of a checkout (typed JSON on stdout).
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    /// Go-generation baseline commit whose `.claude` trees are compared (read with ls-tree only).
    #[arg(long)]
    go_baseline_commit: Option<String>,
    /// Extra repository to look up the baseline commit in when the project lacks it.
    #[arg(long)]
    go_repo: Option<PathBuf>,
}

pub const OPERATION: &str = "baseline.inventory";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::baseline::inventory(
        &args.project,
        args.go_baseline_commit.as_deref(),
        args.go_repo.as_deref(),
    )
}
