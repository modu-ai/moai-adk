//! `card` (#3070): the project's local cards. One word, one action each; the logic is
//! `moai_cli::card::command`, `dispatch` and `work` (card C-53: the work and finish path).

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
    /// Reads the ledger only. Writes files only inside --out. Export every card to a folder
    ///
    /// Reads the ledger only. Writes files only inside --out. Writes `index.json`, `events.jsonl`
    /// and one `C-<n>.md` per card (title, body, criteria, notes, links) into an empty or new
    /// folder; the same ledger gives the same bytes. Nothing already in the folder is replaced,
    /// and a project without a ledger gets `ledger not found` and no folder.
    ///
    /// The export is private to you (new folder 0700, files 0600; an existing empty folder that
    /// others can open is refused: `chmod 700` it first) and holds what the ledger holds: card
    /// text, criteria, notes, links and every event with its payload, which includes reservation
    /// holders (session ids) and lease times. Look at it before you share it. The folder may not be inside
    /// `<HOME>/.moai`. Links in the parent folders are followed; the real path is printed as `out`.
    Export {
        /// The folder to write: new, or empty and closed to other users (its parent must exist).
        #[arg(long)]
        out: PathBuf,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
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
    /// card becomes `ready`. A `review` card is refused: `card reopen` it first. A stale generation or a card held by another session is refused
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
    /// Reads the ledger only. Show the card to work on next
    ///
    /// Reads the ledger only. Prints the cards this session holds (`mine`: generation, lease
    /// end, worktree, branch), then the `ready` cards it could take, by priority and then by
    /// number, then every card it cannot take with the reason (`held` by another session,
    /// `scope_overlap` with another card's reservation, `empty_scope`: `take` refuses a card
    /// without a write scope). `next` names the card to work on now: one this session holds,
    /// else the first `ready` one. A project without a ledger has no cards (`ledger: "absent"`,
    /// exit 0) and nothing is created.
    Next {
        /// The project folder (its profile holds the project id).
        #[arg(default_value = ".")]
        project: PathBuf,
        /// The HOME to read (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Append a note to a card, the text on stdin
    ///
    /// Writes the ledger. Stores the text read from stdin as one note with one `card.note`
    /// event. Notes are only ever added, never changed or removed, and anyone may add one: a
    /// decision is answered by someone who does not hold the card. The event keeps the id of the
    /// session that wrote the note, whatever `--by` says, and `card show` prints it. The text is
    /// at most 1 MiB, with no control character but tabs and line breaks, and no word with the
    /// shape of a secret: `card export` writes notes out. The answer's `ref` names the note.
    Note {
        /// `C-12`.
        card: String,
        /// `decision`, `report`, `observation` or `comment`.
        #[arg(long)]
        kind: String,
        /// A name for the note, unique on its card (default: a random one).
        #[arg(long = "ref")]
        reference: Option<String>,
        /// Who signs it (default: this session's id, or `user` without one). A `decision`
        /// needs it.
        #[arg(long)]
        by: Option<String>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Mark an acceptance criterion met or withdrawn
    ///
    /// Writes the ledger. As this session, holding the card at `--generation` with a live lease:
    /// `--met --evidence REF` marks criterion N met, `--withdraw --reason TEXT` withdraws it. A
    /// stale generation or a card held by another session is refused (`conflict`).
    Criterion {
        /// `C-12`.
        card: String,
        /// The criterion's number on the card, from 1.
        number: i64,
        /// The generation `take` or `handoff` answered with.
        #[arg(long)]
        generation: i64,
        /// The criterion is met; needs `--evidence`.
        #[arg(long, requires = "evidence", conflicts_with_all = ["withdraw", "reason"], required_unless_present = "withdraw")]
        met: bool,
        /// Where the proof is: a command and its result, a run, a commit (one line).
        #[arg(long, requires = "met")]
        evidence: Option<String>,
        /// The criterion no longer applies; needs `--reason`.
        #[arg(long, requires = "reason")]
        withdraw: bool,
        /// Why it is withdrawn (one line).
        #[arg(long, requires = "withdraw")]
        reason: Option<String>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Link a pull request and record its state
    ///
    /// Writes the ledger. Reads the pull request's state from the code host (one GET, nothing is
    /// written there) and stores it on the card's link: `open`, `closed`, or `merged` with the
    /// merge commit. Run it again after the merge; `done` needs a merged link. As this session,
    /// holding the card at `--generation`; another session is refused before the host is asked.
    Pr {
        /// `C-12`.
        card: String,
        /// The pull request's number.
        number: u64,
        /// The generation `take` or `handoff` answered with.
        #[arg(long)]
        generation: i64,
        /// The host project, written `owner/repo`. A card that came from a host project takes
        /// only that project (the default); any other card needs this.
        #[arg(long)]
        repo: Option<String>,
        /// The host: `github` (default: the card's own host, else `github`).
        #[arg(long)]
        from: Option<String>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Send a taken card to review
    ///
    /// Writes the ledger. An `active` card becomes `review`, with one `card.review` event. As
    /// this session, holding the card at `--generation` with a live lease; the reservation
    /// stays.
    Review {
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
    /// Writes the ledger. Take a card in review back to active
    ///
    /// Writes the ledger. A `review` card becomes `active` again, with one `card.reopen` event:
    /// the way back when the review found something to repair. As this session, holding the
    /// card at `--generation` with a live lease; the reservation stays.
    Reopen {
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
    /// Writes the ledger. Close a card whose pull request is merged
    ///
    /// Writes the ledger. A `review` card becomes `done` and its reservation ends, when at least
    /// one criterion is met, none is open and a merged pull request is linked (`card pr` after
    /// the merge). Anything else is refused with the rule that is not met. As this session,
    /// holding the card at `--generation`.
    Done {
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
    /// Writes the ledger. Extend the lease of a taken card
    ///
    /// Writes the ledger. The lease of this session's reservation at `--generation` ends
    /// `--lease-secs` from now, never earlier than it already did. Needs a lease that has not
    /// ended; the generation stays the same.
    Renew {
        /// `C-12`.
        card: String,
        /// The generation `take` or `handoff` answered with.
        #[arg(long)]
        generation: i64,
        /// Lease length in seconds from now, 60 to 86400 (default 14400).
        #[arg(long)]
        lease_secs: Option<i64>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Reads the ledger only. Check that this session may write a path
    ///
    /// Reads the ledger only. Exit 0 only when this session holds the card at `--generation`
    /// with a live lease and every `--path` lies under the reservation's write scope. Anything
    /// else is refused (`conflict`). Nothing is recorded. The paths are compared as text:
    /// a symbolic link on disk is not followed.
    Check {
        /// `C-12`.
        card: String,
        /// The generation `take` or `handoff` answered with.
        #[arg(long)]
        generation: i64,
        /// A path to write, relative to the repository and written with `/` (repeat for more).
        #[arg(long = "path", required = true)]
        path: Vec<String>,
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
    use moai_cli::card::work as w;
    use moai_cli::ledger::workflow::CriterionChange;
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
        Action::Export { out, project, home } => {
            c::export(&project, &resolve_home(home)?, &out)
        }
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
        Action::Next { project, home } => w::next(
            &project,
            &resolve_home(home)?,
            d::session_holder("next", env).ok().as_deref(),
        ),
        Action::Note {
            card,
            kind,
            reference,
            by,
            project,
            home,
        } => {
            let home = resolve_home(home)?;
            // The one action that reads standard input (the help rule in `src/main.rs` reads
            // this arm).
            let text = w::read_note_text(std::io::stdin().lock())?;
            w::note(
                &project,
                &home,
                &card,
                &w::NoteRequest {
                    kind: &kind,
                    text: &text,
                    reference: reference.as_deref(),
                    by: by.as_deref(),
                    session: d::session_holder("note", env).ok().as_deref(),
                },
            )
        }
        Action::Criterion {
            card,
            number,
            generation,
            met: _,
            evidence,
            withdraw: _,
            reason,
            project,
            home,
        } => {
            let change = match (&evidence, &reason) {
                (Some(evidence), None) => CriterionChange::Met { evidence },
                (None, Some(reason)) => CriterionChange::Withdrawn { reason },
                _ => {
                    return Err(
                        "card criterion refused: give --met --evidence REF or --withdraw --reason TEXT"
                            .to_string(),
                    )
                }
            };
            w::criterion(
                &project,
                &resolve_home(home)?,
                &card,
                &d::session_holder("criterion", env)?,
                generation,
                number,
                change,
            )
        }
        Action::Pr {
            card,
            number,
            generation,
            repo,
            from,
            project,
            home,
        } => {
            let home = resolve_home(home)?;
            let holder = d::session_holder("pr", env)?;
            let transport = moai_cli::forge::transport::GhTransport::from_env()?;
            let clock = moai_cli::forge::policy::SystemClock;
            let connector = moai_cli::forge::connector::GithubConnector::new(
                &transport,
                &clock,
                moai_cli::forge::Config::default(),
            );
            w::pr(
                &project,
                &home,
                &card,
                &w::PrRequest {
                    holder: &holder,
                    generation,
                    number,
                    from: from.as_deref(),
                    repo: repo.as_deref(),
                },
                &connector,
            )
        }
        Action::Review {
            card,
            generation,
            project,
            home,
        } => w::review(
            &project,
            &resolve_home(home)?,
            &card,
            &d::session_holder("review", env)?,
            generation,
        ),
        Action::Reopen {
            card,
            generation,
            project,
            home,
        } => w::reopen(
            &project,
            &resolve_home(home)?,
            &card,
            &d::session_holder("reopen", env)?,
            generation,
        ),
        Action::Done {
            card,
            generation,
            project,
            home,
        } => w::done(
            &project,
            &resolve_home(home)?,
            &card,
            &d::session_holder("done", env)?,
            generation,
        ),
        Action::Renew {
            card,
            generation,
            lease_secs,
            project,
            home,
        } => w::renew(
            &project,
            &resolve_home(home)?,
            &card,
            &d::session_holder("renew", env)?,
            generation,
            lease_secs,
        ),
        Action::Check {
            card,
            generation,
            path,
            project,
            home,
        } => w::check(
            &project,
            &resolve_home(home)?,
            &card,
            &d::session_holder("check", env)?,
            generation,
            &path,
        ),
    }
}
