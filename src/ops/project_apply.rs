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
    // (`profile::plan_digest` digests selected fields), and any other unknown key is still
    // refused by the plan type itself.
    let mut plan: serde_json::Value = super::super::plan_input()?;
    if !plan.is_object() {
        return Err("plan input refused: not an object".into());
    }
    for key in ["n", "work_flow_remote"] {
        plan.as_object_mut().expect("checked above").remove(key);
    }
    let plan: moai_cli::Plan =
        serde_json::from_value(plan).map_err(|e| format!("plan input refused: {e}"))?;
    moai_cli::apply(
        &args.project,
        plan,
        &args.plan_digest,
        &args.permission_reference,
    )
}
