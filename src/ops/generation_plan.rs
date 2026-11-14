//! `generation-plan` (03-C #1997): plans a project's generated agent and skill files read-only. Moved out of `src/main.rs` unchanged (#2834 R4a).

use std::path::PathBuf;

/// Preview the project's generated agent and skill files
///
/// Plan project-specific generated files read-only (03-C #1997): typed request on
/// stdin; judgment envelope on stdout; writes nothing.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "generation.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    serde_json::to_value(moai_cli::generation::generation_plan(
        &args.project,
        super::super::input::<moai_cli::generation::GenerationRequest>()?,
    )?)
    .map_err(|_| "encode failed".into())
}
