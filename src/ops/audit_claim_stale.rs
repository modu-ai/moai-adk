//! `audit-claim-stale` (#3015 C7b, N2): audit reservations left on closed or merged pull
//! requests; the logic is `moai_cli::audit_stale`.

use moai_cli::forge::policy::{parse_utc, Clock, SystemClock};
use moai_cli::forge::transport::GhTransport;
use std::path::PathBuf;

/// List audit reservations left on closed pull requests
///
/// Looks at the closed pull requests of the last `--since-hours` (one page of 30, newest
/// update first), reads each one's comments to the end and prints those whose final head still
/// carries a `moai-audit-claim:` line and no `AUDIT … verdict: PASS|FAIL` line for that head
/// (for example after a merge outside the gate). `--lane` keeps the ones that lane reserved or
/// holds. A pull request that cannot be read to the end is named in `unread`, a full page that
/// may hide older ones in `partial`. Reads with GET only and writes nothing: such a reservation
/// has no effect on a closed pull request (#3015 C7b).
#[derive(clap::Args)]
pub struct Args {
    /// Any checkout of the repository; its `origin` remote names OWNER/REPO unless `--repo`.
    #[arg(default_value = ".")]
    project: PathBuf,
    /// `OWNER/REPO` (default: the project's `origin` remote).
    #[arg(long)]
    repo: Option<String>,
    /// Keep only the reservations this lane made or holds.
    #[arg(long)]
    lane: Option<String>,
    /// How far back to look, 1 to 720 hours (default 48).
    #[arg(long, default_value_t = moai_cli::audit_stale::DEFAULT_SINCE_HOURS)]
    since_hours: i64,
    /// `YYYY-MM-DDTHH:MM:SSZ` to judge against; the system clock when absent.
    #[arg(long)]
    now: Option<String>,
}

pub const OPERATION: &str = "audit.claim_stale";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let now = match &args.now {
        Some(text) => parse_utc(text)
            .ok_or_else(|| format!("invalid --now `{text}`: expected YYYY-MM-DDTHH:MM:SSZ"))?,
        None => SystemClock.now(),
    };
    let (owner, repo) = match &args.repo {
        Some(slug) => slug
            .split_once('/')
            .filter(|(o, r)| !o.is_empty() && !r.is_empty() && !r.contains('/'))
            .map(|(o, r)| (o.to_string(), r.to_string()))
            .ok_or_else(|| format!("invalid --repo `{slug}`: expected OWNER/REPO"))?,
        None => moai_cli::lane_reset::origin_repo(&args.project).ok_or(
            "the project's origin remote is not a GitHub repository; pass --repo OWNER/REPO",
        )?,
    };
    moai_cli::audit_stale::list(
        &owner,
        &repo,
        args.lane.as_deref(),
        args.since_hours,
        now,
        &GhTransport::from_env()?,
        &SystemClock,
    )
}
