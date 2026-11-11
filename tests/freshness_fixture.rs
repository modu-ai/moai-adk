//! #2407: the C6 incremental index judged by the FRESHNESS cases of
//! `tests/fixtures/atlas/suites.json` (the fixture is the case authority, #2346 AC-3).
//! Each case gets its own temporary repository: setup → `index-build` → mutation →
//! `index-query`, and the query is compared with the case's `expected`. A second test
//! perturbs each runnable case's expectation and requires the judge to reject it, so a
//! judge that accepts anything cannot pass.
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

const SUITES: &str = "tests/fixtures/atlas/suites.json";

fn cases() -> Vec<Value> {
    let path = Path::new(env!("CARGO_MANIFEST_DIR")).join(SUITES);
    let suites: Value = serde_json::from_slice(&fs::read(path).unwrap()).unwrap();
    let suite = suites["suites"]
        .as_array()
        .unwrap()
        .iter()
        .find(|s| s["id"] == "FRESHNESS")
        .expect("FRESHNESS suite")
        .clone();
    suite["cases"].as_array().unwrap().clone()
}

fn with_env(mut command: Command, dir: &Path) -> Command {
    command
        .current_dir(dir)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("GIT_AUTHOR_NAME", "t")
        .env("GIT_AUTHOR_EMAIL", "t@example.com")
        .env("GIT_COMMITTER_NAME", "t")
        .env("GIT_COMMITTER_EMAIL", "t@example.com");
    command
}

fn git(dir: &Path, args: &[&str]) {
    let mut command = with_env(Command::new("git"), dir);
    command.args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"]);
    let status = command.args(args).status().unwrap();
    assert!(status.success(), "git {args:?} failed in {}", dir.display());
}

fn cli(dir: &Path, op: &str) -> Value {
    let mut child = with_env(Command::new(env!("CARGO_BIN_EXE_moai-cli")), dir)
        .args([op, "."])
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(br#"{"protocol_version": 1}"#)
        .unwrap();
    let out = child.wait_with_output().unwrap();
    assert!(
        out.status.success(),
        "{op} failed: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    serde_json::from_slice::<Value>(&out.stdout).unwrap()["result"].clone()
}

fn pointer_gen(dir: &Path) -> Option<String> {
    let bytes = fs::read(dir.join(".moai-cli/index/current.json")).ok()?;
    let v: Value = serde_json::from_slice(&bytes).ok()?;
    v["generation_id"].as_str().map(str::to_string)
}

/// A case the C6 index cannot judge yet, with the reason.
fn not_run_reason(case: &Value) -> Option<&'static str> {
    (case["mutation"]["operation"] == "fail_scan").then_some(
        "the C6 index records files only; edges arrive with the map/graph records of #2347 \
         (C6b), so `edge_state` has no source to judge",
    )
}

struct Repo {
    _tmp: tempfile::TempDir,
    main: PathBuf,
    worktrees: Vec<(String, PathBuf)>,
}

impl Repo {
    fn dir(&self, worktree: Option<&str>) -> &Path {
        match worktree {
            None => &self.main,
            Some(name) => {
                &self
                    .worktrees
                    .iter()
                    .find(|(n, _)| n == name)
                    .unwrap_or_else(|| panic!("unknown worktree {name}"))
                    .1
            }
        }
    }
}

fn setup(case: &Value) -> Repo {
    let tmp = tempfile::tempdir().unwrap();
    let main = tmp.path().join("project");
    fs::create_dir_all(&main).unwrap();
    git(&main, &["init", "-q", "--initial-branch=main"]);
    for (path, contents) in case["setup"]["tracked_files"].as_object().unwrap() {
        fs::write(main.join(path), contents.as_str().unwrap()).unwrap();
    }
    git(&main, &["add", "."]);
    git(&main, &["commit", "-q", "-m", "init"]);
    let mut worktrees = Vec::new();
    for name in case["setup"]["worktrees"].as_array().into_iter().flatten() {
        let name = name.as_str().unwrap().to_string();
        let path = tmp.path().join(&name);
        git(
            &main,
            &["worktree", "add", "-q", "-b", &name, path.to_str().unwrap()],
        );
        worktrees.push((name, path));
    }
    Repo {
        _tmp: tmp,
        main,
        worktrees,
    }
}

fn mutate(repo: &Repo, m: &Value) {
    let dir = repo.dir(m["worktree"].as_str());
    let path = m["path"].as_str().unwrap_or_default();
    let contents = m["contents"].as_str().unwrap_or_default();
    match m["operation"].as_str().unwrap() {
        "add_tracked" => {
            fs::write(dir.join(path), contents).unwrap();
            git(dir, &["add", path]);
        }
        "add_untracked" | "modify" | "modify_worktree" => {
            fs::write(dir.join(path), contents).unwrap();
        }
        "delete_without_git_rm" => fs::remove_file(dir.join(path)).unwrap(),
        "git_mv" => git(
            dir,
            &["mv", m["from"].as_str().unwrap(), m["to"].as_str().unwrap()],
        ),
        "checkout_new_commit" => {
            assert_eq!(m["same_files"], json!(true), "only same_files is defined");
            git(
                dir,
                &["checkout", "-q", "-b", m["branch"].as_str().unwrap()],
            );
            git(dir, &["commit", "-q", "--allow-empty", "-m", "other"]);
        }
        other => panic!("unknown mutation operation {other}"),
    }
}

/// Compares one directory's query with an expectation (`state`, `changes`,
/// `generation_unchanged`).
fn compare(label: &str, dir: &Path, before: &Option<String>, want: &Value) -> Result<(), String> {
    let got = cli(dir, "index-query");
    if got["state"] != want["state"] {
        return Err(format!(
            "{label}: state {} != {}",
            got["state"], want["state"]
        ));
    }
    for (key, value) in want["changes"].as_object().into_iter().flatten() {
        let seen = &got["changes"][key];
        if seen != value {
            return Err(format!("{label}: changes.{key} {seen} != {value}"));
        }
    }
    if want["generation_unchanged"] == json!(true) && pointer_gen(dir) != *before {
        return Err(format!("{label}: the generation pointer moved"));
    }
    Ok(())
}

/// Runs one case end to end; `Err` names the first difference.
fn judge(case: &Value) -> Result<(), String> {
    let id = case["case_id"].as_str().unwrap();
    let repo = setup(case);
    let mut before: Vec<(Option<String>, Option<String>)> = Vec::new();
    let mut dirs: Vec<Option<String>> = vec![None];
    dirs.extend(repo.worktrees.iter().map(|(n, _)| Some(n.clone())));
    if case["setup"]["build_index"] == json!(true) {
        for d in &dirs {
            let dir = repo.dir(d.as_deref());
            cli(dir, "index-build");
            before.push((d.clone(), pointer_gen(dir)));
        }
    }
    let gen_of = |name: Option<&str>| {
        before
            .iter()
            .find(|(d, _)| d.as_deref() == name)
            .and_then(|(_, g)| g.clone())
    };
    let m = &case["mutation"];
    mutate(&repo, m);
    let target = m["worktree"].as_str();
    compare(id, repo.dir(target), &gen_of(target), &case["expected"])?;
    if let Some(other) = case["expected"].get("other_worktree") {
        let name = other["worktree"].as_str().unwrap();
        compare(
            &format!("{id} ({name})"),
            repo.dir(Some(name)),
            &gen_of(Some(name)),
            other,
        )?;
    }
    if case["expected"]["cache_scope"] == json!("per_worktree") {
        for (name, path) in &repo.worktrees {
            if !path.join(".moai-cli/index/current.json").is_file() {
                return Err(format!(
                    "{id}: worktree {name} has no index pointer of its own"
                ));
            }
        }
    }
    Ok(())
}

#[test]
fn freshness_cases_from_the_fixture() {
    let all = cases();
    assert_eq!(all.len(), 8, "the FRESHNESS suite has eight cases");
    let mut not_run = Vec::new();
    for case in &all {
        let id = case["case_id"].as_str().unwrap();
        if let Some(reason) = not_run_reason(case) {
            eprintln!("not_run {id}: {reason}");
            not_run.push(id.to_string());
            continue;
        }
        if let Err(e) = judge(case) {
            panic!("{e}");
        }
    }
    assert_eq!(not_run, ["F-08"], "only the edge case waits for #2347");
}

/// Mutation per case: a wrong expected state must make the judge fail.
#[test]
fn the_judge_rejects_a_wrong_expectation_in_every_runnable_case() {
    for case in cases() {
        if not_run_reason(&case).is_some() {
            continue;
        }
        let id = case["case_id"].as_str().unwrap().to_string();
        let mut wrong = case.clone();
        let state = wrong["expected"]["state"].as_str().unwrap();
        let flipped = if state == "fresh" { "stale" } else { "fresh" };
        wrong["expected"]["state"] = json!(flipped);
        assert!(
            judge(&wrong).is_err(),
            "{id}: the judge accepted state {flipped}"
        );
    }
}
