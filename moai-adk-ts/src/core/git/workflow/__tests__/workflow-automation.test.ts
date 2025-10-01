// @TEST:GIT-004 | Chain: @SPEC:GIT-001 -> @SPEC:GIT-001 -> @CODE:GIT-004 -> @TEST:GIT-004
// Related: @CODE:GIT-004, @CODE:GIT-004

/**
 * @file Workflow Automation Integration Tests
 * @author MoAI Team
 *
 * @fileoverview 워크플로우 자동화 통합 테스트
 */

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { GitConfig } from '../../../../types/git';
import type { GitManager } from '../../git-manager';
import { WorkflowAutomation } from '../index';

describe('WorkflowAutomation Integration Tests', () => {
  let mockGitManager: GitManager;
  let config: GitConfig;
  let workflow: WorkflowAutomation;

  beforeEach(() => {
    // Mock GitManager
    mockGitManager = {
      createBranch: vi.fn().mockResolvedValue(undefined),
      commitChanges: vi.fn().mockResolvedValue({ hash: 'abc123' }),
      createCheckpoint: vi.fn().mockResolvedValue(undefined),
      createPullRequest: vi.fn().mockResolvedValue('https://github.com/pr/1'),
    } as unknown as GitManager;

    config = {
      mode: 'solo',
      branchStrategy: 'feature',
      commitStyle: 'conventional',
    } as GitConfig;

    workflow = new WorkflowAutomation(mockGitManager, config);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('@TEST:GIT-004-SPEC - SPEC 워크플로우', () => {
    it('should successfully start spec workflow in solo mode', async () => {
      const result = await workflow.startSpecWorkflow(
        'SPEC-001',
        'Test feature'
      );

      expect(result.success).toBe(true);
      expect(result.stage).toBe('spec');
      expect(result.branchName).toBe('spec/SPEC-001');
      expect(result.commitHash).toBe('abc123');
      expect(mockGitManager.createBranch).toHaveBeenCalledWith(
        'spec/SPEC-001',
        'main'
      );
    });

    it('should create draft PR in team mode', async () => {
      workflow = new WorkflowAutomation(mockGitManager, {
        ...config,
        mode: 'team',
      });

      const result = await workflow.startSpecWorkflow(
        'SPEC-002',
        'Team feature'
      );

      expect(result.success).toBe(true);
      expect(result.pullRequestUrl).toBe('https://github.com/pr/1');
      expect(mockGitManager.createPullRequest).toHaveBeenCalled();
    });

    it('should handle errors gracefully', async () => {
      (mockGitManager.createBranch as ReturnType<typeof vi.fn>).mockRejectedValue(
        new Error('Git error')
      );

      const result = await workflow.startSpecWorkflow(
        'SPEC-003',
        'Failed feature'
      );

      expect(result.success).toBe(false);
      expect(result.stage).toBe('init');
      expect(result.message).toContain('Failed to start SPEC workflow');
    });
  });

  describe('@TEST:GIT-004-BUILD - Build 워크플로우', () => {
    it('should complete TDD build workflow', async () => {
      const result = await workflow.runBuildWorkflow('SPEC-001');

      expect(result.success).toBe(true);
      expect(result.stage).toBe('build');
      expect(result.commitHash).toBe('abc123');
      expect(mockGitManager.createCheckpoint).toHaveBeenCalledTimes(3);
      expect(mockGitManager.createCheckpoint).toHaveBeenCalledWith(
        'SPEC-001 TDD RED phase - Tests written'
      );
      expect(mockGitManager.createCheckpoint).toHaveBeenCalledWith(
        'SPEC-001 TDD GREEN phase - Tests passing'
      );
      expect(mockGitManager.createCheckpoint).toHaveBeenCalledWith(
        'SPEC-001 TDD REFACTOR phase - Code optimized'
      );
    });

    it('should handle build failures', async () => {
      (mockGitManager.createCheckpoint as ReturnType<typeof vi.fn>).mockRejectedValue(
        new Error('Checkpoint failed')
      );

      const result = await workflow.runBuildWorkflow('SPEC-001');

      expect(result.success).toBe(false);
      expect(result.stage).toBe('build');
    });
  });

  describe('@TEST:GIT-004-SYNC - Sync 워크플로우', () => {
    it('should complete sync workflow', async () => {
      const result = await workflow.runSyncWorkflow('SPEC-001');

      expect(result.success).toBe(true);
      expect(result.stage).toBe('sync');
      expect(result.commitHash).toBe('abc123');
      expect(mockGitManager.commitChanges).toHaveBeenCalled();
    });

    it('should handle sync failures', async () => {
      (mockGitManager.commitChanges as ReturnType<typeof vi.fn>).mockRejectedValue(
        new Error('Commit failed')
      );

      const result = await workflow.runSyncWorkflow('SPEC-001');

      expect(result.success).toBe(false);
      expect(result.stage).toBe('sync');
    });
  });

  describe('@TEST:GIT-004-FULL - 전체 워크플로우', () => {
    it('should run full workflow successfully', async () => {
      const results = await workflow.runFullSpecWorkflow(
        'SPEC-001',
        'Full test'
      );

      expect(results).toHaveLength(3);
      expect(results[0].stage).toBe('spec');
      expect(results[1].stage).toBe('build');
      expect(results[2].stage).toBe('sync');
      expect(results.every(r => r.success)).toBe(true);
    });

    it('should stop on first failure', async () => {
      (mockGitManager.createBranch as ReturnType<typeof vi.fn>).mockRejectedValue(
        new Error('Branch failed')
      );

      const results = await workflow.runFullSpecWorkflow(
        'SPEC-001',
        'Failed test'
      );

      expect(results).toHaveLength(1);
      expect(results[0].success).toBe(false);
    });
  });

  describe('@TEST:GIT-004-RELEASE - 릴리스 워크플로우', () => {
    it('should create release workflow', async () => {
      const result = await workflow.createRelease('1.0.0', 'Initial release');

      expect(result.success).toBe(true);
      expect(result.branchName).toBe('release/1.0.0');
      expect(result.commitHash).toBe('abc123');
      expect(mockGitManager.createBranch).toHaveBeenCalledWith(
        'release/1.0.0',
        'develop'
      );
    });

    it('should create release PR in team mode', async () => {
      workflow = new WorkflowAutomation(mockGitManager, {
        ...config,
        mode: 'team',
      });

      const result = await workflow.createRelease('1.0.0', 'Team release');

      expect(result.success).toBe(true);
      expect(result.pullRequestUrl).toBe('https://github.com/pr/1');
    });
  });
});
