//! `home-routing-plan` (#2638): preview of the MoAI routing files in the Claude config folders;
//! writes nothing.

use std::path::PathBuf;

/// Preview the routing section in HOME instructions
///
/// Preview the MoAI routing files (#2638) in `~/.claude`, the one place moai keeps them: the
/// routing file `rules/moai.md` and the flow rule `rules/moai-flow.md` are planned together
/// (create, merge, unchanged or conflict), the earlier block in `CLAUDE.md` is offered for
/// removal only with `--remove-old-block`, and the hosts not written are listed with the reason.
/// The folder `CLAUDE_CONFIG_DIR` names and each `--config-dir` get a row that writes nothing
/// (`linked` when its `rules` is a link to `~/.claude/rules`, else `not_linked`). Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// User home root whose `.claude` folder holds the one copy (`CLAUDE_CONFIG_DIR` adds a row).
    home: PathBuf,
    /// Skill package folder carrying `references/home/routing.md` and `references/home/flow.md`.
    package: PathBuf,
    /// Report one more Claude config folder (absolute path) that reads the copy; repeat for several.
    #[arg(long = "config-dir")]
    config_dir: Vec<PathBuf>,
    /// Offer the removal of the earlier `moai-routing` block from `~/.claude/CLAUDE.md`.
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
