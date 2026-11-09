import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  output: 'static',
  redirects: { '/': '/ko/' },
  integrations: [
    starlight({
      title: { ko: 'MoAI CLI 문서', en: 'MoAI CLI Docs' },
      favicon: '/favicon.svg',
      defaultLocale: 'ko',
      locales: {
        ko: { label: '한국어', lang: 'ko' },
        en: { label: 'English', lang: 'en' },
      },
      social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/modu-ai/moai-adk' }],
      sidebar: [
        { slug: '', label: '소개', translations: { en: 'Overview' } },
        {
          label: '사용 가이드',
          translations: { en: 'Guides' },
          items: [
            { slug: 'guides/getting-started' },
            { slug: 'guides/claude-code' },
            { slug: 'guides/codex' },
          ],
        },
        {
          label: '참조',
          translations: { en: 'Reference' },
          items: [{ slug: 'reference/cli' }, { slug: 'reference/safety' }],
        },
      ],
    }),
  ],
});
