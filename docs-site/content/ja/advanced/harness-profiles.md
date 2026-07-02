---
title: ハーネスプロファイルと評価システム
weight: 75
draft: false
---
# ハーネスプロファイルと評価システム


3階層ハーネスレベルと4次元評価プロファイルによる適応型品質検証システムです。

## 概要

MoAI-ADKのハーネス(Harness)は **3階層適応型品質検証システム** です。SPECの複雑度に
応じて自動的に検証の深さを調整します。sync-auditorエージェントが4次元スコアリングで
独立した懐疑的な品質評価を行います。

## 3階層ハーネスレベル

| レベル | 説明 | 適用タイミング | sync-auditor |
|------|------|----------|-----------------|
| **minimal** | 迅速な検証 | 単純な変更 (typo、設定修正) | 省略可能 |
| **standard** | 基本品質検証 | ほとんどの作業 | 任意 |
| **thorough** | 完全検証 + TRUST 5 | 複雑なSPEC、大規模変更 | 必須 |

ハーネスレベルはSPECスコープに基づき、**複雑度推定器**(Complexity Estimator)が自動的に
決定します。

## 4次元スコアリング

sync-auditorは4つの次元でスコアを付けます:

| 次元 | 説明 | デフォルトMust-Pass |
|------|------|---------------|
| **Functionality** | 機能完成度 — 意図した目的を達成しているか | はい |
| **Security** | セキュリティ — OWASP、認証、権限、入力検証 | はい |
| **Craft** | コード品質 — 可読性、構造、テストカバレッジ | いいえ |
| **Consistency** | 一貫性 — プロジェクト規則、コードスタイル遵守 | いいえ |

### スコア範囲

各次元は0.0 ~ 1.0のスコアを受け取ります。

### ルーブリックアンカー

すべての評価基準は4段階のルーブリックアンカーを持ちます:

| スコア | レベル | 意味 |
|------|------|------|
| 0.25 | 未達 | 基本要件を満たしていない |
| 0.50 | 部分的 | 一部充足、改善が必要 |
| 0.75 | 充足 | ほぼ充足、小規模な改善 |
| 1.00 | 優秀 | すべての基準を完全に充足 |

## 評価プロファイル

`.moai/config/evaluator-profiles/`に4つのプロファイルが提供されます:

| プロファイル | 説明 | 適した場合 |
|--------|------|------------|
| `default.md` | バランスの取れた基本プロファイル | ほとんどの作業 |
| `strict.md` | 厳格な基準 | セキュリティ重視の作業 |
| `lenient.md` | 寛容な基準 | プロトタイピング |
| `frontend.md` | フロントエンド特化 | UI/UX作業 |

## 評価者バイアス防止 (5つのメカニズム)

評価者の甘さを防ぐため、5つのメカニズムが機能します:

| # | メカニズム | 説明 |
|---|---------|------|
| 1 | **ルーブリックアンカリング** | スコアにルーブリックによる正当化が必須 |
| 2 | **回帰ベースライン** | 過去のプロジェクトと比較した過度なスコア上昇を検出 |
| 3 | **Must-Passファイアウォール** | 必須基準は他の領域のスコアで補うことはできない |
| 4 | **独立再評価** | 5回ごとに独立再評価 (偏差 > 0.10 の場合は再調整) |
| 5 | **アンチパターン照合** | 既知のアンチパターンが検出された場合、該当次元のスコアを0.50に制限 |

## Evaluator Memory Scope

評価者の判断記憶は **反復ごとに一時的** です。GAN Loopの各反復でsync-auditorは
新しいコンテキストで再起動され、前回の反復の判断根拠は新しいプロンプトに含まれません。
Sprint Contractの状態のみが反復間で維持されます。

## 設定

`.moai/config/sections/harness.yaml`で設定します:

```yaml
harness:
  level: auto              # auto | minimal | standard | thorough
  evaluator:
    memory_scope: per_iteration   # FROZEN — 変更不可
    profiles:
      default: .moai/config/evaluator-profiles/default.md
      strict: .moai/config/evaluator-profiles/strict.md
    aggregation: min              # min | mean
    must_pass_dimensions:
      - Functionality
      - Security
```

## 関連ドキュメント

- [ハーネスエンジニアリング](/ja/core-concepts/harness-engineering) — ハーネスの概念概要
- [TRUST 5品質](/ja/core-concepts/trust-5) — 5つの品質基準
- [Constitutionシステム](/ja/core-concepts/constitution) — FROZEN/Evolvableルール
- GAN Loop — デザイン品質検証の反復 (GAN Loopはadversarial評価者-判別者ループによる、品質改善のための反復検証パターンです)
