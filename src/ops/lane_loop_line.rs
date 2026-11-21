//! `lane-loop-line` (#2904 O5): the line a lane pastes to keep itself running on its host.
//! The logic is `moai_cli::lane_loop`.

/// Print a lane's wake line for its host
///
/// Print the line a lane pastes to keep running without a leader: Claude Code
/// `/loop 5m /moai --mode-agent <duty>`. Claude Code is the only host: any other `--host` word
/// is refused with one line. Reads and writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// `claude`, the only host.
    #[arg(long)]
    host: String,
    /// The lane name as it appears in claims (`session=<lane>`).
    #[arg(long)]
    lane: String,
    /// `OWNER/REPO`, checked when given.
    #[arg(long)]
    repo: Option<String>,
    /// The duty a Claude loop carries after `레인 <lane> — ` (default: the round's restart point
    /// and the resume-note rule).
    #[arg(long)]
    duty: Option<String>,
}

pub const OPERATION: &str = "lane.loop_line";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    use moai_cli::lane_loop as l;
    let host = l::Host::parse(&args.host)?;
    l::lines(host, &args.lane, args.repo.as_deref(), args.duty.as_deref())
}
