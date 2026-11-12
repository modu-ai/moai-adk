//! #2497: the MODE-AGENT and MODE-LEADER workflows as testable step lists, and every
//! `moai-cli <op> [--flag]` they name checked against the built binary's `--help`.
use std::{fs, path::Path, process::Command};

fn workflow(name: &str) -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai/workflows")
            .join(name),
    )
    .unwrap()
}

/// The fixed keys of the `## Procedure` steps (`N. **key** — …`), in order; the step numbers
/// must run 1, 2, 3, … without a gap.
fn procedure_keys(text: &str) -> Vec<String> {
    let start = text.find("\n## Procedure\n").expect("a Procedure section");
    let body = &text[start + 1..];
    let body = &body[..body[3..].find("\n## ").map_or(body.len(), |i| i + 3)];
    let mut keys = Vec::new();
    for line in body.lines() {
        let Some((num, rest)) = line.split_once(". ") else {
            continue;
        };
        if num.is_empty() || !num.bytes().all(|b| b.is_ascii_digit()) {
            continue;
        }
        assert_eq!(
            num,
            (keys.len() + 1).to_string(),
            "step numbers run without a gap: {line}"
        );
        let key = rest
            .strip_prefix("**")
            .and_then(|r| r.split_once("** — "))
            .map(|(k, _)| k)
            .unwrap_or_else(|| panic!("step {num} starts with its **key** — : {line}"));
        assert!(
            !key.is_empty() && key.bytes().all(|b| b.is_ascii_lowercase() || b == b'-'),
            "{key}"
        );
        keys.push(key.to_string());
    }
    keys
}

#[test]
fn mode_agent_first_run_steps_are_fixed() {
    assert_eq!(
        procedure_keys(&workflow("mode-agent.md")),
        [
            "mode-only",
            "banner",
            "state",
            "session-id",
            "claim",
            "claim-verify",
            "heavy-checks",
            "pr-gate",
            "merge-train",
            "hand-off",
            "goal-lines",
            "self-pick",
            "report",
        ]
    );
}

#[test]
fn mode_leader_first_run_steps_are_fixed() {
    assert_eq!(
        procedure_keys(&workflow("mode-leader.md")),
        [
            "codex-refusal",
            "mode-only",
            "banner",
            "state",
            "answer-requests",
            "goal-line",
            "supervise",
            "pr-gate",
            "merge-train",
            "handoff-line",
            "lane-done",
        ]
    );
}

/// Each `` `moai-cli <op> [--flag …]` `` span in a workflow: (op, flags).
fn cli_spans(text: &str) -> Vec<(String, Vec<String>)> {
    let mut out = Vec::new();
    for (i, _) in text.match_indices("`moai-cli ") {
        let rest = &text[i + 1..];
        let span = &rest[..rest.find('`').expect("closed code span")];
        let mut words = span.split_whitespace().skip(1);
        let op = words.next().expect("an op after moai-cli").to_string();
        let flags = words
            .take_while(|w| w.starts_with("--"))
            .map(str::to_string)
            .collect();
        out.push((op, flags));
    }
    out
}

fn help(args: &[&str]) -> String {
    let out = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
        .arg("--help")
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "`moai-cli {} --help` failed: {}",
        args.join(" "),
        String::from_utf8_lossy(&out.stderr)
    );
    String::from_utf8_lossy(&out.stdout).into_owned()
}

/// AC-3: every op a mode workflow names exists in this binary, with every flag it shows.
#[test]
fn ops_named_by_the_mode_workflows_exist_with_their_flags() {
    let top = help(&[]);
    let mut seen = Vec::new();
    for name in ["mode-agent.md", "mode-leader.md"] {
        for (op, flags) in cli_spans(&workflow(name)) {
            assert!(
                top.lines()
                    .any(|l| l.split_whitespace().next() == Some(op.as_str())),
                "{name}: `moai-cli {op}` is not a subcommand of this binary"
            );
            let sub = help(&[op.as_str()]);
            for flag in &flags {
                assert!(
                    sub.contains(flag.as_str()),
                    "{name}: `moai-cli {op}` has no {flag}"
                );
            }
            seen.push(format!(
                "{op}{}",
                flags.iter().map(|f| format!(" {f}")).collect::<String>()
            ));
        }
    }
    seen.sort();
    seen.dedup();
    assert_eq!(
        seen,
        [
            "claim-verify",
            "claim-verify --hook",
            "pr-confirm",
            "queue-next",
            "scope-check"
        ],
        "the op list the workflows name"
    );
}

#[test]
fn spans_parse_ops_and_flags() {
    assert_eq!(
        cli_spans("run `moai-cli claim-verify --hook` and `moai-cli queue-next`"),
        [
            ("claim-verify".to_string(), vec!["--hook".to_string()]),
            ("queue-next".to_string(), vec![])
        ]
    );
}
