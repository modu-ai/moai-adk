// Test setup for migration tests
import { vi } from 'vitest';

// Mock Node.js built-in modules
global.describe = describe;
global.it = it;
global.expect = expect;
global.beforeEach = beforeEach;
global.afterEach = afterEach;