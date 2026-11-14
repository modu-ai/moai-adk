//! `audit-verdict-line` (#2833 D4): the verdict first line for a PR's current head.

use moai_cli::forge::transport::GhTransport;
use moai_cli::gate_collect::{self, Target};
use std::path::PathBuf;

/// Print the verdict first line for a PR's current head
///
/// Print the first line of an audit verdict for a pull request's head as GitHub has it now
/// (#2833 D4): `AUDIT #N (TITLE) head <40hex> base <40hex> — verdict: PASS|FAIL`. The head is
/// read right before printing, so a verdict never names a head that moved during the audit.
/// The base is `--base`, or the merge base of the head and `--tip` in `--dir`. The line reads
/// back through the gate's own parser before it is printed. Reads only; posts nothing.
#[derive(clap::Args)]
pub struct Args {
    /// Pull request number.
    #[arg(long)]
    pr: u64,
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
    /// Local repository holding the head and the tip (when `--base` is not given).
    #[arg(long, default_value = ".")]
    dir: PathBuf,
    /// Local ref or SHA of the integration tip (when `--base` is not given).
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
        &args.title,
        &args.verdict,
        args.base.as_deref(),
        &GhTransport::from_env()?,
    )
}
