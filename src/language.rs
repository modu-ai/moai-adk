//! Per-project writing language (Issue #2293): the `language` key of `.moai-cli/project.yaml`,
//! the setup input step, and the sentences the AGENTS.md managed section renders from it.
//!
//! The setup asks three questions: the conversation language (`user_chat`; the AGENTS.md and
//! REVIEW.md managed sections are written in it), the commit-message language and the document
//! language (`documents`: reports and SPEC documents). Everything else is English without asking:
//! instruction files, skills and rules, messages between agents, code comments, issue and PR
//! records. The older profile items (`reports`, `issue_pr_records`, `agent_messages`,
//! `instruction_files`, `skill_instructions`, `code_comments`) are still accepted and, when a
//! profile sets one, still mean what they meant; the setup neither asks for them nor adds them.
//!
//! The item list, the accepted language codes and each item's default live in the two
//! tables below and nowhere else. The defaults are stated values chosen for the product,
//! not measured ones. The input step never writes the profile: it returns a complete
//! profile that goes through `project-plan`/`project-apply` like any other change.

use serde_json::{json, Map, Value};
use std::collections::BTreeMap;
use std::io::{BufRead, Write};

type Result<T> = std::result::Result<T, String>;

/// The `language` value: item id → language code. A `BTreeMap` so the serialized key order
/// is fixed and a re-run of the input step reproduces the same bytes.
pub type LanguageSettings = BTreeMap<String, String>;

/// Where an item's default comes from when the profile does not set it.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum DefaultSource {
    /// English (`en`).
    English,
    /// The conversation language: the managed-section request language when rendering,
    /// the `--conversation` code (or the profile's `user_chat`) in the input step.
    Conversation,
}

/// One writing item: its profile id, its label in English and Korean, its default, and whether
/// the setup asks for it.
#[derive(Clone, Copy, Debug)]
pub struct Item {
    pub id: &'static str,
    pub label_en: &'static str,
    pub label_ko: &'static str,
    pub default: DefaultSource,
    /// The three items the setup asks; the others are older profile items, kept readable.
    pub asked: bool,
}

/// Every item the `language` key may name, in rendering order; the setup asks the ones marked
/// `asked`, in the order they appear here.
pub const ITEMS: &[Item] = &[
    Item {
        id: "user_chat",
        label_en: "Conversation with the user",
        label_ko: "사용자 대화",
        default: DefaultSource::Conversation,
        // The conversation language is asked once at install and lives in the HOME state
        // folder; the project setup reads that value instead of asking again.
        asked: false,
    },
    Item {
        id: "reports",
        label_en: "Reports to the user",
        label_ko: "사용자 보고",
        default: DefaultSource::English,
        asked: false,
    },
    Item {
        id: "issue_pr_records",
        label_en: "Issue and PR records",
        label_ko: "Issue·PR 기록",
        default: DefaultSource::English,
        asked: false,
    },
    Item {
        id: "commit_messages",
        label_en: "Commit messages",
        label_ko: "커밋 메시지",
        default: DefaultSource::Conversation,
        asked: true,
    },
    Item {
        id: "agent_messages",
        label_en: "Messages between agents",
        label_ko: "에이전트 간 메시지",
        default: DefaultSource::English,
        asked: false,
    },
    Item {
        id: "instruction_files",
        label_en: "Instruction files (AGENTS.md, rules)",
        label_ko: "지침 파일(AGENTS.md, 규칙)",
        default: DefaultSource::English,
        asked: true,
    },
    Item {
        id: "skill_instructions",
        label_en: "Skill instructions",
        label_ko: "skill 지침",
        default: DefaultSource::English,
        asked: false,
    },
    Item {
        id: "code_comments",
        label_en: "Code comments",
        label_ko: "코드 주석",
        default: DefaultSource::English,
        asked: false,
    },
    Item {
        id: "documents",
        label_en: "Documents (reports and SPEC documents)",
        label_ko: "문서(보고서·SPEC)",
        default: DefaultSource::Conversation,
        asked: true,
    },
];

/// Accepted language codes (ISO 639-1) with their English and Korean names. Closed on
/// purpose: every accepted code needs a rendered name.
pub const CODES: &[(&str, &str, &str)] = &[
    ("en", "English", "영어"),
    ("ko", "Korean", "한국어"),
    ("ja", "Japanese", "일본어"),
    ("zh", "Chinese", "중국어"),
];

const INVALID: &str =
    "invalid language setting: unknown item or unsupported language code; no values were logged";

fn item(id: &str) -> Option<&'static Item> {
    ITEMS.iter().find(|i| i.id == id)
}

pub fn is_code(code: &str) -> bool {
    CODES.iter().any(|(c, _, _)| *c == code)
}

fn code_name(code: &str, ui: &str) -> &'static str {
    CODES
        .iter()
        .find(|(c, _, _)| *c == code)
        .map(|(_, en, ko)| if ui == "ko" { *ko } else { *en })
        .expect("validated language code")
}

/// Accepts a `language` value naming at least one known item, each with an accepted code.
pub fn validate(settings: &LanguageSettings) -> Result<()> {
    if settings.is_empty()
        || !settings
            .iter()
            .all(|(id, code)| item(id).is_some() && is_code(code))
    {
        return Err(INVALID.into());
    }
    Ok(())
}

fn default_for(item: &Item, conversation: &str) -> String {
    match item.default {
        DefaultSource::English => "en".into(),
        DefaultSource::Conversation => conversation.into(),
    }
}

/// The code of every item, in `ITEMS` order: the profile value, else the item default.
pub fn resolve(settings: &LanguageSettings, conversation: &str) -> Vec<(&'static Item, String)> {
    ITEMS
        .iter()
        .map(|i| {
            let code = settings
                .get(i.id)
                .cloned()
                .unwrap_or_else(|| default_for(i, conversation));
            (i, code)
        })
        .collect()
}

/// The managed-section lines for a set `language` key, written in `ui` (`en` or `ko`): one line
/// per item the profile sets, in `ITEMS` order, then one sentence that everything not listed is
/// English. An item the profile leaves out gets no line —
/// its language is never guessed (card C-87 S1; before, it was rendered with its default and a
/// conversation-default item took the request language, so `reports` read English beside a
/// Japanese `user_chat`). `profile_rel` is where the project keeps `project.yaml`
/// (`state_root::rel`, #2696 C2).
pub fn render_lines(settings: &LanguageSettings, ui: &str, profile_rel: &str) -> String {
    let mut out = if ui == "ko" {
        format!("항목별 작성 언어(`{profile_rel}`의 `language`):\n")
    } else {
        format!("Writing language by item (`language` in `{profile_rel}`):\n")
    };
    for item in ITEMS {
        let Some(code) = settings.get(item.id) else {
            continue;
        };
        let label = if ui == "ko" {
            item.label_ko
        } else {
            item.label_en
        };
        out.push_str(&format!("- {label}: {} (`{code}`)\n", code_name(code, ui)));
    }
    out.push_str(if ui == "ko" {
        "그 밖의 글(지침 파일, 스킬·규칙, 에이전트 간 메시지, 코드 주석, Issue·PR 기록)은 위에 따로 적지 않으면 영어로 써요.\n"
    } else {
        "All other writing (instruction files, skills and rules, messages between agents, code comments, issue and PR records) is in English unless a line above says otherwise.\n"
    });
    out
}

/// The English name of a supported language code (`ko` → `Korean`), or `None` for any other
/// code.
pub fn english_name(code: &str) -> Option<&'static str> {
    CODES
        .iter()
        .find(|(c, _, _)| *c == code)
        .map(|(_, en, _)| *en)
}

/// The conversation language the input step uses for conversation-default items: the
/// explicit `--conversation` code, else the profile's `user_chat`, else English.
pub fn conversation_language(
    explicit: Option<&str>,
    current: Option<&LanguageSettings>,
) -> Result<String> {
    conversation_language_in(explicit, current, None)
}

/// [`conversation_language`] with the HOME value between the profile and English (#2935 L1):
/// `--conversation`, then the profile's `user_chat`, then `home` (the code
/// [`home_conversation`] read, so this function touches no file), then `en`. The flag is
/// checked; `home` is trusted to be a code already.
pub fn conversation_language_in(
    explicit: Option<&str>,
    current: Option<&LanguageSettings>,
    home: Option<&str>,
) -> Result<String> {
    if let Some(code) = explicit {
        if !is_code(code) {
            return Err(INVALID.into());
        }
        return Ok(code.into());
    }
    Ok(current
        .and_then(|c| c.get("user_chat"))
        .map(String::as_str)
        .or(home)
        .unwrap_or("en")
        .into())
}

/// The file the installer's choice lives in, below the HOME state folder (#2935 L1). It is not
/// the Go generation's `config/sections/language.yaml`: that is a project section file, this is
/// one user value beside `config/launcher.yaml`.
pub const HOME_FILE: [&str; 2] = ["config", "language.yaml"];

/// The largest HOME language file read (bytes); a larger one means no value.
pub const HOME_FILE_LIMIT: u64 = 4096;

/// The longest language code `language-set` stores (BCP-47 allows more; this bounds the file).
pub const TAG_MAX: usize = 35;

/// Whether `code` is language-code-shaped text: `^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$` and at
/// most [`TAG_MAX`] bytes (#2935 L1). The shape only: `fr` and `pt-BR` pass, and so does `KO`;
/// whether MoAI has a name for the language is [`is_code`]. Nothing outside ASCII letters,
/// digits and `-` passes, so a stored value can hold nothing else in its YAML line.
pub fn is_language_tag(code: &str) -> bool {
    if code.len() > TAG_MAX {
        return false;
    }
    let mut parts = code.split('-');
    let primary = parts.next().unwrap_or("");
    (2..=3).contains(&primary.len())
        && primary.bytes().all(|b| b.is_ascii_alphabetic())
        && parts.all(|p| (2..=8).contains(&p.len()) && p.bytes().all(|b| b.is_ascii_alphanumeric()))
}

/// The conversation language recorded in `home`'s state folder, as stored: the
/// `language.conversation_language` of [`HOME_FILE`] when it is [`is_language_tag`]-shaped. It
/// may be a language MoAI has no name for (`fr`); use [`home_conversation_supported`] where a
/// name is needed. A missing, linked, wrong-typed, oversized, non-UTF-8 or malformed file, and a
/// value of another shape, give `None` — never an error.
pub fn home_conversation(home: &std::path::Path) -> Option<String> {
    let root = crate::home_root::path(home);
    let bytes = crate::state_root::read_limited(&root, &HOME_FILE, HOME_FILE_LIMIT)?;
    let text = String::from_utf8(bytes).ok()?;
    crate::go_settings::section_pairs(&text, "language")?
        .into_iter()
        .find(|(key, _)| key == "conversation_language")
        .map(|(_, value)| value)
        .filter(|value| is_language_tag(value))
}

/// [`home_conversation`] when the value is also one of [`CODES`], else `None`: the conversation
/// default of `language-setup`, whose item resolver and names only know those codes. Another
/// language (`fr`) falls through to the next source, silently.
pub fn home_conversation_supported(home: &std::path::Path) -> Option<String> {
    home_conversation(home).filter(|code| is_code(code))
}

/// The bytes `language-set` writes: one key, nothing else.
fn home_file_bytes(code: &str) -> Vec<u8> {
    format!("language:\n  conversation_language: {code}\n").into_bytes()
}

/// `language-set --home <code>` (#2935 L1): records `code` in [`HOME_FILE`] below the HOME state
/// folder, once. Never overwrites (an existing file gives `written: false, reason: "exists"`),
/// never creates the state folder itself (`reason: "no_home_state"`; the installer only
/// reports it) and creates nothing but `config/`. A `config` or target that is a link or the
/// wrong file type is refused before any write. A state folder that is a link to a folder is
/// followed like every other HOME path; a dangling link is no folder. `code` is any
/// [`is_language_tag`] text and is written exactly as given (`pt-BR` stays `pt-BR`).
pub fn set_home(home: &std::path::Path, code: &str) -> Result<Value> {
    use std::fs;
    use std::io::ErrorKind::{AlreadyExists, NotFound};
    if !is_code(code) {
        return Err(INVALID.into());
    }
    let root = crate::home_root::path(home);
    let config = root.join(HOME_FILE[0]);
    let file = config.join(HOME_FILE[1]);
    let report = |written: bool, reason: Option<&str>| {
        json!({
            "path": file.display().to_string(),
            "written": written,
            "code": code,
            "reason": reason,
        })
    };
    match fs::metadata(&root) {
        Ok(m) if m.is_dir() => {}
        Ok(_) => return Err("language-set refused: the HOME state folder is not a folder".into()),
        Err(e) if e.kind() == NotFound => return Ok(report(false, Some("no_home_state"))),
        Err(_) => return Err("language-set refused: the HOME state folder cannot be read".into()),
    }
    match fs::symlink_metadata(&config) {
        Ok(m) if m.file_type().is_dir() => {}
        Ok(_) => return Err("language-set refused: config is a link or not a folder".into()),
        Err(e) if e.kind() == NotFound => {
            // Exclusive: only `config/` is made, and a folder that appeared meanwhile is fine.
            if let Err(e) = fs::create_dir(&config) {
                if e.kind() != AlreadyExists {
                    return Err("language-set refused: config folder create failed".into());
                }
            }
        }
        Err(_) => return Err("language-set refused: config cannot be read".into()),
    }
    let exists = || match fs::symlink_metadata(&file) {
        Ok(m) if m.file_type().is_file() => Ok(true),
        Ok(_) => Err("language-set refused: language.yaml is a link or not a file".to_string()),
        Err(e) if e.kind() == NotFound => Ok(false),
        Err(_) => Err("language-set refused: language.yaml cannot be read".to_string()),
    };
    if exists()? {
        // A later `language-set --home` with a valid code replaces the value: the before
        // bytes go with the publish, so a failed replace leaves the old file.
        let before = fs::read(&file)
            .map_err(|_| "language-set refused: language.yaml cannot be read".to_string())?;
        return match crate::apply::publish_replace(&file, Some(&before), &home_file_bytes(code)) {
            Ok(()) => Ok(report(true, Some("replaced"))),
            Err(e) => Err(e),
        };
    }
    match crate::apply::publish_exclusive(&file, &home_file_bytes(code)) {
        Ok(()) => Ok(report(true, None)),
        // A writer that got there first left an existing file, not a failure.
        Err(_) if exists()? => Ok(report(false, Some("exists"))),
        Err(e) => Err(e),
    }
}

fn suggested(current: Option<&LanguageSettings>, item: &Item, conversation: &str) -> String {
    current
        .and_then(|c| c.get(item.id))
        // `documents` is the successor of `reports`: a stored `reports` value is the suggestion
        // for it, never a copy of it.
        .or_else(|| {
            (item.id == "documents")
                .then(|| current.and_then(|c| c.get("reports")))
                .flatten()
        })
        .cloned()
        .unwrap_or_else(|| default_for(item, conversation))
}

/// The values an empty answer keeps: the profile's own value first, then the Go backup's
/// (#2817), then the item default. Returned as a map naming only items with a value.
fn with_go(current: Option<&LanguageSettings>, go: &LanguageSettings) -> LanguageSettings {
    let mut base = go.clone();
    if let Some(c) = current {
        base.extend(c.iter().map(|(k, v)| (k.clone(), v.clone())));
    }
    base
}

/// The question set a host question tool asks: one question per asked item (three) with the
/// current value, the value an empty answer keeps (`suggested`) and the accepted codes.
pub fn questions(current: Option<&LanguageSettings>, conversation: &str) -> Value {
    questions_with(current, &LanguageSettings::new(), conversation, "default")
}

/// [`questions`] with the Go backup's item values as the next source of `suggested` (#2817).
fn questions_with(
    current: Option<&LanguageSettings>,
    go: &LanguageSettings,
    conversation: &str,
    source: &str,
) -> Value {
    let base = with_go(current, go);
    let options: Vec<Value> = CODES
        .iter()
        .map(|(code, en, ko)| json!({"code": code, "name_en": en, "name_ko": ko}))
        .collect();
    let items: Vec<Value> = ITEMS
        .iter()
        .filter(|i| i.asked)
        .map(|i| {
            json!({
                "id": i.id,
                "question_en": format!("Which language for: {}?", i.label_en),
                "question_ko": format!("{}에 쓸 언어는?", i.label_ko),
                "current": current.and_then(|c| c.get(i.id)),
                "suggested": suggested(Some(&base), i, conversation),
                "default": default_for(i, conversation),
            })
        })
        .collect();
    json!({
        "conversation_language": conversation,
        // Where the conversation language came from: `flag`, `profile`, `installer` (the value
        // the installer saved: a confirmation, not a new question), `go_backup` or `default`.
        "conversation_source": source,
        "items": items,
        "options": options,
    })
}

/// Applies host-tool answers (`{item: code}`); an asked item left out keeps its suggested value.
/// The result names the three asked items explicitly, so a re-run with no answers is byte-stable,
/// and carries an older item only when `current` has it or an answer names it.
pub fn merge_answers(
    current: Option<&LanguageSettings>,
    answers: &Map<String, Value>,
    conversation: &str,
) -> Result<LanguageSettings> {
    let mut out = LanguageSettings::new();
    for (id, value) in answers {
        let code = value.as_str().ok_or(INVALID)?;
        if item(id).is_none() || !is_code(code) {
            return Err(INVALID.into());
        }
    }
    for i in ITEMS {
        let answer = answers.get(i.id).and_then(Value::as_str);
        let code = match (answer, i.asked) {
            (Some(code), _) => code.to_owned(),
            (None, true) => suggested(current, i, conversation),
            (None, false) => match current.and_then(|c| c.get(i.id)) {
                Some(code) => code.clone(),
                None => continue,
            },
        };
        out.insert(i.id.into(), code);
    }
    Ok(out)
}

/// How the setup input step gets its answers.
pub enum SetupMode {
    /// Print the question set for a host question tool; nothing is answered.
    Questions,
    /// Answers a host question tool collected, as `{item: code}`.
    Answers(Map<String, Value>),
    /// Ask item by item on the terminal.
    Interactive,
}

/// The setup input step (`moai-cli language-setup`). Reads the current profile and returns
/// either the question set or a complete profile with every item set, ready for
/// `project-plan`. Writes nothing; `project-apply` stays the only writer of the profile.
///
/// `home` is the user's HOME when the caller has one (#2935 L1): the conversation language
/// `language-set` recorded there is the default after `conversation`, before the profile's
/// earlier `user_chat`, the Go backup's value and English.
pub fn setup<R: BufRead, W: Write>(
    root: &std::path::Path,
    home: Option<&std::path::Path>,
    mode: SetupMode,
    conversation: Option<&str>,
    input: &mut R,
    prompt: &mut W,
) -> Result<Value> {
    let (mut profile, _) = crate::read_profile(root)?;
    let current = profile.language.clone();
    // #2817: the Go backup's values come after the profile's and before the built-in ones.
    let go = crate::go_settings::read(root);
    let home_chat = home.and_then(home_conversation_supported);
    let profile_value = current
        .as_ref()
        .and_then(|c| c.get("user_chat"))
        .filter(|c| is_code(c))
        .cloned();
    // The conversation language, in order: the explicit flag, the value the installer saved in
    // the HOME state folder, the profile's earlier user_chat (installs older than the home
    // record), the Go backup's value, then English. A code outside the accepted list skips its
    // source.
    let (conversation, source) = match (conversation, &home_chat, profile_value, &go.conversation) {
        (Some(code), _, _, _) => (code.to_string(), "flag".to_string()),
        (None, Some(code), _, _) => (code.clone(), "installer".to_string()),
        (None, None, Some(code), _) => (code, "profile".to_string()),
        (None, None, None, Some(code)) if is_code(code) => (code.clone(), "go_backup".to_string()),
        _ => ("en".to_string(), "default".to_string()),
    };
    let base = with_go(current.as_ref(), &go.items);
    let base_ref = (!base.is_empty()).then_some(&base);
    // The profile's own older items stay as they are; a value that only the Go backup has for an
    // item the setup does not ask is never copied in.
    let own = current.clone().unwrap_or_default();
    let mut answered: Vec<String> = Vec::new();
    let mut chosen = match mode {
        SetupMode::Questions => {
            let mut out = json!({"questions": questions_with(current.as_ref(), &go.items, &conversation, &source)});
            if go.found {
                out["go_backup"] = go.to_json();
            }
            return Ok(out);
        }
        SetupMode::Answers(answers) => {
            answered = answers.keys().cloned().collect();
            merge_answers(base_ref, &answers, &conversation)?
        }
        SetupMode::Interactive => {
            if go.found {
                let _ = writeln!(
                    prompt,
                    "Defaults include values read from {}/config/sections (nothing is copied).",
                    crate::state_root::GO_BACKUP
                );
            }
            interactive(base_ref, &conversation, input, prompt)?
        }
    };
    chosen.retain(|k, _| {
        item(k).is_some_and(|i| i.asked) || own.contains_key(k) || answered.contains(k)
    });
    for (k, v) in &own {
        chosen.entry(k.clone()).or_insert_with(|| v.clone());
    }
    // The profile carries the resolved conversation language in its user_chat: the setup
    // never asks it, and the home value wins over an earlier profile value.
    chosen.insert("user_chat".to_string(), conversation.clone());
    let changed: Vec<&str> = ITEMS
        .iter()
        .filter(|i| current.as_ref().and_then(|c| c.get(i.id)) != chosen.get(i.id))
        .map(|i| i.id)
        .collect();
    // project-plan accepts schema v2 only; the version is MoAI-owned schema state.
    profile.schema_version = 2;
    profile.language = Some(chosen.clone());
    let mut out = json!({
        "profile": profile,
        "language": chosen,
        "changed_items": changed,
        "profile_written": false,
        "next_actions": ["pass result.profile to `moai project-plan <project>` on stdin, review the plan, then `moai project-apply`"]
    });
    if go.found {
        out["go_backup"] = go.to_json();
    }
    Ok(out)
}

/// Upper bound on wrong answers to one question before the interactive step gives up.
pub const MAX_ATTEMPTS: usize = 3;

/// The interactive input step: one prompt per asked item (three) on `prompt`, one answer line per
/// item from `input`. An empty line or end of input keeps the suggested value; a code outside
/// `CODES` is asked again, up to `MAX_ATTEMPTS` times.
pub fn interactive<R: BufRead, W: Write>(
    current: Option<&LanguageSettings>,
    conversation: &str,
    input: &mut R,
    prompt: &mut W,
) -> Result<LanguageSettings> {
    let ui = if conversation == "ko" { "ko" } else { "en" };
    let codes: Vec<&str> = CODES.iter().map(|(c, _, _)| *c).collect();
    let mut out = LanguageSettings::new();
    let mut ended = false;
    for i in ITEMS.iter().filter(|i| i.asked) {
        let keep = suggested(current, i, conversation);
        let mut chosen = None;
        for _ in 0..MAX_ATTEMPTS {
            if ended {
                chosen = Some(keep.clone());
                break;
            }
            let label = if ui == "ko" { i.label_ko } else { i.label_en };
            let _ = write!(prompt, "{label} [{}] ({}): ", keep, codes.join("/"));
            let _ = prompt.flush();
            let mut line = String::new();
            let read = input
                .read_line(&mut line)
                .map_err(|_| "stdin read failed")?;
            let answer = line.trim();
            if read == 0 {
                ended = true;
                chosen = Some(keep.clone());
                break;
            }
            if answer.is_empty() {
                chosen = Some(keep.clone());
                break;
            }
            if is_code(answer) {
                chosen = Some(answer.to_owned());
                break;
            }
            let _ = writeln!(
                prompt,
                "unsupported language code; use one of {}",
                codes.join(", ")
            );
        }
        let code = chosen
            .ok_or("too many unsupported answers; nothing was changed; no values were logged")?;
        out.insert(i.id.into(), code);
    }
    Ok(out)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn settings(pairs: &[(&str, &str)]) -> LanguageSettings {
        pairs
            .iter()
            .map(|(k, v)| (k.to_string(), v.to_string()))
            .collect()
    }

    /// The item ids are unique and every default resolves to an accepted code.
    #[test]
    fn tables_are_consistent() {
        let mut ids: Vec<&str> = ITEMS.iter().map(|i| i.id).collect();
        ids.sort();
        ids.dedup();
        assert_eq!(ids.len(), ITEMS.len());
        for conversation in ["en", "ko"] {
            for i in ITEMS {
                assert!(is_code(&default_for(i, conversation)));
            }
        }
    }

    #[test]
    fn validate_rejects_unknown_items_bad_codes_and_empty() {
        assert!(validate(&settings(&[("code_comments", "en")])).is_ok());
        assert!(validate(&settings(&[("code_comment", "en")])).is_err());
        assert!(validate(&settings(&[("code_comments", "EN")])).is_err());
        assert!(validate(&settings(&[("code_comments", "fr")])).is_err());
        assert!(validate(&settings(&[("code_comments", "")])).is_err());
        assert!(validate(&LanguageSettings::new()).is_err());
    }

    #[test]
    fn merge_answers_names_the_three_asked_items_and_keeps_older_ones() {
        let current = settings(&[("code_comments", "ja")]);
        let answers = json!({"user_chat": "zh"});
        let out = merge_answers(Some(&current), answers.as_object().unwrap(), "ko").unwrap();
        // The three asked items are named; an older item is carried only because the profile
        // already had it; an older item the profile never set gets no key.
        assert_eq!(out.len(), 4, "{out:?}");
        assert_eq!(out["user_chat"], "zh");
        assert_eq!(out["commit_messages"], "ko");
        assert_eq!(out["documents"], "ko");
        assert_eq!(out["code_comments"], "ja");
        assert!(!out.contains_key("reports"));
        assert!(!out.contains_key("agent_messages"));
        for bad in [
            json!({"nope": "en"}),
            json!({"reports": "xx"}),
            json!({"reports": 1}),
        ] {
            assert!(merge_answers(None, bad.as_object().unwrap(), "en").is_err());
        }
    }

    #[test]
    fn interactive_keeps_on_empty_and_eof_and_retries_bad_codes() {
        // The three asked items, in table order: user_chat, commit_messages, documents.
        let mut prompt = Vec::new();
        let mut input = "ko\n\nxx\nja\n".as_bytes();
        let out = interactive(None, "en", &mut input, &mut prompt).unwrap();
        assert_eq!(out.len(), 3, "{out:?}");
        assert_eq!(out["user_chat"], "ko");
        // An empty line keeps the suggested value (the conversation language, `en` here).
        assert_eq!(out["commit_messages"], "en");
        // A bad code is asked again; the next line is the answer.
        assert_eq!(out["documents"], "ja");
        assert!(String::from_utf8(prompt)
            .unwrap()
            .contains("unsupported language code"));
        // End of input keeps the suggested value for every remaining item.
        let mut input = "ko\n".as_bytes();
        let out = interactive(None, "en", &mut input, &mut Vec::new()).unwrap();
        assert_eq!(out["user_chat"], "ko");
        assert_eq!(out["commit_messages"], "en");
        assert_eq!(out["documents"], "en");
        assert_eq!(out.len(), 3);
    }

    #[test]
    fn interactive_gives_up_after_max_attempts() {
        let mut input = "xx\nyy\nzz\nko\n".as_bytes();
        assert!(interactive(None, "en", &mut input, &mut Vec::new()).is_err());
        let mut input = "xx\nyy\nko\n".as_bytes();
        let out = interactive(None, "en", &mut input, &mut Vec::new()).unwrap();
        assert_eq!(out["user_chat"], "ko");
    }

    /// #2935 L1: `--conversation`, then `user_chat`, then the HOME value, then `en`.
    #[test]
    fn the_home_value_sits_between_the_profile_and_english() {
        let chat = settings(&[("user_chat", "zh")]);
        let other = settings(&[("code_comments", "zh")]);
        let layered = |explicit, current: Option<&LanguageSettings>, home| {
            conversation_language_in(explicit, current, home).unwrap()
        };
        assert_eq!(layered(Some("ko"), Some(&chat), Some("ja")), "ko");
        assert_eq!(layered(None, Some(&chat), Some("ja")), "zh");
        assert_eq!(layered(None, Some(&other), Some("ja")), "ja");
        assert_eq!(layered(None, None, Some("ja")), "ja");
        assert_eq!(layered(None, None, None), "en");
        assert_eq!(layered(None, Some(&other), None), "en");
        // The flag is still checked, whatever HOME says; the old two-argument form has no HOME.
        assert!(conversation_language_in(Some("xx"), None, Some("ja")).is_err());
        assert_eq!(conversation_language(None, None).unwrap(), "en");
    }

    /// The HOME state folder of a temporary HOME without a marker (asked of `home_root`).
    fn state(home: &std::path::Path) -> std::path::PathBuf {
        crate::home_root::path(home)
    }

    fn home_with(file: Option<&[u8]>) -> tempfile::TempDir {
        let home = tempfile::tempdir().unwrap();
        let config = state(home.path()).join("config");
        std::fs::create_dir_all(&config).unwrap();
        if let Some(bytes) = file {
            std::fs::write(config.join("language.yaml"), bytes).unwrap();
        }
        home
    }

    #[test]
    fn home_conversation_reads_only_a_good_file() {
        let read = |bytes: &[u8]| home_conversation(home_with(Some(bytes)).path());
        assert_eq!(
            read(b"language:\n  conversation_language: ko\n").as_deref(),
            Some("ko")
        );
        assert_eq!(
            read(b"language:\n  conversation_language: \"ja\" # chosen\n").as_deref(),
            Some("ja")
        );
        // Another language is returned as stored; `home_conversation_supported` is the filter.
        for other in ["fr", "pt-BR", "es-419", "KO"] {
            let text = format!("language:\n  conversation_language: {other}\n");
            let home = home_with(Some(text.as_bytes()));
            assert_eq!(home_conversation(home.path()).as_deref(), Some(other));
            let supported = home_conversation_supported(home.path());
            assert_eq!(supported.is_some(), is_code(other), "{other}");
        }
        for bad in [
            &b""[..],
            b"\xff\xfe",
            b"language: ko\n",
            b"language:\n  conversation_language: ko ja\n",
            b"language:\n  conversation_language: k\n",
            b"language:\n  conversation_language: ko-\n",
            b"language:\n  conversation_language: abcd\n",
            b"language:\n  conversation_language: [ko]\n",
            b"conversation_language: ko\n",
            b"language:\n  conversation_language: ko\nlanguage:\n  x: y\n",
        ] {
            assert_eq!(read(bad), None, "{bad:?}");
        }
        assert_eq!(home_conversation(home_with(None).path()), None, "no file");
        let bare = tempfile::tempdir().unwrap();
        assert_eq!(home_conversation(bare.path()), None, "no state folder");
    }

    /// The grammar `^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$` and the 35-byte limit, at each edge.
    #[test]
    fn a_language_tag_has_the_documented_shape() {
        let longest = format!("ko{}-abcde", "-abcdefgh".repeat(3));
        let too_long = format!("{longest}f");
        assert_eq!((longest.len(), too_long.len()), (TAG_MAX, TAG_MAX + 1));
        for good in [
            "en",
            "ko",
            "fr",
            "kor",
            "KO",
            "pt-BR",
            "es-419",
            "ko-ab",
            "ko-abcdefgh",
            "zh-Hant-TW",
            "en-a1",
            &longest,
        ] {
            assert!(is_language_tag(good), "{good}");
        }
        for bad in [
            "",
            "e",
            "k",
            "abcd",
            "ko-",
            "-ko",
            "ko--KR",
            "ko-a",
            "ko-abcdefghi",
            "ko_KR",
            "ko ",
            " ko",
            "ko\n",
            "ko\t",
            "ko'",
            "ko\"",
            "ko;",
            "한국어",
            "ko-한국",
            "k0",
            "1a",
            "../ko",
            "$HOME",
            &too_long,
        ] {
            assert!(!is_language_tag(bad), "{bad:?}");
        }
    }

    #[test]
    fn home_conversation_stops_at_the_size_limit() {
        let head = "language:\n  conversation_language: ko\n";
        let sized = |total: usize| {
            let mut text = String::from(head);
            text.push_str(&"#".repeat(total - head.len() - 1));
            text.push('\n');
            assert_eq!(text.len(), total);
            home_conversation(home_with(Some(text.as_bytes())).path())
        };
        assert_eq!(sized(HOME_FILE_LIMIT as usize).as_deref(), Some("ko"));
        assert_eq!(sized(HOME_FILE_LIMIT as usize + 1), None);
    }

    #[cfg(unix)]
    #[test]
    fn home_conversation_never_follows_a_link() {
        let good = b"language:\n  conversation_language: ko\n";
        let home = home_with(None);
        let dir = state(home.path());
        std::fs::write(home.path().join("real.yaml"), good).unwrap();
        std::os::unix::fs::symlink(
            home.path().join("real.yaml"),
            dir.join("config/language.yaml"),
        )
        .unwrap();
        assert_eq!(home_conversation(home.path()), None, "linked file");
        let home = home_with(None);
        let dir = state(home.path());
        std::fs::create_dir(home.path().join("real")).unwrap();
        std::fs::write(home.path().join("real/language.yaml"), good).unwrap();
        std::fs::remove_dir(dir.join("config")).unwrap();
        std::os::unix::fs::symlink(home.path().join("real"), dir.join("config")).unwrap();
        assert_eq!(home_conversation(home.path()), None, "linked config");
    }

    #[test]
    fn set_home_writes_once_and_what_it_writes_is_read_back() {
        let home = home_with(None);
        let file = state(home.path()).join("config/language.yaml");
        let first = set_home(home.path(), "ja").unwrap();
        assert_eq!(first["written"], true);
        assert_eq!(first["reason"], Value::Null);
        assert_eq!(first["code"], "ja");
        assert_eq!(first["path"], file.display().to_string());
        assert_eq!(
            std::fs::read(&file).unwrap(),
            b"language:\n  conversation_language: ja\n"
        );
        assert_eq!(home_conversation(home.path()).as_deref(), Some("ja"));
        let second = set_home(home.path(), "ko").unwrap();
        assert_eq!(second["written"], false);
        assert_eq!(second["reason"], "exists");
        assert_eq!(
            std::fs::read(&file).unwrap(),
            b"language:\n  conversation_language: ja\n"
        );
    }

    #[test]
    fn set_home_checks_the_code_and_the_state_folder_first() {
        let bare = tempfile::tempdir().unwrap();
        let none = set_home(bare.path(), "ko").unwrap();
        assert_eq!(none["written"], false);
        assert_eq!(none["reason"], "no_home_state");
        assert_eq!(std::fs::read_dir(bare.path()).unwrap().count(), 0);
        assert!(set_home(bare.path(), "k").is_err());
        let home = home_with(None);
        for bad in ["", "k", "abcd", "ko-", "ko ", "ko\n", "ko'", "한국어"] {
            assert!(set_home(home.path(), bad).is_err(), "{bad:?}");
        }
        assert!(!state(home.path()).join("config/language.yaml").exists());
    }

    /// Another language is stored exactly as given and read back as stored, but
    /// `language-setup` does not use it.
    #[test]
    fn set_home_stores_another_language_exactly_as_given() {
        let home = home_with(None);
        let file = state(home.path()).join("config/language.yaml");
        let r = set_home(home.path(), "pt-BR").unwrap();
        assert_eq!(r["written"], true);
        assert_eq!(r["code"], "pt-BR");
        assert_eq!(
            std::fs::read(&file).unwrap(),
            b"language:\n  conversation_language: pt-BR\n"
        );
        assert_eq!(home_conversation(home.path()).as_deref(), Some("pt-BR"));
        assert_eq!(home_conversation_supported(home.path()), None);
    }

    /// Card C-87 S1: the lines name only the items the profile sets; an item it leaves out gets
    /// no line, never a guessed default.
    #[test]
    fn render_lines_lists_only_the_items_the_profile_sets() {
        let s = settings(&[("code_comments", "ko"), ("user_chat", "ja")]);
        let en = render_lines(&s, "en", ".moai/project.yaml");
        // #2696 C2: the heading names where this project keeps its profile.
        assert_eq!(
            en,
            "Writing language by item (`language` in `.moai/project.yaml`):\n\
             - Conversation with the user: Japanese (`ja`)\n\
             - Code comments: Korean (`ko`)\n\
             All other writing (instruction files, skills and rules, messages between agents, code comments, issue and PR records) is in English unless a line above says otherwise.\n"
        );
        let ko = render_lines(&s, "ko", ".moai/project.yaml");
        assert_eq!(
            ko,
            "항목별 작성 언어(`.moai/project.yaml`의 `language`):\n\
             - 사용자 대화: 일본어 (`ja`)\n\
             - 코드 주석: 한국어 (`ko`)\n\
             그 밖의 글(지침 파일, 스킬·규칙, 에이전트 간 메시지, 코드 주석, Issue·PR 기록)은 위에 따로 적지 않으면 영어로 써요.\n"
        );
    }

    /// Each item alone renders exactly its own line, in `ITEMS` order with the others absent, then
    /// the one closing sentence.
    #[test]
    fn each_item_renders_its_own_line() {
        for item in ITEMS {
            let s = settings(&[(item.id, "ko")]);
            let en = render_lines(&s, "en", "p.yaml");
            let lines: Vec<&str> = en.lines().collect();
            // Heading, the item's own line, then the closing sentence about everything else.
            assert_eq!(lines.len(), 3, "{}: {en}", item.id);
            assert_eq!(lines[1], format!("- {}: Korean (`ko`)", item.label_en));
            assert!(
                lines[2].starts_with("All other writing"),
                "{}: {en}",
                item.id
            );
        }
    }
}
