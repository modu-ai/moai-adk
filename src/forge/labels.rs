//! Label sync for a project (#2434 decision 40): `.github/moai/labels.yml` is the source of
//! truth and `moai-cli` brings the repository's labels in line with it.
//!
//! `labels-plan` reads the file and the repository's labels (GET, pages of 100) and plans,
//! per listed label, `create`, `update` (colour or description differ) or `unchanged`. Names
//! compare case-insensitively, as GitHub does; colours compare without `#` and case; a null
//! remote description equals an empty one. A remote label the file does not list is reported
//! as `unlisted` and never deleted: a project keeps its own labels (decision 42, `area:`).
//!
//! `labels-apply` needs the plan digest and a permission reference, re-reads the file and the
//! remote list and refuses a different plan (`stale`), then sends POST/PATCH only, one second
//! apart, through the adapter's backoff. A failed write stops the run and names what was
//! already applied; a re-plan after it plans only what is left.
//!
//! No YAML crate (#2013 ②A): [`parse_labels_yml`] reads exactly the product shape —
//! `schema_version: 2` and one `labels:` list of `name`, `color`, `description` — and refuses
//! anything else rather than skipping it.
use super::policy::{next_link, Clock};
use super::transport::{HttpRequest, Transport};
use super::{Client, Config, ForgeError, API, SCHEMA_VERSION};
use crate::spec::{Canonical, SortedKeys};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::path::Path;

/// The label list a project keeps, relative to the project root.
pub const LABELS_FILE: &str = ".github/moai/labels.yml";
/// GitHub's limits on a label (REST "Create a label"): name 50, description 100 characters.
const NAME_MAX: usize = 50;
const DESCRIPTION_MAX: usize = 100;
const PAGE: &str = "per_page=100";

#[derive(Clone, Debug, Deserialize, Serialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct LabelsRequest {
    pub schema_version: u32,
    pub owner: String,
    pub repo: String,
}

/// One label of the file.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct LabelSpec {
    pub name: String,
    /// Six lower-case hex digits, no `#`.
    pub color: String,
    pub description: String,
}

/// One label as GitHub returns it; other fields are ignored.
#[derive(Clone, Debug, Deserialize)]
pub struct RemoteLabel {
    pub name: String,
    pub color: String,
    #[serde(default)]
    pub description: Option<String>,
}

#[derive(Clone, Debug, Deserialize, Serialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct LabelChange {
    pub name: String,
    /// create | update | unchanged
    pub action: String,
    pub color: String,
    pub description: String,
    /// The remote name when it exists (its case may differ from the file's).
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub remote_name: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub remote_color: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub remote_description: Option<String>,
}

#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
pub struct LabelsPlan {
    pub schema_version: u32,
    pub request: LabelsRequest,
    pub file_sha256: String,
    pub changes: Vec<LabelChange>,
    /// Remote labels the file does not list; never deleted.
    pub unlisted: Vec<String>,
    pub plan_digest: String,
}

fn invalid(message: impl Into<String>) -> ForgeError {
    ForgeError::invalid(format!("invalid labels file: {}", message.into()))
}

/// A YAML scalar of the restricted shape: plain, or single-quoted with `''` for a quote.
fn scalar(raw: &str, line: usize) -> Result<String, ForgeError> {
    let raw = raw.trim_end();
    if let Some(inner) = raw.strip_prefix('\'') {
        let inner = inner
            .strip_suffix('\'')
            .ok_or_else(|| invalid(format!("line {line}: unterminated quote")))?;
        if inner.replace("''", "").contains('\'') {
            return Err(invalid(format!("line {line}: stray quote")));
        }
        return Ok(inner.replace("''", "'"));
    }
    let first = raw.chars().next();
    if matches!(
        first,
        Some('"' | '{' | '[' | '&' | '*' | '!' | '|' | '>' | '#' | '%' | '@' | '`')
    ) || raw.contains(": ")
        || raw.contains(" #")
        || raw.ends_with(':')
    {
        return Err(invalid(format!("line {line}: unsupported scalar")));
    }
    Ok(raw.to_string())
}

/// Parses the product label list; any other shape is an error.
pub fn parse_labels_yml(text: &str) -> Result<Vec<LabelSpec>, ForgeError> {
    if text.contains('\t') || text.contains('\r') {
        return Err(invalid("tabs and carriage returns are not allowed"));
    }
    let mut schema = false;
    let mut in_list = false;
    let mut items: Vec<BTreeMap<String, String>> = Vec::new();
    for (i, line) in text.lines().enumerate() {
        let n = i + 1;
        if line.trim().is_empty() || line.starts_with('#') {
            continue;
        }
        if let Some(rest) = line.strip_prefix("- name: ") {
            if !in_list {
                return Err(invalid(format!("line {n}: item outside `labels:`")));
            }
            items.push(BTreeMap::from([("name".to_string(), scalar(rest, n)?)]));
        } else if let Some(rest) = line.strip_prefix("  ") {
            let item = items
                .last_mut()
                .filter(|_| in_list)
                .ok_or_else(|| invalid(format!("line {n}: field outside an item")))?;
            let (key, value) = rest
                .split_once(": ")
                .ok_or_else(|| invalid(format!("line {n}: expected `key: value`")))?;
            if !matches!(key, "color" | "description") || item.contains_key(key) {
                return Err(invalid(format!("line {n}: unexpected or repeated `{key}`")));
            }
            item.insert(key.to_string(), scalar(value, n)?);
        } else if line == "schema_version: 2" && !schema {
            schema = true;
            in_list = false;
        } else if line == "labels:" && items.is_empty() && !in_list {
            in_list = true;
        } else {
            return Err(invalid(format!(
                "line {n}: only `schema_version: 2` and one `labels:` list are accepted"
            )));
        }
    }
    if !schema || items.is_empty() {
        return Err(invalid(
            "needs `schema_version: 2` and a non-empty `labels:` list",
        ));
    }
    let mut seen = std::collections::BTreeSet::new();
    let mut out = Vec::new();
    for item in items {
        let (Some(name), Some(color), Some(description)) =
            (item.get("name"), item.get("color"), item.get("description"))
        else {
            return Err(invalid("every label needs name, color and description"));
        };
        if name.is_empty() || name.chars().count() > NAME_MAX {
            return Err(invalid(format!("label name length: {name}")));
        }
        if !seen.insert(name.to_lowercase()) {
            return Err(invalid(format!(
                "duplicate label (names ignore case): {name}"
            )));
        }
        if color.len() != 6 || !color.bytes().all(|b| b.is_ascii_hexdigit()) {
            return Err(invalid(format!("color of {name} must be six hex digits")));
        }
        if description.chars().count() > DESCRIPTION_MAX {
            return Err(invalid(format!(
                "description of {name} is over 100 characters"
            )));
        }
        out.push(LabelSpec {
            name: name.clone(),
            color: color.to_ascii_lowercase(),
            description: description.clone(),
        });
    }
    Ok(out)
}

fn plan_digest(p: &LabelsPlan) -> String {
    let mut v = serde_json::to_value(p).expect("plan serializes");
    v["plan_digest"] = Value::Null;
    crate::digest(&SortedKeys.bytes(&v))
}

/// The plan for `specs` against `remote`; a pure function of its inputs.
pub fn plan_labels(req: &LabelsRequest, specs: &[LabelSpec], remote: &[RemoteLabel]) -> LabelsPlan {
    plan_with_file(req, "", specs, remote)
}

fn plan_with_file(
    req: &LabelsRequest,
    file_sha256: &str,
    specs: &[LabelSpec],
    remote: &[RemoteLabel],
) -> LabelsPlan {
    let by_name: BTreeMap<String, &RemoteLabel> =
        remote.iter().map(|r| (r.name.to_lowercase(), r)).collect();
    let changes = specs
        .iter()
        .map(|s| {
            let found = by_name.get(&s.name.to_lowercase());
            let action = match found {
                None => "create",
                Some(r)
                    if r.color.trim_start_matches('#').to_ascii_lowercase() == s.color
                        && r.description.as_deref().unwrap_or("") == s.description =>
                {
                    "unchanged"
                }
                Some(_) => "update",
            };
            LabelChange {
                name: s.name.clone(),
                action: action.into(),
                color: s.color.clone(),
                description: s.description.clone(),
                remote_name: found.map(|r| r.name.clone()),
                remote_color: found.map(|r| r.color.clone()),
                remote_description: found.and_then(|r| r.description.clone()),
            }
        })
        .collect();
    let listed: std::collections::BTreeSet<String> =
        specs.iter().map(|s| s.name.to_lowercase()).collect();
    let mut unlisted: Vec<String> = remote
        .iter()
        .filter(|r| !listed.contains(&r.name.to_lowercase()))
        .map(|r| r.name.clone())
        .collect();
    unlisted.sort();
    let mut plan = LabelsPlan {
        schema_version: SCHEMA_VERSION,
        request: req.clone(),
        file_sha256: file_sha256.to_string(),
        changes,
        unlisted,
        plan_digest: String::new(),
    };
    plan.plan_digest = plan_digest(&plan);
    plan
}

fn valid_name(s: &str) -> bool {
    !s.is_empty()
        && s.len() <= 100
        && s.bytes()
            .all(|b| b.is_ascii_alphanumeric() || b"-_.".contains(&b))
        && s != "."
        && s != ".."
}

fn validate(req: &LabelsRequest) -> Result<(), ForgeError> {
    if req.schema_version != SCHEMA_VERSION {
        return Err(ForgeError::of(
            "unsupported_protocol_version",
            "unsupported request schema_version",
        ));
    }
    if !valid_name(&req.owner) || !valid_name(&req.repo) {
        return Err(ForgeError::invalid("owner and repo must be GitHub names"));
    }
    Ok(())
}

/// Percent-encodes one URL path segment (RFC 3986 unreserved characters kept).
pub fn encode_segment(s: &str) -> String {
    s.bytes()
        .map(|b| match b {
            b'A'..=b'Z' | b'a'..=b'z' | b'0'..=b'9' | b'-' | b'_' | b'.' | b'~' => {
                (b as char).to_string()
            }
            _ => format!("%{b:02X}"),
        })
        .collect()
}

fn base(req: &LabelsRequest) -> String {
    format!("{API}/repos/{}/{}/labels", req.owner, req.repo)
}

/// Every label of the repository, following `Link` pages of 100.
pub fn read_remote<T: Transport, C: Clock>(
    client: &Client<T, C>,
    req: &LabelsRequest,
) -> Result<Vec<RemoteLabel>, ForgeError> {
    validate(req)?;
    let mut out = Vec::new();
    let mut url = Some(format!("{}?{PAGE}", base(req)));
    let mut pages = 0;
    while let Some(u) = url {
        pages += 1;
        if pages > client.config.max_pages {
            return Err(ForgeError::blocked(
                format!("more than {} label pages", client.config.max_pages),
                false,
            ));
        }
        let resp = client.send_with_backoff(&HttpRequest::new("GET", u.clone()))?;
        if resp.status != 200 {
            return Err(ForgeError::transport(format!(
                "GET labels returned HTTP {}",
                resp.status
            )));
        }
        let page: Vec<RemoteLabel> =
            serde_json::from_value(resp.body.clone().unwrap_or(Value::Null))
                .map_err(|_| ForgeError::transport("GET labels returned an unexpected body"))?;
        out.extend(page);
        url = next_link(resp.header("link"));
    }
    Ok(out)
}

fn read_file(project: &Path) -> Result<(String, Vec<LabelSpec>), ForgeError> {
    if !project.is_dir() {
        return Err(ForgeError::invalid("project must be a directory"));
    }
    let bytes = std::fs::read(project.join(LABELS_FILE))
        .map_err(|_| ForgeError::invalid(format!("{LABELS_FILE} is missing or unreadable")))?;
    let text = String::from_utf8(bytes).map_err(|_| invalid("not UTF-8"))?;
    let specs = parse_labels_yml(&text)?;
    Ok((crate::digest(text.as_bytes()), specs))
}

fn plan_now<T: Transport, C: Clock>(
    project: &Path,
    req: &LabelsRequest,
    transport: &T,
    clock: &C,
) -> Result<LabelsPlan, ForgeError> {
    validate(req)?;
    let (file_sha256, specs) = read_file(project)?;
    let client = Client::new(transport, clock, Config::default(), BTreeMap::new());
    let remote = read_remote(&client, req)?;
    Ok(plan_with_file(req, &file_sha256, &specs, &remote))
}

/// `labels-plan`: GET only.
pub fn labels_plan_cli<T: Transport, C: Clock>(
    project: &Path,
    req: LabelsRequest,
    transport: &T,
    clock: &C,
) -> Result<Value, String> {
    let plan = plan_now(project, &req, transport, clock)?;
    Ok(serde_json::to_value(plan).expect("plan serializes"))
}

/// `labels-apply`: re-plans, refuses a different plan, then writes POST/PATCH only.
pub fn labels_apply_cli<T: Transport, C: Clock>(
    project: &Path,
    plan: Value,
    confirmed: &str,
    permission: &str,
    transport: &T,
    clock: &C,
) -> Result<Value, String> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    let p: LabelsPlan = serde_json::from_value(plan)
        .map_err(|_| ForgeError::invalid("the input is not a labels plan"))?;
    let current = plan_now(project, &p.request, transport, clock)?;
    if confirmed != p.plan_digest
        || plan_digest(&p) != p.plan_digest
        || current.plan_digest != p.plan_digest
    {
        return Err("stale or altered plan; re-preview required".into());
    }
    let client = Client::new(transport, clock, Config::default(), BTreeMap::new());
    let (mut created, mut updated) = (Vec::new(), Vec::new());
    let mut last_write: Option<i64> = None;
    for c in p.changes.iter().filter(|c| c.action != "unchanged") {
        if let Some(prev) = last_write {
            let gap = clock.now() - prev;
            if gap < 1 {
                clock.sleep((1 - gap) as u64);
            }
        }
        let body = json!({"color": c.color, "description": c.description});
        let req = if c.action == "create" {
            let mut body = body;
            body["name"] = json!(c.name);
            let mut r = HttpRequest::new("POST", base(&p.request));
            r.body = Some(body);
            r
        } else {
            let remote = c.remote_name.as_deref().unwrap_or(&c.name);
            let mut r = HttpRequest::new(
                "PATCH",
                format!("{}/{}", base(&p.request), encode_segment(remote)),
            );
            r.body = Some(body);
            r
        };
        let resp = client.send_with_backoff(&req);
        last_write = Some(clock.now());
        let failed = match resp {
            Ok(r) if (200..300).contains(&r.status) => None,
            Ok(r) => Some(format!("HTTP {}", r.status)),
            Err(e) => Some(e.message),
        };
        if let Some(why) = failed {
            return Err(format!(
                "stale or failed label write at {} ({why}); applied before it: created {:?}, updated {:?}; re-preview required",
                c.name, created, updated
            ));
        }
        if c.action == "create" {
            created.push(c.name.clone());
        } else {
            updated.push(c.name.clone());
        }
    }
    Ok(json!({
        "created": created,
        "updated": updated,
        "unlisted": p.unlisted,
        "deleted": [],
        "permission_reference": permission,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn segments_keep_unreserved_bytes_and_encode_the_rest() {
        assert_eq!(encode_segment("priority:P0"), "priority%3AP0");
        assert_eq!(encode_segment("a b/é"), "a%20b%2F%C3%A9");
        assert_eq!(encode_segment("needs-decision_1.x~"), "needs-decision_1.x~");
    }
}
