//! Project instruction files: plan (read-only) and apply (hash-rechecked).
//!
//! `CLAUDE.md` stays user-owned. MoAI owns only the text between its managed markers and
//! decides per file with a baseline/current/new three-way comparison. Everything outside
//! the markers is preserved byte for byte. A conflict is reported, never resolved by
//! overwriting. `CLAUDE.local.md` is created once, on the user's pick, and never modified
//! afterwards. Nothing is written under the project's `.claude` folders: moai's common
//! content lives in the Claude config folder. `AGENTS.md` and `AGENTS.local.md` are never
//! written.

use crate::apply::{self, decide, read_opt, relative_ok, root_of, target, Decision};
use crate::retired_hosts::owned_entry_retired;
use crate::state_root;
use crate::{digest, read_profile, Profile};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

pub const AGENTS_SECTION: &str = "moai-workflow";
pub const IGNORE_SECTION: &str = "moai-local";
/// The project file that holds the managed block (Claude Code's own file name).
pub const BLOCK_FILE: &str = "CLAUDE.md";
/// The personal file, created once and only on the user's pick.
pub const LOCAL_FILE: &str = "CLAUDE.local.md";
/// What the file `AGENTS.md` is called; it is read, never written.
const AGENTS_FILE: &str = "AGENTS.md";
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
    /// `none` (no personal file), `ignored` (the personal file is kept out of git) or `tracked`
    /// (read as `ignored`: a personal file is not shared; the preview says so). The personal
    /// file is created only on the pick `personal-file`.
    pub local_file: String,
    /// What the user picked (SPEC-20261119-agents-conflict-choices, SPEC-20261122-claude-md-layout):
    /// ids from [`PICKS`]. Absent when empty, so a request without picks digests as before.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub picks: Vec<String>,
}

/// The ids a request's `picks` may hold. `append-managed` appends the managed section to a
/// CLAUDE.md whose block was removed; `skip-conflicts` leaves each conflicting file as it is;
/// `import` and `move` name the picks of the project picks step (`--pick` of `agents-merge-plan`
/// and `agents-merge-apply`), which writes the block together with the import line or the copied
/// text. The writer alone never carries them out: a `CLAUDE.md` that held only the block would
/// stop Claude Code reading the `AGENTS.md` of the project, so a direct request that picks them
/// plans no `CLAUDE.md` and says so. `personal-file` creates `CLAUDE.local.md` once. The earlier
/// pick `claude-import` is gone: it is refused like any unknown id, and no code reads it by name.
pub const PICKS: [&str; 5] = [
    "append-managed",
    "skip-conflicts",
    "import",
    "move",
    "personal-file",
];

fn picked(r: &AgentsRequest, id: &str) -> bool {
    r.picks.iter().any(|p| p == id)
}

thread_local! {
    /// Set while the project picks step runs the writer for its own pick (see [`CarriedByPicks`]).
    static CARRIED_BY_PICKS: std::cell::Cell<bool> = const { std::cell::Cell::new(false) };
}

/// While it lives, the picks `import` and `move` of a request plan the block into `CLAUDE.md`:
/// the project picks step holds it for the writer calls of its own composite, and writes the
/// import line or the copied text in the same receipt. It is not a request value, so a plan
/// that comes from outside (a command, a stored plan, another caller) cannot set it.
pub(crate) struct CarriedByPicks(());

impl CarriedByPicks {
    pub(crate) fn hold() -> Self {
        CARRIED_BY_PICKS.with(|c| c.set(true));
        CarriedByPicks(())
    }
}

impl Drop for CarriedByPicks {
    fn drop(&mut self) {
        CARRIED_BY_PICKS.with(|c| c.set(false));
    }
}

fn carried_by_picks() -> bool {
    CARRIED_BY_PICKS.with(|c| c.get())
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct FilePlan {
    pub path: String,
    /// create | merge | append | unchanged | conflict | skip (a conflict the user chose to leave)
    pub action: String,
    /// Design intent for generated projects, not a measured git state. The `.gitignore` row
    /// asks git read-only questions (`ignore_block`: up to five calls, one of them a throw-away
    /// `git init` in the temporary folder); no other row runs git.
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
    /// Host notes of earlier versions; the writer no longer fills them (the files they named
    /// are now MoAI's own targets), so the list is empty. The field stays for the schema.
    pub host_notes: Vec<String>,
    pub warnings: Vec<String>,
    /// Kept for the schema and for plans written by an earlier version: the writer no longer
    /// sets it, because `CLAUDE.md` is MoAI's file and does not shadow anything.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub claude_shadow: Option<ClaudeShadow>,
    pub plan_digest: String,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct ClaudeShadow {
    /// `CLAUDE.md` or `.claude/CLAUDE.md`, relative to the project.
    pub path: String,
    /// The line an earlier version's pick appended to it. Kept so a plan that version wrote
    /// still reads; the writer no longer sets it.
    pub import_line: String,
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
    if r.picks.iter().any(|p| !PICKS.contains(&p.as_str())) {
        // Same phrase as above, so the error keeps its `invalid_input` kind (contract.rs).
        return Err(
            "invalid agents request; no values were logged; picks holds an id this tool does not know"
                .into(),
        );
    }
    Ok(())
}

/// What a planned file looks like on disk, read without failing the plan: a link (the file or a
/// folder above it inside the project) and text that is not UTF-8 are refused as a conflict row
/// that writes nothing, not as an error of the whole plan.
enum Probe {
    Absent,
    Text(Vec<u8>),
    Refused(&'static str),
}

fn probe(root: &Path, rel: &str) -> Result<Probe> {
    let path = match target(root, rel) {
        Ok(path) => path,
        Err(e) if e == "symlink target refused" => {
            return Ok(Probe::Refused(
                "the file, or a folder above it, is a link; nothing is written through a link",
            ))
        }
        Err(e) => return Err(e),
    };
    Ok(match read_opt(&path)? {
        None => Probe::Absent,
        Some(bytes) if std::str::from_utf8(&bytes).is_err() => {
            Probe::Refused("the file is not valid UTF-8 text; it is left as it is")
        }
        Some(bytes) => Probe::Text(bytes),
    })
}

/// The conflict row of a file the plan refuses to touch (a link or a file that is not text).
fn refused_row(path: &str, tracked: bool, section: Option<&str>, why: &str) -> FilePlan {
    FilePlan {
        path: path.into(),
        action: "conflict".into(),
        tracked_by_default: tracked,
        expected_sha256: None,
        section_id: section.map(String::from),
        baseline_sha256: None,
        rendered_sha256: None,
        detail: Some(why.into()),
    }
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
        .chain(p.versioning.as_ref().and_then(|v| v.format.as_ref()))
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
/// The language the generated AGENTS.md and REVIEW.md text is written in: the profile's
/// `language.instruction_files` when it is set to one the templates exist for (`en`, `ko`; an
/// older profile's explicit choice), else the conversation language `language.user_chat` when it
/// is one of those, else the request language. A profile without `language` renders exactly as
/// before.
fn instruction_language<'a>(p: &'a Profile, r: &'a AgentsRequest) -> &'a str {
    let supported = |key: &str| {
        p.language
            .as_ref()
            .and_then(|l| l.get(key))
            .map(String::as_str)
            .filter(|code| *code == "en" || *code == "ko")
    };
    supported("instruction_files")
        .or_else(|| supported("user_chat"))
        .unwrap_or(r.language.as_str())
}

/// The first heading of a REVIEW.md this tool creates (card C-60): written once, outside the
/// managed markers, in the instruction language like the managed block. A file that exists keeps
/// its own heading.
fn review_heading(language: &str) -> &'static str {
    if language == "ko" {
        "# 검토 지침\n\n"
    } else {
        "# Review instructions\n\n"
    }
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

/// The most lines the managed block holds between its two marker lines (REQ-3).
pub const BLOCK_MAX_LINES: usize = 15;

/// The writing-language facts on one line: `render_lines` prints a heading, one row per item the
/// profile sets and a closing sentence, which would use most of the block's lines; the rows alone
/// carry the facts. `None` when the profile sets no language.
fn writing_line(p: &Profile, ui: &str, root: &Path) -> Option<String> {
    let settings = p.language.as_ref()?;
    let full =
        crate::language::render_lines(settings, ui, &crate::state_root::rel(root, "project.yaml"));
    let rows: Vec<&str> = full.lines().filter_map(|l| l.strip_prefix("- ")).collect();
    if rows.is_empty() {
        return None;
    }
    let (head, tail) = if ui == "ko" {
        ("작성 언어", "그 밖은 영어")
    } else {
        ("Writing language", "everything else English")
    };
    Some(format!("{head}: {}; {tail}.\n", rows.join("; ")))
}

/// The managed block of `CLAUDE.md`: facts only, within [`BLOCK_MAX_LINES`] lines (REQ-3): the
/// project id, the tools, the check command, the language, the versioning line, the flow with its
/// integration branch, and the places of SPECs and cards. What to do and how to decide is the
/// flow rule's and the skills' to say, and those live in the config folder, not in the block.
fn render_agents(p: &Profile, r: &AgentsRequest, root: &Path) -> String {
    let (begin, end) = markers(AGENTS_SECTION, "html");
    let lang = instruction_language(p, r);
    let ko = lang == "ko";
    let tools = p.tools.join(", ");
    let writing = writing_line(p, lang, root).unwrap_or_default();
    // C-69: the versioning line appears only when the profile records a choice.
    let versioning = p
        .versioning
        .as_ref()
        .map(|v| crate::versioning::render_line(v, lang))
        .unwrap_or_default();
    let specs = format!("{}/specs/", state_root::dir(root));
    let flow = p
        .work_flow
        .as_ref()
        .map(|w| match (ko, w.integration_branch.as_deref()) {
            (true, Some(b)) => format!("작업 흐름: {} (통합 브랜치 `{b}`).\n", w.flow),
            (true, None) => format!("작업 흐름: {}.\n", w.flow),
            (false, Some(b)) => {
                format!("Work flow: {} (integration branch `{b}`).\n", w.flow)
            }
            (false, None) => format!("Work flow: {}.\n", w.flow),
        })
        .unwrap_or_default();
    let body = if ko {
        format!(
            "## MoAI 프로젝트 사실 (관리 구간)\n\n\
프로젝트 `{id}` · 도구: {tools}.\n\
검사: `moai check affected`.\n\
{writing}{versioning}{flow}\
SPEC: `{specs}` · 카드: `moai card list`.\n",
            id = p.project_id,
        )
    } else {
        format!(
            "## MoAI project facts (managed block)\n\n\
Project `{id}` · tools: {tools}.\n\
Checks: `moai check affected`.\n\
{writing}{versioning}{flow}\
SPECs: `{specs}` · cards: `moai card list`.\n",
            id = p.project_id,
        )
    };
    format!("{begin}\n{body}{end}\n")
}

/// The first-creation text of the personal file `CLAUDE.local.md`, in the request language. A
/// personal file is never tracked, so `tracked` renders like `ignored` (REQ-7).
fn render_local(r: &AgentsRequest) -> String {
    match r.language.as_str() {
        "ko" => "# 개인 작업 보충\n\n응답 언어: 한국어\n\n<!-- Claude Code가 CLAUDE.md 옆에서 함께 읽는 개인 파일. 비밀값을 넣지 않으며 repo에 공유하지 않는다. -->\n".into(),
        _ => "# Personal supplement\n\nResponse language: English\n\n<!-- Personal file Claude Code loads next to CLAUDE.md. Keep secrets out; never share it in the repo. -->\n".into(),
    }
}

/// The first-creation texts of the personal file `AGENTS.local.md` that earlier versions wrote
/// (the `ignored` and the `tracked` text, in each language).
const EARLIER_LOCAL_TEXTS: [&str; 4] = [
    "# 개인 작업 보충\n\n응답 언어: 한국어\n\n<!-- MoAI가 호출 시 명시적으로 읽는 개인 파일. 이 저장소에서는 추적을 허용했으므로 비밀값을 넣지 않는다. -->\n",
    "# 개인 작업 보충\n\n응답 언어: 한국어\n\n<!-- MoAI가 호출 시 명시적으로 읽는 개인 파일. 비밀값을 넣지 않으며 repo·Issue에 공유하지 않는다. -->\n",
    "# Personal supplement\n\nResponse language: English\n\n<!-- Personal file MoAI reads explicitly when invoked. Tracking is allowed in this repository, so keep secrets out of it. -->\n",
    "# Personal supplement\n\nResponse language: English\n\n<!-- Personal file MoAI reads explicitly when invoked. Keep secrets out; never share it in the repo or in Issues. -->\n",
];

/// `true` when `bytes` is exactly one of the personal-file texts this tool has written (#2936
/// R10): such a file holds nothing the user wrote, so a move into `CLAUDE.md` skips it.
pub fn is_generated_local(bytes: &[u8]) -> bool {
    let current = |language: &str| {
        render_local(&AgentsRequest {
            protocol_version: 1,
            language: language.into(),
            workflow_reference: String::new(),
            local_file: "ignored".into(),
            picks: Vec::new(),
        })
    };
    ["en", "ko"]
        .iter()
        .any(|language| current(language).as_bytes() == bytes)
        || EARLIER_LOCAL_TEXTS.iter().any(|t| t.as_bytes() == bytes)
}

/// What the project's state folder keeps in git (#3009 S3b, design §2): the profile, the
/// generation manifests and workflows this tool writes into it today, the two generation
/// markers (a fresh checkout classifies the folder by them), the team's own `config/`,
/// `decisions/` and `docs/`, and the SPECs in `specs/` for every project, with or without GitHub
/// Issues (user decision 2026-11-17, SDD always; replaces decision 6892663669). `specs/` stays
/// last so a block written before that decision for a project without Issues keeps its bytes.
/// Everything else in the folder is machine output and is ignored.
const STATE_KEPT: [&str; 9] = [
    "project.yaml",
    "moai-cli.json",
    "moai.json",
    "manifests/",
    "workflows/",
    "config/",
    "decisions/",
    "docs/",
    "specs/",
];

/// The `kind` an owned `.gitignore` block is recorded with: whether it lists `AGENTS.local.md`
/// (#3009 S3b). A block recorded before this field existed listed it, so `None` reads as listed.
const IGNORE_KIND_WITH_LOCAL: &str = "agents:ignore-with-local";
const IGNORE_KIND_STATE_ONLY: &str = "agents:ignore-state";

/// The managed `.gitignore` block (#3009 S3b, leader decision A): the receipts line, the
/// state-folder allow-list, `CLAUDE.local.md` only when `list_local` (the personal file is kept
/// out of git), and the line `AGENTS.local.md` when the block already carried it (`keep_earlier`:
/// that file is still the user's, and the block never drops a line that keeps it out of git).
///
/// `allow` is the list of `!` paths the block carries, in order (see [`ignore_block`]).
fn render_ignore(root: &Path, list_local: bool, keep_earlier: bool, allow: &[String]) -> String {
    let (begin, end) = markers(IGNORE_SECTION, "hash");
    let state = state_root::dir(root);
    let mut body = String::new();
    if keep_earlier {
        body.push_str(&format!("{}\n", LOCAL_NAMES[0]));
    }
    if list_local {
        body.push_str(&format!("{LOCAL_FILE}\n"));
    }
    // HOLD-B1 (user-confirmed 2026-11-08, ADK_1989_CONFIRMED_20261108_01): apply
    // receipts stay out of git alongside the
    // personal file; src/apply.rs::receipts_rel is the single source of the path.
    body.push_str(&format!("{}/\n", apply::receipts_rel(root)));
    body.push_str(&format!("{state}/*\n"));
    for path in allow {
        // `ignore_block` refused a name that cannot be written, so the fallback is not reached.
        let line = escape_ignore_path(path).unwrap_or_else(|| path.clone());
        body.push_str(&format!("!{line}\n"));
    }
    format!("{begin}\n{body}{end}\n")
}

/// `path` as one `.gitignore` pattern that matches exactly that path: a backslash before the
/// characters git reads as pattern syntax (`* ? [ \`) and before trailing spaces. `None` for a
/// name with a line break, which no pattern line can hold.
fn escape_ignore_path(path: &str) -> Option<String> {
    if path.contains(['\n', '\r']) {
        return None;
    }
    let mut out = String::with_capacity(path.len());
    for c in path.chars() {
        if matches!(c, '*' | '?' | '[' | '\\') {
            out.push('\\');
        }
        out.push(c);
    }
    let kept = out.trim_end_matches(' ').len();
    let spaces = out.len() - kept;
    out.truncate(kept);
    for _ in 0..spaces {
        out.push_str("\\ ");
    }
    Some(out)
}

/// The rendered block and what the plan tells about how it was made.
struct IgnoreBlock {
    rendered: String,
    /// Preview lines for the plan's `detail`: paths left ignored for a user rule, paths kept
    /// because git tracks them under the folder an update moves away from.
    notes: Vec<String>,
}

/// The `!` allow-list of the block, shaped by two facts about the project (#3009 S3b, audit
/// 6897773189 B1/B3, leader decision 6897797810):
///
/// - a rule the user wrote wins: outside the block in the project's root `.gitignore`, in
///   `.git/info/exclude`, or in the global excludes file git reads (leader decision 6897897560).
///   A path such a rule already ignores gets no `!` line, so the block never changes what one
///   of the user's rules does. A `!` line in the root `.gitignore` outranks the other two
///   sources, which is why they are read too. The block is committed, so another computer's
///   global rules cannot be seen from here; planning again there fits the block to them;
/// - a project update moves the state folder from `.moai-cli/` to `.moai/`: what git tracks
///   in the earlier folder keeps its place in the new one, so a later `git add -A` does not
///   take the old paths out of git and leave the new ones ignored.
///
/// Plan and apply both call this, so the preview and the write cannot differ.
fn ignore_block(root: &Path, list_local: bool, ignore_text: Option<&str>) -> Result<IgnoreBlock> {
    let state = state_root::dir(root);
    let kept: Vec<String> = STATE_KEPT.iter().map(|k| format!("{state}/{k}")).collect();
    let user = ignore_text.map(outside_block).unwrap_or_default();
    let mut notes: Vec<String> = Vec::new();
    // Names git tracks in the earlier folder, and the path each takes in this one. A name the
    // receipts line already covers is not asked about.
    let receipts = format!("{}/", apply::receipts_rel(root));
    let mut moved: Vec<(String, String)> = Vec::new();
    for name in tracked_in_earlier_folder(root, state, &mut notes)? {
        let path = format!("{state}/{name}");
        if format!("{path}/") == receipts || path == receipts {
            continue;
        }
        if escape_ignore_path(&path).is_none() {
            return Err(format!(
                "git tracks {}/{name}, a name a .gitignore line cannot hold (it has a line break); \
                 rename it before the update",
                crate::PROJECT_STATE_DIR
            ));
        }
        moved.push((name, path));
    }
    // One question to git covers the fixed list and the moved names: a user rule that ignores a
    // path wins over both (audit 6899590801 N1), whether or not git tracked it before.
    let mut asked = kept.clone();
    for (_, path) in &moved {
        if !asked.contains(path) {
            asked.push(path.clone());
        }
    }
    let hits = user_rule_hits(root, &user, &asked, &mut notes)?;
    let mut shown: Vec<String> = Vec::new();
    for h in &hits {
        let line = match moved.iter().find(|(_, path)| *path == h.path) {
            Some((name, _)) => format!(
                "tracking ends: your rule `{}` in {} ignores {}, which git tracks as {}/{name}",
                h.rule,
                h.origin,
                h.path,
                crate::PROJECT_STATE_DIR
            ),
            None => format!(
                "left ignored by your rule `{}` in {}: {}",
                h.rule, h.origin, h.path
            ),
        };
        shown.push(line);
    }
    shown.append(&mut notes);
    let mut allow: Vec<String> = kept
        .into_iter()
        .filter(|p| !hits.iter().any(|h| h.path == *p))
        .collect();
    for (name, path) in moved {
        if allow.contains(&path) || hits.iter().any(|h| h.path == path) {
            continue;
        }
        shown.push(format!(
            "kept because git tracks {}/{name}, which the update moves to {path}",
            crate::PROJECT_STATE_DIR
        ));
        allow.push(path);
    }
    // The block of an earlier version listed `AGENTS.local.md`: that line stays.
    let keep_earlier = ignore_text.is_some_and(|text| block_lists(text, LOCAL_NAMES[0]));
    Ok(IgnoreBlock {
        rendered: render_ignore(root, list_local, keep_earlier, &allow),
        notes: shown,
    })
}

/// `true` when a rule for `name` stands inside this tool's managed `.gitignore` block of `text`.
fn block_lists(text: &str, name: &str) -> bool {
    let (begin, end) = markers(IGNORE_SECTION, "hash");
    matches!(find_block(text, &begin, &end), Ok(Some((_, _, body))) if ignore_names(&body, name))
}

/// The part of a `.gitignore` the user wrote: its text outside this tool's managed block. A
/// text with malformed markers is returned whole (the plan reports those as a conflict).
fn outside_block(text: &str) -> String {
    let (begin, end) = markers(IGNORE_SECTION, "hash");
    match find_block(text, &begin, &end) {
        Ok(Some((start, block_end, _))) => format!("{}{}", &text[..start], &text[block_end..]),
        _ => text.to_string(),
    }
}

/// What a spawn error of `git` is called, so a caller can tell "no git" from "git said no".
const GIT_UNAVAILABLE: &str = "git unavailable";

/// `git` in `dir` with the scrubbed environment; `input` goes to its stdin. stderr is kept only
/// to classify a failure (`ask_project`); it is never shown.
fn git_output(dir: &Path, args: &[&str], input: &[u8]) -> Result<std::process::Output> {
    use std::io::Write;
    use std::process::{Command, Stdio};
    let mut cmd = Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let mut child = cmd
        .arg("-C")
        .arg(dir)
        .args(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .map_err(|_| GIT_UNAVAILABLE.to_string())?;
    if let Some(mut pipe) = child.stdin.take() {
        // A write that fails shows up as git's own exit code.
        let _ = pipe.write_all(input);
    }
    child
        .wait_with_output()
        .map_err(|_| "git failed to run".to_string())
}

/// The one-line summary of a failed git call, without paths (`identity::failure_summary`).
fn git_failure(what: &str, out: &std::process::Output) -> String {
    let summary = crate::identity::failure_summary(&out.stderr);
    let summary = if out.stderr.iter().all(u8::is_ascii_whitespace) {
        "no message".to_string()
    } else {
        summary
    };
    format!("cannot tell {what}: git said: {summary}")
}

/// Records, once, that git is missing so the project's own ignore state could not be read.
fn note_git_missing(notes: &mut Vec<String>) {
    let note =
        "git is not available, so your ignore rules and the files git tracks were not considered";
    if !notes.iter().any(|n| n == note) {
        notes.push(note.to_string());
    }
}

/// A read-only git question about the project, where only two failures mean "nothing to read":
/// the project is not a git repository, and git is not installed (the second is noted in the
/// preview). Any other failure stops the plan, because reading it as "nothing" would drop a
/// protection without a word (audit 6899590801 N2). `Ok(None)` is "nothing to read".
fn ask_project(
    root: &Path,
    args: &[&str],
    what: &str,
    notes: &mut Vec<String>,
) -> Result<Option<Vec<u8>>> {
    let out = match git_output(root, args, b"") {
        Ok(out) => out,
        Err(e) if e == GIT_UNAVAILABLE => {
            note_git_missing(notes);
            return Ok(None);
        }
        Err(e) => return Err(e),
    };
    if out.status.success() {
        return Ok(Some(out.stdout));
    }
    if String::from_utf8_lossy(&out.stderr).contains(crate::identity::NOT_A_REPOSITORY) {
        return Ok(None);
    }
    Err(git_failure(what, &out))
}

/// One allow-list path a user rule already ignores.
struct RuleHit {
    path: String,
    rule: String,
    /// The file the rule stands in: `.gitignore`, `.git/info/exclude`, or the global excludes
    /// file (its home folder shown as `~`).
    origin: String,
}

fn has_rule(text: &str) -> bool {
    text.lines().any(|l| {
        let t = l.trim();
        !t.is_empty() && !t.starts_with('#')
    })
}

/// `path` with the home folder at its start written as `~`.
fn tilde(path: &Path) -> String {
    let shown = path.display().to_string();
    let Some(home) = std::env::var("HOME").ok().filter(|h| !h.is_empty()) else {
        return shown;
    };
    match shown.strip_prefix(home.trim_end_matches('/')) {
        Some(rest) if rest.is_empty() || rest.starts_with('/') => format!("~{rest}"),
        _ => shown,
    }
}

/// Where git looks for the global excludes file when `core.excludesFile` is not set.
fn default_global_excludes() -> Option<std::path::PathBuf> {
    let config = std::env::var_os("XDG_CONFIG_HOME")
        .filter(|v| !v.is_empty())
        .map(std::path::PathBuf::from)
        .or_else(|| {
            std::env::var_os("HOME")
                .filter(|v| !v.is_empty())
                .map(|h| std::path::PathBuf::from(h).join(".config"))
        })?;
    Some(config.join("git").join("ignore"))
}

/// The global excludes file git reads for `root` and its text. A file that is configured but
/// missing, or present but unreadable, gives no rules and a one-line note for the preview
/// (leader decision 6897897560, item 7); an unset setting with no default file gives neither.
fn global_excludes(
    root: &Path,
    notes: &mut Vec<String>,
) -> Result<(Option<std::path::PathBuf>, String)> {
    // Unit tests run with the developer's real HOME; the real global file must not shape them.
    // The binary tests (tests/gitignore_block_cli.rs) run this code with an empty HOME.
    if cfg!(test) {
        return Ok((None, String::new()));
    }
    // `git config --get` answers 0 with the value and 1 when the key is not set; any other
    // answer is a broken configuration, which stops the plan.
    let configured = match git_output(
        root,
        &["config", "--type=path", "--get", "core.excludesFile"],
        b"",
    ) {
        Ok(out) => match out.status.code() {
            Some(0) => Some(String::from_utf8_lossy(&out.stdout).trim().to_string())
                .filter(|p| !p.is_empty())
                .map(std::path::PathBuf::from),
            Some(1) => None,
            _ => return Err(git_failure("what git's core.excludesFile says", &out)),
        },
        Err(e) if e == GIT_UNAVAILABLE => {
            note_git_missing(notes);
            return Ok((None, String::new()));
        }
        Err(e) => return Err(e),
    };
    let was_configured = configured.is_some();
    let Some(path) = configured.or_else(default_global_excludes) else {
        return Ok((None, String::new()));
    };
    match std::fs::read_to_string(&path) {
        Ok(text) => Ok((Some(path), text)),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound && !was_configured => {
            Ok((Some(path), String::new()))
        }
        Err(_) => {
            notes.push(format!(
                "global excludes file {} is missing or unreadable; its rules are not considered",
                tilde(&path)
            ));
            Ok((Some(path), String::new()))
        }
    }
}

/// The text of the project's `.git/info/exclude`: empty when the project is not a repository or
/// the file is not there; a file that is there and cannot be read stops the plan.
fn info_exclude(root: &Path, notes: &mut Vec<String>) -> Result<String> {
    let Some(out) = ask_project(
        root,
        &["rev-parse", "--git-path", "info/exclude"],
        "where the project's .git/info/exclude is",
        notes,
    )?
    else {
        return Ok(String::new());
    };
    let path = String::from_utf8_lossy(&out).trim().to_string();
    if path.is_empty() {
        return Ok(String::new());
    }
    match std::fs::read(root.join(path)) {
        Ok(bytes) => String::from_utf8(bytes)
            .map_err(|_| "cannot read .git/info/exclude: it is not UTF-8 text".to_string()),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(String::new()),
        Err(_) => Err("cannot read .git/info/exclude".to_string()),
    }
}

/// Which of `paths` the user's own ignore rules already ignore, with the rule and the file it
/// stands in: `user_rules` (the root `.gitignore` outside the block), `.git/info/exclude`, and
/// the global excludes file. Git answers: the three texts are written to a throw-away
/// repository in the temporary folder, where git applies them in its own order, and asked with
/// `check-ignore --no-index`, so the project is neither read through git's ignore machinery nor
/// written and a negated rule counts as not ignoring. A directory path ends in `/`. Nothing is
/// asked when no source holds a rule. What cannot be read about the sources goes to `notes`
/// (the preview) or, when git answers wrongly, stops the plan.
fn user_rule_hits(
    root: &Path,
    user_rules: &str,
    paths: &[String],
    notes: &mut Vec<String>,
) -> Result<Vec<RuleHit>> {
    let (global_path, global_text) = global_excludes(root, notes)?;
    let info_text = info_exclude(root, notes)?;
    if paths.is_empty() || !(has_rule(user_rules) || has_rule(&info_text) || has_rule(&global_text))
    {
        return Ok(Vec::new());
    }
    let unavailable =
        "cannot tell which paths your ignore rules ignore: git check-ignore did not answer";
    let tmp = tempfile::tempdir().map_err(|_| "temporary folder unavailable".to_string())?;
    let dir = tmp.path();
    let no_excludes = dir.join("global-excludes");
    std::fs::write(&no_excludes, &global_text)
        .map_err(|_| "temporary folder unavailable".to_string())?;
    std::fs::write(dir.join(".gitignore"), user_rules)
        .map_err(|_| "temporary folder unavailable".to_string())?;
    if !git_output(dir, &["init", "-q"], b"")?.status.success() {
        return Err(unavailable.into());
    }
    let info_dir = dir.join(".git").join("info");
    std::fs::create_dir_all(&info_dir).map_err(|_| "temporary folder unavailable".to_string())?;
    std::fs::write(info_dir.join("exclude"), &info_text)
        .map_err(|_| "temporary folder unavailable".to_string())?;
    let excludes = format!("core.excludesFile={}", no_excludes.display());
    let mut input = Vec::new();
    for path in paths {
        input.extend_from_slice(path.as_bytes());
        input.push(0);
    }
    let out = git_output(
        dir,
        &[
            "-c",
            excludes.as_str(),
            "check-ignore",
            "-v",
            "-z",
            "--no-index",
            "--stdin",
        ],
        &input,
    )?;
    // 0: some path matched; 1: none did; anything else is a failure.
    if !matches!(out.status.code(), Some(0 | 1)) {
        return Err(unavailable.into());
    }
    // `-v -z` prints four NUL-ended fields per match: source, line, pattern, path.
    let fields: Vec<&[u8]> = out.stdout.split(|b| *b == 0).collect();
    let mut hits = Vec::new();
    for record in fields.chunks_exact(4) {
        let (Ok(source), Ok(pattern), Ok(path)) = (
            std::str::from_utf8(record[0]),
            std::str::from_utf8(record[2]),
            std::str::from_utf8(record[3]),
        ) else {
            return Err(unavailable.into());
        };
        // A negated rule matching a path means the path is not ignored.
        if pattern.starts_with('!') {
            continue;
        }
        let origin = if source == ".gitignore" {
            ".gitignore".to_string()
        } else if source.replace('\\', "/").ends_with("info/exclude") {
            ".git/info/exclude".to_string()
        } else {
            global_path
                .as_deref()
                .map_or_else(|| "the global excludes file".to_string(), tilde)
        };
        hits.push(RuleHit {
            path: path.to_string(),
            rule: pattern.to_string(),
            origin,
        });
    }
    Ok(hits)
}

/// Top-level names (`specs/`, `state/`, `project.yaml`) git tracks in the earlier state
/// folder `.moai-cli/` while the project's state folder is `.moai/`: the update moves them
/// there. Empty when the state folder is not `.moai/`, when the project is not a git
/// repository, or when git is not installed (noted in the preview). Any other failure of git,
/// and a tracked name that is not UTF-8 text, stop the plan: a list that silently lacks names
/// would let a later `git add -A` drop them.
fn tracked_in_earlier_folder(
    root: &Path,
    state: &str,
    notes: &mut Vec<String>,
) -> Result<Vec<String>> {
    if state != state_root::MOAI_DIR {
        return Ok(Vec::new());
    }
    let Some(stdout) = ask_project(
        root,
        &["ls-files", "-z", "--", crate::PROJECT_STATE_DIR],
        "which files git tracks in the earlier state folder",
        notes,
    )?
    else {
        return Ok(Vec::new());
    };
    let mut names = std::collections::BTreeSet::new();
    for path in stdout.split(|b| *b == 0).filter(|p| !p.is_empty()) {
        let path = std::str::from_utf8(path).map_err(|_| {
            format!(
                "git tracks a name under {}/ that is not UTF-8 text; rename it before the update",
                crate::PROJECT_STATE_DIR
            )
        })?;
        let Some((_, rest)) = path.split_once('/') else {
            continue;
        };
        names.insert(match rest.split_once('/') {
            Some((first, _)) => format!("{first}/"),
            None => rest.to_string(),
        });
    }
    Ok(names.into_iter().collect())
}

fn ignore_kind(list_local: bool) -> &'static str {
    if list_local {
        IGNORE_KIND_WITH_LOCAL
    } else {
        IGNORE_KIND_STATE_ONLY
    }
}

/// `true` when the manifest's owned `.gitignore` block lists `AGENTS.local.md` (#3009 S3b): its
/// recorded kind says so, or it has none (a block written before the kind existed listed it).
pub(crate) fn ignore_kind_lists_local(kind: Option<&str>) -> bool {
    kind.is_none_or(|k| k == IGNORE_KIND_WITH_LOCAL)
}

/// The line-level preview of a managed block change for the plan's `detail` (#3009 S3b,
/// condition 1): `+` a line the block gains, `-` a line it loses, nothing for a kept line.
fn block_line_preview(current: Option<&[u8]>, begin: &str, end: &str, rendered: &str) -> String {
    let body = |text: &str| -> Vec<String> {
        text.lines()
            .filter(|l| *l != begin && *l != end)
            .map(String::from)
            .collect()
    };
    let old: Vec<String> = current
        .and_then(|b| std::str::from_utf8(b).ok())
        .and_then(|text| {
            find_block(text, begin, end)
                .ok()
                .flatten()
                .map(|(s, e, _)| body(&text[s..e]))
        })
        .unwrap_or_default();
    let new = body(rendered);
    let added: Vec<&String> = new.iter().filter(|l| !old.contains(l)).collect();
    let removed: Vec<&String> = old.iter().filter(|l| !new.contains(l)).collect();
    let mut out = format!(
        "managed block lines: {} added, {} removed",
        added.len(),
        removed.len()
    );
    for l in &added {
        out.push_str(&format!("\n+ {l}"));
    }
    for l in &removed {
        out.push_str(&format!("\n- {l}"));
    }
    out
}

/// The names the personal file has had: `AGENTS.local.md` before Claude Code became the target,
/// `CLAUDE.local.md` after it. Claude Code reads the second and never the first, so a project
/// may hold either record; the readers accept both.
pub(crate) const LOCAL_NAMES: [&str; 2] = ["AGENTS.local.md", "CLAUDE.local.md"];

fn ignore_names(text: &str, name: &str) -> bool {
    text.lines().any(|l| {
        let t = l.trim();
        t == name || t.strip_prefix('/') == Some(name)
    })
}

/// `true` when a rule in `text` names the personal file this version generates
/// (`CLAUDE.local.md`); a rule for the other name is not a rule for that file.
pub(crate) fn ignore_has_rule(text: &str) -> bool {
    ignore_names(text, LOCAL_NAMES[1])
}

/// `true` when a rule in `text` names the personal file under either of its names; for what a
/// recorded block lists, whichever name that generation wrote.
pub(crate) fn ignore_has_any_local_rule(text: &str) -> bool {
    LOCAL_NAMES.iter().any(|name| ignore_names(text, name))
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
    // The personal file counts under either of its names; any other path under its own.
    let names: &[&str] = if LOCAL_NAMES.contains(&path) {
        &LOCAL_NAMES
    } else {
        &[path]
    };
    m.is_some_and(|m| {
        m.created_once.iter().any(|p| names.contains(&p.as_str()))
            // Manifests written before `created_once` existed recorded the personal file in owned_files.
            || m.owned_files
                .iter()
                .any(|f| names.contains(&f.path.as_str()) && f.section_id.is_none())
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

/// The conflict detail of a baseline that has no block in the file any more.
pub(crate) const BLOCK_MISSING: &str = "managed block missing or removed; baseline mismatch";
const APPEND_PICKED: &str =
    "append-managed picked: the managed section is added at the end; every other byte is kept";
const SKIP_PICKED: &str = "skip-conflicts picked: the file is kept as it is";

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
                BLOCK_MISSING.into()
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
    record_owned_kind(owned, path, section, bytes, None);
}

fn record_owned_kind(
    owned: &mut Vec<OwnedFile>,
    path: &str,
    section: &str,
    bytes: &[u8],
    kind: Option<&str>,
) {
    owned.retain(|f| f.path != path);
    owned.push(OwnedFile {
        path: path.into(),
        section_id: Some(section.into()),
        baseline_sha256: digest(bytes),
        kind: kind.map(String::from),
    });
}

/// Whether the personal file's line goes into the managed `.gitignore` block (#3009 S3b): the
/// choice is `ignored` or `tracked` (a personal file is never tracked, so both list it), and no
/// rule of that name stands in the user's own text, the part of the file outside the block. The
/// block's own line is never counted as a user rule (audit 6897773189 B4): with the generation
/// manifest gone, a block that lists the file would otherwise read itself as the user's rule,
/// plan a different block and stay in conflict. A line `!CLAUDE.local.md` in the user's text
/// re-includes the file; the block, which stands after it, would ignore it again, so it does not
/// list the file then.
fn lists_local(request: &AgentsRequest, ignore_text: Option<&str>) -> bool {
    request.local_file != "none"
        && !ignore_text.is_some_and(|text| {
            let user = outside_block(text);
            ignore_has_rule(&user) || ignore_negates_local(&user)
        })
}

/// `true` when a line of `text` re-includes the personal file this version generates.
fn ignore_negates_local(text: &str) -> bool {
    text.lines().any(|l| {
        let t = l.trim();
        t.strip_prefix('!').is_some_and(|rest| {
            rest == LOCAL_NAMES[1] || rest.strip_prefix('/') == Some(LOCAL_NAMES[1])
        })
    })
}

/// What an apply carries over from the previous manifest: the generation linkage (#2127:
/// it belongs to the last generation apply), the owned sections, and the personal files seen
/// once (a legacy entry without a section becomes created-once).
type Carried = (Option<String>, Option<String>, Vec<OwnedFile>, Vec<String>);
fn carried(active: &[&str], previous: Option<Manifest>) -> Carried {
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
        .filter(|f| {
            f.section_id
                .as_deref()
                .is_some_and(|id| !owned_entry_retired(id, active))
        })
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
                    if section == IGNORE_SECTION {
                        let lists = std::str::from_utf8(block).is_ok_and(ignore_has_any_local_rule);
                        record_owned_kind(owned, &f.path, section, block, Some(ignore_kind(lists)));
                    } else {
                        record_owned(owned, &f.path, section, block);
                    }
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
    active: &[&str],
    root: &Path,
    profile: &Profile,
    profile_bytes: &[u8],
    previous: Option<Manifest>,
    request: &AgentsRequest,
    files: &[FilePlan],
) -> Result<Manifest> {
    let manifest_rel = manifest_rel(root);
    let mut carried = carried(active, previous);
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
    active: &[&str],
    root: &Path,
    request: &AgentsRequest,
    files: &[FilePlan],
) -> Result<Value> {
    let root = root_of(root)?;
    let (profile, profile_bytes) = read_profile(&root)?;
    let previous = read_manifest(&root, &profile.project_id)?;
    let m = predict_manifest(
        active,
        &root,
        &profile,
        &profile_bytes,
        previous,
        request,
        files,
    )?;
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
    // The files that used to be listed here (`CLAUDE.md`, `CLAUDE.local.md`) are this tool's own
    // targets now, and `AGENTS.override.md` belonged to a host this tool no longer writes for.
    let host_notes: Vec<String> = Vec::new();

    let agents_rendered = render_agents(&profile, &request, &root);
    // REQ-1, REQ-2: the block goes into `CLAUDE.md` when the project holds no `AGENTS.md`, when
    // `CLAUDE.md` already holds the block or the import of `AGENTS.md` (a settled project keeps
    // getting its block refreshed), or when the project picks step carries out `import` or `move`
    // (it adds the import line or the copied text in the same receipt). Otherwise Claude Code
    // would stop reading the user's `AGENTS.md` the moment this tool wrote a CLAUDE file, so
    // nothing is planned and the preview says a pick is needed; a request that picks `import` or
    // `move` on its own gets the same, because a block alone would hide `AGENTS.md`.
    let claude = probe(&root, BLOCK_FILE)?;
    let agents_md_exists = std::fs::symlink_metadata(root.join(AGENTS_FILE)).is_ok();
    let (block_begin, _) = markers(AGENTS_SECTION, "html");
    let settled = match &claude {
        Probe::Text(bytes) => {
            // `probe` returned text, so it is UTF-8.
            std::str::from_utf8(bytes).is_ok_and(|t| t.contains(&block_begin))
                || crate::claude::imports_agents(&root.join(BLOCK_FILE))
        }
        Probe::Absent | Probe::Refused(_) => false,
    };
    let carried_pick =
        carried_by_picks() && (picked(&request, "import") || picked(&request, "move"));
    let block_wanted = !agents_md_exists || settled || carried_pick;
    let mut claude_row = None;
    if !block_wanted {
        let direct = ["import", "move"]
            .into_iter()
            .filter(|p| picked(&request, p))
            .map(|p| format!("`{p}`"))
            .collect::<Vec<_>>();
        let direct = if direct.is_empty() {
            String::new()
        } else {
            format!(
                "; the pick {} is carried out only by the project picks step (`--pick` of \
                 agents-merge-plan and agents-merge-apply), which writes the block together with \
                 the text, so this request writes no {BLOCK_FILE}",
                direct.join(" and ")
            )
        };
        warnings.push(format!(
            "{AGENTS_FILE} exists, so no {BLOCK_FILE} is planned: Claude Code reads {BLOCK_FILE} \
             instead of {AGENTS_FILE} once it exists. A pick is needed: `move` (recommended: the \
             block and a copy of your text), `import` (the block and an `@{AGENTS_FILE}` line) or \
             `leave` (nothing is written){direct}"
        ));
    } else {
        let row = match &claude {
            Probe::Refused(why) => refused_row(BLOCK_FILE, true, Some(AGENTS_SECTION), why),
            Probe::Absent | Probe::Text(_) => {
                let current = match &claude {
                    Probe::Text(bytes) => Some(bytes.as_slice()),
                    _ => None,
                };
                let mut row = section_plan(
                    BLOCK_FILE,
                    true,
                    AGENTS_SECTION,
                    "html",
                    current,
                    baseline_for(m, BLOCK_FILE),
                    &agents_rendered,
                )?;
                // `append-managed`: a baseline whose block is gone is planned as a first append
                // (no baseline), which keeps every other byte. An edited block or malformed
                // markers stay conflicts: only the missing-block conflict is turned.
                if picked(&request, "append-managed")
                    && row.action == "conflict"
                    && row.detail.as_deref() == Some(BLOCK_MISSING)
                {
                    let appended = section_plan(
                        BLOCK_FILE,
                        true,
                        AGENTS_SECTION,
                        "html",
                        current,
                        None,
                        &agents_rendered,
                    )?;
                    if appended.action == "append" {
                        row = FilePlan {
                            detail: Some(APPEND_PICKED.into()),
                            ..appended
                        };
                    }
                }
                row
            }
        };
        claude_row = Some(row);
    }
    if let Some(row) = claude_row {
        if row.detail.as_deref() == Some(BLOCK_ABSENT) {
            warnings.push(BLOCK_ABSENT.into());
        }
        files.push(row);
    }
    // The managed text follows `instruction_files` when the profile sets it and the
    // conversation language `user_chat` otherwise, when templates exist for it (`en`, `ko`);
    // another code is reported and the request language used.
    {
        let chosen = instruction_language(&profile, &request);
        let language = profile.language.as_ref();
        let named = |key: &str| {
            language
                .and_then(|l| l.get(key))
                .filter(|c| c.as_str() != chosen)
        };
        if let Some((key, code)) = named("instruction_files")
            .map(|c| ("instruction_files", c))
            .or_else(|| {
                language
                    .filter(|l| !l.contains_key("instruction_files"))
                    .and(named("user_chat"))
                    .map(|c| ("user_chat", c))
            })
        {
            warnings.push(format!(
                "managed section rendered in `{chosen}`: no managed-section text in profile language.{key} `{code}`"
            ));
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

    // REQ-6, REQ-7: the personal file is created once, and only on the user's pick. A personal
    // file is not shared, so `tracked` plans exactly like `ignored`; the preview says so.
    if request.local_file == "tracked" {
        warnings.push(format!(
            "local_file `tracked` is read as `ignored`: {LOCAL_FILE} is a personal file and is kept out of git"
        ));
    }
    if picked(&request, "personal-file") {
        if request.local_file == "none" {
            warnings.push(format!(
                "personal-file picked while local_file is `none`: {LOCAL_FILE} is not planned"
            ));
        } else {
            match probe(&root, LOCAL_FILE)? {
                Probe::Refused(why) => files.push(refused_row(LOCAL_FILE, false, None, why)),
                found => {
                    let local = match &found {
                        Probe::Text(bytes) => Some(bytes.as_slice()),
                        _ => None,
                    };
                    let local_rendered = render_local(&request);
                    files.push(FilePlan {
                        path: LOCAL_FILE.into(),
                        action: if local.is_none() && !created_once(m, LOCAL_FILE) {
                            "create"
                        } else {
                            "unchanged"
                        }
                        .into(),
                        tracked_by_default: false,
                        expected_sha256: local.map(digest),
                        section_id: None,
                        baseline_sha256: None,
                        rendered_sha256: Some(digest(local_rendered.as_bytes())),
                        detail: None,
                    });
                }
            }
        }
    }

    // #3009 S3b (leader decision A): the managed `.gitignore` block is always planned, for the
    // state-folder allow-list; `AGENTS.local.md` is listed in it only when the personal file is
    // `ignored` (the behaviour before S3b) and no user rule of the same name already stands
    // outside a block this tool owns.
    {
        let ignore = read_opt(&target(&root, ".gitignore")?)?;
        let ignore_text = match &ignore {
            Some(b) => Some(std::str::from_utf8(b).map_err(|_| "target is not UTF-8")?),
            None => None,
        };
        let list_local = lists_local(&request, ignore_text);
        let user_rule = request.local_file != "none" && !list_local;
        let (begin, end) = markers(IGNORE_SECTION, "hash");
        let block = ignore_block(&root, list_local, ignore_text)?;
        let rendered = block.rendered;
        let mut plan = section_plan(
            ".gitignore",
            true,
            IGNORE_SECTION,
            "hash",
            ignore.as_deref(),
            baseline_for(m, ".gitignore"),
            &rendered,
        )?;
        // Condition 1 of the decision: the block's lines show in the preview, for a new block
        // and for a widened one alike.
        if matches!(plan.action.as_str(), "create" | "append" | "merge") {
            let mut detail = block_line_preview(ignore.as_deref(), &begin, &end, &rendered);
            // Which user rules stood in the way of an allow line, and which tracked paths the
            // block keeps: named in the preview, not only visible as a missing or extra line.
            for note in &block.notes {
                detail.push_str(&format!("\n{note}"));
            }
            plan.detail = Some(detail);
        }
        if user_rule {
            let note = "existing user rule names CLAUDE.local.md; effective git ignore status not verified";
            plan.detail = Some(match plan.detail.take() {
                Some(d) => format!("{note}\n{d}"),
                None => note.into(),
            });
        }
        files.push(plan);
    }

    // The first CLAUDE file of the project makes Claude Code stop reading `AGENTS.md` and
    // `.claude/AGENTS.md`: the plan names each one that exists, whether this request creates the
    // file or a pick (which the warning above asks for) would. Nothing is warned about once a
    // CLAUDE file exists.
    let creates_first = files
        .iter()
        .find(|f| f.action == "create" && (f.path == BLOCK_FILE || f.path == LOCAL_FILE))
        .map(|f| f.path.as_str())
        .or((!block_wanted).then_some(BLOCK_FILE));
    if let Some(creating) = creates_first {
        // The block brings the root AGENTS.md along (a pick copies its text or imports it), so
        // only the files nothing carries over are named.
        warnings.extend(crate::agents_local_merge::hidden_files_warning_except(
            &root,
            creating,
            &["AGENTS.md"],
        ));
    }

    // `skip-conflicts`: a conflicting file is left as it is and the rest of the plan applies.
    if picked(&request, "skip-conflicts") {
        for f in files.iter_mut().filter(|f| f.action == "conflict") {
            let why = f.detail.take().unwrap_or_else(|| "conflict".into());
            f.action = "skip".into();
            f.detail = Some(format!("{SKIP_PICKED} ({why})"));
        }
    }

    let manifest_bytes = read_opt(&target(&root, &manifest_rel)?)?;
    // #2936 Q3 B: the manifest is `unchanged` when every other file is and the bytes an apply
    // would publish equal the file on disk, so a re-run plans `unchanged` throughout. A
    // skipped file adds nothing to the manifest either, so it counts like `unchanged`.
    let manifest_action = match &manifest_bytes {
        None => "create",
        Some(disk)
            if files
                .iter()
                .all(|f| f.action == "unchanged" || f.action == "skip")
                && manifest_bytes_of(&predict_manifest(
                    &crate::generation::HOSTS,
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
        claude_shadow: None,
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
    agents_apply_amended_for(
        &crate::generation::HOSTS,
        root,
        p,
        confirmed,
        permission,
        false,
        amend,
    )
}

/// `agents_apply_amended` whose re-plan uses `agents_plan_for`: generation passes `true`
/// only when its own `amend` hook writes the `workflow_reference` directory (#2274).
pub(crate) fn agents_apply_amended_for(
    active: &[&str],
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
    let (workflow_id, interview_revision, mut owned, mut created) = carried(active, previous);
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
        if f.action == "skip" {
            // The user picked to leave this file: no write, no baseline. The receipt names it so
            // recovery tooling sees why a planned file was not touched.
            receipt.entry(
                &f.path,
                "skip",
                f.expected_sha256.as_deref(),
                f.expected_sha256.as_deref(),
                false,
                f.detail.clone(),
            );
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
            BLOCK_FILE => {
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
                let heading = review_heading(instruction_language(&profile, &p.request));
                let out = splice(before.as_deref(), &b, &e, &rendered, heading)?;
                assert_single_block(&out, &b, &e, &rendered)?;
                (out, Some(REVIEW_SECTION))
            }
            ".gitignore" => {
                let (b, e) = markers(IGNORE_SECTION, "hash");
                let before_text = before.as_deref().and_then(|x| std::str::from_utf8(x).ok());
                let list_local = lists_local(&p.request, before_text);
                let rendered = ignore_block(&root, list_local, before_text)?.rendered;
                let out = splice(before.as_deref(), &b, &e, &rendered, "")?;
                assert_single_block(&out, &b, &e, &rendered)?;
                (out, Some(IGNORE_SECTION))
            }
            LOCAL_FILE => (render_local(&p.request).into_bytes(), None),
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
            // #2936: REVIEW.md is an owned section too; without a baseline a later language or
            // answer change planned `conflict` for a block nobody had edited.
            Some(REVIEW_SECTION) => {
                let rendered =
                    render_review(&profile, &p.request).ok_or("review section has no answers")?;
                record_owned(&mut owned, &f.path, REVIEW_SECTION, rendered.as_bytes());
            }
            Some(IGNORE_SECTION) => {
                // The block just written is the one rendered above; its lines say whether the
                // personal file is listed, and the record keeps that (#3009 S3b).
                let written = std::str::from_utf8(&bytes).map_err(|_| "target is not UTF-8")?;
                let (b, e) = markers(IGNORE_SECTION, "hash");
                let block = find_block(written, &b, &e)?
                    .map(|(s, end, _)| &written[s..end])
                    .ok_or("managed block missing after write")?;
                record_owned_kind(
                    &mut owned,
                    &f.path,
                    IGNORE_SECTION,
                    block.as_bytes(),
                    Some(ignore_kind(ignore_has_any_local_rule(block))),
                );
            }
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

    /// C-69: the managed section names the versioning scheme when the profile records one,
    /// in the section's language; a profile without the key renders byte for byte as before,
    /// and a custom text cannot open or close a managed block.
    #[test]
    fn the_managed_section_names_the_versioning_scheme() {
        let root = tempfile::TempDir::new().unwrap();
        let base = json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude"]});
        let request = |language: &str| AgentsRequest {
            protocol_version: 1,
            language: language.into(),
            workflow_reference: "skills/moai/workflows/".into(),
            local_file: "none".into(),
            picks: Vec::new(),
        };
        let plain: Profile = serde_json::from_value(base.clone()).unwrap();
        let without = render_agents(&plain, &request("en"), root.path());
        assert!(!without.contains("Versioning"), "{without}");

        let with = |versioning: Value| -> Profile {
            let mut v = base.clone();
            v["versioning"] = versioning;
            serde_json::from_value(v).expect("the profile takes a versioning key")
        };
        let calver = with(json!({"scheme": "calver", "format": "YYYY.0M.MICRO"}));
        let text = render_agents(&calver, &request("en"), root.path());
        let line = "Versioning: CalVer (YYYY.0M.MICRO).\n";
        assert!(
            text.contains(&format!("Checks: `moai check affected`.\n{line}SPECs:")),
            "{text}"
        );
        assert_eq!(text.replace(line, ""), without);
        let korean = render_agents(&calver, &request("ko"), root.path());
        assert!(
            korean.contains(
                "검사: `moai check affected`.\n버전 규칙: CalVer (YYYY.0M.MICRO).\nSPEC:"
            ),
            "{korean}"
        );
        // The block holds at most BLOCK_MAX_LINES lines between its markers.
        assert!(text.lines().count() - 2 <= BLOCK_MAX_LINES, "{text}");

        assert!(!injects_markers(&calver, &request("en")));
        let breaking = with(json!({"scheme": "custom", "format": "x --> y"}));
        assert!(injects_markers(&breaking, &request("en")));
    }

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
            picks: Vec::new(),
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

        // en -> ko merges CLAUDE.md (its before bytes create this run's backup dir)
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
    /// unwritable, so the CLAUDE.md merge in a language change cannot store its
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
                picks: Vec::new(),
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
                    picks: Vec::new(),
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
                picks: Vec::new(),
            },
        )
        .unwrap();
        let digest = second.plan_digest.clone();
        let err = agents_apply_amended(root.path(), second, &digest, "fixture", &mut |_, _| Ok(()))
            .unwrap_err();
        fs::set_permissions(&receipts, fs::Permissions::from_mode(0o755)).unwrap();
        assert!(
            err.contains("partial apply") && err.contains("CLAUDE.md"),
            "{err}"
        );
        // The fix names the receipt gap (receipt_failure ran); the buggy path
        // returns the bare step message without it.
        assert!(err.contains("receipt not recorded"), "{err}");
    }

    /// #2125 (audit round 2): the receipt FILE records the failed file row —
    /// path=CLAUDE.md, published=false, the error, the before digest — read from
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
                picks: Vec::new(),
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
                    picks: Vec::new(),
                },
            )
            .unwrap()
            .plan_digest,
            "fixture",
            &mut |_, _| Ok(()),
        )
        .unwrap();
        let before_bytes = fs::read(root.path().join("CLAUDE.md")).unwrap();
        let expected_before = digest(&before_bytes);

        FAIL_FILE_BACKUP.with(|cell| cell.set(true));
        let second = agents_plan(
            root.path(),
            AgentsRequest {
                protocol_version: 1,
                language: "ko".into(),
                workflow_reference: "skills/moai/workflows/".into(),
                local_file: "tracked".into(),
                picks: Vec::new(),
            },
        )
        .unwrap();
        let digest = second.plan_digest.clone();
        let err = agents_apply_amended(root.path(), second, &digest, "fixture", &mut |_, _| Ok(()))
            .unwrap_err();
        FAIL_FILE_BACKUP.with(|cell| cell.set(false));
        assert!(
            err.contains("partial apply") && err.contains("CLAUDE.md"),
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
                if e["path"] == json!("CLAUDE.md") {
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
            picks: Vec::new(),
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

    /// A manifest that records only `created`/`owned` entries for the personal file.
    fn manifest_with(created: &[&str], owned: &[(&str, Option<&str>)]) -> Manifest {
        Manifest {
            schema_version: 1,
            generator_version: "t".into(),
            template_version: "t".into(),
            project_id: "demo".into(),
            profile_sha256: "sha256:p".into(),
            interview_revision: None,
            workflow_id: None,
            workflow_reference: "skills/moai/workflows/".into(),
            source_kind: "t".into(),
            sources: Vec::new(),
            canonical_sha256: None,
            projections: Vec::new(),
            owned_files: owned
                .iter()
                .map(|(path, section)| OwnedFile {
                    path: (*path).into(),
                    section_id: section.map(String::from),
                    baseline_sha256: "sha256:b".into(),
                    kind: None,
                })
                .collect(),
            created_once: created.iter().map(|s| s.to_string()).collect(),
            activation_status: "t".into(),
        }
    }

    /// Either name of the personal file counts as "created once": a project that moved to
    /// `CLAUDE.local.md` must not get an `AGENTS.local.md` made for it again, and the other way
    /// round. Any other path keeps its own record only.
    #[test]
    fn either_personal_file_name_counts_as_created_once() {
        for recorded in LOCAL_NAMES {
            let m = manifest_with(&[recorded], &[]);
            for asked in LOCAL_NAMES {
                assert!(
                    created_once(Some(&m), asked),
                    "{recorded} recorded, {asked} asked"
                );
            }
            // The earlier manifest shape: the file as an owned entry without a section.
            let old = manifest_with(&[], &[(recorded, None)]);
            for asked in LOCAL_NAMES {
                assert!(
                    created_once(Some(&old), asked),
                    "{recorded} owned, {asked} asked"
                );
            }
        }
        let other = manifest_with(&["REVIEW.md"], &[]);
        assert!(!created_once(Some(&other), "AGENTS.local.md"));
        assert!(!created_once(Some(&other), "CLAUDE.local.md"));
        assert!(created_once(Some(&other), "REVIEW.md"));
        assert!(!created_once(None, "AGENTS.local.md"));
        // A managed section of that path is an owned block, not a personal file.
        let sectioned = manifest_with(&[], &[("AGENTS.local.md", Some("moai-workflow"))]);
        assert!(!created_once(Some(&sectioned), "CLAUDE.local.md"));
    }

    /// The ignore readers: what the block lists is read under either name, but the user-rule
    /// check for the file this version generates stays with that file's name, so a rule for the
    /// other name never hides the file that is actually written.
    #[test]
    fn ignore_rules_are_read_per_name_and_across_both() {
        assert!(ignore_has_rule("CLAUDE.local.md\n"));
        assert!(ignore_has_rule("/CLAUDE.local.md\n"));
        assert!(!ignore_has_rule("AGENTS.local.md\n"));
        assert!(!ignore_has_rule("# CLAUDE.local.md\n"));
        for text in [
            "AGENTS.local.md\n",
            "/AGENTS.local.md\n",
            "CLAUDE.local.md\n",
            "  /CLAUDE.local.md  \n",
        ] {
            assert!(ignore_has_any_local_rule(text), "{text:?}");
        }
        for text in [
            "",
            "# CLAUDE.local.md\n",
            "CLAUDE.local.md.bak\n",
            "notes/CLAUDE.local.md\n",
            "*.local.md\n",
        ] {
            assert!(!ignore_has_any_local_rule(text), "{text:?}");
        }
    }

    // relative_paths_are_bounded moved to src/apply.rs with the shared helper.
}
