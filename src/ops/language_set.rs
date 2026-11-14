//! `language-set` (#2935 L1): the installer records the conversation language the user chose.
//! Writes at most one small file in HOME and never over an existing one; the logic is
//! `moai_cli::language::set_home`.

/// Record the conversation language in HOME
///
/// Record the language of the conversation with the model (#2935): writes
/// `config/language.yaml` in the HOME state folder, once. An existing file is never
/// overwritten (`written: false`, `reason: exists`), and a HOME with no MoAI state folder is not
/// written (`reason: no_home_state`; nothing is created). A `config` folder or file that is a
/// link or the wrong type is refused. `moai language-setup` then offers the value as the
/// conversation default. The text MoAI prints stays English whatever the code.
#[derive(clap::Args)]
pub struct Args {
    /// Scope: the user's HOME (`$HOME`). Required, so a later scope can be added.
    #[arg(long)]
    home: bool,
    /// Language code: en, ko, ja or zh.
    #[arg(value_parser = codes())]
    code: String,
}

pub const OPERATION: &str = "language.set";

/// The accepted codes, from the one table (`language::CODES`); a parser built from it lets the
/// usage-error path (`operation_of`) fill the argument with a valid code.
fn codes() -> Vec<&'static str> {
    moai_cli::language::CODES.iter().map(|(c, _, _)| *c).collect()
}

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    // `--home` names the scope. It is a plain flag (the usage-error path fills a required one
    // with a stray value), so a missing one is refused here, before anything is read or written.
    if !args.home {
        return Err("language-set requires --home".into());
    }
    let home = moai_cli::key_vault::home()
        .map_err(|_| "home directory missing; language-set requires HOME".to_string())?;
    moai_cli::language::set_home(&home, &args.code)
}
