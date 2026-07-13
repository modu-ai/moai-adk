---
title: 安全说明
description: "MoAI-ADK v2.20.0-rc1 安全加固变更 — CWE-732/214/345 映射、用户自检流程"
weight: 72
draft: false
tags: ["security", "cwe", "audit"]
---

智能体 Harness 是把执行权限交给智能体的系统。越是移交权限的系统，凭据与更新路径的安全就越构成 Harness 信任的底座。本页整理 MoAI-ADK v2.20.0-rc1 时点引入的 **用户可见安全变更**。每一项都包含 CWE 映射、变更后的行为以及自检命令。

## Why — 本页为何存在

`SPEC-V3R5-SECURITY-CRIT-001`（PR #1032，merge commit `03a2552a2`）修正了 v2.14.0 → v2.20.0-rc1 之间代码评审中发现的 **3 个 P0 release blocker 安全缺陷**。本页把该修正事实以及用户可以在自己环境中确认新保护是否生效的流程，成文为 4 语言官方指南。

三个缺陷都与 GLM 集成 + 自动更新路径相关。

- **CWE-732 / CWE-552** — `.claude/settings.local.json` 文件 mode 强制 `0o600`（仅所有者可读写）
- **CWE-214** — `moai cg` 的 tmux 环境变量注入改经 source-file 而非 argv（GLM token 在 argv 中不可见）
- **CWE-345** — `moai update` 的 checksum 校验为强制（下载失败时拒绝更新）

每一项都由回归测试锁定，阻断未来回归。

## CWE-732 — settings.local.json 权限加固 (Permission Hardening) {#cwe-732}

### 变更内容

`.claude/settings.local.json` 文件在创建、更新时权限被强制设为 **`0o600`**（仅所有者可读写）。此前以 `0o644`（所有者读写 + group/world 可读）创建，在多用户工作站上其他本地用户可以读取 `ANTHROPIC_AUTH_TOKEN` 等敏感凭据。

### 威胁模型

- **攻击者**：同一主机上的低权限本地用户
- **攻击面**：`.claude/settings.local.json` 的 group/world 读权限
- **泄露信息**：GLM API token (`ANTHROPIC_AUTH_TOKEN`)、OAuth refresh token、其他 `settings.Env` 值
- **CWE 映射**：CWE-732 (Incorrect Permission Assignment for Critical Resource)、CWE-552 (Files or Directories Accessible to External Parties)

### 实现位置

- `internal/hook/settings_io.go` — `secureSettingsMode os.FileMode = 0o600` 常量 + `writeSettingsSecure` 辅助函数
- `internal/hook/session_start.go` — `ensureGLMCredentials`、`ensureClaudeEnvFile` 等所有 `settings.local.json` writer
- `internal/hook/session_end.go` — GLM keys write-back 路径

### 自检

确认既有 `settings.local.json` 的权限。

```bash
# Linux
stat -c '%a' .claude/settings.local.json
# 期望值: 600

# macOS
stat -f '%A' .claude/settings.local.json
# 期望值: 600
```

如果权限显示为 `644` 或其他更宽松的值，MoAI-ADK 会在下次会话启动时自动修正为 `0o600`。若要立即修正：

```bash
chmod 0600 .claude/settings.local.json
```

### 影响 (Trade-off)

依赖 `group-readable` 的工作流（由不同 OS 用户读取同一项目目录的极罕见场景）可能会被破坏。这一权衡是有意为之的，安全恢复显然优先。

## CWE-214 — 阻断 tmux IPC token 的 argv 暴露 {#cwe-214}

### 变更内容

`moai cg`（CG 模式）向 tmux 会话环境变量注入 GLM token (`ANTHROPIC_AUTH_TOKEN`) 时，使用 **source-file 通道**（`tmux source-file <tmp>`）取代 **argv 通道**（`tmux set-environment <KEY> <VALUE>`）。token 不再以明文暴露于 `ps auxe`、`/proc/<pid>/cmdline`、auditd 日志、sysmon 追踪与崩溃转储。

CG 模式是代币经济学的核心节省手段（Claude 领队 + GLM 工作者，节省 60-70%），因此其凭据路径的安全尤为重要。

### 实现流程

1. 在 `~/.moai/run/` 下用 `mkstemp` 创建临时文件（自动 mode `0o600` + 显式 `chmod 0o600`）
2. 把 `set-environment -t <session> <KEY> <VALUE>` 这一行写入临时文件
3. 通过 `tmux source-file <tmp>` 让 tmux 读取该文件并注入环境
4. 注入后立即用 `os.Remove` unlink 临时文件

argv 中只暴露临时文件路径，token 本身不暴露。

### 威胁模型

- **攻击者**：同一主机上的本地用户 + 系统日志采集（`ps`、`/proc`、auditd、sysmon）
- **攻击面**：tmux env 注入的 argv 通道
- **泄露信息**：GLM API token 的瞬时可见
- **CWE 映射**：CWE-214 (Invocation of Process Using Visible Sensitive Information)

### 实现位置

- `internal/tmux/session.go` — `InjectSensitiveEnv` 方法、`sensitiveTempDir = ".moai/run"`、`mkstemp` + `chmod 0o600` + `tmux source-file` + `os.Remove`
- `internal/tmux/errors.go` — `ErrTmuxSensitiveInjectFailed` sentinel
- `internal/hook/glm_tmux.go` — 在 `ensureTmuxGLMEnv` 中仅将 `ANTHROPIC_AUTH_TOKEN` 分流到 sensitive 路径（URL、模型名等其余 non-sensitive 值维持既有 argv 路径）

### Non-sensitive 值维持 argv

`CLAUDE_CONFIG_DIR`（目录路径）、`ANTHROPIC_BASE_URL`（URL）、`ANTHROPIC_DEFAULT_*_MODEL`（模型名）等非 token 的值维持 argv 路径。这是显式意图，与 token 泄露风险无关。

### 失败时的行为

若 source-file 注入失败（磁盘写满、tmux source-file 失败等），**不会回退到 argv 造成泄露**，而是返回 `ErrTmuxSensitiveInjectFailed` sentinel error 并中止注入本身。失败时不为了便利而回退 — 这正是此设计的核心。

### 自检

确认 CG 模式运行中 token 是否暴露在 argv 中。

```bash
# 运行 moai cg 后在新 tmux 会话内
ps auxe | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# 期望值: 0 matches (token 不在 argv 中)
```

确认临时文件被正常 unlink。

```bash
ls -la ~/.moai/run/ 2>/dev/null
# 期望值: 空目录或无 stale 文件
```

如果会话结束后 `~/.moai/run/` 有残留文件，可以手动删除（不构成安全威胁 — 是已尝试 unlink 的文件）。

### 用户责任

`~/.moai/.env.glm` source 文件须在用户环境中保持 `0o600` 权限。`moai glm` 命令会自动设置。

```bash
stat -c '%a' ~/.moai/.env.glm    # Linux: 600
stat -f '%A' ~/.moai/.env.glm    # macOS: 600
```

详情：[CG 模式](/zh/multi-llm/cg-mode/)

## CWE-345 — 更新流程的强制 checksum 校验 {#cwe-345}

### 变更内容

`moai update` 的自动更新流程 **无法绕过 checksum 校验**。当 release 的 `checksums.txt` 下载失败或解析失败时，返回 sentinel error `ErrChecksumUnavailable` 并 **中止** 更新流程 — 不会尝试下载二进制。

### Retry 策略

`checksums.txt` 下载以指数退避尝试 **3 次 retry**。

| 尝试 | 等待时间 |
|------|-----------|
| 第 1 次（立即） | 0s |
| 第 2 次 retry | 等待 2s |
| 第 3 次 retry | 等待 4s |
| 无更多 retry | 合计等待 ~6s 后失败 |

（内部实现：base delay 2s × 2^(attempt-1) 指数退避）

所有 retry 失败后以 `ErrChecksumUnavailable` sentinel 结束。**不存在 `--skip-checksum` 之类的绕过选项**。

### Defense-in-depth

若 `version.Checksum` 字段以空字符串状态到达 `downloadAndVerify`，则不进行二进制下载，直接返回 `ErrChecksumUnavailable`。以双重保护（checker 阶段 + updater 阶段）阻断静默绕过。

### 威胁模型

- **攻击者**：网络 MITM（无法全面阻断，但可选择性阻断、限速 `checksums.txt` URL）
- **攻击面**：没有 checksums.txt 也能安装二进制的 silent fallback
- **泄露后果**：无警告地安装未签名的后门二进制
- **CWE 映射**：CWE-345 (Insufficient Verification of Data Authenticity)

### 实现位置

- `internal/update/checker.go` — `downloadChecksumWithRetry(checksumsURL, archiveName, maxAttempts, baseDelay)`（`defaultChecksumMaxAttempts=3`、`defaultChecksumBaseDelay=2*time.Second`）、`ErrChecksumUnavailable` sentinel
- `internal/update/updater.go` — `downloadAndVerify` empty-checksum guard
- domain whitelist（`https://github.com/modu-ai/moai-adk/...`）原样保留（SSRF 面无变化）

### 自检

```bash
# 确认 release 信息 + checksums.txt 存在
moai update --check-only

# 正常流程 (成功时)
moai update
# 输出示例: Downloaded checksums.txt (verified)

# checksums.txt 下载失败时 (有意阻断示例: 断开 VPN 后执行)
moai update
# 输出示例: error: checksum unavailable: persistent retry failure after 3 attempts
```

若显示 `ErrChecksumUnavailable` 消息，请确认以下事项。

1. 确认网络连接（`curl -I https://github.com/modu-ai/moai-adk/releases/latest`）
2. 确认 Proxy / firewall 是否放行 GitHub release asset 域名
3. GitHub CDN 可能出现暂时性故障 — 稍后重试
4. **不提供 `--skip-checksum` 之类的绕过选项** — 这是有意的策略

若被永久阻断，建议手动安装二进制。

```bash
# 手动安装 (由用户自行校验完整性)
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

详情：[更新](/zh/getting-started/update/)

## 自检清单 (Self-Audit Checklist)

可以一次性检查五个项目。

```bash
# 1. CWE-732 — settings.local.json 权限
stat -c '%a' .claude/settings.local.json 2>/dev/null \
  || stat -f '%A' .claude/settings.local.json 2>/dev/null
# 期望值: 600

# 2. CWE-214 — CG 模式运行中 token argv 暴露 (cg 模式激活状态下)
ps auxe 2>/dev/null | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# 期望值: 0 matches

# 3. CWE-214 — tmux sensitive temp 目录一致性
ls -la ~/.moai/run/ 2>/dev/null
# 期望值: 空目录或无 stale 文件

# 4. CWE-345 — 更新流程 checksum 行为
moai update --check-only
# 期望值: release + checksums.txt 正常确认

# 5. GLM source 文件权限 (用户责任)
stat -c '%a' ~/.moai/.env.glm 2>/dev/null \
  || stat -f '%A' ~/.moai/.env.glm 2>/dev/null
# 期望值: 600 (若该文件存在)
```

以上 5 项均满足期望值，即表示 v2.20.0-rc1 安全加固正常生效。

## References

### CHANGELOG

[CHANGELOG `[Unreleased]` v2.20.0-rc1 Security 章节](https://github.com/modu-ai/moai-adk/blob/main/CHANGELOG.md)

### SPEC

- `SPEC-V3R5-SECURITY-CRIT-001` — upstream source of truth, status `implemented` v0.2.0
- PR #1032 merge commit `03a2552a2`

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

- [settings.json 指南](/zh/advanced/settings-json/) — `settings.local.json` 权限章节
- [更新](/zh/getting-started/update/) — checksum 校验章节
- [CG 模式](/zh/multi-llm/cg-mode/) — tmux 环境变量注入安全模型
