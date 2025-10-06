import { describe, test, vi } from 'vitest';
import * as fs from 'node:fs';
import * as path from 'node:path';

describe('Debug Test 3', () => {
  test('should show actual result with parent dir path', async () => {
    const parentDir = path.dirname(process.cwd());
    const testProjectPath = path.join(parentDir, 'moai-test-tmp-' + Date.now());
    fs.mkdirSync(testProjectPath, { recursive: true });

    console.log('Test Path:', testProjectPath);
    console.log('CWD:', process.cwd());

    const options = {
      name: 'test-project',
      yes: true,
      path: testProjectPath,
    };

    const { InitCommand } = await import('@/cli/commands/init/index');
    const { SystemDetector } = await import('@/core/system-checker');
    const detector = new SystemDetector();
    const initCommand = new InitCommand(detector);

    // Mock DoctorCommand
    (initCommand as any).doctorCommand = {
      run: vi.fn().mockResolvedValue({
        allPassed: true,
        results: [],
        missingRequirements: [],
        versionConflicts: [],
        summary: { total: 0, passed: 0, failed: 0 }
      })
    };

    const result = await initCommand.runNonInteractive(options);

    console.log('\n=== RESULT ===');
    console.log('Success:', result.success);
    console.log('Errors:', result.errors);

    // Cleanup
    if (fs.existsSync(testProjectPath)) {
      fs.rmSync(testProjectPath, { recursive: true, force: true });
    }
  }, 30000);
});
