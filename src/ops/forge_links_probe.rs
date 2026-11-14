//! `forge-links-probe` (#2172): probes the GitHub sub-issue and dependency APIs (GET only).
//! Moved out of `src/main.rs` unchanged (#2834 R2b).

use std::path::PathBuf;

/// Probe the GitHub sub-issue and dependency APIs
///
/// Probe whether the sub-issue and dependency APIs answer for one Issue and plan the
/// links: API write intents, or `Parent:`/`Blocked-By:` body lines (GET only; #2172).
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "forge.links_probe";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::forge::links::links_probe_cli(
        &args.project,
        super::super::input::<moai_cli::forge::links::LinksRequest>()?,
        &moai_cli::forge::transport::GhTransport::from_env()?,
        &moai_cli::forge::policy::SystemClock,
    )
}
