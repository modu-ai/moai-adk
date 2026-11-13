//! Flow control (A13 #2279): the WIP cap check and the CI judgments the merge gate shares.
//! Everything here is GET-only — status posting, merging and ruleset changes stay with the
//! leader flow (decision 06:02, 2026-11-09). The integration-gate verdict that lived here
//! (`flow-gate-verdict`) is gone: `gate-check` decides the merge gate, G8 included, and keeps
//! using [`tip_ci_judgment`] (#2674 L3, leader decision 6864050772).
//!
//! GitHub is read through the injected forge transport; tests replay recorded JSON.
use crate::forge::transport::Transport;
use crate::forge::Client;
use serde_json::{json, Value};

type Result<T> = std::result::Result<T, String>;

pub const SCHEMA_VERSION: u32 = 1;
/// The config key names (the numbers themselves are user-specified defaults, not
/// measured baselines — the issue requires them marked as such).
pub const CAP_LANE: &str = "lane_open_pr_cap";
pub const CAP_TOTAL: &str = "total_open_pr_cap";
/// Defaults are the user-approved values (2026-11-09 06:02Z decision: 1 audit-
/// pending PR per lane, 8 total open) — not measured baselines.
pub const DEFAULT_LANE_CAP: u64 = 1;
pub const DEFAULT_TOTAL_CAP: u64 = 8;

fn api(owner: &str, repo: &str, tail: &str) -> String {
    format!("https://api.github.com/repos/{owner}/{repo}/{tail}")
}

/// One open pull request as the WIP check sees it.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct OpenPr {
    pub number: i64,
    pub head_ref: String,
    pub author: String,
    pub draft: bool,
}

fn fetch_open_pulls<T: Transport, C: crate::forge::policy::Clock>(
    client: &Client<'_, T, C>,
    owner: &str,
    repo: &str,
) -> Result<Vec<OpenPr>> {
    let mut out = Vec::new();
    let mut url = Some(api(owner, repo, "pulls?state=open&per_page=100"));
    let mut pages = 0;
    while let Some(u) = url {
        pages += 1;
        if pages > 10 {
            return Err("flow open pull list page limit exceeded (10)".to_string());
        }
        let resp = client.get(&u, None)?;
        let arr = resp
            .body
            .as_ref()
            .and_then(Value::as_array)
            .cloned()
            .unwrap_or_default();
        for p in &arr {
            out.push(OpenPr {
                number: p["number"].as_i64().unwrap_or(0),
                head_ref: p["head"]["ref"].as_str().unwrap_or_default().into(),
                author: p["user"]["login"].as_str().unwrap_or_default().into(),
                draft: p["draft"].as_bool().unwrap_or(false),
            });
        }
        url = resp
            .header("link")
            .and_then(|l| crate::forge::policy::next_link(Some(l)));
    }
    Ok(out)
}

pub(crate) fn fetch_pr<T: Transport, C: crate::forge::policy::Clock>(
    client: &Client<'_, T, C>,
    owner: &str,
    repo: &str,
    number: i64,
) -> Result<Value> {
    let resp = client.get(&api(owner, repo, &format!("pulls/{number}")), None)?;
    resp.body
        .clone()
        .ok_or_else(|| "pr response carried no body".to_string())
}

/// #2563: the aggregate CI check run is renamed `rc-ci` → `ci`. During the transition
/// both spellings name the same logical check: a PR whose head ran before the rename
/// carries `rc-ci`, a newer one carries `ci`. Only these exact names count — a matrix
/// job such as `ci / linux-x86_64` never does.
pub const CI_CHECK_NAMES: [&str; 2] = ["ci", "rc-ci"];

/// Whether a check run named `actual` is a run of the required check `required`.
/// Either CI spelling matches every run of the logical CI check (the union of both
/// names); any other required name matches only itself.
pub fn check_name_matches(required: &str, actual: &str) -> bool {
    if CI_CHECK_NAMES.contains(&required) {
        CI_CHECK_NAMES.contains(&actual)
    } else {
        required == actual
    }
}

/// A PR that repairs a red tip says so with a body line starting `moai-tip-repair:`.
pub const TIP_REPAIR_MARKER: &str = "moai-tip-repair:";

pub fn is_tip_repair(body: &str) -> bool {
    body.lines()
        .any(|l| l.trim_start().starts_with(TIP_REPAIR_MARKER))
}

/// The run id in a check run's `details_url` (`…/actions/runs/<id>/job/<job>`).
fn run_id(run: &Value) -> Option<u64> {
    let url = run["details_url"].as_str()?;
    let rest = url.split("/runs/").nth(1)?;
    rest.split('/').next()?.parse().ok()
}

/// #2665: the tip's own CI, per required check over every attempt on the tip SHA.
/// `green` all success; `red` a failed attempt (failure, timed_out, cancelled — a
/// concurrency cancel leaves the tip without passing CI too — startup_failure,
/// action_required) and no success and none running;
/// `flaky` failure and success on the same SHA (a re-run turned it green — not a stop);
/// `pending` running or not started yet (push CI can start late — not a stop). The
/// worst state across required checks wins: red, then pending, flaky, green.
pub fn tip_ci_judgment(runs: &[Value], required: &[String]) -> Value {
    let failed_conclusion = |c: &str| {
        matches!(
            c,
            "failure" | "timed_out" | "cancelled" | "startup_failure" | "action_required"
        )
    };
    let (mut failed, mut flaky, mut pending) = (Vec::new(), Vec::new(), false);
    for name in required {
        let mine: Vec<&Value> = runs
            .iter()
            .filter(|r| {
                r["name"]
                    .as_str()
                    .is_some_and(|actual| check_name_matches(name, actual))
            })
            .collect();
        let conclusion = |r: &Value| r["conclusion"].as_str().unwrap_or_default().to_string();
        let success = mine.iter().any(|r| conclusion(r) == "success");
        let running = mine.iter().any(|r| conclusion(r).is_empty());
        let fails: Vec<&&Value> = mine
            .iter()
            .filter(|r| failed_conclusion(&conclusion(r)))
            .collect();
        if mine.is_empty() || (running && !success) {
            pending = true;
        } else if !fails.is_empty() && success {
            flaky.push(name.clone());
        } else if !fails.is_empty() {
            for r in fails {
                failed.push(json!({
                    "check": r["name"],
                    "conclusion": conclusion(r),
                    "run_id": run_id(r),
                    "url": r["details_url"],
                    "title": r["output"]["title"],
                }));
            }
        }
    }
    let state = if !failed.is_empty() {
        "red"
    } else if pending {
        "pending"
    } else if !flaky.is_empty() {
        "flaky"
    } else {
        "green"
    };
    json!({"state": state, "failed": failed, "flaky": flaky})
}

// ---- WIP cap (assignment-refusal signal; the refusal itself is the leader's act) ----

fn count_by_author(prs: &[OpenPr], author: &str) -> u64 {
    prs.iter().filter(|p| p.author == author).count() as u64
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct WipVerdict {
    pub lane: String,
    pub lane_open: u64,
    pub total_open: u64,
    pub lane_cap: u64,
    pub total_cap: u64,
    pub exceeded: bool,
}

/// WIP cap check (read-only): one open-PR list call, then the lane count and the
/// total are compared with the caps. `exceeded` is the assignment-refusal signal.
pub fn wip_check<T: Transport, C: crate::forge::policy::Clock>(
    client: &Client<'_, T, C>,
    owner: &str,
    repo: &str,
    lane: &str,
    lane_cap: u64,
    total_cap: u64,
) -> Result<WipVerdict> {
    let prs = fetch_open_pulls(client, owner, repo)?;
    let lane_open = count_by_author(&prs, lane);
    let total_open = prs.len() as u64;
    Ok(WipVerdict {
        lane: lane.to_string(),
        lane_open,
        total_open,
        lane_cap,
        total_cap,
        exceeded: lane_open >= lane_cap || total_open >= total_cap,
    })
}

pub fn wip_check_json<T: Transport, C: crate::forge::policy::Clock>(
    client: &Client<'_, T, C>,
    owner: &str,
    repo: &str,
    lane: &str,
    lane_cap: u64,
    total_cap: u64,
) -> Result<Value> {
    let v = wip_check(client, owner, repo, lane, lane_cap, total_cap)?;
    Ok(json!({
        "schema_version": SCHEMA_VERSION,
        "lane": v.lane,
        "lane_open": v.lane_open,
        "total_open": v.total_open,
        "lane_cap": v.lane_cap,
        "total_cap": v.total_cap,
        "exceeded": v.exceeded,
        "verdict": if v.exceeded { "refuse_assignment" } else { "allow" },
        "writes": false,
    }))
}
