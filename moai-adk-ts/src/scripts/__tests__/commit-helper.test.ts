/**
 * @TEST-COMMIT-HELPER-001: CommitHelper TypeScript 포팅 테스트
 * @연결: @REQ:GIT-COMMIT-001 → @DESIGN:COMMIT-WORKFLOW-003 → @TASK:COMMIT-TS-PORT-001
 *
 * 이 테스트는 Python commit_helper.py의 기능을 TypeScript로 완전 포팅한
 * CommitHelper 클래스의 동작을 검증합니다.
 */

import { describe, test, expect, beforeEach, afterEach, vi } from 'vitest';
import { CommitHelper } from '../commit-helper';
import { GitWorkflow } from '../utils/git-workflow';
import { CommitValidator } from '../validators/commit-validator';
import { FileAnalyzer } from '../analyzers/file-analyzer';
import { MessageGenerator } from '../generators/message-generator';
import * as path from 'path';

// Mock 의존성들
vi.mock('../utils/git-workflow');
vi.mock('../validators/commit-validator');
vi.mock('../analyzers/file-analyzer');
vi.mock('../generators/message-generator');

describe('CommitHelper', () => {
  let commitHelper: CommitHelper;
  let mockGitWorkflow: vi.Mocked<GitWorkflow>;
  let mockValidator: vi.Mocked<CommitValidator>;
  let mockAnalyzer: vi.Mocked<FileAnalyzer>;
  let mockMessageGenerator: vi.Mocked<MessageGenerator>;

  beforeEach(() => {
    // Mock 인스턴스 생성
    mockGitWorkflow = new GitWorkflow() as vi.Mocked<GitWorkflow>;
    mockValidator = new CommitValidator() as vi.Mocked<CommitValidator>;
    mockAnalyzer = new FileAnalyzer() as vi.Mocked<FileAnalyzer>;
    mockMessageGenerator =
      new MessageGenerator() as vi.Mocked<MessageGenerator>;

    // CommitHelper 인스턴스 생성
    commitHelper = new CommitHelper();

    // 의존성 주입
    (commitHelper as any).gitWorkflow = mockGitWorkflow;
    (commitHelper as any).validator = mockValidator;
    (commitHelper as any).analyzer = mockAnalyzer;
    (commitHelper as any).messageGenerator = mockMessageGenerator;
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('constructor', () => {
    it('@TEST-INIT-001: should initialize with default project root and personal mode', () => {
      const helper = new CommitHelper();

      expect(helper).toBeDefined();
      expect((helper as any).mode).toBe('personal');
      expect((helper as any).projectRoot).toBeDefined();
    });

    it('@TEST-INIT-002: should initialize with custom project root', () => {
      const customRoot = '/custom/project/root';
      const helper = new CommitHelper(customRoot);

      expect((helper as any).projectRoot).toBe(path.resolve(customRoot));
    });
  });

  describe('getChangedFiles', () => {
    it('@TEST-CHANGED-FILES-001: should return changed files with analysis', async () => {
      // Arrange
      const mockGitStatus = 'M  src/test.ts\nA  src/new.ts\n';
      mockGitWorkflow.runCommand = vi.fn().mockResolvedValue({
        stdout: mockGitStatus,
        stderr: '',
        exitCode: 0,
      });

      mockAnalyzer.classifyFileChange = vi
        .fn()
        .mockReturnValueOnce('modified')
        .mockReturnValueOnce('added');

      mockAnalyzer.analyzeFileChanges = vi.fn().mockReturnValue({
        total_files: 2,
        categories: {
          python: 0,
          typescript: 2,
          docs: 0,
          config: 0,
          tests: 0,
          other: 0,
        },
        is_simple_change: true,
        has_tests: false,
        is_documentation_only: false,
      });

      // Act
      const result = await commitHelper.getChangedFiles();

      // Assert
      expect(result.success).toBe(true);
      expect(result.files).toHaveLength(2);
      expect(result.files?.[0]).toEqual({
        status: 'M ',
        filename: 'src/test.ts',
        type: 'modified',
      });
      expect(result.files?.[1]).toEqual({
        status: 'A ',
        filename: 'src/new.ts',
        type: 'added',
      });
      expect(result.count).toBe(2);
      expect(result.has_changes).toBe(true);
      expect(result.analysis).toBeDefined();
    });

    it('@TEST-CHANGED-FILES-002: should handle no changes', async () => {
      // Arrange
      mockGitWorkflow.runCommand = vi.fn().mockResolvedValue({
        stdout: '',
        stderr: '',
        exitCode: 0,
      });

      mockAnalyzer.analyzeFileChanges = vi.fn().mockReturnValue({
        total_files: 0,
        categories: {
          python: 0,
          typescript: 0,
          docs: 0,
          config: 0,
          tests: 0,
          other: 0,
        },
        is_simple_change: true,
        has_tests: false,
        is_documentation_only: false,
      });

      // Act
      const result = await commitHelper.getChangedFiles();

      // Assert
      expect(result.success).toBe(true);
      expect(result.files).toHaveLength(0);
      expect(result.count).toBe(0);
      expect(result.has_changes).toBe(false);
    });

    it('@TEST-CHANGED-FILES-003: should handle git command errors', async () => {
      // Arrange
      const errorMessage = 'Git command failed';
      mockGitWorkflow.runCommand = vi
        .fn()
        .mockRejectedValue(new Error(errorMessage));

      // Act
      const result = await commitHelper.getChangedFiles();

      // Assert
      expect(result.success).toBe(false);
      expect(result.error).toBe(errorMessage);
    });
  });

  describe('createSmartCommit', () => {
    it('@TEST-SMART-COMMIT-001: should create smart commit with auto-generated message', async () => {
      // Arrange
      const mockChanges = {
        success: true,
        files: [{ status: 'M ', filename: 'src/test.ts', type: 'modified' }],
        count: 1,
        has_changes: true,
        analysis: { is_simple_change: true },
      };

      vi.spyOn(commitHelper, 'getChangedFiles').mockResolvedValue(mockChanges);

      mockValidator.validateChangeContext = vi
        .fn()
        .mockReturnValue({ valid: true });
      mockValidator.validateFileList = vi.fn().mockReturnValue({ valid: true });
      mockMessageGenerator.generateSmartMessage = vi
        .fn()
        .mockReturnValue('feat: update test functionality');
      mockGitWorkflow.createConstitutionCommit = vi
        .fn()
        .mockResolvedValue('abc123def');

      // Act
      const result = await commitHelper.createSmartCommit();

      // Assert
      expect(result.success).toBe(true);
      expect(result.commit_hash).toBe('abc123def');
      expect(result.message).toBe('feat: update test functionality');
      expect(result.files_changed).toBe(1);
      expect(result.mode).toBe('personal');
    });

    it('@TEST-SMART-COMMIT-002: should use provided custom message', async () => {
      // Arrange
      const customMessage = 'custom: my commit message';
      const mockChanges = {
        success: true,
        files: [{ status: 'M ', filename: 'src/test.ts', type: 'modified' }],
        count: 1,
        has_changes: true,
        analysis: { is_simple_change: true },
      };

      vi.spyOn(commitHelper, 'getChangedFiles').mockResolvedValue(mockChanges);

      mockValidator.validateChangeContext = vi
        .fn()
        .mockReturnValue({ valid: true });
      mockValidator.validateFileList = vi.fn().mockReturnValue({ valid: true });
      mockGitWorkflow.createConstitutionCommit = vi
        .fn()
        .mockResolvedValue('abc123def');

      // Act
      const result = await commitHelper.createSmartCommit(customMessage);

      // Assert
      expect(result.success).toBe(true);
      expect(result.message).toBe(customMessage);
      expect(mockMessageGenerator.generateSmartMessage).not.toHaveBeenCalled();
    });

    it('@TEST-SMART-COMMIT-003: should fail when no changes exist', async () => {
      // Arrange
      const mockChanges = {
        success: true,
        files: [],
        count: 0,
        has_changes: false,
        analysis: {},
      };

      vi.spyOn(commitHelper, 'getChangedFiles').mockResolvedValue(mockChanges);
      mockValidator.validateChangeContext = vi.fn().mockReturnValue({
        valid: false,
        reason: 'No changes to commit',
      });

      // Act
      const result = await commitHelper.createSmartCommit();

      // Assert
      expect(result.success).toBe(false);
      expect(result.error).toBe('No changes to commit');
    });
  });

  describe('createConstitutionCommit', () => {
    it('@TEST-CONSTITUTION-COMMIT-001: should create commit with validated message', async () => {
      // Arrange
      const message = 'feat: add new feature';
      const files = ['src/feature.ts'];

      mockValidator.validateCommitMessage = vi.fn().mockReturnValue({
        valid: true,
        reason: '검증 통과',
      });
      mockValidator.validateFileList = vi.fn().mockReturnValue({
        valid: true,
        reason: '파일 목록 유효',
      });
      mockGitWorkflow.createConstitutionCommit = vi
        .fn()
        .mockResolvedValue('def456abc');

      // Act
      const result = await commitHelper.createConstitutionCommit(
        message,
        files
      );

      // Assert
      expect(result.success).toBe(true);
      expect(result.commit_hash).toBe('def456abc');
      expect(result.message).toBe(message);
      expect(result.validation).toEqual({ valid: true, reason: '검증 통과' });
      expect(result.file_validation).toEqual({
        valid: true,
        reason: '파일 목록 유효',
      });
    });

    it('@TEST-CONSTITUTION-COMMIT-002: should fail with invalid message', async () => {
      // Arrange
      const invalidMessage = 'bad';

      mockValidator.validateCommitMessage = vi.fn().mockReturnValue({
        valid: false,
        reason: '메시지가 너무 짧습니다',
      });

      // Act
      const result =
        await commitHelper.createConstitutionCommit(invalidMessage);

      // Assert
      expect(result.success).toBe(false);
      expect(result.error).toBe('메시지 검증 실패: 메시지가 너무 짧습니다');
    });
  });

  describe('suggestCommitMessage', () => {
    it('@TEST-SUGGEST-MESSAGE-001: should provide multiple commit message suggestions', async () => {
      // Arrange
      const mockChanges = {
        success: true,
        files: [{ status: 'M ', filename: 'src/test.ts', type: 'modified' }],
        count: 1,
        has_changes: true,
        analysis: {
          summary: { modified: 1, added: 0, deleted: 0 },
          is_simple_change: true,
        },
      };

      vi.spyOn(commitHelper, 'getChangedFiles').mockResolvedValue(mockChanges);

      mockValidator.validateChangeContext = vi
        .fn()
        .mockReturnValue({ valid: true });
      mockMessageGenerator.generateSmartMessage = vi
        .fn()
        .mockReturnValue('feat: update test functionality');
      mockMessageGenerator.calculateConfidence = vi.fn().mockReturnValue(0.9);
      mockMessageGenerator.generateContextMessage = vi
        .fn()
        .mockReturnValue('feat(test): update test functionality with context');
      mockMessageGenerator.generateTemplateSuggestions = vi
        .fn()
        .mockReturnValue([
          {
            type: 'template',
            message: 'feat: add new feature',
            confidence: 0.7,
          },
          { type: 'template', message: 'fix: resolve bug', confidence: 0.6 },
        ]);

      // Act
      const result = await commitHelper.suggestCommitMessage('test context');

      // Assert
      expect(result.success).toBe(true);
      expect(result.suggestions).toHaveLength(4); // auto + context + 2 templates
      expect(result.suggestions?.[0]).toEqual({
        type: 'auto',
        message: 'feat: update test functionality',
        confidence: 0.9,
      });
      expect(result.suggestions?.[1]).toEqual({
        type: 'context',
        message: 'feat(test): update test functionality with context',
        confidence: 0.8,
      });
      expect(result.files_changed).toBe(1);
      expect(result.change_summary).toEqual({
        modified: 1,
        added: 0,
        deleted: 0,
      });
    });

    it('@TEST-SUGGEST-MESSAGE-002: should handle context-free suggestions', async () => {
      // Arrange
      const mockChanges = {
        success: true,
        files: [{ status: 'A ', filename: 'src/new.ts', type: 'added' }],
        count: 1,
        has_changes: true,
        analysis: {
          summary: { modified: 0, added: 1, deleted: 0 },
          is_simple_change: true,
        },
      };

      vi.spyOn(commitHelper, 'getChangedFiles').mockResolvedValue(mockChanges);

      mockValidator.validateChangeContext = vi
        .fn()
        .mockReturnValue({ valid: true });
      mockMessageGenerator.generateSmartMessage = vi
        .fn()
        .mockReturnValue('feat: add new file');
      mockMessageGenerator.calculateConfidence = vi.fn().mockReturnValue(0.8);
      mockMessageGenerator.generateTemplateSuggestions = vi
        .fn()
        .mockReturnValue([]);

      // Act
      const result = await commitHelper.suggestCommitMessage(); // no context

      // Assert
      expect(result.success).toBe(true);
      expect(result.suggestions).toHaveLength(1); // only auto suggestion
      expect(result.suggestions?.[0]?.type).toBe('auto');
    });
  });
});
