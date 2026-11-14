//! `leader-round` (#2902 O3b): one action per lane for the leader's round, and what the round
//! shares. The logic is `moai_cli::leader_round`; this file reads the inputs.

use moai_cli::forge::transport::GhTransport;
use moai_cli::lane_reset::{Observations, ScreenObservation};
use moai_cli::leader_round::{self, Act, LaneObservation, ResetVerdict};
use serde_json::{json, Value};
use std::path::{Path, PathBuf};

/// Build the leader's per-lane action list
///
/// Build the leader's action list for one round (#2902 O3b): for each lane in the observations
/// file (its screen as Orca read it, its turn, input box and context use) one action —
/// `clear_redispatch` at a card boundary that `lane-safe-to-reset` finds safe, `answer` for a
/// `LANE-BLOCKED` line, `restart` with the loop line, or `leave` with the reason — and, from
/// GitHub, the open Issues with a close table, the claims whose lease ended, the merge-gate
/// command and the end condition. Reads only: git with `--no-optional-locks`, the session
/// registry without writing, GitHub with GraphQL queries (at most 30 requests).
#[derive(clap::Args)]
pub struct Args {
    /// Any checkout of the repository (the primary one or a worktree).
    project: PathBuf,
    /// JSON file: `schema_version` 1 and `lanes`, each with `lane`, `host`, `read_at`,
    /// `source`, `terminal` (Orca `terminal read --screen --json`), `pane`, `turn`,
    /// `input_box`, `context_used_pct`, `wake_armed`, `background_jobs` and `reset` {claims,
    /// session_cwd}.
    #[arg(long)]
    observations: PathBuf,
    /// Read the GitHub state from a saved `lane-next` snapshot (JSON) instead of reading it.
    #[arg(long)]
    snapshot: Option<PathBuf>,
    /// `OWNER/REPO`: read GitHub now (without `--snapshot`) and name it in the gate command.
    #[arg(long)]
    repo: Option<String>,
    /// `YYYY-MM-DDTHH:MM:SSZ` to judge observation ages and leases against; the system clock
    /// when absent.
    #[arg(long)]
    now: Option<String>,
}

pub const OPERATION: &str = "leader.round";

/// The `lane-safe-to-reset` verdict for one observed lane; a check that cannot run is
/// `unknown` with its reason.
fn reset_verdict(project: &Path, lane: &LaneObservation, now: i64) -> ResetVerdict {
    let obs = Observations {
        schema_version: 1,
        claims: lane.reset.claims.clone(),
        session_cwd: lane.reset.session_cwd.clone(),
        screen: Some(ScreenObservation {
            source: lane.source.clone(),
            observed_at: lane.read_at.clone(),
            background_jobs: lane.background_jobs,
        }),
    };
    let texts = |v: &Value, key: &str| -> Vec<String> {
        v[key]
            .as_array()
            .into_iter()
            .flatten()
            .filter_map(|s| s.as_str().map(str::to_string))
            .collect()
    };
    match moai_cli::lane_reset::check(project, &lane.lane, &obs, now) {
        Ok(v) => match v["verdict"].as_str() {
            Some("safe") => ResetVerdict::Safe,
            Some("not_safe") => ResetVerdict::NotSafe(texts(&v, "findings")),
            _ => ResetVerdict::Unknown(texts(&v, "unknowns")),
        },
        Err(e) => ResetVerdict::Unknown(vec![e]),
    }
}

/// `word` as one POSIX shell word: bare when it holds only safe characters, else single-quoted
/// with each `'` written `'\''` (#2948 audit N3; the `lane_env` and `gate_schedule` quoting).
fn shell_word(word: &str) -> String {
    if !word.is_empty()
        && word
            .chars()
            .all(|c| c.is_ascii_alphanumeric() || "/._-:,+=@".contains(c))
    {
        word.to_string()
    } else {
        format!("'{}'", word.replace('\'', "'\\''"))
    }
}

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let now_text = args.now.unwrap_or_else(moai_cli::audit::utc_now);
    let now = moai_cli::forge::policy::parse_utc(&now_text)
        .ok_or_else(|| format!("invalid --now `{now_text}`: expected YYYY-MM-DDTHH:MM:SSZ"))?;
    let text = std::fs::read_to_string(&args.observations).map_err(|e| {
        format!(
            "observations file unavailable: {}: {e}",
            args.observations.display()
        )
    })?;
    let file = leader_round::parse(&text)?;
    let snap = match (&args.snapshot, &args.repo) {
        (Some(path), _) => {
            let bytes = std::fs::read(path).map_err(|e| format!("snapshot unreadable: {e}"))?;
            Some(
                serde_json::from_slice::<moai_cli::lane_next::Snapshot>(&bytes)
                    .map_err(|e| format!("snapshot invalid: {e}"))?,
            )
        }
        (None, Some(repo)) => {
            let (owner, name) = repo.split_once('/').ok_or("--repo must be OWNER/REPO")?;
            Some(moai_cli::lane_next::collect_fleet(
                owner,
                name,
                &args.project,
                &GhTransport::from_env()?,
            ))
        }
        (None, None) => None,
    };
    let round = leader_round::round(snap.as_ref(), args.repo.as_deref(), &args.project, now);
    let end_met = round["end_condition"]["met"].as_bool();
    let mut lanes = Vec::new();
    for lane in &file.lanes {
        let verdict = reset_verdict(&args.project, lane, now);
        let mut a = leader_round::decide(lane, now, &verdict, end_met);
        if a.action == Act::ClearRedispatch {
            // The holder the lane registered, read without writing; the leader releases it
            // before `/clear` so the next session may register the name at once.
            match moai_cli::lane_registry::show(&args.project, &lane.lane, now) {
                Ok(v) if v["status"] == "held" => {
                    a.release = Some(format!(
                        "moai session-register --release --by leader --name {} --instance {} {}",
                        lane.lane,
                        v["instance"].as_str().unwrap_or(""),
                        shell_word(&args.project.to_string_lossy())
                    ));
                }
                Ok(_) => a
                    .notes
                    .push("session registry: the name is free; nothing to release".into()),
                Err(e) => a.notes.push(format!("session registry: {e}")),
            }
        }
        lanes.push(a);
    }
    Ok(json!({
        "now": moai_cli::forge::policy::format_utc(now),
        "lanes": lanes,
        "round": round,
        "partial": snap.as_ref().map(|s| s.partial.clone()).unwrap_or_default(),
        "api_calls": snap.as_ref().map_or(0, |s| s.api_calls),
        "writes": false,
    }))
}
