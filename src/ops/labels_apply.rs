//! `labels-apply` (#2434): applies an approved label sync plan (POST/PATCH only). Moved out of `src/main.rs` unchanged (#2834 R4a).

use std::path::PathBuf;

/// Sync previewed GitHub labels
///
/// Apply an approved label sync plan (#2434 decision 40): the exact plan on stdin,
/// `--plan-digest` and `--permission-reference`; POST/PATCH only.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    #[arg(long)]
    plan_digest: String,
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "labels.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    if !moai_cli::contract::permission_reference_ok(&args.permission_reference) {
        return Err("permission reference required".into());
    }
    moai_cli::forge::labels::labels_apply_cli(
        &args.project,
        super::super::plan_input::<serde_json::Value>()?,
        &args.plan_digest,
        &args.permission_reference,
        &moai_cli::forge::transport::GhTransport::from_env()?,
        &moai_cli::forge::policy::SystemClock,
    )
}
