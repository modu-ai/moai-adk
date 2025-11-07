---
title: TDD開発ガイド
description: Test-Driven Development完全ガイド - RED、GREEN、REFACTORサイクルで安定したコードを作成
status: stable
---

# TDD (Test-Driven Development) 開発ガイド

**TDD (Test-Driven Development)**は MoAI-ADKの核心原則です。このガイドでは、RED-GREEN-REFACTORサイクルを通じて、テスト駆動開発を実現する方法を学びます。

## 📚 TDDとは?

Test-Driven Developmentは次の順序で進みます:

1. **RED**: 失敗するテストを書く
2. **GREEN**: テストを通過させる最小限のコードを書く
3. **REFACTOR**: コード品質を改善する

このサイクルを繰り返すことで、要求仕様を満たす安定したコードを作成します。

## 🎯 各段階のガイド

### [RED段階](red.md)
- 失敗するテストを書く
- テストケース設計
- 境界値およびエラー処理

### [GREEN段階](green.md)
- 最小限の実装 (YAGNI原則)
- 迅速なテスト通過
- パフォーマンス vs 機能のバランス

### [REFACTOR段階](refactor.md)
- コードのクリーンアップと最適化
- SOLID原則の適用
- 可読性の向上

## 🔄 AlfredによるTDD

Alfred SuperAgentがTDDサイクルを自動化します:

- `/alfred:2-run SPEC-ID`: RED-GREEN-REFACTORを自動実行
- 各段階の自動検証
- Git コミットの自動化

[Alfred ワークフローでTDD開始](../alfred/2-run.md)

## 📊 TDDのメリット

| 項目 | 効果 |
|------|------|
| **テストカバレッジ** | 87%以上を自動達成 |
| **バグ早期発見** | 開発中に95%以上を検出 |
| **リファクタリング安全性** | テストによる完全な保護 |
| **ドキュメンテーション** | テスト自体が実行可能なドキュメント |
| **設計改善** | テスト可能な設計が自動形成 |

## 🚀 次のステップ

- [RED: 失敗するテストを書く](red.md)
- [GREEN: 最小実装で通過](green.md)
- [REFACTOR: コード改善](refactor.md)
- [Alfred 2-run ワークフロー](../alfred/2-run.md)

---

**Learn more**: MoAI-ADKのTDD原則は、SPEC-First開発哲学の核心です。SPEC定義後にTDDで実装すると、要求仕様を完璧に満たすコードが完成します。
