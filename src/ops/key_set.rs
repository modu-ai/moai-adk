//! `key-set` (#2393): stores a provider key, encrypted, in the vault; the value comes from
//! stdin, or with `--prompt` from a hidden prompt on the terminal (card C-54 S1). Moved out
//! of `src/main.rs` (#2834 R3).

use std::io::IsTerminal;

/// Store a provider key in the vault
///
/// Store a provider's key, encrypted, in the vault (#2393): for the current project, or
/// for every project with `--shared`. The value is read from stdin (never argv); an
/// interactive terminal is refused unless `--prompt` asks for the value there with echo
/// off. The answer never contains the value.
#[derive(clap::Args)]
pub struct Args {
    provider: String,
    /// The key every project uses (`~/.moai-cli/vault/_shared/`).
    #[arg(long)]
    shared: bool,
    /// Type the value at a hidden prompt (echo off); only when stdin is a terminal.
    #[arg(long)]
    prompt: bool,
}

pub const OPERATION: &str = "key.set";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    if args.prompt {
        return prompt(&args);
    }
    if std::io::stdin().is_terminal() {
        return Err(
            "key-set reads the value from stdin and refuses a terminal (the \
                    value would be echoed); pipe it in instead"
                .into(),
        );
    }
    let scope = moai_cli::key_vault::scope_here(args.shared)?;
    let value = moai_cli::key_vault::read_value(std::io::stdin())?;
    moai_cli::key_vault::set(
        &moai_cli::key_vault::home()?,
        &scope,
        &args.provider,
        &value,
    )
}

/// `--prompt`: the value is typed at the terminal with echo off. Everything that can refuse
/// without the value (the system, the terminal, the provider name, the scope) is checked
/// before anything is asked; stdin that is not a terminal is never read.
fn prompt(args: &Args) -> Result<serde_json::Value, String> {
    if !cfg!(unix) {
        return Err(moai_cli::key_prompt::UNSUPPORTED.into());
    }
    if !std::io::stdin().is_terminal() {
        return Err(
            "--prompt reads the value from a terminal with echo off, and stdin is not a \
             terminal; nothing was read or stored (run it in your own terminal window, or \
             drop --prompt and pipe the value in)"
                .into(),
        );
    }
    moai_cli::key_vault::check_provider(&args.provider)?;
    let scope = moai_cli::key_vault::scope_here(args.shared)?;
    let home = moai_cli::key_vault::home()?;
    let value = moai_cli::key_prompt::read_hidden(&args.provider)?;
    moai_cli::key_vault::set(&home, &scope, &args.provider, &value)
}
