//! The repo-wide scan judge (#2681, leader decision 6848942944), moved from
//! `tools/repo_scans.py` (#2708 4/4) so the shipped leader workflow needs only `moai-cli`.
//!
//! Some tests read the checked-in tree itself (`tools/repo_scans.json`). Two branches that touch
//! disjoint files still fail once merged when one side changes a scan's rule files and the other
//! side changes a path inside that scan's scope. The gate then runs the listed `checks` on the
//! merge result before merging.
//!
//! The report is `{remerge, hits, checks, reasons}`, the same object the Python judge printed.
//! Any internal error fails safe to `remerge: true` with every check in the table; when no table
//! can be read either, `checks` is empty and a reason says "hold the merge" (#2693 audit N1).

use serde_json::{json, Value};
use std::path::Path;

/// The table's path inside a repository.
pub const TABLE_IN_REPO: &str = "tools/repo_scans.json";
/// A table or a stdin change set larger than this is refused (fail safe).
pub const INPUT_LIMIT: usize = 4 * 1024 * 1024;

/// One scan of the table, checked for the fields the judge reads.
#[derive(Debug, Clone, PartialEq)]
pub struct Scan {
    pub id: String,
    pub rules: Vec<String>,
    pub scope: Vec<(String, Vec<String>)>,
    pub checks: Vec<String>,
    /// The entry as written, so two tables are merged by whole-entry equality.
    pub raw: Value,
}

fn strings(v: &Value, what: &str) -> Result<Vec<String>, String> {
    v.as_array()
        .ok_or_else(|| format!("{what} is not a list"))?
        .iter()
        .map(|s| {
            s.as_str()
                .map(str::to_string)
                .ok_or_else(|| format!("{what} holds a non-string"))
        })
        .collect()
}

/// The scans of a table's text (`{"scans": [...]}`); an error for anything malformed.
pub fn parse_table(text: &str) -> Result<Vec<Scan>, String> {
    if text.len() > INPUT_LIMIT {
        return Err("the table is larger than 4 MiB".into());
    }
    let doc: Value = serde_json::from_str(text).map_err(|e| format!("table: {e}"))?;
    let list = doc
        .get("scans")
        .and_then(Value::as_array)
        .ok_or("table: `scans` is not a list")?;
    let mut out = Vec::new();
    for raw in list {
        let id = raw
            .get("id")
            .and_then(Value::as_str)
            .ok_or("table: a scan has no string `id`")?
            .to_string();
        let rules = strings(raw.get("rules").unwrap_or(&Value::Null), "rules")?;
        let checks = strings(raw.get("checks").unwrap_or(&Value::Null), "checks")?;
        let mut scope = Vec::new();
        for item in raw
            .get("scope")
            .and_then(Value::as_array)
            .ok_or_else(|| format!("table: scan {id} has no `scope` list"))?
        {
            let prefix = item
                .get("prefix")
                .and_then(Value::as_str)
                .ok_or_else(|| format!("table: scan {id} has a scope item without a prefix"))?
                .to_string();
            let suffixes = match item.get("suffixes") {
                None | Some(Value::Null) => Vec::new(),
                Some(v) => strings(v, "suffixes")?,
            };
            scope.push((prefix, suffixes));
        }
        out.push(Scan {
            id,
            rules,
            scope,
            checks,
            raw: raw.clone(),
        });
    }
    Ok(out)
}

/// Entries of every table, first table first, each distinct entry once (#2693 audit B1).
pub fn union(tables: &[Option<Vec<Scan>>]) -> Vec<Scan> {
    let mut out: Vec<Scan> = Vec::new();
    for scan in tables.iter().flatten().flatten() {
        if !out.iter().any(|s| s.raw == scan.raw) {
            out.push(scan.clone());
        }
    }
    out
}

/// Prefix match plus an optional suffix list, no globs (an empty prefix is every path).
pub fn in_scope(path: &str, scope: &[(String, Vec<String>)]) -> bool {
    scope.iter().any(|(prefix, suffixes)| {
        path.starts_with(prefix.as_str())
            && (suffixes.is_empty() || suffixes.iter().any(|s| path.ends_with(s.as_str())))
    })
}

/// The hits: a scan whose rule files changed on one side while a path in its scope changed on
/// the other side.
pub fn judge(pr: &[String], tip: &[String], scans: &[Scan]) -> Vec<Value> {
    let pr: Vec<&String> = pr.iter().filter(|p| !p.is_empty()).collect();
    let tip: Vec<&String> = tip.iter().filter(|p| !p.is_empty()).collect();
    let mut hits = Vec::new();
    for scan in scans {
        for (rules_side, rule_paths, scope_side, scope_paths) in
            [("pr", &pr, "tip", &tip), ("tip", &tip, "pr", &pr)]
        {
            let mut changed: Vec<&String> = rule_paths
                .iter()
                .copied()
                .filter(|p| scan.rules.contains(p))
                .collect();
            if changed.is_empty() {
                continue;
            }
            changed.sort();
            let mut touched: Vec<&String> = scope_paths
                .iter()
                .copied()
                .filter(|p| in_scope(p, &scan.scope))
                .collect();
            if touched.is_empty() {
                continue;
            }
            touched.sort();
            hits.push(json!({
                "id": scan.id,
                "rules_side": rules_side,
                "rules": changed,
                "scope_side": scope_side,
                "scope": touched,
            }));
        }
    }
    hits
}

/// The check commands of every hit scan id, in table order, without duplicates.
pub fn checks_for(ids: &[&str], scans: &[Scan]) -> Vec<String> {
    let mut out: Vec<String> = Vec::new();
    for scan in scans.iter().filter(|s| ids.contains(&s.id.as_str())) {
        for c in &scan.checks {
            if !out.contains(c) {
                out.push(c.clone());
            }
        }
    }
    out
}

fn git(repo: &Path, args: &[&str]) -> Result<Vec<u8>, String> {
    let out = crate::audit_delta::command(repo)
        .args(args)
        .output()
        .map_err(|e| format!("cannot execute Git: {e}"))?;
    if !out.status.success() {
        return Err(format!(
            "git {} failed: {}",
            args.first().copied().unwrap_or(""),
            crate::identity::failure_summary(&out.stderr)
        ));
    }
    Ok(out.stdout)
}

fn commit_of(repo: &Path, rev: &str) -> Result<String, String> {
    if rev.is_empty() || rev.starts_with('-') {
        return Err(format!("invalid revision {rev:?}"));
    }
    let out = git(
        repo,
        &[
            "rev-parse",
            "--verify",
            "--quiet",
            "--end-of-options",
            &format!("{rev}^{{commit}}"),
        ],
    )
    .map_err(|_| format!("unknown revision {rev:?}"))?;
    Ok(String::from_utf8_lossy(&out).trim().to_string())
}

fn changed(repo: &Path, base: &str, rev: &str) -> Result<Vec<String>, String> {
    let out = git(
        repo,
        &["diff", "--name-only", "--no-renames", "-z", base, rev, "--"],
    )?;
    out.split(|b| *b == 0)
        .filter(|p| !p.is_empty())
        .map(|p| {
            String::from_utf8(p.to_vec()).map_err(|_| "a changed path is not UTF-8".to_string())
        })
        .collect()
}

/// The table as committed at `commit`, or `None` when that commit has none (older than #2681).
/// A table that exists but does not parse is an error, so the caller fails safe.
fn load_ref(repo: &Path, commit: &str) -> Result<Option<Vec<Scan>>, String> {
    let spec = format!("{commit}:{TABLE_IN_REPO}");
    if crate::audit_delta::command(repo)
        .args(["cat-file", "-e", &spec])
        .output()
        .map(|o| !o.status.success())
        .unwrap_or(true)
    {
        return Ok(None);
    }
    let bytes = git(repo, &["cat-file", "blob", &spec])?;
    let text = String::from_utf8(bytes).map_err(|_| "the table is not UTF-8".to_string())?;
    parse_table(&text).map(Some)
}

/// The table in a file (the fallback table).
pub fn load_file(path: &Path) -> Result<Vec<Scan>, String> {
    let meta = std::fs::metadata(path).map_err(|e| format!("{}: {e}", path.display()))?;
    if meta.len() > INPUT_LIMIT as u64 {
        return Err("the table is larger than 4 MiB".into());
    }
    let text = std::fs::read_to_string(path).map_err(|e| format!("{}: {e}", path.display()))?;
    parse_table(&text)
}

/// Where the change sets come from.
pub enum Input {
    /// `--refs <tip> <head>`: `git diff --name-only` from their merge-base, tables at both refs.
    Refs(String, String),
    /// `{"pr": [...], "tip": [...]}` on stdin, judged with the fallback table.
    Stdin(String),
    /// Arguments the command does not take; reported as an internal error (fail safe).
    Bad(String),
}

fn stdin_sets(text: &str) -> Result<(Vec<String>, Vec<String>), String> {
    if text.len() > INPUT_LIMIT {
        return Err("stdin is larger than 4 MiB".into());
    }
    let v: Value = serde_json::from_str(text).map_err(|e| format!("stdin: {e}"))?;
    let pr = v.get("pr").ok_or("stdin has no `pr`")?;
    let tip = v.get("tip").ok_or("stdin has no `tip`")?;
    if !pr.is_array() || !tip.is_array() {
        return Err("pr and tip must be lists of paths".into());
    }
    Ok((strings(pr, "pr")?, strings(tip, "tip")?))
}

/// The judge's report. `repo` runs Git (the `--refs` form); `fallback` is the table used for
/// stdin input and when neither ref has a table. Never fails: errors become the fail-safe report.
pub fn report(input: &Input, repo: &Path, fallback: &Path) -> Value {
    let mut scans: Vec<Scan> = Vec::new();
    let run = |scans: &mut Vec<Scan>| -> Result<(Vec<Value>, Vec<String>), String> {
        let (pr, tip) = match input {
            Input::Bad(msg) => return Err(msg.clone()),
            Input::Refs(tip_rev, head_rev) => {
                let tip_c = commit_of(repo, tip_rev)?;
                let head_c = commit_of(repo, head_rev)?;
                let base = String::from_utf8_lossy(&git(repo, &["merge-base", &tip_c, &head_c])?)
                    .trim()
                    .to_string();
                let sets = (
                    changed(repo, &base, &head_c)?,
                    changed(repo, &base, &tip_c)?,
                );
                let tables = [load_ref(repo, &head_c)?, load_ref(repo, &tip_c)?];
                *scans = if tables.iter().any(Option::is_some) {
                    union(&tables)
                } else {
                    load_file(fallback)?
                };
                sets
            }
            Input::Stdin(text) => {
                *scans = load_file(fallback)?;
                stdin_sets(text)?
            }
        };
        let hits = judge(&pr, &tip, scans);
        let ids: Vec<&str> = hits.iter().filter_map(|h| h["id"].as_str()).collect();
        Ok((hits.clone(), checks_for(&ids, scans)))
    };
    match run(&mut scans) {
        Ok((hits, checks)) => {
            let reasons: Vec<String> = hits
                .iter()
                .map(|h| {
                    format!(
                        "{}: rules changed on {}, scope changed on {}",
                        h["id"].as_str().unwrap_or(""),
                        h["rules_side"].as_str().unwrap_or(""),
                        h["scope_side"].as_str().unwrap_or("")
                    )
                })
                .collect();
            json!({"remerge": !hits.is_empty(), "hits": hits, "checks": checks, "reasons": reasons})
        }
        Err(e) => {
            let mut reasons = vec![format!("judge internal error: {e}")];
            if scans.is_empty() {
                match load_file(fallback) {
                    Ok(t) => scans = t,
                    Err(te) => reasons.push(format!("no table readable ({te}): hold the merge")),
                }
            }
            let ids: Vec<&str> = scans.iter().map(|s| s.id.as_str()).collect();
            let checks = checks_for(&ids, &scans);
            json!({"remerge": true, "hits": [], "checks": checks, "reasons": reasons})
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn scan(id: &str, rules: &[&str], prefix: &str, suffixes: &[&str], checks: &[&str]) -> Scan {
        let raw = json!({"id": id, "rules": rules, "checks": checks,
                         "scope": [{"prefix": prefix, "suffixes": suffixes}]});
        parse_table(&json!({"scans": [raw]}).to_string())
            .unwrap()
            .remove(0)
    }

    fn s(v: &[&str]) -> Vec<String> {
        v.iter().map(|x| x.to_string()).collect()
    }

    #[test]
    fn scope_is_a_prefix_with_optional_suffixes() {
        let sc = vec![("src/".to_string(), s(&[".rs"]))];
        assert!(in_scope("src/a.rs", &sc));
        assert!(in_scope("src/forge/a.rs", &sc));
        assert!(!in_scope("src/a.md", &sc));
        assert!(!in_scope("srcx/a.rs", &sc));
        assert!(in_scope("anything", &[(String::new(), Vec::new())]));
    }

    #[test]
    fn both_directions_and_order() {
        let t = vec![
            scan("a", &["r"], "s/", &[], &["x", "y"]),
            scan("b", &["r"], "s/", &[], &["y", "z"]),
        ];
        let hits = judge(&s(&["r"]), &s(&["s/1", ""]), &t);
        assert_eq!(hits.len(), 2);
        assert_eq!(hits[0]["rules_side"], "pr");
        assert_eq!(checks_for(&["b", "a"], &t), s(&["x", "y", "z"]));
        let back = judge(&s(&["s/1"]), &s(&["r"]), &t);
        assert_eq!(back[0]["rules_side"], "tip");
        assert!(judge(&s(&["r"]), &s(&[]), &t).is_empty());
    }

    #[test]
    fn empty_paths_are_dropped_before_the_judge() {
        // an empty prefix is every path, but an empty path is no change (#2708 mutant M8)
        let every = vec![scan("all", &["r"], "", &[], &["c"])];
        assert!(judge(&s(&["r"]), &s(&[""]), &every).is_empty());
        assert!(judge(&s(&[""]), &s(&["r"]), &every).is_empty());
        assert_eq!(judge(&s(&["r"]), &s(&["x"]), &every).len(), 1);
    }

    #[test]
    fn malformed_tables_are_errors() {
        for bad in [
            "{}",
            r#"{"scans": {}}"#,
            r#"{"scans": [{"id": 1}]}"#,
            r#"{"scans": [{"id": "a", "rules": [], "checks": []}]}"#,
            r#"{"scans": [{"id": "a", "rules": [1], "checks": [], "scope": []}]}"#,
            r#"{"scans": [{"id": "a", "rules": [], "checks": [], "scope": [{}]}]}"#,
        ] {
            assert!(parse_table(bad).is_err(), "{bad}");
        }
    }
}
