//! SPEC-20261123-hook-bash-boundary: the judgment of `moai hook bash-boundary` (H2, plan §4).
//!
//! The handler reads the text of one Bash command and answers a deny with a class, or nothing.
//! It splits the text into commands with a small shell lexer (quotes, escapes, substitutions,
//! here-documents, operators, groups), follows only the plain literal forms of `cd`, `pushd` and
//! `popd` (a stack index, `CDPATH` or a bare `pushd` leaves the folder unreadable), looks through
//! the commands that run another command (the closed table of wrappers, `xargs`, `find -exec`,
//! shell strings, `eval`, `trap`, the git subcommands that run a program), treats a word that
//! changes how a command word or a `cd` operand is looked up (`alias`, `hash`, `enable`, the
//! `cdable_vars` and `cdpath` options and variables) as unreadable (D-12), and judges
//! four things: a delete outside the session checkout and the temporary folders, a git state
//! change of another checkout (or a branch-moving command in the shared primary checkout), a
//! direct change of a protected branch ref, and a skipped or replaced commit hook. The one
//! principle: a value a check needs that the text does not hold (an expansion, a substitution, a
//! glob root, an alias, a computed command word) is denied. Variables are never followed.
//!
//! It reads the file system (to find checkouts, resolve links and read `HEAD`) and nothing else:
//! no process, no network, no environment variable and no write; the caller passes in what it
//! needs.

use serde_json::{json, Map, Value};
use std::collections::BTreeMap;
use std::fs;
use std::path::{Component, Path, PathBuf};

/// What the handler needs besides the hook input.
pub struct Env<'a> {
    /// The folder the entry found `.moai/project.yaml` in.
    pub project_root: &'a Path,
    /// `CLAUDE_PROJECT_DIR`, when set.
    pub project_dir: Option<&'a Path>,
    pub home: Option<&'a Path>,
    /// The working folder of the hook process, used when the input has no `cwd`.
    pub process_cwd: &'a Path,
}

const MAX_BYTES: usize = 256 * 1024;
const MAX_DEPTH: usize = 8;
const MAX_COMMANDS: usize = 2048;
const MAX_BRACE: usize = 64;
/// The most bytes read from `.moai/project.yaml` (the real file is a few KiB). A file over the
/// limit is read as absent, so the protected list is then `main` and `master` only (D-4).
const PROJECT_FILE_LIMIT: u64 = 1024 * 1024;

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum Class {
    DeleteOutside,
    GitState,
    ProtectedRef,
    HookSkip,
    Unreadable,
}

impl Class {
    fn word(self) -> &'static str {
        match self {
            Class::DeleteOutside => "delete-outside",
            Class::GitState => "git-state",
            Class::ProtectedRef => "protected-ref",
            Class::HookSkip => "hook-skip",
            Class::Unreadable => "unreadable",
        }
    }

    fn reason(self) -> String {
        let text = match self {
            Class::DeleteOutside => "a delete target is outside this session's checkout and the temporary folders, or is a checkout root, a .git folder or another checkout. Delete only inside your own worktree or a temporary folder, by a literal path.",
            Class::GitState => "this git command changes the state of a checkout that is not this session's own, or moves a branch in the shared primary checkout. Work in your own worktree (cd into it, or use git -C with its literal path).",
            Class::ProtectedRef => "an integration or protected branch is changed by hand here. Land through `moai card land`; other branches can be changed freely.",
            Class::HookSkip => "commit hooks are skipped or replaced here. Fix what the hook reports and run the command again without the skip.",
            Class::Unreadable => "a value this check needs cannot be read from the command text (an expansion, a substitution, a glob root, an alias, a computed command word, or syntax that cannot be split). Write the path, branch or command literally.",
        };
        format!("moai-boundary/{}: {}", self.word(), text)
    }
}

type Verdict = Result<(), Class>;

// ------------------------------------------------------------------ entry

/// The decision for one hook input: a deny object, or `None` for no decision.
pub fn decide(input: &Map<String, Value>, env: &Env) -> Option<Value> {
    if let Some(tool) = input.get("tool_name") {
        if tool.as_str() != Some("Bash") {
            return None;
        }
    }
    let command = input.get("tool_input")?.get("command")?.as_str()?;
    let class = judge(command, input, env).err()?;
    Some(json!({
        "hookSpecificOutput": {
            "hookEventName": "PreToolUse",
            "permissionDecision": "deny",
            "permissionDecisionReason": class.reason(),
        }
    }))
}

fn judge(command: &str, input: &Map<String, Value>, env: &Env) -> Verdict {
    if command.len() > MAX_BYTES {
        return Err(Class::Unreadable);
    }
    if !needs_walk(command) {
        return Ok(());
    }
    let world = World::build(input, env);
    let mut walker = Walker {
        world: &world,
        count: 0,
        others: None,
    };
    let mut state = State {
        cwd: world.cwd.clone(),
        stack: Vec::new(),
        git_env: BTreeMap::new(),
    };
    walker.walk(command, &mut state, 0)
}

/// Words the checks react to. The scan runs on the command with quotes and backslashes removed,
/// so a word split by quoting is still seen; any expansion marker also needs the walk.
const TRIGGERS: &[&str] = &[
    "rm", "unlink", "shred", "find", "git", "xargs", "eval", "exec", "sh", "sudo", "doas", "env",
    "command", "builtin", "nohup", "nice", "time", "parallel", "watch", "busybox", "setsid",
];
const TRIGGERS_MORE: &[&str] = &[
    "stdbuf",
    "ionice",
    "caffeinate",
    // The words that change how a command word or a `cd` operand is looked up (D-12).
    "alias",
    "enable",
    // `set`, `setopt`, `unsetopt` and `shopt` read their option words from the table of D-17: a
    // line with one of them is walked, because a word outside the table is unreadable.
    "set",
    "emulate",
    "cdpath",
    "cdable",
];

/// Whether the command holds a word the checks react to (D-11). A glob character or a brace can
/// spell a command word out of nothing a trigger word names, and a word that starts with `=`
/// is a command word in some shells, so each of them also needs the walk.
pub fn needs_walk(command: &str) -> bool {
    if command.contains(['$', '`', '*', '?', '[', '{'])
        || command.contains("<(")
        || command.contains(">(")
    {
        return true;
    }
    let mut before = ' ';
    for c in command.chars() {
        if c == '='
            && matches!(
                before,
                ' ' | '\t' | '\n' | ';' | '&' | '|' | '(' | '"' | '\''
            )
        {
            return true;
        }
        before = c;
    }
    let flat: String = command
        .chars()
        .filter(|c| !matches!(c, '\'' | '"' | '\\'))
        .flat_map(char::to_lowercase)
        .collect();
    TRIGGERS
        .iter()
        .chain(TRIGGERS_MORE.iter())
        .any(|t| flat.contains(t))
}

// ------------------------------------------------------------------ world

struct Checkout {
    root: PathBuf,
    /// The folder that holds this checkout's `HEAD`.
    gitdir: PathBuf,
    common: PathBuf,
    primary: bool,
}

struct World {
    cwd: Option<PathBuf>,
    own: Option<Checkout>,
    home: Option<PathBuf>,
    protected: Vec<String>,
    /// Resolved boundary roots that are never a target themselves.
    roots: Vec<PathBuf>,
}

impl World {
    fn build(input: &Map<String, Value>, env: &Env) -> World {
        let cwd_in = input
            .get("cwd")
            .and_then(Value::as_str)
            .map(PathBuf::from)
            .filter(|p| p.is_absolute())
            .unwrap_or_else(|| env.process_cwd.to_path_buf());
        let cwd = if cwd_in.is_absolute() {
            resolve(&cwd_in, true)
        } else {
            None
        };
        let here = cwd.as_deref().and_then(checkout_of);
        let project = env.project_dir.and_then(checkout_of);
        let own = match (here, project) {
            (Some(h), Some(p)) if h.common != p.common => Some(p),
            (h, _) => h,
        };
        let mut protected = vec!["main".to_owned(), "master".to_owned()];
        if let Some(name) = integration_branch(env.project_root) {
            protected.push(name.to_lowercase());
        }
        let mut roots = Vec::new();
        for raw in ["/tmp", "/private/tmp", "/var/tmp", "/private/var/tmp"] {
            if let Some(r) = resolve(Path::new(raw), true) {
                if !roots.contains(&r) {
                    roots.push(r);
                }
            }
        }
        if let Some(s) = input
            .get("scratchpad_dir")
            .and_then(Value::as_str)
            .map(PathBuf::from)
            .filter(|p| p.is_absolute())
            .and_then(|p| resolve(&p, true))
        {
            roots.push(s);
        }
        World {
            cwd,
            own,
            home: env.home.map(Path::to_path_buf),
            protected,
            roots,
        }
    }

    fn is_protected(&self, branch: &str) -> bool {
        let lower = branch.to_lowercase();
        self.protected.contains(&lower)
    }
}

fn integration_branch(project_root: &Path) -> Option<String> {
    use std::io::Read;
    let file = fs::File::open(project_root.join(".moai/project.yaml")).ok()?;
    let mut bytes = Vec::new();
    file.take(PROJECT_FILE_LIMIT + 1)
        .read_to_end(&mut bytes)
        .ok()?;
    if bytes.len() as u64 > PROJECT_FILE_LIMIT {
        return None;
    }
    let value: Value = serde_json::from_slice(&bytes).ok()?;
    let name = value
        .get("work_flow")?
        .get("integration_branch")?
        .as_str()?;
    (!name.is_empty()).then(|| name.to_owned())
}

/// Resolves links component by component, as the kernel does for an absolute path. The last
/// component is followed only when `follow_last` is set (`rm link` removes the link itself).
fn resolve(path: &Path, follow_last: bool) -> Option<PathBuf> {
    let mut links = 0u32;
    resolve_inner(path, follow_last, &mut links)
}

fn resolve_inner(path: &Path, follow_last: bool, links: &mut u32) -> Option<PathBuf> {
    let comps: Vec<Component> = path.components().collect();
    let mut cur = PathBuf::from("/");
    for (k, comp) in comps.iter().enumerate() {
        match comp {
            Component::RootDir | Component::Prefix(_) => cur = PathBuf::from("/"),
            Component::CurDir => {}
            Component::ParentDir => {
                cur.pop();
            }
            Component::Normal(name) => {
                let next = cur.join(name);
                if k + 1 == comps.len() && !follow_last {
                    cur = next;
                    continue;
                }
                match fs::symlink_metadata(&next) {
                    Ok(m) if m.file_type().is_symlink() => {
                        *links += 1;
                        if *links > 32 {
                            return None;
                        }
                        let target = fs::read_link(&next).ok()?;
                        let base = if target.is_absolute() {
                            target
                        } else {
                            cur.join(target)
                        };
                        cur = resolve_inner(&base, true, links)?;
                    }
                    _ => cur = next,
                }
            }
        }
    }
    Some(cur)
}

/// The checkout holding `path` (the nearest ancestor with a `.git` entry), read without a process.
fn checkout_of(path: &Path) -> Option<Checkout> {
    let start = resolve(path, true)?;
    let mut cur = Some(start.as_path());
    while let Some(dir) = cur {
        let dotgit = dir.join(".git");
        if let Ok(meta) = fs::symlink_metadata(&dotgit) {
            if meta.is_dir() {
                return Some(Checkout {
                    root: dir.to_path_buf(),
                    gitdir: dotgit.clone(),
                    common: dotgit,
                    primary: true,
                });
            }
            if meta.is_file() {
                let text = fs::read_to_string(&dotgit).ok()?;
                let line = text.lines().next()?;
                let rel = line.strip_prefix("gitdir:")?.trim();
                let gitdir = resolve(&dir.join(rel), true)?;
                let common = match fs::read_to_string(gitdir.join("commondir")) {
                    Ok(c) => resolve(&gitdir.join(c.trim()), true)?,
                    Err(_) => gitdir.clone(),
                };
                return Some(Checkout {
                    root: dir.to_path_buf(),
                    gitdir,
                    common,
                    primary: false,
                });
            }
        }
        cur = dir.parent();
    }
    None
}

/// The branch `HEAD` of a checkout names, if it is on one.
fn head_branch(checkout: &Checkout) -> Option<String> {
    let text = fs::read_to_string(checkout.gitdir.join("HEAD")).ok()?;
    let name = text.lines().next()?.strip_prefix("ref: refs/heads/")?;
    Some(name.trim().to_owned())
}

/// Every checkout root of the repository: the primary and each registered linked worktree.
fn all_checkouts(common: &Path) -> Vec<PathBuf> {
    let mut out = Vec::new();
    if common.file_name().is_some_and(|n| n == ".git") {
        if let Some(parent) = common.parent() {
            out.push(parent.to_path_buf());
        }
    }
    if let Ok(entries) = fs::read_dir(common.join("worktrees")) {
        for entry in entries.flatten() {
            if let Ok(text) = fs::read_to_string(entry.path().join("gitdir")) {
                if let Some(parent) = Path::new(text.trim()).parent() {
                    out.push(parent.to_path_buf());
                }
            }
        }
    }
    out
}

/// The root of a per-user temporary folder (`/private/var/folders/<a>/<b>/T`) holding `path`.
fn folders_root(path: &Path) -> Option<PathBuf> {
    let comps: Vec<_> = path.components().collect();
    if comps.len() < 7 {
        return None;
    }
    let names: Vec<_> = comps
        .iter()
        .map(|c| c.as_os_str().to_str().unwrap_or(""))
        .collect();
    if names[1] == "private" && names[2] == "var" && names[3] == "folders" && names[6] == "T" {
        Some(comps[..7].iter().collect())
    } else {
        None
    }
}

// ------------------------------------------------------------------ lexer

#[derive(Clone, Default)]
struct Word {
    /// Each character with whether it is bare (not quoted, not escaped).
    chars: Vec<(char, bool)>,
    /// The value is decided by an expansion the text does not hold.
    dynamic: bool,
    /// Scripts run by a substitution inside the word.
    subs: Vec<String>,
}

impl Word {
    fn text(&self) -> String {
        self.chars.iter().map(|&(c, _)| c).collect()
    }

    fn literal(text: &str) -> Word {
        Word {
            chars: text.chars().map(|c| (c, false)).collect(),
            ..Word::default()
        }
    }

    /// A readable option word: a dash first (quoted or not, the program sees the same), and
    /// not a lone dash.
    fn is_option(&self) -> bool {
        !self.dynamic && self.chars.len() > 1 && self.chars[0].0 == '-'
    }
}

struct Heredoc {
    body: String,
    quoted: bool,
}

#[derive(Default)]
struct Cmd {
    words: Vec<Word>,
    herestr: Vec<Word>,
    redirs: Vec<Word>,
    heredocs: Vec<usize>,
    stdin_redirect: bool,
    piped_in: bool,
}

enum Item {
    Cmd(Cmd),
    Open,
    Close,
}

struct Script {
    items: Vec<Item>,
    bodies: Vec<Heredoc>,
    unsupported: bool,
}

struct Lexer {
    s: Vec<char>,
    i: usize,
    unsupported: bool,
    items: Vec<Item>,
    bodies: Vec<Heredoc>,
    pending: Vec<usize>,
    /// The delimiter and the tab-strip flag of each pending here-document.
    heredoc_meta: Vec<(String, bool)>,
    cur: Cmd,
    word: Option<Word>,
}

fn lex(src: &str) -> Script {
    let mut lx = Lexer {
        s: src.chars().collect(),
        i: 0,
        unsupported: false,
        items: Vec::new(),
        bodies: Vec::new(),
        pending: Vec::new(),
        heredoc_meta: Vec::new(),
        cur: Cmd::default(),
        word: None,
    };
    lx.run();
    Script {
        items: lx.items,
        bodies: lx.bodies,
        unsupported: lx.unsupported,
    }
}

const WORD_END: [char; 10] = [' ', '\t', '\n', ';', '&', '|', '(', ')', '<', '>'];

impl Lexer {
    fn peek(&self, k: usize) -> Option<char> {
        self.s.get(self.i + k).copied()
    }

    fn finish_word(&mut self) {
        if let Some(w) = self.word.take() {
            self.cur.words.push(w);
        }
    }

    fn end_cmd(&mut self, pipe_next: bool) {
        self.finish_word();
        let done = std::mem::take(&mut self.cur);
        if !(done.words.is_empty()
            && done.herestr.is_empty()
            && done.redirs.is_empty()
            && done.heredocs.is_empty())
        {
            self.items.push(Item::Cmd(done));
        }
        self.cur.piped_in = pipe_next;
    }

    fn run(&mut self) {
        while self.i < self.s.len() {
            let c = self.s[self.i];
            match c {
                ' ' | '\t' => {
                    self.finish_word();
                    self.i += 1;
                }
                '\n' => {
                    self.end_cmd(false);
                    self.i += 1;
                    self.read_heredoc_bodies();
                }
                '\\' if self.peek(1) == Some('\n') && self.word.is_none() => self.i += 2,
                '#' if self.word.is_none() => {
                    while self.i < self.s.len() && self.s[self.i] != '\n' {
                        self.i += 1;
                    }
                }
                ';' => {
                    if self.peek(1) == Some(';') {
                        self.unsupported = true;
                    }
                    self.end_cmd(false);
                    self.i += if self.peek(1) == Some(';') { 2 } else { 1 };
                }
                '&' => {
                    if self.peek(1) == Some('>') {
                        self.i += 1;
                        self.redirect();
                    } else if self.peek(1) == Some('&') {
                        self.end_cmd(false);
                        self.i += 2;
                    } else {
                        self.end_cmd(false);
                        self.i += 1;
                    }
                }
                '|' => match self.peek(1) {
                    Some('|') => {
                        self.end_cmd(false);
                        self.i += 2;
                    }
                    Some('&') => {
                        self.end_cmd(true);
                        self.i += 2;
                    }
                    _ => {
                        self.end_cmd(true);
                        self.i += 1;
                    }
                },
                '(' => {
                    if self.word.is_some() || self.peek(1) == Some('(') {
                        self.unsupported = true;
                    }
                    self.end_cmd(false);
                    self.items.push(Item::Open);
                    self.i += 1;
                }
                ')' => {
                    self.end_cmd(false);
                    self.items.push(Item::Close);
                    self.i += 1;
                }
                '<' | '>' => self.redirect(),
                _ => {
                    let w = self.read_word();
                    self.word = Some(w);
                }
            }
        }
        self.end_cmd(false);
    }

    fn skip_blanks(&mut self) {
        while matches!(self.peek(0), Some(' ' | '\t')) {
            self.i += 1;
        }
    }

    fn redirect(&mut self) {
        // A bare run of digits right before the operator is a file descriptor.
        let is_fd = self.word.as_ref().is_some_and(|w| {
            !w.chars.is_empty() && w.chars.iter().all(|&(c, b)| b && c.is_ascii_digit())
        });
        if is_fd {
            self.word = None;
        } else {
            self.finish_word();
        }
        let c = self.s[self.i];
        if self.peek(1) == Some('(') {
            self.i += 1;
            let w = self.read_process_sub();
            self.cur.words.push(w);
            return;
        }
        if c == '<' && self.peek(1) == Some('<') && self.peek(2) == Some('<') {
            self.i += 3;
            self.skip_blanks();
            let w = self.read_word();
            self.cur.herestr.push(w);
            return;
        }
        if c == '<' && self.peek(1) == Some('<') {
            self.i += 2;
            let strip = self.peek(0) == Some('-');
            if strip {
                self.i += 1;
            }
            self.skip_blanks();
            let delim = self.read_word();
            let quoted = delim.chars.iter().any(|&(_, bare)| !bare);
            let idx = self.bodies.len();
            self.bodies.push(Heredoc {
                body: String::new(),
                quoted,
            });
            self.cur.heredocs.push(idx);
            self.pending.push(idx);
            self.heredoc_meta.push((delim.text(), strip));
            return;
        }
        // Other operators: `<`, `<>`, `<&`, `>`, `>>`, `>&`, `>|`.
        self.i += 1;
        if matches!(self.peek(0), Some('>' | '&' | '|')) || (c == '<' && self.peek(0) == Some('>'))
        {
            self.i += 1;
        }
        if c == '<' {
            self.cur.stdin_redirect = true;
        }
        self.skip_blanks();
        if self.peek(0).is_some_and(|n| !WORD_END.contains(&n)) {
            let w = self.read_word();
            self.cur.redirs.push(w);
        }
    }

    fn read_heredoc_bodies(&mut self) {
        let pending = std::mem::take(&mut self.pending);
        let metas = std::mem::take(&mut self.heredoc_meta);
        for (idx, (delim, strip)) in pending.into_iter().zip(metas) {
            let mut body = String::new();
            while self.i < self.s.len() {
                let start = self.i;
                while self.i < self.s.len() && self.s[self.i] != '\n' {
                    self.i += 1;
                }
                let line: String = self.s[start..self.i].iter().collect();
                if self.i < self.s.len() {
                    self.i += 1;
                }
                let cmp = if strip {
                    line.trim_start_matches('\t')
                } else {
                    line.as_str()
                };
                if cmp == delim {
                    break;
                }
                body.push_str(&line);
                body.push('\n');
            }
            self.bodies[idx].body = body;
        }
    }

    /// `<(…)` or `>(…)`: the cursor is on the opening parenthesis.
    fn read_process_sub(&mut self) -> Word {
        let mut w = Word {
            dynamic: true,
            ..Word::default()
        };
        match match_paren(&self.s, self.i + 1) {
            Some(close) => {
                w.subs.push(self.s[self.i + 1..close].iter().collect());
                self.i = close + 1;
            }
            None => {
                self.unsupported = true;
                self.i = self.s.len();
            }
        }
        w
    }

    fn read_word(&mut self) -> Word {
        let mut w = Word::default();
        while let Some(c) = self.peek(0) {
            if WORD_END.contains(&c) {
                break;
            }
            match c {
                '\\' => {
                    match self.peek(1) {
                        Some('\n') => {}
                        Some(n) => w.chars.push((n, false)),
                        None => w.chars.push(('\\', false)),
                    }
                    self.i += 2;
                }
                '\'' => {
                    self.i += 1;
                    loop {
                        match self.peek(0) {
                            Some('\'') => {
                                self.i += 1;
                                break;
                            }
                            Some(n) => {
                                w.chars.push((n, false));
                                self.i += 1;
                            }
                            None => {
                                self.unsupported = true;
                                break;
                            }
                        }
                    }
                }
                '"' => {
                    self.i += 1;
                    self.read_double(&mut w);
                }
                '$' => self.read_dollar(&mut w, false),
                '`' => self.read_backtick(&mut w),
                _ => {
                    w.chars.push((c, true));
                    self.i += 1;
                }
            }
        }
        w
    }

    fn read_double(&mut self, w: &mut Word) {
        loop {
            match self.peek(0) {
                None => {
                    self.unsupported = true;
                    return;
                }
                Some('"') => {
                    self.i += 1;
                    return;
                }
                Some('\\') => {
                    match self.peek(1) {
                        Some('\n') => {}
                        Some(n @ ('"' | '\\' | '$' | '`')) => w.chars.push((n, false)),
                        Some(n) => {
                            w.chars.push(('\\', false));
                            w.chars.push((n, false));
                        }
                        None => w.chars.push(('\\', false)),
                    }
                    self.i += 2;
                }
                Some('$') => self.read_dollar(w, true),
                Some('`') => self.read_backtick(w),
                Some(n) => {
                    w.chars.push((n, false));
                    self.i += 1;
                }
            }
        }
    }

    fn read_backtick(&mut self, w: &mut Word) {
        self.i += 1;
        let mut inner = String::new();
        loop {
            match self.peek(0) {
                None => {
                    self.unsupported = true;
                    return;
                }
                Some('`') => {
                    self.i += 1;
                    break;
                }
                Some('\\') => match self.peek(1) {
                    Some(n @ ('`' | '\\' | '$')) => {
                        inner.push(n);
                        self.i += 2;
                    }
                    _ => {
                        inner.push('\\');
                        self.i += 1;
                    }
                },
                Some(n) => {
                    inner.push(n);
                    self.i += 1;
                }
            }
        }
        w.dynamic = true;
        w.subs.push(inner);
    }

    fn read_dollar(&mut self, w: &mut Word, in_double: bool) {
        match self.peek(1) {
            Some('(') if self.peek(2) == Some('(') => {
                // Arithmetic: the value is computed; a nested substitution cannot be split.
                match find_double_close(&self.s, self.i + 3) {
                    Some(close) => {
                        let inner: String = self.s[self.i + 3..close].iter().collect();
                        if inner.contains("$(") || inner.contains('`') {
                            self.unsupported = true;
                        }
                        self.i = close + 2;
                    }
                    None => {
                        self.unsupported = true;
                        self.i = self.s.len();
                    }
                }
                w.dynamic = true;
            }
            Some('(') => match match_paren(&self.s, self.i + 2) {
                Some(close) => {
                    w.subs.push(self.s[self.i + 2..close].iter().collect());
                    w.dynamic = true;
                    self.i = close + 1;
                }
                None => {
                    self.unsupported = true;
                    self.i = self.s.len();
                }
            },
            Some('{') => match match_brace(&self.s, self.i + 2) {
                Some(close) => {
                    let inner: String = self.s[self.i + 2..close].iter().collect();
                    if inner.contains("$(") || inner.contains('`') {
                        self.unsupported = true;
                    }
                    w.dynamic = true;
                    self.i = close + 1;
                }
                None => {
                    self.unsupported = true;
                    self.i = self.s.len();
                }
            },
            Some('"') if !in_double => {
                // A locale string: the shell may translate the text, so the value is not the
                // text. The quotes that follow are read as an ordinary double-quoted part.
                w.dynamic = true;
                self.i += 1;
            }
            Some('\'') if !in_double => {
                // ANSI-C quoting: literal only when it holds no escape.
                let mut j = self.i + 2;
                let mut text = Vec::new();
                while j < self.s.len() && self.s[j] != '\'' {
                    if self.s[j] == '\\' {
                        j += 1;
                    }
                    text.push(self.s.get(j).copied().unwrap_or(' '));
                    j += 1;
                }
                if j >= self.s.len() {
                    self.unsupported = true;
                    self.i = self.s.len();
                    return;
                }
                let has_escape = self.s[self.i + 2..j].contains(&'\\');
                if has_escape {
                    w.dynamic = true;
                } else {
                    w.chars.extend(text.into_iter().map(|c| (c, false)));
                }
                self.i = j + 1;
            }
            Some(n) if n.is_ascii_alphabetic() || n == '_' => {
                self.i += 1;
                while self
                    .peek(0)
                    .is_some_and(|c| c.is_ascii_alphanumeric() || c == '_')
                {
                    self.i += 1;
                }
                w.dynamic = true;
            }
            Some(n) if n.is_ascii_digit() || "@*#?$!-".contains(n) => {
                self.i += 2;
                w.dynamic = true;
            }
            _ => {
                w.chars.push(('$', false));
                self.i += 1;
            }
        }
    }
}

/// How many substitution-in-quote levels the matchers below follow before the text is treated
/// as unsplittable (a deeper text is denied, and the recursion stays far from the stack limit).
const MAX_MATCH_LEVEL: usize = 64;

/// The index of the `)` closing a substitution whose body starts at `start`.
fn match_paren(s: &[char], start: usize) -> Option<usize> {
    match_paren_at(s, start, 0)
}

fn match_paren_at(s: &[char], start: usize, level: usize) -> Option<usize> {
    if level > MAX_MATCH_LEVEL {
        return None;
    }
    let mut depth = 1usize;
    let mut i = start;
    while i < s.len() {
        match s[i] {
            '\\' => i += 1,
            '\'' => {
                i += 1;
                while i < s.len() && s[i] != '\'' {
                    i += 1;
                }
                s.get(i)?;
            }
            '"' => i = skip_double(s, i + 1, level + 1)?,
            '`' => {
                i += 1;
                while i < s.len() && s[i] != '`' {
                    if s[i] == '\\' {
                        i += 1;
                    }
                    i += 1;
                }
                s.get(i)?;
            }
            '(' => depth += 1,
            ')' => {
                depth -= 1;
                if depth == 0 {
                    return Some(i);
                }
            }
            _ => {}
        }
        i += 1;
    }
    None
}

/// After a double quote opened before `start`: the index just past its closing quote.
fn skip_double(s: &[char], start: usize, level: usize) -> Option<usize> {
    if level > MAX_MATCH_LEVEL {
        return None;
    }
    let mut i = start;
    while i < s.len() {
        match s[i] {
            '\\' => i += 1,
            '"' => return Some(i),
            '$' if s.get(i + 1) == Some(&'(') => i = match_paren_at(s, i + 2, level + 1)?,
            '`' => {
                i += 1;
                while i < s.len() && s[i] != '`' {
                    if s[i] == '\\' {
                        i += 1;
                    }
                    i += 1;
                }
                s.get(i)?;
            }
            _ => {}
        }
        i += 1;
    }
    None
}

fn match_brace(s: &[char], start: usize) -> Option<usize> {
    let mut depth = 1usize;
    let mut i = start;
    while i < s.len() {
        match s[i] {
            '\\' => i += 1,
            '\'' => {
                i += 1;
                while i < s.len() && s[i] != '\'' {
                    i += 1;
                }
            }
            '"' => i = skip_double(s, i + 1, 0)?,
            '{' => depth += 1,
            '}' => {
                depth -= 1;
                if depth == 0 {
                    return Some(i);
                }
            }
            _ => {}
        }
        i += 1;
    }
    None
}

/// The index of the first `))` at or after `start`.
fn find_double_close(s: &[char], start: usize) -> Option<usize> {
    (start..s.len().saturating_sub(1)).find(|&i| s[i] == ')' && s[i + 1] == ')')
}

/// The substitution bodies inside the text of an unquoted here-document.
fn scan_subs(text: &str) -> Option<Vec<String>> {
    let s: Vec<char> = text.chars().collect();
    let mut out = Vec::new();
    let mut i = 0;
    while i < s.len() {
        match s[i] {
            '\\' => i += 1,
            '$' if s.get(i + 1) == Some(&'(') => {
                if s.get(i + 2) == Some(&'(') {
                    i = find_double_close(&s, i + 3)? + 1;
                } else {
                    let close = match_paren(&s, i + 2)?;
                    out.push(s[i + 2..close].iter().collect());
                    i = close;
                }
            }
            '`' => {
                let mut j = i + 1;
                let mut inner = String::new();
                while j < s.len() && s[j] != '`' {
                    if s[j] == '\\' {
                        j += 1;
                    }
                    inner.push(*s.get(j)?);
                    j += 1;
                }
                s.get(j)?;
                out.push(inner);
                i = j;
            }
            _ => {}
        }
        i += 1;
    }
    Some(out)
}

// ------------------------------------------------------------------ walker

#[derive(Clone)]
struct State {
    /// The working folder after the literal `cd` of the command; `None` when unreadable.
    cwd: Option<PathBuf>,
    stack: Vec<Option<PathBuf>>,
    /// Persistent git-relevant variables (`None` value: set to something unreadable).
    git_env: BTreeMap<String, Option<String>>,
}

struct Walker<'a> {
    world: &'a World,
    count: usize,
    others: Option<(PathBuf, Vec<PathBuf>)>,
}

/// What the stdin of the command is.
struct Stdin<'a> {
    piped: bool,
    redirected: bool,
    heredocs: Vec<&'a Heredoc>,
    herestr: &'a [Word],
}

/// The variables whose value the walk keeps from one command to the next: the ones git reads to
/// find its repository and configuration, and `CDPATH`, which changes where `cd` lands.
fn is_git_env_name(name: &str) -> bool {
    name == "GIT_DIR"
        || name == "GIT_WORK_TREE"
        || name == "GIT_COMMON_DIR"
        || name == "CDPATH"
        || name.starts_with("GIT_CONFIG")
}

/// A command word the text does not fix: an expansion or substitution, a glob character, a
/// brace list, or a first `=` (a command word in some shells). `[` and `[[` are the test command.
fn command_word_unreadable(w: &Word) -> bool {
    if w.dynamic {
        return true;
    }
    let text = w.text();
    if text == "[" || text == "[[" {
        return false;
    }
    w.chars
        .iter()
        .any(|&(c, bare)| bare && matches!(c, '*' | '?' | '[' | '{'))
        || w.chars.first() == Some(&('=', true))
}

/// Whether `cd` looks the operand up in `CDPATH`: a relative name that does not start with `.`
/// or `..` as its first component.
fn searched_by_cdpath(text: &str) -> bool {
    !(text.starts_with('/')
        || text == "."
        || text == ".."
        || text.starts_with("./")
        || text.starts_with("../"))
}

/// The lower-case `cdpath` array of zsh is tied to `CDPATH`; any assignment of it leaves the
/// folder of a later `cd` unreadable (the upper-case variable is followed by value).
const CDPATH_ARRAY: &str = "cdpath";

/// The state a command text that runs inside another command starts from: the working folder and
/// the git variables of the outer state, with the assignments written in front of the command
/// (`CDPATH=/opt bash -c '…'`, `env GIT_DIR=… sh -c '…'`) on top.
fn with_prefix(state: &State, prefix: &BTreeMap<String, Option<String>>) -> State {
    let mut inner = state.clone();
    for (name, value) in prefix {
        if is_git_env_name(name) {
            inner.git_env.insert(name.clone(), value.clone());
        }
    }
    inner
}

/// A word that is, or assigns, the `CDPATH` variable or the `cdpath` array: the name of a
/// variable that `read`, `printf -v`, `mapfile`, `let`, `getopts` or a `for` loop can set.
fn names_cdpath_variable(w: &Word) -> bool {
    if w.dynamic {
        return false;
    }
    let text = w.text();
    let name = text.split_once('=').map_or(text.as_str(), |(n, _)| n);
    matches!(name.trim_end_matches('+'), "CDPATH" | CDPATH_ARRAY)
}

// ------------------------------------------------------------ the allowed-option table (D-17)
//
// The option words of a shell start, `emulate`, `set`, `setopt`, `unsetopt` and `shopt` are read
// from the tables below and from nothing else: a word the tables do not list is unreadable. No
// word is judged by what it might mean, so a new spelling of an option that changes where `cd`
// looks needs no new rule. Sources of the letters: zshoptions(1) "SINGLE LETTER OPTIONS" and
// bash(1), read 2026-11-24T00:48:05Z (SPEC Sources). To read one more word, add it to a table with
// a row of AC-18.

/// Letters after `-` in a short option word: `c` and `s` are the command-line letters of a shell
/// start, `e u x f v C l` are `ERR_EXIT`, `NO_UNSET`, `XTRACE`, `NO_RCS`, `VERBOSE`, `NO_CLOBBER`
/// and `LOGIN` in zsh and the same letters in bash. `n` is `NO_EXEC`: the shell parses and does
/// not run (`bash -n FILE`); a `-c` string given with it is still judged. `o` is read apart (last
/// letter only).
const OPTION_LETTERS_MINUS: &str = "csexufvlCn";
/// Letters after `+` (the same options turned off); `c` and `s` have no `+` form.
const OPTION_LETTERS_PLUS: &str = "exufvC";
/// The names a separate `-o` or `+o` may take, and the first names of `setopt` and `unsetopt`.
const OPTION_NAMES: &[&str] = &["errexit", "nounset", "pipefail", "xtrace", "verbose"];
/// The names `setopt` and `unsetopt` read besides `OPTION_NAMES`.
const SETOPT_NAMES: &[&str] = &["extendedglob", "nullglob", "globdots"];
/// The names `shopt` reads.
const SHOPT_NAMES: &[&str] = &[
    "extglob",
    "nullglob",
    "failglob",
    "globstar",
    "dotglob",
    "nocaseglob",
    "nocasematch",
    "lastpipe",
];
/// The long words of a shell start that name no option (`--rcfile` and `--init-file` are read
/// apart because they take a file).
const SHELL_LONG_OPTIONS: &[&str] = &[
    "--login",
    "--norc",
    "--noprofile",
    "--no-rcs",
    "--version",
    "--help",
];
/// The letters of an `emulate` cluster (`-L`, `-R`, `-l`, `-c STRING`) and the shells it names.
const EMULATE_LETTERS: &str = "LRlc";
const EMULATE_SHELLS: &[&str] = &["zsh", "sh", "ksh", "csh"];

/// What one short option word (`-fe`, `+x`, `-eo`, `-c`) says where a shell reads options.
struct ShortOptions {
    /// The last letter is `o`: the next word is an option name.
    takes_name: bool,
    /// The word holds the `-c` letter (`-lc`).
    command_string: bool,
    /// The word holds the `-s` letter.
    read_stdin: bool,
}

/// The one reading of a short option word for every place that takes shell options. `None` is
/// unreadable: a letter that the table does not list, an `o` that is not the last letter (zsh reads
/// the rest of the word as the option name), a bare `-` or `+`, and a long word.
fn read_short_options(text: &str) -> Option<ShortOptions> {
    let (letters, cluster) = if let Some(rest) = text.strip_prefix('-') {
        (OPTION_LETTERS_MINUS, rest)
    } else {
        (OPTION_LETTERS_PLUS, text.strip_prefix('+')?)
    };
    if cluster.is_empty() {
        return None;
    }
    let mut scan = ShortOptions {
        takes_name: false,
        command_string: false,
        read_stdin: false,
    };
    let count = cluster.chars().count();
    for (at, c) in cluster.chars().enumerate() {
        if c == 'o' && at + 1 == count {
            scan.takes_name = true;
        } else if letters.contains(c) {
            scan.command_string |= c == 'c';
            scan.read_stdin |= c == 's';
        } else {
            return None;
        }
    }
    Some(scan)
}

/// Whether the word after `-o` or `+o` is a name of `names`: a missing word (`set -o` lists the
/// options) is fine, a word an expansion decides or a name outside the table is not.
fn option_name_listed(value: Option<&Word>, names: &[&str]) -> bool {
    value.is_none_or(|v| !v.dynamic && names.contains(&v.text().as_str()))
}

/// Whether the option words of `set`, `setopt`, `unsetopt` or `shopt` are unreadable (D-17). For
/// `set` the option places end at `--` or at the first static word that is no option (the
/// positional parameters follow and are data), and a word that an expansion decides before that
/// may be an option, so it is unreadable. For the other three every word is an option place.
fn option_command_unreadable(name: &str, args: &[Word]) -> bool {
    let mut i = 0;
    while i < args.len() {
        let w = &args[i];
        i += 1;
        if w.dynamic {
            return true;
        }
        let text = w.text();
        if name == "shopt" {
            if matches!(text.as_str(), "-s" | "-u" | "-q" | "-p") {
                continue;
            }
            if !SHOPT_NAMES.contains(&text.as_str()) {
                return true;
            }
            continue;
        }
        if name == "set" && text == "--" {
            return false;
        }
        if text.starts_with(['-', '+']) {
            let Some(scan) = read_short_options(&text) else {
                return true;
            };
            if scan.takes_name {
                if !option_name_listed(args.get(i), OPTION_NAMES) {
                    return true;
                }
                i += 1;
            }
        } else if name == "set" {
            return false;
        } else if !OPTION_NAMES.contains(&text.as_str()) && !SETOPT_NAMES.contains(&text.as_str()) {
            return true;
        }
    }
    false
}

/// The shell string of a util-linux `flock … FILE -c COMMAND` (also `--command` and a cluster
/// that ends in `c`). Options may stand before and after the lock file; the first word after
/// the file that is not an option is the program to run, and ends the scan.
fn flock_string(args: &[Word]) -> Result<Option<String>, Class> {
    let mut i = 0;
    let mut file_seen = false;
    while i < args.len() {
        let w = &args[i];
        if w.dynamic {
            return Err(Class::Unreadable);
        }
        let t = w.text();
        if t == "--" {
            return Ok(None);
        }
        if !w.is_option() {
            if file_seen {
                return Ok(None);
            }
            file_seen = true;
            i += 1;
            continue;
        }
        let value = |at: usize| -> Result<Option<String>, Class> {
            match args.get(at) {
                Some(v) if v.dynamic => Err(Class::Unreadable),
                Some(v) => Ok(Some(v.text())),
                None => Ok(None),
            }
        };
        if t == "--command" {
            return value(i + 1);
        }
        if let Some(v) = t.strip_prefix("--command=") {
            return Ok(Some(v.to_owned()));
        }
        if t.starts_with("--") {
            if matches!(t.as_str(), "--timeout" | "--wait" | "--conflict-exit-code") {
                i += 1;
            }
        } else {
            let cluster = &t[1..];
            if cluster.ends_with('c') {
                return value(i + 1);
            }
            if let Some(p) = cluster.find('c') {
                return Ok(Some(cluster[p + 1..].to_owned()));
            }
            if cluster.ends_with(['w', 'E']) {
                i += 1;
            }
        }
        i += 1;
    }
    Ok(None)
}

/// The shell string of a util-linux `script -c COMMAND` (also `--command` and a cluster that
/// ends in `c`), `None` when the call has none.
fn script_string(args: &[Word]) -> Result<Option<String>, Class> {
    let mut i = 0;
    while i < args.len() {
        let w = &args[i];
        if w.dynamic {
            return Err(Class::Unreadable);
        }
        let t = w.text();
        if t == "--" || !w.is_option() {
            return Ok(None);
        }
        let value = |at: usize| -> Result<Option<String>, Class> {
            match args.get(at) {
                Some(v) if v.dynamic => Err(Class::Unreadable),
                Some(v) => Ok(Some(v.text())),
                None => Ok(None),
            }
        };
        if t == "--command" {
            return value(i + 1);
        }
        if let Some(v) = t.strip_prefix("--command=") {
            return Ok(Some(v.to_owned()));
        }
        if !t.starts_with("--") {
            let cluster = &t[1..];
            if let Some(p) = cluster.find('c') {
                let rest = &cluster[p + 1..];
                return if rest.is_empty() {
                    value(i + 1)
                } else {
                    Ok(Some(rest.to_owned()))
                };
            }
            // A valued option (`-t time`, `-F pipe`, `-I file`, …) takes the next word.
            if cluster.ends_with(['t', 'F', 'I', 'O', 'B', 'T', 'E', 'm', 'o']) {
                i += 1;
            }
        }
        i += 1;
    }
    Ok(None)
}

/// Splits `NAME=value` off the front of the words.
fn split_assign(w: &Word) -> Option<(String, Word)> {
    let eq = w.chars.iter().position(|&(c, b)| c == '=' && b)?;
    let name: String = w.chars[..eq].iter().map(|&(c, _)| c).collect();
    let name = name.strip_suffix('+').unwrap_or(&name).to_owned();
    let valid = !name.is_empty()
        && !name.starts_with(|c: char| c.is_ascii_digit())
        && name.chars().all(|c| c.is_ascii_alphanumeric() || c == '_')
        && w.chars[..eq].iter().all(|&(_, b)| b);
    valid.then(|| {
        (
            name,
            Word {
                chars: w.chars[eq + 1..].to_vec(),
                dynamic: w.dynamic,
                subs: Vec::new(),
            },
        )
    })
}

fn value_of(w: &Word) -> Option<String> {
    (!w.dynamic).then(|| w.text())
}

fn basename(text: &str) -> &str {
    text.rsplit('/').next().unwrap_or(text)
}

const SHELLS: &[&str] = &[
    "sh", "bash", "zsh", "dash", "ksh", "mksh", "ash", "csh", "tcsh", "fish",
];
const DELETERS: &[&str] = &["rm", "rmdir", "unlink", "shred"];
/// Programs that run another command or a script: never fed unreadable stdin arguments.
const RUNNERS: &[&str] = &[
    "rm", "rmdir", "unlink", "shred", "find", "git", "xargs", "eval", "source", ".", "sh", "bash",
    "zsh", "dash", "ksh", "mksh", "ash", "csh", "tcsh", "fish", "env", "sudo", "doas", "command",
    "builtin", "exec", "nohup", "nice", "timeout", "time",
];
const RUNNERS_MORE: &[&str] = &["parallel", "watch", "busybox", "setsid", "stdbuf"];

fn is_runner(name: &str) -> bool {
    RUNNERS.contains(&name) || RUNNERS_MORE.contains(&name)
}

impl Walker<'_> {
    fn walk(&mut self, src: &str, state: &mut State, depth: usize) -> Verdict {
        if depth > MAX_DEPTH {
            return Err(Class::Unreadable);
        }
        let script = lex(src);
        if script.unsupported {
            return Err(Class::Unreadable);
        }
        let mut groups: Vec<Option<PathBuf>> = Vec::new();
        for item in &script.items {
            match item {
                Item::Open => groups.push(state.cwd.clone()),
                Item::Close => {
                    if let Some(saved) = groups.pop() {
                        state.cwd = saved;
                    }
                }
                Item::Cmd(cmd) => {
                    self.count += 1;
                    if self.count > MAX_COMMANDS {
                        return Err(Class::Unreadable);
                    }
                    self.command(cmd, &script.bodies, state, depth)?;
                }
            }
        }
        Ok(())
    }

    fn command(
        &mut self,
        cmd: &Cmd,
        bodies: &[Heredoc],
        state: &mut State,
        depth: usize,
    ) -> Verdict {
        for w in cmd.words.iter().chain(&cmd.herestr).chain(&cmd.redirs) {
            for sub in &w.subs {
                self.walk(sub, &mut state.clone(), depth + 1)?;
            }
        }
        for &idx in &cmd.heredocs {
            if !bodies[idx].quoted {
                let subs = scan_subs(&bodies[idx].body).ok_or(Class::Unreadable)?;
                for sub in subs {
                    self.walk(&sub, &mut state.clone(), depth + 1)?;
                }
            }
        }
        let mut prefix: BTreeMap<String, Option<String>> = BTreeMap::new();
        let mut at = 0;
        while at < cmd.words.len() {
            match split_assign(&cmd.words[at]) {
                Some((name, value)) => {
                    if name == CDPATH_ARRAY {
                        return Err(Class::Unreadable);
                    }
                    prefix.insert(name, value_of(&value));
                    at += 1;
                }
                None => break,
            }
        }
        let mut rest = &cmd.words[at..];
        // Reserved words in front of a command.
        while let Some(first) = rest.first() {
            let text = first.text();
            let bare = !first.dynamic && first.chars.iter().all(|&(_, b)| b);
            if bare
                && [
                    "if", "then", "else", "elif", "do", "while", "until", "!", "{", "}", "fi",
                    "done",
                ]
                .contains(&text.as_str())
            {
                rest = &rest[1..];
                continue;
            }
            if bare && ["for", "select"].contains(&text.as_str()) {
                // `for CDPATH in …` sets the variable for the body.
                return if rest.get(1).is_some_and(names_cdpath_variable) {
                    Err(Class::Unreadable)
                } else {
                    Ok(())
                };
            }
            if bare && ["case", "function", "coproc"].contains(&text.as_str()) {
                return Err(Class::Unreadable);
            }
            break;
        }
        if rest.is_empty() {
            for (name, value) in prefix {
                if is_git_env_name(&name) {
                    state.git_env.insert(name, value);
                }
            }
            return Ok(());
        }
        let stdin = Stdin {
            piped: cmd.piped_in,
            redirected: cmd.stdin_redirect,
            heredocs: cmd.heredocs.iter().map(|&i| &bodies[i]).collect(),
            herestr: &cmd.herestr,
        };
        self.simple(rest, prefix, &stdin, state, depth)
    }

    /// One simple command: peel wrapper programs, then judge the command that runs.
    fn simple(
        &mut self,
        mut words: &[Word],
        mut prefix: BTreeMap<String, Option<String>>,
        stdin: &Stdin,
        state: &mut State,
        depth: usize,
    ) -> Verdict {
        loop {
            let Some(first) = words.first() else {
                return Ok(());
            };
            if command_word_unreadable(first) {
                return Err(Class::Unreadable);
            }
            let text = first.text();
            let name = basename(&text);
            let args = &words[1..];
            match name {
                "cd" | "pushd" | "popd" => {
                    self.change_dir(name, args, state, &prefix);
                    return Ok(());
                }
                "script" => {
                    // util-linux `-c COMMAND` runs a shell string; the BSD form runs the words
                    // after the file name (peeled below).
                    if let Some(text) = script_string(args)? {
                        return self.walk(&text, &mut with_prefix(state, &prefix), depth + 1);
                    }
                }
                "export" | "declare" | "typeset" | "readonly" | "local" => {
                    for w in args {
                        if let Some((n, v)) = split_assign(w) {
                            // A name reference (`declare -n ref=CDPATH`) sets it by another name.
                            if n == CDPATH_ARRAY || names_cdpath_variable(&v) {
                                return Err(Class::Unreadable);
                            }
                            if is_git_env_name(&n) {
                                state.git_env.insert(n, value_of(&v));
                            }
                        } else if !w.dynamic && w.text() == CDPATH_ARRAY {
                            return Err(Class::Unreadable);
                        }
                    }
                    return Ok(());
                }
                // Commands that set a variable by name: the `CDPATH` they name is unreadable.
                "read" | "printf" | "mapfile" | "readarray" | "let" | "getopts" => {
                    return if args.iter().any(names_cdpath_variable) {
                        Err(Class::Unreadable)
                    } else {
                        Ok(())
                    };
                }
                // `trap ACTION SIGNAL…` runs ACTION, a shell string, later: it is judged as if it
                // stood alone.
                "trap" => {
                    let operands: Vec<&Word> = args.iter().filter(|w| !w.is_option()).collect();
                    let Some(action) = operands.first() else {
                        return Ok(());
                    };
                    if action.dynamic {
                        return Err(Class::Unreadable);
                    }
                    let text = action.text();
                    if text.is_empty() || text == "-" || text.chars().all(|c| c.is_ascii_digit()) {
                        return Ok(());
                    }
                    return self.walk(&text, &mut state.clone(), depth + 1);
                }
                // Words that change how a command word or a `cd` operand is looked up: the
                // line is unreadable whatever follows (D-12). The list of words is closed.
                "alias" | "unalias" | "hash" | "enable" => return Err(Class::Unreadable),
                "setopt" | "unsetopt" | "shopt" | "set" => {
                    return if option_command_unreadable(name, args) {
                        Err(Class::Unreadable)
                    } else {
                        Ok(())
                    };
                }
                "flock" => {
                    if let Some(text) = flock_string(args)? {
                        self.walk(&text, &mut with_prefix(state, &prefix), depth + 1)?;
                    }
                }
                "unset" => {
                    for w in args {
                        state.git_env.remove(&w.text());
                    }
                    return Ok(());
                }
                "rm" | "rmdir" | "unlink" | "shred" => return self.delete(name, args, state),
                "find" => return self.find(args, &prefix, stdin, state, depth),
                "git" => return self.git(args, &prefix, state, depth),
                "eval" => return self.eval(args, &with_prefix(state, &prefix), depth),
                "emulate" => return self.emulate(args, &with_prefix(state, &prefix), depth),
                "xargs" => return xargs(args),
                "parallel" => {
                    return if args
                        .iter()
                        .any(|w| w.dynamic || is_runner(basename(&w.text())))
                    {
                        Err(Class::Unreadable)
                    } else {
                        Ok(())
                    }
                }
                "env" => {
                    let mut i = 0;
                    while i < args.len() {
                        let w = &args[i];
                        if w.dynamic {
                            return Err(Class::Unreadable);
                        }
                        let t = w.text();
                        if let Some((n, v)) = split_assign(w) {
                            if n == CDPATH_ARRAY {
                                return Err(Class::Unreadable);
                            }
                            prefix.insert(n, value_of(&v));
                            i += 1;
                        } else if matches!(
                            t.as_str(),
                            "-i" | "-0" | "-v" | "--ignore-environment" | "--null" | "--"
                        ) {
                            i += 1;
                        } else if matches!(t.as_str(), "-u" | "-C" | "--unset" | "--chdir") {
                            i += 2;
                        } else if t.starts_with("--unset=") || t.starts_with("--chdir=") {
                            i += 1;
                        } else if t.starts_with('-') && t.len() > 1 {
                            // `-S` splits a string into a command; any other option is unknown.
                            return Err(Class::Unreadable);
                        } else {
                            break;
                        }
                    }
                    words = &args[i.min(args.len())..];
                    continue;
                }
                // `moai heavy-slot-acquire … -- COMMAND` and `moai card line-lease … -- COMMAND`
                // run the words after the first `--` as a command: they are judged like the same
                // command typed bare. The words before it are moai's own options. Without a `--`
                // no command is named (SPEC-20261125-boundary-unreadable-detail, D-2).
                "moai" | "moai-cli" => {
                    match args.iter().position(|w| !w.dynamic && w.text() == "--") {
                        Some(i) => {
                            words = &args[i + 1..];
                            continue;
                        }
                        None => return Ok(()),
                    }
                }
                "source" | "." => return Ok(()),
                _ => {}
            }
            if SHELLS.contains(&name) {
                return self.shell(args, stdin, &with_prefix(state, &prefix), depth);
            }
            if name == "command"
                && args
                    .iter()
                    .any(|w| matches!(w.text().as_str(), "-v" | "-V"))
            {
                return Ok(());
            }
            let Some(spec) = wrap_spec(name) else {
                return Ok(());
            };
            match peel(&spec, args) {
                Ok(i) => {
                    words = &args[i..];
                    continue;
                }
                Err(Peel::Dynamic) => return Err(Class::Unreadable),
                Err(Peel::Unknown) => {
                    return if args
                        .iter()
                        .any(|w| w.dynamic || is_runner(basename(&w.text())))
                    {
                        Err(Class::Unreadable)
                    } else {
                        Ok(())
                    }
                }
            }
        }
    }

    /// A literal `cd`, `pushd` or `popd`. Only the plain forms are followed: a stack index
    /// (`+N`, `-N`), `-n`, a bare `pushd` (swap) or a `popd` with an argument changes the stack
    /// in ways this text walk does not model, and a relative `cd` target under `CDPATH` may
    /// land anywhere; each of them leaves the folder unreadable.
    fn change_dir(
        &self,
        name: &str,
        args: &[Word],
        state: &mut State,
        prefix: &BTreeMap<String, Option<String>>,
    ) {
        if name != "cd" {
            let stack_form = args
                .iter()
                .any(|w| w.dynamic || w.is_option() || w.text().starts_with('+'));
            if stack_form
                || (name == "pushd" && args.is_empty())
                || (name == "popd" && !args.is_empty())
            {
                state.stack.clear();
                state.cwd = None;
                return;
            }
        }
        if name == "popd" {
            state.cwd = state.stack.pop().unwrap_or(None);
            return;
        }
        if name == "pushd" {
            state.stack.push(state.cwd.clone());
        }
        let cdpath_set = prefix
            .get("CDPATH")
            .or_else(|| state.git_env.get("CDPATH"))
            .is_some_and(|v| v.as_deref() != Some(""));
        let operands: Vec<&Word> = args.iter().filter(|w| !w.is_option()).collect();
        let target = match operands.as_slice() {
            [] => self.world.home.clone(),
            [w] if !w.dynamic && w.text() == "-" => None,
            [w] if cdpath_set && !w.dynamic && searched_by_cdpath(&w.text()) => None,
            [w] => self.path_of(w, state).and_then(|p| p.into_iter().next()),
            _ => None,
        };
        state.cwd = target.map(|p| normalize(&p));
    }

    /// The literal absolute path of a word (after tilde and brace expansion), with its glob
    /// flag; `None` when the text does not hold it.
    fn path_of(&self, w: &Word, state: &State) -> Option<Vec<PathBuf>> {
        let list = self.expand(w, state)?;
        if list.iter().any(|e| e.glob.is_some()) {
            return None;
        }
        Some(list.into_iter().map(|e| e.path).collect())
    }

    fn expand(&self, w: &Word, state: &State) -> Option<Vec<Expanded>> {
        if w.dynamic {
            return None;
        }
        let mut budget = MAX_BRACE;
        let variants = brace_expand(&w.chars, &mut budget)?;
        let mut out = Vec::new();
        for chars in variants {
            out.push(self.one_path(chars, state)?);
        }
        Some(out)
    }

    fn one_path(&self, mut chars: Vec<(char, bool)>, state: &State) -> Option<Expanded> {
        if chars.first() == Some(&('~', true)) {
            let slash = chars
                .iter()
                .position(|&(c, _)| c == '/')
                .unwrap_or(chars.len());
            if slash != 1 {
                return None;
            }
            let home = self.world.home.as_ref()?;
            let mut replaced: Vec<(char, bool)> =
                home.to_string_lossy().chars().map(|c| (c, false)).collect();
            replaced.extend_from_slice(&chars[1..]);
            chars = replaced;
        }
        if chars.is_empty() {
            return None;
        }
        let absolute = chars[0].0 == '/';
        let mut comps: Vec<Vec<(char, bool)>> = Vec::new();
        let mut cur = Vec::new();
        for &(c, b) in &chars {
            if c == '/' {
                comps.push(std::mem::take(&mut cur));
            } else {
                cur.push((c, b));
            }
        }
        let trailing_slash = cur.is_empty();
        if !cur.is_empty() {
            comps.push(cur);
        }
        comps.retain(|c| !c.is_empty());
        let is_glob =
            |c: &Vec<(char, bool)>| c.iter().any(|&(ch, b)| b && matches!(ch, '*' | '?' | '['));
        let glob_at = comps.iter().position(is_glob);
        let mut prefix = if absolute {
            PathBuf::from("/")
        } else {
            state.cwd.clone()?
        };
        let literal_len = glob_at.unwrap_or(comps.len());
        for c in &comps[..literal_len] {
            prefix.push(c.iter().map(|&(ch, _)| ch).collect::<String>());
        }
        let last_special = comps.get(literal_len.wrapping_sub(1)).is_some_and(|c| {
            matches!(
                c.iter().map(|&(ch, _)| ch).collect::<String>().as_str(),
                "." | ".."
            )
        });
        let follow = glob_at.is_some() || trailing_slash || last_special;
        let path = resolve(&prefix, follow)?;
        let glob = match glob_at {
            None => None,
            Some(k) => {
                let g = &comps[k];
                let dot_first = g[0].0 == '.';
                let literal_prefix = !(g[0].1 && matches!(g[0].0, '*' | '?' | '['));
                let climbs = comps[k + 1..]
                    .iter()
                    .any(|c| c.iter().map(|&(ch, _)| ch).collect::<String>() == "..");
                Some(Glob {
                    literal_prefix,
                    hidden_or_climbing: (dot_first
                        && g.iter().any(|&(ch, b)| b && matches!(ch, '*' | '?' | '[')))
                        || climbs,
                })
            }
        };
        Some(Expanded { path, glob })
    }

    // ---------------------------------------------------------- deletion

    fn delete(&mut self, name: &str, args: &[Word], state: &State) -> Verdict {
        let mut after_dd = false;
        let mut i = 0;
        while i < args.len() {
            let w = &args[i];
            i += 1;
            if w.dynamic {
                return Err(Class::Unreadable);
            }
            if !after_dd && w.text() == "--" {
                after_dd = true;
                continue;
            }
            if !after_dd && w.is_option() {
                if name == "shred"
                    && matches!(w.text().as_str(), "-n" | "-s" | "--iterations" | "--size")
                {
                    i += 1;
                }
                continue;
            }
            self.check_target(w, state, false)?;
        }
        Ok(())
    }

    /// One delete target. `keep_root_ok` allows the checkout root itself (a `find` start with a
    /// name filter).
    fn check_target(&mut self, w: &Word, state: &State, keep_root_ok: bool) -> Verdict {
        let list = self.expand(w, state).ok_or(Class::Unreadable)?;
        for e in list {
            self.classify(&e, keep_root_ok)?;
        }
        Ok(())
    }

    fn classify(&mut self, e: &Expanded, root_ok: bool) -> Verdict {
        let t = &e.path;
        if t.components().any(|c| c.as_os_str() == ".git") {
            return Err(Class::DeleteOutside);
        }
        if let Some(g) = &e.glob {
            if g.hidden_or_climbing {
                return Err(Class::DeleteOutside);
            }
        }
        let children_only = e.glob.is_some();
        let world = self.world;
        if let Some(own) = &world.own {
            if t.starts_with(&own.root) {
                if *t == own.root && !(children_only || root_ok) {
                    return Err(Class::DeleteOutside);
                }
                return self.no_other_checkout(t, children_only);
            }
        }
        let mut roots: Vec<PathBuf> = world.roots.clone();
        if let Some(r) = folders_root(t) {
            roots.push(r);
        }
        // A root itself is never a target, whichever shorter root it lies under (the scratch
        // folder under `/tmp` is a root of its own).
        let root_ok_here = root_ok || e.glob.as_ref().is_some_and(|g| g.literal_prefix);
        if !root_ok_here && roots.iter().any(|r| t == r) {
            return Err(Class::DeleteOutside);
        }
        for r in &roots {
            if t.starts_with(r) {
                return self.no_other_checkout(t, children_only);
            }
        }
        Err(Class::DeleteOutside)
    }

    /// The target is, holds or lies inside a checkout other than the session's own.
    fn no_other_checkout(&mut self, t: &Path, children_only: bool) -> Verdict {
        let world = self.world;
        let Some(own) = &world.own else { return Ok(()) };
        if self.others.as_ref().map(|(c, _)| c) != Some(&own.common) {
            self.others = Some((own.common.clone(), all_checkouts(&own.common)));
        }
        let Some((_, others)) = &self.others else {
            return Ok(());
        };
        for o in others {
            if *o == own.root {
                continue;
            }
            // A checkout that holds the session's own (the primary above a linked worktree) is
            // not "another" one for a target inside the own checkout.
            let holds_own = own.root.starts_with(o);
            if !holds_own && (o == t || t.starts_with(o)) {
                return Err(Class::DeleteOutside);
            }
            if o.starts_with(t) && o != t {
                if children_only {
                    let below = o.strip_prefix(t).ok().and_then(|r| r.components().next());
                    if below.is_some_and(|c| c.as_os_str().to_string_lossy().starts_with('.')) {
                        continue;
                    }
                }
                return Err(Class::DeleteOutside);
            }
        }
        Ok(())
    }

    fn find(
        &mut self,
        args: &[Word],
        prefix: &BTreeMap<String, Option<String>>,
        stdin: &Stdin,
        state: &mut State,
        depth: usize,
    ) -> Verdict {
        let mut i = 0;
        while i < args.len() {
            let t = args[i].text();
            if args[i].dynamic {
                return Err(Class::Unreadable);
            }
            if matches!(t.as_str(), "-H" | "-L" | "-P") || t.starts_with("-O") {
                i += 1;
            } else if t == "-D" {
                i += 2;
            } else {
                break;
            }
        }
        let mut starts: Vec<&Word> = Vec::new();
        while i < args.len() {
            let w = &args[i];
            let t = w.text();
            if !w.dynamic && (t.starts_with('-') || t == "(" || t == "!" || t == ",") {
                break;
            }
            starts.push(w);
            i += 1;
        }
        let default_dot = Word::literal(".");
        if starts.is_empty() {
            starts.push(&default_dot);
        }
        let mut filtered = false;
        let mut start_checked = false;
        let check_starts = |this: &mut Self, filtered: bool| -> Verdict {
            for s in &starts {
                this.check_target(s, state, filtered)?;
            }
            Ok(())
        };
        let mut j = i;
        while j < args.len() {
            let w = &args[j];
            if w.dynamic {
                return Err(Class::Unreadable);
            }
            let t = w.text();
            match t.as_str() {
                "-name" | "-iname" | "-path" | "-ipath" | "-wholename" | "-iwholename"
                | "-regex" | "-iregex" => {
                    filtered = true;
                    j += 2;
                }
                "-delete" => {
                    check_starts(self, filtered)?;
                    start_checked = true;
                    j += 1;
                }
                "-exec" | "-execdir" | "-ok" | "-okdir" => {
                    let mut k = j + 1;
                    while k < args.len()
                        && (args[k].dynamic || !matches!(args[k].text().as_str(), ";" | "+"))
                    {
                        k += 1;
                    }
                    let inner: Vec<Word> = args[j + 1..k.min(args.len())].to_vec();
                    // `{}` stands for a file below the start folders: the starts decide.
                    let deleting = inner
                        .first()
                        .is_some_and(|f| !f.dynamic && DELETERS.contains(&basename(&f.text())));
                    if deleting {
                        check_starts(self, filtered)?;
                        start_checked = true;
                    } else {
                        let placeholder = starts.first().map(|s| s.text()).unwrap_or_default();
                        let swapped: Vec<Word> = inner
                            .iter()
                            .map(|x| {
                                if !x.dynamic && x.text() == "{}" {
                                    Word::literal(&placeholder)
                                } else {
                                    x.clone()
                                }
                            })
                            .collect();
                        self.simple(
                            &swapped,
                            prefix.clone(),
                            stdin,
                            &mut state.clone(),
                            depth + 1,
                        )?;
                    }
                    j = k + 1;
                }
                _ => j += 1,
            }
        }
        let _ = start_checked;
        Ok(())
    }

    // ---------------------------------------------------------- shells

    fn shell(&mut self, args: &[Word], stdin: &Stdin, state: &State, depth: usize) -> Verdict {
        let mut command_string = false;
        let mut read_stdin = false;
        let mut i = 0;
        let mut operands: Vec<&Word> = Vec::new();
        while i < args.len() {
            let w = &args[i];
            i += 1;
            if w.dynamic {
                // Before the first operand a word is an option or the script, and an expansion
                // decides which (`"$O" -c …`, `-o$O`): every form is unreadable (D-17).
                if operands.is_empty() {
                    return Err(Class::Unreadable);
                }
                operands.push(w);
                continue;
            }
            let t = w.text();
            if !operands.is_empty() {
                operands.push(w);
            } else if t == "--" {
                operands.extend(&args[i..]);
                break;
            } else if matches!(t.as_str(), "--rcfile" | "--init-file") {
                // The value of these two is a file.
                i += 1;
            } else if t.starts_with("--") {
                // A long option is read only from the table; `--cdablevars` and every other
                // spelling of an option name is unreadable (D-17).
                if !SHELL_LONG_OPTIONS.contains(&t.as_str()) {
                    return Err(Class::Unreadable);
                }
            } else if t.starts_with(['-', '+']) {
                // A short word is read only from the allowed letters; `o` stands last and takes
                // the next word, which must be a name of the table (D-17).
                let Some(scan) = read_short_options(&t) else {
                    return Err(Class::Unreadable);
                };
                if scan.takes_name {
                    if !option_name_listed(args.get(i), OPTION_NAMES) {
                        return Err(Class::Unreadable);
                    }
                    i += 1;
                }
                command_string |= scan.command_string;
                read_stdin |= scan.read_stdin;
            } else {
                operands.push(w);
            }
        }
        if command_string {
            let Some(script) = operands.first() else {
                return Ok(());
            };
            if script.dynamic {
                return Err(Class::Unreadable);
            }
            return self.walk(&script.text(), &mut state.clone(), depth + 1);
        }
        if !operands.is_empty() && !read_stdin {
            // A script file: its contents are not part of the command text.
            return Ok(());
        }
        if !stdin.heredocs.is_empty() || !stdin.herestr.is_empty() {
            for h in &stdin.heredocs {
                self.walk(&h.body, &mut state.clone(), depth + 1)?;
            }
            for w in stdin.herestr {
                if w.dynamic {
                    return Err(Class::Unreadable);
                }
                self.walk(&w.text(), &mut state.clone(), depth + 1)?;
            }
            return Ok(());
        }
        if stdin.piped || stdin.redirected {
            return Err(Class::Unreadable);
        }
        Ok(())
    }

    /// zsh `emulate [-lLR] [shell] [-c STRING]`: it sets shell options like a shell start does, so
    /// every word is read by the allowed-option table (D-17): a cluster of the letters `L R l c`
    /// and a shell name of the table are read, and anything else (an option such as `-o` or `-T`,
    /// a word an expansion decides) is unreadable. The `-c` string is judged as if it stood alone.
    fn emulate(&mut self, args: &[Word], state: &State, depth: usize) -> Verdict {
        let mut i = 0;
        while i < args.len() {
            let w = &args[i];
            i += 1;
            if w.dynamic {
                return Err(Class::Unreadable);
            }
            let t = w.text();
            if let Some(cluster) = t.strip_prefix('-') {
                if cluster.is_empty() || !cluster.chars().all(|c| EMULATE_LETTERS.contains(c)) {
                    return Err(Class::Unreadable);
                }
                if cluster.contains('c') {
                    return match args.get(i) {
                        Some(s) if s.dynamic => Err(Class::Unreadable),
                        Some(s) => self.walk(&s.text(), &mut state.clone(), depth + 1),
                        None => Ok(()),
                    };
                }
            } else if !EMULATE_SHELLS.contains(&t.as_str()) {
                return Err(Class::Unreadable);
            }
        }
        Ok(())
    }

    fn eval(&mut self, args: &[Word], state: &State, depth: usize) -> Verdict {
        if args.iter().any(|w| w.dynamic) {
            return Err(Class::Unreadable);
        }
        let joined = args.iter().map(Word::text).collect::<Vec<_>>().join(" ");
        self.walk(&joined, &mut state.clone(), depth + 1)
    }

    // ---------------------------------------------------------- git

    fn git(
        &mut self,
        args: &[Word],
        prefix: &BTreeMap<String, Option<String>>,
        state: &State,
        depth: usize,
    ) -> Verdict {
        let mut env_view = state.git_env.clone();
        env_view.extend(prefix.clone());
        let mut dirs: Vec<Word> = Vec::new();
        let mut git_dir: Option<Word> = None;
        let mut work_tree: Option<Word> = None;
        let mut i = 0;
        let mut sub: Option<&Word> = None;
        while i < args.len() {
            let w = &args[i];
            i += 1;
            if !w.is_option() {
                sub = Some(w);
                break;
            }
            let t = w.text();
            match t.as_str() {
                "-C" => {
                    dirs.push(args.get(i).ok_or(Class::Unreadable)?.clone());
                    i += 1;
                }
                "-c" => {
                    let kv = args.get(i).ok_or(Class::Unreadable)?;
                    i += 1;
                    config_key_check(kv)?;
                }
                "--git-dir" => {
                    git_dir = Some(args.get(i).ok_or(Class::Unreadable)?.clone());
                    i += 1;
                }
                "--work-tree" => {
                    work_tree = Some(args.get(i).ok_or(Class::Unreadable)?.clone());
                    i += 1;
                }
                "--namespace" | "--super-prefix" => i += 1,
                "--config-env" => {
                    let kv = args.get(i).ok_or(Class::Unreadable)?;
                    i += 1;
                    config_key_check(kv)?;
                }
                _ if t.starts_with("--config-env=") => {
                    config_key_check(&Word::literal(&t["--config-env=".len()..]))?;
                }
                _ if t.starts_with("--git-dir=") => {
                    git_dir = Some(Word::literal(&t["--git-dir=".len()..]));
                }
                _ if t.starts_with("--work-tree=") => {
                    work_tree = Some(Word::literal(&t["--work-tree=".len()..]));
                }
                _ if t.starts_with("--namespace=")
                    || t.starts_with("--exec-path")
                    || t.starts_with("--super-prefix=") => {}
                "-p"
                | "-P"
                | "--paginate"
                | "--no-pager"
                | "--bare"
                | "--no-replace-objects"
                | "--literal-pathspecs"
                | "--glob-pathspecs"
                | "--noglob-pathspecs"
                | "--icase-pathspecs"
                | "--no-optional-locks"
                | "--no-lazy-fetch"
                | "--version"
                | "--help"
                | "-h"
                | "-v"
                | "--html-path"
                | "--man-path"
                | "--info-path" => {}
                _ => return Err(Class::Unreadable),
            }
        }
        let Some(sub_word) = sub else { return Ok(()) };
        if sub_word.dynamic {
            return Err(Class::Unreadable);
        }
        let name = sub_word.text();
        let rest = &args[i..];
        // A subcommand that runs another command: judge that command as if it stood alone.
        self.git_runs_command(&name, rest, prefix, &with_prefix(state, prefix), depth)?;
        let kind = classify_git(&name, rest)?;
        if kind == GitKind::Read {
            return Ok(());
        }
        // Environment that replaces the configuration or the repository.
        for (k, v) in &env_view {
            if !k.starts_with("GIT_CONFIG") {
                continue;
            }
            match k.as_str() {
                "GIT_CONFIG_GLOBAL" | "GIT_CONFIG_SYSTEM" | "GIT_CONFIG" => {
                    if v.as_deref() != Some("/dev/null") {
                        return Err(Class::Unreadable);
                    }
                }
                "GIT_CONFIG_NOSYSTEM" | "GIT_CONFIG_COUNT" => {}
                _ => match v {
                    None => return Err(Class::Unreadable),
                    Some(text) if text.to_lowercase().contains("hookspath") => {
                        return Err(Class::HookSkip)
                    }
                    Some(_) => {}
                },
            }
        }
        let parsed = Parsed::new(&name, rest);
        if name == "config" {
            return config_check(&parsed);
        }
        // The checkout this command acts on.
        let target = self.target_checkout(
            state,
            &dirs,
            git_dir.as_ref(),
            work_tree.as_ref(),
            &env_view,
        )?;
        if let Some(t) = &target {
            let own_root = self.world.own.as_ref().map(|o| &o.root);
            if name == "worktree" {
                let sub_sub = parsed.pos.first().map(|w| w.text()).unwrap_or_default();
                if matches!(sub_sub.as_str(), "remove" | "move" | "lock" | "unlock") {
                    let path = parsed.pos.get(1).ok_or(Class::Unreadable)?;
                    let list = self.expand(path, state).ok_or(Class::Unreadable)?;
                    for e in list {
                        let same = checkout_of(&e.path).map(|c| c.root) == own_root.cloned();
                        if !same || e.path != *own_root.unwrap_or(&PathBuf::new()) {
                            return Err(Class::GitState);
                        }
                    }
                }
            } else if Some(&t.root) != own_root || (t.primary && D5.contains(&name.as_str())) {
                return Err(Class::GitState);
            }
        }
        self.ref_checks(&name, &parsed, target.as_ref())?;
        hook_checks(&name, &parsed, rest)?;
        Ok(())
    }

    /// The git subcommands that hand a command to the system: `rebase -x/--exec`, `difftool
    /// -x/--extcmd`, `submodule foreach`, `bisect run`, the `--*-filter` options of
    /// `filter-branch` and the `--upload-pack`/`--receive-pack`/`--exec` programs of the
    /// transport commands. The command they run is judged as if it stood alone (REQ-9), and a
    /// command text that a substitution decides is unreadable.
    fn git_runs_command(
        &mut self,
        name: &str,
        rest: &[Word],
        prefix: &BTreeMap<String, Option<String>>,
        state: &State,
        depth: usize,
    ) -> Verdict {
        let mut strings: Vec<String> = Vec::new();
        match name {
            "rebase" => option_values(rest, &['x'], &["--exec"], &mut strings)?,
            "difftool" => option_values(rest, &['x'], &["--extcmd"], &mut strings)?,
            "filter-branch" => option_values(
                rest,
                &[],
                &[
                    "--env-filter",
                    "--tree-filter",
                    "--index-filter",
                    "--parent-filter",
                    "--msg-filter",
                    "--commit-filter",
                    "--tag-name-filter",
                    "--setup",
                ],
                &mut strings,
            )?,
            "clone" => option_values(rest, &['u'], &["--upload-pack"], &mut strings)?,
            "grep" => option_values(rest, &['O'], &["--open-files-in-pager"], &mut strings)?,
            "fetch" | "pull" | "push" | "ls-remote" | "archive" => option_values(
                rest,
                &[],
                &["--upload-pack", "--receive-pack", "--exec"],
                &mut strings,
            )?,
            "submodule" | "bisect" => {
                let words: Vec<&Word> = rest.iter().filter(|w| !w.is_option()).collect();
                let first = words.first().filter(|w| !w.dynamic).map(|w| w.text());
                if name == "submodule" && first.as_deref() == Some("foreach") {
                    let at = rest
                        .iter()
                        .position(|w| !w.dynamic && w.text() == "foreach")
                        .unwrap_or(0);
                    let tail: Vec<&Word> = rest[at + 1..]
                        .iter()
                        .skip_while(|w| w.is_option())
                        .collect();
                    if tail.iter().any(|w| w.dynamic) {
                        return Err(Class::Unreadable);
                    }
                    if !tail.is_empty() {
                        strings.push(tail.iter().map(|w| w.text()).collect::<Vec<_>>().join(" "));
                    }
                }
                if name == "bisect" && first.as_deref() == Some("run") {
                    let at = rest
                        .iter()
                        .position(|w| !w.dynamic && w.text() == "run")
                        .unwrap_or(0);
                    let command = &rest[at + 1..];
                    if !command.is_empty() {
                        let stdin = Stdin {
                            piped: false,
                            redirected: false,
                            heredocs: Vec::new(),
                            herestr: &[],
                        };
                        self.simple(
                            command,
                            prefix.clone(),
                            &stdin,
                            &mut state.clone(),
                            depth + 1,
                        )?;
                    }
                }
            }
            _ => {}
        }
        for text in strings {
            self.walk(&text, &mut state.clone(), depth + 1)?;
        }
        Ok(())
    }

    fn target_checkout(
        &self,
        state: &State,
        dirs: &[Word],
        git_dir: Option<&Word>,
        work_tree: Option<&Word>,
        env_view: &BTreeMap<String, Option<String>>,
    ) -> Result<Option<Checkout>, Class> {
        let mut base = state.cwd.clone();
        for d in dirs {
            let list = self
                .path_of(
                    d,
                    &State {
                        cwd: base.clone(),
                        ..state.clone()
                    },
                )
                .ok_or(Class::Unreadable)?;
            base = list.into_iter().next();
            if base.is_none() {
                return Err(Class::Unreadable);
            }
        }
        let env_word = |name: &str| -> Result<Option<Word>, Class> {
            match env_view.get(name) {
                None => Ok(None),
                Some(None) => Err(Class::Unreadable),
                Some(Some(v)) => Ok(Some(Word::literal(v))),
            }
        };
        let work = match work_tree {
            Some(w) => Some(w.clone()),
            None => env_word("GIT_WORK_TREE")?,
        };
        let gdir = match git_dir {
            Some(w) => Some(w.clone()),
            None => env_word("GIT_DIR")?,
        };
        let here = State {
            cwd: base.clone(),
            ..state.clone()
        };
        let dir = if let Some(w) = work {
            self.path_of(&w, &here)
                .ok_or(Class::Unreadable)?
                .into_iter()
                .next()
        } else if let Some(g) = gdir {
            let p = self
                .path_of(&g, &here)
                .ok_or(Class::Unreadable)?
                .into_iter()
                .next()
                .ok_or(Class::Unreadable)?;
            if p.file_name().is_some_and(|n| n == ".git") {
                p.parent().map(Path::to_path_buf)
            } else if let Ok(text) = fs::read_to_string(p.join("gitdir")) {
                Path::new(text.trim()).parent().map(Path::to_path_buf)
            } else {
                return Err(Class::Unreadable);
            }
        } else {
            base
        };
        let dir = dir.ok_or(Class::Unreadable)?;
        Ok(checkout_of(&dir))
    }

    fn ref_checks(&self, name: &str, p: &Parsed, target: Option<&Checkout>) -> Verdict {
        let w = self.world;
        let head = target.and_then(head_branch);
        let head_protected = head.as_deref().is_some_and(|h| w.is_protected(h));
        let protected_name = |text: &str| -> bool {
            let short = text.strip_prefix("refs/heads/").unwrap_or(text);
            w.is_protected(short)
        };
        match name {
            "update-ref" => {
                if p.has("--stdin") {
                    return Err(Class::Unreadable);
                }
                if p.dynamic {
                    return Err(Class::Unreadable);
                }
                if let Some(r) = p.pos.first() {
                    let text = r.text();
                    // `HEAD` is the ref of the branch the checkout is on.
                    if protected_name(&text) || (text == "HEAD" && head_protected) {
                        return Err(Class::ProtectedRef);
                    }
                }
            }
            "symbolic-ref" => {
                if p.dynamic {
                    return Err(Class::Unreadable);
                }
                if let Some(r) = p.pos.first() {
                    if protected_name(&r.text()) {
                        return Err(Class::ProtectedRef);
                    }
                }
            }
            "branch" => {
                if p.dynamic {
                    return Err(Class::Unreadable);
                }
                let delete = p.has_any(&["-d", "-D", "--delete"]);
                let moving = p.has_any(&["-m", "-M", "--move", "-c", "-C", "--copy"]);
                let names: Vec<&Word> = if delete {
                    p.pos.clone()
                } else if moving {
                    p.pos.iter().take(2).copied().collect()
                } else {
                    p.pos.iter().take(1).copied().collect()
                };
                if names.iter().any(|n| protected_name(&n.text())) {
                    return Err(Class::ProtectedRef);
                }
                // `git branch -m <new>` renames the branch the checkout is on.
                if p.has_any(&["-m", "-M", "--move"]) && p.pos.len() < 2 && head_protected {
                    return Err(Class::ProtectedRef);
                }
            }
            "checkout" | "switch" | "worktree" => {
                if p.dynamic {
                    return Err(Class::Unreadable);
                }
                // `--track origin/main` creates the local branch `main`.
                let track = p
                    .opts
                    .iter()
                    .any(|(o, _)| o == "-t" || long_matches(o, "--track"));
                let named = NAME_SLOTS.iter().any(|f| !p.values_of(f).is_empty());
                if track && !named && name != "worktree" {
                    if let Some(start) = p.pos.first() {
                        let text = start.text();
                        let local = text.split_once('/').map_or(text.as_str(), |(_, b)| b);
                        if protected_name(local) {
                            return Err(Class::ProtectedRef);
                        }
                    }
                }
                // `worktree add <path> [<commit-ish>]` without `-b` checks the commit-ish out as a
                // branch, or creates a branch named after the path, when it is not a commit.
                if name == "worktree"
                    && !named
                    && !p.has_any(&["--detach", "-d"])
                    && p.pos.first().is_some_and(|w| w.text() == "add")
                {
                    let created = p
                        .pos
                        .get(2)
                        .map(|w| w.text())
                        .or_else(|| p.pos.get(1).map(|w| basename(&w.text()).to_owned()));
                    if created.is_some_and(|c| protected_name(&c)) {
                        return Err(Class::ProtectedRef);
                    }
                }
                // A plain `checkout main` or `switch main` creates the local branch from the
                // remote-tracking one when it does not exist; a `--` makes the word a tree to
                // restore from, and `--detach` makes it a commit.
                if matches!(name, "checkout" | "switch")
                    && !named
                    && !track
                    && p.tail.is_empty()
                    && !p.has_any(&["--detach", "-d", "--orphan"])
                    && p.pos.len() == 1
                    && protected_name(&p.pos[0].text())
                {
                    return Err(Class::ProtectedRef);
                }
                for flag in [
                    "-b",
                    "-B",
                    "-c",
                    "-C",
                    "--orphan",
                    "--create",
                    "--force-create",
                ] {
                    for v in p.values_of(flag) {
                        if protected_name(&v) {
                            return Err(Class::ProtectedRef);
                        }
                    }
                }
            }
            "push" => {
                if p.dynamic {
                    return Err(Class::Unreadable);
                }
                if p.has_any(&["--mirror", "--all", "--branches"]) {
                    return Err(Class::ProtectedRef);
                }
                let specs: Vec<String> = p.pos.iter().skip(1).map(|x| x.text()).collect();
                if specs.is_empty() {
                    if !p.has_any(&["--tags"]) && head_protected {
                        return Err(Class::ProtectedRef);
                    }
                } else {
                    for spec in &specs {
                        let spec = spec.strip_prefix('+').unwrap_or(spec);
                        if spec.contains('*') {
                            return Err(Class::ProtectedRef);
                        }
                        let dst = match spec.split_once(':') {
                            Some((_, d)) => d.to_owned(),
                            None if spec == "HEAD" || spec == "@" => {
                                head.clone().unwrap_or_default()
                            }
                            None => spec.to_owned(),
                        };
                        if dst.starts_with("refs/") && !dst.starts_with("refs/heads/") {
                            continue;
                        }
                        if protected_name(&dst) {
                            return Err(Class::ProtectedRef);
                        }
                    }
                }
            }
            "fetch" | "pull" => {
                if p.dynamic {
                    return Err(Class::Unreadable);
                }
                for spec in p.pos.iter().skip(1).map(|x| x.text()) {
                    let spec = spec.strip_prefix('+').unwrap_or(&spec).to_owned();
                    if let Some((_, dst)) = spec.split_once(':') {
                        if dst.contains('*') && !dst.starts_with("refs/remotes/") {
                            return Err(Class::ProtectedRef);
                        }
                        if (!dst.starts_with("refs/") || dst.starts_with("refs/heads/"))
                            && protected_name(dst)
                        {
                            return Err(Class::ProtectedRef);
                        }
                    }
                }
            }
            "rebase" => {
                if p.dynamic {
                    return Err(Class::Unreadable);
                }
                // `git rebase <upstream> <branch>` checks `<branch>` out and moves it; with
                // `--root` there is no upstream and the one positional is the branch.
                let moved = if p.has("--root") {
                    p.pos.last()
                } else {
                    p.pos.get(1)
                };
                if let Some(branch) = moved {
                    if protected_name(&branch.text()) {
                        return Err(Class::ProtectedRef);
                    }
                }
            }
            "filter-branch" => {
                if p.dynamic {
                    return Err(Class::Unreadable);
                }
                // The refs it rewrites are the positionals before and after `--`; none names
                // the branch `HEAD` is on, and `--all` names every branch.
                let refs: Vec<&Word> = p.pos.iter().chain(p.tail.iter()).copied().collect();
                if refs.iter().any(|r| r.dynamic) {
                    return Err(Class::Unreadable);
                }
                let texts: Vec<String> = refs.iter().map(|r| r.text()).collect();
                let all = p.has_any(&["--all", "--branches"])
                    || texts
                        .iter()
                        .any(|t| matches!(t.as_str(), "--all" | "--branches"));
                let heads = texts.iter().filter(|t| !t.starts_with('-')).count() == 0
                    || texts.iter().any(|t| t == "HEAD");
                if all || texts.iter().any(|t| protected_name(t)) || (heads && head_protected) {
                    return Err(Class::ProtectedRef);
                }
            }
            _ => {}
        }
        // A command that advances the branch `HEAD` names.
        let advances = match name {
            "commit" | "merge" | "pull" | "rebase" | "cherry-pick" | "revert" | "am" => true,
            "reset" => {
                p.has_any(&["--soft", "--mixed", "--hard", "--merge", "--keep"])
                    || !p.pos.is_empty()
            }
            _ => false,
        };
        if advances && head_protected {
            return Err(Class::ProtectedRef);
        }
        Ok(())
    }
}

struct Glob {
    /// The glob component starts with a literal character.
    literal_prefix: bool,
    /// The glob can match `.`/`..` or a later component climbs out.
    hidden_or_climbing: bool,
}

struct Expanded {
    path: PathBuf,
    glob: Option<Glob>,
}

fn normalize(p: &Path) -> PathBuf {
    let mut out = PathBuf::new();
    for c in p.components() {
        match c {
            Component::ParentDir => {
                out.pop();
            }
            Component::CurDir => {}
            other => out.push(other.as_os_str()),
        }
    }
    if out.as_os_str().is_empty() {
        PathBuf::from("/")
    } else {
        out
    }
}

/// Brace lists (`{a,b}`) of bare braces become separate words; a range or too many results is
/// unreadable.
fn brace_expand(chars: &[(char, bool)], budget: &mut usize) -> Option<Vec<Vec<(char, bool)>>> {
    let Some(open) = chars.iter().position(|&(c, b)| c == '{' && b) else {
        return Some(vec![chars.to_vec()]);
    };
    let mut depth = 0;
    let mut commas = Vec::new();
    let mut close = None;
    for (k, &(c, b)) in chars.iter().enumerate().skip(open) {
        if !b {
            continue;
        }
        match c {
            '{' => depth += 1,
            '}' => {
                depth -= 1;
                if depth == 0 {
                    close = Some(k);
                    break;
                }
            }
            ',' if depth == 1 => commas.push(k),
            _ => {}
        }
    }
    let Some(close) = close else {
        return Some(vec![chars.to_vec()]);
    };
    if commas.is_empty() {
        let inner = &chars[open + 1..close];
        if inner
            .windows(2)
            .any(|w| w[0] == ('.', true) && w[1] == ('.', true))
        {
            return None;
        }
        let tails = brace_expand(&chars[close + 1..], budget)?;
        return Some(
            tails
                .into_iter()
                .map(|t| {
                    let mut v = chars[..=close].to_vec();
                    v.extend(t);
                    v
                })
                .collect(),
        );
    }
    let mut bounds = vec![open];
    bounds.extend(&commas);
    bounds.push(close);
    let mut out = Vec::new();
    for pair in bounds.windows(2) {
        let mut v = chars[..open].to_vec();
        v.extend_from_slice(&chars[pair[0] + 1..pair[1]]);
        v.extend_from_slice(&chars[close + 1..]);
        for r in brace_expand(&v, budget)? {
            if *budget == 0 {
                return None;
            }
            *budget -= 1;
            out.push(r);
        }
    }
    Some(out)
}

// ------------------------------------------------------------------ wrappers

struct WrapSpec {
    flags: &'static [&'static str],
    valued: &'static [&'static str],
    positional: usize,
}

enum Peel {
    Dynamic,
    Unknown,
}

fn wrap_spec(name: &str) -> Option<WrapSpec> {
    let (flags, valued, positional): (&[&str], &[&str], usize) = match name {
        "sudo" => (
            &[
                "-A",
                "-b",
                "-E",
                "-e",
                "-H",
                "-i",
                "-K",
                "-k",
                "-n",
                "-P",
                "-S",
                "-s",
                "--preserve-env",
                "--login",
                "--shell",
                "--non-interactive",
            ],
            &[
                "-u", "-g", "-h", "-p", "-C", "-D", "-R", "-T", "-U", "-r", "-t", "--user",
                "--group", "--host", "--prompt", "--chdir",
            ],
            0,
        ),
        "doas" => (&["-n", "-s"], &["-u", "-C"], 0),
        "nohup" | "builtin" | "busybox" => (&[], &[], 0),
        "setsid" => (&["-c", "-f", "-w"], &[], 0),
        "time" => (
            &["-p", "-l", "-v"],
            &["-f", "-o", "--format", "--output"],
            0,
        ),
        "nice" => (&[], &["-n", "--adjustment"], 0),
        "ionice" => (&["-t"], &["-c", "-n", "-p", "-P", "-u"], 0),
        "timeout" => (
            &["--preserve-status", "--foreground", "-v", "--verbose"],
            &["-s", "-k", "--signal", "--kill-after"],
            1,
        ),
        "stdbuf" => (
            &[],
            &["-i", "-o", "-e", "--input", "--output", "--error"],
            0,
        ),
        "exec" => (&["-c", "-l"], &["-a"], 0),
        "command" => (&["-p"], &[], 0),
        "caffeinate" => (&["-d", "-i", "-m", "-s", "-u"], &["-t", "-w"], 0),
        "watch" => (
            &[
                "-b",
                "-c",
                "-e",
                "-g",
                "-p",
                "-q",
                "-x",
                "-t",
                "-d",
                "--differences",
            ],
            &["-n", "-w", "--interval"],
            0,
        ),
        // macOS `arch [-arch name | -arm64 | -x86_64 | -32 | -64 | -c | -e VAR=x | -d VAR] cmd`.
        "arch" => (
            &["-arm64", "-x86_64", "-x86_64h", "-i386", "-c"],
            &["-arch", "-e", "-d"],
            0,
        ),
        // macOS `taskpolicy [-d p] [-g p] [-c clamp] [-b] [-t tier] [-l tier] [-a] [-s] [-S shims]
        // [-m limit] [-j pri] [-P action] program…`; `-p pid` names no program.
        "taskpolicy" => (
            &["-b", "-B", "-a", "-s", "-x", "-X"],
            &["-d", "-g", "-c", "-t", "-l", "-m", "-j", "-P", "-S", "-p"],
            0,
        ),
        // `xcrun [--sdk name] [--toolchain name] [-v -n -k -l -r -f] tool…` (the `--show-*` forms
        // name no tool).
        "xcrun" => (
            &[
                "-v",
                "--verbose",
                "-n",
                "--no-cache",
                "-k",
                "--kill-cache",
                "-l",
                "--log",
                "-r",
                "--run",
                "-f",
                "--find",
                "--show-sdk-path",
                "--show-sdk-version",
                "--show-sdk-build-version",
                "--show-sdk-platform-path",
                "--show-sdk-platform-version",
                "--show-toolchain-path",
            ],
            &["--sdk", "-sdk", "--toolchain", "-toolchain"],
            0,
        ),
        // `sandbox-exec [-f file] [-n name] [-p string] [-D key=value] command…`.
        "sandbox-exec" => (&[], &["-f", "-n", "-p", "-D"], 0),
        // util-linux `flock [-sxun o] [-w time] [-E code] FILE command…`; the `-c` string is read
        // by `flock_string`, and the lock file is the one positional.
        "flock" => (
            &[
                "-s",
                "-x",
                "-u",
                "-n",
                "-o",
                "-F",
                "-v",
                "--shared",
                "--exclusive",
                "--unlock",
                "--nonblock",
                "--nb",
                "--close",
                "--no-fork",
                "--verbose",
            ],
            &[
                "-w",
                "-E",
                "-c",
                "--timeout",
                "--wait",
                "--conflict-exit-code",
                "--command",
            ],
            1,
        ),
        // BSD `script [-adefkqr] [-t time] [-F pipe] file command…` (the util-linux `-c` string
        // is read by `script_string`).
        "script" => (
            &[
                "-a", "-d", "-e", "-f", "-k", "-p", "-q", "-r", "-W", "--quiet", "--append",
            ],
            &["-t", "-F", "-I", "-O", "-B", "-T", "-E", "-m", "-o"],
            1,
        ),
        _ => return None,
    };
    Some(WrapSpec {
        flags,
        valued,
        positional,
    })
}

/// The index in `args` where the wrapped command starts.
fn peel(spec: &WrapSpec, args: &[Word]) -> Result<usize, Peel> {
    let mut i = 0;
    let mut positional = spec.positional;
    while i < args.len() {
        let w = &args[i];
        if w.dynamic {
            return Err(Peel::Dynamic);
        }
        let t = w.text();
        if t == "--" {
            return Ok(i + 1);
        }
        if w.is_option() {
            if spec.flags.contains(&t.as_str()) || t[1..].chars().all(|c| c.is_ascii_digit()) {
                i += 1;
            } else if spec.valued.contains(&t.as_str()) {
                i += 2;
            } else if (t.starts_with("--")
                && t.split_once('=')
                    .is_some_and(|(k, _)| spec.valued.contains(&k) || spec.flags.contains(&k)))
                || (!t.starts_with("--")
                    && spec.valued.iter().any(|v| v.len() == 2 && t.starts_with(v)))
            {
                i += 1;
            } else {
                return Err(Peel::Unknown);
            }
            continue;
        }
        if positional > 0 {
            positional -= 1;
            i += 1;
            continue;
        }
        break;
    }
    Ok(i.min(args.len()))
}

/// Full names of git options that are also the start of a longer option this file knows. git takes
/// the exact name before any prefix, so these never stand for the longer one (`--force` is its own
/// option of `checkout` and `switch`, not `--force-create`).
const EXACT_WINS: &[&str] = &["--force"];

/// The one rule for a long option written in a git command line (D-16): git takes any
/// unambiguous prefix of a long option name (`--roo` for `--root`, `--upload-pa=…` for
/// `--upload-pack=…`), so `typed` (the word, with or without an attached `=value`) matches `full`
/// when its name part is a non-empty start of it. A prefix that git finds ambiguous is an error
/// there, so matching every prefix never lets through a command git would run. The global options
/// before the subcommand are not prefix-matched by git and do not come here.
fn long_matches(typed: &str, full: &str) -> bool {
    let (Some(t), Some(f)) = (typed.strip_prefix("--"), full.strip_prefix("--")) else {
        return false;
    };
    let name = t.split_once('=').map_or(t, |(n, _)| n);
    if name.is_empty() || !f.starts_with(name) {
        return false;
    }
    name == f || !EXACT_WINS.contains(&format!("--{name}").as_str())
}

/// `typed` names `flag`: a short flag by its text, a long flag by [`long_matches`].
fn same_flag(typed: &str, flag: &str) -> bool {
    if flag.starts_with("--") {
        long_matches(typed, flag)
    } else {
        typed == flag
    }
}

/// A word that a substitution or expansion decides, but whose readable part already names the
/// option that carries a program, with the value written against it (`--upload-pack=$X`,
/// `-u"$X"`, `-qu"$X"`, `--upload-pa$X`): the program is unreadable, like the separate form.
fn attached_dynamic_value(literal: &str, shorts: &[char], longs: &[&str]) -> bool {
    if literal.starts_with("--") {
        longs.iter().any(|long| long_matches(literal, long))
    } else if literal.starts_with('-') && literal.len() > 1 {
        shorts.iter().any(|c| literal[1..].contains(*c))
    } else {
        false
    }
}

/// The values of the options that carry a command: a short letter (`-x CMD`, `-xCMD`, or a
/// cluster that ends in it) and a long name (`--exec CMD`, `--exec=CMD`, and any prefix of the
/// name, [`long_matches`]). A value that a substitution decides is unreadable, written apart from
/// the option or against it.
fn option_values(rest: &[Word], shorts: &[char], longs: &[&str], out: &mut Vec<String>) -> Verdict {
    let mut i = 0;
    while i < rest.len() {
        let w = &rest[i];
        i += 1;
        if w.dynamic {
            if attached_dynamic_value(&w.text(), shorts, longs) {
                return Err(Class::Unreadable);
            }
            continue;
        }
        let t = w.text();
        if t == "--" {
            break;
        }
        if !t.starts_with('-') || t.len() < 2 {
            continue;
        }
        let mut take_next = false;
        if t.starts_with("--") {
            for long in longs {
                if long_matches(&t, long) {
                    match t.split_once('=') {
                        Some((_, v)) => out.push(v.to_owned()),
                        None => take_next = true,
                    }
                }
            }
        } else {
            for c in shorts {
                if let Some(p) = t[1..].find(*c) {
                    let after = &t[1 + p + c.len_utf8()..];
                    if after.is_empty() {
                        take_next = true;
                    } else {
                        out.push(after.to_owned());
                    }
                }
            }
        }
        if take_next {
            match rest.get(i) {
                Some(v) if v.dynamic => return Err(Class::Unreadable),
                Some(v) => out.push(v.text()),
                None => {}
            }
            i += 1;
        }
    }
    Ok(())
}

fn xargs(args: &[Word]) -> Verdict {
    let mut i = 0;
    while i < args.len() {
        let w = &args[i];
        if w.dynamic {
            return Err(Class::Unreadable);
        }
        let t = w.text();
        if !w.is_option() {
            break;
        }
        if matches!(
            t.as_str(),
            "-I" | "-n"
                | "-P"
                | "-L"
                | "-s"
                | "-d"
                | "-E"
                | "-a"
                | "-J"
                | "-R"
                | "-S"
                | "--max-args"
                | "--max-procs"
                | "--delimiter"
                | "--arg-file"
        ) {
            i += 2;
        } else {
            i += 1;
        }
    }
    match args.get(i) {
        None => Ok(()),
        Some(w) if command_word_unreadable(w) => Err(Class::Unreadable),
        Some(w) if is_runner(basename(&w.text())) => Err(Class::Unreadable),
        Some(_) => Ok(()),
    }
}

// ------------------------------------------------------------------ git tables

/// Subcommands that move a branch or rewrite history; denied in the primary checkout (D-5).
const D5: &[&str] = &[
    "checkout",
    "switch",
    "reset",
    "rebase",
    "stash",
    "clean",
    "merge",
    "pull",
    "cherry-pick",
    "revert",
    "am",
];

#[derive(PartialEq, Eq, Clone, Copy)]
enum GitKind {
    Read,
    Write,
}

const GIT_READ: &[&str] = &[
    "status",
    "log",
    "diff",
    "show",
    "rev-parse",
    "rev-list",
    "ls-files",
    "ls-tree",
    "cat-file",
    "describe",
    "blame",
    "grep",
    "shortlog",
    "whatchanged",
    "name-rev",
    "merge-base",
    // Reads trees and writes objects only; it moves no ref and touches no working tree.
    "merge-tree",
    "for-each-ref",
    "show-ref",
    "ls-remote",
    "check-ignore",
    "check-attr",
    "var",
    "version",
    "help",
    "count-objects",
    "fsck",
    "verify-commit",
    "verify-tag",
    "diff-tree",
    "diff-index",
    "diff-files",
    "range-diff",
    "cherry",
    "show-branch",
    "annotate",
    "difftool",
    "archive",
    "bundle",
    "format-patch",
    "request-pull",
    "get-tar-commit-id",
    "column",
    "stripspace",
    "check-ref-format",
    "interpret-trailers",
    "mailinfo",
    "mailsplit",
    "check-mailmap",
    "fmt-merge-msg",
    "verify-pack",
    "show-index",
];
const GIT_WRITE: &[&str] = &[
    "add",
    "am",
    "apply",
    "bisect",
    "checkout",
    "cherry-pick",
    "clean",
    "clone",
    "commit",
    "fetch",
    "filter-branch",
    "gc",
    "init",
    "merge",
    "mv",
    "pull",
    "push",
    "rebase",
    "replace",
    "repack",
    "reset",
    "restore",
    "revert",
    "rm",
    "sparse-checkout",
    "submodule",
    "switch",
    "update-index",
    "update-ref",
    "write-tree",
    "commit-tree",
    "read-tree",
    "checkout-index",
    "mktag",
    "mktree",
    "fast-import",
    "unpack-objects",
    "index-pack",
    "maintenance",
    "prune",
    "pack-refs",
    "notes",
    "hash-object",
    "rerere",
    "mergetool",
    "stash",
];

/// Reads or writes by subcommand; an unknown subcommand (an alias or an external one) is
/// unreadable.
fn classify_git(name: &str, rest: &[Word]) -> Result<GitKind, Class> {
    let first_pos = || rest.iter().find(|w| !w.is_option()).map(Word::text);
    let any_flag = |flags: &[&str]| {
        rest.iter()
            .any(|w| !w.dynamic && flags.iter().any(|f| same_flag(&w.text(), f)))
    };
    let positionals = rest.iter().filter(|w| !w.is_option()).count();
    Ok(match name {
        "branch" => {
            let list_flag = any_flag(&[
                "-l",
                "--list",
                "-a",
                "-r",
                "-v",
                "-vv",
                "--show-current",
                "--contains",
                "--no-contains",
                "--merged",
                "--no-merged",
                "--points-at",
            ]);
            let write_flag = any_flag(&[
                "-d",
                "-D",
                "-m",
                "-M",
                "-c",
                "-C",
                "-f",
                "--force",
                "--delete",
                "--move",
                "--copy",
                "-u",
                "--set-upstream-to",
                "--unset-upstream",
                "--edit-description",
            ]);
            if write_flag || (positionals > 0 && !list_flag) {
                GitKind::Write
            } else {
                GitKind::Read
            }
        }
        "tag" => {
            if any_flag(&[
                "-l",
                "--list",
                "-n",
                "--contains",
                "--points-at",
                "--merged",
                "--no-merged",
                "-v",
                "--verify",
            ]) || positionals == 0
            {
                GitKind::Read
            } else {
                GitKind::Write
            }
        }
        "remote" => match first_pos().as_deref() {
            None | Some("show" | "get-url") => GitKind::Read,
            Some(_) => GitKind::Write,
        },
        "stash" => match first_pos().as_deref() {
            Some("list" | "show") => GitKind::Read,
            _ => GitKind::Write,
        },
        "worktree" => match first_pos().as_deref() {
            Some("list") => GitKind::Read,
            _ => GitKind::Write,
        },
        "reflog" => match first_pos().as_deref() {
            Some("expire" | "delete") => GitKind::Write,
            _ => GitKind::Read,
        },
        "symbolic-ref" => {
            if positionals >= 2 || any_flag(&["-d", "--delete"]) {
                GitKind::Write
            } else {
                GitKind::Read
            }
        }
        "config" => GitKind::Write,
        // git-lfs: the listing subcommands read; the ones that fill or change the working
        // tree, the attributes or the hooks write; any other name is not known.
        "lfs" => match first_pos().as_deref() {
            Some("version" | "env" | "ls-files" | "status" | "logs" | "locks" | "pointer") => {
                GitKind::Read
            }
            Some(
                "pull" | "fetch" | "checkout" | "prune" | "push" | "track" | "untrack" | "migrate"
                | "install" | "uninstall" | "update" | "lock" | "unlock",
            ) => GitKind::Write,
            _ => return Err(Class::Unreadable),
        },
        n if GIT_READ.contains(&n) => GitKind::Read,
        n if GIT_WRITE.contains(&n) => GitKind::Write,
        _ => return Err(Class::Unreadable),
    })
}

/// `-c key=value` and `--config-env=key=NAME`: keys that replace the hooks or run a program.
fn config_key_check(kv: &Word) -> Verdict {
    if kv.dynamic {
        return Err(Class::Unreadable);
    }
    let text = kv.text().to_lowercase();
    let key = text.split('=').next().unwrap_or("");
    if key == "core.hookspath" {
        return Err(Class::HookSkip);
    }
    let runs_program = key.starts_with("alias.")
        || key.starts_with("pager.")
        || key.starts_with("filter.")
        || key.starts_with("include")
        || key.ends_with(".driver")
        || key.ends_with(".textconv")
        || key.ends_with(".program")
        || matches!(
            key,
            "core.fsmonitor"
                | "core.pager"
                | "core.editor"
                | "core.sshcommand"
                | "core.askpass"
                | "core.gitproxy"
                | "core.worktree"
                | "core.gitdir"
                | "diff.external"
                | "credential.helper"
                | "sequence.editor"
        );
    if runs_program {
        return Err(Class::Unreadable);
    }
    Ok(())
}

/// The arguments of a git subcommand, split into options and positionals.
struct Parsed<'a> {
    /// Literal option words with their value, if the next word is the value.
    opts: Vec<(String, Option<String>)>,
    pos: Vec<&'a Word>,
    /// The words after a bare `--` (pathspecs or refs, by subcommand).
    tail: Vec<&'a Word>,
    /// A dynamic word outside a data slot.
    dynamic: bool,
}

/// Options whose next word is data (a message, a file, an author) rather than a decision.
fn data_slots(sub: &str) -> &'static [&'static str] {
    match sub {
        "commit" => &[
            "-m",
            "--message",
            "-F",
            "--file",
            "-C",
            "-c",
            "--reuse-message",
            "--reedit-message",
            "--author",
            "--date",
            "-t",
            "--template",
            "--fixup",
            "--squash",
            "--cleanup",
        ],
        "merge" | "pull" => &[
            "-m",
            "-F",
            "--file",
            "-s",
            "--strategy",
            "-X",
            "--strategy-option",
        ],
        "tag" => &["-m", "-F", "--message", "--file"],
        // The value after each of these is a program text, a strategy or a path, never a ref
        // (`--onto` is the new base; the branch to rebase is the last positional).
        "rebase" => &[
            "-x",
            "--exec",
            "-s",
            "--strategy",
            "-X",
            "--strategy-option",
            "-C",
            "--onto",
        ],
        "filter-branch" => &[
            "--env-filter",
            "--tree-filter",
            "--index-filter",
            "--parent-filter",
            "--msg-filter",
            "--commit-filter",
            "--tag-name-filter",
            "--setup",
            "--subdirectory-filter",
            "--original",
            "--state-branch",
            "-d",
        ],
        // The value of these is a type, a default, a file or a comment, not the key.
        "config" => &["--type", "--default", "-f", "--file", "--blob", "--comment"],
        "update-ref" => &["-m"],
        "push" => &["-o", "--push-option", "--repo", "--receive-pack", "--exec"],
        _ => &[],
    }
}

/// `typed` is the option of a slot (an option whose next word is its value): a short slot by its
/// text, a long slot by [`long_matches`] when no value is attached to it with `=`.
fn slot_named(typed: &str, slot: &str) -> bool {
    if slot.starts_with("--") {
        !typed.contains('=') && long_matches(typed, slot)
    } else {
        typed == slot
    }
}

/// Options whose next word is a branch name, for the ref checks.
const NAME_SLOTS: &[&str] = &[
    "-b",
    "-B",
    "-c",
    "-C",
    "--orphan",
    "--create",
    "--force-create",
];

impl<'a> Parsed<'a> {
    fn new(sub: &str, rest: &'a [Word]) -> Parsed<'a> {
        let slots = data_slots(sub);
        let mut p = Parsed {
            opts: Vec::new(),
            pos: Vec::new(),
            tail: Vec::new(),
            dynamic: false,
        };
        let mut i = 0;
        let mut after_dd = false;
        while i < rest.len() {
            let w = &rest[i];
            i += 1;
            if after_dd {
                p.tail.push(w);
                continue;
            }
            if w.dynamic {
                p.dynamic = true;
                p.pos.push(w);
                continue;
            }
            let t = w.text();
            if t == "--" {
                after_dd = true;
                continue;
            }
            if w.is_option() {
                if slots.iter().any(|s| slot_named(&t, s)) {
                    i += 1;
                    p.opts.push((t, None));
                } else if NAME_SLOTS.iter().any(|s| slot_named(&t, s))
                    && matches!(sub, "checkout" | "switch" | "worktree")
                {
                    let value = rest.get(i);
                    i += 1;
                    match value {
                        Some(v) if v.dynamic => {
                            p.dynamic = true;
                            p.opts.push((t, None));
                        }
                        Some(v) => p.opts.push((t, Some(v.text()))),
                        None => p.opts.push((t, None)),
                    }
                } else {
                    p.opts.push((t, None));
                }
            } else {
                p.pos.push(w);
            }
        }
        p
    }

    fn has(&self, flag: &str) -> bool {
        self.opts.iter().any(|(o, _)| same_flag(o, flag))
    }

    fn has_any(&self, flags: &[&str]) -> bool {
        flags.iter().any(|f| self.has(f))
    }

    fn values_of(&self, flag: &str) -> Vec<String> {
        let mut out: Vec<String> = Vec::new();
        for (o, v) in &self.opts {
            if flag.starts_with("--") {
                // A long flag or any prefix of it, with the value apart or against it.
                if long_matches(o, flag) {
                    match o.split_once('=') {
                        Some((_, attached)) => out.push(attached.to_owned()),
                        None => out.extend(v.clone()),
                    }
                }
                continue;
            }
            if o == flag {
                out.extend(v.clone());
            }
            // `-Bname` attached to the flag.
            if flag.len() == 2 && o.len() > 2 && o.starts_with(flag) && !o.starts_with("--") {
                out.push(o[2..].to_owned());
            }
            if let Some(v) = o.strip_prefix(&format!("{flag}=")) {
                out.push(v.to_owned());
            }
        }
        out
    }
}

/// Reading a config value is allowed; a write of the hooks path is a skip.
fn config_check(p: &Parsed) -> Verdict {
    let read_flag = p.has_any(&[
        "--get",
        "--get-all",
        "--get-regexp",
        "--get-urlmatch",
        "-l",
        "--list",
        "--show-origin",
        "--show-scope",
        "--name-only",
        "--get-color",
        "--get-colorbool",
    ]);
    let write_flag = p.has_any(&[
        "--add",
        "--replace-all",
        "--unset",
        "--unset-all",
        "--rename-section",
        "--remove-section",
        "-e",
        "--edit",
    ]);
    let writes = write_flag || (!read_flag && p.pos.len() >= 2);
    if !writes {
        return Ok(());
    }
    if p.dynamic {
        return Err(Class::Unreadable);
    }
    // The hooks path, and the include keys that pull in another file that may set it.
    if p.pos.iter().any(|w| {
        let key = w.text().to_lowercase();
        key.contains("hookspath") || key.starts_with("include.") || key.starts_with("includeif.")
    }) {
        return Err(Class::HookSkip);
    }
    Ok(())
}

/// Options that skip the commit hooks, in every spelling git accepts.
fn hook_checks(name: &str, p: &Parsed, rest: &[Word]) -> Verdict {
    const HOOK_RUNNERS: [&str; 8] = [
        "commit",
        "merge",
        "push",
        "pull",
        "am",
        "cherry-pick",
        "revert",
        "rebase",
    ];
    if !HOOK_RUNNERS.contains(&name) {
        return Ok(());
    }
    if p.dynamic {
        return Err(Class::Unreadable);
    }
    for (o, _) in &p.opts {
        // Any prefix of `--no-verify` that no other option shares (`--no-v` and longer).
        if o.starts_with("--no-v") && long_matches(o, "--no-verify") {
            return Err(Class::HookSkip);
        }
    }
    if name == "commit" {
        // Short options: `-n` skips the hooks; letters after a value-taking one are its value.
        let slots = data_slots("commit");
        let mut at = 0;
        while at < rest.len() {
            let w = &rest[at];
            at += 1;
            if w.dynamic {
                continue;
            }
            let t = w.text();
            if t == "--" {
                break;
            }
            if slots.contains(&t.as_str()) {
                // The next word is the value of this option, never an option itself.
                at += 1;
                continue;
            }
            if !w.is_option() || t.starts_with("--") {
                continue;
            }
            for c in t[1..].chars() {
                if c == 'n' {
                    return Err(Class::HookSkip);
                }
                if slots.contains(&format!("-{c}").as_str()) || matches!(c, 'u' | 'S') {
                    break;
                }
            }
        }
    }
    Ok(())
}

/// D-16: the one rule for a long git option written as a prefix of its name (AC-16).
#[cfg(test)]
mod long_option_rule_tests {
    use super::{attached_dynamic_value, long_matches, same_flag, slot_named};

    #[test]
    fn a_prefix_of_the_name_matches_with_or_without_an_attached_value() {
        assert!(long_matches("--root", "--root"));
        assert!(long_matches("--roo", "--root"));
        assert!(long_matches("--r", "--root"));
        assert!(long_matches("--upload-pa=git-upload-pack", "--upload-pack"));
        assert!(!long_matches("--rootx", "--root"));
        assert!(!long_matches("--onto", "--root"));
    }

    #[test]
    fn a_lone_dash_pair_a_short_flag_and_a_missing_name_match_nothing() {
        assert!(!long_matches("--", "--root"));
        assert!(!long_matches("--=x", "--root"));
        assert!(!long_matches("-r", "--root"));
        assert!(!long_matches("--root", "-r"));
        assert!(!long_matches("root", "--root"));
    }

    #[test]
    fn the_exact_name_of_another_option_wins_over_a_longer_one() {
        assert!(!long_matches("--force", "--force-create"));
        assert!(long_matches("--force-cr", "--force-create"));
        assert!(long_matches("--force", "--force"));
    }

    #[test]
    fn a_short_flag_is_compared_by_its_text_only() {
        assert!(same_flag("-d", "-d"));
        assert!(!same_flag("-dd", "-d"));
        assert!(same_flag("--detac", "--detach"));
    }

    #[test]
    fn a_slot_takes_its_value_from_the_next_word_only_without_an_equals_sign() {
        assert!(slot_named("--exe", "--exec"));
        assert!(!slot_named("--exe=x", "--exec"));
        assert!(slot_named("-x", "-x"));
        assert!(!slot_named("-xy", "-x"));
    }

    #[test]
    fn a_dynamic_word_that_names_a_program_option_with_the_value_against_it_is_found() {
        let longs = ["--upload-pack"];
        assert!(attached_dynamic_value("--upload-pack=", &[], &longs));
        assert!(attached_dynamic_value("--upload-pa", &[], &longs));
        assert!(attached_dynamic_value("-u", &['u'], &longs));
        assert!(attached_dynamic_value("-qu", &['u'], &longs));
        assert!(!attached_dynamic_value("--depth=", &['u'], &longs));
        assert!(!attached_dynamic_value("-q", &['u'], &longs));
        assert!(!attached_dynamic_value("--", &['u'], &longs));
        assert!(!attached_dynamic_value("", &['u'], &longs));
        assert!(!attached_dynamic_value("x-u", &['u'], &longs));
    }
}
