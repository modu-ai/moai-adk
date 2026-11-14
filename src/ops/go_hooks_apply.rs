//! `go-hooks-apply` (#2877): replace a project's Go-generation hooks after the user confirmed
//! the plan. The logic is `moai_cli::go_hooks`.

use std::path::PathBuf;

/// Replace Go-era project hooks
///
/// Replace the Go-generation hooks of a project (#2877) after the user confirmed `go-hooks-plan`:
/// the scripts move into a receipt folder, the Go entries leave `.claude/settings.json` (other
/// keys stay), and the MoAI hook set is installed the way `hooks-apply` installs it. A failure
/// puts everything back; `go-hooks-revert` undoes a finished apply.
#[derive(clap::Args)]
pub struct Args {
    /// The `plan_digest` of the confirmed `go-hooks-plan`.
    #[arg(long)]
    plan_digest: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
    /// The project (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "go_hooks.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = std::env::var_os("HOME").map(PathBuf::from);
    moai_cli::go_hooks::apply(
        &args.project,
        home.as_deref(),
        &args.plan_digest,
        &args.permission_reference,
    )
}
