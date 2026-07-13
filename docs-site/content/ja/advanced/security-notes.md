---
title: セキュリティノート
description: "MoAI-ADK v2.20.0-rc1 のセキュリティ強化変更点 — CWE-732/214/345 マッピング、ユーザーセルフ監査手順"
weight: 72
draft: false
tags: ["security", "cwe", "audit"]
---

エージェンティック・ハーネスはエージェントに実行権限を渡すシステムです。権限を渡すシステムであるほど、クレデンシャルと更新経路のセキュリティがハーネスの信頼の土台をなします。本ページは MoAI-ADK v2.20.0-rc1 時点で導入された **ユーザーから見えるセキュリティ変更点** を整理します。各項目は CWE マッピング、変更された動作、セルフ点検コマンドを含みます。

## Why — なぜこのページが存在するのか

`SPEC-V3R5-SECURITY-CRIT-001` (PR #1032、merge commit `03a2552a2`) は v2.14.0 → v2.20.0-rc1 間のコードレビューで発見された **P0 release blocker セキュリティ欠陥 3 件** を是正しました。本ページはその是正の事実と、ユーザーが自分の環境で新しい保護が動作しているかを確認できる手順を、4-locale の公式案内として明文化します。

3 つの欠陥はすべて GLM 統合 + 自動更新経路に関連します。

- **CWE-732 / CWE-552** — `.claude/settings.local.json` ファイル mode `0o600` 強制 (所有者専用 read/write)
- **CWE-214** — `moai cg` の tmux 環境変数注入が argv ではなく source-file 経由 (GLM token argv 非可視化)
- **CWE-345** — `moai update` の checksum 検証 mandatory (ダウンロード失敗時は update を拒否)

各項目は回帰テストでロックされており、将来の回帰が遮断されます。

## CWE-732 — settings.local.json 権限強化 (Permission Hardening) {#cwe-732}

### 変更点

`.claude/settings.local.json` ファイルが生成・更新されるとき、ファイル権限が **`0o600`** (所有者のみ read/write) に強制されます。以前は `0o644` (所有者 read/write + group/world read) で生成されており、マルチユーザーワークステーションで他のローカルユーザーが `ANTHROPIC_AUTH_TOKEN` などの機密クレデンシャルを読める状態でした。

### 脅威モデル

- **攻撃者**: 同一ホストの低権限ローカルユーザー
- **攻撃面**: `.claude/settings.local.json` の group/world read 権限
- **漏洩情報**: GLM API token (`ANTHROPIC_AUTH_TOKEN`)、OAuth refresh token、その他の `settings.Env` 値
- **CWE マッピング**: CWE-732 (Incorrect Permission Assignment for Critical Resource)、CWE-552 (Files or Directories Accessible to External Parties)

### 実装場所

- `internal/hook/settings_io.go` — `secureSettingsMode os.FileMode = 0o600` 定数 + `writeSettingsSecure` ヘルパー
- `internal/hook/session_start.go` — `ensureGLMCredentials`, `ensureClaudeEnvFile` などすべての `settings.local.json` writer
- `internal/hook/session_end.go` — GLM keys write-back 経路

### セルフ点検

既存の `settings.local.json` 権限を確認します。

```bash
# Linux
stat -c '%a' .claude/settings.local.json
# 期待値: 600

# macOS
stat -f '%A' .claude/settings.local.json
# 期待値: 600
```

権限が `644` などより緩い値で表示された場合、MoAI-ADK が次のセッション開始時に自動で `0o600` に是正します。すぐに是正するには:

```bash
chmod 0600 .claude/settings.local.json
```

### 影響 (Trade-off)

`group-readable` を期待するワークフロー (同一プロジェクトディレクトリを別の OS ユーザーが read する非常にまれなシナリオ) は壊れる可能性があります。このトレードオフは意図されたものであり、セキュリティ回復が明確な優先事項です。

## CWE-214 — tmux IPC token argv 露出の遮断 {#cwe-214}

### 変更点

`moai cg` (CG モード) が GLM token (`ANTHROPIC_AUTH_TOKEN`) を tmux セッション環境変数に注入するとき、**argv チャネル** (`tmux set-environment <KEY> <VALUE>`) の代わりに **source-file チャネル** (`tmux source-file <tmp>`) を使用します。token はもはや `ps auxe`、`/proc/<pid>/cmdline`、auditd ログ、sysmon トレース、クラッシュダンプに平文で露出しません。

CG モードはトークノミクスの主要な削減手段 (Claude リーダー + GLM ワーカー、60-70% 削減) であるだけに、そのクレデンシャル経路のセキュリティは特に重要です。

### 実装フロー

1. `~/.moai/run/` 配下に一時ファイルを `mkstemp` で作成 (mode `0o600` 自動 + 明示的 `chmod 0o600`)
2. `set-environment -t <session> <KEY> <VALUE>` の 1 行を一時ファイルに記録
3. `tmux source-file <tmp>` で tmux がそのファイルを読んで環境に注入
4. 注入直後に一時ファイルを `os.Remove` で unlink

argv には一時ファイルのパスだけが露出し、token 自体は露出しません。

### 脅威モデル

- **攻撃者**: 同一ホストのローカルユーザー + システムログ収集 (`ps`, `/proc`, auditd, sysmon)
- **攻撃面**: tmux env injection の argv チャネル
- **漏洩情報**: GLM API token の瞬間的な可視化
- **CWE マッピング**: CWE-214 (Invocation of Process Using Visible Sensitive Information)

### 実装場所

- `internal/tmux/session.go` — `InjectSensitiveEnv` メソッド、`sensitiveTempDir = ".moai/run"`、`mkstemp` + `chmod 0o600` + `tmux source-file` + `os.Remove`
- `internal/tmux/errors.go` — `ErrTmuxSensitiveInjectFailed` sentinel
- `internal/hook/glm_tmux.go` — `ensureTmuxGLMEnv` で `ANTHROPIC_AUTH_TOKEN` のみ sensitive 経路に分岐 (それ以外の URL、model 名など non-sensitive 値は既存の argv 経路を維持)

### Non-sensitive 値は argv を維持

`CLAUDE_CONFIG_DIR` (ディレクトリパス)、`ANTHROPIC_BASE_URL` (URL)、`ANTHROPIC_DEFAULT_*_MODEL` (モデル名) など token ではない値は argv 経路を維持します。これは明示的な意図であり、トークン漏洩リスクとは無関係です。

### 失敗時の動作

source-file 注入が失敗した場合 (ディスクフル、tmux source-file 失敗など)、**argv fallback で漏洩することなく** `ErrTmuxSensitiveInjectFailed` sentinel error を返して注入自体を abort します。失敗時に利便性へ後退しないことがこの設計の核心です。

### セルフ点検

CG モード実行中に token が argv に露出していないか確認します。

```bash
# moai cg 実行後、新しい tmux セッション内で
ps auxe | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# 期待値: 0 matches (token が argv にない)
```

一時ファイルが正常に unlink されているか確認します。

```bash
ls -la ~/.moai/run/ 2>/dev/null
# 期待値: 空ディレクトリまたは stale ファイルなし
```

セッション終了後に `~/.moai/run/` に残存ファイルがある場合は手動で削除できます (セキュリティ脅威ではありません — すでに unlink が試行されたファイル)。

### ユーザーの責任

`~/.moai/.env.glm` source ファイルはユーザー環境で `0o600` 権限を維持する必要があります。これは `moai glm` コマンドが自動で設定します。

```bash
stat -c '%a' ~/.moai/.env.glm    # Linux: 600
stat -f '%A' ~/.moai/.env.glm    # macOS: 600
```

詳細: [CG モード](/ja/multi-llm/cg-mode/)

## CWE-345 — Update フローの mandatory checksum 検証 {#cwe-345}

### 変更点

`moai update` の自動更新フローは **checksum 検証を迂回できません**。release の `checksums.txt` ダウンロードが失敗、またはパースが失敗すると sentinel error `ErrChecksumUnavailable` を返して update フローを **abort** します — binary のダウンロードを試みません。

### Retry ポリシー

`checksums.txt` ダウンロードは **3 回 retry** を指数バックオフで試行します。

| 試行 | 待機時間 |
|------|-----------|
| 1 回目 (即時) | 0s |
| 2 回目 retry | 2s 待機 |
| 3 回目 retry | 4s 待機 |
| 追加 retry なし | 合計 ~6s 待機後に失敗 |

(内部実装: base delay 2s × 2^(attempt-1) の指数バックオフ)

すべての retry が失敗すると `ErrChecksumUnavailable` sentinel で終了します。**`--skip-checksum` のような迂回オプションは存在しません**。

### Defense-in-depth

`version.Checksum` フィールドが empty string の状態で `downloadAndVerify` に到達した場合、binary ダウンロードを進めず `ErrChecksumUnavailable` を返します。二重保護 (checker 段階 + updater 段階) でサイレントな迂回を遮断します。

### 脅威モデル

- **攻撃者**: ネットワーク MITM (全体遮断はできなくても `checksums.txt` URL だけ選択遮断・throttle 可能)
- **攻撃面**: checksums.txt なしでも binary がインストールされていた silent fallback
- **漏洩結果**: 署名されていないバックドアバイナリの無警告インストール
- **CWE マッピング**: CWE-345 (Insufficient Verification of Data Authenticity)

### 実装場所

- `internal/update/checker.go` — `downloadChecksumWithRetry(checksumsURL, archiveName, maxAttempts, baseDelay)` (`defaultChecksumMaxAttempts=3`, `defaultChecksumBaseDelay=2*time.Second`)、`ErrChecksumUnavailable` sentinel
- `internal/update/updater.go` — `downloadAndVerify` の empty-checksum guard
- domain whitelist (`https://github.com/modu-ai/moai-adk/...`) は従来どおり維持 (SSRF 表面の変化なし)

### セルフ点検

```bash
# release 情報 + checksums.txt の存在確認
moai update --check-only

# 正常フロー (成功時)
moai update
# 出力例: Downloaded checksums.txt (verified)

# checksums.txt ダウンロード失敗時 (意図的な遮断例: VPN 切断後に実行)
moai update
# 出力例: error: checksum unavailable: persistent retry failure after 3 attempts
```

`ErrChecksumUnavailable` メッセージが表示された場合は次を確認してください。

1. ネットワーク接続確認 (`curl -I https://github.com/modu-ai/moai-adk/releases/latest`)
2. Proxy / firewall が GitHub release asset ドメインを許可しているか確認
3. 一時的な GitHub CDN 障害の可能性 — しばらくして再試行
4. **`--skip-checksum` のような迂回オプションは提供されません** — これは意図されたポリシー

恒久的に遮断される場合は手動 binary インストールを推奨します。

```bash
# 手動インストール (ユーザーが直接整合性を検証)
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
```

詳細: [アップデート](/ja/getting-started/update/)

## セルフ監査チェックリスト (Self-Audit Checklist)

5 項目を一度に点検できます。

```bash
# 1. CWE-732 — settings.local.json 権限
stat -c '%a' .claude/settings.local.json 2>/dev/null \
  || stat -f '%A' .claude/settings.local.json 2>/dev/null
# 期待値: 600

# 2. CWE-214 — CG モード実行中の token argv 露出 (cg モードアクティブ状態で)
ps auxe 2>/dev/null | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# 期待値: 0 matches

# 3. CWE-214 — tmux sensitive temp ディレクトリの整合性
ls -la ~/.moai/run/ 2>/dev/null
# 期待値: 空ディレクトリまたは stale ファイルなし

# 4. CWE-345 — Update flow checksum 動作
moai update --check-only
# 期待値: release + checksums.txt の正常確認

# 5. GLM source ファイル権限 (ユーザーの責任)
stat -c '%a' ~/.moai/.env.glm 2>/dev/null \
  || stat -f '%A' ~/.moai/.env.glm 2>/dev/null
# 期待値: 600 (該当ファイルが存在する場合)
```

上記 5 項目がすべて期待値を満たしていれば、v2.20.0-rc1 のセキュリティ強化が正常に動作しています。

## References

### CHANGELOG

[CHANGELOG `[Unreleased]` v2.20.0-rc1 Security セクション](https://github.com/modu-ai/moai-adk/blob/main/CHANGELOG.md)

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

### 関連ページ

- [settings.json ガイド](/ja/advanced/settings-json/) — `settings.local.json` 権限セクション
- [アップデート](/ja/getting-started/update/) — checksum 検証セクション
- [CG モード](/ja/multi-llm/cg-mode/) — tmux 環境変数注入のセキュリティモデル
