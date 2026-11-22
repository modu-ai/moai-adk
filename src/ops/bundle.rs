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
/// `m-flow/bundle-<date>-<n>` and a worktree under `<project>/.claude/worktrees/`, merges the heads one
/// merge commit each, drops a head that conflicts and names it; it never moves or deletes an
/// existing branch or worktree.
/// `verify` is the one action that runs commands: `Runs commands.` It runs the check list in the
/// worktree of a bundle branch under one heavy slot and writes one new folder under the
/// worktree's `target/` (or `--out`) with `run.json`, the run record `body` reads.
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
        /// The branch bundles are merged into; without it the profile's m-flow
        /// integration branch becomes origin/<branch>, else origin/v2026.11.0.
        #[arg(long)]
        tip: Option<String>,
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
        /// The branch bundles are merged into; without it the profile's m-flow
        /// integration branch becomes origin/<branch>, else origin/v2026.11.0.
        #[arg(long)]
        tip: Option<String>,
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
        /// The branch bundles are merged into; without it the profile's m-flow
        /// integration branch becomes origin/<branch>, else origin/v2026.11.0.
        #[arg(long)]
        tip: Option<String>,
    },
    /// Writes a new branch and worktree. Merge heads, one merge commit each
    ///
    /// Writes a new branch and worktree. Creates `m-flow/bundle-<date>-<n>` (the lowest unused
    /// `n`) and the worktree `<project>/.claude/worktrees/<project>-bundle-<n>` at the tip, merges each
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
        /// The branch bundles are merged into; without it the profile's m-flow
        /// integration branch becomes origin/<branch>, else origin/v2026.11.0.
        #[arg(long)]
        tip: Option<String>,
    },
    /// Runs commands. Run the check list on a bundle branch
    ///
    /// Runs commands. Writes one new folder under the worktree's `target/` (or `--out`). Runs
    /// every command of the check list (build, the whole `--lib`, the home-isolation and
    /// product CLI tests, the tests the branch adds or changes and the `--test` targets of the
    /// SPECs it changes, fmt, clippy, `trailer-check`, `spec check --all`, `spec check <SPEC>
    /// --closure` per SPEC, and the Python unittest discovery) in the worktree that has BRANCH
    /// checked out, under one heavy slot (`heavy-slot-acquire`), and goes on after a failed
    /// command. Writes `run.json` (`commands`, `not_run`, `self_audit`: the record `bundle body
    /// --run` reads) and one log per command in the folder; a command that cannot run is in
    /// `not_run` with its reason, a failure names the heads whose diff touches a file its output
    /// names, or `unattributed`. HEADS are `C-<n>` or `C-<n>=<40 hex>` as for `build`. It
    /// refuses, running and writing nothing, when the branch has no worktree, the worktree has
    /// uncommitted changes to tracked files, a head is in neither the tip nor the branch, the
    /// folder exists, or `--self-audit` is not a JSON object. Never fetches, pushes or opens a
    /// pull request.
    Verify {
        /// `C-<n>` or `C-<n>=<40 hex>`: the heads the branch was built from.
        #[arg(required = true)]
        heads: Vec<String>,
        /// The bundle branch `build` made.
        #[arg(long)]
        branch: String,
        /// A new folder for `run.json` and the logs (default `<worktree>/target/bundle-verify/<sha>`).
        #[arg(long)]
        out: Option<PathBuf>,
        /// A JSON object with the six Self-Audit lines (`plan`, `red`, `mutation`, `clone`,
        /// `host`, `receipt`), copied into `run.json`.
        #[arg(long)]
        self_audit: Option<PathBuf>,
        /// The built `moai-cli` for the trailer, spec and closure checks (default
        /// `<CARGO_TARGET_DIR or the worktree's target>/debug/moai-cli`).
        #[arg(long)]
        moai_cli: Option<PathBuf>,
        /// The Python for the unittest discovery.
        #[arg(long, default_value = "python3")]
        python: String,
        /// The heavy-slot folder (default the machine's own).
        #[arg(long)]
        slot_dir: Option<PathBuf>,
        /// Heavy slots: only a value below the machine's limit takes effect.
        #[arg(long)]
        max: Option<u32>,
        /// Free disk in GB the slot needs: only a value above the machine's limit takes effect.
        #[arg(long)]
        min_free_gb: Option<u32>,
        /// The project folder (its profile holds the project id).
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
        /// The repository to read.
        #[arg(long, default_value = ".")]
        dir: PathBuf,
        /// The branch bundles are merged into; without it the profile's m-flow
        /// integration branch becomes origin/<branch>, else origin/v2026.11.0.
        #[arg(long)]
        tip: Option<String>,
    },
    /// Writes the user HOME. Install an opt-in pack's entries into the user folders and record
    /// the selection in the user-assets manifest (the Go output contract rides in `result.text`)
    Add {
        /// The pack name (the valid names: `bundle list`).
        name: String,
    },
    /// Writes the user HOME. Remove a pack's entries; entries shared with the base set or
    /// another selection survive, and a declared dependency of a preserved asset is deferred
    /// (the Go output contract rides in `result.text`)
    Remove {
        /// The pack name (the valid names: `bundle list`).
        name: String,
    },
    /// Reads only. List the opt-in packs with their skill and agent counts (the Go output
    /// contract rides in `result.text`)
    List,
}

pub const OPERATION: &str = "bundle";

const DEFAULT_TIP: &str = "origin/v2026.11.0";

/// The tip an action judges: an explicit `--tip` wins; without one, the project
/// profile's m-flow integration branch becomes `origin/<branch>` (the stored
/// value drives the integration ref), and without a recorded branch today's
/// default stays, so every other repository keeps its answer.
fn resolve_tip(project: &std::path::Path, tip: Option<&str>) -> Result<String, String> {
    if let Some(t) = tip {
        return Ok(t.to_string());
    }
    let work_flow = moai_cli::profile::read_profile(project)
        .ok()
        .and_then(|(p, _)| p.work_flow);
    if work_flow.as_ref().is_some_and(|w| w.flow == "m-flow") {
        if let Some(branch) = work_flow
            .and_then(|w| w.integration_branch)
            .filter(|b| !b.trim().is_empty())
        {
            return Ok(format!("origin/{branch}"));
        }
    }
    Ok(DEFAULT_TIP.to_string())
}

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    use moai_cli::bundle as b;
    use moai_cli::ledger::command::resolve_home;
    let tip = match &args.action {
        Action::Ready { project, tip, .. }
        | Action::Check { project, tip, .. }
        | Action::Body { project, tip, .. }
        | Action::Build { project, tip, .. }
        | Action::Verify { project, tip, .. } => resolve_tip(project, tip.as_deref())?,
        // The user-assets actions name no tip: they never judge a ref.
        Action::Add { .. } | Action::Remove { .. } | Action::List => String::new(),
    };
    match args.action {
        Action::Ready {
            project, home, dir, ..
        } => b::ready(&project, &resolve_home(home)?, &dir, &tip),
        Action::Check {
            heads,
            project,
            home,
            dir,
            ..
        } => b::check(&project, &resolve_home(home)?, &dir, &tip, &heads),
        Action::Body {
            heads,
            branch,
            run,
            out,
            project,
            home,
            dir,
            ..
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
            ..
        } => b::build::build(
            &project,
            &resolve_home(home)?,
            &dir,
            &tip,
            &heads,
            date.as_deref(),
        ),
        Action::Verify {
            heads,
            branch,
            out,
            self_audit,
            moai_cli,
            python,
            slot_dir,
            max,
            min_free_gb,
            project,
            home,
            dir,
            ..
        } => b::verify::verify(
            &project,
            &resolve_home(home)?,
            &dir,
            &tip,
            &heads,
            &b::verify::Options {
                branch: &branch,
                out: out.as_deref(),
                self_audit: self_audit.as_deref(),
                moai_cli: moai_cli.as_deref(),
                python: &python,
                slot_dir: slot_dir.as_deref(),
                max,
                min_free_gb,
            },
        ),
        Action::Add { name } => moai_cli::user_assets::add(&resolve_home(None)?, &name),
        Action::Remove { name } => moai_cli::user_assets::remove(&resolve_home(None)?, &name),
        Action::List => moai_cli::user_assets::list(),
    }
}
