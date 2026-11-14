//! `run-verify` (05-C #2003): re-checks the pinned contract and runs the approved verification
//! commands; returns a receipt. Moved out of `src/main.rs` unchanged (#2834 R1).

use std::path::PathBuf;

/// Re-check a run's pinned verification
///
/// Temporary (05-C #2003; the name is fixed after P06): re-check the pinned contract and
/// run the approved verification commands in PROJECT as argv; returns a receipt.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "run.verify";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let request = super::super::input::<moai_cli::run::RunRequest>()?;
    if request.protocol_version != moai_cli::contract::PROTOCOL_VERSION {
        return Err("unsupported protocol_version; supported: 1".into());
    }
    let current = moai_cli::plan::current_revision(&args.project)?;
    serde_json::to_value(moai_cli::run::run_verify(&args.project, &request, &current)?)
        .map_err(|_| "encode failed".into())
}
