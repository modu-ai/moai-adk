//! `agents-plan`: previews the AGENTS.md managed section. Moved out of `src/main.rs` unchanged (#2834 R4a).

use std::path::PathBuf;

/// Preview the AGENTS.md managed section
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "agents.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    serde_json::to_value(moai_cli::agents_plan(
        &args.project,
        super::super::input::<moai_cli::AgentsRequest>()?,
    )?)
    .map_err(|_| "encode failed".into())
}
