//! `check` (SPEC-20261120-check-affected): checks over a change. One word, one action; the
//! logic is `moai_cli::check_affected`.

use std::path::PathBuf;

/// Pick the tests a change can break and check them
///
/// `moai check affected <base>..<head>` picks, from the paths the range changes, every test
/// target of the same area (the file, the folder or the operation) and of the area's first
/// word, the changed test files, fixtures and shared helpers, every target
/// `tools/repo_scans.json` names, the Python tests of the area and the tests that name this
/// host's operating system. It adds the library and binary tests, `git diff --check`,
/// `cargo fmt`, `rustfmt --edition 2021` of each changed operation file, `bash -n` of each
/// changed script and `cargo clippy`.
#[derive(clap::Args)]
pub struct Args {
    #[command(subcommand)]
    action: Action,
}

#[derive(clap::Subcommand)]
enum Action {
    /// Runs commands. Pick the tests a range can break, run them, answer in one line
    ///
    /// Runs commands (`--list` only reads). The range is `<base>..<head>` (two dots); the changed
    /// paths are `<base>...<head>`, so a head behind its base's tip counts only its own changes.
    /// The checkout must be at the head with no modified tracked file, otherwise nothing runs.
    /// Every step runs, a failure does not stop the rest; each cargo step holds one heavy slot.
    /// The answer has one `summary` line of `key=value` words in a fixed order: `check-affected`,
    /// base, head, targets, py, scans, the exit status of each step in the order they ran, then
    /// `result=PASS` when every status is 0 and `result=FAIL` otherwise. Writes the step logs
    /// under `<dir>/target/check-affected-<head>` (or `--out`).
    Affected {
        /// The range, `<base>..<head>`.
        range: String,
        /// Print the selection and the steps; run and write nothing.
        #[arg(long)]
        list: bool,
        /// The repository (default: the current folder).
        #[arg(long, default_value = ".")]
        dir: PathBuf,
        /// The heavy slot folder (default: the machine's).
        #[arg(long)]
        slot_dir: Option<PathBuf>,
        /// The most heavy builds at once.
        #[arg(long, default_value_t = 4)]
        max: u32,
        /// The free disk space, in GB, a heavy build needs.
        #[arg(long, default_value_t = 50)]
        min_free_gb: u32,
        /// Where the step logs go.
        #[arg(long)]
        out: Option<PathBuf>,
    },
}

pub const OPERATION: &str = "check.affected";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    match args.action {
        Action::Affected {
            range,
            list,
            dir,
            slot_dir,
            max,
            min_free_gb,
            out,
        } => {
            if list {
                return moai_cli::check_affected::list(&dir, &range);
            }
            let opts = moai_cli::check_affected::RunOptions {
                slot_dir,
                max,
                min_free_gb,
                out,
            };
            moai_cli::check_affected::run(&dir, &range, &opts)
        }
    }
}
