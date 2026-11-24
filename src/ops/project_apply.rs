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

/// The one key `project-plan` adds to its `result` beside the plan itself: the remote check of the
/// integration branch (SPEC-20261119-plan-remote-branch). It is a warning for the reader of the
/// plan, not a part of the plan, so the apply takes the `result` as printed and drops it.
const PLAN_ONLY_KEY: &str = "work_flow_remote";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    // SPEC-20261124-moai-project-interview REQ-11/12: only that key is dropped; any other
    // unknown key still fails the typed `Plan` below as invalid typed JSON input.
    let mut input = super::super::plan_input::<serde_json::Value>()?;
    if let Some(object) = input.as_object_mut() {
        object.remove(PLAN_ONLY_KEY);
    }
    let plan: moai_cli::Plan =
        serde_json::from_value(input).map_err(|_| "invalid typed JSON input".to_string())?;
    moai_cli::apply(
        &args.project,
        plan,
        &args.plan_digest,
        &args.permission_reference,
    )
}
