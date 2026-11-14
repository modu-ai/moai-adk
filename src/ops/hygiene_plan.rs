//! `hygiene-plan` (#2934 S4a, #2930 R-4): what may move to the holding folder. Reads only; the
//! logic is `moai_cli::hygiene`.

use std::path::PathBuf;

/// Preview what may move to the holding folder
///
/// List the leftovers of lane work that may move to `<HOME>/.moai/holding/<UTC date>/` (#2930):
/// linked worktrees `worktree-gc-plan` would remove (merged, clean, unused, idle), and cargo
/// build outputs directly in the temporary folder (a top `CACHEDIR.TAG`, untouched for 2 days,
/// no process working in it or naming it). Each item is `hold` or `keep` with its reason; a
/// fact that cannot be read keeps it. Writes nothing; the answer carries a `plan_digest`.
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
    /// The time to judge ages at, Unix seconds (default: now).
    #[arg(long)]
    now: Option<u64>,
}

pub const OPERATION: &str = "hygiene.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = moai_cli::home_workdir::home()?;
    let tmp = args.tmp.unwrap_or_else(std::env::temp_dir);
    let now = args.now.unwrap_or_else(|| {
        std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .map(|d| d.as_secs())
            .unwrap_or(0)
    });
    moai_cli::hygiene::plan(&moai_cli::hygiene::Inputs {
        project: &args.project,
        integration_ref: &args.integration_ref,
        tmp_root: &tmp,
        home: &home,
        now,
        cwds: moai_cli::worktree_gc::process_cwds(),
        process_text: moai_cli::hygiene::process_text(),
    })
}
