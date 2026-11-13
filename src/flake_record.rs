//! `flake-record` (Issue #2666 F2): one tracking Issue per flaky test. For every flaky row of a
//! `flake-classify` result, find the open Issue whose body starts with the test's hidden marker
//! `<!-- moai:flake test=<test> -->`; add a comment for this run and job, or create the Issue
//! (title `[flake] <test>`) when none is open. A run and job already recorded — its hidden run
//! marker is a line of the Issue body or of one of its comments — writes nothing, so a second
//! run of the same command is a no-op. Writes need a permission reference; a transport error on
//! a write is read back before anything is claimed.

use crate::forge::policy::next_link;
use crate::forge::transport::{HttpRequest, Transport};
use serde_json::{json, Value};

type Result<T> = std::result::Result<T, String>;

/// Pages read at most per listing (100 items each).
const MAX_PAGES: usize = 20;

/// The hidden marker that names the test an Issue tracks: the first line of its body.
pub fn test_marker(test: &str) -> String {
    format!("<!-- moai:flake test={test} -->")
}

/// The hidden marker of one recorded failure: a line of the Issue body or of a comment.
pub fn run_marker(run: u64, job: &str) -> String {
    format!("<!-- moai:flake-run run={run} job={job} -->")
}

/// One flaky test of one job, as `flake-classify` lists it.
#[derive(Clone, Debug, PartialEq)]
pub struct Flaky {
    pub job: String,
    pub test: String,
    pub message: Option<String>,
    /// The head the job ran on, from the result's `first_heads` (flaky means both runs had it).
    pub head: Option<String>,
}

/// The `flaky` rows of a `flake-classify` result (bare, or inside its `{"result": …}` envelope).
pub fn flaky_rows(classified: &Value) -> Result<Vec<Flaky>> {
    let c = classified.get("result").unwrap_or(classified);
    let rows = c["flaky"]
        .as_array()
        .ok_or("not a flake-classify result: no `flaky` list")?;
    rows.iter()
        .map(|r| {
            Ok(Flaky {
                job: r["job"]
                    .as_str()
                    .ok_or("a flaky row has no job")?
                    .to_string(),
                test: r["test"]
                    .as_str()
                    .filter(|t| !t.trim().is_empty() && !t.contains('\n') && !t.contains("-->"))
                    .ok_or("a flaky row has no usable test name")?
                    .to_string(),
                message: r["message"].as_str().map(String::from),
                head: r["job"]
                    .as_str()
                    .and_then(|j| c["first_heads"][j].as_str())
                    .map(String::from),
            })
        })
        .collect()
}

/// Every item of a listing, following `Link: rel="next"` exactly.
fn list<T: Transport>(transport: &T, first: String) -> Result<Vec<Value>> {
    let mut items = Vec::new();
    let mut url = Some(first);
    let mut pages = 0;
    while let Some(u) = url {
        pages += 1;
        if pages > MAX_PAGES {
            return Err(format!("more than {MAX_PAGES} pages; refusing to guess"));
        }
        let resp = transport.send(&HttpRequest::new("GET", &u))?;
        if resp.status != 200 {
            return Err(format!("GET {u} returned HTTP {}", resp.status));
        }
        match resp.body.as_ref().and_then(Value::as_array) {
            Some(a) => items.extend(a.iter().cloned()),
            None => return Err(format!("GET {u} did not return a list")),
        }
        url = next_link(resp.header("link"));
    }
    Ok(items)
}

fn lines_hold(text: Option<&str>, marker: &str) -> bool {
    text.is_some_and(|t| t.lines().any(|l| l.trim() == marker))
}

fn first_line_is(text: Option<&str>, marker: &str) -> bool {
    text.and_then(|t| t.lines().next())
        .is_some_and(|l| l.trim() == marker)
}

fn entry(f: &Flaky, run: u64, run_url: &str) -> String {
    let message = f.message.as_deref().unwrap_or("(no panic line in the log)");
    format!(
        "{}\n- run: {run_url} (run {run})\n- job: `{}`\n- head: `{}`\n- message: `{}`",
        run_marker(run, &f.job),
        f.job,
        f.head.as_deref().unwrap_or("unknown"),
        message.replace('`', "'")
    )
}

/// Records every flaky row of `classified` for the failed run `run` on `owner/repo`.
pub fn record<T: Transport>(
    owner: &str,
    repo: &str,
    run: u64,
    classified: &Value,
    permission: &str,
    transport: &T,
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("a permission reference is required; nothing was read or written".into());
    }
    let rows = flaky_rows(classified)?;
    let api = format!("https://api.github.com/repos/{owner}/{repo}");
    let run_url = format!("https://github.com/{owner}/{repo}/actions/runs/{run}");
    let issues_url = format!("{api}/issues?state=open&per_page=100");
    let mut open: Vec<Value> = list(transport, issues_url.clone())?
        .into_iter()
        .filter(|i| i.get("pull_request").is_none())
        .collect();
    let (mut created, mut commented, mut skipped, mut uncertain) =
        (Vec::new(), Vec::new(), Vec::new(), Vec::new());
    for f in &rows {
        let marker = test_marker(&f.test);
        let run_mark = run_marker(run, &f.job);
        let found: Vec<&Value> = open
            .iter()
            .filter(|i| first_line_is(i["body"].as_str(), &marker))
            .collect();
        if found.len() > 1 {
            let numbers: Vec<Value> = found.iter().map(|i| i["number"].clone()).collect();
            return Err(format!(
                "more than one open Issue tracks {}: {numbers:?}; nothing more written",
                f.test
            ));
        }
        if let Some(issue) = found.first() {
            let number = issue["number"]
                .as_u64()
                .ok_or("an Issue without a number")?;
            let recorded = lines_hold(issue["body"].as_str(), &run_mark)
                || list(
                    transport,
                    format!("{api}/issues/{number}/comments?per_page=100"),
                )?
                .iter()
                .any(|c| lines_hold(c["body"].as_str(), &run_mark));
            let row = json!({"test": f.test, "job": f.job, "issue": number});
            if recorded {
                skipped.push(row);
                continue;
            }
            let body = format!("{}\n\n🗿 MoAI", entry(f, run, &run_url));
            let sent = transport.send(
                &HttpRequest::new("POST", format!("{api}/issues/{number}/comments"))
                    .with_body(json!({"body": body})),
            );
            match sent {
                Ok(r) if r.status == 201 => commented.push(row),
                Ok(r) => return Err(format!("comment on #{number} returned HTTP {}", r.status)),
                // Maybe posted: read the comments back.
                Err(e) => {
                    let back = list(
                        transport,
                        format!("{api}/issues/{number}/comments?per_page=100"),
                    )
                    .map(|cs| cs.iter().any(|c| lines_hold(c["body"].as_str(), &run_mark)));
                    match back {
                        Ok(true) => commented.push(row),
                        Ok(false) => return Err(format!("comment on #{number} failed: {e}")),
                        Err(_) => uncertain.push(json!({"test": f.test, "job": f.job,
                            "issue": number, "error": e})),
                    }
                }
            }
            continue;
        }
        let title = format!("[flake] {}", f.test);
        let body = format!(
            "{marker}\n불안정 시험 추적 (`moai-cli flake-record`, #2666). 같은 head의 뒤 run에서 통과한 실패를 run마다 적어요.\n\n{}\n\n🗿 MoAI",
            entry(f, run, &run_url)
        );
        let sent = transport.send(
            &HttpRequest::new("POST", format!("{api}/issues"))
                .with_body(json!({"title": title, "body": body})),
        );
        let issue = match sent {
            Ok(r) if r.status == 201 => r.body.unwrap_or(Value::Null),
            Ok(r) => return Err(format!("creating {title:?} returned HTTP {}", r.status)),
            // Maybe created: list the open Issues again before claiming anything.
            Err(e) => match list(transport, issues_url.clone()) {
                Ok(again) => match again
                    .into_iter()
                    .find(|i| first_line_is(i["body"].as_str(), &marker))
                {
                    Some(i) => i,
                    None => return Err(format!("creating {title:?} failed: {e}")),
                },
                Err(_) => {
                    uncertain.push(json!({"test": f.test, "job": f.job, "error": e}));
                    continue;
                }
            },
        };
        created.push(json!({"test": f.test, "job": f.job, "issue": issue["number"]}));
        // A later row of the same test (another job) comments on this Issue.
        open.push(issue);
    }
    let writes = created.len() + commented.len();
    Ok(json!({
        "run": run, "created": created, "commented": commented,
        "already_recorded": skipped, "uncertain": uncertain, "writes": writes,
    }))
}
