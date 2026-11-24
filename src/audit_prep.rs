//! Card-addressed audit slots (SPEC-20261124-audit-prep). `audit-prep <card> --head H`
//! takes a fixed audit slot for the card, switches it detached to the head, inherits the lane's
//! build and runs one incremental `cargo test --no-run` of the test targets named; `audit-done
//! <card>` finishes the card's slot and never discards or forces.
//!
//! The slot work is [`crate::audit_slot`]'s, called with the card as holder and no session: a card
//! is the same holder whichever session prepared or finishes its audit (D-6). The new part is the
//! build. It runs in the slot with the slot's own `target/`, without `RUSTC_WRAPPER` (an
//! incremental build and a wrapper clash, exit 101 in the 2026-11-22 audits), inside one heavy
//! slot, and reports what Cargo compiled and what it found fresh from its JSON messages. The
//! sources of the slot were touched by `audit-slot prep`, so the workspace units compile once from
//! the audited sources and a rebuilt test executable reads the slot's folder, never the lane's.

use crate::audit_slot::{self, Who};
use crate::worktree_slot::{self, Caps};
use serde_json::{json, Value};
use std::fs;
use std::io::{BufRead, BufReader, Write};
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

type Result<T> = std::result::Result<T, String>;

/// The most test targets one `audit-prep` builds (SPEC D-4: a mutant needs one or two).
pub const MAX_TESTS: usize = 2;
/// The last lines of a failed build that the refusal names.
const TAIL_LINES: usize = 30;

/// What `audit-prep` is asked to do.
pub struct Prep<'a> {
    pub project: &'a Path,
    pub card: &'a str,
    pub head: &'a str,
    pub lane: Option<&'a Path>,
    /// The `--test` targets to build; none builds the library.
    pub tests: &'a [String],
    pub pid: Option<i64>,
    pub lease_secs: u64,
    pub caps: Caps,
    /// The heavy slot folder (default: the machine's).
    pub slot_dir: Option<PathBuf>,
    pub max: u32,
    pub min_free_gb: u32,
}

/// A card name: `C-` and one to nine digits.
fn check_card(command: &str, card: &str) -> Result<()> {
    let ok = card
        .strip_prefix("C-")
        .is_some_and(|d| (1..=9).contains(&d.len()) && d.bytes().all(|b| b.is_ascii_digit()));
    if ok {
        Ok(())
    } else {
        Err(format!(
            "{command} refused card {card:?}: a card name is C- and digits, like C-1; nothing was changed"
        ))
    }
}

fn check_tests(tests: &[String]) -> Result<()> {
    if tests.len() > MAX_TESTS {
        return Err(format!(
            "audit-prep refused: {} --test targets named, at most two are built (a mutant needs one or two); nothing was changed",
            tests.len()
        ));
    }
    for t in tests {
        if t.is_empty()
            || !t
                .bytes()
                .all(|b| b.is_ascii_alphanumeric() || b == b'_' || b == b'-')
        {
            return Err(format!(
                "audit-prep refused --test {t:?}: a test target name is letters, digits, `_` or `-`; nothing was changed"
            ));
        }
    }
    Ok(())
}

/// A refusal of `audit_slot` in this command's name.
fn reword(e: String) -> String {
    e.replacen("audit-slot refused", "audit-prep refused", 1)
        .replace("run `done ", "run `audit-slot done ")
}

/// Takes the card's audit slot for the head and builds the named test targets once. See the
/// module notes; the answer is the slot's answer with `card` and `cargo` added.
pub fn prep(o: &Prep) -> Result<Value> {
    check_card("audit-prep", o.card)?;
    check_tests(o.tests)?;
    // The slot the card already holds, else the first one without a live holder.
    let name = audit_slot::slot_held_by(o.project, o.card)?
        .unwrap_or_else(|| audit_slot::AUTO.to_string());
    let who = Who {
        holder: o.card,
        pid: o.pid,
        lease_secs: o.lease_secs,
    };
    let mut answer = audit_slot::prep_with(o.project, &name, o.head, o.lane, o.caps, &who, None)
        .map_err(reword)?;
    let common = worktree_slot::common_dir(o.project)?;
    let primary = worktree_slot::primary_of(&common)?;
    let rel = answer["path"]
        .as_str()
        .ok_or("audit-prep: the slot answer names no path")?
        .to_string();
    let slot_name = answer["slot"].as_str().unwrap_or_default().to_string();
    let cargo = build(&primary.join(&rel), &rel, o).map_err(|e| {
        format!(
            "{e}; {slot_name} stays held by {} (run audit-done {} when the audit is over)",
            o.card, o.card
        )
    })?;
    answer["card"] = json!(o.card);
    answer["cargo"] = cargo;
    Ok(answer)
}

/// Finishes the audit of `card`: the slot it holds must be clean, and is then released with its
/// `target/` kept. A change left in the slot refuses it; there is no discard and no force.
pub fn done(project: &Path, card: &str) -> Result<Value> {
    check_card("audit-done", card)?;
    let Some(slot) = audit_slot::slot_held_by(project, card)? else {
        return Err(format!(
            "audit-done refused {card}: it holds no audit slot; nothing was changed"
        ));
    };
    let who = Who {
        holder: card,
        pid: None,
        lease_secs: audit_slot::DEFAULT_LEASE_SECS,
    };
    let mut answer = audit_slot::done_with(project, &slot, false, &who, None).map_err(|e| {
        e.replacen("audit-slot refused", "audit-done refused", 1).replace(
            &format!("`done {slot} --discard` reverts and removes them"),
            &format!(
                "revert them in the slot first (git -C <slot> checkout -- <path>); the leader's repair is `audit-slot done {slot} --holder {card} --discard`"
            ),
        )
    })?;
    answer["card"] = json!(card);
    Ok(answer)
}

/// One `compiler-artifact` message as a unit of the answer.
fn unit_of(v: &Value, slot: &Path, canon: &Path) -> Value {
    let workspace = v["manifest_path"]
        .as_str()
        .map(Path::new)
        .is_some_and(|m| m.starts_with(slot) || m.starts_with(canon));
    json!({
        "name": v["target"]["name"],
        "kind": v["target"]["kind"][0],
        "workspace": workspace,
        "fresh": v["fresh"] == json!(true),
        "executable": v["executable"],
    })
}

/// The last `n` lines of `text`.
fn tail(text: &str, n: usize) -> String {
    let lines: Vec<&str> = text.lines().collect();
    lines[lines.len().saturating_sub(n)..].join("\n")
}

/// Runs the one incremental build in `slot` under a heavy slot. `Err` when no heavy slot was
/// obtained, cargo did not start, or the build failed (then with the compiler's last lines).
fn build(slot: &Path, rel: &str, o: &Prep) -> Result<Value> {
    let target = slot.join("target");
    fs::create_dir_all(&target)
        .map_err(|e| format!("audit-prep: {rel}/target could not be made ({e}); no build ran"))?;
    let log_path = target.join(format!("audit-prep-{}.log", o.card));
    let heavy_log = target.join(format!("audit-prep-{}.heavy-slot.log", o.card));
    let err_file = fs::File::create(&log_path)
        .map_err(|e| format!("audit-prep: the build log could not be made ({e}); no build ran"))?;
    let mut progress: Box<dyn Write> = match fs::File::create(&heavy_log) {
        Ok(f) => Box::new(f),
        Err(_) => Box::new(std::io::sink()),
    };
    let folder = match &o.slot_dir {
        Some(d) => d.clone(),
        None => crate::heavy_slot::default_dir().map_err(|e| e.to_string())?,
    };
    let (max, limits, notes) = crate::heavy_slot::resolve(
        &folder,
        Some(o.max as usize),
        Some(u64::from(o.min_free_gb)),
        None,
        None,
    )
    .map_err(|e| format!("audit-prep: the heavy slot was refused ({e}); no cargo ran"))?;
    for n in notes {
        let _ = writeln!(progress, "{n}");
    }
    let command: Vec<String> = std::env::args_os()
        .map(|a| a.to_string_lossy().into_owned())
        .collect();
    let held = crate::heavy_slot::hold(&folder, max, "audit-prep", &command, limits, &mut progress)
        .map_err(|code| {
            format!(
                "audit-prep: no heavy slot was obtained (exit {code}, see {}); no cargo ran",
                heavy_log.display()
            )
        })?;
    let _ = writeln!(progress, "{}", held.taken_line("audit-prep"));

    let program = std::env::var_os("CARGO").unwrap_or_else(|| "cargo".into());
    let mut args: Vec<String> = [
        "test",
        "--no-run",
        "--locked",
        "--offline",
        "--message-format",
        "json",
    ]
    .map(String::from)
    .to_vec();
    if o.tests.is_empty() {
        args.push("--lib".into());
    }
    for t in o.tests {
        args.push("--test".into());
        args.push(t.clone());
    }
    let mut cmd = Command::new(&program);
    cmd.args(&args)
        .current_dir(slot)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::from(err_file))
        .env_remove("RUSTC_WRAPPER")
        .env_remove("CARGO_BUILD_RUSTC_WRAPPER")
        .env_remove("CARGO_TARGET_DIR")
        .env_remove("CARGO_BUILD_TARGET_DIR")
        .env(crate::heavy_slot::SLOT_FD_ENV, held.env_value());
    if std::env::var_os("CARGO_BUILD_JOBS").is_none() {
        cmd.env(
            "CARGO_BUILD_JOBS",
            crate::heavy_slot::build_jobs_cap(max).to_string(),
        );
    }
    let mut child = cmd.spawn().map_err(|e| {
        format!(
            "audit-prep: {} did not start ({e}); no build ran",
            program.to_string_lossy()
        )
    })?;
    let canon = fs::canonicalize(slot).unwrap_or_else(|_| slot.to_path_buf());
    let mut units: Vec<Value> = Vec::new();
    let mut errors: Vec<String> = Vec::new();
    if let Some(out) = child.stdout.take() {
        for line in BufReader::new(out)
            .lines()
            .map_while(std::result::Result::ok)
        {
            let Ok(v) = serde_json::from_str::<Value>(&line) else {
                continue;
            };
            match v["reason"].as_str() {
                Some("compiler-artifact") => units.push(unit_of(&v, slot, &canon)),
                Some("compiler-message") if v["message"]["level"] == json!("error") => {
                    if let Some(r) = v["message"]["rendered"].as_str() {
                        errors.push(r.to_string());
                    }
                }
                _ => {}
            }
        }
    }
    let status = child
        .wait()
        .map_err(|e| format!("audit-prep: waiting for cargo failed ({e})"))?;
    drop(held);
    let code = status.code().unwrap_or(-1);
    let count = |keep: &dyn Fn(&Value) -> bool| units.iter().filter(|u| keep(u)).count();
    let compiled = count(&|u| u["fresh"] == json!(false));
    let compiled_workspace =
        count(&|u| u["fresh"] == json!(false) && u["workspace"] == json!(true));
    let fresh = units.len() - compiled;
    if code != 0 {
        let stderr = fs::read_to_string(&log_path).unwrap_or_default();
        let text = format!("{}\n{stderr}", errors.join("\n"));
        return Err(format!(
            "audit-prep: cargo test --no-run failed in {rel} (exit {code}; compiled {compiled}, fresh {fresh}; log {}); last lines:\n{}",
            log_path.display(),
            tail(&text, TAIL_LINES)
        ));
    }
    Ok(json!({
        "program": program.to_string_lossy(),
        "args": args,
        "exit": code,
        "log": log_path.to_string_lossy(),
        "units": units,
        "compiled": compiled,
        "fresh": fresh,
        "compiled_workspace": compiled_workspace,
        "compiled_dependencies": compiled - compiled_workspace,
    }))
}
