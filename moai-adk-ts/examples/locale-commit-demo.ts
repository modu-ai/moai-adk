#!/usr/bin/env node
/**
 * @file Locale-based Commit Message Demo
 * @description Demonstrates locale-based commit message generation
 */

import {
  type CommitLocale,
  type TDDStage,
  getTDDCommitMessage,
  getTDDCommitWithTag,
} from '../src/core/git/constants/commit-message-locales';

console.log('🗿 MoAI-ADK Commit Message Locale Demo\n');

// Demo data
const specId = 'AUTH-001';
const stages: TDDStage[] = ['RED', 'GREEN', 'REFACTOR', 'DOCS'];

// Messages for each locale
const messages: Record<CommitLocale, Record<TDDStage, string>> = {
  ko: {
    RED: '로그인 실패 테스트 추가',
    GREEN: '로그인 기능 구현',
    REFACTOR: '로그인 코드 개선',
    DOCS: '로그인 API 문서 작성',
  },
  en: {
    RED: 'add login failure test',
    GREEN: 'implement login feature',
    REFACTOR: 'improve login code',
    DOCS: 'write login API documentation',
  },
  ja: {
    RED: 'ログイン失敗テスト追加',
    GREEN: 'ログイン機能実装',
    REFACTOR: 'ログインコード改善',
    DOCS: 'ログインAPIドキュメント作成',
  },
  zh: {
    RED: '添加登录失败测试',
    GREEN: '实现登录功能',
    REFACTOR: '改进登录代码',
    DOCS: '编写登录API文档',
  },
};

// Demo for each locale
const locales: CommitLocale[] = ['ko', 'en', 'ja', 'zh'];
const localeNames = {
  ko: 'Korean (한국어)',
  en: 'English',
  ja: 'Japanese (日本語)',
  zh: 'Chinese (中文)',
};

for (const locale of locales) {
  console.log(`\n${'='.repeat(60)}`);
  console.log(`📍 Locale: ${localeNames[locale]}`);
  console.log('='.repeat(60));

  for (const stage of stages) {
    console.log(`\n--- ${stage} Phase ---`);

    // Simple message
    const simple = getTDDCommitMessage(locale, stage, messages[locale][stage]);
    console.log('\n✅ Simple message:');
    console.log(simple);

    // With @TAG
    const withTag = getTDDCommitWithTag(
      locale,
      stage,
      messages[locale][stage],
      specId
    );
    console.log('\n✅ With @TAG:');
    console.log(withTag);
  }
}

// Demo git command generation
console.log('\n\n' + '='.repeat(60));
console.log('📦 Example Git Commands');
console.log('='.repeat(60));

console.log('\n--- Korean Project ---');
console.log('git commit -m "$(cat <<\'EOF\'');
console.log(getTDDCommitWithTag('ko', 'RED', '로그인 테스트', 'AUTH-001'));
console.log('EOF\n)"');

console.log('\n--- English Project ---');
console.log('git commit -m "$(cat <<\'EOF\'');
console.log(getTDDCommitWithTag('en', 'GREEN', 'implement auth', 'AUTH-001'));
console.log('EOF\n)"');

console.log('\n--- Japanese Project ---');
console.log('git commit -m "$(cat <<\'EOF\'');
console.log(
  getTDDCommitWithTag('ja', 'REFACTOR', 'コード改善', 'AUTH-001')
);
console.log('EOF\n)"');

console.log('\n--- Chinese Project ---');
console.log('git commit -m "$(cat <<\'EOF\'');
console.log(getTDDCommitWithTag('zh', 'DOCS', '更新文档', 'AUTH-001'));
console.log('EOF\n)"');

// Demo configuration examples
console.log('\n\n' + '='.repeat(60));
console.log('⚙️  Configuration Examples');
console.log('='.repeat(60));

const configs = [
  {
    locale: 'ko',
    desc: 'Korean project',
    example: '로컬 중심 개발 프로젝트',
  },
  {
    locale: 'en',
    desc: 'English project (default)',
    example: 'International team project',
  },
  {
    locale: 'ja',
    desc: 'Japanese project',
    example: '日本のチームプロジェクト',
  },
  {
    locale: 'zh',
    desc: 'Chinese project',
    example: '中国团队项目',
  },
];

for (const config of configs) {
  console.log(`\n--- ${config.desc} ---`);
  console.log('File: .moai/config.json');
  console.log(
    JSON.stringify(
      {
        project: {
          name: 'my-project',
          mode: 'team',
          locale: config.locale,
          description: config.example,
        },
      },
      null,
      2
    )
  );
}

console.log('\n\n✅ Demo complete!\n');
console.log('💡 Tip: Set your locale in .moai/config.json:');
console.log('   "project": { "locale": "ko" }\n');
