import { Analytics } from "@vercel/analytics/next";

const config = {
  defaultTheme: "light",
  logo: {
    text: "🗿 MoAI-ADK",
  },
  i18n: [
    {
      locale: "ko",
      name: "한국어",
      title: "MoAI-ADK 문서",
      description:
        "MoAI-ADK 온라인 문서 - Claude Code를 위한 온라인 제작 인공지능 개발 도구 키트. SPEC 기반 개발, DDD 구현, TRUST 5 품질 시스템으로 완전 자율 자동화.",
      lang: "ko-KR",
      dir: "ltr",
      theme: {
        toc: {
          title: "이 페이지에서",
        },
        editLink: {
          content: "이 페이지 수정 →",
        },
        feedback: {
          content: "질문이 있으신가요? 피드백을 남겨주세요!",
          useLink: "https://github.com/modu-ai/moai-adk/issues/new",
        },
        footer: {
          content: "MIT License © 2025 MoAI-ADK Contributors",
        },
      },
    },
    {
      locale: "en",
      name: "English",
      title: "MoAI-ADK Documentation",
      description:
        "MoAI-ADK online documentation - Agentic Development Kit for Claude Code. Fully automated development with SPEC-based development, DDD implementation, and TRUST 5 quality system.",
      lang: "en-US",
      dir: "ltr",
      theme: {
        toc: {
          title: "On this page",
        },
        editLink: {
          content: "Edit this page →",
        },
        feedback: {
          content: "Have a question? Leave your feedback!",
          useLink: "https://github.com/modu-ai/moai-adk/issues/new",
        },
        footer: {
          content: "MIT License © 2025 MoAI-ADK Contributors",
        },
      },
    },
    {
      locale: "zh",
      name: "简体中文",
      title: "MoAI-ADK 文档",
      description:
        "MoAI-ADK 在线文档 - Claude Code 的 AI 开发工具包。通过 SPEC 驱动开发、DDD 实现和 TRUST 5 质量系统实现完全自动化。",
      lang: "zh-CN",
      dir: "ltr",
      theme: {
        toc: {
          title: "本页内容",
        },
        editLink: {
          content: "编辑此页面 →",
        },
        feedback: {
          content: "有问题？留下您的反馈！",
          useLink: "https://github.com/modu-ai/moai-adk/issues/new",
        },
        footer: {
          content: "MIT License © 2025 MoAI-ADK Contributors",
        },
      },
    },
    {
      locale: "ja",
      name: "日本語",
      title: "MoAI-ADK ドキュメント",
      description:
        "MoAI-ADK オンラインドキュメント - Claude Code 用 AI 開発キット。SPEC ベース開発、DDD 実装、TRUST 5 品質システムで完全自動化。",
      lang: "ja-JP",
      dir: "ltr",
      theme: {
        toc: {
          title: "このページの目次",
        },
        editLink: {
          content: "このページを編集 →",
        },
        feedback: {
          content: "ご質問がありますか？フィードバックをお寄せください！",
          useLink: "https://github.com/modu-ai/moai-adk/issues/new",
        },
        footer: {
          content: "MIT License © 2025 MoAI-ADK Contributors",
        },
      },
    },
  ],
  project: {
    link: (
      <div className="x:flex x:items-center x:gap-2">
        <a
          key="github-stars"
          href="https://github.com/modu-ai/moai-adk"
          target="_blank"
          rel="noopener noreferrer"
        >
          {/* biome-ignore lint/performance/noImgElement: External badge image */}
          <img
            src="https://img.shields.io/github/stars/modu-ai/moai-adk?style=social"
            alt="GitHub stars"
            className="x:inline-block"
            loading="lazy"
          />
        </a>
        <a
          key="github-link"
          href="https://github.com/modu-ai/moai-adk"
          target="_blank"
          rel="noopener noreferrer"
        >
          modu-ai/moai-adk
        </a>
      </div>
    ),
  },
  chat: {
    link: "https://discord.gg/moai-adk",
  },
  docsRepositoryBase: "https://github.com/modu-ai/moai-adk/tree/main/docs",
  footer: {
    content: "MIT License © 2025 MoAI-ADK Contributors",
  },
  toc: {
    backToTop: true,
    float: true,
    title: "이 페이지에서",
  },
  sidebar: {
    defaultMenuCollapseLevel: 1,
    toggleButton: true,
  },
  navigation: {
    prev: true,
    next: true,
  },
  search: false,
  navbar: {},
  editLink: {
    content: "이 페이지 수정 →",
  },
  feedback: {
    content: "질문이 있으신가요? 피드백을 남겨주세요!",
    useLink: "https://github.com/modu-ai/moai-adk/issues/new",
  },
  head: [
    {
      tag: "link",
      rel: "icon",
      href: "/favicon.ico",
      sizes: "32x32",
    },
    {
      tag: "link",
      rel: "icon",
      href: "/favicon.svg",
      type: "image/svg+xml",
    },
    {
      tag: "link",
      rel: "apple-touch-icon",
      href: "/apple-touch-icon.png",
      sizes: "180x180",
    },
    {
      tag: "link",
      rel: "manifest",
      href: "/manifest.json",
    },
    {
      tag: "meta",
      name: "viewport",
      content: "width=device-width, initial-scale=1.0",
    },
    {
      tag: "meta",
      name: "theme-color",
      content: "#000000",
    },
    {
      tag: "meta",
      name: "description",
      content:
        "MoAI-ADK 온라인 문서 - Claude Code를 위한 온라인 제작 인공지능 개발 도구 키트. SPEC 기반 개발, DDD 구현, TRUST 5 품질 시스템으로 완전 자율 자동화.",
    },
    { tag: "meta", property: "og:type", content: "website" },
    { tag: "meta", property: "og:locale", content: "ko_KR" },
    { tag: "meta", property: "og:url", content: "https://adk.mo.ai.kr" },
    {
      tag: "meta",
      property: "og:site_name",
      content: "MoAI-ADK Documentation",
    },
    {
      tag: "meta",
      property: "og:title",
      content: "MoAI-ADK - Claude Code 온라인 제작 인공지능 개발 도구",
    },
    {
      tag: "meta",
      property: "og:description",
      content:
        "SPEC 생성부터 DDD 구현, 문서화까지 완전 자동화하는 AI 개발 도구",
    },
    {
      tag: "meta",
      property: "og:image",
      content: "https://adk.mo.ai.kr/og.png",
    },
    {
      tag: "meta",
      name: "twitter:card",
      content: "summary_large_image",
    },
    {
      tag: "meta",
      name: "twitter:title",
      content: "MoAI-ADK - Claude Code 온라인 제작 인공지능 개발 도구",
    },
    {
      tag: "meta",
      name: "twitter:description",
      content:
        "SPEC 생성부터 DDD 구현, 문서화까지 완전 자율 자동화하는 AI 개발 도구",
    },
    {
      tag: "meta",
      name: "twitter:image",
      content: "https://adk.mo.ai.kr/og.png",
    },
    {
      tag: "meta",
      name: "keywords",
      content:
        "MoAI-ADK, Claude Code, AI 개발, 온라인 제작, 인공지능 개발 도구, 자동화, SPEC, DDD, TRUST 5",
    },
    {
      tag: "link",
      rel: "alternate",
      hrefLang: "ko",
      href: "https://adk.mo.ai.kr",
    },
    {
      tag: "link",
      rel: "alternate",
      hrefLang: "en",
      href: "https://adk.mo.ai.kr/en",
    },
    {
      tag: "link",
      rel: "alternate",
      hrefLang: "zh",
      href: "https://adk.mo.ai.kr/zh",
    },
    {
      tag: "link",
      rel: "alternate",
      hrefLang: "ja",
      href: "https://adk.mo.ai.kr/ja",
    },
    {
      tag: "link",
      rel: "alternate",
      hrefLang: "x-default",
      href: "https://adk.mo.ai.kr",
    },
    {
      tag: "link",
      rel: "canonical",
      href: "https://adk.mo.ai.kr",
    },
    <Analytics key="vercel-analytics" />,
  ],
};

export default config;
