//! `tidy-inventory` (SPEC-20261124-home-tidy-slice2): list every entry of the home moai folder
//! with a class. Reads only; the logic is `moai_cli::tidy`.

use std::path::PathBuf;

/// List every entry of the home moai folder with its class
///
/// List every top-level entry of `<HOME>/.moai` with its class (`owned_live`, `owned_stale`,
/// `retired_tool`, `other_product`, `unknown` or `protected`), size, newest change and the reason,
/// and one row per folder kind that grows (reports, logs, evidence, worktrees, build, receipts)
/// with its count, size and the thresholds it passes. A key, a credential, a link or a special
/// file is listed by its own directory data only and never opened or followed. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// User home root whose `.moai` folder is inspected.
    home: PathBuf,
    /// Override one threshold, `<kind>.<field>=<value>` (repeat for more), for example
    /// `reports.max_files=500`.
    #[arg(long = "threshold", value_name = "KIND.FIELD=VALUE")]
    threshold: Vec<String>,
    /// Measure this memory index file (`MEMORY.md`): lines, bytes and its share of the host's
    /// load cap (repeat for more). The file is only read.
    #[arg(long = "memory-index", value_name = "PATH")]
    memory_index: Vec<PathBuf>,
    /// Entries visited below one top-level entry before its size is reported as `over_limit`.
    #[arg(long, default_value_t = moai_cli::tidy::DEFAULT_ENTRY_LIMIT)]
    entry_limit: usize,
    /// Seconds the whole run may take before the remaining sizes are reported as `over_limit`.
    #[arg(long, default_value_t = moai_cli::tidy::DEFAULT_TIME_LIMIT_SECS)]
    time_limit_secs: u64,
}

pub const OPERATION: &str = "tidy.inventory";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let opts = moai_cli::tidy::Options::new(
        &args.threshold,
        args.memory_index,
        args.entry_limit,
        args.time_limit_secs,
    )?;
    moai_cli::tidy::run_inventory(&args.home, &opts)
}
