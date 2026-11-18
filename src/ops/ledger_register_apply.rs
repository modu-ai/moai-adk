//! `ledger-register-apply` (#3068 P1): create the project's local ledger, only on an identical
//! fresh preview. The logic is `moai_cli::ledger::command`.

use std::path::PathBuf;

/// Create the previewed project ledger
///
/// Create the project's local SQLite ledger (#3068), the only command that makes the folder
/// and the file. It runs only when the preview recomputed from the same inputs has the same
/// `plan_digest`, and refuses when the HOME is the Go generation's `~/.moai`. A valid ledger that
/// is already there is left untouched (`unchanged`); a file that is not this project's ledger is
/// an error and is never replaced.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder (its profile holds the project id).
    project: PathBuf,
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

pub const OPERATION: &str = "ledger.register.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = moai_cli::ledger::command::resolve_home(args.home)?;
    // Only a registration that would create the ledger waits for the HOME move; an existing
    // ledger keeps its idempotent no-op whatever else sits in `~/.moai` (Codex review P2 on #3199).
    let would_create =
        moai_cli::ledger::command::register_plan(&args.project, &home)?["action"] == "register";
    if would_create {
        if let Some(why) = moai_cli::home_install::registration_blocked(&home) {
            return Err(format!("{why}; nothing was written"));
        }
    }
    moai_cli::ledger::command::register_apply(
        &args.project,
        &home,
        &args.plan_digest,
        &args.permission_reference,
    )
}
