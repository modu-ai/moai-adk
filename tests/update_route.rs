//! #2696: `/moai update` (route UPDATE) is the project layer — managed project files, settings
//! merge and the Go-generation move — while the shell command `moai-cli update` keeps the user
//! layer under HOME (user decision #2634 6845653775). The workflow text is checked against the
//! route table and the built binary, and the project flow it names runs in temporary projects
//! with an empty HOME; the real checkout and the real HOME are never read.

#![cfg(unix)]

mod common;
use common::CleanEnv;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    fs,
    path::{Path, PathBuf},
    process::Command,
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const REF: &str = "test:2696";

fn skill(rel: &str) -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai")
            .join(rel),
    )
    .unwrap()
}

#[test]
fn the_update_route_is_the_project_layer() {
    let routes: Value = serde_json::from_str(&skill("references/routes.json")).unwrap();
    assert_eq!(
        routes["routes"]["UPDATE"]["reference"],
        json!("../workflows/update.md")
    );
    let update = skill("workflows/update.md");
    for text in [
        "`/moai update` (`$moai update` in Codex CLI, `/skill:moai update` in OMP) updates the current project",
        // #2936 S1: one read-only preview of every phase, then the printed commands.
        "Run `moai project-update-plan <project>`. It writes nothing",
        "Show the user every phase and every step in one view",
        "Ask once for the `ready` steps as a whole, then a separate yes for each `needs_decision` step",
        // C-51 (user decision 2026-11-16): the session composes the reference from the pick.
        "the words of `apply.argv` (the first is the `moai` command) with `<permission-reference>` replaced by a permission reference this session composes from the user's pick",
        "and `apply.stdin` on stdin (nothing when it is `null`)",
        "After a go-state apply, run `moai project-update-plan <project>` again before anything else",
        "The project is up to date only when `all_unchanged` is `true`",
        "a re-run plans `unchanged`",
        // The user layer is the shell command; since C-51 (user decision 2026-11-16) this
        // session runs it after the preview and the user's pick.
        "The user layer — the installed skill, the binary and the global instructions under HOME — is the shell command `moai update`",
        "this session runs it for the user (user decision 2026-11-16, card C-51): never tell the user to run it in a terminal",
        // #2702 audit B1: a Go-generation project usually has no profile yet; the move comes
        // first, and "nothing to update" holds only without Go state. C-51: the session goes
        // on to PROJECT itself.
        "when the backup phase has a move, do it first (steps 2–4), then go on to the PROJECT route in this session once the user picks that, and never send the user to type another command",
        "only when there is no profile and no Go-generation state is there nothing to update",
        // #2702 audit N2 (#2953 F1): the other Go-state kinds are reported, never written.
        "When the go-state step reads `needs_decision` with a `reason` naming state_root `interrupted`, a move stopped midway",
        "`state_root.generation` is `mixed`, `unknown` or `refuse` (`refuse` covers a `.moai` that is a link or not a folder)",
        "It is reported with its `reason` and nothing is moved (a link is never followed): say what was found and ask as a selection whether to go on with the steps that do not depend on it (recommended) or stop here",
        // #2953 F6: the confirmed replacement is the step's printed command.
        "the step reads `needs_decision` and its printed command is that replacement (`apply.argv` ends with `--confirm-replace <token>`",
        // #2702 audit N1: the user's edits, unreadable targets and HOLD, and the #2383 step.
        "A file or managed part the user edited is planned as `conflict` and kept",
        "Never write a target whose current state could not be read, nor one under a recorded HOLD",
        "For the Claude Code settings, a key MoAI manages that the user set themselves",
        "`claude-settings-apply … --confirm-replace <token>` only after their explicit confirmation",
        // #2936 R-fix R2 (AC-12): the scope of one yes is a sentence, not an inference.
        "That yes belongs to the preview run the user was shown and covers only the steps that read `ready` in it",
        "A step that read `blocked`, `skip` or `re-run after …` in that run was not shown with its content, so no earlier yes covers it",
        "An answer such as \"continue\" or \"install\" is not an approval of a step the user has not seen",
        "the yes given before the backup phase does not carry over",
        // #2936 R-fix R-A: the preview says what the move takes away before the user agrees.
        "the go-state `items`, `go_backup`, `state_moves`, `restores`, `root_files`, `counterparts`, `after_move` and `warnings`",
        "each moved kind reads `reinstalled by <step>`, `user layer after moai update` or `no counterpart \u{2014} will be gone`",
        "`unknown: true` means a bounded read gave up",
        // #2936 S0 (F6, F16, F24): whether the Go folder moves or stays, and where the backup goes.
        "The go-state preview says in `layout` whether the Go folder moves or stays (`folder_moves`, `folder_stays`, `legacy_into_new` or `nothing`): read it out first",
        "`backup_to` is where the backup would go (the apply fills in `<UTC>`) and is `null` when nothing moves; `go_backup` is `null` unless `layout` is `folder_moves`",
        "A `folder_stays` or `nothing` layout with `backup_to` `null` means no backup is made and no step waits for a confirmation on it",
        // #2936 R-fix R-B (O-1): the hooks step.
        "A `hooks` step reads `ready` when `.claude/settings.json` still runs Go-generation hook scripts",
        "its printed `go-hooks-apply` moves those scripts into a receipt folder, takes the Go entries out of `hooks` and installs the MoAI hook set; `permissions` is not touched",
        "It reads `needs_decision` without a printed command when other tools' hook entries share `hooks`",
    ] {
        assert!(update.contains(text), "update.md lacks {text:?}");
    }
    // The host-install procedure is no longer this route's own steps.
    assert!(
        !update.contains("Identify the MoAI skill, settings, and owner on every installed host"),
        "the user-layer procedure moved out of the project route"
    );
}

fn squeezed(text: &str) -> String {
    text.split_whitespace().collect::<Vec<_>>().join(" ")
}

/// C-51 D1 (user decision 2026-11-16): the session runs every command itself after the
/// preview and the user's pick, the user layer included, and no step hands the user a command
/// or a decision "outside this route".
#[test]
fn the_update_route_runs_the_commands_itself_after_the_pick() {
    let update = skill("workflows/update.md");
    for text in [
        "run the preview `moai update --json` (it writes nothing)",
        "it writes under the real HOME, a `guarded` setting, so nothing is applied without the user's pick",
        "After the pick run `moai update --confirm <plan_digest>` with that preview's `plan_digest`",
        "`moai update --confirm-cleanup <token>` (the preview's `cleanup.confirm_token`",
        "`moai update --confirm-home <token>` (the preview's `home.confirm_token`",
        "one at a time, with a new preview between them",
        "the two cannot be combined in one run",
        "`moai update-apply … --confirm-backup-replace <token>`",
        // D3: what only the user can do is a selection too; deletions wait for the pick.
        "this session cannot close another one",
        "\"I closed them, continue (recommended)\"",
        "after the pick make that one change yourself",
        "Never edit that file without the pick",
        // Addendum 1: the printed order, never a step blocked behind an earlier one.
        "Ask about the steps in the printed order, and never about a step that reads `blocked` or `re-run after …` behind an earlier one",
        "the backup phase lists `go-settings` before `go-state`",
    ] {
        assert!(update.contains(text), "update.md lacks {text:?}");
    }
    let flat = squeezed(&update).to_lowercase();
    for gone in [
        "this route never runs it",
        "tell them to run it in a terminal",
        "let the user decide outside this route",
        "then send the user to setup",
        "send the user to `/moai project`",
        "ask the user to close such sessions",
    ] {
        assert!(!flat.contains(gone), "update.md still says {gone:?}");
    }
}

/// C-51 D2 and C-52 D4: one reference says how a question reaches the user — a selection,
/// the recommended option first, through the host's question tool, with the stated plain-text
/// fallback — and the routes that ask name it instead of repeating it.
#[test]
fn every_question_is_a_selection_through_the_hosts_question_tool() {
    let q = skill("references/internal/questions.md");
    for text in [
        "(user decision 2026-11-16, cards C-51 and C-52)",
        "Never tell the user to run a command in a terminal when this session can run it",
        "Every question is a selection",
        "put the recommended one first and mark it",
        "Never ask an open or free-text question when the options can be listed",
        "A yes-or-no ask is a selection of two options with a recommendation",
        "waits for the user's pick with no default applied",
        "Nothing is deleted unless the user picked exactly that",
        "Use the host's question tool whenever the host lists it",
        "At most three questions per call",
        "`moai language-setup <project> --questions`",
        "never as a bullet list",
    ] {
        assert!(q.contains(text), "questions.md lacks {text:?}");
    }
    // One table row per host: the tool, and the fallback when the host does not list it.
    let row = |host: &str| {
        q.lines()
            .find(|l| l.starts_with(&format!("| {host} |")))
            .unwrap_or_else(|| panic!("no row for {host}"))
            .to_string()
    };
    for (host, texts) in [
        ("Claude Code", &["`AskUserQuestion`", "(Recommended)"][..]),
        (
            "Codex CLI",
            &[
                "`request_user_input`",
                "At most 3 questions per call, 2 or 3 options each",
                "`features.default_mode_request_user_input`",
                "the `moai codex` launcher enables it (card C-52)",
                "Never `request_user_input_async` for a question a step waits on",
                "say once that `/plan` enables the dialog",
                "one question per message with numbered choices",
            ][..],
        ),
        (
            "OMP",
            &[
                "`ask`",
                "one HTML form",
                "`<select>`",
                "`checked`",
                "print mode or a subagent",
                "in the final answer",
            ][..],
        ),
    ] {
        let r = row(host);
        assert_eq!(r.matches(" | ").count(), 3, "{host}: four cells");
        for t in texts {
            assert!(r.contains(t), "{host} row lacks {t:?}: {r}");
        }
    }
    // The routes that ask, and the question bank, name the reference.
    for file in [
        "workflows/update.md",
        "workflows/project.md",
        "workflows/setup.md",
        "references/internal/interview.md",
    ] {
        assert!(
            skill(file).contains("`internal/questions`"),
            "{file} does not name `internal/questions`"
        );
    }
    // The case-(2) sentence of SKILL.md keeps the wait and says it is still a selection.
    let skill_md = skill("SKILL.md");
    assert!(skill_md.contains(
        "goes to the user and waits — no deadline applies it by default; it is still asked as a selection with the recommended option first, and once the user has picked, the session runs it (user decision 2026-11-16, card C-51)"
    ));
    // No route asks three or four at a time any more (the Codex CLI limit is three).
    for file in [
        "workflows/project.md",
        "workflows/setup.md",
        "references/internal/interview.md",
    ] {
        assert!(
            !skill(file).contains("three or four"),
            "{file} still asks three or four at a time"
        );
    }
}

/// C-51 D5 and the 2026-11-16 addendum: the work ledger is local cards, never asked and
/// never a development method; previews run without asking and the pick follows them.
#[test]
fn the_ledger_is_local_cards_and_previews_run_without_asking() {
    let project = skill("workflows/project.md");
    for text in [
        "the previews are read-only and run without asking (user decision 2026-11-16, card C-51)",
        "`go-settings` comes before `go-state`",
        "never ask about a step that reads `blocked` or `re-run after …` behind an earlier step",
        "ask once, after the preview is shown",
        "the development method is TDD or DDD only",
        "the work ledger is the local card ledger (`moai card`): it is not asked, it is never recorded under `development_methods`, and GitHub Issues are not offered as a ledger or a method",
        "relay the `--questions` set through the host's question tool",
        "its `suggested` value first",
    ] {
        assert!(project.contains(text), "project.md lacks {text:?}");
    }
    for gone in [
        "second yes",
        "After a yes, run the read-only preview",
        "Ask the user whether to move the earlier MoAI generation's state aside now",
        "else in the conversation",
    ] {
        assert!(!project.contains(gone), "project.md still says {gone:?}");
    }
    let interview = skill("references/internal/interview.md");
    for text in [
        "The work ledger is not a row",
        "these two only",
        "never recorded under `development_methods`",
    ] {
        assert!(interview.contains(text), "interview.md lacks {text:?}");
    }
    let setup = skill("workflows/setup.md");
    for text in [
        "record the first work item with `moai card add`",
        "do not offer GitHub Issues as the ledger",
        "Only when the user asks to export a work item to a GitHub Issue",
        "never tell the user to run it in a terminal",
    ] {
        assert!(setup.contains(text), "setup.md lacks {text:?}");
    }
    assert!(
        !setup.contains("to answer on the terminal"),
        "setup.md still sends the language questions to the terminal"
    );
    let skill_md = skill("SKILL.md");
    assert!(
        !skill_md.contains("syncs docs and Issues"),
        "the description still names Issues"
    );
    assert!(skill_md.contains("syncs docs and work cards through one shared workflow"));
}

/// #2936 S0 (F6, F16, F24): PROJECT step 2 says the same three things as UPDATE step 2.
#[test]
fn project_step_two_names_the_layout_and_the_backup_destination() {
    let project = skill("workflows/project.md");
    for text in [
        "The preview's `layout` says whether the Go folder moves or stays (`folder_moves`, `folder_stays`, `legacy_into_new` or `nothing`): read it out first",
        "`backup_to` is where the backup would go (the apply fills in `<UTC>`) and is `null` when nothing moves; `go_backup` is `null` unless `layout` is `folder_moves`",
        "A `folder_stays` or `nothing` layout with `backup_to` `null` means no backup is made and no step waits for a confirmation on it",
    ] {
        assert!(project.contains(text), "project.md lacks {text:?}");
    }
}

/// Every `` `moai <op> … --flag` `` span in update.md (and the `` `moai-cli <op>` `` spelling
/// kept for `moai-cli update`, #2792 R3) names an op of this binary with every flag it shows.
#[test]
fn ops_named_by_the_update_route_exist_with_their_flags() {
    let update = skill("workflows/update.md");
    let mut seen = 0;
    let mut starts: Vec<usize> = ["`moai ", "`moai-cli "]
        .iter()
        .flat_map(|p| update.match_indices(p).map(|(i, _)| i))
        .collect();
    starts.sort_unstable();
    for i in starts {
        let rest = &update[i + 1..];
        let span = &rest[..rest.find('`').expect("closed code span")];
        let mut words = span.split_whitespace().skip(1);
        let op = words.next().expect("an op").to_string();
        let flags: Vec<&str> = words
            .filter(|w| w.starts_with("--"))
            .map(|w| w.trim_end_matches(|c: char| !(c.is_ascii_alphanumeric() || c == '-')))
            .collect();
        let out = Command::new(EXE)
            .args([op.as_str(), "--help"])
            .output()
            .unwrap();
        assert!(out.status.success(), "`moai {op}` is not a subcommand");
        let help = String::from_utf8_lossy(&out.stdout);
        // #2702 audit N1: a whole flag word, so `--plan` never passes as a prefix of
        // `--plan-digest`.
        let known: Vec<&str> = help
            .split(|c: char| !(c.is_ascii_alphanumeric() || c == '-'))
            .filter(|w| w.starts_with("--"))
            .collect();
        for flag in flags {
            assert!(known.contains(&flag), "`moai {op}` has no {flag}");
        }
        seen += 1;
    }
    assert!(seen >= 8, "update.md names the project ops ({seen})");
}

// ---- the project flow the route runs (temporary project, empty HOME) ----

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
    root: PathBuf,
}

fn sha(bytes: &[u8]) -> String {
    format!("sha256:{:x}", Sha256::digest(bytes))
}

impl Fx {
    fn new() -> Fx {
        let tmp = TempDir::new().unwrap();
        let base = fs::canonicalize(tmp.path()).unwrap();
        let (home, root) = (base.join("home"), base.join("project"));
        fs::create_dir_all(&home).unwrap();
        fs::create_dir_all(&root).unwrap();
        Fx {
            _tmp: tmp,
            home,
            root,
        }
    }

    fn write(&self, rel: &str, body: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }

    /// A Go-generation project: two files the Go deployer recorded unchanged, one the user
    /// edited after deployment, and a SPEC folder.
    fn go_project(&self) {
        let mut files = serde_json::Map::new();
        for (rel, body) in [
            (".moai/config/sections/a.yaml", "a: 1\n"),
            (".claude/workflows/w.md", "# w\n"),
            (".moai/config/sections/b.yaml", "b: 1\n"),
        ] {
            self.write(rel, body);
            let h = sha(body.as_bytes());
            files.insert(
                rel.into(),
                json!({"provenance": "template_managed", "template_hash": h,
                       "deployed_hash": h, "current_hash": h}),
            );
        }
        self.write(".moai/config/sections/b.yaml", "b: 2 # edited\n");
        self.write(".moai/specs/SPEC-1/spec.md", "# spec\n");
        self.write(
            ".moai/manifest.json",
            &json!({"version": "v3.2.1-rc.6", "deployed_at": "2026-11-01T00:00:00Z",
                    "files": files})
            .to_string(),
        );
    }

    fn call(&self, args: &[&str]) -> Value {
        let out = Command::new(EXE)
            .args(args)
            .clean_env()
            .env("HOME", &self.home)
            .env("PATH", "/usr/bin:/bin")
            .env("LC_ALL", "C")
            .output()
            .unwrap();
        let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: {}{}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        });
        assert_eq!(v["ok"], json!(true), "{args:?}: {v}");
        v["result"].clone()
    }

    fn p(&self) -> &str {
        self.root.to_str().unwrap()
    }

    /// Every file under the project with its bytes, the backup area left out.
    fn tree(&self) -> Vec<(PathBuf, Vec<u8>)> {
        let backups = self.root.join(".moai-cli/backups");
        let mut out = Vec::new();
        let mut stack = vec![self.root.clone()];
        while let Some(d) = stack.pop() {
            for e in fs::read_dir(&d).unwrap() {
                let p = e.unwrap().path();
                if p.starts_with(&backups) {
                    continue;
                }
                if p.is_dir() {
                    stack.push(p);
                } else {
                    out.push((p.clone(), fs::read(&p).unwrap()));
                }
            }
        }
        out.sort();
        out
    }

    fn home_is_empty(&self) -> bool {
        fs::read_dir(&self.home).unwrap().next().is_none()
    }
}

fn moves(plan: &Value) -> usize {
    plan["items"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|i| i["action"] == "move")
        .count()
}

/// Detect → preview and cancel (nothing written) → apply → a re-run moves nothing → revert
/// brings the project back. HOME is never written.
#[test]
fn the_go_generation_move_runs_preview_cancel_apply_rerun_and_revert() {
    let fx = Fx::new();
    fx.go_project();
    let before = fx.tree();

    let detect = fx.call(&["project-detect", fx.p()]);
    // The case step 1 must not stop on (#2702 B1): no profile yet, Go state present.
    assert_eq!(detect["profile"], Value::Null, "{detect}");
    assert_eq!(detect["go_state"]["moai_dir"], json!("folder"), "{detect}");
    assert_eq!(detect["go_state"]["go_manifest"], json!(true));

    // Preview, then the user says no: nothing moves.
    let plan = fx.call(&["go-state-plan", fx.p()]);
    assert_eq!(moves(&plan), 2, "{plan}");
    assert_eq!(fx.tree(), before, "a preview writes nothing");
    // #2936 S0: the preview says the folder moves and names both backup destinations; the
    // preview `project-update-plan` embeds says the same.
    assert_eq!(plan["layout"], json!("folder_moves"), "{plan}");
    assert_eq!(plan["go_backup"], json!(".moai-backup"), "{plan}");
    assert_eq!(plan["backup_to"], json!(".moai/backups/<UTC>"), "{plan}");
    let embedded = fx.call(&["project-update-plan", fx.p()]);
    // #2936 L2a: `go-settings` is the backup phase's first step; the go-state step is by name.
    let go_state = embedded["phases"][0]["steps"]
        .as_array()
        .and_then(|s| s.iter().find(|s| s["step"] == "go-state"))
        .expect("a go-state step");
    let preview = &go_state["preview"];
    for key in ["layout", "go_backup", "backup_to"] {
        assert_eq!(preview[key], plan[key], "{key}: {preview}");
    }
    assert_eq!(fx.tree(), before, "the embedded preview writes nothing");

    // Apply after the second yes.
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let applied = fx.call(&[
        "go-state-apply",
        fx.p(),
        "--plan-digest",
        &digest,
        "--permission-reference",
        REF,
    ]);
    // #2696: the whole Go `.moai/` is kept as `.moai-backup`; a file the user edited is kept
    // there byte for byte, and the SPECs are copied into the new `.moai/specs`.
    let stamp = applied["backup"].as_str().unwrap().to_string();
    assert_eq!(applied["go_backup"], json!(".moai-backup"), "{applied}");
    assert_eq!(
        fs::read_to_string(fx.root.join(".moai-backup/config/sections/b.yaml")).unwrap(),
        "b: 2 # edited\n",
        "a file the user edited is kept in the backup"
    );
    assert!(fx.root.join(".moai/specs/SPEC-1/spec.md").exists());
    assert!(fx.root.join(".moai/moai.json").is_file());
    assert!(!fx.root.join(".moai/moai-cli.json").exists());

    // A second run of the route finds nothing more to move, and step 2 is not offered again.
    let again = fx.call(&["go-state-plan", fx.p()]);
    assert_eq!(moves(&again), 0, "{again}");
    assert_eq!(again["layout"], json!("nothing"), "{again}");
    assert_eq!(again["go_backup"], Value::Null, "{again}");
    assert_eq!(again["backup_to"], Value::Null, "{again}");
    let detect = fx.call(&["project-detect", fx.p()]);
    assert_eq!(
        detect["state_root"]["generation"],
        json!("rust"),
        "{detect}"
    );
    assert!(
        !detect["next_actions"].to_string().contains("go-state-plan"),
        "{detect}"
    );

    // The way back the route reports.
    let rplan = fx.call(&["go-state-revert-plan", fx.p(), "--backup", &stamp]);
    let rdigest = rplan["plan_digest"].as_str().unwrap().to_string();
    fx.call(&[
        "go-state-revert",
        fx.p(),
        "--backup",
        &stamp,
        "--plan-digest",
        &rdigest,
        "--permission-reference",
        REF,
    ]);
    // The new `.moai/` is set aside as `.moai.rust-<UTC>` (kept, never deleted); everything
    // else is as it was.
    let outside_aside = |tree: Vec<(PathBuf, Vec<u8>)>| -> Vec<(PathBuf, Vec<u8>)> {
        tree.into_iter()
            .filter(|(p, _)| {
                !p.strip_prefix(&fx.root)
                    .is_ok_and(|r| r.to_string_lossy().starts_with(".moai.rust-"))
            })
            .collect()
    };
    assert_eq!(
        outside_aside(fx.tree()),
        outside_aside(before),
        "revert brings the project back"
    );
    assert!(fx.home_is_empty(), "the project route never writes HOME");
}

/// A project that is already on this generation's `.moai/` (marker present) has nothing to
/// move; one whose only state is the earlier `.moai-cli/` is offered the move into a new
/// `.moai/` (#2696 C4a), and nothing is written by the preview.
#[test]
fn an_already_rust_project_has_nothing_to_move() {
    let fx = Fx::new();
    fx.write(".moai-cli/specs/SPEC-1/spec.md", "# spec\n");
    let legacy = fx.call(&["go-state-plan", fx.p()]);
    assert_eq!(moves(&legacy), 1, "{legacy}");
    assert_eq!(legacy["layout"], json!("legacy_into_new"), "{legacy}");
    assert_eq!(legacy["go_backup"], Value::Null, "{legacy}");
    assert_eq!(
        legacy["backup_to"],
        json!(".moai/backups/<UTC>"),
        "{legacy}"
    );
    assert_eq!(
        legacy["state_moves"][0]["to"],
        json!(".moai/specs"),
        "{legacy}"
    );
    fx.write(
        ".moai/moai-cli.json",
        r#"{"schema_version":1,"generation":"moai-cli"}"#,
    );
    let before = fx.tree();
    let detect = fx.call(&["project-detect", fx.p()]);
    assert_eq!(
        detect["state_root"]["generation"],
        json!("rust"),
        "{detect}"
    );
    let plan = fx.call(&["go-state-plan", fx.p()]);
    assert_eq!(moves(&plan), 0, "{plan}");
    assert_eq!(plan["layout"], json!("nothing"), "{plan}");
    assert_eq!(plan["backup_to"], Value::Null, "{plan}");
    assert_eq!(fx.tree(), before);
    assert!(fx.home_is_empty());
}
