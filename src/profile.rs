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

use crate::{digest, SECRET_SHAPES};
use serde::{Deserialize, Serialize};
use serde_json::{json, Map, Value};
use std::{
    fs,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;
const LIMIT: u64 = 65536;
/// `decisions` stays small by design (a few interview outcomes, not a transcript).
const DECISION_ITEM_LIMIT: usize = 32;
const DECISION_STRING_LIMIT: usize = 256;
const REVISION_LIMIT: usize = 128;
/// Keys owned by the profile schema, as `(key, additive)`: the delta covers exactly these, in
/// this order, and `read_profile` keeps exactly these. An additive key (added after the first
/// profile release) absent from desired, current and baseline is left out of the delta, so a
/// profile without it plans exactly as before the key existed. A length-free slice, so two
/// branches that each add a key merge without a silent length clash (PR #2044). Adding a key
/// touches this table, the `Profile` field and `validate` (#2254).
const PROFILE_KEYS: &[(&str, bool)] = &[
    ("schema_version", false),
    ("project_id", false),
    ("purpose", false),
    ("tools", false),
    ("development_methods", false),
    ("hosts", false),
    ("package_pin", false),
    ("interview_revision", false),
    ("decisions", false),
    ("github_issues", true),
    ("leader", true),
    ("context_max_bytes", true),
    ("archive_grace_days", true),
    ("language", true),
    ("detected", true),
];

fn managed_keys() -> impl Iterator<Item = &'static str> {
    PROFILE_KEYS.iter().map(|&(key, _)| key)
}

/// Leader supervision settings (rc.2 A3 #2161/#2162, user decision D-2 2026-11-09).
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct LeaderSettings {
    /// Seconds between leader PR checks. A user-chosen value (default 300 when absent),
    /// not a measured one; the tick itself never sleeps — scheduling is the caller's.
    pub pr_check_interval_secs: u64,
}

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
    /// The project chose GitHub Issues, so `github-plan` generates the GitHub templates
    /// when its request carries no explicit flag (Issue #2106, REQ-08B-04). Absent means
    /// no choice was recorded; a profile written before this key reads unchanged.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub github_issues: Option<bool>,
    /// Leader supervision settings (#2162). Absent means the default interval; a profile
    /// written before this key reads unchanged.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub leader: Option<LeaderSettings>,
    /// Byte budget `context-bundle` uses when a request names no `max_bytes` (#2184,
    /// REQ-C8-01). Absent means the built-in 256 KiB default; a profile written before this
    /// key reads unchanged.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub context_max_bytes: Option<u64>,
    /// SPEC auto-archive inactivity window in days (#2187, REQ-D1-01). Absent means
    /// the built-in 90-day default (design §12, Go-inherited); a profile written
    /// before this key reads unchanged.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub archive_grace_days: Option<u64>,
    /// Writing language per item (#2293): item id → language code, both checked against the
    /// tables in `crate::language`. Absent means no choice was recorded; a profile written
    /// before this key reads unchanged and the managed section renders as before.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub language: Option<crate::language::LanguageSettings>,
    /// Project facts `project-detect` found (#2496): new or existing, manifests, languages,
    /// language servers, MCP server names and databases. Absent means detection never ran; a
    /// profile written before this key reads unchanged.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub detected: Option<crate::project_detect::Detected>,
}

/// Key-level reconfigure delta of a plan (REQ-03A-02). Keys are sorted in every list so
/// the plan digest is order-stable; every managed key lands in exactly one list (except
/// `github_issues` and `leader` while absent everywhere, which land in none), and
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
    /// Managed keys whose value in the current file differs from the baseline the last
    /// `project-apply` recorded, sorted (#2590). A hand edit after an apply shows here even
    /// when the input agrees with the baseline and `action` stays `unchanged` (the edit is
    /// kept, never overwritten). Empty — and left out of the JSON — without a baseline or a
    /// file, or when the file still matches the baseline.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub drift: Vec<String>,
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
    if p.leader
        .as_ref()
        .is_some_and(|l| l.pr_check_interval_secs == 0)
    {
        return Err("leader pr_check_interval_secs must be a positive number".into());
    }
    if let Some(language) = &p.language {
        crate::language::validate(language)?;
    }
    if let Some(detected) = &p.detected {
        crate::project_detect::validate(detected)?;
    }
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
        || p.context_max_bytes == Some(0)
        || p.archive_grace_days == Some(0)
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
    let dir = crate::state_root::path(&root, "");
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

/// The baseline/current/desired table from `agents.rs decide()`. The 01-B refactor
/// (#1989) landed a shared `apply::decide`, but it compares digests; this table keeps
/// the same semantics for JSON key values, mirrored on purpose. Missing values are
/// distinct from present values, and drift away from the baseline is never overwritten
/// (REQ-03A-03).
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
    for &(key, additive) in PROFILE_KEYS {
        let d = desired_map.get(key);
        let c = current_map.and_then(|m| m.get(key));
        let b = baseline_map.and_then(|m| m.get(key));
        // An additive key absent from desired, current and baseline is left out of the delta.
        if additive && d.is_none() && c.is_none() && b.is_none() {
            continue;
        }
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

/// Managed keys where the current file and the baseline disagree (#2590). Needs both: with
/// no baseline every key is already treated as user-edited, and with no file there is
/// nothing to drift. An additive key absent from both is skipped like in the delta.
fn compute_drift(baseline: Option<&Value>, current: Option<&Value>) -> Vec<String> {
    let (Some(b), Some(c)) = (
        baseline.and_then(Value::as_object),
        current.and_then(Value::as_object),
    ) else {
        return Vec::new();
    };
    let mut drift: Vec<String> = managed_keys()
        .filter(|key| b.get(*key) != c.get(*key))
        .map(String::from)
        .collect();
    drift.sort();
    drift
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
            &p.drift,
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
    let drift = compute_drift(baseline.as_ref(), current.as_ref());
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
        drift,
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

/// Publishes `after` through the 01-B apply engine (#2071): a replace when the plan saw
/// bytes, an exclusive create when it saw none. The plan's bytes are checked here first,
/// so a change since the preview keeps its stale-plan message; the engine re-checks after
/// staging.
fn write_replace(path: &Path, before: Option<&[u8]>, after: &[u8]) -> Result<()> {
    if read(path)?.as_deref() != before {
        return Err("target changed since plan; re-preview required".into());
    }
    match before {
        Some(_) => crate::apply::publish_replace(path, before, after),
        None => crate::apply::publish_exclusive(path, after),
    }
}

fn write_baseline(
    root: &Path,
    bytes: &[u8],
    profile_value: &Value,
) -> Result<(Option<String>, String)> {
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
    let before_digest = before.as_deref().map(digest);
    write_replace(&path, before.as_deref(), &out)?;
    Ok((before_digest, digest(&out)))
}

pub fn apply(root: &Path, p: Plan, confirmed: &str, permission: &str) -> Result<Value> {
    // Caller authorization remains host-owned; this reference is not an authorization verifier.
    if !crate::contract::permission_reference_ok(permission) {
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
    // #2696 C2b: the first profile of a project with no state makes `.moai/` and its marker
    // first, so the paths below and every later writer resolve to `.moai`.
    if p.action == "create" {
        crate::state_root::adopt_new(root, "project-apply")?;
    }
    let target_rel = crate::state_root::rel(root, "project.yaml");
    let baseline_rel = crate::state_root::rel(root, "manifests/profile-baseline.json");
    if p.action == "conflict" {
        // REQ-03A-03: the user-edited keys are named, never their values; nothing is
        // written. REQ-01B-01 mirrored: the refused outcome is recorded so 04-C recovery
        // tooling can see why nothing was applied.
        let request_id = crate::apply::new_request_id("project.apply", &p.plan_digest);
        let mut receipt = crate::apply::Receipt::new(&request_id, "project.apply", root);
        receipt.entry(
            &target_rel,
            "conflict",
            p.expected_sha256.as_deref(),
            None,
            false,
            Some(format!(
                "user-edited keys ({}) preserved",
                p.delta.conflict.join(", ")
            )),
        );
        receipt.status = "refused".into();
        let mut message = format!(
            "conflict in project.yaml: user-edited keys ({}) preserved; nothing applied",
            p.delta.conflict.join(", ")
        );
        if let Err(e) = receipt.save(root) {
            message.push_str(&format!("; receipt not recorded: {e}"));
        }
        return Err(message);
    }
    let request_id = crate::apply::new_request_id("project.apply", &p.plan_digest);
    let mut receipt = crate::apply::Receipt::new(&request_id, "project.apply", root);
    if p.action == "unchanged" {
        // REQ-01B-10: a re-run writes nothing; the receipt records zero publishes.
        let bytes = read(&target(root)?)?.unwrap_or_default();
        receipt.entry(
            &target_rel,
            "unchanged",
            Some(&digest(&bytes)),
            Some(&digest(&bytes)),
            false,
            None,
        );
        receipt.status = "complete".into();
        receipt
            .save(root)
            .map_err(|e| format!("receipt not recorded: {e}"))?;
        return Ok(json!({
            "state": "unchanged",
            "profile_digest": digest(&bytes),
            "interview_revision": p.profile.interview_revision,
            "host_verified": false,
            "receipt": request_id
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
    let before_digest = before.as_deref().map(digest);
    // REQ-01B-08: keep the before bytes before publishing so recovery can restore
    // content; a created file needs no backup (recovery removes it).
    if let Some(b) = before.as_deref() {
        if let Err(e) = crate::apply::save_before_bytes(root, &request_id, &target_rel, b) {
            receipt.entry(
                &target_rel,
                &p.action,
                before_digest.as_deref(),
                None,
                false,
                Some(e.clone()),
            );
            receipt.status = "partial".into();
            let message = format!("backup of {target_rel} before publish failed: {e}");
            return Err(match receipt.save(root) {
                Ok(()) => message,
                Err(se) => format!("{message}; receipt not recorded: {se}"),
            });
        }
    }
    if let Err(e) = write_replace(&path, before.as_deref(), &bytes) {
        receipt.entry(
            &target_rel,
            &p.action,
            before_digest.as_deref(),
            None,
            false,
            Some(e.clone()),
        );
        receipt.status = "partial".into();
        return Err(match receipt.save(root) {
            Ok(()) => e,
            Err(se) => format!("{e}; receipt not recorded: {se}"),
        });
    }
    receipt.entry(
        &target_rel,
        &p.action,
        before_digest.as_deref(),
        Some(&digest(&bytes)),
        true,
        None,
    );
    // The baseline record lives under manifests/ and coexists with the 01-B receipt
    // store under receipts/. REQ-01B-08 mirrored for the record itself: a replaced
    // baseline keeps its before bytes, so recovery restores the whole pre-update state
    // instead of stopping halfway between project.yaml and the baseline.
    let baseline_before_bytes = read(&baseline_path(root)?).ok().flatten();
    if let Some(b) = baseline_before_bytes.as_deref() {
        if let Err(e) = crate::apply::save_before_bytes(root, &request_id, &baseline_rel, b) {
            receipt.entry(&baseline_rel, "replace", None, None, false, Some(e.clone()));
            receipt.status = "partial".into();
            let message = format!(
                "partial apply: applied [project.yaml]; failed at {baseline_rel} backup: {e}"
            );
            return Err(match receipt.save(root) {
                Ok(()) => message,
                Err(se) => format!("{message}; receipt not recorded: {se}"),
            });
        }
    }
    match write_baseline(root, &bytes, &desired) {
        Ok((baseline_before_digest, after_baseline)) => {
            receipt.entry(
                &baseline_rel,
                if baseline_before_digest.is_some() {
                    "replace"
                } else {
                    "create"
                },
                baseline_before_digest.as_deref(),
                Some(&after_baseline),
                true,
                None,
            );
            receipt.status = "complete".into();
        }
        Err(e) => {
            receipt.entry(&baseline_rel, "create", None, None, false, Some(e.clone()));
            receipt.status = "partial".into();
            let message =
                format!("partial apply: applied [project.yaml]; failed at {baseline_rel}: {e}");
            return Err(match receipt.save(root) {
                Ok(()) => message,
                Err(se) => format!("{message}; receipt not recorded: {se}"),
            });
        }
    }
    if let Err(e) = receipt.save(root) {
        return Err(format!(
            "published {target_rel} but its receipt was not recorded: {e}"
        ));
    }
    Ok(json!({
        "state": state,
        "profile_digest": digest(&bytes),
        "interview_revision": p.profile.interview_revision,
        "host_verified": false,
        "receipt": request_id,
        "next_actions": [format!(
            "update profile_sha256 and interview_revision in {} through the 03-C generation apply; project.apply does not write the manifest",
            crate::state_root::rel(root, "manifests/generation.json")
        )]
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
    let filtered: serde_json::Map<String, Value> = map
        .iter()
        .filter(|(k, _)| managed_keys().any(|m| m == k.as_str()))
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
            github_issues: None,
            leader: None,
            context_max_bytes: None,
            archive_grace_days: None,
            language: None,
            detected: None,
        }
    }

    /// `PROFILE_KEYS` lists exactly the `Profile` fields, and the additive ones are exactly
    /// the fields a v2 profile may omit (#2254: the table replaced three hand-kept lists).
    #[test]
    fn profile_keys_match_the_profile_fields() {
        let full = Profile {
            package_pin: Some("pin".into()),
            github_issues: Some(true),
            leader: Some(LeaderSettings {
                pr_check_interval_secs: 300,
            }),
            context_max_bytes: Some(1),
            archive_grace_days: Some(7),
            language: Some([("code_comments".to_string(), "en".to_string())].into()),
            detected: Some(crate::project_detect::Detected {
                project_kind: "new".into(),
                ..Default::default()
            }),
            ..v2()
        };
        let fields: Vec<String> = serde_json::to_value(&full)
            .unwrap()
            .as_object()
            .unwrap()
            .keys()
            .cloned()
            .collect();
        let mut keys: Vec<String> = managed_keys().map(String::from).collect();
        let mut sorted = fields.clone();
        sorted.sort();
        keys.sort();
        assert_eq!(keys, sorted);
        let additive: Vec<&str> = PROFILE_KEYS
            .iter()
            .filter(|(_, a)| *a)
            .map(|(k, _)| *k)
            .collect();
        assert_eq!(
            additive,
            [
                "github_issues",
                "leader",
                "context_max_bytes",
                "archive_grace_days",
                "language",
                "detected"
            ]
        );
    }

    /// AC-1 (#2293): `language` is additive, and an unknown item or an unsupported code is
    /// refused by `validate` and by `read_profile` on a hand-edited file.
    #[test]
    fn language_key_is_additive_and_validated() {
        let base = serde_json::json!({
            "schema_version": 2, "project_id": "demo", "purpose": "p",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"]
        });
        let all = |d: &Delta| {
            [&d.added, &d.changed, &d.removed, &d.unchanged, &d.conflict]
                .iter()
                .any(|l| l.contains(&"language".to_string()))
        };
        assert!(!all(&compute_delta(Some(&base), Some(&base), &base)));
        let mut desired = base.clone();
        desired["language"] = serde_json::json!({"code_comments": "en"});
        assert_eq!(
            compute_delta(Some(&base), Some(&base), &desired).added,
            vec!["language"]
        );
        let with = |pairs: &[(&str, &str)]| Profile {
            language: Some(
                pairs
                    .iter()
                    .map(|(k, v)| (k.to_string(), v.to_string()))
                    .collect(),
            ),
            ..v2()
        };
        assert!(validate(&with(&[("code_comments", "ko")])).is_ok());
        assert!(validate(&with(&[("code_coments", "ko")])).is_err());
        assert!(validate(&with(&[("code_comments", "kr")])).is_err());
        assert!(validate(&with(&[])).is_err());

        let tmp = tempfile::tempdir().unwrap();
        let path = tmp
            .path()
            .join(crate::PROJECT_STATE_DIR)
            .join("project.yaml");
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        for (bad, _) in [
            (
                json!({"code_comments": "en", "emoji": "en"}),
                "unknown item",
            ),
            (json!({"code_comments": "english"}), "unsupported code"),
            (json!({"code_comments": 1}), "non-string code"),
        ] {
            let mut file = serde_json::to_value(v2()).unwrap();
            file["language"] = bad;
            fs::write(&path, serde_json::to_vec(&file).unwrap()).unwrap();
            assert!(read_profile(tmp.path()).is_err());
        }
    }

    /// `read_profile` keeps the `PROFILE_KEYS` and leaves a user's own key out, so the
    /// strict `Profile` still reads a file that carries one.
    #[test]
    fn read_profile_skips_keys_outside_the_table() {
        let tmp = tempfile::tempdir().unwrap();
        let mut file = serde_json::to_value(v2()).unwrap();
        file["my_own_key"] = json!("kept in the file");
        let path = tmp
            .path()
            .join(crate::PROJECT_STATE_DIR)
            .join("project.yaml");
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(&path, serde_json::to_vec(&file).unwrap()).unwrap();
        let (profile, bytes) = read_profile(tmp.path()).unwrap();
        assert_eq!(profile.project_id, "demo");
        assert!(String::from_utf8(bytes).unwrap().contains("my_own_key"));
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

    /// AC-3 (#2162): `leader` is a managed key with the same additive rule as
    /// `github_issues`, and its interval must be positive.
    #[test]
    fn leader_key_is_managed_additive_and_validated() {
        assert!(PROFILE_KEYS.contains(&("leader", true)));
        let base = serde_json::json!({
            "schema_version": 2, "project_id": "demo", "purpose": "p",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"]
        });
        let all = |d: &Delta| {
            [&d.added, &d.changed, &d.removed, &d.unchanged, &d.conflict]
                .iter()
                .any(|l| l.contains(&"leader".to_string()))
        };
        // Absent everywhere: in no list, so an old profile plans exactly as before.
        assert!(!all(&compute_delta(Some(&base), Some(&base), &base)));
        let mut desired = base.clone();
        desired["leader"] = serde_json::json!({"pr_check_interval_secs": 60});
        let delta = compute_delta(Some(&base), Some(&base), &desired);
        assert_eq!(delta.added, vec!["leader"]);
        let mut p = v2();
        p.leader = Some(LeaderSettings {
            pr_check_interval_secs: 60,
        });
        assert!(validate(&p).is_ok());
        p.leader = Some(LeaderSettings {
            pr_check_interval_secs: 0,
        });
        assert_eq!(
            validate(&p).unwrap_err(),
            "leader pr_check_interval_secs must be a positive number"
        );
    }
}
