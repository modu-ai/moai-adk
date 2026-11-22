//! `home-routing-apply` (#2638): apply a previewed routing plan.

use std::path::PathBuf;

/// Write a previewed routing section into HOME
///
/// Apply a previewed MoAI routing plan (#2638): the routing file and the flow rule in each
/// previewed Claude config folder, and the earlier block's removal when it was picked, with a
/// receipt holding the before bytes. The digest must match a fresh preview made with the same
/// `--config-dir` and `--remove-old-block`.
#[derive(clap::Args)]
pub struct Args {
    /// User home root.
    home: PathBuf,
    /// Skill package folder carrying `references/home/routing.md` and `references/home/flow.md`.
    package: PathBuf,
    /// The preview's `plan_digest`.
    #[arg(long)]
    plan_digest: String,
    /// Caller-supplied permission reference (recorded; not an authorization verifier).
    #[arg(long)]
    permission_reference: String,
    /// The extra Claude config folders the preview was made with (absolute path); repeat for several.
    #[arg(long = "config-dir")]
    config_dir: Vec<PathBuf>,
    /// The removal of the earlier block, if the preview was made with it.
    #[arg(long)]
    remove_old_block: bool,
}

pub const OPERATION: &str = "home.routing_apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let opts = moai_cli::home_routing::Options {
        config_dirs: args.config_dir,
        remove_old_block: args.remove_old_block,
    };
    moai_cli::home_routing::apply_with(
        &args.home,
        &args.package,
        &["claude"],
        &opts,
        &args.plan_digest,
        &args.permission_reference,
    )
}
