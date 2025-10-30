// @TEST:NEXTRA-I18N-003
/**
 * Message Loading Tests
 *
 * Tests for translation message files:
 * - Message file structure validation
 * - Key consistency between locales
 * - Required namespaces presence
 */

import koMessages from '@/messages/ko.json';
import enMessages from '@/messages/en.json';

describe('Translation Messages', () => {
  describe('message files', () => {
    it('should load Korean messages', () => {
      expect(koMessages).toBeDefined();
      expect(typeof koMessages).toBe('object');
    });

    it('should load English messages', () => {
      expect(enMessages).toBeDefined();
      expect(typeof enMessages).toBe('object');
    });
  });

  describe('message structure', () => {
    it('should have common namespace in both locales', () => {
      expect(koMessages.common).toBeDefined();
      expect(enMessages.common).toBeDefined();
    });

    it('should have navigation namespace in both locales', () => {
      expect(koMessages.navigation).toBeDefined();
      expect(enMessages.navigation).toBeDefined();
    });

    it('should have home namespace in both locales', () => {
      expect(koMessages.home).toBeDefined();
      expect(enMessages.home).toBeDefined();
    });

    it('should have footer namespace in both locales', () => {
      expect(koMessages.footer).toBeDefined();
      expect(enMessages.footer).toBeDefined();
    });
  });

  describe('key consistency', () => {
    it('should have same keys in common namespace', () => {
      const koKeys = Object.keys(koMessages.common).sort();
      const enKeys = Object.keys(enMessages.common).sort();
      expect(koKeys).toEqual(enKeys);
    });

    it('should have same keys in navigation namespace', () => {
      const koKeys = Object.keys(koMessages.navigation).sort();
      const enKeys = Object.keys(enMessages.navigation).sort();
      expect(koKeys).toEqual(enKeys);
    });

    it('should have same keys in home namespace', () => {
      const koKeys = Object.keys(koMessages.home).sort();
      const enKeys = Object.keys(enMessages.home).sort();
      expect(koKeys).toEqual(enKeys);
    });

    it('should have same keys in footer namespace', () => {
      const koKeys = Object.keys(koMessages.footer).sort();
      const enKeys = Object.keys(enMessages.footer).sort();
      expect(koKeys).toEqual(enKeys);
    });
  });

  describe('required translations', () => {
    it('should have home translation in common namespace', () => {
      expect(koMessages.common.home).toBe('홈');
      expect(enMessages.common.home).toBe('Home');
    });

    it('should have docs translation in common namespace', () => {
      expect(koMessages.common.docs).toBe('문서');
      expect(enMessages.common.docs).toBe('Documentation');
    });

    it('should have title translation in home namespace', () => {
      expect(koMessages.home.title).toBe('MoAI-ADK 문서');
      expect(enMessages.home.title).toBe('MoAI-ADK Documentation');
    });
  });
});
