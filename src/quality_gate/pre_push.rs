//! `moai hook pre-push`: commit-subject convention check run by the git pre-push hook
//! (SPEC-20261120-rust-pre-push). Same call as the Go build: subjects on standard input, one
//! per line; same exit codes (0 pass or warn, 2 enforced violation, 1 load error).
//!
//! The built-in conventions are matched by hand: the crate carries no regex dependency and the
//! three patterns are small enough to write out (the recorded Go table guards the equivalence).

use super::config::yaml;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::Command;

/// Caller-supplied environment, so tests do not depend on the process environment.
#[derive(Debug, Default, Clone)]
pub struct PrePushEnv {
    /// `MOAI_ENFORCE_ON_PUSH`
    pub enforce_on_push: Option<String>,
    /// `MOAI_PRE_PUSH`
    pub pre_push_mode: Option<String>,
    /// `MOAI_GIT_CONVENTION`
    pub convention: Option<String>,
}

pub struct PrePushInvocation {
    /// The working directory: the configuration is read here, as the Go build does.
    pub project_dir: PathBuf,
    /// `CLAUDE_PROJECT_DIR` when it is set: the repository whose history auto detection samples.
    /// `None` means `project_dir`.
    pub repo_dir: Option<PathBuf>,
    pub env: PrePushEnv,
    /// Everything read from standard input.
    pub stdin: Vec<u8>,
}

#[derive(Clone, Copy, PartialEq, Eq, Debug)]
enum Action {
    Skip,
    Warn,
    Enforce,
}

fn parse_action(v: &str) -> Option<Action> {
    match v {
        "skip" => Some(Action::Skip),
        "warn" => Some(Action::Warn),
        "enforce" => Some(Action::Enforce),
        _ => None,
    }
}

#[derive(Clone, Copy, PartialEq, Eq)]
enum ScopeKind {
    /// `\(.+\)`
    Any,
    /// `\([a-z-]+\)`
    LowerDash,
}

#[derive(Clone)]
struct Convention {
    name: &'static str,
    /// The pattern as Go prints it in a violation.
    pattern: &'static str,
    types: &'static [&'static str],
    scope_kind: ScopeKind,
    /// Whether `!` (breaking change) may precede the colon: `!?` in the pattern.
    bang: bool,
    max_length: usize,
    examples: &'static [&'static str],
}

const BUILTIN_ORDER: &[&str] = &["conventional-commits", "angular", "karma"];

fn builtin(name: &str) -> Option<Convention> {
    Some(match name {
        "conventional-commits" => Convention {
            name: "conventional-commits",
            pattern: r"^(build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test)(\(.+\))?!?: .+",
            types: &[
                "build", "chore", "ci", "docs", "feat", "fix", "perf", "refactor", "revert",
                "style", "test",
            ],
            scope_kind: ScopeKind::Any,
            bang: true,
            max_length: 100,
            examples: &[
                "feat(auth): add JWT token validation",
                "fix: resolve null pointer in user service",
                "docs(readme): update installation guide",
            ],
        },
        "angular" => Convention {
            name: "angular",
            pattern: r"^(build|ci|docs|feat|fix|perf|refactor|test)(\([a-z-]+\))?: .+",
            types: &[
                "build", "ci", "docs", "feat", "fix", "perf", "refactor", "test",
            ],
            scope_kind: ScopeKind::LowerDash,
            bang: false,
            max_length: 100,
            examples: &[
                "feat(router): add lazy loading support",
                "fix(compiler): handle edge case in template parser",
            ],
        },
        "karma" => Convention {
            name: "karma",
            pattern: r"^(feat|fix|docs|style|refactor|perf|test|chore)(\(.+\))?: .+",
            types: &[
                "feat", "fix", "docs", "style", "refactor", "perf", "test", "chore",
            ],
            scope_kind: ScopeKind::Any,
            bang: false,
            max_length: 100,
            examples: &[
                "feat(service): add user notification endpoint",
                "test(api): add integration tests for auth module",
            ],
        },
        _ => return None,
    })
}

/// `: ` then at least one character, after an optional `!`.
fn tail_ok(s: &str, bang: bool) -> bool {
    let after_bang = if bang { s.strip_prefix('!') } else { None };
    [Some(s), after_bang]
        .into_iter()
        .flatten()
        .any(|t| t.strip_prefix(": ").is_some_and(|r| !r.is_empty()))
}

fn scope_ok(scope: &str, kind: ScopeKind) -> bool {
    !scope.is_empty()
        && match kind {
            ScopeKind::Any => true,
            ScopeKind::LowerDash => scope.bytes().all(|b| b.is_ascii_lowercase() || b == b'-'),
        }
}

fn pattern_matches(header: &str, conv: &Convention) -> bool {
    conv.types.iter().any(|t| {
        let Some(rest) = header.strip_prefix(t) else {
            return false;
        };
        if rest.starts_with('(') {
            // The scope runs to a `)` that is followed by the tail; try every `)`.
            for (k, c) in rest.char_indices() {
                if c == ')'
                    && k >= 2
                    && scope_ok(&rest[1..k], conv.scope_kind)
                    && tail_ok(&rest[k + 1..], conv.bang)
                {
                    return true;
                }
            }
        }
        tail_ok(rest, conv.bang)
    })
}

#[derive(Debug)]
enum Violation {
    Pattern {
        expected: String,
        suggestion: Vec<u8>,
    },
    InvalidType {
        expected: String,
        actual: String,
    },
    MaxLength {
        expected: String,
        actual: String,
    },
    Required,
}

struct ValidationResult {
    message: Vec<u8>,
    violations: Vec<Violation>,
}

fn extract_type(header: &str) -> &str {
    header
        .char_indices()
        .find(|(_, c)| matches!(c, '(' | ':' | '!'))
        .map(|(i, _)| &header[..i])
        .unwrap_or("")
}

/// Go's `suggestFix`, as bytes: the first byte is lower-cased and a first byte that starts a
/// multi-byte character becomes U+FFFD while its continuation bytes stay (Go slices a byte).
fn suggest_fix(header: &str) -> Vec<u8> {
    // Go's `strings.ToLower` maps each rune to its simple lower case, so `İ` becomes `i`; Rust's
    // full mapping would give `i` plus a combining dot.
    let lower: String = header
        .chars()
        .map(|c| c.to_lowercase().next().unwrap_or(c))
        .collect();
    let kind = if lower.contains("fix") || lower.contains("bug") {
        "fix"
    } else if lower.contains("add") || lower.contains("feat") || lower.contains("new") {
        "feat"
    } else if lower.contains("doc") || lower.contains("readme") {
        "docs"
    } else if lower.contains("test") {
        "test"
    } else if lower.contains("refactor") || lower.contains("clean") {
        "refactor"
    } else {
        "chore"
    };
    let desc = header.trim().as_bytes();
    let mut out = format!("{kind}: ").into_bytes();
    if let Some((&first, rest)) = desc.split_first() {
        if first.is_ascii() {
            out.push(first.to_ascii_lowercase());
        } else {
            out.extend_from_slice("\u{FFFD}".as_bytes());
        }
        out.extend_from_slice(rest);
    }
    out
}

/// Go's `strings.TrimSpace` on bytes: a valid UTF-8 line is trimmed by Unicode white space, any
/// other line by ASCII white space only.
fn trim_subject(line: &[u8]) -> &[u8] {
    match std::str::from_utf8(line) {
        Ok(s) => s.trim().as_bytes(),
        Err(_) => {
            let start = line
                .iter()
                .position(|b| !b.is_ascii_whitespace())
                .unwrap_or(line.len());
            let end = line
                .iter()
                .rposition(|b| !b.is_ascii_whitespace())
                .map_or(start, |i| i + 1);
            &line[start..end]
        }
    }
}

fn validate(message: &[u8], conv: &Convention) -> ValidationResult {
    let first_line = message.split(|b| *b == b'\n').next().unwrap_or(&[]);
    let raw_header = trim_subject(first_line);
    let lossy = String::from_utf8_lossy(raw_header);
    let header: &str = &lossy;
    let mut violations = Vec::new();
    if header.is_empty() {
        violations.push(Violation::Required);
        return ValidationResult {
            message: message.to_vec(),
            violations,
        };
    }
    if conv.max_length > 0 && raw_header.len() > conv.max_length {
        violations.push(Violation::MaxLength {
            expected: format!("max {} characters", conv.max_length),
            actual: format!("{} characters", raw_header.len()),
        });
    }
    if !pattern_matches(header, conv) {
        violations.push(Violation::Pattern {
            expected: conv.pattern.to_string(),
            suggestion: suggest_fix(header),
        });
    } else {
        let ty = extract_type(header);
        if !conv.types.is_empty() && !ty.is_empty() && !conv.types.contains(&ty) {
            violations.push(Violation::InvalidType {
                expected: conv.types.join(", "),
                actual: ty.to_string(),
            });
        }
    }
    ValidationResult {
        message: message.to_vec(),
        violations,
    }
}

fn is_valid(r: &ValidationResult) -> bool {
    r.violations.is_empty()
}

/// Whether Go's `strconv.IsPrint` would print `c` as it is: white space other than the plain
/// space, control characters, the format characters and the private-use area are not printable.
/// (An approximation of the Unicode tables Go carries, covering the characters a commit
/// subject can plausibly hold.)
fn go_printable(c: char) -> bool {
    let u = c as u32;
    if c == ' ' {
        return true;
    }
    if c.is_control() || c.is_whitespace() {
        return false;
    }
    !matches!(
        u,
        0xad | 0x600..=0x605
            | 0x61c
            | 0x6dd
            | 0x70f
            | 0x180e
            | 0x200b..=0x200f
            | 0x202a..=0x202e
            | 0x2060..=0x2064
            | 0x2066..=0x206f
            | 0xfeff
            | 0xfff9..=0xfffb
            | 0xe000..=0xf8ff
            | 0xf0000..=0x10ffff
    )
}

/// Go's `%q` (`strconv.Quote`) on a string that may not be valid UTF-8: a byte that is not part
/// of a character becomes `\xNN`.
fn quote(s: &[u8]) -> String {
    let mut out = String::from("\"");
    let mut rest = s;
    while !rest.is_empty() {
        let (valid, bad) = match std::str::from_utf8(rest) {
            Ok(v) => (v, 0),
            Err(e) => {
                let upto = e.valid_up_to();
                let bad = e.error_len().unwrap_or(rest.len() - upto);
                (std::str::from_utf8(&rest[..upto]).unwrap_or(""), bad)
            }
        };
        for c in valid.chars() {
            match c {
                '"' => out.push_str("\\\""),
                '\\' => out.push_str("\\\\"),
                '\x07' => out.push_str("\\a"),
                '\x08' => out.push_str("\\b"),
                '\x0c' => out.push_str("\\f"),
                '\n' => out.push_str("\\n"),
                '\r' => out.push_str("\\r"),
                '\t' => out.push_str("\\t"),
                '\x0b' => out.push_str("\\v"),
                c if (c as u32) < 0x20 || c as u32 == 0x7f => {
                    out.push_str(&format!("\\x{:02x}", c as u32))
                }
                c if go_printable(c) => out.push(c),
                c if (c as u32) < 0x10000 => out.push_str(&format!("\\u{:04x}", c as u32)),
                c => out.push_str(&format!("\\U{:08x}", c as u32)),
            }
        }
        let used = valid.len();
        for b in &rest[used..used + bad] {
            out.push_str(&format!("\\x{b:02x}"));
        }
        rest = &rest[used + bad..];
    }
    out.push('"');
    out
}

fn format_violation(v: &Violation) -> (&'static str, Vec<u8>) {
    match v {
        Violation::Pattern {
            expected,
            suggestion,
        } => {
            let mut s = format!("Pattern mismatch (expected: {expected})").into_bytes();
            if !suggestion.is_empty() {
                s.extend_from_slice(b"\n    Suggestion: ");
                s.extend_from_slice(suggestion);
            }
            ("[PATTERN]", s)
        }
        Violation::InvalidType { expected, actual } => (
            "[TYPE]",
            format!(
                "Invalid type {} (allowed: {expected})",
                quote(actual.as_bytes())
            )
            .into_bytes(),
        ),
        Violation::MaxLength { expected, actual } => (
            "[LENGTH]",
            format!("Header too long ({actual}, {expected})").into_bytes(),
        ),
        Violation::Required => ("[REQUIRED]", b"Missing required field: header".to_vec()),
    }
}

/// The report of one violating message, as bytes (a Go suggestion can hold raw bytes).
fn format_error(r: &ValidationResult, conv: &Convention) -> Vec<u8> {
    let mut b = format!(
        "Commit message violates {} convention:\n  Message: {}\n\n",
        conv.name,
        quote(&r.message)
    )
    .into_bytes();
    for v in &r.violations {
        let (icon, text) = format_violation(v);
        b.extend_from_slice(format!("  {icon} ").as_bytes());
        b.extend_from_slice(&text);
        b.push(b'\n');
    }
    if !conv.examples.is_empty() {
        b.extend_from_slice(b"\n  Examples of valid messages:\n");
        for ex in conv.examples {
            b.extend_from_slice(format!("    - {ex}\n").as_bytes());
        }
    }
    b
}

// ---- settings --------------------------------------------------------------------------

struct Settings {
    enforce_on_push: bool,
    convention: String,
    detect_enabled: bool,
    sample_size: i64,
    threshold: f64,
    fallback: String,
    max_length: i64,
    /// `git_strategy.mode`
    mode: String,
    /// `git_strategy.<mode>.hooks.pre_push` for the three known modes.
    hook_value: Option<String>,
}

fn read_node(path: &Path, warnings: &mut Vec<String>) -> Option<yaml::Node> {
    let bytes = std::fs::read(path).ok()?;
    let node = String::from_utf8(bytes)
        .ok()
        .as_deref()
        .and_then(super::config::parse_lenient);
    if node.is_none() {
        warnings.push(format!(
            "moai hook pre-push: {} is not UTF-8 or is outside the YAML subset this build reads; using the built-in defaults",
            path.display()
        ));
    }
    node
}

/// What the Go build's config loader refuses: a convention outside the built-in names, a
/// fallback outside the three conventions, a negative sample size or length, or a threshold
/// outside 0..1. The loader then fails as a whole.
fn config_is_valid(s: &Settings) -> bool {
    let convention_ok = s.convention == "auto" || builtin(&s.convention).is_some();
    let fallback_ok = s.fallback.is_empty() || builtin(&s.fallback).is_some();
    convention_ok
        && fallback_ok
        && s.sample_size >= 0
        && (0.0..=1.0).contains(&s.threshold)
        && s.max_length >= 0
}

/// Go has no config when its loader fails: the compiled-in defaults apply, the switch can then
/// come from the environment only, and there is no profile, which reads as `enforce`. A hook
/// that cannot read its settings fails closed, as the Go build does.
fn settings(project_dir: &Path, warnings: &mut Vec<String>) -> Settings {
    let s = settings_raw(project_dir, warnings);
    if config_is_valid(&s) {
        return s;
    }
    Settings {
        enforce_on_push: false,
        convention: "auto".into(),
        detect_enabled: true,
        sample_size: 100,
        threshold: 0.5,
        fallback: "conventional-commits".into(),
        max_length: 100,
        mode: s.mode,
        hook_value: None,
    }
}

fn settings_raw(project_dir: &Path, warnings: &mut Vec<String>) -> Settings {
    let dir = project_dir.join(".moai/config/sections");
    let conv = read_node(&dir.join("git-convention.yaml"), warnings);
    let strat = read_node(&dir.join("git-strategy.yaml"), warnings);
    let get = |n: &Option<yaml::Node>, path: &[&str]| {
        n.as_ref().and_then(|n| n.scalar(path)).map(str::to_string)
    };
    let flag = |n: &Option<yaml::Node>, path: &[&str], default: bool| {
        get(n, path)
            .and_then(|v| yaml::boolean(&v))
            .unwrap_or(default)
    };
    let mode = get(&strat, &["git_strategy", "mode"]).unwrap_or_else(|| "team".into());
    let hook_value = if matches!(mode.as_str(), "manual" | "personal" | "team") {
        Some(
            get(&strat, &["git_strategy", &mode, "hooks", "pre_push"])
                .unwrap_or_else(|| "warn".into()),
        )
    } else {
        None
    };
    Settings {
        enforce_on_push: flag(
            &conv,
            &["git_convention", "validation", "enforce_on_push"],
            false,
        ),
        convention: get(&conv, &["git_convention", "convention"])
            .filter(|c| !c.is_empty())
            .unwrap_or_else(|| "auto".into()),
        detect_enabled: flag(
            &conv,
            &["git_convention", "auto_detection", "enabled"],
            true,
        ),
        sample_size: get(&conv, &["git_convention", "auto_detection", "sample_size"])
            .and_then(|v| v.parse().ok())
            .unwrap_or(100),
        threshold: get(
            &conv,
            &["git_convention", "auto_detection", "confidence_threshold"],
        )
        .and_then(|v| v.parse().ok())
        .unwrap_or(0.5),
        fallback: get(&conv, &["git_convention", "auto_detection", "fallback"])
            .unwrap_or_else(|| "conventional-commits".into()),
        max_length: get(&conv, &["git_convention", "validation", "max_length"])
            .and_then(|v| v.parse().ok())
            .unwrap_or(100),
        mode,
        hook_value,
    }
}

fn enforcement_on(env: &PrePushEnv, s: &Settings) -> bool {
    match env.enforce_on_push.as_deref().filter(|v| !v.is_empty()) {
        Some(v) => v == "true" || v == "1",
        None => s.enforce_on_push,
    }
}

fn resolve_action(env: &PrePushEnv, s: &Settings) -> Action {
    if let Some(a) = env
        .pre_push_mode
        .as_deref()
        .filter(|v| !v.is_empty())
        .and_then(parse_action)
    {
        return a;
    }
    match &s.hook_value {
        Some(v) => parse_action(v).unwrap_or(Action::Enforce),
        // An unknown mode has no profile.
        None => Action::Enforce,
    }
}

// ---- convention loading ------------------------------------------------------------------

fn load_fallback(name: &str) -> Result<Convention, String> {
    let name = if name.is_empty() {
        "conventional-commits"
    } else {
        name
    };
    builtin(name).ok_or_else(|| {
        format!(
            "load convention: fallback \"{name}\" failed: unknown built-in convention: \"{name}\""
        )
    })
}

fn recent_subjects(repo: &Path, limit: i64) -> Option<Vec<String>> {
    let mut cmd = Command::new("git");
    super::step::scrub_git_env(&mut cmd);
    let out = cmd
        .arg("-C")
        .arg(repo)
        .args(["log", &format!("--max-count={limit}"), "--format=%s"])
        .output()
        .ok()
        .filter(|o| o.status.success())?;
    let text = String::from_utf8_lossy(&out.stdout).into_owned();
    Some(
        text.trim()
            .split('\n')
            .filter(|l| !l.trim().is_empty())
            .map(String::from)
            .collect(),
    )
}

fn resolve_auto(repo: &Path, s: &Settings) -> Result<Convention, String> {
    if !s.detect_enabled {
        return load_fallback(&s.fallback);
    }
    let sample = if s.sample_size <= 0 {
        100
    } else {
        s.sample_size
    };
    let Some(subjects) = recent_subjects(repo, sample).filter(|m| !m.is_empty()) else {
        return load_fallback(&s.fallback);
    };
    let mut best: Option<(Convention, f64)> = None;
    for name in BUILTIN_ORDER {
        let conv = builtin(name).expect("built-in");
        let matched = subjects
            .iter()
            .filter(|m| is_valid(&validate(m.as_bytes(), &conv)))
            .count();
        let confidence = matched as f64 / subjects.len() as f64;
        if best.as_ref().is_none_or(|(_, c)| confidence > *c) {
            best = Some((conv, confidence));
        }
    }
    match best {
        Some((conv, confidence)) if confidence >= s.threshold => Ok(conv),
        _ => load_fallback(&s.fallback),
    }
}

/// Whether `run` will read standard input for this directory and environment: enforcement on
/// and an action other than `skip`. A caller that wires standard input asks first, so a hook
/// that is off never blocks on an open terminal.
pub fn reads_input(project_dir: &Path, env: &PrePushEnv) -> bool {
    let mut warnings = Vec::new();
    let s = settings(project_dir, &mut warnings);
    enforcement_on(env, &s) && resolve_action(env, &s) != Action::Skip
}

/// Runs the check; the return value is the process exit code.
pub fn run(inv: &PrePushInvocation, out: &mut dyn Write, err: &mut dyn Write) -> i32 {
    let mut warnings = Vec::new();
    let s = settings(&inv.project_dir, &mut warnings);
    for w in &warnings {
        let _ = writeln!(err, "{w}");
    }
    if !enforcement_on(&inv.env, &s) {
        return 0;
    }
    let action = resolve_action(&inv.env, &s);
    if action == Action::Skip {
        return 0;
    }
    let name = inv
        .env
        .convention
        .clone()
        .filter(|c| !c.is_empty())
        .unwrap_or_else(|| s.convention.clone());
    let loaded = if name == "auto" {
        resolve_auto(inv.repo_dir.as_deref().unwrap_or(&inv.project_dir), &s)
    } else {
        builtin(&name).ok_or_else(|| {
            format!("load convention \"{name}\": unknown built-in convention: \"{name}\"")
        })
    };
    let mut conv = match loaded {
        Ok(c) => c,
        Err(e) => {
            let _ = writeln!(err, "pre-push: load convention: {e}");
            return 1;
        }
    };
    if s.max_length > 0 {
        conv.max_length = s.max_length as usize;
    }
    let subjects: Vec<&[u8]> = inv
        .stdin
        .split(|b| *b == b'\n')
        .map(trim_subject)
        .filter(|l| !l.is_empty())
        .collect();
    if subjects.is_empty() {
        let _ = writeln!(out, "No commit messages to validate.");
        return 0;
    }
    let results: Vec<ValidationResult> = subjects.iter().map(|m| validate(m, &conv)).collect();
    let violations = results.iter().filter(|r| !is_valid(r)).count();
    if violations == 0 {
        let _ = writeln!(
            out,
            "All {} commit(s) follow {} convention.",
            subjects.len(),
            conv.name
        );
        return 0;
    }
    let _ = write!(
        err,
        "{violations} of {} commit(s) violate {} convention:\n\n",
        subjects.len(),
        conv.name
    );
    for r in results.iter().filter(|r| !is_valid(r)) {
        let _ = err.write_all(&format_error(r, &conv));
        let _ = writeln!(err);
    }
    if action == Action::Enforce {
        2
    } else {
        0
    }
}
