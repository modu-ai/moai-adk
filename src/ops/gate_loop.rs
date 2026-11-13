//! `gate-loop` (#2674 S4): one pass of the leader's merge loop.

use moai_cli::forge::audit_status::APP_TOKEN_VAR;
use moai_cli::forge::policy::SystemClock;
use moai_cli::forge::transport::GhTransport;
use moai_cli::gate_loop::{self, LoopTarget};
use std::path::PathBuf;

/// One pass of the leader's merge loop (#2674 S4): runs `gate-merge` on every open,
/// non-draft pull request into the integration branch, and once per head updates the branch
/// on a code overlap with the tip (G6) or asks the claim holder when the verdict's author holds
/// no allowed claim (G2). The leader or a scheduler calls it instead of a shell loop. Needs a
/// permission reference; the heads must already be fetched (a missing one is skipped).
#[derive(clap::Args)]
pub struct Args {
    /// Run a single pass (the only mode; a scheduler repeats it).
    #[arg(long)]
    once: bool,
    /// The operator's permission reference for these writes.
    #[arg(long)]
    permission_reference: String,
    /// `OWNER/REPO`.
    #[arg(long, default_value = "modu-ai/moai-adk")]
    repo: String,
    /// Integration branch whose pull requests are gated.
    #[arg(long, default_value = "v2026.11.0")]
    base: String,
    /// Local repository holding the heads and the tip.
    #[arg(long, default_value = ".")]
    dir: PathBuf,
    /// Local ref or SHA of the integration tip.
    #[arg(long, default_value = "refs/remotes/origin/v2026.11.0")]
    tip: String,
    /// PR numbers whose REVIEW.md draft the user confirmed, comma-separated.
    #[arg(long, value_delimiter = ',')]
    review_confirmed: Vec<u64>,
}

pub const OPERATION: &str = "gate.loop";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    if !args.once {
        return Err("gate-loop runs one pass; pass --once and repeat it from a scheduler".into());
    }
    let (owner, repo) = args
        .repo
        .split_once('/')
        .ok_or("--repo must be OWNER/REPO")?;
    let target = LoopTarget {
        dir: &args.dir,
        owner,
        repo,
        base_branch: &args.base,
        tip: &args.tip,
        review_confirmed_prs: args.review_confirmed.clone(),
    };
    let mut transport = GhTransport::from_env()?;
    let token = std::env::var(APP_TOKEN_VAR)
        .ok()
        .filter(|t| !t.trim().is_empty());
    let app = token.is_some();
    transport.gh_token = token;
    gate_loop::once(&target, &args.permission_reference, app, &transport, &SystemClock)
}
