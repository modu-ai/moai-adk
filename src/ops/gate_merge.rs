//! `gate-merge` (#2674 S2): the leader's merge after the gate passes.

use moai_cli::forge::audit_status::APP_TOKEN_VAR;
use moai_cli::forge::policy::SystemClock;
use moai_cli::forge::transport::GhTransport;
use moai_cli::gate_collect::Target;
use std::path::PathBuf;

/// Merge one pull request only when the leader merge gate passes (#2674 S2): checks every rule
/// G1–G7 like `gate-check --pr`, then posts the `moai/audit` success status on the head, merges
/// with the head pinned (at most three attempts) and posts the merge record. A stopped gate
/// writes nothing. Needs a permission reference.
#[derive(clap::Args)]
pub struct Args {
    /// Pull request number.
    #[arg(long)]
    pr: u64,
    /// The operator's permission reference for these writes.
    #[arg(long)]
    permission_reference: String,
    /// `OWNER/REPO`.
    #[arg(long, default_value = "modu-ai/moai-adk")]
    repo: String,
    /// Local repository holding the head and the tip.
    #[arg(long, default_value = ".")]
    dir: PathBuf,
    /// Local ref or SHA of the integration tip.
    #[arg(long, default_value = "refs/remotes/origin/v2026.11.0")]
    tip: String,
    /// PR numbers whose REVIEW.md draft the user confirmed, comma-separated.
    #[arg(long, value_delimiter = ',')]
    review_confirmed: Vec<u64>,
}

pub const OPERATION: &str = "gate.merge";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let (owner, repo) = args
        .repo
        .split_once('/')
        .ok_or("--repo must be OWNER/REPO")?;
    let target = Target {
        dir: &args.dir,
        owner,
        repo,
        pr: args.pr,
        tip: &args.tip,
        review_confirmed_prs: args.review_confirmed.clone(),
    };
    let mut transport = GhTransport::from_env()?;
    // The status is posted with the audit App token when it is set, like forge-audit-status.
    let token = std::env::var(APP_TOKEN_VAR)
        .ok()
        .filter(|t| !t.trim().is_empty());
    let app = token.is_some();
    transport.gh_token = token;
    moai_cli::gate_merge::merge(&target, &args.permission_reference, app, &transport, &SystemClock)
}
