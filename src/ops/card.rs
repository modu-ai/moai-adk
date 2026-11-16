//! `card` (#3070): the project's local cards. One word, one action each; the logic is
//! `moai_cli::card::command`.

use std::path::PathBuf;

/// Work with the project's local cards
///
/// `moai card <action>` reads and writes the cards of the project's local ledger (#3068, #3070).
/// Every action's help starts with what it does to the ledger: `Reads the ledger only.` or
/// `Writes the ledger.` None of them creates the ledger file; that is `ledger-register-apply`.
#[derive(clap::Args)]
pub struct Args {
    #[command(subcommand)]
    action: Action,
}

#[derive(clap::Subcommand)]
enum Action {
    /// Reads the ledger only. List the project's cards, oldest first
    ///
    /// Reads the ledger only. Prints each card's name (`C-12`), title, kind, state, priority and
    /// source. A project without a ledger has no cards (`ledger: "absent"`, exit 0) and nothing is
    /// created.
    List {
        /// The project folder (its profile holds the project id).
        #[arg(default_value = ".")]
        project: PathBuf,
        /// Only cards in this state (`draft`, `ready`, `active`, `review`, `done`, `hold`).
        #[arg(long)]
        state: Option<String>,
        /// The HOME to read (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Reads the ledger only. Show one card with its criteria, notes and links
    ///
    /// Reads the ledger only. The card is named `C-12` (the ledger's number) or `#3066` (the number
    /// it came from on its code host; refused when more than one card has that number).
    Show {
        /// `C-12` or `#3066`.
        card: String,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME to read (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Add a card in the draft state
    ///
    /// Writes the ledger. Stores one `draft` card, its acceptance criteria (numbered from 1 in the
    /// order given) and one `card.add` event in a single transaction. The ledger must already
    /// exist; this action never creates it (`ledger-register-plan`, then `ledger-register-apply`).
    Add {
        /// The card's title: one line, 1 to 300 characters.
        #[arg(long)]
        title: String,
        /// The card's body (default empty; at most 1 MiB, no NUL).
        #[arg(long, default_value = "")]
        body: String,
        /// `task` (default), `feature`, `fix`, `docs`, `research`, `audit` or `chore`.
        #[arg(long)]
        kind: Option<String>,
        /// `p0` to `p3` (default `p2`).
        #[arg(long)]
        priority: Option<String>,
        /// A path the card may change, relative with `/` (repeat for more).
        #[arg(long = "scope")]
        scope: Vec<String>,
        /// An acceptance criterion, one line (repeat for more; numbered from 1).
        #[arg(long = "ac")]
        criteria: Vec<String>,
        /// Who adds the card (default `user`).
        #[arg(long)]
        by: Option<String>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
}

pub const OPERATION: &str = "card";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    use moai_cli::card::command as c;
    use moai_cli::ledger::command::resolve_home;
    match args.action {
        Action::List {
            project,
            state,
            home,
        } => c::list(&project, &resolve_home(home)?, state.as_deref()),
        Action::Show {
            card,
            project,
            home,
        } => c::show(&project, &resolve_home(home)?, &card),
        Action::Add {
            title,
            body,
            kind,
            priority,
            scope,
            criteria,
            by,
            project,
            home,
        } => c::add(
            &project,
            &resolve_home(home)?,
            &c::AddRequest {
                title: &title,
                body: &body,
                kind: kind.as_deref(),
                priority: priority.as_deref(),
                scope: &scope,
                criteria: &criteria,
                by: by.as_deref(),
            },
        ),
    }
}
