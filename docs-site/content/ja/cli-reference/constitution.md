---
title: moai constitution 憲法
weight: 84
draft: false
---

`moai constitution` は zone レジストリ (FROZEN/EVOLVABLE zone の成文化) を照会・検証します。規則のどの部分が凍結 (FROZEN) されていて変更できないか、どの部分が進化可能 (EVOLVABLE) かを管理するコマンドツリーです。

## サブコマンド

| コマンド | 説明 |
|--------|------|
| `moai constitution list` | zone レジストリ項目を一覧表示 |
| `moai constitution guard` | FROZEN zone 違反を検査 (CI 統合用) |
| `moai constitution amend` | 5-layer セーフティゲートを通過する憲法改正の提案 |
| `moai constitution validate` | ソースファイルに対する zone レジストリのドリフト・不変条件を検証 |

## moai constitution list

```bash
moai constitution list
moai constitution list --zone frozen --format json
```

| フラグ | 説明 |
|--------|------|
| `--zone <frozen\|evolvable>` | zone フィルタ |
| `--file <path>` | ファイルパスフィルタ (部分一致) |
| `--format <table\|json>` | 出力形式 |

## moai constitution guard

```bash
moai constitution guard --violations CONST-V3R2-001,CONST-V3R2-002
```

変更された規則 ID の一覧を受け取り、FROZEN zone 違反を検査します。CI 統合用です。

| フラグ | 説明 |
|--------|------|
| `--violations <ids>` | 変更された規則 ID の一覧 (カンマ区切りまたは繰り返しフラグ) |

## moai constitution amend

```bash
moai constitution amend --rule CONST-V3R2-001 --before "..." --after "..." --evidence "..."
```

FrozenGuard → Canary → ContradictionDetector → RateLimiter → HumanOversight の 5-layer セーフティゲートを通過して初めて適用されます。

| フラグ | 説明 |
|--------|------|
| `--rule <id>` | 規則 ID (CONST-V3R2-NNN) [必須] |
| `--before <text>` | 現行条項のテキスト [必須] |
| `--after <text>` | 新しい条項のテキスト [必須] |
| `--evidence <text>` | 改正根拠 (Frozen zone では必須) |
| `--dry-run` | ファイル変更なしでシミュレーションのみ |

## moai constitution validate

```bash
moai constitution validate
```

レジストリの各項目の条項がソースファイルに存在するかを確認し、zone_class enum・canary_gate の不変条件を検証してドリフトを報告します。

| フラグ | 説明 |
|--------|------|
| `--strict` | 厳格モード (すべての検査を強制) |
| `--fail-on-warning` | 警告をエラーとして扱う (`--strict` を含む) |
| `--format <text\|json>` | 出力形式 |

終了コード: 0=正常、1=ドリフト/エラー、2=致命的 (ソースファイルなし)。

## 関連ドキュメント

- [CLI 概要](/ja/getting-started/cli)
