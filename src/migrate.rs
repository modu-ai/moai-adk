//! #2016 isolated migration rehearsal. No live HOME discovery, launcher, key,
//! PATH, host configuration, binary replacement or Go removal is implemented.
//! The leader owns CLI registration. Filesystem execution is deliberately confined
//! to an explicitly supplied temporary synthetic home; production admission remains
//! a separate integration gate. Callers must serialize access to that fixture.

use crate::apply;
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};
use std::{
    collections::BTreeSet,
    fs,
    io::Read,
    path::{Component, Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;
const LIMIT: u64 = 262_144;

pub fn hash(bytes: &[u8]) -> String {
    format!("{:x}", Sha256::digest(bytes))
}
fn digest<T: Serialize>(value: &T) -> Result<String> {
    Ok(hash(
        &serde_json::to_vec(value).map_err(|_| "encoding failed")?,
    ))
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
pub enum Host {
    Claude,
    Codex,
    Glm,
    OnPrem,
}
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
pub enum Mode {
    Default,
    Automatic,
}
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Selection {
    pub source: String,
    pub target: String,
    pub recorded_sha256: Option<String>,
    pub project_current: Option<String>,
}
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Entry {
    pub selection: Selection,
    pub source_sha256: String,
    pub target_sha256: Option<String>,
    pub project_sha256: Option<String>,
    pub ownership: String,
    pub action: String,
}
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Plan {
    pub home: PathBuf,
    pub hosts: Vec<Host>,
    pub mode: Mode,
    pub summary: String,
    pub entries: Vec<Entry>,
    pub digest: String,
}
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Receipt {
    pub plan: Plan,
    pub published: Vec<bool>,
}
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct RecoveryPlan {
    pub receipt: Receipt,
    pub actions: Vec<String>,
    pub digest: String,
}

pub struct SyntheticHome {
    root: PathBuf,
}
impl SyntheticHome {
    pub fn new(root: &Path) -> Result<Self> {
        let root = fs::canonicalize(root).map_err(|_| "synthetic root unavailable")?;
        let temp = fs::canonicalize(std::env::temp_dir()).map_err(|_| "temp unavailable")?;
        if root == temp || !root.starts_with(&temp) || !root.is_dir() {
            return Err("only an explicit temporary synthetic home is allowed".into());
        }
        Ok(Self { root })
    }

    fn path(&self, rel: &str) -> Result<PathBuf> {
        if fs::symlink_metadata(&self.root)
            .map_err(|_| "root unavailable")?
            .file_type()
            .is_symlink()
        {
            return Err("synthetic root became a link".into());
        }
        let p = Path::new(rel);
        if rel.is_empty()
            || rel.contains('\\')
            || !p.components().all(|c| matches!(c, Component::Normal(_)))
        {
            return Err("invalid relative path".into());
        }
        let mut cursor = self.root.clone();
        for part in p.components() {
            cursor.push(part);
            match fs::symlink_metadata(&cursor) {
                Ok(m) if m.file_type().is_symlink() || (!m.is_file() && !m.is_dir()) => {
                    return Err("link or special file refused".into())
                }
                Ok(_) => (),
                Err(e) if e.kind() == std::io::ErrorKind::NotFound => (),
                Err(_) => return Err("path inspection failed".into()),
            }
        }
        Ok(cursor)
    }
    fn read(&self, rel: &str) -> Result<Option<Vec<u8>>> {
        let p = self.path(rel)?;
        let m = match fs::symlink_metadata(&p) {
            Ok(m) => m,
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(None),
            Err(_) => return Err("metadata failed".into()),
        };
        if !m.is_file() || m.len() > LIMIT {
            return Err("bounded regular file required".into());
        }
        let mut bytes = Vec::new();
        fs::File::open(p)
            .map_err(|_| "read failed")?
            .take(LIMIT + 1)
            .read_to_end(&mut bytes)
            .map_err(|_| "read failed")?;
        if bytes.len() as u64 > LIMIT {
            return Err("file grew beyond limit".into());
        }
        Ok(Some(bytes))
    }
    fn source(&self, rel: &str) -> Result<Vec<u8>> {
        let lower = rel.to_ascii_lowercase();
        if !(rel.starts_with(".moai/config/") || rel.starts_with(".moai/specs/"))
            || [
                ".env",
                "secret",
                "credential",
                "token",
                "key",
                "worktrees",
                "db",
            ]
            .iter()
            .any(|s| lower.contains(s))
        {
            return Err("only selected non-secret Go config and SPEC files are allowed".into());
        }
        let bytes = self.read(rel)?.ok_or("source missing")?;
        let text = std::str::from_utf8(&bytes)
            .map_err(|_| "text source required")?
            .to_ascii_lowercase();
        if [
            "api_key",
            "api-key",
            "apikey",
            "password",
            "secret",
            "bearer ",
            "access_token",
            "auth_token",
            "private key",
        ]
        .iter()
        .any(|s| text.contains(s))
        {
            return Err("potential credential content refused; no values logged".into());
        }
        Ok(bytes)
    }
    pub fn plan(&self, hosts: &[Host], mode: Mode, selections: &[Selection]) -> Result<Plan> {
        if hosts.is_empty() || selections.is_empty() || selections.len() > 128 {
            return Err("bounded explicit selection required".into());
        }
        let mut seen = BTreeSet::new();
        let mut entries = Vec::new();
        for s in selections {
            if !s.target.starts_with(".moai-cli/imports/") || !seen.insert(s.target.clone()) {
                return Err("distinct import targets required".into());
            }
            let src = self.source(&s.source)?;
            let source_sha256 = hash(&src);
            let target_sha256 = self.read(&s.target)?.map(|b| hash(&b));
            let project_sha256 = match &s.project_current {
                None => None,
                Some(path)
                    if path.starts_with("project/.moai/config/")
                        || path.starts_with("project/.moai/specs/") =>
                {
                    self.read(path)?.map(|b| hash(&b))
                }
                Some(_) => return Err("project comparison must name config or SPEC".into()),
            };
            let ownership = match &s.recorded_sha256 {
                None => "unknown",
                Some(h) if h == &source_sha256 => "owned",
                Some(_) => "modified",
            };
            let action = if ownership != "owned" {
                "preserved"
            } else if project_sha256.as_ref().is_some_and(|h| h != &source_sha256) {
                "conflict"
            } else if target_sha256.as_ref() == Some(&source_sha256) {
                "unchanged"
            } else if target_sha256.is_some() {
                "conflict"
            } else {
                "create"
            };
            entries.push(Entry {
                selection: s.clone(),
                source_sha256,
                target_sha256,
                project_sha256,
                ownership: ownership.into(),
                action: action.into(),
            });
        }
        let summary = format!("Hosts: {hosts:?}; mode: {mode:?}. Default requests confirmation; automatic mode can reduce permission prompts and isolation in a host adapter. This rehearsal does not change security modes, keys, PATH or host settings. GLM/OnPrem adapter support is unverified. Global .moai is the source baseline; existing project imports are conflicts. Final summary confirmation must bind this entire plan.");
        let mut plan = Plan {
            home: self.root.clone(),
            hosts: hosts.to_vec(),
            mode,
            summary,
            entries,
            digest: String::new(),
        };
        plan.digest = digest(&plan)?;
        Ok(plan)
    }
    fn backup_rel(plan: &Plan, suffix: &str) -> String {
        format!(".moai-go-backup/{}/{suffix}", plan.digest)
    }
    fn save_receipt(&self, receipt: &Receipt) -> Result<()> {
        let rel = Self::backup_rel(&receipt.plan, "manifest.json");
        let p = self.path(&rel)?;
        let old = self.read(&rel)?;
        let bytes = serde_json::to_vec(receipt).map_err(|_| "receipt encoding failed")?;
        apply::publish_replace(&p, old.as_deref(), &bytes)
    }
    pub fn apply(&self, plan: &Plan, confirmed: &str) -> Result<Receipt> {
        let selections: Vec<_> = plan.entries.iter().map(|e| e.selection.clone()).collect();
        let fresh = self.plan(&plan.hosts, plan.mode.clone(), &selections)?;
        if plan != &fresh || confirmed != fresh.digest {
            return Err("stale, altered or unconfirmed summary".into());
        }
        if plan
            .entries
            .iter()
            .any(|e| e.action == "conflict" || e.ownership != "owned")
        {
            return Err("conflict or preserved source; reselect before apply".into());
        }
        let root = self
            .path(&Self::backup_rel(plan, "backup"))?
            .parent()
            .ok_or("backup path")?
            .to_path_buf();
        fs::create_dir_all(root.parent().ok_or("backup parent")?)
            .map_err(|_| "backup unavailable")?;
        fs::create_dir(&root).map_err(|_| "backup already exists or unavailable")?;
        for (i, e) in plan.entries.iter().enumerate() {
            let bytes = self.source(&e.selection.source)?;
            if hash(&bytes) != e.source_sha256 {
                return Err("source changed before backup".into());
            }
            let rel = Self::backup_rel(plan, &i.to_string());
            apply::publish_exclusive(&self.path(&rel)?, &bytes)?;
            if self.read(&rel)?.as_deref().map(hash).as_ref() != Some(&e.source_sha256) {
                return Err("backup hash mismatch".into());
            }
        }
        let mut receipt = Receipt {
            plan: plan.clone(),
            published: vec![false; plan.entries.len()],
        };
        self.save_receipt(&receipt)?;
        for (i, e) in plan.entries.iter().enumerate() {
            if e.action != "create" {
                continue;
            }
            let bytes = self.source(&e.selection.source)?;
            if hash(&bytes) != e.source_sha256 {
                return Err("source changed; backups retained".into());
            }
            let target = self.path(&e.selection.target)?;
            fs::create_dir_all(target.parent().ok_or("target parent")?)
                .map_err(|_| "target unavailable")?;
            apply::publish_exclusive(&target, &bytes)?;
            receipt.published[i] = true;
            self.save_receipt(&receipt)?;
        }
        Ok(receipt)
    }
    pub fn recovery_plan(&self, receipt: &Receipt) -> Result<RecoveryPlan> {
        if receipt.plan.home != self.root || receipt.published.len() != receipt.plan.entries.len() {
            return Err("receipt scope mismatch".into());
        }
        let saved = self
            .read(&Self::backup_rel(&receipt.plan, "manifest.json"))?
            .ok_or("manifest missing")?;
        let on_disk: Receipt = serde_json::from_slice(&saved).map_err(|_| "invalid manifest")?;
        if &on_disk != receipt {
            return Err("receipt differs from durable manifest".into());
        }
        let mut actions = Vec::new();
        for (i, e) in receipt.plan.entries.iter().enumerate() {
            if !e.selection.target.starts_with(".moai-cli/imports/") {
                return Err("invalid recovery target".into());
            }
            let backup = self
                .read(&Self::backup_rel(&receipt.plan, &i.to_string()))?
                .ok_or("backup missing")?;
            if hash(&backup) != e.source_sha256 {
                return Err("backup hash mismatch".into());
            }
            let current = self.read(&e.selection.target)?.map(|b| hash(&b));
            actions.push(
                if !receipt.published[i] {
                    "preserved"
                } else if current.is_none() {
                    "unchanged"
                } else if current.as_ref() == Some(&e.source_sha256) {
                    "removed_import"
                } else {
                    "conflict"
                }
                .into(),
            );
        }
        let mut p = RecoveryPlan {
            receipt: receipt.clone(),
            actions,
            digest: String::new(),
        };
        p.digest = digest(&p)?;
        Ok(p)
    }
    pub fn recover(&self, plan: &RecoveryPlan, confirmed: &str) -> Result<Vec<String>> {
        let fresh = self.recovery_plan(&plan.receipt)?;
        if plan != &fresh || confirmed != fresh.digest {
            return Err("stale or unconfirmed recovery".into());
        }
        for (e, a) in fresh.receipt.plan.entries.iter().zip(&fresh.actions) {
            if a == "removed_import" {
                let current = self
                    .read(&e.selection.target)?
                    .ok_or("recovery target changed")?;
                if hash(&current) != e.source_sha256 {
                    return Err("recovery target changed".into());
                }
                fs::remove_file(self.path(&e.selection.target)?)
                    .map_err(|_| "import recovery failed")?;
            }
        }
        Ok(fresh.actions)
    }
}
