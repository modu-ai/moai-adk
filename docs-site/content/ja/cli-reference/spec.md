---
title: moai spec ドキュメント管理
weight: 35
draft: false
---

`moai spec` は `.moai/specs/` ディレクトリの SPEC ドキュメントを管理します。ステータス更新、ドリフト検出、受け入れ基準の照会、EARS/GEARS リント、原子的クローズ、era 監査、アーカイブのサブコマンドを提供します。

## サブコマンド

| コマンド | 説明 |
|--------|------|
| `moai spec status` | SPEC ステータスの更新または一覧表示 |
| `moai spec drift` | frontmatter status と git log 間のドリフト検出 |
| `moai spec view <SPEC-ID>` | 受け入れ基準をツリー構造で照会 |
| `moai spec lint [spec.md...]` | EARS 準拠と構造妥当性のリント |
| `moai spec close <SPEC-ID>` | 原子的 4-phase クローズ (status: completed + progress.md backfill) |
| `moai spec audit` | SPEC era 分類と modern-era ステータスドリフト監査 |
| `moai spec archive` | クローズ済み SPEC を `.moai/specs/` の外へアーカイブ |

## moai spec status

```bash
moai spec status <SPEC-ID> <new-status>   # ステータス更新
moai spec status --list                   # 全 SPEC を一覧表示
moai spec status --sync-git               # git log からステータス同期
```

| フラグ | 説明 |
|--------|------|
| `--dry-run` | 書き込みなしで変更をプレビュー |
| `--list` | 全 SPEC とステータスを一覧表示 |
| `--sync-git` | main の git log から SPEC ステータスを同期 |
| `--yes` | `--sync-git` の非対話自動確認 (CI/パイプ必須) |

## moai spec drift

```bash
moai spec drift
```

| フラグ | 説明 |
|--------|------|
| `--json` | JSON 形式で出力 |
| `--exit-code-on-drift` | ドリフト検出時に終了コード 1 |
| `--count` | ドリフト件数のみ出力 |
| `--no-cache` | HEAD-SHA 結果キャッシュを迂回して再計算 |

## moai spec lint

```bash
moai spec lint [spec.md...]
```

| フラグ | 説明 |
|--------|------|
| `--json` | JSON 形式で出力 |
| `--sarif` | SARIF 2.1.0 形式で出力 |
| `--strict` | 警告をエラーとして扱う |
| `--format <fmt>` | 出力形式 (table) |

## moai spec close

```bash
moai spec close SPEC-ID
```

SPEC を単一コミットで `status: completed` へ原子的に遷移させます。

| フラグ | 説明 |
|--------|------|
| `--backfill-only` | progress.md の backfill のみ実行 |
| `--dry-run` | コミットせずプレビュー |
| `--force` | 確認なしで強制クローズ |
| `--json` | JSON 形式で出力 |

## moai spec audit

```bash
moai spec audit
```

`.moai/specs/SPEC-*/` をスキャンして各 SPEC を era ヒューリスティックで分類し、modern-era ステータスドリフトを検出します。

| フラグ | 説明 |
|--------|------|
| `--json` | JSON 形式で出力 |
| `--filter-era <era>` | era でフィルタ |
| `--filter-spec <id>` | SPEC ID でフィルタ |
| `--include-grandfathered` | grandfather era SPEC を含める |
| `--strict` | 厳格モード |

## moai spec archive

```bash
moai spec archive --dry-run   # 対象確認 (移動なし)
moai spec archive --yes       # 計画を適用
```

grace ウィンドウ (デフォルト 90 日) より古い terminal SPEC をアーカイブします。

| フラグ | 説明 |
|--------|------|
| `--dry-run` | 移動なしで対象集合を報告 |
| `--yes` | 移動を確定 (適用に必須) |
| `--grace-days <n>` | grace ウィンドウの日数 (0 = デフォルト 90) |
| `--json` | 計画を JSON で出力 |

## 関連ドキュメント

- [SPEC ベース開発](/ja/core-concepts/spec-based-dev)
- [CLI 概要](/ja/getting-started/cli)
