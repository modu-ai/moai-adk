import { describe, test } from 'vitest';

describe('Env Test', () => {
  test('should check environment variables', () => {
    console.log('NODE_ENV:', process.env.NODE_ENV);
    console.log('VITEST:', process.env.VITEST);
    console.log('ALL ENV:', Object.keys(process.env).filter(k => k.includes('TEST') || k.includes('NODE')));
  });
});
