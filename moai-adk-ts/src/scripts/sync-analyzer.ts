#!/usr/bin/env tsx
// @FEATURE:SYNC-001 | Chain: @REQ:SYNC-001 -> @DESIGN:SYNC-001 -> @TASK:SYNC-001 -> @TEST:SYNC-001
// Related: @DOCS:SYNC-001

/**
 * sync-analyzer.ts - 문서 동기화 전 프로젝트 상태 분석 스크립트
 *
 * 기능:
 * - Git 상태 분석 (변경된 파일, 브랜치)
 * - 문서 신선도 체크 (오래된 문서 감지)
 * - TAG 시스템 상태 확인
 * - 자동 실행 가능성 판단
 * - 구조화된 JSON 결과 출력
 */

import { execSync } from 'child_process';
import { existsSync, statSync } from 'fs';
import { logger } from '../utils/winston-logger.js';

// 타입 정의
interface SyncAnalyzerOptions {
  mode?: 'auto' | 'interactive' | 'force' | 'status';
  target?: string;
  skipConfirmation?: boolean;
}

interface ProjectStatus {
  changedFiles: number;
  uncommittedDocs: number;
  outdatedDocs: number;
  tagFiles: number;
  currentBranch: string;
}

interface SyncPlan {
  mode: string;
  scope: 'full' | 'partial' | 'selective';
  target: string;
  estimatedTime: string;
  risks: string[];
}

interface Recommendations {
  action: 'proceed' | 'review' | 'abort';
  reason: string;
  nextSteps: string[];
}

interface SyncAnalysisResult {
  projectStatus: ProjectStatus;
  autoSyncSafe: boolean;
  syncMode: 'auto' | 'interactive' | 'force' | 'status';
  skipConfirmation: boolean;
  syncPlan: SyncPlan;
  recommendations: Recommendations;
}

// 유틸리티 함수
function execCommand(command: string): string {
  try {
    return execSync(command, { encoding: 'utf-8' }).trim();
  } catch (error) {
    return '';
  }
}

function countLines(output: string): number {
  if (!output) return 0;
  return output.split('\n').filter(line => line.trim()).length;
}

// Git 상태 분석
function analyzeGitStatus(): ProjectStatus {
  // 변경된 파일 수
  const gitStatus = execCommand('git status --porcelain');
  const changedFiles = countLines(gitStatus);

  // 미커밋 문서 수
  const uncommittedDocsOutput = execCommand('git status --porcelain -- "*.md" "docs/"');
  const uncommittedDocs = countLines(uncommittedDocsOutput);

  // 현재 브랜치
  const currentBranch = execCommand('git rev-parse --abbrev-ref HEAD') || 'unknown';

  // 오래된 문서 수 (sync-report.md 대비)
  let outdatedDocs = 0;
  const syncReportPath = '.moai/reports/sync-report.md';
  if (existsSync(syncReportPath)) {
    const syncReportTime = statSync(syncReportPath).mtimeMs;
    const docsToCheck = ['README.md', 'CLAUDE.md', 'docs'];

    docsToCheck.forEach(path => {
      if (existsSync(path)) {
        const stat = statSync(path);
        if (stat.isFile() && stat.mtimeMs < syncReportTime) {
          outdatedDocs++;
        } else if (stat.isDirectory()) {
          // 디렉토리 내 문서 체크는 간소화
          outdatedDocs += 0;
        }
      }
    });
  }

  // TAG 파일 수 - CODE-FIRST: 인덱스 불필요
  // NOTE: [v0.0.3+] .moai/indexes 제거 - 코드 직접 스캔 방식으로 전환
  const tagFiles = 0; // 더 이상 인덱스 파일 사용하지 않음

  return {
    changedFiles,
    uncommittedDocs,
    outdatedDocs,
    tagFiles,
    currentBranch,
  };
}

// 자동 실행 가능성 판단
function isAutoSyncSafe(status: ProjectStatus): boolean {
  // 안전한 변경 임계값: <5 파일, <3 문서, <3 오래된 문서
  return (
    status.changedFiles < 5 &&
    status.uncommittedDocs < 3 &&
    status.outdatedDocs < 3
  );
}

// 동기화 계획 생성
function createSyncPlan(
  options: SyncAnalyzerOptions,
  status: ProjectStatus,
  autoSafe: boolean
): SyncPlan {
  const mode = options.mode || 'auto';
  const target = options.target || '.';

  let scope: 'full' | 'partial' | 'selective' = 'full';
  let estimatedTime = '15-25초';
  const risks: string[] = [];

  // 모드별 계획 조정
  switch (mode) {
    case 'auto':
      if (autoSafe) {
        scope = 'partial';
        estimatedTime = '5-10초';
      } else {
        scope = 'full';
        estimatedTime = '15-25초';
        risks.push('중간 규모 변경 감지 - 수동 확인 권장');
      }
      break;

    case 'force':
      scope = 'full';
      estimatedTime = '20-30초';
      risks.push('강제 전체 동기화 - 모든 문서 재생성');
      break;

    case 'status':
      scope = 'selective';
      estimatedTime = '2-3초';
      break;

    case 'interactive':
      scope = 'full';
      estimatedTime = '15-25초';
      break;
  }

  // 추가 위험 요소 평가
  if (status.changedFiles > 10) {
    risks.push('대량 파일 변경 감지 (10개 이상)');
  }

  if (status.uncommittedDocs > 5) {
    risks.push('다수 문서 미커밋 (5개 이상)');
  }

  if (status.tagFiles === 0) {
    risks.push('TAG 인덱스 파일 없음 - 초기화 필요');
  }

  return {
    mode,
    scope,
    target,
    estimatedTime,
    risks,
  };
}

// 권장사항 생성
function createRecommendations(
  status: ProjectStatus,
  autoSafe: boolean,
  plan: SyncPlan
): Recommendations {
  // abort 조건 확인
  if (status.changedFiles > 50) {
    return {
      action: 'abort',
      reason: '너무 많은 파일 변경 (50개 이상) - 수동 확인 필요',
      nextSteps: [
        'git status로 변경 내역 확인',
        '필요한 변경만 선택적으로 커밋',
        '변경 범위를 줄인 후 재시도',
      ],
    };
  }

  // review 조건 확인
  if (!autoSafe && plan.mode === 'auto') {
    return {
      action: 'review',
      reason: '중간 규모 변경 감지 - 동기화 계획 검토 권장',
      nextSteps: [
        '변경된 파일 목록 확인',
        '동기화 범위 및 예상 시간 검토',
        '승인 후 동기화 진행',
      ],
    };
  }

  // proceed 조건
  if (autoSafe) {
    return {
      action: 'proceed',
      reason: '안전한 변경 감지 - 자동 동기화 가능',
      nextSteps: [
        '즉시 문서 동기화 시작',
        'Phase 2 doc-syncer 실행',
        '고속 모드로 5-10초 내 완료',
      ],
    };
  }

  // 기본 proceed
  return {
    action: 'proceed',
    reason: '동기화 준비 완료',
    nextSteps: [
      '사용자 승인 후 동기화 시작',
      'Phase 2 doc-syncer 실행',
      '완전 모드로 15-25초 내 완료',
    ],
  };
}

// 메인 분석 함수
function analyzeSyncStatus(options: SyncAnalyzerOptions): SyncAnalysisResult {
  // 1. Git 상태 분석
  const projectStatus = analyzeGitStatus();

  // 2. 자동 실행 가능성 판단
  const autoSyncSafe = isAutoSyncSafe(projectStatus);

  // 3. 모드 결정
  const syncMode = options.mode || 'auto';
  let skipConfirmation = options.skipConfirmation ?? false;

  // auto 모드에서 안전하면 확인 스킵
  if (syncMode === 'auto' && autoSyncSafe) {
    skipConfirmation = true;
  }

  // status 모드는 항상 확인 스킵
  if (syncMode === 'status') {
    skipConfirmation = true;
  }

  // 4. 동기화 계획 생성
  const syncPlan = createSyncPlan(options, projectStatus, autoSyncSafe);

  // 5. 권장사항 생성
  const recommendations = createRecommendations(projectStatus, autoSyncSafe, syncPlan);

  return {
    projectStatus,
    autoSyncSafe,
    syncMode,
    skipConfirmation,
    syncPlan,
    recommendations,
  };
}

// CLI 인터페이스
function main() {
  // 명령줄 인수 파싱
  const args = process.argv.slice(2);
  const options: SyncAnalyzerOptions = {
    mode: 'auto',
    target: '.',
    skipConfirmation: false,
  };

  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    if (arg === '--mode' && args[i + 1]) {
      options.mode = args[i + 1] as 'auto' | 'interactive' | 'force' | 'status';
      i++;
    } else if (arg === '--target' && args[i + 1]) {
      options.target = args[i + 1] || '.';
      i++;
    } else if (arg === '--skip-confirmation') {
      options.skipConfirmation = true;
    }
  }

  // 분석 실행
  const result = analyzeSyncStatus(options);

  // JSON 출력
  logger.info(JSON.stringify(result, null, 2));

  // 종료 코드 결정
  if (result.recommendations.action === 'abort') {
    process.exit(1);
  }

  process.exit(0);
}

// 스크립트 실행
if (require.main === module) {
  main();
}

export { analyzeSyncStatus, SyncAnalyzerOptions, SyncAnalysisResult };