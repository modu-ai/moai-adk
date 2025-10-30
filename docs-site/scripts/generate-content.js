// @CODE:CONTENT-GENERATOR-001 - Generate remaining MDX pages efficiently
const fs = require('fs');
const path = require('path');

const PAGES_DIR = path.join(__dirname, '../pages');

// Template for remaining pages
const generateMDXContent = (category, page, lang) => {
  const titles = {
    ko: {
      // Concepts
      'spec-first-tdd': 'SPEC-First TDD',
      'ears-guide': 'EARS 가이드',
      'tag-system': 'TAG 시스템',
      // Workflow
      'workflow-overview': '워크플로우 개요',
      '0-project': '0️⃣ 프로젝트 초기화',
      '1-plan': '1️⃣ SPEC 작성',
      '2-run': '2️⃣ TDD 구현',
      '3-sync': '3️⃣ 문서 동기화',
      // Skills
      'skills-overview': 'Skills 개요',
      'foundation': 'Foundation Tier',
      'essentials': 'Essentials Tier',
      'alfred': 'Alfred Tier',
      'domain': 'Domain Tier',
      // Agents
      'agents-overview': 'Sub-agents 개요',
      'spec-builder': 'spec-builder',
      'git-manager': 'git-manager',
      // CLI
      'commands': '커맨드 개요',
      'alfred-0-project': '/alfred:0-project',
      'alfred-1-plan': '/alfred:1-plan',
      'alfred-2-run': '/alfred:2-run',
      'alfred-3-sync': '/alfred:3-sync',
      // Config
      'project-config': '프로젝트 설정',
      'git-strategy': 'Git 전략',
      // API
      'specs': 'SPEC API',
      'tags': 'TAG API',
      'git': 'Git API',
      // Misc
      'contributing': '기여하기',
      'faq': 'FAQ',
      'troubleshooting': '문제 해결',
      'changelog': '변경 로그',
      'roadmap': '로드맵',
      'community': '커뮤니티',
      'license': '라이선스',
    },
    en: {
      // Concepts
      'spec-first-tdd': 'SPEC-First TDD',
      'ears-guide': 'EARS Guide',
      'tag-system': 'TAG System',
      // Workflow
      'workflow-overview': 'Workflow Overview',
      '0-project': '0️⃣ Project Init',
      '1-plan': '1️⃣ Plan SPEC',
      '2-run': '2️⃣ TDD Implementation',
      '3-sync': '3️⃣ Sync Docs',
      // Skills
      'skills-overview': 'Skills Overview',
      'foundation': 'Foundation Tier',
      'essentials': 'Essentials Tier',
      'alfred': 'Alfred Tier',
      'domain': 'Domain Tier',
      // Agents
      'agents-overview': 'Sub-agents Overview',
      'spec-builder': 'spec-builder',
      'git-manager': 'git-manager',
      // CLI
      'commands': 'Commands Overview',
      'alfred-0-project': '/alfred:0-project',
      'alfred-1-plan': '/alfred:1-plan',
      'alfred-2-run': '/alfred:2-run',
      'alfred-3-sync': '/alfred:3-sync',
      // Config
      'project-config': 'Project Configuration',
      'git-strategy': 'Git Strategy',
      // API
      'specs': 'SPEC API',
      'tags': 'TAG API',
      'git': 'Git API',
      // Misc
      'contributing': 'Contributing',
      'faq': 'FAQ',
      'troubleshooting': 'Troubleshooting',
      'changelog': 'Changelog',
      'roadmap': 'Roadmap',
      'community': 'Community',
      'license': 'License',
    },
  };

  const title = titles[lang][page] || page;
  const description = lang === 'ko' ? `${title} 가이드` : `${title} guide`;

  return `---
title: ${title}
description: ${description}
---

# ${title}

${lang === 'ko' ? '이 페이지는 현재 작성 중입니다.' : 'This page is under construction.'}

${lang === 'ko' ? '## 개요' : '## Overview'}

${lang === 'ko'
  ? `${title}에 대한 자세한 내용은 곧 업데이트될 예정입니다.`
  : `Detailed content for ${title} will be updated soon.`}

${lang === 'ko' ? '## 주요 특징' : '## Key Features'}

- ${lang === 'ko' ? '특징 1' : 'Feature 1'}
- ${lang === 'ko' ? '특징 2' : 'Feature 2'}
- ${lang === 'ko' ? '특징 3' : 'Feature 3'}

${lang === 'ko' ? '## 사용법' : '## Usage'}

\`\`\`bash
# ${lang === 'ko' ? '예제 명령어' : 'Example command'}
${lang === 'ko' ? '# 상세 사용법은 곧 업데이트됩니다' : '# Detailed usage will be updated soon'}
\`\`\`

${lang === 'ko' ? '## 다음 단계' : '## Next Steps'}

- ${lang === 'ko' ? '[홈으로](/ko)' : '[Home](/en)'}
- ${lang === 'ko' ? '[문서 개요](/ko/introduction/overview)' : '[Overview](/en/introduction/overview)'}
`;
};

// Pages to generate
const pagesToGenerate = [
  // Concepts (3 pages)
  { category: 'concepts', page: 'spec-first-tdd' },
  { category: 'concepts', page: 'ears-guide' },
  { category: 'concepts', page: 'tag-system' },
  // Workflow (5 pages)
  { category: 'workflow', page: 'overview' },
  { category: 'workflow', page: '0-project' },
  { category: 'workflow', page: '1-plan' },
  { category: 'workflow', page: '2-run' },
  { category: 'workflow', page: '3-sync' },
  // Skills (5 pages)
  { category: 'skills', page: 'overview' },
  { category: 'skills', page: 'foundation' },
  { category: 'skills', page: 'essentials' },
  { category: 'skills', page: 'alfred' },
  { category: 'skills', page: 'domain' },
  // Agents (3 pages)
  { category: 'agents', page: 'overview' },
  { category: 'agents', page: 'spec-builder' },
  { category: 'agents', page: 'git-manager' },
  // CLI (5 pages)
  { category: 'cli', page: 'commands' },
  { category: 'cli', page: 'alfred-0-project' },
  { category: 'cli', page: 'alfred-1-plan' },
  { category: 'cli', page: 'alfred-2-run' },
  { category: 'cli', page: 'alfred-3-sync' },
  // Config (2 pages)
  { category: 'config', page: 'project-config' },
  { category: 'config', page: 'git-strategy' },
  // API (3 pages)
  { category: 'api', page: 'specs' },
  { category: 'api', page: 'tags' },
  { category: 'api', page: 'git' },
  // Misc (7 pages)
  { category: 'misc', page: 'contributing' },
  { category: 'misc', page: 'faq' },
  { category: 'misc', page: 'troubleshooting' },
  { category: 'misc', page: 'changelog' },
  { category: 'misc', page: 'roadmap' },
  { category: 'misc', page: 'community' },
  { category: 'misc', page: 'license' },
];

// Generate all pages
let generated = 0;
for (const { category, page } of pagesToGenerate) {
  for (const lang of ['ko', 'en']) {
    const filePath = path.join(PAGES_DIR, lang, category, `${page}.mdx`);

    // Skip if already exists
    if (fs.existsSync(filePath)) {
      console.log(`⏭️  Skipped: ${filePath} (already exists)`);
      continue;
    }

    const content = generateMDXContent(category, page, lang);
    fs.writeFileSync(filePath, content, 'utf-8');
    generated++;
    console.log(`✅ Generated: ${filePath}`);
  }
}

console.log(`\n🎉 Total generated: ${generated} files`);
console.log(`📊 Expected: ${pagesToGenerate.length * 2} files (${pagesToGenerate.length} pages × 2 languages)`);
