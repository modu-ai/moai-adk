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
    /// Writes the ledger. Make a draft card ready to be taken
    ///
    /// Writes the ledger. A `draft` (or `hold`) card becomes `ready`, with one `card.ready` event.
    /// The ledger must already exist.
    Ready {
        /// `C-12`.
        card: String,
        /// Who does it (default `user`).
        #[arg(long)]
        by: Option<String>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Take a ready card for a holder
    ///
    /// Writes the ledger. Reserves the card for this session: the next generation, a lease from the database
    /// clock, its write scope (the card's own unless `--scope` is given) and `ready` → `active`.
    /// Refused when the card is not `ready`, is already held, or its scope overlaps another
    /// card's reservation (`conflict`). The answer's `generation` is what `release` and
    /// `handoff` need. The holder is this session's own id (`CLAUDE_CODE_SESSION_ID` or
    /// `CODEX_THREAD_ID`), never typed; a session without one is refused.
    Take {
        /// `C-12`.
        card: String,
        /// Lease length in seconds, 60 to 86400 (default 14400).
        #[arg(long)]
        lease_secs: Option<i64>,
        /// A path pattern the holder may write, repeat for more (default: the card's own scope).
        #[arg(long = "scope")]
        scope: Vec<String>,
        /// The worktree the holder works in.
        #[arg(long)]
        worktree: Option<String>,
        /// The branch the holder works on.
        #[arg(long)]
        branch: Option<String>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Give a taken card back
    ///
    /// Writes the ledger. Ends this session's reservation at the given generation; an `active`
    /// card becomes `ready`. A stale generation or a card held by another session is refused
    /// (`conflict`).
    Release {
        /// `C-12`.
        card: String,
        /// The generation `take` or `handoff` answered with.
        #[arg(long)]
        generation: i64,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Hand a taken card over to another holder
    ///
    /// Writes the ledger. Ends this session's reservation and gives `--to` the next generation with
    /// the same scope, worktree, branch and lease end; the card stays `active`. The old generation
    /// is stale from then on. Needs a live lease.
    Handoff {
        /// `C-12`.
        card: String,
        /// The generation this session holds.
        #[arg(long)]
        generation: i64,
        /// The session id of the receiving holder (hex digits and `-`, 8 to 64 characters).
        #[arg(long)]
        to: String,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
}

pub const OPERATION: &str = "card";

/// One environment variable, for the session id (a value that is not UTF-8 reads as absent).
fn env(name: &str) -> Option<String> {
    std::env::var(name).ok()
}

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    use moai_cli::card::command as c;
    use moai_cli::card::dispatch as d;
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
        Action::Ready {
            card,
            by,
            project,
            home,
        } => d::ready(&project, &resolve_home(home)?, &card, by.as_deref()),
        Action::Take {
            card,
            lease_secs,
            scope,
            worktree,
            branch,
            project,
            home,
        } => d::take(
            &project,
            &resolve_home(home)?,
            &card,
            &d::TakeRequest {
                holder: &d::session_holder("take", env)?,
                lease_secs,
                scope: &scope,
                worktree: worktree.as_deref(),
                branch: branch.as_deref(),
            },
        ),
        Action::Release {
            card,
            generation,
            project,
            home,
        } => d::release(
            &project,
            &resolve_home(home)?,
            &card,
            &d::session_holder("release", env)?,
            generation,
        ),
        Action::Handoff {
            card,
            generation,
            to,
            project,
            home,
        } => d::handoff(
            &project,
            &resolve_home(home)?,
            &card,
            &d::session_holder("handoff", env)?,
            generation,
            &to,
        ),
    }
}
