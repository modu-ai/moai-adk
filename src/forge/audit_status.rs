//! The `moai/audit` commit status, posted by an audit-only GitHub App when its token is set
//! (#2173, B6; decisions 01:1x(11-09) "Forge permissions and fallback").
//!
//! `forge-audit-status` posts one status on an exact head SHA and reads it back once:
//!
//! - The context is the constant `moai/audit`, the context the integration ruleset requires.
//!   It is never caller-settable; another context would be a different check.
//! - WHEN `MOAI_AUDIT_APP_TOKEN` is set, its value is handed to gh as `GH_TOKEN` in the child's
//!   environment only, and the description carries `source=app`. Otherwise gh's own (shared
//!   account) authentication posts it, the description carries `source=shared-account`, and a
//!   warning says the source is not pinned (REQ-B6-01). Only the variable name is ever output.
//! - Pinning the source is a ruleset setting (the required check's expected source app) and
//!   App registration, installation and token minting are operator steps (keep-set); none of
//!   them is done here.
//!
//! The POST is a write: it needs a permission reference, like `forge-report`. A status is
//! idempotent per (sha, context), so no outbox marker is kept; the read-back GET reports what
//! GitHub shows, or `uncertain`.
use super::policy::Clock;
use super::transport::{GhTransport, HttpRequest, Transport};
use super::{Client, Config, ForgeError, API, SCHEMA_VERSION};
use serde::Deserialize;
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::path::Path;

pub const CONTEXT: &str = "moai/audit";
/// The only place the App token is read from. Its value is never logged or stored.
pub const APP_TOKEN_VAR: &str = "MOAI_AUDIT_APP_TOKEN";
/// GitHub truncates longer status descriptions.
const DESCRIPTION_MAX: usize = 140;
const STATES: [&str; 4] = ["success", "failure", "error", "pending"];

#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct AuditStatusRequest {
    pub schema_version: u32,
    pub owner: String,
    pub repo: String,
    pub sha: String,
    pub state: String,
    /// The verdict comment (`https://github.com/...`).
    pub target_url: String,
    #[serde(default)]
    pub description: String,
}

fn valid_name(s: &str) -> bool {
    !s.is_empty()
        && s.len() <= 100
        && s.bytes()
            .all(|b| b.is_ascii_alphanumeric() || b"-_.".contains(&b))
        && s != "."
        && s != ".."
}

fn validate(req: &AuditStatusRequest, permission: &str) -> Result<(), ForgeError> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err(ForgeError::of(
            "permission_reference_missing",
            "permission reference required",
        ));
    }
    if req.schema_version != SCHEMA_VERSION {
        return Err(ForgeError::of(
            "unsupported_protocol_version",
            "unsupported request schema_version",
        ));
    }
    if !valid_name(&req.owner) || !valid_name(&req.repo) {
        return Err(ForgeError::invalid("owner and repo must be GitHub names"));
    }
    let hex = req.sha.len() == 40
        && req
            .sha
            .bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b));
    if !hex {
        return Err(ForgeError::invalid("sha must be a 40 lowercase hex commit"));
    }
    if !STATES.contains(&req.state.as_str()) {
        return Err(ForgeError::invalid(
            "state must be success, failure, error or pending",
        ));
    }
    if !req.target_url.starts_with("https://github.com/")
        || req.target_url.len() > 512
        || req.target_url.chars().any(|c| c.is_control() || c == ' ')
    {
        return Err(ForgeError::invalid(
            "target_url must be an https://github.com/ link",
        ));
    }
    if req.description.chars().any(char::is_control) {
        return Err(ForgeError::invalid(
            "description must be one line without control characters",
        ));
    }
    Ok(())
}

/// The description sent: the caller's text plus the source marker, within GitHub's limit.
fn description(text: &str, source: &str) -> Result<String, ForgeError> {
    let marker = format!("source={source}");
    let full = if text.is_empty() {
        marker
    } else {
        format!("{text} · {marker}")
    };
    if full.chars().count() > DESCRIPTION_MAX {
        return Err(ForgeError::invalid(format!(
            "description must be at most {DESCRIPTION_MAX} characters with the source marker"
        )));
    }
    Ok(full)
}

/// Posts and reads back. `app` says whether the transport carries the App token.
pub fn post<T: Transport, C: Clock>(
    req: &AuditStatusRequest,
    permission: &str,
    app: bool,
    transport: &T,
    clock: &C,
) -> Result<Value, ForgeError> {
    validate(req, permission)?;
    let (source, token_source) = if app {
        ("app", "app")
    } else {
        ("shared-account", "shared_account")
    };
    let desc = description(&req.description, source)?;
    let client = Client::new(transport, clock, Config::default(), BTreeMap::new());
    let (o, r, sha) = (&req.owner, &req.repo, &req.sha);
    let post =
        HttpRequest::new("POST", format!("{API}/repos/{o}/{r}/statuses/{sha}")).with_body(json!({
            "state": req.state, "target_url": req.target_url,
            "description": desc, "context": CONTEXT,
        }));
    let resp = client.send_with_backoff(&post)?;
    if resp.status != 201 {
        return Err(ForgeError::transport(format!(
            "status POST returned HTTP {}",
            resp.status
        )));
    }
    // Read back: the newest `moai/audit` entry on that sha, or uncertain.
    let observed = match client.get(
        &format!("{API}/repos/{o}/{r}/commits/{sha}/statuses?per_page=100"),
        None,
    ) {
        Ok(back) => back
            .body
            .as_ref()
            .and_then(Value::as_array)
            .and_then(|a| a.iter().find(|s| s["context"] == CONTEXT))
            .map(|s| json!({"state": s["state"], "description": s["description"]}))
            .unwrap_or_else(|| json!("uncertain")),
        Err(_) => json!("uncertain"),
    };
    let mut warnings = Vec::new();
    if !app {
        warnings.push(format!(
            "status source is not pinned: posted by the shared account ({APP_TOKEN_VAR} not set)"
        ));
    }
    Ok(json!({
        "schema_version": SCHEMA_VERSION,
        "sha": sha,
        "context": CONTEXT,
        "state": req.state,
        "description": desc,
        "token_source": token_source,
        "token_variable": APP_TOKEN_VAR,
        "observed": observed,
        "warnings": warnings,
        "notes": [
            "pinning the source is the ruleset's expected-source app for the required check; operator keep-set, not done here"
        ],
    }))
}

/// CLI entry (`forge.audit_status`): reads the App token from the environment by name.
pub fn audit_status_cli<C: Clock>(
    project: &Path,
    req: AuditStatusRequest,
    permission: &str,
    mut transport: GhTransport,
    clock: &C,
) -> Result<Value, String> {
    if !project.is_dir() {
        return Err(ForgeError::invalid("project must be a directory").into());
    }
    let token = std::env::var(APP_TOKEN_VAR)
        .ok()
        .filter(|t| !t.trim().is_empty());
    let app = token.is_some();
    transport.gh_token = token;
    Ok(post(&req, permission, app, &transport, clock)?)
}
