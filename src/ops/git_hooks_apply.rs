//! `git-hooks-apply` (SPEC-20261121-go-path-removal): the confirmed git-hook retire and
//! install; the logic is `moai_cli::git_hooks`.

use std::path::PathBuf;

/// Apply the previewed git-hook change
///
/// Apply the previewed git-hook change (SPEC-20261121-go-path-removal): retires the
/// Go-generation hooks with their hash files into `.git/hooks/moai-retired-<stamp>/` and
/// installs the shipped Rust texts in the free names. Only the preview's token applies it;
/// hooks without MoAI marks are never touched.
#[derive(clap::Args)]
pub struct Args {
    /// The repository (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
    /// The confirming token the preview printed.
    #[arg(long = "confirm-git-hooks")]
    confirm_git_hooks: String,
}

pub const OPERATION: &str = "git_hooks.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::git_hooks::apply(
        &args.project,
        &args.confirm_git_hooks,
        &moai_cli::audit::utc_now(),
    )
}
