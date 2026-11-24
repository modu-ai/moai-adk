//! `leader` (SPEC-20261124-leader-brief): what the coordination window reads when it wakes. One
//! action today, `brief`; the logic is `moai_cli::leader_brief`.

use std::path::PathBuf;

/// What the coordination window reads
///
/// `moai leader <action>` reads the ledger and the lane registry for the window that leads the
/// project. It never sends a message, starts a session, writes a note or lands a card.
#[derive(clap::Args)]
pub struct Args {
    #[command(subcommand)]
    action: Action,
}

#[derive(clap::Subcommand)]
enum Action {
    /// Reads the ledger only. The short state a leader reads when it wakes
    ///
    /// Reads the ledger only. Prints at most 4096 bytes: `counts`, the landing queue (`queue`:
    /// card, head and the verdict `ref`), the unanswered blocks (`blocked`) and decisions
    /// (`decisions`), each with the number of its note, what was cut (`omitted`) and what could
    /// not be read (`unread`). It shrinks `card orchestrate plan` and writes nothing. A project
    /// without a ledger has empty lists (`ledger: "absent"`, exit 0) and nothing is created.
    Brief {
        /// The project folder (its profile holds the project id).
        #[arg(default_value = ".")]
        project: PathBuf,
        /// The HOME to read (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
}

pub const OPERATION: &str = "leader.brief";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    match args.action {
        Action::Brief { project, home } => moai_cli::leader_brief::brief(
            &project,
            &moai_cli::ledger::command::resolve_home(home)?,
        ),
    }
}
