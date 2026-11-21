//! `git-hook-apply` (SPEC-20261120-commit-message-card-numbers, repair slices): installs the
//! shipped commit-msg hook into the project's git hook folder. The judgement is
//! `moai_cli::git_hooks`: a plan that answers `install` installs, `unchanged` or
//! `needs_decision` refuses with the decision, so a user's own hook is never replaced.

use std::path::PathBuf;

/// Installs the shipped commit-msg hook for a project.
///
/// `moai git-hook-apply [--dir DIR]` copies the shipped hook (the same bytes as
/// `install/commit-msg`, mode 0755) into the folder `git rev-parse --git-path hooks` names,
/// as `commit-msg`, when that folder holds no hook. A hook with the same bytes already
/// there, a different hook file, a non-folder hooks path, or a non-repository is a refusal
/// naming the decision — the user's own hook is never replaced. Reads git once with the
/// scrubbed environment; every other read is the local filesystem.
#[derive(clap::Args)]
pub struct Args {
    /// The repository (default: the current folder).
    #[arg(long, default_value = ".")]
    dir: PathBuf,
}

pub const OPERATION: &str = "git-hook-apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::git_hooks::install(&args.dir)
}
