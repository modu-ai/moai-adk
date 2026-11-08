//! Project profile (`.moai-cli/project.yaml`) and its key-level reconfigure delta
//! (Issue #1995, 03-A).
//!
//! Schema v2 keeps the seven v1 keys and adds `interview_revision` and `decisions`,
//! each bounded and run through the same portable filter as the v1 values. A plan
//! decides per key with a baseline/current/desired three-way table. The baseline is
//! the profile `project-apply` last wrote, recorded in
//! `.moai-cli/manifests/profile-baseline.json`; the generation manifest records only
//! the digest, and updating it stays 03-C's job. A file without a recoverable baseline
//! is treated as user-edited throughout: any key except the MoAI-owned
//! `schema_version` lands on a conflict when the input would change it, and nothing is
//! written. General YAML (comments, anchors, block style) keeps the v1
//! behaviour: edits are refused with "existing YAML requires preservation review"
//! until the lossless-editing HOLD is decided.

use crate::{digest, PROJECT_STATE_DIR, SECRET_SHAPES};
use serde::{Deserialize, Serialize};
use serde_json::{json, Map, Value};
use std::{
    fs::{self, OpenOptions},
    io::Write,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;
const LIMIT: u64 = 65536;
/// `decisions` stays small by design (a few interview outcomes, not a transcript).
const DECISION_ITEM_LIMIT: usize = 32;
const DECISION_STRING_LIMIT: usize = 256;
const REVISION_LIMIT: usize = 128;
/// Keys owned by the profile schema; the delta covers exactly these.
const MANAGED_KEYS: [&str; 9] = [
    "schema_version",
    "project_id",
    "purpose",
    "tools",
    "development_methods",
    "hosts",
    "package_pin",
    "interview_revision",
    "decisions",
];

/// One interview decision recorded under `decisions` (REQ-03A-01). `source` says where
/// the value came from: `repo` (repository convention), `user` (stated in the
/// interview) or `default` (MoAI fallback the user accepted).
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct DecisionRecord {
    pub id: String,
    pub topic: String,
    pub value: String,
    pub source: String,
    pub scope: String,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Profile {
    pub schema_version: u32,
    pub project_id: String,
    pub purpose: String,
    pub tools: Vec<String>,
    pub development_methods: Vec<String>,
    pub hosts: Vec<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub package_pin: Option<String>,
    /// Interview revision label linking the profile to one interview run; 03-C's
    /// generation apply surfaces it in the generation manifest.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub interview_revision: Option<String>,
    /// Interview decisions 03-B consumes as its candidate-selection basis. Bounded and
    /// portable-filtered; the interview transcript itself never enters the profile.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub decisions: Vec<DecisionRecord>,
}

/// Key-level reconfigure delta of a plan (REQ-03A-02). Keys are sorted in every list so
/// the plan digest is order-stable; every managed key lands in exactly one list, and
/// conflict keys (REQ-03A-03) appear only in `conflict`.
#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct Delta {
    #[serde(default)]
    pub added: Vec<String>,
    #[serde(default)]
    pub changed: Vec<String>,
    #[serde(default)]
    pub removed: Vec<String>,
    #[serde(default)]
    pub unchanged: Vec<String>,
    #[serde(default)]
    pub conflict: Vec<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Plan {
    pub protocol_version: u32,
    pub profile: Profile,
    pub expected_sha256: Option<String>,
    pub target_binding: String,
    /// create | update | unchanged | conflict
    pub action: String,
    pub delta: Delta,
    pub plan_digest: String,
}

/// What `project-apply` last wrote, read back as the reconfigure baseline. Anything
/// unreadable, foreign or malformed means "no baseline": the three-way table then
/// treats every key as user-edited, and the plan lands on a conservative conflict.
#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
struct BaselineRecord {
    schema_version: u32,
    profile_digest: String,
    profile: Value,
}

fn valid_digest(s: &str) -> bool {
    s.len() == 71
        && s.starts_with("sha256:")
        && s[7..]
            .bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}
fn portable(s: &str) -> bool {
    let lower = s.to_ascii_lowercase();
    !s.trim().is_empty()
        && s.len() <= 2048
        && !s.chars().any(char::is_control)
        && !["/", "\\", "~/", ":\\"]
            .iter()
            .chain(SECRET_SHAPES.iter())
            .any(|x| lower.contains(x))
}
fn portable_id(s: &str) -> bool {
    !s.is_empty()
        && s.len() <= 80
        && s.bytes()
            .all(|b| b.is_ascii_alphanumeric() || b == b'-' || b == b'_')
}
fn bounded_decision_string(s: &str) -> bool {
    !s.is_empty() && s.len() <= DECISION_STRING_LIMIT && portable(s)
}

/// Accepts both schema versions so v1 files stay readable (fields default); a plan
/// input must be v2, which `plan` checks separately (REQ-03A-08).
fn validate(p: &Profile) -> Result<()> {
    let decisions_ok = p.decisions.len() <= DECISION_ITEM_LIMIT
        && p.decisions.iter().all(|d| {
            ["repo", "user", "default"].contains(&d.source.as_str())
                && portable_id(&d.id)
                && bounded_decision_string(&d.topic)
                && bounded_decision_string(&d.value)
                && bounded_decision_string(&d.scope)
        });
    if (p.schema_version != 1 && p.schema_version != 2)
        || !portable_id(&p.project_id)
        || !portable(&p.purpose)
        || p.tools.is_empty()
        || p.development_methods.is_empty()
        || p.hosts.is_empty()
        || p.tools.len() > 32
        || p.development_methods.len() > 32
        || p.hosts.len() > 2
        || !p
            .tools
            .iter()
            .chain(&p.development_methods)
            .all(|s| portable(s))
        || !p.hosts.iter().all(|s| s == "claude" || s == "codex")
        || p.package_pin.as_ref().is_some_and(|s| !valid_digest(s))
        || p.interview_revision
            .as_ref()
            .is_some_and(|s| !portable(s) || s.len() > REVISION_LIMIT)
        || !decisions_ok
    {
        return Err("invalid or nonportable profile; no values were logged".into());
    }
    Ok(())
}

/// Resolves a path under `.moai-cli/` and refuses symlinks on the state directory or
/// the path itself.
fn state_path(root: &Path, rel: &str) -> Result<PathBuf> {
    let root = root
        .canonicalize()
        .map_err(|_| "project directory unavailable")?;
    if !root.is_dir() {
        return Err("project must be a directory".into());
    }
    let dir = root.join(PROJECT_STATE_DIR);
    let full = dir.join(rel);
    for path in [&dir, &full] {
        match fs::symlink_metadata(path) {
            Ok(meta) if meta.file_type().is_symlink() => {
                return Err("symlink target refused".into())
            }
            Ok(_) => (),
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => (),
            Err(_) => return Err("target metadata unavailable".into()),
        }
    }
    Ok(full)
}
fn target(root: &Path) -> Result<PathBuf> {
    state_path(root, "project.yaml")
}
fn baseline_path(root: &Path) -> Result<PathBuf> {
    state_path(root, "manifests/profile-baseline.json")
}
pub(crate) fn read(path: &Path) -> Result<Option<Vec<u8>>> {
    match fs::metadata(path) {
        Ok(m) if !m.is_file() || m.len() > LIMIT => {
            Err("target is not a bounded regular file".into())
        }
        Ok(_) => fs::read(path)
            .map(Some)
            .map_err(|_| "target read failed".into()),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(None),
        Err(_) => Err("target metadata unavailable".into()),
    }
}
fn read_baseline(root: &Path) -> Result<Option<Value>> {
    let Ok(Some(bytes)) = read(&baseline_path(root)?) else {
        return Ok(None);
    };
    let Ok(record) = serde_json::from_slice::<BaselineRecord>(&bytes) else {
        return Ok(None);
    };
    if record.schema_version != 1
        || !valid_digest(&record.profile_digest)
        || !record.profile.is_object()
    {
        return Ok(None);
    }
    Ok(Some(record.profile))
}

#[derive(Debug, PartialEq, Eq, Clone, Copy)]
enum KeyDecision {
    KeepCurrent,
    Apply,
    Conflict,
}

/// The baseline/current/desired table from `agents.rs decide()`. Lifting that function
/// into a shared module is the 01-B refactor (#1989, open), so the semantics are
/// mirrored here on purpose. Missing values are distinct from present values, and
/// drift away from the baseline is never overwritten (REQ-03A-03).
fn decide(
    baseline: Option<&Value>,
    current: Option<&Value>,
    desired: Option<&Value>,
) -> KeyDecision {
    if current == desired || desired == baseline {
        KeyDecision::KeepCurrent
    } else if current == baseline {
        KeyDecision::Apply
    } else {
        KeyDecision::Conflict
    }
}

fn compute_delta(baseline: Option<&Value>, current: Option<&Value>, desired: &Value) -> Delta {
    let empty = Map::new();
    let desired_map = desired.as_object().unwrap_or(&empty);
    let current_map = current.and_then(Value::as_object);
    let baseline_map = baseline.and_then(Value::as_object);
    let mut delta = Delta::default();
    for key in MANAGED_KEYS {
        let d = desired_map.get(key);
        let c = current_map.and_then(|m| m.get(key));
        let b = baseline_map.and_then(|m| m.get(key));
        // schema_version is MoAI-owned schema state, never user content (leader decision
        // on the #2052 audit): a v1 file upgrades to v2 even without a baseline, while
        // every other key keeps the conservative no-baseline conflict.
        let decision = if key == "schema_version" {
            if c == d {
                KeyDecision::KeepCurrent
            } else {
                KeyDecision::Apply
            }
        } else {
            decide(b, c, d)
        };
        match decision {
            KeyDecision::KeepCurrent => delta.unchanged.push(key.into()),
            KeyDecision::Apply => match (c.is_some(), d.is_some()) {
                (false, true) => delta.added.push(key.into()),
                (true, false) => delta.removed.push(key.into()),
                _ => delta.changed.push(key.into()),
            },
            KeyDecision::Conflict => delta.conflict.push(key.into()),
        }
    }
    for list in [
        &mut delta.added,
        &mut delta.changed,
        &mut delta.removed,
        &mut delta.unchanged,
        &mut delta.conflict,
    ] {
        list.sort();
    }
    delta
}

fn plan_digest(p: &Plan) -> String {
    digest(
        &serde_json::to_vec(&(
            p.protocol_version,
            &p.profile,
            &p.expected_sha256,
            &p.target_binding,
            &p.action,
            &p.delta,
        ))
        .expect("typed plan"),
    )
}
pub fn plan(root: &Path, profile: Profile) -> Result<Plan> {
    validate(&profile)?;
    if profile.schema_version != 2 {
        return Err("invalid or nonportable profile; no values were logged".into());
    }
    let old = read(&target(root)?)?;
    let current = match &old {
        None => None,
        Some(bytes) => {
            // JSON is a YAML 1.2 subset. General YAML and comments are preserved by refusing edits (REQ-03A-04).
            let value: Value = serde_json::from_slice(bytes)
                .map_err(|_| "existing YAML requires preservation review")?;
            if !value.is_object() {
                return Err("invalid profile object".into());
            }
            Some(value)
        }
    };
    let baseline = read_baseline(root)?;
    let desired = serde_json::to_value(&profile).expect("typed profile");
    let delta = compute_delta(baseline.as_ref(), current.as_ref(), &desired);
    let action = if !delta.conflict.is_empty() {
        "conflict"
    } else if old.is_none() {
        "create"
    } else if delta.added.is_empty() && delta.changed.is_empty() && delta.removed.is_empty() {
        "unchanged"
    } else {
        "update"
    };
    let binding = digest(
        target(root)?
            .to_str()
            .ok_or("non UTF-8 project path")?
            .as_bytes(),
    );
    let mut p = Plan {
        protocol_version: 1,
        profile,
        target_binding: binding,
        expected_sha256: old.as_deref().map(digest),
        action: action.into(),
        delta,
        plan_digest: String::new(),
    };
    p.plan_digest = plan_digest(&p);
    Ok(p)
}

/// Rebuilds the profile object for an `update`: applied keys take the desired value,
/// user-preserved keys and unknown user keys stay byte-for-byte as they are.
/// `serde_json` runs with `preserve_order` (Issue #1992, leader RC1 decision H1 = B),
/// so existing keys keep their file order and keys `delta.added` appends them at the
/// end of the object.
fn splice_update(current: &[u8], desired: &Value, delta: &Delta) -> Result<Vec<u8>> {
    let mut value: Value = serde_json::from_slice(current)
        .map_err(|_| "existing YAML requires preservation review")?;
    let map = value.as_object_mut().ok_or("invalid profile object")?;
    let desired_map = desired.as_object().ok_or("invalid profile object")?;
    for key in delta.added.iter().chain(&delta.changed) {
        let v = desired_map.get(key).ok_or("invalid profile object")?;
        map.insert(key.clone(), v.clone());
    }
    for key in &delta.removed {
        map.remove(key);
    }
    let mut out = serde_json::to_vec_pretty(&value).map_err(|_| "encode failed")?;
    out.push(b'\n');
    Ok(out)
}

fn write_replace(path: &Path, before: Option<&[u8]>, after: &[u8]) -> Result<()> {
    if read(path)?.as_deref() != before {
        return Err("target changed since plan; re-preview required".into());
    }
    let parent = path.parent().ok_or("target parent unavailable")?;
    fs::create_dir_all(parent).map_err(|_| "target directory unavailable")?;
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    let temp = parent.join(format!(".moai-apply-{}-{nanos}.tmp", std::process::id()));
    let mut file = OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&temp)
        .map_err(|_| "temporary file refused")?;
    #[cfg(unix)]
    if let Ok(meta) = fs::metadata(path) {
        let _ = fs::set_permissions(&temp, meta.permissions());
    }
    let written = file
        .write_all(after)
        .and_then(|_| file.sync_all())
        .map_err(|_| "temporary write failed".to_string());
    if written.is_err() || read(path)?.as_deref() != before {
        let _ = fs::remove_file(&temp);
        return Err("target changed during apply; temporary file discarded".into());
    }
    let published = if before.is_none() {
        // Exclusive publish: a hard link fails when the target appeared meanwhile.
        fs::hard_link(&temp, path)
            .map_err(|_| "exclusive create refused")
            .map(|_| ())
    } else {
        fs::rename(&temp, path).map_err(|_| "atomic replace failed")
    };
    if before.is_none() || published.is_err() {
        let _ = fs::remove_file(&temp);
    }
    published?;
    if let Ok(dir) = fs::File::open(parent) {
        let _ = dir.sync_all();
    }
    Ok(())
}

fn write_baseline(root: &Path, bytes: &[u8], profile_value: &Value) -> Result<()> {
    let record = json!({
        "schema_version": 1,
        "profile_digest": digest(bytes),
        "profile": profile_value,
    });
    let mut out = serde_json::to_vec_pretty(&record).map_err(|_| "encode failed")?;
    out.push(b'\n');
    let path = baseline_path(root)?;
    fs::create_dir_all(path.parent().expect("baseline parent"))
        .map_err(|_| "project state directory unavailable")?;
    let before = read(&path)?;
    write_replace(&path, before.as_deref(), &out)
}

pub fn apply(root: &Path, p: Plan, confirmed: &str, permission: &str) -> Result<Value> {
    // Caller authorization remains host-owned; this reference is not an authorization verifier.
    if ["", "null", "placeholder"].contains(&permission.trim())
        || permission.len() > 128
        || permission.chars().any(char::is_control)
    {
        return Err("permission reference required".into());
    }
    let current_plan = plan(root, p.profile.clone())?;
    if p.protocol_version != 1
        || confirmed != p.plan_digest
        || plan_digest(&p) != p.plan_digest
        || current_plan.plan_digest != p.plan_digest
    {
        return Err("stale or altered plan; re-preview required".into());
    }
    if p.action == "conflict" {
        // REQ-03A-03: the user-edited keys are named, never their values; nothing is written.
        return Err(format!(
            "conflict in project.yaml: user-edited keys ({}) preserved; nothing applied",
            p.delta.conflict.join(", ")
        ));
    }
    if p.action == "unchanged" {
        let bytes = read(&target(root)?)?.unwrap_or_default();
        return Ok(json!({
            "state": "unchanged",
            "profile_digest": digest(&bytes),
            "interview_revision": p.profile.interview_revision,
            "host_verified": false
        }));
    }
    let desired = serde_json::to_value(&p.profile).expect("typed profile");
    let (bytes, state) = match p.action.as_str() {
        "create" => {
            let mut bytes = serde_json::to_vec_pretty(&p.profile).expect("typed profile");
            bytes.push(b'\n');
            (bytes, "created")
        }
        "update" => {
            // REQ-03A-05: replace only when the file still hashes to the plan's expectation.
            let current =
                read(&target(root)?)?.ok_or("stale or altered plan; re-preview required")?;
            if Some(digest(&current)) != p.expected_sha256 {
                return Err("stale or altered plan; re-preview required".into());
            }
            (splice_update(&current, &desired, &p.delta)?, "updated")
        }
        _ => return Err("unknown plan action; re-preview required".into()),
    };
    let path = target(root)?;
    fs::create_dir_all(path.parent().expect("target parent"))
        .map_err(|_| "project state directory unavailable")?;
    let path = target(root)?;
    let before = read(&path)?;
    write_replace(&path, before.as_deref(), &bytes)?;
    if let Err(e) = write_baseline(root, &bytes, &desired) {
        return Err(format!(
            "partial apply: applied [project.yaml]; failed at .moai-cli/manifests/profile-baseline.json: {e}"
        ));
    }
    Ok(json!({
        "state": state,
        "profile_digest": digest(&bytes),
        "interview_revision": p.profile.interview_revision,
        "host_verified": false,
        "next_actions": ["update profile_sha256 and interview_revision in .moai-cli/manifests/generation.json through the 03-C generation apply; project.apply does not write the manifest"]
    }))
}

/// Reads the existing profile, keeping only known keys. Unknown user keys are left in the
/// file untouched; the raw bytes are returned so callers can bind to the exact revision.
pub fn read_profile(root: &Path) -> Result<(Profile, Vec<u8>)> {
    let bytes = read(&target(root)?)?.ok_or("project profile missing")?;
    let value: Value =
        serde_json::from_slice(&bytes).map_err(|_| "existing YAML requires preservation review")?;
    let map = value.as_object().ok_or("invalid profile object")?;
    if map.get("package_pin").is_some_and(Value::is_null) {
        return Err("explicit null pin is invalid".into());
    }
    let known = [
        "schema_version",
        "project_id",
        "purpose",
        "tools",
        "development_methods",
        "hosts",
        "package_pin",
        "interview_revision",
        "decisions",
    ];
    let filtered: serde_json::Map<String, Value> = map
        .iter()
        .filter(|(k, _)| known.contains(&k.as_str()))
        .map(|(k, v)| (k.clone(), v.clone()))
        .collect();
    let profile: Profile =
        serde_json::from_value(Value::Object(filtered)).map_err(|_| "invalid profile")?;
    validate(&profile)?;
    Ok((profile, bytes))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn record(id: &str, value: &str, source: &str) -> DecisionRecord {
        DecisionRecord {
            id: id.into(),
            topic: "tools".into(),
            value: value.into(),
            source: source.into(),
            scope: "project".into(),
        }
    }

    fn v2() -> Profile {
        Profile {
            schema_version: 2,
            project_id: "demo".into(),
            purpose: "Build a small editor".into(),
            tools: vec!["Rust".into()],
            development_methods: vec!["TDD".into()],
            hosts: vec!["claude".into()],
            package_pin: None,
            interview_revision: Some("r1".into()),
            decisions: vec![record("d1", "Rust", "user")],
        }
    }

    #[test]
    fn key_decisions_mirror_the_agents_table() {
        let v = |s: &str| Value::String(s.into());
        assert_eq!(decide(None, None, Some(&v("new"))), KeyDecision::Apply);
        assert_eq!(
            decide(Some(&v("old")), Some(&v("old")), Some(&v("new"))),
            KeyDecision::Apply
        );
        // A user edit away from the baseline wins over an input that agrees with the baseline.
        assert_eq!(
            decide(Some(&v("old")), Some(&v("user")), Some(&v("old"))),
            KeyDecision::KeepCurrent
        );
        assert_eq!(
            decide(Some(&v("old")), Some(&v("new")), Some(&v("new"))),
            KeyDecision::KeepCurrent
        );
        // Without a baseline every pending change is a conflict (conservative).
        assert_eq!(
            decide(None, Some(&v("user")), Some(&v("new"))),
            KeyDecision::Conflict
        );
        assert_eq!(
            decide(Some(&v("old")), Some(&v("user")), Some(&v("new"))),
            KeyDecision::Conflict
        );
        assert_eq!(
            decide(Some(&v("old")), None, Some(&v("new"))),
            KeyDecision::Conflict
        );
        assert_eq!(
            decide(Some(&v("old")), Some(&v("old")), None),
            KeyDecision::Apply
        );
    }

    #[test]
    fn v1_and_v2_profiles_validate_but_v1_files_stay_readable() {
        let mut v1 = v2();
        v1.schema_version = 1;
        v1.interview_revision = None;
        v1.decisions = Vec::new();
        assert!(validate(&v1).is_ok());
        assert!(validate(&v2()).is_ok());
        assert!(validate(&Profile {
            schema_version: 3,
            ..v2()
        })
        .is_err());
    }

    #[test]
    fn decision_records_and_revisions_are_bounded_and_portable() {
        let mut many = v2();
        many.decisions = (0..33)
            .map(|i| record(&format!("d{i}"), "Rust", "user"))
            .collect();
        assert!(validate(&many).is_err());
        let mut secret = v2();
        secret.decisions = vec![record("d1", "password=fixture-only", "user")];
        assert!(validate(&secret).is_err());
        let mut source = v2();
        source.decisions = vec![record("d1", "Rust", "maybe")];
        assert!(validate(&source).is_err());
        let mut long = v2();
        long.decisions = vec![record("d1", &"x".repeat(DECISION_STRING_LIMIT + 1), "user")];
        assert!(validate(&long).is_err());
        let mut path = v2();
        path.decisions = vec![record("d1", "/Users/example/private", "user")];
        assert!(validate(&path).is_err());
        let mut revision = v2();
        revision.interview_revision = Some("/Users/example/private".into());
        assert!(validate(&revision).is_err());
    }

    #[test]
    fn delta_lists_are_sorted_and_categorised() {
        let baseline = serde_json::json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
            "interview_revision": "r0"
        });
        // The user edited purpose and added an unknown key after the baseline was written.
        let current = serde_json::json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Build a better editor",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
            "interview_revision": "r0", "custom": {"user": true}
        });
        let mut desired = baseline.clone();
        desired["tools"] = serde_json::json!(["Python"]);
        desired["interview_revision"] = serde_json::json!("r1");
        let delta = compute_delta(Some(&baseline), Some(&current), &desired);
        assert!(delta.added.is_empty() && delta.removed.is_empty() && delta.conflict.is_empty());
        assert_eq!(delta.changed, vec!["interview_revision", "tools"]);
        // purpose keeps the user value (desired == baseline), the unknown key is not managed.
        assert!(delta.unchanged.contains(&"purpose".to_string()));
        assert!(!delta.unchanged.contains(&"custom".to_string()));
    }
}
