//! `init` (#2835 W2b): initialize the current folder like a first launch, or with `--new
//! <name>` make `<HOME>/.moai/projects/<name>/` and initialize it. The logic is
//! `moai_cli::launcher_init::{init_here, init_new}`.

/// Initialize this folder or a new project
///
/// Without `--new`: the current folder, exactly as the first `cc|glm|codex` launch does it
/// (`git init` when needed, the project settings with their marker, `AGENTS.md`; existing files
/// are never touched; HOME itself and the file-system root are refused). With `--new <name>`:
/// create `<HOME>/.moai/projects/<name>/` (HOME is `$HOME`, on Windows `%USERPROFILE%`) and
/// initialize it the same way. A taken name, a name that is not one safe path component, and a
/// `~/.moai` or `~/.moai/projects` that is a symbolic link are refused before any write.
#[derive(clap::Args)]
pub struct Args {
    /// Create `<HOME>/.moai/projects/<name>/` and initialize it.
    #[arg(long, value_name = "NAME")]
    new: Option<String>,
    /// The host the default project settings name: `claude` (default), `codex` or `glm`.
    #[arg(long, default_value = "claude", value_parser = ["claude", "codex", "glm"])]
    host: String,
    /// Report what would be created and write nothing.
    #[arg(long)]
    dry_run: bool,
}

pub const OPERATION: &str = "project.init";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    match args.new {
        Some(name) => {
            let home = moai_cli::home_workdir::home()?;
            moai_cli::launcher_init::init_new(&home, &name, &args.host, args.dry_run)
        }
        None => {
            let cwd = std::env::current_dir()
                .map_err(|_| "the current directory is unreadable".to_string())?;
            moai_cli::launcher_init::init_here(&cwd, &args.host, args.dry_run)
        }
    }
}
