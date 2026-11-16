//! `leader-pulse` (#3015 C5): the text and the `orca` command line a leader pastes every N
//! seconds. The logic is `moai_cli::leader_pulse`.

/// Print the periodic round instruction and its orca command
///
/// Check one round instruction for the leader's own terminal and print it with the interval and
/// the `orca terminal send` command line (#3015 C5). It runs nothing: the skill waits the
/// interval and runs the command once per round. It never creates a terminal, and a pulse whose
/// turn did not start is reported, not sent again. Live Codex rounds are not measured
/// (`verified: not_run`). Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The leader's terminal handle (from `orca terminal list`).
    #[arg(long)]
    terminal: String,
    /// The one-line instruction typed into that terminal.
    #[arg(long)]
    text: String,
    /// Seconds between two pulses (60 to 3600).
    #[arg(long, default_value_t = moai_cli::leader_pulse::DEFAULT_INTERVAL_SECS)]
    interval_secs: u64,
    /// Stop after this many pulses (1 to 10000; leave out to run until stopped).
    #[arg(long)]
    rounds: Option<u32>,
}

pub const OPERATION: &str = "leader.pulse";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::leader_pulse::build(&args.terminal, &args.text, args.interval_secs, args.rounds)
}
