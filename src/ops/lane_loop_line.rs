//! `lane-loop-line` (#2904 O5): the line a lane pastes to keep itself running on its host.
//! The logic is `moai_cli::lane_loop`.

use std::path::PathBuf;

/// Print a lane's wake line for its host
///
/// Print the line a lane pastes to keep running without a leader (#2899 section B): Claude Code
/// `/loop 5m /moai --mode-agent <duty>`; Codex CLI `codex features enable goals`, then
/// `/goal <end condition>`, with the Stop hook the lane needs in `.codex/hooks.json` and its
/// state read from the project (`LANE-BLOCKED <lane> wake …` when it is not installed); the
/// third host `--goal "<end condition>"`. Reads the project's hook file only; writes nothing.
/// Live sessions of the Codex and third-host lines are not measured (`verified: not_run`).
#[derive(clap::Args)]
pub struct Args {
    /// `claude`, `codex` or `omp`.
    #[arg(long)]
    host: String,
    /// The lane name as it appears in claims (`session=<lane>`).
    #[arg(long)]
    lane: String,
    /// `OWNER/REPO` (required for codex: its Stop hook names it).
    #[arg(long)]
    repo: Option<String>,
    /// The duty a Claude loop carries after `레인 <lane> — ` (default: the round's restart point
    /// and the resume-note rule).
    #[arg(long)]
    duty: Option<String>,
    /// The project whose `.codex/hooks.json` is read (codex).
    #[arg(long, default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "lane.loop_line";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    use moai_cli::lane_loop as l;
    let host = l::Host::parse(&args.host)?;
    let mut v = l::lines(host, &args.lane, args.repo.as_deref(), args.duty.as_deref())?;
    if host == l::Host::Codex {
        let repo = args.repo.as_deref().unwrap_or_default();
        let state = l::codex_hook_state(&args.project, &args.lane, repo);
        v["stop_hook"]["state"] = state.into();
        // Codex keeps trust in its own records (`/hooks`); it is never read here.
        v["stop_hook"]["trust"] = "unknown".into();
        v["blocked_line"] = l::codex_blocked_line(&args.lane, state).into();
    }
    Ok(v)
}
