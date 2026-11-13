//! Read-only label axis check (#2740, decision 49 = A of #2434): every open Issue carries
//! exactly one `type:*`, exactly one `priority:*` and one or two `area:*` labels
//! (`.github/moai/README.md`, "Labels and milestones"). This module reports the Issues that
//! break the rule; it never adds or removes a label. Labels are display only, never a lock
//! (AGENTS.md, Authority 2), so the report is a search and tidy-up aid, not an input of
//! dispatch or merge decisions.
//!
//! The axes come from this repository's `.github/moai/labels.yml` shape — `reused` (lists per
//! axis), `added` (`- name:` entries) and `not_used` (lists per group). No YAML crate (#2013
//! ②A): [`parse_axes`] reads exactly that shape and refuses anything else. The product's
//! `labels.yml` shape (`schema_version: 2`, `forge::labels`) is a separate, later scope.
//!
//! GitHub is read with GET only (open Issues, `Link` pagination); pull requests, which the
//! Issues endpoint also returns, are skipped.

use crate::forge::policy::next_link;
use crate::forge::transport::{HttpRequest, Transport};
use serde_json::{json, Value};
use std::collections::BTreeSet;

type Result<T> = std::result::Result<T, String>;

/// Pages of 100 Issues read at most.
const MAX_PAGES: usize = 30;
/// The most `area:*` labels one Issue may carry.
const AREA_MAX: usize = 2;

/// The label axes of one `labels.yml`.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct Axes {
    pub types: BTreeSet<String>,
    pub priorities: BTreeSet<String>,
    pub areas: BTreeSet<String>,
    /// Every label the file lists as in use (`reused` and `added`).
    pub known: BTreeSet<String>,
    /// Labels the file keeps documented but unused (`not_used`).
    pub not_used: BTreeSet<String>,
}

fn unquote(s: &str) -> String {
    let s = s.trim();
    let inner = s
        .strip_prefix('"')
        .and_then(|r| r.strip_suffix('"'))
        .or_else(|| s.strip_prefix('\'').and_then(|r| r.strip_suffix('\'')));
    inner.unwrap_or(s).to_string()
}

fn indent(line: &str) -> usize {
    line.len() - line.trim_start().len()
}

/// Reads this repository's `labels.yml` shape. Unknown top-level keys, an unknown `reused`
/// axis, a schema other than 1 or an empty type, priority or area axis are refused.
pub fn parse_axes(text: &str) -> Result<Axes> {
    let mut axes = Axes::default();
    let mut section = String::new();
    let mut group = String::new();
    let mut schema = None;
    for (n, raw) in text.lines().enumerate() {
        let line = raw.trim_end();
        let body = line.trim_start();
        if body.is_empty() || body.starts_with('#') {
            continue;
        }
        let at = n + 1;
        match indent(line) {
            0 => {
                let (key, value) = body
                    .split_once(':')
                    .ok_or_else(|| format!("labels.yml line {at}: expected `key:`"))?;
                section = key.to_string();
                group.clear();
                match key {
                    "schema_version" => schema = Some(value.trim().to_string()),
                    "reused" | "added" | "not_used" => {}
                    "proposed_not_created" if value.trim() == "[]" => {}
                    _ => return Err(format!("labels.yml line {at}: unknown key `{key}`")),
                }
            }
            _ if body.starts_with("- ") => {
                let item = &body[2..];
                match section.as_str() {
                    "reused" => {
                        let name = unquote(item);
                        match group.as_str() {
                            "type" => axes.types.insert(name.clone()),
                            "priority" => axes.priorities.insert(name.clone()),
                            "area" => axes.areas.insert(name.clone()),
                            "functional" => true,
                            _ => {
                                return Err(format!(
                                    "labels.yml line {at}: item outside a known reused axis"
                                ))
                            }
                        };
                        axes.known.insert(name);
                    }
                    "added" => {
                        let name = item.strip_prefix("name:").ok_or_else(|| {
                            format!("labels.yml line {at}: added entries start with `- name:`")
                        })?;
                        let name = unquote(name);
                        if let Some(set) = [
                            ("type:", &mut axes.types),
                            ("priority:", &mut axes.priorities),
                            ("area:", &mut axes.areas),
                        ]
                        .into_iter()
                        .find(|(p, _)| name.starts_with(p))
                        .map(|(_, s)| s)
                        {
                            set.insert(name.clone());
                        }
                        axes.known.insert(name);
                    }
                    "not_used" if !group.is_empty() => {
                        axes.not_used.insert(unquote(item));
                    }
                    _ => return Err(format!("labels.yml line {at}: list item out of place")),
                }
            }
            _ => {
                let (key, value) = body
                    .split_once(':')
                    .ok_or_else(|| format!("labels.yml line {at}: expected `key:`"))?;
                match section.as_str() {
                    "reused" | "not_used" if value.trim().is_empty() => group = key.to_string(),
                    // `color:` and `description:` lines of an `added` entry.
                    "added" => {}
                    _ => return Err(format!("labels.yml line {at}: unexpected `{key}:`")),
                }
            }
        }
    }
    if schema.as_deref() != Some("1") {
        return Err("labels.yml: schema_version must be 1 (this repository's shape)".into());
    }
    if axes.types.is_empty() || axes.priorities.is_empty() || axes.areas.is_empty() {
        return Err("labels.yml: the type, priority and area axes must each list a label".into());
    }
    if let Some(both) = axes.known.intersection(&axes.not_used).next() {
        return Err(format!("labels.yml: `{both}` is both in use and not_used"));
    }
    Ok(axes)
}

/// The rule violations of one Issue's labels, in a fixed order; empty when it follows the rule.
pub fn violations(labels: &[String], axes: &Axes) -> Vec<Value> {
    let count = |set: &BTreeSet<String>| labels.iter().filter(|l| set.contains(*l)).count();
    let mut out = Vec::new();
    let types = count(&axes.types);
    if types != 1 {
        out.push(json!({"kind": "type_count", "count": types}));
    }
    let priorities = count(&axes.priorities);
    if priorities != 1 {
        out.push(json!({"kind": "priority_count", "count": priorities}));
    }
    let areas = count(&axes.areas);
    if areas == 0 {
        out.push(json!({"kind": "area_missing"}));
    } else if areas > AREA_MAX {
        out.push(json!({"kind": "area_too_many", "count": areas}));
    }
    for l in labels {
        if axes.not_used.contains(l) {
            out.push(json!({"kind": "not_used", "label": l}));
        } else if !axes.known.contains(l) {
            out.push(json!({"kind": "unlisted", "label": l}));
        }
    }
    out
}

fn get_json<T: Transport>(t: &T, url: &str) -> Result<(Value, Option<String>)> {
    let resp = t.send(&HttpRequest::new("GET", url))?;
    if resp.status != 200 {
        return Err(format!("GET {url} returned HTTP {}", resp.status));
    }
    let next = next_link(resp.header("link"));
    Ok((resp.body.unwrap_or(Value::Null), next))
}

/// Reads every open Issue of `owner/repo` (GET only) and reports the ones that break the axis
/// rule, with a count per violation kind. Writes nothing anywhere.
pub fn check<T: Transport>(t: &T, owner: &str, repo: &str, axes: &Axes) -> Result<Value> {
    let mut url = Some(format!(
        "https://api.github.com/repos/{owner}/{repo}/issues?state=open&per_page=100"
    ));
    let (mut pages, mut read) = (0, 0);
    let mut issues = Vec::new();
    let mut kinds = std::collections::BTreeMap::<String, usize>::new();
    while let Some(u) = url {
        pages += 1;
        if pages > MAX_PAGES {
            return Err(format!("more than {MAX_PAGES} issue pages; refused"));
        }
        let (page, next) = get_json(t, &u)?;
        for issue in page.as_array().ok_or("issues page is not a list")? {
            if issue.get("pull_request").is_some() {
                continue;
            }
            read += 1;
            let number = issue["number"].as_u64().ok_or("issue without number")?;
            let labels: Vec<String> = issue["labels"]
                .as_array()
                .ok_or("issue without labels list")?
                .iter()
                .filter_map(|l| l["name"].as_str().map(str::to_string))
                .collect();
            let found = violations(&labels, axes);
            if found.is_empty() {
                continue;
            }
            for v in &found {
                *kinds
                    .entry(v["kind"].as_str().unwrap_or_default().to_string())
                    .or_default() += 1;
            }
            issues.push(json!({"number": number, "labels": labels, "violations": found}));
        }
        url = next;
    }
    issues.sort_by_key(|i| std::cmp::Reverse(i["number"].as_u64()));
    Ok(json!({
        "repo": format!("{owner}/{repo}"),
        "open_issues_read": read,
        "violating": issues.len(),
        "by_kind": kinds,
        "issues": issues,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn axes() -> Axes {
        parse_axes(include_str!("../.github/moai/labels.yml")).unwrap()
    }

    fn v(labels: &[&str]) -> Vec<String> {
        let labels: Vec<String> = labels.iter().map(|s| s.to_string()).collect();
        violations(&labels, &axes())
            .iter()
            .map(|v| match (v["count"].as_u64(), v["label"].as_str()) {
                (Some(n), _) => format!("{}:{n}", v["kind"].as_str().unwrap()),
                (_, Some(l)) => format!("{}:{l}", v["kind"].as_str().unwrap()),
                _ => v["kind"].as_str().unwrap().to_string(),
            })
            .collect()
    }

    #[test]
    fn this_repository_file_parses_into_its_axes() {
        let a = axes();
        assert!(a.types.contains("type:feature") && a.types.contains("type:design"));
        assert!(a.priorities.contains("priority:P0") && a.priorities.len() == 4);
        assert!(a.areas.contains("area:ci") && a.areas.contains("area:statusline"));
        assert!(a.known.contains("host:claude") && a.known.contains("needs:decision"));
        assert!(a.not_used.contains("type:security") && a.not_used.contains("priority:P4"));
        assert!(!a.types.contains("type:security"));
    }

    #[test]
    fn each_violation_kind_is_caught_and_a_correct_issue_has_none() {
        assert!(v(&["type:fix", "priority:P1", "area:cli"]).is_empty());
        assert!(v(&[
            "type:fix",
            "priority:P1",
            "area:cli",
            "host:codex",
            "needs:decision"
        ])
        .is_empty());
        assert_eq!(v(&["priority:P1", "area:cli"]), ["type_count:0"]);
        assert_eq!(
            v(&["type:fix", "type:docs", "priority:P1", "area:cli"]),
            ["type_count:2"]
        );
        assert_eq!(v(&["type:fix", "area:cli"]), ["priority_count:0"]);
        assert_eq!(
            v(&["type:fix", "priority:P1", "priority:P2", "area:cli"]),
            ["priority_count:2"]
        );
        assert_eq!(v(&["type:fix", "priority:P1"]), ["area_missing"]);
        assert_eq!(
            v(&["type:security", "priority:P1", "area:cli"]),
            ["type_count:0", "not_used:type:security"]
        );
        assert_eq!(
            v(&["type:fix", "priority:P1", "area:cli", "shiny"]),
            ["unlisted:shiny"]
        );
    }

    #[test]
    fn two_areas_pass_and_three_are_too_many() {
        assert!(v(&["type:fix", "priority:P1", "area:cli", "area:ci"]).is_empty());
        assert_eq!(
            v(&[
                "type:fix",
                "priority:P1",
                "area:cli",
                "area:ci",
                "area:hooks"
            ]),
            ["area_too_many:3"]
        );
    }

    #[test]
    fn the_parser_refuses_other_shapes() {
        let good = include_str!("../.github/moai/labels.yml");
        for (bad, why) in [
            (good.replace("schema_version: 1", "schema_version: 2"), "schema_version"),
            (format!("{good}\nlabels:\n  - name: x\n"), "unknown key"),
            (good.replace("  functional:", "  colours:"), "known reused axis"),
            (good.replace("  - name: host:claude", "  - host:claude"), "start with `- name:`"),
            (
                good.replace("  priority:\n    - priority:P0\n    - priority:P1\n    - priority:P2\n    - priority:P3\n", "  priority:\n"),
                "must each list",
            ),
            (format!("{good}    - type:fix\n"), "both in use and not_used"),
        ] {
            let e = parse_axes(&bad).unwrap_err();
            assert!(e.contains(why), "{why}: {e}");
        }
    }
}
