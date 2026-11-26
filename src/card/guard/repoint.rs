//! `moai card guard-repoint` (SPEC-20261126-guard-record-repoint): re-point a stale guard record.
//!
//! The guard record is the file [`RECORD_FILE`] in the common git folder. After the profile's
//! integration branch changes it still names the old branch, so the new branch is not guarded.
//! This module shows that state and, only with the preview's digest and a permission reference,
//! rewrites the record to the profile's branch:
//!
//! - the preview ([`repoint`] without flags) answers one of five states and writes nothing; only
//!   `stale` carries a plan digest, which binds the record line and bytes, the target ref and the
//!   common git folder (not the tip of the target branch, which moves with every landing);
//! - `--apply` checks every input first and writes nothing on a refusal, then takes the lock
//!   `moai/integration-branch.lock` with an exclusive create, compares the record again under the
//!   lock, writes the receipt `moai/record-repoint/<stamp>.json` with an exclusive create and
//!   renames the lock over the record; a failing step removes the lock and the receipt;
//! - `--revert <stamp>` puts the old bytes back from that receipt, only while the record still
//!   equals the receipt's new bytes, and writes `<stamp>.revert.json` beside it.
//!
//! The installer, the seed and `card land` never rewrite an existing record; this is the only
//! writer, and only the primary worktree answers `stale` or applies.

use super::{
    common_dir_of, create_file, git, hook_text, is_primary_worktree, read_record, Env, HEADS,
    INTENT_FILE, RECORD_FILE,
};
use crate::card::drift::reference_ok;
use crate::card::landed;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

/// The receipts, in the common git folder (not the hook installer's retire folder).
const RECEIPTS: &str = "moai/record-repoint";
/// What the user's pick stands for in a command the answer prints.
const PICK: &str = "<the user's pick reference>";
/// The receipt layout version.
const SCHEMA: u32 = 1;

/// The options of `card guard-repoint`.
pub struct RepointRequest<'a> {
    /// `--apply`: rewrite the record.
    pub apply: bool,
    /// `--revert <stamp>`: restore the bytes a receipt kept.
    pub revert: Option<&'a str>,
    /// `--plan-digest`: the digest of the preview.
    pub plan_digest: Option<&'a str>,
    /// `--permission-reference`: the reference of the user's yes.
    pub permission_reference: Option<&'a str>,
}

/// A record that no longer names the profile's integration branch.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Staleness {
    /// The branch the record names, without `refs/heads/`.
    pub record: String,
    /// The integration branch the profile names.
    pub profile_branch: String,
}

/// The stale record of `project`'s repository, or `None`: `Some` only in the primary worktree,
/// with a readable record that names another ref than the profile's integration branch. A linked
/// worktree, no record, an unreadable record or profile, a profile with no branch and a current
/// record all answer `None`.
pub fn stale(project: &Path) -> Option<Staleness> {
    match probe(project) {
        Ok(Probe::Stale(plan)) => Some(Staleness {
            record: short(&plan.record),
            profile_branch: short(&plan.target),
        }),
        _ => None,
    }
}

/// `card guard-repoint`: see the module comment.
pub fn repoint(project: &Path, req: &RepointRequest<'_>) -> Result<Value> {
    match (req.apply, req.revert) {
        (true, Some(_)) => Err(refused("--apply and --revert cannot be combined")),
        (true, None) => apply(project, req),
        (false, Some(stamp)) => revert(project, stamp, req),
        (false, None) => {
            if req.plan_digest.is_some() || req.permission_reference.is_some() {
                return Err(refused(
                    "--plan-digest and --permission-reference belong to --apply or --revert; \
                     the preview takes none",
                ));
            }
            preview(project)
        }
    }
}

fn refused(why: impl std::fmt::Display) -> String {
    format!("card guard-repoint refused: {why}")
}

fn short(full_ref: &str) -> String {
    full_ref.strip_prefix(HEADS).unwrap_or(full_ref).to_string()
}

// ---- reading the state -----------------------------------------------------------------------

/// A stale record with what the preview and the apply compare.
struct Plan {
    common: PathBuf,
    /// The full ref the record names.
    record: String,
    /// The full ref of the profile's integration branch.
    target: String,
    /// The record file as read.
    bytes: Vec<u8>,
}

/// What the repository says about the record.
enum Probe {
    Linked,
    NoRecord,
    NoProfileBranch,
    /// The record already names the profile's branch (the full ref).
    Current(String),
    Stale(Plan),
}

fn probe(project: &Path) -> Result<Probe> {
    let common = common_dir_of(project).map_err(refused)?;
    if !is_primary_worktree(project) {
        return Ok(Probe::Linked);
    }
    let bytes = match fs::read(common.join(RECORD_FILE)) {
        Ok(bytes) => bytes,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(Probe::NoRecord),
        Err(e) => return Err(refused(format!("the record cannot be read: {e}"))),
    };
    let Some(record) = read_record(&common) else {
        return Ok(Probe::NoRecord);
    };
    let Some(target) = landed::profile_branch(project)
        .ok()
        .and_then(|name| landed::branch_ref(&name).ok())
    else {
        return Ok(Probe::NoProfileBranch);
    };
    if record == target {
        return Ok(Probe::Current(target));
    }
    Ok(Probe::Stale(Plan {
        common,
        record,
        target,
        bytes,
    }))
}

/// Whether `reference` exists as a local ref (not a tag or a remote-tracking ref of that name).
fn ref_exists(project: &Path, reference: &str) -> bool {
    git(
        Env::Project(project),
        &["show-ref", "--verify", "--quiet", reference],
    )
    .is_ok_and(|out| out.status.success())
}

/// The reference-transaction hook as the common git folder holds it.
fn hook_state(common: &Path) -> &'static str {
    match fs::read(common.join("hooks").join(crate::git_hooks::GUARD_HOOK)) {
        Ok(bytes) if bytes == hook_text().as_bytes() => "managed",
        Ok(_) => "other",
        Err(_) => "absent",
    }
}

fn intent_in_flight(common: &Path) -> bool {
    fs::symlink_metadata(common.join(INTENT_FILE)).is_ok()
}

fn lock_path(common: &Path) -> PathBuf {
    common.join(format!("{RECORD_FILE}.lock"))
}

// ---- digest, hex ------------------------------------------------------------------------------

fn hex_of(bytes: &[u8]) -> String {
    bytes.iter().map(|b| format!("{b:02x}")).collect()
}

fn sha_hex(bytes: &[u8]) -> String {
    hex_of(&Sha256::digest(bytes))
}

fn unhex(text: &str) -> Option<Vec<u8>> {
    if !text.len().is_multiple_of(2) {
        return None;
    }
    (0..text.len())
        .step_by(2)
        .map(|i| u8::from_str_radix(text.get(i..i + 2)?, 16).ok())
        .collect()
}

/// The plan digest: the record line and bytes, the target ref and the common git folder, each
/// part length-prefixed. The tip of the target branch is not part of it.
fn digest_of(plan: &Plan) -> String {
    let common = fs::canonicalize(&plan.common).unwrap_or_else(|_| plan.common.clone());
    let common = common.display().to_string();
    let mut hash = Sha256::new();
    for part in [
        b"moai card guard-repoint plan v1".as_slice(),
        plan.record.as_bytes(),
        &plan.bytes,
        plan.target.as_bytes(),
        common.as_bytes(),
    ] {
        hash.update((part.len() as u64).to_le_bytes());
        hash.update(part);
    }
    hex_of(&hash.finalize())
}

// ---- the preview ------------------------------------------------------------------------------

fn state_answer(state: &str, why: &str) -> Value {
    json!({"action": "guard-repoint", "state": state, "writes": false, "why": why})
}

fn preview(project: &Path) -> Result<Value> {
    Ok(match probe(project)? {
        Probe::Linked => state_answer(
            "linked_worktree",
            "this is a linked worktree: the record is read and changed from the primary worktree",
        ),
        Probe::NoRecord => state_answer(
            "no_record",
            "the repository holds no guard record; the installer seeds one from the profile",
        ),
        Probe::NoProfileBranch => state_answer(
            "no_profile_branch",
            "the profile names no integration branch to re-point the record to",
        ),
        Probe::Current(target) => {
            let mut out = state_answer("current", "the record names the profile's branch");
            out["record"] = json!(target);
            out["target"] = json!(target);
            out
        }
        Probe::Stale(plan) => stale_answer(project, &plan),
    })
}

fn stale_answer(project: &Path, plan: &Plan) -> Value {
    let digest = digest_of(plan);
    let apply = format!(
        "moai card guard-repoint --apply --plan-digest {digest} --permission-reference {PICK}"
    );
    json!({
        "action": "guard-repoint",
        "state": "stale",
        "writes": false,
        "record": plan.record,
        "target": plan.target,
        "branch_exists": ref_exists(project, &plan.target),
        "hook": hook_state(&plan.common),
        "intent_in_flight": intent_in_flight(&plan.common),
        "unguarded_after": plan.record,
        "plan_digest": digest,
        "apply": apply,
    })
}

// ---- the apply --------------------------------------------------------------------------------

/// The reference of the user's yes: present and valid, or the refusal.
fn valid_reference(given: Option<&str>) -> Result<&str> {
    let Some(reference) = given else {
        return Err(refused(
            "--permission-reference is needed, the reference of the user's yes; nothing was \
             changed",
        ));
    };
    if !reference_ok(reference) {
        return Err(refused(
            "--permission-reference holds 1 to 80 letters, digits or . _ : / -; nothing was \
             changed",
        ));
    }
    Ok(reference)
}

fn apply(project: &Path, req: &RepointRequest<'_>) -> Result<Value> {
    let Some(given) = req.plan_digest else {
        return Err(refused(
            "--apply needs --plan-digest from a preview of the same record; nothing was changed",
        ));
    };
    let reference = valid_reference(req.permission_reference)?;
    let plan = match probe(project)? {
        Probe::Stale(plan) => plan,
        Probe::Linked => {
            return Err(refused(
                "this is a linked worktree: the record is changed from the primary worktree \
                 only; nothing was changed",
            ))
        }
        Probe::NoRecord | Probe::NoProfileBranch | Probe::Current(_) => {
            return Err(refused(
                "the record is not stale (run the preview for its state); nothing was changed",
            ))
        }
    };
    if digest_of(&plan) != given {
        return Err(refused(
            "the plan digest is not the digest of the record now (the record or the target \
             changed since the preview): run the preview again; nothing was changed",
        ));
    }
    if !ref_exists(project, &plan.target) {
        return Err(refused(format!(
            "{} is not an existing local branch: create it first; nothing was changed",
            plan.target
        )));
    }
    if intent_in_flight(&plan.common) {
        return Err(refused(
            "a landing is in flight (the file moai/land-intent exists): re-point after it \
             ends; nothing was changed",
        ));
    }
    let lock = lock_path(&plan.common);
    if fs::symlink_metadata(&lock).is_ok() {
        return Err(lock_refusal(&lock));
    }
    write_repoint(&plan, given, reference)
}

fn lock_refusal(lock: &Path) -> String {
    refused(format!(
        "the lock {} exists: another run holds it or one stopped early; remove it by hand \
         after checking that no other `moai card guard-repoint` is running; nothing was changed",
        lock.display()
    ))
}

fn write_repoint(plan: &Plan, digest: &str, reference: &str) -> Result<Value> {
    let now = crate::audit::utc_now();
    let stamp = now.replace(['-', ':'], "");
    let new_bytes = format!("{}\n", plan.target).into_bytes();
    let receipt = json!({
        "schema": SCHEMA,
        "kind": "guard-repoint",
        "stamp": stamp,
        "applied_at": now,
        "old_record": plan.record,
        "new_record": plan.target,
        "old_bytes_sha256": sha_hex(&plan.bytes),
        "new_bytes_sha256": sha_hex(&new_bytes),
        // The receipt keeps the old bytes themselves, so a revert restores them exactly.
        "old_bytes_hex": hex_of(&plan.bytes),
        "profile_branch": short(&plan.target),
        "plan_digest": digest,
        "permission_reference": reference,
        "hook_present": hook_state(&plan.common) != "absent",
    });
    let path = plan.common.join(RECEIPTS).join(format!("{stamp}.json"));
    swap_record(&plan.common, &plan.bytes, &new_bytes, &path, &receipt)?;
    Ok(json!({
        "action": "guard-repoint",
        "applied": true,
        "writes": true,
        "stamp": stamp,
        "receipt": path.display().to_string(),
        "record": plan.target,
        "unguarded": plan.record,
        "revert": format!(
            "moai card guard-repoint --revert {stamp} --permission-reference {PICK}"
        ),
    }))
}

/// Replaces the record bytes `expect` by `new` under the lock and writes `receipt` first. Any step
/// after the lock was taken that fails removes the lock and the receipt and leaves the record as
/// it was; a lock that another run holds is never removed.
fn swap_record(
    common: &Path,
    expect: &[u8],
    new: &[u8],
    receipt_path: &Path,
    receipt: &Value,
) -> Result<()> {
    let lock = lock_path(common);
    let mut file = match create_file(&lock, false) {
        Ok(file) => file,
        Err(e) if e.kind() == std::io::ErrorKind::AlreadyExists => return Err(lock_refusal(&lock)),
        Err(e) => {
            return Err(refused(format!(
                "cannot take the lock {}: {e}; nothing was changed",
                lock.display()
            )))
        }
    };
    let mut receipt_made = false;
    let outcome = write_locked(
        common,
        &mut file,
        expect,
        new,
        (receipt_path, receipt),
        &mut receipt_made,
    );
    drop(file);
    if outcome.is_err() {
        let _ = fs::remove_file(&lock);
        if receipt_made {
            let _ = fs::remove_file(receipt_path);
        }
    }
    outcome
}

fn write_locked(
    common: &Path,
    lock: &mut fs::File,
    expect: &[u8],
    new: &[u8],
    receipt: (&Path, &Value),
    receipt_made: &mut bool,
) -> Result<()> {
    let record = common.join(RECORD_FILE);
    lock.write_all(new)
        .map_err(|e| refused(format!("cannot write the lock: {e}")))?;
    match fs::read(&record) {
        Ok(now) if now == expect => {}
        _ => {
            return Err(refused(
                "the record changed while the lock was taken: run the preview again; nothing \
                 was changed",
            ))
        }
    }
    let (path, body) = receipt;
    write_new_json(path, body, receipt_made)?;
    fs::rename(lock_path(common), &record)
        .map_err(|e| refused(format!("cannot replace the record: {e}")))
}

/// Writes `body` to a file that must not exist yet, making its folder; `made` is set once the
/// file exists, so the caller can remove it.
fn write_new_json(path: &Path, body: &Value, made: &mut bool) -> Result<()> {
    let folder = path
        .parent()
        .ok_or_else(|| refused("no folder for a receipt"))?;
    fs::create_dir_all(folder).map_err(|e| {
        refused(format!(
            "cannot make the receipt folder {}: {e}",
            folder.display()
        ))
    })?;
    let mut file = create_file(path, false).map_err(|e| {
        if e.kind() == std::io::ErrorKind::AlreadyExists {
            refused(format!(
                "the receipt {} exists already (a second run in the same second): try again",
                path.display()
            ))
        } else {
            refused(format!("cannot write the receipt {}: {e}", path.display()))
        }
    })?;
    *made = true;
    let text = serde_json::to_string_pretty(body).map_err(refused)? + "\n";
    file.write_all(text.as_bytes())
        .map_err(|e| refused(format!("cannot write the receipt {}: {e}", path.display())))
}

// ---- the revert -------------------------------------------------------------------------------

/// What a receipt must hold for a revert.
struct Receipt {
    old_record: String,
    new_record: String,
    old_bytes: Vec<u8>,
    new_bytes: Vec<u8>,
}

fn read_receipt(path: &Path, stamp: &str) -> Result<Receipt> {
    let bad = |why: &str| {
        refused(format!(
            "the receipt {} {why}; nothing was changed",
            path.display()
        ))
    };
    match fs::symlink_metadata(path) {
        Ok(meta) if meta.is_file() => {}
        Ok(_) => return Err(bad("is not a regular file")),
        Err(_) => return Err(bad("is missing")),
    }
    let text = fs::read_to_string(path).map_err(|_| bad("cannot be read"))?;
    let value: Value = serde_json::from_str(&text).map_err(|_| bad("is not valid JSON"))?;
    let field = |key: &str| value.get(key).and_then(Value::as_str).map(str::to_string);
    let (Some(old_record), Some(new_record), Some(old_sha), Some(new_sha), Some(old_hex)) = (
        field("old_record"),
        field("new_record"),
        field("old_bytes_sha256"),
        field("new_bytes_sha256"),
        field("old_bytes_hex"),
    ) else {
        return Err(bad("is missing a field a revert needs"));
    };
    if field("kind").as_deref() != Some("guard-repoint") || field("stamp").as_deref() != Some(stamp)
    {
        return Err(bad("is not the receipt of this stamp"));
    }
    let old_bytes = unhex(&old_hex).ok_or_else(|| bad("holds unreadable old bytes"))?;
    let new_bytes = format!("{new_record}\n").into_bytes();
    if sha_hex(&old_bytes) != old_sha || sha_hex(&new_bytes) != new_sha {
        return Err(bad("does not match its own hashes"));
    }
    Ok(Receipt {
        old_record,
        new_record,
        old_bytes,
        new_bytes,
    })
}

fn revert(project: &Path, stamp: &str, req: &RepointRequest<'_>) -> Result<Value> {
    if req.plan_digest.is_some() {
        return Err(refused("--plan-digest belongs to --apply"));
    }
    let reference = valid_reference(req.permission_reference)?;
    if !crate::state_root::utc_stamp(stamp) {
        return Err(refused(format!(
            "the stamp `{stamp}` is not YYYYMMDDTHHMMSSZ"
        )));
    }
    let common = common_dir_of(project).map_err(refused)?;
    let folder = common.join(RECEIPTS);
    let receipt = read_receipt(&folder.join(format!("{stamp}.json")), stamp)?;
    let revert_path = folder.join(format!("{stamp}.revert.json"));
    if fs::symlink_metadata(&revert_path).is_ok() {
        return Err(refused(format!(
            "the receipt of {stamp} is already reverted ({} exists); nothing was changed",
            revert_path.display()
        )));
    }
    let now = crate::audit::utc_now();
    let body = json!({
        "stamp": stamp,
        "reverted_at": now,
        "permission_reference": reference,
        "restored_record": receipt.old_record,
    });
    swap_record(
        &common,
        &receipt.new_bytes,
        &receipt.old_bytes,
        &revert_path,
        &body,
    )?;
    Ok(json!({
        "action": "guard-repoint",
        "reverted": true,
        "writes": true,
        "stamp": stamp,
        "receipt": revert_path.display().to_string(),
        "record": receipt.old_record,
        "was": receipt.new_record,
    }))
}
