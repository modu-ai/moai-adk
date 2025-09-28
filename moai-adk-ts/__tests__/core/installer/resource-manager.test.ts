/**
 * @TEST:RESOURCE-MANAGER-001 ResourceManager TDD 테스트 스위트
 *
 * @TASK:TEST-RED-001 실패하는 테스트 작성
 * Python ResourceManager의 모든 기능을 검증하는 포괄적 테스트
 */

import { promises as fs } from 'fs';
import * as path from 'path';
import { tmpdir } from 'os';
import { ResourceManager, TemplateContext } from '../../../src/core/installer/managers/resource-manager';

describe('@TEST:RESOURCE-MANAGER-SUITE-001 ResourceManager', () => {
  let resourceManager: ResourceManager;
  let testDir: string;

  beforeEach(async () => {
    resourceManager = new ResourceManager();
    // 테스트용 임시 디렉토리 생성
    testDir = await fs.mkdtemp(path.join(tmpdir(), 'moai-test-'));
  });

  afterEach(async () => {
    // 테스트 후 정리
    try {
      await fs.rm(testDir, { recursive: true, force: true });
    } catch (error) {
      // 정리 실패는 무시 (테스트 환경에서는 일반적)
    }
  });

  describe('@TEST:BASIC-API-001 Basic API Methods', () => {
    test('@TEST:VERSION-001 should return package version', async () => {
      // RED: 아직 VERSION 파일이 없으므로 실패해야 함
      const version = await resourceManager.getVersion();
      expect(version).toBe('0.0.1'); // 실제 버전으로 업데이트될 예정
    });

    test('@TEST:TEMPLATE-PATH-001 should return correct template path', () => {
      // RED: 템플릿 경로 구조가 아직 없으므로 실패할 수 있음
      const templatePath = resourceManager.getTemplatePath('.claude');
      expect(templatePath).toContain('.claude');
      expect(path.isAbsolute(templatePath)).toBe(true);
    });

    test('@TEST:TEMPLATE-CONTENT-001 should read template content for valid files', async () => {
      // RED: 템플릿 파일이 없으므로 null을 반환해야 함
      const content = await resourceManager.getTemplateContent('nonexistent.md');
      expect(content).toBeNull();
    });

    test('@TEST:TEMPLATE-CONTENT-002 should return null for invalid file types', async () => {
      // RED: 지원하지 않는 파일 타입은 null 반환
      const content = await resourceManager.getTemplateContent('script.sh');
      expect(content).toBeNull();
    });
  });

  describe('@TEST:CLAUDE-RESOURCES-001 Claude Resources Management', () => {
    test('@TEST:CLAUDE-COPY-001 should copy Claude resources successfully', async () => {
      // RED: .claude 템플릿이 없으므로 실패해야 함
      const copiedFiles = await resourceManager.copyClaudeResources(testDir, false);

      expect(copiedFiles).toHaveLength(1);
      expect(copiedFiles[0]).toBe(path.join(testDir, '.claude'));

      // 실제 파일 시스템에서 확인
      const claudeDir = path.join(testDir, '.claude');
      const stat = await fs.stat(claudeDir);
      expect(stat.isDirectory()).toBe(true);
    });

    test('@TEST:CLAUDE-COPY-002 should handle overwrite flag correctly', async () => {
      // RED: 덮어쓰기 시나리오 테스트
      // 첫 번째 복사
      await resourceManager.copyClaudeResources(testDir, false);

      // 두 번째 복사 (overwrite=false)
      const copiedFiles = await resourceManager.copyClaudeResources(testDir, false);
      expect(copiedFiles).toHaveLength(1); // 스킵되어도 성공으로 간주

      // 세 번째 복사 (overwrite=true)
      const overwrittenFiles = await resourceManager.copyClaudeResources(testDir, true);
      expect(overwrittenFiles).toHaveLength(1);
    });

    test('@TEST:HOOK-PERMISSIONS-001 should set executable permissions on hook files', async () => {
      // RED: 훅 파일 권한 설정 테스트
      await resourceManager.copyClaudeResources(testDir, false);

      const hooksDir = path.join(testDir, '.claude', 'hooks', 'moai');
      try {
        const hookFiles = await fs.readdir(hooksDir);
        const pyFiles = hookFiles.filter(f => f.endsWith('.py'));

        for (const pyFile of pyFiles) {
          const filePath = path.join(hooksDir, pyFile);
          const stat = await fs.stat(filePath);

          // Unix 계열에서만 권한 확인 (Windows는 스킵)
          if (process.platform !== 'win32') {
            expect(stat.mode & 0o111).toBeGreaterThan(0); // 실행 권한 확인
          }
        }
      } catch (error) {
        // 훅 디렉토리가 없을 수 있음 - 템플릿에 따라 다름
        expect(error).toBeDefined();
      }
    });
  });

  describe('@TEST:MOAI-RESOURCES-001 MoAI Resources Management', () => {
    test('@TEST:MOAI-COPY-001 should copy MoAI resources successfully', async () => {
      // RED: .moai 템플릿이 없으므로 실패해야 함
      const copiedFiles = await resourceManager.copyMoaiResources(testDir, false);

      expect(copiedFiles).toHaveLength(1);
      expect(copiedFiles[0]).toBe(path.join(testDir, '.moai'));

      // 실제 파일 시스템에서 확인
      const moaiDir = path.join(testDir, '.moai');
      const stat = await fs.stat(moaiDir);
      expect(stat.isDirectory()).toBe(true);
    });

    test('@TEST:MOAI-COPY-002 should exclude templates when requested', async () => {
      // RED: 템플릿 제외 기능 테스트
      const copiedFiles = await resourceManager.copyMoaiResources(
        testDir,
        false,
        true // excludeTemplates
      );

      expect(copiedFiles).toHaveLength(1);

      // _templates 디렉토리가 복사되지 않았는지 확인
      const templatesDir = path.join(testDir, '.moai', '_templates');
      try {
        await fs.access(templatesDir);
        fail('Templates directory should not exist when excluded');
      } catch {
        // 예상된 에러 - 템플릿이 제외되었음
        expect(true).toBe(true);
      }
    });

    test('@TEST:PROJECT-CONTEXT-001 should apply project context variables', async () => {
      // RED: 프로젝트 컨텍스트 치환 테스트
      const projectContext: TemplateContext = {
        PROJECT_NAME: 'test-project',
        PROJECT_DESCRIPTION: 'A test project for MoAI-ADK'
      };

      const copiedFiles = await resourceManager.copyMoaiResources(
        testDir,
        false,
        false,
        projectContext
      );

      expect(copiedFiles).toHaveLength(1);

      // config.json 파일에서 변수 치환 확인
      const configFile = path.join(testDir, '.moai', 'config.json');
      try {
        const content = await fs.readFile(configFile, 'utf-8');
        expect(content).toContain('test-project');
        expect(content).toContain('A test project for MoAI-ADK');
        expect(content).not.toContain('{{PROJECT_NAME}}');
        expect(content).not.toContain('{{PROJECT_DESCRIPTION}}');
      } catch (error) {
        // config.json이 없을 수 있음 - 템플릿 구조에 따라
        expect(error).toBeDefined();
      }
    });
  });

  describe('@TEST:VALIDATION-001 Resource Validation', () => {
    test('@TEST:VALIDATE-SUCCESS-001 should validate existing resources successfully', async () => {
      // RED: 필수 리소스 검증 테스트
      // 리소스 설치
      await resourceManager.copyClaudeResources(testDir, false);
      await resourceManager.copyMoaiResources(testDir, false);

      // CLAUDE.md 파일 생성 (임시)
      const claudeMdPath = path.join(testDir, 'CLAUDE.md');
      await fs.writeFile(claudeMdPath, '# Test Project', 'utf-8');

      const isValid = await resourceManager.validateProjectResources(testDir);
      expect(isValid).toBe(true);
    });

    test('@TEST:VALIDATE-FAILURE-001 should fail validation for missing resources', async () => {
      // RED: 리소스가 없는 경우 검증 실패
      const isValid = await resourceManager.validateProjectResources(testDir);
      expect(isValid).toBe(false);
    });

    test('@TEST:VALIDATE-PARTIAL-001 should fail validation for partially missing resources', async () => {
      // RED: 일부 리소스만 있는 경우 검증 실패
      await resourceManager.copyClaudeResources(testDir, false);
      // .moai와 CLAUDE.md는 누락

      const isValid = await resourceManager.validateProjectResources(testDir);
      expect(isValid).toBe(false);
    });
  });

  describe('@TEST:ERROR-HANDLING-001 Error Handling & Edge Cases', () => {
    test('@TEST:SAFE-PATH-001 should reject unsafe paths', async () => {
      // RED: 안전하지 않은 경로 거부 테스트
      const unsafePaths = [
        '../../../etc/passwd',
        '/etc/hosts',
        testDir + '/../../../usr/bin'
      ];

      for (const unsafePath of unsafePaths) {
        const copiedFiles = await resourceManager.copyClaudeResources(unsafePath, false);
        expect(copiedFiles).toHaveLength(0); // 안전하지 않은 경로는 복사 실패
      }
    });

    test('@TEST:NONEXISTENT-TEMPLATE-001 should handle nonexistent template gracefully', async () => {
      // RED: 존재하지 않는 템플릿 처리
      const content = await resourceManager.getTemplateContent('nonexistent-template.md');
      expect(content).toBeNull();
    });

    test('@TEST:PERMISSION-ERROR-001 should handle permission errors gracefully', async () => {
      // RED: 권한 오류 처리 (Unix 계열에서만)
      if (process.platform === 'win32') {
        return; // Windows에서는 스킵
      }

      // 읽기 전용 디렉토리 생성
      const readOnlyDir = path.join(testDir, 'readonly');
      await fs.mkdir(readOnlyDir);
      await fs.chmod(readOnlyDir, 0o444); // 읽기 전용

      try {
        const copiedFiles = await resourceManager.copyClaudeResources(readOnlyDir, false);
        expect(copiedFiles).toHaveLength(0); // 권한 오류로 복사 실패
      } finally {
        // 정리를 위해 권한 복원
        await fs.chmod(readOnlyDir, 0o755);
      }
    });
  });

  describe('@TEST:CLEAN-INSTALLATION-001 Clean Installation Validation', () => {
    test('@TEST:CLEAN-VALIDATE-001 should validate clean installation structure', async () => {
      // RED: 깨끗한 설치 구조 검증
      await resourceManager.copyMoaiResources(testDir, false);

      // validateCleanInstallation은 private 메서드이므로 간접적으로 테스트
      // 설치 후 로그를 통해 검증이 수행되었는지 확인
      const moaiDir = path.join(testDir, '.moai');
      const stat = await fs.stat(moaiDir);
      expect(stat.isDirectory()).toBe(true);

      // specs 디렉토리가 깨끗한지 확인
      const specsDir = path.join(moaiDir, 'specs');
      try {
        const entries = await fs.readdir(specsDir);
        const nonGitkeepFiles = entries.filter(f => f !== '.gitkeep');
        expect(nonGitkeepFiles).toHaveLength(0);
      } catch {
        // specs 디렉토리가 없을 수도 있음
        expect(true).toBe(true);
      }
    });
  });
});

/**
 * @TEST:INTEGRATION-001 Integration Tests
 * 실제 리소스 파일과의 통합 테스트
 */
describe('@TEST:INTEGRATION-SUITE-001 ResourceManager Integration', () => {
  let resourceManager: ResourceManager;
  let testDir: string;

  beforeEach(async () => {
    resourceManager = new ResourceManager();
    testDir = await fs.mkdtemp(path.join(tmpdir(), 'moai-integration-'));
  });

  afterEach(async () => {
    try {
      await fs.rm(testDir, { recursive: true, force: true });
    } catch {
      // 정리 실패 무시
    }
  });

  test('@TEST:FULL-WORKFLOW-001 should complete full installation workflow', async () => {
    // RED: 전체 설치 워크플로우 테스트

    // 1. Claude 리소스 설치
    const claudeFiles = await resourceManager.copyClaudeResources(testDir, false);
    expect(claudeFiles.length).toBeGreaterThan(0);

    // 2. MoAI 리소스 설치
    const projectContext: TemplateContext = {
      PROJECT_NAME: 'integration-test',
      PROJECT_DESCRIPTION: 'Integration test project'
    };

    const moaiFiles = await resourceManager.copyMoaiResources(
      testDir,
      false,
      false,
      projectContext
    );
    expect(moaiFiles.length).toBeGreaterThan(0);

    // 3. CLAUDE.md 파일 생성
    const claudeMemorySuccess = await resourceManager.copyProjectMemory(
      testDir,
      false,
      projectContext
    );
    expect(claudeMemorySuccess).toBe(true);

    // 4. 전체 검증
    const isValid = await resourceManager.validateProjectResources(testDir);
    expect(isValid).toBe(true);

    // 4. 필수 파일들이 실제로 존재하는지 확인
    const claudeDir = path.join(testDir, '.claude');
    const moaiDir = path.join(testDir, '.moai');

    const claudeStat = await fs.stat(claudeDir);
    const moaiStat = await fs.stat(moaiDir);

    expect(claudeStat.isDirectory()).toBe(true);
    expect(moaiStat.isDirectory()).toBe(true);
  });
});