//! SPEC-20261123-temp-home-e2e-regression: the landed Claude Code redesign run end to end through
//! the real binary, in a temporary version folder, a temporary HOME and temporary projects. Every
//! command is a copy of the built binary with a clean environment; the real Claude config and
//! MoAI state folders are never read or written. A real `claude` is never started (the operator steps are
//! in the SPEC Result as `not_run`).
#![cfg(unix)]

mod common;

use common::CleanEnv;
use moai_cli::tree_digest;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};
use tempfile::TempDir;

/// The tree of this repository that the test installs as the package.
fn repo(rel: &str) -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join(rel)
}

fn copy_tree(from: &Path, to: &Path) {
    fs::create_dir_all(to).unwrap();
    for entry in fs::read_dir(from).unwrap() {
        let entry = entry.unwrap();
        let (src, dst) = (entry.path(), to.join(entry.file_name()));
        if entry.file_type().unwrap().is_dir() {
            copy_tree(&src, &dst);
        } else {
            fs::copy(&src, &dst).unwrap();
        }
    }
}

/// One temporary world: a root folder that holds the version folder, the HOME and the projects.
struct World {
    _tmp: TempDir,
    /// The canonical root: every path a command reports has to lie under it.
    root: PathBuf,
    home: PathBuf,
    /// `<root>/ver/bin/moai-cli`, with `skills/moai` and `agents` beside its `bin`.
    exe: PathBuf,
}

impl World {
    fn new() -> Self {
        let tmp = TempDir::new().unwrap();
        let root = fs::canonicalize(tmp.path()).unwrap();
        let home = root.join("home");
        fs::create_dir_all(&home).unwrap();
        let ver = root.join("ver");
        let (cli, _moai) = common::install_both_names(&ver.join("bin"));
        copy_tree(&repo("skills/moai"), &ver.join("skills/moai"));
        copy_tree(&repo("agents"), &ver.join("agents"));
        World {
            _tmp: tmp,
            root,
            home,
            exe: cli,
        }
    }

    /// `moai-cli <args>` in `cwd` with a clean environment whose HOME names the temporary home.
    fn run_in(&self, cwd: &Path, args: &[&str]) -> Output {
        self.run_with(cwd, &[], None, args)
    }

    /// The same, with extra variables and bytes on standard input (none: closed).
    fn run_with(
        &self,
        cwd: &Path,
        envs: &[(&str, &str)],
        stdin: Option<&[u8]>,
        args: &[&str],
    ) -> Output {
        let mut cmd = Command::new(&self.exe);
        cmd.args(args)
            .clean_env()
            .env("PATH", "/usr/bin:/bin")
            .env("HOME", &self.home)
            .env("MOAI_ACCOUNT_HOME", &self.home)
            .env("LC_ALL", "C")
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_NOSYSTEM", "1")
            .envs(envs.iter().copied())
            .current_dir(cwd)
            .stdin(if stdin.is_some() {
                Stdio::piped()
            } else {
                Stdio::null()
            })
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        let mut child = cmd.spawn().expect("spawn moai-cli");
        if let Some(bytes) = stdin {
            common::write_stdin(child.stdin.take().unwrap(), bytes);
        }
        child.wait_with_output().expect("run moai-cli")
    }

    fn run(&self, args: &[&str]) -> Output {
        self.run_in(&self.home, args)
    }

    fn digest(&self) -> String {
        tree_digest(&self.home, &[]).unwrap()
    }
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "envelope JSON on stdout: {}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn ok(out: &Output) -> Value {
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert!(out.status.success(), "{v}");
    v["result"].clone()
}

// ---- helpers ------------------------------------------------------------------------------------

/// Every whole word of a string in `v` that starts with `/`, names a folder that exists at the
/// top of the file system (so `/status` is no path) and does not lie under `root`.
fn outside_paths(v: &Value, root: &Path, out: &mut Vec<String>) {
    match v {
        Value::String(s) => {
            for word in s.split(|c: char| c.is_whitespace() || "`\"',;()[]".contains(c)) {
                let Some(rest) = word.strip_prefix('/') else {
                    continue;
                };
                let first = rest.split('/').next().unwrap_or("");
                if first.is_empty() || !Path::new("/").join(first).exists() {
                    continue;
                }
                if !Path::new(word).starts_with(root) {
                    out.push(word.to_string());
                }
            }
        }
        Value::Array(a) => a.iter().for_each(|x| outside_paths(x, root, out)),
        Value::Object(m) => m.values().for_each(|x| outside_paths(x, root, out)),
        _ => {}
    }
}

impl World {
    /// Fails naming every path of `answer` that lies outside the temporary root (REQ-2).
    fn assert_inside(&self, answer: &Value) {
        let mut bad = Vec::new();
        outside_paths(answer, &self.root, &mut bad);
        assert!(
            bad.is_empty(),
            "paths outside {}: {bad:?}",
            self.root.display()
        );
    }

    /// `moai-cli update`: the preview, which writes nothing.
    fn preview(&self) -> Value {
        let before = self.digest();
        let r = ok(&self.run(&["update"]));
        assert_eq!(r["mode"], json!("preview"), "{r}");
        assert_eq!(r["applied"], json!(false), "{r}");
        assert_eq!(self.digest(), before, "a preview writes nothing");
        self.assert_inside(&r);
        r
    }

    /// `moai-cli update --confirm <digest>`.
    fn confirm(&self, digest: &str) -> Value {
        let r = ok(&self.run(&["update", "--confirm", digest]));
        assert_eq!(r["mode"], json!("applied"), "{r}");
        self.assert_inside(&r);
        // Receipts carry second-resolution timestamps.
        std::thread::sleep(std::time::Duration::from_millis(1100));
        r
    }

    fn pkg(&self, rel: &str) -> PathBuf {
        self.root.join("ver/skills/moai").join(rel)
    }

    fn agent_src(&self, rel: &str) -> PathBuf {
        self.root.join("ver/agents").join(rel)
    }
}

fn digest_of(preview: &Value) -> String {
    preview["plan_digest"].as_str().unwrap().to_string()
}

/// The files a preview lists, as `(layer, path, action)`; the path is absolute.
fn listed(preview: &Value) -> Vec<(String, PathBuf, String)> {
    let mut out = Vec::new();
    let s = |v: &Value| v.as_str().unwrap().to_string();
    for h in preview["plan"]["hosts"].as_array().unwrap() {
        let root = PathBuf::from(s(&h["projection_root"]));
        for f in h["files"].as_array().unwrap() {
            out.push(("skill".into(), root.join(s(&f["path"])), s(&f["action"])));
        }
    }
    for h in preview["agents"]["hosts"].as_array().unwrap() {
        let dir = PathBuf::from(s(&h["dir"]));
        for f in h["files"].as_array().unwrap() {
            out.push(("agents".into(), dir.join(s(&f["file"])), s(&f["action"])));
        }
    }
    for h in preview["routing"]["hosts"].as_array().unwrap() {
        for part in ["section", "flow"] {
            let p = &h[part];
            out.push((
                "routing".into(),
                PathBuf::from(s(&p["path"])),
                s(&p["action"]),
            ));
        }
    }
    for t in preview["style"]["targets"].as_array().unwrap() {
        out.push((
            "style".into(),
            PathBuf::from(s(&t["path"])),
            s(&t["action"]),
        ));
    }
    out
}

/// Every regular file under `dir`, relative to it, sorted.
fn files_under(dir: &Path) -> Vec<PathBuf> {
    let mut out = Vec::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap() {
            let p = e.unwrap().path();
            let m = fs::symlink_metadata(&p).unwrap();
            if m.is_dir() {
                stack.push(p);
            } else {
                out.push(p.strip_prefix(dir).unwrap().to_path_buf());
            }
        }
    }
    out.sort();
    out
}

fn is_state_folder(rel: &Path) -> bool {
    rel.starts_with(moai_cli::state_root::MOAI_DIR) || rel.starts_with(moai_cli::PROJECT_STATE_DIR)
}

// ---- scenarios ----------------------------------------------------------------------------------

/// REQ-2 (the guard of the helper): a path outside the root is named, one inside is not, a
/// slash command is no path.
#[test]
fn scenario_root_checker_names_a_path_outside_the_root() {
    let w = World::new();
    let mut bad = Vec::new();
    let inside = w.home.join(".claude/rules/moai.md");
    outside_paths(
        &json!({"a": inside.display().to_string(), "b": ["see /usr/bin/git now", "`/status` lists"]}),
        &w.root,
        &mut bad,
    );
    assert_eq!(bad, vec!["/usr/bin/git".to_string()]);
}

/// REQ-1, REQ-2: every answer of a first preview and its confirmation lies inside the root.
#[test]
fn scenario_root_every_answer_stays_inside_the_temporary_root() {
    let w = World::new();
    let p = w.preview();
    let r = w.confirm(&digest_of(&p));
    w.assert_inside(&r);
    assert!(w.home.starts_with(&w.root));
}

/// REQ-3, REQ-4: a new install writes what the preview listed and nothing else outside MoAI's
/// own state folder; a wrong digest writes nothing.
#[test]
fn scenario_new_install_writes_exactly_the_previewed_files() {
    let w = World::new();
    let project = w.root.join("work/other");
    fs::create_dir_all(&project).unwrap();
    fs::write(project.join("main.rs"), "fn main() {}\n").unwrap();
    let project_before = tree_digest(&project, &[]).unwrap();
    let p = w.preview();
    let rows = listed(&p);
    for layer in ["skill", "agents", "routing", "style"] {
        assert!(
            rows.iter().any(|(l, _, _)| l == layer),
            "no {layer} row: {p}"
        );
    }
    assert!(
        rows.iter().all(|(_, _, a)| a == "create" || a == "set"),
        "{rows:?}"
    );

    // A wrong digest writes nothing.
    let before = w.digest();
    let wrong = format!("sha256:{}", "0".repeat(64));
    let out = w.run(&["update", "--confirm", &wrong]);
    assert!(!out.status.success());
    assert_eq!(envelope(&out)["ok"], json!(false));
    assert_eq!(w.digest(), before, "a foreign digest wrote something");

    w.confirm(&digest_of(&p));
    // The installed agent files are what this tree renders from its sources: Claude Code reads
    // `name` and `description` from their frontmatter.
    let defs = moai_cli::agent_defs::package_agents(
        &w.root.join("ver/agents"),
        &w.root.join("ver/skills/moai"),
    )
    .unwrap();
    assert_eq!(defs.len(), 5, "the five base agents");
    for def in &defs {
        let rendered = moai_cli::agent_defs::render(def, "claude").unwrap();
        // Claude Code keeps the base agents in `agents/moai/`.
        let installed = fs::read_to_string(
            w.home
                .join(".claude/agents/moai")
                .join(format!("{}.md", def.name)),
        )
        .unwrap();
        assert_eq!(installed, rendered.text, "{}", rendered.path);
        assert!(
            installed.starts_with(&format!("---\nname: {}\n", def.name)),
            "{}",
            rendered.path
        );
    }
    let home_files = files_under(&w.home);
    for (layer, path, _) in &rows {
        assert!(
            path.is_file(),
            "{layer}: {} was not written",
            path.display()
        );
    }
    let listed_rel: Vec<PathBuf> = rows
        .iter()
        .map(|(_, p, _)| p.strip_prefix(&w.home).unwrap().to_path_buf())
        .collect();
    for rel in &home_files {
        assert!(
            listed_rel.contains(rel)
                || is_state_folder(rel)
                || rel.starts_with(".metadata_never_index"),
            "a file the preview did not list: {}",
            rel.display()
        );
    }
    assert_eq!(tree_digest(&project, &[]).unwrap(), project_before);
}

/// REQ-3: an update started from inside a git project that has no MoAI profile writes nothing
/// into that project, `.git` included.
#[test]
fn scenario_non_moai_project_is_untouched_by_an_update_started_inside_it() {
    let w = World::new();
    let project = w.root.join("work/plain-git");
    write(&project.join("src/main.rs"), "fn main() {}\n");
    git(&project, &["init", "-q", "-b", "main"]);
    git(&project, &["add", "src/main.rs"]);
    git(&project, &["commit", "-q", "-m", "init"]);
    let before = tree_digest(&project, &[]).unwrap();
    let out = w.run_in(&project, &["update"]);
    let p = ok(&out);
    assert_eq!(p["mode"], json!("preview"), "{p}");
    assert_eq!(
        tree_digest(&project, &[]).unwrap(),
        before,
        "the preview wrote"
    );
    let confirmed = ok(&w.run_in(&project, &["update", "--confirm", &digest_of(&p)]));
    w.assert_inside(&confirmed);
    assert_eq!(
        tree_digest(&project, &[]).unwrap(),
        before,
        "the update wrote into a project that is not MoAI's"
    );
    assert!(w.home.join(".claude/skills/moai/SKILL.md").is_file());
}

/// REQ-5: a changed package writes only the changed files; a run of the same package after
/// that leaves the home as it was.
#[test]
fn scenario_two_updates_write_only_what_changed() {
    let w = World::new();
    let p = w.preview();
    w.confirm(&digest_of(&p));

    // Version 2 of the package: one skill file changed, one added, one agent changed.
    let audit = w.pkg("references/audit.md");
    let mut text = fs::read_to_string(&audit).unwrap();
    text.push_str("\nA line the second package adds.\n");
    fs::write(&audit, text).unwrap();
    fs::write(w.pkg("references/added-note.md"), "# Added\nnote\n").unwrap();
    let plan_agent = w.agent_src("plan.md");
    let mut text = fs::read_to_string(&plan_agent).unwrap();
    text.push_str("\nA line the second package adds.\n");
    fs::write(&plan_agent, text).unwrap();

    let p2 = w.preview();
    let changed: Vec<(String, String, String)> = listed(&p2)
        .into_iter()
        .filter(|(_, _, a)| a != "unchanged")
        .map(|(l, path, a)| {
            (
                l,
                path.file_name().unwrap().to_string_lossy().into_owned(),
                a,
            )
        })
        .collect();
    eprintln!("CHANGED {changed:?}");
    let names: Vec<&str> = changed.iter().map(|(_, n, _)| n.as_str()).collect();
    assert!(names.contains(&"audit.md"), "{changed:?}");
    assert!(names.contains(&"added-note.md"), "{changed:?}");
    assert!(names.contains(&"moai-plan.md"), "{changed:?}");
    assert_eq!(
        changed.len(),
        3,
        "only the three changed files: {changed:?}"
    );
    w.confirm(&digest_of(&p2));
    assert!(
        fs::read_to_string(w.home.join(".claude/skills/moai/references/audit.md"))
            .unwrap()
            .contains("A line the second package adds.")
    );

    // The same package again: nothing to apply and the home is as it was.
    let settled = w.digest();
    let p3 = w.preview();
    assert_eq!(p3["nothing_to_apply"], json!(true), "{p3}");
    assert!(
        listed(&p3)
            .iter()
            .all(|(_, _, a)| a == "unchanged" || a == "set"),
        "{p3}"
    );
    assert_eq!(w.digest(), settled);
}

/// The files outside MoAI's own state folders with their bytes: what a user can see and edit.
fn user_files(home: &Path) -> Vec<(PathBuf, Vec<u8>)> {
    files_under(home)
        .into_iter()
        .filter(|rel| !is_state_folder(rel))
        .map(|rel| {
            let bytes = fs::read(home.join(&rel)).unwrap();
            (rel, bytes)
        })
        .collect()
}

fn write(path: &Path, bytes: impl AsRef<[u8]>) {
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, bytes).unwrap();
}

fn action_of(rows: &[(String, PathBuf, String)], file: &Path) -> String {
    rows.iter()
        .find(|(_, p, _)| p == file)
        .unwrap_or_else(|| panic!("no row for {}: {rows:?}", file.display()))
        .2
        .clone()
}

/// Appends `extra` to the package source `src`.
fn append(src: &Path, extra: &str) {
    let mut text = fs::read_to_string(src).unwrap();
    text.push_str(extra);
    fs::write(src, text).unwrap();
}

/// REQ-6, REQ-7: files of the user's own are never touched; a hand-edited managed file, a
/// managed file that is a link and one that is not UTF-8 are kept with the bytes they had.
#[test]
fn scenario_conflict_and_personal_files_are_kept() {
    let w = World::new();
    let claude = w.home.join(".claude");
    // The user's own files, there before the first install.
    let own_settings = "{\n  \"theme\": \"dark\",\n  \"permissions\": {\"allow\": [\"Bash(ls:*)\"]},\n  \"env\": {\"MY_TOOL\": \"1\"}\n}\n";
    let personal: Vec<(PathBuf, &str)> = vec![
        (claude.join("CLAUDE.md"), "my own notes, not MoAI's\n"),
        (claude.join("rules/mine.md"), "# my own rule\n"),
        (
            claude.join("agents/mine.md"),
            "---\nname: mine\ndescription: mine\n---\nmy agent\n",
        ),
        (
            claude.join("skills/mine/SKILL.md"),
            "---\nname: mine\ndescription: mine\n---\n",
        ),
    ];
    write(&claude.join("settings.json"), own_settings);
    for (p, text) in &personal {
        write(p, text);
    }
    let p = w.preview();
    w.confirm(&digest_of(&p));
    for (p, text) in &personal {
        assert_eq!(fs::read_to_string(p).unwrap(), *text, "{}", p.display());
    }
    // Another tool's keys survive the style setting.
    let settings: Value =
        serde_json::from_slice(&fs::read(claude.join("settings.json")).unwrap()).unwrap();
    let theirs: Value = serde_json::from_str(own_settings).unwrap();
    for (key, value) in theirs.as_object().unwrap() {
        assert_eq!(&settings[key], value, "the key {key} was changed");
    }

    // The user edits a skill file and the routing file; a third file is not UTF-8 and the
    // style file becomes a link. The next package changes all four sources.
    let audit = claude.join("skills/moai/references/audit.md");
    let routing = claude.join("rules/moai.md");
    let flow = claude.join("rules/moai/flow.md");
    let style = claude.join("output-styles/MoAI-Easy.md");
    let edited_audit = format!("{}\nmy own note\n", fs::read_to_string(&audit).unwrap());
    let edited_routing = format!(
        "{}\nmy own routing note\n",
        fs::read_to_string(&routing).unwrap()
    );
    write(&audit, &edited_audit);
    write(&routing, &edited_routing);
    write(&flow, [0xff, 0xfe, 0x00, 0x41]);
    let elsewhere = w.root.join("elsewhere/style-target.md");
    write(&elsewhere, "not MoAI's file\n");
    fs::remove_file(&style).unwrap();
    std::os::unix::fs::symlink(&elsewhere, &style).unwrap();
    append(
        &w.pkg("references/audit.md"),
        "\nthe package's new audit line\n",
    );
    append(
        &w.pkg("references/home/routing.md"),
        "\nthe package's new routing line\n",
    );
    append(
        &w.pkg("references/home/flow.md"),
        "\nthe package's new flow line\n",
    );
    append(
        &w.pkg("references/home/output-style.md"),
        "\nthe package's new style line\n",
    );

    let p2 = w.preview();
    let rows = listed(&p2);
    eprintln!(
        "ACTIONS audit={} routing={} flow={} style={}",
        action_of(&rows, &audit),
        action_of(&rows, &routing),
        action_of(&rows, &flow),
        action_of(&rows, &style)
    );
    for f in [&audit, &routing, &flow, &style] {
        let a = action_of(&rows, f);
        assert!(
            a != "create" && a != "update",
            "{} would be written: {a}",
            f.display()
        );
    }
    w.confirm(&digest_of(&p2));
    assert_eq!(fs::read_to_string(&audit).unwrap(), edited_audit);
    assert_eq!(fs::read_to_string(&routing).unwrap(), edited_routing);
    assert_eq!(fs::read(&flow).unwrap(), [0xff, 0xfe, 0x00, 0x41]);
    assert_eq!(
        fs::read_to_string(&elsewhere).unwrap(),
        "not MoAI's file\n",
        "written through a link"
    );
    assert!(fs::symlink_metadata(&style)
        .unwrap()
        .file_type()
        .is_symlink());
    for (p, text) in &personal {
        assert_eq!(fs::read_to_string(p).unwrap(), *text, "{}", p.display());
    }
}

const PERM: &str = "e2e-approved";

impl World {
    fn h(&self) -> String {
        self.home.display().to_string()
    }

    fn pkg_dir(&self) -> String {
        self.root.join("ver/skills/moai").display().to_string()
    }

    fn agents_dir(&self) -> String {
        self.root.join("ver/agents").display().to_string()
    }

    /// One layer's plan, apply and revert; the receipt of the apply, with the plan's digest.
    fn layer_apply(&self, plan_cmd: &str, apply_cmd: &str, with_agents: bool) -> String {
        let (h, pkg, agents) = (self.h(), self.pkg_dir(), self.agents_dir());
        let mut plan_args = vec![plan_cmd, h.as_str(), pkg.as_str()];
        if with_agents {
            plan_args.push(agents.as_str());
        }
        let plan = ok(&self.run(&plan_args));
        let digest = plan["plan_digest"].as_str().unwrap();
        let mut args = vec![apply_cmd, h.as_str(), pkg.as_str()];
        if with_agents {
            args.push(agents.as_str());
        }
        args.extend(["--plan-digest", digest, "--permission-reference", PERM]);
        let applied = ok(&self.run(&args));
        self.assert_inside(&applied);
        applied["receipt"].as_str().unwrap().to_string()
    }

    fn layer_revert(&self, revert_cmd: &str, receipt: &str) -> Value {
        let h = self.h();
        let r = ok(&self.run(&[
            revert_cmd,
            h.as_str(),
            "--receipt",
            receipt,
            "--permission-reference",
            PERM,
        ]));
        self.assert_inside(&r);
        r
    }
}

/// REQ-8, REQ-9: each layer's revert brings the user's files back to what they were before the
/// apply, and a file the user edited after the apply keeps the edit and is named.
#[test]
fn scenario_rollback_each_layer_returns_the_files_it_changed() {
    let w = World::new();
    let claude = w.home.join(".claude");
    write(
        &claude.join("settings.json"),
        "{\n  \"theme\": \"dark\"\n}\n",
    );
    write(&claude.join("CLAUDE.md"), "my own notes\n");
    for (plan, apply, revert, agents) in [
        (
            "home-routing-plan",
            "home-routing-apply",
            "home-routing-revert",
            false,
        ),
        (
            "agents-home-plan",
            "agents-home-apply",
            "agents-home-revert",
            true,
        ),
        (
            "home-style-plan",
            "home-style-apply",
            "home-style-revert",
            false,
        ),
    ] {
        let before = user_files(&w.home);
        let whole_before = w.digest();
        let files_before = files_under(&w.home);
        let receipt = w.layer_apply(plan, apply, agents);
        assert_ne!(user_files(&w.home), before, "{apply} wrote nothing");
        w.layer_revert(revert, &receipt);
        assert_eq!(
            user_files(&w.home),
            before,
            "{revert} did not return the files"
        );
        let left: Vec<PathBuf> = files_under(&w.home)
            .into_iter()
            .filter(|f| !files_before.contains(f))
            .collect();
        eprintln!(
            "ROLLBACK {revert}: whole tree equal = {}; files the revert leaves: {left:?}",
            w.digest() == whole_before
        );
        // Whatever stays is MoAI's own record of the apply, never a file the user can see.
        assert!(left.iter().all(|f| is_state_folder(f)), "{left:?}");
    }

    // An edit made after the apply survives the revert and is named in its answer.
    let receipt = w.layer_apply("home-routing-plan", "home-routing-apply", false);
    let routing = claude.join("rules/moai.md");
    let edited = format!(
        "{}\nmy edit after the apply\n",
        fs::read_to_string(&routing).unwrap()
    );
    write(&routing, &edited);
    let r = w.layer_revert("home-routing-revert", &receipt);
    assert_eq!(
        fs::read_to_string(&routing).unwrap(),
        edited,
        "the edit was lost"
    );
    assert!(
        r.to_string().contains("moai.md"),
        "the answer does not name the file: {r}"
    );
}

/// REQ-8: an update interrupted halfway is recovered to the skill files it had before.
#[test]
fn scenario_rollback_an_interrupted_update_is_recovered() {
    let w = World::new();
    let h = w.h();
    let stage1 = w.pkg_dir();
    let apply = |stage: &str, plan: &Value| {
        let out = w.run_with(
            &w.home,
            &[],
            Some(plan.to_string().as_bytes()),
            &[
                "update-apply",
                &h,
                stage,
                "--plan-digest",
                plan["plan_digest"].as_str().unwrap(),
                "--permission-reference",
                PERM,
            ],
        );
        envelope(&out)
    };
    let plan1 = ok(&w.run(&["update-plan", &h, &stage1]));
    let v1 = apply(&stage1, &plan1);
    assert_eq!(v1["result"]["status"], json!("applied"), "{v1}");

    let skill = w.home.join(".claude/skills/moai");
    let skill_before = tree_digest(&skill, &[]).unwrap();
    let stage2 = w.root.join("stage2");
    copy_tree(&w.root.join("ver/skills/moai"), &stage2);
    append(&stage2.join("SKILL.md"), "\nthe second package's line\n");
    append(
        &stage2.join("workflows/update.md"),
        "\nthe second package's line\n",
    );

    // A real interruption: the workflows folder cannot be written, after SKILL.md was.
    let workflows = skill.join("workflows");
    let original = fs::metadata(&workflows).unwrap().permissions();
    let mut locked = original.clone();
    locked.set_readonly(true);
    fs::set_permissions(&workflows, locked).unwrap();
    let stage2s = stage2.display().to_string();
    let plan2 = ok(&w.run(&["update-plan", &h, &stage2s]));
    let v2 = apply(&stage2s, &plan2);
    fs::set_permissions(&workflows, original).unwrap();
    assert_eq!(v2["result"]["status"], json!("partial"), "{v2}");
    assert_ne!(
        tree_digest(&skill, &[]).unwrap(),
        skill_before,
        "nothing was half written"
    );

    let rplan = ok(&w.run(&["recovery-plan", &h]));
    let out = w.run_with(
        &w.home,
        &[],
        Some(rplan.to_string().as_bytes()),
        &[
            "recovery-apply",
            &h,
            "--plan-digest",
            rplan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            PERM,
        ],
    );
    let r = ok(&out);
    w.assert_inside(&r);
    assert_eq!(
        tree_digest(&skill, &[]).unwrap(),
        skill_before,
        "the skill files are not what they were before the interrupted update"
    );
}

/// REQ-10, REQ-11: the hook entry answers every registered name with exit 0, an empty output or
/// one JSON document, and writes no file into the project, whatever the input.
#[test]
fn scenario_hook_entry_exits_zero_and_writes_nothing_into_the_project() {
    let w = World::new();
    let moai = w.root.join("work/moai-proj");
    let plain = w.root.join("work/plain");
    write(
        &moai
            .join(moai_cli::state_root::MOAI_DIR)
            .join("project.yaml"),
        "{\"schema_version\":2,\"project_id\":\"demo\"}\n",
    );
    write(&moai.join("src/main.rs"), "fn main() {}\n");
    write(&plain.join("src/main.rs"), "fn main() {}\n");
    let event =
        json!({"hook_event_name": "SessionStart", "session_id": "s-1", "source": "startup"})
            .to_string();
    let inputs: Vec<(&str, Vec<u8>)> = vec![
        ("an event", event.clone().into_bytes()),
        ("empty", Vec::new()),
        ("not json", b"not json".to_vec()),
        ("not an object", b"[1, 2]".to_vec()),
        ("not utf-8", vec![0xff, 0xfe, 0x00]),
    ];
    for name in ["session-context", "bash-boundary", "write-boundary"] {
        for (project, is_moai) in [(&moai, true), (&plain, false)] {
            for (what, bytes) in &inputs {
                let before = tree_digest(project, &[]).unwrap();
                let out = w.run_with(project, &[], Some(bytes), &["hook", name]);
                let at = format!("hook {name} in {} with {what}", project.display());
                assert_eq!(out.status.code(), Some(0), "{at}");
                if !out.stdout.is_empty() {
                    serde_json::from_slice::<Value>(&out.stdout)
                        .unwrap_or_else(|e| panic!("{at}: stdout is not one JSON document: {e}"));
                }
                if !is_moai {
                    assert!(out.stdout.is_empty(), "{at}: printed in a plain project");
                }
                if *what != "an event" {
                    assert!(out.stdout.is_empty(), "{at}: printed for bad input");
                }
                assert_eq!(
                    tree_digest(project, &[]).unwrap(),
                    before,
                    "{at}: wrote a file"
                );
            }
        }
        // The project named by the variable, the working folder elsewhere.
        let before = tree_digest(&moai, &[]).unwrap();
        let out = w.run_with(
            &w.home,
            &[("CLAUDE_PROJECT_DIR", moai.to_str().unwrap())],
            Some(event.as_bytes()),
            &["hook", name],
        );
        assert_eq!(
            out.status.code(),
            Some(0),
            "hook {name} with CLAUDE_PROJECT_DIR"
        );
        assert_eq!(tree_digest(&moai, &[]).unwrap(), before);
    }
}

const S_LANE: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_AUDIT: &str = "cccccccc-0000-4000-8000-00000000000c";
const S_LANDER: &str = "dddddddd-0000-4000-8000-00000000000d";
/// Test data: the profile of a temporary project whose integration branch is `integ`.
const PROFILE: &str = "{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"integ\", \"land_verdict\": \"enforce\"}}";

fn git(dir: &Path, args: &[&str]) -> String {
    let mut all = vec![
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
    ];
    all.extend_from_slice(args);
    common::git::git(dir, &all)
}

/// REQ-12: one card goes from take to landing in a temporary repository and HOME, and the next
/// card is then the one `card next` names. Every name, path and commit here is test data.
#[test]
fn scenario_card_cycle_takes_checks_audits_lands_and_names_the_next_card() {
    let w = World::new();
    let project = w.root.join("work/demo");
    fs::create_dir_all(&project).unwrap();
    git(&project, &["init", "-q", "-b", "main"]);
    write(&project.join("a.txt"), "a\n");
    git(&project, &["add", "a.txt"]);
    git(&project, &["commit", "-q", "-m", "init"]);
    git(&project, &["branch", "integ"]);
    let tip = git(&project, &["rev-parse", "refs/heads/integ"]);
    write(
        &project
            .join(moai_cli::PROJECT_STATE_DIR)
            .join("project.yaml"),
        PROFILE,
    );
    moai_cli::ledger::register(&w.home, "demo").unwrap();
    let p = project.to_str().unwrap().to_string();
    let card = |session: Option<&str>, args: &[&str], stdin: Option<&[u8]>| -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &p]);
        let mut envs = vec![("MOAI_GIT_BIN", "/usr/bin/git")];
        if let Some(s) = session {
            envs.push(("CLAUDE_CODE_SESSION_ID", s));
        }
        w.run_with(&project, &envs, stdin, &all)
    };
    let add = |title: &str| -> String {
        let r = ok(&card(
            None,
            &[
                "add", "--title", title, "--kind", "task", "--scope", "apps/a/", "--ac", "one",
            ],
            None,
        ));
        let name = r["name"].as_str().unwrap().to_string();
        ok(&card(None, &["ready", &name], None));
        name
    };
    let first = add("first fixture card");
    let second = add("second fixture card");

    // Take in a worktree of its own, with the write check before and after.
    let wt = w.root.join("work/wt-first");
    git(
        &project,
        &[
            "worktree",
            "add",
            "-q",
            wt.to_str().unwrap(),
            "-b",
            "lane/first",
            "integ",
        ],
    );
    let taken = ok(&card(
        Some(S_LANE),
        &[
            "take",
            &first,
            "--worktree",
            wt.to_str().unwrap(),
            "--branch",
            "lane/first",
        ],
        None,
    ));
    assert_eq!(taken["reservation"]["generation"], json!(1), "{taken}");
    let inside = ok(&card(
        Some(S_LANE),
        &[
            "check",
            &first,
            "--generation",
            "1",
            "--path",
            "apps/a/x.txt",
        ],
        None,
    ));
    assert_eq!(inside["allowed"], json!(true), "{inside}");
    let outside = card(
        Some(S_LANE),
        &[
            "check",
            &first,
            "--generation",
            "1",
            "--path",
            "docs/other.md",
        ],
        None,
    );
    let outside = envelope(&outside);
    assert_ne!(outside["result"]["allowed"], json!(true), "{outside}");

    // The change, committed in the worktree.
    write(&wt.join("apps/a/x.txt"), "the card's change\n");
    git(&wt, &["add", "apps/a/x.txt"]);
    // The trailers come from the tool, never typed: the landing reads who wrote the head.
    let trailers = ok(&w.run_with(
        &wt,
        &[("CLAUDE_CODE_SESSION_ID", S_LANE)],
        None,
        &["trailers", "--spec", "SPEC-20261123-fixture-card"],
    ));
    let trailer_text = trailers["text"].as_str().unwrap().to_string();
    assert!(trailer_text.contains("Moai-Session: "), "{trailers}");
    git(
        &wt,
        &[
            "commit",
            "-q",
            "-m",
            "the card's change",
            "-m",
            trailer_text.trim_end(),
        ],
    );
    let head = git(&wt, &["rev-parse", "HEAD"]);

    // The ready note, the review state, an independent verdict, then the landing.
    let note = format!("LANE-READY-INTEGRATION {first} head={head}\nbase {tip}\n");
    ok(&card(
        Some(S_LANE),
        &["note", &first, "--kind", "report"],
        Some(note.as_bytes()),
    ));
    ok(&card(
        Some(S_LANE),
        &["review", &first, "--generation", "1"],
        None,
    ));
    let verdict = ok(&card(
        Some(S_AUDIT),
        &[
            "verdict",
            &first,
            "--head",
            &head,
            "--base",
            &tip,
            "--verdict",
            "pass",
            "--tier",
            "1",
            "--evidence",
            "read the diff in the temporary repository",
            "--receipt-head",
            &head,
            "--receipt-step",
            "diff-check=0",
        ],
        None,
    ));
    assert_eq!(verdict["written"], json!(true), "{verdict}");
    // The lander is a third session: a tier 1 PASS counts only from a session other than it.
    let landed = ok(&card(
        Some(S_LANDER),
        &[
            "land",
            &first,
            "--head",
            &head,
            "--met",
            "1=the temporary test",
            "--by",
            "leader",
        ],
        None,
    ));
    assert_eq!(landed["moved"], json!(true), "{landed}");
    assert_eq!(landed["closed"], json!(true), "{landed}");
    assert_eq!(landed["state"], json!("done"), "{landed}");
    let merge_sha = git(&project, &["rev-parse", "refs/heads/integ"]);
    assert_eq!(
        merge_sha, head,
        "the integration branch is not at the landed head"
    );
    eprintln!("CARD-CYCLE landed head {merge_sha} base {tip}");

    // The next card is the second one.
    let next = ok(&w.run_with(
        &project,
        &[("CLAUDE_CODE_SESSION_ID", S_LANE)],
        None,
        &["card", "next", &p],
    ));
    assert_eq!(next["next"], json!(second), "{next}");
    assert_eq!(next["next_is"], json!("ready"), "{next}");
}
