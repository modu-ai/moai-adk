//! Per-project writing language by item (Issue #2293): the `language` key of
//! `.moai-cli/project.yaml`, the setup input step that asks for it item by item, and the
//! sentences the AGENTS.md managed section renders from it.
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

/// One writing item: its profile id, its label in English and Korean, and its default.
#[derive(Clone, Copy, Debug)]
pub struct Item {
    pub id: &'static str,
    pub label_en: &'static str,
    pub label_ko: &'static str,
    pub default: DefaultSource,
}

/// Every item the `language` key may name, in question and rendering order.
pub const ITEMS: &[Item] = &[
    Item {
        id: "user_chat",
        label_en: "Conversation with the user",
        label_ko: "사용자 대화",
        default: DefaultSource::Conversation,
    },
    Item {
        id: "reports",
        label_en: "Reports to the user",
        label_ko: "사용자 보고",
        default: DefaultSource::Conversation,
    },
    Item {
        id: "issue_pr_records",
        label_en: "Issue and PR records",
        label_ko: "Issue·PR 기록",
        default: DefaultSource::Conversation,
    },
    Item {
        id: "commit_messages",
        label_en: "Commit messages",
        label_ko: "커밋 메시지",
        default: DefaultSource::Conversation,
    },
    Item {
        id: "agent_messages",
        label_en: "Messages between agents",
        label_ko: "에이전트 간 메시지",
        default: DefaultSource::English,
    },
    Item {
        id: "instruction_files",
        label_en: "Instruction files (AGENTS.md, rules)",
        label_ko: "지침 파일(AGENTS.md, 규칙)",
        default: DefaultSource::English,
    },
    Item {
        id: "skill_instructions",
        label_en: "Skill instructions",
        label_ko: "skill 지침",
        default: DefaultSource::English,
    },
    Item {
        id: "code_comments",
        label_en: "Code comments",
        label_ko: "코드 주석",
        default: DefaultSource::English,
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

/// The managed-section lines for a set `language` key, written in `ui` (`en` or `ko`).
/// Items the profile leaves out are rendered with their default, `ui` standing in for the
/// conversation language.
pub fn render_lines(settings: &LanguageSettings, ui: &str, profile_rel: &str) -> String {
    render_lines_in(settings, ui, ui, profile_rel)
}

/// [`render_lines`] with the label language (`ui`, the language the surrounding text is in)
/// apart from the conversation language that conversation-default items resolve to (#2706).
/// `profile_rel` is where the project keeps `project.yaml` (`state_root::rel`, #2696 C2).
pub fn render_lines_in(
    settings: &LanguageSettings,
    ui: &str,
    conversation: &str,
    profile_rel: &str,
) -> String {
    let mut out = if ui == "ko" {
        format!("항목별 작성 언어(`{profile_rel}`의 `language`):\n")
    } else {
        format!("Writing language by item (`language` in `{profile_rel}`):\n")
    };
    for (item, code) in resolve(settings, conversation) {
        let label = if ui == "ko" {
            item.label_ko
        } else {
            item.label_en
        };
        out.push_str(&format!("- {label}: {} (`{code}`)\n", code_name(&code, ui)));
    }
    out
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

/// The conversation language recorded in `home`'s state folder, when there is one: the
/// `language.conversation_language` of [`HOME_FILE`]. A missing, linked, wrong-typed, oversized,
/// non-UTF-8 or malformed file, and a value that is not an accepted code, give `None` — never an
/// error, so the caller falls back to English.
pub fn home_conversation(home: &std::path::Path) -> Option<String> {
    let root = crate::home_root::path(home);
    let bytes = crate::state_root::read_limited(&root, &HOME_FILE, HOME_FILE_LIMIT)?;
    let text = String::from_utf8(bytes).ok()?;
    crate::go_settings::section_pairs(&text, "language")?
        .into_iter()
        .find(|(key, _)| key == "conversation_language")
        .map(|(_, value)| value)
        .filter(|value| is_code(value))
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
/// followed like every other HOME path; a dangling link is no folder.
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
        return Ok(report(false, Some("exists")));
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

/// The question set a host question tool asks: one question per item with the current
/// value, the value an empty answer keeps (`suggested`) and the accepted codes.
pub fn questions(current: Option<&LanguageSettings>, conversation: &str) -> Value {
    questions_with(current, &LanguageSettings::new(), conversation)
}

/// [`questions`] with the Go backup's item values as the next source of `suggested` (#2817).
fn questions_with(
    current: Option<&LanguageSettings>,
    go: &LanguageSettings,
    conversation: &str,
) -> Value {
    let base = with_go(current, go);
    let options: Vec<Value> = CODES
        .iter()
        .map(|(code, en, ko)| json!({"code": code, "name_en": en, "name_ko": ko}))
        .collect();
    let items: Vec<Value> = ITEMS
        .iter()
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
    json!({"conversation_language": conversation, "items": items, "options": options})
}

/// Applies host-tool answers (`{item: code}`); an item left out keeps its suggested value.
/// The result names every item explicitly, so a re-run with no answers is byte-stable.
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
        let code = match answers.get(i.id).and_then(Value::as_str) {
            Some(code) => code.to_owned(),
            None => suggested(current, i, conversation),
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
/// `language-set` recorded there is the default after `conversation` and the profile's
/// `user_chat`, before the Go backup's and before English.
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
    let profile_chat = current.as_ref().and_then(|c| c.get("user_chat")).is_some();
    let home_chat = home.and_then(home_conversation);
    let conversation = match (conversation, profile_chat, &home_chat, &go.conversation) {
        (None, false, None, Some(code)) => code.clone(),
        _ => conversation_language_in(conversation, current.as_ref(), home_chat.as_deref())?,
    };
    let base = with_go(current.as_ref(), &go.items);
    let base_ref = (!base.is_empty()).then_some(&base);
    let chosen = match mode {
        SetupMode::Questions => {
            let mut out =
                json!({"questions": questions_with(current.as_ref(), &go.items, &conversation)});
            if go.found {
                out["go_backup"] = go.to_json();
            }
            return Ok(out);
        }
        SetupMode::Answers(answers) => merge_answers(base_ref, &answers, &conversation)?,
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
        "next_actions": ["pass result.profile to `moai-cli project-plan <project>` on stdin, review the plan, then `moai-cli project-apply`"]
    });
    if go.found {
        out["go_backup"] = go.to_json();
    }
    Ok(out)
}

/// Upper bound on wrong answers to one question before the interactive step gives up.
pub const MAX_ATTEMPTS: usize = 3;

/// The interactive input step: one prompt per item on `prompt`, one answer line per item
/// from `input`. An empty line or end of input keeps the suggested value; a code outside
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
    for i in ITEMS {
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
    fn merge_answers_names_every_item_and_rejects_bad_answers() {
        let current = settings(&[("code_comments", "ja")]);
        let answers = json!({"user_chat": "zh"});
        let out = merge_answers(Some(&current), answers.as_object().unwrap(), "ko").unwrap();
        assert_eq!(out.len(), ITEMS.len());
        assert_eq!(out["user_chat"], "zh");
        assert_eq!(out["code_comments"], "ja");
        assert_eq!(out["reports"], "ko");
        assert_eq!(out["agent_messages"], "en");
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
        let mut prompt = Vec::new();
        let mut input = "ko\n\nxx\nja\n".as_bytes();
        let out = interactive(None, "en", &mut input, &mut prompt).unwrap();
        assert_eq!(out["user_chat"], "ko");
        assert_eq!(out["reports"], "en");
        assert_eq!(out["issue_pr_records"], "ja");
        // End of input keeps the suggested value for every remaining item.
        assert_eq!(out["commit_messages"], "en");
        assert_eq!(out["code_comments"], "en");
        assert!(String::from_utf8(prompt)
            .unwrap()
            .contains("unsupported language code"));
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
        for bad in [
            &b""[..],
            b"\xff\xfe",
            b"language: ko\n",
            b"language:\n  conversation_language: fr\n",
            b"language:\n  conversation_language: KO\n",
            b"language:\n  conversation_language: ko ja\n",
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
        assert!(set_home(bare.path(), "xx").is_err());
        let home = home_with(None);
        for bad in ["", "KO", "kor", "k", "fr", "ko "] {
            assert!(set_home(home.path(), bad).is_err(), "{bad:?}");
        }
        assert!(!state(home.path()).join("config/language.yaml").exists());
    }

    #[test]
    fn render_lines_lists_every_item_in_the_ui_language() {
        let s = settings(&[("code_comments", "ko")]);
        let en = render_lines(&s, "en", ".moai/project.yaml");
        assert_eq!(en.lines().count(), ITEMS.len() + 1);
        // #2696 C2: the heading names where this project keeps its profile.
        assert!(en.starts_with("Writing language by item (`language` in `.moai/project.yaml`):\n"));
        assert!(en.contains("- Code comments: Korean (`ko`)\n"));
        assert!(en.contains("- Commit messages: English (`en`)\n"));
        let ko = render_lines(&s, "ko", ".moai/project.yaml");
        assert!(ko.starts_with("항목별 작성 언어(`.moai/project.yaml`의 `language`):\n"));
        assert!(ko.contains("- 코드 주석: 한국어 (`ko`)\n"));
        assert!(ko.contains("- 커밋 메시지: 한국어 (`ko`)\n"));
    }
}
