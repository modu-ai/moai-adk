// @TEST:CONTENT-META-001 - Test _meta.json file structure
import { existsSync, readFileSync } from 'fs';
import { join } from 'path';

describe('@TEST:CONTENT-META-001 - _meta.json structure', () => {
  const pagesDir = join(__dirname, '../pages');

  describe('Korean _meta.json files', () => {
    it('should have root _meta.json for Korean', () => {
      const metaPath = join(pagesDir, 'ko/_meta.json');
      expect(existsSync(metaPath)).toBe(true);

      const meta = JSON.parse(readFileSync(metaPath, 'utf-8'));
      expect(meta).toHaveProperty('index');
      expect(meta).toHaveProperty('introduction');
      expect(meta).toHaveProperty('getting-started');
      expect(meta).toHaveProperty('concepts');
      expect(meta).toHaveProperty('workflow');
      expect(meta).toHaveProperty('skills');
      expect(meta).toHaveProperty('agents');
      expect(meta).toHaveProperty('cli');
      expect(meta).toHaveProperty('config');
      expect(meta).toHaveProperty('api');
      expect(meta).toHaveProperty('misc');
    });

    it('should have introduction/_meta.json for Korean', () => {
      const metaPath = join(pagesDir, 'ko/introduction/_meta.json');
      expect(existsSync(metaPath)).toBe(true);

      const meta = JSON.parse(readFileSync(metaPath, 'utf-8'));
      expect(meta).toHaveProperty('overview');
      expect(meta).toHaveProperty('architecture');
      expect(meta).toHaveProperty('concepts');
    });

    it('should have getting-started/_meta.json for Korean', () => {
      const metaPath = join(pagesDir, 'ko/getting-started/_meta.json');
      expect(existsSync(metaPath)).toBe(true);

      const meta = JSON.parse(readFileSync(metaPath, 'utf-8'));
      expect(meta).toHaveProperty('installation');
      expect(meta).toHaveProperty('quick-start');
      expect(meta).toHaveProperty('first-project');
    });

    it('should have concepts/_meta.json for Korean', () => {
      const metaPath = join(pagesDir, 'ko/concepts/_meta.json');
      expect(existsSync(metaPath)).toBe(true);

      const meta = JSON.parse(readFileSync(metaPath, 'utf-8'));
      expect(meta).toHaveProperty('spec-first-tdd');
      expect(meta).toHaveProperty('ears-guide');
      expect(meta).toHaveProperty('tag-system');
    });

    it('should have workflow/_meta.json for Korean', () => {
      const metaPath = join(pagesDir, 'ko/workflow/_meta.json');
      expect(existsSync(metaPath)).toBe(true);

      const meta = JSON.parse(readFileSync(metaPath, 'utf-8'));
      expect(meta).toHaveProperty('overview');
      expect(meta).toHaveProperty('0-project');
      expect(meta).toHaveProperty('1-plan');
      expect(meta).toHaveProperty('2-run');
      expect(meta).toHaveProperty('3-sync');
    });

    it('should have skills/_meta.json for Korean', () => {
      const metaPath = join(pagesDir, 'ko/skills/_meta.json');
      expect(existsSync(metaPath)).toBe(true);

      const meta = JSON.parse(readFileSync(metaPath, 'utf-8'));
      expect(meta).toHaveProperty('overview');
      expect(meta).toHaveProperty('foundation');
      expect(meta).toHaveProperty('essentials');
      expect(meta).toHaveProperty('alfred');
      expect(meta).toHaveProperty('domain');
    });

    it('should have agents/_meta.json for Korean', () => {
      const metaPath = join(pagesDir, 'ko/agents/_meta.json');
      expect(existsSync(metaPath)).toBe(true);

      const meta = JSON.parse(readFileSync(metaPath, 'utf-8'));
      expect(meta).toHaveProperty('overview');
      expect(meta).toHaveProperty('spec-builder');
      expect(meta).toHaveProperty('git-manager');
    });

    it('should have cli/_meta.json for Korean', () => {
      const metaPath = join(pagesDir, 'ko/cli/_meta.json');
      expect(existsSync(metaPath)).toBe(true);

      const meta = JSON.parse(readFileSync(metaPath, 'utf-8'));
      expect(meta).toHaveProperty('commands');
      expect(meta).toHaveProperty('alfred-0-project');
      expect(meta).toHaveProperty('alfred-1-plan');
      expect(meta).toHaveProperty('alfred-2-run');
      expect(meta).toHaveProperty('alfred-3-sync');
    });
  });

  describe('English _meta.json files', () => {
    it('should have root _meta.json for English', () => {
      const metaPath = join(pagesDir, 'en/_meta.json');
      expect(existsSync(metaPath)).toBe(true);

      const meta = JSON.parse(readFileSync(metaPath, 'utf-8'));
      expect(meta).toHaveProperty('index');
      expect(meta).toHaveProperty('introduction');
      expect(meta).toHaveProperty('getting-started');
      expect(meta).toHaveProperty('concepts');
      expect(meta).toHaveProperty('workflow');
      expect(meta).toHaveProperty('skills');
      expect(meta).toHaveProperty('agents');
      expect(meta).toHaveProperty('cli');
      expect(meta).toHaveProperty('config');
      expect(meta).toHaveProperty('api');
      expect(meta).toHaveProperty('misc');
    });

    it('should have introduction/_meta.json for English', () => {
      const metaPath = join(pagesDir, 'en/introduction/_meta.json');
      expect(existsSync(metaPath)).toBe(true);

      const meta = JSON.parse(readFileSync(metaPath, 'utf-8'));
      expect(meta).toHaveProperty('overview');
      expect(meta).toHaveProperty('architecture');
      expect(meta).toHaveProperty('concepts');
    });
  });
});
