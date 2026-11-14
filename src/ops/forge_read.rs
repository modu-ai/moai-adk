//! `forge-read`: reads GitHub Issues and comments as tasks (GET only). Moved out of
//! `src/main.rs` unchanged (#2834 R2b).

use std::path::PathBuf;

/// Read GitHub Issues and comments as tasks
///
/// Read GitHub Issues and comments as TODO (GET only; typed JSON request on stdin,
/// envelope on stdout). Poll state: `<PROJECT>/.moai-cli/forge/poll-state.json`.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "forge.read";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    // Progress lines on a terminal's stderr, as `shows_progress` turned on for it in main.rs.
    moai_cli::progress::install_stderr();
    moai_cli::forge::forge_read_cli(
        &args.project,
        super::super::input::<moai_cli::forge::ReadRequest>()?,
        &moai_cli::forge::transport::GhTransport::from_env()?,
        &moai_cli::forge::policy::SystemClock,
    )
}
