//! `worktree-install-plan` (#3010 P6a): read-only preview of the command that would install
//! only the packages one profile scope names. Nothing is run and nothing is written.

use std::path::PathBuf;

/// Preview a scoped package install
///
/// Read-only preview of the install command for one scope of the project profile (`scopes`
/// key, #3010). The command is shown, never run: installing fetches packages over the network.
/// What a package manager's documentation does not say is listed under `unverified`.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    /// Name of a scope declared in the project profile.
    #[arg(long)]
    scope: String,
}

pub const OPERATION: &str = "worktree.install.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::worktree_install_plan::plan(&args.project, &args.scope)
}
