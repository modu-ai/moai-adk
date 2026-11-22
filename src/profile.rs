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

use crate::digest;
use serde::{Deserialize, Serialize};
use serde_json::{json, Map, Value};
use std::{
    fs,
    io::Read,
    path::{Path, PathBuf},
    process::{Command, Stdio},
    sync::mpsc,
    time::{Duration, Instant},
};

type Result<T> = std::result::Result<T, String>;
/// The largest file [`read`] accepts, in bytes. Every skill package file is read through it
/// (`package_digest`: install, update, `package-inspect`), so it bounds each package file too.
pub const LIMIT: u64 = 65536;
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
    ("scopes", true),
    ("versioning", true),
    ("work_flow", true),
    ("remote_copy", true),
    ("test_paths", true),
    ("ci", true),
    ("checks", true),
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
    /// The project's fixed id (#3009 S1, decision A 6889383420): the name the HOME store is
    /// keyed by. A request that leaves it out (or empty) gets the file's existing id, or for a
    /// profile that does not exist yet a random one from [`crate::project_id::new_id`]; [`plan`]
    /// fills it in, so a plan and its apply always carry the same id. An existing id, such as
    /// `demo`, is never replaced.
    #[serde(default)]
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
    /// Worktree scopes (#3010): named lists of repository folders, an allow list that a later
    /// piece reads (the scoped install preview; rc.2 does no sparse checkout, user decision
    /// 6892968564). Absent means no scope was declared and a worktree is the whole
    /// repository, as before; a profile written before this key reads unchanged.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub scopes: Option<Vec<crate::scopes::ScopeEntry>>,
    /// The versioning scheme the project's releases follow (card C-69): `semver`, `headver`,
    /// `calver` or `custom`, with the CalVer format or the HeadVer head when chosen. Absent
    /// means the interview recorded no choice; a profile written before this key reads
    /// unchanged and the managed section renders as before.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub versioning: Option<crate::versioning::VersioningSettings>,
    /// The git work flow the project follows (card C-70): `m-flow` (MoAI's own, the
    /// recommended default), `github-flow`, `trunk-based` or `git-flow`, with the integration
    /// branch when it was named. Absent means the interview recorded no choice; a profile
    /// written before this key reads unchanged.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub work_flow: Option<WorkFlowSettings>,
    /// Whether the landing's remote copy is on: the leader turns it on only after a preview
    /// and the user's yes (decision of 2026-11-22); absent or false means the copy pushes
    /// nothing. A bool needs no validator beyond its type: a non-boolean value fails the
    /// read, and the key is listed in [`PROFILE_KEYS`] so it is a managed key.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub remote_copy: Option<bool>,
    /// The project's own test paths beside the per-language table of `crate::tdd`
    /// (SPEC-20261117-green-test-lock D-4): relative folders or files, at most
    /// [`crate::tdd::MAX_DECLARED`]. Absent means none were declared; a profile written before
    /// this key reads unchanged.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub test_paths: Option<Vec<String>>,
    /// The saved answers of the plain CI questions (SPEC-20261122-project-ci-design): what the
    /// project is, which systems its users run, whether and where it deploys, public or
    /// private with the monthly budget, and how strict the checks are. Absent means the
    /// interview recorded none; a profile written before this key reads unchanged.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub ci: Option<crate::project_ci::CiAnswers>,
    /// The check commands the project declares, in the order they run: the one list the
    /// generated workflow and the project's local check both read. Absent means none were
    /// declared; a profile written before this key reads unchanged.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub checks: Option<Vec<crate::project_ci::CheckEntry>>,
}

/// The flows the `work_flow` key names (card C-70), m-flow first.
pub const WORK_FLOWS: &[&str] = &["m-flow", "github-flow", "trunk-based", "git-flow"];

/// The `work_flow` key: one of [`WORK_FLOWS`] and, when named, the integration branch every
/// card branch starts from.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct WorkFlowSettings {
    pub flow: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub integration_branch: Option<String>,
}

/// A branch name git takes, kept to a portable subset: 1 to 100 bytes of ASCII letters, digits
/// and `. - _ /`; no leading `-`, `.` or `/`; no trailing `/` or `.`; no `..`, `//` or `/.`;
/// no `/` component ending in `.lock`; not `HEAD` (as `git check-ref-format --branch` rules).
pub(crate) fn branch_name(s: &str) -> bool {
    !s.is_empty()
        && s.len() <= 100
        && s != "HEAD"
        && s.bytes()
            .all(|b| b.is_ascii_alphanumeric() || matches!(b, b'.' | b'-' | b'_' | b'/'))
        && !s.starts_with(['-', '.', '/'])
        && !s.ends_with(['/', '.'])
        && !s.split('/').any(|part| part.ends_with(".lock"))
        && !["..", "//", "/."].iter().any(|x| s.contains(x))
}

/// What a first launch writes into `tools` and `development_methods` until the project
/// interview replaces it (`launcher_init::default_profile`).
const UNSPECIFIED: &str = "unspecified";

impl Profile {
    /// Whether this is still the placeholder a first launch writes (#2936 R3): `tools` and
    /// `development_methods` are both exactly `["unspecified"]` and no decision came from the
    /// user or the repository. One answered axis or one such decision makes it an interviewed
    /// profile.
    pub fn is_placeholder(&self) -> bool {
        let unspecified = |v: &[String]| v.len() == 1 && v[0] == UNSPECIFIED;
        unspecified(&self.tools)
            && unspecified(&self.development_methods)
            && self.decisions.iter().all(|d| d.source == "default")
    }
}

/// The refusal for a placeholder profile (#2936 R3); its prefix maps to the `blocked` kind and
/// the message names the way out, so no second `next:` line is added.
pub fn placeholder_refusal() -> String {
    "profile blocked: the project profile is still the placeholder a first launch writes \
     (tools and development methods unspecified, no interview answer); run `/moai project` \
     to answer the interview, then plan again"
        .into()
}

/// Refuses a project whose profile is still the placeholder (#2936 R3). A missing or unreadable
/// profile passes: the command that follows reports that itself, as before.
pub fn require_interviewed(root: &Path) -> Result<()> {
    match read_profile(root) {
        Ok((profile, _)) if profile.is_placeholder() => Err(placeholder_refusal()),
        _ => Ok(()),
    }
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
        && !["/", "\\", "~/", ":\\"].iter().any(|x| lower.contains(x))
        // Card C-90: `sk-` only at the start of a word, so `task-based` is ordinary text.
        && !crate::secret_shaped(s)
}
pub(crate) fn portable_id(s: &str) -> bool {
    !s.is_empty()
        && s.len() <= 80
        && s.bytes()
            .all(|b| b.is_ascii_alphanumeric() || b == b'-' || b == b'_')
}
fn bounded_decision_string(s: &str) -> bool {
    !s.is_empty() && s.len() <= DECISION_STRING_LIMIT && portable(s)
}

/// The hosts a profile may list (SPEC-20261118-profile-host-omp); one more entry than these is a
/// duplicate or an unknown name.
pub const KNOWN_HOSTS: [&str; 3] = ["claude", "codex", "omp"];

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
    if let Some(scopes) = &p.scopes {
        crate::scopes::validate(scopes)?;
    }
    if let Some(w) = &p.work_flow {
        if !WORK_FLOWS.contains(&w.flow.as_str())
            || w.integration_branch
                .as_deref()
                .is_some_and(|b| !branch_name(b))
        {
            return Err("invalid work_flow setting; no values were logged".into());
        }
    }
    if let Some(versioning) = &p.versioning {
        crate::versioning::validate(versioning)?;
        // The format is text like every other profile string: no path, no secret shape.
        if versioning.format.as_ref().is_some_and(|f| !portable(f)) {
            return Err("invalid or nonportable profile; no values were logged".into());
        }
    }
    if let Some(paths) = &p.test_paths {
        crate::tdd::validate_declared(paths)?;
    }
    if let Some(answers) = &p.ci {
        crate::project_ci::validate_answers(answers)?;
    }
    if let Some(checks) = &p.checks {
        crate::project_ci::validate_checks(checks)?;
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
        || p.hosts.len() > KNOWN_HOSTS.len()
        || !p
            .tools
            .iter()
            .chain(&p.development_methods)
            .all(|s| portable(s))
        || !p.hosts.iter().all(|s| KNOWN_HOSTS.contains(&s.as_str()))
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

/// The words `read_profile` fails with when the project has no profile file at all.
pub const PROFILE_MISSING: &str = "project profile missing";

/// The folder (`.moai` or `.moai-cli`) whose `project.yaml` a READER opens
/// (SPEC-20261121-mixed-profile-read). A writer follows the read in a mixed layout (the delta
/// audit's B4): `project-apply` updates the newer profile and the older one is left untouched,
/// so what a write changes is what every reader then sees. A reader takes the newer layout's
/// profile when both `.moai/project.yaml` and `.moai-cli/project.yaml` are regular files
/// (neither the file nor the `.moai` folder a link) and the layout is not yet `rust` (its
/// marker may be missing, e.g. a worktree where the marker is ignored by git); otherwise the
/// layout's own folder, exactly as before.
fn read_dir(root: &Path) -> Result<&'static str> {
    let root = root
        .canonicalize()
        .map_err(|_| "project directory unavailable")?;
    let own = crate::state_root::dir(&root);
    let regular = |dir: &str| {
        let folder = root.join(dir);
        fs::symlink_metadata(&folder).is_ok_and(|m| m.file_type().is_dir())
            && fs::symlink_metadata(folder.join("project.yaml"))
                .is_ok_and(|m| m.file_type().is_file())
    };
    if own != crate::state_root::MOAI_DIR
        && regular(crate::state_root::MOAI_DIR)
        && regular(crate::state_root::EARLIER_PROJECT_DIR)
    {
        Ok(crate::state_root::MOAI_DIR)
    } else {
        Ok(own)
    }
}

/// The project-relative path of the profile a reader opens, for messages (`.moai/project.yaml`).
pub fn read_rel(root: &Path) -> String {
    format!(
        "{}/project.yaml",
        read_dir(root).unwrap_or(crate::state_root::EARLIER_PROJECT_DIR)
    )
}

fn read_target(root: &Path) -> Result<PathBuf> {
    // `target` keeps the symlink and directory refusals of the layout's own folder.
    let own = target(root)?;
    let dir = read_dir(root)?;
    if dir == crate::state_root::MOAI_DIR && crate::state_root::dir(root) != dir {
        Ok(root
            .canonicalize()
            .map_err(|_| "project directory unavailable")?
            .join(dir)
            .join("project.yaml"))
    } else {
        Ok(own)
    }
}

/// Whether a reader — and, since the delta audit's B4, a writer — picks the newer `.moai`
/// profile: both profiles are regular files and the layout is not `rust` (the mixed branch of
/// [`read_dir`]). Recovery uses it to keep exactly the profile path a mixed-layout apply wrote
/// (src/apply.rs).
pub fn reads_newer_profile(root: &Path) -> bool {
    crate::state_root::dir(root) != crate::state_root::MOAI_DIR
        && read_dir(root).is_ok_and(|d| d == crate::state_root::MOAI_DIR)
}

/// The bytes of the profile a reader opens ([`read_target`]): `Ok(None)` when the chosen file
/// is not there. The existence checks of readers that only want to know whether a profile is
/// there use this, so they look at the folder that is actually read (the delta audit's N2).
pub fn chosen_profile(root: &Path) -> Result<Option<Vec<u8>>> {
    profile_bytes(root)
}

/// The bytes at [`read_target`], with the read's own errors named by the file (`target read
/// failed` and `target metadata unavailable` said neither before — the delta audit's R2).
fn profile_bytes(root: &Path) -> Result<Option<Vec<u8>>> {
    read(&read_target(root)?).map_err(|e| format!("{}: {e}", read_rel(root)))
}

/// In a mixed layout both profiles belong to one check: each must read and parse (a failure is
/// named `{folder}/project.yaml: reason`) and both must name the same `project_id` — anything
/// else refuses the read, naming both files and both ids (the delta audit's N3). One profile,
/// or two that agree, is fine; a rust layout owns `.moai` outright and has no twin to check.
fn mixed_layout_ids(root: &Path) -> Result<()> {
    let root = root
        .canonicalize()
        .map_err(|_| "project directory unavailable")?;
    // A rust layout owns `.moai` outright: the earlier folder is legacy data, not a twin, and
    // the check lives only where both files exist (the mixed branch of `read_dir`).
    if crate::state_root::dir(&root) == crate::state_root::MOAI_DIR {
        return Ok(());
    }
    let regular = |dir: &str| {
        let folder = root.join(dir);
        fs::symlink_metadata(&folder).is_ok_and(|m| m.file_type().is_dir())
            && fs::symlink_metadata(folder.join("project.yaml"))
                .is_ok_and(|m| m.file_type().is_file())
    };
    if !(regular(crate::state_root::MOAI_DIR) && regular(crate::state_root::EARLIER_PROJECT_DIR)) {
        return Ok(());
    }
    let mut ids: Vec<(&'static str, String)> = Vec::new();
    for name in [
        crate::state_root::MOAI_DIR,
        crate::state_root::EARLIER_PROJECT_DIR,
    ] {
        let path = root.join(name).join("project.yaml");
        let bytes = read(&path)
            .map_err(|e| format!("{name}/project.yaml: {e}"))?
            .ok_or_else(|| format!("{name}/project.yaml: {PROFILE_MISSING}"))?;
        let id = crate::project_id::of_profile_bytes(&bytes)
            .map_err(|e| format!("{name}/project.yaml: {e}"))?;
        ids.push((name, id));
    }
    match ids.as_slice() {
        [(a_name, a), (b_name, b)] if a != b => Err(format!(
            "the two profiles name different projects: {a_name}/project.yaml has `{a}`, \
             {b_name}/project.yaml has `{b}`"
        )),
        _ => Ok(()),
    }
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

/// Whether the saved `hosts` still hold a retired host and the desired list is exactly the saved
/// one without it: such a key is MoAI-owned state moving off a retired host, never a user edit.
fn retired_hosts_update(active: &[&str], current: Option<&Value>, desired: Option<&Value>) -> bool {
    let names = |v: Option<&Value>| -> Option<Vec<String>> {
        v?.as_array()?
            .iter()
            .map(|h| h.as_str().map(str::to_string))
            .collect()
    };
    let (Some(saved), Some(wanted)) = (names(current), names(desired)) else {
        return false;
    };
    let mut filtered = saved.clone();
    crate::retired_hosts::retire_hosts(&mut filtered, active);
    filtered != saved && filtered == wanted
}

fn compute_delta(
    active: &[&str],
    baseline: Option<&Value>,
    current: Option<&Value>,
    desired: &Value,
) -> Delta {
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
        } else if key == "hosts" && retired_hosts_update(active, c, d) {
            KeyDecision::Apply
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
    plan_in(&KNOWN_HOSTS, root, profile)
}

/// [`plan`] for a place whose active hosts are `active`: retired hosts in the input are filtered
/// like a read, and a saved profile that still holds one gets the filtered `hosts` as an update.
pub fn plan_in(active: &[&str], root: &Path, mut profile: Profile) -> Result<Plan> {
    crate::retired_hosts::retire_hosts(&mut profile.hosts, active);
    if profile.project_id.is_empty() {
        profile.project_id = match existing_project_id(root)? {
            Some(id) => id,
            None => crate::project_id::new_id()?,
        };
    }
    validate(&profile)?;
    if profile.schema_version != 2 {
        return Err("invalid or nonportable profile; no values were logged".into());
    }
    let old = profile_bytes(root)?;
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
    let delta = compute_delta(active, baseline.as_ref(), current.as_ref(), &desired);
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
        read_target(root)?
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

/// C-105 slice 2 (SPEC-20261119-plan-remote-branch): whether the profile's integration
/// branch exists on the project's remote, and how far it sits behind the remote's default
/// branch. Called by the `project-plan` entry only — `project-update-plan` shares `plan()`
/// but must not run `git ls-remote` (a network call) in its preview (audit B4). A name that
/// must never reach git as an argument (leading `-`, whitespace, `..`, over 100 bytes,
/// `HEAD`) is refused without running any git command (the profile validation refuses it
/// first — this is the second layer for internal callers); a repository or a remote that is
/// missing answers `None` (the plan output stays as it was); a git failure becomes an
/// `unavailable` warning and the plan still succeeds. The remote check never prompts on
/// the terminal (`GIT_TERMINAL_PROMPT=0`) and never waits forever: the wait is bounded
/// (audit B6) and the child's process group is killed when it lapses; tests shrink the
/// bound with `MOAI_GIT_TIMEOUT_SECS`, the same injection shape as `MOAI_FORGE_GH_BIN`.
/// Only fixed words and verified 40-hex
/// SHAs travel as git arguments — the branch name never does.
pub fn plan_work_flow_remote(root: &Path, work_flow: Option<&WorkFlowSettings>) -> Option<Value> {
    let branch = work_flow?.integration_branch.as_deref()?;
    if !branch_name(branch) {
        return Some(json!({
            "on_remote": false,
            "warnings": [format!(
                "refused: the integration branch is not a safe branch name; no git command ran for it"
            )],
        }));
    }
    let git = |args: &[&str]| -> std::result::Result<String, String> {
        let timeout = std::env::var("MOAI_GIT_TIMEOUT_SECS")
            .ok()
            .and_then(|v| v.parse::<u64>().ok())
            .unwrap_or(10);
        let mut cmd = Command::new("git");
        crate::identity::scrub_git_env(&mut cmd);
        // Its own process group, so everything a remote check starts can be ended with it.
        #[cfg(unix)]
        std::os::unix::process::CommandExt::process_group(&mut cmd, 0);
        let mut child = cmd
            .env("GIT_TERMINAL_PROMPT", "0")
            .arg("-C")
            .arg(root)
            .args(args)
            .stdin(Stdio::null())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .map_err(|e| format!("cannot run git: {e}"))?;
        // Readers hand their bytes over a channel, so a full pipe cannot deadlock the wait.
        let drain = |pipe: Option<Box<dyn Read + Send>>| {
            let (tx, rx) = mpsc::channel();
            std::thread::spawn(move || {
                let mut buf = Vec::new();
                if let Some(mut p) = pipe {
                    let _ = p.read_to_end(&mut buf);
                }
                let _ = tx.send(buf);
            });
            rx
        };
        let stdout = drain(
            child
                .stdout
                .take()
                .map(|p| Box::new(p) as Box<dyn Read + Send>),
        );
        let stderr = drain(
            child
                .stderr
                .take()
                .map(|p| Box::new(p) as Box<dyn Read + Send>),
        );
        let deadline = Instant::now() + Duration::from_secs(timeout);
        let status = loop {
            match child.try_wait() {
                Ok(Some(s)) => break Some(s),
                Ok(None) if Instant::now() >= deadline => {
                    // The group goes with it, so an ssh or helper the remote check started
                    // cannot outlive the timeout holding the pipes.
                    #[cfg(unix)]
                    {
                        let _ = Command::new("kill")
                            .args(["-KILL", "--", &format!("-{}", child.id())])
                            .stdin(Stdio::null())
                            .stdout(Stdio::null())
                            .stderr(Stdio::null())
                            .status();
                    }
                    let _ = child.kill();
                    let _ = child.wait();
                    return Err(format!("timed out after {timeout}s"));
                }
                Ok(None) => std::thread::sleep(Duration::from_millis(10)),
                Err(_) => return Err("git wait failed".to_string()),
            }
        };
        let grace = Duration::from_millis(500);
        let out = stdout.recv_timeout(grace).unwrap_or_default();
        let err_out = stderr.recv_timeout(grace).unwrap_or_default();
        let Some(s) = status else {
            return Err("git wait failed".to_string());
        };
        if s.success() {
            Ok(String::from_utf8_lossy(&out).into_owned())
        } else {
            Err(String::from_utf8_lossy(&err_out).trim().to_string())
        }
    };
    // REQ-5: nothing to check without a repository and a remote.
    let has_remote = git(&["remote"])
        .map(|s| !s.trim().is_empty())
        .unwrap_or(false);
    if !has_remote {
        return None;
    }
    let ls = match git(&["ls-remote", "--symref", "origin"]) {
        Ok(s) => s,
        Err(e) => {
            return Some(json!({
                "on_remote": false,
                "warnings": [format!("unavailable: git ls-remote failed: {e}")],
            }))
        }
    };
    let is_sha = |s: &str| s.len() == 40 && s.bytes().all(|b| b.is_ascii_hexdigit());
    let mut on_remote = false;
    let mut default_target: Option<String> = None;
    let mut default_sha: Option<&str> = None;
    let mut branch_sha: Option<&str> = None;
    for line in ls.lines() {
        // `--symref` prints the symref first, as `ref: <target>\tHEAD`; the ref lines follow
        // as `<sha>\t<ref>`.
        let Some((value, r)) = line.split_once('\t') else {
            continue;
        };
        if r == "HEAD" {
            // The symref line (`ref: <target>\tHEAD`) names the default branch; the plain
            // `HEAD` sha line after it carries no target and must not reset ours.
            if let Some(target) = value.strip_prefix("ref: ") {
                default_target = Some(target.to_string());
            }
        } else if r == format!("refs/heads/{branch}") && is_sha(value) {
            on_remote = true;
            branch_sha = Some(value);
        } else if default_target.as_deref() == Some(r) && is_sha(value) {
            default_sha = Some(value);
        }
    }
    let mut warnings: Vec<String> = Vec::new();
    if on_remote {
        // Audit B1: a measurement that cannot run is a warning, never silence — a stale
        // local clone (the objects missing) used to read as a healthy branch.
        if let (Some(b), Some(d)) = (branch_sha, default_sha) {
            match git(&["rev-list", "--count", &format!("{b}..{d}")]) {
                Ok(count) => match count.trim().parse::<u64>() {
                    Ok(n) if n > 0 => warnings.push(format!(
                        "behind: {branch} is {n} commits behind the remote default branch"
                    )),
                    Ok(_) => {}
                    Err(_) => warnings.push(
                        "unavailable: could not read the behind count from git rev-list output"
                            .to_string(),
                    ),
                },
                Err(e) => warnings.push(format!(
                    "unavailable: could not measure how far {branch} is behind: {e}"
                )),
            }
        }
    } else {
        warnings.push(format!(
            "missing: {branch} does not exist on the remote the branches go to"
        ));
    }
    Some(json!({"on_remote": on_remote, "warnings": warnings}))
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
    apply_in(&KNOWN_HOSTS, root, p, confirmed, permission)
}

/// [`apply`] for a place whose active hosts are `active`.
pub fn apply_in(
    active: &[&str],
    root: &Path,
    p: Plan,
    confirmed: &str,
    permission: &str,
) -> Result<Value> {
    // Caller authorization remains host-owned; this reference is not an authorization verifier.
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    let current_plan = plan_in(active, root, p.profile.clone())?;
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
    // Writes follow reads in a mixed layout (the delta audit's B4): the receipt names the
    // profile that was actually written.
    let target_rel = read_rel(root);
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
        let bytes = profile_bytes(root)?.unwrap_or_default();
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
                profile_bytes(root)?.ok_or("stale or altered plan; re-preview required")?;
            if Some(digest(&current)) != p.expected_sha256 {
                return Err("stale or altered plan; re-preview required".into());
            }
            (splice_update(&current, &desired, &p.delta)?, "updated")
        }
        _ => return Err("unknown plan action; re-preview required".into()),
    };
    let path = read_target(root)?;
    fs::create_dir_all(path.parent().expect("target parent"))
        .map_err(|_| "project state directory unavailable")?;
    let path = read_target(root)?;
    let before = read(&path).map_err(|e| format!("{}: {e}", read_rel(root)))?;
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

/// The `project_id` of the profile file that is already there: `Ok(None)` when the project
/// has no profile, an error when it has one that is not a JSON object with a usable
/// `project_id` (the file is left alone). Reads only.
pub(crate) fn existing_project_id(root: &Path) -> Result<Option<String>> {
    mixed_layout_ids(root)?;
    match profile_bytes(root)? {
        None => Ok(None),
        Some(bytes) => crate::project_id::of_profile_bytes(&bytes)
            .map(Some)
            .map_err(|e| format!("{}: {e}", read_rel(root))),
    }
}

/// Reads the existing profile, keeping only known keys. Unknown user keys are left in the
/// file untouched; the raw bytes are returned so callers can bind to the exact revision.
pub fn read_profile(root: &Path) -> Result<(Profile, Vec<u8>)> {
    read_profile_in(&KNOWN_HOSTS, root)
}

/// [`read_profile`] for a place whose active hosts are `active`: retired hosts leave `hosts`, and
/// a list they emptied reads as `["claude"]` (the file is not changed by a read).
pub fn read_profile_in(active: &[&str], root: &Path) -> Result<(Profile, Vec<u8>)> {
    mixed_layout_ids(root)?;
    let bytes = profile_bytes(root)?.ok_or(PROFILE_MISSING)?;
    let value: Value = serde_json::from_slice(&bytes).map_err(|_| {
        format!(
            "{}: existing YAML requires preservation review",
            read_rel(root)
        )
    })?;
    let map = value
        .as_object()
        .ok_or_else(|| format!("{}: invalid profile object", read_rel(root)))?;
    if map.get("package_pin").is_some_and(Value::is_null) {
        return Err("explicit null pin is invalid".into());
    }
    let filtered: serde_json::Map<String, Value> = map
        .iter()
        .filter(|(k, _)| managed_keys().any(|m| m == k.as_str()))
        .map(|(k, v)| (k.clone(), v.clone()))
        .collect();
    let mut profile: Profile =
        serde_json::from_value(Value::Object(filtered)).map_err(|_| "invalid profile")?;
    crate::retired_hosts::retire_hosts(&mut profile.hosts, active);
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
            scopes: None,
            versioning: None,
            work_flow: None,
            remote_copy: None,
            test_paths: None,
            ci: None,
            checks: None,
        }
    }

    fn with_hosts(hosts: &[&str]) -> Profile {
        Profile {
            hosts: hosts.iter().map(|h| h.to_string()).collect(),
            ..v2()
        }
    }

    /// SPEC-20261118-profile-host-omp AC-1: the three known hosts validate, alone or together.
    #[test]
    fn three_hosts_with_omp_validate() {
        for hosts in [
            &["omp"][..],
            &["claude", "omp"],
            &["claude", "codex"],
            &["claude", "codex", "omp"],
        ] {
            assert!(
                validate(&with_hosts(hosts)).is_ok(),
                "{hosts:?} was refused"
            );
        }
    }

    /// SPEC-20261118-profile-host-omp AC-1: the boundary is three (N/N+1), and a host other
    /// than the three known ones is refused at any count.
    #[test]
    fn four_hosts_and_unknown_hosts_are_refused() {
        for hosts in [
            &["claude", "codex", "omp", "claude"][..],
            &["claude", "codex", "omp", "kiro"],
            &["claude", "codex", "zcode"],
            &["zcode"],
            &["OMP"],
            &[],
        ] {
            assert!(validate(&with_hosts(hosts)).is_err(), "{hosts:?} passed");
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
            scopes: Some(vec![crate::scopes::ScopeEntry {
                name: "web".into(),
                paths: vec!["apps/web".into()],
            }]),
            versioning: Some(crate::versioning::VersioningSettings {
                scheme: "headver".into(),
                format: None,
                head: Some(3),
            }),
            work_flow: Some(WorkFlowSettings {
                flow: "m-flow".into(),
                integration_branch: Some("main".into()),
            }),
            remote_copy: Some(true),
            test_paths: Some(vec!["qa/".into()]),
            ci: Some(crate::project_ci::CiAnswers {
                kind: "cli".into(),
                systems: vec!["linux".into()],
                deploy: crate::project_ci::CiDeploy {
                    enabled: false,
                    target: None,
                },
                visibility: "public".into(),
                monthly_budget_usd: None,
                macos_price_ack: None,
                dependabot: None,
                strict: "report".into(),
            }),
            checks: Some(vec![crate::project_ci::CheckEntry {
                id: "test".into(),
                command: "cargo test".into(),
            }]),
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
                "detected",
                "scopes",
                "versioning",
                "work_flow",
                "remote_copy",
                "test_paths",
                "ci",
                "checks"
            ]
        );
    }

    /// #3010: `scopes` is additive and validated by `validate` and by `read_profile` on a
    /// hand-edited file; an absent key plans exactly as before the key existed.
    #[test]
    fn scopes_key_is_additive_and_validated() {
        let base = serde_json::json!({
            "schema_version": 2, "project_id": "demo", "purpose": "p",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"]
        });
        let listed = |d: &Delta| {
            [&d.added, &d.changed, &d.removed, &d.unchanged, &d.conflict]
                .iter()
                .any(|l| l.contains(&"scopes".to_string()))
        };
        assert!(!listed(&compute_delta(
            &KNOWN_HOSTS,
            Some(&base),
            Some(&base),
            &base
        )));
        let mut desired = base.clone();
        desired["scopes"] = serde_json::json!([{"name": "web", "paths": ["apps/web"]}]);
        assert_eq!(
            compute_delta(&KNOWN_HOSTS, Some(&base), Some(&base), &desired).added,
            vec!["scopes"]
        );
        let with = |name: &str, paths: &[&str]| Profile {
            scopes: Some(vec![crate::scopes::ScopeEntry {
                name: name.into(),
                paths: paths.iter().map(|p| p.to_string()).collect(),
            }]),
            ..v2()
        };
        assert!(validate(&with("web", &["apps/web", "packages/ui"])).is_ok());
        assert!(validate(&with("web", &["../up"])).is_err());
        assert!(validate(&with("Web", &["apps/web"])).is_err());
        assert!(validate(&with("web", &[])).is_err());
        assert!(validate(&Profile {
            scopes: Some(Vec::new()),
            ..v2()
        })
        .is_err());

        let tmp = tempfile::tempdir().unwrap();
        let path = tmp
            .path()
            .join(crate::PROJECT_STATE_DIR)
            .join("project.yaml");
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        for (bad, why) in [
            (json!([{"name": "web", "paths": ["/abs"]}]), "absolute path"),
            (
                json!([{"name": "web", "paths": ["a"], "x": 1}]),
                "unknown field",
            ),
            (json!([{"name": "web"}]), "missing paths"),
            (json!({"web": ["a"]}), "wrong shape"),
        ] {
            let mut file = serde_json::to_value(v2()).unwrap();
            file["scopes"] = bad;
            fs::write(&path, serde_json::to_vec(&file).unwrap()).unwrap();
            assert!(read_profile(tmp.path()).is_err(), "{why}");
        }
        let mut file = serde_json::to_value(v2()).unwrap();
        file["scopes"] = json!([{"name": "web", "paths": ["apps/web"]}]);
        fs::write(&path, serde_json::to_vec(&file).unwrap()).unwrap();
        let (profile, _) = read_profile(tmp.path()).unwrap();
        assert_eq!(profile.scopes.unwrap()[0].paths, ["apps/web"]);
    }

    /// C-69: `versioning` is additive and validated by `read_profile` on a hand-edited file; an
    /// absent key plans exactly as before the key existed and an older profile still loads.
    #[test]
    fn versioning_key_is_additive_and_validated() {
        let base = serde_json::json!({
            "schema_version": 2, "project_id": "demo", "purpose": "p",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"]
        });
        let listed = |d: &Delta| {
            [&d.added, &d.changed, &d.removed, &d.unchanged, &d.conflict]
                .iter()
                .any(|l| l.contains(&"versioning".to_string()))
        };
        assert!(!listed(&compute_delta(
            &KNOWN_HOSTS,
            Some(&base),
            Some(&base),
            &base
        )));
        let mut desired = base.clone();
        desired["versioning"] = serde_json::json!({"scheme": "calver", "format": "YYYY.0M.MICRO"});
        assert_eq!(
            compute_delta(&KNOWN_HOSTS, Some(&base), Some(&base), &desired).added,
            vec!["versioning"]
        );

        let tmp = tempfile::tempdir().unwrap();
        let path = tmp
            .path()
            .join(crate::PROJECT_STATE_DIR)
            .join("project.yaml");
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        for (bad, why) in [
            (json!({"scheme": "romantic"}), "unknown scheme"),
            (json!({"scheme": "SemVer"}), "scheme names are lower-case"),
            (
                json!({"scheme": "semver", "head": 1}),
                "head without HeadVer",
            ),
            (
                json!({"scheme": "headver", "format": "YYYY.0M"}),
                "format without CalVer or custom",
            ),
            (
                json!({"scheme": "calver", "format": "YYYY"}),
                "one-token format",
            ),
            (json!({"scheme": "semver", "x": 1}), "unknown field"),
            (json!({"format": "YYYY.0M"}), "no scheme"),
            (json!("semver"), "wrong shape"),
            (json!(["semver"]), "an array, not an object"),
            (
                json!({"scheme": "headver", "head": 10000}),
                "head over 9999",
            ),
            (
                json!({"scheme": "custom", "format": "run `id` first"}),
                "a backtick in the custom text",
            ),
            (
                json!({"scheme": "calver", "format": "YYYY.0M.0M"}),
                "a CalVer token twice",
            ),
            (json!({"scheme": "headver", "head": -1}), "negative head"),
            (json!({"scheme": "headver", "head": "3"}), "head as text"),
            (
                json!({"scheme": "custom", "format": "see docs/VERSIONS.md"}),
                "a path in the custom text",
            ),
            // C-71 audit N2: `\` is refused like `/`, as the refusal lists say.
            (
                json!({"scheme": "custom", "format": "see docs\\VERSIONS.md"}),
                "a backslash path in the custom text",
            ),
            (
                json!({"scheme": "custom", "format": "a \\ b"}),
                "a lone backslash in the custom text",
            ),
        ] {
            let mut file = serde_json::to_value(v2()).unwrap();
            file["versioning"] = bad;
            fs::write(&path, serde_json::to_vec(&file).unwrap()).unwrap();
            assert!(read_profile(tmp.path()).is_err(), "{why}");
        }
        for good in [
            json!({"scheme": "semver"}),
            json!({"scheme": "headver", "head": 3}),
            json!({"scheme": "headver", "head": 0}),
            json!({"scheme": "calver", "format": "YYYY.0M.MICRO"}),
            json!({"scheme": "custom", "format": "release train letter plus build number"}),
        ] {
            let mut file = serde_json::to_value(v2()).unwrap();
            file["versioning"] = good.clone();
            fs::write(&path, serde_json::to_vec(&file).unwrap()).unwrap();
            let (profile, _) = read_profile(tmp.path()).unwrap();
            assert_eq!(serde_json::to_value(&profile).unwrap()["versioning"], good);
        }
        // A profile written before the key existed loads and does not gain it.
        fs::write(&path, serde_json::to_vec(&v2()).unwrap()).unwrap();
        let (profile, _) = read_profile(tmp.path()).unwrap();
        assert!(serde_json::to_value(&profile)
            .unwrap()
            .get("versioning")
            .is_none());
    }

    /// Card C-70 AC-2: `work_flow` is additive and validated by `read_profile` on a hand-edited
    /// file: one of the four flows, and an integration branch name git takes; an older profile
    /// still loads without it.
    #[test]
    fn work_flow_key_is_additive_and_validated() {
        let base = serde_json::json!({
            "schema_version": 2, "project_id": "demo", "purpose": "p",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"]
        });
        let mut desired = base.clone();
        desired["work_flow"] = json!({"flow": "m-flow", "integration_branch": "main"});
        assert_eq!(
            compute_delta(&KNOWN_HOSTS, Some(&base), Some(&base), &desired).added,
            vec!["work_flow"]
        );
        let tmp = tempfile::tempdir().unwrap();
        let path = tmp
            .path()
            .join(crate::PROJECT_STATE_DIR)
            .join("project.yaml");
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        for (bad, why) in [
            (json!({"flow": "gitlab-flow"}), "unknown flow"),
            (json!({"flow": "M-Flow"}), "flow names are lower-case"),
            (json!({"integration_branch": "main"}), "no flow"),
            (json!({"flow": "m-flow", "x": 1}), "unknown field"),
            (json!("m-flow"), "wrong shape"),
            (
                json!({"flow": "m-flow", "integration_branch": ""}),
                "empty branch",
            ),
            (
                json!({"flow": "m-flow", "integration_branch": "a..b"}),
                "`..`",
            ),
            (
                json!({"flow": "m-flow", "integration_branch": "-x"}),
                "leading dash",
            ),
            (
                json!({"flow": "m-flow", "integration_branch": "a b"}),
                "a space",
            ),
            (
                json!({"flow": "m-flow", "integration_branch": "x.lock"}),
                "`.lock`",
            ),
            (
                json!({"flow": "m-flow", "integration_branch": "a/"}),
                "trailing slash",
            ),
            (
                json!({"flow": "m-flow", "integration_branch": "b".repeat(101)}),
                "101 bytes",
            ),
            // Audit B2: `git check-ref-format --branch` refuses `.lock` at the end of any
            // `/` component, and `HEAD` as a branch name.
            (
                json!({"flow": "m-flow", "integration_branch": "x.lock/y"}),
                "`.lock` ending the first component",
            ),
            (
                json!({"flow": "m-flow", "integration_branch": "a/b.lock/c"}),
                "`.lock` ending a middle component",
            ),
            (
                json!({"flow": "m-flow", "integration_branch": "HEAD"}),
                "HEAD",
            ),
        ] {
            let mut file = serde_json::to_value(v2()).unwrap();
            file["work_flow"] = bad;
            fs::write(&path, serde_json::to_vec(&file).unwrap()).unwrap();
            assert!(read_profile(tmp.path()).is_err(), "{why}");
        }
        for good in [
            json!({"flow": "m-flow", "integration_branch": "v2026.11.0"}),
            json!({"flow": "m-flow"}),
            json!({"flow": "github-flow", "integration_branch": "main"}),
            json!({"flow": "trunk-based", "integration_branch": "trunk"}),
            json!({"flow": "git-flow", "integration_branch": "release/next"}),
            json!({"flow": "m-flow", "integration_branch": "b".repeat(100)}),
            // The N+1 side of the audit B2 edges: one letter off is a branch git takes.
            json!({"flow": "m-flow", "integration_branch": "x.loc/y"}),
            json!({"flow": "m-flow", "integration_branch": "a/b.locks/c"}),
            json!({"flow": "m-flow", "integration_branch": "HEADS"}),
            json!({"flow": "m-flow", "integration_branch": "release/HEAD"}),
        ] {
            let mut file = serde_json::to_value(v2()).unwrap();
            file["work_flow"] = good.clone();
            fs::write(&path, serde_json::to_vec(&file).unwrap()).unwrap();
            let (profile, _) = read_profile(tmp.path()).unwrap();
            assert_eq!(serde_json::to_value(&profile).unwrap()["work_flow"], good);
        }
        fs::write(&path, serde_json::to_vec(&v2()).unwrap()).unwrap();
        let (profile, _) = read_profile(tmp.path()).unwrap();
        assert!(serde_json::to_value(&profile)
            .unwrap()
            .get("work_flow")
            .is_none());
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
        assert!(!all(&compute_delta(
            &KNOWN_HOSTS,
            Some(&base),
            Some(&base),
            &base
        )));
        let mut desired = base.clone();
        desired["language"] = serde_json::json!({"code_comments": "en"});
        assert_eq!(
            compute_delta(&KNOWN_HOSTS, Some(&base), Some(&base), &desired).added,
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

    /// Card C-90: custom text with `sk-` inside a word is portable; a key shape is not.
    #[test]
    fn sk_inside_a_word_is_portable_text() {
        for ok in [
            "task-based releases",
            "risk-free, desk-top first",
            "YYYY.task-N",
        ] {
            assert!(portable(ok), "{ok}");
        }
        for no in [
            "sk-abc123",
            "uses sk-test-not-real",
            "x token=abc",
            "a ghp_x",
        ] {
            assert!(!portable(no), "{no}");
        }
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
        let delta = compute_delta(&KNOWN_HOSTS, Some(&baseline), Some(&current), &desired);
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
        assert!(!all(&compute_delta(
            &KNOWN_HOSTS,
            Some(&base),
            Some(&base),
            &base
        )));
        let mut desired = base.clone();
        desired["leader"] = serde_json::json!({"pr_check_interval_secs": 60});
        let delta = compute_delta(&KNOWN_HOSTS, Some(&base), Some(&base), &desired);
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

    // ---- SPEC-20261121-mixed-profile-read: which profile file a reader opens. ----------------

    const EARLIER: &str = crate::state_root::EARLIER_PROJECT_DIR;
    const MIXED_PLAIN: &str = r#"{"schema_version": 1, "project_id": "demo", "purpose": "fixture", "tools": ["rust"], "development_methods": ["sdd"], "hosts": ["claude"]}"#;
    const MIXED_MFLOW: &str = r#"{"schema_version": 1, "project_id": "demo", "purpose": "fixture", "tools": ["rust"], "development_methods": ["sdd"], "hosts": ["claude"], "work_flow": {"flow": "m-flow", "integration_branch": "integ"}}"#;
    const MIXED_BROKEN: &str = r#"{"schema_version": 1, "project_id": "demo", "purpose": "fixture", "tools": ["rust"], "development_methods": ["sdd"], "hosts": ["claude"], "work_flow": {"flow": "bogus-flow"}}"#;

    fn mixed_put(root: &Path, dir: &str, bytes: &str) {
        fs::create_dir_all(root.join(dir)).unwrap();
        fs::write(root.join(dir).join("project.yaml"), bytes).unwrap();
    }

    fn mixed_flow(root: &Path) -> Option<String> {
        read_profile(root).unwrap().0.work_flow.map(|w| w.flow)
    }

    #[test]
    fn mixed_layout_reads_the_newer_profile_when_both_exist() {
        let t = tempfile::tempdir().unwrap();
        mixed_put(t.path(), EARLIER, MIXED_PLAIN);
        mixed_put(t.path(), ".moai", MIXED_MFLOW);
        assert_eq!(mixed_flow(t.path()).as_deref(), Some("m-flow"));
        assert_eq!(read_rel(t.path()), ".moai/project.yaml");
        assert_eq!(
            existing_project_id(t.path()).unwrap().as_deref(),
            Some("demo")
        );
    }

    #[test]
    fn mixed_layout_keeps_the_layouts_own_folder_when_only_one_profile_exists() {
        let earlier = tempfile::tempdir().unwrap();
        mixed_put(earlier.path(), EARLIER, MIXED_MFLOW);
        assert_eq!(mixed_flow(earlier.path()).as_deref(), Some("m-flow"));
        assert_eq!(read_rel(earlier.path()), format!("{EARLIER}/project.yaml"));

        // Only the newer file and no marker: the layout is not `rust`, the earlier folder has no
        // profile, so the reader still looks where it always did and finds none.
        let newer = tempfile::tempdir().unwrap();
        mixed_put(newer.path(), ".moai", MIXED_MFLOW);
        assert_eq!(
            read_profile(newer.path()).unwrap_err(),
            PROFILE_MISSING.to_string()
        );
        assert_eq!(read_rel(newer.path()), format!("{EARLIER}/project.yaml"));

        // A `rust` layout (valid marker) reads `.moai`, with or without the earlier file.
        for with_earlier in [false, true] {
            let t = tempfile::tempdir().unwrap();
            mixed_put(t.path(), ".moai", MIXED_MFLOW);
            fs::write(
                t.path().join(".moai").join(crate::state_root::RUST_MARKER),
                crate::state_root::marker_bytes("test", &[]),
            )
            .unwrap();
            if with_earlier {
                mixed_put(t.path(), EARLIER, MIXED_PLAIN);
            }
            assert_eq!(mixed_flow(t.path()).as_deref(), Some("m-flow"));
            assert_eq!(read_rel(t.path()), ".moai/project.yaml");
        }
    }

    #[test]
    fn mixed_layout_newer_profile_that_cannot_be_read_is_no_reason_to_fall_back() {
        let t = tempfile::tempdir().unwrap();
        mixed_put(t.path(), EARLIER, MIXED_MFLOW);
        mixed_put(t.path(), ".moai", MIXED_BROKEN);
        assert!(read_profile(t.path()).is_err());
        assert_eq!(read_rel(t.path()), ".moai/project.yaml");
    }

    #[cfg(unix)]
    #[test]
    fn mixed_layout_a_linked_newer_profile_is_not_chosen() {
        let t = tempfile::tempdir().unwrap();
        mixed_put(t.path(), EARLIER, MIXED_PLAIN);
        fs::create_dir_all(t.path().join(".moai")).unwrap();
        let real = t.path().join("elsewhere.yaml");
        fs::write(&real, MIXED_MFLOW).unwrap();
        std::os::unix::fs::symlink(&real, t.path().join(".moai").join("project.yaml")).unwrap();
        assert_eq!(read_rel(t.path()), format!("{EARLIER}/project.yaml"));
        assert_eq!(mixed_flow(t.path()), None);
    }
}
