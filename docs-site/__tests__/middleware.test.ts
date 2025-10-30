// @TEST:NEXTRA-I18N-005
/**
 * Middleware Tests
 *
 * Tests for i18n middleware configuration:
 * - Middleware file structure validation
 * - Routing configuration consistency
 * - Path matcher validation
 *
 * Note: Actual middleware runtime behavior is tested via E2E tests
 * since middleware requires Next.js Web API environment.
 */

import { routing } from '@/i18n/routing';
import * as fs from 'fs';
import * as path from 'path';

describe('i18n Middleware', () => {
  describe('middleware file', () => {
    it('should exist at root directory', () => {
      const middlewarePath = path.join(process.cwd(), 'middleware.ts');
      expect(fs.existsSync(middlewarePath)).toBe(true);
    });

    it('should import createMiddleware from next-intl', () => {
      const middlewarePath = path.join(process.cwd(), 'middleware.ts');
      const content = fs.readFileSync(middlewarePath, 'utf-8');

      expect(content).toContain("import createMiddleware from 'next-intl/middleware'");
    });

    it('should import routing configuration', () => {
      const middlewarePath = path.join(process.cwd(), 'middleware.ts');
      const content = fs.readFileSync(middlewarePath, 'utf-8');

      expect(content).toContain("import { routing } from './i18n/routing'");
    });

    it('should export default middleware', () => {
      const middlewarePath = path.join(process.cwd(), 'middleware.ts');
      const content = fs.readFileSync(middlewarePath, 'utf-8');

      expect(content).toContain('export default createMiddleware(routing)');
    });

    it('should export config with matcher', () => {
      const middlewarePath = path.join(process.cwd(), 'middleware.ts');
      const content = fs.readFileSync(middlewarePath, 'utf-8');

      expect(content).toContain('export const config');
      expect(content).toContain('matcher');
    });
  });

  describe('matcher configuration', () => {
    it('should exclude Next.js internal paths', () => {
      const middlewarePath = path.join(process.cwd(), 'middleware.ts');
      const content = fs.readFileSync(middlewarePath, 'utf-8');

      // Should exclude _next and _vercel
      expect(content).toContain('_next');
      expect(content).toContain('_vercel');
    });

    it('should exclude static files', () => {
      const middlewarePath = path.join(process.cwd(), 'middleware.ts');
      const content = fs.readFileSync(middlewarePath, 'utf-8');

      // Should exclude files with extensions (.*\\..*)
      expect(content).toMatch(/\.\*\\\\\.\.\*/);
    });
  });

  describe('routing integration', () => {
    it('should use same locales as routing config', () => {
      expect(routing.locales).toEqual(['ko', 'en']);
    });

    it('should use same default locale as routing config', () => {
      expect(routing.defaultLocale).toBe('ko');
    });

    it('should use as-needed locale prefix strategy', () => {
      expect(routing.localePrefix).toBe('as-needed');
    });
  });

  describe('middleware tag', () => {
    it('should have correct TAG marker', () => {
      const middlewarePath = path.join(process.cwd(), 'middleware.ts');
      const content = fs.readFileSync(middlewarePath, 'utf-8');

      expect(content).toContain('@CODE:NEXTRA-I18N-006');
    });
  });
});
