//! `audit-done` (SPEC-20261124-audit-prep): finishes the audit slot of a card. The logic is
//! `moai_cli::audit_prep::done`.

use std::path::PathBuf;

/// Finish the audit slot of a card and release it
///
/// Finds the audit slot the card holds. A clean slot is released and keeps its `target/` for the
/// next audit; a slot with a change left in it (a mutant, a stray file) is refused with the files
/// named and nothing changed. There is no discard and no force; revert the change in the slot and
/// run it again. A card that holds no slot is refused.
#[derive(clap::Args)]
pub struct Args {
    /// The card whose audit is over, like `C-364`.
    card: String,
    /// The project folder (any checkout of the repository).
    #[arg(long, default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "audit.done";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::audit_prep::done(&args.project, &args.card)
}
