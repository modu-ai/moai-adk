//! Go `encoding/json` decoding of the statusline payload into the Go `StdinData` shape
//! (#2350 audit 2), as a JSON value the serde types in `input.rs` then read.
//!
//! serde's derive rejects a duplicate key and reads numbers through `f64`/`i64` only, so it
//! blanks payloads Go renders. This module reads the first JSON value keeping every key in
//! order and every number's text, then folds it the way Go v1 `json.Decoder.Decode` fills a
//! `StdinData` (go1.26, measured against `moai.old` v3.2.1-rc.6; the rules are Go's
//! `decode.go`):
//! - a later duplicate key decodes into what the earlier one left: an object into a pointer
//!   struct merges field by field, a scalar or string replaces, a slice decodes element by
//!   element into the earlier elements and takes the new length;
//! - `null` clears a pointer, slice or raw field and leaves a plain value as it was;
//! - `ModelInfo` has its own `UnmarshalJSON`: a string sets id, name and display name, an
//!   object replaces the whole value;
//! - keys match a field exactly first, otherwise case-insensitively (Go's `foldName`, which
//!   also folds `ſ` to `s` and the Kelvin sign to `k`); unknown keys are skipped;
//! - an integer field takes only an integer literal (`-0` included) that fits `int64`; a
//!   float field takes any number that does not overflow; any type mismatch fails the whole
//!   payload, as `Decode` returns the first type error.
use serde_json::{Map, Value};

/// Go's `maxNestingDepth` for the decoder's scanner.
pub const MAX_DEPTH: usize = 10_000;
/// Nesting a `json.RawMessage` field may carry here (serde_json's default recursion limit).
const RAW_DEPTH: usize = 128;

/// One JSON value with object keys in order (duplicates kept) and numbers as text.
#[derive(Clone, Debug, PartialEq)]
pub enum Tree {
    Null,
    Bool(bool),
    Number(String),
    Str(String),
    Array(Vec<Tree>),
    Object(Vec<(String, Tree)>),
}

/// Dropping a 10,000-level tree recursively would exhaust a thread's stack; children are
/// moved onto a heap stack and dropped one level at a time.
impl Drop for Tree {
    fn drop(&mut self) {
        let mut pending: Vec<Tree> = match self {
            Tree::Array(items) => std::mem::take(items),
            Tree::Object(entries) => std::mem::take(entries)
                .into_iter()
                .map(|(_, v)| v)
                .collect(),
            _ => return,
        };
        while let Some(mut t) = pending.pop() {
            match &mut t {
                Tree::Array(items) => pending.append(items),
                Tree::Object(entries) => {
                    pending.extend(std::mem::take(entries).into_iter().map(|(_, v)| v))
                }
                _ => {}
            }
        }
    }
}

struct Reader<'a> {
    s: &'a [u8],
    text: &'a str,
    i: usize,
}

impl Reader<'_> {
    fn ws(&mut self) {
        while self.i < self.s.len() && matches!(self.s[self.i], b' ' | b'\t' | b'\n' | b'\r') {
            self.i += 1;
        }
    }

    fn eat(&mut self, lit: &[u8]) -> Option<()> {
        (self.s.get(self.i..self.i + lit.len()) == Some(lit)).then(|| self.i += lit.len())
    }

    /// One value, iteratively: an explicit stack of open containers, so 10,000 levels do
    /// not use the call stack. Go counts open containers (empty ones too) against the limit.
    fn value(&mut self) -> Option<Tree> {
        enum Open {
            Array(Vec<Tree>),
            Object(Vec<(String, Tree)>, String),
        }
        let mut stack: Vec<Open> = Vec::new();
        loop {
            self.ws();
            let mut done = match *self.s.get(self.i)? {
                b'n' => self.eat(b"null").map(|_| Tree::Null)?,
                b't' => self.eat(b"true").map(|_| Tree::Bool(true))?,
                b'f' => self.eat(b"false").map(|_| Tree::Bool(false))?,
                b'"' => Tree::Str(self.string()?),
                open @ (b'[' | b'{') => {
                    if stack.len() + 1 > MAX_DEPTH {
                        return None;
                    }
                    self.i += 1;
                    self.ws();
                    if open == b'[' {
                        if self.eat(b"]").is_some() {
                            Tree::Array(Vec::new())
                        } else {
                            stack.push(Open::Array(Vec::new()));
                            continue;
                        }
                    } else if self.eat(b"}").is_some() {
                        Tree::Object(Vec::new())
                    } else {
                        let key = self.key()?;
                        stack.push(Open::Object(Vec::new(), key));
                        continue;
                    }
                }
                _ => Tree::Number(self.number()?),
            };
            // Attach the finished value to the open containers, closing them as they end.
            loop {
                match stack.last_mut() {
                    None => return Some(done),
                    Some(Open::Array(items)) => {
                        items.push(done);
                        self.ws();
                        if self.eat(b",").is_some() {
                            break;
                        }
                        self.eat(b"]")?;
                        let Some(Open::Array(items)) = stack.pop() else {
                            return None;
                        };
                        done = Tree::Array(items);
                    }
                    Some(Open::Object(entries, key)) => {
                        entries.push((std::mem::take(key), done));
                        self.ws();
                        if self.eat(b",").is_some() {
                            *key = self.key()?;
                            break;
                        }
                        self.eat(b"}")?;
                        let Some(Open::Object(entries, _)) = stack.pop() else {
                            return None;
                        };
                        done = Tree::Object(entries);
                    }
                }
            }
        }
    }

    /// `"key" :` inside an object.
    fn key(&mut self) -> Option<String> {
        self.ws();
        let key = self.string()?;
        self.ws();
        self.eat(b":")?;
        Some(key)
    }

    /// A string token, decoded by serde_json (escapes were made Go-equivalent before).
    fn string(&mut self) -> Option<String> {
        let start = self.i;
        if self.s.get(self.i) != Some(&b'"') {
            return None;
        }
        self.i += 1;
        loop {
            match *self.s.get(self.i)? {
                b'"' => break,
                b'\\' => self.i += 2,
                _ => self.i += 1,
            }
        }
        self.i += 1;
        serde_json::from_str(self.text.get(start..self.i)?).ok()
    }

    /// A number token by the JSON grammar: `-? (0 | [1-9][0-9]*) (. [0-9]+)? ([eE] [+-]? [0-9]+)?`.
    fn number(&mut self) -> Option<String> {
        let start = self.i;
        let digits = |r: &mut Self| {
            let from = r.i;
            while r.s.get(r.i).is_some_and(u8::is_ascii_digit) {
                r.i += 1;
            }
            r.i > from
        };
        let _ = self.eat(b"-");
        if self.eat(b"0").is_none() && !digits(self) {
            return None;
        }
        if self.eat(b".").is_some() && !digits(self) {
            return None;
        }
        if matches!(self.s.get(self.i), Some(b'e' | b'E')) {
            self.i += 1;
            if matches!(self.s.get(self.i), Some(b'+' | b'-')) {
                self.i += 1;
            }
            if !digits(self) {
                return None;
            }
        }
        Some(self.text[start..self.i].to_string())
    }
}

/// The first JSON value of `text` (leading JSON whitespace skipped, anything after it
/// ignored, as `json.Decoder.Decode` reads one value), or `None` if it is malformed.
pub fn first_value(text: &str) -> Option<Tree> {
    Reader {
        s: text.as_bytes(),
        text,
        i: 0,
    }
    .value()
}

/// How a Go field decodes.
#[derive(Clone, Copy, Debug)]
enum Kind {
    Str,
    /// `*string`: `null` clears it.
    PtrStr,
    Int,
    Float,
    /// `*float64`: `null` clears it.
    PtrFloat,
    Bool,
    /// `*Struct`: `null` clears it; an object merges into it.
    Ptr(&'static [Field]),
    /// A struct value (a slice element): `null` leaves it; an object merges into it.
    Val(&'static [Field]),
    /// `*ModelInfo` with its own `UnmarshalJSON`.
    Model,
    /// `json.RawMessage`: any value replaces it.
    Raw,
    /// `[]T`: `null` clears it; an array decodes into the earlier elements.
    Slice(&'static Kind),
}

#[derive(Debug)]
struct Field(&'static str, Kind);

const REPO: &[Field] = &[
    Field("host", Kind::Str),
    Field("owner", Kind::Str),
    Field("name", Kind::Str),
];
const WORKSPACE: &[Field] = &[
    Field("current_dir", Kind::Str),
    Field("project_dir", Kind::Str),
    Field("git_worktree", Kind::Str),
    Field("repo", Kind::Ptr(REPO)),
];
const COST: &[Field] = &[
    Field("total_usd", Kind::Float),
    Field("total_cost_usd", Kind::Float),
    Field("input_tokens", Kind::Int),
    Field("output_tokens", Kind::Int),
    Field("total_duration_ms", Kind::Int),
    Field("total_lines_added", Kind::Int),
    Field("total_lines_removed", Kind::Int),
];
const CURRENT_USAGE: &[Field] = &[
    Field("input_tokens", Kind::Int),
    Field("cache_creation_input_tokens", Kind::Int),
    Field("cache_read_input_tokens", Kind::Int),
    Field("output_tokens", Kind::Int),
];
const CONTEXT_WINDOW: &[Field] = &[
    Field("used_percentage", Kind::PtrFloat),
    Field("remaining_percentage", Kind::PtrFloat),
    Field("context_window_size", Kind::Int),
    Field("total_input_tokens", Kind::Int),
    Field("total_output_tokens", Kind::Int),
    Field("used", Kind::Int),
    Field("total", Kind::Int),
    Field("current_usage", Kind::Ptr(CURRENT_USAGE)),
];
const NAMED: &[Field] = &[Field("name", Kind::Str)];
const RATE_WINDOW: &[Field] = &[
    Field("used_percentage", Kind::Float),
    Field("resets_at", Kind::Int),
];
const RATE_LIMITS: &[Field] = &[
    Field("five_hour", Kind::Ptr(RATE_WINDOW)),
    Field("seven_day", Kind::Ptr(RATE_WINDOW)),
];
const EFFORT: &[Field] = &[Field("level", Kind::Str)];
const THINKING: &[Field] = &[Field("enabled", Kind::Bool)];
const PR: &[Field] = &[
    Field("number", Kind::Int),
    Field("url", Kind::Str),
    Field("review_state", Kind::Str),
];
const WORKTREE: &[Field] = &[
    Field("name", Kind::Str),
    Field("path", Kind::Str),
    Field("branch", Kind::Str),
    Field("original_cwd", Kind::Str),
    Field("original_branch", Kind::Str),
];
const TASK: &[Field] = &[
    Field("id", Kind::Str),
    Field("name", Kind::PtrStr),
    Field("type", Kind::Str),
    Field("agentType", Kind::PtrStr),
    Field("status", Kind::Str),
    Field("description", Kind::Str),
    Field("label", Kind::Str),
    Field("startTime", Kind::Int),
    Field("model", Kind::PtrStr),
    Field("effort", Kind::Raw),
    Field("contextWindowSize", Kind::Int),
    Field("tokenCount", Kind::Int),
    Field("tokenSamples", Kind::Slice(&Kind::Int)),
    Field("cwd", Kind::Str),
];
const MODEL: &[Field] = &[
    Field("id", Kind::Str),
    Field("display_name", Kind::Str),
    Field("name", Kind::Str),
];
/// Go `StdinData` (internal/statusline/types.go).
const STDIN: &[Field] = &[
    Field("hook_event_name", Kind::Str),
    Field("session_id", Kind::Str),
    Field("session_name", Kind::Str),
    Field("transcript_path", Kind::Str),
    Field("cwd", Kind::Str),
    Field("model", Kind::Model),
    Field("workspace", Kind::Ptr(WORKSPACE)),
    Field("cost", Kind::Ptr(COST)),
    Field("context_window", Kind::Ptr(CONTEXT_WINDOW)),
    Field("output_style", Kind::Ptr(NAMED)),
    Field("rate_limits", Kind::Ptr(RATE_LIMITS)),
    Field("effort", Kind::Ptr(EFFORT)),
    Field("thinking", Kind::Ptr(THINKING)),
    Field("version", Kind::Str),
    Field("pr", Kind::Ptr(PR)),
    Field("agent", Kind::Ptr(NAMED)),
    Field("worktree", Kind::Ptr(WORKTREE)),
    // #2756 AC-8: not in the Go struct; kept raw so an unexpected value never fails the
    // payload (the renderer reads only `warm` and `expires_at`).
    Field("prompt_cache", Kind::Raw),
    Field("exceeds_200k_tokens", Kind::Bool),
    Field("tasks", Kind::Slice(&Kind::Val(TASK))),
];

/// Go `foldName` for an ASCII field name: ASCII letters fold to lower case, `ſ` (U+017F)
/// folds to `s` and the Kelvin sign (U+212A) to `k`; no other character folds to ASCII.
fn fold_name(s: &str) -> String {
    s.chars()
        .map(|c| match c {
            '\u{017F}' => 's',
            '\u{212A}' => 'k',
            c => c.to_ascii_lowercase(),
        })
        .collect()
}

fn field<'f>(fields: &'f [Field], key: &str) -> Option<&'f Field> {
    fields
        .iter()
        .find(|f| f.0 == key)
        .or_else(|| fields.iter().find(|f| fold_name(f.0) == fold_name(key)))
}

/// Go `strconv.ParseInt(s, 10, 64)` over a JSON number token.
fn go_int(raw: &str) -> Option<i64> {
    let digits = raw.strip_prefix('-').unwrap_or(raw);
    if digits.is_empty() || !digits.bytes().all(|b| b.is_ascii_digit()) {
        return None;
    }
    raw.parse().ok()
}

/// Go `strconv.ParseFloat(s, 64)`: an out-of-range value is an error, not infinity.
fn go_float(raw: &str) -> Option<f64> {
    raw.parse::<f64>().ok().filter(|f| f.is_finite())
}

/// A raw message as a value; nesting past serde_json's default limit (128) is refused.
fn raw_value(t: &Tree, depth: usize) -> Option<Value> {
    if depth > RAW_DEPTH {
        return None;
    }
    Some(match t {
        Tree::Null => Value::Null,
        Tree::Bool(b) => Value::Bool(*b),
        Tree::Number(n) => serde_json::from_str(n).ok()?,
        Tree::Str(s) => Value::String(s.clone()),
        Tree::Array(items) => Value::Array(
            items
                .iter()
                .map(|i| raw_value(i, depth + 1))
                .collect::<Option<_>>()?,
        ),
        // A raw message keeps duplicate keys; the last one is what a reader of it sees.
        Tree::Object(entries) => {
            let mut map = Map::new();
            for (k, v) in entries {
                map.insert(k.clone(), raw_value(v, depth + 1)?);
            }
            Value::Object(map)
        }
    })
}

/// `Err` is a Go type error (the whole payload then fails); `Ok(None)` leaves the field unset.
fn fold(kind: Kind, prev: Option<Value>, new: &Tree) -> Result<Option<Value>, ()> {
    use Tree as T;
    Ok(match (kind, new) {
        (Kind::Str, T::Str(s)) | (Kind::PtrStr, T::Str(s)) => Some(Value::String(s.clone())),
        (Kind::Int, T::Number(n)) => Some(Value::from(go_int(n).ok_or(())?)),
        (Kind::Float, T::Number(n)) | (Kind::PtrFloat, T::Number(n)) => {
            Some(Value::from(go_float(n).ok_or(())?))
        }
        (Kind::Bool, T::Bool(b)) => Some(Value::Bool(*b)),
        (Kind::Str | Kind::Int | Kind::Float | Kind::Bool | Kind::Val(_), T::Null) => prev,
        (Kind::PtrStr | Kind::PtrFloat | Kind::Ptr(_) | Kind::Model | Kind::Slice(_), T::Null) => {
            None
        }
        (Kind::Ptr(fields) | Kind::Val(fields), T::Object(entries)) => {
            let mut map = match prev {
                Some(Value::Object(m)) => m,
                _ => Map::new(),
            };
            for (key, value) in entries {
                let Some(Field(name, kind)) = field(fields, key) else {
                    continue;
                };
                match fold(*kind, map.remove(*name), value)? {
                    Some(v) => map.insert(name.to_string(), v),
                    None => None,
                };
            }
            Some(Value::Object(map))
        }
        (Kind::Model, T::Str(s)) => {
            Some(serde_json::json!({"id": s, "display_name": s, "name": s}))
        }
        (Kind::Model, T::Object(_)) => fold(Kind::Ptr(MODEL), None, new)?,
        (Kind::Raw, t) => Some(raw_value(t, 1).ok_or(())?),
        (Kind::Slice(elem), T::Array(items)) => {
            let old = match prev {
                Some(Value::Array(a)) => a,
                _ => Vec::new(),
            };
            let mut out = Vec::with_capacity(items.len());
            for (i, item) in items.iter().enumerate() {
                let base = old.get(i).cloned();
                let zero = match elem {
                    Kind::Val(_) => Value::Object(Map::new()),
                    _ => Value::from(0),
                };
                out.push(fold(*elem, base, item)?.unwrap_or(zero));
            }
            Some(Value::Array(out))
        }
        _ => return Err(()),
    })
}

/// The payload as Go would have filled `StdinData`, as JSON for the serde types; `None` when
/// Go's `Decode` would return an error. A top-level `null` is an empty payload.
pub fn decode(text: &str) -> Option<Value> {
    match first_value(text)? {
        Tree::Null => Some(Value::Object(Map::new())),
        t @ Tree::Object(_) => fold(Kind::Ptr(STDIN), None, &t).ok().flatten(),
        _ => None,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    #[test]
    fn duplicate_keys_merge_pointer_structs_and_replace_the_rest() {
        let v = decode(
            r#"{"context_window": {"used_percentage": 40}, "context_window": {"context_window_size": 200000},
                "session_name": "a", "session_name": "b", "session_name": null}"#,
        )
        .unwrap();
        assert_eq!(
            v["context_window"],
            json!({"used_percentage": 40.0, "context_window_size": 200000})
        );
        assert_eq!(v["session_name"], json!("b"), "null leaves a plain string");
        let v = decode(r#"{"context_window": {"used_percentage": 40, "used_percentage": null}}"#)
            .unwrap();
        assert_eq!(v["context_window"], json!({}), "null clears a *float64");
        let v = decode(r#"{"workspace": {"current_dir": "x"}, "workspace": null}"#).unwrap();
        assert_eq!(v.get("workspace"), None, "null clears a pointer struct");
    }

    #[test]
    fn model_replaces_whole_and_reads_a_string() {
        let v =
            decode(r#"{"model": {"id": "x", "display_name": "A"}, "model": {"id": "y"}}"#).unwrap();
        assert_eq!(v["model"], json!({"id": "y"}));
        let v = decode(r#"{"model": {"display_name": "A"}, "model": "B"}"#).unwrap();
        assert_eq!(
            v["model"],
            json!({"id": "B", "display_name": "B", "name": "B"})
        );
        assert_eq!(
            decode(r#"{"model": {"display_name": "A"}, "model": null}"#)
                .unwrap()
                .get("model"),
            None
        );
        assert_eq!(decode(r#"{"model": 3}"#), None);
    }

    #[test]
    fn slices_decode_into_earlier_elements_and_take_the_new_length() {
        let v = decode(
            r#"{"tasks": [{"id": "1", "status": "running"}, {"id": "2"}], "tasks": [{"status": "done"}]}"#,
        )
        .unwrap();
        assert_eq!(v["tasks"], json!([{"id": "1", "status": "done"}]));
        assert_eq!(
            decode(r#"{"tasks": [{"id": "1"}], "tasks": null}"#)
                .unwrap()
                .get("tasks"),
            None
        );
        assert_eq!(
            decode(r#"{"tasks": [null]}"#).unwrap()["tasks"],
            json!([{}])
        );
    }

    #[test]
    fn keys_fold_like_go() {
        let v = decode(r#"{"MODEL": {"DISPLAY_NAME": "U"}, "ſession_id": "s", "Session_Id": "t"}"#)
            .unwrap();
        assert_eq!(v["model"], json!({"display_name": "U"}));
        assert_eq!(v["session_id"], json!("t"), "the later match wins");
        // An exact match is preferred: `model` then `Model` both land on `model`.
        let v = decode(r#"{"Model": "first", "model": "second"}"#).unwrap();
        assert_eq!(v["model"]["id"], json!("second"));
        // Go folds the long s and the Kelvin sign onto ASCII letters.
        let long_s = decode("{\"\u{017F}ession_id\": \"s\"}").unwrap();
        assert_eq!(long_s["session_id"], json!("s"));
        let kelvin = decode("{\"thin\u{212A}ing\": {\"enabled\": true}}").unwrap();
        assert_eq!(kelvin["thinking"], json!({"enabled": true}));
        // Non-ASCII letters that do not fold to ASCII do not match.
        assert_eq!(decode(r#"{"modél": "x"}"#).unwrap().get("model"), None);
    }

    #[test]
    fn numbers_follow_go_parse_int_and_parse_float() {
        let n = |json: &str| decode(&format!(r#"{{"cost": {{"total_duration_ms": {json}}}}}"#));
        assert_eq!(n("-0").unwrap()["cost"]["total_duration_ms"], json!(0));
        assert_eq!(
            n("9223372036854775807").unwrap()["cost"]["total_duration_ms"],
            json!(i64::MAX)
        );
        assert_eq!(n("9223372036854775808"), None, "int64 overflow");
        assert_eq!(
            n("-9223372036854775808").unwrap()["cost"]["total_duration_ms"],
            json!(i64::MIN)
        );
        assert_eq!(n("-9223372036854775809"), None);
        assert_eq!(n("1.0"), None);
        assert_eq!(n("6e4"), None);
        assert_eq!(n("\"60\""), None);
        let f = |json: &str| decode(&format!(r#"{{"cost": {{"total_usd": {json}}}}}"#));
        assert_eq!(f("6e4").unwrap()["cost"]["total_usd"], json!(60000.0));
        assert_eq!(f("1e308").unwrap()["cost"]["total_usd"], json!(1e308));
        assert_eq!(f("1e309"), None, "ParseFloat range error");
    }

    #[test]
    fn reader_takes_the_first_value_and_rejects_bad_grammar() {
        assert_eq!(
            first_value(" \t\r\n{} trailing"),
            Some(Tree::Object(vec![]))
        );
        assert_eq!(
            first_value("\u{feff}{}"),
            None,
            "a BOM is not JSON whitespace"
        );
        for bad in [
            "",
            "{",
            "{\"a\"}",
            "{\"a\":}",
            "[1,]",
            "-",
            "1.",
            "1e",
            "nul",
            "{'a':1}",
            "{\"a\":01}",
        ] {
            assert!(first_value(bad).is_none(), "{bad:?}");
        }
        assert_eq!(decode("[]"), None);
        assert_eq!(decode("\"s\""), None);
        assert_eq!(decode("null"), Some(json!({})));
    }

    /// A task's raw `effort` may nest up to serde_json's 128 levels (a known difference:
    /// Go keeps any depth up to its 10,000-container limit).
    #[test]
    fn raw_effort_nesting_limit() {
        let effort = |n: usize| {
            format!(
                "{{\"tasks\": [{{\"id\": \"t\", \"effort\": {}1{}}}]}}",
                "[".repeat(n),
                "]".repeat(n)
            )
        };
        assert!(decode(&effort(127)).is_some());
        assert!(decode(&effort(128)).is_none());
    }

    #[test]
    fn nesting_depth_limit_is_go_max_depth() {
        let nest = |n: usize| format!("{{\"extra\": {}1{}}}", "[".repeat(n), "]".repeat(n));
        // The outer object and n arrays are n + 1 open containers; Go allows 10,000.
        assert!(decode(&nest(9_999)).is_some());
        assert!(decode(&nest(10_000)).is_none());
        let empty = |n: usize| format!("{{\"extra\": {}{}}}", "[".repeat(n), "]".repeat(n));
        assert!(decode(&empty(9_999)).is_some());
        assert!(
            decode(&empty(10_000)).is_none(),
            "empty containers count too"
        );
    }
}
