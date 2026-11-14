//! `codex-doctor`: read-only Codex check of the AGENTS chain, skill duplicates and `.codex/`.
//! Moved out of `src/main.rs` unchanged (#2834 R3).

use std::path::PathBuf;

/// Check the Codex instructions and skills
///
/// Read-only Codex doctor: AGENTS chain and budget, skill duplicates, `.codex/` layer.
#[derive(clap::Args)]
pub struct Args {
    /// Project root (Codex walks from here down to the working directory).
    project: PathBuf,
    /// Working directory inside the project (default: the project root).
    #[arg(long)]
    cwd: Option<PathBuf>,
    /// Codex home holding the global AGENTS files (default: $CODEX_HOME, else ~/.codex).
    #[arg(long)]
    codex_home: Option<PathBuf>,
    /// Home holding `.agents/skills` (default: $HOME).
    #[arg(long)]
    home: Option<PathBuf>,
    /// `project_doc_fallback_filenames` entry (repeatable; config.toml is not read).
    #[arg(long)]
    fallback: Vec<String>,
    /// `project_doc_max_bytes` (default 32768; config.toml is not read).
    #[arg(long)]
    max_bytes: Option<u64>,
}

pub const OPERATION: &str = "codex.doctor";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = args
        .home
        .or_else(|| std::env::var_os("HOME").map(PathBuf::from))
        .ok_or("home directory missing; pass --home")?;
    let codex_home = args
        .codex_home
        .or_else(|| std::env::var_os("CODEX_HOME").map(PathBuf::from))
        .unwrap_or_else(|| home.join(".codex"));
    let cwd = args.cwd.unwrap_or_else(|| args.project.clone());
    moai_cli::hosts::codex::doctor(
        &args.project,
        &cwd,
        &codex_home,
        &home,
        &args.fallback,
        args.max_bytes,
    )
}
