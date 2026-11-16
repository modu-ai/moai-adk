//! The host connector (#3071 H1): the four things card code asks of a code host.
//!
//! Card code never talks to a host directly. It holds a [`HostConnector`] and calls one of four
//! operations: read the open issues, put a summary on an issue or PR, read a PR's state, read the
//! check and audit verdicts of a PR. One implementation exists per host: [`GithubConnector`] on
//! the existing [`Transport`] and [`Client`] (tests replay recorded exchanges and never touch the
//! network), and [`FakeConnector`] for the tests of the card commands. GitLab has no
//! implementation yet: [`HostKind::require_implemented`] refuses it before any connector or
//! transport exists, so no request can be made and no token is read.
//!
//! There is no operation that deletes anything on the host. Closing a moved issue is a later
//! slice (H5) and adds its own narrow operation.
//!
//! Issue bodies and comments are returned as text with their author's association. Nothing here
//! reads a control line: whether a comment may carry one is decided by `comment_authority` where
//! the comment becomes a card note (a later slice).
use super::policy::{next_link, Clock};
use super::transport::{redact, HttpRequest, HttpResponse, Transport};
use super::{audit_status, valid_name, Client, Config, ForgeError, API};
use crate::comment_authority;
use serde_json::{json, Value};
use std::{cell::RefCell, collections::BTreeMap, fmt};

/// What a refusal of GitLab says is missing.
const GITLAB_NEEDS: &str =
    "a test project (gitlab.com or a self-hosted server) and an access token kept in the CLI's vault";

/// The code host a card comes from or goes to.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum HostKind {
    Github,
    Gitlab,
}

impl HostKind {
    /// The value stored in `card.source_host`.
    pub fn as_str(self) -> &'static str {
        match self {
            HostKind::Github => "github",
            HostKind::Gitlab => "gitlab",
        }
    }

    /// Reads the host a command line names (`--from`, `--to`).
    pub fn parse(text: &str) -> Result<HostKind, ConnectorError> {
        match text {
            "github" => Ok(HostKind::Github),
            "gitlab" => Ok(HostKind::Gitlab),
            other => Err(ConnectorError::Invalid(format!(
                "unknown host `{other}`; use github or gitlab"
            ))),
        }
    }

    /// Refuses a host that has no transport yet. Called before a connector is built, so a
    /// refused host makes no request and reads no token.
    pub fn require_implemented(self) -> Result<HostKind, ConnectorError> {
        match self {
            HostKind::Github => Ok(self),
            HostKind::Gitlab => Err(ConnectorError::NotImplemented {
                host: "gitlab",
                needs: GITLAB_NEEDS,
            }),
        }
    }
}

/// Why a connector call did not give an answer.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum ConnectorError {
    /// The host has no connector yet; `needs` says what a connector needs first.
    NotImplemented {
        host: &'static str,
        needs: &'static str,
    },
    /// A listing could not be read to its end (page limit). A partial list is never returned,
    /// because a count taken from it would look right and be wrong.
    Incomplete { reason: String },
    /// The input is not something a host can be asked about.
    Invalid(String),
    /// The host or the transport failed.
    Host(ForgeError),
}

impl fmt::Display for ConnectorError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ConnectorError::NotImplemented { host, needs } => {
                write!(f, "{host} is not implemented yet; it needs {needs}")
            }
            ConnectorError::Incomplete { reason } => {
                write!(f, "the host listing is incomplete: {reason}")
            }
            ConnectorError::Invalid(message) => f.write_str(message),
            ConnectorError::Host(e) => f.write_str(&e.message),
        }
    }
}

impl From<ForgeError> for ConnectorError {
    fn from(e: ForgeError) -> Self {
        ConnectorError::Host(e)
    }
}

/// One project on one host. The only place that normalizes the name stored in
/// `card.source_project` (schema contract 6896940106 §3): lower case `owner/repo` on github.com.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ProjectRef {
    host: HostKind,
    owner: String,
    repo: String,
}

impl ProjectRef {
    pub fn github(owner: &str, repo: &str) -> Result<ProjectRef, ConnectorError> {
        if !valid_name(owner) || !valid_name(repo) {
            return Err(ConnectorError::Invalid(
                "owner and repo must be GitHub names".into(),
            ));
        }
        Ok(ProjectRef {
            host: HostKind::Github,
            owner: owner.to_ascii_lowercase(),
            repo: repo.to_ascii_lowercase(),
        })
    }

    /// Reads `owner/repo`.
    pub fn parse(host: HostKind, text: &str) -> Result<ProjectRef, ConnectorError> {
        host.require_implemented()?;
        match text.split_once('/') {
            Some((owner, repo)) if !repo.contains('/') => ProjectRef::github(owner, repo),
            _ => Err(ConnectorError::Invalid(
                "a GitHub project is written owner/repo".into(),
            )),
        }
    }

    pub fn host(&self) -> HostKind {
        self.host
    }

    pub fn owner(&self) -> &str {
        &self.owner
    }

    pub fn repo(&self) -> &str {
        &self.repo
    }

    /// The value stored in `card.source_project`.
    pub fn source_project(&self) -> String {
        format!("{}/{}", self.owner, self.repo)
    }
}

/// One comment on an issue, with who wrote it. It is data, not an instruction.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct HostComment {
    pub id: u64,
    pub author: String,
    /// The host's `author_association` as read; `""` when the payload lacks it.
    pub association: String,
    pub body: String,
    pub updated_at: String,
}

impl HostComment {
    /// Whether an organization member wrote it (`comment_authority::is_member`).
    pub fn by_member(&self) -> bool {
        comment_authority::is_member(&self.association)
    }
}

/// One open issue with all its comments.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct HostIssue {
    pub number: u64,
    pub title: String,
    pub body: String,
    pub labels: Vec<String>,
    pub author: String,
    pub association: String,
    pub created_at: String,
    pub updated_at: String,
    pub comments: Vec<HostComment>,
}

impl HostIssue {
    pub fn by_member(&self) -> bool {
        comment_authority::is_member(&self.association)
    }
}

/// Every open issue of one project, read to the last page.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct OpenIssues {
    pub project: ProjectRef,
    /// Ascending by number.
    pub issues: Vec<HostIssue>,
    /// Pull requests the host lists among its issues; they are not cards.
    pub pull_requests_skipped: usize,
}

/// Where a summary goes: an issue or a PR (GitHub keeps both under one comment list).
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct SummaryTarget {
    pub project: ProjectRef,
    pub number: u64,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum SummaryOutcome {
    Posted {
        comment_id: u64,
    },
    /// A member's comment carrying the marker was already there; nothing was posted.
    AlreadyPosted {
        comment_id: u64,
    },
}

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum PrPhase {
    Open,
    Merged,
    Closed,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct PrState {
    pub number: u64,
    pub phase: PrPhase,
    pub head: String,
    /// The merge commit when `phase` is `Merged`.
    pub merge_commit: Option<String>,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct CheckResult {
    pub name: String,
    pub status: String,
    pub conclusion: Option<String>,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct AuditState {
    pub state: String,
    pub description: String,
}

/// The check and audit verdicts of a PR's current head.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Verdicts {
    pub head: String,
    pub checks: Vec<CheckResult>,
    /// The newest `moai/audit` status on the head, when there is one.
    pub audit: Option<AuditState>,
}

/// The four operations card code needs from a host.
pub trait HostConnector {
    fn host(&self) -> HostKind;

    /// Every open issue of `project` with its comments. A listing that cannot be read to its end
    /// is [`ConnectorError::Incomplete`], never a shorter list.
    fn open_issues(&self, project: &ProjectRef) -> Result<OpenIssues, ConnectorError>;

    /// Puts `body` on the target once. `marker` is one `<!-- moai-… -->` line: when a member's
    /// comment on the target already carries it, nothing is posted. Otherwise the marker is
    /// appended to the body, so the next call finds it.
    fn post_summary(
        &self,
        target: &SummaryTarget,
        body: &str,
        marker: &str,
    ) -> Result<SummaryOutcome, ConnectorError>;

    fn pr_state(&self, project: &ProjectRef, number: u64) -> Result<PrState, ConnectorError>;

    fn verdicts(&self, project: &ProjectRef, number: u64) -> Result<Verdicts, ConnectorError>;
}

/// The longest marker, in bytes.
pub const MARKER_MAX_BYTES: usize = 200;

/// A marker is one HTML comment line: it starts `<!-- moai-`, ends ` -->`, has no `-->` inside
/// (that would close the comment early), no control character (so no second line), and is at
/// most [`MARKER_MAX_BYTES`] bytes.
pub fn check_marker(marker: &str) -> Result<(), ConnectorError> {
    let ok = marker.starts_with("<!-- moai-")
        && marker
            .strip_suffix(" -->")
            .is_some_and(|inner| !inner.contains("-->"))
        && marker.len() <= MARKER_MAX_BYTES
        && !marker.chars().any(char::is_control);
    if ok {
        Ok(())
    } else {
        Err(ConnectorError::Invalid(
            "a marker is one line, `<!-- moai-… -->`, without control characters".into(),
        ))
    }
}

fn text(v: &Value) -> String {
    v.as_str().unwrap_or("").to_string()
}

/// The GitHub implementation, on a transport and a clock handed in.
pub struct GithubConnector<'a, T: Transport, C: Clock> {
    client: Client<'a, T, C>,
}

impl<'a, T: Transport, C: Clock> GithubConnector<'a, T, C> {
    pub fn new(transport: &'a T, clock: &'a C, config: Config) -> Self {
        GithubConnector {
            client: Client::new(transport, clock, config, BTreeMap::new()),
        }
    }

    fn require_github(project: &ProjectRef) -> Result<(), ConnectorError> {
        project.host().require_implemented().map(|_| ())
    }

    /// All pages of a list. `pick` turns one response into its items.
    fn read_all(
        &self,
        first: String,
        pick: &dyn Fn(&HttpResponse) -> Vec<Value>,
    ) -> Result<Vec<Value>, ConnectorError> {
        let project =
            |r: &HttpResponse| json!({"items": pick(r), "next": next_link(r.header("link"))});
        match self.client.pages(first, &project) {
            Ok((items, _)) => Ok(items),
            Err(e) if e.message.starts_with("page limit exceeded") => {
                Err(ConnectorError::Incomplete { reason: e.message })
            }
            Err(e) => Err(e.into()),
        }
    }

    fn array(r: &HttpResponse) -> Vec<Value> {
        r.body
            .as_ref()
            .and_then(Value::as_array)
            .cloned()
            .unwrap_or_default()
    }

    fn comments_url(project: &ProjectRef, number: u64) -> String {
        format!(
            "{API}/repos/{}/{}/issues/{number}/comments?per_page=100",
            project.owner(),
            project.repo()
        )
    }

    fn comment_of(c: &Value) -> Option<HostComment> {
        Some(HostComment {
            id: c["id"].as_u64()?,
            author: text(&c["user"]["login"]),
            association: comment_authority::rest_association(c),
            body: text(&c["body"]),
            updated_at: text(&c["updated_at"]),
        })
    }

    fn pull(&self, project: &ProjectRef, number: u64) -> Result<Value, ConnectorError> {
        let url = format!(
            "{API}/repos/{}/{}/pulls/{number}",
            project.owner(),
            project.repo()
        );
        let resp = self.client.get(&url, None)?;
        resp.body
            .ok_or_else(|| ForgeError::transport("pull request response had no body").into())
    }
}

fn pr_state_of(number: u64, pull: &Value) -> Result<PrState, ConnectorError> {
    let head = pull["head"]["sha"]
        .as_str()
        .ok_or_else(|| ForgeError::transport("pull request response carried no head sha"))?
        .to_string();
    let merged = pull["merged"].as_bool() == Some(true) || pull["merged_at"].is_string();
    let phase = match (pull["state"].as_str(), merged) {
        (Some("open"), _) => PrPhase::Open,
        (Some("closed"), true) => PrPhase::Merged,
        (Some("closed"), false) => PrPhase::Closed,
        _ => {
            return Err(ForgeError::transport("pull request response had no state").into());
        }
    };
    let merge_commit = (phase == PrPhase::Merged)
        .then(|| pull["merge_commit_sha"].as_str().map(str::to_string))
        .flatten();
    Ok(PrState {
        number,
        phase,
        head,
        merge_commit,
    })
}

impl<T: Transport, C: Clock> HostConnector for GithubConnector<'_, T, C> {
    fn host(&self) -> HostKind {
        HostKind::Github
    }

    fn open_issues(&self, project: &ProjectRef) -> Result<OpenIssues, ConnectorError> {
        Self::require_github(project)?;
        let (o, r) = (project.owner(), project.repo());
        let raw = self.read_all(
            format!(
                "{API}/repos/{o}/{r}/issues?state=open&sort=created&direction=asc&per_page=100"
            ),
            &Self::array,
        )?;
        let mut issues = Vec::new();
        let mut pull_requests_skipped = 0;
        for item in &raw {
            if item.get("pull_request").is_some() {
                pull_requests_skipped += 1;
                continue;
            }
            let Some(number) = item["number"].as_u64() else {
                continue;
            };
            let comments = self
                .read_all(Self::comments_url(project, number), &Self::array)?
                .iter()
                .filter_map(Self::comment_of)
                .collect();
            issues.push(HostIssue {
                number,
                title: text(&item["title"]),
                body: text(&item["body"]),
                labels: item["labels"]
                    .as_array()
                    .map(|a| a.iter().map(|l| text(&l["name"])).collect())
                    .unwrap_or_default(),
                author: text(&item["user"]["login"]),
                association: comment_authority::rest_association(item),
                created_at: text(&item["created_at"]),
                updated_at: text(&item["updated_at"]),
                comments,
            });
        }
        issues.sort_by_key(|i| i.number);
        Ok(OpenIssues {
            project: project.clone(),
            issues,
            pull_requests_skipped,
        })
    }

    fn post_summary(
        &self,
        target: &SummaryTarget,
        body: &str,
        marker: &str,
    ) -> Result<SummaryOutcome, ConnectorError> {
        Self::require_github(&target.project)?;
        check_marker(marker)?;
        // The marker counts only on a member's comment: on a public repository anyone could
        // write the marker text first and so stop a summary from ever being posted.
        let existing = self
            .read_all(
                Self::comments_url(&target.project, target.number),
                &Self::array,
            )?
            .iter()
            .filter_map(Self::comment_of)
            .find(|c| c.by_member() && c.body.contains(marker));
        if let Some(c) = existing {
            return Ok(SummaryOutcome::AlreadyPosted { comment_id: c.id });
        }
        let url = Self::comments_url(&target.project, target.number);
        let url = url.split('?').next().unwrap_or(&url).to_string();
        let request = HttpRequest::new("POST", url)
            .with_body(json!({"body": format!("{}\n\n{marker}", redact(body.trim_end()))}));
        let resp = self.client.send_with_backoff(&request)?;
        if resp.status != 201 {
            return Err(ForgeError::transport(format!(
                "summary POST returned HTTP {}",
                resp.status
            ))
            .into());
        }
        let comment_id = resp
            .body
            .as_ref()
            .and_then(|b| b["id"].as_u64())
            .ok_or_else(|| ForgeError::transport("summary POST answered without a comment id"))?;
        Ok(SummaryOutcome::Posted { comment_id })
    }

    fn pr_state(&self, project: &ProjectRef, number: u64) -> Result<PrState, ConnectorError> {
        Self::require_github(project)?;
        pr_state_of(number, &self.pull(project, number)?)
    }

    fn verdicts(&self, project: &ProjectRef, number: u64) -> Result<Verdicts, ConnectorError> {
        Self::require_github(project)?;
        let head = pr_state_of(number, &self.pull(project, number)?)?.head;
        let (o, r) = (project.owner(), project.repo());
        let runs = self.read_all(
            format!("{API}/repos/{o}/{r}/commits/{head}/check-runs?per_page=100"),
            &|resp| {
                resp.body
                    .as_ref()
                    .and_then(|b| b["check_runs"].as_array())
                    .cloned()
                    .unwrap_or_default()
            },
        )?;
        let checks = runs
            .iter()
            .map(|run| CheckResult {
                name: text(&run["name"]),
                status: text(&run["status"]),
                conclusion: run["conclusion"].as_str().map(str::to_string),
            })
            .collect();
        // Newest first; the first `moai/audit` entry is the current one.
        let statuses = self.read_all(
            format!("{API}/repos/{o}/{r}/commits/{head}/statuses?per_page=100"),
            &Self::array,
        )?;
        let audit = statuses
            .iter()
            .find(|s| s["context"] == audit_status::CONTEXT)
            .map(|s| AuditState {
                state: text(&s["state"]),
                description: text(&s["description"]),
            });
        Ok(Verdicts {
            head,
            checks,
            audit,
        })
    }
}

/// One call a [`FakeConnector`] received. There is no variant that deletes: a test that asserts
/// on the call list shows by construction that nothing was removed.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Call {
    OpenIssues {
        project: String,
    },
    PostSummary {
        project: String,
        number: u64,
        marker: String,
    },
    PrState {
        project: String,
        number: u64,
    },
    Verdicts {
        project: String,
        number: u64,
    },
}

/// A connector for the tests of the card commands: answers from what it was given and records
/// every call. It posts nothing anywhere; a summary is "posted" into its own list.
pub struct FakeConnector {
    host: HostKind,
    issues: Vec<HostIssue>,
    incomplete: Option<String>,
    pull_requests_skipped: usize,
    prs: BTreeMap<u64, PrState>,
    verdicts: BTreeMap<u64, Verdicts>,
    posted: RefCell<Vec<(String, u64, String, String)>>,
    calls: RefCell<Vec<Call>>,
}

impl FakeConnector {
    pub fn new(host: HostKind) -> Self {
        FakeConnector {
            host,
            issues: Vec::new(),
            incomplete: None,
            pull_requests_skipped: 0,
            prs: BTreeMap::new(),
            verdicts: BTreeMap::new(),
            posted: RefCell::new(Vec::new()),
            calls: RefCell::new(Vec::new()),
        }
    }

    pub fn with_issues(mut self, issues: Vec<HostIssue>) -> Self {
        self.issues = issues;
        self
    }

    /// Makes `open_issues` fail as an unreadable-to-the-end listing.
    pub fn with_incomplete(mut self, reason: &str) -> Self {
        self.incomplete = Some(reason.to_string());
        self
    }

    /// How many pull requests `open_issues` says it left out of the listing.
    pub fn with_pull_requests_skipped(mut self, count: usize) -> Self {
        self.pull_requests_skipped = count;
        self
    }

    pub fn with_pr(mut self, state: PrState) -> Self {
        self.prs.insert(state.number, state);
        self
    }

    pub fn with_verdicts(mut self, number: u64, verdicts: Verdicts) -> Self {
        self.verdicts.insert(number, verdicts);
        self
    }

    pub fn calls(&self) -> Vec<Call> {
        self.calls.borrow().clone()
    }

    /// `(project, number, marker, body)` of every summary that was posted, in order.
    pub fn posted(&self) -> Vec<(String, u64, String, String)> {
        self.posted.borrow().clone()
    }

    fn note(&self, call: Call) {
        self.calls.borrow_mut().push(call);
    }
}

impl HostConnector for FakeConnector {
    fn host(&self) -> HostKind {
        self.host
    }

    fn open_issues(&self, project: &ProjectRef) -> Result<OpenIssues, ConnectorError> {
        self.note(Call::OpenIssues {
            project: project.source_project(),
        });
        if let Some(reason) = &self.incomplete {
            return Err(ConnectorError::Incomplete {
                reason: reason.clone(),
            });
        }
        Ok(OpenIssues {
            project: project.clone(),
            issues: self.issues.clone(),
            pull_requests_skipped: self.pull_requests_skipped,
        })
    }

    fn post_summary(
        &self,
        target: &SummaryTarget,
        body: &str,
        marker: &str,
    ) -> Result<SummaryOutcome, ConnectorError> {
        check_marker(marker)?;
        let project = target.project.source_project();
        self.note(Call::PostSummary {
            project: project.clone(),
            number: target.number,
            marker: marker.to_string(),
        });
        let mut posted = self.posted.borrow_mut();
        if let Some(at) = posted
            .iter()
            .position(|(p, n, m, _)| *p == project && *n == target.number && m == marker)
        {
            return Ok(SummaryOutcome::AlreadyPosted {
                comment_id: 1000 + at as u64,
            });
        }
        posted.push((project, target.number, marker.to_string(), body.to_string()));
        Ok(SummaryOutcome::Posted {
            comment_id: 999 + posted.len() as u64,
        })
    }

    fn pr_state(&self, project: &ProjectRef, number: u64) -> Result<PrState, ConnectorError> {
        self.note(Call::PrState {
            project: project.source_project(),
            number,
        });
        self.prs
            .get(&number)
            .cloned()
            .ok_or_else(|| ConnectorError::Invalid(format!("no pull request {number} known")))
    }

    fn verdicts(&self, project: &ProjectRef, number: u64) -> Result<Verdicts, ConnectorError> {
        self.note(Call::Verdicts {
            project: project.source_project(),
            number,
        });
        self.verdicts
            .get(&number)
            .cloned()
            .ok_or_else(|| ConnectorError::Invalid(format!("no verdicts for {number} known")))
    }
}
