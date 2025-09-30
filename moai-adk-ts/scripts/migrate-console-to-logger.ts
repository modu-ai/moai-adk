/**
 * @file Automated console.* to Winston logger migration script
 * @tags @TASK:LOGGER-MIGRATION-001 @PERF:TRUST-S-001
 */

import fs from 'fs-extra';
import path from 'node:path';
import { glob } from 'glob';

interface MigrationResult {
  filePath: string;
  originalConsoleCount: number;
  migrated: boolean;
  error?: string;
}

/**
 * Calculate relative import path for winston-logger
 */
function calculateImportPath(filePath: string): string {
  const fileDir = path.dirname(filePath);
  const utilsPath = path.join(process.cwd(), 'src', 'utils');
  const relativePath = path.relative(fileDir, utilsPath);

  // Normalize path for cross-platform and convert to forward slashes
  const normalizedPath = relativePath.replace(/\\/g, '/');

  // Ensure path starts with ./ or ../
  if (!normalizedPath.startsWith('.')) {
    return `./${normalizedPath}/winston-logger.js`;
  }
  return `${normalizedPath}/winston-logger.js`;
}

/**
 * Check if file already imports winston-logger
 */
function hasLoggerImport(content: string): boolean {
  return /from\s+['"].*winston-logger\.js['"]/.test(content);
}

/**
 * Count console.* usage in content
 */
function countConsoleUsage(content: string): number {
  const matches = content.match(/console\.(log|error|warn|debug|info)\(/g);
  return matches ? matches.length : 0;
}

/**
 * Add logger import at the top of the file
 */
function addLoggerImport(content: string, importPath: string): string {
  // Find the position after the last import statement
  const importRegex = /^import\s+.*?;$/gm;
  const imports = Array.from(content.matchAll(importRegex));

  if (imports.length === 0) {
    // No imports found, add at the beginning
    return `import { logger } from '${importPath}';\n\n${content}`;
  }

  // Add after the last import
  const lastImport = imports[imports.length - 1];
  const lastImportEnd = lastImport.index! + lastImport[0].length;

  return (
    content.slice(0, lastImportEnd) +
    `\nimport { logger } from '${importPath}';` +
    content.slice(lastImportEnd)
  );
}

/**
 * Replace console.* with logger.* methods
 */
function replaceConsoleCalls(content: string): string {
  let modified = content;

  // Replace console.log → logger.info
  modified = modified.replace(/console\.log\(/g, 'logger.info(');

  // Replace console.error → logger.error
  modified = modified.replace(/console\.error\(/g, 'logger.error(');

  // Replace console.warn → logger.warn
  modified = modified.replace(/console\.warn\(/g, 'logger.warn(');

  // Replace console.debug → logger.debug
  modified = modified.replace(/console\.debug\(/g, 'logger.debug(');

  // Replace console.info → logger.info
  modified = modified.replace(/console\.info\(/g, 'logger.info(');

  return modified;
}

/**
 * Migrate a single file
 */
async function migrateFile(filePath: string): Promise<MigrationResult> {
  try {
    const originalContent = await fs.readFile(filePath, 'utf-8');
    const originalConsoleCount = countConsoleUsage(originalContent);

    // Skip if no console.* usage
    if (originalConsoleCount === 0) {
      return {
        filePath,
        originalConsoleCount: 0,
        migrated: false,
      };
    }

    let modifiedContent = originalContent;

    // Add logger import if not present
    if (!hasLoggerImport(originalContent)) {
      const importPath = calculateImportPath(filePath);
      modifiedContent = addLoggerImport(modifiedContent, importPath);
    }

    // Replace console calls
    modifiedContent = replaceConsoleCalls(modifiedContent);

    // Verify all console.* replaced
    const remainingConsole = countConsoleUsage(modifiedContent);
    if (remainingConsole > 0) {
      return {
        filePath,
        originalConsoleCount,
        migrated: false,
        error: `${remainingConsole} console.* calls remain after migration`,
      };
    }

    // Write back
    await fs.writeFile(filePath, modifiedContent, 'utf-8');

    return {
      filePath,
      originalConsoleCount,
      migrated: true,
    };
  } catch (error) {
    return {
      filePath,
      originalConsoleCount: 0,
      migrated: false,
      error: error instanceof Error ? error.message : String(error),
    };
  }
}

/**
 * Main migration function
 */
async function migrateConsoleToLogger(): Promise<void> {
  console.log('🚀 Starting console.* to Winston logger migration...\n');

  // Find all TypeScript files except tests and node_modules
  const files = await glob('src/**/*.ts', {
    ignore: ['**/*.test.ts', '**/node_modules/**', '**/__tests__/**'],
    cwd: process.cwd(),
    absolute: true,
  });

  console.log(`📁 Found ${files.length} TypeScript files to scan\n`);

  const results: MigrationResult[] = [];
  let totalMigrated = 0;
  let totalConsoleCount = 0;

  for (const file of files) {
    const result = await migrateFile(file);
    results.push(result);

    if (result.migrated) {
      totalMigrated++;
      totalConsoleCount += result.originalConsoleCount;
      console.log(`✅ Migrated: ${path.relative(process.cwd(), result.filePath)}`);
      console.log(`   Replaced ${result.originalConsoleCount} console.* calls`);
    } else if (result.error) {
      console.log(`❌ Failed: ${path.relative(process.cwd(), result.filePath)}`);
      console.log(`   Error: ${result.error}`);
    }
  }

  // Summary
  console.log('\n' + '='.repeat(60));
  console.log('📊 Migration Summary:');
  console.log('='.repeat(60));
  console.log(`Total files scanned: ${files.length}`);
  console.log(`Files migrated: ${totalMigrated}`);
  console.log(`Total console.* calls replaced: ${totalConsoleCount}`);
  console.log(`Files with errors: ${results.filter(r => r.error).length}`);

  // Error summary
  const errors = results.filter(r => r.error);
  if (errors.length > 0) {
    console.log('\n❌ Files with errors:');
    for (const error of errors) {
      console.log(`  - ${path.relative(process.cwd(), error.filePath)}`);
      console.log(`    ${error.error}`);
    }
  }

  console.log('\n✅ Migration complete!');
  console.log('\n📋 Next steps:');
  console.log('  1. Run: bun run type-check');
  console.log('  2. Run: bun run build');
  console.log('  3. Run: bun test');
  console.log('  4. Verify: rg "console\\.(log|error|warn|debug)" src/');
}

// Run migration
migrateConsoleToLogger().catch((error) => {
  console.error('Fatal error during migration:', error);
  process.exit(1);
});