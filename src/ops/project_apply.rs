//! `project-apply`: writes a previewed project profile. Moved out of `src/main.rs` unchanged (#2834 R4a).

use std::path::PathBuf;

/// Write a previewed project profile
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    #[arg(long)]
    plan_digest: String,
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "project.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::apply(
        &args.project,
        super::super::plan_input::<moai_cli::Plan>()?,
        &args.plan_digest,
        &args.permission_reference,
    )
}
