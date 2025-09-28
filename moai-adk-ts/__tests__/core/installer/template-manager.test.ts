/**
 * TemplateManager 테스트 수트
 * TDD RED 단계: 실패하는 테스트 작성
 */

import * as fs from 'fs/promises';
import * as path from 'path';
import * as os from 'os';
import {
  TemplateManager,
  TemplateContext,
  templateManager
} from '../../../src/core/installer/managers/template-manager';

describe('TemplateManager', () => {
  let manager: TemplateManager;
  let tempDir: string;
  let testTemplatePath: string;

  beforeEach(async () => {
    manager = new TemplateManager();

    // 임시 디렉토리 및 테스트 템플릿 생성
    tempDir = await fs.mkdtemp(path.join(os.tmpdir(), 'template-test-'));
    testTemplatePath = path.join(tempDir, 'test-template.mustache');
  });

  afterEach(async () => {
    // 임시 파일 정리
    await fs.rm(tempDir, { recursive: true });
    manager.clearCache();
  });

  describe('기본 템플릿 렌더링', () => {
    it('should render simple variable substitution', () => {
      const template = 'Hello {{name}}!';
      const context: TemplateContext = { name: 'World' };

      const result = manager.renderTemplate(template, context);

      expect(result).toBe('Hello World!');
    });

    it('should render conditional blocks', () => {
      const template = '{{#isActive}}Active User{{/isActive}}{{^isActive}}Inactive User{{/isActive}}';

      const activeContext: TemplateContext = { isActive: true };
      const inactiveContext: TemplateContext = { isActive: false };

      expect(manager.renderTemplate(template, activeContext)).toBe('Active User');
      expect(manager.renderTemplate(template, inactiveContext)).toBe('Inactive User');
    });

    it('should render lists with iteration', () => {
      const template = '{{#features}}Feature: {{.}}\n{{/features}}';
      const context: TemplateContext = {
        features: ['Authentication', 'Database', 'API']
      };

      const result = manager.renderTemplate(template, context);

      expect(result).toBe('Feature: Authentication\nFeature: Database\nFeature: API\n');
    });

    it('should render complex nested objects', () => {
      const template = '{{#project}}Project: {{name}} ({{version}}){{/project}}';
      const context: TemplateContext = {
        project: {
          name: 'MoAI-ADK',
          version: '0.0.1'
        }
      };

      const result = manager.renderTemplate(template, context);

      expect(result).toBe('Project: MoAI-ADK (0.0.1)');
    });
  });

  describe('파일 기반 템플릿 처리', () => {
    it('should load template from file', async () => {
      const templateContent = 'Project: {{PROJECT_NAME}}\nMode: {{MODE}}';
      await fs.writeFile(testTemplatePath, templateContent);

      const result = await manager.loadTemplate(testTemplatePath);

      expect(result).toBe(templateContent);
    });

    it('should render template file with context', async () => {
      const templateContent = 'Hello {{user}}, welcome to {{app}}!';
      await fs.writeFile(testTemplatePath, templateContent);

      const context: TemplateContext = {
        user: 'Developer',
        app: 'MoAI-ADK'
      };

      const result = await manager.renderTemplateFile(testTemplatePath, context);

      expect(result).toBe('Hello Developer, welcome to MoAI-ADK!');
    });

    it('should throw error for non-existent template file', async () => {
      const nonExistentPath = path.join(tempDir, 'nonexistent.mustache');

      await expect(manager.loadTemplate(nonExistentPath))
        .rejects
        .toThrow('Failed to load template');
    });
  });

  describe('템플릿 캐싱', () => {
    it('should cache loaded templates', async () => {
      const templateContent = 'Cached template: {{value}}';
      await fs.writeFile(testTemplatePath, templateContent);

      // 첫 번째 로드
      await manager.loadTemplate(testTemplatePath);
      expect(manager.getCacheSize()).toBe(1);

      // 두 번째 로드 (캐시에서 가져와야 함)
      const cachedTemplate = await manager.loadTemplate(testTemplatePath);
      expect(cachedTemplate).toBe(templateContent);
      expect(manager.getCacheSize()).toBe(1);
    });

    it('should clear cache when requested', async () => {
      const templateContent = 'Template to be cached';
      await fs.writeFile(testTemplatePath, templateContent);

      await manager.loadTemplate(testTemplatePath);
      expect(manager.getCacheSize()).toBe(1);

      manager.clearCache();
      expect(manager.getCacheSize()).toBe(0);
    });

    it('should support cache-disabled mode', async () => {
      const noCacheManager = new TemplateManager(false);
      const templateContent = 'No cache template';
      await fs.writeFile(testTemplatePath, templateContent);

      await noCacheManager.loadTemplate(testTemplatePath);
      expect(noCacheManager.getCacheSize()).toBe(0);
    });

    it('should cache parsed templates for performance', () => {
      const template = 'Hello {{name}}, welcome to {{app}}!';
      const context1: TemplateContext = { name: 'Alice', app: 'MoAI' };
      const context2: TemplateContext = { name: 'Bob', app: 'MoAI' };

      // 첫 번째 렌더링 (파싱 + 캐싱)
      manager.renderTemplate(template, context1);
      expect(manager.getParsedCacheSize()).toBe(1);

      // 두 번째 렌더링 (캐시된 파싱 사용)
      manager.renderTemplate(template, context2);
      expect(manager.getParsedCacheSize()).toBe(1);
    });

    it('should limit parsed template cache size', () => {
      const limitedManager = new TemplateManager(true, 2); // 최대 2개만 캐싱

      const template1 = 'Template 1: {{value}}';
      const template2 = 'Template 2: {{value}}';
      const template3 = 'Template 3: {{value}}';
      const context: TemplateContext = { value: 'test' };

      limitedManager.renderTemplate(template1, context);
      limitedManager.renderTemplate(template2, context);
      expect(limitedManager.getParsedCacheSize()).toBe(2);

      // 3번째 템플릿 추가 시 첫 번째 제거
      limitedManager.renderTemplate(template3, context);
      expect(limitedManager.getParsedCacheSize()).toBe(2);
    });
  });

  describe('템플릿 검증', () => {
    it('should validate correct Mustache template', () => {
      const validTemplate = 'Hello {{name}}! {{#items}}Item: {{.}}{{/items}}';

      expect(manager.validateTemplate(validTemplate)).toBe(true);
    });

    it('should detect invalid template syntax', () => {
      const invalidTemplate = 'Hello {{name! {{#items}Item: {{.}}{{/invalid}}';

      expect(manager.validateTemplate(invalidTemplate)).toBe(false);
    });

    it('should validate empty template', () => {
      expect(manager.validateTemplate('')).toBe(true);
    });

    it('should validate template with comments', () => {
      const templateWithComments = '{{! This is a comment }}Hello {{name}}!';

      expect(manager.validateTemplate(templateWithComments)).toBe(true);
    });
  });

  describe('에러 처리', () => {
    it('should handle rendering errors gracefully', () => {
      const invalidTemplate = '{{#unclosed}}';
      const context: TemplateContext = {};

      expect(() => manager.renderTemplate(invalidTemplate, context))
        .toThrow('Template rendering failed');
    });

    it('should handle complex context with circular references', () => {
      // 순환 참조가 있는 객체는 JSON 직렬화가 불가능하므로 에러 발생 가능
      const context: any = { name: 'test' };
      context.self = context;

      const template = 'Name: {{name}}';

      // Mustache는 순환 참조를 처리할 수 있어야 함 (또는 적절한 에러 처리)
      expect(() => manager.renderTemplate(template, context)).not.toThrow();
    });
  });

  describe('MoAI-ADK 특화 템플릿 예시', () => {
    it('should render MoAI project template', async () => {
      const moaiTemplate = `{{! MoAI-ADK Project Template }}
# {{PROJECT_NAME}}

**Mode**: {{MODE}}
**Version**: {{VERSION}}

{{#FEATURES}}
- Feature: {{.}}
{{/FEATURES}}

{{#HAS_GIT}}
## Git Configuration
Repository initialized with {{BRANCH_PREFIX}} prefix.
{{/HAS_GIT}}

{{^HAS_GIT}}
## No Git
Local project without version control.
{{/HAS_GIT}}`;

      await fs.writeFile(testTemplatePath, moaiTemplate);

      const context: TemplateContext = {
        PROJECT_NAME: 'My MoAI Project',
        MODE: 'development',
        VERSION: '0.1.0',
        FEATURES: ['CLI', 'TDD', 'Agents'],
        HAS_GIT: true,
        BRANCH_PREFIX: 'feature'
      };

      const result = await manager.renderTemplateFile(testTemplatePath, context);

      expect(result).toContain('# My MoAI Project');
      expect(result).toContain('**Mode**: development');
      expect(result).toContain('- Feature: CLI');
      expect(result).toContain('- Feature: TDD');
      expect(result).toContain('- Feature: Agents');
      expect(result).toContain('Repository initialized with feature prefix');
      expect(result).not.toContain('No Git');
    });

    it('should handle missing context gracefully', () => {
      const template = 'Project: {{PROJECT_NAME}}, Optional: {{OPTIONAL_FIELD}}';
      const context: TemplateContext = {
        PROJECT_NAME: 'Test Project'
        // OPTIONAL_FIELD 누락
      };

      const result = manager.renderTemplate(template, context);

      expect(result).toBe('Project: Test Project, Optional: ');
    });
  });

  describe('성능 테스트', () => {
    it('should render large templates efficiently', () => {
      const largeTemplate = Array(1000).fill('Item {{index}}: {{name}}\n').join('');
      const context: TemplateContext = {
        index: 1,
        name: 'Performance Test'
      };

      const startTime = performance.now();
      const result = manager.renderTemplate(largeTemplate, context);
      const endTime = performance.now();

      expect(result).toContain('Item 1: Performance Test');
      expect(endTime - startTime).toBeLessThan(100); // 100ms 이내
    });

    it('should maintain parsed template cache correctly', () => {
      const template = 'Cached template: {{value}} {{index}}';
      const cachedManager = new TemplateManager(true);

      // 첫 번째 렌더링에서 캐시 생성
      cachedManager.renderTemplate(template, { value: 'test', index: 1 });
      expect(cachedManager.getParsedCacheSize()).toBe(1);

      // 다른 컨텍스트로 동일 템플릿 렌더링 (캐시 재사용)
      const result = cachedManager.renderTemplate(template, { value: 'cached', index: 2 });
      expect(result).toBe('Cached template: cached 2');
      expect(cachedManager.getParsedCacheSize()).toBe(1);
    });
  });

  describe('싱글톤 인스턴스', () => {
    it('should provide singleton template manager', () => {
      expect(templateManager).toBeInstanceOf(TemplateManager);
      expect(templateManager).toBe(templateManager); // 동일한 인스턴스
    });

    it('should use singleton for common operations', () => {
      const template = 'Singleton test: {{value}}';
      const context: TemplateContext = { value: 'works' };

      const result = templateManager.renderTemplate(template, context);

      expect(result).toBe('Singleton test: works');
    });
  });
});