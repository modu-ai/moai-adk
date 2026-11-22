//! `skills-legacy` (#2592): legacy `moai-*` skills beside `moai`; writes nothing.

use std::path::PathBuf;

/// List legacy moai-* skill folders
///
/// List legacy `moai-*` skill folders (every `moai-` entry except `moai` itself) under the
/// Claude skill folder in HOME (the list is `skills_legacy::ROOTS`), with `count` and a
/// `summary` line that says `0건` when there are none. Links are reported, never followed (a folder on the way, an entry or its `SKILL.md`);
/// a `SKILL.md` is read only when it is a regular file of at most 64 KiB, so a FIFO never
/// blocks the check. Reads only; moves, deletes and edits nothing.
#[derive(clap::Args)]
pub struct Args {
    /// User home root.
    home: PathBuf,
}

pub const OPERATION: &str = "skills.legacy";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    Ok(moai_cli::skills_legacy::check(&args.home))
}
