//! `review-check` (#2614 AC-3): format check of review findings before they are posted; reads
//! git only. Moved out of `src/main.rs` unchanged (#2834 R1).

use std::path::PathBuf;

/// Check review findings before they are posted
///
/// Format check of review findings before they are posted (#2614 AC-3): stdin
/// `{"schema_version":1,"findings":[…]}`; each finding needs a severity, a `file:line` in
/// HEAD, a title, evidence quoted from that line, and a confirmed verification. Prints the
/// verdict (any `important` is FAIL), counts, post body and digest. Reads git only.
#[derive(clap::Args)]
pub struct Args {
    /// Repository root.
    project: PathBuf,
    /// The head revision the findings cite.
    #[arg(long)]
    head: String,
}

pub const OPERATION: &str = "review.check";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let findings: serde_json::Value = super::super::input()?;
    serde_json::to_value(moai_cli::review::check(
        &args.project,
        &args.head,
        &findings,
    )?)
    .map_err(|_| "encode failed".into())
}
