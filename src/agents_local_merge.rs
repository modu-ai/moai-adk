//! `agents-merge-plan|apply|revert` (#2936 R10, user decision 6880117157): when a project
//! has an `AGENTS.local.md` with text the user wrote, `/moai update` merges that text into
//! `AGENTS.md`, outside the managed section, once.
//!
//! Where the text comes from: the file in place when it holds more than the text MoAI
//! generates (`agents::is_generated_local`); otherwise the copy a committed `go-state-apply`
//! moved into its backup's `root-files/`. Nothing is moved or deleted here: a file in place
//! stays and a copy goes into the receipt folder; the moved copy is already in its dated
//! backup. The text lands at the end of `AGENTS.md` between two marker lines:
//!
//! ```text
//! <!-- moai:merged-local sha256=<sha256 of the source bytes> -->
//! ...the text...
//! <!-- /moai:merged-local -->
//! ```
//!
//! The markers make a second run `unchanged`. A block edited by hand, or a source that changed
//! since the merge, is a `conflict`: nothing is overwritten. The preview shows the whole text,
//! its position and that it goes into a committed file; a line shaped like a credential makes
//! the plan `blocked` (the value is never printed). `plan_digest` binds the source bytes, the
//! `AGENTS.md` bytes and the action.

use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{fs, path::Path};

type Result<T> = std::result::Result<T, String>;

const TARGET: &str = "AGENTS.md";
const LOCAL: &str = "AGENTS.local.md";
const BEGIN: &str = "<!-- moai:merged-local sha256=";
const BEGIN_TAIL: &str = " -->";
const END: &str = "<!-- /moai:merged-local -->";
/// The start of this generation's managed-section markers (`agents.rs`); a source that holds it
/// would be read as a second managed section.
const MANAGED_MARKER: &str = "<!-- moai:managed";
/// The most a source file may hold.
const SOURCE_LIMIT: usize = 128 * 1024;
const RECEIPT_PREFIX: &str = "agents-local-merge-";
/// The one line the preview shows about where the text goes (user decision 6880117157).
pub const COMMIT_NOTICE: &str = "This text goes into AGENTS.md, a file that is committed to the \
     repository; AGENTS.local.md may be a personal file.";

/// Where the text to merge was read.
struct Source {
    /// `in_place` or `backup`.
    kind: &'static str,
    /// Project-relative path of the file read.
    rel: String,
    bytes: Vec<u8>,
}

enum Found {
    Source(Source),
    /// Nothing to merge, with the reason.
    Nothing(String),
    Blocked(String),
}

struct Decision {
    action: &'static str,
    reason: String,
    new_agents: Option<Vec<u8>>,
    starts_at_line: Option<usize>,
    body: Option<String>,
}

impl Decision {
    fn of(action: &'static str, reason: impl Into<String>) -> Decision {
        Decision {
            action,
            reason: reason.into(),
            new_agents: None,
            starts_at_line: None,
            body: None,
        }
    }
}

/// The text with exactly one trailing newline (a source without one gets it).
fn normalized(text: &str) -> String {
    format!("{}\n", text.trim_end_matches('\n'))
}

fn hex(bytes: &[u8]) -> String {
    crate::digest(bytes)
        .trim_start_matches("sha256:")
        .to_string()
}

fn block_of(sha_hex: &str, body: &str) -> String {
    format!("{BEGIN}{sha_hex}{BEGIN_TAIL}\n{body}{END}\n")
}

/// The block already in `AGENTS.md`: the sha256 its marker names and the text between the marks.
struct Existing {
    sha: String,
    body: String,
}

/// `Ok(None)`: no block. `Err`: a block that cannot be read as one (the plan is a conflict).
fn find_block(agents: &str) -> std::result::Result<Option<Existing>, String> {
    let starts: Vec<usize> = agents.match_indices(BEGIN).map(|(i, _)| i).collect();
    let start = match starts.as_slice() {
        // An end marker alone means the start marker was deleted: the text may still be
        // there, and merging again would double it (#2982 audit N5).
        [] if agents.contains(END) => {
            return Err("the merged-local end marker has no start marker".into())
        }
        [] => return Ok(None),
        [one] => *one,
        _ => return Err("AGENTS.md holds more than one merged-local block".into()),
    };
    if start != 0 && !agents[..start].ends_with('\n') {
        return Err("the merged-local marker is not at the start of a line".into());
    }
    let after = &agents[start + BEGIN.len()..];
    let damaged = || "the merged-local marker is damaged".to_string();
    let close = after.find(BEGIN_TAIL).ok_or_else(damaged)?;
    let sha = &after[..close];
    if sha.len() != 64 || !sha.bytes().all(|b| b.is_ascii_hexdigit()) {
        return Err(damaged());
    }
    let rest = after[close + BEGIN_TAIL.len()..]
        .strip_prefix('\n')
        .ok_or_else(damaged)?;
    if rest.matches(END).count() != 1 {
        return Err("the merged-local block needs exactly one end marker".into());
    }
    let end = rest.find(END).ok_or_else(damaged)?;
    let body = &rest[..end];
    if !(body.is_empty() || body.ends_with('\n')) {
        return Err("the merged-local end marker is not at the start of a line".into());
    }
    Ok(Some(Existing {
        sha: sha.to_ascii_lowercase(),
        body: body.to_string(),
    }))
}

/// Lines of `text` shaped like a credential: `(line number, kind)`. The value is never kept.
fn secret_suspects(text: &str) -> Vec<(usize, &'static str)> {
    text.lines()
        .enumerate()
        .filter_map(|(i, line)| secret_kind(line).map(|kind| (i + 1, kind)))
        .collect()
}

/// A token of `prefix` followed by at least `min` token characters, the prefix not inside a word.
fn has_token(line: &str, prefix: &str, min: usize, upper_digit_only: bool) -> bool {
    line.match_indices(prefix).any(|(i, _)| {
        let word_before = line[..i]
            .chars()
            .next_back()
            .is_some_and(|c| c.is_ascii_alphanumeric());
        let run = line[i + prefix.len()..]
            .chars()
            .take_while(|c| {
                if upper_digit_only {
                    c.is_ascii_uppercase() || c.is_ascii_digit()
                } else {
                    c.is_ascii_alphanumeric() || *c == '-' || *c == '_'
                }
            })
            .count();
        !word_before && run >= min
    })
}

fn secret_kind(line: &str) -> Option<&'static str> {
    let lower = line.to_ascii_lowercase();
    if lower.contains("-----begin") && lower.contains("private key") {
        return Some("private_key");
    }
    for (prefix, min, upper, kind) in [
        ("sk-", 20, false, "api_key"),
        ("ghp_", 20, false, "github_token"),
        ("gho_", 20, false, "github_token"),
        ("github_pat_", 20, false, "github_token"),
        ("xoxb-", 10, false, "slack_token"),
        ("xoxp-", 10, false, "slack_token"),
        ("ghs_", 20, false, "github_token"),
        ("ghu_", 20, false, "github_token"),
        ("glpat-", 20, false, "gitlab_token"),
        ("sk_live_", 16, false, "api_key"),
        ("rk_live_", 16, false, "api_key"),
        ("xoxa-", 10, false, "slack_token"),
        ("xoxr-", 10, false, "slack_token"),
        ("xoxs-", 10, false, "slack_token"),
        ("xapp-", 10, false, "slack_token"),
        ("AIza", 35, false, "api_key"),
        ("hf_", 30, false, "api_key"),
        ("AKIA", 16, true, "aws_key"),
        ("ASIA", 16, true, "aws_key"),
    ] {
        if has_token(line, prefix, min, upper) {
            return Some(kind);
        }
    }
    // A JSON web token: a header and a payload, both starting `eyJ`.
    if line.contains(".eyJ") && has_token(line, "eyJ", 10, false) {
        return Some("jwt");
    }
    // `Authorization: Bearer <token>` or `Basic <base64>`.
    if lower.contains("authorization") {
        for word in ["bearer ", "basic "] {
            for (i, _) in lower.match_indices(word).take(64) {
                let run = line[i + word.len()..]
                    .chars()
                    .take(64)
                    .take_while(|c| c.is_ascii_alphanumeric() || "-._~+/=".contains(*c))
                    .count();
                if run >= 16 {
                    return Some("authorization_header");
                }
            }
        }
    }
    // A URL that carries a password: `scheme://user:password@host`.
    for (i, _) in line.match_indices("://").take(64) {
        let authority: &str = line[i + 3..]
            .split(|c: char| c == '/' || c.is_whitespace())
            .next()
            .unwrap_or("");
        if let Some((userinfo, _)) = authority.split_once('@') {
            if let Some((_, password)) = userinfo.split_once(':') {
                if password.len() >= 3 && !password.starts_with(['<', '$', '{', '*', '[', '(']) {
                    return Some("url_credential");
                }
            }
        }
    }
    for key in [
        "password",
        "passwd",
        "secret",
        "token",
        "api_key",
        "apikey",
        "api-key",
        "secret_key",
        "secret_access_key",
        "access_key",
        "private_key",
    ] {
        // A bounded number of hits and of value characters per line: one huge line must not
        // cost time quadratic in its length (#2982 audit N6).
        for (i, _) in lower.match_indices(key).take(64) {
            let tail = lower[i + key.len()..].trim_start_matches(['"', '\'', '`', ' ', '\t']);
            let Some(tail) = tail.strip_prefix('=').or_else(|| tail.strip_prefix(':')) else {
                continue;
            };
            let at = line.len() - tail.len();
            let value: String = line[at..]
                .trim_start_matches([' ', '\t', '"', '\'', '`'])
                .chars()
                .take(256)
                .take_while(|c| !c.is_whitespace() && !matches!(c, '"' | '\'' | '`'))
                .collect();
            let looks_real = value.len() >= 8
                && value.is_ascii()
                && !value.starts_with(['<', '$', '{', '*', '[', '('])
                && value
                    .bytes()
                    .any(|b| b.is_ascii_digit() || b.is_ascii_uppercase());
            if looks_real {
                return Some("credential_assignment");
            }
        }
    }
    None
}

/// The plan's decision from the two files and the credential scan. Pure.
fn decide(agents: Option<&[u8]>, found: Found, suspects: &[(usize, &'static str)]) -> Decision {
    let src = match found {
        Found::Nothing(reason) => return Decision::of("none", reason),
        Found::Blocked(reason) => return Decision::of("blocked", reason),
        Found::Source(s) => s,
    };
    let Ok(text) = std::str::from_utf8(&src.bytes) else {
        return Decision::of("blocked", format!("{} is not UTF-8 text", src.rel));
    };
    if text.contains(BEGIN) || text.contains(END) || text.contains(MANAGED_MARKER) {
        return Decision::of(
            "blocked",
            format!(
                "{} holds a MoAI marker line; merged, it would be read as a managed block",
                src.rel
            ),
        );
    }
    let Some(agents) = agents else {
        return Decision::of(
            "blocked",
            "AGENTS.md does not exist yet; re-run after the install phase",
        );
    };
    let Ok(agents) = std::str::from_utf8(agents) else {
        return Decision::of("blocked", "AGENTS.md is not UTF-8 text");
    };
    let body = normalized(text);
    let sha = hex(&src.bytes);
    match find_block(agents) {
        Err(reason) => Decision::of("conflict", format!("{reason}; nothing is written")),
        Ok(Some(e)) if e.sha == sha && e.body == body => {
            Decision::of("unchanged", "AGENTS.md already holds this text")
        }
        Ok(Some(e)) if e.sha == sha => Decision::of(
            "conflict",
            "the merged block was edited by hand; it is kept and nothing is written",
        ),
        Ok(Some(_)) => Decision::of(
            "conflict",
            format!(
                "{} changed since it was merged; the block is kept and nothing is written",
                src.rel
            ),
        ),
        Ok(None) if !suspects.is_empty() => Decision::of(
            "blocked",
            format!(
                "{} line(s) of {} look like a credential (lines {}); remove them from the file \
                 and preview again",
                suspects.len(),
                src.rel,
                suspects
                    .iter()
                    .map(|(n, _)| n.to_string())
                    .collect::<Vec<_>>()
                    .join(", ")
            ),
        ),
        Ok(None) => {
            let separator = if agents.is_empty() || agents.ends_with("\n\n") {
                ""
            } else if agents.ends_with('\n') {
                "\n"
            } else {
                "\n\n"
            };
            let head = format!("{agents}{separator}");
            let starts_at_line = head.matches('\n').count() + 1;
            let new_agents = format!("{head}{}", block_of(&sha, &body)).into_bytes();
            // The plan, the revert and every later read of AGENTS.md go through
            // `apply::read_opt`, which refuses a file above this size (#2982 audit B1).
            if new_agents.len() as u64 > crate::apply::RECEIPT_LIMIT {
                return Decision::of(
                    "blocked",
                    format!(
                        "AGENTS.md would grow to {} bytes with this text; the limit is {} bytes",
                        new_agents.len(),
                        crate::apply::RECEIPT_LIMIT
                    ),
                );
            }
            Decision {
                action: "merge",
                reason: "the text of AGENTS.local.md is not in AGENTS.md yet".into(),
                new_agents: Some(new_agents),
                starts_at_line: Some(starts_at_line),
                body: Some(body),
            }
        }
    }
}

/// What the personal file in place holds: text the user wrote, nothing, or MoAI's own text.
enum Content {
    Personal,
    Blank,
    Generated,
}

fn content_of(bytes: &[u8]) -> Content {
    if bytes.iter().all(u8::is_ascii_whitespace) {
        Content::Blank
    } else if crate::agents::is_generated_local(bytes) {
        Content::Generated
    } else {
        Content::Personal
    }
}

fn read_checked(root: &Path, rel: &str) -> std::result::Result<Option<Vec<u8>>, String> {
    let path = crate::apply::target(root, rel).map_err(|e| format!("{rel}: {e}"))?;
    let bytes = crate::apply::read_opt(&path).map_err(|e| format!("{rel}: {e}"))?;
    match bytes {
        Some(b) if b.len() > SOURCE_LIMIT => Err(format!(
            "{rel} is larger than {} KiB; not merged",
            SOURCE_LIMIT / 1024
        )),
        other => Ok(other),
    }
}

fn find_source(root: &Path) -> Result<Found> {
    let mut why_not = format!("no {LOCAL} and no backup copy of one");
    match read_checked(root, LOCAL) {
        Err(reason) => return Ok(Found::Blocked(reason)),
        Ok(None) => {}
        Ok(Some(bytes)) => match content_of(&bytes) {
            Content::Personal => {
                return Ok(Found::Source(Source {
                    kind: "in_place",
                    rel: LOCAL.into(),
                    bytes,
                }))
            }
            Content::Blank => why_not = format!("{LOCAL} is empty"),
            Content::Generated => {
                why_not = format!("{LOCAL} only holds the text MoAI generated");
            }
        },
    }
    let backed = match crate::go_state::backed_up_root_file(root, LOCAL) {
        Ok(b) => b,
        Err(reason) => return Ok(Found::Blocked(reason)),
    };
    let Some(backed) = backed else {
        return Ok(Found::Nothing(why_not));
    };
    let bytes = match read_checked(root, &backed.rel) {
        Err(reason) => return Ok(Found::Blocked(reason)),
        Ok(None) => {
            return Ok(Found::Blocked(format!(
                "{} is missing; the go-state receipt names it",
                backed.rel
            )))
        }
        Ok(Some(b)) => b,
    };
    if crate::digest(&bytes) != backed.sha256 {
        return Ok(Found::Blocked(format!(
            "{} changed since the go-state apply moved it; not merged",
            backed.rel
        )));
    }
    Ok(match content_of(&bytes) {
        Content::Personal => Found::Source(Source {
            kind: "backup",
            rel: backed.rel,
            bytes,
        }),
        Content::Blank => Found::Nothing(format!("the backup copy of {LOCAL} is empty")),
        Content::Generated => Found::Nothing(format!(
            "the backup copy of {LOCAL} only holds the text MoAI generated"
        )),
    })
}

struct Inner {
    json: Value,
    action: &'static str,
    reason: String,
    before: Option<Vec<u8>>,
    new_agents: Option<Vec<u8>>,
    source: Option<Source>,
}

fn plan_inner(root: &Path) -> Result<Inner> {
    let agents_path = crate::apply::target(root, TARGET)?;
    let before = crate::apply::read_opt(&agents_path)?;
    let found = find_source(root)?;
    let suspects = match &found {
        Found::Source(s) => std::str::from_utf8(&s.bytes)
            .map(secret_suspects)
            .unwrap_or_default(),
        _ => Vec::new(),
    };
    let decision = decide(before.as_deref(), clone_found(&found), &suspects);
    let source = match found {
        Found::Source(s) => Some(s),
        _ => None,
    };
    let source_json = source.as_ref().map(|s| {
        json!({
            "kind": s.kind,
            "path": s.rel,
            "sha256": crate::digest(&s.bytes),
            "bytes": s.bytes.len(),
        })
    });
    let suspects_json: Vec<Value> = suspects
        .iter()
        .map(|(line, kind)| json!({"line": line, "kind": kind}))
        .collect();
    let bound = json!({
        "action": decision.action,
        "source": source_json,
        "agents_sha256": before.as_deref().map(crate::digest),
        "secret_suspects": suspects_json,
    });
    let digest = crate::digest(&serde_json::to_vec(&bound).map_err(|_| "encode failed")?);
    let mut json = json!({
        "project": root.display().to_string(),
        "action": decision.action,
        "reason": decision.reason,
        "source": source_json,
        "target": {
            "path": TARGET,
            "exists": before.is_some(),
            "sha256_before": before.as_deref().map(crate::digest),
            "starts_at_line": decision.starts_at_line,
        },
        "commit_notice": COMMIT_NOTICE,
        "secret_suspects": suspects_json,
        "body": decision.body,
        "plan_digest": digest,
    });
    json["next_actions"] = if decision.action == "merge" {
        json!([
            "show the user the body, the position and the commit_notice; after their yes run \
             `moai-cli agents-merge-apply --plan-digest <plan_digest> \
             --permission-reference <ref> <project>`"
        ])
    } else {
        json!([])
    };
    Ok(Inner {
        json,
        action: decision.action,
        reason: decision.reason,
        before,
        new_agents: decision.new_agents,
        source,
    })
}

/// `decide` consumes its `Found`; the plan keeps the source for the receipt, so it decides on
/// a copy.
fn clone_found(found: &Found) -> Found {
    match found {
        Found::Source(s) => Found::Source(Source {
            kind: s.kind,
            rel: s.rel.clone(),
            bytes: s.bytes.clone(),
        }),
        Found::Nothing(r) => Found::Nothing(r.clone()),
        Found::Blocked(r) => Found::Blocked(r.clone()),
    }
}

/// The read-only preview. Writes nothing.
pub fn plan(project: &Path) -> Result<Value> {
    let root = fs::canonicalize(project).map_err(|_| "project folder unavailable")?;
    plan_inner(&root).map(|i| i.json)
}

#[derive(Serialize, Deserialize)]
struct Receipt {
    schema_version: u32,
    id: String,
    at: String,
    plan_digest: String,
    permission_reference: String,
    /// `pending`, `applied`, `failed` or `reverted`.
    status: String,
    agents_before: String,
    agents_after: Option<String>,
    block_sha256: String,
    source_kind: String,
    source_path: String,
    source_sha256: String,
}

fn receipts_dir(root: &Path) -> std::path::PathBuf {
    crate::state_root::path(root, "receipts")
}

/// Owner-only bits for a receipt folder and for what it holds: the receipt keeps a copy of a
/// personal file, which may hold a token (#2982 audit B2; the same bits as `go_state`'s backup).
const PRIVATE_DIR: u32 = 0o700;
const PRIVATE_FILE: u32 = 0o600;

/// Create one folder (not its parents) that only the owner can enter, from its creation.
fn create_private_dir(path: &Path) -> std::io::Result<()> {
    let mut builder = fs::DirBuilder::new();
    #[cfg(unix)]
    {
        use std::os::unix::fs::DirBuilderExt;
        builder.mode(PRIVATE_DIR);
    }
    builder.create(path)
}

/// Create `path` and the folders below `base` on the way to it, each one private. A link on
/// the way is refused.
fn create_private_dirs(base: &Path, rel: &[&str]) -> Result<std::path::PathBuf> {
    let mut cur = base.to_path_buf();
    for part in rel {
        cur.push(part);
        match fs::symlink_metadata(&cur) {
            Ok(m) if m.file_type().is_symlink() => return Err("symlink target refused".into()),
            Ok(m) if m.is_dir() => {}
            Ok(_) => return Err("receipt folder unavailable".into()),
            Err(_) => create_private_dir(&cur).map_err(|_| "receipt folder unavailable")?,
        }
    }
    Ok(cur)
}

/// A new file in a receipt folder: owner-only from its creation, never replacing a file.
fn write_private(path: &Path, bytes: &[u8]) -> Result<()> {
    crate::apply::publish_exclusive_mode(path, bytes, Some(PRIVATE_FILE))
        .map_err(|_| "receipt write failed".into())
}

fn new_receipt_dir(root: &Path) -> Result<(String, std::path::PathBuf)> {
    let state = crate::state_root::path(root, "");
    crate::claude::refuse_symlink_dir(&state)?;
    fs::create_dir_all(&state).map_err(|_| "receipt folder unavailable")?;
    let receipts = create_private_dirs(&state, &["receipts"])?;
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    for n in 0..16 {
        let id = format!("{RECEIPT_PREFIX}{nanos}-{n}");
        let dir = receipts.join(&id);
        if create_private_dir(&dir).is_ok() {
            return Ok((id, dir));
        }
    }
    Err("receipt id unavailable".into())
}

fn write_receipt(dir: &Path, r: &Receipt) -> Result<()> {
    let mut bytes = serde_json::to_vec_pretty(r).map_err(|_| "encode failed")?;
    bytes.push(b'\n');
    let path = dir.join("receipt.json");
    match crate::claude::read_existing(&path)? {
        // The first write creates the file owner-only; a later one keeps the bits it has.
        None => write_private(&path, &bytes),
        Some(before) => crate::apply::publish_replace(&path, Some(&before), &bytes),
    }
}

/// Merges the text after the user confirmed the plan whose digest is `plan_digest`.
pub fn apply(project: &Path, plan_digest: &str, permission: &str) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("a permission reference is required; nothing was written".into());
    }
    let root = fs::canonicalize(project).map_err(|_| "project folder unavailable")?;
    let inner = plan_inner(&root)?;
    if inner.json["plan_digest"].as_str() != Some(plan_digest) {
        return Err("stale or altered plan; re-preview required".into());
    }
    match inner.action {
        "merge" => {}
        "none" | "unchanged" => return Ok(json!({"state": "unchanged", "writes": []})),
        other => return Err(format!("the plan is {other}: {}", inner.reason)),
    }
    let (Some(before), Some(new_agents), Some(source)) =
        (inner.before, inner.new_agents, inner.source)
    else {
        return Err("the plan holds no merge".into());
    };
    if new_agents.len() as u64 > crate::apply::RECEIPT_LIMIT {
        return Err(
            "the merged AGENTS.md would be larger than the limit; nothing was written".into(),
        );
    }
    let agents_path = crate::apply::target(&root, TARGET)?;
    let (id, dir) = new_receipt_dir(&root)?;
    let keep = create_private_dirs(&dir, &["before"])?;
    write_private(&keep.join(TARGET), &before)?;
    // A file in place stays; the receipt keeps the copy that serves as its backup. A moved
    // copy is already in its dated go-state backup.
    if source.kind == "in_place" {
        let copy = create_private_dirs(&dir, &["source"])?;
        write_private(&copy.join(LOCAL), &source.bytes)?;
    }
    let mut receipt = Receipt {
        schema_version: 1,
        id: id.clone(),
        at: crate::baseline::utc_now(),
        plan_digest: plan_digest.to_string(),
        permission_reference: permission.to_string(),
        status: "pending".into(),
        agents_before: crate::digest(&before),
        // Known before the write, so a revert can tell an interrupted apply that did write
        // from one that did not (#2982 audit N2).
        agents_after: Some(crate::digest(&new_agents)),
        block_sha256: hex(&source.bytes),
        source_kind: source.kind.into(),
        source_path: source.rel.clone(),
        source_sha256: crate::digest(&source.bytes),
    };
    write_receipt(&dir, &receipt)?;
    if let Err(e) = crate::apply::publish_replace(&agents_path, Some(&before), &new_agents) {
        receipt.status = "failed".into();
        receipt.agents_after = None;
        write_receipt(&dir, &receipt)?;
        return Err(format!("AGENTS.md was not written ({e})"));
    }
    receipt.status = "applied".into();
    write_receipt(&dir, &receipt)?;
    Ok(json!({
        "state": "applied",
        "project": root.display().to_string(),
        "source": {"kind": source.kind, "path": source.rel},
        "starts_at_line": inner.json["target"]["starts_at_line"],
        "receipt": id,
        "writes": [TARGET],
        "next_actions": [format!(
            "to undo: moai-cli agents-merge-revert --receipt {id} --permission-reference <ref> <project>"
        )],
    }))
}

/// Undoes one merge while `AGENTS.md` still holds the bytes the apply wrote: it gets its
/// earlier bytes back. The source file and its backup are not touched.
pub fn revert(project: &Path, receipt_id: &str, permission: &str) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("a permission reference is required; nothing was written".into());
    }
    if !(receipt_id.starts_with(RECEIPT_PREFIX)
        && receipt_id.len() <= 64
        && receipt_id
            .bytes()
            .all(|b| b.is_ascii_alphanumeric() || b == b'-'))
    {
        return Err("unknown receipt id".into());
    }
    let root = fs::canonicalize(project).map_err(|_| "project directory unavailable")?;
    let dir = receipts_dir(&root).join(receipt_id);
    for d in [
        crate::state_root::path(&root, ""),
        receipts_dir(&root),
        dir.clone(),
    ] {
        crate::claude::refuse_symlink_dir(&d)?;
    }
    let bytes =
        crate::claude::read_existing(&dir.join("receipt.json"))?.ok_or("receipt not found")?;
    let mut r: Receipt = serde_json::from_slice(&bytes).map_err(|_| "receipt is unreadable")?;
    let agents_path = crate::apply::target(&root, TARGET)?;
    match r.status.as_str() {
        "applied" => {}
        "reverted" => return Ok(json!({"state": "unchanged", "writes": []})),
        // An apply that was interrupted: AGENTS.md holds either the bytes it was about to
        // write (undo like an applied one) or its earlier bytes (it wrote nothing).
        "pending" => {
            let now = crate::apply::read_opt(&agents_path)?.ok_or("AGENTS.md is missing")?;
            let held = crate::digest(&now);
            if held == r.agents_before {
                r.status = "failed".into();
                r.agents_after = None;
                write_receipt(&dir, &r)?;
                return Ok(json!({"state": "unchanged", "writes": []}));
            }
            // Any other bytes fall to the check below: AGENTS.md changed; nothing is written.
        }
        _ => return Err("that apply wrote nothing; there is nothing to undo".into()),
    }
    let now = crate::apply::read_opt(&agents_path)?.ok_or("AGENTS.md is missing")?;
    if r.agents_after.as_deref() != Some(crate::digest(&now).as_str()) {
        return Err("AGENTS.md changed since the merge; nothing was written".into());
    }
    crate::claude::refuse_symlink_dir(&dir.join("before"))?;
    let before = crate::claude::read_existing(&dir.join("before").join(TARGET))?
        .ok_or("the receipt has no copy of the earlier AGENTS.md")?;
    if crate::digest(&before) != r.agents_before {
        return Err(
            "the receipt's copy of the earlier AGENTS.md was altered; nothing was written".into(),
        );
    }
    crate::apply::publish_replace(&agents_path, Some(&now), &before)?;
    r.status = "reverted".into();
    write_receipt(&dir, &r)?;
    Ok(json!({
        "state": "reverted",
        "project": root.display().to_string(),
        "receipt": receipt_id,
        "writes": [TARGET],
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn src(text: &str) -> Found {
        Found::Source(Source {
            kind: "in_place",
            rel: LOCAL.into(),
            bytes: text.as_bytes().to_vec(),
        })
    }

    fn action(agents: Option<&str>, found: Found) -> &'static str {
        let suspects = match &found {
            Found::Source(s) => secret_suspects(std::str::from_utf8(&s.bytes).unwrap()),
            _ => Vec::new(),
        };
        decide(agents.map(str::as_bytes), found, &suspects).action
    }

    /// Credential-shaped text is built here, not written out, so no fixture in the source is
    /// itself a secret.
    fn shaped(prefix: &str, run: usize) -> String {
        format!("{prefix}{}", "a".repeat(run))
    }

    #[test]
    fn the_text_lands_after_a_blank_line_at_the_end_with_its_hash() {
        let d = decide(Some(b"# Rules\n"), src("mine"), &[]);
        assert_eq!(d.action, "merge");
        let sha = hex(b"mine");
        assert_eq!(
            String::from_utf8(d.new_agents.unwrap()).unwrap(),
            format!("# Rules\n\n{BEGIN}{sha}{BEGIN_TAIL}\nmine\n{END}\n")
        );
        assert_eq!(d.starts_at_line, Some(3));
    }

    #[test]
    fn the_separator_adapts_to_how_agents_md_ends() {
        let at = |agents: &str| decide(Some(agents.as_bytes()), src("x\n"), &[]).starts_at_line;
        assert_eq!(at(""), Some(1));
        assert_eq!(at("a\n\n"), Some(3));
        assert_eq!(at("a\n"), Some(3));
        assert_eq!(at("a"), Some(3));
    }

    #[test]
    fn a_second_run_is_unchanged_and_an_edit_is_a_conflict() {
        let first = decide(Some(b"# Rules\n"), src("mine"), &[]);
        let merged = String::from_utf8(first.new_agents.unwrap()).unwrap();
        assert_eq!(action(Some(&merged), src("mine")), "unchanged");
        // A source that gained a trailing newline has the same normalized body but a new hash.
        assert_eq!(action(Some(&merged), src("mine\n")), "conflict");
        assert_eq!(action(Some(&merged), src("changed")), "conflict");
        let edited = merged.replace("mine", "edited");
        assert_eq!(action(Some(&edited), src("mine")), "conflict");
    }

    #[test]
    fn a_damaged_or_doubled_block_is_a_conflict() {
        let sha = hex(b"mine");
        let good = format!("{BEGIN}{sha}{BEGIN_TAIL}\nmine\n{END}\n");
        assert_eq!(action(Some(&good), src("mine")), "unchanged");
        assert_eq!(
            action(Some(&format!("{good}{good}")), src("mine")),
            "conflict"
        );
        let short = good.replace(&sha, "abc");
        assert_eq!(action(Some(&short), src("mine")), "conflict");
        let no_end = good.replace(END, "");
        assert_eq!(action(Some(&no_end), src("mine")), "conflict");
        let mid_line = format!("text {good}");
        assert_eq!(action(Some(&mid_line), src("mine")), "conflict");
        // Only the start marker deleted: merging again would double the text.
        let no_start: String = good.lines().skip(1).map(|l| format!("{l}\n")).collect();
        assert_eq!(action(Some(&no_start), src("mine")), "conflict");
    }

    /// The merged `AGENTS.md` must stay readable by `apply::read_opt` (#2982 audit B1).
    #[test]
    fn the_merged_file_may_not_pass_the_receipt_limit() {
        let limit = crate::apply::RECEIPT_LIMIT as usize;
        let block_len = block_of(&hex(b"x"), "x\n").len();
        let agents_of = |extra: usize| format!("{}\n\n", "a".repeat(limit - 2 - block_len + extra));
        let merged_len = |extra: usize| {
            decide(Some(agents_of(extra).as_bytes()), src("x"), &[])
                .new_agents
                .map(|b| b.len())
        };
        assert_eq!(merged_len(0), Some(limit));
        let over = decide(Some(agents_of(1).as_bytes()), src("x"), &[]);
        assert_eq!(over.action, "blocked");
        assert!(over.new_agents.is_none());
        assert!(
            over.reason.contains(&(limit + 1).to_string()),
            "{}",
            over.reason
        );
    }

    #[test]
    fn nothing_to_merge_and_waiting_cases() {
        assert_eq!(action(Some("a\n"), Found::Nothing("none".into())), "none");
        assert_eq!(action(Some("a\n"), Found::Blocked("why".into())), "blocked");
        assert_eq!(action(None, src("mine")), "blocked");
        assert_eq!(action(Some("a\n"), src("a\u{0}b")), "merge");
    }

    #[test]
    fn a_source_with_a_moai_marker_is_blocked() {
        for marker in [
            "<!-- moai:managed:begin x -->".to_string(),
            END.to_string(),
            format!("{BEGIN}{}{BEGIN_TAIL}", hex(b"x")),
        ] {
            assert_eq!(
                action(Some("a\n"), src(&format!("mine\n{marker}\n"))),
                "blocked"
            );
        }
    }

    #[test]
    fn credential_shaped_lines_block_the_merge() {
        let key = format!("key: {}", shaped("sk-", 24));
        assert_eq!(action(Some("a\n"), src(&key)), "blocked");
        let assignment = format!("a\npassword = {}", "Hunter2".repeat(2));
        assert_eq!(action(Some("a\n"), src(&assignment)), "blocked");
        let pem = format!("note\n{}{} RSA PRIVATE KEY{}", "-----", "BEGIN", "-----");
        assert_eq!(action(Some("a\n"), src(&pem)), "blocked");
        // Shapes that are not credentials: a short value, a placeholder, a prefix inside a word.
        assert_eq!(action(Some("a\n"), src("password: short")), "merge");
        assert_eq!(
            action(Some("a\n"), src("token: <your-token-here>")),
            "merge"
        );
        assert_eq!(action(Some("a\n"), src(&shaped("task-", 26))), "merge");
        // An existing block is judged without the scan: the text is already in the file.
        let sha = hex(key.as_bytes());
        let have = format!("{BEGIN}{sha}{BEGIN_TAIL}\n{key}\n{END}\n");
        assert_eq!(action(Some(&have), src(&key)), "unchanged");
    }

    #[test]
    fn suspects_report_the_line_and_kind_only() {
        let text = format!(
            "ok\n{}\n{}\n",
            shaped("ghp_", 30),
            "AKIA".to_string() + &"A".repeat(16)
        );
        assert_eq!(
            secret_suspects(&text),
            vec![(2, "github_token"), (3, "aws_key")]
        );
        assert_eq!(secret_suspects(&shaped("sk-", 19)), vec![]);
        assert_eq!(secret_suspects(&shaped("sk-", 20)), vec![(1, "api_key")]);
    }

    #[test]
    fn token_shapes_have_an_exact_minimum_length() {
        // (prefix, shortest run that counts, kind); one fewer is not a token.
        for (prefix, min, kind) in [
            ("AKIA", 16, "aws_key"),
            ("ASIA", 16, "aws_key"),
            ("ghs_", 20, "github_token"),
            ("ghu_", 20, "github_token"),
            ("glpat-", 20, "gitlab_token"),
            ("sk_live_", 16, "api_key"),
            ("xapp-", 10, "slack_token"),
            ("xoxa-", 10, "slack_token"),
            ("AIza", 35, "api_key"),
            ("hf_", 30, "api_key"),
        ] {
            let body = |n: usize| {
                if prefix.starts_with("AKIA") || prefix.starts_with("ASIA") {
                    format!("{prefix}{}", "A".repeat(n))
                } else {
                    shaped(prefix, n)
                }
            };
            assert_eq!(secret_suspects(&body(min)), vec![(1, kind)], "{prefix}");
            assert_eq!(secret_suspects(&body(min - 1)), vec![], "{prefix} short");
        }
    }

    #[test]
    fn headers_urls_and_web_tokens_are_credentials() {
        let bearer = |n: usize| format!("Authorization: {} {}", "Bearer", "a".repeat(n));
        assert_eq!(
            secret_suspects(&bearer(16)),
            vec![(1, "authorization_header")]
        );
        assert_eq!(secret_suspects(&bearer(15)), vec![]);
        // The word alone, without the header name, is prose.
        assert_eq!(
            secret_suspects(&format!("Bearer {}", "a".repeat(30))),
            vec![]
        );

        let url = |password: &str| format!("DB: {}://app:{password}@db.example/main", "postgres");
        assert_eq!(secret_suspects(&url("pw3")), vec![(1, "url_credential")]);
        assert_eq!(secret_suspects(&url("pw")), vec![]);
        assert_eq!(secret_suspects(&url("<password>")), vec![]);
        assert_eq!(
            secret_suspects("see https://example.com/a:b@c for details"),
            vec![]
        );

        let jwt = format!("{}.{}.sig", shaped("eyJ", 12), shaped("eyJ", 12));
        assert_eq!(secret_suspects(&jwt), vec![(1, "jwt")]);
        assert_eq!(secret_suspects(&shaped("eyJ", 12)), vec![]);
    }

    #[test]
    fn assignments_need_a_value_of_eight_characters_with_a_digit_or_capital() {
        for key in [
            "password",
            "SECRET_KEY",
            "AWS_SECRET_ACCESS_KEY",
            "private_key",
        ] {
            assert_eq!(
                secret_suspects(&format!("{key}=Abcdef12")),
                vec![(1, "credential_assignment")],
                "{key}"
            );
            assert_eq!(
                secret_suspects(&format!("{key}=Abcdef1")),
                vec![],
                "{key} short"
            );
        }
        assert_eq!(secret_suspects("password: abcdefghij"), vec![]);
    }

    /// A single huge line must not cost time quadratic in its length (#2982 audit N6; it took
    /// about 43 seconds in a debug build before the scan was bounded).
    #[test]
    fn one_huge_line_is_scanned_in_bounded_time() {
        let line = "password:".repeat(SOURCE_LIMIT / 9);
        let start = std::time::Instant::now();
        let _ = secret_suspects(&line);
        let spaced = format!("password{}x", " ".repeat(SOURCE_LIMIT - 10));
        let _ = secret_suspects(&spaced);
        assert!(start.elapsed().as_secs() < 5, "{:?}", start.elapsed());
    }
}
