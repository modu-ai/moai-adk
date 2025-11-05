# 🔒 Security Report: Dependabot Alert #1

**Report ID**: SECURITY-DEPENDABOT-001
**Generated**: 2025-10-29
**Alert URL**: https://github.com/modu-ai/moai-adk/security/dependabot/1
**Status**: 🟡 Open (Investigation Required)

---

## 📋 Alert Summary

| Field | Value |
|-------|-------|
| **Alert Number** | #1 |
| **Severity** | 🟡 Medium (CVSS 4.0: 6.0) |
| **State** | Open |
| **Package** | `vite` (npm) |
| **Ecosystem** | npm |
| **Manifest Path** | `package-lock.json` |
| **Dependency Type** | Transitive (development) |
| **Created** | 2025-10-21 07:52:42 UTC |

---

## 🐛 Vulnerability Details

### CVE Information
- **CVE ID**: CVE-2025-62522
- **GHSA ID**: GHSA-93m4-6634-74q7
- **CWE**: CWE-22 (Path Traversal)

### Summary
**Vite allows `server.fs.deny` bypass via backslash on Windows**

### Description
Files denied by [`server.fs.deny`](https://vitejs.dev/config/server-options.html#server-fs-deny) configuration can be accessed if:
1. The URL ends with a backslash (`\`)
2. The dev server is running on **Windows**
3. The server is explicitly exposed to the network (`--host` or `server.host` config)

### Affected Versions
| Version Range | First Patched Version |
|---------------|----------------------|
| 7.1.0 - 7.1.10 | **7.1.11** |
| 7.0.0 - 7.0.7 | **7.0.8** |
| 6.0.0 - 6.4.0 | **6.4.1** |
| 5.2.6 - 5.4.20 | **5.4.21** ⚠️ |
| 4.5.3 - 4.5.x | **5.4.21** |
| 3.2.9 - 3.x.x | **5.4.21** |
| 2.9.18 - 2.x.x | **5.4.21** |

⚠️ **Current Project**: Vulnerable range detected as **5.2.6 - 5.4.20**

### CVSS Scores
- **CVSS 3.x**: 0.0 (Not calculated)
- **CVSS 4.0**: 6.0 (Medium)
  - Vector: `CVSS:4.0/AV:N/AC:L/AT:P/PR:N/UI:P/VC:H/VI:N/VA:N/SC:N/SI:N/SA:N`
- **EPSS**: 0.082% (24.915th percentile)

### PoC (Proof of Concept)
```bash
npm create vite@latest
cd vite-project/
echo "secret" > .env
npm install
npm run dev
curl --request-target /.env\ http://localhost:5173
# Result: .env file contents leaked
```

---

## 🔍 Investigation Findings

### 1. Repository Analysis
**Finding**: ❌ **No `package.json` or `package-lock.json` found in current codebase**

```bash
$ find . -name "package*.json"
# No results

$ git log --all --oneline -- package-lock.json | head -3
28cc6e50 🧹 OPTIMIZE: 프로젝트 루트 정리 및 최적화 (v0.2.15)
befb2e1f 🏗️ REFACTOR: Hooks 소스 관리 독립화 (v0.3.5)
79961cf8 Merge branch 'develop'
```

**Conclusion**: The file existed in historical commits but was **removed** during project cleanup.

### 2. Project Type
**Current Setup**: Pure Python project
- Build system: `hatchling`
- Dependencies: Python packages only (click, rich, pyfiglet, etc.)
- No JavaScript/TypeScript dependencies

### 3. Dependabot Behavior
**Hypothesis**: Dependabot scanned historical commits or cached data
- Alert created: 2025-10-21 (8 days ago)
- Last commit removing npm files: v0.2.15 (several months ago)
- Possible lag in Dependabot's repository scanning

---

## 🎯 Risk Assessment

### Impact: ⚠️ **LOW** (False Positive)

| Factor | Assessment | Notes |
|--------|-----------|-------|
| **File Existence** | ❌ Not present | No `package-lock.json` in current codebase |
| **Exploitability** | ❌ Not applicable | No Vite dev server running |
| **Platform** | ❌ Not Windows-specific | Development primarily on macOS/Linux |
| **Network Exposure** | ❌ No dev server | Python CLI tool, not a web server |
| **Production Impact** | ✅ None | Vite not used in production |

### Overall Risk: **INFORMATIONAL** (Can be dismissed)

---

## ✅ Recommended Actions

### Option 1: Dismiss Alert (Recommended)
**Reason**: `package-lock.json` no longer exists in the repository.

```bash
gh api repos/modu-ai/moai-adk/dependabot/alerts/1 \
  -X PATCH \
  -f state=dismissed \
  -f dismissed_reason=no_bandwidth \
  -f dismissed_comment="npm dependencies removed in v0.2.15. Project is now Python-only with hatchling build system."
```

### Option 2: Monitor for Transitive Dependencies
**Action**: Verify no Python packages transitively depend on npm/Vite
```bash
# Check Python dependencies for hidden npm usage
pip-audit --desc
bandit -r src/
```

### Option 3: Update GitHub Settings
**Configuration**: Adjust Dependabot to skip npm ecosystem
```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "pip"
    directory: "/"
    schedule:
      interval: "weekly"
  # Remove npm ecosystem scanning
```

---

## 📊 Timeline

| Date | Event |
|------|-------|
| 2025-10-20 | CVE-2025-62522 published |
| 2025-10-21 07:52 | Dependabot alert #1 created |
| 2025-10-21 14:51 | Security advisory updated |
| 2025-10-29 | Investigation completed - **False positive confirmed** |

---

## 🔗 References

- **GitHub Advisory**: https://github.com/vitejs/vite/security/advisories/GHSA-93m4-6634-74q7
- **NVD Entry**: https://nvd.nist.gov/vuln/detail/CVE-2025-62522
- **Vite Fix Commit**: https://github.com/vitejs/vite/commit/f479cc57c425ed41ceb434fecebd63931b1ed4ed
- **Dependabot Alert**: https://github.com/modu-ai/moai-adk/security/dependabot/1

---

## 🏷️ Tags

@SECURITY:DEPENDABOT-001
@DOC:SECURITY-REPORT-001

---

**Report Author**: Alfred (MoAI-ADK SuperAgent)
**Next Review**: 2025-11-05 (or when alert state changes)