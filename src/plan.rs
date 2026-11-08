//! PLAN and task DAG admission (Issue #2002, 05-B).
//!
//! A plan orders the tasks of an approved 05-A snapshot and decides, before RUN, whether the
//! work may start: `ready`, `stale` (something it was approved or pinned against changed) or
//! `blocked` (something it needs is missing). The decision is recomputed from its inputs every
//! time; nothing is stored, and execution is always reported as `not_run`.

use crate::digest;
use crate::spec::{content_digest, Canonical, Execution, Outcome, SortedKeys, SpecDocument};
use serde::{Deserialize, Serialize};
use serde_json::json;
use std::collections::{BTreeMap, BTreeSet};
use std::path::Path;
use std::process::Command;

pub type Result<T> = std::result::Result<T, String>;

/// Verification kinds that count as an executable verifier for an acceptance criterion.
const VERIFIER_KINDS: [&str; 4] = ["command", "test", "property", "manual-review"];

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Task {
    pub id: String,
    #[serde(default)]
    pub depends_on: Vec<String>,
    /// Repository-relative paths the task writes.
    #[serde(default)]
    pub write_paths: Vec<String>,
    /// Acceptance criteria of the snapshot this task is verified against.
    #[serde(default)]
    pub verifier_refs: Vec<String>,
    /// Host the task needs, if any; compared with the session's capabilities.
    pub host: Option<String>,
}

/// An approval record. It names what was approved; it never grants host permissions.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Approval {
    pub id: String,
    /// Opaque; owner and epoch fields belong to 06-A.
    pub approver: String,
    pub basis: String,
    pub snapshot_reference: String,
    pub snapshot_digest: String,
    pub allowed_actions: Vec<String>,
    pub revoked: bool,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct AdmissionRequest {
    pub protocol_version: u32,
    pub spec: SpecDocument,
    pub tasks: Vec<Task>,
    pub approval: Approval,
    /// Commit id the plan was pinned to (`HEAD^{commit}` at plan time).
    pub expected_revision: String,
    /// Hosts this session can run tasks on.
    #[serde(default)]
    pub capabilities: Vec<String>,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct Dangling {
    pub from: String,
    pub to: String,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct Blocked {
    pub task: String,
    pub reason: &'static str,
    pub suggestion: String,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct Conflict {
    pub tasks: [String; 2],
    pub paths: Vec<String>,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct Plan {
    pub spec_ref: Option<String>,
    pub snapshot_digest: Option<String>,
    pub expected_revision: String,
    /// Topological order (Kahn, ties broken by id); empty when there is a cycle.
    pub order: Vec<String>,
    /// Strongly connected components that form a cycle (Tarjan), each sorted by id.
    pub cycles: Vec<Vec<String>>,
    pub structure: Outcome,
    pub duplicate_ids: Vec<String>,
    pub dangling: Vec<Dangling>,
    pub blocked: Vec<Blocked>,
    /// Task pairs whose write paths overlap with no dependency between them.
    pub parallel_conflicts: Vec<Conflict>,
    pub execution: Execution,
    pub plan_digest: String,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum Decision {
    Ready,
    Stale,
    Blocked,
}

/// What 05-C receives for a task admitted to RUN.
#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct PinnedContract {
    pub spec_ref: String,
    pub plan_digest: String,
    pub task_id: String,
    pub expected_revision: String,
    pub scope: Vec<String>,
    pub approval_ref: String,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct AdmissionDecision {
    pub decision: Decision,
    pub reasons: Vec<&'static str>,
    pub expected_revision: String,
    pub current_revision: String,
    pub plan: Plan,
    /// One entry per task in order, only when the decision is `ready`.
    pub pinned: Vec<PinnedContract>,
}

fn normalize(path: &str) -> String {
    let mut p = path.trim();
    while let Some(rest) = p.strip_prefix("./") {
        p = rest;
    }
    p.trim_end_matches('/').to_string()
}

// @MX:DEBT: [AUTO] write-scope overlap by path-string prefix; the interface belongs to 06-C #2007
// @MX:CEILING: repository-relative paths; symlinks and case-insensitive filesystems are not seen
// @MX:UPGRADE: replace with the 06-C write_scope interface once it lands
fn paths_overlap(a: &str, b: &str) -> bool {
    a == b || b.starts_with(&format!("{a}/")) || a.starts_with(&format!("{b}/"))
}

/// Tarjan's strongly connected components over `deps` (task -> its dependencies).
fn cycles(deps: &BTreeMap<&str, BTreeSet<&str>>) -> Vec<Vec<String>> {
    struct State<'a> {
        index: BTreeMap<&'a str, usize>,
        low: BTreeMap<&'a str, usize>,
        stack: Vec<&'a str>,
        on_stack: BTreeSet<&'a str>,
        found: Vec<Vec<String>>,
    }
    fn visit<'a>(v: &'a str, deps: &BTreeMap<&'a str, BTreeSet<&'a str>>, s: &mut State<'a>) {
        let i = s.index.len();
        s.index.insert(v, i);
        s.low.insert(v, i);
        s.stack.push(v);
        s.on_stack.insert(v);
        for &w in &deps[v] {
            if !s.index.contains_key(w) {
                visit(w, deps, s);
                let lw = s.low[w];
                let lv = s.low.get_mut(v).expect("visited");
                *lv = (*lv).min(lw);
            } else if s.on_stack.contains(w) {
                let iw = s.index[w];
                let lv = s.low.get_mut(v).expect("visited");
                *lv = (*lv).min(iw);
            }
        }
        if s.low[v] == s.index[v] {
            let mut scc = Vec::new();
            while let Some(w) = s.stack.pop() {
                s.on_stack.remove(w);
                scc.push(w.to_string());
                if w == v {
                    break;
                }
            }
            if scc.len() > 1 || deps[v].contains(v) {
                scc.sort();
                s.found.push(scc);
            }
        }
    }
    let mut s = State {
        index: BTreeMap::new(),
        low: BTreeMap::new(),
        stack: Vec::new(),
        on_stack: BTreeSet::new(),
        found: Vec::new(),
    };
    for &v in deps.keys() {
        if !s.index.contains_key(v) {
            visit(v, deps, &mut s);
        }
    }
    s.found.sort();
    s.found
}

/// Kahn's algorithm with a sorted ready set, so equal inputs give equal orders.
fn order(deps: &BTreeMap<&str, BTreeSet<&str>>) -> Option<Vec<String>> {
    let mut indegree: BTreeMap<&str, usize> = deps.iter().map(|(k, d)| (*k, d.len())).collect();
    let mut dependents: BTreeMap<&str, Vec<&str>> = BTreeMap::new();
    for (task, ds) in deps {
        for d in ds {
            dependents.entry(d).or_default().push(task);
        }
    }
    let mut ready: BTreeSet<&str> = indegree
        .iter()
        .filter(|(_, n)| **n == 0)
        .map(|(k, _)| *k)
        .collect();
    let mut out = Vec::new();
    while let Some(t) = ready.pop_first() {
        out.push(t.to_string());
        for d in dependents.get(t).into_iter().flatten() {
            let n = indegree.get_mut(d).expect("known task");
            *n -= 1;
            if *n == 0 {
                ready.insert(d);
            }
        }
    }
    (out.len() == deps.len()).then_some(out)
}

/// Every task reachable through `depends_on` from `start`.
fn ancestors<'a>(start: &'a str, deps: &BTreeMap<&'a str, BTreeSet<&'a str>>) -> BTreeSet<&'a str> {
    let mut seen = BTreeSet::new();
    let mut todo: Vec<&str> = deps[start].iter().copied().collect();
    while let Some(t) = todo.pop() {
        if seen.insert(t) {
            todo.extend(deps[t].iter().copied());
        }
    }
    seen
}

/// Builds the plan for the snapshot's tasks. Pure: no files, no git, no clock.
pub fn build_plan(spec: &SpecDocument, tasks: &[Task], expected_revision: &str) -> Plan {
    let mut sorted: Vec<&Task> = tasks.iter().collect();
    sorted.sort_by(|a, b| a.id.cmp(&b.id));
    let mut duplicate_ids: Vec<String> = sorted
        .windows(2)
        .filter(|w| w[0].id == w[1].id)
        .map(|w| w[0].id.clone())
        .collect();
    duplicate_ids.dedup();
    sorted.dedup_by(|a, b| a.id == b.id);
    let by_id: BTreeMap<&str, &Task> = sorted.iter().map(|t| (t.id.as_str(), *t)).collect();

    let mut dangling = Vec::new();
    let mut deps: BTreeMap<&str, BTreeSet<&str>> = BTreeMap::new();
    for t in &sorted {
        let entry = deps.entry(t.id.as_str()).or_default();
        for d in &t.depends_on {
            match by_id.get_key_value(d.as_str()) {
                Some((k, _)) => {
                    entry.insert(k);
                }
                None => dangling.push(Dangling {
                    from: t.id.clone(),
                    to: d.clone(),
                }),
            }
        }
    }

    let criteria: BTreeMap<&str, &crate::spec::Acceptance> =
        spec.acceptance.iter().map(|a| (a.id.as_str(), a)).collect();
    let mut blocked = Vec::new();
    for t in &sorted {
        let mut has_verifier = !t.verifier_refs.is_empty();
        for r in &t.verifier_refs {
            match criteria.get(r.as_str()) {
                None => {
                    dangling.push(Dangling {
                        from: t.id.clone(),
                        to: r.clone(),
                    });
                    has_verifier = false;
                }
                Some(ac) => {
                    let v = &ac.verification;
                    if !VERIFIER_KINDS.contains(&v.kind.as_str()) || v.selector.trim().is_empty() {
                        has_verifier = false;
                    }
                }
            }
        }
        if !has_verifier {
            blocked.push(Blocked {
                task: t.id.clone(),
                reason: "verifier-missing",
                suggestion: format!(
                    "add a task before {} that secures a verifier (command, test, property or manual review) for its acceptance criteria",
                    t.id
                ),
            });
        }
    }

    let order = order(&deps);
    let cycles = if order.is_some() {
        Vec::new()
    } else {
        cycles(&deps)
    };

    let mut parallel_conflicts = Vec::new();
    let reach: BTreeMap<&str, BTreeSet<&str>> =
        deps.keys().map(|k| (*k, ancestors(k, &deps))).collect();
    for (i, a) in sorted.iter().enumerate() {
        for b in &sorted[i + 1..] {
            if reach[a.id.as_str()].contains(b.id.as_str())
                || reach[b.id.as_str()].contains(a.id.as_str())
            {
                continue;
            }
            let mut paths = BTreeSet::new();
            for pa in a.write_paths.iter().map(|p| normalize(p)) {
                for pb in b.write_paths.iter().map(|p| normalize(p)) {
                    if paths_overlap(&pa, &pb) {
                        paths.insert(pa.clone());
                        paths.insert(pb);
                    }
                }
            }
            if !paths.is_empty() {
                parallel_conflicts.push(Conflict {
                    tasks: [a.id.clone(), b.id.clone()],
                    paths: paths.into_iter().collect(),
                });
            }
        }
    }

    let spec_ref = spec.snapshot.as_ref().map(|s| s.reference.clone());
    let snapshot_digest = spec.snapshot.as_ref().map(|s| s.digest.clone());
    let plan_digest = digest(&SortedKeys.bytes(&json!({
        "spec_ref": spec_ref,
        "snapshot_digest": snapshot_digest,
        "expected_revision": expected_revision,
        "tasks": sorted,
    })));
    Plan {
        spec_ref,
        snapshot_digest,
        expected_revision: expected_revision.into(),
        order: order.unwrap_or_default(),
        cycles,
        structure: if duplicate_ids.is_empty() && dangling.is_empty() {
            Outcome::Passed
        } else {
            Outcome::Failed
        },
        duplicate_ids,
        dangling,
        blocked,
        parallel_conflicts,
        execution: Execution::NotRun,
        plan_digest,
    }
}

/// Re-reads everything the approval and the pin rest on and decides. `current_revision` is
/// the repository's `HEAD^{commit}` now (see [`current_revision`]).
pub fn admit(req: &AdmissionRequest, current_revision: &str) -> AdmissionDecision {
    let plan = build_plan(&req.spec, &req.tasks, &req.expected_revision);
    let mut stale: Vec<&'static str> = Vec::new();
    let mut blocked: Vec<&'static str> = Vec::new();
    match (&req.spec.snapshot, req.spec.status.as_str()) {
        (Some(snap), "approved") => {
            if snap.digest != content_digest(&req.spec, &SortedKeys) {
                stale.push("snapshot-digest-mismatch");
            }
            if req.approval.snapshot_reference != snap.reference {
                stale.push("approval-reference-mismatch");
            }
            if req.approval.snapshot_digest != snap.digest {
                stale.push("approval-snapshot-mismatch");
            }
        }
        _ => blocked.push("not-approved"),
    }
    if req.expected_revision != current_revision {
        stale.push("expected-revision-mismatch");
    }
    if req.approval.revoked {
        blocked.push("approval-revoked");
    }
    if !req.approval.allowed_actions.iter().any(|a| a == "run") {
        blocked.push("approval-scope");
    }
    if !plan.cycles.is_empty() {
        blocked.push("cycle");
    }
    if !plan.duplicate_ids.is_empty() {
        blocked.push("duplicate-task-id");
    }
    if !plan.dangling.is_empty() {
        blocked.push("dangling-dependency");
    }
    if !plan.blocked.is_empty() {
        blocked.push("verifier-missing");
    }
    let missing_host = req
        .tasks
        .iter()
        .filter_map(|t| t.host.as_ref())
        .any(|h| !req.capabilities.contains(h));
    if missing_host {
        blocked.push("capability-missing");
    }
    let decision = match (stale.is_empty(), blocked.is_empty()) {
        (false, _) => Decision::Stale,
        (true, false) => Decision::Blocked,
        (true, true) => Decision::Ready,
    };
    let pinned = if decision == Decision::Ready {
        let by_id: BTreeMap<&str, &Task> = req.tasks.iter().map(|t| (t.id.as_str(), t)).collect();
        plan.order
            .iter()
            .map(|id| PinnedContract {
                spec_ref: plan.spec_ref.clone().unwrap_or_default(),
                plan_digest: plan.plan_digest.clone(),
                task_id: id.clone(),
                expected_revision: req.expected_revision.clone(),
                scope: by_id[id.as_str()]
                    .write_paths
                    .iter()
                    .map(|p| normalize(p))
                    .collect(),
                approval_ref: req.approval.id.clone(),
            })
            .collect()
    } else {
        Vec::new()
    };
    stale.extend(blocked);
    AdmissionDecision {
        decision,
        reasons: stale,
        expected_revision: req.expected_revision.clone(),
        current_revision: current_revision.into(),
        plan,
        pinned,
    }
}

/// `git rev-parse --verify --end-of-options HEAD^{commit}` in `project`, run with an argv
/// array (no shell). A missing repository or commit is an error, never a guessed revision.
pub fn current_revision(project: &Path) -> Result<String> {
    let out = Command::new("git")
        .arg("-C")
        .arg(project)
        .args([
            "rev-parse",
            "--verify",
            "--quiet",
            "--end-of-options",
            "HEAD^{commit}",
        ])
        .output()
        .map_err(|_| "git unavailable")?;
    if !out.status.success() {
        return Err(format!(
            "git rev-parse exited {}",
            out.status.code().unwrap_or(-1)
        ));
    }
    let rev = String::from_utf8(out.stdout).map_err(|_| "non UTF-8 git output")?;
    let rev = rev.trim();
    let hex = rev
        .bytes()
        .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b));
    if !(hex && (rev.len() == 40 || rev.len() == 64)) {
        return Err("git did not report a commit id".into());
    }
    Ok(rev.to_string())
}
