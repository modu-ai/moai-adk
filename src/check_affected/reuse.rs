//! `check reuse` (SPEC-20261120-merge-reuse): after the local integration branch moved, decide
//! from git alone whether the evidence of an audited merge candidate still holds. [`judge`] is
//! pure (the facts in, a [`Verdict`] out); [`read_inputs`] is the one place that reads git.
//! Nothing is written except the tree objects `git merge-tree --write-tree` makes; no ref, no
//! index and no working file moves.

use super::{
    commit, declared_mods, git, host_os, os_names, select, stem, steps, Selection, TestFile,
};
use crate::repo_scans;
use std::collections::BTreeSet;
use std::path::Path;
use std::process::Command;

type Result<T> = std::result::Result<T, String>;

/// How the head merges into the tip.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Merge {
    Clean { tree: String },
    Conflict,
}

/// The files of `tests/` and the scan list of one tree.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct TreeFacts {
    pub tests: Vec<TestFile>,
    pub scan_list: Option<String>,
}

/// Everything [`judge`] needs.
#[derive(Clone, Debug)]
pub struct Inputs {
    pub base: String,
    pub head: String,
    pub tip: String,
    pub head_in_tip: bool,
    pub base_in_tip: bool,
    pub merge: Merge,
    /// The paths the head changes against the base.
    pub candidate: Vec<String>,
    /// The paths the tip changes against the base.
    pub delta: Vec<String>,
    /// The tests and scan list of the tree the candidate was checked on, and of the tree it would
    /// have on the tip.
    pub old: TreeFacts,
    pub new: TreeFacts,
    /// Which of the overlapping paths exist in the new tree (for the per-file steps).
    pub overlap_present: Vec<String>,
}

#[derive(Clone, Debug)]
pub struct Verdict {
    pub decision: &'static str,
    pub reasons: Vec<String>,
    pub rerun_rust: BTreeSet<String>,
    pub rerun_python: BTreeSet<String>,
    pub rerun_steps: Vec<String>,
    /// The checks of the scans whose rule files moved on one side while a path of their scope
    /// moved on the other (the `repo-scans` judge).
    pub rerun_scans: Vec<String>,
    pub overlap: Vec<String>,
    pub old: Option<Selection>,
    pub new: Option<Selection>,
    pub merge_tree: Option<String>,
}

fn verdict(decision: &'static str, reasons: &[&str]) -> Verdict {
    Verdict {
        decision,
        reasons: reasons.iter().map(|s| s.to_string()).collect(),
        rerun_rust: BTreeSet::new(),
        rerun_python: BTreeSet::new(),
        rerun_steps: Vec::new(),
        rerun_scans: Vec::new(),
        overlap: Vec::new(),
        old: None,
        new: None,
        merge_tree: None,
    }
}

/// The checks of every scan of `tools/repo_scans.json` whose rule files changed on one side while
/// a path in its scope changed on the other: two changes that touch different files and still
/// break once merged. The table is the union of the two trees' tables, as `moai repo-scans
/// --refs` reads it. An unreadable table is an error (the caller reruns everything).
fn scan_crossing(i: &Inputs) -> Result<(Vec<String>, Vec<String>)> {
    let tables = [&i.old.scan_list, &i.new.scan_list]
        .into_iter()
        .map(|text| text.as_deref().map(repo_scans::parse_table).transpose())
        .collect::<Result<Vec<_>>>()?;
    let scans = repo_scans::union(&tables);
    let mut ids: Vec<String> = repo_scans::judge(&i.candidate, &i.delta, &scans)
        .iter()
        .filter_map(|h| h["id"].as_str().map(String::from))
        .collect();
    ids.dedup();
    let wanted: Vec<&str> = ids.iter().map(String::as_str).collect();
    let checks = repo_scans::checks_for(&wanted, &scans);
    Ok((ids, checks))
}

/// The steps that check the whole change rather than a target: what an overlapping base move
/// invalidates (REQ-5).
fn whole_change_steps(sel: &Selection, present: &[String]) -> Vec<String> {
    steps(sel, present, "reuse", "reuse")
        .into_iter()
        .map(|s| s.label)
        .filter(|l| !matches!(l.as_str(), "test_targets" | "build_cli" | "python"))
        .collect()
}

/// The decision (REQ-1 to REQ-6, REQ-10). Pure.
pub fn judge(i: &Inputs, host: &str) -> Verdict {
    if i.tip == i.base {
        return verdict("fast_forward", &["base_is_tip"]);
    }
    if i.head_in_tip {
        return verdict("already_merged", &["head_is_in_tip"]);
    }
    if !i.base_in_tip {
        return verdict("rerun_all", &["base_not_ancestor_of_tip"]);
    }
    let tree = match &i.merge {
        Merge::Clean { tree } => tree.clone(),
        Merge::Conflict => return verdict("rerun_all", &["merge_conflict"]),
    };
    let old = select(&i.candidate, &i.old.tests, i.old.scan_list.as_deref(), host);
    let new = select(&i.candidate, &i.new.tests, i.new.scan_list.as_deref(), host);
    // What the base move can break by its own paths, by the name rules of `check affected`; only
    // the part the candidate's selection holds counts.
    let picked = select(&i.delta, &i.new.tests, None, "");
    let gained_rust: BTreeSet<String> = new.rust.difference(&old.rust).cloned().collect();
    let gained_python: BTreeSet<String> = new.python.difference(&old.python).cloned().collect();
    let picked_rust: BTreeSet<String> = picked.rust.intersection(&new.rust).cloned().collect();
    let picked_python: BTreeSet<String> =
        picked.python.intersection(&new.python).cloned().collect();
    let candidate: BTreeSet<&String> = i.candidate.iter().collect();
    let overlap: Vec<String> = {
        let mut o: Vec<String> = i
            .delta
            .iter()
            .filter(|p| candidate.contains(p))
            .cloned()
            .collect();
        o.sort();
        o.dedup();
        o
    };
    let mut reasons = vec![
        "base_is_ancestor_of_tip".to_string(),
        "merge_is_clean".to_string(),
    ];
    let mut rerun_rust = gained_rust.clone();
    rerun_rust.extend(picked_rust.iter().cloned());
    let mut rerun_python = gained_python.clone();
    rerun_python.extend(picked_python.iter().cloned());
    if !gained_rust.is_empty() || !gained_python.is_empty() {
        reasons.push("selection_gained_a_target".to_string());
    }
    if !picked_rust.is_empty() || !picked_python.is_empty() {
        reasons.push("base_move_reaches_a_selected_target".to_string());
    }
    let mut steps_to_run = Vec::new();
    if overlap.is_empty() {
        reasons.push("no_shared_path".to_string());
    } else {
        reasons.push("shared_path".to_string());
        steps_to_run = whole_change_steps(&new, &i.overlap_present);
    }
    let (scan_ids, scan_checks) = match scan_crossing(i) {
        Ok(found) => found,
        Err(e) => {
            let mut v = verdict("rerun_all", &[]);
            v.reasons.push(format!("scan_table_unreadable: {e}"));
            return v;
        }
    };
    if !scan_ids.is_empty() {
        reasons.push("scan_rules_cross_scope".to_string());
    }
    let covered = old.rust.is_superset(&new.rust) && old.python.is_superset(&new.python);
    let equal = old.rust == new.rust && old.python == new.python && old.scans == new.scans;
    let decision = if rerun_rust.is_empty()
        && rerun_python.is_empty()
        && steps_to_run.is_empty()
        && scan_ids.is_empty()
        && covered
    {
        reasons.push(
            if equal {
                "selection_is_equal"
            } else {
                "selection_is_covered"
            }
            .to_string(),
        );
        "reuse"
    } else {
        "rerun"
    };
    Verdict {
        decision,
        reasons,
        rerun_rust,
        rerun_python,
        rerun_steps: steps_to_run,
        rerun_scans: scan_checks,
        overlap,
        old: Some(old),
        new: Some(new),
        merge_tree: Some(tree),
    }
}

fn is_ancestor(dir: &Path, a: &str, b: &str) -> Result<bool> {
    let mut cmd = Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let o = cmd
        .args(["merge-base", "--is-ancestor", a, b])
        .current_dir(dir)
        .output()
        .map_err(|e| format!("cannot run git: {e}"))?;
    match o.status.code() {
        Some(0) => Ok(true),
        Some(1) => Ok(false),
        _ => Err(format!(
            "git merge-base failed: {}",
            String::from_utf8_lossy(&o.stderr).trim()
        )),
    }
}

/// `git merge-tree --write-tree`: the tree of the merge, or the fact that it conflicts. It writes
/// tree objects and nothing else.
fn merge_tree(dir: &Path, a: &str, b: &str) -> Result<Merge> {
    let mut cmd = Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let o = cmd
        .args(["merge-tree", "--write-tree", "--no-messages", a, b])
        .current_dir(dir)
        .output()
        .map_err(|e| format!("cannot run git: {e}"))?;
    match o.status.code() {
        Some(0) => Ok(Merge::Clean {
            tree: String::from_utf8_lossy(&o.stdout)
                .lines()
                .next()
                .unwrap_or("")
                .trim()
                .to_string(),
        }),
        Some(1) => Ok(Merge::Conflict),
        _ => Err(format!(
            "git merge-tree failed: {}",
            String::from_utf8_lossy(&o.stderr).trim()
        )),
    }
}

fn paths(dir: &Path, args: &[&str]) -> Result<Vec<String>> {
    Ok(git(dir, args)?
        .split('\0')
        .filter(|l| !l.is_empty())
        .map(String::from)
        .collect())
}

/// The test files and the scan list of `tree`, read from git, never from the checkout.
pub fn tree_facts(dir: &Path, tree: &str) -> Result<TreeFacts> {
    // `--full-tree`: the pathspecs are repository paths whatever folder `dir` is.
    let listing = git(dir, &["ls-tree", "-z", "--full-tree", tree, "--", "tests/"])?;
    let mut tests = Vec::new();
    let mut rust: Vec<(String, String)> = Vec::new();
    for entry in listing.split('\0').filter(|l| !l.is_empty()) {
        let Some((meta, path)) = entry.split_once('\t') else {
            continue;
        };
        let mut words = meta.split_whitespace();
        let (_mode, kind, sha) = (words.next(), words.next(), words.next().unwrap_or(""));
        if kind != Some("blob") {
            continue;
        }
        let Some(file) = path.strip_prefix("tests/").filter(|f| !f.contains('/')) else {
            continue;
        };
        let name = stem(file).to_string();
        if file.ends_with(".rs") {
            rust.push((name, sha.to_string()));
        } else if file.ends_with(".py") && name.ends_with("_test") {
            tests.push(TestFile {
                name,
                python: true,
                ..TestFile::default()
            });
        }
    }
    let scan = git(
        dir,
        &[
            "ls-tree",
            "-z",
            "--full-tree",
            tree,
            "--",
            "tools/repo_scans.json",
        ],
    )?;
    let scan_sha = scan
        .split('\0')
        .find(|l| !l.is_empty())
        .and_then(|entry| entry.split_once('\t'))
        .and_then(|(m, _)| m.split_whitespace().nth(2))
        .map(String::from);
    // One `git cat-file --batch` for every blob: a process per test file made the answer slow.
    let mut shas: Vec<&str> = rust.iter().map(|(_, s)| s.as_str()).collect();
    shas.extend(scan_sha.as_deref());
    let mut texts = blobs(dir, &shas)?.into_iter();
    for (name, _) in &rust {
        let text = texts.next().unwrap_or_default();
        tests.push(TestFile {
            name: name.clone(),
            python: false,
            modules: declared_mods(&text),
            os: os_names(&text),
        });
    }
    let scan_list = scan_sha.map(|_| texts.next().unwrap_or_default());
    tests.sort_by(|a, b| (a.python, &a.name).cmp(&(b.python, &b.name)));
    Ok(TreeFacts { tests, scan_list })
}

/// The text of each blob, in the order asked, from one `git cat-file --batch`.
fn blobs(dir: &Path, shas: &[&str]) -> Result<Vec<String>> {
    use std::io::Write;
    use std::process::Stdio;
    if shas.is_empty() {
        return Ok(Vec::new());
    }
    let mut cmd = Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let mut child = cmd
        .args(["cat-file", "--batch"])
        .current_dir(dir)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .map_err(|e| format!("cannot run git: {e}"))?;
    // The request is written from its own thread while this one reads: git stops reading its
    // input once its answer fills the output pipe, so writing everything first can hang when the
    // tree is large.
    let mut stdin = child.stdin.take().ok_or("git cat-file has no stdin")?;
    let mut request = String::new();
    for sha in shas {
        request.push_str(sha);
        request.push('\n');
    }
    let writer = std::thread::spawn(move || stdin.write_all(request.as_bytes()));
    let out = child
        .wait_with_output()
        .map_err(|e| format!("cannot read git: {e}"))?;
    let written = writer.join().map_err(|_| "the request writer panicked")?;
    if !out.status.success() {
        return Err(format!(
            "git cat-file failed: {}",
            String::from_utf8_lossy(&out.stderr).trim()
        ));
    }
    written.map_err(|e| format!("cannot write to git: {e}"))?;
    let bytes = out.stdout;
    let mut texts = Vec::with_capacity(shas.len());
    let mut at = 0;
    for sha in shas {
        let end = bytes[at..]
            .iter()
            .position(|b| *b == b'\n')
            .ok_or("git cat-file answer is cut short")?
            + at;
        let header = String::from_utf8_lossy(&bytes[at..end]).into_owned();
        let words: Vec<&str> = header.split_whitespace().collect();
        let size: usize = match words.as_slice() {
            [_, "blob", size] => size
                .parse()
                .map_err(|_| format!("bad size in `{header}`"))?,
            _ => return Err(format!("git cat-file did not find {sha}: `{header}`")),
        };
        let body = bytes
            .get(end + 1..end + 1 + size)
            .ok_or("git cat-file answer is cut short")?;
        texts.push(String::from_utf8_lossy(body).into_owned());
        at = end + 1 + size + 1;
    }
    Ok(texts)
}

fn exists_in(dir: &Path, tree: &str, path: &str) -> bool {
    git(dir, &["cat-file", "-e", &format!("{tree}:{path}")]).is_ok()
}

/// Read what [`judge`] needs. The revisions were checked by the caller; any failure from here on
/// is an unreadable fact, not a refusal.
pub fn read_inputs(dir: &Path, base: &str, head: &str, tip: &str) -> Result<Inputs> {
    let head_in_tip = is_ancestor(dir, head, tip)?;
    let base_in_tip = is_ancestor(dir, base, tip)?;
    let candidate = paths(
        dir,
        &[
            "diff",
            "--name-only",
            "--no-renames",
            "-z",
            &format!("{base}...{head}"),
        ],
    )?;
    let delta = if base_in_tip {
        paths(
            dir,
            &["diff", "--name-only", "--no-renames", "-z", base, tip],
        )?
    } else {
        Vec::new()
    };
    let merge = if base_in_tip && !head_in_tip && base != tip {
        merge_tree(dir, tip, head)?
    } else {
        Merge::Conflict
    };
    let (old, new, overlap_present) = match &merge {
        Merge::Clean { tree } => {
            let old_tree = match merge_tree(dir, base, head)? {
                Merge::Clean { tree } => tree,
                Merge::Conflict => return Err("the candidate does not merge into its base".into()),
            };
            let overlap: BTreeSet<&String> = delta.iter().collect();
            let present = candidate
                .iter()
                .filter(|p| overlap.contains(p) && exists_in(dir, tree, p))
                .cloned()
                .collect();
            (tree_facts(dir, &old_tree)?, tree_facts(dir, tree)?, present)
        }
        Merge::Conflict => (TreeFacts::default(), TreeFacts::default(), Vec::new()),
    };
    Ok(Inputs {
        base: base.to_string(),
        head: head.to_string(),
        tip: tip.to_string(),
        head_in_tip,
        base_in_tip,
        merge,
        candidate,
        delta,
        old,
        new,
        overlap_present,
    })
}

fn names(set: &BTreeSet<String>) -> Vec<&String> {
    set.iter().collect()
}

fn selection_json(sel: &Option<Selection>) -> serde_json::Value {
    match sel {
        None => serde_json::Value::Null,
        Some(s) => serde_json::json!({
            "rust": names(&s.rust),
            "python": names(&s.python),
            "scans": s.scans.word(),
        }),
    }
}

/// The answer of `check reuse --base <b> --head <h> --tip <t>`.
pub fn run(dir: &Path, base: &str, head: &str, tip: &str) -> Result<serde_json::Value> {
    let base = commit(dir, base)?;
    let head = commit(dir, head)?;
    let tip = commit(dir, tip)?;
    let (v, counts) = match read_inputs(dir, &base, &head, &tip) {
        Ok(inputs) => (
            judge(&inputs, host_os()),
            (inputs.candidate.len(), inputs.delta.len()),
        ),
        Err(e) => (
            {
                let mut v = verdict("rerun_all", &[]);
                v.reasons.push(format!("unreadable: {e}"));
                v
            },
            (0, 0),
        ),
    };
    Ok(serde_json::json!({
        "action": "reuse",
        "decision": v.decision,
        "reasons": v.reasons,
        "rerun": {
            "rust": names(&v.rerun_rust),
            "python": names(&v.rerun_python),
            "steps": v.rerun_steps,
            "scan_checks": v.rerun_scans,
        },
        "overlap": v.overlap,
        "base": base,
        "head": head,
        "tip": tip,
        "candidate_paths": counts.0,
        "delta_paths": counts.1,
        "selection": {"old": selection_json(&v.old), "new": selection_json(&v.new)},
        "merge_tree": v.merge_tree,
    }))
}

/// `check reuse --merge <commit> --tip <t>`: the first parent is the base, the second the head.
pub fn run_merge(dir: &Path, merge: &str, tip: &str) -> Result<serde_json::Value> {
    let merge = commit(dir, merge)?;
    let line = git(dir, &["rev-list", "--parents", "-n", "1", &merge])?;
    let words: Vec<&str> = line.split_whitespace().collect();
    if words.len() != 3 {
        return Err(format!(
            "check reuse refused: `{merge}` is not a merge of exactly two parents"
        ));
    }
    run(dir, words[1], words[2], tip)
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::time::Duration;

    fn git_in(dir: &Path, args: &[&str]) -> String {
        let mut cmd = Command::new("git");
        crate::identity::scrub_git_env(&mut cmd);
        let out = cmd.args(args).current_dir(dir).output().unwrap();
        assert!(out.status.success(), "git {args:?}");
        String::from_utf8_lossy(&out.stdout).trim().to_string()
    }

    /// The request is longer than a pipe holds and the answer is larger than a pipe holds: the
    /// reader must not wait for the whole request to be written first.
    #[test]
    fn blobs_reads_a_large_answer_while_the_request_is_still_being_written() {
        let dir = tempfile::tempdir().unwrap();
        git_in(dir.path(), &["init", "-q"]);
        std::fs::write(dir.path().join("big.txt"), "x".repeat(8_000)).unwrap();
        let sha = git_in(dir.path(), &["hash-object", "-w", "big.txt"]);
        let path = dir.path().to_path_buf();
        let (tx, rx) = std::sync::mpsc::channel();
        std::thread::spawn(move || {
            let shas = vec![sha.as_str(); 8000];
            let _ = tx.send(blobs(&path, &shas).map(|t| (t.len(), t[0].len())));
        });
        let got = rx
            .recv_timeout(Duration::from_secs(60))
            .expect("blobs hung")
            .unwrap();
        assert_eq!(got, (8000, 8_000));
    }
}
