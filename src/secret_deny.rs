//! `secret-deny-plan` (SPEC-20261122-secret-file-deny-rules): the `Read` deny rules a project's
//! root `.claude/settings.json` could carry, built from the key files found by name and from the
//! entries of `.claudeignore`, which Claude Code does not read. It looks at names, reads
//! `.claudeignore` and git's list of tracked files, opens no matched file and writes nothing.
//!
//! `secret-deny-apply` (slice B): [`plan_picks`] previews the picked rules as the final
//! `permissions.deny` list and a digest, and [`apply`] adds only those rules to the root settings
//! file through the writer in `crate::claude`, then moves `.claudeignore` into the settings backup
//! folder when that second pick is made.

use crate::claude;
use crate::worktree_include::glob_match;
use serde_json::{json, Value};
use std::collections::{HashSet, VecDeque};
use std::fs;
use std::path::{Path, PathBuf};
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
    build(project, bound).map(|built| built.value)
}

/// A preview together with what the picks of [`plan_picks`] and [`apply`] are checked against.
struct Built {
    root: PathBuf,
    value: Value,
    /// Every rule the preview offers: the key rules and the converted `.claudeignore` entries.
    offered: Vec<String>,
}

fn build(project: &Path, bound: usize) -> Result<Built, String> {
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

    let list = read_list(&root.join(".claudeignore"));
    // The negations of the carried-over list: (rule, entry).
    let carried: Vec<(String, String)> = match &list {
        List::Text(text) => list_entries(text)
            .filter(|(_, entry)| entry.starts_with('!'))
            .map(|(_, entry)| (deny_rule(entry), entry.to_string()))
            .collect(),
        _ => Vec::new(),
    };
    let claudeignore = match &list {
        List::Absent => json!({"state": "absent", "entries": []}),
        List::NotRead(reason) => json!({
            "state": "not_read",
            "reason": reason,
            "note": NO_EFFECT,
            "entries": [],
        }),
        List::Text(text) => {
            let entries: Vec<Value> = list_entries(text)
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

    let offered: Vec<String> = key_rules
        .iter()
        .chain(claudeignore["entries"].as_array().into_iter().flatten())
        .filter_map(|r| r["rule"].as_str().map(str::to_string))
        .collect();

    let (settings, held) = settings_view(&root);
    // The negations that stand in the settings file come first, then the carried-over ones the
    // file does not hold yet.
    let mut negations: Vec<(String, String, &str)> = held
        .iter()
        .filter_map(|rule| negation_entry(rule).map(|e| (rule.clone(), e.to_string(), "settings")))
        .collect();
    for (rule, entry) in carried {
        if !negations.iter().any(|(have, ..)| *have == rule) {
            negations.push((rule, entry, "claudeignore"));
        }
    }
    let carved = carve_outs(&walk, &negations);

    let value = json!({
        "project": root.display().to_string(),
        "settings_file": root.join(".claude").join("settings.json").display().to_string(),
        "key_rules": key_rules,
        "claudeignore": claudeignore,
        "settings": settings,
        "carve_outs": carved,
        "applies": APPLIES,
        "reach": REACH,
        "tracked_files": if tracked.is_some() { "known" } else { "unknown" },
        "walk": walk.report(bound),
    });
    Ok(Built {
        root,
        value,
        offered,
    })
}

/// The root settings file as the preview reads it, without following a link, and the
/// `permissions.deny` strings it holds: the `settings` block of the answer and the list.
fn settings_view(root: &Path) -> (Value, Vec<String>) {
    let (ignored, sessions) = claude::settings_reach(root);
    let mut deny = Vec::new();
    let state = |file_state: &str, reason: Option<String>| {
        let mut block = json!({"file_state": file_state});
        if let Some(reason) = reason {
            block["reason"] = json!(reason);
        }
        block
    };
    let folder_is_link =
        fs::symlink_metadata(root.join(".claude")).is_ok_and(|meta| meta.file_type().is_symlink());
    let mut block = if folder_is_link {
        state("not_read", Some("the .claude folder is a link".to_string()))
    } else {
        match claude::read_existing(&root.join(".claude").join("settings.json")) {
            Ok(None) => state("absent", None),
            Err(reason) => state("not_read", Some(reason)),
            Ok(Some(bytes)) => match serde_json::from_slice::<Value>(&bytes) {
                Ok(Value::Object(map)) => {
                    deny = map
                        .get("permissions")
                        .and_then(|p| p.get("deny"))
                        .and_then(Value::as_array)
                        .into_iter()
                        .flatten()
                        .filter_map(|rule| rule.as_str().map(str::to_string))
                        .collect();
                    state("present", None)
                }
                _ => state("not_read", Some("not a JSON object".to_string())),
            },
        }
    };
    block["settings_ignored"] = json!(ignored);
    block["worktree_sessions"] = json!(sessions);
    block["negations"] = json!(deny
        .iter()
        .enumerate()
        .filter(|(_, rule)| is_negation(rule))
        .map(|(index, entry)| json!({"index": index, "entry": entry}))
        .collect::<Vec<_>>());
    (block, deny)
}

/// The `!` entries that select a file a key rule matches (REQ-6): one row per negation and key
/// rule, with the files the negation takes out of that rule. `negations` holds the rule as it
/// stands in the list, the entry as `.gitignore`-style matching reads it, and where it came from.
fn carve_outs(walk: &Walk, negations: &[(String, String, &str)]) -> Vec<Value> {
    let mut out = Vec::new();
    for (rule, entry, source) in negations {
        let selected: Vec<&String> = walk
            .files
            .iter()
            .filter(|file| entry_selects(entry, file))
            .collect();
        for shape in KEY_SHAPES {
            let files: Vec<&String> = selected
                .iter()
                .copied()
                .filter(|file| glob_match(shape.as_bytes(), file_name(file).as_bytes()))
                .collect();
            if !files.is_empty() {
                out.push(json!({
                    "negation": rule,
                    "rule": format!("Read({shape})"),
                    "files": files,
                    "source": source,
                    "warning": CARVE_WARNING,
                }));
            }
        }
    }
    out
}

/// The argument of a `Read(...)` rule.
fn read_argument(rule: &str) -> Option<&str> {
    rule.strip_prefix("Read(")
        .and_then(|rest| rest.strip_suffix(')'))
}

/// A `!` entry of a deny list is `Read(!<path>)` (D-5).
pub(crate) fn is_negation(rule: &str) -> bool {
    rule.starts_with("Read(!")
}

/// The `!<path>` an entry of the list carries, as an ignore entry.
fn negation_entry(rule: &str) -> Option<&str> {
    is_negation(rule).then(|| {
        let inner = &rule["Read(".len()..];
        inner.strip_suffix(')').unwrap_or(inner)
    })
}

/// A key-file rule (REQ-6): a `Read(...)` rule that is no negation and whose argument has a
/// credential file shape, the same test that recommends a `.claudeignore` entry. A rule for
/// `.env.production` is one, as is `Read(.env)`; a folder rule such as `Read(**/vendor/**/*)` is not.
pub(crate) fn is_key_rule(rule: &str) -> bool {
    read_argument(rule).is_some_and(credential_shaped)
}

/// The deny list after the picks (REQ-6, D-6): the entries that were there stay as they are, each
/// pick not yet in the list follows (the non-key picks first, then the key picks), and every key
/// rule that stands before the last `!` entry moves to just behind it, in its order, so that no
/// key rule is left for a negation to carve. Nothing else moves, and a list that holds no such
/// key rule is returned as it is.
pub(crate) fn order_deny(
    existing: &[String],
    new_non_key: &[String],
    new_key: &[String],
) -> Vec<String> {
    let mut list = existing.to_vec();
    for rule in new_non_key.iter().chain(new_key) {
        if !list.contains(rule) {
            list.push(rule.clone());
        }
    }
    let Some(last) = list.iter().rposition(|rule| is_negation(rule)) else {
        return list;
    };
    let misplaced: Vec<String> = list
        .iter()
        .take(last)
        .filter(|rule| is_key_rule(rule))
        .cloned()
        .collect();
    if misplaced.is_empty() {
        return list;
    }
    let mut ordered = Vec::with_capacity(list.len());
    let mut behind = 0;
    for (index, rule) in list.into_iter().enumerate() {
        if index < last && is_key_rule(&rule) {
            continue;
        }
        ordered.push(rule);
        if index == last {
            behind = ordered.len();
        }
    }
    ordered.splice(behind..behind, misplaced);
    ordered
}

/// The picks of one preview or apply, checked against the rules the preview offers.
struct Picked {
    built: Built,
    /// The picks in the order given, each once.
    picks: Vec<String>,
    /// The digest of the `.claudeignore` to move, when the second pick is made.
    retire: Option<String>,
    /// Bound into the plan digest next to the settings file and the receipt.
    bind: Vec<u8>,
}

impl Picked {
    fn extend(&self, existing: &[String]) -> Vec<String> {
        let (key, non_key): (Vec<String>, Vec<String>) = self
            .picks
            .iter()
            .cloned()
            .partition(|pick| is_key_rule(pick));
        order_deny(existing, &non_key, &key)
    }
}

fn picked(project: &Path, picks: &[String], retire: bool) -> Result<Picked, String> {
    let built = build(project, WALK_BOUND)?;
    let mut unique: Vec<String> = Vec::new();
    for pick in picks {
        if !unique.contains(pick) {
            unique.push(pick.clone());
        }
    }
    if unique.is_empty() {
        return Err("nothing picked: name at least one rule of the preview with --pick".into());
    }
    if let Some(pick) = unique.iter().find(|pick| !built.offered.contains(pick)) {
        return Err(format!(
            "pick {pick:?} is not a rule the preview offers; nothing written"
        ));
    }
    let retire = retire
        .then(|| claudeignore_digest(&built.root))
        .transpose()?;
    let bind = serde_json::to_vec(&("secret-deny-picks", &unique, &retire)).expect("bind tuple");
    Ok(Picked {
        built,
        picks: unique,
        retire,
        bind,
    })
}

/// The digest of the root `.claudeignore` when it can be moved: a regular file of the size the
/// preview reads. A link is refused, so no apply moves a link or what it names.
fn claudeignore_digest(root: &Path) -> Result<String, String> {
    let path = root.join(".claudeignore");
    let meta = match fs::symlink_metadata(&path) {
        Ok(meta) => meta,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
            return Err("there is no .claudeignore to retire; nothing written".into())
        }
        Err(e) => return Err(format!(".claudeignore: {e}; nothing written")),
    };
    if meta.file_type().is_symlink() {
        return Err(".claudeignore is a link; retiring a link is refused, nothing written".into());
    }
    if !meta.is_file() {
        return Err(".claudeignore is not a regular file; nothing written".into());
    }
    if meta.len() > LIST_LIMIT {
        return Err(format!(
            ".claudeignore is larger than {LIST_LIMIT} bytes; nothing written"
        ));
    }
    fs::read(&path)
        .map(|bytes| crate::digest(&bytes))
        .map_err(|e| format!(".claudeignore: {e}; nothing written"))
}

/// What the answer says about the rules reaching sessions started in a linked worktree.
fn reach_message(ignored: &str) -> &'static str {
    match ignored {
        "not_ignored" => {
            "The rules reach sessions started in a worktree once .claude/settings.json is committed."
        }
        "ignored" => {
            "Git ignores .claude/settings.json, so the rules would not reach worktree sessions: track the file first."
        }
        _ => {
            "Whether git ignores .claude/settings.json is unknown here, so whether the rules reach worktree sessions is unknown."
        }
    }
}

/// The preview with the picks applied to the list: `apply` carries the digest the apply needs,
/// the list the file would hold and whether `.claudeignore` would be moved. Writes nothing. A pick
/// the preview does not offer, a `.claudeignore` that is no regular file when it is to be moved,
/// and a settings file the writer refuses (a link, no JSON object, too large) are refused here.
pub fn plan_picks(project: &Path, picks: &[String], retire: bool) -> Result<Value, String> {
    let picked = picked(project, picks, retire)?;
    let plan = claude::deny_add_plan(
        &picked.built.root,
        &|e: &[String]| picked.extend(e),
        &picked.bind,
    )?;
    let added: Vec<&String> = plan
        .final_deny
        .iter()
        .filter(|rule| !plan.before_deny.contains(rule))
        .collect();
    let mut apply = json!({
        "plan_digest": plan.digest,
        "picks": picked.picks,
        "final_deny": plan.final_deny,
        "adds": added,
        "retire_claudeignore": picked.retire.is_some(),
        "settings_ignored": plan.ignored,
        "worktree_sessions": plan.sessions,
        "note": reach_message(plan.ignored),
        "writes": false,
    });
    if !plan.observed.is_empty() {
        apply["would_refuse"] = json!(plan.observed);
    }
    let mut value = picked.built.value;
    value["apply"] = apply;
    Ok(value)
}

/// Adds the picked rules to `permissions.deny` of the root `.claude/settings.json` and, with the
/// second pick, moves `.claudeignore` into the settings backup folder (never deleting it).
/// Nothing is written unless `plan_digest` is the one [`plan_picks`] answers for the same picks
/// and flag now, and not while git ignores the settings file. The move is the last step: when it
/// fails the rules stay applied and the refusal says so, and a new preview finishes the move.
pub fn apply(
    project: &Path,
    plan_digest: &str,
    permission: &str,
    picks: &[String],
    retire: bool,
) -> Result<Value, String> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    let picked = picked(project, picks, retire)?;
    let root = picked.built.root.clone();
    let applied = claude::deny_add_apply(
        &root,
        permission,
        plan_digest,
        &|e: &[String]| picked.extend(e),
        &picked.bind,
    )?;
    let mut kept = None;
    if picked.retire.is_some() {
        kept = Some(
            retire_claudeignore(&root, applied.backup.as_ref()).map_err(|e| {
                format!(
                    "partial apply: the deny rules are in .claude/settings.json; .claudeignore was not moved: {e}; preview again with --retire-claudeignore to finish it"
                )
            })?,
        );
    }
    let mut message = if applied.written {
        format!(
            "Added the picked rules to .claude/settings.json. {}",
            reach_message(applied.ignored)
        )
    } else {
        format!(
            ".claude/settings.json already holds the picked rules in order; nothing was written. {}",
            reach_message(applied.ignored)
        )
    };
    if let Some(kept) = &kept {
        message.push_str(&format!(" .claudeignore was moved to {kept}."));
    }
    let mut out = json!({
        "applied": true,
        "written": applied.written,
        "state": if applied.written { "applied" } else { "unchanged" },
        "settings_file": ".claude/settings.json",
        "final_deny": applied.final_deny,
        "files_created": if applied.created { json!([".claude/settings.json"]) } else { json!([]) },
        "backup_dir": applied.backup.as_ref().map(|(_, rel)| rel),
        "receipt_refreshed": applied.receipt_refreshed,
        "retired_claudeignore": kept.is_some(),
        "claudeignore_kept_as": kept,
        "settings_ignored": applied.ignored,
        "worktree_sessions": applied.sessions,
        "message": message,
        "host_loaded": "not_run",
    });
    if !applied.observed.is_empty() {
        out["would_refuse"] = json!(applied.observed);
    }
    if applied.unrecorded {
        out["observation_recorded"] = json!(false);
    }
    Ok(out)
}

/// Moves the root `.claudeignore` into the settings backup folder: the one `kept` by this apply
/// when it replaced a settings original, else a new one. A rename, so the bytes are the same and
/// nothing is copied or deleted; the file must still be a regular file and the target must be
/// free. Answers the project-relative path of the kept file.
fn retire_claudeignore(root: &Path, kept: Option<&(PathBuf, String)>) -> Result<String, String> {
    let from = root.join(".claudeignore");
    if !fs::symlink_metadata(&from).is_ok_and(|meta| meta.is_file()) {
        return Err(".claudeignore is no longer a regular file".into());
    }
    let (dir, rel) = match kept {
        Some((dir, rel)) => (dir.clone(), rel.clone()),
        None => claude::backup_folder_of(root)?,
    };
    let to = dir.join(".claudeignore");
    if fs::symlink_metadata(&to).is_ok() {
        return Err(format!("{rel}/.claudeignore is already there"));
    }
    fs::rename(&from, &to).map_err(|e| format!("the move into {rel} failed ({e})"))?;
    Ok(format!("{rel}/.claudeignore"))
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
