/**
 * RED Phase - Failing Tests for Nextra 3.3.1 _meta.json Format
 *
 * These tests verify that _meta.json files are in correct JSON format
 * compatible with Nextra 3.3.1 documentation system.
 */

const fs = require('fs');
const path = require('path');

describe('Nextra 3.3.1 _meta.json Format Compliance', () => {
  const docsRoot = path.join(__dirname, '../pages');

  test('should have _meta.json in Korean pages directory', () => {
    const metaPath = path.join(docsRoot, 'ko/_meta.json');
    expect(fs.existsSync(metaPath)).toBe(true);
  });

  test('should parse _meta.json as valid JSON', () => {
    const metaPath = path.join(docsRoot, 'ko/_meta.json');
    const content = fs.readFileSync(metaPath, 'utf8');

    // This should fail if file contains JavaScript export syntax
    expect(() => {
      JSON.parse(content);
    }).not.toThrow();
  });

  test('should not contain JavaScript export syntax', () => {
    const metaPath = path.join(docsRoot, 'ko/_meta.json');
    const content = fs.readFileSync(metaPath, 'utf8');

    // Should NOT contain JavaScript syntax
    expect(content).not.toContain('export default');
    expect(content).not.toContain('module.exports');
    expect(content).not.toContain(';');
  });

  test('should have valid Korean navigation structure', () => {
    const metaPath = path.join(docsRoot, 'ko/_meta.json');
    const content = fs.readFileSync(metaPath, 'utf8');
    const meta = JSON.parse(content);

    // Required navigation items
    expect(meta).toHaveProperty('title', '한국어');
    expect(meta).toHaveProperty('getting-started', '시작하기');
    expect(meta).toHaveProperty('features', '주요 기능');
    expect(meta).toHaveProperty('skills', 'Skills');
    expect(meta).toHaveProperty('guides', '가이드');
  });

  test('should match page directory structure', () => {
    const metaPath = path.join(docsRoot, 'ko/_meta.json');
    const koPagesDir = path.join(docsRoot, 'ko');
    const content = fs.readFileSync(metaPath, 'utf8');
    const meta = JSON.parse(content);

    // Get all directories in ko pages (excluding _meta.json)
    const directories = fs.readdirSync(koPagesDir)
      .filter(file => {
        const fullPath = path.join(koPagesDir, file);
        return fs.statSync(fullPath).isDirectory() && file !== '_meta.json';
      });

    // Each directory should have corresponding entry in _meta.json
    directories.forEach(dir => {
      expect(meta).toHaveProperty(dir);
    });
  });
});

describe('Nextra 3.3.1 Server Startup', () => {
  test('should be able to start dev server without _meta errors', async () => {
    // This test will pass if _meta.json format is correct
    const { spawn } = require('child_process');

    return new Promise((resolve, reject) => {
      const server = spawn('npm', ['run', 'dev', '--', '--port=3002'], {
        cwd: path.join(__dirname, '..'),
        stdio: ['ignore', 'pipe', 'pipe']
      });

      let output = '';
      let errorOutput = '';

      server.stdout.on('data', (data) => {
        output += data.toString();
      });

      server.stderr.on('data', (data) => {
        errorOutput += data.toString();
      });

      // Kill server after timeout
      setTimeout(() => {
        server.kill('SIGTERM');

        // Should NOT contain _meta related errors after fix
        expect(errorOutput).not.toContain('_meta');
        expect(errorOutput).not.toContain('Unexpected token');
        expect(errorOutput).not.toContain('JSON');
        expect(errorOutput).not.toContain('SyntaxError');

        // Should contain successful startup indicators
        expect(output).toContain('ready') || expect(output).toContain('started');

        resolve();
      }, 5000);
    });
  }, 10000);

  test('should validate JSON format for all _meta.json files', () => {
    const metaFiles = [
      path.join(docsRoot, 'ko/_meta.json'),
      path.join(docsRoot, 'ko/features/_meta.json'),
      path.join(docsRoot, 'ko/getting-started/_meta.json'),
      path.join(docsRoot, 'ko/output-style/_meta.json'),
      path.join(docsRoot, 'ko/skills/_meta.json')
    ];

    metaFiles.forEach(filePath => {
      if (fs.existsSync(filePath)) {
        const content = fs.readFileSync(filePath, 'utf8');
        expect(() => {
          JSON.parse(content);
        }).not.toThrow(`Failed to parse ${filePath}`);
      }
    });
  });
});