---
title: Constitution システム
weight: 35
draft: false
---

MoAI-ADKの不変ルール(FROZEN)と進化可能なルール(Evolvable)を管理する憲法的制約システムです。

## 概要

MoAI-ADKは **Constitution(憲法)** システムを通じて、AIエージェントが任意に変更できない
不変制約(FROZEN Zone)と、学習を通じて改善できる進化可能な制約(Evolvable Zone)を
区別します。これはハーネスエンジニアリングの中核となる安全メカニズムです。

## FROZEN vs Evolvable

### FROZEN Zone (不変)

AIエージェントが絶対に修正できないルールです。人間の開発者のみが変更できます。

**代表項目**:

| 項目 | 説明 | ソース |
|------|------|------|
| TRUST 5 | 5つの品質基準 | moai-constitution.md |
| SPEC + EARS | 仕様書フォーマット | spec-workflow.md |
| AskUserQuestion独占 | ユーザー質問チャネル | agent-common-protocol.md |
| 評価次元4つ | Functionality/Security/Craft/Consistency | harness/scorer.go |
| ルーブリックアンカー4段階 | 0.25/0.50/0.75/1.00 | harness/rubric.go |
| 合格閾値下限 | 最低0.60 (下げられない) | design-constitution.md |
| デザインパイプライン順序 | manager-spec が最初、sync-auditor が最後 | design-constitution.md |

### Evolvable Zone (進化可能)

学習(lessons)と研究(research)を通じて改善提案が可能なルールです。

**代表項目**:

| 項目 | 説明 |
|------|------|
| スキル本文の内容 | moai-domain-*スキルの詳細内容 |
| パイプライン重み | design.yamlのphase_weights |
| 反復上限 | design.yamlのiteration_limits |
| エージェント行動ルール | Surface Assumptions、Enforce Simplicity など |

## Zone Registry

すべてのHARD条項を列挙する **単一の真実の情報源**(Single Source of Truth)です。

### ID割り当てルール

```
CONST-V3R2-NNN (3桁以上のzero-padding)

001-050: 既存のHARD条項
051-099: design constitutionミラーエントリ
100-149: design overflow (自動拡張)
150+: 新規追加
```

### Canary Gate

FROZEN条項は`canary_gate: true`を持ちます。変更前にcanary検証が必須です。

```yaml
# Zone Registry エントリ例
- id: CONST-V3R2-154
  zone: Frozen
  file: internal/harness/scorer.go
  anchor: "#dimension-enum"
  clause: "Dimension enum FROZEN at 4 values"
  canary_gate: true
```

## 安全アーキテクチャ (5階層)

Constitutionシステムは5階層の安全アーキテクチャによって保護されています:

### Layer 1: Frozen Guard

書き込み操作の前に、対象ファイルがFROZEN zoneでないかを確認します。違反時は書き込み
ブロック + ロギング + ユーザー通知。

### Layer 2: Canary Check

提案された変更をメモリ上に適用し、直近3つのプロジェクトを再評価します。スコアの
低下が0.10を超える場合、変更を拒否します。

### Layer 3: Contradiction Detector

新しい学習が既存のルールと矛盾する場合、両方をユーザーに提示します。自動上書きは
一切発生しません。

### Layer 4: Rate Limiter

進化の速度を制限します:

| パラメータ | デフォルト値 | 説明 |
|-----------|--------|------|
| `max_evolution_rate_per_week` | 3 | 週あたりの最大進化回数 |
| `cooldown_hours` | 24 | 進化間の最小待機時間 |
| `max_active_learnings` | 50 | アクティブな学習項目の最大数 |

### Layer 5: Human Oversight

`require_approval: true`の場合、すべての進化提案にユーザー承認が必要です。

## CLIでの活用

```bash
# registry全体の照会
moai constitution list

# Frozen zoneフィルタ
moai constitution list --zone frozen

# 特定ファイルの条項のみ照会
moai constitution list --file internal/harness/scorer.go

# JSON形式出力
moai constitution list --format json
```

## 関連ドキュメント

- [TRUST 5品質](/ja/core-concepts/trust-5) — 5つの品質基準
- [ハーネスエンジニアリング](/ja/core-concepts/harness-engineering) — ハーネスの概念概要
- [SPECベース開発](/ja/core-concepts/spec-based-dev) — SPECワークフロー
