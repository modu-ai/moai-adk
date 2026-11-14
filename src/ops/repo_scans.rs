//! `repo-scans` (#2708 4/4, formerly tools/repo_scans.py): the repo-wide scan judge of the
//! merge train (#2681). The logic is `moai_cli::repo_scans`.

use std::io::Read;
use std::path::PathBuf;

/// List the repo-wide scan tests a merge must run
///
/// List the repo-wide scan tests a merge must run first (#2681; #2708 moved it from
/// tools/repo_scans.py). With `--refs <tip> <head>` the change sets are `git diff --name-only`
/// from their merge-base and the table is the union of `tools/repo_scans.json` at both refs;
/// without it, stdin `{"pr": [...], "tip": [...]}` is judged with the fallback table. The result
/// is `{remerge, hits, checks, reasons}`; any internal error fails safe to `remerge: true` with
/// every check, and no readable table means hold the merge. Reads only.
#[derive(clap::Args)]
pub struct Args {
    /// `<tip> <head>`: judge the two refs of this repository.
    #[arg(long, num_args = 0..=3, value_names = ["TIP", "HEAD"])]
    refs: Option<Vec<String>>,
    /// The repository that runs Git (default: the current directory).
    #[arg(long, default_value = ".")]
    repo: PathBuf,
    /// The fallback table (default: `<repo>/tools/repo_scans.json`).
    #[arg(long)]
    table: Option<PathBuf>,
}

pub const OPERATION: &str = "repo_scans.judge";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    use moai_cli::repo_scans::{report, Input, INPUT_LIMIT, TABLE_IN_REPO};
    let input = match args.refs {
        Some(r) if r.len() == 2 => Input::Refs(r[0].clone(), r[1].clone()),
        Some(r) => Input::Bad(format!(
            "usage: repo-scans --refs <tip> <head> (got {} values)",
            r.len()
        )),
        None => {
            let mut text = String::new();
            match std::io::stdin()
                .take(INPUT_LIMIT as u64 + 1)
                .read_to_string(&mut text)
            {
                Ok(_) => Input::Stdin(text),
                Err(_) => Input::Bad("stdin is not readable UTF-8".into()),
            }
        }
    };
    let table = args
        .table
        .unwrap_or_else(|| args.repo.join(TABLE_IN_REPO));
    Ok(report(&input, &args.repo, &table))
}
