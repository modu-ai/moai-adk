//! `wait` (#2676): one run that polls GitHub with conditional requests until a condition
//! holds, then ends — for a tool's background run or Monitor instead of a hand-written loop.

use moai_cli::forge::policy::SystemClock;
use moai_cli::forge::transport::GhTransport;
use moai_cli::forge::wait::{self, Condition, WaitRequest};

/// Wait until a condition holds on GitHub, then end. KIND is `ci` (every workflow run of
/// `--head` completed), `comment` (a comment on `--issue` newer than `--after`, optionally with
/// a line starting `--prefix`) or `pr` (pull request `--pr` no longer at `--head`, or closed).
/// Polls are conditional requests (an unchanged answer is a 304), one at a time, at least
/// `--interval` and GitHub's `x-poll-interval` apart. A timeout answers `matched: false`.
#[derive(clap::Args)]
pub struct Args {
    /// `ci`, `comment` or `pr`.
    kind: String,
    /// `OWNER/REPO`.
    #[arg(long)]
    repo: String,
    /// 40-hex commit (`ci`, `pr`).
    #[arg(long)]
    head: Option<String>,
    /// Issue or pull request number whose comments to read (`comment`).
    #[arg(long)]
    issue: Option<u64>,
    /// Only comments with a larger id count (`comment`).
    #[arg(long, default_value_t = 0)]
    after: u64,
    /// Only a comment with a line starting with this text counts (`comment`; 1..=200
    /// characters; only the first 4000 characters of a comment are read).
    #[arg(long)]
    prefix: Option<String>,
    /// Pull request number (`pr`).
    #[arg(long)]
    pr: Option<u64>,
    /// Seconds before answering `matched: false` (at most 3600).
    #[arg(long, default_value_t = 1800)]
    timeout: u64,
    /// Seconds between polls (at least 10).
    #[arg(long, default_value_t = 30)]
    interval: u64,
}

pub const OPERATION: &str = "wait";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    // Static progress lines on a stderr terminal (#2806); silent on a pipe.
    moai_cli::progress::install_stderr();
    let (owner, repo) = args
        .repo
        .split_once('/')
        .ok_or("--repo must be OWNER/REPO")?;
    let need = |v: Option<String>, flag: &str| v.ok_or(format!("wait {} needs {flag}", args.kind));
    let condition = match args.kind.as_str() {
        "ci" => Condition::Ci {
            head: need(args.head.clone(), "--head")?,
        },
        "comment" => Condition::Comment {
            issue: args.issue.ok_or("wait comment needs --issue")?,
            after: args.after,
            prefix: args.prefix.clone(),
        },
        "pr" => Condition::Pr {
            pr: args.pr.ok_or("wait pr needs --pr")?,
            head: need(args.head.clone(), "--head")?,
        },
        _ => return Err("unknown wait kind; must be ci, comment or pr".into()),
    };
    let req = WaitRequest {
        owner: owner.to_string(),
        repo: repo.to_string(),
        condition,
        timeout_s: args.timeout,
        interval_s: args.interval,
    };
    wait::wait(&req, &GhTransport::from_env()?, &SystemClock).map_err(|e| e.message)
}
