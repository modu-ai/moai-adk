/**
 * @file Vitest test setup and extensions
 * @author MoAI Team
 * @tags @TEST:SETUP-001
 */

import { expect } from 'vitest';

// Extend Vitest matchers for better testing
interface CustomMatchers<R = unknown> {
  toBeOneOf(expected: any[]): R;
}

declare module 'vitest' {
  interface Assertion<T = any> extends CustomMatchers<T> {}
  interface AsymmetricMatchersContaining extends CustomMatchers {}
}

// Custom matcher for "one of" assertions
expect.extend({
  toBeOneOf(received: any, expected: any[]) {
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

export {};
