//! `lane-safe-to-reset` (#2669): whether a lane may be cleared or redispatched without losing
//! work. Reads only; the logic is `moai_cli::lane_reset`. Since #3015 C1 the tool reads the
//! lane's open claims from GitHub itself (GET and search only); the caller gives the screen
//! as a flag.

use moai_cli::forge::policy::SystemClock;
use moai_cli::forge::transport::GhTransport;
use moai_cli::lane_reset::{CwdObservation, ScreenObservation};
use std::path::PathBuf;

/// Check that a lane's session is safe to clear
///
/// Before a lane's session is cleared or its card dispatched elsewhere (#2669): map the lane
/// to its worktrees, count their uncommitted, unpushed and stashed work, and check the lane's
/// open claims (each with a resume note), session folder and screen (background jobs). `safe`
/// only when every check ran clean; any finding is `not_safe`; a check that could not run is
/// `unknown` with its reason. Reads only.
///
/// The lane's open claims are read from GitHub by this command (a search for the lane's
/// `session=` token, then each thread's comments; GET only) unless the observations file
/// gives `claims`. A read that fails leaves the check `unknown` with the reason. Give the
/// screen with `--screen-jobs N` (the number of background jobs you see on the lane's screen
/// now) and the lane's folder with `--session-cwd`; an observations file is only needed to
/// pass a screen read earlier, with its own time.
///
/// The answer also says what state the lane is in (`lane_state`, #3015 C3b), told from its
/// open claims with their resume notes, the open pull requests those claims lead to, unsaved
/// work, a running writer on the screen and the session lease — not from a quiet screen:
/// `idle` (nothing is owed), `waiting` (every claim waits for a check or an audit), `resumable`
/// (stopped, with a resume note to go on from), `stopped` (stopped with nothing saying where to
/// go on), `working` (a background job runs) or `unknown` (a fact was not read; `reasons` names
/// it). The pull requests are read only when this command reads the claims itself.
///
/// Observations file: schema_version: 1; claims.items[].issue, claim_comment, resume_note: numbers (resume_note may be left out); session_cwd.path: text; screen.observed_at: YYYY-MM-DDTHH:MM:SSZ; screen.background_jobs: a number (not a list); no other keys. Every part except schema_version may be left out
///
/// Example:
/// {"schema_version":1,"claims":{"source":"github","items":[{"issue":7,"claim_comment":70,"resume_note":71}]},"session_cwd":{"source":"orca","path":"/path/to/the/lane/worktree"},"screen":{"source":"orca","observed_at":"2026-11-15T02:30:00Z","background_jobs":0}}
#[derive(clap::Args)]
pub struct Args {
    /// Any checkout of the repository (the primary one or a worktree).
    project: PathBuf,
    /// The lane's name, as in its claim comments (`agent-27`).
    #[arg(long)]
    lane: String,
    /// JSON file with `schema_version` 1 and any of `claims`, `session_cwd`, `screen` (shape
    /// and example below). A missing part is read by this command (`claims`), given by a flag
    /// (`session_cwd`, `screen`), or left `unknown`.
    #[arg(long)]
    observations: Option<PathBuf>,
    /// The number of background jobs on the lane's screen, read just now; the observation is
    /// dated at the command's `--now`. Not allowed together with `screen` in the file.
    #[arg(long)]
    screen_jobs: Option<u64>,
    /// The folder the lane's session runs in. Not allowed together with `session_cwd` in the
    /// file.
    #[arg(long)]
    session_cwd: Option<PathBuf>,
    /// `OWNER/REPO` to read the claims from (default: the project's `origin` remote when it is
    /// a GitHub repository).
    #[arg(long)]
    repo: Option<String>,
    /// `YYYY-MM-DDTHH:MM:SSZ` to judge the screen observation's age against; the system
    /// clock when absent.
    #[arg(long)]
    now: Option<String>,
}

pub const OPERATION: &str = "lane.safe_to_reset";

/// `OWNER/REPO` from the flag or the project's `origin`.
fn repo_of(args: &Args) -> Result<(String, String), String> {
    match &args.repo {
        Some(slug) => slug
            .split_once('/')
            .filter(|(o, r)| !o.is_empty() && !r.is_empty() && !r.contains('/'))
            .map(|(o, r)| (o.to_string(), r.to_string()))
            .ok_or_else(|| format!("invalid --repo `{slug}`: expected OWNER/REPO")),
        None => moai_cli::lane_reset::origin_repo(&args.project).ok_or_else(|| {
            "the project's origin remote is not a GitHub repository; pass --repo OWNER/REPO".into()
        }),
    }
}

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let mut obs = match &args.observations {
        None => moai_cli::lane_reset::Observations {
            schema_version: 1,
            ..Default::default()
        },
        Some(path) => {
            let text = std::fs::read_to_string(path)
                .map_err(|e| format!("observations file unavailable: {}: {e}", path.display()))?;
            moai_cli::lane_reset::parse_observations(&text)?
        }
    };
    let now_text = args.now.clone().unwrap_or_else(moai_cli::audit::utc_now);
    let now = moai_cli::forge::policy::parse_utc(&now_text)
        .ok_or_else(|| format!("invalid --now `{now_text}`: expected YYYY-MM-DDTHH:MM:SSZ"))?;
    if let Some(path) = &args.session_cwd {
        if obs.session_cwd.is_some() {
            return Err("session_cwd is given twice: by --session-cwd and in the observations file; keep one".into());
        }
        obs.session_cwd = Some(CwdObservation {
            source: "flag".into(),
            path: path.clone(),
        });
    }
    if let Some(jobs) = args.screen_jobs {
        if obs.screen.is_some() {
            return Err("the screen is given twice: by --screen-jobs and in the observations file; keep one".into());
        }
        obs.screen = Some(ScreenObservation {
            source: "flag".into(),
            observed_at: now_text.clone(),
            background_jobs: jobs,
        });
    }
    // The claims: the file's, else read from GitHub; a failed or cut read is said in `unknowns`.
    let mut claims_note = None;
    let mut ignored = moai_cli::comment_authority::Dropped::default();
    let claims_read_here = obs.claims.is_none();
    if obs.claims.is_none() {
        let read = repo_of(&args).and_then(|(owner, repo)| {
            let transport = GhTransport::from_env()?;
            let extra = moai_cli::lane_reset::local_lane_issues(&args.project, &args.lane);
            moai_cli::lane_reset::read_claims(
                &owner,
                &repo,
                &args.lane,
                &extra,
                &transport,
                &SystemClock,
            )
        });
        match read {
            Ok(r) => {
                obs.claims = Some(r.claims);
                claims_note = r.partial;
                ignored = r.ignored;
            }
            Err(why) => claims_note = Some(why),
        }
    }
    // The lane's state (#3015 C3b) also needs the open pull requests its claims lead to — read
    // only when this command read the claims itself — and its session lease. A fact that is not
    // read stays `None`, and the state then says `unknown`.
    let mut extras = moai_cli::lane_reset::Extras {
        session_live: moai_cli::lane_registry::session_live(&args.project, &args.lane, now),
        ..Default::default()
    };
    if claims_read_here {
        if let (Some(claims), Ok((owner, repo))) = (&obs.claims, repo_of(&args)) {
            let read = GhTransport::from_env().and_then(|transport| {
                moai_cli::lane_state::read_open_prs(
                    &owner,
                    &repo,
                    &claims.items,
                    &transport,
                    &SystemClock,
                )
            });
            if let Ok(r) = read {
                extras.open_prs = Some(r.prs);
                extras.prs_partial = r.partial;
            }
        }
    }
    let mut out = moai_cli::lane_reset::check_full(
        &args.project,
        &args.lane,
        &obs,
        now,
        claims_note.as_deref(),
        &extras,
    )?;
    if !ignored.is_empty() {
        out["ignored_control_lines"] = ignored.to_json();
    }
    Ok(out)
}
