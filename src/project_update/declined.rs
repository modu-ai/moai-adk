//! The choice steps of the project update the user declined (SPEC-20261125-update-declined-steps).
//!
//! The update asks about `claude-md`, `agents-picks`, `agents-local-merge` and
//! `claude-settings-merge` on every run. A user who answers "change nothing" records that answer in
//! `.moai/manifests/update-declined.json`: the step name and the sha256 of the source the step
//! offered. While the step's source has that sha256 and the step would still ask (`needs_decision`
//! or `conflict`), the preview reads it as `skip`. A source with other bytes is a new question,
//! so the record is not read. `--ask-again` removes the record. Declining writes that one file
//! and nothing else: no project file is written or deleted.

use super::{claude_md, Step, PERMISSION_PLACEHOLDER};
use crate::{apply, claude_settings_merge, state_root};
use serde_json::{json, Map, Value};
use std::{collections::BTreeMap, path::Path};

type Result<T> = std::result::Result<T, String>;

/// The subcommand that records and removes a decline.
pub const OPERATION_WORD: &str = "project-update-decline";
pub const AGENTS_PICKS: &str = "agents-picks";
pub const AGENTS_LOCAL_MERGE: &str = "agents-local-merge";
pub const SETTINGS_MERGE: &str = "claude-settings-merge";
pub const CLAUDE_MD: &str = "claude-md";
/// The steps that can be declined.
pub const STEPS: [&str; 4] = [AGENTS_PICKS, AGENTS_LOCAL_MERGE, SETTINGS_MERGE, CLAUDE_MD];

/// The record, under the project's state folder.
const RECORD: &str = "manifests/update-declined.json";
const SCHEMA_VERSION: u64 = 1;

/// Step name to the source sha256 the user declined.
pub type Record = BTreeMap<String, String>;

fn record_rel(root: &Path) -> String {
    state_root::rel(root, RECORD)
}

/// The first component of `rel` below `root` that is a symbolic link, spelled from the root.
fn linked_component(root: &Path, rel: &str) -> Option<String> {
    let mut path = root.to_path_buf();
    let mut shown = String::new();
    for part in rel.split('/') {
        path.push(part);
        if !shown.is_empty() {
            shown.push('/');
        }
        shown.push_str(part);
        if std::fs::symlink_metadata(&path).is_ok_and(|m| m.file_type().is_symlink()) {
            return Some(shown);
        }
    }
    None
}

/// The record file's bytes (`None` when there is none); a link on the way is refused and named.
fn read_bytes(root: &Path) -> Result<Option<Vec<u8>>> {
    let rel = record_rel(root);
    if let Some(link) = linked_component(root, &rel) {
        return Err(format!("{link} is a symbolic link; nothing was written"));
    }
    let path = apply::target(root, &rel).map_err(|e| format!("{rel}: {e}"))?;
    apply::read_opt(&path).map_err(|e| format!("{rel}: {e}"))
}

fn parse(bytes: &[u8]) -> Option<Record> {
    let v: Value = serde_json::from_slice(bytes).ok()?;
    if v["schema_version"].as_u64() != Some(SCHEMA_VERSION) {
        return None;
    }
    v["declined"]
        .as_object()?
        .iter()
        .map(|(step, e)| Some((step.clone(), e["source_sha256"].as_str()?.to_string())))
        .collect()
}

fn encode(record: &Record) -> Vec<u8> {
    let declined: Map<String, Value> = record
        .iter()
        .map(|(step, sha)| (step.clone(), json!({"source_sha256": sha})))
        .collect();
    let mut bytes =
        serde_json::to_vec_pretty(&json!({"schema_version": SCHEMA_VERSION, "declined": declined}))
            .expect("the record encodes");
    bytes.push(b'\n');
    bytes
}

/// The records now (empty without a file). A link, a file that is not a bounded regular file or
/// one that is not a record is an error; the preview then reads every step as it would without.
pub fn read(root: &Path) -> Result<Record> {
    match read_bytes(root)? {
        None => Ok(Record::new()),
        Some(bytes) => parse(&bytes).ok_or_else(|| format!("{}: not a record", record_rel(root))),
    }
}

/// The digest of the bytes of `AGENTS.md` and `AGENTS.local.md` at the project root, each with
/// its name and whether it exists, so a file that appears, changes or goes changes it. `None`
/// when neither exists or one cannot be read as a plain file.
pub fn picks_source_sha256(root: &Path) -> Option<String> {
    let mut bytes = Vec::new();
    let mut any = false;
    for name in ["AGENTS.md", "AGENTS.local.md"] {
        let path = apply::target(root, name).ok()?;
        let body = apply::read_opt(&path).ok()?;
        bytes.extend_from_slice(name.as_bytes());
        bytes.push(0);
        match body {
            Some(body) => {
                any = true;
                bytes.push(1);
                bytes.extend_from_slice(&(body.len() as u64).to_be_bytes());
                bytes.extend_from_slice(&body);
            }
            None => bytes.push(0),
        }
    }
    any.then(|| crate::digest(&bytes))
}

fn unknown_step(step: &str) -> String {
    format!(
        "{step:?} is not a step that can be declined (one of {}); nothing was written",
        STEPS.join(", ")
    )
}

/// The source sha256 `step` has now, `None` when it has no source. Reads only.
pub fn source_sha256(root: &Path, step: &str) -> Result<Option<String>> {
    match step {
        AGENTS_PICKS => Ok(picks_source_sha256(root)),
        AGENTS_LOCAL_MERGE => Ok(crate::agents_local_merge::plan(root)?["source"]["sha256"]
            .as_str()
            .map(String::from)),
        SETTINGS_MERGE => Ok(claude_settings_merge::candidates(root)?.map(|c| c.sha256)),
        CLAUDE_MD => Ok(claude_md::source_digest(root)),
        other => Err(unknown_step(other)),
    }
}

/// `moai project-update-decline <project> --step <step> --source-sha256 <sha256> …`.
pub fn decline_argv(shown: &str, step: &str, sha: &str) -> Vec<String> {
    [
        "moai",
        OPERATION_WORD,
        shown,
        "--step",
        step,
        "--source-sha256",
        sha,
        "--permission-reference",
        PERMISSION_PLACEHOLDER,
    ]
    .map(String::from)
    .to_vec()
}

/// `moai project-update-decline <project> --step <step> --ask-again …`.
pub fn ask_again_argv(shown: &str, step: &str) -> Vec<String> {
    [
        "moai",
        OPERATION_WORD,
        shown,
        "--step",
        step,
        "--ask-again",
        "--permission-reference",
        PERMISSION_PLACEHOLDER,
    ]
    .map(String::from)
    .to_vec()
}

/// What declining the step does, in one sentence.
fn effect(step: &str) -> &'static str {
    match step {
        AGENTS_PICKS => {
            "declining leaves AGENTS.md, AGENTS.local.md and the CLAUDE files as they are and stops asking while those two files keep their bytes"
        }
        AGENTS_LOCAL_MERGE => {
            "declining leaves CLAUDE.local.md as it is and stops asking while the source file keeps its bytes"
        }
        CLAUDE_MD => {
            "declining keeps the old instruction files where they are, still loaded by the sessions, and stops asking while the files the step would move keep their bytes"
        }
        _ => {
            "declining leaves .claude/settings.json as it is and stops asking while the earlier settings copy keeps its bytes"
        }
    }
}

/// The `decline` object of a step's preview: the source sha256, the effect and the commands.
pub fn decline_object(shown: &str, step: &str, sha: &str, declined: bool) -> Value {
    let mut out = json!({
        "source_sha256": sha,
        "effect": effect(step),
        "argv": decline_argv(shown, step, sha),
    });
    if declined {
        out["ask_again_argv"] = json!(ask_again_argv(shown, step));
    }
    out
}

/// Why a declined step reads `skip`.
fn reason(step: &str) -> String {
    format!(
        "the user declined this step for the source it has now; `moai {OPERATION_WORD} --step \
         {step} --ask-again` asks again, and a source with other bytes asks again by itself"
    )
}

/// The step as the preview shows it with the records applied. A step that would ask
/// (`needs_decision` or `conflict`) and has a `source` sha256 prints its `decline` object; one
/// whose record holds that sha256 reads `skip` with a reason that says so, no command and no
/// choices. Any other step is returned as it came.
pub(super) fn read_step(shown: &str, record: &Record, mut step: Step, sha: Option<String>) -> Step {
    let Some(sha) = sha else { return step };
    if !matches!(step.state, "needs_decision" | "conflict") {
        return step;
    }
    let name = step.name;
    let declined = record.get(name) == Some(&sha);
    let Some(preview) = step.preview.as_object_mut() else {
        return step;
    };
    preview.insert(
        "decline".into(),
        decline_object(shown, name, &sha, declined),
    );
    if declined {
        preview.remove("choices");
        preview.insert("declined".into(), json!(true));
        step.state = "skip";
        step.reason = reason(name);
        step.plan_digest = None;
        step.apply = None;
    }
    step
}

fn write(root: &Path, before: Option<&[u8]>, record: &Record) -> Result<String> {
    let rel = record_rel(root);
    let path = apply::target(root, &rel).map_err(|e| format!("{rel}: {e}"))?;
    if let Some(dir) = path.parent() {
        std::fs::create_dir_all(dir).map_err(|_| format!("{rel}: folder create failed"))?;
    }
    apply::publish_replace(&path, before, &encode(record)).map_err(|e| format!("{rel}: {e}"))?;
    Ok(rel)
}

fn checked(project: &Path, step: &str, reference: &str) -> Result<std::path::PathBuf> {
    if !STEPS.contains(&step) {
        return Err(unknown_step(step));
    }
    if !crate::contract::permission_reference_ok(reference) {
        return Err(
            "permission reference refused: it is empty, a placeholder, too long or has a control \
             character; nothing was written"
                .into(),
        );
    }
    std::fs::canonicalize(project).map_err(|_| crate::contract::PROJECT_PATH_UNAVAILABLE.into())
}

/// Records that the user declined `step` for the source `sha`. Refused, with nothing written,
/// when `sha` is not the source sha256 the step has now, when the step has no source, or when
/// the state folder or the record is a link.
pub fn decline(project: &Path, step: &str, sha: &str, reference: &str) -> Result<Value> {
    let root = checked(project, step, reference)?;
    match source_sha256(&root, step)? {
        Some(now) if now == sha => {}
        Some(now) => {
            return Err(format!(
                "--source-sha256 {sha} is not the source sha256 {step} has now ({now}); run \
                 project-update-plan again; nothing was written"
            ))
        }
        None => {
            return Err(format!(
                "{step} has no source now, so there is nothing to decline; nothing was written"
            ))
        }
    }
    let before = read_bytes(&root)?;
    let mut record = match &before {
        Some(bytes) => parse(bytes)
            .ok_or_else(|| format!("{}: not a record; nothing was written", record_rel(&root)))?,
        None => Record::new(),
    };
    let rel = record_rel(&root);
    if record.get(step).map(String::as_str) == Some(sha) {
        return Ok(
            json!({"step": step, "state": "unchanged", "record": rel, "source_sha256": sha, "writes": []}),
        );
    }
    record.insert(step.to_string(), sha.to_string());
    write(&root, before.as_deref(), &record)?;
    Ok(json!({
        "step": step, "state": "declined", "record": rel, "source_sha256": sha,
        "writes": [rel],
        "ask_again": ask_again_argv(root.to_str().unwrap_or("."), step),
    }))
}

/// Removes the record of `step` and leaves the others; a step with no record writes nothing.
pub fn ask_again(project: &Path, step: &str, reference: &str) -> Result<Value> {
    let root = checked(project, step, reference)?;
    let rel = record_rel(&root);
    let before = read_bytes(&root)?;
    let unchanged = json!({"step": step, "state": "unchanged", "record": rel, "writes": []});
    let Some(bytes) = &before else {
        return Ok(unchanged);
    };
    let mut record =
        parse(bytes).ok_or_else(|| format!("{rel}: not a record; nothing was written"))?;
    if record.remove(step).is_none() {
        return Ok(unchanged);
    }
    write(&root, before.as_deref(), &record)?;
    Ok(json!({"step": step, "state": "asked_again", "record": rel, "writes": [rel]}))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn step(state: &'static str, preview: Value) -> Step {
        Step {
            name: AGENTS_LOCAL_MERGE,
            state,
            reason: "asks".into(),
            preview,
            plan_digest: Some("sha256:d".into()),
            apply: Some((vec!["moai".into()], Value::Null)),
        }
    }

    fn record(sha: &str) -> Record {
        Record::from([(AGENTS_LOCAL_MERGE.to_string(), sha.to_string())])
    }

    /// A record of the same sha256 turns a step that asks into `skip`; another sha256, a step
    /// that does not ask and a step without a source stay as they are.
    #[test]
    fn a_record_of_the_same_source_turns_an_asking_step_into_skip() {
        for state in ["needs_decision", "conflict"] {
            let read = read_step(
                "/p",
                &record("sha256:a"),
                step(state, json!({"choices": [1]})),
                Some("sha256:a".into()),
            );
            assert_eq!(read.state, "skip", "{state}");
            assert!(read.apply.is_none() && read.plan_digest.is_none());
            assert!(read.reason.contains("declined") && read.reason.contains("--ask-again"));
            assert_eq!(read.preview.get("choices"), None);
            assert_eq!(read.preview["decline"]["argv"][1], OPERATION_WORD);
            assert_eq!(read.preview["decline"]["ask_again_argv"][5], "--ask-again");
        }
        // Another source: asks as before, with the decline object that carries the new sha256.
        let other = read_step(
            "/p",
            &record("sha256:a"),
            step("needs_decision", json!({})),
            Some("sha256:b".into()),
        );
        assert_eq!(other.state, "needs_decision");
        assert!(other.apply.is_some());
        assert_eq!(other.preview["decline"]["source_sha256"], "sha256:b");
        assert_eq!(other.preview["decline"].get("ask_again_argv"), None);
        // A step that does not ask, and one with no source, are returned as they came.
        for (state, sha) in [("unchanged", Some("sha256:a")), ("needs_decision", None)] {
            let same = read_step(
                "/p",
                &record("sha256:a"),
                step(state, json!({})),
                sha.map(String::from),
            );
            assert_eq!(same.state, state);
            assert_eq!(same.preview, json!({}));
        }
    }

    #[test]
    fn a_record_round_trips_and_a_foreign_file_is_not_a_record() {
        let r = Record::from([
            (AGENTS_PICKS.to_string(), "sha256:1".to_string()),
            (SETTINGS_MERGE.to_string(), "sha256:2".to_string()),
        ]);
        assert_eq!(parse(&encode(&r)), Some(r));
        for bad in [
            "",
            "[]",
            "{\"schema_version\":2,\"declined\":{}}",
            "{\"schema_version\":1,\"declined\":{\"x\":{}}}",
        ] {
            assert_eq!(parse(bad.as_bytes()), None, "{bad}");
        }
    }
}
