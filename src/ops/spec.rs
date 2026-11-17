//! `spec` (SDD-2, SPEC-20261117-spec-tooling): the project's SPEC documents under
//! `<project>/.moai/specs/<SPEC-ID>/spec.md`. One word, one action each; the parser and the
//! checks are `moai_cli::spec_doc`. Only `take` runs a program (`moai_cli::spec_take`, SDD-7),
//! and only `take --push` and `take --draft-pr` reach the network.

use moai_cli::spec_doc::{self as sd, specs_dir, SpecDoc};
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};

/// Write, check, list, show and take the project's SPECs
///
/// `moai spec <action>` works on `<project>/.moai/specs/<SPEC-ID>/spec.md`, a SPEC id being
/// `SPEC-<yyyymmdd>-<slug>`. `new` writes one new SPEC folder and its `spec.md` and nothing
/// else; `check`, `list` and `show` read only; `take` creates or switches to the SPEC's branch.
/// Go-generation SPECs (`SPEC-AUTH-001`) are read, never written. Only `take --push` and
/// `take --draft-pr` reach the network.
#[derive(clap::Args)]
pub struct Args {
    #[command(subcommand)]
    action: Action,
}

#[derive(clap::Subcommand)]
enum Action {
    /// Writes one new SPEC. Create a draft SPEC from the built-in template
    ///
    /// Writes one new SPEC. Creates `.moai/specs/SPEC-<date>-<slug>/spec.md` with status `draft`
    /// and revision 1 from the template built into this binary (`--minimal`: Intent,
    /// Requirements, Acceptance, Scope and Plan; otherwise also Non-goals, Risks and Decisions).
    /// Refused when that folder exists. Writes nothing else and prints the file's path.
    New {
        /// Lowercase letters and digits in words joined by `-`, at most 48 characters.
        slug: String,
        /// One line, at most 200 characters, without `"` or `\`.
        #[arg(long)]
        title: String,
        /// Use the minimal template.
        #[arg(long)]
        minimal: bool,
        /// The SPEC this one splits from.
        #[arg(long)]
        parent: Option<String>,
        /// The creation date, YYYY-MM-DD (default: today in UTC).
        #[arg(long)]
        date: Option<String>,
        /// The project folder.
        #[arg(long, default_value = ".")]
        project: PathBuf,
    },
    /// Reads only. Check SPEC files and report problems by code
    ///
    /// Reads only. Checks one SPEC or every SPEC against the format and reports each problem with
    /// a stable code, an error or a warning. Exit 0 only when no SPEC has an error. Runs no
    /// program: a `verify:` command is read, never run.
    Check {
        /// The SPEC id.
        #[arg(conflicts_with = "all", required_unless_present = "all")]
        id: Option<String>,
        /// Check every SPEC folder.
        #[arg(long)]
        all: bool,
        /// Also require status `completed` and every AC `met` or `withdrawn` in `## Result`.
        #[arg(long)]
        closure: bool,
        /// The project folder.
        #[arg(long, default_value = ".")]
        project: PathBuf,
    },
    /// Reads only. List the project's SPECs with status, title and parent
    ///
    /// Reads only. One row per SPEC folder, in name order: id, status, title, parent and the
    /// number of errors and warnings. A project without SPECs answers an empty list.
    List {
        /// The project folder.
        #[arg(long, default_value = ".")]
        project: PathBuf,
    },
    /// Reads only. Show one SPEC as JSON
    ///
    /// Reads only. The frontmatter, the sections, every REQ and AC, the scope, the content
    /// digest an approval records, and the findings. `local_cards` lists the cards of this
    /// machine's ledger that name the SPEC (`moai card spec`); they are personal and never
    /// written to the SPEC. Without a project profile or a ledger the list is empty and
    /// `ledger` says why; the SPEC is shown all the same.
    Show {
        /// The SPEC id.
        id: String,
        /// The project folder.
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME whose ledger is read (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
    /// Writes the branch. Start work on a SPEC on its own branch
    ///
    /// Writes the branch. Checks that the SPEC exists, has the form `SPEC-<yyyymmdd>-<slug>` and
    /// no error, then creates `spec/<yyyymmdd>-<slug>` from the base and switches to it, or
    /// switches to it when it exists. Refused, with nothing changed, while the work tree has a
    /// change outside the SPEC's own folder or when the branch carries another SPEC's
    /// `Moai-Spec` trailer. `--card` links a card in the local ledger (`moai card spec`). Prints
    /// the next commands (trailers, push, draft pull request) and runs none of them unless
    /// `--push` (and `--draft-pr`, with the remote `origin` only) is given. git runs with the
    /// user's own configuration, so the repository's hooks (`post-checkout`, `pre-push`) and
    /// `core.fsmonitor` run even without `--push`; git never asks for a password, but ssh may
    /// still ask for a key passphrase on the terminal.
    Take {
        /// The SPEC id.
        id: String,
        /// The branch or commit to start from and to open the pull request against (default:
        /// the base recorded when the branch was created, else the current branch).
        #[arg(long)]
        base: Option<String>,
        /// Link this card (C-number) to the SPEC in the local ledger.
        #[arg(long)]
        card: Option<String>,
        /// The remote for the push.
        #[arg(long, default_value = "origin")]
        remote: String,
        /// Push the branch to the remote (an external write).
        #[arg(long)]
        push: bool,
        /// Also open a draft pull request with `gh` (needs `--push` and the remote `origin`).
        #[arg(long)]
        draft_pr: bool,
        /// The project folder.
        #[arg(long, default_value = ".")]
        project: PathBuf,
        /// The HOME whose ledger `--card` writes (default `$HOME`). Tests pass a temporary folder.
        #[arg(long)]
        home: Option<PathBuf>,
    },
}

pub const OPERATION: &str = "spec";

pub fn run(args: Args) -> Result<Value, String> {
    match args.action {
        Action::New {
            slug,
            title,
            minimal,
            parent,
            date,
            project,
        } => new(
            &project,
            &slug,
            &title,
            minimal,
            parent.as_deref(),
            date.as_deref(),
        ),
        Action::Check {
            id,
            all: _,
            closure,
            project,
        } => check(&project, id.as_deref(), closure),
        Action::List { project } => list(&project),
        Action::Show { id, project, home } => show(&project, &id, home),
        Action::Take {
            id,
            base,
            card,
            remote,
            push,
            draft_pr,
            project,
            home,
        } => {
            let gh = if draft_pr {
                moai_cli::forge::transport::GhTransport::from_env()
                    .map_err(|e| format!("spec take refused: {e}"))?
                    .bin
            } else {
                "gh".to_string()
            };
            moai_cli::spec_take::take(&moai_cli::spec_take::Request {
                project: &project,
                id: &id,
                base: base.as_deref(),
                remote: &remote,
                card: card.as_deref(),
                home: home.as_deref(),
                push,
                draft_pr,
                gh: &gh,
            })
        }
    }
}

const SPECS: &str = sd::SPECS_DIR;

/// A folder name as it may appear in a message: anything but ASCII letters, digits, `-`, `_`
/// and `.` reads `?`.
fn shown(name: &str) -> String {
    name.chars()
        .take(80)
        .map(|c| {
            if c.is_ascii_alphanumeric() || matches!(c, '-' | '_' | '.') {
                c
            } else {
                '?'
            }
        })
        .collect()
}

/// A typed SPEC id, refused unless it has a SPEC id's form (so it is safe to echo).
fn typed_id(action: &str, id: &str) -> Result<(), String> {
    if sd::id_form(id).is_none() {
        return Err(format!(
            "spec {action} refused: the id is not SPEC-<yyyymmdd>-<slug> (or a Go-generation \
             SPEC-…-NNN id)"
        ));
    }
    Ok(())
}

fn new(
    project: &Path,
    slug: &str,
    title: &str,
    minimal: bool,
    parent: Option<&str>,
    date: Option<&str>,
) -> Result<Value, String> {
    let refused = |m: String| format!("spec new refused: {m}");
    let today;
    let created = match date {
        Some(d) => d,
        None => {
            let now = std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .map_err(|_| "system clock is before the Unix epoch".to_string())?
                .as_secs();
            today = moai_cli::hygiene::utc_date(now);
            &today
        }
    };
    let id = sd::new_id(created, slug).map_err(refused)?;
    let text = sd::render(&sd::NewSpec {
        id: &id,
        title,
        created,
        parent,
        minimal,
    })
    .map_err(refused)?;
    let (root, _) = specs_dir(project)?;
    let generation = moai_cli::state_root::classify(&root).generation;
    if matches!(generation, "interrupted" | "refuse") {
        return Err(refused(format!(
            "the project's .moai folder reads `{generation}`; resolve that first (go-state-plan)"
        )));
    }
    let moai = root.join(moai_cli::state_root::MOAI_DIR);
    let specs = moai.join("specs");
    let dir = specs.join(&id);
    let rel = format!("{SPECS}/{id}/{}", sd::SPEC_FILE);
    if fs::symlink_metadata(&dir).is_ok() {
        return Err(format!(
            "conflict in {SPECS}/{id}: the SPEC folder exists; nothing was written"
        ));
    }
    // Only the folders made here are taken back on a failure.
    let mut made: Vec<PathBuf> = Vec::new();
    let result = (|| {
        for d in [&moai, &specs] {
            if fs::symlink_metadata(d).is_err() {
                fs::create_dir(d).map_err(|_| "spec write failed: a folder create failed")?;
                made.push(d.clone());
            }
        }
        fs::create_dir(&dir).map_err(|e| {
            if e.kind() == std::io::ErrorKind::AlreadyExists {
                format!("conflict in {SPECS}/{id}: the SPEC folder exists; nothing was written")
            } else {
                "spec write failed: the SPEC folder create failed".to_string()
            }
        })?;
        made.push(dir.clone());
        // N7: a folder on the way that turned into a link meanwhile is refused before the
        // file is written (the folders made here are taken back).
        for d in [&moai, &specs, &dir] {
            if !fs::symlink_metadata(d).is_ok_and(|m| m.is_dir()) {
                return Err(
                    "spec refused: a SPEC folder on the way is a link or not a folder; it is \
                     never followed"
                        .to_string(),
                );
            }
        }
        moai_cli::apply::publish_exclusive(&dir.join(sd::SPEC_FILE), text.as_bytes())
    })();
    if let Err(e) = result {
        for d in made.iter().rev() {
            let _ = fs::remove_dir(d);
        }
        return Err(e);
    }
    Ok(json!({
        "action": "new",
        "id": id,
        "path": rel,
        "status": "draft",
        "revision": 1,
        "form": if minimal { "minimal" } else { "full" },
        "created": created,
        "writes": true,
    }))
}

/// Every entry of `.moai/specs` that is a folder or a link (a link is reported, not entered),
/// in name order. Plain files there are not SPECs and are skipped.
fn folders(root: &Path) -> Result<Vec<(String, PathBuf)>, String> {
    let specs = root.join(SPECS);
    let mut out = Vec::new();
    let entries = fs::read_dir(&specs)
        .map_err(|_| "spec read failed: the SPEC folder could not be listed")?;
    for e in entries {
        let e = e.map_err(|_| "spec read failed: the SPEC folder could not be listed")?;
        let path = e.path();
        let kind = fs::symlink_metadata(&path)
            .map_err(|_| "spec read failed: a SPEC folder could not be examined")?;
        if kind.is_dir() || kind.file_type().is_symlink() {
            out.push((e.file_name().to_string_lossy().into_owned(), path));
        }
    }
    out.sort();
    Ok(out)
}

fn report(name: &str, doc: &SpecDoc) -> Value {
    json!({
        "id": doc.id.clone().unwrap_or_else(|| name.to_string()),
        "path": format!("{SPECS}/{name}/{}", sd::SPEC_FILE),
        "errors": doc.errors(),
        "warnings": doc.warnings(),
        "findings": doc.findings,
    })
}

fn check(project: &Path, id: Option<&str>, closure: bool) -> Result<Value, String> {
    if let Some(id) = id {
        typed_id("check", id)?;
    }
    let (root, present) = specs_dir(project)?;
    let targets: Vec<(String, PathBuf)> = match id {
        Some(id) => {
            let dir = root.join(SPECS).join(id);
            if !present || fs::symlink_metadata(&dir).is_err() {
                return Err(format!("spec not found: {id}"));
            }
            vec![(id.to_string(), dir)]
        }
        None if present => folders(&root)?,
        None => Vec::new(),
    };
    let mut files = Vec::new();
    let (mut errors, mut warnings) = (0, 0);
    let mut listed = Vec::new();
    for (name, dir) in &targets {
        let doc = sd::load(dir, closure);
        errors += doc.errors();
        warnings += doc.warnings();
        for f in doc
            .findings
            .iter()
            .filter(|f| f.severity == sd::Severity::Error)
        {
            let line = f.line.map(|l| format!(":{l}")).unwrap_or_default();
            listed.push(format!("{SPECS}/{}/spec.md{line} {}", shown(name), f.code));
        }
        files.push(report(name, &doc));
    }
    if errors > 0 {
        const SHOWN: usize = 50;
        let more = listed.len().saturating_sub(SHOWN);
        let mut text = listed
            .into_iter()
            .take(SHOWN)
            .collect::<Vec<_>>()
            .join("; ");
        if more > 0 {
            text.push_str(&format!("; and {more} more (spec show <id> lists them)"));
        }
        return Err(format!(
            "spec check refused: {errors} error(s) and {warnings} warning(s) in {} SPEC(s): {text}",
            targets.len()
        ));
    }
    Ok(json!({
        "action": "check",
        "closure": closure,
        "specs": if present { "present" } else { "absent" },
        "count": files.len(),
        "errors": 0,
        "warnings": warnings,
        "files": files,
    }))
}

fn list(project: &Path) -> Result<Value, String> {
    let (root, present) = specs_dir(project)?;
    let mut rows = Vec::new();
    if present {
        for (name, dir) in folders(&root)? {
            let doc = sd::load(&dir, false);
            rows.push(json!({
                "id": doc.id.clone().unwrap_or_else(|| name.clone()),
                "folder": name,
                "status": doc.status,
                "title": doc.title,
                "parent": doc.parent,
                "legacy": doc.id_form == Some(sd::IdForm::Legacy) || doc.status_legacy,
                "errors": doc.errors(),
                "warnings": doc.warnings(),
            }));
        }
    }
    Ok(json!({
        "action": "list",
        "specs": if present { "present" } else { "absent" },
        "count": rows.len(),
        "rows": rows,
    }))
}

fn show(project: &Path, id: &str, home: Option<PathBuf>) -> Result<Value, String> {
    typed_id("show", id)?;
    let doc = sd::find(project, id)?;
    Ok(json!({
        "action": "show",
        "path": format!("{SPECS}/{id}/{}", sd::SPEC_FILE),
        "spec": doc,
        "local_cards": local_cards(project, id, home),
    }))
}

/// The cards of this machine's ledger that name SPEC `id` now (SDD-4): personal, so `local`.
/// No HOME, no project profile, no ledger or a ledger that cannot be read lists no card and
/// says why in `ledger`; it never stops `show`.
fn local_cards(project: &Path, id: &str, home: Option<PathBuf>) -> Value {
    use moai_cli::ledger::{self, command, LedgerError};
    let none = |why: &str| json!({"local": true, "ledger": why, "count": 0, "cards": []});
    let Ok(home) = command::resolve_home(home) else {
        return none("no_home");
    };
    let Ok(project_id) = command::project_id_of(project) else {
        return none("no_project");
    };
    let read = match ledger::open_read_only(&home, &project_id) {
        Ok(l) => moai_cli::card::spec_link::cards_of(l.conn(), id).map_err(LedgerError::from),
        Err(e) => Err(e),
    };
    match read {
        Ok(cards) => {
            json!({"local": true, "ledger": "ready", "count": cards.len(), "cards": cards})
        }
        Err(LedgerError::Absent) => none("absent"),
        Err(e) => none(e.code()),
    }
}
