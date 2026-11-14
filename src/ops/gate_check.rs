//! `gate-check` (#2674 S1): the leader merge gate rules over one pull request snapshot.

use moai_cli::forge::transport::GhTransport;
use moai_cli::gate::{self, Snapshot};
use moai_cli::gate_collect::{self, Target};
use std::io::Read;
use std::path::PathBuf;

/// Check every rule of the leader merge gate
///
/// The leader merge gate (#2674), read-only: decides whether every rule G1–G12 holds for one pull
/// request (G8: the tip's last completed push CI is not red, or this PR is its repair, #2777;
/// G9: the auditor is not an author; G10: every PR commit but a merge has a `Moai-Claim`
/// trailer; G11: the PR body has a `Moai-Claim: ` line; G12: an open PR below this one has a
/// PASS holding every commit this PR would bring in, #2859) and prints `pass`, or the rule that stopped and why. With
/// `--pr` it reads the pull
/// request from GitHub (GET only) and the local repository (the head and tip must already be
/// fetched); without it, it decides from a snapshot on stdin. Posts nothing, merges nothing.
#[derive(clap::Args)]
pub struct Args {
    /// Pull request number to read live.
    #[arg(long)]
    pr: Option<u64>,
    /// `OWNER/REPO` (with `--pr`).
    #[arg(long, default_value = "modu-ai/moai-adk")]
    repo: String,
    /// Local repository holding the head and the tip (with `--pr`).
    #[arg(long, default_value = ".")]
    dir: PathBuf,
    /// Local ref or SHA of the integration tip (with `--pr`).
    #[arg(long, default_value = "refs/remotes/origin/v2026.11.0")]
    tip: String,
    /// PR numbers whose REVIEW.md draft the user confirmed, comma-separated (with `--pr`).
    #[arg(long, value_delimiter = ',')]
    review_confirmed: Vec<u64>,
}

pub const OPERATION: &str = "gate.check";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    // Static progress lines on a stderr terminal (#2806); silent on a pipe.
    moai_cli::progress::install_stderr();
    if let Some(pr) = args.pr {
        let (owner, repo) = args
            .repo
            .split_once('/')
            .ok_or("--repo must be OWNER/REPO")?;
        let target = Target {
            dir: &args.dir,
            owner,
            repo,
            pr,
            tip: &args.tip,
            review_confirmed_prs: args.review_confirmed.clone(),
        };
        let snapshot = gate_collect::collect(&target, &GhTransport::from_env()?)?;
        return Ok(decide(&snapshot));
    }
    let mut text = String::new();
    std::io::stdin()
        .take(4 * 1024 * 1024 + 1)
        .read_to_string(&mut text)
        .map_err(|_| "stdin is not readable UTF-8".to_string())?;
    if text.len() > 4 * 1024 * 1024 {
        return Err("snapshot larger than 4 MiB; refused".into());
    }
    let snapshot: Snapshot =
        serde_json::from_str(&text).map_err(|e| format!("invalid gate snapshot: {e}"))?;
    Ok(decide(&snapshot))
}

fn decide(snapshot: &Snapshot) -> serde_json::Value {
    match gate::check(snapshot) {
        Ok((base, overlap)) => {
            let mut out = serde_json::json!({
                "pr": snapshot.pr,
                "head": snapshot.head,
                "pass": true,
                "verdict_base": base,
                "overlap": overlap,
            });
            // #2833 D1: a PASS carried by a mechanical delta shows its evidence.
            if let Some(d) = gate::accepted_auto_delta(snapshot) {
                out["auto_delta"] = serde_json::to_value(d).unwrap_or_default();
            }
            out
        }
        Err(stop) => serde_json::json!({
            "pr": snapshot.pr,
            "head": snapshot.head,
            "pass": false,
            "rule": stop.rule,
            "reason": stop.reason,
        }),
    }
}
