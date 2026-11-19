//! `bundle` (card C-119, SPEC-20261119-bundle-ready-check and SPEC-20261119-bundle-build): the
//! READY heads of the ledger, the checks of each head, the bundle pull request body and the
//! integration branch. One word, one action each; the logic is `moai_cli::bundle`.

use std::path::PathBuf;

/// Assemble the M-Flow bundle from READY heads
///
/// `moai bundle <action>` reads the READY heads, judges them, writes the bundle pull request
/// body and builds the branch; it helps the leader assemble one M-Flow bundle pull request from the
/// heads that lanes marked `LANE-READY-INTEGRATION`. Every action's help starts with what it
/// does: `Reads only.` (no ledger row, ref, branch, card note or remote is written) or, for
/// `body --out`, `Reads only. Writes one new file.` None of them fetches, pushes or opens a pull
/// request: fetch first, and the leader chooses the heads, pushes and opens the pull request.
/// `build` is the one action that writes: `Writes a new branch and worktree.` It creates
/// `m-flow/bundle-<date>-<n>` and a worktree under `<HOME>/.moai/worktrees/`, merges the heads one
/// merge commit each, drops a head that conflicts and names it; it never moves or deletes an
/// existing branch or worktree.
/// `check` and `body` run `git merge-tree --write-tree` (git 2.38 or newer), which leaves
/// unreferenced tree and blob objects in the object store that `git gc` removes; no ref moves.
#[derive(clap::Args)]
pub struct Args {
    #[command(subcommand)]
    action: Action,
}

#[derive(clap::Subcommand)]
enum Action {
    /// Reads only. List the READY heads: the newest readable line of each card
    ///
    /// Reads only. For each card whose newest note starts with `LANE-READY-INTEGRATION` and reads
    /// as `LANE-READY-INTEGRATION C-<n> head=<40 hex>` (then the end of the line or a space, with
    /// an optional ` spec=<SPEC id>` right after the head), prints the card, the note number, the
    /// head, the SPEC ids of the `Moai-Spec` trailers of `<tip>..<head>`, whether the head is
    /// present here and on `origin`, and the tip's SHA and commit time. A card whose newest marker
    /// note does not read is listed under `unreadable` and gives no head. Never fetches.
    Ready {
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
        /// The repository to read.
        #[arg(long, default_value = ".")]
        dir: PathBuf,
        /// The branch bundles are merged into.
        #[arg(long, default_value = "origin/v2026.11.0")]
        tip: String,
    },
    /// Reads only. Judge heads: every check, all findings in one answer
    ///
    /// Reads only. HEADS are `C-<n>` (the card's newest readable READY head) or `C-<n>=<40 hex>`.
    /// Every head gets every check: `no_ready_line`, `head_missing` (fetch first),
    /// `no_merge_base` (an unrelated history; the other heads are still judged), `not_on_origin`,
    /// `already_merged`, `audit_missing` (the changed paths `<tip>...<head>` set an
    /// audit floor of 1 or more and no note of the card is a `PASS` verdict naming exactly the
    /// head; the nearest `PASS` on an ancestor is shown), `trailers`, `spec`, `result_missing`
    /// and `conflict` (each head with the tip and with every earlier head, by the text merge
    /// GitHub runs). `ok` is false when anything was found. Never fetches.
    Check {
        /// `C-<n>` or `C-<n>=<40 hex>`, in the order they would be merged.
        #[arg(required = true)]
        heads: Vec<String>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
        /// The repository to read.
        #[arg(long, default_value = ".")]
        dir: PathBuf,
        /// The branch bundles are merged into.
        #[arg(long, default_value = "origin/v2026.11.0")]
        tip: String,
    },
    /// Reads only. Writes one new file with --out. Write the bundle pull request body
    ///
    /// Reads only. Writes one new file with --out. From the heads, the bundle branch and a run
    /// record file (JSON: `commands` with `check`, `command`, `exit` and optional `counts`;
    /// `not_run` with `check` and `reason`; `self_audit` with `plan`, `red`, `mutation`, `clone`,
    /// `host`, `receipt`; the check keys are build, lib_tests, home_isolation_scan, product_cli,
    /// fmt, clippy, trailer_check, spec_check, result_coverage, unittest) writes the body: the
    /// `audit-tier:` line, the `Moai-Spec:` lines, the heads table, the `Integration:` line with
    /// the branch's SHA, `## Self-Audit` and `## Result`. A fact with no record, or any finding of
    /// `check` for a head, refuses the whole body and names every one; the body must pass the
    /// Self-Audit checker.
    Body {
        /// `C-<n>` or `C-<n>=<40 hex>`: the heads the branch was built from.
        #[arg(required = true)]
        heads: Vec<String>,
        /// The bundle branch (or any revision) whose SHA is pinned.
        #[arg(long)]
        branch: String,
        /// The run record file (JSON).
        #[arg(long)]
        run: PathBuf,
        /// A new file to write the body to (an existing path is refused).
        #[arg(long)]
        out: Option<PathBuf>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
        /// The repository to read.
        #[arg(long, default_value = ".")]
        dir: PathBuf,
        /// The branch bundles are merged into.
        #[arg(long, default_value = "origin/v2026.11.0")]
        tip: String,
    },
    /// Writes a new branch and worktree. Merge heads, one merge commit each
    ///
    /// Writes a new branch and worktree. Creates `m-flow/bundle-<date>-<n>` (the lowest unused
    /// `n`) and the worktree `<HOME>/.moai/worktrees/<project>-bundle-<n>` at the tip, merges each
    /// head in the order given with one merge commit under the text merge GitHub runs, and prints
    /// the pinned SHA. HEADS are `C-<n>` (the card's newest readable READY head) or
    /// `C-<n>=<40 hex>`. A head whose merge conflicts is aborted, left out and named with its
    /// files, and `rebuild` is the command that builds again without it (a new branch: an
    /// existing branch or worktree is never moved or deleted). Heads already in the tip or in the
    /// branch are skipped and named. Never fetches, pushes or opens a pull request.
    Build {
        /// `C-<n>` or `C-<n>=<40 hex>`, in the order they are merged.
        #[arg(required = true)]
        heads: Vec<String>,
        /// The date part of the branch name, YYYYMMDD (default today, UTC).
        #[arg(long)]
        date: Option<String>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger and the worktrees folder (default `$HOME`).
        #[arg(long)]
        home: Option<PathBuf>,
        /// The repository to build in.
        #[arg(long, default_value = ".")]
        dir: PathBuf,
        /// The branch bundles are merged into.
        #[arg(long, default_value = "origin/v2026.11.0")]
        tip: String,
    },
}

pub const OPERATION: &str = "bundle";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    use moai_cli::bundle as b;
    use moai_cli::ledger::command::resolve_home;
    match args.action {
        Action::Ready {
            project,
            home,
            dir,
            tip,
        } => b::ready(&project, &resolve_home(home)?, &dir, &tip),
        Action::Check {
            heads,
            project,
            home,
            dir,
            tip,
        } => b::check(&project, &resolve_home(home)?, &dir, &tip, &heads),
        Action::Body {
            heads,
            branch,
            run,
            out,
            project,
            home,
            dir,
            tip,
        } => b::body(
            &project,
            &resolve_home(home)?,
            &dir,
            &tip,
            &heads,
            &branch,
            &run,
            out.as_deref(),
        ),
        Action::Build {
            heads,
            date,
            project,
            home,
            dir,
            tip,
        } => b::build::build(
            &project,
            &resolve_home(home)?,
            &dir,
            &tip,
            &heads,
            date.as_deref(),
        ),
    }
}
