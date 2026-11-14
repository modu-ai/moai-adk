//! `github-plan` (08-B #2013): plans a project's GitHub templates read-only. Moved out of `src/main.rs` unchanged (#2834 R4a).

use std::path::PathBuf;

/// Preview the project's GitHub templates
///
/// Plan the GitHub templates of a project (08-B #2013): typed request on stdin
/// (`github_issues` opts in); read-only.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "github.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    serde_json::to_value(moai_cli::github_templates::github_plan(
        &args.project,
        super::super::input::<moai_cli::github_templates::GithubRequest>()?,
    )?)
    .map_err(|_| "encode failed".into())
}
