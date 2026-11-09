//! Semantic parity check (D6 #2192): classify key differences between a template
//! tree and a generated tree, with an optional third (doc) set, reusing the 00-A
//! classification vocabulary from `docs/baseline-rc1.md`: a key present in exactly
//! one set names that set (`template-only` = 누락/missing, `generated-only` =
//! 추가/extra, `doc-only` = doc set only); a key in two or more sets is `identical`
//! when all content digests agree and otherwise `differing` (값 상이/value
//! mismatch). Plain filesystem walks only — no git, nothing written, credential
//! values are not a concern because the caller names the trees.

use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{collections::BTreeMap, fs, path::Path};

type Result<T> = std::result::Result<T, String>;

const FILE_LIMIT: u64 = 262_144;
const WALK_DEPTH: usize = 32;
const WALK_ENTRIES: usize = 8192;
const WALK_FILES: usize = 4096;
const WALK_BYTES: usize = 16 * 1024 * 1024;

fn walk_set(
    root: &Path,
    dir: &Path,
    out: &mut BTreeMap<String, String>,
    entries: &mut usize,
    total: &mut usize,
    depth: usize,
) -> Result<()> {
    if depth > WALK_DEPTH {
        return Err("parity nesting limit exceeded".into());
    }
    for entry in fs::read_dir(dir).map_err(|_| "parity directory unavailable".to_string())? {
        *entries += 1;
        if *entries > WALK_ENTRIES {
            return Err("parity entry limit exceeded".into());
        }
        let path = entry
            .map_err(|_| "parity entry unavailable".to_string())?
            .path();
        let meta =
            fs::symlink_metadata(&path).map_err(|_| "parity metadata unavailable".to_string())?;
        if meta.file_type().is_symlink() {
            return Err("parity input contains a symlink; refused".into());
        }
        if meta.is_dir() {
            walk_set(root, &path, out, entries, total, depth + 1)?;
        } else if meta.is_file() {
            if meta.len() > FILE_LIMIT {
                return Err("parity input file exceeds the size limit".into());
            }
            let rel = path
                .strip_prefix(root)
                .map_err(|_| "parity path invalid".to_string())?
                .to_str()
                .ok_or("non UTF-8 parity path".to_string())?
                .replace('\\', "/");
            let bytes = fs::read(&path).map_err(|_| "parity input read failed".to_string())?;
            *total += bytes.len();
            if *total > WALK_BYTES {
                return Err("parity input too large".into());
            }
            out.insert(rel, format!("sha256:{:x}", Sha256::digest(&bytes)));
        } else {
            return Err("parity special file refused".into());
        }
        if out.len() > WALK_FILES {
            return Err("parity file limit exceeded".into());
        }
    }
    Ok(())
}

fn scan_set(root: &Path) -> Result<BTreeMap<String, String>> {
    let meta =
        fs::symlink_metadata(root).map_err(|_| "parity directory unavailable".to_string())?;
    if meta.file_type().is_symlink() {
        return Err("parity directory is a symlink; refused".into());
    }
    if !meta.is_dir() {
        return Err("parity directory is not a directory".into());
    }
    let mut out = BTreeMap::new();
    let mut entries = 0usize;
    let mut total = 0usize;
    walk_set(root, root, &mut out, &mut entries, &mut total, 0)?;
    Ok(out)
}

/// Classifies one key across the three sets with the 00-A vocabulary: a single-set
/// key names its set, a multi-set key is `identical` when every digest agrees and
/// `differing` otherwise. `kind` is the REQ-D6-01 grouping (누락 missing / 추가
/// extra / 값 상이 value_mismatch / identical / doc_only).
fn classify(
    t: Option<&String>,
    g: Option<&String>,
    d: Option<&String>,
) -> (&'static str, &'static str) {
    let present = [t.is_some(), g.is_some(), d.is_some()];
    let count = present.iter().filter(|x| **x).count();
    if count == 1 {
        if t.is_some() {
            ("template-only", "missing")
        } else if g.is_some() {
            ("generated-only", "extra")
        } else {
            ("doc-only", "doc_only")
        }
    } else {
        let digests: Vec<&str> = [t, g, d]
            .into_iter()
            .flatten()
            .map(String::as_str)
            .collect();
        if digests.windows(2).all(|w| w[0] == w[1]) {
            ("identical", "identical")
        } else {
            ("differing", "value_mismatch")
        }
    }
}

/// Read-only parity report over template ↔ generated (↔ optional doc set)
/// (REQ-D6-01: differences separated by kind; AC-2 report format). Writes nothing.
pub fn parity_check(template: &Path, generated: &Path, doc: Option<&Path>) -> Result<Value> {
    let t = scan_set(template)?;
    let g = scan_set(generated)?;
    let d = match doc {
        Some(p) => Some(scan_set(p)?),
        None => None,
    };

    let mut keys: Vec<String> = t.keys().cloned().collect();
    for k in g.keys() {
        if !t.contains_key(k) {
            keys.push(k.clone());
        }
    }
    if let Some(dd) = &d {
        for k in dd.keys() {
            if !t.contains_key(k) && !g.contains_key(k) {
                keys.push(k.clone());
            }
        }
    }
    keys.sort();

    let mut missing = 0i64;
    let mut extra = 0i64;
    let mut value_mismatch = 0i64;
    let mut identical = 0i64;
    let mut doc_only = 0i64;
    let mut files = Vec::new();
    for key in &keys {
        let (class, kind) = classify(
            t.get(key),
            g.get(key),
            d.as_ref().and_then(|dd| dd.get(key)),
        );
        match kind {
            "missing" => missing += 1,
            "extra" => extra += 1,
            "value_mismatch" => value_mismatch += 1,
            "identical" => identical += 1,
            _ => doc_only += 1,
        }
        files.push(json!({
            "path": key,
            "class": class,
            "kind": kind,
            "template_sha256": t.get(key),
            "generated_sha256": g.get(key),
            "doc_sha256": d.as_ref().and_then(|dd| dd.get(key)),
        }));
    }

    Ok(json!({
        "counts": {
            "missing": missing,
            "extra": extra,
            "value_mismatch": value_mismatch,
            "identical": identical,
            "doc_only": doc_only,
            "total": keys.len(),
        },
        "files": files,
        "sets": {
            "template": template.to_string_lossy(),
            "generated": generated.to_string_lossy(),
            "doc": doc.map(|p| p.to_string_lossy()),
        },
        "writes": false,
    }))
}
