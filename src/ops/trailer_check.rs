//! `trailer-check` (SPEC-20261117-session-trailer, SDD-6; SPEC-20261119-verified-trailer): the
//! identity trailers of a range.

use std::path::PathBuf;

/// Report commits with missing or malformed trailers
///
/// Read the commits of RANGE (merges skipped, oldest first) and report each missing or
/// malformed `Moai-Spec:` / `Moai-Session:` / `Moai-Verified:` trailer (user decision U-B
/// 2026-11-17). A missing `Moai-Spec` fails only on a `spec/` branch; a malformed value always
/// fails; a missing `Moai-Session` is reported and never fails (the audit gate then reads the
/// authors as unknown). A missing `Moai-Verified` line fails only after the rule's start commit
/// — the range's first commit carrying the line, or the commit `--verified-since` names (itself
/// exempt); before it the line reads as `none`. With `--card-numbers-since`, a commit after
/// that commit whose subject or body carries the card-number shape `C-<n>` or `c-<n>` fails as
/// `card_number` (quoted ledger output, `AC-<n>`, versions and hexadecimal SHAs pass); without
/// the flag the rule is off. `--message-file` judges one draft commit message with that scan
/// instead of a range, always on — the commit-msg hook's mode. `ok` is true when nothing
/// fails; otherwise the result is printed and the command exits 1. A value holding the log's
/// field separator is `malformed_trailers`. Reads only.
#[derive(clap::Args)]
pub struct Args {
    /// Revision range, for example `<base>..HEAD`; leave out with `--message-file`.
    range: Option<String>,
    /// Repository to read.
    #[arg(long, default_value = ".")]
    dir: PathBuf,
    /// The branch the range is judged for, when the checkout is detached (CI).
    #[arg(long)]
    branch: Option<String>,
    /// The revision the Moai-Verified rule starts at, in place of the range's first commit
    /// carrying the line (that commit itself stays exempt).
    #[arg(long)]
    verified_since: Option<String>,
    /// The revision the card-number rule starts at (that commit itself stays exempt); the rule
    /// is off without it.
    #[arg(long)]
    card_numbers_since: Option<String>,
    /// A draft commit message to judge with the card-number scan instead of a range (the
    /// commit-msg hook's input, git hands the file).
    #[arg(long)]
    message_file: Option<PathBuf>,
}

pub const OPERATION: &str = "trailers.check";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    if let Some(file) = args.message_file.as_deref() {
        if args.range.is_some()
            || args.branch.is_some()
            || args.verified_since.is_some()
            || args.card_numbers_since.is_some()
        {
            return Err("give --message-file alone: the draft scan needs no range or rules".into());
        }
        return moai_cli::trailers::message_file_check(file);
    }
    let range = args.range.as_deref().ok_or(
        "give a revision range such as <base>..HEAD, or --message-file for one draft message",
    )?;
    moai_cli::trailers::check_rules(
        &args.dir,
        range,
        args.branch.as_deref(),
        args.verified_since.as_deref(),
        args.card_numbers_since.as_deref(),
    )
}
