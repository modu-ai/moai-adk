//! `moai update` for a person at a terminal (Issue #2894, leader decision D3-U on #2799 D3).
//!
//! When stdout is a terminal and `--json` is not given, `moai update` shows these screens
//! instead of the JSON envelope; a pipe, a host tool, a skill or a lane still gets the envelope
//! byte for byte. The screens are drawn from the same `update.run` result the envelope carries,
//! so the two can never say different things. Layout follows the Mole survey (#2894
//! 6873522840): a `◆ MoAI update` head with a muted context line, a blank line and a bold
//! heading per stage, two-space indented rows, a symbol *and* a word on every row (never color
//! alone), counts after a `·`, one "all up to date" screen, `[y/N]` with no as the default.
//!
//! Every text taken from the result goes through [`crate::ui::visible`] (#2818 B1), so a file
//! or reason cannot move the cursor or retitle the window. Lines are cut to the terminal width
//! by display cells (Korean takes two).

use crate::ui::{visible, Painter, Token};
use crate::ui_text::Lang;
use serde_json::Value;
use std::sync::atomic::{AtomicBool, AtomicU8, Ordering};
use unicode_width::{UnicodeWidthChar, UnicodeWidthStr};

/// What happened to the question on this run.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Asked {
    /// Not asked: nothing to write, or no terminal to answer on.
    NotAsked,
    /// Answered yes; the result is the applied one.
    Yes,
    /// Answered no (or Enter, end of input): nothing was written.
    No,
    /// Answered yes, but the plan changed before the write: refused, and the result is a new
    /// preview.
    Stale,
}

/// The MoAI project the command runs in, for its line (U item 7, 7b). Read only.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Project {
    pub path: String,
    /// `Some(n)`: the project still has the Go-generation `.moai/`, with `n` entries the move
    /// would carry (`go-state-plan`); `None`: a project of this generation, whose settings the
    /// session's `/moai update` checks (`moai update` does not plan project files).
    pub go_items: Option<usize>,
}

impl Project {
    /// The folder name shown in the end block.
    fn name(&self) -> &str {
        self.path
            .trim_end_matches('/')
            .rsplit('/')
            .next()
            .unwrap_or(&self.path)
    }
}

/// How to draw: the painter (color, symbols, language), the width in cells, the HOME shown as
/// `~`, and the project line.
#[derive(Clone, Debug)]
pub struct View {
    pub painter: Painter,
    pub width: usize,
    pub home: String,
    pub project: Option<Project>,
}

// --- the process-wide switch -------------------------------------------------------------

static HUMAN: AtomicBool = AtomicBool::new(false);
static ASKED: AtomicU8 = AtomicU8::new(0);

/// Set once by `main` before the command runs: draw screens instead of the envelope.
pub fn set_human(on: bool) {
    HUMAN.store(on, Ordering::SeqCst);
}

pub fn human() -> bool {
    HUMAN.load(Ordering::SeqCst)
}

/// Recorded by the question; read back when the result is drawn.
pub fn set_asked(a: Asked) {
    ASKED.store(
        match a {
            Asked::NotAsked => 0,
            Asked::Yes => 1,
            Asked::No => 2,
            Asked::Stale => 3,
        },
        Ordering::SeqCst,
    );
}

pub fn asked() -> Asked {
    match ASKED.load(Ordering::SeqCst) {
        1 => Asked::Yes,
        2 => Asked::No,
        3 => Asked::Stale,
        _ => Asked::NotAsked,
    }
}

// --- words ---------------------------------------------------------------------------------

#[derive(Clone, Copy)]
enum W {
    Installed,
    Package,
    NotInstalled,
    ProjectWord,
    Checking,
    Changes,
    Applying,
    Skills,
    Agents,
    Routing,
    Style,
    BinNames,
    Hooks,
    UpToDate,
    ToUpdate,
    New,
    Conflict,
    Skipped,
    Blocked,
    OutsideLayout,
    NoHookSet,
    NotGitTree,
    AllUpToDate,
    NothingApplied,
    NotWritten,
    NextConfirm,
    StaleRefused,
    Done,
    Partial,
    Revert,
    Reload,
    SameAs,
    More,
    GoPending,
    ProjectNext,
    ThisFolder,
    GoWaiting,
    SessionChecks,
    Changed,
    Linked,
    Cleanup,
    Archive,
    Keep,
    CleanupNext,
    Cleaned,
}

fn w(lang: Lang, word: W) -> &'static str {
    use W::*;
    let ko = lang == Lang::Ko;
    match word {
        Installed => {
            if ko {
                "설치본"
            } else {
                "installed"
            }
        }
        Package => {
            if ko {
                "패키지"
            } else {
                "package"
            }
        }
        NotInstalled => {
            if ko {
                "없음"
            } else {
                "none"
            }
        }
        ProjectWord => {
            if ko {
                "프로젝트"
            } else {
                "project"
            }
        }
        Checking => {
            if ko {
                "검사"
            } else {
                "Checking"
            }
        }
        Changes => {
            if ko {
                "바뀔 것"
            } else {
                "Changes"
            }
        }
        Applying => {
            if ko {
                "적용"
            } else {
                "Applied"
            }
        }
        Skills => {
            if ko {
                "스킬"
            } else {
                "skills"
            }
        }
        Agents => {
            if ko {
                "에이전트"
            } else {
                "agents"
            }
        }
        Routing => {
            if ko {
                "라우팅"
            } else {
                "routing"
            }
        }
        Style => {
            if ko {
                "출력 스타일"
            } else {
                "output style"
            }
        }
        BinNames => {
            if ko {
                "실행 파일 이름"
            } else {
                "command names"
            }
        }
        Hooks => "hooks",
        UpToDate => {
            if ko {
                "최신"
            } else {
                "up to date"
            }
        }
        ToUpdate => {
            if ko {
                "갱신"
            } else {
                "to update"
            }
        }
        New => {
            if ko {
                "새로"
            } else {
                "new"
            }
        }
        Conflict => {
            if ko {
                "충돌"
            } else {
                "conflict"
            }
        }
        Skipped => {
            if ko {
                "건너뜀"
            } else {
                "skipped"
            }
        }
        Blocked => {
            if ko {
                "막힘"
            } else {
                "blocked"
            }
        }
        OutsideLayout => {
            if ko {
                "버전 폴더 밖에서 실행 중"
            } else {
                "running outside the version folders"
            }
        }
        NoHookSet => {
            if ko {
                "이 프로젝트에 MoAI hook 없음"
            } else {
                "no MoAI hook set in this project"
            }
        }
        NotGitTree => {
            if ko {
                "git 작업 트리 밖이에요"
            } else {
                "not inside a git work tree"
            }
        }
        AllUpToDate => {
            if ko {
                "모두 최신이에요"
            } else {
                "Everything is up to date"
            }
        }
        NothingApplied => {
            if ko {
                "적용할 수 있는 것이 없어요 — 막힌 곳을 먼저 풀어 주세요"
            } else {
                "Nothing can be applied — clear the blocked part first"
            }
        }
        NotWritten => {
            if ko {
                "쓰지 않았어요."
            } else {
                "Nothing was written."
            }
        }
        NextConfirm => {
            if ko {
                "다음: 터미널에서 다시 실행해 답하거나"
            } else {
                "next: run it again in a terminal to answer, or"
            }
        }
        StaleRefused => {
            if ko {
                "확인 사이에 계획이 바뀌어 쓰지 않았어요 — 새 미리보기예요"
            } else {
                "The plan changed after the preview; nothing was written — here is the new one"
            }
        }
        Done => {
            if ko {
                "다 바꿨어요"
            } else {
                "Done"
            }
        }
        Partial => {
            if ko {
                "일부만 바꿨어요"
            } else {
                "Partly applied"
            }
        }
        Revert => {
            if ko {
                "되돌리기"
            } else {
                "revert"
            }
        }
        Reload => {
            if ko {
                "실행 중인 세션은 /reload 가 필요해요"
            } else {
                "running sessions need /reload"
            }
        }
        SameAs => {
            if ko {
                "같은 파일"
            } else {
                "the same files"
            }
        }
        More => {
            if ko {
                "외"
            } else {
                "more"
            }
        }
        GoPending => {
            if ko {
                "아직 이전 전(Go 세대) — 옮길 항목"
            } else {
                "not moved yet (Go generation) — items to move:"
            }
        }
        ProjectNext => {
            if ko {
                // A no-break space keeps each command on one line when the sentence wraps.
                "다음: 기존 프로젝트는 각 프로젝트의 세션에서 /moai\u{a0}update (Codex CLI 는 $moai\u{a0}update)를 실행해 프로젝트 설정을 꼭 업데이트하세요."
            } else {
                "Next: in each existing project, run /moai\u{a0}update in a session there ($moai\u{a0}update in Codex CLI) to update the project's settings."
            }
        }
        ThisFolder => {
            if ko {
                "이 폴더"
            } else {
                "this folder"
            }
        }
        GoWaiting => {
            if ko {
                "Go 세대 · 이전 대기"
            } else {
                "Go generation · waiting to move:"
            }
        }
        SessionChecks => {
            if ko {
                "MoAI 프로젝트 · 설정 확인은 세션의 /moai update 가 해요"
            } else {
                "MoAI project · the session's /moai update checks its settings"
            }
        }
        Changed => {
            if ko {
                "바뀜"
            } else {
                "changed"
            }
        }
        Linked => {
            if ko {
                "연결"
            } else {
                "link"
            }
        }
        Cleanup => {
            if ko {
                "정리(Go 남은 것)"
            } else {
                "cleanup (Go leftovers)"
            }
        }
        Archive => {
            if ko {
                "보관"
            } else {
                "to archive"
            }
        }
        Keep => {
            if ko {
                "유지"
            } else {
                "kept"
            }
        }
        CleanupNext => {
            if ko {
                "정리는 따로 확인해요:"
            } else {
                "the cleanup is confirmed on its own:"
            }
        }
        Cleaned => {
            if ko {
                "정리했어요 — 보관"
            } else {
                "Cleaned up — archived:"
            }
        }
    }
}

/// `n` with its unit word: `3개` / `3`.
fn count(lang: Lang, n: usize) -> String {
    if lang == Lang::Ko {
        format!("{n}개")
    } else {
        n.to_string()
    }
}

/// `word n` in the language's order: `최신 28개`, `28 up to date`.
fn counted(lang: Lang, word: W, n: usize) -> String {
    if lang == Lang::Ko {
        format!("{} {}", w(lang, word), count(lang, n))
    } else {
        format!("{n} {}", w(lang, word))
    }
}

// --- rows ----------------------------------------------------------------------------------

/// The state a row shows: a symbol and a word, never color alone.
#[derive(Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Debug)]
enum Mark {
    Ok,
    Skip,
    Update,
    Warn,
    Error,
}

fn mark(p: &Painter, m: Mark) -> String {
    match m {
        Mark::Ok => p.mark(Token::Ok),
        Mark::Warn => p.mark(Token::Warn),
        Mark::Error => p.mark(Token::Error),
        Mark::Update => p.paint(Token::Info, if p.unicode { "↑" } else { "^" }),
        Mark::Skip => p.paint(Token::Muted, if p.unicode { "–" } else { "-" }),
    }
}

struct Row {
    mark: Mark,
    label: String,
    text: String,
}

fn row(mark: Mark, label: String, text: String) -> Row {
    Row { mark, label, text }
}

/// Counts of a file list's `action`s: (unchanged, create, update, conflict, other blocked).
fn tally(files: &[Value]) -> (usize, usize, usize, usize, usize) {
    let mut t = (0, 0, 0, 0, 0);
    for f in files {
        match f["action"].as_str().unwrap_or("") {
            "unchanged" => t.0 += 1,
            "create" | "append" | "merge" | "set" => t.1 += 1,
            "update" | "replace" | "replace_copy" => t.2 += 1,
            "conflict" => t.3 += 1,
            _ => t.4 += 1,
        }
    }
    t
}

fn tally_text(lang: Lang, t: (usize, usize, usize, usize, usize)) -> (Mark, String) {
    let mut parts = Vec::new();
    if t.0 > 0 {
        parts.push(counted(lang, W::UpToDate, t.0));
    }
    if t.1 > 0 {
        parts.push(counted(lang, W::New, t.1));
    }
    if t.2 > 0 {
        parts.push(counted(lang, W::ToUpdate, t.2));
    }
    if t.3 > 0 {
        parts.push(counted(lang, W::Conflict, t.3));
    }
    if t.4 > 0 {
        parts.push(counted(lang, W::Skipped, t.4));
    }
    let mark = if t.3 > 0 || t.4 > 0 {
        Mark::Warn
    } else if t.1 + t.2 > 0 {
        Mark::Update
    } else {
        Mark::Ok
    };
    (mark, parts.join(" · "))
}

fn blocked_row(lang: Lang, label: String, reason: &Value) -> Row {
    let reason = reason.as_str().unwrap_or("");
    row(
        Mark::Error,
        label,
        format!("{}: {}", w(lang, W::Blocked), reason),
    )
}

/// The `검사` rows of a preview, one per layer (skills per host).
fn preview_rows(v: &Value, lang: Lang) -> Vec<Row> {
    let mut rows = Vec::new();
    for h in v["plan"]["hosts"].as_array().into_iter().flatten() {
        let label = format!(
            "{} {}",
            w(lang, W::Skills),
            h["host"].as_str().unwrap_or("")
        );
        if h["preflight"].as_str() != Some("ok") {
            rows.push(blocked_row(lang, label, &h["blocked_reason"]));
            continue;
        }
        let (m, text) = tally_text(lang, tally(h["files"].as_array().map_or(&[], |a| a)));
        rows.push(row(m, label, text));
    }
    // Agents: one row over the hosts.
    let hosts: Vec<&Value> = v["agents"]["hosts"]
        .as_array()
        .into_iter()
        .flatten()
        .collect();
    if !hosts.is_empty() {
        let mut all = (0, 0, 0, 0, 0);
        let mut names = Vec::new();
        let mut blocked = Vec::new();
        for h in &hosts {
            let name = h["host"].as_str().unwrap_or("");
            if h["preflight"].as_str() != Some("ok") {
                blocked.push(format!(
                    "{name} {}: {}",
                    w(lang, W::Blocked),
                    h["blocked_reason"].as_str().unwrap_or("")
                ));
                continue;
            }
            let files = h["files"].as_array().map_or(&[][..], |a| a);
            let t = tally(files);
            all = (
                all.0 + t.0,
                all.1 + t.1,
                all.2 + t.2,
                all.3 + t.3,
                all.4 + t.4,
            );
            names.push(format!("{name} {}", files.len()));
        }
        let (mut m, text) = tally_text(lang, all);
        let mut text = format!("{} — {text}", names.join(" · "));
        if !blocked.is_empty() {
            m = Mark::Error;
            text = format!("{text} · {}", blocked.join(" · "));
        }
        rows.push(row(m, w(lang, W::Agents).into(), text));
    }
    // Routing: the files and their state.
    let mut parts = Vec::new();
    let mut worst = Mark::Ok;
    for h in v["routing"]["hosts"].as_array().into_iter().flatten() {
        let file = short(h["file"].as_str().unwrap_or(""));
        if h["preflight"].as_str() != Some("ok") {
            worst = worst.max(Mark::Error);
            parts.push(format!(
                "{file} {}: {}",
                w(lang, W::Blocked),
                h["blocked_reason"].as_str().unwrap_or("")
            ));
            continue;
        }
        let (m, word) = match h["section"]["action"].as_str().unwrap_or("") {
            "unchanged" => (Mark::Ok, w(lang, W::UpToDate)),
            "create" => (Mark::Update, w(lang, W::New)),
            "update" => (Mark::Update, w(lang, W::ToUpdate)),
            "conflict" => (Mark::Warn, w(lang, W::Conflict)),
            _ => (Mark::Warn, w(lang, W::Skipped)),
        };
        worst = worst.max(m);
        parts.push(format!("{file} {word}"));
    }
    if !parts.is_empty() {
        rows.push(row(worst, w(lang, W::Routing).into(), parts.join(" · ")));
    }
    // Output style.
    if let Some(targets) = v["style"]["targets"].as_array().filter(|t| !t.is_empty()) {
        let (m, text) = tally_text(lang, tally(targets));
        rows.push(row(m, w(lang, W::Style).into(), text));
    }
    // Command names.
    let b = &v["bin_names"];
    if let Some(state) = b["state"].as_str() {
        let label = w(lang, W::BinNames).to_string();
        rows.push(match state {
            "outside_layout" => row(
                Mark::Skip,
                label,
                format!("{}: {}", w(lang, W::Skipped), w(lang, W::OutsideLayout)),
            ),
            "unchanged" => row(Mark::Ok, label, w(lang, W::UpToDate).into()),
            "create" | "replace_copy" => {
                row(Mark::Update, label, format!("moai {}", w(lang, W::Linked)))
            }
            _ => blocked_row(lang, label, &b["reason"]),
        });
    }
    // Hooks.
    let hk = &v["hooks"];
    let hook_hosts = hk["hosts"].as_array().map_or(&[][..], |a| a);
    if hook_hosts.is_empty() {
        if let Some(s) = hk["skipped"].as_array().and_then(|s| s.first()) {
            let reason = s["reason"].as_str().unwrap_or("");
            // The two reasons `update_hooks` gives most often, in the screen's language; any
            // other reason is shown as the result has it.
            let reason = match reason {
                "no MoAI hook set is installed in this project" => {
                    w(lang, W::NoHookSet).to_string()
                }
                "not inside a git work tree" => w(lang, W::NotGitTree).to_string(),
                other => other.to_string(),
            };
            rows.push(row(
                Mark::Skip,
                w(lang, W::Hooks).into(),
                format!("{}: {reason}", w(lang, W::Skipped)),
            ));
        }
    } else {
        let (m, text) = tally_text(lang, tally(hook_hosts));
        rows.push(row(m, w(lang, W::Hooks).into(), text));
    }
    for s in v["skipped"].as_array().into_iter().flatten() {
        rows.push(row(
            Mark::Skip,
            format!(
                "{} {}",
                s["part"].as_str().unwrap_or(""),
                s["host"].as_str().unwrap_or("")
            )
            .trim()
            .to_string(),
            format!(
                "{}: {}",
                w(lang, W::Skipped),
                s["reason"].as_str().unwrap_or("")
            ),
        ));
    }
    // #2872 stage 2: Go leftovers the cleanup would archive (confirmed apart, never by the y).
    let c = &v["cleanup"];
    if c["confirm_token"].is_string() && v["cleanup_result"].is_null() {
        let n = |k: &str| c["counts"][k].as_u64().unwrap_or(0) as usize;
        rows.push(row(
            Mark::Warn,
            w(lang, W::Cleanup).into(),
            format!(
                "{} · {}",
                counted(lang, W::Archive, n("archive")),
                counted(lang, W::Keep, n("keep"))
            ),
        ));
    }
    rows
}

/// A path with the user's folder as `~` is shown by the caller; here only the file name.
fn short(path: &str) -> String {
    path.rsplit('/').next().unwrap_or(path).to_string()
}

// --- drawing -------------------------------------------------------------------------------

fn cells(s: &str) -> usize {
    UnicodeWidthStr::width(s)
}

/// `s` cut to `max` cells with `…` (or `...`) inside the limit.
fn cut(s: &str, max: usize, unicode: bool) -> String {
    if cells(s) <= max {
        return s.to_string();
    }
    let dots = if unicode { "…" } else { "..." };
    let room = max.saturating_sub(cells(dots));
    let mut out = String::new();
    let mut used = 0;
    for c in s.chars() {
        let cw = c.width().unwrap_or(0);
        if used + cw > room {
            break;
        }
        used += cw;
        out.push(c);
    }
    out.push_str(dots);
    out
}

/// Lines of `text` (already safe) wrapped to `width` cells after `indent` spaces; the first
/// line starts at `first` cells.
fn wrap(text: &str, first: usize, indent: usize, width: usize) -> Vec<String> {
    let mut lines = Vec::new();
    let mut cur = String::new();
    let mut room = width.saturating_sub(first).max(8);
    for word in text.split(' ') {
        let need = if cur.is_empty() {
            cells(word)
        } else {
            cells(&cur) + 1 + cells(word)
        };
        if need > room && !cur.is_empty() {
            lines.push(std::mem::take(&mut cur));
            room = width.saturating_sub(indent).max(8);
        }
        if !cur.is_empty() {
            cur.push(' ');
        }
        cur.push_str(word);
    }
    if !cur.is_empty() {
        lines.push(cur);
    }
    // A single word longer than the room is cut.
    lines
        .into_iter()
        .enumerate()
        .map(|(i, l)| {
            let r = if i == 0 {
                width.saturating_sub(first)
            } else {
                width.saturating_sub(indent)
            };
            cut(&l, r.max(4), true)
        })
        .collect()
}

struct Out<'a> {
    v: &'a View,
    buf: String,
}

impl Out<'_> {
    fn p(&self) -> &Painter {
        &self.v.painter
    }
    fn lang(&self) -> Lang {
        self.v.painter.lang
    }
    fn home(&self, s: &str) -> String {
        let h = self.v.home.trim_end_matches('/');
        if !h.is_empty() && (s == h || s.starts_with(&format!("{h}/"))) {
            format!("~{}", &s[h.len()..])
        } else {
            s.to_string()
        }
    }
    fn blank(&mut self) {
        self.buf.push('\n');
    }
    /// One line of muted context, cut to the width.
    fn muted(&mut self, text: &str) {
        for l in wrap(&visible(text), 2, 2, self.v.width) {
            let line = format!("  {}\n", self.p().paint(Token::Muted, &l));
            self.buf.push_str(&line);
        }
    }
    fn heading(&mut self, word: W) {
        let line = format!("{}\n", self.p().paint(Token::Strong, w(self.lang(), word)));
        self.buf.push_str(&line);
    }
    /// A status line at the left edge: mark, then the sentence.
    fn status(&mut self, m: Mark, text: &str) {
        let safe = visible(text);
        let lines = wrap(&safe, 2, 2, self.v.width);
        for (i, l) in lines.iter().enumerate() {
            let line = if i == 0 {
                format!(
                    "{} {}\n",
                    mark(self.p(), m),
                    self.p().paint(Token::Strong, l)
                )
            } else {
                format!("  {}\n", self.p().paint(Token::Strong, l))
            };
            self.buf.push_str(&line);
        }
    }
    /// Rows with their labels in one column.
    fn rows(&mut self, rows: &[Row]) {
        let label_w = rows.iter().map(|r| cells(&r.label)).max().unwrap_or(0);
        let narrow = self.v.width < 60;
        for r in rows {
            let label = visible(&r.label);
            let text = visible(&r.text);
            let m = mark(self.p(), r.mark);
            if narrow {
                // Label on the mark line, text below it.
                let head = cut(&label, self.v.width.saturating_sub(4), self.p().unicode);
                let line = format!("  {m} {head}\n");
                self.buf.push_str(&line);
                for l in wrap(&text, 4, 4, self.v.width) {
                    self.buf.push_str(&format!("    {l}\n"));
                }
                continue;
            }
            let pad = " ".repeat(label_w.saturating_sub(cells(&label)));
            let first = 2 + 2 + label_w + 2;
            let lines = wrap(&text, first, first, self.v.width);
            for (i, l) in lines.iter().enumerate() {
                let line = if i == 0 {
                    format!("  {m} {label}{pad}  {l}\n")
                } else {
                    format!("{}{l}\n", " ".repeat(first))
                };
                self.buf.push_str(&line);
            }
        }
    }
    fn header(&mut self, v: &Value) {
        let lang = self.lang();
        let line = format!(
            "{} {}\n",
            self.p().mark(Token::Accent),
            self.p().paint(Token::Strong, "MoAI update")
        );
        self.buf.push_str(&line);
        let plan = &v["plan"];
        let from = plan["source_version"]
            .as_str()
            .unwrap_or(w(lang, W::NotInstalled));
        let to = plan["target_version"]
            .as_str()
            .or_else(|| v["result"]["package_version"].as_str())
            .unwrap_or("?");
        let arrow = if from == to {
            "="
        } else if self.p().unicode {
            "→"
        } else {
            "->"
        };
        if !plan.is_null() {
            self.muted(&format!(
                "{} {from} {arrow} {} {to}",
                w(lang, W::Installed),
                w(lang, W::Package)
            ));
        } else {
            self.muted(&format!("{} {to}", w(lang, W::Package)));
        }
        let mut ctx = format!("HOME {}", self.home(&self.v.home.clone()));
        if let Some(pr) = &self.v.project {
            ctx.push_str(&format!(
                " · {} {}",
                w(lang, W::ProjectWord),
                self.home(&pr.path)
            ));
        }
        self.muted(&ctx);
    }
    fn project_line(&mut self) {
        let lang = self.lang();
        if let Some(Project {
            go_items: Some(n), ..
        }) = &self.v.project
        {
            let rows = [row(
                Mark::Warn,
                w(lang, W::ProjectWord).into(),
                format!("{} {}", w(lang, W::GoPending), count(lang, *n)),
            )];
            self.rows(&rows);
        }
    }
    /// #2872 stage 2: the cleanup is confirmed with its own token, never by the `y`; the command
    /// is shown whole so it can be copied.
    fn cleanup_next(&mut self, v: &Value) {
        let Some(flag) = v["cleanup"]["apply_flag"].as_str() else {
            return;
        };
        if !v["cleanup_result"].is_null() {
            return;
        }
        self.blank();
        let line = format!(
            "{} {}\n",
            self.p().mark(Token::Info),
            visible(w(self.lang(), W::CleanupNext))
        );
        self.buf.push_str(&line);
        let line = format!(
            "  {}\n",
            self.p()
                .paint(Token::Code, &visible(&format!("moai update {flag}")))
        );
        self.buf.push_str(&line);
    }
    /// A cleanup that ran (`--confirm-cleanup`): how many entries were archived and the revert.
    fn cleanup_done(&mut self, v: &Value) {
        let r = &v["cleanup_result"];
        if r.is_null() {
            return;
        }
        let lang = self.lang();
        self.blank();
        let moved = r["moved"].as_u64().unwrap_or(0) as usize;
        self.status(
            Mark::Ok,
            &format!("{} {}", w(lang, W::Cleaned), count(lang, moved)),
        );
        if let Some(c) = v["revert"]["cleanup"].as_str() {
            let c = self.home(c);
            let line = format!(
                "  {} {}\n",
                self.p()
                    .paint(Token::Muted, &format!("{}:", w(lang, W::Revert))),
                self.p().paint(Token::Code, &visible(&c))
            );
            self.buf.push_str(&line);
        }
    }
    /// #2894 7b (U-9): the end block of a finished run and of "all up to date" — always, inside
    /// a project or not: update each existing project from a session there. Inside a MoAI
    /// project, its name and state join the block.
    fn project_next(&mut self) {
        let lang = self.lang();
        self.blank();
        let text = w(lang, W::ProjectNext);
        let lines = wrap(text, 2, 2, self.v.width);
        for (i, l) in lines.iter().enumerate() {
            let lead = if i == 0 {
                format!("{} ", self.p().mark(Token::Info))
            } else {
                "  ".to_string()
            };
            let line = format!("{lead}{}\n", self.p().paint(Token::Strong, l));
            self.buf.push_str(&line);
        }
        if let Some(pr) = self.v.project.clone() {
            let state = match pr.go_items {
                Some(n) => format!("{} {}", w(lang, W::GoWaiting), count(lang, n)),
                None => w(lang, W::SessionChecks).to_string(),
            };
            let text = format!("{}: {} — {state}", w(lang, W::ThisFolder), pr.name());
            for l in wrap(&visible(&text), 2, 2, self.v.width) {
                let line = format!("  {}\n", self.p().paint(Token::Strong, &l));
                self.buf.push_str(&line);
            }
        }
    }
    /// The files that change, per layer; unchanged ones are only counted above.
    fn changes(&mut self, v: &Value) {
        let lang = self.lang();
        let mut groups: Vec<(String, Vec<(Mark, String)>)> = Vec::new();
        for h in v["plan"]["hosts"].as_array().into_iter().flatten() {
            if h["preflight"].as_str() != Some("ok") {
                continue;
            }
            let items: Vec<(Mark, String)> = h["files"]
                .as_array()
                .into_iter()
                .flatten()
                .filter_map(|f| file_change(f, "path"))
                .collect();
            if !items.is_empty() {
                groups.push((
                    format!(
                        "{} {}",
                        w(lang, W::Skills),
                        h["host"].as_str().unwrap_or("")
                    ),
                    items,
                ));
            }
        }
        let mut agent_items = Vec::new();
        for h in v["agents"]["hosts"].as_array().into_iter().flatten() {
            for f in h["files"].as_array().into_iter().flatten() {
                if let Some((m, name)) = file_change(f, "file") {
                    agent_items.push((m, format!("{} {name}", h["host"].as_str().unwrap_or(""))));
                }
            }
        }
        if !agent_items.is_empty() {
            groups.push((w(lang, W::Agents).into(), agent_items));
        }
        let routing: Vec<(Mark, String)> = v["routing"]["hosts"]
            .as_array()
            .into_iter()
            .flatten()
            .filter(|h| h["preflight"].as_str() == Some("ok"))
            .filter_map(|h| {
                let m = change_mark(h["section"]["action"].as_str().unwrap_or(""))?;
                Some((m, self.home(h["file"].as_str().unwrap_or(""))))
            })
            .collect();
        if !routing.is_empty() {
            groups.push((w(lang, W::Routing).into(), routing));
        }
        let style: Vec<(Mark, String)> = v["style"]["targets"]
            .as_array()
            .into_iter()
            .flatten()
            .filter_map(|t| {
                let m = change_mark(t["action"].as_str().unwrap_or(""))?;
                Some((m, self.home(t["path"].as_str().unwrap_or(""))))
            })
            .collect();
        if !style.is_empty() {
            groups.push((w(lang, W::Style).into(), style));
        }
        let b = &v["bin_names"];
        if matches!(b["state"].as_str(), Some("create" | "replace_copy")) {
            groups.push((
                w(lang, W::BinNames).into(),
                vec![(Mark::Update, self.home(b["alias"].as_str().unwrap_or("")))],
            ));
        }
        for h in v["hooks"]["hosts"].as_array().into_iter().flatten() {
            if let Some(m) = change_mark(h["action"].as_str().unwrap_or("")) {
                groups.push((
                    w(lang, W::Hooks).into(),
                    vec![(m, h["host"].as_str().unwrap_or("").to_string())],
                ));
            }
        }
        if groups.is_empty() {
            return;
        }
        self.blank();
        self.heading(W::Changes);
        // Two skill hosts with the same files: the second says so instead of repeating them.
        let label_w = groups.iter().map(|(l, _)| cells(l)).max().unwrap_or(0);
        let mut prev: Option<Vec<String>> = None;
        const SHOWN: usize = 6;
        for (label, items) in &groups {
            let names: Vec<String> = items.iter().map(|(_, n)| n.clone()).collect();
            let label_s = visible(label).into_owned();
            let pad = " ".repeat(label_w.saturating_sub(cells(&label_s)));
            let first = 2 + label_w + 2;
            let narrow = self.v.width < 60;
            if prev.as_ref() == Some(&names) && label.starts_with(w(lang, W::Skills)) {
                let text = if lang == Lang::Ko {
                    format!("{} {}", w(lang, W::SameAs), count(lang, items.len()))
                } else {
                    format!("the same {} files", items.len())
                };
                if narrow {
                    self.buf.push_str(&format!("  {label_s}\n    {text}\n"));
                } else {
                    self.buf.push_str(&format!("  {label_s}{pad}  {text}\n"));
                }
                continue;
            }
            for (i, (m, name)) in items.iter().take(SHOWN).enumerate() {
                let lead = if narrow {
                    if i == 0 {
                        self.buf.push_str(&format!("  {label_s}\n"));
                    }
                    "    ".to_string()
                } else if i == 0 {
                    format!("  {label_s}{pad}  ")
                } else {
                    " ".repeat(first)
                };
                let room = self.v.width.saturating_sub(cells(&lead) + 2);
                let shown = cut(&visible(name), room, self.p().unicode);
                self.buf
                    .push_str(&format!("{lead}{} {shown}\n", mark(self.p(), *m)));
            }
            if items.len() > SHOWN {
                let lead = if narrow {
                    "    ".to_string()
                } else {
                    " ".repeat(first)
                };
                let more = if lang == Lang::Ko {
                    format!(
                        "… {} {}",
                        w(lang, W::More),
                        count(lang, items.len() - SHOWN)
                    )
                } else {
                    format!("… {} {}", items.len() - SHOWN, w(lang, W::More))
                };
                self.buf.push_str(&format!("{lead}{more}\n"));
            }
            if label.starts_with(w(lang, W::Skills)) {
                prev = Some(names);
            }
        }
    }
}

fn change_mark(action: &str) -> Option<Mark> {
    match action {
        "create" | "append" | "merge" | "set" | "update" | "replace" | "replace_copy" => {
            Some(Mark::Update)
        }
        "conflict" => Some(Mark::Warn),
        _ => None,
    }
}

fn file_change(f: &Value, key: &str) -> Option<(Mark, String)> {
    let m = change_mark(f["action"].as_str().unwrap_or(""))?;
    Some((m, f[key].as_str().unwrap_or("").to_string()))
}

fn writes(v: &Value) -> bool {
    v["nothing_to_apply"].as_bool() == Some(false)
}

/// The screen shown before `[y/N]` (on stderr): head, the check rows, the changes.
pub fn question(preview: &Value, view: &View) -> String {
    let mut o = Out {
        v: view,
        buf: String::new(),
    };
    o.header(preview);
    o.blank();
    o.heading(W::Checking);
    let rows = preview_rows(preview, o.lang());
    o.rows(&rows);
    o.project_line();
    o.changes(preview);
    o.blank();
    o.buf
}

/// The screen printed on stdout when the command ends.
pub fn screen(result: &Value, asked: Asked, view: &View) -> String {
    if result["mode"].as_str() == Some("applied") {
        return applied(result, view);
    }
    let mut o = Out {
        v: view,
        buf: String::new(),
    };
    let lang = o.lang();
    let rows = preview_rows(result, lang);
    let blocked = rows.iter().any(|r| r.mark == Mark::Error);
    match asked {
        // The question already showed the head, rows and changes: one line now.
        Asked::No => {
            o.status(Mark::Skip, w(lang, W::NotWritten));
            o.cleanup_next(result);
            return o.buf;
        }
        Asked::Stale => {
            o.status(Mark::Warn, w(lang, W::StaleRefused));
            o.blank();
        }
        Asked::NotAsked | Asked::Yes => {}
    }
    o.header(result);
    o.blank();
    if !writes(result) && !blocked {
        o.status(Mark::Ok, w(lang, W::AllUpToDate));
        let mut totals = Vec::new();
        for h in result["plan"]["hosts"].as_array().into_iter().flatten() {
            totals.push(format!(
                "{} {} {}",
                w(lang, W::Skills),
                h["host"].as_str().unwrap_or(""),
                h["files"].as_array().map_or(0, Vec::len)
            ));
        }
        let agents: usize = result["agents"]["hosts"]
            .as_array()
            .into_iter()
            .flatten()
            .map(|h| h["files"].as_array().map_or(0, Vec::len))
            .sum();
        totals.push(format!("{} {agents}", w(lang, W::Agents)));
        let routing = result["routing"]["hosts"].as_array().map_or(0, Vec::len);
        totals.push(format!("{} {routing}", w(lang, W::Routing)));
        o.muted(&totals.join(" · "));
        let others: Vec<Row> = rows
            .into_iter()
            .filter(|r| matches!(r.mark, Mark::Skip | Mark::Warn))
            .collect();
        if !others.is_empty() {
            o.blank();
            o.rows(&others);
        }
        o.cleanup_done(result);
        o.cleanup_next(result);
        o.project_next();
        return o.buf;
    }
    o.heading(W::Checking);
    o.rows(&rows);
    o.project_line();
    if writes(result) {
        o.changes(result);
    }
    o.blank();
    if blocked && !writes(result) {
        o.status(Mark::Error, w(lang, W::NothingApplied));
    } else if let Some(cmd) = result["confirm_with"].as_str() {
        let line = format!(
            "{} {}\n",
            o.p().mark(Token::Info),
            visible(w(lang, W::NextConfirm))
        );
        o.buf.push_str(&line);
        let line = format!("  {}\n", o.p().paint(Token::Code, &visible(cmd)));
        o.buf.push_str(&line);
    }
    o.cleanup_done(result);
    o.cleanup_next(result);
    o.buf
}

/// The end of a confirmed run: what each layer did, how to revert.
fn applied(v: &Value, view: &View) -> String {
    let mut o = Out {
        v: view,
        buf: String::new(),
    };
    let lang = o.lang();
    let mut rows = Vec::new();
    let mut total = 0usize;
    let mut bad = false;
    for h in v["result"]["hosts"].as_array().into_iter().flatten() {
        let label = format!(
            "{} {}",
            w(lang, W::Skills),
            h["host"].as_str().unwrap_or("")
        );
        let n = h["files"].as_array().map_or(0, Vec::len);
        if h["status"].as_str() == Some("applied") {
            total += n;
            rows.push(row(
                Mark::Ok,
                label,
                format!("{} {}", count(lang, n), w(lang, W::Changed)),
            ));
        } else {
            bad = true;
            rows.push(blocked_row(lang, label, &h["blocked_reason"]));
        }
    }
    let mut part = |label: W, r: &Value, rows: &mut Vec<Row>| {
        let label = w(lang, label).to_string();
        match r["state"].as_str() {
            Some("applied") => {
                let n = r["written"].as_array().map_or(0, Vec::len);
                total += n;
                rows.push(row(
                    Mark::Ok,
                    label,
                    format!("{} {}", count(lang, n), w(lang, W::Changed)),
                ));
            }
            Some("unchanged") | None => {}
            Some("skipped") => rows.push(row(Mark::Skip, label, w(lang, W::Skipped).into())),
            Some(other) => {
                bad = true;
                let msg = r["message"]
                    .as_str()
                    .or(r["reason"].as_str())
                    .unwrap_or(other);
                rows.push(row(Mark::Warn, label, msg.to_string()));
            }
        }
    };
    part(W::Agents, &v["agents_result"], &mut rows);
    part(W::Routing, &v["routing_result"], &mut rows);
    part(W::Style, &v["style_result"], &mut rows);
    part(W::BinNames, &v["bin_names_result"], &mut rows);
    part(W::Hooks, &v["hooks_result"], &mut rows);
    for s in v["skipped"].as_array().into_iter().flatten() {
        rows.push(row(
            Mark::Skip,
            s["part"].as_str().unwrap_or("").to_string(),
            format!(
                "{}: {}",
                w(lang, W::Skipped),
                s["reason"].as_str().unwrap_or("")
            ),
        ));
    }
    // A blank line after the answer typed at the question.
    o.blank();
    o.heading(W::Applying);
    o.rows(&rows);
    o.blank();
    let head = if bad { W::Partial } else { W::Done };
    let m = if bad { Mark::Warn } else { Mark::Ok };
    let total = if lang == Lang::Ko {
        count(lang, total)
    } else {
        format!("{total} {}", w(lang, W::Changed))
    };
    o.status(m, &format!("{} — {total}", w(lang, head)));
    let reverts: Vec<String> = ["agents", "routing", "hooks"]
        .iter()
        .filter_map(|k| v["revert"][*k].as_str())
        .map(|c| o.home(c))
        .collect();
    // Revert commands stay whole (never cut) so they can be copied.
    for c in &reverts {
        let line = format!(
            "  {} {}\n",
            o.p()
                .paint(Token::Muted, &format!("{}:", w(lang, W::Revert))),
            o.p().paint(Token::Code, &visible(c))
        );
        o.buf.push_str(&line);
    }
    let reload = v["result"]["hosts"]
        .as_array()
        .into_iter()
        .flatten()
        .any(|h| h["reload_required"].as_str() == Some("yes"));
    if reload {
        o.muted(w(lang, W::Reload));
    }
    o.cleanup_done(v);
    // A partly applied run ends with what failed, not the next step (#2894 7b).
    if !bad {
        o.project_next();
    }
    o.buf
}

// --- the running process ------------------------------------------------------------------

/// Whether this run shows screens: stdout is a terminal, `--json` was not given and
/// `MOAI_OUTPUT` is not `quiet` (D3-U). Everything else keeps the envelope byte for byte.
pub fn wants_screens(json_flag: bool) -> bool {
    use std::io::IsTerminal;
    !json_flag
        && std::io::stdout().is_terminal()
        && crate::ui_summary::output_mode(std::env::var_os("MOAI_OUTPUT").as_deref())
            != crate::ui_summary::OutputMode::Quiet
}

/// Whether the question can be asked: stdin and stderr are both terminals (the rule of
/// `progress::ask_terminal`).
pub fn can_ask() -> bool {
    use std::io::IsTerminal;
    std::io::stdin().is_terminal() && std::io::stderr().is_terminal()
}

/// The MoAI project holding `cwd`, looked for in `cwd` and its parents but never in `home`
/// itself (HOME holds the installed state, which is not a project). Which generation owns the
/// folder is `state_root::classify`'s verdict (#2725): `go` — and a stopped or mixed move — is a
/// project still to move, counted with the read-only `go_state::plan`; `rust`, or an earlier
/// project state folder (`state_root::is_new_project` false), is a project of this generation.
/// Reads only.
pub fn project_here(cwd: &std::path::Path, home: &std::path::Path) -> Option<Project> {
    for dir in cwd.ancestors() {
        if dir == home {
            continue;
        }
        let path = dir.display().to_string();
        match crate::state_root::classify(dir).generation {
            "go" | "mixed" | "interrupted" => {
                let moves = crate::go_state::plan(dir)
                    .map(|p| p.items.iter().filter(|i| i.action == "move").count())
                    .unwrap_or(0);
                return Some(Project {
                    path,
                    go_items: Some(moves),
                });
            }
            "rust" => {
                return Some(Project {
                    path,
                    go_items: None,
                })
            }
            "none" if !crate::state_root::is_new_project(dir) => {
                return Some(Project {
                    path,
                    go_items: None,
                })
            }
            _ => {}
        }
    }
    None
}

/// The view for `stream` in this process: its painter (`MOAI_OUTPUT=plain` drops color and
/// UTF-8 symbols, not the language), its width, HOME and the project of the current folder.
pub fn view_for(stream: crate::ui::Stream) -> View {
    let env = crate::ui::Env::from_process();
    let normal = || Painter::for_stream(stream, crate::ui::ColorFlag::Auto);
    let painter = match crate::ui_summary::output_mode(std::env::var_os("MOAI_OUTPUT").as_deref()) {
        crate::ui_summary::OutputMode::Plain => Painter::plain().with_lang(normal().lang),
        _ => normal(),
    };
    let home = std::env::var_os("HOME")
        .map(std::path::PathBuf::from)
        .unwrap_or_default();
    let cwd = std::env::current_dir().unwrap_or_default();
    View {
        painter,
        width: crate::ui::width(stream, &env)
            .unwrap_or(crate::ui_summary::DEFAULT_WIDTH)
            .max(20),
        home: home.display().to_string(),
        project: project_here(&cwd, &home),
    }
}
