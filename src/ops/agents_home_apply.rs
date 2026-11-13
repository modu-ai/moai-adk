//! `agents-home-apply` (#2592 S2): apply a previewed base agent install.

use std::path::PathBuf;

/// Apply a previewed base agent install (#2592): only create and update are written, with a
/// receipt holding the before bytes. The digest must match a fresh preview.
#[derive(clap::Args)]
pub struct Args {
    /// User home root.
    home: PathBuf,
    /// Skill package folder whose `references/` the agents' `reads` lists name.
    package: PathBuf,
    /// Agent source folder outside the skill (`agents/` beside `skills/`, #2721).
    agents: PathBuf,
    /// The preview's `plan_digest`.
    #[arg(long)]
    plan_digest: String,
    /// Caller-supplied permission reference (recorded; not an authorization verifier).
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "agents.home_apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::agent_install::apply(
        &args.home,
        &args.package,
        &args.agents,
        &args.plan_digest,
        &args.permission_reference,
    )
}
