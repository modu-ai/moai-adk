/**
 * @TEST:TEMPLATE-UTILS-001 Template Utils Complete Test Suite
 *
 * Python template_manager.py와 template_processor.py 기능들의 TypeScript 포팅 검증
 * @FEATURE:TEMPLATE-UTILS-001 완전한 템플릿 유틸리티 기능 테스트
 */

import {
  unifiedSubstituteTemplateVariables,
  processMultipleVariableFormats,
  expandNestedVariables,
  mergeTemplateContexts,
  applyProjectContext,
  shouldProcessAsTemplate,
  fileExists,
} from '../template-utils';
import type { TemplateContext } from '../template-processor';
import * as fs from 'fs/promises';
import * as path from 'path';
import * as os from 'os';

describe('Template Utils - Python Porting Tests', () => {
  let tempDir: string;
  let testContext: TemplateContext;

  beforeEach(async () => {
    // 임시 디렉토리 생성
    tempDir = await fs.mkdtemp(path.join(os.tmpdir(), 'moai-template-test-'));

    // 테스트 컨텍스트 설정
    testContext = {
      PROJECT_NAME: 'test-project',
      PROJECT_TYPE: 'python',
      AUTHOR: 'Test Developer',
      VERSION: '1.0.0',
      ENVIRONMENT: 'development',
      PROJECT_development_CONFIG: 'dev-config.json',
      PROJECT_production_CONFIG: 'prod-config.json',
    };
  });

  afterEach(async () => {
    // 임시 디렉토리 정리
    await fs.rm(tempDir, { recursive: true, force: true });
  });

  describe('@TEST:UNIFIED-SUBSTITUTE-001 통합 변수 치환 시스템', () => {
    it('should handle all variable formats like Python implementation', () => {
      const template = `
        Project: [PROJECT_NAME]
        Type: {{PROJECT_TYPE}}
        Author: \${AUTHOR}
        Version: $VERSION
        Unknown: [UNKNOWN_VAR]
        Mixed: {{PROJECT_TYPE}} by [AUTHOR] v$VERSION
      `;

      const result = unifiedSubstituteTemplateVariables(template, testContext);

      expect(result).toContain('Project: test-project');
      expect(result).toContain('Type: python');
      expect(result).toContain('Author: Test Developer');
      expect(result).toContain('Version: 1.0.0');
      expect(result).toContain('Unknown: [UNKNOWN_VAR]'); // 알려지지 않은 변수는 원본 유지
      expect(result).toContain('Mixed: python by Test Developer v1.0.0');
    });

    it('should preserve original content on error', () => {
      // 잘못된 컨텍스트로 오류 유발
      const template = 'Test content';
      const invalidContext = null as any;

      const result = unifiedSubstituteTemplateVariables(
        template,
        invalidContext
      );

      expect(result).toBe(template); // 원본 내용 보존
    });

    it('should handle edge cases like Python', () => {
      const template = `
        $PROJECT_NAME_extended should not match
        $PROJECT_NAME should match
        \${PROJECT_NAME} should match
        {{PROJECT_NAME}} should match
        [PROJECT_NAME] should match
      `;

      const result = unifiedSubstituteTemplateVariables(template, testContext);

      expect(result).toContain('$PROJECT_NAME_extended should not match'); // 긴 변수명은 매치 안됨
      expect(result).toContain('test-project should match'); // $PROJECT_NAME 매치
      expect(result).toContain('test-project should match'); // ${PROJECT_NAME} 매치
      expect(result).toContain('test-project should match'); // {{PROJECT_NAME}} 매치
      expect(result).toContain('test-project should match'); // [PROJECT_NAME] 매치
    });
  });

  describe('@TEST:MULTIPLE-FORMATS-001 다중 변수 포맷 처리', () => {
    it('should process all supported variable formats', () => {
      const template = 'Project [PROJECT_NAME] version ${VERSION} by $AUTHOR';
      const result = processMultipleVariableFormats(template, testContext);

      expect(result).toBe(
        'Project test-project version 1.0.0 by Test Developer'
      );
    });

    it('should preserve unknown variables', () => {
      const template = 'Known: [PROJECT_NAME], Unknown: [UNKNOWN_VAR]';
      const result = processMultipleVariableFormats(template, testContext);

      expect(result).toBe('Known: test-project, Unknown: [UNKNOWN_VAR]');
    });
  });

  describe('@TEST:NESTED-VARIABLES-001 중첩 변수 확장', () => {
    it('should expand nested variables correctly', () => {
      const template = 'Config: {{PROJECT_{{ENVIRONMENT}}_CONFIG}}';
      const result = expandNestedVariables(template, testContext);

      expect(result).toBe('Config: dev-config.json');
    });

    it('should handle complex nested patterns', () => {
      const template =
        'Complex: {{PROJECT_{{ENVIRONMENT}}_CONFIG}} for {{PROJECT_NAME}}';
      const result = expandNestedVariables(template, testContext);

      expect(result).toBe('Complex: dev-config.json for {{PROJECT_NAME}}');
    });

    it('should prevent infinite loops', () => {
      const circularContext = {
        A: 'B',
        B: '{{A}}',
      };

      const template = '{{{{A}}}}';
      const result = expandNestedVariables(template, circularContext);

      // 무한 루프 방지 확인 (최대 5회 반복 후 중단)
      expect(result).toBeTruthy();
    });
  });

  describe('@TEST:CONTEXT-MERGING-001 컨텍스트 병합', () => {
    it('should merge contexts with correct priority', () => {
      const primary = { A: 'primary', B: 'primary' };
      const secondary1 = { B: 'secondary1', C: 'secondary1' };
      const secondary2 = { C: 'secondary2', D: 'secondary2' };

      const result = mergeTemplateContexts(primary, secondary1, secondary2);

      expect(result['A']).toBe('primary'); // primary만 존재
      expect(result['B']).toBe('primary'); // primary 우선
      expect(result['C']).toBe('secondary2'); // 나중 secondary 우선
      expect(result['D']).toBe('secondary2'); // secondary2만 존재
    });

    it('should return immutable result', () => {
      const primary = { A: 'test' };
      const result = mergeTemplateContexts(primary);

      expect(Object.isFrozen(result)).toBe(true);
    });
  });

  describe('@TEST:APPLY-PROJECT-CONTEXT-001 프로젝트 컨텍스트 적용', () => {
    it('should apply context to file and modify content', async () => {
      const templatePath = path.join(tempDir, 'test-template.txt');
      const templateContent = 'Project: [PROJECT_NAME]\nAuthor: {{AUTHOR}}';

      await fs.writeFile(templatePath, templateContent);

      const result = await applyProjectContext(templatePath, testContext);

      expect(result).toBe(true);

      const modifiedContent = await fs.readFile(templatePath, 'utf-8');
      expect(modifiedContent).toBe(
        'Project: test-project\nAuthor: Test Developer'
      );
    });

    it('should return false for non-existent file', async () => {
      const nonExistentPath = path.join(tempDir, 'non-existent.txt');

      const result = await applyProjectContext(nonExistentPath, testContext);

      expect(result).toBe(false);
    });

    it('should not modify file if no changes needed', async () => {
      const templatePath = path.join(tempDir, 'no-change.txt');
      const templateContent = 'No variables here';

      await fs.writeFile(templatePath, templateContent);

      // 컨텍스트 적용
      const result = await applyProjectContext(templatePath, testContext);

      expect(result).toBe(true);

      const finalContent = await fs.readFile(templatePath, 'utf-8');
      expect(finalContent).toBe(templateContent); // 내용 변경 없음
    });
  });

  describe('@TEST:SHOULD-PROCESS-TEMPLATE-001 템플릿 처리 여부 판단', () => {
    it('should correctly identify text files for template processing', () => {
      const textFiles = [
        'config.json',
        'README.md',
        'script.py',
        'component.ts',
        'styles.css',
        'Dockerfile', // 확장자 없는 텍스트 파일
        '.gitignore',
      ];

      textFiles.forEach(file => {
        expect(shouldProcessAsTemplate(file)).toBe(true);
      });
    });

    it('should correctly identify binary files to skip', () => {
      const binaryFiles = [
        'image.png',
        'video.mp4',
        'archive.zip',
        'executable.exe',
        'library.dll',
        'database.sqlite',
      ];

      binaryFiles.forEach(file => {
        expect(shouldProcessAsTemplate(file)).toBe(false);
      });
    });
  });

  describe('@TEST:FILE-EXISTS-001 파일 존재 확인', () => {
    it('should return true for existing file', async () => {
      const testFile = path.join(tempDir, 'exists.txt');
      await fs.writeFile(testFile, 'test content');

      const exists = await fileExists(testFile);

      expect(exists).toBe(true);
    });

    it('should return false for non-existent file', async () => {
      const nonExistentFile = path.join(tempDir, 'does-not-exist.txt');

      const exists = await fileExists(nonExistentFile);

      expect(exists).toBe(false);
    });
  });

  describe('@TEST:INTEGRATION-001 통합 워크플로우 테스트', () => {
    it('should handle complete template processing workflow', async () => {
      // 복잡한 템플릿 생성
      const complexTemplate = `
# Project Configuration
project_name: [PROJECT_NAME]
project_type: {{PROJECT_TYPE}}
author: \${AUTHOR}
version: $VERSION

# Environment-specific config
config_file: {{PROJECT_{{ENVIRONMENT}}_CONFIG}}

# Mixed formats
description: "Project {{PROJECT_TYPE}} by [AUTHOR] version $VERSION"
      `;

      const templatePath = path.join(tempDir, 'complex-template.yml');
      await fs.writeFile(templatePath, complexTemplate);

      // 1단계: 중첩 변수 확장
      let processed = await fs.readFile(templatePath, 'utf-8');
      processed = expandNestedVariables(processed, testContext);

      // 2단계: 통합 변수 치환
      processed = unifiedSubstituteTemplateVariables(processed, testContext);

      // 3단계: 결과 파일에 쓰기
      const resultPath = path.join(tempDir, 'processed-template.yml');
      await fs.writeFile(resultPath, processed);

      // 검증
      const finalContent = await fs.readFile(resultPath, 'utf-8');

      expect(finalContent).toContain('project_name: test-project');
      expect(finalContent).toContain('project_type: python');
      expect(finalContent).toContain('author: Test Developer');
      expect(finalContent).toContain('version: 1.0.0');
      expect(finalContent).toContain('config_file: dev-config.json');
      expect(finalContent).toContain(
        'description: "Project python by Test Developer version 1.0.0"'
      );
    });
  });
});
