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
    /// Codex home holding the global AGENTS files (default: $CODEX_HOME for your own home, an
    /// absolute path; else ~/.codex).
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
    // `--codex-home` wins; else the one host-folder rule (card C-79): `CODEX_HOME` for the
    // account's home (absolute only), else `<home>/.codex`.
    let codex_home = match args.codex_home {
        Some(dir) => dir,
        None => moai_cli::host_home::resolve(moai_cli::host_home::Host::Codex, &home)?,
    };
    // The report is JSON text: a folder name that is not UTF-8 is refused, never a panic
    // (C-79 audit N3).
    if codex_home.to_str().is_none() {
        return Err(
            "the Codex home (CODEX_HOME or --codex-home) is not UTF-8; codex-doctor reads only \
             UTF-8 paths"
                .into(),
        );
    }
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
