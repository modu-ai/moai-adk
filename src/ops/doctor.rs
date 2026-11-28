//! `doctor`: read-only diagnosis of the moai layer of a Claude Code setup
//! (SPEC-20261127-doctor-moai-layer).

use std::path::PathBuf;

/// Diagnose the moai layer of your Claude Code setup
///
/// Read-only: checks the Claude config folder, the Claude Code version against the features
/// moai uses, the login source and an API key that overrides it, the managed locks and the
/// inbound message value in the settings files, and the host's capabilities, and names the
/// Claude Code checks for everything else. Prints variable names and states only, never a secret
/// value.
#[derive(clap::Args)]
pub struct Args {
    /// The project whose `.claude/settings*.json` are read (default: the current folder).
    #[arg(long, default_value = ".")]
    project: PathBuf,
    /// Also run `claude auth status` once and keep its auth method, config folder and expiry.
    #[arg(long)]
    auth_status: bool,
    /// Also print the whole capability table, every host and feature with its docs line.
    #[arg(long)]
    table: bool,
    /// Read the managed settings files from this folder (default: the platform's own folder).
    #[arg(long)]
    managed_dir: Option<PathBuf>,
}

pub const OPERATION: &str = "doctor";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    Ok(moai_cli::moai_doctor::run(
        &moai_cli::moai_doctor::Options {
            project: args.project,
            auth_status: args.auth_status,
            managed_dir: args.managed_dir,
            table: args.table,
            // The command edge resolves HOME once; `gather` reads none (home-env-readers).
            home: moai_cli::key_vault::home(),
        },
    ))
}
