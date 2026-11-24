//! Project detection for the `/moai project` route (Issue #2496; user decisions 16 = B and
//! 17 = all three, #1975 comment 6815357071).
//!
//! `project-detect` reads the project folder and writes nothing. It tells a new project from
//! an existing one (a manifest or a code file makes it existing), names the languages and the
//! language servers they call for, the MCP servers the project already configures, and the
//! databases its dependencies name. Only project facts are recorded: no HOME read, no process
//! start, because `.moai-cli/project.yaml` is tracked. The one machine fact, `lsp_plugins`
//! (SPEC-20261123-lsp-official-plugins), looks at the file names of the PATH folders for each
//! language's official LSP server and is reported beside `detected`, never recorded. The `versioning`
//! fact (card C-69) is reported beside `detected` and never recorded; it reads the project's
//! tag names from the files of its own `.git` and starts no program either. MCP configuration is
//! read for server names only and is never created or changed (decision 21 keeps MCP itself
//! on HOLD). The result carries a complete profile for `project-plan`; `index-build` is only
//! suggested, never run.
//!
//! Manifests below the root (card C-61, `plugins/<name>/package.json` and its like) are found
//! by name to a bounded folder depth; a nested manifest's bytes are never read, so the
//! dependencies that name `databases` stay a root-only fact (#2535). The tools those names
//! imply are reported beside `detected` for the tools question.

use crate::PROJECT_STATE_DIR;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{
    collections::{BTreeMap, BTreeSet, VecDeque},
    fs,
    path::{Path, PathBuf},
    rc::Rc,
};

type Result<T> = std::result::Result<T, String>;

/// The `detected` profile value. Every list is sorted and free of duplicates, so a re-run on
/// the same files yields the same bytes.
#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Detected {
    /// `new` (no manifest, no code file) or `existing`.
    pub project_kind: String,
    /// Manifest paths found for the project: a root manifest by its file name, one below the
    /// root by its path relative to the root (C-61, to [`MAX_DEPTH`] folder levels).
    pub manifests: Vec<String>,
    pub languages: Vec<String>,
    /// Language servers the languages call for; not a statement that one is installed.
    pub lsp: Vec<String>,
    /// MCP server names configured in the project's own files.
    pub mcp_servers: Vec<String>,
    /// Database kinds named by dependencies or a schema file.
    pub databases: Vec<String>,
}

/// The manifest catalog (SPEC-20261124-project-detect-monorepo): which file names and name
/// suffixes are manifests, the language and tools each implies, and the folder names the walks
/// skip. It is data, not code: the engine below knows no manifest name. A project adds rows in
/// its own file ([`EXTENSION_FILE`]); a row never replaces a shipped one.
const CATALOG_JSON: &str = include_str!("project_detect_catalog.json");

/// The project's own catalog rows, under the `config/` folder of its state folder.
const EXTENSION_FILE: &str = "config/detect-catalog.json";

/// Folders that hold MoAI's own state or a host's files, never the project's manifests or code.
/// They are not tied to a language; every other name to skip comes from the catalog.
const HOST_AND_STATE_DIRS: &[&str] = &[
    ".git",
    PROJECT_STATE_DIR,
    crate::state_root::MOAI_DIR,
    ".claude",
    ".codex",
    ".omp",
];

/// One catalog file: the shipped one and a project's extension share this shape.
#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
struct CatalogFile {
    schema: u32,
    #[serde(default)]
    manifests: Vec<CatalogRow>,
    #[serde(default)]
    skip_dirs: Vec<String>,
}

/// One ecosystem: the manifest names and suffixes it uses, the language and tools they imply.
#[derive(Clone, Deserialize)]
#[serde(deny_unknown_fields)]
struct CatalogRow {
    ecosystem: String,
    language: String,
    #[serde(default)]
    files: Vec<String>,
    #[serde(default)]
    suffixes: Vec<String>,
    #[serde(default)]
    tools: Vec<String>,
}

/// The catalog a detection reads: the shipped rows first, then the rows a project added.
#[derive(Clone, Default)]
pub(crate) struct Catalog {
    rows: Vec<CatalogRow>,
    skip_dirs: Vec<String>,
}

/// What the project's extension file did, reported beside `detected` (`manifest_catalog`).
#[derive(Clone, Debug, PartialEq, Eq)]
pub(crate) struct ExtensionReport {
    /// `absent`, `read` or `refused`.
    pub extension: &'static str,
    /// Rows the extension file holds (0 unless it was read).
    pub rows: usize,
    /// Why the file was refused; only its rows were left out.
    pub problem: Option<String>,
}

impl ExtensionReport {
    fn absent() -> Self {
        Self {
            extension: "absent",
            rows: 0,
            problem: None,
        }
    }

    fn refused(problem: impl Into<String>) -> Self {
        Self {
            extension: "refused",
            rows: 0,
            problem: Some(problem.into()),
        }
    }

    pub(crate) fn to_json(&self) -> Value {
        json!({"extension": self.extension, "rows": self.rows, "problem": self.problem})
    }
}

/// The rows every value of which can be recorded and that name at least one file or suffix.
fn catalog_file_valid(file: &CatalogFile) -> Result<()> {
    if file.schema != 1 {
        return Err("schema must be 1".into());
    }
    if file.manifests.len() > MAX_ITEMS || file.skip_dirs.len() > MAX_ITEMS {
        return Err(format!("more than {MAX_ITEMS} rows or folders"));
    }
    for row in &file.manifests {
        let names = [&row.ecosystem, &row.language]
            .into_iter()
            .chain(&row.files)
            .chain(&row.tools);
        let suffixes_ok = row
            .suffixes
            .iter()
            .all(|s| s.len() > 1 && s.starts_with('.') && recordable(s));
        let lists_ok = row.files.len() <= MAX_ITEMS
            && row.suffixes.len() <= MAX_ITEMS
            && row.tools.len() <= MAX_ITEMS;
        if !names.into_iter().all(|n| recordable(n))
            || !suffixes_ok
            || !lists_ok
            || (row.files.is_empty() && row.suffixes.is_empty())
        {
            return Err("a row holds a value that cannot be recorded, or names no file".into());
        }
    }
    if !file.skip_dirs.iter().all(|d| recordable(d)) {
        return Err("a folder name cannot be recorded".into());
    }
    Ok(())
}

/// The languages the shipped catalog's manifest rows name, sorted and without repeats.
pub fn shipped_manifest_languages() -> Vec<String> {
    let languages: BTreeSet<String> = shipped().rows.iter().map(|r| r.language.clone()).collect();
    languages.into_iter().collect()
}

/// The shipped catalog, parsed once. A parse failure leaves it empty; a unit test pins that the
/// shipped file parses and holds the rows the earlier tables held.
fn shipped() -> &'static Catalog {
    static SHIPPED: std::sync::OnceLock<Catalog> = std::sync::OnceLock::new();
    SHIPPED.get_or_init(|| {
        serde_json::from_str::<CatalogFile>(CATALOG_JSON)
            .ok()
            .filter(|f| catalog_file_valid(f).is_ok())
            .map(|f| Catalog {
                rows: f.manifests,
                skip_dirs: f.skip_dirs,
            })
            .unwrap_or_default()
    })
}

impl Catalog {
    /// The shipped catalog and, when the project has one, its extension rows. The extension is
    /// read as a regular file only (a link is never followed) of at most [`MAX_FILE`] bytes; a
    /// file that cannot be used is reported and none of its rows is taken.
    pub(crate) fn for_project(root: &Path) -> (Catalog, ExtensionReport) {
        let mut catalog = shipped().clone();
        let state = crate::state_root::dir(root);
        let parts: Vec<&str> = std::iter::once(state)
            .chain(EXTENSION_FILE.split('/'))
            .collect();
        let report = match file_in(root, &parts) {
            Some((path, meta)) if meta.len() <= MAX_FILE => {
                match fs::read_to_string(path)
                    .map_err(|_| "the file is not valid UTF-8".to_string())
                    .and_then(|text| {
                        serde_json::from_str::<CatalogFile>(&text)
                            .map_err(|_| "the file is not a valid catalog".to_string())
                    })
                    .and_then(|file| catalog_file_valid(&file).map(|_| file))
                {
                    Ok(file) => {
                        let rows = file.manifests.len();
                        catalog.add_extension(file);
                        ExtensionReport {
                            extension: "read",
                            rows,
                            problem: None,
                        }
                    }
                    Err(problem) => ExtensionReport::refused(problem),
                }
            }
            Some(_) => ExtensionReport::refused("the file is larger than the file limit"),
            None => {
                if fs::symlink_metadata(root.join(state).join(EXTENSION_FILE)).is_ok() {
                    ExtensionReport::refused(
                        "the file is not a regular file (a link is not followed)",
                    )
                } else {
                    ExtensionReport::absent()
                }
            }
        };
        (catalog, report)
    }

    /// Adds a project's rows: a file name or suffix a shipped row already has stays with the
    /// shipped row, and a row left with no name is dropped.
    fn add_extension(&mut self, file: CatalogFile) {
        for mut row in file.manifests {
            row.files
                .retain(|f| !self.rows.iter().any(|r| r.files.contains(f)));
            row.suffixes
                .retain(|s| !self.rows.iter().any(|r| r.suffixes.contains(s)));
            if !(row.files.is_empty() && row.suffixes.is_empty()) {
                self.rows.push(row);
            }
        }
        for dir in file.skip_dirs {
            if !self.skip_dirs.contains(&dir) {
                self.skip_dirs.push(dir);
            }
        }
    }

    /// The row a file name belongs to: an exact name first, then a suffix.
    fn row_for(&self, name: &str) -> Option<&CatalogRow> {
        self.rows
            .iter()
            .find(|r| r.files.iter().any(|f| f == name))
            .or_else(|| {
                self.rows.iter().find(|r| {
                    r.suffixes
                        .iter()
                        .any(|s| name.len() > s.len() && name.ends_with(s.as_str()))
                })
            })
    }

    fn skips(&self, name: &str) -> bool {
        self.skip_dirs.iter().any(|d| d == name)
    }
}

/// Whether the code walks skip a folder of this name: the host and state folders and the shipped
/// catalog's names.
pub(crate) fn skip_dir(name: &str) -> bool {
    HOST_AND_STATE_DIRS.contains(&name) || shipped().skips(name)
}

/// Code file extension → language.
const EXTENSIONS: &[(&str, &str)] = &[
    ("rs", "rust"),
    ("py", "python"),
    ("js", "javascript"),
    ("jsx", "javascript"),
    ("ts", "typescript"),
    ("tsx", "typescript"),
    ("go", "go"),
    ("java", "java"),
    ("kt", "kotlin"),
    ("rb", "ruby"),
    ("php", "php"),
    ("c", "c"),
    ("h", "c"),
    ("cpp", "cpp"),
    ("cs", "csharp"),
    ("swift", "swift"),
    ("lua", "lua"),
];

/// The marketplace of the official Claude Code LSP plugins.
const LSP_MARKETPLACE: &str = "claude-plugins-official";

/// Language → (the server name recorded in `detected.lsp`, the official LSP plugin of that
/// marketplace, the executable the plugin runs from PATH). Read 2026-11-23 from the marketplace's
/// own `plugin.json` files (SPEC-20261123-lsp-official-plugins).
const LSP_PLUGINS: &[(&str, &str, &str, &str)] = &[
    (
        "rust",
        "rust-analyzer",
        "rust-analyzer-lsp",
        "rust-analyzer",
    ),
    ("python", "pyright", "pyright-lsp", "pyright-langserver"),
    (
        "javascript",
        "typescript-language-server",
        "typescript-lsp",
        "typescript-language-server",
    ),
    (
        "typescript",
        "typescript-language-server",
        "typescript-lsp",
        "typescript-language-server",
    ),
    ("go", "gopls", "gopls-lsp", "gopls"),
    ("java", "jdtls", "jdtls-lsp", "jdtls"),
    ("kotlin", "kotlin-lsp", "kotlin-lsp", "kotlin-lsp"),
    ("ruby", "ruby-lsp", "ruby-lsp", "ruby-lsp"),
    ("php", "intelephense", "php-lsp", "intelephense"),
    ("c", "clangd", "clangd-lsp", "clangd"),
    ("cpp", "clangd", "clangd-lsp", "clangd"),
    ("csharp", "csharp-ls", "csharp-lsp", "csharp-ls"),
    ("swift", "sourcekit-lsp", "swift-lsp", "sourcekit-lsp"),
    (
        "lua",
        "lua-language-server",
        "lua-lsp",
        "lua-language-server",
    ),
];

/// Dependency name → database kind (exact names, compared lower-case).
const DB_DEPENDENCIES: &[(&str, &str)] = &[
    ("postgres", "postgres"),
    ("tokio-postgres", "postgres"),
    ("pg", "postgres"),
    ("psycopg", "postgres"),
    ("psycopg2", "postgres"),
    ("psycopg2-binary", "postgres"),
    ("asyncpg", "postgres"),
    ("github.com/jackc/pgx/v5", "postgres"),
    ("github.com/lib/pq", "postgres"),
    ("mysql", "mysql"),
    ("mysql2", "mysql"),
    ("pymysql", "mysql"),
    ("mysqlclient", "mysql"),
    ("github.com/go-sql-driver/mysql", "mysql"),
    ("rusqlite", "sqlite"),
    ("sqlite3", "sqlite"),
    ("better-sqlite3", "sqlite"),
    ("mongodb", "mongodb"),
    ("pymongo", "mongodb"),
    ("mongoose", "mongodb"),
    ("redis", "redis"),
    ("ioredis", "redis"),
];

/// Walk bounds: folder depth below the root and directory entries looked at.
pub const MAX_DEPTH: usize = 3;
pub const MAX_ENTRIES: usize = 5000;
/// Largest manifest or configuration file read.
const MAX_FILE: u64 = 256 * 1024;
/// Most bytes of `.gitignore` files below the root one manifest walk reads in all.
const MAX_IGNORE_TOTAL: u64 = 4 * MAX_FILE;
/// Largest list kept in `detected`.
const MAX_ITEMS: usize = 64;

/// The regular file at `root/<parts…>`, checked component by component: every folder on the
/// way must be a real folder and the last part a regular file, none of them a symbolic link
/// (#2496 pre-audit B1: a symlinked middle folder such as `prisma -> /elsewhere` must not lead
/// outside). A change between this check and the read is not guarded against.
fn file_in(root: &Path, parts: &[&str]) -> Option<(PathBuf, fs::Metadata)> {
    let (last, folders) = parts.split_last()?;
    let mut path = root.to_path_buf();
    for folder in folders {
        path.push(folder);
        let meta = fs::symlink_metadata(&path).ok()?;
        if !meta.file_type().is_dir() {
            return None;
        }
    }
    path.push(last);
    let meta = fs::symlink_metadata(&path).ok()?;
    if !meta.file_type().is_file() {
        return None;
    }
    Some((path, meta))
}

/// The text of [`file_in`], at most `MAX_FILE` bytes.
pub(crate) fn read_in(root: &Path, parts: &[&str]) -> Option<String> {
    let (path, meta) = file_in(root, parts)?;
    if meta.len() > MAX_FILE {
        return None;
    }
    fs::read_to_string(path).ok()
}

pub(crate) fn is_file_in(root: &Path, parts: &[&str]) -> bool {
    file_in(root, parts).is_some()
}

/// A name that may be recorded: short, printable, no path separators, no spaces, and not dots
/// alone (`.` and `..` name a folder step, never a file, language or tool).
fn recordable(name: &str) -> bool {
    !name.is_empty()
        && name.len() <= 80
        && !name.bytes().all(|b| b == b'.')
        && name
            .bytes()
            .all(|b| b.is_ascii_alphanumeric() || matches!(b, b'-' | b'_' | b'.' | b'@'))
}

/// A path `detected.manifests` may hold (C-61): relative, `/`-separated, every component
/// [`recordable`], no `..` or `.` step, no backslash, no empty or trailing separator, at most
/// 160 bytes. A root manifest is the bare file name — one component — so old values stay valid.
fn recordable_relpath(path: &str) -> bool {
    !path.is_empty()
        && path.len() <= 160
        && !path.starts_with('/')
        && !path.ends_with('/')
        && !path.contains('\\')
        && !path.contains("//")
        && path
            .split('/')
            .all(|c| c != ".." && c != "." && recordable(c))
}

/// Languages of the code files below `root`; empty when the walk found no code file.
fn code_languages(root: &Path) -> BTreeSet<String> {
    let mut found = BTreeSet::new();
    let mut seen = 0usize;
    let mut stack: Vec<(PathBuf, usize)> = vec![(root.to_path_buf(), 0)];
    while let Some((dir, depth)) = stack.pop() {
        let Ok(entries) = fs::read_dir(&dir) else {
            continue;
        };
        let mut entries: Vec<_> = entries.flatten().collect();
        entries.sort_by_key(|e| e.file_name());
        for entry in entries {
            seen += 1;
            if seen > MAX_ENTRIES {
                return found;
            }
            let Ok(kind) = entry.file_type() else {
                continue;
            };
            let name = entry.file_name();
            let name = name.to_string_lossy();
            if kind.is_dir() {
                if depth < MAX_DEPTH && !skip_dir(&name) {
                    stack.push((entry.path(), depth + 1));
                }
            } else if kind.is_file() {
                if let Some(ext) = name.rsplit_once('.').map(|(_, e)| e) {
                    if let Some((_, lang)) = EXTENSIONS.iter().find(|(e, _)| *e == ext) {
                        found.insert((*lang).to_string());
                    }
                }
            }
        }
    }
    found
}

/// One line of an ignore list in `.gitignore` syntax, as `worktree_include` reads it. Its
/// matcher is module-private and outside this card's write scope, so the semantics are restated
/// here on purpose (SPEC-20261119-nested-manifests D-1); unifying them is the leader's call.
struct IgnoreLine {
    negate: bool,
    dir_only: bool,
    anchored: bool,
    glob: String,
}

/// The lines of an ignore-list file: blanks and comments skipped, `!` a negation, a trailing
/// `/` a folder-only pattern, `\#` and `\!` literal first characters.
fn ignore_lines(text: &str) -> Vec<IgnoreLine> {
    let mut lines = Vec::new();
    for raw in text.lines() {
        let trimmed = raw.trim_end_matches('\r').trim_end();
        if trimmed.is_empty() || trimmed.starts_with('#') {
            continue;
        }
        let (negate, body) = match trimmed.strip_prefix('!') {
            Some(rest) => (true, rest),
            None => (false, trimmed),
        };
        let body = body
            .strip_prefix('\\')
            .filter(|b| b.starts_with('#') || b.starts_with('!'))
            .unwrap_or(body);
        let dir_only = body.ends_with('/');
        let body = body.trim_end_matches('/');
        if body.is_empty() {
            continue;
        }
        lines.push(IgnoreLine {
            negate,
            dir_only,
            anchored: body.contains('/'),
            glob: body.trim_start_matches('/').to_string(),
        });
    }
    lines
}

/// `.gitignore` glob over a `/`-separated path: `**` crosses folders, `*` and `?` stay inside
/// one name, `[...]` is a character class.
fn glob_match(pattern: &[u8], text: &[u8]) -> bool {
    match pattern.first() {
        None => text.is_empty(),
        Some(b'*') if pattern.get(1) == Some(&b'*') => {
            let rest = &pattern[2..];
            let rest = rest.strip_prefix(b"/").unwrap_or(rest);
            (0..=text.len()).any(|i| {
                (i == 0 || text[i - 1] == b'/') && glob_match(rest, &text[i..]) || rest.is_empty()
            })
        }
        Some(b'*') => (0..=text.len())
            .take_while(|&i| i == 0 || text[i - 1] != b'/')
            .any(|i| glob_match(&pattern[1..], &text[i..])),
        Some(b'?') => !text.is_empty() && text[0] != b'/' && glob_match(&pattern[1..], &text[1..]),
        Some(b'[') => {
            let Some(end) = pattern.iter().skip(1).position(|&c| c == b']') else {
                return text.first() == Some(&b'[') && glob_match(&pattern[1..], &text[1..]);
            };
            let class = &pattern[1..1 + end];
            let (negated, class) = match class.first() {
                Some(b'!') | Some(b'^') => (true, &class[1..]),
                _ => (false, class),
            };
            let Some(&c) = text.first() else {
                return false;
            };
            let mut hit = false;
            let mut j = 0;
            while j < class.len() {
                if j + 2 < class.len() && class[j + 1] == b'-' {
                    hit |= class[j] <= c && c <= class[j + 2];
                    j += 3;
                } else {
                    hit |= class[j] == c;
                    j += 1;
                }
            }
            hit != negated && c != b'/' && glob_match(&pattern[2 + end..], &text[1..])
        }
        Some(&p) => text.first() == Some(&p) && glob_match(&pattern[1..], &text[1..]),
    }
}

/// Whether `line`'s pattern matches `path` itself (`is_file` false for a folder), judged
/// without case, as git's default on macOS is (`core.ignorecase`), like `worktree_include`.
fn ignore_matches_at(line: &IgnoreLine, path: &str, is_file: bool) -> bool {
    if line.dir_only && is_file {
        return false;
    }
    let glob = line.glob.to_ascii_lowercase();
    let path = path.to_ascii_lowercase();
    if line.anchored {
        glob_match(glob.as_bytes(), path.as_bytes())
    } else {
        let name = path.rsplit('/').next().unwrap_or(&path);
        glob_match(glob.as_bytes(), name.as_bytes())
    }
}

/// The last line of `lines` matching `path` itself (`is_file` false for a folder). The walk
/// never enters a folder its list excludes, so "a parent directory is excluded" needs no
/// shortcut here — the file under an excluded folder is simply never reached, and a deeper
/// `.gitignore` may re-include a folder a shallower line excluded.
fn last_matching<'a>(lines: &'a [IgnoreLine], path: &str, is_file: bool) -> Option<&'a IgnoreLine> {
    lines
        .iter()
        .rev()
        .find(|l| ignore_matches_at(l, path, is_file))
}

/// The ignore lists that apply to one folder of the manifest walk: its own list and, through
/// `parent`, the list of every folder above it, each judging paths from its own folder down
/// (`base` is that folder's depth in components). The deepest list with a matching line wins,
/// as git reads nested `.gitignore` files. A chain is shared by the folders below it, because
/// the walk reads a whole depth before the next one.
struct IgnoreNode {
    base: usize,
    lines: Vec<IgnoreLine>,
    parent: Option<Rc<IgnoreNode>>,
}

impl IgnoreNode {
    /// Whether the folder or file at `parts` (components from the project root) is excluded.
    fn excludes(&self, parts: &[&str], is_file: bool) -> bool {
        let mut node = Some(self);
        while let Some(n) = node {
            if n.base < parts.len() {
                let rel = parts[n.base..].join("/");
                if let Some(l) = last_matching(&n.lines, &rel, is_file) {
                    return !l.negate;
                }
            }
            node = n.parent.as_deref();
        }
        false
    }
}

/// The ignore lines that apply at the project root: `.git/info/exclude` first (the lower
/// precedence), then the root `.gitignore`; `.git` is read only when it is a real folder.
fn root_ignore_lines(root: &Path) -> Vec<IgnoreLine> {
    let mut lines = Vec::new();
    if fs::symlink_metadata(root.join(".git")).is_ok_and(|m| m.file_type().is_dir()) {
        lines.extend(
            read_in(root, &[".git", "info", "exclude"])
                .map(|t| ignore_lines(&t))
                .unwrap_or_default(),
        );
    }
    lines.extend(
        read_in(root, &[".gitignore"])
            .map(|t| ignore_lines(&t))
            .unwrap_or_default(),
    );
    lines
}

/// The manifests a walk found and whether it read every folder it meant to.
pub(crate) struct ManifestWalk {
    pub found: BTreeSet<String>,
    /// False when the walk stopped at [`MAX_ENTRIES`] folder entries or could not read a folder
    /// or an entry of one.
    pub complete: bool,
}

/// Manifest files below the root (C-61 AC-1): folder depth at most [`MAX_DEPTH`], names only —
/// a nested manifest's bytes are never read — no symbolic link followed (the walk looks at
/// `DirEntry::file_type`, which does not follow one). Every folder of one depth is read before
/// any folder one level deeper, so a large early folder cannot use the budget up before a later
/// package is reached. Folders whose name starts with a dot (tool, cache and backup state), the
/// catalog's folder names and folders the ignore lists exclude are never entered; a manifest
/// file at the root is listed whatever the ignore lists say. An ignore list that cannot be read
/// (a link, over [`MAX_FILE`], not UTF-8, past the [`MAX_IGNORE_TOTAL`] the walk reads in all)
/// counts as no lines, so that failure can only read more, never less. A folder or an entry
/// the walk cannot read is left out and the walk is not complete. The `seen` budget mirrors the
/// code walk's [`MAX_ENTRIES`].
pub(crate) fn walk_manifests(root: &Path, catalog: &Catalog) -> ManifestWalk {
    let mut found = BTreeSet::new();
    let mut seen = 0usize;
    let mut unreadable = false;
    let mut ignore_budget = MAX_IGNORE_TOTAL;
    let top = Rc::new(IgnoreNode {
        base: 0,
        lines: root_ignore_lines(root),
        parent: None,
    });
    // Each queued folder holds the list chain of its parent; its own list is read when the
    // folder is taken from the queue, so only the lists of folders being walked are in memory.
    let mut queue: VecDeque<(PathBuf, Vec<String>, Rc<IgnoreNode>)> =
        VecDeque::from([(root.to_path_buf(), Vec::new(), top)]);
    while let Some((dir, rel, parent)) = queue.pop_front() {
        let ignores = if rel.is_empty() {
            parent
        } else {
            Rc::new(IgnoreNode {
                base: rel.len(),
                lines: budgeted_ignore_lines(&dir, &mut ignore_budget),
                parent: Some(parent),
            })
        };
        let Ok(entries) = fs::read_dir(&dir) else {
            unreadable = true;
            continue;
        };
        let mut entries: Vec<_> = entries
            .filter_map(|e| {
                if e.is_err() {
                    unreadable = true;
                }
                e.ok()
            })
            .collect();
        entries.sort_by_key(|e| e.file_name());
        for entry in entries {
            seen += 1;
            if seen > MAX_ENTRIES {
                return ManifestWalk {
                    found,
                    complete: false,
                };
            }
            let Ok(kind) = entry.file_type() else {
                unreadable = true;
                continue;
            };
            let name = entry.file_name();
            let name = name.to_string_lossy();
            let mut parts = rel.clone();
            parts.push(name.to_string());
            let parts_ref: Vec<&str> = parts.iter().map(String::as_str).collect();
            if kind.is_dir() {
                if name.starts_with('.')
                    || catalog.skips(&name)
                    || HOST_AND_STATE_DIRS.contains(&name.as_ref())
                    || ignores.excludes(&parts_ref, false)
                {
                    continue;
                }
                if rel.len() < MAX_DEPTH {
                    queue.push_back((entry.path(), parts, Rc::clone(&ignores)));
                }
            } else if kind.is_file()
                && catalog.row_for(&name).is_some()
                && (rel.is_empty() || !ignores.excludes(&parts_ref, true))
            {
                found.insert(parts.join("/"));
            }
        }
    }
    ManifestWalk {
        found,
        complete: !unreadable,
    }
}

/// The lines of the `.gitignore` in `dir`, counted against `budget` (the bytes of ignore lists
/// still allowed, at most [`MAX_IGNORE_TOTAL`] for one walk). A list that cannot be read, is
/// over [`MAX_FILE`] or would pass the budget counts as no lines, so that can only let more
/// manifests in, never hide one.
fn budgeted_ignore_lines(dir: &Path, budget: &mut u64) -> Vec<IgnoreLine> {
    let Some((path, meta)) = file_in(dir, &[".gitignore"]) else {
        return Vec::new();
    };
    if meta.len() > MAX_FILE || meta.len() > *budget {
        return Vec::new();
    }
    let Ok(text) = fs::read_to_string(path) else {
        return Vec::new();
    };
    *budget -= meta.len();
    ignore_lines(&text)
}

/// [`walk_manifests`] over the project's own catalog, for callers that need the paths alone.
pub(crate) fn nested_manifests(root: &Path) -> BTreeSet<String> {
    walk_manifests(root, &Catalog::for_project(root).0).found
}

/// The tools the detected manifest names imply (C-61 AC-2), beside `detected` and never inside
/// it: tool name → the sorted manifest paths that imply it, at most [`MAX_ITEMS`] tools. The
/// paths come from `detected.manifests`, so they are already capped and validated.
fn tools_fact(manifests: &[String], catalog: &Catalog) -> BTreeMap<String, Vec<String>> {
    let mut map: BTreeMap<String, BTreeSet<String>> = BTreeMap::new();
    for path in manifests {
        let name = path.rsplit('/').next().unwrap_or(path);
        if let Some(row) = catalog.row_for(name) {
            for tool in &row.tools {
                map.entry(tool.clone()).or_default().insert(path.clone());
            }
        }
    }
    map.into_iter()
        .take(MAX_ITEMS)
        .map(|(tool, paths)| (tool, paths.into_iter().collect()))
        .collect()
}

fn toml_doc(text: &str) -> Option<toml_edit::DocumentMut> {
    text.parse::<toml_edit::DocumentMut>().ok()
}

fn table_keys(item: Option<&toml_edit::Item>) -> Vec<String> {
    item.and_then(toml_edit::Item::as_table_like)
        .map(|t| t.iter().map(|(k, _)| k.to_string()).collect())
        .unwrap_or_default()
}

/// The package name at the front of a requirement such as `psycopg[binary]>=3; python_version…`.
fn requirement_name(spec: &str) -> String {
    spec.trim()
        .split(|c: char| !(c.is_ascii_alphanumeric() || matches!(c, '-' | '_' | '.')))
        .next()
        .unwrap_or("")
        .to_ascii_lowercase()
}

/// Dependency names the Python files at the root declare: `pyproject.toml` (`[project]` and
/// Poetry dependencies) and `requirements.txt`. Names only, as the requirement names carry
/// them.
fn python_dependency_names(root: &Path) -> BTreeSet<String> {
    let mut names = BTreeSet::new();
    if let Some(doc) = read_in(root, &["pyproject.toml"]).and_then(|t| toml_doc(&t)) {
        if let Some(list) = doc
            .get("project")
            .and_then(|p| p.get("dependencies"))
            .and_then(toml_edit::Item::as_array)
        {
            names.extend(list.iter().filter_map(|v| v.as_str()).map(requirement_name));
        }
        names.extend(table_keys(
            doc.get("tool")
                .and_then(|t| t.get("poetry"))
                .and_then(|p| p.get("dependencies")),
        ));
    }
    if let Some(text) = read_in(root, &["requirements.txt"]) {
        names.extend(
            text.lines()
                .map(str::trim)
                .filter(|l| !l.is_empty() && !l.starts_with('#') && !l.starts_with('-'))
                .map(requirement_name),
        );
    }
    names.into_iter().map(|n| n.to_ascii_lowercase()).collect()
}

/// Dependency names declared in the root manifests (names only; versions are ignored).
fn dependency_names(root: &Path) -> BTreeSet<String> {
    let mut names = BTreeSet::new();
    if let Some(doc) = read_in(root, &["Cargo.toml"]).and_then(|t| toml_doc(&t)) {
        for section in ["dependencies", "dev-dependencies", "build-dependencies"] {
            names.extend(table_keys(doc.get(section)));
        }
    }
    if let Some(v) =
        read_in(root, &["package.json"]).and_then(|t| serde_json::from_str::<Value>(&t).ok())
    {
        for section in ["dependencies", "devDependencies"] {
            if let Some(map) = v[section].as_object() {
                names.extend(map.keys().cloned());
            }
        }
    }
    names.extend(python_dependency_names(root));
    if let Some(text) = read_in(root, &["go.mod"]) {
        for line in text.lines() {
            let line = line.trim().trim_start_matches("require").trim();
            if let Some(module) = line.split_whitespace().next() {
                if module.contains('/') {
                    names.insert(module.to_string());
                }
            }
        }
    }
    names.into_iter().map(|n| n.to_ascii_lowercase()).collect()
}

/// The datasource provider named in `prisma/schema.prisma`, as a database kind.
fn prisma_database(root: &Path) -> Option<String> {
    let text = read_in(root, &["prisma", "schema.prisma"])?;
    let provider = text.lines().find_map(|l| {
        let l = l.trim();
        l.strip_prefix("provider")
            .map(str::trim)
            .and_then(|r| r.strip_prefix('='))
            .map(|r| r.trim().trim_matches('"').to_string())
    })?;
    match provider.as_str() {
        "postgresql" | "postgres" | "cockroachdb" => Some("postgres".into()),
        "mysql" => Some("mysql".into()),
        "sqlite" => Some("sqlite".into()),
        "mongodb" => Some("mongodb".into()),
        "sqlserver" => Some("sqlserver".into()),
        _ => None,
    }
}

/// MCP server names from the project's `.mcp.json` (`mcpServers`). The `.codex` folder is not
/// read: moai no longer manages it. Values are never read further.
fn mcp_servers(root: &Path) -> BTreeSet<String> {
    let mut names = BTreeSet::new();
    if let Some(v) =
        read_in(root, &[".mcp.json"]).and_then(|t| serde_json::from_str::<Value>(&t).ok())
    {
        if let Some(map) = v["mcpServers"].as_object() {
            names.extend(map.keys().cloned());
        }
    }
    names
}

fn capped(set: BTreeSet<String>) -> Vec<String> {
    set.into_iter()
        .filter(|n| recordable(n))
        .take(MAX_ITEMS)
        .collect()
}

/// [`capped`] for `detected.manifests`, whose entries are relative paths, not bare names.
fn capped_paths(set: BTreeSet<String>) -> Vec<String> {
    set.into_iter()
        .filter(|p| recordable_relpath(p))
        .take(MAX_ITEMS)
        .collect()
}

/// Detects the facts of the project at `root`. Reads only; follows no symbolic link.
/// The languages of the manifest files at `root` alone, sorted: the cheap part of [`detect`]
/// (no walk of the code files), for readers that must stay fast, such as the tool-preflight
/// hook (`crate::tdd::TestPaths`).
pub fn manifest_languages(root: &Path) -> Vec<String> {
    let (catalog, _) = Catalog::for_project(root);
    let languages: BTreeSet<String> = catalog
        .rows
        .iter()
        .filter(|row| {
            row.files
                .iter()
                .any(|file| is_file_in(root, &[file.as_str()]))
        })
        .map(|row| row.language.clone())
        .collect();
    languages.into_iter().collect()
}

/// What one reading of the project showed: the facts that are recorded and the facts about the
/// reading itself, which are reported beside them and never recorded.
pub(crate) struct Detection {
    pub detected: Detected,
    pub catalog: Catalog,
    pub extension: ExtensionReport,
    /// False when the manifest walk stopped at its entry budget.
    pub walk_complete: bool,
    /// Manifests found that `detected.manifests` leaves out because of [`MAX_ITEMS`].
    pub over_cap: usize,
}

pub fn detect(root: &Path) -> Result<Detected> {
    detect_full(root).map(|d| d.detected)
}

pub(crate) fn detect_full(root: &Path) -> Result<Detection> {
    let meta = fs::symlink_metadata(root).map_err(|_| "project directory unavailable")?;
    if !meta.is_dir() {
        return Err("project must be a directory (not a symbolic link)".into());
    }
    let (catalog, extension) = Catalog::for_project(root);
    let mut manifests = BTreeSet::new();
    // A manifest at the root is found by its name whatever the walk's budget or ignore lists say.
    for row in &catalog.rows {
        for file in &row.files {
            if is_file_in(root, &[file.as_str()]) {
                manifests.insert(file.clone());
            }
        }
    }
    let walk = walk_manifests(root, &catalog);
    let code = code_languages(root);
    let existing = !manifests.is_empty() || !walk.found.is_empty() || !code.is_empty();
    manifests.extend(walk.found);
    let mut languages: BTreeSet<String> = manifests
        .iter()
        .filter_map(|path| catalog.row_for(path.rsplit('/').next().unwrap_or(path)))
        .map(|row| row.language.clone())
        .collect();
    languages.extend(code);
    let recordable_manifests = manifests.iter().filter(|p| recordable_relpath(p)).count();
    let lsp: BTreeSet<String> = languages
        .iter()
        .filter_map(|l| {
            LSP_PLUGINS
                .iter()
                .find(|(lang, ..)| lang == l)
                .map(|(_, server, ..)| (*server).to_string())
        })
        .collect();
    let deps = dependency_names(root);
    let mut databases: BTreeSet<String> = DB_DEPENDENCIES
        .iter()
        .filter(|(dep, _)| deps.contains(*dep))
        .map(|(_, kind)| (*kind).to_string())
        .collect();
    databases.extend(prisma_database(root));
    let detected = Detected {
        project_kind: if existing { "existing" } else { "new" }.into(),
        manifests: capped_paths(manifests),
        languages: capped(languages),
        lsp: capped(lsp),
        mcp_servers: capped(mcp_servers(root)),
        databases: capped(databases),
    };
    Ok(Detection {
        detected,
        catalog,
        extension,
        walk_complete: walk.complete,
        over_cap: recordable_manifests.saturating_sub(MAX_ITEMS),
    })
}

/// Accepts a `detected` value as `detect` produces it.
pub fn validate(d: &Detected) -> Result<()> {
    let manifests_ok = d.manifests.len() <= MAX_ITEMS
        && d.manifests.iter().all(|p| recordable_relpath(p))
        && d.manifests.windows(2).all(|w| w[0] < w[1]);
    let lists = [&d.languages, &d.lsp, &d.mcp_servers, &d.databases];
    let lists_ok = lists.iter().all(|l| {
        l.len() <= MAX_ITEMS && l.iter().all(|n| recordable(n)) && l.windows(2).all(|w| w[0] < w[1])
    });
    if !["new", "existing"].contains(&d.project_kind.as_str()) || !manifests_ok || !lists_ok {
        return Err("invalid detected project facts; no values were logged".into());
    }
    Ok(())
}

/// Build and test command candidates and CI files as facts (#2639), reported beside
/// `detected`, never inside it, so the profile is unchanged. Each command is a fixed template
/// chosen by what a root manifest declares — a script body or a value from the project's files
/// never enters the output — and nothing is run: the interview confirms them with the user.
fn commands_fact(root: &Path) -> Value {
    let mut build = BTreeSet::new();
    let mut test = BTreeSet::new();
    let mut lint = BTreeSet::new();
    let mut format = BTreeSet::new();
    if is_file_in(root, &["Cargo.toml"]) {
        build.insert("cargo build".to_string());
        test.insert("cargo test".to_string());
        lint.insert("cargo clippy --all-targets -- -D warnings".to_string());
        format.insert("cargo fmt --check".to_string());
    }
    if let Some(v) =
        read_in(root, &["package.json"]).and_then(|t| serde_json::from_str::<Value>(&t).ok())
    {
        let pm = [
            ("pnpm-lock.yaml", "pnpm"),
            ("yarn.lock", "yarn"),
            ("bun.lockb", "bun"),
            ("bun.lock", "bun"),
        ]
        .iter()
        .find(|(lock, _)| is_file_in(root, &[lock]))
        .map_or("npm", |(_, pm)| pm);
        if let Some(scripts) = v["scripts"].as_object() {
            if scripts.contains_key("build") {
                build.insert(format!("{pm} run build"));
            }
            if scripts.contains_key("test") {
                // `bun test` runs Bun's own test runner, not the script (Codex review on #2645).
                let run_test = if pm == "bun" {
                    "bun run test"
                } else {
                    &format!("{pm} test")
                };
                test.insert(run_test.to_string());
            }
            // Only a script that checks: `format` usually rewrites files, so it is not offered.
            if scripts.contains_key("lint") {
                lint.insert(format!("{pm} run lint"));
            }
            if scripts.contains_key("format:check") {
                format.insert(format!("{pm} run format:check"));
            }
        }
    }
    if is_file_in(root, &["go.mod"]) {
        build.insert("go build ./...".to_string());
        test.insert("go test ./...".to_string());
        lint.insert("go vet ./...".to_string());
        format.insert("test -z \"$(gofmt -l .)\"".to_string());
    }
    if dependency_names(root).contains("pytest") || python_test_names(root).contains("pytest") {
        test.insert("pytest".to_string());
    }
    // Lint and format tools count only when a Python file declares them: a Rust or Node
    // dependency that is also called `ruff` offers nothing.
    let python_names: BTreeSet<String> = python_dependency_names(root)
        .into_iter()
        .chain(python_test_names(root))
        .collect();
    if python_names.contains("ruff") {
        lint.insert("ruff check .".to_string());
        format.insert("ruff format --check .".to_string());
    }
    if python_names.contains("black") {
        format.insert("black --check .".to_string());
    }
    if let Some(text) = read_in(root, &["Makefile"]) {
        // A Makefile target is offered only under a name that checks, never one that rewrites.
        for (target, set) in [
            ("build", &mut build),
            ("test", &mut test),
            ("lint", &mut lint),
            ("format-check", &mut format),
        ] {
            if text.lines().any(|l| l.starts_with(&format!("{target}:"))) {
                set.insert(format!("make {target}"));
            }
        }
    }
    if is_file_in(root, &["pom.xml"]) {
        build.insert("mvn package".to_string());
        test.insert("mvn test".to_string());
    }
    if is_file_in(root, &["build.gradle"]) || is_file_in(root, &["build.gradle.kts"]) {
        let gradle = if is_file_in(root, &["gradlew"]) {
            "./gradlew"
        } else {
            "gradle"
        };
        build.insert(format!("{gradle} build"));
        test.insert(format!("{gradle} test"));
    }
    json!({
        "build": build.into_iter().collect::<Vec<_>>(),
        "test": test.into_iter().collect::<Vec<_>>(),
        "lint": lint.into_iter().collect::<Vec<_>>(),
        "format": format.into_iter().collect::<Vec<_>>(),
        "ci_files": ci_files(root),
    })
}

/// Dependency names in the `pyproject.toml` sections that commonly hold test tools and that
/// `dependency_names` does not read: `[dependency-groups]`, `[project.optional-dependencies]`,
/// Poetry's `[tool.poetry.group.<name>.dependencies]` and `[tool.poetry.dev-dependencies]`
/// (Codex review on #2645). Names only.
fn python_test_names(root: &Path) -> BTreeSet<String> {
    let mut names = BTreeSet::new();
    let Some(doc) = read_in(root, &["pyproject.toml"]).and_then(|t| toml_doc(&t)) else {
        return names;
    };
    let mut from_arrays = |item: Option<&toml_edit::Item>| {
        if let Some(table) = item.and_then(toml_edit::Item::as_table_like) {
            for (_, list) in table.iter() {
                if let Some(array) = list.as_array() {
                    names.extend(
                        array
                            .iter()
                            .filter_map(|v| v.as_str())
                            .map(requirement_name),
                    );
                }
            }
        }
    };
    from_arrays(doc.get("dependency-groups"));
    from_arrays(
        doc.get("project")
            .and_then(|p| p.get("optional-dependencies")),
    );
    let poetry = doc.get("tool").and_then(|t| t.get("poetry"));
    names.extend(table_keys(poetry.and_then(|p| p.get("dev-dependencies"))));
    if let Some(groups) = poetry
        .and_then(|p| p.get("group"))
        .and_then(toml_edit::Item::as_table_like)
    {
        for (_, group) in groups.iter() {
            names.extend(table_keys(group.get("dependencies")));
        }
    }
    names.into_iter().map(|n| n.to_ascii_lowercase()).collect()
}

/// Workflow file names under `.github/workflows/`, without following a symbolic link.
fn ci_files(root: &Path) -> Vec<String> {
    let real_dir = |p: &Path| fs::symlink_metadata(p).is_ok_and(|m| m.is_dir());
    let dir = root.join(".github").join("workflows");
    if !real_dir(&root.join(".github")) || !real_dir(&dir) {
        return Vec::new();
    }
    let names: BTreeSet<String> = fs::read_dir(&dir)
        .map(|entries| {
            entries
                .flatten()
                .filter(|e| e.file_type().is_ok_and(|t| t.is_file()))
                .filter_map(|e| e.file_name().into_string().ok())
                .filter(|n| n.ends_with(".yml") || n.ends_with(".yaml"))
                .collect()
        })
        .unwrap_or_default();
    capped(names)
}

/// Whether `name` is an executable file in one of the folders of `path` (the value of the PATH
/// variable). It only looks at file metadata: nothing is run, written or fetched, and an empty
/// entry (which the shell reads as the current folder) is skipped.
///
/// A relative `PATH` entry is honoured as a shell does: it is read against the current folder.
fn executable_on_path(name: &str, path: Option<&std::ffi::OsStr>) -> bool {
    let Some(path) = path else {
        return false;
    };
    std::env::split_paths(path)
        .filter(|dir| !dir.as_os_str().is_empty())
        .any(|dir| {
            let candidates: &[&str] = if cfg!(windows) {
                &[".exe", ".cmd", ".bat"]
            } else {
                &[""]
            };
            candidates.iter().any(|ext| {
                fs::metadata(dir.join(format!("{name}{ext}"))).is_ok_and(|m| {
                    #[cfg(unix)]
                    let runnable = {
                        use std::os::unix::fs::PermissionsExt;
                        m.permissions().mode() & 0o111 != 0
                    };
                    #[cfg(not(unix))]
                    let runnable = true;
                    m.is_file() && runnable
                })
            })
        })
}

/// The official Claude Code LSP plugin each detected language calls for, one row per plugin
/// (SPEC-20261123-lsp-official-plugins), reported beside `detected`, never inside it: whether a
/// server is installed is a fact about this machine, and `project.yaml` is tracked. `installed`
/// says whether the plugin's server executable is on PATH; whether the plugin itself is enabled
/// is not read here (that is HOME state): the PROJECT route asks `claude plugin list`.
fn lsp_plugins_fact(languages: &[String], path: Option<&std::ffi::OsStr>) -> Value {
    let mut plugins: BTreeMap<&str, (BTreeSet<&str>, &str)> = BTreeMap::new();
    for lang in languages {
        if let Some((l, _, plugin, executable)) =
            LSP_PLUGINS.iter().find(|(l, ..)| *l == lang.as_str())
        {
            plugins
                .entry(plugin)
                .or_insert_with(|| (BTreeSet::new(), executable))
                .0
                .insert(l);
        }
    }
    let rows: Vec<Value> = plugins
        .into_iter()
        .map(|(plugin, (langs, executable))| {
            json!({
                "plugin": plugin,
                "marketplace": LSP_MARKETPLACE,
                "languages": langs.into_iter().collect::<Vec<_>>(),
                "executable": executable,
                "installed": executable_on_path(executable, path),
            })
        })
        .collect();
    Value::Array(rows)
}

/// Go-generation project state as a fact (#2561), read only and without following links:
/// what `.moai` is (`absent`, `folder`, `link` or `other`) and whether a real folder holds a
/// regular `manifest.json`. Reported beside `detected`, never inside it, so the profile is
/// unchanged. A `.moai/` that holds only the project's tracked SPECs is no state and reads
/// `absent`, as in `state_root::classify` (SDD B1).
fn go_state_fact(root: &Path) -> Value {
    let moai = root.join(".moai");
    let moai_dir = match fs::symlink_metadata(&moai) {
        Err(_) => "absent",
        Ok(m) if m.file_type().is_symlink() => "link",
        Ok(_) if crate::state_root::specs_only(&moai) => "absent",
        Ok(m) if m.is_dir() => "folder",
        Ok(_) => "other",
    };
    let go_manifest = moai_dir == "folder"
        && fs::symlink_metadata(moai.join("manifest.json"))
            .map(|m| m.file_type().is_file())
            .unwrap_or(false);
    json!({"moai_dir": moai_dir, "go_manifest": go_manifest})
}

/// `project-detect`: the facts, and — when the project has a profile — that profile with
/// `detected` set, ready for `project-plan`. Writes nothing.
pub fn detect_cli(root: &Path) -> Result<Value> {
    let Detection {
        detected,
        catalog,
        extension,
        walk_complete,
        over_cap,
    } = detect_full(root)?;
    // C-61 AC-2: the tools the manifest names imply, beside `detected` for the tools question.
    let tools = tools_fact(&detected.manifests, &catalog);
    let shown = root.display();
    let index = format!(
        "suggest only, run after the user agrees: moai index-build {shown} with stdin {{\"protocol_version\":1}}"
    );
    let go_state = go_state_fact(root);
    let state_root = crate::state_root::classify(root);
    // #2696: after the move `.moai/` is a folder of this generation too, so the step follows
    // the classification, not the folder kind.
    let go_step = match state_root.generation {
        "go" => Some(format!(
            "Go-generation .moai/ found: ask the user first; only after the user agrees run the \
             read-only preview `moai go-state-plan {shown}`, show it, and apply only after a \
             second agreement (see the PROJECT workflow)"
        )),
        // #2696 C4a: the state is still in the earlier `.moai-cli/`; the same preview moves it.
        "none" if state_root.legacy_root => Some(format!(
            "the project state is still in the earlier .moai-cli/: ask the user first; only \
             after the user agrees run the read-only preview `moai go-state-plan {shown}`, \
             show it, and apply only after a second agreement (see the PROJECT workflow)"
        )),
        "interrupted" => Some(format!(
            "a Go-state move stopped midway: show the read-only `moai go-state-revert-plan \
             {shown} --backup <UTC>` (the stamp is the folder name under .moai/backups/) and move \
             nothing else until the user decides"
        )),
        _ => None,
    };
    let (profile, first) = match crate::read_profile(root) {
        Ok((mut profile, _)) => {
            profile.schema_version = 2;
            profile.detected = Some(detected.clone());
            (
                json!(profile),
                format!("pass result.profile to `moai project-plan {shown}` on stdin, review the plan, then `moai project-apply`"),
            )
        }
        Err(e) if e == "project profile missing" => (
            Value::Null,
            format!(
                "no {}/project.yaml yet: create the profile first (a first launch or SETUP), then run project-detect again",
                // Where the profile will be created (#2696 C2b), not where it is read now.
                crate::state_root::create_dir(root)
            ),
        ),
        Err(e) => return Err(e),
    };
    let mut next_actions = vec![first, index];
    next_actions.extend(go_step);
    if !walk_complete || over_cap > 0 {
        // SPEC-20261124-project-detect-monorepo REQ-8: a partial list must not silently replace
        // a longer recorded one.
        next_actions.push(
            "the manifest list is partial (the walk stopped at its entry budget or the list was \
             capped): never record it over an earlier detected.manifests list without showing \
             the user which entries would drop"
                .to_string(),
        );
    }
    Ok(json!({
        "detected": detected,
        "manifest_walk": {"complete": walk_complete, "over_cap": over_cap},
        "manifest_catalog": extension.to_json(),
        "tools": tools,
        "profile": profile,
        "profile_written": false,
        "go_state": go_state,
        // #2725 S2a: whose `.moai/` this is, before anything may write or move it.
        "state_root": state_root.to_json(),
        "commands": commands_fact(root),
        // SPEC-20261123-lsp-official-plugins: the official LSP plugin of each language and
        // whether its server is on PATH, a fact beside `detected` like the commands.
        "lsp_plugins": lsp_plugins_fact(
            &detected.languages,
            std::env::var_os("PATH").as_deref(),
        ),
        // C-69: the versioning scheme the tag names and root manifests already show, a fact
        // beside `detected` like the commands; the interview offers keeping it.
        "versioning": crate::versioning::fact(root),
        // SPEC-20261122-project-scale: small, large single tree or monorepo, a fact beside
        // `detected` that the user's pick can override; never recorded inside `detected`.
        "scale": crate::project_scale_detect::read(root),
        "next_actions": next_actions,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    /// The ignore chain of a folder from its lists, shallowest first.
    fn chain(levels: Vec<(usize, Vec<IgnoreLine>)>) -> Rc<IgnoreNode> {
        let mut parent = None;
        for (base, lines) in levels {
            parent = Some(Rc::new(IgnoreNode {
                base,
                lines,
                parent,
            }));
        }
        parent.unwrap()
    }

    #[test]
    fn tables_name_only_recordable_values() {
        for (lang, server, plugin, executable) in LSP_PLUGINS {
            for name in [lang, server, plugin, executable] {
                assert!(recordable(name), "{name}");
            }
            assert!(EXTENSIONS.iter().any(|(_, l)| l == lang), "{lang}");
        }
        for (_, k) in DB_DEPENDENCIES {
            assert!(recordable(k));
        }
    }

    /// SPEC-20261124-project-detect-monorepo REQ-4: the shipped catalog parses, every value in it
    /// can be recorded, and it holds the twelve manifest names the earlier tables held with the
    /// same language and tools.
    #[test]
    fn the_shipped_catalog_holds_the_earlier_rows_with_recordable_values() {
        let file: CatalogFile = serde_json::from_str(CATALOG_JSON).unwrap();
        catalog_file_valid(&file).unwrap();
        let catalog = shipped();
        assert_eq!(catalog.rows.len(), file.manifests.len());
        let earlier: &[(&str, &str, &[&str])] = &[
            ("Cargo.toml", "rust", &["cargo"]),
            ("package.json", "javascript", &["node", "npm"]),
            ("tsconfig.json", "typescript", &["typescript"]),
            ("pyproject.toml", "python", &["python"]),
            ("setup.py", "python", &["python"]),
            ("requirements.txt", "python", &["python"]),
            ("go.mod", "go", &["go"]),
            ("pom.xml", "java", &["maven"]),
            ("build.gradle", "java", &["gradle"]),
            ("build.gradle.kts", "kotlin", &["gradle"]),
            ("Gemfile", "ruby", &["ruby"]),
            ("composer.json", "php", &["php"]),
        ];
        for (name, language, tools) in earlier {
            let row = catalog.row_for(name).unwrap_or_else(|| panic!("{name}"));
            assert_eq!(&row.language, language, "{name}");
            assert_eq!(&row.tools, tools, "{name}");
        }
        for dir in [
            ".venv",
            "venv",
            "target",
            "node_modules",
            "dist",
            "build",
            "vendor",
        ] {
            assert!(catalog.skips(dir), "{dir}");
        }
    }

    /// REQ-5: an extension row never replaces a shipped one, a suffix row matches a name that
    /// is longer than the suffix, and an exact name wins over a suffix.
    #[test]
    fn extension_rows_are_added_and_never_replace_a_shipped_row() {
        let mut catalog = shipped().clone();
        let extension: CatalogFile = serde_json::from_str(
            r#"{"schema": 1, "manifests": [
                {"ecosystem": "mine", "language": "elixir", "files": ["package.json", "mix.exs"], "tools": ["mix"]},
                {"ecosystem": "nuget", "language": "csharp", "suffixes": [".csproj"]},
                {"ecosystem": "again", "language": "csharp", "files": ["go.mod"]}
            ], "skip_dirs": ["generated-out", "target"]}"#,
        )
        .unwrap();
        catalog_file_valid(&extension).unwrap();
        catalog.add_extension(extension);
        assert_eq!(
            catalog.row_for("package.json").unwrap().language,
            "javascript"
        );
        assert_eq!(catalog.row_for("go.mod").unwrap().language, "go");
        assert_eq!(catalog.row_for("mix.exs").unwrap().language, "elixir");
        assert_eq!(catalog.row_for("App.csproj").unwrap().language, "csharp");
        assert!(catalog.row_for(".csproj").is_none());
        assert!(catalog.skips("generated-out"));
        assert_eq!(
            catalog.skip_dirs.iter().filter(|d| *d == "target").count(),
            1
        );
    }

    /// REQ-6: a catalog file with a value that cannot be recorded, an unknown key, a wrong
    /// schema or a row without a name is refused whole.
    #[test]
    fn a_catalog_file_with_an_unrecordable_value_is_refused() {
        for bad in [
            r#"{"schema": 2}"#,
            r#"{"schema": 1, "manifests": [{"ecosystem": "a b", "language": "x", "files": ["f"]}]}"#,
            r#"{"schema": 1, "manifests": [{"ecosystem": "a", "language": "x", "files": ["../f"]}]}"#,
            r#"{"schema": 1, "manifests": [{"ecosystem": "a", "language": "x"}]}"#,
            r#"{"schema": 1, "manifests": [{"ecosystem": "a", "language": "x", "suffixes": ["csproj"]}]}"#,
            r#"{"schema": 1, "skip_dirs": ["a/b"]}"#,
        ] {
            let parsed = serde_json::from_str::<CatalogFile>(bad).unwrap();
            assert!(catalog_file_valid(&parsed).is_err(), "{bad}");
        }
        assert!(serde_json::from_str::<CatalogFile>(r#"{"schema": 1, "extra": 1}"#).is_err());
    }

    /// C-61: a manifest path is a relative path of recordable components; nothing else passes.
    #[test]
    fn recordable_relpaths_accept_only_relative_component_paths() {
        for ok in [
            "package.json",
            "plugins/moai-atp/package.json",
            "a/b/c/package.json",
        ] {
            assert!(recordable_relpath(ok), "{ok}");
        }
        for bad in [
            "",
            "/package.json",
            "a//b/package.json",
            "a/../package.json",
            "./package.json",
            "a/b/",
            "a\\b",
            "a b/package.json",
            &format!("{}.json", "x".repeat(160)),
        ] {
            assert!(!recordable_relpath(bad), "{bad}");
        }
    }

    /// C-61: the tools fact maps each tool to the manifests that imply it, sorted and capped.
    #[test]
    fn tools_fact_maps_and_sorts_manifest_paths() {
        let manifests = vec![
            "Cargo.toml".to_string(),
            "plugins/x/package.json".to_string(),
            "plugins/x/server/package.json".to_string(),
        ];
        let fact = tools_fact(&manifests, shipped());
        let keys: Vec<&String> = fact.keys().collect();
        assert_eq!(keys, ["cargo", "node", "npm"]);
        assert_eq!(
            fact["node"],
            vec!["plugins/x/package.json", "plugins/x/server/package.json"]
        );
        // A name with no tools row is simply absent.
        let gemfile = tools_fact(&["Gemfile".to_string()], shipped());
        assert_eq!(gemfile.keys().collect::<Vec<_>>(), ["ruby"]);
    }

    /// C-61: ignore lines follow the `.gitignore` reading the walk depends on — folder-only
    /// patterns, unanchored names, negation order, character classes, anchored patterns — and
    /// the stack lets a deeper list override a shallower one.
    #[test]
    fn ignore_lines_follow_gitignore_reading() {
        let lines = ignore_lines("gen/\n*.cache\n!x.cache\n");
        // "gen/" excludes a folder named gen at any depth, never a file named gen.
        let gen = &lines[0];
        assert!(ignore_matches_at(gen, "a/gen", false));
        assert!(!ignore_matches_at(gen, "a/gen", true));
        // Within one list the last matching line decides: "!x.cache" re-includes x.cache.
        assert!(!last_matching(&lines, "a/b/y.cache", true).unwrap().negate);
        assert!(last_matching(&lines, "a/b/x.cache", true).unwrap().negate);
        // A file pattern never excludes the folder holding the file.
        assert!(last_matching(&lines, "a/b", false).is_none());
        // A character class and an anchored pattern.
        let lines = ignore_lines("build[0-9]/\n/docs\n");
        assert!(ignore_matches_at(&lines[0], "build7", false));
        assert!(ignore_matches_at(&lines[0], "a/build7", false));
        assert!(!ignore_matches_at(&lines[0], "build7", true));
        assert!(ignore_matches_at(&lines[1], "docs", false));
        assert!(!ignore_matches_at(&lines[1], "a/docs", false));
        // The stack: the deepest list with a matching line decides — a deeper "!old" overrides
        // a shallower "old/" for the folder, and the walk reads the file under it.
        let stack = chain(vec![
            (0, ignore_lines("old/\n")),
            (1, ignore_lines("!old\n")),
        ]);
        assert!(!stack.excludes(&["pkg", "old"], false));
        assert!(!stack.excludes(&["pkg", "old", "package.json"], true));
        // A folder the root list excludes is never entered, whatever a deeper list says about
        // the files inside it.
        let stack = chain(vec![
            (0, ignore_lines("old/\n")),
            (1, ignore_lines("!old/**\n")),
        ]);
        assert!(stack.excludes(&["pkg", "old"], false));
    }

    #[test]
    fn requirement_names_drop_versions_and_extras() {
        assert_eq!(requirement_name("psycopg[binary]>=3.1"), "psycopg");
        assert_eq!(requirement_name("Redis==5.0; python_version>'3'"), "redis");
        assert_eq!(requirement_name("  asyncpg  "), "asyncpg");
    }

    #[test]
    fn validate_refuses_unsorted_unknown_or_path_like_values() {
        let ok = Detected {
            project_kind: "existing".into(),
            manifests: vec!["Cargo.toml".into()],
            ..Default::default()
        };
        assert!(validate(&ok).is_ok());
        // C-61: a manifest may be a relative path; a step out or an absolute path is refused.
        let nested = Detected {
            manifests: vec!["Cargo.toml".into(), "plugins/moai-atp/package.json".into()],
            ..ok.clone()
        };
        assert!(validate(&nested).is_ok());
        for bad in ["../x", "/package.json", "a//b"] {
            let escape = Detected {
                manifests: vec![bad.into()],
                ..ok.clone()
            };
            assert!(validate(&escape).is_err(), "{bad}");
        }
        let bad_kind = Detected {
            project_kind: "maybe".into(),
            ..ok.clone()
        };
        assert!(validate(&bad_kind).is_err());
        let unsorted = Detected {
            languages: vec!["rust".into(), "go".into()],
            ..ok.clone()
        };
        assert!(validate(&unsorted).is_err());
        let path = Detected {
            mcp_servers: vec!["../x".into()],
            ..ok
        };
        assert!(validate(&path).is_err());
    }
}
