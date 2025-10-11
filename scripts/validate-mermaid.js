#!/usr/bin/env node
/**
 * Mermaid Diagram Validator
 *
 * Extracts and validates all Mermaid diagrams from Markdown files
 */

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import { dirname } from 'path';
import { spawn } from 'child_process';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const DOCS_DIR = path.join(__dirname, '../docs');
const TEMP_DIR = path.join(__dirname, '../.temp-mermaid');

// Colors for output
const colors = {
  reset: '\x1b[0m',
  green: '\x1b[32m',
  red: '\x1b[31m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  gray: '\x1b[90m',
};

function log(message, color = 'reset') {
  console.log(`${colors[color]}${message}${colors.reset}`);
}

/**
 * Extract Mermaid diagrams from markdown file
 */
function extractMermaidDiagrams(filePath) {
  const content = fs.readFileSync(filePath, 'utf8');
  const diagrams = [];
  const regex = /```mermaid\n([\s\S]*?)```/g;
  let match;
  let index = 0;

  while ((match = regex.exec(content)) !== null) {
    diagrams.push({
      index: index++,
      content: match[1].trim(),
      lineStart: content.substring(0, match.index).split('\n').length,
    });
  }

  return diagrams;
}

/**
 * Find all markdown files with Mermaid diagrams
 */
function findMermaidFiles(dir, fileList = []) {
  const files = fs.readdirSync(dir);

  for (const file of files) {
    const filePath = path.join(dir, file);
    const stat = fs.statSync(filePath);

    if (stat.isDirectory()) {
      // Skip certain directories
      if (
        file === 'node_modules' ||
        file === '.vitepress' ||
        file === 'api' ||
        file.startsWith('.')
      ) {
        continue;
      }
      findMermaidFiles(filePath, fileList);
    } else if (file.endsWith('.md')) {
      const diagrams = extractMermaidDiagrams(filePath);
      if (diagrams.length > 0) {
        fileList.push({
          path: filePath,
          relativePath: path.relative(DOCS_DIR, filePath),
          diagrams,
        });
      }
    }
  }

  return fileList;
}

/**
 * Validate a single Mermaid diagram
 */
async function validateDiagram(diagramContent, file, diagramIndex) {
  // Create temp directory if it doesn't exist
  if (!fs.existsSync(TEMP_DIR)) {
    fs.mkdirSync(TEMP_DIR, { recursive: true });
  }

  const tempInputFile = path.join(TEMP_DIR, `diagram-${diagramIndex}.mmd`);
  const tempOutputFile = path.join(TEMP_DIR, `diagram-${diagramIndex}.svg`);

  try {
    // Write diagram to temp file
    fs.writeFileSync(tempInputFile, diagramContent);

    // Try to render with mmdc using spawn
    await new Promise((resolve, reject) => {
      const process = spawn('npx', [
        'mmdc',
        '-i',
        tempInputFile,
        '-o',
        tempOutputFile,
        '-t',
        'neutral',
        '-b',
        'transparent',
      ]);

      let stderr = '';

      process.stderr.on('data', (data) => {
        stderr += data.toString();
      });

      process.on('close', (code) => {
        if (code === 0) {
          resolve();
        } else {
          reject(new Error(stderr || `Process exited with code ${code}`));
        }
      });

      process.on('error', (error) => {
        reject(error);
      });
    });

    // Clean up temp files
    if (fs.existsSync(tempInputFile)) fs.unlinkSync(tempInputFile);
    if (fs.existsSync(tempOutputFile)) fs.unlinkSync(tempOutputFile);

    return { success: true };
  } catch (error) {
    // Clean up temp files
    if (fs.existsSync(tempInputFile)) fs.unlinkSync(tempInputFile);
    if (fs.existsSync(tempOutputFile)) fs.unlinkSync(tempOutputFile);

    return {
      success: false,
      error: error.message,
    };
  }
}

/**
 * Main validation function
 */
async function main() {
  log('\n📊 Mermaid Diagram Validator', 'blue');
  log('='.repeat(50), 'gray');

  const files = findMermaidFiles(DOCS_DIR);

  if (files.length === 0) {
    log('\n✅ No Mermaid diagrams found', 'green');
    return;
  }

  const totalDiagrams = files.reduce((sum, f) => sum + f.diagrams.length, 0);

  log(`\nFound ${totalDiagrams} diagram(s) in ${files.length} file(s)\n`, 'blue');

  let errors = 0;
  let successes = 0;

  for (const file of files) {
    log(`📄 ${file.relativePath}`, 'blue');

    for (const diagram of file.diagrams) {
      const id = `${path.basename(file.path)}-diagram-${diagram.index}`;
      process.stdout.write(`  └─ Diagram ${diagram.index + 1} (line ${diagram.lineStart})... `);

      const result = await validateDiagram(diagram.content, file.path, id);

      if (result.success) {
        log('✅ Valid', 'green');
        successes++;
      } else {
        log('❌ Invalid', 'red');
        log(`     Error: ${result.error}`, 'red');
        errors++;
      }
    }

    console.log();
  }

  // Clean up temp directory
  if (fs.existsSync(TEMP_DIR)) {
    fs.rmSync(TEMP_DIR, { recursive: true });
  }

  // Summary
  log('='.repeat(50), 'gray');
  log('\n📋 Validation Summary', 'blue');
  log(`  ✅ Valid:   ${successes}`, 'green');
  log(`  ❌ Invalid: ${errors}`, errors > 0 ? 'red' : 'gray');
  log(`  📊 Total:   ${totalDiagrams}`, 'blue');

  if (errors > 0) {
    log('\n❌ Mermaid validation failed', 'red');
    process.exit(1);
  } else {
    log('\n✅ All Mermaid diagrams are valid!', 'green');
    process.exit(0);
  }
}

main().catch((error) => {
  log(`\n❌ Unexpected error: ${error.message}`, 'red');
  console.error(error);
  process.exit(1);
});
