//! SPEC-20261122-output-style-moai-easy: the shipped MoAI-Easy text names only Claude Code, and
//! the installer takes a style file that equals the current render or an earlier shipped render
//! as MoAI's own while a real user edit stays a conflict. Against a temporary HOME only.
//!
//! The three earlier renders are kept as fixtures (`tests/fixtures/home-style/render-1.md` to
//! `render-3.md`), rebuilt from the commits of the text history; their sha256 are the ones of
//! the SPEC decision D-4.

mod common;
use moai_cli::home_style::{apply, plan, render_style_file, revert, Plan, SOURCE};
use std::fs;
use std::path::{Path, PathBuf};

const PERM: &str = "test-permission";
const STYLE: &str = ".claude/output-styles/MoAI-Easy.md";
/// The record of what MoAI last wrote, below the state folder of a HOME without a marker.
const RECORD: &str = ".moai-cli/home-style/claude.json";
const RENDERS: [&str; 3] = ["render-1.md", "render-2.md", "render-3.md"];

fn package() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
}

fn fixture(name: &str) -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("tests/fixtures/home-style")
            .join(name),
    )
    .unwrap()
}

fn temp_home() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::tempdir().unwrap();
    let path = fs::canonicalize(dir.path()).unwrap();
    (dir, path)
}

fn src() -> String {
    fs::read_to_string(package().join(SOURCE)).unwrap()
}

fn write(home: &Path, rel: &str, bytes: &str) {
    let p = home.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, bytes).unwrap();
}

fn record_naming(text: &str) -> String {
    format!(
        "{{\n  \"schema_version\": 1,\n  \"owned_sha256\": \"{}\"\n}}\n",
        moai_cli::digest(text.as_bytes())
    )
}

fn style_action(p: &Plan) -> String {
    p.targets
        .iter()
        .find(|t| t.kind == "style_file")
        .map(|t| t.action.clone())
        .unwrap()
}

/// A temporary package holding only what a plan reads: `SKILL.md` and the style source.
fn temp_package(source: &str) -> tempfile::TempDir {
    let pkg = tempfile::tempdir().unwrap();
    fs::write(pkg.path().join("SKILL.md"), "package\n").unwrap();
    let p = pkg.path().join(SOURCE);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(&p, source).unwrap();
    pkg
}

/// AC-1 (REQ-1, REQ-2, REQ-3): the shipped text.
#[test]
fn the_shipped_source_names_only_claude_code_and_keeps_the_picked_sections() {
    let s = src();
    for word in ["Codex", "every coding tool", "Claude Code, Codex"] {
        assert!(!s.contains(word), "{word}");
    }
    for item in [
        "## Voice",
        "## Rhythm",
        "## Size",
        "## Banner",
        "## Progress board",
        "## Format",
        "four steps",
        "Understand",
        "Never call something done without evidence",
        "A simple question gets a short plain answer",
        "`<emoji> MoAI 🗿 <label> ─…`",
        "46 `─`",
        "▓▓▓▓▓▓░░░░",
    ] {
        assert!(s.contains(item), "{item}");
    }
    // The star marker and the old banner wording are gone.
    assert!(!s.contains('★'), "a star marker is left");
    assert!(!s.contains("MoAI-Easy ★"));
}

/// AC-3 (REQ-4): the description is the one of the SPEC, with no double quote inside.
#[test]
fn the_rendered_description_has_no_double_quote_and_names_the_banner() {
    let out = render_style_file(&src());
    let line = out
        .lines()
        .find(|l| l.starts_with("description: "))
        .unwrap();
    assert_eq!(
        line,
        "description: \"Plain words, four steps (understand, plan, do, check), evidence before done, MoAI 🗿 banners and a progress board.\""
    );
    assert!(!line["description: \"".len()..line.len() - 1].contains('"'));
}

/// The sha256 of the three earlier renders, kept in the SPEC (D-4), match the fixtures.
#[test]
fn the_fixtures_are_the_three_renders_of_the_spec() {
    let want = [
        "93e53607dd194da86abfbe7ea993303fff1d17af1a1a1ae6868788938ed785d2",
        "d4e1851a65eb0ed21d16de888925dc6ae8a7bcebcae45973e1b8cce884d75e67",
        "690c534220f9bb1614e27d608e56306da973e2fe7b53e2bc1a4d42d7d2b9c65e",
    ];
    for (name, sha) in RENDERS.iter().zip(want) {
        assert_eq!(
            moai_cli::digest(fixture(name).as_bytes()),
            format!("sha256:{sha}"),
            "{name}"
        );
    }
}

/// The sha256 of the render of the shipped text today. A change of the template or of the
/// frontmatter makes this test fail until the outgoing render is appended to
/// `EARLIER_RENDERS` and this constant names the new one (AC-7, REQ-8).
const CURRENT_RENDER_SHA256: &str =
    "a31d966b0486f501839121832eb6a99d3fc718996b22a901e9e17884a0ab369e";

/// AC-7 (REQ-8): the list holds 64 hex digits each, all different, equal to the fixtures in
/// order, and never the current render, which is pinned here.
#[test]
fn the_earlier_renders_are_pinned_and_never_hold_the_current_one() {
    let list = moai_cli::home_style::EARLIER_RENDERS;
    for entry in list {
        assert_eq!(entry.len(), 64, "{entry}");
        assert!(entry
            .bytes()
            .all(|b| b.is_ascii_hexdigit() && !b.is_ascii_uppercase()));
    }
    let mut sorted = list.to_vec();
    sorted.sort_unstable();
    sorted.dedup();
    assert_eq!(sorted.len(), list.len(), "two entries are the same text");
    for (entry, name) in list.iter().zip(RENDERS) {
        assert_eq!(
            moai_cli::digest(fixture(name).as_bytes()),
            format!("sha256:{entry}"),
            "{name}"
        );
    }
    let current = moai_cli::digest(render_style_file(&src()).as_bytes());
    assert_eq!(
        current,
        format!("sha256:{CURRENT_RENDER_SHA256}"),
        "the render of the shipped text changed: append the outgoing render to EARLIER_RENDERS \
         (with a fixture), then pin the new one here"
    );
    assert!(!list.contains(&CURRENT_RENDER_SHA256));
}

/// AC-4 (REQ-5): a file equal to the current render is taken as owned although the record names
/// an earlier text or no text; the bytes stay and the record follows, so the next template
/// change updates it.
#[test]
fn a_style_file_equal_to_the_current_text_is_taken_as_owned() {
    let current = render_style_file(&src());
    for record in [
        Some(fixture("render-1.md")),
        Some(fixture("render-3.md")),
        None,
    ] {
        let (_t, home) = temp_home();
        write(&home, STYLE, &current);
        if let Some(r) = &record {
            write(&home, RECORD, &record_naming(r));
        }
        let p = plan(&home, &package()).unwrap();
        assert_eq!(style_action(&p), "merge", "record {:?}", record.is_some());
        apply(&home, &package(), &p.plan_digest, PERM).unwrap();
        assert_eq!(fs::read_to_string(home.join(STYLE)).unwrap(), current);
        assert_eq!(
            fs::read_to_string(home.join(RECORD)).unwrap(),
            record_naming(&current)
        );
        // Owned now: a changed template is an update, never a conflict.
        let next = temp_package(&format!("{}One more line.\n", src()));
        let p = plan(&home, next.path()).unwrap();
        assert_eq!(style_action(&p), "merge");
        // And an unchanged template leaves it alone.
        let p = plan(&home, &package()).unwrap();
        assert_eq!(style_action(&p), "unchanged");
    }
}

/// AC-5 (REQ-6): a file equal to any earlier shipped render is updated to the current one.
#[test]
fn a_style_file_equal_to_an_earlier_shipped_text_is_updated() {
    let current = render_style_file(&src());
    for name in RENDERS {
        for record in [None, Some("render-1.md")] {
            let (_t, home) = temp_home();
            write(&home, STYLE, &fixture(name));
            if let Some(r) = record {
                write(&home, RECORD, &record_naming(&fixture(r)));
            }
            let p = plan(&home, &package()).unwrap();
            assert_eq!(style_action(&p), "merge", "{name} {record:?}");
            apply(&home, &package(), &p.plan_digest, PERM).unwrap();
            assert_eq!(
                fs::read_to_string(home.join(STYLE)).unwrap(),
                current,
                "{name} {record:?}"
            );
        }
    }
}

/// AC-6 (REQ-7): a user edit on top of the current render, an earlier render or the text the
/// record names stays a conflict and keeps every byte.
#[test]
fn a_real_user_edit_stays_a_conflict_on_every_known_text() {
    let mut bases = vec![render_style_file(&src())];
    bases.extend(RENDERS.iter().map(|n| fixture(n)));
    for base in bases {
        let (_t, home) = temp_home();
        let edited = format!("{base}my own line\n");
        write(&home, STYLE, &edited);
        write(&home, RECORD, &record_naming(&base));
        let p = plan(&home, &package()).unwrap();
        assert_eq!(style_action(&p), "conflict");
        apply(&home, &package(), &p.plan_digest, PERM).unwrap();
        assert_eq!(fs::read_to_string(home.join(STYLE)).unwrap(), edited);
    }
    // A text that is none of them, with no record, is a conflict too.
    let (_t, home) = temp_home();
    write(&home, STYLE, "---\nname: MoAI-Easy\n---\n\nmy own style\n");
    let p = plan(&home, &package()).unwrap();
    assert_eq!(style_action(&p), "conflict");
}

/// AC-8 (REQ-9): reverting an adoption or a replacement gives back the before bytes and the
/// record as it was (an absent record stays absent).
#[test]
fn reverting_an_adopted_or_replaced_style_file_gives_back_bytes_and_record() {
    let current = render_style_file(&src());
    // Adopt: the file is the current render, the record names an earlier text.
    let (_t, home) = temp_home();
    write(&home, STYLE, &current);
    let before_record = record_naming(&fixture("render-1.md"));
    write(&home, RECORD, &before_record);
    let p = plan(&home, &package()).unwrap();
    let id = apply(&home, &package(), &p.plan_digest, PERM).unwrap()["receipt"]
        .as_str()
        .unwrap()
        .to_string();
    assert_eq!(
        fs::read_to_string(home.join(RECORD)).unwrap(),
        record_naming(&current)
    );
    revert(&home, &id, PERM).unwrap();
    assert_eq!(fs::read_to_string(home.join(STYLE)).unwrap(), current);
    assert_eq!(
        fs::read_to_string(home.join(RECORD)).unwrap(),
        before_record
    );
    // Replace: the file is an earlier render and there was no record.
    let (_t, home) = temp_home();
    let earlier = fixture("render-2.md");
    write(&home, STYLE, &earlier);
    let p = plan(&home, &package()).unwrap();
    let id = apply(&home, &package(), &p.plan_digest, PERM).unwrap()["receipt"]
        .as_str()
        .unwrap()
        .to_string();
    assert_eq!(fs::read_to_string(home.join(STYLE)).unwrap(), current);
    revert(&home, &id, PERM).unwrap();
    assert_eq!(fs::read_to_string(home.join(STYLE)).unwrap(), earlier);
    assert!(
        !home.join(RECORD).exists()
            || !fs::read_to_string(home.join(RECORD))
                .unwrap()
                .contains(&moai_cli::digest(current.as_bytes())),
        "the record still names the text the revert took away"
    );
}

/// AC-9 (REQ-10): every target is below the HOME and a project folder with its own
/// `.claude/output-styles` is never written.
#[test]
fn the_style_is_never_written_below_a_project_folder() {
    let (_t, home) = temp_home();
    let (_p, project) = home_like_project();
    let before = tree(&project);
    let p = plan(&home, &package()).unwrap();
    for t in &p.targets {
        assert!(
            Path::new(&t.path).starts_with(&home),
            "{} is not below the HOME",
            t.path
        );
        assert!(
            !Path::new(&t.path).starts_with(&project),
            "{} is below the project",
            t.path
        );
    }
    apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(tree(&project), before);
}

fn home_like_project() -> (tempfile::TempDir, PathBuf) {
    let (t, path) = temp_home();
    write(&path, ".claude/output-styles/MoAI-Easy.md", "project own\n");
    write(&path, ".claude/settings.json", "{}\n");
    (t, path)
}

fn tree(root: &Path) -> Vec<(String, String)> {
    let mut out = Vec::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(dir) = stack.pop() {
        for e in fs::read_dir(&dir).unwrap() {
            let p = e.unwrap().path();
            if p.is_dir() {
                stack.push(p);
            } else {
                out.push((
                    p.strip_prefix(root).unwrap().display().to_string(),
                    moai_cli::digest(&fs::read(&p).unwrap()),
                ));
            }
        }
    }
    out.sort();
    out
}
