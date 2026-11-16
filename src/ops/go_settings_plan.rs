//! `go-settings-plan` (#2936 L2a): what cleaning the Go-made `.moai/` references out of a
//! project's `.claude/settings.json` would do. Reads only; the logic is
//! `moai_cli::go_settings_clean`.

use std::path::PathBuf;

/// Preview cleaning Go-era settings keys
///
/// Preview cleaning the Go-generation `.moai/` references out of `.claude/settings.json`
/// (#2936 L2a): a `statusLine` or `subagentStatusLine` that runs the Go `.moai/status_line.sh`
/// exactly, and `plansDirectory` set to exactly `.moai/plans`. The two status lines go or become
/// the Rust launcher's direct path once that launcher is shown to be the Rust build (otherwise
/// the plan is `held` and writes nothing); any value that differs by a character is yours and
/// is only named. Key names only, never a value. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The project (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "go_settings.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = std::env::var_os("HOME").map(PathBuf::from);
    moai_cli::go_settings_clean::plan(&args.project, home.as_deref())
}
