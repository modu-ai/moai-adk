---
title: セキュリティノート
description: "MoAI-ADK v2.20.0-rc1 のセキュリティ強化 — CWE-732/214/345 マッピングと利用者向けセルフ監査手順"
weight: 72
draft: false
tags: ["security", "cwe", "audit"]
---

# セキュリティノート (Security Notes)

本ページは MoAI-ADK v2.20.0-rc1 で導入された **ユーザーに見えるセキュリティ変更** を整理します。各項目には CWE マッピング、変更された動作、セルフ監査コマンドを記載します。

## このページが存在する理由 (Why)

`SPEC-V3R5-SECURITY-CRIT-001` (PR #1032、マージコミット `03a2552a2`) は、v2.14.0 → v2.20.0-rc1 のコードレビューで発見された **P0 リリースブロッカーのセキュリティ欠陥 3 件** を修正しました。本ページはその修正事実と、新しい保護がユーザー環境で正しく動作しているかを確認できる手順を、公式 4-locale ドキュメントとして明文化します。

3 件はいずれも GLM 統合および自動アップデート経路に関連しています:

- **CWE-732 / CWE-552** — `.claude/settings.local.json` のファイルモードを **`0o600`** (所有者のみ read/write) に強制。
- **CWE-214** — `moai cg` の tmux 環境変数注入を argv ではなく source-file 経由に変更 (GLM トークンを argv から不可視化)。
- **CWE-345** — `moai update` の checksum 検証を mandatory 化 (ダウンロード失敗時に update を拒否)。

各修正は回帰テストでロックされており、将来の回帰を阻止します。

## CWE-732 — settings.local.json 権限強化 (Permission Hardening) {#cwe-732}

### 変更内容

`.claude/settings.local.json` の生成・更新時に、ファイル権限を **`0o600`** (所有者のみ read/write) に強制します。以前は `0o644` (所有者 read/write + group/world read) で作成されていたため、複数ユーザー環境のワークステーションで別のローカルユーザーが `ANTHROPIC_AUTH_TOKEN` などの機密資格情報を読み取れる可能性がありました。

### 脅威モデル

- **攻撃者**: 同一ホストの低権限ローカルユーザー
- **攻撃表面**: `.claude/settings.local.json` の group/world read 権限
- **漏洩情報**: GLM API トークン (`ANTHROPIC_AUTH_TOKEN`)、OAuth refresh トークン、その他 `settings.Env` 値
- **CWE マッピング**: CWE-732 (Incorrect Permission Assignment for Critical Resource)、CWE-552 (Files or Directories Accessible to External Parties)

### 実装位置

- `internal/hook/settings_io.go` — 定数 `secureSettingsMode os.FileMode = 0o600` + `writeSettingsSecure` ヘルパー
- `internal/hook/session_start.go` — `settings.local.json` 書き込み経路すべて (`ensureGLMCredentials`、`ensureClaudeEnvFile`)
- `internal/hook/session_end.go` — GLM keys write-back パス

### セルフ監査

既存の `settings.local.json` 権限を確認:

```bash
# Linux
stat -c '%a' .claude/settings.local.json
# 期待値: 600

# macOS
stat -f '%A' .claude/settings.local.json
# 期待値: 600
```

権限が `644` またはそれ以上に緩い場合、MoAI-ADK は次回セッション開始時に自動で `0o600` に修正します。即時修正するには:

```bash
chmod 0600 .claude/settings.local.json
```

### 影響 (Trade-off)

`group-readable` を期待するワークフロー (同じプロジェクトディレクトリを別の OS ユーザーが read する非常にまれなシナリオ) は破壊されます。このトレードオフは意図的であり、セキュリティ回復が明確に優先されます。

## CWE-214 — tmux IPC トークン argv 露出の遮断 {#cwe-214}

### 変更内容

`moai cg` (CG モード) が GLM トークン (`ANTHROPIC_AUTH_TOKEN`) を tmux セッション環境変数に注入する際、**argv チャンネル** (`tmux set-environment <KEY> <VALUE>`) ではなく **source-file チャンネル** (`tmux source-file <tmp>`) を使用します。トークンは `ps auxe`、`/proc/<pid>/cmdline`、auditd ログ、sysmon トレース、クラッシュダンプに平文で露出しなくなりました。

### 実装フロー

1. `~/.moai/run/` 配下に `mkstemp` で一時ファイルを作成 (デフォルト mode `0o600` + 明示的 `chmod 0o600`)。
2. `set-environment -t <session> <KEY> <VALUE>` の 1 行を一時ファイルに書き込み。
3. `tmux source-file <tmp>` を呼び出し、tmux にそのファイルを読み込ませて環境に注入。
4. 注入直後に `os.Remove` で一時ファイルを unlink。

argv に露出するのは一時ファイルのパスのみで、トークン自体は決して argv に現れません。

### 脅威モデル

- **攻撃者**: 同一ホストのローカルユーザー + システムログコレクター (`ps`、`/proc`、auditd、sysmon)
- **攻撃表面**: tmux 環境変数注入の argv チャンネル
- **漏洩情報**: GLM API トークンの瞬間的露出
- **CWE マッピング**: CWE-214 (Invocation of Process Using Visible Sensitive Information)

### 実装位置

- `internal/tmux/session.go` — `InjectSensitiveEnv` メソッド、`sensitiveTempDir = ".moai/run"`、`mkstemp` + `chmod 0o600` + `tmux source-file` + `os.Remove`
- `internal/tmux/errors.go` — `ErrTmuxSensitiveInjectFailed` sentinel
- `internal/hook/glm_tmux.go` — `ensureTmuxGLMEnv` で `ANTHROPIC_AUTH_TOKEN` のみ sensitive パスに分岐 (それ以外の URL、モデル名などの non-sensitive 値は argv 経路を維持)

### Non-sensitive 値は argv を維持

`CLAUDE_CONFIG_DIR` (ディレクトリパス)、`ANTHROPIC_BASE_URL` (URL)、`ANTHROPIC_DEFAULT_*_MODEL` (モデル名) などトークンでない値は argv 経路を維持します。これは明示的な意図であり、トークン漏洩リスクとは無関係です。

### 失敗時の動作

source-file 注入が失敗した場合 (ディスク満杯、`tmux source-file` 失敗など) は **argv にフォールバックして漏洩させず**、`ErrTmuxSensitiveInjectFailed` sentinel エラーを返して注入自体を abort します。

### セルフ監査

CG モード実行中にトークンが argv に露出していないか確認:

```bash
# moai cg 実行後、新しい tmux セッション内で
ps auxe | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# 期待値: 0 matches (トークンが argv に存在しない)
```

一時ファイルが正常に unlink されているか確認:

```bash
ls -la ~/.moai/run/ 2>/dev/null
# 期待値: 空のディレクトリ、または stale ファイルなし
```

セッション終了後に `~/.moai/run/` に残存ファイルがある場合は手動で削除可能です (セキュリティ脅威ではなく、既に unlink が試みられたファイルです)。

### ユーザー責任

`~/.moai/.env.glm` source ファイルは `0o600` 権限を維持する必要があります。これは `moai glm` コマンドが自動的に設定します:

```bash
stat -c '%a' ~/.moai/.env.glm    # Linux: 600
stat -f '%A' ~/.moai/.env.glm    # macOS: 600
```

詳細は [CG モード](/ja/multi-llm/cg-mode/) を参照してください。

## CWE-345 — Update フローの mandatory checksum 検証 {#cwe-345}

### 変更内容

`moai update` の自動アップデートフローは **checksum 検証を回避できません**。リリースの `checksums.txt` ダウンロードが失敗、またはパースが失敗した場合、sentinel エラー `ErrChecksumUnavailable` を返してアップデートフローを **abort** します — binary ダウンロードは試行しません。

### Retry ポリシー

`checksums.txt` ダウンロードは指数バックオフで **3 回 retry** します:

| 試行 | 待機時間 |
|------|----------|
| 1 回目 (即時) | 0s |
| 2 回目 retry | 2s 待機 |
| 3 回目 retry | 4s 待機 |
| 追加 retry なし | 合計 約 6s 待機後に失敗 |

(内部実装: base delay 2s × 2^(attempt-1) 指数バックオフ)

すべての retry が失敗すると `ErrChecksumUnavailable` sentinel で終了します。**`--skip-checksum` のような回避オプションは存在しません**。

### Defense-in-depth

`version.Checksum` フィールドが empty string の状態で `downloadAndVerify` に到達した場合、binary ダウンロードを進めず `ErrChecksumUnavailable` を返します。二重保護 (checker 段階 + updater 段階) で silent bypass を阻止します。

### 脅威モデル

- **攻撃者**: ネットワーク MITM (全体は遮断できないが `checksums.txt` URL のみ選択的に遮断・throttle 可能)
- **攻撃表面**: checksums.txt なしで binary がインストールされていた silent fallback
- **結果**: 署名されていないバックドアバイナリの無警告インストール
- **CWE マッピング**: CWE-345 (Insufficient Verification of Data Authenticity)

### 実装位置

- `internal/update/checker.go` — `downloadChecksumWithRetry(checksumsURL, archiveName, maxAttempts, baseDelay)` (`defaultChecksumMaxAttempts=3`、`defaultChecksumBaseDelay=2*time.Second`)、`ErrChecksumUnavailable` sentinel
- `internal/update/updater.go` — `downloadAndVerify` empty-checksum ガード
- ドメインホワイトリスト (`https://github.com/modu-ai/moai-adk/...`) は既存どおり保持 (SSRF 表面変化なし)

### セルフ監査

```bash
# リリース情報 + checksums.txt の存在確認
moai update --check-only

# 正常フロー (成功時)
moai update
# 出力例: Downloaded checksums.txt (verified)

# checksums.txt ダウンロード失敗時 (意図的遮断例: VPN 切断後の実行)
moai update
# 出力例: error: checksum unavailable: persistent retry failure after 3 attempts
```

`ErrChecksumUnavailable` メッセージが表示された場合:

1. ネットワーク接続を確認 (`curl -I https://github.com/modu-ai/moai-adk/releases/latest`)
2. Proxy / firewall が GitHub release asset ドメインを許可しているか確認
3. 一時的な GitHub CDN 障害の可能性 — しばらく後に再試行
4. **`--skip-checksum` のような回避オプションは提供されません** — これは意図された方針

恒久的に遮断される場合は手動 binary インストールを推奨します:

```bash
# 手動インストール (ユーザー自身が無欠性を検証)
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

詳細は [アップデート](/ja/getting-started/update/) を参照してください。

## セルフ監査チェックリスト (Self-Audit Checklist)

```bash
# 1. CWE-732 — settings.local.json 権限
stat -c '%a' .claude/settings.local.json 2>/dev/null \
  || stat -f '%A' .claude/settings.local.json 2>/dev/null
# 期待値: 600

# 2. CWE-214 — CG モード実行中のトークン argv 露出
ps auxe 2>/dev/null | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# 期待値: 0 matches

# 3. CWE-214 — tmux sensitive temp ディレクトリ整合性
ls -la ~/.moai/run/ 2>/dev/null
# 期待値: 空のディレクトリ、または stale ファイルなし

# 4. CWE-345 — Update フロー checksum 動作
moai update --check-only
# 期待値: release + checksums.txt 正常確認

# 5. GLM source ファイル権限 (ユーザー責任)
stat -c '%a' ~/.moai/.env.glm 2>/dev/null \
  || stat -f '%A' ~/.moai/.env.glm 2>/dev/null
# 期待値: 600 (ファイルが存在する場合)
```

上記 5 項目すべてが期待値を満たす場合、v2.20.0-rc1 のセキュリティ強化は正常に動作しています。

## References

### CHANGELOG

[CHANGELOG `[Unreleased]` v2.20.0-rc1 Security セクション](https://github.com/modu-ai/moai-adk/blob/main/CHANGELOG.md)

### SPEC

- `SPEC-V3R5-SECURITY-CRIT-001` — upstream source of truth、status `implemented` v0.2.0
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

### 関連ページ

- [settings.json ガイド](/ja/advanced/settings-json/) — `settings.local.json` 権限セクション
- [アップデート](/ja/getting-started/update/) — checksum 検証セクション
- [CG モード](/ja/multi-llm/cg-mode/) — tmux 環境変数注入のセキュリティモデル
