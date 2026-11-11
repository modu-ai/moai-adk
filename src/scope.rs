//! write_scope overlap between active claims (Issue #2007, 06-C; REQ-06C-05/06).
//!
//! Two claims conflict when some path could fall under both of their write_scope entries.
//! A conflict marks the new claim `blocked(reason=scope_conflict)` unless the leader names an
//! integration order (`--integration-order <task_ref>`), in which case it is allowed and the
//! order is recorded. [`check`] computes the decision and returns the record change as
//! `record_update_request`; [`check_and_record`] also commits it to the claim journal through
//! 06-B's `Journal::commit` when the record came from the journal (`--task`).
//!
//! A handed-off claim (`pr_ready`) whose PRs have all closed is no longer a conflict candidate
//! (#2404): [`handed_off_status`] reads the PR state from the forge for the overlapping
//! `pr_ready` claims only, and [`check_with`] lists an ended one under `skipped_ended`. A state
//! that cannot be settled keeps the conflict as `uncertain`; nothing of the other claim's
//! record is written.
//!
//! The overlap test is the interface 05-B's private `plan::paths_overlap` is meant to move to
//! (decision 2026-11-08 19:1x); 06-C does not edit `src/plan.rs`.
use crate::claim::{ClaimRecord, ACTIVE_STATES};
use serde_json::{json, Value};

/// Paths every claim shares by default (REQ-06C-06): two active claims that both reach one of
/// these conflict exactly like any other overlap.
pub const SHARED_INTERFACE_PATHS: &[&str] = &[
    "Cargo.lock",
    "Cargo.toml",
    ".github/**",
    ".moai-cli/manifests/**",
    "AGENTS.md",
];

/// The path's segments, lowercased and with `.` and `..` resolved lexically; `None` when a
/// `..` climbs above the repository root. Lowercasing is conservative: on a case-sensitive
/// file system two spellings are two files, but treating them as one only adds conflicts,
/// never hides one (#2392, design boundary table row 6).
fn segments(path: &str) -> Option<Vec<String>> {
    let mut out: Vec<String> = Vec::new();
    for s in path.trim().split('/') {
        match s {
            "" | "." => {}
            ".." => {
                out.pop()?;
            }
            _ => out.push(s.to_lowercase()),
        }
    }
    Some(out)
}

/// `*` inside one segment matches any run of characters within that segment.
fn segment_match(pattern: &str, text: &str) -> bool {
    let parts: Vec<&str> = pattern.split('*').collect();
    if parts.len() == 1 {
        return pattern == text;
    }
    let (first, last) = (parts[0], parts[parts.len() - 1]);
    if !text.starts_with(first) || text.len() < first.len() + last.len() || !text.ends_with(last) {
        return false;
    }
    let mut rest = &text[first.len()..text.len() - last.len()];
    for mid in &parts[1..parts.len() - 1] {
        match rest.find(mid) {
            Some(i) => rest = &rest[i + mid.len()..],
            None => return false,
        }
    }
    true
}

fn segments_compatible(a: &str, b: &str) -> bool {
    match (a.contains('*'), b.contains('*')) {
        (false, false) => a == b,
        (true, false) => segment_match(a, b),
        (false, true) => segment_match(b, a),
        // Two wildcard segments: assume some name matches both (conservative).
        (true, true) => true,
    }
}

// @MX:ANCHOR: [AUTO] write_scope overlap interface owned by 06-C #2007
// @MX:REASON: scope-check, the shared interface list and (later) 05-B plan admission all decide
//   ownership conflicts through this one predicate; changing it changes who may write where
// @MX:DEBT: [AUTO] std-only matching: exact path, directory prefix, `**` (anything below) and
//   `*` within one segment; no character classes, braces or escapes (Cargo.toml additions are
//   forbidden for this Issue, so no glob crate)
// @MX:CEILING: repository-relative paths compared as text; case variants and `..` are folded
//   conservatively (#2392); links are not resolved here — `check_and_record` inspects the
//   scope on disk and blocks any link it reaches (#2392 F6, `physical_findings`)
// @MX:UPGRADE: adopt a standard glob matcher once a dependency decision allows one
/// True when some path could be written under both `a` and `b`. A plain path also covers
/// everything below it (as 05-B's prefix rule does), and `**` covers everything below. Paths
/// are compared case-insensitively with `.` and `..` resolved; a `..` that leaves the
/// repository overlaps everything, because the path it names is unknown.
pub fn overlap(a: &str, b: &str) -> bool {
    let (Some(sa), Some(sb)) = (segments(a), segments(b)) else {
        return true;
    };
    for (x, y) in sa.iter().zip(sb.iter()) {
        if *x == "**" || *y == "**" {
            return true;
        }
        if !segments_compatible(x, y) {
            return false;
        }
    }
    true
}

fn is_shared(path: &str) -> bool {
    SHARED_INTERFACE_PATHS.iter().any(|s| overlap(s, path))
}

/// Every pair of overlapping entries, in input order.
pub fn overlaps(mine: &[String], theirs: &[String]) -> Vec<Value> {
    let mut out = Vec::new();
    for m in mine {
        for t in theirs {
            if overlap(m, t) {
                out.push(json!({
                    "mine": m,
                    "theirs": t,
                    "shared_interface": is_shared(m) && is_shared(t),
                }));
            }
        }
    }
    out
}

/// The record state a lane leaves once its PR is opened (written by `pr-confirm`).
const HANDED_OFF_STATE: &str = "pr_ready";

/// The other claims whose PR state decides whether they still count (#2404): active, not
/// `mine`, `pr_ready`, and overlapping `mine`'s write_scope. Only these are looked up.
fn handed_off_overlaps<'a>(mine: &ClaimRecord, records: &'a [ClaimRecord]) -> Vec<&'a ClaimRecord> {
    records
        .iter()
        .filter(|o| {
            o.task_key != mine.task_key
                && o.state == HANDED_OFF_STATE
                && !overlaps(&mine.write_scope, &o.write_scope).is_empty()
        })
        .collect()
}

/// Whether a handed-off record's work has ended, judged from its PRs on the forge (#2404).
/// A record links to its PRs only through `pr_create` effects; the `response_id` of a
/// confirmed one is read as the PR number when it is all digits. The record has ended only
/// when it has at least one confirmed `pr_create`, every one of them is `closed` (merged or
/// not), and none is still `intended` or `uncertain` (#2329: several PRs per epoch).
/// Anything else that cannot be settled — a lookup failure, a non-GitHub task, a response id
/// that is not a number, a PR without a `state` — is `uncertain` and keeps the conflict.
pub fn pr_status<T: crate::forge::transport::Transport, C: crate::forge::policy::Clock>(
    record: &ClaimRecord,
    client: &crate::forge::Client<'_, T, C>,
) -> Value {
    let uncertain =
        |reason: String| json!({"ended": false, "pr_state": "uncertain", "reason": reason});
    let tr = &record.task_ref;
    if tr.provider != "github" || tr.host != "github.com" {
        return uncertain(format!("{}/{} is not a GitHub task", tr.provider, tr.host));
    }
    let creates: Vec<&Value> = record
        .effects
        .iter()
        .filter(|e| e["kind"] == json!("pr_create"))
        .collect();
    if creates.iter().any(|e| e["state"] != json!("confirmed")) {
        return uncertain("a pr_create effect is not confirmed".into());
    }
    if creates.is_empty() {
        return uncertain("no confirmed pr_create effect".into());
    }
    let mut prs = Vec::new();
    let mut all_closed = true;
    for e in creates {
        let Some(number) = e["response_id"]
            .as_str()
            .filter(|r| !r.is_empty() && r.bytes().all(|b| b.is_ascii_digit()))
            .and_then(|r| r.parse::<i64>().ok())
        else {
            return uncertain("a pr_create response id is not a PR number".into());
        };
        let pr = match crate::flow::fetch_pr(client, &tr.owner, &tr.repo, number) {
            Ok(pr) => pr,
            Err(e) => return uncertain(format!("PR #{number} lookup failed: {e}")),
        };
        let Some(state) = pr["state"].as_str() else {
            return uncertain(format!("PR #{number} response has no state"));
        };
        all_closed &= state == "closed";
        prs.push(json!({"number": number, "state": state, "merged_at": pr["merged_at"]}));
    }
    json!({"ended": all_closed, "pr_state": if all_closed { "closed" } else { "open" }, "prs": prs})
}

/// True when some `pr_ready` claim overlaps `mine` — the only case that needs the forge.
pub fn needs_pr_status(mine: &ClaimRecord, records: &[ClaimRecord]) -> bool {
    !handed_off_overlaps(mine, records).is_empty()
}

/// [`pr_status`] for every handed-off record that overlaps `mine`, keyed by task key. No
/// overlap with a `pr_ready` record means no forge call at all. Read before the journal lock.
pub fn handed_off_status<T: crate::forge::transport::Transport, C: crate::forge::policy::Clock>(
    mine: &ClaimRecord,
    records: &[ClaimRecord],
    client: &crate::forge::Client<'_, T, C>,
) -> std::collections::BTreeMap<String, Value> {
    handed_off_overlaps(mine, records)
        .into_iter()
        .map(|o| (o.task_key.clone(), pr_status(o, client)))
        .collect()
}

/// The scope decision for `mine` against every other active claim in `records`.
/// `integration_order` is the task the leader put first; a conflict with it is allowed.
pub fn check(
    mine: &ClaimRecord,
    records: &[ClaimRecord],
    integration_order: Option<&str>,
) -> Value {
    check_with(
        mine,
        records,
        integration_order,
        &std::collections::BTreeMap::new(),
    )
}

/// [`check`] with the forge view of handed-off records ([`handed_off_status`]): a record whose
/// PRs have all closed is reported under `skipped_ended` instead of conflicting; an unsettled
/// one keeps its conflict and carries the `pr_status` that explains why.
pub fn check_with(
    mine: &ClaimRecord,
    records: &[ClaimRecord],
    integration_order: Option<&str>,
    handed_off: &std::collections::BTreeMap<String, Value>,
) -> Value {
    let mut conflicts = Vec::new();
    let mut skipped_ended = Vec::new();
    for other in records {
        if other.task_key == mine.task_key || !ACTIVE_STATES.contains(&other.state.as_str()) {
            continue;
        }
        let overlap = overlaps(&mine.write_scope, &other.write_scope);
        if overlap.is_empty() {
            continue;
        }
        let status = handed_off.get(&other.task_key);
        if status.is_some_and(|s| s["ended"] == json!(true)) {
            skipped_ended.push(json!({
                "other_task": other.task_key,
                "other_issue_number": other.task_ref.issue_number,
                "pr_status": status,
            }));
            continue;
        }
        let ordered = integration_order.is_some_and(|o| {
            o == other.task_key
                || o == other.logical_key
                || o == format!("#{}", other.task_ref.issue_number)
        });
        let mut conflict = json!({
            "scope_conflict": {
                "other_task": other.task_key,
                "other_logical_key": other.logical_key,
                "other_issue_number": other.task_ref.issue_number,
                "overlap": overlap,
            },
            "allowed_by_integration_order": ordered,
        });
        if let Some(s) = status {
            conflict["pr_status"] = s.clone();
        }
        conflicts.push(conflict);
    }
    let blocked = conflicts
        .iter()
        .any(|c| c["allowed_by_integration_order"] == false);
    let decision = if blocked {
        "blocked"
    } else if conflicts.is_empty() {
        "clear"
    } else {
        "allowed_with_order"
    };
    let request = if blocked {
        json!({
            "task_key": mine.task_key,
            "epoch": mine.epoch,
            "state": "blocked",
            "reason": "scope_conflict",
        })
    } else if conflicts.is_empty() {
        Value::Null
    } else {
        json!({
            "task_key": mine.task_key,
            "epoch": mine.epoch,
            "integration_order": integration_order,
        })
    };
    json!({
        "decision": decision,
        "task_key": mine.task_key,
        "conflicts": conflicts,
        "skipped_ended": skipped_ended,
        "shared_interface_paths": SHARED_INTERFACE_PATHS,
        "record_update_request": request,
        "record_update_owner": crate::worktree::RECORD_UPDATE_OWNER,
    })
}

/// Directory entries one physical inspection may read before it gives up (#2392 F6).
const INSPECT_BUDGET: usize = 100_000;

struct Inspection<'a> {
    pattern: &'a str,
    budget: usize,
    findings: Vec<Value>,
}

impl Inspection<'_> {
    fn finding(&mut self, path: &str, reason: &str) {
        self.findings
            .push(json!({"scope": self.pattern, "path": path, "reason": reason}));
    }

    /// Lists `dir` without following links; `None` once a finding has been recorded.
    fn list(
        &mut self,
        dir: &std::path::Path,
        rel: &str,
    ) -> Option<Vec<(String, std::fs::DirEntry)>> {
        let entries = match std::fs::read_dir(dir) {
            Ok(entries) => entries,
            // A missing part is a new path the claim may create.
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Some(Vec::new()),
            Err(_) => {
                self.finding(rel, "unreadable");
                return None;
            }
        };
        let mut out = Vec::new();
        for entry in entries {
            if self.budget == 0 {
                self.finding(rel, "inspection_budget");
                return None;
            }
            self.budget -= 1;
            let Ok(entry) = entry else {
                self.finding(rel, "unreadable");
                return None;
            };
            out.push((entry.file_name().to_string_lossy().into_owned(), entry));
        }
        Some(out)
    }

    /// The entry at `path`: a link is a finding; a directory is returned for descent.
    fn observe(&mut self, path: &std::path::Path, rel: &str) -> Option<bool> {
        match std::fs::symlink_metadata(path) {
            Ok(m) if m.file_type().is_symlink() => {
                self.finding(rel, "symlink");
                None
            }
            Ok(m) => Some(m.is_dir()),
            Err(_) => {
                self.finding(rel, "unreadable");
                None
            }
        }
    }

    /// Everything below `dir`, which a plain path or `**` covers.
    fn subtree(&mut self, dir: &std::path::Path, rel: &str) {
        let Some(entries) = self.list(dir, rel) else {
            return;
        };
        for (name, entry) in entries {
            let child = join_rel(rel, &name);
            if self.observe(&entry.path(), &child) == Some(true) {
                self.subtree(&entry.path(), &child);
            }
            if !self.findings.is_empty() {
                return;
            }
        }
    }
}

fn join_rel(rel: &str, name: &str) -> String {
    if rel.is_empty() {
        name.to_string()
    } else {
        format!("{rel}/{name}")
    }
}

/// Physical findings for one write_scope entry under `root`, read with at most `budget`
/// directory entries. The raw spelling is walked (no lexical `..` folding, which would erase
/// `link/../x`); names match case-insensitively, as [`overlap`] does, so a case variant reaches
/// the entry it aliases on a case-insensitive file system and only adds blocks elsewhere.
fn inspect_entry(root: &std::path::Path, pattern: &str, budget: usize) -> Vec<Value> {
    let mut ins = Inspection {
        pattern,
        budget,
        findings: Vec::new(),
    };
    let trimmed = pattern.trim();
    if trimmed.starts_with('/') {
        ins.finding(trimmed, "absolute");
        return ins.findings;
    }
    let parts: Vec<&str> = trimmed
        .split('/')
        .filter(|s| !s.is_empty() && *s != ".")
        .collect();
    if parts.contains(&"..") {
        ins.finding(trimmed, "parent_component");
        return ins.findings;
    }
    // The directories reached so far, with their root-relative spelling on disk.
    let mut frontier = vec![(root.to_path_buf(), String::new())];
    for (i, part) in parts.iter().enumerate() {
        if *part == "**" {
            for (dir, rel) in &frontier {
                ins.subtree(dir, rel);
                if !ins.findings.is_empty() {
                    break;
                }
            }
            return ins.findings;
        }
        let last = i + 1 == parts.len();
        let wanted = part.to_lowercase();
        let mut next = Vec::new();
        for (dir, rel) in &frontier {
            let Some(entries) = ins.list(dir, rel) else {
                return ins.findings;
            };
            for (name, entry) in entries {
                if !segments_compatible(&wanted, &name.to_lowercase()) {
                    continue;
                }
                let child = join_rel(rel, &name);
                match ins.observe(&entry.path(), &child) {
                    None => return ins.findings,
                    // A plain path covers everything below it.
                    Some(true) if last => ins.subtree(&entry.path(), &child),
                    Some(true) => next.push((entry.path(), child)),
                    Some(false) => {}
                }
                if !ins.findings.is_empty() {
                    return ins.findings;
                }
            }
        }
        frontier = next;
    }
    ins.findings
}

// @MX:NOTE: [AUTO] #2392 F6: a check-time observation of the scope-check worktree only; a link
//   created after the check, and links in other worktrees, are outside this guard
/// Why `scope` may not be trusted as written under `root` (#2392 F6): every link reached by an
/// entry (literal, `*`, `**` or a covered directory) and every absolute, `..`, unreadable or
/// over-budget entry, root-relative. Links are never followed, so an internal alias, a dangling
/// link and a link out of the project all block alike. Empty when the scope is plain.
pub fn physical_findings(root: &std::path::Path, scope: &[String]) -> Vec<Value> {
    scope
        .iter()
        .flat_map(|p| inspect_entry(root, p, INSPECT_BUDGET))
        .collect()
}

/// [`check`] plus the record write: `blocked` with a `scope_conflict` effect, or an
/// `integration_order` effect when the leader's order allowed the overlap. A repeated
/// check with the same outcome adds no event.
///
/// The write_scope is also inspected on disk under the `project` work tree (#2392 F6): a
/// [`physical_findings`] entry blocks with reason `scope_symlink`, which no integration order
/// lifts, and the inspected root is reported under `physical_check`.
pub fn check_and_record<C: crate::forge::policy::Clock>(
    project: &std::path::Path,
    mine: &ClaimRecord,
    from_journal: bool,
    records: &[ClaimRecord],
    integration_order: Option<&str>,
    handed_off: &std::collections::BTreeMap<String, Value>,
    clock: &C,
) -> std::result::Result<Value, crate::contract::OpError> {
    use crate::worktree::{commit_update, not_written, open_locked_journal};
    let mut out = check_with(mine, records, integration_order, handed_off);
    // The measured top level, never the record's own worktree path; read before the lock.
    let root = crate::worktree::identity(project)?.path;
    let findings = physical_findings(std::path::Path::new(&root), &mine.write_scope);
    out["physical_check"] = json!({"root": root, "findings": findings});
    let effect = if !findings.is_empty() {
        out["decision"] = json!("blocked");
        out["record_update_request"] = json!({
            "task_key": mine.task_key,
            "epoch": mine.epoch,
            "state": "blocked",
            "reason": "scope_symlink",
        });
        Some(json!({"kind": "scope_symlink", "findings": findings, "conflicts": out["conflicts"]}))
    } else {
        match out["decision"].as_str() {
            Some("blocked") => {
                Some(json!({"kind": "scope_conflict", "conflicts": out["conflicts"]}))
            }
            Some("allowed_with_order") => {
                Some(json!({"kind": "integration_order", "order": integration_order}))
            }
            _ => None,
        }
    };
    out["record_update"] = match (effect, from_journal) {
        (None, _) => json!({"applied": false, "reason": "no scope conflict"}),
        (Some(_), false) => not_written(),
        (Some(effect), true) => {
            let blocked = out["decision"] == "blocked";
            let journal = open_locked_journal(project)?;
            commit_update(
                &journal,
                &mine.task_key,
                mine.epoch,
                "scope_check",
                clock,
                |r| {
                    if blocked {
                        r.state = "blocked".into();
                    }
                    if r.effects.last() != Some(&effect) {
                        r.effects.push(effect);
                    }
                },
            )?
        }
    };
    Ok(out)
}

#[cfg(test)]
mod tests {
    use super::{inspect_entry, overlap};

    fn reasons(root: &std::path::Path, pattern: &str, budget: usize) -> Vec<String> {
        inspect_entry(root, pattern, budget)
            .iter()
            .map(|f| {
                format!(
                    "{}:{}",
                    f["reason"].as_str().unwrap(),
                    f["path"].as_str().unwrap()
                )
            })
            .collect()
    }

    #[test]
    fn inspection_refuses_absolute_parent_and_budget() {
        let tmp = tempfile::tempdir().unwrap();
        let root = tmp.path();
        std::fs::create_dir_all(root.join("safe/a")).unwrap();
        std::fs::write(root.join("safe/a/x"), b"").unwrap();
        std::fs::write(root.join("safe/y"), b"").unwrap();
        assert_eq!(reasons(root, "/etc/**", 100), ["absolute:/etc/**"]);
        assert_eq!(
            reasons(root, "safe/../x", 100),
            ["parent_component:safe/../x"]
        );
        // Four entries are read (`safe` in the root, `a` and `y`, then `x`): 4 suffices, 3 not.
        assert!(reasons(root, "safe/**", 4).is_empty());
        assert!(reasons(root, "safe", 4).is_empty());
        assert_eq!(reasons(root, "safe/**", 3), ["inspection_budget:safe/a"]);
        // A missing tail is a new path, not a finding.
        assert!(reasons(root, "safe/new/deeper.rs", 100).is_empty());
    }

    #[test]
    fn overlap_rules() {
        assert!(overlap("src/lib.rs", "src/lib.rs"));
        assert!(overlap("src", "src/lib.rs"));
        assert!(overlap("./src/", "src/a.rs"));
        assert!(overlap(".github/**", ".github/workflows/rc-ci.yml"));
        assert!(overlap("src/*.rs", "src/lib.rs"));
        assert!(overlap("docs/**", "docs/a/*.md"));
        assert!(!overlap("src/lib.rs", "src/main.rs"));
        assert!(!overlap("src/*.rs", "src/lib.md"));
        assert!(!overlap("docs/a.md", "src/a.md"));
        assert!(!overlap("src/lib.rs", "src/lib.rs.bak"));
    }
}
