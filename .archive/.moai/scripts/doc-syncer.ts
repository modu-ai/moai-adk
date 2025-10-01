#!/usr/bin/env tsx
// CODE-DOC-SYNCER-001: 문서 동기화 및 Living Document 관리 스크립트
// 연결: SPEC-DOC-SYNC-001 → SPEC-DOC-SYNCER-001 → CODE-DOC-SYNC-001

import { program } from 'commander';
import { promises as fs } from 'fs';
import path from 'path';
import chalk from 'chalk';

interface DocSyncerOptions {
  spec?: string;
  all?: boolean;
  target?: 'readme' | 'api' | 'release' | 'all';
  format?: 'markdown' | 'html' | 'json';
  push?: boolean;
}

interface DocumentSection {
  id: string;
  title: string;
  content: string;
  lastUpdated: string;
  tags: string[];
}

interface SyncReport {
  timestamp: string;
  synced: string[];
  errors: string[];
  warnings: string[];
  stats: {
    totalDocs: number;
    updatedDocs: number;
    newDocs: number;
    errorCount: number;
  };
}

interface SpecSummary {
  id: string;
  title: string;
  status: string;
  priority: string;
  tags: string[];
  implementation?: {
    completed: boolean;
    coverage?: number;
    testsPassed?: number;
  };
}

async function loadSpecSummaries(): Promise<SpecSummary[]> {
  const specsDir = '.moai/specs';
  const summaries: SpecSummary[] = [];

  try {
    const entries = await fs.readdir(specsDir, { withFileTypes: true });
    const specDirs = entries.filter(entry => entry.isDirectory() && entry.name.startsWith('SPEC-'));

    for (const dir of specDirs) {
      try {
        const specPath = path.join(specsDir, dir.name, 'spec.md');
        const metadataPath = path.join(specsDir, dir.name, 'metadata.json');

        let metadata: any = {};
        try {
          const metadataContent = await fs.readFile(metadataPath, 'utf-8');
          metadata = JSON.parse(metadataContent);
        } catch {
          // 메타데이터 파일이 없으면 SPEC 파일에서 추출
          const specContent = await fs.readFile(specPath, 'utf-8');
          metadata = extractMetadataFromSpec(specContent, dir.name);
        }

        summaries.push({
          id: dir.name,
          title: metadata.title || '제목 없음',
          status: metadata.status || 'draft',
          priority: metadata.priority || 'medium',
          tags: metadata.tags || [],
          implementation: metadata.implementation
        });
      } catch (error) {
        console.warn(chalk.yellow(`⚠️  ${dir.name} 처리 중 오류: ${error.message}`));
      }
    }
  } catch (error) {
    console.warn(chalk.yellow(`⚠️  SPEC 디렉토리 읽기 실패: ${error.message}`));
  }

  return summaries.sort((a, b) => a.id.localeCompare(b.id));
}

function extractMetadataFromSpec(content: string, specId: string): any {
  const titleMatch = content.match(/^# SPEC-\d+:\s*(.+)$/m);
  const priorityMatch = content.match(/우선순위:\s*(\w+)/i);
  const statusMatch = content.match(/상태:\s*(\w+)/i);

  return {
    title: titleMatch ? titleMatch[1].trim() : specId,
    priority: priorityMatch ? priorityMatch[1] : 'medium',
    status: statusMatch ? statusMatch[1] : 'draft',
    tags: []
  };
}

async function syncReadme(specs: SpecSummary[]): Promise<{ success: boolean; message: string }> {
  try {
    const readmePath = 'README.md';
    let readmeContent = '';

    // 기존 README.md 읽기
    try {
      readmeContent = await fs.readFile(readmePath, 'utf-8');
    } catch {
      // README.md가 없으면 새로 생성
      readmeContent = generateBasicReadme();
    }

    // SPEC 섹션 업데이트
    const specSection = generateSpecSection(specs);
    readmeContent = updateReadmeSection(readmeContent, 'SPEC 목록', specSection);

    // 진행 상황 섹션 업데이트
    const progressSection = generateProgressSection(specs);
    readmeContent = updateReadmeSection(readmeContent, '프로젝트 진행 상황', progressSection);

    await fs.writeFile(readmePath, readmeContent);

    return {
      success: true,
      message: `README.md 업데이트 완료 (${specs.length}개 SPEC)`
    };
  } catch (error) {
    return {
      success: false,
      message: `README.md 업데이트 실패: ${error.message}`
    };
  }
}

function generateBasicReadme(): string {
  return `# 프로젝트 이름

프로젝트 설명을 여기에 작성하세요.

## 시작하기

### 요구사항

- Node.js 18+
- Git

### 설치

\`\`\`bash
npm install
\`\`\`

### 사용법

\`\`\`bash
npm start
\`\`\`

## SPEC 목록

<!-- SPEC_SECTION_START -->
<!-- SPEC_SECTION_END -->

## 프로젝트 진행 상황

<!-- PROGRESS_SECTION_START -->
<!-- PROGRESS_SECTION_END -->

## 기여하기

1. 이 저장소를 Fork 하세요
2. 새로운 브랜치를 생성하세요 (\`git checkout -b feature/새기능\`)
3. 변경사항을 커밋하세요 (\`git commit -am '새 기능 추가'\`)
4. 브랜치에 Push 하세요 (\`git push origin feature/새기능\`)
5. Pull Request를 생성하세요

## 라이센스

이 프로젝트는 MIT 라이센스 하에 있습니다.
`;
}

function generateSpecSection(specs: SpecSummary[]): string {
  if (specs.length === 0) {
    return '현재 작성된 SPEC이 없습니다.\n\n`/moai:1-spec`을 사용하여 첫 번째 SPEC을 생성하세요.';
  }

  let section = '| SPEC ID | 제목 | 상태 | 우선순위 | 구현 상태 |\n';
  section += '|---------|------|------|----------|----------|\n';

  for (const spec of specs) {
    const statusBadge = getStatusBadge(spec.status);
    const priorityBadge = getPriorityBadge(spec.priority);
    const implStatus = spec.implementation?.completed ? '✅ 완료' : '⏳ 진행중';

    section += `| [${spec.id}](.moai/specs/${spec.id}/spec.md) | ${spec.title} | ${statusBadge} | ${priorityBadge} | ${implStatus} |\n`;
  }

  return section;
}

function generateProgressSection(specs: SpecSummary[]): string {
  const totalSpecs = specs.length;
  const completedSpecs = specs.filter(s => s.implementation?.completed).length;
  const inProgressSpecs = specs.filter(s => s.status === 'in_progress' || s.status === 'review').length;
  const draftSpecs = specs.filter(s => s.status === 'draft').length;

  const progressPercentage = totalSpecs > 0 ? Math.round((completedSpecs / totalSpecs) * 100) : 0;

  let section = `### 전체 진행률: ${progressPercentage}% (${completedSpecs}/${totalSpecs})\n\n`;

  // 진행률 바
  const progressBar = generateProgressBar(progressPercentage);
  section += `${progressBar}\n\n`;

  // 상태별 통계
  section += '#### 상태별 통계\n\n';
  section += `- 📝 Draft: ${draftSpecs}개\n`;
  section += `- 🔄 In Progress: ${inProgressSpecs}개\n`;
  section += `- ✅ Completed: ${completedSpecs}개\n\n`;

  // 우선순위별 통계
  const priorityStats = {
    critical: specs.filter(s => s.priority === 'critical').length,
    high: specs.filter(s => s.priority === 'high').length,
    medium: specs.filter(s => s.priority === 'medium').length,
    low: specs.filter(s => s.priority === 'low').length
  };

  section += '#### 우선순위별 통계\n\n';
  section += `- 🔴 Critical: ${priorityStats.critical}개\n`;
  section += `- 🟠 High: ${priorityStats.high}개\n`;
  section += `- 🟡 Medium: ${priorityStats.medium}개\n`;
  section += `- 🟢 Low: ${priorityStats.low}개\n`;

  return section;
}

function getStatusBadge(status: string): string {
  switch (status.toLowerCase()) {
    case 'draft': return '📝 Draft';
    case 'review': return '👀 Review';
    case 'approved': return '✅ Approved';
    case 'implemented': return '🚀 Implemented';
    case 'in_progress': return '🔄 In Progress';
    default: return `📄 ${status}`;
  }
}

function getPriorityBadge(priority: string): string {
  switch (priority.toLowerCase()) {
    case 'critical': return '🔴 Critical';
    case 'high': return '🟠 High';
    case 'medium': return '🟡 Medium';
    case 'low': return '🟢 Low';
    default: return `⚪ ${priority}`;
  }
}

function generateProgressBar(percentage: number): string {
  const totalBars = 20;
  const filledBars = Math.round((percentage / 100) * totalBars);
  const emptyBars = totalBars - filledBars;

  return `${'█'.repeat(filledBars)}${'░'.repeat(emptyBars)} ${percentage}%`;
}

function updateReadmeSection(content: string, sectionName: string, newContent: string): string {
  const startMarker = `<!-- ${sectionName.toUpperCase().replace(/\s/g, '_')}_SECTION_START -->`;
  const endMarker = `<!-- ${sectionName.toUpperCase().replace(/\s/g, '_')}_SECTION_END -->`;

  const startIndex = content.indexOf(startMarker);
  const endIndex = content.indexOf(endMarker);

  if (startIndex === -1 || endIndex === -1) {
    // 마커가 없으면 끝에 추가
    return content + `\n\n## ${sectionName}\n\n${startMarker}\n${newContent}\n${endMarker}\n`;
  }

  const before = content.substring(0, startIndex + startMarker.length);
  const after = content.substring(endIndex);

  return `${before}\n${newContent}\n${after}`;
}

async function generateApiDocs(specs: SpecSummary[]): Promise<{ success: boolean; message: string }> {
  try {
    const apiDocsDir = 'docs/api';
    await fs.mkdir(apiDocsDir, { recursive: true });

    // API 문서 인덱스 생성
    let indexContent = '# API 문서\n\n';
    indexContent += '이 섹션은 프로젝트의 API 문서를 포함합니다.\n\n';

    const apiSpecs = specs.filter(spec =>
      spec.tags.some(tag => tag.includes('API')) ||
      spec.title.toLowerCase().includes('api')
    );

    if (apiSpecs.length > 0) {
      indexContent += '## API SPEC 목록\n\n';
      for (const spec of apiSpecs) {
        indexContent += `- [${spec.id}: ${spec.title}](../.moai/specs/${spec.id}/spec.md)\n`;
      }
    } else {
      indexContent += '현재 API 관련 SPEC이 없습니다.\n';
    }

    await fs.writeFile(path.join(apiDocsDir, 'index.md'), indexContent);

    return {
      success: true,
      message: `API 문서 생성 완료 (${apiSpecs.length}개 API SPEC)`
    };
  } catch (error) {
    return {
      success: false,
      message: `API 문서 생성 실패: ${error.message}`
    };
  }
}

async function generateReleaseNotes(specs: SpecSummary[]): Promise<{ success: boolean; message: string }> {
  try {
    const completedSpecs = specs.filter(s => s.implementation?.completed);

    if (completedSpecs.length === 0) {
      return {
        success: true,
        message: '완료된 SPEC이 없어 릴리스 노트를 생성하지 않습니다.'
      };
    }

    const releaseNotesPath = 'CHANGELOG.md';
    const currentDate = new Date().toISOString().split('T')[0];

    let releaseContent = `# 변경 이력\n\n## [Unreleased] - ${currentDate}\n\n`;

    // 우선순위별로 그룹화
    const criticalFeatures = completedSpecs.filter(s => s.priority === 'critical');
    const highFeatures = completedSpecs.filter(s => s.priority === 'high');
    const mediumFeatures = completedSpecs.filter(s => s.priority === 'medium');
    const lowFeatures = completedSpecs.filter(s => s.priority === 'low');

    if (criticalFeatures.length > 0) {
      releaseContent += '### 🔴 Critical 기능\n\n';
      for (const spec of criticalFeatures) {
        releaseContent += `- **${spec.id}**: ${spec.title}\n`;
      }
      releaseContent += '\n';
    }

    if (highFeatures.length > 0) {
      releaseContent += '### 🟠 주요 기능\n\n';
      for (const spec of highFeatures) {
        releaseContent += `- **${spec.id}**: ${spec.title}\n`;
      }
      releaseContent += '\n';
    }

    if (mediumFeatures.length > 0) {
      releaseContent += '### 🟡 개선사항\n\n';
      for (const spec of mediumFeatures) {
        releaseContent += `- **${spec.id}**: ${spec.title}\n`;
      }
      releaseContent += '\n';
    }

    if (lowFeatures.length > 0) {
      releaseContent += '### 🟢 기타 변경사항\n\n';
      for (const spec of lowFeatures) {
        releaseContent += `- **${spec.id}**: ${spec.title}\n`;
      }
      releaseContent += '\n';
    }

    // 기존 CHANGELOG 내용과 병합
    try {
      const existingContent = await fs.readFile(releaseNotesPath, 'utf-8');
      const existingReleases = existingContent.replace(/^# 변경 이력\n\n## \[Unreleased\].*?\n\n/s, '');
      releaseContent += existingReleases;
    } catch {
      // 기존 파일이 없으면 무시
    }

    await fs.writeFile(releaseNotesPath, releaseContent);

    return {
      success: true,
      message: `릴리스 노트 생성 완료 (${completedSpecs.length}개 완료된 SPEC)`
    };
  } catch (error) {
    return {
      success: false,
      message: `릴리스 노트 생성 실패: ${error.message}`
    };
  }
}

async function saveSyncReport(report: SyncReport): Promise<void> {
  const reportsDir = '.moai/reports';
  await fs.mkdir(reportsDir, { recursive: true });

  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  const reportPath = path.join(reportsDir, `sync-report-${timestamp}.json`);

  await fs.writeFile(reportPath, JSON.stringify(report, null, 2));
}

async function syncDocuments(options: DocSyncerOptions): Promise<{ success: boolean; message: string; report: SyncReport }> {
  const syncReport: SyncReport = {
    timestamp: new Date().toISOString(),
    synced: [],
    errors: [],
    warnings: [],
    stats: {
      totalDocs: 0,
      updatedDocs: 0,
      newDocs: 0,
      errorCount: 0
    }
  };

  try {
    console.log(chalk.blue('📋 SPEC 정보 로딩 중...'));
    const specs = await loadSpecSummaries();
    syncReport.stats.totalDocs = specs.length;

    const targets = options.target === 'all'
      ? ['readme', 'api', 'release']
      : [options.target || 'readme'];

    for (const target of targets) {
      try {
        let result: { success: boolean; message: string };

        switch (target) {
          case 'readme':
            result = await syncReadme(specs);
            break;
          case 'api':
            result = await generateApiDocs(specs);
            break;
          case 'release':
            result = await generateReleaseNotes(specs);
            break;
          default:
            throw new Error(`알 수 없는 대상: ${target}`);
        }

        if (result.success) {
          syncReport.synced.push(target);
          syncReport.stats.updatedDocs++;
        } else {
          syncReport.errors.push(`${target}: ${result.message}`);
          syncReport.stats.errorCount++;
        }
      } catch (error) {
        syncReport.errors.push(`${target}: ${error.message}`);
        syncReport.stats.errorCount++;
      }
    }

    // 동기화 리포트 저장
    await saveSyncReport(syncReport);

    const success = syncReport.errors.length === 0;
    const message = `문서 동기화 ${success ? '완료' : '부분 완료'}: ${syncReport.synced.length}개 성공, ${syncReport.errors.length}개 실패`;

    return { success, message, report: syncReport };

  } catch (error) {
    syncReport.errors.push(`전체 동기화 실패: ${error.message}`);
    syncReport.stats.errorCount++;

    return {
      success: false,
      message: `문서 동기화 실패: ${error.message}`,
      report: syncReport
    };
  }
}

program
  .name('doc-syncer')
  .description('MoAI 문서 동기화 및 Living Document 관리')
  .option('-s, --spec <spec-id>', '특정 SPEC 동기화')
  .option('-a, --all', '모든 문서 동기화')
  .option('-t, --target <target>', '동기화 대상 (readme|api|release|all)', 'readme')
  .option('-f, --format <format>', '출력 형식 (markdown|html|json)', 'markdown')
  .option('--push', 'Git에 변경사항 푸시')
  .action(async (options: DocSyncerOptions) => {
    try {
      console.log(chalk.blue('📚 문서 동기화 시작...'));

      const result = await syncDocuments(options);

      if (result.success) {
        console.log(chalk.green('✅'), result.message);
      } else {
        console.log(chalk.yellow('⚠️'), result.message);
      }

      // 상세 리포트 출력
      if (result.report.synced.length > 0) {
        console.log(chalk.cyan('\n📄 동기화된 문서:'));
        for (const doc of result.report.synced) {
          console.log(chalk.green(`  ✅ ${doc}`));
        }
      }

      if (result.report.errors.length > 0) {
        console.log(chalk.red('\n❌ 오류:'));
        for (const error of result.report.errors) {
          console.log(chalk.red(`  ❌ ${error}`));
        }
      }

      console.log(JSON.stringify({
        success: result.success,
        synced: result.report.synced,
        errors: result.report.errors,
        stats: result.report.stats,
        nextSteps: result.success ? [
          '문서가 성공적으로 동기화되었습니다',
          'Git 커밋을 고려하세요'
        ] : [
          '오류를 수정하고 다시 시도하세요'
        ]
      }, null, 2));

      process.exit(result.success ? 0 : 1);
    } catch (error) {
      console.error(chalk.red('❌ 스크립트 실행 실패:'), error.message);
      console.log(JSON.stringify({
        success: false,
        error: error.message
      }, null, 2));
      process.exit(1);
    }
  });

program.parse();