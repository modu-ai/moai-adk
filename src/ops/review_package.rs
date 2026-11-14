//! `review-package` (#2614): the deterministic review package of one audit; reads git, writes
//! nothing. Moved out of `src/main.rs` unchanged (#2834 R1).

use std::path::PathBuf;

/// Build the review package for an audit
///
/// The deterministic review package for an AUDIT (#2614): the diff from the merge base to
/// HEAD, the changed files, and the review instructions (`REVIEW.md` and its includes) and
/// context (`AGENTS.md`, `CLAUDE.md`) read from BASE only. Reads git; writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// Repository root.
    project: PathBuf,
    /// The base revision (the pull request's base branch tip).
    #[arg(long)]
    base: String,
    /// The head revision under review.
    #[arg(long)]
    head: String,
}

pub const OPERATION: &str = "review.package";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    serde_json::to_value(moai_cli::review::package(
        &args.project,
        &args.base,
        &args.head,
    )?)
    .map_err(|_| "encode failed".into())
}
