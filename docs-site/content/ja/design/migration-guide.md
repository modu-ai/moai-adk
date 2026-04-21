---
title: 移行ガイド
description: 既存の.agency/プロジェクトを新設計システムに変換
weight: 60
draft: false
---

# 移行ガイド

SPEC-AGENCY-ABSORB-001に従い、`/agency`コマンドが`/moai design`に統合されました。既存の`.agency/`ディレクトリを持つプロジェクトは、**マイグレーション**で新システムに移行できます。

## マイグレーション対象

以下のいずれかに該当する場合、マイグレーション必要:

- `.agency/`ディレクトリが存在
- 既存agencyの学習/観測を使用中
- 旧エージェント(agency-copywriter、agency-designer等)を活用中

**マイグレーション後:**

- `.agency/` → `.agency.archived/`(バックアップ)
- `.moai/project/brand/`(新規作成)
- `.moai/config/sections/design.yaml`(新規作成)
- 前learnings統合可能

## マイグレーション実行

### ステップ1: 事前確認

```bash
# .agency/存在確認
ls -la .agency/
# brand-voice.md
# visual-identity.md
# learnings/
# observations/
```

### ステップ2: ドライラン(選択)

マイグレーション結果を事前確認:

```bash
moai migrate agency --dry-run
```

### ステップ3: 実際のマイグレーション実行

```bash
moai migrate agency
```

**実行順序(6フェーズ):**

1. **検証** — `.agency/`存在確認、ディスク空き確認
2. **Staging** — 一時ディレクトリにコピー
3. **コンテキスト移行** — brandi ファイルを`.moai/project/brand/`にコピー
4. **設定マージ** — learnings/observationsを`.moai/config/`にマージ
5. **Learning移行** — 既存ヒューリスティックスを新構造に変換
6. **Atomic Swap** — `.agency.archived/`に元本バックアップ、完了

**完了時:**
```
マイグレーション完了 [TX-abc123def456]

移行ファイル: 47個
  ✓ .moai/project/brand/ 3個
  ✓ .moai/config/sections/design.yaml生成
  ✓ .moai/research/設定マージ

バックアップ位置: .agency.archived/

次ステップ:
  /moai design
```

## マイグレーションオプション

### --forceオプション

既存対象ディレクトリ上書き:

```bash
moai migrate agency --force
```

**注意:** `.moai/project/brand/`が既存なら上書き。事前バックアップ推奨。

### --resumeオプション

中断したマイグレーション再開:

```bash
# 前回SIGINT で中断した場合
moai migrate agency --resume TX-abc123def456
```

チェックポイントファイル: `~/.moai/.migrate-tx-<txID>.json`

## マイグレーションエラーコード

| エラーコード | 原因 | 解決方法 |
|---|---|---|
| `MIGRATE_NO_SOURCE` | `.agency/`なし | 既存agencyディレクトリ確認 |
| `MIGRATE_TARGET_EXISTS` | `.moai/project/brand/`既存 | `--force`オプション使用 |
| `MIGRATE_ARCHIVE_EXISTS` | `.agency.archived/`既存 | 既存バックアップ削除またはと移動 |
| `MIGRATE_DISK_FULL` | ディスク容量不足 | 空き容量確保(最少100MB) |
| `MIGRATE_MERGE_CONFLICT` | tech-preferences.md競合 | 既存`.moai/project/tech.md`バックアップ、再試行 |
| `MIGRATE_INTERRUPT` | SIGINT/SIGTERM受信 | `--resume`オプション で再開 |
| `MIGRATE_CHECKPOINT_CORRUPT` | チェックポイント破損 | `~/.moai/.migrate-tx-*.json`削除、再試行 |

## マイグレーション結果

### 生成ファイル構造

```
.moai/
├── project/
│   └── brand/
│       ├── brand-voice.md
│       ├── visual-identity.md
│       └── target-audience.md
├── config/
│   └── sections/
│       └── design.yaml
└── research/
    ├── learnings/
    └── observations/

.agency.archived/
├── brand-voice.md
├── visual-identity.md
├── learnings/
└── observations/
```

### Learning マージ

既存`.agency/learnings/`項目:
- 新構造に変換
- `.moai/research/learnings/`にマージ
- MIGRATED タグ付加

## ロールバック

マイグレーション後、前状態に戻すには:

### オプション1: バックアップから復元

```bash
# バックアップから復元
mv .agency.archived .agency

# 新規作成ファイル削除
rm -rf .moai/project/brand
rm .moai/config/sections/design.yaml
```

### オプション2: Gitで復元

マイグレーションがgit commitを生成した場合:

```bash
git log --oneline | grep migrate
# abc1234 chore: migrate agency to moai design system

git revert abc1234
```

## マイグレーション後

1. **ブランドコンテキスト確認**
   ```bash
   cat .moai/project/brand/brand-voice.md
   ```

2. **新設計ワークフロー開始**
   ```
   /moai design
   ```

3. **既存learnings確認**
   ```bash
   ls .moai/research/learnings/
   ```

4. **選択: 既存.agency/削除**
   ```bash
   rm -rf .agency.archived
   ```

## マイグレーション状態確認

マイグレーション結果サマリー照会:

```bash
# マイグレーションログ照会
cat ~/.moai/.migrate-tx-abc123def456.json

# またはCLIで確認
moai status design
```

## SIGINT/SIGTERM処理

マイグレーション中に中断:

**Ctrl+C入力時:**
```
マイグレーション中断 [TX-abc123def456]
完了フェーズ: validation, staging, context-transfer
未完了フェーズ: config-merge, learning-transfer, atomic-swap

再開:
  moai migrate agency --resume TX-abc123def456
```

## FAQ

### Q: マイグレーション後、既存/agencyコマンド使用可?

**A:** いいえ。`/agency`はサポート終了。`/moai design`を使用。

### Q: マイグレーション複数回実行可?

**A:** 初回完了後`.agency/`は`.agency.archived/`に変更され、2回目実行でsource not foundエラー。`--force`で上書き可。

### Q: Learningsが喪失されない?

**A:** いいえ。すべてlearnings/observations は`.moai/research/`にマージ、バックアップも`.agency.archived/`に保存。

### Q: マイグレーション中ネットワーク切断?

**A:** 完了フェーズは保存されるため、ネットワーク復旧後`--resume`で再開可。
