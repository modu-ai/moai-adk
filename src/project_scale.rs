//! The large-codebase planner (SPEC-20261122-project-scale REQ-6, REQ-7, REQ-10, REQ-11;
//! SPEC-20261122-project-scale-settings REQ-1 to REQ-7): from the scale the user picked and the
//! folders they picked in the interview, the plan lists every Claude Code file and settings key
//! it would create or change, each with the documentation line it comes from.
//!
//! The plan is read-only: it reads the project and writes nothing, and it starts no program.
//! A file or key the project already holds with another value is a conflict and is kept. A
//! picked value that looks like a credential, a path that leaves the repository and a pick that
//! is not a detected package, folder or language refuse the whole plan and name the pick.
//! Only Claude Code's own keys are planned; the settings writer checks them again on apply.

use crate::apply::{read_opt, root_of, target};
use crate::digest;
use crate::spec::{Canonical, SortedKeys};
use serde::{Deserialize, Serialize};
use serde_json::{json, Map, Value};
use std::collections::BTreeSet;
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// The documentation lines every row cites, with the date they were read.
const DOCS: &str = include_str!("project_scale_docs.json");

/// Bounds of the picks: entries per list and characters per entry.
pub const PICKS_MAX: usize = 256;
pub const PICK_CHARS_MAX: usize = 256;
/// The documentation's target for an instruction file.
pub const FILE_LINES_MAX: usize = 200;
/// Facts listed in a package `CLAUDE.md` before the rest is counted.
const FACTS_MAX: usize = 20;
/// The marketplace of the code intelligence plugins.
const MARKETPLACE: &str = "claude-plugins-official";
/// Dependency folders a worktree links from the root instead of copying.
const DEPENDENCY_FOLDERS: [&str; 4] = ["node_modules", ".venv", "venv", "vendor"];

/// Detected language, plugin and the language server binary the plugin needs
/// (https://code.claude.com/docs/en/plugins/code-intelligence, read 2026-11-22).
const PLUGINS: [(&str, &str, &str); 14] = [
    ("c", "clangd-lsp", "clangd"),
    ("cpp", "clangd-lsp", "clangd"),
    ("csharp", "csharp-lsp", "csharp-ls"),
    ("go", "gopls-lsp", "gopls"),
    ("java", "jdtls-lsp", "jdtls"),
    ("kotlin", "kotlin-lsp", "kotlin-lsp"),
    ("lua", "lua-lsp", "lua-language-server"),
    ("php", "php-lsp", "intelephense"),
    ("python", "pyright-lsp", "pyright-langserver"),
    ("ruby", "ruby-lsp", "ruby-lsp"),
    ("rust", "rust-analyzer-lsp", "rust-analyzer"),
    ("swift", "swift-lsp", "sourcekit-lsp"),
    ("typescript", "typescript-lsp", "typescript-language-server"),
    ("javascript", "typescript-lsp", "typescript-language-server"),
];

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct ScaleRequest {
    pub protocol_version: u32,
    /// The scale the user picked: `small`, `large_tree` or `monorepo`.
    pub kind: String,
    /// Packages that get a `CLAUDE.md` with their detected facts (monorepo).
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub package_files: Vec<String>,
    /// Packages the user never works in: one `claudeMdExcludes` glob each (monorepo).
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub exclude_packages: Vec<String>,
    /// Tracked generated or vendored folders: one `Read` deny rule each.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub deny_read: Vec<String>,
    /// Detected languages whose code intelligence plugin is turned on.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub plugins: Vec<String>,
    /// Folders a worktree checks out besides `.claude`; empty when sparse worktrees were not picked.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub sparse_paths: Vec<String>,
    /// The one ignore line for `.claude/worktrees/`, asked as its own pick.
    #[serde(default, skip_serializing_if = "std::ops::Not::not")]
    pub ignore_worktrees: bool,
    /// Packages sessions start from: a self-contained settings file each (monorepo).
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub start_packages: Vec<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct ScaleRow {
    /// The repository-relative file the row creates or changes.
    pub file: String,
    /// The settings key, for a row that changes one key of a settings file.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub key: Option<String>,
    /// create | merge | unchanged | conflict
    pub action: String,
    /// The planned value: the file's text, the key's value, or the ignore line.
    pub value: Value,
    pub doc_url: String,
    pub doc_quote: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub detail: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ScalePlan {
    pub protocol_version: u32,
    pub request: ScaleRequest,
    /// What `project-detect` reads now; the user may have picked another kind.
    pub detected_kind: String,
    pub rows: Vec<ScaleRow>,
    /// Install commands and the language server each needs, as text: the plan installs nothing.
    pub install: Vec<String>,
    /// Plain facts for the preview, in English; the interview relays them in the user's language.
    pub notes: Vec<String>,
    /// The day the documentation lines were read.
    pub docs_read_on: String,
    pub plan_digest: String,
}

fn docs() -> Value {
    serde_json::from_str(DOCS).expect("the documentation lines parse")
}

/// Why a picked path is not a plain path inside the repository, or `None`.
fn path_refusal(s: &str) -> Option<&'static str> {
    if s.trim().is_empty() || s.chars().count() > PICK_CHARS_MAX {
        return Some("empty or longer than the limit");
    }
    if s.chars().any(char::is_control) {
        return Some("it holds a control character");
    }
    if crate::claude::credential_like(s) {
        return Some("it looks like a credential");
    }
    let b = s.as_bytes();
    if s.starts_with(['/', '\\', '~'])
        || (b.len() >= 2 && b[1] == b':' && b[0].is_ascii_alphabetic())
    {
        return Some("it is not a relative path");
    }
    if s.split(['/', '\\']).any(|p| p == "..") {
        return Some("a `..` leaves the repository");
    }
    None
}

/// A picked path without a trailing `/` or a leading `./`.
fn clean(s: &str) -> String {
    s.trim_start_matches("./").trim_end_matches('/').to_string()
}

/// Checks one list of picked paths and returns them cleaned, sorted and without repeats.
fn picked_paths(key: &str, list: &[String]) -> Result<Vec<String>> {
    if list.len() > PICKS_MAX {
        return Err(format!(
            "{key} refused: more than {PICKS_MAX} entries; nothing is planned"
        ));
    }
    let mut out = BTreeSet::new();
    for s in list {
        if let Some(why) = path_refusal(s) {
            return Err(format!("{key} refused: {why}; nothing is planned"));
        }
        out.insert(clean(s));
    }
    Ok(out.into_iter().collect())
}

/// A real folder (not a link) at `rel` below `root`.
fn real_dir(root: &Path, rel: &str) -> bool {
    let mut path = root.to_path_buf();
    for part in rel.split('/') {
        path.push(part);
        match std::fs::symlink_metadata(&path) {
            Ok(m) if m.is_dir() => {}
            _ => return false,
        }
    }
    true
}

/// The relative path from folder `from` to folder `to`, both repository-relative.
fn relative(from: &str, to: &str) -> String {
    let a: Vec<&str> = from.split('/').collect();
    let b: Vec<&str> = to.split('/').collect();
    let common = a.iter().zip(&b).take_while(|(x, y)| x == y).count();
    let mut parts: Vec<&str> = vec![".."; a.len() - common];
    parts.extend(&b[common..]);
    parts.join("/")
}

/// A settings file as read: `Ok(None)` when it is missing, `Err` with the reason when it is not
/// a JSON object.
type Current = std::result::Result<Option<Map<String, Value>>, String>;

/// The settings object of `rel`.
fn settings_of(root: &Path, rel: &str) -> Result<Current> {
    let Some(bytes) = read_opt(&target(root, rel)?)? else {
        return Ok(Ok(None));
    };
    Ok(match serde_json::from_slice::<Value>(&bytes) {
        Ok(Value::Object(map)) => Ok(Some(map)),
        _ => Err(format!("{rel} is not a JSON object; it is kept")),
    })
}

fn get_dotted<'a>(map: &'a Map<String, Value>, key: &str) -> Option<&'a Value> {
    let mut parts = key.split('.');
    let mut v = map.get(parts.next()?)?;
    for part in parts {
        v = v.get(part)?;
    }
    Some(v)
}

/// The action for one key: lists and plugin objects merge their new entries, any other value
/// that differs is a conflict.
fn key_action(current: &Current, key: &str, planned: &Value) -> (&'static str, Option<String>) {
    let map = match current {
        Err(why) => return ("conflict", Some(why.clone())),
        Ok(None) => return ("create", None),
        Ok(Some(map)) => map,
    };
    let Some(now) = get_dotted(map, key) else {
        return ("merge", None);
    };
    if now == planned {
        return ("unchanged", None);
    }
    let merges = matches!(key, "permissions.deny" | "claudeMdExcludes");
    match (now, planned) {
        (Value::Array(have), Value::Array(want)) if merges => {
            if want.iter().all(|w| have.contains(w)) {
                ("unchanged", None)
            } else {
                (
                    "merge",
                    Some(format!("added to the {} entries already there", have.len())),
                )
            }
        }
        (Value::Object(have), Value::Object(want)) if key == "enabledPlugins" => {
            if want.iter().all(|(k, v)| have.get(k) == Some(v)) {
                ("unchanged", None)
            } else if want.keys().any(|k| have.contains_key(k)) {
                (
                    "conflict",
                    Some("a picked plugin is already set to another value; it is kept".into()),
                )
            } else {
                (
                    "merge",
                    Some(format!("added to the {} plugins already there", have.len())),
                )
            }
        }
        _ => (
            "conflict",
            Some("the key already holds another value; it is kept".into()),
        ),
    }
}

/// The action for a whole file: a file that exists with other text is a conflict.
fn file_action(root: &Path, rel: &str, text: &str) -> Result<(&'static str, Option<String>)> {
    let path = target(root, rel)?;
    if std::fs::symlink_metadata(&path).is_ok_and(|m| m.file_type().is_symlink()) {
        return Ok(("conflict", Some("the path is a link; it is kept".into())));
    }
    Ok(match read_opt(&path)? {
        None => ("create", None),
        Some(bytes) if bytes == text.as_bytes() => ("unchanged", None),
        Some(_) => (
            "conflict",
            Some("the file exists and was not written by this setup; it is kept".into()),
        ),
    })
}

/// The `CLAUDE.md` of a package: a header and the detected facts only.
fn package_text(package: &str, manifests: &BTreeSet<String>) -> String {
    let prefix = format!("{package}/");
    let mut own = Vec::new();
    let mut nested = 0usize;
    for path in manifests {
        let Some(rest) = path.strip_prefix(&prefix) else {
            continue;
        };
        if rest.contains('/') {
            nested += 1;
        } else {
            own.push(rest.to_string());
        }
    }
    let mut text = format!(
        "# {package}\n\nFacts MoAI detected in this package. The conventions are for the people who own this folder to write below.\n\n"
    );
    for name in own.iter().take(FACTS_MAX) {
        text.push_str(&format!("- Manifest: `{name}`\n"));
    }
    if own.len() > FACTS_MAX {
        text.push_str(&format!(
            "- Other manifests here: {}\n",
            own.len() - FACTS_MAX
        ));
    }
    if nested > 0 {
        text.push_str(&format!(
            "- Folders below with their own manifest: {nested}\n"
        ));
    }
    text
}

/// The ignore line for the worktree folder is already in `.gitignore` or the local exclude file.
fn worktrees_ignored(root: &Path) -> bool {
    [".gitignore", ".git/info/exclude"].iter().any(|rel| {
        read_opt(&root.join(rel)).ok().flatten().is_some_and(|b| {
            String::from_utf8_lossy(&b).lines().any(|l| {
                matches!(
                    l.trim(),
                    ".claude/worktrees/" | ".claude/worktrees" | "/.claude/worktrees/"
                )
            })
        })
    })
}

fn plan_digest(p: &ScalePlan) -> String {
    let mut v = serde_json::to_value(p).expect("plan serializes");
    v["plan_digest"] = Value::Null;
    digest(&SortedKeys.bytes(&v))
}

/// `project-scale-plan`: read-only.
pub fn scale_plan(root: &Path, request: ScaleRequest) -> Result<ScalePlan> {
    if request.protocol_version != 1 {
        return Err("unsupported protocol_version; supported: 1".into());
    }
    if !matches!(request.kind.as_str(), "small" | "large_tree" | "monorepo") {
        return Err("kind refused: it must be small, large_tree or monorepo".into());
    }
    let root = root_of(root)?;
    let docs = docs();
    let line = |id: &str| -> (String, String) {
        let l = &docs["lines"][id];
        (
            l["url"].as_str().unwrap_or_default().to_string(),
            l["quote"].as_str().unwrap_or_default().to_string(),
        )
    };
    let scale = crate::project_scale_detect::read(&root);
    // Every folder below the root that holds a manifest is a package; unlike the detector's
    // list, which stops at its bound, a pick is checked against all of them.
    let manifests = crate::project_detect::nested_manifests(&root);
    let packages: BTreeSet<String> = manifests
        .iter()
        .filter_map(|path| path.rsplit_once('/').map(|(folder, _)| folder.to_string()))
        .collect();
    let mut plan = ScalePlan {
        protocol_version: 1,
        request: request.clone(),
        detected_kind: scale.kind.clone(),
        rows: Vec::new(),
        install: Vec::new(),
        notes: Vec::new(),
        docs_read_on: docs["read_on"].as_str().unwrap_or_default().to_string(),
        plan_digest: String::new(),
    };

    // REQ-11: every pick is checked before anything is planned.
    let package_files = picked_paths("package_files", &request.package_files)?;
    let exclude_packages = picked_paths("exclude_packages", &request.exclude_packages)?;
    let start_packages = picked_paths("start_packages", &request.start_packages)?;
    let deny_read = picked_paths("deny_read", &request.deny_read)?;
    let sparse_paths = picked_paths("sparse_paths", &request.sparse_paths)?;
    if request.plugins.len() > PICKS_MAX {
        return Err(format!(
            "plugins refused: more than {PICKS_MAX} entries; nothing is planned"
        ));
    }
    for (key, list) in [
        ("package_files", &package_files),
        ("exclude_packages", &exclude_packages),
        ("start_packages", &start_packages),
    ] {
        if list.is_empty() {
            continue;
        }
        if request.kind != "monorepo" {
            return Err(format!(
                "{key} refused: packages apply to a monorepo only; nothing is planned"
            ));
        }
        if let Some(p) = list.iter().find(|p| !packages.contains(*p)) {
            return Err(format!(
                "{key} refused: {} is not a detected package; nothing is planned",
                p.chars().take(64).collect::<String>()
            ));
        }
    }
    for (key, list) in [("deny_read", &deny_read), ("sparse_paths", &sparse_paths)] {
        if let Some(p) = list.iter().find(|p| !real_dir(&root, p)) {
            return Err(format!(
                "{key} refused: {} is not a folder in the repository; nothing is planned",
                p.chars().take(64).collect::<String>()
            ));
        }
    }
    let mut plugins: BTreeSet<(&str, &str)> = BTreeSet::new();
    if !request.plugins.is_empty() {
        let detected = crate::project_detect::detect(&root)
            .map(|d| d.languages)
            .unwrap_or_default();
        for language in &request.plugins {
            let found = PLUGINS.iter().find(|(l, _, _)| *l == language.as_str());
            match found {
                Some((l, plugin, binary)) if detected.iter().any(|d| d == l) => {
                    plugins.insert((plugin, binary));
                }
                _ => {
                    let shown: String = language
                        .chars()
                        .filter(|c| !c.is_control())
                        .take(64)
                        .collect();
                    return Err(format!(
                        "plugins refused: {shown} is not a detected language with a code intelligence plugin; nothing is planned"
                    ));
                }
            }
        }
    }

    if request.kind == "small" {
        plan.notes.push(
            "The scale picked is small, so nothing is planned: a project this size needs no large-codebase setup."
                .into(),
        );
        plan.plan_digest = plan_digest(&plan);
        return Ok(plan);
    }

    let mut rows: Vec<ScaleRow> = Vec::new();
    let push = |rows: &mut Vec<ScaleRow>,
                file: &str,
                key: Option<&str>,
                (action, detail): (&str, Option<String>),
                value: Value,
                doc: &str| {
        let (doc_url, doc_quote) = line(doc);
        rows.push(ScaleRow {
            file: file.into(),
            key: key.map(str::to_string),
            action: action.into(),
            value,
            doc_url,
            doc_quote,
            detail,
        });
    };

    // REQ-1: one CLAUDE.md per picked package.
    for package in &package_files {
        let file = format!("{package}/CLAUDE.md");
        let text = package_text(package, &manifests);
        let action = file_action(&root, &file, &text)?;
        push(
            &mut rows,
            &file,
            None,
            action,
            Value::String(text),
            "package_claude_md",
        );
    }
    // REQ-3: one claudeMdExcludes glob per package the user never works in.
    if !exclude_packages.is_empty() {
        let file = ".claude/settings.local.json";
        let value = json!(exclude_packages
            .iter()
            .map(|p| format!("**/{p}/**/CLAUDE.md"))
            .collect::<Vec<_>>());
        let current = settings_of(&root, file)?;
        let action = key_action(&current, "claudeMdExcludes", &value);
        push(
            &mut rows,
            file,
            Some("claudeMdExcludes"),
            action,
            value,
            "claude_md_excludes",
        );
    }
    let root_settings = ".claude/settings.json";
    let current_root = settings_of(&root, root_settings)?;
    // REQ-4: one Read deny rule per picked generated or vendored folder.
    if !deny_read.is_empty() {
        let value = json!(deny_read
            .iter()
            .map(|d| format!("Read(./{d}/**/*)"))
            .collect::<Vec<_>>());
        let action = key_action(&current_root, "permissions.deny", &value);
        push(
            &mut rows,
            root_settings,
            Some("permissions.deny"),
            action,
            value,
            "deny_read",
        );
    }
    // REQ-5: the plugin keys, and the install commands as text.
    if !plugins.is_empty() {
        let mut value = Map::new();
        for (plugin, binary) in &plugins {
            value.insert(format!("{plugin}@{MARKETPLACE}"), Value::Bool(true));
            plan.install.push(format!(
                "/plugin install {plugin}@{MARKETPLACE} (it needs the {binary} language server on PATH; the plugin does not include it)"
            ));
        }
        let value = Value::Object(value);
        let action = key_action(&current_root, "enabledPlugins", &value);
        push(
            &mut rows,
            root_settings,
            Some("enabledPlugins"),
            action,
            value,
            "enabled_plugins",
        );
    }
    // REQ-6: sparse worktrees with .claude first, the root dependency folders linked.
    if !sparse_paths.is_empty() {
        let mut paths = vec![".claude".to_string()];
        paths.extend(sparse_paths.iter().filter(|p| *p != ".claude").cloned());
        let value = json!(paths);
        let action = key_action(&current_root, "worktree.sparsePaths", &value);
        push(
            &mut rows,
            root_settings,
            Some("worktree.sparsePaths"),
            action,
            value,
            "sparse_paths",
        );
        let linked: Vec<&str> = DEPENDENCY_FOLDERS
            .iter()
            .copied()
            .filter(|d| real_dir(&root, d))
            .collect();
        if !linked.is_empty() {
            let value = json!(linked);
            let action = key_action(&current_root, "worktree.symlinkDirectories", &value);
            push(
                &mut rows,
                root_settings,
                Some("worktree.symlinkDirectories"),
                action,
                value,
                "symlink_directories",
            );
        }
        plan.notes.push(
            "With sparsePaths set, git turns on extensions.worktreeConfig in the shared .git/config."
                .into(),
        );
    }
    if request.ignore_worktrees {
        let action = if worktrees_ignored(&root) {
            ("unchanged", None)
        } else {
            (
                "merge",
                Some("one line in the primary checkout's local exclude file, the place the worktree setup already uses".into()),
            )
        };
        push(
            &mut rows,
            ".git/info/exclude",
            None,
            action,
            json!(".claude/worktrees/"),
            "worktree_ignore",
        );
    }
    // REQ-7: a self-contained settings file per package sessions start from, naming its
    // sibling packages that exist.
    for package in &start_packages {
        let siblings: Vec<String> = packages
            .iter()
            .filter(|s| {
                *s != package
                    && !s.starts_with(&format!("{package}/"))
                    && !package.starts_with(&format!("{s}/"))
                    && real_dir(&root, s)
            })
            .map(|s| relative(package, s))
            .take(PICKS_MAX)
            .collect();
        let value = json!({"permissions": {"additionalDirectories": siblings}});
        let file = format!("{package}/.claude/settings.json");
        let mut text = serde_json::to_string_pretty(&value).map_err(|_| "encode failed")?;
        text.push('\n');
        let action = file_action(&root, &file, &text)?;
        push(
            &mut rows,
            &file,
            None,
            action,
            value,
            "additional_directories",
        );
    }

    // REQ-11: a planned value that looks like a credential refuses the whole plan.
    if let Some(r) = rows
        .iter()
        .find(|r| crate::claude::holds_credential(&r.value))
    {
        return Err(format!(
            "{} refused: a planned value looks like a credential; nothing is planned",
            r.key.as_deref().unwrap_or(&r.file)
        ));
    }
    if let Some(r) = rows.iter().find(|r| {
        r.value
            .as_str()
            .is_some_and(|t| t.lines().count() > FILE_LINES_MAX)
    }) {
        return Err(format!(
            "{} refused: more than {FILE_LINES_MAX} lines",
            r.file
        ));
    }

    // REQ-10: where to start Claude.
    let (inherit_url, inherit_quote) = line("package_settings");
    let (plan_url, _) = line("plan_survives");
    plan.notes.push(format!(
        "Where to start Claude: start Claude at the repository root to reach every package, or in a package folder to work inside it. Project settings in `.claude/settings.json` aren't inherited from parent directories the way CLAUDE.md files are, so a package you start in needs its own settings file ({inherit_url}: \"{inherit_quote}\"). For a long task, write the plan in plan mode: the plan Claude wrote in plan mode is re-injected from disk after compaction ({plan_url})."
    ));
    if !deny_read.is_empty() {
        plan.notes.push(
            "A Read deny rule stops Claude's own file tools; a program Claude runs can still open those files, so it is not a security boundary."
                .into(),
        );
    }
    if !start_packages.is_empty() {
        plan.notes.push(
            "additionalDirectories takes effect only after the user trusts the folder, and it grants file access only."
                .into(),
        );
    }
    plan.rows = rows;
    plan.plan_digest = plan_digest(&plan);
    Ok(plan)
}
