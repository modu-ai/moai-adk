//! `home-archive-apply` (SPEC-20261123-home-state-retire): archive the earlier HOME state
//! folder. The logic is `moai_cli::home_state_retire`.

use std::path::PathBuf;

/// Archive the earlier HOME state folder under ~/.moai/archive
///
/// Rename the earlier state folder, whole, to `~/.moai/archive/moai-cli-<UTC>/` after writing a
/// manifest and a move journal (one rename, same file system, never overwriting). Runs only when
/// the preview recomputed from the same inputs has the same `plan_digest` (or `content_digest`)
/// and the action `retire`. Copies, opens and deletes nothing: bytes, modes and the vault's
/// owner-only permissions stay as they were. A journal of the same `--now` that a crash left is
/// finished instead. On a real HOME the leader runs it only after the user confirms.
#[derive(clap::Args)]
pub struct Args {
    /// The HOME to change (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
    /// The same `--now` the preview read (for `--plan-digest`; names the archive).
    #[arg(long)]
    now: String,
    /// The preview's `plan_digest`.
    #[arg(long, conflicts_with = "content_digest")]
    plan_digest: Option<String>,
    /// The preview's `content_digest`, which holds no clock.
    #[arg(long)]
    content_digest: Option<String>,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "home_archive.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = match args.home {
        Some(h) => h,
        None => std::env::var_os("HOME")
            .filter(|h| !h.is_empty())
            .map(PathBuf::from)
            .ok_or("HOME is not set; pass --home")?,
    };
    let (confirm, digest) = match (args.plan_digest, args.content_digest) {
        (Some(d), _) => (moai_cli::home_state_retire::Confirm::Plan, d),
        (None, Some(d)) => (moai_cli::home_state_retire::Confirm::Content, d),
        (None, None) => return Err("--plan-digest or --content-digest is required".into()),
    };
    moai_cli::home_state_retire::apply(
        &home,
        &args.now,
        confirm,
        &digest,
        &args.permission_reference,
        &moai_cli::home_state_retire::Facts::current(),
    )
}
