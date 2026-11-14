//! `codex-config-apply`: applies a previewed project Codex config. Moved out of `src/main.rs` unchanged (#2834 R4a).

use std::path::PathBuf;

/// Write previewed Codex config keys
///
/// Apply a previewed project config with the same stdin, digest and permission reference.
/// User HOME and CODEX_HOME config targets are refused; trust is never changed.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    #[arg(long)]
    plan_digest: String,
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "codex.config_apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = super::super::codex_config_home(&args.project)?;
    let desired = super::super::input()?;
    moai_cli::hosts::codex_config::config_apply(
        &args.project,
        &home,
        &desired,
        &args.plan_digest,
        &args.permission_reference,
    )
}
