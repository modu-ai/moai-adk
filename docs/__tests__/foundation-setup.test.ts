/**
 * TEST-FOUNDATION-SETUP-001: Nextra foundation setup tests
 * Nextra 기반 문서 시스템 기능 테스트
 *
 * 기본 설정 검증:
 * 1. Nextra 테마 정상적으로 로드되는지
 * 2. 다국어 설정이 올바르게 구성되어 있는지
 * 3. 모바일/태블릿 반응형 디자인이 작동하는지
 * 4. 검색 기능이 활성화되어 있는지
 * 5. 깃허브 연동이 올바르게 설정되어 있는지
 */

describe('Nextra Foundation Setup Tests', () => {
  describe('Configuration Validation', () => {
    it('should have proper Nextra configuration', () => {
      // 테스트: Nextra 설정이 올바르게 구성되어 있는지
      const fs = require('fs');
      const path = require('path');

      expect(fs.existsSync(path.join(__dirname, '../next.config.js'))).toBe(true);
      expect(fs.existsSync(path.join(__dirname, '../theme.config.tsx'))).toBe(true);
      expect(fs.existsSync(path.join(__dirname, '../package.json'))).toBe(true);
    });

    it('should have correct theme configuration structure', () => {
      // 테스트: 테마 설정이 올바르게 구성되어 있는지
      const fs = require('fs');
      const path = require('path');
      const themeConfigPath = path.join(__dirname, '../theme.config.tsx');

      if (fs.existsSync(themeConfigPath)) {
        const content = fs.readFileSync(themeConfigPath, 'utf8');
        expect(content).toContain('logo:');
        expect(content).toContain('project:');
        expect(content).toContain('i18n:');
        expect(content).toContain('search:');
      }
    });

    it('should support multiple languages', () => {
      // 테스트: 다국어 설정이 올바르게 구성되어 있는지
      const fs = require('fs');
      const path = require('path');
      const themeConfigPath = path.join(__dirname, '../theme.config.tsx');

      if (fs.existsSync(themeConfigPath)) {
        const content = fs.readFileSync(themeConfigPath, 'utf8');
        expect(content).toContain('ko');
        expect(content).toContain('en');
        expect(content).toContain('ja');
        expect(content).toContain('zh');
        expect(content).toContain('한국어');
        expect(content).toContain('English');
        expect(content).toContain('日本語');
        expect(content).toContain('中文');
      }
    });

    it('should have proper search configuration', () => {
      // 테스트: 검색 기능이 올바르게 구성되어 있는지
      const fs = require('fs');
      const path = require('path');
      const themeConfigPath = path.join(__dirname, '../theme.config.tsx');

      if (fs.existsSync(themeConfigPath)) {
        const content = fs.readFileSync(themeConfigPath, 'utf8');
        expect(content).toContain('search:');
        expect(content).toContain('placeholder:');
      }
    });
  });

  describe('Theme Features', () => {
    it('should have responsive design enabled', () => {
      // 테스트: 반응형 디자인이 활성화되어 있는지
      const fs = require('fs');
      const path = require('path');
      const themeConfigPath = path.join(__dirname, '../theme.config.tsx');

      if (fs.existsSync(themeConfigPath)) {
        const content = fs.readFileSync(themeConfigPath, 'utf8');
        expect(content).toContain('toc:');
        expect(content).toContain('float:');
        expect(content).toContain('backToTop:');
      }
    });

    it('should have proper navigation setup', () => {
      // 테스트: 네비게이션이 올바르게 설정되어 있는지
      const fs = require('fs');
      const path = require('path');
      const themeConfigPath = path.join(__dirname, '../theme.config.tsx');

      if (fs.existsSync(themeConfigPath)) {
        const content = fs.readFileSync(themeConfigPath, 'utf8');
        expect(content).toContain('navigation:');
        expect(content).toContain('sidebar:');
        expect(content).toContain('toggleButton:');
      }
    });

    it('should have dark mode enabled', () => {
      // 테스트: 다크 모드가 활성화되어 있는지
      const fs = require('fs');
      const path = require('path');
      const themeConfigPath = path.join(__dirname, '../theme.config.tsx');

      if (fs.existsSync(themeConfigPath)) {
        const content = fs.readFileSync(themeConfigPath, 'utf8');
        expect(content).toContain('darkMode:');
        expect(content).toContain('nextThemes:');
      }
    });
  });

  describe('External Integrations', () => {
    it('should have GitHub repository configured', () => {
      // 테스트: 깃허브 저장소가 올바르게 설정되어 있는지
      const fs = require('fs');
      const path = require('path');
      const themeConfigPath = path.join(__dirname, '../theme.config.tsx');

      if (fs.existsSync(themeConfigPath)) {
        const content = fs.readFileSync(themeConfigPath, 'utf8');
        expect(content).toContain('docsRepositoryBase:');
        expect(content).toContain('project:');
        expect(content).toContain('github.com/modu-ai/moai-adk');
      }
    });

    it('should have feedback system configured', () => {
      // 테스트: 피드백 시스템이 설정되어 있는지
      const fs = require('fs');
      const path = require('path');
      const themeConfigPath = path.join(__dirname, '../theme.config.tsx');

      if (fs.existsSync(themeConfigPath)) {
        const content = fs.readFileSync(themeConfigPath, 'utf8');
        expect(content).toContain('feedback:');
        expect(content).toContain('질문이 있나요?');
      }
    });
  });

  describe('Build Configuration', () => {
    it('should have proper Next.js configuration', () => {
      // 테스트: Next.js 설정이 올바르게 구성되어 있는지
      const fs = require('fs');
      const path = require('path');
      const nextConfigPath = path.join(__dirname, '../next.config.js');

      if (fs.existsSync(nextConfigPath)) {
        const content = fs.readFileSync(nextConfigPath, 'utf8');
        expect(content).toContain('nextConfig');
        expect(content).toContain('nextra');
        expect(content).toContain('experimental:');
        expect(content).toContain('webpackBuildWorker:');
      }
    });

    it('should have proper optimizations enabled', () => {
      // 테스트: 최적화 설정이 올바르게 구성되어 있는지
      const fs = require('fs');
      const path = require('path');
      const nextConfigPath = path.join(__dirname, '../next.config.js');

      if (fs.existsSync(nextConfigPath)) {
        const content = fs.readFileSync(nextConfigPath, 'utf8');
        expect(content).toContain('compiler:');
        expect(content).toContain('compress:');
        expect(content).toContain('poweredByHeader:');
      }
    });
  });

  describe('Package Dependencies', () => {
    it('should have required Nextra dependencies', () => {
      // 테스트: 필요한 Nextra 의존성이 설치되어 있는지
      const fs = require('fs');
      const path = require('path');
      const packageJsonPath = path.join(__dirname, '../package.json');

      if (fs.existsSync(packageJsonPath)) {
        const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
        expect(packageJson.dependencies).toBeDefined();
        expect(packageJson.dependencies.next).toBeDefined();
        expect(packageJson.dependencies.nextra).toBeDefined();
        expect(packageJson.dependencies['nextra-theme-docs']).toBeDefined();
        expect(packageJson.dependencies.react).toBeDefined();
        expect(packageJson.dependencies['react-dom']).toBeDefined();
      }
    });

    it('should have test dependencies', () => {
      // 테스트: 테스트 관련 의존성이 설치되어 있는지
      const fs = require('fs');
      const path = require('path');
      const packageJsonPath = path.join(__dirname, '../package.json');

      if (fs.existsSync(packageJsonPath)) {
        const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
        expect(packageJson.devDependencies).toBeDefined();
        expect(packageJson.devDependencies.jest).toBeDefined();
        expect(packageJson.devDependencies['ts-jest']).toBeDefined();
        expect(packageJson.devDependencies['@types/jest']).toBeDefined();
      }
    });
  });
});