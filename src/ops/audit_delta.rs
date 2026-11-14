//! `audit-delta`: compares one local two-parent merge with a controlled replay; evidence only.
//! Moved out of `src/main.rs` unchanged (#2834 R1).

use std::path::PathBuf;

/// Compare a two-parent merge with a replay
///
/// Compare one local two-parent merge with a controlled replay; evidence only, not AUDIT PASS.
#[derive(clap::Args)]
pub struct Args {
    #[arg(default_value = ".")]
    project: PathBuf,
    /// Previously audited commit (40 lowercase hex).
    #[arg(long)]
    pass_head: String,
    /// Integrated base tip (40 lowercase hex).
    #[arg(long)]
    tip: String,
    /// Candidate merge commit (40 lowercase hex).
    #[arg(long)]
    head: String,
}

pub const OPERATION: &str = "audit.delta";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    serde_json::to_value(moai_cli::audit_delta::inspect(
        &args.project,
        moai_cli::audit_delta::Request {
            pass_head: args.pass_head,
            tip: args.tip,
            head: args.head,
        },
    )?)
    .map_err(|_| "encode failed".into())
}
