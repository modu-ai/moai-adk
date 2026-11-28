//! `tidy-restore` (SPEC-20261126-home-tidy-apply): put the files of a tidy archive back. The
//! logic is `moai_cli::tidy::restore`.

use std::path::PathBuf;

/// Put the files of a tidy archive back
///
/// Read `<HOME>/.moai/archive/<NAME>` and restore its items (those named with `--pick`, all of
/// them without one). An item is checked whole against its `list.json` before any of its files
/// is written; each file is created only where nothing exists, with the listed mode and time.
/// Never overwrites or removes what exists (such paths are named), keeps the archive, follows no
/// link, starts no command and writes nothing outside `<HOME>/.moai`.
#[derive(clap::Args)]
pub struct Args {
    /// User home root whose `.moai` folder holds the archive.
    home: PathBuf,
    /// The archive name `tidy-apply` answered (a UTC time).
    #[arg(long, value_name = "NAME")]
    archive: String,
    /// One item path of the archive to restore (repeat for more).
    #[arg(long = "pick", value_name = "PATH")]
    pick: Vec<String>,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "tidy.restore";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::tidy::restore::restore(
        &args.home,
        &args.archive,
        &args.pick,
        &args.permission_reference,
    )
}
