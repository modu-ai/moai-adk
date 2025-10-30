// @TEST:NEXTRA-I18N-001
/**
 * i18n Configuration Tests
 *
 * Tests for i18n routing configuration:
 * - Locale validation
 * - Default locale settings
 * - Locale prefix configuration
 */

import { routing } from '@/i18n/routing';

describe('i18n Configuration', () => {
  describe('routing configuration', () => {
    it('should define supported locales', () => {
      expect(routing.locales).toBeDefined();
      expect(Array.isArray(routing.locales)).toBe(true);
    });

    it('should support Korean (ko) locale', () => {
      expect(routing.locales).toContain('ko');
    });

    it('should support English (en) locale', () => {
      expect(routing.locales).toContain('en');
    });

    it('should have exactly 2 locales', () => {
      expect(routing.locales).toHaveLength(2);
    });

    it('should set Korean (ko) as default locale', () => {
      expect(routing.defaultLocale).toBe('ko');
    });

    it('should have locale prefix configuration', () => {
      expect(routing.localePrefix).toBeDefined();
    });

    it('should use as-needed locale prefix strategy', () => {
      expect(routing.localePrefix).toBe('as-needed');
    });

    it('should have locales in correct order (ko, en)', () => {
      expect(routing.locales).toEqual(['ko', 'en']);
    });
  });

  describe('locale validation', () => {
    it('should not include unsupported locales', () => {
      expect(routing.locales).not.toContain('ja');
      expect(routing.locales).not.toContain('zh');
      expect(routing.locales).not.toContain('fr');
    });

    it('should have default locale in locales array', () => {
      expect(routing.locales).toContain(routing.defaultLocale);
    });
  });
});
