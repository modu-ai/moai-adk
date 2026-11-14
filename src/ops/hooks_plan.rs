//! `hooks-plan` (rc.2 I1 #2238): previews the MoAI hook set for one host. Moved out of `src/main.rs` unchanged (#2834 R4b).

use std::path::PathBuf;

/// Preview the MoAI hook set for one host
///
/// Preview the MoAI hook set for one host in PROJECT (rc.2 I1 #2238); writes nothing.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    /// `claude` (.claude/settings.json `hooks`) or `codex` (.codex/hooks.json).
    #[arg(long)]
    host: String,
    #[command(flatten)]
    lane: super::super::LaneLoopArgs,
}

pub const OPERATION: &str = "hooks.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = std::env::var_os("HOME").map(PathBuf::from);
    let mut out = moai_cli::hooks::plan_with(
        &args.project,
        &args.host,
        home.as_deref(),
        args.lane.lane_loop(),
    )?;
    super::super::merge_binary_state(&mut out);
    // Outside the plan digest: the #2391 hook state from local evidence, read only.
    let path_env = std::env::var("PATH").unwrap_or_default();
    out["hook_status"] =
        moai_cli::hooks::status(&args.project, &args.host, home.as_deref(), &path_env)?;
    Ok(out)
}
