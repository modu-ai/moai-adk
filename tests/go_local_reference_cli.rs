//! SPEC-20261126-go-local-settings-reference: a Go-generation project whose local settings name
//! its Go folder gets one pick, `--rewrite-local-references`, that backs the file up, points the
//! references at the Go backup folder and moves the folder. Every case builds a temporary
//! project inside a temporary HOME; the real HOME and checkout are never read. The state folder
//! names come from the resolver constants, never from a literal.

#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::state_root::{GO_BACKUP, MOAI_DIR};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    fs,
    io::Write as _,
    os::unix::fs::{symlink, PermissionsExt},
    path::{Path, PathBuf},
    process::Stdio,
};

const PICK: &str = "--rewrite-local-references";
const LOCAL: &str = ".claude/settings.local.json";
const SHARED: &str = ".claude/settings.json";
/// A proven `.claude/` file that moves even when the Go folder stays: a refusal must keep it.
const FREE: &str = ".claude/workflows/free.md";
const REF: &str = "test-pick";
/// A value of the local settings that stands for a token: no command output may hold it.
const KEPT_VALUE: &str = "fixture-value-that-stays-private";

fn sha(bytes: &[u8]) -> String {
    format!("sha256:{:x}", Sha256::digest(bytes))
}

/// `rel` inside the folder named `base`.
fn under(base: &str, rel: &str) -> String {
    format!("{base}/{rel}")
}

/// `rel` inside the Go folder.
fn go(rel: &str) -> String {
    under(MOAI_DIR, rel)
}

#[derive(Debug, PartialEq, Eq)]
enum Node {
    Dir,
    Link(PathBuf),
    File(Vec<u8>, u32),
}

struct Fx {
    home: TempHome,
    root: PathBuf,
    files: serde_json::Map<String, Value>,
}

impl Fx {
    /// A Go project: a config file and four hook scripts under the Go folder, one proven
    /// `.claude/` file outside it, and the Go manifest that records them.
    fn new() -> Fx {
        let home = TempHome::new();
        let root = fs::canonicalize(home.path()).unwrap().join("work/demo");
        fs::create_dir_all(&root).unwrap();
        let mut fx = Fx {
            home,
            root,
            files: serde_json::Map::new(),
        };
        fx.deployed(&go("config/sections/a.yaml"), "a: 1\n");
        fx.deployed(&go("status_line.sh"), "#!/bin/sh\necho line\n");
        for name in ["a", "b", "c", "d"] {
            fx.deployed(&go(&format!("hooks/{name}.sh")), "#!/bin/sh\n");
        }
        fx.deployed(FREE, "# free\n");
        fx.manifest();
        fx
    }

    fn write(&self, rel: &str, body: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }

    fn deployed(&mut self, rel: &str, body: &str) {
        self.write(rel, body);
        let h = sha(body.as_bytes());
        self.files.insert(
            rel.into(),
            json!({"provenance": "template_managed", "template_hash": h,
                   "deployed_hash": h, "current_hash": h}),
        );
    }

    fn manifest(&self) {
        self.write(
            &go("manifest.json"),
            &json!({"version": "v3.2.1-rc.6", "deployed_at": "2026-11-01T00:00:00Z",
                    "files": self.files})
            .to_string(),
        );
    }

    /// The local settings, owner-only as a file that may hold a token is.
    fn local(&self, text: &str) {
        self.write(LOCAL, text);
        self.chmod(LOCAL, 0o600);
    }

    fn chmod(&self, rel: &str, mode: u32) {
        fs::set_permissions(self.root.join(rel), fs::Permissions::from_mode(mode)).unwrap();
    }

    fn mode(&self, rel: &str) -> u32 {
        fs::metadata(self.root.join(rel))
            .unwrap()
            .permissions()
            .mode()
            & 0o777
    }

    fn read(&self, rel: &str) -> String {
        fs::read_to_string(self.root.join(rel)).unwrap()
    }

    fn shown(&self) -> &str {
        self.root.to_str().unwrap()
    }

    /// Every entry of the project: kind, bytes and mode, links never followed.
    fn tree(&self) -> Vec<(PathBuf, Node)> {
        let mut out = Vec::new();
        let mut stack = vec![self.root.clone()];
        while let Some(d) = stack.pop() {
            for e in fs::read_dir(&d).unwrap() {
                let p = e.unwrap().path();
                let m = fs::symlink_metadata(&p).unwrap();
                if m.file_type().is_symlink() {
                    out.push((p.clone(), Node::Link(fs::read_link(&p).unwrap())));
                } else if m.is_dir() {
                    out.push((p.clone(), Node::Dir));
                    stack.push(p);
                } else {
                    let mode = m.permissions().mode() & 0o777;
                    out.push((p.clone(), Node::File(fs::read(&p).unwrap(), mode)));
                }
            }
        }
        out.sort_by(|a, b| a.0.cmp(&b.0));
        out
    }

    /// One command's JSON envelope and everything it printed.
    fn run(&self, args: &[&str], stdin: &[u8], envs: &[(&str, &str)]) -> (Value, String) {
        let mut child = self
            .home
            .command(args)
            .env("PATH", "/usr/bin:/bin")
            .env("LC_ALL", "C")
            .envs(envs.iter().copied())
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .expect("run moai");
        // A command that reads no input may have exited already.
        let _ = child.stdin.take().unwrap().write_all(stdin);
        let out = child.wait_with_output().unwrap();
        let text = format!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
        let v = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| panic!("{args:?}: {text}"));
        (v, text)
    }

    fn ok(&self, args: &[&str], stdin: &[u8]) -> Value {
        let (v, text) = self.run(args, stdin, &[]);
        assert_eq!(v["ok"], json!(true), "{args:?}: {text}");
        v["result"].clone()
    }

    /// `go-state-plan`, with or without the pick; the preview writes nothing.
    fn plan(&self, pick: bool) -> Value {
        let before = self.tree();
        let mut args = vec!["go-state-plan", self.shown()];
        if pick {
            args.push(PICK);
        }
        let plan = self.ok(&args, b"");
        assert_eq!(self.tree(), before, "a preview writes nothing");
        plan
    }

    /// `go-state-apply` with the digest of `plan`, under the fault point `fault`.
    fn apply(&self, plan: &Value, pick: bool, fault: Option<&str>) -> (Value, String) {
        let digest = plan["plan_digest"].as_str().unwrap().to_string();
        let mut args = vec![
            "go-state-apply",
            self.shown(),
            "--plan-digest",
            &digest,
            "--permission-reference",
            REF,
        ];
        if pick {
            args.push(PICK);
        }
        let envs: Vec<(&str, &str)> = fault.map(|f| ("MOAI_TEST_FAULT", f)).into_iter().collect();
        self.run(&args, b"", &envs)
    }

    /// The picked preview applied; the receipt.
    fn applied(&self) -> Value {
        let plan = self.plan(true);
        let (v, text) = self.apply(&plan, true, None);
        assert_eq!(v["ok"], json!(true), "{text}");
        assert_eq!(v["result"]["state"], json!("committed"), "{text}");
        assert!(!text.contains(KEPT_VALUE), "{text}");
        v["result"].clone()
    }

    /// The picked apply must refuse and leave the project byte for byte; its message.
    fn refuses(&self, plan: &Value) -> String {
        let before = self.tree();
        let (v, text) = self.apply(plan, true, None);
        assert_eq!(v["ok"], json!(false), "{text}");
        assert_eq!(self.tree(), before, "a refusal writes nothing: {text}");
        v["error"]["message"].as_str().unwrap().to_string()
    }

    fn revert_plan(&self, backup: &str) -> Value {
        self.ok(
            &["go-state-revert-plan", self.shown(), "--backup", backup],
            b"",
        )
    }

    fn revert(&self, backup: &str, plan: &Value) -> (Value, String) {
        let digest = plan["plan_digest"].as_str().unwrap().to_string();
        self.run(
            &[
                "go-state-revert",
                self.shown(),
                "--backup",
                backup,
                "--plan-digest",
                &digest,
                "--permission-reference",
                REF,
            ],
            b"",
            &[],
        )
    }
}

fn item<'a>(plan: &'a Value, rel: &str) -> &'a Value {
    plan["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["path"] == rel)
        .unwrap_or_else(|| panic!("{rel} not in plan: {plan}"))
}

fn references(plan: &Value) -> Vec<(String, String, String)> {
    plan["local_references"]["references"]
        .as_array()
        .unwrap_or_else(|| panic!("no local_references: {plan}"))
        .iter()
        .map(|r| {
            let s = |k: &str| r[k].as_str().unwrap().to_string();
            (s("pointer"), s("path"), s("class"))
        })
        .collect()
}

fn blocked(plan: &Value) -> String {
    plan["local_references"]["blocked"].to_string()
}

/// A status line that names one script of the Go folder, with `command` as the whole value.
fn status(command: &str) -> String {
    json!({"statusLine": {"type": "command", "command": command}}).to_string()
}

/// Every project form: relative, `./`, the project folder variable in its quotings, and the
/// absolute project root; in a status line, hooks and permission rules.
fn all_forms(fx: &Fx) -> String {
    let root = fx.shown();
    format!(
        r#"{{
  "env": {{"EXAMPLE_TOKEN": "{KEPT_VALUE}"}},
  "statusLine": {{"type": "command", "command": "bash {m}/status_line.sh"}},
  "hooks": {{"Stop": [{{"hooks": [
    {{"type": "command", "command": "\"$CLAUDE_PROJECT_DIR\"/{m}/hooks/a.sh"}},
    {{"type": "command", "command": "\"${{CLAUDE_PROJECT_DIR}}/{m}/hooks/b.sh\""}},
    {{"type": "command", "command": "'{root}/{m}/hooks/c.sh' --flag"}},
    {{"type": "command", "command": "./{m}/hooks/d.sh"}}
  ]}}]}},
  "permissions": {{
    "deny": ["Read(./{m}/config/**)", "Read(./.env)"],
    "ask": ["Bash({m}/hooks/a.sh:*)"]
  }}
}}
"#,
        m = MOAI_DIR
    )
}

// ---- AC-1: the preview ----

#[test]
fn plan_lists_each_reference_with_pointer_path_and_class_and_no_other_text() {
    let fx = Fx::new();
    let secret = "fixture-token-must-not-show";
    fx.local(
        &json!({
            "env": {"API_KEY": "fixture-env-value-must-not-show"},
            "statusLine": {"type": "command",
                "command": format!("TOKEN={secret} bash {}", go("status_line.sh"))},
            "permissions": {"deny": [format!("Read(./{})", go("config/**"))]},
        })
        .to_string(),
    );
    for pick in [false, true] {
        let (v, text) = {
            let mut args = vec!["go-state-plan", fx.shown()];
            if pick {
                args.push(PICK);
            }
            fx.run(&args, b"", &[])
        };
        assert_eq!(v["ok"], json!(true), "{text}");
        assert!(!text.contains(secret), "{text}");
        assert!(!text.contains("fixture-env-value-must-not-show"), "{text}");
        assert!(!text.contains("TOKEN="), "{text}");
    }
    let plan = fx.plan(false);
    assert_eq!(plan["local_references"]["file"], json!(LOCAL), "{plan}");
    assert_eq!(plan["local_references"]["pick"], json!(PICK), "{plan}");
    assert_eq!(plan["local_references"]["picked"], json!(false), "{plan}");
    assert_eq!(plan["local_references"]["blocked"], json!([]), "{plan}");
    let mut got = references(&plan);
    got.sort();
    assert_eq!(
        got,
        vec![
            (
                "/permissions/deny/0".to_string(),
                go("config/**"),
                "project".to_string()
            ),
            (
                "/statusLine/command".to_string(),
                go("status_line.sh"),
                "project".to_string()
            ),
        ]
    );
    let folder = item(&plan, MOAI_DIR);
    assert_eq!(folder["action"], json!("leave"), "{plan}");
    let why = folder["reason"].as_str().unwrap();
    assert!(
        why.starts_with(&format!("settings reference {MOAI_DIR}/")),
        "{why}"
    );
    assert!(why.contains(PICK), "{why}");
}

#[test]
fn plan_with_the_pick_shows_the_move_under_another_digest() {
    let fx = Fx::new();
    fx.local(&all_forms(&fx));
    let plain = fx.plan(false);
    let picked = fx.plan(true);
    assert_eq!(plain["layout"], json!("folder_stays"), "{plain}");
    assert_eq!(picked["layout"], json!("folder_moves"), "{picked}");
    assert_eq!(item(&picked, MOAI_DIR)["action"], json!("move"));
    assert_eq!(picked["local_references"]["picked"], json!(true));
    assert_ne!(plain["plan_digest"], picked["plan_digest"]);
    let row = picked["root_files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|r| r["path"] == LOCAL)
        .unwrap();
    assert_eq!(row["action"], json!("rewrite"), "{picked}");
    assert!(
        picked["warnings"].to_string().contains(GO_BACKUP),
        "{picked}"
    );
}

#[test]
fn plan_of_a_project_without_the_hold_says_nothing_new() {
    let fx = Fx::new();
    fx.local(&json!({"env": {"EDITOR": "vi"}}).to_string());
    let plan = fx.plan(false);
    assert!(plan.get("local_references").is_none(), "{plan}");
    assert_eq!(item(&plan, MOAI_DIR)["action"], json!("move"), "{plan}");
    assert!(!plan.to_string().contains(PICK), "{plan}");
}

// ---- AC-2: references to the HOME folder keep holding ----

#[test]
fn home_references_keep_the_folder_and_the_pick_refuses_them() {
    for form in ["~", "$HOME", "${HOME}", "<absolute>"] {
        let fx = Fx::new();
        let home = fs::canonicalize(fx.home.path()).unwrap();
        let prefix = if form == "<absolute>" {
            home.to_str().unwrap().to_string()
        } else {
            form.to_string()
        };
        let path = format!("{prefix}/{}", go("bin"));
        fx.local(&json!({"env": {"PATH": format!("/usr/bin:{path}")}}).to_string());
        let plain = fx.plan(false);
        assert_eq!(plain["layout"], json!("folder_stays"), "{form}: {plain}");
        assert_eq!(references(&plain)[0].2, "other", "{form}: {plain}");
        assert_ne!(blocked(&plain), "[]", "{form}: {plain}");
        let picked = fx.plan(true);
        assert_eq!(picked["layout"], json!("folder_stays"), "{form}: {picked}");
        assert_eq!(picked["plan_digest"], plain["plan_digest"], "{form}");
        let why = fx.refuses(&picked);
        assert!(why.contains("/env/PATH"), "{form}: {why}");
        assert!(fx.root.join(FREE).is_file(), "{form}");
    }
}

// ---- AC-3: the apply ----

#[test]
fn apply_rewrites_only_the_references_keeps_the_rules_and_reaches_a_profile() {
    let fx = Fx::new();
    let before = all_forms(&fx);
    fx.local(&before);
    let receipt = fx.applied();
    let b = GO_BACKUP;
    // The exact bytes: only the folder name changes, and each deny and ask rule keeps its
    // original beside the rewritten one.
    let m = MOAI_DIR;
    let expected = before
        .replace(
            &format!("\"Read(./{m}/config/**)\""),
            &format!("\"Read(./{m}/config/**)\", \"Read(./{b}/config/**)\""),
        )
        .replace(
            &format!("\"Bash({m}/hooks/a.sh:*)\""),
            &format!("\"Bash({m}/hooks/a.sh:*)\", \"Bash({b}/hooks/a.sh:*)\""),
        )
        .replace(
            &format!("bash {m}/status_line.sh"),
            &format!("bash {b}/status_line.sh"),
        )
        .replace(
            &format!("\\\"/{m}/hooks/a.sh"),
            &format!("\\\"/{b}/hooks/a.sh"),
        )
        .replace(&format!("}}/{m}/hooks/b.sh"), &format!("}}/{b}/hooks/b.sh"))
        .replace(&format!("/{m}/hooks/c.sh"), &format!("/{b}/hooks/c.sh"))
        .replace(
            &format!("\"./{m}/hooks/d.sh\""),
            &format!("\"./{b}/hooks/d.sh\""),
        );
    assert_eq!(fx.read(LOCAL), expected);
    assert_eq!(fx.mode(LOCAL), 0o600, "never wider than it was");
    for rel in [
        "status_line.sh",
        "hooks/a.sh",
        "hooks/b.sh",
        "hooks/c.sh",
        "hooks/d.sh",
        "config",
    ] {
        assert!(fx.root.join(under(b, rel)).exists(), "{rel}");
    }
    let stamp = receipt["backup"].as_str().unwrap();
    let copy = go(&format!("backups/{stamp}/root-files/{LOCAL}"));
    assert_eq!(
        fx.read(&copy),
        before,
        "the original bytes are in the backup"
    );
    assert_eq!(fx.mode(&copy), 0o600);
    assert!(fx.root.join(go("moai.json")).is_file());
    // The one pick reached a project the ordinary profile commands finish.
    let detect = fx.ok(&["project-detect", fx.shown()], b"");
    let profile = serde_json::to_vec(&json!({
        "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
        "detected": detect["detected"],
    }))
    .unwrap();
    let plan = fx.ok(&["project-plan", fx.shown()], &profile);
    fx.ok(
        &[
            "project-apply",
            fx.shown(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            REF,
        ],
        &serde_json::to_vec(&plan).unwrap(),
    );
    assert!(fx.root.join(go("project.yaml")).is_file());
    assert!(!fx.root.join(moai_cli::PROJECT_STATE_DIR).exists());
    let update = fx.ok(&["project-update-plan", fx.shown()], b"");
    let go_state = step(&update, "go-state");
    assert_ne!(go_state["state"], json!("blocked"), "{go_state}");
}

#[test]
fn apply_writes_the_stamped_backup_name_when_the_plain_one_is_taken() {
    let fx = Fx::new();
    fx.write(&under(GO_BACKUP, "keep.txt"), "taken\n");
    fx.local(&status(&format!("bash {}", go("status_line.sh"))));
    let receipt = fx.applied();
    let name = receipt["go_backup"].as_str().unwrap();
    assert_ne!(name, GO_BACKUP);
    assert_eq!(
        fx.read(LOCAL),
        status(&format!("bash {name}/status_line.sh"))
    );
    assert!(fx.root.join(under(name, "status_line.sh")).is_file());
}

// ---- AC-4: refusals, the unpicked run and a failure after the first write ----

#[test]
fn refuse_every_form_the_pick_cannot_rewrite() {
    let m = MOAI_DIR;
    let cases: Vec<(&str, String)> = vec![
        ("a nested folder", status(&format!("bash sub/{m}/x.sh"))),
        ("a parent folder", status(&format!("bash ../{m}/x.sh"))),
        (
            "another folder named like it",
            status(&format!("bash my{m}/x.sh")),
        ),
        ("another variable", status(&format!("$OTHER/{m}/x.sh"))),
        (
            "after a folder change",
            status(&format!("cd sub && {m}/x.sh")),
        ),
        (
            "a quote that closes another path",
            status(&format!("\"$HOME/\"\"{m}/x.sh\"")),
        ),
        ("an object key", format!(r#"{{"env": {{"{m}/x": "1"}}}}"#)),
        (
            "an escaped spelling beside a plain one",
            format!(r#"{{"env": {{"A": "{m}/x", "B": "{m}\/y"}}}}"#),
        ),
        ("not an object", format!(r#"["{m}/x"]"#)),
        ("not JSON", format!("{{ {m}/x")),
        (
            "too large",
            format!(
                r#"{{"env": {{"A": "{m}/x", "B": "{}"}}}}"#,
                "x".repeat(1024 * 1024)
            ),
        ),
    ];
    for (name, text) in cases {
        let fx = Fx::new();
        fx.local(&text);
        let picked = fx.plan(true);
        assert_eq!(picked["layout"], json!("folder_stays"), "{name}");
        assert_ne!(blocked(&picked), "[]", "{name}: {picked}");
        let why = fx.refuses(&picked);
        assert!(why.contains(PICK), "{name}: {why}");
        assert!(fx.root.join(FREE).is_file(), "{name}");
    }
}

#[test]
fn refuse_a_link_a_changed_file_another_hold_and_a_pick_with_nothing_to_rewrite() {
    // A link: the preview itself refuses, as without the pick.
    let fx = Fx::new();
    fx.write("elsewhere.json", &status(&go("status_line.sh")));
    fs::create_dir_all(fx.root.join(".claude")).unwrap();
    symlink(fx.root.join("elsewhere.json"), fx.root.join(LOCAL)).unwrap();
    let before = fx.tree();
    let (v, text) = fx.run(&["go-state-plan", fx.shown(), PICK], b"", &[]);
    assert_eq!(v["ok"], json!(false), "{text}");
    let (v, text) = fx.apply(&json!({"plan_digest": "sha256:0"}), true, None);
    assert_eq!(v["ok"], json!(false), "{text}");
    assert_eq!(fx.tree(), before);

    // Changed since the preview.
    let fx = Fx::new();
    fx.local(&status(&go("status_line.sh")));
    let picked = fx.plan(true);
    fx.local(&status(&go("hooks/a.sh")));
    let why = fx.refuses(&picked);
    assert!(why.contains("stale or altered plan"), "{why}");

    // The shared settings hold the folder too.
    let fx = Fx::new();
    fx.local(&status(&go("status_line.sh")));
    fx.write(SHARED, &status(&go("hooks/a.sh")));
    let picked = fx.plan(true);
    assert_eq!(picked["layout"], json!("folder_stays"), "{picked}");
    let why = fx.refuses(&picked);
    assert!(why.contains(SHARED), "{why}");

    // A nested checkout under the Go folder holds it too.
    let fx = Fx::new();
    fx.local(&status(&go("status_line.sh")));
    fx.write(&go("nested/.git/HEAD"), "ref: refs/heads/main\n");
    let picked = fx.plan(true);
    assert_eq!(picked["layout"], json!("folder_stays"), "{picked}");
    fx.refuses(&picked);

    // Nothing to rewrite: the pick is not silently ignored.
    let fx = Fx::new();
    fx.local(&json!({"env": {"EDITOR": "vi"}}).to_string());
    let picked = fx.plan(true);
    let why = fx.refuses(&picked);
    assert!(why.contains(PICK), "{why}");
}

#[test]
fn refuse_nothing_changes_for_the_local_settings_without_the_pick() {
    let fx = Fx::new();
    let text = all_forms(&fx);
    fx.local(&text);
    let plan = fx.plan(false);
    let (v, out) = fx.apply(&plan, false, None);
    assert_eq!(v["ok"], json!(true), "{out}");
    assert_eq!(fx.read(LOCAL), text);
    assert!(
        fx.root.join(go("manifest.json")).is_file(),
        "the folder stays"
    );
    assert!(!fx.root.join(GO_BACKUP).exists());
    // The digest of an unpicked preview does not confirm the picked apply.
    let fx = Fx::new();
    fx.local(&all_forms(&fx));
    let plain = fx.plan(false);
    let why = fx.refuses(&plain);
    assert!(why.contains("stale or altered plan"), "{why}");
}

#[test]
fn refuse_a_failure_before_the_rewrite_leaves_the_original_bytes() {
    for point in ["go_state_before_root_files", "go_state_before_rewrite"] {
        let fx = Fx::new();
        let text = all_forms(&fx);
        fx.local(&text);
        let picked = fx.plan(true);
        let (v, out) = fx.apply(&picked, true, Some(point));
        assert_eq!(v["ok"], json!(false), "{point}: {out}");
        let why = v["error"]["message"].as_str().unwrap();
        assert!(why.contains(point) && why.contains(PICK), "{point}: {why}");
        assert_eq!(fx.read(LOCAL), text, "{point}");
        assert_eq!(fx.mode(LOCAL), 0o600, "{point}");
    }
}

#[test]
fn refuse_a_failure_after_the_rewrite_puts_the_original_bytes_back_and_reverts() {
    let fx = Fx::new();
    let text = all_forms(&fx);
    fx.local(&text);
    let picked = fx.plan(true);
    let (v, out) = fx.apply(&picked, true, Some("go_state_after_rewrite"));
    assert_eq!(v["ok"], json!(false), "{out}");
    let why = v["error"]["message"].as_str().unwrap();
    assert!(
        why.contains("go_state_after_rewrite") && why.contains("original bytes"),
        "{why}"
    );
    assert_eq!(fx.read(LOCAL), text);
    assert_eq!(fx.mode(LOCAL), 0o600);
    let stamp = only_backup(&fx);
    let plan = fx.revert_plan(&stamp);
    let (v, out) = fx.revert(&stamp, &plan);
    assert_eq!(v["ok"], json!(true), "{out}");
    assert_eq!(fx.read(LOCAL), text);
    assert!(fx.root.join(go("status_line.sh")).is_file());
}

// ---- AC-5: the revert ----

/// The one apply stamp under the new state folder's backups.
fn only_backup(fx: &Fx) -> String {
    let names: Vec<String> = fs::read_dir(fx.root.join(go("backups")))
        .unwrap()
        .map(|e| e.unwrap())
        .filter(|e| e.path().is_dir())
        .map(|e| e.file_name().into_string().unwrap())
        .collect();
    assert_eq!(names.len(), 1, "{names:?}");
    names[0].clone()
}

/// The files kept under any `replaced-by-revert/<UTC>/` of the project, by their bytes.
fn set_aside(fx: &Fx) -> Vec<String> {
    fx.tree()
        .into_iter()
        .filter_map(|(p, n)| match n {
            Node::File(bytes, _)
                if p.to_str().unwrap().contains("/replaced-by-revert/") && p.ends_with(LOCAL) =>
            {
                Some(String::from_utf8(bytes).unwrap())
            }
            _ => None,
        })
        .collect()
}

#[test]
fn revert_restores_the_original_bytes_and_keeps_the_later_file() {
    for edit in [None, Some("{\"mine\": true}\n")] {
        let fx = Fx::new();
        let text = all_forms(&fx);
        fx.local(&text);
        let receipt = fx.applied();
        if let Some(edit) = edit {
            fs::write(fx.root.join(LOCAL), edit).unwrap();
        }
        let current = fx.read(LOCAL);
        let stamp = receipt["backup"].as_str().unwrap();
        let plan = fx.revert_plan(stamp);
        assert_eq!(
            revert_action(&plan),
            "set_aside_then_restore_original",
            "{plan}"
        );
        let (v, out) = fx.revert(stamp, &plan);
        assert_eq!(v["ok"], json!(true), "{out}");
        assert!(!out.contains(KEPT_VALUE), "{out}");
        assert_eq!(fx.read(LOCAL), text);
        assert_eq!(fx.mode(LOCAL), 0o600);
        assert_eq!(set_aside(&fx), vec![current]);
        assert!(fx.root.join(go("status_line.sh")).is_file());
        assert!(!fx.root.join(GO_BACKUP).exists());
    }
}

#[test]
fn revert_restores_a_deleted_file_and_refuses_a_link() {
    let fx = Fx::new();
    let text = all_forms(&fx);
    fx.local(&text);
    let receipt = fx.applied();
    let stamp = receipt["backup"].as_str().unwrap();
    fs::remove_file(fx.root.join(LOCAL)).unwrap();
    fx.write("elsewhere.json", "{}\n");
    symlink(fx.root.join("elsewhere.json"), fx.root.join(LOCAL)).unwrap();
    let plan = fx.revert_plan(stamp);
    assert_eq!(revert_action(&plan), "conflict", "{plan}");
    let before = fx.tree();
    let (v, out) = fx.revert(stamp, &plan);
    assert_eq!(v["ok"], json!(false), "{out}");
    assert_eq!(fx.tree(), before);
    fs::remove_file(fx.root.join(LOCAL)).unwrap();
    let plan = fx.revert_plan(stamp);
    assert_eq!(revert_action(&plan), "restore_original", "{plan}");
    let (v, out) = fx.revert(stamp, &plan);
    assert_eq!(v["ok"], json!(true), "{out}");
    assert_eq!(fx.read(LOCAL), text);
    assert_eq!(fx.mode(LOCAL), 0o600);
}

#[test]
fn revert_refuses_when_the_backup_copy_changed() {
    let fx = Fx::new();
    fx.local(&all_forms(&fx));
    let receipt = fx.applied();
    let stamp = receipt["backup"].as_str().unwrap();
    fx.write(&go(&format!("backups/{stamp}/root-files/{LOCAL}")), "{}\n");
    let plan = fx.revert_plan(stamp);
    assert_eq!(revert_action(&plan), "missing", "{plan}");
    let before = fx.tree();
    let (v, out) = fx.revert(stamp, &plan);
    assert_eq!(v["ok"], json!(false), "{out}");
    assert_eq!(fx.tree(), before);
}

fn revert_action(plan: &Value) -> &str {
    item(plan, LOCAL)["action"].as_str().unwrap()
}

// ---- AC-6: the named next step ----

fn step<'a>(plan: &'a Value, name: &str) -> &'a Value {
    plan["phases"]
        .as_array()
        .unwrap()
        .iter()
        .flat_map(|ph| ph["steps"].as_array().unwrap())
        .find(|s| s["step"] == json!(name))
        .unwrap_or_else(|| panic!("no {name} step: {plan}"))
}

#[test]
fn next_step_is_named_by_the_update_plan_and_the_project_workflow() {
    // With other proven files to move, and with nothing at all to move while the folder
    // is held.
    for others_move in [true, false] {
        let fx = Fx::new();
        if !others_move {
            fs::remove_file(fx.root.join(FREE)).unwrap();
        }
        fx.local(&status(&format!("bash {}", go("status_line.sh"))));
        let update = fx.ok(&["project-update-plan", fx.shown()], b"");
        let why = step(&update, "go-state")["reason"].as_str().unwrap();
        assert!(why.contains(PICK), "{others_move}: {why}");
        // A session never edits the file by hand, so the step must not ask for that.
        assert!(
            !why.contains("edit or remove these keys yourself"),
            "{others_move}: {why}"
        );
    }
    let text = fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/workflows/project.md"),
    )
    .unwrap();
    let line = text
        .lines()
        .find(|l| l.contains(PICK))
        .expect("project.md names the pick");
    assert!(line.contains("never edit"), "{line}");
}
