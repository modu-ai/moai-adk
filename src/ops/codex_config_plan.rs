//! `codex-config-plan`: previews managed dotted keys in PROJECT/.codex/config.toml. Moved out of `src/main.rs` unchanged (#2834 R4a).

use std::path::PathBuf;

/// Preview the managed Codex config keys
///
/// Preview managed dotted keys from stdin in PROJECT/.codex/config.toml; never writes. A key
/// the user set differently is a conflict, listed in `diff` with both values. With
/// `--replace-with-backup` the conflict is previewed as a replacement: the keys, the copy of the
/// file that would be kept and the digest the matching apply needs.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    /// Preview replacing the keys the user set differently, after keeping a copy of the file
    #[arg(long)]
    replace_with_backup: bool,
}

pub const OPERATION: &str = "codex.config_plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = super::super::codex_config_home(&args.project)?;
    let desired = super::super::input()?;
    moai_cli::hosts::codex_config::config_plan_with(
        &args.project,
        &home,
        &desired,
        args.replace_with_backup,
    )
}
