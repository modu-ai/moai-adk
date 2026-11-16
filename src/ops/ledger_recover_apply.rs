//! `ledger-recover-apply` (#3068 P6): recover a damaged project ledger, only on an identical
//! fresh preview. The logic is `moai_cli::ledger_recover`.

use std::path::PathBuf;

/// Recover the previewed project ledger
///
/// Build the new ledger at `<ledger>.recovering`, check it (`integrity_check`,
/// `foreign_key_check`), keep the damaged file under `<ledger>.damaged-<UTC>` and rename the new
/// file over the ledger path (#3068 P6). It runs only when the preview recomputed from the same
/// inputs has the same `plan_digest` and a permission reference is given. Nothing is deleted: the
/// damaged file stays, byte for byte, under its new name. Running it again with the same plan
/// answers `unchanged` and writes nothing. A usable ledger that is not this plan's result is
/// never replaced.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder (its profile holds the project id).
    project: PathBuf,
    /// The same source the preview read.
    #[arg(long)]
    source: PathBuf,
    /// The same HOME the preview read (default `$HOME`).
    #[arg(long)]
    home: Option<PathBuf>,
    /// The same claim journal folder as the preview.
    #[arg(long)]
    journal: Option<PathBuf>,
    /// The same observation file as the preview.
    #[arg(long)]
    observations: Option<PathBuf>,
    /// The same `--now` as the preview, when the preview was given one.
    #[arg(long)]
    now: Option<String>,
    /// The preview's `plan_digest`.
    #[arg(long)]
    plan_digest: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "ledger.recover.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = moai_cli::ledger::command::resolve_home(args.home)?;
    let now_text = args.now.unwrap_or_else(moai_cli::audit::utc_now);
    let now = moai_cli::forge::policy::parse_utc(&now_text)
        .ok_or_else(|| format!("invalid --now `{now_text}`: expected YYYY-MM-DDTHH:MM:SSZ"))?;
    moai_cli::ledger_recover::apply(
        &moai_cli::ledger_recover::Inputs {
            project: &args.project,
            home: &home,
            source: &args.source,
            journal: args.journal.as_deref(),
            observations: args.observations.as_deref(),
            now,
        },
        &args.plan_digest,
        &args.permission_reference,
        &moai_cli::ledger_recover::GitProbe,
    )
}
