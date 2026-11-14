//! #2894 U-3 and U-6: the human `moai update` screens, drawn from saved `update.run` results
//! (`tests/fixtures/update_tux/*.json`, taken from a temporary HOME and with the paths fixed),
//! compared with `tests/fixtures/ui/{en,ko}/update-*.golden`. Regenerate the goldens with
//! `MOAI_UPDATE_GOLDEN_WRITE=1 cargo test --test update_tux_golden` and review the diff.

use moai_cli::ui::Painter;
use moai_cli::ui_text::Lang;
use moai_cli::update_tux::{question, screen, Asked, Project, View};
use serde_json::Value;
use std::{fs, path::PathBuf};
use unicode_width::UnicodeWidthStr;

fn root() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
}

fn fixture(name: &str) -> Value {
    let p = root().join(format!("tests/fixtures/update_tux/{name}.json"));
    serde_json::from_str(&fs::read_to_string(&p).unwrap()).unwrap()
}

#[derive(Clone, Copy)]
enum Look {
    /// No color, UTF-8 symbols (`MOAI_OUTPUT=plain` on a UTF-8 terminal, `NO_COLOR`).
    Plain,
    /// Color and UTF-8 symbols.
    Color,
    /// No color, ASCII symbols (`TERM=dumb`).
    Ascii,
}

fn view(lang: Lang, look: Look, width: usize, project: Option<Project>) -> View {
    let (color, unicode) = match look {
        Look::Plain => (false, true),
        Look::Color => (true, true),
        Look::Ascii => (false, false),
    };
    View {
        painter: Painter {
            color,
            unicode,
            lang,
        },
        width,
        home: "/home/u".into(),
        project,
    }
}

fn go_project() -> Option<Project> {
    Some(Project {
        path: "/home/u/work/site".into(),
        go_items: Some(47),
    })
}

/// #2872 stage 2 on the screens: Go leftovers the cleanup would archive are a row, and the
/// command that confirms them (its own token, never the `y`) is shown whole; once a cleanup ran,
/// its count and revert are shown and the command is not offered again.
#[test]
fn the_cleanup_is_shown_with_its_own_command() {
    let token = "sha256:".to_string() + &"c".repeat(64);
    let pending = serde_json::json!({"counts": {"archive": 2, "keep": 1, "total": 3},
        "confirm_token": token, "apply_flag": format!("--confirm-cleanup {token}")});
    let mut changes = fixture("preview-changes");
    changes["cleanup"] = pending.clone();
    let mut latest = fixture("preview-latest");
    latest["cleanup"] = pending;
    let command = format!("moai update --confirm-cleanup {token}");
    for lang in [Lang::En, Lang::Ko] {
        let v = view(lang, Look::Plain, 80, None);
        let word = if lang == Lang::Ko {
            "정리(Go 남은 것)"
        } else {
            "cleanup (Go leftovers)"
        };
        let counts = if lang == Lang::Ko {
            "보관 2개 · 유지 1개"
        } else {
            "2 to archive · 1 kept"
        };
        let q = question(&changes, &v);
        assert!(q.contains(word) && q.contains(counts), "{q}");
        for text in [
            screen(&changes, Asked::NotAsked, &v),
            screen(&changes, Asked::No, &v),
            screen(&latest, Asked::NotAsked, &v),
        ] {
            assert!(text.lines().any(|l| l.trim() == command), "{text}");
        }
        let mut done = latest.clone();
        done["cleanup_result"] = serde_json::json!({"state": "committed", "moved": 2,
            "stamp": "20261114T080000Z"});
        done["revert"] = serde_json::json!({
            "cleanup": "moai cleanup-revert-plan <HOME> --stamp 20261114T080000Z"});
        let text = screen(&done, Asked::NotAsked, &v);
        let cleaned = if lang == Lang::Ko {
            "정리했어요 — 보관 2개"
        } else {
            "Cleaned up — archived: 2"
        };
        assert!(text.contains(cleaned), "{text}");
        assert!(text.contains("moai cleanup-revert-plan <HOME> --stamp 20261114T080000Z"));
        assert!(!text.contains("--confirm-cleanup"), "{text}");
    }
}

fn moai_project() -> Option<Project> {
    Some(Project {
        path: "/home/u/work/app".into(),
        go_items: None,
    })
}

/// #2894 7b (U-9): a finished run and "all up to date" always end with the block that sends
/// each existing project to its own session's `/moai update`, inside a project or not, with the
/// project's name and state when there is one; a refusal, a block, a refused stale plan and a
/// preview that waits for its answer do not.
#[test]
fn the_project_update_block_ends_finished_and_up_to_date_screens_only() {
    let changes = fixture("preview-changes");
    let latest = fixture("preview-latest");
    let applied = fixture("applied");
    let blocked = fixture("preview-blocked");
    // A partly applied run (one part failed after the skills were written) ends with what
    // failed, not with the next step.
    let mut partial = applied.clone();
    partial["routing_result"] =
        serde_json::json!({"state": "error", "message": "routing file changed during apply"});
    for (lang, block, codex) in [
        (
            Lang::Ko,
            "각 프로젝트의 세션에서 /moai update",
            "$moai update",
        ),
        (
            Lang::En,
            "run /moai update in a session there",
            "$moai update",
        ),
    ] {
        for project in [None, go_project(), moai_project()] {
            let v = view(lang, Look::Plain, 80, project.clone());
            for text in [
                screen(&applied, Asked::Yes, &v),
                screen(&latest, Asked::NotAsked, &v),
            ] {
                // The sentence wraps to the width; read it with its lines joined (a no-break
                // space holds each command together).
                let joined = flat(&text);
                assert!(joined.contains(block) && joined.contains(codex), "{text}");
                // The block is the end of the screen: the last paragraph is the block.
                let end = flat(text.trim_end().rsplit("\n\n").next().unwrap_or(""));
                assert!(end.contains(block) && end.contains(codex), "{text}");
                match &project {
                    Some(p) if p.go_items.is_some() => assert!(text.contains("site —"), "{text}"),
                    Some(_) => assert!(text.contains("app —"), "{text}"),
                    None => assert!(!text.contains(" — Go") && !text.contains("app —")),
                }
            }
            for text in [
                screen(&changes, Asked::No, &v),
                screen(&changes, Asked::NotAsked, &v),
                screen(&changes, Asked::Stale, &v),
                screen(&blocked, Asked::NotAsked, &v),
                question(&changes, &v),
                screen(&partial, Asked::Yes, &v),
            ] {
                let flat = flat(&text);
                assert!(!flat.contains(block) && !flat.contains(codex), "{text}");
            }
        }
    }
}

/// `text` with wrapped lines joined and no-break spaces read as spaces.
fn flat(text: &str) -> String {
    text.replace("\n  ", " ").replace('\u{a0}', " ")
}

fn check(name: &str, lang: Lang, text: &str, width: usize) {
    let dir = if lang == Lang::Ko { "ko" } else { "en" };
    let path = root().join(format!("tests/fixtures/ui/{dir}/update-{name}.golden"));
    if std::env::var_os("MOAI_UPDATE_GOLDEN_WRITE").is_some() {
        fs::write(&path, text).unwrap();
    }
    let want = fs::read_to_string(&path).unwrap_or_default();
    assert_eq!(
        text,
        want,
        "{} differs; regenerate and review",
        path.display()
    );
    // Every line fits the width it was drawn for (escape codes take no cells) — except the
    // confirming and revert commands, which stay whole so they can be copied.
    for line in text.lines().filter(|l| {
        !l.contains("moai update --confirm sha256:") && !l.contains("-revert <HOME> --receipt")
    }) {
        let shown = strip_ansi(line);
        assert!(
            UnicodeWidthStr::width(shown.as_str()) <= width,
            "{name}: line wider than {width}: {shown:?}"
        );
    }
}

fn strip_ansi(s: &str) -> String {
    let mut out = String::new();
    let mut chars = s.chars().peekable();
    while let Some(c) = chars.next() {
        if c == '\u{1b}' {
            for d in chars.by_ref() {
                if d.is_ascii_alphabetic() {
                    break;
                }
            }
        } else {
            out.push(c);
        }
    }
    out
}

/// U-3: each screen in both languages; the plain look at 80 columns, plus color, ASCII and 40
/// columns for the screens a person meets first.
#[test]
fn the_update_screens_match_their_goldens() {
    let changes = fixture("preview-changes");
    let latest = fixture("preview-latest");
    let applied = fixture("applied");
    let blocked = fixture("preview-blocked");
    for lang in [Lang::Ko, Lang::En] {
        let v = view(lang, Look::Plain, 80, None);
        check("question", lang, &question(&changes, &v), 80);
        check(
            "changes-not-asked",
            lang,
            &screen(&changes, Asked::NotAsked, &v),
            80,
        );
        check("declined", lang, &screen(&changes, Asked::No, &v), 80);
        check("stale", lang, &screen(&changes, Asked::Stale, &v), 80);
        check("latest", lang, &screen(&latest, Asked::NotAsked, &v), 80);
        check("applied", lang, &screen(&applied, Asked::Yes, &v), 80);
        let narrow = view(lang, Look::Plain, 40, None);
        check(
            "blocked-40",
            lang,
            &screen(&blocked, Asked::NotAsked, &narrow),
            40,
        );
        check("question-40", lang, &question(&changes, &narrow), 40);
        let go = view(lang, Look::Plain, 80, go_project());
        check(
            "go-project",
            lang,
            &screen(&latest, Asked::NotAsked, &go),
            80,
        );
        let moai = view(lang, Look::Plain, 80, moai_project());
        check(
            "applied-project",
            lang,
            &screen(&applied, Asked::Yes, &moai),
            80,
        );
    }
    check(
        "question-color",
        Lang::Ko,
        &question(&changes, &view(Lang::Ko, Look::Color, 80, None)),
        80,
    );
    check(
        "question-ascii-40",
        Lang::En,
        &question(&changes, &view(Lang::En, Look::Ascii, 40, None)),
        40,
    );
}

/// U-6: text read from the result (a path, a reason) cannot reach the terminal as a control
/// sequence: an escape in a file name is shown as `\x1b`.
#[test]
fn control_characters_in_the_result_are_shown_not_sent() {
    let mut changes = fixture("preview-changes");
    // A file that changes, so its name is drawn in the change list too.
    let files = changes["plan"]["hosts"][0]["files"].as_array_mut().unwrap();
    let changed = files
        .iter_mut()
        .find(|f| f["action"] != "unchanged")
        .expect("the fixture changes a file");
    changed["path"] = Value::String("SKILL\u{1b}]0;owned\u{7}.md".into());
    changes["hooks"]["skipped"][0]["reason"] = Value::String("bad\u{1b}[2Jreason".into());
    for look in [Look::Plain, Look::Color] {
        let v = view(Lang::En, look, 80, None);
        for text in [
            question(&changes, &v),
            screen(&changes, Asked::NotAsked, &v),
        ] {
            let without_style = strip_ansi(&text);
            assert!(!without_style.contains('\u{1b}'), "{without_style}");
            assert!(!text.contains('\u{7}'));
            assert!(without_style.contains("\\x1b"), "{without_style}");
        }
    }
}

/// The screens never show the JSON envelope or the full digest a person would have to copy.
#[test]
fn no_screen_prints_json_or_asks_to_copy_a_digest_when_it_asks() {
    let changes = fixture("preview-changes");
    let v = view(Lang::En, Look::Plain, 80, None);
    for text in [
        question(&changes, &v),
        screen(&changes, Asked::No, &v),
        screen(&fixture("applied"), Asked::Yes, &v),
    ] {
        assert!(!text.contains("\"protocol_version\""), "{text}");
        assert!(!text.contains("sha256:"), "{text}");
    }
    // Only the not-asked screen (no terminal to answer on) names the confirming command.
    assert!(screen(&changes, Asked::NotAsked, &v).contains("moai update --confirm sha256:"));
}
