//! AGENTS.md managed-section generation: plan (read-only) and apply (hash-rechecked).
//!
//! AGENTS.md stays user-owned. MoAI owns only the text between its managed markers and
//! decides per file with a baseline/current/new three-way comparison. Everything outside
//! the markers is preserved byte for byte. A conflict is reported, never resolved by
//! overwriting. AGENTS.local.md is created once and never modified afterwards.

use crate::apply::{self, decide, read_opt, relative_ok, root_of, target, Decision};
use crate::{digest, read_profile, Profile};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

pub const AGENTS_SECTION: &str = "moai-workflow";
pub const IGNORE_SECTION: &str = "moai-local";
/// The REVIEW.md section the PROJECT interview fills (#2639).
pub const REVIEW_SECTION: &str = "moai-review";
/// Interview answers the REVIEW.md section carries, in the order it lists them.
const REVIEW_ANSWERS: [&str; 2] = ["review_blocks", "protected"];
/// The generation manifest of `root`, under its state folder (#2725 S2b).
fn manifest_rel(root: &Path) -> String {
    crate::state_root::rel(root, "manifests/generation.json")
}
/// The manifest path of an earlier-root (`.moai-cli`) fixture project.
#[cfg(test)]
const MANIFEST: &str = crate::project_state!("manifests/generation.json");
const TEMPLATE_VERSION: &str = "rc1";

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct AgentsRequest {
    pub protocol_version: u32,
    /// `en` or `ko`: language of the rendered managed text.
    pub language: String,
    /// Repo-relative directory the managed section points at for detailed procedures.
    /// Must exist at plan time; a missing reference blocks the plan.
    pub workflow_reference: String,
    /// `none` (do not create), `ignored` (create once and add a .gitignore rule) or
    /// `tracked` (create once, no ignore rule). The personal file is opt-in by design.
    pub local_file: String,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct FilePlan {
    pub path: String,
    /// create | merge | append | unchanged | conflict
    pub action: String,
    /// Design intent for generated projects, not a measured git state (no git command runs).
    pub tracked_by_default: bool,
    pub expected_sha256: Option<String>,
    pub section_id: Option<String>,
    pub baseline_sha256: Option<String>,
    pub rendered_sha256: Option<String>,
    pub detail: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct AgentsPlan {
    pub protocol_version: u32,
    pub request: AgentsRequest,
    pub project_id: String,
    pub profile_sha256: String,
    pub target_binding: String,
    pub files: Vec<FilePlan>,
    /// Pre-existing host instruction files that may shadow or replace AGENTS.md. They are
    /// preserved, never edited; actual host loading stays not_run.
    pub host_notes: Vec<String>,
    pub warnings: Vec<String>,
    pub plan_digest: String,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
struct OwnedFile {
    path: String,
    section_id: Option<String>,
    baseline_sha256: String,
    /// What produced the entry (`workflow`, `native-role`, `skill-projection`);
    /// 03-C #2109 added it as an additive field. Manifests written before it
    /// existed (and agents.rs's own writes) leave it unset, and the
    /// `generation:…` section_id tag remains the read-compatible discriminator
    /// for those.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    kind: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
struct Manifest {
    schema_version: u32,
    generator_version: String,
    template_version: String,
    project_id: String,
    profile_sha256: String,
    interview_revision: Option<String>,
    /// Template field kept for parity; project workflow ids are not assigned yet.
    workflow_id: Option<String>,
    workflow_reference: String,
    source_kind: String,
    sources: Vec<Value>,
    canonical_sha256: Option<String>,
    projections: Vec<String>,
    owned_files: Vec<OwnedFile>,
    /// Personal files MoAI created once, or found already present at an apply. User-owned;
    /// no digest is recorded; never recreated after the user deletes them.
    #[serde(default)]
    created_once: Vec<String>,
    activation_status: String,
}

fn validate(r: &AgentsRequest) -> Result<()> {
    if r.protocol_version != 1
        || !(r.language == "en" || r.language == "ko")
        || !relative_ok(&r.workflow_reference)
        || !["none", "ignored", "tracked"].contains(&r.local_file.as_str())
    {
        return Err("invalid agents request; no values were logged".into());
    }
    Ok(())
}

pub(crate) fn markers(id: &str, comment: &str) -> (String, String) {
    match comment {
        "html" => (
            format!("<!-- moai:managed:begin id={id} -->"),
            format!("<!-- moai:managed:end id={id} -->"),
        ),
        _ => (
            format!("# moai:managed:begin id={id}"),
            format!("# moai:managed:end id={id}"),
        ),
    }
}

/// Locates exactly one managed block. Returns the body between the marker lines, or an
/// error when markers are malformed (duplicate, unterminated, or orphaned end).
fn find_block(text: &str, begin: &str, end: &str) -> Result<Option<(usize, usize, String)>> {
    let mut starts = text.match_indices(begin).map(|(i, _)| i);
    let mut ends = text.match_indices(end).map(|(i, _)| i);
    let (start, finish) = match (starts.next(), ends.next()) {
        (None, None) => return Ok(None),
        (Some(s), Some(e)) if e > s && starts.next().is_none() && ends.next().is_none() => (s, e),
        _ => return Err("malformed managed markers".into()),
    };
    let after_end = finish + end.len();
    let block_end = match text[after_end..].find('\n') {
        Some(n) => after_end + n + 1,
        None => text.len(),
    };
    let body_start = start + begin.len();
    let body = text[body_start..finish].to_owned();
    Ok(Some((start, block_end, body)))
}

/// Values rendered into a managed block must not be able to open or close a block.
fn injects_markers(p: &Profile, r: &AgentsRequest) -> bool {
    let tokens = ["<!--", "-->", "moai:managed:"];
    std::iter::once(&p.project_id)
        .chain(std::iter::once(&p.purpose))
        .chain(&p.tools)
        .chain(&p.development_methods)
        .chain(&p.hosts)
        .chain(std::iter::once(&r.workflow_reference))
        .chain(review_answers(p).into_iter().map(|(_, v)| v))
        .any(|v| tokens.iter().any(|t| v.contains(t)))
}

/// The answers REVIEW.md carries: `review_blocks` and `protected` decisions the user or the
/// repository gave. A default (an axis left unanswered, `HOLD`) is never written as a rule.
fn review_answers(p: &Profile) -> Vec<(&str, &String)> {
    REVIEW_ANSWERS
        .iter()
        .flat_map(|id| {
            p.decisions
                .iter()
                .filter(move |d| d.id == *id && d.source != "default")
                .map(|d| (d.id.as_str(), &d.value))
        })
        .collect()
}

/// REVIEW.md managed section (#2639): `None` when the interview gave no review answer, so a
/// profile without them plans no REVIEW.md at all.
/// The language the generated instruction text is written in (#2706): the profile's
/// `language.instruction_files` when it is one the templates exist for (`en`, `ko`), else the
/// request language. A profile without the choice renders exactly as before.
fn instruction_language<'a>(p: &'a Profile, r: &'a AgentsRequest) -> &'a str {
    p.language
        .as_ref()
        .and_then(|l| l.get("instruction_files"))
        .map(String::as_str)
        .filter(|code| *code == "en" || *code == "ko")
        .unwrap_or(r.language.as_str())
}

fn render_review(p: &Profile, r: &AgentsRequest) -> Option<String> {
    let answers = review_answers(p);
    if answers.is_empty() {
        return None;
    }
    let (begin, end) = markers(REVIEW_SECTION, "html");
    let ko = instruction_language(p, r) == "ko";
    let mut body = if ko {
        "## 프로젝트 리뷰 규칙 (관리 구간)\n\n`/moai project` 인터뷰 답이다. 이 구간 밖에 적은 규칙은 MoAI가 건드리지 않는다.\n\n".to_string()
    } else {
        "## Project review rules (managed section)\n\nFrom the `/moai project` interview. Rules written outside this section are never touched by MoAI.\n\n".to_string()
    };
    for (id, value) in answers {
        let line = match (id, ko) {
            ("review_blocks", false) => "Always stop in review",
            ("review_blocks", true) => "리뷰에서 항상 막는다",
            (_, false) => "Never change without asking first",
            (_, true) => "먼저 묻지 않고 바꾸지 않는다",
        };
        body.push_str(&format!("- {line}: {value}\n"));
    }
    Some(format!("{begin}\n{body}{end}\n"))
}

fn render_agents(p: &Profile, r: &AgentsRequest, root: &Path) -> String {
    let (begin, end) = markers(AGENTS_SECTION, "html");
    let list = |v: &[String]| v.join(", ");
    let lang = instruction_language(p, r);
    let share = match (lang, r.local_file.as_str()) {
        ("ko", "tracked") => "credential과 원문 비밀 로그는 repo와 Issue에 공유하지 않는다.",
        ("ko", _) => "개인 `AGENTS.local.md`, credential, 원문 비밀 로그는 repo와 Issue에 공유하지 않는다.",
        (_, "tracked") => "Never share credentials or raw secret logs in the repo or in Issues.",
        _ => "Never share the personal `AGENTS.local.md`, credentials or raw secret logs in the repo or in Issues.",
    };
    // #2293: the writing-language lines appear only when the profile sets `language`, so a
    // profile without it renders byte-identical to the earlier template.
    let writing = p
        .language
        .as_ref()
        .map(|l| {
            format!(
                "{}\n",
                crate::language::render_lines_in(
                    l,
                    lang,
                    &r.language,
                    &crate::state_root::rel(root, "project.yaml"),
                )
            )
        })
        .unwrap_or_default();
    let body = if lang == "ko" {
        format!(
            "## MoAI 작업 흐름 (관리 구간)\n\n\
프로젝트 `{id}` · 도구: {tools} · 개발 방식: {methods} · host: {hosts}.\n\
목적: {purpose}\n\n\
{writing}\
승인한 SPEC과 Issue 범위를 확인한 뒤 계획 → 구현 → 실제 검증 → 독립 감사 → 보고 순서로 진행한다.\n\
사실 확인은 스스로 하고 근거를 남긴다. 되돌릴 수 없거나, 허용 범위를 넘거나(비용·권한·외부 전송), 돈·법·외부 계약·비밀·실데이터·보호 대상 설정(권한·키·실제 HOME과 전역 설정·삭제·보호 설정)에 닿는 선택은 사용자를 기다리고, 나머지는 허용 범위 안에서 정한 뒤 기록한다.\n\
실행하지 않은 검증을 PASS로 보고하지 않는다. passed·failed·not_run·blocked·stale을 구분한다.\n\
어떤 검사를 로컬에서 돌리고 언제 정확한 head의 CI 결과를 대신 인용할 수 있는지는 VERIFY workflow의 \"Check-run rules\" 절(단일 출처)을 따른다.\n\
MoAI binary가 없어도 native 도구로 가능한 작업을 수행한다. 필수 gate의 검증된 대체가 없으면 관련 실행·발행을 막고 blocked로 보고한다.\n\
상세 절차는 `{wf}`에서 필요한 workflow만 읽는다.\n\
{share}\n",
            id = p.project_id,
            tools = list(&p.tools),
            methods = list(&p.development_methods),
            hosts = list(&p.hosts),
            purpose = p.purpose,
            wf = r.workflow_reference,
        )
    } else {
        format!(
            "## MoAI workflow (managed section)\n\n\
Project `{id}` · tools: {tools} · methods: {methods} · hosts: {hosts}.\n\
Purpose: {purpose}\n\n\
{writing}\
Confirm the approved SPEC and Issue scope, then proceed plan → implement → real verification → independent audit → report.\n\
Answer a factual question yourself and cite the evidence. A choice that cannot be undone, that goes beyond the allowed scope (cost, permissions, external transmission), or that touches money, legal terms, external contracts, secrets, real data or a guarded setting (permissions, keys, the real HOME and global settings, deletions, protection settings) waits for the user; decide any other choice within the allowed scope and record it.\n\
Never report an unexecuted check as PASS. Distinguish passed, failed, not_run, blocked and stale.\n\
Which checks run locally, and when an exact-head CI result may be cited instead, follow the \"Check-run rules\" section of the VERIFY workflow (the single source).\n\
Without the MoAI binary, do what native tools allow. When a required gate has no verified substitute, block the dependent action and report blocked.\n\
Read only the workflow you need from `{wf}`.\n\
{share}\n",
            id = p.project_id,
            tools = list(&p.tools),
            methods = list(&p.development_methods),
            hosts = list(&p.hosts),
            purpose = p.purpose,
            wf = r.workflow_reference,
        )
    };
    format!("{begin}\n{body}{end}\n")
}

fn render_local(r: &AgentsRequest) -> String {
    match (r.language.as_str(), r.local_file.as_str()) {
        ("ko", "tracked") => "# 개인 작업 보충\n\n응답 언어: 한국어\n\n<!-- MoAI가 호출 시 명시적으로 읽는 개인 파일. 이 저장소에서는 추적을 허용했으므로 비밀값을 넣지 않는다. -->\n".into(),
        ("ko", _) => "# 개인 작업 보충\n\n응답 언어: 한국어\n\n<!-- MoAI가 호출 시 명시적으로 읽는 개인 파일. 비밀값을 넣지 않으며 repo·Issue에 공유하지 않는다. -->\n".into(),
        (_, "tracked") => "# Personal supplement\n\nResponse language: English\n\n<!-- Personal file MoAI reads explicitly when invoked. Tracking is allowed in this repository, so keep secrets out of it. -->\n".into(),
        _ => "# Personal supplement\n\nResponse language: English\n\n<!-- Personal file MoAI reads explicitly when invoked. Keep secrets out; never share it in the repo or in Issues. -->\n".into(),
    }
}

fn render_ignore(root: &Path) -> String {
    let (begin, end) = markers(IGNORE_SECTION, "hash");
    // HOLD-B1 (user-confirmed 2026-11-08, ADK_1989_CONFIRMED_20261108_01): apply
    // receipts stay out of git alongside the
    // personal file; src/apply.rs::receipts_rel is the single source of the path.
    format!(
        "{begin}\nAGENTS.local.md\n{}/\n{end}\n",
        apply::receipts_rel(root)
    )
}

fn ignore_has_rule(text: &str) -> bool {
    text.lines().any(|l| {
        let t = l.trim();
        t == "AGENTS.local.md" || t == "/AGENTS.local.md"
    })
}

fn read_manifest(root: &Path, project_id: &str) -> Result<Option<Manifest>> {
    let Some(bytes) = read_opt(&target(root, &manifest_rel(root))?)? else {
        return Ok(None);
    };
    let m: Manifest = serde_json::from_slice(&bytes)
        .map_err(|_| "generation manifest unreadable; inspect before regenerating")?;
    if m.schema_version != 1 || m.project_id != project_id {
        return Err("generation manifest belongs to another schema or project".into());
    }
    Ok(Some(m))
}

fn created_once(m: Option<&Manifest>, path: &str) -> bool {
    m.is_some_and(|m| {
        m.created_once.iter().any(|p| p == path)
            // Manifests written before `created_once` existed recorded the personal file in owned_files.
            || m.owned_files.iter().any(|f| f.path == path && f.section_id.is_none())
    })
}

fn baseline_for(m: Option<&Manifest>, path: &str) -> Option<String> {
    m.and_then(|m| {
        m.owned_files
            .iter()
            .find(|f| f.path == path)
            .map(|f| f.baseline_sha256.clone())
    })
}

const BLOCK_ABSENT: &str =
    "managed block absent; user deletion preserved (re-adopt requires an explicit request)";

/// Three-way plan of one managed file. Exposed (Issue #1989, 5b) so 03-C's
/// `generation.rs` can call the engine per `FileKind` instead of reimplementing it.
pub fn section_plan(
    path: &str,
    tracked: bool,
    section: &str,
    comment: &str,
    current: Option<&[u8]>,
    baseline: Option<String>,
    rendered: &str,
) -> Result<FilePlan> {
    let (begin, end) = markers(section, comment);
    let rendered_digest = digest(rendered.as_bytes());
    let mut plan = FilePlan {
        path: path.into(),
        action: String::new(),
        tracked_by_default: tracked,
        expected_sha256: current.map(digest),
        section_id: Some(section.into()),
        baseline_sha256: baseline.clone(),
        rendered_sha256: Some(rendered_digest.clone()),
        detail: None,
    };
    let (block_digest, present) = match current {
        None => (None, false),
        Some(bytes) => {
            let text = std::str::from_utf8(bytes).map_err(|_| "target is not UTF-8")?;
            match find_block(text, &begin, &end) {
                Ok(Some((s, e, _))) => (Some(digest(&text.as_bytes()[s..e])), true),
                Ok(None) => (None, true),
                Err(_) => {
                    plan.action = "conflict".into();
                    plan.detail = Some("malformed managed markers".into());
                    return Ok(plan);
                }
            }
        }
    };
    let decision = decide(
        baseline.as_deref(),
        block_digest.as_deref(),
        Some(&rendered_digest),
    );
    plan.action = match (decision, present, block_digest.is_some()) {
        (Decision::KeepCurrent, _, _) => {
            // A block removed after a baseline stays removed; say so instead of passing silently.
            if present && block_digest.is_none() && baseline.is_some() {
                plan.detail = Some(BLOCK_ABSENT.into());
            }
            "unchanged"
        }
        (Decision::Apply, false, _) => "create",
        (Decision::Apply, true, false) => "append",
        (Decision::Apply, true, true) => "merge",
        (Decision::Conflict, _, _) => {
            plan.detail = Some(if block_digest.is_none() {
                "managed block missing or removed; baseline mismatch".into()
            } else {
                "managed block edited after baseline; preserve and review".into()
            });
            "conflict"
        }
    }
    .into();
    Ok(plan)
}

fn plan_digest(p: &AgentsPlan) -> String {
    digest(
        &serde_json::to_vec(&(
            p.protocol_version,
            &p.request,
            &p.project_id,
            &p.profile_sha256,
            &p.target_binding,
            &p.files,
            &p.host_notes,
        ))
        .expect("typed plan"),
    )
}

/// Records `bytes` as the baseline of one managed section, replacing an earlier entry.
fn record_owned(owned: &mut Vec<OwnedFile>, path: &str, section: &str, bytes: &[u8]) {
    owned.retain(|f| f.path != path);
    owned.push(OwnedFile {
        path: path.into(),
        section_id: Some(section.into()),
        baseline_sha256: digest(bytes),
        kind: None,
    });
}

/// What an apply carries over from the previous manifest: the generation linkage (#2127:
/// it belongs to the last generation apply), the owned sections, and the personal files seen
/// once (a legacy entry without a section becomes created-once).
type Carried = (Option<String>, Option<String>, Vec<OwnedFile>, Vec<String>);
fn carried(previous: Option<Manifest>) -> Carried {
    let Some(m) = previous else {
        return (None, None, Vec::new(), Vec::new());
    };
    let mut created = m.created_once;
    for legacy in m.owned_files.iter().filter(|f| f.section_id.is_none()) {
        if !created.contains(&legacy.path) {
            created.push(legacy.path.clone());
        }
    }
    let owned = m
        .owned_files
        .into_iter()
        .filter(|f| f.section_id.is_some())
        .collect();
    (m.workflow_id, m.interview_revision, owned, created)
}

/// Recovery after an interrupted apply, for an `unchanged` row whose file exists: a managed
/// block that already equals the rendered text becomes the baseline (first write or
/// interrupted update); a personal file is recorded as seen, whatever its content, so it is
/// never recreated after the user deletes it (presence, not ownership).
fn adopt_unchanged(
    owned: &mut Vec<OwnedFile>,
    created: &mut Vec<String>,
    f: &FilePlan,
    rendered: &str,
    bytes: &[u8],
) {
    match f.section_id.as_deref() {
        None => {
            if !created.contains(&f.path) {
                created.push(f.path.clone());
            }
        }
        Some(section) if f.baseline_sha256.as_deref() != Some(rendered) => {
            let comment = if f.path == ".gitignore" {
                "hash"
            } else {
                "html"
            };
            let (b, e) = markers(section, comment);
            if let Some((s, end, _)) = std::str::from_utf8(bytes)
                .ok()
                .and_then(|text| find_block(text, &b, &e).ok().flatten())
            {
                let block = &bytes[s..end];
                if digest(block) == rendered {
                    record_owned(owned, &f.path, section, block);
                }
            }
        }
        Some(_) => {}
    }
}

/// The manifest an apply publishes (before a caller's hook), from what it carried and recorded.
fn compose_manifest(
    profile: &Profile,
    profile_bytes: &[u8],
    request: &AgentsRequest,
    carried: Carried,
) -> Manifest {
    let (workflow_id, interview_revision, owned, created) = carried;
    Manifest {
        schema_version: 1,
        generator_version: env!("CARGO_PKG_VERSION").into(),
        template_version: TEMPLATE_VERSION.into(),
        project_id: profile.project_id.clone(),
        profile_sha256: digest(profile_bytes),
        interview_revision,
        workflow_id,
        workflow_reference: request.workflow_reference.clone(),
        source_kind: "generated".into(),
        sources: Vec::new(),
        canonical_sha256: None,
        projections: profile.hosts.clone(),
        owned_files: owned,
        created_once: created,
        activation_status: "pending-validation".into(),
    }
}

/// The manifest bytes as published: pretty JSON and a final newline.
fn manifest_bytes_of(m: &Manifest) -> Vec<u8> {
    let mut bytes = serde_json::to_vec_pretty(m).expect("typed manifest");
    bytes.push(b'\n');
    bytes
}

/// The manifest an apply of `files` would publish before a caller's hook, when every planned
/// file other than the manifest is `unchanged` (#2936 Q3 B). Rows of any other action are not
/// modelled, so a caller asks only in that case.
fn predict_manifest(
    root: &Path,
    profile: &Profile,
    profile_bytes: &[u8],
    previous: Option<Manifest>,
    request: &AgentsRequest,
    files: &[FilePlan],
) -> Result<Manifest> {
    let manifest_rel = manifest_rel(root);
    let mut carried = carried(previous);
    for f in files {
        if f.path == manifest_rel || f.action != "unchanged" {
            continue;
        }
        let Some(rendered) = &f.rendered_sha256 else {
            continue;
        };
        let Some(bytes) = read_opt(&target(root, &f.path)?)? else {
            continue;
        };
        adopt_unchanged(&mut carried.2, &mut carried.3, f, rendered, &bytes);
    }
    Ok(compose_manifest(profile, profile_bytes, request, carried))
}

/// [`predict_manifest`] for generation (#2936 Q3 B): the value its apply hook then amends.
pub(crate) fn predicted_manifest_value(
    root: &Path,
    request: &AgentsRequest,
    files: &[FilePlan],
) -> Result<Value> {
    let root = root_of(root)?;
    let (profile, profile_bytes) = read_profile(&root)?;
    let previous = read_manifest(&root, &profile.project_id)?;
    let m = predict_manifest(&root, &profile, &profile_bytes, previous, request, files)?;
    Ok(serde_json::to_value(&m).expect("typed manifest"))
}

/// The bytes `agents_apply` publishes for an amended manifest value; `None` when the value is
/// outside the schema (that apply would refuse it).
pub(crate) fn amended_manifest_bytes(value: Value) -> Option<Vec<u8>> {
    serde_json::from_value::<Manifest>(value)
        .ok()
        .map(|m| manifest_bytes_of(&m))
}

pub fn agents_plan(root: &Path, request: AgentsRequest) -> Result<AgentsPlan> {
    agents_plan_for(root, request, false)
}

/// `agents_plan` with one exception to the reference check (#2274, REQ-2): when
/// `reference_created_here`, the caller writes the `workflow_reference` directory in the same
/// apply (generation's own `.moai-cli/workflows/` location), so it need not exist yet. The path
/// is still resolved inside the project and every other refusal still applies (REQ-3). A
/// direct `agents-plan` always passes `false` and stays fail-closed.
pub(crate) fn agents_plan_for(
    root: &Path,
    request: AgentsRequest,
    reference_created_here: bool,
) -> Result<AgentsPlan> {
    validate(&request)?;
    let root = root_of(root)?;
    let manifest_rel = manifest_rel(&root);
    let (profile, profile_bytes) = read_profile(&root)?;
    let manifest = read_manifest(&root, &profile.project_id)?;
    let m = manifest.as_ref();
    let mut files = Vec::new();
    let mut warnings = Vec::new();
    let reference = target(&root, request.workflow_reference.trim_end_matches('/'))?;
    if !reference_created_here && !reference.is_dir() {
        return Err("workflow reference directory missing; generate it first".into());
    }
    if injects_markers(&profile, &request) {
        return Err("profile value contains managed-marker syntax; refused".into());
    }
    let host_notes: Vec<String> = [
        "CLAUDE.md",
        ".claude/CLAUDE.md",
        "CLAUDE.local.md",
        "AGENTS.override.md",
    ]
    .iter()
    .filter(|rel| root.join(rel).exists())
    .map(|rel| format!("{rel} present; preserved; may shadow or replace AGENTS.md on its host (loading not_run)"))
    .collect();

    let agents = read_opt(&target(&root, "AGENTS.md")?)?;
    files.push(section_plan(
        "AGENTS.md",
        true,
        AGENTS_SECTION,
        "html",
        agents.as_deref(),
        baseline_for(m, "AGENTS.md"),
        &render_agents(&profile, &request, &root),
    )?);

    if let Some(agents_file) = files.first() {
        let rendered_len = render_agents(&profile, &request, &root).len();
        let (b, e) = markers(AGENTS_SECTION, "html");
        let projected = match &agents {
            Some(bytes) => {
                let block_len = std::str::from_utf8(bytes)
                    .ok()
                    .and_then(|t| find_block(t, &b, &e).ok().flatten())
                    .map_or(0, |(s, end, _)| end - s);
                bytes.len() - block_len + rendered_len
            }
            None => rendered_len,
        };
        if agents_file.detail.as_deref() == Some(BLOCK_ABSENT) {
            warnings.push(BLOCK_ABSENT.into());
        }
        // #2706: the managed text follows the profile's `instruction_files` when templates
        // exist for it (`en`, `ko`); another code is reported and the request language used.
        if let Some(code) = profile
            .language
            .as_ref()
            .and_then(|l| l.get("instruction_files"))
            .filter(|code| code.as_str() != instruction_language(&profile, &request))
        {
            warnings.push(format!(
                "managed section rendered in `{}`: no managed-section text in profile language.instruction_files `{code}`",
                instruction_language(&profile, &request)
            ));
        }
        if agents_file.action != "conflict" && projected > 32 * 1024 {
            warnings
                .push("AGENTS.md may exceed the 32 KiB default instruction budget of Codex".into());
        }
    }

    if let Some(review) = render_review(&profile, &request) {
        let current = read_opt(&target(&root, "REVIEW.md")?)?;
        files.push(section_plan(
            "REVIEW.md",
            true,
            REVIEW_SECTION,
            "html",
            current.as_deref(),
            baseline_for(m, "REVIEW.md"),
            &review,
        )?);
    }

    if request.local_file != "none" {
        let local = read_opt(&target(&root, "AGENTS.local.md")?)?;
        let local_rendered = render_local(&request);
        files.push(FilePlan {
            path: "AGENTS.local.md".into(),
            action: if local.is_none() && !created_once(m, "AGENTS.local.md") {
                "create"
            } else {
                "unchanged"
            }
            .into(),
            tracked_by_default: request.local_file == "tracked",
            expected_sha256: local.as_deref().map(digest),
            section_id: None,
            baseline_sha256: None,
            rendered_sha256: Some(digest(local_rendered.as_bytes())),
            detail: None,
        });
    }

    if request.local_file == "ignored" {
        let ignore = read_opt(&target(&root, ".gitignore")?)?;
        let ignore_text = match &ignore {
            Some(b) => Some(std::str::from_utf8(b).map_err(|_| "target is not UTF-8")?),
            None => None,
        };
        if ignore_text.is_some_and(ignore_has_rule) && baseline_for(m, ".gitignore").is_none() {
            files.push(FilePlan {
                path: ".gitignore".into(),
                action: "unchanged".into(),
                tracked_by_default: true,
                expected_sha256: ignore.as_deref().map(digest),
                section_id: Some(IGNORE_SECTION.into()),
                baseline_sha256: None,
                rendered_sha256: None,
                detail: Some("existing user rule names AGENTS.local.md; effective git ignore status not verified".into()),
            });
        } else {
            files.push(section_plan(
                ".gitignore",
                true,
                IGNORE_SECTION,
                "hash",
                ignore.as_deref(),
                baseline_for(m, ".gitignore"),
                &render_ignore(&root),
            )?);
        }
    }

    let manifest_bytes = read_opt(&target(&root, &manifest_rel)?)?;
    // #2936 Q3 B: the manifest is `unchanged` when every other file is and the bytes an apply
    // would publish equal the file on disk, so a re-run plans `unchanged` throughout.
    let manifest_action = match &manifest_bytes {
        None => "create",
        Some(disk)
            if files.iter().all(|f| f.action == "unchanged")
                && manifest_bytes_of(&predict_manifest(
                    &root,
                    &profile,
                    &profile_bytes,
                    manifest.clone(),
                    &request,
                    &files,
                )?) == *disk =>
        {
            "unchanged"
        }
        Some(_) => "merge",
    };
    files.push(FilePlan {
        path: manifest_rel.clone(),
        action: manifest_action.into(),
        tracked_by_default: true,
        expected_sha256: manifest_bytes.as_deref().map(digest),
        section_id: None,
        baseline_sha256: None,
        rendered_sha256: None,
        detail: None,
    });

    let binding = digest(root.to_str().ok_or("non UTF-8 project path")?.as_bytes());
    let mut plan = AgentsPlan {
        protocol_version: 1,
        request,
        project_id: profile.project_id.clone(),
        profile_sha256: digest(&profile_bytes),
        target_binding: binding,
        files,
        host_notes,
        warnings,
        plan_digest: String::new(),
    };
    plan.plan_digest = plan_digest(&plan);
    Ok(plan)
}

/// Compatibility publish used by 02-B's `claude.rs`: exclusive create when no before
/// bytes exist, atomic replace otherwise. The body lives in the shared engine
/// (`crate::apply`, Issue #1989, 01-B owner of the agents.rs refactor); the pre-check
/// keeps the historical "changed since plan" message for existing callers.
pub(crate) fn write_replace(path: &Path, before: Option<&[u8]>, after: &[u8]) -> Result<()> {
    if read_opt(path)?.as_deref() != before {
        return Err("target changed since plan; re-preview required".into());
    }
    if before.is_none() {
        apply::publish_exclusive(path, after)
    } else {
        apply::publish_replace(path, before, after)
    }
}

/// The written file must contain exactly one managed block whose bytes equal the rendering.
pub fn assert_single_block(out: &[u8], begin: &str, end: &str, rendered: &str) -> Result<()> {
    let text = std::str::from_utf8(out).map_err(|_| "rendered output is not UTF-8")?;
    match find_block(text, begin, end)? {
        Some((s, e, _)) if &text[s..e] == rendered => Ok(()),
        _ => Err("rendered block not uniquely locatable; refused".into()),
    }
}

/// Splices the rendered block into `current` (create prefix / in-block replace / append).
/// Exposed (Issue #1989, 5b) for 03-C's generation renderer.
pub fn splice(
    current: Option<&[u8]>,
    begin: &str,
    end: &str,
    rendered: &str,
    create_prefix: &str,
) -> Result<Vec<u8>> {
    Ok(match current {
        None => format!("{create_prefix}{rendered}").into_bytes(),
        Some(bytes) => {
            let text = std::str::from_utf8(bytes).map_err(|_| "target is not UTF-8")?;
            match find_block(text, begin, end)? {
                Some((s, e, _)) => format!("{}{}{}", &text[..s], rendered, &text[e..]).into_bytes(),
                None => {
                    let sep = if text.is_empty() || text.ends_with("\n\n") {
                        ""
                    } else if text.ends_with('\n') {
                        "\n"
                    } else {
                        "\n\n"
                    };
                    format!("{text}{sep}{rendered}").into_bytes()
                }
            }
        }
    })
}

pub fn agents_apply(
    root: &Path,
    p: AgentsPlan,
    confirmed: &str,
    permission: &str,
) -> Result<Value> {
    agents_apply_amended(root, p, confirmed, permission, &mut |_, _| Ok(()))
}

/// `agents_apply` with one caller hook between the agents file publishes and the
/// single manifest publish (Issue #2114). `amend` may publish its own files and edit
/// the manifest value; the result must still decode as the manifest schema. An `amend`
/// error returns before the manifest is touched, so the previous manifest — with what
/// an earlier apply recorded — stays on disk instead of a reset intermediate.
// @MX:NOTE [#2114] generation_apply's only path to the manifest; one publish per apply.
pub(crate) fn agents_apply_amended(
    root: &Path,
    p: AgentsPlan,
    confirmed: &str,
    permission: &str,
    amend: &mut dyn FnMut(&Path, &mut Value) -> Result<()>,
) -> Result<Value> {
    agents_apply_amended_for(root, p, confirmed, permission, false, amend)
}

/// `agents_apply_amended` whose re-plan uses `agents_plan_for`: generation passes `true`
/// only when its own `amend` hook writes the `workflow_reference` directory (#2274).
pub(crate) fn agents_apply_amended_for(
    root: &Path,
    p: AgentsPlan,
    confirmed: &str,
    permission: &str,
    reference_created_here: bool,
    amend: &mut dyn FnMut(&Path, &mut Value) -> Result<()>,
) -> Result<Value> {
    // Caller authorization remains host-owned; this reference is not an authorization verifier.
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    let current = agents_plan_for(root, p.request.clone(), reference_created_here)?;
    if p.protocol_version != 1
        || confirmed != p.plan_digest
        || plan_digest(&p) != p.plan_digest
        || current.plan_digest != p.plan_digest
    {
        return Err("stale or altered plan; re-preview required".into());
    }
    if let Some(c) = p.files.iter().find(|f| f.action == "conflict") {
        // REQ-01B-01: a conflict writes no target byte; the refused outcome is recorded
        // so 04-C recovery tooling can see why nothing was applied.
        let request_id = apply::new_request_id("agents.apply", &p.plan_digest);
        let mut receipt = apply::Receipt::new(&request_id, "agents.apply", root);
        for f in p.files.iter().filter(|f| f.action == "conflict") {
            receipt.entry(
                &f.path,
                "conflict",
                f.expected_sha256.as_deref(),
                f.rendered_sha256.as_deref(),
                false,
                f.detail.clone(),
            );
        }
        receipt.status = "refused".into();
        let mut message = format!(
            "conflict in {}: {}; nothing applied",
            c.path,
            c.detail.as_deref().unwrap_or("review required")
        );
        if let Err(e) = receipt.save(root) {
            message.push_str(&format!("; receipt not recorded: {e}"));
        }
        return Err(message);
    }
    let root = root_of(root)?;
    let manifest_rel = manifest_rel(&root);
    let request_id = apply::new_request_id("agents.apply", &p.plan_digest);
    let mut receipt = apply::Receipt::new(&request_id, "agents.apply", &root);
    let (profile, profile_bytes) = read_profile(&root)?;
    let previous = read_manifest(&root, &profile.project_id)?;
    // #2127: the generation linkage belongs to the last generation apply; an
    // agents-only apply carries it forward instead of resetting it.
    let (workflow_id, interview_revision, mut owned, mut created) = carried(previous);
    let mut applied = Vec::new();
    let step = |applied: &[String], path: &str, e: String| {
        format!(
            "partial apply: applied [{}]; failed at {path}: {e}",
            applied.join(", ")
        )
    };
    // Saves the receipt at a failure point and appends a note when the receipt itself
    // could not be recorded — the receipt gap is never silent (REQ-01B-06).
    let receipt_failure =
        |receipt: &apply::Receipt, root: &Path, message: &str| match receipt.save(root) {
            Ok(()) => message.to_string(),
            Err(e) => format!("{message}; receipt not recorded: {e}"),
        };

    for f in &p.files {
        if f.path == manifest_rel {
            continue;
        }
        if f.action == "unchanged" {
            // Recovery after an interrupted apply: a managed block that already equals the
            // rendered text becomes the baseline (first write or interrupted update); a
            // personal file that equals its template is recorded as created once.
            let Some(rendered) = &f.rendered_sha256 else {
                continue;
            };
            let path = target(&root, &f.path)?;
            let Some(bytes) = read_opt(&path)? else {
                continue;
            };
            receipt.entry(
                &f.path,
                "unchanged",
                Some(&digest(&bytes)),
                Some(&digest(&bytes)),
                false,
                None,
            );
            adopt_unchanged(&mut owned, &mut created, f, rendered, &bytes);
            continue;
        }
        let path = target(&root, &f.path)?;
        let before = read_opt(&path)?;
        if before.as_deref().map(digest) != f.expected_sha256 {
            // REQ-01B-02: the file and everything after it stays untouched; the
            // stale_target outcome goes into the receipt.
            receipt.entry(
                &f.path,
                "stale_target",
                before.as_deref().map(digest).as_deref(),
                None,
                false,
                Some("target changed since plan".into()),
            );
            return Err(receipt_failure(
                &receipt,
                &root,
                &step(&applied, &f.path, "target changed since plan".into()),
            ));
        }
        let (bytes, section) = match f.path.as_str() {
            "AGENTS.md" => {
                let (b, e) = markers(AGENTS_SECTION, "html");
                let rendered = render_agents(&profile, &p.request, &root);
                let prefix = format!("# {}\n\n", profile.project_id);
                let out = splice(before.as_deref(), &b, &e, &rendered, &prefix)?;
                assert_single_block(&out, &b, &e, &rendered)?;
                (out, Some(AGENTS_SECTION))
            }
            "REVIEW.md" => {
                let (b, e) = markers(REVIEW_SECTION, "html");
                let rendered =
                    render_review(&profile, &p.request).ok_or("review section has no answers")?;
                let out = splice(
                    before.as_deref(),
                    &b,
                    &e,
                    &rendered,
                    "# Review instructions\n\n",
                )?;
                assert_single_block(&out, &b, &e, &rendered)?;
                (out, Some(REVIEW_SECTION))
            }
            ".gitignore" => {
                let (b, e) = markers(IGNORE_SECTION, "hash");
                let rendered = render_ignore(&root);
                let out = splice(before.as_deref(), &b, &e, &rendered, "")?;
                assert_single_block(&out, &b, &e, &rendered)?;
                (out, Some(IGNORE_SECTION))
            }
            "AGENTS.local.md" => (render_local(&p.request).into_bytes(), None),
            _ => return Err("unknown planned path".into()),
        };
        // REQ-01B-08: keep the before bytes before publishing so recovery can restore
        // content; a created file needs no backup (recovery removes it).
        if let Some(b) = before.as_deref() {
            let backup = apply::save_before_bytes(&root, &request_id, &f.path, b);
            #[cfg(test)]
            let backup = if FAIL_FILE_BACKUP.with(std::cell::Cell::get) {
                Err("injected file backup failure".to_string())
            } else {
                backup
            };
            if let Err(e) = backup {
                // #2125: a failed file backup is recorded like the #2123 manifest
                // path, so the receipt still names the file that was not published.
                receipt
                    .entries
                    .push(failed_backup_entry(&f.path, &f.action, b, &e));
                return Err(receipt_failure(
                    &receipt,
                    &root,
                    &step(&applied, &f.path, e),
                ));
            }
        }
        let publish = if before.is_none() {
            apply::publish_exclusive(&path, &bytes)
        } else {
            apply::publish_replace(&path, before.as_deref(), &bytes)
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
            return Err(receipt_failure(
                &receipt,
                &root,
                &step(&applied, &f.path, e),
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
        if let Err(e) = receipt.save(&root) {
            return Err(format!(
                "partial apply: applied [{}]; published {} but its receipt was not recorded: {e}",
                applied.join(", "),
                f.path
            ));
        }
        // Baseline is the managed text MoAI wrote, not the whole user file.
        match section {
            Some(AGENTS_SECTION) => record_owned(
                &mut owned,
                &f.path,
                AGENTS_SECTION,
                render_agents(&profile, &p.request, &root).as_bytes(),
            ),
            Some(IGNORE_SECTION) => record_owned(
                &mut owned,
                &f.path,
                IGNORE_SECTION,
                render_ignore(&root).as_bytes(),
            ),
            _ => {
                if !created.contains(&f.path) {
                    created.push(f.path.clone());
                }
            }
        }
        applied.push(f.path.clone());
    }

    let manifest = compose_manifest(
        &profile,
        &profile_bytes,
        &p.request,
        (workflow_id, interview_revision, owned, created),
    );
    let manifest_path = target(&root, &manifest_rel)?;
    let before = read_opt(&manifest_path)?;
    let manifest_plan = p
        .files
        .iter()
        .find(|f| f.path == manifest_rel)
        .ok_or("manifest missing from plan")?;
    if before.as_deref().map(digest) != manifest_plan.expected_sha256 {
        receipt.entry(
            &manifest_rel,
            "stale_target",
            before.as_deref().map(digest).as_deref(),
            None,
            false,
            Some("target changed since plan".into()),
        );
        return Err(receipt_failure(
            &receipt,
            &root,
            &step(&applied, &manifest_rel, "target changed since plan".into()),
        ));
    }
    let mut value = serde_json::to_value(&manifest).expect("typed manifest");
    if let Err(e) = amend(&root, &mut value) {
        return Err(receipt_failure(&receipt, &root, &e));
    }
    let manifest: Manifest = match serde_json::from_value(value) {
        Ok(m) => m,
        Err(_) => {
            // #2119: the files the hook and this apply published stay on record; the
            // refused manifest is recorded like any other failed target.
            let e = "amended manifest outside the schema".to_string();
            receipt.entry(
                &manifest_rel,
                &manifest_plan.action,
                before.as_deref().map(digest).as_deref(),
                None,
                false,
                Some(e.clone()),
            );
            return Err(receipt_failure(
                &receipt,
                &root,
                &step(&applied, &manifest_rel, e),
            ));
        }
    };
    let manifest_bytes = manifest_bytes_of(&manifest);
    if before.as_deref() != Some(manifest_bytes.as_slice()) {
        if let Some(b) = before.as_deref() {
            if let Err(e) = apply::save_before_bytes(&root, &request_id, &manifest_rel, b) {
                // #2123: a failed backup is recorded like the #2119 decode failure,
                // so the receipt still names the manifest that was not published.
                receipt.entry(
                    &manifest_rel,
                    &manifest_plan.action,
                    Some(&digest(b)),
                    None,
                    false,
                    Some(e.clone()),
                );
                return Err(receipt_failure(
                    &receipt,
                    &root,
                    &step(&applied, &manifest_rel, e),
                ));
            }
        }
        let publish = if before.is_none() {
            apply::publish_exclusive(&manifest_path, &manifest_bytes)
        } else {
            apply::publish_replace(&manifest_path, before.as_deref(), &manifest_bytes)
        };
        if let Err(e) = publish {
            receipt.entry(
                &manifest_rel,
                &manifest_plan.action,
                before.as_deref().map(digest).as_deref(),
                None,
                false,
                Some(e.clone()),
            );
            return Err(receipt_failure(
                &receipt,
                &root,
                &step(&applied, &manifest_rel, e),
            ));
        }
        receipt.entry(
            &manifest_rel,
            &manifest_plan.action,
            before.as_deref().map(digest).as_deref(),
            Some(&digest(&manifest_bytes)),
            true,
            None,
        );
        if let Err(e) = receipt.save(&root) {
            return Err(format!(
                "partial apply: applied [{}]; published {manifest_rel} but its receipt was not recorded: {e}",
                applied.join(", ")
            ));
        }
        applied.push(manifest_rel.clone());
    }
    receipt.status = "complete".into();
    receipt.save(&root).map_err(|e| {
        format!(
            "applied [{}]; receipt not recorded: {e}",
            applied.join(", ")
        )
    })?;
    Ok(json!({
        "state": if applied.is_empty() { "unchanged" } else { "applied" },
        "receipt": request_id,
        "applied": applied,
        "files_created": !applied.is_empty(),
        "manifest_sha256": digest(&manifest_bytes),
        "host_loaded": false,
        "execution_verified": false,
        "host_notes": p.host_notes
    }))
}

/// The receipt row for a failed before-bytes backup (#2125): the file is named
/// with `published=false`, the error, and its before digest, so recovery tooling
/// sees the file that never reached its backup.
fn failed_backup_entry(
    path: &str,
    action: &str,
    before: &[u8],
    error: &str,
) -> apply::ReceiptEntry {
    apply::ReceiptEntry {
        path: path.into(),
        decision: action.into(),
        before_sha256: Some(digest(before)),
        after_sha256: None,
        published: false,
        error: Some(error.to_string()),
    }
}

// #2125: test-only switch — set before an apply to make the next file-loop
// backup fail, so the receipt row can be observed from the receipt file.
#[cfg(test)]
thread_local! {
    static FAIL_FILE_BACKUP: std::cell::Cell<bool> = const { std::cell::Cell::new(false) };
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;

    /// #2123: the manifest's before-bytes backup fails (its backup directory is
    /// blocked by a regular file, planted by the hook after this run created its
    /// `<request-id>.before/`). The apply fails and the agents.apply receipt records
    /// the manifest row as not published, with the error.
    #[test]
    fn manifest_backup_failure_is_recorded_in_receipt() {
        let root = tempfile::TempDir::new().unwrap();
        fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
        fs::create_dir_all(root.path().join(crate::PROJECT_STATE_DIR)).unwrap();
        let profile = json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"],"interview_revision":"rc1-fixture-1"});
        fs::write(
            root.path().join(crate::project_state!("project.yaml")),
            serde_json::to_vec_pretty(&profile).unwrap(),
        )
        .unwrap();
        let request = |language: &str| AgentsRequest {
            protocol_version: 1,
            language: language.into(),
            workflow_reference: "skills/moai/workflows/".into(),
            local_file: "tracked".into(),
        };
        let first = agents_plan(root.path(), request("en")).unwrap();
        let digest = first.plan_digest.clone();
        agents_apply(root.path(), first, &digest, "fixture").unwrap();
        let receipts = root.path().join(apply::receipts_rel(root.path()));
        let backups = |dir: &Path| -> Vec<std::path::PathBuf> {
            fs::read_dir(dir)
                .unwrap()
                .map(|e| e.unwrap().path())
                .filter(|p| p.extension().and_then(|e| e.to_str()) == Some("before"))
                .collect()
        };
        assert!(
            backups(&receipts).is_empty(),
            "the first apply only created"
        );

        // en -> ko merges AGENTS.md (its before bytes create this run's backup dir)
        // and changes the manifest, so the manifest backup is attempted.
        let second = agents_plan(root.path(), request("ko")).unwrap();
        let digest = second.plan_digest.clone();
        let err = agents_apply_amended(root.path(), second, &digest, "fixture", &mut |_, _| {
            let dirs = backups(&receipts);
            assert_eq!(dirs.len(), 1, "this run's backup dir: {dirs:?}");
            fs::write(
                dirs[0].join(crate::PROJECT_STATE_DIR),
                b"blocks the manifest backup",
            )
            .unwrap();
            Ok(())
        })
        .unwrap_err();
        assert!(
            err.contains("partial apply") && err.contains(MANIFEST),
            "{err}"
        );

        let mut manifest_rows = Vec::new();
        for entry in fs::read_dir(&receipts).unwrap() {
            let path = entry.unwrap().path();
            if path.extension().and_then(|e| e.to_str()) != Some("json") {
                continue;
            }
            let v: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
            if v["operation"] != json!("agents.apply") || v["status"] == json!("complete") {
                continue;
            }
            assert_eq!(v["status"], json!("partial"), "{v}");
            manifest_rows.extend(
                v["entries"]
                    .as_array()
                    .unwrap()
                    .iter()
                    .filter(|e| e["path"] == json!(MANIFEST))
                    .cloned(),
            );
        }
        assert_eq!(manifest_rows.len(), 1, "{manifest_rows:?}");
        assert_eq!(manifest_rows[0]["published"], json!(false));
        assert!(
            manifest_rows[0]["error"]
                .as_str()
                .is_some_and(|e| !e.is_empty()),
            "{manifest_rows:?}"
        );
    }

    /// #2119 path 1: a hook that succeeds but leaves the manifest outside the schema
    /// fails the apply, and the agents.apply receipt records that manifest failure.
    /// #2125 RED: a FILE backup failure in the loop returns through the bare `?`
    /// without a receipt row and without the receipt-gap note (REQ-01B-06: the
    /// receipt gap is never silent). The injection makes the receipts directory
    /// unwritable, so the AGENTS.md merge in a language change cannot store its
    /// before bytes; the fixed path records the file row (published=false, the
    /// error, the before digest) and calls receipt_failure, whose save attempt
    /// names the gap even though the directory refuses the file.
    #[cfg(unix)]
    #[test]
    fn file_backup_failure_names_the_receipt_gap() {
        use std::os::unix::fs::PermissionsExt;
        let root = tempfile::TempDir::new().unwrap();
        fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
        fs::create_dir_all(root.path().join(crate::PROJECT_STATE_DIR)).unwrap();
        let profile = json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"],"interview_revision":"rc1-fixture-1"});
        fs::write(
            root.path().join(crate::project_state!("project.yaml")),
            serde_json::to_vec_pretty(&profile).unwrap(),
        )
        .unwrap();
        let first = agents_plan(
            root.path(),
            AgentsRequest {
                protocol_version: 1,
                language: "en".into(),
                workflow_reference: "skills/moai/workflows/".into(),
                local_file: "tracked".into(),
            },
        )
        .unwrap();
        agents_apply_amended(
            root.path(),
            first,
            &agents_plan(
                root.path(),
                AgentsRequest {
                    protocol_version: 1,
                    language: "en".into(),
                    workflow_reference: "skills/moai/workflows/".into(),
                    local_file: "tracked".into(),
                },
            )
            .unwrap()
            .plan_digest,
            "fixture",
            &mut |_, _| Ok(()),
        )
        .unwrap();
        let receipts = root.path().join(apply::receipts_rel(root.path()));
        fs::set_permissions(&receipts, fs::Permissions::from_mode(0o555)).unwrap();
        let second = agents_plan(
            root.path(),
            AgentsRequest {
                protocol_version: 1,
                language: "ko".into(),
                workflow_reference: "skills/moai/workflows/".into(),
                local_file: "tracked".into(),
            },
        )
        .unwrap();
        let digest = second.plan_digest.clone();
        let err = agents_apply_amended(root.path(), second, &digest, "fixture", &mut |_, _| Ok(()))
            .unwrap_err();
        fs::set_permissions(&receipts, fs::Permissions::from_mode(0o755)).unwrap();
        assert!(
            err.contains("partial apply") && err.contains("AGENTS.md"),
            "{err}"
        );
        // The fix names the receipt gap (receipt_failure ran); the buggy path
        // returns the bare step message without it.
        assert!(err.contains("receipt not recorded"), "{err}");
    }

    /// #2125 (audit round 2): the receipt FILE records the failed file row —
    /// path=AGENTS.md, published=false, the error, the before digest — read from
    /// `receipts/*.json` after an apply whose file backup was injected to fail
    /// (#2122's amended_manifest_decode_failure style, through the #2125 test
    /// switch). Deleting the receipt.entries push breaks this assertion.
    #[test]
    fn file_backup_failure_row_is_recorded_in_the_receipt_file() {
        let root = tempfile::TempDir::new().unwrap();
        fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
        fs::create_dir_all(root.path().join(crate::PROJECT_STATE_DIR)).unwrap();
        let profile = json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"],"interview_revision":"rc1-fixture-1"});
        fs::write(
            root.path().join(crate::project_state!("project.yaml")),
            serde_json::to_vec_pretty(&profile).unwrap(),
        )
        .unwrap();
        let first = agents_plan(
            root.path(),
            AgentsRequest {
                protocol_version: 1,
                language: "en".into(),
                workflow_reference: "skills/moai/workflows/".into(),
                local_file: "tracked".into(),
            },
        )
        .unwrap();
        agents_apply_amended(
            root.path(),
            first,
            &agents_plan(
                root.path(),
                AgentsRequest {
                    protocol_version: 1,
                    language: "en".into(),
                    workflow_reference: "skills/moai/workflows/".into(),
                    local_file: "tracked".into(),
                },
            )
            .unwrap()
            .plan_digest,
            "fixture",
            &mut |_, _| Ok(()),
        )
        .unwrap();
        let before_bytes = fs::read(root.path().join("AGENTS.md")).unwrap();
        let expected_before = digest(&before_bytes);

        FAIL_FILE_BACKUP.with(|cell| cell.set(true));
        let second = agents_plan(
            root.path(),
            AgentsRequest {
                protocol_version: 1,
                language: "ko".into(),
                workflow_reference: "skills/moai/workflows/".into(),
                local_file: "tracked".into(),
            },
        )
        .unwrap();
        let digest = second.plan_digest.clone();
        let err = agents_apply_amended(root.path(), second, &digest, "fixture", &mut |_, _| Ok(()))
            .unwrap_err();
        FAIL_FILE_BACKUP.with(|cell| cell.set(false));
        assert!(
            err.contains("partial apply") && err.contains("AGENTS.md"),
            "{err}"
        );

        // The written agents.apply receipt names the failed file row.
        let mut rows = Vec::new();
        for entry in fs::read_dir(root.path().join(apply::receipts_rel(root.path()))).unwrap() {
            let path = entry.unwrap().path();
            if path.extension().and_then(|e| e.to_str()) != Some("json") {
                continue;
            }
            let v: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
            if v["operation"] != json!("agents.apply") || v["status"] != json!("partial") {
                continue;
            }
            for e in v["entries"].as_array().unwrap() {
                if e["path"] == json!("AGENTS.md") {
                    rows.push((v["status"].clone(), e.clone()));
                }
            }
        }
        assert_eq!(rows.len(), 1, "{rows:?}");
        assert_eq!(rows[0].0, json!("partial"));
        assert_eq!(rows[0].1["published"], json!(false));
        assert_eq!(rows[0].1["decision"], json!("merge"));
        assert!(
            rows[0].1["error"].as_str().is_some_and(|e| !e.is_empty()),
            "{rows:?}"
        );
        assert_eq!(
            rows[0].1["before_sha256"].as_str(),
            Some(expected_before.as_str())
        );
        assert_eq!(rows[0].1["after_sha256"], Value::Null);
    }

    #[test]
    fn amended_manifest_decode_failure_is_recorded_in_receipt() {
        let root = tempfile::TempDir::new().unwrap();
        fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
        fs::create_dir_all(root.path().join(crate::PROJECT_STATE_DIR)).unwrap();
        let profile = json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"],"interview_revision":"rc1-fixture-1"});
        fs::write(
            root.path().join(crate::project_state!("project.yaml")),
            serde_json::to_vec_pretty(&profile).unwrap(),
        )
        .unwrap();
        let request = AgentsRequest {
            protocol_version: 1,
            language: "en".into(),
            workflow_reference: "skills/moai/workflows/".into(),
            local_file: "tracked".into(),
        };
        let p = agents_plan(root.path(), request).unwrap();
        let digest = p.plan_digest.clone();
        let err = agents_apply_amended(root.path(), p, &digest, "fixture", &mut |_, v| {
            v["outside_the_schema"] = json!(true);
            Ok(())
        })
        .unwrap_err();
        assert!(err.contains("amended manifest outside the schema"), "{err}");
        assert!(
            !root.path().join(MANIFEST).exists(),
            "no manifest byte moved"
        );

        let mut manifest_rows = Vec::new();
        for entry in fs::read_dir(root.path().join(apply::receipts_rel(root.path()))).unwrap() {
            let path = entry.unwrap().path();
            if path.extension().and_then(|e| e.to_str()) != Some("json") {
                continue;
            }
            let v: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
            if v["operation"] != json!("agents.apply") {
                continue;
            }
            assert_eq!(v["status"], json!("partial"), "{v}");
            manifest_rows.extend(
                v["entries"]
                    .as_array()
                    .unwrap()
                    .iter()
                    .filter(|e| e["path"] == json!(MANIFEST))
                    .cloned(),
            );
        }
        assert_eq!(manifest_rows.len(), 1, "{manifest_rows:?}");
        assert_eq!(manifest_rows[0]["published"], json!(false));
        assert!(
            manifest_rows[0]["error"]
                .as_str()
                .is_some_and(|e| e.contains("outside the schema")),
            "{manifest_rows:?}"
        );
    }

    /// #2125: the row recorded for a failed before-bytes backup names the file,
    /// stays unpublished, carries the error, and pins the before digest — deleting
    /// any of those from the row breaks this assertion.
    #[test]
    fn failed_backup_entry_records_the_gap_row() {
        let before = b"AGENTS.md before bytes";
        let row = failed_backup_entry(
            "AGENTS.md",
            "merge",
            before,
            "receipt backup directory unavailable",
        );
        assert_eq!(row.path, "AGENTS.md");
        assert_eq!(row.decision, "merge");
        assert!(!row.published);
        let expected_before = digest(before);
        assert_eq!(row.before_sha256.as_deref(), Some(expected_before.as_str()));
        assert_eq!(row.after_sha256, None);
        assert!(row.error.as_deref().is_some_and(|e| !e.is_empty()));
    }

    #[test]
    fn owned_files_without_kind_still_deserialize() {
        // #2109 AC-2: manifests written before the additive `kind` field existed
        // (agents.rs and 03-C writes) read unchanged; an entry carrying kind reads
        // too, and both shapes round-trip through the manifest type.
        let old = json!({
            "path": "AGENTS.md",
            "section_id": "moai-workflow",
            "baseline_sha256": "sha256:971e386609930516d7c92f3ad25ef05793f923f978c6b01a11f5a018e94f5b0b"
        });
        let parsed: OwnedFile = serde_json::from_value(old).unwrap();
        assert_eq!(parsed.kind, None);
        assert_eq!(parsed.section_id.as_deref(), Some("moai-workflow"));
        let new = json!({
            "path": crate::project_state!("workflows/review-flow/SKILL.md"),
            "kind": "workflow",
            "section_id": "generation:workflow:review-flow",
            "baseline_sha256": "sha256:971e386609930516d7c92f3ad25ef05793f923f978c6b01a11f5a018e94f5b0b"
        });
        let parsed: OwnedFile = serde_json::from_value(new).unwrap();
        assert_eq!(parsed.kind.as_deref(), Some("workflow"));
    }

    #[test]
    fn block_detection_and_malformed_markers() {
        let (b, e) = markers("x", "html");
        let text = format!("user\n\n{b}\nbody\n{e}\nmore\n");
        let (s, end, body) = find_block(&text, &b, &e).unwrap().unwrap();
        assert_eq!(&text[s..end], format!("{b}\nbody\n{e}\n"));
        assert_eq!(body, "\nbody\n");
        assert!(find_block(&format!("{b}\n{b}\n{e}\n"), &b, &e).is_err());
        assert!(find_block(&format!("{b}\nbody\n"), &b, &e).is_err());
        assert!(find_block(&format!("{e}\n{b}\n"), &b, &e).is_err());
        assert!(find_block("plain", &b, &e).unwrap().is_none());
    }

    // relative_paths_are_bounded moved to src/apply.rs with the shared helper.
}
