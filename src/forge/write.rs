//! Edit-preserving write helpers (REQ-FORGE-007/008). Not exposed as CLI commands here; 07-B
//! and 07-C call them. Label and assignee sets are only ever changed through the dedicated
//! add (POST) and remove (DELETE) endpoints, never by sending a whole set with PATCH or PUT,
//! and every write is preceded by a re-read that must match what the caller last read.
//! GitHub does not support conditional requests on unsafe methods, so the guard is
//! read-before-write rather than If-Match.
use super::{
    policy::Clock,
    transport::{HttpRequest, Transport},
    Client, IssueKey,
};
use serde_json::{json, Value};
use std::cell::Cell;

type Result<T> = std::result::Result<T, String>;

/// What the caller saw on its last read; a write proceeds only if the issue still matches.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct LastRead {
    pub updated_at: String,
    /// `crate::digest` of the issue body text as last read (`""` for an empty body).
    pub body_digest: String,
}

#[derive(Debug, PartialEq, Eq)]
pub enum WriteOutcome {
    Applied,
    /// The issue changed since the caller's read; nothing was sent.
    Conflict {
        updated_at: String,
    },
}

/// Serialises writes and keeps at least one second between them (GitHub best practices).
pub struct Writer<'a, T: Transport, C: Clock> {
    pub client: &'a Client<'a, T, C>,
    last_write: Cell<Option<i64>>,
}

impl<'a, T: Transport, C: Clock> Writer<'a, T, C> {
    pub fn new(client: &'a Client<'a, T, C>) -> Self {
        Writer {
            client,
            last_write: Cell::new(None),
        }
    }

    fn issue_url(&self, key: &IssueKey, tail: &str) -> String {
        format!(
            "https://api.github.com/repos/{}/{}/issues/{}{}",
            key.owner, key.repo, key.number, tail
        )
    }

    fn check(&self, key: &IssueKey, last: &LastRead) -> Result<Option<String>> {
        let resp = self.client.get(&self.issue_url(key, ""), None)?;
        let body = resp.body.unwrap_or(Value::Null);
        if body["id"].as_i64() != Some(key.id) {
            return Err("issue id changed under the same number; refusing to write".into());
        }
        let updated_at = body["updated_at"].as_str().unwrap_or_default().to_string();
        let digest = crate::digest(body["body"].as_str().unwrap_or("").as_bytes());
        Ok((updated_at != last.updated_at || digest != last.body_digest).then_some(updated_at))
    }

    fn write(&self, method: &str, url: String, body: Option<Value>) -> Result<()> {
        if let Some(prev) = self.last_write.get() {
            let gap = self.client.clock.now() - prev;
            if gap < 1 {
                self.client.clock.sleep((1 - gap) as u64);
            }
        }
        let mut req = HttpRequest::new(method, url);
        req.body = body;
        let resp = self.client.send_with_backoff(&req)?;
        self.last_write.set(Some(self.client.clock.now()));
        if (200..300).contains(&resp.status) {
            Ok(())
        } else {
            Err(format!("{method} returned HTTP {}", resp.status))
        }
    }

    fn guarded(
        &self,
        key: &IssueKey,
        last: &LastRead,
        method: &str,
        tail: String,
        body: Option<Value>,
    ) -> Result<WriteOutcome> {
        if let Some(updated_at) = self.check(key, last)? {
            return Ok(WriteOutcome::Conflict { updated_at });
        }
        self.write(method, self.issue_url(key, &tail), body)?;
        Ok(WriteOutcome::Applied)
    }

    /// POST /issues/{n}/labels: adds labels and keeps the existing ones.
    pub fn add_labels(
        &self,
        key: &IssueKey,
        last: &LastRead,
        labels: &[&str],
    ) -> Result<WriteOutcome> {
        self.guarded(
            key,
            last,
            "POST",
            "/labels".into(),
            Some(json!({ "labels": labels })),
        )
    }

    /// DELETE /issues/{n}/labels/{name}: removes exactly one label.
    pub fn remove_label(
        &self,
        key: &IssueKey,
        last: &LastRead,
        label: &str,
    ) -> Result<WriteOutcome> {
        let name = super::labels::encode_segment(label);
        self.guarded(key, last, "DELETE", format!("/labels/{name}"), None)
    }

    /// POST /issues/{n}/assignees: adds up to 10 assignees and keeps the existing ones.
    pub fn add_assignees(
        &self,
        key: &IssueKey,
        last: &LastRead,
        logins: &[&str],
    ) -> Result<WriteOutcome> {
        if logins.len() > 10 {
            return Err("at most 10 assignees per request".into());
        }
        self.guarded(
            key,
            last,
            "POST",
            "/assignees".into(),
            Some(json!({ "assignees": logins })),
        )
    }

    /// DELETE /issues/{n}/assignees: removes only the named assignees.
    pub fn remove_assignees(
        &self,
        key: &IssueKey,
        last: &LastRead,
        logins: &[&str],
    ) -> Result<WriteOutcome> {
        self.guarded(
            key,
            last,
            "DELETE",
            "/assignees".into(),
            Some(json!({ "assignees": logins })),
        )
    }
}
