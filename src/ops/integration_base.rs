//! `integration-base` (SPEC-20261121-integration-base-sha, slice S1): the local commit an m-flow
//! lane starts from. The logic is `moai_cli::spec_take::{integration_base, worktree_base}`.

use moai_cli::contract::PROJECT_PATH_UNAVAILABLE;
use moai_cli::spec_take::{self, IntegrationBase};
use serde_json::{json, Value};
use std::path::PathBuf;

/// Print the local commit that lanes start from
///
/// Reads the project profile: when its work flow is `m-flow` and it names an integration
/// branch, prints that branch and the full commit id of `refs/heads/<branch>` — the local
/// branch only, never a tag and never a remote-tracking ref, so a diverged `origin` cannot stand
/// in for it. When that local branch does not exist it stops with the next step and prints no
/// commit id. Any other flow (or no readable profile) prints `policy: "none"` and the reason.
/// With `--worktree PATH` it also reads that worktree: its `HEAD`, whether it holds the local
/// integration commit (`contains_base`), the commit recorded for its branch (`recorded_sha`)
/// and whether that equals the local commit. Reads only; writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder.
    #[arg(default_value = ".")]
    project: PathBuf,
    /// A worktree to read against the local integration commit.
    #[arg(long)]
    worktree: Option<PathBuf>,
}

pub const OPERATION: &str = "integration.base";

pub fn run(args: Args) -> Result<Value, String> {
    let project =
        std::fs::canonicalize(&args.project).map_err(|_| PROJECT_PATH_UNAVAILABLE.to_string())?;
    Ok(match spec_take::integration_base(&project)? {
        IntegrationBase::NoPolicy { reason } => json!({
            "action": "integration-base", "policy": "none", "reason": reason,
            "branch": null, "sha": null, "source": null, "worktree": null, "writes": false,
        }),
        IntegrationBase::Local { branch, sha } => {
            let worktree = match &args.worktree {
                Some(path) => {
                    let w = spec_take::worktree_base(path, &sha)?;
                    json!({
                        "path": path.to_string_lossy(), "head": w.head, "branch": w.branch,
                        "contains_base": w.contains_base, "recorded_sha": w.recorded_sha,
                        "matches_recorded": w.matches_recorded,
                    })
                }
                None => Value::Null,
            };
            json!({
                "action": "integration-base", "policy": "m-flow", "branch": branch,
                "sha": sha, "source": "local", "worktree": worktree, "writes": false,
            })
        }
    })
}
