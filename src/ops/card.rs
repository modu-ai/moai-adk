//! `card` (#3070): the project's local cards. One word, one action each; the logic is
//! `moai_cli::card::command`, `dispatch` and `work` (card C-53: the work and finish path), and
//! `spec_link` for `card spec` (SDD-4).

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
        /// Only cards whose title starts with this text (case-sensitive, taken as typed; not empty,
        /// no control character). The flaky rule's cards are `flaky: <test>`: `--title-prefix
        /// "flaky: "`.
        #[arg(long)]
        title_prefix: Option<String>,
        /// Show withdrawn cards too: the default list omits every card a `card.withdraw` event
        /// marks.
        #[arg(long)]
        all: bool,
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
    /// Reads the ledger only. Show the events of one card, oldest first
    ///
    /// Reads the ledger only. The ledger is opened read-only with a wait of at most 250
    /// milliseconds, the events are read in one read transaction (at most `--limit`, 1 to 500,
    /// default 50) and the connection is closed before anything is printed, so a reader never
    /// holds a writer's commit. Each event has its time, kind, session, lane name (the author of
    /// the newest note that session wrote on this card, else null) and the plain fields of its
    /// payload; a note event also shows the note's author and first line. The answer names the
    /// current owner, and `next` and `end` are opaque cursors: hand `next` back as `--after` to
    /// read what follows, with no gap and no overlap; `--after` never waits for new events. The
    /// `renew` bookkeeping events are left out unless `--all`; a rejected call (`fence_reject`)
    /// always shows. The card is named `C-12`.
    Timeline {
        /// `C-12`.
        card: String,
        /// The `next` or `end` of an earlier answer for this card: only the events after it.
        #[arg(long)]
        after: Option<String>,
        /// The most events to read, 1 to 500 (default 50).
        #[arg(long, allow_negative_numbers = true)]
        limit: Option<i64>,
        /// Also show the bookkeeping events (`renew`).
        #[arg(long)]
        all: bool,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME to read (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Reads the ledger only. Show what a dispatch should know before it starts: commits and
    /// PRs already naming the card, its SPEC folders, and decision notes colliding with the
    /// card's scope or criteria
    ///
    /// Reads the ledger only. Also reads the project's git history (nothing is written, no
    /// network is used): commits of `--integration-ref` (default: the current branch's
    /// upstream) whose subject, body or `Moai-Spec:` trailer names the card, the SPEC the
    /// ledger links to it, the SPEC folders naming it, and every `decision` note sharing a
    /// path with the card's write scope or a phrase from its criteria (card C-96, W3).
    Precheck {
        /// `C-12` or `#3066`.
        card: String,
        /// The integration ref to scan (default: the current branch's upstream).
        #[arg(long)]
        integration_ref: Option<String>,
        /// Scan at most this many commits (default 500; the report names a cut-short scan).
        #[arg(long)]
        max_commits: Option<usize>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
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
    /// Writes the ledger. Set the write scope of a card, or widen a held card's
    ///
    /// Writes the ledger. Replaces the write scope of a `draft` or `ready` card that nobody holds,
    /// with one `card.scope` event that keeps the old scope, the new one and `--by`. `card next`
    /// skips a card without a write scope (`empty_scope`); this is how such a card gets one.
    /// A card that has an active reservation (`active` or `review`) is widened instead, under
    /// the same generation: no release and no retake, and the holder, lease, worktree and branch
    /// stay. A held card's scope only grows: the list must keep every pattern the reservation
    /// has (a list that leaves one out is a `conflict` naming it), and the added patterns must
    /// not overlap another card's reservation (a `conflict` naming that card). `--reason` is
    /// required for a held card; the `card.scope` event records `--by`, the reason, the
    /// generation and the holder; the card's own scope is not changed, so a later `card take`
    /// without `--scope` still reads the old one. `--generation` is the generation the caller
    /// read (`card show`): given, the card's live reservation must be at that generation, else
    /// `conflict` naming the generation it has (or that the card is not held) and nothing is
    /// written; left out, nothing is checked. Each entry is checked as `add --scope` checks it.
    /// Any other state is refused (`conflict`) with its state and holder, and nothing is
    /// written.
    Scope {
        /// `C-12`.
        card: String,
        /// A path the card may change, relative with `/` (repeat for more; replaces them all, and
        /// for a held card the list must contain the current patterns plus the new ones).
        #[arg(long = "scope", required = true)]
        scope: Vec<String>,
        /// Who sets it, one line (for example `leader`).
        #[arg(long)]
        by: String,
        /// Why, one line of 1 to 500 characters; required to widen a held card.
        #[arg(long)]
        reason: Option<String>,
        /// The generation the card has now (`card show` prints it); a card at another
        /// generation, or one nobody holds, is a `conflict`. Left out, nothing is checked.
        #[arg(long, value_parser = clap::value_parser!(i64).range(1..))]
        generation: Option<i64>,
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
    /// `CODEX_THREAD_ID`), never typed; a session without one is refused. The answer's `spec`
    /// is the SPEC the card names (`card spec`) and its status; `warnings` names a card that
    /// changes the repository without an approved SPEC, and `--require-approved-spec` refuses
    /// that card instead (`invalid_input`, nothing taken).
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
        /// Refuse, instead of warn, when a card that changes the repository has no approved SPEC.
        #[arg(long)]
        require_approved_spec: bool,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Withdraw a draft or ready card
    ///
    /// Writes the ledger. Marks the card withdrawn: it moves to `hold`, one `card.withdraw`
    /// event records the reason and the by with the writing session, `card take` refuses it
    /// naming the reason, and the default `card list` hides it (`--all` shows it). The ledger
    /// schema is unchanged: the six state values are pinned by its CHECK, so the marker is the
    /// free event kind, not a new value. The card must be `draft` or `ready` and held by
    /// nobody.
    Withdraw {
        /// The card to withdraw (`C-12` or `#3066`).
        card: String,
        /// Why the card is withdrawn (one line, 1..=500 characters).
        #[arg(long)]
        reason: String,
        /// Who withdraws it (one line, 1..=200 characters).
        #[arg(long)]
        by: String,
        /// The project folder (its profile holds the project id).
        #[arg(default_value = ".")]
        project: PathBuf,
        /// The HOME to use (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Record that this card's head landed on the local integration branch
    ///
    /// Writes the ledger. A `review` card gains one `card.landed` event naming the head and the
    /// by, with the writing session as the actor (M-Flow, SPEC-20261120-card-done-mflow). The
    /// head must be 40 lowercase hex. `card done` then closes the card without a pull request
    /// when the profile's `work_flow` is `m-flow` and the head sits inside the profile's
    /// integration branch.
    Landed {
        /// The card whose head landed (`C-12` or `#3066`).
        card: String,
        /// The head sha that landed (40 lowercase hex).
        #[arg(long)]
        head: String,
        /// Who records the landing (one line, 1..=200 characters).
        #[arg(long)]
        by: String,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME to use (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Write the typed verdict of an audit on a card in review
    ///
    /// Writes the ledger. One `card.verdict` event for a card in `review`: the commit it judges
    /// (`--head`, a full 40-digit id), `--result` `pass` or `fail` and an optional `--evidence`
    /// reference, with this session as the actor. Refused (`conflict`) when this session holds or
    /// ever held a reservation of the card (the agent that implemented a change never writes its
    /// verdict), when an auditor was assigned and it is another session, and when the card is not
    /// in review; nothing is written then. The same session repeating the same head and result
    /// writes no second row (`written` is false). No landing reads the row yet.
    Verdict {
        /// The card, `C-12`.
        card: String,
        /// The commit the verdict judges (40 lowercase hex).
        #[arg(long)]
        head: String,
        /// `pass` or `fail`.
        #[arg(long)]
        result: String,
        /// A reference to what the audit read (a note id, a report path), one line.
        #[arg(long)]
        evidence: Option<String>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME to use (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Name the auditor of a card in review
    ///
    /// Writes the ledger. One `card.audit_assigned` event for a card in `review` with the
    /// auditor's identity and `--by leader`; only the leader word assigns. Refused (`conflict`)
    /// when the identity holds or ever held a reservation of the card, and when the card is not
    /// in review; nothing is written then.
    Assign {
        /// The card, `C-12`.
        card: String,
        /// The auditor's identity (a session id), one word.
        #[arg(long)]
        auditor: String,
        /// Who assigns: `leader`.
        #[arg(long)]
        by: String,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME to use (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Link a card to the SPEC it works on, or unlink it
    ///
    /// Writes the ledger. Names the SPEC (`SPEC-<yyyymmdd>-<slug>`) the card works on, with one
    /// `card.spec` event; `--unlink` removes the link. The SPEC must be in the project's
    /// `.moai/specs` and pass `moai spec check` without an error; it is only read, and it never
    /// names the card: the link lives in this machine's ledger. `card next` and `card take` then
    /// print the SPEC's status and warn while a card that changes the repository (any kind but
    /// `research` and `audit`) has no approved SPEC. The same link again writes nothing.
    Spec {
        /// `C-12`.
        card: String,
        /// The SPEC id.
        #[arg(required_unless_present = "unlink", conflicts_with = "unlink")]
        spec: Option<String>,
        /// Remove the card's link.
        #[arg(long)]
        unlink: bool,
        /// Who does it (default `user`).
        #[arg(long)]
        by: Option<String>,
        /// The project folder (its profile holds the project id; its `.moai/specs` holds the SPEC).
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
    /// Writes the ledger. Take over a card whose reservation has ended
    ///
    /// Writes the ledger. For a card whose holder is gone: the reservation at `--generation` must be
    /// `held` with a lease that has ended. Refusals write nothing and come in four kinds:
    /// `conflict` (a live lease, another generation, a `done` card, a reservation that is not
    /// `held`, or a card nobody holds), `not_found` (no such card), `usage` (exit 2: an option
    /// that is missing or cannot be read, such as no `--generation` or `--generation x`) and
    /// `invalid_input` (exit 1: a value out of range, a blank or longer `--reason`, a `--by` over
    /// 80 characters, a `--lease-secs` outside 60 to 86400, or a `--reason` with a word shaped
    /// like a secret). The reservation
    /// becomes `reclaimed`, and this session gets the next generation with the same scope, worktree
    /// and branch and a new lease; the card keeps its state, notes, criteria and links, so this
    /// session can then close it with `card criterion`, `card pr` and `card done`. To give a card
    /// back instead: a card in `review` is reopened with `card reopen` first, then given back with
    /// `card release` (`release` refuses a card in `review`). One `card.reclaim` event names the old
    /// holder, generation and lease end, the new holder, `--by` and `--reason`. Nothing checks that
    /// the old holder is really gone: say in `--reason` what was found (process ended, pull request
    /// merged, uncommitted work). The old generation is stale from then on. The new holder is this
    /// session's own id, never typed.
    Reclaim {
        /// `C-12`.
        card: String,
        /// The generation the card has now (`card show` prints it).
        #[arg(long)]
        generation: i64,
        /// Who reclaims it, one line of 1 to 80 characters: the actor of the event.
        #[arg(long)]
        by: String,
        /// What was found out about the old holder, one line of 1 to 500 characters.
        #[arg(long)]
        reason: String,
        /// Lease length of the new reservation in seconds, 60 to 86400 (default 14400).
        #[arg(long)]
        lease_secs: Option<i64>,
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
    /// else the first `ready` one. Each held and ready card carries its SPEC and that SPEC's
    /// status (`spec`); `warnings` names each card that changes the repository without an
    /// approved SPEC. A project without a ledger has no cards (`ledger: "absent"`, exit 0) and
    /// nothing is created.
    Next {
        /// The project folder (its profile holds the project id).
        #[arg(default_value = ".")]
        project: PathBuf,
        /// The HOME to read (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Reads the ledger only. List the lane notes that still wait for the leader
    ///
    /// Reads the ledger only. A request is a note whose first line starts with `감사 요청` (or is
    /// a `LANE-DONE` line that says `audit requested`: `audit_request`), `LANE-BLOCKED`
    /// (`blocked`), `LANE-READY-INTEGRATION` (`ready_for_integration`) or `DECISION-QUESTION`
    /// (`decision_question`). On each card that is not `done`, the newest request of each
    /// family is listed unless the card holds a later note (a larger note id) written by the
    /// leader. Notes imported from a code host are left out. The items come oldest first with
    /// their age in seconds; `LANE-NOTE` progress lines are not requests. A project without a
    /// ledger has nothing waiting (`ledger: "absent"`, exit 0) and nothing is created.
    Inbox {
        /// The project folder (its profile holds the project id).
        #[arg(default_value = ".")]
        project: PathBuf,
        /// The author name of the leader's notes (default `leader`).
        #[arg(long)]
        leader: Option<String>,
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
    /// Writes the ledger. Add acceptance criteria to a card
    ///
    /// Writes the ledger. Appends each `--ac` text to the card as an open criterion, numbered
    /// after the card's highest (a withdrawn one keeps its number), with one
    /// `card.criterion_add` event that records the session that added them next to `--by` (null
    /// when the caller has no session id): criteria are only added, never changed or removed. In
    /// the answer, `session` is that id and `criteria` is the number of the card's criteria
    /// (withdrawn ones included), not its highest number. It needs no generation and no holder, so
    /// a card nobody holds, such as a migrated Issue card, can be given one; the reservation of a
    /// held card is left as it is; a `done` card is refused.
    CriterionAdd {
        /// `C-12`.
        card: String,
        /// An acceptance criterion, one line (repeat for more; numbered after the card's last).
        #[arg(long = "ac", required = true)]
        criteria: Vec<String>,
        /// Who adds it (default `user`).
        #[arg(long)]
        by: Option<String>,
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
    /// Writes the ledger. Send a taken card to review and give its paths back
    ///
    /// Writes the ledger. An `active` card becomes `review`, with one `card.review` event. As
    /// this session, holding the card at `--generation` with a live lease. Every path of the
    /// reservation that no `--keep` names is given back in the same transaction, so it is free
    /// for any card at once; the answer lists `given_back` and `kept`. A `--keep` path the
    /// reservation does not hold is refused and nothing is written. A card sent back with
    /// `card reopen` widens its scope again with `card scope` and a reason.
    Review {
        /// `C-12`.
        card: String,
        /// The generation `take` or `handoff` answered with.
        #[arg(long)]
        generation: i64,
        /// A path of the reservation to keep for the review (repeat for more); the others are
        /// given back.
        #[arg(long)]
        keep: Vec<String>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Give paths of a held card back
    ///
    /// Writes the ledger. The holder removes the named `--path`s from the reservation of this
    /// `--generation`; a path given back is free for any card at once. A path the reservation
    /// does not hold, a stale generation or another session is refused and no reservation, scope
    /// or card changes (a stale generation or another session leaves the ledger's usual
    /// `fence_reject` event). One `card.scope` event names the paths. A reopened card widens its scope again with
    /// `card scope` and a reason.
    GiveBack {
        /// `C-12`.
        card: String,
        /// The generation `take` or `handoff` answered with.
        #[arg(long)]
        generation: i64,
        /// A path of the reservation to give back (repeat for more).
        #[arg(long = "path", required = true)]
        paths: Vec<String>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Lease a shared file for one command
    ///
    /// Writes the ledger. Adds the named `--path`s to the reservation, runs the command given
    /// after `--` and gives the added paths back in one call: a short lease for a line in a
    /// shared file. While another card's reservation covers a path the call retries until
    /// `--wait-secs` (default 60, at most 600), then is refused naming that card and writes
    /// nothing. The command's standard output goes to standard error, so the answer stays one
    /// JSON envelope. A command that exits non-zero or cannot start is refused with its status
    /// named, after the lease gives the paths back. A path given back is free for any card at
    /// once. If this process is killed while the command runs, the leased paths stay in the
    /// reservation until `card give-back` frees them. A reopened card widens its scope again
    /// with `card scope` and a reason.
    LineLease {
        /// `C-12`.
        card: String,
        /// The generation `take` or `handoff` answered with.
        #[arg(long)]
        generation: i64,
        /// A path to lease (repeat for more).
        #[arg(long = "path", required = true)]
        paths: Vec<String>,
        /// The longest wait for another card's reservation, in seconds.
        #[arg(long)]
        wait_secs: Option<u64>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
        /// The command to run while the lease is held, after `--`.
        #[arg(last = true, required = true)]
        command: Vec<String>,
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
    /// Writes the ledger. Close a card whose pull request is merged, or whose head landed on
    /// the local integration branch
    ///
    /// Writes the ledger. A `review` card becomes `done` and its reservation ends, when at least
    /// one criterion is met, none is open and a merged pull request is linked (`card pr` after
    /// the merge) — or, when the profile's `work_flow` is `m-flow`, the card's `card landed`
    /// head is contained in the profile's integration branch. Anything else is refused with the
    /// rule that is not met. As this session, holding the card at `--generation`.
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
    /// Writes the ledger. Close a ready, active or review card whose landing or report is proven
    ///
    /// Writes the ledger. Closes the card in one transaction: every open criterion is marked met
    /// with its `--met N=REF` (one per open criterion), the card becomes `done` and a reservation
    /// it holds is released, with one `card.landed` and one `card.finish` event naming `--by`.
    /// A card that changes the repository needs a landing: `--head`, else its merged pull request
    /// link, else the head `card landed` recorded, and git must show that head inside the
    /// profile's integration branch (`work_flow` m-flow). When an open criterion of such a card
    /// has no `--met`, the proven landing is recorded once (one `card.landed` event with the
    /// head, the tip, the generation and `--by`), nothing else changes, and the answer says
    /// `closed` false and lists `open_criteria`. A `research` or `audit` card gives exactly one
    /// of `--evidence-report` (a file under `<HOME>/.moai/reports` or the project folder) or
    /// `--evidence-note` (a report or decision note of the card) and no `--head`, and every open
    /// criterion needs its `--met`. Another session's reservation whose lease has not ended is
    /// refused; your own, or one that has ended, passes. With `--generation N` the call is
    /// refused before any write unless N is the generation of the card's newest reservation (0
    /// for a card that never had one). A card that is already `done` is refused, and when two
    /// calls reach one card at the same time one closes it and the other is refused. Every
    /// refusal writes nothing.
    Finish {
        /// The card to close (`C-12`).
        card: String,
        /// The head that landed (40 lowercase hex); not for a research or audit card.
        #[arg(long)]
        head: Option<String>,
        /// One open criterion and its evidence, `N=REF` (repeat for each open criterion; REF is
        /// one line of 1..=500 characters).
        #[arg(long = "met")]
        met: Vec<String>,
        /// A report file under `<HOME>/.moai/reports` or the project folder (research and audit
        /// cards only; a relative path is read from the project folder).
        #[arg(long)]
        evidence_report: Option<String>,
        /// The `ref` of a report or decision note of the card (research and audit cards only).
        #[arg(long)]
        evidence_note: Option<String>,
        /// The generation of the card's newest reservation you expect (0 when it never had
        /// one); another number refuses the call before any write. Not checked when absent.
        #[arg(long)]
        generation: Option<i64>,
        /// Who finishes the card (one line, 1..=200 characters).
        #[arg(long)]
        by: String,
        /// The project folder (its profile holds the project id and the integration branch).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger and the integration branch. Move the branch to a candidate commit and close the card
    ///
    /// Writes the ledger and, when the candidate is not inside the tip yet, the profile's
    /// integration branch (`work_flow` m-flow): one `git update-ref` that moves the branch to
    /// `--head` only while it still points at the tip that was read, then the same close as `card
    /// finish` with `--head` as the landed head. Only `--by leader` lands until the typed verdict
    /// exists. Everything `card finish` would refuse is refused before the branch moves; so is a
    /// candidate that is no commit, a branch checked out in a worktree, a SPEC that fails at the
    /// candidate and a candidate that does not have the tip on its first-parent line
    /// (`stale_base`). A branch that moved in between is refused as `tip_moved`. A close refused
    /// after the move leaves the branch moved and says so; the same call run again skips the move
    /// and closes. A card already `done` with the same head answers done and writes nothing.
    Land {
        /// The card to land (`C-12`).
        card: String,
        /// The candidate commit (40 lowercase hex) that has the integration tip on its
        /// first-parent line.
        #[arg(long)]
        head: String,
        /// One open criterion and its evidence, `N=REF`, as for `card finish`.
        #[arg(long = "met")]
        met: Vec<String>,
        /// The generation of the card's newest reservation you expect; not checked when absent.
        #[arg(long)]
        generation: Option<i64>,
        /// Who lands the card: only `leader` is accepted.
        #[arg(long)]
        by: String,
        /// The project folder (its profile holds the project id and the integration branch).
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

/// `card criterion-add C-n --ac TEXT… [--by W]`: what is wrong with the request is found before
/// the ledger is opened (the card's name, each text, `--by`), then the ledger appends the
/// criteria in one transaction (`workflow::add_criteria_as`, cards C-108 and C-125). The event and
/// the answer carry the caller's session id as `card.note` does (`null` without one); `--by`
/// stays the author.
fn criterion_add(
    project: &std::path::Path,
    home: &std::path::Path,
    reference: &str,
    entries: &[String],
    by: Option<&str>,
) -> Result<serde_json::Value, String> {
    use moai_cli::card::command as c;
    use moai_cli::card::dispatch as d;
    use moai_cli::card::input::{self, CardRef};
    use moai_cli::ledger::workflow::{self, WorkflowError};
    let refused = |why: String| format!("card criterion-add refused: {why}");
    let card_id = match input::parse_ref(reference).map_err(|e| refused(e.to_string()))? {
        CardRef::Id(n) => n,
        CardRef::SourceNumber(_) => {
            return Err(refused(
                "name the card by its C-number (`card show #N` finds it)".to_string(),
            ))
        }
    };
    let texts = input::criterion_texts(entries).map_err(|e| refused(e.to_string()))?;
    let by = c::actor(by, "criterion-add")?;
    // No session id is not a refusal here: nobody has to hold the card, so the event says `null`.
    let session = d::session_holder("criterion-add", env).ok();
    let (project_id, mut ledger) = c::open_for_write(project, home)?;
    let added = workflow::add_criteria_as(&mut ledger, card_id, &by, session.as_deref(), &texts)
        .map_err(|e| match e {
            WorkflowError::InvalidInput(why) => refused(why),
            other => other.to_string(),
        })?;
    Ok(serde_json::json!({
        "action": "criterion-add", "project_id": project_id,
        "name": input::card_name(card_id), "card_id": card_id,
        "by": by, "session": session, "added": added.added, "criteria": added.total,
        "rev": added.rev, "events": 1, "writes": true,
    }))
}

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    use moai_cli::card::command as c;
    use moai_cli::card::dispatch as d;
    use moai_cli::card::work as w;
    use moai_cli::ledger::command::resolve_home;
    use moai_cli::ledger::workflow::CriterionChange;
    match args.action {
        Action::List {
            project,
            state,
            title_prefix,
            all,
            home,
        } => c::list_filtered(
            &project,
            &resolve_home(home)?,
            state.as_deref(),
            title_prefix.as_deref(),
            all,
        ),
        Action::Show {
            card,
            project,
            home,
        } => {
            let home = resolve_home(home)?;
            let mut shown = c::show(&project, &home, &card)?;
            shown["roles"] = moai_cli::card::roles::roles_of_shown(&project, &home, &shown)?;
            Ok(shown)
        }
        Action::Timeline {
            card,
            after,
            limit,
            all,
            project,
            home,
        } => moai_cli::card::timeline::timeline(
            &project,
            &resolve_home(home)?,
            d::session_holder("timeline", env).ok().as_deref(),
            &card,
            after.as_deref(),
            limit,
            all,
        ),
        Action::Precheck {
            card,
            integration_ref,
            max_commits,
            project,
            home,
        } => moai_cli::card::precheck::precheck(
            &card,
            integration_ref.as_deref(),
            max_commits,
            &project,
            home.as_deref(),
        )
        .map_err(|r| r.message),
        Action::Export { out, project, home } => c::export(&project, &resolve_home(home)?, &out),
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
        Action::Scope {
            card,
            scope,
            by,
            reason,
            generation,
            project,
            home,
        } => d::scope(
            &project,
            &resolve_home(home)?,
            &card,
            &scope,
            &by,
            reason.as_deref(),
            generation,
        ),
        Action::Withdraw {
            card,
            reason,
            by,
            project,
            home,
        } => c::withdraw(&project, &resolve_home(home)?, &card, &reason, &by),
        Action::Landed {
            card,
            head,
            by,
            project,
            home,
        } => w::landed(&project, &resolve_home(home)?, &card, &head, &by),
        Action::Take {
            card,
            lease_secs,
            scope,
            worktree,
            branch,
            require_approved_spec,
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
                require_approved_spec,
            },
        ),
        Action::Verdict {
            card,
            head,
            result,
            evidence,
            project,
            home,
        } => moai_cli::card::roles::verdict(
            &project,
            &resolve_home(home)?,
            &card,
            &moai_cli::card::roles::VerdictRequest {
                head: &head,
                result: &result,
                evidence: evidence.as_deref(),
                session: &d::session_holder("verdict", env)?,
            },
        ),
        Action::Assign {
            card,
            auditor,
            by,
            project,
            home,
        } => moai_cli::card::roles::assign(&project, &resolve_home(home)?, &card, &auditor, &by),
        Action::Spec {
            card,
            spec,
            unlink: _,
            by,
            project,
            home,
        } => moai_cli::card::spec_link::link(
            &project,
            &resolve_home(home)?,
            &card,
            spec.as_deref(),
            by.as_deref(),
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
        Action::Reclaim {
            card,
            generation,
            by,
            reason,
            lease_secs,
            project,
            home,
        } => d::reclaim(
            &project,
            &resolve_home(home)?,
            &card,
            &d::ReclaimRequest {
                to_holder: &d::session_holder("reclaim", env)?,
                generation,
                by: &by,
                reason: &reason,
                lease_secs,
            },
        ),
        Action::Next { project, home } => w::next(
            &project,
            &resolve_home(home)?,
            d::session_holder("next", env).ok().as_deref(),
        ),
        Action::Inbox {
            project,
            leader,
            home,
        } => moai_cli::card::inbox::inbox(&project, &resolve_home(home)?, leader.as_deref()),
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
                _ => return Err(
                    "card criterion refused: give --met --evidence REF or --withdraw --reason TEXT"
                        .to_string(),
                ),
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
        Action::CriterionAdd {
            card,
            criteria,
            by,
            project,
            home,
        } => criterion_add(
            &project,
            &resolve_home(home)?,
            &card,
            &criteria,
            by.as_deref(),
        ),
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
            keep,
            project,
            home,
        } => w::review_keeping(
            &project,
            &resolve_home(home)?,
            &card,
            &d::session_holder("review", env)?,
            generation,
            &keep,
        ),
        Action::GiveBack {
            card,
            generation,
            paths,
            project,
            home,
        } => w::give_back(
            &project,
            &resolve_home(home)?,
            &card,
            &d::session_holder("give-back", env)?,
            generation,
            &paths,
        ),
        Action::LineLease {
            card,
            generation,
            paths,
            wait_secs,
            project,
            home,
            command,
        } => w::line_lease(
            &project,
            &resolve_home(home)?,
            &card,
            &d::session_holder("line-lease", env)?,
            generation,
            &paths,
            wait_secs,
            &command,
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
        Action::Finish {
            card,
            head,
            met,
            evidence_report,
            evidence_note,
            generation,
            by,
            project,
            home,
        } => {
            // No session id is not a refusal: the call passes the gate only on an ended lease.
            let session = d::session_holder("finish", env).ok();
            moai_cli::card::finish::finish(
                &project,
                &resolve_home(home)?,
                &card,
                &moai_cli::card::finish::FinishRequest {
                    session: session.as_deref(),
                    head: head.as_deref(),
                    met: &met,
                    evidence_report: evidence_report.as_deref(),
                    evidence_note: evidence_note.as_deref(),
                    generation,
                    by: &by,
                },
            )
        }
        Action::Land {
            card,
            head,
            met,
            generation,
            by,
            project,
            home,
        } => {
            let session = d::session_holder("land", env).ok();
            moai_cli::card::land::land(
                &project,
                &resolve_home(home)?,
                &card,
                &moai_cli::card::land::LandRequest {
                    session: session.as_deref(),
                    head: &head,
                    met: &met,
                    generation,
                    by: &by,
                },
            )
        }
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
