//! `home-retire-apply` (#2871 U2): move the previewed Go items of `~/.moai` to a dated
//! backup. The logic is `moai_cli::home_retire`.

use std::path::PathBuf;

/// Retire the Go items in ~/.moai
///
/// Move the entries a fresh preview marks `retire` to `~/.moai-go-retired/<UTC>/` (one rename
/// each, never overwriting, same file system), after writing a manifest and a move journal
/// (#2871). Runs only when the preview recomputed from the same inputs has the same
/// `plan_digest` and the action `retire`. Copies and deletes nothing; a running program's
/// entry is skipped. On a real HOME the leader runs it only after the user confirms.
#[derive(clap::Args)]
pub struct Args {
    /// The HOME to change (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
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

pub const OPERATION: &str = "home_retire.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = match args.home {
        Some(h) => h,
        None => std::env::var_os("HOME")
            .filter(|h| !h.is_empty())
            .map(PathBuf::from)
            .ok_or("HOME is not set; pass --home")?,
    };
    let moai_home = std::env::var("MOAI_HOME").ok();
    moai_cli::home_retire::apply(
        &home,
        moai_home.as_deref(),
        &args.now,
        &args.plan_digest,
        &args.permission_reference,
    )
}
