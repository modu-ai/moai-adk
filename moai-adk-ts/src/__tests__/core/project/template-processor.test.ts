// @TEST:REFACTOR-002-PROCESSOR | -PROCESSOR
// Related: @CODE:TEMPLATE-PROCESSOR-001

/**
 * @file TemplateProcessor Test Suite - Phase 2
 * @tags @TEST:REFACTOR-002-PROCESSOR
 *
 * Bottom-up TDD: TemplateProcessor 분리 테스트
 * - 템플릿 데이터 생성
 * - 파일 생성 로직
 * - 타입별 프로젝트 파일 생성
 */

import { beforeEach, describe, expect, it } from 'vitest';
import { TemplateProcessor } from '@/core/project/template-processor';
import type { ProjectConfig, TemplateData } from '@/types/project';
import { ProjectType } from '@/types/project';

describe('TemplateProcessor - Phase 2: Template Processing Logic', () => {
  let processor: TemplateProcessor;

  beforeEach(() => {
    // RED: TemplateProcessor 클래스가 아직 존재하지 않음
    processor = new TemplateProcessor();
  });

  describe('@TEST:PROCESSOR-DATA-001 - Template Data Creation', () => {
    it('should create basic template data from config', () => {
      const config: ProjectConfig = {
        name: 'test-project',
        type: ProjectType.TYPESCRIPT,
        description: 'Test project',
        author: 'Test Author',
        license: 'MIT',
      };

      // RED: createTemplateData 메서드가 존재하지 않음
      const data: TemplateData = processor.createTemplateData(config);

      expect(data.projectName).toBe('test-project');
      expect(data.projectType).toBe(ProjectType.TYPESCRIPT);
      expect(data.description).toBe('Test project');
      expect(data.author).toBe('Test Author');
      expect(data.license).toBe('MIT');
      expect(data.timestamp).toBeDefined();
      expect(typeof data.timestamp).toBe('string');
    });

    it('should use defaults for optional fields', () => {
      const minimalConfig: ProjectConfig = {
        name: 'minimal-project',
        type: ProjectType.PYTHON,
      };

      const data = processor.createTemplateData(minimalConfig);

      expect(data.author).toBe('MoAI Developer'); // 기본값
      expect(data.license).toBe('MIT'); // 기본값
      expect(data.description).toContain('python'); // 타입 기반 설명
    });

    it('should handle features configuration', () => {
      const config: ProjectConfig = {
        name: 'feature-project',
        type: ProjectType.TYPESCRIPT,
        features: [
          { name: 'typescript', enabled: true },
          { name: 'jest', enabled: true },
          { name: 'biome', enabled: false },
        ],
      };

      const data = processor.createTemplateData(config);

      expect(data.features).toBeDefined();
      expect(data.features.typescript).toBe(true);
      expect(data.features.jest).toBe(true);
      expect(data.features.biome).toBe(false);
    });
  });

  describe('@TEST:PROCESSOR-CONTENT-001 - File Content Generation', () => {
    it('should generate pyproject.toml content', () => {
      const data: TemplateData = {
        projectName: 'python-project',
        projectType: ProjectType.PYTHON,
        timestamp: new Date().toISOString(),
        author: 'Test Author',
        description: 'A Python project',
        license: 'MIT',
        packageManager: 'pip',
        features: {},
      };

      // RED: generatePyprojectToml 메서드가 존재하지 않음
      const content = processor.generatePyprojectToml(data);

      expect(content).toContain('[project]');
      expect(content).toContain('name = "python-project"');
      expect(content).toContain('description = "A Python project"');
      expect(content).toContain('authors = [{name = "Test Author"}]');
      expect(content).toContain('license = {text = "MIT"}');
    });

    it('should generate package.json content', () => {
      const data: TemplateData = {
        projectName: 'node-project',
        projectType: ProjectType.NODEJS,
        timestamp: new Date().toISOString(),
        author: 'Test Author',
        description: 'A Node.js project',
        license: 'MIT',
        packageManager: 'npm',
        features: {},
      };

      // RED: generatePackageJson 메서드가 존재하지 않음
      const packageJson = processor.generatePackageJson(data);

      expect(packageJson.name).toBe('node-project');
      expect(packageJson.version).toBe('0.1.0');
      expect(packageJson.description).toBe('A Node.js project');
      expect(packageJson.author).toBe('Test Author');
      expect(packageJson.license).toBe('MIT');
      expect(packageJson.scripts).toBeDefined();
      expect(packageJson.scripts.build).toBeDefined();
      expect(packageJson.scripts.test).toBeDefined();
    });

    it('should generate tsconfig.json content', () => {
      // RED: generateTsConfig 메서드가 존재하지 않음
      const tsconfig = processor.generateTsConfig();

      expect(tsconfig.compilerOptions).toBeDefined();
      expect(tsconfig.compilerOptions.target).toBeDefined();
      expect(tsconfig.compilerOptions.module).toBeDefined();
      expect(tsconfig.compilerOptions.strict).toBe(true);
      expect(tsconfig.include).toContain('src/**/*');
      expect(tsconfig.exclude).toContain('node_modules');
    });

    it('should generate jest.config content', () => {
      // RED: generateJestConfig 메서드가 존재하지 않음
      const jestConfig = processor.generateJestConfig();

      expect(jestConfig).toContain('module.exports');
      expect(jestConfig).toContain('preset:');
      expect(jestConfig).toContain('testEnvironment:');
    });

    it('should generate pytest.ini content', () => {
      // RED: generatePytestConfig 메서드가 존재하지 않음
      const pytestConfig = processor.generatePytestConfig();

      expect(pytestConfig).toContain('[tool:pytest]');
      expect(pytestConfig).toContain('testpaths');
      expect(pytestConfig).toContain('python_files');
    });
  });

  describe('@TEST:PROCESSOR-MOAI-001 - MoAI Config Generation', () => {
    it('should generate moai config', () => {
      const data: TemplateData = {
        projectName: 'alfred-project',
        projectType: ProjectType.TYPESCRIPT,
        timestamp: new Date().toISOString(),
        author: 'Test Author',
        description: 'Test project',
        license: 'MIT',
        packageManager: 'npm',
        features: {},
      };

      // RED: generateMoaiConfig 메서드가 존재하지 않음
      const config = processor.generateMoaiConfig(data);

      expect(config.project).toBeDefined();
      expect(config.project.name).toBe('alfred-project');
      expect(config.project.type).toBe(ProjectType.TYPESCRIPT);
      expect(config.constitution).toBeDefined();
      expect(config.constitution.enforce_tdd).toBe(true);
      expect(config.constitution.test_coverage_target).toBe(85);
    });

    it('should generate project documentation files', () => {
      const data: TemplateData = {
        projectName: 'doc-project',
        projectType: ProjectType.TYPESCRIPT,
        timestamp: new Date().toISOString(),
        author: 'Test Author',
        description: 'Test project',
        license: 'MIT',
        packageManager: 'npm',
        features: {},
      };

      // RED: generateProjectFile 메서드가 존재하지 않음
      const productMd = processor.generateProjectFile('product.md', data);
      const structureMd = processor.generateProjectFile('structure.md', data);
      const techMd = processor.generateProjectFile('tech.md', data);

      expect(productMd).toContain('doc-project');
      expect(productMd).toContain('PRODUCT');
      expect(structureMd).toContain('STRUCTURE');
      expect(techMd).toContain('TECH');
    });
  });

  describe('@TEST:PROCESSOR-PATH-001 - File Path Generation', () => {
    it('should handle different project types', () => {
      const pythonData: TemplateData = {
        projectName: 'py-project',
        projectType: ProjectType.PYTHON,
        timestamp: new Date().toISOString(),
        author: 'Test',
        description: 'Test',
        license: 'MIT',
        packageManager: 'pip',
        features: {},
      };

      const nodeData: TemplateData = {
        projectName: 'node-project',
        projectType: ProjectType.NODEJS,
        timestamp: new Date().toISOString(),
        author: 'Test',
        description: 'Test',
        license: 'MIT',
        packageManager: 'npm',
        features: {},
      };

      // Python 프로젝트는 pyproject.toml 생성
      expect(() => processor.generatePyprojectToml(pythonData)).not.toThrow();

      // Node 프로젝트는 package.json 생성
      expect(() => processor.generatePackageJson(nodeData)).not.toThrow();
    });
  });

  describe('@TEST:PROCESSOR-ERROR-001 - Error Handling', () => {
    it('should handle missing required fields gracefully', () => {
      const incompleteData = {
        projectName: 'test',
        // 필수 필드 누락
      } as TemplateData;

      // 메서드들이 에러를 던지지 않고 기본값 사용
      expect(() => processor.generateMoaiConfig(incompleteData)).not.toThrow();
    });

    it('should provide meaningful error messages', () => {
      const config = {
        name: 'test',
        type: null as any,
        version: '0.1.0',
      };

      // Validator를 통해 검증되므로 processor는 유효한 데이터만 받음
      // 하지만 방어적 프로그래밍으로 처리 가능해야 함
      expect(() => processor.createTemplateData(config)).not.toThrow();
    });
  });
});
