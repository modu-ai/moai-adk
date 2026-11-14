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
        "`/moai update` (`$moai update` in Codex CLI) updates the current project",
        "moai go-state-plan",
        "moai project-plan",
        "moai agents-plan",
        "moai generation-plan",
        "a re-run plans `unchanged`",
        // The user layer is the shell command, named and never run from this route.
        "The user layer — the installed skill, the binary and the global instructions under HOME — is the shell command `moai update`",
        "this route never runs it",
        // #2702 audit B1: a Go-generation project usually has no profile yet; the move comes
        // first, and "nothing to update" holds only without Go state.
        // #2696: after the move `.moai/` is this generation's folder too, so the route follows
        // the classification, not the folder kind.
        "When `result.profile` is `null` but `result.state_root.generation` is `go` or `interrupted`, do step 2 first",
        "only when there is no profile and no Go-generation state is there nothing to update",
        // #2702 audit N2: the other Go-state kinds are reported, never written.
        "which covers a `.moai` that is a link or not a folder) is reported with its `reason` and nothing is moved",
        "When it is `interrupted`, a move stopped midway",
        // #2702 audit N1: the user's edits, unreadable targets and HOLD, and the #2383 step.
        "A file or managed part the user edited is planned as `conflict` and kept",
        "Never write a target whose current state could not be read, nor one under a recorded HOLD",
        "`moai claude-settings-plan <project>`",
        "`claude-settings-apply … --confirm-replace <token>` only after their explicit confirmation",
    ] {
        assert!(update.contains(text), "update.md lacks {text:?}");
    }
    // The host-install procedure is no longer this route's own steps.
    assert!(
        !update.contains("Identify the MoAI skill, settings, and owner on every installed host"),
        "the user-layer procedure moved out of the project route"
    );
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
    assert!(fx.root.join(".moai/moai-cli.json").is_file());

    // A second run of the route finds nothing more to move, and step 2 is not offered again.
    let again = fx.call(&["go-state-plan", fx.p()]);
    assert_eq!(moves(&again), 0, "{again}");
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
    assert_eq!(fx.tree(), before);
    assert!(fx.home_is_empty());
}
