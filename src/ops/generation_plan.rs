//! `generation-plan` (03-C #1997): plans a project's generated agent and skill files read-only. Moved out of `src/main.rs` unchanged (#2834 R4a).

use std::path::PathBuf;

/// Preview the project's generated agent and skill files
///
/// Plan project-specific generated files read-only (03-C #1997): typed request on
/// stdin; judgment envelope on stdout; writes nothing.
///
/// The request is one JSON object. A complete one, with one project agent and one project
/// skill (`host` is required on every agent; the lists may be left out; the pick
/// `harness-only` plans only the agents and skills). It has no space inside a text, so the
/// help screen cannot fold it in the middle of one; a purpose is one line of plain text:
///
/// {"protocol_version":1,"language":"en","workflow_reference":".moai/workflows/project/","local_file":"none","workflow_id":"project","picks":["harness-only"],"native_roles":[{"role_id":"prj-test-runner","host":"claude","purpose":"Runs-the-tests"}],"project_skills":[{"id":"prj-checks","purpose":"Checks-before-done"}]}
///
/// Save the `result` of the answer as plan.json: `generation-apply` reads it on stdin.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "generation.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::profile::require_interviewed(&args.project)?;
    serde_json::to_value(moai_cli::generation::generation_plan(
        &args.project,
        super::super::input::<moai_cli::generation::GenerationRequest>()?,
    )?)
    .map_err(|_| "encode failed".into())
}
