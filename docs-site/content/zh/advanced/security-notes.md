---
title: 安全说明
description: "MoAI-ADK v2.20.0-rc1 安全加固 — CWE-732/214/345 映射与用户自检流程"
weight: 72
draft: false
tags: ["security", "cwe", "audit"]
---

# 安全说明 (Security Notes)

本页面整理了 MoAI-ADK v2.20.0-rc1 引入的 **用户可见安全变更**。每一项包含 CWE 映射、行为变更与自检命令。

## 为何存在本页 (Why)

`SPEC-V3R5-SECURITY-CRIT-001` (PR #1032、合并提交 `03a2552a2`) 修复了在 v2.14.0 → v2.20.0-rc1 代码审查中发现的 **3 项 P0 发布阻塞安全缺陷**。本页将这些修复以及用户用于验证保护是否在自身环境中正常工作的程序,作为正式的 4-locale 文档明文化。

3 项缺陷均与 GLM 集成 + 自动更新路径相关:

- **CWE-732 / CWE-552** — `.claude/settings.local.json` 文件权限强制为 **`0o600`** (仅所有者 read/write)。
- **CWE-214** — `moai cg` 的 tmux 环境变量注入改为通过 source-file 而非 argv (GLM token 不再可见于 argv)。
- **CWE-345** — `moai update` 的 checksum 验证强制化 (下载失败将拒绝更新)。

每项修复均由回归测试锁定,阻止未来回退。

## CWE-732 — settings.local.json 权限加固 (Permission Hardening) {#cwe-732}

### 变更内容

`.claude/settings.local.json` 在创建和更新时,文件权限强制为 **`0o600`** (仅所有者 read/write)。之前使用 `0o644` (所有者 read/write + group/world read),在多用户工作站上其他本地用户可读取 `ANTHROPIC_AUTH_TOKEN` 等敏感凭证。

### 威胁模型

- **攻击者**: 同一主机的低权限本地用户
- **攻击面**: `.claude/settings.local.json` 的 group/world read 权限
- **泄露信息**: GLM API token (`ANTHROPIC_AUTH_TOKEN`)、OAuth refresh token、其他 `settings.Env` 值
- **CWE 映射**: CWE-732 (Incorrect Permission Assignment for Critical Resource)、CWE-552 (Files or Directories Accessible to External Parties)

### 实现位置

- `internal/hook/settings_io.go` — 常量 `secureSettingsMode os.FileMode = 0o600` + `writeSettingsSecure` helper
- `internal/hook/session_start.go` — 所有 `settings.local.json` writer (`ensureGLMCredentials`、`ensureClaudeEnvFile`)
- `internal/hook/session_end.go` — GLM keys write-back 路径

### 自检

检查现有 `settings.local.json` 权限:

```bash
# Linux
stat -c '%a' .claude/settings.local.json
# 期望值: 600

# macOS
stat -f '%A' .claude/settings.local.json
# 期望值: 600
```

若权限显示为 `644` 或更宽松值,MoAI-ADK 将在下次会话开始时自动修正为 `0o600`。如需立即修正:

```bash
chmod 0600 .claude/settings.local.json
```

### 影响 (Trade-off)

期望文件为 `group-readable` 的工作流 (同一项目目录被另一 OS 用户读取的极少见场景) 会被破坏。这一权衡是有意为之,安全恢复明显优先。

## CWE-214 — tmux IPC token argv 暴露阻断 {#cwe-214}

### 变更内容

`moai cg` (CG 模式) 在向 tmux 会话环境注入 GLM token (`ANTHROPIC_AUTH_TOKEN`) 时,改用 **source-file 通道** (`tmux source-file <tmp>`) 而非 **argv 通道** (`tmux set-environment <KEY> <VALUE>`)。token 不再以明文形式出现在 `ps auxe`、`/proc/<pid>/cmdline`、auditd 日志、sysmon 跟踪和崩溃转储中。

### 实现流程

1. 在 `~/.moai/run/` 下通过 `mkstemp` 创建临时文件 (默认 mode `0o600` + 显式 `chmod 0o600`)。
2. 向临时文件写入一行 `set-environment -t <session> <KEY> <VALUE>`。
3. 调用 `tmux source-file <tmp>` 让 tmux 读取该文件并注入到环境。
4. 注入完成后立即通过 `os.Remove` unlink 临时文件。

argv 中仅暴露临时文件路径,token 本身永不出现在 argv。

### 威胁模型

- **攻击者**: 同一主机的本地用户 + 系统日志采集 (`ps`、`/proc`、auditd、sysmon)
- **攻击面**: tmux 环境变量注入的 argv 通道
- **泄露信息**: GLM API token 的瞬时可见
- **CWE 映射**: CWE-214 (Invocation of Process Using Visible Sensitive Information)

### 实现位置

- `internal/tmux/session.go` — `InjectSensitiveEnv` 方法、`sensitiveTempDir = ".moai/run"`、`mkstemp` + `chmod 0o600` + `tmux source-file` + `os.Remove`
- `internal/tmux/errors.go` — `ErrTmuxSensitiveInjectFailed` sentinel
- `internal/hook/glm_tmux.go` — `ensureTmuxGLMEnv` 仅对 `ANTHROPIC_AUTH_TOKEN` 走 sensitive 路径; 其他 URL、模型名等 non-sensitive 值保留 argv 路径

### Non-sensitive 值保留 argv

`CLAUDE_CONFIG_DIR` (目录路径)、`ANTHROPIC_BASE_URL` (URL)、`ANTHROPIC_DEFAULT_*_MODEL` (模型名) 等非 token 值保留 argv 路径。此为显式设计,与 token 泄露风险无关。

### 失败时行为

若 source-file 注入失败 (磁盘满、`tmux source-file` 失败等),实现 **不会回退到 argv 导致泄露**,而是返回 `ErrTmuxSensitiveInjectFailed` sentinel 错误并中止注入。

### 自检

CG 模式运行期间验证 token 不暴露于 argv:

```bash
# moai cg 后在新 tmux 会话内
ps auxe | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# 期望值: 0 matches (token 不存在于 argv)
```

确认临时文件被正常 unlink:

```bash
ls -la ~/.moai/run/ 2>/dev/null
# 期望值: 空目录或无 stale 文件
```

若会话结束后 `~/.moai/run/` 仍有残余文件,可手动删除 (并非安全威胁 — 这些是已经尝试过 unlink 的文件)。

### 用户责任

`~/.moai/.env.glm` source 文件应保持 `0o600` 权限。`moai glm` 命令会自动设置:

```bash
stat -c '%a' ~/.moai/.env.glm    # Linux: 600
stat -f '%A' ~/.moai/.env.glm    # macOS: 600
```

详见 [CG 模式](/zh/multi-llm/cg-mode/)。

## CWE-345 — Update 流程强制 checksum 验证 {#cwe-345}

### 变更内容

`moai update` 自动更新流程 **无法绕过 checksum 验证**。如果 release 的 `checksums.txt` 下载失败或解析失败,系统返回 sentinel 错误 `ErrChecksumUnavailable` 并 **中止** 更新流程 — 不尝试下载 binary。

### Retry 策略

`checksums.txt` 下载以指数退避重试 **3 次**:

| 尝试 | 等待时间 |
|------|----------|
| 第 1 次 (立即) | 0s |
| 第 2 次 retry | 等待 2s |
| 第 3 次 retry | 等待 4s |
| 无更多 retry | 合计 约 6s 等待后失败 |

(内部实现: base delay 2s × 2^(attempt-1) 指数退避)

所有 retry 均失败时以 `ErrChecksumUnavailable` sentinel 结束。**不存在 `--skip-checksum` 类绕过选项**。

### Defense-in-depth

若 `version.Checksum` 字段为 empty string 状态时到达 `downloadAndVerify`,不会进行 binary 下载,而是返回 `ErrChecksumUnavailable`。双重防护 (checker 阶段 + updater 阶段) 阻断 silent bypass。

### 威胁模型

- **攻击者**: 网络 MITM (无法全部阻断但能选择性阻断/限速 `checksums.txt` URL)
- **攻击面**: 无 checksums.txt 时仍可安装 binary 的 silent fallback
- **结果**: 未经签名的后门 binary 无警告安装
- **CWE 映射**: CWE-345 (Insufficient Verification of Data Authenticity)

### 实现位置

- `internal/update/checker.go` — `downloadChecksumWithRetry(checksumsURL, archiveName, maxAttempts, baseDelay)` (`defaultChecksumMaxAttempts=3`、`defaultChecksumBaseDelay=2*time.Second`)、`ErrChecksumUnavailable` sentinel
- `internal/update/updater.go` — `downloadAndVerify` empty-checksum 保护
- 域名白名单 (`https://github.com/modu-ai/moai-adk/...`) 维持原样 (SSRF 表面无变化)

### 自检

```bash
# 验证 release 信息 + checksums.txt 可达
moai update --check-only

# 正常流程 (成功时)
moai update
# 输出示例: Downloaded checksums.txt (verified)

# checksums.txt 下载失败时 (示例: VPN 断开后执行)
moai update
# 输出示例: error: checksum unavailable: persistent retry failure after 3 attempts
```

如果出现 `ErrChecksumUnavailable` 信息:

1. 检查网络连通性 (`curl -I https://github.com/modu-ai/moai-adk/releases/latest`)
2. 确认 proxy / firewall 允许 GitHub release asset 域名
3. 考虑临时性 GitHub CDN 故障 — 稍后重试
4. **不提供 `--skip-checksum` 绕过选项** — 这是有意的策略

如长期阻塞,建议手动 binary 安装:

```bash
# 手动安装 (用户自行验证完整性)
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

详见 [更新](/zh/getting-started/update/)。

## 自检清单 (Self-Audit Checklist)

```bash
# 1. CWE-732 — settings.local.json 权限
stat -c '%a' .claude/settings.local.json 2>/dev/null \
  || stat -f '%A' .claude/settings.local.json 2>/dev/null
# 期望值: 600

# 2. CWE-214 — CG 模式运行期间 token argv 暴露
ps auxe 2>/dev/null | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# 期望值: 0 matches

# 3. CWE-214 — tmux sensitive temp 目录完整性
ls -la ~/.moai/run/ 2>/dev/null
# 期望值: 空目录或无 stale 文件

# 4. CWE-345 — Update 流程 checksum 行为
moai update --check-only
# 期望值: release + checksums.txt 正常确认

# 5. GLM source 文件权限 (用户责任)
stat -c '%a' ~/.moai/.env.glm 2>/dev/null \
  || stat -f '%A' ~/.moai/.env.glm 2>/dev/null
# 期望值: 600 (如文件存在)
```

上述 5 项均满足期望值时,v2.20.0-rc1 安全加固正常运行。

## References

### CHANGELOG

[CHANGELOG `[Unreleased]` v2.20.0-rc1 Security 部分](https://github.com/modu-ai/moai-adk/blob/main/CHANGELOG.md)

### SPEC

- `SPEC-V3R5-SECURITY-CRIT-001` — upstream source of truth, status `implemented` v0.2.0
- PR #1032 合并提交 `03a2552a2`

### Commits

- `b48bd86cb` — M1 settings.local.json 0o600 hardening (CWE-732/552)
- `10776c4b8` — M2 tmux sensitive env source-file injection (CWE-214)
- `ee1335282` — M3 mandatory checksum verification with retry (CWE-345)
- `b4e7115cb` — M4 cross-cutting verification + frontmatter

### CWE / OWASP

- [CWE-732](https://cwe.mitre.org/data/definitions/732.html) — Incorrect Permission Assignment for Critical Resource
- [CWE-552](https://cwe.mitre.org/data/definitions/552.html) — Files or Directories Accessible to External Parties
- [CWE-214](https://cwe.mitre.org/data/definitions/214.html) — Invocation of Process Using Visible Sensitive Information
- [CWE-345](https://cwe.mitre.org/data/definitions/345.html) — Insufficient Verification of Data Authenticity

### 相关页面

- [settings.json 指南](/zh/advanced/settings-json/) — `settings.local.json` 权限部分
- [更新](/zh/getting-started/update/) — checksum 验证部分
- [CG 模式](/zh/multi-llm/cg-mode/) — tmux 环境变量注入安全模型
