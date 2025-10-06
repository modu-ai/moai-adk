import { describe, test, expect, vi } from 'vitest';
import * as fs from 'node:fs';
import * as path from 'node:path';
import * as os from 'node:os';

describe('Debug Test', () => {
  test('should show actual result', async () => {
    const testProjectPath = path.join(os.tmpdir(), '.test-project-init-' + Date.now());
    fs.mkdirSync(testProjectPath, { recursive: true });

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
    console.log(JSON.stringify(result, null, 2));
    console.log('Success:', result.success);
    console.log('Errors:', result.errors);

    // Cleanup
    if (fs.existsSync(testProjectPath)) {
      fs.rmSync(testProjectPath, { recursive: true, force: true });
    }
  }, 30000);
});
