//! SPEC-20261125-go-residue-apply: the apply and the revert of the go-residue step. The apply
//! moves the files of the chosen bundles (the proven `go_same` set and each explicitly picked
//! `go_changed` path) into `.moai/backups/<UTC>/go-residue/<same path>` and deletes nothing;
//! the revert puts them back from the receipt. Both go through one path guard
//! ([`crate::go_residue::contained`]) before they read, write or move anything, the receipt
//! is saved (pending) and read back before the first file moves, and the revert trusts the
//! receipt for nothing: every path in it passes the guard again and every restore is judged
//! by what is on disk (an empty original place and a backup whose sha256 equals the entry),
//! not by the receipt's `moved` flags.

use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{fs, path::Path};

use crate::go_residue::{contained, secret_path, Survey};
use crate::go_state::{
    make_dirs, make_private_dirs, move_exclusive, parent_of, read_bounded_text, sha256_file,
    still_as_previewed,
};
use crate::state_root::{utc_stamp, GO_STATE_RECEIPT_MAX};

type Result<T> = std::result::Result<T, String>;

/// The receipt's file name under `.moai/backups/<UTC>/`. It is not `go-state-receipt.json`, so
/// the state classifier and the go-state revert never read it.
const RECEIPT_FILE: &str = "go-residue-receipt.json";
/// The folder under the stamp folder that holds the moved files.
const AREA: &str = "go-residue";
/// One size limit for writing and reading the receipt.
const RECEIPT_MAX: u64 = GO_STATE_RECEIPT_MAX;
/// A receipt may name files that are private to the user.
const PRIVATE_FILE: u32 = 0o600;
/// How many following seconds a stamp may be moved on when its folder is taken already.
const STAMP_TRIES: i64 = 120;

/// The refusal for a digest that is not the fresh plan's (the phrase makes
/// `contract::classify` say `stale_plan`).
const STALE: &str =
    "stale or altered plan: what go-residue-apply would move no longer matches the \
     preview that was confirmed (a file changed, appeared or went, or the digest is from an \
     older version); nothing was written; run `moai project-update-plan <project>` again, read \
     the new go-residue step and confirm again with its plan_digest";

/// One file the apply may move.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Candidate {
    pub path: String,
    /// `go_same` or `go_changed`.
    pub class: String,
    pub sha256: String,
    pub tracked: bool,
}

/// The files the apply may move, from one survey: a `go_same` or `go_changed` item the survey
/// proved (it has a digest) and suggests to archive.
pub fn candidates(survey: &Survey) -> Vec<Candidate> {
    survey
        .items
        .iter()
        .filter(|i| {
            matches!(i.class, "go_same" | "go_changed")
                && i.suggested_action == "archive"
                && !secret_path(&i.path)
        })
        .filter_map(|i| {
            Some(Candidate {
                path: i.path.clone(),
                class: i.class.to_string(),
                sha256: i.sha256.clone()?,
                tracked: i.tracked,
            })
        })
        .collect()
}

/// The digest of the candidates and the manifest that proved them: what a confirmation binds.
pub fn plan_digest(proof: &str, candidates: &[Candidate]) -> String {
    let mut text = format!("proof\0{proof}\n");
    for c in candidates {
        text.push_str(&format!(
            "{}\0{}\0{}\0{}\n",
            c.path, c.class, c.sha256, c.tracked
        ));
    }
    format!("sha256:{:x}", Sha256::digest(text.as_bytes()))
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
struct Moved {
    path: String,
    sha256: String,
    #[serde(default)]
    class: String,
    #[serde(default)]
    tracked: bool,
    moved: bool,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
struct Receipt {
    schema_version: u32,
    project: String,
    backup: String,
    plan_digest: String,
    /// The caller's `--permission-reference`, recorded as given (not an authorization).
    permission_reference: String,
    /// The manifest that proved the files.
    proof: String,
    /// `pending` → `committed` → `reverted`.
    state: String,
    moves: Vec<Moved>,
}

fn receipt_rel(stamp: &str) -> String {
    format!(".moai/backups/{stamp}/{RECEIPT_FILE}")
}

fn area_rel(stamp: &str) -> String {
    format!(".moai/backups/{stamp}/{AREA}")
}

fn guard(root: &Path, rel: &str) -> Result<()> {
    if contained(root, rel) && !secret_path(rel) {
        Ok(())
    } else {
        Err(format!(
            "go-residue refused: {rel} is absolute, holds a `..` component, passes through a \
             link or names a secret-shaped file; nothing was written"
        ))
    }
}

fn bytes_of(receipt: &Receipt) -> Result<Vec<u8>> {
    let bytes = serde_json::to_vec_pretty(receipt).map_err(|_| "encode failed".to_string())?;
    // The writer keeps the limit its reader uses: a receipt that could not be read back is
    // never written.
    if bytes.len() as u64 > RECEIPT_MAX {
        return Err(format!(
            "go-residue refused: the receipt is larger than {} MiB; nothing was written",
            RECEIPT_MAX / (1024 * 1024)
        ));
    }
    Ok(bytes)
}

/// The first stamp, from now on, whose backup folder does not exist yet: two applies in one
/// second keep two receipts.
fn free_stamp(root: &Path) -> Result<String> {
    let now = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_secs() as i64)
        .unwrap_or(0);
    for i in 0..STAMP_TRIES {
        let stamp = crate::baseline::utc_from_epoch(now + i).replace(['-', ':'], "");
        if fs::symlink_metadata(root.join(format!(".moai/backups/{stamp}"))).is_err() {
            return Ok(stamp);
        }
    }
    Err("go-residue-apply refused: no free backup stamp in the next two minutes; nothing was written".into())
}

/// Apply the chosen bundles of the preview whose digest the caller confirmed. Nothing is
/// written unless a fresh preview has exactly that digest.
pub fn apply(
    project: &Path,
    plan_digest_given: &str,
    permission_reference: &str,
    bundle_proven: bool,
    picks: &[String],
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("a permission reference is required".into());
    }
    if !bundle_proven && picks.is_empty() {
        return Err(
            "go-residue-apply refused: choose `--bundle proven` or at least one `--pick <path>`; \
             nothing was written"
                .into(),
        );
    }
    let root = fs::canonicalize(project)
        .map_err(|_| crate::contract::PROJECT_PATH_UNAVAILABLE.to_string())?;
    let survey = crate::go_residue::survey(&root);
    let all = candidates(&survey);
    let proof = survey.proof.clone().unwrap_or_default();
    if plan_digest(&proof, &all) != plan_digest_given {
        return Err(STALE.into());
    }
    let mut chosen: Vec<Candidate> = Vec::new();
    if bundle_proven {
        chosen.extend(all.iter().filter(|c| c.class == "go_same").cloned());
    }
    for pick in picks {
        match all
            .iter()
            .find(|c| &c.path == pick && c.class == "go_changed")
        {
            Some(c) if !chosen.contains(c) => chosen.push(c.clone()),
            Some(_) => {}
            None => {
                return Err(format!(
                    "go-residue-apply refused: {pick} is not a changed Go file of the previewed \
                     plan (only a `go_changed` file is picked by path); nothing was written"
                ))
            }
        }
    }
    if chosen.is_empty() {
        return Err(
            "go-residue-apply refused: the chosen bundles hold no file; nothing was written".into(),
        );
    }
    chosen.sort_by(|a, b| a.path.cmp(&b.path));
    let stamp = free_stamp(&root)?;
    apply_chosen(
        &root,
        &stamp,
        &chosen,
        plan_digest_given,
        permission_reference,
        &proof,
    )
}

/// The apply of an already chosen list under `stamp`: the guard on every path, the pending
/// receipt saved and read back, then one move per file (each hashed once more right before),
/// the read-back of every moved file, and the receipt `committed`.
fn apply_chosen(
    root: &Path,
    stamp: &str,
    chosen: &[Candidate],
    digest: &str,
    permission: &str,
    proof: &str,
) -> Result<Value> {
    let area = area_rel(stamp);
    for c in chosen {
        guard(root, &c.path)?;
        guard(root, &format!("{area}/{}", c.path))?;
    }
    let mut receipt = Receipt {
        schema_version: 1,
        project: root.to_string_lossy().into_owned(),
        backup: stamp.to_string(),
        plan_digest: digest.to_string(),
        permission_reference: permission.to_string(),
        proof: proof.to_string(),
        state: "pending".into(),
        moves: chosen
            .iter()
            .map(|c| Moved {
                path: c.path.clone(),
                sha256: c.sha256.clone(),
                class: c.class.clone(),
                tracked: c.tracked,
                moved: false,
            })
            .collect(),
    };
    let pending = bytes_of(&receipt)?;
    let receipt_path = root.join(receipt_rel(stamp));
    make_private_dirs(root, &format!(".moai/backups/{stamp}")).map_err(|e| {
        format!("go-residue-apply refused: the backup area is not usable ({e}); nothing was moved")
    })?;
    crate::apply::publish_exclusive_mode(&receipt_path, &pending, Some(PRIVATE_FILE)).map_err(
        |e| format!("go-residue-apply refused: the pending receipt could not be saved ({e}); nothing was moved"),
    )?;
    if fs::read(&receipt_path).ok().as_deref() != Some(pending.as_slice()) {
        return Err(
            "go-residue-apply refused: the pending receipt does not read back; nothing was moved"
                .into(),
        );
    }

    let revert_hint = format!(
        "run `moai go-residue-revert <project> --receipt {stamp} --permission-reference <ref>` to put the moved files back"
    );
    for (moved, c) in chosen.iter().enumerate() {
        let from = root.join(&c.path);
        let to_rel = format!("{area}/{}", c.path);
        let step = (|| -> Result<()> {
            make_private_dirs(root, &parent_of(&to_rel)?)?;
            still_as_previewed(&from, &c.path, &c.sha256)?;
            move_exclusive(&from, &root.join(&to_rel))
        })();
        if let Err(e) = step {
            // Record what did move, so the revert and a reader see it; best effort.
            for m in receipt.moves.iter_mut().take(moved) {
                m.moved = true;
            }
            if moved > 0 {
                if let Ok(after) = bytes_of(&receipt) {
                    let _ = crate::apply::publish_replace_limited(
                        &receipt_path,
                        Some(&pending),
                        &after,
                        RECEIPT_MAX,
                    );
                }
                return Err(format!(
                    "partial apply: {moved} of {} files moved; stopped at {}: {e}; {revert_hint}",
                    chosen.len(),
                    c.path
                ));
            }
            return Err(format!(
                "go-residue-apply refused: {}: {e}; no file moved",
                c.path
            ));
        }
    }
    // Read every moved file back before the receipt says `committed`.
    for c in chosen {
        let back = sha256_file(&root.join(format!("{area}/{}", c.path)));
        if !back.is_ok_and(|h| h == c.sha256) || fs::symlink_metadata(root.join(&c.path)).is_ok() {
            return Err(format!(
                "partial apply: {} files moved but {} does not read back from the backup; the \
                 receipt stays pending; {revert_hint}",
                chosen.len(),
                c.path
            ));
        }
    }
    receipt.state = "committed".into();
    for m in &mut receipt.moves {
        m.moved = true;
    }
    let committed = bytes_of(&receipt)?;
    crate::apply::publish_replace_limited(&receipt_path, Some(&pending), &committed, RECEIPT_MAX)
        .map_err(|e| {
        format!(
            "partial apply: {} files moved but the receipt could not be marked committed \
                 ({e}); {revert_hint}",
            chosen.len()
        )
    })?;
    let tracked_moved = chosen.iter().filter(|c| c.tracked).count();
    Ok(json!({
        "state": "committed",
        "receipt": stamp,
        "backup": area,
        "plan_digest": digest,
        "moved": chosen.iter().map(|c| c.path.clone()).collect::<Vec<_>>(),
        "tracked_moved": tracked_moved,
        "warning": (tracked_moved > 0).then(|| format!(
            "{tracked_moved} of the moved files are tracked by git; a checkout brings them back \
             unless the deletions are committed"
        )),
        "revert": {"argv": [
            "moai", "go-residue-revert", root.to_string_lossy(), "--receipt", stamp,
            "--permission-reference", crate::project_update::PERMISSION_PLACEHOLDER,
        ]},
    }))
}

/// Put the files of one apply back. The receipt is untrusted: every path in it passes the guard
/// first (nothing moves when one fails), and each file goes back only when its original place
/// is empty and the backup copy's sha256 is the entry's; anything else is left as it is and
/// reported.
pub fn revert(project: &Path, stamp: &str, permission_reference: &str) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("a permission reference is required".into());
    }
    if !utc_stamp(stamp) {
        return Err("invalid receipt name: expected the apply's UTC stamp YYYYMMDDTHHMMSSZ".into());
    }
    let root = fs::canonicalize(project)
        .map_err(|_| crate::contract::PROJECT_PATH_UNAVAILABLE.to_string())?;
    let rel = receipt_rel(stamp);
    let text = read_bounded_text(&root, &rel, RECEIPT_MAX)?
        .ok_or_else(|| format!("{rel}: no go-residue receipt"))?;
    let mut receipt: Receipt =
        serde_json::from_str(&text).map_err(|_| format!("{rel}: unreadable receipt"))?;
    if receipt.backup != stamp || receipt.project != root.to_string_lossy() {
        return Err("receipt scope mismatch; refused".into());
    }
    if !matches!(receipt.state.as_str(), "pending" | "committed" | "reverted") {
        return Err("receipt state is not one an apply writes; refused".into());
    }
    let area = area_rel(stamp);
    for m in &receipt.moves {
        guard(&root, &m.path)?;
        guard(&root, &format!("{area}/{}", m.path))?;
        let hex_ok = m
            .sha256
            .strip_prefix("sha256:")
            .is_some_and(|h| h.len() == 64 && h.bytes().all(|b| b.is_ascii_hexdigit()));
        if !hex_ok {
            return Err(format!(
                "{}: the receipt holds no sha256 for it; refused",
                m.path
            ));
        }
    }
    if receipt.state == "reverted" {
        return Ok(
            json!({"state": "reverted", "receipt": stamp, "restored": [],
                         "conflicts": [], "missing": [], "note": "already reverted"}),
        );
    }
    let (mut restored, mut conflicts, mut missing) = (Vec::new(), Vec::new(), Vec::new());
    for m in &receipt.moves {
        let original = root.join(&m.path);
        let backup = root.join(format!("{area}/{}", m.path));
        let original_there = fs::symlink_metadata(&original).is_ok();
        let backup_there = fs::symlink_metadata(&backup).is_ok();
        if !backup_there {
            // Never moved (the original is still there) or lost (neither is there).
            if !original_there {
                missing.push(m.path.clone());
            }
            continue;
        }
        if original_there || !sha256_file(&backup).is_ok_and(|h| h == m.sha256) {
            conflicts.push(m.path.clone());
            continue;
        }
        let put_back = (|| -> Result<()> {
            make_dirs(&root, &parent_of(&m.path)?)?;
            still_as_previewed(&backup, &m.path, &m.sha256)?;
            move_exclusive(&backup, &original)
        })();
        match put_back {
            Ok(()) => restored.push(m.path.clone()),
            Err(e) => {
                return Err(format!(
                    "partial apply: the revert restored {} files and stopped at {}: {e}; run the \
                     same revert again to continue",
                    restored.len(),
                    m.path
                ))
            }
        }
    }
    let clean = conflicts.is_empty() && missing.is_empty();
    if clean {
        let before = fs::read(root.join(&rel)).map_err(|_| "receipt read failed".to_string())?;
        receipt.state = "reverted".into();
        for m in &mut receipt.moves {
            m.moved = false;
        }
        let after = bytes_of(&receipt)?;
        crate::apply::publish_replace_limited(&root.join(&rel), Some(&before), &after, RECEIPT_MAX)
            .map_err(|e| format!("partial apply: the files are back but the receipt could not be marked reverted ({e})"))?;
    }
    Ok(json!({
        "state": if clean { "reverted" } else { "conflict" },
        "receipt": stamp,
        "restored": restored,
        "conflicts": conflicts,
        "missing": missing,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;
    use tempfile::TempDir;

    fn project() -> (TempDir, std::path::PathBuf) {
        let tmp = TempDir::new().unwrap();
        let root = fs::canonicalize(tmp.path()).unwrap().join("project");
        fs::create_dir_all(root.join(".moai/policies")).unwrap();
        (tmp, root)
    }

    fn file(root: &Path, rel: &str, body: &str) -> Candidate {
        fs::create_dir_all(root.join(rel).parent().unwrap()).unwrap();
        fs::write(root.join(rel), body).unwrap();
        Candidate {
            path: rel.into(),
            class: "go_same".into(),
            sha256: format!("sha256:{:x}", Sha256::digest(body.as_bytes())),
            tracked: false,
        }
    }

    /// REQ-3 (the apply's own guard): a candidate that leaves the project, in every shape, is
    /// refused before anything is created, even when the plan that named it was doctored.
    #[test]
    fn the_apply_refuses_a_candidate_that_leaves_the_project() {
        let (tmp, root) = project();
        let base = root.parent().unwrap();
        fs::create_dir_all(base.join("outside")).unwrap();
        fs::write(base.join("outside/x.md"), "# outside\n").unwrap();
        fs::create_dir_all(base.join("outside_dir")).unwrap();
        std::os::unix::fs::symlink("../../outside_dir", root.join(".moai/linkdir")).unwrap();
        fs::write(base.join("outside_dir/y.md"), "# y\n").unwrap();
        let abs = base.join("outside/x.md").to_str().unwrap().to_string();
        for path in [
            "../outside/x.md".to_string(),
            abs,
            ".moai/linkdir/y.md".to_string(),
            ".moai/.env".to_string(),
            ".moai/config/credentials.json".to_string(),
        ] {
            let c = Candidate {
                path,
                class: "go_same".into(),
                sha256: format!("sha256:{:x}", Sha256::digest(b"# outside\n")),
                tracked: false,
            };
            let refused = apply_chosen(
                &root,
                "20261125T000000Z",
                std::slice::from_ref(&c),
                "d",
                "ref",
                "p",
            );
            assert!(refused.is_err(), "{}: {refused:?}", c.path);
            assert!(
                fs::symlink_metadata(root.join(".moai/backups")).is_err(),
                "{}: the backup area was created",
                c.path
            );
        }
        assert_eq!(fs::read(base.join("outside/x.md")).unwrap(), b"# outside\n");
        assert_eq!(fs::read(base.join("outside_dir/y.md")).unwrap(), b"# y\n");
        drop(tmp);
    }

    /// REQ-7 and REQ-9 (AC-5): a destination that already exists stops the apply at that file,
    /// the receipt still lists every planned move, and the revert restores what did move.
    #[test]
    fn a_taken_destination_stops_the_apply_and_the_revert_restores_what_moved() {
        let (tmp, root) = project();
        let stamp = "20261125T000000Z";
        let a = file(&root, ".moai/policies/a.md", "# a\n");
        let b = file(&root, ".moai/policies/b.md", "# b\n");
        let c = file(&root, ".moai/policies/c.md", "# c\n");
        let taken = root.join(format!("{}/.moai/policies/b.md", area_rel(stamp)));
        fs::create_dir_all(taken.parent().unwrap()).unwrap();
        fs::write(&taken, "# someone else's file\n").unwrap();
        let err = apply_chosen(
            &root,
            stamp,
            &[a.clone(), b.clone(), c.clone()],
            "d",
            "ref",
            "p",
        )
        .unwrap_err();
        assert!(err.contains("partial apply"), "{err}");
        assert!(err.contains(".moai/policies/b.md"), "{err}");
        assert!(fs::symlink_metadata(root.join(&a.path)).is_err(), "a moved");
        assert_eq!(fs::read(root.join(&b.path)).unwrap(), b"# b\n", "b stays");
        assert_eq!(fs::read(root.join(&c.path)).unwrap(), b"# c\n", "c stays");
        let receipt: Receipt =
            serde_json::from_slice(&fs::read(root.join(receipt_rel(stamp))).unwrap()).unwrap();
        assert_eq!(receipt.state, "pending");
        assert_eq!(receipt.moves.len(), 3, "every planned move stays listed");
        let back = revert(&root, stamp, "ref").unwrap();
        assert_eq!(back["restored"], json!([".moai/policies/a.md"]), "{back}");
        assert_eq!(fs::read(root.join(&a.path)).unwrap(), b"# a\n");
        assert_eq!(fs::read(&taken).unwrap(), b"# someone else's file\n");
        drop(tmp);
    }
}
