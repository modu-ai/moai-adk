//! `ledger-recover-plan` (#3068 P6): what recovering a damaged project ledger would build. Reads
//! only; the logic is `moai_cli::ledger_recover`.

use std::path::PathBuf;

/// Preview recovering a damaged project ledger
///
/// Preview building a new ledger from the best readable copy of a damaged one (#3068 P6): the
/// file itself, its `.bak` copy, or a ledger-shaped file you name with `--source`. Each table is
/// read with tolerance; a read error stops that table at its last good row and the answer says
/// how many rows were lost. A reservation comes back `held` only when the claim journal says its
/// lease is live, its worktree exists on its branch, and the observation file shows its branch on
/// the remote; every other active reservation becomes `held_for_review`, which still keeps the
/// card from being taken. The answer carries a `plan_digest` for `ledger-recover-apply`. Writes
/// nothing, not even to the source; the damaged file is never deleted or overwritten.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder (its profile holds the project id).
    project: PathBuf,
    /// The best readable copy of the ledger.
    #[arg(long)]
    source: PathBuf,
    /// The HOME to use (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
    /// A claim journal folder (default: the project's own).
    #[arg(long)]
    journal: Option<PathBuf>,
    /// A JSON file of `journal`, `worktrees` and `remotes` facts to add to what is observed.
    #[arg(long)]
    observations: Option<PathBuf>,
    /// `YYYY-MM-DDTHH:MM:SSZ` to judge leases against; the system clock when absent.
    #[arg(long)]
    now: Option<String>,
}

pub const OPERATION: &str = "ledger.recover.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = moai_cli::ledger::command::resolve_home(args.home)?;
    let now_text = args.now.unwrap_or_else(moai_cli::audit::utc_now);
    let now = moai_cli::forge::policy::parse_utc(&now_text)
        .ok_or_else(|| format!("invalid --now `{now_text}`: expected YYYY-MM-DDTHH:MM:SSZ"))?;
    moai_cli::ledger_recover::plan(
        &moai_cli::ledger_recover::Inputs {
            project: &args.project,
            home: &home,
            source: &args.source,
            journal: args.journal.as_deref(),
            observations: args.observations.as_deref(),
            now,
        },
        &moai_cli::ledger_recover::GitProbe,
    )
}
