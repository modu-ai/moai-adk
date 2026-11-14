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

/// #2871 I2 S3 on the screens: each legacy key file is one plain line — provider, file and what
/// the keys layer does — in English and Korean; the import is a change, the applied run shows
/// the `keys-revert` command, and no value appears.
#[test]
fn the_keys_rows_are_one_plain_line_each() {
    let keys = serde_json::json!({
        "rows": [
            {"provider": "glm", "source": "/home/u/.moai/.env.glm", "action": "import", "reason": null},
            {"provider": "glm", "source": "/home/u/.moai-cli/.env.glm", "action": "keep",
             "reason": "another glm file is imported"},
            {"provider": "openai", "source": "/home/u/.moai/.env.openai", "action": "keep",
             "reason": "unknown provider: only glm is imported"},
        ],
        "destination": "/home/u/.moai-cli/vault/_shared/glm.key",
        "plan_digest": "sha256:0", "value_shown": false,
    });
    let mut changes = fixture("preview-changes");
    changes["keys"] = keys;
    for lang in [Lang::En, Lang::Ko] {
        let v = view(lang, Look::Plain, 100, None);
        let (import, keep) = if lang == Lang::Ko {
            ("금고로 가져오기(원본 유지)", "그대로 둠")
        } else {
            ("import into the vault (original kept)", "keep")
        };
        let q = question(&changes, &v);
        for want in [
            format!("glm .moai/.env.glm — {import}"),
            format!("glm .moai-cli/.env.glm — {keep}: another glm file is imported"),
            format!("openai .moai/.env.openai — {keep}: unknown provider: only glm is imported"),
        ] {
            let lines: Vec<&str> = q.lines().filter(|l| l.contains(&want)).collect();
            assert_eq!(lines.len(), 1, "{want}: {q}");
        }
        assert!(q.contains("~/.moai-cli/vault/_shared/glm.key"), "{q}");
        let mut done = fixture("preview-changes");
        done["mode"] = serde_json::json!("applied");
        done["keys_result"] = serde_json::json!({"state": "applied",
            "written": ["/home/u/.moai-cli/vault/_shared/glm.key"], "receipt": "keys-1-2"});
        done["revert"] = serde_json::json!({
            "keys": "moai keys-revert <HOME> --receipt keys-1-2 --permission-reference <reference>"});
        let text = screen(&done, Asked::Yes, &v);
        assert!(
            text.contains("moai keys-revert <HOME> --receipt keys-1-2"),
            "{text}"
        );
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
    // A file ends with one newline (`git diff --check`, AUDIT #2908 B2): a screen's trailing
    // blank lines are stored and compared as one newline; the question's own blank line is
    // asserted in `goldens_end_with_one_newline_and_the_question_keeps_its_blank_line`.
    let stored = format!("{}\n", text.trim_end_matches('\n'));
    if std::env::var_os("MOAI_UPDATE_GOLDEN_WRITE").is_some() {
        fs::write(&path, &stored).unwrap();
    }
    let want = fs::read_to_string(&path).unwrap_or_default();
    assert_eq!(
        stored,
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

/// PR #2908 review (Codex P2 4572504797, 4572504811): a skill host the apply left as it was
/// (`unchanged`) or replaced after a backup (`replaced`) is not a failure, and a host's count is
/// the files it published, not every file the result lists.
#[test]
fn the_applied_screen_counts_published_files_and_keeps_healthy_hosts_ok() {
    let mut applied = fixture("applied");
    let hosts = applied["result"]["hosts"].as_array_mut().unwrap();
    // claude: nothing to write. codex: one file of 31 published by a backup-then-replace.
    hosts[0]["status"] = "unchanged".into();
    for f in hosts[0]["files"].as_array_mut().unwrap() {
        f["published"] = false.into();
    }
    hosts[1]["status"] = "replaced".into();
    for (i, f) in hosts[1]["files"]
        .as_array_mut()
        .unwrap()
        .iter_mut()
        .enumerate()
    {
        f["published"] = (i == 0).into();
    }
    let v = view(Lang::En, Look::Plain, 80, None);
    let text = screen(&applied, Asked::Yes, &v);
    assert!(
        !text.contains("Partly applied") && !text.contains("blocked"),
        "{text}"
    );
    // 1 skill file + 15 agents + 2 routing + 3 style + 1 command name.
    assert!(text.contains("✓ Done — 22 changed"), "{text}");
    assert!(
        text.lines()
            .any(|l| l.contains("skills codex") && l.contains(" 1 changed")),
        "{text}"
    );
    assert!(
        !text
            .lines()
            .any(|l| l.contains("skills claude") && l.contains("changed")),
        "{text}"
    );
    // The project step still ends a run whose every change succeeded.
    assert!(
        flat(&text).contains("run /moai update in a session there"),
        "{text}"
    );
    // A host that failed is still a partial run, with the error its files carry.
    hosts_failed(&mut applied);
    let text = screen(&applied, Asked::Yes, &v);
    assert!(
        text.contains("Partly applied") && text.contains("disk full"),
        "{text}"
    );
}

/// AUDIT #2908 B1: a confirmed cleanup that failed — before any move or after some (`update_cleanup`
/// stops with "stopped after n moved, the receipt … stays pending") — is an error on the screen
/// with its message, never "Cleaned up ✓", and the run ends as partly applied.
#[test]
fn a_failed_cleanup_is_shown_as_failed_with_its_message() {
    for message in [
        "cleanup token changed after the layers were applied; preview again",
        "a target changed during the move: stopped after 2 moved, the receipt 20261114T080000Z stays pending",
    ] {
        let mut applied = fixture("applied");
        applied["cleanup_result"] = serde_json::json!({"state": "error", "message": message});
        for lang in [Lang::En, Lang::Ko] {
            let text = screen(&applied, Asked::Yes, &view(lang, Look::Plain, 80, None));
            let flat_text = flat(&text);
            let (cleaned, partial) = if lang == Lang::Ko {
                ("정리했어요", "일부만 바꿨어요")
            } else {
                ("Cleaned up", "Partly applied")
            };
            assert!(!text.contains(cleaned), "{text}");
            assert!(text.contains(partial), "{text}");
            assert!(flat_text.contains(&message[..40]), "{text}");
            assert!(text.lines().any(|l| l.starts_with("✗ ")), "{text}");
        }
        // The not-asked screen that shows a cleanup result also shows the failure.
        let mut latest = fixture("preview-latest");
        latest["cleanup_result"] = serde_json::json!({"state": "error", "message": message});
        let text = screen(&latest, Asked::NotAsked, &view(Lang::En, Look::Plain, 80, None));
        assert!(!text.contains("Cleaned up") && flat(&text).contains(&message[..40]), "{text}");
    }
}

/// AUDIT #2908 B2: a golden file never ends with a blank line (`git diff --check`); the question
/// screen still ends with the blank line that sets the `[y/N]` prompt apart.
#[test]
fn goldens_end_with_one_newline_and_the_question_keeps_its_blank_line() {
    // The named goldens this file writes (no tree walk: tools/repo_scans.json rule).
    const NAMES: &[&str] = &[
        "applied-project",
        "applied",
        "blocked-40",
        "changes-not-asked",
        "declined",
        "go-project",
        "latest",
        "question-40",
        "question-ascii-40",
        "question-color",
        "question",
        "stale",
    ];
    for dir in ["en", "ko"] {
        for name in NAMES {
            let p = root().join(format!("tests/fixtures/ui/{dir}/update-{name}.golden"));
            // Not every golden exists in both languages (color is ko only, ascii en only).
            let Ok(text) = fs::read_to_string(&p) else {
                continue;
            };
            assert!(
                !text.ends_with("\n\n"),
                "{} ends with a blank line",
                p.display()
            );
        }
    }
    let q = question(
        &fixture("preview-changes"),
        &view(Lang::En, Look::Plain, 80, None),
    );
    assert!(q.ends_with("\n\n"), "{q:?}");
}

fn hosts_failed(applied: &mut Value) {
    let h = &mut applied["result"]["hosts"][0];
    h["status"] = "failed".into();
    h["blocked_reason"] = Value::Null;
    h["files"][0]["error"] = "disk full".into();
}

/// PR #2908 review (Codex P2 4572504814): the end of a confirmed run lists every layer's revert,
/// output style included, in the reverse of the order the layers were applied (agents, routing,
/// style, command names, hooks), so a file two layers changed is restored last-first.
#[test]
fn the_applied_screen_lists_reverts_last_applied_first_with_the_style() {
    let mut applied = fixture("applied");
    applied["revert"]["hooks"] =
        "moai hooks-revert <PROJECT> --receipt hooks-1 --permission-reference <reference>".into();
    let text = screen(&applied, Asked::Yes, &view(Lang::En, Look::Plain, 80, None));
    let order: Vec<&str> = text
        .lines()
        .filter_map(|l| l.trim().strip_prefix("revert: moai "))
        .map(|l| l.split(' ').next().unwrap_or(""))
        .collect();
    assert_eq!(
        order,
        [
            "hooks-revert",
            "home-style-revert",
            "home-routing-revert",
            "agents-home-revert"
        ],
        "{text}"
    );
}

/// PR #2908 review (Codex P2 4572504820, 4572504824): routing that is appended to an existing
/// file or merged into its block is a change, never "skipped"; a blocked output-style target is an
/// error row with its path and reason, so the preview is not "everything is up to date".
#[test]
fn the_preview_shows_routing_appends_and_blocked_style_targets() {
    let v = view(Lang::En, Look::Plain, 80, None);
    for (action, word) in [("append", "new"), ("merge", "to update")] {
        let mut p = fixture("preview-latest");
        p["routing"]["hosts"][0]["section"]["action"] = action.into();
        let q = question(&p, &v);
        let line = q.lines().find(|l| l.contains("routing")).unwrap_or("");
        assert!(
            line.contains(word) && !line.contains("skipped"),
            "{action}: {q}"
        );
    }
    let mut p = fixture("preview-latest");
    let t = &mut p["style"]["targets"][0];
    t["action"] = "blocked".into();
    t["detail"] = "is a symlink; refusing to write through it".into();
    let path = t["path"].as_str().unwrap().to_string();
    let text = screen(&p, Asked::NotAsked, &v);
    let flat_text = flat(&text);
    assert!(!text.contains("Everything is up to date"), "{text}");
    // The reason wraps to the width; its start stays on the row.
    assert!(
        flat_text.contains("blocked: is a symlink; refusing"),
        "{text}"
    );
    assert!(text.contains("Nothing can be applied"), "{text}");
    let name = path.rsplit('/').next().unwrap();
    assert!(flat_text.contains(name), "{text}");
}
