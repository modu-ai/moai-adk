//! `check-tree` (#3009 S3a): warnings about the tracked MoAI folders of a project. Reads only;
//! the logic is `moai_cli::check_tree`.

use std::path::PathBuf;

/// Warn about what the tracked MoAI folders hold
///
/// Read the files git tracks under `.moai/` and `.moai-cli/` of a project (default `.`) and
/// warn about: a file over 1 MiB, text that holds `/Users/<name>` or `/home/<name>`, a session
/// UUID, a file named like a credential, and a project id a person chose instead of the random
/// one. It always answers (this is the warning stage), names the path and a count or a size
/// but never the matched text, and writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder (a git work tree).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "check.tree";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::check_tree::check(&args.project)
}
