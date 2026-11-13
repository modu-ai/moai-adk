//! Base MoAI agents (#2592, stage 2 S1): one tool-neutral source per role, rendered into each
//! host's agent file. The sources live in the `moai` skill package under
//! `references/agents/<role>.md` (decision A: reference documents, never separate skills), so
//! they ship and update with the package and never show in a user skill list.
//!
//! A source is Markdown with a small frontmatter — `name`, `description`, `readonly`
//! (`true|false`), `isolation` (`none|worktree`), `shape` (`sequential|fanout`) — and the
//! instructions as the body. Rendering only uses what the host itself offers (#2592: no
//! substitute implementation); a source field the host cannot express is never dropped
//! silently but listed in `not_supported`. Pure functions: nothing here reads HOME or writes.

use serde::Serialize;
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// Where the sources live inside the skill package.
pub const SOURCE_DIR: &str = "references/agents";
/// Hosts with an agent renderer: Claude Code and Codex (S1), OMP (#2633, user decision
/// 6841920695 and the 2026-11-13 confirmation of five supported hosts, #2592).
pub const HOSTS: [&str; 3] = ["claude", "codex", "omp"];
const DESCRIPTION_LIMIT: usize = 1024;
const BODY_LIMIT: usize = 16 * 1024;

#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct AgentDef {
    pub name: String,
    pub description: String,
    pub readonly: bool,
    /// `none` or `worktree`.
    pub isolation: String,
    /// `sequential` or `fanout`; how a parent starts it, recorded for the workflows (S4).
    pub shape: String,
    pub body: String,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct Rendered {
    pub host: String,
    /// Path relative to the host's user folder (`~/.claude`, `~/.codex`).
    pub path: String,
    pub text: String,
    /// Source fields this host cannot express, each with the reason.
    pub not_supported: Vec<String>,
}

fn valid_name(name: &str) -> bool {
    name.strip_prefix("moai-").is_some_and(|role| {
        !role.is_empty()
            && role.len() <= 32
            && role
                .bytes()
                .all(|b| b.is_ascii_lowercase() || b.is_ascii_digit())
    })
}

/// Parses one source file. Every key is required and appears once; nothing else is allowed.
pub fn parse(text: &str) -> Result<AgentDef> {
    let rest = text
        .strip_prefix("---\n")
        .ok_or("agent source must start with a --- frontmatter line")?;
    let end = rest
        .find("\n---\n")
        .ok_or("agent source frontmatter is not closed")?;
    let mut keys: Vec<(String, String)> = Vec::new();
    for line in rest[..end].lines() {
        let (k, v) = line
            .split_once(':')
            .ok_or_else(|| format!("frontmatter line is not `key: value`: {line:?}"))?;
        let k = k.trim().to_string();
        if keys.iter().any(|(seen, _)| *seen == k) {
            return Err(format!("frontmatter key repeated: {k}"));
        }
        keys.push((k, v.trim().to_string()));
    }
    const KNOWN: [&str; 5] = ["name", "description", "readonly", "isolation", "shape"];
    if let Some((k, _)) = keys.iter().find(|(k, _)| !KNOWN.contains(&k.as_str())) {
        return Err(format!("unknown frontmatter key: {k}"));
    }
    let get = |k: &str| {
        keys.iter()
            .find(|(n, _)| n == k)
            .map(|(_, v)| v.clone())
            .ok_or_else(|| format!("frontmatter key missing: {k}"))
    };
    let name = get("name")?;
    if !valid_name(&name) {
        return Err(format!(
            "agent name must be moai-<lowercase role>: {name:?}"
        ));
    }
    let description = get("description")?;
    if description.is_empty()
        || description.chars().count() > DESCRIPTION_LIMIT
        || description.chars().any(char::is_control)
    {
        return Err(format!(
            "{name}: description must be one line of 1..={DESCRIPTION_LIMIT} characters"
        ));
    }
    let readonly = match get("readonly")?.as_str() {
        "true" => true,
        "false" => false,
        other => {
            return Err(format!(
                "{name}: readonly must be true or false, not {other:?}"
            ))
        }
    };
    let isolation = get("isolation")?;
    if !["none", "worktree"].contains(&isolation.as_str()) {
        return Err(format!("{name}: isolation must be none or worktree"));
    }
    let shape = get("shape")?;
    if !["sequential", "fanout"].contains(&shape.as_str()) {
        return Err(format!("{name}: shape must be sequential or fanout"));
    }
    let body = rest[end + 5..].trim_start_matches('\n').to_string();
    if body.trim().is_empty() || body.len() > BODY_LIMIT {
        return Err(format!(
            "{name}: instructions must be 1..={BODY_LIMIT} bytes"
        ));
    }
    if body.lines().any(|l| l.trim_end() == "---") {
        return Err(format!("{name}: instructions must not hold a --- line"));
    }
    Ok(AgentDef {
        name,
        description,
        readonly,
        isolation,
        shape,
        body,
    })
}

/// Every source under `<package>/references/agents/`, sorted by file name. A package without
/// the folder has no agents.
pub fn package_agents(package: &Path) -> Result<Vec<AgentDef>> {
    let dir = package.join(SOURCE_DIR);
    let entries = match std::fs::read_dir(&dir) {
        Ok(entries) => entries,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(Vec::new()),
        Err(_) => return Err("agent source folder unreadable".into()),
    };
    let mut names: Vec<String> = Vec::new();
    for entry in entries {
        let entry = entry.map_err(|_| "agent source folder unreadable")?;
        let name = entry.file_name().to_string_lossy().into_owned();
        if !entry
            .file_type()
            .map_err(|_| "agent source unreadable")?
            .is_file()
            || !name.ends_with(".md")
        {
            return Err(format!(
                "{SOURCE_DIR}/{name}: only regular .md files belong here"
            ));
        }
        names.push(name);
    }
    names.sort();
    let mut defs = Vec::new();
    for file in names {
        let text = std::fs::read_to_string(dir.join(&file))
            .map_err(|_| format!("{SOURCE_DIR}/{file}: not readable UTF-8"))?;
        let def = parse(&text).map_err(|e| format!("{SOURCE_DIR}/{file}: {e}"))?;
        let role = file.trim_end_matches(".md");
        if def.name != format!("moai-{role}") {
            return Err(format!("{SOURCE_DIR}/{file}: name must be moai-{role}"));
        }
        defs.push(def);
    }
    Ok(defs)
}

fn yaml_scalar(s: &str) -> String {
    // A double-quoted YAML scalar; the description is one line without control characters.
    format!("\"{}\"", s.replace('\\', "\\\\").replace('"', "\\\""))
}

/// Claude Code subagent file `agents/<name>.md` (https://code.claude.com/docs/en/sub-agents):
/// `name`, `description`, `disallowedTools` for read-only roles, `isolation: worktree`.
fn render_claude(def: &AgentDef) -> Rendered {
    let mut fm = format!(
        "---\nname: {}\ndescription: {}\n",
        def.name,
        yaml_scalar(&def.description)
    );
    if def.readonly {
        fm.push_str("disallowedTools: Write, Edit, NotebookEdit\n");
    }
    if def.isolation == "worktree" {
        fm.push_str("isolation: worktree\n");
    }
    fm.push_str("---\n\n");
    Rendered {
        host: "claude".into(),
        path: format!("agents/{}.md", def.name),
        text: fm + &def.body,
        not_supported: Vec::new(),
    }
}

/// Codex agent file `agents/<name>.toml`: `name`, `description`, `developer_instructions`.
/// Codex agents inherit the parent sandbox and have no isolation field, so read-only and
/// worktree isolation are instructions only (#2592 6840103566 §2).
fn render_codex(def: &AgentDef) -> Rendered {
    use crate::generation::toml_string;
    let mut not_supported = Vec::new();
    let mut body = def.body.clone();
    if def.readonly {
        not_supported.push(
            "readonly: a Codex agent inherits the parent sandbox; read-only is an instruction only"
                .into(),
        );
        body.push_str("\nRead-only: this host cannot enforce it, so do not write any file.\n");
    }
    if def.isolation == "worktree" {
        not_supported.push(
            "isolation: Codex agents have no worktree isolation; they run in the parent folder"
                .into(),
        );
    }
    Rendered {
        host: "codex".into(),
        path: format!("agents/{}.toml", def.name),
        text: format!(
            "name = {}\ndescription = {}\ndeveloper_instructions = {}\n",
            toml_string(&def.name),
            toml_string(&def.description),
            toml_string(&body)
        ),
        not_supported,
    }
}

/// Built-in tools an OMP agent gets. OMP reads `tools` as an allow-list with no deny form,
/// and a list holding `task` lets the agent start any other agent even with `spawns: false`
/// (`discovery/helpers.ts`), so the list is explicit and never holds `task`: the one-level
/// nesting rule is enforced by the host. Read 2026-11-13 from github.com/can1357/oh-my-pi
/// (`src/tools/builtin-names.ts`, `docs/task-agent-discovery.md`).
const OMP_TOOLS: [&str; 8] = [
    "read",
    "grep",
    "glob",
    "find",
    "bash",
    "web_search",
    "edit",
    "write",
];
/// The tools a read-only OMP agent drops. `bash` stays, as in the Claude Code file, so an
/// audit can still run checks.
const OMP_WRITE_TOOLS: [&str; 2] = ["edit", "write"];

/// OMP agent file `agents/<name>.md` under `~/.omp/agent/` (Markdown with YAML frontmatter;
/// `name`, `description` and an explicit `tools` list). OMP skips `.claude/agents`, so it
/// needs its own file. Worktree isolation is chosen per task call (`isolated: true` with
/// `task.isolation.enabled`), not in the agent file, so it is named in `not_supported`.
fn render_omp(def: &AgentDef) -> Rendered {
    let tools: Vec<&str> = OMP_TOOLS
        .iter()
        .copied()
        .filter(|t| !(def.readonly && OMP_WRITE_TOOLS.contains(t)))
        .collect();
    let mut not_supported = Vec::new();
    if def.isolation == "worktree" {
        not_supported.push(
            "isolation: OMP isolates per task call (isolated: true with task.isolation.enabled), \
             not in the agent file"
                .into(),
        );
    }
    Rendered {
        host: "omp".into(),
        path: format!("agents/{}.md", def.name),
        text: format!(
            "---\nname: {}\ndescription: {}\ntools: {}\n---\n\n{}",
            def.name,
            yaml_scalar(&def.description),
            tools.join(", "),
            def.body
        ),
        not_supported,
    }
}

/// The agent file for one host.
pub fn render(def: &AgentDef, host: &str) -> Result<Rendered> {
    match host {
        "claude" => Ok(render_claude(def)),
        "codex" => Ok(render_codex(def)),
        "omp" => Ok(render_omp(def)),
        other => Err(format!("agent host not supported yet: {other:?}")),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    const SAMPLE: &str = "---\nname: moai-x\ndescription: An agent.\nreadonly: true\nisolation: worktree\nshape: sequential\n---\n\nDo it.\n";

    #[test]
    fn a_source_parses_and_rejects_bad_frontmatter() {
        let def = parse(SAMPLE).unwrap();
        assert_eq!(def.name, "moai-x");
        assert!(def.readonly);
        assert_eq!(def.body, "Do it.\n");
        for (bad, why) in [
            (SAMPLE.replace("name: moai-x", "name: x"), "agent name"),
            (SAMPLE.replace("name: moai-x", "name: moai-X"), "agent name"),
            (
                SAMPLE.replace("readonly: true", "readonly: yes"),
                "readonly",
            ),
            (
                SAMPLE.replace("isolation: worktree", "isolation: vm"),
                "isolation",
            ),
            (SAMPLE.replace("shape: sequential", "shape: swarm"), "shape"),
            (SAMPLE.replace("shape: sequential\n", ""), "missing: shape"),
            (
                SAMPLE.replace("shape: sequential", "shape: sequential\nmodel: opus"),
                "unknown",
            ),
            (
                SAMPLE.replace("shape: sequential", "shape: sequential\nshape: fanout"),
                "repeated",
            ),
            (SAMPLE.replace("Do it.", "Do it.\n---\nmore"), "--- line"),
            (SAMPLE.replace("Do it.\n", ""), "instructions"),
            (SAMPLE.replacen("---\n", "", 1), "must start"),
        ] {
            let e = parse(&bad).unwrap_err();
            assert!(e.contains(why), "{why}: {e}");
        }
    }

    #[test]
    fn renderers_use_only_host_fields_and_name_what_they_cannot_express() {
        let def = parse(SAMPLE).unwrap();
        let claude = render(&def, "claude").unwrap();
        assert_eq!(claude.path, "agents/moai-x.md");
        assert_eq!(
            claude.text,
            "---\nname: moai-x\ndescription: \"An agent.\"\ndisallowedTools: Write, Edit, NotebookEdit\nisolation: worktree\n---\n\nDo it.\n"
        );
        assert!(claude.not_supported.is_empty());
        let codex = render(&def, "codex").unwrap();
        assert_eq!(codex.path, "agents/moai-x.toml");
        assert!(codex
            .text
            .starts_with("name = \"moai-x\"\ndescription = \"An agent.\"\n"));
        assert!(codex.text.contains("do not write any file"));
        assert_eq!(codex.not_supported.len(), 2);
        let writer = parse(
            &SAMPLE
                .replace("readonly: true", "readonly: false")
                .replace("isolation: worktree", "isolation: none"),
        )
        .unwrap();
        assert!(!render(&writer, "claude")
            .unwrap()
            .text
            .contains("disallowedTools"));
        assert!(render(&writer, "codex").unwrap().not_supported.is_empty());
        assert!(render(&def, "goose").is_err());
    }

    /// #2633: the OMP file names an explicit tool list without `task` (one nesting level,
    /// enforced by the host), drops `edit`/`write` for a read-only role, and names the
    /// per-call worktree isolation it cannot put in the file.
    #[test]
    fn omp_file_has_an_explicit_tool_list_without_task() {
        let def = parse(SAMPLE).unwrap();
        let omp = render(&def, "omp").unwrap();
        assert_eq!(omp.path, "agents/moai-x.md");
        assert_eq!(
            omp.text,
            "---\nname: moai-x\ndescription: \"An agent.\"\ntools: read, grep, glob, find, bash, web_search\n---\n\nDo it.\n"
        );
        assert_eq!(omp.not_supported.len(), 1);
        assert!(omp.not_supported[0].starts_with("isolation:"));
        let writer = parse(
            &SAMPLE
                .replace("readonly: true", "readonly: false")
                .replace("isolation: worktree", "isolation: none"),
        )
        .unwrap();
        let omp = render(&writer, "omp").unwrap();
        assert!(omp
            .text
            .contains("\ntools: read, grep, glob, find, bash, web_search, edit, write\n"));
        assert!(omp.not_supported.is_empty());
        for role in [&def, &writer] {
            let text = render(role, "omp").unwrap().text;
            let tools = text.lines().find(|l| l.starts_with("tools: ")).unwrap();
            assert!(!tools.split(", ").any(|t| t.ends_with("task")), "{tools}");
        }
    }
}
