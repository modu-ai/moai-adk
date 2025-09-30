// @FEATURE:TAG-VAL-001 | Chain: @REQ:TAG-001 -> @DESIGN:TAG-001 -> @TASK:TAG-001 -> @TEST:TAG-001
// Related: @API:TAG-001

/**
 * @file TAG system validation script
 * @author MoAI Team
 */

import { execSync } from 'node:child_process';
import * as fs from 'node:fs';
import * as path from 'node:path';

interface ValidationResult {
  totalFiles: number;
  filesWithTags: number;
  completeTAGBlocks: number;
  incompleteTAGBlocks: number;
  filesWithoutTags: string[];
  orphanedTags: Array<{ file: string; tag: string; line: number }>;
  brokenChains: Array<{ file: string; issue: string }>;
  invalidFormatTags: Array<{ file: string; tag: string; line: number }>;
}

const TAG_BLOCK_PATTERN = /^\/\/ @FEATURE:(\w+-\d{3}) \| Chain: @REQ:(\w+-\d{3}) -> @DESIGN:(\w+-\d{3}) -> @TASK:(\w+-\d{3}) -> @TEST:(\w+-\d{3})/;
const TAG_PATTERN = /@(REQ|DESIGN|TASK|TEST|FEATURE|API|UI|DATA):(\w+-\d{3})/g;
const DOMAIN_PATTERN = /^[A-Z]{2,6}-\d{3}$/;

function getAllTypeScriptFiles(dir: string): string[] {
  const files: string[] = [];
  const entries = fs.readdirSync(dir, { withFileTypes: true });

  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name);

    if (entry.isDirectory()) {
      // Skip node_modules, dist, coverage
      if (!['node_modules', 'dist', 'coverage', '.git'].includes(entry.name)) {
        files.push(...getAllTypeScriptFiles(fullPath));
      }
    } else if (entry.name.endsWith('.ts') && !entry.name.endsWith('.d.ts')) {
      files.push(fullPath);
    }
  }

  return files;
}

function validateFile(filePath: string): {
  hasTagBlock: boolean;
  isComplete: boolean;
  tags: Array<{ type: string; id: string; line: number }>;
  issues: string[];
} {
  const content = fs.readFileSync(filePath, 'utf-8');
  const lines = content.split('\n');

  const result = {
    hasTagBlock: false,
    isComplete: false,
    tags: [] as Array<{ type: string; id: string; line: number }>,
    issues: [] as string[],
  };

  // Check first few lines for TAG BLOCK
  const firstFewLines = lines.slice(0, 10).join('\n');
  const tagBlockMatch = firstFewLines.match(TAG_BLOCK_PATTERN);

  if (tagBlockMatch) {
    result.hasTagBlock = true;

    // Validate TAG BLOCK completeness
    const [, featureId, reqId, designId, taskId, testId] = tagBlockMatch;

    // Check if all IDs use the same domain
    const baseDomain = featureId.split('-')[0];
    const allSameDomain = [reqId, designId, taskId, testId].every(
      id => id.startsWith(baseDomain)
    );

    // Check if domain follows naming convention
    if (!DOMAIN_PATTERN.test(featureId)) {
      result.issues.push(`Invalid TAG ID format: ${featureId}`);
    }

    result.isComplete = allSameDomain;
    if (!allSameDomain) {
      result.issues.push(`Inconsistent domain in TAG chain`);
    }
  } else {
    // Check if file has any TAG annotations (incomplete)
    if (firstFewLines.includes('@FEATURE:') ||
        firstFewLines.includes('@REQ:') ||
        firstFewLines.includes('@tags')) {
      result.hasTagBlock = true; // Partial TAG exists
      result.issues.push('Incomplete TAG BLOCK format');
    }
  }

  // Extract all TAGs in file
  let match;
  let lineNum = 0;
  for (const line of lines) {
    lineNum++;
    const regex = /@(REQ|DESIGN|TASK|TEST|FEATURE|API|UI|DATA):(\w+-\d{3})/g;
    while ((match = regex.exec(line)) !== null) {
      result.tags.push({
        type: match[1],
        id: match[2],
        line: lineNum,
      });
    }
  }

  return result;
}

function runValidation(): ValidationResult {
  const srcDir = path.join(process.cwd(), 'src');
  const files = getAllTypeScriptFiles(srcDir);

  const result: ValidationResult = {
    totalFiles: files.length,
    filesWithTags: 0,
    completeTAGBlocks: 0,
    incompleteTAGBlocks: 0,
    filesWithoutTags: [],
    orphanedTags: [],
    brokenChains: [],
    invalidFormatTags: [],
  };

  for (const file of files) {
    const relativePath = path.relative(process.cwd(), file);
    const validation = validateFile(file);

    if (validation.hasTagBlock) {
      result.filesWithTags++;

      if (validation.isComplete) {
        result.completeTAGBlocks++;
      } else {
        result.incompleteTAGBlocks++;
        result.brokenChains.push({
          file: relativePath,
          issue: validation.issues.join(', '),
        });
      }
    } else {
      result.filesWithoutTags.push(relativePath);
    }

    // Check for orphaned TAGs (TAGs not in first TAG BLOCK)
    if (validation.tags.length > 0 && !validation.isComplete) {
      for (const tag of validation.tags) {
        if (tag.line > 10) { // TAGs after line 10 might be orphaned
          result.orphanedTags.push({
            file: relativePath,
            tag: `@${tag.type}:${tag.id}`,
            line: tag.line,
          });
        }
      }
    }
  }

  return result;
}

function generateReport(result: ValidationResult): string {
  const percentage = Math.round((result.completeTAGBlocks / result.totalFiles) * 100);

  let report = `
╔═══════════════════════════════════════════════════════════════════════╗
║                    TAG SYSTEM VALIDATION REPORT                        ║
╚═══════════════════════════════════════════════════════════════════════╝

📊 STATISTICS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Total Files:              ${result.totalFiles}
  Files with TAGs:          ${result.filesWithTags} (${Math.round((result.filesWithTags/result.totalFiles)*100)}%)
  Complete TAG BLOCKs:      ${result.completeTAGBlocks} (${percentage}%)
  Incomplete TAG BLOCKs:    ${result.incompleteTAGBlocks}
  Files without TAGs:       ${result.filesWithoutTags.length}
  Orphaned TAGs:            ${result.orphanedTags.length}
  Broken Chains:            ${result.brokenChains.length}

`;

  if (result.filesWithoutTags.length > 0) {
    report += `
⚠️  FILES WITHOUT TAG BLOCKS (${result.filesWithoutTags.length})
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`;
    for (const file of result.filesWithoutTags.slice(0, 10)) {
      report += `  • ${file}\n`;
    }
    if (result.filesWithoutTags.length > 10) {
      report += `  ... and ${result.filesWithoutTags.length - 10} more\n`;
    }
  }

  if (result.brokenChains.length > 0) {
    report += `
🔗 BROKEN CHAINS (${result.brokenChains.length})
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`;
    for (const chain of result.brokenChains.slice(0, 10)) {
      report += `  • ${chain.file}\n    Issue: ${chain.issue}\n`;
    }
  }

  if (result.orphanedTags.length > 0) {
    report += `
🏷️  ORPHANED TAGS (${result.orphanedTags.length})
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`;
    for (const orphan of result.orphanedTags.slice(0, 10)) {
      report += `  • ${orphan.file}:${orphan.line} - ${orphan.tag}\n`;
    }
  }

  report += `
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`;

  if (percentage === 100) {
    report += `
✅ VALIDATION PASSED - 100% TAG BLOCK COVERAGE!
`;
  } else if (percentage >= 90) {
    report += `
⚡ EXCELLENT - ${percentage}% TAG BLOCK COVERAGE
`;
  } else if (percentage >= 75) {
    report += `
👍 GOOD - ${percentage}% TAG BLOCK COVERAGE
`;
  } else {
    report += `
⚠️  NEEDS IMPROVEMENT - ${percentage}% TAG BLOCK COVERAGE
`;
  }

  report += `
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`;

  return report;
}

// Main execution
try {
  console.log('\n🔍 Starting TAG system validation...\n');

  const result = runValidation();
  const report = generateReport(result);

  console.log(report);

  // Exit with error if validation fails
  const coverageThreshold = 85;
  const coverage = Math.round((result.completeTAGBlocks / result.totalFiles) * 100);

  if (coverage < coverageThreshold) {
    console.error(`\n❌ Validation failed: TAG BLOCK coverage (${coverage}%) is below threshold (${coverageThreshold}%)\n`);
    process.exit(1);
  }

  if (result.brokenChains.length > 0) {
    console.error(`\n❌ Validation failed: ${result.brokenChains.length} broken TAG chains detected\n`);
    process.exit(1);
  }

  console.log('\n✅ TAG system validation passed!\n');
  process.exit(0);

} catch (error) {
  console.error('\n❌ Validation error:', error);
  process.exit(1);
}
