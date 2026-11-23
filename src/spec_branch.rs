//! The SPEC id of a work branch (SPEC-20261119-spec-branch-name-first, card C-112).
//!
//! A SPEC work branch is named `spec/<slug>-<yyyymmdd>` (name first); the older
//! `spec/<yyyymmdd>-<slug>` (date first) is still read, and either may end in one repair suffix
//! `-r<N>`. Both forms give the SPEC id `SPEC-<yyyymmdd>-<slug>`. A name can read as two ids
//! (`spec/20261117-demo-20261118`); such a name is never guessed: it is settled only by a SPEC
//! folder under `.moai/specs/` of exactly one of the readings, and otherwise refused by the
//! readers that need one SPEC.
//!
//! One reader for the SPEC-first hook, `moai tdd` and `moai spec take`: no file system or git
//! access besides the optional folder check of [`resolve`].

use crate::spec_doc::{self, IdForm};
use std::path::Path;

/// What a branch name says about its SPEC.
#[derive(Debug, PartialEq, Eq)]
pub enum Resolved {
    /// Not a SPEC work branch in either form.
    NotSpec,
    /// Exactly one SPEC id.
    One(String),
    /// Two or more ids that nothing settles; the ids are in the order of [`candidates`].
    Ambiguous(Vec<String>),
}

/// The SPEC ids a branch name can be read as, without looking at any folder: the name as it
/// stands and the name without a trailing `-r<N>`, each read date first and name first. A
/// reading counts only when it has a date of eight digits and a slug that `spec new` accepts.
pub fn candidates(branch: &str) -> Vec<String> {
    let mut out = Vec::new();
    let Some(rest) = branch.strip_prefix("spec/") else {
        return out;
    };
    // The slug rule is ASCII only, so a name with any other byte is no SPEC branch (and byte
    // offsets below cannot fall inside a character).
    if rest.is_empty() || rest.contains('/') || !rest.is_ascii() {
        return out;
    }
    push_readings(rest, &mut out);
    if let Some((head, number)) = rest.rsplit_once("-r") {
        if !number.is_empty() && number.bytes().all(|b| b.is_ascii_digit()) {
            push_readings(head, &mut out);
        }
    }
    out
}

/// The readings of `name` (an ASCII name without a repair suffix) added to `out`: date first,
/// then name first, each only when it is a SPEC id and not yet listed.
fn push_readings(name: &str, out: &mut Vec<String>) {
    let mut add = |date: &str, slug: &str| {
        let date_ok = date.len() == 8 && date.bytes().all(|b| b.is_ascii_digit());
        if date_ok && slug_ok(slug) {
            let id = format!("SPEC-{date}-{slug}");
            if !out.contains(&id) {
                out.push(id);
            }
        }
    };
    if let Some((date, tail)) = name.split_at_checked(8) {
        if let Some(slug) = tail.strip_prefix('-') {
            add(date, slug);
        }
    }
    if name.len() > 9 {
        let (head, date) = name.split_at(name.len() - 8);
        if let Some(slug) = head.strip_suffix('-') {
            add(date, slug);
        }
    }
}

/// A slug of lower-case letters, digits and hyphens with no hyphen at either end: the rule the
/// SPEC-first hook always read a branch by. It is looser than `spec_doc::valid_slug` (which
/// refuses a hyphen pair and caps the length at 48): the hook never had a cap, and it read
/// `spec/20261117-a--b` as a SPEC branch, so a branch with a longer slug or a hyphen pair must
/// not slip past it now (audit note 3228, N1; D-11: the hook is not made weaker).
fn slug_ok(s: &str) -> bool {
    !s.is_empty()
        && !s.starts_with('-')
        && !s.ends_with('-')
        && s.bytes()
            .all(|c| c.is_ascii_lowercase() || c.is_ascii_digit() || c == b'-')
}

/// The SPEC of `branch`: first the one its name reads (see [`resolve_by_name`]); when that is not
/// an existing SPEC file, the SPEC whose `branch:` frontmatter line names the branch
/// (SPEC-20261124-lane-branch-after-spec). The link is looked for only then, so a branch whose
/// name reads an existing SPEC is judged as before and reads no other file, and `specs_dir`
/// `None` looks at nothing.
pub fn resolve(branch: &str, specs_dir: Option<&Path>) -> Resolved {
    let named = resolve_by_name(branch, specs_dir);
    let Some(dir) = specs_dir else {
        return named;
    };
    match &named {
        // A name that reads no SPEC id is no SPEC branch: nothing to settle.
        Resolved::NotSpec => return named,
        Resolved::One(id) if spec_file_exists(dir, id) => return named,
        _ => {}
    }
    let mut linked = linked_specs(branch, dir);
    match linked.len() {
        0 => named,
        1 => Resolved::One(linked.remove(0)),
        _ => Resolved::Ambiguous(linked),
    }
}

/// Whether `<specs_dir>/<id>/spec.md` is a regular file (a link does not count).
fn spec_file_exists(specs_dir: &Path, id: &str) -> bool {
    std::fs::symlink_metadata(specs_dir.join(id).join("spec.md")).is_ok_and(|m| m.is_file())
}

/// A SPEC file is read for its `branch:` line up to this many bytes: the frontmatter is at the
/// top, and a hundred SPECs are read in one scan.
const LINK_READ_LIMIT: u64 = 16 * 1024;

/// The ids of the SPECs under `specs_dir` that name `branch` in a `branch:` frontmatter line,
/// sorted. Only the direct children of `specs_dir` are read and no link is followed: a child
/// that is not a real folder, a `spec.md` that is not a regular file, a folder name that is not a
/// SPEC id, an `id:` that differs from the folder name and a status other than `draft` or
/// `approved` are all left out.
fn linked_specs(branch: &str, specs_dir: &Path) -> Vec<String> {
    use std::io::Read;
    let Ok(entries) = std::fs::read_dir(specs_dir) else {
        return Vec::new();
    };
    let mut ids = Vec::new();
    for entry in entries.flatten() {
        // `DirEntry::file_type` does not follow a link: a link to a folder is not a folder here.
        if !entry.file_type().is_ok_and(|t| t.is_dir()) {
            continue;
        }
        let Some(name) = entry.file_name().to_str().map(str::to_string) else {
            continue;
        };
        if spec_doc::id_form(&name) != Some(IdForm::New) {
            continue;
        }
        let file = entry.path().join("spec.md");
        if !std::fs::symlink_metadata(&file).is_ok_and(|m| m.is_file()) {
            continue;
        }
        let Ok(handle) = crate::hook_guards::open_for_reading(&file) else {
            continue;
        };
        let mut bytes = Vec::new();
        if handle
            .take(LINK_READ_LIMIT)
            .read_to_end(&mut bytes)
            .is_err()
        {
            continue;
        }
        if names_branch(&String::from_utf8_lossy(&bytes), &name, branch) {
            ids.push(name);
        }
    }
    ids.sort();
    ids
}

/// Whether the frontmatter of `text` (the lines between the first two `---` lines) has the lines
/// `id: <id>`, `branch: <branch>` and a `status:` of `draft` or `approved`, each unindented.
fn names_branch(text: &str, id: &str, branch: &str) -> bool {
    let mut lines = text.lines();
    if lines.next().map(str::trim_end) != Some("---") {
        return false;
    }
    let (mut same_id, mut same_branch, mut live) = (false, false, false);
    for line in lines.take_while(|l| l.trim_end() != "---") {
        if let Some(v) = frontmatter_value(line, "id") {
            same_id = v == id;
        } else if let Some(v) = frontmatter_value(line, "branch") {
            same_branch = v == branch;
        } else if let Some(v) = frontmatter_value(line, "status") {
            live = matches!(v, "draft" | "approved");
        }
    }
    same_id && same_branch && live
}

/// The value of an unindented `<key>: <value>` line, without surrounding blanks or one pair of
/// quotes.
fn frontmatter_value<'a>(line: &'a str, key: &str) -> Option<&'a str> {
    let value = line.strip_prefix(key)?.strip_prefix(':')?.trim();
    let quoted = |q: char| value.len() >= 2 && value.starts_with(q) && value.ends_with(q);
    Some(if quoted('"') || quoted('\'') {
        &value[1..value.len() - 1]
    } else {
        value
    })
}

/// The SPEC the name of `branch` reads: the single candidate, or among several the one whose
/// folder exists in `specs_dir` when exactly one does. `specs_dir` `None` looks at no folder.
fn resolve_by_name(branch: &str, specs_dir: Option<&Path>) -> Resolved {
    let mut found = candidates(branch);
    match found.len() {
        0 => Resolved::NotSpec,
        1 => Resolved::One(found.remove(0)),
        _ => {
            if let Some(dir) = specs_dir {
                // The ids hold only `a-z`, `0-9` and `-`, so joining one cannot leave `dir`.
                let mut with_folder = found.iter().filter(|id| dir.join(id).is_dir());
                if let (Some(only), None) = (with_folder.next(), with_folder.next()) {
                    return Resolved::One(only.clone());
                }
            }
            Resolved::Ambiguous(found)
        }
    }
}

/// The two branch names of a new-form SPEC id: `(name_first, date_first)`, so
/// `SPEC-20261117-login-fix` gives `spec/login-fix-20261117` and `spec/20261117-login-fix`.
/// `None` for an id that is not of the form `SPEC-<yyyymmdd>-<slug>`.
pub fn names_of(id: &str) -> Option<(String, String)> {
    if spec_doc::id_form(id) != Some(IdForm::New) {
        return None;
    }
    let (date, slug) = id.strip_prefix("SPEC-")?.split_once('-')?;
    Some((format!("spec/{slug}-{date}"), format!("spec/{date}-{slug}")))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn ids(branch: &str) -> Vec<String> {
        candidates(branch)
    }

    #[test]
    fn both_forms_read_as_the_same_spec_id() {
        let want = vec!["SPEC-20261117-demo".to_string()];
        assert_eq!(ids("spec/demo-20261117"), want);
        assert_eq!(ids("spec/20261117-demo"), want);
        assert_eq!(
            ids("spec/login-fix-20261117"),
            vec!["SPEC-20261117-login-fix".to_string()]
        );
    }

    #[test]
    fn a_repair_suffix_is_read_off_before_the_name_is_read() {
        let want = vec!["SPEC-20261117-demo".to_string()];
        assert_eq!(ids("spec/demo-20261117-r2"), want);
        assert_eq!(ids("spec/demo-20261117-r12"), want);
        // Date first: the name as it stands is a reading too (a slug may end in `-r2`), so the
        // two readings wait for a SPEC folder to tell them apart.
        assert_eq!(
            ids("spec/20261117-demo-r2"),
            vec![
                "SPEC-20261117-demo-r2".to_string(),
                "SPEC-20261117-demo".to_string()
            ]
        );
        // `-r` without digits is part of the slug, not a suffix.
        assert_eq!(
            ids("spec/20261117-demo-r"),
            vec!["SPEC-20261117-demo-r".to_string()]
        );
    }

    #[test]
    fn a_name_that_reads_two_ways_gives_both_ids() {
        assert_eq!(
            ids("spec/20261117-demo-20261118"),
            vec![
                "SPEC-20261117-demo-20261118".to_string(),
                "SPEC-20261118-20261117-demo".to_string()
            ]
        );
    }

    #[test]
    fn the_date_is_eight_digits_with_no_calendar_check() {
        assert_eq!(ids("spec/20269999-x"), vec!["SPEC-20269999-x".to_string()]);
        assert_eq!(ids("spec/x-20269999"), vec!["SPEC-20269999-x".to_string()]);
    }

    #[test]
    fn a_long_slug_is_still_a_spec_branch() {
        // `moai spec new` caps a slug (48 characters) but the SPEC-first hook never did; a
        // branch with a longer slug must still be read, not slip past the rule.
        let slug = "a".repeat(60);
        let id = format!("SPEC-20261117-{slug}");
        assert_eq!(ids(&format!("spec/{slug}-20261117")), vec![id.clone()]);
        assert_eq!(ids(&format!("spec/20261117-{slug}")), vec![id]);
    }

    /// Audit of C-112 (note 3228, N1): the SPEC-first hook read a slug with a hyphen pair inside
    /// (`spec/20261117-a--b`) before the resolver; the resolver keeps that, in both forms.
    #[test]
    fn a_hyphen_pair_inside_a_slug_still_reads_as_a_spec_branch() {
        let id = vec!["SPEC-20261117-a--b".to_string()];
        assert_eq!(ids("spec/20261117-a--b"), id);
        assert_eq!(ids("spec/a--b-20261117"), id);
        assert!(ids("spec/20261117--b").is_empty());
        assert!(ids("spec/20261117-b-").is_empty());
    }

    #[test]
    fn names_that_are_not_spec_branches_read_as_no_id() {
        for branch in [
            "main",
            "spec/",
            "spec/Demo-20261117",
            "spec/demo-2026111",
            "spec/demo-",
            "spec/-20261117",
            "spec/20261117",
            "spec/20261117-",
            "spec/demo",
            "spec/demo--20261117",
            "spec/demo-20261117/sub",
            "spec/demo_20261117",
            "issue/demo-20261117",
            "demo-20261117",
            "spec/demo-20261117-",
        ] {
            assert!(ids(branch).is_empty(), "{branch}");
            assert_eq!(resolve(branch, None), Resolved::NotSpec, "{branch}");
        }
    }

    #[test]
    fn one_reading_resolves_with_no_folder_to_look_at() {
        assert_eq!(
            resolve("spec/demo-20261117", None),
            Resolved::One("SPEC-20261117-demo".to_string())
        );
        assert_eq!(
            resolve("spec/20261117-demo", None),
            Resolved::One("SPEC-20261117-demo".to_string())
        );
    }

    #[test]
    fn several_readings_are_settled_by_exactly_one_spec_folder() {
        let dir = tempfile::TempDir::new().unwrap();
        let specs = dir.path();
        let both = vec![
            "SPEC-20261117-demo-20261118".to_string(),
            "SPEC-20261118-20261117-demo".to_string(),
        ];
        let name = "spec/20261117-demo-20261118";
        assert_eq!(resolve(name, None), Resolved::Ambiguous(both.clone()));
        assert_eq!(
            resolve(name, Some(specs)),
            Resolved::Ambiguous(both.clone())
        );
        std::fs::create_dir(specs.join("SPEC-20261118-20261117-demo")).unwrap();
        assert_eq!(
            resolve(name, Some(specs)),
            Resolved::One("SPEC-20261118-20261117-demo".to_string())
        );
        std::fs::create_dir(specs.join("SPEC-20261117-demo-20261118")).unwrap();
        assert_eq!(resolve(name, Some(specs)), Resolved::Ambiguous(both));
    }

    #[test]
    fn a_repair_suffix_name_is_settled_by_the_spec_folder() {
        let dir = tempfile::TempDir::new().unwrap();
        let specs = dir.path();
        let name = "spec/20261117-demo-r2";
        assert!(matches!(resolve(name, Some(specs)), Resolved::Ambiguous(_)));
        std::fs::create_dir(specs.join("SPEC-20261117-demo")).unwrap();
        assert_eq!(
            resolve(name, Some(specs)),
            Resolved::One("SPEC-20261117-demo".to_string())
        );
    }

    #[test]
    fn a_missing_specs_folder_looks_like_no_evidence() {
        let dir = tempfile::TempDir::new().unwrap();
        let missing = dir.path().join("no-such");
        assert!(matches!(
            resolve("spec/20261117-demo-r2", Some(&missing)),
            Resolved::Ambiguous(_)
        ));
    }

    #[test]
    fn the_two_branch_names_of_an_id() {
        assert_eq!(
            names_of("SPEC-20261117-login-fix"),
            Some((
                "spec/login-fix-20261117".to_string(),
                "spec/20261117-login-fix".to_string()
            ))
        );
        assert_eq!(names_of("SPEC-AUTH-001"), None);
        assert_eq!(names_of("C-87"), None);
        assert_eq!(names_of(""), None);
    }

    #[test]
    fn the_names_of_an_id_read_back_as_that_id() {
        for id in [
            "SPEC-20261117-login-fix",
            "SPEC-20261119-a",
            "SPEC-20261101-x-y-z",
        ] {
            let (new, old) = names_of(id).unwrap();
            assert_eq!(resolve(&new, None), Resolved::One(id.to_string()), "{new}");
            assert_eq!(resolve(&old, None), Resolved::One(id.to_string()), "{old}");
        }
    }
}
