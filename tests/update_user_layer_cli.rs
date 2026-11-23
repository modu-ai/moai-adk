//! #2638: `moai-cli update` installs the user layer — the skill copies, the base agents and the
//! routing section of the global instruction files — from the real `skills/moai` package placed
//! beside a copy of the binary, inside a temporary HOME. The real HOME is never touched.

#![cfg(unix)]

mod common;
use common::CleanEnv;

use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;

fn package() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
}

fn copy_tree(from: &Path, to: &Path) {
    fs::create_dir_all(to).unwrap();
    for entry in fs::read_dir(from).unwrap() {
        let p = entry.unwrap().path();
        let dest = to.join(p.file_name().unwrap());
        if p.is_dir() {
            copy_tree(&p, &dest);
        } else {
            fs::copy(&p, &dest).unwrap();
        }
    }
}

/// A temporary HOME with `~/.moai-cli/versions/a/bin/moai-cli`, the real package beside it and
/// the agent sources in `agents/` next to `skills/` (#2721).
fn setup() -> (TempDir, PathBuf, PathBuf, PathBuf) {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    let folder = home.join(".moai-cli/versions/a");
    fs::create_dir_all(folder.join("bin")).unwrap();
    let exe = folder.join("bin/moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    let pkg = folder.join("skills/moai");
    copy_tree(&package(), &pkg);
    copy_tree(
        &Path::new(env!("CARGO_MANIFEST_DIR")).join("agents"),
        &folder.join("agents"),
    );
    (tmp, home, exe, pkg)
}

/// SPEC-20261122-html-report-install: the skill folder a release carries beside `skills/moai`.
const HTML_REPORT: &str = moai_cli::update::extra_skill::SKILLS[0];

/// Copies the repository's `skills/moai-html-report` beside the fixture's `skills/moai`, as a
/// release that carries it lays it out; returns the copy.
fn with_html_report(pkg: &Path) -> PathBuf {
    let dest = pkg.parent().unwrap().join(HTML_REPORT);
    copy_tree(
        &Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills")
            .join(HTML_REPORT),
        &dest,
    );
    dest
}

/// Every file under `dir`, relative, with its bytes.
fn tree(dir: &Path) -> Vec<(PathBuf, Vec<u8>)> {
    let mut out: Vec<(PathBuf, Vec<u8>)> = state(dir)
        .into_iter()
        .map(|(p, b)| (p.strip_prefix(dir).unwrap().to_path_buf(), b))
        .collect();
    out.sort();
    out
}

/// The receipts of the html-report skill under HOME's state folder.
fn html_report_receipts(home: &Path) -> Vec<Value> {
    let tail = Path::new("hosts").join(HTML_REPORT).join("claude.json");
    state(home)
        .into_iter()
        .filter(|(p, _)| p.ends_with(&tail))
        .map(|(_, b)| serde_json::from_slice(&b).unwrap())
        .collect()
}

/// #2721 AC-3: an install from a package that still carried the agent sources inside the skill
/// (`references/agents/`) moves to the new layout. The base agents are found by name, so they
/// stay unchanged; the five old skill-tree files are reported `remove_blocked` and kept — no
/// host loads an agent from a skill folder, and the updater never deletes.
#[test]
fn an_install_with_agents_inside_the_skill_moves_to_the_new_layout() {
    let (_t, home, exe, pkg) = setup();
    let old = pkg.join("references/agents");
    copy_tree(&Path::new(env!("CARGO_MANIFEST_DIR")).join("agents"), &old);
    let p = result(&run(&exe, &home, &["update"]));
    let top = p["plan_digest"].as_str().unwrap().to_string();
    let r = result(&run(&exe, &home, &["update", "--confirm", &top]));
    assert_eq!(r["agents_result"]["state"], json!("applied"), "{r}");
    let copied = skill_dir(&home, "claude").join("references/agents/impl.md");
    assert!(copied.is_file());
    // The next release ships the sources only beside the skill.
    fs::remove_dir_all(&old).unwrap();
    let p = result(&run(&exe, &home, &["update"]));
    for host in p["agents"]["hosts"].as_array().unwrap() {
        for f in host["files"].as_array().unwrap() {
            assert_eq!(f["action"], json!("unchanged"), "{f}");
        }
    }
    let claude = p["plan"]["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == json!("claude"))
        .unwrap();
    let stale: Vec<&str> = claude["files"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| {
            f["path"]
                .as_str()
                .unwrap()
                .starts_with("references/agents/")
        })
        .map(|f| f["action"].as_str().unwrap())
        .collect();
    assert_eq!(stale, ["remove_blocked"; 5], "{p}");
    assert!(copied.is_file(), "a preview deletes nothing");
}

/// #2721: a release folder without `agents/` (or with a symlink there) is refused with the
/// expected path; nothing is written.
#[test]
fn a_release_without_a_real_agent_folder_is_refused() {
    let (_t, home, exe, pkg) = setup();
    let agents = pkg.parent().unwrap().parent().unwrap().join("agents");
    fs::remove_dir_all(&agents).unwrap();
    let before = state(&home);
    let out = run(&exe, &home, &["update"]);
    assert!(!out.status.success());
    let text = String::from_utf8_lossy(&out.stdout);
    assert!(text.contains("no agent sources installed"), "{text}");
    std::os::unix::fs::symlink(
        Path::new(env!("CARGO_MANIFEST_DIR")).join("agents"),
        &agents,
    )
    .unwrap();
    let out = run(&exe, &home, &["update"]);
    assert!(!out.status.success());
    // Refused where the release folder is resolved, before any source is read.
    let text = String::from_utf8_lossy(&out.stdout);
    assert!(
        text.contains("installed agent source folder is a symlink"),
        "{text}"
    );
    // A file in its place is refused the same way.
    fs::remove_file(&agents).unwrap();
    fs::write(&agents, "x").unwrap();
    let out = run(&exe, &home, &["update"]);
    let text = String::from_utf8_lossy(&out.stdout);
    assert!(
        text.contains("the installed agent source path is not a folder"),
        "{text}"
    );
    assert_eq!(state(&home), before);
}

fn run(exe: &Path, home: &Path, args: &[&str]) -> Output {
    Command::new(exe)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .current_dir(home)
        .stdin(Stdio::null())
        .output()
        .unwrap()
}

fn result(out: &Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// Every file under HOME except the version folder, with its bytes.
fn state(home: &Path) -> Vec<(PathBuf, Vec<u8>)> {
    fn walk(dir: &Path, skip: &Path, out: &mut Vec<(PathBuf, Vec<u8>)>) {
        let Ok(entries) = fs::read_dir(dir) else {
            return;
        };
        for e in entries {
            let p = e.unwrap().path();
            if p.starts_with(skip) {
                continue;
            }
            if p.is_dir() {
                walk(&p, skip, out);
            } else {
                out.push((p.clone(), fs::read(&p).unwrap()));
            }
        }
    }
    let mut out = Vec::new();
    walk(home, &home.join(".moai-cli/versions"), &mut out);
    out.sort();
    out
}

fn agents_dir(home: &Path, host: &str) -> PathBuf {
    assert_eq!(host, "claude", "the home update writes Claude only");
    home.join(".claude/agents/moai")
}

fn skill_dir(home: &Path, agent_host: &str) -> PathBuf {
    assert_eq!(agent_host, "claude", "the home update writes Claude only");
    home.join(".claude/skills/moai")
}

#[test]
fn update_previews_and_installs_the_whole_user_layer_then_has_nothing_to_apply() {
    let (_t, home, exe, _) = setup();
    let before = state(&home);
    let p = result(&run(&exe, &home, &["update"]));
    assert_eq!(state(&home), before, "a preview writes nothing");
    assert_eq!(p["nothing_to_apply"], json!(false), "{p}");
    let top = p["plan_digest"].as_str().unwrap();
    assert_eq!(
        p["confirm_with"],
        json!(format!("moai update --confirm {top}"))
    );
    assert_eq!(p["agents"]["hosts"].as_array().unwrap().len(), 1, "{p}");
    assert_eq!(p["routing"]["hosts"].as_array().unwrap().len(), 1, "{p}");
    // #2875: the style file and the default style, both Claude's.
    assert_eq!(p["style"]["targets"].as_array().unwrap().len(), 2, "{p}");
    // SPEC-20261122-output-style-install-checks: the style write brings the `/status` pointer,
    // in its own `notices` list and never among the skipped writes.
    let status = |list: &Value| {
        list.as_array()
            .unwrap()
            .iter()
            .filter(|n| n["reason"].as_str().is_some_and(|r| r.contains("/status")))
            .count()
    };
    assert_eq!(status(&p["notices"]), 1, "{p}");
    assert_eq!(status(&p["skipped"]), 0, "{p}");
    // A part's own digest is not a confirmation of the whole layer.
    for part in ["plan", "agents", "routing", "style"] {
        let d = p[part]["plan_digest"].as_str().unwrap();
        assert_ne!(d, top);
        let out = run(&exe, &home, &["update", "--confirm", d]);
        assert!(!out.status.success(), "{part}");
        assert_eq!(state(&home), before, "a refused confirm writes nothing");
    }
    let r = result(&run(&exe, &home, &["update", "--confirm", top]));
    assert_eq!(r["applied"], json!(true), "{r}");
    assert_eq!(r["agents_result"]["state"], json!("applied"), "{r}");
    assert_eq!(r["routing_result"]["state"], json!("applied"), "{r}");
    assert_eq!(r["style_result"]["state"], json!("applied"), "{r}");
    assert_eq!(
        status(&r["notices"]),
        1,
        "the result keeps the pointer: {r}"
    );
    assert!(r["revert"]["style"]
        .as_str()
        .unwrap()
        .starts_with("moai home-style-revert"));
    assert!(r["revert"]["agents"]
        .as_str()
        .unwrap()
        .starts_with("moai agents-home-revert"));
    assert!(r["revert"]["routing"]
        .as_str()
        .unwrap()
        .starts_with("moai home-routing-revert"));
    // #2936 E1: the commands are only safe last applied first (style, then routing, then agents):
    // reverting routing first is skipped as changed since the apply. The envelope says so.
    assert_eq!(
        r["revert_order"],
        json!(["style", "routing", "agents"]),
        "{r}"
    );
    // The routing file carries the section and the flow rule file sits beside it; the user's
    // `CLAUDE.md` is not touched, and no Codex or shared-agents file is made.
    assert!(fs::read_to_string(home.join(".claude/rules/moai.md"))
        .unwrap()
        .contains("(/moai)"));
    assert!(home.join(".claude/rules/moai/flow.md").is_file());
    assert!(!home.join(".claude/CLAUDE.md").exists());
    for gone in [".codex", ".agents", ".omp"] {
        assert!(!home.join(gone).exists(), "{gone} was made");
    }
    // #2875: MoAI-Easy is Claude Code's default style.
    assert!(home.join(".claude/output-styles/MoAI-Easy.md").is_file());
    let settings: serde_json::Value =
        serde_json::from_slice(&fs::read(home.join(".claude/settings.json")).unwrap()).unwrap();
    assert_eq!(settings["outputStyle"], json!("MoAI-Easy"));
    // Every `references/…` path an installed agent names exists in the skill copy it reads.
    let mut checked = 0;
    for host in ["claude"] {
        let dir = agents_dir(&home, host);
        let files: Vec<PathBuf> = fs::read_dir(&dir)
            .unwrap()
            .map(|e| e.unwrap().path())
            .collect();
        assert_eq!(files.len(), 5, "{host}: {files:?}");
        for f in files {
            let text = fs::read_to_string(&f).unwrap();
            for part in text.split('`').filter(|s| s.starts_with("references/")) {
                assert!(
                    skill_dir(&home, host).join(part).is_file(),
                    "{host} {}: {part}",
                    f.display()
                );
                checked += 1;
            }
        }
    }
    assert!(checked > 0, "no references checked");
    // A second run has nothing left to do.
    let again = result(&run(&exe, &home, &["update"]));
    assert_eq!(again["nothing_to_apply"], json!(true), "{again}");
    // Each part counts on its own: with only the section (then only the agents) reverted, the
    // preview still has something to apply.
    let h = home.to_str().unwrap();
    for (op, part) in [
        ("home-routing-revert", "routing_result"),
        ("agents-home-revert", "agents_result"),
        ("home-style-revert", "style_result"),
    ] {
        let receipt = r[part]["receipt"].as_str().unwrap();
        result(&run(
            &exe,
            &home,
            &[
                op,
                h,
                "--receipt",
                receipt,
                "--permission-reference",
                "fixture",
            ],
        ));
        let p = result(&run(&exe, &home, &["update"]));
        assert_eq!(p["nothing_to_apply"], json!(false), "{op}: {p}");
        // Put the part back, so the next case misses only its own part.
        let top = p["plan_digest"].as_str().unwrap();
        result(&run(&exe, &home, &["update", "--confirm", top]));
        let full = result(&run(&exe, &home, &["update"]));
        assert_eq!(full["nothing_to_apply"], json!(true), "{op}: {full}");
    }
}

/// #2936 E1 (U1): the printed `revert_order` is the order that really takes the user layer back.
/// The user's own instruction files and agent start as they were, byte for byte, once style,
/// routing and agents are reverted in that order with the printed commands; the commands are run
/// as printed, `<HOME>` and `<reference>` filled in.
#[test]
fn the_printed_revert_order_gives_back_the_users_own_files_byte_for_byte() {
    let (_t, home, exe, _) = setup();
    let mine = [
        (".claude/CLAUDE.md", "# My Claude notes\n"),
        (".claude/settings.json", "{\"theme\":\"dark\"}\n"),
        (".claude/agents/my-own.md", "---\nname: my-own\n---\nmine\n"),
    ];
    for (rel, text) in mine {
        let path = home.join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, text).unwrap();
    }
    let p = result(&run(&exe, &home, &["update"]));
    let top = p["plan_digest"].as_str().unwrap();
    let r = result(&run(&exe, &home, &["update", "--confirm", top]));
    assert_eq!(r["applied"], json!(true), "{r}");
    for (rel, text) in mine {
        // The layer wrote into `settings.json` and beside the others: the routing text is its own
        // file now, so the user's `CLAUDE.md` is left as it was.
        let now = fs::read_to_string(home.join(rel)).unwrap();
        if rel == ".claude/settings.json" {
            assert_ne!(now, text, "{rel} was not touched by the apply");
        } else {
            assert_eq!(now, text, "{rel} was touched by the apply");
        }
    }
    let order: Vec<String> = r["revert_order"]
        .as_array()
        .unwrap()
        .iter()
        .map(|v| v.as_str().unwrap().to_string())
        .collect();
    assert_eq!(order, ["style", "routing", "agents"], "{r}");
    let h = home.to_str().unwrap();
    for step in &order {
        let printed = r["revert"][step.as_str()].as_str().unwrap();
        let words: Vec<String> = printed
            .split_whitespace()
            .skip(1)
            .map(|w| match w {
                "<HOME>" => h.to_string(),
                "<reference>" => "fixture".to_string(),
                other => other.to_string(),
            })
            .collect();
        let args: Vec<&str> = words.iter().map(String::as_str).collect();
        let out = result(&run(&exe, &home, &args));
        assert_ne!(out["state"], json!("conflict"), "{step}: {out}");
        assert_ne!(out["retryable"], json!(false), "{step}: {out}");
    }
    for (rel, text) in mine {
        assert_eq!(
            fs::read_to_string(home.join(rel)).unwrap(),
            text,
            "{rel} is not as the user had it"
        );
    }
    let left: Vec<String> = fs::read_dir(home.join(".claude/agents"))
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    assert_eq!(left, ["my-own.md"], "only the user's own agent remains");
}

/// Skills already current (installed through `update-plan`/`update-apply`), agents and section
/// missing: `update` still has something to apply and installs only those two parts.
#[test]
fn current_skills_with_missing_agents_still_have_something_to_apply() {
    let (_t, home, exe, pkg) = setup();
    let (h, p) = (home.to_str().unwrap(), pkg.to_str().unwrap());
    let plan = result(&run(&exe, &home, &["update-plan", h, p]));
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let mut child = Command::new(&exe)
        .args([
            "update-apply",
            h,
            p,
            "--plan-digest",
            &digest,
            "--permission-reference",
            "fixture-approved",
        ])
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", &home)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    use std::io::Write as _;
    child
        .stdin
        .take()
        .unwrap()
        .write_all(plan.to_string().as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    assert_eq!(result(&out)["status"], json!("applied"));
    let p = result(&run(&exe, &home, &["update"]));
    for h in p["plan"]["hosts"].as_array().unwrap() {
        assert!(
            h["files"]
                .as_array()
                .unwrap()
                .iter()
                .all(|f| f["action"] == json!("unchanged")),
            "{h}"
        );
    }
    assert_eq!(p["nothing_to_apply"], json!(false), "{p}");
    let top = p["plan_digest"].as_str().unwrap().to_string();
    let r = result(&run(&exe, &home, &["update", "--confirm", &top]));
    assert_eq!(r["applied"], json!(true), "{r}");
    assert_eq!(r["agents_result"]["state"], json!("applied"), "{r}");
    assert!(home.join(".claude/rules/moai.md").is_file());
}

#[test]
fn a_blocked_skill_copy_gets_no_agents_and_no_section() {
    let (_t, home, exe, _) = setup();
    // A skill this tool did not install blocks the Claude copy (`~/.claude/skills/moai`).
    let foreign = home.join(".claude/skills/moai");
    fs::create_dir_all(&foreign).unwrap();
    fs::write(
        foreign.join("SKILL.md"),
        "---\nname: moai\ndescription: someone else's\n---\n",
    )
    .unwrap();
    let p = result(&run(&exe, &home, &["update"]));
    let hosts = |part: &str| -> Vec<String> {
        p[part]["hosts"]
            .as_array()
            .unwrap()
            .iter()
            .map(|h| h["host"].as_str().unwrap().to_string())
            .collect()
    };
    assert!(hosts("agents").is_empty(), "{p}");
    assert!(hosts("routing").is_empty(), "{p}");
    let skipped: Vec<String> = p["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| {
            format!(
                "{}:{}",
                s["part"].as_str().unwrap(),
                s["host"].as_str().unwrap()
            )
        })
        .collect();
    assert_eq!(
        skipped,
        ["agents:claude", "routing:claude", "style:claude"],
        "{p}"
    );
    assert!(p["style"]["targets"].as_array().unwrap().is_empty(), "{p}");
    // The other parts still go in; the foreign skill and the agents and section stay out.
    let top = p["plan_digest"].as_str().unwrap().to_string();
    let r = result(&run(&exe, &home, &["update", "--confirm", &top]));
    assert_eq!(r["applied"], json!(true), "{r}");
    assert!(!agents_dir(&home, "claude").exists());
    assert!(!home.join(".claude/rules/moai.md").exists());
    assert_eq!(
        fs::read_to_string(foreign.join("SKILL.md")).unwrap(),
        "---\nname: moai\ndescription: someone else's\n---\n"
    );
}

/// #2872 stage 2: Go leftovers a cleanup plan archives — a Go `~/.moai` (no live session) whose
/// user-asset manifest proves the sibling skill `~/.claude/skills/moai-proven/`.
fn go_leftovers(home: &Path) {
    use sha2::{Digest, Sha256};
    let sha = |b: &[u8]| format!("{:x}", Sha256::digest(b));
    let put = |rel: &str, body: &str| {
        let p = home.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    };
    put(".moai/state/active-sessions.json", "[]");
    put(".claude/skills/moai-proven/SKILL.md", "proven\n");
    put(".claude/skills/moai-proven/refs/x.md", "x\n");
    // Keys are `<slug>/<path>`, joined at run time: a literal `…-skills/<path>` reads as a
    // repository skill path to tests/skill_references_test.py.
    let slug = "claude-skills";
    let mut files = serde_json::Map::new();
    for (rel, body) in [
        ("moai-proven/SKILL.md", "proven\n"),
        ("moai-proven/refs/x.md", "x\n"),
    ] {
        files.insert(
            format!("{slug}/{rel}"),
            json!({"sha256": sha(body.as_bytes()), "bundle": "core"}),
        );
    }
    put(
        ".moai/user-assets.json",
        &json!({"schema_version": 1, "bundles": {}, "files": files, "collisions": []}).to_string(),
    );
}

const PREVIEW_KEYS: [&str; 14] = [
    "mode",
    "applied",
    "nothing_to_apply",
    "package",
    "plan_digest",
    "plan",
    "agents",
    "routing",
    "style",
    "bin_names",
    "hooks",
    "skipped",
    "notices",
    "confirm_with",
];

fn keys(v: &Value) -> Vec<String> {
    v.as_object().unwrap().keys().cloned().collect()
}

fn refused(out: &Output) -> String {
    assert!(!out.status.success());
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    v["error"]["message"].as_str().unwrap().to_string()
}

/// #2872 stage 2: with nothing to archive the preview is exactly what it was — no `cleanup`
/// key — and a cleanup token is refused before anything is written.
#[test]
fn without_go_leftovers_the_preview_has_no_cleanup_and_a_token_is_refused() {
    let (_t, home, exe, _) = setup();
    let before = state(&home);
    let p = result(&run(&exe, &home, &["update"]));
    assert_eq!(keys(&p), PREVIEW_KEYS, "{p}");
    let e = refused(&run(
        &exe,
        &home,
        &["update", "--confirm-cleanup", "cleanup:0000000000000000"],
    ));
    assert!(e.contains("cleanup token"), "{e}");
    assert_eq!(state(&home), before, "a refused cleanup writes nothing");
}

/// #2872 stage 2: Go leftovers show a cleanup preview with its token; the cleanup is not part
/// of `plan_digest`, and `--confirm <digest>` applies the layer but moves nothing.
#[test]
fn go_leftovers_are_previewed_outside_the_digest_and_confirm_moves_none() {
    let (_t, home, exe, _) = setup();
    let plain = result(&run(&exe, &home, &["update"]));
    go_leftovers(&home);
    let p = result(&run(&exe, &home, &["update"]));
    assert_eq!(
        p["plan_digest"], plain["plan_digest"],
        "cleanup is in the digest"
    );
    let mut with: Vec<&str> = PREVIEW_KEYS.to_vec();
    with.push("cleanup");
    // #2871 U2b: a Go registry file makes this a Go HOME, so the home part is previewed too —
    // beside the cleanup, outside the digest, and (this narrow PATH hides `lsof`) with nothing
    // offered.
    with.push("home");
    assert_eq!(keys(&p), with, "{p}");
    let token = p["cleanup"]["confirm_token"].as_str().unwrap();
    assert!(token.starts_with("cleanup:"), "{p}");
    assert_eq!(p["cleanup"]["counts"]["archive"], json!(1), "{p}");
    assert_eq!(p["cleanup"]["writes"], json!(false));
    let top = p["plan_digest"].as_str().unwrap();
    let r = result(&run(&exe, &home, &["update", "--confirm", top]));
    assert_eq!(r["applied"], json!(true), "{r}");
    assert!(r.get("cleanup_result").is_none(), "{r}");
    assert!(r["revert"].get("cleanup").is_none(), "{r}");
    assert!(home.join(".claude/skills/moai-proven/SKILL.md").is_file());
    assert!(!home.join(".moai-cli/cleanup").exists());
}

/// #2872 stage 2: a wrong or stale cleanup token is refused before any write, with or without
/// `--confirm`; the right one alone archives the leftovers while the layer stays a preview, and
/// `cleanup-revert` puts the bytes back.
#[test]
fn the_cleanup_token_alone_archives_and_cleanup_revert_restores() {
    let (_t, home, exe, _) = setup();
    go_leftovers(&home);
    let p = result(&run(&exe, &home, &["update"]));
    let token = p["cleanup"]["confirm_token"].as_str().unwrap().to_string();
    let top = p["plan_digest"].as_str().unwrap().to_string();
    let before = state(&home);
    let wrong = "cleanup:0000000000000000";
    for args in [
        vec!["update", "--confirm-cleanup", wrong],
        vec!["update", "--confirm", &top, "--confirm-cleanup", wrong],
    ] {
        let e = refused(&run(&exe, &home, &args));
        assert!(e.contains("cleanup token"), "{args:?}: {e}");
        assert_eq!(
            state(&home),
            before,
            "{args:?}: a refused token writes nothing"
        );
    }
    let r = result(&run(&exe, &home, &["update", "--confirm-cleanup", &token]));
    assert_eq!(r["mode"], json!("preview"), "{r}");
    assert_eq!(r["applied"], json!(false), "{r}");
    assert!(
        !home.join(".claude/skills/moai").exists(),
        "the layer stays a preview"
    );
    assert_eq!(r["cleanup_result"]["state"], json!("committed"), "{r}");
    assert_eq!(r["cleanup_result"]["moved"], json!(1), "{r}");
    let stamp = r["cleanup_result"]["stamp"].as_str().unwrap();
    assert_eq!(
        r["revert"]["cleanup"],
        json!(format!("moai cleanup-revert-plan <HOME> --stamp {stamp}"))
    );
    // The cleanup's revert is in the list the agent and the screen follow (AUDIT #3133 B2).
    assert_eq!(r["revert_order"], json!(["cleanup"]), "{r}");
    let archive = home.join(".moai-cli/cleanup").join(stamp);
    assert!(archive.join("receipt.json").is_file());
    assert_eq!(
        fs::read(archive.join(".claude/skills/moai-proven/refs/x.md")).unwrap(),
        b"x\n"
    );
    assert!(!home.join(".claude/skills/moai-proven").exists());
    // The same token again is stale: nothing is left to archive.
    let after = state(&home);
    refused(&run(&exe, &home, &["update", "--confirm-cleanup", &token]));
    assert_eq!(state(&home), after);
    let h = home.to_str().unwrap();
    result(&run(
        &exe,
        &home,
        &["cleanup-revert-plan", h, "--stamp", stamp],
    ));
    let rv = result(&run(
        &exe,
        &home,
        &[
            "cleanup-revert",
            h,
            "--stamp",
            stamp,
            "--permission-reference",
            "fixture",
        ],
    ));
    assert_eq!(rv["state"], json!("reverted"), "{rv}");
    assert_eq!(
        fs::read(home.join(".claude/skills/moai-proven/SKILL.md")).unwrap(),
        b"proven\n"
    );
    assert_eq!(
        fs::read(home.join(".claude/skills/moai-proven/refs/x.md")).unwrap(),
        b"x\n"
    );
}

/// #2872 stage 2: `--confirm` and `--confirm-cleanup` together apply the layer, then archive.
#[test]
fn confirm_with_the_cleanup_token_applies_the_layer_then_archives() {
    let (_t, home, exe, _) = setup();
    go_leftovers(&home);
    let p = result(&run(&exe, &home, &["update"]));
    let token = p["cleanup"]["confirm_token"].as_str().unwrap();
    let top = p["plan_digest"].as_str().unwrap();
    let r = result(&run(
        &exe,
        &home,
        &["update", "--confirm", top, "--confirm-cleanup", token],
    ));
    assert_eq!(r["mode"], json!("applied"), "{r}");
    assert_eq!(r["applied"], json!(true), "{r}");
    assert_eq!(r["cleanup_result"]["state"], json!("committed"), "{r}");
    let stamp = r["cleanup_result"]["stamp"].as_str().unwrap();
    assert_eq!(
        r["revert"]["cleanup"],
        json!(format!("moai cleanup-revert-plan <HOME> --stamp {stamp}"))
    );
    assert!(home.join(".claude/skills/moai/SKILL.md").is_file());
    assert!(!home.join(".claude/skills/moai-proven").exists());
}

#[test]
fn a_package_changed_after_the_preview_is_refused_with_nothing_written() {
    let (_t, home, exe, pkg) = setup();
    let p = result(&run(&exe, &home, &["update"]));
    let top = p["plan_digest"].as_str().unwrap().to_string();
    let src = pkg.join("references/home/routing.md");
    let text = fs::read_to_string(&src).unwrap();
    fs::write(&src, text.replace("before asking", "before you ask")).unwrap();
    let before = state(&home);
    let out = run(&exe, &home, &["update", "--confirm", &top]);
    assert!(!out.status.success());
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert!(
        v["error"]["message"].as_str().unwrap().contains("stale"),
        "{v}"
    );
    assert_eq!(state(&home), before, "nothing was written");
}

/// SPEC-20261122-html-report-install AC-1: a release that carries `skills/moai-html-report`
/// previews it as its own part, and the terminal screen shows it as its own skills row beside
/// the moai skill's row. A preview writes nothing.
#[test]
fn the_preview_lists_the_html_report_skill_on_its_own_row_beside_the_moai_skill() {
    let (_t, home, exe, pkg) = setup();
    let src = with_html_report(&pkg);
    let before = state(&home);
    let p = result(&run(&exe, &home, &["update"]));
    assert_eq!(state(&home), before, "a preview writes nothing");
    let mut with: Vec<&str> = PREVIEW_KEYS.to_vec();
    with.push("extra_skills");
    assert_eq!(keys(&p), with, "{p}");
    let parts = p["extra_skills"]["parts"].as_array().unwrap();
    assert_eq!(parts.len(), 1, "{p}");
    let part = &parts[0];
    assert_eq!(part["skill"], json!(HTML_REPORT), "{p}");
    assert_eq!(part["host"], json!("claude"), "{p}");
    assert_eq!(part["preflight"], json!("ok"), "{p}");
    assert_eq!(
        part["projection_root"],
        json!(home.join(".claude/skills").join(HTML_REPORT)),
        "{p}"
    );
    let files = part["files"].as_array().unwrap();
    assert_eq!(files.len(), tree(&src).len(), "{p}");
    assert!(files.iter().all(|f| f["action"] == json!("create")), "{p}");
    // The moai skill keeps its own plan; the html-report files are not in it.
    let moai = &p["plan"]["hosts"][0];
    assert_eq!(moai["projection_root"], json!(skill_dir(&home, "claude")));
    let view = moai_cli::update_tux::View {
        painter: moai_cli::ui::Painter::plain(),
        width: 120,
        home: home.display().to_string(),
        project: None,
    };
    let text = common::screens::strip_colors(&moai_cli::update_tux::screen(
        &p,
        moai_cli::update_tux::Asked::NotAsked,
        &view,
    ));
    let row = |line: &str| -> Vec<String> {
        line.split_whitespace()
            .skip(1)
            .take(3)
            .map(str::to_string)
            .collect()
    };
    assert!(
        text.lines()
            .any(|l| row(l) == ["skills", "claude", HTML_REPORT] && l.contains("new")),
        "{text}"
    );
    assert!(
        text.lines().any(|l| {
            let r = row(l);
            r.len() == 3 && r[..2] == ["skills", "claude"] && r[2] != HTML_REPORT
        }),
        "{text}"
    );
}

/// SPEC-20261122-html-report-install AC-2: a confirmed update copies both skill folders byte
/// for byte, the html-report skill stays out of the `/` menu (`user-invocable: false`), its
/// receipt is committed under its own name, and the next run has nothing to apply.
#[test]
fn confirm_installs_both_skill_folders_and_the_html_report_is_not_user_invocable() {
    let (_t, home, exe, pkg) = setup();
    let src = with_html_report(&pkg);
    let p = result(&run(&exe, &home, &["update"]));
    let top = p["plan_digest"].as_str().unwrap();
    let r = result(&run(&exe, &home, &["update", "--confirm", top]));
    assert_eq!(r["applied"], json!(true), "{r}");
    assert_eq!(r["extra_skills_result"]["state"], json!("applied"), "{r}");
    let skills = home.join(".claude/skills");
    assert_eq!(tree(&skills.join("moai")), tree(&pkg));
    let installed = skills.join(HTML_REPORT);
    assert_eq!(tree(&installed), tree(&src));
    let text = fs::read_to_string(installed.join("SKILL.md")).unwrap();
    let head = text.strip_prefix("---\n").unwrap();
    let head = &head[..head.find("\n---\n").unwrap()];
    let lines: Vec<&str> = head.lines().map(str::trim_end).collect();
    assert!(
        lines.contains(&format!("name: {HTML_REPORT}").as_str()),
        "{head}"
    );
    assert!(lines.contains(&"user-invocable: false"), "{head}");
    let receipts = html_report_receipts(&home);
    assert_eq!(receipts.len(), 1, "{receipts:?}");
    assert_eq!(receipts[0]["status"], json!("committed"));
    assert_eq!(receipts[0]["host"], json!("claude"));
    assert_eq!(
        receipts[0]["owned"].as_array().unwrap().len(),
        tree(&src).len()
    );
    let again = result(&run(&exe, &home, &["update"]));
    assert_eq!(again["nothing_to_apply"], json!(true), "{again}");
    let part = &again["extra_skills"]["parts"][0];
    assert_eq!(part["preflight"], json!("ok"), "{again}");
    assert!(
        part["files"]
            .as_array()
            .unwrap()
            .iter()
            .all(|f| f["action"] == json!("unchanged")),
        "{again}"
    );
}

/// SPEC-20261122-html-report-install AC-3: a `moai-html-report` folder whose package file no
/// moai receipt owns blocks only that part; the reason names the folder, the confirmed update
/// leaves its bytes as they are and still installs the moai skill.
#[test]
fn an_html_report_folder_moai_did_not_install_is_named_and_left_untouched() {
    let (_t, home, exe, pkg) = setup();
    with_html_report(&pkg);
    let foreign = home.join(".claude/skills").join(HTML_REPORT);
    fs::create_dir_all(&foreign).unwrap();
    let mine = "---\nname: moai-html-report\ndescription: my own copy\n---\nmine\n";
    fs::write(foreign.join("SKILL.md"), mine).unwrap();
    let p = result(&run(&exe, &home, &["update"]));
    let part = &p["extra_skills"]["parts"][0];
    assert_eq!(part["preflight"], json!("blocked"), "{p}");
    let why = part["blocked_reason"].as_str().unwrap();
    assert!(why.contains(foreign.to_str().unwrap()), "{why}");
    assert!(why.contains("moai did not install"), "{why}");
    assert_eq!(p["plan"]["hosts"][0]["preflight"], json!("ok"), "{p}");
    let top = p["plan_digest"].as_str().unwrap();
    let r = result(&run(&exe, &home, &["update", "--confirm", top]));
    assert_eq!(r["applied"], json!(true), "{r}");
    assert_eq!(
        r["extra_skills_result"]["parts"][0]["status"],
        json!("blocked"),
        "{r}"
    );
    assert_eq!(
        tree(&foreign),
        [(PathBuf::from("SKILL.md"), mine.as_bytes().to_vec())]
    );
    assert!(skill_dir(&home, "claude").join("SKILL.md").is_file());
    assert!(html_report_receipts(&home).is_empty());
}

/// SPEC-20261122-html-report-install: the next version folder takes the html-report baseline
/// from the receipt the earlier folder committed, as the moai skill does, so the folder moai
/// installed is not taken for one it did not.
#[test]
fn a_second_version_folder_adopts_the_html_report_receipt_instead_of_blocking_it() {
    let (_t, home, exe, pkg) = setup();
    with_html_report(&pkg);
    let p = result(&run(&exe, &home, &["update"]));
    let top = p["plan_digest"].as_str().unwrap();
    result(&run(&exe, &home, &["update", "--confirm", top]));
    // Beside the first version folder, whatever the state folder is named: exe is
    // `<versions>/a/bin/moai-cli`.
    let versions = exe
        .parent()
        .and_then(Path::parent)
        .and_then(Path::parent)
        .unwrap();
    let next = versions.join("b");
    fs::create_dir_all(next.join("bin")).unwrap();
    let exe_b = next.join("bin/moai-cli");
    common::copy_executable(Path::new(EXE), &exe_b);
    let pkg_b = next.join("skills/moai");
    copy_tree(&package(), &pkg_b);
    with_html_report(&pkg_b);
    copy_tree(
        &Path::new(env!("CARGO_MANIFEST_DIR")).join("agents"),
        &next.join("agents"),
    );
    let p = result(&run(&exe_b, &home, &["update"]));
    let part = &p["extra_skills"]["parts"][0];
    assert_eq!(part["preflight"], json!("ok"), "{p}");
    assert!(part["receipt_from"].is_string(), "{p}");
    assert!(
        part["files"]
            .as_array()
            .unwrap()
            .iter()
            .all(|f| f["action"] == json!("unchanged")),
        "{p}"
    );
}

/// [`run`] with one more variable set, such as the debug builds' fault seam.
fn run_with(exe: &Path, home: &Path, args: &[&str], var: (&str, &str)) -> Output {
    Command::new(exe)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .env(var.0, var.1)
        .current_dir(home)
        .stdin(Stdio::null())
        .output()
        .unwrap()
}

/// The fault seam that fails the last publish of a skill folder beside the moai skill.
const LAST_PUBLISH_FAULT: (&str, &str) = ("MOAI_TEST_FAULT", "extra_skill_last_publish");

/// The receipt files of the html-report skill under HOME's state folder.
fn html_report_receipt_files(home: &Path) -> Vec<PathBuf> {
    let tail = Path::new("hosts").join(HTML_REPORT).join("claude.json");
    state(home)
        .into_iter()
        .map(|(p, _)| p)
        .filter(|p| p.ends_with(&tail))
        .collect()
}

/// The installed html-report folder under the temporary HOME.
fn html_report_dir(home: &Path) -> PathBuf {
    home.join(".claude/skills").join(HTML_REPORT)
}

/// The preview's html-report part and its blocked reason (empty when it is not blocked).
fn html_report_part(p: &Value) -> (Value, String) {
    let part = p["extra_skills"]["parts"][0].clone();
    let why = part["blocked_reason"].as_str().unwrap_or("").to_string();
    (part, why)
}

/// SPEC-20261122-html-report-install AC-6: a path of a package file that the apply would fail on
/// blocks the part at preview time and the reason names that path: a regular file named
/// `references` where the package needs a folder, a folder named `SKILL.md` where it needs a
/// file, and a folder moai cannot write in. The confirmed update writes nothing there.
#[test]
fn a_path_the_apply_would_fail_on_blocks_the_html_report_part_at_preview() {
    let (_t, home, exe, pkg) = setup();
    with_html_report(&pkg);
    let dest = html_report_dir(&home);
    fs::create_dir_all(&dest).unwrap();
    let blocker = dest.join("references");
    fs::write(&blocker, "not a folder\n").unwrap();
    let p = result(&run(&exe, &home, &["update"]));
    let (part, why) = html_report_part(&p);
    assert_eq!(part["preflight"], json!("blocked"), "{p}");
    assert!(why.contains(blocker.to_str().unwrap()), "{why}");
    let top = p["plan_digest"].as_str().unwrap();
    let r = result(&run(&exe, &home, &["update", "--confirm", top]));
    assert_eq!(
        r["extra_skills_result"]["parts"][0]["status"],
        json!("blocked"),
        "{r}"
    );
    assert_eq!(
        tree(&dest),
        [(PathBuf::from("references"), b"not a folder\n".to_vec())]
    );
    assert!(html_report_receipts(&home).is_empty());
    // A folder where the package puts a file.
    fs::remove_file(&blocker).unwrap();
    let folder = dest.join("SKILL.md");
    fs::create_dir_all(&folder).unwrap();
    let (part, why) = html_report_part(&result(&run(&exe, &home, &["update"])));
    assert_eq!(part["preflight"], json!("blocked"), "{part}");
    assert!(why.contains(folder.to_str().unwrap()), "{why}");
    // A folder of the package that moai cannot write in.
    fs::remove_dir(&folder).unwrap();
    fs::create_dir_all(&blocker).unwrap();
    use std::os::unix::fs::PermissionsExt;
    fs::set_permissions(&blocker, fs::Permissions::from_mode(0o555)).unwrap();
    let (part, why) = html_report_part(&result(&run(&exe, &home, &["update"])));
    fs::set_permissions(&blocker, fs::Permissions::from_mode(0o755)).unwrap();
    assert_eq!(part["preflight"], json!("blocked"), "{part}");
    assert!(why.contains(blocker.to_str().unwrap()), "{why}");
    assert!(html_report_receipts(&home).is_empty());
}

/// SPEC-20261122-html-report-install AC-7: a publish that fails during the apply (the fault seam
/// fails the last one) undoes this apply's own files: the files it created are removed with the
/// folders it made, the files it replaced get their earlier bytes back, and the receipt is
/// `recovered`, not left pending. No half-installed skill stays and the next preview is not
/// blocked.
#[test]
fn a_failed_html_report_apply_undoes_its_own_files_and_the_next_preview_is_not_blocked() {
    let (_t, home, exe, pkg) = setup();
    let src = with_html_report(&pkg);
    let dest = html_report_dir(&home);
    let p = result(&run(&exe, &home, &["update"]));
    let top = p["plan_digest"].as_str().unwrap().to_string();
    let r = result(&run_with(
        &exe,
        &home,
        &["update", "--confirm", &top],
        LAST_PUBLISH_FAULT,
    ));
    let part = &r["extra_skills_result"]["parts"][0];
    assert_eq!(part["status"], json!("failed"), "{r}");
    assert_eq!(part["receipt_status"], json!("recovered"), "{r}");
    assert!(!dest.exists(), "a half-installed skill: {:?}", tree(&dest));
    // The way back it prints is the undo that ran, not `recovery-plan`.
    let revert = r["revert"]["extra_skills"].as_str().unwrap();
    assert!(revert.contains("removes the files it created"), "{r}");
    assert!(!revert.contains("recovery-plan"), "{r}");
    let receipts = html_report_receipts(&home);
    assert_eq!(receipts.len(), 1, "{receipts:?}");
    assert_eq!(receipts[0]["status"], json!("recovered"), "{receipts:?}");
    // The next preview plans the whole folder again and its confirm installs it.
    let p = result(&run(&exe, &home, &["update"]));
    let (part, why) = html_report_part(&p);
    assert_eq!(part["preflight"], json!("ok"), "{why}");
    assert!(
        part["files"]
            .as_array()
            .unwrap()
            .iter()
            .all(|f| f["action"] == json!("create")),
        "{p}"
    );
    let top = p["plan_digest"].as_str().unwrap();
    let r = result(&run(&exe, &home, &["update", "--confirm", top]));
    assert_eq!(r["extra_skills_result"]["state"], json!("applied"), "{r}");
    assert_eq!(tree(&dest), tree(&src));
    // An update that fails puts back the bytes it replaced.
    let installed = tree(&dest);
    for rel in ["SKILL.md", "references/fonts.md"] {
        let mut bytes = fs::read(src.join(rel)).unwrap();
        bytes.extend_from_slice(b"\nchanged in the next release\n");
        fs::write(src.join(rel), bytes).unwrap();
    }
    let updates = |p: &Value| {
        p["extra_skills"]["parts"][0]["files"]
            .as_array()
            .unwrap()
            .iter()
            .filter(|f| f["action"] == json!("update"))
            .count()
    };
    let p = result(&run(&exe, &home, &["update"]));
    assert_eq!(updates(&p), 2, "{p}");
    let top = p["plan_digest"].as_str().unwrap().to_string();
    let r = result(&run_with(
        &exe,
        &home,
        &["update", "--confirm", &top],
        LAST_PUBLISH_FAULT,
    ));
    let part = &r["extra_skills_result"]["parts"][0];
    assert_eq!(part["status"], json!("failed"), "{r}");
    assert_eq!(part["receipt_status"], json!("recovered"), "{r}");
    assert_eq!(tree(&dest), installed, "the replaced bytes are back");
    assert_eq!(html_report_receipts(&home)[0]["status"], json!("recovered"));
    let p = result(&run(&exe, &home, &["update"]));
    let (part, why) = html_report_part(&p);
    assert_eq!(part["preflight"], json!("ok"), "{why}");
    assert_eq!(updates(&p), 2, "{p}");
}

/// SPEC-20261122-html-report-install AC-8: the confirm digest binds the skill folders beside
/// the moai skill: a file of `skills/moai-html-report` changed after the preview is refused as a
/// stale plan and nothing is written, the moai skill included.
#[test]
fn an_html_report_package_changed_after_the_preview_is_refused_with_nothing_written() {
    let (_t, home, exe, pkg) = setup();
    let src = with_html_report(&pkg);
    let p = result(&run(&exe, &home, &["update"]));
    let top = p["plan_digest"].as_str().unwrap().to_string();
    let file = src.join("references/fonts.md");
    let mut bytes = fs::read(&file).unwrap();
    bytes.extend_from_slice(b"\nchanged after the preview\n");
    fs::write(&file, bytes).unwrap();
    let before = state(&home);
    let out = run(&exe, &home, &["update", "--confirm", &top]);
    assert!(!out.status.success());
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert!(
        v["error"]["message"].as_str().unwrap().contains("stale"),
        "{v}"
    );
    assert_eq!(state(&home), before, "nothing was written");
}

/// SPEC-20261122-html-report-install AC-9: a pending receipt of the html-report skill blocks the
/// part, in this version folder and in the next one. The reason names the receipt and the folder
/// and gives the way out; it does not send the user to `recovery-plan`, which reads only the moai
/// skill's receipts. The confirmed update leaves the folder as it is.
#[test]
fn a_pending_html_report_receipt_blocks_the_part_and_names_the_way_out() {
    let (_t, home, exe, pkg) = setup();
    with_html_report(&pkg);
    let dest = html_report_dir(&home);
    let p = result(&run(&exe, &home, &["update"]));
    let top = p["plan_digest"].as_str().unwrap();
    result(&run(&exe, &home, &["update", "--confirm", top]));
    let files = html_report_receipt_files(&home);
    assert_eq!(files.len(), 1, "{files:?}");
    let receipt = &files[0];
    let mut v: Value = serde_json::from_slice(&fs::read(receipt).unwrap()).unwrap();
    v["status"] = json!("pending");
    fs::write(receipt, serde_json::to_vec_pretty(&v).unwrap()).unwrap();
    let way_out = |why: &str| {
        assert!(why.contains(receipt.to_str().unwrap()), "{why}");
        assert!(why.contains(dest.to_str().unwrap()), "{why}");
        assert!(why.contains("move the folder aside"), "{why}");
        assert!(!why.contains("recovery-plan"), "{why}");
    };
    let p = result(&run(&exe, &home, &["update"]));
    let (part, why) = html_report_part(&p);
    assert_eq!(part["preflight"], json!("blocked"), "{p}");
    way_out(&why);
    let installed = tree(&dest);
    let top = p["plan_digest"].as_str().unwrap();
    let r = result(&run(&exe, &home, &["update", "--confirm", top]));
    assert_eq!(
        r["extra_skills_result"]["parts"][0]["status"],
        json!("blocked"),
        "{r}"
    );
    assert_eq!(tree(&dest), installed);
    // The next version folder finds the earlier folder's pending receipt the same way: exe is
    // `<versions>/a/bin/moai-cli`, whatever the state folder is named.
    let next = exe
        .parent()
        .and_then(Path::parent)
        .and_then(Path::parent)
        .unwrap()
        .join("b");
    fs::create_dir_all(next.join("bin")).unwrap();
    let exe_b = next.join("bin/moai-cli");
    common::copy_executable(Path::new(EXE), &exe_b);
    let pkg_b = next.join("skills/moai");
    copy_tree(&package(), &pkg_b);
    with_html_report(&pkg_b);
    copy_tree(
        &Path::new(env!("CARGO_MANIFEST_DIR")).join("agents"),
        &next.join("agents"),
    );
    let (part, why) = html_report_part(&result(&run(&exe_b, &home, &["update"])));
    assert_eq!(part["preflight"], json!("blocked"), "{part}");
    way_out(&why);
}

/// SPEC-20261122-html-report-install AC-10: a symlink or a file where the html-report folder
/// goes blocks the part, and the reason names that folder.
#[test]
fn a_symlink_or_a_file_at_the_html_report_folder_is_named_in_the_reason() {
    let (_t, home, exe, pkg) = setup();
    with_html_report(&pkg);
    let dest = html_report_dir(&home);
    let elsewhere = home.join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    fs::create_dir_all(dest.parent().unwrap()).unwrap();
    std::os::unix::fs::symlink(&elsewhere, &dest).unwrap();
    let before = state(&home);
    let (part, why) = html_report_part(&result(&run(&exe, &home, &["update"])));
    assert_eq!(part["preflight"], json!("blocked"), "{part}");
    assert!(why.contains(dest.to_str().unwrap()), "{why}");
    assert!(why.contains("symlink"), "{why}");
    assert_eq!(state(&home), before, "a preview writes nothing");
    fs::remove_file(&dest).unwrap();
    fs::write(&dest, "a file\n").unwrap();
    let (part, why) = html_report_part(&result(&run(&exe, &home, &["update"])));
    assert_eq!(part["preflight"], json!("blocked"), "{part}");
    assert!(why.contains(dest.to_str().unwrap()), "{why}");
    assert!(why.contains("not a folder"), "{why}");
}

/// SPEC-20261122-html-report-install AC-11: on a first install the html-report receipt records
/// that the top skills folder did not exist before this run, though the moai skill's apply of the
/// same run made it first, so the result asks for the skill reload.
#[test]
fn a_first_html_report_install_records_that_the_skills_folder_was_new_to_this_run() {
    let (_t, home, exe, pkg) = setup();
    with_html_report(&pkg);
    assert!(!home.join(".claude/skills").exists());
    let p = result(&run(&exe, &home, &["update"]));
    let top = p["plan_digest"].as_str().unwrap();
    let r = result(&run(&exe, &home, &["update", "--confirm", top]));
    let part = &r["extra_skills_result"]["parts"][0];
    assert_eq!(part["status"], json!("applied"), "{r}");
    assert_eq!(part["reload_required"], json!("yes"), "{r}");
    let receipts = html_report_receipts(&home);
    assert_eq!(
        receipts[0]["top_skills_existed_before"],
        json!(false),
        "{receipts:?}"
    );
}

/// SPEC-20261122-claude-md-home-routing REQ-5 and REQ-8, through the real `moai update`: the
/// preview writes nothing, the confirm writes the rules files and leaves the earlier
/// `moai-routing` block of the user's `CLAUDE.md` byte for byte, and only the separate pick
/// (`--remove-old-block` of the home routing commands) removes exactly that block.
#[test]
fn update_confirm_never_removes_the_old_block_only_its_own_pick_does() {
    let (_t, home, exe, pkg) = setup();
    let block = "<!-- moai:managed:begin id=moai-routing -->\n## MoAI\n\nold text (/moai)\n\
                 <!-- moai:managed:end id=moai-routing -->\n";
    let text = format!("# Mine\n\nKeep this.\n\n{block}\nAfter.\n");
    let claude_md = home.join(".claude/CLAUDE.md");
    fs::create_dir_all(claude_md.parent().unwrap()).unwrap();
    fs::write(&claude_md, &text).unwrap();
    let state_file = moai_cli::home_root::path(&home).join("home-routing/claude.json");
    fs::create_dir_all(state_file.parent().unwrap()).unwrap();
    fs::write(
        &state_file,
        json!({"schema_version": 1, "block_sha256": moai_cli::digest(block.as_bytes())})
            .to_string(),
    )
    .unwrap();
    let old_block_action = |p: &Value| -> Value {
        p["routing"]["hosts"]
            .as_array()
            .unwrap()
            .iter()
            .find(|h| h["host"] == json!("claude"))
            .unwrap()["old_block"]["action"]
            .clone()
    };
    // The preview names the earlier block as waiting for the rules file and writes nothing.
    let before = state(&home);
    let p = result(&run(&exe, &home, &["update"]));
    assert_eq!(old_block_action(&p), json!("waiting"), "{p}");
    assert_eq!(state(&home), before, "a preview writes nothing");
    assert!(!home.join(".claude/rules/moai.md").exists());
    // The confirm writes the rules files; the user's file is untouched.
    let top = p["plan_digest"].as_str().unwrap().to_string();
    let r = result(&run(&exe, &home, &["update", "--confirm", &top]));
    assert_eq!(r["routing_result"]["state"], json!("applied"), "{r}");
    assert!(home.join(".claude/rules/moai.md").is_file());
    assert!(home.join(".claude/rules/moai/flow.md").is_file());
    assert_eq!(fs::read_to_string(&claude_md).unwrap(), text);
    // With the rules file in place the block is only offered; a second confirm still keeps it.
    let p = result(&run(&exe, &home, &["update"]));
    assert_eq!(old_block_action(&p), json!("offered"), "{p}");
    if p["nothing_to_apply"] == json!(false) {
        let top = p["plan_digest"].as_str().unwrap();
        result(&run(&exe, &home, &["update", "--confirm", top]));
    }
    assert_eq!(fs::read_to_string(&claude_md).unwrap(), text);
    // The separate pick removes exactly the block and nothing else.
    let h = home.to_str().unwrap();
    let pkg = pkg.to_str().unwrap();
    let plan = result(&run(
        &exe,
        &home,
        &["home-routing-plan", h, pkg, "--remove-old-block"],
    ));
    let digest = plan["plan_digest"].as_str().unwrap();
    let applied = result(&run(
        &exe,
        &home,
        &[
            "home-routing-apply",
            h,
            pkg,
            "--plan-digest",
            digest,
            "--permission-reference",
            "test-permission",
            "--remove-old-block",
        ],
    ));
    assert_eq!(applied["state"], json!("applied"), "{applied}");
    assert_eq!(
        fs::read_to_string(&claude_md).unwrap(),
        text.replacen(block, "", 1)
    );
}
