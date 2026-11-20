//! Step execution (the Go build's `executeStep` and `runStep`): skip rules, a child run without
//! a shell in its own process group, a time limit, the failure text.

use super::config::GateConfig;
use super::detect::SKIP_DIRS;
use super::summary::{command_line, Summary, REASON_DISABLED};
use super::table::{Step, PYTEST};
use std::ffi::{OsStr, OsString};
use std::io::Read;
use std::os::unix::fs::PermissionsExt;
use std::os::unix::process::{CommandExt, ExitStatusExt};
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant, SystemTime};

use super::fmtutil::go_duration;

/// How long a step's output pipes may stay open after the step itself has ended.
const WAIT_GRACE: Duration = Duration::from_secs(2);
const DESCENDANT_NOTE: &str =
    "termination was signalled to the step's whole process group, so processes the step started were signalled too";
/// Entries one source-file scan visits before it gives up and lets the step run.
const SOURCE_SCAN_ENTRY_CAP: usize = 20000;

/// The git variables that would point a step's git calls at another repository.
const GIT_ENV_REMOVED: &[&str] = &[
    "GIT_DIR",
    "GIT_COMMON_DIR",
    "GIT_WORK_TREE",
    "GIT_INDEX_FILE",
    "GIT_OBJECT_DIRECTORY",
    "GIT_ALTERNATE_OBJECT_DIRECTORIES",
    "GIT_NAMESPACE",
    "GIT_PREFIX",
    "GIT_SUPER_PREFIX",
    "GIT_CEILING_DIRECTORIES",
    "GIT_QUARANTINE_PATH",
];

pub struct StepCtx {
    /// The whole run must end by then.
    pub deadline: Instant,
    /// The PATH the tools are looked up on and run with.
    pub path: OsString,
}

pub struct Runner<'a> {
    pub cfg: &'a GateConfig,
    pub ctx: StepCtx,
    pub summary: Summary,
    staged: Option<Option<Vec<String>>>,
}

fn is_executable(p: &Path) -> bool {
    p.metadata()
        .map(|m| !m.is_dir() && m.permissions().mode() & 0o111 != 0)
        .unwrap_or(false)
}

/// `exec.LookPath`: a name with a slash is used as it is; a bare name is searched on `path`.
pub fn which(binary: &str, path: &OsStr) -> Option<PathBuf> {
    if binary.contains('/') {
        return is_executable(Path::new(binary)).then(|| PathBuf::from(binary));
    }
    for dir in std::env::split_paths(path) {
        let dir = if dir.as_os_str().is_empty() {
            PathBuf::from(".")
        } else {
            dir
        };
        let cand = dir.join(binary);
        if is_executable(&cand) {
            return Some(cand);
        }
    }
    None
}

fn extension(name: &str) -> &str {
    name.rfind('.').map(|i| &name[i..]).unwrap_or("")
}

fn has_ext(name: &str, exts: &[String]) -> bool {
    let ext = extension(name);
    exts.iter().any(|w| ext.eq_ignore_ascii_case(w))
}

/// `(found, determined)`: whether the module holds a file with one of the extensions; not
/// determined when the walk hit its entry cap or the root cannot be read.
pub fn project_has_source_file(dir: &Path, exts: &[String]) -> (bool, bool) {
    let mut visited = 0usize;
    let mut stack = vec![dir.to_path_buf()];
    let mut first = true;
    while let Some(d) = stack.pop() {
        let rd = match std::fs::read_dir(&d) {
            Ok(rd) => rd,
            Err(_) if first => return (false, false),
            Err(_) => continue,
        };
        first = false;
        visited += 1;
        let mut entries: Vec<_> = rd.flatten().collect();
        entries.sort_by_key(|e| e.file_name());
        // Reverse so the stack pops in name order, like the Go walk.
        for e in entries.into_iter().rev() {
            visited += 1;
            if visited > SOURCE_SCAN_ENTRY_CAP {
                return (false, false);
            }
            let name = e.file_name().to_string_lossy().into_owned();
            match e.file_type() {
                Ok(t) if t.is_dir() => {
                    if !SKIP_DIRS.contains(&name.as_str()) {
                        stack.push(e.path());
                    }
                }
                Ok(_) if has_ext(&name, exts) => return (true, true),
                _ => {}
            }
        }
    }
    (false, true)
}

fn kill_group(pid: u32) {
    if pid > 0 {
        unsafe { libc::kill(-(pid as libc::pid_t), libc::SIGKILL) };
    }
}

fn is_dotnet_restore_failure(output: &str) -> bool {
    [
        "Restore operation failed",
        "NU1202",
        "NETSDK1005",
        "not supported on this platform",
    ]
    .iter()
    .any(|m| output.contains(m))
}

impl<'a> Runner<'a> {
    pub fn new(cfg: &'a GateConfig, ctx: StepCtx) -> Self {
        Runner {
            cfg,
            ctx,
            summary: Summary::new(),
            staged: None,
        }
    }

    fn git_env(&self, cmd: &mut Command) {
        for key in GIT_ENV_REMOVED {
            cmd.env_remove(key);
        }
        cmd.env("PATH", &self.ctx.path);
    }

    /// Staged file names, or `None` when git cannot say (the caller then runs the step).
    fn staged_files(&mut self, dir: &Path) -> Option<&Vec<String>> {
        if self.staged.is_none() {
            let files = (|| {
                let git = which("git", &self.ctx.path)?;
                let mut cmd = Command::new(git);
                cmd.args(["diff", "--cached", "--name-only"])
                    .current_dir(dir)
                    .stdin(Stdio::null())
                    .stderr(Stdio::null());
                self.git_env(&mut cmd);
                let out = cmd.output().ok().filter(|o| o.status.success())?;
                let list: Vec<String> = String::from_utf8_lossy(&out.stdout)
                    .lines()
                    .map(|l| l.trim().to_string())
                    .filter(|l| !l.is_empty())
                    .collect();
                (!list.is_empty()).then_some(list)
            })();
            self.staged = Some(files);
        }
        self.staged.as_ref().and_then(Option::as_ref)
    }

    /// Applies the skip rules, then runs the step: `(passed, text)`.
    pub fn execute_step(&mut self, step: &Step, dir: &Path, timeout: Duration) -> (bool, String) {
        if let Some((_, false)) = self
            .cfg
            .disabled_steps
            .iter()
            .find(|(k, _)| *k == step.name)
        {
            self.summary.disabled(&step.name, REASON_DISABLED);
            return (true, String::new());
        }
        if step.optional && which(&step.binary, &self.ctx.path).is_none() {
            self.summary.skipped(
                &step.name,
                &format!(
                    "the optional tool \"{}\" was not found on PATH",
                    step.binary
                ),
            );
            return (true, String::new());
        }
        if !step.config_files.is_empty()
            && !step
                .config_files
                .iter()
                .any(|f| dir.join(f).metadata().map(|m| !m.is_dir()).unwrap_or(false))
        {
            self.summary.skipped(
                &step.name,
                &format!(
                    "none of its config files exist in the project directory ({})",
                    step.config_files.join(", ")
                ),
            );
            return (true, String::new());
        }
        if !step.changed_exts.is_empty() {
            let exts = step.changed_exts.clone();
            if let Some(staged) = self.staged_files(dir) {
                if !staged.iter().any(|f| has_ext(f, &exts)) {
                    self.summary.skipped(
                        &step.name,
                        &format!(
                            "no staged file carries one of its extensions ({})",
                            exts.join(", ")
                        ),
                    );
                    return (true, String::new());
                }
            }
        }
        if !step.source_exts.is_empty() {
            let (found, determined) = project_has_source_file(dir, &step.source_exts);
            if determined && !found {
                self.summary.skipped(
                    &step.name,
                    &format!(
                        "the project holds no source file with one of its extensions ({})",
                        step.source_exts.join(", ")
                    ),
                );
                return (true, String::new());
            }
        }
        self.run_step(&step.name, dir, timeout, &step.binary, &step.args)
    }

    fn run_step(
        &mut self,
        name: &str,
        dir: &Path,
        timeout: Duration,
        binary: &str,
        args: &[String],
    ) -> (bool, String) {
        let started_wall = SystemTime::now();
        let started = Instant::now();
        let step_deadline = started + timeout;
        let parent_binds = self.ctx.deadline <= step_deadline;
        let parent_budget = self.ctx.deadline.saturating_duration_since(started);
        let deadline = step_deadline.min(self.ctx.deadline);
        let result = self.spawn_and_wait(
            name,
            dir,
            binary,
            args,
            deadline,
            timeout,
            parent_binds,
            parent_budget,
        );
        self.summary.executed(
            name,
            &command_line(binary, args),
            started_wall,
            started.elapsed(),
        );
        result
    }

    #[allow(clippy::too_many_arguments)]
    fn spawn_and_wait(
        &self,
        name: &str,
        dir: &Path,
        binary: &str,
        args: &[String],
        deadline: Instant,
        timeout: Duration,
        parent_binds: bool,
        parent_budget: Duration,
    ) -> (bool, String) {
        let failed = |text: String| (false, format!("quality gate failed: {name}\n\n{text}"));
        let Some(exe) = which(binary, &self.ctx.path) else {
            return failed(format!(
                "exec: \"{binary}\": executable file not found in $PATH"
            ));
        };
        let mut cmd = Command::new(&exe);
        cmd.args(args)
            .current_dir(dir)
            .stdin(Stdio::null())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .process_group(0);
        self.git_env(&mut cmd);
        let mut child = match cmd.spawn() {
            Ok(c) => c,
            Err(e) => return failed(format!("fork/exec {}: {}", exe.display(), e)),
        };
        let pid = child.id();
        let buffers: [Arc<Mutex<Vec<u8>>>; 2] = [Arc::default(), Arc::default()];
        let (tx, rx) = std::sync::mpsc::channel::<()>();
        let mut pipes: Vec<Box<dyn Read + Send>> = Vec::new();
        if let Some(s) = child.stderr.take() {
            pipes.push(Box::new(s));
        }
        if let Some(s) = child.stdout.take() {
            pipes.push(Box::new(s));
        }
        let readers = pipes.len();
        for (i, mut pipe) in pipes.into_iter().enumerate() {
            let buf = Arc::clone(&buffers[i]);
            let tx = tx.clone();
            std::thread::spawn(move || {
                let mut chunk = [0u8; 8192];
                while let Ok(n) = pipe.read(&mut chunk) {
                    if n == 0 {
                        break;
                    }
                    buf.lock().unwrap().extend_from_slice(&chunk[..n]);
                }
                let _ = tx.send(());
            });
        }
        drop(tx);
        let mut killed = false;
        let status = loop {
            match child.try_wait() {
                Ok(Some(st)) => break st,
                Ok(None) if Instant::now() >= deadline => {
                    kill_group(pid);
                    killed = true;
                    match child.wait() {
                        Ok(st) => break st,
                        Err(e) => return failed(e.to_string()),
                    }
                }
                Ok(None) => std::thread::sleep(Duration::from_millis(5)),
                Err(e) => {
                    kill_group(pid);
                    return failed(e.to_string());
                }
            }
        };
        // The pipes close when the step and everything it started let go of them; a descendant
        // that still holds one after the grace is cut off with the group.
        let mut closed = 0;
        let until = Instant::now() + WAIT_GRACE;
        while closed < readers {
            let left = until.saturating_duration_since(Instant::now());
            if rx.recv_timeout(left).is_err() {
                break;
            }
            closed += 1;
        }
        let pipes_held = closed < readers;
        kill_group(pid);
        if status.success() {
            if pipes_held {
                return (
                    false,
                    format!(
                        "quality gate could not read {name} to completion: a process it started still held its output stream {} after it exited; {DESCENDANT_NOTE}",
                        go_duration(WAIT_GRACE)
                    ),
                );
            }
            return (true, String::new());
        }
        let text = |i: usize| String::from_utf8_lossy(&buffers[i].lock().unwrap()).into_owned();
        let combined = format!("{}{}", text(0), text(1));
        let output = combined.trim().to_string();
        let code = status.code();
        if name == PYTEST && code == Some(5) {
            return (true, String::new());
        }
        if killed || Instant::now() >= deadline {
            if parent_binds {
                return (
                    false,
                    format!(
                        "quality gate timed out: the overall gate budget ran out while running {name} ({} of it remained when the step started; the step's own {} budget was not exceeded); {DESCENDANT_NOTE}",
                        go_duration(super::fmtutil::round_ms(parent_budget)),
                        go_duration(timeout)
                    ),
                );
            }
            return (
                false,
                format!(
                    "quality gate timed out: {name} exceeded {}; {DESCENDANT_NOTE}",
                    go_duration(timeout)
                ),
            );
        }
        if name.to_lowercase().contains("dotnet") && is_dotnet_restore_failure(&combined) {
            return (true, String::new());
        }
        let output = if output.is_empty() {
            match (code, status.signal()) {
                (Some(c), _) => format!("exit status {c}"),
                (None, Some(sig)) => format!("signal: {}", signal_name(sig)),
                _ => "exit status -1".to_string(),
            }
        } else {
            output
        };
        failed(output)
    }
}

fn signal_name(sig: i32) -> String {
    match sig {
        libc::SIGKILL => "killed".into(),
        libc::SIGTERM => "terminated".into(),
        libc::SIGSEGV => "segmentation fault".into(),
        libc::SIGABRT => "aborted".into(),
        libc::SIGINT => "interrupt".into(),
        other => other.to_string(),
    }
}
