/**
 * @FEATURE-COMMIT-HELPER-001: TypeScript CommitHelper 구현
 * @연결: @REQ:GIT-COMMIT-001 → @DESIGN:COMMIT-WORKFLOW-003 → @TASK:COMMIT-TS-PORT-001
 *
 * MoAI 커밋 도우미 v1.0.0 (TypeScript 포팅)
 * 자동 메시지 생성 및 TRUST 원칙 준수 커밋 시스템
 */

import * as path from 'path';
import { GitWorkflow } from './utils/git-workflow';
import { CommitValidator } from './validators/commit-validator';
import { FileAnalyzer } from './analyzers/file-analyzer';
import { MessageGenerator } from './generators/message-generator';
import { ProjectHelper } from './utils/project-helper';

export interface FileChange {
  status: string;
  filename: string;
  type: string;
}

export interface ChangedFilesResult {
  success: boolean;
  files?: FileChange[];
  count?: number;
  has_changes?: boolean;
  analysis?: any;
  error?: string;
}

export interface CommitResult {
  success: boolean;
  commit_hash?: string;
  message?: string;
  files_changed?: number;
  mode?: string;
  analysis?: any;
  validation?: any;
  file_validation?: any;
  error?: string;
}

export interface SuggestionResult {
  success: boolean;
  suggestions?: Array<{
    type: string;
    message: string;
    confidence: number;
  }>;
  files_changed?: number;
  change_summary?: any;
  analysis?: any;
  error?: string;
}

export class CommitHelper {
  private projectRoot: string;
  private gitWorkflow: GitWorkflow;
  private validator: CommitValidator;
  private analyzer: FileAnalyzer;
  private messageGenerator: MessageGenerator;
  private config: any;
  private mode: string;

  constructor(projectRoot?: string) {
    this.projectRoot = projectRoot
      ? path.resolve(projectRoot)
      : this.findProjectRoot();
    this.gitWorkflow = new GitWorkflow(this.projectRoot);
    this.validator = new CommitValidator();
    this.analyzer = new FileAnalyzer();
    this.messageGenerator = new MessageGenerator();

    this.config = ProjectHelper.loadConfig(this.projectRoot);
    this.mode = this.config?.mode || 'personal';
  }

  private findProjectRoot(): string {
    // 현재 디렉토리에서 시작해서 상위로 올라가며 .git 디렉토리를 찾음
    let currentDir = process.cwd();
    const root = path.parse(currentDir).root;

    while (currentDir !== root) {
      try {
        const gitDir = path.join(currentDir, '.git');
        if (require('fs').existsSync(gitDir)) {
          return currentDir;
        }
      } catch (error) {
        // 무시하고 계속
      }
      currentDir = path.dirname(currentDir);
    }

    // .git을 찾지 못하면 현재 디렉토리 반환
    return process.cwd();
  }

  /**
   * @API-GET-CHANGED-FILES-001: 변경된 파일 목록 조회
   */
  async getChangedFiles(): Promise<ChangedFilesResult> {
    try {
      const result = await this.gitWorkflow.runCommand([
        'git',
        'status',
        '--porcelain',
      ]);

      const changedFiles: FileChange[] = [];
      for (const line of result.stdout.split('\n')) {
        if (line.trim()) {
          const status = line.slice(0, 2);
          const filename = line.slice(3).trim();
          changedFiles.push({
            status,
            filename,
            type: this.analyzer.classifyFileChange(status),
          });
        }
      }

      const analysis = this.analyzer.analyzeFileChanges(changedFiles);

      return {
        success: true,
        files: changedFiles,
        count: changedFiles.length,
        has_changes: changedFiles.length > 0,
        analysis,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  /**
   * @API-CREATE-SMART-COMMIT-001: 스마트 커밋 생성
   */
  async createSmartCommit(
    message?: string,
    files?: string[]
  ): Promise<CommitResult> {
    try {
      // 변경사항 확인 및 검증
      const changes = await this.getChangedFiles();
      if (!changes.success) {
        return changes as CommitResult;
      }

      const changeValidation = this.validator.validateChangeContext(changes);
      if (!changeValidation.valid) {
        return { success: false, error: changeValidation.reason };
      }

      // 파일 목록 검증
      const fileValidation = this.validator.validateFileList(files);
      if (!fileValidation.valid) {
        return { success: false, error: fileValidation.reason };
      }

      // 메시지가 없으면 자동 생성
      let commitMessage = message;
      if (!commitMessage) {
        commitMessage = this.messageGenerator.generateSmartMessage(
          changes.files || []
        );
      }

      // 커밋 실행
      const commitHash = await this.gitWorkflow.createConstitutionCommit(
        commitMessage,
        files
      );

      return {
        success: true,
        commit_hash: commitHash,
        message: commitMessage,
        files_changed: changes.count || 0,
        mode: this.mode,
        analysis: changes.analysis,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  /**
   * @API-CREATE-CONSTITUTION-COMMIT-001: TRUST 원칙 기반 커밋 생성
   */
  async createConstitutionCommit(
    message: string,
    files?: string[]
  ): Promise<CommitResult> {
    try {
      // 메시지 검증
      const validation = this.validator.validateCommitMessage(message);
      if (!validation.valid) {
        return {
          success: false,
          error: `메시지 검증 실패: ${validation.reason}`,
        };
      }

      // 파일 목록 검증
      const fileValidation = this.validator.validateFileList(files);
      if (!fileValidation.valid) {
        return { success: false, error: fileValidation.reason };
      }

      // 커밋 실행
      const commitHash = await this.gitWorkflow.createConstitutionCommit(
        message,
        files
      );

      return {
        success: true,
        commit_hash: commitHash,
        message,
        validation,
        file_validation: fileValidation,
        mode: this.mode,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  /**
   * @API-SUGGEST-COMMIT-MESSAGE-001: 커밋 메시지 제안
   */
  async suggestCommitMessage(context?: string): Promise<SuggestionResult> {
    try {
      const changes = await this.getChangedFiles();
      if (!changes.success) {
        return changes as SuggestionResult;
      }

      const changeValidation = this.validator.validateChangeContext(changes);
      if (!changeValidation.valid) {
        return { success: false, error: changeValidation.reason };
      }

      const suggestions: Array<{
        type: string;
        message: string;
        confidence: number;
      }> = [];

      // 파일 변경 기반 제안
      const autoMessage = this.messageGenerator.generateSmartMessage(
        changes.files || []
      );
      suggestions.push({
        type: 'auto',
        message: autoMessage,
        confidence: this.messageGenerator.calculateConfidence(
          changes.files || []
        ),
      });

      // 컨텍스트 기반 제안
      if (context) {
        const contextMessage = this.messageGenerator.generateContextMessage(
          context,
          changes.files || []
        );
        suggestions.push({
          type: 'context',
          message: contextMessage,
          confidence: 0.8,
        });
      }

      // 템플릿 기반 제안
      const templateSuggestions =
        this.messageGenerator.generateTemplateSuggestions();
      suggestions.push(...templateSuggestions);

      return {
        success: true,
        suggestions,
        files_changed: changes.count || 0,
        change_summary: changes.analysis?.summary || {},
        analysis: changes.analysis,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }
}
