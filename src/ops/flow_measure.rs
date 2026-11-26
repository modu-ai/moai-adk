//! `flow-measure` (SPEC-20261125-flow-measure): how long the stages of the card flow took. The
//! logic is `moai_cli::flow_measure`.

use std::path::PathBuf;

/// How long the stages of the card flow took
///
/// Reads the ledger only and writes nothing. With card names (`C-<number>`) it prints per card the
/// seconds from submit to verdict and from pass to done, the failed verdicts, the auditor session,
/// the head and the reservation, and the P50 and P90 of pass to done (nearest rank, raw values
/// too). With `--hours N` it prints the window: the `card.add`, `card.landed` and `card.finish`
/// events counted apart, P50 and P90 and the sample count of each stage, the share of first
/// verdicts that failed, and the `LANE-BLOCKED` notes by kind. A span whose end is not in the
/// ledger is `null`, never zero. The answer is JSON; its `table` lines are the same numbers for a
/// person (`jq -r '.result.table[]'`). A project without a ledger answers `ledger: "absent"`
/// (exit 0) and nothing is created.
#[derive(clap::Args)]
pub struct Args {
    /// The cards to measure, each `C-<number>`.
    cards: Vec<String>,
    /// Count the events of the last N hours (a whole number, 1 or more) and summarise the stages.
    #[arg(long)]
    hours: Option<String>,
    /// The project folder (its profile holds the project id).
    #[arg(long, default_value = ".")]
    project: PathBuf,
    /// The HOME to read (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
    /// The end of the window in unix seconds (default: the clock of the ledger database).
    #[arg(long)]
    now: Option<i64>,
}

pub const OPERATION: &str = "flow.measure";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let request =
        moai_cli::flow_measure::parse_request(&args.cards, args.hours.as_deref(), args.now)?;
    moai_cli::flow_measure::measure(
        &args.project,
        &moai_cli::ledger::command::resolve_home(args.home)?,
        &request,
    )
}
