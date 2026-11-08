//! Selected promotion of local TODOs to GitHub Issues (07-B #2009, P07/D80).
//!
//! Two steps, like `project-plan`/`project-apply`:
//! - `preview` reads `tasks.md` (path given by the caller; RC1 interim decision: no default
//!   location), the mapping file and the repository's Issue list, and returns one row per
//!   *selected* task: `local_id`, `state`, `target`, `redacted_title`, `redacted_body_digest`,
//!   plus a `preview_digest`. It writes nothing anywhere, local or remote.
//! - `apply` recomputes the preview, refuses a different digest (`stale`), and for each
//!   `preview_only` row records an outbox intent, creates the Issue with an operation marker,
//!   and records `published` only after a fresh GET shows the same id, node_id, number,
//!   html_url and the marker. A result it cannot know (transport error, timeout, 5xx) is
//!   recorded `uncertain` and is never created again automatically; a later run looks for the
//!   marker first and links the Issue if it exists.
//!
//! Unselected tasks never appear in any output or request. Duplicate detection reads the
//! Issue list (RC1 interim decision B: list + local marker match, no search API) and only
//! marks candidates. The disclosure gate requires `disclosure_review: approved` and refuses
//! credential-shaped text and personal absolute paths; it is conservative, not a secret
//! scanner (the shared redaction module belongs to 03-B #1996; gitleaks is HOLD).
use super::{
    outbox::{self, OutboxEntry, OutboxState, Phase},
    policy::{format_utc, next_link, Clock},
    transport::{HttpRequest, HttpResponse, Transport},
    Client, Config, ForgeError, API,
};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{
    collections::{BTreeMap, BTreeSet},
    fs::{self, OpenOptions},
    io::Write,
    path::{Path, PathBuf},
};

const OPERATION: &str = "promote";
const PROMOTIONS: &str = "forge/promotions.json";
const LIMIT: u64 = 65536;

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Target {
    pub owner: String,
    pub repo: String,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct PromoteRequest {
    pub schema_version: u32,
    /// Path to `tasks.md`, absolute or relative to the project. Required (no default).
    pub tasks_path: String,
    pub target: Target,
    /// Task ids the user chose; nothing else is read into the output.
    pub selected: Vec<String>,
    #[serde(default)]
    pub config: Config,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct LocalTask {
    pub id: String,
    pub title: String,
    pub body: Vec<String>,
    pub disclosure_approved: bool,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct TasksFile {
    pub project_id: String,
    pub tasks: Vec<LocalTask>,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum RecordState {
    Published,
    Uncertain,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct PromotionRecord {
    pub local_id: String,
    pub marker: String,
    pub state: RecordState,
    pub attempt: u32,
    /// The 07-A external reference once published.
    pub remote: Option<Value>,
    pub recorded_at: String,
}

#[derive(Clone, Debug, Default, PartialEq, Serialize, Deserialize)]
pub struct Promotions {
    pub schema_version: u32,
    pub project_id: String,
    pub records: Vec<PromotionRecord>,
}

fn valid_id(s: &str) -> bool {
    !s.is_empty()
        && s.len() <= 64
        && s.bytes()
            .all(|b| b.is_ascii_alphanumeric() || b"-_.".contains(&b))
        && s != "."
        && s != ".."
}

/// Parses the design-22 template: `Project: <id>` header and task lines
/// `- [ ] <id> · <title> · key=value…`, each followed by `  - Key: value` lines.
/// `Evidence` and `External issue` stay local; `disclosure_review: approved` (or
/// `Disclosure review: approved`) is the gate flag; other lines and `key=value` attributes
/// except `state` form the Issue body.
pub fn parse_tasks(text: &str) -> Result<TasksFile, ForgeError> {
    let bad = |why: &str| ForgeError::invalid(format!("invalid tasks file: {why}"));
    let mut project_id = None;
    let mut tasks: Vec<LocalTask> = Vec::new();
    for line in text.lines() {
        if let Some(p) = line.strip_prefix("Project:") {
            let p = p.trim();
            if !valid_id(p) {
                return Err(bad("Project must be a portable id"));
            }
            project_id = Some(p.to_string());
        } else if let Some(rest) = line
            .strip_prefix("- [ ] ")
            .or_else(|| line.strip_prefix("- [x] "))
        {
            let mut parts = rest.split(" · ");
            let id = parts.next().unwrap_or("").trim().to_string();
            let title = parts.next().unwrap_or("").trim().to_string();
            if !valid_id(&id) || title.is_empty() || title.len() > 256 {
                return Err(bad("each task needs a portable id and a title"));
            }
            if tasks.iter().any(|t| t.id == id) {
                return Err(bad("duplicate task id"));
            }
            let body = parts
                .filter_map(|a| a.trim().split_once('='))
                .filter(|(k, _)| k.trim() != "state")
                .map(|(k, v)| format!("- {}: {}", k.trim(), v.trim()))
                .collect();
            tasks.push(LocalTask {
                id,
                title,
                body,
                disclosure_approved: false,
            });
        } else if let Some(sub) = line.strip_prefix("  - ") {
            let Some(task) = tasks.last_mut() else {
                continue;
            };
            let (key, value) = sub.split_once(':').unwrap_or((sub, ""));
            let (key, value) = (key.trim(), value.trim());
            match key.to_ascii_lowercase().replace(' ', "_").as_str() {
                "evidence" | "external_issue" => {}
                "disclosure_review" => task.disclosure_approved = value == "approved",
                _ => task.body.push(format!("- {key}: {value}")),
            }
        }
    }
    let project_id = project_id.ok_or_else(|| bad("missing Project header"))?;
    Ok(TasksFile { project_id, tasks })
}

/// Conservative disclosure gate: the profile filter's credential shapes, a few more common
/// ones (cloud keys, bearer headers, chat tokens) and personal absolute paths.
fn sensitive(s: &str) -> bool {
    let lower = s.to_ascii_lowercase();
    crate::SECRET_SHAPES
        .iter()
        .chain(&[
            "akia",
            "bearer ",
            "xoxb-",
            "xoxp-",
            "aws_secret",
            "/users/",
            "/home/",
            "~/",
            ":\\",
        ])
        .any(|x| lower.contains(x))
        || s.chars().any(|c| c.is_control() && c != '\t')
}

fn redact_line(s: &str) -> String {
    if sensitive(s) {
        "[redacted]".into()
    } else {
        s.into()
    }
}

fn tokens(title: &str) -> BTreeSet<String> {
    title
        .to_lowercase()
        .split(|c: char| !c.is_alphanumeric())
        .filter(|t| t.len() >= 2)
        .map(str::to_string)
        .collect()
}

/// Same words (Jaccard ≥ 0.8) or the same normalized title.
fn similar(a: &str, b: &str) -> bool {
    let (x, y) = (tokens(a), tokens(b));
    let union = x.union(&y).count();
    union > 0 && x.intersection(&y).count() * 5 >= union * 4
}

fn resolve_tasks(project: &Path, path: &str) -> Result<PathBuf, ForgeError> {
    let p = Path::new(path);
    let p = if p.is_absolute() {
        p.to_path_buf()
    } else {
        project.join(p)
    };
    match fs::symlink_metadata(&p) {
        Ok(m) if m.is_file() && !m.file_type().is_symlink() && m.len() <= LIMIT => Ok(p),
        _ => Err(ForgeError::invalid(
            "invalid tasks file: must be a bounded regular file, not a symlink",
        )),
    }
}

fn promotions_path(project: &Path) -> PathBuf {
    project.join(crate::PROJECT_STATE_DIR).join(PROMOTIONS)
}

/// The mapping and the digest of the bytes read (`absent` when there is no file).
pub fn load_promotions(project: &Path) -> Result<(Promotions, String), ForgeError> {
    match fs::read(promotions_path(project)) {
        Ok(bytes) => {
            let p: Promotions = serde_json::from_slice(&bytes)
                .map_err(|_| ForgeError::io("promotions file unreadable"))?;
            Ok((p, crate::digest(&bytes)))
        }
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
            Ok((Promotions::default(), "absent".into()))
        }
        Err(_) => Err(ForgeError::io("promotions file unreadable")),
    }
}

/// Writes the mapping only if the file still has the digest that was read (REQ-PROMO-008).
pub fn save_promotions(
    project: &Path,
    p: &Promotions,
    read_digest: &str,
) -> Result<(), ForgeError> {
    let (_, now) = load_promotions(project)?;
    if now != read_digest {
        return Err(ForgeError::of(
            "conflict",
            "conflict in promotions.json: changed by another writer since it was read; nothing overwritten",
        ));
    }
    let mut dir = project.to_path_buf();
    for part in [crate::PROJECT_STATE_DIR, "forge"] {
        dir.push(part);
        match fs::symlink_metadata(&dir) {
            Ok(m) if m.file_type().is_symlink() || !m.is_dir() => {
                return Err(ForgeError::invalid(
                    "invalid forge state directory: a symlink or not a directory; refused",
                ))
            }
            Ok(_) => {}
            Err(_) => fs::create_dir(&dir)
                .map_err(|_| ForgeError::io("forge state directory unavailable"))?,
        }
    }
    let path = promotions_path(project);
    let tmp = path.with_extension(format!("json.tmp-{}", std::process::id()));
    let bytes =
        serde_json::to_vec_pretty(p).map_err(|_| ForgeError::of("internal", "encode failed"))?;
    let mut f = OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&tmp)
        .map_err(|_| ForgeError::io("promotions write failed"))?;
    f.write_all(&bytes)
        .and_then(|_| f.sync_all())
        .map_err(|_| ForgeError::io("promotions write failed"))?;
    fs::rename(&tmp, &path).map_err(|_| {
        let _ = fs::remove_file(&tmp);
        ForgeError::io("promotions replace failed")
    })
}

/// Everything a preview computed; `apply` works from the same values.
pub struct Preview {
    pub result: Value,
    pub digest: String,
    tasks: TasksFile,
    rows: Vec<Row>,
    promotions: Promotions,
    promotions_digest: String,
    repository_id: i64,
}

struct Row {
    local_id: String,
    state: &'static str,
    title: String,
    body: String,
    /// An existing Issue carrying this task's marker.
    marker_hit: Option<Value>,
}

fn external_ref(
    req: &PromoteRequest,
    repository_id: i64,
    issue: &Value,
    etag: Option<&str>,
) -> Value {
    json!({
        "provider": "github", "host": "github.com",
        "project": {"owner": req.target.owner, "repo": req.target.repo, "repository_id": repository_id},
        "issue": {"id": issue["id"], "node_id": issue["node_id"], "number": issue["number"]},
        "updated_at": issue["updated_at"], "etag": etag,
    })
}

/// Reads, never writes (REQ-PROMO-001). Fails as a whole on any transport error, so an
/// offline run changes nothing (REQ-PROMO-010).
pub fn preview<T: Transport, C: Clock>(
    project: &Path,
    req: &PromoteRequest,
    client: &Client<'_, T, C>,
) -> Result<Preview, ForgeError> {
    if req.schema_version != 1 {
        return Err(ForgeError::of(
            "unsupported_protocol_version",
            "unsupported request schema_version",
        ));
    }
    if !super::valid_name(&req.target.owner) || !super::valid_name(&req.target.repo) {
        return Err(ForgeError::invalid(
            "invalid target: owner and repo must be GitHub names",
        ));
    }
    let mut selected = req.selected.clone();
    selected.sort();
    selected.dedup();
    if selected.is_empty()
        || selected.len() != req.selected.len()
        || !selected.iter().all(|s| valid_id(s))
    {
        return Err(ForgeError::invalid(
            "invalid selection: list distinct portable task ids",
        ));
    }
    let path = resolve_tasks(project, &req.tasks_path)?;
    let bytes = fs::read(&path).map_err(|_| ForgeError::io("tasks file unreadable"))?;
    let text = String::from_utf8(bytes.clone())
        .map_err(|_| ForgeError::invalid("tasks file is not UTF-8"))?;
    let tasks = parse_tasks(&text)?;
    if selected
        .iter()
        .any(|s| !tasks.tasks.iter().any(|t| &t.id == s))
    {
        return Err(ForgeError::invalid(
            "invalid selection: a selected task is missing from the tasks file",
        ));
    }
    let (promotions, promotions_digest) = load_promotions(project)?;
    if !promotions.project_id.is_empty() && promotions.project_id != tasks.project_id {
        return Err(ForgeError::of(
            "conflict",
            "conflict in promotions.json: it belongs to another project; refused",
        ));
    }

    let repo_url = format!("{API}/repos/{}/{}", req.target.owner, req.target.repo);
    let repo = client.get(&repo_url, None)?.body.unwrap_or(Value::Null);
    let repository_id = repo["id"]
        .as_i64()
        .ok_or_else(|| ForgeError::transport("repository response carried no id"))?;
    let visibility = repo["visibility"]
        .as_str()
        .map(str::to_string)
        .unwrap_or_else(|| {
            if repo["private"] == json!(true) {
                "private".into()
            } else {
                "public".into()
            }
        });

    // Decision B: the Issue list plus a local marker match. Bodies are reduced to their
    // markers at once; titles are kept only for the duplicate check.
    let list_url = format!(
        "{API}/repos/{}/{}/issues?state=all&sort=updated&direction=asc&per_page=100",
        req.target.owner, req.target.repo
    );
    let project_page = |r: &HttpResponse| {
        let etag = r.header("etag");
        let items: Vec<Value> = r
            .body
            .as_ref()
            .and_then(Value::as_array)
            .map(|a| {
                a.iter()
                    .filter(|i| i.get("pull_request").is_none())
                    .map(|i| {
                        let markers: Vec<Value> =
                            outbox::parse_op_markers(i["body"].as_str().unwrap_or(""))
                                .into_iter()
                                .map(|m| json!([m.operation, m.project_id, m.local_id, m.attempt]))
                                .collect();
                        json!({"id": i["id"], "node_id": i["node_id"], "number": i["number"],
                               "title": i["title"], "html_url": i["html_url"],
                               "updated_at": i["updated_at"], "etag": etag, "markers": markers})
                    })
                    .collect()
            })
            .unwrap_or_default();
        json!({"items": items, "next": next_link(r.header("link"))})
    };
    let (issues, _) = client.pages(list_url, &project_page)?;

    let mut rows = Vec::new();
    for task in tasks.tasks.iter().filter(|t| selected.contains(&t.id)) {
        let record = promotions.records.iter().find(|r| r.local_id == task.id);
        let hit = issues.iter().find(|i| {
            i["markers"].as_array().is_some_and(|ms| {
                ms.iter().any(|m| {
                    m[0] == OPERATION
                        && m[1] == tasks.project_id.as_str()
                        && m[2] == task.id.as_str()
                })
            })
        });
        // An outbox entry for the attempt apply would use next, left by a run that stopped
        // before it could record the mapping: only a `blocked` (refused) one may be retried.
        let next_attempt = record.map(|r| r.attempt + 1).unwrap_or(1);
        let outbox_unknown = outbox::op_id(OPERATION, &tasks.project_id, &task.id, next_attempt)
            .ok()
            .map(|id| outbox::load(project, &id))
            .transpose()?
            .flatten()
            .is_some_and(|e| e.state != OutboxState::Blocked);
        let body = task.body.join("\n");
        let blocked = !task.disclosure_approved
            || sensitive(&task.title)
            || task.body.iter().any(|l| sensitive(l));
        let state = match (record.map(|r| r.state), hit) {
            (Some(RecordState::Published), _) | (_, Some(_)) => "already_published",
            (Some(RecordState::Uncertain), None) => "uncertain",
            _ if outbox_unknown => "uncertain",
            _ if blocked => "blocked_review",
            _ if issues
                .iter()
                .any(|i| similar(i["title"].as_str().unwrap_or(""), &task.title)) =>
            {
                "duplicate_candidate"
            }
            _ => "preview_only",
        };
        let redacted_title = redact_line(&task.title);
        let redacted_body: String = task
            .body
            .iter()
            .map(|l| redact_line(l))
            .collect::<Vec<_>>()
            .join("\n");
        rows.push(Row {
            local_id: task.id.clone(),
            state,
            title: if blocked {
                redacted_title
            } else {
                task.title.clone()
            },
            body: if blocked { redacted_body } else { body },
            marker_hit: hit.cloned(),
        });
    }
    let target =
        json!({"owner": req.target.owner, "repo": req.target.repo, "visibility": visibility});
    let row_json: Vec<Value> = rows
        .iter()
        .map(|r| {
            json!({"local_id": r.local_id, "state": r.state, "target": target,
                   "redacted_title": redact_line(&r.title),
                   "redacted_body_digest": crate::digest(r.body.lines().map(redact_line).collect::<Vec<_>>().join("\n").as_bytes())})
        })
        .collect();
    let digest = crate::digest(
        &serde_json::to_vec(&(
            1,
            crate::digest(&bytes),
            &promotions_digest,
            &target,
            &selected,
            &row_json,
        ))
        .map_err(|_| ForgeError::of("internal", "encode failed"))?,
    );
    let result = json!({"schema_version": 1, "project_id": tasks.project_id, "rows": row_json, "preview_digest": digest});
    Ok(Preview {
        result,
        digest,
        tasks,
        rows,
        promotions,
        promotions_digest,
        repository_id,
    })
}

fn issue_url(req: &PromoteRequest, number: &Value) -> String {
    format!(
        "{API}/repos/{}/{}/issues/{}",
        req.target.owner, req.target.repo, number
    )
}

/// GET the Issue again; `published` only if id, node_id, number, html_url and marker agree.
fn confirm<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    req: &PromoteRequest,
    created: &Value,
    marker: &str,
) -> Result<Option<(Value, Option<String>)>, ForgeError> {
    let resp = client.get(&issue_url(req, &created["number"]), None)?;
    let got = resp.body.clone().unwrap_or(Value::Null);
    let same = ["id", "node_id", "number", "html_url"]
        .iter()
        .all(|k| got[*k] == created[*k] && !got[*k].is_null());
    let marked = got["body"].as_str().is_some_and(|b| b.contains(marker));
    Ok((same && marked).then(|| (got, resp.header("etag").map(str::to_string))))
}

/// REQ-PROMO-004..008. Returns per-row outcomes; stops at the first result it cannot know.
pub fn apply<T: Transport, C: Clock>(
    project: &Path,
    req: &PromoteRequest,
    preview_digest: &str,
    permission: &str,
    client: &Client<'_, T, C>,
) -> Result<Value, ForgeError> {
    // Caller authorization remains host-owned; the reference is not an authorization verifier.
    if ["", "null", "placeholder"].contains(&permission.trim())
        || permission.len() > 128
        || permission.chars().any(char::is_control)
    {
        return Err(ForgeError::of(
            "permission_reference_missing",
            "permission reference required",
        ));
    }
    let pv = preview(project, req, client)?;
    if pv.digest != preview_digest {
        return Err(ForgeError::of(
            "stale_plan",
            "stale or altered plan; re-preview required",
        ));
    }
    let project_id = pv.tasks.project_id.clone();
    let mut records: BTreeMap<String, PromotionRecord> = pv
        .promotions
        .records
        .iter()
        .map(|r| (r.local_id.clone(), r.clone()))
        .collect();
    let mut outcomes = Vec::new();
    let mut stop: Option<String> = None;
    let mut wrote = false;
    let now = || format_utc(client.clock.now());

    for row in &pv.rows {
        if stop.is_some() {
            break;
        }
        let prev = records.get(&row.local_id).cloned();
        match (row.state, &row.marker_hit) {
            ("already_published", Some(hit))
                if prev.as_ref().map(|r| r.state) != Some(RecordState::Published) =>
            {
                // Reconcile: the marker is on the remote; link it after a fresh GET.
                let marker = hit["markers"]
                    .as_array()
                    .and_then(|ms| ms.iter().find(|m| m[2] == row.local_id.as_str()))
                    .map(|m| {
                        outbox::op_marker(
                            OPERATION,
                            &project_id,
                            &row.local_id,
                            m[3].as_u64().unwrap_or(1) as u32,
                        )
                    })
                    .unwrap_or_default();
                match confirm(client, req, hit, &marker)? {
                    Some((got, etag)) => {
                        let attempt = prev.as_ref().map(|r| r.attempt).unwrap_or(1);
                        if let Ok(id) =
                            outbox::op_id(OPERATION, &project_id, &row.local_id, attempt)
                        {
                            if let Some(mut e) = outbox::load(project, &id)? {
                                if e.state == OutboxState::PublishUnknown {
                                    e.advance(OutboxState::Confirmed, Phase { at: now(), detail: json!({"number": got["number"], "via": "marker"}) })?;
                                    outbox::save(project, &e)?;
                                }
                            }
                        }
                        records.insert(
                            row.local_id.clone(),
                            PromotionRecord {
                                local_id: row.local_id.clone(),
                                marker,
                                state: RecordState::Published,
                                attempt,
                                remote: Some(external_ref(
                                    req,
                                    pv.repository_id,
                                    &got,
                                    etag.as_deref(),
                                )),
                                recorded_at: now(),
                            },
                        );
                        outcomes.push(json!({"local_id": row.local_id, "outcome": "linked", "number": got["number"]}));
                    }
                    None => {
                        stop = Some(format!(
                            "failed at {}: marker Issue did not confirm",
                            row.local_id
                        ));
                    }
                }
            }
            ("preview_only", _) => {
                let attempt = prev.as_ref().map(|r| r.attempt + 1).unwrap_or(1);
                let marker = outbox::op_marker(OPERATION, &project_id, &row.local_id, attempt);
                let op_id = outbox::op_id(OPERATION, &project_id, &row.local_id, attempt)?;
                let mut entry = OutboxEntry::new(&op_id, &marker);
                let body = format!("{}\n\n{}", row.body, marker)
                    .trim_start()
                    .to_string();
                let payload = json!({"title": row.title, "body": body});
                entry.advance(OutboxState::Publishing, Phase {
                    at: now(),
                    detail: json!({"payload_digest": crate::digest(payload.to_string().as_bytes())}),
                })?;
                outbox::save(project, &entry)?;
                // The clock counts whole seconds, so a full second is slept between writes
                // rather than trusting a one-second difference that may be milliseconds.
                if wrote {
                    client.clock.sleep(1);
                }
                let sent = client.send_with_backoff(&HttpRequest {
                    method: "POST".into(),
                    url: format!(
                        "{API}/repos/{}/{}/issues",
                        req.target.owner, req.target.repo
                    ),
                    headers: Default::default(),
                    body: Some(payload),
                });
                wrote = true;
                let uncertain = |entry: &mut OutboxEntry, why: String| -> Result<(), ForgeError> {
                    entry.advance(
                        OutboxState::PublishUnknown,
                        Phase {
                            at: now(),
                            detail: json!({"reason": why}),
                        },
                    )?;
                    outbox::save(project, entry)
                };
                match sent {
                    Ok(resp) if (200..300).contains(&resp.status) => {
                        let created = resp.body.clone().unwrap_or(Value::Null);
                        entry.record_response(Phase {
                            at: now(),
                            detail: json!({"status": resp.status, "number": created["number"]}),
                        })?;
                        outbox::save(project, &entry)?;
                        match confirm(client, req, &created, &marker) {
                            Ok(Some((got, etag))) => {
                                entry.advance(
                                    OutboxState::Confirmed,
                                    Phase {
                                        at: now(),
                                        detail: json!({"number": got["number"]}),
                                    },
                                )?;
                                outbox::save(project, &entry)?;
                                records.insert(
                                    row.local_id.clone(),
                                    PromotionRecord {
                                        local_id: row.local_id.clone(),
                                        marker,
                                        state: RecordState::Published,
                                        attempt,
                                        remote: Some(external_ref(
                                            req,
                                            pv.repository_id,
                                            &got,
                                            etag.as_deref(),
                                        )),
                                        recorded_at: now(),
                                    },
                                );
                                outcomes.push(json!({"local_id": row.local_id, "outcome": "published", "number": got["number"]}));
                            }
                            Ok(None) => {
                                entry.advance(OutboxState::Conflict, Phase { at: now(), detail: json!({"reason": "re-read did not match the response"}) })?;
                                outbox::save(project, &entry)?;
                                records.insert(
                                    row.local_id.clone(),
                                    PromotionRecord {
                                        local_id: row.local_id.clone(),
                                        marker,
                                        state: RecordState::Uncertain,
                                        attempt,
                                        remote: None,
                                        recorded_at: now(),
                                    },
                                );
                                stop = Some(format!(
                                    "uncertain at {}: the created Issue did not confirm",
                                    row.local_id
                                ));
                            }
                            Err(e) => {
                                uncertain(&mut entry, e.kind.to_string())?;
                                records.insert(
                                    row.local_id.clone(),
                                    PromotionRecord {
                                        local_id: row.local_id.clone(),
                                        marker,
                                        state: RecordState::Uncertain,
                                        attempt,
                                        remote: None,
                                        recorded_at: now(),
                                    },
                                );
                                stop = Some(format!(
                                    "uncertain at {}: confirmation read failed",
                                    row.local_id
                                ));
                            }
                        }
                    }
                    Ok(resp) if resp.status >= 500 || resp.status == 408 => {
                        uncertain(&mut entry, format!("HTTP {}", resp.status))?;
                        records.insert(
                            row.local_id.clone(),
                            PromotionRecord {
                                local_id: row.local_id.clone(),
                                marker,
                                state: RecordState::Uncertain,
                                attempt,
                                remote: None,
                                recorded_at: now(),
                            },
                        );
                        stop = Some(format!(
                            "uncertain at {}: HTTP {}",
                            row.local_id, resp.status
                        ));
                    }
                    Ok(resp) => {
                        entry.advance(
                            OutboxState::Blocked,
                            Phase {
                                at: now(),
                                detail: json!({"status": resp.status}),
                            },
                        )?;
                        outbox::save(project, &entry)?;
                        stop = Some(format!("failed at {}: HTTP {}", row.local_id, resp.status));
                    }
                    Err(e) if e.kind == "blocked" => {
                        // 401/403 or an exhausted rate limit: GitHub refused the request.
                        entry.advance(
                            OutboxState::Blocked,
                            Phase {
                                at: now(),
                                detail: json!({"reason": "blocked"}),
                            },
                        )?;
                        outbox::save(project, &entry)?;
                        stop = Some(format!("failed at {}: forge blocked", row.local_id));
                    }
                    Err(e) => {
                        // Sent or not, the outcome is unknown: never resend automatically.
                        uncertain(&mut entry, e.kind.to_string())?;
                        records.insert(
                            row.local_id.clone(),
                            PromotionRecord {
                                local_id: row.local_id.clone(),
                                marker,
                                state: RecordState::Uncertain,
                                attempt,
                                remote: None,
                                recorded_at: now(),
                            },
                        );
                        stop = Some(format!("uncertain at {}: {}", row.local_id, e.kind));
                    }
                }
            }
            (state, _) => outcomes
                .push(json!({"local_id": row.local_id, "outcome": "not_sent", "state": state})),
        }
    }

    let promotions = Promotions {
        schema_version: 1,
        project_id: project_id.clone(),
        records: records.into_values().collect(),
    };
    if promotions.records != pv.promotions.records {
        save_promotions(project, &promotions, &pv.promotions_digest)?;
    }
    if let Some(why) = stop {
        let done: Vec<&str> = outcomes
            .iter()
            .filter(|o| o["outcome"] == "published" || o["outcome"] == "linked")
            .filter_map(|o| o["local_id"].as_str())
            .collect();
        return Err(ForgeError::of(
            "partial",
            format!("partial apply: published [{}]; {why}", done.join(", ")),
        ));
    }
    Ok(
        json!({"schema_version": 1, "project_id": project_id, "preview_digest": pv.digest, "outcomes": outcomes}),
    )
}

/// CLI entries (`todo.promote.plan`, `todo.promote.apply`): errors are their messages.
pub fn preview_cli<T: Transport, C: Clock>(
    project: &Path,
    req: PromoteRequest,
    transport: &T,
    clock: &C,
) -> Result<Value, String> {
    let client = Client::new(transport, clock, req.config.clone(), Default::default());
    Ok(preview(project, &req, &client)?.result)
}

pub fn apply_cli<T: Transport, C: Clock>(
    project: &Path,
    req: PromoteRequest,
    preview_digest: &str,
    permission: &str,
    transport: &T,
    clock: &C,
) -> Result<Value, String> {
    let client = Client::new(transport, clock, req.config.clone(), Default::default());
    Ok(apply(project, &req, preview_digest, permission, &client)?)
}
