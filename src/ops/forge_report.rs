//! `forge-report`: creates or updates the one report comment of a run. Moved out of
//! `src/main.rs` unchanged (#2834 R2b), with its own stdin read and 64 KiB message.

use std::io::Read;
use std::path::PathBuf;

/// Create or update a run's report comment
///
/// Create or update the one report comment of a run on its Issue (typed JSON payload on
/// stdin; envelope on stdout). Large output goes to artifact references, not the body.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "forge.report";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let mut bytes = Vec::new();
    std::io::stdin()
        .take(65537)
        .read_to_end(&mut bytes)
        .map_err(|_| "stdin read failed")?;
    if bytes.len() > 65536 {
        return Err(
            "input exceeds 64 KiB; put large output in an artifact reference, not the payload"
                .into(),
        );
    }
    let payload: moai_cli::forge::report::ReportPayload =
        serde_json::from_slice(&bytes).map_err(|_| "invalid typed JSON input")?;
    moai_cli::forge::report::report_cli(
        &args.project,
        payload,
        &args.permission_reference,
        &moai_cli::forge::transport::GhTransport::from_env()?,
        &moai_cli::forge::policy::SystemClock,
    )
}
