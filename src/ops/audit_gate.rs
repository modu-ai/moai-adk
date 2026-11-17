//! `audit-gate` (SPEC-20261117-audit-gate): the CI audit-gate job's decision for one pull
//! request head.

use moai_cli::audit_gate::{self, Input};
use moai_cli::forge::transport::GhTransport;
use std::io::Read;

/// Decide whether a PR head has the audit its tier needs
///
/// The CI `audit-gate` job's decision (SPEC-20261117-audit-gate), read-only: from the PR's
/// verdict comments it prints `pass`, `fail` (a posted FAIL with no later PASS by the same
/// auditor, at every tier) or `missing` (tier 1 and 2: no PASS on exactly this head by an
/// auditor lane that is not a `Moai-Session:` author). Only members' comments count. With
/// `--pr` it reads the pull request from GitHub (GET only); without it, it decides from an
/// input on stdin. `--require-pass` makes anything but `pass` an error, for the job's exit
/// code. Whether the PASS cites the CI run is not checked yet. Posts nothing.
#[derive(clap::Args)]
pub struct Args {
    /// Pull request number to read live.
    #[arg(long)]
    pr: Option<u64>,
    /// The head the check is for, 40 hex (with `--pr`; default the PR's current head).
    #[arg(long)]
    head: Option<String>,
    /// `OWNER/REPO` (with `--pr`).
    #[arg(long, default_value = "modu-ai/moai-adk")]
    repo: String,
    /// Fail unless the decision is `pass`.
    #[arg(long)]
    require_pass: bool,
}

pub const OPERATION: &str = "audit.gate";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let input = match args.pr {
        Some(pr) => {
            let (owner, repo) = args
                .repo
                .split_once('/')
                .ok_or("--repo must be OWNER/REPO")?;
            if let Some(h) = &args.head {
                if !moai_cli::gate::is_full_sha(h) {
                    return Err("--head must be 40 lowercase hex".into());
                }
            }
            audit_gate::collect(
                owner,
                repo,
                pr,
                args.head.as_deref(),
                &GhTransport::from_env()?,
            )?
        }
        None => {
            let mut text = String::new();
            std::io::stdin()
                .take(4 * 1024 * 1024 + 1)
                .read_to_string(&mut text)
                .map_err(|_| "stdin is not readable UTF-8".to_string())?;
            if text.len() > 4 * 1024 * 1024 {
                return Err("audit-gate input larger than 4 MiB; refused".into());
            }
            serde_json::from_str::<Input>(&text)
                .map_err(|e| format!("invalid audit-gate input: {e}"))?
        }
    };
    let decision = audit_gate::decide(&input);
    if args.require_pass && decision["state"] != "pass" {
        return Err(format!(
            "audit gate {}: {} ({})",
            decision["state"].as_str().unwrap_or("unknown"),
            decision["reason"].as_str().unwrap_or(""),
            decision["rule"].as_str().unwrap_or("")
        ));
    }
    Ok(decision)
}
