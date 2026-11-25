//! `build-number` (SPEC-20261125-install-build-number): the install build number is handed out
//! by moai, never picked by hand. The logic is `moai_cli::build_number`.

use std::path::PathBuf;

/// Hand out the next install build number
///
/// `next` records one new `{build}` number of the project's HeadVer head and prints it: always
/// above every number it recorded before and above every installed version folder of the head,
/// and never given twice, also when the build that asked for it fails. Use it for the build you
/// install (`MOAI_LOCAL_BUILD=<number>`); a build that only checks a change takes no number.
/// `show` reads the last number and the one `next` would give, and writes nothing. The record is
/// `build-number.log` in the project's state folder under HOME, which project registration
/// creates; a record that is not whole numbers is refused and left as it is.
#[derive(clap::Args)]
pub struct Args {
    #[command(subcommand)]
    action: Action,
    /// The project folder (any checkout of the repository).
    #[arg(long, global = true, default_value = ".")]
    project: PathBuf,
    /// The HOME that holds the record (default `$HOME`; tests pass a temporary folder).
    #[arg(long, global = true)]
    home: Option<PathBuf>,
}

#[derive(clap::Subcommand)]
enum Action {
    /// Record and print the next install build number.
    Next,
    /// Print the last recorded number and the next one; writes nothing.
    Show,
}

pub const OPERATION: &str = "build.number";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = moai_cli::ledger::command::resolve_home(args.home)?;
    match args.action {
        Action::Next => {
            let found = moai_cli::session_current::resolve(|name| std::env::var(name).ok());
            moai_cli::build_number::next(&args.project, &home, found["source_session_id"].as_str())
        }
        Action::Show => moai_cli::build_number::show(&args.project, &home),
    }
}
