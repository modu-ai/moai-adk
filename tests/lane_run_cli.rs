//! SPEC-20261122-nonstop-lane slice 2: `lane-run show|set|prompt|goal`. The pick of running a lane
//! non-stop or one card at a time lives in the project profile's top-level `lane_run` key
//! (`default` for the project, `lanes` per lane); `set` writes it without asking, `show` reads
//! the lane's value before the project's and names the host facts that turn it into `single`,
//! `prompt` prints the self-paced `/loop` text for a non-stop lane, and `goal` prints a `/goal`
//! condition for one card. The binary runs in a clean environment on temporary projects.
mod common;

use common::CleanEnv;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;

const PROFILE: &str = "{\"schema_version\": 2, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\"}}\n";

struct Project {
    _dir: tempfile::TempDir,
    root: PathBuf,
    home: tempfile::TempDir,
}

/// A temporary project whose profile sits in the folder the resolver names for a new project.
fn project() -> Project {
    let dir = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    let profile = moai_cli::state_root::path(&root, "project.yaml");
    fs::create_dir_all(profile.parent().unwrap()).unwrap();
    fs::write(profile, PROFILE).unwrap();
    Project {
        _dir: dir,
        root,
        home: tempfile::tempdir().unwrap(),
    }
}

impl Project {
    fn profile_path(&self) -> PathBuf {
        self.root.join(moai_cli::profile::read_rel(&self.root))
    }

    fn profile(&self) -> Value {
        serde_json::from_slice(&fs::read(self.profile_path()).unwrap()).unwrap()
    }

    fn run_env(&self, args: &[&str], env: &[(&str, &str)]) -> (i32, Value) {
        let mut cmd = Command::new(common::MOAI_BIN);
        cmd.clean_env()
            .env("HOME", self.home.path())
            .current_dir(&self.root)
            .arg("lane-run")
            .args(args)
            .args(["--project", self.root.to_str().unwrap()]);
        for (k, v) in env {
            cmd.env(k, v);
        }
        let out = cmd.output().unwrap();
        let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: not JSON: {} / {}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        });
        (out.status.code().unwrap_or(-1), v)
    }

    fn run(&self, args: &[&str]) -> (i32, Value) {
        self.run_env(args, &[])
    }

    /// The result of a call that must succeed.
    fn ok(&self, args: &[&str]) -> Value {
        let (code, v) = self.run(args);
        assert_eq!((code, &v["ok"]), (0, &json!(true)), "{args:?}: {v}");
        v["result"].clone()
    }

    /// A call that must refuse, with the profile file byte-identical afterwards.
    fn refused(&self, args: &[&str]) -> Value {
        let before = fs::read(self.profile_path()).ok();
        let (code, v) = self.run(args);
        assert_ne!(code, 0, "{args:?}: {v}");
        assert_eq!(v["ok"], json!(false), "{args:?}: {v}");
        assert_eq!(fs::read(self.profile_path()).ok(), before, "{args:?}");
        v
    }
}

fn without_lane_run(mut v: Value) -> Value {
    v.as_object_mut().unwrap().remove("lane_run");
    v
}

// ---- pick_ (AC-2: REQ-2, REQ-3)

#[test]
fn pick_is_unset_then_the_project_value_then_the_lane_value_first() {
    let p = project();
    let before = p.profile();
    let shown = p.ok(&["show"]);
    assert_eq!(shown["saved"], json!("unset"), "{shown}");
    assert_eq!(shown["source"], json!("none"), "{shown}");
    assert_eq!(shown["effective"], json!("single"), "{shown}");

    let set = p.ok(&["set", "nonstop"]);
    assert_eq!(set["value"], json!("nonstop"), "{set}");
    let shown = p.ok(&["show"]);
    assert_eq!(
        (&shown["saved"], &shown["source"], &shown["effective"]),
        (&json!("nonstop"), &json!("project"), &json!("nonstop")),
        "{shown}"
    );

    p.ok(&["set", "single", "--lane", "agent-3"]);
    let lane = p.ok(&["show", "--lane", "agent-3"]);
    assert_eq!(
        (&lane["saved"], &lane["source"], &lane["effective"]),
        (&json!("single"), &json!("lane"), &json!("single")),
        "{lane}"
    );
    let other = p.ok(&["show", "--lane", "agent-4"]);
    assert_eq!(
        (&other["saved"], &other["source"]),
        (&json!("nonstop"), &json!("project")),
        "{other}"
    );

    // Nothing but the one key changed, and every other key keeps its place.
    let after = p.profile();
    assert_eq!(
        after["lane_run"],
        json!({"default": "nonstop", "lanes": {"agent-3": "single"}})
    );
    let keys = |v: &Value| -> Vec<String> { v.as_object().unwrap().keys().cloned().collect() };
    assert_eq!(keys(&without_lane_run(after.clone())), keys(&before));
    assert_eq!(without_lane_run(after), before);

    // The profile reader takes the key and keeps it.
    let (profile, _) = moai_cli::profile::read_profile(&p.root).unwrap();
    assert_eq!(
        serde_json::to_value(&profile).unwrap()["lane_run"],
        json!({"default": "nonstop", "lanes": {"agent-3": "single"}})
    );
}

#[test]
fn pick_set_again_changes_only_the_named_value() {
    let p = project();
    p.ok(&["set", "nonstop"]);
    p.ok(&["set", "nonstop", "--lane", "agent-1"]);
    p.ok(&["set", "single"]);
    assert_eq!(
        p.profile()["lane_run"],
        json!({"default": "single", "lanes": {"agent-1": "nonstop"}})
    );
    let shown = p.ok(&["show", "--lane", "agent-1"]);
    assert_eq!(shown["effective"], json!("nonstop"), "{shown}");
}

#[test]
fn pick_refuses_a_wrong_value_a_wrong_lane_and_a_missing_profile_with_nothing_written() {
    let p = project();
    p.refused(&["set", "fast"]);
    p.refused(&["set", "NONSTOP"]);
    p.refused(&["set", "nonstop", "--lane", "../x"]);
    p.refused(&["set", "nonstop", "--lane", ""]);
    p.refused(&["set", "nonstop", "--lane", &"a".repeat(65)]);
    p.refused(&["show", "--lane", "a b"]);
    assert!(p.profile().get("lane_run").is_none());

    // A project without a profile: refused, and no state folder appears.
    let empty = tempfile::tempdir().unwrap();
    let bare = Project {
        root: fs::canonicalize(empty.path()).unwrap(),
        _dir: empty,
        home: tempfile::tempdir().unwrap(),
    };
    let (code, v) = bare.run(&["set", "nonstop"]);
    assert_ne!(code, 0, "{v}");
    let left: Vec<_> = fs::read_dir(&bare.root).unwrap().collect();
    assert!(left.is_empty(), "{left:?}");
}

#[test]
fn pick_an_unknown_field_under_lane_run_is_refused_on_read() {
    let p = project();
    let mut v = p.profile();
    v["lane_run"] = json!({"default": "nonstop", "every": "day"});
    fs::write(p.profile_path(), serde_json::to_vec_pretty(&v).unwrap()).unwrap();
    assert!(moai_cli::profile::read_profile(&p.root).is_err());
    p.refused(&["show"]);

    v["lane_run"] = json!({"default": "always"});
    fs::write(p.profile_path(), serde_json::to_vec_pretty(&v).unwrap()).unwrap();
    p.refused(&["show"]);
}

/// A hand-edited file is checked lane by lane: a lane name `session-register` would refuse, or a
/// lane value that is no pick, makes the whole read refuse, on `show` and on `set`.
#[test]
fn pick_a_bad_lane_name_or_lane_value_in_the_file_is_refused_on_read() {
    let p = project();
    let mut v = p.profile();
    for bad in [
        json!({"default": "nonstop", "lanes": {"a b": "single"}}),
        json!({"default": "nonstop", "lanes": {"../x": "single"}}),
        json!({"lanes": {"agent-1": "always"}}),
        json!({"lanes": {"agent-1": "single", "agent-2": "NONSTOP"}}),
    ] {
        v["lane_run"] = bad.clone();
        fs::write(p.profile_path(), serde_json::to_vec_pretty(&v).unwrap()).unwrap();
        assert!(moai_cli::profile::read_profile(&p.root).is_err(), "{bad}");
        for args in [&["show"][..], &["set", "single", "--lane", "agent-1"][..]] {
            let e = p.refused(args);
            assert_eq!(
                e["error"]["message"],
                json!("invalid lane_run setting; no values were logged"),
                "{bad} {args:?}: {e}"
            );
        }
    }
    // The same lanes, written well, read.
    v["lane_run"] = json!({"default": "nonstop", "lanes": {"agent-1": "single"}});
    fs::write(p.profile_path(), serde_json::to_vec_pretty(&v).unwrap()).unwrap();
    assert_eq!(
        p.ok(&["show", "--lane", "agent-1"])["saved"],
        json!("single")
    );
}

/// `lane_run` is an object or absent, by one rule on `show` and on `set`: `null` and any other
/// non-object value are refused, never read as unset or as an empty setting.
#[test]
fn pick_a_null_or_non_object_lane_run_is_refused_on_show_and_on_set() {
    let p = project();
    let mut v = p.profile();
    // The profile reader takes these as unset or empty; lane-run refuses them by its own rule.
    // A string or a number already fails the reader. Either way show and set say the same.
    for (bad, why) in [
        (
            json!(null),
            "invalid lane_run setting; no values were logged",
        ),
        (json!([]), "invalid lane_run setting; no values were logged"),
        (json!("nonstop"), "invalid profile"),
        (json!(1), "invalid profile"),
    ] {
        v["lane_run"] = bad.clone();
        fs::write(p.profile_path(), serde_json::to_vec_pretty(&v).unwrap()).unwrap();
        for args in [
            &["show"][..],
            &["set", "nonstop"][..],
            &["prompt", "--lane", "a"][..],
        ] {
            let e = p.refused(args);
            assert_eq!(e["error"]["message"], json!(why), "{bad} {args:?}: {e}");
        }
    }
}

/// `set` writes through a temporary file, which is made 0600: the profile keeps its own mode.
#[cfg(unix)]
#[test]
fn pick_set_keeps_the_profile_file_mode() {
    use std::os::unix::fs::PermissionsExt;
    let p = project();
    for mode in [0o644, 0o600, 0o664] {
        fs::set_permissions(p.profile_path(), fs::Permissions::from_mode(mode)).unwrap();
        p.ok(&["set", "nonstop"]);
        let now = fs::metadata(p.profile_path()).unwrap().permissions().mode() & 0o777;
        assert_eq!(now, mode, "{mode:o} became {now:o}");
    }
}

// ---- facts_ (AC-5: REQ-9)

fn fact<'a>(shown: &'a Value, name: &str) -> &'a Value {
    shown["facts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["name"] == json!(name))
        .unwrap_or_else(|| panic!("no fact {name}: {shown}"))
}

#[test]
fn facts_a_disabled_scheduler_makes_the_pick_single() {
    let p = project();
    p.ok(&["set", "nonstop"]);
    for on in ["1", "true", "TRUE", "yes", "on", "On", " 1 ", "true\n"] {
        let (code, v) = p.run_env(&["show"], &[("CLAUDE_CODE_DISABLE_CRON", on)]);
        assert_eq!(code, 0, "{v}");
        let shown = &v["result"];
        assert_eq!(shown["saved"], json!("nonstop"), "{on}: {shown}");
        assert_eq!(shown["effective"], json!("single"), "{on}: {shown}");
        let f = fact(shown, "CLAUDE_CODE_DISABLE_CRON");
        assert_eq!(
            (&f["read"], &f["blocks"]),
            (&json!(true), &json!(true)),
            "{on}: {f}"
        );
    }
    for off in ["0", "false", "", "no", "off", "2", "enabled"] {
        let (_, v) = p.run_env(&["show"], &[("CLAUDE_CODE_DISABLE_CRON", off)]);
        assert_eq!(v["result"]["effective"], json!("nonstop"), "{off:?}: {v}");
    }
    let shown = p.ok(&["show"]);
    assert_eq!(shown["effective"], json!("nonstop"), "{shown}");
    let f = fact(&shown, "CLAUDE_CODE_DISABLE_CRON");
    assert_eq!(
        (&f["read"], &f["blocks"]),
        (&json!(true), &json!(false)),
        "{f}"
    );
}

#[test]
fn facts_that_moai_cannot_read_are_reported_unread_and_never_block() {
    let p = project();
    p.ok(&["set", "nonstop"]);
    let shown = p.ok(&["show"]);
    for name in ["loop_available", "claude_code_version"] {
        let f = fact(&shown, name);
        assert_eq!(
            (&f["read"], &f["blocks"]),
            (&json!(false), &json!(false)),
            "{f}"
        );
    }
    // The version floors are told as text.
    let floors = shown["floors"].as_str().unwrap();
    assert!(
        floors.contains("2.1.248") && floors.contains("2.1.310"),
        "{floors}"
    );
}

// ---- prompt_ (AC-3: REQ-4, REQ-5, REQ-6, REQ-12)

/// No card number, no key-shaped run and no absolute path outside the project.
fn assert_clean(text: &str, root: &Path) {
    let words: Vec<&str> = text
        .split(|c: char| c.is_whitespace() || "`'\"(),;".contains(c))
        .collect();
    for w in &words {
        let card = w.strip_prefix("C-").or_else(|| w.strip_prefix('#'));
        assert!(
            !card.is_some_and(|n| !n.is_empty() && n.chars().all(|c| c.is_ascii_digit())),
            "card number {w:?} in {text}"
        );
        let run = w
            .chars()
            .filter(|c| c.is_ascii_alphanumeric() || *c == '_' || *c == '-')
            .count();
        assert!(run < 32 || w.len() != run, "key-shaped {w:?}");
        if w.starts_with('/') || w.starts_with("~/") {
            assert!(
                Path::new(w).starts_with(root),
                "absolute path {w:?} outside the project in {text}"
            );
        }
    }
}

#[test]
fn prompt_for_a_nonstop_lane_names_every_rule_and_nothing_else() {
    let p = project();
    p.ok(&["set", "nonstop"]);
    let out = p.ok(&["prompt", "--lane", "agent-7"]);
    assert_eq!(out["pick"], json!("nonstop"), "{out}");
    let text = out["prompt"].as_str().unwrap();
    for needle in [
        "agent-7",
        "current card",
        "READY",
        "moai card next",
        "moai card show",
        "Monitor",
        "verdict",
        "LANE-BLOCKED",
        "never end a turn",
        "ScheduleWakeup",
        "stop: true",
        "idle",
    ] {
        assert!(text.contains(needle), "{needle:?} missing: {text}");
    }
    assert!(
        !text.starts_with('/'),
        "a plain text, not a command: {text}"
    );
    // Only commands that exist today: the watch command lands with a later slice.
    assert!(!text.contains("lane-run watch"), "{text}");
    assert_clean(text, &p.root);
    // The same lane and pick give the same text.
    assert_eq!(
        p.ok(&["prompt", "--lane", "agent-7"])["prompt"],
        out["prompt"]
    );
}

#[test]
fn prompt_for_a_single_or_unset_lane_prints_no_loop() {
    let p = project();
    let unset = p.ok(&["prompt", "--lane", "agent-7"]);
    assert_eq!(unset["prompt"], Value::Null, "{unset}");
    assert_eq!(unset["pick"], json!("single"), "{unset}");
    assert_eq!(unset["started"], json!(false), "{unset}");
    p.ok(&["set", "nonstop"]);
    p.ok(&["set", "single", "--lane", "agent-7"]);
    let single = p.ok(&["prompt", "--lane", "agent-7"]);
    assert_eq!(single["prompt"], Value::Null, "{single}");
    assert_eq!(single["pick"], json!("single"), "{single}");
    assert!(
        single["note"]
            .as_str()
            .unwrap()
            .contains("no loop is started"),
        "{single}"
    );
    // A non-stop pick with the scheduler off gives no loop either.
    p.ok(&["set", "nonstop", "--lane", "agent-7"]);
    let (_, v) = p.run_env(
        &["prompt", "--lane", "agent-7"],
        &[("CLAUDE_CODE_DISABLE_CRON", "1")],
    );
    assert_eq!(v["result"]["prompt"], Value::Null, "{v}");
}

#[test]
fn prompt_refuses_a_lane_name_it_does_not_accept() {
    let p = project();
    p.ok(&["set", "nonstop"]);
    for bad in ["", "a b", "../x", "x/y", "a".repeat(65).as_str()] {
        let (code, v) = p.run(&["prompt", "--lane", bad]);
        assert_ne!(code, 0, "{bad:?}: {v}");
        assert_eq!(v["error"]["kind"], json!("invalid_input"), "{bad:?}: {v}");
        assert!(
            v["error"]["message"]
                .as_str()
                .unwrap()
                .starts_with("invalid lane name:"),
            "{bad:?}: {v}"
        );
    }
    // A name that looks like an option never reaches the check: the parser refuses it.
    let (code, v) = p.run(&["prompt", "--lane", "-a"]);
    assert_ne!(code, 0, "{v}");
    assert_eq!(v["error"]["kind"], json!("usage"), "{v}");
}

// ---- goal_ (AC-6: REQ-10)

#[test]
fn goal_text_names_the_card_end_state_check_and_turns_within_the_limit() {
    let p = project();
    let out = p.ok(&["goal", "--card", "C-999999", "--turns", "200"]);
    let text = out["condition"].as_str().unwrap();
    assert!(text.chars().count() <= 4000, "{}", text.len());
    for needle in ["C-999999", "READY", "head", "base", "git", "200 turns"] {
        assert!(text.contains(needle), "{needle:?} missing: {text}");
    }
    assert_eq!(out["started"], json!(false), "{out}");
    assert_eq!(out["max_chars"], json!(4000), "{out}");
    // The default turn limit is printed too.
    let plain = p.ok(&["goal", "--card", "C-12"]);
    assert!(
        plain["condition"].as_str().unwrap().contains(" turns"),
        "{plain}"
    );
}

#[test]
fn goal_refuses_a_wrong_card_or_turn_limit() {
    let p = project();
    for (args, why) in [
        (vec!["goal", "--card", "12"], "invalid card name"),
        (vec!["goal", "--card", "C-0"], "invalid card name"),
        (vec!["goal", "--card", "C-012"], "invalid card name"),
        (vec!["goal", "--card", "C-1234567"], "invalid card name"),
        (
            vec!["goal", "--card", "C-12", "--turns", "0"],
            "invalid turn limit",
        ),
        (
            vec!["goal", "--card", "C-12", "--turns", "201"],
            "invalid turn limit",
        ),
    ] {
        let (code, v) = p.run(&args);
        assert_ne!(code, 0, "{args:?}: {v}");
        assert_eq!(v["error"]["kind"], json!("invalid_input"), "{args:?}: {v}");
        assert!(
            v["error"]["message"].as_str().unwrap().starts_with(why),
            "{args:?}: {v}"
        );
    }
}
