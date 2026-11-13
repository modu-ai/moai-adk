//! GitHub TODO reader and polling adapter (07-A #2008).
//!
//! `forge-read` reads the active Issues and Issue comments of one repository with conditional
//! GETs, a `since` cursor and `Link` pagination, serially, and returns stable references
//! `(provider, host, project{owner, repo, repository_id}, issue{id, node_id, number})`. Poll
//! state lives in `<project>/.moai-cli/forge/poll-state.json` (REQ-FORGE-009). The read path
//! sends GET only; the write helpers in [`write`] are library functions for 07-C.
//!
//! The CLI wraps the result in the 01-C envelope (`crate::contract`); errors become messages
//! whose kind `contract::classify` derives, so each message here carries its kind's phrase.
pub mod audit_status;
pub mod gitlab;
pub mod labels;
pub mod links;
pub mod outbox;
pub mod policy;
pub mod report;
pub mod transport;
pub mod write;

use policy::{format_utc, next_link, parse_utc, Clock, Verdict};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{
    cell::RefCell,
    collections::BTreeMap,
    fs::{self, OpenOptions},
    io::Write,
    path::{Path, PathBuf},
};
use transport::{HttpRequest, HttpResponse, Transport};

pub const SCHEMA_VERSION: u32 = 1;
pub const API: &str = "https://api.github.com";
const POLL_STATE: &str = "forge/poll-state.json";

/// A failure of the adapter. `kind` is a `contract::Kind` name, and `contract::classify` on
/// `message` yields the same kind (tests check every constructor), so the CLI envelope and the
/// library agree. Besides the 00-D vocabulary this adapter uses two kinds registered in
/// `contract`: `blocked` (authentication, permission, or a rate limit that outlived the retry
/// budget) and `transport` (gh missing, no HTTP response, or an unexpected status).
/// `retryable` is finer than the envelope's per-kind value: an exhausted rate limit is
/// `blocked` but retryable here, because waiting until the reset will succeed.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ForgeError {
    pub kind: &'static str,
    pub message: String,
    pub retryable: bool,
}

impl ForgeError {
    fn of(kind: &'static str, message: impl Into<String>) -> Self {
        ForgeError {
            kind,
            message: message.into(),
            retryable: false,
        }
    }
    fn io(message: impl Into<String>) -> Self {
        Self::of("io", message)
    }
    fn blocked(message: impl Into<String>, retryable: bool) -> Self {
        ForgeError {
            kind: "blocked",
            message: format!("forge blocked: {}", message.into()),
            retryable,
        }
    }
    fn invalid(message: impl Into<String>) -> Self {
        Self::of("invalid_input", message)
    }
    fn transport(message: impl Into<String>) -> Self {
        ForgeError {
            kind: "transport",
            message: format!("gh transport: {}", transport::redact(&message.into())),
            retryable: true,
        }
    }
}

impl From<ForgeError> for String {
    fn from(e: ForgeError) -> String {
        e.message
    }
}

/// Settings with interim values (leader, 2026-11-08). `poll_interval_s` is unmeasured (P11),
/// so no default is assumed; GitHub's `x-poll-interval` always wins when present.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Config {
    #[serde(default)]
    pub poll_interval_s: Option<u64>,
    #[serde(default = "default_overlap")]
    pub overlap_s: i64,
    #[serde(default = "default_retries")]
    pub max_retries: u32,
    #[serde(default = "default_max_pages")]
    pub max_pages: u32,
}
fn default_overlap() -> i64 {
    60
}
fn default_retries() -> u32 {
    3
}
fn default_max_pages() -> u32 {
    50
}
impl Default for Config {
    fn default() -> Self {
        Config {
            poll_interval_s: None,
            overlap_s: default_overlap(),
            max_retries: default_retries(),
            max_pages: default_max_pages(),
        }
    }
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ReadRequest {
    pub schema_version: u32,
    pub owner: String,
    pub repo: String,
    /// `open` (default), `closed` or `all`.
    #[serde(default = "default_state")]
    pub state: String,
    #[serde(default = "default_true")]
    pub include_comments: bool,
    #[serde(default)]
    pub config: Config,
}
fn default_state() -> String {
    "open".into()
}
fn default_true() -> bool {
    true
}

/// Identifies one issue for writes; `number` alone is never a key (REQ-FORGE-001).
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct IssueKey {
    pub owner: String,
    pub repo: String,
    pub repository_id: i64,
    pub id: i64,
    pub number: i64,
}

/// One cached response: its validator and the projected items it produced.
#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq)]
pub struct CachedPage {
    pub etag: String,
    pub value: Value,
}

#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq)]
pub struct PollState {
    pub schema_version: u32,
    pub provider: String,
    pub host: String,
    pub owner: String,
    pub repo: String,
    pub repository_id: Option<i64>,
    pub last_success_at: Option<String>,
    pub issues_since: Option<String>,
    pub comments_since: Option<String>,
    pub rate_limit: Option<Value>,
    /// URL → validator and projection. Only URLs used by the last successful run are kept.
    pub pages: BTreeMap<String, CachedPage>,
}

/// One serial client (REQ-FORGE-004): every request goes through `send_with_backoff`.
pub struct Client<'a, T: Transport, C: Clock> {
    pub transport: &'a T,
    pub clock: &'a C,
    pub config: Config,
    cache: BTreeMap<String, CachedPage>,
    used: RefCell<BTreeMap<String, CachedPage>>,
    log: RefCell<Vec<Value>>,
    last_rate: RefCell<Option<Value>>,
    poll_interval_header: RefCell<Option<u64>>,
}

impl<'a, T: Transport, C: Clock> Client<'a, T, C> {
    pub fn new(
        transport: &'a T,
        clock: &'a C,
        config: Config,
        cache: BTreeMap<String, CachedPage>,
    ) -> Self {
        Client {
            transport,
            clock,
            config,
            cache,
            used: RefCell::new(BTreeMap::new()),
            log: RefCell::new(Vec::new()),
            last_rate: RefCell::new(None),
            poll_interval_header: RefCell::new(None),
        }
    }

    fn note_headers(&self, resp: &HttpResponse) {
        if let Some(limit) = resp.header("x-ratelimit-limit") {
            let n = |k: &str| resp.header(k).and_then(|v| v.parse::<i64>().ok());
            *self.last_rate.borrow_mut() = Some(json!({
                "limit": limit.parse::<i64>().ok(),
                "remaining": n("x-ratelimit-remaining"),
                "reset": n("x-ratelimit-reset").map(format_utc),
                "resource": resp.header("x-ratelimit-resource"),
            }));
        }
        if let Some(p) = resp.header("x-poll-interval").and_then(|v| v.parse().ok()) {
            *self.poll_interval_header.borrow_mut() = Some(p);
        }
    }

    /// Sends one request, waiting as GitHub's headers say (REQ-FORGE-003). After
    /// `max_retries` waits the result is `blocked`.
    pub fn send_with_backoff(&self, req: &HttpRequest) -> Result<HttpResponse, ForgeError> {
        let mut attempt = 0;
        loop {
            let resp = self.transport.send(req).map_err(ForgeError::transport)?;
            self.note_headers(&resp);
            self.log
                .borrow_mut()
                .push(json!({"method": req.method, "url": req.url, "status": resp.status}));
            match policy::judge(&resp, attempt, self.clock.now()) {
                Verdict::Use => return Ok(resp),
                Verdict::Blocked(why) => return Err(ForgeError::blocked(why, false)),
                Verdict::Retry(secs) if attempt < self.config.max_retries => {
                    self.clock.sleep(secs);
                    attempt += 1;
                }
                Verdict::Retry(_) => {
                    return Err(ForgeError::blocked(
                        format!("rate limited after {} retries", self.config.max_retries),
                        true,
                    ))
                }
            }
        }
    }

    /// Conditional GET (REQ-FORGE-002). Returns the response; a 304 is answered from `cache`.
    pub fn get(
        &self,
        url: &str,
        project: Option<&dyn Fn(&HttpResponse) -> Value>,
    ) -> Result<HttpResponse, ForgeError> {
        let cached = self.cache.get(url).cloned();
        let mut headers = BTreeMap::new();
        if let Some(c) = &cached {
            headers.insert("If-None-Match".to_string(), c.etag.clone());
        }
        let mut req = HttpRequest::new("GET", url);
        req.headers = headers;
        let resp = self.send_with_backoff(&req)?;
        match (resp.status, cached) {
            (304, Some(c)) => {
                self.used.borrow_mut().insert(url.into(), c);
                Ok(resp)
            }
            (304, None) => Err(ForgeError::transport("304 without a cached response")),
            (200..=299, _) => {
                if let (Some(etag), Some(project)) = (resp.header("etag"), project) {
                    let value = project(&resp);
                    self.used.borrow_mut().insert(
                        url.into(),
                        CachedPage {
                            etag: etag.into(),
                            value,
                        },
                    );
                }
                Ok(resp)
            }
            (404, _) => Err(ForgeError::of(
                "not_found",
                "forge resource not found or not visible to this account (404)",
            )),
            (s, _) => Err(ForgeError::transport(format!("GET returned HTTP {s}"))),
        }
    }

    /// Follows `Link: rel="next"` exactly; returns (projected items, all-304, page count).
    fn pages(
        &self,
        first: String,
        project: &dyn Fn(&HttpResponse) -> Value,
    ) -> Result<(Vec<Value>, bool), ForgeError> {
        let mut url = Some(first);
        let mut items = Vec::new();
        let mut all_cached = true;
        let mut n = 0;
        while let Some(u) = url {
            n += 1;
            if n > self.config.max_pages {
                return Err(ForgeError::invalid(format!(
                    "page limit exceeded (config.max_pages {})",
                    self.config.max_pages
                )));
            }
            let resp = self.get(&u, Some(project))?;
            let page = if resp.status == 304 {
                self.used.borrow()[&u].value.clone()
            } else {
                all_cached = false;
                self.used
                    .borrow()
                    .get(&u)
                    .map(|c| c.value.clone())
                    .unwrap_or_else(|| project(&resp))
            };
            items.extend(page["items"].as_array().cloned().unwrap_or_default());
            url = page["next"].as_str().map(str::to_string);
        }
        Ok((items, all_cached))
    }
}

fn valid_name(s: &str) -> bool {
    !s.is_empty()
        && s.len() <= 100
        && s.bytes()
            .all(|b| b.is_ascii_alphanumeric() || b"-_.".contains(&b))
        && s != "."
        && s != ".."
}

fn label_names(v: &Value) -> Vec<Value> {
    v.as_array()
        .map(|a| {
            a.iter()
                .filter_map(|l| l["name"].as_str())
                .map(|s| json!(s))
                .collect()
        })
        .unwrap_or_default()
}

/// Reads one repository. Pure over `transport` and `clock`; the caller persists the state.
pub fn read<T: Transport, C: Clock>(
    req: &ReadRequest,
    prior: Option<PollState>,
    transport: &T,
    clock: &C,
) -> Result<(Value, PollState), ForgeError> {
    if req.schema_version != SCHEMA_VERSION {
        return Err(ForgeError::of(
            "unsupported_protocol_version",
            "unsupported request schema_version",
        ));
    }
    if !valid_name(&req.owner) || !valid_name(&req.repo) {
        return Err(ForgeError::invalid("owner and repo must be GitHub names"));
    }
    if !["open", "closed", "all"].contains(&req.state.as_str()) {
        return Err(ForgeError::invalid("state must be open, closed or all"));
    }
    if req.config.overlap_s < 0 || req.config.max_pages == 0 {
        return Err(ForgeError::invalid(
            "overlap_s must be >= 0 and max_pages > 0",
        ));
    }
    let prior =
        prior.filter(|p| p.owner == req.owner && p.repo == req.repo && p.provider == "github");
    let cache = prior.as_ref().map(|p| p.pages.clone()).unwrap_or_default();
    let client = Client::new(transport, clock, req.config.clone(), cache);
    let observed_at = format_utc(clock.now());

    let repo_url = format!("{API}/repos/{}/{}", req.owner, req.repo);
    let repo_project = |r: &HttpResponse| {
        let b = r.body.clone().unwrap_or(Value::Null);
        json!({"repository_id": b["id"], "full_name": b["full_name"]})
    };
    client.get(&repo_url, Some(&repo_project))?;
    let repo = client
        .used
        .borrow()
        .get(&repo_url)
        .map(|c| c.value.clone())
        .unwrap_or(Value::Null);
    let repository_id = repo["repository_id"]
        .as_i64()
        .ok_or_else(|| ForgeError::transport("repository response carried no id"))?;
    let project = json!({"owner": req.owner, "repo": req.repo, "repository_id": repository_id});

    let since = |cursor: &Option<String>| {
        cursor
            .as_ref()
            .map(|s| format!("&since={s}"))
            .unwrap_or_default()
    };
    let issues_url = format!(
        "{API}/repos/{}/{}/issues?state={}&sort=updated&direction=asc&per_page=100{}",
        req.owner,
        req.repo,
        req.state,
        since(&prior.as_ref().and_then(|p| p.issues_since.clone()))
    );
    let issue_project = |r: &HttpResponse| {
        let etag = r.header("etag");
        let items: Vec<Value> = r
            .body
            .as_ref()
            .and_then(Value::as_array)
            .map(|a| {
                a.iter()
                    .map(|i| {
                        let kind = if i.get("pull_request").is_some() { "pull_request" } else { "issue" };
                        json!({
                            "kind": kind,
                            "ref": {
                                "provider": "github", "host": "github.com", "project": project,
                                "issue": {"id": i["id"], "node_id": i["node_id"], "number": i["number"]},
                                "updated_at": i["updated_at"], "etag": etag,
                            },
                            "title": i["title"], "state": i["state"], "state_reason": i["state_reason"],
                            "labels": label_names(&i["labels"]),
                        })
                    })
                    .collect()
            })
            .unwrap_or_default();
        json!({"items": items, "next": next_link(r.header("link"))})
    };
    let (issue_items, issues_cached) = client.pages(issues_url, &issue_project)?;

    let (comment_items, comments_cached) = if req.include_comments {
        let url = format!(
            "{API}/repos/{}/{}/issues/comments?sort=updated&direction=asc&per_page=100{}",
            req.owner,
            req.repo,
            since(&prior.as_ref().and_then(|p| p.comments_since.clone()))
        );
        let comment_project = |r: &HttpResponse| {
            let items: Vec<Value> = r
                .body
                .as_ref()
                .and_then(Value::as_array)
                .map(|a| {
                    a.iter()
                        .map(|c| {
                            let number = c["issue_url"]
                                .as_str()
                                .and_then(|u| u.rsplit('/').next())
                                .and_then(|n| n.parse::<i64>().ok());
                            json!({"kind": "comment", "id": c["id"], "node_id": c["node_id"],
                                   "issue_number": number, "updated_at": c["updated_at"]})
                        })
                        .collect()
                })
                .unwrap_or_default();
            json!({"items": items, "next": next_link(r.header("link"))})
        };
        client.pages(url, &comment_project)?
    } else {
        (Vec::new(), true)
    };

    // Overlap reads can repeat an item; keep the newest copy per stable id.
    let dedupe = |items: Vec<Value>,
                  id: &dyn Fn(&Value) -> Option<i64>,
                  ts: &dyn Fn(&Value) -> Option<String>| {
        let mut by_id: BTreeMap<i64, Value> = BTreeMap::new();
        for it in items {
            if let Some(k) = id(&it) {
                let newer = by_id.get(&k).map(|old| ts(&it) >= ts(old)).unwrap_or(true);
                if newer {
                    by_id.insert(k, it);
                }
            }
        }
        by_id.into_values().collect::<Vec<_>>()
    };
    let issues = dedupe(issue_items, &|v| v["ref"]["issue"]["id"].as_i64(), &|v| {
        v["ref"]["updated_at"].as_str().map(str::to_string)
    });
    let comments = dedupe(comment_items, &|v| v["id"].as_i64(), &|v| {
        v["updated_at"].as_str().map(str::to_string)
    });

    // The next cursor comes from the data, not the clock, so an unchanged repository asks the
    // same URL again and is answered with 304. GitHub's `since` is "after"; the overlap window
    // (GitLab's "on or after" adapter keeps the same rule) covers boundary updates.
    let cursor = |items: &[Value], ts: &dyn Fn(&Value) -> Option<&str>, old: Option<String>| {
        items
            .iter()
            .filter_map(|v| ts(v).and_then(parse_utc))
            .max()
            .map(|t| format_utc(t - req.config.overlap_s))
            .or(old)
    };
    let prior_issues_since = prior.as_ref().and_then(|p| p.issues_since.clone());
    let prior_comments_since = prior.as_ref().and_then(|p| p.comments_since.clone());
    let state = PollState {
        schema_version: SCHEMA_VERSION,
        provider: "github".into(),
        host: "github.com".into(),
        owner: req.owner.clone(),
        repo: req.repo.clone(),
        repository_id: Some(repository_id),
        last_success_at: Some(observed_at.clone()),
        issues_since: cursor(
            &issues,
            &|v| v["ref"]["updated_at"].as_str(),
            prior_issues_since,
        ),
        comments_since: if req.include_comments {
            cursor(
                &comments,
                &|v| v["updated_at"].as_str(),
                prior_comments_since,
            )
        } else {
            prior_comments_since
        },
        rate_limit: client.last_rate.borrow().clone(),
        pages: client.used.borrow().clone(),
    };
    let next_poll_after = match (
        *client.poll_interval_header.borrow(),
        req.config.poll_interval_s,
    ) {
        (Some(h), Some(c)) => Some(h.max(c)),
        (a, b) => a.or(b),
    };
    let result = json!({
        "schema_version": SCHEMA_VERSION,
        "observed_at": observed_at,
        "last_success_at": observed_at,
        "previous_success_at": prior.as_ref().and_then(|p| p.last_success_at.clone()),
        "project": project,
        "cache_hit": issues_cached && comments_cached,
        "rate_limit": client.last_rate.borrow().clone(),
        "next_poll_after_s": next_poll_after,
        "config": {
            "overlap_s": req.config.overlap_s, "max_retries": req.config.max_retries,
            "max_pages": req.config.max_pages, "poll_interval_s": req.config.poll_interval_s,
            "note": "overlap_s and max_retries are interim values (2026-11-08); poll_interval_s is unmeasured and x-poll-interval wins",
        },
        "items": issues,
        "comments": comments,
        "requests": client.log.borrow().clone(),
    });
    Ok((result, state))
}

/// Stable identity of a read item: `(provider, host, repository_id, issue.id)`. Owner and repo
/// names are display data — a renamed repository keeps its id — and `number` is never part of
/// the key, because numbers repeat across repositories (REQ-FORGE-001).
pub fn identity(item: &Value) -> Option<(String, String, i64, i64)> {
    let r = &item["ref"];
    Some((
        r["provider"].as_str()?.to_string(),
        r["host"].as_str()?.to_string(),
        r["project"]["repository_id"].as_i64()?,
        r["issue"]["id"].as_i64()?,
    ))
}

fn state_path(project: &Path) -> Result<PathBuf, ForgeError> {
    let meta = fs::symlink_metadata(project)
        .map_err(|_| ForgeError::of("not_found", "project directory unavailable"))?;
    if !meta.is_dir() {
        return Err(ForgeError::invalid("project must be a directory"));
    }
    let mut dir = project.to_path_buf();
    for part in [crate::PROJECT_STATE_DIR, "forge"] {
        dir.push(part);
        match fs::symlink_metadata(&dir) {
            Ok(m) if m.file_type().is_symlink() || !m.is_dir() => {
                return Err(ForgeError::invalid(
                    "poll state directory is a symlink or not a directory; refused",
                ))
            }
            Ok(_) => {}
            Err(_) => fs::create_dir(&dir)
                .map_err(|_| ForgeError::io("poll state directory unavailable"))?,
        }
    }
    Ok(project.join(crate::PROJECT_STATE_DIR).join(POLL_STATE))
}

/// Loads the poll state; an unreadable or foreign file is ignored (the next success replaces it).
pub fn load_state(project: &Path) -> Option<PollState> {
    let path = project.join(crate::PROJECT_STATE_DIR).join(POLL_STATE);
    let bytes = fs::read(path).ok()?;
    serde_json::from_slice::<PollState>(&bytes)
        .ok()
        .filter(|s| s.schema_version == SCHEMA_VERSION)
}

/// Temporary file + rename in the same directory (REQ-FORGE-009). Writes nothing under `.moai/`.
pub fn save_state(project: &Path, state: &PollState) -> Result<(), ForgeError> {
    let path = state_path(project)?;
    let tmp = path.with_extension(format!("json.tmp-{}", std::process::id()));
    let bytes = serde_json::to_vec_pretty(state)
        .map_err(|_| ForgeError::of("internal", "encode failed"))?;
    let mut f = OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&tmp)
        .map_err(|_| ForgeError::io("poll state write failed"))?;
    f.write_all(&bytes)
        .and_then(|_| f.sync_all())
        .map_err(|_| ForgeError::io("poll state write failed"))?;
    fs::rename(&tmp, &path).map_err(|_| {
        let _ = fs::remove_file(&tmp);
        ForgeError::io("poll state replace failed")
    })
}

/// CLI entry: poll, persist on success, and return the result. `main` wraps it in the 01-C
/// envelope (`forge.read`); an error is its message, which `contract::classify` maps to the
/// same kind as `ForgeError::kind`.
pub fn forge_read_cli<T: Transport, C: Clock>(
    project: &Path,
    req: ReadRequest,
    transport: &T,
    clock: &C,
) -> Result<Value, String> {
    let prior = load_state(project);
    let (result, state) = read(&req, prior, transport, clock)?;
    save_state(project, &state)?;
    Ok(result)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn every_error_kind_matches_the_contract_classifier() {
        let errors = [
            ForgeError::blocked("authentication failed (401)", false),
            ForgeError::blocked("rate limited after 3 retries", true),
            ForgeError::transport("cannot start gh: NotFound"),
            ForgeError::transport("refused: only https://api.github.com/ URLs are requested"),
            ForgeError::of(
                "not_found",
                "forge resource not found or not visible to this account (404)",
            ),
            ForgeError::of(
                "unsupported_protocol_version",
                "unsupported request schema_version",
            ),
            ForgeError::of("not_found", "project directory unavailable"),
            ForgeError::invalid("project must be a directory"),
            ForgeError::invalid("page limit exceeded (config.max_pages 50)"),
            ForgeError::invalid("poll state directory is a symlink or not a directory; refused"),
            ForgeError::io("poll state directory unavailable"),
            ForgeError::io("poll state write failed"),
            ForgeError::io("poll state replace failed"),
            ForgeError::of("internal", "encode failed"),
            // links (#2172)
            ForgeError::invalid("owner and repo must be GitHub names"),
            ForgeError::invalid("issue numbers must be positive"),
            ForgeError::invalid("issue links must be to other issues"),
            ForgeError::invalid("blocked_by must be at most 50 issues (one relationship type)"),
            ForgeError::invalid("blocked_by must be free of repeated issues"),
            // audit status (#2173)
            ForgeError::of(
                "permission_reference_missing",
                "permission reference required",
            ),
            ForgeError::invalid("sha must be a 40 lowercase hex commit"),
            ForgeError::invalid("state must be success, failure, error or pending"),
            ForgeError::invalid("target_url must be an https://github.com/ link"),
            ForgeError::invalid("description must be one line without control characters"),
            ForgeError::invalid(
                "description must be at most 140 characters with the source marker",
            ),
            ForgeError::transport("status POST returned HTTP 422"),
        ];
        for e in errors {
            assert_eq!(
                crate::contract::classify(&e.message).as_str(),
                e.kind,
                "{}",
                e.message
            );
        }
    }
}
