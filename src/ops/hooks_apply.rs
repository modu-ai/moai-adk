//! `hooks-apply`: writes a previewed hook set and its receipt. Moved out of `src/main.rs` unchanged (#2834 R4b).

use std::path::PathBuf;

/// Write a previewed hook set
///
/// Write a previewed hook set and its receipt; requires the plan digest.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    #[arg(long)]
    host: String,
    #[arg(long)]
    plan_digest: String,
    #[arg(long)]
    permission_reference: String,
    #[command(flatten)]
    lane: super::super::LaneLoopArgs,
}

pub const OPERATION: &str = "hooks.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = std::env::var_os("HOME").map(PathBuf::from);
    let mut out = moai_cli::hooks::apply_with(
        &args.project,
        &args.host,
        home.as_deref(),
        &args.plan_digest,
        &args.permission_reference,
        args.lane.lane_loop(),
    )?;
    super::super::merge_binary_state(&mut out);
    Ok(out)
}
