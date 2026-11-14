//! `decision-record`: appends one native decision record to a run's log. Moved out of
//! `src/main.rs` unchanged (#2834 R2b).

use std::path::PathBuf;

/// Append one decision record to a run
///
/// Append one native decision record (stdin) to `.moai-cli/decisions/<RUN>.jsonl`.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    #[arg(long)]
    run: String,
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "decision.record";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::decision::append(
        &args.project,
        &args.run,
        &super::super::input()?,
        &args.permission_reference,
    )
}
