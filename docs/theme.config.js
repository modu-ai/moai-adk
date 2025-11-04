export default {
  github: "https://github.com/modu-ai/moai-adk",
  docsRepositoryBase: "https://github.com/modu-ai/moai-adk/tree/main/docs",
  titleSuffix: " – MoAI-ADK Docs",
  logo: (
    <>
      <span className="mr-2 font-extrabold hidden md:inline text-blue-500">MoAI-ADK</span>
      <span className="text-gray-600 font-normal hidden md:inline">Agentic Development Kit</span>
    </>
  ),
  head: (
    <>
      <meta name="msapplication-TileColor" content="#050816" />
      <meta name="theme-color" content="#050816" />
      <meta name="viewport" content="width=device-width, initial-scale=1.0" />
      <meta httpEquiv="Content-Language" content="ko" />
      <meta name="description" content="MoAI-ADK: SPEC → TDD → 코드 → 문서를 자동화하는 한국어 문서" />
      <meta name="og:description" content="MoAI-ADK: SPEC → TDD → 코드 → 문서를 자동화하는 한국어 문서" />
      <meta name="twitter:card" content="summary_large_image" />
      <meta name="twitter:image" content="https://github.com/modu-ai/moai-adk/raw/main/docs/public/og.png" />
      <meta name="twitter:site:domain" content="moai-adk.vercel.app" />
      <meta name="twitter:url" content="https://moai-adk.vercel.app" />
      <meta name="og:title" content="MoAI-ADK Agentic Development Kit" />
      <meta name="og:image" content="https://github.com/modu-ai/moai-adk/raw/main/docs/public/og.png" />
      <meta name="apple-mobile-web-app-title" content="MoAI-ADK" />
    </>
  ),
  search: true,
  prevLinks: true,
  nextLinks: true,
  footer: true,
  footerEditLink: "GitHub에서 이 페이지 수정하기",
  footerText: <>MIT {new Date().getFullYear()} © Modu-AI.</>,
  unstable_faviconGlyph: "🤖",
};
