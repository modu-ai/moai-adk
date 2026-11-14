//! `audit-verdict-line` (#2833 D4): the verdict first line for the head an auditor examined.

use moai_cli::forge::transport::GhTransport;
use moai_cli::gate_collect::{self, Target};
use std::path::PathBuf;

/// Print the verdict first line for the audited head
///
/// Print the first line of an audit verdict for the head the auditor examined (#2833 D4):
/// `AUDIT #N (TITLE) head <40hex> base <40hex> — verdict: PASS|FAIL`. Right before printing it
/// reads the pull request's head on GitHub: when it is no longer `--head`, the line is printed
/// only if the move is a mechanical delta (only the base branch merged in, so the gate carries
/// a PASS on `--head`); otherwise it refuses and names the delta to audit. The base is
/// `--base`, or the merge base of the head and `--tip` in `--dir`. The line reads back through
/// the gate's own parser before it is printed. Reads only; posts nothing.
#[derive(clap::Args)]
pub struct Args {
    /// Pull request number.
    #[arg(long)]
    pr: u64,
    /// The head the audit examined, 40 hex.
    #[arg(long)]
    head: String,
    /// The verdict word.
    #[arg(long, value_parser = ["PASS", "FAIL"])]
    verdict: String,
    /// Short title in the parentheses, one line.
    #[arg(long)]
    title: String,
    /// Base to name instead of the merge base, 40 hex (a delta re-audit names the earlier head).
    #[arg(long)]
    base: Option<String>,
    /// `OWNER/REPO`.
    #[arg(long, default_value = "modu-ai/moai-adk")]
    repo: String,
    /// Local repository holding the heads and the tip.
    #[arg(long, default_value = ".")]
    dir: PathBuf,
    /// Local ref or SHA of the integration tip.
    #[arg(long, default_value = "refs/remotes/origin/v2026.11.0")]
    tip: String,
}

pub const OPERATION: &str = "audit.verdict_line";

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
        review_confirmed_prs: Vec::new(),
    };
    gate_collect::verdict_line_for(
        &target,
        &args.head,
        &args.title,
        &args.verdict,
        args.base.as_deref(),
        &GhTransport::from_env()?,
    )
}
