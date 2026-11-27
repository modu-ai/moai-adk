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
enum OrchestrateAction {
    /// Reads the ledger only. Print the plan of one round; nothing is sent, started or landed
    ///
    /// Reads the ledger only. Prints `window`, `lanes`, `actions` (`clear`, `dispatch`, `audit`,
    /// `blocked`, each with a stable `key`), `land_queue`, `superseded`
    /// (passed heads that are no longer their card's current head, with no `card land` line),
    /// `attention` and `host`. Two plans of
    /// an unchanged ledger and registry are equal. `--agents-json` reads a file in the shape of
    /// `claude agents --json`; without it the host is `not_read`. A project without a ledger has
    /// an empty plan (`ledger: "absent"`, exit 0) and nothing is created.
    Plan {
        /// The project folder (its profile holds the project id).
        #[arg(default_value = ".")]
        project: PathBuf,
        /// A file in the shape of `claude agents --json` (a JSON array); never run for you.
        #[arg(long)]
        agents_json: Option<PathBuf>,
        /// The HOME to read (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
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
        /// The SPEC the card names (`SPEC-<yyyymmdd>-<slug>`). The card and its link are one
        /// write, and a SPEC that already has a card that is not `done` refuses the add unless
        /// `--follow-up`, `--follow-up-of` or `--split-of` is given.
        #[arg(long, value_name = "SPEC_ID")]
        spec: Option<String>,
        /// With `--spec`: a follow-up card of the SPEC. A SPEC may have several open follow-up
        /// cards.
        #[arg(
            long = "follow-up",
            requires = "spec",
            conflicts_with_all = ["split_of", "follow_up_of"]
        )]
        follow_up: bool,
        /// With `--spec`: a follow-up card that names the card it follows; that card must name
        /// the same SPEC.
        #[arg(
            long = "follow-up-of",
            requires = "spec",
            value_name = "CARD",
            conflicts_with = "split_of"
        )]
        follow_up_of: Option<String>,
        /// With `--spec`: the card this one is split from; it must name the same SPEC.
        #[arg(long = "split-of", requires = "spec", value_name = "CARD")]
        split_of: Option<String>,
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
    /// A card with no `open` or `met` criterion is refused and the answer names `card criterion-add`:
    /// such a card could never be closed. The ledger must already exist.
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
    /// has (a list that leaves one out is a `conflict` naming it), and a pattern that another card's reservation also covers
    /// is added all the same and named in `overlaps` (no file lock). `--reason` is
    /// required for a held card; the `card.scope` event records `--by`, the reason, the
    /// generation and the holder; the card's own scope is not changed, so a later `card take`
    /// without `--scope` still reads the old one. `--generation` is the generation the caller
    /// read (`card show`): given, the card's live reservation must be at that generation, else
    /// `conflict` naming the generation it has (or that the card is not held) and nothing is
    /// written; left out, nothing is checked. Each entry is checked as `add --scope` checks it.
    /// An entry that the scope does not hold yet and that overlaps a `shared_files` entry of the
    /// project profile is refused: write that file in short with `card line-lease`.
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
    /// clock, its write scope and `ready` → `active`. The scope is the given `--scope`, else the
    /// card's own; a card with no scope of its own is filled from the `write:` list of the SPEC
    /// it names (every entry must be a path or a glob, a path named twice is merged), else from
    /// its last reservation, and the filled or given scope is then kept on the card (the answer's
    /// `scope_source` is `given`, `card`, `spec` or `last_reservation`). With no source the take
    /// is refused and names `card take <C-n> --scope <path>` and `card scope <C-n> --scope <path>`.
    /// A `--scope` value is one path: a comma in it is refused.
    /// A path that overlaps a `shared_files` entry of the project profile is never reserved
    /// whole: a given or own scope that holds one is refused, a scope filled from a SPEC or a
    /// last reservation leaves it out and names it in `warnings`; write the file in short with
    /// `card line-lease`.
    /// Refused when the card is not `ready`, is already held, has no `open` or `met` criterion
    /// (the answer names `card criterion-add`). A scope that overlaps another
    /// card's reservation is no reason to refuse: the answer's `overlaps` names it. The answer's `generation` is what `release` and
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
        /// A path pattern the holder may write, repeat for more (default: the card's own scope,
        /// else the SPEC's `write:` list, else its last reservation's scope).
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
    /// Writes the ledger. Withdraw a draft, ready, active or review card
    ///
    /// Writes the ledger. Marks the card withdrawn: it moves to `hold`, one `card.withdraw`
    /// event records the reason and the by with the writing session, `card take` refuses it
    /// naming the reason, and the default `card list` hides it (`--all` shows it). The ledger
    /// schema is unchanged: the six state values are pinned by its CHECK, so the marker is the
    /// free event kind, not a new value. A `draft`, `ready`, `active` or `review` card that
    /// nobody holds is withdrawn by anyone. A card whose `held` lease has ended is withdrawn only
    /// with `--by leader`, which ends that reservation as `released` and records who held it; a
    /// live lease, and a reservation that is not `held`, are refused.
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
    /// Writes the ledger. Fold a duplicate card into another
    ///
    /// Writes the ledger. The card moves to `hold`, one `card.fold` event records the card it
    /// folds into, the reason and the by with the writing session, and one `relates` link runs
    /// from the card to the target. Its notes, criteria and links stay readable, `card next`
    /// never offers it, `card take` refuses it naming the target, and the default `card list`
    /// hides it (`--all` shows it). The target must exist and must not be `hold`. The card must
    /// be one `card withdraw` takes, with the same rule for a lapsed lease: only `--by leader`
    /// folds a card whose `held` lease has ended, and that ends the reservation as `released`.
    /// `card ready` takes a folded card back.
    Fold {
        /// The card to fold (`C-12` or `#3066`).
        card: String,
        /// The card it folds into (`C-14`).
        #[arg(long)]
        into: String,
        /// Why the card is folded (one line, 1..=500 characters).
        #[arg(long)]
        reason: String,
        /// Who folds it (one line, 1..=200 characters).
        #[arg(long)]
        by: String,
        /// The project folder (its profile holds the project id).
        #[arg(default_value = ".")]
        project: PathBuf,
        /// The HOME to use (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Return the cards whose lease ended long ago to ready
    ///
    /// Writes the ledger. Every `active` card whose `held` lease ended at least the grace period
    /// ago (the profile key `lease_sweep_hours`, 24 hours without it) goes back to `ready`: its
    /// reservation ends as `released`, one `card.sweep` event and one note on the card name the
    /// old holder, generation, lease end, worktree and branch. It is not a reclaim: the card
    /// keeps its generation, uses none of its three reclaims, and its branch, worktree and notes
    /// are not touched. A card inside the grace period, a card in `review`, a reservation that
    /// is not `held` and a lease renewed after the listing are left as they are and listed as
    /// skipped. Only `--by leader` sweeps. Name one card to sweep only it; `--dry-run` lists
    /// and writes nothing.
    Sweep {
        /// One card to sweep (`C-12`); without it, every card with an ended lease past the grace.
        card: Option<String>,
        /// List the cards that would be swept and the ones left alone, and write nothing.
        #[arg(long)]
        dry_run: bool,
        /// Who sweeps; only `leader` is accepted.
        #[arg(long)]
        by: String,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Put a structured HOLD on a draft or ready card
    ///
    /// Writes the ledger. One `card.hold` event records the category (`money`, `legal`,
    /// `external_contract`, `secret`, `real_data` or `other`), the reason and the by with the
    /// writing session; the card keeps its state. `card next` then lists the card as skipped
    /// with the reason `hold`, `card take` refuses it before any reservation or worktree, and
    /// `card show` carries the HOLD. Only a person's `card unhold` lifts it: a note that says
    /// HOLD or lifts it in free text changes nothing. A card that is held, done, withdrawn or
    /// already on HOLD is refused and nothing is written.
    Hold {
        /// The card (`C-12` or `#3066`).
        card: String,
        /// `money`, `legal`, `external_contract`, `secret`, `real_data` or `other`.
        #[arg(long)]
        category: String,
        /// Why the card is on HOLD (one line, 1..=500 characters).
        #[arg(long)]
        reason: String,
        /// Who sets it (one line, 1..=200 characters).
        #[arg(long)]
        by: String,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME to use (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Lift the structured HOLD of a card
    ///
    /// Writes the ledger. One `card.unhold` event ends the `card.hold` in force; this command is
    /// the only writer of that event, so run it only when the decision behind the HOLD is made.
    /// A card with no HOLD in force is refused and nothing is written.
    Unhold {
        /// The card (`C-12` or `#3066`).
        card: String,
        /// Why the HOLD is lifted (one line, 1..=500 characters).
        #[arg(long)]
        reason: String,
        /// Who lifts it (one line, 1..=200 characters).
        #[arg(long)]
        by: String,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME to use (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Record that a card needs another card done first
    ///
    /// Writes the ledger. One active `blocks` link runs from the needed card to this one and one
    /// `card.prereq` event records the by. `card next` lists this card as skipped with the
    /// reason `prerequisite` naming the cards that are not done, and `card take` refuses it
    /// before any reservation. A second call for the same pair changes nothing (`created:
    /// false`). A card that needs itself, or one the needed card already needs, is refused.
    Prereq {
        /// The card that waits (`C-12` or `#3066`).
        card: String,
        /// The card that has to be done first (`C-14`).
        #[arg(long)]
        needs: String,
        /// Who records it (one line, 1..=200 characters).
        #[arg(long)]
        by: String,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME to use (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Reads the ledger only. List open cards whose text says HOLD but that have no structured HOLD
    ///
    /// Reads the ledger only. Each open card (draft, ready, active or review) with no `card.hold`
    /// in force whose body or a note has the capital word HOLD is listed with the note id (or
    /// the body) and a short quote of the line. The list is a hint: nothing is written and
    /// `card next` and `card take` ignore it. When a candidate is a real HOLD, a person runs
    /// `card hold` for it.
    HoldCandidates {
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME to read (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Record that this card's head landed on the local integration branch
    ///
    /// Writes the ledger. A `review` card gains one `card.landed` event naming the head and the
    /// by, with the writing session as the actor (NeXT-Flow, SPEC-20261120-card-done-mflow). The
    /// head must be 40 lowercase hex. A `ready` card (sent back after its audit) gains the event
    /// only with `--by leader`, when a verdict of the card names exactly the head, the newest
    /// verdict of every reviewer on it is a PASS and no reviewer held a reservation of the card;
    /// the event then carries `from_state` `ready`. Any other state is refused naming `ready`
    /// and `review`. `card done` then closes the card without a pull request
    /// when the profile's work flow is NeXT-Flow and the head sits inside the profile's
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
    /// Writes the ledger. Name the auditor of a card in review
    ///
    /// Writes the ledger. One `card.audit_assigned` event for a card in `review` with the
    /// auditor's identity and `--by leader`; only the leader word assigns. Refused (`conflict`)
    /// when the identity holds or ever held a reservation of the card, in any spelling (ASCII
    /// case and a `:<12hex>` worktree suffix are ignored), and when the card is not in review. A
    /// `ready` card takes the assignment only for the reviewer of a PASS verdict of the card
    /// (`conflict` for any other identity); no other state does.
    /// Refused too when the value is not a session identity, a lane name such as `agent-4`
    /// included: a session id (hex digits and dashes, 8 to 64 characters), optionally followed
    /// by `:` and 12 hex digits. Nothing is written then.
    Assign {
        /// The card, `C-12`.
        card: String,
        /// The auditor's identity (a session id, optionally `:<12hex>`), one word.
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
    ///
    /// With `--by leader --reason <why>` this is the leader's release of another session's
    /// reservation: only the confirmed leader session (the live registered `leader` entry in
    /// leader mode) may use it. It ends another session's `held` reservation as `released` when
    /// its lease ended or every early proof holds (the launcher exited, the worktree is clean or
    /// gone, no process works in it); a `review` card goes back through `active`, and the card
    /// returns to ready. Refusals: `not_leader`, `lease_live` (the first missing proof is named),
    /// `verdict_pass` (a card in `review` with a `pass` verdict is landed with `card land`
    /// instead) and `holder_changed`. No reclaim is counted: a card at the reclaim limit is taken
    /// again with `card take` after the release.
    Release {
        /// `C-12`.
        card: String,
        /// The generation `take` or `handoff` answered with.
        #[arg(long)]
        generation: i64,
        /// Who releases for the holder: `leader`, together with `--reason`.
        #[arg(long, requires = "reason")]
        by: Option<String>,
        /// Why the reservation of the other session is released (with `--by leader`).
        #[arg(long, requires = "by")]
        reason: Option<String>,
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
    /// session's own id, never typed. The new holder continues in the recorded worktree. Once the
    /// lease has ended, commits that sit on no remote-tracking ref do not stop the takeover: the
    /// event and the answer name them as `kept_commits` (the count and up to twenty ids, newest
    /// first). While the lease is still live, such commits stop it (`conflict`, nothing written),
    /// as does a reservation that is not `held`. An uncommitted change in that worktree stops
    /// either takeover (`conflict`, nothing written): commit or give it up first.
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
    /// Writes the ledger. Continue your own card after a /clear or a restart
    ///
    /// Writes the ledger. For the same lane after a `/clear`: the lane runs `session-register
    /// --continue` first, which keeps the old session id in its registry entry, then this command
    /// moves the card's reservation at `--generation` to this session at the next generation, with
    /// the same write scope, worktree and branch and a new lease from the ledger clock. The proofs
    /// are the registry's own record and are read before anything is written: this session's entry
    /// lists the holder session among its previous sessions (`not_same_lane`), no entry still names
    /// the holder session (`holder_registered`), and the reservation, the running folder and the
    /// entry name one git worktree (`other_worktree`); each is a `conflict` that names the code. A
    /// card that is `done`, a reservation that is not `held` and a generation that is not the
    /// current one are refused with the answer `card reclaim` gives for the same state, and a
    /// reservation that changed holder while the proofs were read is `holder_changed`. Nothing is
    /// written by a refusal. The card keeps its state (also `review`), notes, criteria and links;
    /// one `card.adopt` event names both sessions, both generations, the lane and the proof, and
    /// the reclaim count of the card is untouched. The worktree's changes, its processes and the
    /// launcher's lock are not read: the same lane in the same tree has the work. The old
    /// generation is stale from then on, so a session that is still running under the old id is
    /// refused by `card check`. The new holder is this session's own id, never typed.
    Adopt {
        /// `C-12`.
        card: String,
        /// The generation the card has now (`card show` prints it).
        #[arg(long)]
        generation: i64,
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
    /// Writes the ledger. Offer a taken card to one named session
    ///
    /// Writes the ledger. The holder offers the card to the session `--to`: one `card.offer` event
    /// names the holder, the generation, the target, the end of the offer's validity
    /// (`--valid-secs`, 60 to 86400, default 1800) and the optional `--handover` line. The
    /// reservation, the generation and the card's state stay as they are, so the holder owns the
    /// card until the target accepts. A new offer replaces the older one, and `--cancel` voids the
    /// newest offer (`no_offer` when there is none). Fenced like `handoff`: only the holder at the
    /// current generation with a live lease offers, and a refused holder call is recorded as one
    /// `fence_reject`. The reclaim count of the card is untouched.
    Offer {
        /// `C-12`.
        card: String,
        /// The generation this session holds.
        #[arg(long)]
        generation: i64,
        /// The session id of the session offered the card (hex digits and `-`, 8 to 64
        /// characters).
        #[arg(long, required_unless_present = "cancel", conflicts_with = "cancel")]
        to: Option<String>,
        /// Void the newest offer instead of making one.
        #[arg(long)]
        cancel: bool,
        /// Seconds the offer stays valid, 60 to 86400 (default 1800).
        #[arg(long, conflicts_with = "cancel")]
        valid_secs: Option<i64>,
        /// One line for the receiver: the work that is kept, the next step, open approvals and
        /// remote requests (1 to 500 characters, kept in the ledger for good).
        #[arg(long, conflicts_with = "cancel")]
        handover: Option<String>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Accept a card another session offered you
    ///
    /// Writes the ledger. Run by the session the newest offer of the card names, at the offered
    /// `--generation`: one transaction ends the old reservation as `released` and gives this
    /// session the next generation with the same scope, worktree and branch and a new lease, and
    /// writes one `card.accept` event. Refusals write nothing and each names its code in the
    /// answer: the state of the card in the words `card reclaim` uses, `no_offer` (no offer at
    /// that generation), `offer_cancelled`, `not_target` (another session was named),
    /// `offer_expired` and `lease_expired` (the holder's lease ended: the card goes the reclaim
    /// way). A session that is not named can never take a card this way. The new holder is this
    /// session's own id, never typed.
    Accept {
        /// `C-12`.
        card: String,
        /// The generation the offer was made at (the card's current one).
        #[arg(long)]
        generation: i64,
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
    /// `empty_scope`: a card without a write
    /// scope that has no usable SPEC `write:` list or earlier reservation to fill it from, so
    /// `take` refuses it; a card that `take` could fill is listed `ready` with the scope it would
    /// get and its `scope_source`; `no_criteria`: a `ready` card with no criterion that is not
    /// withdrawn, so `take` refuses it, given only when no other reason applies).
    /// `next` names the card to work on now: one this session holds,
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
    /// Reads the ledger only. Answer many cards, the live reservations and notes by number in one call
    ///
    /// Reads the ledger only. Pick at least one of `--card`, `--reservations` and `--note`; the
    /// answer is one JSON object with a `cards`, `reservations` and `notes` list for each asked
    /// for. `--card C-12,C-13` (at most 50, `#n` is refused) gives each card's `head` (the head of
    /// its newest hand-back note while that note belongs to its present work, else null), its
    /// newest `verdict` on any head, its `head_verdict` (the newest verdict on that head), the
    /// count of its criteria by status and its open criteria; `--notes N` (0 to 20, needs
    /// `--card`) adds the headers of its newest N notes with the first line cut to 200
    /// characters. A card with a verdict row that does not read has no verdict and counts it as
    /// `verdicts_unreadable`. `--reservations` lists every reservation that holds a card (held,
    /// expired candidate or held for review) by card, with its session, worktree, branch and
    /// whether its lease has ended. `--note 10436,10437` (at most 10) gives each note whole. A
    /// card or a note that does not exist refuses the call and nothing partial is printed. Text is
    /// shown as stored, as `card show` does. `card land` still decides what may land.
    Query {
        /// Cards, each `C-12`, comma separated.
        #[arg(
            long = "card",
            value_delimiter = ',',
            required_unless_present_any = ["reservations", "note"]
        )]
        card: Vec<String>,
        /// List every reservation that holds a card.
        #[arg(long)]
        reservations: bool,
        /// Note numbers to show whole, comma separated.
        #[arg(long = "note", value_delimiter = ',', value_name = "ID")]
        note: Vec<i64>,
        /// Also show the headers of the newest N notes of each card, 0 to 20 (needs `--card`).
        #[arg(long, requires = "card", allow_negative_numbers = true)]
        notes: Option<i64>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME to read (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Reads the ledger only. The coordination window's plan of one round
    ///
    /// Reads the ledger only. `moai card orchestrate plan` prints what one round of the
    /// coordination window would do and does none of it: the registry's lanes and whether each is
    /// free, busy or unknown; the cards to dispatch to the free lanes; the lanes to clear; the
    /// heads that need an audit; the audited heads that wait to land; and what waits for the user.
    /// It starts no host program and writes no file.
    Orchestrate {
        #[command(subcommand)]
        action: OrchestrateAction,
    },
    /// Reads the ledger only. Wait for the next note on the named cards, or for a timeout
    ///
    /// Reads the ledger only. Fixes a start mark (`--since`, else the largest note id in the
    /// ledger), then reads the ledger every `--every` seconds (default 15, 1 to 300). A poll
    /// keeps the local notes above the mark that are on a named card, or whose first line holds
    /// `all lanes` in any case on a card that is not `done`, and leaves out the notes of each
    /// `--not-author` name. The first poll that keeps any ends the wait with all of them, oldest
    /// first, each with its first line cut to 200 characters (`outcome: "note"`); when the
    /// timeout passes first the answer is `outcome: "timeout"` with the seconds waited. Both exit
    /// 0, and nothing is printed while it waits. `--timeout` must be given: 1 second to 12
    /// hours, written `<n>`, `<n>s`, `<n>m` or `<n>h`. An unknown card, an absent ledger or a
    /// ledger that is refused is refused before the wait, and nothing is created. A poll that
    /// finds the ledger busy counts as a poll that found nothing, and the timeout is counted from
    /// before the first read of the ledger. A wait started in the background with `&` from a
    /// shell that is not interactive ignores SIGINT (Ctrl-C); end it with SIGTERM
    /// (`kill <pid>`).
    Wait {
        /// A card to watch, `C-<n>` or `#<number>` as `card show` reads it (repeat, or join with
        /// commas).
        #[arg(long = "card", required = true, value_delimiter = ',')]
        cards: Vec<String>,
        /// The longest wait: `<n>`, `<n>s`, `<n>m` or `<n>h`, 1 second to 12 hours.
        #[arg(long, value_parser = moai_cli::card::wait::timeout_arg)]
        timeout: u64,
        /// The time between two polls, written as `--timeout` is: 1 to 300 seconds (default 15).
        #[arg(long, value_parser = moai_cli::card::wait::every_arg)]
        every: Option<u64>,
        /// Return only notes with a larger note id (default: the largest note id at the start).
        #[arg(long, value_parser = clap::value_parser!(i64).range(0..))]
        since: Option<i64>,
        /// Leave out the notes this author wrote, for example the lane's own name (repeat for
        /// more).
        #[arg(long = "not-author")]
        not_authors: Vec<String>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
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
    /// session that wrote the note, whatever `--by` says, and `card show` prints it. When stdin
    /// is a terminal the command refuses at once with a usage error and waits for nothing: give
    /// the text from a file or a heredoc. The text is
    /// at most 1 MiB, with no control character but tabs and line breaks, and no word with the
    /// shape of a secret: `card export` writes notes out. The answer's `ref` names the note.
    ///
    /// A note whose first line starts with `LANE-READY-INTEGRATION`, `REPAIRED` or `READY ` hands
    /// a head back and must read `<marker> <card> head=<40 lowercase hex>`. It is kept only when
    /// a worktree that is at that head holds `target/check-affected-<head12>/summary.txt` with
    /// the `result=PASS` line of a bare or `--card` `moai check affected` run of that head (a
    /// `--lane` run is no receipt). Otherwise the note is refused with the way out and nothing is written.
    /// A first line with a blank before the marker is refused the same way, and the `base` of the
    /// receipt's summary line must be an ancestor of that head.
    ///
    /// The same notes need exactly one `leftovers:` line in the body: `leftovers: none`, or
    /// `leftovers: <path> | <path>` for what the lane could not remove (a throwaway copy, a
    /// temporary HOME and a `mktemp` folder belong under `<worktree>/target/<name>/` and are
    /// removed by a relative path before READY). Without a readable line the note is refused and
    /// nothing is written. The answer of a kept note adds `shared_checkout_new`: the untracked or
    /// ignored files that appeared in the primary checkout since the card was reserved (nothing
    /// is refused for them yet), or `shared_checkout: "not_checked"` with the reason.
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
    /// Writes the ledger. Hand in a finished card: gates, receipt, READY note and review in one
    ///
    /// Writes the ledger, and only after every step before the write passed. As this session,
    /// holding the card at `--generation`: (1) the reserved worktree is clean, its HEAD is the
    /// tip of the reserved branch and no criterion is `open` (mark it with `card criterion`);
    /// (2) the fast gates all run, so one refusal lists every failure: `diff_check`, `fmt`,
    /// `ops_rustfmt`, `tables` and `trailers`; a failing gate starts no long step; (3) a PASS
    /// summary of the same head and base is quoted, otherwise `check affected --card` runs once;
    /// (4) the READY note is written from the receipt: first line `LANE-READY-INTEGRATION
    /// <card> head=<40hex>`, base, tier (the floor of the changed paths), receipt line, fast-gate
    /// exits, each criterion, the paths outside the reserved scope on one `outside-scope:` line
    /// (never a refusal), the `--mutants` lines and `leftovers: none`; (5) the card goes to
    /// `review` and its paths are given back, as `card review` does. The answer names the head,
    /// the tier and the next card. It writes no verdict and lands nothing. A failing step
    /// leaves the ledger as it was and names the step and its log.
    Submit {
        /// `C-12`.
        card: String,
        /// The generation `take` or `handoff` answered with.
        #[arg(long)]
        generation: i64,
        /// A file of `guard-mutant: <test> | <code removed> | exit <n>` lines; a tier 2 head
        /// needs it (`guard-mutant: none | <reason>` for no new guard test).
        #[arg(long)]
        mutants: Option<PathBuf>,
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
    /// shared file. Another card's reservation of a path is no reason to wait or refuse; the
    /// `--wait-secs` option is still accepted (at most 600) and is not used.
    /// The command's standard output goes to standard error, so the answer stays one
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
    /// the merge) — or, when the profile's work flow is NeXT-Flow, the card's `card landed`
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
    /// profile's integration branch (NeXT-Flow). When an open criterion of such a card
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
    /// integration branch (NeXT-Flow): one `git update-ref` that moves the branch to
    /// `--head` only while it still points at the tip that was read, then the same close as `card
    /// finish` with `--head` as the landed head. Only `--by leader` lands until the typed verdict
    /// exists. Everything `card finish` would refuse is refused before the branch moves; so is a
    /// candidate that is no commit, a SPEC that fails at the candidate, a candidate that does not
    /// have the tip on its first-parent line (`stale_base`) and a worktree that has the branch
    /// checked out and cannot be brought along: a tracked change, an operation in progress, a
    /// HEAD that is not the tip, a file no commit tracks where the head adds one, a sparse
    /// checkout, a skip-worktree or assume-unchanged entry, a submodule change or a second
    /// worktree on the branch. A worktree that has the branch checked out is brought to the head
    /// right after the move (`git read-tree -m -u`), and the answer's `worktree_follow` says
    /// `followed`, `finished` or `none`. A branch that moved in between is refused as
    /// `tip_moved`. A close or a follow refused after the move leaves the branch moved and says
    /// so; the same call run again skips the move, finishes the worktree and closes. A card
    /// already `done` with the same head answers done and writes nothing.
    ///
    /// The close needs an independent PASS of the exact head (SPEC-20261124-land-closes-card): when
    /// `card verdict` recorded one that the verdict check relies on, every open criterion not
    /// named by a `--met` becomes met with the line `verdict event <n> PASS head <12 hex>` and the
    /// card is done. Without such a PASS the call is refused as `no_pass` in every mode of
    /// `work_flow.land_verdict`. `--keep-open` lands a slice instead: the landing is recorded
    /// (`card.landed` with `keep_open` true), nothing is closed and the answer names the open
    /// criteria; the mode of `land_verdict` still judges the verdict, and `--keep-open` is refused
    /// with a `--met` option or on a card with no open criterion.
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
        /// Land a slice: record the landing and close nothing; the answer names the open criteria.
        #[arg(long)]
        keep_open: bool,
        /// The project folder (its profile holds the project id and the integration branch).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the integration branch and the ledger. Land 2 to 4 audited heads on one wide check
    ///
    /// `card land-batch C-1=<head> C-2=<head> ... --by leader` joins the heads one after another
    /// onto the integration tip (a head the tip already holds or stands on needs no join), runs
    /// the check once on the last join in a private worktree and, when it passes, lands each head
    /// through the same landing as `card land`, in the order given, each by its own compare-and-
    /// swap. A join candidate is landed with its audited head named, so the verdict of the audited
    /// head covers it. When the check fails on more than one head the first half runs on its own
    /// and the failing head is found by halving; a passing part lands and the rest is checked
    /// again on the new tip, because no result is reused once the tip has moved. A head that cannot
    /// be joined or fails alone is dropped and the others still land. A refusal of the landing
    /// (verdict, scans, a moved tip) stops the batch at that head and the answer lists the heads
    /// not run. The check is the wide `check affected` of this program on the join; a wide run
    /// that passed with a full-selection summary covers the scans of its heads. Only the session
    /// of the live `leader` entry of the lane registry may call it. A tip that moves after a
    /// passing check is joined and checked again, three times at most (`tip_kept_moving`). The
    /// answer has `landed`, `dropped` (with the reason), `not_run`, `runs` (the range, exit, log
    /// and scan cover of each check) and `stopped`.
    LandBatch {
        /// Each `C-n=<head>`: a card and its audited head (40 lowercase hex), 2 to 4 in all.
        #[arg(required = true)]
        pairs: Vec<String>,
        /// Who lands: only `leader` is accepted, from the registered leader session.
        #[arg(long)]
        by: String,
        /// The project folder (its profile holds the project id and the integration branch).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the integration branch and the ledger. Merge another branch into the integration branch
    ///
    /// `card land-sync <from> --spec <SPEC-ID> --by leader` merges the tip of the local branch
    /// `<from>` onto the integration tip as one merge commit (the tip is the first parent), runs
    /// the wide `check affected` of this program on that commit in a private worktree and, only
    /// after a full-selection PASS, moves the integration branch to it by one compare-and-swap
    /// through the same intent record `card land` writes, so the protection hook is not weakened.
    /// The branch must be listed in `work_flow.sync_sources` of the profile (an empty list refuses
    /// every branch) and `--spec` must name an approved SPEC present at the tip. Only the session of
    /// the live `leader` entry of the lane registry may call it. A tip that is not the one the merge
    /// was built on is refused as `stale_base`; a branch the tip already holds answers
    /// `moved: false`. Two card-less ledger rows (`card.sync_started`, `card.synced`) record the
    /// move so the next landing is not refused as drift. The answer has `moved`, `merge`,
    /// `to_tip`, `rev_list_count` (`git rev-list --count <branch>..<from>` after the move),
    /// `worktree_follow`, `runs` and `remote_copy`.
    LandSync {
        /// The local branch to merge in (listed in `work_flow.sync_sources`).
        from: String,
        /// The approved SPEC that decided this sync (written as the `Moai-Spec` trailer).
        #[arg(long)]
        spec: String,
        /// Who syncs: only `leader` is accepted, from the registered leader session.
        #[arg(long)]
        by: String,
        /// The project folder (its profile holds the project id and the integration branch).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the ledger. Record the audit verdict of one exact head as an event of the card
    ///
    /// Writes one `card.verdict` event of the card: the head and the base it was judged against
    /// (40 lowercase hex, commits of this repository, the base an ancestor of the head), `pass`
    /// or `fail`, the tier the reviewer judged (0 to 2), the evidence (one line that passes the
    /// secret screen) and, for a PASS above tier 0, the check receipt (`--receipt-head` and
    /// `--receipt-step NAME=EXIT` for each step). The reviewer is the session id the host sets;
    /// no option names one. The same call again writes nothing and answers the first event.
    /// `card land` reads these events and lands a head only on an independent PASS of it.
    Verdict {
        /// The card the audit belongs to (`C-12`).
        card: String,
        /// The audited head (40 lowercase hex).
        #[arg(long)]
        head: String,
        /// The base the head was judged against (40 lowercase hex, an ancestor of the head).
        #[arg(long)]
        base: String,
        /// `pass` or `fail`.
        #[arg(long)]
        verdict: String,
        /// The tier the reviewer judged, 0 to 2.
        #[arg(long, value_parser = clap::value_parser!(u8).range(0..=2))]
        tier: u8,
        /// What was read and run, on one line.
        #[arg(long)]
        evidence: String,
        /// The head the check receipt names.
        #[arg(long)]
        receipt_head: Option<String>,
        /// One step of the receipt, `NAME=EXIT`; repeat for each step.
        #[arg(long = "receipt-step", requires = "receipt_head")]
        receipt_step: Vec<String>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the git folder and the ledger only with `--restore`. Show where the integration branch stands against the ledger, and put it back
    ///
    /// The plain view writes nothing. `--restore` writes one file, the preview's one-use nonce in
    /// the common git folder, and `--apply` also moves the branch and writes the ledger row. Compares the integration branch's tip
    /// with the ledger: it must be the result commit of the newest successful `card land` of the
    /// branch (or, for a landing that was interrupted, the tip it started from or its target).
    /// `state` is `ok`, `drift` or `no_landing` (the ledger never moved the branch). A commit
    /// added by hand and a branch copied or rewound with `git branch -C` are drift. `--restore`
    /// previews the move back to the recorded commit: the commits it drops and a plan digest bound
    /// to the nonce; nothing moves. With `--apply` the move is made, and only with `--by leader`,
    /// the digest of the latest preview of the same state (a digest is used once) and
    /// `--permission-reference` naming the user's yes: the branch moves under the guard (a restore
    /// intent that names the recorded landing and its ledger row, the one move that may be no
    /// fast-forward) and one `card.tip_restored` row is written.
    Drift {
        /// Preview the move back to the recorded commit.
        #[arg(long)]
        restore: bool,
        /// Make the move of `--restore`; needs `--plan-digest`, `--permission-reference` and `--by`.
        #[arg(long)]
        apply: bool,
        /// The `plan_digest` of a preview of the same state.
        #[arg(long)]
        plan_digest: Option<String>,
        /// The reference of the user's yes (1 to 80 letters, digits or . _ : / -).
        #[arg(long)]
        permission_reference: Option<String>,
        /// Who restores the branch: only `leader` is accepted.
        #[arg(long)]
        by: Option<String>,
        /// The project folder (its profile holds the project id and the integration branch).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the git folder only with `--apply` or `--revert`. Show a stale guard record and re-point it
    ///
    /// The guard record is the file `moai/integration-branch` in the common git folder. After the
    /// profile's integration branch changes it still names the old branch, so the new branch is
    /// not guarded. The plain view writes nothing: it answers `state` (`stale`, `current`,
    /// `no_record`, `no_profile_branch` or `linked_worktree`) and, for `stale`, a plan digest and
    /// the exact `--apply` command. `--apply` needs `--plan-digest` and `--permission-reference`
    /// naming the user's yes: it rewrites the record under a lock, writes a receipt under
    /// `moai/record-repoint/` and answers the receipt, its stamp and the `--revert` command.
    /// `--revert <stamp>` restores the old bytes only while the record still equals the receipt's
    /// new bytes. The old branch is not guarded after the re-point; the preview says so.
    GuardRepoint {
        /// Make the re-point; needs `--plan-digest` and `--permission-reference`.
        #[arg(long, conflicts_with = "revert")]
        apply: bool,
        /// Put the old record back from the receipt of this stamp; needs `--permission-reference`.
        #[arg(long)]
        revert: Option<String>,
        /// The `plan_digest` of the preview of the same record.
        #[arg(long)]
        plan_digest: Option<String>,
        /// The reference of the user's yes (1 to 80 letters, digits or . _ : / -).
        #[arg(long)]
        permission_reference: Option<String>,
        /// The project folder (its profile holds the integration branch).
        #[arg(long, default_value = ".")]
        project: PathBuf,
    },
    /// Writes the git folder only. Guard the integration branch (run by git, never typed)
    ///
    /// Run by the `reference-transaction` hook, never by hand: git runs it with the state of a
    /// reference transaction (`prepared` or `committed`) and the transaction's lines on standard
    /// input. At `prepared` it refuses, with exit 1, every create, move or delete of the
    /// integration branch that the intent record of `card land` (`moai/land-intent` in the common
    /// git folder) does not name exactly, every case variant of the branch's name and every
    /// symbolic alias of it; `pack-refs` and `gc` pass. The branch is the one the record
    /// `moai/integration-branch` names. At `committed` it removes the intent record of a move that
    /// went through. Any other state does nothing.
    Guard {
        /// The state git gives the hook: `prepared` or `committed`.
        state: String,
    },
    /// Copies the integration branch to its remote, fast-forward only
    ///
    /// Writes no ledger row and, while the profile key is off, touches no network at all. The
    /// on-demand run of the copy a landing makes: the remote is read first and pushed only
    /// behind or absent, never forced. A copy that could not finish answers pending and is
    /// retried by this command or the next landing; diverged histories and a remote tip this
    /// repository does not know answer hold and move nothing.
    RemoteCopy {
        /// The project folder (its profile holds the project id and the integration branch).
        #[arg(long, default_value = ".")]
        project: PathBuf,
    },
    /// Writes the ledger. Extend the lease of a taken card
    ///
    /// Writes the ledger. The lease of this session's reservation at `--generation` ends
    /// `--lease-secs` from now, never earlier than it already did. The holder renews its own reservation whether or not the lease has
    /// ended, at the same generation, until another session reclaims it or a sweep returns the card.
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
    /// a symbolic link on disk is not followed. A path outside the scope that overlaps a
    /// `shared_files` entry of the project profile ends the refusal with the `card line-lease`
    /// form for that path.
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
            spec,
            follow_up,
            follow_up_of,
            split_of,
            by,
            project,
            home,
        } => {
            let request = c::AddRequest {
                title: &title,
                body: &body,
                kind: kind.as_deref(),
                priority: priority.as_deref(),
                scope: &scope,
                criteria: &criteria,
                by: by.as_deref(),
            };
            match spec.as_deref() {
                None => c::add(&project, &resolve_home(home)?, &request),
                Some(spec) => {
                    use moai_cli::card::spec_link::Relation;
                    let relation = match (follow_up, follow_up_of.as_deref(), split_of.as_deref()) {
                        (true, _, _) => Relation::FollowUp,
                        (false, Some(origin), _) => Relation::FollowUpOf(origin),
                        (false, None, Some(origin)) => Relation::SplitOf(origin),
                        (false, None, None) => Relation::Plain,
                    };
                    c::add_with_spec(&project, &resolve_home(home)?, &request, spec, relation)
                }
            }
        }
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
        } => w::withdraw(&project, &resolve_home(home)?, &card, &reason, &by),
        Action::Hold {
            card,
            category,
            reason,
            by,
            project,
            home,
        } => w::hold(
            &project,
            &resolve_home(home)?,
            &card,
            &category,
            &reason,
            &by,
        ),
        Action::Unhold {
            card,
            reason,
            by,
            project,
            home,
        } => w::unhold(&project, &resolve_home(home)?, &card, &reason, &by),
        Action::Prereq {
            card,
            needs,
            by,
            project,
            home,
        } => w::prereq(&project, &resolve_home(home)?, &card, &needs, &by),
        Action::HoldCandidates { project, home } => {
            moai_cli::card::hold_candidates::candidates(&project, &resolve_home(home)?)
        }
        Action::Fold {
            card,
            into,
            reason,
            by,
            project,
            home,
        } => w::fold(&project, &resolve_home(home)?, &card, &into, &reason, &by),
        Action::Sweep {
            card,
            dry_run,
            by,
            project,
            home,
        } => moai_cli::card::sweep::sweep(
            &project,
            &resolve_home(home)?,
            card.as_deref(),
            &by,
            dry_run,
        ),
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
                running_dir: std::env::current_dir().ok().as_deref(),
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
            by,
            reason,
            project,
            home,
        } => match (by, reason) {
            (Some(by), Some(reason)) => d::release_by_leader(
                &project,
                &resolve_home(home)?,
                &card,
                generation,
                d::session_holder("release", env).ok().as_deref(),
                &by,
                &reason,
            ),
            _ => d::release(
                &project,
                &resolve_home(home)?,
                &card,
                &d::session_holder("release", env)?,
                generation,
            ),
        },
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
                running_dir: std::env::current_dir().ok().as_deref(),
            },
        ),
        Action::Adopt {
            card,
            generation,
            lease_secs,
            project,
            home,
        } => {
            let running_dir = std::env::current_dir().map_err(|e| {
                format!("card adopt refused: the running folder cannot be read: {e}")
            })?;
            moai_cli::card::adopt::adopt(
                &project,
                &resolve_home(home)?,
                &card,
                &moai_cli::card::adopt::AdoptRequest {
                    to_holder: &d::session_holder("adopt", env)?,
                    generation,
                    lease_secs,
                    running_dir: &running_dir,
                },
            )
        }
        Action::Offer {
            card,
            generation,
            to,
            cancel,
            valid_secs,
            handover,
            project,
            home,
        } => moai_cli::card::offer::offer(
            &project,
            &resolve_home(home)?,
            &card,
            &moai_cli::card::offer::OfferRequest {
                holder: &d::session_holder("offer", env)?,
                generation,
                to: to.as_deref(),
                cancel,
                valid_secs,
                handover: handover.as_deref(),
            },
        ),
        Action::Accept {
            card,
            generation,
            lease_secs,
            project,
            home,
        } => moai_cli::card::offer::accept(
            &project,
            &resolve_home(home)?,
            &card,
            &moai_cli::card::offer::AcceptRequest {
                accepter: &d::session_holder("accept", env)?,
                generation,
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
        Action::Query {
            card,
            reservations,
            note,
            notes,
            project,
            home,
        } => moai_cli::card::query::query(
            &project,
            &resolve_home(home)?,
            &moai_cli::card::query::QueryRequest {
                cards: &card,
                reservations,
                note_ids: &note,
                notes,
            },
        ),
        Action::Orchestrate {
            action:
                OrchestrateAction::Plan {
                    project,
                    agents_json,
                    home,
                },
        } => moai_cli::card::orchestrate::plan(
            &project,
            &resolve_home(home)?,
            agents_json.as_deref(),
        ),
        Action::Wait {
            cards,
            timeout,
            every,
            since,
            not_authors,
            project,
            home,
        } => moai_cli::card::wait::wait(
            &project,
            &resolve_home(home)?,
            &moai_cli::card::wait::WaitRequest {
                cards: &cards,
                timeout_secs: timeout,
                every_secs: every.unwrap_or(moai_cli::card::wait::EVERY_DEFAULT_SECS),
                since,
                not_authors: &not_authors,
            },
        ),
        Action::Note {
            card,
            kind,
            reference,
            by,
            project,
            home,
        } => {
            // Standard input on a terminal: nobody types the text there, so refuse at once
            // instead of waiting for it (SPEC-20261126-card-note-tty-stdin).
            if std::io::IsTerminal::is_terminal(&std::io::stdin()) {
                return Err(
                    "card note refused: standard input is a terminal; give the note \
                            text from a file or a heredoc, for example `moai card note <card> \
                            --kind <kind> < note.md`"
                        .to_string(),
                );
            }
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
        Action::Submit {
            card,
            generation,
            mutants,
            project,
            home,
        } => {
            // The tests' seam for the long check program; the command line has no flag for it
            // and a release build does not read the variable at all.
            #[cfg(debug_assertions)]
            let submit_check = env("MOAI_TEST_SUBMIT_CHECK").map(PathBuf::from);
            #[cfg(not(debug_assertions))]
            let submit_check: Option<PathBuf> = None;
            moai_cli::card::submit::submit(
                &project,
                &resolve_home(home)?,
                &moai_cli::card::submit::SubmitRequest {
                    card: &card,
                    holder: &d::session_holder("submit", env)?,
                    generation,
                    mutants: mutants.as_deref(),
                    check_cmd: submit_check,
                },
            )
        }
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
            keep_open,
            project,
            home,
        } => {
            let session = d::session_holder("land", env).ok();
            let mut landed = moai_cli::card::land::land(
                &project,
                &resolve_home(home)?,
                &card,
                &moai_cli::card::land::LandRequest {
                    session: session.as_deref(),
                    head: &head,
                    met: &met,
                    generation,
                    by: &by,
                    keep_open,
                    audited: None,
                },
            )?;
            // The remote copy never fails the landing; its state rides in the answer.
            let copy = moai_cli::card::remote_copy::copy(&project, "land");
            if let serde_json::Value::Object(map) = &mut landed {
                map.insert("remote_copy".to_string(), copy.json());
            }
            Ok(landed)
        }
        Action::LandBatch {
            pairs,
            by,
            project,
            home,
        } => {
            let session = d::session_holder("land", env).ok();
            // The tests' seam for the check program; the command line has no flag for it and a
            // release build does not read the variable at all.
            #[cfg(debug_assertions)]
            let check_cmd = env("MOAI_TEST_LAND_BATCH_CHECK").map(PathBuf::from);
            #[cfg(not(debug_assertions))]
            let check_cmd: Option<PathBuf> = None;
            let mut answer = moai_cli::card::land::batch::land_batch(
                &project,
                &resolve_home(home)?,
                &moai_cli::card::land::batch::BatchRequest {
                    session: session.as_deref(),
                    by: &by,
                    pairs: &pairs,
                    check_cmd: check_cmd.as_deref(),
                },
            )?;
            // One copy for the whole batch, after the last landing.
            if answer["landed"].as_array().is_some_and(|l| !l.is_empty()) {
                let copy = moai_cli::card::remote_copy::copy(&project, "land");
                answer["remote_copy"] = copy.json();
            }
            Ok(answer)
        }
        Action::LandSync {
            from,
            spec,
            by,
            project,
            home,
        } => {
            let session = d::session_holder("land", env).ok();
            // The tests' seam for the check program; the command line has no flag for it and a
            // release build does not read the variable at all.
            #[cfg(debug_assertions)]
            let check_cmd = env("MOAI_TEST_LAND_SYNC_CHECK").map(PathBuf::from);
            #[cfg(not(debug_assertions))]
            let check_cmd: Option<PathBuf> = None;
            let mut answer = moai_cli::card::land::sync::land_sync(
                &project,
                &resolve_home(home)?,
                &moai_cli::card::land::sync::SyncRequest {
                    session: session.as_deref(),
                    by: &by,
                    from: &from,
                    spec: &spec,
                    check_cmd: check_cmd.as_deref(),
                },
            )?;
            // The remote copy never fails the sync; its state rides in the answer.
            if answer["moved"].as_bool() == Some(true) {
                let copy = moai_cli::card::remote_copy::copy(&project, "land");
                answer["remote_copy"] = copy.json();
            }
            Ok(answer)
        }
        Action::Verdict {
            card,
            head,
            base,
            verdict,
            tier,
            evidence,
            receipt_head,
            receipt_step,
            project,
            home,
        } => {
            let session = d::session_holder("verdict", env).ok();
            moai_cli::card::verdict::verdict(
                &project,
                &resolve_home(home)?,
                &card,
                &moai_cli::card::verdict::VerdictRequest {
                    session: session.as_deref(),
                    head: &head,
                    base: &base,
                    verdict: &verdict,
                    tier,
                    evidence: &evidence,
                    receipt_head: receipt_head.as_deref(),
                    steps: &receipt_step,
                },
            )
        }
        Action::RemoteCopy { project } => Ok(moai_cli::card::remote_copy::entry(&project)),
        Action::Drift {
            restore,
            apply,
            plan_digest,
            permission_reference,
            by,
            project,
            home,
        } => moai_cli::card::drift::drift(
            &project,
            &resolve_home(home)?,
            &moai_cli::card::drift::DriftRequest {
                restore,
                apply,
                plan_digest: plan_digest.as_deref(),
                permission_reference: permission_reference.as_deref(),
                by: by.as_deref(),
            },
        ),
        Action::GuardRepoint {
            apply,
            revert,
            plan_digest,
            permission_reference,
            project,
        } => moai_cli::card::guard::repoint::repoint(
            &project,
            &moai_cli::card::guard::repoint::RepointRequest {
                apply,
                revert: revert.as_deref(),
                plan_digest: plan_digest.as_deref(),
                permission_reference: permission_reference.as_deref(),
            },
        ),
        Action::Guard { state } => {
            let mut input = String::new();
            std::io::Read::read_to_string(&mut std::io::stdin(), &mut input)
                .map_err(|e| format!("card guard refused: cannot read the transaction: {e}"))?;
            moai_cli::card::guard::run(&state, &input)?;
            Ok(serde_json::json!({"action": "guard", "state": state, "allowed": true}))
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
