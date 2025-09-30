#!/usr/bin/env tsx
/**
 * @FEATURE:TAG-UPDATER-001 | Chain: @REQ:TAG-UPDATER-001 -> @DESIGN:TAG-UPDATER-001 -> @TASK:TAG-UPDATER-001 -> @TEST:TAG-UPDATER-001
 * Related: @API:TAG-UPDATER-001, @DATA:TAG-UPDATER-001
 *
 * TAG 시스템 업데이트 스크립트
 * - TAG 체인 자동 생성
 * - TAG 무결성 검증
 * - TAG 인덱스 갱신
 */

import { program } from 'commander';
import { promises as fs } from 'fs';
import path from 'path';
import chalk from 'chalk';

interface TagUpdaterOptions {
  scan?: boolean;
  repair?: boolean;
  validate?: boolean;
  cleanup?: boolean;
  backup?: boolean;
  force?: boolean;
}

interface TagEntry {
  id: string;
  type: TagType;
  category: TagCategory;
  title: string;
  description?: string;
  status: TagStatus;
  priority: TagPriority;
  parents: string[];
  children: string[];
  files: string[];
  createdAt: string;
  updatedAt: string;
  author?: string;
  metadata?: Record<string, unknown>;
}

interface TagDatabase {
  version: string;
  tags: Record<string, TagEntry>;
  indexes: {
    byType: Record<TagType, string[]>;
    byCategory: Record<TagCategory, string[]>;
    byStatus: Record<TagStatus, string[]>;
    byFile: Record<string, string[]>;
  };
  metadata: {
    totalTags: number;
    lastUpdated: string;
    checksum?: string;
  };
}

type TagType = 'REQ' | 'DESIGN' | 'TASK' | 'TEST' | 'VISION' | 'STRUCT' | 'TECH' | 'ADR' | 'FEATURE' | 'API' | 'UI' | 'DATA' | 'PERF' | 'SEC' | 'DOCS' | 'TAG';
type TagCategory = 'PRIMARY' | 'STEERING' | 'IMPLEMENTATION' | 'QUALITY';
type TagStatus = 'pending' | 'in_progress' | 'completed' | 'blocked';
type TagPriority = 'critical' | 'high' | 'medium' | 'low';

interface TagScanResult {
  foundTags: TagEntry[];
  orphanedTags: string[];
  brokenReferences: string[];
  duplicateTags: string[];
  filesScanned: number;
}

interface TagUpdateResult {
  success: boolean;
  message: string;
  scanResult?: TagScanResult;
  updatedTags: number;
  repairedTags: number;
  removedTags: number;
}

const TAG_CATEGORIES: Record<TagType, TagCategory> = {
  'REQ': 'PRIMARY',
  'DESIGN': 'PRIMARY',
  'TASK': 'PRIMARY',
  'TEST': 'PRIMARY',
  'VISION': 'STEERING',
  'STRUCT': 'STEERING',
  'TECH': 'STEERING',
  'ADR': 'STEERING',
  'FEATURE': 'IMPLEMENTATION',
  'API': 'IMPLEMENTATION',
  'UI': 'IMPLEMENTATION',
  'DATA': 'IMPLEMENTATION',
  'PERF': 'QUALITY',
  'SEC': 'QUALITY',
  'DOCS': 'QUALITY',
  'TAG': 'QUALITY'
};

const TAG_PATTERN = /@([A-Z]+):([A-Z0-9-]+)/g;

async function loadTagDatabase(): Promise<TagDatabase> {
  const tagDbPath = '.moai/indexes/tags.json';

  try {
    const content = await fs.readFile(tagDbPath, 'utf-8');
    return JSON.parse(content) as TagDatabase;
  } catch {
    // 데이터베이스가 없으면 빈 구조 생성
    return createEmptyDatabase();
  }
}

function createEmptyDatabase(): TagDatabase {
  return {
    version: '1.0.0',
    tags: {},
    indexes: {
      byType: {} as Record<TagType, string[]>,
      byCategory: {} as Record<TagCategory, string[]>,
      byStatus: {} as Record<TagStatus, string[]>,
      byFile: {}
    },
    metadata: {
      totalTags: 0,
      lastUpdated: new Date().toISOString()
    }
  };
}

async function saveTagDatabase(database: TagDatabase): Promise<void> {
  const tagDbPath = '.moai/indexes/tags.json';

  // 인덱스 디렉토리 생성
  await fs.mkdir(path.dirname(tagDbPath), { recursive: true });

  // 메타데이터 업데이트
  database.metadata.totalTags = Object.keys(database.tags).length;
  database.metadata.lastUpdated = new Date().toISOString();

  // 인덱스 재구성
  database.indexes = rebuildIndexes(database.tags);

  await fs.writeFile(tagDbPath, JSON.stringify(database, null, 2));
}

function rebuildIndexes(tags: Record<string, TagEntry>): TagDatabase['indexes'] {
  const indexes: TagDatabase['indexes'] = {
    byType: {} as Record<TagType, string[]>,
    byCategory: {} as Record<TagCategory, string[]>,
    byStatus: {} as Record<TagStatus, string[]>,
    byFile: {}
  };

  // 인덱스 초기화
  for (const type of Object.keys(TAG_CATEGORIES) as TagType[]) {
    indexes.byType[type] = [];
  }

  for (const category of Object.values(TAG_CATEGORIES) as TagCategory[]) {
    indexes.byCategory[category] = [];
  }

  const statuses: TagStatus[] = ['pending', 'in_progress', 'completed', 'blocked'];
  for (const status of statuses) {
    indexes.byStatus[status] = [];
  }

  // 인덱스 구성
  for (const [tagId, tag] of Object.entries(tags)) {
    // 타입별 인덱스
    if (!indexes.byType[tag.type]) {
      indexes.byType[tag.type] = [];
    }
    indexes.byType[tag.type].push(tagId);

    // 카테고리별 인덱스
    if (!indexes.byCategory[tag.category]) {
      indexes.byCategory[tag.category] = [];
    }
    indexes.byCategory[tag.category].push(tagId);

    // 상태별 인덱스
    if (!indexes.byStatus[tag.status]) {
      indexes.byStatus[tag.status] = [];
    }
    indexes.byStatus[tag.status].push(tagId);

    // 파일별 인덱스
    for (const file of tag.files) {
      if (!indexes.byFile[file]) {
        indexes.byFile[file] = [];
      }
      indexes.byFile[file].push(tagId);
    }
  }

  return indexes;
}

async function scanProjectForTags(): Promise<TagScanResult> {
  const foundTags: TagEntry[] = [];
  const filesScanned: string[] = [];
  const tagReferences = new Map<string, { files: string[]; contexts: string[] }>();

  console.log(chalk.blue('🔍 프로젝트 파일 스캔 중...'));

  await scanDirectory('.', tagReferences, filesScanned);

  console.log(chalk.blue(`📁 ${filesScanned.length}개 파일에서 TAG 추출 중...`));

  // TAG 엔트리 생성
  for (const [tagId, info] of tagReferences) {
    const tagEntry = createTagEntryFromId(tagId, info.files);
    if (tagEntry) {
      foundTags.push(tagEntry);
    }
  }

  // 기존 데이터베이스와 비교하여 고아 TAG 찾기
  const database = await loadTagDatabase();
  const currentTagIds = new Set(foundTags.map(t => t.id));
  const orphanedTags = Object.keys(database.tags).filter(id => !currentTagIds.has(id));

  // 중복 TAG 찾기
  const tagIdCounts = new Map<string, number>();
  for (const tag of foundTags) {
    tagIdCounts.set(tag.id, (tagIdCounts.get(tag.id) || 0) + 1);
  }
  const duplicateTags = Array.from(tagIdCounts.entries())
    .filter(([, count]) => count > 1)
    .map(([id]) => id);

  // 끊어진 참조 찾기
  const brokenReferences: string[] = [];
  for (const tag of foundTags) {
    for (const parentId of tag.parents) {
      if (!currentTagIds.has(parentId)) {
        brokenReferences.push(`${tag.id} → ${parentId}`);
      }
    }
    for (const childId of tag.children) {
      if (!currentTagIds.has(childId)) {
        brokenReferences.push(`${tag.id} → ${childId}`);
      }
    }
  }

  return {
    foundTags,
    orphanedTags,
    brokenReferences,
    duplicateTags,
    filesScanned: filesScanned.length
  };
}

async function scanDirectory(
  dirPath: string,
  tagReferences: Map<string, { files: string[]; contexts: string[] }>,
  filesScanned: string[]
): Promise<void> {
  try {
    const entries = await fs.readdir(dirPath, { withFileTypes: true });

    for (const entry of entries) {
      const fullPath = path.join(dirPath, entry.name);

      if (entry.isDirectory()) {
        // 제외할 디렉토리들
        if (!['node_modules', '.git', '.vscode', 'dist', 'build', 'target', '.next'].includes(entry.name)) {
          await scanDirectory(fullPath, tagReferences, filesScanned);
        }
      } else if (entry.isFile()) {
        await scanFile(fullPath, tagReferences, filesScanned);
      }
    }
  } catch (error) {
    // 접근 권한 없는 디렉토리는 무시
  }
}

async function scanFile(
  filePath: string,
  tagReferences: Map<string, { files: string[]; contexts: string[] }>,
  filesScanned: string[]
): Promise<void> {
  try {
    // 특정 파일 형식만 스캔
    const ext = path.extname(filePath).toLowerCase();
    const scanExtensions = ['.md', '.ts', '.js', '.py', '.java', '.go', '.rs', '.cpp', '.c', '.h'];

    if (!scanExtensions.includes(ext)) {
      return;
    }

    const content = await fs.readFile(filePath, 'utf-8');
    filesScanned.push(filePath);

    // TAG 패턴 매칭
    let match;
    TAG_PATTERN.lastIndex = 0; // 정규식 상태 리셋

    while ((match = TAG_PATTERN.exec(content)) !== null) {
      const fullMatch = match[0];
      const tagType = match[1] as TagType;
      const tagSuffix = match[2] || '001';

      // 유효한 TAG 타입인지 확인
      if (TAG_CATEGORIES[tagType]) {
        const tagId = `@${tagType}:${tagSuffix}`;

        if (!tagReferences.has(tagId)) {
          tagReferences.set(tagId, { files: [], contexts: [] });
        }

        const ref = tagReferences.get(tagId)!;
        if (!ref.files.includes(filePath)) {
          ref.files.push(filePath);
        }

        // 컨텍스트 추출 (TAG 주변 텍스트)
        const lines = content.split('\n');
        const lineIndex = content.substring(0, match.index).split('\n').length - 1;
        const contextLine = lines[lineIndex]?.trim();
        if (contextLine && !ref.contexts.includes(contextLine)) {
          ref.contexts.push(contextLine);
        }
      }
    }
  } catch (error) {
    // 파일 읽기 실패는 무시
  }
}

function createTagEntryFromId(tagId: string, files: string[]): TagEntry | null {
  const match = tagId.match(/@([A-Z]+)-(.+)/);
  if (!match) return null;

  const [, typeStr, suffix] = match;
  const type = typeStr as TagType;
  const category = TAG_CATEGORIES[type];

  if (!category) return null;

  return {
    id: tagId,
    type,
    category,
    title: `${type} ${suffix}`,
    status: 'pending',
    priority: 'medium',
    parents: [],
    children: [],
    files,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    author: process.env.USER || 'system'
  };
}

async function repairTagDatabase(database: TagDatabase, scanResult: TagScanResult): Promise<number> {
  let repairedCount = 0;

  // 새로 발견된 TAG들 추가
  for (const newTag of scanResult.foundTags) {
    if (!database.tags[newTag.id]) {
      database.tags[newTag.id] = newTag;
      repairedCount++;
    } else {
      // 파일 목록 업데이트
      const existingTag = database.tags[newTag.id];
      for (const file of newTag.files) {
        if (!existingTag.files.includes(file)) {
          existingTag.files.push(file);
          existingTag.updatedAt = new Date().toISOString();
          repairedCount++;
        }
      }
    }
  }

  // 고아 TAG 제거 또는 상태 변경
  for (const orphanId of scanResult.orphanedTags) {
    if (database.tags[orphanId]) {
      database.tags[orphanId].status = 'blocked';
      database.tags[orphanId].updatedAt = new Date().toISOString();
      repairedCount++;
    }
  }

  // 끊어진 참조 수정
  for (const brokenRef of scanResult.brokenReferences) {
    const [fromId, toId] = brokenRef.split(' → ');
    const fromTag = database.tags[fromId];

    if (fromTag) {
      // 부모/자식 관계에서 끊어진 참조 제거
      fromTag.parents = fromTag.parents.filter(id => id !== toId);
      fromTag.children = fromTag.children.filter(id => id !== toId);
      fromTag.updatedAt = new Date().toISOString();
      repairedCount++;
    }
  }

  return repairedCount;
}

async function validateTagDatabase(database: TagDatabase): Promise<string[]> {
  const issues: string[] = [];

  // 순환 참조 검사
  for (const [tagId, tag] of Object.entries(database.tags)) {
    if (hasCircularReference(tagId, database.tags, new Set())) {
      issues.push(`순환 참조 발견: ${tagId}`);
    }
  }

  // Primary Chain 검증
  const primaryTags = Object.values(database.tags).filter(t => t.category === 'PRIMARY');
  for (const tag of primaryTags) {
    if (tag.type === 'REQ' && tag.children.length === 0) {
      issues.push(`@REQ TAG에 연결된 @DESIGN이 없습니다: ${tag.id}`);
    }
    if (tag.type === 'DESIGN' && !tag.parents.some(p => database.tags[p]?.type === 'REQ')) {
      issues.push(`@DESIGN TAG에 연결된 @REQ가 없습니다: ${tag.id}`);
    }
  }

  // 파일 존재 여부 검사
  for (const [tagId, tag] of Object.entries(database.tags)) {
    for (const file of tag.files) {
      try {
        await fs.access(file);
      } catch {
        issues.push(`TAG ${tagId}가 참조하는 파일이 존재하지 않습니다: ${file}`);
      }
    }
  }

  return issues;
}

function hasCircularReference(
  tagId: string,
  tags: Record<string, TagEntry>,
  visited: Set<string>
): boolean {
  if (visited.has(tagId)) {
    return true;
  }

  const tag = tags[tagId];
  if (!tag) return false;

  visited.add(tagId);

  for (const childId of tag.children) {
    if (hasCircularReference(childId, tags, new Set(visited))) {
      return true;
    }
  }

  return false;
}

async function backupTagDatabase(): Promise<string> {
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  const backupPath = `.moai/indexes/tags-backup-${timestamp}.json`;

  try {
    const currentDb = await fs.readFile('.moai/indexes/tags.json', 'utf-8');
    await fs.writeFile(backupPath, currentDb);
    return backupPath;
  } catch (error) {
    throw new Error(`백업 실패: ${error.message}`);
  }
}

async function updateTags(options: TagUpdaterOptions): Promise<TagUpdateResult> {
  try {
    let database = await loadTagDatabase();
    let scanResult: TagScanResult | undefined;
    let updatedTags = 0;
    let repairedTags = 0;
    let removedTags = 0;

    // 백업 생성
    if (options.backup) {
      const backupPath = await backupTagDatabase();
      console.log(chalk.blue(`💾 백업 생성: ${backupPath}`));
    }

    // 스캔 실행
    if (options.scan || options.repair) {
      scanResult = await scanProjectForTags();
      console.log(chalk.green(`🔍 스캔 완료: ${scanResult.foundTags.length}개 TAG 발견`));

      if (scanResult.orphanedTags.length > 0) {
        console.log(chalk.yellow(`⚠️  고아 TAG: ${scanResult.orphanedTags.length}개`));
      }

      if (scanResult.brokenReferences.length > 0) {
        console.log(chalk.red(`❌ 끊어진 참조: ${scanResult.brokenReferences.length}개`));
      }
    }

    // 수리 실행
    if (options.repair && scanResult) {
      repairedTags = await repairTagDatabase(database, scanResult);
      console.log(chalk.green(`🔧 수리 완료: ${repairedTags}개 TAG 수정`));
    }

    // 검증 실행
    if (options.validate) {
      const issues = await validateTagDatabase(database);
      if (issues.length > 0) {
        console.log(chalk.red('❌ 검증 실패:'));
        for (const issue of issues) {
          console.log(chalk.red(`  - ${issue}`));
        }
        if (!options.force) {
          throw new Error(`검증 실패: ${issues.length}개 문제 발견`);
        }
      } else {
        console.log(chalk.green('✅ 검증 통과'));
      }
    }

    // 정리 실행
    if (options.cleanup) {
      const beforeCount = Object.keys(database.tags).length;

      // 블록된 상태인 TAG들 제거
      const blockedTags = Object.entries(database.tags)
        .filter(([, tag]) => tag.status === 'blocked')
        .map(([id]) => id);

      for (const blockedId of blockedTags) {
        delete database.tags[blockedId];
      }

      removedTags = beforeCount - Object.keys(database.tags).length;
      console.log(chalk.blue(`🧹 정리 완료: ${removedTags}개 TAG 제거`));
    }

    // 데이터베이스 저장
    if (options.scan || options.repair || options.cleanup || options.force) {
      await saveTagDatabase(database);
      updatedTags = Object.keys(database.tags).length;
      console.log(chalk.green(`💾 데이터베이스 저장 완료: ${updatedTags}개 TAG`));
    }

    return {
      success: true,
      message: `TAG 업데이트 완료: ${updatedTags}개 TAG, ${repairedTags}개 수리, ${removedTags}개 제거`,
      scanResult,
      updatedTags,
      repairedTags,
      removedTags
    };

  } catch (error) {
    return {
      success: false,
      message: `TAG 업데이트 실패: ${error.message}`,
      updatedTags: 0,
      repairedTags: 0,
      removedTags: 0
    };
  }
}

program
  .name('tag-updater')
  .description('MoAI TAG 인덱스 업데이트 및 관리')
  .option('-s, --scan', '프로젝트 파일에서 TAG 스캔')
  .option('-r, --repair', 'TAG 데이터베이스 수리')
  .option('-v, --validate', 'TAG 데이터베이스 검증')
  .option('-c, --cleanup', '사용하지 않는 TAG 정리')
  .option('-b, --backup', '실행 전 백업 생성')
  .option('-f, --force', '검증 실패 시에도 강제 실행')
  .action(async (options: TagUpdaterOptions) => {
    try {
      console.log(chalk.blue('🏷️  TAG 업데이트 시작...'));

      const result = await updateTags(options);

      if (result.success) {
        console.log(chalk.green('✅'), result.message);
      } else {
        console.log(chalk.red('❌'), result.message);
      }

      // 스캔 결과 출력
      if (result.scanResult) {
        console.log(chalk.cyan('\n📊 스캔 결과:'));
        console.log(`  📁 스캔된 파일: ${result.scanResult.filesScanned}개`);
        console.log(`  🏷️  발견된 TAG: ${result.scanResult.foundTags.length}개`);
        if (result.scanResult.orphanedTags.length > 0) {
          console.log(chalk.yellow(`  👻 고아 TAG: ${result.scanResult.orphanedTags.length}개`));
        }
        if (result.scanResult.brokenReferences.length > 0) {
          console.log(chalk.red(`  🔗 끊어진 참조: ${result.scanResult.brokenReferences.length}개`));
        }
        if (result.scanResult.duplicateTags.length > 0) {
          console.log(chalk.yellow(`  🔄 중복 TAG: ${result.scanResult.duplicateTags.length}개`));
        }
      }

      console.log(JSON.stringify({
        success: result.success,
        stats: {
          updatedTags: result.updatedTags,
          repairedTags: result.repairedTags,
          removedTags: result.removedTags
        },
        scanResult: result.scanResult ? {
          filesScanned: result.scanResult.filesScanned,
          foundTags: result.scanResult.foundTags.length,
          orphanedTags: result.scanResult.orphanedTags.length,
          brokenReferences: result.scanResult.brokenReferences.length,
          duplicateTags: result.scanResult.duplicateTags.length
        } : null,
        nextSteps: result.success ? [
          'TAG 시스템이 업데이트되었습니다',
          '문서 동기화를 위해 doc-syncer를 실행하세요'
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