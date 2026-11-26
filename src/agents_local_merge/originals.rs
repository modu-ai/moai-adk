//! The pick `remove-originals`: deletes the two originals `AGENTS.md` and `AGENTS.local.md` once
//! their text is whole in the CLAUDE files, and keeps their bytes in the receipt of the pick.
//!
//! - The plan reads each original and counts its text as found only when its CLAUDE file holds
//!   that text whole (`CLAUDE.md` for `AGENTS.md`, `CLAUDE.local.md` for `AGENTS.local.md`). A plan
//!   action `unchanged` of the text move is not "found": an earlier generation merged personal
//!   text into `AGENTS.md`, which no CLAUDE file holds. A CLAUDE file that imports `@AGENTS.md` or
//!   is a link holds no copy, so nothing is offered then.
//! - The apply is all or nothing. It re-reads every original and compares its sha256 with the
//!   plan, writes a private copy of each original into the receipt folder, saves the receipt,
//!   compares again and only then deletes. A failure after the first deletion goes back through the
//!   same receipt.
//! - The revert writes each deleted original back with its bytes and unix mode, leaves a file that
//!   exists again at that path untouched and names it, and takes back the one `.gitignore` line
//!   that keeps the receipt folder out of git.
//!
//! Nothing here deletes a file outside the two original names.

use super::picks::{digest_of, exists, holds_block, text_of, CLAUDE_FILES};
use super::{
    content_of, create_private_dirs, find_block_for, find_guidance, imports_agents_md, plan_ignore,
    read_checked, remove_block, revert_ignore_file, write_private, Content, Found, IgnoreWrite,
    Kind, Result, GITIGNORE, GUIDANCE_TARGET, LEGACY_TARGET, LOCAL, TARGET,
};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{fs, path::Path};

/// The part of the pick receipt that records the deleted originals.
pub(super) const PART: &str = "originals";
/// The folder of the receipt that holds the private copy of each deleted original.
const COPIES: &str = "originals";

/// One original with what the plan found of its text.
pub(super) struct Row {
    rel: &'static str,
    bytes: Vec<u8>,
    sha256: String,
    mode: u32,
    /// `found` (its text is whole in a CLAUDE file), `managed_only` (it holds only MoAI's old
    /// managed section, which `CLAUDE.md` replaces) or `no_text` (blank, or text MoAI generated).
    verdict: &'static str,
    /// The CLAUDE files that hold the text, each with the sha256 it had when the plan read it.
    held_by: Vec<(&'static str, String)>,
}

impl Row {
    fn to_json(&self) -> Value {
        json!({
            "path": self.rel,
            "sha256": self.sha256,
            "bytes": self.bytes.len(),
            "mode": self.mode,
            "verdict": self.verdict,
            "found_in": self.held_by.first().map(|h| h.0),
            "also_found_in": self.held_by.iter().skip(1).map(|h| h.0).collect::<Vec<_>>(),
        })
    }
}

/// What `remove-originals` would do to the project as it is now.
pub(super) struct Plan {
    rows: Vec<Row>,
    /// Text of an original that no CLAUDE file receives (the old managed section of `AGENTS.md`).
    not_copied: Vec<Value>,
    pub(super) warnings: Vec<String>,
    /// The one `.gitignore` line for the receipt folder, when git does not ignore it yet.
    ignore: Option<IgnoreWrite>,
}

pub(super) enum Planned {
    Ready(Plan),
    /// The pick is unavailable, with the reason.
    Refused(String),
}

impl Plan {
    /// The originals the apply deletes, in the order it deletes them.
    pub(super) fn deletes(&self) -> Vec<String> {
        self.rows.iter().map(|r| r.rel.to_string()).collect()
    }

    /// The files the apply writes: only `.gitignore`, and only when its line is missing.
    pub(super) fn writes(&self) -> Vec<String> {
        self.ignore
            .iter()
            .map(|_| GITIGNORE.to_string())
            .collect::<Vec<_>>()
    }

    /// The part of the preview: one verdict per original and the text no CLAUDE file receives.
    pub(super) fn part_json(&self) -> Value {
        json!({
            "part": PART,
            "originals": self.rows.iter().map(Row::to_json).collect::<Vec<_>>(),
            "not_copied": self.not_copied,
        })
    }

    /// The digest the apply is bound to: every original's bytes, mode and verdict, the CLAUDE files
    /// that hold its text and the `.gitignore` write.
    pub(super) fn digest(&self, root: &Path) -> Result<String> {
        digest_of(&json!({
            "pick": "remove-originals",
            "project": root.display().to_string(),
            "originals": self.rows.iter().map(|r| json!({
                "path": r.rel, "sha256": r.sha256, "mode": r.mode, "verdict": r.verdict,
                "held_by": r.held_by,
            })).collect::<Vec<_>>(),
            "gitignore": self.ignore.as_ref().map(|g| json!({
                "before": g.before.as_deref().map(crate::digest),
                "after": crate::digest(&g.new_bytes),
            })),
        }))
    }
}

#[cfg(unix)]
fn mode_of(meta: &fs::Metadata) -> u32 {
    use std::os::unix::fs::PermissionsExt;
    meta.permissions().mode() & 0o7777
}

#[cfg(not(unix))]
fn mode_of(_: &fs::Metadata) -> u32 {
    0o644
}

/// An original that is a regular UTF-8 file of at most 128 KiB, with its mode. `None`: it is not
/// there. A link, a file that is not text and a file above the limit are refused with the reason.
fn read_original(root: &Path, rel: &'static str) -> Result<Option<(Vec<u8>, u32)>> {
    let Some(bytes) = read_checked(root, rel)? else {
        return Ok(None);
    };
    if std::str::from_utf8(&bytes).is_err() {
        return Err(format!("{rel} is not UTF-8 text; nothing is deleted"));
    }
    let meta =
        fs::symlink_metadata(root.join(rel)).map_err(|_| format!("{rel}: metadata unavailable"))?;
    Ok(Some((bytes, mode_of(&meta))))
}

/// A CLAUDE file as text with the sha256 of its bytes; `None` when it is absent or not text.
fn claude_text(root: &Path, rel: &str) -> Option<(String, String)> {
    let bytes = read_checked(root, rel).ok()??;
    let sha = crate::digest(&bytes);
    String::from_utf8(bytes).ok().map(|t| (t, sha))
}

/// `remove-originals` for the project at `root`, or the reason it is not available. Writes nothing.
pub(super) fn plan(root: &Path) -> Planned {
    match plan_inner(root) {
        Ok(plan) => Planned::Ready(plan),
        Err(reason) => Planned::Refused(reason),
    }
}

fn plan_inner(root: &Path) -> Result<Plan> {
    if !exists(root, LEGACY_TARGET) && !exists(root, LOCAL) {
        return Err("no AGENTS.md and no AGENTS.local.md in this project".into());
    }
    // A CLAUDE file that imports `@AGENTS.md` or links to it holds no copy of the text.
    for rel in CLAUDE_FILES {
        crate::apply::target(root, rel).map_err(|e| format!("{rel}: {e}"))?;
        if text_of(root, rel).is_some_and(|t| imports_agents_md(&t)) {
            return Err(format!(
                "{rel} imports AGENTS.md, so no CLAUDE file holds a copy of its text"
            ));
        }
    }
    let mut rows = Vec::new();
    let mut not_copied = Vec::new();
    let mut warnings = Vec::new();
    if let Some((bytes, mode)) = read_original(root, LEGACY_TARGET)? {
        let text = String::from_utf8(bytes.clone()).map_err(|_| "AGENTS.md is not UTF-8 text")?;
        let mut held_by = Vec::new();
        let verdict = match find_guidance(root)? {
            Found::Blocked(reason) => return Err(reason),
            Found::Source(s) => {
                let (_, sha) = claude_text(root, GUIDANCE_TARGET)
                    .filter(|(t, _)| holds_block(t, &s.text))
                    .ok_or("the text of AGENTS.md is not whole in CLAUDE.md")?;
                held_by.push((GUIDANCE_TARGET, sha));
                "found"
            }
            // Only MoAI's old managed section: the managed block of `CLAUDE.md` replaces it.
            Found::Nothing(_) => {
                let (begin, _) = crate::agents::markers(crate::agents::AGENTS_SECTION, "html");
                let (_, sha) = claude_text(root, GUIDANCE_TARGET)
                    .filter(|(t, _)| t.contains(&begin))
                    .ok_or(
                        "AGENTS.md holds only MoAI's old managed section and CLAUDE.md does not \
                         hold the managed section yet",
                    )?;
                held_by.push((GUIDANCE_TARGET, sha));
                "managed_only"
            }
        };
        // Personal text an earlier generation merged into `AGENTS.md` must be whole in
        // `CLAUDE.local.md` too: the plan of the text move reads `unchanged` for it.
        if let Some(existing) = find_block_for(Kind::Local, &text)? {
            let (_, sha) = claude_text(root, TARGET)
                .filter(|(t, _)| holds_block(t, existing.body.as_bytes()))
                .ok_or(
                    "the personal text an earlier generation merged into AGENTS.md is not whole \
                     in CLAUDE.local.md",
                )?;
            held_by.push((TARGET, sha));
        }
        let (begin, end) = crate::agents::markers(crate::agents::AGENTS_SECTION, "html");
        if let Some(removed) = remove_block(&text, &begin, &end) {
            let id = crate::agents::AGENTS_SECTION;
            let lines = removed.lines;
            not_copied.push(json!({"path": LEGACY_TARGET, "section": id, "lines": lines}));
            warnings.push(format!(
                "{LEGACY_TARGET} holds MoAI's old managed section (marker id {id}, {lines} line{}): \
                 no CLAUDE file receives that text, so text you wrote inside it survives only in \
                 the receipt copy of {LEGACY_TARGET}",
                if lines == 1 { "" } else { "s" }
            ));
        }
        rows.push(row(LEGACY_TARGET, bytes, mode, verdict, held_by));
    }
    if let Some((bytes, mode)) = read_original(root, LOCAL)? {
        let (verdict, held_by) = match content_of(&bytes) {
            Content::Blank | Content::Generated => ("no_text", Vec::new()),
            Content::Personal => {
                let (_, sha) = claude_text(root, TARGET)
                    .filter(|(t, _)| holds_block(t, &bytes))
                    .ok_or("the text of AGENTS.local.md is not whole in CLAUDE.local.md")?;
                ("found", vec![(TARGET, sha)])
            }
        };
        rows.push(row(LOCAL, bytes, mode, verdict, held_by));
    }
    let facts = plan_ignore(root, false);
    warnings.extend(facts.warnings);
    Ok(Plan {
        rows,
        not_copied,
        warnings,
        ignore: facts.write,
    })
}

fn row(
    rel: &'static str,
    bytes: Vec<u8>,
    mode: u32,
    verdict: &'static str,
    held_by: Vec<(&'static str, String)>,
) -> Row {
    Row {
        rel,
        sha256: crate::digest(&bytes),
        bytes,
        mode,
        verdict,
        held_by,
    }
}

/// What the receipt keeps of one deleted original.
#[derive(Serialize, Deserialize, Clone)]
pub(super) struct FileRecord {
    path: String,
    sha256: String,
    bytes: u64,
    /// The unix permission bits the original had.
    mode: u32,
}

/// What the receipt keeps of the `.gitignore` line the apply added.
#[derive(Serialize, Deserialize, Clone)]
pub(super) struct IgnoreRecord {
    existed: bool,
    before: Option<String>,
    after: String,
}

/// Compares every original with the plan, all before any is touched: a link, a file that is not
/// text, a file above the limit, a missing file or bytes that differ from the plan refuse the
/// whole set.
pub(super) fn verify(root: &Path, rows: &[Row]) -> Result<()> {
    for row in rows {
        match read_original(root, row.rel)? {
            Some((bytes, _)) if crate::digest(&bytes) == row.sha256 => {}
            _ => {
                return Err(format!(
                    "{} changed after the preview; nothing was deleted",
                    row.rel
                ))
            }
        }
    }
    Ok(())
}

/// The first half of the apply: compares the originals with the plan, writes a private copy of
/// each into the receipt folder `dir` (owner-only, never replacing a file) and returns what the
/// receipt records. No original and no project file changes.
pub(super) fn stage(
    root: &Path,
    dir: &Path,
    plan: &Plan,
) -> Result<(Vec<FileRecord>, Option<IgnoreRecord>)> {
    verify(root, &plan.rows)?;
    let copies = create_private_dirs(dir, &[COPIES])?;
    for row in &plan.rows {
        write_private(&copies.join(row.rel), &row.bytes)?;
    }
    if let Some(Some(before)) = plan.ignore.as_ref().map(|g| g.before.as_ref()) {
        let keep = create_private_dirs(dir, &["before"])?;
        write_private(&keep.join(GITIGNORE), before)?;
    }
    let files = plan
        .rows
        .iter()
        .map(|r| FileRecord {
            path: r.rel.to_string(),
            sha256: r.sha256.clone(),
            bytes: r.bytes.len() as u64,
            mode: r.mode,
        })
        .collect();
    let ignore = plan.ignore.as_ref().map(|g| IgnoreRecord {
        existed: g.before.is_some(),
        before: g.before.as_deref().map(crate::digest),
        after: crate::digest(&g.new_bytes),
    });
    Ok((files, ignore))
}

/// The second half of the apply, after the receipt names the files: the `.gitignore` line first
/// (when it cannot be written nothing else has changed), then the originals. Returns the files
/// written.
pub(super) fn commit(
    root: &Path,
    plan: &Plan,
    after_each: &mut dyn FnMut(usize) -> Result<()>,
) -> Result<Vec<String>> {
    let mut writes = Vec::new();
    if let Some(g) = &plan.ignore {
        let path = crate::apply::target(root, GITIGNORE)?;
        crate::apply::publish_replace(&path, g.before.as_deref(), &g.new_bytes)
            .map_err(|e| format!("{GITIGNORE} was not written ({e})"))?;
        writes.push(GITIGNORE.to_string());
    }
    delete(root, &plan.rows, after_each)?;
    Ok(writes)
}

/// Deletes the originals, all or none: every one is compared with the plan before the first is
/// removed. `after_each` runs after each removal with its index (a failure there stops the loop).
pub(super) fn delete(
    root: &Path,
    rows: &[Row],
    after_each: &mut dyn FnMut(usize) -> Result<()>,
) -> Result<()> {
    verify(root, rows)?;
    for (i, row) in rows.iter().enumerate() {
        let path = crate::apply::target(root, row.rel)?;
        fs::remove_file(&path).map_err(|_| format!("{} could not be deleted", row.rel))?;
        after_each(i)?;
    }
    Ok(())
}

/// Writes one deleted original back from the receipt copy with its recorded mode. `Ok(true)`: it
/// was written. `Ok(false)`: the file is there with the recorded bytes already. `Err`: a file is
/// at the path (it stays as it is), or the receipt cannot give the bytes back.
fn restore_one(root: &Path, dir: &Path, f: &FileRecord) -> Result<bool> {
    if f.path != LEGACY_TARGET && f.path != LOCAL {
        return Err("the receipt names a file this command never writes".into());
    }
    let path = crate::apply::target(root, &f.path)?;
    if fs::symlink_metadata(&path).is_ok() {
        return match crate::apply::read_opt(&path) {
            Ok(Some(now)) if crate::digest(&now) == f.sha256 => Ok(false),
            _ => Err("a file exists again at this path; it was left as it is".into()),
        };
    }
    let copies = dir.join(COPIES);
    crate::claude::refuse_symlink_dir(&copies)?;
    let copy = crate::claude::read_existing(&copies.join(&f.path))?
        .ok_or("the receipt has no copy of this file")?;
    if crate::digest(&copy) != f.sha256 {
        return Err(
            "the receipt's copy of this file was altered; nothing was written for it".into(),
        );
    }
    crate::apply::publish_exclusive_mode(&path, &copy, Some(f.mode))?;
    Ok(true)
}

/// Undoes the part: each deleted original goes back in the reverse order of the deletion, then the
/// `.gitignore` line goes. Returns the files written and what was kept, each naming its file.
pub(super) fn revert(
    root: &Path,
    dir: &Path,
    receipt_id: &str,
    files: &[FileRecord],
    ignore: Option<&IgnoreRecord>,
) -> (Vec<String>, Vec<Value>) {
    let mut writes = Vec::new();
    let mut kept = Vec::new();
    for f in files.iter().rev() {
        match restore_one(root, dir, f) {
            Ok(true) => writes.push(f.path.clone()),
            Ok(false) => {}
            Err(reason) => kept.push(json!({
                "part": PART, "path": f.path, "receipt": receipt_id, "reason": reason,
            })),
        }
    }
    if let Some(g) = ignore {
        // A `.gitignore` that changed since is left as it is without a word, like the receipts of
        // the text merges; only a receipt that cannot be read is named.
        if let Err(reason) = revert_ignore_file(
            root,
            dir,
            GITIGNORE,
            g.existed,
            g.before.as_deref(),
            Some(&g.after),
        ) {
            kept.push(json!({
                "part": PART, "path": GITIGNORE, "receipt": receipt_id, "reason": reason,
            }));
        }
    }
    (writes, kept)
}

#[cfg(test)]
mod tests {
    use super::*;

    const AGENTS: &str = "# Team rules\n\nWrite tests first.\n";
    const MINE: &str = "My own notes\nPrefer tabs.\n";

    /// A project whose CLAUDE files already hold the text of both originals.
    fn project() -> (tempfile::TempDir, std::path::PathBuf) {
        let tmp = tempfile::TempDir::new().unwrap();
        let root = fs::canonicalize(tmp.path()).unwrap();
        for (rel, text) in [
            ("AGENTS.md", AGENTS),
            ("CLAUDE.md", AGENTS),
            ("AGENTS.local.md", MINE),
            ("CLAUDE.local.md", MINE),
        ] {
            fs::write(root.join(rel), text).unwrap();
        }
        (tmp, root)
    }

    fn ready(root: &Path) -> Plan {
        match plan(root) {
            Planned::Ready(plan) => plan,
            Planned::Refused(reason) => panic!("refused: {reason}"),
        }
    }

    fn noop() -> impl FnMut(usize) -> Result<()> {
        |_| Ok(())
    }

    #[test]
    fn delete_removes_both_originals_and_nothing_else() {
        let (_tmp, root) = project();
        let plan = ready(&root);
        delete(&root, &plan.rows, &mut noop()).unwrap();
        assert!(!root.join("AGENTS.md").exists() && !root.join("AGENTS.local.md").exists());
        assert!(root.join("CLAUDE.md").exists() && root.join("CLAUDE.local.md").exists());
    }

    #[test]
    fn delete_keeps_a_file_that_changed_after_the_plan() {
        let (_tmp, root) = project();
        let plan = ready(&root);
        fs::write(
            root.join("AGENTS.local.md"),
            format!("{MINE}one more line\n"),
        )
        .unwrap();
        let err = delete(&root, &plan.rows, &mut noop()).unwrap_err();
        assert!(
            err.contains("AGENTS.local.md changed after the preview"),
            "{err}"
        );
        assert!(
            root.join("AGENTS.md").exists(),
            "the unchanged original was deleted"
        );
        assert!(fs::read_to_string(root.join("AGENTS.local.md"))
            .unwrap()
            .contains("one more line"));
    }

    #[test]
    fn delete_is_all_or_nothing_when_the_second_original_changed() {
        let (_tmp, root) = project();
        let plan = ready(&root);
        // The first row is untouched and comes first; the second one changed.
        fs::write(root.join("AGENTS.local.md"), "other\n").unwrap();
        assert!(delete(&root, &plan.rows, &mut noop()).is_err());
        assert_eq!(fs::read_to_string(root.join("AGENTS.md")).unwrap(), AGENTS);
        assert!(root.join("AGENTS.local.md").exists());
    }

    #[cfg(unix)]
    #[test]
    fn delete_refuses_a_link_in_place_of_an_original() {
        let (_tmp, root) = project();
        let plan = ready(&root);
        fs::rename(root.join("AGENTS.md"), root.join("elsewhere.md")).unwrap();
        std::os::unix::fs::symlink("elsewhere.md", root.join("AGENTS.md")).unwrap();
        assert!(delete(&root, &plan.rows, &mut noop()).is_err());
        assert!(root.join("AGENTS.local.md").exists());
        assert!(root.join("elsewhere.md").exists());
    }

    #[test]
    fn the_plan_needs_the_whole_text_of_each_original_in_its_own_claude_file() {
        // The personal text: absent, partial or unrelated in CLAUDE.local.md.
        for claude_local in [None, Some("My own notes\n"), Some("unrelated\n")] {
            let (_tmp, root) = project();
            match claude_local {
                Some(text) => fs::write(root.join("CLAUDE.local.md"), text).unwrap(),
                None => fs::remove_file(root.join("CLAUDE.local.md")).unwrap(),
            }
            match plan(&root) {
                Planned::Refused(reason) => {
                    assert!(
                        reason.contains("AGENTS.local.md"),
                        "{claude_local:?}: {reason}"
                    )
                }
                Planned::Ready(_) => panic!("{claude_local:?}: found without its whole text"),
            }
        }
        // The team text: partial in CLAUDE.md.
        let (_tmp, root) = project();
        fs::write(root.join("CLAUDE.md"), "# Team rules\n").unwrap();
        match plan(&root) {
            Planned::Refused(reason) => assert!(reason.contains("AGENTS.md"), "{reason}"),
            Planned::Ready(_) => panic!("found without its whole text"),
        }
    }

    #[test]
    fn the_plan_counts_a_merged_local_block_as_not_found() {
        let (_tmp, root) = project();
        fs::remove_file(root.join("AGENTS.local.md")).unwrap();
        fs::remove_file(root.join("CLAUDE.local.md")).unwrap();
        let body = MINE;
        let sha = crate::digest(body.as_bytes());
        let block = format!(
            "{}{} -->\n{body}{}\n",
            super::super::BEGIN,
            sha.trim_start_matches("sha256:"),
            super::super::END
        );
        fs::write(root.join("AGENTS.md"), format!("{AGENTS}\n{block}")).unwrap();
        match plan(&root) {
            Planned::Refused(reason) => assert!(reason.contains("CLAUDE.local.md"), "{reason}"),
            Planned::Ready(_) => panic!("a text that sits only in AGENTS.md was found"),
        }
        fs::write(root.join("CLAUDE.local.md"), MINE).unwrap();
        assert!(matches!(plan(&root), Planned::Ready(_)));
    }

    #[test]
    fn a_stale_original_does_not_stage_a_copy() {
        let (_tmp, root) = project();
        let plan = ready(&root);
        let dir = root.join("receipt");
        fs::create_dir(&dir).unwrap();
        fs::write(root.join("AGENTS.md"), "changed\n").unwrap();
        assert!(stage(&root, &dir, &plan).is_err());
        assert!(!dir.join(COPIES).exists());
    }
}
