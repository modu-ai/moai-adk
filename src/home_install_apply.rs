//! #2871 U3c (PLAN comment 6881562064, user decision D-M1 6876633358, leader decision 6886405676):
//! the apply and the revert of the move from `~/.moai-cli` to the `~/.moai` top level that
//! [`crate::home_install::plan`] previews. Each `move` entry of a fresh, identical preview goes by
//! one rename (the per-item journal of [`crate::move_journal`]); then the `path` links are
//! re-pointed at the moved `versions`; the schema-2 marker is written last, so every reader
//! flips to `~/.moai` only when everything is in place. Nothing is copied or deleted, a secret
//! is judged by `lstat` and never opened, and a link is never followed.
//!
//! Apply order: permission reference, the journal folder is a real folder, the manifest of this
//! stamp (an unfinished install resumes without a new preview; a finished one is `unchanged`),
//! otherwise a fresh preview whose `plan_digest` must match, a live Rust session of the clone
//! blocks, the install lock (`home-install/apply.lock`, never waited for), the same checks again
//! under the lock, the immutable manifest, `~/.moai`, the journalled moves (`versions` and `path`
//! last), a rollback when any entry did not move, the links, the vault facts, the marker.
//!
//! Revert order: permission reference, `~/.moai` is a real folder, the marker is still the one
//! the apply wrote (an edited marker is refused) and goes into the journal folder, the links go
//! back where their target is still the new one, the journal is replayed in reverse, and
//! `~/.moai` is removed only when the apply created it and it is empty.
//!
//! Every step leaves a state the other half understands: a crash after any step is finished by
//! applying again with the same `--now` and digest, or undone by the revert.

use crate::home_install::{self, ALIAS_JOURNAL_LIMIT, JOURNAL_DIR, KIND, MANIFEST_PREFIX};
use crate::home_layout;
use crate::home_state::{stamp, InUse};
use crate::move_journal::{self, JournalState, MoveEnv, PlannedMove};
use crate::state_root::{
    classify_home, utc_stamp, LEGACY_GENERATION, LEGACY_MARKER, MOAI_DIR, NEW_GENERATION,
    NEW_MARKER,
};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};

/// The lock file inside the journal folder.
const LOCK_FILE: &str = "apply.lock";
/// The marker the revert takes out of `~/.moai` is kept here, beside the manifest.
const REVERTED_MARKER_PREFIX: &str = "reverted-marker-";
/// Who the marker names as its writer.
const CREATED_BY: &str = "home-install-apply";
/// Debug builds only: `MOAI_CLI_TEST_INSTALL_FAULT=<point>` stops the apply with an error at that
/// point, leaving everything as a crash would. The points are `after_manifest`, `after_mkdir`,
/// the journal's own (`after_journal`, `after_mark_moving`, `after_rename`, `after_move:<n>`,
/// `before_commit`), `after_commit`, `after_stub:<n>`, `before_marker` and `after_marker`.
pub const FAULT_VAR: &str = "MOAI_CLI_TEST_INSTALL_FAULT";
/// A manifest holds at most 500 entries, 500 aliases and 50 links; a larger file is not one.
const MANIFEST_LIMIT: u64 = 4 * 1024 * 1024;
/// A marker is a few lines; a larger file is not the one the apply wrote.
const MARKER_LIMIT: u64 = 64 * 1024;
/// The entries that move last: if anything fails early, the build and its links are untouched.
const MOVE_LAST: [&str; 2] = ["versions", "path"];
const STALE: &str =
    "stale or altered plan: the recomputed digest differs; re-preview required, nothing was written";

type Res<T> = Result<T, String>;

#[cfg(debug_assertions)]
fn fault_knob() -> Option<String> {
    std::env::var(FAULT_VAR).ok().filter(|v| !v.is_empty())
}

#[cfg(not(debug_assertions))]
fn fault_knob() -> Option<String> {
    None
}

/// Stop here when the test knob names this point (never in a release build).
fn stop_at(point: &str) -> Res<()> {
    if fault_knob().as_deref() == Some(point) {
        return Err(format!("injected crash at {point}"));
    }
    Ok(())
}

/// The folders the install works in.
struct Roots {
    old: PathBuf,
    journal: PathBuf,
    moai: PathBuf,
}

impl Roots {
    fn of(home: &Path) -> Roots {
        let old = home_install::old_root(home);
        Roots {
            journal: old.join(JOURNAL_DIR),
            moai: home.join(MOAI_DIR),
            old,
        }
    }

    fn manifest(&self, stamp: &str) -> PathBuf {
        self.journal.join(format!("{MANIFEST_PREFIX}{stamp}.json"))
    }

    fn move_journal(&self, stamp: &str) -> PathBuf {
        move_journal::journal_path(&self.journal, stamp)
    }

    fn reverted_marker(&self, stamp: &str) -> PathBuf {
        self.journal
            .join(format!("{REVERTED_MARKER_PREFIX}{stamp}.json"))
    }
}

#[derive(Clone, Debug, Serialize, Deserialize)]
struct Item {
    name: String,
    kind: String,
}

/// A `path` link that is re-pointed: its target before the move. The target after is derived
/// from it, never read back from the file.
#[derive(Clone, Debug, Serialize, Deserialize)]
struct StubMove {
    name: String,
    target: String,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
struct MarkerPlan {
    /// The exact bytes the apply writes, so a resume by another build writes the same ones.
    content: String,
    sha256: String,
}

/// The immutable record of one install: what was decided, before anything moved.
#[derive(Clone, Debug, Serialize, Deserialize)]
struct Manifest {
    schema_version: u32,
    kind: String,
    stamp: String,
    home: String,
    created_at: String,
    plan_digest: String,
    content_digest: String,
    permission_reference: String,
    /// Whether `~/.moai` did not exist when the apply decided, so only then may it be removed.
    moai_created: bool,
    /// The permission bits of `~/.moai-cli`; `~/.moai` and the marker mirror them.
    dir_mode: u32,
    items: Vec<Item>,
    aliases: BTreeMap<String, String>,
    stubs: Vec<StubMove>,
    vault: Value,
    marker: MarkerPlan,
    kept: Vec<Value>,
}

/// A name that is one path component: a manifest is a file in the user's HOME, so nothing in it
/// may lead a rename or a link out of the folders it names.
fn plain_name(n: &str) -> bool {
    !n.is_empty() && n != "." && n != ".." && !n.contains(['/', '\\', '\0'])
}

impl Manifest {
    fn validate(&self, home: &Path, stamp: &str) -> Res<()> {
        let bad = |why: &str| {
            Err(format!(
                "home install manifest {stamp} is invalid: {why}; nothing was changed"
            ))
        };
        if self.schema_version != 1 || self.kind != KIND || self.stamp != stamp {
            return bad("another schema, kind or stamp");
        }
        if self.home != home.display().to_string() {
            return bad("it belongs to another HOME");
        }
        if self.items.is_empty() || self.items.iter().any(|i| !plain_name(&i.name)) {
            return bad("an entry name is not one path component");
        }
        let old_versions = home_install::old_root(home).join("versions");
        if self
            .stubs
            .iter()
            .any(|s| !plain_name(&s.name) || new_target(home, &s.target, &old_versions).is_none())
        {
            return bad("a link is not one into the earlier versions folder");
        }
        if crate::digest(self.marker.content.as_bytes()) != self.marker.sha256 {
            return bad("the marker digest does not match its content");
        }
        if self.marker_file().is_none() {
            return bad("the marker names a generation this build does not know");
        }
        Ok(())
    }

    /// The file name of the marker this install wrote, from the `generation` inside the saved
    /// bytes: `moai-cli` is the earlier name, `moai` the new one (#2871 B1). A build that resumes
    /// or reverts an install another build began must write and find the file under the name
    /// that matches the bytes, or it publishes a name and generation that `state_root` rejects as
    /// a pair, or overlooks the marker of a finished install. A new install carries the new
    /// generation, so it is written under the new name.
    fn marker_file(&self) -> Option<&'static str> {
        let v: Value = serde_json::from_str(&self.marker.content).ok()?;
        match v["generation"].as_str()? {
            LEGACY_GENERATION => Some(LEGACY_MARKER),
            NEW_GENERATION => Some(NEW_MARKER),
            _ => None,
        }
    }
}

/// Where a link into the earlier `versions` folder points once that folder has moved.
fn new_target(home: &Path, old_target: &str, old_versions: &Path) -> Option<PathBuf> {
    let rest = Path::new(old_target).strip_prefix(old_versions).ok()?;
    Some(home_layout::rust_path(home, "versions").join(rest))
}

fn io<E: std::fmt::Display>(p: &Path, e: E) -> String {
    format!("{}: {e}", p.display())
}

fn canon(p: &Path) -> PathBuf {
    fs::canonicalize(p).unwrap_or_else(|_| p.to_path_buf())
}

fn lstat(p: &Path) -> std::io::Result<fs::Metadata> {
    fs::symlink_metadata(p)
}

fn not_found(e: &std::io::Error) -> bool {
    e.kind() == std::io::ErrorKind::NotFound
}

/// Create a folder with these permission bits exactly (the umask would narrow the mode of the
/// create call alone).
fn make_dir(p: &Path, mode: u32) -> Res<()> {
    let b = fs::DirBuilder::new();
    #[cfg(unix)]
    let b = {
        let mut b = b;
        std::os::unix::fs::DirBuilderExt::mode(&mut b, mode);
        b
    };
    b.create(p).map_err(|e| io(p, e))?;
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(p, fs::Permissions::from_mode(mode)).map_err(|e| io(p, e))?;
    }
    #[cfg(not(unix))]
    let _ = mode;
    Ok(())
}

fn sync_parent(p: &Path) {
    if let Some(dir) = p.parent().and_then(|d| fs::File::open(d).ok()) {
        let _ = dir.sync_all();
    }
}

/// Create `path` exclusively (0600) and write `bytes` into it, fsynced.
fn write_new(path: &Path, bytes: &[u8]) -> Res<()> {
    use std::io::Write;
    let mut o = fs::OpenOptions::new();
    o.write(true).create_new(true);
    #[cfg(unix)]
    std::os::unix::fs::OpenOptionsExt::mode(&mut o, 0o600);
    let mut f = o.open(path).map_err(|e| io(path, e))?;
    f.write_all(bytes)
        .and_then(|()| f.sync_all())
        .map_err(|e| io(path, e))?;
    sync_parent(path);
    Ok(())
}

/// The journal folder is a real folder or absent. A link there would send the lock, the manifest
/// and the journal to wherever it points.
fn check_journal_root(journal: &Path) -> Res<()> {
    match lstat(journal) {
        Err(e) if not_found(&e) => Ok(()),
        Err(e) => Err(format!("{}; nothing was written", io(journal, e))),
        Ok(m) if m.is_dir() => Ok(()),
        Ok(_) => Err(format!(
            "home install refused: {} is a link or not a folder; nothing was written",
            journal.display()
        )),
    }
}

/// The names in the journal folder; a name or an entry that cannot be read fails the call, since
/// a manifest or journal among them could be the unfinished one.
fn journal_names(journal: &Path) -> Res<Vec<String>> {
    let mut names = Vec::new();
    for e in fs::read_dir(journal).map_err(|e| io(journal, e))? {
        let e = e.map_err(|e| io(journal, e))?;
        let name = e.file_name().into_string().map_err(|_| {
            format!(
                "{} holds a name that is not UTF-8; nothing was changed",
                journal.display()
            )
        })?;
        names.push(name);
    }
    names.sort();
    Ok(names)
}

/// The stamps of the install manifests in `journal`, oldest first.
fn manifest_stamps(journal: &Path) -> Res<Vec<String>> {
    Ok(journal_names(journal)?
        .iter()
        .filter_map(|n| {
            n.strip_prefix(MANIFEST_PREFIX)?
                .strip_suffix(".json")
                .filter(|s| utc_stamp(s))
                .map(str::to_string)
        })
        .collect())
}

/// How many install journals the folder holds.
fn journal_count(journal: &Path) -> Res<usize> {
    Ok(journal_names(journal)?
        .iter()
        .filter(|n| n.starts_with(move_journal::PREFIX) && n.ends_with(".json"))
        .count())
}

fn read_manifest(home: &Path, roots: &Roots, stamp: &str) -> Res<Manifest> {
    let path = roots.manifest(stamp);
    match lstat(&path) {
        Ok(m) if m.is_file() && m.len() <= MANIFEST_LIMIT => {}
        Ok(_) => return Err(format!("{} is not a plain manifest file", path.display())),
        Err(e) => return Err(io(&path, e)),
    }
    let bytes = fs::read(&path).map_err(|e| io(&path, e))?;
    let m: Manifest = serde_json::from_slice(&bytes).map_err(|e| io(&path, e))?;
    m.validate(home, stamp)?;
    Ok(m)
}

/// Where the marker of this install is in `~/.moai`: the name its saved bytes go with.
fn marker_path(roots: &Roots, m: &Manifest) -> Res<PathBuf> {
    let name = m.marker_file().ok_or_else(|| {
        format!(
            "home install manifest {} is invalid: the marker names a generation this build does not know; nothing was changed",
            m.stamp
        )
    })?;
    Ok(roots.moai.join(name))
}

/// Whether the marker the manifest names is in `~/.moai`: the same bytes, by digest.
fn marker_is_ours(roots: &Roots, m: &Manifest) -> Res<bool> {
    let p = marker_path(roots, m)?;
    match lstat(&p) {
        Err(e) if not_found(&e) => Ok(false),
        Err(e) => Err(io(&p, e)),
        Ok(md) if md.is_file() && md.len() <= MARKER_LIMIT => {
            let bytes = fs::read(&p).map_err(|e| io(&p, e))?;
            Ok(crate::digest(&bytes) == m.marker.sha256)
        }
        Ok(_) => Ok(false),
    }
}

/// Where an install stands, from its journal and the marker.
#[derive(Debug, PartialEq, Eq)]
enum State {
    /// Started and not finished: apply again to finish it.
    Unfinished,
    /// Done: moved, links re-pointed, marker written.
    Installed,
    /// Rolled back or reverted: nothing is left to do.
    Ended,
    /// A revert took the marker out and has not finished.
    Reverting,
    /// A revert left an entry where it was.
    Partial,
}

fn state_of(roots: &Roots, m: &Manifest) -> Res<State> {
    let jp = roots.move_journal(&m.stamp);
    let journal = match lstat(&jp) {
        Err(e) if not_found(&e) => None,
        Err(e) => return Err(io(&jp, e)),
        Ok(_) => Some(move_journal::read_journal(&jp)?.state),
    };
    let reverting = lstat(&roots.reverted_marker(&m.stamp)).is_ok();
    Ok(match journal {
        Some(JournalState::Reverted) => State::Ended,
        Some(JournalState::RevertPartial) => State::Partial,
        _ if reverting => State::Reverting,
        None | Some(JournalState::Pending) => State::Unfinished,
        Some(JournalState::Committed) => {
            if marker_is_ours(roots, m)? {
                State::Installed
            } else {
                State::Unfinished
            }
        }
    })
}

/// What the apply does with this call.
enum Next {
    /// A fresh preview that matched: start an install.
    Fresh(Box<Value>),
    /// An install of this stamp is unfinished: finish it.
    Resume(Box<Manifest>),
    /// An install of this stamp is finished.
    Unchanged(Box<Manifest>),
}

/// Judge this call from what is on disk, reading only. An error here has written nothing.
fn precheck(
    home: &Path,
    moai_home: Option<&str>,
    now: &str,
    stamp: &str,
    plan_digest: &str,
) -> Res<Next> {
    let roots = Roots::of(home);
    if lstat(&roots.journal).is_ok() {
        let stamps = manifest_stamps(&roots.journal)?;
        if stamps.iter().any(|s| s == stamp) {
            let m = read_manifest(home, &roots, stamp)?;
            if m.plan_digest != plan_digest {
                return Err(STALE.into());
            }
            return match state_of(&roots, &m)? {
                State::Installed => Ok(Next::Unchanged(Box::new(m))),
                State::Unfinished => Ok(Next::Resume(Box::new(m))),
                State::Ended => Err(format!(
                    "install {stamp} was rolled back or reverted; preview again and apply with a new --now; nothing was written"
                )),
                State::Reverting | State::Partial => Err(format!(
                    "a revert of install {stamp} is unfinished; run home-install-revert again; nothing was written"
                )),
            };
        }
        for s in &stamps {
            let other = read_manifest(home, &roots, s)?;
            if matches!(
                state_of(&roots, &other)?,
                State::Unfinished | State::Reverting
            ) {
                return Err(format!(
                    "install {s} is unfinished: apply again with its own --now {} and plan_digest, or run home-install-revert; nothing was written",
                    other.created_at
                ));
            }
        }
    }
    let fresh = home_install::plan(home, moai_home, now)?;
    if fresh["plan_digest"].as_str() != Some(plan_digest) {
        return Err(STALE.into());
    }
    if fresh["action"] != "install" {
        return Err(format!(
            "the preview's action is `{}`, not `install`; nothing was written",
            fresh["action"].as_str().unwrap_or("?")
        ));
    }
    // The lookup that tells readers an earlier install id reads this many journals; a further one
    // would push the oldest one's ids out of it.
    if lstat(&roots.journal).is_ok() && journal_count(&roots.journal)? >= ALIAS_JOURNAL_LIMIT {
        return Err(format!(
            "home install refused: {} already holds {ALIAS_JOURNAL_LIMIT} install journals (the limit); nothing was written",
            roots.journal.display()
        ));
    }
    Ok(Next::Fresh(Box::new(fresh)))
}

/// The install lock: created with the journal folder, held until the returned file drops, never
/// waited for.
fn take_lock(roots: &Roots) -> Res<fs::File> {
    check_journal_root(&roots.journal)?;
    if lstat(&roots.journal).is_err() {
        make_dir(&roots.journal, 0o700)?;
    }
    let path = roots.journal.join(LOCK_FILE);
    match lstat(&path) {
        Err(e) if not_found(&e) => {}
        Ok(m) if m.is_file() => {}
        Ok(_) => {
            return Err(format!(
                "home install refused: {} is a link or not a plain file; nothing was written",
                path.display()
            ))
        }
        Err(e) => return Err(io(&path, e)),
    }
    let mut o = fs::OpenOptions::new();
    o.write(true).create(true).truncate(false);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        o.mode(0o600).custom_flags(libc::O_NOFOLLOW);
    }
    let f = o.open(&path).map_err(|e| io(&path, e))?;
    match f.try_lock() {
        Ok(()) => Ok(f),
        Err(fs::TryLockError::WouldBlock) => Err(
            "conflict in home install: another apply or revert holds the install lock; nothing was written"
                .into(),
        ),
        Err(fs::TryLockError::Error(e)) => Err(io(&path, e)),
    }
}

/// A live Rust session of the clone blocks the apply, and so does a registry that cannot be read.
fn sessions_gate(project: &Path) -> Res<Value> {
    let found = home_install::live_sessions_of(project);
    match &found {
        Ok((_, 0)) => {}
        Ok((_, n)) => {
            return Err(format!(
                "home install blocked: rust_session_live:{n}; end the Rust sessions registered in this clone and apply again (other clones are not visible from the HOME); nothing was written"
            ))
        }
        Err(e) => {
            return Err(format!(
                "home install blocked: rust_session_registry_unknown ({e}); nothing was written"
            ))
        }
    }
    Ok(home_install::session_fact(&found))
}

/// What a call needs while it runs.
struct Ctx<'a> {
    home: &'a Path,
    roots: Roots,
    m: Manifest,
    now: &'a str,
    sessions: Value,
    resumed: bool,
}

/// Move `home`'s `~/.moai-cli` items to `~/.moai`: the entries of a fresh preview that has this
/// `plan_digest` and the action `install`, or, when an install of this stamp was left unfinished,
/// the rest of it (a resume does not preview again). `project` is the clone whose Rust sessions
/// are checked. Writes nothing when it refuses.
pub fn apply(
    home: &Path,
    moai_home: Option<&str>,
    project: &Path,
    now: &str,
    plan_digest: &str,
    permission_reference: &str,
) -> Res<Value> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("permission reference required".into());
    }
    if !home.is_absolute() {
        return Err("HOME is not an absolute path".into());
    }
    let stamp = stamp(now)?;
    let roots = Roots::of(home);
    check_journal_root(&roots.journal)?;
    if let Next::Unchanged(m) = precheck(home, moai_home, now, &stamp, plan_digest)? {
        return unchanged(home, &roots, &m);
    }
    let sessions = sessions_gate(project)?;
    let _lock = take_lock(&roots)?;
    // Under the lock the same checks run again: what they judged before may have changed.
    match precheck(home, moai_home, now, &stamp, plan_digest)? {
        Next::Unchanged(m) => unchanged(home, &roots, &m),
        Next::Resume(m) => run(&Ctx {
            home,
            roots,
            m: *m,
            now,
            sessions,
            resumed: true,
        }),
        Next::Fresh(plan) => {
            let m = build_manifest(home, &roots, now, &stamp, &plan, permission_reference)?;
            let bytes = serde_json::to_vec_pretty(&m).map_err(|e| e.to_string())?;
            write_new(&roots.manifest(&stamp), &[bytes.as_slice(), b"\n"].concat())?;
            stop_at("after_manifest")?;
            run(&Ctx {
                home,
                roots,
                m,
                now,
                sessions,
                resumed: false,
            })
        }
    }
}

fn unchanged(home: &Path, roots: &Roots, m: &Manifest) -> Res<Value> {
    Ok(json!({
        "schema_version": 1,
        "state": "unchanged",
        "stamp": m.stamp,
        "home": home.display().to_string(),
        "manifest": roots.manifest(&m.stamp).display().to_string(),
        "journal": roots.move_journal(&m.stamp).display().to_string(),
        "generation": classify_home(home).generation,
        "writes": false,
    }))
}

/// The entries of the preview that move, `versions` and `path` last.
fn moves_of(plan: &Value) -> Vec<Item> {
    let mut items: Vec<Item> = plan["entries"]
        .as_array()
        .map(|a| {
            a.iter()
                .filter(|e| e["action"] == "move")
                .filter_map(|e| {
                    Some(Item {
                        name: e["name"].as_str()?.to_string(),
                        kind: e["kind"].as_str()?.to_string(),
                    })
                })
                .collect()
        })
        .unwrap_or_default();
    // A stable sort: the others keep the preview's order.
    items.sort_by_key(|i| {
        MOVE_LAST
            .iter()
            .position(|l| i.name.eq_ignore_ascii_case(l))
            .map_or(0, |p| p + 1)
    });
    items
}

fn build_manifest(
    home: &Path,
    roots: &Roots,
    now: &str,
    stamp: &str,
    plan: &Value,
    permission_reference: &str,
) -> Res<Manifest> {
    let text = |k: &str| plan[k].as_str().unwrap_or_default().to_string();
    let aliases = plan["alias_table"]
        .as_object()
        .map(|o| {
            o.iter()
                .filter_map(|(k, v)| Some((k.clone(), v.as_str()?.to_string())))
                .collect()
        })
        .unwrap_or_default();
    let stubs = plan["stub"]
        .as_array()
        .map(|a| {
            a.iter()
                .filter(|s| s["repoint"] == json!(true))
                .filter_map(|s| {
                    Some(StubMove {
                        name: s["name"].as_str()?.to_string(),
                        target: s["target"].as_str()?.to_string(),
                    })
                })
                .collect()
        })
        .unwrap_or_default();
    let marker = String::from_utf8(crate::state_root::home_marker_bytes(CREATED_BY, stamp))
        .map_err(|e| e.to_string())?;
    let m = Manifest {
        schema_version: 1,
        kind: KIND.to_string(),
        stamp: stamp.to_string(),
        home: home.display().to_string(),
        created_at: now.to_string(),
        plan_digest: text("plan_digest"),
        content_digest: text("content_digest"),
        permission_reference: permission_reference.to_string(),
        moai_created: lstat(&roots.moai).is_err_and(|e| not_found(&e)),
        dir_mode: home_install::mode_of(&roots.old).map_or(0o700, |m| m & 0o777),
        items: moves_of(plan),
        aliases,
        stubs,
        vault: plan["vault"].clone(),
        marker: MarkerPlan {
            sha256: crate::digest(marker.as_bytes()),
            content: marker,
        },
        kept: plan["entries"]
            .as_array()
            .map(|a| {
                a.iter()
                    .filter(|e| e["action"] != "move")
                    .cloned()
                    .collect()
            })
            .unwrap_or_default(),
    };
    m.validate(home, stamp)?;
    Ok(m)
}

/// `~/.moai` exists as a real folder: created with the mode of `~/.moai-cli` when this install
/// is the one that decided to create it.
fn ensure_moai(roots: &Roots, m: &Manifest) -> Res<()> {
    match lstat(&roots.moai) {
        Ok(md) if md.is_dir() => Ok(()),
        Ok(_) => Err(format!(
            "home install refused: {} is a link or not a folder; nothing was moved",
            roots.moai.display()
        )),
        Err(e) if not_found(&e) && m.moai_created => make_dir(&roots.moai, m.dir_mode),
        Err(e) if not_found(&e) => Err(format!(
            "home install refused: {} existed when this install began and is gone; nothing was moved",
            roots.moai.display()
        )),
        Err(e) => Err(io(&roots.moai, e)),
    }
}

/// `versions` and `path` hold the running build and the links to it; they move even then.
fn runs_in_place(p: &Path) -> bool {
    p.file_name()
        .and_then(|n| n.to_str())
        .is_some_and(|n| MOVE_LAST.iter().any(|l| n.eq_ignore_ascii_case(l)))
}

fn move_items(c: &Ctx) -> Res<move_journal::Report> {
    let items: Vec<PlannedMove> =
        c.m.items
            .iter()
            .map(|i| PlannedMove {
                from: c.roots.old.join(&i.name),
                to: home_layout::rust_path(c.home, &i.name),
            })
            .collect();
    // The programs that run now decide: an entry one of them uses is skipped, not moved, and a
    // table that cannot be read skips everything.
    let procs = InUse::read();
    let check = |p: &Path| {
        if runs_in_place(p) {
            return move_journal::InUse::No;
        }
        match procs.uses(p, &canon(p)) {
            Ok(None) => move_journal::InUse::No,
            Ok(Some(why)) => move_journal::InUse::Yes(why.to_string()),
            Err(_) => move_journal::InUse::Unknown,
        }
    };
    let knob = fault_knob();
    let mut env = MoveEnv::new(&check);
    env.fault = knob.as_deref();
    move_journal::apply(&c.roots.journal, KIND, &items, &c.m.aliases, c.now, &env).map_err(|e| {
        format!(
            "{e}; the journal in {} lets home-install-revert put back what moved",
            c.roots.journal.display()
        )
    })
}

fn run(c: &Ctx) -> Res<Value> {
    ensure_moai(&c.roots, &c.m)?;
    stop_at("after_mkdir")?;
    let moved = move_items(c)?;
    stop_at("after_commit")?;
    if moved.items.iter().any(|i| i.outcome != "moved") {
        return roll_back(c, &moved);
    }
    let stubs = repoint_stubs(c)?;
    let vault = vault_report(c);
    stop_at("before_marker")?;
    let wrote = write_marker(&c.roots, &c.m)?;
    stop_at("after_marker")?;
    let generation = classify_home(c.home).generation;
    if generation != "rust-first" {
        return Err(format!(
            "the marker is in place but the HOME reads as `{generation}`, not rust-first; run home-install-revert --home {} --stamp {}",
            c.home.display(),
            c.m.stamp
        ));
    }
    Ok(json!({
        "schema_version": 1,
        "state": "applied",
        "resumed": c.resumed,
        "stamp": c.m.stamp,
        "home": c.home.display().to_string(),
        "manifest": c.roots.manifest(&c.m.stamp).display().to_string(),
        "journal": moved.journal,
        "journal_state": moved.state,
        "moved": moved.items.len(),
        "not_moved": 0,
        "items": item_outcomes(&moved),
        "stubs": stubs,
        "vault": vault,
        "marker": {
            "path": marker_path(&c.roots, &c.m)?.display().to_string(),
            "written": wrote,
        },
        "generation": generation,
        "sessions": c.sessions,
        "revert": format!(
            "home-install-revert --home {} --stamp {} --permission-reference <ref>",
            c.home.display(),
            c.m.stamp
        ),
        "partial": false,
    }))
}

fn item_outcomes(r: &move_journal::Report) -> Vec<Value> {
    r.items
        .iter()
        .map(|i| {
            json!({
                "name": Path::new(&i.from).file_name().map(|n| n.to_string_lossy().into_owned()),
                "outcome": i.outcome,
            })
        })
        .collect()
}

/// An entry did not move (a program uses it, or the table of programs cannot be read): put back
/// what did, and remove the `~/.moai` this install created, so the HOME is as it was.
fn roll_back(c: &Ctx, moved: &move_journal::Report) -> Res<Value> {
    let undone = move_journal::revert(&c.roots.move_journal(&c.m.stamp))?;
    let moai = remove_created_moai(&c.roots, &c.m);
    let state = if undone.state == JournalState::Reverted {
        "rolled_back"
    } else {
        "revert_partial"
    };
    Ok(json!({
        "schema_version": 1,
        "state": state,
        "resumed": c.resumed,
        "stamp": c.m.stamp,
        "home": c.home.display().to_string(),
        "manifest": c.roots.manifest(&c.m.stamp).display().to_string(),
        "journal": undone.journal,
        "journal_state": undone.state,
        "moved": 0,
        "not_moved": moved.items.len(),
        "items": item_outcomes(moved),
        "moai": moai,
        "sessions": c.sessions,
        "partial": true,
        "rerun": "an entry could not be moved (a program uses it, or the process table could not be read) and the entries that had moved were put back: stop the program, run home-install-plan again and apply with its new digest",
    }))
}

/// `~/.moai` goes only when this install created it and it is empty (`remove_dir` refuses
/// anything else).
fn remove_created_moai(roots: &Roots, m: &Manifest) -> &'static str {
    if !m.moai_created {
        return "kept_not_created_by_install";
    }
    match lstat(&roots.moai) {
        Ok(md) if md.is_dir() => {}
        Err(_) => return "absent",
        Ok(_) => return "kept_not_a_folder",
    }
    match fs::remove_dir(&roots.moai) {
        Ok(()) => "removed",
        Err(_) => "kept_not_empty",
    }
}

#[derive(Debug, PartialEq, Eq)]
enum Swap {
    Done,
    Already,
    /// Not the link this install knows: left alone.
    Foreign,
}

#[cfg(unix)]
fn make_link(target: &Path, at: &Path) -> Res<()> {
    std::os::unix::fs::symlink(target, at).map_err(|e| io(at, e))
}

#[cfg(not(unix))]
fn make_link(_target: &Path, at: &Path) -> Res<()> {
    Err(format!("{}: links need a unix file system", at.display()))
}

/// Point the link `at` from `from` to `to`: a temporary link renamed over it, so the name is never
/// missing. A link that already points at `to` is left (a rerun), and a link that points anywhere
/// else is not this install's and is left too.
fn swap_link(at: &Path, from: &Path, to: &Path, stamp: &str) -> Res<Swap> {
    let current = match lstat(at) {
        Ok(m) if m.file_type().is_symlink() => fs::read_link(at).map_err(|e| io(at, e))?,
        _ => return Ok(Swap::Foreign),
    };
    if current == to {
        return Ok(Swap::Already);
    }
    if current != from {
        return Ok(Swap::Foreign);
    }
    let name = at.file_name().map(|n| n.to_string_lossy().into_owned());
    let tmp = at.with_file_name(format!(".{}.install-{stamp}.tmp", name.unwrap_or_default()));
    match lstat(&tmp) {
        // A temporary link of a run that stopped here is ours.
        Ok(m) if m.file_type().is_symlink() => fs::remove_file(&tmp).map_err(|e| io(&tmp, e))?,
        Ok(_) => return Err(format!("{} is not a link; refused", tmp.display())),
        Err(_) => {}
    }
    make_link(to, &tmp)?;
    fs::rename(&tmp, at).map_err(|e| io(at, e))?;
    Ok(Swap::Done)
}

fn repoint_stubs(c: &Ctx) -> Res<Value> {
    let dir = c.roots.moai.join("path");
    let old_versions = c.roots.old.join("versions");
    let (mut done, mut already, mut foreign) = (Vec::new(), Vec::new(), Vec::new());
    for (i, s) in c.m.stubs.iter().enumerate() {
        let to = new_target(c.home, &s.target, &old_versions)
            .ok_or_else(|| format!("link {} does not point into the earlier versions", s.name))?;
        let kind = swap_link(&dir.join(&s.name), Path::new(&s.target), &to, &c.m.stamp)?;
        match kind {
            Swap::Done => done.push(s.name.clone()),
            Swap::Already => already.push(s.name.clone()),
            Swap::Foreign => foreign.push(s.name.clone()),
        }
        stop_at(&format!("after_stub:{}", i + 1))?;
    }
    Ok(json!({"repointed": done, "already": already, "left": foreign}))
}

/// The vault by `lstat` alone, before and after: a move keeps its modes, and a loose vault is
/// reported and never changed.
fn vault_report(c: &Ctx) -> Value {
    let after = home_install::vault_facts(&c.roots.moai);
    json!({
        "before": c.m.vault,
        "after": after,
        "unchanged": after == c.m.vault,
        "loose": after["loose"],
    })
}

/// The marker, last: exclusive, with the mode of `~/.moai-cli` minus the execute bits. A marker
/// already there is accepted only when it is the one this install wrote (a rerun).
fn write_marker(roots: &Roots, m: &Manifest) -> Res<bool> {
    let path = marker_path(roots, m)?;
    match lstat(&path) {
        Err(e) if not_found(&e) => {}
        Err(e) => return Err(io(&path, e)),
        Ok(_) => {
            return if marker_is_ours(roots, m)? {
                Ok(false)
            } else {
                Err(format!(
                    "conflict in home install: {} is already there and is not the marker of this install; nothing was changed",
                    path.display()
                ))
            }
        }
    }
    crate::apply::publish_exclusive_mode(
        &path,
        m.marker.content.as_bytes(),
        Some(m.dir_mode & 0o666),
    )?;
    Ok(true)
}

/// Put back what an install moved: the marker goes into the journal folder (only the one the
/// apply wrote), the links go back where they still point at the new target, and the journal is
/// replayed in reverse. `stamp` names the install (the newest when `None`). Deletes nothing.
pub fn revert(home: &Path, stamp: Option<&str>, permission_reference: &str) -> Res<Value> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("permission reference required".into());
    }
    if !home.is_absolute() {
        return Err("HOME is not an absolute path".into());
    }
    let roots = Roots::of(home);
    // The replay renames into `~/.moai`; a link there would send the entries to wherever it
    // points. Absent is left to the replay.
    if lstat(&roots.moai).is_ok_and(|m| !m.is_dir()) {
        return Err(format!(
            "{} is a link or not a folder; nothing was moved",
            roots.moai.display()
        ));
    }
    if !lstat(&roots.journal).is_ok_and(|m| m.is_dir()) {
        return Err(format!(
            "{} is not a folder; nothing was moved",
            roots.journal.display()
        ));
    }
    let stamp = match stamp {
        Some(s) if utc_stamp(s) => s.to_string(),
        Some(s) => return Err(format!("stamp `{s}` is not YYYYMMDDTHHMMSSZ")),
        None => manifest_stamps(&roots.journal)?.pop().ok_or_else(|| {
            format!(
                "install manifest not found in {}; nothing was moved",
                roots.journal.display()
            )
        })?,
    };
    let m = read_manifest(home, &roots, &stamp)?;
    let _lock = take_lock(&roots)?;
    revert_locked(home, &roots, &m)
}

fn revert_locked(home: &Path, roots: &Roots, m: &Manifest) -> Res<Value> {
    let jp = roots.move_journal(&m.stamp);
    if lstat(&jp).is_ok() && move_journal::read_journal(&jp)?.state == JournalState::Reverted {
        return Ok(json!({
            "schema_version": 1,
            "state": "reverted",
            "stamp": m.stamp,
            "home": home.display().to_string(),
            "journal": jp.display().to_string(),
            "changed": false,
            "writes": false,
        }));
    }
    let marker = take_marker_back(roots, m)?;
    let stubs = restore_stubs(home, roots, m)?;
    if lstat(&jp).is_err() {
        // The apply stopped before its journal: an empty one gives this install an end, so a
        // later apply does not take it for unfinished.
        let never_moved = |_: &Path| move_journal::InUse::No;
        move_journal::apply(
            &roots.journal,
            KIND,
            &[],
            &BTreeMap::new(),
            &m.created_at,
            &MoveEnv::new(&never_moved),
        )?;
    }
    let undone = move_journal::revert(&jp)?;
    let moai = remove_created_moai(roots, m);
    let restored_links = stubs["restored"].as_array().is_some_and(|a| !a.is_empty());
    Ok(json!({
        "schema_version": 1,
        "state": undone.state,
        "stamp": m.stamp,
        "home": home.display().to_string(),
        "journal": undone.journal,
        "changed": undone.changed || marker == "moved_into_install_journal" || restored_links,
        "items": item_outcomes(&undone),
        "marker": marker,
        "stubs": stubs,
        "moai": moai,
    }))
}

/// Take the marker out of `~/.moai` into the journal folder, but only the one the apply wrote.
fn take_marker_back(roots: &Roots, m: &Manifest) -> Res<&'static str> {
    let path = marker_path(roots, m)?;
    match lstat(&path) {
        Err(e) if not_found(&e) => return Ok("absent"),
        Err(e) => return Err(io(&path, e)),
        Ok(_) => {}
    }
    if !marker_is_ours(roots, m)? {
        return Err(format!(
            "conflict in home install: the marker {} was changed since the apply (digest mismatch); nothing was changed",
            path.display()
        ));
    }
    let to = roots.reverted_marker(&m.stamp);
    move_journal::rename_noreplace(&path, &to).map_err(|e| io(&to, e))?;
    sync_parent(&to);
    Ok("moved_into_install_journal")
}

fn restore_stubs(home: &Path, roots: &Roots, m: &Manifest) -> Res<Value> {
    let dir = roots.moai.join("path");
    let old_versions = roots.old.join("versions");
    let (mut restored, mut left) = (Vec::new(), Vec::new());
    for s in &m.stubs {
        let now_target = new_target(home, &s.target, &old_versions)
            .ok_or_else(|| format!("link {} does not point into the earlier versions", s.name))?;
        match swap_link(
            &dir.join(&s.name),
            &now_target,
            Path::new(&s.target),
            &m.stamp,
        )? {
            Swap::Done => restored.push(s.name.clone()),
            Swap::Already | Swap::Foreign => left.push(s.name.clone()),
        }
    }
    Ok(json!({"restored": restored, "left": left}))
}

#[cfg(all(test, unix))]
mod tests {
    use super::*;

    fn item(n: &str) -> Value {
        json!({"name": n, "kind": "folder", "action": "move"})
    }

    /// `versions` and `path` move last, in that order, whatever their case; the rest keep the
    /// preview's order.
    #[test]
    fn versions_and_path_move_last() {
        let plan = json!({"entries": [
            item("path"), item("agents"), item("Versions"), item("vault"),
            json!({"name": "bin", "kind": "folder", "action": "keep"}), item("backups"),
        ]});
        let names: Vec<String> = moves_of(&plan).into_iter().map(|i| i.name).collect();
        assert_eq!(names, ["agents", "vault", "backups", "Versions", "path"]);
    }

    #[test]
    fn a_name_is_one_path_component() {
        for ok in ["agents", ".hidden", "a b", "home-state-receipt.json"] {
            assert!(plain_name(ok), "{ok}");
        }
        for bad in ["", ".", "..", "a/b", "../x", "a\\b", "a\0b"] {
            assert!(!plain_name(bad), "{bad:?}");
        }
    }

    #[test]
    fn a_link_target_follows_the_versions_folder() {
        let home = Path::new("/h");
        let old = Path::new("/h/.moai-cli/versions");
        assert_eq!(
            new_target(home, "/h/.moai-cli/versions/1.0.0/bin/moai-cli", old),
            Some(PathBuf::from("/h/.moai/versions/1.0.0/bin/moai-cli"))
        );
        for bad in ["/elsewhere/bin/moai-cli", "/h/.moai-cli/versionsX/a", "rel"] {
            assert_eq!(new_target(home, bad, old), None, "{bad}");
        }
    }

    /// The link swap is the same step both ways: it acts only on the link it knows, leaves one
    /// that is already done, and never touches anything else.
    #[test]
    fn swap_link_acts_only_on_the_link_it_knows() {
        let dir = tempfile::tempdir().unwrap();
        let at = dir.path().join("moai");
        let (a, b, c) = (
            dir.path().join("a"),
            dir.path().join("b"),
            dir.path().join("c"),
        );
        std::os::unix::fs::symlink(&a, &at).unwrap();
        assert_eq!(swap_link(&at, &a, &b, "S").unwrap(), Swap::Done);
        assert_eq!(fs::read_link(&at).unwrap(), b);
        assert_eq!(swap_link(&at, &a, &b, "S").unwrap(), Swap::Already);
        assert_eq!(swap_link(&at, &a, &c, "S").unwrap(), Swap::Foreign);
        assert_eq!(fs::read_link(&at).unwrap(), b);
        fs::remove_file(&at).unwrap();
        assert_eq!(
            swap_link(&at, &a, &b, "S").unwrap(),
            Swap::Foreign,
            "absent"
        );
        fs::write(&at, "x").unwrap();
        assert_eq!(
            swap_link(&at, &a, &b, "S").unwrap(),
            Swap::Foreign,
            "a file"
        );
        assert_eq!(fs::read_to_string(&at).unwrap(), "x");
    }

    /// A temporary link a stopped run left is replaced; a real file under that name is not.
    #[test]
    fn swap_link_replaces_its_own_leftover_and_refuses_a_file() {
        let dir = tempfile::tempdir().unwrap();
        let at = dir.path().join("moai");
        let (a, b) = (dir.path().join("a"), dir.path().join("b"));
        std::os::unix::fs::symlink(&a, &at).unwrap();
        let tmp = dir.path().join(".moai.install-S.tmp");
        std::os::unix::fs::symlink(dir.path().join("stale"), &tmp).unwrap();
        assert_eq!(swap_link(&at, &a, &b, "S").unwrap(), Swap::Done);
        assert!(!tmp.exists() && fs::symlink_metadata(&tmp).is_err());
        fs::remove_file(&at).unwrap();
        std::os::unix::fs::symlink(&a, &at).unwrap();
        fs::write(&tmp, "a real file").unwrap();
        assert!(swap_link(&at, &a, &b, "S").is_err());
        assert_eq!(fs::read_to_string(&tmp).unwrap(), "a real file");
        assert_eq!(fs::read_link(&at).unwrap(), a);
    }
}
