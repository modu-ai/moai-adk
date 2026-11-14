//! `archive-apply` (D2 #2188): moves the archive-eligible SPECs of a matching plan. Moved out of
//! `src/main.rs` unchanged (#2834 R2a).

use std::path::PathBuf;

/// Archive the previewed SPECs
///
/// Move the archive-eligible SPECs (D2 #2188, REQ-D2-01): recomputes the plan and
/// requires its `plan_digest`; each SPEC is hashed, copied, read back and indexed
/// before its source is removed. Any mismatch cancels the move and keeps the source.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    #[arg(long)]
    plan_digest: String,
    #[arg(long)]
    permission_reference: String,
    #[arg(long)]
    grace_days: Option<u64>,
}

pub const OPERATION: &str = "archive.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::archive::apply(
        &args.project,
        std::time::SystemTime::now(),
        args.grace_days,
        &args.plan_digest,
        &args.permission_reference,
    )
}
