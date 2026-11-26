//! `project-check` (SPEC-20261122-project-ci-design REQ-5): the local check of a project. The
//! logic is `moai_cli::project_check`.

use std::path::PathBuf;

/// Run the commands the project declares
///
/// Runs the check commands saved in the project profile, in the declared order, in the project
/// folder: the same list the generated CI workflow runs. Every step runs even when an earlier
/// one failed; a command still running after `--timeout-secs` is stopped and counted as a
/// failure. The answer names each step with its exit status and ends with one `summary` line and
/// `result=PASS` or `result=FAIL`. A project with no declared checks is refused with a plain
/// explanation. It writes nothing itself; the commands are the ones a person confirmed.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder (its profile holds the declared checks).
    project: PathBuf,
    /// Stop a command that runs longer than this many seconds.
    #[arg(long, default_value_t = moai_cli::project_check::DEFAULT_TIMEOUT_SECS)]
    timeout_secs: u64,
}

pub const OPERATION: &str = "project.check";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::project_check::run(
        &args.project,
        &moai_cli::project_check::Options {
            timeout_secs: args.timeout_secs,
        },
    )
}
