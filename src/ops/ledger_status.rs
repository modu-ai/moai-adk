//! `ledger-status` (#3068 P1): read-only report of the project's local ledger. The logic is
//! `moai_cli::ledger::command`.

use std::path::PathBuf;

/// Show the project's ledger state
///
/// Report whether the project's local ledger exists, its schema version, that it is not in WAL
/// mode, and how many cards it holds by state (#3068). Opens the file read-only; an absent ledger
/// is a normal answer (`exists: false`, exit 0) and creates no folder or file. A file that is not
/// this project's usable ledger is reported with its `problem` code and is never repaired.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder (its profile holds the project id).
    project: PathBuf,
    /// The HOME to read (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
}

pub const OPERATION: &str = "ledger.status";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = moai_cli::ledger::command::resolve_home(args.home)?;
    moai_cli::ledger::command::status(&args.project, &home)
}
