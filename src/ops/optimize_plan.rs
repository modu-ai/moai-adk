//! `optimize-plan` (#2724): read-only preview of leftover worktrees, build folders and
//! reports. The logic is `moai_cli::optimize`.

use std::path::PathBuf;

/// Preview space that can be freed
///
/// Read-only space preview (#2724): linked worktrees after their merge (squash merges
/// included), cargo build folders under the roots you name, and reports — size, last change,
/// retention, the reasons each is kept, and how to restore it. A worktree is a `remove`
/// candidate only with an independent-audit PASS for its exact HEAD; a live claim, a HOLD or
/// an unobserved part keeps it. Deletes nothing; applying stays `worktree-gc-apply` after the
/// user confirms.
#[derive(clap::Args)]
pub struct Args {
    /// Any checkout of the repository.
    project: PathBuf,
    /// Local ref of the integration branch, e.g. `origin/v2026.11.0`. Nothing is fetched.
    #[arg(long)]
    integration_ref: String,
    /// A change newer than this many minutes keeps a row (at least 60).
    #[arg(long, default_value_t = moai_cli::worktree_gc::DEFAULT_RECENT_MINUTES)]
    recent_minutes: u64,
    /// Time budget in milliseconds; past it the remaining worktrees are `skip` and sizes
    /// are marked partial.
    #[arg(long, default_value_t = moai_cli::worktree_gc::DEFAULT_BUDGET_MS)]
    budget_ms: u64,
    /// A build folder outside the worktrees whose children count only past this many days.
    #[arg(long, default_value_t = moai_cli::optimize::DEFAULT_RETENTION_DAYS)]
    retention_days: u64,
    /// Absolute folder whose child folders are cargo build folders to list (repeatable, e.g.
    /// `/tmp` for `/tmp/moai-rc-target-*`); none by default.
    #[arg(long = "build-root")]
    build_roots: Vec<PathBuf>,
    /// JSON file: `schema_version` 1 and optional `audits` {source, items [{pr, head,
    /// verdict_comment, verdict}]}, `claims` {source, items [{issue, claim_comment}]},
    /// `holds` {source, items [{issue}]}. A missing part keeps the rows that need it.
    #[arg(long)]
    observations: Option<PathBuf>,
    /// Worktrees per project before the preview says the project is over (a whole number of
    /// at least 1). Without it: `[worktrees] cap` in the HOME `hook-config.toml`, else 40. The
    /// cap is only shown; it blocks and removes nothing.
    #[arg(long, allow_hyphen_values = true)]
    worktree_cap: Option<String>,
    /// Test seam: the current time in Unix seconds instead of the clock.
    #[arg(long, hide = true)]
    now_unix: Option<u64>,
}

pub const OPERATION: &str = "optimize.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let obs = match &args.observations {
        None => moai_cli::optimize::Observations {
            schema_version: 1,
            ..Default::default()
        },
        Some(path) => {
            let text = std::fs::read_to_string(path)
                .map_err(|e| format!("observations file unavailable: {}: {e}", path.display()))?;
            serde_json::from_str(&text).map_err(|e| format!("invalid observations: {e}"))?
        }
    };
    let home = moai_cli::hooks::env_home();
    // The cap is read first: an unusable setting refuses before any worktree is read.
    let cap = moai_cli::worktree_cap::resolve_cap(args.worktree_cap.as_deref(), home.as_deref())?;
    let recent = std::time::Duration::from_secs(args.recent_minutes.saturating_mul(60));
    let budget = std::time::Duration::from_millis(args.budget_ms);
    let mut plan = moai_cli::optimize::plan(&moai_cli::optimize::Request {
        project: &args.project,
        integration_ref: &args.integration_ref,
        recent,
        budget,
        now_unix: args.now_unix,
        observations: &obs,
        build_roots: &args.build_roots,
        retention_days: args.retention_days,
        home: home.as_deref(),
    })?;
    // SPEC-20261120-worktree-cap-and-cleanup: the count against the cap and the cleanup
    // candidates sit beside the existing rows; none of it enters `preview_digest`.
    let (cap_report, cleanup) = moai_cli::worktree_cap::preview(
        &args.project,
        &args.integration_ref,
        recent,
        budget,
        args.now_unix,
        home.as_deref(),
        &cap,
    )?;
    if let Some(map) = plan.as_object_mut() {
        map.insert("worktree_cap".into(), cap_report);
        map.insert("cleanup".into(), cleanup);
    }
    Ok(plan)
}
