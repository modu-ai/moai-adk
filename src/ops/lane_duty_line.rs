//! `lane-duty-line` (#2902 O3a): the line a leader types into a lane; the logic is
//! `moai_cli::lane_duty`.

/// Print the line a leader types into a lane
///
/// Print the one line a leader types into a lane's terminal to start its non-stop loop
/// (#2902 O3a): `/moai --mode-agent 레인 <lane> — <duty>` (`$moai` on Codex), with
/// ` 먼저: <task>` when `--first` names the first card. A task with a control, line-separator
/// or bidirectional character, or over 300 characters, is refused. Reads and writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The lane's name, as in its claim comments (1 to 64 of A-Z a-z 0-9 . _ -).
    #[arg(long)]
    lane: String,
    /// The first task the lane takes, one line of at most 300 characters.
    #[arg(long)]
    first: Option<String>,
    /// The lane's host: `claude` (default), `codex` or `omp`.
    #[arg(long, default_value = "claude", value_parser = ["claude", "codex", "omp"])]
    host: String,
}

pub const OPERATION: &str = "lane.duty_line";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let line = moai_cli::lane_duty::line(&args.lane, args.first.as_deref(), &args.host)?;
    Ok(serde_json::json!({ "line": line }))
}
