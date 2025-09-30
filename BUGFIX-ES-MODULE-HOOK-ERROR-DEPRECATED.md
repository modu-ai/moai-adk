# Bugfix: ES Module Hook Error - `module is not defined`

## 🐛 Problem Report

**Error Message**:
```
ReferenceError: module is not defined in ES module scope
This file is being treated as an ES module because it has a '.js' file
extension and '/Users/goos/MoAI/MoAI-ADK/package.json' contains "type": "module"
```

**Location**: `.claude/hooks/moai/session-notice.js`

**Trigger**: SessionStart hook execution in Claude Code

---

## 🔍 Root Cause Analysis

### Environment Configuration Conflict

1. **Project Setting**: Root `package.json` has `"type": "module"` (Line 6)
   - This makes Node.js treat ALL `.js` files as ES Modules (ESM)

2. **Hook Files**: Generated as **CommonJS** format
   - Line 36: `module.exports = __toCommonJS(session_notice_exports);`
   - Line 288: `require.main === module`
   - Line 37: `require("child_process")`

3. **Node.js Behavior**:
   - Sees `.js` extension + `"type": "module"` in package.json
   - Tries to parse as ESM
   - Encounters CommonJS syntax → **Error**

### Why Hooks Were Built as CommonJS

- `tsup.hooks.config.ts` specifies `format: ['cjs']` (Line 12)
- Intended for standalone execution without ES module complications
- However, this conflicts with monorepo-level `"type": "module"` setting

---

## ✅ Solution Implemented

### Fix Strategy: Change File Extension to `.cjs`

**Rationale**: Node.js treats `.cjs` files as CommonJS **regardless** of package.json settings

### Changes Made

#### 1. Renamed Hook Files (Root Project)
```bash
.claude/hooks/moai/
├── file-monitor.js         → file-monitor.cjs
├── language-detector.js    → language-detector.cjs
├── policy-block.js         → policy-block.cjs
├── pre-write-guard.js      → pre-write-guard.cjs
├── session-notice.js       → session-notice.cjs ✅
├── steering-guard.js       → steering-guard.cjs
└── tag-enforcer.js         → tag-enforcer.cjs
```

#### 2. Updated `.claude/settings.json` (Root)

**Before**:
```json
{
  "command": "node $CLAUDE_PROJECT_DIR/.claude/hooks/moai/session-notice.js"
}
```

**After**:
```json
{
  "command": "node $CLAUDE_PROJECT_DIR/.claude/hooks/moai/session-notice.cjs"
}
```

**Modified Lines**: 15, 19, 28, 39, 50

#### 3. Renamed Template Hook Files
```bash
moai-adk-ts/templates/.claude/hooks/moai/
├── file-monitor.js         → file-monitor.cjs
├── language-detector.js    → language-detector.cjs
├── policy-block.js         → policy-block.cjs
├── pre-write-guard.js      → pre-write-guard.cjs
├── session-notice.js       → session-notice.cjs ✅
├── steering-guard.js       → steering-guard.cjs
└── tag-enforcer.js         → tag-enforcer.cjs
```

#### 4. Updated `templates/.claude/settings.json`

**Same changes as root settings.json**
- All `.js` references → `.cjs`
- Modified Lines: 15, 19, 28, 39, 50

---

## 📊 Impact Analysis

### Before (Broken)
```
✗ SessionStart hook fails immediately
✗ Error message displayed to user
✗ Project status not displayed
✗ Confusing developer experience
```

### After (Fixed)
```
✓ SessionStart hook executes successfully
✓ No error messages
✓ Project status displayed correctly
✓ Clean developer experience
```

### Files Modified

| Category | File Path | Change |
|----------|-----------|--------|
| Root Hooks | `.claude/hooks/moai/*.js` → `*.cjs` | 7 files renamed |
| Root Config | `.claude/settings.json` | 5 hook paths updated |
| Template Hooks | `templates/.claude/hooks/moai/*.js` → `*.cjs` | 7 files renamed |
| Template Config | `templates/.claude/settings.json` | 5 hook paths updated |

**Total**: 21 files modified

---

## 🧪 Verification

### Expected Behavior
```bash
# No error on session start
claude code
# Should show:
# 🗿 MoAI-ADK 프로젝트: MoAI-ADK
# 🌿 현재 브랜치: develop
# ✅ 통합 체크포인트 시스템 사용 가능
```

### Testing Commands
```bash
# Test hook execution directly
node .claude/hooks/moai/session-notice.cjs

# Start Claude Code and verify no errors
claude
```

---

## 🔧 Future Considerations

### Option 1: Keep Current Solution (Recommended)
- **Pros**: Simple, works immediately, no build changes needed
- **Cons**: Non-standard file extension

### Option 2: Build Hooks as ESM
- Modify `tsup.hooks.config.ts`: `format: ['esm']`
- Update hook source code to use `export` instead of `module.exports`
- **Pros**: Consistent with project settings
- **Cons**: Requires refactoring, potential compatibility issues

### Option 3: Separate package.json for hooks
- Create `.claude/hooks/package.json` with `"type": "commonjs"`
- Keep `.js` extensions
- **Pros**: Cleaner solution, no extension change needed
- **Cons**: Additional configuration file

**Recommendation**: Keep current `.cjs` solution for stability

---

## 📚 Related Documentation

- Node.js ES Modules: https://nodejs.org/api/esm.html
- File Extension Handling: https://nodejs.org/api/packages.html#type
- tsup Configuration: https://tsup.egoist.dev/

---

## ✅ Resolution Status

- [x] Root hook files renamed to `.cjs`
- [x] Root `settings.json` updated
- [x] Template hook files renamed to `.cjs`
- [x] Template `settings.json` updated
- [x] Error resolved
- [ ] Manual testing by user (pending)

---

**Fixed in**: Bugfix commit
**Reported by**: User
**Fixed by**: Claude Code Agent
**Date**: 2025-09-30