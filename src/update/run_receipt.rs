//! The run receipt of one confirmed `moai update` (card C-513, S7 of the update redesign).
//!
//! `update_command_core` writes the skill copies first and reports a later part's failure beside
//! them, so a run can end half applied. This module reads the finished result once and says, in
//! one record, which parts landed, which are left and why, and whether the skill files the run
//! wrote still hold the bytes it meant to write: each one is read back from disk and its sha256 is
//! compared with the `after_sha256` the result names. The end screen is drawn from this record, not
//! from the parts one by one.
//!
//! Only the files of the moai skill and of the skill folders beside it carry an expected digest in
//! the result. Every other part of a run is listed in `not_compared` with a reason and its own
//! receipt, whatever its state: it either writes files the result gives no digest for, or only
//! moves and links. A skill part is listed too when a file it planned to change was not compared.
//! The record claims no comparison it did not make.
//!
//! The record is one file in the HOME state folder, `installations/<install id>/update-run.json`,
//! replaced atomically by every run that wrote something; a confirmed plan that writes nothing
//! leaves the folder as it was, its `run` still sits in the result. A write failure is
//! reported as `receipt_error` and never replaces the apply result.

use super::{current_install_id, now_rfc3339, projection_root, sha256_hex, skills_parent};
use crate::apply::{publish_replace, read_opt};
use serde_json::{json, Value};
use std::{
    fs,
    path::{Component, Path, PathBuf},
};

/// The file name of the record inside `installations/<install id>/`.
pub const FILE_NAME: &str = "update-run.json";
const SCHEMA_VERSION: u32 = 1;
/// A written file larger than this is not read back; it is reported, never skipped silently.
const READ_LIMIT: u64 = 16 * 1024 * 1024;

/// The reason `update_command_core` gives a part it did not run because the skill copy of its
/// host did not apply. The record reads it back as an open part.
pub(crate) const GATE_REASON: &str = "the skill copy did not apply";

/// How a part of the run ended.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum End {
    /// Applied, or nothing to apply.
    Done,
    /// Left as it was on purpose: a file in conflict, or a file of the person's that was kept.
    /// A rerun does not change it; the person chooses.
    Held,
    /// Not applied: failed, refused, blocked or not run. A rerun applies it once its cause is gone.
    Open,
}

struct Part {
    part: &'static str,
    host: Option<String>,
    skill: Option<String>,
    state: String,
    end: End,
    reason: Option<String>,
}

impl Part {
    fn json(&self) -> Value {
        let mut v = json!({"part": self.part, "state": self.state});
        if let Some(h) = &self.host {
            v["host"] = json!(h);
        }
        if let Some(s) = &self.skill {
            v["skill"] = json!(s);
        }
        if let Some(r) = &self.reason {
            v["reason"] = json!(r);
        }
        v
    }
}

fn text(v: &Value) -> Option<String> {
    v.as_str().filter(|s| !s.is_empty()).map(str::to_string)
}

fn conflicts(files: &Value) -> usize {
    files
        .as_array()
        .into_iter()
        .flatten()
        .filter(|f| f["action"].as_str() == Some("conflict"))
        .count()
}

/// The parts of the moai skill result, one per host.
fn skill_parts(r: &Value) -> Vec<Part> {
    let mut out = Vec::new();
    for h in r["hosts"].as_array().into_iter().flatten() {
        let status = h["status"].as_str().unwrap_or("").to_string();
        let held = conflicts(&h["files"]);
        let (end, reason) = match status.as_str() {
            "applied" | "replaced" | "unchanged" if held == 0 => (End::Done, None),
            "applied" | "replaced" | "unchanged" | "conflict" => {
                (End::Held, Some(format!("{held} in conflict")))
            }
            _ => (
                End::Open,
                text(&h["blocked_reason"])
                    .or_else(|| {
                        h["files"]
                            .as_array()
                            .into_iter()
                            .flatten()
                            .find_map(|f| text(&f["error"]))
                    })
                    .or_else(|| text(&h["status"])),
            ),
        };
        out.push(Part {
            part: "skills",
            host: text(&h["host"]),
            skill: None,
            state: status,
            end,
            reason,
        });
    }
    out
}

/// The parts of the skill folders beside the moai skill. A result without `parts` is a refused or
/// failed step and counts as one open part.
fn extra_parts(r: &Value) -> Vec<Part> {
    if r.is_null() {
        return Vec::new();
    }
    let Some(parts) = r["parts"].as_array() else {
        return vec![Part {
            part: "extra_skills",
            host: None,
            skill: None,
            state: r["state"].as_str().unwrap_or("").to_string(),
            end: End::Open,
            reason: text(&r["reason"]).or_else(|| text(&r["message"])),
        }];
    };
    parts
        .iter()
        .map(|p| {
            let status = p["status"].as_str().unwrap_or("").to_string();
            let held = conflicts(&p["files"]);
            let (end, reason) = match status.as_str() {
                "applied" | "unchanged" if held == 0 => (End::Done, None),
                "applied" | "unchanged" | "conflict" => {
                    (End::Held, Some(format!("{held} in conflict")))
                }
                _ => (
                    End::Open,
                    text(&p["blocked_reason"])
                        .or_else(|| text(&p["reason"]))
                        .or_else(|| text(&p["message"]))
                        .or_else(|| text(&p["status"])),
                ),
            };
            Part {
                part: "extra_skills",
                host: text(&p["host"]),
                skill: text(&p["skill"]),
                state: status,
                end,
                reason,
            }
        })
        .collect()
}

/// A part that reports `state` and `written`; `done` says which states end it cleanly.
fn state_part(part: &'static str, r: &Value, done: impl Fn(&Value) -> bool) -> Option<Part> {
    if r.is_null() {
        return None;
    }
    let state = r["state"].as_str().unwrap_or("").to_string();
    let kept = r["kept"].as_array().is_some_and(|k| !k.is_empty());
    let (end, reason) = if done(r) {
        (End::Done, None)
    } else if part == "agents" && state == "partial" && kept {
        // A flat agent file that is not ours was kept on purpose.
        (End::Held, Some("a file of yours was kept".to_string()))
    } else if state == "outside_layout" {
        // The executable is not in the layout the part links from: nothing to write here, and a
        // rerun finds the same.
        (End::Held, text(&r["reason"]).or_else(|| text(&r["state"])))
    } else {
        (
            End::Open,
            text(&r["message"])
                .or_else(|| text(&r["reason"]))
                .or_else(|| text(&r["state"])),
        )
    };
    Some(Part {
        part,
        host: None,
        skill: None,
        state,
        end,
        reason,
    })
}

fn applied_or_unchanged(r: &Value) -> bool {
    matches!(
        r["state"].as_str(),
        Some("applied" | "unchanged" | "skipped")
    )
}

/// Every part of a finished run, in the order the run applies them.
fn parts_of(out: &Value) -> Vec<Part> {
    let mut parts = skill_parts(&out["result"]);
    parts.extend(extra_parts(&out["extra_skills_result"]));
    for (name, key) in [
        ("agents", "agents_result"),
        ("routing", "routing_result"),
        ("style", "style_result"),
        ("bin_names", "bin_names_result"),
        ("hooks", "hooks_result"),
        ("keys", "keys_result"),
    ] {
        parts.extend(state_part(name, &out[key], applied_or_unchanged));
    }
    // The parts with their own token end the way the end screen reads them: a cleanup is done
    // when it committed, the home part when it applied and its journal committed, the switch and
    // the git-hook part when they applied.
    parts.extend(state_part("cleanup", &out["cleanup_result"], |r| {
        r["state"].as_str() == Some("committed")
    }));
    parts.extend(state_part("home", &out["home_result"], |r| {
        r["state"].as_str() == Some("applied") && r["journal_state"] == "committed"
    }));
    parts.extend(state_part("switch", &out["switch_result"], |r| {
        r["state"].as_str() == Some("applied")
    }));
    parts.extend(state_part("git_hooks", &out["git_hooks_result"], |r| {
        r["state"].as_str() == Some("applied")
    }));
    // A part the run did not start because the skill copy of its host did not apply.
    for s in out["skipped"].as_array().into_iter().flatten() {
        if s["reason"].as_str() != Some(GATE_REASON) {
            continue;
        }
        let part = match s["part"].as_str() {
            Some("agents") => "agents",
            Some("routing") => "routing",
            Some("style") => "style",
            _ => continue,
        };
        parts.push(Part {
            part,
            host: text(&s["host"]),
            skill: None,
            state: "not_run".to_string(),
            end: End::Open,
            reason: text(&s["reason"]),
        });
    }
    parts
}

/// The folder a relative path of a result lives in must stay inside it: a result path with a
/// parent or root component is not read.
fn inside(root: &Path, rel: &str) -> Option<PathBuf> {
    let rel = Path::new(rel);
    rel.components()
        .all(|c| matches!(c, Component::Normal(_)))
        .then(|| root.join(rel))
}

fn bare(sha: &str) -> &str {
    sha.strip_prefix("sha256:").unwrap_or(sha)
}

/// Reads one written file back and compares it with the digest the result named.
fn check_one(part: &str, host: &str, rel: &str, root: &Path, expected: &str) -> Value {
    let mut v = json!({"part": part, "host": host, "path": rel, "expected_sha256": expected});
    let check = match inside(root, rel) {
        None => "unsafe_path".to_string(),
        Some(path) => match fs::symlink_metadata(&path) {
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => "missing".to_string(),
            Err(_) => "unreadable".to_string(),
            Ok(m) if !m.is_file() => "not_a_file".to_string(),
            Ok(m) if m.len() > READ_LIMIT => "too_large".to_string(),
            Ok(_) => match fs::read(&path) {
                Err(_) => "unreadable".to_string(),
                Ok(bytes) => {
                    let seen = sha256_hex(&bytes);
                    v["observed_sha256"] = json!(seen);
                    if bare(&seen) == bare(expected) {
                        "match".to_string()
                    } else {
                        "mismatch".to_string()
                    }
                }
            },
        },
    };
    v["check"] = json!(check);
    v
}

/// The written files of a result's file list that name an expected digest.
fn written(files: &Value) -> impl Iterator<Item = (&str, &str)> {
    files.as_array().into_iter().flatten().filter_map(|f| {
        if f["published"].as_bool() != Some(true) {
            return None;
        }
        Some((f["path"].as_str()?, f["after_sha256"].as_str()?))
    })
}

/// Reads back every skill file this run wrote: the moai skill and the folders beside it.
fn check_files(home: &Path, out: &Value) -> Vec<Value> {
    let mut checks = Vec::new();
    for h in out["result"]["hosts"].as_array().into_iter().flatten() {
        let host = h["host"].as_str().unwrap_or("");
        let Ok(root) = projection_root(home, host) else {
            continue;
        };
        for (rel, expected) in written(&h["files"]) {
            checks.push(check_one("skills", host, rel, &root, expected));
        }
    }
    for p in out["extra_skills_result"]["parts"]
        .as_array()
        .into_iter()
        .flatten()
    {
        let (Some(host), Some(skill)) = (p["host"].as_str(), p["skill"].as_str()) else {
            continue;
        };
        let Some(root) = skills_parent(home, host)
            .ok()
            .and_then(|parent| inside(&parent, skill))
        else {
            continue;
        };
        for (rel, expected) in written(&p["files"]) {
            checks.push(check_one("extra_skills", host, rel, &root, expected));
        }
    }
    checks
}

/// The files a skill part set out to change, published or not: a file left in conflict or kept
/// for the person counts.
fn planned_files(out: &Value, part: &str) -> usize {
    let entries = if part == "skills" {
        &out["result"]["hosts"]
    } else {
        &out["extra_skills_result"]["parts"]
    };
    entries
        .as_array()
        .into_iter()
        .flatten()
        .flat_map(|e| e["files"].as_array().into_iter().flatten())
        .filter(|f| {
            f["published"].as_bool() == Some(true) || f["action"].as_str() != Some("unchanged")
        })
        .count()
}

/// Why the record holds no comparison for a part outside the skill files.
fn why_not_compared(part: &str) -> &'static str {
    match part {
        "bin_names" => "no digest to compare: it only links bin/moai and moves a copy aside",
        "cleanup" => "no digest to compare: it only moves files into an archive",
        "home" => "no digest to compare: it only moves entries into a dated folder",
        "switch" => {
            "no digest to compare: it puts a link in place of the old script, and this record \
             does not read the kept copy back"
        }
        _ => "this record does not read back the files of this part",
    }
}

/// Every part whose planned files this record did not all compare, whatever its state, with the
/// reason and the receipt the part keeps: each part outside the skill files, and a skill part
/// that did not apply or has a planned file without a digest comparison.
fn not_compared(out: &Value, parts: &[Part], checks: &[Value]) -> Vec<Value> {
    let mut list: Vec<Value> = Vec::new();
    for p in parts {
        if list.iter().any(|v| v["part"] == p.part) {
            continue;
        }
        let same = || parts.iter().filter(|q| q.part == p.part);
        let mut state = &p.state;
        let mut v = json!({"part": p.part});
        if matches!(p.part, "skills" | "extra_skills") {
            let planned = planned_files(out, p.part);
            let compared = checks
                .iter()
                .filter(|c| {
                    c["part"] == p.part && matches!(c["check"].as_str(), Some("match" | "mismatch"))
                })
                .count();
            if compared < planned {
                v["reason"] = json!(format!(
                    "{} of {planned} planned files were not compared",
                    planned - compared
                ));
            } else if same().any(|q| q.end == End::Open) {
                v["reason"] = json!("the part did not apply");
            } else {
                continue;
            }
            if let Some(q) = same().find(|q| q.end != End::Done) {
                state = &q.state;
            }
        } else {
            v["reason"] = json!(why_not_compared(p.part));
            let result = &out[format!("{}_result", p.part)];
            if let Some(r) = text(&result["receipt"]).or_else(|| text(&result["stamp"])) {
                v["receipt"] = json!(r);
            }
        }
        v["state"] = json!(state);
        list.push(v);
    }
    list
}

/// The record of a finished run, built from its result. Reads the written skill files; writes
/// nothing.
pub fn build(home: &Path, out: &Value) -> Value {
    let parts = parts_of(out);
    let checks = check_files(home, out);
    let bad_files = checks
        .iter()
        .filter(|c| c["check"].as_str() != Some("match"))
        .count();
    let mut remaining: Vec<&'static str> = Vec::new();
    for p in parts.iter().filter(|p| p.end == End::Open) {
        if !remaining.contains(&p.part) {
            remaining.push(p.part);
        }
    }
    let mut held: Vec<&'static str> = Vec::new();
    for p in parts.iter().filter(|p| p.end == End::Held) {
        if !held.contains(&p.part) {
            held.push(p.part);
        }
    }
    let complete = remaining.is_empty() && held.is_empty() && bad_files == 0;
    let not_compared = not_compared(out, &parts, &checks);
    let mut run = json!({
        "schema_version": SCHEMA_VERSION,
        "status": if complete { "complete" } else { "partial" },
        "at": now_rfc3339(),
        "plan_digest": out["plan_digest"],
        "package_version": out["result"]["package_version"],
        "wrote": out["applied"] == json!(true),
        "parts": parts.iter().map(Part::json).collect::<Vec<_>>(),
        "remaining": remaining,
        "held": held,
        "files_checked": checks.len(),
        "files_not_matching": bad_files,
        "files": checks,
        "not_compared": not_compared,
    });
    if !remaining.is_empty() {
        run["next"] = json!("moai update");
    }
    run
}

/// Whether this run leaves a record: only a run that wrote something does. A confirmed plan that
/// writes nothing (every host blocked, or nothing to apply) leaves the folder as it was, as
/// before; its `run` still sits in the result for the screen and a pipe to read.
fn leaves_record(run: &Value) -> bool {
    run["wrote"] == json!(true)
}

fn write_record(home: &Path, run: &Value) -> Result<PathBuf, String> {
    let id = current_install_id(home)?;
    let dir = crate::home_root::path(home).join("installations").join(id);
    if fs::symlink_metadata(&dir).is_ok_and(|m| m.file_type().is_symlink()) {
        return Err("the installation folder is a symlink; refused".into());
    }
    fs::create_dir_all(&dir).map_err(|_| "the installation folder is unavailable".to_string())?;
    let path = dir.join(FILE_NAME);
    let mut bytes =
        serde_json::to_vec_pretty(run).map_err(|_| "the run receipt did not encode".to_string())?;
    bytes.push(b'\n');
    let before = read_opt(&path)?;
    publish_replace(&path, before.as_deref(), &bytes)?;
    Ok(path)
}

/// Builds the record of a finished run and, when the run leaves one, writes it. A write failure
/// is named in `receipt_error` and changes nothing else.
pub fn finish(home: &Path, out: &Value) -> Value {
    let mut run = build(home, out);
    if leaves_record(&run) {
        match write_record(home, &run) {
            Ok(path) => run["receipt"] = json!(path.to_string_lossy()),
            Err(e) => run["receipt_error"] = json!(e),
        }
    }
    run
}

#[cfg(test)]
mod tests {
    use super::*;

    fn done(state: &str) -> Value {
        json!({"state": state})
    }

    #[test]
    fn a_clean_result_is_complete_and_leaves_nothing_open() {
        let out = json!({
            "applied": false,
            "result": {"hosts": [{"host": "claude", "status": "unchanged", "files": []}]},
            "agents_result": done("unchanged"),
        });
        let run = build(Path::new("/nonexistent-home"), &out);
        assert_eq!(run["status"], json!("complete"), "{run}");
        assert_eq!(run["remaining"], json!([]));
        assert!(!leaves_record(&run), "{run}");
    }

    #[test]
    fn a_failed_later_part_is_named_as_remaining_and_the_record_is_partial() {
        let out = json!({
            "applied": true,
            "result": {"hosts": [{"host": "claude", "status": "applied", "files": []}]},
            "agents_result": {"state": "error", "message": "boom"},
            "routing_result": done("applied"),
        });
        let run = build(Path::new("/nonexistent-home"), &out);
        assert_eq!(run["status"], json!("partial"), "{run}");
        assert_eq!(run["remaining"], json!(["agents"]), "{run}");
        assert_eq!(run["next"], json!("moai update"));
        assert!(leaves_record(&run));
        let agents = run["parts"]
            .as_array()
            .unwrap()
            .iter()
            .find(|p| p["part"] == "agents")
            .unwrap();
        assert_eq!(agents["reason"], json!("boom"));
    }

    #[test]
    fn a_conflict_and_a_kept_file_are_held_not_remaining() {
        let out = json!({
            "applied": true,
            "result": {"hosts": [{"host": "claude", "status": "applied",
                "files": [{"path": "a.md", "action": "conflict"}]}]},
            "agents_result": {"state": "partial", "kept": ["x.md"], "written": ["y.md"]},
        });
        let run = build(Path::new("/nonexistent-home"), &out);
        assert_eq!(run["status"], json!("partial"), "{run}");
        assert_eq!(run["remaining"], json!([]), "{run}");
        assert_eq!(run["held"], json!(["skills", "agents"]), "{run}");
        assert!(!run.as_object().unwrap().contains_key("next"));
    }

    #[test]
    fn a_link_outside_the_layout_is_held_and_a_blocked_run_writes_no_record() {
        let out = json!({
            "applied": false,
            "result": {"hosts": [{"host": "claude", "status": "unchanged", "files": []}]},
            "bin_names_result": {"state": "outside_layout", "reason": "not in a version folder"},
        });
        let run = build(Path::new("/nonexistent-home"), &out);
        assert_eq!(run["remaining"], json!([]), "{run}");
        assert_eq!(run["held"], json!(["bin_names"]), "{run}");
        assert!(!leaves_record(&run), "nothing was written: {run}");
        let blocked = json!({
            "applied": false,
            "result": {"hosts": [{"host": "claude", "status": "blocked", "files": []}]},
        });
        let run = build(Path::new("/nonexistent-home"), &blocked);
        assert_eq!(run["remaining"], json!(["skills"]), "{run}");
        assert!(
            !leaves_record(&run),
            "a blocked run writes no record: {run}"
        );
    }

    #[test]
    fn a_part_the_gate_skipped_is_remaining() {
        let out = json!({
            "applied": false,
            "result": {"hosts": [{"host": "claude", "status": "blocked", "files": [],
                "blocked_reason": "symlink"}]},
            "skipped": [
                {"part": "agents", "host": "claude", "reason": GATE_REASON},
                {"part": "routing", "host": "other", "reason": "not installed"},
            ],
        });
        let run = build(Path::new("/nonexistent-home"), &out);
        assert_eq!(run["remaining"], json!(["skills", "agents"]), "{run}");
    }

    #[test]
    fn every_part_without_a_full_comparison_is_listed_whatever_its_state() {
        let out = json!({
            "applied": true,
            "result": {"hosts": [{"host": "claude", "status": "applied", "files": [
                {"path": "mine.md", "action": "conflict", "published": false},
                {"path": "same.md", "action": "unchanged", "published": false},
            ]}]},
            "extra_skills_result": {"parts": [
                {"host": "claude", "skill": "x", "status": "unchanged",
                 "files": [{"path": "same.md", "action": "unchanged", "published": false}]},
            ]},
            "agents_result": {"state": "partial", "kept": ["x.md"], "written": ["y.md"],
                "receipt": "agents-1"},
            "style_result": done("unchanged"),
            "bin_names_result": done("applied"),
            "hooks_result": {"state": "partial", "receipt": "hooks-1"},
            "cleanup_result": {"state": "committed", "stamp": "c-1"},
            "home_result": {"state": "applied", "journal_state": "committed", "stamp": "h-1"},
            "switch_result": {"state": "applied", "stamp": "s-1"},
            "git_hooks_result": {"state": "applied", "stamp": "g-1"},
            "skipped": [{"part": "routing", "host": "claude", "reason": GATE_REASON}],
        });
        let run = build(Path::new("/nonexistent-home"), &out);
        let entry = |name: &str| {
            let found: Vec<&Value> = run["not_compared"]
                .as_array()
                .unwrap()
                .iter()
                .filter(|p| p["part"] == name)
                .collect();
            assert_eq!(found.len(), 1, "{name}: {run}");
            found[0].clone()
        };
        for (name, state, receipt) in [
            ("agents", "partial", Some("agents-1")),
            ("style", "unchanged", None),
            ("hooks", "partial", Some("hooks-1")),
            ("git_hooks", "applied", Some("g-1")),
            ("routing", "not_run", None),
        ] {
            let e = entry(name);
            assert_eq!(e["state"], json!(state), "{e}");
            assert_eq!(e["reason"], json!(why_not_compared("agents")), "{e}");
            assert_eq!(e["receipt"].as_str(), receipt, "{e}");
        }
        // The parts that write no file to compare say that instead.
        for name in ["bin_names", "cleanup", "home", "switch"] {
            let e = entry(name);
            let reason = e["reason"].as_str().unwrap();
            assert!(reason.starts_with("no digest to compare"), "{e}");
        }
        // The kept skill file was planned and never compared; the unchanged one was not planned.
        let skills = entry("skills");
        assert_eq!(
            skills["reason"],
            json!("1 of 1 planned files were not compared"),
            "{skills}"
        );
        // A skill folder with nothing planned is not listed.
        assert!(
            !run["not_compared"]
                .as_array()
                .unwrap()
                .iter()
                .any(|p| p["part"] == "extra_skills"),
            "{run}"
        );
    }

    #[test]
    fn a_skill_file_that_could_not_be_read_or_a_part_that_did_not_apply_is_listed() {
        let out = json!({
            "applied": true,
            "result": {"hosts": [{"host": "claude", "status": "applied", "files": [
                {"path": "gone.md", "action": "create", "published": true,
                 "after_sha256": "sha256:00"},
            ]}]},
            "extra_skills_result": {"state": "refused", "reason": "no package"},
        });
        let run = build(Path::new("/nonexistent-home"), &out);
        assert_eq!(run["files"][0]["check"], json!("missing"), "{run}");
        let list: Vec<(&str, &str)> = run["not_compared"]
            .as_array()
            .unwrap()
            .iter()
            .map(|p| (p["part"].as_str().unwrap(), p["reason"].as_str().unwrap()))
            .collect();
        assert_eq!(
            list,
            [
                ("skills", "1 of 1 planned files were not compared"),
                ("extra_skills", "the part did not apply"),
            ],
            "{run}"
        );
    }

    #[test]
    fn a_result_path_that_leaves_its_folder_is_never_read() {
        let root = Path::new("/nonexistent-root");
        assert!(inside(root, "../x").is_none());
        assert!(inside(root, "/etc/passwd").is_none());
        assert!(inside(root, "a/../b").is_none());
        assert_eq!(inside(root, "a/b.md"), Some(root.join("a/b.md")));
        let v = check_one("skills", "claude", "../x", root, "sha256:00");
        assert_eq!(v["check"], json!("unsafe_path"));
    }

    #[test]
    fn a_written_file_is_read_back_and_compared() {
        let dir = tempfile::TempDir::new().unwrap();
        let root = dir.path();
        fs::write(root.join("a.md"), b"one").unwrap();
        let good = sha256_hex(b"one");
        assert_eq!(
            check_one("skills", "claude", "a.md", root, &good)["check"],
            json!("match")
        );
        // The same digest without its prefix still matches.
        assert_eq!(
            check_one("skills", "claude", "a.md", root, bare(&good))["check"],
            json!("match")
        );
        assert_eq!(
            check_one("skills", "claude", "a.md", root, &sha256_hex(b"two"))["check"],
            json!("mismatch")
        );
        assert_eq!(
            check_one("skills", "claude", "gone.md", root, &good)["check"],
            json!("missing")
        );
        fs::create_dir(root.join("d")).unwrap();
        assert_eq!(
            check_one("skills", "claude", "d", root, &good)["check"],
            json!("not_a_file")
        );
    }
}
