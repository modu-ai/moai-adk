//! Project-specific file generation: plan read-only, apply with the merged 3-way
//! decision and receipts (Issue #1997, 03-C).
//!
//! The AGENTS.md managed section, the personal file opt-in and the `.gitignore`
//! section are the existing agents machinery: `generation` derives an
//! `AgentsRequest`, calls the public `agents_plan` for its read-only judgment and
//! the `agents_apply` engine for its publication (receipts, baseline recovery, manifest
//! write included) through the crate-internal `agents_apply_amended` hook, so the
//! generation state joins the one manifest publish (#2114). `src/agents.rs` is 01-B
//! #1989's; its public API and behavior are unchanged; the
//! generation-specific files — portable project workflows under
//! `.moai-cli/workflows/<id>/`, native role files `.claude/agents/<role>.md` and
//! `.codex/agents/<role>.toml`, and skill projections `.claude/skills/<id>/SKILL.md`
//! (`.agents/skills/…`) — are planned and published in this module through the public
//! 01-B apply engine (`decide`, `publish_exclusive`/`publish_replace`,
//! `save_before_bytes`, `Receipt`). Baselines of generation files live in the same
//! `generation.json` `owned_files` list, recorded with an additive `kind` field
//! (`workflow`, `native-role`, `skill-projection`; #2109) plus a `section_id` tag of
//! the shape `generation:<kind>:…` kept so manifests written before the kind field
//! stay readable (agents.rs's `OwnedFile` rejects unknown fields, and its own writes
//! leave `kind` unset).
//!
//! The delegation is stale-safe: `generation_apply` first re-plans the whole
//! generation request and refuses any drift, then re-plans the agents portion and
//! hands `agents_apply` its own freshly computed digest — a plan that passed the
//! outer check passes the inner one unchanged.
//!
//! Boundaries: `.claude/settings.json` and `.codex/config.toml` merges stay out
//! (HOLD-2 A, owner 02-B #1992) — Codex role files carry only `name`, `description`
//! and `developer_instructions`, never the keys that widen parent permissions.
//! Markdown merging stays at managed-section granularity (HOLD-1 A; the diff3 study
//! shows line-level merging has no general non-conflict guarantee). The subcommand
//! names are `generation-plan`/`generation-apply` (HOLD-3 A). External skill content
//! is never copied: a projection references a `reviewed` source with an attribution
//! line, and a source whose license is not reviewed blocks every file that cites it
//! (REQ-03C-08).

use crate::agents::{
    agents_apply_amended_for, agents_plan_for, AgentsPlan, AgentsRequest, FilePlan,
};
use crate::apply::{
    decide, new_request_id, publish_exclusive, publish_replace, read_opt, root_of,
    save_before_bytes, Decision, Receipt,
};
use crate::{digest, read_profile};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{fs, path::Path};

type Result<T> = std::result::Result<T, String>;

/// The generation manifest of `root`, under its state folder (#2725 S2b).
fn manifest_rel(root: &Path) -> String {
    crate::state_root::rel(root, "manifests/generation.json")
}
const MAX_ROLES: usize = 16;
const MAX_PROJECTIONS: usize = 16;
/// Generated text stays short: Codex reads the whole AGENTS.md chain within 32 KiB
/// by default, and the managed section must leave room for user content.
const PURPOSE_LIMIT: usize = 500;

/// Skill names follow the Agent Skills specification (same rule as 03-B).
fn valid_skill_name(s: &str) -> bool {
    !s.is_empty()
        && s.len() <= 64
        && s.bytes()
            .all(|b| b.is_ascii_lowercase() || b.is_ascii_digit() || b == b'-')
        && !s.starts_with('-')
        && !s.ends_with('-')
        && !s.contains("--")
}

fn valid_revision(s: &str) -> bool {
    s.len() == 40
        && s.bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}

/// REQ-03C-07: no rendered string may carry managed-marker tokens or a frontmatter
/// delimiter line, and role purposes must stay single-line short.
fn render_safe(s: &str) -> Result<()> {
    let lower = s.to_ascii_lowercase();
    if lower.contains("<!--")
        || lower.contains("-->")
        || lower.contains("moai:managed:")
        || s.lines().any(|l| l.trim() == "---")
    {
        return Err("rendered value contains managed-marker or frontmatter syntax; refused".into());
    }
    Ok(())
}

fn bounded_purpose(s: &str) -> bool {
    !s.trim().is_empty() && s.len() <= PURPOSE_LIMIT && !s.contains('\n')
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct NativeRole {
    pub role_id: String,
    /// claude | codex
    pub host: String,
    pub purpose: String,
    #[serde(default)]
    pub read_only_intent: bool,
    /// Claude-only allowed-tools list (Codex files never carry tool keys).
    #[serde(default)]
    pub tools: Vec<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct SkillProjection {
    /// Installed skill name; also the projected directory name (REQ-03C-09).
    pub id: String,
    /// `source_url` of an entry in `.moai-cli/manifests/sources.json`.
    pub source_url: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub resolved_revision: Option<String>,
    /// claude | codex (default: claude)
    #[serde(default)]
    pub host_targets: Vec<String>,
}

/// A project skill (#2706 P2): its body comes from the profile's interview answers, not from
/// the request, so `project.yaml` stays the one source.
#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ProjectSkill {
    /// `prj-<name>`; also the skill directory name.
    pub id: String,
    /// claude | codex (default: both).
    #[serde(default)]
    pub hosts: Vec<String>,
    /// One line; the skill's `description`.
    pub purpose: String,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct GenerationRequest {
    pub protocol_version: u32,
    /// en | ko (feeds the delegated agents rendering too).
    pub language: String,
    /// Repo-relative directory that must exist (agents_plan contract).
    pub workflow_reference: String,
    /// none | ignored | tracked (the personal-file opt-in, passed through).
    pub local_file: String,
    /// Portable workflow id; the generated directory name.
    pub workflow_id: String,
    #[serde(default)]
    pub native_roles: Vec<NativeRole>,
    #[serde(default)]
    pub skill_projections: Vec<SkillProjection>,
    /// Absent when empty, so a request without project skills digests as before.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub project_skills: Vec<ProjectSkill>,
}

fn project_skill_hosts(s: &ProjectSkill) -> Vec<String> {
    if s.hosts.is_empty() {
        vec!["claude".into(), "codex".into()]
    } else {
        s.hosts.clone()
    }
}

/// One planned file. The agents portion carries the `agents.rs` fields it produced;
/// generation files add their `source_refs` and kind tag.
#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct GenerationFilePlan {
    pub path: String,
    /// create | merge | append | unchanged | conflict
    pub action: String,
    pub tracked_by_default: bool,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub expected_sha256: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub baseline_sha256: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub rendered_sha256: Option<String>,
    #[serde(default)]
    pub source_refs: Vec<String>,
    /// agents:<section> | generation:workflow:<id> |
    /// generation:native-role:<host>:<role> | generation:skill-projection:<host>:<id>
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub kind: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub detail: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct GenerationPlan {
    pub protocol_version: u32,
    pub request: GenerationRequest,
    pub target_binding: String,
    pub files: Vec<GenerationFilePlan>,
    /// Pass-through of the agents host_notes (CLAUDE.md family) plus generation
    /// warnings (read-only-intent, license gate).
    pub host_notes: Vec<String>,
    pub warnings: Vec<String>,
    pub plan_digest: String,
}

fn validate(r: &GenerationRequest) -> Result<()> {
    if r.protocol_version != 1
        || !(r.language == "en" || r.language == "ko")
        || !["none", "ignored", "tracked"].contains(&r.local_file.as_str())
        || !valid_skill_name(&r.workflow_id)
        || r.workflow_reference.trim().is_empty()
        || r.workflow_reference.len() > 256
        || r.native_roles.len() > MAX_ROLES
        || r.skill_projections.len() > MAX_PROJECTIONS
        || r.project_skills.len() > MAX_PROJECTIONS
    {
        return Err("invalid generation request; no values were logged".into());
    }
    for role in &r.native_roles {
        // A project role is named `prj-<name>` (leader design 6842177115, #2706): it never
        // shadows a user-layer `moai-*` agent, and a name without the prefix is refused. The
        // bare `prj-` fails `valid_skill_name` (no trailing `-`).
        if !valid_skill_name(&role.role_id)
            || !role.role_id.starts_with("prj-")
            || !(role.host == "claude" || role.host == "codex")
            || !bounded_purpose(&role.purpose)
            || role.tools.len() > 32
            || !role
                .tools
                .iter()
                .all(|t| !t.is_empty() && t.len() <= 128 && !t.contains(char::is_whitespace))
        {
            return Err("invalid generation request; no values were logged".into());
        }
    }
    for p in &r.skill_projections {
        if !valid_skill_name(&p.id)
            || p.source_url.trim().is_empty()
            || p.source_url.len() > 2048
            || p.resolved_revision
                .as_ref()
                .is_some_and(|rev| !valid_revision(rev))
            || p.host_targets.iter().any(|h| h != "claude" && h != "codex")
        {
            return Err("invalid generation request; no values were logged".into());
        }
    }
    let mut seen_ids = std::collections::BTreeSet::new();
    for ps in &r.project_skills {
        if !seen_ids.insert(ps.id.as_str())
            || ps
                .hosts
                .iter()
                .collect::<std::collections::BTreeSet<_>>()
                .len()
                != ps.hosts.len()
        {
            return Err("invalid generation request: a project skill id or host repeats".into());
        }
        if !valid_skill_name(&ps.id)
            || !ps.id.starts_with("prj-")
            || !bounded_purpose(&ps.purpose)
            || ps.hosts.iter().any(|h| h != "claude" && h != "codex")
        {
            return Err("invalid generation request; no values were logged".into());
        }
        render_safe(&ps.purpose)?;
    }
    for role in &r.native_roles {
        render_safe(&role.purpose)?;
        render_safe(&role.role_id)?;
    }
    render_safe(&r.workflow_id)?;
    Ok(())
}

/// Reads `generation.json` as a Value. agents.rs owns the typed schema; the
/// generation side only rewrites keys the schema already knows.
fn manifest_value(root: &Path) -> Result<Value> {
    let path = root.join(manifest_rel(root));
    match fs::metadata(&path) {
        Ok(m) if !m.is_file() || m.len() > 262_144 => {
            Err("target is not a bounded regular file".into())
        }
        Ok(_) => {
            let bytes = fs::read(&path).map_err(|_| "target read failed".to_string())?;
            serde_json::from_slice(&bytes)
                .map_err(|_| "generation manifest unreadable; inspect before generating".into())
        }
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(Value::Null),
        Err(_) => Err("target metadata unavailable".into()),
    }
}

/// The recorded baseline digest of one generation file, from the manifest's
/// `owned_files` (entries tagged with a `generation:` section id).
fn generation_baseline(manifest: &Value, path: &str) -> Option<String> {
    manifest
        .get("owned_files")
        .and_then(Value::as_array)
        .and_then(|list| {
            list.iter()
                .find(|f| f.get("path").and_then(Value::as_str) == Some(path))
                .and_then(|f| f.get("baseline_sha256"))
                .and_then(Value::as_str)
                .map(str::to_string)
        })
}

/// License gate (REQ-03C-08): the source entry in `sources.json` must carry
/// `license.status: reviewed` for anything to reference it. Returns the attribution
/// pair (resolved revision, license note) when reviewed.
fn source_reviewed(root: &Path, source_url: &str) -> Result<Option<(String, String)>> {
    let path = crate::state_root::path(root, "manifests/sources.json");
    let bytes = match fs::read(&path) {
        Ok(b) => b,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(None),
        Err(_) => return Err("target read failed".into()),
    };
    let manifest: Value = serde_json::from_slice(&bytes)
        .map_err(|_| "sources manifest unreadable; inspect before generating".to_string())?;
    for s in manifest
        .get("sources")
        .and_then(Value::as_array)
        .into_iter()
        .flatten()
    {
        if s.get("source_url").and_then(Value::as_str) == Some(source_url) {
            let status = s
                .get("license")
                .and_then(|l| l.get("status"))
                .and_then(Value::as_str)
                .unwrap_or("unknown");
            if status != "reviewed" {
                return Ok(None);
            }
            let revision = s
                .get("resolved_revision")
                .and_then(Value::as_str)
                .unwrap_or("unpinned");
            let license = s
                .get("license")
                .and_then(|l| l.get("spdx"))
                .and_then(Value::as_str)
                .or_else(|| {
                    s.get("license")
                        .and_then(|l| l.get("evidence"))
                        .and_then(Value::as_str)
                })
                .unwrap_or("evidence recorded");
            return Ok(Some((revision.to_string(), license.to_string())));
        }
    }
    Ok(None)
}

pub(crate) fn toml_string(s: &str) -> String {
    let mut out = String::from("\"");
    for c in s.chars() {
        match c {
            '"' => out.push_str("\\\""),
            '\\' => out.push_str("\\\\"),
            '\u{8}' => out.push_str("\\b"),
            '\t' => out.push_str("\\t"),
            '\n' => out.push_str("\\n"),
            '\u{c}' => out.push_str("\\f"),
            '\r' => out.push_str("\\r"),
            c if (c as u32) < 0x20 => out.push_str(&format!("\\u{:04X}", c as u32)),
            c => out.push(c),
        }
    }
    out.push('"');
    out
}

fn render_workflow_skill(workflow_id: &str, language: &str) -> String {
    if language == "ko" {
        format!("# {id} 워크플로\n\n이 프로젝트 전용 워크플로입니다. 절차와 판단 기준은 `references/workflow.md`에 정리합니다.\n",
            id = workflow_id)
    } else {
        format!("# {id} workflow\n\nA project-specific workflow. Procedures and judgment notes live in `references/workflow.md`.\n",
            id = workflow_id)
    }
}

fn render_workflow_reference(workflow_id: &str, language: &str) -> String {
    if language == "ko" {
        format!("# {id} 절차\n\n1. 대상과 범위를 확인한다.\n2. 검증 가능한 완료 조건을 먼저 정의한다.\n3. 결과와 증거를 기록한다.\n",
            id = workflow_id)
    } else {
        format!("# {id} procedure\n\n1. Confirm the target and scope.\n2. Define a verifiable done condition first.\n3. Record the result and the evidence.\n",
            id = workflow_id)
    }
}

fn render_claude_role(role: &NativeRole) -> String {
    let tools = if role.tools.is_empty() {
        String::new()
    } else {
        format!("tools: {}\n", role.tools.join(", "))
    };
    format!(
        "---\nname: {id}\ndescription: {purpose}\n{tools}---\n\n{purpose}\n",
        id = role.role_id,
        purpose = role.purpose,
        tools = tools,
    )
}

fn render_codex_role(role: &NativeRole) -> String {
    format!(
        "name = {name}\ndescription = {desc}\ndeveloper_instructions = {instr}\n",
        name = toml_string(&role.role_id),
        desc = toml_string(&role.purpose),
        instr = toml_string(&role.purpose),
    )
}

/// A YAML double-quoted scalar: a purpose such as `Run tests: cargo test` keeps the
/// frontmatter valid (pre-audit 6854178087 N2). `render_safe` already refused line breaks.
fn yaml_quoted(s: &str) -> String {
    format!("\"{}\"", s.replace('\\', "\\\\").replace('"', "\\\""))
}

/// The first component from `root` down to `rel` (the file included) that is a symbolic link,
/// if any. Writing or removing through it would act outside the project (pre-audit
/// 6854178087 B1, 6854760758 B1), so such a path is a conflict, never followed.
fn link_on_path(root: &Path, rel: &str) -> Option<String> {
    let mut at = root.to_path_buf();
    let mut shown = String::new();
    for part in Path::new(rel).components() {
        at.push(part);
        if !shown.is_empty() {
            shown.push('/');
        }
        shown.push_str(&part.as_os_str().to_string_lossy());
        match fs::symlink_metadata(&at) {
            Ok(m) if m.file_type().is_symlink() => return Some(shown),
            Ok(_) => {}
            Err(_) => return None,
        }
    }
    None
}

/// Where a project skill of `host` lives; `None` for a host this tool does not write.
fn project_skill_path(host: &str, id: &str) -> Option<String> {
    match host {
        "codex" => Some(format!(".agents/skills/{id}/SKILL.md")),
        "claude" => Some(format!(".claude/skills/{id}/SKILL.md")),
        _ => None,
    }
}

/// The project skill body (#2706 P2): the profile's interview answers, one line each; an answer
/// recorded with `source: default` (an axis left unanswered) is never written.
/// The profile a project skill is built from, in the project's state folder (#2725 S2b).
fn project_profile_rel(root: &Path) -> String {
    crate::state_root::rel(root, "project.yaml")
}

fn render_project_skill(
    skill: &ProjectSkill,
    language: &str,
    profile: &crate::Profile,
    source: &str,
) -> Result<String> {
    let ko = language == "ko";
    let answer = |id: &str| {
        profile
            .decisions
            .iter()
            .filter(|d| d.id == id && d.source != "default")
            .map(|d| d.value.clone())
            .collect::<Vec<_>>()
    };
    let sections: [(&str, &str, &str); 4] = [
        ("test_command", "Run the tests", "시험 실행"),
        ("build_command", "Build", "빌드"),
        (
            "review_blocks",
            "Review always stops",
            "리뷰에서 항상 막는 것",
        ),
        (
            "protected",
            "Never change without asking first",
            "먼저 묻지 않고 바꾸지 않는 것",
        ),
    ];
    let mut body = if ko {
        format!("# {}\n\n`/moai project` 인터뷰 답으로 만든 프로젝트 스킬이다. 원본은 `{source}`이다.\n", skill.id)
    } else {
        format!("# {}\n\nProject skill built from the `/moai project` interview answers; the source is `{source}`.\n", skill.id)
    };
    for (id, en, kr) in sections {
        for value in answer(id) {
            render_safe(&value)?;
            body.push_str(&format!("\n- {}: {value}", if ko { kr } else { en }));
        }
    }
    body.push('\n');
    Ok(format!(
        "---\nname: {id}\ndescription: {desc}\n---\n\n{body}",
        id = skill.id,
        desc = yaml_quoted(&skill.purpose),
    ))
}

fn render_projection_skill(
    projection: &SkillProjection,
    language: &str,
    source: &(String, String),
) -> Result<String> {
    let attribution = format!(
        "source: {} @ {} (license: {})",
        projection.source_url, source.0, source.1
    );
    render_safe(&attribution)?;
    let body = if language == "ko" {
        format!("# {id} 참조\n\n이 파일은 외부 skill을 복제하지 않고 참조만 기록합니다.\n{attribution}\n",
            id = projection.id)
    } else {
        format!("# {id} reference\n\nThis file references an external skill without copying it.\n{attribution}\n",
            id = projection.id)
    };
    Ok(format!(
        "---\nname: {id}\ndescription: {desc}\n---\n\n{body}",
        id = projection.id,
        desc = if language == "ko" {
            "외부 skill 참조 기록"
        } else {
            "external skill reference record"
        },
    ))
}

fn plan_digest(p: &GenerationPlan) -> String {
    digest(
        &serde_json::to_vec(&(
            p.protocol_version,
            &p.request,
            &p.target_binding,
            &p.files,
            &p.host_notes,
            &p.warnings,
        ))
        .expect("typed plan"),
    )
}

fn file_from_agents(f: &FilePlan) -> GenerationFilePlan {
    GenerationFilePlan {
        path: f.path.clone(),
        action: f.action.clone(),
        tracked_by_default: f.tracked_by_default,
        expected_sha256: f.expected_sha256.clone(),
        baseline_sha256: f.baseline_sha256.clone(),
        rendered_sha256: f.rendered_sha256.clone(),
        source_refs: Vec::new(),
        kind: Some(format!(
            "agents:{}",
            f.section_id.clone().unwrap_or_else(|| "personal".into())
        )),
        detail: f.detail.clone(),
    }
}

fn three_way_action(
    baseline: Option<&str>,
    current_digest: Option<&str>,
    rendered_digest: &str,
) -> &'static str {
    match decide(baseline, current_digest, Some(rendered_digest)) {
        Decision::KeepCurrent => "unchanged",
        Decision::Apply => "merge",
        Decision::Conflict => "conflict",
    }
}

/// The create/merge refinement of an Apply decision: no current bytes means create.
fn is_create(current: Option<&Vec<u8>>) -> bool {
    current.is_none()
}

/// `generation-plan`: the read-only judgment (REQ-03C-01). The agents portion is the
/// public `agents_plan`; nothing is written.
/// Canonical `workflow_reference` form: exactly one trailing slash. The recorded
/// configuration and the existing golden AGENTS.md renders carry the slash, and
/// `agents.rs` (untouched here) embeds the value verbatim — so the two spellings of
/// one meaning are normalized to the slash-terminated form before planning (Issue
/// #2095). A canonical request renders byte-identically to before.
fn normalize_workflow_reference(raw: &str) -> String {
    format!("{}/", raw.trim_end_matches('/'))
}

/// The canonical `workflow_reference` names a directory this request's own apply writes:
/// `.moai-cli/workflows/` or `.moai-cli/workflows/<workflow_id>/` (#2274, REQ-1). Only then
/// may the agents plan skip its "directory exists" check; any other reference keeps it.
fn reference_is_own_output(root: &Path, request: &GenerationRequest) -> bool {
    let base = format!("{}/workflows/", crate::state_root::dir(root));
    request.workflow_reference == base
        || request.workflow_reference == format!("{base}{}/", request.workflow_id)
}

pub fn generation_plan(root: &Path, request: GenerationRequest) -> Result<GenerationPlan> {
    let mut request = request;
    request.workflow_reference = normalize_workflow_reference(&request.workflow_reference);
    validate(&request)?;
    let root = root
        .canonicalize()
        .map_err(|_| "project directory unavailable")?;
    let agents_request = AgentsRequest {
        protocol_version: 1,
        language: request.language.clone(),
        workflow_reference: request.workflow_reference.clone(),
        local_file: request.local_file.clone(),
    };
    let agents = agents_plan_for(
        &root,
        agents_request,
        reference_is_own_output(&root, &request),
    )?;
    let manifest = manifest_value(&root)?;
    let mut files: Vec<GenerationFilePlan> = agents.files.iter().map(file_from_agents).collect();
    let mut warnings: Vec<String> = agents.warnings.clone();
    let host_notes: Vec<String> = agents.host_notes.clone();

    // Portable workflow (REQ-03C-01: one workflow per request).
    let workflow_rel = crate::state_root::rel(
        &root,
        &format!("workflows/{}/SKILL.md", request.workflow_id),
    );
    let workflow_ref_rel = crate::state_root::rel(
        &root,
        &format!("workflows/{}/references/workflow.md", request.workflow_id),
    );
    let rendered_skill = render_workflow_skill(&request.workflow_id, &request.language);
    let rendered_ref = render_workflow_reference(&request.workflow_id, &request.language);
    for (rel, rendered) in [
        (&workflow_rel, rendered_skill),
        (&workflow_ref_rel, rendered_ref),
    ] {
        let current = fs::read(root.join(rel)).ok();
        let baseline = generation_baseline(&manifest, rel);
        let rendered_digest = digest(rendered.as_bytes());
        let action = three_way_action(
            baseline.as_deref(),
            current.as_deref().map(digest).as_deref(),
            &rendered_digest,
        );
        let action = if action == "merge" && is_create(current.as_ref()) {
            "create"
        } else {
            action
        };
        files.push(GenerationFilePlan {
            path: rel.into(),
            action: action.into(),
            tracked_by_default: true,
            expected_sha256: current.as_deref().map(digest),
            baseline_sha256: baseline,
            rendered_sha256: Some(rendered_digest),
            source_refs: Vec::new(),
            kind: Some(format!("generation:workflow:{}", request.workflow_id)),
            detail: None,
        });
    }

    // Native roles (REQ-03C-05/06): nothing under .claude/agents/ or .codex/agents/
    // is planned when the request lists no roles.
    for role in &request.native_roles {
        let (rel, rendered) = match role.host.as_str() {
            "claude" => (
                format!(".claude/agents/{}.md", role.role_id),
                render_claude_role(role),
            ),
            _ => (
                format!(".codex/agents/{}.toml", role.role_id),
                render_codex_role(role),
            ),
        };
        if role.host == "codex" && role.read_only_intent {
            warnings.push(
                "read_only_intent is guidance; a Codex agent file cannot create a sandbox".into(),
            );
        }
        let current = fs::read(root.join(&rel)).ok();
        let baseline = generation_baseline(&manifest, &rel);
        let rendered_digest = digest(rendered.as_bytes());
        let action = three_way_action(
            baseline.as_deref(),
            current.as_deref().map(digest).as_deref(),
            &rendered_digest,
        );
        let action = if action == "merge" && is_create(current.as_ref()) {
            "create"
        } else {
            action
        };
        files.push(GenerationFilePlan {
            path: rel,
            action: action.into(),
            tracked_by_default: false,
            expected_sha256: current.as_deref().map(digest),
            baseline_sha256: baseline,
            rendered_sha256: Some(rendered_digest),
            source_refs: Vec::new(),
            kind: Some(format!(
                "generation:native-role:{}:{}",
                role.host, role.role_id
            )),
            detail: None,
        });
    }

    // Skill projections (REQ-03C-08/09): license gate first, attribution rendered.
    for p in &request.skill_projections {
        let hosts: Vec<String> = if p.host_targets.is_empty() {
            vec!["claude".into()]
        } else {
            p.host_targets.clone()
        };
        let Some(source) = source_reviewed(&root, &p.source_url)? else {
            warnings.push(format!(
                "source {} is not license-reviewed; projection {} is not planned",
                p.source_url, p.id
            ));
            continue;
        };
        let rendered = render_projection_skill(p, &request.language, &source)?;
        let rendered_digest = digest(rendered.as_bytes());
        for host in &hosts {
            let rel = match host.as_str() {
                "codex" => format!(".agents/skills/{}/SKILL.md", p.id),
                _ => format!(".claude/skills/{}/SKILL.md", p.id),
            };
            let dest_dir = root
                .join(&rel)
                .parent()
                .expect("projection parent")
                .to_path_buf();
            let current = fs::read(root.join(&rel)).ok();
            let baseline = generation_baseline(&manifest, &rel);
            // REQ-03C-09: an existing skill directory without a recorded baseline is
            // a conflict — MoAI never claims a directory it did not project before.
            let unclaimed = baseline.is_none()
                && (current.is_some() || fs::symlink_metadata(&dest_dir).is_ok());
            let action = if unclaimed {
                "conflict"
            } else {
                let decided = three_way_action(
                    baseline.as_deref(),
                    current.as_deref().map(digest).as_deref(),
                    &rendered_digest,
                );
                if decided == "merge" && is_create(current.as_ref()) {
                    "create"
                } else {
                    decided
                }
            };
            files.push(GenerationFilePlan {
                path: rel,
                action: action.into(),
                tracked_by_default: false,
                expected_sha256: current.as_deref().map(digest),
                baseline_sha256: baseline,
                rendered_sha256: Some(rendered_digest.clone()),
                source_refs: vec![p.source_url.clone()],
                kind: Some(format!("generation:skill-projection:{}:{}", host, p.id)),
                detail: unclaimed
                    .then_some("existing skill directory without a manifest baseline".to_string()),
            });
        }
    }

    // Project skills (#2706 P2): one file per host at the host's own skill folder, body from
    // the profile; an existing folder without a recorded baseline is never claimed.
    if !request.project_skills.is_empty() {
        let (profile, _) = read_profile(&root)?;
        for ps in &request.project_skills {
            let rendered =
                render_project_skill(ps, &request.language, &profile, &project_profile_rel(&root))?;
            let rendered_digest = digest(rendered.as_bytes());
            for host in project_skill_hosts(ps) {
                let rel = project_skill_path(&host, &ps.id).ok_or("unknown skill host")?;
                let dest_dir = root
                    .join(&rel)
                    .parent()
                    .expect("skill parent")
                    .to_path_buf();
                let link = link_on_path(&root, &rel);
                let current = if link.is_some() {
                    None
                } else {
                    fs::read(root.join(&rel)).ok()
                };
                let baseline = generation_baseline(&manifest, &rel);
                let unclaimed = baseline.is_none()
                    && (current.is_some() || fs::symlink_metadata(&dest_dir).is_ok());
                let action = if link.is_some() || unclaimed {
                    "conflict"
                } else {
                    let decided = three_way_action(
                        baseline.as_deref(),
                        current.as_deref().map(digest).as_deref(),
                        &rendered_digest,
                    );
                    if decided == "merge" && is_create(current.as_ref()) {
                        "create"
                    } else {
                        decided
                    }
                };
                files.push(GenerationFilePlan {
                    path: rel,
                    action: action.into(),
                    tracked_by_default: true,
                    expected_sha256: current.as_deref().map(digest),
                    baseline_sha256: baseline,
                    rendered_sha256: Some(rendered_digest.clone()),
                    source_refs: vec![project_profile_rel(&root)],
                    kind: Some(format!("generation:project-skill:{host}:{}", ps.id)),
                    detail: match (&link, unclaimed) {
                        (Some(l), _) => Some(format!(
                            "{l} is a symbolic link; a project skill is never written through it"
                        )),
                        (None, true) => {
                            Some("existing skill directory without a manifest baseline".to_string())
                        }
                        (None, false) => None,
                    },
                });
            }
        }
    }

    // Project skills no longer requested (#2706 P2b): a file this tool wrote and nobody edited
    // is planned `remove` (apply keeps its bytes in the receipt first); an edited one is a
    // conflict and stays. A recorded file that is already gone only leaves the list.
    let requested: Vec<String> = request
        .project_skills
        .iter()
        .flat_map(|ps| {
            project_skill_hosts(ps)
                .into_iter()
                .map(move |h| format!("generation:project-skill:{h}:{}", ps.id))
        })
        .collect();
    let owned: Vec<(String, String, String)> = manifest
        .get("owned_files")
        .and_then(Value::as_array)
        .map(|list| {
            list.iter()
                .filter_map(|f| {
                    let kind = f.get("section_id")?.as_str()?;
                    if !kind.starts_with("generation:project-skill:") {
                        return None;
                    }
                    Some((
                        kind.to_string(),
                        f.get("path")?.as_str()?.to_string(),
                        f.get("baseline_sha256")?.as_str()?.to_string(),
                    ))
                })
                .collect()
        })
        .unwrap_or_default();
    for (kind, rel, baseline) in owned {
        if requested.contains(&kind) || files.iter().any(|f| f.path == rel) {
            continue;
        }
        // Pre-audit 6854760758 B1: the recorded path is trusted only when it is the path the
        // kind names, and nothing on the way is a symbolic link; otherwise it is a conflict.
        let named = kind
            .strip_prefix("generation:project-skill:")
            .and_then(|rest| rest.split_once(':'))
            .and_then(|(host, id)| project_skill_path(host, id));
        let link = link_on_path(&root, &rel);
        let refused = if named.as_deref() != Some(rel.as_str()) {
            Some("the recorded path is not the path of its skill; nothing removed".to_string())
        } else {
            link.map(|l| format!("{l} is a symbolic link; nothing removed through it"))
        };
        let current = if refused.is_some() {
            None
        } else {
            fs::read(root.join(&rel)).ok()
        };
        let edited = current.as_deref().is_some_and(|c| digest(c) != baseline);
        let detail = refused.or_else(|| {
            edited.then_some("no longer requested, but edited after generation; kept".to_string())
        });
        files.push(GenerationFilePlan {
            path: rel,
            action: if detail.is_some() {
                "conflict"
            } else {
                "remove"
            }
            .into(),
            tracked_by_default: true,
            expected_sha256: current.as_deref().map(digest),
            baseline_sha256: Some(baseline),
            rendered_sha256: None,
            source_refs: vec![project_profile_rel(&root)],
            kind: Some(kind),
            detail,
        });
    }

    let binding = digest(root.to_str().ok_or("non UTF-8 project path")?.as_bytes());
    let mut plan = GenerationPlan {
        protocol_version: 1,
        request,
        target_binding: binding,
        files,
        host_notes,
        warnings,
        plan_digest: String::new(),
    };
    plan.plan_digest = plan_digest(&plan);
    Ok(plan)
}

/// Recovers the rendered text for one generation file from the plan request.
fn rendered_text_for(root: &Path, p: &GenerationPlan, f: &GenerationFilePlan) -> Result<String> {
    let kind = f.kind.as_deref().unwrap_or("");
    if let Some(_id) = kind.strip_prefix("generation:workflow:") {
        if f.path.ends_with("SKILL.md") {
            return Ok(render_workflow_skill(
                &p.request.workflow_id,
                &p.request.language,
            ));
        }
        return Ok(render_workflow_reference(
            &p.request.workflow_id,
            &p.request.language,
        ));
    }
    if let Some(rest) = kind.strip_prefix("generation:native-role:") {
        let Some((host, role_id)) = rest.split_once(':') else {
            return Err("unknown planned path".into());
        };
        let role = p
            .request
            .native_roles
            .iter()
            .find(|r| r.role_id == role_id && r.host == host)
            .ok_or("unknown planned path")?;
        return match host {
            "claude" => Ok(render_claude_role(role)),
            _ => Ok(render_codex_role(role)),
        };
    }
    if let Some(rest) = kind.strip_prefix("generation:project-skill:") {
        let Some((host, id)) = rest.split_once(':') else {
            return Err("unknown planned path".into());
        };
        let skill = p
            .request
            .project_skills
            .iter()
            .find(|s| s.id == id && project_skill_hosts(s).iter().any(|h| h == host))
            .ok_or("unknown planned path")?;
        let (profile, _) = read_profile(root)?;
        return render_project_skill(
            skill,
            &p.request.language,
            &profile,
            &project_profile_rel(root),
        );
    }
    if let Some(rest) = kind.strip_prefix("generation:skill-projection:") {
        let Some((host, id)) = rest.split_once(':') else {
            return Err("unknown planned path".into());
        };
        let projection = p
            .request
            .skill_projections
            .iter()
            .find(|sp| sp.id == id)
            .ok_or("unknown planned path")?;
        let hosts: Vec<String> = if projection.host_targets.is_empty() {
            vec!["claude".into()]
        } else {
            projection.host_targets.clone()
        };
        if !hosts.iter().any(|h| h == host) {
            return Err("unknown planned path".into());
        }
        let source =
            source_reviewed(root, &projection.source_url)?.ok_or("unknown planned path")?;
        return render_projection_skill(projection, &p.request.language, &source);
    }
    Err("unknown planned path".into())
}

/// Removes one generated file the plan marked `remove` (#2706 P2b): only while its bytes still
/// equal the plan's, after saving them in the receipt (so `apply-recover` restores it), then the
/// folder it leaves empty. A file already gone is recorded and skipped.
fn remove_owned(
    root: &Path,
    request_id: &str,
    f: &GenerationFilePlan,
    receipt: &mut Receipt,
) -> Result<()> {
    let path_abs = root.join(&f.path);
    let before = read_opt(&path_abs)?;
    if before.as_deref().map(digest) != f.expected_sha256 {
        receipt.status = "partial".into();
        receipt.entry(
            &f.path,
            "remove",
            before.as_deref().map(digest).as_deref(),
            None,
            false,
            Some("target changed since plan".into()),
        );
        let _ = receipt.save(root);
        return Err(format!(
            "partial apply: failed at {}: target changed since plan",
            f.path
        ));
    }
    let Some(bytes) = before else {
        receipt.entry(
            &f.path,
            "remove",
            None,
            None,
            false,
            Some("already gone".into()),
        );
        return Ok(());
    };
    save_before_bytes(root, request_id, &f.path, &bytes)?;
    fs::remove_file(&path_abs).map_err(|_| format!("remove failed: {}", f.path))?;
    if let Some(dir) = path_abs.parent() {
        let _ = fs::remove_dir(dir); // only succeeds when the folder is now empty
    }
    receipt.entry(&f.path, "remove", Some(&digest(&bytes)), None, true, None);
    Ok(())
}

/// `generation-apply` (REQ-03C-03/04/11): re-plans the whole request and refuses any
/// drift, applies the agents portion through the `agents_apply` engine with its own
/// freshly computed digest, publishes the generation files with the 01-B engine from
/// its hook, and records the generation baselines plus `workflow_id`,
/// `interview_revision` and `profile_sha256` in the one `generation.json` publish that
/// comes last (#2114). A mid-sequence failure reports `partial apply` with what landed
/// and leaves the previous manifest in place.
pub fn generation_apply(
    root: &Path,
    p: GenerationPlan,
    confirmed: &str,
    permission: &str,
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    // Same canonical form at the apply entry: an old plan recorded before #2095 with
    // the raw spelling then fails the digest re-check and must re-preview.
    let mut p = p;
    p.request.workflow_reference = normalize_workflow_reference(&p.request.workflow_reference);
    let current = generation_plan(root, p.request.clone())?;
    if p.protocol_version != 1
        || confirmed != p.plan_digest
        || plan_digest(&p) != p.plan_digest
        || current.plan_digest != p.plan_digest
    {
        return Err("stale or altered plan; re-preview required".into());
    }
    if let Some(f) = p.files.iter().find(|f| f.action == "conflict") {
        return Err(format!(
            "conflict in {}: {}; nothing applied",
            f.path,
            f.detail.as_deref().unwrap_or("review required")
        ));
    }
    let root = root_of(root)?;

    // 1. The agents portion: AGENTS.md managed section, personal opt-in, .gitignore,
    //    and the single manifest publish agents.rs owns. The freshly re-planned digest is the
    //    inner confirmation; the outer generation check already refused drift.
    let agents_request = AgentsRequest {
        protocol_version: 1,
        language: p.request.language.clone(),
        workflow_reference: p.request.workflow_reference.clone(),
        local_file: p.request.local_file.clone(),
    };
    let own_output = reference_is_own_output(&root, &p.request);
    let agents_plan: AgentsPlan = agents_plan_for(&root, agents_request.clone(), own_output)?;
    // 2 and 3 run inside the agents hook, after the agents files and before the one
    //    manifest publish (#2114): a failure here leaves the previous manifest intact.
    let request_id = new_request_id("generation.apply", &p.plan_digest);
    let (profile, profile_bytes) = read_profile(&root)?;
    let mut files_created: Vec<String> = Vec::new();
    let mut files_merged: Vec<String> = Vec::new();
    // Lives past the hook: it reads `complete` only after the manifest publish (#2119).
    let mut receipt = Receipt::new(&request_id, "generation.apply", &root);
    let mut hook = |root: &Path, manifest: &mut Value| -> Result<()> {
        // 2. The generation files, through the 01-B engine with a receipt.
        for f in &p.files {
            let Some(kind) = &f.kind else {
                continue;
            };
            if !kind.starts_with("generation:") || f.path == manifest_rel(root) {
                continue;
            }
            if f.action == "unchanged" {
                continue;
            }
            if kind.starts_with("generation:project-skill:") {
                if let Some(l) = link_on_path(root, &f.path) {
                    return Err(format!(
                        "conflict in {}: {l} became a symbolic link after the plan; nothing more applied",
                        f.path
                    ));
                }
            }
            if f.action == "remove" {
                remove_owned(root, &request_id, f, &mut receipt)?;
                continue;
            }
            let rendered_text = rendered_text_for(root, &p, f)?;
            let bytes = rendered_text.into_bytes();
            let path_abs = root.join(&f.path);
            let before = read_opt(&path_abs)?;
            let fail = |receipt: &Receipt, message: String| -> String {
                match receipt.save(root) {
                    Ok(()) => message,
                    Err(se) => format!("{message}; receipt not recorded: {se}"),
                }
            };
            if before.as_deref().map(digest).as_deref() != f.expected_sha256.as_deref() {
                receipt.entry(
                    &f.path,
                    &f.action,
                    before.as_deref().map(digest).as_deref(),
                    None,
                    false,
                    Some("target changed since plan".into()),
                );
                receipt.status = "partial".into();
                return Err(fail(
                    &receipt,
                    format!(
                        "partial apply: applied [{}]; failed at {}: target changed since plan",
                        files_created.join(", "),
                        f.path
                    ),
                ));
            }
            if let Some(b) = before.as_deref() {
                if let Err(e) = save_before_bytes(root, &request_id, &f.path, b) {
                    receipt.entry(
                        &f.path,
                        &f.action,
                        before.as_deref().map(digest).as_deref(),
                        None,
                        false,
                        Some(e.clone()),
                    );
                    receipt.status = "partial".into();
                    return Err(fail(
                        &receipt,
                        format!(
                            "partial apply: applied [{}]; failed at {}: {e}",
                            files_created.join(", "),
                            f.path
                        ),
                    ));
                }
            }
            let publish = if before.is_none() {
                publish_exclusive(&path_abs, &bytes)
            } else {
                publish_replace(&path_abs, before.as_deref(), &bytes)
            };
            if let Err(e) = publish {
                receipt.entry(
                    &f.path,
                    &f.action,
                    before.as_deref().map(digest).as_deref(),
                    None,
                    false,
                    Some(e.clone()),
                );
                receipt.status = "partial".into();
                return Err(fail(
                    &receipt,
                    format!(
                        "partial apply: applied [{}]; failed at {}: {e}",
                        files_created.join(", "),
                        f.path
                    ),
                ));
            }
            receipt.entry(
                &f.path,
                &f.action,
                before.as_deref().map(digest).as_deref(),
                Some(&digest(&bytes)),
                true,
                None,
            );
            if before.is_none() {
                files_created.push(f.path.clone());
            } else {
                files_merged.push(f.path.clone());
            }
        }

        // 3. The manifest: record generation baselines and the profile linkage on the
        //    value agents_apply is about to publish. Only keys the agents.rs schema
        //    already knows are touched (agents_apply re-decodes the typed schema).
        if let Some(map) = manifest.as_object_mut() {
            let list = map
                .entry("owned_files")
                .or_insert_with(|| json!([]))
                .as_array_mut()
                .expect("owned_files array");
            list.retain(|f| {
                let path = f.get("path").and_then(Value::as_str).unwrap_or("");
                !p.files.iter().any(|pf| {
                    pf.kind
                        .as_deref()
                        .is_some_and(|k| k.starts_with("generation:"))
                        && pf.path == path
                })
            });
            for f in &p.files {
                let Some(kind) = &f.kind else {
                    continue;
                };
                if !kind.starts_with("generation:") || f.action == "conflict" {
                    continue;
                }
                // A `remove` plan carries no rendered digest, so it never re-enters the list.
                let Some(rendered) = &f.rendered_sha256 else {
                    continue;
                };
                list.push(json!({
                    "path": f.path,
                    // The bare kind (#2109) is the primary discriminator; the
                    // generation:… section_id tag stays for readers of manifests
                    // written before the kind field existed.
                    "kind": kind.strip_prefix("generation:").and_then(|rest| {
                        rest.split(':').next()
                    }),
                    "section_id": kind,
                    "baseline_sha256": rendered,
                }));
            }
            map.insert("workflow_id".into(), json!(p.request.workflow_id));
            map.insert(
                "interview_revision".into(),
                json!(profile.interview_revision),
            );
            map.insert("profile_sha256".into(), json!(digest(&profile_bytes)));
        }
        // The manifest entry itself is recorded in the agents.apply receipt, which keeps
        // the pre-apply manifest bytes for recovery. Saved as `partial` here: the
        // manifest publish is still ahead.
        receipt
            .save(root)
            .map_err(|e| format!("published files but the receipt was not recorded: {e}"))
    };
    let agents_result = agents_apply_amended_for(
        &root,
        agents_plan.clone(),
        &agents_plan.plan_digest,
        permission,
        own_output,
        &mut hook,
    )?;
    receipt.status = "complete".into();
    receipt
        .save(&root)
        .map_err(|e| format!("published files but the receipt was not recorded: {e}"))?;
    let agents_applied: Vec<String> = agents_result
        .get("applied")
        .and_then(Value::as_array)
        .into_iter()
        .flatten()
        .filter_map(|v| v.as_str().map(str::to_string))
        .filter(|path| *path != manifest_rel(&root))
        .collect();
    let mut all_created = agents_applied.clone();
    all_created.extend(files_created.iter().cloned());
    Ok(json!({
        "state": "applied",
        "files_created": all_created,
        "files_merged": files_merged,
        "host_loaded": false,
        "execution_verified": false,
        "receipt": request_id,
    }))
}
