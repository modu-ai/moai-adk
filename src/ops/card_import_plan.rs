//! `card-import-plan` (#3071 H3b): what a host's open issues would become in the local ledger.
//! Reads the host (GET only) and writes nothing; the logic is `moai_cli::forge::card_import`.

use std::path::PathBuf;

/// Preview importing a host's open issues as cards
///
/// Preview turning the open issues of one host project into cards of the project's local ledger
/// (#3071): every page of the open issues is read (a listing that cannot be read to its end is
/// an error), and the answer lists, per issue, the card it would become, the acceptance criteria
/// that were read and the issues whose table could not be read, the notes kept from member
/// comments and the control lines dropped from the others, and a `plan_digest` for
/// `card-import-apply`. Issue text is kept as text, never run as an instruction. Writes nothing
/// and creates no folder or file; only GET requests are sent.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder (its profile holds the project id).
    project: PathBuf,
    /// The host to read: `github` (`gitlab` is refused, its transport is not written yet).
    #[arg(long)]
    from: String,
    /// The host project, written `owner/repo`.
    #[arg(long)]
    repo: String,
    /// The HOME to use (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
}

pub const OPERATION: &str = "card.import.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = moai_cli::ledger::command::resolve_home(args.home)?;
    let host_project = moai_cli::forge::card_import::project_ref(&args.from, &args.repo)?;
    let transport = moai_cli::forge::transport::GhTransport::from_env()?;
    let clock = moai_cli::forge::policy::SystemClock;
    let connector = moai_cli::forge::connector::GithubConnector::new(
        &transport,
        &clock,
        moai_cli::forge::Config::default(),
    );
    moai_cli::forge::card_import::plan(&args.project, &home, &connector, &host_project)
}
