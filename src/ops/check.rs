//! `check` (SPEC-20261120-check-affected): checks over a change. One word, one action; the
//! logic is `moai_cli::check_affected`.

use std::path::PathBuf;

/// Pick the tests a change can break and check them
///
/// `moai check affected <base>..<head>` picks, from the paths the range changes, every test
/// target of the same area (the file, the folder or the operation) and of the area's first
/// word, the changed test files, fixtures and shared helpers, every target
/// `tools/repo_scans.json` names and the tests that name this host's operating system. It
/// picks and runs no Python test. It adds the library and binary tests, `git diff --check`,
/// `cargo fmt`, `rustfmt --edition 2021` of each changed operation file, `bash -n` of each
/// changed script and `cargo clippy`; the tests run in one call.
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
    /// The cheap steps all run; when one fails, nothing after them runs. The cargo steps (clippy,
    /// one `test` call by `cargo nextest run`, or `cargo test` when nextest is missing) run under
    /// one heavy slot held from the first to the last; a slot that is not obtained gives its exit
    /// code to each of them and no cargo starts. The `test` step runs with a cap of 4 threads
    /// (`-j 4` for nextest, `RUST_TEST_THREADS=4` for `cargo test`); a `NEXTEST_TEST_THREADS` or
    /// `RUST_TEST_THREADS` the caller set stays as it is, and the slot folder's `limits.json` does
    /// not carry the cap yet. The `test` step also gets a private `TMPDIR`: a new
    /// `moai-affected-<id>` folder under the first root of the caller's `TMPDIR`, the macOS user
    /// temporary folder and `/tmp` that lies outside the worktree, inside no git repository and
    /// has no session id shape in its path (the other steps keep the caller's `TMPDIR`; with no
    /// such root the test step inherits it). A passed `test` step removes that folder and
    /// nothing else; a failed one keeps it. The answer's `test_tmpdir` says `removed`, `kept`
    /// (with its path), `left` (with the reason it was not removed), `inherited` (with the
    /// reason) or `not_run`. The answer has one `summary` line of `key=value` words in a fixed
    /// order: `check-affected`, base, head, targets, scans, `selection=lane` (only a `--lane` run
    /// writes it: its `summary.txt` is no receipt for a READY or REPAIRED note, a bare run is),
    /// `selection=card` (only a `--card` run that narrowed writes it; its `summary.txt` is a
    /// receipt for a READY or REPAIRED note, and a `--card` run that fell back to the full set
    /// writes the bare line) followed by `card_by=<reasons>` (the sorted `area:<name>`,
    /// `module:<word>` and `units:<module path>` reasons why the set is narrow, or `none`),
    /// `test-threads=<n>` (the cap the
    /// `test` step ran with), the exit status of each step in the order they ran, then
    /// `result=PASS` when every status is 0 and `result=FAIL` otherwise. A repository with no
    /// `Cargo.toml` at its top folder and no `.rs` file at the head gets no cargo step: it runs
    /// `git diff --check`, `bash -n` and one `sh -c` step per entry of `checks` in
    /// `.moai/project.yaml` (a heavy step each, in the declared order). When it declares none (no
    /// profile, one that cannot be read, or no `checks`) and no step failed, the result is
    /// `not_run`, never `PASS`: the line ends in `result=not_run`, the answer's `not_run.next`
    /// says what to write in the profile, and the READY note gate takes no receipt from it.
    /// Writes the step logs,
    /// `heavy-slot.log` and `summary.txt` (the summary line, whatever its result: the receipt
    /// `card note` reads before it keeps a READY or REPAIRED note of this head) under
    /// `<build folder>/check-affected-<head>` (or `--out`). The bare command
    /// selects the full set; `--lane` selects the precise lane set, a subset of it: a scan of
    /// `tools/repo_scans.json` joins only when a changed path is in its scope or is one of its
    /// rule files, a change to `tests/common/<sub>.rs` selects only the tests that name
    /// `common::<sub>`, and a test tagged with the host OS joins only by its area. A changed
    /// `skills/` or `docs/` path selects the tests whose source names that path, in both sets.
    /// The answer's `selection` field says `lane` or `full`; with `--integration-ref <ref>` its
    /// `range_note` warns when the base lies behind the merge base with that ref.
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
        /// Select the precise lane set instead of the full set (the default).
        #[arg(long)]
        lane: bool,
        /// Select what this card's change reaches: the tests the range changes or adds, the
        /// tests that name a changed module, the scans it touches. A source path that no name
        /// narrows (a `mod.rs`, a word that more than 10 targets carry, a path no test names)
        /// takes the tests of its area in `tests/areas.toml`, and the library's unit tests narrow
        /// to the changed modules; a wide path (dependencies, build or common settings,
        /// permission guards, ledger storage, the audit floor) gives the full set.
        #[arg(long, conflicts_with = "lane")]
        card: bool,
        /// The integration branch: a range whose base lies behind its merge base with the head
        /// gets a `range_note` in the answer.
        #[arg(long)]
        integration_ref: Option<String>,
        /// Run the full selection and read no earlier run: without this flag a run on a head
        /// whose earlier run failed re-runs only the tests that failed there (`rerun_of=` in the
        /// summary), and a head that moved runs the diff's targets and the earlier failures.
        #[arg(long)]
        full: bool,
    },
    /// Reads the ledger. Print the tests that passed only on a retry, with their counts
    ///
    /// Reads `check-flaky.jsonl` in the project's state folder: the test step of `check affected`
    /// retries a failing test twice (`--retries 2 --flaky-result pass`), lists one that passed on
    /// a retry as FLAKY in the summary line, in `flaky.txt` of the check folder and in this ledger
    /// (one line per run). The answer lists each name with its count; a name seen twice has
    /// `status` `fix`: fix that test. An unreadable ledger answers an error and no counts.
    Flaky {
        /// The repository (default: the current folder).
        #[arg(long, default_value = ".")]
        dir: PathBuf,
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
            lane,
            card,
            integration_ref,
            full,
        } => {
            let pick = if card {
                moai_cli::check_affected::Pick::Card
            } else if lane {
                moai_cli::check_affected::Pick::Lane
            } else {
                moai_cli::check_affected::Pick::Full
            };
            if list {
                return moai_cli::check_affected::list(
                    &dir,
                    &range,
                    pick,
                    integration_ref.as_deref(),
                );
            }
            let opts = moai_cli::check_affected::RunOptions {
                slot_dir,
                max,
                min_free_gb,
                out,
                pick,
                integration_ref,
                full,
            };
            let answer = moai_cli::check_affected::run(&dir, &range, &opts)?;
            // SPEC-20261124-ready-receipt-gate: the summary line is the receipt `card note` reads.
            if let (Some(out), Some(summary)) = (answer["out"].as_str(), answer["summary"].as_str())
            {
                moai_cli::card::ready_receipt::write_summary(std::path::Path::new(out), summary)?;
            }
            Ok(answer)
        }
        Action::Flaky { dir } => moai_cli::check_flaky::report(&dir),
    }
}
