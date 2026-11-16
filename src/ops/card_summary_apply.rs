//! `card-summary-apply` (#3071 H4, AC7): post the planned summary once. Sends one POST at most;
//! the logic is `moai_cli::forge::card_export`.

use std::path::PathBuf;

/// Post a card's summary to its issue, once
///
/// Post the text shown by `card-summary-plan` to the card's issue or pull request (#3071), only
/// when the plan is still the same (`plan_digest`) and a permission reference is given. The host
/// is read first: when a member's comment already carries the card's marker nothing is posted and
/// the answer is `already_posted`. This is the only command that writes to the host.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder (its profile holds the project id).
    project: PathBuf,
    /// The card, written `C-<number>`.
    #[arg(long)]
    card: String,
    /// The host to post to: `github`.
    #[arg(long)]
    to: String,
    /// The same target number the plan used.
    #[arg(long)]
    number: Option<u64>,
    /// The same HOME the plan read (default `$HOME`).
    #[arg(long)]
    home: Option<PathBuf>,
    /// The plan's `plan_digest`.
    #[arg(long)]
    plan_digest: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "card.summary.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = moai_cli::ledger::command::resolve_home(args.home)?;
    let transport = moai_cli::forge::transport::GhTransport::from_env()?;
    let clock = moai_cli::forge::policy::SystemClock;
    let connector = moai_cli::forge::connector::GithubConnector::new(
        &transport,
        &clock,
        moai_cli::forge::Config::default(),
    );
    moai_cli::forge::card_export::apply(
        &moai_cli::forge::card_export::Request {
            project: &args.project,
            home: &home,
            card: &args.card,
            to: &args.to,
            number: args.number,
        },
        &connector,
        &args.plan_digest,
        &args.permission_reference,
    )
}
