import type { MetadataRoute } from "next";

export default function sitemap(): MetadataRoute.Sitemap {
  const baseUrl = "https://adk.mo.ai.kr";
  const lastModified = new Date();

  // Define locales and their paths
  const locales = [
    { locale: "", lang: "ko" },
    { locale: "/en", lang: "en" },
    { locale: "/zh", lang: "zh" },
    { locale: "/ja", lang: "ja" },
  ];

  // Main routes
  const mainRoutes = [
    "",
    "/getting-started/installation",
    "/getting-started/introduction",
    "/getting-started/update",
    "/core-concepts/what-is-moai-adk",
    "/core-concepts/spec-based-dev",
    "/core-concepts/ddd",
    "/core-concepts/trust-5",
    "/workflow-commands/moai-plan",
    "/workflow-commands/moai-run",
    "/workflow-commands/moai-sync",
    "/workflow-commands/moai-project",
    "/utility-commands/moai",
    "/utility-commands/moai-fix",
    "/utility-commands/moai-loop",
    "/utility-commands/moai-feedback",
    "/claude-code/index",
    "/claude-code/quickstart",
    "/claude-code/settings",
    "/claude-code/skills",
    "/claude-code/sub-agents",
    "/claude-code/extensions",
    "/claude-code/chrome",
    "/advanced/agent-guide",
    "/advanced/skill-guide",
    "/advanced/builder-agents",
    "/advanced/hooks-guide",
    "/advanced/settings-json",
    "/advanced/mcp-servers",
    "/advanced/stitch-guide",
    "/moai-rank/index",
    "/moai-rank/dashboard",
    "/moai-rank/faq",
    "/worktree/index",
    "/worktree/faq",
  ];

  // Generate sitemap entries for all locales
  const sitemapEntries: MetadataRoute.Sitemap = [];

  for (const { locale } of locales) {
    for (const route of mainRoutes) {
      sitemapEntries.push({
        url: `${baseUrl}${locale}${route}`,
        lastModified,
        changeFrequency: "weekly",
        priority: route === "" ? 1 : 0.8,
        alternates: {
          languages: Object.fromEntries(
            locales.map((l) => [l.lang, `${baseUrl}${l.locale}${route}`]),
          ),
        },
      });
    }
  }

  return sitemapEntries;
}
