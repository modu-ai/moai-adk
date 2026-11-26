//! `claude-md-plan`: what a project's old instruction files hold, and which of their lines are
//! the project's own. Reads only; the logic is `moai_cli::project_update::claude_md`.

use std::path::PathBuf;

/// Preview the old instruction files and their project lines
///
/// Preview the instruction files of a Go-generation project: `CLAUDE.md`, `AGENTS.md`,
/// `AGENTS.local.md`, `CLAUDE.local.md` and every file under the state folder's
/// `docs/local-rules/` that exists, each with its sha256, line count, git tracking, what it was
/// compared with and the numbers of the lines that are the project's own. A file whose bytes
/// equal the Go manifest's record has none; with the recorded version found in git history only
/// the lines that version lacks count; otherwise every line counts. A source that is a link,
/// lies below a link folder, is not UTF-8 text or is larger than 1 MiB is named under `refused`.
///
/// The project lines are shown as hunks, one per heading section, with the id
/// `<path>:<first line>`; a hunk equal to an earlier one only names it. A line that holds a
/// credential-shaped value is counted and never printed. `--keep <id>=shared` puts a hunk into
/// the new `CLAUDE.md`, which follows the managed block of the profile; `--keep <id>=personal`
/// puts it into `CLAUDE.local.md`. The answer shows both texts, warnings, one plan digest and
/// one `claude-md-apply` command. A refused source, an unusable profile or a shared hunk with a
/// credential-shaped line is named under `apply_blocked`, and then no command is printed.
/// Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The project (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
    /// Keep one hunk of the answer: `<id>=shared` (the new `CLAUDE.md`) or `<id>=personal`
    /// (`CLAUDE.local.md`). Repeat once per hunk.
    #[arg(long = "keep", value_name = "ID=DESTINATION")]
    keep: Vec<String>,
}

pub const OPERATION: &str = "claude_md.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::project_update::claude_md::plan(&args.project, &args.keep)
}
