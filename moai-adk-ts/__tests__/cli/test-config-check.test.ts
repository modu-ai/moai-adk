import { describe, test, vi } from 'vitest';
import * as fs from 'node:fs';
import * as path from 'node:path';

describe('Config Check Test', () => {
  test('should check actual config file content', async () => {
    // Set test environment
    process.env.NODE_ENV = 'test';

    const testProjectPath = path.join(process.cwd(), '.test-project-config');
    fs.mkdirSync(testProjectPath, { recursive: true });
    fs.writeFileSync(
      path.join(testProjectPath, 'package.json'),
      JSON.stringify({ name: 'test-project' }, null, 2)
    );

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

    console.log('\nResult success:', result.success);
    
    const configPath = path.join(testProjectPath, '.moai', 'config.json');
    if (fs.existsSync(configPath)) {
      const configContent = fs.readFileSync(configPath, 'utf-8');
      console.log('\nConfig file content:');
      console.log(configContent);
    } else {
      console.log('\nConfig file does not exist!');
    }

    // Cleanup
    if (fs.existsSync(testProjectPath)) {
      fs.rmSync(testProjectPath, { recursive: true, force: true });
    }
  }, 30000);
});
