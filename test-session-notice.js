// Test script to verify session-notice behavior
const path = require('path');
const fs = require('fs');

const projectRoot = process.cwd();
console.log('Project Root:', projectRoot);
console.log('');

// Test 1: Check .moai directory
const moaiDir = path.join(projectRoot, '.moai');
const moaiExists = fs.existsSync(moaiDir);
console.log(`✓ Check .moai directory: ${moaiExists ? '✅ EXISTS' : '❌ NOT FOUND'}`);
console.log(`  Path: ${moaiDir}`);
console.log('');

// Test 2: Check .claude/commands/alfred directory
const alfredCommands = path.join(projectRoot, '.claude', 'commands', 'alfred');
const alfredExists = fs.existsSync(alfredCommands);
console.log(`✓ Check .claude/commands/alfred: ${alfredExists ? '✅ EXISTS' : '❌ NOT FOUND'}`);
console.log(`  Path: ${alfredCommands}`);
console.log('');

// Test 3: Check old moai path (should not be used)
const moaiCommands = path.join(projectRoot, '.claude', 'commands', 'moai');
const moaiCommandsExists = fs.existsSync(moaiCommands);
console.log(`✓ Check .claude/commands/moai (old path): ${moaiCommandsExists ? '⚠️ EXISTS (not used)' : '✅ NOT FOUND'}`);
console.log(`  Path: ${moaiCommands}`);
console.log('');

// Test 4: isMoAIProject logic
const isMoAIProject = moaiExists && alfredExists;
console.log('='.repeat(50));
console.log(`isMoAIProject() result: ${isMoAIProject ? '✅ TRUE (initialized)' : '❌ FALSE (not initialized)'}`);
console.log('='.repeat(50));
console.log('');

if (isMoAIProject) {
  console.log('✅ PASS: Project should NOT show initialization message');
} else {
  console.log('❌ FAIL: Project WILL show initialization message');
  if (!moaiExists) console.log('  Reason: .moai directory missing');
  if (!alfredExists) console.log('  Reason: .claude/commands/alfred directory missing');
}
