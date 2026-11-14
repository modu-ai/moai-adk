//! `agents-apply`: writes a previewed AGENTS.md managed section. Moved out of `src/main.rs` unchanged (#2834 R4a).

use std::path::PathBuf;

/// Write a previewed AGENTS.md managed section
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    #[arg(long)]
    plan_digest: String,
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "agents.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::agents_apply(
        &args.project,
        super::super::plan_input::<moai_cli::AgentsPlan>()?,
        &args.plan_digest,
        &args.permission_reference,
    )
}
