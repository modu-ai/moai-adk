//! The picks `move`, `leave` and `remove-originals` of `moai update` for a project that holds an
//! `AGENTS.md` or an `AGENTS.local.md` (SPEC-20261122-claude-md-project-picks,
//! SPEC-20261124-claude-md-product-next): the composite of one pick, its preview, its apply and
//! its revert.
//!
//! - `move`: the managed block goes into `CLAUDE.md` through the generation writer, the text a
//!   person wrote in `AGENTS.md` goes below it ([`Kind::Guidance`]), and the text a person wrote in
//!   `AGENTS.local.md` goes into `CLAUDE.local.md` together with its one `.gitignore` line
//!   ([`Kind::Local`]). A personal file MoAI generated holds nothing to copy.
//! - `leave`: nothing is written; Claude Code keeps reading `AGENTS.md`.
//! - `remove-originals`: once the text is whole in the CLAUDE files, the two originals are deleted,
//!   all or none, and their bytes stay in the receipt (`originals`).
//!
//! `AGENTS.md` and `AGENTS.local.md` are written by no pick and deleted only by `remove-originals`.
//! Each part runs from a plan of its own made just before it runs (a part digests the bytes the
//! part before it changed), and the one digest of a pick binds the plans of every part on the
//! project as it is before the apply. One receipt (`agents-pick-…`) lists the parts; the revert
//! undoes them in the reverse order and names a file that changed after the apply instead of
//! touching it, while the other parts still go back. A receipt that an earlier build wrote for the
//! retired pick `import` holds a part `import-line`, which the revert still undoes.
//!
//! A project is offered `move` and `leave` while the text of an original is not yet in its CLAUDE
//! file. Once `CLAUDE.md` already holds the text of `AGENTS.md` (a merged block or pasted by hand)
//! and nothing of `AGENTS.local.md` is left to move, it is offered `remove-originals` alone, or
//! nothing when a text is not whole in its CLAUDE file or a CLAUDE file imports `AGENTS.md`.

use super::{
    find_guidance, held_by_claude_md, hidden_files_warning_except, imports_agents_md,
    merge_confirmed, new_receipt_dir_named, originals, preview_of, read_checked, receipts_dir,
    revert_for, revert_legacy_import, write_private, Found, Kind, Result, GUIDANCE_TARGET,
    HELD_BY_CLAUDE_MD, LEGACY_TARGET, LOCAL, TARGET,
};
use crate::generation::{self, GenerationRequest};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
};

/// The start of the id of a composite receipt.
pub const RECEIPT_PREFIX: &str = "agents-pick-";

const MOVE_EFFECT: &str = "write the managed section into CLAUDE.md, copy the text you wrote in \
     AGENTS.md below it and the text you wrote in AGENTS.local.md into CLAUDE.local.md; AGENTS.md \
     and AGENTS.local.md stay unchanged unless you pick remove-originals after the move, the one \
     way they are deleted";
const REMOVE_EFFECT: &str = "delete the original AGENTS.md and AGENTS.local.md, whose text is \
     already in CLAUDE.md and CLAUDE.local.md; all or none are deleted, their bytes and modes stay \
     in the receipt of this pick and agents-merge-revert writes them back; a tracked AGENTS.md \
     shows up as a git deletion, and another tool that reads AGENTS.md stops seeing it";
const LEAVE_EFFECT: &str =
    "write nothing; Claude Code keeps reading AGENTS.md as it does now, and \
     these picks are offered again at the next update";
const LEFT_MESSAGE: &str = "Nothing was written. Claude Code keeps reading AGENTS.md as it does \
     now; the picks are offered again at the next update.";
/// What the warning of a first `CLAUDE.local.md` over an `AGENTS.md` says: the pick writes
/// `CLAUDE.md` first, so that warning no longer applies to its preview.
const HIDDEN_WARNING_PART: &str = "only while no CLAUDE.md";
/// The files that carry the text of the project for Claude Code, besides `AGENTS.md`.
pub(super) const CLAUDE_FILES: [&str; 3] = ["CLAUDE.md", ".claude/CLAUDE.md", "CLAUDE.local.md"];

/// One of the three picks.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum Pick {
    Move,
    Leave,
    RemoveOriginals,
}

impl Pick {
    /// The order the picks are shown in.
    pub const ALL: [Pick; 3] = [Pick::Move, Pick::Leave, Pick::RemoveOriginals];

    pub fn id(self) -> &'static str {
        match self {
            Pick::Move => "move",
            Pick::Leave => "leave",
            Pick::RemoveOriginals => "remove-originals",
        }
    }

    pub fn parse(text: &str) -> Result<Pick> {
        Pick::ALL
            .into_iter()
            .find(|p| p.id() == text)
            .ok_or_else(|| {
                "the pick must be move, leave or remove-originals; nothing was written".into()
            })
    }

    fn effect(self) -> &'static str {
        match self {
            Pick::Move => MOVE_EFFECT,
            Pick::Leave => LEAVE_EFFECT,
            Pick::RemoveOriginals => REMOVE_EFFECT,
        }
    }
}

pub(super) fn digest_of(value: &Value) -> Result<String> {
    Ok(crate::digest(
        &serde_json::to_vec(value).map_err(|_| "encode failed")?,
    ))
}

pub(super) fn exists(root: &Path, rel: &str) -> bool {
    fs::symlink_metadata(root.join(rel)).is_ok()
}

/// A regular, readable, UTF-8 text file of the project; a link, a file above the limit and text
/// that is not UTF-8 read as no text.
pub(super) fn text_of(root: &Path, rel: &str) -> Option<String> {
    let bytes = read_checked(root, rel).ok()??;
    String::from_utf8(bytes).ok()
}

/// `true` when `haystack` holds `text` as whole lines (line ends are compared as LF).
pub(super) fn holds_block(haystack: &str, text: &[u8]) -> bool {
    let Ok(text) = std::str::from_utf8(text) else {
        return false;
    };
    let needle = text.replace("\r\n", "\n");
    let needle = needle.trim_end_matches('\n');
    if needle.is_empty() {
        return false;
    }
    let hay = haystack.replace("\r\n", "\n");
    hay.match_indices(needle).any(|(at, _)| {
        let end = at + needle.len();
        (at == 0 || hay.as_bytes()[at - 1] == b'\n')
            && (end == hay.len() || hay.as_bytes()[end] == b'\n')
    })
}

/// What the project holds, read once. Writes nothing.
struct State {
    root: PathBuf,
    has_agents: bool,
    has_local: bool,
    guidance: Option<Value>,
    local: Option<Value>,
}

impl State {
    fn read(root: &Path) -> Result<State> {
        let has_agents = exists(root, LEGACY_TARGET);
        let has_local = exists(root, LOCAL);
        // `offer` has asked about a Go-generation `CLAUDE.md` before it reads the state.
        let plan = |kind, wanted: bool| wanted.then(|| preview_of(kind, root, false)).transpose();
        Ok(State {
            root: root.to_path_buf(),
            has_agents,
            has_local,
            guidance: plan(Kind::Guidance, has_agents)?,
            local: plan(Kind::Local, has_local)?,
        })
    }

    fn action(plan: &Option<Value>) -> Option<&str> {
        plan.as_ref().and_then(|p| p["action"].as_str())
    }

    /// Why `move` is not needed for a project that holds the files, or `None` when it is. A plan
    /// action `unchanged` only says nothing is left to move, not that the text is in a CLAUDE file
    /// (an earlier generation merged personal text into `AGENTS.md`): the offer to delete the
    /// originals asks [`originals::plan`], which reads the CLAUDE files themselves.
    fn settled(&self) -> Option<String> {
        let claude = text_of(&self.root, GUIDANCE_TARGET);
        if claude.as_deref().is_some_and(imports_agents_md) {
            return Some("CLAUDE.md already imports AGENTS.md".into());
        }
        let agents_done = !self.has_agents
            || match find_guidance(&self.root) {
                Ok(Found::Source(s)) => claude.as_deref().is_some_and(|c| holds_block(c, &s.text)),
                // `AGENTS.md` holds nothing of the person's: the block is all a pick would write.
                Ok(Found::Nothing(_)) => claude.as_deref().is_some_and(|c| {
                    c.contains(&crate::agents::markers(crate::agents::AGENTS_SECTION, "html").0)
                }),
                _ => false,
            };
        let local_done =
            !self.has_local || matches!(State::action(&self.local), Some("none" | "unchanged"));
        (agents_done && local_done).then(|| {
            "CLAUDE.md already holds the text of AGENTS.md and nothing of AGENTS.local.md is left \
             to move"
                .into()
        })
    }

    /// REQ-8: Claude Code reads only the CLAUDE files once one exists, so a regular one that
    /// neither imports `AGENTS.md` nor holds its text hides it.
    fn shadow_warnings(&self) -> Vec<String> {
        let Ok(Found::Source(source)) = find_guidance(&self.root) else {
            return Vec::new();
        };
        CLAUDE_FILES
            .iter()
            .filter_map(|rel| {
                let text = text_of(&self.root, rel)?;
                (!imports_agents_md(&text) && !holds_block(&text, &source.text)).then(|| {
                    format!(
                        "Claude Code reads only the CLAUDE files in this project, so it does not \
                         read the text of AGENTS.md: {rel} neither imports AGENTS.md nor holds its \
                         text"
                    )
                })
            })
            .collect()
    }
}

/// One pick as the preview shows it.
struct Candidate {
    pick: Pick,
    blocked: Vec<Value>,
    plan_digest: String,
    writes: Vec<String>,
    /// The files the pick deletes (`remove-originals` only).
    deletes: Vec<String>,
    parts: Vec<Value>,
    warnings: Vec<String>,
}

impl Candidate {
    fn state(&self) -> &'static str {
        if self.blocked.is_empty() {
            "ready"
        } else {
            "blocked"
        }
    }

    fn to_json(&self) -> Value {
        json!({
            "id": self.pick.id(),
            "recommended": self.pick == Pick::Move,
            "state": self.state(),
            "blocked": self.blocked,
            "effect": self.pick.effect(),
            "plan_digest": self.plan_digest,
            "writes": self.writes,
            "deletes": self.deletes,
            "parts": self.parts,
            "warnings": self.warnings,
        })
    }
}

/// The part names a receipt and a preview use.
const BLOCK: &str = "block";
const AGENTS_TEXT: &str = "agents-text";
const PERSONAL_TEXT: &str = "personal-text";
/// The part of a receipt that an earlier build wrote for the retired pick `import`: it is only
/// ever undone.
const LEGACY_IMPORT_PART: &str = "import-line";

/// `move` or `leave` as the preview shows it; `remove-originals` has its own [`removal_candidate`].
fn candidate(st: &State, request: &GenerationRequest, pick: Pick) -> Result<Candidate> {
    let mut c = Candidate {
        pick,
        blocked: Vec::new(),
        plan_digest: String::new(),
        writes: Vec::new(),
        deletes: Vec::new(),
        parts: Vec::new(),
        warnings: Vec::new(),
    };
    let mut bound: Vec<Value> = Vec::new();
    if pick == Pick::Leave {
        for plan in [&st.guidance, &st.local].into_iter().flatten() {
            bound.push(plan["plan_digest"].clone());
        }
    } else {
        block_part(st, request, &mut c, &mut bound);
        if st.has_agents {
            engine_part(AGENTS_TEXT, &st.guidance, &mut c, &mut bound);
        }
        if st.has_local {
            engine_part(PERSONAL_TEXT, &st.local, &mut c, &mut bound);
        }
    }
    let mut seen = std::collections::HashSet::new();
    c.writes.retain(|w| seen.insert(w.clone()));
    c.plan_digest = digest_of(&json!({
        "pick": pick.id(),
        "project": st.root.display().to_string(),
        "blocked": c.blocked,
        "bound": bound,
    }))?;
    Ok(c)
}

/// The generation writer's part: the plan of the whole generation request with the pick in it.
fn block_part(st: &State, request: &GenerationRequest, c: &mut Candidate, bound: &mut Vec<Value>) {
    let mut picked = request.clone();
    picked.picks = vec![c.pick.id().to_string()];
    // The writer plans the block for `move` only for this composite, which adds the text of
    // AGENTS.md in the same receipt.
    let _carried = crate::agents::CarriedByPicks::hold();
    let plan = match generation::generation_plan(&st.root, picked) {
        Ok(plan) => plan,
        Err(e) => {
            c.blocked.push(json!({"part": BLOCK, "reason": e}));
            return;
        }
    };
    for f in plan.files.iter().filter(|f| f.action == "conflict") {
        c.blocked.push(json!({
            "part": BLOCK,
            "reason": format!("{}: {}", f.path, f.detail.as_deref().unwrap_or("conflict")),
        }));
    }
    let rows: Vec<Value> = plan
        .files
        .iter()
        .filter(|f| f.action != "unchanged")
        .map(|f| json!({"path": f.path, "action": f.action}))
        .collect();
    c.writes.extend(
        plan.files
            .iter()
            .filter(|f| f.action != "unchanged")
            .map(|f| f.path.clone()),
    );
    c.parts
        .push(json!({"part": BLOCK, "files": rows, "plan_digest": plan.plan_digest}));
    bound.push(json!(plan.plan_digest));
}

/// The part of one merge kind, from the plan the engine made on the project as it is now.
fn engine_part(name: &str, plan: &Option<Value>, c: &mut Candidate, bound: &mut Vec<Value>) {
    let Some(plan) = plan else {
        return;
    };
    let reason = plan["reason"].as_str().unwrap_or_default();
    match plan["action"].as_str() {
        Some("merge") => {
            if let Some(path) = plan["target"]["path"].as_str() {
                c.writes.push(path.to_string());
            }
            // The personal file's line, or the receipt folder's line when no managed block has it.
            if plan["gitignore"]["action"] == "add"
                || plan["gitignore"]["receipts"]["action"] == "add"
            {
                c.writes.push(".gitignore".into());
            }
            // The pick writes `CLAUDE.md` before the personal file, so the warning that a first
            // CLAUDE file hides `AGENTS.md` does not apply to it.
            c.warnings.extend(
                plan["warnings"]
                    .as_array()
                    .into_iter()
                    .flatten()
                    .filter_map(Value::as_str)
                    .filter(|w| !w.contains(HIDDEN_WARNING_PART))
                    .map(String::from),
            );
            c.parts.push(json!({
                "part": name,
                "action": "merge",
                "reason": reason,
                "source": plan["source"],
                "target": plan["target"],
                "gitignore": plan["gitignore"],
                "body": plan["body"],
                "commit_notice": plan["commit_notice"],
                "line_endings": plan["line_endings"],
                "plan_digest": plan["plan_digest"],
            }));
        }
        Some(action @ ("none" | "unchanged")) => c.parts.push(json!({
            "part": name, "action": action, "reason": reason, "plan_digest": plan["plan_digest"],
        })),
        _ => c.blocked.push(json!({"part": name, "reason": reason})),
    }
    bound.push(plan["plan_digest"].clone());
}

/// The `remove-originals` choice of the preview, from a plan that found every text in its CLAUDE
/// file.
fn removal_choice(root: &Path, plan: &originals::Plan) -> Result<Value> {
    Ok(json!({
        "id": Pick::RemoveOriginals.id(),
        "recommended": false,
        "state": "ready",
        "blocked": [],
        "effect": Pick::RemoveOriginals.effect(),
        "plan_digest": plan.digest(root)?,
        "writes": plan.writes(),
        "deletes": plan.deletes(),
        "parts": [plan.part_json()],
        "warnings": plan.warnings,
    }))
}

/// The preview of the picks of a project. Writes nothing. `request` builds the generation request
/// the writer plans with; it is called only when `move` and `leave` are offered.
///
/// `offer` is `claude-md` (the root `CLAUDE.md` is a Go-generation instruction file, which the
/// `claude-md` step of the update preview decides about), `none` (no `AGENTS.md` and no
/// `AGENTS.local.md`), `settled` (nothing left to pick), `unavailable` (the request cannot be
/// built, so the writer cannot be planned), `originals` (the text is in the CLAUDE files;
/// `choices` holds `remove-originals` alone) or `picks`, which carries `choices` in the order
/// `move`, `leave` and the shadow `warnings`.
pub fn offer(root: &Path, request: impl FnOnce() -> Result<GenerationRequest>) -> Result<Value> {
    let project = root.display().to_string();
    // The first CLAUDE file of the project hides `AGENTS.md` and `.claude/AGENTS.md` from Claude
    // Code. A project that holds only `.claude/AGENTS.md` is warned and offered no pick: the
    // picks move the root files, and moving that file is a separate feature. `move` carries the
    // root AGENTS.md into the new CLAUDE.md, so it stays read and is not named; only
    // `.claude/AGENTS.md` (nothing carries it) is.
    let hidden: Vec<String> = hidden_files_warning_except(root, GUIDANCE_TARGET, &[LEGACY_TARGET])
        .into_iter()
        .collect();
    let other = |offer: &str, reason: &str| -> Result<Value> {
        Ok(json!({
            "project": project, "offer": offer, "reason": reason, "warnings": hidden,
            "choices": [],
        }))
    };
    // Before anything reads the files: an import line of that file is no finished pick.
    if held_by_claude_md(root) {
        return other("claude-md", HELD_BY_CLAUDE_MD);
    }
    if !exists(root, LEGACY_TARGET) && !exists(root, LOCAL) {
        return other(
            "none",
            "no AGENTS.md and no AGENTS.local.md in this project",
        );
    }
    let st = State::read(root)?;
    if let Some(reason) = st.settled() {
        // Nothing is left to move. The originals may go when every text of theirs is whole in its
        // CLAUDE file; otherwise nothing is offered.
        return match originals::plan(root) {
            originals::Planned::Ready(plan) => {
                let choice = removal_choice(root, &plan)?;
                Ok(json!({
                    "project": project,
                    "offer": "originals",
                    "reason": "the text of the originals is in the CLAUDE files; remove-originals \
                               deletes AGENTS.md and AGENTS.local.md and keeps their bytes in the \
                               receipt",
                    "warnings": hidden.into_iter().chain(plan.warnings.iter().cloned()).collect::<Vec<_>>(),
                    "choices": [choice],
                }))
            }
            originals::Planned::Refused(_) => other("settled", &reason),
        };
    }
    let request = match request() {
        Ok(request) => request,
        Err(reason) => return other("unavailable", &reason),
    };
    let choices: Vec<Value> = [Pick::Move, Pick::Leave]
        .into_iter()
        .map(|p| candidate(&st, &request, p).map(|c| c.to_json()))
        .collect::<Result<_>>()?;
    Ok(json!({
        "project": project,
        "offer": "picks",
        "reason": "the project holds AGENTS.md or AGENTS.local.md; pick move or leave",
        "warnings": hidden.into_iter().chain(st.shadow_warnings()).collect::<Vec<_>>(),
        "choices": choices,
    }))
}

/// The preview of one pick. `state` is `ready`, `blocked` (the reasons are in `blocked`, and
/// naming the file) or `none` (no pick is offered, so there is nothing to apply). Writes nothing.
fn plan_with(
    root: &Path,
    pick: Pick,
    request: impl FnOnce() -> Result<GenerationRequest>,
) -> Result<Value> {
    let offered = offer(root, request)?;
    let none = |state: &str, reason: &Value| -> Result<Value> {
        Ok(json!({
            "project": offered["project"],
            "pick": pick.id(),
            "state": state,
            "offer": offered["offer"],
            "reason": reason,
            "blocked": if state == "blocked" { json!([{"part": "request", "reason": reason}]) } else { json!([]) },
            "plan_digest": digest_of(&json!({
                "pick": pick.id(), "offer": offered["offer"], "reason": reason,
                "project": offered["project"],
            }))?,
            "shadow_warnings": offered["warnings"],
        }))
    };
    // Blocked rather than `none`, which an apply answers as nothing to do.
    if offered["offer"] == "claude-md" {
        return none("blocked", &offered["reason"]);
    }
    // The deletion is available only when every text is whole in its CLAUDE file: a direct call
    // before that is refused with the reason, never answered as nothing to do.
    if pick == Pick::RemoveOriginals && offered["offer"] != "originals" {
        let reason = match originals::plan(root) {
            originals::Planned::Refused(reason) => reason,
            originals::Planned::Ready(_) => {
                "text of an original is still to move; pick move first".to_string()
            }
        };
        return none("blocked", &json!(reason));
    }
    match offered["offer"].as_str() {
        Some("picks") => {}
        Some("originals") if pick == Pick::RemoveOriginals => {}
        Some("unavailable") => return none("blocked", &offered["reason"]),
        _ => return none("none", &offered["reason"]),
    }
    let Some(mut chosen) = offered["choices"]
        .as_array()
        .into_iter()
        .flatten()
        .find(|c| c["id"] == pick.id())
        .cloned()
    else {
        return none(
            "blocked",
            &json!(format!("the pick {} is not offered", pick.id())),
        );
    };
    chosen["pick"] = json!(pick.id());
    chosen["project"] = offered["project"].clone();
    chosen["offer"] = json!("picks");
    chosen["shadow_warnings"] = offered["warnings"].clone();
    Ok(chosen)
}

/// The read-only preview of one pick of the project at `project`.
pub fn plan(project: &Path, pick: &str) -> Result<Value> {
    let pick = Pick::parse(pick)?;
    let root = fs::canonicalize(project).map_err(|_| crate::contract::PROJECT_PATH_UNAVAILABLE)?;
    plan_with(&root, pick, || {
        crate::project_update::rebuilt_request(&root)
    })
}

#[derive(Serialize, Deserialize, Default)]
struct PartRecord {
    part: String,
    /// The receipts of the engines the part ran, in the order they are undone.
    receipts: Vec<String>,
    #[serde(default)]
    reverted: bool,
    /// The originals a `remove-originals` part deleted, in the order it deleted them.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    files: Vec<originals::FileRecord>,
    /// The `.gitignore` line a `remove-originals` part added.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    ignore: Option<originals::IgnoreRecord>,
}

#[derive(Serialize, Deserialize)]
struct PickReceipt {
    schema_version: u32,
    id: String,
    at: String,
    pick: String,
    plan_digest: String,
    permission_reference: String,
    /// `pending`, `applied`, `partial` (a part failed), `failed` (nothing was written),
    /// `reverted` or `reverted_partly`.
    status: String,
    parts: Vec<PartRecord>,
    /// The files a revert left alone, named.
    #[serde(default)]
    kept: Vec<Value>,
}

fn save(dir: &Path, r: &PickReceipt) -> Result<()> {
    let mut bytes = serde_json::to_vec_pretty(r).map_err(|_| "encode failed")?;
    bytes.push(b'\n');
    let path = dir.join("receipt.json");
    match crate::claude::read_existing(&path)? {
        None => write_private(&path, &bytes),
        Some(before) => crate::apply::publish_replace(&path, Some(&before), &bytes),
    }
}

/// `"a, b"` of the reasons a pick is refused, each naming its part and file.
fn reasons(blocked: &Value) -> String {
    blocked
        .as_array()
        .into_iter()
        .flatten()
        .map(|b| {
            format!(
                "{}: {}",
                b["part"].as_str().unwrap_or_default(),
                b["reason"].as_str().unwrap_or_default()
            )
        })
        .collect::<Vec<_>>()
        .join("; ")
}

fn paths_of(value: &Value) -> impl Iterator<Item = String> + '_ {
    value
        .as_array()
        .into_iter()
        .flatten()
        .filter_map(Value::as_str)
        .map(String::from)
}

/// Applies one pick after the user confirmed the plan whose digest is `plan_digest`.
pub fn apply(project: &Path, pick: &str, plan_digest: &str, permission: &str) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("a permission reference is required; nothing was written".into());
    }
    let pick = Pick::parse(pick)?;
    let root = fs::canonicalize(project).map_err(|_| crate::contract::PROJECT_PATH_UNAVAILABLE)?;
    // Asked before the digest, so a digest of an earlier preview gets the same answer.
    if held_by_claude_md(&root) {
        return Err(format!(
            "the pick {} is refused: {HELD_BY_CLAUDE_MD}; nothing was written",
            pick.id()
        ));
    }
    let request = crate::project_update::rebuilt_request(&root);
    let planned = plan_with(&root, pick, || request.clone())?;
    if planned["plan_digest"].as_str() != Some(plan_digest) {
        return Err("stale or altered plan; re-preview required".into());
    }
    match planned["state"].as_str() {
        Some("none") => {
            return Ok(json!({"state": "unchanged", "writes": [], "reason": planned["reason"]}))
        }
        Some("ready") => {}
        _ => {
            return Err(format!(
                "the pick {} is refused: {}; nothing was written",
                pick.id(),
                reasons(&planned["blocked"])
            ))
        }
    }
    if pick == Pick::Leave {
        return Ok(json!({
            "state": "left",
            "pick": "leave",
            "writes": [],
            "message": LEFT_MESSAGE,
            "warnings": planned["shadow_warnings"],
        }));
    }
    // `remove-originals` plans nothing with the generation writer.
    let request = if pick == Pick::RemoveOriginals {
        None
    } else {
        Some(request?)
    };
    let (id, dir) = new_receipt_dir_named(RECEIPT_PREFIX, &root)?;
    let mut receipt = PickReceipt {
        schema_version: 1,
        id: id.clone(),
        at: crate::baseline::utc_now(),
        pick: pick.id().to_string(),
        plan_digest: plan_digest.to_string(),
        permission_reference: permission.to_string(),
        status: "pending".into(),
        parts: Vec::new(),
        kept: Vec::new(),
    };
    save(&dir, &receipt)?;
    let mut writes: Vec<String> = Vec::new();
    let mut deletes: Vec<String> = Vec::new();
    let mut parts_out: Vec<Value> = Vec::new();
    let outcome = match &request {
        Some(request) => run_parts(
            &root,
            pick,
            request,
            permission,
            &mut receipt,
            &dir,
            &mut writes,
            &mut parts_out,
        ),
        None => run_removal(
            &root,
            plan_digest,
            &mut receipt,
            &dir,
            &mut writes,
            &mut deletes,
            &mut parts_out,
            &mut |_| Ok(()),
        ),
    };
    if let Err(e) = outcome {
        return Err(fail_with_rollback(&root, &dir, &mut receipt, e, permission));
    }
    receipt.status = "applied".into();
    save(&dir, &receipt)?;
    let mut seen = std::collections::HashSet::new();
    writes.retain(|w| seen.insert(w.clone()));
    let originals: Vec<&str> = [LEGACY_TARGET, LOCAL]
        .into_iter()
        .filter(|f| exists(&root, f))
        .collect();
    Ok(json!({
        "state": "applied",
        "pick": pick.id(),
        "project": root.display().to_string(),
        "receipt": id,
        "writes": writes,
        "deletes": deletes,
        "parts": parts_out,
        "unchanged": originals,
        "next_actions": [format!(
            "to undo: moai agents-merge-revert --receipt {id} --permission-reference <ref> <project>"
        )],
    }))
}

/// A pick is all or nothing: when a part fails after others wrote, the parts that did write go
/// back through the same receipt. The message says what is left, if anything.
fn fail_with_rollback(
    root: &Path,
    dir: &Path,
    receipt: &mut PickReceipt,
    cause: String,
    permission: &str,
) -> String {
    let id = receipt.id.clone();
    let applied: Vec<String> = receipt.parts.iter().map(|p| p.part.clone()).collect();
    receipt.status = if applied.is_empty() {
        "failed"
    } else {
        "partial"
    }
    .into();
    let saved = save(dir, receipt);
    if applied.is_empty() {
        let mut message = cause;
        if let Err(se) = saved {
            message.push_str(&format!("; receipt not recorded: {se}"));
        }
        return message;
    }
    let rolled = match saved {
        Ok(()) => revert(root, &id, permission),
        Err(se) => Err(format!("receipt not recorded: {se}")),
    };
    match rolled {
        Ok(out) if out["state"] == "reverted" => format!(
            "{cause}; the parts written before it [{}] were rolled back, nothing is left changed \
             (receipt {id})",
            applied.join(", ")
        ),
        Ok(out) => format!(
            "partial apply: applied [{}]; {cause}; the rollback kept {}; undo the rest with \
             agents-merge-revert --receipt {id}",
            applied.join(", "),
            out["kept"]
                .as_array()
                .into_iter()
                .flatten()
                .filter_map(|k| k["path"].as_str())
                .collect::<Vec<_>>()
                .join(", ")
        ),
        Err(re) => format!(
            "partial apply: applied [{}]; {cause}; the rollback failed: {re}; undo them with \
             agents-merge-revert --receipt {id}",
            applied.join(", ")
        ),
    }
}

/// The parts of a `move`, each from a plan made now. A part that fails after
/// another has written leaves the receipt to say what to undo.
#[allow(clippy::too_many_arguments)]
fn run_parts(
    root: &Path,
    pick: Pick,
    request: &GenerationRequest,
    permission: &str,
    receipt: &mut PickReceipt,
    dir: &Path,
    writes: &mut Vec<String>,
    parts_out: &mut Vec<Value>,
) -> Result<()> {
    // 1. The managed block, through the generation writer with the pick in its request.
    let mut picked = request.clone();
    picked.picks = vec![pick.id().to_string()];
    let _carried = crate::agents::CarriedByPicks::hold();
    let plan = generation::generation_plan(root, picked)?;
    if let Some(f) = plan.files.iter().find(|f| f.action == "conflict") {
        return Err(format!(
            "failed at {BLOCK}: conflict in {}: {}",
            f.path,
            f.detail.as_deref().unwrap_or("review required")
        ));
    }
    if plan.files.iter().any(|f| f.action != "unchanged") {
        let out = generation::generation_apply(root, plan.clone(), &plan.plan_digest, permission)
            .map_err(|e| format!("failed at {BLOCK}: {e}"))?;
        let written: Vec<String> = paths_of(&out["files_created"])
            .chain(paths_of(&out["files_merged"]))
            .collect();
        let receipts: Vec<String> = ["agents_receipt", "receipt"]
            .iter()
            .filter_map(|k| out[*k].as_str().map(String::from))
            .collect();
        writes.append(&mut written.clone());
        receipt.parts.push(PartRecord {
            part: BLOCK.into(),
            receipts: receipts.clone(),
            ..PartRecord::default()
        });
        save(dir, receipt)?;
        parts_out.push(json!({"part": BLOCK, "writes": written, "receipts": receipts}));
    }
    // 2. and 3. The text merges, each from a plan made after the part before it wrote.
    let merges: &[(Kind, &str, &str)] = &[
        (Kind::Guidance, AGENTS_TEXT, LEGACY_TARGET),
        (Kind::Local, PERSONAL_TEXT, LOCAL),
    ];
    for (kind, name, source) in merges {
        if !exists(root, source) {
            continue;
        }
        let plan = preview_of(*kind, root, false).map_err(|e| format!("failed at {name}: {e}"))?;
        match plan["action"].as_str() {
            Some("merge") => {}
            Some("none" | "unchanged") => continue,
            _ => {
                return Err(format!(
                    "failed at {name}: {}",
                    plan["reason"].as_str().unwrap_or("the plan is refused")
                ))
            }
        }
        let digest = plan["plan_digest"].as_str().unwrap_or_default();
        // The token of a plan that warns about a first CLAUDE file is never passed here: the
        // block is written first, so a plan that still carries one is refused.
        let out = merge_confirmed(*kind, root, digest, permission, None)
            .map_err(|e| format!("failed at {name}: {e}"))?;
        let Some(engine_receipt) = out["receipt"].as_str() else {
            continue;
        };
        let written: Vec<String> = paths_of(&out["writes"]).collect();
        writes.extend(written.clone());
        receipt.parts.push(PartRecord {
            part: (*name).into(),
            receipts: vec![engine_receipt.to_string()],
            ..PartRecord::default()
        });
        save(dir, receipt)?;
        parts_out.push(json!({"part": name, "writes": written, "receipts": [engine_receipt]}));
    }
    Ok(())
}

/// The part of `remove-originals`, from a plan made now. The order is the one that keeps the
/// project whole at every step: compare, write the private copies, name the files in the receipt,
/// then write `.gitignore` and delete. A failure after the receipt names the files goes back
/// through the same receipt. `after_each` runs after each deletion with its index.
#[allow(clippy::too_many_arguments)]
fn run_removal(
    root: &Path,
    plan_digest: &str,
    receipt: &mut PickReceipt,
    dir: &Path,
    writes: &mut Vec<String>,
    deletes: &mut Vec<String>,
    parts_out: &mut Vec<Value>,
    after_each: &mut dyn FnMut(usize) -> Result<()>,
) -> Result<()> {
    let originals::Planned::Ready(plan) = originals::plan(root) else {
        return Err(format!(
            "failed at {}: the originals are no longer ready",
            originals::PART
        ));
    };
    // The plan this apply was confirmed on, read again just before the first file is touched.
    if plan.digest(root)? != plan_digest {
        return Err("stale or altered plan; re-preview required".into());
    }
    let (files, ignore) = originals::stage(root, dir, &plan)
        .map_err(|e| format!("failed at {}: {e}", originals::PART))?;
    receipt.parts.push(PartRecord {
        part: originals::PART.into(),
        files,
        ignore,
        ..PartRecord::default()
    });
    save(dir, receipt)?;
    let written = originals::commit(root, &plan, after_each)
        .map_err(|e| format!("failed at {}: {e}", originals::PART))?;
    writes.extend(written.clone());
    deletes.extend(plan.deletes());
    parts_out.push(json!({
        "part": originals::PART, "writes": written, "deletes": plan.deletes(),
    }));
    Ok(())
}

/// Undoes one pick: the parts in the reverse order of the apply. A file that changed after the
/// apply is named under `kept` and left as it is; the other parts still go back.
pub fn revert(project: &Path, receipt_id: &str, permission: &str) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("a permission reference is required; nothing was written".into());
    }
    if !(receipt_id.starts_with(RECEIPT_PREFIX)
        && receipt_id.len() <= 64
        && receipt_id
            .bytes()
            .all(|b| b.is_ascii_alphanumeric() || b == b'-'))
    {
        return Err("unknown receipt id".into());
    }
    let root = fs::canonicalize(project).map_err(|_| "project directory unavailable")?;
    let dir = receipts_dir(&root).join(receipt_id);
    for d in [
        crate::state_root::path(&root, ""),
        receipts_dir(&root),
        dir.clone(),
    ] {
        crate::claude::refuse_symlink_dir(&d)?;
    }
    let bytes =
        crate::claude::read_existing(&dir.join("receipt.json"))?.ok_or("receipt not found")?;
    let mut receipt: PickReceipt =
        serde_json::from_slice(&bytes).map_err(|_| "receipt is unreadable")?;
    match receipt.status.as_str() {
        "applied" | "partial" | "pending" => {}
        "reverted" => return Ok(json!({"state": "unchanged", "writes": []})),
        "reverted_partly" => {
            return Ok(json!({"state": "partial", "writes": [], "kept": receipt.kept}))
        }
        _ => return Err("that apply wrote nothing; there is nothing to undo".into()),
    }
    let mut writes: Vec<String> = Vec::new();
    let mut kept: Vec<Value> = Vec::new();
    for part in receipt.parts.iter_mut().rev().filter(|p| !p.reverted) {
        if part.part == originals::PART {
            let (written, not_restored) =
                originals::revert(&root, &dir, receipt_id, &part.files, part.ignore.as_ref());
            writes.extend(written);
            kept.extend(not_restored);
            part.reverted = true;
            continue;
        }
        // The part `import-line` is what the retired pick `import` wrote: only its undo is left.
        let engine = match part.part.as_str() {
            AGENTS_TEXT | LEGACY_IMPORT_PART => Some(GUIDANCE_TARGET),
            PERSONAL_TEXT => Some(TARGET),
            _ => None,
        };
        for id in &part.receipts {
            match engine {
                Some(target) => {
                    let undone = match part.part.as_str() {
                        AGENTS_TEXT => revert_for(Kind::Guidance, &root, id, permission),
                        PERSONAL_TEXT => revert_for(Kind::Local, &root, id, permission),
                        _ => revert_legacy_import(&root, id, permission),
                    };
                    match undone {
                        Ok(out) => writes.extend(paths_of(&out["writes"])),
                        Err(reason) => kept.push(json!({
                            "part": part.part, "path": target, "receipt": id, "reason": reason,
                        })),
                    }
                }
                None => match crate::apply::recover(&root, id, permission) {
                    Ok(out) => {
                        writes.extend(paths_of(&out["reverted"]));
                        for skipped in out["skipped"].as_array().into_iter().flatten() {
                            kept.push(json!({
                                "part": part.part, "path": skipped["path"], "receipt": id,
                                "reason": skipped["reason"],
                            }));
                        }
                    }
                    Err(reason) => kept.push(json!({
                        "part": part.part, "path": GUIDANCE_TARGET, "receipt": id,
                        "reason": reason,
                    })),
                },
            }
        }
        part.reverted = true;
    }
    receipt.status = if kept.is_empty() {
        "reverted"
    } else {
        "reverted_partly"
    }
    .into();
    receipt.kept = kept.clone();
    save(&dir, &receipt)?;
    writes.dedup();
    Ok(json!({
        "state": if kept.is_empty() { "reverted" } else { "partial" },
        "project": root.display().to_string(),
        "receipt": receipt_id,
        "writes": writes,
        "kept": kept,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    const PERMISSION: &str = "fixture-approved";

    fn receipt_with(id: &str, parts: Vec<PartRecord>) -> PickReceipt {
        PickReceipt {
            schema_version: 1,
            id: id.to_string(),
            at: "2026-11-22T00:00:00Z".into(),
            pick: "move".into(),
            plan_digest: "digest".into(),
            permission_reference: PERMISSION.into(),
            status: "pending".into(),
            parts,
            kept: Vec::new(),
        }
    }

    /// A project folder with an empty receipt folder for `id`.
    fn project(id: &str) -> (tempfile::TempDir, PathBuf, PathBuf) {
        let tmp = tempfile::TempDir::new().unwrap();
        let root = fs::canonicalize(tmp.path()).unwrap();
        let dir = receipts_dir(&root).join(id);
        fs::create_dir_all(&dir).unwrap();
        (tmp, root, dir)
    }

    fn status_on_disk(dir: &Path) -> String {
        let bytes = fs::read(dir.join("receipt.json")).unwrap();
        let v: Value = serde_json::from_slice(&bytes).unwrap();
        v["status"].as_str().unwrap().to_string()
    }

    #[test]
    fn failing_before_any_part_wrote_leaves_a_failed_receipt_and_the_cause_unchanged() {
        let id = format!("{RECEIPT_PREFIX}1-0");
        let (_tmp, root, dir) = project(&id);
        let mut receipt = receipt_with(&id, Vec::new());
        let message = fail_with_rollback(
            &root,
            &dir,
            &mut receipt,
            "failed at block: x".into(),
            PERMISSION,
        );
        assert_eq!(message, "failed at block: x");
        assert_eq!(status_on_disk(&dir), "failed");
    }

    #[test]
    fn failing_after_a_part_wrote_rolls_it_back_through_the_same_receipt() {
        let id = format!("{RECEIPT_PREFIX}2-0");
        let (_tmp, root, dir) = project(&id);
        // The block part with no engine receipts left to undo: the rollback has nothing to keep.
        let part = PartRecord {
            part: BLOCK.into(),
            ..PartRecord::default()
        };
        let mut receipt = receipt_with(&id, vec![part]);
        let message = fail_with_rollback(
            &root,
            &dir,
            &mut receipt,
            "failed at agents-text: x".into(),
            PERMISSION,
        );
        assert!(message.contains("failed at agents-text: x"), "{message}");
        assert!(
            message.contains("were rolled back, nothing is left changed"),
            "{message}"
        );
        assert!(!message.contains("partial apply"), "{message}");
        assert_eq!(status_on_disk(&dir), "reverted");
    }

    const TEAM: &str = "# Team rules\n\nWrite tests first.\n";
    const MINE: &str = "My own notes\nPrefer tabs.\n";

    /// A project whose CLAUDE files hold the text of both originals, with a receipt folder for
    /// `id` and the receipt saved as the apply saves it before its first write.
    fn removal_project(id: &str) -> (tempfile::TempDir, PathBuf, PathBuf, PickReceipt) {
        let (tmp, root, dir) = project(id);
        for (rel, text) in [
            ("AGENTS.md", TEAM),
            ("CLAUDE.md", TEAM),
            ("AGENTS.local.md", MINE),
            ("CLAUDE.local.md", MINE),
        ] {
            fs::write(root.join(rel), text).unwrap();
        }
        let mut receipt = receipt_with(id, Vec::new());
        receipt.pick = "remove-originals".into();
        save(&dir, &receipt).unwrap();
        (tmp, root, dir, receipt)
    }

    /// Runs the removal with a failure injected after the deletion at `fail_at`.
    fn failing_removal(
        root: &Path,
        dir: &Path,
        receipt: &mut PickReceipt,
        fail_at: usize,
    ) -> String {
        let originals::Planned::Ready(plan) = originals::plan(root) else {
            panic!("the originals are not ready");
        };
        let digest = plan.digest(root).unwrap();
        let (mut writes, mut deletes, mut parts) = (Vec::new(), Vec::new(), Vec::new());
        run_removal(
            root,
            &digest,
            receipt,
            dir,
            &mut writes,
            &mut deletes,
            &mut parts,
            &mut |i| {
                if i == fail_at {
                    Err("disk full".into())
                } else {
                    Ok(())
                }
            },
        )
        .unwrap_err()
    }

    #[test]
    fn a_failure_after_the_first_original_was_deleted_puts_it_back_through_the_same_receipt() {
        let id = format!("{RECEIPT_PREFIX}4-0");
        let (_tmp, root, dir, mut receipt) = removal_project(&id);
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            fs::set_permissions(root.join("AGENTS.md"), fs::Permissions::from_mode(0o755)).unwrap();
        }
        let cause = failing_removal(&root, &dir, &mut receipt, 0);
        assert!(cause.contains("disk full"), "{cause}");
        assert!(
            !root.join("AGENTS.md").exists(),
            "the first original was deleted before the failure"
        );
        let message = fail_with_rollback(&root, &dir, &mut receipt, cause, PERMISSION);
        assert!(
            message.contains("were rolled back, nothing is left changed"),
            "{message}"
        );
        assert_eq!(fs::read_to_string(root.join("AGENTS.md")).unwrap(), TEAM);
        assert_eq!(
            fs::read_to_string(root.join("AGENTS.local.md")).unwrap(),
            MINE
        );
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            let mode = fs::metadata(root.join("AGENTS.md"))
                .unwrap()
                .permissions()
                .mode();
            assert_eq!(mode & 0o7777, 0o755);
        }
        assert_eq!(status_on_disk(&dir), "reverted");
    }

    #[test]
    fn a_rollback_names_the_original_it_could_not_put_back() {
        let id = format!("{RECEIPT_PREFIX}5-0");
        let (_tmp, root, dir, mut receipt) = removal_project(&id);
        let cause = failing_removal(&root, &dir, &mut receipt, 0);
        // The private copy was altered since: it must not be written back.
        fs::write(dir.join("originals/AGENTS.md"), "tampered\n").unwrap();
        let message = fail_with_rollback(&root, &dir, &mut receipt, cause, PERMISSION);
        assert!(
            message.starts_with("partial apply: applied [originals]"),
            "{message}"
        );
        assert!(message.contains("the rollback kept AGENTS.md"), "{message}");
        assert!(!root.join("AGENTS.md").exists());
        assert_eq!(status_on_disk(&dir), "reverted_partly");
    }

    #[test]
    fn a_stale_plan_digest_deletes_nothing_and_records_no_part() {
        let id = format!("{RECEIPT_PREFIX}6-0");
        let (_tmp, root, dir, mut receipt) = removal_project(&id);
        let (mut writes, mut deletes, mut parts) = (Vec::new(), Vec::new(), Vec::new());
        let err = run_removal(
            &root,
            "sha256:not-the-plan",
            &mut receipt,
            &dir,
            &mut writes,
            &mut deletes,
            &mut parts,
            &mut |_| Ok(()),
        )
        .unwrap_err();
        assert!(err.contains("stale or altered plan"), "{err}");
        assert!(receipt.parts.is_empty());
        assert!(root.join("AGENTS.md").exists() && root.join("AGENTS.local.md").exists());
    }

    #[test]
    fn a_rollback_that_cannot_undo_a_part_names_the_file_it_kept() {
        let id = format!("{RECEIPT_PREFIX}3-0");
        let (_tmp, root, dir) = project(&id);
        // An engine receipt that is not on disk: that part cannot be undone and is named.
        let part = PartRecord {
            part: AGENTS_TEXT.into(),
            receipts: vec!["agents-guidance-0-0".into()],
            ..PartRecord::default()
        };
        let mut receipt = receipt_with(&id, vec![part]);
        let message = fail_with_rollback(
            &root,
            &dir,
            &mut receipt,
            "failed at personal-text: x".into(),
            PERMISSION,
        );
        assert!(
            message.starts_with("partial apply: applied [agents-text]"),
            "{message}"
        );
        assert!(
            message.contains(&format!("the rollback kept {GUIDANCE_TARGET}")),
            "{message}"
        );
        assert!(
            message.contains(&format!("agents-merge-revert --receipt {id}")),
            "{message}"
        );
        assert_eq!(status_on_disk(&dir), "reverted_partly");
    }
}
