//! Project detection for the `/moai project` route (Issue #2496; user decisions 16 = B and
//! 17 = all three, #1975 comment 6815357071).
//!
//! `project-detect` reads the project folder and writes nothing. It tells a new project from
//! an existing one (a manifest or a code file makes it existing), names the languages and the
//! language servers they call for, the MCP servers the project already configures, and the
//! databases its dependencies name. Only project facts are recorded: no HOME read, no PATH
//! probe, no process start, because `.moai-cli/project.yaml` is tracked. The `versioning`
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
    collections::{BTreeMap, BTreeSet},
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

/// Root manifest file → language.
const MANIFESTS: &[(&str, &str)] = &[
    ("Cargo.toml", "rust"),
    ("package.json", "javascript"),
    ("tsconfig.json", "typescript"),
    ("pyproject.toml", "python"),
    ("setup.py", "python"),
    ("requirements.txt", "python"),
    ("go.mod", "go"),
    ("pom.xml", "java"),
    ("build.gradle", "java"),
    ("build.gradle.kts", "kotlin"),
    ("Gemfile", "ruby"),
    ("composer.json", "php"),
];

/// Manifest file name → the tools its presence implies (C-61 AC-2): names only, so a nested
/// manifest's bytes are never read for this; the interview confirms the pick.
const MANIFEST_TOOLS: &[(&str, &[&str])] = &[
    ("Cargo.toml", &["cargo"]),
    ("package.json", &["node", "npm"]),
    ("tsconfig.json", &["typescript"]),
    ("pyproject.toml", &["python"]),
    ("setup.py", &["python"]),
    ("requirements.txt", &["python"]),
    ("go.mod", &["go"]),
    ("pom.xml", &["maven"]),
    ("build.gradle", &["gradle"]),
    ("build.gradle.kts", &["gradle"]),
    ("Gemfile", &["ruby"]),
    ("composer.json", &["php"]),
];

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
];

/// Language → the language server it calls for.
const LSP: &[(&str, &str)] = &[
    ("rust", "rust-analyzer"),
    ("python", "pyright"),
    ("javascript", "typescript-language-server"),
    ("typescript", "typescript-language-server"),
    ("go", "gopls"),
    ("java", "jdtls"),
    ("kotlin", "kotlin-language-server"),
    ("ruby", "ruby-lsp"),
    ("php", "intelephense"),
    ("c", "clangd"),
    ("cpp", "clangd"),
    ("csharp", "omnisharp"),
    ("swift", "sourcekit-lsp"),
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

/// Folders the code walk never enters: VCS and MoAI state, the host folders MoAI writes hook
/// and agent files into (`.omp/extensions/moai-hooks.ts` is TypeScript, #2936 R9), build
/// output, vendored deps.
const SKIP_DIRS: &[&str] = &[
    ".git",
    PROJECT_STATE_DIR,
    ".claude",
    ".codex",
    ".omp",
    "target",
    "node_modules",
    ".venv",
    "venv",
    "dist",
    "build",
    "vendor",
];
/// Walk bounds: folder depth below the root and directory entries looked at.
pub const MAX_DEPTH: usize = 3;
pub const MAX_ENTRIES: usize = 5000;
/// Largest manifest or configuration file read.
const MAX_FILE: u64 = 256 * 1024;
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
fn read_in(root: &Path, parts: &[&str]) -> Option<String> {
    let (path, meta) = file_in(root, parts)?;
    if meta.len() > MAX_FILE {
        return None;
    }
    fs::read_to_string(path).ok()
}

fn is_file_in(root: &Path, parts: &[&str]) -> bool {
    file_in(root, parts).is_some()
}

/// A name that may be recorded: short, printable, no path separators, no spaces.
fn recordable(name: &str) -> bool {
    !name.is_empty()
        && name.len() <= 80
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
                if depth < MAX_DEPTH && !SKIP_DIRS.contains(&name.as_ref()) {
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

/// The ignore lists the manifest walk has collected: one entry per folder entered, judging
/// paths from its own folder down (`base` is that folder's depth in components). The deepest
/// list with a matching line wins, as git reads nested `.gitignore` files.
#[derive(Clone)]
struct IgnoreStack(Vec<(usize, Rc<Vec<IgnoreLine>>)>);

impl IgnoreStack {
    /// Whether the folder or file at `parts` (components from the project root) is excluded.
    fn excludes(&self, parts: &[&str], is_file: bool) -> bool {
        for (base, lines) in self.0.iter().rev() {
            if *base >= parts.len() {
                continue;
            }
            let rel = parts[*base..].join("/");
            if let Some(l) = last_matching(lines, &rel, is_file) {
                return !l.negate;
            }
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

/// Manifest files below the root (C-61 AC-1): folder depth at most [`MAX_DEPTH`], names only —
/// a nested manifest's bytes are never read — no symbolic link followed (the walk looks at
/// `DirEntry::file_type`, which does not follow one), folders [`SKIP_DIRS`] names, folders whose
/// name starts with a dot (tool, cache and backup state: SPEC-20261121-detect-walk-breadth
/// REQ-2) and folders the ignore lists exclude never entered. An ignore list that cannot be
/// read (a link, over [`MAX_FILE`], not UTF-8) counts as no lines, so that failure can only
/// read more, never less. The walk reads every folder of one depth before any folder one level
/// deeper (REQ-1), so when the `seen` budget ([`MAX_ENTRIES`], like the code walk's) runs out
/// the shallow manifests are already found and `complete` is false.
fn nested_manifests(root: &Path) -> ManifestWalk {
    let mut found = BTreeSet::new();
    let mut seen = 0usize;
    let mut complete = true;
    let mut level: Vec<(PathBuf, Vec<String>, IgnoreStack)> = vec![(
        root.to_path_buf(),
        Vec::new(),
        IgnoreStack(vec![(0, Rc::new(root_ignore_lines(root)))]),
    )];
    'levels: while !level.is_empty() {
        let mut next = Vec::new();
        for (dir, rel, stack) in level {
            let Ok(entries) = fs::read_dir(&dir) else {
                continue;
            };
            let mut entries: Vec<_> = entries.flatten().collect();
            entries.sort_by_key(|e| e.file_name());
            for entry in entries {
                seen += 1;
                if seen > MAX_ENTRIES {
                    complete = false;
                    break 'levels;
                }
                let Ok(kind) = entry.file_type() else {
                    continue;
                };
                let name = entry.file_name();
                let name = name.to_string_lossy();
                let mut parts = rel.clone();
                parts.push(name.to_string());
                let parts_ref: Vec<&str> = parts.iter().map(String::as_str).collect();
                if kind.is_dir() {
                    if SKIP_DIRS.contains(&name.as_ref())
                        || name.starts_with('.')
                        || stack.excludes(&parts_ref, false)
                    {
                        continue;
                    }
                    if rel.len() < MAX_DEPTH {
                        let mut below = stack.clone();
                        below.0.push((
                            parts.len(),
                            Rc::new(
                                read_in(&dir, &[&name, ".gitignore"])
                                    .map(|t| ignore_lines(&t))
                                    .unwrap_or_default(),
                            ),
                        ));
                        next.push((entry.path(), parts, below));
                    }
                } else if kind.is_file()
                    && MANIFESTS.iter().any(|(file, _)| *file == name)
                    && !stack.excludes(&parts_ref, true)
                {
                    found.insert(parts.join("/"));
                }
            }
        }
        level = next;
    }
    ManifestWalk { found, complete }
}

/// What the manifest walk below the root found, and whether it read everything it could.
struct ManifestWalk {
    found: BTreeSet<String>,
    complete: bool,
}

/// How complete `detected.manifests` is (SPEC-20261121-detect-walk-breadth REQ-3), reported
/// beside `detected` and never recorded in it: `complete` is false when the walk stopped at
/// [`MAX_ENTRIES`]; `over_cap` counts the manifests found that the list leaves out because of
/// [`MAX_ITEMS`].
#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize)]
pub struct ManifestWalkFact {
    pub complete: bool,
    pub over_cap: usize,
}

/// The tools the detected manifest names imply (C-61 AC-2), beside `detected` and never inside
/// it: tool name → the sorted manifest paths that imply it, at most [`MAX_ITEMS`] tools. The
/// paths come from `detected.manifests`, so they are already capped and validated.
fn tools_fact(manifests: &[String]) -> BTreeMap<String, Vec<String>> {
    let mut map: BTreeMap<String, BTreeSet<String>> = BTreeMap::new();
    for path in manifests {
        let name = path.rsplit('/').next().unwrap_or(path);
        if let Some((_, tools)) = MANIFEST_TOOLS.iter().find(|(file, _)| *file == name) {
            for tool in *tools {
                map.entry((*tool).to_string())
                    .or_default()
                    .insert(path.clone());
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

/// MCP server names from the project's `.mcp.json` (`mcpServers`) and
/// `.codex/config.toml` (`[mcp_servers.<name>]`). Values are never read further.
fn mcp_servers(root: &Path) -> BTreeSet<String> {
    let mut names = BTreeSet::new();
    if let Some(v) =
        read_in(root, &[".mcp.json"]).and_then(|t| serde_json::from_str::<Value>(&t).ok())
    {
        if let Some(map) = v["mcpServers"].as_object() {
            names.extend(map.keys().cloned());
        }
    }
    if let Some(doc) = read_in(root, &[".codex", "config.toml"]).and_then(|t| toml_doc(&t)) {
        names.extend(table_keys(doc.get("mcp_servers")));
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
    let languages: BTreeSet<String> = MANIFESTS
        .iter()
        .filter(|(file, _)| is_file_in(root, &[file]))
        .map(|(_, lang)| (*lang).to_string())
        .collect();
    languages.into_iter().collect()
}

pub fn detect(root: &Path) -> Result<Detected> {
    detect_with_walk(root).map(|(detected, _)| detected)
}

/// [`detect`] and how complete its manifest list is.
pub fn detect_with_walk(root: &Path) -> Result<(Detected, ManifestWalkFact)> {
    let meta = fs::symlink_metadata(root).map_err(|_| "project directory unavailable")?;
    if !meta.is_dir() {
        return Err("project must be a directory (not a symbolic link)".into());
    }
    let mut manifests = BTreeSet::new();
    let mut languages = BTreeSet::new();
    for (file, lang) in MANIFESTS {
        if is_file_in(root, &[file]) {
            manifests.insert((*file).to_string());
            languages.insert((*lang).to_string());
        }
    }
    let walk = nested_manifests(root);
    let code = code_languages(root);
    let existing = !manifests.is_empty() || !walk.found.is_empty() || !code.is_empty();
    manifests.extend(walk.found);
    languages.extend(code);
    let lsp: BTreeSet<String> = languages
        .iter()
        .filter_map(|l| {
            LSP.iter()
                .find(|(lang, _)| lang == l)
                .map(|(_, s)| (*s).to_string())
        })
        .collect();
    let deps = dependency_names(root);
    let mut databases: BTreeSet<String> = DB_DEPENDENCIES
        .iter()
        .filter(|(dep, _)| deps.contains(*dep))
        .map(|(_, kind)| (*kind).to_string())
        .collect();
    databases.extend(prisma_database(root));
    let recordable_manifests = manifests.iter().filter(|p| recordable_relpath(p)).count();
    let fact = ManifestWalkFact {
        complete: walk.complete,
        over_cap: recordable_manifests.saturating_sub(MAX_ITEMS),
    };
    let detected = Detected {
        project_kind: if existing { "existing" } else { "new" }.into(),
        manifests: capped_paths(manifests),
        languages: capped(languages),
        lsp: capped(lsp),
        mcp_servers: capped(mcp_servers(root)),
        databases: capped(databases),
    };
    Ok((detected, fact))
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
    if is_file_in(root, &["Cargo.toml"]) {
        build.insert("cargo build".to_string());
        test.insert("cargo test".to_string());
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
        }
    }
    if is_file_in(root, &["go.mod"]) {
        build.insert("go build ./...".to_string());
        test.insert("go test ./...".to_string());
    }
    if dependency_names(root).contains("pytest") || python_test_names(root).contains("pytest") {
        test.insert("pytest".to_string());
    }
    if let Some(text) = read_in(root, &["Makefile"]) {
        for target in ["build", "test"] {
            if text.lines().any(|l| l.starts_with(&format!("{target}:"))) {
                let set = if target == "build" {
                    &mut build
                } else {
                    &mut test
                };
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
    let (detected, walk) = detect_with_walk(root)?;
    // C-61 AC-2: the tools the manifest names imply, beside `detected` for the tools question.
    let tools = tools_fact(&detected.manifests);
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
    // SPEC-20261121-detect-walk-breadth REQ-4: a partial list is said so, once.
    if !walk.complete || walk.over_cap > 0 {
        let mut why = Vec::new();
        if !walk.complete {
            why.push(format!("the walk stopped at {MAX_ENTRIES} folder entries"));
        }
        if walk.over_cap > 0 {
            why.push(format!(
                "{} more manifests than the {MAX_ITEMS} the list holds",
                walk.over_cap
            ));
        }
        next_actions.push(format!(
            "manifest list is partial ({}): do not record it over an earlier detected.manifests list without showing the user which entries would drop",
            why.join("; ")
        ));
    }
    Ok(json!({
        "detected": detected,
        "tools": tools,
        // SPEC-20261121-detect-walk-breadth REQ-3: how complete `detected.manifests` is.
        "manifest_walk": walk,
        "profile": profile,
        "profile_written": false,
        "go_state": go_state,
        // #2725 S2a: whose `.moai/` this is, before anything may write or move it.
        "state_root": state_root.to_json(),
        "commands": commands_fact(root),
        // C-69: the versioning scheme the tag names and root manifests already show, a fact
        // beside `detected` like the commands; the interview offers keeping it.
        "versioning": crate::versioning::fact(root),
        "next_actions": next_actions,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn tables_name_only_recordable_values() {
        for (f, l) in MANIFESTS {
            assert!(recordable(f) && recordable(l));
        }
        for (_, s) in LSP {
            assert!(recordable(s));
        }
        for (_, k) in DB_DEPENDENCIES {
            assert!(recordable(k));
        }
    }

    /// C-61: every tools row names a real manifest file, and its values are recordable names.
    #[test]
    fn manifest_tools_rows_name_real_manifests_and_recordable_tools() {
        for (file, tools) in MANIFEST_TOOLS {
            assert!(MANIFESTS.iter().any(|(f, _)| f == file), "{file}");
            assert!(!tools.is_empty(), "{file}");
            for tool in *tools {
                assert!(recordable(tool), "{tool}");
            }
        }
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
        let fact = tools_fact(&manifests);
        let keys: Vec<&String> = fact.keys().collect();
        assert_eq!(keys, ["cargo", "node", "npm"]);
        assert_eq!(
            fact["node"],
            vec!["plugins/x/package.json", "plugins/x/server/package.json"]
        );
        // A name with no tools row is simply absent.
        let gemfile = tools_fact(&["Gemfile".to_string()]);
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
        let stack = IgnoreStack(vec![
            (0, ignore_lines("old/\n").into()),
            (1, ignore_lines("!old\n").into()),
        ]);
        assert!(!stack.excludes(&["pkg", "old"], false));
        assert!(!stack.excludes(&["pkg", "old", "package.json"], true));
        // A folder the root list excludes is never entered, whatever a deeper list says about
        // the files inside it.
        let stack = IgnoreStack(vec![
            (0, ignore_lines("old/\n").into()),
            (1, ignore_lines("!old/**\n").into()),
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
