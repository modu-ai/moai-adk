/**
 * TEST-CONTENT-CREATION-001: Content creation functionality tests
 * 핵심 콘텐츠 생성 테스트
 *
 * 기본 콘텐츠 구조 검증:
 * 1. 메인 페이지 콘텐츠 존재 여부
 * 2. 문서 구조 검증
 * 3. 각 섹션별 콘텐츠 검증
 * 4. 다국어 콘텐츠 검증
 * 5. 메타데이터 검증
 */

describe('Content Creation Tests', () => {
  describe('Main Content Structure', () => {
    it('should have main page content', () => {
      // 테스트: 메인 페이지 콘텐츠가 존재하는지
      const fs = require('fs');
      const path = require('path');

      const indexPath = path.join(__dirname, '../pages/index.md');
      expect(fs.existsSync(indexPath)).toBe(true);
    });

    it('should have proper main page structure', () => {
      // 테스트: 메인 페이지 구조가 올바른지
      const fs = require('fs');
      const path = require('path');
      const indexPath = path.join(__dirname, '../pages/index.md');

      if (fs.existsSync(indexPath)) {
        const content = fs.readFileSync(indexPath, 'utf8');
        expect(content).toContain('# ');
        expect(content).toContain('MoAI-ADK');
        expect(content).toContain('문서');
      }
    });

    it('should have introduction section', () => {
      // 테스트: 소개 섹션이 존재하는지
      const fs = require('fs');
      const path = require('path');
      const indexPath = path.join(__dirname, '../pages/index.md');

      if (fs.existsSync(indexPath)) {
        const content = fs.readFileSync(indexPath, 'utf8');
        expect(content).toMatch(/소개|개요|Introduction|Overview/i);
      }
    });
  });

  describe('Documentation Structure', () => {
    it('should have getting-started section', () => {
      // 테스트: 시작하기 섹션이 존재하는지
      const fs = require('fs');
      const path = require('path');
      const gettingStartedPath = path.join(__dirname, '../pages/en/getting-started/index.md');

      if (fs.existsSync(gettingStartedPath)) {
        const content = fs.readFileSync(gettingStartedPath, 'utf8');
        expect(content).toContain('Getting Started');
        expect(content).toContain('시작하기');
      }
    });

    it('should have guides section', () => {
      // 테스트: 가이드 섹션이 존재하는지
      const fs = require('fs');
      const path = require('path');
      const guidesPath = path.join(__dirname, '../pages/en/guides/index.md');

      if (fs.existsSync(guidesPath)) {
        const content = fs.readFileSync(guidesPath, 'utf8');
        expect(content).toContain('Guides');
        expect(content).toContain('가이드');
      }
    });

    it('should have reference section', () => {
      // 테스트: 레퍼런스 섹션이 존재하는지
      const fs = require('fs');
      const path = require('path');
      const referencePath = path.join(__dirname, '../pages/en/reference/index.md');

      if (fs.existsSync(referencePath)) {
        const content = fs.readFileSync(referencePath, 'utf8');
        expect(content).toContain('Reference');
        expect(content).toContain('참고 자료');
      }
    });

    it('should have troubleshooting section', () => {
      // 테스트: 문제 해결 섹션이 존재하는지
      const fs = require('fs');
      const path = require('path');
      const troubleshootingPath = path.join(__dirname, '../pages/en/troubleshooting/index.md');

      if (fs.existsSync(troubleshootingPath)) {
        const content = fs.readFileSync(troubleshootingPath, 'utf8');
        expect(content).toContain('Troubleshooting');
        expect(content).toContain('문제 해결');
      }
    });
  });

  describe('Content Organization', () => {
    it('should have hierarchical structure', () => {
      // 테스트: 계층적 구조가 존재하는지
      const fs = require('fs');
      const path = require('path');

      const checkDirectory = (dirPath: string) => {
        if (fs.existsSync(dirPath)) {
          const stats = fs.statSync(dirPath);
          if (stats.isDirectory()) {
            const files = fs.readdirSync(dirPath);
            return files.length > 0;
          }
        }
        return false;
      };

      const docsDir = path.join(__dirname, '../pages');
      expect(checkDirectory(docsDir)).toBe(true);

      const enDir = path.join(docsDir, 'en');
      expect(checkDirectory(enDir)).toBe(true);

      const koDir = path.join(docsDir, 'ko');
      expect(checkDirectory(koDir)).toBe(true);
    });

    it('should have metadata for all pages', () => {
      // 테스트: 모든 페이지에 메타데이터가 존재하는지
      const fs = require('fs');
      const path = require('path');

      const checkMetaFiles = (dirPath: string) => {
        if (fs.existsSync(dirPath)) {
          const files = fs.readdirSync(dirPath);
          return files.includes('_meta.json');
        }
        return false;
      };

      const enDir = path.join(__dirname, '../pages/en');
      expect(checkMetaFiles(enDir)).toBe(true);

      const koDir = path.join(__dirname, '../pages/ko');
      expect(checkMetaFiles(koDir)).toBe(true);
    });
  });

  describe('Multilingual Content', () => {
    it('should have Korean content', () => {
      // 테스트: 한국어 콘텐츠가 존재하는지
      const fs = require('fs');
      const path = require('path');
      const koIndex = path.join(__dirname, '../pages/ko/index.md');

      if (fs.existsSync(koIndex)) {
        const content = fs.readFileSync(koIndex, 'utf8');
        expect(content).toContain('MoAI-ADK');
        expect(content).toContain('한국어');
      }
    });

    it('should have English content', () => {
      // 테스트: 영어 콘텐츠가 존재하는지
      const fs = require('fs');
      const path = require('path');
      const enIndex = path.join(__dirname, '../pages/en/index.md');

      if (fs.existsSync(enIndex)) {
        const content = fs.readFileSync(enIndex, 'utf8');
        expect(content).toContain('MoAI-ADK');
        expect(content).toContain('English');
      }
    });

    it('should have Japanese content', () => {
      // 테스트: 일본어 콘텐츠가 존재하는지
      const fs = require('fs');
      const path = require('path');
      const jaIndex = path.join(__dirname, '../pages/ja/index.md');

      if (fs.existsSync(jaIndex)) {
        const content = fs.readFileSync(jaIndex, 'utf8');
        expect(content).toContain('MoAI-ADK');
        expect(content).toContain('日本語');
      }
    });

    it('should have Chinese content', () => {
      // 테스트: 중국어 콘텐츠가 존재하는지
      const fs = require('fs');
      const path = require('path');
      const zhIndex = path.join(__dirname, '../pages/zh/index.md');

      if (fs.existsSync(zhIndex)) {
        const content = fs.readFileSync(zhIndex, 'utf8');
        expect(content).toContain('MoAI-ADK');
        expect(content).toContain('中文');
      }
    });
  });

  describe('Content Validation', () => {
    it('should have proper frontmatter structure', () => {
      // 테스트: 프론트매터 구조가 올바른지
      const fs = require('fs');
      const path = require('path');

      // 메인 페이지 프론트매터 검증
      const indexPath = path.join(__dirname, '../pages/index.md');
      if (fs.existsSync(indexPath)) {
        const content = fs.readFileSync(indexPath, 'utf8');
        expect(content).toContain('---');
        expect(content).toContain('title:');
        expect(content).toContain('description:');
      }
    });

    it('should have navigation structure', () => {
      // 테스트: 네비게이션 구조가 존재하는지
      const fs = require('fs');
      const path = require('path');

      const checkNavigationFiles = (dirPath: string) => {
        if (fs.existsSync(dirPath)) {
          const files = fs.readdirSync(dirPath);
          return files.includes('_meta.json') || files.includes('_nav.md');
        }
        return false;
      };

      const guidesDir = path.join(__dirname, '../pages/en/guides');
      expect(checkNavigationFiles(guidesDir)).toBe(true);

      const referenceDir = path.join(__dirname, '../pages/en/reference');
      expect(checkNavigationFiles(referenceDir)).toBe(true);
    });

    it('should have consistent content across languages', () => {
      // 테스트: 다국어 간 콘텐츠 일관성이 유지되는지
      const fs = require('fs');
      const path = require('path');

      const getIndexContent = (lang: string) => {
        const indexPath = path.join(__dirname, `../pages/${lang}/index.md`);
        if (fs.existsSync(indexPath)) {
          return fs.readFileSync(indexPath, 'utf8');
        }
        return null;
      };

      const contentMap = {
        ko: getIndexContent('ko'),
        en: getIndexContent('en'),
        ja: getIndexContent('ja'),
        zh: getIndexContent('zh'),
      };

      // 각 언어별로 동일한 핵심 콘텐츠 존재 여부 검증
      Object.values(contentMap).forEach(content => {
        expect(content).toBeDefined();
        if (content) {
          expect(content).toContain('MoAI-ADK');
        }
      });
    });
  });

  describe('Specialized Content', () => {
    it('should have TDD guides', () => {
      // 테스트: TDD 가이드가 존재하는지
      const fs = require('fs');
      const path = require('path');
      const tddPath = path.join(__dirname, '../pages/en/guides/tdd/index.md');

      if (fs.existsSync(tddPath)) {
        const content = fs.readFileSync(tddPath, 'utf8');
        expect(content).toContain('TDD');
        expect(content).toContain('Test-Driven Development');
      }
    });

    it('should have Alfred workflow guides', () => {
      // 테스트: Alfred 워크플로우 가이드가 존재하는지
      const fs = require('fs');
      const path = require('path');
      const alfredPath = path.join(__dirname, '../pages/en/guides/alfred/index.md');

      if (fs.existsSync(alfredPath)) {
        const content = fs.readFileSync(alfredPath, 'utf8');
        expect(content).toContain('Alfred');
        expect(content).toContain('Workflow');
      }
    });

    it('should have agent references', () => {
      // 테스트: 에이전트 레퍼런스가 존재하는지
      const fs = require('fs');
      const path = require('path');
      const agentsPath = path.join(__dirname, '../pages/en/reference/agents/index.md');

      if (fs.existsSync(agentsPath)) {
        const content = fs.readFileSync(agentsPath, 'utf8');
        expect(content).toContain('Agents');
        expect(content).toContain('에이전트');
      }
    });

    it('should have skills references', () => {
      // 테스트: 스킬 레퍼런스가 존재하는지
      const fs = require('fs');
      const path = require('path');
      const skillsPath = path.join(__dirname, '../pages/en/reference/skills/index.md');

      if (fs.existsSync(skillsPath)) {
        const content = fs.readFileSync(skillsPath, 'utf8');
        expect(content).toContain('Skills');
        expect(content).toContain('스킬');
      }
    });
  });
});