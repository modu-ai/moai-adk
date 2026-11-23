//! The whole-file targets of the routing layer: `rules/moai.md` and `rules/moai/flow.md` inside
//! a Claude config folder, and the removal of the earlier `moai-routing` block from that
//! folder's `CLAUDE.md` (SPEC-20261122-claude-md-home-routing, SPEC-20261122-claude-md-flow-rule).
//!
//! A rules file is moai's own only while it holds the bytes moai last wrote, the way the style
//! file is (`home_style`): the record is one state file per config folder,
//! `home-routing/rules-<12 hex of the sha256 of the folder path>.json`, holding both files'
//! last-written digests. Anything else is a `conflict` and is kept. A link on the way (the
//! config folder, `rules`, `rules/moai` or the file itself) and a file that is not UTF-8 are
//! conflicts too; only a refused `CLAUDE_CONFIG_DIR` (or an unreadable state file) blocks a
//! folder. The earlier block stays in the user's `CLAUDE.md` until the removal is picked; the
//! removal needs the routing file in place and moai's own, and the block bytes equal to the
//! baseline the earlier version recorded in `home-routing/claude.json`.
//!
//! Nothing here is public: `home_routing` plans, applies and reverts through these functions.

use super::*;
use crate::host_home::{place_in, Host};

/// The routing text: at most this many lines.
pub(super) const ROUTING_LINES: usize = 12;
/// The flow text: at most this many lines and bytes.
pub(super) const FLOW_LINES: usize = 80;
pub(super) const FLOW_BYTES: usize = 5 * 1024;

/// The routing source, checked: the checks of the block source plus the line limit and the
/// `paths` key.
pub(super) fn routing_source(package: &Path) -> Result<String> {
    let text = source(package)?;
    if text.lines().count() > ROUTING_LINES {
        return Err(format!("{SOURCE} must be at most {ROUTING_LINES} lines"));
    }
    if frontmatter_has_paths(&text) {
        return Err(format!("{SOURCE} must not carry a paths key"));
    }
    Ok(text)
}

/// The flow source, checked; `None` when the package has none (a package from before it).
pub(super) fn flow_source(package: &Path) -> Result<Option<String>> {
    let path = package.join(FLOW_SOURCE);
    let text = match fs::symlink_metadata(&path) {
        Ok(m) if m.is_file() => {
            fs::read_to_string(&path).map_err(|_| format!("{FLOW_SOURCE} is not readable UTF-8"))?
        }
        Ok(_) => return Err(format!("{FLOW_SOURCE} is not a regular file")),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(None),
        Err(_) => return Err(format!("{FLOW_SOURCE} is unreadable")),
    };
    if text.starts_with("---") {
        return Err(format!(
            "{FLOW_SOURCE} must not start with frontmatter (the rule carries no paths key)"
        ));
    }
    if text.lines().count() > FLOW_LINES {
        return Err(format!("{FLOW_SOURCE} must be at most {FLOW_LINES} lines"));
    }
    if text.len() > FLOW_BYTES {
        return Err(format!("{FLOW_SOURCE} must be at most 5 KiB"));
    }
    if text.is_empty() || !text.ends_with('\n') {
        return Err(format!("{FLOW_SOURCE} must end with a newline"));
    }
    Ok(Some(text))
}

fn frontmatter_has_paths(text: &str) -> bool {
    let mut lines = text.lines();
    if lines.next().map(str::trim_end) != Some("---") {
        return false;
    }
    lines
        .take_while(|l| l.trim_end() != "---")
        .any(|l| l.trim_start().starts_with("paths") && l.contains(':'))
}

/// The 12 hex digits that name a config folder's state file.
pub(super) fn folder_id(dir: &Path) -> Result<String> {
    let s = dir.to_str().ok_or("the config folder path is not UTF-8")?;
    let digest = sha(s.as_bytes());
    let hex = digest.strip_prefix("sha256:").unwrap_or(&digest);
    Ok(hex[..12].to_string())
}

/// The files of one config folder moai last wrote.
#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub(super) struct RulesState {
    schema_version: u32,
    routing_sha256: Option<String>,
    flow_sha256: Option<String>,
}

fn state_file(home: &Path, id: &str) -> PathBuf {
    state_root(home).join(format!("rules-{id}.json"))
}

pub(super) fn read_rules_state(home: &Path, id: &str) -> Result<RulesState> {
    plain_dir_or_absent(&crate::home_root::path(home))?;
    plain_dir_or_absent(&state_root(home))?;
    match read_opt(&state_file(home, id))? {
        None => Ok(RulesState {
            schema_version: SCHEMA_VERSION,
            ..RulesState::default()
        }),
        Some(bytes) => {
            let state: RulesState = serde_json::from_slice(&bytes).map_err(|_| {
                "routing state of the config folder is unreadable; inspect it first"
            })?;
            if state.schema_version != SCHEMA_VERSION {
                return Err("routing state of the config folder has an unknown schema".into());
            }
            Ok(state)
        }
    }
}

fn update_rules_state(home: &Path, id: &str, kind: &str, owned: Option<String>) -> Result<()> {
    let mut state = read_rules_state(home, id)?;
    match kind {
        "routing" => state.routing_sha256 = owned,
        _ => state.flow_sha256 = owned,
    }
    write_json(&state_file(home, id), &state)
}

/// What a target holds.
pub(super) enum Cur {
    Absent,
    Bytes(Vec<u8>),
    /// A link on the way, a file that is not regular or not UTF-8, an unreadable file.
    Conflict(String),
}

/// The folder (the config folder or the one a followed link leads to) under which `rel` lies.
fn base_of(home: &Path, dir: &Path, rel: &str) -> Result<PathBuf> {
    Ok(place_in(Host::Claude, home, dir, rel)?.base)
}

/// `base` and every folder from it down to the parent of `rel` must be plain or absent.
fn chain_plain(base: &Path, rel: &str) -> Result<()> {
    plain_dir_or_absent(base)?;
    let mut d = base.to_path_buf();
    let parts: Vec<&str> = rel.split('/').collect();
    for part in &parts[..parts.len() - 1] {
        d.push(part);
        plain_dir_or_absent(&d)?;
    }
    Ok(())
}

fn read_target(base: &Path, rel: &str) -> Cur {
    if let Err(e) = chain_plain(base, rel) {
        return Cur::Conflict(e);
    }
    let path = base.join(rel);
    match current_file(&path) {
        Err(e) => Cur::Conflict(e),
        Ok(None) => Cur::Absent,
        Ok(Some(b)) if std::str::from_utf8(&b).is_err() => {
            Cur::Conflict(format!("{} is not valid UTF-8; kept", path.display()))
        }
        Ok(Some(b)) => Cur::Bytes(b),
    }
}

/// The file's path and what it holds; a refused place is a conflict, never a block.
fn target(home: &Path, dir: &Path, rel: &str) -> (PathBuf, Cur) {
    match base_of(home, dir, rel) {
        Ok(base) => {
            let cur = read_target(&base, rel);
            (base.join(rel), cur)
        }
        Err(e) => (dir.join(rel), Cur::Conflict(e)),
    }
}

fn is_write(action: &str) -> bool {
    action == "create" || action == "merge"
}

/// The three-way plan of a whole file: ours only while it holds the bytes we last wrote.
fn whole_file_plan(path: &Path, cur: &Cur, owned: Option<&str>, rendered: &str) -> FilePlan {
    let mut p = FilePlan {
        path: path.to_string_lossy().into_owned(),
        action: String::new(),
        tracked_by_default: false,
        expected_sha256: None,
        section_id: None,
        baseline_sha256: owned.map(str::to_string),
        rendered_sha256: Some(sha(rendered.as_bytes())),
        detail: None,
    };
    let action = match cur {
        Cur::Conflict(why) => {
            p.detail = Some(why.clone());
            "conflict"
        }
        Cur::Absent => "create",
        Cur::Bytes(b) => {
            let digest = sha(b);
            p.expected_sha256 = Some(digest.clone());
            if b.as_slice() == rendered.as_bytes() {
                if owned == Some(digest.as_str()) {
                    "unchanged"
                } else {
                    p.detail =
                        Some("the file equals the current text; recorded as MoAI's own".into());
                    "merge"
                }
            } else if owned == Some(digest.as_str()) {
                "merge"
            } else {
                p.detail = Some(
                    "the file is not the one MoAI last wrote (edited or someone else's); preserve and review"
                        .into(),
                );
                "conflict"
            }
        }
    };
    p.action = action.into();
    p
}

/// `begin` … `end` in `text`: the start of the begin line and the end of the end line's
/// newline; `None` without markers, an error when they are doubled or out of order.
fn find_block(text: &str, begin: &str, end: &str) -> Result<Option<(usize, usize)>> {
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
    Ok(Some((start, block_end)))
}

const OLD_BLOCK_WARNING: &str = "a session of a profile that imports this CLAUDE.md may lose the routing text, because such a profile reads no rules folder";

/// The earlier block in the folder's `CLAUDE.md`: offered, removed on the pick, kept with its
/// reason, or waiting for the routing file. `None` when there is nothing to offer.
fn old_block_plan(
    home: &Path,
    dir: &Path,
    baseline: Option<&str>,
    pick: bool,
    routing_in_place: bool,
) -> Option<OldBlockPlan> {
    let (path, cur) = target(home, dir, "CLAUDE.md");
    let mut plan = OldBlockPlan {
        path: path.to_string_lossy().into_owned(),
        action: String::new(),
        expected_sha256: None,
        baseline_sha256: baseline.map(str::to_string),
        detail: None,
    };
    let keep = |mut p: OldBlockPlan, why: String| {
        p.action = "kept".into();
        p.detail = Some(why);
        Some(p)
    };
    let bytes = match cur {
        Cur::Absent => return None,
        Cur::Conflict(why) => return pick.then(|| keep(plan, why)).flatten(),
        Cur::Bytes(b) => b,
    };
    let text = std::str::from_utf8(&bytes).expect("checked UTF-8");
    let (begin, end) = markers(SECTION, "html");
    let (s, e) = match find_block(text, &begin, &end) {
        Ok(None) => return None,
        Ok(Some(range)) => range,
        Err(why) => return pick.then(|| keep(plan, why)).flatten(),
    };
    plan.expected_sha256 = Some(sha(&bytes));
    let Some(base) = baseline else {
        return keep(
            plan,
            "no record of what moai wrote there; the block is kept".into(),
        );
    };
    if sha(&text.as_bytes()[s..e]) != base {
        return keep(
            plan,
            "the block was edited after moai wrote it; it is kept".into(),
        );
    }
    if !routing_in_place {
        plan.action = "waiting".into();
        plan.detail = Some(format!(
            "the block stays until {RULES_FILE} is in place and moai's own; run again after it is written"
        ));
        return Some(plan);
    }
    plan.action = if pick { "remove" } else { "offered" }.into();
    plan.detail = Some(format!(
        "the block is the same text as {RULES_FILE}; removing it is a separate pick: {OLD_BLOCK_WARNING}"
    ));
    Some(plan)
}

/// One config folder's plan, with the plans the apply works from.
pub(super) struct Folder {
    pub row: HostPlan,
    dir: Option<PathBuf>,
    id: String,
    routing: Option<FilePlan>,
    flow: Option<FilePlan>,
    old: Option<OldBlockPlan>,
}

impl Folder {
    pub fn writes(&self) -> bool {
        self.row.preflight == "ok"
            && (self.routing.iter().any(|p| is_write(&p.action))
                || self.flow.iter().any(|p| is_write(&p.action))
                || self.old.iter().any(|p| p.action == "remove"))
    }
}

pub(super) fn plan_folder(
    home: &Path,
    dir: Result<PathBuf>,
    is_default: bool,
    routing_text: &str,
    flow_text: Option<&str>,
    pick: bool,
) -> Folder {
    let blocked = |dir: &Path, reason: String| Folder {
        row: HostPlan {
            host: "claude".into(),
            config_dir: dir.to_string_lossy().into_owned(),
            file: dir.join(RULES_FILE).to_string_lossy().into_owned(),
            preflight: "blocked".into(),
            blocked_reason: Some(reason),
            section: None,
            flow: None,
            old_block: None,
        },
        dir: None,
        id: String::new(),
        routing: None,
        flow: None,
        old: None,
    };
    let dir = match dir {
        Ok(d) => d,
        Err(e) => return blocked(&home.join(Host::Claude.default_dir()), e),
    };
    let state = folder_id(&dir).and_then(|id| read_rules_state(home, &id).map(|s| (id, s)));
    let (id, state) = match state {
        Ok(v) => v,
        Err(e) => return blocked(&dir, e),
    };
    let (rpath, rcur) = target(home, &dir, RULES_FILE);
    let routing = whole_file_plan(&rpath, &rcur, state.routing_sha256.as_deref(), routing_text);
    let flow = flow_text.map(|t| {
        let (path, cur) = target(home, &dir, FLOW_FILE);
        whole_file_plan(&path, &cur, state.flow_sha256.as_deref(), t)
    });
    let in_place =
        matches!(&rcur, Cur::Bytes(b) if state.routing_sha256.as_deref() == Some(sha(b).as_str()));
    // Only the default folder has a baseline: the earlier version wrote its block there alone.
    let baseline = if is_default {
        read_state(home, "claude").ok().and_then(|s| s.block_sha256)
    } else {
        None
    };
    let old = old_block_plan(home, &dir, baseline.as_deref(), pick, in_place);
    // `update` asks the routing row alone whether anything is written: when only the flow file
    // changes, the row carries that action so the layer still counts as a write.
    let mut section = routing.clone();
    if section.action == "unchanged" {
        if let Some(f) = flow.as_ref().filter(|f| is_write(&f.action)) {
            section.action = f.action.clone();
            section.detail = Some(format!(
                "{RULES_FILE} is unchanged; the flow rule file {} is {}",
                f.path, f.action
            ));
        }
    }
    Folder {
        row: HostPlan {
            host: "claude".into(),
            config_dir: dir.to_string_lossy().into_owned(),
            file: rpath.to_string_lossy().into_owned(),
            preflight: "ok".into(),
            blocked_reason: None,
            section: Some(section),
            flow: flow.clone(),
            old_block: old.clone(),
        },
        dir: Some(dir),
        id,
        routing: Some(routing),
        flow,
        old,
    }
}

/// The missing folders from `base`'s first existing ancestor down to the parent of `rel`,
/// created top down; their paths, for the receipt.
fn make_parents(base: &Path, rel: &str) -> Result<Vec<String>> {
    let mut chain = vec![base.to_path_buf()];
    let mut d = base.to_path_buf();
    let parts: Vec<&str> = rel.split('/').collect();
    for part in &parts[..parts.len() - 1] {
        d.push(part);
        chain.push(d.clone());
    }
    let mut above = Vec::new();
    let mut up = base.parent();
    while let Some(a) = up {
        if fs::symlink_metadata(a).is_ok() {
            break;
        }
        above.push(a.to_path_buf());
        up = a.parent();
    }
    above.reverse();
    let mut created = Vec::new();
    for dir in above.into_iter().chain(chain) {
        match fs::symlink_metadata(&dir) {
            Ok(_) => {}
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
                fs::create_dir(&dir).map_err(|_| format!("{} unavailable", dir.display()))?;
                created.push(dir.to_string_lossy().into_owned());
            }
            Err(_) => return Err(format!("{} is unreadable", dir.display())),
        }
    }
    chain_plain(base, rel)?;
    Ok(created)
}

struct Journal<'a> {
    home: &'a Path,
    receipt_dir: &'a Path,
    receipt: &'a mut Receipt,
}

impl Journal<'_> {
    fn save(&self) -> Result<()> {
        write_json(&self.receipt_dir.join("receipt.json"), &*self.receipt)
    }
}

fn write_rules_file(
    j: &mut Journal,
    dir: &Path,
    id: &str,
    kind: &str,
    rel: &str,
    plan: &FilePlan,
    text: &str,
) -> Result<()> {
    let base = base_of(j.home, dir, rel)?;
    let before = match read_target(&base, rel) {
        Cur::Absent => None,
        Cur::Bytes(b) => Some(b),
        Cur::Conflict(why) => return Err(why),
    };
    if before.as_deref().map(sha) != plan.expected_sha256 {
        return Err(format!(
            "{} changed since the plan; re-preview required",
            plan.path
        ));
    }
    let target = base.join(rel);
    let created = make_parents(&base, rel)?;
    if let Some(b) = &before {
        publish_exclusive(&j.receipt_dir.join(format!("{kind}-{id}.before")), b)?;
    }
    match &before {
        Some(b) => publish_replace(&target, Some(b), text.as_bytes())?,
        None => publish_exclusive(&target, text.as_bytes())?,
    }
    j.receipt.entries.push(ReceiptEntry {
        host: "claude".into(),
        kind: Some(kind.into()),
        config_dir: Some(dir.to_string_lossy().into_owned()),
        action: plan.action.clone(),
        before_sha256: before.as_deref().map(sha),
        after_sha256: sha(text.as_bytes()),
        baseline_before: plan.baseline_sha256.clone(),
        created_dirs: created,
    });
    j.save()?;
    update_rules_state(j.home, id, kind, Some(sha(text.as_bytes())))
}

fn remove_old_block(j: &mut Journal, dir: &Path, id: &str, plan: &OldBlockPlan) -> Result<()> {
    let base = base_of(j.home, dir, "CLAUDE.md")?;
    let Cur::Bytes(before) = read_target(&base, "CLAUDE.md") else {
        return Err(format!(
            "{} changed since the plan; re-preview required",
            plan.path
        ));
    };
    if Some(sha(&before)) != plan.expected_sha256 {
        return Err(format!(
            "{} changed since the plan; re-preview required",
            plan.path
        ));
    }
    let text = std::str::from_utf8(&before).map_err(|_| "target is not UTF-8")?;
    let (begin, end) = markers(SECTION, "html");
    let (s, e) = find_block(text, &begin, &end)?.ok_or("the block is gone; re-preview required")?;
    if Some(sha(&text.as_bytes()[s..e])) != plan.baseline_sha256 {
        return Err("the block changed since the plan; re-preview required".into());
    }
    let out = format!("{}{}", &text[..s], &text[e..]);
    publish_exclusive(
        &j.receipt_dir.join(format!("old_block-{id}.before")),
        &before,
    )?;
    let target = base.join("CLAUDE.md");
    publish_replace(&target, Some(&before), out.as_bytes())?;
    j.receipt.entries.push(ReceiptEntry {
        host: "claude".into(),
        kind: Some("old_block".into()),
        config_dir: Some(dir.to_string_lossy().into_owned()),
        action: "remove".into(),
        before_sha256: Some(sha(&before)),
        after_sha256: sha(out.as_bytes()),
        baseline_before: plan.baseline_sha256.clone(),
        created_dirs: Vec::new(),
    });
    j.save()?;
    write_json(
        &state_root(j.home).join("claude.json"),
        &HostState {
            schema_version: SCHEMA_VERSION,
            block_sha256: None,
        },
    )
}

/// Writes one folder's files, the routing file, the flow file, then the removal of the earlier
/// block; the paths written.
pub(super) fn apply_folder(
    home: &Path,
    receipt_dir: &Path,
    receipt: &mut Receipt,
    f: &Folder,
    routing_text: &str,
    flow_text: Option<&str>,
) -> Result<Vec<String>> {
    let (Some(dir), true) = (&f.dir, f.row.preflight == "ok") else {
        return Ok(Vec::new());
    };
    let mut j = Journal {
        home,
        receipt_dir,
        receipt,
    };
    let mut written = Vec::new();
    if let Some(p) = f.routing.as_ref().filter(|p| is_write(&p.action)) {
        write_rules_file(&mut j, dir, &f.id, "routing", RULES_FILE, p, routing_text)?;
        written.push(p.path.clone());
    }
    if let (Some(p), Some(text)) = (f.flow.as_ref().filter(|p| is_write(&p.action)), flow_text) {
        write_rules_file(&mut j, dir, &f.id, "flow", FLOW_FILE, p, text)?;
        written.push(p.path.clone());
    }
    if let Some(p) = f.old.as_ref().filter(|p| p.action == "remove") {
        remove_old_block(&mut j, dir, &f.id, p)?;
        written.push(p.path.clone());
    }
    Ok(written)
}

/// One entry of a receipt undone: `Ok(Ok(path))` when restored, `Ok(Err(skipped))` when left
/// as it is, with the entry's file named.
pub(super) fn revert_entry(
    home: &Path,
    receipt_dir: &Path,
    e: &ReceiptEntry,
) -> Result<std::result::Result<String, serde_json::Value>> {
    let kind = e.kind.as_deref().unwrap_or("old_block");
    let rel = match kind {
        "routing" => RULES_FILE,
        "flow" => FLOW_FILE,
        "old_block" => "CLAUDE.md",
        _ => return Err("routing receipt names an unknown entry kind".into()),
    };
    let skip = |path: &Path, reason: String| {
        Ok(Err(serde_json::json!({
            "host": e.host, "kind": kind, "path": path.to_string_lossy(), "reason": reason,
        })))
    };
    let dir = match &e.config_dir {
        Some(d) if !Path::new(d).is_absolute() => {
            return Err("routing receipt names a relative config folder".into())
        }
        Some(d) => PathBuf::from(d),
        None => match crate::host_home::resolve_named(&e.host, home) {
            Ok(d) => d,
            Err(why) => {
                return skip(&home.join(rel), format!("host_folder_not_plain: {why}"));
            }
        },
    };
    let base = match base_of(home, &dir, rel).and_then(|b| chain_plain(&b, rel).map(|()| b)) {
        Ok(b) => b,
        Err(why) => return skip(&dir.join(rel), format!("host_folder_not_plain: {why}")),
    };
    let target = base.join(rel);
    let current = match current_file(&target) {
        Ok(c) => c,
        Err(why) => return skip(&target, format!("file_not_plain: {why}")),
    };
    if current.as_deref().map(sha) != Some(e.after_sha256.clone()) {
        return skip(&target, "user_changed_since_apply".into());
    }
    let current = current.expect("checked above");
    match &e.before_sha256 {
        None => {
            fs::remove_file(&target).map_err(|_| format!("{} not removed", target.display()))?;
            for d in e.created_dirs.iter().rev() {
                // Only a real, empty folder this apply made goes; `remove_dir` refuses the rest.
                if fs::symlink_metadata(d).is_ok_and(|m| m.is_dir()) {
                    let _ = fs::remove_dir(d);
                }
            }
        }
        Some(before_sha) => {
            let name = if e.kind.is_some() {
                format!("{kind}-{}.before", folder_id(&dir)?)
            } else {
                format!("{}.before", e.host)
            };
            let before = read_opt(&receipt_dir.join(name))?
                .ok_or("before bytes missing from the receipt")?;
            if &sha(&before) != before_sha {
                return Err("before bytes do not match the receipt".into());
            }
            publish_replace(&target, Some(&current), &before)?;
        }
    }
    if kind == "old_block" {
        let mut state = read_state(home, &e.host)?;
        state.block_sha256 = e.baseline_before.clone();
        write_json(&state_root(home).join(format!("{}.json", e.host)), &state)?;
    } else {
        update_rules_state(home, &folder_id(&dir)?, kind, e.baseline_before.clone())?;
    }
    Ok(Ok(target.to_string_lossy().into_owned()))
}
