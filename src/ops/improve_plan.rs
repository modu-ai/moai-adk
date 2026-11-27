//! `improve-plan` (SPEC-20261121-self-improvement-loop, slice S2): a read-only proposal of one to
//! three "lessons" for one project skill or agent file, held by static gates. The logic is in
//! this file with its unit tests; the command line tests are `tests/improve_plan_cli.rs`.

use moai_cli::{help, redact, state_root};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::collections::{BTreeMap, BTreeSet};
use std::io::Read;
use std::path::{Component, Path, PathBuf};

/// Plan a change to a project skill or agent file without writing anything
///
/// Reads one target file below the project folder (a skill or agent the project's generation
/// manifest owns, grade A) and shows what one to three lesson items would do to it: the old and
/// new sha-256, the item lines, whether the file still equals its generated baseline, and a
/// `plan_digest`. Items are merged by id inside the managed block `moai-lessons`. A proposal that
/// fails a gate is refused with every failing reason (`not_allowed`, `test_weakening`, `items`,
/// `drift`, `secret`, `front_matter`, `budget`). Writes nothing and says `writes: false`.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder.
    #[arg(default_value = ".")]
    project: PathBuf,
    /// The file to change, a path below the project folder (for example `.claude/agents/moai/prj-x.md`).
    #[arg(long)]
    target: String,
    /// One item as JSON (repeat for one to three): `{"op":"add|replace","id":"…","rule":"…","reason":"…","source":["…"]}` or `{"op":"retire","id":"…"}`.
    #[arg(long = "item")]
    items: Vec<String>,
}

pub const OPERATION: &str = "improve.plan";

/// The bytes of the target and of the manifest that are read.
const FILE_BYTES: u64 = 1024 * 1024;
const MANIFEST_BYTES: u64 = 4 * 1024 * 1024;
/// Size budgets (SPEC AC-12): lines of a skill, lines of an agent file, characters of a
/// description.
const SKILL_LINES: usize = 500;
const AGENT_LINES: usize = 200;
const DESCRIPTION_CHARS: usize = 1024;
/// Item limits (SPEC AC-11).
const MAX_ITEMS: usize = 3;
const MAX_SOURCES: usize = 5;
const TEXT_CHARS: usize = 200;
const SOURCE_CHARS: usize = 120;
const ID_CHARS: usize = 40;
/// The managed block of the lessons (`src/agents.rs` writes the same marker form).
const BEGIN: &str = "<!-- moai:managed:begin id=moai-lessons -->";
const END: &str = "<!-- moai:managed:end id=moai-lessons -->";
/// The manifest sections a grade A target may carry (host `claude` only, SPEC D10).
const SECTION_SKILL: &str = "generation:project-skill:claude:";
const SECTION_AGENT: &str = "generation:native-role:claude:";
/// File names a cited path may end in (the same set `improve-signals` reads).
const EXTENSIONS: [&str; 9] = ["rs", "md", "py", "json", "yml", "yaml", "toml", "sh", "txt"];

const NOT_ALLOWED: &str = "not_allowed";
const TEST_WEAKENING: &str = "test_weakening";
const ITEMS: &str = "items";
const DRIFT: &str = "drift";
const SECRET: &str = "secret";
const FRONT_MATTER: &str = "front_matter";
const BUDGET: &str = "budget";

/// One failed gate: its reason word and a line that never repeats an item's text.
#[derive(Debug, Clone, PartialEq)]
struct Failure {
    reason: &'static str,
    detail: String,
}

fn fail(reason: &'static str, detail: &str) -> Failure {
    Failure {
        reason,
        detail: detail.to_string(),
    }
}

#[derive(Debug, Clone, Copy, PartialEq)]
enum Op {
    Add,
    Replace,
    Retire,
}

impl Op {
    fn word(self) -> &'static str {
        match self {
            Op::Add => "add",
            Op::Replace => "replace",
            Op::Retire => "retire",
        }
    }
}

/// One item of a proposal, checked for its shape.
#[derive(Debug, Clone, PartialEq)]
struct Item {
    op: Op,
    id: String,
    rule: String,
    reason: String,
    source: Vec<String>,
}

impl Item {
    /// The line the item is in the managed block (the form the operations document fixes).
    fn line(&self) -> String {
        format!(
            "- [{}] {} | why: {} | source: {} | status: trial | helped: 0 | hurt: 0",
            self.id,
            self.rule,
            self.reason,
            self.source.join(", ")
        )
    }
}

pub fn run(args: Args) -> Result<Value, String> {
    let root = std::fs::canonicalize(&args.project)
        .map_err(|_| moai_cli::contract::PROJECT_PATH_UNAVAILABLE.to_string())?;
    let rel = args.target.as_str();
    let mut failures: Vec<Failure> = Vec::new();

    // Gate: the target path (allow-list, test path).
    let (path_failures, readable) = path_gate(&root, rel);
    failures.extend(path_failures);
    let owned = read_manifest(&root).and_then(|m| owned_entry(&m, rel));
    if owned.is_none() || !is_project_path(rel) {
        failures.push(fail(
            NOT_ALLOWED,
            // Not worded with "generation manifest": `contract::classify` reads that phrase as a
            // conflict, and a refusal here is `invalid_input`.
            "the target is not a project skill or agent file the generated-files record owns",
        ));
    }
    if is_test_path(rel) {
        failures.push(fail(
            TEST_WEAKENING,
            "the target is a test or an evaluation set file",
        ));
    }

    // Gates: the items.
    let (items, item_failures) = parse_items(&args.items);
    let items_ok = item_failures.is_empty();
    failures.extend(item_failures);
    failures.extend(secret_gate(&items));
    failures.extend(drift_gate(&root, &items));

    // Gates that read the target: front matter, size, merge and budget.
    let mut old_bytes: Vec<u8> = Vec::new();
    let mut new_text: Option<String> = None;
    if readable {
        match read_target(&root, rel) {
            Err(detail) => failures.push(fail(NOT_ALLOWED, &detail)),
            Ok(None) => failures.push(fail(
                BUDGET,
                "the target is larger than the read limit of 1 MiB",
            )),
            Ok(Some(bytes)) => {
                old_bytes = bytes;
                match std::str::from_utf8(&old_bytes) {
                    Err(_) => failures.push(fail(FRONT_MATTER, "the target is not UTF-8 text")),
                    Ok(text) => {
                        if let Err(detail) = front_matter(text) {
                            failures.push(fail(FRONT_MATTER, &detail));
                        }
                        if items_ok {
                            match merge(text, &items) {
                                Err(Merge::Markers) => failures.push(fail(
                                    NOT_ALLOWED,
                                    "the managed block moai-lessons is not one begin and one end marker",
                                )),
                                Err(Merge::Items(detail)) => failures.push(fail(ITEMS, &detail)),
                                Ok(merged) => {
                                    let budget = budget_for(rel);
                                    let lines = merged.lines().count();
                                    if !within_budget(lines, budget) {
                                        failures.push(fail(
                                            BUDGET,
                                            &format!("the new file has {lines} lines, over the budget of {budget}"),
                                        ));
                                    }
                                    new_text = Some(merged);
                                }
                            }
                        }
                    }
                }
            }
        }
    }

    if !failures.is_empty() {
        return Err(refusal(&failures));
    }
    let (owned_kind, baseline) = owned.expect("an owned target passed the allow-list");
    let new = new_text.expect("a merged text exists when no gate failed");
    let old_text = String::from_utf8_lossy(&old_bytes);
    let old_sha = sha_hex(&old_bytes);
    let new_sha = sha_hex(new.as_bytes());
    let generation = if !baseline.is_empty() && baseline == old_sha {
        "unchanged"
    } else {
        "conflict"
    };
    Ok(json!({
        "action": "plan", "grade": "A", "target": rel, "kind": owned_kind,
        "generation": generation, "old_sha256": old_sha, "new_sha256": new_sha,
        "old_lines": old_text.lines().count(), "new_lines": new.lines().count(),
        "budget_lines": budget_for(rel),
        "items": items.iter().map(item_json).collect::<Vec<_>>(),
        "plan_digest": plan_digest(rel, &old_sha, &new_sha, &baseline, &items),
        "writes": false,
    }))
}

fn item_json(item: &Item) -> Value {
    match item.op {
        Op::Retire => json!({"op": "retire", "id": item.id}),
        op => json!({"op": op.word(), "id": item.id, "line": item.line()}),
    }
}

/// The refusal message: the sorted reasons, then one detail per failure.
fn refusal(failures: &[Failure]) -> String {
    let reasons: BTreeSet<&str> = failures.iter().map(|f| f.reason).collect();
    let mut details: Vec<String> = Vec::new();
    for f in failures {
        let d = format!("{}: {}", f.reason, f.detail);
        if !details.contains(&d) {
            details.push(d);
        }
    }
    format!(
        "improve-plan refused: reasons={} · {}",
        reasons.into_iter().collect::<Vec<_>>().join(","),
        details.join(" · ")
    )
}

fn sha_hex(bytes: &[u8]) -> String {
    format!("{:x}", Sha256::digest(bytes))
}

// ---------------------------------------------------------------------------------------------
// The target path.

/// The failures of the path itself and whether the target is a regular file below the project
/// folder that no link and no `..` leads to (so it may be read).
fn path_gate(root: &Path, rel: &str) -> (Vec<Failure>, bool) {
    let bad = |detail: &str| (vec![fail(NOT_ALLOWED, detail)], false);
    if rel.is_empty() || rel.contains('\\') || rel.contains('\0') {
        return bad("the target path is empty or not a plain relative path");
    }
    let path = Path::new(rel);
    if path.is_absolute() {
        return bad("the target path is absolute");
    }
    let parts: Vec<&std::ffi::OsStr> = path
        .components()
        .filter_map(|c| match c {
            Component::Normal(p) => Some(Ok(p)),
            Component::CurDir => None,
            _ => Some(Err(())),
        })
        .collect::<Result<_, _>>()
        .unwrap_or_default();
    if parts.is_empty() || path.components().any(|c| !matches!(c, Component::Normal(_))) {
        return bad("the target path has a `.` or `..` part");
    }
    let mut cur = root.to_path_buf();
    for (i, part) in parts.iter().enumerate() {
        cur.push(part);
        let Ok(meta) = std::fs::symlink_metadata(&cur) else {
            return bad("the target does not exist below the project folder");
        };
        if meta.file_type().is_symlink() {
            return bad("the target path goes through a symbolic link");
        }
        let last = i + 1 == parts.len();
        if last && !meta.is_file() {
            return bad("the target is not a regular file");
        }
        if !last && !meta.is_dir() {
            return bad("the target path goes through something that is not a folder");
        }
    }
    (Vec::new(), true)
}

/// The grade A shape: `.claude/skills/<moai-prj-…|prj-…>/SKILL.md`, `.claude/agents/prj-….md` or
/// `.claude/agents/moai/prj-….md`.
fn is_project_path(rel: &str) -> bool {
    let parts: Vec<&str> = rel.split('/').collect();
    match parts.as_slice() {
        [".claude", "skills", dir, "SKILL.md"] => {
            dir.strip_prefix("moai-prj-")
                .or_else(|| dir.strip_prefix("prj-"))
                .is_some_and(|name| !name.is_empty())
        }
        [".claude", "agents", file] | [".claude", "agents", "moai", file] => {
            file.strip_prefix("prj-")
                .and_then(|n| n.strip_suffix(".md"))
                .is_some_and(|name| !name.is_empty())
        }
        _ => false,
    }
}

/// A test or an evaluation set, by the path alone (no word list on any text).
fn is_test_path(rel: &str) -> bool {
    let mut parts = rel.split('/').filter(|p| !p.is_empty()).peekable();
    while let Some(part) = parts.next() {
        if matches!(part, "tests" | "test" | "evals" | "__tests__") {
            return true;
        }
        if parts.peek().is_none() {
            let stem = part.split('.').next().unwrap_or(part);
            return stem.starts_with("test_")
                || stem.ends_with("_test")
                || part.contains(".test.")
                || part.contains(".spec.");
        }
    }
    false
}

/// The generation manifest of the project, or `None` when it is absent, a link or too large.
fn read_manifest(root: &Path) -> Option<Value> {
    let path = state_root::path(root, "manifests/generation.json");
    let meta = std::fs::symlink_metadata(&path).ok()?;
    if !meta.is_file() || meta.len() > MANIFEST_BYTES {
        return None;
    }
    serde_json::from_slice(&std::fs::read(path).ok()?).ok()
}

/// The `(kind, baseline_sha256)` of the `owned_files` entry that owns `rel` as a grade A file.
fn owned_entry(manifest: &Value, rel: &str) -> Option<(&'static str, String)> {
    manifest
        .get("owned_files")?
        .as_array()?
        .iter()
        .filter(|e| e.get("path").and_then(Value::as_str) == Some(rel))
        .find_map(|e| {
            let section = e.get("section_id").and_then(Value::as_str)?;
            let kind = if section.starts_with(SECTION_SKILL) {
                "project-skill"
            } else if section.starts_with(SECTION_AGENT) {
                "native-role"
            } else {
                return None;
            };
            let baseline = e
                .get("baseline_sha256")
                .and_then(Value::as_str)
                .unwrap_or("")
                .to_string();
            Some((kind, baseline))
        })
}

/// The bytes of the target, or `None` when it is over the read limit.
fn read_target(root: &Path, rel: &str) -> Result<Option<Vec<u8>>, String> {
    let mut bytes = Vec::new();
    std::fs::File::open(root.join(rel))
        .and_then(|f| f.take(FILE_BYTES + 1).read_to_end(&mut bytes))
        .map_err(|_| "the target cannot be read".to_string())?;
    Ok((bytes.len() as u64 <= FILE_BYTES).then_some(bytes))
}

// ---------------------------------------------------------------------------------------------
// The items.

/// The items of a proposal and the failures of their shape and count.
fn parse_items(raw: &[String]) -> (Vec<Item>, Vec<Failure>) {
    let mut failures = Vec::new();
    let mut items: Vec<Item> = Vec::new();
    if raw.is_empty() || raw.len() > MAX_ITEMS {
        failures.push(fail(ITEMS, "a proposal has one to three items"));
    }
    for (n, text) in raw.iter().enumerate() {
        match parse_item(text) {
            Ok(item) => {
                if items.iter().any(|i| i.id == item.id) {
                    failures.push(fail(ITEMS, &format!("item {} repeats an id", n + 1)));
                }
                items.push(item);
            }
            Err(why) => failures.push(fail(ITEMS, &format!("item {}: {why}", n + 1))),
        }
    }
    (items, failures)
}

fn parse_item(raw: &str) -> Result<Item, String> {
    let v: Value = serde_json::from_str(raw).map_err(|_| "not a JSON object".to_string())?;
    let obj = v.as_object().ok_or("not a JSON object")?;
    let op = match obj.get("op").and_then(Value::as_str) {
        Some("add") => Op::Add,
        Some("replace") => Op::Replace,
        Some("retire") => Op::Retire,
        _ => return Err("op is add, replace or retire".into()),
    };
    let allowed: &[&str] = if op == Op::Retire {
        &["op", "id"]
    } else {
        &["op", "id", "rule", "reason", "source"]
    };
    if obj.keys().any(|k| !allowed.contains(&k.as_str())) {
        return Err("a field the op does not take".into());
    }
    let id = obj.get("id").and_then(Value::as_str).unwrap_or("");
    if !valid_id(id) {
        return Err("id is 1 to 40 characters of a-z, 0-9, - and _".into());
    }
    if op == Op::Retire {
        return Ok(Item {
            op,
            id: id.to_string(),
            rule: String::new(),
            reason: String::new(),
            source: Vec::new(),
        });
    }
    let text = |key: &str| -> Result<String, String> {
        let s = obj.get(key).and_then(Value::as_str).unwrap_or("");
        let n = s.chars().count();
        if n == 0 || n > TEXT_CHARS || !plain_line(s) {
            return Err(format!("{key} is 1 to 200 characters of one plain line"));
        }
        Ok(s.to_string())
    };
    let rule = text("rule")?;
    let reason = text("reason")?;
    let source: Vec<String> = obj
        .get("source")
        .and_then(Value::as_array)
        .ok_or("source is a list of one to five signal ids")?
        .iter()
        .map(|s| s.as_str().map(String::from))
        .collect::<Option<_>>()
        .ok_or("source holds strings")?;
    if source.is_empty() || source.len() > MAX_SOURCES {
        return Err("source holds one to five signal ids".into());
    }
    if source
        .iter()
        .any(|s| s.is_empty() || s.chars().count() > SOURCE_CHARS || !plain_line(s))
    {
        return Err("a source id is 1 to 120 characters of one plain line".into());
    }
    Ok(Item {
        op,
        id: id.to_string(),
        rule,
        reason,
        source,
    })
}

fn valid_id(id: &str) -> bool {
    let n = id.chars().count();
    (1..=ID_CHARS).contains(&n)
        && id.chars().next().is_some_and(|c| c.is_ascii_lowercase() || c.is_ascii_digit())
        && id
            .chars()
            .all(|c| c.is_ascii_lowercase() || c.is_ascii_digit() || c == '-' || c == '_')
}

/// One line with no control character and nothing that could end the item or the block.
fn plain_line(s: &str) -> bool {
    !s.chars().any(char::is_control)
        && !s.contains('|')
        && !s.contains("<!--")
        && !s.contains("-->")
}

/// Every text of an item that carries a secret shape.
fn secret_gate(items: &[Item]) -> Vec<Failure> {
    let shaped = |s: &str| redact::redact_secret_shapes(s) != s;
    let hit = items.iter().any(|i| {
        shaped(&i.id)
            || shaped(&i.rule)
            || shaped(&i.reason)
            || i.source.iter().any(|s| shaped(s))
    });
    if hit {
        vec![fail(SECRET, "an item holds a word shaped like a secret")]
    } else {
        Vec::new()
    }
}

/// A path that does not exist and a `moai` command the binary does not have, named by an item.
fn drift_gate(root: &Path, items: &[Item]) -> Vec<Failure> {
    let mut out = Vec::new();
    for item in items.iter().filter(|i| i.op != Op::Retire) {
        let text = format!("{}\n{}", item.rule, item.reason);
        if cited_paths(&text)
            .iter()
            .any(|p| std::fs::symlink_metadata(root.join(p)).is_err())
        {
            out.push(fail(DRIFT, "an item names a project path that does not exist"));
        }
        if moai_words(&text).iter().any(|w| help::group_of(w).is_none()) {
            out.push(fail(DRIFT, "an item names a moai command this binary does not have"));
        }
    }
    out.dedup();
    out
}

/// The relative file paths a text names (a word of letters, digits and `_ . / -` ending in a
/// known extension), each once. A copy of the reading `improve-signals` uses for audit notes.
fn cited_paths(text: &str) -> BTreeSet<String> {
    let mut out = BTreeSet::new();
    for word in
        text.split(|c: char| !(c.is_ascii_alphanumeric() || matches!(c, '_' | '.' | '/' | '-')))
    {
        let word = word.strip_prefix("./").unwrap_or(word).trim_end_matches('.');
        if word.is_empty()
            || word.len() > 200
            || word.starts_with('/')
            || word.starts_with('-')
            || word.contains("..")
        {
            continue;
        }
        let Some((stem, ext)) = word.rsplit_once('.') else {
            continue;
        };
        if !stem.is_empty() && !stem.ends_with('/') && EXTENSIONS.contains(&ext) {
            out.insert(word.to_string());
        }
    }
    out
}

/// The command word of every backticked `` `moai <word> …` `` span (prose is not read).
fn moai_words(text: &str) -> Vec<String> {
    text.split('`')
        .skip(1)
        .step_by(2)
        .filter_map(|span| {
            let rest = span.trim().strip_prefix("moai ")?.trim_start();
            let word: String = rest
                .chars()
                .take_while(|c| c.is_ascii_lowercase() || c.is_ascii_digit() || *c == '-')
                .collect();
            (!word.is_empty() && !word.starts_with('-')).then_some(word)
        })
        .collect()
}

// ---------------------------------------------------------------------------------------------
// The target text: front matter, merge, budgets.

/// The front matter at the top of the file: `---`, a `name`, a `description` of at most 1,024
/// characters, `---`.
fn front_matter(text: &str) -> Result<(), String> {
    let mut lines = text.lines();
    if lines.next() != Some("---") {
        return Err("the target does not start with front matter".into());
    }
    let (mut name, mut description) = (String::new(), String::new());
    let (mut closed, mut folding) = (false, false);
    for l in lines {
        if l == "---" {
            closed = true;
            break;
        }
        if let Some(v) = l.strip_prefix("name:") {
            name = v.trim().to_string();
            folding = false;
        } else if let Some(v) = l.strip_prefix("description:") {
            let v = v.trim();
            folding = matches!(v, ">" | "|" | ">-" | "|-");
            description = if folding { String::new() } else { unquote(v) };
        } else if folding && l.starts_with(char::is_whitespace) {
            if !description.is_empty() {
                description.push(' ');
            }
            description.push_str(l.trim());
        } else {
            folding = false;
        }
    }
    if !closed {
        return Err("the front matter has no closing line".into());
    }
    if name.is_empty() {
        return Err("the front matter has no name".into());
    }
    if description.is_empty() {
        return Err("the front matter has no description".into());
    }
    if description.chars().count() > DESCRIPTION_CHARS {
        return Err("the description is over 1,024 characters".into());
    }
    Ok(())
}

fn unquote(v: &str) -> String {
    for q in ['"', '\''] {
        if v.len() >= 2 && v.starts_with(q) && v.ends_with(q) {
            return v[1..v.len() - 1].to_string();
        }
    }
    v.to_string()
}

#[derive(Debug, PartialEq)]
enum Merge {
    /// The block markers are not one begin and one end in that order.
    Markers,
    /// An item cannot be applied (the reason is a line that repeats no item text).
    Items(String),
}

fn marker_line(line: &str, marker: &str) -> bool {
    line.trim_end_matches(['\r', '\n']) == marker
}

fn item_prefix(id: &str) -> String {
    format!("- [{id}] ")
}

/// The text with the items merged by id inside the block `moai-lessons`: `add` goes last in the
/// block, `replace` keeps its place, `retire` removes a `trial` item. Without a block `add` makes
/// one at the end of the file. Every byte outside the block stays as it was.
fn merge(text: &str, items: &[Item]) -> Result<String, Merge> {
    let lines: Vec<&str> = text.split_inclusive('\n').collect();
    let begins: Vec<usize> = (0..lines.len()).filter(|&i| marker_line(lines[i], BEGIN)).collect();
    let ends: Vec<usize> = (0..lines.len()).filter(|&i| marker_line(lines[i], END)).collect();
    if text.matches(BEGIN).count() != begins.len() || text.matches(END).count() != ends.len() {
        return Err(Merge::Markers);
    }
    let (head, mut block, tail): (String, Vec<String>, String) = match (begins.as_slice(), ends.as_slice()) {
        ([], []) => {
            if items.iter().any(|i| i.op != Op::Add) {
                return Err(Merge::Items("replace and retire need the block moai-lessons".into()));
            }
            let mut head = text.to_string();
            if !head.is_empty() && !head.ends_with('\n') {
                head.push('\n');
            }
            head.push('\n');
            head.push_str(BEGIN);
            head.push('\n');
            (head, Vec::new(), format!("{END}\n"))
        }
        ([b], [e]) if b < e => (
            lines[..=*b].concat(),
            lines[b + 1..*e].iter().map(|l| l.to_string()).collect(),
            lines[*e..].concat(),
        ),
        _ => return Err(Merge::Markers),
    };
    for item in items {
        let prefix = item_prefix(&item.id);
        let at = block.iter().position(|l| l.starts_with(&prefix));
        match (item.op, at) {
            (Op::Add, Some(_)) => return Err(Merge::Items("add names an id that exists".into())),
            (Op::Add, None) => block.push(format!("{}\n", item.line())),
            (Op::Replace, Some(i)) => block[i] = format!("{}\n", item.line()),
            (Op::Retire, Some(i)) => {
                if !block[i].contains(" | status: trial | ") {
                    return Err(Merge::Items("retire names an item that is not a trial".into()));
                }
                block.remove(i);
            }
            (Op::Replace | Op::Retire, None) => {
                return Err(Merge::Items("replace and retire name an id that does not exist".into()))
            }
        }
    }
    Ok(format!("{head}{}{tail}", block.concat()))
}

/// The line budget of a target by its place: skills 500 lines, agent files 200.
fn budget_for(rel: &str) -> usize {
    if rel.starts_with(".claude/skills/") {
        SKILL_LINES
    } else {
        AGENT_LINES
    }
}

/// N lines pass and N+1 do not. The lines are those of the new bytes, so a proposal that retires
/// one trial item and adds one is judged at the length it leaves.
fn within_budget(lines: usize, budget: usize) -> bool {
    lines <= budget
}

// ---------------------------------------------------------------------------------------------
// The digest.

/// The digest of a plan: its inputs only, never the project folder or the clock.
fn plan_digest(rel: &str, old_sha: &str, new_sha: &str, baseline: &str, items: &[Item]) -> String {
    let mut h = Sha256::new();
    h.update(b"improve-plan@1\n");
    for part in [rel, old_sha, new_sha, baseline] {
        h.update(part.as_bytes());
        h.update(b"\n");
    }
    for item in items {
        let canonical: BTreeMap<&str, Value> = BTreeMap::from([
            ("id", json!(item.id)),
            ("op", json!(item.op.word())),
            ("reason", json!(item.reason)),
            ("rule", json!(item.rule)),
            ("source", json!(item.source)),
        ]);
        h.update(serde_json::to_string(&canonical).unwrap_or_default().as_bytes());
        h.update(b"\n");
    }
    format!("sha256:{:x}", h.finalize())
}

#[cfg(test)]
mod tests {
    use super::*;

    fn add(id: &str) -> Item {
        Item {
            op: Op::Add,
            id: id.into(),
            rule: "Rule.".into(),
            reason: "Why.".into(),
            source: vec!["s".into()],
        }
    }

    fn retire(id: &str) -> Item {
        Item {
            op: Op::Retire,
            id: id.into(),
            rule: String::new(),
            reason: String::new(),
            source: Vec::new(),
        }
    }

    /// A text of `n` lines that holds the block with the trial item `old`.
    fn text_of(n: usize) -> String {
        let mut t = format!("{BEGIN}\n{}\n{END}\n", add("old").line());
        while t.lines().count() < n {
            t.push_str("filler\n");
        }
        t
    }

    #[test]
    fn improve_plan_budgets_pass_at_n_and_refuse_at_n_plus_one() {
        for (budget, rel) in [
            (500, ".claude/skills/moai-prj-x/SKILL.md"),
            (200, ".claude/agents/moai/prj-x.md"),
        ] {
            assert_eq!(budget_for(rel), budget);
            assert!(within_budget(budget, budget), "N passes");
            assert!(!within_budget(budget + 1, budget), "N+1 is refused");
            // One more item on a file at N-1 is N; on a file at N it is N+1; with one trial item
            // retired it is N again.
            let lines = |text: &str, items: &[Item]| merge(text, items).unwrap().lines().count();
            assert_eq!(lines(&text_of(budget - 1), &[add("new")]), budget);
            assert_eq!(lines(&text_of(budget), &[add("new")]), budget + 1);
            assert_eq!(lines(&text_of(budget), &[add("new"), retire("old")]), budget);
        }
        assert_eq!(budget_for(".claude/skills/prj-x/SKILL.md"), SKILL_LINES);
    }

    #[test]
    fn improve_plan_the_description_limit_is_1024_characters() {
        let with = |n: usize| format!("---\nname: a\ndescription: {}\n---\n", "d".repeat(n));
        assert!(front_matter(&with(1024)).is_ok());
        assert!(front_matter(&with(1025)).is_err());
        assert!(front_matter("---\nname: a\ndescription: >\n  one\n  two\n---\n").is_ok());
        assert!(front_matter("---\nname: a\ndescription: >\n---\n").is_err());
        assert!(front_matter("---\nname: a\ndescription: \"quoted\"\n---\n").is_ok());
    }

    #[test]
    fn improve_plan_a_test_is_known_by_its_path_alone() {
        for yes in ["tests/a.rs", "pkg/tests/a.md", "a/evals/x.md", "x/test/y.md", "src/a_test.rs", "src/test_a.py", "w/a.test.ts", "w/a.spec.ts", "a/__tests__/b.js"] {
            assert!(is_test_path(yes), "{yes}");
        }
        for no in [".claude/skills/moai-prj-x/SKILL.md", ".claude/agents/moai/prj-x.md", "docs/testing-guide.md", "src/contest.rs", "src/attest.md"] {
            assert!(!is_test_path(no), "{no}");
        }
    }

    #[test]
    fn improve_plan_the_allow_list_shape_is_grade_a_only() {
        for yes in [".claude/skills/moai-prj-x/SKILL.md", ".claude/skills/prj-x/SKILL.md", ".claude/agents/prj-x.md", ".claude/agents/moai/prj-x.md"] {
            assert!(is_project_path(yes), "{yes}");
        }
        for no in [".claude/skills/moai/SKILL.md", ".claude/skills/moai-prj-/SKILL.md", ".claude/skills/other/SKILL.md", ".claude/agents/other.md", ".claude/agents/moai/other.md", ".claude/agents/moai/prj-x.txt", ".claude/settings.json", "CLAUDE.md", "CLAUDE.local.md", ".claude/skills/moai-prj-x/extra.md", ".claude/agents/moai/deep/prj-x.md"] {
            assert!(!is_project_path(no), "{no}");
        }
    }

    #[test]
    fn improve_plan_the_digest_is_a_function_of_its_inputs_only() {
        let base = plan_digest("a.md", "o", "n", "b", &[add("x")]);
        assert_eq!(base, plan_digest("a.md", "o", "n", "b", &[add("x")]));
        assert_ne!(base, plan_digest("b.md", "o", "n", "b", &[add("x")]));
        assert_ne!(base, plan_digest("a.md", "O", "n", "b", &[add("x")]));
        assert_ne!(base, plan_digest("a.md", "o", "N", "b", &[add("x")]));
        assert_ne!(base, plan_digest("a.md", "o", "n", "B", &[add("x")]));
        assert_ne!(base, plan_digest("a.md", "o", "n", "b", &[add("y")]));
        let mut other = add("x");
        other.reason = "Other.".into();
        assert_ne!(base, plan_digest("a.md", "o", "n", "b", &[other]));
    }

    #[test]
    fn improve_plan_the_block_markers_have_the_form_agents_rs_writes() {
        assert_eq!(BEGIN, "<!-- moai:managed:begin id=moai-lessons -->");
        assert_eq!(END, "<!-- moai:managed:end id=moai-lessons -->");
    }

    #[test]
    fn improve_plan_a_command_is_read_only_inside_backticks() {
        assert_eq!(moai_words("Run `moai card show C-1` and `moai improve-plan`, not moai skill."), vec!["card", "improve-plan"]);
        assert!(moai_words("`moai --version` and `moai ` and ``").is_empty());
        assert!(help::group_of("card").is_some());
    }

    #[test]
    fn improve_plan_merge_keeps_every_byte_outside_the_block() {
        let text = format!("before\n{BEGIN}\n{}\n{END}\nafter without newline", add("old").line());
        let merged = merge(&text, &[add("new")]).unwrap();
        assert!(merged.starts_with("before\n"));
        assert!(merged.ends_with(&format!("{END}\nafter without newline")));
        assert!(merge("no block", &[retire("x")]).is_err());
        assert_eq!(merge(&format!("{END}\n{BEGIN}\n"), &[add("x")]), Err(Merge::Markers));
        assert_eq!(merge(&format!("a {BEGIN} b\n"), &[add("x")]), Err(Merge::Markers));
    }
}
