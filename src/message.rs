//! Durable receipt protocol, not a work/ownership or business-acceptance protocol.
//!
//! Payload identity is `crate::digest` of compact UTF-8 serde_json JSON with recursively
//! lexicographically sorted object keys. Array order and serde_json number representation
//! are retained. Retry observations are outside the immutable message. Session IDs are
//! declarations; numeric GitHub author IDs prove the account only, not a host/session.
//! The persistent lock file is an OS advisory-lock target, not a PID/liveness claim.
//! The rendered comment (markers included) uses forge-report's shared character limit;
//! stdin limits remain the CLI's responsibility. Shared secret-shape detection examines
//! decoded string keys/values, never JSON escaping. `received` records an explicit ACK
//! action after authenticating the receiver; remote-only reconciliation creates no receipt.
//! Definitive HTTP 4xx refusals except 408 allow a new attempt only on a new explicit
//! invocation after fresh lookup. Unknown outcomes never permit blind retransmission.
use crate::contract::{Kind, OpError};
use crate::forge::{
    outbox,
    policy::{next_link, parse_utc, Clock},
    transport::{HttpRequest, Transport},
    Client, Config,
};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{
    collections::{BTreeMap, BTreeSet},
    fs::{self, OpenOptions},
    io::Write,
    path::{Component, Path, PathBuf},
};

type Result<T, E = OpError> = std::result::Result<T, E>;
const API: &str = "https://api.github.com";
const END: &str = "<!-- moai:message:end -->";

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Actor {
    pub actor_id: u64,
    pub session_id: String,
}
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Task {
    pub owner: String,
    pub repo: String,
    pub repository_id: u64,
    pub issue_id: u64,
    pub issue_node_id: String,
    pub issue_number: u64,
    pub task_id: String,
    pub revision: String,
}
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Message {
    pub schema_version: u32,
    pub report_id: String,
    pub project_id: String,
    pub task: Task,
    pub sender: Actor,
    pub receiver: Actor,
    pub owner_epoch: u64,
    pub leader_epoch: u64,
    pub event: String,
    pub candidate_digest: String,
    pub created_at: String,
    pub approval_refs: Vec<String>,
    pub hold_refs: Vec<String>,
    pub payload: Value,
}
#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct RecordRequest {
    pub message: Message,
}
#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct AckRequest {
    pub message: Message,
    pub payload_digest: String,
}
#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ReconcileRequest {
    pub message: Message,
}

#[derive(Default, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
struct Observation {
    attempted: bool,
    attempts: u64,
    attempted_at: Option<i64>,
    response_status: Option<u16>,
    comment_id: Option<u64>,
    observed_at: Option<i64>,
}
#[derive(Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
struct Entry {
    schema_version: u32,
    message: Message,
    payload_digest: String,
    record: Observation,
    ack: Observation,
    received: bool,
}
#[derive(Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
struct Wire {
    schema_version: u32,
    kind: String,
    message: Message,
    payload_digest: String,
}
struct Seen {
    record: Option<u64>,
    ack: Option<u64>,
}
fn err(kind: Kind, text: &str) -> OpError {
    OpError::new(kind, text)
}
fn invalid(text: &str) -> OpError {
    err(Kind::InvalidInput, text)
}
fn conflict() -> OpError {
    err(
        Kind::Conflict,
        "message immutable identity or remote author/context conflict",
    )
}
fn portable(s: &str) -> bool {
    !s.is_empty()
        && s.len() <= 64
        && s != "."
        && s != ".."
        && s.bytes()
            .all(|b| b.is_ascii_alphanumeric() || b"-_.".contains(&b))
}
// Task/session/revision IDs are opaque literal data, never path or command fragments.
fn literal_id(s: &str) -> bool {
    !s.is_empty() && !s.chars().any(char::is_control)
}
fn digest_valid(s: &str) -> bool {
    s.len() == 71
        && s.starts_with("sha256:")
        && s[7..]
            .bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}
fn sensitive(s: &str) -> bool {
    crate::redact::contains_secret_shape(s)
}
fn sensitive_value(value: &Value) -> bool {
    match value {
        Value::String(s) => sensitive(s),
        Value::Array(items) => items.iter().any(sensitive_value),
        Value::Object(map) => map
            .iter()
            .any(|(key, value)| sensitive(key) || sensitive_value(value)),
        _ => false,
    }
}
pub(crate) fn validate(message: &mut Message) -> Result<String> {
    let t = &message.task;
    if message.schema_version != 1
        || !portable(&message.report_id)
        || !portable(&message.project_id)
        || !portable(&t.owner)
        || !portable(&t.repo)
        || !literal_id(&t.task_id)
        || t.repository_id == 0
        || t.issue_id == 0
        || t.issue_number == 0
        || t.issue_node_id.is_empty()
        || !literal_id(&t.revision)
        || message.sender.actor_id == 0
        || message.receiver.actor_id == 0
        || !literal_id(&message.sender.session_id)
        || !literal_id(&message.receiver.session_id)
        || message.owner_epoch == 0
        || message.leader_epoch == 0
        || !portable(&message.event)
        || !digest_valid(&message.candidate_digest)
        || parse_utc(&message.created_at).is_none()
        || message
            .approval_refs
            .iter()
            .chain(&message.hold_refs)
            .any(|s| s.trim().is_empty())
    {
        return Err(invalid(
            "invalid message schema, identity, epoch, revision, reference or digest",
        ));
    }
    let fields = [
        &message.report_id,
        &message.project_id,
        &t.owner,
        &t.repo,
        &t.issue_node_id,
        &t.task_id,
        &t.revision,
        &message.sender.session_id,
        &message.receiver.session_id,
        &message.event,
        &message.candidate_digest,
        &message.created_at,
    ];
    if fields
        .into_iter()
        .chain(&message.approval_refs)
        .chain(&message.hold_refs)
        .any(|s| sensitive(s))
        || sensitive_value(&message.payload)
    {
        return Err(err(Kind::Blocked, "message blocked_sensitive"));
    }
    message.payload.sort_all_objects();
    let bytes = serde_json::to_vec(&message.payload)
        .map_err(|_| err(Kind::Internal, "message encoding failed"))?;
    Ok(crate::digest(&bytes))
}
pub(crate) fn permission_gate(permission: &str) -> Result<()> {
    if ["", "null", "placeholder"].contains(&permission.trim())
        || permission.len() > 128
        || permission.chars().any(char::is_control)
    {
        return Err(err(
            Kind::PermissionReferenceMissing,
            "permission reference required",
        ));
    }
    if sensitive(permission) {
        return Err(err(Kind::Blocked, "message blocked_sensitive"));
    }
    Ok(())
}

// These prechecks reject known symlinks, not concurrent filesystem substitution.
fn check_path(path: &Path) -> Result<()> {
    let mut current = PathBuf::new();
    for part in path.components() {
        if matches!(part, Component::ParentDir) {
            return Err(invalid("message path traversal refused"));
        }
        current.push(part.as_os_str());
        match fs::symlink_metadata(&current) {
            Ok(m) if m.file_type().is_symlink() => {
                return Err(invalid("message symlink path refused"))
            }
            Ok(_) => {}
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => {}
            Err(_) => return Err(err(Kind::Io, "message path unavailable")),
        }
    }
    Ok(())
}
fn sync_directory(path: &Path) -> Result<()> {
    let result = fs::File::open(path).and_then(|f| f.sync_all());
    #[cfg(all(test, unix))]
    let result = result.and_then(|()| durability_tests::reported_result(path));
    result.map_err(|_| err(Kind::Io, "message directory sync failed"))?;
    #[cfg(all(test, unix))]
    durability_tests::synced(path);
    Ok(())
}
fn directories(project: &Path) -> Result<PathBuf> {
    check_path(project)?;
    if !project.is_dir() {
        return Err(invalid("message project must be an existing directory"));
    }
    let root = project.join(crate::PROJECT_STATE_DIR).join("messages");
    for path in [
        project.join(crate::PROJECT_STATE_DIR),
        root.clone(),
        root.join("outbox"),
        root.join("inbox"),
    ] {
        check_path(&path)?;
        match fs::create_dir(&path) {
            Ok(()) => {
                #[cfg(all(test, unix))]
                durability_tests::created(&path);
            }
            Err(e) if e.kind() == std::io::ErrorKind::AlreadyExists && path.is_dir() => {}
            Err(_) => return Err(err(Kind::Io, "message directory unavailable")),
        }
        // An existing path may be left by an interrupted invocation before its parent
        // entry reached storage. Never infer durability from AlreadyExists.
        sync_directory(
            path.parent()
                .ok_or_else(|| invalid("message parent missing"))?,
        )?;
        sync_directory(&path)?;
    }
    Ok(root)
}
fn lock(root: &Path) -> Result<fs::File> {
    let path = root.join(".lock");
    check_path(&path)?;
    if fs::symlink_metadata(&path).is_ok_and(|m| !m.is_file()) {
        return Err(invalid("message lock is not a regular file"));
    }
    let file = OpenOptions::new()
        .read(true)
        .write(true)
        .create(true)
        .truncate(false)
        .open(path)
        .map_err(|_| err(Kind::Io, "message lock unavailable"))?;
    match file.try_lock() {
        Ok(()) => Ok(file),
        Err(fs::TryLockError::WouldBlock) => Err(err(
            Kind::Unavailable,
            "message lock wait: another invocation holds the advisory lock",
        )),
        Err(fs::TryLockError::Error(_)) => Err(err(Kind::Io, "message lock unavailable")),
    }
}
fn load(path: &Path, message: &Message, digest: &str) -> Result<Option<Entry>> {
    check_path(path)?;
    match fs::symlink_metadata(path) {
        Ok(m) if !m.is_file() => return Err(invalid("message entry is not a regular file")),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(None),
        Err(_) => return Err(err(Kind::Io, "message entry unavailable")),
        _ => {}
    }
    let bytes = fs::read(path).map_err(|_| err(Kind::Io, "message entry unreadable"))?;
    let entry: Entry =
        serde_json::from_slice(&bytes).map_err(|_| invalid("malformed message entry preserved"))?;
    if entry.schema_version != 1 || entry.message != *message || entry.payload_digest != digest {
        return Err(conflict());
    }
    if entry.record.comment_id == Some(0)
        || entry.ack.comment_id == Some(0)
        || (entry.record.attempted != (entry.record.attempts > 0))
        || (entry.ack.attempted != (entry.ack.attempts > 0))
        || (entry.record.attempted != entry.record.attempted_at.is_some())
        || (entry.ack.attempted != entry.ack.attempted_at.is_some())
        || (entry.received && entry.record.comment_id.is_none())
        || (entry.ack.comment_id.is_some() && entry.record.comment_id.is_none())
    {
        return Err(invalid("inconsistent message observations preserved"));
    }
    Ok(Some(entry))
}
fn save(path: &Path, entry: &Entry) -> Result<()> {
    check_path(path)?;
    let tmp = path.with_extension(format!("json.tmp-{}", std::process::id()));
    check_path(&tmp)?;
    let mut f = OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&tmp)
        .map_err(|_| {
            err(
                Kind::Io,
                "message temporary entry unavailable; existing file preserved",
            )
        })?;
    let result = (|| {
        serde_json::to_writer(&mut f, entry)
            .map_err(|_| err(Kind::Io, "message entry write failed"))?;
        f.write_all(b"\n")
            .and_then(|_| f.sync_all())
            .map_err(|_| err(Kind::Io, "message entry sync failed"))?;
        fs::rename(&tmp, path).map_err(|_| err(Kind::Io, "message entry replace failed"))?;
        sync_directory(
            path.parent()
                .ok_or_else(|| invalid("message parent missing"))?,
        )?;
        Ok(())
    })();
    if result.is_err() {
        let _ = fs::remove_file(&tmp);
    }
    result
}
fn new_entry(message: Message, digest: String) -> Entry {
    Entry {
        schema_version: 1,
        message,
        payload_digest: digest,
        record: Observation::default(),
        ack: Observation::default(),
        received: false,
    }
}
fn url(message: &Message) -> String {
    format!(
        "{API}/repos/{}/{}/issues/{}/comments",
        message.task.owner, message.task.repo, message.task.issue_number
    )
}
fn remote_error(error: crate::forge::ForgeError) -> OpError {
    OpError::from(String::from(error))
}
fn verify_target<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    message: &Message,
) -> Result<()> {
    let t = &message.task;
    let repo = client
        .get(&format!("{API}/repos/{}/{}", t.owner, t.repo), None)
        .map_err(remote_error)?
        .body
        .ok_or_else(|| err(Kind::Transport, "message repository response missing"))?;
    if repo["id"].as_u64() != Some(t.repository_id) {
        return Err(conflict());
    }
    let issue = client
        .get(
            &format!(
                "{API}/repos/{}/{}/issues/{}",
                t.owner, t.repo, t.issue_number
            ),
            None,
        )
        .map_err(remote_error)?
        .body
        .ok_or_else(|| err(Kind::Transport, "message issue response missing"))?;
    if issue["id"].as_u64() != Some(t.issue_id)
        || issue["node_id"].as_str() != Some(t.issue_node_id.as_str())
        || issue["number"].as_u64() != Some(t.issue_number)
        || issue.get("pull_request").is_some_and(|v| !v.is_null())
    {
        return Err(conflict());
    }
    Ok(())
}
fn marker(message: &Message, ack: bool) -> String {
    outbox::op_marker(
        if ack { "message-ack" } else { "message-record" },
        &message.project_id,
        &message.report_id,
        1,
    )
}
fn render(message: &Message, payload_digest: &str, ack: bool) -> Result<String> {
    #[derive(Serialize)]
    struct Projection<'a> {
        schema_version: u32,
        kind: &'a str,
        message: &'a Message,
        payload_digest: &'a str,
    }
    let wire = Projection {
        schema_version: 1,
        kind: if ack { "ack" } else { "record" },
        message,
        payload_digest,
    };
    let mut bytes = marker(message, ack).into_bytes();
    bytes.push(b'\n');
    serde_json::to_writer(&mut bytes, &wire)
        .map_err(|_| err(Kind::Internal, "message encoding failed"))?;
    bytes.push(b'\n');
    bytes.extend_from_slice(END.as_bytes());
    let body =
        String::from_utf8(bytes).map_err(|_| err(Kind::Internal, "message encoding failed"))?;
    if body.chars().count() > crate::forge::report::DEFAULT_MAX_BODY_CHARS {
        return Err(err(
            Kind::InputTooLarge,
            "message rendered comment exceeds forge report body limit",
        ));
    }
    Ok(body)
}
fn observe<T: Transport, C: Clock>(client: &Client<'_, T, C>, entry: &Entry) -> Result<Seen> {
    let base = url(&entry.message);
    let mut next = Some(format!("{base}?per_page=100"));
    let mut visited = BTreeSet::new();
    let mut seen = Seen {
        record: None,
        ack: None,
    };
    while let Some(page) = next {
        if visited.len() >= 1000 || !visited.insert(page.clone()) {
            return Err(err(
                Kind::Blocked,
                "message pagination incomplete or cyclic",
            ));
        }
        // Follow GitHub pagination only within this exact Issue's comments collection.
        if !page.starts_with(&format!("{base}?")) {
            return Err(invalid("message pagination target refused"));
        }
        let response = client.get(&page, None).map_err(remote_error)?;
        let link = next_link(response.header("link"));
        if response
            .header("link")
            .is_some_and(|s| s.contains("rel=\"next\""))
            && link.is_none()
        {
            return Err(err(Kind::Transport, "message malformed pagination link"));
        }
        let body = response
            .body
            .ok_or_else(|| err(Kind::Transport, "message comments response missing"))?;
        let comments = body
            .as_array()
            .ok_or_else(|| err(Kind::Transport, "message comments response is not an array"))?;
        for comment in comments {
            let text = comment["body"]
                .as_str()
                .ok_or_else(|| err(Kind::Blocked, "message comment body unavailable"))?;
            let first = text.split('\n').next().unwrap_or("");
            let marks = outbox::parse_op_markers(first);
            let mut matching = marks.iter().filter(|m| {
                m.local_id == entry.message.report_id
                    && (m.operation == "message-record" || m.operation == "message-ack")
            });
            let Some(m) = matching.next() else {
                continue;
            };
            if matching.next().is_some() || marks.len() != 1 {
                return Err(conflict());
            }
            let ack = m.operation == "message-ack";
            if first != marker(&entry.message, ack) {
                return Err(conflict());
            }
            let mut lines = text.split('\n');
            lines.next();
            let encoded = lines.next().ok_or_else(conflict)?;
            if lines.next() != Some(END) || lines.next().is_some() {
                return Err(conflict());
            }
            let wire: Wire = serde_json::from_str(encoded).map_err(|_| conflict())?;
            if wire.schema_version != 1
                || wire.kind != if ack { "ack" } else { "record" }
                || wire.message != entry.message
                || wire.payload_digest != entry.payload_digest
            {
                return Err(conflict());
            }
            let actor = if ack {
                entry.message.receiver.actor_id
            } else {
                entry.message.sender.actor_id
            };
            if comment["user"]["id"].as_u64() != Some(actor) {
                return Err(conflict());
            }
            let id = comment["id"]
                .as_u64()
                .filter(|id| *id > 0)
                .ok_or_else(conflict)?;
            let slot = if ack { &mut seen.ack } else { &mut seen.record };
            // Identical duplicate publications remain one logical packet/receipt.
            *slot = Some(slot.map_or(id, |old| old.min(id)));
        }
        next = link;
    }
    if seen.ack.is_some() && seen.record.is_none() {
        return Err(conflict());
    }
    // Previously confirmed comments disappearing are unknown, not permission to republish.
    if (entry.record.comment_id.is_some() && seen.record.is_none())
        || (entry.ack.comment_id.is_some() && seen.ack.is_none())
    {
        return Err(err(
            Kind::Blocked,
            "message previously observed publication is now missing",
        ));
    }
    Ok(seen)
}
fn apply_seen(entry: &mut Entry, seen: Seen, now: i64) {
    entry.record.comment_id = seen.record;
    entry.ack.comment_id = seen.ack;
    if seen.record.is_some() {
        entry.record.observed_at = Some(now);
    }
    if seen.ack.is_some() {
        entry.ack.observed_at = Some(now);
    }
}
fn known_refusal(observation: &Observation) -> bool {
    observation
        .response_status
        .is_some_and(|s| (400..500).contains(&s) && s != 408)
}
fn can_attempt(observation: &Observation) -> bool {
    !observation.attempted || known_refusal(observation)
}
fn publication_missing(status: Option<u16>) -> OpError {
    if status.is_some_and(|s| (400..500).contains(&s) && s != 408) {
        err(
            Kind::Blocked,
            "message publication refused; durable attempt preserved; reconcile required",
        )
    } else {
        err(Kind::Transport, "message publish_unknown: publication not observed; durable attempt preserved; reconcile required, no blind retry")
    }
}
fn guard_peer(entry: Option<&Entry>, seen: &Seen, ack: bool) -> Result<()> {
    if let Some(entry) = entry {
        if (entry.record.comment_id.is_some() && seen.record.is_none())
            || (entry.ack.comment_id.is_some() && seen.ack.is_none())
        {
            return Err(err(
                Kind::Blocked,
                "message previously observed peer publication is now missing",
            ));
        }
        let observation = if ack { &entry.ack } else { &entry.record };
        if !can_attempt(observation) && (if ack { seen.ack } else { seen.record }).is_none() {
            return Err(publication_missing(observation.response_status));
        }
    }
    Ok(())
}
fn result(entry: &Entry) -> Value {
    let state = if entry.ack.comment_id.is_some() {
        "received_ack"
    } else if entry.ack.attempted && known_refusal(&entry.ack) {
        "blocked"
    } else if entry.ack.attempted {
        "publish_unknown"
    } else if entry.received {
        "received"
    } else if entry.record.comment_id.is_some() {
        "queued"
    } else if entry.record.attempted && known_refusal(&entry.record) {
        "blocked"
    } else if entry.record.attempted {
        "publish_unknown"
    } else {
        "intent"
    };
    json!({"report_id":entry.message.report_id,"payload_digest":entry.payload_digest,"state":state,
        "queued":entry.record.comment_id.is_some(),"received_ack":entry.ack.comment_id.is_some(),
        "received":entry.received,
        "record_comment_id":entry.record.comment_id,"ack_comment_id":entry.ack.comment_id,
        "business_acceptance":false,"business_completion":false})
}
fn verify_actor<T: Transport, C: Clock>(client: &Client<'_, T, C>, actor: u64) -> Result<()> {
    let user = client
        .get(&format!("{API}/user"), None)
        .map_err(remote_error)?
        .body
        .ok_or_else(|| {
            err(
                Kind::Transport,
                "message authenticated user response missing",
            )
        })?;
    if user["id"].as_u64() != Some(actor) {
        return Err(err(
            Kind::Blocked,
            "message authenticated GitHub actor does not match publisher",
        ));
    }
    Ok(())
}
fn publish<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    entry: &mut Entry,
    path: &Path,
    ack: bool,
    body: String,
) -> Result<()> {
    let observation = if ack {
        &mut entry.ack
    } else {
        &mut entry.record
    };
    observation.attempted = true;
    observation.attempts += 1;
    observation.attempted_at = Some(client.clock.now());
    observation.response_status = None;
    save(path, entry)?;
    // Client::send_with_backoff can resend writes on rate responses. Deliberately send
    // exactly once through Transport; every outcome, even an error, requires fresh GET.
    let response = client
        .transport
        .send(&HttpRequest::new("POST", url(&entry.message)).with_body(json!({"body":body})));
    let status = response.as_ref().ok().map(|r| r.status);
    if ack {
        entry.ack.response_status = status;
    } else {
        entry.record.response_status = status;
    }
    save(path, entry)?;
    let seen = observe(client, entry)?;
    apply_seen(entry, seen, client.clock.now());
    save(path, entry)?;
    // A non-observed attempt stays publish_unknown forever until a remote read finds it.
    // An absent read is not proof that the original POST cannot arrive later.
    if (if ack {
        entry.ack.comment_id
    } else {
        entry.record.comment_id
    })
    .is_none()
    {
        return Err(publication_missing(status));
    }
    Ok(())
}
fn client<'a, T: Transport, C: Clock>(transport: &'a T, clock: &'a C) -> Client<'a, T, C> {
    Client::new(transport, clock, Config::default(), BTreeMap::new())
}
fn paths(root: &Path, id: &str) -> (PathBuf, PathBuf) {
    (
        root.join("outbox").join(format!("{id}.json")),
        root.join("inbox").join(format!("{id}.json")),
    )
}

pub fn record_cli<T: Transport, C: Clock>(
    project: &Path,
    mut request: RecordRequest,
    permission: &str,
    transport: &T,
    clock: &C,
) -> Result<Value> {
    permission_gate(permission)?;
    let digest = validate(&mut request.message)?;
    let body = render(&request.message, &digest, false)?;
    let root = directories(project)?;
    let _lock = lock(&root)?;
    let (path, inbox) = paths(&root, &request.message.report_id);
    let existing = load(&path, &request.message, &digest)?;
    let peer = load(&inbox, &request.message, &digest)?;
    let mut entry = existing.unwrap_or_else(|| new_entry(request.message, digest));
    save(&path, &entry)?; // Intent is durable before even the first network read.
    let client = client(transport, clock);
    verify_target(&client, &entry.message)?;
    let seen = observe(&client, &entry)?;
    guard_peer(peer.as_ref(), &seen, false)?;
    apply_seen(&mut entry, seen, clock.now());
    save(&path, &entry)?;
    if entry.record.comment_id.is_none() && can_attempt(&entry.record) {
        verify_actor(&client, entry.message.sender.actor_id)?;
        publish(&client, &mut entry, &path, false, body)?;
    }
    if entry.record.comment_id.is_none() {
        return Err(publication_missing(entry.record.response_status));
    }
    Ok(result(&entry))
}
pub fn ack_cli<T: Transport, C: Clock>(
    project: &Path,
    mut request: AckRequest,
    permission: &str,
    transport: &T,
    clock: &C,
) -> Result<Value> {
    permission_gate(permission)?;
    let digest = validate(&mut request.message)?;
    let body = render(&request.message, &digest, true)?;
    if digest != request.payload_digest {
        return Err(conflict());
    }
    let root = directories(project)?;
    let _lock = lock(&root)?;
    let (outbox, path) = paths(&root, &request.message.report_id);
    let existing = load(&path, &request.message, &digest)?;
    let peer = load(&outbox, &request.message, &digest)?;
    let mut entry = existing.unwrap_or_else(|| new_entry(request.message, digest));
    let client = client(transport, clock);
    verify_target(&client, &entry.message)?;
    let seen = observe(&client, &entry)?;
    guard_peer(peer.as_ref(), &seen, true)?;
    if seen.record.is_none() {
        return Err(err(
            Kind::Blocked,
            "message ACK requires freshly observed sender publication",
        ));
    }
    verify_actor(&client, entry.message.receiver.actor_id)?;
    apply_seen(&mut entry, seen, clock.now());
    entry.received = true;
    save(&path, &entry)?;
    if entry.ack.comment_id.is_none() && can_attempt(&entry.ack) {
        publish(&client, &mut entry, &path, true, body)?;
    }
    if entry.ack.comment_id.is_none() {
        return Err(publication_missing(entry.ack.response_status));
    }
    Ok(result(&entry))
}
pub fn reconcile_cli<T: Transport, C: Clock>(
    project: &Path,
    mut request: ReconcileRequest,
    transport: &T,
    clock: &C,
) -> Result<Value> {
    let digest = validate(&mut request.message)?;
    let _body = render(&request.message, &digest, false)?;
    let root = directories(project)?;
    let _lock = lock(&root)?;
    let (outbox, inbox) = paths(&root, &request.message.report_id);
    let mut outgoing = load(&outbox, &request.message, &digest)?;
    let mut incoming = load(&inbox, &request.message, &digest)?;
    let mut probe = new_entry(request.message, digest);
    let client = client(transport, clock);
    verify_target(&client, &probe.message)?;
    let reference = outgoing.as_ref().or(incoming.as_ref()).unwrap_or(&probe);
    let seen = observe(&client, reference)?;
    // Validate both journals against the same snapshot before changing either.
    for entry in outgoing.iter().chain(incoming.iter()) {
        if (entry.record.comment_id.is_some() && seen.record.is_none())
            || (entry.ack.comment_id.is_some() && seen.ack.is_none())
        {
            return Err(err(
                Kind::Blocked,
                "message previously observed publication is now missing",
            ));
        }
    }
    if outgoing.is_none() && incoming.is_none() {
        apply_seen(&mut probe, seen, clock.now());
        if probe.record.comment_id.is_some() {
            save(&inbox, &probe)?;
        }
        return Ok(result(&probe));
    }
    if let Some(entry) = outgoing.as_mut() {
        apply_seen(
            entry,
            Seen {
                record: seen.record,
                ack: seen.ack,
            },
            clock.now(),
        );
        save(&outbox, entry)?;
    }
    if let Some(entry) = incoming.as_mut() {
        apply_seen(entry, seen, clock.now());
        save(&inbox, entry)?;
    }
    Ok(result(
        incoming.as_ref().or(outgoing.as_ref()).unwrap_or(&probe),
    ))
}

/// Verified S1 publications for one task, including server timestamps. No local writes.
pub(crate) struct Publication {
    pub message: Message,
    pub payload_digest: String,
    pub ack: bool,
    pub comment_id: u64,
    pub created_at: i64,
}
pub(crate) fn dispatch_publications<T: Transport, C: Clock>(
    anchor: &Message,
    transport: &T,
    clock: &C,
) -> Result<Vec<Publication>> {
    let client = client(transport, clock);
    verify_target(&client, anchor)?;
    let base = url(anchor);
    let mut next = Some(format!("{base}?per_page=100"));
    let mut visited = BTreeSet::new();
    let mut candidates = Vec::new();
    let mut relevant_ids = BTreeSet::from([anchor.report_id.clone()]);
    let relevant = |m: &Message| {
        m.project_id == anchor.project_id
            && m.task.task_id == anchor.task.task_id
            && (m.event == "dispatch" || m.event == "dispatch-decision")
    };
    // Discover all relevant packet IDs before any context/event filtering. A collision
    // can precede its relevant packet on another page, project, task or message kind.
    while let Some(page) = next {
        if visited.len() >= 1000 || !visited.insert(page.clone()) {
            return Err(err(
                Kind::Blocked,
                "dispatch pagination incomplete or cyclic",
            ));
        }
        if !page.starts_with(&format!("{base}?")) {
            return Err(invalid("dispatch pagination target refused"));
        }
        let response = client.get(&page, None).map_err(remote_error)?;
        let link = next_link(response.header("link"));
        if response
            .header("link")
            .is_some_and(|s| s.contains("rel=\"next\""))
            && link.is_none()
        {
            return Err(err(Kind::Transport, "dispatch malformed pagination link"));
        }
        let body = response
            .body
            .ok_or_else(|| err(Kind::Transport, "dispatch comments missing"))?;
        for comment in body
            .as_array()
            .ok_or_else(|| err(Kind::Transport, "dispatch comments not an array"))?
        {
            let text = comment["body"].as_str().ok_or_else(conflict)?;
            let marks = outbox::parse_op_markers(text.split('\n').next().unwrap_or_default());
            if !marks
                .iter()
                .any(|m| m.operation == "message-record" || m.operation == "message-ack")
            {
                continue;
            }
            let wire = text
                .split('\n')
                .nth(1)
                .and_then(|encoded| serde_json::from_str::<Wire>(encoded).ok());
            if wire.is_none() && marks.iter().any(|m| m.project_id == anchor.project_id) {
                return Err(conflict());
            }
            if let Some(wire) = &wire {
                if relevant(&wire.message) {
                    relevant_ids.insert(wire.message.report_id.clone());
                }
            }
            candidates.push((comment.clone(), marks, wire));
        }
        next = link;
    }
    let mut result = Vec::new();
    for (comment, marks, wire) in candidates {
        if !wire.as_ref().is_some_and(|w| relevant(&w.message))
            && !marks.iter().any(|m| relevant_ids.contains(&m.local_id))
        {
            continue;
        }
        let mut wire = wire.ok_or_else(conflict)?;
        let text = comment["body"].as_str().ok_or_else(conflict)?;
        let digest = validate(&mut wire.message)?;
        let ack = wire.kind == "ack";
        if wire.schema_version != 1
            || !["record", "ack"].contains(&wire.kind.as_str())
            || wire.payload_digest != digest
            || text != render(&wire.message, &digest, ack)?
        {
            return Err(conflict());
        }
        let actor = if ack {
            wire.message.receiver.actor_id
        } else {
            wire.message.sender.actor_id
        };
        if comment["user"]["id"].as_u64() != Some(actor) {
            return Err(conflict());
        }
        let id = comment["id"]
            .as_u64()
            .filter(|id| *id > 0)
            .ok_or_else(conflict)?;
        let at = comment["created_at"]
            .as_str()
            .and_then(parse_utc)
            .ok_or_else(conflict)?;
        // Return colliding foreign-context/event packets too. The consumer must check
        // immutable identities across this set before discarding any packet kind.
        result.push(Publication {
            message: wire.message,
            payload_digest: digest,
            ack,
            comment_id: id,
            created_at: at,
        });
    }
    Ok(result)
}

#[cfg(all(test, unix))]
mod durability_tests {
    use super::*;
    use crate::forge::transport::HttpResponse;
    use std::cell::{Cell, RefCell};
    use tempfile::TempDir;

    #[derive(Default)]
    struct Probe {
        created: Vec<PathBuf>,
        synced: Vec<PathBuf>,
        fail: Option<(PathBuf, usize)>,
    }
    thread_local! {
        static PROBE: RefCell<Probe> = RefCell::new(Probe::default());
    }
    pub(super) fn created(path: &Path) {
        PROBE.with(|p| p.borrow_mut().created.push(path.to_path_buf()));
    }
    // Inject the syscall result before the product converts it, not an early OpError.
    // Tests therefore detect ignoring an actual open/sync result inside the helper too.
    pub(super) fn reported_result(path: &Path) -> std::io::Result<()> {
        PROBE.with(|p| {
            let mut p = p.borrow_mut();
            if let Some((target, skip)) = &mut p.fail {
                if target == path {
                    if *skip == 0 {
                        return Err(std::io::Error::other("injected directory sync failure"));
                    }
                    *skip -= 1;
                }
            }
            Ok(())
        })
    }
    pub(super) fn synced(path: &Path) {
        PROBE.with(|p| p.borrow_mut().synced.push(path.to_path_buf()));
    }
    fn reset(fail: Option<(PathBuf, usize)>) {
        PROBE.with(|p| {
            *p.borrow_mut() = Probe {
                fail,
                ..Probe::default()
            }
        });
    }
    fn assert_synced_chain(project: &Path) {
        PROBE.with(|p| {
            let p = p.borrow();
            for path in [
                project.to_path_buf(),
                project.join(".moai-cli"),
                project.join(".moai-cli/messages"),
                project.join(".moai-cli/messages/outbox"),
                project.join(".moai-cli/messages/inbox"),
            ] {
                assert!(
                    p.synced.contains(&path),
                    "directory not synced before POST: {path:?}"
                );
            }
            for path in &p.created {
                assert!(
                    p.synced.iter().any(|s| Some(s.as_path()) == path.parent()),
                    "new directory parent not synced before POST: {path:?}"
                );
            }
        });
    }
    struct Time;
    impl Clock for Time {
        fn now(&self) -> i64 {
            1_794_355_200
        }
        fn sleep(&self, _: u64) {}
    }
    struct Remote {
        project: RefCell<PathBuf>,
        comments: RefCell<Vec<Value>>,
        actor: Cell<u64>,
        posts: Cell<usize>,
        calls: Cell<usize>,
        hidden_kind: Cell<Option<&'static str>>,
    }
    impl Transport for Remote {
        fn send(&self, request: &HttpRequest) -> std::result::Result<HttpResponse, String> {
            self.calls.set(self.calls.get() + 1);
            if request.method == "POST" {
                let project = self.project.borrow();
                assert_synced_chain(&project);
                let body = request.body.as_ref().unwrap()["body"].as_str().unwrap();
                let ack = body.contains("\"kind\":\"ack\"");
                let entry = journal(&project, ack);
                assert!(if ack {
                    entry.ack.attempted
                } else {
                    entry.record.attempted
                });
                self.posts.set(self.posts.get() + 1);
                let mut comments = self.comments.borrow_mut();
                let id = 100 + comments.len() as u64;
                comments.push(json!({"id":id,"body":body,"user":{"id":self.actor.get()}}));
                return Ok(HttpResponse::new(201).with_body(json!({"id":id})));
            }
            assert_eq!(request.method, "GET");
            let body = if request.url.ends_with("/user") {
                json!({"id":self.actor.get()})
            } else if request.url.contains("/comments?") {
                let comments: Vec<_> = self
                    .comments
                    .borrow()
                    .iter()
                    .filter(|c| {
                        !self.hidden_kind.get().is_some_and(|kind| {
                            c["body"]
                                .as_str()
                                .unwrap()
                                .contains(&format!("\"kind\":\"{kind}\""))
                        })
                    })
                    .cloned()
                    .collect();
                json!(comments)
            } else if request.url.ends_with("/issues/3") {
                json!({"id":2,"node_id":"I_2","number":3})
            } else if request.url.ends_with("/repos/org/repo") {
                json!({"id":1})
            } else {
                panic!("unexpected request")
            };
            Ok(HttpResponse::new(200).with_body(body))
        }
    }
    fn message() -> Message {
        serde_json::from_value(json!({"schema_version":1,"report_id":"r1","project_id":"p1",
            "task":{"owner":"org","repo":"repo","repository_id":1,"issue_id":2,"issue_node_id":"I_2","issue_number":3,"task_id":"T1","revision":"rev1"},
            "sender":{"actor_id":11,"session_id":"sender"},"receiver":{"actor_id":22,"session_id":"receiver"},
            "owner_epoch":1,"leader_epoch":1,"event":"report","candidate_digest":crate::digest(b"candidate"),"created_at":"2026-11-11T00:00:00Z",
            "approval_refs":["issue-comment:10"],"hold_refs":[],"payload":{"text":"hello"}})).unwrap()
    }
    fn invoke(project: &Path, remote: &Remote, ack: bool) -> Result<Value> {
        if ack {
            ack_cli(
                project,
                AckRequest {
                    message: message(),
                    payload_digest: crate::digest(br#"{"text":"hello"}"#),
                },
                "decision:1",
                remote,
                &Time,
            )
        } else {
            record_cli(
                project,
                RecordRequest { message: message() },
                "decision:1",
                remote,
                &Time,
            )
        }
    }
    fn temp() -> TempDir {
        TempDir::new_in(fs::canonicalize(std::env::temp_dir()).unwrap()).unwrap()
    }
    fn setup(ack: bool, reused: bool) -> (TempDir, Remote) {
        let project = temp();
        let remote = Remote {
            project: RefCell::new(project.path().to_path_buf()),
            comments: RefCell::new(vec![]),
            actor: Cell::new(11),
            posts: Cell::new(0),
            calls: Cell::new(0),
            hidden_kind: Cell::new(None),
        };
        if ack {
            let sender = temp();
            *remote.project.borrow_mut() = sender.path().to_path_buf();
            reset(None);
            invoke(sender.path(), &remote, false).unwrap();
            remote.actor.set(22);
            *remote.project.borrow_mut() = project.path().to_path_buf();
        }
        if reused {
            fs::create_dir_all(project.path().join(".moai-cli/messages/outbox")).unwrap();
            fs::create_dir_all(project.path().join(".moai-cli/messages/inbox")).unwrap();
        }
        reset(None);
        (project, remote)
    }
    fn journal(project: &Path, ack: bool) -> Entry {
        let path = project.join(if ack {
            ".moai-cli/messages/inbox/r1.json"
        } else {
            ".moai-cli/messages/outbox/r1.json"
        });
        serde_json::from_slice(&fs::read(path).unwrap()).unwrap()
    }
    fn posting_requires_synced_chain(reused: bool) {
        for ack in [false, true] {
            let (project, remote) = setup(ack, reused);
            let before = remote.posts.get();
            invoke(project.path(), &remote, ack).unwrap();
            assert_eq!(remote.posts.get(), before + 1);
            reset(None);
            invoke(project.path(), &remote, ack).unwrap();
            assert_eq!(remote.posts.get(), before + 1);
            assert_synced_chain(project.path());
        }
    }
    #[test]
    fn audit_new_ancestors_synced_before_post() {
        posting_requires_synced_chain(false);
    }
    #[test]
    fn existing_ancestors_are_resynced_before_post() {
        posting_requires_synced_chain(true);
    }

    #[test]
    fn directory_sync_failure_blocks_new_and_existing_paths_then_restart_recovers() {
        for ack in [false, true] {
            for reused in [false, true] {
                for relative in [
                    "",
                    ".moai-cli",
                    ".moai-cli/messages",
                    ".moai-cli/messages/outbox",
                    ".moai-cli/messages/inbox",
                ] {
                    let (project, remote) = setup(ack, reused);
                    let posts = remote.posts.get();
                    let calls = remote.calls.get();
                    let comments = remote.comments.borrow().clone();
                    reset(Some((project.path().join(relative), 0)));
                    assert_eq!(
                        invoke(project.path(), &remote, ack).unwrap_err().kind,
                        Kind::Io
                    );
                    assert_eq!(remote.posts.get(), posts);
                    assert_eq!(remote.calls.get(), calls);
                    assert_eq!(*remote.comments.borrow(), comments);
                    reset(None);
                    invoke(project.path(), &remote, ack).unwrap();
                    assert_eq!(remote.posts.get(), posts + 1);
                }
            }
        }
    }
    #[test]
    fn intent_sync_failure_can_restart_once_but_attempt_sync_failure_never_reposts() {
        // Record saves the intent, the fresh observation, then the POST attempt.
        for skip in [1, 2, 3] {
            let (project, remote) = setup(false, false);
            reset(Some((
                project.path().join(".moai-cli/messages/outbox"),
                skip,
            )));
            assert_eq!(
                invoke(project.path(), &remote, false).unwrap_err().kind,
                Kind::Io
            );
            assert_eq!(remote.posts.get(), 0);
            assert_eq!(journal(project.path(), false).record.attempted, skip == 3);
            reset(None);
            if skip < 3 {
                invoke(project.path(), &remote, false).unwrap();
                assert_eq!(remote.posts.get(), 1);
            } else {
                assert_eq!(
                    invoke(project.path(), &remote, false).unwrap_err().kind,
                    Kind::Transport
                );
                assert_eq!(remote.posts.get(), 0);
            }
        }
    }
    #[test]
    fn post_sync_failure_keeps_attempt_and_restart_does_not_repeat_hidden_record_or_ack() {
        for ack in [false, true] {
            let (project, remote) = setup(ack, false);
            let posts = remote.posts.get();
            let leaf = if ack { "inbox" } else { "outbox" };
            // Record has one extra pre-attempt observation save; ACK has none.
            let skip = if ack { 3 } else { 4 };
            reset(Some((
                project.path().join(".moai-cli/messages").join(leaf),
                skip,
            )));
            assert_eq!(
                invoke(project.path(), &remote, ack).unwrap_err().kind,
                Kind::Io
            );
            assert_eq!(remote.posts.get(), posts + 1);
            let entry = journal(project.path(), ack);
            assert!(if ack {
                entry.ack.attempted
            } else {
                entry.record.attempted
            });
            remote
                .hidden_kind
                .set(Some(if ack { "ack" } else { "record" }));
            reset(None);
            assert_eq!(
                invoke(project.path(), &remote, ack).unwrap_err().kind,
                Kind::Transport
            );
            assert_eq!(remote.posts.get(), posts + 1);
            remote.hidden_kind.set(None);
            reset(None);
            invoke(project.path(), &remote, ack).unwrap();
            assert_eq!(remote.posts.get(), posts + 1);
        }
    }

    #[test]
    fn reconcile_sync_failure_preserves_receipts_and_performs_no_remote_calls() {
        let (project, remote) = setup(false, false);
        invoke(project.path(), &remote, false).unwrap();
        let path = project.path().join(".moai-cli/messages/outbox/r1.json");
        let before = fs::read(&path).unwrap();
        let calls = remote.calls.get();
        let posts = remote.posts.get();
        reset(Some((project.path().to_path_buf(), 0)));
        assert_eq!(
            reconcile_cli(
                project.path(),
                ReconcileRequest { message: message() },
                &remote,
                &Time
            )
            .unwrap_err()
            .kind,
            Kind::Io
        );
        assert_eq!(remote.calls.get(), calls);
        assert_eq!(fs::read(&path).unwrap(), before);
        reset(None);
        reconcile_cli(
            project.path(),
            ReconcileRequest { message: message() },
            &remote,
            &Time,
        )
        .unwrap();
        assert_eq!(remote.posts.get(), posts);
    }
}
