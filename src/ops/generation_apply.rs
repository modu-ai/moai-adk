//! `generation-apply` (03-C #1997): applies an approved generation plan. Moved out of `src/main.rs` unchanged (#2834 R4a).

use std::path::PathBuf;

/// Write previewed generated agent and skill files
///
/// Apply an approved generation plan (03-C #1997): the exact plan on stdin,
/// `--plan-digest` and `--permission-reference`.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    #[arg(long)]
    plan_digest: String,
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "generation.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::generation::generation_apply(
        &args.project,
        super::super::plan_input::<moai_cli::generation::GenerationPlan>()?,
        &args.plan_digest,
        &args.permission_reference,
    )
}
