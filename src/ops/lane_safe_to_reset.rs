//! `lane-safe-to-reset` (#2669): whether a lane may be cleared or redispatched without losing
//! work. Reads only; the logic is `moai_cli::lane_reset`.

use std::path::PathBuf;

/// Before a lane's session is cleared or its card dispatched elsewhere (#2669): map the lane
/// to its worktrees, count their uncommitted, unpushed and stashed work, and check the
/// caller's observations of the lane's open claims (each with a resume note), session folder
/// and screen (background jobs). `safe` only when every check ran clean; any finding is
/// `not_safe`; a check that could not run is `unknown` with its reason. Reads only.
#[derive(clap::Args)]
pub struct Args {
    /// Any checkout of the repository (the primary one or a worktree).
    project: PathBuf,
    /// The lane's name, as in its claim comments (`agent-27`).
    #[arg(long)]
    lane: String,
    /// JSON file: `schema_version` 1 and optional `claims` {source, items [{issue,
    /// claim_comment, resume_note}]}, `session_cwd` {source, path}, `screen` {source,
    /// observed_at, background_jobs}. A missing part leaves its check unknown.
    #[arg(long)]
    observations: Option<PathBuf>,
    /// `YYYY-MM-DDTHH:MM:SSZ` to judge the screen observation's age against; the system
    /// clock when absent.
    #[arg(long)]
    now: Option<String>,
}

pub const OPERATION: &str = "lane.safe_to_reset";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let obs = match &args.observations {
        None => moai_cli::lane_reset::Observations {
            schema_version: 1,
            ..Default::default()
        },
        Some(path) => {
            let text = std::fs::read_to_string(path)
                .map_err(|e| format!("observations file unavailable: {}: {e}", path.display()))?;
            serde_json::from_str(&text).map_err(|e| format!("invalid observations: {e}"))?
        }
    };
    let now_text = args.now.unwrap_or_else(moai_cli::audit::utc_now);
    let now = moai_cli::forge::policy::parse_utc(&now_text)
        .ok_or_else(|| format!("invalid --now `{now_text}`: expected YYYY-MM-DDTHH:MM:SSZ"))?;
    moai_cli::lane_reset::check(&args.project, &args.lane, &obs, now)
}
