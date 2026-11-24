//! `git-hooks-plan` (SPEC-20261121-go-path-removal): what replacing a repository's MoAI-managed
//! git hooks would do. Reads only; the logic is `moai_cli::git_hooks`.

use std::path::PathBuf;

/// Preview the git-hook retire and install
///
/// Preview the git hooks `moai` manages in a repository (SPEC-20261121-go-path-removal): the
/// Go-generation hooks (recognised by their hash files and MoAI-ADK headers) retire, the Rust
/// hook texts install in the free names, and every hook without MoAI marks stays byte for byte.
/// Writes nothing; the answer carries the confirming token of the apply.
#[derive(clap::Args)]
pub struct Args {
    /// The repository (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "git_hooks.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::git_hooks::plan(&args.project)
}
