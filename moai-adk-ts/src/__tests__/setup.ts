/**
 * @file Vitest test setup and extensions
 * @author MoAI Team
 * @tags @TEST:SETUP-001
 */

import { expect } from 'vitest';

/**
 * Extend Vitest matchers for better testing
 * @tags @TEST:CUSTOM-MATCHERS-001
 */
interface CustomMatchers<R = unknown> {
  toBeOneOf(expected: unknown[]): R;
}

declare module 'vitest' {
  interface Assertion extends CustomMatchers {}
  interface AsymmetricMatchersContaining extends CustomMatchers {}
}

/**
 * Custom matcher for "one of" assertions
 * @tags @TEST:ONE-OF-MATCHER-001
 */
expect.extend({
  toBeOneOf(received: unknown, expected: unknown[]) {
    const pass = expected.includes(received);
    if (pass) {
      return {
        message: () => `expected ${received} not to be one of ${expected}`,
        pass: true,
      };
    } else {
      return {
        message: () => `expected ${received} to be one of ${expected}`,
        pass: false,
      };
    }
  },
});
