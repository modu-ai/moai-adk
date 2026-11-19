//! `project-plan`: previews the project profile. Moved out of `src/main.rs` unchanged (#2834 R4a).

use std::path::PathBuf;

/// Preview the project profile
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "project.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    // REQ-03A-08: an unknown input key or a non-v2 schema_version is refused with
    // the profile message; no input value is echoed.
    let profile: moai_cli::Profile =
        serde_json::from_value(super::super::input::<serde_json::Value>()?)
            .map_err(|_| "invalid or nonportable profile; no values were logged".to_string())?;
    let work_flow = profile.work_flow.clone();
    let mut answer = serde_json::to_value(moai_cli::plan(&args.project, profile)?)
        .map_err(|_| "encode failed".to_string())?;
    // C-105 slice 2 (SPEC-20261119-plan-remote-branch), audit B4: the remote check
    // belongs to this entry alone — `project-update-plan` shares `plan()` but must
    // not run `git ls-remote` (a network call) in its preview. The value this op
    // returns becomes the envelope's `result`, so the object attaches here directly.
    if let Some(wf) = moai_cli::profile::plan_work_flow_remote(&args.project, work_flow.as_ref()) {
        answer["work_flow_remote"] = wf;
    }
    Ok(answer)
}
