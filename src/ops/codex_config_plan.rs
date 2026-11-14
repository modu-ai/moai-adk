//! `codex-config-plan`: previews managed dotted keys in PROJECT/.codex/config.toml. Moved out of `src/main.rs` unchanged (#2834 R4a).

use std::path::PathBuf;

/// Preview the managed Codex config keys
///
/// Preview managed dotted keys from stdin in PROJECT/.codex/config.toml; never writes.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "codex.config_plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = super::super::codex_config_home(&args.project)?;
    let desired = super::super::input()?;
    moai_cli::hosts::codex_config::config_plan(&args.project, &home, &desired)
}
