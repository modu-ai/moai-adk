//! `home-install-apply` (#2871 U3c): move this generation's `~/.moai-cli` items to the `~/.moai`
//! top level. The logic is `moai_cli::home_install_apply`.

use std::path::PathBuf;

/// Move the ~/.moai-cli items to ~/.moai
///
/// Rename each entry a fresh preview marks `move` from `~/.moai-cli` to `~/.moai` (#2871), after
/// writing a manifest and a move journal, with `versions` and `path` last; re-point the `path`
/// links at the moved `versions`; write the schema-2 marker last, so the HOME reads as `~/.moai`
/// only when everything is in place. Runs only when the preview recomputed from the same inputs
/// has the same `plan_digest` and the action `install`, and no Rust session is registered in the
/// clone at `--project`. An entry a program uses is not moved and the install is rolled back.
/// Copies and deletes nothing; a secret is judged by `lstat` and never opened. An install that
/// stopped is finished by the same command. On a real HOME the leader runs it only after the
/// user confirms.
#[derive(clap::Args)]
pub struct Args {
    /// The HOME to change. Required: the real HOME is never a default.
    #[arg(long)]
    home: PathBuf,
    /// A folder inside the clone whose Rust sessions are checked (default: the current folder).
    #[arg(long)]
    project: Option<PathBuf>,
    /// The same `--now` the preview read.
    #[arg(long)]
    now: String,
    /// The preview's `plan_digest`.
    #[arg(long)]
    plan_digest: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "home_install.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let project = match args.project {
        Some(p) => p,
        None => std::env::current_dir().map_err(|e| format!("current folder unavailable: {e}"))?,
    };
    let moai_home = std::env::var("MOAI_HOME").ok();
    moai_cli::home_install_apply::apply(
        &args.home,
        moai_home.as_deref(),
        &project,
        &args.now,
        &args.plan_digest,
        &args.permission_reference,
    )
}
