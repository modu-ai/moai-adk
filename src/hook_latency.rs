//! `moai hook latency` (SPEC-20261123-hook-entry-skeleton): measures how long
//! `<bin> hook <name>` takes and prints p50, p90 and the maximum in milliseconds. It judges no
//! number: the targets of the plan belong to the slices that add the handlers.
//!
//! Two cases are measured per run: a working folder that is not a MoAI project, and a MoAI
//! project with a registered project folder under a temporary HOME. The folders are made under
//! the system temporary folder and removed afterwards; the real HOME is never read. With
//! `--values` the tool only summarises the given numbers and starts nothing.

use serde_json::{json, Value};
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::Instant;

const DEFAULT_SAMPLES: usize = 30;
const MAX_SAMPLES: usize = 1000;
const DEFAULT_NAME: &str = "session-context";

/// The nearest-rank percentile: the smallest value with at least `pct` percent of the values at
/// or below it. `None` for an empty list or a `pct` outside 1 to 100. Integer arithmetic on the
/// rank, so 90 percent of 20 values is exactly the 18th.
pub fn percentile(values: &[f64], pct: usize) -> Option<f64> {
    if values.is_empty() || pct == 0 || pct > 100 {
        return None;
    }
    let mut sorted = values.to_vec();
    sorted.sort_by(f64::total_cmp);
    let rank = (pct * sorted.len()).div_ceil(100);
    sorted.get(rank - 1).copied()
}

fn round3(ms: f64) -> f64 {
    (ms * 1000.0).round() / 1000.0
}

fn summary(values: &[f64]) -> Value {
    let max = values.iter().copied().fold(0.0, f64::max);
    json!({
        "samples": values.len(),
        "p50_ms": round3(percentile(values, 50).unwrap_or(0.0)),
        "p90_ms": round3(percentile(values, 90).unwrap_or(0.0)),
        "max_ms": round3(max),
    })
}

/// Millisecond values separated by commas: each a finite number of at least 0.
pub fn parse_values(list: &str) -> Result<Vec<f64>, String> {
    let mut out = Vec::new();
    for piece in list.split(',') {
        let piece = piece.trim();
        let value: f64 = piece
            .parse()
            .map_err(|_| format!("--values: `{piece}` is not a number"))?;
        if !value.is_finite() || value < 0.0 {
            return Err(format!(
                "--values: `{piece}` is not a finite number of at least 0"
            ));
        }
        out.push(value);
    }
    Ok(out)
}

/// `release`, `debug` or `unknown`, read from the folders of the binary's path.
pub fn profile_label(bin: &Path) -> &'static str {
    for part in bin.components().rev() {
        match part.as_os_str().to_str() {
            Some("release") => return "release",
            Some("debug") => return "debug",
            _ => {}
        }
    }
    "unknown"
}

/// Runs the tool with the words after `hook latency`; the answer is one JSON line.
pub fn run(args: &[String]) -> Result<String, String> {
    let (mut bin, mut samples, mut name, mut values) =
        (None, DEFAULT_SAMPLES, DEFAULT_NAME.to_owned(), None);
    let mut it = args.iter();
    while let Some(flag) = it.next() {
        let mut value = |what: &str| {
            it.next()
                .cloned()
                .ok_or_else(|| format!("{what} needs a value"))
        };
        match flag.as_str() {
            "--bin" => bin = Some(PathBuf::from(value("--bin")?)),
            "--samples" => {
                let raw = value("--samples")?;
                samples = raw
                    .parse()
                    .map_err(|_| format!("--samples: `{raw}` is not a whole number"))?;
            }
            "--name" => name = value("--name")?,
            "--values" => values = Some(value("--values")?),
            other => return Err(format!("unknown option `{other}`")),
        }
    }
    if let Some(list) = values {
        let values = parse_values(&list)?;
        let mut answer = json!({ "tool": "hook latency" });
        answer
            .as_object_mut()
            .expect("an object")
            .extend(summary(&values).as_object().expect("an object").clone());
        return Ok(answer.to_string());
    }
    let bin = bin.ok_or("--bin <path of the moai binary> is required")?;
    if !(1..=MAX_SAMPLES).contains(&samples) {
        return Err(format!("--samples must be 1 to {MAX_SAMPLES}"));
    }
    if !super::hook_entry::is_registered(&name) {
        return Err(format!(
            "--name must be one of {}",
            super::hook_entry::NAMES.join(", ")
        ));
    }
    if !fs::metadata(&bin).is_ok_and(|m| m.is_file()) {
        return Err(format!("--bin: `{}` is not a file", bin.display()));
    }
    let cases = measure(&bin, &name, samples)?;
    Ok(json!({
        "tool": "hook latency",
        "bin": bin.display().to_string(),
        "profile": profile_label(&bin),
        "name": name,
        "samples": samples,
        "cases": cases,
    })
    .to_string())
}

/// A good input for each name: the event the host would send.
fn fixture(name: &str) -> Vec<u8> {
    let value = match name {
        "session-context" => json!({"hook_event_name": "SessionStart", "source": "startup"}),
        "bash-boundary" => json!({"hook_event_name": "PreToolUse", "tool_name": "Bash",
            "tool_input": {"command": "ls"}}),
        _ => json!({"hook_event_name": "PreToolUse", "tool_name": "Write",
            "tool_input": {"file_path": "/tmp/x", "content": "x"}}),
    };
    value.to_string().into_bytes()
}

fn measure(bin: &Path, name: &str, samples: usize) -> Result<Vec<Value>, String> {
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map_or(0, |d| d.as_nanos());
    let root =
        std::env::temp_dir().join(format!("moai-hook-latency-{}-{nanos}", std::process::id()));
    let result = measure_in(&root, bin, name, samples);
    // The folder was made by this call just above, under the temporary folder.
    let _ = fs::remove_dir_all(&root);
    result
}

fn measure_in(root: &Path, bin: &Path, name: &str, samples: usize) -> Result<Vec<Value>, String> {
    let (plain, project, home) = (root.join("plain"), root.join("project"), root.join("home"));
    let made = (|| -> std::io::Result<()> {
        fs::create_dir(root)?;
        fs::create_dir_all(&plain)?;
        fs::create_dir_all(project.join(".moai"))?;
        fs::write(
            project.join(".moai/project.yaml"),
            "{\"schema_version\": 2, \"project_id\": \"latency\"}\n",
        )?;
        fs::create_dir_all(home.join(".moai/projects/latency"))
    })();
    made.map_err(|e| format!("cannot prepare the measuring folders: {e}"))?;
    let input = fixture(name);
    let mut cases = Vec::new();
    for (case, cwd) in [("not_moai_project", &plain), ("moai_project", &project)] {
        let mut values = Vec::with_capacity(samples);
        for _ in 0..samples {
            values.push(sample(bin, name, cwd, &home, &input)?);
        }
        let mut entry = json!({ "case": case });
        entry
            .as_object_mut()
            .expect("an object")
            .extend(summary(&values).as_object().expect("an object").clone());
        cases.push(entry);
    }
    Ok(cases)
}

/// One run of `<bin> hook <name>`, in milliseconds from before the start to after the exit.
fn sample(bin: &Path, name: &str, cwd: &Path, home: &Path, input: &[u8]) -> Result<f64, String> {
    let start = Instant::now();
    let mut child = Command::new(bin)
        .args(["hook", name])
        .current_dir(cwd)
        .env("HOME", home)
        .env_remove("CLAUDE_PROJECT_DIR")
        .stdin(Stdio::piped())
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .spawn()
        .map_err(|e| format!("cannot start `{}`: {e}", bin.display()))?;
    if let Some(mut stdin) = child.stdin.take() {
        let _ = stdin.write_all(input);
    }
    let status = child
        .wait()
        .map_err(|e| format!("cannot wait for `{}`: {e}", bin.display()))?;
    let ms = start.elapsed().as_secs_f64() * 1000.0;
    if status.success() {
        Ok(ms)
    } else {
        Err(format!("`hook {name}` ended with {status}, not 0"))
    }
}
