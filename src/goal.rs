//! The standing goal of a clone (Issue #2679): an instruction such as "keep going until no
//! Issue is open" is given once, kept in a file, and read by the leader and every lane at the
//! start of a round, so a restart or a cleared session does not need the person to say it
//! again.
//!
//! Store: `<git common dir>/.moai-cli/goal/goal.json` — the same anchor as the claim journal,
//! so the primary worktree and every linked worktree of one clone read one file; it is never
//! committed and survives a session restart.
//!
//! Only a change asks the person: `goal-plan` previews the change and its `plan_digest`
//! without writing; `goal-apply` writes only with that digest, recomputed against the file as
//! it is then. Reading (`goal-show`) never asks and says `state: none` when no goal is set.
use crate::apply::{publish_exclusive, publish_replace, read_opt};
use crate::{digest, identity};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;
/// The stored goal with its exact bytes, or none.
type Stored = Option<(Goal, Vec<u8>)>;

pub const SCHEMA_VERSION: u32 = 1;
/// Longest accepted `text`, `end_condition`, `source` or exclusion `reason`, in characters.
pub const MAX_TEXT: usize = 2000;

/// One Issue or pull request the goal leaves out, with why.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Exclusion {
    pub number: u64,
    pub reason: String,
}

/// The typed request of `goal-plan` and `goal-apply`.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Request {
    pub schema_version: u32,
    /// `set` or `clear`.
    pub action: String,
    /// The instruction in the person's words. Required for `set`.
    #[serde(default)]
    pub text: String,
    /// When the goal is met, e.g. "no open pull request and no startable Issue".
    #[serde(default)]
    pub end_condition: String,
    #[serde(default)]
    pub exclusions: Vec<Exclusion>,
    /// Where the instruction came from (who and when, or a comment link).
    #[serde(default)]
    pub source: String,
}

/// The stored goal.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Goal {
    pub schema_version: u32,
    /// false after a `clear`; the last goal stays readable.
    pub active: bool,
    pub text: String,
    pub end_condition: String,
    pub exclusions: Vec<Exclusion>,
    pub source: String,
    /// When this wording was first set; kept while only the end condition, the exclusions or
    /// the source change.
    pub started_at: String,
    pub updated_at: String,
    /// 1 for the first goal, +1 on every applied change.
    pub revision: u64,
}

/// The store path for `project`, a folder inside a git repository.
pub fn store(project: &Path) -> Result<PathBuf> {
    let id = identity::identity(project)?;
    if id["identity_status"].as_str() != Some("ok") {
        return Err("the goal file requires a git repository".into());
    }
    let common = id["clone_anchor"]
        .as_str()
        .ok_or("the goal file requires a git repository")?;
    Ok(Path::new(common)
        .join(crate::CLONE_STATE_DIR)
        .join("goal")
        .join("goal.json"))
}

fn read_goal(path: &Path) -> Result<Stored> {
    let Some(bytes) = read_opt(path)? else {
        return Ok(None);
    };
    let goal: Goal = serde_json::from_slice(&bytes)
        .map_err(|e| format!("goal file unreadable: {}: {e}", path.display()))?;
    if goal.schema_version != SCHEMA_VERSION {
        return Err(format!(
            "goal file has schema_version {}; this binary reads {SCHEMA_VERSION}",
            goal.schema_version
        ));
    }
    Ok(Some((goal, bytes)))
}

/// `goal-show`: the stored goal, or `state: none` when there is none. Reads only.
pub fn show(project: &Path) -> Result<Value> {
    let path = store(project)?;
    Ok(match read_goal(&path)? {
        None => json!({
            "state": "none",
            "store": path.to_string_lossy(),
            "goal": Value::Null,
            "detail": "no goal is set for this clone",
        }),
        Some((goal, bytes)) => json!({
            "state": if goal.active { "active" } else { "cleared" },
            "store": path.to_string_lossy(),
            "goal": goal,
            "sha256": digest(&bytes),
        }),
    })
}

fn check_text(field: &str, value: &str, required: bool) -> Result<()> {
    if required && value.trim().is_empty() {
        return Err(format!("{field} is required"));
    }
    if value.chars().count() > MAX_TEXT {
        return Err(format!("{field} is longer than {MAX_TEXT} characters"));
    }
    if crate::redact::contains_secret_shape(value) {
        return Err(format!(
            "{field} looks like it carries a credential; refused"
        ));
    }
    Ok(())
}

fn validate(r: &Request) -> Result<()> {
    if r.schema_version != SCHEMA_VERSION {
        return Err(format!("schema_version must be {SCHEMA_VERSION}"));
    }
    match r.action.as_str() {
        "set" => {
            check_text("text", &r.text, true)?;
            check_text("end_condition", &r.end_condition, true)?;
            check_text("source", &r.source, true)?;
            let mut seen = std::collections::BTreeSet::new();
            for e in &r.exclusions {
                if e.number == 0 {
                    return Err("exclusion numbers must be positive".into());
                }
                if !seen.insert(e.number) {
                    return Err(format!("exclusion #{} is listed twice", e.number));
                }
                check_text("exclusion reason", &e.reason, true)?;
            }
            Ok(())
        }
        "clear" => {
            if !r.text.is_empty() || !r.end_condition.is_empty() || !r.exclusions.is_empty() {
                return Err("clear takes no text, end_condition or exclusions".into());
            }
            check_text("source", &r.source, true)
        }
        other => Err(format!("action must be set or clear, not `{other}`")),
    }
}

fn sorted(mut exclusions: Vec<Exclusion>) -> Vec<Exclusion> {
    exclusions.sort_by_key(|e| e.number);
    exclusions
}

/// What `goal-plan` previews: the change kind, the fields that change and the digest that
/// binds the request to the file as it is now.
fn preview(path: &Path, request: &Request) -> Result<(Value, Stored)> {
    validate(request)?;
    let current = read_goal(path)?;
    let cur = current.as_ref().map(|(g, _)| g);
    let (change, changed): (&str, Vec<&str>) = match (request.action.as_str(), cur) {
        ("clear", None) => ("none", vec![]),
        ("clear", Some(g)) if !g.active => ("none", vec![]),
        ("clear", Some(_)) => ("clear", vec!["active"]),
        ("set", None) => (
            "create",
            vec!["text", "end_condition", "exclusions", "source"],
        ),
        ("set", Some(g)) => {
            let mut fields = Vec::new();
            if !g.active {
                fields.push("active");
            }
            if g.text != request.text {
                fields.push("text");
            }
            if g.end_condition != request.end_condition {
                fields.push("end_condition");
            }
            if g.exclusions != sorted(request.exclusions.clone()) {
                fields.push("exclusions");
            }
            if g.source != request.source {
                fields.push("source");
            }
            if fields.is_empty() {
                ("unchanged", fields)
            } else {
                ("update", fields)
            }
        }
        _ => unreachable!("validated"),
    };
    let current_sha = current.as_ref().map(|(_, b)| digest(b));
    let mut normalized = request.clone();
    normalized.exclusions = sorted(normalized.exclusions);
    let bound = serde_json::to_vec(&(current_sha.as_deref(), &normalized))
        .map_err(|_| "encode failed".to_string())?;
    let plan_digest = digest(&bound);
    let out = json!({
        "store": path.to_string_lossy(),
        "change": change,
        "changed_fields": changed,
        "current": cur,
        "requested": normalized,
        "current_sha256": current_sha,
        "plan_digest": plan_digest,
        "asks_person": !matches!(change, "unchanged" | "none"),
        "writes": false,
    });
    Ok((out, current))
}

/// Reads a typed request file.
pub fn read_request(path: &Path) -> Result<Request> {
    let bytes =
        std::fs::read(path).map_err(|e| format!("request unavailable: {}: {e}", path.display()))?;
    serde_json::from_slice(&bytes).map_err(|e| format!("invalid goal request: {e}"))
}

/// `goal-plan`: previews a change of the goal. Writes nothing.
pub fn plan(project: &Path, request: &Request) -> Result<Value> {
    Ok(preview(&store(project)?, request)?.0)
}

/// `goal-apply`: writes the previewed change when `plan_digest` still matches a fresh preview
/// (the file unchanged since the preview, the same request). `now` is
/// `YYYY-MM-DDTHH:MM:SSZ`.
pub fn apply(project: &Path, request: &Request, plan_digest: &str, now: &str) -> Result<Value> {
    if crate::decision_file::deadline_secs(now).is_none() {
        return Err(format!(
            "invalid time `{now}`: expected YYYY-MM-DDTHH:MM:SSZ"
        ));
    }
    let path = store(project)?;
    let (fresh, current) = preview(&path, request)?;
    if fresh["plan_digest"] != plan_digest {
        return Err("stale or altered plan; re-preview required".into());
    }
    let change = fresh["change"].as_str().unwrap_or_default().to_string();
    if change == "unchanged" || change == "none" {
        return Ok(json!({"state": change, "store": path.to_string_lossy(), "written": false}));
    }
    let next = match (&current, request.action.as_str()) {
        (Some((g, _)), "clear") => Goal {
            active: false,
            updated_at: now.into(),
            revision: g.revision + 1,
            source: request.source.clone(),
            ..g.clone()
        },
        (cur, _) => {
            let prior = cur.as_ref().map(|(g, _)| g);
            let same_words = prior.is_some_and(|g| g.active && g.text == request.text);
            Goal {
                schema_version: SCHEMA_VERSION,
                active: true,
                text: request.text.clone(),
                end_condition: request.end_condition.clone(),
                exclusions: sorted(request.exclusions.clone()),
                source: request.source.clone(),
                started_at: match prior {
                    Some(g) if same_words => g.started_at.clone(),
                    _ => now.into(),
                },
                updated_at: now.into(),
                revision: prior.map_or(1, |g| g.revision + 1),
            }
        }
    };
    let mut bytes = serde_json::to_vec_pretty(&next).map_err(|_| "encode failed".to_string())?;
    bytes.push(b'\n');
    if let Some(parent) = path.parent() {
        std::fs::create_dir_all(parent)
            .map_err(|e| format!("goal folder unavailable: {}: {e}", parent.display()))?;
    }
    match &current {
        None => publish_exclusive(&path, &bytes)?,
        Some((_, before)) => publish_replace(&path, Some(before), &bytes)?,
    }
    Ok(json!({
        "state": change,
        "store": path.to_string_lossy(),
        "written": true,
        "goal": next,
        "sha256": digest(&bytes),
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn set(text: &str) -> Request {
        Request {
            schema_version: 1,
            action: "set".into(),
            text: text.into(),
            end_condition: "no open pull request and no startable Issue".into(),
            exclusions: vec![],
            source: "user 2026-11-13T08:40Z".into(),
        }
    }

    #[test]
    fn requests_are_validated() {
        assert!(validate(&set("keep going")).is_ok());
        assert!(validate(&set("  ")).is_err());
        let mut r = set("x");
        r.action = "replace".into();
        assert!(validate(&r).is_err());
        let mut r = set("x");
        r.exclusions = vec![
            Exclusion {
                number: 7,
                reason: "HOLD".into(),
            },
            Exclusion {
                number: 7,
                reason: "again".into(),
            },
        ];
        assert!(validate(&r).unwrap_err().contains("twice"));
        let mut r = set("x");
        r.end_condition = String::new();
        assert!(validate(&r).is_err());
        let r = set(&"y".repeat(MAX_TEXT + 1));
        assert!(validate(&r).unwrap_err().contains("longer"));
        let mut r = set("x");
        r.action = "clear".into();
        assert!(validate(&r).unwrap_err().contains("clear takes"));
    }
}
