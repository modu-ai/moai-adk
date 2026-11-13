//! `home-routing-apply` (#2638): apply a previewed routing section.

use std::path::PathBuf;

/// Apply a previewed MoAI routing section (#2638): only the managed block changes, with a
/// receipt holding the before bytes. The digest must match a fresh preview.
#[derive(clap::Args)]
pub struct Args {
    /// User home root.
    home: PathBuf,
    /// Skill package folder carrying `references/home/routing.md`.
    package: PathBuf,
    /// The preview's `plan_digest`.
    #[arg(long)]
    plan_digest: String,
    /// Caller-supplied permission reference (recorded; not an authorization verifier).
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "home.routing_apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::home_routing::apply(
        &args.home,
        &args.package,
        &args.plan_digest,
        &args.permission_reference,
    )
}
