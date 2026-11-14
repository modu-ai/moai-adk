//! `forge-audit-status` (#2173): posts the `moai/audit` commit status on an exact SHA. Moved out
//! of `src/main.rs` unchanged (#2834 R2b).

use std::path::PathBuf;

/// Post the audit commit status on a SHA
///
/// Post the `moai/audit` commit status on an exact SHA (typed JSON on stdin). Uses the
/// audit App token from `MOAI_AUDIT_APP_TOKEN` when set, else warns (#2173).
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "forge.audit_status";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::forge::audit_status::audit_status_cli(
        &args.project,
        super::super::input::<moai_cli::forge::audit_status::AuditStatusRequest>()?,
        &args.permission_reference,
        moai_cli::forge::transport::GhTransport::from_env()?,
        &moai_cli::forge::policy::SystemClock,
    )
}
