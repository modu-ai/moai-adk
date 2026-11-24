//! The questions behind the model and effort choice (SPEC-20261118-model-effort-questions, card
//! C-47 slice 2a): the tool pick, the model and effort selections of each picked tool and role,
//! the reader of one answer line, a question loop over any reader and writer, the plan of the
//! `launcher.models` key and the digest-bound apply to the launcher settings file given by path.
//!
//! The module resolves no HOME and reads no terminal itself: the caller passes the reader, the
//! writer and the path. Every refusal names a key path or a reason and never quotes an answer
//! line or saved text. The update command, the commands that expose the plan and the apply, and
//! the skill step are slice 2b.
use crate::model_effort_settings::{
    default_choice, flat_choice, role_entry_paths, supported, unsupported_entries,
    with_flat_choice, Choice,
};
use serde_json::{Map, Value};
use sha2::{Digest, Sha256};
use std::fs;
use std::io::{self, BufRead, Read, Write};
use std::path::Path;

/// The tools a person can pick, in question order (user decision 2026-11-21: Claude Code and GLM
/// only; Codex support is withdrawn). A pick writes its own tool alone.
pub const TOOL_CHOICES: [&str; 2] = ["claude", "glm"];
/// Bad answers to one question before the loop gives up.
const MAX_TRIES: usize = 3;
/// The launcher settings file is small; a larger one is not read.
const MAX_FILE_BYTES: u64 = 1024 * 1024;
/// An answer line with more content bytes than this (newline not counted) is one invalid answer;
/// the rest of the line is dropped.
const MAX_LINE_BYTES: u64 = 4096;

/// Whether a question takes one option or several.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Kind {
    One,
    Many,
}

/// One option: the value that is saved and the text shown.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct QOption {
    pub value: String,
    pub label: String,
}

/// A question: `defaults` are indexes into `options`.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Question {
    pub id: String,
    pub prompt: String,
    pub kind: Kind,
    pub options: Vec<QOption>,
    pub defaults: Vec<usize>,
}

fn tool_label(tool: &str) -> &'static str {
    match tool {
        "claude" => "Claude",
        "glm" => "GLM",
        _ => "",
    }
}

fn option(value: &str, label: &str) -> QOption {
    QOption {
        value: value.to_string(),
        label: label.to_string(),
    }
}

/// Which tools to set up. The default is the installed tools the caller passes.
pub fn tool_question(installed: &[&str]) -> Question {
    Question {
        id: "tools".into(),
        prompt: "Which coding tools do you want to set up?".into(),
        kind: Kind::Many,
        options: TOOL_CHOICES
            .iter()
            .map(|tool| option(tool, tool_label(tool)))
            .collect(),
        defaults: TOOL_CHOICES
            .iter()
            .enumerate()
            .filter(|(_, tool)| installed.contains(tool))
            .map(|(index, _)| index)
            .collect(),
    }
}

/// The table row the questions offer first when nothing is saved: the `leader` row of
/// [`default_choice`], the model and effort a launch is closest to (SPEC-20261124-launcher-retire-
/// role-flags: a launch has no role, so the row is a starting point of the screen, not a role).
fn table_default(tool: &str) -> Option<Choice> {
    default_choice(tool, "leader")
}

/// The model of a tool: the default first, then the rest of the table in order.
pub fn model_question(tool: &str) -> Option<Question> {
    model_question_with(tool, None)
}

/// [`model_question`] where a saved supported choice is the default instead of the table default,
/// so that Enter keeps what the person saved earlier.
pub fn model_question_with(tool: &str, saved: Option<&Choice>) -> Option<Question> {
    let default = match saved {
        Some(choice) => choice.clone(),
        None => table_default(tool)?,
    };
    let rows = supported(tool)?;
    let mut options = vec![option(&default.model, &default.model)];
    options.extend(
        rows.iter()
            .filter(|row| row.id != default.model)
            .map(|row| option(row.id, row.id)),
    );
    Some(Question {
        id: format!("{tool}.model"),
        prompt: format!("{}: which model?", tool_label(tool)),
        kind: Kind::One,
        options,
        defaults: vec![0],
    })
}

/// The effort of a model for a tool: the efforts that model offers in table order. The default is
/// the table default when the model offers it, else `high`, else the first.
pub fn effort_question(tool: &str, model: &str) -> Option<Question> {
    effort_question_with(tool, model, None)
}

/// [`effort_question`] where the saved effort is the default when the saved model is the model
/// asked about and still offers it.
pub fn effort_question_with(tool: &str, model: &str, saved: Option<&Choice>) -> Option<Question> {
    let preferred = match saved {
        Some(choice) if choice.model == model => choice.effort.clone(),
        _ => table_default(tool)?.effort,
    };
    let row = supported(tool)?.iter().find(|row| row.id == model)?;
    let position = |name: &str| row.efforts.iter().position(|e| *e == name);
    let default = position(&preferred)
        .or_else(|| position("high"))
        .unwrap_or(0);
    Some(Question {
        id: format!("{tool}.effort"),
        prompt: format!("{}: which effort for that model?", tool_label(tool)),
        kind: Kind::One,
        options: row.efforts.iter().map(|e| option(e, e)).collect(),
        defaults: vec![default],
    })
}

/// What one answer line means.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Answer {
    Values(Vec<String>),
    Invalid,
    Cancelled,
}

/// Reads one answer line. `None` is the end of input. An empty line is the default, numbers are
/// 1-based options separated by commas or blanks; a single choice takes exactly one. The line is
/// never copied into the result.
pub fn read_answer(question: &Question, line: Option<&str>) -> Answer {
    let Some(line) = line else {
        return Answer::Cancelled;
    };
    let tokens: Vec<&str> = line
        .split(|c: char| c == ',' || c.is_whitespace())
        .filter(|token| !token.is_empty())
        .collect();
    let mut chosen: Vec<usize> = if tokens.is_empty() {
        question.defaults.clone()
    } else {
        let mut numbers = Vec::new();
        for token in tokens {
            if !token.bytes().all(|b| b.is_ascii_digit()) {
                return Answer::Invalid;
            }
            match token.parse::<usize>() {
                Ok(n) if (1..=question.options.len()).contains(&n) => {
                    if !numbers.contains(&(n - 1)) {
                        numbers.push(n - 1);
                    }
                }
                _ => return Answer::Invalid,
            }
        }
        numbers
    };
    if question.kind == Kind::One && chosen.len() != 1 {
        return Answer::Invalid;
    }
    chosen.sort_unstable();
    Answer::Values(
        chosen
            .into_iter()
            .map(|i| question.options[i].value.clone())
            .collect(),
    )
}

/// One pick: the model and effort of a tool (one pick per host, no role).
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Pick {
    pub tool: String,
    pub model: String,
    pub effort: String,
}

/// How the question loop ended.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Outcome {
    Picks(Vec<Pick>),
    Cancelled,
}

fn show<W: Write>(out: &mut W, question: &Question) -> io::Result<()> {
    writeln!(out, "{}", question.prompt)?;
    for (index, option) in question.options.iter().enumerate() {
        let mark = if question.defaults.contains(&index) {
            " (default)"
        } else {
            ""
        };
        writeln!(out, "  {}) {}{mark}", index + 1, option.label)?;
    }
    let hint = match question.kind {
        Kind::Many => "Numbers separated by commas, Enter for the default",
        Kind::One => "One number, Enter for the default",
    };
    write!(out, "{hint}: ")?;
    out.flush()
}

/// Reads one line of at most [`MAX_LINE_BYTES`]. `None` is the end of input. A longer line is
/// dropped up to its newline and read as `?`, which no question accepts.
pub fn read_line_bounded<R: BufRead>(input: &mut R) -> io::Result<Option<String>> {
    let mut bytes = Vec::new();
    let read = (&mut *input)
        .take(MAX_LINE_BYTES + 1)
        .read_until(b'\n', &mut bytes)?;
    if read == 0 {
        return Ok(None);
    }
    if bytes.last() != Some(&b'\n') && read as u64 == MAX_LINE_BYTES + 1 {
        loop {
            let mut rest = Vec::new();
            let more = (&mut *input)
                .take(MAX_LINE_BYTES)
                .read_until(b'\n', &mut rest)?;
            if more == 0 || rest.last() == Some(&b'\n') {
                break;
            }
        }
        return Ok(Some("?".to_string()));
    }
    Ok(Some(String::from_utf8_lossy(&bytes).into_owned()))
}

/// Asks one question up to [`MAX_TRIES`] times; `None` is cancelled (end of input or too many
/// bad answers).
fn ask<R: BufRead, W: Write>(
    input: &mut R,
    out: &mut W,
    question: &Question,
) -> io::Result<Option<Vec<String>>> {
    for _ in 0..MAX_TRIES {
        show(out, question)?;
        let line = read_line_bounded(input)?;
        match read_answer(question, line.as_deref()) {
            Answer::Values(values) => return Ok(Some(values)),
            Answer::Cancelled => return Ok(None),
            Answer::Invalid => writeln!(out, "That is not one of the numbers above.")?,
        }
    }
    Ok(None)
}

/// Runs the questions: the tool pick first, then for each picked tool (in option order) one model
/// and one effort. An empty tool pick asks nothing more.
pub fn run<R: BufRead, W: Write>(
    input: &mut R,
    out: &mut W,
    installed: &[&str],
) -> io::Result<Outcome> {
    run_with(input, out, installed, &Map::new())
}

/// [`run`] where the saved supported choices of the launcher object are the defaults of the
/// questions: Enter keeps what the person saved, and only an entry nobody saved gets the table
/// default.
pub fn run_with<R: BufRead, W: Write>(
    input: &mut R,
    out: &mut W,
    installed: &[&str],
    launcher: &Map<String, Value>,
) -> io::Result<Outcome> {
    let Some(tools) = ask(input, out, &tool_question(installed))? else {
        return Ok(Outcome::Cancelled);
    };
    let mut picks = Vec::new();
    for tool in TOOL_CHOICES
        .iter()
        .filter(|t| tools.iter().any(|p| p.as_str() == **t))
    {
        let saved = flat_choice(launcher, tool);
        let Some(models) = model_question_with(tool, saved.as_ref()) else {
            continue;
        };
        let Some(model) = ask(input, out, &models)? else {
            return Ok(Outcome::Cancelled);
        };
        let model = model[0].clone();
        let Some(efforts) = effort_question_with(tool, &model, saved.as_ref()) else {
            continue;
        };
        let Some(effort) = ask(input, out, &efforts)? else {
            return Ok(Outcome::Cancelled);
        };
        picks.push(Pick {
            tool: (*tool).to_string(),
            model,
            effort: effort[0].clone(),
        });
    }
    Ok(Outcome::Picks(picks))
}

/// A saved or new value of one tool in a plan.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Change {
    pub tool: String,
    /// The supported value saved before; `None` when nothing or an unsupported value was saved.
    pub from: Option<Choice>,
    pub to: Choice,
}

/// What a set of picks would do to the launcher object.
#[derive(Clone, Debug, PartialEq)]
pub struct Plan {
    pub changes: Vec<Change>,
    /// Key paths of the saved `leader` and `agent` entries the write replaces (the launcher no
    /// longer reads them; SPEC-20261124-launcher-retire-role-flags REQ-6).
    pub replaced: Vec<String>,
    /// Key paths of saved entries outside the support table that the picks leave unchanged.
    pub kept_unsupported: Vec<String>,
    pub after: Map<String, Value>,
    /// SHA-256 over the sorted content before, the sorted picks and the sorted content after.
    pub digest: String,
}

/// Key order is not meaningful, so the digest is taken over a sorted rendering.
fn canon(value: &Value) -> String {
    match value {
        Value::Object(map) => {
            let mut keys: Vec<&String> = map.keys().collect();
            keys.sort();
            let body: Vec<String> = keys
                .into_iter()
                .map(|k| format!("{}:{}", Value::String(k.clone()), canon(&map[k])))
                .collect();
            format!("{{{}}}", body.join(","))
        }
        Value::Array(items) => {
            let body: Vec<String> = items.iter().map(canon).collect();
            format!("[{}]", body.join(","))
        }
        other => other.to_string(),
    }
}

fn targets(tool: &str) -> Vec<&str> {
    // A pick writes its own tool alone (the earlier OMP pick also set glm; OMP is not a tool
    // any more, and a saved OMP entry is ignored like a saved Codex entry).
    vec![tool]
}

/// Plans the picks on a launcher object. Refuses a structurally broken `launcher.models` and a
/// pick the support table does not accept; the input is not changed.
pub fn plan(launcher: &Map<String, Value>, picks: &[Pick]) -> Result<Plan, String> {
    let unsupported = unsupported_entries(launcher)?;
    let mut after = launcher.clone();
    let mut changes = Vec::new();
    let mut replaced = Vec::new();
    let mut role_entries = Vec::new();
    for pick in picks {
        for tool in targets(&pick.tool) {
            after = with_flat_choice(&after, tool, &pick.model, &pick.effort)?;
            changes.push(Change {
                tool: tool.to_string(),
                from: flat_choice(launcher, tool),
                to: Choice {
                    model: pick.model.clone(),
                    effort: pick.effort.clone(),
                },
            });
            // The write replaces the whole `launcher.models.<tool>` object: its own path and
            // every path under it, whatever the support table said about them.
            replaced.push(format!("launcher.models.{tool}"));
            role_entries.extend(role_entry_paths(launcher, tool));
        }
    }
    let under = |path: &str, tool_path: &String| {
        path == tool_path
            || path
                .strip_prefix(tool_path.as_str())
                .is_some_and(|r| r.starts_with('.'))
    };
    let kept_unsupported = unsupported
        .into_iter()
        .filter(|path| !replaced.iter().any(|tool_path| under(path, tool_path)))
        .collect();
    let mut sorted: Vec<[&str; 3]> = picks
        .iter()
        .map(|p| [&*p.tool, &*p.model, &*p.effort])
        .collect();
    sorted.sort_unstable();
    let text = format!(
        "{}\n{:?}\n{}",
        canon(&Value::Object(launcher.clone())),
        sorted,
        canon(&Value::Object(after.clone()))
    );
    Ok(Plan {
        changes,
        replaced: role_entries,
        kept_unsupported,
        after,
        digest: format!("{:x}", Sha256::digest(text.as_bytes())),
    })
}

/// The bytes of the settings file, `None` when it is absent. The entry is checked by `lstat` before
/// it is opened, so a link, a folder or a special file is never read, and a file over one MiB is
/// refused.
fn read_bytes(path: &Path) -> Result<Option<Vec<u8>>, String> {
    match fs::symlink_metadata(path) {
        Err(e) if e.kind() == io::ErrorKind::NotFound => return Ok(None),
        Err(_) => return Err("launcher settings cannot be read".into()),
        Ok(meta) if !meta.file_type().is_file() => {
            return Err("launcher settings must be a regular file".into())
        }
        Ok(meta) if meta.len() > MAX_FILE_BYTES => {
            return Err("launcher settings are larger than one MiB".into())
        }
        Ok(_) => {}
    }
    let bytes = fs::read(path).map_err(|_| "launcher settings cannot be read".to_string())?;
    if bytes.len() as u64 > MAX_FILE_BYTES {
        return Err("launcher settings are larger than one MiB".into());
    }
    Ok(Some(bytes))
}

/// The text of the settings file, `None` when it is absent (same rules as a plan reads it).
pub fn read_text(path: &Path) -> Result<Option<String>, String> {
    read_bytes(path)?
        .map(|bytes| {
            String::from_utf8(bytes).map_err(|_| "launcher settings must be UTF-8 text".to_string())
        })
        .transpose()
}

/// The root object of the launcher settings file; an absent file is an empty object. The entry is
/// checked by `lstat` before it is opened, so a link, a folder or a special file is never read.
fn read_root(path: &Path) -> Result<Map<String, Value>, String> {
    let Some(bytes) = read_bytes(path)? else {
        return Ok(Map::new());
    };
    match serde_json::from_slice::<Value>(&bytes) {
        Ok(Value::Object(root)) => Ok(root),
        Ok(_) => Err("launcher settings must be an object".into()),
        Err(_) => Err("launcher settings must use valid JSON-subset syntax".into()),
    }
}

/// The `launcher` object of the settings file (empty when the file or the key is absent), read
/// by the same rules as a plan: a link, a folder, a special file or more than one MiB is refused.
pub fn read_launcher(path: &Path) -> Result<Map<String, Value>, String> {
    launcher_of(&read_root(path)?)
}

fn launcher_of(root: &Map<String, Value>) -> Result<Map<String, Value>, String> {
    match root.get("launcher") {
        None => Ok(Map::new()),
        Some(Value::Object(launcher)) => Ok(launcher.clone()),
        Some(_) => Err("launcher settings launcher must be an object".into()),
    }
}

/// The preview: plans the picks on the file at `path` and writes nothing.
pub fn plan_file(path: &Path, picks: &[Pick]) -> Result<Plan, String> {
    plan(&launcher_of(&read_root(path)?)?, picks)
}

/// Distinguishes the temporary files of one process.
static TEMP_SERIAL: std::sync::atomic::AtomicU64 = std::sync::atomic::AtomicU64::new(0);

/// Writes `root` next to `path` and renames it over `path`: the mode of an existing file is kept,
/// a new file is private to the owner.
fn write_atomic(path: &Path, root: &Map<String, Value>) -> Result<(), String> {
    write_text(path, &render_root(root)?)
}

/// The text a write of `root` puts into the file: pretty JSON and one newline.
pub fn render_root(root: &Map<String, Value>) -> Result<String, String> {
    let mut text = serde_json::to_string_pretty(&Value::Object(root.clone()))
        .map_err(|_| "launcher settings cannot be written".to_string())?;
    text.push('\n');
    Ok(text)
}

/// The text of the file after the apply of `plan` on the file now at `path`: the other root keys
/// as they are, the launcher object of the plan.
pub fn after_text(path: &Path, plan: &Plan) -> Result<String, String> {
    let mut next = read_root(path)?;
    next.insert("launcher".to_string(), Value::Object(plan.after.clone()));
    render_root(&next)
}

/// Removes the settings file when it is a regular file (never a link, folder or special file) and
/// its folder is not a link: a link there would carry the removal to a file outside the home
/// (C-47 slice 5 audit B1, the same check as [`write_text`]).
pub fn remove_settings(path: &Path) -> Result<(), String> {
    if let Some(parent) = path.parent() {
        if fs::symlink_metadata(parent).is_ok_and(|m| m.file_type().is_symlink()) {
            return Err("the folder of the launcher settings must not be a link".into());
        }
    }
    match fs::symlink_metadata(path) {
        Ok(meta) if meta.file_type().is_file() => {
            fs::remove_file(path).map_err(|_| "launcher settings cannot be removed".to_string())
        }
        Err(e) if e.kind() == io::ErrorKind::NotFound => Ok(()),
        _ => Err("launcher settings must be a regular file".into()),
    }
}

/// Writes `text` next to `path` and renames it over `path` (the rules of [`write_atomic`]).
pub fn write_text(path: &Path, text: &str) -> Result<(), String> {
    let parent = path
        .parent()
        .ok_or_else(|| "launcher settings path has no folder".to_string())?;
    if fs::symlink_metadata(parent).is_ok_and(|m| m.file_type().is_symlink()) {
        return Err("the folder of the launcher settings must not be a link".into());
    }
    #[cfg(unix)]
    let mode = {
        use std::os::unix::fs::PermissionsExt;
        fs::symlink_metadata(path)
            .map(|m| m.permissions().mode() & 0o777)
            .unwrap_or(0o600)
    };
    // The reader refuses a file over the limit, so a file that would grow past it is not written.
    if text.len() as u64 > MAX_FILE_BYTES {
        return Err("launcher settings would be larger than one MiB; nothing was changed".into());
    }
    fs::create_dir_all(parent)
        .map_err(|_| "launcher settings folder cannot be created".to_string())?;
    // A name no other writer uses; a file that exists under it is never opened or removed.
    let stamp = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map_or(0, |d| d.as_nanos());
    let serial = TEMP_SERIAL.fetch_add(1, std::sync::atomic::Ordering::Relaxed);
    let temp = parent.join(format!(
        ".launcher.tmp-{}-{stamp}-{serial}",
        std::process::id()
    ));
    let mut options = fs::OpenOptions::new();
    options.write(true).create_new(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        options.mode(mode);
    }
    // Only a file this call created is ever removed.
    let mut file = options
        .open(&temp)
        .map_err(|_| "launcher settings cannot be written; nothing was changed".to_string())?;
    let written = (|| -> io::Result<()> {
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            file.set_permissions(fs::Permissions::from_mode(mode))?;
        }
        file.write_all(text.as_bytes())?;
        file.sync_all()?;
        Ok(())
    })();
    drop(file);
    if written.and_then(|()| fs::rename(&temp, path)).is_err() {
        let _ = fs::remove_file(&temp);
        return Err("launcher settings cannot be written; nothing was changed".into());
    }
    Ok(())
}

/// The apply: plans again from the file and the picks, writes only when the digest equals the
/// preview's, keeps every other root key, and reads the result back.
pub fn apply(path: &Path, picks: &[Pick], digest: &str) -> Result<Plan, String> {
    let root = read_root(path)?;
    let current = launcher_of(&root)?;
    let planned = plan(&current, picks)?;
    if planned.digest != digest {
        return Err(
            "the launcher settings or the picks changed since the preview; preview again, nothing was written"
                .into(),
        );
    }
    // Nothing to change: the file is left as it is, byte for byte, and an absent file stays absent.
    if planned.after == current {
        return Ok(planned);
    }
    let mut next = root;
    next.insert("launcher".to_string(), Value::Object(planned.after.clone()));
    write_atomic(path, &next)?;
    if read_root(path)? != next {
        return Err(
            "launcher settings were replaced but read back differently than written; check the file"
                .into(),
        );
    }
    Ok(planned)
}

/// The key vault scope of the project `dir` belongs to, resolved with `home` as the HOME: the
/// workspace identity then reads the home state of `home` and never the process environment's
/// HOME, so a run under a temporary home leaves the real one unread. `None` outside a git work
/// tree or when the identity does not resolve. The folder name is the one
/// `key_vault::project_scope` makes (same identity fields, same `folder_name`); that function
/// reads the environment HOME and is not changed here.
pub fn glm_key_scope(dir: &Path, home: &Path) -> Option<crate::key_vault::Scope> {
    let id = crate::identity::identity_in(dir, Some(home)).ok()?;
    if id["identity_status"] != "ok" || id["work_tree"].is_null() {
        return None;
    }
    let clone = id["clone_id"].as_str().unwrap_or("");
    let clone_hex = clone.strip_prefix("sha256:").unwrap_or(clone);
    if clone_hex.len() < 8 {
        return None;
    }
    // The primary worktree is the folder that holds the common git directory.
    let anchor = Path::new(id["clone_anchor"].as_str().unwrap_or(""));
    let primary = if anchor.file_name().is_some_and(|n| n == ".git") {
        anchor.parent().unwrap_or(anchor)
    } else {
        anchor
    };
    let primary_name = primary
        .file_name()
        .map(|n| n.to_string_lossy().trim_end_matches(".git").to_string())
        .unwrap_or_default();
    Some(crate::key_vault::Scope::Project(
        crate::key_vault::folder_name(id["project_id"].as_str(), &primary_name, clone_hex),
    ))
}

/// The model step's glm key question (SPEC-20261121-glm-key-at-pick): when the picks name glm,
/// the key is asked right here, before the launcher record. With `hidden` (the caller checked
/// that stdin and stderr are terminals) the value is read through the hidden prompt and never
/// echoed; otherwise it is one line of `input`, and nothing here touches a terminal. The value
/// goes to the key vault under `home` the way the key set command stores it and appears in no
/// output. An empty answer skips — the pick stays, and the first glm run keeps asking for the
/// key. A key the launcher would already find (this project's, then the shared one) is kept
/// unless the answer names a new one. The key is stored when it is answered, so a later no at
/// the launcher save question leaves it in the vault. Outside a git work tree the project has
/// no key folder: the question is skipped with one line saying how to store the key.
pub fn glm_key_step<R: BufRead, W: Write>(
    input: &mut R,
    out: &mut W,
    home: &Path,
    picks: &[Pick],
    hidden: bool,
) -> io::Result<()> {
    if !picks.iter().any(|p| p.tool == "glm") {
        return Ok(());
    }
    let cwd = std::env::current_dir().ok();
    let Some(scope) = cwd.as_deref().and_then(|dir| glm_key_scope(dir, home)) else {
        writeln!(
            out,
            "The glm key is kept per project and this folder is not inside one; the key entry was skipped (run `moai key-set glm --prompt` inside a project, or add --shared)."
        )?;
        return Ok(());
    };
    // The launcher's order (`key_vault::read`): the project's entry, then the shared one. A
    // project entry that exists but cannot be read (a broken file) ends the search there, as it
    // does for the launcher, and counts as no usable key: the answer then replaces it.
    let stored = match crate::key_vault::read_entry(home, &scope, "glm") {
        Ok(Some(value)) => crate::key_vault::usable_value(&value),
        Ok(None) => crate::key_vault::read_entry(home, &crate::key_vault::Scope::Shared, "glm")
            .ok()
            .flatten()
            .and_then(|value| crate::key_vault::usable_value(&value)),
        Err(_) => None,
    };
    if stored.is_some() {
        writeln!(
            out,
            "A GLM key is already in the key vault. Press Enter to keep it, or type n to enter a new one."
        )?;
        let answer = read_line_bounded(input)?;
        let wants_new = answer.as_deref().is_some_and(|a| {
            let word = a.trim();
            word.eq_ignore_ascii_case("n") || word.eq_ignore_ascii_case("new")
        });
        if !wants_new {
            return Ok(());
        }
    } else {
        writeln!(
            out,
            "The glm pick needs a GLM API key. The input is hidden on a terminal and is stored only in the key vault; press Enter to skip."
        )?;
    }
    let value = if hidden {
        // The hidden prompt reads the terminal itself; `input` is not read here, because a
        // read of an empty line buffer would wait for one more line.
        match crate::key_prompt::read_hidden("glm") {
            Ok(value) => Some(value),
            Err(reason) => {
                // An empty answer is the skip the question offers, not a failure to report.
                if reason.starts_with("no usable value") {
                    writeln!(out, "No key was entered; the key entry was skipped.")?;
                } else {
                    writeln!(out, "The key entry was skipped: {reason}")?;
                }
                None
            }
        }
    } else {
        read_line_bounded(input)?.and_then(|line| crate::key_vault::usable_value(line.trim()))
    };
    let Some(value) = value else {
        return Ok(());
    };
    // A refused write is told on the screen and the step goes on, so the model pick and the
    // save question after it are not lost; the first glm run keeps asking for the key.
    if let Err(reason) = crate::key_vault::set(home, &scope, "glm", &value) {
        writeln!(out, "The key could not be stored: {reason}")?;
    }
    Ok(())
}
