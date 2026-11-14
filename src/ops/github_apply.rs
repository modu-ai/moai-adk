//! `github-apply` (08-B #2013): applies an approved GitHub template plan. Moved out of `src/main.rs` unchanged (#2834 R4a).

use std::path::PathBuf;

/// Write previewed GitHub templates
///
/// Apply an approved GitHub template plan (08-B #2013): the exact plan on stdin,
/// `--plan-digest` and `--permission-reference`.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    #[arg(long)]
    plan_digest: String,
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "github.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::github_templates::github_apply(
        &args.project,
        super::super::plan_input::<moai_cli::github_templates::GithubPlan>()?,
        &args.plan_digest,
        &args.permission_reference,
    )
}
