//! Pure segment computations ported from the Go statusline (moai-adk-go 4fa2aa9db2,
//! `internal/statusline`): no environment, no files, no clock. Each function names the Go
//! function it reproduces; integer arithmetic truncates the way Go `int` division does.

use super::input::SubagentTask;

/// Auto-compact trigger used to scale the context bar (memory.go `defaultAutoCompactPct`).
pub const DEFAULT_AUTO_COMPACT_PCT: i64 = 85;
const DEFAULT_CONTEXT_WINDOW: i64 = 200_000;
// internal/config handoff band constants.
const HANDOFF_SOFT_LARGE_PCT: f64 = 50.0;
const HANDOFF_SOFT_STANDARD_PCT: f64 = 90.0;
const HANDOFF_LARGE_WINDOW_CUTOFF: i64 = 500_000;
const HANDOFF_HARD_CEILING_CAP_PCT: i64 = 95;
const HANDOFF_HARD_CEILING_MARGIN_PCT: i64 = 10;

/// `ShortenModelName` (metrics.go): `claude-opus-4-1-20250805` → `Opus 4.1`.
pub fn shorten_model_name(model: &str) -> String {
    if model.is_empty() {
        return String::new();
    }
    let Some(rest) = model.strip_prefix("claude-") else {
        return model.strip_suffix("[1m]").unwrap_or(model).to_string();
    };
    let mut name = rest.to_string();
    let mut parts: Vec<String> = name.split('-').map(str::to_string).collect();
    if parts.len() > 1 {
        let last = &parts[parts.len() - 1];
        if last.len() == 8 && last.parse::<i64>().is_ok() {
            name = parts[..parts.len() - 1].join("-");
            parts = name.split('-').map(str::to_string).collect();
        }
    }
    let title = |p: &str| -> String {
        let mut chars = p.chars();
        match chars.next() {
            Some(first) => {
                first.to_uppercase().collect::<String>() + &chars.as_str().to_lowercase()
            }
            None => String::new(),
        }
    };
    let mut model_name = String::new();
    let mut version: Vec<String> = Vec::new();
    for (i, part) in parts.iter().enumerate() {
        let lower = part.to_lowercase();
        if lower == "sonnet" || lower == "opus" || lower == "haiku" {
            model_name = title(part);
            version = parts[..i].to_vec();
            version.extend_from_slice(&parts[i + 1..]);
            break;
        }
    }
    if model_name.is_empty() && !parts.is_empty() {
        model_name = title(&parts[0]);
        version = parts[1..].to_vec();
    }
    let version = version.join(".");
    if version.is_empty() {
        model_name
    } else {
        format!("{model_name} {version}")
    }
}

/// The context-window reading behind the CW bar (`MemoryData`).
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq)]
pub struct Memory {
    pub tokens_used: i64,
    pub token_budget: i64,
    pub context_window_size: i64,
    pub available: bool,
}

/// What `CollectMemory` reads from `context_window`; `None` when the object is absent.
#[derive(Clone, Copy, Debug, Default)]
pub struct ContextReading {
    pub used_percentage: Option<f64>,
    pub context_window_size: i64,
    pub used: i64,
    pub total: i64,
    /// input + cache creation + cache read, when `current_usage` is present.
    pub current_usage_tokens: Option<i64>,
}

/// `CollectMemory` (memory.go). `size_override` is the GLM / env override (0 = none);
/// `auto_compact` scales the budget so the bar reaches 100% where auto-compact fires.
pub fn collect_memory(cw: Option<ContextReading>, size_override: i64, auto_compact: i64) -> Memory {
    let Some(cw) = cw else {
        return Memory::default();
    };
    let mut size = cw.context_window_size;
    if size <= 0 {
        size = cw.total;
    }
    if size <= 0 {
        size = DEFAULT_CONTEXT_WINDOW;
    }
    if size_override > 0 {
        size = size_override;
    }
    let budget = size * auto_compact / 100;
    let full = |tokens_used| Memory {
        tokens_used,
        token_budget: budget,
        context_window_size: size,
        available: true,
    };
    if let Some(pct) = cw.used_percentage {
        return full((size as f64 * pct / 100.0) as i64);
    }
    if let Some(tokens) = cw.current_usage_tokens {
        return full(tokens);
    }
    if cw.used > 0 || cw.total > 0 {
        return Memory {
            tokens_used: cw.used,
            token_budget: cw.total * auto_compact / 100,
            context_window_size: size,
            available: true,
        };
    }
    full(0)
}

/// `usagePercent`: truncated, capped at 100, 0 for an empty budget.
pub fn usage_percent(used: i64, total: i64) -> i64 {
    if total <= 0 {
        return 0;
    }
    (used * 100 / total).min(100)
}

/// `cacheHitPercent`: read / (read + creation), omitted when creation ≤ 0.
pub fn cache_hit_percent(read: i64, creation: i64) -> Option<i64> {
    if creation <= 0 {
        return None;
    }
    let denom = read + creation;
    if denom <= 0 {
        return None;
    }
    Some(read * 100 / denom)
}

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum HandoffStage {
    None,
    Soft,
    Hard,
}

/// `handoffGuideStage` with `hardCeilingPct` / `softThresholdPct` (renderer.go).
pub fn handoff_stage(memory: &Memory, auto_compact: i64) -> HandoffStage {
    let size = memory.context_window_size;
    if size <= 0 {
        return HandoffStage::None;
    }
    let raw = memory.tokens_used as f64 * 100.0 / size as f64;
    let soft = if size >= HANDOFF_LARGE_WINDOW_CUTOFF {
        HANDOFF_SOFT_LARGE_PCT
    } else {
        HANDOFF_SOFT_STANDARD_PCT
    };
    let ceil = (auto_compact + HANDOFF_HARD_CEILING_MARGIN_PCT).min(HANDOFF_HARD_CEILING_CAP_PCT);
    let hard = if (ceil as f64) < soft {
        soft
    } else {
        ceil as f64
    };
    if raw >= hard {
        HandoffStage::Hard
    } else if raw >= soft {
        HandoffStage::Soft
    } else {
        HandoffStage::None
    }
}

/// `getAutoCompactThreshold`: an override in 1..=100, else the default.
pub fn auto_compact_threshold(override_value: Option<&str>) -> i64 {
    override_value
        .and_then(|v| v.parse::<i64>().ok())
        .filter(|v| *v > 0 && *v <= 100)
        .unwrap_or(DEFAULT_AUTO_COMPACT_PCT)
}

/// `renderSessionTime`: `⏳ Xm`, `⏳ Xh Ym`, `⏳ Xd Yh`; empty for 0.
pub fn session_time(ms: i64) -> String {
    if ms <= 0 {
        return String::new();
    }
    let minutes = ms / 60_000;
    let hours = minutes / 60;
    if hours >= 24 {
        format!("⏳ {}d {}h", hours / 24, hours % 24)
    } else if hours >= 1 {
        format!("⏳ {}h {}m", hours, minutes % 60)
    } else {
        format!("⏳ {minutes}m")
    }
}

/// `formatResetTimeRelative`: `rolling` for a zero or past reset. `now_ms` carries the
/// sub-second part Go's `time.Until` sees, so a reset exactly one hour ahead in whole
/// seconds reads `59m`, as in Go.
pub fn reset_relative(resets_at: i64, now_ms: i64) -> String {
    if resets_at <= 0 {
        return "rolling".into();
    }
    let remaining_ms = resets_at * 1000 - now_ms;
    if remaining_ms <= 0 {
        return "rolling".into();
    }
    let hours = remaining_ms / 3_600_000;
    let minutes = (remaining_ms / 60_000) % 60;
    if hours >= 24 {
        let mut parts = vec![format!("{}D", hours / 24)];
        if hours % 24 > 0 {
            parts.push(format!("{}h", hours % 24));
        }
        if minutes > 0 {
            parts.push(format!("{minutes}m"));
        }
        return parts.join(", ");
    }
    if hours > 0 {
        if minutes > 0 {
            return format!("{hours}h {minutes}m");
        }
        return format!("{hours}h");
    }
    format!("{minutes}m")
}

/// `formatResetTimeAbsolute` as `Jan 2`, in UTC (Go uses the local zone; the harness
/// pins TZ=UTC, and this binary carries no time-zone database).
pub fn reset_absolute(resets_at: i64) -> String {
    if resets_at <= 0 {
        return "rolling".into();
    }
    let (_, month, day) = civil_from_days(resets_at.div_euclid(86_400));
    const MONTHS: [&str; 12] = [
        "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
    ];
    format!("{} {day}", MONTHS[(month - 1) as usize])
}

/// Days since 1970-01-01 → (year, month 1..=12, day 1..=31) (Howard Hinnant's algorithm).
fn civil_from_days(days: i64) -> (i64, i64, i64) {
    let z = days + 719_468;
    let era = z.div_euclid(146_097);
    let doe = z - era * 146_097;
    let yoe = (doe - doe / 1460 + doe / 36_524 - doe / 146_096) / 365;
    let doy = doe - (365 * yoe + yoe / 4 - yoe / 100);
    let mp = (5 * doy + 2) / 153;
    let day = doy - (153 * mp + 2) / 5 + 1;
    let month = if mp < 10 { mp + 3 } else { mp - 9 };
    let year = yoe + era * 400 + i64::from(month <= 2);
    (year, month, day)
}

/// What `git --no-optional-locks status --porcelain --branch` reports (manager.go).
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct GitStatus {
    pub branch: String,
    pub staged: i64,
    pub modified: i64,
    pub untracked: i64,
    pub ahead: i64,
    pub behind: i64,
}

/// `Status` line parsing plus `parseStatusBranchHeader` (manager.go).
pub fn parse_porcelain(out: &str) -> GitStatus {
    let mut status = GitStatus::default();
    for line in out.split('\n') {
        if let Some(header) = line.strip_prefix("## ") {
            let (branch, ahead, behind) = parse_branch_header(header);
            status.branch = branch;
            status.ahead = ahead;
            status.behind = behind;
            continue;
        }
        let bytes = line.as_bytes();
        if bytes.len() < 3 {
            continue;
        }
        let (x, y) = (bytes[0], bytes[1]);
        if x == b'?' && y == b'?' {
            status.untracked += 1;
            continue;
        }
        if x != b' ' && x != b'?' {
            status.staged += 1;
        }
        if y == b'M' || y == b'D' {
            status.modified += 1;
        }
    }
    status
}

fn parse_branch_header(header: &str) -> (String, i64, i64) {
    let mut rest = header.strip_prefix("No commits yet on ").unwrap_or(header);
    let (mut ahead, mut behind) = (0, 0);
    if let Some(idx) = rest.rfind(" [") {
        if rest.ends_with(']') {
            for part in rest[idx + 2..rest.len() - 1].split(", ") {
                let Some((field, value)) = part.split_once(' ') else {
                    continue;
                };
                let Ok(n) = value.parse::<i64>() else {
                    continue;
                };
                match field {
                    "ahead" => ahead = n,
                    "behind" => behind = n,
                    _ => {}
                }
            }
            rest = &rest[..idx];
        }
    }
    if let Some(idx) = rest.find("...") {
        rest = &rest[..idx];
    }
    if rest.contains([' ', '\t']) {
        return (String::new(), ahead, behind);
    }
    (rest.to_string(), ahead, behind)
}

/// `renderSubagentOutput`: one `{"id","content"}` line per row with a non-empty id.
pub fn subagent_rows(tasks: &[SubagentTask]) -> String {
    let mut out = String::new();
    for task in tasks {
        let id = task.id.as_deref().unwrap_or("");
        if id.is_empty() {
            continue;
        }
        let row = serde_json::json!({"id": id, "content": subagent_row(task)});
        out.push_str(&go_json(&row));
        out.push('\n');
    }
    out
}

/// Go `json.Marshal` escapes `<`, `>` and `&` (and U+2028/U+2029); serde_json does not.
fn go_json(v: &serde_json::Value) -> String {
    let plain = v.to_string();
    let mut out = String::with_capacity(plain.len());
    for c in plain.chars() {
        match c {
            '<' => out.push_str("\\u003c"),
            '>' => out.push_str("\\u003e"),
            '&' => out.push_str("\\u0026"),
            '\u{2028}' => out.push_str("\\u2028"),
            '\u{2029}' => out.push_str("\\u2029"),
            c => out.push(c),
        }
    }
    out
}

/// `formatSubagentTaskRow`: `⚙ [agentType] name (status) · N tokens`.
fn subagent_row(task: &SubagentTask) -> String {
    let mut row = String::from("⚙ ");
    let badge = sanitize(task.agent_type.as_deref().unwrap_or(""));
    if !badge.is_empty() {
        row.push_str(&format!("[{badge}] "));
    }
    row.push_str(&subagent_display_name(task));
    let status = sanitize(task.status.as_deref().unwrap_or(""));
    if !status.is_empty() {
        row.push_str(&format!(" ({status})"));
    }
    let tokens = task.token_count.unwrap_or(0);
    if tokens > 0 {
        row.push_str(&format!(" · {} tokens", token_count(tokens)));
    }
    row
}

fn subagent_display_name(task: &SubagentTask) -> String {
    for candidate in [&task.name, &task.label, &task.description] {
        let clean = sanitize(candidate.as_deref().unwrap_or(""));
        if !clean.is_empty() {
            return clean;
        }
    }
    format!("#{}", task.id.as_deref().unwrap_or(""))
}

/// `sanitizeSubagentText`: drop C0, DEL and C1 controls.
fn sanitize(s: &str) -> String {
    s.chars()
        .filter(|c| !(*c < '\u{20}' || *c == '\u{7f}' || ('\u{80}'..='\u{9f}').contains(c)))
        .collect()
}

/// `formatTokenCount`: 850 → `850`, 12_300 → `12.3k`, 1_234_567 → `1.2M`.
fn token_count(n: i64) -> String {
    if n >= 1_000_000 {
        format!("{:.1}M", n as f64 / 1_000_000.0)
    } else if n >= 1_000 {
        format!("{:.1}k", n as f64 / 1_000.0)
    } else {
        n.to_string()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn model_names_shorten_like_go() {
        assert_eq!(shorten_model_name("claude-opus-4-1-20250805"), "Opus 4.1");
        assert_eq!(shorten_model_name("claude-sonnet-5-5"), "Sonnet 5.5");
        assert_eq!(shorten_model_name("claude-3-5-haiku-20241022"), "Haiku 3.5");
        assert_eq!(shorten_model_name("glm-5.3[1m]"), "glm-5.3");
        assert_eq!(shorten_model_name("claude-fable-5-1"), "Fable 5.1");
    }

    #[test]
    fn memory_scales_to_the_auto_compact_budget() {
        let cw = |pct, size| ContextReading {
            used_percentage: Some(pct),
            context_window_size: size,
            ..Default::default()
        };
        let m = collect_memory(Some(cw(8.0, 200_000)), 0, 85);
        assert_eq!((m.tokens_used, m.token_budget), (16_000, 170_000));
        assert_eq!(usage_percent(m.tokens_used, m.token_budget), 9);
        let m = collect_memory(Some(cw(93.0, 200_000)), 0, 85);
        assert_eq!(usage_percent(m.tokens_used, m.token_budget), 100);
        assert_eq!(handoff_stage(&m, 85), HandoffStage::Soft);
        let usage = ContextReading {
            context_window_size: 200_000,
            current_usage_tokens: Some(50_000),
            ..Default::default()
        };
        let m = collect_memory(Some(usage), 0, 85);
        assert_eq!(usage_percent(m.tokens_used, m.token_budget), 29);
        assert!(!collect_memory(None, 0, 85).available);
    }

    #[test]
    fn handoff_stage_bands() {
        let m = |used, size| Memory {
            tokens_used: used,
            token_budget: 1,
            context_window_size: size,
            available: true,
        };
        assert_eq!(handoff_stage(&m(179_999, 200_000), 85), HandoffStage::None);
        assert_eq!(handoff_stage(&m(180_000, 200_000), 85), HandoffStage::Soft);
        assert_eq!(handoff_stage(&m(190_000, 200_000), 85), HandoffStage::Hard);
        assert_eq!(
            handoff_stage(&m(500_000, 1_000_000), 85),
            HandoffStage::Soft
        );
        // A degenerate override collapses the hard ceiling onto the soft band.
        assert_eq!(handoff_stage(&m(180_000, 200_000), 20), HandoffStage::Hard);
    }

    #[test]
    fn auto_compact_override_is_bounded() {
        assert_eq!(auto_compact_threshold(None), 85);
        assert_eq!(auto_compact_threshold(Some("70")), 70);
        assert_eq!(auto_compact_threshold(Some("100")), 100);
        assert_eq!(auto_compact_threshold(Some("101")), 85);
        assert_eq!(auto_compact_threshold(Some("0")), 85);
        assert_eq!(auto_compact_threshold(Some("x")), 85);
    }

    #[test]
    fn cache_hit_needs_a_fresh_write() {
        assert_eq!(cache_hit_percent(2000, 5000), Some(28));
        assert_eq!(cache_hit_percent(10_000, 0), None);
    }

    #[test]
    fn times_format_like_go() {
        assert_eq!(session_time(45_000), "⏳ 0m");
        assert_eq!(session_time(5_400_000), "⏳ 1h 30m");
        assert_eq!(session_time(90_000_000), "⏳ 1d 1h");
        assert_eq!(session_time(0), "");
        let now = 200_000; // 200 s in ms
        assert_eq!(reset_relative(100, now), "rolling");
        assert_eq!(reset_relative(200, now), "rolling");
        assert_eq!(reset_relative(200 + 4 * 3600 + 30 * 60, now), "4h 30m");
        assert_eq!(reset_relative(200 + 3600, now), "1h");
        assert_eq!(reset_relative(200 + 3600, now + 1), "59m");
        assert_eq!(reset_relative(200 + 26 * 3600 + 60, now), "1D, 2h, 1m");
        assert_eq!(reset_relative(200 + 59, now), "0m");
        assert_eq!(reset_absolute(1_738_857_600), "Feb 6");
        assert_eq!(reset_absolute(0), "rolling");
        assert_eq!(reset_absolute(951_782_400), "Feb 29");
    }

    #[test]
    fn porcelain_counts_and_header() {
        let s = parse_porcelain(
            "## main...origin/main [ahead 1, behind 2]\n M a\n?? b\nA  c\nMM d\nR  x -> y\n",
        );
        assert_eq!(s.branch, "main");
        assert_eq!((s.ahead, s.behind), (1, 2));
        assert_eq!((s.staged, s.modified, s.untracked), (3, 2, 1));
        assert_eq!(
            parse_porcelain("## No commits yet on main\n").branch,
            "main"
        );
        assert_eq!(parse_porcelain("## HEAD (no branch)\n").branch, "");
        assert_eq!(
            parse_porcelain("## feat/x...origin/feat/x\n").branch,
            "feat/x"
        );
    }

    #[test]
    fn subagent_rows_skip_empty_ids_and_sanitize() {
        let task = |id: &str, name: Option<&str>| SubagentTask {
            id: Some(id.into()),
            name: name.map(Into::into),
            agent_type: Some("Explore".into()),
            status: Some("running".into()),
            token_count: Some(12_000),
            ..Default::default()
        };
        let out = subagent_rows(&[
            task("t1", Some("ex\u{1b}plorer")),
            task("", Some("x")),
            task("t3", None),
        ]);
        assert_eq!(
            out,
            "{\"id\":\"t1\",\"content\":\"⚙ [Explore] explorer (running) · 12.0k tokens\"}\n\
             {\"id\":\"t3\",\"content\":\"⚙ [Explore] #t3 (running) · 12.0k tokens\"}\n"
        );
        assert_eq!(subagent_rows(&[]), "");
        assert_eq!(token_count(850), "850");
        // Go %.1f rounds the exact binary value half to even: 1.25 → "1.2".
        assert_eq!(token_count(1_250), "1.2k");
        assert_eq!(token_count(1_234_567), "1.2M");
        assert_eq!(
            go_json(&serde_json::json!("<a&b>")),
            "\"\\u003ca\\u0026b\\u003e\""
        );
    }
}
