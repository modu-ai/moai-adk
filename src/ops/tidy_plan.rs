//! `tidy-plan` (SPEC-20261124-home-tidy-slice2): the read-only tidy plan of the home moai
//! folder. The logic is `moai_cli::tidy`.

use std::path::PathBuf;

/// Preview the tidy of the home moai folder
///
/// Preview the tidy of `<HOME>/.moai`: each candidate has an action (`archive`, `delete` or
/// `keep`), the rule that made it a candidate and the reason it is kept. A folder kind over a
/// threshold offers what its age rule reaches, oldest first; what needs git proof is kept as
/// `proof_pending`; another product's entries, unknown entries and protected entries are never
/// offered. The only `delete` is an empty folder of a retired tool. `digest` is the digest of
/// the candidate list. Writes nothing and starts no command.
#[derive(clap::Args)]
pub struct Args {
    /// User home root whose `.moai` folder is inspected.
    home: PathBuf,
    /// Override one threshold, `<kind>.<field>=<value>` (repeat for more), for example
    /// `reports.max_files=500`.
    #[arg(long = "threshold", value_name = "KIND.FIELD=VALUE")]
    threshold: Vec<String>,
    /// Measure this memory index file (`MEMORY.md`); above 80 percent of the host's load cap the
    /// plan prints the three-group memory tidy for the user to run in a Claude session.
    #[arg(long = "memory-index", value_name = "PATH")]
    memory_index: Vec<PathBuf>,
    /// Entries visited below one top-level entry before its size is reported as `over_limit`.
    #[arg(long, default_value_t = moai_cli::tidy::DEFAULT_ENTRY_LIMIT)]
    entry_limit: usize,
    /// Seconds the whole run may take before the remaining sizes are reported as `over_limit`.
    #[arg(long, default_value_t = moai_cli::tidy::DEFAULT_TIME_LIMIT_SECS)]
    time_limit_secs: u64,
}

pub const OPERATION: &str = "tidy.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let opts = moai_cli::tidy::Options::new(
        &args.threshold,
        args.memory_index,
        args.entry_limit,
        args.time_limit_secs,
    )?;
    moai_cli::tidy::run_plan(&args.home, &opts)
}
