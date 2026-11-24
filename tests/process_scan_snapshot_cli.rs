//! The process scans and the deletion re-check (SPEC-20261121-process-scan-snapshot): tests
//! inject the table from a file, a stuck `lsof` is stopped, and the re-check right before a
//! deletion runs its own scan — through the binary in a temporary HOME or folder only.

mod common;
use common::home::TempHome;
use common::{git, MOAI_BIN};
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::Command,
    time::{Duration, Instant},
};

const NOW: &str = "2026-11-21T05:00:00Z";

/// A temporary HOME whose `.moai` looks like the Go generation's.
fn go_home() -> TempHome {
    let home = TempHome::new();
    for (rel, body) in [
        (
            ".moai/user-assets.json",
            r#"{"schema_version": 1, "files": {}}"#,
        ),
        (".moai/state/active-sessions.json", "[]"),
        (".moai/credentials.yaml", "token: PLACEHOLDER\n"),
        (".moai/config/participation.yaml", "x: 1\n"),
    ] {
        let p = home.path().join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }
    home
}

/// A `lsof` script that logs its arguments, one line per call, then runs `then`.
fn stub_lsof(dir: &Path, log: &Path, then: &str) {
    fs::create_dir_all(dir).unwrap();
    let script = dir.join("lsof");
    fs::write(
        &script,
        format!("#!/bin/sh\necho \"$*\" >> '{}'\n{then}\n", log.display()),
    )
    .unwrap();
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(&script, fs::Permissions::from_mode(0o755)).unwrap();
    }
}

fn plan(home: &TempHome, stub_dir: &Path, env: &[(&str, &str)]) -> Value {
    let path = format!("{}:/usr/bin:/bin", stub_dir.display());
    let mut c = home.command(&[
        "home-state-plan",
        "--home",
        home.path().to_str().unwrap(),
        "--now",
        NOW,
    ]);
    c.env("PATH", path);
    // These tests judge the scan itself, so the clean-env default `empty` is replaced by the
    // real scan (a stub `lsof` on PATH); a test that injects a table file overrides it below.
    c.env(common::PROCESS_TABLE_SWITCH, "real");
    for (k, v) in env {
        c.env(k, v);
    }
    let out = c.output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn blocks(p: &Value) -> Vec<String> {
    p["blocks"]
        .as_array()
        .cloned()
        .unwrap_or_default()
        .iter()
        .filter_map(|b| b.as_str().map(str::to_string))
        .collect()
}

/// Linux reads `/proc` and never calls `lsof`.
fn reads_proc() -> bool {
    Path::new("/proc/self/fd").exists()
}

/// REQ-1: the working-folder check asks only for the cwd descriptors (`-a -d cwd`) and the
/// open-file check is the machine listing; each runs at most once per command, cheaply.
#[test]
fn the_cwd_check_asks_only_cwd_descriptors_and_the_open_check_the_machine_listing() {
    if reads_proc() {
        return;
    }
    let home = go_home();
    let stub = home.path().join("stub");
    let log = home.path().join("lsof.log");
    stub_lsof(
        &stub,
        &log,
        "printf 'fcwd\\nn/nowhere\\nf3\\nn/nowhere/file\\n'",
    );
    plan(&home, &stub, &[]);
    let calls = fs::read_to_string(&log).unwrap_or_default();
    let calls: Vec<&str> = calls.lines().collect();
    assert_eq!(calls.len(), 2, "{calls:?}");

    // The first scan answers the working-folder question: cwd descriptors only.
    let cwd_words: Vec<&str> = calls[0].split_whitespace().collect();
    for flag in ["-w", "-n", "-P", "-b", "-a", "-d", "cwd"] {
        assert!(
            cwd_words.contains(&flag),
            "{flag} is missing from `lsof {}`",
            calls[0]
        );
    }

    // The open-file scan is the machine listing, without the cwd filter.
    let open_words: Vec<&str> = calls[1].split_whitespace().collect();
    for flag in ["-w", "-n", "-P", "-b", "-F", "fn"] {
        assert!(
            open_words.contains(&flag),
            "{flag} is missing from `lsof {}`",
            calls[1]
        );
    }
    assert!(
        !open_words.contains(&"cwd"),
        "the open-file scan filters cwd: `lsof {}`",
        calls[1]
    );
}

/// REQ-5: a table file drives the decisions and `lsof` is never run.
#[test]
fn an_injected_table_file_drives_the_decisions_without_lsof() {
    if reads_proc() || !cfg!(debug_assertions) {
        return;
    }
    let home = go_home();
    let moai = fs::canonicalize(home.path().join(".moai")).unwrap();
    let stub = home.path().join("stub");
    let log = home.path().join("lsof.log");
    stub_lsof(&stub, &log, "echo 'n/nowhere'");
    let table = home.path().join("table.txt");

    // A working folder inside: named by the working-folder check (and, as lsof lists a working
    // folder among open files, by the open-file check too).
    fs::write(&table, format!("cwd {}\n", moai.join("config").display())).unwrap();
    let b = blocks(&plan(
        &home,
        &stub,
        &[("MOAI_TEST_PROCESS_TABLE", table.to_str().unwrap())],
    ));
    assert!(b.contains(&"in_use".to_string()), "{b:?}");

    // Only an open file inside.
    fs::write(
        &table,
        format!("open {}\n", moai.join("credentials.yaml").display()),
    )
    .unwrap();
    let b = blocks(&plan(
        &home,
        &stub,
        &[("MOAI_TEST_PROCESS_TABLE", table.to_str().unwrap())],
    ));
    assert_eq!(b, ["file_open_in_use"], "{b:?}");

    // A sibling whose name only starts the same, and paths elsewhere: nothing is blocked.
    let sibling = format!("{}-other/x", moai.display());
    fs::write(
        &table,
        format!("cwd {sibling}\nopen {sibling}\nopen /elsewhere/file\n"),
    )
    .unwrap();
    let b = blocks(&plan(
        &home,
        &stub,
        &[("MOAI_TEST_PROCESS_TABLE", table.to_str().unwrap())],
    ));
    assert!(b.is_empty(), "{b:?}");

    assert!(!log.exists(), "lsof ran: {:?}", fs::read_to_string(&log));
}

/// REQ-6: an `lsof` that never answers is stopped and the check says it cannot tell.
#[test]
fn a_stuck_lsof_is_stopped_and_the_check_reports_unavailable() {
    if reads_proc() || !cfg!(debug_assertions) {
        return;
    }
    let home = go_home();
    let stub = home.path().join("stub");
    let log = home.path().join("lsof.log");
    // A few lines and then a hang: the stop-by-deadline path, not the empty-output one.
    stub_lsof(
        &stub,
        &log,
        "printf 'p1\\nfcwd\\nn/nowhere\\n'; exec sleep 20",
    );
    let start = Instant::now();
    let p = plan(&home, &stub, &[("MOAI_TEST_LSOF_TIMEOUT_MS", "500")]);
    let took = start.elapsed();
    assert!(took < Duration::from_secs(12), "took {took:?}");
    let b = blocks(&p);
    assert!(b.iter().any(|x| x == "process_check_unavailable"), "{b:?}");
}

// ---------------------------------------------------------------------------
// The deletion re-check (worktree-gc-apply), on a synthetic repository only.
// ---------------------------------------------------------------------------

/// A day from now: every fixture change is older than the default window.
fn later() -> u64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs()
        + 86_400
}

/// A synthetic repository with two merged, clean, built worktrees `a` and `b` — the removal
/// candidates of `worktree-gc-plan`/`apply` (the worktree_gc_apply fixture, reduced) — in a
/// temporary folder only. Never this machine's real worktrees.
struct GcFixture {
    _dir: tempfile::TempDir,
    root: PathBuf,
    home: PathBuf,
    wts: [PathBuf; 2],
}

fn gc_fixture() -> GcFixture {
    let dir = tempfile::TempDir::new().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let root = base.join("repo");
    let home = base.join("home");
    fs::create_dir_all(&root).unwrap();
    fs::create_dir_all(&home).unwrap();
    let g = |args: &[&str], cwd: &Path| {
        let out = git::command()
            .args([
                "-c",
                "user.name=fixture",
                "-c",
                "user.email=fixture@example.invalid",
            ])
            .args(args)
            .current_dir(cwd)
            .output()
            .unwrap();
        assert!(
            out.status.success(),
            "git {}: {}",
            args.join(" "),
            String::from_utf8_lossy(&out.stderr)
        );
    };
    g(&["init", "-q", "-b", "main"], &root);
    fs::write(root.join(".gitignore"), ".moai-cli/\ntarget/\n").unwrap();
    fs::write(root.join("a.txt"), "a\n").unwrap();
    g(&["add", ".gitignore", "a.txt"], &root);
    g(&["commit", "-q", "-m", "init"], &root);
    g(&["branch", "integ"], &root);
    let mut wts = [PathBuf::new(), PathBuf::new()];
    for (i, name) in ["a", "b"].iter().enumerate() {
        let p = root.join(".moai-cli/worktrees").join(name);
        g(
            &[
                "worktree",
                "add",
                "-q",
                "-b",
                &format!("issue/1-{name}"),
                p.to_str().unwrap(),
            ],
            &root,
        );
        // Built and finished: cargo output only, nothing else in the worktree.
        fs::create_dir_all(p.join("target/debug/deps")).unwrap();
        fs::write(
            p.join("target/CACHEDIR.TAG"),
            "Signature: 8a477f597d28d172789f06886806bc55\n",
        )
        .unwrap();
        fs::write(p.join("target/debug/.cargo-lock"), "").unwrap();
        fs::write(p.join("target/debug/deps/lib.rlib"), vec![7u8; 4096]).unwrap();
        wts[i] = fs::canonicalize(&p).unwrap();
    }
    GcFixture {
        _dir: dir,
        root,
        home,
        wts,
    }
}

fn gc_run(fx: &GcFixture, args: &[String], stub_dir: Option<&Path>, env: &[(&str, &str)]) -> Value {
    let mut c = Command::new(MOAI_BIN);
    c.args(args)
        .current_dir(&fx.root)
        .env("HOME", &fx.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null");
    if let Some(stub) = stub_dir {
        c.env("PATH", format!("{}:/usr/bin:/bin", stub.display()));
    }
    for (k, v) in env {
        c.env(k, v);
    }
    let out = c.output().unwrap();
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "envelope on stdout; stderr {}",
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn gc_plan(fx: &GcFixture, now: u64, stub_dir: Option<&Path>, env: &[(&str, &str)]) -> Value {
    let args = [
        "worktree-gc-plan".to_string(),
        fx.root.to_str().unwrap().to_string(),
        "--integration-ref".into(),
        "integ".into(),
        "--now-unix".into(),
        now.to_string(),
    ];
    let v = gc_run(fx, &args, stub_dir, env);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn gc_apply(
    fx: &GcFixture,
    now: u64,
    digest: &str,
    stub_dir: Option<&Path>,
    env: &[(&str, &str)],
) -> Value {
    let args = [
        "worktree-gc-apply".to_string(),
        fx.root.to_str().unwrap().to_string(),
        "--integration-ref".into(),
        "integ".into(),
        "--now-unix".into(),
        now.to_string(),
        "--plan-digest".into(),
        digest.to_string(),
        "--permission-reference".into(),
        "fixture-approved".into(),
    ];
    let v = gc_run(fx, &args, stub_dir, env);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn entry<'a>(v: &'a Value, field: &str, name: &str) -> &'a Value {
    v[field]
        .as_array()
        .unwrap()
        .iter()
        .find(|c| {
            c["path"]
                .as_str()
                .is_some_and(|p| p.ends_with(&format!("/worktrees/{name}")))
        })
        .unwrap_or_else(|| panic!("no worktrees/{name} in {field}: {v}"))
}

/// B2: the re-check right before a deletion runs its own scan and never trusts the plan's
/// snapshot. From the second scan on the table says a process is working inside `b`: the
/// fresh read refuses `b`; a cached one would remove it.
#[test]
fn the_deletion_recheck_runs_its_own_scan() {
    if reads_proc() {
        return;
    }
    let fx = gc_fixture();
    let stub = fx.home.join("stub");
    let log = fx.home.join("lsof.log");
    let cnt = fx.home.join("cnt");
    fs::create_dir_all(&stub).unwrap();
    fs::create_dir_all(&cnt).unwrap();
    let wt_b = fx.wts[1].display();
    // The counter is per calling process (`$PPID` is the moai binary, one process per
    // command): every command's first scan — the plan command's and the apply command's
    // recomputation — sees an empty table; every scan after those (the deletion re-checks)
    // sees a process working inside `b`.
    fs::write(
        stub.join("lsof"),
        format!(
            "#!/bin/sh\nn=$(cat '{cnt}/'$PPID'.count' 2>/dev/null || echo 0)\necho \"$*\" >> '{log}'\necho $((n+1)) > '{cnt}/'$PPID'.count'\nif [ \"$n\" -eq 0 ]; then echo 'n/nowhere/at/all'; else printf 'fcwd\\nn/%s\\n' '{wt_b}'; fi\n",
            cnt = cnt.display(),
            log = log.display(),
        ),
    )
    .unwrap();
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(stub.join("lsof"), fs::Permissions::from_mode(0o755)).unwrap();
    }

    let now = later();
    let p = gc_plan(&fx, now, Some(&stub), &[]);
    // The plan's own scan saw nothing: both worktrees are removal candidates.
    assert_eq!(entry(&p, "candidates", "a")["action"], "remove", "{p}");
    assert_eq!(entry(&p, "candidates", "b")["action"], "remove", "{p}");
    let digest = p["plan_digest"].as_str().unwrap().to_string();

    let v = gc_apply(&fx, now, &digest, Some(&stub), &[]);
    let a = entry(&v, "results", "a");
    let b = entry(&v, "results", "b");
    assert_eq!(a["status"], "removed", "{v}");
    assert_eq!(b["status"], "refused", "{v}");
    assert!(
        b["detail"]
            .as_str()
            .unwrap_or_default()
            .contains("process_in_use"),
        "{v}"
    );
    assert!(fx.wts[1].exists(), "the refused worktree stayed in place");

    // And the re-check ran scans of its own, not the plan's one.
    let calls = fs::read_to_string(&log).unwrap_or_default().lines().count();
    assert!(
        calls >= 3,
        "expected the plan's scan plus the per-entry re-check scans, got {calls}"
    );
}

/// B3: an `lsof` that does not answer within its time limit is stopped, the table is unknown,
/// and the deletion re-check refuses — the worktree stays in place.
#[test]
fn an_lsof_cut_off_by_the_time_limit_refuses_the_removal() {
    if reads_proc() || !cfg!(debug_assertions) {
        return;
    }
    let fx = gc_fixture();
    // Every command's first scan answers (the plan command's and the apply command's
    // recomputation, per-process counter); every scan after those — the deletion re-check —
    // hangs past the limit.
    let stub = fx.home.join("stub");
    let cnt = fx.home.join("cnt");
    fs::create_dir_all(&stub).unwrap();
    fs::create_dir_all(&cnt).unwrap();
    fs::write(
        stub.join("lsof"),
        format!(
            "#!/bin/sh\nn=$(cat '{cnt}/'$PPID'.count' 2>/dev/null || echo 0)\necho $((n+1)) > '{cnt}/'$PPID'.count'\nif [ \"$n\" -eq 0 ]; then echo 'n/nowhere/at/all'; else printf 'p1\\nfcwd\\nn/nowhere\\n'; exec sleep 20; fi\n",
            cnt = cnt.display(),
        ),
    )
    .unwrap();
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(stub.join("lsof"), fs::Permissions::from_mode(0o755)).unwrap();
    }

    let now = later();
    let p = gc_plan(&fx, now, Some(&stub), &[]);
    assert_eq!(entry(&p, "candidates", "a")["action"], "remove", "{p}");
    let digest = p["plan_digest"].as_str().unwrap().to_string();

    let v = gc_apply(
        &fx,
        now,
        &digest,
        Some(&stub),
        &[("MOAI_TEST_LSOF_TIMEOUT_MS", "500")],
    );
    let a = entry(&v, "results", "a");
    assert_eq!(a["status"], "refused", "{v}");
    assert!(
        a["detail"]
            .as_str()
            .unwrap_or_default()
            .contains("process_check_unavailable"),
        "{v}"
    );
    assert!(fx.wts[0].exists(), "the worktree stayed in place");
}

// ---- REQ-15 (SPEC-20261124-cleanup-leftovers, slice E1): the re-check right before a home move ----

/// A `lsof` whose first two calls (one command's working-folder scan and open-file scan) see an
/// empty table and whose later calls see a process working inside `busy`. The counter is per
/// calling process (`$PPID` is the moai binary), so every command starts with its quiet pair:
/// the plan command's scan and the apply command's recomputation of the plan are quiet, a scan
/// after those — the re-check right before the move — sees the process.
#[cfg(unix)]
fn busy_after_the_first_read(stub: &Path, busy: &Path) {
    let cnt = stub.join("cnt");
    fs::create_dir_all(&cnt).unwrap();
    stub_lsof(
        stub,
        &stub.join("lsof.log"),
        &format!(
            "n=$(cat '{cnt}/'$PPID'.count' 2>/dev/null || echo 0); echo $((n+1)) > '{cnt}/'$PPID'.count'; if [ \"$n\" -lt 2 ]; then echo 'n/nowhere/at/all'; else printf 'fcwd\\nn%s\\n' '{busy}'; fi",
            cnt = cnt.display(),
            busy = busy.display(),
        ),
    );
}

/// One `moai` command in the temporary HOME with the stub `lsof` first on PATH and the real scan.
#[cfg(unix)]
fn home_json(home: &TempHome, stub: &Path, args: &[&str]) -> Value {
    let mut c = home.command(args);
    c.env("PATH", format!("{}:/usr/bin:/bin", stub.display()));
    c.env(common::PROCESS_TABLE_SWITCH, "real");
    let out = c.output().unwrap();
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)))
}

/// REQ-15: `home-retire-apply` reads the processes again right before it moves. A program that
/// started working in `cache` after the command's first scan keeps `cache` in place; a read of
/// the command cache would move it.
#[cfg(unix)]
#[test]
fn the_retire_move_recheck_runs_its_own_scan() {
    if reads_proc() {
        return;
    }
    let home = go_home();
    let h = home.path().to_str().unwrap().to_string();
    fs::create_dir_all(home.path().join(".moai/cache")).unwrap();
    fs::write(home.path().join(".moai/cache/a.bin"), "x").unwrap();
    let busy = fs::canonicalize(home.path().join(".moai/cache")).unwrap();
    let stub_dir = tempfile::TempDir::new().unwrap();
    let stub = stub_dir.path();
    busy_after_the_first_read(stub, &busy);

    let p = home_json(
        &home,
        stub,
        &["home-retire-plan", "--home", &h, "--now", NOW],
    );
    assert_eq!(p["ok"], json!(true), "{p}");
    assert!(
        p["result"]["counts"]["retire"].as_u64().unwrap_or(0) > 0,
        "the plan's own scan saw nothing, so `cache` is planned to move: {p}"
    );
    let digest = p["result"]["plan_digest"].as_str().unwrap().to_string();

    let a = home_json(
        &home,
        stub,
        &[
            "home-retire-apply",
            "--home",
            &h,
            "--now",
            NOW,
            "--plan-digest",
            &digest,
            "--permission-reference",
            "fixture-approved",
        ],
    );
    assert_eq!(a["ok"], json!(true), "{a}");
    assert!(
        a.to_string().contains("process_cwd"),
        "the move was skipped for the working process: {a}"
    );
    assert!(
        home.path().join(".moai/cache/a.bin").is_file(),
        "the busy entry stayed in place: {a}"
    );
}

/// REQ-15: `home-archive-apply` reads the processes again right before it moves the earlier
/// folder, after the plan was recomputed and the manifest written.
#[cfg(unix)]
#[test]
fn the_archive_move_recheck_runs_its_own_scan() {
    if reads_proc() {
        return;
    }
    let home = TempHome::new();
    let h = home.path().to_str().unwrap().to_string();
    for (rel, body) in [
        (".moai-cli/state/usage-limits.json", "{}\n"),
        (".moai-cli/receipts/r.json", "{}\n"),
        (".moai/state/active-sessions.json", "[]\n"),
    ] {
        let p = home.path().join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }
    let busy = fs::canonicalize(home.path().join(".moai-cli/state")).unwrap();
    let stub_dir = tempfile::TempDir::new().unwrap();
    let stub = stub_dir.path();
    busy_after_the_first_read(stub, &busy);

    let p = home_json(
        &home,
        stub,
        &["home-archive-plan", "--home", &h, "--now", NOW],
    );
    assert_eq!(p["ok"], json!(true), "{p}");
    assert_eq!(p["result"]["action"], json!("retire"), "{p}");
    let digest = p["result"]["plan_digest"].as_str().unwrap().to_string();

    let a = home_json(
        &home,
        stub,
        &[
            "home-archive-apply",
            "--home",
            &h,
            "--now",
            NOW,
            "--plan-digest",
            &digest,
            "--permission-reference",
            "fixture-approved",
        ],
    );
    assert!(
        a.to_string().contains("process_cwd"),
        "the move was skipped for the working process: {a}"
    );
    assert!(
        home.path()
            .join(".moai-cli/state/usage-limits.json")
            .is_file(),
        "the earlier folder stayed in place: {a}"
    );
}
