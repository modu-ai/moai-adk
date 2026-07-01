---
title: 貢献する
weight: 110
draft: false
---

MoAI-ADK はオープンソースプロジェクトであり、皆様の貢献をお待ちしています。このガイドはプロジェクトへの貢献方法を説明します。

## クイックスタート

1. リポジトリを**フォーク**します
2. フィーチャーブランチを作成: `git checkout -b feature/my-feature`
3. テストを作成 (新しいコードは TDD、既存コードは特性化テスト)
4. すべてのテストが通ることを確認: `make test`
5. リントが通ることを確認: `make lint`
6. コードをフォーマット: `make fmt`
7. Conventional Commit メッセージでコミット
8. プルリクエストを作成

## コード品質要件

| 項目 | 基準 |
|------|------|
| テストカバレッジ | **85%** 以上 |
| リントエラー | **0** 個 |
| タイプエラー | **0** 個 |
| コミットメッセージ | Conventional Commits 形式 |

## コミットメッセージ形式

```
<type>(<scope>): <description>

[オプションの本文]

[オプションのフッター]
```

### タイプ

| タイプ | 説明 |
|--------|------|
| `feat` | 新しい機能 |
| `fix` | バグ修正 |
| `docs` | ドキュメント変更 |
| `style` | コードフォーマット (機能変更なし) |
| `refactor` | リファクタリング (機能変更なし) |
| `perf` | パフォーマンス改善 |
| `test` | テスト追加/変更 |
| `chore` | ビルド/ツール変更 |
| `revert` | 前のコミットを戻す |

### 例

```
feat(template): add SessionEnd hook to settings.json generator
fix(cli): prevent race condition in hook execution
test(settings): add TestEnsureGlobalSettingsEnv test cases
docs(readme): update agent count and statistics
```

## 開発環境セットアップ

### 必須ツール

- **Go 1.26+** — コア開発言語
- **Git** — バージョン管理
- **make** — ビルドコマンド

### 主要コマンド

```bash
make build        # プロジェクトをビルド
make test         # テストを実行
make test-race    # Race condition を検出
make lint         # リンターを実行
make fmt          # コードをフォーマット
make install      # ローカルにインストール
make clean        # ビルド成果物をクリーンアップ
```

## プルリクエストガイド

### PR 作成時

- 明確で簡潔なタイトル (70 文字以内)
- 変更内容の要約 (Summary セクション)
- テスト計画 (Test Plan セクション)
- 関連イシューの参照 (例: `Fixes #123`)

### PR チェックリスト

- [ ] テストを追加/更新
- [ ] すべてのテストが通過 (`make test`)
- [ ] リントが通過 (`make lint`)
- [ ] コミットメッセージが Conventional Commits 形式
- [ ] ドキュメントを更新 (必要に応じて)

## コミュニティ

- **イシュートラッカー**: [GitHub Issues](https://github.com/modu-ai/moai-adk/issues) — バグレポート、機能リクエスト
- **公式ドキュメント**: [adk.mo.ai.kr](https://adk.mo.ai.kr)

## ライセンス

[Apache License 2.0](https://github.com/modu-ai/moai-adk/blob/main/LICENSE) — 自由に使用、変更、配布できます。
