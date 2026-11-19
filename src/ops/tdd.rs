//! `tdd` (card C-86 slice 3, SPEC-20261117-green-test-lock): this worktree's TDD cycle and its
//! GREEN test lock. The record and the log live in the worktree's own git folder; the logic is
//! `moai_cli::tdd`.

use serde_json::Value;
use std::path::PathBuf;
use std::time::Duration;

/// Record the TDD cycle that locks the RED tests during GREEN
///
/// `moai tdd <action>` keeps this worktree's TDD cycle in `<git-dir>/moai/tdd.json` and logs every
/// step in `<git-dir>/moai/tdd-log.jsonl` (never committed). While the cycle is in GREEN or
/// REFACTOR, the tool-preflight hook refuses writes to the tests the RED commit changed and new
/// test files. Runs on a SPEC work branch `spec/<slug>-<yyyymmdd>` (or the older
/// `spec/<yyyymmdd>-<slug>`).
#[derive(clap::Args)]
pub struct Args {
    #[command(subcommand)]
    action: Action,
}

#[derive(clap::Subcommand)]
enum Action {
    /// Writes the record. Record the head commit as the RED of a new cycle
    ///
    /// Writes the record. The head commit carries the trailer `Moai-Phase: red` and changes a
    /// test path; every --check runs at the head and must fail with assertion failures only (a
    /// build, network or empty-selection failure is not RED). The cycle starts in GREEN.
    Red {
        /// A check that shows the RED (an argv, no shell), 1 to 8.
        #[arg(long = "check", required = true)]
        checks: Vec<String>,
        /// Seconds each check may run.
        #[arg(long, default_value_t = 1800)]
        timeout_secs: u64,
        /// A folder inside the worktree.
        #[arg(long, default_value = ".")]
        project: PathBuf,
    },
    /// Writes the record. Move to REFACTOR once every RED check passes
    ///
    /// Writes the record. Runs the recorded checks at the head; all must pass. The lock stays.
    Green {
        /// Seconds each check may run.
        #[arg(long, default_value_t = 1800)]
        timeout_secs: u64,
        /// A folder inside the worktree.
        #[arg(long, default_value = ".")]
        project: PathBuf,
    },
    /// Writes the record. End the lock to change a test, with the reason
    ///
    /// Writes the record. Removes it and logs the reason; the next step is a new RED.
    Reopen {
        /// Why a locked test has to change (one line, up to 500 characters).
        #[arg(long)]
        reason: String,
        /// A folder inside the worktree.
        #[arg(long, default_value = ".")]
        project: PathBuf,
    },
    /// Writes the record. End the cycle; the log stays
    Done {
        /// A folder inside the worktree.
        #[arg(long, default_value = ".")]
        project: PathBuf,
    },
    /// Reads only. Show the cycle's phase, record and log
    Status {
        /// A folder inside the worktree.
        #[arg(long, default_value = ".")]
        project: PathBuf,
    },
}

pub const OPERATION: &str = "tdd";

pub fn run(args: Args) -> Result<Value, String> {
    use moai_cli::tdd;
    match args.action {
        Action::Red {
            checks,
            timeout_secs,
            project,
        } => tdd::red(&project, &checks, Duration::from_secs(timeout_secs)),
        Action::Green {
            timeout_secs,
            project,
        } => tdd::green(&project, Duration::from_secs(timeout_secs)),
        Action::Reopen { reason, project } => tdd::reopen(&project, &reason),
        Action::Done { project } => tdd::done(&project),
        Action::Status { project } => tdd::status(&project),
    }
}
