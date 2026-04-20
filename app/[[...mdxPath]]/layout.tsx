import { getPageMap } from "nextra/page-map";
import { Footer, Layout, Navbar } from "nextra-theme-docs";

import LanguageSelector from "../../components/LanguageSelector";

const TOC_TITLES: Record<string, string> = {
  ko: "이 페이지에서",
  en: "On this page",
  ja: "このページの目次",
  zh: "本页内容",
};

const SUPPORTED_LOCALES = ["ko", "en", "ja", "zh"];

type LayoutProps = {
  children: React.ReactNode;
  params: Promise<{ mdxPath?: string[] }>;
};

export default async function LocaleLayout({ children, params }: LayoutProps) {
  const { mdxPath } = await params;
  const firstSegment = mdxPath?.[0] ?? "ko";
  const locale = SUPPORTED_LOCALES.includes(firstSegment) ? firstSegment : "ko";

  const pageMap = await getPageMap(`/${locale}`);

  const navbar = (
    <Navbar
      logo={
        <span
          style={{
            fontSize: "1.5rem",
            fontWeight: "bold",
            display: "flex",
            alignItems: "center",
            gap: "8px",
          }}
        >
          <span role="img" aria-label="moai">
            🗿
          </span>{" "}
          MoAI-ADK
        </span>
      }
    >
      <a
        href="https://github.com/modu-ai/moai-adk"
        target="_blank"
        rel="noopener noreferrer"
        className="x:inline-flex x:items-center x:shrink-0"
      >
        {/* biome-ignore lint/performance/noImgElement: External badge image */}
        <img
          src="https://img.shields.io/github/stars/modu-ai/moai-adk?style=social"
          alt="GitHub stars"
          loading="lazy"
        />
      </a>
      <LanguageSelector />
    </Navbar>
  );

  const footer = <Footer>MIT License © 2025 MoAI-ADK Contributors</Footer>;

  return (
    <Layout
      key={locale}
      pageMap={pageMap}
      navbar={navbar}
      footer={footer}
      docsRepositoryBase="https://github.com/modu-ai/moai-adk/tree/main/docs"
      search={false}
      sidebar={{
        defaultMenuCollapseLevel: 1,
        toggleButton: true,
        autoCollapse: true,
      }}
      toc={{
        backToTop: true,
        title: TOC_TITLES[locale] ?? "On this page",
        float: true,
      }}
      navigation={{ prev: true, next: true }}
    >
      {children}
    </Layout>
  );
}
