//! `lane-loop-line` (#2904 O5): the line a lane pastes to keep itself running. Claude Code is
//! the only host. The logic is `moai_cli::lane_loop`.

/// Print a lane's wake line for its host
///
/// Print the line a lane pastes to keep running without a leader (#2899 section B): Claude Code
/// `/loop 5m <duty>`. Claude Code is the only host. Reads and writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// `claude`, the only host.
    #[arg(long, value_parser = ["claude"])]
    host: String,
    /// The lane name as it appears in claims (`session=<lane>`).
    #[arg(long)]
    lane: String,
    /// `OWNER/REPO`, checked when given.
    #[arg(long)]
    repo: Option<String>,
    /// The duty the loop carries after `레인 <lane> — ` (default: the round's restart point and
    /// the resume-note rule).
    #[arg(long)]
    duty: Option<String>,
}

pub const OPERATION: &str = "lane.loop_line";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    // The command line refuses every other host word before this runs.
    let Args {
        host: _,
        lane,
        repo,
        duty,
    } = args;
    moai_cli::lane_loop::lines(&lane, repo.as_deref(), duty.as_deref())
}
