//! `secret-deny-plan` (SPEC-20261122-secret-file-deny-rules): the `Read` deny rules a project's
//! root `.claude/settings.json` could carry, built from the key files found by name and from the
//! entries of `.claudeignore`, which Claude Code does not read. It looks at names, reads
//! `.claudeignore` and git's list of tracked files, opens no matched file and writes nothing.

use crate::worktree_include::glob_match;
use serde_json::{json, Value};
use std::collections::{HashSet, VecDeque};
use std::fs;
use std::path::Path;
use std::process::{Command, Stdio};

/// The credential file shapes a key rule is proposed for (REQ-2; the credential table of
/// https://code.claude.com/docs/en/agent-sdk/secure-deployment, read 2026-11-26).
pub const KEY_SHAPES: [&str; 7] = [
    ".env",
    ".env.*",
    "*.pem",
    "*.key",
    "*-service-account.json",
    ".npmrc",
    ".pypirc",
];

/// The walk stops after this many entries (D-8).
pub const WALK_BOUND: usize = 100_000;

/// A `.claudeignore` larger than this is not read.
const LIST_LIMIT: u64 = 256 * 1024;

const APPLIES: &str = "Deny rules apply without the workspace trust step, headless runs included, \
    except in a run that does not load project settings (--setting-sources user, or an SDK \
    session without them).";

const REACH: &str = "A Read deny rule reaches Claude Code's file tools and the Bash file commands \
    and redirections it recognizes; it does not reach NotebookEdit or a process that opens the \
    file itself.";

const BLOCK_WARNING: &str = "also blocks editing and writing the matching paths, not only reading";

const CARVE_WARNING: &str = "an exception: carves the matching paths out of earlier deny rules of \
    the same settings file, lifting their block on reading and editing";

const NO_EFFECT: &str = "Claude Code does not read .claudeignore, so the file has no effect, \
    whether or not any of these rules is applied.";

/// The preview for the project folder `project`.
pub fn plan(project: &Path) -> Result<Value, String> {
    plan_with_bound(project, WALK_BOUND)
}

/// [`plan`] with the walk bound given by the caller.
pub fn plan_with_bound(project: &Path, bound: usize) -> Result<Value, String> {
    let root = fs::canonicalize(project)
        .map_err(|e| format!("project folder {}: {e}", project.display()))?;
    if !root.is_dir() {
        return Err(format!("{} is not a folder", root.display()));
    }
    let walk = walk(&root, bound);
    let tracked = tracked_files(&root);
    // `null` when git cannot say, so that an unknown answer never reads as "none tracked".
    let tracked_among = |matched: &[&String]| match &tracked {
        Some(set) => json!(matched
            .iter()
            .filter(|f| set.contains(f.as_str()))
            .collect::<Vec<_>>()),
        None => Value::Null,
    };

    let key_rules: Vec<Value> = KEY_SHAPES
        .iter()
        .filter_map(|shape| {
            let matched: Vec<&String> = walk
                .files
                .iter()
                .filter(|f| glob_match(shape.as_bytes(), file_name(f).as_bytes()))
                .collect();
            (!matched.is_empty()).then(|| {
                json!({
                    "rule": format!("Read({shape})"),
                    "shape": shape,
                    "matched": matched,
                    "tracked": tracked_among(&matched),
                    "warning": BLOCK_WARNING,
                })
            })
        })
        .collect();

    let claudeignore = match read_list(&root.join(".claudeignore")) {
        List::Absent => json!({"state": "absent", "entries": []}),
        List::NotRead(reason) => json!({
            "state": "not_read",
            "reason": reason,
            "note": NO_EFFECT,
            "entries": [],
        }),
        List::Text(text) => {
            let entries: Vec<Value> = list_entries(&text)
                .map(|(line, entry)| {
                    let matched: Vec<&String> = walk
                        .files
                        .iter()
                        .filter(|f| entry_selects(entry, f))
                        .collect();
                    let negation = entry.starts_with('!');
                    json!({
                        "line": line,
                        "entry": entry,
                        "rule": deny_rule(entry),
                        "matched": matched,
                        "tracked": tracked_among(&matched),
                        "recommended": credential_shaped(entry),
                        "warning": if negation { CARVE_WARNING } else { BLOCK_WARNING },
                    })
                })
                .collect();
            json!({"state": "present", "note": NO_EFFECT, "entries": entries})
        }
    };

    Ok(json!({
        "project": root.display().to_string(),
        "settings_file": root.join(".claude").join("settings.json").display().to_string(),
        "key_rules": key_rules,
        "claudeignore": claudeignore,
        "applies": APPLIES,
        "reach": REACH,
        "tracked_files": if tracked.is_some() { "known" } else { "unknown" },
        "walk": walk.report(bound),
    }))
}

/// What the walk saw: file paths from the root with `/` between the names, sorted.
struct Walk {
    files: Vec<String>,
    seen: usize,
    stopped: bool,
    nested: Vec<String>,
    unreadable: Vec<String>,
}

impl Walk {
    fn report(&self, bound: usize) -> Value {
        let note = if self.stopped {
            Some(format!(
                "The walk stopped at its bound of {bound} entries, so the file lists are partial."
            ))
        } else if !self.unreadable.is_empty() {
            Some("Some folders could not be read, so the file lists are partial.".to_string())
        } else {
            None
        };
        json!({
            "bound": bound,
            "entries_seen": self.seen,
            "complete": note.is_none(),
            "note": note,
            "skipped_nested_checkouts": self.nested,
            "unreadable_folders": self.unreadable,
        })
    }
}

/// Lists the files under `root`, shallow folders first, so that a large folder deep in the tree
/// cannot spend the bound before the files near the root are seen. Ignored files are included.
/// `.git` and a folder below the root with its own `.git` entry are not entered, and no symbolic
/// link is followed: a link is listed under its own name.
fn walk(root: &Path, bound: usize) -> Walk {
    let mut walk = Walk {
        files: Vec::new(),
        seen: 0,
        stopped: false,
        nested: Vec::new(),
        unreadable: Vec::new(),
    };
    let mut queue = VecDeque::from([String::new()]);
    while let Some(folder) = queue.pop_front() {
        let path = root.join(&folder);
        if !folder.is_empty() && fs::symlink_metadata(path.join(".git")).is_ok() {
            walk.nested.push(folder);
            continue;
        }
        let shown = if folder.is_empty() { "." } else { &folder };
        let Ok(reader) = fs::read_dir(&path) else {
            walk.unreadable.push(shown.to_string());
            continue;
        };
        let mut entries = Vec::new();
        for entry in reader {
            if walk.seen == bound {
                walk.stopped = true;
                break;
            }
            match entry {
                Ok(entry) => {
                    walk.seen += 1;
                    entries.push(entry);
                }
                Err(_) => walk.unreadable.push(shown.to_string()),
            }
        }
        entries.sort_by_key(|e| e.file_name());
        for entry in entries {
            let name = entry.file_name().to_string_lossy().into_owned();
            if name == ".git" {
                continue;
            }
            let relative = if folder.is_empty() {
                name
            } else {
                format!("{folder}/{name}")
            };
            // The entry's own type: a link to a folder is a link, never entered.
            match entry.file_type() {
                Ok(kind) if kind.is_dir() => queue.push_back(relative),
                Ok(_) => walk.files.push(relative),
                Err(_) => walk.unreadable.push(relative),
            }
        }
        if walk.stopped {
            break;
        }
    }
    walk.files.sort();
    walk.nested.sort();
    walk
}

/// The files git tracks under `root`, or `None` when git cannot say.
fn tracked_files(root: &Path) -> Option<HashSet<String>> {
    let mut cmd = Command::new("git");
    // A repository's own fsmonitor program must not run for a preview.
    cmd.args(["--no-optional-locks", "-c", "core.fsmonitor=false", "-C"])
        .arg(root)
        .args(["ls-files", "-z"]);
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .env("GIT_OPTIONAL_LOCKS", "0")
        .env("GIT_TERMINAL_PROMPT", "0")
        .stdin(Stdio::null())
        .stderr(Stdio::null())
        .output()
        .ok()?;
    out.status.success().then(|| {
        out.stdout
            .split(|b| *b == 0)
            .filter(|name| !name.is_empty())
            .map(|name| String::from_utf8_lossy(name).into_owned())
            .collect()
    })
}

enum List {
    Absent,
    NotRead(String),
    Text(String),
}

/// Reads the ignore list itself, only when it is a regular file: a link could name a key file.
fn read_list(path: &Path) -> List {
    let meta = match fs::symlink_metadata(path) {
        Ok(meta) => meta,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return List::Absent,
        Err(e) => return List::NotRead(e.to_string()),
    };
    if !meta.is_file() {
        return List::NotRead("not a regular file".to_string());
    }
    if meta.len() > LIST_LIMIT {
        return List::NotRead(format!("larger than {LIST_LIMIT} bytes"));
    }
    match fs::read_to_string(path) {
        Ok(text) => List::Text(text),
        Err(e) => List::NotRead(e.to_string()),
    }
}

/// The entries of the list with their 1-based line numbers: blank lines and comments left out.
fn list_entries(text: &str) -> impl Iterator<Item = (usize, &str)> {
    text.trim_start_matches('\u{feff}')
        .lines()
        .enumerate()
        .map(|(i, line)| (i + 1, line.trim_end()))
        .filter(|(_, line)| !line.is_empty() && !line.starts_with('#'))
}

/// The `Read` deny rule for one entry (D-5): a folder entry takes the documented form for a
/// folder's contents; every other entry, negations included, is carried as written.
fn deny_rule(entry: &str) -> String {
    if !entry.starts_with('!') && entry.ends_with('/') {
        let folder = entry.trim_end_matches('/');
        if folder.contains('/') {
            return format!("Read({folder}/**/*)");
        }
        if !folder.is_empty() {
            return format!("Read(**/{folder}/**/*)");
        }
    }
    format!("Read({entry})")
}

/// Whether the entry's pattern selects `file`, read as `.gitignore` reads it: without a `/`
/// inside it matches a name at any depth, with one it is anchored at the root; a folder it
/// matches brings its files; a trailing `/` matches folders only. Case is kept as written.
fn entry_selects(entry: &str, file: &str) -> bool {
    let body = entry.strip_prefix('!').unwrap_or(entry);
    let folder_only = body.ends_with('/');
    let body = body.trim_end_matches('/');
    let anchored = body.contains('/');
    let pattern = body.strip_prefix("./").unwrap_or(body);
    let pattern = pattern.trim_start_matches('/').as_bytes();
    let names: Vec<&str> = file.split('/').collect();
    let depth = names.len() - usize::from(folder_only);
    !pattern.is_empty()
        && (1..=depth).any(|n| {
            if anchored {
                glob_match(pattern, names[..n].join("/").as_bytes())
            } else {
                glob_match(pattern, names[n - 1].as_bytes())
            }
        })
}

/// Whether the entry itself has a credential file shape: the only entries recommended (D-4). A
/// negation never is; a folder entry has no last name, so no shape matches it.
fn credential_shaped(entry: &str) -> bool {
    !entry.starts_with('!')
        && KEY_SHAPES
            .iter()
            .any(|shape| glob_match(shape.as_bytes(), file_name(entry).as_bytes()))
}

fn file_name(path: &str) -> &str {
    path.rsplit('/').next().unwrap_or(path)
}
