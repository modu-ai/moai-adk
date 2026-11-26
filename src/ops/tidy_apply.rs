//! `tidy-apply` (SPEC-20261126-home-tidy-apply): archive or delete the picked items of the
//! tidy plan. The logic is `moai_cli::tidy::apply`.

use std::path::PathBuf;

/// Archive or delete the picked items of the tidy plan
///
/// Recompute `tidy-plan` and, only when its `digest` equals `--confirm`, act on the items named
/// with `--pick`. An archive item is packed into `<HOME>/.moai/archive/<UTC>` as a compressed
/// archive with a `list.json`, then read back and compared with the list and the source.
/// Removes the listed files one by one only after that check (a folder only when empty; never
/// recursively). A delete item is an empty folder. An item with a secret-shaped name, a link, a
/// special or unreadable part is skipped whole, as is a retired-tool item while a Go session may
/// be live. Starts no command and changes nothing outside `<HOME>/.moai`.
#[derive(clap::Args)]
pub struct Args {
    /// User home root whose `.moai` folder is tidied.
    home: PathBuf,
    /// One plan item path to act on (repeat for more).
    #[arg(long = "pick", value_name = "PATH")]
    pick: Vec<String>,
    /// The `digest` of the plan the picks were made from.
    #[arg(long, value_name = "DIGEST")]
    confirm: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
    /// The same thresholds the plan was read with, `<kind>.<field>=<value>` (repeat for more).
    #[arg(long = "threshold", value_name = "KIND.FIELD=VALUE")]
    threshold: Vec<String>,
    /// Entries visited below one item before it is skipped as `over_limit`.
    #[arg(long, default_value_t = moai_cli::tidy::DEFAULT_ENTRY_LIMIT)]
    entry_limit: usize,
    /// Seconds one walk may take before what remains is `over_limit`.
    #[arg(long, default_value_t = moai_cli::tidy::DEFAULT_TIME_LIMIT_SECS)]
    time_limit_secs: u64,
}

pub const OPERATION: &str = "tidy.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let opts = moai_cli::tidy::Options::new(
        &args.threshold,
        Vec::new(),
        args.entry_limit,
        args.time_limit_secs,
    )?;
    let request = moai_cli::tidy::apply::Request {
        picks: &args.pick,
        confirm: &args.confirm,
        permission: &args.permission_reference,
    };
    moai_cli::tidy::apply::apply(&args.home, &opts, &request)
}
