//! The workflow files the CI design generates (SPEC-20261122-project-ci-workflows): the text of
//! `moai-check.yml`, `moai-nightly.yml`, `moai-deploy.yml` and, on the user's pick,
//! `dependabot.yml`, built from the saved answers, the declared checks and the project's work
//! flow. Pure: it reads no file, runs nothing and writes nothing; the plan and apply that put
//! the text on disk come later.
//!
//! Every generated file passes [`lint`] before it is returned: a file that breaks one of the
//! safety rules is refused, never emitted (REQ-5). The work-flow mapping is the one table
//! [`FLOWS`], so a renamed flow changes one row.

use crate::project_ci::{CheckEntry, CiAnswers};

type Result<T> = std::result::Result<T, String>;

/// A generated file: its path under the project and its text.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct File {
    pub path: String,
    pub text: String,
}

/// What [`generate`] returns: the files in a fixed order and the plain-words notes the preview
/// shows beside them.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Generated {
    pub files: Vec<File>,
    pub notes: Vec<String>,
}

/// What the generator reads.
pub struct Input<'a> {
    /// The profile's `work_flow.flow`; `None` or an unknown value gives pull-request checks only.
    pub flow: Option<&'a str>,
    /// The profile's integration branch; `main` when none was named.
    pub integration_branch: Option<&'a str>,
    pub answers: &'a CiAnswers,
    pub checks: &'a [CheckEntry],
    /// The detected languages of the project, for the dependency cache.
    pub languages: &'a [String],
    /// The language of the comment block at the top of each file.
    pub comments_in_korean: bool,
}

/// `actions/checkout` v7.0.1, pinned to the full commit (the tag's commit, read from the
/// official repository on 2026-11-22).
pub const CHECKOUT_PIN: &str = "actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1";
/// `actions/cache` v6.1.0, pinned to the full commit (read the same way).
pub const CACHE_PIN: &str = "actions/cache@55cc8345863c7cc4c66a329aec7e433d2d1c52a9 # v6.1.0";

/// Which pull requests a flow checks.
enum PullRequests {
    /// No pull-request trigger.
    No,
    /// Every pull request, whatever its target.
    Any,
    /// Pull requests into these branches; `@integration` is the profile's integration branch.
    Into(&'static [&'static str]),
}

/// One row of the work-flow mapping: the triggers of a flow.
struct FlowRow {
    flow: &'static str,
    /// Branches whose pushes run the check; `@integration` is the integration branch.
    push: &'static [&'static str],
    pull_requests: PullRequests,
    /// The flow adds one nightly full run.
    nightly: bool,
    /// A newer push cancels the older run of the same pull request (never a running push).
    cancel_per_pull_request: bool,
}

const INTEGRATION: &str = "@integration";

/// The one table that maps a work flow to its triggers (REQ-1, REQ-2). The profile's `m-flow`
/// value becomes the NeXT-Flow value when the rename lands: only its row's name changes here.
const FLOWS: &[FlowRow] = &[
    FlowRow {
        flow: "m-flow",
        push: &[INTEGRATION],
        pull_requests: PullRequests::No,
        nightly: true,
        cancel_per_pull_request: false,
    },
    FlowRow {
        flow: "github-flow",
        push: &[INTEGRATION],
        pull_requests: PullRequests::Into(&[INTEGRATION]),
        nightly: false,
        cancel_per_pull_request: true,
    },
    FlowRow {
        flow: "trunk-based",
        push: &[INTEGRATION],
        pull_requests: PullRequests::Any,
        nightly: false,
        cancel_per_pull_request: true,
    },
    FlowRow {
        flow: "git-flow",
        push: &["develop"],
        pull_requests: PullRequests::Into(&["develop", "release/*", INTEGRATION]),
        nightly: false,
        cancel_per_pull_request: true,
    },
];

/// The row for a flow with no entry in [`FLOWS`]: pull-request checks only.
const UNKNOWN_FLOW: FlowRow = FlowRow {
    flow: "",
    push: &[],
    pull_requests: PullRequests::Any,
    nightly: false,
    cancel_per_pull_request: true,
};

/// Integration branch when the profile names none.
const DEFAULT_BRANCH: &str = "main";
/// The nightly run starts at this minute, off the hour, when runs queue up least.
const NIGHTLY_CRON: &str = "17 3 * * *";
/// Job timeouts in minutes.
const CHECK_TIMEOUT: u32 = 20;
const NIGHTLY_TIMEOUT: u32 = 45;
const DEPLOY_TIMEOUT: u32 = 30;

/// A dependency cache: the stack's name, the folders to keep and the files whose hash is the key.
struct Stack {
    name: &'static str,
    paths: &'static [&'static str],
    keys: &'static [&'static str],
}

fn stack_for(language: &str, checks: &[CheckEntry]) -> Option<Stack> {
    let uses = |word: &str| {
        checks
            .iter()
            .any(|c| c.command.split_whitespace().next() == Some(word))
    };
    Some(match language {
        "rust" => Stack {
            name: "rust",
            paths: &["~/.cargo/registry", "~/.cargo/git", "target"],
            keys: &["**/Cargo.lock"],
        },
        "javascript" | "typescript" if uses("pnpm") => Stack {
            name: "pnpm",
            paths: &["~/.local/share/pnpm/store"],
            keys: &["**/pnpm-lock.yaml"],
        },
        "javascript" | "typescript" if uses("yarn") => Stack {
            name: "yarn",
            paths: &["~/.cache/yarn"],
            keys: &["**/yarn.lock"],
        },
        "javascript" | "typescript" if uses("bun") => Stack {
            name: "bun",
            paths: &["~/.bun/install/cache"],
            keys: &["**/bun.lock", "**/bun.lockb"],
        },
        "javascript" | "typescript" => Stack {
            name: "npm",
            paths: &["~/.npm"],
            keys: &["**/package-lock.json"],
        },
        "python" => Stack {
            name: "python",
            paths: &["~/.cache/pip"],
            keys: &["**/requirements*.txt", "**/pyproject.toml"],
        },
        "go" => Stack {
            name: "go",
            paths: &["~/go/pkg/mod", "~/.cache/go-build"],
            keys: &["**/go.sum"],
        },
        "java" | "kotlin" if checks.iter().any(|c| c.command.contains("gradle")) => Stack {
            name: "gradle",
            paths: &["~/.gradle/caches", "~/.gradle/wrapper"],
            keys: &["**/*.gradle*", "**/gradle-wrapper.properties"],
        },
        "java" | "kotlin" => Stack {
            name: "maven",
            paths: &["~/.m2/repository"],
            keys: &["**/pom.xml"],
        },
        "ruby" => Stack {
            name: "ruby",
            paths: &["vendor/bundle"],
            keys: &["**/Gemfile.lock"],
        },
        "php" => Stack {
            name: "php",
            paths: &["~/.composer/cache/files"],
            keys: &["**/composer.lock"],
        },
        _ => return None,
    })
}

/// The one cache step of a job: every detected stack in one step, one key.
fn cache_step(languages: &[String], checks: &[CheckEntry]) -> String {
    let mut stacks: Vec<Stack> = Vec::new();
    for language in languages {
        if let Some(stack) = stack_for(language, checks) {
            if !stacks.iter().any(|s| s.name == stack.name) {
                stacks.push(stack);
            }
        }
    }
    stacks.sort_by_key(|s| s.name);
    let (label, paths, keys): (String, Vec<&str>, Vec<&str>) = if stacks.is_empty() {
        ("generic".into(), vec!["~/.cache"], vec!["**/*.lock"])
    } else {
        (
            stacks.iter().map(|s| s.name).collect::<Vec<_>>().join("-"),
            stacks
                .iter()
                .flat_map(|s| s.paths.iter().copied())
                .collect(),
            stacks.iter().flat_map(|s| s.keys.iter().copied()).collect(),
        )
    };
    let hash = keys
        .iter()
        .map(|k| format!("'{k}'"))
        .collect::<Vec<_>>()
        .join(", ");
    let mut out = String::new();
    out.push_str("      - name: \"Cache dependencies\"\n");
    out.push_str(&format!("        uses: {CACHE_PIN}\n"));
    out.push_str("        with:\n          path: |\n");
    for p in paths {
        out.push_str(&format!("            {p}\n"));
    }
    out.push_str(&format!(
        "          key: ${{{{ runner.os }}}}-moai-{label}-${{{{ hashFiles({hash}) }}}}\n"
    ));
    out.push_str(&format!(
        "          restore-keys: |\n            ${{{{ runner.os }}}}-moai-{label}-\n"
    ));
    out
}

/// The plain-words comment block at the top of a file: what it does and how to turn it off.
fn header(extra: &[&str], korean: bool) -> String {
    let mut lines: Vec<&str> = if korean {
        vec![
            "MoAI 가 만든 파일이에요. 프로젝트에 선언한 검사 명령을 선언한 순서대로 실행해요.",
            "프로필에 선언한 명령만 실행하고 다른 것은 실행하지 않아요.",
            "끄려면 이 파일을 지우거나, Actions 탭에서 이 워크플로를 골라 Disable workflow 를 누르세요.",
        ]
    } else {
        vec![
            "MoAI generated this file. It runs the checks you declared for this project, in the order you declared them.",
            "Only the commands in your project profile run here; nothing else does.",
            "To turn it off: delete this file, or open the Actions tab, pick this workflow and choose Disable workflow.",
        ]
    };
    lines.extend_from_slice(extra);
    let mut out = String::new();
    for l in lines {
        out.push_str(&format!("# {l}\n"));
    }
    out.push('\n');
    out
}

/// The checkout step: pinned, and it leaves no credentials in the workspace.
fn checkout_step() -> String {
    format!("      - uses: {CHECKOUT_PIN}\n        with:\n          persist-credentials: false\n")
}

fn steps_for(languages: &[String], checks: &[CheckEntry], only_build: bool) -> String {
    let mut out = format!("    steps:\n{}", checkout_step());
    out.push_str(&cache_step(languages, checks));
    for c in checks.iter().filter(|c| !only_build || c.id == "build") {
        out.push_str(&format!(
            "      - name: \"{}\"\n        run: |\n          {}\n",
            c.id, c.command
        ));
    }
    out
}

fn quoted(list: &[String]) -> String {
    let items: Vec<String> = list.iter().map(|b| format!("\"{b}\"")).collect();
    format!("[{}]", items.join(", "))
}

/// Builds the files for an input. Every returned file has passed [`lint`].
pub fn generate(input: &Input) -> Result<Generated> {
    crate::project_ci::validate_answers(input.answers)?;
    crate::project_ci::validate_checks(input.checks)
        .map_err(|_| "no valid declared checks; nothing was planned".to_string())?;
    let integration = input.integration_branch.unwrap_or(DEFAULT_BRANCH);
    if !crate::profile::branch_name(integration) {
        return Err("invalid integration branch; no values were logged".into());
    }
    let known = input
        .flow
        .and_then(|f| FLOWS.iter().find(|row| row.flow == f));
    let row = known.unwrap_or(&UNKNOWN_FLOW);
    let resolve = |names: &[&str]| -> Vec<String> {
        names
            .iter()
            .map(|n| if *n == INTEGRATION { integration } else { n }.to_string())
            .collect()
    };
    let korean = input.comments_in_korean;
    let mut notes = Vec::new();
    if known.is_none() {
        notes.push(match input.flow {
            Some(flow) => format!(
                "The work flow \"{flow}\" has no CI row yet, so the generated check runs on pull requests only (pull-request checks only)."
            ),
            None => "No work flow is recorded, so the generated check runs on pull requests only (pull-request checks only).".to_string(),
        });
    }

    let mut on = String::from("on:\n");
    if !row.push.is_empty() {
        on.push_str(&format!(
            "  push:\n    branches: {}\n",
            quoted(&resolve(row.push))
        ));
    }
    match &row.pull_requests {
        PullRequests::No => {}
        PullRequests::Any => {
            on.push_str("  pull_request:\n    types: [opened, synchronize, reopened]\n")
        }
        PullRequests::Into(branches) => on.push_str(&format!(
            "  pull_request:\n    branches: {}\n",
            quoted(&resolve(branches))
        )),
    }
    let concurrency = if row.cancel_per_pull_request {
        "concurrency:\n  group: ${{ github.workflow }}-${{ github.event.pull_request.number || github.ref }}\n  cancel-in-progress: ${{ github.event_name == 'pull_request' }}\n"
    } else {
        "concurrency:\n  group: ${{ github.workflow }}-${{ github.ref }}\n  cancel-in-progress: false\n"
    };
    let mut check = header(&[], korean);
    check.push_str("name: \"MoAI check\"\n");
    check.push_str(&on);
    check.push_str("permissions:\n  contents: read\n");
    check.push_str(concurrency);
    check.push_str("jobs:\n  check:\n    name: \"check\"\n    runs-on: ubuntu-latest\n");
    check.push_str(&format!("    timeout-minutes: {CHECK_TIMEOUT}\n"));
    check.push_str(&steps_for(input.languages, input.checks, false));

    let has_build = input.checks.iter().any(|c| c.id == "build");
    for (system, runner) in [("windows", "windows-latest"), ("macos", "macos-latest")] {
        if !input.answers.systems.iter().any(|s| s == system) {
            continue;
        }
        if !has_build {
            notes.push(format!(
                "{system} was named, but no build check is declared, so no {system} job was planned."
            ));
            continue;
        }
        check.push_str(&format!(
            "  build-{system}:\n    name: \"build-{system}\"\n    runs-on: {runner}\n"
        ));
        check.push_str(&format!("    timeout-minutes: {CHECK_TIMEOUT}\n"));
        check.push_str("    defaults:\n      run:\n        shell: bash\n");
        check.push_str(&steps_for(input.languages, input.checks, true));
    }
    if input.answers.systems.iter().any(|s| s == "macos") {
        let p = crate::project_ci_estimate::prices().per_minute_mills;
        notes.push(format!(
            "macOS jobs bill about {} dollars per minute (Linux {}, Windows {}); a private repository pays it and a public one does not.",
            crate::project_ci_estimate::dollars(p.macos),
            crate::project_ci_estimate::dollars(p.linux),
            crate::project_ci_estimate::dollars(p.windows),
        ));
    }
    let mut files = vec![File {
        path: ".github/workflows/moai-check.yml".into(),
        text: check,
    }];

    if row.nightly {
        let mut nightly = header(
            &[if korean {
                "공개 저장소는 활동이 없으면 60일 뒤 예약 실행이 저절로 꺼져요."
            } else {
                "A public repository's scheduled run switches itself off after 60 days without activity."
            }],
            korean,
        );
        nightly.push_str("name: \"MoAI nightly\"\n");
        nightly.push_str(&format!(
            "on:\n  schedule:\n    - cron: \"{NIGHTLY_CRON}\"\n"
        ));
        nightly.push_str("permissions:\n  contents: read\n");
        nightly.push_str(
            "concurrency:\n  group: ${{ github.workflow }}\n  cancel-in-progress: false\n",
        );
        nightly.push_str("jobs:\n  nightly:\n    name: \"nightly\"\n    runs-on: ubuntu-latest\n");
        nightly.push_str(&format!("    timeout-minutes: {NIGHTLY_TIMEOUT}\n"));
        nightly.push_str(&steps_for(input.languages, input.checks, false));
        files.push(File {
            path: ".github/workflows/moai-nightly.yml".into(),
            text: nightly,
        });
        notes.push(
            "The nightly run is a scheduled workflow; a scheduled run of a public repository switches itself off after 60 days without activity. A failed run only reports and blocks no landing.".into(),
        );
    }

    if let Some(target) = input
        .answers
        .deploy
        .target
        .as_deref()
        .filter(|_| input.answers.deploy.enabled)
    {
        let mut deploy = header(
            &[if korean {
                "자동으로 시작하지 않아요. Actions 탭에서 직접 실행할 때만 돌아요."
            } else {
                "It never starts by itself: it runs only when you start it from the Actions tab."
            }],
            korean,
        );
        deploy.push_str("name: \"MoAI deploy\"\non:\n  workflow_dispatch:\n");
        deploy.push_str("permissions:\n  contents: read\n");
        deploy.push_str(
            "concurrency:\n  group: ${{ github.workflow }}\n  cancel-in-progress: false\n",
        );
        deploy.push_str("jobs:\n  deploy:\n    name: \"deploy\"\n    runs-on: ubuntu-latest\n");
        deploy.push_str(&format!("    environment: \"{target}\"\n"));
        deploy.push_str(&format!("    timeout-minutes: {DEPLOY_TIMEOUT}\n"));
        deploy.push_str(&format!("    steps:\n{}", checkout_step()));
        deploy.push_str(&cache_step(input.languages, input.checks));
        // No deploy command is declared, so none is generated: the job body is a comment that
        // tells the user where their own steps go (a run step would be an undeclared command).
        deploy.push_str(
            "      # Add your own deploy steps here. MoAI generates no deploy command because none is declared.\n",
        );
        files.push(File {
            path: ".github/workflows/moai-deploy.yml".into(),
            text: deploy,
        });
        notes.push(format!(
            "Add required reviewers to the environment \"{target}\" so a deploy waits for a person: Settings, Environments, {target}, Required reviewers. That is a repository setting this tool cannot apply, and on private repositories it needs a plan that offers it."
        ));
        notes.push(
            "No secret is referenced. If your deploy steps need one, add it under Settings, Secrets and variables, Actions, New repository secret, and use it by name in the deploy steps.".into(),
        );
    }

    if input.answers.dependabot == Some(true) {
        let mut text = String::new();
        for l in [
            "MoAI generated this file. Dependabot keeps the pinned versions of the actions in your workflows up to date.",
            "It covers the GitHub Actions ecosystem only and changes no other file.",
            "To turn it off: delete this file.",
        ] {
            text.push_str(&format!("# {l}\n"));
        }
        text.push_str("\nversion: 2\nupdates:\n  - package-ecosystem: \"github-actions\"\n    directory: \"/\"\n    schedule:\n      interval: \"weekly\"\n");
        files.push(File {
            path: ".github/dependabot.yml".into(),
            text,
        });
    }

    if input.answers.strict == "block" {
        notes.push(
            "To block merges on failed checks, add the job \"check\" as a required status check: Settings, Branches, your branch rule, Require status checks to pass, then pick \"check\". This tool shows the steps and applies nothing.".into(),
        );
    }

    for f in &files {
        let findings = lint(&f.text);
        if !findings.is_empty() {
            return Err(format!(
                "generated file {} breaks a safety rule: {}",
                f.path,
                findings.join("; ")
            ));
        }
    }
    Ok(Generated { files, notes })
}

/// The indent (leading spaces) of a line.
fn indent(line: &str) -> usize {
    line.len() - line.trim_start().len()
}

/// The findings of the safety rules for one generated file, empty when it is clean.
pub fn lint(text: &str) -> Vec<String> {
    let mut found = Vec::new();
    let lines: Vec<&str> = text.lines().collect();
    if lines.len() < 3 || !lines.iter().take(3).all(|l| l.starts_with('#')) {
        found.push("the file must begin with a comment block of plain words".to_string());
    }
    if crate::secret_shaped(text) {
        found.push("the file holds text in the shape of a secret".to_string());
    }
    if text.contains("pull_request_target") {
        found.push("the pull_request_target trigger is never used".to_string());
    }
    if lines.iter().any(|l| l.trim() == "inputs:") {
        found.push("no input may reach a shell".to_string());
    }
    // A skipped workflow leaves a required check pending, so no path filter is ever written.
    if lines.iter().any(|l| {
        let t = l.trim_start();
        t.starts_with("paths:") || t.starts_with("paths-ignore:")
    }) {
        found.push("no path filter may be used".to_string());
    }
    for (i, l) in lines.iter().enumerate() {
        if l.contains("uses: actions/checkout@")
            && !(lines.get(i + 1) == Some(&"        with:")
                && lines.get(i + 2) == Some(&"          persist-credentials: false"))
        {
            found.push("the checkout must set persist-credentials to false".to_string());
        }
    }
    // A file with no `jobs:` (the Dependabot file) has none of the workflow rules.
    let Some(jobs_at) = lines.iter().position(|l| *l == "jobs:") else {
        return found;
    };
    if !text.contains("\npermissions:\n  contents: read\n") {
        found.push("the top level must set read-only permissions".to_string());
    }
    if lines.iter().any(|l| l.trim_end().ends_with(": write")) {
        found.push("no permission may be write".to_string());
    }
    if !text.contains("\nconcurrency:\n") {
        found.push("the workflow needs a concurrency group".to_string());
    }
    for l in &lines {
        let t = l.trim_start().trim_start_matches("- ");
        if let Some(value) = t.strip_prefix("uses:") {
            let value = value.trim();
            let pinned = value.split('@').nth(1).is_some_and(|rest| {
                rest.len() > 40
                    && rest.as_bytes()[..40].iter().all(u8::is_ascii_hexdigit)
                    && rest[40..].starts_with(" # v")
            });
            if !pinned {
                found.push(format!("an action is not pinned to a full commit: {value}"));
            }
        }
    }
    let mut i = 0;
    while i < lines.len() {
        let line = lines[i];
        let t = line.trim_start().trim_start_matches("- ");
        if let Some(value) = t.strip_prefix("run:") {
            let value = value.trim();
            let base = indent(line);
            let mut body = String::from(value);
            if matches!(value, "|" | "|-" | ">" | ">-") {
                body.clear();
                while i + 1 < lines.len()
                    && (lines[i + 1].trim().is_empty() || indent(lines[i + 1]) > base)
                {
                    i += 1;
                    body.push_str(lines[i]);
                    body.push('\n');
                }
            }
            if body.contains("${{") {
                found.push("event data or an expression may not appear in a run value".to_string());
            }
        }
        i += 1;
    }
    // Per job: a timeout and a cache step.
    // (job key, has a timeout, has a cache, has its stable name)
    let mut jobs: Vec<(String, bool, bool, bool)> = Vec::new();
    for l in &lines[jobs_at + 1..] {
        if indent(l) == 2 && l.trim_end().ends_with(':') && !l.trim().is_empty() {
            jobs.push((l.trim().to_string(), false, false, false));
        } else if let Some(job) = jobs.last_mut() {
            if l.starts_with("    timeout-minutes: ") {
                job.1 = true;
            }
            if l.contains("uses: actions/cache@") {
                job.2 = true;
            }
            // The required-check name: the job key `check` must be named exactly "check".
            if *l == "    name: \"check\"" {
                job.3 = true;
            }
        }
    }
    if jobs.is_empty() {
        found.push("the workflow has no job".to_string());
    }
    for (name, timeout, cache, stable) in jobs {
        if !timeout {
            found.push(format!("job {name} has no timeout"));
        }
        if !cache {
            found.push(format!("job {name} has no dependency cache"));
        }
        if name == "check:" && !stable {
            found.push("the check job must be named exactly \"check\"".to_string());
        }
    }
    found
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::project_ci::CiDeploy;

    fn answers() -> CiAnswers {
        CiAnswers {
            kind: "cli".into(),
            systems: vec!["linux".into()],
            deploy: CiDeploy {
                enabled: false,
                target: None,
            },
            visibility: "public".into(),
            monthly_budget_usd: None,
            macos_price_ack: None,
            dependabot: None,
            strict: "report".into(),
        }
    }

    fn checks() -> Vec<CheckEntry> {
        [
            ("lint", "cargo clippy --all-targets -- -D warnings"),
            ("build", "cargo build"),
            ("test", "cargo test"),
        ]
        .iter()
        .map(|(id, command)| CheckEntry {
            id: (*id).into(),
            command: (*command).into(),
        })
        .collect()
    }

    fn languages() -> Vec<String> {
        vec!["rust".into()]
    }

    fn make(flow: Option<&str>, a: &CiAnswers, c: &[CheckEntry], langs: &[String]) -> Generated {
        generate(&Input {
            flow,
            integration_branch: Some("trunk"),
            answers: a,
            checks: c,
            languages: langs,
            comments_in_korean: false,
        })
        .expect("generates")
    }

    fn file<'a>(g: &'a Generated, name: &str) -> &'a str {
        g.files
            .iter()
            .find(|f| f.path == format!(".github/workflows/{name}"))
            .map(|f| f.text.as_str())
            .unwrap_or_else(|| {
                panic!(
                    "no {name} in {:?}",
                    g.files.iter().map(|f| &f.path).collect::<Vec<_>>()
                )
            })
    }

    fn has_file(g: &Generated, path: &str) -> bool {
        g.files.iter().any(|f| f.path == path)
    }

    /// The `on:` block of a file: from the line `on:` to the next top-level key.
    fn on_block(text: &str) -> String {
        let mut out = Vec::new();
        let mut inside = false;
        for line in text.lines() {
            if line == "on:" {
                inside = true;
                continue;
            }
            if inside && !line.starts_with(' ') && !line.is_empty() {
                break;
            }
            if inside {
                out.push(line);
            }
        }
        out.join("\n")
    }

    /// REQ-1, AC-1: each work-flow value gets the triggers of the design table; the integration
    /// branch comes from the profile, never a fixed name.
    #[test]
    fn each_flow_gets_the_triggers_of_its_row() {
        let c = checks();
        let l = languages();
        let m = make(Some("m-flow"), &answers(), &c, &l);
        let on = on_block(file(&m, "moai-check.yml"));
        assert!(on.contains("  push:\n    branches: [\"trunk\"]"), "{on}");
        assert!(
            !on.contains("pull_request"),
            "m-flow has no pull_request: {on}"
        );

        let g = make(Some("github-flow"), &answers(), &c, &l);
        let on = on_block(file(&g, "moai-check.yml"));
        assert!(on.contains("  push:\n    branches: [\"trunk\"]"), "{on}");
        assert!(
            on.contains("  pull_request:\n    branches: [\"trunk\"]"),
            "{on}"
        );

        let t = make(Some("trunk-based"), &answers(), &c, &l);
        let on = on_block(file(&t, "moai-check.yml"));
        assert!(on.contains("  push:\n    branches: [\"trunk\"]"), "{on}");
        assert!(on.contains("  pull_request:"), "{on}");
        assert!(!on.contains("pull_request:\n    branches"), "{on}");

        let f = make(Some("git-flow"), &answers(), &c, &l);
        let on = on_block(file(&f, "moai-check.yml"));
        assert!(on.contains("  push:\n    branches: [\"develop\"]"), "{on}");
        assert!(
            on.contains("  pull_request:\n    branches: [\"develop\", \"release/*\", \"trunk\"]"),
            "{on}"
        );
    }

    /// REQ-1 failure line: an unknown or absent flow still gets pull-request checks, and the
    /// preview says so.
    #[test]
    fn an_unknown_flow_generates_pull_request_checks_only_and_says_so() {
        for flow in [Some("kanban"), None] {
            let g = make(flow, &answers(), &checks(), &languages());
            let on = on_block(file(&g, "moai-check.yml"));
            assert!(on.contains("  pull_request:"), "{on}");
            assert!(!on.contains("push:"), "{on}");
            assert!(
                g.notes
                    .iter()
                    .any(|n| n.contains("pull-request checks only")),
                "{:?}",
                g.notes
            );
        }
    }

    /// REQ-2: m-flow alone adds the nightly full run, at a minute that is not on the hour, and
    /// the notes say that a public repository's scheduled run switches itself off.
    #[test]
    fn only_m_flow_gets_the_nightly_run_at_an_off_peak_minute() {
        let c = checks();
        let l = languages();
        let m = make(Some("m-flow"), &answers(), &c, &l);
        let nightly = file(&m, "moai-nightly.yml");
        let cron = nightly
            .lines()
            .find(|l| l.contains("cron:"))
            .expect("a schedule");
        let minute: u32 = cron
            .split('"')
            .nth(1)
            .and_then(|s| s.split(' ').next())
            .and_then(|s| s.parse().ok())
            .expect("a numeric minute");
        assert!(minute != 0 && minute != 30 && minute < 60, "{cron}");
        assert!(
            m.notes.iter().any(|n| n.contains("60 days")),
            "{:?}",
            m.notes
        );
        for flow in ["github-flow", "trunk-based", "git-flow"] {
            let g = make(Some(flow), &answers(), &c, &l);
            assert!(
                !has_file(&g, ".github/workflows/moai-nightly.yml"),
                "{flow}"
            );
            assert!(!g.notes.iter().any(|n| n.contains("60 days")), "{flow}");
        }
    }

    /// REQ-3: one Linux runner by default; another system appears only when named, and then
    /// builds without running the tests; macOS shows its price.
    #[test]
    fn other_systems_appear_only_when_named_and_only_build() {
        let c = checks();
        let l = languages();
        let plain = make(Some("github-flow"), &answers(), &c, &l);
        let text = file(&plain, "moai-check.yml");
        assert!(text.contains("runs-on: ubuntu-latest"));
        assert!(!text.contains("windows-latest") && !text.contains("macos-latest"));

        let mut a = answers();
        a.systems = vec!["linux".into(), "macos".into(), "windows".into()];
        a.macos_price_ack = Some(true);
        let many = make(Some("github-flow"), &a, &c, &l);
        let text = file(&many, "moai-check.yml");
        let section = |job: &str| -> String {
            let start = text
                .find(&format!("  {job}:\n"))
                .unwrap_or_else(|| panic!("no job {job}"));
            let rest = &text[start + 1..];
            let end = rest.find("\n  build-").map_or(rest.len(), |i| i + 1);
            rest[..end].to_string()
        };
        for (job, runner) in [
            ("build-windows", "windows-latest"),
            ("build-macos", "macos-latest"),
        ] {
            let s = section(job);
            assert!(s.contains(&format!("runs-on: {runner}")), "{s}");
            assert!(
                s.contains("cargo build") && !s.contains("cargo test"),
                "{s}"
            );
            assert!(!s.contains("clippy"), "{s}");
        }
        assert!(
            many.notes
                .iter()
                .any(|n| n.contains("macOS") && n.contains("0.062")),
            "{:?}",
            many.notes
        );
        // No declared build check: no extra system job, and the notes say why.
        let no_build: Vec<CheckEntry> = c.iter().filter(|c| c.id != "build").cloned().collect();
        let none = make(Some("github-flow"), &a, &no_build, &l);
        assert!(!file(&none, "moai-check.yml").contains("build-windows"));
        assert!(
            none.notes.iter().any(|n| n.contains("no build check")),
            "{:?}",
            none.notes
        );
    }

    /// REQ-4, AC-2: the four project kinds over several stacks and every system answer; every
    /// file passes the lint, and every job carries a timeout, a cache and its workflow a
    /// concurrency group.
    #[test]
    fn every_generated_job_has_a_timeout_a_cache_and_a_concurrency_group() {
        let stacks: &[(&[&str], &str)] = &[
            (&["rust"], "cargo build"),
            (&["javascript"], "pnpm run build"),
            (&["python"], "pytest"),
            (&["go"], "go build ./..."),
            (&["java"], "./gradlew build"),
            (&[], "make build"),
        ];
        for kind in ["web", "api", "cli", "library"] {
            for (langs, build) in stacks {
                for flow in ["m-flow", "github-flow", "trunk-based", "git-flow"] {
                    let mut a = answers();
                    a.kind = kind.into();
                    a.systems = vec!["linux".into(), "macos".into(), "windows".into()];
                    a.macos_price_ack = Some(true);
                    a.deploy = CiDeploy {
                        enabled: true,
                        target: Some("production".into()),
                    };
                    a.dependabot = Some(true);
                    let c = vec![
                        CheckEntry {
                            id: "build".into(),
                            command: (*build).into(),
                        },
                        CheckEntry {
                            id: "test".into(),
                            command: "true".into(),
                        },
                    ];
                    let l: Vec<String> = langs.iter().map(|s| s.to_string()).collect();
                    let g = make(Some(flow), &a, &c, &l);
                    for f in &g.files {
                        assert_eq!(
                            lint(&f.text),
                            Vec::<String>::new(),
                            "{} {kind} {flow}",
                            f.path
                        );
                        if f.path.ends_with("dependabot.yml") {
                            continue;
                        }
                        let in_jobs = f.text.split("\njobs:\n").nth(1).expect("jobs");
                        let job_names = in_jobs
                            .lines()
                            .filter(|l| {
                                l.starts_with("  ") && !l.starts_with("   ") && l.ends_with(':')
                            })
                            .count();
                        assert_eq!(
                            job_names,
                            in_jobs.matches("    timeout-minutes: ").count(),
                            "{}",
                            f.path
                        );
                        assert_eq!(
                            job_names,
                            in_jobs.matches("uses: actions/cache@").count(),
                            "{}",
                            f.path
                        );
                        assert!(f.text.contains("\nconcurrency:\n"), "{}", f.path);
                    }
                }
            }
        }
    }

    /// REQ-4 (design): the run steps are exactly the declared commands, in the declared order,
    /// and nothing else runs.
    #[test]
    fn the_run_steps_are_the_declared_commands_in_the_declared_order() {
        let c = checks();
        let g = make(Some("github-flow"), &answers(), &c, &languages());
        let text = file(&g, "moai-check.yml");
        let runs: Vec<&str> = text
            .lines()
            .scan(false, |after_run, l| {
                let was = *after_run;
                *after_run = l.trim_end().ends_with("run: |");
                Some(if was { Some(l.trim()) } else { None })
            })
            .flatten()
            .collect();
        let want: Vec<&str> = c.iter().map(|c| c.command.as_str()).collect();
        assert_eq!(runs, want);
    }

    /// REQ-5, AC-3: the file begins with plain words, the token is read-only at the top, every
    /// action is pinned to a full commit, no event data reaches a run value and the dangerous
    /// trigger never appears.
    #[test]
    fn the_safety_rules_hold_for_every_generated_file() {
        let mut a = answers();
        a.deploy = CiDeploy {
            enabled: true,
            target: Some("production".into()),
        };
        a.dependabot = Some(true);
        let g = make(Some("m-flow"), &a, &checks(), &languages());
        assert_eq!(
            g.files.len(),
            4,
            "{:?}",
            g.files.iter().map(|f| &f.path).collect::<Vec<_>>()
        );
        for f in &g.files {
            assert!(f.text.starts_with("# "), "{}", f.path);
            assert!(
                f.text.lines().take(3).all(|l| l.starts_with('#')),
                "{}",
                f.path
            );
            assert!(!f.text.contains("pull_request_target"), "{}", f.path);
            assert!(!f.text.contains("inputs:"), "{}", f.path);
            if f.path.ends_with("dependabot.yml") {
                continue;
            }
            assert!(
                f.text.contains("\npermissions:\n  contents: read\n"),
                "{}",
                f.path
            );
            for l in f.text.lines().filter(|l| l.contains("uses:")) {
                let at = l.split('@').nth(1).expect("a pin");
                assert!(
                    at.len() > 40 && at[..40].bytes().all(|b| b.is_ascii_hexdigit()),
                    "{l}"
                );
                assert!(l.contains(" # v"), "{l}");
            }
        }
        // The header says how to turn the file off.
        let header: String = file(&g, "moai-check.yml")
            .lines()
            .take_while(|l| l.starts_with('#'))
            .collect::<Vec<_>>()
            .join("\n");
        assert!(
            header.contains("delete this file") && header.contains("Disable workflow"),
            "{header}"
        );
        let ko = generate(&Input {
            flow: Some("github-flow"),
            integration_branch: None,
            answers: &answers(),
            checks: &checks(),
            languages: &languages(),
            comments_in_korean: true,
        })
        .unwrap();
        assert!(
            file(&ko, "moai-check.yml").starts_with("# MoAI"),
            "{}",
            file(&ko, "moai-check.yml")
        );
        assert!(
            file(&ko, "moai-check.yml").contains("파일을 지우"),
            "korean header"
        );
    }

    /// REQ-5: the lint names each rule it enforces; a text that breaks one is refused.
    #[test]
    fn the_lint_finds_each_broken_rule() {
        let good = file(
            &make(Some("github-flow"), &answers(), &checks(), &languages()),
            "moai-check.yml",
        )
        .to_string();
        assert_eq!(lint(&good), Vec::<String>::new());
        let broken: &[(&str, String)] = &[
            (
                "no comment block",
                good.replacen("# ", "", 1)
                    .trim_start_matches('#')
                    .to_string(),
            ),
            (
                "write permission",
                good.replace("contents: read", "contents: write"),
            ),
            (
                "no top-level permissions",
                good.replace("\npermissions:\n  contents: read\n", "\n"),
            ),
            (
                "tag pin",
                good.replace("3d3c42e5aac5ba805825da76410c181273ba90b1", "v7"),
            ),
            (
                "short pin",
                good.replace("3d3c42e5aac5ba805825da76410c181273ba90b1", "3d3c42e"),
            ),
            (
                "event data in run",
                good.replace("cargo build", "echo ${{ github.event.pull_request.title }}"),
            ),
            (
                "dangerous trigger",
                good.replace("pull_request:", "pull_request_target:"),
            ),
            ("no timeout", good.replace("timeout-minutes", "x-timeout")),
            ("no cache", good.replace("actions/cache@", "actions/other@")),
            (
                "no concurrency",
                good.replace("\nconcurrency:\n", "\nx-concurrency:\n"),
            ),
            ("a key shape", format!("{good}# sk-abc123def\n")),
            ("an input", format!("{good}  inputs:\n")),
            // Audit B-N1: a path filter lets a required check be skipped and stay pending, and
            // a changed job name breaks the required-check name.
            (
                "a path filter",
                good.replace(
                    "  pull_request:\n",
                    "  pull_request:\n    paths: [\"src/**\"]\n",
                ),
            ),
            (
                "a path-ignore filter",
                good.replace(
                    "  pull_request:\n",
                    "  pull_request:\n    paths-ignore: [\"docs/**\"]\n",
                ),
            ),
            (
                "a renamed check job",
                good.replace("name: \"check\"", "name: \"Check job\""),
            ),
            (
                "a check job without a name",
                good.replace("    name: \"check\"\n", ""),
            ),
            // Audit B-N5: the checkout keeps no credentials in the workspace.
            (
                "credentials kept by checkout",
                good.replace("persist-credentials: false", "persist-credentials: true"),
            ),
            (
                "checkout without the setting",
                good.replace("        with:\n          persist-credentials: false\n", ""),
            ),
        ];
        for (why, text) in broken {
            assert!(!lint(text).is_empty(), "{why} passed the lint");
        }
    }

    /// REQ-7, AC-4: a deploy workflow exists only after the yes, names its environment and
    /// never starts by itself.
    #[test]
    fn the_deploy_workflow_exists_only_after_the_yes_and_names_its_environment() {
        let c = checks();
        let l = languages();
        let no = make(Some("github-flow"), &answers(), &c, &l);
        assert!(!has_file(&no, ".github/workflows/moai-deploy.yml"));
        assert!(!no.files.iter().any(|f| f.text.contains("environment:")));
        let mut a = answers();
        a.deploy = CiDeploy {
            enabled: true,
            target: Some("staging".into()),
        };
        let yes = make(Some("github-flow"), &a, &c, &l);
        let text = file(&yes, "moai-deploy.yml");
        assert!(text.contains("    environment: \"staging\"\n"), "{text}");
        // Audit B-N3: no run step of an undeclared command; the job body is a comment that
        // tells the user where their own deploy steps go.
        assert!(!text.contains("run:") && !text.contains("echo"), "{text}");
        assert!(
            text.contains("\n      # Add your own deploy steps here"),
            "{text}"
        );
        assert!(on_block(text).contains("workflow_dispatch"), "{text}");
        assert!(
            !on_block(text).contains("push") && !on_block(text).contains("pull_request"),
            "{text}"
        );
        assert!(
            yes.notes.iter().any(|n| n.contains("required reviewers")),
            "{:?}",
            yes.notes
        );
        assert!(
            !text.contains("secrets."),
            "no secret is referenced: {text}"
        );
    }

    /// REQ-8, AC-5: the blocking choice keeps one stable job name and no path filter, and the
    /// settings steps are text that names the job; no workflow ever filters on paths.
    #[test]
    fn the_blocking_choice_keeps_one_job_name_and_no_path_filter() {
        for strict in ["report", "block"] {
            let mut a = answers();
            a.strict = strict.into();
            for flow in ["m-flow", "github-flow", "trunk-based", "git-flow"] {
                let g = make(Some(flow), &a, &checks(), &languages());
                let text = file(&g, "moai-check.yml");
                assert!(
                    text.contains("\n  check:\n    name: \"check\"\n"),
                    "{flow} {strict}"
                );
                for f in &g.files {
                    assert!(
                        !f.text.contains("paths:") && !f.text.contains("paths-ignore:"),
                        "{}",
                        f.path
                    );
                }
                let mentions = g
                    .notes
                    .iter()
                    .any(|n| n.contains("required status check") && n.contains("\"check\""));
                assert_eq!(
                    mentions,
                    strict == "block",
                    "{flow} {strict}: {:?}",
                    g.notes
                );
            }
        }
    }

    /// The Dependabot file is the user's separate pick and covers the Actions ecosystem only.
    #[test]
    fn dependabot_is_written_only_on_the_pick_and_for_actions_only() {
        let g = make(Some("github-flow"), &answers(), &checks(), &languages());
        assert!(!has_file(&g, ".github/dependabot.yml"));
        let mut a = answers();
        a.dependabot = Some(true);
        let g = make(Some("github-flow"), &a, &checks(), &languages());
        let text = g
            .files
            .iter()
            .find(|f| f.path == ".github/dependabot.yml")
            .expect("file")
            .text
            .clone();
        assert!(
            text.contains("package-ecosystem: \"github-actions\""),
            "{text}"
        );
        assert_eq!(text.matches("package-ecosystem").count(), 1, "{text}");
    }

    /// The generator checks its input again: a command or a name the profile would refuse is an
    /// error here too, and the text is the same on every run.
    #[test]
    fn the_input_is_checked_again_and_the_output_is_stable() {
        let bad = vec![CheckEntry {
            id: "t".into(),
            command: "echo ${{ x }}".into(),
        }];
        assert!(generate(&Input {
            flow: Some("github-flow"),
            integration_branch: None,
            answers: &answers(),
            checks: &bad,
            languages: &[],
            comments_in_korean: false,
        })
        .is_err());
        for branch in ["", "-x", "a b", "a\"b", "x/../y", "x\ny"] {
            assert!(
                generate(&Input {
                    flow: Some("github-flow"),
                    integration_branch: Some(branch),
                    answers: &answers(),
                    checks: &checks(),
                    languages: &[],
                    comments_in_korean: false,
                })
                .is_err(),
                "{branch:?}"
            );
        }
        let none: Vec<CheckEntry> = Vec::new();
        assert!(
            generate(&Input {
                flow: Some("github-flow"),
                integration_branch: None,
                answers: &answers(),
                checks: &none,
                languages: &[],
                comments_in_korean: false,
            })
            .is_err(),
            "a project with no declared checks gets no planned check"
        );
        let a = make(Some("m-flow"), &answers(), &checks(), &languages());
        let b = make(Some("m-flow"), &answers(), &checks(), &languages());
        assert_eq!(a, b);
    }

    fn all_features() -> CiAnswers {
        let mut a = answers();
        a.systems = vec!["linux".into(), "macos".into(), "windows".into()];
        a.macos_price_ack = Some(true);
        a.deploy = CiDeploy {
            enabled: true,
            target: Some("production".into()),
        };
        a.dependabot = Some(true);
        a
    }

    /// Audit B-N2 (mutant 6a): the generator checks the answers again; an answer set the
    /// profile would refuse is an error here whatever the caller did.
    #[test]
    fn the_generator_checks_the_answers_again() {
        let broken: Vec<(&str, CiAnswers)> = vec![
            (
                "unknown kind",
                CiAnswers {
                    kind: "game".into(),
                    ..answers()
                },
            ),
            (
                "macos without its yes",
                CiAnswers {
                    systems: vec!["linux".into(), "macos".into()],
                    ..answers()
                },
            ),
            (
                "no linux",
                CiAnswers {
                    systems: vec!["windows".into()],
                    ..answers()
                },
            ),
            (
                "unknown strictness",
                CiAnswers {
                    strict: "x".into(),
                    ..answers()
                },
            ),
            (
                "a deploy place with a space",
                CiAnswers {
                    deploy: CiDeploy {
                        enabled: true,
                        target: Some("my site".into()),
                    },
                    ..answers()
                },
            ),
            (
                "a deploy place in the shape of a key",
                CiAnswers {
                    deploy: CiDeploy {
                        enabled: true,
                        target: Some("ghp_abc".into()),
                    },
                    ..answers()
                },
            ),
            (
                "a budget on a public repository",
                CiAnswers {
                    monthly_budget_usd: Some(5),
                    ..answers()
                },
            ),
        ];
        for (why, a) in &broken {
            let out = generate(&Input {
                flow: Some("github-flow"),
                integration_branch: None,
                answers: a,
                checks: &checks(),
                languages: &languages(),
                comments_in_korean: false,
            });
            assert!(out.is_err(), "{why} was accepted");
        }
    }

    /// Audit B-N3: every run step of every generated file is one of the declared commands, in
    /// a job that is not the deploy job; nothing else is ever run.
    #[test]
    fn every_run_step_in_every_file_is_a_declared_command() {
        let c = checks();
        for flow in ["m-flow", "github-flow", "trunk-based", "git-flow"] {
            let g = make(Some(flow), &all_features(), &c, &languages());
            for f in &g.files {
                let mut after_run = false;
                for l in f.text.lines() {
                    if after_run {
                        assert!(
                            c.iter().any(|c| c.command == l.trim()),
                            "{flow} {}: an undeclared run step: {l}",
                            f.path
                        );
                    }
                    after_run = l.trim_end().ends_with("run: |");
                }
            }
        }
    }

    /// Audit B-N5: every checkout leaves no credentials in the workspace.
    #[test]
    fn checkout_keeps_no_credentials() {
        let g = make(Some("m-flow"), &all_features(), &checks(), &languages());
        let mut seen = 0;
        for f in &g.files {
            let lines: Vec<&str> = f.text.lines().collect();
            for (i, l) in lines.iter().enumerate() {
                if l.contains("uses: actions/checkout@") {
                    seen += 1;
                    assert_eq!(lines[i + 1], "        with:", "{}", f.path);
                    assert_eq!(
                        lines[i + 2],
                        "          persist-credentials: false",
                        "{}",
                        f.path
                    );
                }
            }
        }
        assert!(
            seen >= 4,
            "the check, build, nightly and deploy jobs: {seen}"
        );
    }

    /// Audit B-N4: names and the environment are quoted, so no value can change the shape of the
    /// YAML around it.
    #[test]
    fn names_and_environments_are_quoted() {
        let g = make(Some("m-flow"), &all_features(), &checks(), &languages());
        for f in g
            .files
            .iter()
            .filter(|f| !f.path.ends_with("dependabot.yml"))
        {
            for l in f.text.lines() {
                let t = l.trim_start().trim_start_matches("- ");
                for key in ["name:", "environment:"] {
                    if let Some(value) = t.strip_prefix(key) {
                        let value = value.trim();
                        assert!(
                            value.len() >= 2 && value.starts_with('"') && value.ends_with('"'),
                            "{}: unquoted {key} {value}",
                            f.path
                        );
                    }
                }
            }
        }
        let text = file(&g, "moai-deploy.yml");
        assert!(
            text.contains("\n    environment: \"production\"\n"),
            "{text}"
        );
    }
}
