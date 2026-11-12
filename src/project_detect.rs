//! Project detection for the `/moai project` route (Issue #2496; user decisions 16 = B and
//! 17 = all three, #1975 comment 6815357071).
//!
//! `project-detect` reads the project folder and writes nothing. It tells a new project from
//! an existing one (a manifest or a code file makes it existing), names the languages and the
//! language servers they call for, the MCP servers the project already configures, and the
//! databases its dependencies name. Only project facts are recorded: no HOME read, no PATH
//! probe, no process start, because `.moai-cli/project.yaml` is tracked. MCP configuration is
//! read for server names only and is never created or changed (decision 21 keeps MCP itself
//! on HOLD). The result carries a complete profile for `project-plan`; `index-build` is only
//! suggested, never run.

use crate::PROJECT_STATE_DIR;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{
    collections::BTreeSet,
    fs,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;

/// The `detected` profile value. Every list is sorted and free of duplicates, so a re-run on
/// the same files yields the same bytes.
#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Detected {
    /// `new` (no manifest, no code file) or `existing`.
    pub project_kind: String,
    /// Manifest file names found at the project root.
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

/// Folders the code walk never enters: VCS and MoAI state, build output, vendored deps.
const SKIP_DIRS: &[&str] = &[
    ".git",
    PROJECT_STATE_DIR,
    ".claude",
    ".codex",
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

/// A regular file at `path`, never through a symbolic link, at most `MAX_FILE` bytes.
fn read_small(path: &Path) -> Option<String> {
    let meta = fs::symlink_metadata(path).ok()?;
    if !meta.file_type().is_file() || meta.len() > MAX_FILE {
        return None;
    }
    fs::read_to_string(path).ok()
}

fn is_file(path: &Path) -> bool {
    fs::symlink_metadata(path).is_ok_and(|m| m.file_type().is_file())
}

/// A name that may be recorded: short, printable, no path separators, no spaces.
fn recordable(name: &str) -> bool {
    !name.is_empty()
        && name.len() <= 80
        && name
            .bytes()
            .all(|b| b.is_ascii_alphanumeric() || matches!(b, b'-' | b'_' | b'.' | b'@'))
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
    if let Some(doc) = read_small(&root.join("Cargo.toml")).and_then(|t| toml_doc(&t)) {
        for section in ["dependencies", "dev-dependencies", "build-dependencies"] {
            names.extend(table_keys(doc.get(section)));
        }
    }
    if let Some(v) =
        read_small(&root.join("package.json")).and_then(|t| serde_json::from_str::<Value>(&t).ok())
    {
        for section in ["dependencies", "devDependencies"] {
            if let Some(map) = v[section].as_object() {
                names.extend(map.keys().cloned());
            }
        }
    }
    if let Some(doc) = read_small(&root.join("pyproject.toml")).and_then(|t| toml_doc(&t)) {
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
    if let Some(text) = read_small(&root.join("requirements.txt")) {
        names.extend(
            text.lines()
                .map(str::trim)
                .filter(|l| !l.is_empty() && !l.starts_with('#') && !l.starts_with('-'))
                .map(requirement_name),
        );
    }
    if let Some(text) = read_small(&root.join("go.mod")) {
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
    let text = read_small(&root.join("prisma").join("schema.prisma"))?;
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
        read_small(&root.join(".mcp.json")).and_then(|t| serde_json::from_str::<Value>(&t).ok())
    {
        if let Some(map) = v["mcpServers"].as_object() {
            names.extend(map.keys().cloned());
        }
    }
    let codex = root.join(".codex");
    let codex_dir = fs::symlink_metadata(&codex).is_ok_and(|m| m.file_type().is_dir());
    if codex_dir {
        if let Some(doc) = read_small(&codex.join("config.toml")).and_then(|t| toml_doc(&t)) {
            names.extend(table_keys(doc.get("mcp_servers")));
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

/// Detects the facts of the project at `root`. Reads only; follows no symbolic link.
pub fn detect(root: &Path) -> Result<Detected> {
    let meta = fs::symlink_metadata(root).map_err(|_| "project directory unavailable")?;
    if !meta.is_dir() {
        return Err("project must be a directory (not a symbolic link)".into());
    }
    let mut manifests = BTreeSet::new();
    let mut languages = BTreeSet::new();
    for (file, lang) in MANIFESTS {
        if is_file(&root.join(file)) {
            manifests.insert((*file).to_string());
            languages.insert((*lang).to_string());
        }
    }
    let code = code_languages(root);
    let existing = !manifests.is_empty() || !code.is_empty();
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
    Ok(Detected {
        project_kind: if existing { "existing" } else { "new" }.into(),
        manifests: capped(manifests),
        languages: capped(languages),
        lsp: capped(lsp),
        mcp_servers: capped(mcp_servers(root)),
        databases: capped(databases),
    })
}

/// Accepts a `detected` value as `detect` produces it.
pub fn validate(d: &Detected) -> Result<()> {
    let lists = [
        &d.manifests,
        &d.languages,
        &d.lsp,
        &d.mcp_servers,
        &d.databases,
    ];
    let lists_ok = lists.iter().all(|l| {
        l.len() <= MAX_ITEMS && l.iter().all(|n| recordable(n)) && l.windows(2).all(|w| w[0] < w[1])
    });
    if !["new", "existing"].contains(&d.project_kind.as_str()) || !lists_ok {
        return Err("invalid detected project facts; no values were logged".into());
    }
    Ok(())
}

/// `project-detect`: the facts, and — when the project has a profile — that profile with
/// `detected` set, ready for `project-plan`. Writes nothing.
pub fn detect_cli(root: &Path) -> Result<Value> {
    let detected = detect(root)?;
    let shown = root.display();
    let index = format!(
        "suggest only, run after the user agrees: moai-cli index-build {shown} with stdin {{\"protocol_version\":1}}"
    );
    match crate::read_profile(root) {
        Ok((mut profile, _)) => {
            profile.schema_version = 2;
            profile.detected = Some(detected.clone());
            Ok(json!({
                "detected": detected,
                "profile": profile,
                "profile_written": false,
                "next_actions": [
                    format!("pass result.profile to `moai-cli project-plan {shown}` on stdin, review the plan, then `moai-cli project-apply`"),
                    index,
                ]
            }))
        }
        Err(e) if e == "project profile missing" => Ok(json!({
            "detected": detected,
            "profile": null,
            "profile_written": false,
            "next_actions": [
                "no .moai-cli/project.yaml yet: create the profile first (a first launch or SETUP), then run project-detect again",
                index,
            ]
        })),
        Err(e) => Err(e),
    }
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
