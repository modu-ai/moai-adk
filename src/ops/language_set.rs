//! `language-set` (#2935 L1): the installer records the conversation language the user chose.
//! Writes at most one small file in HOME and never over an existing one; the logic is
//! `moai_cli::language::set_home`.

use clap::builder::PossibleValue;
use std::ffi::OsStr;

/// Record the conversation language in HOME
///
/// Record the language of the conversation with the model (#2935): writes
/// `config/language.yaml` in the HOME state folder, once. The code is written as given. An
/// existing file is never overwritten (`written: false`, `reason: exists`), and a HOME with no
/// MoAI state folder is not written (`reason: no_home_state`; nothing is created). A `config`
/// folder or file that is a link or the wrong type is refused. `moai language-setup` offers the
/// value as the conversation default when it is en, ko, ja or zh. The text MoAI prints stays
/// English whatever the code.
#[derive(clap::Args)]
pub struct Args {
    /// Scope: the user's HOME (`$HOME`). Without it the command refuses (exit 1), so a later
    /// scope can be added.
    #[arg(long)]
    home: bool,
    /// Language code: 2-3 letters and optional `-` groups, at most 35 characters (en, pt-BR)
    #[arg(value_parser = Code, hide_possible_values = true)]
    code: String,
}

pub const OPERATION: &str = "language.set";

/// Any language-code-shaped text (`language::is_language_tag`); anything else is a usage error.
/// It names the table's codes as its possible values only because the usage-error path
/// (`operation_of`) fills a required argument from the first one; they are hidden from the help
/// and do not limit what is accepted.
#[derive(Clone)]
struct Code;

impl clap::builder::TypedValueParser for Code {
    type Value = String;

    fn parse_ref(
        &self,
        cmd: &clap::Command,
        _arg: Option<&clap::Arg>,
        value: &OsStr,
    ) -> Result<String, clap::Error> {
        match value.to_str() {
            Some(code) if moai_cli::language::is_language_tag(code) => Ok(code.to_owned()),
            _ => Err(clap::Error::raw(
                clap::error::ErrorKind::ValueValidation,
                "not a language code: use 2-3 letters and optional `-` groups of 2-8 letters or digits, at most 35 characters (en, ko, pt-BR)\n",
            )
            .with_cmd(cmd)),
        }
    }

    fn possible_values(&self) -> Option<Box<dyn Iterator<Item = PossibleValue> + '_>> {
        Some(Box::new(
            moai_cli::language::CODES
                .iter()
                .map(|(c, _, _)| PossibleValue::new(*c)),
        ))
    }
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
