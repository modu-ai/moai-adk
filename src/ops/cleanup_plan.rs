//! `cleanup-plan` (#2872 stage 1): preview archiving what an update or the Go generation left in
//! HOME. Reads only; the logic is `moai_cli::update_cleanup`.

use std::path::PathBuf;

/// Preview archiving leftover MoAI files in HOME
///
/// Preview the HOME cleanup (#2872): files in the host `moai` skill folders the staged package
/// no longer carries, Go-era `moai-*` skills and agents, and Go `~/.moai` leftovers. Each item
/// is `archive` only when proven unchanged (the host receipt or the Go user-asset manifest),
/// otherwise `keep` with a reason; secrets are never read, links never followed. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// User home root whose host folders and `~/.moai` are inspected.
    home: PathBuf,
    /// Package directory carrying SKILL.md (the staged package, as for update-plan).
    stage: PathBuf,
}

pub const OPERATION: &str = "cleanup.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let staged = moai_cli::update::stage_package(&args.stage)?;
    moai_cli::update_cleanup::plan(&args.home, &staged)
}
