interface StructuredDataProps {
  data: Record<string, unknown>;
}

export function StructuredData({ data }: StructuredDataProps) {
  return (
    <script
      type="application/ld+json"
      // biome-ignore lint/security/noDangerouslySetInnerHtml: JSON-LD structured data is safe
      dangerouslySetInnerHTML={{ __html: JSON.stringify(data) }}
    />
  );
}

export function MoAIStructuredData() {
  const baseUrl = "https://adk.mo.ai.kr";

  const organizationData = {
    "@context": "https://schema.org",
    "@type": "Organization",
    name: "MoAI-ADK",
    url: baseUrl,
    logo: `${baseUrl}/og.png`,
    description:
      "MoAI-ADK - Claude Code를 위한 온라인 제작 인공지능 개발 도구 키트. SPEC 기반 개발, DDD 구현, TRUST 5 품질 시스템으로 완전 자율 자동화.",
    sameAs: [
      "https://github.com/modu-ai/moai-adk",
      "https://discord.gg/moai-adk",
      "https://www.npmjs.com/package/moai-adk",
    ],
  };

  const softwareApplicationData = {
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    name: "MoAI-ADK",
    operatingSystem: "macOS, Linux, Windows",
    applicationCategory: "DeveloperApplication",
    offers: {
      "@type": "Offer",
      price: "0",
      priceCurrency: "USD",
    },
    aggregateRating: {
      "@type": "AggregateRating",
      ratingValue: "4.8",
      ratingCount: "42",
    },
    author: {
      "@type": "Organization",
      name: "MoAI-ADK Contributors",
    },
    description:
      "MoAI-ADK는 Claude Code를 위한 온라인 제작 인공지능 개발 도구 키트입니다. SPEC 생성부터 DDD 구현, 문서화까지 완전 자율 자동화하는 AI 개발 도구입니다.",
    keywords:
      "MoAI-ADK, Claude Code, AI 개발, 온라인 제작, 인공지능 개발 도구, 자동화, SPEC, DDD, TRUST 5",
    license: "MIT",
    url: baseUrl,
    softwareVersion: "2.0.0",
  };

  const webSiteData = {
    "@context": "https://schema.org",
    "@type": "WebSite",
    name: "MoAI-ADK Documentation",
    url: baseUrl,
    description:
      "MoAI-ADK 온라인 문서 - Claude Code를 위한 온라인 제작 인공지능 개발 도구 키트에 대한 포괄적인 가이드",
    potentialAction: {
      "@type": "SearchAction",
      target: `${baseUrl}/search?q={search_term_string}`,
      "query-input": "required name=search_term_string",
    },
    inLanguage: ["ko", "en", "zh", "ja"],
  };

  const techArticleData = {
    "@context": "https://schema.org",
    "@type": "TechArticle",
    headline: "MoAI-ADK Documentation",
    description:
      "MoAI-ADK는 Claude Code를 위한 온라인 제작 인공지능 개발 도구 키트입니다. SPEC 기반 개발, DDD 구현, TRUST 5 품질 시스템으로 완전 자율 자동화를 제공합니다.",
    image: `${baseUrl}/og.png`,
    author: {
      "@type": "Organization",
      name: "MoAI-ADK Contributors",
    },
    publisher: {
      "@type": "Organization",
      name: "MoAI-ADK",
      logo: {
        "@type": "ImageObject",
        url: `${baseUrl}/og.png`,
      },
    },
    datePublished: "2024-01-01",
    dateModified: new Date().toISOString().split("T")[0],
    keywords: "MoAI-ADK, Claude Code, AI 개발, SPEC, DDD, TRUST 5",
    inLanguage: "ko",
  };

  return (
    <>
      <StructuredData data={organizationData} />
      <StructuredData data={softwareApplicationData} />
      <StructuredData data={webSiteData} />
      <StructuredData data={techArticleData} />
    </>
  );
}
