/**
 * @FEATURE-MESSAGE-GENERATOR-001: TypeScript MessageGenerator 구현
 * @연결: @TASK:MESSAGE-GEN-001 → @FEATURE:SMART-MESSAGE-001 → @API:GENERATE-MESSAGE
 */

import type { FileChange } from '../commit-helper';

export interface MessageSuggestion {
  type: string;
  message: string;
  confidence: number;
}

export class MessageGenerator {
  private readonly conventionalTypes = [
    { type: 'feat', description: '새로운 기능 추가' },
    { type: 'fix', description: '버그 수정' },
    { type: 'docs', description: '문서 변경' },
    { type: 'style', description: '코드 스타일 변경' },
    { type: 'refactor', description: '리팩토링' },
    { type: 'test', description: '테스트 추가/수정' },
    { type: 'chore', description: '기타 변경사항' },
  ];

  generateSmartMessage(files: FileChange[]): string {
    if (files.length === 0) {
      return 'chore: empty commit';
    }

    const analysis = this.analyzeChanges(files);
    const type = this.inferCommitType(analysis);
    const scope = this.inferScope(files);
    const description = this.generateDescription(analysis, files);

    if (scope) {
      return `${type}(${scope}): ${description}`;
    }

    return `${type}: ${description}`;
  }

  generateContextMessage(context: string, files: FileChange[]): string {
    const baseMessage = this.generateSmartMessage(files);
    const cleanContext = context.trim().toLowerCase();

    // 컨텍스트에서 스코프 추출 시도
    if (cleanContext) {
      const [type, rest] = baseMessage.split(': ');
      return `${type}(${cleanContext}): ${rest}`;
    }

    return baseMessage;
  }

  generateTemplateSuggestions(): MessageSuggestion[] {
    return this.conventionalTypes.map(item => ({
      type: 'template',
      message: `${item.type}: ${item.description}`,
      confidence: 0.5,
    }));
  }

  calculateConfidence(files: FileChange[]): number {
    if (files.length === 0) {
      return 0.1;
    }

    let confidence = 0.5; // 기본 신뢰도

    // 단순 변경 (파일 수가 적을수록 높은 신뢰도)
    if (files.length === 1) {
      confidence += 0.3;
    } else if (files.length <= 3) {
      confidence += 0.2;
    } else if (files.length <= 5) {
      confidence += 0.1;
    }

    // 동일한 유형의 변경이 많을수록 높은 신뢰도
    const changeTypes = new Set(files.map(f => f.type));
    if (changeTypes.size === 1) {
      confidence += 0.2;
    }

    // 확장자 일관성
    const extensions = new Set(
      files.map(f => this.getFileExtension(f.filename))
    );
    if (extensions.size === 1) {
      confidence += 0.1;
    }

    return Math.min(confidence, 1.0);
  }

  private analyzeChanges(files: FileChange[]): {
    added: number;
    modified: number;
    deleted: number;
    renamed: number;
    hasTests: boolean;
    hasDocs: boolean;
    hasConfig: boolean;
  } {
    const analysis = {
      added: 0,
      modified: 0,
      deleted: 0,
      renamed: 0,
      hasTests: false,
      hasDocs: false,
      hasConfig: false,
    };

    for (const file of files) {
      switch (file.type) {
        case 'added':
          analysis.added++;
          break;
        case 'modified':
          analysis.modified++;
          break;
        case 'deleted':
          analysis.deleted++;
          break;
        case 'renamed':
          analysis.renamed++;
          break;
      }

      const filename = file.filename.toLowerCase();
      if (filename.includes('test') || filename.includes('spec')) {
        analysis.hasTests = true;
      }
      if (filename.endsWith('.md') || filename.includes('doc')) {
        analysis.hasDocs = true;
      }
      if (filename.includes('config') || filename.includes('package.json')) {
        analysis.hasConfig = true;
      }
    }

    return analysis;
  }

  private inferCommitType(analysis: any): string {
    if (
      analysis.hasDocs &&
      !analysis.hasTests &&
      analysis.added === 0 &&
      analysis.modified > 0
    ) {
      return 'docs';
    }

    if (analysis.hasTests && analysis.added === 0) {
      return 'test';
    }

    if (analysis.hasConfig) {
      return 'chore';
    }

    if (analysis.added > 0) {
      return 'feat';
    }

    if (analysis.modified > 0) {
      return 'fix';
    }

    if (analysis.deleted > 0) {
      return 'refactor';
    }

    return 'chore';
  }

  private inferScope(files: FileChange[]): string | null {
    if (files.length === 0) {
      return null;
    }

    // 공통 디렉토리 추출
    const directories = files
      .map(f => f.filename.split('/')[0])
      .filter(dir => dir && dir.includes('/') === false); // 루트 파일 제외

    if (directories.length === 0) {
      return null;
    }

    // 가장 빈번한 디렉토리
    const dirCounts = directories.reduce(
      (acc, dir) => {
        if (dir) {
          acc[dir] = (acc[dir] || 0) + 1;
        }
        return acc;
      },
      {} as Record<string, number>
    );

    const mostCommonDir = Object.entries(dirCounts).sort(
      ([, a], [, b]) => b - a
    )[0];

    // 모든 파일이 같은 디렉토리에 있으면 스코프로 사용
    if (mostCommonDir && mostCommonDir[1] === files.length) {
      return mostCommonDir[0];
    }

    return null;
  }

  private generateDescription(analysis: any, files: FileChange[]): string {
    if (analysis.hasDocs) {
      return 'update documentation';
    }

    if (analysis.hasTests) {
      return 'update tests';
    }

    if (analysis.hasConfig) {
      return 'update configuration';
    }

    if (analysis.added > 0 && analysis.modified === 0) {
      return `add ${files.length} new file${files.length > 1 ? 's' : ''}`;
    }

    if (analysis.modified > 0 && analysis.added === 0) {
      return `update ${files.length} file${files.length > 1 ? 's' : ''}`;
    }

    if (analysis.deleted > 0) {
      return `remove ${analysis.deleted} file${analysis.deleted > 1 ? 's' : ''}`;
    }

    return 'update project files';
  }

  private getFileExtension(filename: string): string {
    const lastDot = filename.lastIndexOf('.');
    return lastDot !== -1 ? filename.substring(lastDot) : '';
  }
}
