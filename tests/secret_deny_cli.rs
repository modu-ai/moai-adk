//! `secret-deny-plan` (SPEC-20261122-secret-file-deny-rules, slice A) and `secret-deny-apply`
//! (slice B): temporary projects and HOME folders only. Both verbs run the built binary; the walk
//! bound is driven at library level.

#![cfg(unix)]

mod common;

use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs;
use std::os::unix::fs::{symlink, PermissionsExt};
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use tempfile::TempDir;

/// A value planted in a key file; deliberately not shaped like any credential.
const PLANTED: &str = "planted-value-not-a-secret-7f3a";

/// This repository's `.claudeignore` as read on 2026-11-26, comments and order kept.
const REPOSITORY_LIST: &str = include_str!("fixtures/secret_deny/claudeignore");

/// Its entries in file order with the rule each one converts to (the table under the SPEC's
/// Plan) and `yes` for a recommended entry: `entry<TAB>rule<TAB>yes|no`.
const EXPECTED_RULES: &str = include_str!("fixtures/secret_deny/rules.tsv");

fn expected_rules() -> Vec<(&'static str, &'static str, bool)> {
    EXPECTED_RULES
        .lines()
        .map(|line| {
            let mut cells = line.split('\t');
            let mut cell = || cells.next().expect("three cells in rules.tsv");
            (cell(), cell(), cell() == "yes")
        })
        .collect()
}

fn project() -> (TempDir, PathBuf) {
    let dir = TempDir::new().expect("temp dir");
    let root = dir.path().join("project");
    fs::create_dir_all(&root).expect("project folder");
    common::git::git_ok(&root, &["init"]);
    (dir, root)
}

fn put(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().expect("parent")).expect("parent folder");
    fs::write(path, text).expect("write fixture file");
}

/// Runs the preview of the built binary; answers (stdout, stderr, envelope).
fn preview(root: &Path) -> (String, String, Value) {
    let home = TempDir::new().expect("verb home");
    let out = Command::new(common::MOAI_BIN)
        .arg("secret-deny-plan")
        .arg(root)
        .current_dir(root)
        .env("PATH", "/usr/sbin:/usr/bin:/bin:/sbin")
        .env("HOME", home.path())
        .env("LC_ALL", "C")
        .stdin(Stdio::null())
        .output()
        .expect("run secret-deny-plan");
    let stdout = String::from_utf8_lossy(&out.stdout).into_owned();
    let stderr = String::from_utf8_lossy(&out.stderr).into_owned();
    assert_eq!(
        out.status.code(),
        Some(0),
        "preview exits 0: {stdout}{stderr}"
    );
    let envelope: Value = serde_json::from_slice(&out.stdout).expect("envelope");
    assert_eq!(envelope["operation"], "secret_deny.plan");
    assert_eq!(
        fs::read_dir(home.path()).expect("home").count(),
        0,
        "the preview writes nothing under HOME"
    );
    (stdout, stderr, envelope)
}

fn strings(value: &Value) -> Vec<String> {
    value
        .as_array()
        .expect("array")
        .iter()
        .map(|v| v.as_str().expect("string").to_string())
        .collect()
}

fn rule_names(list: &Value) -> Vec<String> {
    list.as_array()
        .expect("rule array")
        .iter()
        .map(|r| r["rule"].as_str().expect("rule").to_string())
        .collect()
}

fn find<'a>(list: &'a Value, rule: &str) -> &'a Value {
    list.as_array()
        .expect("rule array")
        .iter()
        .find(|r| r["rule"] == rule)
        .unwrap_or_else(|| panic!("no rule {rule} in {list}"))
}

/// Every path under `root` (`.git` aside) with what it holds: bytes, a link target or a folder.
fn snapshot(root: &Path) -> BTreeMap<PathBuf, Vec<u8>> {
    fn visit(dir: &Path, into: &mut BTreeMap<PathBuf, Vec<u8>>) {
        for entry in fs::read_dir(dir).expect("read dir") {
            let path = entry.expect("entry").path();
            if path.file_name().is_some_and(|n| n == ".git") {
                continue;
            }
            let meta = fs::symlink_metadata(&path).expect("metadata");
            if meta.is_dir() {
                into.insert(path.clone(), b"folder".to_vec());
                visit(&path, into);
            } else {
                into.insert(path.clone(), fs::read(&path).expect("read"));
            }
        }
    }
    let mut into = BTreeMap::new();
    visit(root, &mut into);
    into
}

/// [`snapshot`] for a tree that may hold links: a symbolic link is recorded as its target and
/// never followed, so a linked folder neither fails the read nor hides a write through it.
fn tree(root: &Path) -> BTreeMap<PathBuf, Vec<u8>> {
    fn visit(dir: &Path, into: &mut BTreeMap<PathBuf, Vec<u8>>) {
        for entry in fs::read_dir(dir).expect("read dir") {
            let path = entry.expect("entry").path();
            if path.file_name().is_some_and(|n| n == ".git") {
                continue;
            }
            let meta = fs::symlink_metadata(&path).expect("metadata");
            if meta.file_type().is_symlink() {
                let target = fs::read_link(&path).expect("link target");
                into.insert(path, format!("link:{}", target.display()).into_bytes());
            } else if meta.is_dir() {
                into.insert(path.clone(), b"folder".to_vec());
                visit(&path, into);
            } else {
                into.insert(path.clone(), fs::read(&path).expect("read"));
            }
        }
    }
    let mut into = BTreeMap::new();
    visit(root, &mut into);
    into
}

/// What the built binary answered; judging the exit code is the caller's job.
struct Run {
    code: Option<i32>,
    /// Standard output and standard error together, for assertion messages.
    text: String,
    envelope: Value,
}

impl Run {
    fn operation(&self) -> &str {
        self.envelope["operation"].as_str().unwrap_or("")
    }

    /// The `result` of an answer, or the `error` of a refusal.
    fn body(&self) -> &Value {
        match self.envelope.get("result") {
            Some(result) => result,
            None => &self.envelope["error"],
        }
    }

    fn message(&self) -> String {
        self.body()["message"].as_str().unwrap_or("").to_lowercase()
    }
}

/// Runs one `secret-deny-*` verb of the built binary with a temporary HOME that must stay empty.
fn run_verb(verb: &str, root: &Path, args: &[String]) -> Run {
    let home = TempDir::new().expect("verb home");
    let out = Command::new(common::MOAI_BIN)
        .arg(verb)
        .arg(root)
        .args(args)
        .current_dir(root)
        .env("PATH", "/usr/sbin:/usr/bin:/bin:/sbin")
        .env("HOME", home.path())
        .env("LC_ALL", "C")
        .env_remove("CLAUDE_CONFIG_DIR")
        .stdin(Stdio::null())
        .output()
        .unwrap_or_else(|e| panic!("run {verb}: {e}"));
    assert_eq!(
        fs::read_dir(home.path()).expect("home").count(),
        0,
        "{verb} writes nothing under HOME"
    );
    Run {
        code: out.status.code(),
        text: format!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        ),
        envelope: serde_json::from_slice(&out.stdout).unwrap_or(Value::Null),
    }
}

fn pick_args(picks: &[&str], retire: bool) -> Vec<String> {
    let mut args = Vec::new();
    for pick in picks {
        args.push("--pick".to_string());
        args.push((*pick).to_string());
    }
    if retire {
        args.push("--retire-claudeignore".to_string());
    }
    args
}

fn plan_run(root: &Path, picks: &[&str], retire: bool) -> Run {
    run_verb("secret-deny-plan", root, &pick_args(picks, retire))
}

/// The `result` of a preview that must succeed.
fn plan_with(root: &Path, picks: &[&str], retire: bool) -> Value {
    let run = plan_run(root, picks, retire);
    assert_eq!(run.code, Some(0), "the preview exits 0: {}", run.text);
    assert_eq!(run.operation(), "secret_deny.plan", "{}", run.text);
    run.envelope["result"].clone()
}

fn digest_of(preview: &Value) -> String {
    preview["apply"]["plan_digest"]
        .as_str()
        .unwrap_or_else(|| panic!("no apply.plan_digest in {preview}"))
        .to_string()
}

fn apply_run(root: &Path, digest: &str, picks: &[&str], retire: bool) -> Run {
    let mut args = vec![
        "--plan-digest".to_string(),
        digest.to_string(),
        "--permission-reference".to_string(),
        "test-secret-deny".to_string(),
    ];
    args.extend(pick_args(picks, retire));
    run_verb("secret-deny-apply", root, &args)
}

/// A preview and its apply, both of which must succeed and report `applied`.
fn apply_ok(root: &Path, picks: &[&str], retire: bool) -> Run {
    let preview = plan_with(root, picks, retire);
    let run = apply_run(root, &digest_of(&preview), picks, retire);
    assert_eq!(run.code, Some(0), "the apply exits 0: {}", run.text);
    assert!(
        run.operation().starts_with("secret_deny."),
        "this verb's own answer: {}",
        run.text
    );
    assert_eq!(run.body()["applied"].as_bool(), Some(true), "{}", run.text);
    run
}

/// A refusal that is the verb's own answer (never the generic unknown-command one), claims no
/// success and gives a message naming one of `reasons`.
fn assert_refused(run: &Run, reasons: &[&str]) {
    assert!(
        run.operation().starts_with("secret_deny."),
        "a refusal from the verb itself: {}",
        run.text
    );
    let body = run.body();
    assert!(
        run.code != Some(0)
            || body["written"].as_bool() == Some(false)
            || body["applied"].as_bool() == Some(false),
        "the apply says no: {}",
        run.text
    );
    assert_ne!(body["applied"].as_bool(), Some(true), "{}", run.text);
    let message = run.message();
    assert!(
        reasons.iter().any(|reason| message.contains(reason)),
        "the message names one of {reasons:?}: {}",
        run.text
    );
}

fn write_settings(root: &Path, value: &Value) {
    let text = serde_json::to_string_pretty(value).expect("settings JSON");
    put(root, ".claude/settings.json", &format!("{text}\n"));
}

fn read_settings(root: &Path) -> Value {
    let bytes = fs::read(root.join(".claude/settings.json")).expect("settings file");
    serde_json::from_slice(&bytes).expect("settings file is JSON")
}

fn deny_list(root: &Path) -> Vec<String> {
    strings(&read_settings(root)["permissions"]["deny"])
}

/// A `!` entry of a deny list: `Read(!<path>)` (D-5).
fn is_negation(entry: &str) -> bool {
    entry.starts_with("Read(!")
}

/// A key-file rule (REQ-6): a `Read(...)` rule that is no negation and whose argument, a leading
/// `./`, `/` or `**/` stripped, is one of the credential file shapes the planner proposes.
fn is_key_rule(entry: &str) -> bool {
    let Some(argument) = entry
        .strip_prefix("Read(")
        .and_then(|rest| rest.strip_suffix(')'))
    else {
        return false;
    };
    if argument.starts_with('!') {
        return false;
    }
    let argument = argument
        .strip_prefix("./")
        .or_else(|| argument.strip_prefix("**/"))
        .or_else(|| argument.strip_prefix('/'))
        .unwrap_or(argument);
    moai_cli::secret_deny::KEY_SHAPES.contains(&argument)
}

/// Every key-file rule of `deny` stands after its last `!` entry.
fn assert_keys_after_negations(deny: &[String]) {
    let last = deny
        .iter()
        .rposition(|entry| is_negation(entry))
        .unwrap_or_else(|| panic!("no negation in {deny:?}"));
    let keys: Vec<usize> = (0..deny.len()).filter(|i| is_key_rule(&deny[*i])).collect();
    assert!(!keys.is_empty(), "no key rule in {deny:?}");
    for i in keys {
        assert!(
            i > last,
            "{} is not after the last negation: {deny:?}",
            deny[i]
        );
    }
}

/// The entries of `deny` that are key-file rules, in list order.
fn key_rules(deny: &[String]) -> Vec<&str> {
    deny.iter()
        .map(String::as_str)
        .filter(|e| is_key_rule(e))
        .collect()
}

/// The entries of `deny` that are no key-file rule, in list order.
fn other_rules(deny: &[String]) -> Vec<&str> {
    deny.iter()
        .map(String::as_str)
        .filter(|e| !is_key_rule(e))
        .collect()
}

/// The carve-out the preview names for the negation `negation`.
fn carve<'a>(carve_outs: &'a Value, negation: &str) -> &'a Value {
    carve_outs
        .as_array()
        .expect("carve_outs array")
        .iter()
        .find(|c| c["negation"] == negation)
        .unwrap_or_else(|| panic!("no carve-out for {negation} in {carve_outs}"))
}

/// Every regular file under the settings backup folder with its bytes; the `.gitignore` the
/// folder keeps for itself is left out.
fn backed_up(root: &Path) -> Vec<(PathBuf, Vec<u8>)> {
    fn visit(dir: &Path, own: &Path, into: &mut Vec<(PathBuf, Vec<u8>)>) {
        let Ok(reader) = fs::read_dir(dir) else {
            return;
        };
        for entry in reader {
            let path = entry.expect("entry").path();
            let meta = fs::symlink_metadata(&path).expect("metadata");
            if meta.is_dir() {
                visit(&path, own, into);
            } else if meta.is_file() && path != own.join(".gitignore") {
                into.push((path.clone(), fs::read(&path).expect("read")));
            }
        }
    }
    let backups = moai_cli::state_root::path(root, "settings-backups");
    let mut into = Vec::new();
    visit(&backups, &backups, &mut into);
    into
}

/// Runs one `claude-settings-plan` / `-apply` with `home` as HOME; answers (exit code, envelope).
fn claude_settings(args: &[&str], home: &Path, desired: &Value) -> (Option<i32>, Value) {
    let mut child = Command::new(common::MOAI_BIN)
        .args(args)
        .env("HOME", home)
        .env_remove("CLAUDE_CONFIG_DIR")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .expect("spawn claude-settings");
    common::feed_stdin(&mut child, desired.to_string().as_bytes());
    let out = child.wait_with_output().expect("claude-settings output");
    (
        out.status.code(),
        serde_json::from_slice(&out.stdout).unwrap_or(Value::Null),
    )
}

#[test]
fn preview_lists_rules_and_writes_nothing() {
    let (_dir, root) = project();
    for rel in [
        ".env",
        "apps/web/.env",
        "apps/web/.env.local",
        "certs/server.pem",
        "certs/server.key",
        "keys/ci-service-account.json",
        ".npmrc",
        "README.md",
        "src/env.rs",
        "service-account.json",
    ] {
        put(&root, rel, "x\n");
    }
    put(&root, ".claude/settings.json", "{\"model\": \"opus\"}\n");
    let before = snapshot(&root);

    let (_, _, envelope) = preview(&root);
    let result = &envelope["result"];
    let keys = &result["key_rules"];
    // `.pypirc` matched nothing, so it proposes no rule.
    assert_eq!(
        rule_names(keys),
        [
            "Read(.env)",
            "Read(.env.*)",
            "Read(*.pem)",
            "Read(*.key)",
            "Read(*-service-account.json)",
            "Read(.npmrc)",
        ]
    );
    assert_eq!(
        strings(&find(keys, "Read(.env)")["matched"]),
        [".env", "apps/web/.env"]
    );
    assert_eq!(
        strings(&find(keys, "Read(.env.*)")["matched"]),
        ["apps/web/.env.local"]
    );
    assert_eq!(
        strings(&find(keys, "Read(*-service-account.json)")["matched"]),
        ["keys/ci-service-account.json"]
    );
    let settings = result["settings_file"].as_str().expect("settings file");
    assert!(
        settings.ends_with("project/.claude/settings.json"),
        "{settings}"
    );
    assert_eq!(result["claudeignore"]["state"], "absent");
    assert_eq!(result["walk"]["complete"], true);
    assert_eq!(snapshot(&root), before, "the preview changes no file");
}

#[test]
fn claudeignore_entries_are_picked_one_by_one() {
    let (_dir, root) = project();
    let table = expected_rules();
    assert_eq!(table.len(), 34, "the SPEC table holds 34 entries");
    // Forms the repository's list does not hold: a negation and a root-anchored folder.
    let list = format!("{REPOSITORY_LIST}!.env.example\n/generated/\n!certs/dev.pem\n");
    put(&root, ".claudeignore", &list);
    put(&root, ".env.production", "x\n");
    put(&root, ".env.example", "x\n");
    put(&root, "web/node_modules/pkg/index.js", "x\n");
    put(&root, "build.log", "x\n");
    put(&root, "generated/a.txt", "x\n");
    put(&root, "sub/generated/b.txt", "x\n");
    put(&root, "docs-site/public/page.html", "x\n");
    put(&root, "logs", "a file, not a folder\n");
    common::git::git_ok(&root, &["add", "-f", ".env.production", "build.log"]);

    let (_, _, envelope) = preview(&root);
    let ignore = &envelope["result"]["claudeignore"];
    assert_eq!(ignore["state"], "present");
    let entries = ignore["entries"].as_array().expect("entries");
    assert_eq!(entries.len(), table.len() + 3, "no blank or comment line");
    for (entry, expected) in entries.iter().zip(table) {
        assert_eq!(entry["entry"], expected.0);
        assert_eq!(entry["rule"], expected.1, "{}", expected.0);
        assert_eq!(entry["recommended"], expected.2, "{}", expected.0);
        let warning = entry["warning"].as_str().expect("warning");
        assert!(warning.contains("blocks editing"), "{warning}");
    }
    // Line numbers are those of the file, comments and blank lines counted.
    assert_eq!(entries[0]["line"], 5);

    let list = &ignore["entries"];
    let negation = find(list, "Read(!.env.example)");
    assert_eq!(negation["recommended"], false);
    assert_eq!(strings(&negation["matched"]), [".env.example"]);
    assert!(
        negation["warning"].as_str().unwrap().contains("carves"),
        "a negation says what it does instead: {negation}"
    );
    // A negation naming a key file is still never recommended: it would open the file.
    assert_eq!(find(list, "Read(!certs/dev.pem)")["recommended"], false);
    let anchored = find(list, "Read(/generated/**/*)");
    assert_eq!(strings(&anchored["matched"]), ["generated/a.txt"]);

    let production = find(list, "Read(.env.production)");
    assert_eq!(strings(&production["matched"]), [".env.production"]);
    assert_eq!(strings(&production["tracked"]), [".env.production"]);
    let modules = find(list, "Read(**/node_modules/**/*)");
    assert_eq!(
        strings(&modules["matched"]),
        ["web/node_modules/pkg/index.js"]
    );
    assert!(strings(&modules["tracked"]).is_empty());
    assert_eq!(
        strings(&find(list, "Read(*.log)")["tracked"]),
        ["build.log"]
    );
    assert_eq!(
        strings(&find(list, "Read(docs-site/public/**/*)")["matched"]),
        ["docs-site/public/page.html"]
    );
    // A folder entry matches no file of that name, and an entry with no file is still listed.
    assert!(strings(&find(list, "Read(**/logs/**/*)")["matched"]).is_empty());
    assert!(strings(&find(list, "Read(*.pyc)")["matched"]).is_empty());
}

#[test]
fn matched_file_is_never_opened() {
    let (_dir, root) = project();
    put(&root, ".env", &format!("TOKEN={PLANTED}\n"));
    put(&root, "locked/server.key", PLANTED);
    let locked = root.join("locked/server.key");
    fs::set_permissions(&locked, fs::Permissions::from_mode(0o000)).expect("chmod");
    put(&root, ".claudeignore", ".env\n*.key\n");

    let (stdout, stderr, envelope) = preview(&root);
    let keys = &envelope["result"]["key_rules"];
    assert_eq!(strings(&find(keys, "Read(.env)")["matched"]), [".env"]);
    assert_eq!(
        strings(&find(keys, "Read(*.key)")["matched"]),
        ["locked/server.key"]
    );
    let all = format!("{stdout}{stderr}");
    assert!(
        !all.contains(PLANTED),
        "a matched file's content is printed"
    );
    assert!(
        !all.to_ascii_lowercase().contains("permission denied"),
        "a matched file was opened: {all}"
    );
    fs::set_permissions(&locked, fs::Permissions::from_mode(0o600)).expect("chmod back");
}

#[test]
fn a_linked_ignore_list_is_not_read() {
    let (_dir, root) = project();
    put(&root, ".env", &format!("{PLANTED}\n"));
    symlink(root.join(".env"), root.join(".claudeignore")).expect("list link");

    let (stdout, stderr, envelope) = preview(&root);
    let ignore = &envelope["result"]["claudeignore"];
    assert_eq!(ignore["state"], "not_read", "{ignore}");
    assert!(ignore["entries"].as_array().expect("entries").is_empty());
    assert!(
        !format!("{stdout}{stderr}").contains(PLANTED),
        "the link's target was read as the list"
    );
}

#[test]
fn the_preview_runs_no_program_the_repository_configures() {
    let (dir, root) = project();
    put(&root, "README.md", "x\n");
    common::git::git_ok(&root, &["add", "README.md"]);
    let marker = dir.path().join("fsmonitor-ran");
    let program = dir.path().join("fsmonitor.sh");
    fs::write(&program, format!("#!/bin/sh\n: > '{}'\n", marker.display())).expect("write program");
    fs::set_permissions(&program, fs::Permissions::from_mode(0o755)).expect("chmod");
    common::git::git_ok(
        &root,
        &["config", "core.fsmonitor", program.to_str().expect("utf-8")],
    );

    let (_, _, envelope) = preview(&root);
    assert_eq!(envelope["result"]["tracked_files"], "known");
    assert!(!marker.exists(), "the preview ran the repository's program");
}

#[test]
fn the_walk_is_bounded_and_stays_in_the_project() {
    let (dir, root) = project();
    let outside = dir.path().join("outside");
    put(&outside, "stolen.pem", "x\n");
    put(&root, ".gitignore", "ignored/\n");
    put(&root, "ignored/deploy.key", "x\n");
    put(&root, "nested/.git/HEAD", "ref: refs/heads/main\n");
    put(&root, "nested/inner.pem", "x\n");
    put(&root, "linked-checkout/.git", "gitdir: /nowhere\n");
    put(&root, "linked-checkout/other.pem", "x\n");
    put(&root, ".git/hooks/own.pem", "x\n");
    symlink(&outside, root.join("linked-folder")).expect("folder link");

    let (_, _, envelope) = preview(&root);
    let result = &envelope["result"];
    assert_eq!(
        strings(&find(&result["key_rules"], "Read(*.key)")["matched"]),
        ["ignored/deploy.key"]
    );
    assert_eq!(
        rule_names(&result["key_rules"]),
        ["Read(*.key)"],
        "nothing from .git, a nested checkout or behind a link"
    );
    assert_eq!(
        strings(&result["walk"]["skipped_nested_checkouts"]),
        ["linked-checkout", "nested"]
    );
    assert_eq!(result["walk"]["bound"], 100_000);
    assert_eq!(result["walk"]["complete"], true);

    // Past the bound the list says it is partial; the walk takes shallow entries first, so a
    // key file at the root is found however large a folder below it is.
    for n in 0..40 {
        put(&root, &format!("bulk/file-{n:02}.txt"), "x\n");
    }
    put(&root, ".env", "x\n");
    let partial = moai_cli::secret_deny::plan_with_bound(&root, 12).expect("plan");
    assert_eq!(partial["walk"]["bound"], 12);
    assert_eq!(partial["walk"]["complete"], false);
    let note = partial["walk"]["note"].as_str().expect("note");
    assert!(note.contains("partial") && note.contains("12"), "{note}");
    assert_eq!(
        strings(&find(&partial["key_rules"], "Read(.env)")["matched"]),
        [".env"]
    );
}

#[test]
fn preview_states_rule_forms_and_trust_behavior() {
    let (_dir, root) = project();
    put(&root, ".env", "x\n");
    put(&root, "a.pem", "x\n");
    put(&root, ".claudeignore", "*.log\n!keep.log\nbuild/\n");

    let (_, _, envelope) = preview(&root);
    let result = &envelope["result"];
    let mut rules = rule_names(&result["key_rules"]);
    rules.extend(rule_names(&result["claudeignore"]["entries"]));
    assert_eq!(rules.len(), 5, "{rules:?}");
    for rule in &rules {
        assert!(rule.starts_with("Read("), "{rule}");
    }
    let applies = result["applies"].as_str().expect("applies");
    for part in ["trust", "headless", "--setting-sources user", "SDK"] {
        assert!(applies.contains(part), "{part} missing: {applies}");
    }
    let no_effect = result["claudeignore"]["note"].as_str().expect("note");
    assert!(no_effect.contains("no effect"), "{no_effect}");
}

/// SPEC-20261122-secret-file-deny-rules AC-5: a user `!` entry stands in the file already, a
/// carried-over negation is applied in a second run, and after each run every key-file rule is
/// behind the last `!` entry, entries that were in the file before included.
#[test]
fn key_rules_stay_after_every_negation() {
    let (_dir, root) = project();
    for rel in [".env", ".env.example", "certs/dev.pem", "certs/server.pem"] {
        put(&root, rel, "x\n");
    }
    put(&root, ".claudeignore", "!.env.example\n");
    // A user key rule stands before the user's negation, which carves one of the `*.pem` files.
    let user = [
        "Read(./.npmrc)",
        "Bash(rm:*)",
        "Read(!certs/dev.pem)",
        "Read(./notes/**)",
    ];
    write_settings(
        &root,
        &json!({"model": "opus", "permissions": {"deny": user}}),
    );

    // The preview names both carve-outs before anything is picked.
    let preview = plan_with(&root, &[], false);
    let settings = &preview["settings"];
    assert_eq!(settings["file_state"], "present");
    assert_eq!(
        settings["negations"],
        json!([{"index": 2, "entry": "Read(!certs/dev.pem)"}])
    );
    let own = carve(&preview["carve_outs"], "Read(!certs/dev.pem)");
    assert_eq!(own["rule"], "Read(*.pem)");
    assert_eq!(strings(&own["files"]), ["certs/dev.pem"]);
    assert_eq!(own["source"], "settings");
    let carried = carve(&preview["carve_outs"], "Read(!.env.example)");
    assert_eq!(carried["rule"], "Read(.env.*)");
    assert_eq!(strings(&carried["files"]), [".env.example"]);
    assert_eq!(carried["source"], "claudeignore");

    // Run one: a key rule is added; the user's key rule moves behind the user's negation.
    let one = ["Read(.env)"];
    let planned = plan_with(&root, &one, false);
    let foreseen = strings(&planned["apply"]["final_deny"]);
    assert_keys_after_negations(&foreseen);
    apply_ok(&root, &one, false);
    let deny = deny_list(&root);
    assert_eq!(deny, foreseen, "the file holds the list the preview showed");
    assert_keys_after_negations(&deny);
    assert_eq!(
        other_rules(&deny),
        ["Bash(rm:*)", "Read(!certs/dev.pem)", "Read(./notes/**)"],
        "the other entries keep their order: {deny:?}"
    );
    assert_eq!(key_rules(&deny), ["Read(./.npmrc)", "Read(.env)"]);

    // Run two: a carried-over negation and a key rule. The negation lands behind the entries
    // that were there, and every key rule, run one's included, stays behind it.
    let two = ["Read(*.pem)", "Read(!.env.example)"];
    let planned = plan_with(&root, &two, false);
    assert_eq!(
        carve(&planned["carve_outs"], "Read(!.env.example)")["source"],
        "claudeignore",
        "the carve-out is still named before the pick"
    );
    let foreseen = strings(&planned["apply"]["final_deny"]);
    assert_keys_after_negations(&foreseen);
    apply_ok(&root, &two, false);
    let deny = deny_list(&root);
    assert_eq!(deny, foreseen, "the file holds the list the preview showed");
    assert_keys_after_negations(&deny);
    assert_eq!(
        other_rules(&deny),
        [
            "Bash(rm:*)",
            "Read(!certs/dev.pem)",
            "Read(./notes/**)",
            "Read(!.env.example)"
        ],
        "a new negation follows the entries that were there: {deny:?}"
    );
    assert_eq!(
        key_rules(&deny),
        ["Read(./.npmrc)", "Read(.env)", "Read(*.pem)"],
        "key rules keep their order, new ones follow: {deny:?}"
    );
    assert_eq!(read_settings(&root)["model"], "opus");

    // The same picks again leave the list as it is: no rule twice, no move.
    let settled = read_settings(&root);
    let planned = plan_with(&root, &two, false);
    assert_eq!(strings(&planned["apply"]["final_deny"]), deny);
    let again = apply_run(&root, &digest_of(&planned), &two, false);
    assert!(
        again.operation().starts_with("secret_deny."),
        "{}",
        again.text
    );
    assert_eq!(read_settings(&root), settled, "{}", again.text);
    assert_eq!(deny_list(&root), deny);
}

/// A key rule whose name is longer than a shape (`.env.production`, which the preview recommends
/// for its credential shape) is a key rule too: when a user negation follows it, it moves behind.
#[test]
fn a_longer_credential_name_is_a_key_rule_too() {
    let (_dir, root) = project();
    put(&root, ".env.production", "x\n");
    put(&root, ".claudeignore", ".env.production\n");
    write_settings(
        &root,
        &json!({"permissions": {"deny": [
            "Read(.env.production)",
            "Read(!.env.example)",
            "Read(./notes/**)",
        ]}}),
    );
    let picks = ["Read(.env.production)"];
    let preview = plan_with(&root, &picks, false);
    assert_eq!(
        strings(&preview["apply"]["final_deny"]),
        [
            "Read(!.env.example)",
            "Read(.env.production)",
            "Read(./notes/**)"
        ]
    );
    apply_ok(&root, &picks, false);
    assert_eq!(
        deny_list(&root),
        [
            "Read(!.env.example)",
            "Read(.env.production)",
            "Read(./notes/**)"
        ]
    );
}

/// SPEC-20261122-secret-file-deny-rules AC-6: the apply adds rules to a file that holds other
/// keys and user permissions, keeps the original, and takes no ownership of `permissions`.
#[test]
fn apply_adds_rules_without_taking_ownership() {
    let (_dir, root) = project();
    put(&root, ".env", "x\n");
    write_settings(
        &root,
        &json!({
            "theme": "dark",
            "env": {"EDITOR": "vi"},
            "permissions": {
                "allow": ["Bash(git status)", "Read(./docs/**)"],
                "ask": ["Bash(git push:*)"],
                "deny": ["Bash(rm:*)", "Read(./secrets/**)"],
            },
        }),
    );

    // A receipt that owns only `model`, made by the settings writer itself.
    let home = TempDir::new().expect("settings home");
    let project_arg = root.to_str().expect("utf-8");
    let desired = json!({"model": "opus"});
    let (code, planned) = claude_settings(
        &["claude-settings-plan", project_arg],
        home.path(),
        &desired,
    );
    assert_eq!(code, Some(0), "{planned}");
    let digest = planned["result"]["plan_digest"].as_str().expect("digest");
    let (code, applied) = claude_settings(
        &[
            "claude-settings-apply",
            project_arg,
            "--plan-digest",
            digest,
            "--permission-reference",
            "test-secret-deny",
        ],
        home.path(),
        &desired,
    );
    assert_eq!(code, Some(0), "{applied}");
    let receipt_path = moai_cli::state_root::path(&root, "manifests/claude-settings.json");
    let read_receipt = || -> Value {
        serde_json::from_slice(&fs::read(&receipt_path).expect("receipt")).expect("receipt JSON")
    };
    let receipt_before = read_receipt();
    assert_eq!(receipt_before["owned"], json!({"model": "opus"}));
    let original = fs::read(root.join(".claude/settings.json")).expect("settings file");
    let earlier = read_settings(&root);

    apply_ok(&root, &["Read(.env)"], false);

    // Every earlier value is kept; only the picked rule is new.
    let after = read_settings(&root);
    for key in ["theme", "env", "model"] {
        assert_eq!(after[key], earlier[key], "{key}");
    }
    assert_eq!(
        after.as_object().expect("object").len(),
        earlier.as_object().expect("object").len(),
        "no key was added: {after}"
    );
    for list in ["allow", "ask"] {
        assert_eq!(
            after["permissions"][list], earlier["permissions"][list],
            "{list}"
        );
    }
    assert_eq!(
        strings(&after["permissions"]["deny"]),
        ["Bash(rm:*)", "Read(./secrets/**)", "Read(.env)"]
    );

    // The original bytes are in the settings backup folder.
    let kept = backed_up(&root);
    assert!(
        kept.iter().any(|(_, bytes)| *bytes == original),
        "the original file is not in the backup folder: {:?}",
        kept.iter().map(|(path, _)| path).collect::<Vec<_>>()
    );

    // MoAI owns what it owned before and nothing of `permissions`.
    let receipt_after = read_receipt();
    assert_eq!(receipt_after["owned"], receipt_before["owned"]);
    assert!(receipt_after["owned"].get("permissions").is_none());
    // Its file fingerprint follows the file the apply wrote, and no interrupted-apply mark stays.
    let written = fs::read(root.join(".claude/settings.json")).expect("settings file");
    assert_ne!(written, original, "the apply changed the file");
    assert_eq!(
        receipt_after["settings_sha256"],
        json!(moai_cli::digest(&written)),
        "{receipt_after}"
    );
    assert!(receipt_after.get("pending").is_none(), "{receipt_after}");
    // So a request for the same key without `permissions` plans no removal of it.
    let (code, again) = claude_settings(
        &["claude-settings-plan", project_arg],
        home.path(),
        &desired,
    );
    assert_eq!(code, Some(0), "{again}");
    assert_eq!(again["result"]["action"], "unchanged", "{again}");
    assert_eq!(again["result"]["conflicts"], json!([]), "{again}");
    assert_eq!(again["result"]["planned_sha256"], Value::Null, "{again}");
}

/// SPEC-20261122-secret-file-deny-rules AC-7: while git ignores the root settings file the apply
/// writes nothing and names the fix; a file git does not ignore is written and the answer says
/// to commit it; outside a repository it writes and says unknown.
#[test]
fn ignored_root_settings_refuses_the_write() {
    // (a) The root settings file is ignored: nothing changes, not even a retired list.
    let (_dir, root) = project();
    put(&root, ".gitignore", ".claude/\n");
    put(&root, ".env", "x\n");
    put(&root, ".claudeignore", ".env\n");
    write_settings(&root, &json!({"model": "opus"}));
    let before = tree(&root);
    let pick = ["Read(.env)"];
    for retire in [false, true] {
        let preview = plan_with(&root, &pick, retire);
        assert_eq!(preview["settings"]["settings_ignored"], "ignored");
        let run = apply_run(&root, &digest_of(&preview), &pick, retire);
        assert_refused(&run, &["track"]);
        let message = run.message();
        assert!(message.contains(".claude/settings.json"), "{message}");
        assert!(
            !message.contains("settings.local") && !message.contains("local settings"),
            "the message names only the fix of tracking the file: {message}"
        );
        assert_eq!(tree(&root), before, "retire={retire}: nothing was written");
    }

    // (b) Untracked but not ignored: written, and the answer says the rules reach worktrees once
    // the file is committed.
    let (_dir_b, root_b) = project();
    put(&root_b, ".env", "x\n");
    write_settings(&root_b, &json!({"model": "opus"}));
    let preview = plan_with(&root_b, &pick, false);
    assert_eq!(preview["settings"]["settings_ignored"], "not_ignored");
    let run = apply_ok(&root_b, &pick, false);
    let message = run.message();
    assert!(
        message.contains("commit") && message.contains("worktree"),
        "{message}"
    );
    assert_eq!(deny_list(&root_b), ["Read(.env)"]);
    assert_eq!(read_settings(&root_b)["model"], "opus");
    // A file that is absent is created holding only the rule.
    let (_dir_c, root_c) = project();
    put(&root_c, ".env", "x\n");
    apply_ok(&root_c, &pick, false);
    assert_eq!(
        read_settings(&root_c),
        json!({"permissions": {"deny": ["Read(.env)"]}})
    );

    // (c) Not a git repository: written, and the answer says unknown.
    let dir_d = TempDir::new().expect("temp dir");
    let root_d = dir_d.path().join("project");
    fs::create_dir_all(&root_d).expect("project folder");
    assert!(
        !common::git::git_raw(&root_d, &["rev-parse", "--git-dir"]).success,
        "the temporary folder lies inside a repository; point TMPDIR elsewhere"
    );
    put(&root_d, ".env", "x\n");
    let preview = plan_with(&root_d, &pick, false);
    assert_eq!(preview["settings"]["settings_ignored"], "unknown");
    let run = apply_ok(&root_d, &pick, false);
    assert!(run.message().contains("unknown"), "{}", run.text);
    assert_eq!(deny_list(&root_d), ["Read(.env)"]);
}

/// SPEC-20261122-secret-file-deny-rules AC-8: `.claudeignore` stays unless the separate pick to
/// retire it is made; then it moves into the backup folder with the same bytes; a link is refused.
#[test]
fn retiring_claudeignore_needs_the_second_pick() {
    // Not exactly a list a person would write: a comment in Korean, no final newline.
    const LIST: &[u8] = "# 그대로 옮겨요\n.env\n*.log\nlast-line-without-newline".as_bytes();
    let fixture = || {
        let (dir, root) = project();
        put(&root, ".env", "x\n");
        fs::write(root.join(".claudeignore"), LIST).expect("list");
        (dir, root)
    };
    let pick = ["Read(.env)"];

    // Without the flag the file stays where it is and the backup folder does not hold it.
    let (_dir, root) = fixture();
    let preview = plan_with(&root, &pick, false);
    assert_eq!(preview["apply"]["retire_claudeignore"], false);
    apply_ok(&root, &pick, false);
    assert_eq!(
        fs::read(root.join(".claudeignore")).expect("list stays"),
        LIST
    );
    assert!(
        backed_up(&root).iter().all(|(_, bytes)| bytes != LIST),
        "the list was moved without the pick"
    );

    // With the flag the root file is gone and the backup folder holds its bytes.
    let (_dir, root) = fixture();
    let preview = plan_with(&root, &pick, true);
    assert_eq!(preview["apply"]["retire_claudeignore"], true);
    apply_ok(&root, &pick, true);
    assert_eq!(
        fs::symlink_metadata(root.join(".claudeignore"))
            .map(|_| ())
            .map_err(|e| e.kind()),
        Err(std::io::ErrorKind::NotFound),
        "the retired list is gone from the root"
    );
    let moved: Vec<_> = backed_up(&root)
        .into_iter()
        .filter(|(_, bytes)| bytes == LIST)
        .collect();
    assert_eq!(moved.len(), 1, "one copy of the list in the backup folder");
    assert_eq!(deny_list(&root), ["Read(.env)"]);

    // A linked `.claudeignore` is refused, at the preview or at the apply, and nothing moves.
    let (dir, root) = project();
    put(&root, ".env", "x\n");
    let outside = dir.path().join("outside");
    put(&outside, "list-target.txt", ".env\n");
    symlink(outside.join("list-target.txt"), root.join(".claudeignore")).expect("list link");
    let (before, before_outside) = (tree(&root), tree(&outside));
    let planned = plan_run(&root, &pick, true);
    assert_eq!(planned.operation(), "secret_deny.plan", "{}", planned.text);
    if planned.code == Some(0) {
        let run = apply_run(&root, &digest_of(&planned.envelope["result"]), &pick, true);
        assert_refused(&run, &["link"]);
    } else {
        assert!(planned.message().contains("link"), "{}", planned.text);
    }
    assert_eq!(tree(&root), before, "nothing moved, nothing was written");
    assert_eq!(
        tree(&outside),
        before_outside,
        "the link target is untouched"
    );
}

/// SPEC-20261122-secret-file-deny-rules REQ-7 (dangerous path): a digest from another pick set,
/// a digest the file changed under, and a settings file that is no JSON object write nothing.
#[test]
fn apply_refuses_a_stale_digest_and_a_non_object_settings_file() {
    let (_dir, root) = project();
    put(&root, ".env", "x\n");
    put(&root, "apps/web/.env.local", "x\n");
    write_settings(&root, &json!({"model": "opus"}));
    let reasons = ["stale", "digest", "re-preview"];

    // A digest made for one pick, used for two.
    let one = ["Read(.env)"];
    let digest = digest_of(&plan_with(&root, &one, false));
    let before = tree(&root);
    let run = apply_run(&root, &digest, &["Read(.env)", "Read(.env.*)"], false);
    assert_refused(&run, &reasons);
    assert_eq!(tree(&root), before, "a digest of another pick set wrote");

    // A digest the settings file changed under.
    let digest = digest_of(&plan_with(&root, &one, false));
    write_settings(&root, &json!({"model": "changed-after-the-preview"}));
    let before = tree(&root);
    let run = apply_run(&root, &digest, &one, false);
    assert_refused(&run, &reasons);
    assert_eq!(tree(&root), before, "a digest of an earlier file wrote");

    // A settings file that is a JSON array: refused at the preview or at the apply.
    let (_dir, root) = project();
    put(&root, ".env", "x\n");
    put(&root, ".claude/settings.json", "[]\n");
    let before = tree(&root);
    let planned = plan_run(&root, &one, false);
    assert_eq!(planned.operation(), "secret_deny.plan", "{}", planned.text);
    let digest = planned.envelope["result"]["apply"]["plan_digest"]
        .as_str()
        .map_or_else(|| format!("sha256:{}", "0".repeat(64)), str::to_string);
    let run = apply_run(&root, &digest, &one, false);
    assert_refused(&run, &["object", "preservation", "json"]);
    assert_eq!(
        tree(&root),
        before,
        "a non-object settings file was written"
    );
}

/// SPEC-20261122-secret-file-deny-rules REQ-7 (dangerous path): a settings file or a `.claude`
/// folder that became a link after a valid preview is refused, and nothing behind it changes.
/// The link leads to a copy with the same bytes, so the digest still matches.
#[test]
fn apply_refuses_a_linked_settings_file_or_claude_folder() {
    let one = ["Read(.env)"];
    let reasons = ["link"];

    // A linked settings file.
    let (dir, root) = project();
    put(&root, ".env", "x\n");
    write_settings(&root, &json!({"model": "keep"}));
    let digest = digest_of(&plan_with(&root, &one, false));
    let outside = dir.path().join("outside");
    let settings = root.join(".claude/settings.json");
    put(
        &outside,
        "real.json",
        &fs::read_to_string(&settings).expect("settings text"),
    );
    fs::remove_file(&settings).expect("remove file");
    symlink(outside.join("real.json"), &settings).expect("file link");
    let (before, before_outside) = (tree(&root), tree(&outside));
    let run = apply_run(&root, &digest, &one, false);
    assert_refused(&run, &reasons);
    assert_eq!(tree(&root), before, "the project changed");
    assert_eq!(tree(&outside), before_outside, "the link target changed");
    assert!(fs::symlink_metadata(&settings)
        .expect("link stays")
        .file_type()
        .is_symlink());

    // A linked `.claude` folder.
    let (dir, root) = project();
    put(&root, ".env", "x\n");
    write_settings(&root, &json!({"model": "keep"}));
    let digest = digest_of(&plan_with(&root, &one, false));
    let outside = dir.path().join("outside");
    fs::create_dir_all(&outside).expect("outside folder");
    fs::rename(root.join(".claude"), outside.join("claude")).expect("move folder");
    symlink(outside.join("claude"), root.join(".claude")).expect("folder link");
    let (before, before_outside) = (tree(&root), tree(&outside));
    let run = apply_run(&root, &digest, &one, false);
    assert_refused(&run, &reasons);
    assert_eq!(tree(&root), before, "the project changed");
    assert_eq!(
        tree(&outside),
        before_outside,
        "the folder behind the link changed"
    );
    assert!(fs::symlink_metadata(root.join(".claude"))
        .expect("link stays")
        .file_type()
        .is_symlink());
}

/// The two route texts, compiled in so `tools/repo_scans.json` selects this target when one of
/// them changes.
const PROJECT_ROUTE: &str = include_str!("../skills/moai/workflows/project.md");
const UPDATE_ROUTE: &str = include_str!("../skills/moai/workflows/update.md");

/// The line of `text` that holds `marker`, or a failure that names the file.
fn route_line<'a>(text: &'a str, marker: &str, file: &str) -> &'a str {
    text.lines()
        .find(|l| l.contains(marker))
        .unwrap_or_else(|| panic!("{file} has no line with `{marker}`"))
}

/// Slice C, AC-10 (REQ-11): the project-setup route offers the preview as one pick and applies a
/// rule only after the user's pick; the update route names the preview and has no apply step.
#[test]
fn the_routes_offer_the_preview_and_apply_only_on_a_pick() {
    let (project, update) = (PROJECT_ROUTE, UPDATE_ROUTE);

    let step = route_line(
        project,
        "SPEC-20261122-secret-file-deny-rules",
        "project.md",
    );
    for needle in [
        "moai secret-deny-plan <project>",
        "read-only",
        "one selection",
        "moai secret-deny-apply <project>",
        "--plan-digest <plan_digest>",
        "--permission-reference <ref>",
        "--pick",
        "--retire-claudeignore",
        "guarded",
        "only after the user",
        "declined or unanswered",
    ] {
        assert!(step.contains(needle), "project.md step lacks `{needle}`");
    }
    // The retire pick is a second, separate pick, and the route never applies without a pick.
    assert!(
        step.contains("separate pick"),
        "the retire pick is not separate"
    );

    let named = route_line(update, "secret-deny-plan", "update.md");
    for needle in ["read-only", "/moai project", "applies nothing"] {
        assert!(named.contains(needle), "update.md line lacks `{needle}`");
    }
    for apply_word in ["secret-deny-apply", "--pick", "--retire-claudeignore"] {
        assert!(
            !update.contains(apply_word),
            "update.md must not carry `{apply_word}`: the update route has no apply step for it"
        );
    }
}
