//! `moai gate`: the heavy commit-quality gate, run in Rust (SPEC-20261120-rust-quality-gate).
//!
//! It does what the Go build's `moai gate` does: read `gate.yaml`, detect the project's
//! languages, run each language's vet, lint and test steps, print a summary, exit 0 or 1.
//! Axes this build does not run yet are named in the summary, never skipped silently.

pub mod config;
pub mod detect;
mod fmtutil;
pub mod lock;
mod step;
mod summary;
pub mod table;

use config::GateConfig;
use fmtutil::go_duration;
use std::ffi::OsString;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::time::{Duration, Instant};
use step::{Runner, StepCtx};
use summary::{condense_skip_reason, join_blocks, Summary, REASON_SKIP_TESTS};
use table::Step;

/// What one `moai gate` run needs from its caller. `path` is the PATH the steps' tools are
/// looked up on and run with (the real PATH in production, a fixture directory in tests).
pub struct Invocation {
    pub project_dir: PathBuf,
    /// `MOAI_PRECOMMIT=1`: the git pre-commit hook is the caller.
    pub precommit_marker: bool,
    pub path: OsString,
}

const PRECOMMIT_NOTICE: &str = "[moai gate] pre-commit context: heavy gate not opted in — skipping the project-wide steps (opt in via gate.pre_commit.enabled: true in .moai/config/sections/gate.yaml)";

const BUDGET_FLOOR: Duration = Duration::from_secs(10 * 60);
const GRAPH_FRESHNESS_ALLOWANCE: Duration = Duration::from_secs(30);
const AST_GREP_SCAN_TIMEOUT: Duration = Duration::from_secs(30);
const SLACK_MIN: Duration = Duration::from_secs(60);

/// Runs the gate; diagnostics go to `err`; the return value is the process exit code
/// (0 pass, 1 fail).
pub fn run(inv: &Invocation, err: &mut dyn Write) -> i32 {
    let (cfg, warnings) = config::load(&inv.project_dir);
    for w in &warnings {
        let _ = writeln!(err, "{w}");
    }
    if inv.precommit_marker && !cfg.pre_commit_enabled {
        let _ = writeln!(err, "{PRECOMMIT_NOTICE}");
        return 0;
    }
    let waited = lock::wait(&inv.project_dir, cfg.lock_wait, err);
    if !waited.line.is_empty() {
        let _ = writeln!(err, "{}", waited.line);
    }
    let _lock = waited.lock;

    let detected = if cfg.enabled {
        detect::detect(&inv.project_dir)
    } else {
        Vec::new()
    };
    let toolchains: Vec<(table::Toolchain, PathBuf)> = detected
        .iter()
        .map(|d| {
            (
                table::resolve(d.table_index, &d.root, &cfg.go_build_tags),
                d.root.clone(),
            )
        })
        .collect();
    let (budget, warn) = overall_budget(&cfg, &toolchains);
    if !warn.is_empty() {
        let _ = writeln!(err, "{warn}");
    }
    let (passed, output) = gate(&cfg, inv, &toolchains, budget);
    if !output.is_empty() {
        let _ = writeln!(err, "{output}");
    }
    if passed {
        0
    } else {
        1
    }
}

fn gate(
    cfg: &GateConfig,
    inv: &Invocation,
    toolchains: &[(table::Toolchain, PathBuf)],
    budget: Duration,
) -> (bool, String) {
    if !cfg.enabled || toolchains.is_empty() {
        return (true, String::new());
    }
    let mut runner = Runner::new(
        cfg,
        StepCtx {
            deadline: Instant::now() + budget,
            path: inv.path.clone(),
        },
    );
    seed(&mut runner.summary, cfg, toolchains);
    if cfg.graph_freshness_enabled {
        runner.summary.axis_not_ported("graph-freshness");
    }
    if cfg.ast_grep_enabled {
        runner.summary.axis_not_ported("ast-grep");
    }
    let mut notes = String::new();
    for (tc, root) in toolchains {
        match static_steps(&mut runner, tc, root) {
            Ok(out) => notes = append(&notes, &out),
            Err(out) => return (false, with_summary(&out, &runner.summary)),
        }
    }
    for (tc, root) in toolchains {
        let Some(test) = &tc.test else { continue };
        if cfg.skip_tests {
            runner.summary.skipped(&test.name, REASON_SKIP_TESTS);
            continue;
        }
        if tc.node {
            runner.summary.not_ported(&test.name);
            continue;
        }
        let (ok, out) = runner.execute_step(test, root, cfg.test_timeout);
        if !ok {
            return (false, with_summary(&out, &runner.summary));
        }
        notes = append(&notes, &out);
    }
    (true, with_summary(&notes, &runner.summary))
}

fn with_summary(out: &str, summary: &Summary) -> String {
    join_blocks(out, &summary.render())
}

fn append(existing: &str, add: &str) -> String {
    let add = add.trim();
    if add.is_empty() {
        existing.to_string()
    } else if existing.trim().is_empty() {
        add.to_string()
    } else {
        format!("{existing}\n{add}")
    }
}

/// The summary rows in the order the Go build seeds them: every vet step, the type check, then
/// each language's lint steps and test step.
fn seed(summary: &mut Summary, cfg: &GateConfig, toolchains: &[(table::Toolchain, PathBuf)]) {
    for (tc, _) in toolchains {
        for s in &tc.vet {
            summary.seed(&s.name);
        }
    }
    if cfg.typecheck_enabled {
        summary.seed("typecheck");
    }
    for (tc, _) in toolchains {
        for s in &tc.lint {
            summary.seed(&s.name);
        }
        if let Some(t) = &tc.test {
            summary.seed(&t.name);
        }
    }
}

/// Vet, type check and lint of one language: `Ok(notes)` or `Err(failure text)`.
fn static_steps(runner: &mut Runner, tc: &table::Toolchain, root: &Path) -> Result<String, String> {
    let cfg = runner.cfg;
    let mut notes = String::new();
    for s in &tc.vet {
        let (ok, out) = runner.execute_step(s, root, cfg.vet_timeout);
        if !ok {
            return Err(out);
        }
        notes = append(&notes, &out);
    }
    if cfg.typecheck_enabled {
        let command = cfg.typecheck_command.trim();
        if !command.is_empty() {
            let mut fields = command.split_whitespace();
            let step = Step {
                name: "typecheck".into(),
                binary: fields.next().unwrap_or_default().into(),
                args: fields.map(String::from).collect(),
                optional: true,
                ..Step::default()
            };
            let (ok, out) = runner.execute_step(&step, root, cfg.typecheck_timeout);
            if !ok {
                return Err(out);
            }
            notes = append(&notes, &out);
        } else if tc.node {
            runner.summary.not_ported("typecheck");
        } else {
            let reason = "typecheck: skipped (no default for this language; set gate.typecheck.command to enable one)";
            runner
                .summary
                .skipped("typecheck", &condense_skip_reason("typecheck", reason));
            notes = append(&notes, reason);
        }
    }
    for s in &tc.lint {
        if tc.node {
            runner.summary.not_ported(&s.name);
            continue;
        }
        let (ok, out) = runner.execute_step(s, root, cfg.lint_timeout);
        if !ok {
            return Err(out);
        }
        notes = append(&notes, &out);
    }
    Ok(notes)
}

/// The time the whole run may take: the sum of the step limits plus slack, at least ten
/// minutes and at least the largest single step, or `gate.timeouts.total` when it is set.
fn overall_budget(
    cfg: &GateConfig,
    toolchains: &[(table::Toolchain, PathBuf)],
) -> (Duration, String) {
    if !cfg.enabled {
        return (BUDGET_FLOOR, String::new());
    }
    let mut sum = Duration::ZERO;
    let mut largest = Duration::ZERO;
    let mut note = |d: Duration, n: usize| {
        if n > 0 {
            sum += d * n as u32;
            largest = largest.max(d);
        }
    };
    for (tc, _) in toolchains {
        note(cfg.vet_timeout, tc.vet.len());
        if cfg.typecheck_enabled && !cfg.typecheck_command.trim().is_empty() {
            note(cfg.typecheck_timeout, 1);
        }
        note(cfg.lint_timeout, tc.lint.len());
        if tc.test.is_some() && !cfg.skip_tests {
            note(cfg.test_timeout, 1);
        }
    }
    if cfg.graph_freshness_enabled {
        note(GRAPH_FRESHNESS_ALLOWANCE, 1);
    }
    if cfg.ast_grep_enabled {
        note(AST_GREP_SCAN_TIMEOUT, 1);
    }
    let derived = (sum + SLACK_MIN.max(sum / 10))
        .max(BUDGET_FLOOR)
        .max(largest);
    if cfg.total_timeout.is_zero() {
        return (derived, String::new());
    }
    if cfg.total_timeout < largest {
        return (
            largest,
            format!(
                "gate: timeouts.total={} is below the largest single step budget {}; using {}",
                go_duration(cfg.total_timeout),
                go_duration(largest),
                go_duration(largest)
            ),
        );
    }
    (cfg.total_timeout, String::new())
}
