//! SPEC-20261122-settings-write-check: what the project settings writer (`claude-settings-plan` /
//! `-apply`, `hooks-plan` / `-apply` and the status line apply behind them) says and records about
//! a write it would refuse once the checks enforce. Every command is the `moai` binary of this
//! build, run in a cleared environment with a temporary HOME and a temporary project that sit side
//! by side; the only `moai` on its PATH is a fake script, and no real settings or `claude` is read.
//!
//! The shape these tests read, which the product has no constant for yet:
//!
//! - a plan or apply answer: `would_refuse`, a non-empty array, absent when there is no entry;
//! - one entry: `rule` (non-empty rule id) and `key` (dotted key path); an entry about a rewrite
//!   that loses content also has `file` (the project-relative settings file), and names either the
//!   affected top-level `key` or a repeated-key `rule` (its id contains `repeat`);
//! - one observation line: `unix_ms`, `file`, and `would_refuse` (the `rule` and `key` of each
//!   entry, `[]` without one); `observation_recorded: false` in an answer whose line was not written.
#![cfg(unix)]

mod common;

use common::cli::tree_state;
use common::CleanEnv as _;
use moai_cli::spec::{Canonical as _, SortedKeys};
use serde_json::{json, Map, Value};
use std::{
    fs,
    io::Write as _,
    os::unix::fs::symlink,
    path::PathBuf,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const SETTINGS: &str = ".claude/settings.json";
const LOCAL: &str = ".claude/settings.local.json";
/// REQ-9: the settings observation log, under the project's state folder.
const LOG: &str = "receipts/settings-write.jsonl";
const FIELD: &str = "would_refuse";
const RECORDED: &str = "observation_recorded";
const REFERENCE: &str = "settings-write-test";
const SWITCHES: [&str; 2] = [
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS",
    "CLAUDE_CODE_EXPERIMENTAL_CODE_MODE",
];

/// One temporary world: `root/project`, `root/home` and `root/bin` with the fake `moai`.
struct World {
    _dir: TempDir,
    root: PathBuf,
    project: PathBuf,
    home: PathBuf,
}

impl World {
    fn new() -> Self {
        let dir = TempDir::new().unwrap();
        let root = fs::canonicalize(dir.path()).unwrap();
        let w = World {
            project: root.join("project"),
            home: root.join("home"),
            root,
            _dir: dir,
        };
        for d in [&w.project, &w.home, &w.root.join("bin")] {
            fs::create_dir_all(d).unwrap();
        }
        // Answers the two probes of the hook and status line gates as the Rust build does.
        common::write_executable(
            &w.root.join("bin/moai"),
            &format!(
                "#!/bin/sh\ncase \"$1\" in\n  statusline) printf '%s\\n' '{}' ;;\n  *) printf '%s\\n' 'moai 1.2.3 (rust)' ;;\nesac\n",
                moai_cli::status_line::RUST_HELP_LINE
            ),
        );
        w
    }

    /// Runs `moai ARGS` with `input` on stdin; returns the whole envelope.
    fn run(&self, args: &[&str], input: &[u8]) -> Value {
        let mut child = Command::new(common::MOAI_BIN)
            .clean_env()
            .args(args)
            .current_dir(&self.root)
            .env("HOME", &self.home)
            .env("USERPROFILE", &self.home)
            .env(
                "PATH",
                format!("{}:/usr/bin:/bin", self.root.join("bin").display()),
            )
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        let mut pipe = child.stdin.take().unwrap();
        // A command that refuses its arguments may exit before it reads stdin.
        let _ = pipe.write_all(input);
        drop(pipe);
        let out = child.wait_with_output().unwrap();
        serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: not JSON: {}",
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    fn p(&self) -> &str {
        self.project.to_str().unwrap()
    }

    /// `claude-settings-plan` with `extra` options.
    fn plan_envelope(&self, extra: &[&str], request: &Value) -> Value {
        let mut args = vec!["claude-settings-plan", self.p()];
        args.extend_from_slice(extra);
        self.run(&args, request.to_string().as_bytes())
    }

    /// `claude-settings-apply` of `digest` with `extra` options.
    fn apply_envelope(&self, extra: &[&str], request: &Value, digest: &str) -> Value {
        let mut args = vec![
            "claude-settings-apply",
            self.p(),
            "--plan-digest",
            digest,
            "--permission-reference",
            REFERENCE,
        ];
        args.extend_from_slice(extra);
        self.run(&args, request.to_string().as_bytes())
    }

    fn plan(&self, extra: &[&str], request: &Value) -> Value {
        ok(self.plan_envelope(extra, request))
    }

    /// Plans `request` and applies that plan; the apply wrote what the plan announced.
    fn write_through(&self, extra: &[&str], request: &Value) -> (Value, Value) {
        let plan = self.plan(extra, request);
        let applied = ok(self.apply_envelope(extra, request, digest_of(&plan)));
        if !plan["planned_sha256"].is_null() {
            assert_eq!(
                applied["settings_sha256_after"], plan["planned_sha256"],
                "{request}: {applied}"
            );
        }
        (plan, applied)
    }

    fn hooks_plan(&self) -> Value {
        ok(self.run(&["hooks-plan", self.p(), "--host", "claude"], b""))
    }

    fn hooks_apply(&self, plan: &Value) -> Value {
        ok(self.run(
            &[
                "hooks-apply",
                self.p(),
                "--host",
                "claude",
                "--plan-digest",
                digest_of(plan),
                "--permission-reference",
                REFERENCE,
            ],
            b"",
        ))
    }

    fn put(&self, rel: &str, bytes: &[u8]) {
        let path = self.project.join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, bytes).unwrap();
    }

    fn read(&self, rel: &str) -> Vec<u8> {
        fs::read(self.project.join(rel)).unwrap()
    }

    fn log_path(&self) -> PathBuf {
        moai_cli::state_root::path(&self.project, LOG)
    }

    /// The observation lines so far; none when there is no log.
    fn lines(&self) -> Vec<Value> {
        match fs::read_to_string(self.log_path()) {
            Ok(text) => text
                .lines()
                .map(|l| serde_json::from_str(l).unwrap_or_else(|_| panic!("not a line: {l}")))
                .collect(),
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => Vec::new(),
            Err(e) => panic!("the log cannot be read: {e}"),
        }
    }
}

fn ok(envelope: Value) -> Value {
    assert_eq!(envelope["ok"], json!(true), "{envelope}");
    envelope["result"].clone()
}

fn digest_of(plan: &Value) -> &str {
    plan["plan_digest"]
        .as_str()
        .unwrap_or_else(|| panic!("no plan digest: {plan}"))
}

/// The bytes the writer renders today when `request` sets its keys over `before`.
fn rewritten(before: Option<&[u8]>, request: &Value) -> Vec<u8> {
    let mut map: Map<String, Value> =
        before.map_or_else(Map::new, |b| serde_json::from_slice(b).unwrap());
    let request = request.as_object().unwrap();
    let mut keys: Vec<&String> = request.keys().collect();
    keys.sort();
    for key in keys {
        map.insert(key.clone(), request[key].clone());
    }
    pretty(&Value::Object(map))
}

fn pretty(v: &Value) -> Vec<u8> {
    let mut bytes = serde_json::to_vec_pretty(v).unwrap();
    bytes.push(b'\n');
    bytes
}

/// The project-scope plan digest as it is today, for a project without an ownership receipt and a
/// plan without a conflict: an observation must not enter it.
fn todays_digest(request: &Value, before: Option<&[u8]>, after: &[u8]) -> String {
    let tuple = (
        1,
        moai_cli::digest(&SortedKeys.bytes(request)),
        before.map(moai_cli::digest),
        None::<String>,
        Some(moai_cli::digest(after)),
        Vec::<String>::new(),
    );
    moai_cli::digest(&serde_json::to_vec(&tuple).unwrap())
}

/// One write over a fresh project (no receipt), held to today's plan, digest and bytes.
struct Written {
    label: String,
    world: World,
    plan: Value,
    applied: Value,
}

fn write_as_today(before: Option<&[u8]>, request: Value) -> Written {
    let world = World::new();
    if let Some(bytes) = before {
        world.put(SETTINGS, bytes);
    }
    let after = rewritten(before, &request);
    let (plan, applied) = world.write_through(&[], &request);
    let label = request.to_string();
    let action = if before.is_some() { "update" } else { "create" };
    assert_eq!(plan["action"], json!(action), "{label}: {plan}");
    assert_eq!(
        plan["planned_sha256"],
        json!(moai_cli::digest(&after)),
        "{label}: {plan}"
    );
    assert_eq!(
        plan["plan_digest"],
        json!(todays_digest(&request, before, &after)),
        "{label}: {plan}"
    );
    assert_eq!(applied["state"], json!("applied"), "{label}: {applied}");
    assert_eq!(world.read(SETTINGS), after, "{label}");
    Written {
        label,
        world,
        plan,
        applied,
    }
}

/// The `would_refuse` entries of an answer, which must have some.
fn entries(answer: &Value, label: &str) -> Vec<Value> {
    let list = answer
        .get(FIELD)
        .unwrap_or_else(|| panic!("{label}: no `would_refuse` field: {answer}"))
        .as_array()
        .unwrap_or_else(|| panic!("{label}: `would_refuse` is not a list: {answer}"))
        .clone();
    assert!(!list.is_empty(), "{label}: an empty `would_refuse`");
    for entry in &list {
        assert!(
            entry["rule"].as_str().is_some_and(|r| !r.is_empty()),
            "{label}: an entry without a rule: {entry}"
        );
    }
    list
}

fn no_field(answer: &Value, label: &str) {
    assert!(
        answer.get(FIELD).is_none(),
        "{label}: unexpected `would_refuse`: {answer}"
    );
}

/// Whether the entry's key path is `path` or lies below it.
fn under(entry: &Value, path: &str) -> bool {
    entry["key"].as_str().is_some_and(|k| {
        k == path || k.starts_with(&format!("{path}.")) || k.starts_with(&format!("{path}["))
    })
}

/// The plan and the apply list the same entries, all of them about `path`.
fn reported_under(w: &Written, path: &str) {
    let planned = entries(&w.plan, &w.label);
    for entry in &planned {
        assert!(under(entry, path), "{}: not about {path}: {entry}", w.label);
    }
    assert_eq!(entries(&w.applied, &w.label), planned, "{}", w.label);
}

/// `(rule, key)` of every entry of a list.
fn pairs(list: &[Value]) -> Vec<(Value, Value)> {
    list.iter()
        .map(|e| (e["rule"].clone(), e["key"].clone()))
        .collect()
}

// SPEC-20261122-settings-write-check AC-1 (REQ-1)
#[test]
fn invalid_shapes_are_reported_not_refused() {
    let line = |extra: Value| {
        let mut v = json!({"type": "command", "command": "my-status"});
        for (k, val) in extra.as_object().unwrap() {
            v[k] = val.clone();
        }
        v
    };
    let invalid = [
        ("statusLine", json!("my-status")),
        ("statusLine", json!(null)),
        ("statusLine", json!({"command": "my-status"})),
        (
            "statusLine",
            json!({"type": "text", "command": "my-status"}),
        ),
        ("statusLine", json!({"type": "command"})),
        ("statusLine", json!({"type": "command", "command": ""})),
        ("statusLine", json!({"type": "command", "command": 7})),
        ("statusLine", line(json!({"padding": "1"}))),
        ("statusLine", line(json!({"refreshInterval": 0}))),
        ("statusLine", line(json!({"refreshInterval": 0.5}))),
        ("statusLine", line(json!({"refreshInterval": "60"}))),
        ("statusLine", line(json!({"hideVimModeIndicator": "yes"}))),
        ("outputStyle", json!("")),
        ("outputStyle", json!(5)),
        ("outputStyle", json!(true)),
        ("outputStyle", json!(null)),
        ("outputStyle", json!(["Explanatory"])),
        ("outputStyle", json!({"name": "Explanatory"})),
    ];
    let valid = [
        json!({"statusLine": line(json!({}))}),
        json!({"statusLine": line(json!({"padding": 0, "refreshInterval": 1, "hideVimModeIndicator": true}))}),
        json!({"statusLine": line(json!({"padding": 2.5, "refreshInterval": 1.5, "hideVimModeIndicator": false}))}),
        json!({"outputStyle": "Explanatory"}),
    ];
    for request in valid {
        let w = write_as_today(None, request);
        no_field(&w.plan, &w.label);
        no_field(&w.applied, &w.label);
    }
    let written: Vec<(&str, Written)> = invalid
        .into_iter()
        .map(|(key, bad)| (key, write_as_today(None, json!({ key: bad }))))
        .collect();
    for (key, w) in &written {
        reported_under(w, key);
    }
}

// SPEC-20261122-settings-write-check AC-2 (REQ-4)
#[test]
fn untouched_user_values_are_not_reported() {
    let held = json!({
        "statusLine": {"type": "text"},
        "outputStyle": "",
        "env": {"HOME": "/opt/h", "EDITOR": "$EDITOR", SWITCHES[0]: "1"},
        "permissions": {"allow": ["Write(docs/**)", "Bash(git * main)"]},
    });
    let held_bytes = pretty(&held);

    // The user's own values stay out of the answer when another key changes, and when the request
    // names the very values the file holds.
    let mut same_again = held.clone();
    same_again["model"] = json!("opus");
    for request in [json!({"model": "opus"}), same_again] {
        let world = World::new();
        world.put(SETTINGS, &held_bytes);
        let (plan, applied) = world.write_through(&[], &request);
        assert_eq!(plan["action"], json!("update"), "{plan}");
        assert_eq!(
            world.read(SETTINGS),
            rewritten(Some(&held_bytes), &json!({"model": "opus"}))
        );
        no_field(&plan, "a held value");
        no_field(&applied, "a held value");
    }

    // A value MoAI wrote and the next request leaves as it is, inside a key that does change.
    let owned = World::new();
    let first = json!({
        "env": {"HOME": "/opt/h"},
        "permissions": {"allow": ["Write(docs/**)"]},
        "outputStyle": "Explanatory",
    });
    owned.write_through(&[], &first);
    let second = json!({
        "env": {"HOME": "/opt/h", "EDITOR": "vim"},
        "permissions": {"allow": ["Write(docs/**)", "Read(src/**)"]},
        "outputStyle": "Explanatory",
    });
    let (plan, applied) = owned.write_through(&[], &second);
    assert_eq!(applied["state"], json!("applied"), "{applied}");
    no_field(&plan, "an unchanged entry of a changed key");
    no_field(&applied, "an unchanged entry of a changed key");

    // One new listed value beside the held ones: that one only.
    let third = json!({
        "env": {"HOME": "/opt/h", "EDITOR": "vim", "TMPDIR": "/opt/t"},
        "permissions": {"allow": ["Write(docs/**)", "Read(src/**)"]},
        "outputStyle": "Explanatory",
    });
    let added = owned.write_through(&[], &third);
    // A key MoAI owns changed to a misshapen value.
    let mut fourth = third.clone();
    fourth["outputStyle"] = json!("");
    let changed = owned.write_through(&[], &fourth);
    // A misshapen change beside a misshapen value the user holds.
    let beside = write_as_today(
        Some(&pretty(&json!({"statusLine": {"type": "text"}}))),
        json!({"outputStyle": ""}),
    );

    for (label, (plan, applied), key) in [
        ("a new env entry", &added, "env.TMPDIR"),
        ("a changed output style", &changed, "outputStyle"),
        (
            "beside a held status line",
            &(beside.plan.clone(), beside.applied.clone()),
            "outputStyle",
        ),
    ] {
        for answer in [plan, applied] {
            let list = entries(answer, label);
            assert_eq!(list.len(), 1, "{label}: {list:?}");
            assert_eq!(list[0]["key"], json!(key), "{label}");
        }
    }
}

// SPEC-20261122-settings-write-check AC-3 (REQ-2)
#[test]
fn questionable_values_are_reported_not_refused() {
    let mut listed: Vec<(String, Value)> = Vec::new();
    for (list, rule) in [
        ("allow", "Write(docs/**)"),
        ("deny", "NotebookEdit(notes/a.ipynb)"),
        ("ask", "MultiEdit(src/**)"),
        ("allow", "Glob(src/**)"),
        ("deny", "Grep(src/**)"),
        ("allow", "Bash(git * main)"),
        ("allow", "Bash(git -C * status *)"),
    ] {
        listed.push((
            format!("permissions.{list}"),
            json!({"permissions": { list: [rule] }}),
        ));
    }
    let mut env = vec![
        ("EDITOR", "$HOME/bin/vim"),
        ("TOOLS_DIR", "~/tools"),
        ("SEARCH_DIRS", "/opt/a:$b"),
        (SWITCHES[0], "1"),
        (SWITCHES[1], "1"),
        (SWITCHES[0], "true"),
    ];
    for name in [
        "CLAUDE_CONFIG_DIR",
        "CLAUDE_CODE_TMPDIR",
        "HOME",
        "TMPDIR",
        "TMP",
        "TEMP",
        "XDG_CONFIG_HOME",
        "XDG_CACHE_HOME",
    ] {
        env.push((name, "/opt/x"));
    }
    for (name, value) in env {
        listed.push((format!("env.{name}"), json!({"env": { name: value }})));
    }

    let unlisted = [
        json!({"permissions": {"allow": ["Bash(git status *)"]}}),
        json!({"permissions": {"allow": ["Bash(ls *)"]}}),
        json!({"permissions": {"allow": ["Edit(docs/**)"]}}),
        json!({"permissions": {"allow": ["Read(src/**)"]}}),
        json!({"permissions": {"allow": ["Write"]}}),
        json!({"permissions": {"allow": ["Write(*)"]}}),
        json!({"permissions": {"allow": ["Glob(*)"]}}),
        // The wildcard rule is about allow rules.
        json!({"permissions": {"deny": ["Bash(git * main)"]}}),
        json!({"env": {"EDITOR": "vim"}}),
        json!({"env": {"BACKUP_SUFFIX": "a~b"}}),
        json!({"env": {"HOMEPAGE": "/opt/x"}}),
        json!({"env": {"MY_TMPDIR": "/opt/x"}}),
        json!({"env": {"TEMPLATE": "/opt/x"}}),
        json!({"env": {"XDGX": "/opt/x"}}),
        json!({"env": {"MY_XDG_X": "/opt/x"}}),
        json!({"env": { SWITCHES[0]: "0" }}),
        json!({"env": { SWITCHES[1]: "0" }}),
        json!({"env": { SWITCHES[0]: "" }}),
        json!({"env": { SWITCHES[1]: "" }}),
    ];
    for request in unlisted {
        let w = write_as_today(None, request);
        no_field(&w.plan, &w.label);
        no_field(&w.applied, &w.label);
    }
    let written: Vec<(String, Written)> = listed
        .into_iter()
        .map(|(path, request)| (path, write_as_today(None, request)))
        .collect();
    for (path, w) in &written {
        reported_under(w, path);
    }
}

// SPEC-20261122-settings-write-check AC-4 (REQ-3)
#[test]
fn forbidden_keys_are_never_written() {
    let world = World::new();
    world.put(
        SETTINGS,
        b"{\n  \"theme\": \"dark\",\n  \"sandbox\": {\"enabled\": true}\n}\n",
    );
    let project_before = tree_state(&world.project);

    let mut cases: Vec<(Value, String)> = Vec::new();
    for (key, value) in [
        ("attribution", json!({"commit": ""})),
        ("fastModePerSessionOptIn", json!(true)),
        ("includeGitInstructions", json!(false)),
        ("disableAllHooks", json!(true)),
        ("sandbox", json!({"enabled": false})),
        ("hooks", json!({"Stop": []})),
        ("apiKeyHelper", json!("my-helper")),
    ] {
        cases.push((
            json!({ key: value }),
            format!("unmanaged claude settings key refused: {key}"),
        ));
    }
    for mode in ["bypassPermissions", "auto"] {
        cases.push((
            json!({"permissions": {"defaultMode": mode}}),
            "widening modes are refused".into(),
        ));
    }
    for value in [json!(1), json!(true), json!(null), json!({"a": "b"})] {
        cases.push((
            json!({"env": {"EDITOR": value}}),
            "the value is not a string".into(),
        ));
    }
    for list in ["allow", "deny", "ask"] {
        for rule in [json!(""), json!("   "), json!(5), json!(null)] {
            cases.push((
                json!({"permissions": { list: [rule] }}),
                format!("permissions.{list} rules must be non-empty one-line strings"),
            ));
        }
    }

    for (request, why) in cases {
        let stale = format!("sha256:{}", "0".repeat(64));
        for envelope in [
            world.plan_envelope(&[], &request),
            world.apply_envelope(&[], &request, &stale),
        ] {
            assert_eq!(envelope["ok"], json!(false), "{request}: {envelope}");
            let message = envelope["error"]["message"].as_str().unwrap_or_default();
            assert!(message.contains(&why), "{request}: {envelope}");
        }
        assert_eq!(tree_state(&world.project), project_before, "{request}");
    }
}

/// One lossy original: its bytes and the top-level key a respelled number sits under (`None`: the
/// file repeats a key).
const LOSSY: [(&[u8], Option<&str>); 6] = [
    (b"{\"cleanupPeriodDays\": 1E2}\n", Some("cleanupPeriodDays")),
    (b"{\"outer\": {\"ratio\": 1.10}}\n", Some("outer")),
    (
        b"{\"big\": [123456789012345678901234567890]}\n",
        Some("big"),
    ),
    (b"{\"zero\": -0}\n", Some("zero")),
    (b"{\"theme\": \"a\", \"theme\": \"b\"}\n", None),
    (b"{\"outer\": {\"k\": 1, \"k\": 2}}\n", None),
];

/// Some entry of `answer` names `file` and the loss: the top-level key, or the repeated-key rule.
fn names_the_loss(answer: &Value, file: &str, key: Option<&str>, label: &str) {
    let list = entries(answer, label);
    let named = list.iter().any(|e| {
        e["file"] == json!(file)
            && match key {
                Some(key) => e["key"] == json!(key),
                None => e["rule"].as_str().is_some_and(|r| r.contains("repeat")),
            }
    });
    assert!(
        named,
        "{label}: no entry names {file} and {key:?}: {list:?}"
    );
}

// SPEC-20261122-settings-write-check AC-6 (REQ-6)
#[test]
fn a_rewrite_that_would_lose_content_is_reported() {
    let request = json!({"model": "opus"});
    let written: Vec<(Option<&str>, Written)> = LOSSY
        .iter()
        .map(|(before, key)| {
            // The fixture does lose content: rendering it again does not give the same values back.
            let again = pretty(&serde_json::from_slice(before).unwrap());
            let compact = |b: &[u8]| -> Vec<u8> {
                b.iter()
                    .copied()
                    .filter(|c| !c.is_ascii_whitespace())
                    .collect()
            };
            assert_ne!(compact(&again), compact(before));
            (*key, write_as_today(Some(before), request.clone()))
        })
        .collect();

    // The local file is rewritten the same way and named as itself.
    let local = World::new();
    local.put(LOCAL, LOSSY[0].0);
    let local_answers = local.write_through(&["--scope", "local"], &request);
    assert_eq!(local.read(LOCAL), rewritten(Some(LOSSY[0].0), &request));

    // A confirmed replacement rewrites the file too.
    let replaced = World::new();
    let conflicting: &[u8] = b"{\"model\": \"sonnet\", \"ratio\": 1.10}\n";
    replaced.put(SETTINGS, conflicting);
    let conflict = replaced.plan(&[], &request);
    assert_eq!(conflict["action"], json!("conflict"), "{conflict}");
    let token = conflict["confirm_replace"]["confirm_token"]
        .as_str()
        .unwrap()
        .to_string();
    let replacement = ok(replaced.apply_envelope(
        &["--confirm-replace", &token],
        &request,
        digest_of(&conflict),
    ));
    assert_eq!(replacement["state"], json!("replaced"), "{replacement}");
    assert_eq!(
        replaced.read(SETTINGS),
        rewritten(Some(conflicting), &request)
    );

    // The hook writer: the same bytes as over the same values without the loss.
    let hooks: Vec<(Option<&str>, Value, Value)> = [LOSSY[0], LOSSY[4]]
        .into_iter()
        .map(|(before, key)| {
            let (lossy, clean) = (World::new(), World::new());
            lossy.put(SETTINGS, before);
            clean.put(SETTINGS, &pretty(&serde_json::from_slice(before).unwrap()));
            let (plan, clean_plan) = (lossy.hooks_plan(), clean.hooks_plan());
            assert_eq!(plan["planned_sha256"], clean_plan["planned_sha256"]);
            let applied = lossy.hooks_apply(&plan);
            let clean_applied = clean.hooks_apply(&clean_plan);
            assert_eq!(applied["state"], json!("applied"), "{applied}");
            assert_eq!(lossy.read(SETTINGS), clean.read(SETTINGS));
            no_field(&clean_plan, "a hooks plan without a loss");
            no_field(&clean_applied, "a hooks apply without a loss");
            (key, plan, applied)
        })
        .collect();

    // Nothing is rewritten: the request is already in the file.
    for (before, _) in LOSSY {
        let world = World::new();
        let mut text = b"{\"model\": \"opus\", \"kept\": ".to_vec();
        text.extend_from_slice(before);
        text.push(b'}');
        world.put(SETTINGS, &text);
        let (plan, applied) = world.write_through(&[], &request);
        assert_eq!(plan["action"], json!("unchanged"), "{plan}");
        assert_eq!(applied["state"], json!("unchanged"), "{applied}");
        assert_eq!(world.read(SETTINGS), text);
        no_field(&plan, "an unchanged plan");
        no_field(&applied, "an unchanged apply");
    }

    for (key, w) in &written {
        // Nested originals report the outer key.
        for answer in [&w.plan, &w.applied] {
            names_the_loss(answer, SETTINGS, *key, &format!("{key:?}"));
        }
    }
    for answer in [&local_answers.0, &local_answers.1] {
        names_the_loss(answer, LOCAL, LOSSY[0].1, "local");
    }
    names_the_loss(&replacement, SETTINGS, Some("ratio"), "replacement");
    for (key, plan, applied) in &hooks {
        for answer in [plan, applied] {
            names_the_loss(answer, SETTINGS, *key, "hooks");
        }
    }
}

// SPEC-20261122-settings-write-check AC-7 (REQ-7)
#[test]
fn a_valid_write_keeps_every_user_key() {
    let world = World::new();
    let before: &[u8] = r#"{
  "zeta": {"b": 2, "a": 1},
  "permissions": {"deny": ["Read(./.env)"], "allow": ["Bash(ls *)", "Bash(git status *)"]},
  "hooks": {"Stop": [{"hooks": [{"type": "command", "command": "my-check"}]}]},
  "enabledPlugins": {"x@y": false},
  "alpha": [3, 1, 2],
  "env": {"ZED": "1", "ABLE": "2"},
  "note": "한글"
}
"#
    .as_bytes();
    world.put(SETTINGS, before);
    let request = json!({"outputStyle": "Explanatory", "model": "opus"});
    let (_, applied) = world.write_through(&[], &request);
    assert_eq!(applied["state"], json!("applied"), "{applied}");
    // Byte equality covers every value, every list entry and the order of the keys.
    assert_eq!(world.read(SETTINGS), rewritten(Some(before), &request));

    // The user changes a key MoAI wrote.
    let mut map: Map<String, Value> = serde_json::from_slice(&world.read(SETTINGS)).unwrap();
    map.insert("model".into(), json!("sonnet"));
    let edited = pretty(&Value::Object(map.clone()));
    world.put(SETTINGS, &edited);
    // A request that keeps the old value leaves the user's one, and changes the other key.
    let next = json!({"outputStyle": "Concise", "model": "opus"});
    let (_, applied) = world.write_through(&[], &next);
    assert_eq!(applied["state"], json!("applied"), "{applied}");
    map.insert("outputStyle".into(), json!("Concise"));
    let kept = pretty(&Value::Object(map));
    assert_eq!(world.read(SETTINGS), kept);

    // A request for a third value is a conflict: without a confirmation nothing is written.
    let third = json!({"outputStyle": "Concise", "model": "haiku"});
    let plan = world.plan(&[], &third);
    assert_eq!(plan["action"], json!("conflict"), "{plan}");
    let refused = world.apply_envelope(&[], &third, digest_of(&plan));
    assert_eq!(refused["ok"], json!(false), "{refused}");
    assert_eq!(world.read(SETTINGS), kept);
}

// SPEC-20261122-settings-write-check AC-8 (REQ-8)
#[test]
fn no_value_reaches_a_new_answer_or_line() {
    const PLAIN: &str = "plantedplain41";
    const DIGITS: &str = "2345678901234";
    // Built here so that no file holds a credential-shaped text.
    let credential = format!("{}{}", "ghp_", "a".repeat(30));
    let held = format!(
        "{{\"first\": \"{credential}\", \"second\": \"{PLAIN}\", \
         \"big\": {{\"n\": 1{DIGITS}567890123456, \"note\": \"{credential}\"}}, \
         \"twice\": \"{credential}\", \"twice\": \"{PLAIN}\"}}\n"
    );
    let cases: [(Option<&[u8]>, Value); 5] = [
        (
            None,
            json!({"statusLine": {"type": PLAIN, "command": format!("run-{PLAIN}")}}),
        ),
        (None, json!({"outputStyle": {"name": PLAIN}})),
        (
            None,
            json!({"permissions": {"allow": [
                format!("Write({PLAIN}/**)"),
                format!("Bash({PLAIN} * main)"),
            ]}}),
        ),
        (
            None,
            json!({"env": {
                "HOME": format!("/{PLAIN}"),
                "EDITOR": format!("${PLAIN}"),
                SWITCHES[0]: PLAIN,
            }}),
        ),
        (Some(held.as_bytes()), json!({"model": PLAIN})),
    ];
    let written: Vec<Written> = cases
        .into_iter()
        .map(|(before, request)| write_as_today(before, request))
        .collect();
    for w in &written {
        let planned = entries(&w.plan, &w.label);
        let applied = entries(&w.applied, &w.label);
        let lines = w.world.lines();
        assert_eq!(lines.len(), 1, "{}: one observation line", w.label);
        assert!(
            lines[0][FIELD].as_array().is_some_and(|l| !l.is_empty()),
            "{}: {}",
            w.label,
            lines[0]
        );
        let log = fs::read_to_string(w.world.log_path()).unwrap();
        let new_texts = [
            Value::Array(planned).to_string(),
            Value::Array(applied).to_string(),
            w.plan
                .get(RECORDED)
                .map_or_else(String::new, Value::to_string),
            w.applied
                .get(RECORDED)
                .map_or_else(String::new, Value::to_string),
            log,
        ];
        for text in &new_texts {
            for planted in [PLAIN, DIGITS, credential.as_str()] {
                assert!(
                    !text.contains(planted),
                    "{}: a value reached a new answer or line: {text}",
                    w.label
                );
            }
        }
    }
}

/// The line names `file`, has a time and lists exactly the rule and key of `answer`'s entries.
fn assert_line(line: &Value, file: &str, answer: &Value, label: &str) {
    assert_eq!(line["file"], json!(file), "{label}: {line}");
    assert!(
        line["unix_ms"].as_u64().is_some_and(|t| t > 0),
        "{label}: {line}"
    );
    let listed = line[FIELD]
        .as_array()
        .unwrap_or_else(|| panic!("{label}: no list: {line}"));
    let answered = answer
        .get(FIELD)
        .and_then(Value::as_array)
        .cloned()
        .unwrap_or_default();
    assert_eq!(pairs(listed), pairs(&answered), "{label}: {line}");
}

// SPEC-20261122-settings-write-check AC-9 (REQ-9, REQ-10)
#[test]
fn every_changing_apply_is_recorded() {
    let world = World::new();
    // A project with a fixed id beside a HOME state folder: the layout in which the shared receipt
    // writer would put a log under HOME when it is given the HOME.
    let state = world.project.join(moai_cli::PROJECT_STATE_DIR);
    fs::create_dir_all(&state).unwrap();
    fs::write(
        state.join("project.yaml"),
        "{\"schema_version\":2,\"project_id\":\"demo\"}\n",
    )
    .unwrap();
    fs::create_dir_all(moai_cli::home_root::path(&world.home)).unwrap();
    assert!(
        moai_cli::project_home::resolve(&world.home, &world.project)
            .unwrap()
            .write_area("receipts")
            .unwrap()
            .is_some(),
        "the fixture has a HOME receipts folder to avoid"
    );
    let home_before = tree_state(&world.home);
    // Each step: its name, the lines after it and its answer.
    let mut steps: Vec<(&str, Vec<Value>, Value)> = Vec::new();
    let mut done = |name: &'static str, answer: Value| {
        steps.push((name, world.lines(), answer));
    };

    let first = json!({"model": "opus"});
    done("plan", world.plan(&[], &first));
    done("create", world.write_through(&[], &first).1);
    let (_, again) = world.write_through(&[], &first);
    assert_eq!(again["state"], json!("unchanged"), "{again}");
    done("unchanged", again);

    let second = json!({"model": "opus", "outputStyle": ""});
    done("reported", world.write_through(&[], &second).1);

    let third = json!({"model": "fable", "outputStyle": ""});
    let stale = format!("sha256:{}", "0".repeat(64));
    let refused = world.apply_envelope(&[], &third, &stale);
    assert_eq!(refused["ok"], json!(false), "{refused}");
    done("refused", refused);

    done(
        "local",
        world.write_through(&["--scope", "local"], &first).1,
    );

    // The status line apply: MoAI's own line, behind its gate.
    let mut with_line = second.clone();
    with_line["statusLine"] = moai_cli::status_line::command_of(moai_cli::hooks::HOOK_BINARY);
    let (_, line_applied) = world.write_through(&[], &with_line);
    assert_eq!(
        line_applied["status_line_migration"]["state"],
        json!("created"),
        "{line_applied}"
    );
    done("status line", line_applied);

    // A confirmed replacement of a key the user changed.
    let mut map: Map<String, Value> = serde_json::from_slice(&world.read(SETTINGS)).unwrap();
    map.insert("model".into(), json!("sonnet"));
    world.put(SETTINGS, &pretty(&Value::Object(map)));
    let mut fourth = with_line.clone();
    fourth["model"] = json!("haiku");
    let conflict = world.plan(&[], &fourth);
    let token = conflict["confirm_replace"]["confirm_token"]
        .as_str()
        .unwrap_or_else(|| panic!("no replacement offered: {conflict}"))
        .to_string();
    let replaced = ok(world.apply_envelope(
        &["--confirm-replace", &token],
        &fourth,
        digest_of(&conflict),
    ));
    assert_eq!(replaced["state"], json!("replaced"), "{replaced}");
    done("replaced", replaced);

    let hooks_plan = world.hooks_plan();
    done("hooks plan", hooks_plan.clone());
    let hooked = world.hooks_apply(&hooks_plan);
    assert_eq!(hooked["state"], json!("applied"), "{hooked}");
    done("hooks", hooked);
    let hooked_again = world.hooks_apply(&world.hooks_plan());
    assert_eq!(hooked_again["state"], json!("unchanged"), "{hooked_again}");
    done("hooks unchanged", hooked_again);

    // A log that cannot be appended to: the apply goes on and says so.
    let blocked: Vec<(&str, World, Value)> = ["a link", "a dangling link", "a folder"]
        .into_iter()
        .map(|kind| {
            let w = World::new();
            w.write_through(&[], &first);
            let log = w.log_path();
            fs::create_dir_all(log.parent().unwrap()).unwrap();
            if fs::symlink_metadata(&log).is_ok() {
                fs::remove_file(&log).unwrap();
            }
            let target = w.root.join("outside.txt");
            match kind {
                "a link" => {
                    fs::write(&target, b"keep\n").unwrap();
                    symlink(&target, &log).unwrap();
                }
                "a dangling link" => symlink(&target, &log).unwrap(),
                _ => fs::create_dir(&log).unwrap(),
            }
            let (_, applied) = w.write_through(&[], &second);
            assert_eq!(applied["state"], json!("applied"), "{kind}: {applied}");
            assert_eq!(
                w.read(SETTINGS),
                rewritten(Some(&rewritten(None, &first)), &second),
                "{kind}"
            );
            match kind {
                "a link" => assert_eq!(fs::read(&target).unwrap(), b"keep\n"),
                _ => assert!(fs::symlink_metadata(&target).is_err(), "{kind}"),
            }
            let meta = fs::symlink_metadata(&log).unwrap();
            assert_eq!(meta.file_type().is_symlink(), kind != "a folder", "{kind}");
            (kind, w, applied)
        })
        .collect();

    assert_eq!(
        tree_state(&world.home),
        home_before,
        "nothing is recorded under HOME"
    );

    // The step's name, the number of lines after it and, when it adds one, that line's file.
    let expected: [(&str, usize, Option<&str>); 11] = [
        ("plan", 0, None),
        ("create", 1, Some(SETTINGS)),
        ("unchanged", 1, None),
        ("reported", 2, Some(SETTINGS)),
        ("refused", 2, None),
        ("local", 3, Some(LOCAL)),
        ("status line", 4, Some(SETTINGS)),
        ("replaced", 5, Some(SETTINGS)),
        ("hooks plan", 5, None),
        ("hooks", 6, Some(SETTINGS)),
        ("hooks unchanged", 6, None),
    ];
    assert_eq!(steps.len(), expected.len());
    for ((name, lines, answer), (step, count, file)) in steps.iter().zip(expected) {
        assert_eq!(*name, step);
        assert_eq!(lines.len(), count, "{step}: {lines:?}");
        if let Some(file) = file {
            assert_line(lines.last().unwrap(), file, answer, step);
            assert_ne!(answer.get(RECORDED), Some(&json!(false)), "{step}");
        }
        if step == "reported" {
            let list = entries(answer, step);
            assert!(list.iter().all(|e| under(e, "outputStyle")), "{list:?}");
        } else {
            // A value the apply leaves as it is adds no entry.
            no_field(answer, step);
        }
    }

    for (kind, _, applied) in &blocked {
        assert_eq!(
            applied.get(RECORDED),
            Some(&json!(false)),
            "{kind}: {applied}"
        );
    }
}
