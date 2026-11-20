//! SPEC-20261120-security-scan-notify: `moai hook-dispatch --event security-scan` reads the text a
//! write tool just wrote (PostToolUse, Claude Code and Codex), looks for six critical and high
//! classes and only ever notifies the model: the rule `security_scan` in `[hooks.modes]` of the
//! HOME hook-config.toml is `observe` by default (one `would_notify` line, no output), `enforce`
//! answers with `additionalContext` and `off` does nothing. Every case runs the real binary in a
//! temporary project and HOME; secret-shaped words are assembled at run time.
#![cfg(unix)]
mod common;

use common::CleanEnv;
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
use std::process::{Command, Output, Stdio};
use std::time::Instant;

struct Fx {
    _dir: tempfile::TempDir,
    root: PathBuf,
    proj: PathBuf,
    home: PathBuf,
}

fn fx() -> Fx {
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    let (proj, home) = (root.join("proj"), root.join("home"));
    fs::create_dir_all(proj.join("src")).unwrap();
    fs::create_dir_all(proj.join(".moai-cli")).unwrap();
    fs::create_dir_all(home.join(".moai-cli")).unwrap();
    Fx {
        _dir: dir,
        root,
        proj,
        home,
    }
}

/// What one run printed: (exit, stdout, stderr).
type Ran = (i32, String, String);

impl Fx {
    fn config(&self, text: &str) {
        fs::write(self.home.join(".moai-cli/hook-config.toml"), text).unwrap();
    }

    /// The rule `security_scan` in the mode `mode` (a bare word of the three modes, or any word).
    fn mode(&self, mode: &str) {
        self.config(&format!("[hooks.modes]\nsecurity_scan = \"{mode}\"\n"));
    }

    fn dispatch(&self, host: &str, event: &str, stdin: &[u8]) -> Ran {
        let mut child = Command::new(common::MOAI_BIN)
            .clean_env()
            .env("PATH", std::env::var_os("PATH").unwrap_or_default())
            .env("HOME", &self.home)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .current_dir(&self.proj)
            .args(["hook-dispatch", "--host", host, "--event", event])
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        common::feed_stdin(&mut child, stdin);
        let out: Output = child.wait_with_output().unwrap();
        (
            out.status.code().unwrap_or(-1),
            String::from_utf8_lossy(&out.stdout).trim().to_string(),
            String::from_utf8_lossy(&out.stderr).trim().to_string(),
        )
    }

    fn scan(&self, host: &str, stdin: &[u8]) -> Ran {
        self.dispatch(host, "security-scan", stdin)
    }

    /// The engine log lines of the project.
    fn log(&self) -> Vec<Value> {
        fs::read_to_string(self.proj.join(".moai-cli/receipts/hook-engine.jsonl"))
            .unwrap_or_default()
            .lines()
            .filter_map(|l| serde_json::from_str(l).ok())
            .collect()
    }

    fn would_notify(&self) -> Vec<Value> {
        self.log()
            .into_iter()
            .filter(|l| l["kind"] == json!("would_notify"))
            .collect()
    }

    fn abs(&self, rel: &str) -> String {
        self.proj.join(rel).display().to_string()
    }

    /// A Claude Code PostToolUse input for a write tool.
    fn claude(&self, tool: &str, tool_input: Value) -> Vec<u8> {
        json!({"session_id": "s1", "hook_event_name": "PostToolUse", "tool_name": tool,
               "tool_use_id": "t1", "tool_input": tool_input,
               "tool_response": {"success": true}, "cwd": self.proj})
        .to_string()
        .into_bytes()
    }

    fn write(&self, rel: &str, content: &str) -> Vec<u8> {
        self.claude(
            "Write",
            json!({"file_path": self.abs(rel), "content": content}),
        )
    }

    /// A Codex PostToolUse input for apply_patch (only `tool_input.command` carries the patch).
    fn patch(&self, patch: &str) -> Vec<u8> {
        json!({"session_id": "s1", "hook_event_name": "PostToolUse", "tool_name": "apply_patch",
               "tool_use_id": "t1", "tool_input": {"command": patch},
               "tool_response": "Success", "cwd": self.proj})
        .to_string()
        .into_bytes()
    }
}

/// The classes of the table: (class, severity, one line that holds only that class).
fn samples() -> Vec<(&'static str, &'static str, String)> {
    let name = format!("{}{}", "pass", "word");
    vec![
        (
            "unsafe-deserialization",
            "high",
            "data = pickle.loads(blob)".to_string(),
        ),
        (
            "dom-injection-xss",
            "high",
            "el.innerHTML = html;".to_string(),
        ),
        (
            "hardcoded-secret",
            "critical",
            format!("{name} = \"hunter2\""),
        ),
        (
            "code-injection-eval",
            "high",
            "result = eval(user_text)".to_string(),
        ),
        (
            "sql-injection",
            "high",
            "q = \"SELECT * FROM users WHERE id = \" + uid".to_string(),
        ),
        ("command-injection", "high", "os.system(cmd)".to_string()),
    ]
}

/// The `additionalContext` of a PostToolUse answer (and the shape around it).
fn context(stdout: &str) -> String {
    let v: Value = serde_json::from_str(stdout).unwrap_or_else(|e| panic!("{e}: {stdout}"));
    assert_eq!(
        v["hookSpecificOutput"]["hookEventName"], "PostToolUse",
        "{v}"
    );
    assert_eq!(
        v.as_object().unwrap().len(),
        1,
        "only hookSpecificOutput: {v}"
    );
    assert_eq!(
        v["hookSpecificOutput"].as_object().unwrap().len(),
        2,
        "only the event name and the context: {v}"
    );
    v["hookSpecificOutput"]["additionalContext"]
        .as_str()
        .unwrap_or_else(|| panic!("no additionalContext: {v}"))
        .to_string()
}

/// The finding entries of a notification: lines `- <class> (<severity>) <path>:<line>`.
fn entries(text: &str) -> Vec<String> {
    text.lines()
        .filter(|l| l.starts_with("- ") && !l.starts_with("- and "))
        .map(str::to_string)
        .collect()
}

/// The classes a Write of `text` to src/a.py is notified for (enforce), sorted.
fn classes_of(f: &Fx, text: &str) -> Vec<String> {
    let (code, out, err) = f.scan("claude", &f.write("src/a.py", text));
    assert_eq!((code, err.as_str()), (0, ""), "{text:?}");
    if out.is_empty() {
        return Vec::new();
    }
    let mut classes: Vec<String> = entries(&context(&out))
        .iter()
        .map(|e| e[2..].split(' ').next().unwrap().to_string())
        .collect();
    classes.sort();
    classes
}

/// AC-1, REQ-1 / REQ-7: in enforce mode every class is notified through every Claude write tool,
/// as the documented PostToolUse `additionalContext` object, with class, severity, relative path
/// and line, no matched text, within 800 bytes.
#[test]
fn notify_every_class_on_every_claude_write_tool() {
    let f = fx();
    f.mode("enforce");
    for (class, severity, line) in samples() {
        let inputs = [
            (
                "src/a.py",
                2,
                f.write("src/a.py", &format!("# header\n{line}\n")),
            ),
            (
                "src/a.py",
                1,
                f.claude(
                    "Edit",
                    json!({"file_path": f.abs("src/a.py"), "old_string": "x", "new_string": line}),
                ),
            ),
            (
                "src/a.py",
                2,
                f.claude(
                    "MultiEdit",
                    json!({"file_path": f.abs("src/a.py"), "edits": [
                        {"old_string": "a", "new_string": "ok"},
                        {"old_string": "b", "new_string": line}]}),
                ),
            ),
            (
                "nb.ipynb",
                1,
                f.claude(
                    "NotebookEdit",
                    json!({"notebook_path": f.abs("nb.ipynb"), "new_source": line}),
                ),
            ),
        ];
        for (path, at, input) in inputs {
            let (code, out, err) = f.scan("claude", &input);
            assert_eq!((code, err.as_str()), (0, ""), "{class}");
            let text = context(&out);
            assert_eq!(
                entries(&text),
                vec![format!("- {class} ({severity}) {path}:{at}")],
                "{class}: {text}"
            );
            assert!(text.len() <= 800, "{class}: {} bytes", text.len());
            assert!(!text.contains(&line), "{class}: the matched text is echoed");
            assert!(
                text.contains("advisory") && text.contains("nothing was blocked"),
                "{class}: {text}"
            );
        }
    }
}

/// AC-1, REQ-1: Codex apply_patch is notified for added lines only, with the path of the file
/// the line was added to and the line number among that file's added lines.
#[test]
fn notify_codex_apply_patch_added_lines_only() {
    let f = fx();
    f.mode("enforce");
    for (class, severity, line) in samples() {
        let patch = format!(
            "*** Begin Patch\n*** Add File: src/b.py\n+# header\n+{line}\n\
             *** Update File: src/c.py\n@@\n-{line}\n context {line}\n+ok\n*** End Patch"
        );
        let (code, out, err) = f.scan("codex", &f.patch(&patch));
        assert_eq!((code, err.as_str()), (0, ""), "{class}");
        let text = context(&out);
        assert_eq!(
            entries(&text),
            vec![format!("- {class} ({severity}) src/b.py:2")],
            "{class}: removed and context lines must not count: {text}"
        );
    }
}

/// REQ-7: at most 5 findings and 800 bytes; the rest is counted, never listed; a long path does
/// not push the text past the bound; odd path characters are not echoed.
#[test]
fn notify_lists_at_most_five_findings_within_800_bytes() {
    let f = fx();
    f.mode("enforce");
    let text: String = (1..=7).map(|n| format!("eval(v{n})\n")).collect();
    let (_, out, _) = f.scan("claude", &f.write("src/a.py", &text));
    let note = context(&out);
    assert_eq!(entries(&note).len(), 5, "{note}");
    assert!(note.contains("- and 2 more"), "{note}");
    assert!(note.len() <= 800, "{} bytes", note.len());

    let long = format!("src/{}/a.py", "d".repeat(190));
    let (_, out, _) = f.scan("claude", &f.write(&long, &text));
    let note = context(&out);
    assert!(note.len() <= 800, "{} bytes: {note}", note.len());
    let listed = entries(&note).len();
    assert!(listed >= 1, "{note}");
    let more: usize = note
        .lines()
        .find_map(|l| l.strip_prefix("- and ")?.split(' ').next()?.parse().ok())
        .unwrap_or(0);
    assert_eq!(
        listed + more,
        7,
        "every finding is listed or counted: {note}"
    );
}

#[test]
fn notify_never_echoes_odd_path_characters() {
    let f = fx();
    f.mode("enforce");
    let input = f.claude(
        "Write",
        json!({"file_path": f.abs("src/a\n- fake (critical) x:1.py"), "content": "eval(x)"}),
    );
    let (code, out, _) = f.scan("claude", &input);
    assert_eq!(code, 0);
    let note = context(&out);
    assert!(!note.contains("fake"), "{note}");
    assert_eq!(entries(&note).len(), 1, "{note}");
}

/// AC-3: a class is reported once per line; two classes on one line are two findings; line
/// numbers count from 1.
#[test]
fn shapes_report_a_class_once_per_line() {
    let f = fx();
    f.mode("enforce");
    let (_, out, _) = f.scan(
        "claude",
        &f.write("src/a.py", "ok\neval(a); eval(b)\neval(os.system(c))\n"),
    );
    assert_eq!(
        entries(&context(&out)),
        vec![
            "- code-injection-eval (high) src/a.py:2",
            "- code-injection-eval (high) src/a.py:3",
            "- command-injection (high) src/a.py:3",
        ]
    );
}

/// AC-3: each shape is found at its edge and not one character beyond it.
#[test]
fn shapes_match_their_table_row() {
    let f = fx();
    f.mode("enforce");
    let name = format!("{}{}", "pass", "word");
    let sql_quote = "`DELETE FROM t ` + x".to_string();
    let positive: Vec<(&str, String)> = vec![
        ("unsafe-deserialization", "yaml.load(x)".into()),
        ("unsafe-deserialization", "yaml.load (x)".into()),
        ("unsafe-deserialization", "pickle.load(f)".into()),
        ("unsafe-deserialization", "pickle.loads(b)".into()),
        ("unsafe-deserialization", "torch.load(p)".into()),
        ("unsafe-deserialization", "Marshal.load(d)".into()),
        ("unsafe-deserialization", "new ObjectInputStream(s)".into()),
        (
            "unsafe-deserialization",
            "new  ObjectInputStream (s)".into(),
        ),
        ("unsafe-deserialization", "unserialize($x)".into()),
        ("unsafe-deserialization", "a;unserialize ($x)".into()),
        ("unsafe-deserialization", "$o.unserialize($x)".into()),
        ("dom-injection-xss", "el.innerHTML=a".into()),
        ("dom-injection-xss", "a.innerHTML  =  b".into()),
        (
            "dom-injection-xss",
            "<p dangerouslySetInnerHTML={x}>".into(),
        ),
        ("dom-injection-xss", "document.write(x)".into()),
        ("dom-injection-xss", "document.write (x)".into()),
        ("dom-injection-xss", "<div v-html=\"x\">".into()),
        ("dom-injection-xss", "<div v-html = \"x\">".into()),
        ("hardcoded-secret", format!("{name}:=\"x\"")),
        ("hardcoded-secret", format!("{name} <- 'x'")),
        ("hardcoded-secret", format!("{}{}", "API", "_KEY = \"abc\"")),
        ("hardcoded-secret", format!("{}{}", "api", "Key = 'abc'")),
        ("hardcoded-secret", format!("{}{}", "api", "-key = \"abc\"")),
        ("hardcoded-secret", format!("Pass{}", "Word = \"abc\"")),
        ("hardcoded-secret", format!("my_{name} = \"x\"")),
        ("code-injection-eval", "eval(x)".into()),
        ("code-injection-eval", "eval (x)".into()),
        ("code-injection-eval", "x.eval(y)".into()),
        ("code-injection-eval", "(eval(x))".into()),
        ("code-injection-eval", "new Function(\"a\")".into()),
        ("code-injection-eval", "new  Function (\"a\")".into()),
        ("sql-injection", "\"SELECT * FROM t\" + x".into()),
        ("sql-injection", "' select a from b ' + c".into()),
        ("sql-injection", sql_quote),
        ("sql-injection", "\"drop table x\"+y".into()),
        ("sql-injection", "query(\"a\" + b)".into()),
        ("sql-injection", "execute(x + y)".into()),
        ("sql-injection", "cursor.execute(sql + tail)".into()),
        ("sql-injection", "Query (a + b)".into()),
        ("command-injection", "shell=True".into()),
        ("command-injection", "shell = True".into()),
        ("command-injection", "os.system(c)".into()),
        ("command-injection", "os.system (c)".into()),
        ("command-injection", "child_process.exec(c)".into()),
        ("command-injection", "run([\"sh -c\", x])".into()),
        ("command-injection", "'bash  -c ls'".into()),
        ("command-injection", "`bash -c x`".into()),
    ];
    for (class, text) in &positive {
        assert_eq!(classes_of(&f, text), vec![class.to_string()], "{text:?}");
    }
    let negative: Vec<String> = vec![
        // prose, near misses and comparisons
        "yaml.safe_load(x)".into(),
        "yaml.load_all(x)".into(),
        "yaml.load".into(),
        "pickle.dumps(x)".into(),
        "pickle.loader(x)".into(),
        "myunserialize(x)".into(),
        "unserialize_x(y)".into(),
        "ObjectInputStream(s)".into(),
        "a.innerHTML == b".into(),
        "a.innerHTML === b".into(),
        "a.innerHTML += b".into(),
        "a.innerHTML;".into(),
        "innerHTML = x".into(),
        "document.writeln(x)".into(),
        "<div v-html>".into(),
        format!("{name} = \"\""),
        format!("{name} == \"x\""),
        format!("{name} = os.getenv(\"X\")"),
        format!("{name} = x"),
        format!("{name}: \"x\""),
        format!("{name}_hash = \"x\""),
        format!("{}{}", "api", "key = `abc`"),
        format!("{}{}", "api", " key = \"x\""),
        "evaluate(x)".into(),
        "retrieval(x)".into(),
        "medieval(x)".into(),
        "my_eval(x)".into(),
        "eval = 1".into(),
        "the eval function".into(),
        "eval".into(),
        "new Function".into(),
        "\"selection\" + x".into(),
        "\"SELECT a FROM b\"".into(),
        "\"SELECT a\" - x".into(),
        "query(\"a\")".into(),
        "execute(a, b)".into(),
        "query(a) + b".into(),
        "query(a + b".into(),
        // Go parity: a quote inside the string hides the first shape
        "\"SELECT * FROM t WHERE n = '\" + n".into(),
        "shell=False".into(),
        "shell = true".into(),
        "os.systemd(c)".into(),
        "os.system".into(),
        "child_process.execSync(c)".into(),
        "child_process.execFile(c)".into(),
        "\"ssh -c x\"".into(),
        "sh -c x".into(),
        "\"sh -cx\"".into(),
        "\"sh -x\"".into(),
        "\"bash\" -c".into(),
    ];
    for text in &negative {
        assert_eq!(classes_of(&f, text), Vec::<String>::new(), "{text:?}");
    }
}

/// AC-3, REQ-4: the look-ahead is 512 bytes beyond the fixed lead-in; one byte more is no match.
#[test]
fn shapes_look_ahead_edge_is_512_bytes() {
    let f = fx();
    f.mode("enforce");
    let name = format!("{}{}", "pass", "word");
    // The whole tail of a shape lies in the 512 bytes after its lead-in. Here `)` is byte 511
    // after `(` (the 512th) → found; the 513th → not.
    let paren = |inner: usize| format!("query({}+)", "a".repeat(inner));
    assert_eq!(classes_of(&f, &paren(510)), vec!["sql-injection"]);
    assert_eq!(classes_of(&f, &paren(511)), Vec::<String>::new());
    // The name is the lead-in: blank, `=`, blank and the quote take 4 bytes, so the closing quote
    // of a 507-byte value is byte 511 → found; one byte more → not.
    let secret = |len: usize| format!("{name} = \"{}\"", "a".repeat(len));
    assert_eq!(classes_of(&f, &secret(507)), vec!["hardcoded-secret"]);
    assert_eq!(classes_of(&f, &secret(508)), Vec::<String>::new());
    // The whole tail of the quoted shape (text, closing quote, blank, `+`) is the 512 bytes after
    // the keyword: with one blank before the text and one before the `+`, the `+` is byte 511
    // when the text is 508 bytes long.
    let sql = |gap: usize| format!("\"SELECT {}\" + x", "b".repeat(gap));
    assert_eq!(classes_of(&f, &sql(508)), vec!["sql-injection"]);
    assert_eq!(classes_of(&f, &sql(509)), Vec::<String>::new());
    // Blanks between a name and its `(` count too.
    let blanks = |n: usize| format!("eval{}(x)", " ".repeat(n));
    assert_eq!(classes_of(&f, &blanks(511)), vec!["code-injection-eval"]);
    assert_eq!(classes_of(&f, &blanks(512)), Vec::<String>::new());
}

/// AC-2, REQ-2 / REQ-3: no entry, an unknown word and `observe` print nothing and write one
/// `would_notify` line without path, command or text; `off` writes nothing; no finding writes no
/// line in any mode.
#[test]
fn mode_observe_writes_one_would_notify_line_and_prints_nothing() {
    for config in [
        None,
        Some("[hooks.modes]\nsecurity_scan = \"observe\"\n"),
        Some("[hooks.modes]\nsecurity_scan = \"yes\"\n"),
        Some("[hooks.modes]\nsecurity_scan = \"Enforce\"\n"),
        Some("[hooks.modes]\nmerge_confirm = \"enforce\"\n"),
    ] {
        let f = fx();
        if let Some(text) = config {
            f.config(text);
        }
        let input = f.write("src/a.py", "ok\nresult = eval(user_text)\nos.system(c)\n");
        let (code, out, err) = f.scan("claude", &input);
        assert_eq!(
            (code, out.as_str(), err.as_str()),
            (0, "", ""),
            "{config:?}"
        );
        let lines = f.would_notify();
        assert_eq!(lines.len(), 1, "{config:?}: {:?}", f.log());
        let l = &lines[0];
        assert_eq!(l["rule"], json!("security_scan"));
        assert_eq!(
            l["classes"],
            json!(["code-injection-eval", "command-injection"])
        );
        assert_eq!(l["findings"], json!(2));
        assert_eq!(l["host"], json!("claude"));
        assert!(l["session"].is_string() && l["call"].is_string());
        assert!(l["unix_ms"].as_u64().unwrap() > 0);
        let raw = l.to_string();
        for leak in ["a.py", "result =", "os.system", "user_text", "src/"] {
            assert!(!raw.contains(leak), "{leak} leaked into {raw}");
        }
        assert!(
            f.log().iter().all(|x| x["kind"] != json!("would_block")),
            "a notify rule is not a block rule"
        );
    }
}

#[test]
fn mode_off_prints_nothing_and_writes_nothing() {
    let f = fx();
    f.mode("off");
    let (code, out, err) = f.scan("claude", &f.write("src/a.py", "eval(x)\n"));
    assert_eq!((code, out.as_str(), err.as_str()), (0, "", ""));
    assert!(f.log().is_empty(), "{:?}", f.log());
}

#[test]
fn mode_without_a_finding_writes_no_line_in_any_mode() {
    for mode in ["off", "observe", "enforce"] {
        let f = fx();
        f.mode(mode);
        let (code, out, _) = f.scan("claude", &f.write("src/a.py", "print('hello')\n"));
        assert_eq!((code, out.as_str()), (0, ""), "{mode}");
        assert!(f.log().is_empty(), "{mode}: {:?}", f.log());
    }
}

#[test]
fn mode_enforce_writes_no_would_notify_line_and_never_blocks() {
    let f = fx();
    f.mode("enforce");
    let (code, out, _) = f.scan("claude", &f.write("src/a.py", "eval(x)\n"));
    assert_eq!(code, 0);
    let v: Value = serde_json::from_str(&out).unwrap();
    for key in ["decision", "permissionDecision", "continue", "stopReason"] {
        assert!(
            v.get(key).is_none() && v["hookSpecificOutput"].get(key).is_none(),
            "{key}"
        );
    }
    assert!(f.would_notify().is_empty());
    // Codex gets the same shape from the same duty.
    let patch = "*** Begin Patch\n*** Add File: src/b.py\n+eval(x)\n*** End Patch";
    let (code, out, _) = f.scan("codex", &f.patch(patch));
    assert_eq!(code, 0);
    assert!(context(&out).contains("code-injection-eval"));
}

/// AC-4, REQ-5: a NUL character, a text one byte over 1 MiB and targets outside the project give
/// no output and no line, while exactly 1 MiB is scanned.
#[test]
fn skips_binary_oversized_and_outside_targets() {
    let f = fx();
    f.mode("enforce");
    let quiet = |input: Vec<u8>, what: &str| {
        let (code, out, err) = f.scan("claude", &input);
        assert_eq!((code, out.as_str(), err.as_str()), (0, "", ""), "{what}");
    };
    quiet(f.write("src/a.py", "eval(x)\n\0"), "NUL");
    let max = 1024 * 1024;
    let text = |len: usize| format!("eval(x)\n{}", "a".repeat(len - "eval(x)\n".len()));
    quiet(f.write("src/a.py", &text(max + 1)), "1 MiB + 1");
    let (code, out, _) = f.scan("claude", &f.write("src/a.py", &text(max)));
    assert_eq!(code, 0);
    assert!(
        context(&out).contains("code-injection-eval"),
        "exactly 1 MiB is scanned"
    );
    let elsewhere = f.root.join("elsewhere/a.py").display().to_string();
    for (path, what) in [
        (elsewhere.as_str(), "absolute, outside"),
        ("../a.py", "relative, up"),
        ("sub/../../a.py", "relative, up through a child"),
    ] {
        quiet(
            f.claude("Write", json!({"file_path": path, "content": "eval(x)"})),
            what,
        );
    }
    // `./src/a.py` is inside and is listed without the `./`.
    let (_, out, _) = f.scan(
        "claude",
        &f.claude(
            "Write",
            json!({"file_path": "./src/a.py", "content": "eval(x)"}),
        ),
    );
    assert_eq!(
        entries(&context(&out)),
        vec!["- code-injection-eval (high) src/a.py:1"]
    );
    // A path that reaches the project through a link (as `/var` does for `/private/var`) is
    // judged by where the file really is: inside → listed with its project-relative path, a link
    // to a folder outside → quiet.
    fs::write(f.proj.join("src/a.py"), "x").unwrap();
    std::os::unix::fs::symlink(&f.proj, f.root.join("alias")).unwrap();
    let through = f.root.join("alias/src/a.py").display().to_string();
    let (_, out, _) = f.scan(
        "claude",
        &f.claude("Write", json!({"file_path": through, "content": "eval(x)"})),
    );
    assert_eq!(
        entries(&context(&out)),
        vec!["- code-injection-eval (high) src/a.py:1"]
    );
    fs::create_dir_all(f.root.join("elsewhere")).unwrap();
    fs::write(f.root.join("elsewhere/b.py"), "x").unwrap();
    std::os::unix::fs::symlink(f.root.join("elsewhere"), f.root.join("outlink")).unwrap();
    let through = f.root.join("outlink/b.py").display().to_string();
    quiet(
        f.claude("Write", json!({"file_path": through, "content": "eval(x)"})),
        "link to a folder outside",
    );
    // A patch that touches one file outside and one inside lists the inside one only.
    let patch = "*** Begin Patch\n*** Add File: ../out.py\n+eval(x)\n\
                 *** Add File: src/in.py\n+os.system(c)\n*** End Patch";
    let (_, out, _) = f.scan("codex", &f.patch(patch));
    assert_eq!(
        entries(&context(&out)),
        vec!["- command-injection (high) src/in.py:1"]
    );
    // No text, or a tool that writes no file, is nothing to scan.
    quiet(
        f.claude("Write", json!({"file_path": f.abs("src/a.py")})),
        "no content",
    );
    quiet(
        f.claude("Bash", json!({"command": "eval(x)"})),
        "Bash is not a write tool",
    );
    assert!(f.log().is_empty(), "{:?}", f.log());
}

/// AC-5, REQ-8 / REQ-9: malformed and cut input are quiet skips; OMP and another event name are
/// engine errors.
#[test]
fn contract_malformed_input_is_a_quiet_skip_and_the_wrong_duty_an_error() {
    let f = fx();
    f.mode("enforce");
    let whole = f.write("src/a.py", "eval(x)\n");
    let cut = &whole[..whole.len() / 2];
    for (host, input) in [
        ("claude", b"{oops".to_vec()),
        ("codex", b"".to_vec()),
        ("claude", cut.to_vec()),
        ("codex", b"[1,2]".to_vec()),
    ] {
        let (code, out, err) = f.scan(host, &input);
        assert_eq!((code, out.as_str(), err.as_str()), (0, "", ""), "{host}");
    }
    let (code, out, err) = f.scan("omp", &whole);
    assert_eq!((code, out.as_str()), (1, ""));
    assert!(err.contains("engine error"), "{err}");
    for name in ["PreToolUse", "PostToolUseFailure", "Stop"] {
        let input = json!({"hook_event_name": name, "tool_name": "Write",
            "tool_input": {"file_path": f.abs("src/a.py"), "content": "eval(x)"}})
        .to_string();
        let (code, out, err) = f.scan("claude", input.as_bytes());
        assert_eq!((code, out.as_str()), (1, ""), "{name}");
        assert!(err.contains("engine error"), "{name}: {err}");
    }
    // A missing event name is accepted, like the other PostToolUse duties.
    let input = json!({"tool_name": "Write", "tool_input": {"file_path": f.abs("src/a.py"),
        "content": "eval(x)"}})
    .to_string();
    let (code, out, _) = f.scan("claude", input.as_bytes());
    assert_eq!(code, 0);
    assert!(context(&out).contains("code-injection-eval"));
}

/// AC-5, REQ-10: the new event is not wired into any generated setting, and the other duties
/// answer the same write exactly as before.
#[test]
fn contract_existing_duties_and_generated_files_are_unchanged() {
    let f = fx();
    f.mode("enforce");
    let tool_input = json!({"file_path": f.abs("src/a.py"), "content": "eval(x)\nos.system(c)\n"});
    for (event, host_event) in [
        ("tool-preflight", "PreToolUse"),
        ("tool-observe", "PostToolUse"),
    ] {
        let input = json!({"session_id": "s1", "hook_event_name": host_event,
            "tool_name": "Write", "tool_use_id": "t1", "tool_input": tool_input,
            "tool_response": {"success": true}, "cwd": f.proj})
        .to_string();
        let (code, out, err) = f.dispatch("claude", event, input.as_bytes());
        assert_eq!((code, out.as_str(), err.as_str()), (0, "", ""), "{event}");
    }
    for host in ["claude", "codex"] {
        let out = Command::new(common::MOAI_BIN)
            .clean_env()
            .env("PATH", std::env::var_os("PATH").unwrap_or_default())
            .env("HOME", &f.home)
            .current_dir(&f.proj)
            .args(["hooks-plan", f.proj.to_str().unwrap(), "--host", host])
            .output()
            .unwrap();
        assert!(out.status.success(), "{host}: {out:?}");
        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
        let hooks = v["result"]["hooks"].to_string();
        assert!(!hooks.contains("security-scan"), "{host}: {hooks}");
    }
}

/// AC-6, REQ-11: `hooks-plan` reports the new duty next to the plan.
#[test]
fn contract_capabilities_list_security_scan() {
    let f = fx();
    for (host, tools) in [
        ("claude", vec!["Write", "Edit", "MultiEdit", "NotebookEdit"]),
        ("codex", vec!["apply_patch"]),
    ] {
        let out = Command::new(common::MOAI_BIN)
            .clean_env()
            .env("PATH", std::env::var_os("PATH").unwrap_or_default())
            .env("HOME", &f.home)
            .current_dir(&f.proj)
            .args(["hooks-plan", f.proj.to_str().unwrap(), "--host", host])
            .output()
            .unwrap();
        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
        let cap = &v["result"]["engine_capabilities"]["security-scan"];
        assert_eq!(cap["event"], json!("PostToolUse"), "{host}: {cap}");
        assert_eq!(cap["blocks"], json!(false), "{host}");
        assert_eq!(cap["mode_key"], json!("security_scan"), "{host}");
        assert_eq!(cap["default_mode"], json!("observe"), "{host}");
        assert_eq!(
            cap["answer"],
            json!("hookSpecificOutput.additionalContext"),
            "{host}"
        );
        let listed: Vec<&str> = cap["tools"]
            .as_array()
            .unwrap_or_else(|| panic!("{host}: {cap}"))
            .iter()
            .filter_map(Value::as_str)
            .collect();
        assert_eq!(listed, tools, "{host}");
    }
    let out = Command::new(common::MOAI_BIN)
        .clean_env()
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("HOME", &f.home)
        .current_dir(&f.proj)
        .args(["hooks-plan", f.proj.to_str().unwrap(), "--host", "omp"])
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(
        v["result"]["engine_capabilities"]["security-scan"]["supported"],
        json!(false)
    );
}

/// AC-7, REQ-4: near-miss text of 1 MiB on one line finishes with the right verdict; the time is
/// printed, the bound is generous (a quadratic scan would not meet it).
#[test]
fn linear_scan_time_on_one_mib_of_near_misses() {
    let f = fx();
    f.mode("enforce");
    let name = format!("{}{}", "pass", "word");
    let families: Vec<(String, Vec<&str>)> = vec![
        ("query(a+".into(), vec!["code-injection-eval"]),
        (
            format!("{name} = \""),
            vec!["code-injection-eval", "hardcoded-secret"],
        ),
        ("\"SELECT a ".into(), vec!["code-injection-eval"]),
        ("eval          ".into(), vec!["code-injection-eval"]),
        ("shell =       ".into(), vec!["code-injection-eval"]),
    ];
    let mib = 1024 * 1024;
    for (unit, expected) in families {
        // Just under the 1 MiB bound, so the closing `eval(z)` line is still scanned.
        let body = unit.repeat((mib - 64) / unit.len());
        let text = format!("{body}\neval(z)\n");
        let start = Instant::now();
        let got = classes_of(&f, &text);
        let took = start.elapsed();
        eprintln!("linear {unit:?}: {took:?}");
        assert_eq!(got, expected, "{unit:?}");
        assert!(took.as_secs() < 8, "{unit:?}: {took:?}");
    }
}
