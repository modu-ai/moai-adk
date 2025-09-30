# Bugfix: Hook Module Resolution - Proper Solution

## 🎯 Problem Summary

**Initial Error**:
```
ReferenceError: module is not defined in ES module scope
```

**Root Cause**: Hooks built as CommonJS (`.js`) conflicted with root `package.json` setting `"type": "module"`

---

## ✅ Final Solution: Local package.json Approach

### Strategy Overview

Instead of renaming files to `.cjs`, we:
1. **Restored TypeScript sources** from git history
2. **Added local `package.json`** to hooks directories
3. **Rebuilt hooks** with proper configuration
4. **Kept `.js` extension** (cleaner, standard)

---

## 📝 Implementation Steps

### Step 1: Restore TypeScript Sources

**Recovered from git commit `c02e55c`**:
```bash
git checkout c02e55c -- moai-adk-ts/src/claude/hooks
```

**Restored files**:
- `src/claude/hooks/session/session-notice.ts`
- `src/claude/hooks/security/policy-block.ts`
- `src/claude/hooks/security/pre-write-guard.ts`
- `src/claude/hooks/security/steering-guard.ts`
- `src/claude/hooks/workflow/file-monitor.ts`
- `src/claude/hooks/workflow/language-detector.ts`

### Step 2: Updated tsup Configuration

**File**: `moai-adk-ts/tsup.hooks.config.ts`

```typescript
import { defineConfig } from 'tsup';

export default defineConfig({
  entry: {
    // Security hooks
    'templates/.claude/hooks/moai/policy-block': 'src/claude/hooks/security/policy-block.ts',
    'templates/.claude/hooks/moai/pre-write-guard': 'src/claude/hooks/security/pre-write-guard.ts',
    'templates/.claude/hooks/moai/steering-guard': 'src/claude/hooks/security/steering-guard.ts',

    // Session hooks
    'templates/.claude/hooks/moai/session-notice': 'src/claude/hooks/session/session-notice.ts',

    // Workflow hooks
    'templates/.claude/hooks/moai/file-monitor': 'src/claude/hooks/workflow/file-monitor.ts',
    'templates/.claude/hooks/moai/language-detector': 'src/claude/hooks/workflow/language-detector.ts',
  },
  format: ['cjs'],
  target: 'node18',
  outDir: '.',
  outExtension: ({ format }) => ({
    js: `.js`, // Force .js extension (package.json declares "type": "commonjs")
  }),
  clean: false,
  sourcemap: false,
  minify: false,
  splitting: false,
  bundle: true,
  external: [],
  platform: 'node',
  shims: true,
});
```

**Key Changes**:
- **Entry points**: All hooks mapped from `src/claude/hooks/`
- **Output**: `templates/.claude/hooks/moai/*.js`
- **outExtension**: Force `.js` even for CommonJS format

### Step 3: Added Local package.json

**File**: `templates/.claude/hooks/moai/package.json`
```json
{
  "type": "commonjs",
  "description": "MoAI-ADK Claude Code Hooks - CommonJS modules"
}
```

**File**: `.claude/hooks/moai/package.json`
```json
{
  "type": "commonjs",
  "description": "MoAI-ADK Claude Code Hooks - CommonJS modules"
}
```

**Effect**: Node.js treats `.js` files in this directory as CommonJS, regardless of parent `package.json`

### Step 4: Reverted .cjs to .js

**Renamed files**:
```bash
.claude/hooks/moai/
├── file-monitor.cjs      → file-monitor.js
├── language-detector.cjs → language-detector.js
├── policy-block.cjs      → policy-block.js
├── pre-write-guard.cjs   → pre-write-guard.js
├── session-notice.cjs    → session-notice.js
├── steering-guard.cjs    → steering-guard.js
└── tag-enforcer.cjs      → tag-enforcer.js
```

**Updated settings.json**:
```json
{
  "hooks": {
    "SessionStart": [{
      "hooks": [{
        "command": "node $CLAUDE_PROJECT_DIR/.claude/hooks/moai/session-notice.js",
        "type": "command"
      }]
    }]
  }
}
```

### Step 5: Build & Copy

```bash
cd moai-adk-ts
bun run build:hooks
cp templates/.claude/hooks/moai/*.js ../.claude/hooks/moai/
```

**Build Output**:
```
✓ templates/.claude/hooks/moai/policy-block.js      46.24 KB
✓ templates/.claude/hooks/moai/session-notice.js     9.18 KB
✓ templates/.claude/hooks/moai/steering-guard.js    46.27 KB
✓ templates/.claude/hooks/moai/pre-write-guard.js   46.26 KB
✓ templates/.claude/hooks/moai/file-monitor.js        6.10 KB
✓ templates/.claude/hooks/moai/language-detector.js   7.99 KB
```

---

## 🧪 Verification

### Test Command
```bash
node .claude/hooks/moai/session-notice.js
```

### Expected Output
```
🗿 MoAI-ADK 프로젝트: moai-adk-ts
🌿 현재 브랜치: develop
📝 변경사항: 243개 파일
📝 SPEC 진행률: 0/0
✅ 통합 체크포인트 시스템 사용 가능
```

### Result
✅ **Success!** No module errors, clean execution

---

## 📊 Solution Comparison

| Approach | Pros | Cons | Status |
|----------|------|------|--------|
| **Rename to .cjs** | Quick fix | Non-standard, harder to maintain | ❌ Reverted |
| **Local package.json** | Standard, maintainable, proper separation | Requires extra file | ✅ **Implemented** |
| **Build as ESM** | Matches root config | Complex refactoring, compatibility issues | ❌ Not needed |

---

## 🎯 Why This Solution is Better

### 1. **Proper Module System Declaration**
- Each directory explicitly declares its module system
- No ambiguity for Node.js
- Works with any parent `package.json` settings

### 2. **Standard File Extensions**
- `.js` is standard for Node.js scripts
- No special handling needed in tooling
- Better IDE support

### 3. **Maintainable Build Pipeline**
- TypeScript sources in `src/claude/hooks/`
- Clear build configuration in `tsup.hooks.config.ts`
- Easy to add new hooks

### 4. **Future-Proof**
- Works with current Node.js LTS (v18+)
- Compatible with future Node.js versions
- Follows Node.js module resolution spec

---

## 📁 File Structure

```
moai-adk-ts/
├── src/
│   └── claude/
│       └── hooks/
│           ├── security/
│           │   ├── policy-block.ts
│           │   ├── pre-write-guard.ts
│           │   └── steering-guard.ts
│           ├── session/
│           │   └── session-notice.ts
│           └── workflow/
│               ├── file-monitor.ts
│               └── language-detector.ts
├── templates/
│   └── .claude/
│       └── hooks/
│           └── moai/
│               ├── package.json        # "type": "commonjs"
│               ├── policy-block.js     # Built from TypeScript
│               ├── pre-write-guard.js
│               ├── steering-guard.js
│               ├── session-notice.js
│               ├── file-monitor.js
│               └── language-detector.js
└── tsup.hooks.config.ts               # Build configuration
```

---

## 🔄 Build Workflow

1. **Edit TypeScript source** in `src/claude/hooks/`
2. **Run build**: `bun run build:hooks`
3. **Copy to root**: `cp templates/.claude/hooks/moai/*.js ../.claude/hooks/moai/`
4. **Test**: `node .claude/hooks/moai/session-notice.js`

---

## 📚 References

- **Node.js Packages**: https://nodejs.org/api/packages.html#type
- **Node.js Module Resolution**: https://nodejs.org/api/modules.html
- **tsup Documentation**: https://tsup.egoist.dev/

---

## ✅ Resolution Checklist

- [x] TypeScript sources restored from git
- [x] tsup.hooks.config.ts updated
- [x] Local package.json added to hooks directories
- [x] Files renamed from .cjs to .js
- [x] settings.json updated
- [x] Hooks rebuilt successfully
- [x] Hook execution verified
- [x] No module resolution errors

---

**Status**: ✅ **Resolved**
**Method**: Local package.json with CommonJS declaration
**Date**: 2025-09-30
**By**: Claude Code Agent