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

/// Detects the facts of the project at `root`. Reads only; follows no symbolic link.
pub fn detect(root: &Path) -> Result<Detected> {
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
/// unchanged.
fn go_state_fact(root: &Path) -> Value {
    let moai = root.join(".moai");
    let moai_dir = match fs::symlink_metadata(&moai) {
        Err(_) => "absent",
        Ok(m) if m.file_type().is_symlink() => "link",
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
    let detected = detect(root)?;
    let shown = root.display();
    let index = format!(
        "suggest only, run after the user agrees: moai-cli index-build {shown} with stdin {{\"protocol_version\":1}}"
    );
    let go_state = go_state_fact(root);
    let state_root = crate::state_root::classify(root);
    // #2696: after the move `.moai/` is a folder of this generation too, so the step follows
    // the classification, not the folder kind.
    let go_step = match state_root.generation {
        "go" => Some(format!(
            "Go-generation .moai/ found: ask the user first; only after the user agrees run the \
             read-only preview `moai-cli go-state-plan {shown}`, show it, and apply only after a \
             second agreement (see the PROJECT workflow)"
        )),
        // #2696 C4a: the state is still in the earlier `.moai-cli/`; the same preview moves it.
        "none" if state_root.legacy_root => Some(format!(
            "the project state is still in the earlier .moai-cli/: ask the user first; only \
             after the user agrees run the read-only preview `moai-cli go-state-plan {shown}`, \
             show it, and apply only after a second agreement (see the PROJECT workflow)"
        )),
        "interrupted" => Some(format!(
            "a Go-state move stopped midway: show the read-only `moai-cli go-state-revert-plan \
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
                format!("pass result.profile to `moai-cli project-plan {shown}` on stdin, review the plan, then `moai-cli project-apply`"),
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
    Ok(json!({
        "detected": detected,
        "profile": profile,
        "profile_written": false,
        "go_state": go_state,
        // #2725 S2a: whose `.moai/` this is, before anything may write or move it.
        "state_root": state_root.to_json(),
        "commands": commands_fact(root),
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
