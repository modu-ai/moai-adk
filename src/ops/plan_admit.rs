//! `plan-admit` (05-B #2002): RUN admission of an approved plan; reads `HEAD^{commit}` with
//! git, writes nothing. Moved out of `src/main.rs` unchanged (#2834 R1).

use std::path::PathBuf;

/// Admit a plan to a run
///
/// Temporary (05-B #2002; the name is fixed after P06): RUN admission of an approved plan.
/// Typed request on stdin; reads the project's `HEAD^{commit}` with git, writes nothing.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "plan.admit";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let request = super::super::input::<moai_cli::plan::AdmissionRequest>()?;
    if request.protocol_version != moai_cli::contract::PROTOCOL_VERSION {
        return Err("unsupported protocol_version; supported: 1".into());
    }
    let current = moai_cli::plan::current_revision(&args.project)?;
    serde_json::to_value(moai_cli::plan::admit(&request, &current))
        .map_err(|_| "encode failed".into())
}
