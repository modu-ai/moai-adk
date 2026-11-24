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
    /// `"always": true`: a path of the head's change set in the scope is a hit by itself, without
    /// a change of a rule file (SPEC-20261123-land-repo-scans).
    pub always: bool,
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

/// The command a scan's `test_targets` stand for, up to the `--test` pairs.
const CARGO_TEST: &str = "cargo test --offline --locked";

/// A cargo integration-test target name as the scan registry spells it: `[a-z0-9_]+`. The one
/// function for the values of the `--test` options of a registry line: `gate::registry_only_grew`
/// reads them with it too (SPEC-20261124-registry-line-add-tier, REQ-11). The function of the
/// same name in `bundle/verify.rs` is another rule for another list (the names of test files a
/// bundle check lists: it also takes `-`, never a leading one), and is not this one.
pub fn is_target_name(name: &str) -> bool {
    !name.is_empty()
        && name
            .chars()
            .all(|c| c.is_ascii_lowercase() || c.is_ascii_digit() || c == '_')
}

/// The check commands of one scan entry of the table, the one place that reads them
/// (SPEC-20261124-repo-scans-targets-per-line): the command built from `test_targets`, in the
/// array's order, then the entries of `checks`. A scan without `test_targets` gives its `checks`
/// as listed, and `checks` may be left out when `test_targets` is there. A `test_targets` that is
/// no list of target names, and a scan with neither field, are errors that name the scan.
pub fn checks_of(entry: &Value) -> Result<Vec<String>, String> {
    let id = entry.get("id").and_then(Value::as_str).unwrap_or("?");
    let mut out = Vec::new();
    let targets = entry.get("test_targets").filter(|v| !v.is_null());
    if let Some(targets) = targets {
        let list = targets
            .as_array()
            .ok_or_else(|| format!("table: scan {id}: `test_targets` is not a list"))?;
        let mut command = CARGO_TEST.to_string();
        for target in list {
            let name = target
                .as_str()
                .filter(|name| is_target_name(name))
                .ok_or_else(|| {
                    format!(
                        "table: scan {id}: `test_targets` holds {target}, which is no target \
                         name ([a-z0-9_]+)"
                    )
                })?;
            command.push_str(" --test ");
            command.push_str(name);
        }
        if !list.is_empty() {
            out.push(command);
        }
    }
    match entry.get("checks").filter(|v| !v.is_null()) {
        None if targets.is_some() => {}
        None => {
            return Err(format!(
                "table: scan {id} has neither `test_targets` nor `checks`"
            ))
        }
        Some(checks) => {
            out.extend(strings(checks, "checks").map_err(|e| format!("table: scan {id}: {e}"))?)
        }
    }
    Ok(out)
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
        let checks = checks_of(raw)?;
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
        let always = match raw.get("always") {
            None | Some(Value::Null) => false,
            Some(Value::Bool(b)) => *b,
            Some(_) => {
                return Err(format!(
                    "table: scan {id} has an `always` that is not a boolean"
                ))
            }
        };
        out.push(Scan {
            id,
            rules,
            scope,
            checks,
            always,
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

/// The own hits: an `always` scan with a path of the head's change set in its scope, whatever
/// changed on the other side. A scan already in `hits` is not listed again, and a change of the
/// tip side alone is never an own hit.
pub fn own_hits(pr: &[String], hits: &[Value], scans: &[Scan]) -> Vec<Value> {
    let pr: Vec<&String> = pr.iter().filter(|p| !p.is_empty()).collect();
    let mut out = Vec::new();
    for scan in scans.iter().filter(|s| s.always) {
        if hits
            .iter()
            .any(|h| h["id"].as_str() == Some(scan.id.as_str()))
        {
            continue;
        }
        let mut touched: Vec<&String> = pr
            .iter()
            .copied()
            .filter(|p| in_scope(p, &scan.scope))
            .collect();
        if touched.is_empty() {
            continue;
        }
        touched.sort();
        out.push(json!({"id": scan.id, "scope_side": "pr", "scope": touched}));
    }
    out
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
    /// What one judge run answers: the hits across the two sides, the own hits and the checks.
    type Judged = Result<(Vec<Value>, Vec<Value>, Vec<String>), String>;
    let run = |scans: &mut Vec<Scan>| -> Judged {
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
        let own = own_hits(&pr, &hits, scans);
        let ids: Vec<&str> = hits
            .iter()
            .chain(own.iter())
            .filter_map(|h| h["id"].as_str())
            .collect();
        Ok((hits.clone(), own.clone(), checks_for(&ids, scans)))
    };
    match run(&mut scans) {
        Ok((hits, own, checks)) => {
            let mut reasons: Vec<String> = hits
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
            reasons.extend(own.iter().map(|h| {
                format!(
                    "{}: scope changed on pr, the scan runs on every such change",
                    h["id"].as_str().unwrap_or("")
                )
            }));
            // `remerge` is set by the hits across the two sides alone; `own_hits` appears only
            // when there is one, so a report without it keeps its earlier shape.
            let mut report = json!({
                "remerge": !hits.is_empty(), "hits": hits, "checks": checks, "reasons": reasons,
            });
            if !own.is_empty() {
                report["own_hits"] = Value::Array(own);
            }
            report
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

/// One scan a landing needs: its id and its check commands.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Needed {
    pub id: String,
    pub checks: Vec<String>,
}

/// The scans one pair of refs needs: the hits across the two sides and the own hits of `head_rev`.
/// `Ok(None)` when neither ref holds a table; a table that is there and cannot be read is an error.
fn needed_between(
    repo: &Path,
    tip_rev: &str,
    head_rev: &str,
) -> Result<Option<Vec<Needed>>, String> {
    let tip_c = commit_of(repo, tip_rev)?;
    let head_c = commit_of(repo, head_rev)?;
    let tables = [load_ref(repo, &head_c)?, load_ref(repo, &tip_c)?];
    if tables.iter().all(Option::is_none) {
        return Ok(None);
    }
    let scans = union(&tables);
    let base = String::from_utf8_lossy(&git(repo, &["merge-base", &tip_c, &head_c])?)
        .trim()
        .to_string();
    let pr = changed(repo, &base, &head_c)?;
    let tip = changed(repo, &base, &tip_c)?;
    let hits = judge(&pr, &tip, &scans);
    let own = own_hits(&pr, &hits, &scans);
    let ids: Vec<&str> = hits
        .iter()
        .chain(own.iter())
        .filter_map(|h| h["id"].as_str())
        .collect();
    Ok(Some(
        scans
            .iter()
            .filter(|s| ids.contains(&s.id.as_str()))
            .map(|s| Needed {
                id: s.id.clone(),
                checks: s.checks.clone(),
            })
            .collect(),
    ))
}

/// The scans the merge tree of a landing needs: `head` is the commit the branch would move to and
/// `tip` the branch's tip, on the first-parent line of `head`. Judged for the range `tip..head` as
/// one change, and for each merge commit of that line against its two parents (the lane's own
/// delta against what the tip gained since the lane's base). A scan is listed once, in the order
/// it first appears; a repository with no table at any of the refs needs none.
pub fn needed(repo: &Path, tip: &str, head: &str) -> Result<Vec<Needed>, String> {
    let mut pairs: Vec<(String, String)> = vec![(tip.to_string(), head.to_string())];
    let range = format!("{tip}..{head}");
    let line = git(repo, &["rev-list", "--first-parent", "--parents", &range])?;
    for record in String::from_utf8_lossy(&line).lines() {
        let parents: Vec<&str> = record.split_whitespace().skip(1).collect();
        if let [first, second] = parents[..] {
            pairs.push((first.to_string(), second.to_string()));
        }
    }
    let mut out: Vec<Needed> = Vec::new();
    for (a, b) in pairs {
        for need in needed_between(repo, &a, &b)?.unwrap_or_default() {
            if !out.iter().any(|n| n.id == need.id) {
                out.push(need);
            }
        }
    }
    Ok(out)
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

    #[test]
    fn always_is_a_boolean_and_an_own_hit_needs_the_head_side_only() {
        let table = json!({"scans": [
            {"id": "a", "always": true, "rules": ["r"], "checks": ["c"],
             "scope": [{"prefix": "s/"}]},
            {"id": "b", "rules": ["r"], "checks": ["d"], "scope": [{"prefix": "s/"}]}]})
        .to_string();
        let scans = parse_table(&table).unwrap();
        assert!(scans[0].always && !scans[1].always);
        let own = own_hits(&s(&["s/2", "s/1", ""]), &[], &scans);
        assert_eq!(own.len(), 1);
        assert_eq!(own[0]["id"], "a");
        assert_eq!(own[0]["scope"], json!(["s/1", "s/2"]));
        assert!(own_hits(&s(&["x"]), &[], &scans).is_empty());
        // a scan already in the hits across the two sides is not listed again
        let hits = judge(&s(&["r", "s/2"]), &s(&["s/1"]), &scans);
        assert!(own_hits(&s(&["r", "s/2"]), &hits, &scans).is_empty());
        let bad = json!({"scans": [{"id": "a", "always": "yes", "rules": [], "checks": [],
                                    "scope": []}]})
        .to_string();
        assert!(parse_table(&bad).is_err());
    }

    /// One scan `zz-scan` with the fields of `extra` added to a rule file and a scope.
    fn one(extra: Value) -> Result<Vec<Scan>, String> {
        let mut entry = json!({"id": "zz-scan", "rules": ["r"], "scope": [{"prefix": "s/"}]});
        for (key, value) in extra.as_object().unwrap() {
            entry[key] = value.clone();
        }
        parse_table(&json!({"scans": [entry]}).to_string())
    }

    const CARGO: &str = "cargo test --offline --locked";

    /// SPEC-20261124-repo-scans-targets-per-line REQ-1.
    #[test]
    fn test_targets_build_one_cargo_command_before_the_listed_checks() {
        let lib = format!("{CARGO} --lib state_dir");
        let scans =
            one(json!({"test_targets": ["alpha_cli", "beta_cli"], "checks": [lib]})).unwrap();
        assert_eq!(
            scans[0].checks,
            s(&[
                &format!("{CARGO} --test alpha_cli --test beta_cli"),
                &format!("{CARGO} --lib state_dir")
            ])
        );
        // the array's own order is kept: the sort is the table test's rule, not the reader's
        let unsorted = one(json!({"test_targets": ["beta_cli", "alpha_cli"]})).unwrap();
        assert_eq!(
            unsorted[0].checks,
            s(&[&format!("{CARGO} --test beta_cli --test alpha_cli")])
        );
        // the built command is what the judge hands to the gate
        assert_eq!(
            checks_for(&["zz-scan"], &scans),
            scans[0].checks,
            "checks_for reads the built commands"
        );
    }

    /// SPEC-20261124-repo-scans-targets-per-line REQ-2.
    #[test]
    fn a_scan_without_test_targets_keeps_its_checks_as_listed() {
        let plain = format!("{CARGO} --test b_cli --test a_cli");
        let listed = [plain.as_str(), "python3 -m unittest tests.x_test"];
        let scans = one(json!({"checks": listed})).unwrap();
        assert_eq!(scans[0].checks, s(&listed));
    }

    /// SPEC-20261124-repo-scans-targets-per-line REQ-3: the message names the scan and the field.
    #[test]
    fn a_malformed_test_targets_is_refused() {
        for bad in [
            json!("alpha_cli"),
            json!({"alpha_cli": 1}),
            json!([1]),
            json!([null]),
            json!([""]),
            json!(["Alpha_cli"]),
            json!(["alpha cli"]),
            json!(["alpha-cli"]),
            json!(["--lib"]),
            json!(["alpha_cli", "--lib"]),
            json!(["alpha_cli::t"]),
        ] {
            let err = one(json!({"test_targets": bad.clone(), "checks": []}))
                .expect_err(&format!("{bad} must be refused"));
            assert!(
                err.contains("zz-scan") && err.contains("test_targets"),
                "{bad}: {err}"
            );
        }
    }

    /// SPEC-20261124-repo-scans-targets-per-line REQ-3: `checks` may be left out when
    /// `test_targets` is there, and a scan with neither is refused.
    #[test]
    fn a_scan_with_test_targets_may_omit_checks_and_one_with_neither_is_refused() {
        let only = one(json!({"test_targets": ["alpha_cli"]})).unwrap();
        assert_eq!(only[0].checks, s(&[&format!("{CARGO} --test alpha_cli")]));
        let err = one(json!({})).expect_err("a scan with neither field is refused");
        assert!(err.contains("zz-scan"), "{err}");
        assert!(
            err.contains("test_targets") && err.contains("checks"),
            "{err}"
        );
    }
}
