//! `worktree-move-plan` / `worktree-move-apply` (Issue #2723): an inventory of a clone's
//! linked worktrees with a preview of moving each under one common root
//! `<root>/<project key>/<name>`, and a move that runs only on a fresh, identical preview.
//!
//! The root defaults to `$HOME/<HOME_STATE_DIR>/worktrees` ([`default_root`]; user decision
//! #1975 6859326988 puts it in the HOME state folder, whose name lives only in
//! `crate::HOME_STATE_DIR`); `--root` names another absolute folder. The project key is
//! `<repository folder name>-<first 12 hex of clone_id>`, the clone id of `identity` (sha256
//! of the canonical common git folder), so two projects with the same name on one machine
//! never share a folder.
//!
//! A worktree is a move candidate only when every fact was read and none blocks it. Blocked:
//! the main worktree (git cannot move it), a bare entry, a missing (prunable) or locked
//! worktree, one holding submodules (git cannot move it), uncommitted or untracked files,
//! commits on no remote-tracking ref (no fetch is made), a stash entry of its branch, a
//! process working inside it, the caller's own directory inside it, a branch the caller's
//! observations list as held (HOLD, unconfirmed handoff) and, without those observations, the
//! hold check itself (`hold_check_not_run`). A fact that cannot be read blocks with its name.
//!
//! The plan writes nothing (git runs with `--no-optional-locks` and a scrubbed environment).
//! The apply recomputes the plan, refuses before any write unless the digest is the same,
//! re-checks each candidate for processes and open files right before moving it, and moves it
//! with `git worktree move` (never `--force`). Then it checks that git lists the worktree at
//! its new path with the same HEAD and branch and that git works inside it; if not, it runs
//! `git worktree repair` there once and checks again. The answer names, for each moved
//! worktree, the `git worktree move` that puts it back. Nothing else under the project's
//! state folders is written; the only folder created is `<root>/<project key>`.

use crate::identity::scrub_git_env;
use crate::worktree_report::parse_worktree_list;
use serde::Deserialize;
use serde_json::{json, Value};
use std::fs;
use std::path::{Component, Path, PathBuf};
use std::process::Command;

type Result<T> = std::result::Result<T, String>;

/// The observations file (typed JSON, unknown keys rejected). Without `holds` the hold check
/// has not run and every worktree is blocked by it.
#[derive(Clone, Debug, Default, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Observations {
    pub schema_version: u32,
    #[serde(default)]
    pub holds: Option<HoldsObservation>,
}

/// Branches that must not move, as read from the forge (HOLD rows, claims whose handoff is
/// not confirmed). An empty list is an observation that nothing is held.
#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct HoldsObservation {
    pub source: String,
    pub branches: Vec<HeldBranch>,
}

#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct HeldBranch {
    pub branch: String,
    /// `hold` or `handoff_pending`.
    pub reason: String,
}

const HOLD_REASONS: &[&str] = &["hold", "handoff_pending"];

/// Reads the observations file; none means `schema_version` 1 with no part observed.
pub fn read_observations(path: Option<&Path>) -> Result<Observations> {
    match path {
        None => Ok(Observations {
            schema_version: 1,
            ..Default::default()
        }),
        Some(p) => {
            let text = fs::read_to_string(p)
                .map_err(|e| format!("observations file unavailable: {}: {e}", p.display()))?;
            serde_json::from_str(&text).map_err(|e| format!("invalid observations: {e}"))
        }
    }
}

fn git(dir: &Path, args: &[&str]) -> Option<String> {
    let mut cmd = Command::new("git");
    scrub_git_env(&mut cmd);
    let out = cmd
        .arg("--no-optional-locks")
        .arg("-C")
        .arg(dir)
        .args(args)
        .output()
        .ok()?;
    out.status
        .success()
        .then(|| String::from_utf8_lossy(&out.stdout).into_owned())
}

/// A folder name of 1 to 64 of ASCII letters, digits, `.`, `_`, `-`; anything else becomes
/// `-`, and a name that would be empty or start with `.` is prefixed with `p`.
fn safe_name(raw: &str) -> String {
    let mut s: String = raw
        .chars()
        .map(|c| {
            if c.is_ascii_alphanumeric() || matches!(c, '.' | '_' | '-') {
                c
            } else {
                '-'
            }
        })
        .take(64)
        .collect();
    if s.is_empty() || s.starts_with('.') {
        s.insert(0, 'p');
    }
    s
}

/// `<folder name>-<12 hex>` from the primary worktree's folder name and the clone id.
pub fn project_key(primary_name: &str, clone_id: &str) -> Result<String> {
    let hex = clone_id.strip_prefix("sha256:").unwrap_or(clone_id);
    if hex.len() < 12 || !hex.bytes().all(|b| b.is_ascii_hexdigit()) {
        return Err("clone id unavailable".into());
    }
    Ok(format!("{}-{}", safe_name(primary_name), &hex[..12]))
}

/// The root must be absolute and plain (no `.` or `..` part); it is not created here.
fn check_root(root: &Path) -> Result<()> {
    if !root.is_absolute() {
        return Err("--root must be an absolute path".into());
    }
    // `components()` drops an inner `.`, so the text is read part by part.
    let text = root.to_string_lossy();
    if text.split(['/', '\\']).any(|p| p == "." || p == "..")
        || root.components().any(|c| matches!(c, Component::ParentDir))
    {
        return Err("--root must not contain `.` or `..` parts".into());
    }
    Ok(())
}

/// `path` with its longest existing ancestor canonicalised (symlinks resolved), the rest kept.
fn resolve_lexical(path: &Path) -> PathBuf {
    let mut rest = Vec::new();
    let mut cur = path.to_path_buf();
    loop {
        if let Ok(c) = fs::canonicalize(&cur) {
            let mut out = c;
            for part in rest.iter().rev() {
                out.push(part);
            }
            return out;
        }
        match (cur.file_name().map(|n| n.to_os_string()), cur.parent()) {
            (Some(name), Some(parent)) => {
                rest.push(name);
                cur = parent.to_path_buf();
            }
            _ => return path.to_path_buf(),
        }
    }
}

fn has_submodules(wt: &Path) -> Option<bool> {
    if wt.join(".gitmodules").exists() {
        return Some(true);
    }
    let staged = git(wt, &["ls-files", "--stage"])?;
    Some(staged.lines().any(|l| l.starts_with("160000 ")))
}

/// The common root when the caller names none: `$HOME/<HOME_STATE_DIR>/worktrees` (#2723,
/// user decision #1975 6859326988: the HOME state folder is the one place for it; the folder
/// name comes only from `crate::HOME_STATE_DIR`, so the switch of #2725 follows by itself).
pub fn default_root() -> Result<PathBuf> {
    let home = std::env::var_os("HOME")
        .filter(|h| !h.is_empty())
        .map(PathBuf::from)
        .ok_or("HOME is not set; pass --root")?;
    if !home.is_absolute() {
        return Err("HOME is not an absolute path; pass --root".into());
    }
    Ok(home.join(crate::HOME_STATE_DIR).join("worktrees"))
}

/// The plan for the clone holding `project`; `root` defaults to [`default_root`].
pub fn plan(project: &Path, root: Option<&Path>, obs: &Observations) -> Result<Value> {
    let owned;
    let root = match root {
        Some(r) => r,
        None => {
            owned = default_root()?;
            owned.as_path()
        }
    };
    if obs.schema_version != 1 {
        return Err("observations schema_version must be 1".into());
    }
    if let Some(h) = &obs.holds {
        if h.source.trim().is_empty() {
            return Err("holds.source must name where the holds were read".into());
        }
        if let Some(b) = h
            .branches
            .iter()
            .find(|b| !HOLD_REASONS.contains(&b.reason.as_str()))
        {
            return Err(format!(
                "holds reason `{}` must be hold or handoff_pending",
                b.reason
            ));
        }
    }
    check_root(root)?;
    let top = git(
        project,
        &["rev-parse", "--path-format=absolute", "--show-toplevel"],
    )
    .map(|s| PathBuf::from(s.trim()))
    .ok_or("not inside a git work tree")?;
    let listing =
        git(&top, &["worktree", "list", "--porcelain"]).ok_or("git worktree list failed")?;
    let entries = parse_worktree_list(&listing);
    let primary = entries
        .first()
        .map(|e| fs::canonicalize(&e.path).unwrap_or_else(|_| e.path.clone()))
        .ok_or("no worktree listed")?;
    let id = crate::identity::identity(&primary)?;
    let clone_id = id["clone_id"].as_str().ok_or("clone id unavailable")?;
    let name = primary
        .file_name()
        .map(|n| n.to_string_lossy().into_owned())
        .unwrap_or_default();
    let key = project_key(&name, clone_id)?;
    let root_resolved = resolve_lexical(root);
    let all_paths: Vec<PathBuf> = entries
        .iter()
        .filter(|e| !e.bare)
        .map(|e| fs::canonicalize(&e.path).unwrap_or_else(|_| e.path.clone()))
        .collect();
    if let Some(inside) = all_paths.iter().find(|p| root_resolved.starts_with(p)) {
        return Err(format!(
            "--root lies inside the worktree {}; choose a folder outside every worktree",
            inside.display()
        ));
    }
    let project_root = root_resolved.join(&key);
    let cwd = std::env::current_dir()
        .ok()
        .and_then(|p| fs::canonicalize(p).ok());
    let cwds = crate::worktree_gc::process_cwds().map(|v| {
        v.into_iter()
            .map(|p| fs::canonicalize(&p).unwrap_or(p))
            .collect::<Vec<_>>()
    });
    let stashes = crate::lane_reset::stash_counts(&primary);

    let mut items = Vec::new();
    for (i, e) in entries.iter().enumerate() {
        let path = fs::canonicalize(&e.path).unwrap_or_else(|_| e.path.clone());
        let base = path
            .file_name()
            .map(|n| safe_name(&n.to_string_lossy()))
            .unwrap_or_else(|| "p".into());
        let destination = project_root.join(&base);
        let mut reasons: Vec<String> = Vec::new();
        let mut action = "move";
        if e.bare {
            reasons.push("bare".into());
        } else if i == 0 {
            reasons.push("main_worktree".into());
        } else if path.parent() == Some(project_root.as_path()) {
            action = "keep";
        } else if e.prunable {
            reasons.push("missing".into());
        } else {
            if e.locked {
                reasons.push("locked".into());
            }
            match has_submodules(&path) {
                Some(true) => reasons.push("has_submodules".into()),
                Some(false) => {}
                None => reasons.push("submodule_check_unknown".into()),
            }
            match git(&path, &["status", "--porcelain", "--untracked-files=no"]) {
                Some(s) if s.trim().is_empty() => {}
                Some(_) => reasons.push("uncommitted".into()),
                None => reasons.push("status_unknown".into()),
            }
            match crate::worktree_gc::census(&path) {
                Some((0, _, _)) => {}
                Some(_) => reasons.push("untracked".into()),
                None => reasons.push("untracked_unknown".into()),
            }
            match git(
                &path,
                &["rev-list", "--count", "HEAD", "--not", "--remotes"],
            )
            .and_then(|t| t.trim().parse::<u64>().ok())
            {
                Some(0) => {}
                Some(_) => reasons.push("unpushed".into()),
                None => reasons.push("unpushed_unknown".into()),
            }
            let label = e.branch.clone().unwrap_or_else(|| "(no branch)".into());
            match &stashes {
                Some(s) if s.contains(&label) => reasons.push("stash".into()),
                Some(_) => {}
                None => reasons.push("stash_unknown".into()),
            }
            match &cwds {
                Some(list) if list.iter().any(|c| c.starts_with(&path)) => {
                    reasons.push("in_use".into())
                }
                Some(_) => {}
                None => reasons.push("process_check_unavailable".into()),
            }
            if cwd.as_ref().is_some_and(|c| c.starts_with(&path)) {
                reasons.push("current_directory".into());
            }
            match &obs.holds {
                None => reasons.push("hold_check_not_run".into()),
                Some(h) => {
                    for b in &h.branches {
                        if e.branch.as_deref() == Some(b.branch.as_str()) {
                            reasons.push(b.reason.clone());
                        }
                    }
                }
            }
            if destination.exists() || fs::symlink_metadata(&destination).is_ok() {
                reasons.push("destination_exists".into());
            }
        }
        if !reasons.is_empty() {
            action = "blocked";
        }
        items.push(json!({
            "path": path.display().to_string(),
            "branch": e.branch,
            "head": e.head,
            "action": action,
            "reasons": reasons,
            "destination": if action == "blocked" && (e.bare || i == 0) { Value::Null } else { json!(destination.display().to_string()) },
        }));
    }
    // Two worktrees whose names map to one destination: neither moves.
    let moving: Vec<String> = items
        .iter()
        .filter(|it| it["action"] == "move")
        .filter_map(|it| it["destination"].as_str().map(str::to_string))
        .collect();
    for it in items.iter_mut() {
        let dest = it["destination"].as_str().map(str::to_string);
        if it["action"] == "move"
            && dest
                .as_ref()
                .is_some_and(|d| moving.iter().filter(|m| *m == d).count() > 1)
        {
            it["action"] = json!("blocked");
            it["reasons"] = json!(["destination_collision"]);
        }
    }
    let bound: Vec<Value> = items
        .iter()
        .map(|it| json!({"path": it["path"], "branch": it["branch"], "head": it["head"], "action": it["action"], "reasons": it["reasons"], "destination": it["destination"]}))
        .collect();
    let plan_digest = crate::digest(
        serde_json::to_string(&json!({"root": project_root, "items": bound}))
            .unwrap_or_default()
            .as_bytes(),
    );
    let count = |a: &str| items.iter().filter(|it| it["action"] == a).count();
    let state = if count("move") == 0 {
        "unchanged"
    } else {
        "changes"
    };
    Ok(json!({
        "schema_version": 1,
        "project_key": key,
        "root": root_resolved.display().to_string(),
        "project_root": project_root.display().to_string(),
        "state": state,
        "totals": {"move": count("move"), "keep": count("keep"), "blocked": count("blocked")},
        "items": items,
        "holds_source": obs.holds.as_ref().map(|h| h.source.clone()),
        "plan_digest": plan_digest,
        "apply": "worktree-move-apply with the same --root and observations, this plan_digest and a permission reference, only after the user approves the moves",
        "network": "disabled",
        "writes": false,
    }))
}

fn listed(primary: &Path) -> Option<Vec<crate::worktree_report::WorktreeEntry>> {
    git(primary, &["worktree", "list", "--porcelain"]).map(|t| parse_worktree_list(&t))
}

/// After a move: git lists `dst` with the same HEAD and branch, no longer lists `src`, and
/// git runs inside `dst` at that HEAD.
fn verify(
    primary: &Path,
    src: &Path,
    dst: &Path,
    head: Option<&str>,
    branch: Option<&str>,
) -> bool {
    let Some(list) = listed(primary) else {
        return false;
    };
    let canon = |p: &Path| fs::canonicalize(p).unwrap_or_else(|_| p.to_path_buf());
    let at = list.iter().find(|e| canon(&e.path) == canon(dst));
    let still_src = list.iter().any(|e| e.path == src);
    let inside = git(dst, &["rev-parse", "HEAD"]).map(|s| s.trim().to_string());
    match at {
        Some(e) => {
            !still_src
                && e.head.as_deref() == head
                && e.branch.as_deref() == branch
                && inside.as_deref() == head
        }
        None => false,
    }
}

/// Act on a plan: recompute, compare digests, then move each `move` item.
pub fn apply(
    project: &Path,
    root: Option<&Path>,
    obs: &Observations,
    plan_digest: &str,
    permission_reference: &str,
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("permission reference required".into());
    }
    let fresh = plan(project, root, obs)?;
    if fresh["plan_digest"].as_str() != Some(plan_digest) {
        return Err("stale or altered plan: the recomputed digest differs; re-preview required, nothing was written".into());
    }
    let project_root = PathBuf::from(fresh["project_root"].as_str().unwrap_or_default());
    let primary = fresh["items"]
        .as_array()
        .and_then(|a| a.first())
        .and_then(|it| it["path"].as_str())
        .map(PathBuf::from)
        .ok_or("no primary worktree")?;
    let mut results = Vec::new();
    let mut not_done = 0usize;
    for it in fresh["items"].as_array().into_iter().flatten() {
        if it["action"] != "move" {
            continue;
        }
        let src = PathBuf::from(it["path"].as_str().unwrap_or_default());
        let dst = PathBuf::from(it["destination"].as_str().unwrap_or_default());
        let head = it["head"].as_str();
        let branch = it["branch"].as_str();
        let outcome: std::result::Result<&str, String> = (|| {
            if let Some(why) = crate::worktree_gc::busy_now(&src, false) {
                return Err(format!("refused: {why}"));
            }
            if dst.exists() || fs::symlink_metadata(&dst).is_ok() {
                return Err("refused: destination_exists".into());
            }
            fs::create_dir_all(&project_root)
                .map_err(|e| format!("could not create {}: {e}", project_root.display()))?;
            let mut cmd = Command::new("git");
            scrub_git_env(&mut cmd);
            let o = cmd
                .arg("-C")
                .arg(&primary)
                .args(["worktree", "move"])
                .arg(&src)
                .arg(&dst)
                .output()
                .map_err(|_| "git worktree move could not run".to_string())?;
            if !o.status.success() {
                return Err(format!(
                    "git worktree move refused: {}",
                    String::from_utf8_lossy(&o.stderr).trim()
                ));
            }
            if verify(&primary, &src, &dst, head, branch) {
                return Ok("moved");
            }
            let mut cmd = Command::new("git");
            scrub_git_env(&mut cmd);
            let _ = cmd
                .arg("-C")
                .arg(&dst)
                .args(["worktree", "repair"])
                .output();
            if verify(&primary, &src, &dst, head, branch) {
                Ok("moved_after_repair")
            } else {
                Err("moved but not verified: git does not list it at the new path with the same HEAD and branch".into())
            }
        })();
        let revert = format!(
            "git -C {} worktree move {} {}",
            primary.display(),
            dst.display(),
            src.display()
        );
        let (status, detail, revert) = match outcome {
            Ok(s) => (s, Value::Null, json!(revert)),
            Err(e) => {
                not_done += 1;
                let moved = e.starts_with("moved but");
                (
                    if e.starts_with("refused") {
                        "refused"
                    } else {
                        "failed"
                    },
                    json!(e),
                    if moved { json!(revert) } else { Value::Null },
                )
            }
        };
        results.push(json!({
            "path": src.display().to_string(),
            "destination": dst.display().to_string(),
            "head": head,
            "branch": branch,
            "status": status,
            "detail": detail,
            "revert": revert,
        }));
    }
    Ok(json!({
        "schema_version": 1,
        "project_key": fresh["project_key"],
        "project_root": fresh["project_root"],
        "plan_digest": plan_digest,
        "state": if results.is_empty() { "unchanged" } else if not_done == 0 { "applied" } else { "partial" },
        "results": results,
        "not_done": not_done,
        "rerun": "worktree-move-plan again: every moved worktree is `keep`",
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn project_key_is_name_and_twelve_hex() {
        let id = format!("sha256:{}", "ab12".repeat(16));
        assert_eq!(
            project_key("moai-adk", &id).unwrap(),
            "moai-adk-ab12ab12ab12"
        );
        assert_eq!(
            project_key("my repo/x", &id).unwrap(),
            "my-repo-x-ab12ab12ab12"
        );
        assert_eq!(
            project_key(".hidden", &id).unwrap(),
            "p.hidden-ab12ab12ab12"
        );
        assert_eq!(project_key("", &id).unwrap(), "p-ab12ab12ab12");
        assert!(project_key("x", "sha256:zz").is_err());
        assert!(project_key("x", "sha256:abcdef01234").is_err());
        assert_eq!(project_key(&"n".repeat(80), &id).unwrap().len(), 64 + 13);
    }

    #[test]
    fn root_must_be_absolute_and_plain() {
        assert!(check_root(Path::new("rel/dir")).is_err());
        assert!(check_root(Path::new("/a/../b")).is_err());
        assert!(check_root(Path::new("/a/./b")).is_err());
        assert!(check_root(Path::new("/a/b")).is_ok());
    }
}
