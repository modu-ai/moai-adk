export const meta = {
  name: "hns-oss-docs-run",
  description: "(user-owned) oss-docs harness Runner — README 4-locale + docs-site publishing pipeline. Phases: scope -> author (canonical locale) -> translate (parallel derived-locale fan-out) -> verify. Non-interactive only; publishing (commit/push/PR) stays orchestrator/human-gated.",
  version: "1.0.0",
};

// hns-oss-docs-run.js — Runner for the oss-docs user-owned harness.
//
// [USER-OWNED] hns- namespace. `moai update` preserves it. NOT distributed to
// user projects (never touches internal/template/templates/).
//
// HARD constraints:
//   (i)   This Runner MUST NOT call AskUserQuestion / mcp__askuser — a
//         dynamic-workflow script cannot prompt the user mid-run (asymmetric
//         boundary, agent-common-protocol.md § User Interaction Boundary).
//         Specialists return blocker reports; the orchestrator holds every
//         human gate BEFORE this Runner is launched.
//   (ii)  This Runner NEVER commits or pushes. No `git commit`, `git push`,
//         `gh pr` anywhere in any dispatched prompt. Publishing is
//         orchestrator/human-gated OUTSIDE this run (Vercel auto-deploys on
//         push, so an accidental push IS an accidental production deploy).
//   (iii) Determinism (dynamic-workflows.md): the script body MUST NOT call
//         Date.now() or Math.random(). Timestamps, if needed, are injected
//         via `args` or stamped onto results AFTER the run returns.
//
// Manifest (SSOT): .claude/commands/harness/oss-docs/manifest.json. The Runner
// dispatches each specialist per its declared `primitive` verbatim. All three
// specialists declare `isolation: "none"`; no worktree is created.

export const MANIFEST_PATH = ".claude/commands/harness/oss-docs/manifest.json";

// Mirrors manifest.sprint_contract (SSOT is the manifest; keep in sync).
export const SPRINT_CONTRACT = {
  dimensions: ["locale-parity", "build-clean", "style-compliance", "content-fidelity"],
  thresholds: {
    "locale-parity": 1.0,
    "build-clean": 1.0,
    "style-compliance": 0.95,
    "content-fidelity": 0.9,
  },
  must_pass: ["locale-parity", "build-clean"],
};

// Canonical-locale chain per hns-oss-docs-i18n-rules:
//   docs-site: ko canonical -> derive en, ja, zh (same PR)
//   README:    en canonical (README.md) -> derive ko, ja, zh (same PR)
export function derivedLocalesFor(scope) {
  const targets = [];
  if (scope === "readme-only" || scope === "both") {
    targets.push(
      { surface: "readme", locale: "ko", canonical: "en", file: "README.ko.md" },
      { surface: "readme", locale: "ja", canonical: "en", file: "README.ja.md" },
      { surface: "readme", locale: "zh", canonical: "en", file: "README.zh.md" }
    );
  }
  if (scope === "docs-only" || scope === "both") {
    targets.push(
      { surface: "docs-site", locale: "en", canonical: "ko", file: "docs-site/content/en/" },
      { surface: "docs-site", locale: "ja", canonical: "ko", file: "docs-site/content/ja/" },
      { surface: "docs-site", locale: "zh", canonical: "ko", file: "docs-site/content/zh/" }
    );
  }
  return targets;
}

// Group derived-locale targets so the fan-out spawns AT MOST 3 parallel
// workers (one per derived locale), each handling every surface for its
// locale — per the manifest locale-translator "One parallel worker per
// derived locale" contract.
export function groupByLocale(targets) {
  const byLocale = new Map();
  for (const t of targets) {
    if (!byLocale.has(t.locale)) byLocale.set(t.locale, []);
    byLocale.get(t.locale).push(t);
  }
  return Array.from(byLocale.entries()).map(([locale, items]) => ({ locale, items }));
}

export function parseJsonBlock(text, fallback) {
  if (typeof text !== "string") return fallback;
  const start = text.indexOf("{");
  const end = text.lastIndexOf("}");
  if (start === -1 || end === -1 || end <= start) return fallback;
  try {
    return JSON.parse(text.slice(start, end + 1));
  } catch {
    return fallback;
  }
}

export async function run({ agent, args }) {
  const task = (args && args.task) ? String(args.task) : "Maintain README 4-locale set and docs-site per the SSOT redesign report.";
  const dry = Boolean(args && args.dry);

  // ---------------------------------------------------------------- Phase 1
  // Scope — one read-only agent classifies the task.
  const scopeRaw = await agent({
    agentType: "Explore",
    effort: "low",
    isolation: "none",
    label: "oss-docs:scope",
    prompt:
      `Classify this oss-docs task and enumerate target files. Task: "${task}". ` +
      `Read .moai/reports/readme-docs-redesign-20260713.md if it exists for context. ` +
      `Surfaces: README 4-locale set (README.md en-canonical + README.ko.md/README.ja.md/README.zh.md) ` +
      `and the Hugo geekdoc docs-site at docs-site/ (config hugo.toml, content/{ko,en,ja,zh}, ko-canonical). ` +
      `Return ONLY a JSON object: {"scope": "readme-only"|"docs-only"|"both", "target_files": ["..."], "rationale": "..."}. ` +
      `Do NOT modify any file.`,
  });
  const scoped = parseJsonBlock(scopeRaw, { scope: "both", target_files: [], rationale: "scope parse fallback" });
  const scope = ["readme-only", "docs-only", "both"].includes(scoped.scope) ? scoped.scope : "both";

  // ---------------------------------------------------------------- Phase 2
  // Author — canonical-locale authoring (content-author specialist).
  // Canonical: en for README (README.md), ko for docs-site pages.
  const authorResult = await agent({
    agentType: "hns-oss-docs-content-author-specialist",
    effort: "high",
    isolation: "none",
    label: "oss-docs:author",
    prompt:
      `Author/rewrite the CANONICAL-locale source content only. Scope: ${scope}. ` +
      `Target files (from scope phase): ${JSON.stringify(scoped.target_files)}. ` +
      `Canonical surfaces: README.md (English) when scope includes README; ` +
      `docs-site/content/ko/ pages when scope includes docs. Do NOT translate — ` +
      `derived locales are produced by a later phase. ` +
      `At start, invoke Skill("hns-oss-docs-i18n-rules") and Skill("hns-oss-docs-readme-sync"). ` +
      `Consume .moai/reports/readme-docs-redesign-20260713.md as the SSOT design reference. ` +
      `HARD: Mermaid TD-only; icon shortcodes over body emoji; URL blacklist ` +
      `(only adk.mo.ai.kr is valid); never run git commit/push/gh pr. ` +
      (dry ? `DRY RUN: report the planned edits as a diff summary, write NOTHING. ` : ``) +
      `If docs-site navigation/menu/redirect changes are needed, list them for the ` +
      `structure-curator instead of editing shared config yourself. ` +
      `Return a markdown report: files written, sections changed, curator handoff list.`,
  });

  // ---------------------------------------------------------------- Phase 3
  // Translate — parallel fan-out, one agent per derived locale (3 max).
  const localeGroups = groupByLocale(derivedLocalesFor(scope));
  const translateResults = await Promise.all(
    localeGroups.map((group) =>
      agent({
        agentType: "hns-oss-docs-locale-translator-specialist",
        effort: "medium",
        isolation: "none",
        label: `oss-docs:translate:${group.locale}`,
        prompt:
          `Derive the ${group.locale} locale from the canonical output of the author phase. ` +
          `Targets for this locale: ${JSON.stringify(group.items)}. ` +
          `Author-phase report (canonical source of truth):\n${authorResult}\n` +
          `At start, invoke Skill("hns-oss-docs-i18n-rules") and Skill("hns-oss-docs-readme-sync"). ` +
          `HARD: preserve facts/figures/code blocks/icon shortcodes/Mermaid direction verbatim; ` +
          `apply per-locale emphasis-marker spacing; URL blacklist (only adk.mo.ai.kr); ` +
          `same-PR 4-locale obligation — every canonical change you were handed MUST land in this locale. ` +
          (dry ? `DRY RUN: report planned edits only, write NOTHING. ` : ``) +
          `Never run git commit/push/gh pr. Return: files written + per-file section-count parity note.`,
      })
    )
  );

  // ---------------------------------------------------------------- Phase 4
  // Verify — run the verify recipe and score the sprint_contract dimensions.
  const verifyRaw = await agent({
    agentType: "hns-oss-docs-content-author-specialist",
    effort: "medium",
    isolation: "none",
    label: "oss-docs:verify",
    prompt:
      `Run the oss-docs verify recipe READ-ONLY (no file writes, no git). ` +
      `At start, invoke Skill("hns-oss-docs-verify") and execute its inlined checks: ` +
      `(1) cd docs-site && hugo --minify --gc — must complete warning-free; ` +
      `(2) test -f docs-site/public/sitemap.xml; ` +
      `(3) URL-blacklist grep (docs.moai-ai.dev|adk.moai.com|adk.moai.kr) over docs-site/content and README*.md; ` +
      `(4) Mermaid LR/RL direction grep over docs-site/content; ` +
      `(5) 4-locale file-existence + section-count parity across content/{ko,en,ja,zh} and README 4-file '^## ' heading counts; ` +
      `(6) body-emoji scan. ` +
      `Score each sprint_contract dimension in [0,1]: locale-parity (parity checks), ` +
      `build-clean (hugo build + sitemap), style-compliance (Mermaid/emoji/emphasis greps), ` +
      `content-fidelity (blacklist + facts preserved vs author report). ` +
      `Return ONLY a JSON object: {"locale-parity": n, "build-clean": n, "style-compliance": n, "content-fidelity": n, "notes": "..."}.`,
  });
  const verify = parseJsonBlock(verifyRaw, {
    "locale-parity": 0,
    "build-clean": 0,
    "style-compliance": 0,
    "content-fidelity": 0,
    notes: "verify parse fallback — treat as FAIL",
  });

  const mustPassOk = SPRINT_CONTRACT.must_pass.every(
    (dim) => Number(verify[dim]) >= SPRINT_CONTRACT.thresholds[dim]
  );

  return {
    manifest: MANIFEST_PATH,
    scope,
    dry,
    files_changed: scoped.target_files,
    author_report: authorResult,
    translate_reports: translateResults,
    verify: {
      "locale-parity": Number(verify["locale-parity"]) || 0,
      "build-clean": Number(verify["build-clean"]) || 0,
      "style-compliance": Number(verify["style-compliance"]) || 0,
      "content-fidelity": Number(verify["content-fidelity"]) || 0,
    },
    must_pass_ok: mustPassOk,
    note:
      "Non-interactive pipeline only. Publishing (commit, push, PR, Vercel deploy) " +
      "is orchestrator/human-gated OUTSIDE this Runner — this run never commits or pushes.",
  };
}
