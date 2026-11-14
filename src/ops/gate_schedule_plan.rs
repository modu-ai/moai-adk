//! `gate-schedule-plan` (#2906 S2): preview of a launchd or cron entry that runs
//! `gate-loop --once`; writes nothing. The logic is `moai_cli::gate_schedule`.

use std::path::PathBuf;

/// What the scheduler entry is made from (shared with `gate-schedule-apply`).
#[derive(clap::Args)]
pub struct Entry {
    /// User home root (the LaunchAgent folder and the state root are under it).
    pub home: PathBuf,
    /// `macos` (LaunchAgent plist) or `linux` (crontab file); `windows` is refused.
    #[arg(long)]
    pub os: String,
    /// `OWNER/REPO` the pass gates.
    #[arg(long, default_value = "modu-ai/moai-adk")]
    pub repo: String,
    /// Absolute path of the `moai` binary the entry runs (default: this binary).
    #[arg(long)]
    pub moai_bin: Option<PathBuf>,
    /// Absolute repository folder the pass runs in (default: the current folder).
    #[arg(long)]
    pub dir: Option<PathBuf>,
    /// The entry's PATH (default: this process's PATH).
    #[arg(long)]
    pub path: Option<String>,
    /// Another `NAME=VALUE` for the entry; token-like names are refused. Repeatable.
    #[arg(long)]
    pub env: Vec<String>,
    /// A merging entry instead of the default `--dry-run` one.
    #[arg(long)]
    pub merge: bool,
    /// The operator's permission reference the merging entry passes to `gate-loop`.
    #[arg(long)]
    pub merge_permission_reference: Option<String>,
}

impl Entry {
    pub fn request(&self) -> Result<moai_cli::gate_schedule::Request, String> {
        let moai_bin = match &self.moai_bin {
            Some(p) => p.clone(),
            None => std::env::current_exe().map_err(|_| "this binary's path unavailable")?,
        };
        let dir = match &self.dir {
            Some(d) => d.clone(),
            None => std::env::current_dir().map_err(|_| "current folder unavailable")?,
        };
        let path = match &self.path {
            Some(p) => p.clone(),
            None => std::env::var("PATH").map_err(|_| "PATH unavailable; pass --path")?,
        };
        Ok(moai_cli::gate_schedule::Request {
            home: self.home.clone(),
            os: self.os.clone(),
            repo: self.repo.clone(),
            moai_bin,
            dir,
            path,
            env: self.env.clone(),
            merge: self.merge,
            merge_permission_reference: self.merge_permission_reference.clone(),
        })
    }
}

/// Preview a scheduler entry for gate-loop
///
/// Preview a scheduler entry that runs `gate-loop --once` (#2906 S2): a macOS LaunchAgent
/// (`StartInterval 90`) or a Linux crontab file (`*/2`, under `flock -n`), the target path, the
/// enable and disable commands as text, and the `plan_digest`. The entry runs `--dry-run`
/// unless `--merge` is given with its permission reference; it never carries
/// `--review-confirmed`, and a token-like variable is refused. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    #[command(flatten)]
    entry: Entry,
}

pub const OPERATION: &str = "gate.schedule_plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let plan = moai_cli::gate_schedule::plan(&args.entry.request()?)?;
    serde_json::to_value(plan).map_err(|_| "encode failed".into())
}
