//! `goal-apply` (#2679): write a previewed change of the standing goal.

use std::path::PathBuf;

/// Write a previewed change of the standing goal (#2679), only with the `plan_digest` of a
/// fresh `goal-plan` for the same request and the same file. An `unchanged` or `none` plan
/// writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// A folder inside the project's git repository (any worktree of the clone).
    project: PathBuf,
    /// The same JSON request the preview read.
    #[arg(long)]
    request: PathBuf,
    /// The preview's `plan_digest`.
    #[arg(long)]
    plan_digest: String,
    /// `YYYY-MM-DDTHH:MM:SSZ`; the system clock when absent.
    #[arg(long)]
    now: Option<String>,
}

pub const OPERATION: &str = "goal.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let request = moai_cli::goal::read_request(&args.request)?;
    let now = args.now.unwrap_or_else(moai_cli::audit::utc_now);
    moai_cli::goal::apply(&args.project, &request, &args.plan_digest, &now)
}
