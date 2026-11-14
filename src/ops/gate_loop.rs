//! `gate-loop` (#2674 S4): one pass of the leader's merge loop.

use moai_cli::forge::audit_status::APP_TOKEN_VAR;
use moai_cli::forge::policy::SystemClock;
use moai_cli::forge::transport::GhTransport;
use moai_cli::gate_loop::{self, LoopTarget};
use std::path::PathBuf;

/// Run one pass of the leader merge loop
///
/// One pass of the leader's merge loop (#2674 S4): runs `gate-merge` on every open pull
/// request into the integration branch, oldest number first, drafts included (a passing draft
/// is marked ready just before its merge, #2860), and once per head updates the branch on a
/// code overlap with the tip (G6) or asks the claim holder when the verdict's author holds no
/// allowed claim (G2). The leader or a scheduler calls it instead of a shell loop. Needs a
/// permission reference, except with `--dry-run`, which writes nothing and lists the decision
/// for each pull request. The heads must be local (a missing one is skipped); `--fetch` first
/// fetches the base branch and every open head from `origin` into named remote refs.
#[derive(clap::Args)]
pub struct Args {
    /// Run a single pass (the only mode; a scheduler repeats it).
    #[arg(long)]
    once: bool,
    /// Read only: report what the pass would do for each pull request, write nothing (#2860).
    #[arg(long)]
    dry_run: bool,
    /// The operator's permission reference for these writes (not needed with --dry-run).
    #[arg(long, default_value = "")]
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
    /// Fetch the base branch and every open pull request head from `origin` first (#2906).
    #[arg(long)]
    fetch: bool,
}

pub const OPERATION: &str = "gate.loop";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    // Static progress lines on a stderr terminal (#2806); silent on a pipe.
    moai_cli::progress::install_stderr();
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
        fetch: args.fetch,
    };
    // Only the status is posted with the audit App token (Commit statuses permission only);
    // reads, merges, branch updates and notes use gh's own authentication (#2735 audit F3).
    let transport = GhTransport::from_env()?;
    let mut status_transport = GhTransport::from_env()?;
    let token = std::env::var(APP_TOKEN_VAR)
        .ok()
        .filter(|t| !t.trim().is_empty());
    let app = token.is_some();
    status_transport.gh_token = token;
    gate_loop::once(
        &target,
        &args.permission_reference,
        app,
        args.dry_run,
        &transport,
        &status_transport,
        &SystemClock,
    )
}
