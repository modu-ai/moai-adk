//! `audit-reduce` (05-D #2004): independent audit of a candidate; receipts on stdin never count.
//! Moved out of `src/main.rs` unchanged (#2834 R1).

use std::path::PathBuf;

/// Run an independent audit of a candidate
///
/// Temporary (05-D #2004; the name is fixed after P06): independent audit of PROJECT. Re-runs
/// the approved verification itself and decides admission; receipts on stdin never count.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "audit.reduce";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let request = super::super::input::<moai_cli::audit::AuditInput>()?;
    if request.protocol_version != moai_cli::contract::PROTOCOL_VERSION {
        return Err("unsupported protocol_version; supported: 1".into());
    }
    serde_json::to_value(moai_cli::audit::audit(&args.project, &request)?)
        .map_err(|_| "encode failed".into())
}
