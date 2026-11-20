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
    // The plan answer carries fields only its reader uses (`n`, `work_flow_remote`); the
    // instructions feed that answer to this command unchanged, so exactly those two keys are
    // dropped before the strict plan type sees the input. The plan digest never covered them
    // (`profile::plan_digest` digests selected fields). The stripped value goes back through
    // the same typed-JSON contract, so every refusal keeps its documented message: a whole
    // envelope is still `invalid typed JSON input`, and so is any other unknown key.
    let mut plan: serde_json::Value = super::super::plan_input()?;
    if let Some(object) = plan.as_object_mut() {
        for key in ["n", "work_flow_remote"] {
            object.remove(key);
        }
    }
    let bytes = serde_json::to_vec(&plan).map_err(|_| "encode failed".to_string())?;
    let plan: moai_cli::Plan = moai_cli::contract::typed_json(&bytes)?;
    moai_cli::apply(
        &args.project,
        plan,
        &args.plan_digest,
        &args.permission_reference,
    )
}
