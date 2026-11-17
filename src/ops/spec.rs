//! `spec` (SDD-2, SPEC-20261117-spec-tooling): the project's SPEC documents under
//! `<project>/.moai/specs/<SPEC-ID>/spec.md`. One word, one action each; the parser and the
//! checks are `moai_cli::spec_doc`. No action runs a program or reaches the network.

use moai_cli::spec_doc::{self as sd, SpecDoc};
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};

/// Write, check, list and show the project's SPECs
///
/// `moai spec <action>` works on `<project>/.moai/specs/<SPEC-ID>/spec.md`, a SPEC id being
/// `SPEC-<yyyymmdd>-<slug>`. `new` writes one new SPEC folder and its `spec.md` and nothing
/// else; `check`, `list` and `show` read only. Go-generation SPECs (`SPEC-AUTH-001`) are read,
/// never written. No action runs a program or reaches the network.
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
    /// digest an approval records, and the findings.
    Show {
        /// The SPEC id.
        id: String,
        /// The project folder.
        #[arg(long, default_value = ".")]
        project: PathBuf,
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
        Action::Show { id, project } => show(&project, &id),
    }
}

const SPECS: &str = ".moai/specs";

/// The project root and whether `.moai/specs` is there. `.moai` or `.moai/specs` that is a link
/// or not a folder is refused; nothing is followed.
fn specs_dir(project: &Path) -> Result<(PathBuf, bool), String> {
    let root = moai_cli::apply::root_of(project)?;
    let mut cur = root.clone();
    for part in [moai_cli::state_root::MOAI_DIR, "specs"] {
        cur = cur.join(part);
        match fs::symlink_metadata(&cur) {
            Ok(m) if m.is_dir() => {}
            Ok(_) => {
                return Err(format!(
                    "spec refused: {} is a link or not a folder; it is never followed",
                    rel_of(&root, &cur)
                ))
            }
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok((root, false)),
            Err(_) => return Err("spec read failed: the SPEC folder could not be examined".into()),
        }
    }
    Ok((root, true))
}

fn rel_of(root: &Path, p: &Path) -> String {
    p.strip_prefix(root)
        .map(|r| r.to_string_lossy().replace('\\', "/"))
        .unwrap_or_default()
}

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

fn show(project: &Path, id: &str) -> Result<Value, String> {
    typed_id("show", id)?;
    let (root, present) = specs_dir(project)?;
    let dir = root.join(SPECS).join(id);
    if !present || fs::symlink_metadata(&dir).is_err() {
        return Err(format!("spec not found: {id}"));
    }
    let doc = sd::load(&dir, false);
    Ok(json!({
        "action": "show",
        "path": format!("{SPECS}/{id}/{}", sd::SPEC_FILE),
        "spec": doc,
    }))
}
