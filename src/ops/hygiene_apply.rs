//! `hygiene-apply` (#2934 S4b, #2930 R-4): moves what a matching `hygiene-plan` holds into
//! the holding folder. The logic is `moai_cli::hygiene::apply`.

use std::path::PathBuf;

/// Move the previewed leftovers to the holding folder
///
/// Act on a `hygiene-plan` (#2930): recomputes the plan with the same inputs and refuses
/// before any write unless its digest equals `--plan-digest`, or when the holding folder or
/// its `manifest.tsv` is a link. Each `hold` item is re-checked for use (working folder,
/// command line, open files) right before it moves; its manifest line is written first, then
/// one same-file-system rename moves it to `<HOME>/.moai/holding/<UTC date>/<n>-<name>` —
/// never a copy. An item that is in use or cannot be renamed is skipped and reported. Moved
/// worktrees are pruned from git's list (`git worktree prune`).
#[derive(clap::Args)]
pub struct Args {
    /// Any checkout of the repository whose worktrees are looked at.
    project: PathBuf,
    /// The integration branch a worktree must be merged into, e.g. `origin/v2026.11.0`.
    #[arg(long)]
    integration_ref: String,
    /// The temporary folder whose direct children are looked at (default: the system one).
    #[arg(long)]
    tmp: Option<PathBuf>,
    /// The home folder whose `.moai/holding` receives the items (default: HOME).
    #[arg(long)]
    home: Option<PathBuf>,
    /// The time to judge ages and name the holding folder at, Unix seconds (default: now).
    #[arg(long)]
    now: Option<u64>,
    #[arg(long)]
    plan_digest: String,
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "hygiene.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = match args.home {
        Some(h) => h,
        None => moai_cli::home_workdir::home()?,
    };
    let tmp = args.tmp.unwrap_or_else(std::env::temp_dir);
    let now = args.now.unwrap_or_else(|| {
        std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .map(|d| d.as_secs())
            .unwrap_or(0)
    });
    moai_cli::hygiene::apply(
        &moai_cli::hygiene::Inputs {
            project: &args.project,
            integration_ref: &args.integration_ref,
            tmp_root: &tmp,
            home: &home,
            now,
            cwds: moai_cli::worktree_gc::process_cwds(),
            process_text: moai_cli::hygiene::process_text(),
        },
        &args.plan_digest,
        &args.permission_reference,
    )
}
