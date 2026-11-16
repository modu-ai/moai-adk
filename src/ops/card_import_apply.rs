//! `card-import-apply` (#3071 H3b): write the previewed import into the local ledger. Reads the
//! host again (GET only); the logic is `moai_cli::forge::card_import`.

use std::path::PathBuf;

/// Import a host's open issues as cards, only the previewed set
///
/// Write the open issues of one host project into the project's local ledger as cards (#3071),
/// only when the issues the host gives now have the `plan_digest` of the preview the user was
/// shown, and a permission reference is given. An issue that already has a card (same host,
/// project and number) is left exactly as it is. All cards are written in one transaction or
/// none. The ledger must already exist (`ledger-register-apply`); without one nothing is read or
/// written. Cards are drafts: nothing is reserved, started or closed, and no issue is closed or
/// changed on the host.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder (its profile holds the project id).
    project: PathBuf,
    /// The host the preview read: `github`.
    #[arg(long)]
    from: String,
    /// The host project, written `owner/repo`.
    #[arg(long)]
    repo: String,
    /// The same HOME the preview read (default `$HOME`).
    #[arg(long)]
    home: Option<PathBuf>,
    /// The preview's `plan_digest`.
    #[arg(long)]
    plan_digest: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "card.import.apply";

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
    moai_cli::forge::card_import::apply(
        &args.project,
        &home,
        &connector,
        &host_project,
        &args.plan_digest,
        &args.permission_reference,
    )
}
