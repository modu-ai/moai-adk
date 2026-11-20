//! `gate.yaml` over the built-in defaults (the Go build's `config.GateConfig`).
//!
//! The YAML subset reader of the status line is included by path, not by `crate::` name, so the
//! tests can compile this module on its own; it is a pure `std` file.

#[path = "../statusline/config.rs"]
pub(super) mod yaml;

use std::path::{Path, PathBuf};
use std::time::Duration;
use yaml::Node;

#[derive(Debug, Clone, PartialEq)]
pub struct GateConfig {
    pub enabled: bool,
    pub skip_tests: bool,
    pub vet_timeout: Duration,
    pub lint_timeout: Duration,
    pub test_timeout: Duration,
    pub typecheck_timeout: Duration,
    /// `Duration::ZERO` means "derive it from the steps".
    pub total_timeout: Duration,
    pub lock_wait: Duration,
    pub typecheck_enabled: bool,
    pub typecheck_command: String,
    /// Step name -> value; an entry whose value is `false` skips that step.
    pub disabled_steps: Vec<(String, bool)>,
    pub pre_commit_enabled: bool,
    pub ast_grep_enabled: bool,
    pub graph_freshness_enabled: bool,
    pub go_build_tags: String,
}

/// The Go build's `NewDefaultGateConfig`.
pub fn defaults() -> GateConfig {
    GateConfig {
        enabled: true,
        skip_tests: false,
        vet_timeout: Duration::from_secs(30),
        lint_timeout: Duration::from_secs(60),
        test_timeout: Duration::from_secs(120),
        typecheck_timeout: Duration::from_secs(300),
        total_timeout: Duration::ZERO,
        lock_wait: Duration::from_secs(30),
        typecheck_enabled: true,
        typecheck_command: String::new(),
        disabled_steps: Vec::new(),
        pre_commit_enabled: false,
        ast_grep_enabled: true,
        graph_freshness_enabled: true,
        go_build_tags: String::new(),
    }
}

/// The project directory: `CLAUDE_PROJECT_DIR` when it is set and not blank (made absolute
/// against `cwd`), else `cwd`.
pub fn resolve_project_dir(claude_project_dir: Option<&str>, cwd: &Path) -> PathBuf {
    match claude_project_dir.map(str::trim).filter(|s| !s.is_empty()) {
        Some(dir) => {
            let p = Path::new(dir);
            if p.is_absolute() {
                p.to_path_buf()
            } else {
                cwd.join(p)
            }
        }
        None => cwd.to_path_buf(),
    }
}

/// The most a timeout may be: a year. A larger value would overflow the clock arithmetic, and
/// no commit hook is meant to wait longer.
const TIMEOUT_CAP_SECS: u64 = 366 * 24 * 3600;

/// `yaml::parse` after dropping what the Go build writes into its config files and this build
/// never reads: an empty flow collection (`key: {}`, `key: []`) becomes a plain `key:`, and a
/// block list (`- item`, with whatever is nested under it) is skipped. A flow collection with
/// content is still outside the subset, so the file is refused as a whole.
pub fn parse_lenient(text: &str) -> Option<Node> {
    let mut out = String::with_capacity(text.len());
    // While skipping a list: the indent of its first `- ` line.
    let mut skipping: Option<usize> = None;
    for line in text.lines() {
        let trimmed = line.trim_start();
        let indent = line.len() - trimmed.len();
        let is_item = trimmed == "-" || trimmed.starts_with("- ");
        if let Some(d) = skipping {
            if !trimmed.is_empty() && !trimmed.starts_with('#') {
                if indent > d || (indent == d && is_item) {
                    continue;
                }
                skipping = None;
            } else {
                continue;
            }
        }
        if is_item {
            skipping = Some(indent);
            continue;
        }
        out.push_str(&empty_flow_to_plain(line));
        out.push('\n');
    }
    yaml::parse(&out)
}

/// `key: {}` / `key: []` (optionally followed by a comment) without the flow part.
fn empty_flow_to_plain(line: &str) -> String {
    let code = line.split(" #").next().unwrap_or(line).trim_end();
    for tail in [" {}", " []"] {
        if let Some(head) = code.strip_suffix(tail) {
            if head.trim_end().ends_with(':') {
                return head.trim_end().to_string();
            }
        }
    }
    line.to_string()
}

/// Reads `<project>/.moai/config/sections/gate.yaml` over the defaults and the Go build tags
/// from `.moai/config/build-tags`. The second value is the warnings to print: a file outside
/// the supported YAML subset is not guessed at, the defaults stand.
pub fn load(project_dir: &Path) -> (GateConfig, Vec<String>) {
    let mut cfg = defaults();
    let mut warnings = Vec::new();
    let file = project_dir.join(".moai/config/sections/gate.yaml");
    match std::fs::read(&file) {
        Ok(bytes) => match String::from_utf8(bytes).ok().as_deref().and_then(parse_lenient) {
            Some(root) => apply(&mut cfg, &root),
            None => warnings.push(format!(
                "moai gate: {} is not UTF-8 or is outside the YAML subset this build reads; using the built-in defaults",
                file.display()
            )),
        },
        Err(_) => {}
    }
    cfg.go_build_tags = read_build_tags(project_dir);
    (cfg, warnings)
}

fn apply(cfg: &mut GateConfig, root: &Node) {
    let flag = |path: &[&str], slot: &mut bool| {
        if let Some(b) = root.scalar(path).and_then(yaml::boolean) {
            *slot = b;
        }
    };
    flag(&["gate", "enabled"], &mut cfg.enabled);
    flag(&["gate", "skip_tests"], &mut cfg.skip_tests);
    flag(
        &["gate", "typecheck", "enabled"],
        &mut cfg.typecheck_enabled,
    );
    flag(
        &["gate", "pre_commit", "enabled"],
        &mut cfg.pre_commit_enabled,
    );
    flag(
        &["gate", "ast_grep_gate", "enabled"],
        &mut cfg.ast_grep_enabled,
    );
    flag(
        &["gate", "graph_freshness", "enabled"],
        &mut cfg.graph_freshness_enabled,
    );
    if let Some(cmd) = root.scalar(&["gate", "typecheck", "command"]) {
        cfg.typecheck_command = cmd.to_string();
    }
    let secs = |key: &str| {
        root.scalar(&["gate", "timeouts", key])
            .and_then(|s| s.parse::<i64>().ok())
    };
    // A zero or negative value is the default (total: "derive it"), as in the Go accessors.
    let pick = |key: &str, default: Duration| match secs(key) {
        Some(n) if n > 0 => Duration::from_secs((n as u64).min(TIMEOUT_CAP_SECS)),
        _ => default,
    };
    cfg.vet_timeout = pick("vet", cfg.vet_timeout);
    cfg.lint_timeout = pick("lint", cfg.lint_timeout);
    cfg.test_timeout = pick("test", cfg.test_timeout);
    cfg.typecheck_timeout = pick("typecheck", cfg.typecheck_timeout);
    cfg.total_timeout = pick("total", Duration::ZERO);
    cfg.lock_wait = pick("lock_wait", cfg.lock_wait);
    if let Some(Node::Map(steps)) = root.get(&["gate", "disabled_steps"]) {
        for (name, value) in steps {
            if let Node::Scalar(v) = value {
                if let Some(b) = yaml::boolean(v) {
                    cfg.disabled_steps.push((name.clone(), b));
                }
            }
        }
    }
}

/// The first non-blank, non-comment line of `.moai/config/build-tags`, trimmed.
fn read_build_tags(project_dir: &Path) -> String {
    let Ok(text) = std::fs::read_to_string(project_dir.join(".moai/config/build-tags")) else {
        return String::new();
    };
    text.lines()
        .map(str::trim)
        .find(|l| !l.is_empty() && !l.starts_with('#'))
        .unwrap_or("")
        .to_string()
}
