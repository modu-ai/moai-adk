//! The readers of the go-residue survey (SPEC-20261126-go-residue-in-use): which project files
//! still read a Go-generation file, and which readers could not be read at all.
//!
//! A reader is an instruction file (`CLAUDE.md` and its kin, `.claude/rules/**/*.md`) with the
//! files it imports through `@path`, the two settings files with the scripts their strings name,
//! and the status line's own configuration files. Only a name is ever taken from a reader: a
//! setting value is never kept, so what leaves this module is a project path and a kind
//! (`import`, `settings`, `script`, `statusline`) or a project path and a fixed reason.
//!
//! Over-keeping is the safe error here: a reader that cannot be read, or an import or script
//! that leaves the project, makes the survey keep every Go file (REQ-3); a name that matches by
//! accident only keeps one file more. Reading goes through `go_state::read_bounded_text`, so a
//! link on the path, more than 4 MiB and text that is not UTF-8 are all refused the same way. A
//! reader whose path holds a secret-shaped name (`go_state::has_secret_component`) is not opened
//! at all and counts as unread (SPEC-20261127-go-residue-secret-readers).

use serde_json::Value;
use std::{
    collections::{BTreeMap, BTreeSet, HashMap},
    fs,
    io::Read as _,
    path::Path,
};

use crate::go_state::{has_secret_component, linked_component, read_bounded_text};

/// The largest reader file read: the limit Claude Code puts on an `@` import.
const READ_LIMIT: u64 = 4 * 1024 * 1024;
/// How many hops an `@` import or a script that names a script is followed.
const HOPS: usize = 4;
/// Reader files read, and `.claude/rules` entries listed, in one survey. Past it the survey
/// treats the readers as unread, so a huge tree cannot make it read without end.
const FILE_LIMIT: usize = 1000;
/// How deep `.claude/rules` is listed.
const RULES_DEPTH: usize = 8;
/// The head of an output style file that is searched for its frontmatter `name`.
const STYLE_HEAD: u64 = 16 * 1024;

/// The instruction files read from the project root and from `.claude/`.
const INSTRUCTION_FILES: [&str; 5] = [
    "CLAUDE.md",
    ".claude/CLAUDE.md",
    "CLAUDE.local.md",
    "AGENTS.md",
    "AGENTS.local.md",
];
const RULES_DIR: &str = ".claude/rules";
const SETTINGS_FILES: [&str; 2] = [".claude/settings.json", ".claude/settings.local.json"];
/// The status line's own files (`src/statusline/state.rs`: the Go paths it reads).
const STATUSLINE_FILES: [&str; 4] = [
    ".moai/config/sections/statusline.yaml",
    ".moai/config/sections/llm.yaml",
    ".moai/config/sections/system.yaml",
    ".moai/statusline.local.yaml",
];
/// What a `statusline` reader shows as its path: the status line is a command of the installed
/// program, not a project file.
const STATUSLINE_READER: &str = "moai statusline";
/// A path passed to one of these is a script (REQ-3).
const INTERPRETERS: &[&str] = &[
    "bash", "sh", "zsh", "node", "python3", "ruby", "perl", "deno", "bun",
];
/// A path ending in one of these is a script (REQ-3).
const SCRIPT_SUFFIXES: &[&str] = &[
    ".sh", ".bash", ".zsh", ".js", ".mjs", ".cjs", ".ts", ".py", ".rb", ".ps1",
];
/// Characters that end a name inside a shell command.
const SEPARATORS: &[char] = &[';', '|', '&', '<', '>', '(', ')', '=', ','];

/// One reader of a file: its project path and how it names the file.
pub(crate) type Reader = (String, &'static str);

/// A reader that exists and could not be read, with why.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Unread {
    pub file: String,
    pub reason: String,
}

#[derive(Debug, Default)]
pub(crate) struct Readers {
    /// Target file -> the readers that name it (sorted).
    pub used_by: BTreeMap<String, BTreeSet<Reader>>,
    /// Readers that could not be read, in the order they were met.
    pub unread: Vec<Unread>,
}

/// Where a name points, seen from the project.
#[derive(Debug, PartialEq, Eq)]
enum Place {
    /// A normalised project-relative path.
    Inside(String),
    /// Leaves the project (`..` past the root, `~`, `$HOME`, another absolute path).
    Outside,
    /// Not a path to a file: empty or the project root itself.
    Nothing,
}

/// The readers of `targets`, the project paths of the survey's `go_same` and `go_changed`
/// files. Reads only: nothing is written and no link is followed.
pub(crate) fn collect(project: &Path, targets: &BTreeSet<String>) -> Readers {
    let mut c = Collector {
        project,
        root: project.to_string_lossy().trim_end_matches('/').to_string(),
        targets,
        reads: 0,
        out: Readers::default(),
    };
    c.imports();
    c.settings();
    for file in STATUSLINE_FILES {
        c.hit(file, STATUSLINE_READER, "statusline");
    }
    c.out
}

struct Collector<'a> {
    project: &'a Path,
    /// The project path without a trailing slash, to see an absolute name inside it.
    root: String,
    targets: &'a BTreeSet<String>,
    /// Reader files read so far, against [`FILE_LIMIT`].
    reads: usize,
    out: Readers,
}

impl Collector<'_> {
    /// `reader` names `target`; only a surveyed file is recorded, and never a file by itself.
    fn hit(&mut self, target: &str, reader: &str, kind: &'static str) {
        if target != reader && self.targets.contains(target) {
            self.out
                .used_by
                .entry(target.to_string())
                .or_default()
                .insert((reader.to_string(), kind));
        }
    }

    fn unread(&mut self, file: &str, reason: &str) {
        let entry = Unread {
            file: file.to_string(),
            reason: reason.to_string(),
        };
        if !self.out.unread.contains(&entry) {
            self.out.unread.push(entry);
        }
    }

    /// The text of a reader file: `None` when it does not exist (which blocks nothing) or when
    /// it could not be read (which is recorded as an unread reader).
    fn read(&mut self, rel: &str) -> Option<String> {
        // A secret-shaped name on any component of the path is a reader that is not opened: only
        // the entry's existence is looked at, so no secret ever reaches the survey. The shared
        // rule is the one the rest of the update uses; this module keeps no list of its own.
        if has_secret_component(rel) {
            if fs::symlink_metadata(self.project.join(rel)).is_ok() {
                self.unread(rel, "has a secret-shaped name");
            }
            return None;
        }
        self.reads += 1;
        if self.reads > FILE_LIMIT {
            self.unread(rel, "the survey reads at most 1000 reader files");
            return None;
        }
        match read_bounded_text(self.project, rel, READ_LIMIT) {
            Ok(text) => text,
            Err(_) => {
                let why = why_unread(self.project, rel);
                self.unread(rel, why);
                None
            }
        }
    }

    /// The instruction files, the rules, and the files they import, at most [`HOPS`] deep.
    fn imports(&mut self) {
        let mut level: BTreeSet<String> = INSTRUCTION_FILES.iter().map(|s| s.to_string()).collect();
        level.extend(self.rule_files());
        let mut seen = BTreeSet::new();
        for hop in 0..=HOPS {
            let mut next = BTreeSet::new();
            for file in std::mem::take(&mut level) {
                if !seen.insert(file.clone()) {
                    continue;
                }
                // An import of a folder imports nothing; only a root file that is one is odd.
                if hop > 0 && is_folder(self.project, &file) {
                    continue;
                }
                let Some(text) = self.read(&file) else {
                    continue;
                };
                let base = folder_of(&file);
                for name in import_names(&text) {
                    let primary = locate(&self.root, base, &name);
                    // A file in a folder (`.claude/rules/x.md`) that names a project-rooted path
                    // is also counted as naming it from the project root: Claude Code reads the
                    // path from the importing file, and keeping one file more is the safe error.
                    let rooted = match (&primary, base.is_empty()) {
                        (Place::Inside(first), false) => match locate(&self.root, "", &name) {
                            Place::Inside(path) if path != *first => Some(path),
                            _ => None,
                        },
                        _ => None,
                    };
                    match primary {
                        Place::Inside(path) => {
                            self.hit(&path, &file, "import");
                            if hop < HOPS {
                                next.insert(path);
                            }
                        }
                        Place::Outside => self.unread(&file, "imports a path outside the project"),
                        Place::Nothing => {}
                    }
                    if let Some(path) = rooted {
                        self.hit(&path, &file, "import");
                        if hop < HOPS {
                            next.insert(path);
                        }
                    }
                }
            }
            level = next;
        }
    }

    /// The `.md` files under `.claude/rules`, sorted. A link on the way, or a listing cut at its
    /// limits, makes `.claude/rules` an unread reader.
    fn rule_files(&mut self) -> Vec<String> {
        if linked_component(self.project, RULES_DIR) {
            self.unread(RULES_DIR, "has a link on its path");
            return Vec::new();
        }
        if !is_folder(self.project, RULES_DIR) {
            return Vec::new();
        }
        let mut files = Vec::new();
        let mut cut = false;
        let mut stack = vec![(RULES_DIR.to_string(), 0usize)];
        let mut listed = 0usize;
        'walk: while let Some((dir, depth)) = stack.pop() {
            let Ok(entries) = fs::read_dir(self.project.join(&dir)) else {
                continue;
            };
            let mut names: Vec<String> = entries
                .filter_map(|e| e.ok())
                .filter_map(|e| e.file_name().into_string().ok())
                .collect();
            names.sort();
            for name in names {
                listed += 1;
                if listed > FILE_LIMIT {
                    cut = true;
                    break 'walk;
                }
                let child = format!("{dir}/{name}");
                let Ok(meta) = fs::symlink_metadata(self.project.join(&child)) else {
                    continue;
                };
                if meta.file_type().is_symlink() {
                    // A link is never followed. A file link is a reader that cannot be read; a
                    // folder link hides files, so it stops the survey the same way.
                    if name.ends_with(".md") {
                        files.push(child);
                    } else if fs::metadata(self.project.join(&child)).is_ok_and(|m| m.is_dir()) {
                        self.unread(&child, "has a link on its path");
                    }
                } else if meta.is_dir() {
                    if depth < RULES_DEPTH {
                        stack.push((child, depth + 1));
                    } else {
                        cut = true;
                    }
                } else if name.ends_with(".md") {
                    files.push(child);
                }
            }
        }
        if cut {
            self.unread(RULES_DIR, "the listing is cut at its limit");
        }
        files.sort();
        files
    }

    /// The two settings files: every string outside `permissions` is a name, `outputStyle`
    /// names a style file, and the scripts the strings name are read in turn.
    fn settings(&mut self) {
        let mut scripts: Vec<(String, usize)> = Vec::new();
        for rel in SETTINGS_FILES {
            let Some(text) = self.read(rel) else {
                continue;
            };
            let Ok(value) = serde_json::from_str::<Value>(&text) else {
                self.unread(rel, "is not valid JSON");
                continue;
            };
            let mut strings = Vec::new();
            collect_strings(&value, true, &mut strings);
            for s in &strings {
                self.command(rel, s, "settings", true, 1, &mut scripts);
            }
            if let Some(style) = value.get("outputStyle").and_then(Value::as_str) {
                self.style(rel, style);
            }
        }
        self.follow(scripts);
    }

    /// The scripts that settings named, and the scripts those name, at most [`HOPS`] deep.
    fn follow(&mut self, mut queue: Vec<(String, usize)>) {
        let mut seen = BTreeSet::new();
        let mut i = 0;
        while i < queue.len() {
            let (script, hop) = queue[i].clone();
            i += 1;
            if !seen.insert(script.clone()) || is_folder(self.project, &script) {
                continue;
            }
            let Some(text) = self.read(&script) else {
                continue;
            };
            self.command(&script, &text, "script", false, hop + 1, &mut queue);
        }
    }

    /// The names in one string of a settings file or in the text of a script. A name inside the
    /// project that is a surveyed file is recorded; a script inside the project is queued for
    /// reading when its hop is within [`HOPS`]; a script outside the project makes `reader` an
    /// unread reader when `in_settings` (REQ-3), and is skipped inside a script text.
    fn command(
        &mut self,
        reader: &str,
        text: &str,
        kind: &'static str,
        in_settings: bool,
        hop: usize,
        queue: &mut Vec<(String, usize)>,
    ) {
        let mut after_interpreter = false;
        for chunk in text.split_whitespace() {
            let bare: String = chunk
                .chars()
                .filter(|c| !matches!(c, '"' | '\'' | '`'))
                .collect();
            if bare.is_empty() {
                continue;
            }
            // The path passed to an interpreter, after its flags, is a script.
            let mut script_argument = false;
            if after_interpreter && !bare.starts_with('-') {
                script_argument = bare.contains('/') || bare.contains('.');
                after_interpreter = false;
            }
            let program = bare.rsplit('/').next().unwrap_or(&bare);
            if INTERPRETERS.contains(&program) {
                after_interpreter = true;
            }
            let mut pieces = vec![bare.clone()];
            pieces.extend(
                bare.split(SEPARATORS)
                    .filter(|p| !p.is_empty() && *p != bare)
                    .map(str::to_string),
            );
            for piece in pieces {
                if piece.contains("://") {
                    continue;
                }
                let is_script =
                    script_argument || SCRIPT_SUFFIXES.iter().any(|suffix| piece.ends_with(suffix));
                match locate(&self.root, "", &piece) {
                    Place::Inside(path) => {
                        self.hit(&path, reader, kind);
                        if is_script && hop <= HOPS {
                            queue.push((path, hop));
                        }
                    }
                    Place::Outside if is_script && in_settings => {
                        self.unread(reader, "names a script outside the project");
                    }
                    Place::Outside | Place::Nothing => {}
                }
            }
        }
    }

    /// `outputStyle` names the style file whose file name, or whose frontmatter `name`, equals
    /// it (without case), or whose path tail does (`dir/name`). Over-keeping is the safe error.
    fn style(&mut self, reader: &str, style: &str) {
        let want = style.trim().to_lowercase();
        if want.is_empty() {
            return;
        }
        let styles: Vec<String> = self
            .targets
            .iter()
            .filter(|t| t.split('/').any(|part| part == "output-styles"))
            .cloned()
            .collect();
        for target in styles {
            let lower = Path::new(&target.to_lowercase())
                .with_extension("")
                .to_string_lossy()
                .into_owned();
            let stem = lower.rsplit('/').next().unwrap_or(&lower).to_string();
            let named = stem == want
                || lower.ends_with(&format!("/{want}"))
                || frontmatter_name(self.project, &target)
                    .is_some_and(|name| name.to_lowercase() == want);
            if named {
                self.hit(&target, reader, "settings");
            }
        }
    }
}

/// Every string of a settings value outside the top-level `permissions`: values and keys.
fn collect_strings(value: &Value, top: bool, out: &mut Vec<String>) {
    match value {
        Value::String(s) => out.push(s.clone()),
        Value::Array(items) => items.iter().for_each(|v| collect_strings(v, false, out)),
        Value::Object(map) => {
            for (key, v) in map {
                if top && key == "permissions" {
                    continue;
                }
                out.push(key.clone());
                collect_strings(v, false, out);
            }
        }
        _ => {}
    }
}

/// The `name:` of an output style's frontmatter, read from the head of the file.
fn frontmatter_name(project: &Path, rel: &str) -> Option<String> {
    if linked_component(project, rel) {
        return None;
    }
    let path = project.join(rel);
    let meta = fs::symlink_metadata(&path).ok().filter(|m| m.is_file())?;
    let mut buf = Vec::new();
    fs::File::open(&path)
        .ok()?
        .take(meta.len().min(STYLE_HEAD))
        .read_to_end(&mut buf)
        .ok()?;
    let text = String::from_utf8_lossy(&buf);
    let mut lines = text.lines();
    (lines.next()?.trim_end() == "---").then_some(())?;
    for line in lines {
        let line = line.trim_end();
        if line == "---" {
            return None;
        }
        if let Some(value) = line.strip_prefix("name:") {
            let value = value.trim().trim_matches(|c| c == '"' || c == '\'');
            return (!value.is_empty()).then(|| value.to_string());
        }
    }
    None
}

fn is_folder(project: &Path, rel: &str) -> bool {
    fs::symlink_metadata(project.join(rel)).is_ok_and(|m| m.is_dir())
}

/// The folder part of a project path, `""` for a file in the root.
fn folder_of(rel: &str) -> &str {
    rel.rsplit_once('/').map_or("", |(dir, _)| dir)
}

/// Why `read_bounded_text` refused `rel`, in words with no content of the file.
fn why_unread(project: &Path, rel: &str) -> &'static str {
    if linked_component(project, rel) {
        return "has a link on its path";
    }
    match fs::symlink_metadata(project.join(rel)) {
        Ok(m) if !m.is_file() => "is not a regular file",
        Ok(m) if m.len() > READ_LIMIT => "is larger than 4 MiB",
        _ => "cannot be read as UTF-8 text",
    }
}

/// Where `name` points, read from the folder `base` of a project file (`""` for the root). The
/// project folder itself, as `$CLAUDE_PROJECT_DIR` or as the absolute path, is the root.
fn locate(root: &str, base: &str, name: &str) -> Place {
    let mut name = name;
    for var in ["$CLAUDE_PROJECT_DIR", "${CLAUDE_PROJECT_DIR}"] {
        if let Some(rest) = name.strip_prefix(var) {
            if rest.is_empty() || rest.starts_with('/') {
                name = rest.trim_start_matches('/');
                break;
            }
        }
    }
    if name.is_empty() {
        return Place::Nothing;
    }
    if name.starts_with('~') || name.starts_with('$') {
        return Place::Outside;
    }
    let joined = if name.starts_with('/') {
        match name
            .strip_prefix(root)
            .filter(|rest| rest.is_empty() || rest.starts_with('/'))
        {
            Some(rest) => rest.to_string(),
            None => return Place::Outside,
        }
    } else if base.is_empty() {
        name.to_string()
    } else {
        format!("{base}/{name}")
    };
    let mut parts: Vec<&str> = Vec::new();
    for part in joined.split('/') {
        match part {
            "" | "." => {}
            ".." => {
                if parts.pop().is_none() {
                    return Place::Outside;
                }
            }
            part => parts.push(part),
        }
    }
    if parts.is_empty() {
        Place::Nothing
    } else {
        Place::Inside(parts.join("/"))
    }
}

/// The `@path` imports of an instruction file: a `@` at the start of a line or after
/// whitespace, the path up to the next whitespace (`\ ` is a space in it). Code spans and
/// fenced blocks are skipped, and a quoted path is not an import.
fn import_names(text: &str) -> Vec<String> {
    let mut names = Vec::new();
    let mut fence: Option<(char, usize)> = None;
    for line in text.lines() {
        let indent = line.len() - line.trim_start_matches(' ').len();
        let body = &line[indent..];
        if let Some((mark, size)) = fence {
            let run = body.chars().take_while(|&c| c == mark).count();
            if indent < 4 && run >= size && body[run..].trim().is_empty() {
                fence = None;
            }
            continue;
        }
        let mark = body.chars().next().filter(|c| matches!(c, '`' | '~'));
        if let Some(mark) = mark.filter(|_| indent < 4) {
            let run = body.chars().take_while(|&c| c == mark).count();
            // A backtick fence's info string holds no backtick: this is a code span otherwise.
            if run >= 3 && (mark == '~' || !body[run..].contains('`')) {
                fence = Some((mark, run));
                continue;
            }
        }
        at_names(&blank_code_spans(line), &mut names);
    }
    names
}

/// The line with every inline code span (a backtick run to the next run of the same length)
/// replaced by one space. Linear: the closing run of each opening run is found once.
fn blank_code_spans(line: &str) -> String {
    if !line.contains('`') {
        return line.to_string();
    }
    let chars: Vec<char> = line.chars().collect();
    // The backtick runs as (start, end).
    let mut runs: Vec<(usize, usize)> = Vec::new();
    let mut i = 0;
    while i < chars.len() {
        if chars[i] == '`' {
            let start = i;
            while i < chars.len() && chars[i] == '`' {
                i += 1;
            }
            runs.push((start, i));
        } else {
            i += 1;
        }
    }
    // The next run of the same length after each run.
    let mut next_same: Vec<Option<usize>> = vec![None; runs.len()];
    let mut latest: HashMap<usize, usize> = HashMap::new();
    for (idx, (start, end)) in runs.iter().enumerate().rev() {
        next_same[idx] = latest.get(&(end - start)).copied();
        latest.insert(end - start, idx);
    }
    let mut blank = vec![false; chars.len()];
    let mut idx = 0;
    while idx < runs.len() {
        match next_same[idx] {
            Some(close) => {
                for slot in &mut blank[runs[idx].0..runs[close].1] {
                    *slot = true;
                }
                idx = close + 1;
            }
            None => idx += 1,
        }
    }
    let mut out = String::with_capacity(line.len());
    let mut in_span = false;
    for (c, blanked) in chars.iter().zip(&blank) {
        if *blanked {
            if !in_span {
                out.push(' ');
            }
            in_span = true;
        } else {
            out.push(*c);
            in_span = false;
        }
    }
    out
}

/// The names after a `@` that starts a line or follows whitespace. A name is also kept without
/// the sentence punctuation that may trail it.
fn at_names(line: &str, names: &mut Vec<String>) {
    let chars: Vec<char> = line.chars().collect();
    let mut i = 0;
    while i < chars.len() {
        let boundary = i == 0 || chars[i - 1].is_whitespace();
        if chars[i] != '@' || !boundary {
            i += 1;
            continue;
        }
        let mut j = i + 1;
        let mut name = String::new();
        while j < chars.len() {
            if chars[j] == '\\' && chars.get(j + 1) == Some(&' ') {
                name.push(' ');
                j += 2;
            } else if chars[j].is_whitespace() || chars[j] == '\\' {
                break;
            } else {
                name.push(chars[j]);
                j += 1;
            }
        }
        i = j.max(i + 1);
        if name.is_empty() || name.starts_with(['"', '\'']) {
            continue;
        }
        let trimmed = name.trim_end_matches(['.', ',', ';', ':', '!', '?', ')', ']', '}']);
        if !trimmed.is_empty() && trimmed != name {
            names.push(trimmed.to_string());
        }
        names.push(name);
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    /// REQ-2: an import is a `@` at a line start or after whitespace; a code span, a fenced
    /// block, a quoted path and a `@` inside a word are not imports; `\ ` is a space.
    #[test]
    fn imports_skip_code_quotes_and_words() {
        let text = "@first.md\nsee @second.md, and `@span.md` or ``@ticks `x` .md``\n\
                    mail me@example.com or \"@quoted.md\" or @\"alt.md\"\n\
                    ```text\n@fenced.md\n```\n~~~\n@tilde.md\n~~~\n    ```\n@after.md\n\
                    @with\\ space.md\n";
        let names = import_names(text);
        for want in ["first.md", "second.md", "after.md", "with space.md"] {
            assert!(names.iter().any(|n| n == want), "{want}: {names:?}");
        }
        for not in [
            "span.md",
            "fenced.md",
            "tilde.md",
            "quoted.md",
            "alt.md",
            "example.com",
        ] {
            assert!(!names.iter().any(|n| n.contains(not)), "{not}: {names:?}");
        }
    }

    /// An unclosed fence hides the rest of the file, and a fence closes only on a long enough
    /// run of its own mark.
    #[test]
    fn a_fence_closes_on_its_own_mark_only() {
        let names = import_names("````\n```\n@inside.md\n````\n@outside.md\n");
        assert_eq!(names, ["outside.md"]);
        assert!(import_names("```\n@never.md\n").is_empty());
    }

    /// Many backtick runs of different lengths stay linear, and an unmatched run is text.
    #[test]
    fn code_spans_pair_equal_runs() {
        assert_eq!(blank_code_spans("a `b` c"), "a   c");
        assert_eq!(blank_code_spans("a ``b ` c`` d"), "a   d");
        assert_eq!(blank_code_spans("a ` b"), "a ` b");
        let many: String = (1..2000).map(|n| format!("{} x ", "`".repeat(n))).collect();
        assert!(!blank_code_spans(&many).is_empty());
    }

    /// REQ-3: where a name points. `..` past the root, `~`, `$HOME` and another absolute path
    /// leave the project; the project variable and the absolute project path stay in it.
    #[test]
    fn names_are_placed_inside_or_outside_the_project() {
        let root = "/work/project";
        let inside = |p: &str| Place::Inside(p.to_string());
        assert_eq!(locate(root, "docs", "../a/b.md"), inside("a/b.md"));
        assert_eq!(locate(root, "", "./x/./y.md"), inside("x/y.md"));
        assert_eq!(locate(root, "docs", "../../a.md"), Place::Outside);
        assert_eq!(locate(root, "", "~/a.md"), Place::Outside);
        assert_eq!(locate(root, "", "$HOME/.tool/bin"), Place::Outside);
        assert_eq!(locate(root, "", "/etc/hosts"), Place::Outside);
        assert_eq!(locate(root, "", "/work/project2/a.md"), Place::Outside);
        assert_eq!(locate(root, "", "/work/project/a/b.md"), inside("a/b.md"));
        assert_eq!(
            locate(root, "", "$CLAUDE_PROJECT_DIR/.claude/x.sh"),
            inside(".claude/x.sh")
        );
        assert_eq!(
            locate(root, "", "${CLAUDE_PROJECT_DIR}/x.sh"),
            inside("x.sh")
        );
        assert_eq!(locate(root, "", "/work/project"), Place::Nothing);
    }
}
