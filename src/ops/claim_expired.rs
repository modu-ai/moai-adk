//! `claim-expired` (#3015 C7a, N19): the open claims whose lease has ended; the logic is
//! `moai_cli::claim_expired`.

use moai_cli::forge::policy::{parse_utc, SystemClock};
use moai_cli::forge::transport::GhTransport;
use std::path::PathBuf;

/// List the open claims whose lease has ended
///
/// Searches the open Issues and PRs that carry a `moai-claim:` line (one page of 100), reads
/// each thread's comments to the end and prints the holders (by the rule of `claim-verify`)
/// whose lease is over (`lease_until <= now`), longest-expired first, with `expired_secs`. A
/// thread that cannot be read to the end is named in `unread`, never counted as having no
/// expired claim; a cut search is named in `partial`. Reads with GET only. It takes nothing
/// over: a takeover still needs `claim-timeout-check` and a handoff line (#3015 C7a).
#[derive(clap::Args)]
pub struct Args {
    /// Any checkout of the repository; its `origin` remote names OWNER/REPO unless `--repo`.
    #[arg(default_value = ".")]
    project: PathBuf,
    /// `OWNER/REPO` (default: the project's `origin` remote).
    #[arg(long)]
    repo: Option<String>,
    /// `YYYY-MM-DDTHH:MM:SSZ` to judge leases against; the system clock when absent.
    #[arg(long)]
    now: Option<String>,
}

pub const OPERATION: &str = "claim.expired";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let now_text = args.now.unwrap_or_else(moai_cli::audit::utc_now);
    let now = parse_utc(&now_text)
        .ok_or_else(|| format!("invalid --now `{now_text}`: expected YYYY-MM-DDTHH:MM:SSZ"))?;
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
    moai_cli::claim_expired::list(&owner, &repo, now, &GhTransport::from_env()?, &SystemClock)
}

