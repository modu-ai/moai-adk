//! `home-routing-plan` (#2638): preview of the MoAI routing files in the Claude config folders;
//! writes nothing.

use std::path::PathBuf;

/// Preview the routing section in HOME instructions
///
/// Preview the MoAI routing files (#2638) in each Claude config folder: the routing file
/// `rules/moai.md` and the flow rule `rules/moai/flow.md` are planned together (create, merge,
/// unchanged or conflict), the earlier block in the folder's `CLAUDE.md` is offered for removal
/// only with `--remove-old-block`, and the hosts not written are listed with the reason.
/// Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// User home root whose Claude config folder (`.claude`, or `CLAUDE_CONFIG_DIR`) is inspected.
    home: PathBuf,
    /// Skill package folder carrying `references/home/routing.md` and `references/home/flow.md`.
    package: PathBuf,
    /// Plan one more Claude config folder (absolute path); repeat for several.
    #[arg(long = "config-dir")]
    config_dir: Vec<PathBuf>,
    /// Offer the removal of the earlier `moai-routing` block from each folder's `CLAUDE.md`.
    #[arg(long)]
    remove_old_block: bool,
}

pub const OPERATION: &str = "home.routing_plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let opts = moai_cli::home_routing::Options {
        config_dirs: args.config_dir,
        remove_old_block: args.remove_old_block,
    };
    serde_json::to_value(moai_cli::home_routing::plan_with(
        &args.home,
        &args.package,
        &["claude"],
        &opts,
    )?)
    .map_err(|_| "encode failed".into())
}
