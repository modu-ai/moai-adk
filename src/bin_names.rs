//! The two command names in an installed version folder (#2791 AC-1, design #2725 R2).
//!
//! The decided layout installs a build at `<home>/<HOME_STATE_DIR>/versions/<v>/bin/moai-cli`
//! (`moai-cli.exe` on Windows). `moai` is the public name (#2790 R1 builds both from one
//! `main`); here the folder gets `bin/moai` as a hard link to `bin/moai-cli`, so the two names
//! are one file and the installation keeps the one identity `install::identity_exe` gives it.
//!
//! Plan → apply with a plan digest and a caller-supplied permission reference. Only the
//! version folder this binary runs from is touched; PATH, shell files and any other `moai`
//! (the Go `moai` included) are never written — `install-inspect` reports those. Nothing is
//! deleted: a separate `moai` copy of the same build is moved aside to
//! `moai.replaced-<UTC>` before the link is made.

use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

/// The file names in a version folder's `bin/` (with `.exe` on Windows).
fn exe_names() -> (String, String) {
    let ext = std::env::consts::EXE_SUFFIX;
    (
        format!("{}{ext}", crate::install::PRODUCT),
        format!("moai{ext}"),
    )
}

/// What the plan found; the apply acts only on `create` and `replace_copy`.
#[derive(Debug, Clone)]
pub struct Plan {
    /// `outside_layout`, `create`, `unchanged`, `replace_copy` or `blocked`.
    pub state: &'static str,
    /// Why nothing will be written, for `outside_layout` and `blocked`.
    pub reason: Option<String>,
    /// `bin/moai-cli` of this build's version folder.
    pub canonical: Option<PathBuf>,
    /// `bin/moai` beside it.
    pub alias: Option<PathBuf>,
    /// The sha256 of `bin/moai-cli`; the apply refuses when it changed.
    pub canonical_sha256: Option<String>,
    /// The sha256 of an existing `bin/moai`, if any.
    pub alias_sha256: Option<String>,
    pub plan_digest: String,
}

impl Plan {
    pub fn writes(&self) -> bool {
        matches!(self.state, "create" | "replace_copy")
    }

    pub fn to_json(&self) -> Value {
        let text = |p: &Option<PathBuf>| p.as_ref().map(|p| p.to_string_lossy().into_owned());
        json!({
            "state": self.state,
            "reason": self.reason,
            "canonical": text(&self.canonical),
            "alias": text(&self.alias),
            "canonical_sha256": self.canonical_sha256,
            "alias_sha256": self.alias_sha256,
            "plan_digest": self.plan_digest,
        })
    }
}

fn finish(
    state: &'static str,
    reason: Option<String>,
    canonical: Option<PathBuf>,
    alias: Option<PathBuf>,
    canonical_sha256: Option<String>,
    alias_sha256: Option<String>,
) -> Result<Plan> {
    let mut plan = Plan {
        state,
        reason,
        canonical,
        alias,
        canonical_sha256,
        alias_sha256,
        plan_digest: String::new(),
    };
    let mut body = plan.to_json();
    body.as_object_mut().expect("object").remove("plan_digest");
    plan.plan_digest = crate::digest(&serde_json::to_vec(&body).map_err(|_| "encode failed")?);
    Ok(plan)
}

/// The version folder's `bin/` this executable runs from, when it sits exactly in the
/// layout `<home>/<HOME_STATE_DIR>/versions/<v>/bin/` under either name.
fn layout_bin(home: &Path, exe: &Path) -> Option<PathBuf> {
    let canonical_home = fs::canonicalize(home).ok()?;
    let canonical_exe = fs::canonicalize(exe).ok()?;
    let bin = canonical_exe.parent()?;
    let versions = crate::home_root::path(&canonical_home).join("versions");
    (bin.file_name()? == "bin" && bin.parent()?.parent()? == versions).then(|| bin.to_path_buf())
}

/// Preview the two names for the executable `exe` (the running binary in the CLI). Reads
/// only; an existing `bin/moai` that is not the same bytes is run once with `--version`.
pub fn plan_for(home: &Path, exe: &Path) -> Result<Plan> {
    let Some(bin) = layout_bin(home, exe) else {
        return finish(
            "outside_layout",
            Some(format!(
                "this binary is not in the version folder layout ~/{}/versions/<version>/bin/; nothing is linked",
                crate::home_root::dir(home)
            )),
            None,
            None,
            None,
            None,
        );
    };
    let (cli_name, alias_name) = exe_names();
    let canonical = bin.join(cli_name);
    let alias = bin.join(alias_name);
    let canonical_ok = fs::symlink_metadata(&canonical).is_ok_and(|m| m.file_type().is_file());
    if !canonical_ok {
        return finish(
            "blocked",
            Some("bin/moai-cli is missing or not a regular file".into()),
            Some(canonical),
            Some(alias),
            None,
            None,
        );
    }
    let canonical_sha = crate::install::sha256_file(&canonical)?;
    let Ok(meta) = fs::symlink_metadata(&alias) else {
        return finish(
            "create",
            None,
            Some(canonical),
            Some(alias),
            Some(canonical_sha),
            None,
        );
    };
    if !meta.file_type().is_file() {
        return finish(
            "blocked",
            Some("bin/moai is a link or a folder; it is left as it is".into()),
            Some(canonical),
            Some(alias),
            Some(canonical_sha),
            None,
        );
    }
    let alias_sha = crate::install::sha256_file(&alias)?;
    if alias_sha == canonical_sha {
        return finish(
            "unchanged",
            None,
            Some(canonical),
            Some(alias),
            Some(canonical_sha),
            Some(alias_sha),
        );
    }
    // A separate `moai` of this same build (the release bundle carries both binaries) is
    // replaced by the link; anything else is not this tool's to touch.
    let line = crate::install::probe_version(&alias);
    let mine = format!("moai-cli {}", crate::VERSION_SIGNATURE);
    if line.as_deref() == Some(mine.as_str()) {
        finish(
            "replace_copy",
            None,
            Some(canonical),
            Some(alias),
            Some(canonical_sha),
            Some(alias_sha),
        )
    } else {
        finish(
            "blocked",
            Some(format!(
                "bin/moai is another program (--version: {}); it is left as it is",
                line.as_deref().unwrap_or("no output")
            )),
            Some(canonical),
            Some(alias),
            Some(canonical_sha),
            Some(alias_sha),
        )
    }
}

/// Make the link the plan for `exe` previewed. `now` names the aside copy (compact UTC).
pub fn apply_for(
    home: &Path,
    exe: &Path,
    plan_digest: &str,
    permission: &str,
    now: &str,
) -> Result<Value> {
    if permission.trim().is_empty() {
        return Err("a permission reference is required; nothing was written".into());
    }
    let plan = plan_for(home, exe)?;
    if plan.plan_digest != plan_digest {
        return Err("stale or altered plan; re-preview required, nothing was written".into());
    }
    let (Some(canonical), Some(alias)) = (plan.canonical.clone(), plan.alias.clone()) else {
        return Ok(json!({"state": plan.state, "plan_digest": plan.plan_digest, "written": []}));
    };
    let mut aside = None;
    match plan.state {
        "create" => {}
        "replace_copy" => {
            let stamp: String = now.chars().filter(|c| c.is_ascii_alphanumeric()).collect();
            let target = alias.with_file_name(format!(
                "moai.replaced-{stamp}{}",
                std::env::consts::EXE_SUFFIX
            ));
            if fs::symlink_metadata(&target).is_ok() {
                return Err(format!(
                    "{} already exists; nothing was written",
                    target.display()
                ));
            }
            fs::rename(&alias, &target).map_err(|e| {
                format!("could not move the separate bin/moai aside ({e}); nothing was written")
            })?;
            aside = Some(target);
        }
        _ => {
            return Ok(json!({
                "state": plan.state,
                "reason": plan.reason,
                "plan_digest": plan.plan_digest,
                "written": [],
            }))
        }
    }
    if let Err(e) = fs::hard_link(&canonical, &alias) {
        if let Some(a) = &aside {
            let back = fs::rename(a, &alias);
            return Err(format!(
                "could not link bin/moai ({e}); the separate copy is {}",
                if back.is_ok() {
                    "back in place".to_string()
                } else {
                    format!("at {}", a.display())
                }
            ));
        }
        return Err(format!(
            "could not link bin/moai ({e}); nothing was written"
        ));
    }
    let linked = crate::install::sha256_file(&alias)?;
    if Some(&linked) != plan.canonical_sha256.as_ref() {
        return Err("bin/moai does not hold bin/moai-cli's bytes after linking".into());
    }
    Ok(json!({
        "state": "applied",
        "plan_digest": plan.plan_digest,
        "linked": alias.to_string_lossy(),
        "to": canonical.to_string_lossy(),
        "moved_aside": aside.map(|a| a.to_string_lossy().into_owned()),
        "written": [alias.to_string_lossy()],
    }))
}
