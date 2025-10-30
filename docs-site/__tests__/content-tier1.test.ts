// @TEST:CONTENT-TIER1-001 - Test Tier 1 pages (Home, Introduction, Getting Started, Concepts, Workflow)
import { existsSync, readFileSync } from 'fs';
import { join } from 'path';

describe('@TEST:CONTENT-TIER1-001 - Tier 1 pages structure', () => {
  const pagesDir = join(__dirname, '../pages');

  describe('Korean Tier 1 pages', () => {
    // Home (1 page)
    it('should have index.mdx for Korean', () => {
      const pagePath = join(pagesDir, 'ko/index.mdx');
      expect(existsSync(pagePath)).toBe(true);

      const content = readFileSync(pagePath, 'utf-8');
      expect(content).toContain('# MoAI-ADK');
    });

    // Introduction (3 pages)
    it('should have introduction/overview.mdx for Korean', () => {
      const pagePath = join(pagesDir, 'ko/introduction/overview.mdx');
      expect(existsSync(pagePath)).toBe(true);

      const content = readFileSync(pagePath, 'utf-8');
      expect(content).toContain('---');
      expect(content).toContain('title:');
    });

    it('should have introduction/architecture.mdx for Korean', () => {
      const pagePath = join(pagesDir, 'ko/introduction/architecture.mdx');
      expect(existsSync(pagePath)).toBe(true);
    });

    it('should have introduction/concepts.mdx for Korean', () => {
      const pagePath = join(pagesDir, 'ko/introduction/concepts.mdx');
      expect(existsSync(pagePath)).toBe(true);
    });

    // Getting Started (3 pages)
    it('should have getting-started/installation.mdx for Korean', () => {
      const pagePath = join(pagesDir, 'ko/getting-started/installation.mdx');
      expect(existsSync(pagePath)).toBe(true);
    });

    it('should have getting-started/quick-start.mdx for Korean', () => {
      const pagePath = join(pagesDir, 'ko/getting-started/quick-start.mdx');
      expect(existsSync(pagePath)).toBe(true);
    });

    it('should have getting-started/first-project.mdx for Korean', () => {
      const pagePath = join(pagesDir, 'ko/getting-started/first-project.mdx');
      expect(existsSync(pagePath)).toBe(true);
    });

    // Concepts (3 pages)
    it('should have concepts/spec-first-tdd.mdx for Korean', () => {
      const pagePath = join(pagesDir, 'ko/concepts/spec-first-tdd.mdx');
      expect(existsSync(pagePath)).toBe(true);
    });

    it('should have concepts/ears-guide.mdx for Korean', () => {
      const pagePath = join(pagesDir, 'ko/concepts/ears-guide.mdx');
      expect(existsSync(pagePath)).toBe(true);
    });

    it('should have concepts/tag-system.mdx for Korean', () => {
      const pagePath = join(pagesDir, 'ko/concepts/tag-system.mdx');
      expect(existsSync(pagePath)).toBe(true);
    });

    // Workflow (5 pages)
    it('should have workflow/overview.mdx for Korean', () => {
      const pagePath = join(pagesDir, 'ko/workflow/overview.mdx');
      expect(existsSync(pagePath)).toBe(true);
    });

    it('should have workflow/0-project.mdx for Korean', () => {
      const pagePath = join(pagesDir, 'ko/workflow/0-project.mdx');
      expect(existsSync(pagePath)).toBe(true);
    });

    it('should have workflow/1-plan.mdx for Korean', () => {
      const pagePath = join(pagesDir, 'ko/workflow/1-plan.mdx');
      expect(existsSync(pagePath)).toBe(true);
    });

    it('should have workflow/2-run.mdx for Korean', () => {
      const pagePath = join(pagesDir, 'ko/workflow/2-run.mdx');
      expect(existsSync(pagePath)).toBe(true);
    });

    it('should have workflow/3-sync.mdx for Korean', () => {
      const pagePath = join(pagesDir, 'ko/workflow/3-sync.mdx');
      expect(existsSync(pagePath)).toBe(true);
    });
  });

  describe('English Tier 1 pages', () => {
    // Home (1 page)
    it('should have index.mdx for English', () => {
      const pagePath = join(pagesDir, 'en/index.mdx');
      expect(existsSync(pagePath)).toBe(true);

      const content = readFileSync(pagePath, 'utf-8');
      expect(content).toContain('# MoAI-ADK');
    });

    // Introduction (3 pages)
    it('should have introduction/overview.mdx for English', () => {
      const pagePath = join(pagesDir, 'en/introduction/overview.mdx');
      expect(existsSync(pagePath)).toBe(true);
    });

    it('should have introduction/architecture.mdx for English', () => {
      const pagePath = join(pagesDir, 'en/introduction/architecture.mdx');
      expect(existsSync(pagePath)).toBe(true);
    });

    it('should have introduction/concepts.mdx for English', () => {
      const pagePath = join(pagesDir, 'en/introduction/concepts.mdx');
      expect(existsSync(pagePath)).toBe(true);
    });
  });
});
