//! `card-summary-plan` (#3071 H4): the summary a card would post to its issue. Reads the ledger
//! only; the logic is `moai_cli::forge::card_export`.

use std::path::PathBuf;

/// Preview the summary a card would post to its issue
///
/// Show the exact text, the target issue or pull request, the marker that keeps a second post
/// from happening, and a `plan_digest` for `card-summary-apply` (#3071). The card must have come
/// from the host named by `--to`. Reads the project's local ledger only: no request is sent and
/// nothing is written.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder (its profile holds the project id).
    project: PathBuf,
    /// The card, written `C-<number>`.
    #[arg(long)]
    card: String,
    /// The host to post to: `github`. It must be the host the card came from.
    #[arg(long)]
    to: String,
    /// Post to this issue or pull request number instead of the card's own issue.
    #[arg(long)]
    number: Option<u64>,
    /// The HOME to use (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
}

pub const OPERATION: &str = "card.summary.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = moai_cli::ledger::command::resolve_home(args.home)?;
    moai_cli::forge::card_export::plan(&moai_cli::forge::card_export::Request {
        project: &args.project,
        home: &home,
        card: &args.card,
        to: &args.to,
        number: args.number,
    })
}
