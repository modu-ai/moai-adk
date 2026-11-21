//! The output style banners carry the moai mark: `<emoji> MoAI 🗿 <name> ─…` (user instruction
//! 2026-11-21). Nothing shipped keeps the old `MoAI-Easy ★` or `MoAI ★` wordings, no bare `★`
//! banner marker remains in the shipped styles, and the style references pin the new form.
//! Walks the shipped tree like the other repo scans; listed in tools/repo_scans.json.

mod style_banner {
    use std::{
        fs,
        path::{Path, PathBuf},
    };

    const ROOT: &str = env!("CARGO_MANIFEST_DIR");
    const OLD_BANNERS: [&str; 2] = ["MoAI-Easy ★", "MoAI ★"];

    fn files(dir: &Path, out: &mut Vec<PathBuf>) {
        for e in fs::read_dir(dir).unwrap() {
            let p = e.unwrap().path();
            if p.is_dir() {
                files(&p, out);
            } else {
                out.push(p);
            }
        }
    }

    #[test]
    fn shipped_paths_keep_no_old_banner_wording() {
        let root = Path::new(ROOT);
        // `templates/` holds the shipped style templates; `skills/moai/` is the shipped skill
        // package whose references state the banner.
        let mut shipped = Vec::new();
        for dir in ["templates", "skills/moai"] {
            files(&root.join(dir), &mut shipped);
        }
        assert!(
            shipped.len() > 30,
            "scan found too few files: {}",
            shipped.len()
        );
        let mut offenders = Vec::new();
        for path in shipped {
            let Ok(text) = fs::read_to_string(&path) else {
                continue;
            };
            let mut hits: Vec<&str> = OLD_BANNERS
                .into_iter()
                .filter(|old| text.contains(old))
                .collect();
            if !hits.is_empty() {
                offenders.push(format!("{}: {}", path.display(), hits.join(", ")));
            }
        }
        assert!(
            offenders.is_empty(),
            "old banner wording or bare star marker in shipped paths:\n{}",
            offenders.join("\n")
        );
    }

    #[test]
    fn the_style_references_pin_the_new_banner_form() {
        for rel in [
            "skills/moai/references/home/output-style.md",
            "templates/claude/output-styles/moai-easy.md",
            "templates/claude/output-styles/moai.md",
            "templates/claude/output-styles/moai-learn.md",
        ] {
            let text = fs::read_to_string(Path::new(ROOT).join(rel)).unwrap();
            assert!(text.contains("MoAI 🗿"), "{rel} lost the new banner mark");
            // every banner mark in a shipped style is the moai emoji: a bare star left in a
            // style file is a banner the change missed
            assert!(!text.contains('★'), "{rel} still carries a star marker");
        }
    }
}
